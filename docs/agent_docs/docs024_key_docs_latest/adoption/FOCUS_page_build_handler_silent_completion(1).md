# FOCUS: page-build-handler silently completes failed work

**Status**: original (2026-04-19) characterisation below. **2026-06-09 update: modes 1–3 are now resolved in current code; one residual gap remains (`complete_work_item` clobber). See the update section directly below.**

---

## 2026-06-09 UPDATE — re-verified against current code (mostly resolved)

Re-checked each mode against the live `claimed-item-timeout` reaper, the build-dispatch-loop def, `validate_page_content.go`, and `load_work_item_actions.go`. The April characterisation is largely stale — the positive-evidence fix has been implemented:

- **Mode 1 (reaper lost-response auto-complete): RESOLVED.** The `claimed-item-timeout` scheduled task now auto-completes a claimed item only with **positive evidence**: a `completed_by_evidence` CTE requiring `page_components` (`component_id` + non-empty `rendered_html` + `updated_at > claimed_at`) for `needs_content_page`, `pages.deployed_at > claimed_at` for `page_rerender`, and a head-slot update for `needs_design`. No evidence → it does not complete. (A prior false-positive — gamesdesign homepage with `build_status='deployed'` but zero components — is explicitly hardened against via the `page_components` + `updated_at > claimed_at` checks.)
- **Mode 3 (claim-timeout → complete): RESOLVED.** The reaper's `reset` CTE moves stuck claimed items (>40 min, no evidence) to `triaged` (or `failed` at `attempt_count + 1 >= max_attempts`) with `attempt_count` incremented. "Claim timed out — handler pod likely died" now labels a RESET, not a completion.
- **Mode 2 (validate_content routed to complete): RESOLVED.** `ValidatePageContentAction` returns an error on blockers; the page-build-handler def routes `validate_content` `error_step → mark_needs_review` (`fail_work_item` with `status_override='needs_human_review'`). Validation failures are flagged, not completed.
- **Positive-evidence helpers:** implemented inline in the reaper SQL rather than as separate Go functions — same effect.
- **Monitor (suggested step 5):** the detection query returns **0 rows** on the live site — no current silent-completion residue.

### Residual gap (the one thing still open): `complete_work_item` is unguarded
`CompleteWorkItemAction` (`load_work_item_actions.go`) does `UPDATE site_work_items SET status='complete' … WHERE id=$1` with **no status guard**. The build-dispatch-loop's `mark_complete` step calls it on every successful handler saga, and page-build-handler's `complete_error` is a SUCCESS-labelled `complete_workflow`. Two consequences:
1. **Clobber.** A handler that flagged its item `needs_human_review` (the existing `mark_needs_review`, and the planned `mark_no_sections`/S2) is re-stamped `complete` by `mark_complete` immediately afterward. Evidence chain: the 2026-06-06 skinner-box retry hit `complete_error` on a sectionless page and ended `status='complete'` with no `page_components` — confirming `complete_error → mark_complete` fires; combined with the unconditional UPDATE, a just-set flag is overwritten. (No direct sample yet because `mark_needs_review` has never fired — zero validation failures to date.)
2. **Dispatch-level silent completion.** Any page-build path ending at `complete_error` (deploy fail, save fail, content skip) returns saga-success → the item is marked `complete`. The reaper cannot catch these (item is already `complete`, not `claimed`); only the monitor query can.

**Fix A (applied 2026-06-09): guard `complete_work_item`.** `… WHERE id=$1 AND status NOT IN ('needs_human_review','failed','unresolved','rejected','wont_fix','verified','blocked')`, returning `completed = rows>0` and logging when skipped. No-op on the normal `claimed→complete` path; preserves deliberate flags. This is the "handler returned explicit success" case of the completion rule, so positive evidence isn't needed here — it just must stop clobbering flags. **Prerequisite for S2 (`mark_no_sections`) to be effective.**

**Fix B (deferred, low urgency given monitor=0): `complete_error` semantics.** `complete_error` shouldn't read as success for genuine-failure paths (deploy/save fail). Either flag the item before completing, or stop labelling `complete_error` as a success completion. Lower priority while no residue exists and the reaper handles lost/timed-out claims.

---

## ORIGINAL (2026-04-19) characterisation — retained for history

**Status**: observed and characterised; not yet fixed. Captured from the 2026-04-19 `install-theme-nav-specs` handoff (parallel chat).

## The pattern

`page-build-handler` ends work items with `status = 'complete'` in three distinct scenarios that all share a common symptom: the handler didn't actually do what the item asked for, but the item gets marked complete anyway. Downstream consumers see "item completed successfully" and don't retry or escalate, so the broken state sticks.

The three failure modes:

### 1. Reaper auto-completion on lost responses

**Error field value**: `"Auto-completed: work verified done despite lost response"`

**Cause**: When the orchestration response is lost (pod died, network blip, Kafka delivery failure), the reaper eventually fires and tries to decide whether the work is actually done. It uses weak evidence — some combination of page record state, orchestration state, elapsed time — but doesn't check for positive evidence of completion like `page_components` rows or git commits.

**Result**: items get marked complete on the hope that the work finished despite the missing response. Sometimes true; often false.

### 2. validate_content failures routed inconsistently

**Error field value**: `"validate_content failed: <N> blockers, <M> errors"`

**Cause**: The content validator (`validate_page_content`) detects placeholders, cross-site contamination, unrendered templates, etc. In some sites' orchestrations this routes to `needs_human_review` (correct). For other sites — including gamedesign.uk observed directly — it lands on `complete`. The routing is inconsistent, meaning the same validation result produces different terminal states depending on context that isn't clearly documented.

**Result**: sites with real content problems get flagged as done and deployed anyway.

### 3. 40-minute blind-reaper completion on claim timeouts

**Error field value**: `"Claim timed out — handler pod likely died"`

**Cause**: A handler claims a work item but never releases it — the pod dies, OOMs, deadlocks, whatever. After 40 minutes the blind reaper fires. Its current behaviour is to mark the item `complete`. The correct behaviour for a claim timeout should be to reset to `triaged` so the next dispatch cycle can retry it.

**Result**: items lost to pod failures are marked complete rather than retried. The site never gets the work the item was supposed to produce.

## Why this matters

All three modes share one architectural flaw: **"we're done trying" is being treated as "the work is done."** These are not the same. A pipeline that can't distinguish "succeeded" from "gave up" will silently leak broken state into production.

The consequence is that other handlers that depend on page-build-handler's output (rerender, image-build-handler, populate-nav) see `complete` status and assume they have something to work with. When they don't, their own failures become downstream symptoms of the silent completion, and debugging requires tracing back through multiple handler chains.

## What a proper fix looks like

The rule should be: **complete an item only when (a) the handler returned an explicit success response, OR (b) positive evidence exists in the database that the work was done.**

Positive-evidence checks for page-build-handler specifically:
- `page_components` rows exist for this page_id with non-null `component_id` and `rendered_html`
- `pages.build_status = 'deployed'` for this page
- A git commit has happened for this page's rendered HTML

Without at least one of these, the item should not move to `complete`. Options for the alternative status:
- `needs_investigation` — new status class, handled by a dedicated agent or HITL
- `failed` — the simplest option; retry-able
- `needs_human_review` — reuses existing HITL machinery

Per-failure-mode handling:
- **Mode 1 (lost response)**: either confirm with positive evidence, or move to `needs_investigation`
- **Mode 2 (validate_content failure)**: route ALL cases to `needs_human_review` consistently — never to `complete`
- **Mode 3 (claim timeout)**: reset to `triaged` with `attempt_count += 1`, not mark complete

## Why this isn't fixed yet

Three reasons:

1. **Scope**. Each of the three modes touches a different part of the pipeline (reaper logic, validate_content routing, claim-timeout logic). A proper fix is probably a small project in itself.

2. **The three modes are symptoms, not the root cause**. The root cause is that `complete` is being used as "terminal state" rather than "work succeeded." Fixing the three modes individually patches symptoms without addressing the muddled semantics of the `status` column.

3. **Priority**. The composition work (025, 026) was more urgent because it blocked sites from rendering correctly at all. The silent-completion issue produces broken sites that look fine — less obvious damage but more insidious. Worth fixing, not worth blocking other work for.

## Related in project knowledge

- `HANDOFF_2026-04-19_install-theme-nav-specs_1_.md` — original observations in "Page-build-handler silently completes failed work" section
- `HANDOFF_2026-04-18_enrichment_bug_diagnosed_and_patched.md` — the enrichment bug was a similar class of silent-looking-ok failure (component_id NULL didn't break anything loud, just degraded output)
- `016_debugging_guide_v2.md` — could usefully get a new symptom-table row for this class once the fix is in

## Suggested work order when this is picked up

1. **Inventory the reaper logic** — find the code paths that set `status = 'complete'` from the reaper. There may be more than three.
2. **Add positive-evidence helpers** — small Go functions like `pageHasComponents(ctx, db, pageID)`, `pageIsDeployed(ctx, db, pageID)`. Reusable across handlers.
3. **Patch the three known modes** — one at a time, each with a test case.
4. **Audit other handlers** — page-build-handler is the one that surfaced this, but component-creator, rerender-pages, webdesign-agent, and page-content-writer all go through similar paths. Same failure pattern is likely present elsewhere.
5. **Add monitoring** — a query that finds work items marked `complete` in the last N hours where no positive evidence exists. Run nightly, raise an alert if non-zero.

## Workaround until fixed

The parallel chat's suggested triage:

```sql
-- Find complete items with no corresponding page_components
SELECT wi.id, wi.site_id, wi.item_type,
       wi.completed_at, LEFT(wi.error, 100) AS err
  FROM site_work_items wi
  JOIN pages p ON p.site_id = wi.site_id
              AND p.name = wi.spec ->> 'page_name'
 WHERE wi.status = 'complete'
   AND wi.item_type IN ('needs_content_page', 'content_rewrite')
   AND wi.completed_at > NOW() - INTERVAL '7 days'
   AND NOT EXISTS (
       SELECT 1 FROM page_components pc
        WHERE pc.page_id = p.id
          AND pc.component_id IS NOT NULL
   );
```

Requeue any rows this returns by setting `status = 'triaged'` and `attempt_count = 0`. Not a fix — just a way to catch and retry the stuck work items until the architectural fix is in.
