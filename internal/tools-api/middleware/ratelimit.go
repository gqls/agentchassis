package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/clientip"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware returns a gin middleware that enforces a per-IP
// token-bucket rate limit (in-memory, per-pod — acceptable for v1).
// Requests that exceed the limit are rejected with the shared JSON error shape.
func RateLimitMiddleware(rps float64, burst int) gin.HandlerFunc {
	var mu sync.Mutex
	limiters := make(map[string]*rate.Limiter)

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		l, ok := limiters[ip]
		if !ok {
			l = rate.NewLimiter(rate.Limit(rps), burst)
			limiters[ip] = l
		}
		return l
	}

	return func(c *gin.Context) {
		// clientip.From, not c.ClientIP(): behind Caddy the latter is a constant,
		// so this map held ONE bucket for every visitor. See clientip's package doc.
		limiter := getLimiter(clientip.From(c))
		if !limiter.Allow() {
			httperr.JSONError(c, 429, "rate limit exceeded")
			c.Abort()
			return
		}
		c.Next()
	}
}
