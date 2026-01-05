// FILE: platform/orchestration/actions/entity_state_actions.go

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// APPEND ENTITY STATE
// ============================================================================

// AppendEntityStateAction appends data to the entity state log
// Config:
//
//	entity_id_field: path to entity ID in collected_data (e.g., "input_data.domain")
//	entity_type: type of entity (e.g., "domain", "project")
//	namespace: NULL for shared, or specific namespace (defaults to agent_type if "auto")
//	path: dot-notation path (e.g., "brand.tone", "research.products")
//	data_field: path to data in collected_data to store
func AppendEntityStateAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	// 1. Get entity ID
	entityIDField, _ := config["entity_id_field"].(string)
	entityID, err := resolveEntityID(entityIDField, params.CollectedData)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve entity_id: %w", err)
	}

	// 2. Get entity type
	entityType, _ := config["entity_type"].(string)
	if entityType == "" {
		entityType = "default"
	}

	// 3. Get namespace
	namespace := resolveNamespace(config, params)

	// 4. Get path
	path, _ := config["path"].(string)
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	// 5. Get data to store
	dataField, _ := config["data_field"].(string)
	data, err := resolveDataField(dataField, params.CollectedData)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve data_field: %w", err)
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// 6. Insert into log
	var insertedID int64
	err = params.DB.QueryRowContext(ctx, `
        INSERT INTO entity_state_log 
            (entity_id, entity_type, namespace, path, data, created_by_agent_type, orchestration_id, correlation_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `,
		entityID,
		entityType,
		nullableString(namespace),
		path,
		dataJSON,
		os.Getenv("AGENT_TYPE"),
		params.ExecutionContext.OrchestrationID,
		params.ExecutionContext.CorrelationID,
	).Scan(&insertedID)

	if err != nil {
		return nil, fmt.Errorf("failed to insert entity state: %w", err)
	}

	logger.Info("Appended entity state",
		zap.String("entity_id", entityID),
		zap.String("namespace", namespace),
		zap.String("path", path),
		zap.Int64("log_id", insertedID))

	return map[string]interface{}{
		"success":   true,
		"entity_id": entityID,
		"path":      path,
		"log_id":    insertedID,
	}, nil
}

// ============================================================================
// READ LATEST ENTITY STATE
// ============================================================================

// ReadLatestEntityStateAction reads the most recent entry for each path
// Config:
//
//	entity_id_field: path to entity ID
//	entity_type: optional filter
//	namespace: NULL for shared, "auto" for agent_type, or specific
//	paths: array of path patterns to read (supports wildcards with %)
func ReadLatestEntityStateAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	// 1. Get entity ID
	entityIDField, _ := config["entity_id_field"].(string)
	entityID, err := resolveEntityID(entityIDField, params.CollectedData)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve entity_id: %w", err)
	}

	// 2. Get namespace
	namespace := resolveNamespace(config, params)

	// 3. Get paths to read
	pathPatterns := getPathPatterns(config)

	// 4. Query for latest entries
	result := make(map[string]interface{})

	for _, pattern := range pathPatterns {
		rows, err := params.DB.QueryContext(ctx, `
            SELECT DISTINCT ON (path) path, data, created_at, created_by_agent_type
            FROM entity_state_log
            WHERE entity_id = $1 
              AND (namespace = $2 OR ($2 IS NULL AND namespace IS NULL))
              AND path LIKE $3
              AND superseded_by IS NULL
            ORDER BY path, created_at DESC
        `, entityID, nullableString(namespace), pattern)

		if err != nil {
			logger.Error("Failed to query entity state", zap.Error(err))
			continue
		}

		for rows.Next() {
			var path string
			var data json.RawMessage
			var createdAt time.Time
			var createdBy sql.NullString

			if err := rows.Scan(&path, &data, &createdAt, &createdBy); err != nil {
				logger.Error("Failed to scan row", zap.Error(err))
				continue
			}

			var parsedData interface{}
			json.Unmarshal(data, &parsedData)

			result[path] = map[string]interface{}{
				"data":       parsedData,
				"created_at": createdAt,
				"created_by": createdBy.String,
			}
		}
		rows.Close()
	}

	logger.Info("Read latest entity state",
		zap.String("entity_id", entityID),
		zap.String("namespace", namespace),
		zap.Int("entries_found", len(result)))

	return map[string]interface{}{
		"entity_id": entityID,
		"namespace": namespace,
		"state":     result,
	}, nil
}

// ============================================================================
// READ ENTITY HISTORY
// ============================================================================

// ReadEntityHistoryAction reads the full history for a path
// Config:
//
//	entity_id_field: path to entity ID
//	namespace: NULL for shared, "auto" for agent_type, or specific
//	path: specific path to get history for
//	limit: max entries to return (default 50)
func ReadEntityHistoryAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	// 1. Get entity ID
	entityIDField, _ := config["entity_id_field"].(string)
	entityID, err := resolveEntityID(entityIDField, params.CollectedData)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve entity_id: %w", err)
	}

	// 2. Get namespace
	namespace := resolveNamespace(config, params)

	// 3. Get path
	path, _ := config["path"].(string)
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	// 4. Get limit
	limit := 50
	if l, ok := config["limit"].(float64); ok {
		limit = int(l)
	}

	// 5. Query history
	rows, err := params.DB.QueryContext(ctx, `
        SELECT id, data, created_at, created_by_agent_type, orchestration_id, superseded_by
        FROM entity_state_log
        WHERE entity_id = $1 
          AND (namespace = $2 OR ($2 IS NULL AND namespace IS NULL))
          AND path = $3
        ORDER BY created_at DESC
        LIMIT $4
    `, entityID, nullableString(namespace), path, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query entity history: %w", err)
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var id int64
		var data json.RawMessage
		var createdAt time.Time
		var createdBy sql.NullString
		var orchestrationID sql.NullString
		var supersededBy sql.NullInt64

		if err := rows.Scan(&id, &data, &createdAt, &createdBy, &orchestrationID, &supersededBy); err != nil {
			logger.Error("Failed to scan row", zap.Error(err))
			continue
		}

		var parsedData interface{}
		json.Unmarshal(data, &parsedData)

		entry := map[string]interface{}{
			"id":         id,
			"data":       parsedData,
			"created_at": createdAt,
		}
		if createdBy.Valid {
			entry["created_by"] = createdBy.String
		}
		if orchestrationID.Valid {
			entry["orchestration_id"] = orchestrationID.String
		}
		if supersededBy.Valid {
			entry["superseded_by"] = supersededBy.Int64
		}

		history = append(history, entry)
	}

	logger.Info("Read entity history",
		zap.String("entity_id", entityID),
		zap.String("path", path),
		zap.Int("entries_found", len(history)))

	return map[string]interface{}{
		"entity_id": entityID,
		"namespace": namespace,
		"path":      path,
		"history":   history,
		"count":     len(history),
	}, nil
}

// ============================================================================
// READ MY STATE (Agent-namespaced convenience action)
// ============================================================================

// ReadMyStateAction reads the calling agent's namespaced state
// Automatically uses the agent's type as the namespace
// Config:
//
//	entity_id_field: path to entity ID
//	paths: array of path patterns (default: ["%"] for all)
func ReadMyStateAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Force namespace to this agent's type
	config := params.StepConfig.Config
	config["namespace"] = "auto"
	params.StepConfig.Config = config

	return ReadLatestEntityStateAction(ctx, params)
}

// ============================================================================
// WRITE MY STATE (Agent-namespaced convenience action)
// ============================================================================

// WriteMyStateAction writes to the calling agent's namespaced state
// Automatically uses the agent's type as the namespace
// Config:
//
//	entity_id_field: path to entity ID
//	path: path within agent's namespace
//	data_field: data to store
func WriteMyStateAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Force namespace to this agent's type
	config := params.StepConfig.Config
	config["namespace"] = "auto"
	params.StepConfig.Config = config

	return AppendEntityStateAction(ctx, params)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func resolveEntityID(fieldPath string, collectedData map[string]interface{}) (string, error) {
	if fieldPath == "" {
		return "", fmt.Errorf("entity_id_field is required")
	}

	value := datahelpers.ExtractNestedField(collectedData, fieldPath)
	if value == nil {
		return "", fmt.Errorf("entity_id not found at path: %s", fieldPath)
	}

	entityID, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("entity_id at %s is not a string", fieldPath)
	}

	return entityID, nil
}

func resolveNamespace(config map[string]interface{}, params ActionParams) string {
	namespace, _ := config["namespace"].(string)

	if namespace == "auto" || namespace == "" {
		// Use agent type as namespace
		return os.Getenv("AGENT_TYPE")
	}
	if namespace == "shared" {
		return ""
	}
	return namespace
}

func resolveDataField(fieldPath string, collectedData map[string]interface{}) (interface{}, error) {
	if fieldPath == "" {
		return nil, fmt.Errorf("data_field is required")
	}

	value := datahelpers.ExtractNestedField(collectedData, fieldPath)
	if value == nil {
		return nil, fmt.Errorf("data not found at path: %s", fieldPath)
	}

	return value, nil
}

func getPathPatterns(config map[string]interface{}) []string {
	// Check for paths array
	if paths, ok := config["paths"].([]interface{}); ok {
		result := make([]string, 0, len(paths))
		for _, p := range paths {
			if ps, ok := p.(string); ok {
				result = append(result, ps)
			}
		}
		if len(result) > 0 {
			return result
		}
	}

	// Check for single path
	if path, ok := config["path"].(string); ok && path != "" {
		return []string{path}
	}

	// Default to all
	return []string{"%"}
}

/*// resolveFieldPath navigates a dot-notation path through nested maps
func resolveFieldPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		if current == nil {
			return nil
		}

		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case map[interface{}]interface{}:
			current = v[part]
		default:
			return nil
		}
	}

	return current
}*/

/*func init() {
	registry.Register("append_entity_state", registry.ActionDefinition{
		Func:        AppendEntityStateAction,
		Category:    registry.CategoryData,
		Description: "Appends data to entity state log",
		Status:      registry.StatusActive,
	})

	registry.Register("read_latest_entity_state", registry.ActionDefinition{
		Func:        ReadLatestEntityStateAction,
		Category:    registry.CategoryData,
		Description: "Reads most recent entity state entries",
		Status:      registry.StatusActive,
	})

	registry.Register("read_entity_history", registry.ActionDefinition{
		Func:        ReadEntityHistoryAction,
		Category:    registry.CategoryData,
		Description: "Reads full history for an entity state path",
		Status:      registry.StatusActive,
	})

	registry.Register("read_my_state", registry.ActionDefinition{
		Func:        ReadMyStateAction,
		Category:    registry.CategoryData,
		Description: "Reads this agent type's namespaced entity state",
		Status:      registry.StatusActive,
	})

	registry.Register("write_my_state", registry.ActionDefinition{
		Func:        WriteMyStateAction,
		Category:    registry.CategoryData,
		Description: "Writes to this agent type's namespaced entity state",
		Status:      registry.StatusActive,
	})
}*/
