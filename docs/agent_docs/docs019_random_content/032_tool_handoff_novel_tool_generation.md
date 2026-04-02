# Handoff: Tool Pipeline — Novel Tool Generation

## Context

We're building a system that intelligently plans and builds multipage websites given domain names. Sites get interactive tools (calculators, converters, games etc.) suggested and deployed automatically by an agent pipeline. The tool pipeline was recently overhauled to fix a bug where a password checker was being deployed to every site (including gas wholesalers). The fix removed hard-coded tag matching and replaced it with LLM-based evaluation.

Transcript of prior work: `/mnt/transcripts/2026-03-28-18-44-42-tool-pipeline-llm-logging-finance-tools.txt`

Key docs in project knowledge:
- `012b_tool_lifecycle_guide_v2.md` — full pipeline: agents, work items, discovery checks, quality tiers
- `010b_tool_library_guide_v2.md` — library inventory (15 tools), deployment, templates, quality standards
- `001g_development_guide_new_agents_v7.md` — agent creation guidelines, bug history, rules
- `003e_contracts_and_standards_v5.md` — workflow contracts, action specs
- `002d_system_architecture_v4.md` — overall system architecture

---

## Current State (deployed and working)

### SQL — all in database
- **15 library tools** (14 with HTML, clip-path-builder still empty): 8 original + 7 finance calculators
- **tool-suggester** — fixed: parameterised queries, `.specs.` paths, Sonnet model, `ensure_site_record` step, `save_tool_spec` step writes evaluation to `site_specs` aspect `tools`, prompt includes `rejected_tools` in output
- **tool-deployer** — handles library forks only (novel tools fail — see "Next Task")
- **tool-improver** — status `active`, fixes tools based on issue descriptions
- **tool-auditor** — registered, LLM code review agent (Tier 2 quality)
- **tool_health** — registered in design-discovery-agent checks list
- **Finance calculators**: stamp duty, affordability, repayment, overpayment, bridging, buy-to-let, equity release

### Go — deployed in current build
- `check_missing_tools.go` — removed `universalTags`/`matchToolToSite`, creates `evaluate_tools` items for LLM evaluation
- `check_tool_health.go` — Tier 1 structural checks + Tier 2 audit queue
- `deploy_tool_action.go` — forks library tools, creates tool page with content sections (`hero-tool`, `tool-guide-intro`, tool widget at position 2, `tool-cta`), creates companion guide page at `/guides/{name}-guide.html`, creates `needs_content_page` work items for both
- `create_tool_component_action.go` — fixed pages INSERT columns
- `tool_admin_handlers.go` — GET/DELETE/POST sites/:id/tools, GET tools/library
- `confirm_work_item_handler.go` — POST work-items/:id/confirm (HITL → improve_tool)
- `get_pages_to_build_actions.go` — page_name fix

### Test in progress
Gaswholesalers (`5fe15466-4e2e-4ff2-981e-98c1b7074002`) has an `evaluate_tools` item at `triaged`. Tool-suggester previously suggested "Fleet Fuel Cost Estimator" (a novel tool — no library match) which failed at tool-deployer because `tool_component_id` was null. The password tool was removed. Re-evaluation was triggered with the fixed prompt and spec-write step.

Check status:
```sql
-- Did tool-suggester run?
SELECT id, status, error, completed_at
FROM site_work_items 
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002' 
  AND item_type = 'evaluate_tools'
ORDER BY created_at DESC LIMIT 1;

-- Did it write the spec?
SELECT jsonb_pretty(data) 
FROM site_specs 
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002' 
  AND aspect = 'tools' AND is_current = true;

-- Any new add_tool items?
SELECT summary, status, spec->>'tool_component_id' as tool_id, spec->>'library_source' as lib_src
FROM site_work_items 
WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002' 
  AND item_type = 'add_tool'
ORDER BY created_at DESC;
```

---

## Next Task: Novel Tool Generation

### Problem
When tool-suggester evaluates a site and no library tool fits, it suggests a novel tool with `tool_component_id: null` and `library_source: null`. Tool-deployer receives the `add_tool` work item but fails immediately because it requires `tool_component_id` to fork from.

Example: gaswholesalers gets "Fleet Fuel Cost Estimator" suggested. No library tool exists. The spec includes a good description of what the tool should do. Tool-deployer can't handle it.

### What needs to happen
A novel tool needs to be created from scratch via LLM, then deployed. Two approaches:

**Option A: Route to a separate `tool-generator` agent**
- tool-suggester's loop checks `tool_component_id`: if non-null → `add_tool` with handler `tool-deployer`; if null → `add_tool` with handler `tool-generator`
- `tool-generator` takes the description, generates HTML/CSS/JS via LLM, saves to `content_components`, then deploys (same as deployer but creates the tool first)
- The agent definition placeholder already exists in docs but no SQL/Go is written

**Option B: Expand tool-deployer to handle both cases**
- If `tool_component_id` is present → fork (existing path)
- If `tool_component_id` is null → generate via LLM, save as new library tool, then fork

**Recommendation from prior discussion**: Option A (separate agent, cleaner logs, separate responsibilities). The routing change is in tool-suggester's `create_items_loop` sub-workflow.

### What the generator needs to produce
A self-contained HTML template following the established pattern:
```html
<style>
    /* CSS with var(--color-*) and fallbacks, @media breakpoints */
</style>
<main class="container">
    <!-- Tool UI -->
</main>
<script>
(function() {
    // IIFE, all functions scoped, no external deps
})();
</script>
```

The LLM prompt should include:
- The tool description from the work item spec
- Site context (industry, audience, brand DNA)
- Template structure rules (CSS variables, self-contained, responsive, accessible)
- Example of a working tool (e.g. the stamp duty calculator template) as context for quality

### Key files to look at
- `deploy_tool_action.go` — current fork-based deployment (the pattern to follow for the deploy step after generation)
- `create_tool_component_action.go` — creates a new content_component + page + page_component (already exists but needs the LLM generation step before it)
- `tool-suggester` agent definition — the loop that creates `add_tool` items needs routing logic
- `tool-improver` agent definition — its LLM prompt pattern for HTML generation is a good reference for the generator prompt
- Finance calculator templates in `finance_tools_batch1/2/3.sql` — examples of well-structured self-contained tools

### Database schema for tools
```
content_components:
  id, name, display_name, function, category, component_level='tool',
  render_mode, is_dark_section, is_active, description,
  semantic_tags (jsonb), html_template (text), input_schema (jsonb),
  forked_from (uuid, null for library originals)

Unique index: idx_cc_tool_function_unique 
  ON (function) WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true
```

### Kubernetes
```
namespace: ai-persona-system
kafka: -n kafka (personae-kafka-cluster-combined-pool-prod-0/1/2)
```

---

## Architecture Reminders

- Every agent is an orchestrator
- Workflows are defined in SQL (agent_definitions.default_config)
- Complex logic goes in Go actions, workflows stay simple
- Don't create sub-workflows — spawn sub-agents instead
- Agents respond to their parent's response topic, not their own
- Use `logger.Info` not `logger.Debug` (Debug doesn't show in logs)
- Always check database schemas before writing SQL
- Reuse existing functions/structs before creating new ones
- Keep workflow variable names in sync with what actions expect
- ON CONFLICT for tools: `WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true`

---

## Three-Tier Tool Quality

| Tier | Component | Status |
|------|-----------|--------|
| 1 — Structural | `tool_health` check (Go regex) | Deployed |
| 2 — LLM review | `tool-auditor` agent (Sonnet) | Deployed |
| 3 — Visual | `tool-visual-tester` (Puppeteer, separate pod) | Planned |

The tool-auditor prompt reviews full HTML for JS logic bugs, mobile issues, UX problems, accessibility gaps. Findings with confidence `certain`/`likely` → `improve_tool`. Findings with confidence `possible` → `needs_human_review` → admin confirms or dismisses.

---

## Files in outputs (latest versions)

| File | Purpose |
|------|---------|
| `check_missing_tools.go` | Discovery check — evaluate_tools trigger |
| `check_tool_health.go` | Structural + audit queue |
| `deploy_tool_action.go` | Fork + content sections + companion guide |
| `create_tool_component_action.go` | Novel tool creation (needs LLM step added) |
| `tool_admin_handlers.go` | Admin REST endpoints |
| `confirm_work_item_handler.go` | HITL confirm endpoint |
| `063_tool_auditor_agent.sql` | Auditor agent definition |
| `finance_tools_batch1/2/3.sql` | 7 finance calculators |
| `012_tool_lifecycle_guide.md` | Full pipeline documentation |
| `010_tool_library_guide.md` | Library inventory + deployment guide |

---

## Other planned work (not started)
- **Shared tool scripts** — `tool_shared_scripts` table so finance calculators share `calculators.js` (deferred for simplicity)
- **Price data visualisation** — chart tools fetching `/data/prices.json` from a separate price collection workflow
- **Game variants** — give LLM a working game and request theme swaps, each variant is its own library tool
- **Headless browser testing** (Tier 3) — Puppeteer pod for visual regression
