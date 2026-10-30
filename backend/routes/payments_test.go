package routes

import (
	"splitpay/backend/config"
	"splitpay/backend/models"
	"testing"
)

func TestGatewayPaymentMatchesExactCapturedOrder(t *testing.T) {
	attempt := models.PaymentAttempt{GatewayOrderID: "order_123", AmountPaise: 1250}
	payment := config.RazorpayPayment{ID: "pay_123", OrderID: "order_123", Amount: 1250, Currency: "INR", Status: "captured", Captured: true}
	if !gatewayPaymentMatches(attempt, payment) {
		t.Fatal("valid captured test payment rejected")
	}
	payment.Amount++
	if gatewayPaymentMatches(attempt, payment) {
		t.Fatal("wrong amount accepted")
	}
	payment.Amount--
	payment.OrderID = "order_other"
	if gatewayPaymentMatches(attempt, payment) {
		t.Fatal("wrong order accepted")
	}
	payment.OrderID = "order_123"
	payment.Captured = false
	if gatewayPaymentMatches(attempt, payment) {
		t.Fatal("uncaptured payment accepted")
	}
}
