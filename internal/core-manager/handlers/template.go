// FILE: internal/core-manager/handlers/template.go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// TemplateHandler handles template-related operations
type TemplateHandler struct {
	personaRepo models.PersonaRepository
	logger      *zap.Logger
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(personaRepo models.PersonaRepository, logger *zap.Logger) *TemplateHandler {
	return &TemplateHandler{
		personaRepo: personaRepo,
		logger:      logger,
	}
}

// CreateTemplateRequest represents a request to create a template
type CreateTemplateRequest struct {
	Name        string                 `json:"name" binding:"required" example:"Customer Support Agent"`
	Description string                 `json:"description" example:"A helpful customer support agent template"`
	Category    string                 `json:"category" binding:"required" example:"support"`
	Config      map[string]interface{} `json:"config" binding:"required"`
}

// TemplateListResponse represents a list of templates
type TemplateListResponse struct {
	Templates []models.Persona `json:"templates"`
	Count     int              `json:"count" example:"10"`
}

// HandleCreateTemplate creates a new template
func (h *TemplateHandler) HandleCreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	template := &models.Persona{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Config:      req.Config,
		IsTemplate:  true,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	createdTemplate, err := h.personaRepo.CreateTemplate(c.Request.Context(), template)
	if err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, createdTemplate)
}

// HandleListTemplates returns all templates
func (h *TemplateHandler) HandleListTemplates(c *gin.Context) {
	templates, err := h.personaRepo.ListTemplates(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve templates"})
		return
	}

	response := TemplateListResponse{
		Templates: templates,
		Count:     len(templates),
	}
	c.JSON(http.StatusOK, response)
}

// HandleGetTemplate returns a specific template
func (h *TemplateHandler) HandleGetTemplate(c *gin.Context) {
	templateID := c.Param("id")
	template, err := h.personaRepo.GetTemplateByID(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	c.JSON(http.StatusOK, template)
}

// HandleUpdateTemplate updates an existing template
func (h *TemplateHandler) HandleUpdateTemplate(c *gin.Context) {
	templateID := c.Param("id")

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	template := &models.Persona{
		ID:          uuid.MustParse(templateID),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Config:      req.Config,
		UpdatedAt:   time.Now(),
	}

	updatedTemplate, err := h.personaRepo.UpdateTemplate(c.Request.Context(), template)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		return
	}
	c.JSON(http.StatusOK, updatedTemplate)
}

// HandleDeleteTemplate deletes a template
func (h *TemplateHandler) HandleDeleteTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if err := h.personaRepo.DeleteTemplate(c.Request.Context(), templateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}
	c.Status(http.StatusNoContent)
}
