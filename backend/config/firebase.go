package config

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var FirebaseAuth *auth.Client

// InitializeFirebase initializes Firebase Admin SDK
func InitializeFirebase() {
	serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
	if serviceAccountPath == "" {
		log.Println("⚠️  FIREBASE_SERVICE_ACCOUNT_PATH not set")
		log.Println("⚠️  Continuing without Firebase (download service account key from Firebase Console)")
		return
	}

	opt := option.WithCredentialsFile(serviceAccountPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Printf("❌ Firebase initialization error: %v\n", err)
		log.Println("⚠️  Continuing without Firebase")
		return
	}

	FirebaseAuth, err = app.Auth(context.Background())
	if err != nil {
		log.Printf("❌ Firebase Auth initialization error: %v\n", err)
		log.Println("⚠️  Continuing without Firebase")
		return
	}

	fmt.Println("✅ Firebase Admin initialized")
}

// GetFirebaseAuth returns the Firebase Auth client
func GetFirebaseAuth() *auth.Client {
	return FirebaseAuth
}
