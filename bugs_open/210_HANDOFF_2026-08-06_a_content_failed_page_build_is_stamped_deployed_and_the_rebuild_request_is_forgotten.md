# 210 — a content-failed page build is stamped `deployed`, so the rebuild request is silently forgotten

**Filed 2026-08-06** by the `bugfix_208_owned_page_commit_before_guard` lane, as the follow-up
that bug's fix deliberately did **not** include. **OPEN, unowned.** Severity: medium — silent
loss of a build request, no data destroyed.

> **Filed as a code-derived finding, and the label is doing work.** The mechanism below is read
> first-hand from the live code and is not inferred. What is **NOT** established is how often it
> fires in production, and the obvious way to measure it is confounded — see "How to measure
> this properly", which is the main thing this file is for. Per the owner ruling of 2026-07-31 I
> did **not** run `090`: this is a local, single-function claim with no cross-cutting root cause,
> not the class that ruling covers. If a fixing thread wants to assert anything wider than the
> paragraph below, that changes.

## The mechanism

`UpdatePageStatusAction` (`platform/orchestration/actions/v3_site_actions.go`) has two guards
before it stamps a page `deployed`, both added to stop exactly this class:

- `pageHasComponents` — refuse if the page has no rendered components;
- `pageSectionShortfall` — refuse if fewer components were rendered than the plan asked for.

Neither sees a **rebuild whose content generation failed**, because both are satisfied by the
*previous* build's output:

1. `assemble_page` skips (its `checkUpstreamContentFailure` branch, or an empty content field) and
   returns `{"html":"", "skipped":true}`.
2. `deploy_page` (`git_commit`) honours the skip via `checkUpstreamSkipped` and commits nothing.
   **Correct — the live page is untouched.**
3. `update_page_status` runs with `status: "deployed"`. `pageHasComponents` is **true** (the prior
   build's components are still there) and `pageSectionShortfall` finds **no shortfall** (they
   still match the plan), so both guards pass.
4. The row is stamped `build_status='deployed'`, `deployed_at=NOW()`, and
   `built_from_plan_version = the site's current plan`.

**Net:** a page that asked to be rebuilt, wasn't, and now says it was. It leaves `needs_rebuild`,
so no future selection picks it up; and the `built_from_plan_version` stamp makes
`ReconcileSitePlanAction`'s `decideEmit` return `skip_built`, so the reconciler will not revisit
it either. The request is not failed, not retried and not queued — it is **forgotten**, and the
page keeps serving its old content while the fleet believes it is current.

Same family as `bugs_open/208` (a status claim for work that did not happen) and as the
`bugs_open/099` shape (*a FAILED step shows COMPLETED with `error` NULL*).

## Why 208's fix deliberately stopped short of this

208 added exactly this guard, but keyed to **its own** skip marker
(`strings.Contains(skipReason, ownedPageSkipReasonPrefix)`), so it fires only for an owned page
the ownership guard refused. Widening it to *any* assembly skip is the fix for this bug — and it
is a **fleet-wide retry-behaviour change**: pages that are currently stamped-and-forgotten would
instead stay at `needs_rebuild` and be **re-selected on every subsequent run**. On a page whose
content generation fails deterministically (a bad spec, an unsatisfiable component, a quota wall
— see `bugs_open/202`) that is an unbounded retry loop paying for an LLM call each time.

That trade-off is the whole of this bug. It is not a one-line change wearing a disguise.

`TestUpdatePageStatus_OrdinarySkipStillStamps` and `TestSavePageSections_OrdinarySkipIsNotClaimed`
pin the current narrow scope, so **widening it will fail two named tests** — deliberately, so the
change is a decision rather than a side effect. Update them in the same commit.

## How to measure this properly — read this before running the obvious query

The tempting proxy is "pages stamped `deployed` after their newest `page_components` write":

```sql
WITH pc AS (SELECT page_id, max(updated_at) AS newest FROM page_components GROUP BY page_id)
SELECT s.domain, p.name, p.deployed_at, pc.newest
FROM pages p JOIN sites s ON s.id=p.site_id JOIN pc ON pc.page_id=p.id
WHERE p.build_status='deployed' AND p.deployed_at > pc.newest + interval '2 hours';
```

**I ran it, got 15+ hits with gaps of thousands of hours, and it is NOT evidence of this bug.**
`page-rerender` legitimately re-assembles and re-commits a page from its **existing**
`page_components` without writing any, so a large `deployed_at > newest_component` gap is the
*expected* signature of a healthy rerender. The proxy cannot separate the two, and every hit I
looked at is consistent with the innocent explanation. Recorded as a rejected measurement so the
next thread does not spend the same hour concluding the same wrong thing.

What would actually discriminate, in rough order of cost:

1. **Catch it live.** `agent_error_log` / the orchestration's own step records for a run where
   `assembled_page.skipped=true` **and** `update_page_status` reported `updated:true` in the same
   iteration. Terminal `orchestration_states` are reaped at ~24h, so this must be caught inside
   the window — set it up before you need it.
2. **A cheap permanent signal:** have `update_page_status` record the skip-vs-stamp combination
   when it happens (an `agent_error_log` row, as 208's guard now does for its fail-open window).
   That converts an unmeasurable claim into a count, and it is a smaller change than the fix.
3. **`build_status` history**, if anything retains it — a page going `needs_rebuild → deployed`
   with no `page_components` write and no site-repo commit between the two timestamps is the
   unambiguous shape. Check whether any audit table keeps it before assuming it does.

## Fix candidates, ordered by what closes the door

1. **Refuse the stamp on any assembly skip, and requeue explicitly.** Leave `build_status` at
   `needs_rebuild` **and** file a work item recording the failed attempt, so the retry is bounded
   by the existing two-strike machinery in `insertWorkItem` (`load_work_item_actions.go`: two
   terminal predecessors → `unresolved`, and a <3h predecessor is suppressed within-cycle) rather
   than by nothing. This is the candidate that answers the retry-loop objection instead of
   inheriting it — the bounding mechanism already exists and is used elsewhere.
2. **Stamp a distinct status** (`rebuild_failed`) rather than `deployed` or `needs_rebuild`.
   Honest and un-forgettable, but it is a new value in a status vocabulary several queries
   enumerate — check `workItemTerminalStatuses`' sibling lists and every `build_status IN (...)`
   before proposing it, and expect the RFC question about whether it changes what the shared
   vocabulary guarantees.
3. **Refuse the stamp and do nothing else.** One line, and it is the unbounded retry loop. Only
   acceptable with a measured answer to "how many pages fail content generation repeatedly?"
4. **Leave it and document it.** Defensible only if measurement (above) shows it effectively never
   fires — which nobody knows today.

## Related

- `bugs_open/208` — same family, and the reason this is scoped out; its guard is the template.
- `bugs_open/099` — a FAILED step reporting COMPLETED with `error` NULL.
- `bugs_open/202` — Gemini quota 429s blocking fleet page builds, i.e. a live source of exactly
  the deterministic content-generation failures that make candidate 3 dangerous.
- `bugs_closed/037` — `needs_rebuild` as an under-protected state.
- PBP-036 (concept register) — the ownership guard whose scoping decision created this file.
