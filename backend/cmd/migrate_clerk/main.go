// Run from backend: go run ./cmd/migrate_clerk (audit) or go run ./cmd/migrate_clerk --apply.
// Requires MONGODB_URI, MONGODB_DATABASE, FIREBASE_SERVICE_ACCOUNT_PATH, and CLERK_SECRET_KEY.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
	"splitpay/backend/models"
)

type clerkRecord struct {
	ID                    string `json:"id"`
	ExternalID            string `json:"external_id"`
	PrimaryEmailAddressID string `json:"primary_email_address_id"`
	EmailAddresses        []struct {
		ID           string `json:"id"`
		EmailAddress string `json:"email_address"`
		Verification struct {
			Status string `json:"status"`
		} `json:"verification"`
	} `json:"email_addresses"`
}

func (user clerkRecord) verifiedEmail() string {
	for _, address := range user.EmailAddresses {
		if address.ID == user.PrimaryEmailAddressID && address.Verification.Status == "verified" {
			return address.EmailAddress
		}
	}
	return ""
}

func main() {
	apply := flag.Bool("apply", false, "create Clerk users and link Mongo profiles")
	flag.Parse()
	_ = godotenv.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	for _, key := range []string{"MONGODB_URI", "FIREBASE_SERVICE_ACCOUNT_PATH", "CLERK_SECRET_KEY"} {
		if os.Getenv(key) == "" {
			log.Fatalf("%s is required", key)
		}
	}
	name := os.Getenv("MONGODB_DATABASE")
	if name == "" {
		name = "split-it"
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database(name)
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")))
	if err != nil {
		log.Fatal(err)
	}
	auth, err := app.Auth(ctx)
	if err != nil {
		log.Fatal(err)
	}
	profiles := db.Collection("users")
	cursor, err := profiles.Find(ctx, bson.M{"firebaseUid": bson.M{"$exists": true, "$ne": ""}})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		log.Fatal(err)
	}
	// A group can predate its owner's profile. Stop cutover if any owner/member has no profile.
	uidSet := map[string]bool{}
	for _, user := range users {
		uidSet[user.FirebaseUID] = true
	}
	groups, err := db.Collection("groups").Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer groups.Close(ctx)
	missing := map[string]bool{}
	for groups.Next(ctx) {
		var group models.Group
		if err := groups.Decode(&group); err != nil {
			log.Fatal(err)
		}
		if group.UserID != "" && !uidSet[group.UserID] {
			missing[group.UserID] = true
		}
		for _, member := range group.Members {
			if member.UID != "" && !uidSet[member.UID] {
				missing[member.UID] = true
			}
		}
	}
	if err := groups.Err(); err != nil {
		log.Fatal(err)
	}
	if len(missing) > 0 {
		log.Fatalf("%d group identities have no user profile; create/reconcile profiles before migration", len(missing))
	}
	ready := 0
	for _, profile := range users {
		firebaseUser, err := auth.GetUser(ctx, profile.FirebaseUID)
		if err != nil {
			log.Printf("BLOCK %s: Firebase account lookup: %v", profile.FirebaseUID, err)
			continue
		}
		if !firebaseUser.EmailVerified || firebaseUser.Email == "" || !strings.EqualFold(firebaseUser.Email, profile.Email) {
			log.Printf("BLOCK %s: Firebase email is unverified or disagrees with profile", profile.FirebaseUID)
			continue
		}
		clerkUser, err := findClerk(ctx, profile.FirebaseUID)
		if err != nil {
			log.Printf("BLOCK %s: Clerk lookup: %v", profile.FirebaseUID, err)
			continue
		}
		if clerkUser.ID == "" && *apply {
			clerkUser, err = createClerk(ctx, profile.FirebaseUID, firebaseUser.Email, profile.Name)
			if err != nil {
				log.Printf("BLOCK %s: Clerk creation: %v", profile.FirebaseUID, err)
				continue
			}
		}
		if clerkUser.ID != "" && clerkUser.ExternalID != profile.FirebaseUID {
			log.Printf("BLOCK %s: Clerk external ID mismatch", profile.FirebaseUID)
			continue
		}
		if clerkUser.ID != "" && !strings.EqualFold(clerkUser.verifiedEmail(), firebaseUser.Email) {
			log.Printf("BLOCK %s: Clerk primary email is unverified or mismatched", profile.FirebaseUID)
			continue
		}
		if *apply {
			result, err := profiles.UpdateOne(ctx, bson.M{"_id": profile.ID, "$or": []bson.M{{"clerkUserId": bson.M{"$exists": false}}, {"clerkUserId": clerkUser.ID}}}, bson.M{"$set": bson.M{"internalUid": profile.FirebaseUID, "clerkUserId": clerkUser.ID, "emailVerified": true, "updatedAt": time.Now()}})
			if err != nil || result.MatchedCount != 1 {
				log.Printf("BLOCK %s: Mongo link failed: %v", profile.FirebaseUID, err)
				continue
			}
		}
		ready++
	}
	log.Printf("%d/%d profiles %s. Do not cut over until all are ready.", ready, len(users), map[bool]string{true: "linked", false: "eligible or already provisioned"}[*apply])
	if ready != len(users) {
		os.Exit(1)
	}
}

func clerkRequest(ctx context.Context, method, endpoint string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.clerk.com/v1"+endpoint, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CLERK_SECRET_KEY"))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Clerk status %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
}

func findClerk(ctx context.Context, uid string) (clerkRecord, error) {
	var records []clerkRecord
	if err := clerkRequest(ctx, http.MethodGet, "/users?external_id%5B%5D="+url.QueryEscape(uid)+"&limit=2", nil, &records); err != nil {
		return clerkRecord{}, err
	}
	for _, record := range records {
		if record.ExternalID == uid {
			return record, nil
		}
	}
	if len(records) > 0 {
		return clerkRecord{}, errors.New("unexpected external ID result")
	}
	return clerkRecord{}, nil
}

func createClerk(ctx context.Context, uid, email, name string) (clerkRecord, error) {
	var record clerkRecord
	parts := strings.Fields(name)
	first := ""
	last := ""
	if len(parts) > 0 {
		first = parts[0]
		last = strings.Join(parts[1:], " ")
	}
	body := map[string]any{"email_address": []string{email}, "external_id": uid, "first_name": first, "last_name": last, "skip_password_requirement": true}
	err := clerkRequest(ctx, http.MethodPost, "/users", body, &record)
	return record, err
}
