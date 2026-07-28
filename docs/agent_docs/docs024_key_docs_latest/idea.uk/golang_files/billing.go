package main

// billing.go — payment provider behind an interface (webhook = source of truth).
// StripeProvider talks to Stripe over net/http and verifies webhook signatures
// with crypto/hmac (no SDK). FakeProvider is for local end-to-end testing.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type WebhookEvent struct {
	EventID string
	Type    string
	OrderID string // our order id, carried in metadata
	Paid    bool
}

type Provider interface {
	// priceGBP is passed per CALL, not held on the provider: the amount is a
	// property of the order, not of the process. Holding it on the provider is
	// what made a second tier impossible without this change.
	CreateCheckout(orderID, email string, priceGBP int) (sessionID, checkoutURL string, err error)
	ParseWebhook(payload []byte, sigHeader string) (WebhookEvent, error)
}

// ── Stripe ───────────────────────────────────────────────────────────────────
type StripeProvider struct {
	secretKey     string
	webhookSecret string
	publicBaseURL string
}

func (p *StripeProvider) CreateCheckout(orderID, email string, priceGBP int) (string, string, error) {
	// Refuse rather than charge a wrong amount. A zero here would bill £0 and a
	// negative would be rejected by Stripe with a far less obvious error; either
	// way the buyer's charge must never be inferred from a missing value.
	if priceGBP <= 0 {
		return "", "", fmt.Errorf("refusing to create a checkout for £%d — price must be positive", priceGBP)
	}
	form := url.Values{}
	form.Set("mode", "payment")
	if email != "" {
		form.Set("customer_email", email)
	}
	form.Set("line_items[0][quantity]", "1")
	form.Set("line_items[0][price_data][currency]", "gbp")
	form.Set("line_items[0][price_data][unit_amount]", strconv.Itoa(priceGBP*100))
	form.Set("line_items[0][price_data][product_data][name]", "idea.uk — verified AI opportunity report")
	form.Set("line_items[0][price_data][product_data][description]", "Ranked, web-verified candidate ideas for your business.")
	form.Set("metadata[order_id]", orderID)
	form.Set("success_url", fmt.Sprintf("%s/order/success?o=%s", p.publicBaseURL, orderID))
	form.Set("cancel_url", fmt.Sprintf("%s/order/cancel?o=%s", p.publicBaseURL, orderID))

	req, _ := http.NewRequest("POST", "https://api.stripe.com/v1/checkout/sessions",
		strings.NewReader(form.Encode()))
	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("stripe %d: %s", resp.StatusCode, b)
	}
	var out struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", "", err
	}
	return out.ID, out.URL, nil
}

func (p *StripeProvider) ParseWebhook(payload []byte, sigHeader string) (WebhookEvent, error) {
	if err := verifyStripeSignature(payload, sigHeader, p.webhookSecret); err != nil {
		return WebhookEvent{}, err
	}
	var e struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object struct {
				PaymentStatus string            `json:"payment_status"`
				Metadata      map[string]string `json:"metadata"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &e); err != nil {
		return WebhookEvent{}, err
	}
	paid := e.Type == "checkout.session.completed" && e.Data.Object.PaymentStatus == "paid"
	return WebhookEvent{e.ID, e.Type, e.Data.Object.Metadata["order_id"], paid}, nil
}

// verifyStripeSignature checks the Stripe-Signature header: t=timestamp,v1=hmac.
// expected = hex(HMAC-SHA256(secret, "timestamp.payload")).
func verifyStripeSignature(payload []byte, sigHeader, secret string) error {
	var ts, v1 string
	for _, part := range strings.Split(sigHeader, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			ts = kv[1]
		case "v1":
			v1 = kv[1]
		}
	}
	if ts == "" || v1 == "" {
		return errors.New("stripe signature: missing t or v1")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(v1)) {
		return errors.New("stripe signature: mismatch")
	}
	return nil
}

// ── Fake (local/testing only — NEVER in production) ─────────────────────────
type FakeProvider struct{ publicBaseURL string }

func (p *FakeProvider) CreateCheckout(orderID, email string, priceGBP int) (string, string, error) {
	return "fake_" + orderID,
		fmt.Sprintf("%s/order/success?o=%s&fake=1", p.publicBaseURL, orderID), nil
}

func (p *FakeProvider) ParseWebhook(payload []byte, _ string) (WebhookEvent, error) {
	var b struct {
		EventID string `json:"event_id"`
		Type    string `json:"type"`
		OrderID string `json:"order_id"`
		Paid    bool   `json:"paid"`
	}
	if err := json.Unmarshal(payload, &b); err != nil {
		return WebhookEvent{}, err
	}
	if b.EventID == "" {
		b.EventID = "evt_fake"
	}
	if b.Type == "" {
		b.Type = "checkout.session.completed"
	}
	return WebhookEvent{b.EventID, b.Type, b.OrderID, b.Paid}, nil
}
