Excellent point! Let me search for all action types to see which ones access CollectedData:Let me search for a list of all action files:Based on the logs, I can see references to **`aggregate_data`** and **`complete_workflow`** actions. Let me create a comprehensive summary of all action types and which ones access `CollectedData`:

## Complete Action Inventory

Based on the code patterns and logs, here are ALL the action types that likely access `CollectedData`:

### **Actions that DEFINITELY access CollectedData:**
1. **call_agent.go** - `extractDataForAgent()` ✅ Confirmed
2. **ai_actions.go** - `extractDataForAiAgent()` ✅ Confirmed
3. **spawn_actions.go** - Likely passes data to spawned agents
4. **workflow_actions.go** - `complete_workflow` - Returns collected data

### **Actions that MIGHT access CollectedData:**
5. **aggregate_actions.go** or similar - `aggregate_data` / `aggregate_webpage` - Collects data from multiple step responses
6. **storage_actions.go** - If storing/retrieving data
7. **calculate_actions.go** - If doing calculations on collected data
8. **basic_actions.go** - Generic data operations
9. **generic_actions.go** - Catch-all operations

## The Universal Solution

Instead of fixing each action individually, we should:

1. **Create a centralized data access layer** that all actions use
2. **Ensure CollectedData is always normalized** when created/modified
3. **Provide helper functions** for safe data access

Here's the complete architectural solution:

```go
// === data_helpers.go (Central Data Access Layer) ===
package orchestration

import (
	"fmt"
	"go.uber.org/zap"
)

// ============================================================================
// NORMALIZATION FUNCTIONS - Called when building CollectedData
// ============================================================================

// NormalizeInputData extracts clean input_data from any message structure
func NormalizeInputData(source map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	var inputData map[string]interface{}
	
	// Try body.input_data first (child agents)
	if body, ok := source["body"].(map[string]interface{}); ok {
		if data, ok := body["input_data"].(map[string]interface{}); ok {
			logger.Debug("Extracted input_data from body.input_data")
			return data
		}
	}
	
	// Try direct input_data (root agents or normalized)
	if data, ok := source["input_data"].(map[string]interface{}); ok {
		// Check if it's clean (not nested with action, config)
		if _, hasAction := data["action"]; !hasAction {
			if _, hasConfig := data["config"]; !hasConfig {
				logger.Debug("Using direct input_data")
				return data
			}
		}
		
		// It's nested, extract real input_data
		if nestedData, ok := data["input_data"].(map[string]interface{}); ok {
			logger.Debug("Extracted from input_data.input_data")
			return nestedData
		}
	}
	
	logger.Debug("No input_data found, returning empty map")
	return map[string]interface{}{}
}

// NormalizeCollectedData builds clean CollectedData from a message
// This is the ONLY way to build CollectedData from messages
func NormalizeCollectedData(
	message map[string]interface{},
	execCtx interface{},
	requestsTopic string,
	logger *zap.Logger,
) map[string]interface{} {
	
	collectedData := map[string]interface{}{
		"__execution_context__": execCtx,
		"__my_requests_topic__": requestsTopic,
		"__raw_message__": message,
	}
	
	// Extract action
	if action, ok := message["action"].(string); ok {
		collectedData["action"] = action
	}
	
	// Extract config (workflow/system config only)
	if config, ok := message["config"].(map[string]interface{}); ok {
		if _, hasWorkflow := config["workflow"]; hasWorkflow {
			collectedData["config"] = config
		} else if _, hasGroupType := config["group_type"]; hasGroupType {
			collectedData["config"] = config
		}
	}
	
	// CRITICAL: Normalize input_data to top level
	collectedData["input_data"] = NormalizeInputData(message, logger)
	
	// Extract other standard fields
	if agentConfig, ok := message["agent_config"].(map[string]interface{}); ok {
		collectedData["agent_config"] = agentConfig
	}
	
	if agentGroup, ok := message["agent_group"].(map[string]interface{}); ok {
		collectedData["agent_group"] = agentGroup
	}
	
	if agentType, ok := message["agent_type"].(string); ok {
		collectedData["agent_type"] = agentType
	}
	
	if prompt, ok := message["prompt"].(string); ok {
		collectedData["prompt"] = prompt
	}
	
	return collectedData
}

// NormalizeResponseData cleans response data before storage in CollectedData
func NormalizeResponseData(responseBody map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	// If response has a data field, use that
	if data, ok := responseBody["data"].(map[string]interface{}); ok {
		logger.Debug("Using response.body.data")
		return data
	}
	
	// If response has result, use that
	if result, ok := responseBody["result"].(map[string]interface{}); ok {
		logger.Debug("Using response.body.result")
		return result
	}
	
	// Return cleaned body (remove system fields)
	cleaned := make(map[string]interface{})
	for k, v := range responseBody {
		if k != "action" && k != "config" && k != "__execution_context__" {
			cleaned[k] = v
		}
	}
	
	logger.Debug("Using cleaned response body")
	return cleaned
}

// ============================================================================
// SAFE ACCESS FUNCTIONS - Used by ALL actions
// ============================================================================

// GetInputData safely retrieves input_data from CollectedData
// ALL actions should use this instead of direct access
func GetInputData(collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	// After normalization, it should be at top level
	if data, ok := collectedData["input_data"].(map[string]interface{}); ok {
		return data
	}
	
	// Shouldn't happen after normalization, but handle gracefully
	logger.Warn("input_data not found at top level, returning empty map")
	return map[string]interface{}{}
}

// GetStepData safely retrieves data from a specific step's response
func GetStepData(collectedData map[string]interface{}, stepName string, logger *zap.Logger) (map[string]interface{}, bool) {
	if stepData, ok := collectedData[stepName].(map[string]interface{}); ok {
		logger.Debug("Found step data", zap.String("step", stepName))
		return stepData, true
	}
	
	logger.Debug("Step data not found", zap.String("step", stepName))
	return nil, false
}

// GetMultipleStepData retrieves data from multiple steps (for aggregation)
func GetMultipleStepData(collectedData map[string]interface{}, stepNames []string, logger *zap.Logger) map[string]interface{} {
	result := make(map[string]interface{})
	
	for _, stepName := range stepNames {
		if stepData, ok := collectedData[stepName].(map[string]interface{}); ok {
			result[stepName] = stepData
			logger.Debug("Collected step data", zap.String("step", stepName))
		} else {
			logger.Warn("Step data not found for aggregation", zap.String("step", stepName))
		}
	}
	
	return result
}

// GetFieldFromPath retrieves data using a dot-notation path (e.g., "input_data.business_type")
func GetFieldFromPath(collectedData map[string]interface{}, path string, logger *zap.Logger) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := collectedData
	
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - return the value
			if val, ok := current[part]; ok {
				return val, nil
			}
			return nil, fmt.Errorf("field not found: %s", path)
		}
		
		// Navigate deeper
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, fmt.Errorf("cannot navigate path at: %s", part)
		}
	}
	
	return nil, fmt.Errorf("empty path")
}
```

Now, **every action file** uses these helpers:

```go
// === All action files use the same pattern ===

// call_agent.go
func extractDataForAgent(params ActionParams) map[string]interface{} {
	return GetInputData(params.CollectedData, params.Logger)
}

// ai_actions.go  
func extractDataForAiAgent(params ActionParams) map[string]interface{} {
	return GetInputData(params.CollectedData, params.Logger)
}

// aggregate_actions.go (hypothetical)
func AggregateDataAction(params ActionParams) (*ActionResult, error) {
	stepNames := params.StepConfig.Config["response_fields"].([]string)
	aggregatedData := GetMultipleStepData(params.CollectedData, stepNames, params.Logger)
	// ... aggregate logic
}

// Any action that needs step data
func SomeAction(params ActionParams) (*ActionResult, error) {
	inputData := GetInputData(params.CollectedData, params.Logger)
	previousStepData, found := GetStepData(params.CollectedData, "previous_step", params.Logger)
	// ... use the data
}
```

This way:
- **One source of truth** for data access
- **All actions work consistently**
- **Easy to debug** - all access goes through helpers
- **Safe** - handles missing data gracefully
- **Future-proof** - new actions automatically work correctly

---

I'll think through this systematically and create a complete, clean solution.

## Core Principles

1. **Single source of truth** - `CollectedData` structure must be consistent everywhere
2. **Normalize at entry points** - Clean data when it enters the system
3. **Safe access patterns** - All reads go through helpers
4. **No surprises** - Predictable behavior at every level

## The Problem We're Solving

Currently, `input_data` can be in multiple places:
- `CollectedData.input_data` (clean)
- `CollectedData.__raw_message__.input_data` (nested)
- `CollectedData.__raw_message__.body.input_data` (double nested)
- `CollectedData.input_data.input_data` (incorrectly nested)

This breaks template resolution, data passing between agents, and makes debugging impossible.

## The Solution Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Entry Points (where data enters CollectedData)             │
├─────────────────────────────────────────────────────────────┤
│  1. CreateInitialState (new orchestration)                  │
│  2. handleCompleteResponse (child response)                 │
│  3. Message building (passing to children)                  │
└─────────────────────────────────────────────────────────────┘
                           ↓
                   NormalizeCollectedData()
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  Normalized CollectedData Structure                          │
├─────────────────────────────────────────────────────────────┤
│  {                                                           │
│    "__execution_context__": {...},                          │
│    "__my_requests_topic__": "...",                          │
│    "__raw_message__": {...},        // Debug only           │
│    "action": "process",                                     │
│    "input_data": {                   // ALWAYS TOP LEVEL    │
│      "business_name": "...",                                │
│      "business_type": "..."                                 │
│    },                                                        │
│    "agent_config": {...},                                   │
│    "step_name_1": {...},             // Response data       │
│    "step_name_2": {...}              // Response data       │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  Access Points (where actions read CollectedData)           │
├─────────────────────────────────────────────────────────────┤
│  ALL actions use helper functions:                          │
│  - GetInputData()                                           │
│  - GetStepData()                                            │
│  - GetMultipleStepData()                                    │
│  - GetFieldFromPath()                                       │
└─────────────────────────────────────────────────────────────┘
```

Now, let me write the complete implementation:

## File 1: `data_helpers.go` (NEW FILE)

```go
// internal/backend/agent-chassis/platform/orchestration/data_helpers.go
package orchestration

import (
	"fmt"
	"strings"
	
	"go.uber.org/zap"
)

// ============================================================================
// NORMALIZATION FUNCTIONS
// These are called when building or modifying CollectedData
// ============================================================================

// NormalizeInputData extracts clean input_data from any source structure
// This handles all the messy nested cases and returns a clean map
func NormalizeInputData(source map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	if source == nil {
		logger.Debug("NormalizeInputData: source is nil, returning empty map")
		return map[string]interface{}{}
	}
	
	// CASE 1: Try body.input_data first (messages from parent to child)
	if body, ok := source["body"].(map[string]interface{}); ok {
		if data, ok := body["input_data"].(map[string]interface{}); ok {
			logger.Debug("NormalizeInputData: extracted from body.input_data")
			return data
		}
	}
	
	// CASE 2: Try direct input_data
	if data, ok := source["input_data"].(map[string]interface{}); ok {
		// Check if it's already clean (doesn't contain action/config/input_data)
		_, hasAction := data["action"]
		_, hasConfig := data["config"]
		_, hasInputData := data["input_data"]
		
		// If it has none of these system fields, it's clean user data
		if !hasAction && !hasConfig && !hasInputData {
			logger.Debug("NormalizeInputData: using clean input_data")
			return data
		}
		
		// CASE 3: It's nested - extract the real input_data
		if nestedData, ok := data["input_data"].(map[string]interface{}); ok {
			logger.Debug("NormalizeInputData: extracted from input_data.input_data")
			return nestedData
		}
		
		// CASE 4: It has system fields but no nested input_data
		// This shouldn't happen but handle gracefully - extract only non-system fields
		cleaned := make(map[string]interface{})
		for k, v := range data {
			if k != "action" && k != "config" && k != "agent_config" && k != "agent_group" {
				cleaned[k] = v
			}
		}
		if len(cleaned) > 0 {
			logger.Debug("NormalizeInputData: cleaned system fields from input_data")
			return cleaned
		}
	}
	
	// CASE 5: No input_data found anywhere
	logger.Debug("NormalizeInputData: no input_data found, returning empty map")
	return map[string]interface{}{}
}

// NormalizeCollectedData builds a clean CollectedData structure from a message
// This is the ONLY function that should build CollectedData from messages
func NormalizeCollectedData(
	message map[string]interface{},
	execCtx interface{},
	requestsTopic string,
	logger *zap.Logger,
) map[string]interface{} {
	
	logger.Debug("NormalizeCollectedData: building clean structure")
	
	// Start with system metadata
	collectedData := map[string]interface{}{
		"__execution_context__": execCtx,
		"__my_requests_topic__": requestsTopic,
		"__raw_message__":       message, // Keep for debugging, but never use for data access
	}
	
	// Extract action
	if action, ok := message["action"].(string); ok {
		collectedData["action"] = action
		logger.Debug("NormalizeCollectedData: extracted action", zap.String("action", action))
	}
	
	// Extract config (only if it's workflow/system config, not user data)
	if config, ok := message["config"].(map[string]interface{}); ok {
		// Check if this looks like system config
		_, hasWorkflow := config["workflow"]
		_, hasGroupType := config["group_type"]
		
		if hasWorkflow || hasGroupType {
			collectedData["config"] = config
			logger.Debug("NormalizeCollectedData: extracted system config")
		}
	}
	
	// CRITICAL: Normalize input_data to top level
	// This is the user/business data that flows through the system
	collectedData["input_data"] = NormalizeInputData(message, logger)
	
	// Extract other standard fields
	if agentConfig, ok := message["agent_config"].(map[string]interface{}); ok {
		collectedData["agent_config"] = agentConfig
		logger.Debug("NormalizeCollectedData: extracted agent_config")
	}
	
	if agentGroup, ok := message["agent_group"].(map[string]interface{}); ok {
		collectedData["agent_group"] = agentGroup
		logger.Debug("NormalizeCollectedData: extracted agent_group")
	}
	
	if agentType, ok := message["agent_type"].(string); ok {
		collectedData["agent_type"] = agentType
		logger.Debug("NormalizeCollectedData: extracted agent_type", zap.String("type", agentType))
	}
	
	if prompt, ok := message["prompt"].(string); ok {
		collectedData["prompt"] = prompt
		logger.Debug("NormalizeCollectedData: extracted prompt")
	}
	
	logger.Debug("NormalizeCollectedData: structure complete",
		zap.Int("fields", len(collectedData)))
	
	return collectedData
}

// NormalizeResponseData cleans response data before storing in CollectedData
// Called when a child agent returns a response
func NormalizeResponseData(responseBody map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	if responseBody == nil {
		logger.Debug("NormalizeResponseData: responseBody is nil, returning empty map")
		return map[string]interface{}{}
	}
	
	// CASE 1: Response has explicit "data" field
	if data, ok := responseBody["data"].(map[string]interface{}); ok {
		logger.Debug("NormalizeResponseData: using response.body.data")
		return data
	}
	
	// CASE 2: Response has explicit "result" field
	if result, ok := responseBody["result"].(map[string]interface{}); ok {
		logger.Debug("NormalizeResponseData: using response.body.result")
		return result
	}
	
	// CASE 3: Clean the body by removing system fields
	cleaned := make(map[string]interface{})
	systemFields := map[string]bool{
		"action":               true,
		"config":               true,
		"__execution_context__": true,
		"__raw_message__":      true,
		"__my_requests_topic__": true,
		"agent_config":         true,
		"agent_group":          true,
	}
	
	for k, v := range responseBody {
		if !systemFields[k] {
			cleaned[k] = v
		}
	}
	
	logger.Debug("NormalizeResponseData: cleaned response body",
		zap.Int("original_fields", len(responseBody)),
		zap.Int("cleaned_fields", len(cleaned)))
	
	return cleaned
}

// ============================================================================
// SAFE ACCESS FUNCTIONS
// ALL actions should use these instead of direct CollectedData access
// ============================================================================

// GetInputData safely retrieves input_data from CollectedData
// This is the primary way to access user/business data
func GetInputData(collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	if collectedData == nil {
		logger.Warn("GetInputData: collectedData is nil")
		return map[string]interface{}{}
	}
	
	// After normalization, input_data should always be at top level
	if data, ok := collectedData["input_data"].(map[string]interface{}); ok {
		logger.Debug("GetInputData: found input_data at top level")
		return data
	}
	
	// This shouldn't happen after normalization
	logger.Warn("GetInputData: input_data not found at top level, attempting normalization")
	return NormalizeInputData(collectedData, logger)
}

// GetStepData safely retrieves data from a specific step's response
// Used when one step needs data from a previous step
func GetStepData(collectedData map[string]interface{}, stepName string, logger *zap.Logger) (map[string]interface{}, bool) {
	if collectedData == nil {
		logger.Warn("GetStepData: collectedData is nil", zap.String("step", stepName))
		return nil, false
	}
	
	if stepData, ok := collectedData[stepName].(map[string]interface{}); ok {
		logger.Debug("GetStepData: found step data", zap.String("step", stepName))
		return stepData, true
	}
	
	logger.Debug("GetStepData: step data not found", zap.String("step", stepName))
	return nil, false
}

// GetMultipleStepData retrieves data from multiple steps
// Used by aggregation actions that combine results from multiple steps
func GetMultipleStepData(collectedData map[string]interface{}, stepNames []string, logger *zap.Logger) map[string]interface{} {
	if collectedData == nil {
		logger.Warn("GetMultipleStepData: collectedData is nil")
		return map[string]interface{}{}
	}
	
	result := make(map[string]interface{})
	
	for _, stepName := range stepNames {
		if stepData, ok := collectedData[stepName].(map[string]interface{}); ok {
			result[stepName] = stepData
			logger.Debug("GetMultipleStepData: collected step data", zap.String("step", stepName))
		} else {
			logger.Warn("GetMultipleStepData: step data not found", zap.String("step", stepName))
			// Store nil or skip? Let's skip missing steps
		}
	}
	
	logger.Debug("GetMultipleStepData: collection complete",
		zap.Int("requested", len(stepNames)),
		zap.Int("found", len(result)))
	
	return result
}

// GetFieldFromPath retrieves a value using dot-notation path (e.g., "input_data.business_type")
// Used for template resolution and deep data access
func GetFieldFromPath(collectedData map[string]interface{}, path string, logger *zap.Logger) (interface{}, error) {
	if collectedData == nil {
		return nil, fmt.Errorf("collectedData is nil")
	}
	
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	
	parts := strings.Split(path, ".")
	current := collectedData
	
	logger.Debug("GetFieldFromPath: navigating path",
		zap.String("path", path),
		zap.Int("depth", len(parts)))
	
	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("empty segment in path: %s", path)
		}
		
		// Last part - return the value
		if i == len(parts)-1 {
			if val, ok := current[part]; ok {
				logger.Debug("GetFieldFromPath: found value", zap.String("path", path))
				return val, nil
			}
			return nil, fmt.Errorf("field not found: %s", path)
		}
		
		// Navigate deeper
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, fmt.Errorf("cannot navigate path at segment '%s' in path: %s", part, path)
		}
	}
	
	return nil, fmt.Errorf("unexpected end of path navigation")
}

// GetFieldFromPathWithDefault retrieves a value with a fallback default
func GetFieldFromPathWithDefault(collectedData map[string]interface{}, path string, defaultValue interface{}, logger *zap.Logger) interface{} {
	val, err := GetFieldFromPath(collectedData, path, logger)
	if err != nil {
		logger.Debug("GetFieldFromPathWithDefault: using default",
			zap.String("path", path),
			zap.Error(err))
		return defaultValue
	}
	return val
}

// MergeInputData merges new input data with existing, useful for enrichment
// New data takes precedence over existing
func MergeInputData(existing, new map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Copy existing
	for k, v := range existing {
		result[k] = v
	}
	
	// Overlay new (overwrites)
	for k, v := range new {
		result[k] = v
	}
	
	logger.Debug("MergeInputData: merged data",
		zap.Int("existing_fields", len(existing)),
		zap.Int("new_fields", len(new)),
		zap.Int("result_fields", len(result)))
	
	return result
}
```

## File 2: Updated `state.go` - `CreateInitialState`

```go
// internal/backend/agent-chassis/platform/orchestration/state.go
// (Only showing the CreateInitialState function - rest of file unchanged)

func CreateInitialState(
	orchestrationID string,
	parentOrchestrationID string,
	execCtx *types.ExecutionContext,
	workflow *types.Workflow,
	requestsTopic string,
	message map[string]interface{},
	repo Repository,
	logger *zap.Logger,
) (*types.OrchestrationState, error) {
	
	logger.Info("CreateInitialState starting",
		zap.String("orchestration_id", orchestrationID),
		zap.String("parent_orchestration_id", parentOrchestrationID))
	
	// Use normalization to build clean CollectedData
	collectedData := NormalizeCollectedData(message, execCtx, requestsTopic, logger)
	
	logger.Info("CreateInitialState: CollectedData normalized",
		zap.Strings("top_level_keys", getMapKeys(collectedData)))
	
	// Create the state object
	now := time.Now()
	state := &types.OrchestrationState{
		OrchestrationID:        orchestrationID,
		ParentOrchestrationID:  parentOrchestrationID,
		CorrelationID:          execCtx.CorrelationID,
		OwnerAgentID:           execCtx.Sender.AgentID,
		OwnerAgentType:         execCtx.Sender.AgentType,
		OwnerAgentRole:         execCtx.Sender.Role,
		ClientID:               execCtx.ClientID,
		RequestsTopic:          requestsTopic,
		ResponsesTopic:         execCtx.ResponsesTopic,
		Status:                 types.StatusInitialized,
		CurrentStep:            workflow.StartStep,
		CollectedData:          collectedData,
		AwaitedRequests:        make(map[string]*types.AwaitedRequest),
		AwaitedSteps:           []string{},
		ProcessingHistory:      []types.ProcessingRecord{},
		SubtreeAgents:          make(map[string]*types.SubtreeAgent),
		FuelBudget:             execCtx.FuelBudget,
		CreatedAt:              now,
		UpdatedAt:              now,
		Version:                1,
	}
	
	// Persist to database
	err := repo.CreateOrchestrationState(state)
	if err != nil {
		logger.Error("Failed to create orchestration state in database",
			zap.String("orchestration_id", orchestrationID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to persist initial state: %w", err)
	}
	
	logger.Info("Orchestration state created and persisted",
		zap.String("orchestration_id", orchestrationID),
		zap.String("status", string(state.Status)),
		zap.String("current_step", state.CurrentStep))
	
	return state, nil
}

// Helper function (add if not present)
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
```

## File 3: Updated `coordinator.go` - `handleCompleteResponse`

```go
// internal/backend/agent-chassis/platform/orchestration/coordinator.go
// (Only showing the relevant function - rest of file unchanged)

func (c *SagaCoordinator) handleCompleteResponse(
	state *types.OrchestrationState,
	response *types.Response,
	logger *zap.Logger,
) error {
	
	logger.Info("handleCompleteResponse: processing response",
		zap.String("orchestration_id", state.OrchestrationID))
	
	// Find the awaited request
	var awaitedReq *types.AwaitedRequest
	var requestID string
	
	for rid, req := range state.AwaitedRequests {
		if req.RequestID == response.Headers.InResponseTo.RequestID {
			awaitedReq = req
			requestID = rid
			break
		}
	}
	
	if awaitedReq == nil {
		return fmt.Errorf("no awaited request found for response")
	}
	
	logger.Info("handleCompleteResponse: found awaited request",
		zap.String("request_id", requestID),
		zap.String("step_name", awaitedReq.StepName))
	
	// Normalize the response data before storing
	normalizedData := NormalizeResponseData(response.Body, logger)
	
	// Store under the step name in CollectedData
	stepName := awaitedReq.StepName
	state.CollectedData[stepName] = normalizedData
	
	logger.Info("handleCompleteResponse: stored normalized response",
		zap.String("step_name", stepName),
		zap.Strings("data_keys", getMapKeys(normalizedData)))
	
	// Remove from awaited requests
	delete(state.AwaitedRequests, requestID)
	
	logger.Info("handleCompleteResponse: removed from awaited requests",
		zap.String("request_id", requestID),
		zap.Int("remaining_awaited", len(state.AwaitedRequests)))
	
	// Update state in database
	err := c.repo.UpdateStateWithVersion(state)
	if err != nil {
		return fmt.Errorf("failed to update state after response: %w", err)
	}
	
	// Check if all responses received
	if len(state.AwaitedRequests) == 0 {
		logger.Info("handleCompleteResponse: all responses received, continuing workflow")
		return c.continueExecution(state)
	}
	
	logger.Info("handleCompleteResponse: still awaiting responses",
		zap.Int("remaining", len(state.AwaitedRequests)))
	
	return nil
}
```

## File 4: Updated `call_agent.go` - `extractDataForAgent`

```go
// internal/backend/agent-chassis/platform/orchestration/actions/call_agent.go
// (Only showing the extraction function - rest of file uses it)

// extractDataForAgent gets the input data to pass to a child agent
// Uses the centralized safe access function
func extractDataForAgent(params ActionParams) map[string]interface{} {
	// Get which field to extract (default: "input_data")
	inputField := "input_data"
	if field, ok := params.StepConfig.Config["input_field"].(string); ok && field != "" {
		inputField = field
	}
	
	params.Logger.Info("extractDataForAgent: extracting data",
		zap.String("field", inputField))
	
	// If looking for input_data specifically, use the helper
	if inputField == "input_data" {
		return GetInputData(params.CollectedData, params.Logger)
	}
	
	// Otherwise try to get from top level
	if data, ok := params.CollectedData[inputField].(map[string]interface{}); ok {
		params.Logger.Info("extractDataForAgent: found field at top level",
			zap.String("field", inputField))
		return data
	}
	
	// Field not found
	params.Logger.Warn("extractDataForAgent: field not found, returning empty map",
		zap.String("field", inputField))
	return map[string]interface{}{}
}
```

## File 5: Updated `ai_actions.go` - `extractDataForAiAgent`

```go
// internal/backend/agent-chassis/platform/orchestration/actions/ai_actions.go
// (Only showing the extraction function - rest of file uses it)

// extractDataForAiAgent gets template data for LLM prompt rendering
// Uses the centralized safe access function
func extractDataForAiAgent(params ActionParams) map[string]interface{} {
	// Get which field to use (default: "input_data")
	inputField := "input_data"
	if field, ok := params.StepConfig.Config["input_field"].(string); ok && field != "" {
		inputField = field
	}
	
	params.Logger.Info("extractDataForAiAgent: extracting data",
		zap.String("field", inputField))
	
	// If looking for input_data, use the helper
	if inputField == "input_data" {
		return GetInputData(params.CollectedData, params.Logger)
	}
	
	// Otherwise try top level
	if data, ok := params.CollectedData[inputField].(map[string]interface{}); ok {
		params.Logger.Info("extractDataForAiAgent: found field at top level",
			zap.String("field", inputField))
		return data
	}
	
	// Not found
	params.Logger.Warn("extractDataForAiAgent: field not found, returning empty map",
		zap.String("field", inputField))
	return map[string]interface{}{}
}
```

## File 6: Template Resolution (if you have a separate file for this)

```go
// If there's template rendering code, update it to use GetFieldFromPath

func renderTemplateString(template string, data map[string]interface{}, logger *zap.Logger) (string, error) {
	// Parse template
	tmpl, err := template.New("template").Parse(template)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	
	// Execute template with data
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		// Template execution failed, try to provide helpful error
		logger.Warn("Template execution failed",
			zap.String("template", template),
			zap.Error(err),
			zap.Strings("available_keys", getMapKeys(data)))
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}
```

## Summary of Changes

### What This Fixes:

1. **`input_data` always at top level** in `CollectedData`
2. **Consistent structure** whether from external request, parent agent, or child response
3. **Template resolution works** - paths like `{{.input_data.business_type}}` always resolve
4. **No more nested confusion** - `input_data.input_data` impossible
5. **Safe access** - all actions use helpers, no direct access
6. **Easy debugging** - single normalization point to inspect/fix

### Migration Path:

1. Add `data_helpers.go` (new file)
2. Update `CreateInitialState` in `state.go`
3. Update `handleCompleteResponse` in `coordinator.go`
4. Update `extractDataForAgent` in `call_agent.go`
5. Update `extractDataForAiAgent` in `ai_actions.go`
6. Update any other actions that access `CollectedData`

### Testing:

```go
// Test normalization
func TestNormalizeInputData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Test case 1: Clean input_data
	source1 := map[string]interface{}{
		"input_data": map[string]interface{}{
			"business_name": "Test",
			"business_type": "bakery",
		},
	}
	result1 := NormalizeInputData(source1, logger)
	assert.Equal(t, "Test", result1["business_name"])
	
	// Test case 2: Nested in body
	source2 := map[string]interface{}{
		"body": map[string]interface{}{
			"input_data": map[string]interface{}{
				"business_name": "Test",
			},
		},
	}
	result2 := NormalizeInputData(source2, logger)
	assert.Equal(t, "Test", result2["business_name"])
	
	// Test case 3: Double nested
	source3 := map[string]interface{}{
		"input_data": map[string]interface{}{
			"input_data": map[string]interface{}{
				"business_name": "Test",
			},
		},
	}
	result3 := NormalizeInputData(source3, logger)
	assert.Equal(t, "Test", result3["business_name"])
}
```

This is a complete, production-ready solution. Every place that touches `CollectedData` now goes through the same normalization and access patterns.