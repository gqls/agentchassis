// FILE: platform/orchestration/actions/hitl_persistence.go
// Persistence functions for HITL (Human-in-the-Loop) input requests

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// InputRequest represents a persisted HITL request
type InputRequest struct {
	RequestID       string                 `json:"request_id"`
	OrchestrationID string                 `json:"orchestration_id"`
	CorrelationID   string                 `json:"correlation_id"`
	StepID          string                 `json:"step_id"`
	StepName        string                 `json:"step_name"`
	RequestType     string                 `json:"request_type"`
	AgentType       string                 `json:"agent_type"`
	AgentID         string                 `json:"agent_id"`
	Title           string                 `json:"title"`
	Message         string                 `json:"message"`
	Data            map[string]interface{} `json:"data"`
	UIConfig        map[string]interface{} `json:"ui_config"`
	ReplyToTopic    string                 `json:"reply_to_topic"`
	TimeoutSeconds  int                    `json:"timeout_seconds"`
	CreatedAt       time.Time              `json:"created_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	Status          string                 `json:"status"`
	Response        map[string]interface{} `json:"response,omitempty"`
	RespondedBy     string                 `json:"responded_by,omitempty"`
	RespondedAt     *time.Time             `json:"responded_at,omitempty"`
}

// storeInputRequest persists a HITL request to the database
func storeInputRequest(
	ctx context.Context,
	db interface{},
	requestID string,
	requestType string,
	execCtx *types.ExecutionContext,
	data map[string]interface{},
	logger *zap.Logger,
) error {
	sqlDB, ok := db.(*sql.DB)
	if !ok || sqlDB == nil {
		logger.Warn("storeInputRequest: Database not available, request will not be persisted",
			zap.String("request_id", requestID))
		return nil // Not a fatal error - Kafka flow still works
	}

	// Extract optional fields from data
	title := ""
	if t, ok := data["title"].(string); ok {
		title = t
	}
	message := ""
	if m, ok := data["message"].(string); ok {
		message = m
	}

	// Get timeout from data or use default
	timeoutSeconds := 3600
	if t, ok := data["timeout_seconds"].(int); ok {
		timeoutSeconds = t
	} else if t, ok := data["timeout_seconds"].(float64); ok {
		timeoutSeconds = int(t)
	}

	// Calculate expiry
	expiresAt := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)

	// Extract UI config if present
	var uiConfigJSON []byte
	if uiConfig, ok := data["ui_config"].(map[string]interface{}); ok {
		uiConfigJSON, _ = json.Marshal(uiConfig)
	}

	// Extract reply_to_topic
	replyToTopic := ""
	if rt, ok := data["reply_to_topic"].(string); ok {
		replyToTopic = rt
	} else if execCtx != nil && execCtx.ResponsesTopic != "" {
		replyToTopic = execCtx.ResponsesTopic
	}

	// Marshal the full data
	dataJSON, err := json.Marshal(data)
	if err != nil {
		logger.Error("storeInputRequest: Failed to marshal data",
			zap.String("request_id", requestID),
			zap.Error(err))
		return err
	}

	// Get agent info from execution context
	agentType := ""
	agentID := ""
	orchestrationID := ""
	correlationID := ""
	stepID := ""
	stepName := ""

	if execCtx != nil {
		if execCtx.Sender.AgentType != "" {
			agentType = execCtx.Sender.AgentType
		}
		if execCtx.Sender.AgentID != "" {
			agentID = execCtx.Sender.AgentID
		}
		orchestrationID = execCtx.OrchestrationID
		correlationID = execCtx.CorrelationID
		stepID = execCtx.StepID
		stepName = execCtx.StepName
	}

	query := `
		INSERT INTO input_requests (
			request_id, orchestration_id, correlation_id, step_id, step_name,
			request_type, agent_type, agent_id, title, message,
			data, ui_config, reply_to_topic, timeout_seconds,
			created_at, expires_at, status
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14,
			NOW(), $15, 'pending'
		)
		ON CONFLICT (request_id) DO UPDATE SET
			status = 'pending',
			expires_at = $15,
			data = $11
	`

	_, err = sqlDB.ExecContext(ctx, query,
		requestID, // $1
		datahelpers.NullableString(orchestrationID), // $2
		datahelpers.NullableString(correlationID),   // $3
		datahelpers.NullableString(stepID),          // $4
		datahelpers.NullableString(stepName),        // $5
		requestType,                                 // $6
		datahelpers.NullableString(agentType),       // $7
		datahelpers.NullableString(agentID),         // $8
		datahelpers.NullableString(title),           // $9
		datahelpers.NullableString(message),         // $10
		dataJSON,                                    // $11
		nullableJSON(uiConfigJSON),                  // $12
		replyToTopic,                                // $13
		timeoutSeconds,                              // $14
		expiresAt,                                   // $15
	)

	if err != nil {
		logger.Error("storeInputRequest: Failed to insert request",
			zap.String("request_id", requestID),
			zap.Error(err))
		return err
	}

	logger.Info("storeInputRequest: Request persisted",
		zap.String("request_id", requestID),
		zap.String("request_type", requestType),
		zap.String("orchestration_id", orchestrationID),
		zap.String("reply_to_topic", replyToTopic),
		zap.Time("expires_at", expiresAt),
	)

	return nil
}

// updateInputRequest updates a HITL request with the response
func updateInputRequest(
	ctx context.Context,
	db interface{},
	requestID string,
	status string,
	response map[string]interface{},
	logger *zap.Logger,
) error {
	sqlDB, ok := db.(*sql.DB)
	if !ok || sqlDB == nil {
		logger.Warn("updateInputRequest: Database not available",
			zap.String("request_id", requestID))
		return nil
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		logger.Error("updateInputRequest: Failed to marshal response",
			zap.String("request_id", requestID),
			zap.Error(err))
		return err
	}

	// Extract responded_by if present
	respondedBy := ""
	if rb, ok := response["responded_by"].(string); ok {
		respondedBy = rb
	} else if rb, ok := response["user"].(string); ok {
		respondedBy = rb
	} else if rb, ok := response["email"].(string); ok {
		respondedBy = rb
	}

	query := `
		UPDATE input_requests
		SET status = $2,
			response = $3,
			responded_by = $4,
			responded_at = NOW()
		WHERE request_id = $1
	`

	result, err := sqlDB.ExecContext(ctx, query,
		requestID,
		status,
		responseJSON,
		datahelpers.NullableString(respondedBy),
	)

	if err != nil {
		logger.Error("updateInputRequest: Failed to update request",
			zap.String("request_id", requestID),
			zap.Error(err))
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	logger.Info("updateInputRequest: Request updated",
		zap.String("request_id", requestID),
		zap.String("status", status),
		zap.String("responded_by", respondedBy),
		zap.Int64("rows_affected", rowsAffected),
	)

	return nil
}

// GetPendingInputRequests retrieves all pending HITL requests
func GetPendingInputRequests(ctx context.Context, db *sql.DB, logger *zap.Logger) ([]InputRequest, error) {
	query := `
		SELECT 
			request_id, orchestration_id, correlation_id, step_id, step_name,
			request_type, agent_type, agent_id, title, message,
			data, ui_config, reply_to_topic, timeout_seconds,
			created_at, expires_at, status
		FROM input_requests
		WHERE status = 'pending'
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at ASC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("GetPendingInputRequests: Query failed", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var requests []InputRequest
	for rows.Next() {
		var req InputRequest
		var dataJSON, uiConfigJSON []byte
		var orchestrationID, correlationID, stepID, stepName, agentType, agentID sql.NullString
		var title, message sql.NullString

		err := rows.Scan(
			&req.RequestID,
			&orchestrationID,
			&correlationID,
			&stepID,
			&stepName,
			&req.RequestType,
			&agentType,
			&agentID,
			&title,
			&message,
			&dataJSON,
			&uiConfigJSON,
			&req.ReplyToTopic,
			&req.TimeoutSeconds,
			&req.CreatedAt,
			&req.ExpiresAt,
			&req.Status,
		)
		if err != nil {
			logger.Error("GetPendingInputRequests: Scan failed", zap.Error(err))
			continue
		}

		req.OrchestrationID = orchestrationID.String
		req.CorrelationID = correlationID.String
		req.StepID = stepID.String
		req.StepName = stepName.String
		req.AgentType = agentType.String
		req.AgentID = agentID.String
		req.Title = title.String
		req.Message = message.String

		if len(dataJSON) > 0 {
			json.Unmarshal(dataJSON, &req.Data)
		}
		if len(uiConfigJSON) > 0 {
			json.Unmarshal(uiConfigJSON, &req.UIConfig)
		}

		requests = append(requests, req)
	}

	logger.Info("GetPendingInputRequests: Retrieved pending requests",
		zap.Int("count", len(requests)))

	return requests, nil
}

// GetInputRequestByID retrieves a specific HITL request
func GetInputRequestByID(ctx context.Context, db *sql.DB, requestID string, logger *zap.Logger) (*InputRequest, error) {
	query := `
		SELECT 
			request_id, orchestration_id, correlation_id, step_id, step_name,
			request_type, agent_type, agent_id, title, message,
			data, ui_config, reply_to_topic, timeout_seconds,
			created_at, expires_at, status, response, responded_by, responded_at
		FROM input_requests
		WHERE request_id = $1
	`

	var req InputRequest
	var dataJSON, uiConfigJSON, responseJSON []byte
	var orchestrationID, correlationID, stepID, stepName, agentType, agentID sql.NullString
	var title, message, respondedBy sql.NullString
	var respondedAt sql.NullTime

	err := db.QueryRowContext(ctx, query, requestID).Scan(
		&req.RequestID,
		&orchestrationID,
		&correlationID,
		&stepID,
		&stepName,
		&req.RequestType,
		&agentType,
		&agentID,
		&title,
		&message,
		&dataJSON,
		&uiConfigJSON,
		&req.ReplyToTopic,
		&req.TimeoutSeconds,
		&req.CreatedAt,
		&req.ExpiresAt,
		&req.Status,
		&responseJSON,
		&respondedBy,
		&respondedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logger.Error("GetInputRequestByID: Query failed",
			zap.String("request_id", requestID),
			zap.Error(err))
		return nil, err
	}

	req.OrchestrationID = orchestrationID.String
	req.CorrelationID = correlationID.String
	req.StepID = stepID.String
	req.StepName = stepName.String
	req.AgentType = agentType.String
	req.AgentID = agentID.String
	req.Title = title.String
	req.Message = message.String
	req.RespondedBy = respondedBy.String

	if respondedAt.Valid {
		req.RespondedAt = &respondedAt.Time
	}

	if len(dataJSON) > 0 {
		json.Unmarshal(dataJSON, &req.Data)
	}
	if len(uiConfigJSON) > 0 {
		json.Unmarshal(uiConfigJSON, &req.UIConfig)
	}
	if len(responseJSON) > 0 {
		json.Unmarshal(responseJSON, &req.Response)
	}

	return &req, nil
}

// ExpireOldInputRequests marks expired requests as 'expired'
func ExpireOldInputRequests(ctx context.Context, db *sql.DB, logger *zap.Logger) (int64, error) {
	query := `
		UPDATE input_requests
		SET status = 'expired'
		WHERE status = 'pending'
		  AND expires_at < NOW()
	`

	result, err := db.ExecContext(ctx, query)
	if err != nil {
		logger.Error("ExpireOldInputRequests: Failed", zap.Error(err))
		return 0, err
	}

	count, _ := result.RowsAffected()
	if count > 0 {
		logger.Info("ExpireOldInputRequests: Expired requests",
			zap.Int64("count", count))
	}

	return count, nil
}

// Helper functions

func nullableJSON(data []byte) interface{} {
	if len(data) == 0 {
		return nil
	}
	return data
}
