// FILE: internal/core-manager/middleware/auth.go
package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gqls/agentchassis/platform/config"
	"go.uber.org/zap"
)

// AuthClaims represents the JWT claims
type AuthClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	ClientID    string   `json:"client_id"`
	Role        string   `json:"role"`
	Tier        string   `json:"tier"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// AuthMiddlewareConfig holds configuration for auth middleware
type AuthMiddlewareConfig struct {
	JWTSecret      []byte
	AuthServiceURL string
	Logger         *zap.Logger
}

// NewAuthMiddlewareConfig creates auth middleware configuration from service config
func NewAuthMiddlewareConfig(cfg *config.ServiceConfig, logger *zap.Logger) (*AuthMiddlewareConfig, error) {
	// Get JWT secret from environment
	jwtSecretEnvVar := "JWT_SECRET_KEY"
	if cfg.Custom != nil {
		if envVar, ok := cfg.Custom["jwt_secret_env_var"].(string); ok {
			jwtSecretEnvVar = envVar
		}
	}

	jwtSecret := os.Getenv(jwtSecretEnvVar)
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT secret not found in environment variable %s", jwtSecretEnvVar)
	}

	// Get auth service URL
	authServiceURL := "http://auth-service:8081"
	if cfg.Custom != nil {
		if url, ok := cfg.Custom["auth_service_url"].(string); ok {
			authServiceURL = url
		}
	}

	return &AuthMiddlewareConfig{
		JWTSecret:      []byte(jwtSecret),
		AuthServiceURL: authServiceURL,
		Logger:         logger,
	}, nil
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(config *AuthMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Also check X-Authorization header (for proxied requests)
			authHeader = c.GetHeader("X-Authorization")
		}

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return config.JWTSecret, nil
		})

		if err != nil {
			config.Logger.Debug("Token validation failed", zap.Error(err))

			// Optionally validate with auth service
			if isValid := validateWithAuthService(config, tokenString); !isValid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
		}

		if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
			// Set user context
			c.Set("user_claims", claims)
			c.Set("user_id", claims.UserID)
			c.Set("client_id", claims.ClientID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.Role)
			c.Set("user_tier", claims.Tier)
			c.Set("user_permissions", claims.Permissions)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
	}
}

// validateWithAuthService validates token with auth service (optional fallback)
func validateWithAuthService(config *AuthMiddlewareConfig, token string) bool {
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("POST", config.AuthServiceURL+"/api/v1/auth/validate", nil)
	if err != nil {
		config.Logger.Error("Failed to create validation request", zap.Error(err))
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		config.Logger.Error("Failed to validate with auth service", zap.Error(err))
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var result struct {
		Valid bool `json:"valid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	return result.Valid
}

// TenantMiddleware sets up tenant context
func TenantMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No user claims found"})
			c.Abort()
			return
		}

		authClaims, ok := claims.(*AuthClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid claims type"})
			c.Abort()
			return
		}

		// Set tenant context for database operations
		c.Set("client_id", authClaims.ClientID)
		c.Set("user_id", authClaims.UserID)

		c.Next()
	}
}

// AdminOnly restricts access to admin users
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No role found"})
			c.Abort()
			return
		}

		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
