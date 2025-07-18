package api

import (
	"context"
	"fmt"
	"github.com/gqls/agentchassis/internal/core-manager/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/core-manager/admin"
	"github.com/gqls/agentchassis/internal/core-manager/database"
	"github.com/gqls/agentchassis/internal/core-manager/middleware"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Server represents the Core Manager API server
type Server struct {
	ctx           context.Context
	cfg           *config.ServiceConfig
	logger        *zap.Logger
	router        *gin.Engine
	httpServer    *http.Server
	personaRepo   models.PersonaRepository
	kafkaProducer kafka.Producer
}

// NewServer creates a new API server instance
func NewServer(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger, templatesDB, clientsDB *pgxpool.Pool) (*Server, error) {
	// Initialize repositories
	personaRepo := database.NewPersonaRepository(templatesDB, clientsDB, logger)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Initialize Kafka Producer for admin handlers
	kafkaProducer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer for admin handlers: %w", err)
	}

	// Initialize auth middleware config
	authConfig, err := middleware.NewAuthMiddlewareConfig(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth middleware: %w", err)
	}

	server := &Server{
		ctx:           ctx,
		cfg:           cfg,
		logger:        logger,
		router:        router,
		personaRepo:   personaRepo,
		kafkaProducer: kafkaProducer,
	}

	// Setup routes with configured auth middleware
	server.setupRoutes(authConfig)

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	return server, nil
}

// setupRoutes configures all API routes with auth config
func (s *Server) setupRoutes(authConfig *middleware.AuthMiddlewareConfig) {
	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(s.cfg, s.logger)
	templateHandler := handlers.NewTemplateHandler(s.personaRepo, s.logger)
	instanceHandler := handlers.NewInstanceHandler(s.personaRepo, s.logger)

	// Initialize admin handlers
	personaRepoImpl := s.personaRepo.(*database.PersonaRepository)
	clientHandlers := admin.NewClientHandlers(personaRepoImpl.ClientsDB(), s.logger)
	systemHandlers := admin.NewSystemHandlers(personaRepoImpl.ClientsDB(), personaRepoImpl.TemplatesDB(), s.kafkaProducer, s.logger)
	agentAdminHandlers := admin.NewAgentHandlers(personaRepoImpl.ClientsDB(), personaRepoImpl.TemplatesDB(), s.kafkaProducer, s.logger, s.personaRepo)

	// Health check (no auth)
	s.router.GET("/health", healthHandler.HandleHealth)

	// API v1 group with authentication
	apiV1 := s.router.Group("/api/v1")
	apiV1.Use(middleware.AuthMiddleware(authConfig))
	{
		// Template Management (Admin Only)
		templates := apiV1.Group("/templates")
		templates.Use(middleware.AdminOnly())
		{
			templates.POST("", templateHandler.HandleCreateTemplate)
			templates.GET("", templateHandler.HandleListTemplates)
			templates.GET("/:id", templateHandler.HandleGetTemplate)
			templates.PUT("/:id", templateHandler.HandleUpdateTemplate)
			templates.DELETE("/:id", templateHandler.HandleDeleteTemplate)
		}

		// Persona Instance Management (Tenant-scoped)
		instances := apiV1.Group("/personas/instances")
		instances.Use(middleware.TenantMiddleware(s.logger))
		{
			instances.POST("", instanceHandler.HandleCreateInstance)
			instances.GET("", instanceHandler.HandleListInstances)
			instances.GET("/:id", instanceHandler.HandleGetInstance)
			instances.PATCH("/:id", instanceHandler.HandleUpdateInstance)
			instances.DELETE("/:id", instanceHandler.HandleDeleteInstance)
		}

		// Admin Management (Admin Only)
		adminGroup := apiV1.Group("/admin")
		adminGroup.Use(middleware.AdminOnly())
		{
			// Client Management
			adminGroup.POST("/clients", clientHandlers.HandleCreateClient)
			adminGroup.GET("/clients", clientHandlers.HandleListClients)
			adminGroup.GET("/clients/:client_id/usage", clientHandlers.HandleGetClientUsage)

			// System & Workflow Management
			adminGroup.GET("/system/status", systemHandlers.HandleGetSystemStatus)
			adminGroup.GET("/system/kafka/topics", systemHandlers.HandleListKafkaTopics)
			adminGroup.GET("/workflows", systemHandlers.HandleListWorkflows)
			adminGroup.GET("/workflows/:correlation_id", systemHandlers.HandleGetWorkflow)
			adminGroup.POST("/workflows/:correlation_id/resume", systemHandlers.HandleResumeWorkflow)

			// Agent Definition Management
			adminGroup.GET("/agent-definitions", systemHandlers.HandleListAgentDefinitions)
			adminGroup.PUT("/agent-definitions/:type_name", systemHandlers.HandleUpdateAgentDefinition)

			// Agent Instance Management
			adminGroup.PUT("/clients/:client_id/instances/:instance_id/config", agentAdminHandlers.HandleUpdateInstanceConfig)
		}
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting Core Manager API server", zap.String("address", s.httpServer.Addr))
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.kafkaProducer.Close()
	return s.httpServer.Shutdown(ctx)
}

// Address returns the server's address
func (s *Server) Address() string {
	return s.httpServer.Addr
}
