// FILE: internal/core-manager/admin/client_handlers.go
package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ClientHandlers handles admin operations for client management
type ClientHandlers struct {
	clientsDB *pgxpool.Pool
	logger    *zap.Logger
}

// NewClientHandlers creates new client admin handlers
func NewClientHandlers(clientsDB *pgxpool.Pool, logger *zap.Logger) *ClientHandlers {
	return &ClientHandlers{
		clientsDB: clientsDB,
		logger:    logger,
	}
}

// CreateClientRequest represents a request to create a new client
type CreateClientRequest struct {
	ClientID    string                 `json:"client_id" binding:"required,alphanum,min=3,max=50"`
	DisplayName string                 `json:"display_name" binding:"required"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// ClientInfo represents client information
type ClientInfo struct {
	ClientID    string                 `json:"client_id"`
	DisplayName string                 `json:"display_name"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	IsActive    bool                   `json:"is_active"`
}

// HandleCreateClient creates a new client with schema
func (h *ClientHandlers) HandleCreateClient(c *gin.Context) {
	var req CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate client_id format (alphanumeric, underscores allowed)
	if !isValidClientID(req.ClientID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client_id format. Use only alphanumeric characters and underscores"})
		return
	}

	// Check if client already exists
	exists, err := h.clientExists(c.Request.Context(), req.ClientID)
	if err != nil {
		h.logger.Error("Failed to check client existence", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check client existence"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Client already exists"})
		return
	}

	// Create client schema
	if err := h.createClientSchema(c.Request.Context(), req.ClientID); err != nil {
		h.logger.Error("Failed to create client schema", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client schema"})
		return
	}

	// Store client info in a clients table (we'll need to create this)
	if err := h.storeClientInfo(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to store client info", zap.Error(err))
		// Note: Schema is already created, this is a partial failure
	}

	h.logger.Info("Client created successfully", zap.String("client_id", req.ClientID))
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Client created successfully",
		"client_id": req.ClientID,
	})
}

// HandleListClients lists all clients
func (h *ClientHandlers) HandleListClients(c *gin.Context) {
	clients, err := h.listClients(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list clients", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list clients"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": clients,
		"count":   len(clients),
	})
}

// HandleGetClientUsage gets usage statistics for a client
func (h *ClientHandlers) HandleGetClientUsage(c *gin.Context) {
	clientID := c.Param("client_id")

	if !isValidClientID(clientID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client_id format"})
		return
	}

	usage, err := h.getClientUsage(c.Request.Context(), clientID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
			return
		}
		h.logger.Error("Failed to get client usage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client usage"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// Helper functions

func isValidClientID(clientID string) bool {
	// Only allow alphanumeric and underscores
	for _, char := range clientID {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	return len(clientID) >= 3 && len(clientID) <= 50
}

func (h *ClientHandlers) clientExists(ctx context.Context, clientID string) (bool, error) {
	// Check if schema exists
	query := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.schemata 
			WHERE schema_name = $1
		)
	`
	schemaName := fmt.Sprintf("client_%s", clientID)

	var exists bool
	err := h.clientsDB.QueryRow(ctx, query, schemaName).Scan(&exists)
	return exists, err
}

func (h *ClientHandlers) createClientSchema(ctx context.Context, clientID string) error {
	// Call the stored procedure to create client schema
	query := `SELECT create_client_schema($1)`
	_, err := h.clientsDB.Exec(ctx, query, clientID)
	return err
}

func (h *ClientHandlers) storeClientInfo(ctx context.Context, req *CreateClientRequest) error {
	// First, ensure we have a clients info table
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS clients_info (
			client_id VARCHAR(50) PRIMARY KEY,
			display_name VARCHAR(255) NOT NULL,
			settings JSONB DEFAULT '{}',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`

	if _, err := h.clientsDB.Exec(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to create clients_info table: %w", err)
	}

	// Insert client info
	insertQuery := `
		INSERT INTO clients_info (client_id, display_name, settings)
		VALUES ($1, $2, $3)
		ON CONFLICT (client_id) DO NOTHING
	`

	_, err := h.clientsDB.Exec(ctx, insertQuery, req.ClientID, req.DisplayName, req.Settings)
	return err
}

func (h *ClientHandlers) listClients(ctx context.Context) ([]ClientInfo, error) {
	// First ensure the table exists
	h.storeClientInfo(ctx, &CreateClientRequest{}) // This will create table if needed

	query := `
		SELECT client_id, display_name, settings, is_active, created_at
		FROM clients_info
		WHERE is_active = true
		ORDER BY created_at DESC
	`

	rows, err := h.clientsDB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []ClientInfo
	for rows.Next() {
		var client ClientInfo
		var createdAt sql.NullTime
		var settings sql.NullString

		err := rows.Scan(&client.ClientID, &client.DisplayName, &settings, &client.IsActive, &createdAt)
		if err != nil {
			h.logger.Error("Failed to scan client row", zap.Error(err))
			continue
		}

		if createdAt.Valid {
			client.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z")
		}

		clients = append(clients, client)
	}

	// Also check for schemas without entries in clients_info
	schemaQuery := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name LIKE 'client_%'
		AND schema_name NOT IN (
			SELECT 'client_' || client_id FROM clients_info
		)
	`

	schemaRows, err := h.clientsDB.Query(ctx, schemaQuery)
	if err == nil {
		defer schemaRows.Close()
		for schemaRows.Next() {
			var schemaName string
			if err := schemaRows.Scan(&schemaName); err == nil {
				clientID := strings.TrimPrefix(schemaName, "client_")
				clients = append(clients, ClientInfo{
					ClientID:    clientID,
					DisplayName: clientID + " (Legacy)",
					IsActive:    true,
					CreatedAt:   "Unknown",
				})
			}
		}
	}

	return clients, nil
}

// ClientUsageStats represents usage statistics for a client
type ClientUsageStats struct {
	ClientID            string `json:"client_id"`
	TotalUsers          int    `json:"total_users"`
	ActiveUsers         int    `json:"active_users"`
	TotalInstances      int    `json:"total_instances"`
	ActiveInstances     int    `json:"active_instances"`
	TotalWorkflows      int    `json:"total_workflows"`
	WorkflowsLast30Days int    `json:"workflows_last_30_days"`
	TotalMemoryEntries  int    `json:"total_memory_entries"`
	TotalFuelConsumed   int64  `json:"total_fuel_consumed"`
}

func (h *ClientHandlers) getClientUsage(ctx context.Context, clientID string) (*ClientUsageStats, error) {
	stats := &ClientUsageStats{ClientID: clientID}

	// Get user counts from auth database (would need access to auth DB)
	// For now, we'll focus on what we can get from clients DB

	// Count agent instances
	instanceQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true) as active
		FROM client_%s.agent_instances
	`, clientID)

	err := h.clientsDB.QueryRow(ctx, instanceQuery).Scan(&stats.TotalInstances, &stats.ActiveInstances)
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		return nil, err
	}

	// Count workflows
	workflowQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE started_at > NOW() - INTERVAL '30 days') as recent
		FROM client_%s.workflow_executions
	`, clientID)

	err = h.clientsDB.QueryRow(ctx, workflowQuery).Scan(&stats.TotalWorkflows, &stats.WorkflowsLast30Days)
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		h.logger.Warn("Failed to get workflow stats", zap.Error(err))
	}

	// Count memory entries
	memoryQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM client_%s.agent_memory
	`, clientID)

	err = h.clientsDB.QueryRow(ctx, memoryQuery).Scan(&stats.TotalMemoryEntries)
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		h.logger.Warn("Failed to get memory stats", zap.Error(err))
	}

	// Get fuel consumption
	fuelQuery := fmt.Sprintf(`
		SELECT COALESCE(SUM(fuel_consumed), 0)
		FROM client_%s.usage_analytics
	`, clientID)

	err = h.clientsDB.QueryRow(ctx, fuelQuery).Scan(&stats.TotalFuelConsumed)
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		h.logger.Warn("Failed to get fuel stats", zap.Error(err))
	}

	return stats, nil
}
