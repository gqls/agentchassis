// platform/orchestration/actions/orchestration_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func StartOrchestrationAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Starting orchestration action",
		zap.Any("collected_data_keys", getMapKeys(params.CollectedData)),
		zap.String("current_step", params.CurrentStep))

	// Try the step name first (correct behavior)
	var spawnResult interface{}
	var found bool

	// The previous step should be "spawn_website_team"
	spawnResult, found = params.CollectedData["spawn_website_team"]
	if !found {
		// Fallback to action name (this shouldn't happen but it is)
		params.Logger.Warn("Data not found under step name, trying action name")
		spawnResult, found = params.CollectedData["spawn_group"]
		if !found {
			params.Logger.Error("Spawn result not found anywhere",
				zap.Any("available_keys", getMapKeys(params.CollectedData)))
			return nil, fmt.Errorf("spawn result not found in collected data (looked for 'spawn_website_team' and 'spawn_group')")
		}
	}

	spawnData, ok := spawnResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("spawn result is not a map, got %T", spawnResult)
	}

	params.Logger.Info("Found spawn data",
		zap.String("group_id", fmt.Sprintf("%v", spawnData["group_id"])),
		zap.String("group_name", fmt.Sprintf("%v", spawnData["group_name"])),
		zap.Any("agents", spawnData["agents"]))

	// Get the workflow
	var workflowJSON json.RawMessage
	if workflow, ok := spawnData["workflow"]; ok {

		params.Logger.Info("Workflow from spawn data",
			zap.Any("workflow_raw", workflow),
			zap.String("workflow_type", fmt.Sprintf("%T", workflow)))

		switch w := workflow.(type) {
		case json.RawMessage:
			workflowJSON = w
		case []byte:
			workflowJSON = json.RawMessage(w)
		case map[string]interface{}:
			bytes, err := json.Marshal(w)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal workflow: %w", err)
			}
			workflowJSON = json.RawMessage(bytes)
		default:
			return nil, fmt.Errorf("workflow has unexpected type: %T", w)
		}

		params.Logger.Info("Workflow JSON to be used",
			zap.String("workflow_json", string(workflowJSON)),
		)

	} else {

		params.Logger.Info("Workflow erroring from spawn data",
			zap.Any("workflow_raw (empty?)", workflow),
			zap.String("workflow_type", fmt.Sprintf("%T", workflow)))

		return nil, fmt.Errorf("workflow not found in spawn result")
	}

	// Create new correlation ID for the new orchestration
	newCorrelationID := uuid.New().String()

	// Prepare headers for the new orchestration
	newHeaders := make(map[string]string)
	for k, v := range params.Headers {
		newHeaders[k] = v
	}
	newHeaders["correlation_id"] = newCorrelationID
	newHeaders["parent_correlation_id"] = params.Headers["correlation_id"]

	// Add spawned agents to headers if available
	if agentsRaw, ok := spawnData["agents"]; ok {
		params.Logger.Info("Found agents in spawn data",
			zap.String("agents_type", fmt.Sprintf("%T", agentsRaw)))

		switch agents := agentsRaw.(type) {
		case map[string]string:
			for role, agentID := range agents {
				newHeaders[fmt.Sprintf("agent_%s", role)] = agentID
			}
			params.Logger.Info("Added agent mappings to headers (string map)",
				zap.Int("agent_count", len(agents)))

		case map[string]interface{}:
			count := 0
			for role, agentIDRaw := range agents {
				if agentID, ok := agentIDRaw.(string); ok {
					newHeaders[fmt.Sprintf("agent_%s", role)] = agentID
					count++
				}
			}
			params.Logger.Info("Added agent mappings to headers (interface map)",
				zap.Int("agent_count", count))

		default:
			params.Logger.Warn("Unexpected type for agents",
				zap.String("type", fmt.Sprintf("%T", agentsRaw)))
		}
	} else {
		params.Logger.Warn("No agents found in spawn data")
	}

	params.Logger.Info("Creating new orchestration",
		zap.String("new_correlation_id", newCorrelationID),
		zap.String("parent_correlation_id", params.Headers["correlation_id"]))

	// Get the SagaCoordinator
	type orchestratorInterface interface {
		CreateNewOrchestration(context.Context, string, map[string]string, json.RawMessage) error
	}

	orchestrator, ok := params.SagaCoordinator.(orchestratorInterface)
	if !ok || orchestrator == nil {
		return nil, fmt.Errorf("SagaCoordinator not available or doesn't implement required interface")
	}

	// Create the new orchestration
	err := orchestrator.CreateNewOrchestration(ctx, newCorrelationID, newHeaders, workflowJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestration: %w", err)
	}

	params.Logger.Info("New orchestration created successfully",
		zap.String("new_correlation_id", newCorrelationID),
		zap.String("group_id", fmt.Sprintf("%v", spawnData["group_id"])))

	return map[string]interface{}{
		"status":             "orchestration_started",
		"new_correlation_id": newCorrelationID,
		"group_id":           spawnData["group_id"],
	}, nil
}

func getMapKeys(m map[string]interface{}) []string {
	if m == nil {
		return []string{}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
