# 344 — the dispatch loop's `mark_complete` overwrites a ladder-re-triaged item to `complete` two seconds after its failure, so 307's fix turns the §2.2 daily bleed from honest reds into FALSE GREENS

**Filed 2026-08-21** by the session verifying `bugs_open/307`'s live acceptance. **OPEN, unowned.
LIVE and biting production** — one natural row within 18 hours of the 307 roll, demonstrated
twice more on a canary the same minute.

## Why this is filed on first-hand verification instead of a `090` run

Per the 2026-07-31 ruling: every arm of the mechanism below is quoted from code read this
session (four functions, four files), and the chain was then **demonstrated live twice** —
once induced (a canary on a pool site, torn down after) and once caught happening naturally on
`mortgagecalculator.co.uk` 42 seconds earlier. Nothing inferred; the disconfirming outcome
(the canary staying `triaged`) was written down as prediction P1 before the run and did not
happen.

## The defect, in one line

When a handler fails a step routed to a `status: 'failed'` write and then ends via a
success-labelled `complete_workflow`, the new failure ladder (WII-024) correctly re-triages the
item with a cooldown — and the dispatch loop's `mark_complete`, which runs on every returned
saga, **overwrites that `triaged` to `complete` seconds later**, because `triaged` is not in
the completion guard. The retry is cancelled and a failed build is recorded as done.

## Why this is NEW damage, not old damage restated

Pre-307, the same chain wrote terminal `failed` at the handler, and `failed` IS in the
completion guard — so `mark_complete` was refused and the item stayed an **honest red** (that
refusal is the 113-vs-2 `handled_by` split measured in `bugs_closed/301`'s contribution).
Post-307, the handler's failure write lands `triaged` (the ladder working as designed), which
the guard was never asked to protect. **The guard hole was harmless before 2026-08-20 16:09Z
and load-bearing after it.** Net effect on the §2.2 population (the "daily bleed", ~29
failures/day pre-fix, 72% of all failures): before the fix they died honestly at attempt 1;
after it they are falsely completed at attempt 1. No retry either way — and now no honest
record either. A false `complete` is this estate's most-filed defect class.

## The chain, each arm read this session

| # | arm | file:line | what it does |
|---|---|---|---|
| 1 | handler step fails → `error_step: mark_item_failed` | page-build-handler live config (`update_work_item_status`, `status: 'failed'`) | the §2.2 path, present on 5 live agents (307 §2.2) |
| 2 | the failure ladder runs | `work_item_failure_ladder.go` (WII-024) | **correct**: attempt+1, `status='triaged'` (attempt < max), `retry_after = now()+30m×attempt`, claim cleared, error preserved |
| 3 | handler continues to `complete_error` | live config: `complete_workflow` with a `success_message` | the child orchestration ends **COMPLETED** — "a FAILED step shows COMPLETED" (`bugs_open/099` landmine) |
| 4 | parent is told "complete" | `coordinator.go` `notifyParentOfSuccess` (~:3901): `Status: "complete"` **unconditionally** | no saga verdict travels; `__step_error` stays in the child's `collected_data` |
| 5 | loop calls `mark_complete` on every returned saga | build-dispatch-loop live config (`complete_work_item`, next_step of `call_handler`) | by design — the guard is supposed to protect deliberate statuses |
| 6 | gate 1 passes | `complete_work_item_verification.go:335` `handlerReportedFailure` | reads `response.status`, which arm 4 hardcoded to `"complete"` — this is `bugs_closed/196`'s fix, and this chain walks around it because the child never *reports* failure |
| 7 | gates 1b/2 pass | `verifyBeforeComplete` | `content_rewrite`/`needs_page`/`needs_imagery`… have no registered verifier and no `noChangeGates` entry |
| 8 | the overwrite | `load_work_item_actions.go:1025-1033` | `UPDATE … SET status='complete' … WHERE status NOT IN ('needs_human_review','failed','unresolved','rejected','wont_fix','verified','blocked')` — **`triaged` passes**; the code comment says so explicitly: *"Completing from an in-progress status (claimed/triaged/approved/detected/…) is unaffected"* |

Arm 8's guard list is the pre-307 sibling literal; 307's own `workItemCompletionGuardStatuses`
(which added `cancelled` and `deferred`) does not add `triaged` either — deliberately, and
correctly for the world it was written in: completing from an in-progress status was the normal
claim flow. What changed is that **`triaged` became a post-failure state** the moment the ladder
went live.

## The evidence — one natural, one induced, same two-second fingerprint

**Natural** [MEASURED 2026-08-21 ~10:40Z], `mortgagecalculator.co.uk`, item `0c65f9fa-ddce-4e83-a6a8-4f252b3cf3cb`
(`content_rewrite`, `gap_plan_add_scorecard-simulator_…`, created by `content-gap-planner` 08-15):

- real failure: `step process_sections_loop_iter_1_render_section failed: … component
  "mechanism-flow": content does not match the declared field type(s)…`
- ladder write ~10:32:50Z: `attempt_count=1`, `retry_after=11:02:50Z` (= +30 m exactly)
- **`completed_at=10:32:52Z`** — overwritten to `complete` two seconds later, `handled_by=
  'build-dispatch-loop'`, success-shaped `result.response`. The scorecard-simulator page did NOT
  get its content; the record says it did. **`retry_after` (11:02) is later than `completed_at`
  (10:32) — an item that "completed" while mid-cooldown, which is the queryable fingerprint.**

**Induced** (canary `f4f15466-07a0-4167-86e7-9cfdb554179d`, pool-web-tech.internal, torn down
after — see teardown note at the bottom): inserted 10:31:34Z with a spec carrying neither
`page_name` nor `page_id` (deterministic hard error at `load_page_record_action.go:197`,
non-transient text, single item so the burst conjunction cannot fire). Predictions recorded
before the run: P1 = row ends `triaged`/1/+30m (contract holds); P2 = row ends `complete` at
attempt 1 (this defect). **P2 observed**: ladder write 10:33:32Z (`attempt_count=1`,
`retry_after=11:03:32Z`, claim cleared, error preserved) → `complete` at **10:33:34Z**,
`handled_by='build-dispatch-loop'`.

*(The transient arm is NOT affected and was separately proven on natural traffic 2026-08-20
~18:34Z/18:54Z — those failures happen in the LOOP's own `mark_failed` (`fail_work_item` →
`done`), which no `mark_complete` follows. See 307 §9.)*

## Exposure

Every work item dispatched by `build-dispatch-loop` whose handler (a) fails a step routed to a
`failed`-writing error step and (b) then ends via a success-labelled `complete_workflow` — which
is the structure of all five §2.2 agents (`page-build-handler.mark_item_failed → complete_error`,
`image-build-handler.mark_work_item_failed → …`, image-source-unsatisfiable-handler,
image-url-404-handler, required-fields-missing-handler), i.e. the exact population 307's ladder
was extended to cover. Items dispatched by routes with no completing parent keep their retry
correctly. [MEASURED] one natural occurrence in the first 18 h post-roll (a quiet window);
pre-fix baseline for this population was ~29/day.

## Fix candidates, ordered by what closes the door

1. **Make completion refuse an item whose `retry_after` is in the future** — add
   `AND (retry_after IS NULL OR retry_after <= NOW())` to `CompleteWorkItemAction`'s UPDATE (and
   `failUnverifiedCompletion`'s, for symmetry), reporting `skipped` like the existing guard. This
   is the discriminating predicate, not a heuristic: a legitimate retry-then-success claims the
   item only AFTER its cooldown passes, so at completion time `retry_after <= now()` (proven on
   the two natural transient rows: stamps 18:34:00/18:54:33, claims 18:34:25/18:56:58, completions
   after). A future `retry_after` at completion time can only mean **this very saga's failure
   write already scheduled a retry** — completing it contradicts a decision made seconds ago by
   the same contract. Closes the door for every current and future parent, with no change to the
   coordinator seam and no new status vocabulary. ⚠ Shared-action change → council, and the other
   `complete_work_item` callers must be enumerated and told (RULING 2026-07-29 §3).
2. **Propagate a saga verdict from `completeWorkflow`** — a child that traversed an error step /
   carries `__step_error` should not reach the parent as bare `Status: "complete"`; gate 1 already
   exists to read a verdict, and this would give it one. The root repair, and the widest blast
   radius: `notifyParentOfSuccess` is fleet-shared, every parent's conditional on child status is
   in scope, and `bugs_closed/274` (result delivery) + `bugs_open/217` (the failure sibling) live
   on the same seam. Architecture-scope; do not ship inside this bug's patch.
3. **Add `triaged` to the completion guard list** — smallest diff, WRONG: it breaks any legitimate
   flow that completes an item another actor re-triaged mid-run for unrelated reasons, and it
   protects only this status word rather than the decision ("a retry is scheduled") the word
   currently happens to carry.
4. **Reorder the handlers so the failure write is the LAST step** (no success-labelled completion
   after it) — config-only, but it must be done per agent for five agents and every future one;
   "operators must remember X" is a defect (MEMORY: order-fix-candidates-by-what-closes-the-door).

## How to verify a fix

The canary recipe in `bugfix_307_terminal_write_contract/RUNBOOK_terminal_write_contract.md`
(§"The close canary") reproduces this deterministically in ~90 s: after the fix, the canary's
attempt-1 state (`triaged`/1/`retry_after`+30m) must SURVIVE the loop's completion call, the pod
must log the completion-skip, and the retry must then actually run at cooldown expiry. The
disconfirming control: a healthy item that failed once, waited out its cooldown, was re-claimed
and genuinely succeeded must still complete (the two 2026-08-20 transient rows are the recorded
shape). Queryable damage census: `status='complete' AND retry_after > completed_at`.

## Damage to repair when fixed

`0c65f9fa` (mortgagecalculator.co.uk scorecard-simulator content) is falsely complete — its
`item_key`'s dedup slot is released (`complete` is outside `idx_swi_dedup`'s open set), so
re-detection can eventually re-file it, but nothing guarantees when. Any later rows matching the
census above join this list. Re-open or re-file them once the door is shut.

## Relations

`bugs_open/307` §9 (the fix that made this reachable; its close is BLOCKED on this) · register
**WII-024** (the ladder — working as designed; this is the writer one step downstream) ·
`bugs_closed/196` + `bugs_closed/017` (gate 1's ancestry — child *reports* failure; this chain is
the success-labelled sibling that walks around it) · `bugs_open/099` (a FAILED step shows
COMPLETED — the landmine this chain is built on) · `bugs_open/217` (`notifyParentOfFailure`'s own
defect, same seam as candidate 2) · `bugs_open/341` (the fifth SQL ladder writer — untouched by
this) · `bugs_closed/301` contribution (the 113-vs-2 `handled_by` split that shows the guard
refusing pre-307).

## Canary teardown record

Canary `f4f15466` was driven through the remaining ladder arms after the discovery (attempts
2–3 — results recorded in 307 §9) and then DELETEd; verify with
`SELECT count(*) FROM site_work_items WHERE item_key='canary_307_close_20260821'` → 0.
Its `agent_error_log` rows remain (harmless, and they date the demonstration).

---

## §2 — 2026-08-21: candidate 1 BUILT (owner-directed), plus two findings about the census itself

**Go half committed `0f80f5ea1`, council corr `2c21e214-e459-420c-b451-3c66efa8bba9`
(`Council-Submitted:`, no verdict claimed). INERT until the next chassis roll** — the owner has
deliberately deferred the roll, so this defect is still live in production meanwhile.

Built as candidate 1: `AND (retry_after IS NULL OR retry_after <= NOW())` on **both**
completion-shaped writers, refusing for opposite reasons — `CompleteWorkItemAction` to preserve the
scheduled retry, `failUnverifiedCompletion` to avoid charging a second attempt for one fault (that
path is reached *after* the ladder has already counted it). Rendered from one function,
`workItemRetryNotPendingSQL` in `work_items_common.go`; writing the drift test found **two real
inline copies** left over from `307` (the dispatch selection and the atomic claim), so all five Go
sites now render from one place. Five mutations, each caught by a named test, including the
disconfirming control — an item that failed, waited and genuinely succeeded must **still** complete,
or this is a completion outage rather than a fix.

### ⚠ Finding 1 — the damage census in "Damage to repair when fixed" is STRUCTURALLY BLIND to archived rows

`site_work_items_archive` **has no `retry_after` column** (verified 2026-08-21: the two tables both
have 38 columns and differ by exactly two — live has `retry_after`, archive has `domain`). The
archiver copies by an explicit column list, so nothing is broken — but the fingerprint column is
simply **not carried across**.

Consequence for this file's census, `status='complete' AND retry_after > completed_at`: it can only
ever see the **live ~7-day window**, and for anything already archived the evidence does not exist
in any form — this is not "stale after 7 days", it is *absent by construction*. **So the damage must
be swept BEFORE rows age out**, or the archive needs the column. As of 2026-08-21 11:5xZ the live
census reads **0** — including `0c65f9fa`, the natural row this bug was filed on, which is no longer
matchable. Do not read that 0 as "no damage"; read it as "the window has moved".

*(Not fixed here: adding a column to the archive and backfilling it is a schema change for a
different bug, and it cannot recover rows already archived.)*

### ⚠ Finding 2 — `claimed-item-timeout` CANNOT produce this fingerprint, so no SQL half is needed

`bugs_open/341` §5b — which I recorded from a peer contribution — says this sweep's two
auto-COMPLETE arms stand in the same relationship to a mid-cooldown row as `mark_complete` does, and
therefore need the same predicate. **That is wrong, and it is wrong for exactly the reason §2b of
that file already gave about its RESET arm**: all three arms carry `WHERE wi.status = 'claimed'`.

A ladder-re-triaged row is `triaged` with `claimed_by`/`claimed_at` **cleared by the ladder itself**,
so no arm of this sweep can reach it. And a row cannot be claimed *and* mid-cooldown, because the
claim path refuses to re-claim before the stamp expires. Measured: **0 rows** at `status='claimed'`
carry any `retry_after` at all, and **0** fingerprint rows are attributable to the sweep
(`error LIKE 'Auto-completed%'`).

So the predicate would be dead SQL there. **This is the third time in two days that applying a
checklist to "the sweep" as a unit produced the wrong answer** — the correct question each time was
*which population does this arm select*, and for all three arms the answer is the same one that
makes them safe.

## §3 — what remains

- **The roll.** Committed ≠ live; this is still biting until the chassis rolls.
- **Candidate 2** (a real saga verdict out of `completeWorkflow`, so a child that traversed an error
  step stops reaching its parent as bare `Status: "complete"`) remains the root repair and remains
  architecture-scope — `bugs_open/217` sits on the same seam. Unowned.
- **Post-roll verification**: the `307` lane's canary recipe reproduces this in ~90 s. After the
  roll the canary's attempt-1 state must SURVIVE the loop's completion call, and the pod must log
  `CompleteWorkItemAction: skipped` with `reason=retry_scheduled`.
