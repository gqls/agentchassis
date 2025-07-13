// FILE: internal/core-manager/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthClaims represents the JWT claims
type AuthClaims struct {
	UserID   string `json:"user_id"`
	ClientID string `json:"client_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Parse token
		token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
			// TODO: Get actual JWT secret from config
			return []byte("your-secret-key"), nil
		})

		if err != nil {
			logger.Error("Failed to parse JWT", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
			c.Set("user_claims", claims)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
	}
}

// TenantMiddleware sets up tenant context
func TenantMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("user_claims").(*AuthClaims)

		// Set tenant context for database operations
		c.Set("client_id", claims.ClientID)
		c.Set("user_id", claims.UserID)

		c.Next()
	}
}

// AdminOnly restricts access to admin users
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("user_claims").(*AuthClaims)

		if claims.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
