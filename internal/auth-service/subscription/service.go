// FILE: internal/auth-service/subscription/service.go
package subscription

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles subscription business logic
type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService creates a new subscription service
func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// GetSubscription retrieves a user's subscription
func (s *Service) GetSubscription(ctx context.Context, userID string) (*Subscription, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// CreateSubscription creates a new subscription
func (s *Service) CreateSubscription(ctx context.Context, req *CreateSubscriptionRequest) (*Subscription, error) {
	subscription := &Subscription{
		ID:            uuid.New().String(),
		UserID:        req.UserID,
		Tier:          req.Tier,
		Status:        StatusActive,
		StartDate:     time.Now(),
		PaymentMethod: req.PaymentMethodID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if req.TrialDays > 0 {
		trialEnd := time.Now().AddDate(0, 0, req.TrialDays)
		subscription.TrialEndsAt = &trialEnd
		subscription.Status = StatusTrialing
	}

	if err := s.repo.Create(ctx, subscription); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return subscription, nil
}

// UpdateSubscription updates an existing subscription
func (s *Service) UpdateSubscription(ctx context.Context, userID string, req *UpdateSubscriptionRequest) (*Subscription, error) {
	subscription, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Tier != nil {
		subscription.Tier = *req.Tier
	}

	if req.PaymentMethodID != nil {
		subscription.PaymentMethod = *req.PaymentMethodID
	}

	subscription.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, subscription); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return subscription, nil
}

// CancelSubscription cancels a subscription
func (s *Service) CancelSubscription(ctx context.Context, userID string) error {
	now := time.Now()
	return s.repo.Cancel(ctx, userID, now)
}

// GetUsageStats retrieves usage statistics for a user
func (s *Service) GetUsageStats(ctx context.Context, userID string) (*UsageStats, error) {
	return s.repo.GetUsageStats(ctx, userID)
}

// CheckQuota checks if a user has quota for a specific resource
func (s *Service) CheckQuota(ctx context.Context, userID string, resource string) (bool, error) {
	subscription, err := s.GetSubscription(ctx, userID)
	if err != nil {
		return false, err
	}

	tier, err := s.repo.GetTier(ctx, subscription.Tier)
	if err != nil {
		return false, err
	}

	usage, err := s.GetUsageStats(ctx, userID)
	if err != nil {
		return false, err
	}

	switch resource {
	case "personas":
		return tier.MaxPersonas == -1 || usage.PersonasCount < tier.MaxPersonas, nil
	case "projects":
		return tier.MaxProjects == -1 || usage.ProjectsCount < tier.MaxProjects, nil
	case "content":
		return tier.MaxContentItems == -1 || usage.ContentCount < tier.MaxContentItems, nil
	default:
		return false, fmt.Errorf("unknown resource type: %s", resource)
	}
}
