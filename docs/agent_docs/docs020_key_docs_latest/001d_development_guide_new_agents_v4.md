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

**Example — dynamic dispatch loop:**
- Needed: Dispatch work items to different handler agents at runtime
- Assumed: `spawn_agent` only supports static `agent_type` — need custom topic resolution
- Reality: `spawn_agent` already has `agent_type_field` (dynamic resolution from collected_data). Three hours spent designing a "universal topic fallback" that wasn't needed
- Lesson: `grep -n "agent_type_field" platform/orchestration/actions/*.go` before assuming a capability is missing

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

### Actions are the unit of work — don't split them into wrapper + core

All action logic lives inside the action function. Private helpers in the same file are fine (`insertWorkItem`, `upsertSite`, `lookupPageID`), but don't create exported "core logic" functions that actions wrap.

**The wrong pattern:**

```go
// WriteSiteSpec is the "core logic" — exported so other Go code can call it
func WriteSiteSpec(ctx context.Context, db *sql.DB, p WriteSiteSpecParams, logger *zap.Logger) (map[string]interface{}, error) {
    // all the real work here
}

// WriteSiteSpecAction is a "thin wrapper" around the core function
func WriteSiteSpecAction(ctx context.Context, params ActionParams) (interface{}, error) {
    // extract inputs, then call WriteSiteSpec()
}
```

**Why this happens:** You anticipate that another action (say `seed_build_queue`) will need to call `WriteSiteSpec()` directly from Go, so you pre-extract the logic into a callable function. This feels like good engineering — separation of concerns, testability, DRY.

**Why it's wrong:**

- No other action in the codebase does this. You're inventing a two-tier architecture for a caller that doesn't exist yet.
- Composition happens through workflows, not Go-calling-Go. If `seed_build_queue` needs to write a spec, its workflow includes a `write_site_spec` step. That's how the system sequences work.
- The exported function creates a second API surface with its own parameter struct (`WriteSiteSpecParams`) that duplicates what `ActionInputSpec` already defines. Now two contracts describe the same action.
- When you eventually change the action, you have to update both the exported function signature and the action wrapper.

**The right pattern:**

```go
func WriteSiteSpecAction(ctx context.Context, params ActionParams) (interface{}, error) {
    // extract inputs
    // do the work (queries, transactions, logic)
    // return result
}
```

If a private helper would reduce repetition within the same file, keep it unexported or use helpers.go or datahelpers:

```go
func siteSpecDeepMerge(dst, src map[string]interface{}) map[string]interface{} { ... }
```

**If you later genuinely need the logic callable from Go:** Extract a private helper at that point, same as `insertWorkItem` is private to `write_build_items_action.go`. Don't pre-build the abstraction.

**Test:** Does your action file export anything besides the `XxxAction` function and the `XxxInputSpec` variable? If yes, you're probably over-abstracting.

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
{ "action": "call_agent", "config": { "target_role": "webdesigner" } }
```

### How call_agent finds the spawned agent

`call_agent` has two lookup strategies. Understanding which one to use avoids a common trap.

**`target_role` → `findAgentByRole` (preferred)**

Scans ALL keys in collected_data. Matches on the `role` field inside spawn results. This is how every existing workflow does it — page-content-writer's research agent, site-work-orchestrator's planner call, etc.

```json
"spawn_research_agent": { "config": { "role": "researcher", "agent_type": "research-agent" } }
"call_researcher":      { "config": { "target_role": "researcher" } }
```

Works regardless of what the spawn step's `output_field` is named.

**`agent_type` → `findAgentByType` (has a filter trap)**

Scans only collected_data keys that start with `spawn_`. If your spawn step's `output_field` is `"webdesign_agent"` or `"asset_deployer_agent"`, findAgentByType will never find it — it's looking for keys starting with `spawn_`.

**Rule: Always use `target_role` to find spawned agents.** It's more reliable (scans everything) and decouples the call from the spawn step's output_field naming.

### Dynamic dispatch: spawn→call in loops

When you don't know the agent type at workflow-definition time (e.g. a dispatch loop processing work items with different handler types), both `spawn_agent` and `call_agent` support dynamic type resolution:

```json
"spawn_handler": {
    "action": "spawn_agent",
    "config": {
        "role": "fix_handler",
        "agent_type_field": "current_fix_item.handler_agent"
    },
    "output_field": "spawn_handler"
},
"call_handler": {
    "action": "call_agent",
    "config": {
        "target_role": "fix_handler",
        "input_mapping": {
            "site_id": "site_record.site_id",
            "domain": "site_record.domain",
            "asset_id?": "current_fix_item.spec.asset_id"
        }
    }
}
```

The pattern: **fixed role, dynamic type.** `agent_type_field` resolves the type from collected_data at runtime. `target_role` finds the agent by role at call time. Each loop iteration spawns whatever type the work item says, and the call finds it by role. Same standard spawn→call pattern, just with dynamic resolution.

This is exactly how page-content-writer spawns research-agent (fixed role `"researcher"`, static type) — the dispatch loop just adds dynamic type resolution on top of the same mechanism.

**Trap we fell into:** When first building the dispatch loop, we tried to bypass spawn entirely by constructing topic names directly (a "universal topic fallback"). The reasoning was "we can't pre-spawn every handler type." But `spawn_agent` already supports `agent_type_field` — it resolves the type dynamically from collected_data. There was no problem to solve. The standard spawn→call pattern works for dynamic dispatch without any Go changes. Always check existing capabilities before inventing new mechanisms.

### What the orchestrator passes to handlers

The orchestrator passes raw identifiers. The handler loads its own context.

```json
"input_mapping": {
    "site_id": "site_record.site_id",
    "domain": "site_record.domain",
    "asset_id?": "current_fix_item.spec.asset_id",
    "purpose?": "current_fix_item.spec.purpose"
}
```

The `?` suffix makes fields optional — silently skipped if absent. Handlers that don't need `asset_id` ignore it. Handlers that do (asset-deployer) resolve the rest themselves (e.g. looking up the s3:// URI from the asset record).

The handler doesn't know it was dispatched by a work item. It receives `site_id`, `domain`, and whatever spec fields apply — the same inputs it would get from a direct CLI call. This keeps handlers independently callable and testable.

**Test:** Can you call the handler agent directly from the CLI with just `site_id` and `domain`? If not, you've leaked orchestrator concerns into the handler.

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


----

## Lessons Learned — Build Pipeline Post-Mortem

Add these sections to 001b_development_guide_new_agents_v3.md.

---

### Specialist vs Handler: the persistence boundary

**Problem we hit:** The dispatch loop called `page-content-writer` directly as a handler. It generated HTML and returned it, but nothing was saved to `page_components`. The work item was marked complete with the HTML trapped in `site_work_items.result`. The rerender then assembled empty pages.

**Root cause:** `page-content-writer` is a **specialist** — it generates content and returns it to its caller. The old monolithic orchestrators (pageflow-builder, site-work-orchestrator) had post-processing steps after calling it: `assemble_page → git_commit → save_page_sections → update_page_status`. When we moved to the dispatch loop, we dropped those steps because the loop just does `call_handler → mark_complete`.

**Rule: If a specialist agent returns data but doesn't persist it, it cannot be used as a dispatch handler directly.** Wrap it in a handler agent that adds the persistence steps.

The test: **Does the agent write its own outputs to the database?**
- `webdesign-agent`: Yes — writes CSS to `site_components`. Can be a handler directly.
- `image-generator`: Yes — the dispatch has `store_asset` afterwards... actually no, the current dispatch loop doesn't do that either. This needs the same wrapper treatment.
- `page-content-writer`: No — returns `page_html` and `sections_metadata` in its response. Needs a wrapper (`page-build-handler`) that calls `save_page_sections` and `update_page_status`.

**Handler checklist for dispatch loop compatibility:**
1. Does the agent persist its outputs to the database? (page_components, site_components, assets, site_specs)
2. Does it update status on the records it touches? (pages.build_status, etc.)
3. Can you call it with just `site_id` + `domain` + `spec` and have everything land in the right tables?

If the answer to any of these is no, you need a wrapper handler agent.

---

### Work item domain is NOT the site domain

**Problem we hit:** `WriteBuildItemsAction` created `needs_design` with `domain: "design"`. The dispatch loop filters `item_domain: "build"`. CSS generation was never dispatched.

**The `domain` column on `site_work_items` is a namespace for categorising work items** (like "build", "maintenance", "marketing"). It is NOT the website domain (like "gaswholesalers.com"). These are two completely different concepts sharing the same column name.

**Rule: All items in the initial build pipeline must use `domain: "build"`.** The dispatch loop filters by this. If you create items with a different domain, they won't be picked up.

The site domain (gaswholesalers.com) comes from `input_data.domain` which is passed to the dispatch loop from the trigger. It's forwarded to handlers via `input_mapping`. The work item's `domain` field is never passed to handlers as the site domain.

**Naming trap:** The dispatch loop's `input_mapping` had `"domain": "pending.first_item.domain"` which would pass "build" to the handler instead of "gaswholesalers.com". The handler needs the site domain. Fixed to `"domain": "input_data.domain"`.

---

### Every pipeline must end with assembly and deployment

**Problem we hit:** The planner created content page items and design items, but no rerender or deploy step. When all items completed, the site had sections in the database but nothing was assembled or committed to git.

**Rule: The planner (or WriteBuildItemsAction) must create terminal work items that assemble and deploy.** The minimum chain for a brochure site:

```
needs_domain_research (1) → needs_briefing (2) → needs_site_plan (3)
  → needs_logo (5) → needs_hero_image (5) → needs_design (8)
  → needs_content_page × N (10-17) → needs_rerender (20)
```

`needs_rerender` is the terminal item. It:
- Optionally re-renders site_components (header/footer/head with CSS)
- Assembles each page from page_components + site_components
- Git commits each page
- Triggers deployment via GitHub Actions

Without it, the pipeline produces data but no website.

---

### Pre-flight checklist for new work item types

Before adding a new `item_type` to the pipeline:

1. **Domain:** Is it `"build"`? If not, will the dispatch loop's `item_domain` filter find it?
2. **Handler agent:** Is the handler self-contained? Does it persist its outputs? (See specialist vs handler above)
3. **Priority:** Does it run after its dependencies? Lower number = runs first.
4. **depends_on:** If it must wait for specific items, are the UUIDs populated?
5. **Spec:** Does the spec contain everything the handler needs, or does the handler load context itself?
6. **Terminal item:** If this is the last step, does something trigger assembly/deployment?
7. **Input mapping:** Does the dispatch loop's `input_mapping` pass what the handler expects? Check that `domain` means site domain, not work item namespace.

---

### Testing a new pipeline end-to-end

After any change to the build pipeline, verify with this query sequence:

```sql
-- 1. All items created with correct domain and handler
SELECT item_type, domain, handler_agent, priority, status
FROM site_work_items WHERE site_id = '...' ORDER BY priority;

-- 2. After completion: page_components populated
SELECT p.name, COUNT(pc.id) as components, SUM(LENGTH(pc.rendered_html)) as html_bytes
FROM page_components pc JOIN pages p ON pc.page_id = p.id
WHERE p.site_id = '...' GROUP BY p.name;

-- 3. Site components (CSS/header/footer) exist
SELECT slot_name, LENGTH(rendered_html) as len FROM site_components WHERE site_id = '...';

-- 4. Pages have correct status
SELECT name, build_status FROM pages WHERE site_id = '...';

-- 5. Git commits happened (check git-adapter logs)
kubectl -n ai-persona-system logs -l app=git-adapter --tail=50 | grep DOMAIN
```

If query 2 returns 0 rows, the handler didn't persist. If query 3 is empty, CSS wasn't generated. If query 4 shows "deployed" but query 2 is empty, the status was updated without content.

---

Optional fields in dispatch loop input_mapping must use the ? suffix
The dispatch loop serves every work item type through a single call_handler step. Different item types have different spec shapes — needs_rerender has spec.refresh_site_components, content pages have spec.name and spec.sections, and needs_design has an empty spec: {}. If the call_handler input_mapping references a path like "refresh_site_components": "pending.first_item.spec.refresh_site_components", the Go input_mapping resolver (resolveInputMapping in coordinator.go) will hard-fail the entire call when that path doesn't exist — even if the handler being called doesn't need that field. The fix is the ? suffix on the destination key: "refresh_site_components?": "pending.first_item.spec.refresh_site_components". This tells the resolver to skip silently when the source path is missing. Rule: any field in the dispatch loop's input_mapping that comes from spec.* or any other item-type-specific path must use the ? suffix. Only site_id, domain, and work_item_id — fields guaranteed to exist on every work item — should be non-optional.


