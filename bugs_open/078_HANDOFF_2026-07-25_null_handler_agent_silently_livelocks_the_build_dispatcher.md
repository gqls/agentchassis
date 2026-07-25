# 078 — a single work item with `handler_agent IS NULL` silently livelocks the fleet build dispatcher

**Filed 2026-07-25** (bugfix-028 session, found while trying to rebuild one relojistas page).
**Status: OPEN.** Fleet-wide outage class, and cheap to cause: **one hand-written `INSERT`
that omits `handler_agent` stops builds on every site.**

Related but NOT the same as `bugs_open/029` (*hung spawns saturate the dispatch group*).
029's signature is orchestrations stuck in `AWAITING_RESPONSES`. **This one's signature is
the opposite: dispatch orchestrations COMPLETE, briskly and forever, having done nothing.**
Read both before diagnosing a stalled build queue.

## The mechanism

Two queries disagree about which items exist, and nothing reconciles them.

**1. Site selection** — `build-pipeline-trigger.find_dispatchable_site` (a
`query_database` step in the agent's workflow config, fires every 120s):

```sql
SELECT DISTINCT ON (wi.site_id) wi.site_id::text, s.domain
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE wi.status IN ('triaged','approved') AND wi.attempt_count < wi.max_attempts
  AND NOT EXISTS (SELECT 1 FROM site_work_items active
                  WHERE active.site_id = wi.site_id AND active.status='claimed')
ORDER BY wi.site_id, wi.priority ASC LIMIT 1
```

It says nothing about `handler_agent`. One `triaged` row makes a site dispatchable.

**2. Item loading** — `build-dispatch-loop.load_items` →
`LoadWorkItemsAction` (`platform/orchestration/actions/load_work_item_actions.go:547`).
Its SQL also says nothing about `handler_agent` — but the **row scan** does:

```go
var handlerAgent, status string          // :609  — plain string, not sql.NullString
...
err := rows.Scan(..., &handlerAgent, &status, ...)   // :616
if err != nil {
    logger.Warn("LoadWorkItemsAction: scan error", zap.Error(err))
    continue                              // :624  — row silently dropped
}
```

`handler_agent` is a **nullable** column. Scanning SQL NULL into `string` fails with
*"converting NULL to string is unsupported"*, so the row is **skipped by `continue`** —
logged at `Warn`, never surfaced, never counted, and the loop returns
`pending: {has_items:false, item_count:0, items:null}`.

**The livelock:** site selection counts the row, item loading refuses it, the loop completes
having claimed nothing, the row stays `triaged`, and 120 seconds later the trigger picks
**the same site again**. Forever.

And because selection is `ORDER BY wi.site_id … LIMIT 1` — one site per tick, ordered by an
arbitrary UUID — a NULL-handler row on a **low-`site_id`** site starves *every* site sorting
above it. This is not degradation; it is a total stop.

## Observed (2026-07-25, live)

`leopardessconsulting.co.uk` (`4851f6fc…`, the **lowest** site_id in the active set) held one
`page_rerender` row created 17:59 by another session (`created_by = operator:bugfix_023`,
summary *"Re-render who-we-help sections: prove migration 211 gating removes the 6 dead
case-study controls"*) with `handler_agent IS NULL`.

- **680 of 681** `page_rerender` rows fleet-wide carry `handler_agent = 'page-rerender'`.
  That one row was the only exception — a hand-written `INSERT` that omitted the column.
- Dispatch orchestrations for that site completed every ~20s, back to back, each with
  `item_count: 0`.
- **No work item completed anywhere on the fleet between 17:42 and at least 18:26**, while
  `webdesign.co.uk` sat on 95 `triaged` items and `scheduled_tasks.last_triggered_at`
  advanced on schedule the whole time. Nothing errored. Nothing alerted.

```sql
-- the detector: one row, and the whole queue is down
SELECT s.domain, swi.id, swi.item_type, swi.summary
FROM site_work_items swi JOIN sites s ON s.id = swi.site_id
WHERE swi.handler_agent IS NULL AND swi.status IN ('triaged','approved');
```

## Why it is so easy to cause

`insertWorkItem` (Go) always sets `handler_agent`. **Operators writing work items by hand do
not** — there is no `NOT NULL`, no default, no check, and no complaint at insert time. The
row looks perfectly normal in `psql` (NULL renders as blank, indistinguishable from `''` in
aligned output — `''` would work fine; NULL is fatal). `016b` §9 already warns that a
hand-written `INSERT` bypasses the Go-side guards in `insertWorkItem`; this is the sharpest
instance of that, because the cost lands on every other site rather than on your own row.

## Fix candidates

1. **Scan into `sql.NullString`** (`load_work_item_actions.go:609`) and treat NULL as `""`.
   One-line, removes the silent drop. The item then loads and fails *loudly* at
   `spawn_handler` (empty `agent_type_field`) against its own `attempt_count`, which is the
   correct blast radius: one row's problem stays one row's problem.
2. **Make the two queries agree.** Add `AND wi.handler_agent IS NOT NULL` to
   `find_dispatchable_site` so a site is never selected for work the loader will refuse.
   Config-only, live immediately — the fastest mitigation, and it converts a fleet stall
   into one stuck row.
3. **`ALTER TABLE site_work_items ALTER COLUMN handler_agent SET NOT NULL`** (after a
   backfill), or a `DEFAULT`. Closes the door at the source. Needs a check for legitimate
   NULL users first — `SELECT count(*) FROM site_work_items WHERE handler_agent IS NULL`
   was 1 fleet-wide at filing, so the backfill is trivial today.
4. **Make the loop's zero-item outcome legible.** `item_count: 0` on a site the selector
   just called dispatchable is a contradiction the platform should log at `Error`, not
   discover by hand. A `Warn` inside a `continue` is invisible at fleet scale.

Recommend **1 + 2 together** (defence at both ends), then 3 as the durable close.

## Interim repair applied (2026-07-25 18:2x)

I set the one offending row's handler to the value 680 of its 681 siblings carry:

```sql
UPDATE site_work_items SET handler_agent = 'page-rerender', updated_at = now()
WHERE id = '4ed13402-cc32-4f68-8fdc-84b38da8ced9'
  AND handler_agent IS NULL AND item_type = 'page_rerender';
```

This restores the originating session's evident intent rather than overriding it (their item
is a page re-render; `page-rerender` is an active agent type). It is reversible.
**`operator:bugfix_023` — this is your row; nothing else about it was touched.**

> **IMPORTANT, and the reason this case stays OPEN beyond the code fix:** after that repair
> the loader query returns the row correctly (verified by running it verbatim), **but the
> queue had still not drained as of 18:26** — no new dispatch orchestration was created after
> 18:23:24 even though the trigger fired at 18:24:58. So the NULL scan is **a** cause, proven
> and repaired, but **not the only thing wrong with the build queue right now.** Do not read
> this case as "queue stall explained". The residual may be `bugs_open/029`, may be
> `bugs_open/030`, or may be a third thing. **[UNVERIFIED]** — I did not diagnose it; I was
> closing an unrelated case and stopped at the boundary.

## How to verify a fix

1. Insert a `triaged`, `pipeline='build'` work item with `handler_agent` NULL on a site with
   a **low** `site_id`.
2. Watch `orchestration_states` for `build-dispatch-loop` runs: pre-fix they complete every
   ~20s with `collected_data->'pending'->>'item_count' = 0`, forever, and no other site's
   items progress.
3. Post-fix: either the site is not selected (candidate 2), or the item loads and fails on
   its own `attempt_count` (candidate 1). **Either way, other sites' builds must keep
   running** — that is the actual acceptance test, not the fate of the bad row.

## Related

- `bugs_open/029` — builds halt fleet-wide via *hung* spawns. Different signature (stuck vs
  completing), same blast radius. Check both when the queue is dead.
- `bugs_open/030` — dispatch queue serialisation / single-consumer backlog.
- `bugs_open/028` (`028-page-build-noop`) — where this was found; its last page could not be
  rebuilt because of this stall.
- `016b` §9 *"The build dispatcher picks ONE site per tick ordered by `site_id`"* — the
  ordering fact that turns one bad row into a fleet outage.
