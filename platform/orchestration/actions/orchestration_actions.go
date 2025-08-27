// platform/orchestration/actions/orchestration_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

func StartOrchestrationAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Starting orchestration action",
		zap.Any("DEBUGOACTIONS1: collected_data_keys", getMapKeys(params.CollectedData)),
		zap.String("DEBUGOACTIONS1 :current_step", params.CurrentStep))

	// CHECK: Has this step already created a child orchestration?
	stepKey := fmt.Sprintf("%s_started", params.CurrentStep)
	if _, alreadyStarted := params.CollectedData[stepKey]; alreadyStarted {
		params.Logger.Warn("Child orchestration already started for this step",
			zap.String("DEBUGOACTIONS1 :step", params.CurrentStep))

		// Try to find the existing child ID
		if existingChild, ok := params.CollectedData[params.CurrentStep]; ok {
			if childMap, ok := existingChild.(map[string]interface{}); ok {
				if childID, ok := childMap["new_correlation_id"].(string); ok && childID != "" {
					// Return the existing result
					return existingChild, nil
				}
			}
		}

		// If we can't find the child ID, we have a problem
		return nil, fmt.Errorf("orchestration already started but cannot find child ID")
	}

	// Mark that we're starting this orchestration
	params.CollectedData[stepKey] = true

	// The previous step should have the spawn result
	// In your workflow: spawn_website_team -> start_website_workflow
	// So we need to look at the previous step's result

	// Try the step name first (correct behavior)
	var spawnResult interface{}
	var found bool

	// Try to find the most recent spawn result
	// Look through execution path to find the last spawn action
	for stepName, data := range params.CollectedData {
		if dataMap, ok := data.(map[string]interface{}); ok {
			// Check if this is a spawn result (has workflow and agents)
			if _, hasWorkflow := dataMap["workflow"]; hasWorkflow {
				if _, hasAgents := dataMap["agents"]; hasAgents {
					spawnResult = dataMap
					found = true
					params.Logger.Info("Found spawn result from step",
						zap.String("step_name", stepName))
					// Don't break - keep looking for the most recent one
				}
			}
			// Also check if it has group_id (another indicator of spawn result)
			if _, hasGroupID := dataMap["group_id"]; hasGroupID {
				if _, hasAgents := dataMap["agents"]; hasAgents {
					spawnResult = dataMap
					found = true
					params.Logger.Info("Found spawn result from step",
						zap.String("step_name", stepName))
				}
			}
		}
	}

	if !found {
		params.Logger.Error("No spawn result found in collected data",
			zap.Any("available_keys", getMapKeys(params.CollectedData)))
		return nil, fmt.Errorf("spawn result not found")
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
	// Store parent correlation ID and agent type in collected data for child to access later
	params.CollectedData["parent_correlation_id"] = params.Headers["correlation_id"]
	params.CollectedData["parent_agent_type"] = params.Headers["agent_type"]
	newHeaders["parent_agent_type"] = params.AgentType

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
		zap.String("DEBUGOACTIONS2: new_correlation_id", newCorrelationID),
		zap.String("DEBUGOACTIONS2: parent_correlation_id", params.Headers["correlation_id"]))

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
		zap.String("DEBUGOACTIONS3: new_correlation_id", newCorrelationID),
		zap.String("DEBUGOACTIONS3: group_id", fmt.Sprintf("%v", spawnData["group_id"])))

	// Start timeout monitor for child orchestration
	go func() {
		// Configure timeout based on workflow complexity
		timeout := 5 * time.Minute // Default timeout

		// You could make this configurable based on the workflow type
		if workflowTimeout, ok := params.StepConfig.Config["child_timeout_minutes"].(float64); ok {
			timeout = time.Duration(workflowTimeout) * time.Minute
		}

		params.Logger.Info("Starting timeout monitor for child orchestration",
			zap.String("DEBUGOACTIONS4: child_correlation_id", newCorrelationID),
			zap.String("DEBUGOACTIONS4: parent_correlation_id", params.Headers["correlation_id"]),
			zap.Duration("timeout", timeout))

		// Wait for the timeout period
		time.Sleep(timeout)

		// After timeout, check if parent is still waiting for this child
		// We need to check the parent's state, not the child's
		parentCorrelationID := params.Headers["correlation_id"]

		// Since we're in a goroutine, we need to be careful about DB access
		// Create a new context for this check
		checkCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Check parent state to see if it's still waiting
		if params.DB != nil {
			var status string
			var awaitedSteps []byte

			query := `
                SELECT status, awaited_steps 
                FROM orchestrator_state 
                WHERE correlation_id = $1
            `

			err := params.DB.QueryRowContext(checkCtx, query, parentCorrelationID).Scan(&status, &awaitedSteps)
			if err != nil {
				params.Logger.Error("Failed to check parent state for timeout",
					zap.String("DEBUGOACTIONS5: parent_correlation_id", parentCorrelationID),
					zap.Error(err))
				return
			}

			// Parse awaited steps
			var awaited []string
			if err := json.Unmarshal(awaitedSteps, &awaited); err != nil {
				params.Logger.Error("Failed to parse awaited steps",
					zap.Error(err))
				return
			}

			// Check if parent is still waiting for this child
			stillWaiting := false
			for _, step := range awaited {
				if step == newCorrelationID {
					stillWaiting = true
					break
				}
			}

			if stillWaiting && status == "AWAITING_RESPONSES" {
				params.Logger.Error("Child orchestration timeout - parent still waiting",
					zap.String("DEBUGOACTIONS6: child_correlation_id", newCorrelationID),
					zap.String("DEBUGOACTIONS6: parent_correlation_id", parentCorrelationID),
					zap.Duration("DEBUGOACTIONS6: timeout_after", timeout))

				// Send timeout notification to parent
				timeoutResponse := models.TaskResponse{
					Success: false,
					Error:   fmt.Sprintf("Child orchestration timeout after %v", timeout),
					Data: map[string]interface{}{
						"status":         "timeout",
						"correlation_id": newCorrelationID,
						"error":          fmt.Sprintf("Child orchestration %s timed out after %v", newCorrelationID, timeout),
						"timeout_at":     time.Now().UTC(),
					},
				}

				responseBytes, _ := json.Marshal(timeoutResponse)

				// Send timeout notification to parent
				timeoutHeaders := map[string]string{
					"correlation_id":        parentCorrelationID,
					"causation_id":          newCorrelationID,
					"parent_correlation_id": parentCorrelationID,
					"message_type":          "orchestration_timeout",
				}

				// Use the producer to send timeout notification
				if params.Producer != nil {
					err := params.Producer.Produce(checkCtx,
						"system.agent.generic.responses",
						timeoutHeaders,
						[]byte(parentCorrelationID),
						responseBytes)

					if err != nil {
						params.Logger.Error("Failed to send timeout notification to parent",
							zap.Error(err))
					} else {
						params.Logger.Info("Timeout notification sent to parent",
							zap.String("parent_correlation_id", parentCorrelationID))
					}
				}

				// Optional: Also check and update child orchestration status
				var childStatus string
				err = params.DB.QueryRowContext(checkCtx,
					"SELECT status FROM orchestrator_state WHERE correlation_id = $1",
					newCorrelationID).Scan(&childStatus)

				if err == nil && childStatus == "RUNNING" {
					// Child is still running after timeout - mark it as failed
					_, err = params.DB.ExecContext(checkCtx, `
                        UPDATE orchestrator_state 
                        SET status = 'FAILED', 
                            error = $2,
                            updated_at = NOW()
                        WHERE correlation_id = $1 AND status = 'RUNNING'
                    `, newCorrelationID, fmt.Sprintf("Timeout after %v", timeout))

					if err == nil {
						params.Logger.Info("Marked child orchestration as failed due to timeout",
							zap.String("DEBUGOACTIONS7: child_correlation_id", newCorrelationID))
					}
				}
			} else {
				params.Logger.Info("Timeout check passed - parent no longer waiting or child completed",
					zap.String("DEBUGOACTIONS8: child_correlation_id", newCorrelationID),
					zap.String("DEBUGOACTIONS8: parent_status", status),
					zap.Bool("DEBUGOACTIONS8: still_waiting", stillWaiting))
			}
		}
	}()

	return map[string]interface{}{
		"status":             "orchestration_started",
		"new_correlation_id": newCorrelationID,
		"group_id":           spawnData["group_id"],
		"await_response":     true,             // THIS IS KEY!
		"request_id":         newCorrelationID, // Parent will wait for this
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
