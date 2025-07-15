// FILE: cmd/auth-service/main.go
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/auth-service/admin"
	"github.com/gqls/agentchassis/internal/auth-service/auth"
	"github.com/gqls/agentchassis/internal/auth-service/gateway"
	"github.com/gqls/agentchassis/internal/auth-service/jwt"
	"github.com/gqls/agentchassis/internal/auth-service/middleware"
	"github.com/gqls/agentchassis/internal/auth-service/project"
	"github.com/gqls/agentchassis/internal/auth-service/subscription"
	"github.com/gqls/agentchassis/internal/auth-service/user"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/database"
	"github.com/gqls/agentchassis/platform/logger"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

func main() {
	// --- Step 1: Load Configuration using the Platform Library ---
	configPath := flag.String("config", "configs/auth-service.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to load configuration: %v", err)
	}

	// --- Step 2: Initialize Logger using the Platform Library ---
	appLogger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	appLogger.Info("Auth Service starting",
		zap.String("service_name", cfg.ServiceInfo.Name),
		zap.String("version", cfg.ServiceInfo.Version),
		zap.String("environment", cfg.ServiceInfo.Environment),
	)

	// --- Step 3: Initialize Database Connection using the Platform Library ---
	// The auth service uses MySQL.
	db, err := database.NewMySQLConnection(context.Background(), cfg.Infrastructure.AuthDatabase, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to the auth database", zap.Error(err))
	}
	defer db.Close()

	// --- Step 4: Initialize All Services and Handlers ---

	// Extract JWT configuration from environment and config
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		appLogger.Fatal("JWT_SECRET_KEY environment variable not set")
	}

	// Get JWT expiry from config
	jwtExpiryMinutes := 60 // default
	if cfg.Custom != nil {
		if expiry, ok := cfg.Custom["jwt_expiry_access_minutes"].(float64); ok {
			jwtExpiryMinutes = int(expiry)
		}
	}

	// Initialize JWT service
	jwtSvc, err := jwt.NewService(jwtSecret, jwtExpiryMinutes, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize JWT Service", zap.Error(err))
	}

	// Initialize repositories
	userRepo := user.NewRepository(db, appLogger)
	//adminRepo := admin.NewRepository(db, appLogger, nil) // admin repo doesn't need config
	projectRepo := project.NewRepository(db, appLogger)
	subscriptionRepo := subscription.NewRepository(db, appLogger)

	// Initialize services
	userSvc := user.NewService(userRepo, appLogger)
	authSvc := auth.NewService(userSvc, jwtSvc, appLogger)
	gatewaySvc := gateway.NewService(cfg, appLogger)
	subscriptionSvc := subscription.NewService(subscriptionRepo, appLogger)

	// Initialize handlers
	authHandlers := auth.NewHandlers(authSvc)
	userHandlers := user.NewHandlers(userSvc)
	projectHandler := project.NewHTTPHandler(projectRepo, appLogger)
	subscriptionHandlers := subscription.NewHandlers(subscriptionSvc)
	subscriptionAdminHandlers := subscription.NewAdminHandlers(subscriptionSvc, appLogger)
	gatewayHandler := gateway.NewHTTPHandler(gatewaySvc, appLogger)
	adminHandlers := admin.NewHandlers(userRepo, appLogger)

	// --- Step 5: Setup Routing and Middleware ---
	// Using Gin router for consistency with handlers
	router := gin.New()
	router.Use(gin.Recovery())

	// Public routes (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": cfg.ServiceInfo.Name,
			"version": cfg.ServiceInfo.Version,
		})
	})

	// Auth endpoints (public)
	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/register", authHandlers.HandleRegister)
		authGroup.POST("/login", authHandlers.HandleLogin)
		authGroup.POST("/refresh", authHandlers.HandleRefresh)
		authGroup.POST("/validate", authHandlers.HandleValidate)
		authGroup.POST("/logout", middleware.RequireAuth(jwtSvc, appLogger), authHandlers.HandleLogout)
	}

	// User endpoints (protected)
	userGroup := router.Group("/api/v1/user")
	userGroup.Use(middleware.RequireAuth(jwtSvc, appLogger))
	{
		userGroup.GET("/profile", userHandlers.HandleGetCurrentUser)
		userGroup.PUT("/profile", userHandlers.HandleUpdateCurrentUser)
		userGroup.POST("/password", userHandlers.HandleChangePassword)
		userGroup.DELETE("/delete", userHandlers.HandleDeleteAccount)
	}

	// Subscription endpoints (protected)
	subGroup := router.Group("/api/v1/subscription")
	subGroup.Use(middleware.RequireAuth(jwtSvc, appLogger))
	{
		subGroup.GET("", subscriptionHandlers.HandleGetSubscription)
		subGroup.GET("/usage", subscriptionHandlers.HandleGetUsageStats)
		subGroup.GET("/check-quota", subscriptionHandlers.HandleCheckQuota)
	}

	// Project endpoints (protected)
	projectGroup := router.Group("/api/v1/projects")
	projectGroup.Use(middleware.RequireAuth(jwtSvc, appLogger))
	{
		projectGroup.GET("", wrapHTTPHandler(projectHandler.ListProjects))
		projectGroup.POST("", wrapHTTPHandler(projectHandler.CreateProject))
		projectGroup.GET("/:id", wrapProjectHandler(projectHandler.GetProject))
		projectGroup.PUT("/:id", wrapProjectHandler(projectHandler.UpdateProject))
		projectGroup.DELETE("/:id", wrapProjectHandler(projectHandler.DeleteProject))
	}

	// Admin endpoints (protected + admin role)
	adminGroup := router.Group("/api/v1/admin")
	adminGroup.Use(middleware.RequireAuth(jwtSvc, appLogger))
	adminGroup.Use(middleware.RequireRole("admin"))
	{
		// User management (handled by auth-service)
		adminGroup.GET("/users", adminHandlers.HandleListUsers)
		adminGroup.GET("/users/:user_id", adminHandlers.HandleGetUser)
		adminGroup.PUT("/users/:user_id", adminHandlers.HandleUpdateUser)
		adminGroup.DELETE("/users/:user_id", adminHandlers.HandleDeleteUser)
		adminGroup.GET("/users/:user_id/activity", adminHandlers.HandleGetUserActivity)
		adminGroup.POST("/users/:user_id/permissions", adminHandlers.HandleGrantPermission)
		adminGroup.DELETE("/users/:user_id/permissions/:permission_name", adminHandlers.HandleRevokePermission)

		// Subscription management (handled by auth-service)
		adminGroup.GET("/subscriptions", subscriptionAdminHandlers.HandleListSubscriptions)
		adminGroup.POST("/subscriptions", subscriptionAdminHandlers.HandleCreateSubscription)
		adminGroup.PUT("/subscriptions/:user_id", wrapAdminSubscriptionHandler(subscriptionAdminHandlers.HandleUpdateSubscription))

		// Routes to be proxied to core-manager
		adminGroup.Any("/clients", gatewayHandler.HandleAdminRoutes)
		adminGroup.Any("/clients/*path", gatewayHandler.HandleAdminRoutes)
		adminGroup.Any("/system/*path", gatewayHandler.HandleAdminRoutes)
		adminGroup.Any("/workflows/*path", gatewayHandler.HandleAdminRoutes)
		adminGroup.Any("/agent-definitions/*path", gatewayHandler.HandleAdminRoutes)

	}

	// Gateway proxy endpoints (protected)
	gatewayGroup := router.Group("/api/v1")
	gatewayGroup.Use(middleware.RequireAuth(jwtSvc, appLogger))
	{
		// Template management (admin only)
		templateGroup := gatewayGroup.Group("/templates")
		templateGroup.Use(middleware.RequireRole("admin"))
		{
			templateGroup.Any("", gatewayHandler.HandleTemplateRoutes)
			templateGroup.Any("/*path", gatewayHandler.HandleTemplateRoutes)
		}

		// Instance management
		gatewayGroup.Any("/personas/instances", gatewayHandler.HandleInstanceRoutes)
		gatewayGroup.Any("/personas/instances/*path", gatewayHandler.HandleInstanceRoutes)
	}

	// WebSocket endpoint
	router.GET("/ws", middleware.RequireAuth(jwtSvc, appLogger), gatewayHandler.HandleWebSocket)

	// Apply CORS middleware
	allowedOrigins := []string{"*"} // default
	if cfg.Custom != nil {
		if origins, ok := cfg.Custom["allowed_origins"].([]interface{}); ok {
			allowedOrigins = make([]string, len(origins))
			for i, origin := range origins {
				allowedOrigins[i] = origin.(string)
			}
		}
	}

	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	// --- Step 6: Start Server and Handle Graceful Shutdown ---
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: corsConfig.Handler(router),
	}

	// Start server in a goroutine
	go func() {
		appLogger.Info("Auth Service listening", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal("Auth Service listen and serve error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("Shutdown signal received, shutting down auth server...")

	// Graceful shutdown with timeout
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		appLogger.Fatal("Auth Server forced to shutdown due to error", zap.Error(err))
	}
	appLogger.Info("Auth Server exited gracefully")
}

// wrapHTTPHandler wraps standard http handlers to work with gin
func wrapHTTPHandler(fn func(http.ResponseWriter, *http.Request)) gin.HandlerFunc {
	return func(c *gin.Context) {
		fn(c.Writer, c.Request)
	}
}

// wrapProjectHandler wraps project handlers that take an ID parameter
func wrapProjectHandler(fn func(http.ResponseWriter, *http.Request, string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		fn(c.Writer, c.Request, id)
	}
}

// wrapAdminSubscriptionHandler wraps the admin subscription update handler
func wrapAdminSubscriptionHandler(fn func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// The user_id is already available as a URL parameter
		c.Set("param_user_id", c.Param("user_id"))
		fn(c)
	}
}
