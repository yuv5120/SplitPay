// Run from backend with DEMO_EMAIL_A and DEMO_EMAIL_B after both users sign in once.
// Adds one isolated, repeatable reviewer group; never edits existing groups.
package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"splitpay/backend/models"
)

func main() {
	_ = godotenv.Load()
	uri := os.Getenv("MONGODB_URI")
	a := strings.ToLower(strings.TrimSpace(os.Getenv("DEMO_EMAIL_A")))
	b := strings.ToLower(strings.TrimSpace(os.Getenv("DEMO_EMAIL_B")))
	if uri == "" || a == "" || b == "" || a == b {
		log.Fatal("set MONGODB_URI and two distinct DEMO_EMAIL_A / DEMO_EMAIL_B values")
	}
	name := os.Getenv("MONGODB_DATABASE")
	if name == "" {
		name = "split-it"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database(name)
	lookup := func(email string) models.User {
		var user models.User
		if err := db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
			log.Fatalf("sign in first as %s: %v", email, err)
		}
		return user
	}
	owner, guest := lookup(a), lookup(b)
	uid := func(user models.User) string {
		if user.InternalUID != "" {
			return user.InternalUID
		}
		return user.FirebaseUID
	}
	if uid(owner) == "" || uid(guest) == "" {
		log.Fatal("both users need stable internal IDs")
	}
	groupID := "reviewer-demo-" + owner.ID.Hex() + "-" + guest.ID.Hex()
	now := time.Now()
	memberA, memberB := "demo-owner", "demo-guest"
	seedMembers := []models.Member{{ID: memberA, Name: owner.Name, UID: uid(owner)}, {ID: memberB, Name: guest.Name, UID: uid(guest)}}
	plan, err := models.NewPurchasePlan(models.PurchasePlanInput{Title: "Weekend team stay", Details: "Choose a place before booking", VotingDays: 7,
		Options: []models.PurchaseOptionInput{{Label: "Hotel", AmountPaise: 600000}, {Label: "Apartment", AmountPaise: 450000}}}, seedMembers, uid(guest), now)
	if err != nil {
		log.Fatal(err)
	}
	plan, err = plan.Vote(memberB, plan.Options[1].ID, now)
	if err != nil {
		log.Fatal(err)
	}
	var existing models.Group
	if err := db.Collection("groups").FindOne(ctx, bson.M{"id": groupID}).Decode(&existing); err == nil {
		if len(existing.PurchasePlans) == 0 {
			result, err := db.Collection("groups").UpdateOne(ctx, bson.M{"_id": existing.ID, "purchasePlans.0": bson.M{"$exists": false}}, bson.M{"$push": bson.M{"purchasePlans": plan}})
			if err != nil {
				log.Fatal(err)
			}
			if result.ModifiedCount == 1 {
				log.Printf("added purchase plan to existing demo group: %s", groupID)
			}
		}
		log.Printf("demo group ready: %s", groupID)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Fatal(err)
	}
	group := models.Group{
		ID: primitive.NewObjectID(), GroupID: groupID, Name: "Reviewer demo · shared flat", UserID: uid(owner),
		Members: seedMembers,
		Expenses: []models.Expense{
			{ID: "demo-internet", Description: "Internet · approved", AmountPaise: 120000, PaidBy: memberA, Participants: []string{memberA, memberB}, Date: now, CreatedBy: uid(owner), Status: models.ExpenseActive, Version: 1, Cycle: 1, Approvals: []models.ExpenseApproval{{MemberID: memberA, Status: "approved"}, {MemberID: memberB, Status: "approved"}}},
			{ID: "demo-groceries", Description: "Groceries · awaiting approval", AmountPaise: 84000, PaidBy: memberB, Participants: []string{memberA, memberB}, Date: now, CreatedBy: uid(guest), Status: models.ExpenseProposed, Version: 1, Cycle: 1, Approvals: []models.ExpenseApproval{{MemberID: memberA, Status: "pending"}, {MemberID: memberB, Status: "approved"}}},
		},
		Recurring:          []models.RecurringCost{{ID: "demo-electricity", Description: "Electricity", EstimatePaise: 520000, DueDay: 10, PaidBy: memberA, Participants: []string{memberA, memberB}, CreatedBy: uid(owner), Version: 1, CreatedAt: now}},
		PurchasePlans:      []models.PurchasePlan{plan},
		Activity:           []models.Activity{models.NewActivity("demo", uid(owner), "Reviewer demo group created")},
		MonthlyBudgetPaise: 800000, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.Collection("groups").InsertOne(ctx, group); err != nil {
		log.Fatal(err)
	}
	_, err = db.Collection("expenseComments").InsertOne(ctx, bson.M{"groupId": group.ID, "expenseId": "demo-groceries", "authorUid": uid(guest), "text": "Can you check this grocery bill before we settle?", "createdAt": now, "_id": primitive.NewObjectID()})
	if err != nil {
		log.Printf("demo comment could not be added: %v", err)
	}
	log.Printf("created %s. Sign in as %s and %s to review", groupID, a, b)
}
