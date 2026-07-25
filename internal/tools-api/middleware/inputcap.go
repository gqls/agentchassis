package middleware

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
)

// InputCapMiddleware returns a gin middleware that rejects request bodies
// larger than maxBytes with a 413, using the shared JSON error shape.
// It restores c.Request.Body so downstream JSON binding still works (hard constraint 9).
func InputCapMiddleware(maxBytes int) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		if len(body) > maxBytes {
			httperr.JSONError(c, 413, "payload too large")
			c.Abort()
			return
		}
		// Restore body so downstream handlers can bind JSON normally.
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Next()
	}
}
