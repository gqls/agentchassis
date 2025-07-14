// FILE: internal/core-manager/admin/agent_handlers.go
package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// AgentHandlers manages agent-related admin operations
type AgentHandlers struct {
	clientsDB     *pgxpool.Pool
	templatesDB   *pgxpool.Pool
	kafkaProducer kafka.Producer
	logger        *zap.Logger
}

// NewAgentHandlers creates new agent management handlers
func NewAgentHandlers(clientsDB, templatesDB *pgxpool.Pool, kafkaProducer kafka.Producer, logger *zap.Logger) *AgentHandlers {
	return &AgentHandlers{
		clientsDB:     clientsDB,
		templatesDB:   templatesDB,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}
}

// AgentDefinitionRequest for creating/updating agent definitions
type AgentDefinitionRequest struct {
	Type          string                 `json:"type" binding:"required"`
	DisplayName   string                 `json:"display_name" binding:"required"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category" binding:"required,oneof=data-driven code-driven adapter"`
	DefaultConfig map[string]interface{} `json:"default_config"`
}

// AgentInstanceDetails provides detailed info about an agent instance
type AgentInstanceDetails struct {
	ID           string                 `json:"id"`
	TemplateID   string                 `json:"template_id"`
	TemplateName string                 `json:"template_name"`
	ClientID     string                 `json:"client_id"`
	OwnerUserID  string                 `json:"owner_user_id"`
	Name         string                 `json:"name"`
	Config       map[string]interface{} `json:"config"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Usage        AgentUsageStats        `json:"usage"`
	Health       AgentHealthStatus      `json:"health"`
}

// AgentUsageStats tracks usage metrics for an agent
type AgentUsageStats struct {
	TotalExecutions   int        `json:"total_executions"`
	SuccessfulRuns    int        `json:"successful_runs"`
	FailedRuns        int        `json:"failed_runs"`
	AverageRunTime    float64    `json:"average_run_time_ms"`
	LastExecutionTime *time.Time `json:"last_execution_time,omitempty"`
	FuelConsumed      int64      `json:"fuel_consumed"`
}

// AgentHealthStatus represents current health of an agent
type AgentHealthStatus struct {
	Status        string    `json:"status"` // healthy, degraded, unhealthy
	LastCheckTime time.Time `json:"last_check_time"`
	ErrorRate     float64   `json:"error_rate_percent"`
	ResponseTime  float64   `json:"avg_response_time_ms"`
	QueueDepth    int       `json:"queue_depth"`
}

// HandleCreateAgentDefinition creates a new agent type
func (h *AgentHandlers) HandleCreateAgentDefinition(c *gin.Context) {
	var req AgentDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if type already exists
	var exists bool
	err := h.clientsDB.QueryRow(c.Request.Context(),
		"SELECT EXISTS(SELECT 1 FROM agent_definitions WHERE type = $1 AND deleted_at IS NULL)",
		req.Type,
	).Scan(&exists)

	if err != nil {
		h.logger.Error("Failed to check agent existence", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check agent existence"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Agent type already exists"})
		return
	}

	// Insert new agent definition
	id := uuid.New()
	configBytes, _ := json.Marshal(req.DefaultConfig)

	query := `
		INSERT INTO agent_definitions 
		(id, type, display_name, description, category, default_config, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, true)
	`

	_, err = h.clientsDB.Exec(c.Request.Context(), query,
		id, req.Type, req.DisplayName, req.Description, req.Category, configBytes,
	)

	if err != nil {
		h.logger.Error("Failed to create agent definition", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent definition"})
		return
	}

	// Create Kafka topics for the new agent type
	h.createAgentTopics(c.Request.Context(), req.Type)

	h.logger.Info("Agent definition created",
		zap.String("id", id.String()),
		zap.String("type", req.Type))

	c.JSON(http.StatusCreated, gin.H{
		"id":      id.String(),
		"type":    req.Type,
		"message": "Agent definition created successfully",
	})
}

// HandleListAgentInstances lists all agent instances with filtering
func (h *AgentHandlers) HandleListAgentInstances(c *gin.Context) {
	clientID := c.Query("client_id")
	agentType := c.Query("agent_type")
	isActive := c.Query("is_active")

	instances := []AgentInstanceDetails{}

	// Get all client schemas
	schemaRows, err := h.clientsDB.Query(c.Request.Context(), `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name LIKE 'client_%'
	`)
	if err != nil {
		h.logger.Error("Failed to list schemas", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list instances"})
		return
	}
	defer schemaRows.Close()

	for schemaRows.Next() {
		var schemaName string
		if err := schemaRows.Scan(&schemaName); err != nil {
			continue
		}

		currentClientID := strings.TrimPrefix(schemaName, "client_")
		if clientID != "" && currentClientID != clientID {
			continue
		}

		// Query instances from this client schema
		query := fmt.Sprintf(`
			SELECT 
				ai.id, ai.template_id, ai.owner_user_id, ai.name, 
				ai.config, ai.is_active, ai.created_at, ai.updated_at,
				pt.name as template_name
			FROM %s.agent_instances ai
			LEFT JOIN persona_templates pt ON ai.template_id = pt.id
			WHERE 1=1
		`, schemaName)

		args := []interface{}{}
		argCount := 0

		if agentType != "" {
			argCount++
			query += fmt.Sprintf(" AND pt.category = $%d", argCount)
			args = append(args, agentType)
		}

		if isActive != "" {
			argCount++
			query += fmt.Sprintf(" AND ai.is_active = $%d", argCount)
			args = append(args, isActive == "true")
		}

		rows, err := h.clientsDB.Query(c.Request.Context(), query, args...)
		if err != nil {
			h.logger.Error("Failed to query instances", zap.Error(err))
			continue
		}

		for rows.Next() {
			var instance AgentInstanceDetails
			var configJSON []byte
			var templateName sql.NullString

			err := rows.Scan(
				&instance.ID, &instance.TemplateID, &instance.OwnerUserID,
				&instance.Name, &configJSON, &instance.IsActive,
				&instance.CreatedAt, &instance.UpdatedAt, &templateName,
			)
			if err != nil {
				continue
			}

			instance.ClientID = currentClientID
			if templateName.Valid {
				instance.TemplateName = templateName.String
			}
			json.Unmarshal(configJSON, &instance.Config)

			// Get usage stats
			instance.Usage = h.getAgentUsageStats(c.Request.Context(), currentClientID, instance.ID)
			instance.Health = h.getAgentHealth(c.Request.Context(), instance.ID)

			instances = append(instances, instance)
		}
		rows.Close()
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": instances,
		"count":     len(instances),
	})
}

// HandleGetAgentInstance returns detailed information about a specific agent
func (h *AgentHandlers) HandleGetAgentInstance(c *gin.Context) {
	agentID := c.Param("agent_id")
	clientID := c.Param("client_id")

	if clientID == "" {
		// Need to find which client owns this agent
		clientID = h.findClientForAgent(c.Request.Context(), agentID)
		if clientID == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
	}

	query := fmt.Sprintf(`
		SELECT 
			ai.id, ai.template_id, ai.owner_user_id, ai.name, 
			ai.config, ai.is_active, ai.created_at, ai.updated_at,
			pt.name as template_name
		FROM client_%s.agent_instances ai
		LEFT JOIN persona_templates pt ON ai.template_id = pt.id
		WHERE ai.id = $1
	`, clientID)

	var instance AgentInstanceDetails
	var configJSON []byte
	var templateName sql.NullString

	err := h.clientsDB.QueryRow(c.Request.Context(), query, agentID).Scan(
		&instance.ID, &instance.TemplateID, &instance.OwnerUserID,
		&instance.Name, &configJSON, &instance.IsActive,
		&instance.CreatedAt, &instance.UpdatedAt, &templateName,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	instance.ClientID = clientID
	if templateName.Valid {
		instance.TemplateName = templateName.String
	}
	json.Unmarshal(configJSON, &instance.Config)

	// Get detailed usage and health
	instance.Usage = h.getAgentUsageStats(c.Request.Context(), clientID, agentID)
	instance.Health = h.getAgentHealth(c.Request.Context(), agentID)

	// Get recent executions
	executions := h.getRecentExecutions(c.Request.Context(), clientID, agentID, 10)

	c.JSON(http.StatusOK, gin.H{
		"agent":      instance,
		"executions": executions,
	})
}

// HandleToggleAgentStatus enables/disables an agent instance
func (h *AgentHandlers) HandleToggleAgentStatus(c *gin.Context) {
	agentID := c.Param("agent_id")
	clientID := c.Param("client_id")

	if clientID == "" {
		clientID = h.findClientForAgent(c.Request.Context(), agentID)
		if clientID == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
	}

	var req struct {
		IsActive bool   `json:"is_active"`
		Reason   string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := fmt.Sprintf(`
		UPDATE client_%s.agent_instances 
		SET is_active = $2, updated_at = NOW()
		WHERE id = $1
	`, clientID)

	result, err := h.clientsDB.Exec(c.Request.Context(), query, agentID, req.IsActive)
	if err != nil {
		h.logger.Error("Failed to toggle agent status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent status"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Log the action
	h.logger.Info("Agent status toggled",
		zap.String("agent_id", agentID),
		zap.String("client_id", clientID),
		zap.Bool("is_active", req.IsActive),
		zap.String("reason", req.Reason))

	// Send notification if disabling
	if !req.IsActive {
		h.notifyAgentDisabled(c.Request.Context(), clientID, agentID, req.Reason)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Agent %s successfully", map[bool]string{true: "enabled", false: "disabled"}[req.IsActive]),
		"agent_id":  agentID,
		"is_active": req.IsActive,
	})
}

// HandleRestartAgent sends a restart command to an agent
func (h *AgentHandlers) HandleRestartAgent(c *gin.Context) {
	agentID := c.Param("agent_id")

	// Send restart command via Kafka
	command := map[string]interface{}{
		"command":   "restart",
		"agent_id":  agentID,
		"timestamp": time.Now().UTC(),
	}

	commandBytes, _ := json.Marshal(command)
	headers := map[string]string{
		"correlation_id": uuid.NewString(),
		"command_type":   "agent_restart",
		"agent_id":       agentID,
	}

	err := h.kafkaProducer.Produce(c.Request.Context(),
		"system.agent.commands",
		headers,
		[]byte(agentID),
		commandBytes,
	)

	if err != nil {
		h.logger.Error("Failed to send restart command", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send restart command"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Restart command sent",
		"agent_id": agentID,
	})
}

// Helper methods

func (h *AgentHandlers) createAgentTopics(ctx context.Context, agentType string) {
	topics := []string{
		fmt.Sprintf("tasks.high.%s", agentType),
		fmt.Sprintf("tasks.normal.%s", agentType),
		fmt.Sprintf("tasks.low.%s", agentType),
		fmt.Sprintf("responses.%s", agentType),
		fmt.Sprintf("dlq.%s", agentType),
	}

	for _, topic := range topics {
		// In production, you'd use Kafka admin client
		h.logger.Info("Would create Kafka topic", zap.String("topic", topic))
	}
}

func (h *AgentHandlers) findClientForAgent(ctx context.Context, agentID string) string {
	// Search through all client schemas
	rows, err := h.clientsDB.Query(ctx, `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name LIKE 'client_%'
	`)
	if err != nil {
		return ""
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			continue
		}

		var exists bool
		query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s.agent_instances WHERE id = $1)", schemaName)
		if err := h.clientsDB.QueryRow(ctx, query, agentID).Scan(&exists); err == nil && exists {
			return strings.TrimPrefix(schemaName, "client_")
		}
	}

	return ""
}

func (h *AgentHandlers) getAgentUsageStats(ctx context.Context, clientID, agentID string) AgentUsageStats {
	stats := AgentUsageStats{}

	// Get execution stats from workflow_executions
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'COMPLETED') as successful,
			COUNT(*) FILTER (WHERE status = 'FAILED') as failed,
			AVG(EXTRACT(EPOCH FROM (completed_at - started_at)) * 1000) as avg_runtime,
			MAX(completed_at) as last_execution
		FROM client_%s.workflow_executions
		WHERE agent_instance_id = $1
	`, clientID)

	var lastExecution sql.NullTime
	h.clientsDB.QueryRow(ctx, query, agentID).Scan(
		&stats.TotalExecutions,
		&stats.SuccessfulRuns,
		&stats.FailedRuns,
		&stats.AverageRunTime,
		&lastExecution,
	)

	if lastExecution.Valid {
		stats.LastExecutionTime = &lastExecution.Time
	}

	// Get fuel consumption
	fuelQuery := fmt.Sprintf(`
		SELECT COALESCE(SUM(fuel_consumed), 0)
		FROM client_%s.usage_analytics
		WHERE metadata->>'agent_id' = $1
	`, clientID)
	h.clientsDB.QueryRow(ctx, fuelQuery, agentID).Scan(&stats.FuelConsumed)

	return stats
}

func (h *AgentHandlers) getAgentHealth(ctx context.Context, agentID string) AgentHealthStatus {
	// In production, this would query your metrics system
	return AgentHealthStatus{
		Status:        "healthy",
		LastCheckTime: time.Now(),
		ErrorRate:     2.5,
		ResponseTime:  145.3,
		QueueDepth:    3,
	}
}

func (h *AgentHandlers) getRecentExecutions(ctx context.Context, clientID, agentID string, limit int) []map[string]interface{} {
	executions := []map[string]interface{}{}

	query := fmt.Sprintf(`
		SELECT 
			id, correlation_id, status, started_at, completed_at,
			input_data, output_data, error_message
		FROM client_%s.workflow_executions
		WHERE agent_instance_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`, clientID)

	rows, err := h.clientsDB.Query(ctx, query, agentID, limit)
	if err != nil {
		return executions
	}
	defer rows.Close()

	for rows.Next() {
		var exec struct {
			ID            string
			CorrelationID string
			Status        string
			StartedAt     time.Time
			CompletedAt   sql.NullTime
			InputData     json.RawMessage
			OutputData    json.RawMessage
			ErrorMessage  sql.NullString
		}

		if err := rows.Scan(&exec.ID, &exec.CorrelationID, &exec.Status,
			&exec.StartedAt, &exec.CompletedAt, &exec.InputData,
			&exec.OutputData, &exec.ErrorMessage); err != nil {
			continue
		}

		execution := map[string]interface{}{
			"id":             exec.ID,
			"correlation_id": exec.CorrelationID,
			"status":         exec.Status,
			"started_at":     exec.StartedAt,
			"duration_ms":    nil,
		}

		if exec.CompletedAt.Valid {
			duration := exec.CompletedAt.Time.Sub(exec.StartedAt).Milliseconds()
			execution["completed_at"] = exec.CompletedAt.Time
			execution["duration_ms"] = duration
		}

		if exec.ErrorMessage.Valid {
			execution["error"] = exec.ErrorMessage.String
		}

		executions = append(executions, execution)
	}

	return executions
}

func (h *AgentHandlers) notifyAgentDisabled(ctx context.Context, clientID, agentID, reason string) {
	notification := map[string]interface{}{
		"event_type": "AGENT_DISABLED",
		"client_id":  clientID,
		"agent_id":   agentID,
		"reason":     reason,
		"timestamp":  time.Now().UTC(),
	}

	notificationBytes, _ := json.Marshal(notification)
	headers := map[string]string{
		"correlation_id": uuid.NewString(),
		"event_type":     "agent_disabled",
	}

	h.kafkaProducer.Produce(ctx, "system.notifications.admin", headers,
		[]byte(agentID), notificationBytes)
}
