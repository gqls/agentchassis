# Implementation Plan: Fixing Data Path Issues and Standardizing Deployment

## Overview

This plan addresses four interconnected problems:

| # | Problem | Root Cause | Solution |
|---|---------|------------|----------|
| 1 | Git deploy fails "no files to commit" | deployer-agent expects data at different path than pageflow-builder sends | Remove redundant deploy step |
| 2 | Template tags not rendered (`{{.headline}}` in output) | Content from LLM not reaching render context | Fix path references in workflow |
| 3 | Unpredictable data paths | Each call_agent adds nesting; paths depend on workflow depth | Use input/output contracts consistently |
| 4 | Inconsistent deploy patterns | Different builders use different mechanisms | Standardize to per-page commits |

---

## Phase 0: Validation Setup (Step 0.1)

### Goal: Establish tooling to prevent future path mismatches

#### Step 0.1.1: Get the Workflow Validator

Copy the workflow_validator tool to your local repo:
- `main.go` - the validator
- `go.mod` - module definition
- `README.md` - usage instructions
- `sample_agents.json` - example for testing

#### Step 0.1.2: Export your agent definitions

Run this SQL to export the agents we'll be modifying:

```sql
-- Save this output to agents_to_validate.json
SELECT json_agg(json_build_object(
  'id', id::text,
  'type', type,
  'display_name', display_name,
  'default_config', default_config,
  'input_contract', input_contract,
  'output_contract', output_contract
)) 
FROM agent_definitions 
WHERE type IN (
  'pageflow-builder', 
  'page-content-writer', 
  'deployer-agent',
  'content-reviewer',
  'site-planner'
);
```

#### Step 0.1.3: Run validation on current state

```bash
cd workflow_validator
go run main.go -agents agents_to_validate.json -verbose
```

**Document the output** - this is our baseline of known issues.

---

## Phase 1: Fix Git Deploy (Remove Redundant Step)

### Background

Currently pageflow-builder does:
1. Loop: write → review → assemble → **git_commit** (per page) ← THIS WORKS
2. After loop: trigger_site_deploy → calls deployer-agent ← THIS FAILS

The per-page git_commit in the loop already commits each page. The second deploy is redundant and fails due to path mismatch.

---

### Step 1.1: Query current workflow state

**SQL to run:**
```sql
SELECT 
  jsonb_pretty(config->'workflow'->'steps'->'build_pages_loop'->'next_step') as loop_next_step,
  jsonb_pretty(config->'workflow'->'steps'->'spawn_reviewer'->'next_step') as reviewer_next_step,
  config->'workflow'->'steps' ? 'trigger_site_deploy' as has_trigger_step,
  config->'workflow'->'steps' ? 'spawn_deployer' as has_spawn_deployer
FROM agent_definitions 
WHERE agent_type = 'pageflow-builder';
```

**Expected output:**
- `loop_next_step`: `"trigger_site_deploy"`
- `reviewer_next_step`: `"spawn_deployer"`
- `has_trigger_step`: true
- `has_spawn_deployer`: true

**Test:** Just run the query, record results, no changes yet.

---

### Step 1.2: Change loop next_step to skip trigger_site_deploy

**SQL file: `001_fix_pageflow_skip_deploy.sql`**

```sql
-- Step 1.2: Make build_pages_loop go directly to update_site_status
-- This skips the failing deployer-agent call

BEGIN;

UPDATE agent_definitions
SET config = jsonb_set(
    config,
    '{workflow,steps,build_pages_loop,next_step}',
    '"update_site_status"'
),
updated_at = NOW()
WHERE agent_type = 'pageflow-builder';

-- Verify
SELECT 
    config->'workflow'->'steps'->'build_pages_loop'->>'next_step' as new_next_step
FROM agent_definitions 
WHERE agent_type = 'pageflow-builder';

COMMIT;
```

**Test:**
1. Apply the SQL
2. Verify output shows `new_next_step` = `update_site_status`
3. Trigger a build (small test site)
4. Check logs - should NOT see "deployer-agent" being called after the loop
5. Should complete without "no files to commit" error

**Rollback if needed:**
```sql
UPDATE agent_definitions
SET config = jsonb_set(
    config,
    '{workflow,steps,build_pages_loop,next_step}',
    '"trigger_site_deploy"'
)
WHERE agent_type = 'pageflow-builder';
```

---

### Step 1.3: Update spawn_reviewer to skip spawn_deployer

**SQL file: `002_fix_pageflow_skip_spawn_deployer.sql`**

```sql
-- Step 1.3: Make spawn_reviewer go directly to ensure_site_record
-- This removes the spawning of an agent we no longer use

BEGIN;

UPDATE agent_definitions
SET config = jsonb_set(
    config,
    '{workflow,steps,spawn_reviewer,next_step}',
    '"ensure_site_record"'
),
updated_at = NOW()
WHERE agent_type = 'pageflow-builder';

-- Verify
SELECT 
    config->'workflow'->'steps'->'spawn_reviewer'->>'next_step' as new_next_step
FROM agent_definitions 
WHERE agent_type = 'pageflow-builder';

COMMIT;
```

**Test:**
1. Apply the SQL
2. Verify output shows `new_next_step` = `ensure_site_record`
3. Trigger a build
4. Check logs - should NOT see "spawn_deployer" step executing

---

### Step 1.4: Remove unused steps and update output_fields

**SQL file: `003_fix_pageflow_cleanup.sql`**

```sql
-- Step 1.4: Remove trigger_site_deploy and spawn_deployer steps
-- Also update complete to not reference deployment_result

BEGIN;

-- Remove the unused steps
UPDATE agent_definitions
SET config = config #- '{workflow,steps,trigger_site_deploy}' 
                    #- '{workflow,steps,spawn_deployer}',
updated_at = NOW()
WHERE agent_type = 'pageflow-builder';

-- Update complete step output_fields
UPDATE agent_definitions
SET config = jsonb_set(
    config,
    '{workflow,steps,complete,config,output_fields}',
    '["site_record", "pages_built"]'
)
WHERE agent_type = 'pageflow-builder';

-- Verify
SELECT 
    config->'workflow'->'steps' ? 'trigger_site_deploy' as has_trigger,
    config->'workflow'->'steps' ? 'spawn_deployer' as has_spawn_deployer,
    config->'workflow'->'steps'->'complete'->'config'->'output_fields' as output_fields
FROM agent_definitions 
WHERE agent_type = 'pageflow-builder';

COMMIT;
```

**Test:**
1. Apply the SQL
2. Verify: `has_trigger` = false, `has_spawn_deployer` = false
3. Verify: `output_fields` = `["site_record", "pages_built"]`
4. Trigger a build - should complete successfully

---

## Phase 2: Diagnose Template Rendering

### Step 2.1: Add diagnostic logging to RenderComponentAction

**Goal:** See exactly what data is/isn't available when rendering

Find `RenderComponentAction` in your actions code (likely `component_actions.go` or `v3_site_actions.go`).

Add this diagnostic block right after `config := params.StepConfig.Config`:

```go
// DIAGNOSTIC LOGGING - Remove after debugging
params.Logger.Info("RenderComponentAction: DIAGNOSTIC",
    zap.Strings("collected_data_keys", getTopLevelKeys(params.CollectedData)),
)

// Check content_from path
if contentFrom, ok := config["content_from"].(string); ok {
    contentData := datahelpers.ExtractNestedField(params.CollectedData, contentFrom)
    params.Logger.Info("RenderComponentAction: content_from diagnostic",
        zap.String("path", contentFrom),
        zap.Bool("found", contentData != nil),
        zap.String("type", fmt.Sprintf("%T", contentData)),
    )
    if contentData != nil {
        if m, ok := contentData.(map[string]interface{}); ok {
            params.Logger.Info("RenderComponentAction: content_from keys",
                zap.Strings("keys", getMapKeysFromInterface(m)),
            )
        }
    }
}

// Helper if not exists
func getTopLevelKeys(m map[string]interface{}) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}
```

**Test:**
1. Deploy the updated code
2. Trigger a build
3. Search logs for `RenderComponentAction: DIAGNOSTIC`
4. Note what fields ARE available vs what content_from expects

---

### Step 2.2: Check page-content-writer workflow paths

**SQL to examine:**
```sql
SELECT 
    jsonb_pretty(
        config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'->'generate_content'->'output_field'
    ) as generate_output,
    jsonb_pretty(
        config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'->'render_section'->'config'->'content_from'
    ) as render_content_from
FROM agent_definitions 
WHERE agent_type = 'page-content-writer';
```

**Expected finding:**
- `generate_output` should be the base field name (e.g., `"generated_content"`)
- `render_content_from` might have `.result` suffix that doesn't match

---

### Step 2.3: Fix path mismatch (if found)

Based on diagnostics, if render_section expects `generated_content.result` but data is at `generated_content`:

**SQL file: `004_fix_content_path.sql`**

```sql
BEGIN;

UPDATE agent_definitions
SET config = jsonb_set(
    config,
    '{workflow,steps,process_sections_loop,config,substeps,render_section,config,content_from}',
    '"generated_content"'  -- Remove .result if that's the issue
),
updated_at = NOW()
WHERE agent_type = 'page-content-writer';

-- Verify
SELECT 
    config->'workflow'->'steps'->'process_sections_loop'->'config'->'substeps'->'render_section'->'config'->>'content_from'
FROM agent_definitions 
WHERE agent_type = 'page-content-writer';

COMMIT;
```

**Test:**
1. Apply if diagnostics confirmed this is the issue
2. Trigger a build
3. Check if HTML now has actual content instead of `{{.headline}}`

---

## Phase 3: Update Input/Output Contracts

### Step 3.1: Document current contract usage

**SQL to audit:**
```sql
SELECT 
    type,
    input_contract IS NOT NULL as has_input,
    output_contract IS NOT NULL as has_output
FROM agent_definitions
WHERE type IN (
    'pageflow-builder',
    'page-content-writer', 
    'deployer-agent',
    'content-reviewer',
    'site-planner'
)
ORDER BY type;
```

### Step 3.2: Update pageflow-builder contracts to match reality

**SQL file: `005_update_contracts.sql`**

```sql
BEGIN;

-- Update pageflow-builder input contract
UPDATE agent_definitions
SET input_contract = '{
    "expects": {
        "input_data.domain": "string - the domain name",
        "input_data.objective": "string - what the site should achieve",
        "reviewed_brief": "object - completed questionnaire answers with company_name, services, about_us, etc"
    },
    "required": ["input_data.domain", "reviewed_brief"]
}'::jsonb
WHERE agent_type = 'pageflow-builder';

-- Update pageflow-builder output contract (no longer includes deployment_result)
UPDATE agent_definitions
SET output_contract = '{
    "produces": {
        "site_record": "object with site_id, domain, status",
        "pages_built": "object with count, results array, iterations"
    }
}'::jsonb
WHERE agent_type = 'pageflow-builder';

-- Update page-content-writer contracts
UPDATE agent_definitions
SET input_contract = '{
    "expects": {
        "current_page": "object with id, name, title, sections array",
        "site_record": "object with site_id, domain",
        "reviewed_brief": "object with company_name, services, about_us, target_audience, tone",
        "style_collection": "object with colors, typography, component refs (optional)"
    },
    "required": ["current_page", "site_record", "reviewed_brief"]
}'::jsonb,
output_contract = '{
    "produces": {
        "page_html": "string - rendered HTML for the page",
        "sections": "array of rendered section objects",
        "research_ids": "array of UUIDs referencing research_results (if research was done)"
    }
}'::jsonb
WHERE agent_type = 'page-content-writer';

COMMIT;
```

---

## Phase 4: Validate Changes

### Step 4.1: Re-run workflow validator

```bash
# Export updated agents
psql -d clients_db -c "SELECT json_agg(...) FROM agent_definitions WHERE ..." > updated_agents.json

# Run validation
go run main.go -agents updated_agents.json -verbose
```

**Expected:** Fewer errors/warnings than baseline

### Step 4.2: End-to-end test

1. Create a new test site with a simple domain
2. Monitor logs throughout the flow
3. Verify:
    - No "no files to commit" error
    - HTML contains actual content, not template tags
    - Site status updates to "deployed"

---

## Summary: Files to Apply in Order

| Order | File | Purpose | Reversible |
|-------|------|---------|------------|
| 0.1 | workflow_validator/* | Validation tool | N/A |
| 1.2 | 001_fix_pageflow_skip_deploy.sql | Skip deployer-agent | Yes |
| 1.3 | 002_fix_pageflow_skip_spawn_deployer.sql | Don't spawn deployer | Yes |
| 1.4 | 003_fix_pageflow_cleanup.sql | Remove unused steps | Backup first |
| 2.3 | 004_fix_content_path.sql | Fix content_from path | Yes |
| 3.2 | 005_update_contracts.sql | Update contracts | Yes |

Run validation and test after each step before proceeding.

---

## Appendix: Diagnostic Queries

### Check what a workflow step expects vs produces

```sql
-- See full step config
SELECT jsonb_pretty(config->'workflow'->'steps'->'STEP_NAME')
FROM agent_definitions
WHERE agent_type = 'AGENT_TYPE';

-- See all steps and their output_fields
SELECT 
    key as step_name,
    value->>'action' as action,
    value->>'output_field' as output_field,
    value->>'next_step' as next_step
FROM agent_definitions,
     jsonb_each(config->'workflow'->'steps') 
WHERE agent_type = 'pageflow-builder';
```

### Check call_agent input_fields

```sql
-- See what fields are passed to child agents
SELECT 
    key as step_name,
    value->'config'->>'agent_type' as target_agent,
    value->'config'->'input_fields' as input_fields
FROM agent_definitions,
     jsonb_each(config->'workflow'->'steps') 
WHERE agent_type = 'pageflow-builder'
  AND value->>'action' = 'call_agent';
```

