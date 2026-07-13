# HANDOFF — 2026-05-21 — FAQ prevention deployed, news files_field verified

## One-paragraph summary

The FAQ empty-items bug is fully resolved and its prevention is deployed.
Root cause was a duplicate-content-surface plan (`generic-text-block` +
`faq` on one page) made by the planners — NOT a content-writer fault,
proven by an isolated build test that produced a populated 9-item
accordion. The live gaswholesalers faq page was repaired through the
pipeline (q_count=10, deployed). Prevention shipped on three fronts: both
planner prompts edited (live now, no rebuild), and the dead
`validate_components` flag implemented plus an archetype-aware default
(Go, deployed in chassis v1.0.1029). Separately, the news `files_field`
fix from the prior session was confirmed as the reusable mechanism and a
site-wide rerender of ai-agent-orchestration.com is in flight to bring its
stale news.html current.

## What shipped this session

### FAQ prevention — all deployed
- **content-gap-planner prompt** (`fix1_content_gap_planner_prompt.sql`):
  removed the hardcoded `generic-text-block` from the new_page example,
  fixed the add_to_page example to `faq`, added section-selection rules
  (don't pair structured components with generic-text-block; use function
  names not display names). Confirmed live (`gap_has_rule = t`).
- **site-planner prompt** (`fix2_site_planner_prompt.sql`): component list
  now shows `[function]` first with an instruction to use only that
  (kills the `"FAQ Section"` display-name leak); added faq/pricing to
  standard mappings; added the no-pairing rule; fixed `call_to_action` →
  `call-to-action`. Confirmed live (`site_has_faq_mapping = t`).
- **Go (chassis v1.0.1029):**
  - `validate_components` flag now actually implemented in
    `ValidateSitePlanAction` — resolves each section name to a real
    `content_components.function` (function check → NormalizeComponentFunction
    → display_name/name DB lookup → drop+log). Was a dead flag before.
  - Same resolver wired into `applyNewPage` and `applyAddToPage` (the
    gap-planner path doesn't route through validate_site_plan).
  - `defaultSectionsForPage` — archetype-aware default (faq/contact/
    pricing/about get their shapes; unknown keeps generic).
  - Files: `v3_site_actions.go`, `apply_gap_plan_action.go` (patched
    against the user's actual uploads, gofmt-clean).

### Prompts take effect immediately
`loadAgentDefinition` reads per-spawn (no cache), so the prompt fixes are
already active. The Go fixes are in the deployed image.

## In flight — let it run

**ai-agent-orchestration.com site-wide rerender**, batch queued
2026-05-21 10:07:38, 30+ pages. The user chose to let it run its course.
- news.html work item: `d2239c4b-76a4-4450-b91e-65131a1e9a36`
  (page_id `4a5758fc-0bde-4236-a7a6-f02211d4f427`, site
  `2a8ebf9c-20a2-4c39-b191-840b012371da`).
- Diagnosis: news.html is a **stale render**, not broken — its
  news-listing component is linked, `js_content` is 3092 bytes, and the
  site has 849 feed items. Last committed 3 weeks ago, before the
  files_field fix, so its `news-listing.js` was never deployed. The
  rerender should deploy it (expect `file_count=2`).

Verify when it lands:
```sql
SELECT orchestration_id, status,
       collected_data->'deploy_result'->'response'->'data'->>'files_count' AS file_count,
       jsonb_pretty(collected_data->'deploy_result'->'response'->'data'->'files') AS files
FROM orchestration_states
WHERE owner_agent_type = 'page-rerender'
  AND collected_data->'input_data'->>'page_id' = '4a5758fc-0bde-4236-a7a6-f02211d4f427'
  AND created_at > NOW() - INTERVAL '4 hours'
ORDER BY created_at DESC LIMIT 3;
```
```bash
cd ~/projects/sites && git fetch origin --quiet
git ls-tree origin/master ai-agent-orchestration.com/tools/assets/   # expect news-listing.js
```

**Watch the batch drains cleanly.** 30 pages queued at once leans on the
two known-fragile, still-unaddressed areas (consumer-group fix held back,
collected_data/OOM bloat). A long `triaged` stall = that resurfacing, not
news-specific. Reapers should clear genuine stalls.

## Explicitly held back

- **Consumer-group fix** — NOT in v1.0.1029. Needs its own chassis rebuild
  + Kafka topic resets, scheduled deliberately. Structurally-correct fix
  is ready and documented; it's orthogonal to today's changes.

## Where to resume (priority order)

1. Confirm the ai-agent-orchestration.com batch (esp. news.html) lands
   cleanly — `file_count=2`, news-listing.js in git, browser shows listing.
2. Validate the prevention end-to-end: next time a planner runs (or
   trigger against a test brief), confirm a faq page comes out
   `["hero","faq","call-to-action"]` with no generic-text-block and no
   display-name section. Same isolated-build pattern used to diagnose.
3. Then pick from the structural backlog (see TODO_2026-05-21):
   consumer-group fix + topic resets; collected_data/OOM bloat; the
   `needs_section_data`-on-success question; planner depth (briefs +
   stale-plan write-back); post-build structured-field validation (Fix D).
4. User-visible polish: logo.png/favicon.ico 404s; multi-file commit
   message; tool-gas-unit-converter rerender.

## Key facts to carry forward

- **files_field fix** (page-rerender deploy_page) is the reusable mechanism
  for deploying any component's `js_content` to `/tools/assets/{function}.js`.
  Applies to all sites/components. This is why the ai-agent-orchestration
  news rerender will work.
- **"Renders empty" is a data-binding diagnosis, not a template one** —
  walk rendered HTML → page_components (orphan? data?) → content_components
  (what key?) → content_data/content_item → sections-vs-plan. Don't rerender
  before clearing the data path.
- **Two rerender paths:** site-wide `rerender-pages` creates
  `site_work_items` (item_type=page_rerender); single-page `page-rerender`
  is an orchestration only (no work item).
- **gaswholesalers** site `5fe15466-4e2e-4ff2-981e-98c1b7074002`; faq
  repaired and live.
- Chassis at **v1.0.1029**. Namespace `ai-persona-system`, Kafka in
  `kafka` ns (cluster `personae-kafka-cluster`).

## Doc index

Updated worklist: `TODO_remaining_work_2026-05-21.md`.
FAQ: `faq_empty_items_prevention_findings.md`,
`page_content_creation_flow.md`,
`016_debugging_guide_addendum_faq_diagnosis.md`,
`site_planner_depth_and_freshness_concerns.md`,
`validate_components_implementation.md`, `planner_prompt_fixes_defect1.md`.
Code/SQL shipped: `v3_site_actions.go`, `apply_gap_plan_action.go`,
`fix1_content_gap_planner_prompt.sql`, `fix2_site_planner_prompt.sql`.
News: `006_news_feed_pipeline_addendum_rendering.md`,
`component_asset_pipeline_concerns.md`, `rerender_pages_workflow_findings.md`.
