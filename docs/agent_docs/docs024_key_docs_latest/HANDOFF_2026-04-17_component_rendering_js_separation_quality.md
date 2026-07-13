# HANDOFF — Component Rendering, JS Separation, Quality Tracking

**Date:** 2026-04-17
**Site:** vonc.com (primary), system-wide impact
**Improvement loop:** OFF throughout this session

---

## Problem Origin

vonc.com is a Tier 3 submission (interactive-platform, "Spark" social game). CSS was rendering as visible text on deployed pages. Root cause: component templates were truncated at max_tokens=4000, producing unclosed `<section>` tags. The browser rendered subsequent CSS/JS as text.

---

## What Was Done (Deployed)

### 1. Component-creator agent — context-aware generation
**File:** `fix_01_component_creator_context.sql` (applied)
- Workflow expanded: `ensure_site_record → read_site_spec → generate_template → store_component`
- Prompt now includes mission_brief, design_intent, content_direction, classification from site_specs
- max_tokens bumped 4000 → 16000

### 2. JS content separation (Approach 2)
**Files:** `store_generated_component_action.go`, `rerender_single_page_action.go`, `fix_page_rerender_workflow.sql` (all deployed)
- Added `js_content TEXT` column to `content_components`
- `separateInlineJS()` in store action: extracts inline `<script>` blocks from template, stores in `js_content`, replaces with `<script src="/tools/assets/{function}.js">`
- `collectJSAssets()` in rerender action: queries `js_content` for page's components, returns `files` map
- page-rerender workflow updated: `files_field: rendered_page.files` for multi-file git commit
- **Verified working:** `provocation-feed` on provocations page shows `SCRIPT SRC`, `archetype-combinations` on archetypes shows `SCRIPT SRC`

### 3. plan_sections improvements
**File:** `plan_sections_action.go` (deployed)
- Template integrity check: `template_valid` via `LIKE '%</section>%'` — truncated templates skipped, fall to Path 3 for regeneration
- `designDirection` extracted from `design_intent.style_direction` before section loop
- `sectionDescription()` method on sourceResolver reads from site_plan spec
- Fixed duplicate Path 3 block (merge artifact)

### 4. page-build-handler contract fix
**Files:** `fix_page_build_handler_contract.sql` (applied), `fix_load_page_record_fallback.go` (deployed)
- `load_page_record` config changed from `input_data.page_name` → `input_data.spec.page_name` (matching save_sections, update_status)
- Go action has defensive fallback chain: tries configured path, then `input_data.spec.page_name`, then other common paths
- **Root cause of gauntlet/archetypes pages getting 0 components:** `input_data.page_name` was nil because the dispatch loop only populates `input_data.spec.page_name`

### 5. Component-creator prompt revision (tiered field classification)
**File:** `revise_component_creator_prompt.sql` (applied)
- Section 3 (TEMPLATE VARIABLES) rewritten with three tiers:
  - **Tier A — Voice content:** `source: "llm"`, `required: true`. Headlines, CTAs, intro paragraphs. 3-10 per component.
  - **Tier B — Tunable labels:** `source: "static"`, `required: false`, with `fallback`. Button text, stat labels, badge text. 5-20 per component.
  - **Tier C — Site data:** `source: "site_specs.*"` or `site_assets.*`.
- Added template/schema sync invariant: every `{{.x}}` must have a schema entry and vice versa
- Added JS separation note: "pipeline will extract inline `<script>` blocks automatically"

### 6. Component quality tracking
**Files:** `migration_component_quality.sql` (applied), `compute_component_quality_action.go` (pending Go deploy), `agent_component_quality_auditor.sql` (applied)
- New columns on `content_components`: `template_variable_count`, `schema_field_count`, `template_closed`, `schema_template_synced`, `has_data_component`, `quality_score` (0-100), `quality_checked_at`, `quality_issues` (jsonb array)
- Scoring formula: starts at 100, deducts -50 for 0 variables on section components, -30 for unbalanced section tags, -30 for variables with empty schema, -20 for template/schema mismatch, -10 for missing data-component
- `compute_component_quality` Go action: standalone or batch mode, callable from store_generated_component inline or by the auditor agent
- `ScoreAndPersistComponent()` helper for inline use from store_generated_component
- `component-quality-auditor` agent definition created (active, analyst category)
- Backfill work item needs inserting (see Pending Actions below)

### 7. system.internal site
- Created for maintenance/library-level work items that aren't site-specific
- `id: eac60db8-b032-432b-b36d-76f37632045d`, `domain: system.internal`
- `brand_dna: {"is_system": true, "description": "Internal site for library-level work items and system components. Never deployed."}`
- `network_id` is NULL (no FK — the site was created without one to avoid the networks table requirement)

### 8. Contracts doc updated
**File:** `003_contracts_and_standards.md` (updated)
- Handler agent contract: expanded with input data path table, good/bad config examples, within-workflow consistency rule, action-level defense pattern
- JS Content Separation Contract: full data flow, asset path convention, relationship to js_snippets, rules
- Component Quality Contract: columns, scoring formula, when quality gets computed, what low quality triggers

---

## Current vonc.com State

### Pages

| Page | build_status | Components | Content |
|------|-------------|------------|---------|
| index | deployed | 4 | provocation-card (INLINE JS), lobby-grid (INLINE JS), brief-explanation, gauntlet-cta |
| provocations | deployed | 1 | provocation-feed (SCRIPT SRC ✓) |
| about | deployed | 4 | hero, platform-comparison, game-master-explanation, gauntlet-cta — all OK |
| how-it-works | deployed | 8 | Full content |
| archetypes | deployed | 3 | hero, archetype-grid (INLINE JS), archetype-combinations (SCRIPT SRC ✓) |
| gauntlet | deployed | 0 | **EMPTY — needs rebuild** |
| contact | deployed | 0 | Empty — needs rebuild |
| membership | deployed | 0 | Empty — needs rebuild |
| tool pages (4) | planned/deployed | 0-1 each | Various states |

### Missing components (deleted, need regeneration)

| function | Why deleted | Status |
|----------|-------------|--------|
| gauntlet-interface | 0 template variables, 35 schema fields (out of sync) | Deleted, needs `needs_new_component` item |
| archetype-result-card | empty input_schema `{}` | Deleted, needs `needs_new_component` item |
| archetype-combinations | 0 template variables, 56 schema fields (out of sync) | Deleted, needs `needs_new_component` item |
| provocation-feed | 0 template variables, 16 schema fields (out of sync) | Deleted, needs `needs_new_component` item |

### Existing components with 0 template variables (system-wide)

43 generated components have 0 template variables. This is a pre-existing condition from the old prompt. Many are deployed across multiple sites and "work" because content is baked in. The quality auditor will target these for regeneration once the Go action is deployed. Do NOT mass-delete — other sites depend on them.

Notable ones on vonc:
- `archetype-grid`, `lobby-grid`, `gauntlet-cta`, `provocation-card`, `brief-explanation`, `platform-comparison`, `game-master-explanation` — all have 0 variables but are used on deployed pages

### Work items

| item_type | status | count | Notes |
|-----------|--------|-------|-------|
| needs_section_data | needs_human_review | 12 | Sections waiting for data — gauntlet-interface and archetype-result-card are in here |
| empty_section | needs_human_review/unresolved | 5 | Empty sections detected by improvement loop |
| content_rewrite | needs_human_review | 1 | Validation failure |
| missing_css | unresolved | 1 | vonc has no style_collection_id set |
| nav_drift | triaged | 1 | Nav tables may need refresh |

---

## Pending Actions (Next Session)

### Immediate: Unblock vonc

1. **Insert backfill work item** for component-quality-auditor (using system.internal site):
```sql
INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary,
  spec, priority, handler_agent, status, created_by, item_key
)
VALUES (
  'eac60db8-b032-432b-b36d-76f37632045d',
  'backfill', 'maintenance', 'component_quality_scan', 'low',
  'Score all existing components (backfill)',
  '{"scan_all": true}'::jsonb,
  10, 'component-quality-auditor', 'triaged', 'manual',
  'backfill_component_quality_' || extract(epoch from now())::int
);
```

2. **Run vonc unblock SQL** (`unblock_vonc_after_quality.sql`):
   - Clears page_components for gauntlet + archetypes
   - Resets build_status to needs_rebuild
   - Creates `needs_content_page` work items
   - Dispatch loop → plan_sections finds missing components → creates `needs_new_component` items → component-creator generates with revised prompt → pages rebuild

3. **Monitor regeneration** — the new components should have:
   - Template variables matching schema fields (schema_template_synced = true)
   - JS separation (js_content populated, html_template has `<script src>`)
   - quality_score > 60

```sql
-- Monitor
SELECT function, section_type,
  (LENGTH(html_template) - LENGTH(REPLACE(html_template, '{{', ''))) / 2 as var_count,
  CASE WHEN js_content IS NOT NULL AND js_content != '' THEN 'YES' ELSE 'no' END as has_js,
  CASE WHEN html_template LIKE '%/tools/assets/%' THEN 'YES' ELSE 'no' END as has_script_src,
  CASE WHEN html_template LIKE '%</section>%' THEN 'OK' ELSE 'TRUNCATED' END as closed,
  quality_score,
  created_at
FROM content_components
WHERE created_from = 'generated'
  AND section_type IN ('gauntlet-interface', 'archetype-result-card', 'archetype-combinations', 'provocation-feed')
ORDER BY created_at DESC;
```

### Go Deploy Required

These Go files need deploying (in order):

1. **`compute_component_quality_action.go`** — NEW file. Register in:
   - `registry.go`: `"compute_component_quality": { Handler: ComputeComponentQualityAction, Category: "site", Description: "Score and store component quality metrics", IsLocal: true }`
   - `local_actions.go`: `"compute_component_quality": true`

2. **Integration into `store_generated_component_action.go`** — After the INSERT block, add call to `ScoreAndPersistComponent()` (code at bottom of compute_component_quality_action.go file). This gives every newly-generated component a quality score on creation.

3. **`fix_load_page_record_fallback.go`** — Already deployed (the code change to load_page_record_action.go with the fallback chain). Confirm it's in production.

### After vonc is unblocked

1. **contact, membership pages** — also have 0 components. Need work items created.
2. **missing_css** — vonc has no style_collection_id. The webdesign-agent generated design specs but never deployed CSS. Needs investigation.
3. **Improvement loop** — currently OFF. Once vonc has content on all pages, consider re-enabling for vonc only.
4. **Quality auditor backfill** — won't run until compute_component_quality Go action is deployed and registered. The backfill work item is queued (or needs inserting per step 1 above).
5. **Document system.internal convention** — add to contracts doc as convention for maintenance work items.
6. **Component regeneration pipeline** — the quality auditor creates `needs_component_regeneration` items but the component-creator may need a workflow path to handle regeneration (delete old → create new → rerender affected pages). This workflow path doesn't exist yet — currently component-creator only handles `needs_new_component`.

---

## Key Files in Outputs

| File | Purpose | Status |
|------|---------|--------|
| `plan_sections_action.go` | Template integrity, design direction, section descriptions | Deployed |
| `store_generated_component_action.go` | JS separation (separateInlineJS), js_content INSERT | Deployed |
| `rerender_single_page_action.go` | collectJSAssets, files map return | Deployed |
| `fix_page_rerender_workflow.sql` | files_field for multi-file git commit | Applied |
| `fix_page_build_handler_contract.sql` | load_page_record uses input_data.spec.page_name | Applied |
| `fix_load_page_record_fallback.go` | Defensive fallback chain in Go action | Deployed |
| `revise_component_creator_prompt.sql` | Tiered field classification (A/B/C) | Applied |
| `migration_component_quality.sql` | Quality tracking columns | Applied |
| `compute_component_quality_action.go` | Quality scoring Go action + ScoreAndPersistComponent helper | **Pending Go deploy** |
| `agent_component_quality_auditor.sql` | Auditor agent definition | Applied |
| `unblock_vonc_after_quality.sql` | Clear pages + create build items | **Pending execution** |
| `003_contracts_and_standards.md` | Updated contracts doc | Updated |

---

## Architecture Decisions Made

1. **js_content column over js_snippets table** — component JS is 1:1 with the component's HTML structure. js_snippets stays for shared design effects (parallax, scroll reveals).

2. **Spec is primary input (contract rule)** — all handler workflow configs must use `input_data.spec.*` paths. Top-level flattened paths (`input_data.page_name`) are unreliable because they depend on the dispatch loop's optional `?` input_mapping.

3. **Tiered field classification for components** — Tier A (voice content, llm/required), Tier B (tunable labels, static/optional with fallback), Tier C (site data). Prevents both "35 fields all required" and "0 fields everything hardcoded" failure modes.

4. **Quality score as first-class field** — measurable, indexable, actionable. Planner prefers higher scores, auditor targets low scores for regeneration.

5. **system.internal site** — convention for library-level work items that don't belong to any customer site.

---

## Diagnostic Queries

### Component health check
```sql
SELECT function, section_type,
  CASE WHEN html_template LIKE '%</section>%' THEN 'OK' ELSE 'TRUNCATED' END as closed,
  (LENGTH(html_template) - LENGTH(REPLACE(html_template, '{{', ''))) / 2 as var_count,
  (SELECT COUNT(*) FROM jsonb_object_keys(COALESCE(input_schema->'fields', '{}'::jsonb))) as schema_fields,
  CASE WHEN js_content IS NOT NULL AND js_content != '' THEN 'YES' ELSE 'no' END as has_js,
  quality_score
FROM content_components
WHERE is_active = true AND created_from = 'generated'
ORDER BY quality_score ASC NULLS FIRST;
```

### vonc page status
```sql
SELECT p.name, p.build_status,
  (SELECT COUNT(*) FROM page_components pc WHERE pc.page_id = p.id) as components,
  (SELECT SUM(LENGTH(pc.rendered_html)) FROM page_components pc WHERE pc.page_id = p.id) as html_bytes
FROM pages p
WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'vonc.com')
ORDER BY p.nav_order;
```

### JS separation verification
```sql
SELECT pc.slot_name, p.name as page,
  CASE WHEN pc.rendered_html LIKE '%/tools/assets/%' THEN 'SCRIPT SRC'
       WHEN pc.rendered_html LIKE '%<script>%' THEN 'INLINE JS'
       ELSE 'no js' END as js_status,
  CASE WHEN pc.rendered_html LIKE '%</section>%' THEN 'OK' ELSE 'MISSING CLOSE' END as closed,
  LENGTH(pc.rendered_html) as len
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'vonc.com')
ORDER BY p.name, pc.position;
```

### Work item flow monitoring
```sql
SELECT item_type, status, COUNT(*),
  MAX(updated_at)::text as last_activity
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'vonc.com')
  AND item_type IN ('needs_content_page', 'needs_new_component')
  AND status NOT IN ('complete', 'wont_fix')
GROUP BY item_type, status
ORDER BY item_type, status;
```

### Quality leaderboard (after backfill runs)
```sql
SELECT function, component_level, quality_score,
  template_variable_count as vars, schema_field_count as schema,
  template_closed as closed, schema_template_synced as synced,
  quality_issues->0 as first_issue
FROM content_components
WHERE is_active = true AND quality_score IS NOT NULL
ORDER BY quality_score ASC, function
LIMIT 30;
```

### Error log for vonc
```sql
SELECT occurred_at, agent_type, step_name, action,
  LEFT(error_message, 200) as error,
  LEFT(context::text, 200) as ctx
FROM agent_error_log
WHERE domain = 'vonc.com'
  AND occurred_at > now() - interval '6 hours'
ORDER BY occurred_at DESC
LIMIT 15;
```
