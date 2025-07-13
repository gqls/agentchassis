// FILE: internal/auth-service/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gqls/ai-persona-system/internal/auth-service/jwt"
	"go.uber.org/zap"
)

// RequireAuth validates JWT tokens
func RequireAuth(jwtService *jwt.Service, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			logger.Debug("Token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("client_id", claims.ClientID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_tier", claims.Tier)
		c.Set("user_permissions", claims.Permissions)

		c.Next()
	}
}
