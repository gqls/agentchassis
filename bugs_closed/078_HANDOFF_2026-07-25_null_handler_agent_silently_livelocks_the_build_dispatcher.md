# 078 — a single work item with `handler_agent IS NULL` silently livelocks the fleet build dispatcher

**Filed 2026-07-25** (bugfix-028 session, found while trying to rebuild one relojistas page).
**Status: CLOSED 2026-07-26 — fixed and LIVE.** The column is now `NOT NULL DEFAULT ''`
(migration `217`, applied and recorded), so the state this case describes **cannot be
represented any more**. See "Resolution" at the foot of this file.

**It recurred before it was fixed.** Filed 07-25 against `leopardessconsulting.co.uk`;
on 07-26 a *different* session made the identical hand-written `INSERT` against
`gaswholesalers.com` and took the fleet down again for 42 minutes. Two independent
sessions, one day apart, same one-column omission. That recurrence is the whole
argument for closing the door in the schema rather than asking operators to remember
a column.

Fleet-wide outage class, and cheap to cause: **one hand-written `INSERT`
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
> queue still had not drained as of 18:31** — **no dispatch orchestration has been created
> since 18:23:25** even though `build-pipeline-trigger.last_triggered_at` kept advancing
> (18:24:58, 18:29:59). So the NULL scan is **a** cause, proven and repaired, but **not the
> only thing wrong with the build queue.** Do not read this case as "queue stall explained".

### Second cause — NOT diagnosed, but here is the evidence already gathered

State at 18:31, with two sites holding eligible work (`leopardess` 1 item, `webdesign` 95;
both `pipeline='build'`, `status='triaged'`, `attempt_count 0 < 3`, `approval_mode auto`,
neither site `locked_at`) — so `find_dispatchable_site` *should* return a site and
`spawn_dispatch` *should* run. It does not. **The trigger fires and produces nothing.**

The tell is in `scheduled_tasks`: when a dispatch actually runs, the loop's
`notify_scheduler` step does `UPDATE scheduled_tasks SET last_completed_at = NOW()`, so
`last_completed_at` lands *after* `last_triggered_at` (18:19:59 → 18:21:27). Since 18:24 the
two are **identical** on every tick — the idle path (`complete_idle` /
`notify_scheduler_idle`), meaning `check_has_site` saw no site.

**Ruled out** (each checked, each negative — this is the value of the note):

- **Not `029` (hung spawns saturating the `dispatch` group).** `SELECT status, count(*) FROM
  orchestration_states WHERE status IN ('AWAITING_RESPONSES','EXECUTING_STEP')` returns
  **zero** `AWAITING_RESPONSES` and 2 `EXECUTING_STEP`, neither a dispatch agent. Nothing is
  stuck waiting on a child.
- **Not pod-pool exhaustion.** `dispatch` group `max_concurrent` is 8; there are **2 running**
  `build-dispatch-loop` pods (plus 5 `Completed` not yet reaped) and **0**
  `build-pipeline-trigger` pods.
- **Not site locks.** `locked_at IS NULL` on all four active sites.
- **Not item ineligibility.** The loader's SQL, run verbatim against leopardess, returns the
  row.

So the failure is between "eligible rows exist" and "`find_dispatchable_site` returns a
site". **Start there** — run that query verbatim as the agent would, and check whether the
`query_database` step is returning its result into `dispatchable` at all (an
`output_format: "object"` step that yields no row leaves the field unset, and
`check_has_site` then reads a missing field, which is the same as false). Note `orchestration_states`
is pruned at 24h, so gather evidence before it ages out.

**[UNVERIFIED]** — I did not diagnose this; I was closing an unrelated case
(`028-page-build-noop`) and stopped at the boundary rather than assert a mechanism I had not
tested. The negative results above are measured, not inferred.

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

---

# Resolution — 2026-07-26 (bugfix_078 session)

## The second occurrence, measured

I picked this ticket up and found **the fleet already down again**, by the same
mechanism, caused by a session unrelated to the first:

| | measurement (live, 2026-07-26) |
|---|---|
| offending row | `709f0338-8b5a-443d-9bf0-ce11d0b9418e` on `gaswholesalers.com`, `page_rerender`, `triaged`, `handler_agent IS NULL`, `created_by = operator:bugfix_049`, created 16:42:41 |
| ordering | `gaswholesalers` `5fe15466…` sorts **below** `webdesign.co.uk` `6b49db8e…`, so it was picked first every tick |
| trigger | selected `gaswholesalers.com` at 17:32, 17:35, 17:37, 17:40, 17:42 — **every** tick |
| loop | `item_count: 0`, `has_items: false` on **every** run |
| throughput | last completion **17:00:32**, found at 17:42 — **42 minutes of zero completions fleet-wide**, 2 sites holding `triaged` work, 0 claims |

**Causal proof, before/after on the live system.** Setting that one row's handler
(`page-rerender`, the value 704 of its 705 siblings carry, registered and active) moved
`build-dispatch-loop` from `item_count: 0, 0, 0` to **`1`** at 17:45:30, with a
completion landing immediately and `webdesign.co.uk` — starved throughout — claimed and
building by 17:55. Nothing else was changed.

`operator:bugfix_049` — that was your row, repaired the same way the 07-25 one was, to
restore its evident intent. Nothing else about it was touched.

## What was fixed

**1. Migration `217_site_work_items_handler_agent_not_null.sql` — the durable close,
LIVE (applied out of band, `--record-only` in the ledger).**

```sql
UPDATE site_work_items SET handler_agent = '' WHERE handler_agent IS NULL;  -- 129 rows
ALTER TABLE site_work_items
  ALTER COLUMN handler_agent SET DEFAULT '',
  ALTER COLUMN handler_agent SET NOT NULL;
```

This is what closes the case, and it needed no image roll. `''` and NULL were already
two spellings of one deliberate state ("no handler" — flag-only / human-review items,
`check_image_url_404.go:109`, `lock_helpers.go:117`,
`render_site_components_action.go:693`); 169 rows already used `''` against 121 using
NULL. **Only NULL is fatal, and the whole outage turns on that asymmetry:** `''` scans
fine, so the item *loads*, gets *claimed*, and the site is then mutex'd by its own
claimed row — every other site keeps building. NULL is dropped before any of that can
happen.

*Safety, each checked live before applying:* the 129 rewritten rows were all
`needs_human_review` (121) or `complete` (8) — **zero** `triaged`/`approved`, so the
backfill could not change what the dispatcher sees. And **no writer emits an explicit
NULL**: `insertWorkItem` passes a Go `string`, and the three paths that actually created
the NULL rows (`resolve_internal_links_action.go:257`, `plan_sections_action.go:1862`,
`reconcile_site_plan_action.go:255`) **omit the column from the INSERT entirely** — so
they pick up the new default and keep working untouched. That is the point: the fix
lands without editing a single one of them.

**2. `load_work_item_actions.go` — never silently drop a row. [INERT until the next roll]**
`handler_agent` now scans into `sql.NullString`; the drop path logs at `Error` (not
`Warn`) naming the site, counts drops, surfaces `rows_dropped` in the action's output,
and raises an explicit `LIVELOCK RISK` error when every candidate row was dropped. This
also covers a scan failure on *any* column, not just this one.

**3. `claim_work_item_action.go` — an unroutable item is blocked, not dispatched.
[INERT until the next roll]** An empty/NULL handler is the degenerate case of "handler
not registered" and now takes that same existing exit (`blocked`, claim released, clear
error), instead of reaching `spawn_agent` and failing three times under a message that
names the wrong problem — see the induced-fault evidence below.

## Verification — both branches induced live, not inferred

| test | result |
|---|---|
| **A.** The outage's own INSERT shape (column omitted), on a site | lands `handler_agent = ''`, `is_null = f` — **the fatal state is unrepresentable** |
| **B.** An INSERT writing an explicit `NULL` | **rejected at insert time**: `null value in column "handler_agent" … violates not-null constraint`. Fails loudly against its author instead of silently against the fleet |
| **C.** Arm a `''` row `triaged` on `pool-savings-investing.internal` — the **globally lowest `site_id`**, i.e. the exact worst case that starved the fleet twice | trigger selected it 17:57:44; `build-dispatch-loop` returned **`item_count: 1`** (pre-fix this was `0`), the row was **claimed** at 17:58:02 — so the site self-mutexes and others proceed — and the failure landed on **its own `attempt_count` (0→1)**, bounded by `max_attempts` in both the selector's and the loader's SQL |

Test C is the acceptance test the case file asked for, and it passed on the worst-case
site rather than a convenient one. Migration `217` additionally carries an in-transaction
probe that re-runs test A and `RAISE EXCEPTION`s if the default ever stops working, so a
future re-application cannot quietly regress it.

Test C also *justifies* fix 3: the recorded failure was

```
spawn_agent: configuration extraction failed … agent_type is required
(provide 'agent_type' or 'agent_type_field')
```

which blames the step's configuration — the field *is* provided, it just resolves to
empty. Nothing in that message says "this work item has no handler". After the roll it
is blocked on attempt 1 with `No handler_agent set — item cannot be routed to any agent`.

**Not yet verified, and stated as such:** fixes 2 and 3 are Go and are **inert until the
next image roll**. They are defence-in-depth behind the migration, which is what actually
extinguishes the defect. Post-roll marker (assert the *new* string appears — the old
behaviour left no string to disappear):
`kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "LIVELOCK RISK"'`

## Deliberately NOT done: fix candidate 2

Candidate 2 (add `handler_agent IS NOT NULL` to `find_dispatchable_site`) was **not**
applied. Another live session owns that query — its migration
`213_dispatch_gate_matches_dispatcher.sql` (`bugs_open/029`, uncommitted and unapplied
in the tree at the time of writing) rewrites it wholesale. Editing the same JSONB path
underneath a concurrent thread is exactly the collision CLAUDE.md warns about, and after
`217` the clause is redundant: the column can no longer be NULL.

## Handed on, not fixed here: the same livelock has a second door

`find_dispatchable_site` does **not** check `depends_on`; the loader **does**
(`load_work_item_actions.go:562-571`). So a site whose only `triaged` item is
dependency-blocked is selected forever and loads nothing — **identical mechanism, and
this fix does not touch it.** The `bugfix_029` dispatch-gate PLAN lists this as its
watchdog's "KNOWN BENIGN CASE"; it is not benign, it is this bug with a different
predicate.

**[UNOBSERVED] today** — 0 `triaged`/`approved` rows carry `depends_on` (3 rows ever
have). Recorded for the owning lane rather than fixed in it.

## The undiagnosed "second cause" — transferred, not solved

> **CORRECTED/RESOLVED 2026-07-26:** the caution above ("the queue still had not drained
> as of 18:31 … do not read this case as *queue stall explained*") was right to be
> written and is now discharged **by transfer, not by proof.** The `bugfix_029`
> dispatch-gate thread has since measured, live, that the trigger's gate `pre_query` and
> `find_dispatchable_site` **disagree**: the gate ignores the claimed-item mutex the
> dispatcher enforces, so it reported 2 pending sites while the dispatcher could dispatch
> nothing, and the trigger fired every 120s onto `complete_idle`. That is the same
> "trigger fires and produces nothing" symptom recorded here at 18:23–18:31, and it is
> their fix (migration `213`), not this one. **[INFERRED]** — I did not reproduce the
> 07-25 window, and I am not asserting the two are the same event; I am naming where the
> evidence now lives so this case does not stay open on another lane's work.

The NULL mechanism itself **is** proven, fixed, and extinguished — twice reproduced,
once by induced fault after the fix.
