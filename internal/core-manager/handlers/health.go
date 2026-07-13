package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/platform/config"
	"go.uber.org/zap"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	cfg    *config.ServiceConfig
	logger *zap.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(cfg *config.ServiceConfig, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		cfg:    cfg,
		logger: logger,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status" example:"healthy"`
	Service string `json:"service" example:"core-manager"`
	Version string `json:"version" example:"1.0.0"`
}

// HandleHealth returns the service health status
func (h *HealthHandler) HandleHealth(c *gin.Context) {
	response := HealthResponse{
		Status:  "healthy",
		Service: h.cfg.ServiceInfo.Name,
		Version: h.cfg.ServiceInfo.Version,
	}
	c.JSON(http.StatusOK, response)
}
