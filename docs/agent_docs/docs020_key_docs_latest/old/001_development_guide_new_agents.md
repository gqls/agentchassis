# 001 — Development Guide

Practical daily reference for building, debugging, and maintaining agents. Read this before writing any new code.

---

## Pre-Flight: Does This Already Exist? (STEP ZERO)

This is the most important step. Skip it and you will waste hours building something that already exists under a different name.

**Real example:** We built `asset-deploy-agent` with a new `load_undeployed_assets` action, a new agent definition, and a new registry entry. The existing `asset-deployer` agent already did the same thing. It was listed in `isStorageEnabledAgent`, referenced in `deploy_image_asset_action.go` comments, and had a full agent definition in the database. Three hours wasted.

### 0a. Search agent_definitions for similar agents

```sql
SELECT type, display_name, status, agent_category,
       input_contract->'required' as requires,
       substring(description, 1, 120) as desc_snippet
FROM agent_definitions
WHERE deleted_at IS NULL
  AND (
    type ILIKE '%<keyword>%'
    OR display_name ILIKE '%<keyword>%'
    OR description ILIKE '%<keyword>%'
  )
ORDER BY type;
```

Search for every noun in your proposed agent name. If building "asset-deploy-agent", search: `asset`, `deploy`, `image`, `deployer`.

### 0b. Search the action registry

```bash
grep -n "Handler:" platform/orchestration/actions/registry.go | grep -i "<keyword>"
grep -rn "RegisterActionInputSpec" platform/orchestration/actions/ | grep -i "<keyword>"
```

### 0c. Search Go code for similar functions

```bash
grep -rn "func.*<Keyword>" platform/orchestration/actions/*.go
grep -rn "type.*<Keyword>" platform/orchestration/actions/*.go
```

### 0d. Check gate functions (storage, LLM access lists)

```bash
grep -A 20 "isStorageEnabledAgent\|isLLMEnabledAgent" platform/orchestration/actions/*.go
```

If your proposed agent name appears here under a different spelling, it was anticipated and may already exist.

### 0e. Search workflows that use similar actions

```sql
SELECT type FROM agent_definitions
WHERE default_config::text ILIKE '%<action_name>%';
```

### 0f. Document what you found

Write a brief note before proceeding:

```
Existing agents checked: [search terms used]
  - asset-deployer: deploys single image via deploy_image_asset. Status: experimental.
  - image-generator: generates but doesn't deploy.

Existing actions checked:
  - deploy_image_asset: downloads S3, optimizes, commits to git
  - store_generated_image: stores to S3 only

Decision: Reuse asset-deployer. No new agent needed.
```

### 0g. Check input_contract compatibility

If an existing agent is close, verify it accepts the data you'd send:

```sql
SELECT type, input_contract, output_contract
FROM agent_definitions WHERE type = '<existing_agent>';
```

If the contract needs a small tweak (e.g. making `site_id` optional when `domain` suffices), that's a patch — not a new agent.

**Rule: If you cannot demonstrate that no existing agent or action covers the need (by showing the search results), do not create a new one.**

---

## API Verification Reference

Before using any type, method, or field in Go code, verify it exists. 10 seconds of grep saves hours of debugging.

### ActionInputs methods (complete list)

```go
.Get(key) string
.GetMap(key) map[string]interface{}
.GetInt(key, default) int
.GetBool(key, default) bool
.Has(key) bool
.GetRaw(key) interface{}
```

There is **no** `GetString`, `GetUUID`, `GetFloat`, or `GetSlice`.

### ExecutionContext identity fields

```go
params.ExecutionContext.Sender.AgentType     // ✓ who sent this
params.ExecutionContext.Sender.AgentID       // ✓
params.ExecutionContext.Sender.PodName       // ✓
params.ExecutionContext.ToAgentType          // ✓ who it's addressed to
params.ExecutionContext.FromAgentType        // ✓
```

There is **no** `params.ExecutionContext.AgentType` directly.

### Database constraint values

```sql
-- agent_category allowed values:
SELECT conname, pg_get_constraintdef(oid)
FROM pg_constraint WHERE conname = 'check_ad_category';
-- Result: strategist, executor, analyst, integrator, coordinator, specialist
```

### Rule: grep before using

```bash
grep "func.*ActionInputs\|type ActionInputs" platform/orchestration/datahelpers/*.go
grep "type ExecutionContext struct" -A 60 platform/orchestration/types/*.go
```

---

## Reuse Before Creating

Before creating new code, check for existing actions/functions that can be **patched** with small enhancements.

**Example — git_commit:**
- Existing: Uses `page_field` → forces `.html` extension
- Needed: Deploy CSS files
- Solution: 4-line patch adding `file_path` config override
- Avoided: New `git_commit_file` action with duplicated logic

**Example — WebscrapeAction:**
- Existing: Hardcoded `input_data.target_url`
- Needed: Flexible URL from `input_data.url`
- Solution: ~25-line patch adding `url_field` config
- Avoided: New `webscrape_request` action duplicating adapter calls

**Example — asset deploy:**
- Existing: `asset-deployer` agent with full workflow, `deploy_image_asset` action
- Needed: Deploy undeployed assets found by discovery checks
- Solution: Point discovery work items at existing `asset-deployer` (handler_agent field)
- Avoided: New `asset-deploy-agent`, `load_undeployed_assets` action, new registry entry — three files that duplicated what already existed under a slightly different name

**Check existing code for:**
- Similar actions that could be extended
- Config options that could be added
- Shared helper functions

---

## Core Design Principles

### Agents own their domain

Each specialist agent is self-contained and independently callable. The agent handles its own data gathering rather than relying on callers.

```
Builder workflow:
  call_agent: { "site_id": "uuid" }    ← Simple

Agent workflow:
  load_my_data → analyze → generate → deploy    ← Agent handles complexity
```

### Callers pass raw data, agents derive what they need

Parent orchestrators pass raw domain identifiers. The child agent decides how to use them.

**Good:** `{ "district_code": "BT4", "area_name": "Belfast", "business_type": "vet" }`
**Bad:** `{ "query": "veterinary practice BT4 Belfast UK" }` (leaks child's search strategy)

**Test:** If you changed how the child agent works internally, would any caller need updating? If yes, you've leaked responsibility upward.

### Workflows simple, complexity in Go

Workflow JSON should be declarative and readable. Complex logic belongs in Go actions.

| Belongs in workflow config | Belongs in Go |
|---|---|
| `query_template`: how to compose a search | Looping through results and filtering |
| `input_fields`: which fields an action needs | Conditional branching based on data |
| `num_results`: 10 | DB queries and data transformation |
| `include_pages`: true | Error handling and retry logic |

Templates and config declarations are fine in workflow JSON — they express intent, not logic.

### Don't create subworkflows in SQL

Spawn sub-agents with their own workflows instead. This keeps logs clear, maintenance easier, and responsibilities separate.

### Agents respond to caller's topic

When spawned as part of a workflow, agents respond to their parent's responses topic, not their own.

### Spawn before call

In builder workflows, agents must be spawned before they can be called:

```json
{ "action": "spawn_agent", "config": { "role": "webdesigner", "agent_type": "webdesign-agent" } }
```

Then later:

```json
{ "action": "call_agent", "config": { "agent_type": "webdesign-agent", "target_role": "webdesigner" } }
```

---

## Checklist for New Specialist Agent

After completing Step 0 (pre-flight), proceed through these steps:

### 1. Define the domain

What single responsibility does this agent own?

### 2. Design the load action

What data does the agent need? Create `load_*_for_<purpose>` action — but first check 0b/0c to see if one already exists.

### 3. Define ActionInputSpec

```go
var MyAgentInputSpec = datahelpers.ActionInputSpec{
    Required: []string{"site_id"},
    Optional: []string{"include_pages", "max_items"},
    Defaults: map[string]interface{}{
        "max_items": 10,
    },
    Deprecated: datahelpers.BuildDeprecationMap([]string{"site_id", "domain"}),
}

func init() {
    datahelpers.RegisterActionInputSpec("my_action", MyAgentInputSpec)
}
```

**Warning:** Check optional field names against the "Field Name Collisions" section below. Names like `content_data`, `status`, `domain` will be found via nested lookup in `site_record.*` even if the caller never sent them.

### 4. Use ExtractActionInputs in action

```go
inputs, err := datahelpers.ExtractActionInputs(
    params.CollectedData,
    params.StepConfig.Config,
    MyAgentInputSpec,
    params.Logger,
)
if err != nil {
    return nil, err
}
```

### 5. Design the workflow

Keep it simple: `load → analyze → generate → deploy → complete`

### 6. Ensure standalone mode

Agent works with just an ID/domain.

### 7. Plan integration

How will builders spawn and call this agent?

### 8. Register action

Add to `action_registry.go`.

### 9. Create agent definition

SQL insert with workflow, contracts, tags.

### 10. Test both modes

Standalone call and integrated in a builder workflow.

---

## Standardized Input Extraction

### The three layers

| Layer | Where | Purpose | Pattern |
|---|---|---|---|
| **input_mapping** | `call_agent` config | Caller maps data to child's input_data | `"site_id": "site_record.site_id"` |
| **input_fields** | Action config | Action declares needed fields | `"input_fields": ["site_id", "domain"]` |
| **ActionInputSpec** | Go action code | Standardized extraction with validation | `ExtractActionInputs(...)` |

### input_mapping (caller's responsibility)

```json
{
  "action": "call_agent",
  "config": {
    "agent_type": "page-rerender",
    "target_role": "page_renderer",
    "input_mapping": {
      "page_id": "current_page.page_id",
      "site_id": "rerender_pages.site_id",
      "domain": "rerender_pages.domain"
    }
  }
}
```

### Optional fields in input_mapping

Use the `?` suffix when a destination field may not exist in the source:

```json
{
  "input_mapping": {
    "domain": "input_data.domain",
    "page_name?": "input_data.page_name",
    "field_updates?": "input_data.field_updates"
  }
}
```

`ResolveInputMapping` checks `strings.HasSuffix(destField, "?")`. If the source path doesn't exist, the field is silently skipped. The `?` is stripped before storing.

**Use when:** Agent supports multiple modes (content_edit sends field_updates, component_swap sends new_component_function). **Don't use for:** Required fields that must always be present — let the mapping fail loudly.

### Field name collisions

`ExtractActionInputs` has a backward-compat nested lookup that checks these parent objects:

```go
nestedSources := []struct{ parent, child string }{
    {"current_page", field},
    {"rerender_pages", field},
    {"site_record", field},
    {"input_data", field},
}
```

**Any optional field name in your ActionInputSpec will also match `site_record.<your_field>`, `input_data.<your_field>`, etc.** If those parent objects contain a key with the same name, ExtractActionInputs silently picks it up.

**Real example:** The section-editor's spec had `content_data` as optional. `ensure_site_record` puts the site plan into `collected_data["site_record"]["content_data"]`. ExtractActionInputs found `site_record.content_data` via nested lookup and treated it as the caller's replacement data — overwriting the hero section with the site plan.

**Rules:**

1. Never name an optional field the same as a common column/key in `sites`, `pages`, or `site_record`. Watch out for: `content_data`, `status`, `domain`, `name`, `title`, `description`, `config`, `metadata`.

2. If your field could collide, prefix it: `replacement_content_data` not `content_data`, `item_domain` not `domain`.

3. Check collected_data at runtime — if `ensure_site_record` runs before your action, `site_record.*` is in scope.

4. When in doubt, use explicit `input_fields` in your step config and verify the extraction path by checking logs.

### Config value patterns

Values in workflow config are either **paths** (resolved from collected_data) or **literals** (used directly). ExtractActionInputs treats everything as a path. For literal values, read directly from `params.StepConfig.Config`:

| Config key | Value | Type | Read via |
|---|---|---|---|
| `site_id` | `"site_record.site_id"` | path | ExtractActionInputs |
| `work_item_id` | `"current_item.id"` | path | ExtractActionInputs |
| `item_domain` | `"build"` | literal | direct config read |
| `handler_agent` | `"page-content-writer"` | literal | direct config read |
| `max_items` | `20` | literal | direct config read |
| `error_message` | `"Content review not approved"` | literal | direct config read |

### Dedicated load actions

Each specialist agent gets a dedicated `load_*` action:

| Agent | Load Action | Returns |
|---|---|---|
| webdesign-agent | `load_site_for_design` | site info, pages, components, colors, typography |
| link-manager | `load_site_for_links` | pages, navigation structure, internal links |
| seo-agent | `load_site_for_seo` | pages, meta tags, content for analysis |
| section-editor | `load_edit_context` | target page_component, template, page info |

---

## Agent Message Structure

All agents receive work via Kafka messages. A message has three layers:

### Kafka headers

| Header | Purpose | Example |
|---|---|---|
| `correlation_id` | Links all messages in one job | UUID |
| `orchestration_id` | Identifies the orchestration_states row | UUID |
| `request_id` | Unique ID for this specific message | UUID |
| `message_type` | `request` or `response` | `request` |
| `action` | What the agent should do | `process` or `orchestrate` |
| `responses_topic` | Where to send the reply | `system.agent.generic.responses` |
| `sender_agent_type` | Who sent this | `cli` |

The `responses_topic` header is important — agents always reply to the **caller's** responses topic, not their own.

### Message body structure

```json
{
  "headers": {
    "correlation_id": "...",
    "orchestration_id": "...",
    "message_type": "request",
    "action": "process",
    "sender": {
      "agent_id": "cli-user",
      "agent_type": "cli",
      "pod_name": "cli"
    }
  },
  "config": {
    "workflow": { ... }
  },
  "input_data": { ... }
}
```

Three top-level keys: `headers` (routing), `config` (workflow definition), `input_data` (domain payload).

### The spawn+call pattern

Most CLI triggers use the generic agent as a thin launcher:

```
CLI message → system.agent.generic.requests
  → generic agent runs inline workflow:
      1. spawn_agent (creates specialist)
      2. call_agent (forwards input_data)
      3. complete (returns result)
        → specialist runs its own full workflow
```

The inline workflow goes in `config.workflow`. The `input_mapping` in `call_agent` controls which fields get forwarded.

### Topic naming

```
system.agent.{agent-type}.requests    — where agents receive work
system.agent.{agent-type}.responses   — where agents send results
```

### Generating IDs from bash

```bash
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
```

### HITL responses

When a workflow hits a human-in-the-loop step, it creates an `awaited_requests` row and pauses. The HITL response goes to the agent's **responses** topic with `message_type: "response"`, `in_response_to_request_id: <from awaited_requests>`, `status: "complete"`, `sender_agent_type: "human"`.

---

## Orchestration State

Each orchestration creates a row in `orchestration_states`:
- `workflow_plan` — the workflow definition
- `collected_data` — accumulates step outputs as the workflow progresses
- `current_step` — which step is executing
- `status` — RUNNING, AWAITING_RESPONSES, COMPLETED, FAILED

The `collected_data` is how steps communicate. Step A stores output at `output_field`, Step B reads it via `input_fields` or `ExtractFields`.

---

## Stale Orchestration Sweeper

Timeout handling uses in-process goroutines. These die when pods restart, leaving orchestrations stuck in AWAITING_RESPONSES forever. This is the #1 cause of pipeline stalls.

### Three failure modes

1. **Response lost** — child completed but Kafka message never received
2. **Timeout goroutine lost** — pod that started the goroutine restarted
3. **Child stuck** — child itself is in AWAITING_RESPONSES (cascading stall)

### Approach

A periodic DB sweep running on existing agent-chassis pods. No new service. Uses `FOR UPDATE SKIP LOCKED` so multiple pods can run the sweep safely.

Runs every 60 seconds. For each expired request (`timeout_at < NOW() - 30 seconds`), classifies:

**A. Child COMPLETED — response was lost:** Synthesize a completion response from the child's `final_result`, produce to parent's responses_topic. Most common case.

**B. Child FAILED:** Forward the failure to the parent. Parent's error handling takes over.

**C. No child found, or child still running:** Increment retry_version. If < 3: re-send request. If >= 3: mark as expired, fail the parent orchestration.

### Edge cases handled

- Race with normal response delivery: parent's dedup logic handles duplicate responses
- Parent already completed/failed: check status before synthesizing, mark awaited_request as processed
- Cascading stalls: processes deepest (oldest timeout_at) first
- Job topics expired (1hr retention): directly update parent's orchestration_state to advance past the awaited step

---

## Diagnostic Queries

### Check what a workflow step expects vs produces

```sql
SELECT key as step_name,
       value->>'action' as action,
       value->>'output_field' as output_field,
       value->>'next_step' as next_step
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE type = '<agent_type>';
```

### Check call_agent input mappings

```sql
SELECT key as step_name,
       value->'config'->>'agent_type' as target_agent,
       value->'config'->'input_mapping' as mapping
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE type = '<agent_type>'
  AND value->>'action' = 'call_agent';
```

### Find agents that use a specific action

```sql
SELECT type FROM agent_definitions
WHERE default_config::text ILIKE '%<action_name>%'
  AND deleted_at IS NULL;
```

### Check for stale orchestrations

```sql
SELECT os.orchestration_id, os.owner_agent_type, os.status,
       os.current_step, os.updated_at,
       NOW() - os.updated_at as stuck_for
FROM orchestration_states os
WHERE os.status = 'AWAITING_RESPONSES'
  AND os.updated_at < NOW() - INTERVAL '1 hour'
ORDER BY os.updated_at;
```

### Check awaited requests

```sql
SELECT ar.request_id, ar.step_name, ar.target_agent_type,
       ar.timeout_at, ar.retry_version, ar.status
FROM awaited_requests ar
WHERE ar.status = 'waiting'
  AND ar.timeout_at < NOW()
ORDER BY ar.timeout_at;
```

### Kubernetes quick reference

```bash
kubectl -n ai-persona-system get pods
kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep "$CORRELATION_ID"
kubectl -n kafka get pods
```

---

## Migration Guide: Converting Existing Actions

### Step 1: Identify current patterns
Look for `*_field` config keys and manual extraction code.

### Step 2: Create ActionInputSpec
Map old patterns to new field names in `Deprecated` map.

### Step 3: Replace extraction code
Replace 40+ lines of boilerplate with single `ExtractActionInputs()` call.

### Step 4: Test both patterns
Ensure old workflows still work (with deprecation warnings).

### Step 5: Update workflow configs
Migrate from `*_field` to `input_fields` pattern.

### Step 6: Remove deprecated support
After all workflows migrated, remove deprecated patterns from spec.