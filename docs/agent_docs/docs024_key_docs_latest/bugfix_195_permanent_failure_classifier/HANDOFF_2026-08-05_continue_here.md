# HANDOFF 2026-08-05 — `bugs_closed/195`, continue here

> ## ⚠ SUPERSEDED — THE LANE IS CLOSED AND COUNCIL-APPROVED. Nothing here is owed.
>
> **`195` is CLOSED, live on chassis `v1.0.1252`+ (now `v1.0.1254`), induction-proven, and
> council APPROVED** (`9b1254f0`, round 2, 08-05 10:13Z, 5 advisory objections none
> high-severity). File moved to `bugs_closed/`.
>
> **Every owed step in the sequence below was discharged:**
> 1. **Resubmitted** — round 2 sent and **APPROVED**. (Beware: I first misread it as REVISE
>    because `max()` on a text column sorts `'revise' > 'approved'`. Use `ORDER BY created_at
>    DESC LIMIT 1`.)
> 2. **Induction read and PASSED** — corr `631232ed`: exactly 1 `agent_error_log` row,
>    `VALIDATION_ERROR_DROPPED`, `matched_needle = code:WORKFLOW_INVALID` (typed path), 0
>    duplicate rows. Baseline 0/needle-HIT 0. Plus the before/after that proves the
>    unconditional record: `PROCESSING_FAILED` **0 → 102 rows/24h** across the roll. Probe
>    agent deleted.
> 3. **Closed**, moved, `016b` §9 + §10, RSH-005, `MEMORY_closed.md`.
>
> **Two things happened AFTER the closure that a reader should know:**
> - The one medium objection (my "dead code" claim resting on grep-for-callers, which the
>   landmine set flags as unreliable for entry-point-shaped code) was **re-proved on an
>   independent axis**: `AgentServer` is **absent from the shipped binary** (0) while a live
>   sibling method is present (3). Limit: proves it for `agent-chassis` only.
> - **Another session's code review (`f887ed1ad`) found four errors in my shipped work**, two
>   of them mine: a self-contradicting comment I wrote (F1), and the two bare type assertions
>   I left in the very file where I added `AsDomainError` to replace them (F8). Both corrected
>   by that session; I have not reverted anything. See NOTES.
>
> **Still live and unowned, both filed at the council's direction:**
> **`bugs_open/196`** (failure responses reach the parent stamped `complete` with
> `Success: true` — arguably the more serious finding, and it falsifies `bugs_open/029`'s
> hung-parent reading for this path) and **`bugs_open/197`** (the sibling retryable-side
> classifier, still deciding by substring over prose — F8 above was a partial down-payment).
>
> Kept unedited below as the record of what was owed at 09:57Z and how it was discharged.
> **Do not pick work up from it.**

**Read this, then `NOTES_permanent_failure_classifier.md` bottom-up.** Lane is
`bugfix_195_permanent_failure_classifier`. Everything below is measured unless marked.

---

## State in one paragraph

Diagnosed, fixed at source, **LIVE on chassis `v1.0.1252` (both replicas pod-verified
2026-08-05)**. Council returned **REVISE**; every objection has been answered in code or by
measurement and committed, but **the resubmission has not been sent** — that is the first
owed action. An induction probe is in flight to produce the behavioural proof that closes the
bug. Nothing is blocked.

## What the bug was

`ValidateWorkflow` rejection renders as `WORKFLOW_INVALID: Invalid workflow configuration
(caused by: … requires a topic)`. The permanent-vs-transient decision was a **case-sensitive
substring match over that prose** (`{"is required","validation","invalid","missing"}`) — none
match: *"is required"* loses to the wording, *"invalid"* loses to the **capital I**. So the
fleet's commonest permanent config error was classified transient and
`recordDroppedValidationError` (the durable record `bugs_closed/034` provides) was never
called, **because the branch that calls it was never entered**. Total invisibility.

## What shipped (commit `28ef7a044`, live on v1.0.1252)

| change | where |
|---|---|
| `MatchedPermanentFailure` — typed `DomainError.Code` via `errors.As`, closed list `{ErrWorkflowInvalid, ErrValidation}`; needle list **untouched**, demoted to a fallback for untyped errors; returns an audit **token** (`code:WORKFLOW_INVALID` vs a bare needle) | `platform/messaging/validation_drop.go` |
| `AsDomainError` / `CodeOf` — chain-safe, survive `%w` | `platform/errors/errors.go` |
| both classifier call sites, in ONE commit (splitting them is the drift 034 closed); bare `err.(*DomainError)` → `errors.AsDomainError` | `processor.go#handleError`, `agentbase/agent.go` |
| `recordFailedProcessing` — an `agent_error_log` row on **every** non-dropped failure, so visibility no longer depends on classification being right | `agentbase/agent.go` |

Tests mutation-proven (reverting to the prose-only body fails three cases with the intended
messages). Registered **RSH-005**. `016b` §9 entry written.

## THE OWED SEQUENCE — in order

### 1. RESUBMIT to the council — NOT YET DONE, do this first

Verdict on `9b1254f0` was **REVISE**, gated by `guardian`. All objections are answered and
committed (`80a765320`); the resubmission itself was never sent. Use:

```bash
cd /home/ant/projects/agentchassis   # <- the trigger path is relative; a stale cd cost me a run
RESUBMIT_CORR=9b1254f0-2686-4a52-b736-1e212634ace6 \
  ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```

What to say in the round-2 rationale (all of it is already true and committed):

- **`editquality` (medium) — was RIGHT, and the answer is a measurement.** It asked whether
  `handleError`'s failure path can reach a caller other than `agent.go`'s `processMessage`.
  **It can:** `platform/agentbase/server.go:110` (`AgentServer.processMessage`), which logs to
  the pod and commits with **no classification and no `agent_error_log` row** — invisible
  exactly as 195 describes. **But it is dead code:** `AgentServer` is referenced *only inside
  `server.go`*, and `NewAgentServer` has **zero callers in the tree** (measured). So nothing
  live bypasses the recorder. Left in place and **documented where it fires** — a comment says
  reviving it without porting `recordFailedProcessing` reintroduces 195 wholesale.
- **`bug_historian` (medium + low) — two deferrals are now TRACKED items**, which was its
  stated REVISE condition: **`bugs_open/196`** (the success-shaped failure envelope) and
  **`bugs_open/197`** (the sibling retryable-side substring classifier).
- **Three lows, "you asserted, you didn't cite":** the sibling thin wrapper is
  `recordDroppedValidationError` at `agent.go:1389` and `validation_drop.go:150`, both thin
  wrappers over the one writer `orchestration.LogAgentError` — the shape council `180d7c68`
  permits, forbidding forked INSERTs. Put that in `grounded_in`.

The round-1 submission JSON is gone with its temp dir; rebuild from `PLAN_2026-08-04_*.md`,
which carries the same 8 edits.

### 2. Read the induction probe's result

Dispatched 2026-08-05 ~09:57Z, **result not yet seen at handoff time**:

```
agent:  test-195-invalid-workflow   (seeded probe; single step {"action":"complete"} — invalid)
CORR:   631232ed-7902-43c6-967b-f8402cb0eec5
ORCH:   e1fbd8f4-7aa1-424d-9cce-ac7d51155790
name:   ind195-105711
```

```sql
SELECT error_code, context->>'matched_needle', severity, created_at
FROM agent_error_log WHERE context->>'correlation_id'='631232ed-7902-43c6-967b-f8402cb0eec5';
```

- **PASS** = exactly one row, `VALIDATION_ERROR_DROPPED`, `matched_needle = 'code:WORKFLOW_INVALID'`.
  Baseline to beat: **0 rows**, needle-HIT **0** (the filer's probe `34268b8a`, chassis v1.0.1250).
- **Falsified by:** no row (fix not working); a **bare needle** token (typed path not taken);
  an extra `PROCESSING_FAILED` row for the same correlation (double-recording — the drop and
  the transient recorder must be mutually exclusive).
- Per the 173 RUNBOOK's TRAP 2, if no row: **grep the chassis log for the correlation FIRST**.
  In the log → latency refuted, get the reason. Not in the log → genuinely queued
  (publish→start measured at 29 min elsewhere), so wait rather than re-dispatch.

**Owed cleanup either way:**
```sql
DELETE FROM agent_definitions WHERE type='test-195-invalid-workflow';
```

### 3. Then close it

Bar is fixed AND live. With step 2 green: move
`bugs_open/195_HANDOFF_2026-08-04_*.md` → `bugs_closed/`, replace the STATUS banner with the
closing evidence, add the `016b` **§10** index row, and update `MEMORY_closed.md`.

## Things that will bite you

- **`grep -ac 'PROCESSING_FAILED'` returns 0 on a binary that fully carries this change** —
  measured 2026-08-05 on v1.0.1252. Short literals compile to immediate comparisons that never
  reach rodata. **My own register verify-later said to grep exactly that**, and it is corrected
  in place. Use the long string instead:
  `grep -ac 'processing failed and was NOT classified permanent'` → **1** on both replicas,
  with `grep -ac 'message dropped without retry or error response'` → **1** as the positive
  control (proves the probe reads the binary you think it does).
- **One `grep` per `kubectl exec`** — batching several into one `sh -c` times out on a binary
  this size. The image has no `strings`.
- **The trigger script path is relative.** A `cd` from an earlier command persists in this
  shell; I lost a submission run to exactly that and briefly thought the fleet had moved the
  file. `cd /home/ant/projects/agentchassis` first.
- **`errors.ValidationError(` greps `recordDroppedValidationError(` too** — a substring hit on
  an unrelated symbol that made a blast radius look 5× larger. I was measuring a
  substring-matching bug with a substring match. Use `-w` or read the hits.
- **Mutation testing must fail CLOSED.** My recipe's backup went to a `$SCRATCH` path that had
  been cleared; `cp` failed, `set -e` did not stop the script, and a shared-tree source file
  sat mutated until the trailing `diff` caught it. Assert the backup exists *before* mutating,
  or mutate only committed files so `git show HEAD:<path>` restores them. Full account in
  `WRONG_CALLS.md`; the `192` RUNBOOK recipe is corrected.
- **Never nest a heredoc whose terminator appears in the payload** (`<<'PY'` around text
  containing `PY`) — it closes early and leaks shell fragments.

## Spawned by this lane, unowned, deliberately not chased

- **`bugs_open/196`** — failures reach the parent stamped `complete` with `Success: true`, and
  the coordinator dispatches on the header, so the parent marks the step **succeeded** and
  records the error blob as its data. Filed with its unmeasured half stated: no live awaiting
  parent was induced, and the whole severity rests on that one measurement, which the file
  specifies. Falsifies `bugs_open/029`'s hung-parent reading for this path (029 has been told).
- **`bugs_open/197`** — the sibling retryable-side classifier (`isRecoverableError`,
  `IsRecoverable`) still deciding by substring over prose. Filed with **no live instance**,
  deliberately — waiting for one is the failure mode, and 195 is the proof.

## Paper trail

`PLAN_` / `NOTES_` / `README_where_we_are` / this handoff in this directory ·
`WRONG_CALLS.md` (the mutation-backup misstep) · `016b` §9 · register **RSH-005** +
index row · notices appended to `bugs_open/029` and inside `bugs_open/195` itself.

Commits: `28ef7a044` (fix) → `eedf59539` (PLAN+NOTES) → `0dd98510a` (status banner) →
`dc3dff8bc` (016b §9 + notify 029) → `a0ee2ee53` (README) → `2ce45dc43` (census verified) →
`80a765320` (revise r1: dead-caller measurement + 196 + 197).
