# 015 — Workflow Data Path Validation Guide

Mandatory checklist before deploying any new or modified agent workflow. Prevents the most common runtime failure: a step referencing a path that doesn't exist because an earlier step skipped, returned a different shape, or was wired to the wrong source.

---

## When to Use This Guide

- Creating a new agent definition
- Adding or reordering steps in an existing workflow
- Changing an `input_mapping` or `output_field`
- Changing what an LLM prompt returns
- Changing which action a step uses
- Connecting an audit finding to a handler agent

---

## Step 1: Schema Check

Before writing any SQL that references database tables:

```
\d table_name
```

Verify every column name, type, and constraint. Column names in workflow SQL are not validated at definition time — they fail at runtime.

Common mistakes:
- `slug` vs `url` (pages table has `url`, not `slug`)
- `spec_type` vs `aspect` (site_specs uses `aspect`)
- `spec_data` vs `data` (site_specs uses `data`)
- Missing `api_key_env_var` in `ai_service` config

---

## Step 2: Trace Every Data Path

For each step in the workflow, fill in this table:

| Step | Reads | Source step | Can source skip? | Writes | Skip shape |
|------|-------|------------|------------------|--------|------------|
| ensure_site_record | input_data.domain | message | No | site_record | N/A — errors on missing |
| load_page_record | site_record.site_id, input_data.page_name | ensure_site_record, message | No, Yes (page_name might be missing) | page_record | null (0 rows) |
| call_content_writer | current_page.sections (via input_mapping) | page_record | Yes (query found no rows) | page_content | {skipped: true, reason: "..."} |
| save_sections | page_content.response.page_body | call_content_writer | Yes (skipped) | sections_saved | {skipped: true, sections_saved: 0} |
| deploy_page | sections_saved.page_id | save_sections | Yes (skipped) | — | CRASH |

The last row reveals the bug before deployment.

### How to trace

**For each step, check what it reads:**

1. **Config literals** — values in `config` that are strings without dots (these are static, safe)
2. **Path-resolved values** — values in `config` or `input_mapping` with dots: `"site_record.site_id"`. Trace backward: which step wrote `site_record`? What's in `.site_id`?
3. **Params arrays** — `"params": ["site_record.site_id"]`. Same tracing needed.
4. **Template variables** — `{{.site_record.domain}}` in prompt templates. Same tracing.

**For each step, check what it writes:**

1. **output_field** — the key under which this step's result is stored in `collected_data`
2. **Result shape** — what keys does the action return? Check the Go source or a working log.

**For each step, check if it can produce no useful output:**

| Action type | Skip condition | Skip result |
|-------------|---------------|-------------|
| `query_database` (array) | 0 rows returned | `null` (not `[]`) |
| `query_database` (single_row) | 0 rows returned | `null` |
| `execute_llm_prompt` | Template render fails | `{type: "text", result: ""}` |
| `call_agent` | Target agent skips | `{response: {skipped: true, reason: "..."}}` |
| `call_agent` | Target agent errors | Goes to `error_step` if configured |
| `scrape_web` | URL unreachable | `{error: "...", pages_scraped: 0}` |
| `read_site_spec` | No specs exist | `{found: false, data: {}}` |
| `claim_work_item` | Already claimed | `{claimed: false}` |
| `conditional` | — | Writes nothing, only routes |

---

## Step 3: Guard Every Skip

If step N can skip, step N+1 must not blindly read N's output.

**Pattern: Conditional guard**

```json
{
    "check_has_content": {
        "action": "conditional",
        "config": {
            "condition": "page_content.response.skipped != true",
            "then_step": "save_sections",
            "else_step": "complete_error"
        }
    }
}
```

**Pattern: Error step fallback**

```json
{
    "load_page_record": {
        "action": "query_database",
        "next_step": "spawn_content_writer",
        "error_step": "spawn_content_writer"
    }
}
```

The `error_step` means: if this step fails (query error, no rows), continue anyway. The next step must handle the missing data.

**Pattern: Optional input_mapping fields**

```json
{
    "input_mapping": {
        "site_id": "site_record.site_id",
        "page_id?": "page_record.id",
        "sections?": "page_record.sections"
    }
}
```

The `?` suffix means the field is optional — the call proceeds without it if the path is missing.

---

## Step 4: Verify Action Config Support

Many actions have hardcoded paths that ignore config fields. Before using an action in a new context, check the Go source:

```
Does the action actually read config["my_custom_field"]?
Or does it hardcode looking at collected_data["site_plan"]["pages"]?
```

Common offenders:
- `sync_pages_to_db` — hardcodes looking for `site_plan.pages` or `page_plan.pages`, ignores custom `pages_field` config
- `write_build_items` — expects specific plan structure from the planner pipeline
- `save_page_sections` — expects `html_field` to point to actual HTML content

**Rule:** If the Go action doesn't support the config you need, write a new action rather than trying to reshape data to fit.

---

## Step 5: Check Input/Output Contracts

```sql
SELECT type, input_contract, output_contract
FROM agent_definitions WHERE type = 'my-agent';
```

If calling this agent from another agent's workflow, the caller's `input_mapping` must produce what the callee's `input_contract.required` expects.

Trace the mapping:

```
Caller's input_mapping:
    "site_id": "site_record.site_id"        → resolves to UUID ✓
    "current_page": "input_data.current_page" → resolves to audit spec ✗ (missing sections)

Callee expects (page-content-writer):
    required: current_page with sections array
```

---

## Step 6: Check Template Variables Against Data

For `execute_llm_prompt` steps, every `{{.field}}` in the prompt must exist in the template data at render time.

```
Template: {{.site_specs.specs.content_direction.voice}}

Trace: site_specs comes from read_site_spec output
  → read_site_spec loads all current aspects
  → content_direction aspect exists? Check site_specs table.
  → voice key exists in the content_direction data? Check the data.
```

**Go template failure modes:**
- `{{.missing_field}}` → empty string (silent)
- `{{.missing_field.nested}}` → template execution error → empty result
- `{{range .nil_array}}` → template execution error → empty result
- `{{if .field}}{{.field.nested}}{{end}}` → safe, skips if nil

**Rule:** Always wrap deep paths in `{{if}}` blocks. Always wrap `{{range}}` in `{{if}}` checks.

---

## Step 7: Verify the Handoff Between Audit Finding and Handler

When an audit agent creates a work item, the handler receives `spec` as its working context. The handler's workflow reads paths from this spec.

**Check:** What does the handler need vs what does the audit write?

| Handler needs | Where it reads from | Audit provides? |
|---------------|-------------------|-----------------|
| page sections | current_page.sections | No — audit spec has page_name only |
| page_id | current_page.id or page_record.id | No — audit spec doesn't include IDs |
| site_id | input_data.site_id | Yes — dispatch loop passes this |
| domain | input_data.domain | Yes — dispatch loop passes this |

**Fix pattern:** Handler loads missing data from DB:

```json
{
    "load_page_record": {
        "action": "query_database",
        "config": {
            "query": "SELECT id, name, sections::text, page_type FROM pages WHERE site_id = $1 AND name = $2",
            "params": ["site_record.site_id", "input_data.page_name"]
        },
        "output_field": "page_record"
    }
}
```

**Rule:** Handlers should be self-sufficient. They receive identifiers (site_id, page_name, work_item_id) and load everything else from the DB. Never assume the work item spec contains structural data.

---

## Checklist Summary

Before deploying a workflow change:

- [ ] **Schema:** `\d table_name` for every table referenced in queries
- [ ] **Paths:** Every dot-path in config/input_mapping traced to its source step
- [ ] **Skips:** Every step's skip behaviour documented, downstream steps guarded
- [ ] **Actions:** Go source checked for hardcoded paths vs configurable paths
- [ ] **Contracts:** Caller's input_mapping produces what callee's input_contract requires
- [ ] **Templates:** Every `{{.field}}` exists in template data, deep paths wrapped in `{{if}}`
- [ ] **Handoffs:** Handlers load structural data from DB, don't rely on work item spec

---

## Common Data Shapes

Reference for tracing — what each action typically returns.

### ensure_site_record
```json
{"site_id": "uuid", "domain": "example.com", "status": "deployed",
 "content_data": {...}, "network_id": "uuid", "created": "timestamp"}
```

### read_site_spec (single aspect)
```json
{"found": true, "spec_id": "uuid", "aspect": "identity",
 "site_id": "uuid", "source": "classifier", "data": {...}}
```

### read_site_spec (all aspects)
```json
{"site_id": "uuid", "aspects": ["identity", "design_intent"],
 "aspect_count": 2, "specs": {"identity": {...}, "design_intent": {...}}}
```

### query_database (array, with rows)
```json
[{"name": "index", "title": "Home"}, {"name": "about", "title": "About"}]
```

### query_database (array, no rows)
```json
null
```

### query_database (single_row, with row)
```json
{"id": "uuid", "name": "index", "sections": "[\"hero\",\"features\"]"}
```

### query_database (single_row, no row)
```json
null
```

### execute_llm_prompt (success)
```json
{"type": "text", "result": "...parsed JSON or text..."}
```

### execute_llm_prompt (template failure or empty LLM response)
```json
{"type": "text", "result": ""}
```

### call_agent (success)
```json
{"response": {...agent output...}, "response_status": "complete",
 "response_received_at": "timestamp"}
```

### call_agent (target skipped)
```json
{"response": {"skipped": true, "reason": "no sections defined"},
 "response_status": "complete", "response_received_at": "timestamp"}
```

### claim_work_item (claimed)
```json
{"claimed": true, "work_item_id": "uuid", "claimed_by": "agent-id"}
```

### claim_work_item (already claimed)
```json
{"claimed": false, "reason": "already claimed by other-agent"}
```

### spawn_agent
```json
{"agent_id": "uuid", "agent_type": "page-content-writer",
 "requests_topic": "job.xxx.requests", "initialized": true,
 "role": "content_writer", "await_response": true}
```

### load_work_items (has items)
```json
{"has_items": true, "count": 5,
 "items": [{"id": "uuid", "item_type": "...", "handler_agent": "...", "spec": {...}}]}
```

### load_work_items (no items)
```json
{"has_items": false, "count": 0, "items": []}
```
