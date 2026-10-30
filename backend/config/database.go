package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

// ConnectDB initializes MongoDB connection
func ConnectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("❌ MONGODB_URI environment variable is not set")
	}

	// MongoDB driver will automatically configure TLS for mongodb+srv:// URIs
	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetServerSelectionTimeout(30 * time.Second).
		SetConnectTimeout(30 * time.Second).
		SetMaxPoolSize(10).
		SetMinPoolSize(1)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("❌ MongoDB Connection Error: %v", err)
	}

	// Ping the database with a longer timeout
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer pingCancel()

	err = client.Ping(pingCtx, nil)
	if err != nil {
		log.Fatalf("❌ MongoDB Ping Error: %v", err)
	}

	// Keep the old default for existing deployments; new installs set MONGODB_DATABASE.
	databaseName := os.Getenv("MONGODB_DATABASE")
	if databaseName == "" {
		databaseName = "split-it"
	}
	DB = client.Database(databaseName)
	indexCtx, indexCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer indexCancel()
	if _, err := DB.Collection("groups").Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}, {Key: "id", Value: 1}},
	}); err != nil {
		log.Fatalf("Cannot create groups lookup index: %v", err)
	}
	if _, err := DB.Collection("users").Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys: bson.D{{Key: "firebaseUid", Value: 1}},
	}); err != nil {
		log.Fatalf("Cannot create users lookup index: %v", err)
	}
	if _, err := DB.Collection("users").Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys: bson.D{{Key: "clerkUserId", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true),
	}); err != nil {
		log.Fatalf("Cannot create Clerk user lookup index: %v", err)
	}
	if _, err := DB.Collection("users").Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys: bson.D{{Key: "internalUid", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true),
	}); err != nil {
		log.Fatalf("Cannot create internal user lookup index: %v", err)
	}
	if _, err := DB.Collection("groupInvitations").Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys: bson.D{{Key: "tokenHash", Value: 1}}, Options: options.Index().SetUnique(true),
	}); err != nil {
		log.Fatalf("Cannot create invite lookup index: %v", err)
	}
	if _, err := DB.Collection("groupInvitations").Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0),
	}); err != nil {
		log.Fatalf("Cannot create invite expiry index: %v", err)
	}
	if _, err := DB.Collection("expenseComments").Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys: bson.D{{Key: "groupId", Value: 1}, {Key: "expenseId", Value: 1}, {Key: "createdAt", Value: 1}},
	}); err != nil {
		log.Fatalf("Cannot create comment lookup index: %v", err)
	}
	if _, err := DB.Collection("expenseNotifications").Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys: bson.D{{Key: "groupId", Value: 1}, {Key: "recipientUid", Value: 1}, {Key: "createdAt", Value: -1}},
	}); err != nil {
		log.Fatalf("Cannot create notification lookup index: %v", err)
	}

	fmt.Println("✅ MongoDB Connected")
}

// GetDB returns the database instance
func GetDB() *mongo.Database {
	return DB
}
