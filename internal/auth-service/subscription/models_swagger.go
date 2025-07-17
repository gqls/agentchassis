package subscription

// NOTE: This file contains model definitions for Swagger documentation.

import "time"

// Subscription represents a user's subscription with complete details
// @Description User subscription information including tier, status, and billing details
type Subscription struct {
	// Unique subscription identifier
	// @example sub_123e4567-e89b-12d3-a456-426614174000
	ID string `json:"id" example:"sub_123e4567-e89b-12d3-a456-426614174000"`

	// User ID associated with this subscription
	// @example 123e4567-e89b-12d3-a456-426614174000
	UserID string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Subscription tier
	// @example premium
	Tier string `json:"tier" example:"premium" enums:"free,basic,premium,enterprise"`

	// Subscription status
	// @example active
	Status string `json:"status" example:"active" enums:"active,trialing,past_due,canceled,expired"`

	// Subscription start date
	// @example 2024-01-15T10:30:00Z
	StartDate time.Time `json:"start_date" example:"2024-01-15T10:30:00Z"`

	// Subscription end date (null for active subscriptions)
	// @example 2024-12-31T23:59:59Z
	EndDate *time.Time `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`

	// Trial end date (null if not in trial)
	// @example 2024-02-14T23:59:59Z
	TrialEndsAt *time.Time `json:"trial_ends_at,omitempty" example:"2024-02-14T23:59:59Z"`

	// Cancellation date (null if not cancelled)
	// @example 2024-06-30T15:00:00Z
	CancelledAt *time.Time `json:"cancelled_at,omitempty" example:"2024-06-30T15:00:00Z"`

	// Payment method identifier
	// @example pm_1234567890
	PaymentMethod string `json:"payment_method" example:"pm_1234567890"`

	// Creation timestamp
	// @example 2024-01-15T10:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Last update timestamp
	// @example 2024-07-17T14:45:00Z
	UpdatedAt time.Time `json:"updated_at" example:"2024-07-17T14:45:00Z"`
}

// SubscriptionTier defines the features and limits for each subscription tier
// @Description Subscription tier details including pricing and resource limits
type SubscriptionTier struct {
	// Tier identifier
	// @example tier_premium
	ID string `json:"id" example:"tier_premium"`

	// Tier name (internal)
	// @example premium
	Name string `json:"name" example:"premium"`

	// Display name for UI
	// @example Premium Plan
	DisplayName string `json:"display_name" example:"Premium Plan"`

	// Tier description
	// @example Perfect for growing teams and businesses
	Description string `json:"description" example:"Perfect for growing teams and businesses"`

	// Monthly price in USD
	// @example 49.99
	PriceMonthly float64 `json:"price_monthly" example:"49.99"`

	// Yearly price in USD (discounted)
	// @example 479.99
	PriceYearly float64 `json:"price_yearly" example:"479.99"`

	// Maximum number of personas (-1 for unlimited)
	// @example 50
	MaxPersonas int `json:"max_personas" example:"50"`

	// Maximum number of projects (-1 for unlimited)
	// @example 20
	MaxProjects int `json:"max_projects" example:"20"`

	// Maximum content items (-1 for unlimited)
	// @example 10000
	MaxContentItems int `json:"max_content_items" example:"10000"`

	// List of feature names included in this tier
	// @example ["Advanced Analytics", "API Access", "Priority Support", "Custom Integrations"]
	Features []string `json:"features" example:"Advanced Analytics,API Access,Priority Support,Custom Integrations"`

	// Whether this tier is currently available
	// @example true
	IsActive bool `json:"is_active" example:"true"`
}

// UsageStats tracks resource usage for quota management
// @Description Current resource usage statistics for the user
type UsageStats struct {
	// User ID
	// @example 123e4567-e89b-12d3-a456-426614174000
	UserID string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Number of active personas
	// @example 12
	PersonasCount int `json:"personas_count" example:"12"`

	// Number of active projects
	// @example 5
	ProjectsCount int `json:"projects_count" example:"5"`

	// Number of content items
	// @example 1234
	ContentCount int `json:"content_count" example:"1234"`

	// Last time stats were updated
	// @example 2024-07-17T14:30:00Z
	LastUpdated time.Time `json:"last_updated" example:"2024-07-17T14:30:00Z"`
}

// CreateSubscriptionRequest for creating new subscriptions
// @Description Request payload for creating a new subscription
type CreateSubscriptionRequest struct {
	// User ID to create subscription for
	// @example 123e4567-e89b-12d3-a456-426614174000
	UserID string `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Subscription tier to assign
	// @example premium
	Tier string `json:"tier" binding:"required" example:"premium" enums:"free,basic,premium,enterprise"`

	// Payment method ID (from payment provider)
	// @example pm_1234567890
	PaymentMethodID string `json:"payment_method_id,omitempty" example:"pm_1234567890"`

	// Number of trial days (0 for no trial)
	// @example 14
	TrialDays int `json:"trial_days,omitempty" example:"14"`
}

// UpdateSubscriptionRequest for modifying subscriptions
// @Description Request payload for updating an existing subscription
type UpdateSubscriptionRequest struct {
	// New tier (optional)
	// @example enterprise
	Tier *string `json:"tier,omitempty" example:"enterprise"`

	// New payment method ID (optional)
	// @example pm_0987654321
	PaymentMethodID *string `json:"payment_method_id,omitempty" example:"pm_0987654321"`
}

// QuotaCheckResponse for quota verification results
// @Description Response for quota check requests
type QuotaCheckResponse struct {
	// Whether the user has available quota
	// @example true
	HasQuota bool `json:"has_quota" example:"true"`

	// Resource type that was checked
	// @example personas
	Resource string `json:"resource" example:"personas"`

	// Current usage count
	// @example 12
	CurrentUsage int `json:"current_usage,omitempty" example:"12"`

	// Maximum allowed for the user's tier
	// @example 50
	MaxAllowed int `json:"max_allowed,omitempty" example:"50"`

	// Remaining quota
	// @example 38
	Remaining int `json:"remaining,omitempty" example:"38"`
}

// SubscriptionListResponse for paginated subscription lists
// @Description Paginated list of subscriptions (admin endpoint)
type SubscriptionListResponse struct {
	// List of subscriptions
	Subscriptions []Subscription `json:"subscriptions"`

	// Total number of subscriptions matching filters
	// @example 156
	TotalCount int `json:"total_count" example:"156"`

	// Current page number
	// @example 1
	Page int `json:"page" example:"1"`

	// Items per page
	// @example 50
	Limit int `json:"limit" example:"50"`

	// Total number of pages
	// @example 4
	TotalPages int `json:"total_pages,omitempty" example:"4"`
}

// CheckoutSession for payment processing
// @Description Checkout session information for payment processing
type CheckoutSession struct {
	// Session ID from payment provider
	// @example cs_test_a1b2c3d4e5f6
	ID string `json:"id" example:"cs_test_a1b2c3d4e5f6"`

	// URL to redirect user for payment
	// @example https://checkout.stripe.com/pay/cs_test_a1b2c3d4e5f6
	URL string `json:"url" example:"https://checkout.stripe.com/pay/cs_test_a1b2c3d4e5f6"`

	// URL to redirect on successful payment
	// @example https://app.example.com/subscription/success
	SuccessURL string `json:"success_url" example:"https://app.example.com/subscription/success"`

	// URL to redirect on cancelled payment
	// @example https://app.example.com/subscription/cancelled
	CancelURL string `json:"cancel_url" example:"https://app.example.com/subscription/cancelled"`

	// Session expiry timestamp (Unix timestamp)
	// @example 1689696000
	ExpiresAt int64 `json:"expires_at" example:"1689696000"`
}
