package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	// Get database name from URI or use default
	DB = client.Database("split-it")

	fmt.Println("✅ MongoDB Connected")
}

// GetDB returns the database instance
func GetDB() *mongo.Database {
	return DB
}
