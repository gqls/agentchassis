The user wants me to create a comprehensive document that:

Documents the decision to eliminate agent_group_definitions
Traces through the design decisions we made
Provides the code/SQL needed to implement everything

Let me organize this into a clear document covering:

The Problem - spawn_group spawns agents but no one orchestrates them
The Insight - every agent is an orchestrator, so groups should just BE agents
The Decision - eliminate agent_group_definitions, use agent_definitions
Entity State - how to persist data across orchestrations
The Log-Based Design - append, don't overwrite
Implementation - SQL and Go code

# Agent Architecture Simplification & Entity State

## Overview

This document captures the design decisions and implementation for:
1. Eliminating `agent_group_definitions` in favour of unified agent model
2. Implementing persistent entity state with log-based storage
3. Simplifying `spawn_group` to spawn an orchestrator agent

---

## Part 1: The Problem

### Current Architecture Issue

When `spawn_group` is called:
1. It spawns all worker agents (strategist, architect, writer, etc.)
2. Workers initialize and wait for messages
3. The group's `orchestration_workflow` is returned as data but **never executed**
4. No one sends `call_agent` messages to the workers
5. Workers sit idle forever

```
intake-orchestrator
  └── spawn_builder_group (spawns 5 agents) ✓
      └── landing-page-builder agents initialized ✓
          └── But nobody orchestrates them! ✗
```

### Root Cause

`agent_group_definitions` has an `orchestration_workflow` but no agent to run it.
The workflow describes WHAT should happen but there's no WHO to execute it.

---

## Part 2: The Insight

### "Every Agent Is An Orchestrator"

If every agent can orchestrate (spawn other agents, call them, coordinate work),
then a "group" is just an agent whose workflow happens to spawn and coordinate
other agents.

### What's the Actual Difference?

| agent_group_definitions | agent_definitions |
|------------------------|-------------------|
| `group_type` | `type` |
| `name` | `display_name` |
| `orchestration_workflow` | `default_config.workflow` |
| `agent_configs` | Implicit in workflow spawn steps |
| `briefing_questionnaire` | Can add this field |

They're essentially the same thing.

---

## Part 3: The Decision

### Eliminate agent_group_definitions

A "group" becomes an agent type. For example, `landing-page-builder` is an agent
whose workflow spawns workers and calls them.

**Before:**
```
agent_group_definitions (template)
  → spawn_group 
    → spawns workers
    → ??? who orchestrates ???
```

**After:**
```
agent_definitions (template)
  → spawn_agent
    → creates orchestrator pod
    → orchestrator runs its workflow
    → workflow spawns workers and calls them
```

### spawn_group Simplification

`spawn_group` becomes a thin wrapper around `spawn_agent`:

```go
func SpawnGroupAction(ctx context.Context, params ActionParams) (interface{}, error) {
    // 1. Resolve group_type (e.g., "landing-page-builder")
    groupType, err := resolveGroupType(params)
    if err != nil {
        return nil, err
    }
    
    // 2. spawn_group is now just spawn_agent with group_type as agent_type
    spawnParams := params
    spawnParams.StepConfig = models.Step{
        Action: "spawn_agent",
        Config: map[string]interface{}{
            "agent_type": groupType,
            "role":       "orchestrator",
        },
    }
    
    // 3. The spawned agent runs its workflow which spawns workers
    return SpawnAgentAction(ctx, spawnParams)
}
```

---

## Part 4: Migration

### Step 1: Add briefing_questionnaire to agent_definitions

```sql
ALTER TABLE agent_definitions 
ADD COLUMN IF NOT EXISTS briefing_questionnaire JSONB DEFAULT '{}'::jsonb;

COMMENT ON COLUMN agent_definitions.briefing_questionnaire IS 
'Optional questionnaire for briefing agents to execute when working with this agent type';
```

### Step 2: Create Agent Definitions for Each Group Type

For each row in `agent_group_definitions`, create a corresponding `agent_definitions` row.

```sql
-- Landing Page Builder Agent
INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config,
    briefing_questionnaire
)
VALUES (
    gen_random_uuid(),
    'landing-page-builder',
    'Landing Page Builder',
    'Orchestrates the complete landing page build workflow - spawns specialist agents and coordinates them to build conversion-focused landing pages',
    'orchestrator',
    '{
      "workflow": {
        "start_step": "spawn_strategist",
        "steps": {
          "spawn_strategist": {
            "action": "spawn_agent",
            "config": {"agent_type": "site-strategist", "role": "strategist"},
            "next_step": "spawn_architect",
            "description": "Spawn strategist"
          },
          "spawn_architect": {
            "action": "spawn_agent",
            "config": {"agent_type": "landing-page-architect", "role": "architect"},
            "next_step": "spawn_writer",
            "description": "Spawn landing page architect"
          },
          "spawn_writer": {
            "action": "spawn_agent",
            "config": {"agent_type": "content-writer", "role": "writer"},
            "next_step": "spawn_assembler",
            "description": "Spawn content writer"
          },
          "spawn_assembler": {
            "action": "spawn_agent",
            "config": {"agent_type": "html-assembler", "role": "assembler"},
            "next_step": "spawn_deployer",
            "description": "Spawn HTML assembler"
          },
          "spawn_deployer": {
            "action": "spawn_agent",
            "config": {"agent_type": "site-deployer", "role": "deployer"},
            "next_step": "call_strategist",
            "description": "Spawn deployer"
          },
          "call_strategist": {
            "action": "call_agent",
            "config": {
              "agent_type": "site-strategist",
              "target_role": "strategist",
              "input_fields": ["input_data", "brief_data"],
              "timeout_seconds": 120
            },
            "output_field": "build_plan",
            "next_step": "call_architect",
            "description": "Generate build plan from brief"
          },
          "call_architect": {
            "action": "call_agent",
            "config": {
              "agent_type": "landing-page-architect",
              "target_role": "architect",
              "input_fields": ["build_plan", "brief_data", "input_data"],
              "timeout_seconds": 120
            },
            "output_field": "template_data",
            "next_step": "call_writer",
            "description": "Assemble page template from components"
          },
          "call_writer": {
            "action": "call_agent",
            "config": {
              "agent_type": "content-writer",
              "target_role": "writer",
              "input_fields": ["template_data", "build_plan", "brief_data", "input_data"],
              "timeout_seconds": 300
            },
            "output_field": "content_data",
            "next_step": "call_assembler",
            "description": "Generate content for template placeholders"
          },
          "call_assembler": {
            "action": "call_agent",
            "config": {
              "agent_type": "html-assembler",
              "target_role": "assembler",
              "input_fields": ["content_data", "template_data", "brief_data", "input_data"],
              "timeout_seconds": 120
            },
            "output_field": "final_html",
            "next_step": "call_deployer",
            "description": "Assemble final HTML with CSS/JS"
          },
          "call_deployer": {
            "action": "call_agent",
            "config": {
              "agent_type": "site-deployer",
              "target_role": "deployer",
              "input_fields": ["final_html", "input_data"],
              "timeout_seconds": 180
            },
            "output_field": "deployment_result",
            "next_step": "complete",
            "description": "Deploy to git repository"
          },
          "complete": {
            "action": "complete_workflow",
            "description": "Landing page build complete"
          }
        }
      },
      "processing_mode": "orchestration",
      "timeout_seconds": 900
    }'::jsonb,
    true,
    '["orchestration", "site-building", "landing-page"]'::jsonb,
    'docker.io/aqls/agent-chassis',
    'v1.0.478',
    '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
    '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
    '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
    '{
      "sections": [
        {
          "name": "brand",
          "title": "Brand & Identity",
          "questions": [
            {"field": "brand_name", "type": "text", "label": "Brand/Company Name", "required": true},
            {"field": "tagline", "type": "text", "label": "Tagline or Slogan", "required": false},
            {"field": "tone", "type": "select", "label": "Brand Tone", "options": ["professional", "friendly", "bold", "playful", "technical"], "default": "professional"}
          ]
        },
        {
          "name": "value_proposition",
          "title": "Value Proposition",
          "questions": [
            {"field": "primary_benefit", "type": "textarea", "label": "What is the main benefit for visitors?", "required": true},
            {"field": "unique_selling_points", "type": "textarea", "label": "What makes you different? (List 3-5 points)", "required": true},
            {"field": "target_audience", "type": "text", "label": "Who is your ideal customer?", "required": true}
          ]
        },
        {
          "name": "conversion",
          "title": "Conversion Goals",
          "questions": [
            {"field": "primary_cta", "type": "text", "label": "Primary Call-to-Action (e.g., Sign Up, Buy Now)", "required": true},
            {"field": "primary_cta_url", "type": "text", "label": "Primary CTA Link/Action", "required": false},
            {"field": "secondary_cta", "type": "text", "label": "Secondary CTA (e.g., Learn More)", "required": false}
          ]
        },
        {
          "name": "social_proof",
          "title": "Trust & Social Proof",
          "questions": [
            {"field": "has_testimonials", "type": "boolean", "label": "Do you have customer testimonials?", "default": false},
            {"field": "client_count", "type": "text", "label": "Number of customers/users (if applicable)", "required": false},
            {"field": "notable_clients", "type": "text", "label": "Notable clients or partners", "required": false}
          ]
        }
      ]
    }'::jsonb
)
ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    briefing_questionnaire = EXCLUDED.briefing_questionnaire,
    description = EXCLUDED.description,
    updated_at = now();


-- Content Site Builder Agent
INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config,
    briefing_questionnaire
)
VALUES (
    gen_random_uuid(),
    'content-site-builder',
    'Content Site Builder',
    'Orchestrates the complete content/publishing site build workflow',
    'orchestrator',
    '{
      "workflow": {
        "start_step": "spawn_strategist",
        "steps": {
          "spawn_strategist": {
            "action": "spawn_agent",
            "config": {"agent_type": "site-strategist", "role": "strategist"},
            "next_step": "spawn_architect"
          },
          "spawn_architect": {
            "action": "spawn_agent",
            "config": {"agent_type": "content-site-architect", "role": "architect"},
            "next_step": "spawn_writer"
          },
          "spawn_writer": {
            "action": "spawn_agent",
            "config": {"agent_type": "content-writer", "role": "writer"},
            "next_step": "spawn_assembler"
          },
          "spawn_assembler": {
            "action": "spawn_agent",
            "config": {"agent_type": "html-assembler", "role": "assembler"},
            "next_step": "spawn_deployer"
          },
          "spawn_deployer": {
            "action": "spawn_agent",
            "config": {"agent_type": "site-deployer", "role": "deployer"},
            "next_step": "call_strategist"
          },
          "call_strategist": {
            "action": "call_agent",
            "config": {
              "agent_type": "site-strategist",
              "target_role": "strategist",
              "input_fields": ["input_data", "brief_data"],
              "timeout_seconds": 120
            },
            "output_field": "build_plan",
            "next_step": "call_architect"
          },
          "call_architect": {
            "action": "call_agent",
            "config": {
              "agent_type": "content-site-architect",
              "target_role": "architect",
              "input_fields": ["build_plan", "brief_data", "input_data"],
              "timeout_seconds": 120
            },
            "output_field": "template_data",
            "next_step": "call_writer"
          },
          "call_writer": {
            "action": "call_agent",
            "config": {
              "agent_type": "content-writer",
              "target_role": "writer",
              "input_fields": ["template_data", "build_plan", "brief_data", "input_data"],
              "timeout_seconds": 300
            },
            "output_field": "content_data",
            "next_step": "call_assembler"
          },
          "call_assembler": {
            "action": "call_agent",
            "config": {
              "agent_type": "html-assembler",
              "target_role": "assembler",
              "input_fields": ["content_data", "template_data", "brief_data", "input_data"],
              "timeout_seconds": 120
            },
            "output_field": "final_html",
            "next_step": "call_deployer"
          },
          "call_deployer": {
            "action": "call_agent",
            "config": {
              "agent_type": "site-deployer",
              "target_role": "deployer",
              "input_fields": ["final_html", "input_data"],
              "timeout_seconds": 180
            },
            "output_field": "deployment_result",
            "next_step": "complete"
          },
          "complete": {
            "action": "complete_workflow",
            "description": "Content site build complete"
          }
        }
      },
      "processing_mode": "orchestration",
      "timeout_seconds": 900
    }'::jsonb,
    true,
    '["orchestration", "site-building", "content-site"]'::jsonb,
    'docker.io/aqls/agent-chassis',
    'v1.0.478',
    '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
    '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
    '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
    '{
      "sections": [
        {
          "name": "publication",
          "title": "Publication Identity",
          "questions": [
            {"field": "publication_name", "type": "text", "label": "Publication/Site Name", "required": true},
            {"field": "tagline", "type": "text", "label": "Tagline", "required": false},
            {"field": "editorial_tone", "type": "select", "label": "Editorial Tone", "options": ["news_formal", "magazine_polished", "blog_casual", "technical"], "default": "magazine_polished"}
          ]
        },
        {
          "name": "content_structure",
          "title": "Content Structure",
          "questions": [
            {"field": "categories", "type": "textarea", "label": "Content Categories (one per line)", "required": true},
            {"field": "publishing_frequency", "type": "select", "label": "Publishing Frequency", "options": ["daily", "weekly", "occasional"], "default": "weekly"}
          ]
        },
        {
          "name": "monetization",
          "title": "Monetization",
          "questions": [
            {"field": "monetization_model", "type": "select", "label": "Revenue Model", "options": ["advertising", "subscription", "affiliate", "none"], "default": "advertising"},
            {"field": "newsletter_signup", "type": "boolean", "label": "Include Newsletter Signup?", "default": true}
          ]
        }
      ]
    }'::jsonb
)
ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    briefing_questionnaire = EXCLUDED.briefing_questionnaire,
    updated_at = now();
```

### Step 3: Update intake-orchestrator Workflow

Change `spawn_group` to `spawn_agent`:

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,spawn_builder}',
    '{
      "action": "spawn_agent",
      "config": {
        "agent_type_field": "confirmed_type.recommended_group",
        "role": "builder"
      },
      "output_field": "spawned_builder",
      "next_step": "complete",
      "description": "Spawn the appropriate builder agent"
    }'::jsonb
)
WHERE type = 'intake-orchestrator';
```

Or update the intake-orchestrator to use the new pattern.

### Step 4: Simplify SpawnGroupAction

```go
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
        value := resolveFieldPath(fieldPath, collectedData)
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
```

### Step 5: Deprecate agent_group_definitions (Later)

Once all workflows are migrated:

```sql
-- Rename to indicate deprecation
ALTER TABLE agent_group_definitions RENAME TO agent_group_definitions_deprecated;

-- Or drop if confident
-- DROP TABLE agent_group_definitions;
```

---

## Part 5: Entity State - Persistent Data Across Orchestrations

### The Need

Currently, orchestration state is tied to ONE workflow run. When workflow completes,
data is frozen. A new workflow for the same domain starts fresh.

We want data that persists across workflows:
- Research findings accumulate
- Brand guidelines persist
- Build history is tracked
- Agent learnings are retained

### Design Principles

1. **Log-based storage** - append, don't overwrite
2. **Namespaced** - agents can have their own data space
3. **Message-based stays primary** - DB is available when needed, not mandatory
4. **Different accumulation patterns** - additive, evolutionary, versioned

### Data Accumulation Patterns

| Pattern | Example | Behaviour |
|---------|---------|-----------|
| Additive | Research, product ideas | Keep adding, rarely supersede |
| Evolutionary | Brand tone, strategy | Latest usually best, history shows journey |
| Versioned | Deployments, builds | Each is a valid snapshot |
| Singleton | Domain name, objective | Rarely changes |

### Schema

```sql
-- Entity state log - append-only storage
CREATE TABLE entity_state_log (
    id BIGSERIAL PRIMARY KEY,
    
    -- Entity identification
    entity_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100),           -- 'domain', 'project', 'customer', etc.
    namespace VARCHAR(100),             -- NULL=shared, or agent_type, or custom
    
    -- Data
    path VARCHAR(255) NOT NULL,         -- 'brand.tone', 'research.products', etc.
    data JSONB NOT NULL,
    
    -- Context
    created_at TIMESTAMP DEFAULT now(),
    created_by_agent_type VARCHAR(100),
    orchestration_id UUID,
    correlation_id VARCHAR(100),
    
    -- For future intelligent supersession
    superseded_by BIGINT REFERENCES entity_state_log(id),
    supersession_reason TEXT
);

-- Index for efficient lookups
CREATE INDEX idx_entity_state_lookup 
ON entity_state_log(entity_id, namespace, path, created_at DESC);

-- Index for finding active (non-superseded) entries
CREATE INDEX idx_entity_state_active 
ON entity_state_log(entity_id, namespace, path) 
WHERE superseded_by IS NULL;

-- Index by entity type for bulk operations
CREATE INDEX idx_entity_state_type 
ON entity_state_log(entity_type, entity_id);

-- Comments
COMMENT ON TABLE entity_state_log IS 'Append-only log of entity state changes, supporting persistent data across orchestrations';
COMMENT ON COLUMN entity_state_log.namespace IS 'NULL for shared data, agent_type for agent-specific data, or custom namespace';
COMMENT ON COLUMN entity_state_log.path IS 'Dot-notation path within namespace, e.g., brand.tone, research.products';
COMMENT ON COLUMN entity_state_log.superseded_by IS 'Points to newer entry that supersedes this one (for compaction)';
```

### Actions Implementation

```go
// FILE: platform/orchestration/actions/entity_state_actions.go

package actions

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "os"
    "strings"
    "time"

    "github.com/google/uuid"
    "go.uber.org/zap"
)

// ============================================================================
// APPEND ENTITY STATE
// ============================================================================

// AppendEntityStateAction appends data to the entity state log
// Config:
//   entity_id_field: path to entity ID in collected_data (e.g., "input_data.domain")
//   entity_type: type of entity (e.g., "domain", "project")
//   namespace: NULL for shared, or specific namespace (defaults to agent_type if "auto")
//   path: dot-notation path (e.g., "brand.tone", "research.products")
//   data_field: path to data in collected_data to store
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
    err = params.DB.QueryRow(ctx, `
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
//   entity_id_field: path to entity ID
//   entity_type: optional filter
//   namespace: NULL for shared, "auto" for agent_type, or specific
//   paths: array of path patterns to read (supports wildcards with %)
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
        rows, err := params.DB.Query(ctx, `
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
//   entity_id_field: path to entity ID
//   namespace: NULL for shared, "auto" for agent_type, or specific
//   path: specific path to get history for
//   limit: max entries to return (default 50)
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
    rows, err := params.DB.Query(ctx, `
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
//   entity_id_field: path to entity ID
//   paths: array of path patterns (default: ["%"] for all)
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
//   entity_id_field: path to entity ID
//   path: path within agent's namespace
//   data_field: data to store
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

    value := resolveFieldPath(fieldPath, collectedData)
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

    value := resolveFieldPath(fieldPath, collectedData)
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

func nullableString(s string) interface{} {
    if s == "" {
        return nil
    }
    return s
}

// resolveFieldPath navigates a dot-notation path through nested maps
func resolveFieldPath(path string, data map[string]interface{}) interface{} {
    parts := strings.Split(path, ".")
    var current interface{} = data

    for _, part := range parts {
        if m, ok := current.(map[string]interface{}); ok {
            current = m[part]
        } else {
            return nil
        }
    }

    return current
}
```

### Register Actions

```go
// FILE: platform/orchestration/actions/entity_state_actions.go (add init function)

func init() {
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
}
```

---

## Part 6: Example Usage

### Research Agent Workflow

An agent that builds on previous research:

```json
{
  "workflow": {
    "start_step": "load_previous_research",
    "steps": {
      "load_previous_research": {
        "action": "read_entity_history",
        "config": {
          "entity_id_field": "input_data.domain",
          "namespace": "shared",
          "path": "research.products",
          "limit": 20
        },
        "output_field": "previous_research",
        "next_step": "do_research"
      },
      "do_research": {
        "action": "execute_llm_prompt",
        "config": {
          "prompt_template": "Research new products for {{.input_data.domain}}.\n\nAlready known products:\n{{range .previous_research.history}}• {{.data}}\n{{end}}\n\nFind NEW products not in this list.",
          "output_field": "new_products"
        },
        "next_step": "save_findings"
      },
      "save_findings": {
        "action": "append_entity_state",
        "config": {
          "entity_id_field": "input_data.domain",
          "entity_type": "domain",
          "namespace": "shared",
          "path": "research.products",
          "data_field": "new_products"
        },
        "next_step": "complete"
      },
      "complete": {
        "action": "complete_workflow"
      }
    }
  }
}
```

### Site Strategist Using Its Own State

```json
{
  "workflow": {
    "start_step": "check_my_previous_work",
    "steps": {
      "check_my_previous_work": {
        "action": "read_my_state",
        "config": {
          "entity_id_field": "input_data.domain",
          "paths": ["build_plans", "model_preferences"]
        },
        "output_field": "my_history",
        "next_step": "create_plan"
      },
      "create_plan": {
        "action": "execute_llm_prompt",
        "config": {
          "prompt_template": "Create a build plan for {{.input_data.domain}}.\n\n{{if .my_history.state}}Previous approaches:\n{{.my_history.state}}\n\nBuild on what worked, avoid what didn't.{{end}}"
        },
        "output_field": "build_plan",
        "next_step": "save_plan"
      },
      "save_plan": {
        "action": "write_my_state",
        "config": {
          "entity_id_field": "input_data.domain",
          "entity_type": "domain",
          "path": "build_plans",
          "data_field": "build_plan"
        },
        "next_step": "complete"
      },
      "complete": {
        "action": "complete_workflow"
      }
    }
  }
}
```

---

## Part 7: Summary

### Changes Made

1. **Unified agent model** - Groups become agents with orchestration workflows
2. **Simplified spawn_group** - Now delegates to spawn_agent
3. **Log-based entity state** - Persistent data across orchestrations
4. **Agent-namespaced storage** - Each agent type can have its own data space
5. **Additive by default** - Append, don't overwrite

### Files to Create/Modify

| File | Action |
|------|--------|
| `platform/orchestration/actions/spawn_group.go` | Simplify to delegate to spawn_agent |
| `platform/orchestration/actions/entity_state_actions.go` | New file with state actions |
| `migrations/XXX_entity_state_log.sql` | Create entity_state_log table |
| `migrations/XXX_builder_agents.sql` | Create builder agent definitions |

### Future Enhancements

1. **Intelligent compaction** - LLM-based consolidation of historical entries
2. **Cross-agent queries** - Read other agents' namespaces when needed
3. **Entity relationships** - Link related entities (domain → builds → deployments)
4. **Retention policies** - Auto-archive old entries