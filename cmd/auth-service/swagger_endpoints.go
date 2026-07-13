package main

// healthCheck godoc
// @Summary      Health check
// @Description  Check if the auth service is running and healthy
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]interface{} "Service is healthy"
// @Failure      503 {object} map[string]interface{} "Service unavailable"
// @Router       /health [get]
// @ID           healthCheck

// handleWebSocket godoc
// @Summary      WebSocket connection
// @Description  Establish a WebSocket connection for real-time communication
// @Tags         WebSocket
// @Success      101 {string} string "Switching Protocols"
// @Failure      400 {object} map[string]interface{} "Bad request"
// @Failure      401 {object} map[string]interface{} "Unauthorized - no valid token"
// @Failure      426 {object} map[string]interface{} "Upgrade required"
// @Router       /ws [get]
// @Security     Bearer
// @ID           websocketConnect
