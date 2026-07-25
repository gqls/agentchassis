package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRouter constructs the gin engine with all routes for the tools-api service.
// Later stages attach middleware and handlers to apiGroup.
func NewRouter(pool *pgxpool.Pool, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// Local health check — not platform/health per spec hard constraint 6.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// apiGroup is retained for s3/s4/s5/s6 to attach middleware and routes.
	apiGroup := r.Group("/api/v1/tools/gauntlet")
	_ = apiGroup

	return r
}
