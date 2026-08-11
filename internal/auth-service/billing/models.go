package billing

import "time"

// Voucher mirrors the vouchers table (clients_db, migration 391).
type Voucher struct {
	ID                 string     `json:"id"`
	Code               string     `json:"code"`
	DropsPriceToPence  int        `json:"drops_price_to_pence"`
	RecipientName      *string    `json:"recipient_name,omitempty"`
	ExpiresAt          time.Time  `json:"expires_at"`
	RedeemedAt         *time.Time `json:"redeemed_at,omitempty"`
	RedeemedByClientID *string    `json:"redeemed_by_client_id,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// Order mirrors the billing_orders table (clients_db, migration 391).
type Order struct {
	ID                 string     `json:"id"`
	ClientID           string     `json:"client_id"`
	Kind               string     `json:"kind"`
	AmountPence        int        `json:"amount_pence"`
	VoucherID          *string    `json:"voucher_id,omitempty"`
	Status             string     `json:"status"`
	Provider           string     `json:"provider"`
	ProviderSessionID  *string    `json:"provider_session_id,omitempty"`
	ProviderCustomerID *string    `json:"provider_customer_id,omitempty"`
	PaidAt             *time.Time `json:"paid_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// ListPricePence is the ruled £149 all-in list price (owner, 2026-08-11).
const ListPricePence = 14900

// The ruled voucher variants: codes drop the all-in price to £10 or £55.
var RuledVoucherPences = map[int]bool{1000: true, 5500: true}

// Order statuses. Deliberately no 'refunded' — refunds are manual and
// unadvertised (owner ruling 2026-08-11); code must not model them.
const (
	OrderCreated = "created"
	OrderPaid    = "paid"
)

// Payment timing switch values (billing_settings.payment_timing).
const (
	TimingAfterApproval = "after_approval"
	TimingUpfront       = "upfront"
)
