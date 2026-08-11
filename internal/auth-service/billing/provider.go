// Package billing is the payment surface for the £149 site-build product
// (owner rulings 2026-08-11, ai_site_selling_automation PLAN §1b/§1c/§2.7).
//
// The provider sits behind an interface with the webhook as the ONLY source
// of truth — a browser redirect never grants anything. This is the idea.uk
// pattern (proven with real money), reimplemented here against clients_db:
// pattern, not code port.
package billing

import "encoding/json"

// WebhookEvent is the provider-agnostic result of parsing a webhook payload.
type WebhookEvent struct {
	EventID    string
	Type       string
	OrderID    string // our billing_orders.id, carried in checkout metadata
	CustomerID string // provider-side customer id, if the event carries one
	Paid       bool
	Raw        json.RawMessage
}

// Provider creates checkouts and parses webhooks. Amounts are passed per
// call, not held on the provider: the price is a property of the order.
type Provider interface {
	// CreateCheckout returns (sessionID, checkoutURL). amountPence must be
	// positive — implementations refuse rather than risk charging £0.
	CreateCheckout(orderID, email string, amountPence int, description string) (string, string, error)

	// ParseWebhook verifies the signature and normalises the event.
	// An error means the payload is untrusted; nothing may act on it.
	ParseWebhook(payload []byte, sigHeader string) (WebhookEvent, error)
}

// FakeProvider is for tests ONLY. It is never constructed in main.go: the
// only paths to it are test files, so production cannot select it by
// misconfiguration.
type FakeProvider struct {
	PublicBaseURL string
	// FailCheckout, when set, makes CreateCheckout return this error.
	FailCheckout error
}

func (p *FakeProvider) CreateCheckout(orderID, email string, amountPence int, description string) (string, string, error) {
	if p.FailCheckout != nil {
		return "", "", p.FailCheckout
	}
	return "fake_" + orderID, p.PublicBaseURL + "/pay/fake?o=" + orderID, nil
}

func (p *FakeProvider) ParseWebhook(payload []byte, _ string) (WebhookEvent, error) {
	var b struct {
		EventID    string `json:"event_id"`
		Type       string `json:"type"`
		OrderID    string `json:"order_id"`
		CustomerID string `json:"customer_id"`
		Paid       bool   `json:"paid"`
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
	return WebhookEvent{EventID: b.EventID, Type: b.Type, OrderID: b.OrderID, CustomerID: b.CustomerID, Paid: b.Paid, Raw: payload}, nil
}
