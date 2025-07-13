package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Service handles JWT operations
type Service struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	logger          *zap.Logger
}

// Claims represents the JWT claims
type Claims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	ClientID    string   `json:"client_id"`
	Role        string   `json:"role"`
	Tier        string   `json:"tier"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// RefreshClaims for refresh tokens
type RefreshClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// NewService creates a new JWT service
func NewService(secretKey string, accessMinutes int, logger *zap.Logger) (*Service, error) {
	if secretKey == "" {
		return nil, fmt.Errorf("JWT secret key cannot be empty")
	}

	return &Service{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  time.Duration(accessMinutes) * time.Minute,
		refreshTokenTTL: 7 * 24 * time.Hour, // 7 days
		logger:          logger,
	}, nil
}

// GenerateTokens creates both access and refresh tokens
func (s *Service) GenerateTokens(userID, email, clientID, role, tier string, permissions []string) (string, string, error) {
	accessToken, err := s.generateAccessToken(userID, email, clientID, role, tier, permissions)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// generateAccessToken creates an access token with full claims
func (s *Service) generateAccessToken(userID, email, clientID, role, tier string, permissions []string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:      userID,
		Email:       email,
		ClientID:    clientID,
		Role:        role,
		Tier:        tier,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ai-persona-system",
			Subject:   userID,
			ID:        fmt.Sprintf("%d", now.Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// generateRefreshToken creates a refresh token with minimal claims
func (s *Service) generateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID,
			ID:        fmt.Sprintf("refresh_%d", now.Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates and parses an access token
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// ValidateRefreshToken validates a refresh token
func (s *Service) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid refresh token")
}

// RefreshAccessToken creates a new access token from a refresh token
func (s *Service) RefreshAccessToken(refreshToken string, getUserFunc func(userID string) (*UserInfo, error)) (string, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Get updated user details
	userInfo, err := getUserFunc(claims.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to get user details: %w", err)
	}

	// Generate new access token with current user info
	return s.generateAccessToken(
		userInfo.UserID,
		userInfo.Email,
		userInfo.ClientID,
		userInfo.Role,
		userInfo.Tier,
		userInfo.Permissions,
	)
}

// UserInfo holds user information for token generation
type UserInfo struct {
	UserID      string
	Email       string
	ClientID    string
	Role        string
	Tier        string
	Permissions []string
}
