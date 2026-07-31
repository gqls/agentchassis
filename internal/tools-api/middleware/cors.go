package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
	"github.com/gqls/agentchassis/internal/tools-api/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CORSMiddleware enforces an Origin allowlist derived from sites with status='deployed'.
// On a match it sets Access-Control-Allow-* headers and attaches site_id/site_domain
// to the gin context (hard constraint 8). Preflight OPTIONS requests are answered
// with 204 and aborted so no handler body runs.
func CORSMiddleware(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		site, err := store.ActiveSiteByOrigin(c.Request.Context(), pool, origin)
		if err != nil || site == nil {
			httperr.JSONError(c, 403, "origin not allowed")
			c.Abort()
			return
		}

		c.Header("Access-Control-Allow-Origin", origin)
		// GET added 2026-07-31 for the public round record (GET /round/:slug).
		// A simple cross-origin GET is not preflighted, so the endpoint would
		// work without this — but an accurate Allow-Methods is what stops the
		// next person adding a header to that fetch and hitting a preflight
		// failure whose cause is three files away.
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Set("site_id", site.ID)
		c.Set("site_domain", site.Domain)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
