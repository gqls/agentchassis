// FILE: internal/core-manager/api/server.go
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/internal/core-manager/admin" // Import the new admin package
	"github.com/gqls/agentchassis/internal/core-manager/database"
	"github.com/gqls/agentchassis/internal/core-manager/middleware"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka" // Import kafka for admin handlers
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
	kafkaProducer kafka.Producer // Add kafka producer
}

// NewServer function updated to include kafkaProducer
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
	server.setupRoutesWithAuth(router, authConfig)

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	return server, nil
}

// setupRoutesWithAuth configures all API routes with auth config
func (s *Server) setupRoutesWithAuth(router *gin.Engine, authConfig *middleware.AuthMiddlewareConfig) {
	// Health check (no auth)
	router.GET("/health", s.handleHealth)

	// API v1 group with authentication
	apiV1 := router.Group("/api/v1")
	apiV1.Use(middleware.AuthMiddleware(authConfig))
	{
		// Template Management (Admin Only)
		templates := apiV1.Group("/templates")
		templates.Use(middleware.AdminOnly())
		{
			templates.POST("", s.handleCreateTemplate)
			templates.GET("", s.handleListTemplates)
			templates.GET("/:id", s.handleGetTemplate)
			templates.PUT("/:id", s.handleUpdateTemplate)
			templates.DELETE("/:id", s.handleDeleteTemplate)
		}

		// Persona Instance Management (Tenant-scoped)
		instances := apiV1.Group("/personas/instances")
		instances.Use(middleware.TenantMiddleware(s.logger))
		{
			instances.POST("", s.handleCreateInstance)
			instances.GET("", s.handleListInstances)
			instances.GET("/:id", s.handleGetInstance)
			instances.PATCH("/:id", s.handleUpdateInstance)
			instances.DELETE("/:id", s.handleDeleteInstance)
		}

		// Admin Management (Admin Only)
		adminGroup := apiV1.Group("/admin")
		adminGroup.Use(middleware.AdminOnly())
		{
			// Initialize admin handlers
			clientHandlers := admin.NewClientHandlers(s.personaRepo.(*database.PersonaRepository).ClientsDB(), s.logger)
			systemHandlers := admin.NewSystemHandlers(s.personaRepo.(*database.PersonaRepository).ClientsDB(), s.personaRepo.(*database.PersonaRepository).TemplatesDB(), s.kafkaProducer, s.logger)

			// Initialize AgentHandlers with the personaRepo dependency
			agentAdminHandlers := admin.NewAgentHandlers(s.personaRepo.(*database.PersonaRepository).ClientsDB(), s.personaRepo.(*database.PersonaRepository).TemplatesDB(), s.kafkaProducer, s.logger, s.personaRepo)

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
	s.kafkaProducer.Close() // Close the producer on shutdown
	return s.httpServer.Shutdown(ctx)
}

// Address returns the server's address
func (s *Server) Address() string {
	return s.httpServer.Addr
}

// Health check handler
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": s.cfg.ServiceInfo.Name,
		"version": s.cfg.ServiceInfo.Version,
	})
}

// Template Handlers

func (s *Server) handleCreateTemplate(c *gin.Context) {
	var req models.Persona
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	req.ID = uuid.New()
	req.IsTemplate = true
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	createdTemplate, err := s.personaRepo.CreateTemplate(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to create template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, createdTemplate)
}

func (s *Server) handleListTemplates(c *gin.Context) {
	templates, err := s.personaRepo.ListTemplates(c.Request.Context())
	if err != nil {
		s.logger.Error("Failed to list templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (s *Server) handleGetTemplate(c *gin.Context) {
	templateID := c.Param("id")
	template, err := s.personaRepo.GetTemplateByID(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	c.JSON(http.StatusOK, template)
}

func (s *Server) handleUpdateTemplate(c *gin.Context) {
	templateID := c.Param("id")
	var req models.Persona
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.ID, _ = uuid.Parse(templateID)
	req.UpdatedAt = time.Now()

	updatedTemplate, err := s.personaRepo.UpdateTemplate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		return
	}
	c.JSON(http.StatusOK, updatedTemplate)
}

func (s *Server) handleDeleteTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if err := s.personaRepo.DeleteTemplate(c.Request.Context(), templateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Instance Handlers

// handleCreateInstance with proper context
func (s *Server) handleCreateInstance(c *gin.Context) {
	claims := c.MustGet("user_claims").(*middleware.AuthClaims)

	var req struct {
		TemplateID   string `json:"template_id" binding:"required,uuid"`
		InstanceName string `json:"instance_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	instance, err := s.personaRepo.CreateInstanceFromTemplate(c.Request.Context(),
		req.TemplateID, claims.UserID, req.InstanceName)
	if err != nil {
		s.logger.Error("Failed to create instance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create instance"})
		return
	}
	c.JSON(http.StatusCreated, instance)
}

func (s *Server) handleGetInstance(c *gin.Context) {
	instanceID := c.Param("id")
	instance, err := s.personaRepo.GetInstanceByID(c.Request.Context(), instanceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}
	c.JSON(http.StatusOK, instance)
}

func (s *Server) handleListInstances(c *gin.Context) {
	claims := c.MustGet("user_claims").(*middleware.AuthClaims)
	instances, err := s.personaRepo.ListInstances(c.Request.Context(), claims.UserID)
	if err != nil {
		s.logger.Error("Failed to list instances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve instances"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"instances": instances})
}

func (s *Server) handleUpdateInstance(c *gin.Context) {
	instanceID := c.Param("id")
	var req struct {
		Name   *string                `json:"name"`
		Config map[string]interface{} `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedInstance, err := s.personaRepo.UpdateInstance(c.Request.Context(), instanceID, req.Name, req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instance"})
		return
	}
	c.JSON(http.StatusOK, updatedInstance)
}

func (s *Server) handleDeleteInstance(c *gin.Context) {
	instanceID := c.Param("id")
	if err := s.personaRepo.DeleteInstance(c.Request.Context(), instanceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete instance"})
		return
	}
	c.Status(http.StatusNoContent)
}
