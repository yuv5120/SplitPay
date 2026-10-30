package config

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestVerifyClerkTokenRequiresAudienceAndTimeClaims(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	oldKey, oldIssuer, oldParty := clerkPublicKey, clerkIssuer, clerkParty
	defer func() { clerkPublicKey, clerkIssuer, clerkParty = oldKey, oldIssuer, oldParty }()
	clerkPublicKey = &key.PublicKey
	clerkIssuer = "https://example.clerk.accounts.dev"
	clerkParty = "http://localhost:3000"
	now := time.Now().Unix()
	base := jwt.MapClaims{"iss": clerkIssuer, "azp": clerkParty, "sub": "user_123", "sid": "sess_123", "exp": now + 300, "nbf": now - 5}
	sign := func(claims jwt.MapClaims) string {
		value, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	if _, err := VerifyClerkToken(sign(base)); err != nil {
		t.Fatalf("valid session rejected: %v", err)
	}
	for _, claim := range []string{"exp", "nbf", "iss", "azp", "sub", "sid"} {
		claims := jwt.MapClaims{}
		for k, v := range base {
			if k != claim {
				claims[k] = v
			}
		}
		if _, err := VerifyClerkToken(sign(claims)); err == nil {
			t.Errorf("missing %s accepted", claim)
		}
	}
	expired := jwt.MapClaims{}
	for k, v := range base {
		expired[k] = v
	}
	expired["exp"] = now - 1
	if _, err := VerifyClerkToken(sign(expired)); err == nil {
		t.Error("expired session accepted")
	}
}
