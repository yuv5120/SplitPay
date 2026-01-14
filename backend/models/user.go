package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirebaseUID   string             `bson:"firebaseUid" json:"firebaseUid"`
	Email         string             `bson:"email" json:"email"`
	Name          string             `bson:"name" json:"name"`
	Phone         string             `bson:"phone" json:"phone"`
	EmailVerified bool               `bson:"emailVerified" json:"emailVerified"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// UserResponse is the response structure for user data
type UserResponse struct {
	ID            string    `json:"id"`
	FirebaseUID   string    `json:"firebaseUid"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Phone         string    `json:"phone"`
	EmailVerified bool      `json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
}
