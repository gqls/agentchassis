> # ⚠ SUPERSEDED by `HANDOFF_2026-08-18b_continue_here.md` (2026-08-18 18:40Z)
>
> **Two of its claims are now wrong and are corrected there:**
> 1. Its "ONE OPEN TASK" — prove Part A behaviourally — is **DONE**. `call_dispatch` rv1 was
>    granted **00:15:00** (its declared 900s) at 18:28:21Z, `processed` inside the window.
> 2. Its lead *"the INITIAL wait may have its own conversion gap"* is **REFUTED** — `call_diagnoser`
>    rv0 is 30:00 in **29 of 29** rows across its whole history; every 180s reading is a RETRY.
>
> Its traps table and its "do not re-import either framing" warning remain correct and are carried
> forward. Read 18b.

# HANDOFF — 2026-08-18 — `bugs_open/029`, continue here

**Read this, then `NOTES_retry_kills_live_child.md` (technical log, newest at the bottom),
then `SUMMARY_2026-08-18_retry_window_inversion.md` (the read-out).
`README_where_we_are.md` is the owner's plain-prose log — append, never rewrite.**

## State in one line

**Part A is fixed, council-APPROVED, and LIVE on `v1.0.1309` — but BEHAVIOURALLY UNPROVEN,
and 029 itself is still OPEN.** The thing that actually freezes a build job is unexplained.

## The three things you must not get wrong

1. **029's title is wrong and so is its second framing.** The concurrency-group story was
   refuted in-file 2026-07-21. The "retry kills a still-working child" story is **mine and I
   withdrew it** on 2026-08-18 — the child is already dead ~7–10 min before the replay
   arrives, because the takeover arm requires `last_activity` >5 min stale and every state
   write bumps it. Do not re-import either.
2. **An APPROVED verdict here does NOT mean the bug is fixed.** Part A is one contributing
   defect. Every artefact says "part A only, does not close 029" in those words, deliberately.
3. **The fix is live but unproven, and those are different claims.** See below.

## What shipped

`retryWindow()` in `platform/orchestration/coordinator.go`, called from
`handleRecoverableError`. A replayed request now waits the window its **step declared**
instead of a recomputed one. The block it replaced capped every retry at **5 minutes** and
dropped to **3** for any declaration over 30 — an inversion: the longer a step declared, the
less it got. **33 steps across 25 agent types** declare >300s, up to an 86400s human-approval
step that was given three minutes on retry.

- Commits `bf7646a29` (fix + 5 tests) and `2a3d30ec3` (reuse: routed through the existing
  `getTimeout`, per the council's `reuse_agent` seat).
- Registered **RSH-010**; council corr `7c92389a-617f-4abc-b03b-0ef84ca2239f`, **APPROVED
  round 3**, `Council-Reviewed:` written on `2a3d30ec3`.
- Tests: `platform/orchestration/retry_window_declared_test.go`, **all five proven to fail by
  mutation.** ⚠ One of them (`TestRetryWindowIsMonotoneInTheDeclaration`) drives the row
  window **from the declaration on purpose** — with a constant row window it cannot fail
  against the old code and asserts nothing. There is a comment saying so. **Do not "simplify"
  it.**

## THE ONE OPEN TASK, and it is armed

**Prove the fix behaviourally.** It is live on `v1.0.1309` (verified at the artefact, controls
both ways) but the discriminating population has not occurred: post-roll there were 237
awaited rows and only **3 retries, all on steps declaring NO timeout**, which yield the 180s
default under old and new code alike.

```sql
-- THE test. Needs a retry on a step declaring >300s.
SELECT step_name, retry_version, (timeout_at-sent_at)::interval(0) AS window, status
  FROM awaited_requests
 WHERE sent_at > '2026-08-18 15:45:31+00' AND retry_version >= 1
   AND step_name IN ('call_dispatch','process_item_iter_0_call_handler',
       'process_item_iter_1_call_handler','process_item_iter_2_call_handler',
       'process_item_iter_3_call_handler','process_item_iter_4_call_handler','call_diagnoser')
 ORDER BY sent_at;
```

**Expect:** `call_dispatch` **15:00**, iter handlers **20:00**, `call_diagnoser` **30:00**.
**Under the old code:** 05:00 — or 03:00 for the 1800s case.
**Built-in positive control:** rv0 windows must be UNCHANGED (15:00 / 20:00). If rv0 has moved,
something else broke. If both are 05:00, the plan lookup is failing and the fallback is taken.

## Then: the real bug

**What kills the child's continuation after its FIRST spawn handshake?** The child completes
the handshake, then dies with no further state write, leaving `EXECUTING_STEP` with an empty
awaited set — invisible to `TimeoutMonitor` and to the retry driver, reaped only by the
4-hour arm. **This is the actual "hung spawn" and it is unexplained.**

- Candidate `[UNVERIFIED]`: the durable ticker's recovery runs under a shared **60-second**
  context (`cleanupExpiredAwaitedRequests`, `coordinator.go:4264`), so a continuation that
  spawns agents cannot finish inside it. But the fast-path timer uses `context.Background()`,
  so it cannot be the whole story.
- **File a `090` on THAT symptom, not the withdrawn one.** Check the queue first.

## Two more leads, both recorded and neither started

- **The INITIAL wait may have its own conversion gap.** `call_diagnoser` declares 1800s and its
  **rv0** window was 180s, while `call_dispatch` got its declared 900s on the same path shape.
  `[UNVERIFIED]`. Registration sites: `coordinator.go` ~2321 and ~2456, both
  `TimeoutAt: time.Now().Add(getTimeout(step))`; neither calls `ConvertStepTimeout` — only
  `executeStep` (:1174) does, on a by-value copy.
- **The dormant twin stays unfixed, deliberately.** `helpers.go retryTimedOutRequest` carries a
  hard 30s. Caller graph run (the council forced it): `NewOrchestratorHelper` has **ZERO**
  callers, so nothing constructs it. **Scoped claim: no caller in THIS repo** — both
  constructors are exported. Re-run
  `grep -rn "TimeoutMonitor\|OrchestratorHelper" --include="*.go" . | grep -v /helpers.go`
  rather than trusting this line.

## Traps that cost me time today — all in `RUNBOOK` and `LANDMINES.md`

| trap | the check |
|---|---|
| The freeze time is `last_activity`, **never** `updated_at` — on a reaped row `updated_at` is when the REAPER wrote, giving a uniform ~4h26m that is the reaper's own threshold | use `last_activity`; a suspiciously tight cluster means check the instrument |
| `handleRequestTimeout` names **two** methods — live on `SagaCoordinator`, dormant on `TimeoutMonitor` | walk the caller graph UPWARD to the constructors |
| A takeover log-grep returns zero **with the control also zero** — chassis log retention is **~4 min** | blindness, not absence; never cite it |
| The `build provenance` line is a STARTUP line and scrolls | absent ≠ unstamped; use the binary probe |
| **Probing the binary for your OWN commit sha returns absent on a binary that carries your fix** | the binary holds the sha it was BUILT from; use `git merge-base --is-ancestor` |
| `03:00` now means two opposite things | old inversion = a step declaring 1800s+; new code = a step declaring nothing |
| A wedge census filtered by `item_type` cannot see its own answer | the per-site mutex has no `item_type` clause |

## Peer lane

`site_ai_agent_orchestration` is a stakeholder, fully briefed, signed off, and has corrected
its own docs three times in step. Its handoff points **at this lane** for the mechanism rather
than restating it — **do not "helpfully" fill that pointer back in**; it is deliberate, on
`bugs_open/048` grounds (048 restated 029's refuted cause as fact and it travelled).

## Wrong calls from this lane, so you don't repeat them

Six in `WRONG_CALLS.md` for 2026-08-18: **five instruments that could not fail** (a timestamp
written by the thing measured; a filtered count; a log grep whose control was also zero; an
aggregate hiding a regression; an **asserted absence** with no query) **plus one fact that
could not be relevant** (a true statement offered in good faith, sitting beside a conclusion
it had no connection to). The check that catches all six: **name the value the query would
have returned if the thing were false — and for a true fact, name the conclusion it supports
and how.** "It points the same way" is coincidence with good manners, not support.
