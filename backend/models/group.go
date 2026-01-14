package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Member represents a member in a group
type Member struct {
	ID   string `bson:"id" json:"id"`
	Name string `bson:"name" json:"name"`
}

// Expense represents an expense in a group
type Expense struct {
	ID           string    `bson:"id" json:"id"`
	Description  string    `bson:"description" json:"description"`
	Amount       float64   `bson:"amount" json:"amount"`
	PaidBy       string    `bson:"paidBy" json:"paidBy"`
	Participants []string  `bson:"participants" json:"participants"`
	Date         time.Time `bson:"date" json:"date"`
}

// Group represents a group with members and expenses
type Group struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	GroupID   string             `bson:"id" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Members   []Member           `bson:"members" json:"members"`
	Expenses  []Expense          `bson:"expenses" json:"expenses"`
	UserID    string             `bson:"userId" json:"userId"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// GroupResponse is the response structure for group data
type GroupResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Members   []Member  `json:"members"`
	Expenses  []Expense `json:"expenses"`
	CreatedAt time.Time `json:"createdAt"`
}
