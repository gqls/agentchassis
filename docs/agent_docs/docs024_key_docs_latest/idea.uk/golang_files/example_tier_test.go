package main

// example_tier_test.go — the £8 example place.
//
// A full report at a lower price in exchange for permission to publish it
// anonymously. This touches money and it records a promise made to a person, so
// the tests below are about the ways it could go wrong quietly: billing the wrong
// amount, or holding a consent nobody gave.

import (
	"net/url"
	"strings"
	"testing"
)

func tierApp(t *testing.T, price, places int) (*App, *[][3]string) {
	t.Helper()
	app, sent := newTestApp(func(c *Config) {
		c.PriceGBP = 29
		c.ExamplePriceGBP = price
		c.ExampleMaxPlaces = places
	})
	return app, sent
}

func exampleForm() url.Values {
	f := validForm()
	f.Set("tier", "example")
	f.Set("publish_consent", "yes")
	return f
}

func onlyOrder(t *testing.T, app *App) *Order {
	t.Helper()
	if n := len(app.store.Orders); n != 1 {
		t.Fatalf("want exactly 1 order, got %d", n)
	}
	for _, o := range app.store.Orders {
		return o
	}
	return nil
}

// The tier ships OFF. A binary rolling with the code but without the config must
// never start selling £8 reports on its own.
func TestExampleTierIsOffByDefault(t *testing.T) {
	app, _ := newTestApp(func(c *Config) { c.PriceGBP = 29 })
	postForm(t, app, exampleForm())
	o := onlyOrder(t, app)
	if o.PriceGBP != 29 || o.PublishConsent {
		t.Errorf("tier active with no config: price=%d consent=%v", o.PriceGBP, o.PublishConsent)
	}
}

// Asking for the tier AND ticking the box gets the lower price, and the consent
// is recorded on the order rather than inferred later from the price.
func TestExamplePlaceSetsPriceAndRecordsConsent(t *testing.T) {
	app, _ := tierApp(t, 8, 10)
	postForm(t, app, exampleForm())
	o := onlyOrder(t, app)
	if o.PriceGBP != 8 {
		t.Errorf("price = %d, want 8", o.PriceGBP)
	}
	if !o.PublishConsent {
		t.Error("consent not recorded on the order")
	}
}

// The cheap price and the permission are ONE decision. Granting the price without
// the tick would leave us billing £8 for something we have no right to publish.
func TestExamplePriceRequiresTheConsentBox(t *testing.T) {
	app, _ := tierApp(t, 8, 10)
	f := exampleForm()
	f.Del("publish_consent") // asked for the tier, did not tick
	postForm(t, app, f)
	o := onlyOrder(t, app)
	if o.PriceGBP != 29 {
		t.Errorf("price = %d — the discount was given without consent", o.PriceGBP)
	}
	if o.PublishConsent {
		t.Error("consent recorded when the box was not ticked")
	}
}

// And the mirror: a tick without asking for the tier must not silently opt someone
// into publication at full price.
func TestConsentWithoutTierDoesNotRecordConsent(t *testing.T) {
	app, _ := tierApp(t, 8, 10)
	f := validForm()
	f.Set("publish_consent", "yes")
	postForm(t, app, f)
	o := onlyOrder(t, app)
	if o.PublishConsent {
		t.Error("consent recorded for a standard order — nobody agreed to publication")
	}
	if o.PriceGBP != 29 {
		t.Errorf("price = %d, want 29", o.PriceGBP)
	}
}

// The cap bounds how many people we have made a publication promise to. Past it,
// the request still succeeds — at the standard price.
func TestExamplePlacesRunOut(t *testing.T) {
	app, _ := tierApp(t, 8, 1)
	postForm(t, app, exampleForm()) // takes the only place
	postForm(t, app, exampleForm()) // one too many

	var cheap, full int
	for _, o := range app.store.Orders {
		if o.PriceGBP == 8 {
			cheap++
		} else {
			full++
		}
	}
	if cheap != 1 || full != 1 {
		t.Errorf("cap not enforced: %d cheap, %d full (want 1 and 1)", cheap, full)
	}
}

// The counter must include orders we later declined or expired: the cap exists to
// bound the promises made, and declining to run a report does not un-ask.
func TestConsentedCountIncludesTerminalOrders(t *testing.T) {
	s, _ := NewStore("")
	s.Save(&Order{ID: "a", PublishConsent: true, Status: "declined"})
	s.Save(&Order{ID: "b", PublishConsent: true, Status: "expired"})
	s.Save(&Order{ID: "c", Status: "delivered"})
	if got := s.ConsentedCount(); got != 2 {
		t.Errorf("ConsentedCount = %d, want 2 (declined and expired still count)", got)
	}
}

// The amount charged comes from the ORDER, not from config read at checkout time —
// otherwise an order taken at £8 could be billed at whatever the process happens
// to be configured for when the operator gets round to approving it.
func TestPayLinkBillsTheOrdersOwnPrice(t *testing.T) {
	app, sent := tierApp(t, 8, 10)
	o := &Order{ID: "ord_x", Email: "b@example.com", Name: "Ada", Domain: "an idea", PriceGBP: 8}
	app.store.Save(o)
	if _, err := app.sendPayLink(o); err != nil {
		t.Fatalf("sendPayLink: %v", err)
	}
	var body string
	for _, m := range *sent {
		if strings.Contains(m[1], "ready to pay") {
			body = m[2]
		}
	}
	if body == "" {
		t.Fatal("no pay-link email sent")
	}
	if !strings.Contains(body, "£8") {
		t.Errorf("pay-link email does not quote £8:\n%.300s", body)
	}
	if strings.Contains(body, "£29") {
		t.Errorf("pay-link email quotes the standard price for an £8 order:\n%.300s", body)
	}
}

// A legacy row written before tiers existed has PriceGBP 0 and must fall back to
// the standard price — never to £0.
func TestLegacyOrderFallsBackToStandardPrice(t *testing.T) {
	o := &Order{ID: "old"}
	if got := o.Price(29); got != 29 {
		t.Errorf("legacy order price = %d, want 29", got)
	}
	o2 := &Order{ID: "new", PriceGBP: 8}
	if got := o2.Price(29); got != 8 {
		t.Errorf("tier order price = %d, want 8", got)
	}
}

// Stripe would accept a £0 line item and charge nothing. Refusing loudly beats
// creating a checkout whose amount was never decided.
func TestCheckoutRefusesANonPositivePrice(t *testing.T) {
	p := &StripeProvider{publicBaseURL: "http://test"}
	for _, price := range []int{0, -5} {
		if _, _, err := p.CreateCheckout("ord_x", "b@example.com", price); err == nil {
			t.Errorf("CreateCheckout(price=%d) returned no error", price)
		}
	}
}
