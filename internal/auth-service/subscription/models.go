package subscription

import (
	"time"
)

// Subscription represents a user's subscription
type Subscription struct {
	ID                   string     `json:"id" db:"id" example:"sub_123e4567-e89b-12d3-a456-426614174000"`
	UserID               string     `json:"user_id" db:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Tier                 string     `json:"tier" db:"tier" example:"premium"`
	Status               string     `json:"status" db:"status" example:"active"`
	StartDate            time.Time  `json:"start_date" db:"start_date" example:"2024-01-15T10:30:00Z"`
	EndDate              *time.Time `json:"end_date,omitempty" db:"end_date" example:"2024-12-31T23:59:59Z"`
	TrialEndsAt          *time.Time `json:"trial_ends_at,omitempty" db:"trial_ends_at" example:"2024-02-14T23:59:59Z"`
	CancelledAt          *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at" example:"2024-06-30T15:00:00Z"`
	PaymentMethod        string     `json:"payment_method" db:"payment_method" example:"pm_1234567890"`
	StripeCustomerID     string     `json:"-" db:"stripe_customer_id"`
	StripeSubscriptionID string     `json:"-" db:"stripe_subscription_id"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at" example:"2024-07-17T14:45:00Z"`
}

// SubscriptionTier defines tier details
type SubscriptionTier struct {
	ID              string   `json:"id" db:"id" example:"tier_premium"`
	Name            string   `json:"name" db:"name" example:"premium"`
	DisplayName     string   `json:"display_name" db:"display_name" example:"Premium Plan"`
	Description     string   `json:"description" db:"description" example:"Perfect for growing teams and businesses"`
	PriceMonthly    float64  `json:"price_monthly" db:"price_monthly" example:"49.99"`
	PriceYearly     float64  `json:"price_yearly" db:"price_yearly" example:"479.99"`
	MaxPersonas     int      `json:"max_personas" db:"max_personas" example:"50"`
	MaxProjects     int      `json:"max_projects" db:"max_projects" example:"20"`
	MaxContentItems int      `json:"max_content_items" db:"max_content_items" example:"10000"`
	Features        []string `json:"features" db:"features" example:"Advanced Analytics,API Access,Priority Support,Custom Integrations"`
	IsActive        bool     `json:"is_active" db:"is_active" example:"true"`
}

// UsageStats tracks user's resource usage
type UsageStats struct {
	UserID        string    `json:"user_id" db:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	PersonasCount int       `json:"personas_count" db:"personas_count" example:"12"`
	ProjectsCount int       `json:"projects_count" db:"projects_count" example:"5"`
	ContentCount  int       `json:"content_count" db:"content_count" example:"1234"`
	LastUpdated   time.Time `json:"last_updated" db:"last_updated" example:"2024-07-17T14:30:00Z"`
}

// SubscriptionStatus constants
const (
	StatusActive   = "active"
	StatusTrialing = "trialing"
	StatusPastDue  = "past_due"
	StatusCanceled = "canceled"
	StatusExpired  = "expired"
)

// Tier constants
const (
	TierFree       = "free"
	TierBasic      = "basic"
	TierPremium    = "premium"
	TierEnterprise = "enterprise"
)

// CreateSubscriptionRequest for new subscriptions
type CreateSubscriptionRequest struct {
	UserID          string `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Tier            string `json:"tier" binding:"required" example:"premium"`
	PaymentMethodID string `json:"payment_method_id" example:"pm_1234567890"`
	TrialDays       int    `json:"trial_days" example:"14"`
}

// UpdateSubscriptionRequest for subscription changes
type UpdateSubscriptionRequest struct {
	Tier            *string `json:"tier" example:"enterprise"`
	PaymentMethodID *string `json:"payment_method_id" example:"pm_0987654321"`
}

// CheckoutSession for payment processing
type CheckoutSession struct {
	ID         string `json:"id" example:"cs_test_a1b2c3d4e5f6"`
	URL        string `json:"url" example:"https://checkout.stripe.com/pay/cs_test_a1b2c3d4e5f6"`
	SuccessURL string `json:"success_url" example:"https://app.example.com/subscription/success"`
	CancelURL  string `json:"cancel_url" example:"https://app.example.com/subscription/cancelled"`
	ExpiresAt  int64  `json:"expires_at" example:"1689696000"`
}
