package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/core-manager/middleware"
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// InstanceHandler handles persona instance operations
type InstanceHandler struct {
	personaRepo models.PersonaRepository
	logger      *zap.Logger
}

// NewInstanceHandler creates a new instance handler
func NewInstanceHandler(personaRepo models.PersonaRepository, logger *zap.Logger) *InstanceHandler {
	return &InstanceHandler{
		personaRepo: personaRepo,
		logger:      logger,
	}
}

// CreateInstanceRequest represents a request to create an instance
type CreateInstanceRequest struct {
	TemplateID   string `json:"template_id" binding:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	InstanceName string `json:"instance_name" binding:"required" example:"My Support Agent"`
}

// UpdateInstanceRequest represents a request to update an instance
type UpdateInstanceRequest struct {
	Name   *string                `json:"name,omitempty" example:"Updated Agent Name"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// InstanceListResponse represents a list of instances
type InstanceListResponse struct {
	Instances []models.Persona `json:"instances"`
	Count     int              `json:"count" example:"5"`
}

// HandleCreateInstance creates a new instance from a template
func (h *InstanceHandler) HandleCreateInstance(c *gin.Context) {
	claims := c.MustGet("user_claims").(*middleware.AuthClaims)

	var req CreateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	instance, err := h.personaRepo.CreateInstanceFromTemplate(c.Request.Context(),
		req.TemplateID, claims.UserID, req.InstanceName)
	if err != nil {
		h.logger.Error("Failed to create instance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create instance"})
		return
	}
	c.JSON(http.StatusCreated, instance)
}

// HandleGetInstance returns a specific instance
func (h *InstanceHandler) HandleGetInstance(c *gin.Context) {
	instanceID := c.Param("id")
	instance, err := h.personaRepo.GetInstanceByID(c.Request.Context(), instanceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}
	c.JSON(http.StatusOK, instance)
}

// HandleListInstances returns all instances for the current user
func (h *InstanceHandler) HandleListInstances(c *gin.Context) {
	claims := c.MustGet("user_claims").(*middleware.AuthClaims)
	instances, err := h.personaRepo.ListInstances(c.Request.Context(), claims.UserID)
	if err != nil {
		h.logger.Error("Failed to list instances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve instances"})
		return
	}

	response := InstanceListResponse{
		Instances: instances,
		Count:     len(instances),
	}
	c.JSON(http.StatusOK, response)
}

// HandleUpdateInstance updates an existing instance
func (h *InstanceHandler) HandleUpdateInstance(c *gin.Context) {
	instanceID := c.Param("id")

	var req UpdateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updatedInstance, err := h.personaRepo.UpdateInstance(c.Request.Context(), instanceID, req.Name, req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update instance"})
		return
	}
	c.JSON(http.StatusOK, updatedInstance)
}

// HandleDeleteInstance deletes an instance
func (h *InstanceHandler) HandleDeleteInstance(c *gin.Context) {
	instanceID := c.Param("id")
	if err := h.personaRepo.DeleteInstance(c.Request.Context(), instanceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete instance"})
		return
	}
	c.Status(http.StatusNoContent)
}
