# Handoff: Mission-Driven Sites, Component Selector, and vonc.com Build

Session date: 2026-04-02 through 2026-04-03

---

## What was built

### 1. Mission-Driven Site Support (Tier 3 submissions)

The domain-submitter now accepts structured mission data alongside domain submissions. Three tiers:

- **Tier 1** — Domain only (existing, unchanged). Classifier infers everything.
- **Tier 2** — Domain + objective. Classifier uses objective as hint.
- **Tier 3** — Domain + mission + roadmap + briefs. Pre-defined product vision drives classification, planning, and component creation.

**Data separation principle:** Structured JSON (`mission`, `roadmap`) for machine consumers (content writers, plan_sections). Plain text (`mission_brief`, `roadmap_brief`) for LLM prompts. Briefs wrapped in `{"text": "..."}` because write_site_spec requires JSON objects.

**Domain-submitter workflow:**
```
ensure_site_record → store_contact_info → store_submission_spec
  → persist_mission → persist_mission_brief → persist_roadmap → persist_roadmap_brief
  → create_research_item → complete
```
Persist steps have `error_step` inside `config` (not at step level). Tier 1/2 submissions cascade through error_steps to `create_research_item`.

**Status:** Deployed and tested. vonc.com Tier 3 submission completed. All 5 aspects persisted (submission, mission, mission_brief, roadmap, roadmap_brief).

### 2. Classifier reads mission_brief from site_specs

Added `read_site_specs` step before `classify_and_extract`. The prompt references `{{.site_specs.specs.mission_brief.text}}` and `{{.site_specs.specs.roadmap_brief.text}}` via `{{if}}` guards. Sites without missions render empty — classifier works from research alone.

Added `interactive-platform` to site_type enum and `game-like` to tone_suggestion enum.

**Status:** Deployed. vonc.com classified as `interactive-platform` with confidence 0.97.

### 3. Planner reads roadmap_brief

Planner prompt includes roadmap_brief with strong instruction: "IMPORTANT — ROADMAP OVERRIDES THE COMPONENT LIST." When a roadmap is present, the planner outputs the roadmap's pages and section_types rather than standard brochure components.

**Status:** Deployed. vonc.com planner output section_types from roadmap (provocation-card, lobby-grid, gauntlet-interface, etc.).

### 4. Component Selector

**Go code:** `component_selector.go` — `SelectComponentByType()` queries content_components by `section_type` with scoring (35% site_type, 15% page_type, 30% quality, 10% specificity, 10% usage). `CreateNeedsNewComponentItem()` creates work items for missing section types using check-first dedup (not ON CONFLICT).

**Schema migration:** 8 new columns on content_components (`section_type`, `suitable_site_types`, `suitable_page_types`, `content_shape`, `visual_density`, `usage_count`, `avg_quality_score`, `created_from`). All 54 existing components backfilled with metadata. Hero consolidation: page-specific heroes share `section_type = 'hero'` with `suitable_page_types` differentiating them.

**plan_sections integration:** 3-path resolution per section:
- Path 1: Function match (existing sites, direct name lookup)
- Path 2: Selector match (finds component by section_type + scoring)
- Path 3: Not found → creates `needs_new_component` work item, defers section

**Status:** Deployed and working. vonc.com pages triggered Path 3, created 11 needs_new_component items, all completed by component-creator.

### 5. Component Creator Agent

Agent type: `component-creator`. Workflow: `generate_template (execute_llm_prompt) → store_component (store_generated_component) → complete`. Uses claude-sonnet-4-5.

Prompt references `{{.input_data.spec.section_type}}` etc. (spec fields from dispatch loop's input_mapping).

**Status:** Deployed. Generated 11 Spark components (provocation-card, lobby-grid, gauntlet-cta, brief-explanation, provocation-feed, gauntlet-interface, archetype-result-card, platform-comparison, game-master-explanation, archetype-grid, archetype-combinations). Also generated components for gamedesign.uk and robot-hands.com.

---

## Current state — vonc.com

### What's working
- Site record: `active`, site_id `e1e22a7d-0552-405a-85b3-1b1e51384df5`
- 12 site_specs aspects (submission, mission, mission_brief, roadmap, roadmap_brief, identity, classification, content_direction, design_intent, briefing, strategy, site_plan)
- Classification: `interactive-platform`, confidence 0.97, tone `game-like`
- Site plan: 6 pages with roadmap section_types (index, provocations, gauntlet, about, archetypes, contact)
- 11 Spark components generated in content_components library
- Content pages being processed (some complete, some in queue)

### What's broken — BLOCKING DEPLOYMENT
**All 22 generated components have corrupted html_template data.** The `html_template` column contains the raw JSON blob `{"function": "...", "html_template": "<style>..."}` instead of just the HTML. PostgreSQL `::jsonb` extraction fails because the LLM output contains unescaped quotes in SVG paths.

**Root cause:** The `store_generated_component` action's `parseGeneratedTemplate` function received LLM output wrapped in markdown code blocks (` ```json ... ``` `). `json.Unmarshal` failed on the backtick prefix. The fallback treated the entire string as raw HTML and stored it verbatim — including the JSON wrapper.

**Fix required (Go deploy):** The updated `store_generated_component_action.go` includes `stripCodeBlocks()` which removes markdown wrappers before JSON parsing. This file is in `/mnt/user-data/outputs/store_generated_component_action.go`.

### Recovery sequence after deploying the Go fix

```sql
-- 1. Delete all broken generated components
DELETE FROM content_components
WHERE created_from = 'generated'
  AND html_template LIKE '{%"html_template"%';

-- 2. Reset needs_new_component items to regenerate
UPDATE site_work_items
SET status = 'triaged', error = NULL, result = '{}'::jsonb,
    attempt_count = 0, claimed_by = NULL, claimed_at = NULL, completed_at = NULL
WHERE item_type = 'needs_new_component'
  AND status = 'complete';

-- 3. After components regenerate and verify clean templates:
SELECT function, LEFT(html_template, 60) as preview
FROM content_components WHERE created_from = 'generated';
-- Should start with <style> not {

-- 4. Clear vonc page_components and reset content pages
DELETE FROM page_components
WHERE page_id IN (SELECT id FROM pages WHERE site_id = (SELECT id FROM sites WHERE domain = 'vonc.com'));

UPDATE site_work_items
SET status = 'triaged', error = NULL, result = '{}'::jsonb,
    attempt_count = 0, claimed_by = NULL, claimed_at = NULL, completed_at = NULL
WHERE site_id = (SELECT id FROM sites WHERE domain = 'vonc.com')
  AND item_type = 'needs_content_page';

-- 5. Delete rerender items (let pipeline recreate)
DELETE FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'vonc.com')
  AND item_type IN ('needs_rerender', 'page_rerender');

-- 6. Remove extra pages not in roadmap
DELETE FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'vonc.com')
  AND name NOT IN ('index', 'provocations', 'gauntlet', 'about', 'archetypes', 'contact');
```

---

## Other fixes applied during this session

### error_step placement (all agents)
`error_step` must be inside `step.Config`, not at the step level. Fixed on: domain-submitter (4 persist steps), classifier (write_content_direction_spec, write_design_intent_spec).

### chk_created_from_valid constraint
Added `'tool'` and `'forked'` to allowed values. Was blocking the tool-generator.

### Improvement-sweep scheduler query
`wi.domain` renamed to `wi.pipeline` in the pre_query.

### Missing Kafka topic
`system.agent.build-dispatch-loop.process` needs recreation:
```bash
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --topic system.agent.build-dispatch-loop.process \
  --partitions 1 \
  --replication-factor 1 \
  --config retention.ms=3600000
```

---

## Go files to deploy

All in `/mnt/user-data/outputs/`:

| File | Changes | Priority |
|---|---|---|
| `component_selector.go` | `pipeline` not `domain` in INSERT. Check-first dedup instead of ON CONFLICT. | Deployed |
| `store_generated_component_action.go` | `stripCodeBlocks()` before JSON parse. `truncate()` helper for logging. | **NEEDS DEPLOY** |
| `plan_sections_action.go` | 3-path selector integration. Already deployed. | Deployed |

### store_generated_component_action.go — changes since last deploy
- Added `stripCodeBlocks()` function — strips ` ```json ... ``` ` wrappers
- Added `truncate()` helper for logging
- `parseGeneratedTemplate` calls `stripCodeBlocks` on string results before `json.Unmarshal`
- Better logging: includes length and first 50 chars when falling back to raw HTML

---

## Design decisions documented

**Pages always render with what they have.** They are never blocked waiting for components. Deferred sections use fallback components. The improvement loop discovers when better components become available and upgrades sections. This matches the framework philosophy of continuous evolution — sites get better over time, not stuck waiting for dependencies.

**Discovery agent pattern for component upgrades:**
```sql
-- Find sections using fallback where a matching section_type now exists
SELECT p.name, pc.slot_name, cc.function as current, better.function as available
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
JOIN content_components cc ON pc.component_id = cc.id
JOIN content_components better ON better.section_type = pc.slot_name
  AND better.is_active = true AND better.function != cc.function
WHERE cc.function IN ('generic-text-block', 'generic-hero')
  AND p.site_id = $site_id;
```
Not implemented as an agent yet — identified as the right approach for the improvement loop.

**Dispatch loop stays generic.** Handlers read their data from `input_data.spec.*`. No handler-specific field promotions in the dispatch loop's input_mapping.

**Briefs as `{"text": "..."}` objects.** write_site_spec requires JSON objects. Prompts access via `.text` field. The `{{if .site_specs.specs.mission_brief}}` guard references the object (correct), the render uses `.text` (correct).

---

## Monitoring queries

```sql
-- vonc.com overall status
SELECT item_type, wi.status, COUNT(*)
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE s.domain = 'vonc.com'
GROUP BY item_type, wi.status
ORDER BY item_type, wi.status;

-- Generated components health check
SELECT function, section_type,
  CASE WHEN html_template LIKE '{%"function"%' THEN 'BROKEN (JSON blob)'
       WHEN html_template LIKE '<style>%' THEN 'OK (HTML)'
       WHEN html_template LIKE '<%' THEN 'OK (HTML)'
       ELSE 'UNKNOWN' END as status,
  LENGTH(html_template) as len
FROM content_components WHERE created_from = 'generated'
ORDER BY function;

-- Active dispatch across all sites
SELECT s.domain, wi.item_type, wi.status, wi.claimed_at
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE wi.status = 'claimed'
ORDER BY wi.claimed_at DESC;

-- Scheduler health
SELECT name, enabled, last_triggered_at, last_completed_at
FROM scheduled_tasks
WHERE enabled = true
ORDER BY name;
```
