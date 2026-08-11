package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

// fakeStore implements Store in memory with the same semantics the SQL
// enforces: single-use vouchers, event dedup, server-computed amounts.
type fakeStore struct {
	vouchers map[string]*Voucher
	orders   map[string]*Order
	events   map[string]bool
	timing   string
	emails   map[string]string
	seq      int
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		vouchers: map[string]*Voucher{},
		orders:   map[string]*Order{},
		events:   map[string]bool{},
		timing:   TimingAfterApproval,
		emails:   map[string]string{},
	}
}

func (f *fakeStore) CreateVoucher(_ context.Context, pence int, name string, expires time.Time) (Voucher, error) {
	f.seq++
	v := Voucher{ID: fmt.Sprintf("v%d", f.seq), Code: fmt.Sprintf("WD-TEST-%d", f.seq), DropsPriceToPence: pence, ExpiresAt: expires, CreatedAt: time.Now()}
	if name != "" {
		v.RecipientName = &name
	}
	f.vouchers[v.Code] = &v
	return v, nil
}

func (f *fakeStore) ListVouchers(context.Context) ([]Voucher, error) { return nil, nil }

func (f *fakeStore) CreateOrder(_ context.Context, clientID, code string) (Order, error) {
	amount := ListPricePence
	var voucherID *string
	if code != "" {
		v, ok := f.vouchers[code]
		if !ok || v.RedeemedAt != nil || !v.ExpiresAt.After(time.Now()) {
			return Order{}, ErrVoucherInvalid
		}
		now := time.Now()
		v.RedeemedAt = &now
		v.RedeemedByClientID = &clientID
		amount = v.DropsPriceToPence
		voucherID = &v.ID
	}
	f.seq++
	o := Order{ID: fmt.Sprintf("o%d", f.seq), ClientID: clientID, Kind: "site_build", AmountPence: amount, VoucherID: voucherID, Status: OrderCreated, Provider: "stripe", CreatedAt: time.Now()}
	f.orders[o.ID] = &o
	return o, nil
}

func (f *fakeStore) SetOrderSession(_ context.Context, orderID, sessionID string) error {
	if o, ok := f.orders[orderID]; ok {
		o.ProviderSessionID = &sessionID
	}
	return nil
}

func (f *fakeStore) ListOrders(context.Context, int) ([]Order, error) { return nil, nil }

func (f *fakeStore) ProcessEvent(_ context.Context, provider string, ev WebhookEvent) (bool, error) {
	key := provider + "/" + ev.EventID
	if f.events[key] {
		return false, nil
	}
	f.events[key] = true
	if ev.Paid {
		if o, ok := f.orders[ev.OrderID]; ok && o.Status != OrderPaid {
			now := time.Now()
			o.Status = OrderPaid
			o.PaidAt = &now
		}
	}
	return true, nil
}

func (f *fakeStore) GetClientEmail(_ context.Context, clientID string) (string, error) {
	return f.emails[clientID], nil
}
func (f *fakeStore) GetPaymentTiming(context.Context) (string, error) { return f.timing, nil }
func (f *fakeStore) SetPaymentTiming(_ context.Context, t string) error {
	f.timing = t
	return nil
}

func newTestService(store Store, provider Provider) *Service {
	return NewService(store, provider, zap.NewNop())
}

func TestCreateOrderListPrice(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(store, &FakeProvider{PublicBaseURL: "https://x"})
	order, url, err := svc.CreateOrder(context.Background(), "c1", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if order.AmountPence != 14900 {
		t.Fatalf("list price must be 14900p (£149 ruling), got %d", order.AmountPence)
	}
	if url == "" {
		t.Fatal("expected a checkout url")
	}
}

func TestVoucherDropsPriceAndIsSingleUse(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(store, &FakeProvider{})
	v, err := svc.CreateVoucher(context.Background(), 1000, "Alex", time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	order, _, err := svc.CreateOrder(context.Background(), "c1", v.Code, "")
	if err != nil {
		t.Fatal(err)
	}
	if order.AmountPence != 1000 {
		t.Fatalf("voucher must drop price to 1000p, got %d", order.AmountPence)
	}
	// second use must fail — single-use is the ruled invariant
	if _, _, err := svc.CreateOrder(context.Background(), "c2", v.Code, ""); !errors.Is(err, ErrVoucherInvalid) {
		t.Fatalf("second redemption must fail with ErrVoucherInvalid, got %v", err)
	}
}

func TestVoucherRuledVariantsOnly(t *testing.T) {
	svc := newTestService(newFakeStore(), &FakeProvider{})
	for _, pence := range []int{0, 100, 14900, 5501} {
		if _, err := svc.CreateVoucher(context.Background(), pence, "", time.Now().Add(time.Hour)); err == nil {
			t.Fatalf("voucher variant %dp must be rejected (ruled: 1000|5500)", pence)
		}
	}
	for _, pence := range []int{1000, 5500} {
		if _, err := svc.CreateVoucher(context.Background(), pence, "", time.Now().Add(time.Hour)); err != nil {
			t.Fatalf("ruled variant %dp rejected: %v", pence, err)
		}
	}
}

func TestExpiredVoucherRejected(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(store, &FakeProvider{})
	// build an expired voucher directly in the store (service refuses to create one)
	expired, _ := store.CreateVoucher(context.Background(), 1000, "", time.Now().Add(-time.Hour))
	if _, _, err := svc.CreateOrder(context.Background(), "c1", expired.Code, ""); !errors.Is(err, ErrVoucherInvalid) {
		t.Fatalf("expired voucher must be rejected, got %v", err)
	}
}

func TestCheckoutFailureKeepsVoucherConsumed(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(store, &FakeProvider{FailCheckout: errors.New("stripe down")})
	v, _ := svc.CreateVoucher(context.Background(), 5500, "", time.Now().Add(time.Hour))
	_, _, err := svc.CreateOrder(context.Background(), "c1", v.Code, "")
	if err == nil {
		t.Fatal("expected checkout failure to surface")
	}
	// The documented bias: the voucher stays consumed, attached to the order.
	if store.vouchers[v.Code].RedeemedAt == nil {
		t.Fatal("voucher must remain consumed after checkout failure")
	}
}

func TestWebhookMarksPaidOnceAndDedups(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(store, &FakeProvider{})
	order, _, err := svc.CreateOrder(context.Background(), "c1", "", "")
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]any{"event_id": "evt_1", "order_id": order.ID, "paid": true})
	if err := svc.HandleWebhook(context.Background(), payload, ""); err != nil {
		t.Fatal(err)
	}
	if store.orders[order.ID].Status != OrderPaid {
		t.Fatal("order must be paid after verified webhook")
	}
	firstPaidAt := *store.orders[order.ID].PaidAt
	// duplicate delivery: no error, no double effect
	if err := svc.HandleWebhook(context.Background(), payload, ""); err != nil {
		t.Fatal(err)
	}
	if !store.orders[order.ID].PaidAt.Equal(firstPaidAt) {
		t.Fatal("duplicate webhook must not touch the order")
	}
}

func TestUnconfiguredProviderRefuses(t *testing.T) {
	svc := newTestService(newFakeStore(), nil)
	if _, _, err := svc.CreateOrder(context.Background(), "c1", "", ""); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
	if err := svc.HandleWebhook(context.Background(), []byte("{}"), ""); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestPaymentTimingSwitchVocabulary(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(store, nil)
	if err := svc.SetPaymentTiming(context.Background(), "upfront"); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetPaymentTiming(context.Background(), "on_delivery"); err == nil {
		t.Fatal("unruled timing value must be rejected")
	}
	got, _ := svc.GetPaymentTiming(context.Background())
	if got != TimingUpfront {
		t.Fatalf("timing = %q, want upfront", got)
	}
}

// --- Stripe signature verification (the webhook's whole authentication) ---

func signStripe(payload []byte, secret, ts string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(payload)
	return "t=" + ts + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestStripeSignatureVerification(t *testing.T) {
	p := NewStripeProvider("sk_test", "whsec_test", "https://webdesign.uk")
	payload := []byte(`{"id":"evt_9","type":"checkout.session.completed","data":{"object":{"payment_status":"paid","customer":"cus_1","metadata":{"order_id":"abc"}}}}`)

	ev, err := p.ParseWebhook(payload, signStripe(payload, "whsec_test", "1700000000"))
	if err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	if !ev.Paid || ev.OrderID != "abc" || ev.CustomerID != "cus_1" {
		t.Fatalf("event misparsed: %+v", ev)
	}

	if _, err := p.ParseWebhook(payload, signStripe(payload, "wrong_secret", "1700000000")); err == nil {
		t.Fatal("wrong-secret signature must be rejected")
	}
	tampered := append([]byte{}, payload...)
	tampered[len(tampered)-2] = 'X'
	if _, err := p.ParseWebhook(tampered, signStripe(payload, "whsec_test", "1700000000")); err == nil {
		t.Fatal("tampered payload must be rejected")
	}
	if _, err := p.ParseWebhook(payload, ""); err == nil {
		t.Fatal("missing signature header must be rejected")
	}
}

func TestStripeRefusesNonPositiveAmount(t *testing.T) {
	p := NewStripeProvider("sk_test", "whsec_test", "https://webdesign.uk")
	for _, pence := range []int{0, -100} {
		if _, _, err := p.CreateCheckout("o1", "a@b.c", pence, "x"); err == nil {
			t.Fatalf("amount %dp must be refused, never sent to Stripe", pence)
		}
	}
}
