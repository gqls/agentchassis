package subscription

import (
	"time"
)

// Subscription represents a user's subscription
type Subscription struct {
	ID                   string     `json:"id" db:"id"`
	UserID               string     `json:"user_id" db:"user_id"`
	Tier                 string     `json:"tier" db:"tier"`
	Status               string     `json:"status" db:"status"`
	StartDate            time.Time  `json:"start_date" db:"start_date"`
	EndDate              *time.Time `json:"end_date,omitempty" db:"end_date"`
	TrialEndsAt          *time.Time `json:"trial_ends_at,omitempty" db:"trial_ends_at"`
	CancelledAt          *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	PaymentMethod        string     `json:"payment_method" db:"payment_method"`
	StripeCustomerID     string     `json:"-" db:"stripe_customer_id"`
	StripeSubscriptionID string     `json:"-" db:"stripe_subscription_id"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// SubscriptionTier defines tier details
type SubscriptionTier struct {
	ID              string   `json:"id" db:"id"`
	Name            string   `json:"name" db:"name"`
	DisplayName     string   `json:"display_name" db:"display_name"`
	Description     string   `json:"description" db:"description"`
	PriceMonthly    float64  `json:"price_monthly" db:"price_monthly"`
	PriceYearly     float64  `json:"price_yearly" db:"price_yearly"`
	MaxPersonas     int      `json:"max_personas" db:"max_personas"`
	MaxProjects     int      `json:"max_projects" db:"max_projects"`
	MaxContentItems int      `json:"max_content_items" db:"max_content_items"`
	Features        []string `json:"features" db:"features"`
	IsActive        bool     `json:"is_active" db:"is_active"`
}

// UsageStats tracks user's resource usage
type UsageStats struct {
	UserID        string    `json:"user_id" db:"user_id"`
	PersonasCount int       `json:"personas_count" db:"personas_count"`
	ProjectsCount int       `json:"projects_count" db:"projects_count"`
	ContentCount  int       `json:"content_count" db:"content_count"`
	LastUpdated   time.Time `json:"last_updated" db:"last_updated"`
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
	UserID          string `json:"user_id" binding:"required"`
	Tier            string `json:"tier" binding:"required"`
	PaymentMethodID string `json:"payment_method_id"`
	TrialDays       int    `json:"trial_days"`
}

// UpdateSubscriptionRequest for subscription changes
type UpdateSubscriptionRequest struct {
	Tier            *string `json:"tier"`
	PaymentMethodID *string `json:"payment_method_id"`
}

// CheckoutSession for payment processing
type CheckoutSession struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
	ExpiresAt  int64  `json:"expires_at"`
}
