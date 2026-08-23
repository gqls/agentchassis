# 344 — the dispatch loop's `mark_complete` overwrites a ladder-re-triaged item to `complete` two seconds after its failure, so 307's fix turns the §2.2 daily bleed from honest reds into FALSE GREENS

**Filed 2026-08-21** by the session verifying `bugs_open/307`'s live acceptance.
**CLOSED 2026-08-21 — FIXED, LIVE ON v1.0.1322 AND CANARY-PROVEN** (see §4). Council `2c21e214`
APPROVED at round 2. ~~OPEN, unowned. LIVE and biting production~~ — one natural row within 18 hours of the 307 roll, demonstrated
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

---

## CONTRIB 2026-08-21 (`staged_component_build` lane) — a second change landed on this same step today; they are orthogonal, and here is why

**Heads-up, not an objection.** Migration `539` (submitted `d1a94170-8ec9-4b96-ae41-182cd052d291`,
commit in `docs/agent_docs/sql_for_agents/539_build_dispatch_loop_declares_commit_sha.sql`) adds one
config key to **the same step this bug is about** — `build-dispatch-loop`'s nested
`workflow.steps.process_item.config.sub_workflow.steps.mark_complete`:

```
"commit_sha?": "handler_result.response.commit_sha"
```

**Why it does not touch your defect.** 539 changes only *where the `commit_sha` input comes from*
(an explicit path instead of the resolver's whole-tree search). It does not change **whether
`mark_complete` runs**, **which statuses the completion guard protects**, or **what it writes**. Your
chain — a `triaged` item being overwritten to `complete` because `triaged` is not in the guard — is
untouched in both directions: 539 neither causes it nor mitigates it.

**Two interactions worth knowing, both small:**
1. **539 stamps `commit_sha` onto whatever `mark_complete` completes** — including, until this bug is
   fixed, the falsely-completed items you describe. That does not make the false green worse (the
   false `complete` is the damage), but it means a `result.commit_sha` on a post-539 item is **not**
   evidence the build succeeded. Worth knowing if anyone reaches for that field as a health signal.
2. **Fixing this bug will REDUCE completions**, which is the demand signal 539's own verification
   reads. If 539's post-apply window looks quiet after your fix lands, that is your fix working, not
   539 failing — check the demand control (`bdl` loop runs in the window) before concluding anything.

**Nothing is asked of you.** If you would rather 539 waited until the guard is fixed, say so in this
file and this lane will hold it — but our reading is that it need not, since 539's only effect on a
falsely-completed item is the presence of an extra field on a row that should not exist at all.

Context for why 539 exists: it is RFC_029 step 5's last live blocker (`bdl`/`commit_sha`, 387
conflict rows in 24 h), and the path was supplied by the `bugs_open/315` lane, who built the
handler-side half across nine agents. Full reasoning:
`docs/agent_docs/docs024_key_docs_latest/staged_component_build/HANDOFF_2026-08-20_continue_here.md` §2.7.

---

## §4 — CLOSED 2026-08-21: fixed, live on v1.0.1322, canary-proven, council APPROVED r2

**Go half `0f80f5ea1` + round-1 revisions `45ce175c8`.** Council corr
`2c21e214-e459-420c-b451-3c66efa8bba9` — **REVISE at round 1, APPROVED at round 2** with two
advisory objections, both dispositioned below.

### Live proof, not inference

- **Fleet on one stamp**: `bac189921`, 59 pods, `0f80f5ea1` an ancestor; the roll passed through a
  MIXED state (12 new / 275 old) on the way, so any stamp query taken mid-roll answers with two rows.
- **The canary**: attempts 1 and 2 re-triaged with their cooldown stamps and **SURVIVED the loop's
  completion call** — `completed_at` NULL, where the previous binary stamped `complete` in 2 s.
  The guard-skip arm was demanded live by flipping a claimed canary to `wont_fix` mid-run: the
  failure write skipped, the row was untouched, and the skip line was captured verbatim from the
  ephemeral per-job pod (`agent-page-build-handler-e3cd6375-vzcr4`) — **not** from a long-lived
  chassis replica, where it never appears.
- **Natural census**: **0** false-green rows (`retry_after > completed_at`) since the roll, against
  28 claims / 26 completions — so the zero has a demand control behind it.
- **Pre-fix damage list is EMPTY**: `0c65f9fa` was re-driven past the defect by its own lane and now
  sits `needs_human_review`/2.

### Round-2 advisories, dispositioned

1. **`guidelines` (medium) — the durable skip write touches an OPEN row, and the stale reaper keys
   on `updated_at`.** The seat was right to make me show this rather than assert it. Read live:
   `WHERE status='triaged' AND pipeline='build' AND updated_at < NOW() - INTERVAL '48 hours' AND
   claimed_at IS NULL`. So **any** write postpones a reap, not merely a periodic one — my round-2
   phrasing ("the landmine is about periodic writes") was loose. **It is nonetheless bounded, and
   here is the argument I owed:** the write fires *only* when a completion is refused, which fires
   only when a failure wrote that row moments earlier — so it is a second `updated_at` bump inside
   one event, against a **48-hour** threshold. On the `already_flagged_or_terminal` branch the row
   is not `triaged` at all, so the reaper never watched it. There is no cadence, and an item whose
   completions are being refused is by definition being worked, which is the opposite of what the
   reaper exists to catch.
2. **`editquality` (medium) — the sketch showed two incompatible refusal blocks.** True, **and it
   was a defect in my SUBMISSION, not in the code**: resubmitting, I appended the round-2 block to
   the sketch without removing the round-1 text, so the seats were handed a diff that contradicted
   itself. Verified in the tree: `pendingRetry` appears **0** times, one `reason :=` initialiser,
   one refusal path. *Reviewers judge the sketch* — a stale sketch is a real cost even when the
   code is right, and this is the second time this lane has spent a seat's attention on a rationale
   defect rather than a code one.
3. `editquality` (low) — the sketch implied rather than showed both `LoadWorkItemsAction` call
   sites. The census settles it: four call sites across three files, all rendering.
4. `bug_historian` / `guardian` (missing) — the durable write's `RETURNING` cannot silently no-op:
   the scan error is logged and never fatal, so a refusal is never turned into an error by
   bookkeeping. And migration `524` is correctly **not** in this submission's edit set — it belongs
   to `bugs_open/341` and shipped there.

### What is NOT here, and where it went

- **No SQL half was needed** — see `341` §5c. All three arms of `claimed-item-timeout` select
  `status='claimed'`, which a ladder-re-triaged row is not. Measured: 0 claimed rows carry any
  `retry_after`; 0 false-greens attributable to that sweep.
- **Candidate 2** (a real saga verdict out of `completeWorkflow`) remains the root repair and remains
  architecture-scope, unowned, on the seam `bugs_open/217` also sits on.
- **The outage-scale watch lives in `bugs_closed/307` only**, with its reopen trigger. It is
  deliberately not duplicated here: two records of one watch is the drift shape this estate keeps
  filing bugs about. If it fires, it reopens `307`, and this mechanism is re-examined from there.

---

## §5 — 2026-08-23: the closure RE-VERIFIED, and the CONTRACT's remaining reach measured (follow-through lane)

**344 stays CLOSED. Nothing here re-opens it.** A later session was asked to "fix bug 344",
found it closed, and re-verified the closure rather than trusting it. What follows is the
contract's reach — a different question from the bug, and the one that was left open.

### (a) The closure holds — re-verified, with a demand control

| check | result 2026-08-23 |
|---|---|
| `0f80f5ea1` / `45ce175c8` ancestors of HEAD | ✓ / ✓ |
| fleet | `v1.0.1330`, 37 chassis pods on one tag |
| predicate in `CompleteWorkItemAction` | ✓ `load_work_item_actions.go:1129` |
| predicate in `failUnverifiedCompletion` | ✓ `complete_work_item_verification.go:438` |
| damage census `status='complete' AND retry_after > completed_at` | **0** |
| **demand control** | 592 completions / 582 claims in 24 h, and **16** rows completed while *carrying* `retry_after` — every one with `retry_after <= completed_at`, the legitimate after-cooldown path |

The demand control is the load-bearing row: the census **could** have returned a positive, because
the population it filters on exists and flows daily. Without it the 0 would be worth nothing.

### (b) The contract is enforced at ONE `complete` writer of about eleven

344 shipped the predicate to two writers. **As of 2026-08-23**, these also write
`site_work_items.status='complete'`, can reach a `triaged` row, and did not carry it (verified by
direct read, not by grep count):

| writer | why it can meet a re-triaged row |
|---|---|
| `v3_site_actions.go:6291` `UpdateWorkItemStatusAction` | has the *status* guard, not the retry one; its own comment calls it "a third writer of `complete`". **Live config census: `image-build-handler`, `image-source-unsatisfiable-handler`, `image-url-404-handler`, `required-fields-missing-handler` all complete through it — 4 of the 5 agents §Exposure above names** |
| `apply_gap_plan_action.go:1180` `markOriginalComplete` | **the only writer that NAMES `triaged` in its own WHERE** — `status IN ('triaged','claimed')` — so it selects for exactly what the ladder writes. **FIXED 2026-08-23, see (d)** |
| `site_admin_handlers.go:1118`, `confirm_work_item_handler.go:213` | `WHERE id = $1`, no status guard at all — human dispositions |
| `work_items_common.go:444`, `plan_sections_action.go:2633`, `cmd/verifier-remit-check:538`, `cmd/brief-negation-check:548` | `triaged` not excluded |

⚠ **The count carries its date** (owner ruling 2026-08-22) — a census goes stale **by addition**.
Re-run before quoting: `git log --since=2026-08-23 --diff-filter=A -- platform/orchestration/actions/`.

**Sizing, honestly: latent, not bleeding.** Checked each of the four exposed agents' step graphs —
no live workflow config routes fail→complete inside one saga, and the damage census is 0. The
argument for closing these doors is that they are open, not that blood is flowing.

### (c) The predicate is a PROXY, and it can be silently disarmed — now a LANDMINE

`retry_after IS NULL` means **"claimable now"** to the claim path (`claim_work_item_action.go:109`
says so) and **"no retry pending"** to the completion guard. Those agree for a healthy row and
**contradict** for a zero-backoff ladder write, where `status` becomes `triaged` on one condition
(`:622`) and `retry_after` becomes NULL on a *wider* one (`:594`). Three live routes to that state:
`DISABLE_WORK_ITEM_RETRY_BACKOFF` (a kill switch named after backoff that also disarms this guard),
a `reaper_policies.backoff_minutes <= 0` row (operator-editable, no build), and
`retryAfterColumnPresent` latching false. **None armed 2026-08-23** (`__default__`=30,
`initial_verification`=20; no `DISABLE_WORK_ITEM_*` set in the cluster).

> **CORRECTION, recorded rather than quietly dropped:** this lane first listed
> `WORK_ITEM_BURST_COOLDOWN_MINUTES=0` as a fourth live route. It is **not** — `envInt` requires
> `n > 0` (`work_item_failure_ladder.go:297-304`) and falls back to the 15-minute default. The
> statement-level hazard survives for any caller passing 0 directly; the env route is closed.

### (d) Gap D — the contract's own effectiveness is UNMEASURABLE, and that is how this lane got it wrong

`result->'completion_skipped'` cannot be used to count refusals. The success path writes
`result = $2::jsonb`, a REPLACE, so a genuine later completion **wipes** the marker — `:1153`
says so. The two reasons have opposite survival odds: `retry_scheduled` marks an item that will be
retried and may then succeed (marker erased); `already_flagged_or_terminal` marks a terminal row
that never completes (marker permanent). So the observed **22 markers, all
`already_flagged_or_terminal`, zero `retry_scheduled`** is an artefact of the instrument, **not**
evidence that the guard has never fired — which is what this lane wrote down first. Logged in
`WRONG_CALLS.md` and `LANDMINES.md` (and it is the *second* key under `result` with this defect —
`result._verification` has it too, found 2026-08-08).

**The fix is one append-only log line** (`agent_error_log`, retains a month), not a change to the
guard. Not built here.

### (e) What was BUILT here, and what was deliberately NOT — and why

**BUILT** (commit `2dd05c5b2`, council corr `af5135d6-8ca2-4453-b33e-a299dcd6a622`): the contract
added to `markOriginalComplete`, plus the half nobody had noticed — it discarded **both** the error
and the rowcount from its `ExecContext`, so a refusal by the new predicate *or* by the status filter
already present reached no reader at all. Mutation-proven four ways (predicate removed → RED;
outcome discarded → RED; predicate re-inlined → RED on the drift test; predicate inverted → RED on
the disconfirming control), file restored byte-identically. The drift census in
`complete_work_item_retry_guard_test.go` re-derived from **four call sites across three files** to
**five across four, as of 2026-08-23**, with the staleness recipe attached.

**NOT BUILT, and this is a blocking constraint rather than a choice.** The other three gaps live in
files carrying other sessions' large, currently-**RED**, uncommitted work:

| file | uncommitted | last touched | holds |
|---|---|---|---|
| `v3_site_actions.go` | +23/−45 | **1 min before this was written** | Gap (b)'s highest-value site |
| `load_work_item_actions.go` | +139/−17 | 9 min before | Gap (d), and one of Gap C's two literals |
| `work_item_failure_ladder.go` (+test) | +181/−12, +312/−29 | 2026-08-22 | `bugs_open/345`'s repeat-termination rule; Gap (c) lives here |

`go test ./platform/orchestration/actions/` is **red in the working tree** and **green on a clean
`git archive HEAD`** — so the red is theirs, in flight. Editing any of these would take
half-finished work as a **same-file passenger** under this change's message, which is the one thing
a pathspec commit cannot prevent. **So they are recorded, not written.** Whoever picks them up:
re-check `git status` first; the constraint may have cleared.

### (f) RFC_043 Q2 — examined, and NOT taken, for the same reason

The owner ruled 2026-08-21 that the three completion-guard lists *"should become one with it's own
change and review"*, calling it **"unowned and unscheduled"**. One of the two inline literals is in
`load_work_item_actions.go`, which is being actively edited; converging one of two would be worse
than neither. **Still unowned.**

What this lane *did* contribute is the instrument for the owner's stated disconfirming question
("does any live flow legitimately complete a `cancelled` row?"), because the obvious data census is
**vacuous**: `result ? 'previous_status'` is present on **0 of 9,656** rows, so it could not have
come out otherwise. Read the **writers** instead — that argument *can* return a positive:

- **`cancelled`** — there is **no Go writer** of `site_work_items.status='cancelled'` anywhere
  (the only `SET status = 'cancelled'` in the tree is on `awaited_requests`, `state.go:2076`). Live
  rows carry `handled_by` values like `brochure_215_o2_thread` — hand-written by sessions.
- **`deferred`** — the birth status of non-dispatchable `capability_gap` rows (empty
  `handler_agent`), never a transition.
- **Neither is reachable by the dispatch path at all**: `ClaimWorkItemAction` admits only
  `status IN ('triaged','approved')` (`claim_work_item_action.go:102`).

Remaining for whoever takes Q2: enumerate the SQL-side writers in `sql_for_agents/` and the four
live `scheduled_tasks` sweeps that touch these statuses, and **date the enumeration**.

### (g) A design road NOT taken, recorded so it is not re-derived

A Postgres `BEFORE UPDATE` trigger on `site_work_items` would bind every writer at once — present,
future, Go, SQL, admin and `cmd/` — and `RETURN NULL` would reproduce the `rows==0` outcome every
caller already handles (precedent: migration `216` put `updated_at` on this table by trigger for
exactly that reason). **Rejected here** on RFC_043 Q3, owner 2026-08-21: *"Shared contracts live in
GO; a SQL sweep that needs one gets MOVED, not mirrored."* A trigger gives the contract a second
home, which is the drift shape this estate keeps filing. The counter-argument — that a trigger is a
*relocation* rather than a mirror, and would let the SQL copy in `claimed-item-timeout` be retired,
lowering the media count — is real and is **architecture-scope**: it needs an RFC and an owner
ruling, not a bug patch. Recorded, not pursued.

**Relations added:** RFC_043 Q2/Q3/Q4 · register **WII-003**/**WII-024** · `bugs_open/345` (whose
in-flight work blocks Gap (c)) · `bugs_open/354` + `bugs_open/341` (the root repair and the SQL
ladder copy, both **owned by the `bugs_open/307` lane** — untouched here by design).
