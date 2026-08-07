# 217 — `notifyParentOfFailure` hardcodes `error_unrecoverable`: the third unclassified failure sender, and on a child-ORCHESTRATION failure it answers the parent first

**Filed 2026-08-07** by the `bugfix_207_sender_convergence` lane, found live during 207's
post-roll induction on v1.0.1262. **Status: OPEN, UNOWNED.** **Severity: medium — same
class as 207 (retry-quality, nothing corrupted, nothing silent), at the third and last
unconverged failure sender.**

> **VERIFICATION STATEMENT (owner ruling 2026-07-31).** Self-evidencing single-site
> finding: the deciding code is a hardcoded literal, read and cited; the behaviour was
> captured live on the wire during 207's induction. No 090 run for this file — the
> mechanism is one function with no classification logic to misread. (Its cross-cutting
> COMPANION, the dead recoverable arm, is `bugs_open/216` and does have a 090 run.)

## The mechanism

`SagaCoordinator.notifyParentOfFailure` (`coordinator.go:3899-3946`) is what answers an
awaiting parent when a **child orchestration** fails (`failWorkflow` → notify). It builds
the response with:

```go
Status:      "error_unrecoverable",   // coordinator.go:3924 — hardcoded
…
Recoverable: false,                    // coordinator.go:3937 — hardcoded
```

It has only `errorMsg string` in hand — no error object, no classifier call. Every
child-orchestration failure reaches the parent terminal, whatever its nature.

## Why 207's fix does not cover this (measured, not asserted)

In 207's live induction (corr `b155c554-0753-4f57-97a0-fcaec5d229d8`, v1.0.1262), one
deadline-exceeded child failure produced **two** answers to the same awaited request
R=`cef0a691…`:

| sent | sender | status | body.error |
|---|---|---|---|
| 08:20:33 | `notifyParentOfFailure` (child's coordinator) | **`error_unrecoverable`** | `code=CHILD_ORCHESTRATION_FAILED, recoverable=false` — message carries `…context deadline exceeded` |
| 08:20:34 | processor sender (207's converged seam) | **`error_recoverable`** | `recoverable=true`, matched needle `deadline exceeded` |

The hardcoded verdict fired **first**, and the first response claims the awaited request
(`ClaimAwaitedRequest`; the loser is `DUPLICATE_SKIPPED`) — the same pre-emption RSH-006
landmine 2 documented one seam down. On the real `call_agent` flow a child runs as an
orchestration, so **this hardcoded sender is the primary answer for mid-workflow child
failures** — 207's converged senders win only the failures that never reach the child's
coordinator (workflow-start and synchronous processing failures).

## The decision — same shape as 207's: converge or decline on the record

1. **Converge.** Classify before stamping: thread the failure through
   `messaging.RetryDisposition` (permanent first, then transient, else terminal —
   RSH-007). `notifyParentOfFailure` holds only a string; either thread the real error
   down from `failWorkflow`, or classify the string (the needle helpers exist; note
   `matchedTransientNeedle` is unexported and `RetryDisposition` takes an `error`).
   Import direction (`orchestration` → `messaging`) needs checking for cycles.
2. **Decline, on the record:** "a child orchestration that failed its own workflow is
   terminal by policy — retry the STEP, not the orchestration." That is a defensible
   position (the child may have side-effects half-applied), but it must be said here,
   because today it is an accident of a hardcoded literal, not a decision.

**Sequencing constraint: fix `bugs_open/216` first or together.** While 216 stands, a
converged `error_recoverable` from this sender is re-armed and then refused — converging
this seam alone converts hardcoded-terminal into re-arm-then-terminal, which is the same
outcome with better-looking bookkeeping.

## Relations

- `bugs_open/207` (fixed, live) — the first two senders; this is the third **chassis**
  sender. The full literal census at filing (`grep -rn '"error_unrecoverable"' platform/
  internal/`, non-test) shows, besides the converged processor pair and this site:
  `TimeoutMonitor.sendTimeoutResponse` (`helpers.go:409`) stamps a child-orchestration
  TIMEOUT terminal with `Code:"TIMEOUT", Recoverable:false` — a timeout is the textbook
  transient, so this is a sibling suspect; `[UNMEASURED]` whether that monitor is live
  beside the durable ticker path, check before converging it. `sendErrorResponseOLD`
  (`processor.go:2051`) is dead code (zero callers). `getErrorStatus` /
  `determineStatus` are classification-driven mappers, not deciders. The **adapter
  services** (thunder, analyser, browserrunner) carry many per-case hardcoded stamps of
  their own — separate services with their own failure semantics, out of this file's
  scope, named here so the census is honest.
- `bugs_open/216` — the dead response-driven recoverable arm this sender's convergence
  would deliver into.
- `bugs_closed/196` — established the failure-status envelope this sender predates.
