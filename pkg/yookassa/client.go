// Package yookassa is a minimal client for the YooKassa Payments API v3
// (https://yookassa.ru/developers/api). It only implements the subset needed
// by CyberMate: creating a redirect payment, fetching a payment by id (used to
// independently verify webhook notifications), and issuing refunds.
package yookassa

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL           = "https://api.yookassa.ru/v3"
	defaultTimeout    = 20 * time.Second
	maxResponseBytes  = 1 << 20 // 1 MiB, payment objects are small JSON documents
	pathPayments      = "/payments"
	pathRefunds       = "/refunds"
	confirmationRedir = "redirect"
)

// Client talks to the YooKassa Payments API using shop id + secret key basic auth.
type Client struct {
	shopID     string
	secretKey  string
	httpClient *http.Client
}

// New creates a YooKassa client. shopID/secretKey come from the merchant's
// personal cabinet (Настройки → API ключи). If either is empty, Enabled()
// reports false and callers should refuse to start a checkout.
func New(shopID, secretKey string) *Client {
	return &Client{
		shopID:    strings.TrimSpace(shopID),
		secretKey: strings.TrimSpace(secretKey),
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Enabled reports whether credentials are configured.
func (c *Client) Enabled() bool {
	return c != nil && c.shopID != "" && c.secretKey != ""
}

// Amount is a monetary value as required by the YooKassa API (decimal string).
type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// RubAmount formats a ruble amount the way YooKassa expects ("129.00").
func RubAmount(rub float64) Amount {
	return Amount{Value: fmt.Sprintf("%.2f", rub), Currency: "RUB"}
}

// Confirmation describes how the buyer confirms the payment.
type Confirmation struct {
	Type            string `json:"type"`
	ReturnURL       string `json:"return_url,omitempty"`
	ConfirmationURL string `json:"confirmation_url,omitempty"`
}

// Payment is the subset of the YooKassa payment object we care about.
type Payment struct {
	ID           string            `json:"id"`
	Status       string            `json:"status"`
	Paid         bool              `json:"paid"`
	Amount       Amount            `json:"amount"`
	Confirmation Confirmation      `json:"confirmation"`
	Description  string            `json:"description,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    string            `json:"created_at,omitempty"`
	Test         bool              `json:"test,omitempty"`
}

// CreatePaymentRequest are the inputs for starting a redirect checkout.
type CreatePaymentRequest struct {
	Amount         Amount
	Description    string
	ReturnURL      string
	Metadata       map[string]string
	IdempotenceKey string
}

// CreatePayment starts a "smart payment" (redirect confirmation) and returns
// the created payment object, including Confirmation.ConfirmationURL to send
// the buyer to.
func (c *Client) CreatePayment(ctx context.Context, req CreatePaymentRequest) (Payment, error) {
	body := map[string]any{
		"amount":  req.Amount,
		"capture": true,
		"confirmation": Confirmation{
			Type:      confirmationRedir,
			ReturnURL: req.ReturnURL,
		},
	}
	if req.Description != "" {
		if len(req.Description) > 128 {
			req.Description = req.Description[:128]
		}
		body["description"] = req.Description
	}
	if len(req.Metadata) > 0 {
		body["metadata"] = req.Metadata
	}

	idempotenceKey := strings.TrimSpace(req.IdempotenceKey)
	if idempotenceKey == "" {
		var err error
		idempotenceKey, err = randomHex(16)
		if err != nil {
			return Payment{}, fmt.Errorf("yookassa: generate idempotence key: %w", err)
		}
	}

	var out Payment
	if err := c.do(ctx, http.MethodPost, pathPayments, idempotenceKey, body, &out); err != nil {
		return Payment{}, err
	}
	return out, nil
}

// GetPayment fetches the authoritative payment state from YooKassa. Always
// call this after receiving a webhook notification instead of trusting the
// notification body directly (YooKassa does not sign notification payloads).
func (c *Client) GetPayment(ctx context.Context, paymentID string) (Payment, error) {
	paymentID = strings.TrimSpace(paymentID)
	if paymentID == "" {
		return Payment{}, fmt.Errorf("yookassa: payment id is required")
	}
	var out Payment
	if err := c.do(ctx, http.MethodGet, pathPayments+"/"+paymentID, "", nil, &out); err != nil {
		return Payment{}, err
	}
	return out, nil
}

// Refund is the subset of the YooKassa refund object we care about.
type Refund struct {
	ID        string `json:"id"`
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	Amount    Amount `json:"amount"`
}

// CreateRefund issues a full or partial refund for a succeeded payment.
func (c *Client) CreateRefund(ctx context.Context, paymentID string, amount Amount, idempotenceKey string) (Refund, error) {
	paymentID = strings.TrimSpace(paymentID)
	if paymentID == "" {
		return Refund{}, fmt.Errorf("yookassa: payment id is required")
	}
	idempotenceKey = strings.TrimSpace(idempotenceKey)
	if idempotenceKey == "" {
		var err error
		idempotenceKey, err = randomHex(16)
		if err != nil {
			return Refund{}, fmt.Errorf("yookassa: generate idempotence key: %w", err)
		}
	}

	body := map[string]any{
		"payment_id": paymentID,
		"amount":     amount,
	}

	var out Refund
	if err := c.do(ctx, http.MethodPost, pathRefunds, idempotenceKey, body, &out); err != nil {
		return Refund{}, err
	}
	return out, nil
}

// APIError is returned when YooKassa responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("yookassa: request failed with status %d: %s", e.StatusCode, e.Body)
}

func (c *Client) do(ctx context.Context, method, path, idempotenceKey string, body any, out any) error {
	if !c.Enabled() {
		return fmt.Errorf("yookassa: client is not configured (missing shop id / secret key)")
	}

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("yookassa: encode request: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("yookassa: build request: %w", err)
	}
	req.SetBasicAuth(c.shopID, c.secretKey)
	req.Header.Set("Content-Type", "application/json")
	if idempotenceKey != "" {
		req.Header.Set("Idempotence-Key", idempotenceKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("yookassa: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("yookassa: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("yookassa: decode response: %w", err)
		}
	}
	return nil
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
