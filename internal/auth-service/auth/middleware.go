package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/auth-service/jwt"
	"go.uber.org/zap"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtService *jwt.Service, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
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
		c.Set("token_id", claims.ID)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequirePermission checks if user has specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("user_permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No permissions found"})
			c.Abort()
			return
		}

		userPerms, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid permissions format"})
			c.Abort()
			return
		}

		// Check permission
		hasPermission := false
		for _, p := range userPerms {
			if p == permission || p == "*" { // "*" is superuser
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole checks if user has specific role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No role found"})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid role format"})
			c.Abort()
			return
		}

		// Check role
		hasRole := false
		for _, r := range roles {
			if r == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTier checks if user has minimum subscription tier
func RequireTier(minTier string) gin.HandlerFunc {
	tierLevels := map[string]int{
		"free":       0,
		"basic":      1,
		"premium":    2,
		"enterprise": 3,
	}

	return func(c *gin.Context) {
		userTier, exists := c.Get("user_tier")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No subscription tier found"})
			c.Abort()
			return
		}

		tier, ok := userTier.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid tier format"})
			c.Abort()
			return
		}

		userLevel, exists := tierLevels[tier]
		if !exists {
			userLevel = 0 // Default to free
		}

		requiredLevel, exists := tierLevels[minTier]
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid required tier"})
			c.Abort()
			return
		}

		if userLevel < requiredLevel {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         fmt.Sprintf("This feature requires %s tier or higher", minTier),
				"current_tier":  tier,
				"required_tier": minTier,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
