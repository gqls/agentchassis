package auth

import (
	"context"
	"fmt"
	"github.com/gqls/agentchassis/internal/auth-service/jwt"
	"github.com/gqls/agentchassis/internal/auth-service/user"
	"go.uber.org/zap"
)

// Service handles authentication logic
type Service struct {
	userService *user.Service
	jwtService  *jwt.Service
	logger      *zap.Logger
}

// NewService creates a new auth service
func NewService(userService *user.Service, jwtService *jwt.Service, logger *zap.Logger) *Service {
	return &Service{
		userService: userService,
		jwtService:  jwtService,
		logger:      logger,
	}
}

// TokenResponse represents the auth response
type TokenResponse struct {
	AccessToken  string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string    `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string    `json:"token_type" example:"Bearer"`
	ExpiresIn    int       `json:"expires_in" example:"3600"`
	User         *UserInfo `json:"user"`
}

// UserInfo in token response
type UserInfo struct {
	ID            string   `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email         string   `json:"email" example:"john.doe@example.com"`
	ClientID      string   `json:"client_id" example:"client-123"`
	Role          string   `json:"role" example:"user"`
	Tier          string   `json:"tier" example:"premium"`
	EmailVerified bool     `json:"email_verified" example:"true"`
	Permissions   []string `json:"permissions" example:"read:agents,write:agents,read:workflows"`
}

// Register handles user registration
func (s *Service) Register(ctx context.Context, req *user.CreateUserRequest) (*TokenResponse, error) {
	// Create user
	newUser, err := s.userService.Register(ctx, req)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwtService.GenerateTokens(
		newUser.ID,
		newUser.Email,
		newUser.ClientID,
		newUser.Role,
		newUser.SubscriptionTier,
		newUser.Permissions,
	)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err))
		return nil, fmt.Errorf("failed to generate tokens")
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		User: &UserInfo{
			ID:            newUser.ID,
			Email:         newUser.Email,
			ClientID:      newUser.ClientID,
			Role:          newUser.Role,
			Tier:          newUser.SubscriptionTier,
			EmailVerified: newUser.EmailVerified,
			Permissions:   newUser.Permissions,
		},
	}, nil
}

// Login handles user login
func (s *Service) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	// Validate credentials
	user, err := s.userService.Login(ctx, email, password)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwtService.GenerateTokens(
		user.ID,
		user.Email,
		user.ClientID,
		user.Role,
		user.SubscriptionTier,
		user.Permissions,
	)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err))
		return nil, fmt.Errorf("failed to generate tokens")
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: &UserInfo{
			ID:            user.ID,
			Email:         user.Email,
			ClientID:      user.ClientID,
			Role:          user.Role,
			Tier:          user.SubscriptionTier,
			EmailVerified: user.EmailVerified,
			Permissions:   user.Permissions,
		},
	}, nil
}

// RefreshToken handles token refresh
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Validate refresh token and get new access token
	getUserFunc := func(userID string) (*jwt.UserInfo, error) {
		return s.userService.GetUserInfo(ctx, userID)
	}

	accessToken, err := s.jwtService.RefreshAccessToken(refreshToken, getUserFunc)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user info for response
	claims, _ := s.jwtService.ValidateToken(accessToken)

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Return same refresh token
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: &UserInfo{
			ID:          claims.UserID,
			Email:       claims.Email,
			ClientID:    claims.ClientID,
			Role:        claims.Role,
			Tier:        claims.Tier,
			Permissions: claims.Permissions,
		},
	}, nil
}

// Logout handles user logout (for future implementation with token blacklisting)
func (s *Service) Logout(ctx context.Context, tokenID string) error {
	// In a stateless JWT system, logout is typically handled client-side
	// For added security, you could implement token blacklisting here
	s.logger.Info("User logged out", zap.String("token_id", tokenID))
	return nil
}

// ValidateToken validates an access token
func (s *Service) ValidateToken(tokenString string) (*jwt.Claims, error) {
	return s.jwtService.ValidateToken(tokenString)
}
