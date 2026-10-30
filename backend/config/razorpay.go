package config

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type RazorpayClient struct {
	KeyID, keySecret string
	httpClient       *http.Client
	baseURL          string
}

type RazorpayOrder struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Receipt  string `json:"receipt"`
}

type RazorpayPayment struct {
	ID       string `json:"id"`
	OrderID  string `json:"order_id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Captured bool   `json:"captured"`
}

var razorpay *RazorpayClient

func InitializeRazorpayTest() error {
	id, secret := os.Getenv("RAZORPAY_TEST_KEY_ID"), os.Getenv("RAZORPAY_TEST_KEY_SECRET")
	if id == "" && secret == "" {
		return nil
	}
	if !strings.HasPrefix(id, "rzp_test_") || secret == "" {
		return errors.New("Razorpay Test Mode requires RAZORPAY_TEST_KEY_ID (rzp_test_...) and RAZORPAY_TEST_KEY_SECRET; live keys are refused")
	}
	razorpay = &RazorpayClient{KeyID: id, keySecret: secret, httpClient: &http.Client{Timeout: 8 * time.Second}, baseURL: "https://api.razorpay.com/v1"}
	return nil
}

func GetRazorpayTest() *RazorpayClient { return razorpay }

func (r *RazorpayClient) request(ctx context.Context, method, path string, body any, target any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, r.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.SetBasicAuth(r.KeyID, r.keySecret)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Razorpay API returned %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(target)
}

func (r *RazorpayClient) CreateOrder(ctx context.Context, amount int64, receipt string) (RazorpayOrder, error) {
	var order RazorpayOrder
	err := r.request(ctx, http.MethodPost, "/orders", map[string]any{"amount": amount, "currency": "INR", "receipt": receipt}, &order)
	if err != nil {
		return order, err
	}
	if order.ID == "" || order.Amount != amount || order.Currency != "INR" || order.Receipt != receipt {
		return order, errors.New("Razorpay order response did not match request")
	}
	return order, nil
}

func (r *RazorpayClient) FetchPayment(ctx context.Context, id string) (RazorpayPayment, error) {
	var payment RazorpayPayment
	if !strings.HasPrefix(id, "pay_") {
		return payment, errors.New("invalid Razorpay payment ID")
	}
	err := r.request(ctx, http.MethodGet, "/payments/"+id, nil, &payment)
	return payment, err
}

func (r *RazorpayClient) FetchOrderPayments(ctx context.Context, orderID string) ([]RazorpayPayment, error) {
	if !strings.HasPrefix(orderID, "order_") {
		return nil, errors.New("invalid Razorpay order ID")
	}
	var response struct {
		Items []RazorpayPayment `json:"items"`
	}
	err := r.request(ctx, http.MethodGet, "/orders/"+orderID+"/payments", nil, &response)
	return response.Items, err
}

func (r *RazorpayClient) VerifySignature(orderID, paymentID, signature string) bool {
	if orderID == "" || paymentID == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(r.keySecret))
	mac.Write([]byte(orderID + "|" + paymentID))
	want := mac.Sum(nil)
	got, err := hex.DecodeString(signature)
	return err == nil && hmac.Equal(got, want)
}

func VerifyRazorpayWebhook(body []byte, signature string) bool {
	secret := os.Getenv("RAZORPAY_TEST_WEBHOOK_SECRET")
	if secret == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	want := mac.Sum(nil)
	got, err := hex.DecodeString(signature)
	return err == nil && hmac.Equal(got, want)
}
