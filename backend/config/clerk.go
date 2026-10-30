package config

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type ClerkIdentity struct {
	UserID    string
	SessionID string
}

type ClerkUser struct {
	ID                    string `json:"id"`
	ExternalID            string `json:"external_id"`
	FirstName             string `json:"first_name"`
	LastName              string `json:"last_name"`
	PrimaryEmailAddressID string `json:"primary_email_address_id"`
	EmailAddresses        []struct {
		ID           string `json:"id"`
		EmailAddress string `json:"email_address"`
		Verification struct {
			Status string `json:"status"`
		} `json:"verification"`
	} `json:"email_addresses"`
}

var clerkPublicKey *rsa.PublicKey
var clerkIssuer string
var clerkParty string
var clerkSecret string
var clerkHTTP = &http.Client{Timeout: 5 * time.Second}

func InitializeClerk() error {
	key := strings.ReplaceAll(os.Getenv("CLERK_JWT_KEY"), `\n`, "\n")
	clerkIssuer = strings.TrimSuffix(os.Getenv("CLERK_ISSUER"), "/")
	clerkParty = strings.TrimSuffix(os.Getenv("CLERK_AUTHORIZED_PARTY"), "/")
	clerkSecret = os.Getenv("CLERK_SECRET_KEY")
	if key == "" || clerkIssuer == "" || clerkParty == "" || clerkSecret == "" {
		return errors.New("CLERK_JWT_KEY, CLERK_ISSUER, CLERK_AUTHORIZED_PARTY, and CLERK_SECRET_KEY are required")
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(key))
	if err != nil {
		return fmt.Errorf("invalid Clerk JWT public key: %w", err)
	}
	clerkPublicKey = publicKey
	return nil
}

func VerifyClerkToken(raw string) (ClerkIdentity, error) {
	if clerkPublicKey == nil {
		return ClerkIdentity{}, errors.New("Clerk is not initialized")
	}
	claims := jwt.MapClaims{}
	token, err := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()})).ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) { return clerkPublicKey, nil })
	if err != nil || token == nil || !token.Valid {
		return ClerkIdentity{}, errors.New("invalid or expired Clerk session")
	}
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) || !claims.VerifyNotBefore(time.Now().Unix(), true) {
		return ClerkIdentity{}, errors.New("Clerk session time claims missing or invalid")
	}
	if !claims.VerifyIssuer(clerkIssuer, true) {
		return ClerkIdentity{}, errors.New("wrong Clerk issuer")
	}
	if claims["azp"] != clerkParty {
		return ClerkIdentity{}, errors.New("wrong Clerk authorized party")
	}
	if status, ok := claims["sts"].(string); ok && status != "" && status != "active" {
		return ClerkIdentity{}, errors.New("Clerk session is not active")
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return ClerkIdentity{}, errors.New("Clerk subject missing")
	}
	sid, ok := claims["sid"].(string)
	if !ok || sid == "" {
		return ClerkIdentity{}, errors.New("Clerk session missing")
	}
	return ClerkIdentity{UserID: sub, SessionID: sid}, nil
}

func FetchClerkUser(ctx context.Context, userID string) (ClerkUser, error) {
	if clerkSecret == "" {
		return ClerkUser{}, errors.New("Clerk secret not configured")
	}
	endpoint := "https://api.clerk.com/v1/users/" + url.PathEscape(userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ClerkUser{}, err
	}
	req.Header.Set("Authorization", "Bearer "+clerkSecret)
	resp, err := clerkHTTP.Do(req)
	if err != nil {
		return ClerkUser{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ClerkUser{}, fmt.Errorf("Clerk user lookup failed with status %d", resp.StatusCode)
	}
	var user ClerkUser
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&user); err != nil {
		return ClerkUser{}, err
	}
	if user.ID != userID {
		return ClerkUser{}, errors.New("Clerk user mismatch")
	}
	return user, nil
}

func (user ClerkUser) VerifiedPrimaryEmail() (string, bool) {
	for _, email := range user.EmailAddresses {
		if email.ID == user.PrimaryEmailAddressID && email.Verification.Status == "verified" {
			value := strings.ToLower(strings.TrimSpace(email.EmailAddress))
			if value != "" {
				return value, true
			}
		}
	}
	return "", false
}
