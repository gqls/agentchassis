package middleware

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/clientip"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
	"github.com/gqls/agentchassis/platform/httpguard"
)

// BandedRateLimit is the three-line gin adapter httpguard deliberately does not
// ship (its package doc: net/http only, callers write their own). One Limiter
// per ROUTE, because httpguard has no notion of endpoints — that is how the
// gripper group gets "6/hour and 20/day on /session, 60/hour and 200/day on
// /chat, 3/hour and 10/day on /submit" instead of the one flat RPS bucket the
// gauntlet group runs (RateLimitMiddleware). Same visitor key as everywhere
// else in this service: clientip.From, never c.ClientIP().
//
// A refusal carries Retry-After (whole seconds, rounded up) so a well-behaved
// widget can tell the visitor when to come back rather than guessing.
func BandedRateLimit(l *httpguard.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, retryAfter := l.Allow(clientip.From(c))
		if !ok {
			secs := int(math.Ceil(retryAfter.Seconds()))
			if secs < 1 {
				secs = 1
			}
			c.Header("Retry-After", strconv.Itoa(secs))
			httperr.JSONError(c, 429, "rate limit exceeded")
			c.Abort()
			return
		}
		c.Next()
	}
}
