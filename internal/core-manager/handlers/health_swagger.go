package handlers

// NOTE: This file contains swagger annotations for the health handler.
// All types are defined in health.go

// HandleHealth godoc
// @Summary      Health check
// @Description  Returns the health status of the Core Manager service
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200 {object} handlers.HealthResponse "Service is healthy"
// @Router       /health [get]
// @ID           healthCheck
