// FILE: internal/auth-service/auth/interfaces.go
package auth

import (
	"context"
	"github.com/gqls/agentchassis/internal/auth-service/jwt"
	"github.com/gqls/agentchassis/internal/auth-service/user"
)

// ServiceInterface defines the methods that the auth service must implement
type ServiceInterface interface {
	Register(ctx context.Context, req *user.CreateUserRequest) (*TokenResponse, error)
	Login(ctx context.Context, email, password string) (*TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	Logout(ctx context.Context, tokenID string) error
	ValidateToken(tokenString string) (*jwt.Claims, error)
}
