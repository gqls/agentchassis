// FILE: internal/core-manager/admin/system_handlers.go
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
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// SystemHandlers handles admin system monitoring operations
type SystemHandlers struct {
	clientsDB     *pgxpool.Pool
	templatesDB   *pgxpool.Pool
	kafkaProducer kafka.Producer
	logger        *zap.Logger
}

// NewSystemHandlers creates new system monitoring handlers
func NewSystemHandlers(clientsDB, templatesDB *pgxpool.Pool, kafkaProducer kafka.Producer, logger *zap.Logger) *SystemHandlers {
	return &SystemHandlers{
		clientsDB:     clientsDB,
		templatesDB:   templatesDB,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}
}

// SystemStatus represents overall system health
type SystemStatus struct {
	Status      string                    `json:"status"`
	Timestamp   time.Time                 `json:"timestamp"`
	Services    map[string]ServiceStatus  `json:"services"`
	Databases   map[string]DatabaseStatus `json:"databases"`
	KafkaStatus KafkaStatus               `json:"kafka"`
}

// ServiceStatus represents a service health status
type ServiceStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
	Uptime  string `json:"uptime,omitempty"`
}

// DatabaseStatus represents database health
type DatabaseStatus struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Connections int    `json:"active_connections"`
	Size        string `json:"size,omitempty"`
}

// KafkaStatus represents Kafka cluster status
type KafkaStatus struct {
	Status      string `json:"status"`
	BrokerCount int    `json:"broker_count"`
	TopicCount  int    `json:"topic_count"`
	ConsumerLag int64  `json:"total_consumer_lag,omitempty"`
}

// HandleGetSystemStatus returns aggregated system status
func (h *SystemHandlers) HandleGetSystemStatus(c *gin.Context) {
	status := SystemStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  make(map[string]ServiceStatus),
		Databases: make(map[string]DatabaseStatus),
	}

	// Check database connections
	// Clients DB
	if err := h.clientsDB.Ping(c.Request.Context()); err != nil {
		status.Status = "degraded"
		status.Databases["clients_db"] = DatabaseStatus{
			Name:   "clients_db",
			Status: "unhealthy",
		}
	} else {
		stats := h.clientsDB.Stat()
		status.Databases["clients_db"] = DatabaseStatus{
			Name:        "clients_db",
			Status:      "healthy",
			Connections: int(stats.AcquiredConns()),
		}
	}

	// Templates DB
	if err := h.templatesDB.Ping(c.Request.Context()); err != nil {
		status.Status = "degraded"
		status.Databases["templates_db"] = DatabaseStatus{
			Name:   "templates_db",
			Status: "unhealthy",
		}
	} else {
		stats := h.templatesDB.Stat()
		status.Databases["templates_db"] = DatabaseStatus{
			Name:        "templates_db",
			Status:      "healthy",
			Connections: int(stats.AcquiredConns()),
		}
	}

	// Get database sizes
	h.getDatabaseSizes(c.Request.Context(), &status)

	// Check Kafka (simplified - in production you'd want more detailed checks)
	status.KafkaStatus = h.getKafkaStatus(c.Request.Context())

	c.JSON(http.StatusOK, status)
}

// HandleListKafkaTopics lists all Kafka topics
func (h *SystemHandlers) HandleListKafkaTopics(c *gin.Context) {
	// This would require a Kafka admin client
	// For now, return known topics
	knownTopics := []string{
		"orchestrator.state-changes",
		"human.approvals",
		"system.events",
		"system.notifications.ui",
		"system.commands.workflow.resume",
		"system.agent.reasoning.process",
		"system.responses.reasoning",
		"system.adapter.image.generate",
		"system.responses.image",
		"system.adapter.web.search",
		"system.responses.websearch",
		"system.agent.generic.process",
		"system.tasks.copywriter",
		"system.tasks.researcher",
		"system.tasks.content-creator",
		"system.tasks.multimedia-creator",
	}

	c.JSON(http.StatusOK, gin.H{
		"topics": knownTopics,
		"count":  len(knownTopics),
	})
}

// WorkflowListRequest represents workflow filtering parameters
type WorkflowListRequest struct {
	Status    string `form:"status"`
	ClientID  string `form:"client_id"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Limit     int    `form:"limit,default=50"`
	Offset    int    `form:"offset,default=0"`
}

// HandleListWorkflows lists workflow executions
func (h *SystemHandlers) HandleListWorkflows(c *gin.Context) {
	var req WorkflowListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflows, err := h.listWorkflows(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list workflows", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list workflows"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows": workflows,
		"count":     len(workflows),
	})
}

// HandleGetWorkflow returns detailed workflow state
func (h *SystemHandlers) HandleGetWorkflow(c *gin.Context) {
	correlationID := c.Param("correlation_id")

	workflow, err := h.getWorkflowState(c.Request.Context(), correlationID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
			return
		}
		h.logger.Error("Failed to get workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflow"})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// HandleResumeWorkflow manually resumes or terminates a workflow
func (h *SystemHandlers) HandleResumeWorkflow(c *gin.Context) {
	correlationID := c.Param("correlation_id")

	var req struct {
		Action   string                 `json:"action" binding:"required,oneof=resume terminate"`
		Feedback map[string]interface{} `json:"feedback,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get workflow state to find client_id
	workflow, err := h.getWorkflowState(c.Request.Context(), correlationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	if req.Action == "terminate" {
		// Update workflow status to FAILED
		err = h.updateWorkflowStatus(c.Request.Context(), correlationID, "FAILED", "Manually terminated by admin")
		if err != nil {
			h.logger.Error("Failed to terminate workflow", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to terminate workflow"})
			return
		}
	} else if req.Action == "resume" {
		// Send resume command via Kafka
		resumePayload := map[string]interface{}{
			"approved": true,
			"feedback": req.Feedback,
		}

		payloadBytes, _ := json.Marshal(resumePayload)
		headers := map[string]string{
			"correlation_id": correlationID,
			"client_id":      workflow["client_id"].(string),
			"admin_action":   "true",
		}

		err = h.kafkaProducer.Produce(c.Request.Context(),
			orchestration.ResumeWorkflowTopic,
			headers,
			[]byte(correlationID),
			payloadBytes,
		)

		if err != nil {
			h.logger.Error("Failed to send resume command", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resume workflow"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        fmt.Sprintf("Workflow %s successfully", req.Action+"d"),
		"correlation_id": correlationID,
	})
}

// HandleListAgentDefinitions lists all agent types
func (h *SystemHandlers) HandleListAgentDefinitions(c *gin.Context) {
	query := `
		SELECT id, type, display_name, description, category, default_config, is_active
		FROM agent_definitions
		WHERE deleted_at IS NULL
		ORDER BY category, type
	`

	rows, err := h.clientsDB.Query(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list agent definitions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list agent definitions"})
		return
	}
	defer rows.Close()

	var definitions []map[string]interface{}
	for rows.Next() {
		var def struct {
			ID            string
			Type          string
			DisplayName   string
			Description   string
			Category      string
			DefaultConfig json.RawMessage
			IsActive      bool
		}

		err := rows.Scan(&def.ID, &def.Type, &def.DisplayName,
			&def.Description, &def.Category, &def.DefaultConfig, &def.IsActive)
		if err != nil {
			h.logger.Error("Failed to scan agent definition", zap.Error(err))
			continue
		}

		var config map[string]interface{}
		json.Unmarshal(def.DefaultConfig, &config)

		definitions = append(definitions, map[string]interface{}{
			"id":             def.ID,
			"type":           def.Type,
			"display_name":   def.DisplayName,
			"description":    def.Description,
			"category":       def.Category,
			"default_config": config,
			"is_active":      def.IsActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"definitions": definitions,
		"count":       len(definitions),
	})
}

// HandleUpdateAgentDefinition updates an agent type configuration
func (h *SystemHandlers) HandleUpdateAgentDefinition(c *gin.Context) {
	typeName := c.Param("type_name")

	var req struct {
		DisplayName   *string                `json:"display_name"`
		Description   *string                `json:"description"`
		DefaultConfig map[string]interface{} `json:"default_config"`
		IsActive      *bool                  `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update query
	updates := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argCount := 0

	if req.DisplayName != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("display_name = $%d", argCount))
		args = append(args, *req.DisplayName)
	}

	if req.Description != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("description = $%d", argCount))
		args = append(args, *req.Description)
	}

	if req.DefaultConfig != nil {
		argCount++
		configBytes, _ := json.Marshal(req.DefaultConfig)
		updates = append(updates, fmt.Sprintf("default_config = $%d", argCount))
		args = append(args, configBytes)
	}

	if req.IsActive != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("is_active = $%d", argCount))
		args = append(args, *req.IsActive)
	}

	argCount++
	args = append(args, typeName)

	query := fmt.Sprintf(
		"UPDATE agent_definitions SET %s WHERE type = $%d AND deleted_at IS NULL",
		strings.Join(updates, ", "),
		argCount,
	)

	result, err := h.clientsDB.Exec(c.Request.Context(), query, args...)
	if err != nil {
		h.logger.Error("Failed to update agent definition", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent definition"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent definition not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent definition updated successfully",
		"type":    typeName,
	})
}

// Helper methods

func (h *SystemHandlers) getDatabaseSizes(ctx context.Context, status *SystemStatus) {
	// Get clients DB size
	var clientsSize string
	err := h.clientsDB.QueryRow(ctx,
		"SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&clientsSize)
	if err == nil {
		dbStatus := status.Databases["clients_db"]
		dbStatus.Size = clientsSize
		status.Databases["clients_db"] = dbStatus
	}

	// Get templates DB size
	var templatesSize string
	err = h.templatesDB.QueryRow(ctx,
		"SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&templatesSize)
	if err == nil {
		dbStatus := status.Databases["templates_db"]
		dbStatus.Size = templatesSize
		status.Databases["templates_db"] = dbStatus
	}
}

func (h *SystemHandlers) getKafkaStatus(ctx context.Context) KafkaStatus {
	// In a real implementation, you'd use a Kafka admin client
	// For now, return a simplified status
	return KafkaStatus{
		Status:      "healthy",
		BrokerCount: 3,  // Based on the k8s config
		TopicCount:  20, // Approximate
	}
}

func (h *SystemHandlers) listWorkflows(ctx context.Context, req WorkflowListRequest) ([]map[string]interface{}, error) {
	query := `
		SELECT correlation_id, client_id, status, current_step, 
		       created_at, updated_at, error
		FROM orchestrator_state
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	if req.Status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}

	if req.ClientID != "" {
		argCount++
		query += fmt.Sprintf(" AND client_id = $%d", argCount)
		args = append(args, req.ClientID)
	}

	if req.StartDate != "" {
		argCount++
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		argCount++
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, req.EndDate)
	}

	query += " ORDER BY created_at DESC"

	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, req.Limit)

	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, req.Offset)

	rows, err := h.clientsDB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []map[string]interface{}
	for rows.Next() {
		var w struct {
			CorrelationID string
			ClientID      string
			Status        string
			CurrentStep   string
			CreatedAt     time.Time
			UpdatedAt     time.Time
			Error         sql.NullString
		}

		err := rows.Scan(&w.CorrelationID, &w.ClientID, &w.Status,
			&w.CurrentStep, &w.CreatedAt, &w.UpdatedAt, &w.Error)
		if err != nil {
			continue
		}

		workflow := map[string]interface{}{
			"correlation_id": w.CorrelationID,
			"client_id":      w.ClientID,
			"status":         w.Status,
			"current_step":   w.CurrentStep,
			"created_at":     w.CreatedAt,
			"updated_at":     w.UpdatedAt,
		}

		if w.Error.Valid {
			workflow["error"] = w.Error.String
		}

		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

func (h *SystemHandlers) getWorkflowState(ctx context.Context, correlationID string) (map[string]interface{}, error) {
	var state orchestration.OrchestrationState
	var awaitedStepsJSON, collectedDataJSON, initialRequestDataJSON, finalResultJSON []byte
	var errorNull sql.NullString

	query := `
		SELECT correlation_id, client_id, status, current_step, awaited_steps, 
		       collected_data, initial_request_data, final_result, error, 
		       created_at, updated_at
		FROM orchestrator_state
		WHERE correlation_id = $1
	`

	err := h.clientsDB.QueryRow(ctx, query, correlationID).Scan(
		&state.CorrelationID,
		&state.ClientID,
		&state.Status,
		&state.CurrentStep,
		&awaitedStepsJSON,
		&collectedDataJSON,
		&initialRequestDataJSON,
		&finalResultJSON,
		&errorNull,
		&state.CreatedAt,
		&state.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal(awaitedStepsJSON, &state.AwaitedSteps)
	json.Unmarshal(collectedDataJSON, &state.CollectedData)
	state.InitialRequestData = initialRequestDataJSON
	state.FinalResult = finalResultJSON

	if errorNull.Valid {
		state.Error = errorNull.String
	}

	// Convert to map for response
	result := map[string]interface{}{
		"correlation_id":       state.CorrelationID,
		"client_id":            state.ClientID,
		"status":               state.Status,
		"current_step":         state.CurrentStep,
		"awaited_steps":        state.AwaitedSteps,
		"collected_data":       state.CollectedData,
		"initial_request_data": json.RawMessage(state.InitialRequestData),
		"final_result":         json.RawMessage(state.FinalResult),
		"error":                state.Error,
		"created_at":           state.CreatedAt,
		"updated_at":           state.UpdatedAt,
	}

	return result, nil
}

func (h *SystemHandlers) updateWorkflowStatus(ctx context.Context, correlationID, status, errorMsg string) error {
	query := `
		UPDATE orchestrator_state 
		SET status = $2, error = $3, updated_at = NOW()
		WHERE correlation_id = $1
	`

	_, err := h.clientsDB.Exec(ctx, query, correlationID, status, errorMsg)
	return err
}
