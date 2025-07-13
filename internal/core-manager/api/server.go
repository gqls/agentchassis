// FILE: internal/core-manager/api/server.go
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gqls/ai-persona-system/internal/core-manager/database"
	"github.com/gqls/ai-persona-system/internal/core-manager/middleware"
	"github.com/gqls/ai-persona-system/pkg/models"
	"github.com/gqls/ai-persona-system/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Server represents the Core Manager API server
type Server struct {
	ctx         context.Context
	cfg         *config.ServiceConfig
	logger      *zap.Logger
	router      *gin.Engine
	httpServer  *http.Server
	personaRepo models.PersonaRepository
}

// NewServer creates a new API server instance
func NewServer(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger, templatesDB, clientsDB *pgxpool.Pool) (*Server, error) {
	// Initialize repositories
	personaRepo := database.NewPersonaRepository(templatesDB, clientsDB, logger)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	server := &Server{
		ctx:         ctx,
		cfg:         cfg,
		logger:      logger,
		router:      router,
		personaRepo: personaRepo,
	}

	// Setup routes
	server.setupRoutes(router)

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	return server, nil
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", s.handleHealth)

	// API v1 group with authentication
	apiV1 := router.Group("/api/v1")
	apiV1.Use(middleware.AuthMiddleware(s.logger))
	apiV1.Use(middleware.TenantMiddleware(s.logger))
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

		// Persona Instance Management
		instances := apiV1.Group("/personas/instances")
		{
			instances.POST("", s.handleCreateInstance)
			instances.GET("", s.handleListInstances)
			instances.GET("/:id", s.handleGetInstance)
			instances.PATCH("/:id", s.handleUpdateInstance)
			instances.DELETE("/:id", s.handleDeleteInstance)
		}

		// Projects
		projects := apiV1.Group("/projects")
		{
			projects.POST("", s.handleCreateProject)
			projects.GET("", s.handleListProjects)
			projects.GET("/:id", s.handleGetProject)
			projects.PUT("/:id", s.handleUpdateProject)
			projects.DELETE("/:id", s.handleDeleteProject)
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

func (s *Server) handleGetInstance(c *gin.Context) {
	instanceID := c.Param("id")
	instance, err := s.personaRepo.GetInstanceByID(c.Request.Context(), instanceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}
	c.JSON(http.StatusOK, instance)
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

// Project Handlers (placeholders)

func (s *Server) handleCreateProject(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (s *Server) handleListProjects(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"projects": []interface{}{}})
}

func (s *Server) handleGetProject(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (s *Server) handleUpdateProject(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (s *Server) handleDeleteProject(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}
