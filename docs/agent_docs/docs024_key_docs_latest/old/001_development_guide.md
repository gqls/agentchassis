# 001 — Development Guide

Practical daily reference for building, debugging, and maintaining agents. Read this before writing any new code.

Consolidated from: 001i_v9, 004_unblocking_items, 005_extended_thinking, 001_config_driven, 014_loop_mechanisms, 016_workflow_data_path_validation, 031_guidelines_model_update, 037_unresolved.

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

**Example — RAG field path resolution:**
- Needed: Resolve dot-path strings like `"input_data.industry"` in RAG actions
- Built: `resolveRAGFieldPath` — a new function doing the same thing as 17 existing ones
- Reality: `datahelpers.ExtractNestedFieldString` already does exactly this, with `.response` auto-unwrapping as a bonus
- Lesson: grep datahelpers before writing any extraction function (see "Field Path Resolution" section below)

**Check existing code for:**
- Similar actions that could be extended
- Config options that could be added
- Shared helper functions

---

## Field Path Resolution: Use the Canonical Functions

The codebase has accumulated 18+ functions that resolve dot-separated field paths (`"foo.bar.baz"`) from nested maps. This happened because each was written at a different time with slightly different behaviour. **Do not add another one.**

### Canonical functions (all in `datahelpers` package)

| Function | Use case | Returns |
|---|---|---|
| `ExtractNestedField(data, path)` | General-purpose path resolution | `interface{}` |
| `ExtractNestedFieldString(data, path)` | When you need a string | `string` (empty if not found) |
| `ExtractNestedFieldMap(data, path)` | When you need a map | `map[string]interface{}` (nil if not found) |
| `GetFieldFromPath(data, path, logger)` | When you want an error on missing | `interface{}, error` |
| `GetFieldFromPathWithDefault(data, path, default, logger)` | With fallback | `interface{}` |

`ExtractNestedField` is the primary one — it handles `.response` auto-unwrapping which most of the others try to replicate.

### Functions that exist but should NOT be duplicated

These are in the actions package and are essentially duplicates with minor variations. They still work but new code should not use them:

```
resolveFieldPath            — in entity_state_actions.go and workflow_actions.go
resolveFieldPathForSpawn    — in spawn_actions.go (args swapped!)
resolveFieldPathCallAgent   — in call_agent.go (returns string only)
resolveFieldPathQuestionnaire — in fetch_agent_questionnaire.go
resolveFieldValue           — in conditional_branch_action.go
extractFieldValue           — in multipage_actions.go
```

### Common helpers to reuse (not recreate)

| Need | Use | Package |
|---|---|---|
| Nullable string for SQL | `NullableString(s)` | datahelpers |
| Nullable int for SQL | `NullableInt(n)` | datahelpers |
| Truncate with "..." | `TruncateString(s, maxLen)` | datahelpers |
| String from map with default | `GetStringField(m, key, default)` | datahelpers |
| Int from map with default | `GetIntField(m, key, default)` | datahelpers |
| Bool from map with default | `GetBoolField(m, key, default)` | datahelpers |
| Null-if-empty (actions pkg) | `nullIfEmpty(s)` | actions/helpers.go |

**Before adding any utility function**, grep the datahelpers package. If the function exists under a different name, use the existing one.

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

### Every pod-running agent needs a parent that spawned it

This is the pattern that governs *who gets their own pod*, and it's the rule we violated when we first wrote `site-adoption-agent`.

**Baseline rule:** in this system, agents become dedicated Kubernetes Job pods *only* when they're reached via `spawn_agent` → `call_agent` from a parent workflow. That's what `SpawnAgentAction` does — it calls `spawnAgentKubernetesJobFromDefinition`, creates a Job, waits for startup, and sends the initialization message. Without a `spawn_agent` step upstream, there is no pod creation.

**The outlier mistake:** you can *also* start a workflow by sending a message to `system.agent.generic.requests` with `config.agent_type: "my-agent-type"`. That's the system's entry point for callers who don't have a spawn relationship with the target — manual triggers today, and in future probably also UI actions, user-proxy agents, cross-tree invocations, and anything else that doesn't sit inside an existing orchestration. The scheduler already uses this path for heartbeats. What all these callers share is not "user-ness" but that they have no `job.X.requests` topic to post to, so they land on the shared entry point.

What's misleading is what happens next: the chassis extracts `config.agent_type`, loads the agent definition, and runs the workflow — but **inside one of the long-lived `agent-chassis` pods** that consumes that topic. The workflow runs, yes. It never gets its own pod. Its logs interleave with every other workflow those pods are currently handling. And if you want to tail it cleanly, you can't.

That's fine for genuinely short coordination — a wrapper that spawns a child and waits for the response is a few seconds of in-chassis work. It's wrong for any agent that does real work for more than that: adoption, build orchestration, long crawls, anything where you'd want a clean `kubectl logs <pod>` stream.

**The fix is the wrapper-orchestrator pattern.** Create a tiny parent agent whose entire workflow is:

```
start → spawn_<child> → call_<child> → complete
```

The parent runs briefly in the chassis pod (spawn + call + forward the response). The child is spawned as a Job with its own pod, runs the real work there, and terminates. Clean logs, one correlation per pod lifetime, aligned with every other spawn→call flow in the system.

The canonical minimal wrapper is `med-export-orchestrator`:

```json
{
  "workflow": {
    "start_step": "spawn_exporter",
    "steps": {
      "spawn_exporter": {
        "action": "spawn_agent",
        "config": { "role": "json_exporter", "agent_type": "med-json-exporter" },
        "next_step": "call_exporter"
      },
      "call_exporter": {
        "action": "call_agent",
        "config": {
          "agent_type": "med-json-exporter",
          "target_role": "json_exporter",
          "input_mapping": { "input_data": "input_data" },
          "timeout_seconds": 120
        },
        "next_step": "complete"
      },
      "complete": { "action": "complete_workflow" }
    }
  }
}
```

Three steps. `input_mapping` just passes `input_data` through. The child agent is untouched. This is also how `med-url-map-orchestrator`, `med-price-scrape-orchestrator`, and `site-adoption-orchestrator` are structured.

**Test: does this agent need a wrapper?** If anything reaches this agent via the generic entry point (manual trigger, scheduler, UI, proxy agent, cross-tree caller), and **this agent's own workflow steps** do substantive work in-chassis (LLM calls, crawls, multi-table DB writes, anything taking more than a handful of seconds), it needs a wrapper. Coordination work (conditionals, HITL review, spawning and calling children) doesn't count — that's what an orchestrator is supposed to do in-chassis, and it's what `intake-orchestrator` already does correctly. The question is whether the **substantive** work has been delegated to a spawned child.

If this agent is only ever reached via `spawn_agent` from some other workflow, no wrapper is needed — the parent already gives it a pod.

**Why this matters beyond logs.** Long-running workflows hold chassis pod capacity. A 30-minute adoption running in-chassis blocks that pod's consumer slot for 30 minutes; five concurrent adoptions can starve the chassis. Wrapped, the chassis spends a few seconds on orchestration per adoption and the work runs in its own dedicated pod elsewhere in the cluster.

### Topics: the generic entry point vs per-spawn dedicated topics

The system uses two quite different kafka patterns depending on whether a message is *starting a new workflow from outside an existing tree* or *flowing between agents inside one*. Getting these confused leads to the kinds of mistakes that produced this rule.

**How a pod decides what to subscribe to.** Every chassis-based pod reads `KAFKA_TOPIC(S)` from its environment at startup. That env var is set differently depending on how the pod came into existence:

- **Fixed-topic deployments** (adapters, the "generic" chassis, any long-lived Deployment): `KAFKA_TOPIC(S)` is baked into the Deployment manifest at deploy time. The `web-scrape-adapter` pods listen on `system.agent.webscrape-adapter.process`, the "generic" `agent-chassis` pods listen on `system.agent.generic.requests`, and so on. The `agent_definitions.topics` jsonb column is a *declaration* of where this agent expects to be reachable, but it's the Deployment manifest that actually subscribes — the jsonb is informational unless something (a Helm chart, a kustomize overlay) pulls it into a manifest.

- **Spawned Job pods:** `KAFKA_TOPIC(S)` is set per-spawn by `spawnAgentKubernetesJobFromDefinition` (see `setupAgentTopics`). It's a `job.<stable-identity>.requests` topic, created in kafka by `createTopics` just before the Job launches. The stable identity is `<correlationID[:8]>-<orchestrationID[:8]>-<agent_type>-<parent_step>`, so concurrent spawns in different orchestrations get different topics. Sequential spawns of the same type at the same step in one orchestration (e.g. a dispatch loop) reuse the same topic — the previous Job has already completed before the next is spawned, so reuse is safe. The Job pod subscribes only to its own topic. When the workflow completes the Job terminates; kafka retention cleans the topic up.

**What runs where.**

- **`system.agent.generic.requests` — the generic entry point.** Consumed by the "generic" `agent-chassis` Deployment (currently three replicas, `AGENT_TYPE=generic` baked in). When a message lands here, one of those pods reads `config.agent_type` via `MessageProcessor.extractGroupInfo`, loads the matching agent definition, and runs the workflow **in-process in that chassis pod**.

  Anything that needs to start a workflow from outside an existing parent-spawn relationship posts here. Today the main callers are manual triggers (`trigger-*.sh`) and the kafka-scheduler (heartbeats, periodic tasks). In future this may also include user-proxy agents translating chat input into dispatches, admin UI actions, or cross-tree invocations from a workflow whose parent doesn't know the target's spawn topic. What they share is not "who" they are but that they have no spawned topic to use — so they land on the common entry point.

  This mechanism is likely to evolve (a formalised entry API, typed dispatch endpoints, auth/routing in front), so don't treat `system.agent.generic.requests` as a permanent interface — treat it as "the current door". The *rule* about pod lifecycle below survives any door change.

- **`job.<stable-identity>.requests` — per-spawn dedicated topics.** Consumed by one specific spawned Job pod, for a specific workflow instance. No sharing, no content-based demux. Clean logs, clean lifecycle. This is what agent-to-agent traffic inside a workflow looks like. Always use this when you can — if you have a parent workflow, `spawn_agent` it.

- **Per-type fixed topics** (`system.agent.webscrape-adapter.process`, etc.). Consumed by long-lived adapter Deployments. Used when the "agent" is really a long-running service (a scraper, an image generator, a git adapter) that you don't want to spawn per-task.

**The rule is about pod lifecycle, not caller identity.**

The real decision is where the substantive work runs, not who triggered it:

- **Work that runs in-chassis** blocks a shared pod slot, interleaves its logs with everything else that pod is handling, and shares memory/CPU with all other in-chassis workflows. Fine for seconds, bad for minutes, harmful for tens of minutes.

- **Work that runs in a spawned Job pod** has its own pod, its own logs, its own resource limits from `agent_definitions.resources`, and terminates cleanly on completion. This is what the system was designed for and almost every agent already does this.

The fix when you find work running in-chassis is always the same: wrap it in an orchestrator whose sole job is to spawn it. The wrapper runs in-chassis briefly (the spawn + call + return); the real work moves into its own pod.

**How to decide for a new agent.**

| Agent role | Entry | Where substantive work runs |
|---|---|---|
| Spawned sub-agent (always prefer this when a parent exists) | `job.<stable-identity>.requests` (created per spawn by the parent) | Dedicated Job pod, terminates on completion |
| Orchestrator invoked via the generic entry point | `system.agent.generic.requests` (via `config.agent_type`) | The orchestrator itself runs briefly in-chassis; if it does more than trivial coordination, it should spawn a child and delegate |
| Wrapper orchestrator (canonical pattern for "I need a dedicated pod but have no parent to spawn me") | `system.agent.generic.requests` | Wrapper runs in-chassis for seconds; spawns the real worker immediately |
| Long-lived adapter / service | Fixed `system.agent.<type>.*` topic (set on the Deployment) | Persistent Deployment, not Job |

The mistake we made with `site-adoption-agent` was putting a substantive workflow in the second row and leaving it there. Adoption takes minutes, calls LLMs, crawls and parses pages, writes many rows across multiple tables. That shouldn't run in-chassis. The fix (`site-adoption-orchestrator` as the trigger target, following the third row) moves the real work into the first row.

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

## LLM Infrastructure

This section covers cross-cutting LLM concerns: model selection, call logging, the Ollama adapter, and RAG. These are **infrastructure**, not agents — they don't have agent definitions or workflows.

### Model aliases

Agent definitions use aliases like `claude-sonnet-4-6` or `claude-haiku-4-5`. These resolve to dated API strings via `model_aliases.go`. Update that file when Anthropic releases new models.

Current model strategy:

| Agent role | Model | Reasoning |
|---|---|---|
| Planning (chief-strategist, site-planner) | claude-sonnet-4-6 | High-leverage structural decisions |
| Research, reasoning, analysis | claude-sonnet-4-6 | Complex analysis, style extraction |
| Content generation (section creators, copywriter) | claude-sonnet-4-6 | Quality content matching site voice |
| Adoption (site classification, content direction) | claude-sonnet-4-6 | Reasoning about site structure and style |
| Orchestration (website-builder) | claude-haiku-4-5 | Minimal LLM use, mostly routing |
| Classification (site-classifier, future local) | claude-haiku-4-5 → ollama | Short structured output, candidate for fine-tuning |

**Rule:** `claude-sonnet-4-6` is the standard model for all LLM steps unless there is a specific reason to use something else (e.g. haiku for orchestration routing where output is trivial, or ollama for fine-tuned classification). When in doubt, use `claude-sonnet-4-6`.

**Rule:** Use short aliases (`claude-sonnet-4-6`, `claude-haiku-4-5`) not full versioned names in agent definitions. The alias map handles version resolution. Never use strings like `claude-sonnet-4-5-20250514`.

### LLM call logging

Every `execute_llm_prompt` call is logged to the `llm_call_log` table. This serves two purposes: operational visibility and training data collection for fine-tuning local models.

The logger is fire-and-forget (goroutine with 5s timeout). It captures: agent_type, step_name, model (alias + resolved), full prompt template, rendered prompt, response text, token counts, latency, success/error.

Key queries:

```sql
-- What's happening right now
SELECT agent_type, model, COUNT(*) as calls,
       ROUND(AVG(latency_ms)) as avg_ms, ROUND(AVG(output_tokens)) as avg_tokens
FROM llm_call_log WHERE created_at > NOW() - INTERVAL '1 hour'
GROUP BY agent_type, model ORDER BY calls DESC;

-- Training data readiness for a specific agent
SELECT COUNT(*) as examples FROM llm_call_log
WHERE agent_type = 'site-classifier' AND success = true;

-- Use the stats view
SELECT * FROM llm_call_stats;
```

Cleanup: call `SELECT cleanup_old_llm_logs();` from pg_cron or the maintenance-catch-all. Keeps 90 days for successful calls, 180 days for errors. At ~30KB/call, budget ~1GB/month at 1000 calls/day.

### Ollama adapter

Ollama runs as a permanent CPU adapter container in the `ai-persona-system` namespace, alongside existing adapters (web-search, scraping, git). It serves two model types:

**Embeddings (nomic-embed-text, 137M params):** ~50-100ms per call on CPU. Used by `rag_lookup` and `rag_index` actions.

**Classification (fine-tuned 7B, quantized):** ~10-30s per call on CPU. Candidates: site-classifier, vet-practice-verifier, domain-analyst. These produce short structured outputs where 30s latency is acceptable because they run once per build.

Not suitable for: content generation (long outputs from 7B on CPU = minutes), anything needing <2s response time.

Agent definitions reference Ollama like any other provider:

```json
"ai_service": {
    "provider": "ollama",
    "model": "site-classifier-v1",
    "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
}
```

The `createAIClient` switch in `ai_actions.go` routes to `aiservice.NewOllamaClient`. The Ollama client implements the same `AIService` interface as the Anthropic client, including writing `__usage_input_tokens` / `__usage_output_tokens` back into the options map for the logger.

Memory budget: ~8GB covers nomic-embed-text (~1GB) + a quantized 7B (~4GB) loaded simultaneously. No GPU needed. Models persist on a PVC so they survive pod restarts.

Concurrency note: CPU inference is sequential within a model. Two simultaneous classification requests queue. At current batch sizes this is fine; if it becomes a bottleneck, add a second replica.

### RAG actions

Two actions for retrieval-augmented generation, registered as `rag_lookup` and `rag_index` in the action registry.

**`rag_lookup`** — embed query → vector search → return top-k chunks. Falls back to trigram text search if Ollama is unavailable, so content agents don't hard-fail when the embedding service is down. Output includes `rag_context` (combined text for prompt injection) and `search_method` (vector/trigram/failed).

**`rag_index`** — chunk content → embed → store in `knowledge_base` table. Embedding failures are non-fatal: chunks get stored without embeddings and are still searchable via trigram. Dedup uses SHA256 content hash with `ON CONFLICT DO NOTHING`.

The `knowledge_base` table is a **shared resource** (unlike `agent_memory` which is per-instance). Any agent can read from it via `rag_lookup` and any agent can write to it via `rag_index`.

Embedding column is `vector(768)` for nomic-embed-text. Changing to a different embedding model requires ALTERing the column and re-embedding all rows. The `embedding_model` column tracks what was used per row.

Workflow integration example — add `rag_lookup` before a content generation step:

```json
{
    "lookup_industry_knowledge": {
        "action": "rag_lookup",
        "config": {
            "query_field": "input_data.industry",
            "collection": "industry_sites",
            "top_k": 5,
            "embedding_service": { "provider": "ollama", "model": "nomic-embed-text" }
        },
        "next_step": "generate_content",
        "output_field": "industry_knowledge"
    }
}
```

Then in the content prompt template: `{{.industry_knowledge.rag_context}}`

### Fine-tuning path

The training data pipeline: LLM call logging → export → LoRA fine-tune on GPU → GGUF export → load into Ollama → update agent definition to `provider: ollama`.

1. Accumulate 200+ successful examples for the target agent (1-2 weeks of production traffic)
2. Export with `training_data_export.sql` queries (Alpaca or ChatML format)
3. Fine-tune with LoRA on a GPU machine (unsloth framework, 15-60 min on 3090/4090)
4. Export to GGUF, create Ollama model, pull into the adapter
5. Update agent_definition: `provider: ollama`, `model: site-classifier-v1`
6. A/B test against Claude by running both and comparing outputs

Candidates for fine-tuning (short-output classification/extraction tasks):

| Agent | Output type | Why local |
|---|---|---|
| site-classifier | JSON classification | Runs every build, output is structured and short |
| vet-practice-verifier | JSON boolean + evidence | High volume, simple yes/no + extract |
| domain-analyst | JSON metadata | Structured extraction from domain info |
| briefing-agent | JSON questionnaire | Query generation from raw inputs |
| content-researcher | Search queries | Short query strings from context |

### Agent vs infrastructure: where to draw the line

Not everything needs to be an agent. The test: **does it have its own domain of responsibility, need its own workflow, and benefit from being independently spawnable and debuggable?**

| Thing | Is it an agent? | Why |
|---|---|---|
| site-classifier | Yes | Owns classification domain, has workflow, independently testable |
| LLM call logger | No | Cross-cutting instrumentation inside execute_llm_prompt |
| Ollama provider | No | Transport layer, same as Anthropic client |
| rag_lookup/rag_index | No (actions) | Building blocks, like web_search or db_query |
| knowledge-indexer | Yes (future) | When we need scrape→index→refresh orchestration |
| maintenance scheduler | Yes | Owns batch scheduling domain, has workflow |

The `rag_lookup` and `rag_index` are actions because they're single operations any workflow can use. When we need a full pipeline (discover sites to scrape → scrape → chunk → embed → store → periodically refresh), that orchestration becomes a knowledge-indexer agent. Until then, actions are sufficient — following the "Reuse Before Creating" principle.

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

### Optional fields in dispatch loop input_mapping must use the ? suffix

The dispatch loop serves every work item type through a single call_handler step. Different item types have different spec shapes — needs_rerender has spec.refresh_site_components, content pages have spec.name and spec.sections, and needs_design has an empty spec: {}. If the call_handler input_mapping references a path like "refresh_site_components": "pending.first_item.spec.refresh_site_components", the Go input_mapping resolver (resolveInputMapping in coordinator.go) will hard-fail the entire call when that path doesn't exist — even if the handler being called doesn't need that field.

The fix is the `?` suffix on the destination key: `"refresh_site_components?": "pending.first_item.spec.refresh_site_components"`. This tells the resolver to skip silently when the source path is missing.

**Rule:** any field in the dispatch loop's input_mapping that comes from `spec.*` or any other item-type-specific path must use the `?` suffix. Only `site_id`, `domain`, and `work_item_id` — fields guaranteed to exist on every work item — should be non-optional.

---

## Implementation Status: LLM Optimization

### Deployed and verified

**SQL migrations (deployed to clients_db):**
- `081_llm_model_upgrades_and_logging.sql` — model alias updates + llm_call_log table + cleanup function + stats view. Table required post-deploy schema fix: added `agent_id` column, relaxed `step_name` and `prompt_rendered` to nullable (Go code and table schema were misaligned).
- `082_rag_knowledge_base.sql` — knowledge_base table with pgvector + trigram indexes + stats view. Not yet populated.

**Go code deployed (in agent-chassis image):**
- `platform/aiservice/anthropic.go` — Usage struct captures token counts, writes `__usage_input_tokens`/`__usage_output_tokens` back to options map
- `platform/orchestration/actions/llm_call_logger.go` — fire-and-forget logging via `LogLLMCall()` with `LLMCallLogParams` struct
- `platform/orchestration/actions/ai_actions.go` — logging calls in `ExecuteLLMPromptAction` (success + failure paths), `ollama` case in `createAIClient`
- `platform/orchestration/actions/get_pages_to_build_actions.go` — added `page_name` alias in both `scanPageRowsForBuild` and `scanPageRowsForBuildPgx`
- `buildActionParams` in coordinator.go — added `AgentType: state.OwnerAgentType` (was missing, caused empty agent_type on all action params)

**Verified in production (March 2026):**
- 57+ rows in `llm_call_log` from both `ExecuteLLMPromptAction` (page-content-writer) and direct callers (ch-llm-review)
- Token counts captured on all rows (input_tokens, output_tokens, latency_ms)
- `agent_type` populated correctly after `buildActionParams` fix
- Stats view (`llm_call_stats`) working

### Not yet deployed

**Go code produced but not deployed:**
- `platform/aiservice/ollama.go` — Ollama provider implementing AIService interface
- `platform/orchestration/actions/rag_actions.go` — `rag_lookup` and `rag_index` actions
- `platform/orchestration/datahelpers/nullable_helpers.go` — NullableInt, NullableInt64
- `platform/orchestration/actions/registry.go` — add rag_lookup, rag_index to GlobalActionRegistry

**Kubernetes (not yet deployed):**
- Ollama adapter kustomize manifests (base + production overlay + Makefile targets)

**Support files:**
- `training_data_export.sql` — queries for extracting fine-tuning data
- `deployment_guide.md` — step-by-step implementation order

### Remaining deployment order
1. ~~SQL migrations (081, 082)~~ ✓ done
2. ~~Go code patches + new files → rebuild agent-chassis image~~ ✓ done (logging + anthropic token capture)
3. ~~Verify logging works~~ ✓ done — 57+ rows captured
4. Deploy Ollama adapter when ready for RAG / local models
5. Deploy RAG actions (rag_lookup, rag_index) + registry update
6. Wire RAG into workflows (add rag_lookup before content steps, rag_index after scraping)
7. Accumulate training data (200+ examples per agent type, 1-2 weeks)
8. Fine-tune first local model (site-classifier)
9. A/B test local model vs Claude

### Not yet built (future work)
- knowledge-indexer agent (orchestrates scrape→index→refresh cycle) — build when pipeline demands it
- Maintenance-catch-all integration for llm_call_log cleanup scheduling
- REINDEX CONCURRENTLY scheduling for knowledge_base ivfflat index
- A/B testing infrastructure for local vs cloud model comparison
- Field path resolution cleanup (migrate 9+ duplicates in actions package to datahelpers calls)

--

Appendix A - bugs and fixes lessons learned

# Bugs and Fixes — Session Tally

Add to 001e_development_guide as "Lessons Learned" appendix entries.

---

## 1. QueryDatabaseAction doesn't support parameterised queries

**Symptom:** Audit agents failed at `load_brief` and `load_design_context` steps with `expected 1 arguments, got 0`.

**Root cause:** `QueryDatabaseAction` executes raw SQL strings. The `$1` placeholders in the query had no mechanism to receive parameters. The existing agents used Go template interpolation (`{{.input_data.site_id}}`) embedded in the SQL string instead.

**Fix:** Added `params` config field support to `QueryDatabaseAction` — resolves dot-paths from collected_data and passes as `queryArgs...` to `QueryContext`.

**Rule:** New `query_database` usage MUST use `$1` placeholders with `"params"` array. Never embed values via `{{.field}}` template interpolation — SQL injection risk. (Added to contracts doc 003d.)

**Legacy migration needed:** `tool-suggester` and `tool-improver` still use template interpolation.

---

## 2. Missing api_key_env_var in LLM config

**Symptom:** Audit agents reached LLM steps then failed with `ai_service.api_key_env_var not configured`.

**Root cause:** Agent definitions had `"ai_service": {"model": "...", "provider": "anthropic", "max_tokens": 4000}` but were missing `"api_key_env_var": "ANTHROPIC_API_KEY"`.

**Fix:** Added the missing field to all audit agent definitions.

**Rule:** Every `execute_llm_prompt` step MUST have `api_key_env_var` in its `ai_service` config. Copy from an existing working agent (e.g. `page-content-writer`) rather than writing from scratch.

**Checklist item:** When creating agent definitions with LLM steps, verify the ai_service config has: `model`, `provider`, `max_tokens`, `api_key_env_var`.

---

## 3. Wrong column names in SQL queries (site_specs schema)

**Symptom:** Content-quality-auditor failed with `column ss.spec_type does not exist`.

**Root cause:** The `site_specs` table uses `aspect` (not `spec_type`) and `data` (not `spec_data`). The queries were written from memory without checking the schema.

**Fix:** Updated all audit agent queries to use correct column names.

**Rule:** ALWAYS check `\d table_name` before writing SQL in agent definitions. Column names in agent workflow SQL are not validated at definition time — they fail at runtime.

---

## 4. Model alias mismatch

**Symptom:** Audit agents had `"model": "claude-sonnet-4-5-20250514"` which is a Claude 4.0 version string, not a valid alias or model name.

**Fix:** Changed to `"claude-sonnet-4-6"` (alias) which resolves correctly via `model_aliases.go`.

**Rule:** Use short aliases (`claude-sonnet-4-6`, `claude-haiku-4-5`) not full versioned names in agent definitions. The alias map handles version resolution.

---

## 5. Specialist vs handler — page-content-writer persistence gap

**Symptom:** `page-content-writer` generated blog content but it was never saved to `page_components`. The work item was marked complete with HTML trapped in `site_work_items.result`.

**Root cause:** `page-content-writer` is a specialist — returns data to caller. The dispatch loop calls handlers, which must persist their own outputs. Already documented but hit again.

**Fix:** Created `page-build-handler` wrapper agent. Updated `empty_sections` check to route to `page-build-handler` instead of `page-content-writer`.

**Rule:** Before routing a work item to a handler, check: does the agent persist its outputs to the database? If not, use a wrapper handler. (Already in dev guide — reinforce.)

---

## 6. sync_pages_to_db expects specific data shapes

**Symptom:** `blog-content-planner` tried to use `sync_pages_to_db` to create blog post pages. Failed with `no pages found in page_plan` because the LLM output was at `site_plan.result.pages` not where the action looks.

**Root cause:** `sync_pages_to_db` hardcodes looking for pages at specific paths (`site_plan.pages`, `page_plan.pages`). The config field `pages_field` is not supported. The action was built for the planner pipeline and assumes that data shape.

**Fix:** Created `create_blog_posts` Go action that handles the LLM output parsing flexibly (tries multiple paths, handles string/map/array).

**Rule:** When reusing existing actions in new workflows, verify the action actually reads the config fields you're setting. Check the Go source, not just the workflow config. Many actions have hardcoded paths that aren't configurable.

**Broader lesson:** Keep workflows simple, put complexity in Go actions. Trying to wire `sync_pages_to_db` + `write_build_items` into the blog planner created a fragile chain. A single purpose-built action is more reliable.

---

## 7. Go template rendering fails silently on nil values

**Symptom:** `execute_llm_prompt` returned empty string. The blog planner's LLM prompt used `{{range .existing_posts}}` where `existing_posts` was nil (query returned no rows).

**Root cause:** Go's `text/template` engine silently produces empty output when template execution fails. No error is returned — the prompt renders as empty and the LLM gets nothing to respond to.

**Fix:** The `create_blog_posts` action handles nil gracefully. For templates: always guard range with `{{if .field}}{{range .field}}...{{end}}{{end}}` AND ensure the field is `[]` not `nil` when empty.

**Rule:** When `query_database` returns no rows with `output_format: "array"`, the result is `null` not `[]`. Template `{{range}}` on `null` fails silently. Always wrap in `{{if}}` checks. Test templates with empty/nil inputs.

---

## 8. Pages table has `url` not `slug`

**Symptom:** `blog-content-planner` query failed with `column "slug" does not exist`.

**Root cause:** Wrote SQL referencing `slug` column which doesn't exist. The column is `url`.

**Fix:** Changed query to use `url`.

**Rule:** Same as #3 — always check schema before writing SQL. `\d pages`.

---

## 9. Agent definition not created before triggering

**Symptom:** `blog-content-planner` trigger produced no orchestration — went straight to complete with empty output.

**Root cause:** The SQL to create the agent definition hadn't been run yet. The generic chassis received the message, looked up `blog-content-planner` in `agent_definitions`, found nothing, and fell through to an empty workflow.

**Fix:** Run the SQL first, then trigger.

**Rule:** After creating an agent definition SQL file, run it before testing. The chassis doesn't log "agent definition not found" — it silently executes an empty workflow.

**Improvement needed:** The chassis should log a warning when it can't find an agent definition for the requested type.

---

## 10. site_specs data not available to audit agents

**Symptom:** Content-quality-auditor reported "no target audience defined" for sites that had the data in `content_data.response`.

**Root cause:** Older sites (built by `pageflow-builder`) have planning data in `sites.content_data` but no `site_specs` rows. Audit agents query `site_specs` and find nothing.

**Fix:** Backfilled `site_specs` for existing sites. Added `read_site_spec` fallback to read from `content_data`. Added `write_site_spec` steps to both `pageflow-builder` and `site-work-orchestrator` workflows.

**Rule:** When adding agents that read site configuration, use `read_site_spec` (which handles fallback) not direct `query_database` on `site_specs`. And when writing planning data, write to both `content_data` (working state) and `site_specs` (versioned record).


## 11. Invalid Anthropic API version header

**Symptom:** API calls failed with version-related errors.

**Root cause:** Used `2025-01-01` which is not a valid Anthropic API version.

**Fix:** Changed to `2023-06-01`.

**Rule:** Use `2023-06-01` as the `anthropic-version` header. When extended thinking is eventually needed, check Anthropic's docs for the correct version — don't guess.

---

## 12. buildActionParams never set AgentType

**Symptom:** `params.AgentType` was empty string in every action executed via the coordinator. Discovered when `llm_call_log` rows had empty `agent_type` column. Also affects `agent_error_log` entries and any diagnostic that uses `params.AgentType`.

**Root cause:** `buildActionParams()` in coordinator.go constructs `ActionParams` but never set the `AgentType` field. The struct has `AgentType string` but `buildActionParams` only set `CurrentStep`, `DB`, `Producer`, etc. — `AgentType` was left as zero value.

**Fix:** Added `AgentType: state.OwnerAgentType` to the `ActionParams` construction in `buildActionParams()`. `state.OwnerAgentType` is the agent type stored in `orchestration_states` — it's `"page-content-writer"`, `"site-classifier"`, etc.

**Impact:** This affected every action executed through the coordinator, not just LLM logging. Any code path that read `params.AgentType` was getting empty string. The fix is system-wide.

**Rule:** When adding fields to `ActionParams`, verify they're populated in `buildActionParams()`. The struct definition and the construction site are in different files — fields can be added to one and forgotten in the other.

---

## 13. scanPageRowsForBuild missing page_name key

**Symptom:** `load_page_record` failed with `missing required fields: [page_name]` when processing `needs_content_page` work items created by `WriteBuildItemsAction`.

**Root cause:** `scanPageRowsForBuild` builds a page map with `"name": name` but not `"page_name"`. `WriteBuildItemsAction` marshals this map directly as the work item spec (`json.Marshal(page)`). The dispatch loop's `input_mapping` has `"page_name?": "current_item.spec.page_name"` — resolves to nil, silently skipped via `?`. The handler's `load_page_record` action requires `page_name` — fails.

Other sources of `needs_content_page` items (`content-gap-planner`, `blog-content-planner`) already include `page_name` in their specs because they build the spec map manually rather than marshalling the raw page query result.

**Fix:** Added `"page_name": name` alongside `"name": name` in both `scanPageRowsForBuild` and `scanPageRowsForBuildPgx`. Backfilled existing work items:
```sql
UPDATE site_work_items 
SET spec = spec || jsonb_build_object('page_name', spec->>'name')
WHERE spec ? 'name' AND NOT (spec ? 'page_name')
  AND item_type = 'needs_content_page';
```

**Rule:** When a Go function produces data that becomes a work item spec (via `json.Marshal`), check what field names the downstream handler expects. The dispatch loop's `input_mapping` is the contract — if it maps `page_name?`, the spec must contain `page_name`. The `?` suffix hides the mismatch by silently skipping.

---

## 14. llm_call_log table schema didn't match Go INSERT columns

**Symptom:** `llm_call_log` table had 0 rows despite LLM calls succeeding. No errors visible because `LogLLMCall` runs in a fire-and-forget goroutine — INSERT failures are logged as warnings, not errors, and easily missed in high-volume pod logs.

**Root cause:** The Go code inserts into `agent_id` column, but the table (from migration 081) has `client_id` instead. The INSERT also sends `nullIfEmpty(params.StepName)` for a `NOT NULL` column, and `nullIfEmpty(params.PromptRendered)` for another `NOT NULL` column. Every INSERT silently failed in the goroutine.

This happened because the migration SQL and the Go code were written in separate sessions and the column names drifted. The Go code was deployed without testing against the actual table.

**Fix:** ALTER table to match the deployed Go code:
```sql
ALTER TABLE llm_call_log ADD COLUMN IF NOT EXISTS agent_id VARCHAR(255);
ALTER TABLE llm_call_log ALTER COLUMN step_name DROP NOT NULL;
ALTER TABLE llm_call_log ALTER COLUMN prompt_rendered DROP NOT NULL;
```

**Rule:** When deploying fire-and-forget logging, test the INSERT against the actual table schema before considering it done. Goroutine errors are invisible in normal log review. After deploying logging code, always verify with `SELECT COUNT(*) FROM llm_call_log` after triggering a few LLM calls.

**Broader lesson:** Fire-and-forget patterns need explicit verification. The goroutine swallows errors gracefully (by design — logging should never block the workflow), but this means you must actively check that rows are appearing, not assume silence means success.

---


## 15. PostgreSQL to_jsonb() fails with "could not determine polymorphic type"

**Symptom:** `UPDATE agent_definitions SET default_config = jsonb_set(..., to_jsonb(E'long string'))` fails with `could not determine polymorphic type because input has type unknown`.

**Root cause:** `to_jsonb()` is a polymorphic function — it accepts `anyelement`. When PostgreSQL sees an untyped string literal (even with `E''` prefix), it can't infer which overload to use.

**Fix:** Cast the string explicitly:

```sql
-- WRONG: untyped literal
to_jsonb(E'some text with \n newlines')

-- RIGHT: plain string with explicit cast
to_jsonb('multi-line
string content
goes here'::text)
```

Use plain single-quoted strings with real newlines rather than E-strings with `\n` — they're easier to read and avoid the escaping layer. PostgreSQL handles multi-line single-quoted strings natively.

**Rule:** Always use `to_jsonb('...'::text)` when converting string literals to JSONB. This comes up frequently when updating LLM prompt templates in agent definitions via `jsonb_set()`.

---

## Domain Submission: Trigger Script Reference

The `domain-submitter` agent is the entry point for all new site builds. Trigger scripts send a Kafka message to `system.agent.generic.requests` with `"agent_type": "domain-submitter"` and an `input_data` payload. The domain-submitter creates the site record, persists data to `site_specs`, and creates the first work item (`needs_domain_research`) for the build pipeline.

### input_data fields

**Always required:**

| Field | Type | Purpose |
|---|---|---|
| `domain` | string | The site domain (e.g. "vonc.com") |

**Optional (existing):**

| Field | Type | Purpose |
|---|---|---|
| `email` | string | Contact email, stored on site record |
| `phone` | string | Contact phone, stored on site record |
| `objective` | string | Free text hint for the classifier |

**Optional (mission-driven sites):**

| Field | Type | Stored as | Read by |
|---|---|---|---|
| `mission` | JSON object | `site_specs` aspect `mission` | Content writers (page-level content_context, archetype descriptions) |
| `mission_brief` | string | `site_specs` aspect `mission_brief` | Classifier prompt, planner prompt |
| `roadmap` | JSON object | `site_specs` aspect `roadmap` | plan_sections (page specs, section_types for component selector) |
| `roadmap_brief` | string | `site_specs` aspect `roadmap_brief` | Planner prompt |

### Three tiers of domain submission

**Tier 1 — Domain only.** Classifier researches the domain and infers everything. All current sites use this.

```json
{"domain": "dartsonline.com", "email": "darts@example.com", "phone": "+44 ..."}
```

**Tier 2 — Domain with objective.** Classifier uses the objective as a hint alongside research.

```json
{"domain": "dartsonline.com", "objective": "Online darts shop with scoring tools", "email": "..."}
```

**Tier 3 — Domain with mission and roadmap.** Classifier uses mission_brief as strong guidance. Planner uses roadmap_brief to know what pages and section_types to output. Research still runs — validates and supplements.

```json
{
  "domain": "vonc.com",
  "email": "...",
  "objective": "Short summary for the classifier",
  "mission": { "structured JSON for machine consumers" },
  "mission_brief": "Plain text readable by any model",
  "roadmap": { "structured JSON with phases, pages, section_types" },
  "roadmap_brief": "Plain text summary of current phase and pages"
}
```

### Rules for writing briefs

Briefs are the human-readable version of the structured data. They must be plain text that any model (including small ones) can parse. No JSON, no nested structures, no formatting that requires intelligence to decode.

**mission_brief** should cover: what it is, positioning, tagline, key differentiators (as a bullet list), target users, content tone (what it is and what it isn't), AI role if relevant.

**roadmap_brief** should cover: current phase name, what it delivers, design direction. Then a list of pages with their purpose and section_types. Future phases get a one-line summary each with a note that they're not for building now.

### When to use structured vs brief

The structured `mission` and `roadmap` are only needed when downstream machine consumers need specific fields — content_context per page (archetype descriptions, comparison text), section_types arrays for plan_sections to pass to the component selector. If you don't have page-level structured context, you can skip the structured versions and just provide the briefs.

The classifier and planner only read the briefs. They never access nested fields in the structured mission or roadmap.

### When to use which tier

- Normal commercial site, no special requirements → Tier 1
- Commercial site with specific direction → Tier 2
- Pre-planned site with known pages and novel section types → Tier 3

### Domain-submitter workflow (current)

```
ensure_site_record → store_contact_info → store_submission_spec
  → persist_mission → persist_mission_brief → persist_roadmap → persist_roadmap_brief
  → create_research_item → complete
```

The persist steps use `error_step` to skip gracefully when the field isn't in `input_data`. Tier 1 and 2 submissions pass through the persist steps without storing anything — they proceed straight to `create_research_item` via the error_step chain.

---

## 16. error_step must be inside step.Config, not at the step level

**Symptom:** Workflow step fails, but instead of routing to the error_step, the entire orchestration fails with `FAILED` status.

**Root cause:** The coordinator checks `step.Config["error_step"]` — not `step.ErrorStep` or any step-level field. When `error_step` is placed at the step level in the agent definition JSON, it gets parsed into the Step struct but the coordinator never reads it for error routing.

```json
// WRONG — coordinator never sees this
"persist_mission": {
    "action": "write_site_spec",
    "config": { "aspect": "mission", ... },
    "error_step": "persist_mission_brief",     // ← step level, ignored
    "next_step": "persist_mission_brief"
}

// RIGHT — coordinator reads this
"persist_mission": {
    "action": "write_site_spec",
    "config": {
        "aspect": "mission",
        "error_step": "persist_mission_brief"   // ← inside config, found
    },
    "next_step": "persist_mission_brief"
}
```

**Rule:** Always put `error_step` inside the `config` map. The coordinator's `routeToErrorStepOrFail` function reads `step.Config["error_step"]`. Step-level `error_step` is silently ignored — the workflow fails instead of routing.

**Broader lesson:** Several existing agent definitions had this bug (classifier write steps) but it never surfaced because those steps don't fail in normal operation. When adding error_step to any step, verify with `value->'config'->>'error_step'` not `value->>'error_step'`.

---

## 17. write_site_spec rejects plain strings — wrap text in JSON objects

**Symptom:** `spec_data must be a JSON object, got string` when persisting text briefs via write_site_spec.

**Root cause:** The `site_specs.data` column is JSONB. The `write_site_spec` action validates that spec_data is a JSON object (map), not a scalar. Plain text strings like mission briefs are not valid JSONB objects.

**Fix:** Wrap text values in a JSON object: `{"text": "the brief content..."}`. The prompt template accesses via `{{.site_specs.specs.mission_brief.text}}`.

```json
// WRONG — plain string, rejected by write_site_spec
"mission_brief": "Spark is an AI-driven social platform..."

// RIGHT — JSON object, accepted
"mission_brief": {"text": "Spark is an AI-driven social platform..."}
```

**Rule:** Any value persisted via write_site_spec must be a JSON object. For text content, wrap in `{"text": "..."}`. For structured data (mission, roadmap), it's already an object.

---

## 18. Schema column renames — always check the live schema

**Symptom:** `column "domain" does not exist` when inserting into site_work_items.

**Root cause:** The `domain` column on site_work_items was renamed to `pipeline` in a migration. The project's `some_schemas` dump still showed `domain`. Go code written against the stale dump used the wrong column name. The INSERT silently failed, caught by error handling, and logged as a warning in pod logs (not in agent_error_log).

**Fix:** Always run `\d table_name` against the live database before writing SQL or Go code that references columns. Never trust cached schema dumps in the repository.

**Rule:** The live database is the source of truth for column names. Schema dumps go stale. `\d site_work_items` takes 2 seconds and prevents hours of debugging invisible failures.

### Second case: version_note → change_description on component_versions (April 2026)

`UpdateComponentHTMLAction` wrote to `component_versions.version_note`. The live schema had `change_description`. The INSERT failed on every call. But because the snapshot was marked best-effort (the failure was logged as a Warn and ignored), the bug was invisible in normal flow — no observable failure, just silent data loss. Every component update that should have left a history row left nothing.

This was discovered by code audit during the April 2026 component regeneration work, when we added `StoreGeneratedComponentAction` regeneration semantics that rely on snapshots. Two years' worth of `update_component_html` calls produced no `component_versions` rows before the fix.

**Lesson:** `best-effort` operations need active monitoring. A persistent snapshot failure should eventually surface — either in metrics (e.g. "snapshots attempted vs committed" ratio) or via a smoke test that asserts at least ONE row got written. Silent best-effort is really "silent no-effort" when the code path is always broken.

---

## 19. LLM output with markdown code blocks breaks JSON parsing

**Symptom:** Generated component templates stored as JSON blobs instead of HTML. Pages render raw `{"function": "provocation-card", "html_template": "<style>..."}` text.

**Root cause:** The component-creator's LLM prompt asks for JSON output. The LLM wraps it in ` ```json ... ``` ` markdown code blocks. `json.Unmarshal` fails on the backtick-prefixed string. The parser's fallback treats the entire string (including the JSON structure) as "raw HTML" and stores it directly in `html_template`.

**Fix:** Strip markdown code blocks before JSON parsing:

```go
func stripCodeBlocks(s string) string {
    s = strings.TrimSpace(s)
    if strings.HasPrefix(s, "```") {
        if idx := strings.Index(s, "\n"); idx != -1 {
            s = s[idx+1:]
        }
        if strings.HasSuffix(s, "```") {
            s = s[:len(s)-3]
        }
        s = strings.TrimSpace(s)
    }
    return s
}
```

**Rule:** Any action that parses LLM output as JSON must strip markdown code blocks first. LLMs wrap JSON in code blocks even when told not to. This is defensive, not optional.

**Data recovery:** When JSON is stored as the html_template, PostgreSQL's `::jsonb->>'html_template'` can extract it — but only if the JSON is valid. LLM output often contains unescaped quotes in SVG paths (`d="M2 6l3 3 5-5"`) that break JSON parsing. In that case, delete the broken components and regenerate them with the fixed parser.

---

## 20. Handlers read work item spec from input_data.spec, not input_data directly

**Symptom:** Component-creator prompt receives empty section_type. LLM generates generic template. Store action finds generic function name already exists, returns "already_exists".

**Root cause:** The dispatch loop's input_mapping passes:
```json
{
    "spec": "current_item.spec",
    "site_id": "current_item.site_id",
    "work_item_id": "current_item.id"
}
```

The work item's spec fields (section_type, site_type, etc.) arrive at `input_data.spec.section_type`, not `input_data.section_type`. The component-creator's prompt referenced `{{.input_data.section_type}}` which rendered empty.

**Fix:** Agent prompt templates must reference `{{.input_data.spec.section_type}}` etc.

**Rule:** In handlers called via the dispatch loop, the work item's spec is at `input_data.spec.*`. Go actions using `ExtractActionInputs` handle nested lookups automatically. Prompt templates using Go template syntax (`{{.field}}`) must reference the full path.

**The dispatch loop is generic.** Don't add handler-specific field promotions to its input_mapping. Handlers should know where their data lives. The `page_name?` promotion is a legacy convenience, not a pattern to follow.

---

## Extended Thinking Configuration

When `budget_tokens` is set in an LLM step's `ai_service` config, the Anthropic client adds extended thinking support:

```json
"ai_service": {
    "model": "claude-opus-4-6",
    "provider": "anthropic",
    "max_tokens": 4000,
    "api_key_env_var": "ANTHROPIC_API_KEY",
    "budget_tokens": 10000
}
```

When `budget_tokens` is set: the client adds `{"thinking": {"type": "enabled", "budget_tokens": N}}`, temperature is removed (Anthropic requirement), response parsing skips thinking blocks and returns only the text block, response time increases by 30-90 seconds.

When `budget_tokens` is absent or 0: standard behaviour, no thinking, temperature applied as normal.

---

## Work Item Lifecycle: Blocking, Unblocking, and Unresolved

### Three ways items get blocked

| How it got blocked | How it gets unblocked |
|---|---|
| Claim action: handler agent not in agent_definitions | Automatic: feasibility-recheck scheduled task promotes to triaged when agent appears |
| Spec says status: blocked | Manual: human or classifier changes spec status to planned, creates work item |
| Manual block by human | Manual: `UPDATE site_work_items SET status = 'triaged' WHERE id = '...'` |

For cases 1 and 3:

```sql
-- Unblock all items for a specific handler (after deploying that handler)
UPDATE site_work_items
SET status = 'triaged', error = NULL, updated_at = NOW()
WHERE handler_agent = 'tool-deployer' AND status = 'blocked';
```

For case 2 (promoting a spec entry):

```sql
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, priority,
    handler_agent, status, created_by, summary, spec
) VALUES (
    '<site_id>', 'spec', 'build', 'needs_content_page', 60,
    'page-build-handler', 'triaged', 'human',
    'Build pricing page (previously blocked)',
    '{"page_name": "pricing", "sections": ["hero", "pricing-tiers", "faq", "call-to-action"]}'::jsonb
);
```

### The unresolved mechanism

Located in `load_work_item_actions.go`, in `insertWorkItem`:

When a discovery check calls `insertWorkItem`, the action queries for terminal items (complete/failed) with the same `item_key` in the last 7 days. If the newest terminal item is less than 3 hours old, the new item is suppressed entirely. If 2+ terminal items exist (indicating repeated fix-then-rediscovery), the new item is created with status `unresolved` and summary prefixed with `[unresolved after N attempts]`.

Unresolved items are not dispatched — they sit visible for investigation. The `attempt_count = 0` is expected — these are new items born as unresolved, not retried items.

---

## Nav Sync: Config-Driven Page Deactivation

### Problem

Header and footer navigation display pages from previous builds while hiding newly built pages. Two issues compound:

1. `SyncPagesToDBAction` uses `ON CONFLICT (site_id, name)` — only overwrites matching names. Pages from prior builds persist with `in_header=true`.
2. Nav queries filter on `status IN ('deployed', 'active')` but not on `build_status`.

### Fixes

1. `GetHeaderNavFromPages` and `GetFooterNavFromPages`: add `AND build_status = 'deployed'` to WHERE clause.
2. `SyncPagesToDBAction`: before syncing, deactivate pages not in the current plan: `UPDATE pages SET in_header = false, in_footer = false WHERE site_id = $1 AND name != ALL($2)`.

### Design decision

New build flow: deactivate stale pages. Future maintenance/adopt-site flows: should NOT deactivate — they preserve existing content. This is config-driven (`deactivate_stale_pages: true` in workflow step config).

---

## Summary of rules for the dev guide

1. **Check schema before writing SQL** — `\d table_name` every time
2. **Use parameterised queries** — `$1` + `"params"`, never `{{.field}}` in SQL
3. **Include api_key_env_var** in every LLM step config
4. **Use model aliases** — `claude-sonnet-4-6` not full version strings
5. **Check specialist vs handler** — does the agent persist its outputs?
6. **Verify action config support** — read the Go source, not just the docs
7. **Guard templates against nil** — `{{if}}` before `{{range}}`, test with empty data
8. **Run agent definition SQL before testing** — chassis silently handles missing defs
9. **Use read_site_spec for site config** — handles fallback from content_data
10. **Keep workflows simple** — put parsing/creation logic in Go actions
11. **Verify ActionParams population** — check `buildActionParams()` sets every field your action reads
12. **Match spec keys to input_mapping** — if dispatch maps `page_name?`, spec must contain `page_name`
13. **Verify fire-and-forget logging** — goroutine errors are invisible, always check `SELECT COUNT(*)` after deploy
14. **Test INSERT against actual schema** — column name drift between SQL migrations and Go code is common
15. **Cast string literals in to_jsonb()** — use `to_jsonb('...'::text)`, untyped literals fail with polymorphic type error
16. **Put error_step inside step.Config** — coordinator reads `step.Config["error_step"]`, step-level error_step is ignored
17. **Wrap text values for write_site_spec** — plain strings rejected, use `{"text": "..."}`
18. **Check live schema, not dumps** — `\d table_name` against the live DB, cached dumps go stale
19. **Strip markdown code blocks from LLM output** — LLMs wrap JSON in ``` blocks, parse fails silently
20. **Handlers read spec from input_data.spec** — dispatch loop passes work item spec at `input_data.spec.*`, not top-level

---

## Appendix B — Workflow Data Path Validation Checklist

Mandatory checklist before deploying any new or modified agent workflow. Prevents the most common runtime failure: a step referencing a path that doesn't exist because an earlier step skipped, returned a different shape, or was wired to the wrong source.

### When to Use This Checklist

- Creating a new agent definition
- Adding or reordering steps in an existing workflow
- Changing an `input_mapping` or `output_field`
- Changing what an LLM prompt returns
- Changing which action a step uses
- Connecting an audit finding to a handler agent

### Step 1: Schema Check

Before writing any SQL that references database tables: `\d table_name`. Common column name mistakes: `slug` vs `url` (pages), `spec_type` vs `aspect` (site_specs), `spec_data` vs `data` (site_specs), `domain` vs `pipeline` (site_work_items).

### Step 2: Trace Every Data Path

For each step, fill in: what it reads, which step produced it, whether that step can skip, what it writes, and what the skip shape looks like. The last row that shows a crash reveals the bug before deployment.

### Step 3: Guard Every Skip

If step N can skip, step N+1 must not blindly read N's output. Patterns: conditional guard (`page_content.response.skipped != true`), error_step fallback, optional input_mapping fields (`?` suffix).

### Step 4: Verify Action Config Support

Many actions have hardcoded paths that ignore config fields. Check the Go source: does the action read `config["my_field"]` or hardcode `collected_data["site_plan"]["pages"]`?

### Step 5: Check Input/Output Contracts

```sql
SELECT type, input_contract, output_contract
FROM agent_definitions WHERE type = 'my-agent';
```

Caller's `input_mapping` must produce what callee's `input_contract.required` expects.

### Step 6: Check Template Variables

Every `{{.field}}` in a prompt must exist in template data at render time. Go template failure modes: `{{.missing}}` → empty string (silent), `{{.missing.nested}}` → execution error → empty result, `{{range .nil_array}}` → error → empty. Always wrap in `{{if}}`.

### Step 7: Verify Audit-to-Handler Handoff

Handlers should be self-sufficient. They receive identifiers (site_id, page_name, work_item_id) and load everything else from the DB. Never assume the work item spec contains structural data.

### Common Data Shapes Reference

| Action | Success shape | Skip/empty shape |
|---|---|---|
| `ensure_site_record` | `{site_id, domain, status, content_data}` | Errors on missing |
| `read_site_spec` (all) | `{specs: {identity: {...}}, aspect_count: N}` | `{found: false, data: {}}` |
| `query_database` (array) | `[{...}, {...}]` | `null` (not `[]`) |
| `query_database` (single_row) | `{id, name, ...}` | `null` |
| `execute_llm_prompt` | `{type: "text", result: "..."}` | `{type: "text", result: ""}` |
| `call_agent` (success) | `{response: {...}, response_status: "complete"}` | `{response: {skipped: true}}` |
| `claim_work_item` | `{claimed: true, work_item_id: "..."}` | `{claimed: false}` |
| `load_work_items` | `{has_items: true, count: N, items: [...]}` | `{has_items: false, count: 0, items: []}` |
| `spawn_agent` | `{agent_id, agent_type, role, requests_topic}` | N/A |

### Pre-Deploy Checklist

- [ ] `\d table_name` for every table referenced in queries
- [ ] Every dot-path in config/input_mapping traced to its source step
- [ ] Every step's skip behaviour documented, downstream steps guarded
- [ ] Go source checked for hardcoded paths vs configurable paths
- [ ] Caller's input_mapping produces what callee's input_contract requires
- [ ] Every `{{.field}}` exists in template data, deep paths wrapped in `{{if}}`
- [ ] Handlers load structural data from DB, don't rely on work item spec

---

## Appendix C — Loop Mechanisms

Full reference for the loop system. Replaces 014_loop_mechanisms_guide.md.


## Overview

Loops allow a workflow to iterate over a collection of items, executing a series of substeps for each item. Loops are not Go for-loops — they are dynamic workflow expansion. At runtime, the loop step injects N × M steps into the workflow plan (N items × M substeps per iteration), then the normal step-execution engine drives through them sequentially.

This guide covers everything needed to define, debug, and extend loops in agent workflows.

---

## How Loops Work — The Full Lifecycle

### Phase 1: Definition (SQL workflow)

A loop is defined as a step with `"action": "loop"` in the agent's workflow definition. It specifies what collection to iterate over, what variable name to use for the current item, and what substeps to run per iteration.

### Phase 2: LoopAction (Go action)

When `continueExecution` reaches the loop step, it calls `LoopAction` in `loop_actions.go`. This action does not execute the iterations — it reads the config, resolves the collection from `CollectedData`, parses the substeps into `models.Step` structs, determines their order, and returns a special result map with `"loop_action": true`.

### Phase 3: Loop Expansion (coordinator)

Back in `executeLocalAction` (coordinator.go), the code detects `loop_action: true` in the result and calls `handleLoopExpansion` in `loop_expansion_handler.go`. This is where the real work happens:

1. For each item in the collection, for each substep in order, a new step is injected into `state.WorkflowPlan.Steps` with a name following the pattern `{loop_name}_iter_{N}_{substep_name}`.

2. Each injected step gets iteration-specific config injected: `loop_iteration`, `loop_item_index`, `loop_var_name`, `loop_name`, `continue_on_error`.

3. Each item is stored in `CollectedData` at `{loop_name}_item_{N}`.

4. A `{loop_name}_complete` step is created with action `loop_complete` that aggregates results after all iterations finish.

5. `state.CurrentStep` is set to the first injected step.

6. State is persisted to the database.

7. `executeLocalAction` calls `continueExecution` recursively to begin executing the first iteration's substeps.

8. After the recursive call returns, `executeLocalAction` returns `ErrLoopExpansionHandled` — a sentinel that tells the outer `continueExecution` for-loop to exit cleanly without falling through to the loop step's `NextStep`.

### Phase 4: Iteration Execution

The injected steps execute as normal workflow steps. Before each loop substep runs, `setLoopVariable` sets the current item in `CollectedData` under the configured variable name (e.g. `current_section` or `current_item`). It also propagates outputs from earlier substeps in the same iteration so they're accessible under their base names (not just their iteration-suffixed names).

### Phase 5: Iteration Boundaries

When an iteration's last substep finishes, its `NextStep` points to either:
- The first substep of the next iteration (e.g. `{loop}_iter_{N+1}_{first_substep}`), or
- `{loop_name}_complete` if this was the last iteration.

This is resolved at expansion time by `resolveIterationNextStep`.

### Phase 6: Loop Complete

`LoopCompleteAction` runs as the final step. It iterates through `0..total_iterations-1`, collecting each iteration's substep outputs from `CollectedData` using the `substep_output_fields` list (set during expansion). It assembles these into an array of result objects and stores them under the loop step's `output_field`.

After `loop_complete`, execution continues to whatever `next_step` was defined on the original loop step.

---

## Defining a Loop in a Workflow

### Minimal Example

```json
"process_items": {
    "action": "loop",
    "config": {
        "iterate_over": "loaded_data.items",
        "loop_var": "current_item",
        "max_iterations": 20,
        "sub_workflow": {
            "start_step": "do_work",
            "steps": {
                "do_work": {
                    "action": "some_action",
                    "config": {
                        "input_from": "current_item.field_name"
                    },
                    "output_field": "work_result",
                    "description": "Process the item"
                }
            }
        }
    },
    "next_step": "after_loop",
    "description": "Process each item",
    "output_field": "items_processed"
}
```

### Config Fields

| Field | Required | Description |
|-------|----------|-------------|
| `iterate_over` or `items_field` | Yes | Dot-notation path into `CollectedData` to find the array. E.g. `"pending.items"` or `"section_components.components"`. |
| `loop_var` or `item_variable` | No (default: `"loop_item"`) | Name under which the current item is stored in `CollectedData` before each substep runs. |
| `max_iterations` | No (default: 20) | Safety limit. Collections larger than this are truncated. |
| `continue_on_error` | No (default: false) | When true, a failed substep records the error and skips to the next iteration instead of failing the workflow. |
| `allow_missing` | No (default: false) | When true, a missing collection at `iterate_over` returns an empty result rather than an error. |
| `sub_workflow.start_step` | Recommended | Name of the first substep. If omitted, the framework auto-detects by finding the substep with no incoming `next_step` references. Specify it explicitly to avoid ambiguity. |
| `sub_workflow.steps` | Yes | Map of substep definitions. Same structure as top-level workflow steps. |

The older `substeps` key (without the `sub_workflow` wrapper) also works but `sub_workflow` is the standard form used by workflow definitions.

### The Loop Step Itself

| Field | Description |
|-------|-------------|
| `next_step` | Where to go after the loop completes (after `loop_complete` runs). |
| `output_field` | The key in `CollectedData` where the aggregated results array is stored. |
| `description` | For logging. |

---

## Substep Definitions

Substeps are defined the same as regular workflow steps. Each substep has `action`, `config`, `output_field`, `next_step`, and `description`.

### How next_step Works Inside Substeps

Substep `next_step` values are resolved at expansion time:

1. **References another substep name** → Prefixed with iteration context. `"next_step": "generate_content"` becomes `{loop}_iter_{N}_generate_content`.

2. **Empty string or omitted** → This is the terminal substep of the iteration. It chains to the first substep of the next iteration, or to `{loop}_complete` on the last iteration.

3. **References a step name that is NOT a substep** → Left as-is. This would exit the loop early (unusual, but supported for edge cases like `"done"` mapping to `loop_complete`).

### How Config References Are Prefixed

Config fields that reference other substeps (`then_step`, `else_step`, `fallback_step`, `error_step`, `on_success`, `on_failure`) are also prefixed with the iteration index at expansion time. This means conditionals inside loops work correctly.

Config fields that reference data produced by other substeps (`content_from`, `context_from`, `data_from`, `source_field`, `input_from`, `result_from`, `content_field`, `commit_from`) have their values prefixed when they match a substep's `output_field`. For example, if substep `generate_content` has `output_field: "generated_content"`, and a later substep has `"content_from": "generated_content.result"`, the expansion rewrites this to `"content_from": "generated_content_{N}.result"` for iteration N.

### OutputField Suffixing

Each substep's `output_field` is made unique per iteration by `makeIterationOutputField`. A substep with `output_field: "claim_result"` stores its result at `claim_result_0`, `claim_result_1`, etc. in `CollectedData`.

However, `setLoopVariable` propagates these back to their base names before each substep runs. So within an iteration, a substep can reference `claim_result` without the suffix and it will resolve to the current iteration's value.

---

## Accessing Data Inside Loop Substeps

### The Loop Variable

Before each substep executes, `setLoopVariable` sets:
- `CollectedData[loop_var]` = the current item (e.g. `CollectedData["current_item"]` = the Nth item from the collection)
- `CollectedData["__current_loop_item_key__"]` = `"{loop_name}_item_{N}"` (for debugging)

### Outputs from Earlier Substeps in the Same Iteration

`setLoopVariable` also propagates iteration-specific outputs to their base names. If iteration 2's `claim` substep stored a result at `claim_result_2`, then before the `check_claim` substep runs, `setLoopVariable` copies `claim_result_2` to `claim_result`. This means substep configs can reference other substep outputs by their base `output_field` name without worrying about iteration suffixes.

### Outputs from the Parent Workflow

Substeps can access anything in `CollectedData` that was set before the loop started. For example, if a prior step stored `site_record`, any substep can reference `site_record.domain`.

### Outputs from Other Iterations

These are accessible via their suffixed keys: `claim_result_0`, `claim_result_1`, etc. But typically substeps only care about their own iteration's data.

---

## The Dispatch Loop Pattern

The most common loop pattern in the system is the dispatch loop: claim an item, spawn a handler, call the handler, mark complete. Here is the production example from `build-dispatch-loop`:

```json
"process_items": {
    "action": "loop",
    "config": {
        "items_field": "pending.items",
        "item_variable": "current_item",
        "max_iterations": 50,
        "continue_on_error": true,
        "sub_workflow": {
            "start_step": "claim",
            "steps": {
                "claim": {
                    "action": "claim_work_item",
                    "config": { "work_item_id": "current_item.id" },
                    "next_step": "check_claim",
                    "output_field": "claim_result",
                    "description": "Atomically claim item"
                },
                "check_claim": {
                    "action": "conditional",
                    "config": {
                        "condition": "claim_result.claimed == true",
                        "then_step": "spawn_handler",
                        "else_step": "done"
                    },
                    "description": "Skip if already claimed"
                },
                "spawn_handler": {
                    "action": "spawn_agent",
                    "config": {
                        "role": "handler",
                        "agent_type_field": "current_item.handler_agent"
                    },
                    "next_step": "call_handler",
                    "output_field": "handler_spawned",
                    "description": "Spawn handler (dynamic type per item)"
                },
                "call_handler": {
                    "action": "call_agent",
                    "config": {
                        "target_role": "handler",
                        "error_step": "mark_failed",
                        "input_mapping": {
                            "site_id": "current_item.site_id",
                            "domain": "input_data.domain",
                            "work_item_id": "current_item.id",
                            "item_type": "current_item.item_type",
                            "spec": "current_item.spec"
                        },
                        "timeout_seconds": 300
                    },
                    "next_step": "mark_complete",
                    "output_field": "handler_result",
                    "description": "Call handler agent"
                },
                "mark_complete": {
                    "action": "complete_work_item",
                    "config": {
                        "work_item_id": "current_item.id",
                        "result": "handler_result"
                    },
                    "next_step": "done",
                    "output_field": "item_completed",
                    "description": "Mark item complete"
                },
                "mark_failed": {
                    "action": "fail_work_item",
                    "config": {
                        "work_item_id": "current_item.id",
                        "error_message": "Handler failed"
                    },
                    "next_step": "done",
                    "output_field": "item_failed",
                    "description": "Mark item failed"
                },
                "done": {
                    "action": "loop_complete",
                    "description": "Item done"
                }
            }
        }
    },
    "next_step": "complete",
    "output_field": "items_processed",
    "description": "Process each item: claim → spawn → call → mark"
}
```

Key design points:
- `continue_on_error: true` means one failed item doesn't kill the whole batch.
- `claim_work_item` uses atomic database claims to prevent double-dispatch across concurrent instances.
- `spawn_handler` uses `agent_type_field` for dynamic type resolution — the handler agent type comes from the work item data, not the workflow definition.
- `call_handler` uses `target_role: "handler"` to find the just-spawned agent. Always use role-based lookup, never type-based.
- The `done` step with `action: "loop_complete"` is the iteration's terminal substep. Its empty `next_step` means the framework chains to the next iteration or to `{loop}_complete`.
- `error_step: "mark_failed"` on `call_handler` routes handler failures to mark the item as failed rather than crashing the loop.

---

## Conditionals Inside Loops

Loops can contain conditional substeps. The `then_step` and `else_step` values reference other substep names (without iteration prefix — the expansion handles prefixing).

```json
"check_render_mode": {
    "action": "conditional",
    "config": {
        "condition": "current_section.render_mode == 'agent'",
        "then_step": "call_researcher",
        "else_step": "render_from_template"
    },
    "description": "Route based on render mode"
}
```

At expansion time for iteration 2, this becomes:

```
Step: process_sections_loop_iter_2_check_render_mode
Config:
  then_step: process_sections_loop_iter_2_call_researcher
  else_step: process_sections_loop_iter_2_render_from_template
```

The conditional action returns a `next_step` override in its result, which the `continueExecution` for-loop picks up via `getNextStepFromResult`.

---

## Substep Ordering and start_step

The framework determines substep execution order by following `next_step` chains, starting from `start_step`. The algorithm (`buildSubstepOrder`):

1. Use `start_step` if specified and exists in the substeps map.
2. Otherwise auto-detect: find the substep that no other substep references as its `next_step` (the entry point with no incoming edges).
3. Follow the `next_step` chain from there, collecting each step.
4. Any substeps not in the chain are appended at the end (with a warning log).

The order matters because it determines which substep is "first" for each iteration (used for iteration boundary chaining) and which is "last" (its empty `next_step` triggers the cross-iteration chain).

When using conditionals, some substeps will not be in the main chain because they're only reachable via `then_step`/`else_step`. These show up in logs as "Substep not in chain, appending" — this is expected and harmless. The chain order is only used for determining the first substep and for `substep_output_fields` ordering in the `_complete` step.

**Recommendation:** Always specify `start_step` explicitly. Auto-detection depends on `next_step` chain analysis which can be ambiguous when conditionals create branching paths.

---

## Error Handling

### continue_on_error

When `continue_on_error: true` is set on the loop config:

1. If a substep fails, `shouldContinueLoopOnError` checks the step's config for the `continue_on_error` flag (propagated during expansion).
2. `skipToNextLoopIteration` records the error at `{loop_name}_iter_{N}_error` in `CollectedData`, increments `{loop_name}_error_count`, and advances `CurrentStep` to the next iteration's first substep (or `_complete` if it was the last iteration).
3. Execution continues from the new step.

### error_step on Individual Substeps

Substeps can define an `error_step` in their config (e.g. `"error_step": "mark_failed"`). This is handled by the normal `routeToErrorStepOrFail` mechanism. Since `error_step` values that reference other substeps are prefixed during expansion, the error routes correctly within the iteration.

### Async Failures (Timeouts)

When a `call_agent` substep times out, `handleRequestTimeout` uses `skipToNextLoopIterationForAsync` which is the same as `skipToNextLoopIteration` but also cleans up the awaited request.

---

## Loop Completion and Result Aggregation

The `{loop_name}_complete` step runs `LoopCompleteAction`. Its config contains:

```json
{
    "loop_name": "process_items",
    "total_iterations": 4,
    "substep_output_fields": ["claim_result", "handler_result", "item_completed"]
}
```

`substep_output_fields` is built automatically during expansion from the substeps that have non-empty `output_field` values. `LoopCompleteAction` iterates through each iteration index, looks up `{field}_{N}` in `CollectedData` for each field, and assembles per-iteration result objects.

The aggregated array is stored at the loop step's `output_field` (e.g. `items_processed`), making it available to subsequent workflow steps.

---

## The Race Condition and ErrLoopExpansionHandled

### The Problem

When `executeLocalAction` finishes loop expansion, it calls `continueExecution` recursively to begin executing iteration substeps. If a substep calls a child agent that responds very quickly (under ~1 second), the response handler goroutine can advance the workflow state in the database before the recursive `continueExecution` returns.

When the recursive call returns nil, the outer `continueExecution` for-loop resumes with stale context. It checks the DB for `AWAITING_RESPONSES` status — but the response handler already moved past that. It then falls through to the step-transition logic using `currentStepConfig.NextStep`, which is the original loop step's `NextStep` (e.g. `"complete"`). This skips all remaining iterations.

### The Fix

`executeLocalAction` returns `ErrLoopExpansionHandled` after the recursive `continueExecution` call. The outer `continueExecution` catches this sentinel before any other error handling and returns nil — never reaching the stale `NextStep` transition.

This is safe in all timing scenarios:
- **Fast response:** Thread B (response handler) already advanced the workflow. Outer returns nil. Thread B continues.
- **Slow response:** Recursive CE returns nil (state persisted as AWAITING). Outer returns nil via sentinel. Workflow resumes when response arrives.
- **Error:** If the recursive `continueExecution` returns a real error, that error is returned instead of the sentinel, and normal error handling applies.

### Implication for New Loops

You don't need to do anything special to handle this — the sentinel is in the framework code, not in workflow definitions. But be aware that loops with fast-responding child agents are the trigger. If you're debugging a loop that completes too early (skipping iterations), check for this pattern first.

---

## Naming Conventions

| Thing | Pattern | Example |
|-------|---------|---------|
| Loop step name | Descriptive, often ending in `_loop` | `process_items`, `process_sections_loop` |
| Injected step | `{loop_name}_iter_{N}_{substep_name}` | `process_items_iter_2_claim` |
| Completion step | `{loop_name}_complete` | `process_items_complete` |
| Item storage key | `{loop_name}_item_{N}` | `process_items_item_3` |
| Iteration output | `{output_field}_{N}` | `claim_result_2` |
| Error key | `{loop_name}_iter_{N}_error` | `process_items_iter_1_error` |
| Error count | `{loop_name}_error_count` | `process_items_error_count` |
| Loop metadata | `loop_metadata` | Shared key — beware if multiple loops exist |

### The loop_metadata Caveat

`loop_metadata` is a shared key in `CollectedData`. If your workflow has two sequential loops, the second loop's expansion overwrites `loop_metadata`. The `_complete` step config has its own `loop_name` and `total_iterations` (set during expansion), so `LoopCompleteAction` reads from config first and only falls back to `loop_metadata` for backward compatibility. This means multiple sequential loops work correctly — but don't rely on `loop_metadata` for anything except debugging.

---

## Diagnostic Queries

### Check what the expansion produced

```sql
SELECT
    jsonb_object_keys(workflow_plan->'Steps') as step_name
FROM orchestration_states
WHERE orchestration_id = '<id>'
ORDER BY step_name;
```

Filter to just the loop's steps:

```sql
SELECT
    key as step_name,
    value->>'action' as action,
    value->>'next_step' as next_step,
    value->>'output_field' as output_field
FROM orchestration_states,
     jsonb_each(workflow_plan->'Steps')
WHERE orchestration_id = '<id>'
  AND key LIKE 'process_items_iter_%'
ORDER BY key;
```

### Check CollectedData for iteration results

```sql
SELECT
    key,
    CASE
        WHEN length(value::text) > 200 THEN left(value::text, 200) || '...'
        ELSE value::text
    END as value_preview
FROM orchestration_states,
     jsonb_each(collected_data::jsonb)
WHERE orchestration_id = '<id>'
  AND (key LIKE '%_iter_%' OR key LIKE '%_item_%' OR key = 'loop_metadata')
ORDER BY key;
```

### Check for stuck loops

```sql
SELECT
    orchestration_id,
    current_step,
    status,
    last_activity,
    NOW() - last_activity as stale_for
FROM orchestration_states
WHERE current_step LIKE '%_iter_%'
  AND status = 'AWAITING_RESPONSES'
  AND last_activity < NOW() - INTERVAL '5 minutes';
```

---

## Checklist for Adding a New Loop

1. **Ensure the collection exists.** The step that produces the collection must store it at a `CollectedData` path that matches your `iterate_over`/`items_field`. Test that the path resolves — dot-notation traversal uses `getNestedValueForLoop`.

2. **Choose a meaningful loop_var name.** This is how substeps access the current item. Use something descriptive: `current_item`, `current_section`, `current_fix_item`.

3. **Define substeps with explicit next_step chains.** Each substep should have a `next_step` pointing to the next substep in the flow, except the terminal substep which should have an empty or omitted `next_step`.

4. **Specify start_step.** Don't rely on auto-detection.

5. **Set output_field on substeps that produce data.** Only substeps with `output_field` have their results tracked in the aggregation.

6. **Set output_field on the loop step itself.** This is where the aggregated results go.

7. **Set continue_on_error if appropriate.** For dispatch loops processing independent items, this is almost always what you want. For loops where iteration N depends on iteration N-1, leave it false.

8. **Set max_iterations.** The default of 20 is usually fine, but if you know you'll process more items (e.g. dispatch loops), set it higher. The production dispatch loop uses 50.

9. **Test with fast-responding handlers.** If your loop calls child agents, test with agents that respond in under a second to confirm the race condition fix (ErrLoopExpansionHandled) is working.

10. **Don't nest loops inside loops in the same workflow.** If you need nested iteration, spawn a sub-agent whose workflow contains its own loop. This keeps logs clear and responsibilities separate.

---

## Files Reference

| File | What It Does |
|------|-------------|
| `platform/orchestration/actions/loop_actions.go` | `LoopAction` — reads config, resolves collection, parses substeps, returns expansion data |
| `platform/orchestration/loop_expansion_handler.go` | `handleLoopExpansion` — injects steps into workflow plan, creates `_complete` step |
| `platform/orchestration/coordinator.go` | `executeLocalAction` — detects loop expansion, calls `handleLoopExpansion`, recursive `continueExecution`, returns `ErrLoopExpansionHandled` |
| `platform/orchestration/coordinator.go` | `continueExecution` — main execution loop, catches sentinel, manages state transitions |
| `platform/orchestration/coordinator.go` | `setLoopVariable` — sets current item and propagates iteration outputs |
| `platform/orchestration/loop_error_handler.go` | `skipToNextLoopIteration`, `shouldContinueLoopOnError`, `findFirstSubstep`, `parseLoopStepName` |
| `platform/orchestration/actions/loop_actions.go` | `LoopCompleteAction` — aggregates iteration results |


