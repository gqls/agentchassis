# Adoption Pipeline — Handoff Document v4

Session dates: 2026-04-02, 2026-04-03, 2026-04-06

---

## What Was Fixed Across These Sessions

### 1. Registry fix: select_representative_content
Action existed in Go but was never registered in `registry.go`. Added. This unblocked the content direction pipeline.

### 2. Component discovery pipeline unblocked (two bugs in CreateNeedsNewComponentItem)
- **Bug 1:** Column `domain` doesn't exist on `site_work_items` — should be `pipeline`
- **Bug 2:** Missing `::jsonb` cast on spec parameter
Both fixed and deployed. Component-creator agent has now processed real items. Templates generated: tool-list, guide-list, game-list (gamedesign.uk), provocation-card, lobby-grid, brief-explanation, platform-comparison (vonc.com).

### 3. Claimed-item-timeout improved
Two-phase: 15-min evidence-based auto-complete (checks page deployed, CSS generated), 40-min blind reset. Prevents wasteful rebuilds from response-routing loss.

### 4. Dispatch loop reaper threshold
30-minute idle threshold for `build-dispatch-loop` specifically, 90 minutes for everything else. Applied via SQL update to `stale-orchestration-reaper` pre_query.

### 5. Health check task target
`ai-endpoint-health-check` was targeting `build-dispatch-loop` instead of `endpoint-health-checker`. Fixed.

### 6. Stale model references
`claude-sonnet-4-5` → `claude-sonnet-4-6` in ai_service configs.

### 7. Site archetype classification step added to adoption workflow
New `classify_archetype` step added between `analyze_site` and `select_content`. SQL applied, verified:
```
analyze_site → classify_archetype → select_content → derive_content_direction → apply_plan
```
**Go change still needed:** Insert archetype extraction block in `apply_adoption_plan_action.go` at line 219. Patch file: `patch_apply_adoption_plan_archetype.go`. Without this, the LLM produces the archetype but `apply_plan` doesn't write it as a spec.

### 8. Tool recreation handler agent (2026-04-06)
New `tool-recreation-handler` agent created and deployed. Routes interactive pages away from `page-build-handler` to a two-stage LLM pipeline that analyses then recreates working tools.

**Go change deployed:** `apply_adoption_plan_action.go` work item creation loop now routes pages with `interactive_features` to `tool-recreation-handler` instead of `page-build-handler`:
- `item_type: "needs_tool_recreation"` (was `needs_content_page`)
- `handler_agent: "tool-recreation-handler"` (was `page-build-handler`)
- `mode` stays `"recreate"` (so `load_existing_content` still works)

**New Go action deployed:** `check_tool_completeness` — validates LLM output isn't truncated by checking for a completion marker and balanced script/style tags. Registered in `registry.go`.

### 9. agent_definitions schema correction (2026-04-06)
First INSERT attempt failed — handoff doc had wrong column names. Corrected mapping documented below in Key Schema Notes.

---

## Current Adoption Workflow (11 steps)

```
ensure_site_record → crawl_site → format_crawl → check_crawl_content
  → analyze_site (LLM: structure, identity, design, pages, sections)
  → classify_archetype (LLM: character, purpose, constraints, design patterns)
  → select_content (Go: pick representative pages)
  → derive_content_direction (LLM: writing style guide, ~17K chars)
  → apply_plan (Go: write specs, create pages, create work items)
  → complete
```

After `apply_plan`, the dispatch loop picks up work items:
- Static pages → `page-build-handler` (item_type: `needs_content_page`)
- Interactive pages → `tool-recreation-handler` (item_type: `needs_tool_recreation`)
- Design → `webdesign-agent` (item_type: `needs_design`)
- Rerender → `rerender-pages` (item_type: `needs_rerender`)

---

## Tool Recreation Handler

New agent for recreating interactive tools/games from crawled source code. Two-stage LLM pipeline.

### Workflow
```
ensure_site_record → load_page_record → check_page_found
  → load_site_specs → load_existing_content → load_related_context
  → analyze_tool (Sonnet: functional spec from source + context)
  → recreate_tool (Opus: produce working HTML/CSS/JS, 64K max tokens)
  → check_completeness (Go: verify not truncated, strip marker)
  → validate_tool → save_page_sections → update_page_status
  → spawn_rerender → deploy_page → complete
```

### Stage 1: analyze_tool (Sonnet)
Reads rawHtml source, site archetype, content direction, related pages. Produces a JSON functional spec covering: purpose, target user, interaction flow, technical spec (algorithms, state, events, rendering), visual spec, data model (inputs/outputs/state), edge cases, dependencies, site context, improvement notes.

### Stage 2: recreate_tool (Opus, 64K tokens)
Receives the functional spec + original source as reference. Produces a complete self-contained HTML page with embedded CSS and JS. Must include completion marker `<!-- tool-recreation-complete -->`. Output is inner page content (no html/head/body tags) — site chrome is injected by the deployment system.

### Completeness check
Go action `check_tool_completeness` verifies:
- Completion marker present
- Balanced `<script>` and `<style>` tags
- Output > 500 chars
Non-blocking — logs warnings but continues, since a truncated tool may still be partially functional.

### Design decisions
- **Why two stages?** Separates reasoning (understanding the tool) from coding (producing the tool). The functional spec acts as a contract between the two, and can be reviewed independently.
- **Why Opus for recreation?** Code generation quality matters more than speed here. Sonnet produced broken tools in earlier attempts.
- **Why not direct deploy?** Original approach considered but rejected. LLM recreation produces maintainable code that fits our component system, can be improved by the improvement loop, and doesn't carry baggage (broken external refs, CMS deps, analytics scripts).
- **Template `{{ }}` in rawHtml is safe.** Go's `text/template` doesn't re-parse substituted data values — only the template string is parsed for directives.

---

## The Core Problem: Adoption Produces Generic Sites

gamedesign.uk is a dark-themed developer utility platform with working JavaScript tools (Drop Rate Simulator, Progression Architect), playable games (Jelly Invaders, A* Pathfinding), and technical guides. The adoption pipeline correctly captures identity, design palette, interactive features, and content direction. But the build pipeline then:

1. Ignores crawled CSS and applies a generic style collection
2. Feeds content to the LLM content writer, which can't reproduce JavaScript applications
3. Uses generic brochure components (hero → features → CTA) for pages that had custom layouts
4. The improvement loop then audits against generic standards and makes it worse

**The site has been restored to its original state and is ready for re-crawling.**

**Mitigation deployed (2026-04-06):** Interactive pages now route to `tool-recreation-handler` which uses a code-generation-focused prompt with Opus. Static content pages still go through `page-build-handler`. Problems 1, 3, and 4 remain for static pages.

---

## Site Archetype Classification

Designed and tested this session. The LLM produces a multi-dimensional classification:

- **label**: human-readable 2-4 word description ("Game Dev Utility Hub")
- **character**: feel, polish, budget impression, age impression, commercial intent, density
- **design**: palette mood, layout approach, typography feel, imagery, animation, responsive
- **content**: primary/secondary types, voice, media
- **purpose**: array of what the site exists to do
- **content_model**: array of content types present
- **interaction_patterns**: array of what users DO
- **revenue_model**: none/advertising/affiliate/e-commerce/etc
- **visual_character**: array of style tags
- **audience**: array of who it's for
- **structure**: index layout, listing style, navigation, page depth
- **constraints**: array of things the improvement loop should NEVER do

Tested against gamedesign.uk — output was accurate and actionable. The prompt is in the `classify_archetype` step config in the adoption agent definition.

The archetype is a snapshot, not an aspiration. It describes what the site IS. The strategist determines what it should become.

---

## What Needs Building Next

### Priority 1: Deploy the Go archetype patch
Insert the archetype extraction block in `apply_adoption_plan_action.go` at line 219. Without this, `classify_archetype` runs but the output isn't persisted as a spec. Patch in `patch_apply_adoption_plan_archetype.go`.

### Priority 2: Re-crawl gamedesign.uk with rawHtml capture
The previous crawl did NOT capture rawHtml — only markdown. The `has_raw_html` column is false for all adoption_page records. The firecrawl config already requests `["markdown", "rawHtml"]` in `scrape_config.formats`. The original site is now restored and ready. Old adoption data must be cleaned before re-triggering (see test script).

### Priority 3: Design fingerprint extraction
Extract actual CSS values from crawled rawHtml — hex colours, font families, CSS variables, spacing values. Store as `site_specs` aspect `design_fingerprint`. The webdesign-agent reads this instead of guessing from the vague design spec ("dark background", "bright accent"). This is the difference between reproducing the site's look and producing another generic dark theme.

### Priority 4: Iterate on tool recreation prompts
The tool-recreation-handler is deployed but untested against real crawled tools. First run will show whether Opus produces working code. Things to watch:
- Are calculations mathematically correct?
- Do event handlers bind properly?
- Does canvas/SVG rendering work?
- Is the output truncated (check_completeness will flag this)?

If results are poor, consider: adjusting the prompt structure, adding a third "review and fix" LLM pass, or trying a different model for specific tool types.

### Priority 5: Component recreation from functional specs
For components that can't be directly deployed (broken references, CMS dependencies), the adoption analysis should produce per-component functional specs. The component-creator builds from spec using our contracts and CSS system. The result works like the original but is maintainable by our agents.

### Priority 6: Adoption-aware improvement loop
- Design auditors compare against design_fingerprint, not generic standards
- Content auditors respect content_direction writing rules
- Pages built by tool-recreation-handler excluded from content rewrites
- Site archetype constraints checked before any suggestion is applied

---

## Database State: gamedesign.uk

```
Site ID: 15a6cb16-5a86-4541-a8e4-d7106239b6a4
Domain:  gamedesign.uk
Database: clients_db

Pre-cleanup state (2026-04-06):
  Pages:      13 deployed (5 main + 4 blog posts + 3 tool pages + bayesian-ranking-guide)
  Work items: 1388 rows (mostly improvement loop churn)
  Specs:      identity, design, content_direction, structure
              site_archetype NOT YET WRITTEN (Go patch not deployed)
  Research:   adoption_page records exist but NO rawHtml

The original site has been restored and is live for re-crawling.
Cleanup SQL provided in test_tool_recreation.sh — clears pages, specs,
work items, research results, and resets build_status to pending.
```

---

## Diagnosed But Not Applied: Coordinator Go Fixes

### Zombie dispatch loop pods — full root cause

Loop-expanded steps are lost from `workflow_plan` during concurrent state updates. When a later iteration fails, the timeout handler can't find the step, `failWorkflow` hits optimistic lock and can't save FAILED status. Pod stays alive.

**Identified fixes (not applied — too risky without dedicated testing):**
- `failWorkflow`: retry UpdateState with fresh version when fresh status is AWAITING_RESPONSES
- `handleRequestTimeout` line ~3039: check return value of `handleRecoverableError` (currently discarded)

**Applied mitigation:** 30-minute dispatch loop reaper threshold.

**Files reviewed:** coordinator.go, spawn_actions.go, loop_expansion_handler.go, loop_error_handler.go

---

## Key Schema Notes

- `site_work_items`: uses `pipeline` not `domain`, has `item_key` with partial unique index
- `sites`: no `deploy_status` column, uses `build_status`
- `scheduled_tasks`: uses `name` not `task_name`
- `orchestration_states`: uses `owner_agent_type`, workflow_plan key is lowercase `steps`
- `site_specs`: data column is `data` not `spec_data`
- `agent_definitions` column mapping (doc name → actual):
  - `container_image` → `image_repository`
  - `resource_config` → `resources`
  - `topic_config` → `topics`
  - `delegation_config` → `delegation_preferences`
  - `role` → `agent_category`
  - `tags` → `domain_tags`
  - `questionnaire` → `briefing_questionnaire`
  - `timeout_seconds` → `idle_timeout_seconds`
  - `image_tag` is correct
  - Has `agent_category` CHECK constraint: strategist, executor, analyst, integrator, coordinator, specialist
  - Has `status` CHECK constraint: active, experimental, deprecated, demo, template

---

## Kubernetes Reference

```bash
kubectl -n ai-persona-system get pods
kubectl -n kafka get pods
# Kafka cluster: personae-kafka-cluster-combined-pool-prod-{0,1,2}
# Database: clients_db (not personas)
```

Deployment: git push → GitHub Actions → Backblaze S3.

---

## Files Created This Session (2026-04-06)

| File | Purpose | Status |
|------|---------|--------|
| `check_tool_completeness_action.go` | Go action: verify LLM output completeness | Deployed |
| `registry_entry_check_tool_completeness.go` | Registry entry for above | Deployed |
| `patch_route_interactive_pages.go` | Go patch: route interactive pages to tool-recreation-handler | Deployed |
| `tool_recreation_handler_agent_v2.sql` | Agent definition (fixed column names) | Ready to apply |
| `test_tool_recreation.sh` | Test script: cleanup, insert, trigger, verify | Reference |

---

## Guidelines Amendments

Detailed amendments documented in `028e_amendments.md` covering:
- 028e: workflow steps (7→11), content direction ownership, component discovery operational notes, phase plan updates, new principle (adopted state is the floor), dispatch loop/timeout notes
- 001g: column name `domain` → `pipeline`, new lesson on INSERT column drift and `::jsonb` cast pattern
- 028h (this session): agent_definitions column mapping, site_specs `data` not `spec_data`, database is `clients_db`, tool-recreation-handler architecture
