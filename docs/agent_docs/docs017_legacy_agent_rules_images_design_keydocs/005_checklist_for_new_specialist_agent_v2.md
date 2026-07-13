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

## Decision: Use input_fields, Not Explicit Paths

Actions should declare **which top-level fields they need**, not explicit nested paths with `input_data.` prefixes.

**Avoid - Explicit paths:**
```json
{
  "action": "validate_page_content",
  "config": {
    "html_field": "input_data.page_content.response.page_html",
    "site_id_field": "input_data.site_record.site_id"
  }
}
```

**Problems:**
- Caller must know whether data is under `input_data.*` or not
- Path breaks when data structure changes
- Inconsistent across workflows (sometimes `input_data.`, sometimes not)
- Config becomes fragile and hard to debug

**Correct - Use input_fields:**
```json
{
  "action": "validate_page_content",
  "config": {
    "input_fields": ["page_content", "site_record"],
    "check_internal_links": true
  }
}
```

**In the Go action:**
```go
// Get input_fields from config (with sensible defaults)
inputFields := []string{"page_content", "site_record"}
if fields, ok := config["input_fields"].([]interface{}); ok {
    inputFields = make([]string, len(fields))
    for i, f := range fields {
        inputFields[i], _ = f.(string)
    }
}

// Use standard extractor - handles input_data prefix automatically
extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

// Access nested data from extracted objects
if pageContent, ok := extracted["page_content"].(map[string]interface{}); ok {
    if response, ok := pageContent["response"].(map[string]interface{}); ok {
        html, _ = response["page_html"].(string)
    }
}
```

**Why this works:**
- `datahelpers.ExtractFields()` already handles looking in both `CollectedData` and `CollectedData["input_data"]`
- Action config stays simple and doesn't care about data location
- Same pattern used by `execute_llm_prompt` and other core actions
- Consistent across all agents regardless of how they're called

**When to use explicit paths:**
- Only for truly optional config overrides where the caller has special knowledge
- Always provide sensible defaults in the action code

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

## Checklist for New Specialist Agent

1. **Define the domain** - What single responsibility does this agent own?

2. **Design the load action** - What data does the agent need? Create `load_*_for_<purpose>` action.

3. **Check existing code** - Can existing actions be patched instead of creating new ones?

4. **Use input_fields pattern** - Actions declare field names, use `datahelpers.ExtractFields()`, avoid explicit paths.

5. **Define input schema** - What's the minimum input? What's the full context schema?

6. **Design the workflow** - Keep it simple: load → analyze → generate → deploy → complete

7. **Ensure standalone mode** - Agent works with just an ID/domain

8. **Plan integration** - How will builders spawn and call this agent?

9. **Register action** - Add to `action_registry.go`

10. **Create agent definition** - SQL insert with workflow, contracts, tags

11. **Test both modes** - Standalone call and integrated in a builder workflow