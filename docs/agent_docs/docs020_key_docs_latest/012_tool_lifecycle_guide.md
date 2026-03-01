# 012 — Tool Lifecycle Guide

How interactive tools are suggested, deployed, and improved across sites.

---

## Overview

Tools are self-contained interactive HTML components (calculators, converters, estimators, etc.) stored in `content_components` with `component_level = 'tool'`. The system manages them through three agents and a discovery check, connected by the standard work item pipeline.

The key design decision: **tool selection is an LLM judgment call, not a catalogue lookup.** The system doesn't just match tools from a library by keyword — it evaluates what would genuinely help a site's visitors given the industry, audience, and services. A gas wholesaler gets a unit converter; a photographer gets a booking calculator.

---

## Components

### Discovery Check: `missing_tools`

**File:** `discovery_checks/check_missing_tools.go`

A structural check, not an evaluator. It asks two questions:

1. Does this site have any tools deployed? (COUNT on `content_components` via `page_components`)
2. Has a tool evaluation happened in the last 7 days? (checks `site_work_items` for `item_type = 'evaluate_tools'`)

If both are no, it creates a single `evaluate_tools` work item with `handler_agent: tool-suggester`. It does not look at the library, does not try to match by affinity, and does not decide which tools are appropriate.

**Work item created:** `evaluate_tools` → `tool-suggester`

### Agent: `tool-suggester`

**Migration:** 062  
**Category:** analyst / specialist  
**Handles:** `evaluate_tools` work items  
**Input contract:** `site_id`, `domain`

Evaluates what tools would benefit a site using LLM judgment. Workflow steps:

1. **read_specs** — `read_site_spec` loads all aspects (identity, classification, brand_dna)
2. **load_pages** — `query_database` gets deployed/planned pages with names, slugs, purposes
3. **load_existing_tools** — `query_database` checks what tools are already on the site (avoids duplicates)
4. **load_library_tools** — `query_database` loads the library catalogue (up to 30 tools) for reference
5. **suggest_tools** — `execute_llm_prompt` evaluates what tools would help, considering industry and audience

The LLM prompt includes concrete examples per industry and explicitly instructs against irrelevant suggestions. It returns 2–5 suggestions, each with:

- `name` and `function` (kebab-case)
- `description` of what it does and why it helps
- `priority` (1–5)
- `target_page` — which existing page to add it to, or `new` for a dedicated tools page
- `library_source` — function name from library if forkable, or null if it needs building from scratch
- `complexity` — simple / moderate / complex

6. **create_items_loop** — loops over suggestions, creating an `add_tool` work item per suggestion with the full suggestion object in `spec_data`

**Work items created:** N × `add_tool` → `tool-deployer`

### Agent: `tool-deployer`

**Migration:** 061  
**Category:** executor / specialist  
**Handles:** `add_tool` work items  
**Input contract:** `site_id`

Deploys a tool from the library to a site using the fork-on-deploy model. Workflow steps:

1. **load_item** — `load_work_items` gets the next `add_tool` item for this site
2. **check_has_item** — conditional guard
3. **deploy_tool** — `deploy_tool_to_site` action does the heavy lifting:
    - Loads the library tool by `tool_component_id` from the work item spec
    - Checks if already deployed (fork exists for this site) — returns early if so
    - Forks the tool — INSERT into `content_components` with `forked_from` pointing to the library original
    - Creates a tool page at `/tools/{function}.html`
    - Creates a `page_component` linking fork to page
4. **complete_item** — marks work item complete

After completion, the improvement loop's next sweep picks up the new page via `needs_rerender` and deploys it through the normal render/git/deploy pipeline.

**Go action:** `deploy_tool_to_site` in `deploy_tool_action.go`

**Fork model:** The site owns its copy. Changes to the library tool do not cascade to existing forks. This means each site's tools can diverge independently — which is what tool-improver relies on.

### Agent: `tool-improver`

**Migration:** 062  
**Category:** specialist  
**Handles:** `improve_tool` work items  
**Input contract:** `site_id`, `component_id`, `issue`

Incrementally improves a deployed tool based on an issue description. Workflow steps:

1. **load_tool** — `query_database` loads the tool's current HTML, CSS, JS from `content_components`, plus its page context (slug, page_id, page_name)
2. **check_tool_found** — conditional guard, completes with `skipped` status if not found
3. **load_brand_context** — `read_site_spec` loads all aspects so improvements match site style
4. **improve_tool** — `execute_llm_prompt` rewrites the HTML to fix the specific issue. The prompt enforces: CSS variable usage (no hardcoded hex), mobile compatibility, self-contained output, no external dependencies
5. **update_component** — `update_component_html` action saves the improved HTML. Optionally snapshots the previous version to `component_versions` first
6. **create_rerender_item** — creates a `needs_rerender` work item so the page gets rebuilt and deployed

**Go action:** `update_component_html` in `update_component_html_action.go`

The action also marks associated `page_components` as `build_status = 'pending'` so the rerender pipeline picks them up.

**Work item created:** `needs_rerender` → `rerender-pages`

---

## Work Item Flow

```
Improvement sweep (scheduler, every 10 minutes)
│
├─ discovery check: missing_tools
│  └─ evaluate_tools           handler: tool-suggester
│     └─ add_tool (×N)         handler: tool-deployer
│        └─ [needs_rerender]   handler: rerender-pages
│
├─ future: check_tool_rendering
│  └─ improve_tool             handler: tool-improver
│     └─ needs_rerender        handler: rerender-pages
│
└─ manual / content-reviewer report
   └─ improve_tool             handler: tool-improver
      └─ needs_rerender        handler: rerender-pages
```

---

## Work Item Types

| item_type | handler_agent | spec fields | created by |
|---|---|---|---|
| `evaluate_tools` | `tool-suggester` | `industry`, `site_type` | discovery check (missing_tools) |
| `add_tool` | `tool-deployer` | `name`, `function`, `description`, `library_source`, `target_page`, `complexity` | tool-suggester |
| `improve_tool` | `tool-improver` | `component_id`, `issue`, `check` (optional) | manual or future discovery check |

---

## Go Actions

| Action | File | Registry Category | Purpose |
|---|---|---|---|
| `deploy_tool_to_site` | `deploy_tool_action.go` | site | Fork library tool, create page, link component |
| `update_component_html` | `update_component_html_action.go` | site | Update `html_template` with optional version snapshot |

Both need registry entries in `registry.go`:

```go
"deploy_tool_to_site": {
    Handler:     DeployToolToSiteAction,
    Category:    "site",
    Description: "Fork library tool and create tool page for a site",
    IsLocal:     true,
},
"update_component_html": {
    Handler:     UpdateComponentHTMLAction,
    Category:    "site",
    Description: "Update html_template of a content_component with optional version snapshot",
    IsLocal:     true,
},
```

---

## Content Components Model

Library tools live in `content_components` with:
- `component_level = 'tool'`
- `forked_from IS NULL` (these are the originals)
- Unique index on `function` within active, unforked tools

Site forks are also in `content_components` with:
- `forked_from = <library tool UUID>`
- `name` suffixed with domain slug (e.g. `tool-vat-calculator-gaswholesalers-co-uk`)
- Owned by the site — can diverge independently

Tool pages:
- `page_type = 'tool'`
- URL pattern: `/tools/{function}.html`
- Nav label follows pattern: `Tools / {display_name}`

---

## Limitations and Future Work

**Not yet implemented:**
- `check_tool_rendering` discovery check — would detect mobile rendering issues, JS errors, or broken layouts automatically. Currently `improve_tool` items are created manually.
- Tool generation from scratch — when `library_source` is null in a suggestion, there's no agent yet that generates new tool HTML. The tool-deployer only handles library forks. A `tool-generator` agent would fill this gap, taking the suggestion's `description` and `function` and producing HTML/JS via LLM.
- The `create_work_item` action reads `summary` as a config literal. A patch (adding `summary` to the Optional inputs list) allows it to resolve from collected_data, so the tool-suggester's loop can pass the tool name as summary. Without this patch, all `add_tool` items get a generic summary.

**Design decisions:**
- Tools are a post-build concern. The site planner doesn't think about tools — the first improvement sweep after build handles them. This keeps the initial build fast and tool decisions informed by the actual deployed site.
- Each site owns its tool forks. No cascading updates from library changes. This is intentional — a tool that's been improved for a specific site shouldn't be overwritten by a library update.
- The 7-day cooldown on `evaluate_tools` prevents repeated evaluation spam. If a site genuinely has no tools after evaluation, it won't be re-evaluated for a week.


kafka-scheduler (every 120s)
→ build-pipeline-trigger (orchestrate action)
→ design-discovery-agent (runs checks including missing_tools)
→ missing_tools check sees: zero tools deployed + no evaluation in 7 days
→ creates evaluate_tools work item, handler_agent = 'tool-suggester'
→ triage_detected_items promotes it to 'triaged'
→ site-work-orchestrator fix_items_loop
→ loads triaged items
→ sees handler_agent = 'tool-suggester'
→ spawn_agent + call_agent with site_id and domain
→ tool-suggester runs its workflow
→ creates N × add_tool items → tool-deployer


# Manual trigger would look like:
INSERT INTO site_work_items (
site_id, source, domain, item_type, severity, summary,
spec, priority, handler_agent, status, created_by, item_key
) VALUES (
'<site_uuid>', 'manual', 'build', 'improve_tool', 'medium',
'Fix mobile rendering on unit converter',
'{"component_id": "<component_uuid>", "issue": "Tool overflows on screens narrower than 400px, inputs stack but submit button is clipped"}'::jsonb,
60, 'tool-improver', 'triaged', 'admin',
'improve_tool_<component_uuid_prefix>'
);