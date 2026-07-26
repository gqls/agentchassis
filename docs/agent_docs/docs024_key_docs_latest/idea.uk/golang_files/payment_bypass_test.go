package main

// payment_bypass_test.go — the `fake=1` shortcut on /order/success must be
// available ONLY when FakeProvider is configured.
//
// Why this test exists: /order/success IS the Stripe success_url, and the
// matching cancel_url hands the order id to the buyer. With the shortcut
// ungated, a buyer who started a real checkout and cancelled could re-enter
// /order/success?o=<their id>&fake=1 and be delivered a £29 report unpaid.
// Found live on the box 2026-07-26 (bugs_open/085).

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestAppWithProvider mirrors newTestApp but lets the caller pick the
// provider, so the Stripe (production) arm can be exercised without touching
// Stripe — nothing here calls CreateCheckout.
func newTestAppWithProvider(p Provider, tweaks ...func(*Config)) (*App, *[][3]string) {
	cfg := Config{
		PriceGBP: 29, AutoDeliver: true, ReviewBeforePay: true, PublicBaseURL: "http://test",
		InternalAPIKey: "testkey", OperatorEmail: "ops@test", MaxActive: 2,
		AllowedOrigins: []string{"*"},
	}
	for _, tw := range tweaks {
		tw(&cfg)
	}
	store, _ := NewStore("")
	app := NewApp(cfg, store, p)
	app.engine = func(d, a, s string) (renderedReport, error) {
		return renderedReport{Text: "# Stub report for " + d, HTML: "<h1>Stub</h1>"}, nil
	}
	sent := &[][3]string{}
	app.deliver = func(to, subj, body string) { *sent = append(*sent, [3]string{to, subj, body}) }
	app.deliverHTML = func(to, subj, text, htmlBody string) { *sent = append(*sent, [3]string{to, subj, text}) }
	app.dispatch = func(f func()) { f() }
	return app, sent
}

// seedAwaitingPayment puts one order in the exact state the bypass targets:
// approved, pay link sent, buyer has NOT paid. Its report is already stored
// (review-before-pay), so a successful bypass delivers immediately.
func seedAwaitingPayment(app *App) string {
	id := "ord_bypass_test"
	app.store.Save(&Order{
		ID: id, Name: "Test", Email: "buyer@test", Domain: "a bakery",
		Audience: "local", Assets: "none", Status: "awaiting_payment",
		Report: "# The report", ReportHTML: "<h1>The report</h1>",
	})
	return id
}

func TestFakePaymentShortcutRefusedUnderStripe(t *testing.T) {
	app, sent := newTestAppWithProvider(&StripeProvider{
		secretKey: "sk_test_notused", webhookSecret: "whsec_notused",
		publicBaseURL: "http://test", priceGBP: 29,
	})
	srv := httptest.NewServer(app.routes())
	defer srv.Close()
	id := seedAwaitingPayment(app)

	resp, err := srv.Client().Get(srv.URL + "/order/success?o=" + id + "&fake=1")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	o, ok := app.store.Get(id)
	if !ok {
		t.Fatal("order vanished")
	}
	if o.Status != "awaiting_payment" {
		t.Fatalf("PAYMENT BYPASS: status moved to %q under StripeProvider; want it held at awaiting_payment", o.Status)
	}
	for _, s := range *sent {
		if s[0] == "buyer@test" {
			t.Fatalf("PAYMENT BYPASS: report emailed to the buyer unpaid (subject %q)", s[1])
		}
	}
}

// The legitimate local-test path must still work, or the shortcut's only
// purpose is gone.
func TestFakePaymentShortcutStillWorksUnderFakeProvider(t *testing.T) {
	app, sent := newTestAppWithProvider(&FakeProvider{publicBaseURL: "http://test"})
	srv := httptest.NewServer(app.routes())
	defer srv.Close()
	id := seedAwaitingPayment(app)

	resp, err := srv.Client().Get(srv.URL + "/order/success?o=" + id + "&fake=1")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	o, _ := app.store.Get(id)
	if o.Status != "delivered" {
		t.Fatalf("FakeProvider shortcut broken: status %q, want delivered", o.Status)
	}
	if !anySent(sent, "buyer@test", "The report") {
		t.Fatal("FakeProvider shortcut broken: report not delivered to the buyer")
	}
}

// The page itself is harmless to serve (Stripe redirects real buyers here), so
// it must still answer 200 rather than erroring on the refused shortcut.
func TestOrderSuccessPageStillRendersUnderStripe(t *testing.T) {
	app, _ := newTestAppWithProvider(&StripeProvider{publicBaseURL: "http://test", priceGBP: 29})
	srv := httptest.NewServer(app.routes())
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/order/success?o=ord_unknown&fake=1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200", resp.StatusCode)
	}
}
