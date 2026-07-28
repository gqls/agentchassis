package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/config"
	"github.com/gqls/agentchassis/internal/tools-api/handlers"
	"github.com/gqls/agentchassis/internal/tools-api/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRouter constructs the gin engine with all routes for the tools-api service.
// Later stages attach middleware and handlers to apiGroup.
func NewRouter(pool *pgxpool.Pool, cfg *config.Config) *gin.Engine {
	r := gin.New()

	// gin.Logger() before Recovery so every request is recorded, including one
	// that panics. Added for bugs_open/083: the island ran gin.New() with only
	// Recovery, so `docker compose logs tools-api` showed nothing but the startup
	// banner. With no request log there was no denominator — the 503s could be
	// described as "bursty" from client-side sampling but no honest overall RATE
	// could be quoted, which is why 083 §1 carries an [UNMEASURED] marker.
	// Per-request status + latency here is what turns that into a measurement.
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Local health check — not platform/health per spec hard constraint 6.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// apiGroup is retained for s3/s4/s5/s6 to attach middleware and routes.
	apiGroup := r.Group("/api/v1/tools/gauntlet")
	apiGroup.Use(middleware.CORSMiddleware(pool))

	// Rate limiting and input cap are ordered after CORS so that preflight
	// OPTIONS requests (already aborted with 204 by CORSMiddleware) never
	// reach these checks.
	apiGroup.Use(
		middleware.RateLimitMiddleware(cfg.RateLimitRPS, cfg.RateLimitBurst),
		middleware.InputCapMiddleware(cfg.MaxBodyBytes),
	)

	// /round — POST creates a round row and returns {round_id, provocation}.
	// OPTIONS is registered explicitly so gin matches the path and runs the
	// CORS middleware; CORSMiddleware aborts with 204 before the no-op body
	// is ever reached.
	apiGroup.POST("/round", handlers.RoundHandler(pool))
	apiGroup.OPTIONS("/round", func(c *gin.Context) {})

	// /position — POST returns {counter_position, challenge} via one LLM call.
	// OPTIONS is registered explicitly for preflight support.
	apiGroup.POST("/position", handlers.PositionHandler(pool, cfg))
	apiGroup.OPTIONS("/position", func(c *gin.Context) {})

	// /defend — POST returns {verdict, reasons} via one LLM call.
	// OPTIONS is registered explicitly for preflight support.
	apiGroup.POST("/defend", handlers.DefendHandler(pool, cfg))
	apiGroup.OPTIONS("/defend", func(c *gin.Context) {})

	return r
}
