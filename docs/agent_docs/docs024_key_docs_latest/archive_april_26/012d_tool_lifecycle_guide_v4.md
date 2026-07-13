# 012 — Tool Lifecycle Guide

How interactive tools are suggested, deployed, improved, and quality-checked across sites.

---

## Overview

Tools are self-contained interactive HTML components (calculators, converters, estimators, etc.) stored in `content_components` with `component_level = 'tool'`. The system manages them through five agents and two discovery checks, connected by the standard work item pipeline.

**Key design decisions:**

1. **Tool selection is an LLM judgment call, not a catalogue lookup.** The system evaluates what would genuinely help a site's visitors given the industry, audience, and services. If no tools are appropriate, it suggests none. A gas wholesaler gets a unit converter; a photographer gets a booking calculator; neither gets a password checker.

2. **Tool quality is checked in three tiers.** Structural checks (fast, cheap) catch broken tools. LLM code review (Sonnet-tier) catches logic bugs, mobile issues, and UX problems. Headless browser testing (future) catches rendering issues the LLM can't see. The first two tiers are automated; the third is planned.

3. **Tool removal is a human decision.** The system can identify that a tool is broken and fix it, but whether a tool belongs on a site is an admin judgment. The admin dashboard provides removal controls.

---

## Components

### Discovery Check: `missing_tools`

**File:** `discovery_checks/check_missing_tools.go`

A structural check, not an evaluator. It asks two questions:

1. Does this site have any tools deployed? (COUNT on `content_components` via `page_components`)
2. Has a tool evaluation happened in the last 7 days? (checks `site_work_items` for `item_type = 'evaluate_tools'`)

If both are no, it creates a single `evaluate_tools` work item with `handler_agent: tool-suggester`. It does not look at the library, does not try to match by affinity, and does not decide which tools are appropriate.

**Work item created:** `evaluate_tools` → `tool-suggester`

### Discovery Check: `tool_health`

**File:** `discovery_checks/check_tool_health.go`

Runs on every improvement sweep for sites that have deployed tools. Two-tier check:

**Tier 1 — Structural checks (in Go, no LLM cost):**
- Page deployed — `build_status` is `deployed` or `active` (blocker if not)
- HTML present — `page_component.rendered_html` is non-empty (blocker)
- Template present — fork `html_template` is non-empty (blocker)
- Has `<script>` block — tool is interactive (error if missing)
- Has `<style>` block — tool has layout (warning)
- Has `@media` breakpoint — tool is mobile-responsive (warning)
- No hardcoded hex colours outside `var()` fallbacks (warning, >3 instances)
- No external `fetch()`, CDN references, or external `src=` (warning)

Structural blockers create `improve_tool` items directly — tool-improver can fix these without LLM review of the code.

**Tier 2 — LLM audit queue:**
Tools that pass structural checks (no blockers) are queued for LLM code review by creating `audit_tool` items for tool-auditor. 30-day cooldown per tool to avoid repeated audits.

**Work items created:**
- `improve_tool` → `tool-improver` (structural blockers)
- `audit_tool` → `tool-auditor` (LLM code review)

### Agent: `tool-suggester`

**Migration:** 062  
**Category:** specialist  
**Handles:** `evaluate_tools` work items  
**Input contract:** `site_id`, `domain`

Evaluates what tools would benefit a site using LLM judgment. Workflow steps:

1. **read_specs** — `read_site_spec` loads all aspects (identity, classification, brand_dna)
2. **load_pages** — `query_database` gets active pages with names, URLs, titles
3. **load_existing_tools** — `query_database` checks what tools are already on the site (avoids duplicates)
4. **load_library_tools** — `query_database` loads the library catalogue (up to 30 tools with templates) for reference
5. **suggest_tools** — `execute_llm_prompt` evaluates what tools would help, considering industry and audience

The LLM prompt includes concrete examples per industry and explicitly instructs against irrelevant suggestions. It can return 0-5 suggestions. Returning zero is correct when no tools are appropriate — the prompt includes examples of bad suggestions (e.g. "password checker for a gas wholesaler"). Each suggestion includes:

- `name` and `function` (kebab-case)
- `description` of what it does and why it helps
- `priority` (1–5)
- `target_page` — which existing page to add it to, or `new` for a dedicated tools page
- `library_source` — function name from library if forkable, or null if it needs building from scratch
- `complexity` — simple / moderate / complex
- `related_pages` — 1-3 existing page names whose topics connect naturally to the tool (for cross-linking)

6. **save_tool_spec** — `write_site_spec` persists the evaluation (reasoning, suggestions, rejections) to `site_specs` aspect `tools`
7. **create_items_loop** — loops over suggestions. Each suggestion is routed by a conditional check:
    - If `tool_component_id` is non-null → `add_tool` work item with `handler_agent: tool-deployer` (library fork)
    - If `tool_component_id` is null → `add_tool` work item with `handler_agent: tool-generator` (novel tool, LLM generation)
8. **create_cross_links** — `create_tool_cross_link_items` action creates `content_rewrite` work items for each suggestion's `related_pages`, threading rewrite guidance to the content writer

**Work items created:** N × `add_tool` → `tool-deployer` (library) or `tool-generator` (novel), plus N × `content_rewrite` → `page-build-handler` (cross-links)

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

### Agent: `tool-auditor`

**Migration:** 063  
**Category:** specialist  
**Handles:** `audit_tool` work items  
**Input contract:** `site_id`, `domain`, `component_id`

LLM-based code review of deployed tools. This is Tier 2 quality checking — the structural checks in `tool_health` (Tier 1) catch broken tools cheaply; tool-auditor catches logic bugs, mobile issues, and UX problems that require reasoning through the code.

Workflow steps:

1. **ensure_site_record** — standard site context loading
2. **load_tool** — `query_database` loads the tool's full HTML/CSS/JS, rendered HTML, page context
3. **load_site_context** — `read_site_spec` loads industry/audience context for relevant review
4. **llm_audit** — `execute_llm_prompt` sends the full source to Sonnet for code review

The LLM prompt checks six categories:
- **JS logic bugs** — uninitialised variables, missing event listeners, division by zero, DOM reference mismatches
- **Mobile & touch** — layout at 375px, touch targets ≥44px, clipped/overflowing elements, hover-only interactions
- **UX & interaction** — clear first action, feedback on interaction, visible labels, working copy/download
- **CSS & styling** — hardcoded colours, missing CSS variable usage, `!important` conflicts
- **Accessibility** — input labels, alt text, keyboard operability
- **Self-containment** — external APIs, CDN dependencies, ID collisions

Each finding has a confidence level:
- `certain` — the bug is traceable in the code → creates `improve_tool` item for automatic fix
- `likely` — strong evidence → creates `improve_tool` item
- `possible` — worth checking but uncertain → creates `needs_human_review` item for HITL

5. **create_items_loop** — loops over findings, routes to `improve_tool` or `needs_human_review` based on confidence

The LLM also returns a `quality_score` (1-10) which is stored in the work item result for tracking tool quality over time.

**Work items created:**
- `improve_tool` → `tool-improver` (certain/likely findings)
- `needs_human_review` → HITL (possible findings)

### Agent: `tool-generator`

**Migration:** 062b  
**Category:** specialist  
**Handles:** `add_tool` items where `tool_component_id` is null (no library tool to fork)  
**Input contract:** `site_id`, `domain`, `spec` (with function, name, description)

Creates a new tool from scratch via LLM when no suitable library tool exists. Routing from tool-suggester is automatic — the `check_is_library` conditional in the loop sub_workflow sends novel suggestions here.

Workflow steps:

1. **ensure_site_record** — standard site context
2. **load_brand_context** — `read_site_spec` loads brand/identity for matching site style
3. **generate_tool_html** — `execute_llm_prompt` generates the full HTML/CSS/JS
4. **save_tool** — `create_tool_component` action does the heavy lifting:
    - Strips markdown code fences from LLM output (via `datahelpers.StripCodeFences`)
    - Creates `content_components` row with `component_level = 'tool'`, `created_from = 'generated'`
    - Creates tool page at `/tools/{function}.html` with `page_type = 'tool'`
    - Creates `page_components` at position 2 with `rendered_html` set and `build_status = 'deployed'`
    - Adds "Tools" nav group and nav item (via `addToolToNav`)
    - Creates `needs_content_page` work item → `page-build-handler` writes hero, intro, and CTA sections around the tool widget
    - Creates companion guide page at `/guides/{function}-guide.html` + `needs_content_page` work item

The generated tool follows the same template structure as library tools: `<style>` + `<div class="tool-container">` + `<script>`, CSS variables for colours, self-contained with no external dependencies.

**Go action:** `create_tool_component` in `create_tool_component_action.go`

**Work items created:**
- `needs_content_page` → `page-build-handler` (tool page content)
- `needs_content_page` → `page-build-handler` (companion guide)

### Cross-Linking: `create_tool_cross_link_items`

**Migration:** 071  
**Go action:** `create_tool_cross_link_items` in `create_tool_cross_link_items.go`  
**Called by:** tool-suggester's `create_cross_links` step (after `create_items_loop`)

After tool-suggester creates `add_tool` work items, it runs a second step that creates `content_rewrite` work items so existing content pages reference the new tools contextually. The LLM already sees all pages during evaluation and returns `related_pages` per suggestion (1-3 page names whose topics connect to the tool).

The action:
1. Loads page map (name → ID) for the site
2. Iterates suggestions, extracts `related_pages` arrays
3. Skips pages starting with `tool-` (tool-to-tool cross-linking isn't useful)
4. Creates `content_rewrite` work items for `page-build-handler` with:
   - `page_name` in spec (required — dispatch loop maps `spec.page_name` → `input_data.page_name`)
   - `suggestion` field with natural-language guidance for weaving in a tool reference
   - Dedup via `item_key = tool_crosslink:{function}:{page}:{site}`
   - Priority 110

When the dispatch loop processes these items, page-build-handler passes `spec.suggestion` through as `rewrite_guidance` to the content writer. The writer's per-section LLM prompt includes a `## Rewrite Guidance` block that instructs it to integrate the tool reference naturally.

**Note on prompt location:** The content writer's per-section prompt is nested inside:
```
process_sections_loop → config → sub_workflow → steps → generate_content → config → prompt_template
```
Top-level `jsonb_each(default_config->'workflow'->'steps')` will NOT find it.

**Work items created:** N × `content_rewrite` → `page-build-handler`

---

## Work Item Flow

```
Improvement sweep (scheduler, every 10 minutes)
│
├─ discovery check: missing_tools
│  └─ evaluate_tools                handler: tool-suggester
│     ├─ add_tool (×N, library)     handler: tool-deployer
│     │  └─ needs_content_page      handler: page-build-handler
│     │  └─ [needs_rerender]        handler: rerender-pages
│     ├─ add_tool (×N, novel)       handler: tool-generator
│     │  └─ needs_content_page (×2) handler: page-build-handler (tool page + guide)
│     │  └─ [needs_rerender]        handler: rerender-pages
│     └─ content_rewrite (×N)       handler: page-build-handler (cross-links)
│        └─ [needs_rerender]        handler: rerender-pages
│
├─ discovery check: tool_health
│  ├─ improve_tool (structural)     handler: tool-improver
│  │  └─ needs_rerender             handler: rerender-pages
│  └─ audit_tool (LLM review)      handler: tool-auditor
│     ├─ improve_tool (fixable)     handler: tool-improver
│     │  └─ needs_rerender          handler: rerender-pages
│     └─ needs_human_review         handler: hitl-review
│
└─ admin dashboard
   ├─ deploy tool (library)         handler: tool-deployer
   ├─ remove tool                   direct (DELETE endpoint)
   └─ improve_tool (manual)         handler: tool-improver
      └─ needs_rerender             handler: rerender-pages
```

---

## Work Item Types

| item_type | handler_agent | spec fields | created by |
|---|---|---|---|
| `evaluate_tools` | `tool-suggester` | `check`, `reason` | discovery check (missing_tools) |
| `add_tool` | `tool-deployer` or `tool-generator` | `tool_component_id` (null for novel), `name`, `function`, `description`, `library_source`, `target_page`, `complexity` | tool-suggester |
| `improve_tool` | `tool-improver` | `component_id`, `issue`, `check` (optional), `fix_suggestion` (optional) | tool_health check, tool-auditor, or manual |
| `audit_tool` | `tool-auditor` | `component_id`, `check`, `page_id`, `page_name` | tool_health check |
| `needs_human_review` | hitl-review | `component_id`, `tool_function`, `issue`, `fix_suggestion`, `confidence`, `category` | tool-auditor (low-confidence findings) |
| `content_rewrite` | `page-build-handler` | `page_name`, `page_id`, `suggestion`, `tool_function`, `tool_display_name`, `source` | tool-suggester (cross-linking) |

---

## Go Actions

| Action | File | Purpose |
|---|---|---|
| `deploy_tool_to_site` | `deploy_tool_action.go` | Fork library tool, create page + content sections + companion guide, link component |
| `create_tool_component` | `create_tool_component_action.go` | Create novel tool from LLM-generated HTML, create page with rendered_html at position 2, companion guide, nav entry, content work items |
| `create_tool_cross_link_items` | `create_tool_cross_link_items.go` | Create `content_rewrite` items for related pages; filters tool pages; includes `page_name` in spec |
| `update_component_html` | `update_component_html_action.go` | Update `html_template` with optional version snapshot |

## Discovery Checks

| Check | File | Purpose |
|---|---|---|
| `missing_tools` | `check_missing_tools.go` | Trigger tool evaluation when site has no tools |
| `tool_health` | `check_tool_health.go` | Structural quality checks + queue LLM audit |

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

### Three-tier quality plan

| Tier | Agent/Check | What it catches | Cost | Status |
|---|---|---|---|---|
| 1 — Structural | `tool_health` discovery check | Broken: no HTML, no script, not deployed, empty template | Zero (Go regex) | Built |
| 2 — LLM review | `tool-auditor` agent | Logic bugs, mobile issues, UX problems, accessibility gaps | ~1 Sonnet call per tool | Built |
| 3 — Visual | `tool-visual-tester` agent (future) | Rendering failures, layout breakage at specific viewports, canvas issues | Headless browser pod | Planned |

Tier 3 requires its own container (Puppeteer/Playwright) running as a separate agent. It would receive a page URL and viewport list, render the page, take screenshots, and optionally run basic interaction scripts (click buttons, fill inputs, check for JS errors in console). The headless browser container should be isolated — it's a resource-heavy, potentially unsafe environment (rendering arbitrary HTML). Design notes:
- Separate Kubernetes deployment with its own resource limits (CPU/memory for Chromium)
- Receives work items of type `visual_test_tool` with `page_url` and `viewports`
- Returns screenshots (stored in S3) and pass/fail results per viewport
- tool-auditor or tool_health could create visual_test_tool items for tools that pass LLM review
- Screenshots could be sent to an LLM for visual interpretation (Tier 2.5)

### Not yet implemented
- **Tool visual testing** (Tier 3) — headless browser rendering checks. See three-tier plan above.
- **Tool removal automation** — currently admin-only via dashboard. No automated removal of inappropriate tools (by design — relevance is a human judgment).
- **News → tool linking** — when news article writing is added, pass deployed tool list to the writer's prompt. The `rewrite_guidance` pattern is established.

### Implemented (was previously listed as future)
- `tool_health` discovery check — structural quality checks + LLM audit queue. Replaces the old `check_tool_rendering` placeholder.
- `tool-generator` agent — creates tools from scratch via LLM when no library tool exists.
- **Novel tool routing** — tool-suggester's `create_items_loop` routes suggestions with `tool_component_id: null` to `tool-generator` (via `check_is_library` conditional in the loop sub_workflow). Library tools route to `tool-deployer`. Tested end-to-end with gaswholesalers.com: "Gas Unit Converter" generated, deployed, content written, companion guide created, nav entry added.
- `tool-auditor` agent — LLM code review of deployed tools.
- Admin tool management endpoints — list, deploy, remove tools via dashboard.
- `create_tool_component` action — full parity with `deploy_tool_to_site`: creates component, page, page_component with rendered_html at position 2, nav entry, `needs_content_page` work items for tool page and companion guide.
- **Cross-linking** — tool-suggester returns `related_pages` per suggestion. `create_tool_cross_link_items` action creates `content_rewrite` work items. `rewrite_guidance` threads through page-build-handler → page-content-writer so the LLM weaves tool references into existing page content. Verified: tool links land in deployed HTML.

### Design decisions
- Tools are a post-build concern. The site planner doesn't think about tools — the first improvement sweep after build handles them. This keeps the initial build fast and tool decisions informed by the actual deployed site.
- Each site owns its tool forks. No cascading updates from library changes. This is intentional — a tool that's been improved for a specific site shouldn't be overwritten by a library update.
- The 7-day cooldown on `evaluate_tools` prevents repeated evaluation spam. If a site genuinely has no tools after evaluation, it won't be re-evaluated for a week.
- The 30-day cooldown on `audit_tool` prevents repeated LLM audits of the same tool. If tool-improver fixes an issue, the next audit happens after 30 days.
- Tool relevance (whether a tool belongs on a site) is an admin decision, not an automated one. The system can identify broken tools and fix them, but removing a working tool requires human judgment.

### Bug history
- **Missing_tools discovery check used universal tags** — the `matchToolToSite` function classified `security`/`password`/`privacy` as universal, deploying a password checker to every site (including gas wholesalers). Fixed by removing tag-based matching entirely — the discovery check now creates `evaluate_tools` items for LLM evaluation instead of `add_tool` items for direct deployment.
- **tool-suggester SQL bugs** — queried `p.slug` (doesn't exist, should be `p.url`), `p.purpose` (doesn't exist), `p.sort_order` (should be `p.nav_order`). Used template interpolation instead of parameterised queries. site_specs paths missing `.specs.` prefix. All fixed in the agent definition.
- **create_tool_component_action.go** — INSERT referenced `pages.slug`, `pages.sort_order`, `pages.purpose` which don't exist. Fixed to use `url`, `nav_order`, `meta_description`.
- **addToolToNav used wrong column names** — `site_nav_groups` has `group_key`/`group_label`/`group_type`/`position`, not `label`/`sort_order`/`in_header`/`in_footer`. `site_nav_items` requires `site_id` and uses `position` not `sort_order`. The function was silently failing (best-effort, non-fatal) on every tool deployment. Fixed to match the actual schema and use `NOT EXISTS` guard instead of `ON CONFLICT` (no unique constraint on nav items).
- **created_from check constraint** — `content_components.chk_created_from_valid` allows `manual`, `generated`, `adopted`, `tool`, `forked`. Initial code used `'tool-generator'` which failed the constraint. Fixed to use `'generated'`.
- **Novel tools routed to tool-deployer** — tool-suggester hard-coded `handler_agent: "tool-deployer"` for all suggestions. Novel tools (null `tool_component_id`) failed at tool-deployer. Fixed by adding `check_is_library` conditional in the loop sub_workflow that routes to `tool-generator` when `tool_component_id` is null.
- **Cross-link spec missing page_name** — `build-dispatch-loop` maps `spec.page_name` but the cross-link spec only had `"page"`. Page-build-handler got empty page_name. Fixed by adding `"page_name": pageName` to spec.
- **072 prompt injection missed nested prompt** — DO block searched top-level workflow steps via `jsonb_each`, but page-content-writer's LLM prompt is nested inside `process_sections_loop → sub_workflow → generate_content`. Fixed by targeting the correct jsonb path directly.
- **content-quality-auditor check_empty_pages** — `HAVING NOT EXISTS (... WHERE lpc.page_id = p.id ...)` referenced `p.id` not in `GROUP BY p.name`. Fixed: `GROUP BY p.id, p.name`.


## Trigger Flows

### Tool suggestion flow

```
kafka-scheduler (every 120s)
→ build-pipeline-trigger (orchestrate action)
→ design-discovery-agent (runs checks including missing_tools)
→ missing_tools check sees: zero tools deployed + no evaluation in 7 days
→ creates evaluate_tools work item, handler_agent = 'tool-suggester'
→ triage_detected_items promotes it to 'triaged'
→ build-dispatch-loop
→ spawns tool-suggester with site_id and domain
→ tool-suggester runs its workflow (LLM evaluates, may return 0 suggestions)
→ save_tool_spec writes evaluation to site_specs aspect 'tools'
→ create_items_loop routes each suggestion:
  ├─ tool_component_id != null → add_tool item, handler: tool-deployer (fork)
  └─ tool_component_id == null → add_tool item, handler: tool-generator (LLM creates)
→ create_cross_links step: creates content_rewrite items for related pages
→ tool-deployer forks library tool, creates page + content sections + companion guide
→ tool-generator: LLM generates HTML → create_tool_component creates component, page,
  page_component (with rendered_html), nav entry, content work items, companion guide
→ page-build-handler writes hero/intro/CTA sections (and rewrite_guidance for cross-links)
→ rerender-pages → page-rerender → git → deployed
```

### Tool quality flow

```
kafka-scheduler (every 120s)
→ design-discovery-agent (runs checks including tool_health)
→ tool_health check:
  ├─ Tier 1: structural checks on each deployed tool
  │  → creates improve_tool items for blockers (tool-improver fixes directly)
  └─ Tier 2: queues audit_tool for tools passing structural checks
     → build-dispatch-loop spawns tool-auditor
     → tool-auditor sends full HTML to LLM for code review
     → creates improve_tool items (certain/likely) or needs_human_review (possible)
```

### Manual triggers

```sql
-- Trigger tool evaluation for a site
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
) VALUES (
    '<site_uuid>', 'admin', 'build', 'evaluate_tools', 'low',
    'Evaluate tool needs',
    '{"reason": "manual"}'::jsonb,
    130, 'tool-suggester', 'triaged', 'admin',
    'evaluate_tools:<site_uuid>'
);

-- Request improvement of a specific tool
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
) VALUES (
    '<site_uuid>', 'admin', 'build', 'improve_tool', 'medium',
    'Fix mobile rendering on unit converter',
    '{"component_id": "<component_uuid>", "issue": "Tool overflows on screens narrower than 400px, inputs stack but submit button is clipped"}'::jsonb,
    60, 'tool-improver', 'triaged', 'admin',
    'improve_tool_<component_uuid_prefix>'
);

-- Request LLM audit of a specific tool
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
) VALUES (
    '<site_uuid>', 'admin', 'build', 'audit_tool', 'low',
    'LLM code review: <tool display name>',
    '{"component_id": "<component_uuid>", "check": "manual"}'::jsonb,
    140, 'tool-auditor', 'triaged', 'admin',
    'audit_tool:<tool_function>:<site_uuid>'
);

-- Manually trigger novel tool generation (bypasses tool-suggester)
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
) VALUES (
    '<site_uuid>', 'admin', 'build', 'add_tool', 'low',
    '<Tool Display Name>',
    '{"name": "<Tool Display Name>", "function": "tool-<kebab-name>", "description": "<What it does and why>", "priority": 1, "target_page": "new", "library_source": null, "tool_component_id": null, "complexity": "moderate"}'::jsonb,
    120, 'tool-generator', 'triaged', 'admin',
    'add_tool_novel:tool-<kebab-name>:<site_uuid>'
);
```