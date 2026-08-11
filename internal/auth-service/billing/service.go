package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ErrNotConfigured is returned while no payment provider is wired (no Stripe
// keys in the environment). The service is deliberately buildable and
// deployable in this state — owner ruling 2026-08-11: set it up even if
// unused at first.
var ErrNotConfigured = errors.New("billing provider not configured")

const productDescription = "webdesign.uk — AI-built website, all-in"

type Service struct {
	store    Store
	provider Provider // nil until Stripe keys are supplied
	logger   *zap.Logger
}

func NewService(store Store, provider Provider, logger *zap.Logger) *Service {
	return &Service{store: store, provider: provider, logger: logger}
}

func (s *Service) Configured() bool { return s.provider != nil }

// CreateVoucher enforces the ruled variants: codes drop the £149 price to
// £10 or £55 (owner ruling 2026-08-11). Other amounts need a new ruling, not
// a wider API.
func (s *Service) CreateVoucher(ctx context.Context, dropsPriceToPence int, recipientName string, expiresAt time.Time) (Voucher, error) {
	if !RuledVoucherPences[dropsPriceToPence] {
		return Voucher{}, fmt.Errorf("drops_price_to_pence must be 1000 (£10) or 5500 (£55), got %d", dropsPriceToPence)
	}
	if !expiresAt.After(time.Now()) {
		return Voucher{}, errors.New("expires_at must be in the future")
	}
	return s.store.CreateVoucher(ctx, dropsPriceToPence, recipientName, expiresAt)
}

func (s *Service) ListVouchers(ctx context.Context) ([]Voucher, error) {
	return s.store.ListVouchers(ctx)
}

// CreateOrder creates the order row (redeeming a voucher atomically when a
// code is given), then asks the provider for a checkout session. The amount
// is always computed server-side. Returns the order and the checkout URL to
// hand to the customer.
//
// If checkout creation fails after a voucher was redeemed, the voucher stays
// consumed and attached to the 'created' order — retrying payment for that
// order is the recovery, not re-typing the code. That bias is deliberate: a
// code that could be re-redeemed on provider flakiness is a double-discount.
func (s *Service) CreateOrder(ctx context.Context, clientID, voucherCode, emailOverride string) (Order, string, error) {
	if s.provider == nil {
		return Order{}, "", ErrNotConfigured
	}
	email := emailOverride
	if email == "" {
		var err error
		email, err = s.store.GetClientEmail(ctx, clientID)
		if err != nil {
			return Order{}, "", err
		}
	}
	order, err := s.store.CreateOrder(ctx, clientID, voucherCode)
	if err != nil {
		return Order{}, "", err
	}
	sessionID, url, err := s.provider.CreateCheckout(order.ID, email, order.AmountPence, productDescription)
	if err != nil {
		s.logger.Error("checkout creation failed; order remains 'created'",
			zap.String("order_id", order.ID), zap.Error(err))
		return order, "", fmt.Errorf("order %s created but checkout failed: %w", order.ID, err)
	}
	if err := s.store.SetOrderSession(ctx, order.ID, sessionID); err != nil {
		// The session exists at the provider; losing the linkage is
		// recoverable via the webhook's metadata order_id. Log, don't fail.
		s.logger.Error("failed to store checkout session id", zap.String("order_id", order.ID), zap.Error(err))
	}
	return order, url, nil
}

func (s *Service) ListOrders(ctx context.Context, limit int) ([]Order, error) {
	return s.store.ListOrders(ctx, limit)
}

// HandleWebhook verifies, dedups and applies one webhook delivery. The
// signature check inside ParseWebhook is the authentication; an error there
// means the payload is untrusted and nothing happens.
func (s *Service) HandleWebhook(ctx context.Context, payload []byte, sigHeader string) error {
	if s.provider == nil {
		return ErrNotConfigured
	}
	ev, err := s.provider.ParseWebhook(payload, sigHeader)
	if err != nil {
		return err
	}
	fresh, err := s.store.ProcessEvent(ctx, "stripe", ev)
	if err != nil {
		return err
	}
	if !fresh {
		s.logger.Info("duplicate webhook delivery ignored", zap.String("event_id", ev.EventID))
		return nil
	}
	if ev.Paid {
		s.logger.Info("order paid", zap.String("order_id", ev.OrderID), zap.String("event_id", ev.EventID))
	}
	return nil
}

func (s *Service) GetPaymentTiming(ctx context.Context) (string, error) {
	return s.store.GetPaymentTiming(ctx)
}

// SetPaymentTiming flips the ruled switch: payment after approval while the
// system is being tested, up-front later (owner ruling 2026-08-11).
func (s *Service) SetPaymentTiming(ctx context.Context, timing string) error {
	if timing != TimingAfterApproval && timing != TimingUpfront {
		return fmt.Errorf("payment_timing must be %q or %q", TimingAfterApproval, TimingUpfront)
	}
	return s.store.SetPaymentTiming(ctx, timing)
}
