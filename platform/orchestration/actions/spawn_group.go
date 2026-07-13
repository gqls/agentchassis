// FILE: platform/orchestration/actions/spawn_group_from_db.go
// This extends SpawnGroupAction to support:
// - Dynamic group_type resolution from collected_data
// - Group definition lookup from agent_group_definitions table
// - Workflow and agent configs from database
//
// Use this when you want to spawn a group based on a classification result
// or other dynamic input, rather than hardcoded config.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// FILE: platform/orchestration/actions/spawn_group.go

// SpawnGroupAction spawns an orchestrator agent for a group type
// This is now a thin wrapper around SpawnAgentAction - the group_type
// maps directly to an agent_type that contains the orchestration workflow.
func SpawnGroupAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("SpawnGroupAction starting - will delegate to SpawnAgentAction")

	config := params.StepConfig.Config

	// 1. Resolve group_type (static or dynamic from field)
	groupType, err := resolveGroupType(config, params.CollectedData, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve group_type: %w", err)
	}

	params.Logger.Info("Resolved group type, spawning as agent",
		zap.String("group_type", groupType))

	// 2. Build spawn_agent config
	spawnConfig := map[string]interface{}{
		"agent_type": groupType,
		"role":       "orchestrator",
	}

	// Copy any additional config (like send_init_data)
	if sendInit, ok := config["send_init_data"].(bool); ok {
		spawnConfig["send_init_data"] = sendInit
	}

	// 3. Prepare input data to pass to the spawned orchestrator
	inputData := prepareInputDataForSpawn(config, params.CollectedData, params.Logger)

	// Store input_data in collected_data so SpawnAgentAction can access it
	if params.CollectedData == nil {
		params.CollectedData = make(map[string]interface{})
	}
	params.CollectedData["__spawn_input_data__"] = inputData

	// 4. Create spawn params
	spawnParams := params
	spawnParams.StepConfig = models.Step{
		Action:      "spawn_agent",
		Config:      spawnConfig,
		NextStep:    params.StepConfig.NextStep,
		OutputField: params.StepConfig.OutputField,
		Description: fmt.Sprintf("Spawn %s orchestrator", groupType),
	}

	// 5. Delegate to SpawnAgentAction
	result, err := SpawnAgentAction(ctx, spawnParams)
	if err != nil {
		return nil, fmt.Errorf("failed to spawn group orchestrator: %w", err)
	}

	// 6. Add group-specific metadata to result
	if resultMap, ok := result.(map[string]interface{}); ok {
		resultMap["group_type"] = groupType
		resultMap["is_group_orchestrator"] = true
	}

	params.Logger.Info("SpawnGroupAction completed - orchestrator agent spawned",
		zap.String("group_type", groupType))

	return result, nil
}

// resolveGroupType gets the group type from static config or dynamic field
func resolveGroupType(config map[string]interface{}, collectedData map[string]interface{}, logger *zap.Logger) (string, error) {
	// Static group_type
	if groupType, ok := config["group_type"].(string); ok && groupType != "" {
		return groupType, nil
	}

	// Dynamic from field path
	if fieldPath, ok := config["group_type_field"].(string); ok && fieldPath != "" {
		value := datahelpers.ExtractNestedField(collectedData, fieldPath)
		if groupType, ok := value.(string); ok && groupType != "" {
			// Apply suffix if configured
			if suffix, ok := config["group_type_suffix"].(string); ok {
				groupType = groupType + suffix
			}
			return groupType, nil
		}
		return "", fmt.Errorf("field %s did not resolve to a string", fieldPath)
	}

	return "", fmt.Errorf("spawn_group requires 'group_type' or 'group_type_field' in config")
}

// prepareInputDataForSpawn gathers input data to pass to spawned orchestrator
func prepareInputDataForSpawn(config map[string]interface{}, collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	inputData := make(map[string]interface{})

	// Get input_fields from config
	inputFields, _ := config["input_fields"].([]interface{})

	for _, field := range inputFields {
		fieldName, ok := field.(string)
		if !ok {
			continue
		}
		if value, exists := collectedData[fieldName]; exists {
			inputData[fieldName] = value
		}
	}

	// Always include input_data if present
	if existingInput, ok := collectedData["input_data"]; ok {
		inputData["input_data"] = existingInput
	}

	return inputData
}

// SpawnGroupFromDBAction spawns an agent group by looking up its definition
// from the agent_group_definitions table. This is useful when the group type
// is determined dynamically (e.g., from a classifier).
//
// Config:
//   - group_type: string - explicit group type to spawn
//   - group_type_field: string - dot-path to read group_type from collected_data
//   - group_type_suffix: string - appended to resolved type (e.g., "-builder")
//   - input_fields: []string - fields from collected_data to pass to spawned group
//   - version: int - specific group version (optional, defaults to latest)
//
// Example configs:
//
//	Static group type:
//	{"group_type": "landing-page-builder"}
//
//	Dynamic from field:
//	{"group_type_field": "confirmed_type.recommended_group"}
//
//	Dynamic with suffix:
//	{"group_type_field": "classification.site_type", "group_type_suffix": "-builder"}
//	// If classification.site_type = "content", spawns "content-builder"
func SpawnGroupActionNewerOld(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("SpawnGroupFromDBAction starting",
		zap.Any("config", params.StepConfig))

	config := params.StepConfig.Config

	// 1. Resolve group_type (static or dynamic)
	groupType, err := resolveGroupTypeForSpawn(config, params.CollectedData, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve group_type: %w", err)
	}

	params.Logger.Info("Resolved group type for spawn",
		zap.String("group_type", groupType),
	)

	// 2. Get database connection
	db := params.DB
	if db == nil {
		return nil, fmt.Errorf("database connection required for SpawnGroupFromDBAction")
	}

	// 3. Fetch group definition from database
	groupDef, err := fetchGroupDefinitionFromDB(ctx, db, groupType, config, params.Logger)
	params.Logger.Info("Looking for chosen group type in db",
		zap.Any("chosen group definition", groupDef),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group definition from db: %w", err)
	}

	params.Logger.Info("Fetched group definition from DB",
		zap.String("group_type", groupType),
		zap.String("group_name", groupDef.Name),
		zap.Int("version", groupDef.Version),
		zap.Int("agent_count", len(groupDef.AgentConfigs)),
	)

	// 4. Generate group ID and name (same pattern as existing SpawnGroupAction)
	groupID := uuid.New().String()
	groupName := GenerateGroupName(groupType)

	// 5. Handle request ID tracking (same pattern as existing)
	requestID := params.Headers["request_id"]
	if params.Headers["reply_to_request_id"] == "" {
		params.Headers["reply_to_request_id"] = params.Headers["request_id"]
	}
	if params.Headers["parent_responses_topic"] == "" {
		params.Headers["parent_responses_topic"] = params.Headers["responsesTopic"]
	}
	if requestID == "" {
		requestID = uuid.New().String()
		params.Headers["request_id"] = requestID
	}

	params.Logger.Info("Spawning agent group from DB definition",
		zap.String("group_id", groupID),
		zap.String("group_name", groupName),
		zap.String("group_type", groupType),
		zap.Int("agent_count", len(groupDef.AgentConfigs)),
		zap.String("request_id", requestID),
	)

	// 6. Prepare input data to pass to spawned group
	inputData := prepareInputDataForGroup(config, params.CollectedData, params.Logger)

	// 7. Track spawned agents (same pattern as existing)
	spawnedAgents := make(map[string]string) // role -> agentID
	spawnedSubtrees := make(map[string]*types.SubtreeInfo)

	// 8. Spawn each agent defined in the group
	for _, agentConfig := range groupDef.AgentConfigs {
		// Create spawn config for this agent
		spawnStepConfig := models.Step{
			Action: "spawn_agent",
			Config: map[string]interface{}{
				"agent_type": agentConfig.AgentType,
				"role":       agentConfig.Role,
				"group_id":   groupID,
				"group_name": groupName,
			},
		}

		// Add image tag if pinned
		if agentConfig.ImageTag != "" {
			spawnStepConfig.Config["image_tag"] = agentConfig.ImageTag
		}

		// Use SpawnAgentAction to spawn individual agent
		spawnParams := params // Copy params
		spawnParams.StepConfig = spawnStepConfig

		result, err := SpawnAgentAction(ctx, spawnParams)
		if err != nil {
			params.Logger.Error("Failed to spawn agent",
				zap.String("role", agentConfig.Role),
				zap.String("agent_type", agentConfig.AgentType),
				zap.Error(err))
			continue
		}

		agentResult, ok := result.(map[string]interface{})
		if !ok {
			params.Logger.Error("Unexpected result type from SpawnAgentAction",
				zap.String("role", agentConfig.Role))
			continue
		}

		agentID, ok := agentResult["agent_id"].(string)
		if !ok {
			params.Logger.Error("No agent_id in result",
				zap.String("role", agentConfig.Role))
			continue
		}

		spawnedAgents[agentConfig.Role] = agentID

		// Track subtree info
		if subtreeInfo, ok := agentResult["subtree_info"].(*types.SubtreeInfo); ok {
			spawnedSubtrees[agentID] = subtreeInfo
		}

		params.Logger.Info("Successfully spawned agent in group",
			zap.String("role", agentConfig.Role),
			zap.String("agent_id", agentID),
			zap.String("agent_type", agentConfig.AgentType),
			zap.String("group_id", groupID))
	}

	params.Logger.Info("All agents in group spawned",
		zap.String("group_id", groupID),
		zap.Int("total_spawned", len(spawnedAgents)),
		zap.Any("spawned_agents", spawnedAgents))

	// 9. Marshal workflow from group definition
	workflowBytes, err := json.Marshal(groupDef.OrchestrationWorkflow)
	if err != nil {
		params.Logger.Error("Failed to marshal workflow", zap.Error(err))
		workflowBytes = []byte(`{}`)
	}

	// 10. Create group subtree info (same pattern as existing)
	groupSubtree := &types.SubtreeInfo{
		AgentID:       groupID,
		AgentType:     "group",
		AgentName:     groupName,
		ParentAgentID: params.ExecutionContext.FromAgentID,
		Children:      spawnedSubtrees,
		CreatedAt:     time.Now(),
		LastActiveAt:  time.Now(),
		Performance: &types.PerformanceMetrics{
			TasksCompleted: 0,
			TasksFailed:    0,
			LastUpdated:    time.Now(),
		},
	}

	// Check if we should wait for responses (default: false for spawn_group)
	// When wait_for_completion is false, we spawn the group and immediately
	// proceed to the next step without waiting for agent initialization responses.
	waitForCompletion := false
	if wait, ok := config["wait_for_completion"].(bool); ok {
		waitForCompletion = wait
	}

	params.Logger.Info("SpawnGroupFromDBAction completed",
		zap.String("group_id", groupID),
		zap.String("group_type", groupType),
		zap.Int("version", groupDef.Version),
		zap.Int("agents_spawned", len(spawnedAgents)))

	// 11. Return result (same pattern as existing SpawnGroupAction)
	return map[string]interface{}{
		"group_id":      groupID,
		"group_name":    groupName,
		"group_type":    groupType,
		"group_version": groupDef.Version,
		"agents":        spawnedAgents,
		"workflow":      workflowBytes,
		"request_id":    requestID,
		// "await_response":         true, // Wait for group initialization
		"await_response":         waitForCompletion,
		"target_agent_type":      groupType,
		"subtree_info":           groupSubtree,
		"spawn_count":            len(spawnedAgents),
		"input_data":             inputData,
		"briefing_questionnaire": groupDef.BriefingQuestionnaire,
	}, nil
}

// ============================================================================
// Types
// ============================================================================

type DBGroupDefinition struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	GroupType             string                 `json:"group_type"`
	Version               int                    `json:"version"`
	Description           string                 `json:"description"`
	AgentConfigs          []DBAgentConfig        `json:"agent_configs"`
	OrchestrationWorkflow map[string]interface{} `json:"orchestration_workflow"`
	BriefingQuestionnaire map[string]interface{} `json:"briefing_questionnaire"`
}

type DBAgentConfig struct {
	Role      string `json:"role"`
	AgentType string `json:"agent_type"`
	ImageTag  string `json:"image_tag,omitempty"`
}

// ============================================================================
// Helper Functions
// ============================================================================

func resolveGroupTypeForSpawn(config map[string]interface{}, collectedData map[string]interface{}, logger *zap.Logger) (string, error) {
	// Option 1: Explicit group_type in config
	if groupType, ok := config["group_type"].(string); ok && groupType != "" {
		return groupType, nil
	}

	// Option 2: Dynamic from field path
	fieldPath, ok := config["group_type_field"].(string)
	if !ok || fieldPath == "" {
		return "", fmt.Errorf("either 'group_type' or 'group_type_field' must be specified")
	}

	// Navigate to the field
	value, err := getNestedValueForSpawn(collectedData, fieldPath)
	if err != nil {
		return "", fmt.Errorf("failed to get value at path '%s': %w", fieldPath, err)
	}

	groupType, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value at path '%s' is not a string: %T", fieldPath, value)
	}

	// Optional suffix
	if suffix, ok := config["group_type_suffix"].(string); ok {
		groupType = groupType + suffix
	}

	return groupType, nil
}

func getNestedValueForSpawn(data map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, exists := v[part]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found at position %d", part, i)
			}
			current = val
		case string:
			// Try to parse as JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(v), &parsed); err != nil {
				return nil, fmt.Errorf("cannot navigate into string at '%s'", part)
			}
			val, exists := parsed[part]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found in parsed JSON at position %d", part, i)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot navigate into type %T at '%s'", current, part)
		}
	}

	return current, nil
}

func fetchGroupDefinitionFromDB(ctx context.Context, db *sql.DB, groupType string, config map[string]interface{}, logger *zap.Logger) (*DBGroupDefinition, error) {
	var query string
	var args []interface{}

	// Check if specific version requested
	if version, ok := config["version"].(float64); ok {
		query = `
			SELECT id, name, group_type, version, COALESCE(description, ''),
			       agent_configs, orchestration_workflow, 
			       COALESCE(briefing_questionnaire, '{}'::jsonb)
			FROM agent_group_definitions 
			WHERE group_type = $1 AND version = $2`
		args = []interface{}{groupType, int(version)}
	} else {
		// Get latest version
		query = `
			SELECT id, name, group_type, version, COALESCE(description, ''),
			       agent_configs, orchestration_workflow,
			       COALESCE(briefing_questionnaire, '{}'::jsonb)
			FROM agent_group_definitions 
			WHERE group_type = $1 
			ORDER BY version DESC 
			LIMIT 1`
		args = []interface{}{groupType}
	}

	var def DBGroupDefinition
	var agentConfigsJSON, workflowJSON, questionnaireJSON []byte

	err := db.QueryRowContext(ctx, query, args...).Scan(
		&def.ID,
		&def.Name,
		&def.GroupType,
		&def.Version,
		&def.Description,
		&agentConfigsJSON,
		&workflowJSON,
		&questionnaireJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("group type '%s' not found in database", groupType)
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal(agentConfigsJSON, &def.AgentConfigs); err != nil {
		return nil, fmt.Errorf("failed to parse agent_configs: %w", err)
	}
	if err := json.Unmarshal(workflowJSON, &def.OrchestrationWorkflow); err != nil {
		return nil, fmt.Errorf("failed to parse orchestration_workflow: %w", err)
	}
	if err := json.Unmarshal(questionnaireJSON, &def.BriefingQuestionnaire); err != nil {
		return nil, fmt.Errorf("failed to parse briefing_questionnaire: %w", err)
	}

	logger.Info("Loaded group definition from database",
		zap.String("group_type", def.GroupType),
		zap.Int("version", def.Version),
		zap.Int("agent_count", len(def.AgentConfigs)),
	)

	return &def, nil
}

func prepareInputDataForGroup(config map[string]interface{}, collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	inputData := make(map[string]interface{})

	// Get fields to pass
	inputFields, ok := config["input_fields"].([]interface{})
	if !ok {
		// Default: pass all collected data
		for k, v := range collectedData {
			inputData[k] = v
		}
		return inputData
	}

	// Extract specified fields
	for _, field := range inputFields {
		fieldName, ok := field.(string)
		if !ok {
			continue
		}

		// Support dot notation
		if strings.Contains(fieldName, ".") {
			value, err := getNestedValueForSpawn(collectedData, fieldName)
			if err != nil {
				logger.Warn("Could not extract field for group input",
					zap.String("field", fieldName),
					zap.Error(err),
				)
				continue
			}
			// Store at leaf key name
			parts := strings.Split(fieldName, ".")
			inputData[parts[len(parts)-1]] = value
		} else {
			if value, exists := collectedData[fieldName]; exists {
				inputData[fieldName] = value
			}
		}
	}

	return inputData
}

// ============================================================================
// REGISTRY UPDATE
// ============================================================================
// Add to GlobalActionRegistry in registry.go:
//
// "spawn_group_from_db": SpawnGroupFromDBAction,
//
// Note: This complements the existing "spawn_group" action.
// Use spawn_group when agent configs are in the workflow config.
// Use spawn_group_from_db when agent configs should come from database.
//
