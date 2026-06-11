# FOCUS: Dispatch Loop Diagnostic

**Status:** Open. Workarounds (manual triggers) in place. Investigation paused 2026-05-14 to keep imagery loop verification moving.

**Why this exists:** during 2G+2H verification on robot-hands.com, every `needs_imagery` work item ever emitted (8 of them) has remained in status `detected` despite dispatch (`build-dispatch-loop` agent) running successfully every ~60 seconds. The handler chain works end-to-end when manually triggered. Something between item emission and handler invocation is broken. This doc captures what we know so future sessions can pick up without re-deriving it.

## Observable state (as of 2026-05-14)

- 8 `needs_imagery` work items on robot-hands.com, all `status='detected'`, `claimed_at` null, `attempt_count=0`, `triaged_at` null.
- `build-dispatch-loop` (version 1, is_active=true) runs every ~60s. Last 10 runs all COMPLETED, none claimed our items.
- Manual `kcat` triggers DO process these items successfully end-to-end. Hero_home asset deployed via this route. So the handler chain itself works; the gap is purely in dispatch's claim step.
- Other item types are also accumulating in `detected` (audit_finding_audience: 2, needs_briefing: 1, etc.) — suggests this isn't specific to `needs_imagery`, it's systemic.

## Root cause hypothesis

**Discovery checks emit at status='detected', dispatch claims at status IN ('triaged','approved').** There's a missing automation that transitions detected→triaged.

Evidence:

1. **Index definitions explicitly filter on dispatchable statuses:**
   ```
   idx_swi_handler        WHERE status = ANY (ARRAY['triaged','approved'])
   idx_swi_site_pending   WHERE status = ANY (ARRAY['triaged','approved'])
   ```
   Postgres partial indexes are written to match the queries they serve. These indexes exist to make dispatch's "find claimable items" query fast — and they only cover triaged/approved. Therefore dispatch's query filters on those statuses.

2. **Workflow definition is consistent with this.** `build-dispatch-loop` has a `load_items` step calling `load_work_items` action (config: `item_pipeline: "build"`, `max_items: 5`). The action's source lives at `platform/orchestration/actions/load_work_items.go` (or similar) — we haven't read it yet, but the index strongly implies its WHERE clause.

3. **Schema-level corroboration.** `site_work_items` has both `triaged_at` and `claimed_at` columns. A column exists per state transition that needs auditing. The presence of `triaged_at` means the system explicitly tracks when items become triaged — i.e., it's a meaningful state, not just synonymous with "detected".

4. **Our discovery check insert.** `check_unfulfilled_imagery_plan.go` inserts rows with default status (`'detected'` per the schema default). No triage step is invoked after emission.

So the question becomes: **what mechanism transitions detected→triaged for the item types that DO get claimed?** Possibilities:

- a) A triage agent runs on a schedule and bulk-promotes detected→triaged.
- b) Discovery checks themselves insert at status='triaged' for items that need no human review.
- c) An admin/operator manually triages items via the dashboard.
- d) The status default used to be 'triaged' historically; recent change broke this.

## Investigation steps not yet taken

In rough order of cheapness:

1. **Grep `load_work_items` action source for the actual filter:**
   ```bash
   grep -rn "load_work_items\|loadWorkItems\|LoadWorkItems" platform/ | head -20
   ```
   The function body will contain the WHERE clause. Confirms or rebuts the index-based inference.

2. **Find item types currently being claimed and their emission paths.** Any type with a non-trivial count in `status='complete'` and `claimed_at IS NOT NULL` has a working dispatch path. Trace one such type back to its discovery check / insert path to see what status it emits at:
   ```sql
   SELECT item_type, COUNT(*) FILTER (WHERE claimed_at IS NOT NULL) AS claimed,
                       COUNT(*) FILTER (WHERE status = 'complete') AS completed,
                       COUNT(*) FILTER (WHERE status = 'detected') AS detected
   FROM site_work_items
   GROUP BY item_type
   HAVING COUNT(*) > 5
   ORDER BY claimed DESC;
   ```

3. **Look for a triage agent.** Anything in `agent_definitions` with "triage" in name or description:
   ```sql
   SELECT type, version, is_active, default_config #>> '{description}'
   FROM agent_definitions
   WHERE type ILIKE '%triage%' OR default_config::text ILIKE '%triage%' LIMIT 20;
   ```

4. **Check scheduled tasks.** There's a `scheduled_tasks` table referenced by `build-dispatch-loop` (the `notify_scheduler` step does `UPDATE scheduled_tasks ... WHERE name = 'build-pipeline-trigger'`). Other rows there may indicate a triage task:
   ```sql
   SELECT name, last_completed_at, last_started_at
   FROM scheduled_tasks
   ORDER BY name;
   ```

5. **Read recent FAILED orchestrations on triage-shaped agents.** If a triage agent exists and is crashing, its orchestrations would be FAILED.

## What this is NOT

- **Not a kafka issue.** Kafka delivery works (manual triggers land).
- **Not a handler issue.** Image-build-handler works end-to-end when invoked directly.
- **Not a handler_agent column issue.** Items have correct `handler_agent='image-build-handler'`. Confirmed by inspection.
- **Not specific to needs_imagery.** 8 of our items are needs_imagery but other item types are also stuck in detected. Systemic.

## Workarounds currently in use

For loops blocked on dispatch: manual `kcat` trigger per work item. Pattern documented in debugging guide section 9 ("Trigger sent literal placeholder strings"). Use the psql-jsonb-builder form to avoid escape issues:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -t -A -c "
SELECT jsonb_build_object(
  'action', 'orchestrate',
  'config', jsonb_build_object('agent_type', 'image-build-handler'),
  'input_data', jsonb_build_object(
    'site_id', '<site_id>',
    'domain', '<domain>',
    'work_item_id', '<work_item_id>',
    'item_type', 'needs_imagery',
    'spec', spec::jsonb
  )
)::text
FROM site_work_items
WHERE id = '<work_item_id>';
" > /tmp/trigger.json
```

Then pipe `/tmp/trigger.json` to kcat with the standard headers.

## Bookkeeping consequence

Manual triggers do NOT update `site_work_items.status` automatically, even after `phase_2g_followup_mark_work_item_complete.sql` is applied. Reason: the chassis's manual-trigger path doesn't follow the same claim/release cycle as dispatch. The `mark_work_item_complete` step at the end of image-build-handler DOES update the work item when the workflow has `input_data.work_item_id` set — which is preserved through manual triggers. So:

- After fix deployed: manual trigger correctly marks work item complete via the action.
- Before fix deployed: required manual SQL UPDATE.

This works around the dispatch issue at the per-item level but doesn't address the systemic backlog of items stuck in `detected`.

## What we'd actually fix once dispatch is understood

If the hypothesis is correct (detected→triaged transition missing), three plausible fixes:

**A) Auto-triage on emission for low-risk types.** Discovery checks could insert at status='triaged' directly when the work doesn't need review. `needs_imagery` doesn't need human approval — the plan is locked, the prompt is fixed, the cost is bounded. Fix: change the INSERT in `check_unfulfilled_imagery_plan.go` to set `status='triaged'`. One-line change. Backfill existing rows with an UPDATE.

**B) Add a triage agent / step.** If review IS wanted (e.g., approve images before generation costs are incurred), build a triage layer that does the detected→triaged transition with whatever logic is appropriate.

**C) Adjust dispatch to claim 'detected' directly.** Bypass triage. Probably wrong — there's a reason the state exists.

A is the obvious fix for `needs_imagery` specifically, and probably for most discovery-check-emitted items. B is the broader architectural answer if review gates are wanted in general. C is a band-aid.

## Decision points pending

- **Should `needs_imagery` items auto-triage on emission?** Practical answer is probably yes (cost is bounded, prompt is plan-locked, deletion is cheap). Architectural answer depends on whether we want a human approval gate before image generation. Currently no such gate exists; the legacy `unfulfilled_image_prompt` items processed the same way without one. So consistency argues for auto-triage.
- **Bulk backfill existing rows or process them one-by-one?** Once auto-triage is in place for new emissions, the 8 existing items can be backfilled with `UPDATE site_work_items SET status='triaged', triaged_at=NOW() WHERE status='detected' AND item_type='needs_imagery'`. They'd then get claimed naturally by dispatch.
- **Audit other detected-stuck item types.** Once the imagery case is fixed, query for other types stuck in detected and decide which to auto-triage vs which need real review. Likely many.

## Cross-references

- `/PLAN_imagery_loop_closure.md` — broader plan; this dispatch issue listed under Open Items.
- `016_debugging_guide.md` section 9 — workarounds (manual trigger patterns).
- `phase_2g_followup_mark_work_item_complete.sql` and `phase_2g_followup_mark_work_item_failed.sql` — handler-side bookkeeping that closes the loop AFTER an item is processed; doesn't affect the dispatch claim step.

## Next concrete steps in priority order

1. Run query #1 (grep `load_work_items`) to confirm the filter. 2 minutes.
2. Run query #2 (counts by item_type) to see scale of the systemic problem. 1 minute.
3. Run query #3 (triage agent lookup) to see if a triage layer already exists. 1 minute.
4. Based on findings, decide between fix-A (auto-triage on emission) and fix-B (build triage layer).
5. Implement, backfill, watch one dispatch tick claim a needs_imagery item.
