# Agent Design Decisions & Principles

Reference document for implementing new specialist agents in the orchestration framework.

## Core Principle: Agents Own Their Domain

Each specialist agent is **self-contained** and **independently callable**. The agent handles its own data gathering rather than relying on callers to provide everything.

**Why:** Multiple builder workflows (pageflow-builder, landing-page-builder, etc.) will call the same specialists. If callers must assemble complex input data, that logic gets duplicated. If agents load their own data, builders stay simple.

```
Builder workflow:
  call_agent: { "site_id": "uuid" }    ← Simple

Agent workflow:
  load_my_data → analyze → generate → deploy    ← Agent handles complexity
```

## Decision: Dedicated Load Actions

Each specialist agent gets a dedicated `load_*` action that gathers exactly what that agent needs.

| Agent | Load Action | Returns |
|-------|-------------|---------|
| webdesign-agent | `load_site_for_design` | site info, pages, components, colors, typography |
| link-manager | `load_site_for_links` | pages, navigation structure, internal links |
| seo-agent | `load_site_for_seo` | pages, meta tags, content for analysis |

**Why:**
- Different agents need different data shapes
- Keeps DB queries focused and efficient
- Agent can be tested standalone with just an ID

## Decision: Reuse Before Creating

Before creating new code, check for existing actions/functions that can be **patched** with small enhancements.

**Example - git_commit:**
- Existing: Uses `page_field` → forces `.html` extension
- Needed: Deploy CSS files
- Solution: 4-line patch adding `file_path` config override
- Avoided: New `git_commit_file` action with duplicated logic

**Example - WebscrapeAction:**
- Existing: Hardcoded `input_data.target_url`
- Needed: Flexible URL from `input_data.url`
- Solution: ~25-line patch adding `url_field` config
- Avoided: New `webscrape_request` action duplicating adapter calls

**Check existing code for:**
- Similar actions that could be extended
- Config options that could be added
- Shared helper functions

## Decision: Workflows Simple, Complexity in Go

Workflow definitions (JSON in agent_definitions) should be **declarative and readable**. Complex logic belongs in Go actions.

**Good - Simple workflow step:**
```json
{
  "action": "load_site_for_design",
  "config": {
    "include_pages": true
  }
}
```

**Avoid - Complex logic in workflow:**
```json
{
  "action": "transform_data",
  "config": {
    "if": "input_data.site_context != null",
    "then": { "mappings": { "site_context": "input_data.site_context" } },
    "else": { "call": "load_site_for_design", ... }
  }
}
```

If you need conditionals, branching, or data manipulation, put it in a Go action.

---

## Decision: Standardized Input Extraction

**All actions must use `ActionInputSpec` and `ExtractActionInputs()`** for input extraction. This replaces ad-hoc boilerplate in each action.

### The Three Layers

| Layer | Where | Purpose | Pattern |
|-------|-------|---------|---------|
| **input_mapping** | `call_agent` config | Caller maps data to child's input_data | `"site_id": "site_record.site_id"` |
| **input_fields** | Action config | Action declares needed fields | `"input_fields": ["site_id", "domain"]` |
| **ActionInputSpec** | Go action code | Standardized extraction with validation | `ExtractActionInputs(...)` |

### Workflow Config: input_mapping (Caller's Responsibility)

When calling a child agent, the caller uses `input_mapping` to prepare data:

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

This extracts values from the caller's CollectedData and puts them in the child's `input_data`:
```json
{ "input_data": { "page_id": "uuid", "site_id": "uuid", "domain": "example.com" } }
```

### Workflow Config: input_fields (Action's Declaration)

Actions declare which fields they need in their workflow step:

```json
{
  "action": "rerender_single_page",
  "config": {
    "input_fields": ["page_id", "site_id", "domain"],
    "max_nav_items": 6
  }
}
```

### Go Code: ActionInputSpec (Standardized Extraction)

**Every action** defines an ActionInputSpec and uses `ExtractActionInputs()`:

```go
// Define spec at package level
var RerenderSinglePageInputSpec = datahelpers.ActionInputSpec{
    Required: []string{"page_id", "site_id"},
    Optional: []string{"domain", "max_nav_items"},
    Defaults: map[string]interface{}{
        "max_nav_items": 6,
    },
    // Deprecated patterns - will log warnings but still work
    Deprecated: map[string]string{
        "page_id_field": "page_id",
        "site_id_field": "site_id",
        "domain_field":  "domain",
    },
}

func init() {
    // Register for documentation/contract generation
    datahelpers.RegisterActionInputSpec("rerender_single_page", RerenderSinglePageInputSpec)
}

func RerenderSinglePageAction(ctx context.Context, params ActionParams) (interface{}, error) {
    // One call handles everything
    inputs, err := datahelpers.ExtractActionInputs(
        params.CollectedData,
        params.StepConfig.Config,
        RerenderSinglePageInputSpec,
        params.Logger,
    )
    if err != nil {
        return nil, fmt.Errorf("input extraction failed: %w", err)
    }

    // Clean typed access
    pageIDStr := inputs.Get("page_id")      // string
    siteIDStr := inputs.Get("site_id")      // string
    maxNav := inputs.GetInt("max_nav_items", 6)  // int with default
    
    // ... actual business logic
}
```

### What ExtractActionInputs Does

1. **Tries input_fields from config** (preferred pattern)
2. **Falls back to deprecated *_field patterns** (logs warning)
3. **Checks nested objects** for backward compat (current_page.page_id, etc.)
4. **Validates required fields** - returns error if missing
5. **Applies defaults** for optional fields

### Deprecation Path

Old patterns still work but log warnings:

```
WARN: Using deprecated config pattern
      deprecated_key=page_id_field
      path=current_page.page_id
      use_instead=input_fields: ["page_id"]
```

### Benefits

- **No more boilerplate** - 40+ lines of extraction code per action eliminated
- **Consistent errors** - "missing required fields: [page_id]" everywhere
- **Deprecation tracking** - Know which workflows need migration
- **Self-documenting** - Spec defines what action needs
- **Contract generation** - Can auto-generate input_contract from spec

---

## Decision: No Container Config in Agent Definitions

Agent definitions should NOT include `container_image` or `image_tag`. The `spawn_actions.go` handles container configuration dynamically.

**Avoid:**
```json
{
  "type": "webdesign-agent",
  "container_image": "docker.io/aqls/agent-chassis",
  "image_tag": "v1.0.728"
}
```

**Correct:** Omit these fields entirely.

## Decision: Standardized Interface Schemas

When multiple sources can provide input to an agent, define a **standardized schema** that all sources conform to.

**Example - site_context schema:**
```json
{
  "domain": "string (required)",
  "company_name": "string",
  "industry": "string", 
  "color_palette": { "primary": "#hex", ... },
  "typography": { "font_family": "...", ... },
  "all_component_functions": ["hero", "cta", ...],
  "source": "database|scrape|manual"
}
```

**Sources that produce site_context:**
- `load_site_for_design` action (from database)
- `site-scraper` agent (from live URL)
- Direct input (manual/API)

**Consumer:**
- `webdesign-agent` accepts site_context regardless of source

This enables pipelines like: scrape competitor → feed to design agent → apply to your site.

## Decision: Standalone + Integrated

Every specialist agent must work in two modes:

1. **Standalone** - Called directly with minimal input
   ```json
   { "agent_type": "webdesign-agent", "data": { "site_id": "uuid" } }
   ```

2. **Integrated** - Called from builder workflows
   ```json
   {
     "action": "call_agent",
     "config": {
       "agent_type": "webdesign-agent",
       "input_mapping": { "site_id": "site_record.site_id" }
     }
   }
   ```

**Why:**
- Standalone enables testing, debugging, maintenance
- Integrated enables composition into larger workflows
- Same agent code serves both use cases

## Decision: Agents Respond to Caller's Topic

When spawned as part of a workflow, agents respond to their **parent's responses topic**, not their own. This is handled by the framework but important to understand.

## Decision: Spawn Before Call

In builder workflows, agents must be **spawned** before they can be **called**:

```json
{
  "spawn_webdesign_agent": {
    "action": "spawn_agent",
    "config": { "role": "webdesigner", "agent_type": "webdesign-agent" },
    "next_step": "...",
    "output_field": "webdesign_agent"
  }
}
```

Then later:
```json
{
  "apply_site_design": {
    "action": "call_agent",
    "config": {
      "agent_type": "webdesign-agent",
      "target_role": "webdesigner",
      "input_mapping": { "site_id": "site_record.site_id" }
    }
  }
}
```

---

## Checklist for New Specialist Agent

### 1. Define the domain
What single responsibility does this agent own?

### 2. Design the load action
What data does the agent need? Create `load_*_for_<purpose>` action.

### 3. Check existing code
Can existing actions be patched instead of creating new ones?

### 4. Define ActionInputSpec
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

### 5. Use ExtractActionInputs in action
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

### 6. Design the workflow
Keep it simple: load → analyze → generate → deploy → complete

### 7. Ensure standalone mode
Agent works with just an ID/domain

### 8. Plan integration
How will builders spawn and call this agent?

### 9. Register action
Add to `action_registry.go`

### 10. Create agent definition
SQL insert with workflow, contracts, tags

### 11. Test both modes
Standalone call and integrated in a builder workflow

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


---
note that agent_definitions has index on type, version and not just type

clients_db=# \d agent_definitions;
Table "public.agent_definitions"
Column         |           Type           | Collation | Nullable |                                                           Default                                                           
------------------------+--------------------------+-----------+----------+-----------------------------------------------------------------------------------------------------------------------------
id                     | uuid                     |           | not null | gen_random_uuid()
type                   | character varying(100)   |           | not null |
display_name           | character varying(255)   |           | not null |
description            | text                     |           |          |
category               | character varying(50)    |           | not null |
default_config         | jsonb                    |           | not null | '{}'::jsonb
is_active              | boolean                  |           |          | true
created_at             | timestamp with time zone |           | not null | now()
updated_at             | timestamp with time zone |           | not null | now()
deleted_at             | timestamp with time zone |           |          |
capabilities           | jsonb                    |           |          | '[]'::jsonb
image_repository       | character varying(255)   |           |          | 'docker.io/aqls/agent-chassis'::character varying
image_tag              | character varying(100)   |           |          | 'latest'::character varying
command                | text[]                   |           |          |
resources              | jsonb                    |           |          | '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb
topics                 | jsonb                    |           |          | '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb
health_config          | jsonb                    |           |          | '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb
env_vars               | jsonb                    |           |          | '[]'::jsonb
version                | integer                  |           |          | 1
previous_version_id    | uuid                     |           |          |
task_workflow          | jsonb                    |           |          |
orchestrator_workflow  | jsonb                    |           |          |
orchestration_workflow | json                     |           |          |
delegation_preferences | jsonb                    |           |          | '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb
agent_category         | text                     |           |          |
status                 | text                     |           |          | 'experimental'::text
domain_tags            | jsonb                    |           |          | '[]'::jsonb
briefing_questionnaire | jsonb                    |           |          | '{}'::jsonb
usage_count            | integer                  |           |          | 0
is_snapshot            | boolean                  |           |          | false
input_contract         | jsonb                    |           |          |
output_contract        | jsonb                    |           |          |
Indexes:
"agent_definitions_pkey" PRIMARY KEY, btree (id)
"agent_definitions_type_version_key" UNIQUE CONSTRAINT, btree (type, version)
"idx_ad_category" btree (agent_category)
"idx_ad_domain_tags" gin (domain_tags)
"idx_ad_status" btree (status)
"idx_agent_definitions_category" btree (category) WHERE is_active = true
"idx_agent_definitions_type_active" btree (type, is_active) WHERE deleted_at IS NULL
"idx_agent_definitions_type_version" btree (type, version DESC)
"idx_agent_definitions_usage" btree (usage_count DESC) WHERE is_active = true
"idx_agent_definitions_version" btree (type, version)
Check constraints:
"check_ad_category" CHECK (agent_category IS NULL OR (agent_category = ANY (ARRAY['strategist'::text, 'executor'::text, 'analyst'::text, 'integrator'::text, 'coordinator'::text, 'specialist'::text])))
"check_ad_status" CHECK (status = ANY (ARRAY['active'::text, 'experimental'::text, 'deprecated'::text, 'demo'::text, 'template'::text]))
Foreign-key constraints:
"agent_definitions_previous_version_id_fkey" FOREIGN KEY (previous_version_id) REFERENCES agent_definitions(id)
Referenced by:
TABLE "agent_definitions" CONSTRAINT "agent_definitions_previous_version_id_fkey" FOREIGN KEY (previous_version_id) REFERENCES agent_definitions(id)

