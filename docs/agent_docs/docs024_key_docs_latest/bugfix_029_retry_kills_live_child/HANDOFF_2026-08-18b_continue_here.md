# HANDOFF — 2026-08-18b — `bugs_open/029`, continue here

**Supersedes `HANDOFF_2026-08-18_continue_here.md`** (bannered). Read this, then
`NOTES_retry_kills_live_child.md` (technical log, newest at the bottom).
`README_where_we_are.md` is the owner's plain-prose log — append, never rewrite.

## State in one line

**Part A is fixed, council-APPROVED, LIVE on `v1.0.1309` and now BEHAVIOURALLY PROVEN.
`bugs_open/029` is still OPEN** — the wedge it is named for now has a sharp, measured signature
and no explanation.

## The three things you must not get wrong

1. **029's title is wrong and so is its second framing.** The concurrency-group story was
   refuted in-file 2026-07-21. The "retry kills a still-working child" story was withdrawn on
   2026-08-18 — the child is already dead ~7–10 min before the replay arrives. **Do not
   re-import either.**
2. **Part A being PROVEN does not close 029.** It fixed how long a replayed request waits. It
   did not explain what freezes a build job, and it did not touch it.
3. **`03:00` means two opposite things.** Old inversion = a step declaring 1800s+. New code = a
   step declaring **nothing** (the correct default). Check the declaration, and check
   `retry_version`, before concluding anything from a window.

## What is DONE (was "the one open task")

**Proven 2026-08-18 18:28:21Z, 2h43m after the roll:**

```
step_name     | rv | sent_at             | timeout_at          | window   | status
call_dispatch |  1 | 2026-08-18 18:28:21 | 2026-08-18 18:43:21 | 00:15:00 | processed
```

`call_dispatch` declares 900s; the retry got **15:00, exactly**. `status = processed` is the
load-bearing half — the child answered **inside** the window, where the old code expires it at
18:33:21 and escalates to rv2.

**One row is enough because the old distribution has zero variance:** all pre-roll
`call_dispatch` retries, all history — rv1 n=56, rv2 n=30, rv3 n=133, **219 retries, none above
05:00**. **Positive control passes:** rv0 unchanged (`call_dispatch` 15:00 n=87, iter handlers
20:00 n=89, `call_diagnoser` 30:00). RSH-010, the index row and the bug file are all updated.

## THE OPEN TASK: the wedge, which now has a signature

**What kills the parent's continuation after a spawn handshake succeeds?** Measured across the
**entire** wedge population (18 rows), three properties hold:

1. **All 18 are `build-dispatch-loop` frozen at a `process_item_iter_N_spawn_handler` step**, and
   `last_activity - created_at` is **25:14–25:22 in 17 of 18**. A lifetime uniform to eight
   seconds is a **timer**, not a hang.
2. **18 of 18 are entered from the ERROR path** — the *preceding* iteration's
   `process_item_iter_(N-1)_call_handler` is terminal in status `error`. Never the happy path.
3. **The final spawn step is registered TWICE in 17 of 18** (gap up to 9m37s), and the parent's
   **last state write lands 9–16 seconds after that final send, in all 18.**

`iter_N_call_handler` is **never registered**, so the row holds no `waiting` awaited request and
is invisible to everything but the 4-hour reaper.

`[INFERRED]` and worth testing, not asserting: 25:15 ≈ the 1200s `call_handler` declaration +
the old 300s retry cap. If that is right, Part A makes the *entry condition* rarer without
touching the freeze. **Do not let that arithmetic become a claim the wedge is fixed.**

Candidate from the previous session, still `[UNVERIFIED]`: `cleanupExpiredAwaitedRequests`
(`coordinator.go:4313`) runs recovery under a shared **60-second** context — but the fast-path
timer uses `context.Background()`, so it cannot be the whole story. `handleSpawnRetry`
(`coordinator.go:1649`) is a fresh candidate for the **duplicate** registration.

### The 090 is authored and seeded but NOT dispatched — and why

- `090`'s coverage check refuses on four rows that are all **one `failed` item from 2026-08-12**.
  `failed` is not in the exclusion list (`NOT IN ('complete','cancelled','rejected')`), so a
  terminal item reads as live coverage. **`FORCE=1` is the documented override and applies here.**
- **The real blocker: the diagnosis reads `origin`, and `origin` does not have this fix.**
  `retryWindow` is on **no remote branch**; `f0117fb8b` (the commit the live binary was built
  from) is on none; local HEAD is **233 commits ahead**, all dated 2026-08-18, origin's tip being
  12:39 BST. A run against that tree reads the OLD capping block in `handleRecoverableError` and
  has every reason to name the truncation as the cause — **a correct answer about the wrong
  tree, arriving with citations.**
- **Routed to the owner as a push decision.** Publishing 233 commits belonging to a dozen
  concurrent lanes is not one session's call, and forward-only means it cannot be undone.
- The full symptom text is in `NOTES` (2026-08-18 17:05Z section) and needs only the tree
  question settled. Seed scope used: `coordinator.go` at `continueExecution`, `handleSpawnRetry`,
  `createContinuationContext`, `handleRecoverableError`.

> ⚠ **THE WEDGE EVIDENCE EXPIRES DURING 2026-08-19.** `orchestration_states` retains **~26
> hours**. All 18 instances are from 08-17. A run filed after they age out reads a table with no
> instances in it. A dump is in this lane's scratch, but the diagnosis loop reads the live DB.

## Leads: one CLOSED, one unchanged

- ~~**The INITIAL wait may have its own conversion gap.**~~ **REFUTED 2026-08-18.**
  `call_diagnoser` rv0 is **30:00 in 29 of 29 rows** across its whole history; every 180s reading
  sits at rv1/rv2/rv3, i.e. the retry path Part A already fixed. Generalised: **18 agent/step
  pairs joined to their own agent's declaration, 18 honoured, zero mismatches.** Do not re-open
  on a single 180s row — check `retry_version` first.
- **The dormant twin stays unfixed, deliberately.** `helpers.go retryTimedOutRequest` carries a
  hard 30s; `NewOrchestratorHelper` has ZERO callers. **Scoped claim: no caller in THIS repo.**
  Re-run the grep rather than trusting this line.

## Traps — the previous handoff's table still stands, plus these

| trap | the check |
|---|---|
| **A step name is not a key.** `call_handler` is 2100 for `diagnose-dispatch-loop` and 1200 for `report-dispatch-loop` | join declarations on **(agent, step)**; an aggregate across agents fabricates truncations that do not exist (it did, here, today) |
| **`retry_version=0` is a SURVIVORSHIP filter.** `awaited_requests` is `PK (request_id)` — retries **overwrite** `sent_at`/`timeout_at` | a defect that shortens a window makes its own rows retry and leave the population you are counting; say what the census cannot see |
| **`agent_definitions` is not the whole universe of declarations** — `call_verifier`, `call_ingester`, `call_planner` are in no active definition | they come from per-run `workflow_plan`; "not declared here" ≠ "not declared" |
| **A first occurrence sitting exactly on the retention boundary is the boundary talking** — the oldest retained row IS the first wedge row | check `min(created_at)` before reading a start date off a census |
| A `failed` diagnose item still blocks `090`'s coverage check | read the four rows; if terminal and stale, `FORCE=1` is what it is for |

## Peer lane

`site_ai_agent_orchestration` is a stakeholder, briefed and signed off. Its handoff points **at
this lane** for the mechanism rather than restating it — **do not "helpfully" fill that pointer
back in**; it is deliberate, on `bugs_open/048` grounds.
