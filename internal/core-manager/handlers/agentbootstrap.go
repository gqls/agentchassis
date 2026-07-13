package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	logger    *zap.Logger
	clientsDB *sql.DB
}

type BootstrapRequest struct {
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	ClientID  string `json:"client_id"`
}

type BootstrapResponse struct {
	Message string                 `json:"message"`
	Token   string                 `json:"token,omitempty"`
	Config  map[string]interface{} `json:"config,omitempty"`
}

func NewBootstrapHandler(logger *zap.Logger, clientsDB *sql.DB) *Handler {
	return &Handler{
		logger:    logger,
		clientsDB: clientsDB,
	}
}

func (h *Handler) HandleAgentBootstrap(c *gin.Context) {
	// 1. Validate bootstrap key
	bootstrapKey := c.GetHeader("X-Bootstrap-Key")
	if bootstrapKey == "" {
		h.logger.Warn("Agent bootstrap attempt failed: X-Bootstrap-Key header is missing")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: X-Bootstrap-Key header required"})
		return
	}

	expectedKey := os.Getenv("AGENT_BOOTSTRAP_KEY")
	if expectedKey == "" {
		h.logger.Error("AGENT_BOOTSTRAP_KEY environment variable is not set")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if bootstrapKey != expectedKey {
		h.logger.Warn("Agent bootstrap attempt failed: Invalid bootstrap key")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid bootstrap key"})
		return
	}

	// 2. Parse request body
	var req BootstrapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid bootstrap request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 3. Update agent status in database
	if err := h.updateAgentStatus(req.ClientID, req.AgentID, "running"); err != nil {
		h.logger.Error("Failed to update agent status",
			zap.String("agent_id", req.AgentID),
			zap.Error(err))
		// Don't fail bootstrap - just log
	}

	// 4. Record bootstrap event
	if err := h.recordBootstrapEvent(req); err != nil {
		h.logger.Error("Failed to record bootstrap event",
			zap.String("agent_id", req.AgentID),
			zap.Error(err))
		// Don't fail bootstrap - just log
	}

	h.logger.Info("Agent bootstrap successful",
		zap.String("agent_id", req.AgentID),
		zap.String("agent_type", req.AgentType),
		zap.String("client_id", req.ClientID))

	// 5. Return success with optional config
	resp := BootstrapResponse{
		Message: "Bootstrap successful",
		// In future, could generate agent-specific JWT here
		// Token: generateAgentJWT(req.AgentID),
	}

	// Add any dynamic configuration for the agent
	if config := h.getAgentDynamicConfig(req.AgentType); config != nil {
		resp.Config = config
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) updateAgentStatus(clientID, agentID, status string) error {
	query := fmt.Sprintf(`
        UPDATE client_%s.agent_instances 
        SET last_seen_at = NOW(),
            runtime_status = $1,
            config = jsonb_set(
                COALESCE(config, '{}'::jsonb),
                '{runtime_info,last_bootstrap}',
                to_jsonb(NOW()::text)
            )
        WHERE id = $2
    `, clientID)

	_, err := h.clientsDB.ExecContext(context.Background(), query, status, agentID)
	return err
}

func (h *Handler) recordBootstrapEvent(req BootstrapRequest) error {
	query := `
        INSERT INTO system_events (
            event_type, 
            entity_type, 
            entity_id,
            client_id, 
            metadata,
            severity,
            source,
            created_at
        ) VALUES (
            'agent_bootstrap',
            'agent',
            $1,
            $2,
            $3,
            'info',
            'core-manager',
            NOW()
        )
    `

	metadata, _ := json.Marshal(map[string]interface{}{
		"agent_type": req.AgentType,
		"agent_id":   req.AgentID,
		"timestamp":  time.Now().Format(time.RFC3339),
	})

	_, err := h.clientsDB.ExecContext(context.Background(),
		query,
		req.AgentID,
		req.ClientID,
		metadata)

	if err != nil {
		h.logger.Debug("Failed to record bootstrap event", zap.Error(err))
	}
	return nil
}

func (h *Handler) getAgentDynamicConfig(agentType string) map[string]interface{} {
	// Return any dynamic configuration for this agent type
	// This could be loaded from database or config
	return map[string]interface{}{
		"log_level": "info",
		// Add more config as needed
	}
}
