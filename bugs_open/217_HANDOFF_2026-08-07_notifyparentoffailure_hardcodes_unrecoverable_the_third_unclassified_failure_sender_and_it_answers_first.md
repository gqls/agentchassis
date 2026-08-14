# 217 — `notifyParentOfFailure` hardcodes `error_unrecoverable`: the third unclassified failure sender, and on a child-ORCHESTRATION failure it answers the parent first

**Filed 2026-08-07** by the `bugfix_207_sender_convergence` lane, found live during 207's
post-roll induction on v1.0.1262. **Status: OPEN, UNOWNED.** **Severity: medium — same
class as 207 (retry-quality, nothing corrupted, nothing silent), at the third and last
unconverged failure sender.**

> **RE-OBSERVED LIVE on v1.0.1266, 2026-08-08 16:07:45** (216's acceptance induction,
> corr `32a4c28e…`): the same child-orchestration failure drew BOTH envelopes on
> `system.generic.responses` — this sender's hardcoded
> `status=error_unrecoverable / CHILD_ORCHESTRATION_FAILED / recoverable:false` at
> 16:07:45, then the converged sender's `error_recoverable` at 16:07:46. One second
> apart, unclassified answers first, exactly as filed. **The gate this file named is
> now open: `bugs_open/216` is FIXED + LIVE + PROVEN (v1.0.1266)** — a recoverable
> verdict now produces a real replay, so converging this sender buys actual retries,
> not re-arm-then-terminal.

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

---

## 2026-08-08 — DECISION: CONVERGE (option 1). FIX BUILT; inert until its chassis roll.

Taken by a fresh session (ownership checked: who-owns, live transcripts, work-item
queue — all clear; literals re-verified at HEAD before starting).

**The import-direction question resolves to a CYCLE**: `platform/messaging` imports
`platform/orchestration` (processor.go:21, validation_drop.go:27 — the drop recorder
calls `orchestration.LogAgentError`), so the coordinator can never import
`messaging.RetryDisposition`. The fix moves the classification core (both Matched*
classifiers, their needle lists, `RetryDisposition`) to `platform/errors` — a
stdlib-only leaf that already owns `DomainError` and the typed codes — and leaves
re-export shims in messaging so every caller and every pinning test is unchanged
(agentbase's source-scan pin included). One implementation, reachable from both layers.

**The fix** (`coordinator.go notifyParentOfFailure`): classify before stamping —
`perrors.RetryDisposition(errors.New(errorMsg))`; status and `ErrorInfo.Recoverable`
move in lockstep; `Code: CHILD_ORCHESTRATION_FAILED` and the message stay verbatim
(their consumers key on them). New log line carries the disposition + matched token:
`retry disposition decided at the child-orchestration failure sender by the sequenced shared classifier`.

**Blast radius, measured before submission** (`severity='fatal'` rows are written only
by this sender, after the parent-exists check — the fatal rows ARE its sends):
14d population 11,970 → **6,239 (52%) flip to error_recoverable** (connection 4,163,
dominated by browser-runner `ERR_TUNNEL_CONNECTION_FAILED`; deadline exceeded 1,989,
firecrawl POSTs; timeout 87) · 4,756 permanent-terminal · 975 unclassified-terminal.
Chain depth measured 0 or 1 only (8,557 / 3,417 rows, none deeper) ⇒ retry
amplification across levels is bounded at (1+3)² = 16 innermost executions worst-case;
monitors: RSH-006 storm-watch + weekly fatal-row rate. Named follow-up if it fires:
terminal-exhaustion marker minted at the cap site (NOT shipped here — the 124
seam-in-a-bug-patch lesson).

**Sibling suspect RESOLVED: `TimeoutMonitor.sendTimeoutResponse` is DEAD CODE.**
`NewTimeoutMonitor`, `MonitorChildOrchestration`, `MonitorRequest(` and
`TimeoutMonitor{` have zero call sites outside helpers.go and tests (two spellings
searched). Nothing to converge.

**Tests**: `notify_parent_disposition_test.go` (216's harness). Mutation-verified both
ways: hardcoding the literal back fails the transient pin; swapping RetryDisposition
for MatchedTransientFailure fails the sequencing pin (`pq: invalid connection` must
stay terminal). All four packages green.

**Council**: `Council-Submitted: 471a969e-3546-4d34-bc9c-a481aca7f1d6` (plan:
lane PLAN_2026-08-08_217_notify_parent_disposition.md).

**Close criteria (post-roll — verify at the artefact, never the tag):**
1. Pod-grep every replica: POS `retry disposition decided at the child-orchestration failure sender` ≥1,
   NEG a string the diff removed — the old envelope had no removable literal, so use
   the marker pair from `scripts/pick-pod-marker.py`.
2. Induction (SEED_test_207_probe recipe): a deadline-exceeded child-orchestration
   failure draws `error_recoverable` from THIS sender (the envelope that used to be the
   16:07:45-shaped hardcoded terminal), and the parent's `awaited_requests.retry_version >= 1`.
3. Storm watch: retry_version histogram — mass 0–1, wall at 3, ZERO above; and
   week-over-week `severity='fatal'` row rate for cross-level amplification.

## 2026-08-08 (later) — council verdict READ: **APPROVED round 1** (`471a969e…`)

"Approved with 2 advisory objection(s), none high-severity" — 9 reviewers, 8 abstained.
Commit `b19ef6930` carries `Council-Submitted:`; 098 credits it automatically.
Objections and answers are on the record in the lane NOTES. Two things they improved:

1. **The close-criteria monitors are now concrete SQL, not prose** (guardian's point —
   they are operator-run queries, not deployed automation):
   ```sql
   -- storm watch: mass at 0-1, hard wall at 3, ZERO above (no backoff exists on the arm)
   SELECT retry_version, count(*) FROM awaited_requests
   WHERE sent_at > now() - interval '24 hours' GROUP BY 1 ORDER BY 1;
   -- amplification watch: week-over-week fatal-notification rate (this sender's own rows)
   SELECT date_trunc('day', occurred_at) d, count(*) FROM agent_error_log
   WHERE severity='fatal' AND occurred_at > now() - interval '21 days' GROUP BY 1 ORDER BY 1;
   ```
2. **TimeoutMonitor deadness now stands on three legs** (prior-art's asserted-absence
   challenge): bare-type grep → one COMMENT outside helpers.go; deployments/ → zero;
   `agent_definitions.default_config` → 0 rows. No construction site, no receiver, no
   config route.

## 2026-08-08 (night) — **FIXED + LIVE + PROVEN on v1.0.1269. All three close criteria met.**

Stays in `bugs_open/` per the owner's 08-06 ruling. The sender-convergence series
(207 → 216 → 217) is COMPLETE: all three chassis failure senders decide through
`RetryDisposition`, and the classifier has one implementation in `platform/errors`.

1. **Pod-grep** both v1.0.1269 replicas: POS sender literal 1, NEG synthetic 0
   (marker pair from `scripts/pick-pod-marker.py b19ef6930`).
2. **Induction** (216 runbook recipe; corr `1b4b43f2…`, R `a64d935a…`): THIS sender
   answered the deadline-exceeded child-orchestration failure with
   `error_recoverable` / `CHILD_ORCHESTRATION_FAILED` / `recoverable:true` at
   22:14:04.503Z (wire: legacy responses part 2 offset 36692) — the envelope that was
   hardcoded terminal in the 16:07:45 re-observation. It still answers FIRST; the
   processor's converged envelope 1s later now agrees. Re-driving the parent:
   `retry_version` 0→1 + `status='waiting'`, parent unfailed, and **the replay
   consumed from the void topic at offset 1** (`retry_version:1`, fresh
   message_id/timestamp) — a real retry, not a bookkeeping bump.
3. **Storm watch** (24h): retry_version histogram 2,392 / 58 / 4 / 9, hard wall at 3,
   ZERO above; fatal-rate steady across the roll (41–146/hr, no spike). Keep the
   week-over-week fatal-rate check running per the monitor SQL above — the
   amplification worst-case (16 at depth ≤ 1) remains theoretical, unobserved.

Evidence with every id and timestamp: lane NOTES 2026-08-08 (night) entry.

## Observation from the 268 lane (2026-08-14 evening) — two same-day instances of a child result failing VALIDATION on delivery, work fine both times

Contributed per who-owns (this lane owns the failure-sender seam); not a new
bug file. Two `content_rewrite` runs on dartsonline.com, ~30 min apart, both
ended `complete_error` with an EMPTY `error` column and `__step_error`:
`{"failed_step":"deploy_page","message":"workflow completed but its result
could not be delivered to the parent (failed_transient): message validation
failed (code: CHILD_ORCHESTRATION_FAILED)"}` — orchestrations
`49fa9f6b-43de-4e5c-b125-52ebf2bcbf6b` (18:25:44Z) and
`8183390d-a914-4b02-b9c5-3ba2e2e6e1a3` (18:53:xx). Both times the child's
WORK completed and DEPLOYED (DB rows updated 18:24:55/18:51:30; live pages
last-modified 18:26:48/18:52:21) while the `site_work_items` row went
`failed`. So: a validation-rejected child RESULT presents to operators as a
failed work item with the artefact fine — the "failed item, fine work" shape,
reproducible, and it looks adjacent to this file's unclassified-sender
question (which envelope answers, and how it is classified, decides the
item's fate; the work's fate was already decided). Items:
`20fd61a1-6fa6-4cc9-8fe0-41f43a790483`, `a61e48ba-e0f8-41ad-a748-fe55d874f503`.
