# Handoff: Design Adoption Pipeline — Fingerprint, Intent & CSS Fetching

Date: 2026-04-12

---

## What We Built This Session

A design fingerprint extraction and design intent pipeline for the site adoption workflow. The goal: when adopting an existing site, capture its actual CSS values (colours, fonts, layout) and pass them to the webdesign-agent so the rebuilt site looks like the original instead of a generic brochure.

---

## Current State: Where It Stopped

The adoption was re-triggered for gamedesign.uk. The **fingerprint extraction step works** (found 20 pages, extracted 8 colours from inline styles, detected dark sections) but the **external CSS fetching workflow hasn't been tested yet** — the workflow SQL (`phase_css_fetch_workflow.sql`) needs to be applied and the adoption re-run.

The previous run also failed at `classify_archetype` because `analyze_site` wasn't returning parsed JSON. The fix (`output_format: json` on the analyze_site step) has been applied to the DB.

### What's deployed (Go code in cluster)

| File | What it does | Status |
|------|-------------|--------|
| `extract_design_fingerprint_action.go` | Parses rawHTML for colours, fonts, CSS vars, layout patterns. Collects external CSS URLs. | ✅ Deployed, tested, works |
| `enrich_fingerprint_with_css_action.go` | Merges fetched CSS content into fingerprint. Reuses `fp*` parsing functions. | ✅ Deployed, not yet tested |
| `apply_adoption_plan_action.go` | Writes `design_reference` spec from fingerprint (patch 1d already applied before this session). Patch 1e (enrich work item spec with fingerprint) — **needs applying**. | ⚠️ 1d done, 1e patch ready but not confirmed applied |
| Registry entries | Both `extract_design_fingerprint` and `enrich_fingerprint_with_css` registered | ✅ Deployed |

### What's applied (SQL in DB)

| Change | Status |
|--------|--------|
| Phase 1c: `extract_fingerprint` step in adoption workflow | ✅ Applied |
| Phase 2b: Webdesign-agent three-way prompt (design_intent → design_reference → generate) | ✅ Applied |
| Phase 2e: `generate_design_intent` + `write_design_intent` steps in adoption workflow | ✅ Applied |
| `output_format: json` on `analyze_site` step | ✅ Applied |
| CSS fetch workflow (check_has_external_css → fetch_primary_css → enrich_fingerprint) | ❌ **NOT YET APPLIED** |

### What needs doing next

1. **Apply the CSS fetch workflow SQL** (`phase_css_fetch_workflow.sql` in project outputs). This adds three steps between `extract_fingerprint` and `analyze_site`:
   - `check_has_external_css` — conditional on `design_fingerprint.has_external_css == true`
   - `fetch_primary_css` — `firecrawl_scrape` via webscrape adapter, URL from `design_fingerprint.primary_css_url`
   - `enrich_fingerprint` — `enrich_fingerprint_with_css`, merges CSS data into fingerprint

2. **Re-trigger gamedesign.uk adoption** and verify:
   - Fingerprint step extracts external CSS URLs (look for `external_css_urls: 1` in logs)
   - CSS fetch step retrieves `global.css` via webscrape adapter
   - Enrich step parses it and finds `--bg-color`, `--primary-color`, `--font-family` etc
   - analyze_site returns valid JSON (output_format fix)
   - apply_plan writes `design_reference` spec with enriched data
   - generate_design_intent produces semantic brief
   - write_design_intent persists `design_intent` spec

3. **Check the resulting specs** after adoption completes:
   ```sql
   SELECT aspect, source, LEFT(data::text, 200) as preview
   FROM site_specs
   WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
     AND is_current = true
     AND aspect IN ('design_reference', 'design_intent', 'identity', 'site_archetype', 'content_direction')
   ORDER BY aspect;
   ```

4. **Pause work items and snapshot** before letting the build proceed:
   ```sql
   UPDATE site_work_items SET status = 'paused'
   WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
     AND pipeline = 'build' AND status = 'triaged';

   SELECT take_site_snapshot(
       (SELECT id FROM sites WHERE domain = 'gamedesign.uk'),
       'manual', NULL,
       'Post-adoption with design_reference and design_intent - before build',
       'admin'
   );
   ```

5. **Unpause and watch the build** — specifically watch the webdesign-agent run. It should now read `design_intent` from site_specs and generate CSS matching gamedesign.uk's original palette (`#121212` bg, `#00bcd4` accent, system UI fonts).

6. **Apply patch 1e** if not already done — enriches the `needs_design` work item spec with fingerprint data (currently uses LLM design description as fallback).

---

## Adoption Workflow (current deployed state)

```
ensure_site_record
  → crawl_site (firecrawl)
  → format_crawl (summaries)
  → check_crawl_content
  → extract_fingerprint (Go — extracts colours/fonts/layout from rawHTML, collects CSS URLs)
  → [PENDING: check_has_external_css → fetch_primary_css → enrich_fingerprint]
  → analyze_site (LLM — output_format: json)
  → classify_archetype (LLM)
  → select_content (Go)
  → derive_content_direction (LLM)
  → apply_plan (Go — writes specs, pages, work items)
  → generate_design_intent (LLM — semantic brief from fingerprint + identity)
  → write_design_intent (write_site_spec)
  → complete
```

---

## Webdesign-Agent Prompt (deployed)

Three-way priority in the `analyze_design` step:

1. **`design_intent` exists** → creative freedom within described character. Reference values as starting points, not targets. Agent explains choices in design_notes.
2. **`design_reference` exists, no `design_intent`** → reproduce faithfully. Use reference values directly. Don't invent new palette.
3. **Neither exists** → generate from industry/audience/identity. Standard new-build path.

The prompt was updated via DO block with `$do$`/`$prompt$` dollar-quoting (single-quote escaping broke twice with other methods).

---

## Key Decisions Made

1. **`design_reference`** = concrete extracted values (historical, immutable). **`design_intent`** = semantic direction (evolving, creative freedom). Replaces the old vague `design` spec.

2. **design_intent is semantic not prescriptive** — describes character ("dark IDE aesthetic, functional not atmospheric") with reference values as guidance, not targets. The webdesign-agent has creative freedom. This lets the improvement loop evolve the palette.

3. **Palette is locked until design_intent exists.** First adoption build: only design_reference → locked reproduction. Once design_intent is written (auto-generated at adoption, or by strategist/human later): creative freedom within character.

4. **Audit loop proposes but doesn't apply** design changes directly. Creates work items. Design_intent updates flow through strategist or human.

5. **External CSS fetched via webscrape adapter** (Option B), not direct HTTP from Go action. Keeps all external fetching through the same adapter infrastructure. Error handling: if fetch fails, pipeline continues with thin fingerprint.

6. **Three-stage processing**: Go design extraction → LLM classification → Go content extraction. Design data extracted deterministically from CSS. LLM focuses on classification and reasoning.

---

## Phase 0: Nav Fixes (from separate investigation)

Go patches ready but not yet deployed. These are independent of the design pipeline.

| # | Task | Status |
|---|------|--------|
| 0a | `plan_sections` filter — strip header/footer from sections | Patch ready |
| 0b | InjectHeader/Footer skip-if-present guard | Patch ready |
| 0c | Responsive fix idempotency (case-sensitive check) | Patch ready |
| 0g | LogoURL as direct RenderContext field (5 files) | Patches ready |

Patches are in project outputs as `patch_0a_plan_sections_filter.go`, `patch_0b_inject_skip_guard.go`, `patch_0c_responsive_fix_idempotency.go`, `patch_0g_logo_url_all_files.go`.

---

## Phase Status Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 0 | Nav fixes | Go patches ready, not deployed |
| 1 | Fingerprint extraction | ✅ Go deployed. SQL applied. CSS fetch workflow pending. |
| 2a | load_site_for_design update | ✅ Not needed — read_site_spec loads all aspects |
| 2b | Webdesign prompt three-way priority | ✅ Applied to DB |
| 2c | Baseline awareness (existing CSS) | Not started |
| 2e | Generate design_intent in adoption workflow | ✅ Applied to DB |
| 3 | Site-design-planner agent (nav, layout specs) | Not started |
| 4 | Requirement-driven components, recipes, identity library | Not started |

---

## Files in Project Outputs

| File | Purpose |
|------|---------|
| `design_adoption_work_plan.md` | Full work plan with decisions, status, phases |
| `007_adoption_pipeline.md` | Updated adoption doc |
| `extract_design_fingerprint_action.go` | Latest patched version (with CSS URL collection) |
| `enrich_fingerprint_with_css_action.go` | New action for merging fetched CSS |
| `phase_css_fetch_workflow.sql` | Workflow SQL for CSS fetch steps — **APPLY THIS NEXT** |
| `phase1_bc_registry_and_workflow.sql` | Already applied |
| `phase2b_webdesign_prompt_update.sql` | Already applied |
| `phase2e_adoption_design_intent.sql` | Already applied |
| `patch_1e_work_item_spec.go` | Enriches needs_design work item — may need applying |
| `patch_0a_plan_sections_filter.go` | Nav fix — not deployed |
| `patch_0b_inject_skip_guard.go` | Nav fix — not deployed |
| `patch_0c_responsive_fix_idempotency.go` | Nav fix — not deployed |
| `patch_0g_logo_url_all_files.go` | Logo fix across 5 files — not deployed |

---

## gamedesign.uk Reference

The original site (before adoption) uses:
- `--bg-color: #121212` (near-black background)
- `--surface-color: #1e1e1e` (dark card surfaces)
- `--primary-color: #00bcd4` (cyan accent)
- `--text-main: #e0e0e0` (light text)
- `--font-family: 'Segoe UI', Roboto, Helvetica, Arial, sans-serif` (system UI, NOT monospace)
- Dark theme throughout, spacious layout, pillar cards with heavy shadows
- Minimal header: brand name with cyan `.uk` accent, three nav links, no CTA button
- One-line centred footer

The previous adoption build produced a completely different site — wrong colours (#0f1923/#00d4ff), wrong fonts (JetBrains Mono), wrong layout (generic brochure), wrong header (8 nav items + CTA). The fingerprint pipeline is designed to prevent this by giving the webdesign-agent the actual values.

---

## Diagnostic Queries

```sql
-- Check adoption workflow state
SELECT status, current_step, LEFT(error::text, 300) as error
FROM orchestration_states
WHERE orchestration_id IN (
    SELECT orchestration_id FROM orchestration_states
    WHERE collected_data::text LIKE '%gamedesign.uk%'
    ORDER BY created_at DESC LIMIT 1
);

-- Check what specs exist for gamedesign.uk
SELECT aspect, source, LEFT(data::text, 200) as preview, created_at
FROM site_specs
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND is_current = true
ORDER BY aspect;

-- Check work items
SELECT item_type, status, handler_agent, LEFT(spec::text, 100) as spec_preview
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND pipeline = 'build'
ORDER BY priority;

-- Verify CSS fetch workflow is wired correctly
SELECT
    default_config->'workflow'->'steps'->'extract_fingerprint'->>'next_step' as after_fingerprint,
    default_config->'workflow'->'steps'->'check_has_external_css'->'config'->>'then_step' as css_yes,
    default_config->'workflow'->'steps'->'check_has_external_css'->'config'->>'else_step' as css_no,
    default_config->'workflow'->'steps'->'fetch_primary_css'->>'next_step' as after_fetch,
    default_config->'workflow'->'steps'->'enrich_fingerprint'->>'next_step' as after_enrich
FROM agent_definitions WHERE type = 'site-adoption-agent';
```
