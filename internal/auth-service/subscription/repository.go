// FILE: internal/auth-service/subscription/repository.go
package subscription

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Repository handles subscription data access
type Repository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewRepository creates a new subscription repository
func NewRepository(db *sql.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// GetByUserID retrieves a subscription by user ID
func (r *Repository) GetByUserID(ctx context.Context, userID string) (*Subscription, error) {
	var s Subscription
	query := `
		SELECT id, user_id, tier, status, start_date, end_date, trial_ends_at, 
		       cancelled_at, payment_method, stripe_customer_id, stripe_subscription_id,
		       created_at, updated_at
		FROM subscriptions
		WHERE user_id = ?
	`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.Tier, &s.Status, &s.StartDate, &s.EndDate,
		&s.TrialEndsAt, &s.CancelledAt, &s.PaymentMethod,
		&s.StripeCustomerID, &s.StripeSubscriptionID,
		&s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, err
	}

	return &s, nil
}

// Create creates a new subscription
func (r *Repository) Create(ctx context.Context, s *Subscription) error {
	query := `
		INSERT INTO subscriptions (id, user_id, tier, status, start_date, payment_method,
		                          trial_ends_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.UserID, s.Tier, s.Status, s.StartDate,
		s.PaymentMethod, s.TrialEndsAt, s.CreatedAt, s.UpdatedAt,
	)

	return err
}

// Update updates an existing subscription
func (r *Repository) Update(ctx context.Context, s *Subscription) error {
	query := `
		UPDATE subscriptions 
		SET tier = ?, status = ?, payment_method = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		s.Tier, s.Status, s.PaymentMethod, s.UpdatedAt, s.ID,
	)

	return err
}

// Cancel cancels a subscription
func (r *Repository) Cancel(ctx context.Context, userID string, cancelledAt time.Time) error {
	query := `
		UPDATE subscriptions 
		SET status = ?, cancelled_at = ?, updated_at = ?
		WHERE user_id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		StatusCanceled, cancelledAt, time.Now(), userID,
	)

	return err
}

// GetTier retrieves tier information
func (r *Repository) GetTier(ctx context.Context, tierName string) (*SubscriptionTier, error) {
	var t SubscriptionTier
	var featuresJSON string

	query := `
		SELECT id, name, display_name, description, price_monthly, price_yearly,
		       max_personas, max_projects, max_content_items, features, is_active
		FROM subscription_tiers
		WHERE name = ? AND is_active = true
	`

	err := r.db.QueryRowContext(ctx, query, tierName).Scan(
		&t.ID, &t.Name, &t.DisplayName, &t.Description,
		&t.PriceMonthly, &t.PriceYearly,
		&t.MaxPersonas, &t.MaxProjects, &t.MaxContentItems,
		&featuresJSON, &t.IsActive,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(featuresJSON), &t.Features)

	return &t, nil
}

// GetUsageStats retrieves usage statistics
func (r *Repository) GetUsageStats(ctx context.Context, userID string) (*UsageStats, error) {
	var stats UsageStats
	stats.UserID = userID

	// This would need to query across multiple tables/schemas
	// For now, returning mock data
	stats.PersonasCount = 0
	stats.ProjectsCount = 0
	stats.ContentCount = 0
	stats.LastUpdated = time.Now()

	return &stats, nil
}

// ListSubscriptionsParams contains parameters for listing subscriptions
type ListSubscriptionsParams struct {
	Limit  int
	Offset int
	Status string
	Tier   string
}

// ListAll retrieves a paginated list of all subscriptions
func (r *Repository) ListAll(ctx context.Context, params ListSubscriptionsParams) ([]Subscription, int, error) {
	query := `SELECT id, user_id, tier, status, start_date, end_date, created_at FROM subscriptions WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM subscriptions WHERE 1=1`

	// This service runs on MySQL: placeholders are `?`, matching every other
	// query in this file. These were `$N` (Postgres) — dead code until now.
	args := []interface{}{}

	if params.Status != "" {
		query += " AND status = ?"
		countQuery += " AND status = ?"
		args = append(args, params.Status)
	}
	if params.Tier != "" {
		query += " AND tier = ?"
		countQuery += " AND tier = ?"
		args = append(args, params.Tier)
	}

	// Get total count (before the limit/offset args are appended)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get subscription count: %w", err)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, params.Limit, params.Offset)

	// Get subscriptions
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.Tier, &s.Status, &s.StartDate, &s.EndDate, &s.CreatedAt); err != nil {
			r.logger.Error("Failed to scan subscription row", zap.Error(err))
			continue
		}
		subscriptions = append(subscriptions, s)
	}

	return subscriptions, total, nil
}
