package billing

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ErrVoucherInvalid covers unknown, already-redeemed and expired codes with
// one message on purpose: the caller cannot distinguish them without probing
// other people's codes, and the atomic redemption below cannot either.
var ErrVoucherInvalid = errors.New("voucher code is invalid, already used, or expired")

// Store is what the service needs from persistence; Repository implements it
// against clients_db, tests implement it in memory.
type Store interface {
	CreateVoucher(ctx context.Context, dropsPriceToPence int, recipientName string, expiresAt time.Time) (Voucher, error)
	ListVouchers(ctx context.Context) ([]Voucher, error)
	// CreateOrder inserts an order, atomically redeeming voucherCode for
	// clientID in the same transaction when it is non-empty. The amount is
	// computed here — list price, or the voucher's — never caller-supplied.
	CreateOrder(ctx context.Context, clientID, voucherCode string) (Order, error)
	SetOrderSession(ctx context.Context, orderID, sessionID string) error
	ListOrders(ctx context.Context, limit int) ([]Order, error)
	// ProcessEvent records the webhook event (dedup on provider+event id) and,
	// for a paid event, marks the order paid and links the provider customer
	// id to clients.external_id — one transaction. Returns false when the
	// event was a duplicate and nothing was done.
	ProcessEvent(ctx context.Context, provider string, ev WebhookEvent) (bool, error)
	GetClientEmail(ctx context.Context, clientID string) (string, error)
	GetPaymentTiming(ctx context.Context) (string, error)
	SetPaymentTiming(ctx context.Context, timing string) error
}

type Repository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewRepository(pool *pgxpool.Pool, logger *zap.Logger) *Repository {
	return &Repository{pool: pool, logger: logger}
}

// voucher codes: unambiguous alphabet (no 0/O/1/I), server-generated.
const codeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

func newVoucherCode() (string, error) {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, 0, 13)
	out = append(out, 'W', 'D', '-')
	for i, c := range b {
		if i == 5 {
			out = append(out, '-')
		}
		out = append(out, codeAlphabet[int(c)%len(codeAlphabet)])
	}
	return string(out), nil
}

func (r *Repository) CreateVoucher(ctx context.Context, dropsPriceToPence int, recipientName string, expiresAt time.Time) (Voucher, error) {
	code, err := newVoucherCode()
	if err != nil {
		return Voucher{}, err
	}
	var v Voucher
	err = r.pool.QueryRow(ctx, `
		INSERT INTO vouchers (code, drops_price_to_pence, recipient_name, expires_at)
		VALUES ($1, $2, NULLIF($3, ''), $4)
		RETURNING id, code, drops_price_to_pence, recipient_name, expires_at, redeemed_at, redeemed_by_client_id, created_at`,
		code, dropsPriceToPence, recipientName, expiresAt,
	).Scan(&v.ID, &v.Code, &v.DropsPriceToPence, &v.RecipientName, &v.ExpiresAt, &v.RedeemedAt, &v.RedeemedByClientID, &v.CreatedAt)
	return v, err
}

func (r *Repository) ListVouchers(ctx context.Context) ([]Voucher, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, drops_price_to_pence, recipient_name, expires_at, redeemed_at, redeemed_by_client_id, created_at
		FROM vouchers ORDER BY created_at DESC LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Voucher
	for rows.Next() {
		var v Voucher
		if err := rows.Scan(&v.ID, &v.Code, &v.DropsPriceToPence, &v.RecipientName, &v.ExpiresAt, &v.RedeemedAt, &v.RedeemedByClientID, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *Repository) CreateOrder(ctx context.Context, clientID, voucherCode string) (Order, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback(ctx)

	amount := ListPricePence
	var voucherID *string
	if voucherCode != "" {
		// The single-use invariant IS this statement: only one transaction can
		// move redeemed_at from NULL, so a raced second redemption gets zero
		// rows and reports the code as used.
		var vid string
		err := tx.QueryRow(ctx, `
			UPDATE vouchers
			SET redeemed_at = now(), redeemed_by_client_id = $2
			WHERE code = $1 AND redeemed_at IS NULL AND expires_at > now()
			RETURNING id, drops_price_to_pence`,
			voucherCode, clientID,
		).Scan(&vid, &amount)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return Order{}, ErrVoucherInvalid
			}
			return Order{}, err
		}
		voucherID = &vid
	}

	var o Order
	err = tx.QueryRow(ctx, `
		INSERT INTO billing_orders (client_id, amount_pence, voucher_id)
		VALUES ($1, $2, $3)
		RETURNING id, client_id, kind, amount_pence, voucher_id, status, provider, provider_session_id, provider_customer_id, paid_at, created_at`,
		clientID, amount, voucherID,
	).Scan(&o.ID, &o.ClientID, &o.Kind, &o.AmountPence, &o.VoucherID, &o.Status, &o.Provider, &o.ProviderSessionID, &o.ProviderCustomerID, &o.PaidAt, &o.CreatedAt)
	if err != nil {
		return Order{}, err
	}
	return o, tx.Commit(ctx)
}

func (r *Repository) SetOrderSession(ctx context.Context, orderID, sessionID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE billing_orders SET provider_session_id = $2 WHERE id = $1`,
		orderID, sessionID)
	return err
}

func (r *Repository) ListOrders(ctx context.Context, limit int) ([]Order, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, client_id, kind, amount_pence, voucher_id, status, provider, provider_session_id, provider_customer_id, paid_at, created_at
		FROM billing_orders ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.ClientID, &o.Kind, &o.AmountPence, &o.VoucherID, &o.Status, &o.Provider, &o.ProviderSessionID, &o.ProviderCustomerID, &o.PaidAt, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) ProcessEvent(ctx context.Context, provider string, ev WebhookEvent) (bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	// The PK on (provider, provider_event_id) is the idempotency: a conflict
	// means a previous delivery already did the work below and committed.
	var orderID any
	if ev.OrderID != "" {
		if id, err := uuid.Parse(ev.OrderID); err == nil {
			orderID = id
		}
	}
	tag, err := tx.Exec(ctx, `
		INSERT INTO billing_events (provider, provider_event_id, type, order_id, payload)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider, provider_event_id) DO NOTHING`,
		provider, ev.EventID, ev.Type, orderID, ev.Raw)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		return false, nil // duplicate delivery — already processed
	}

	if ev.Paid && orderID != nil {
		tag, err := tx.Exec(ctx, `
			UPDATE billing_orders
			SET status = 'paid', paid_at = now(), provider_customer_id = NULLIF($2, '')
			WHERE id = $1 AND status <> 'paid'`,
			orderID, ev.CustomerID)
		if err != nil {
			return false, err
		}
		if tag.RowsAffected() == 0 {
			r.logger.Warn("paid webhook matched no unpaid order", zap.String("order_id", ev.OrderID), zap.String("event_id", ev.EventID))
		}
		if ev.CustomerID != "" {
			// Stripe linkage stays on clients.external_id (migration 375's
			// stated intent). First writer wins; never overwrite.
			if _, err := tx.Exec(ctx, `
				UPDATE clients SET external_id = $2
				WHERE id = (SELECT client_id FROM billing_orders WHERE id = $1)
				  AND external_id IS NULL`,
				orderID, ev.CustomerID); err != nil {
				return false, err
			}
		}
	}
	return true, tx.Commit(ctx)
}

func (r *Repository) GetClientEmail(ctx context.Context, clientID string) (string, error) {
	var email *string
	err := r.pool.QueryRow(ctx, `SELECT email FROM clients WHERE id = $1`, clientID).Scan(&email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("client %s not found", clientID)
		}
		return "", err
	}
	if email == nil {
		return "", nil
	}
	return *email, nil
}

func (r *Repository) GetPaymentTiming(ctx context.Context) (string, error) {
	var timing string
	err := r.pool.QueryRow(ctx, `SELECT payment_timing FROM billing_settings WHERE id = 1`).Scan(&timing)
	return timing, err
}

func (r *Repository) SetPaymentTiming(ctx context.Context, timing string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE billing_settings SET payment_timing = $1, updated_at = now() WHERE id = 1`, timing)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("billing_settings row missing — migration 391 not applied")
	}
	return nil
}
