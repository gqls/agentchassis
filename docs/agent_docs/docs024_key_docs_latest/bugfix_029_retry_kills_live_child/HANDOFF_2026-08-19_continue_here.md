# HANDOFF — 2026-08-19 — `bugs_open/029`, continue here

Supersedes `HANDOFF_2026-08-18b_continue_here.md` (bannered), which supersedes `..._2026-08-18_...`.
Read this, then `NOTES_retry_kills_live_child.md` (newest at the bottom).
`README_where_we_are.md` is the owner's plain-prose log — append, never rewrite.

## State in one line

**Part A is fixed, approved, live and PROVEN. 029 stays OPEN — but it is no longer "actively
biting and unexplained"; it is "rare, bursty, unexplained, and not currently reproducing".**

## What is DONE — do not redo any of it

| | evidence |
|---|---|
| **Part A proven behaviourally** | `call_dispatch` rv1 granted **00:15:00** (its declared 900s) at 2026-08-18 18:28:21Z, `status=processed` — answered INSIDE the window. Against **219 pre-roll retries, none above 05:00**. rv0 control unchanged |
| **Still shipping on `v1.0.1314`** | build point `d3590ca46` PRESENT both replicas, `deadbeef…` control absent; `bf7646a29`/`2a3d30ec3` are ancestors of it. rv0 windows unchanged on the new build |
| **Initial-wait lead REFUTED** | `call_diagnoser` rv0 is 30:00 in **29/29** rows, whole history; 18 agent/step pairs joined per owning agent, **18 honoured, 0 mismatches** |
| **Wedge signature measured** | 18/18 entered from the ERROR path; final spawn registered **twice** in 17/18; last state write **9–16s** after that final send; lifetimes **25:14–25:22** |
| **Dormant twin re-checked 08-19** | `helpers.go retryTimedOutRequest` (hard 30s) — still no constructor in this repo; only a test-file *comment* references `TimeoutMonitor` outside `helpers.go`. Scoped claim holds |

## THE OPEN QUESTION, unchanged

**What kills the parent a few seconds after its spawn handshake SUCCEEDS, on the path taken when
the previous iteration's `call_handler` ended in `error`?** Candidates, both `[UNVERIFIED]`:
`cleanupExpiredAwaitedRequests` (`coordinator.go:4313`, shared 60s context — but the fast path
uses `context.Background()`, so not the whole story) and `handleSpawnRetry` (`coordinator.go:1649`,
a candidate for the **duplicate** registration).

## What changed overnight, and it cuts both ways

- ✅ **The 090's blocker is GONE.** Origin advanced; `retryWindow` **is** on
  `origin/087_towards_multiple_domains`; HEAD is **17** commits ahead, not 233. A diagnosis filed
  now reads a tree that carries the fix.
- ❌ **The evidence is GONE.** `orchestration_states` starts 2026-08-18 07:58; **0** wedged rows
  retained. All 18 aged out on schedule. The loop reads the LIVE DB, so a 090 filed today would
  have **no instances to cite**.

**So do not file the 090 today on an empty table.** File it when there are instances — the
symptom text is ready in NOTES (2026-08-18 17:05Z), the seed scope is
`coordinator.go:continueExecution,handleSpawnRetry,createContinuationContext,handleRecoverableError`,
and `FORCE=1` is legitimate (the only coverage hit is a terminal `failed` item from 08-12).

## ⚠ THE TRAP THAT WILL CATCH THE NEXT READER

**Post-roll quiet is NOT the fix working.** The wedge's entry condition — a `call_handler`
reaching terminal `error` — by day:

| 08-12 | 08-13 | 08-14 | 08-15 | 08-16 | **08-17** | **08-18** | 08-19 |
|---|---|---|---|---|---|---|---|
| 0/448 | 0/232 | 0/1241 | 1/1302 | 0/481 | **30/1436** | **0/1603** | 0/20 |

**Thirty of thirty-one are one day**, and **08-18 was already zero across 1,603 rows — mostly on
the UNFIXED binary.** Six of eight days are zero. Any comparison of a pooled pre-roll rate
against post-roll silence is fiction (I built one, hedged it, and it was still worthless —
`WRONG_CALLS.md` 2026-08-19). **A quiet period only means something if the baseline is not also
quiet.**

## What is actually left on this lane

1. **Wait for the next burst, then act inside ~26 hours.** That is the whole of the remaining
   work on the wedge. Check with the RUNBOOK's wedge census; if rows exist, transcribe them AND
   file the 090 the same day.
2. **Decide whether the 26-hour evidence window is acceptable.** If this bug is worth solving,
   something has to capture the wedge rows before retention eats them — otherwise every burst
   costs a full evidence cycle to notice and loses it before it can be diagnosed. **Not built,
   not designed, and it is a real decision, not a task.**
3. **Nothing else.** Part A is closed out; the initial-wait lead is refuted; the dormant twin is
   deliberately unfixed with a re-verified scoped claim.

## Can 029 be closed? NO — and the honest reason

The bar is **fixed AND live**. Part A is both, but Part A was never the wedge — every artefact
says "part A only" in those words. The wedge mechanism is **unexplained**, and its absence since
08-17 is indistinguishable from the six other quiet days in the record. **Closing on that
evidence would be closing on a baseline.** Its one observed burst overlaps a GitHub API incident,
which is a lead nobody has pulled.
