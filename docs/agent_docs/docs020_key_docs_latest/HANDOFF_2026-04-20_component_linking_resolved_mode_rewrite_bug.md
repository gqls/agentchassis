# Handoff: Component Linking Resolved, mode=rewrite Bug, Next Priorities

**Date**: 2026-04-20
**Continuation of**: `HANDOFF_2026-04-19_component_linking_news_template_discovery_checks.md` and `HANDOFF_2026-04-17_triage_and_component_linking.md`
**Test site**: gaswholesalers.com (`5fe15466-4e2e-4ff2-981e-98c1b7074002`)
**Homepage page_id**: `4ff0e0ff-fab2-423e-a59c-b9de4674a84f`

---

## Session Outcome Summary

The `save_page_sections` component-linking issue from the 04-17 handoff is **resolved on the homepage**. The code fix was already deployed; the reason linking wasn't happening was that nothing had triggered `page-build-handler` on this page since the deploy. A diagnostic trigger revealed a separate issue: `mode: rewrite` in the work item spec causes the content writer to produce output in a shape `save_page_sections` can't read, so it silently skips. A second diagnostic run without `mode: rewrite` linked all 7 components correctly.

| Issue | 04-17 state | 04-19 state | 04-20 state |
|---|---|---|---|
| `save_page_sections` component_id | P1 — file ready, not deployed | Parked — code on production but not linking | **Resolved on homepage**. 7/7 linked. Root cause: `mode: rewrite` in spec caused skip. |
| news-listing template `data-component` | P1 — SQL drafted | SQL verified, tx ready | **Committed**. Template now includes `data-component="news-listing"`. |
| mode: rewrite causes section-save skip | unknown | unknown | **NEW** P1 — investigate and fix or document |
| Stale 04-14 unlinked rows across other sites/pages | known via diagnostic | known via diagnostic | Will self-heal on next natural rebuild, pending |
| Checks falling off completeness-discovery-agent | P2 | Closed | Closed |
| content-feed-refresh scheduler routing | P2 | P2 | P2 — next |
| Header nav missing News link | P3 | P3 | P3 |
| Improvement-sweep site starvation | P3 | P3 | P3 |
| Topic splitting | Feature | Feature | Feature |

---

## 1. `save_page_sections` component_id — RESOLVED on homepage

### Evidence

Two diagnostic work items were inserted against the gaswholesalers homepage:

**Run 1** (`diagnostic_enrichment_4ff0e0ff_20260419`, spec included `mode: rewrite`):
- Status: `complete`, 9.5 min wall clock, deploy succeeded, `/index.html` committed to git
- `sections_saved`: `{"skipped": true, "sections_saved": 0, "reason": "no HTML content and no sections metadata"}`
- `page_components` rows unchanged (still dated 2026-04-14)

**Run 2** (`diagnostic_enrichment_no_mode_4ff0e0ff_20260420`, identical spec but `mode` key omitted):
- Status: `complete`
- `sections_saved`: `{"sections_saved": 7, "skipped": null}`
- All 7 `page_components` rows now have `component_id` populated
- `slot_name` for position 4 normalized from `differentiators-section` to `differentiators` — proves the "prefer data-component over metadata" enrichment branch fires correctly

### Verification query (re-run any time)

```sql
SELECT pc.position, pc.slot_name, pc.component_id IS NOT NULL AS linked
FROM page_components pc
WHERE pc.page_id = '4ff0e0ff-fab2-423e-a59c-b9de4674a84f'
ORDER BY pc.position;
-- Expected: all 7 linked = true, slot_name for position 4 = 'differentiators'
```

### What this means for other stale pages

The diagnostic query from 04-19 showed 14 pages across 3 sites with `linked_components = 0`. These have rows written by pre-fix code. They will self-heal on their next natural `page-build-handler` run, provided that run does **not** pass `mode: rewrite` in the spec. No manual intervention is needed unless we want to force it.

---

## 2. news-listing template — COMMITTED

Template updated on 04-19:
- `id`: `11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f`
- Now starts: `<section data-component="news-listing" class="news-listing-section" id="news-listing">`

The one existing `page_components` row that uses this template (gaswholesalers `/news.html`, position 2) is still unlinked because it holds HTML snapshotted before the template change. It will link on the next rebuild of that page.

---

## 3. NEW P1: `mode: rewrite` causes section-save skip

### What we observed

When a `site_work_items` spec contains `"mode": "rewrite"`, the `page-build-handler` workflow completes successfully and deploys the page to git, but `save_page_sections` returns `{"skipped": true, "reason": "no HTML content and no sections metadata"}` and `page_components` rows are never updated.

Same page, same handler, same workflow definition — without the `mode` key, save_sections runs normally and persists 7 sections.

### Why this matters

- Sites get deployed but their `page_components` stay out of sync with the deployed HTML
- `rerender-pages` runs against stale `page_components` and produces inconsistent output
- Discovery checks like `unlinked_page_components` keep firing even after rebuilds
- The workflow reports success, so there's no error signal anywhere

### What we don't yet know

- Whether `mode: rewrite` was ever a deliberate supported mode
- Where `mode` is read in `page-content-writer` and what branch it takes
- Whether the broken shape (missing `page_content.response.sections_metadata` and `validation_result.clean_html`) is the writer's doing or something earlier
- Whether other `mode` values exist (e.g. `recreate`, which is referenced in `load_existing_content.mode`)

### Next-step investigation plan

1. Grep the codebase for `"mode": "rewrite"` usages — identify who's emitting it
2. Trace `page-content-writer`'s handling of `build_mode` / `input_data.spec.mode`
3. Compare writer output shape between a normal run and a `mode: rewrite` run
4. Decide: fix the mode path to produce compatible output, or remove all callers that emit `mode: rewrite`

This is today's next task (section 6 below).

---

## 4. Schema and Workflow Notes Confirmed This Session

- `page-build-handler` was re-saved at 2026-04-19 18:44 — workflow definition is current and matches the backup SQL
- `complete` step's `output_fields: ["sections_saved", "deploy_result"]` filters the workflow result — intermediate step outputs (`page_content`, `validation_result`, etc.) are populated during the run but stripped from the final `result.response` row. Don't infer "step didn't run" from "its output_field isn't in the result."
- `complete_error`'s `output_fields: ["page_content", "site_record"]` — seeing those two in a result means the error branch fired
- `page-rerender` and `rerender-site` agents **do not** call `save_page_sections` — only `page-build-handler`, `site-work-orchestrator`, `page-rebuild`, `pageflow-builder`, and `tool-recreation-handler` do

---

## 5. Open Issues (re-listed, current state)

### P1 — Active

| # | Issue | Notes |
|---|---|---|
| 1 | **`mode: rewrite` causes section-save skip** | NEW. Workflow reports success, deploy succeeds, but `page_components` not updated. Investigation plan in section 3. |

### P2 — Queued

| # | Issue | Notes |
|---|---|---|
| 2 | content-feed-refresh scheduler routing | Trigger never fires via scheduler; manual kcat workaround in use. Two possible fixes sketched in 04-17 handoff. |
| 3 | Stale unlinked `page_components` across other sites | 14 pages across 3 sites identified 04-19. Will self-heal on next natural rebuild (without `mode: rewrite`). No forced action unless we want faster cleanup. |

### P3 — Low priority

| # | Issue | Notes |
|---|---|---|
| 4 | Header nav missing News link | Footer has it; header re-render didn't include it. |
| 5 | Improvement-sweep site starvation | Oldest `updated_at` site always wins. |
| 6 | Stale comment in discovery_checks Go source | Shows outdated 8-item checks array as a copy-paste example. Safe but a landmine. |
| 7 | `updateAgentWorkflow` replaces whole workflow tree | Currently safe (no caller), but will silently erase unrelated steps when an automated proposal generator gets wired up. Switch to deep-merge before then. |

### Feature work

| # | Item | Notes |
|---|---|---|
| 8 | Topic splitting (split single Grok source into geopolitical/regulatory/pricing/transition) | SQL-only; plan in `031_news_content_diversity_plan.md`. |

---

## 6. Next Action

Investigate `mode: rewrite` per section 3. Starts with a codebase grep to find who emits it and who consumes it. Results and a proposed fix will be appended to the next handoff, or — if the investigation reveals the mode should simply not be in use — documented here with the list of call sites that need changing.

---

## Monitoring Queries

```sql
-- Component linking rate across recent rebuilds
SELECT
    p.site_id, p.id AS page_id, p.name,
    MAX(pc.updated_at) AS last_rebuild,
    COUNT(*) AS total_components,
    COUNT(pc.component_id) AS linked_components
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
WHERE pc.updated_at > now() - interval '7 days'
GROUP BY p.site_id, p.id, p.name
HAVING COUNT(*) > COUNT(pc.component_id)
ORDER BY last_rebuild DESC
LIMIT 20;

-- Recent page-build-handler runs: did save_sections run or skip?
SELECT
    swi.page_id, swi.item_key, swi.completed_at,
    (swi.result->'response'->'sections_saved'->>'skipped')::text        AS save_skipped,
    (swi.result->'response'->'sections_saved'->>'sections_saved')::int  AS save_count,
    (swi.spec->>'mode')                                                 AS spec_mode
FROM site_work_items swi
WHERE swi.handler_agent = 'page-build-handler'
  AND swi.completed_at > now() - interval '7 days'
  AND swi.status = 'complete'
  AND swi.result->'response' ? 'sections_saved'
ORDER BY swi.completed_at DESC
LIMIT 30;

-- Checks list on each discovery agent (sanity check — haven't shortened)
SELECT
    type,
    jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') AS check_count,
    updated_at
FROM agent_definitions
WHERE type IN ('completeness-discovery-agent', 'design-discovery-agent', 'quality-discovery-agent')
  AND deleted_at IS NULL
ORDER BY type;
```
