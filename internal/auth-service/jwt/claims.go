package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// GetTokenFromString extracts claims without validation (for logging/debugging only)
func GetTokenFromString(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid claims type")
}

// IsTokenExpired checks if a token is expired without full validation
func IsTokenExpired(claims *Claims) bool {
	return claims.ExpiresAt.Time.Before(time.Now())
}
