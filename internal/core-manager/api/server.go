package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gqls/agentchassis/internal/core-manager/handlers"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/core-manager/admin"
	"github.com/gqls/agentchassis/internal/core-manager/database"
	"github.com/gqls/agentchassis/internal/core-manager/middleware"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
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
//func NewServer(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger, templatesDB, clientsDB *sql.DB) (*Server, error) {

func NewServer(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger, templatesDB, clientsDB *sql.DB) (*Server, error) {
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
	siteAdminHandlers := admin.NewSiteAdminHandlers(personaRepoImpl.ClientsDB(), s.logger)
	pageAdminHandlers := admin.NewPageAdminHandlers(personaRepoImpl.ClientsDB(), s.logger)
	specAdminHandlers := admin.NewSpecAdminHandlers(personaRepoImpl.ClientsDB(), s.logger)

	// Initialize the bootstrap handler
	bootstrapHandler := handlers.NewBootstrapHandler(s.logger, personaRepoImpl.ClientsDB())

	// Health check (no auth)
	s.router.GET("/health", healthHandler.HandleHealth)

	// NOTE: Admin dashboard SPA and auth proxy are served by api-gateway (nginx).
	// Core-manager only handles authenticated API requests.

	// Agent Bootstrap Endpoint (Special Authentication, bypasses AuthMiddleware)
	// This endpoint is for agents to register with a bootstrap key, not a JWT.
	s.router.POST("/api/v1/agents/bootstrap", bootstrapHandler.HandleAgentBootstrap)

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
			adminGroup.POST("/agent-definitions", agentAdminHandlers.HandleCreateAgentDefinition) // NEW
			adminGroup.PUT("/agent-definitions/:type_name", systemHandlers.HandleUpdateAgentDefinition)

			// Kafka Topic Management for Agents (NEW)
			adminGroup.GET("/agent-definitions/:type/topics/verify", agentAdminHandlers.HandleVerifyAgentTopics)
			adminGroup.POST("/agent-definitions/:type/topics/recreate", agentAdminHandlers.HandleRecreateAgentTopics)

			// Agent Instance Management
			adminGroup.GET("/agent-instances", agentAdminHandlers.HandleListAgentInstances)                 // NEW
			adminGroup.GET("/agent-instances/:agent_id", agentAdminHandlers.HandleGetAgentInstance)         // NEW
			adminGroup.PUT("/agent-instances/:agent_id/status", agentAdminHandlers.HandleToggleAgentStatus) // NEW
			adminGroup.POST("/agent-instances/:agent_id/restart", agentAdminHandlers.HandleRestartAgent)    // NEW
			adminGroup.PUT("/clients/:client_id/instances/:instance_id/config", agentAdminHandlers.HandleUpdateInstanceConfig)

			// Topic Management
			adminGroup.POST("/system/cleanup-topics", systemHandlers.HandleCleanupStaleTopics)

			// Site Administration
			siteGroup := adminGroup.Group("/sites")
			{
				siteGroup.GET("", siteAdminHandlers.HandleListSites)
				siteGroup.GET("/:site_id", siteAdminHandlers.HandleGetSite)
				siteGroup.PATCH("/:site_id", siteAdminHandlers.HandleUpdateSite)
				siteGroup.PATCH("/:site_id/specs/:aspect", siteAdminHandlers.HandleUpdateSiteSpec)

				// Spec Direction Control (Phase 4)
				siteGroup.GET("/:site_id/specs", specAdminHandlers.HandleListSpecs)
				siteGroup.POST("/:site_id/specs/:aspect/pin", specAdminHandlers.HandlePinSpec)
				siteGroup.POST("/:site_id/specs/:aspect/unpin", specAdminHandlers.HandleUnpinSpec)
				siteGroup.POST("/:site_id/specs/:aspect/propagate", specAdminHandlers.HandlePropagateSpec)

				// Page Structure (Phase 2)
				siteGroup.GET("/:site_id/pages", pageAdminHandlers.HandleListPages)
				siteGroup.GET("/:site_id/pages/:page_name/components", pageAdminHandlers.HandleListComponents)
				siteGroup.PATCH("/:site_id/pages/:page_name/components/:component_id", pageAdminHandlers.HandleUpdateComponent)
				siteGroup.POST("/:site_id/pages/:page_name/components/:component_id/lock", pageAdminHandlers.HandleLockComponent)
				siteGroup.POST("/:site_id/pages/:page_name/components/:component_id/unlock", pageAdminHandlers.HandleUnlockComponent)
				siteGroup.DELETE("/:site_id/pages/:page_name/components/:component_id", pageAdminHandlers.HandleRemoveComponent)
				siteGroup.POST("/:site_id/pages/:page_name/restore-section", pageAdminHandlers.HandleRestoreSection)
			}

			// Work Item Administration + HITL Review
			workItemGroup := adminGroup.Group("/work-items")
			{
				workItemGroup.POST("", siteAdminHandlers.HandleCreateWorkItem)
				workItemGroup.GET("", siteAdminHandlers.HandleListWorkItems)
				workItemGroup.GET("/:item_id", siteAdminHandlers.HandleGetWorkItem)
				workItemGroup.PATCH("/:item_id", siteAdminHandlers.HandleUpdateWorkItem)
				workItemGroup.POST("/:item_id/retry", siteAdminHandlers.HandleRetryWorkItem)
				workItemGroup.POST("/:item_id/resolve", siteAdminHandlers.HandleResolveWorkItem)
				workItemGroup.POST("/:item_id/approve", siteAdminHandlers.HandleApproveWorkItem)
			}
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
