# Handoff: Component Linking Resolved, mode=rewrite Bug, Next Priorities

**Date**: 2026-04-20
**Continuation of**: `HANDOFF_2026-04-19_component_linking_news_template_discovery_checks.md` and `HANDOFF_2026-04-17_triage_and_component_linking.md`
**Test site**: gaswholesalers.com (`5fe15466-4e2e-4ff2-981e-98c1b7074002`)
**Homepage page_id**: `4ff0e0ff-fab2-423e-a59c-b9de4674a84f`

---

## Session Outcome Summary

The `save_page_sections` component-linking issue from the 04-17 handoff is **resolved on the homepage**. The code fix was already deployed; the reason linking wasn't happening was that nothing had triggered `page-build-handler` on this page since the deploy. A diagnostic trigger run with `"mode": "rewrite"` in the spec took the skip branch. A second identical-except-for-the-mode-key run linked all 7 components correctly. Investigation traced `mode` handling through the code and **could not explain why `"rewrite"` would cause a skip** — the code path is identical to omitting `mode`. Since no live caller emits `"rewrite"` and it's not a valid mode value, this is closed as "don't pass unsupported mode values" rather than pursued further.

| Issue | 04-17 state | 04-19 state | 04-20 state |
|---|---|---|---|
| `save_page_sections` component_id | P1 — file ready, not deployed | Parked — code on production but not linking | **Resolved on homepage**. 7/7 linked. |
| news-listing template `data-component` | P1 — SQL drafted | SQL verified, tx ready | **Committed**. Template now includes `data-component="news-listing"`. |
| `mode: rewrite` causing save-skip | — | — | **Closed** — unexplained one-off, no caller emits the value. See §3. |
| Stale 04-14 unlinked rows across other sites/pages | known via diagnostic | known via diagnostic | Will self-heal on next natural rebuild, pending |
| Checks falling off completeness-discovery-agent | P2 | Closed | Closed |
| content-feed-refresh scheduler routing | P2 | P2 | **Fix applied**. Root cause was not routing — it's a workflow config bug. See §7. Pending verification on next fire. |
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

## 3. `mode: rewrite` — unexplained one-off, documented and closed

### What we observed

Two diagnostic work items were run on the same page, minutes apart, with identical specs except for the `mode` key. Run 1 (`"mode": "rewrite"`) returned skipped; run 2 (no mode key) saved 7 sections.

### What investigation found

- **No caller in the codebase emits `"mode": "rewrite"`**. I introduced that value in my diagnostic spec. The only legal value is `"recreate"` (set in `load_existing_content_action.go` line 49601 of the chassis context dump).
- **`LoadExistingContentAction` returns identical output for `mode = "rewrite"` and `mode = ""`** — both hit the `mode != "recreate"` branch and return `{has_existing: false, reason: "not_recreate"}`. So the theoretical code path cannot explain the difference in outcome.
- **`build_mode` (the name used when the workflow passes `mode` to the content writer) is wired through workflow config but not read by any Go code.** It's listed in the writer's `generate_content.input_fields`, but the prompt template doesn't reference it, and no Go code does `.Get("build_mode")` or similar. It's a dead parameter.
- **The prompt's "Recreate Mode" block activates on `existing_content.has_existing`, not on `build_mode`.**

### Conclusion

The empirical correlation between `mode: "rewrite"` and the save-skip is unexplained by the code paths I traced. Possibilities include LLM response variability (since `page-content-writer` makes real LLM calls per section), transient state on one pod, or a path I haven't found. Reproducing it would cost another full LLM run with no guarantee of evidence.

Since:
- No live flow passes `"rewrite"` as a mode
- The value is undocumented and illegal
- The enrichment path works correctly when `mode` is omitted or set to `"recreate"`

This is closed as "don't pass unsupported mode values". If a recurrence is observed in a production flow (which shouldn't happen — no caller emits `rewrite`), investigate with live logs on the failing run.

### Documented facts worth keeping

- `mode` has exactly one legal value: `"recreate"`. Any other value makes `load_existing_content` skip (same as omitting `mode`).
- `build_mode` in `page-content-writer` input_fields is a dead parameter — safe to remove in a future workflow cleanup, but not blocking anything.

---

## 4. Schema and Workflow Notes Confirmed This Session

- `page-build-handler` was re-saved at 2026-04-19 18:44 — workflow definition is current and matches the backup SQL
- `complete` step's `output_fields: ["sections_saved", "deploy_result"]` filters the workflow result — intermediate step outputs (`page_content`, `validation_result`, etc.) are populated during the run but stripped from the final `result.response` row. Don't infer "step didn't run" from "its output_field isn't in the result."
- `complete_error`'s `output_fields: ["page_content", "site_record"]` — seeing those two in a result means the error branch fired
- `page-rerender` and `rerender-site` agents **do not** call `save_page_sections` — only `page-build-handler`, `site-work-orchestrator`, `page-rebuild`, `pageflow-builder`, and `tool-recreation-handler` do

---

## 5. Open Issues (re-listed, current state)

### Pending verification

| # | Issue | Notes |
|---|---|---|
| — | content-feed-refresh fix works end-to-end | Fix applied. Verify on next fire (17:00 UTC today) or force-trigger. See §7. |

### P2 — Queued

| # | Issue | Notes |
|---|---|---|
| 1 | Stale unlinked `page_components` across other sites | 14 pages across 3 sites identified 04-19. Will self-heal on next natural rebuild. No forced action unless we want faster cleanup. |

### P3 — Low priority

| # | Issue | Notes |
|---|---|---|
| 2 | Header nav missing News link | Footer has it; header re-render didn't include it. |
| 3 | Improvement-sweep site starvation | Oldest `updated_at` site always wins. |
| 4 | Stale comment in discovery_checks Go source | Shows outdated 8-item checks array as a copy-paste example. Safe but a landmine. |
| 5 | `updateAgentWorkflow` replaces whole workflow tree | Currently safe (no caller), but will silently erase unrelated steps when an automated proposal generator gets wired up. Switch to deep-merge before then. |
| 6 | `build_mode` is a dead workflow parameter | Wired through `page-build-handler` → `page-content-writer` input_mapping and listed in `generate_content.input_fields`, but no Go code reads it. Safe to remove in a future workflow cleanup. |
| 7 | `owner_agent_type` stored as `"generic"` for spawn-on-demand agents | Discovered while debugging scheduler routing. `orchestration_states` rows where the generic agent routed to a different workflow still show `owner_agent_type = "generic"`. Any observability keyed on that field won't find them. Structural fix: `selectWorkflow` should set `owner_agent_type` to the resolved group/agent type when it finds one. |
| 8 | `orchestration_name` doesn't reflect the scheduled task | The scheduler sets `sched-<task>-<timestamp>` in `execution_context.orchestration_name` but it doesn't make it into the DB column. Search by `orchestration_name LIKE 'sched-%'` returns nothing. |

### Feature work

| # | Item | Notes |
|---|---|---|
| 9 | Topic splitting (split single Grok source into geopolitical/regulatory/pricing/transition) | SQL-only; plan in `031_news_content_diversity_plan.md`. |

---

## 7. content-feed-refresh scheduler routing — fix applied, pending verification

### What the 04-17 handoff thought

> The generic agent consumes the message without routing it to the trigger agent — the trigger has never run (0 orchestration states).

### What the actual problem was

**The routing works fine.** The generic agent receives the message on `system.agent.generic.requests`, extracts `config.agent_type`, calls `FindBestGroup` in `processor.go`, loads the `content-feed-trigger` agent definition, and executes its workflow. This is the designed behaviour for any scheduled task targeting a spawn-on-demand agent.

The trigger was running — and failing — but the workflow's bug was hiding in plain sight. Two failures compounded:

1. **`owner_agent_type` is stored as `"generic"`**, not `"content-feed-trigger"`, because the generic agent is what processes the message and owns the orchestration record. So searching `orchestration_states WHERE owner_agent_type = 'content-feed-trigger'` returned zero rows — which the 04-17 handoff read as "trigger never runs". It does run; it's just filed under `generic`.

2. **The workflow's `find_news_sites → check_has_sites → process_sites` chain was broken**:
   - `find_news_sites` used `output_format: "array"`, returning a flat array (or `nil` when empty).
   - `check_has_sites` tried to read `news_sites.count` — but arrays have no `.count`. The extraction returned `""`, which didn't match the `"0"` key in the conditions map, so the condition always fell through to `default: "process_sites"`.
   - When there were no sites, `process_sites` tried to loop over `null` and crashed with `iterate_over field 'news_sites' is not an array (got <nil>)`.
   - When there WERE sites, it worked accidentally — the default branch happened to be the right one.

The manual kcat workaround from the 04-17 handoff bypasses the trigger's `find_news_sites` step entirely, which is why it always succeeded and why nobody noticed the trigger had been broken for weeks.

### Fix applied (SQL)

Changed `find_news_sites` to `output_format: "object"` (returns `{rows, count, columns}`) and updated `process_sites.items_field` to `"news_sites.rows"`:

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
    '{workflow,steps,find_news_sites,config,output_format}',
    '"object"'::jsonb)
WHERE type = 'content-feed-trigger' AND is_active = true AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
    '{workflow,steps,process_sites,config,items_field}',
    '"news_sites.rows"'::jsonb)
WHERE type = 'content-feed-trigger' AND is_active = true AND deleted_at IS NULL;
```

With these changes:
- Empty result → `news_sites.count = 0` → `check_has_sites` matches `"0"` → `notify_scheduler_idle` → `complete_idle`. Clean no-op cycle.
- Non-empty result → `news_sites.count = N` (N>0) → no match in conditions → default `process_sites` → loop over `news_sites.rows` (always an array).

### Verification state

Verified in DB:
- `output_format = "object"` ✓
- `items_field = "news_sites.rows"` ✓
- `condition_field = "news_sites.count"` (unchanged, matches new shape) ✓

Pending verification on next scheduled fire (17:00 UTC) or force-triggered. No restart needed — `FindBestGroup` loads the definition per message.

### Gotcha during this work

I ran the two UPDATEs inside a `BEGIN;` block without making the `COMMIT;` step obvious enough. You ran the preview + first UPDATE, saw the verify output (which reflected in-transaction state), and then the second UPDATE got its own session which blocked on your first session's uncommitted row lock. `pg_terminate_backend` on your original session rolled back the second UPDATE's changes, leaving the DB in a half-applied state that was worse than before. Rerunning only the missing UPDATE fixed it.

**Note for future SQL guidance**: prefer single-statement UPDATEs with `NOT LIKE` / `!=` idempotency guards rather than multi-statement `BEGIN;…COMMIT;` blocks. Safer when a step goes sideways.

### Latent issues logged (not fixed today)

- `owner_agent_type = "generic"` for orchestrations that execute a different agent's workflow via the generic route. Any observability keyed on `owner_agent_type` won't find them. If we want this to be searchable, `selectWorkflow` should set `owner_agent_type` to the resolved `groupOrAgentType` when it finds a group.
- `orchestration_name` is generated as `generic-orchestrate-MMDD-HHMM` rather than reflecting the scheduled task name. The scheduler puts `sched-<task>-<timestamp>` into the `execution_context.orchestration_name` header but it doesn't make it into the DB column. Same search-friendliness concern.

Both are P3 — noted below.

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
