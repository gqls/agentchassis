package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/internal/auth-service/jwt"
	"go.uber.org/zap"
)

// Service handles business logic for users
type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService creates a new user service
func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *CreateUserRequest) (*User, error) {
	// Validate email format
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user already exists
	existingUser, _ := s.repo.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Create user
	user, err := s.repo.CreateUser(ctx, req)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	// TODO: Send verification email

	s.logger.Info("User registered successfully",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email))

	return user, nil
}

// Login validates user credentials
func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("Login attempt for non-existent user", zap.String("email", email))
		return nil, fmt.Errorf("invalid credentials")
	}

	if !s.repo.ValidatePassword(user, password) {
		s.logger.Warn("Invalid password attempt", zap.String("user_id", user.ID))
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	// Update last login
	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		s.logger.Error("Failed to update last login", zap.Error(err))
	}

	return user, nil
}

// GetUser retrieves user details
func (s *Service) GetUser(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// GetUserInfo returns user info for JWT token generation
func (s *Service) GetUserInfo(ctx context.Context, userID string) (*jwt.UserInfo, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &jwt.UserInfo{
		UserID:      user.ID,
		Email:       user.Email,
		ClientID:    user.ClientID,
		Role:        user.Role,
		Tier:        user.SubscriptionTier,
		Permissions: user.Permissions,
	}, nil
}

// UpdateUser updates user information
func (s *Service) UpdateUser(ctx context.Context, userID string, req *UpdateUserRequest) error {
	return s.repo.UpdateUser(ctx, userID, req)
}

// ChangePassword changes user password
func (s *Service) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Validate current password
	if !s.repo.ValidatePassword(user, req.CurrentPassword) {
		return fmt.Errorf("current password is incorrect")
	}

	// Update password
	return s.repo.UpdatePassword(ctx, userID, req.NewPassword)
}

// DeleteUser deletes a user account
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	return s.repo.DeleteUser(ctx, userID)
}
