package billing

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
	"time"
)

// StripeProvider talks to Stripe over net/http and verifies webhook
// signatures with crypto/hmac. No SDK: the two calls we make (create a
// payment-mode checkout session, verify a webhook) are one form POST and one
// HMAC, and the estate already runs this shape in production (idea.uk).
type StripeProvider struct {
	secretKey     string
	webhookSecret string
	publicBaseURL string
	httpClient    *http.Client
}

func NewStripeProvider(secretKey, webhookSecret, publicBaseURL string) *StripeProvider {
	return &StripeProvider{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		httpClient:    &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *StripeProvider) CreateCheckout(orderID, email string, amountPence int, description string) (string, string, error) {
	// Refuse rather than charge a wrong amount: a zero would bill £0 and a
	// negative draws an unhelpful Stripe error; the buyer's charge must never
	// be inferred from a missing value.
	if amountPence <= 0 {
		return "", "", fmt.Errorf("refusing to create a checkout for %dp — amount must be positive", amountPence)
	}
	form := url.Values{}
	form.Set("mode", "payment")
	if email != "" {
		form.Set("customer_email", email)
	}
	form.Set("line_items[0][quantity]", "1")
	form.Set("line_items[0][price_data][currency]", "gbp")
	form.Set("line_items[0][price_data][unit_amount]", strconv.Itoa(amountPence))
	form.Set("line_items[0][price_data][product_data][name]", description)
	form.Set("metadata[order_id]", orderID)
	form.Set("success_url", fmt.Sprintf("%s/pay/success?o=%s", p.publicBaseURL, orderID))
	form.Set("cancel_url", fmt.Sprintf("%s/pay/cancel?o=%s", p.publicBaseURL, orderID))

	req, err := http.NewRequest("POST", "https://api.stripe.com/v1/checkout/sessions",
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
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
				Customer      string            `json:"customer"`
				Metadata      map[string]string `json:"metadata"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &e); err != nil {
		return WebhookEvent{}, err
	}
	return WebhookEvent{
		EventID:    e.ID,
		Type:       e.Type,
		OrderID:    e.Data.Object.Metadata["order_id"],
		CustomerID: e.Data.Object.Customer,
		Paid:       e.Type == "checkout.session.completed" && e.Data.Object.PaymentStatus == "paid",
		Raw:        payload,
	}, nil
}

// verifyStripeSignature checks the Stripe-Signature header (t=timestamp,
// v1=hmac): expected = hex(HMAC-SHA256(secret, "timestamp.payload")).
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
