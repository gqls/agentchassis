Long-Term Plan (implementation_plan.md)
Phase 1: Input Mapping + Contracts

Replace input_fields with explicit input_mapping
Add contract validation (hard fail on missing required fields)
Add deprecation warnings for __raw_message__ fallbacks
Eventually remove __raw_message__ accumulation

Phase 2: Sequential Fan-Out

New sequential_fan_out action
Spawns one child at a time, waits, spawns next
Each child is separate orchestration (visible in logs)

Phase 3: page-builder Agent

New worker agent that builds ONE page
Uses existing agents via target_role (deployer-agent for git, etc.)
Conditional image generation

Phase 4: Update pageflow-builder

Replace build_pages_loop with build_pages using sequential_fan_out
Point to page-builder as child agent type

Key Decisions Made

Fail fast on child errors (simpler, can add resilience later)
Use agents for external operations (deployer-agent for git, not local git_commit action)
Use parent's spawned agents via target_role routing (page-builder doesn't spawn its own)
Remove __raw_message__ gradually with deprecation warnings first

--------

# Simplified Iteration Model - Implementation Plan

## Overview

Replace complex loop/substep injection with:
1. Explicit `input_mapping` on all agent calls (no path hunting)
2. Runtime contract validation (hard fail on missing fields)
3. `sequential_fan_out` action that spawns independent child orchestrations
4. `page-builder` worker agent with its own simple workflow

## Current vs Proposed Architecture

### Current (Complex)
```
pageflow-builder orchestration
├── Step: build_pages_loop (action: loop)
│   ├── Injects: build_pages_loop_iter_0_write_page_content
│   ├── Injects: build_pages_loop_iter_0_review_page_content
│   ├── Injects: build_pages_loop_iter_0_assemble_page
│   ├── Injects: build_pages_loop_iter_0_deploy_page
│   ├── Injects: build_pages_loop_iter_0_update_page_status
│   ├── Injects: build_pages_loop_iter_0_complete_page
│   ├── Injects: build_pages_loop_iter_1_write_page_content
│   └── ... (6 steps × N pages = many injected steps)
└── All share ONE orchestration_states row
```

### Proposed (Simple)
```
pageflow-builder orchestration (orch-parent)
├── Step: build_pages (action: sequential_fan_out)
│   ├── Spawns: page-builder (orch-child-0) for page "home"
│   ├── Waits for orch-child-0 to complete
│   ├── Spawns: page-builder (orch-child-1) for page "about"
│   ├── Waits for orch-child-1 to complete
│   └── ... continues sequentially
└── Collects results, continues to update_site_status

page-builder orchestration (orch-child-0)
├── write_content → call page-content-writer
├── review_content → call content-reviewer
├── assemble → assemble_page action
├── deploy → git_commit action
├── update_status → update_page_status action
└── complete → return result to parent
```

Each child has its own orchestration_states row, own logs, independent lifecycle.

---

## Phase 1: Input Mapping + Contract Validation

### 1.1 Schema Changes

**agent_definitions table** - enforce contract structure:
```sql
-- Add constraint to validate contract JSON structure
ALTER TABLE agent_definitions 
ADD CONSTRAINT valid_input_contract 
CHECK (
  input_contract IS NULL OR (
    input_contract ? 'required' AND 
    jsonb_typeof(input_contract->'required') = 'array'
  )
);
```

### 1.2 New Types (coordinator.go or types.go)

```go
// InputMapping defines explicit source paths for input fields
// Key = destination field name (what child receives)
// Value = source path in CollectedData (where to get it)
type InputMapping map[string]string

// InputContract defines what an agent expects
type InputContract struct {
    Required []string `json:"required"`
    Optional []string `json:"optional,omitempty"`
}

// OutputContract defines what an agent produces  
type OutputContract struct {
    Produces []string `json:"produces"`
}
```

### 1.3 New Functions (data_helpers.go or new file input_mapping.go)

```go
// ResolveInputMapping builds input data using explicit paths
// Returns error if any mapping path not found (hard fail)
func ResolveInputMapping(
    collectedData map[string]interface{},
    mapping InputMapping,
    logger *zap.Logger,
) (map[string]interface{}, error) {
    result := make(map[string]interface{})
    
    for destField, sourcePath := range mapping {
        // Handle special $item token (for fan_out)
        if sourcePath == "$item" {
            // This will be handled by the fan_out action
            continue
        }
        
        value, found := getValueAtExactPath(collectedData, sourcePath)
        if !found {
            availablePaths := listAvailablePaths(collectedData, 2) // depth 2
            return nil, fmt.Errorf(
                "input_mapping failed: source path '%s' not found for field '%s'\nAvailable paths: %v",
                sourcePath, destField, availablePaths,
            )
        }
        result[destField] = value
    }
    
    return result, nil
}

// getValueAtExactPath retrieves value at exact dot-notation path
// No fallbacks, no hunting - just the exact path
func getValueAtExactPath(data map[string]interface{}, path string) (interface{}, bool) {
    parts := strings.Split(path, ".")
    current := interface{}(data)
    
    for _, part := range parts {
        switch v := current.(type) {
        case map[string]interface{}:
            val, exists := v[part]
            if !exists {
                return nil, false
            }
            current = val
        default:
            return nil, false
        }
    }
    
    return current, true
}

// ValidateInputContract checks that data satisfies agent's input contract
func ValidateInputContract(
    agentType string,
    data map[string]interface{},
    contract InputContract,
) error {
    var missing []string
    
    for _, required := range contract.Required {
        if _, exists := data[required]; !exists {
            missing = append(missing, required)
        }
    }
    
    if len(missing) > 0 {
        return fmt.Errorf(
            "contract violation for agent '%s': missing required fields: %v\nProvided fields: %v",
            agentType, missing, mapKeys(data),
        )
    }
    
    return nil
}

// listAvailablePaths returns available paths in data for error messages
func listAvailablePaths(data map[string]interface{}, maxDepth int) []string {
    var paths []string
    var walk func(prefix string, d map[string]interface{}, depth int)
    
    walk = func(prefix string, d map[string]interface{}, depth int) {
        for k, v := range d {
            path := k
            if prefix != "" {
                path = prefix + "." + k
            }
            paths = append(paths, path)
            
            if depth < maxDepth {
                if nested, ok := v.(map[string]interface{}); ok {
                    walk(path, nested, depth+1)
                }
            }
        }
    }
    
    walk("", data, 0)
    sort.Strings(paths)
    return paths
}

func mapKeys(m map[string]interface{}) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    return keys
}
```

### 1.4 Modify call_agent Action (actions/call_agent.go)

```go
func CallAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
    config := params.StepConfig.Config
    agentType := config["agent_type"].(string)
    
    // NEW: Check for input_mapping (preferred) vs input_fields (legacy)
    var inputData map[string]interface{}
    var err error
    
    if mapping, ok := config["input_mapping"].(map[string]interface{}); ok {
        // New explicit mapping approach
        inputMapping := make(InputMapping)
        for k, v := range mapping {
            inputMapping[k] = v.(string)
        }
        
        inputData, err = ResolveInputMapping(params.CollectedData, inputMapping, params.Logger)
        if err != nil {
            return nil, fmt.Errorf("step %s: %w", params.StepConfig.Name, err)
        }
    } else if inputFields, ok := config["input_fields"].([]interface{}); ok {
        // Legacy approach - log deprecation warning
        params.Logger.Warn("DEPRECATED: using input_fields instead of input_mapping",
            zap.String("step", params.StepConfig.Name),
            zap.String("agent_type", agentType),
        )
        inputData = extractDataForAgentLegacy(params, inputFields)
    } else {
        return nil, fmt.Errorf("step %s: no input_mapping or input_fields specified", params.StepConfig.Name)
    }
    
    // NEW: Validate against child agent's input contract
    childDef, err := getAgentDefinition(ctx, agentType, params.DB)
    if err != nil {
        return nil, fmt.Errorf("failed to get agent definition for %s: %w", agentType, err)
    }
    
    if childDef.InputContract != nil {
        if err := ValidateInputContract(agentType, inputData, *childDef.InputContract); err != nil {
            return nil, err
        }
    }
    
    // Send ONLY the mapped data - no __raw_message__, no __execution_context__
    // Child will build its own execution context from message headers
    requestID := uuid.New().String()
    
    // ... rest of existing call_agent logic (send message, return await_response) ...
}
```

### 1.5 Modify spawn_agent Similarly

Same pattern - use `input_mapping` if present, validate contract, send clean data.

### 1.6 Remove __raw_message__ Accumulation

The `__raw_message__` is currently added at each level and used as a fallback when data isn't found at expected paths. With explicit input_mapping, this fallback becomes unnecessary.

**Where it's added (data_helpers.go):**
```go
// REMOVE THIS:
if _, exists := msgCtx.CollectedData["__raw_message__"]; !exists {
    msgCtx.CollectedData["__raw_message__"] = rawMsg
}
```

**Where it's used as fallback (multiple files):**
```go
// These fallback patterns become unnecessary:
searchPaths := []string{
    "input_data." + fieldName,
    "__raw_message__." + fieldName,  // REMOVE
}

// And:
if raw, ok := params.CollectedData["__raw_message__"].(map[string]interface{}); ok {
    if val, exists := raw[fieldName]; exists {
        // REMOVE this fallback
    }
}
```

**Migration Strategy:**

1. **First**: Add input_mapping support with contract validation (new code path)
2. **Then**: Log warnings when __raw_message__ fallback is used:
   ```go
   if val := findInRawMessage(collectedData, fieldName); val != nil {
       logger.Warn("DEPRECATED: Found field via __raw_message__ fallback - use input_mapping instead",
           zap.String("field", fieldName),
           zap.String("step", stepName),
       )
       return val
   }
   ```
3. **Monitor**: Check logs for these warnings, fix workflows that rely on fallback
4. **Finally**: Remove __raw_message__ addition and fallback code

**Files to Modify:**
- `platform/orchestration/data_helpers.go` - Remove __raw_message__ accumulation
- `platform/actions/call_agent.go` - Remove __raw_message__ fallback searches
- `platform/actions/spawn_agent.go` - Remove __raw_message__ fallback searches
- `platform/orchestration/coordinator.go` - Any __raw_message__ fallbacks

**Benefit:**
- Smaller messages (no duplicate data at each level)
- Clearer data flow (data is where input_mapping says it is)
- Faster path resolution (no hunting through fallbacks)

### 1.7 Add Contracts to Existing Agents

For contract validation to work, we need to add `input_contract` to the agents that page-builder will call. These can be added incrementally.

**page-content-writer:**
```sql
UPDATE agent_definitions 
SET input_contract = '{
    "required": ["current_page", "site_record", "reviewed_brief"],
    "optional": ["style_collection", "db_sync", "generated_images"]
}'::jsonb
WHERE type = 'page-content-writer';
```

**content-reviewer:**
```sql
UPDATE agent_definitions 
SET input_contract = '{
    "required": ["current_page", "page_content"],
    "optional": ["reviewed_brief"]
}'::jsonb
WHERE type = 'content-reviewer';
```

**deployer-agent:**
```sql
UPDATE agent_definitions 
SET input_contract = '{
    "required": ["site_record"],
    "optional": ["page", "assembled_page", "pages_built"]
}'::jsonb
WHERE type = 'deployer-agent';
```

**image-generator:**
```sql
UPDATE agent_definitions 
SET input_contract = '{
    "required": ["page"],
    "optional": ["site_record", "reviewed_brief", "prompt"]
}'::jsonb
WHERE type = 'image-generator';
```

**Note:** These contracts define what the agent EXPECTS to receive. The `input_mapping` in the calling step defines WHERE to get each field. Contract validation ensures the mapping produces all required fields before sending.

---

## Phase 2: Sequential Fan-Out Action

### 2.1 Fan-Out State Structure

```go
// FanOutState tracks progress of sequential fan-out
type FanOutState struct {
    StepName       string                   `json:"step_name"`
    TotalItems     int                      `json:"total_items"`
    CurrentIndex   int                      `json:"current_index"`
    Results        []interface{}            `json:"results"`
    ChildAgentType string                   `json:"child_agent_type"`
    CurrentChildID string                   `json:"current_child_id,omitempty"`
    PendingRequest string                   `json:"pending_request,omitempty"`
    Status         string                   `json:"status"` // "running", "complete", "failed"
    Error          string                   `json:"error,omitempty"`
}

// Store in CollectedData under key like "__fan_out_state_build_pages"
```

### 2.2 Sequential Fan-Out Action (actions/sequential_fan_out.go)

```go
func SequentialFanOutAction(ctx context.Context, params ActionParams) (interface{}, error) {
    config := params.StepConfig.Config
    stepName := params.StepConfig.Name
    
    itemsField := config["items_field"].(string)
    childAgentType := config["child_agent_type"].(string)
    inputMappingRaw := config["input_mapping"].(map[string]interface{})
    
    // Convert input mapping
    inputMapping := make(InputMapping)
    for k, v := range inputMappingRaw {
        inputMapping[k] = v.(string)
    }
    
    // Get items to iterate
    itemsRaw, found := getValueAtExactPath(params.CollectedData, itemsField)
    if !found {
        return nil, fmt.Errorf("items_field '%s' not found", itemsField)
    }
    items, ok := itemsRaw.([]interface{})
    if !ok {
        return nil, fmt.Errorf("items_field '%s' is not an array", itemsField)
    }
    
    // Get or initialize fan-out state
    stateKey := "__fan_out_state_" + stepName
    var fanOutState *FanOutState
    
    if existing, ok := params.CollectedData[stateKey].(map[string]interface{}); ok {
        // Resuming after child response
        fanOutState = parseFanOutState(existing)
    } else {
        // First call - initialize
        fanOutState = &FanOutState{
            StepName:       stepName,
            TotalItems:     len(items),
            CurrentIndex:   0,
            Results:        make([]interface{}, len(items)),
            ChildAgentType: childAgentType,
            Status:         "running",
        }
    }
    
    // Check if we're done
    if fanOutState.CurrentIndex >= fanOutState.TotalItems {
        params.Logger.Info("Sequential fan-out complete",
            zap.String("step", stepName),
            zap.Int("total_items", fanOutState.TotalItems),
        )
        return map[string]interface{}{
            "results":  fanOutState.Results,
            "complete": true,
        }, nil
    }
    
    // Build input for current item
    currentItem := items[fanOutState.CurrentIndex]
    
    inputData, err := ResolveInputMapping(params.CollectedData, inputMapping, params.Logger)
    if err != nil {
        return nil, err
    }
    
    // Replace $item placeholder with current item
    for k, v := range inputMapping {
        if v == "$item" {
            inputData[k] = currentItem
        }
    }
    
    // Validate against child's contract
    childDef, err := getAgentDefinition(ctx, childAgentType, params.DB)
    if err != nil {
        return nil, err
    }
    if childDef.InputContract != nil {
        if err := ValidateInputContract(childAgentType, inputData, *childDef.InputContract); err != nil {
            return nil, fmt.Errorf("iteration %d: %w", fanOutState.CurrentIndex, err)
        }
    }
    
    // Spawn child agent
    requestID := uuid.New().String()
    childOrchID := uuid.New().String()
    
    params.Logger.Info("Sequential fan-out spawning child",
        zap.String("step", stepName),
        zap.Int("iteration", fanOutState.CurrentIndex),
        zap.Int("total", fanOutState.TotalItems),
        zap.String("child_agent_type", childAgentType),
        zap.String("child_orchestration_id", childOrchID),
        zap.String("request_id", requestID),
    )
    
    // Update fan-out state
    fanOutState.PendingRequest = requestID
    fanOutState.CurrentChildID = childOrchID
    
    // Store updated state
    params.State.CollectedData[stateKey] = fanOutState
    
    // Send to child (similar to spawn_agent)
    err = sendSpawnRequest(ctx, params, childAgentType, childOrchID, requestID, inputData)
    if err != nil {
        return nil, err
    }
    
    return map[string]interface{}{
        "await_response":    true,
        "request_id":        requestID,
        "target_agent_type": childAgentType,
        "target_agent_id":   childOrchID,
        "fan_out":           true,
        "iteration_index":   fanOutState.CurrentIndex,
        "total_iterations":  fanOutState.TotalItems,
    }, nil
}
```

### 2.3 Fan-Out Response Handling (coordinator.go)

When a fan-out child completes, we need to:
1. Store result at correct index
2. Increment index
3. Re-execute the step for next item (or continue if done)

```go
// In handleCompleteResponse or similar

func (s *SagaCoordinator) handleFanOutResponse(
    ctx context.Context,
    state *OrchestrationState,
    stepName string,
    response interface{},
    logger *zap.Logger,
) error {
    stateKey := "__fan_out_state_" + stepName
    
    fanOutStateRaw, ok := state.CollectedData[stateKey].(map[string]interface{})
    if !ok {
        return fmt.Errorf("fan_out state not found for step %s", stepName)
    }
    fanOutState := parseFanOutState(fanOutStateRaw)
    
    // Store result at current index
    fanOutState.Results[fanOutState.CurrentIndex] = response
    fanOutState.CurrentIndex++
    fanOutState.PendingRequest = ""
    fanOutState.CurrentChildID = ""
    
    logger.Info("Fan-out iteration complete",
        zap.String("step", stepName),
        zap.Int("completed_index", fanOutState.CurrentIndex-1),
        zap.Int("total", fanOutState.TotalItems),
    )
    
    // Update state
    state.CollectedData[stateKey] = fanOutState
    
    // Check if all done
    if fanOutState.CurrentIndex >= fanOutState.TotalItems {
        // All iterations complete - store final results in output_field
        stepConfig := state.WorkflowPlan.Steps[stepName]
        if stepConfig.OutputField != "" {
            state.CollectedData[stepConfig.OutputField] = fanOutState.Results
        }
        
        // Clean up fan-out state
        delete(state.CollectedData, stateKey)
        
        // Continue to next step
        return nil // Normal flow continues
    }
    
    // More items - re-execute the step
    // This is key: we DON'T advance CurrentStep, we re-run the same step
    // The step will pick up the incremented CurrentIndex from fan-out state
    state.Status = StatusExecutingStep // Reset to executing
    
    return nil
}
```

### 2.4 Integrate into Existing Response Flow

Modify `processAwaitResponse` or `handleCompleteResponse` to detect fan-out:

```go
// When processing a response, check if it's for a fan-out step
if isFanOutStep(state, stepName) {
    return s.handleFanOutResponse(ctx, state, stepName, response, logger)
}
```

---

## Phase 3: page-builder Agent Definition

### 3.1 Design Decisions

**Using Existing Agents:**
- `page-content-writer` - already handles research internally (spawns research-agent)
- `content-reviewer` - reviews content
- `deployer-agent` - handles git operations (allows switching GitHub→GitLab by updating one agent)
- `image-generator` - generates images when needed

**Local Actions:**
- `assemble_page` - assembles HTML from components
- `update_page_status` - marks page as deployed in database

**Why deployer-agent instead of git_commit action:**
If you switch git providers (GitHub → GitLab), you update deployer-agent's adapter config, not every workflow that commits. The abstraction is at the agent level.

### 3.2 SQL to Insert Agent Definition

```sql
INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    version,
    input_contract,
    output_contract,
    default_config,
    category,
    capabilities,
    created_at,
    updated_at
) VALUES (
    'page-builder',
    'Page Builder',
    'Builds a single page: researches, writes content, reviews, assembles, deploys via deployer-agent',
    1,
    '{
        "required": ["page", "site_record", "reviewed_brief"],
        "optional": ["style_collection", "db_sync"]
    }'::jsonb,
    '{
        "produces": ["page_content", "assembled_page", "deploy_result", "page_status"]
    }'::jsonb,
    '{
        "processing_mode": "orchestrator",
        "timeout_seconds": 600,
        "workflow": {
            "start_step": "check_images_needed",
            "steps": {
                "check_images_needed": {
                    "action": "conditional",
                    "description": "Check if page needs generated images",
                    "config": {
                        "condition": "page.needs_images == true",
                        "then_step": "generate_images",
                        "else_step": "write_content"
                    }
                },
                "generate_images": {
                    "action": "call_agent",
                    "description": "Generate images for this page",
                    "config": {
                        "agent_type": "image-generator",
                        "target_role": "image_generator",
                        "input_mapping": {
                            "page": "page",
                            "site_record": "site_record",
                            "reviewed_brief": "reviewed_brief"
                        },
                        "timeout_seconds": 120
                    },
                    "output_field": "generated_images",
                    "next_step": "write_content"
                },
                "write_content": {
                    "action": "call_agent",
                    "description": "Write content for this page (includes research)",
                    "config": {
                        "agent_type": "page-content-writer",
                        "target_role": "content_writer",
                        "input_mapping": {
                            "current_page": "page",
                            "site_record": "site_record",
                            "reviewed_brief": "reviewed_brief",
                            "style_collection": "style_collection",
                            "db_sync": "db_sync",
                            "generated_images": "generated_images"
                        },
                        "timeout_seconds": 300
                    },
                    "output_field": "page_content",
                    "next_step": "review_content"
                },
                "review_content": {
                    "action": "call_agent",
                    "description": "Review page content",
                    "config": {
                        "agent_type": "content-reviewer",
                        "target_role": "reviewer",
                        "input_mapping": {
                            "current_page": "page",
                            "page_content": "page_content",
                            "reviewed_brief": "reviewed_brief"
                        },
                        "timeout_seconds": 300
                    },
                    "output_field": "reviewed_content",
                    "next_step": "assemble"
                },
                "assemble": {
                    "action": "assemble_page",
                    "description": "Assemble full page HTML from components",
                    "config": {
                        "content_field": "page_content.response.page_html",
                        "add_navigation": false
                    },
                    "output_field": "assembled_page",
                    "next_step": "deploy"
                },
                "deploy": {
                    "action": "call_agent",
                    "description": "Deploy page via deployer-agent (git commit)",
                    "config": {
                        "agent_type": "deployer-agent",
                        "target_role": "deployer",
                        "input_mapping": {
                            "page": "page",
                            "site_record": "site_record",
                            "assembled_page": "assembled_page"
                        },
                        "timeout_seconds": 180
                    },
                    "output_field": "deploy_result",
                    "next_step": "update_status"
                },
                "update_status": {
                    "action": "update_page_status",
                    "description": "Mark page as deployed in database",
                    "config": {
                        "status": "deployed",
                        "commit_from": "deploy_result.commit_sha",
                        "page_id_field": "page.id"
                    },
                    "output_field": "page_status",
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow",
                    "description": "Page build complete",
                    "config": {
                        "output_fields": ["page_content", "assembled_page", "deploy_result", "page_status"]
                    }
                }
            }
        }
    }'::jsonb,
    'worker',
    '["page-building", "content-orchestration"]'::jsonb,
    NOW(),
    NOW()
);
```

### 3.3 Note on Agent Spawning

page-builder does NOT pre-spawn agents. It uses agents already spawned by pageflow-builder via `target_role` routing:
- pageflow-builder spawns: content_writer, reviewer, deployer, image_generator
- page-builder calls them using `target_role` to find the existing instances
- Same correlation_id flows through, so role lookup works

---

## Phase 4: Update pageflow-builder Workflow

### 4.1 Replace build_pages_loop with sequential_fan_out

The fan-out needs to pass everything page-builder requires:

```sql
-- First, add the new build_pages step
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,build_pages}',
    '{
        "action": "sequential_fan_out",
        "description": "Build each page via page-builder worker",
        "config": {
            "items_field": "pages_to_build.pages",
            "child_agent_type": "page-builder",
            "input_mapping": {
                "page": "$item",
                "site_record": "site_record",
                "reviewed_brief": "reviewed_brief",
                "style_collection": "style_collection",
                "db_sync": "db_sync"
            }
        },
        "output_field": "pages_built",
        "next_step": "update_site_status"
    }'::jsonb
)
WHERE type = 'pageflow-builder';

-- Remove the old build_pages_loop step
UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,build_pages_loop}'
WHERE type = 'pageflow-builder';

-- Update get_pages_to_build to point to new step name
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,get_pages_to_build,next_step}',
    '"build_pages"'
)
WHERE type = 'pageflow-builder';
```

### 4.2 Data Flow Verification

Before the fan-out, pageflow-builder's CollectedData should have:
- `pages_to_build.pages` - array of page objects from `get_pages_to_build` action
- `site_record` - from `ensure_site_record` action
- `reviewed_brief` - from input_data (passed from briefing-agent)
- `style_collection` - from `select_style_collection` action
- `db_sync` - from `sync_pages_to_db` action

Each page-builder child receives:
- `page` - the specific page object (from `$item`)
- `site_record`, `reviewed_brief`, `style_collection`, `db_sync` - passed through

### 4.3 Spawned Agents Available to page-builder

pageflow-builder spawns these agents before build_pages:
- `spawn_planner` → site-planner (role: planner)
- `spawn_content_writer` → page-content-writer (role: content_writer)
- `spawn_reviewer` → content-reviewer (role: reviewer)
- `spawn_deployer` → deployer-agent (role: deployer)
- `spawn_image_generator` → image-generator (role: image_generator) [conditional]

page-builder uses `target_role` to call these existing agents. The same `correlation_id` flows through, so role lookup finds them.

---

## Phase 5: Migration & Rollback Plan

### 5.1 Feature Flag Approach

Add config to enable new behavior per-agent:

```go
// In agent definition default_config
{
    "use_input_mapping": true,     // Enable new mapping (default false initially)
    "validate_contracts": true,    // Enable contract validation
    "use_fan_out": true           // Use fan_out instead of loop
}
```

### 5.2 Gradual Rollout

1. Deploy code with all new features behind flags (all false)
2. Enable `use_input_mapping` + `validate_contracts` for one agent
3. Test thoroughly
4. Enable for more agents
5. Finally enable `use_fan_out` for pageflow-builder
6. Monitor for issues
7. Remove flags and legacy code

### 5.3 Rollback

If issues arise:
- Flip flags back to false
- Old loop code still exists and works
- No data migration needed

---

## Testing Checklist

### Unit Tests
- [ ] ResolveInputMapping with valid paths
- [ ] ResolveInputMapping with missing path (should error)
- [ ] ValidateInputContract with all required fields
- [ ] ValidateInputContract with missing required field (should error)
- [ ] FanOutState serialization/deserialization
- [ ] Sequential fan-out with 0 items (should skip)
- [ ] Sequential fan-out with 1 item
- [ ] Sequential fan-out with N items

### Integration Tests
- [ ] End-to-end pageflow-builder with fan_out
- [ ] page-builder completes successfully
- [ ] Parent collects all results
- [ ] Error in one child doesn't break others (or does, depending on design)
- [ ] Contract validation catches missing field early

### Log Verification
- [ ] Each child has distinct orchestration_id
- [ ] Parent logs show "spawning child" for each iteration
- [ ] Child logs show complete workflow execution
- [ ] Easy to grep by orchestration_id to see full flow

---

## Expected Outcomes

### Before (Current Logs)
```
orchestration=478f0f11 step=build_pages_loop_iter_0_write_page_content msg="executing step"
orchestration=478f0f11 step=build_pages_loop_iter_0_write_page_content msg="awaiting response"
orchestration=478f0f11 step=build_pages_loop_iter_0_review_page_content msg="executing step"
... (hard to follow, all same orchestration ID)
```

### After (Simplified Logs)
```
orchestration=478f0f11 step=build_pages msg="sequential_fan_out starting" items=5
orchestration=478f0f11 step=build_pages msg="spawning child" iteration=0 child_orch=abc123
orchestration=abc123 agent=page-builder step=write_content msg="executing step"
orchestration=abc123 agent=page-builder step=write_content msg="awaiting response"
orchestration=abc123 agent=page-builder step=review_content msg="executing step"
orchestration=abc123 agent=page-builder step=complete msg="workflow complete"
orchestration=478f0f11 step=build_pages msg="child complete" iteration=0
orchestration=478f0f11 step=build_pages msg="spawning child" iteration=1 child_orch=def456
orchestration=def456 agent=page-builder step=write_content msg="executing step"
... (each child clearly separated by orchestration ID)
```

### Error Messages Before
```
ERROR: site_id not found at site_record.site_id
```

### Error Messages After
```
ERROR: step write_content calling page-content-writer: 
  contract violation for agent 'page-content-writer': missing required fields: [site_record]
  Provided fields: [page, reviewed_brief, style_collection]
  
  Hint: Check input_mapping in step 'write_content' of agent 'page-builder'
```

---

## Open Questions / Design Decisions

### Q1: Error Handling in Fan-Out - DECIDED: Fail Fast

When one child fails, stop the fan-out immediately. Parent reports error, no more children spawned.

Rationale: Simpler to implement and debug. If page 2 fails, there's likely a systemic issue that would cause pages 3-5 to fail too. Can add "continue on error" mode later if needed.

### Q2: Agent Usage - DECIDED: Use Agents for External Operations

page-builder uses agents (not local actions) for:
- **deployer-agent** for git operations - allows switching GitHub→GitLab by updating one agent
- **image-generator** for image creation
- **page-content-writer** for content (which internally uses research-agent)
- **content-reviewer** for review

Local actions only for:
- `assemble_page` - internal HTML assembly
- `update_page_status` - database update

### Q3: Spawned Agents - DECIDED: Use Parent's Spawned Agents

page-builder does NOT spawn its own agents. It uses agents already spawned by pageflow-builder via `target_role` routing. This avoids spawning duplicate agents for each page.

### Q4: Re-executing Steps for Fan-Out

When a fan-out child responds, we need to re-execute the same step (not advance to next).

Current flow: response → remove awaited request → check allDone → if done, advance to next step

Fan-out flow: response → store result → increment index → if more items, re-execute same step

**Approach**: After storing the child's result, if more items remain, call `continueExecution` again WITHOUT advancing `CurrentStep`. The fan-out action will pick up the incremented index from its state and spawn the next child.

```go
// In fan-out response handling
if fanOutState.CurrentIndex < fanOutState.TotalItems {
    // More items - re-run this step (don't advance CurrentStep)
    state.Status = StatusExecutingStep
    return s.continueExecution(ctx, state, execCtx)
}
// All done - normal flow advances to next_step
```

### Q5: Correlation ID Flow

page-builder children need to find agents spawned by pageflow-builder. This works because:
1. pageflow-builder spawns agents with `correlation_id = X`
2. pageflow-builder spawns page-builder with same `correlation_id = X`
3. page-builder calls agents using `target_role`
4. Role lookup uses `correlation_id` to find the right agents

**Verify during implementation**: Ensure correlation_id is passed through spawn and preserved in child orchestration.

---

## Implementation Order

1. **Input mapping + contract validation** (Phase 1)
    - Lowest risk, highest value
    - Can test independently
    - Backward compatible with legacy input_fields

2. **Stop __raw_message__ accumulation** (Part of Phase 1)
    - May require fixing some fallback code first
    - Test carefully

3. **sequential_fan_out action** (Phase 2)
    - New action, doesn't break existing loops
    - Can test with a simple test workflow first

4. **page-builder agent definition** (Phase 3)
    - Just a database insert
    - No code changes needed

5. **Update pageflow-builder workflow** (Phase 4)
    - The risky part - changes production behavior
    - Do last, with feature flag

---

## Files to Modify

### New Files
- `platform/orchestration/input_mapping.go` - ResolveInputMapping, ValidateInputContract, getValueAtExactPath
- `platform/actions/sequential_fan_out.go` - SequentialFanOutAction

### Modified Files (Code)
- `platform/actions/call_agent.go` - Add input_mapping support, contract validation
- `platform/actions/spawn_agent.go` - Add input_mapping support, contract validation
- `platform/orchestration/coordinator.go` - Fan-out response handling, re-execute step logic
- `platform/orchestration/data_helpers.go` - Remove __raw_message__ accumulation, add deprecation warnings
- `platform/orchestration/types.go` - Add FanOutState, InputContract, OutputContract types
- `platform/actions/actions.go` - Register sequential_fan_out action

### Database Changes
- Add `input_contract` to page-content-writer, content-reviewer, deployer-agent, image-generator
- Insert page-builder agent definition
- Update pageflow-builder workflow (replace loop with fan_out)

---

## Summary: What Changes vs What Stays

### Stays the Same
- Agent spawning mechanism (spawn_agent action)
- Agent-to-agent communication (Kafka topics)
- Role-based routing (target_role finds spawned agents)
- Individual agent workflows (page-content-writer, content-reviewer, etc.)
- Orchestration state storage (orchestration_states table)
- Awaited requests tracking (awaited_requests table)

### Changes
- **input_fields** → **input_mapping** (explicit paths, no fallback hunting)
- **No validation** → **Contract validation** (hard fail on missing fields)
- **__raw_message__ fallbacks** → **Removed** (data is where mapping says)
- **loop action with substep injection** → **sequential_fan_out with child orchestrations**
- **One orchestration for all loop iterations** → **Separate orchestration per page**
- **Step names like build_pages_loop_iter_0_write_page_content** → **Clean step names in each child**

### Log Visibility Improvement

**Before:**
```
orch=478f step=build_pages_loop_iter_0_write_page_content action=call_agent
orch=478f step=build_pages_loop_iter_0_write_page_content status=awaiting
orch=478f step=build_pages_loop_iter_0_review_page_content action=call_agent
... (all same orchestration ID, confusing step names)
```

**After:**
```
orch=478f step=build_pages action=sequential_fan_out iteration=0/5 child=abc1
orch=abc1 agent=page-builder step=write_content action=call_agent
orch=abc1 agent=page-builder step=write_content status=awaiting
orch=abc1 agent=page-builder step=review_content action=call_agent
orch=abc1 agent=page-builder step=complete msg="workflow complete"
orch=478f step=build_pages msg="iteration 0 complete"
orch=478f step=build_pages action=sequential_fan_out iteration=1/5 child=def2
orch=def2 agent=page-builder step=write_content action=call_agent
... (each child clearly separate, grep by orch ID to see full page build)
```

### Error Message Improvement

**Before:**
```
ERROR: site_id not found at site_record.site_id
```

**After:**
```
ERROR: step write_content in agent page-builder: 
  contract violation for 'page-content-writer': missing required field 'site_record'
  Provided: [page, reviewed_brief, style_collection]
  Check input_mapping in step config
```