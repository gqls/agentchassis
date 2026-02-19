# Agent Design Decisions & Principles

Reference document for implementing new specialist agents in the orchestration framework.

## Core Principle: Agents Own Their Domain

Each specialist agent is **self-contained** and **independently callable**. The agent handles its own data gathering rather than relying on callers to provide everything.

**Why:** Multiple builder workflows (pageflow-builder, landing-page-builder, etc.) will call the same specialists. If callers must assemble complex input data, that logic gets duplicated. If agents load their own data, builders stay simple.

```
Builder workflow:
  call_agent: { "site_id": "uuid" }    â† Simple

Agent workflow:
  load_my_data â†’ analyze â†’ generate â†’ deploy    â† Agent handles complexity
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

## Decision: Callers Pass Raw Data, Agents Derive What They Need

Parent orchestrators and dispatchers pass **raw domain identifiers** to child agents. The child agent decides how to use them â€” callers should not build derived or computed values on behalf of the child.

**Good â€” caller passes raw data:**
```json
{
  "input_data": {
    "district_code": "BT4",
    "area_name": "Belfast",
    "business_type": "veterinary practice"
  }
}
```
The child agent's workflow then uses `query_template` to compose its own search query from those inputs.

**Bad â€” caller pre-builds derived data:**
```json
{
  "input_data": {
    "query": "veterinary practice BT4 Belfast UK"
  }
}
```
This leaks the child's search strategy into the parent. If the child changes how it searches, the parent must also change.

**Why:**
- The agent owns its domain â€” including how it transforms inputs into actions
- Keeps agents independently testable (pass `district_code`, get results)
- Multiple callers (orchestrator, shell script, another agent) don't duplicate derivation logic
- The agent can change its internal strategy without touching callers

**Test:** If you changed how the child agent works internally, would any caller need updating? If yes, you've leaked responsibility upward.

## Decision: Reuse Before Creating

Before creating new code, check for existing actions/functions that can be **patched** with small enhancements.

**Example - git_commit:**
- Existing: Uses `page_field` â†’ forces `.html` extension
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

### Declarative Config Is Not Complexity

Templates and config declarations in workflow JSON are fine â€” they express **intent**, not **logic**. The distinction is:

| Belongs in workflow config | Belongs in Go |
|---------------------------|---------------|
| `query_template`: how to compose a search query from inputs | Looping through results and filtering |
| `input_fields`: which fields an action needs | Conditional branching based on data |
| `num_results`: 10 | DB queries and data transformation |
| `include_pages`: true | Error handling and retry logic |

**Good â€” declarative template in workflow config:**
```json
{
  "action": "web_search",
  "config": {
    "query_template": "{{.input_data.business_type}} {{.input_data.district_code}} {{.input_data.area_name}} UK",
    "num_results": 10
  }
}
```
This is the agent declaring "here is how I build my search query from my inputs." It's readable, testable, and changes without recompiling.

**Bad â€” moving this to Go in the caller** to keep the workflow "simple":
```go
query := fmt.Sprintf("%s %s %s UK", businessType, districtCode, areaName)
// ... pass query to child agent's input_data
```
This moves domain knowledge out of the agent that owns it and into a parent that shouldn't care.

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

This enables pipelines like: scrape competitor â†’ feed to design agent â†’ apply to your site.

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

## Decision: Orchestrator Boundaries

Orchestrators and child agents have distinct responsibilities. Getting this boundary wrong causes tight coupling and maintenance pain.

**Orchestrator's job:**
- Know *what* needs doing and *in what order*
- Load batches of work from the database
- Dispatch child agents with raw domain identifiers
- Track overall progress

**Agent's job:**
- Know *how* to do the work
- Own its search strategy, data transformation, and domain logic
- Be independently callable and testable

**Example â€” area sweep:**
```
Orchestrator:  "Here are 50 districts that need sweeping"
               â†’ passes: { district_code, area_name, business_type }

Discoverer:    "I know how to search for businesses in a district"
               â†’ builds its own query, filters results, inserts candidates
```

The orchestrator doesn't know or care how the discoverer searches. It just hands off raw identifiers and collects results.

**Warning sign:** If changing an agent's internal approach requires updating its parent orchestrator, the boundary is wrong.

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

**Warning:** Check optional field names against the "Avoid Field Names That Collide with Common Nested Objects" section below. Names like `content_data`, `status`, `domain` will be found via nested lookup in `site_record.*` even if the caller never sent them.

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
Keep it simple: load â†’ analyze â†’ generate â†’ deploy â†’ complete

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

## Decision: Avoid Field Names That Collide with Common Nested Objects

`ExtractActionInputs` has a backward-compat nested lookup that checks these parent objects for any field it hasn't found directly:

```go
nestedSources := []struct{ parent, child string }{
    {"current_page", field},
    {"rerender_pages", field},
    {"site_record", field},
    {"input_data", field},
}
```

**This means any optional field name in your ActionInputSpec will also match `site_record.<your_field>`, `input_data.<your_field>`, etc.** If those parent objects happen to contain a key with the same name, ExtractActionInputs silently picks it up — even if the caller never sent it.

**Real example — the site plan contamination bug:**

The section-editor's spec had `content_data` as optional. `ensure_site_record` puts the site plan into `collected_data["site_record"]["content_data"]`. ExtractActionInputs found `site_record.content_data` via nested lookup and treated it as the caller's replacement data — overwriting the hero section with the site plan.

**Rules:**

1. **Never name an optional field the same as a common column/key** in `sites`, `pages`, or `site_record`. Watch out for: `content_data`, `status`, `domain`, `name`, `title`, `description`, `config`, `metadata`.

2. **If your field could collide, prefix it** with the operation context:
    - `replacement_content_data` instead of `content_data`
    - `target_page_name` instead of `page_name` (if `current_page.page_name` exists)

3. **Check collected_data at runtime** — if `ensure_site_record` runs before your action, `site_record.*` is in scope. If `load_edit_context` runs, `edit_context.*` is in scope. Any field name matching a key inside those objects will be found by the nested lookup.

4. **When in doubt, use explicit `input_fields`** in your step config and verify the extraction path by checking logs.

**Test:** Search for your field name across all action output_fields and common table columns. If it appears anywhere else in the pipeline, rename it.

---

## Decision: Optional Fields in input_mapping

When calling a child agent via `call_agent`, the `input_mapping` is strict by default — if a source path doesn't exist in collected_data, the mapping fails with an error. For agents that support multiple modes (where different fields are relevant per mode), use the `?` suffix on destination field names.

**Mechanism:** `ResolveInputMapping` checks `strings.HasSuffix(destField, "?")`. If the source path doesn't exist, the field is silently skipped. The `?` is stripped before storing the result.

```json
{
  "action": "call_agent",
  "config": {
    "agent_type": "section-editor",
    "input_mapping": {
      "domain": "input_data.domain",
      "edit_type": "input_data.edit_type",
      "page_name?": "input_data.page_name",
      "slot_name?": "input_data.slot_name",
      "field_updates?": "input_data.field_updates",
      "replacement_content_data?": "input_data.content_data",
      "new_component_function?": "input_data.new_component_function",
      "page_component_id?": "input_data.page_component_id"
    }
  }
}
```

**When to use:**
- Agent supports multiple edit modes (content_edit sends field_updates, component_swap sends new_component_function)
- Target identification has alternatives (page_component_id OR page_name+slot_name)
- Some fields are only relevant for certain callers

**When NOT to use:**
- Required fields that must always be present — let the mapping fail loudly
- Fields where a missing value indicates a caller bug