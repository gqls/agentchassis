# FOCUS: page-build-handler silently completes failed work

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
