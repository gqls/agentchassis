// FILE: cmd/auth-service/main.go
// This is the refactored main entrypoint for the auth-service.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// Internal auth-service packages
	"github.com/gqls/personae-auth-service/internal/admin"
	"github.com/gqls/personae-auth-service/internal/auth"
	"github.com/gqls/personae-auth-service/internal/gateway"
	"github.com/gqls/personae-auth-service/internal/jwt"
	"github.com/gqls/personae-auth-service/internal/middleware"
	"github.com/gqls/personae-auth-service/internal/project"
	"github.com/gqls/personae-auth-service/internal/subscription"
	"github.com/gqls/personae-auth-service/internal/user"

	// --- Use platform packages ---
	"github.com/gqls/ai-persona-system/platform/config"
	"github.com/gqls/ai-persona-system/platform/database"
	"github.com/gqls/ai-persona-system/platform/logger"

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
	// This logic remains largely the same, but now it's cleaner because
	// the dependencies are initialized above in a standard way.

	// Extract custom config needed for JWT service
	// (This assumes we add a generic `custom` map[string]interface{} to our platform config struct)
	// For now, we'll manually construct a temporary config for JWT.
	jwtSecret := os.Getenv("JWT_SECRET_KEY") // Read from environment
	if jwtSecret == "" {
		appLogger.Fatal("JWT_SECRET_KEY environment variable not set")
	}

	tempJwtCfg := &auth_config.Config{ // Assuming auth_config is the old config package
		JWTSecretKey:           jwtSecret,
		JWTExpiryAccessMinutes: 60, // This would also come from the new config structure
	}

	jwtSvc, err := jwt.NewService(tempJwtCfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize JWT Service", zap.Error(err))
	}

	// Initialize repositories
	userRepo := user.NewRepository(db, appLogger, tempJwtCfg) // Pass old config for now
	adminRepo := admin.NewRepository(db, appLogger, tempJwtCfg)
	projectRepo := project.NewRepository(db, appLogger)
	subscriptionRepo := subscription.NewRepository(db, appLogger)

	// Initialize services
	userSvc := user.NewService(userRepo, appLogger)
	authSvc := auth.NewService(userSvc, jwtSvc, appLogger)
	gatewaySvc := gateway.NewService(tempJwtCfg, appLogger)
	subscriptionSvc := subscription.NewService(subscriptionRepo, appLogger)

	// Initialize handlers
	authAPIHandler := auth.NewHTTPHandler(authSvc, userSvc, jwtSvc, adminRepo, gatewaySvc, appLogger, tempJwtCfg)
	projectHandler := project.NewHTTPHandler(projectRepo, appLogger)
	subscriptionHandler := subscription.NewHTTPHandler(subscriptionSvc, appLogger)
	subscriptionAdminHandler := subscription.NewAdminHandlers(subscriptionSvc, appLogger)
	gatewayHandler := gateway.NewHTTPHandler(gatewaySvc, jwtSvc, appLogger, subscriptionSvc, projectRepo)

	// --- Step 5: Setup Routing and Middleware ---
	mux := http.NewServeMux()

	// Public routes (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": cfg.ServiceInfo.Name,
			"version": cfg.ServiceInfo.Version,
		})
	})

	// Auth endpoints
	mux.HandleFunc("/api/v1/auth/register", authAPIHandler.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", authAPIHandler.HandleLogin)
	mux.HandleFunc("/api/v1/auth/refresh", authAPIHandler.HandleRefresh)
	mux.HandleFunc("/api/v1/auth/validate", authAPIHandler.HandleValidate)

	// Protected routes group
	protectedMux := http.NewServeMux()

	// User management (protected)
	protectedMux.HandleFunc("/api/v1/auth/logout", authAPIHandler.HandleLogout)
	protectedMux.HandleFunc("/api/v1/user/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.HandleGetCurrentUser(w, r)
		case http.MethodPut:
			userHandler.HandleUpdateCurrentUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/v1/user/password", userHandler.HandleChangePassword)
	protectedMux.HandleFunc("/api/v1/user/delete", userHandler.HandleDeleteAccount)

	// Subscription endpoints (protected)
	protectedMux.HandleFunc("/api/v1/subscription", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			subscriptionHandler.HandleGetSubscription(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/v1/subscription/usage", subscriptionHandler.HandleGetUsageStats)
	protectedMux.HandleFunc("/api/v1/subscription/check-quota", subscriptionHandler.HandleCheckQuota)

	// Project endpoints (protected)
	protectedMux.HandleFunc("/api/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectHandler.ListProjects(w, r)
		case http.MethodPost:
			projectHandler.CreateProject(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/v1/projects/", func(w http.ResponseWriter, r *http.Request) {
		// Extract project ID from path
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 5 {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}
		projectID := parts[4]

		switch r.Method {
		case http.MethodGet:
			projectHandler.GetProject(w, r, projectID)
		case http.MethodPut:
			projectHandler.UpdateProject(w, r, projectID)
		case http.MethodDelete:
			projectHandler.DeleteProject(w, r, projectID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Admin routes
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/api/v1/admin/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			subscriptionAdminHandler.HandleCreateSubscription(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	adminMux.HandleFunc("/api/v1/admin/subscriptions/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 6 {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
		userID := parts[5]

		switch r.Method {
		case http.MethodPut:
			subscriptionAdminHandler.HandleUpdateSubscription(w, r, userID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Gateway proxy endpoints (protected) - these proxy to core-manager
	gatewayMux := http.NewServeMux()

	// Template management (admin only)
	gatewayMux.HandleFunc("/api/v1/templates", gatewayHandler.HandleTemplateRoutes)
	gatewayMux.HandleFunc("/api/v1/templates/", gatewayHandler.HandleTemplateRoutes)

	// Instance management
	gatewayMux.HandleFunc("/api/v1/personas/instances", gatewayHandler.HandleInstanceRoutes)
	gatewayMux.HandleFunc("/api/v1/personas/instances/", gatewayHandler.HandleInstanceRoutes)

	// WebSocket proxy
	gatewayMux.HandleFunc("/ws", gatewayHandler.HandleWebSocket)

	// Apply middleware chains
	// 1. Apply auth middleware to protected routes
	protectedHandler := middleware.RequireAuth(jwtSvc, appLogger)(protectedMux)

	// 2. Apply admin middleware to admin routes
	adminHandler := middleware.RequireAuth(jwtSvc, appLogger)(
		middleware.RequireRole("admin")(adminMux),
	)

	// 3. Apply auth middleware to gateway routes
	gatewayHandler := middleware.RequireAuth(jwtSvc, appLogger)(gatewayMux)

	// Combine all handlers
	mux.Handle("/api/v1/user/", protectedHandler)
	mux.Handle("/api/v1/subscription", protectedHandler)
	mux.Handle("/api/v1/subscription/", protectedHandler)
	mux.Handle("/api/v1/projects", protectedHandler)
	mux.Handle("/api/v1/projects/", protectedHandler)
	mux.Handle("/api/v1/admin/", adminHandler)
	mux.Handle("/api/v1/templates", gatewayHandler)
	mux.Handle("/api/v1/templates/", gatewayHandler)
	mux.Handle("/api/v1/personas/", gatewayHandler)
	mux.Handle("/ws", gatewayHandler)

	// Apply global middleware
	loggedMux := middleware.LoggingMiddleware(appLogger)(mux)

	// Configure CORS
	corsOptions := cors.Options{
		AllowedOrigins:   cfg.Custom["allowed_origins"].([]string),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300,
	}
	handlerWithCORS := cors.New(corsOptions).Handler(loggedMux)
}
