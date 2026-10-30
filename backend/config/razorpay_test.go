package config

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRazorpayTestOrderAndVerification(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		user, password, ok := req.BasicAuth()
		if !ok || user != "rzp_test_example" || password != "secret" {
			t.Error("missing basic auth")
		}
		switch req.URL.Path {
		case "/orders":
			if req.Method != http.MethodPost {
				t.Error("wrong order method")
			}
			w.Write([]byte(`{"id":"order_123","amount":1250,"currency":"INR","receipt":"attempt123"}`))
		case "/payments/pay_123":
			w.Write([]byte(`{"id":"pay_123","order_id":"order_123","amount":1250,"currency":"INR","status":"captured","captured":true}`))
		default:
			t.Errorf("unexpected path %s", req.URL.Path)
		}
	}))
	defer server.Close()
	r := &RazorpayClient{KeyID: "rzp_test_example", keySecret: "secret", httpClient: server.Client(), baseURL: server.URL}
	order, err := r.CreateOrder(context.Background(), 1250, "attempt123")
	if err != nil || order.ID != "order_123" {
		t.Fatalf("create order: %+v, %v", order, err)
	}
	payment, err := r.FetchPayment(context.Background(), "pay_123")
	if err != nil || payment.Status != "captured" || payment.Amount != order.Amount {
		t.Fatalf("fetch payment: %+v, %v", payment, err)
	}
	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte("order_123|pay_123"))
	signature := hex.EncodeToString(mac.Sum(nil))
	if !r.VerifySignature(order.ID, payment.ID, signature) || r.VerifySignature(order.ID, "pay_456", signature) || r.VerifySignature(order.ID, payment.ID, strings.Repeat("0", 64)) {
		t.Fatal("checkout signature validation failed")
	}
}

func TestRazorpayRejectsMismatchedOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(`{"id":"order_123","amount":1251,"currency":"INR","receipt":"attempt123"}`))
	}))
	defer server.Close()
	r := &RazorpayClient{KeyID: "rzp_test_example", keySecret: "secret", httpClient: server.Client(), baseURL: server.URL}
	if _, err := r.CreateOrder(context.Background(), 1250, "attempt123"); err == nil {
		t.Fatal("mismatched amount accepted")
	}
}

func TestRazorpayWebhookSignature(t *testing.T) {
	t.Setenv("RAZORPAY_TEST_WEBHOOK_SECRET", "webhook-secret")
	body := []byte(`{"event":"payment.captured"}`)
	mac := hmac.New(sha256.New, []byte("webhook-secret"))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))
	if !VerifyRazorpayWebhook(body, signature) || VerifyRazorpayWebhook(append(body, ' '), signature) || VerifyRazorpayWebhook(body, "bad") {
		t.Fatal("webhook signature validation failed")
	}
}
