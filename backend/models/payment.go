package models

import "time"

const (
	PaymentPending        = "pending"
	PaymentSucceeded      = "succeeded"
	PaymentFailed         = "failed"
	PaymentCancelled      = "cancelled"
	PaymentReviewRequired = "review_required"
)

// PaymentAttempt records simulated or Razorpay test-mode checkout activity.
type PaymentAttempt struct {
	ID               string     `bson:"id" json:"id"`
	From             string     `bson:"from" json:"from"`
	To               string     `bson:"to" json:"to"`
	AmountPaise      int64      `bson:"amountPaise" json:"amountPaise"`
	Method           string     `bson:"method" json:"method"`
	GatewayOrderID   string     `bson:"gatewayOrderId,omitempty" json:"gatewayOrderId,omitempty"`
	GatewayPaymentID string     `bson:"gatewayPaymentId,omitempty" json:"gatewayPaymentId,omitempty"`
	Status           string     `bson:"status" json:"status"`
	BaseRevision     int64      `bson:"baseRevision" json:"-"`
	CreatedAt        time.Time  `bson:"createdAt" json:"createdAt"`
	CompletedAt      *time.Time `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
	RecipientReadAt  *time.Time `bson:"recipientReadAt,omitempty" json:"-"`
}

type ReminderEvent struct {
	ID          string     `bson:"id" json:"id"`
	From        string     `bson:"from" json:"from"`
	To          string     `bson:"to" json:"to"`
	AmountPaise int64      `bson:"amountPaise" json:"amountPaise"`
	CreatedAt   time.Time  `bson:"createdAt" json:"createdAt"`
	ReadAt      *time.Time `bson:"readAt,omitempty" json:"-"`
}

func OutstandingTo(balances map[string]int64, from, to string) int64 {
	for _, suggestion := range SuggestSettlements(balances) {
		if suggestion.From == from && suggestion.To == to {
			return suggestion.AmountPaise
		}
	}
	return 0
}
