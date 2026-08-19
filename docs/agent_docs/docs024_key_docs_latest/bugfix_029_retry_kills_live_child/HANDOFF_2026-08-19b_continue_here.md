# HANDOFF — 2026-08-19b — `bugs_open/029`, continue here

Supersedes `HANDOFF_2026-08-19_continue_here.md` (still accurate on Part A, RSH-011 and the
baseline trap — read it for those). Then `NOTES_retry_kills_live_child.md` §9 and §10.
`README_where_we_are.md` is the owner's plain-prose log — append, never rewrite.

## State in one line

**029 stays OPEN.** Part A fixed/approved/live/proven, unchanged. The wedge now has a **named class
with three verified source sites** and a **retired wrong subsystem**; a fourth 090 is in flight; and
one real framework gap in the diagnosis harness was found and is **not yet fixed**.

> **UPDATED ~17:00Z — the 090 verdict is IN and the plan is written. Both are recorded below;
> the "IN FLIGHT" table is kept for the audit trail, not as outstanding work.**

## ✅ RESOLVED SINCE THIS FILE WAS WRITTEN

- **090 `d52c3407` verdict: `UNVERIFIABLE` — and it is a REAL abstention, the first run to READ the
  wedge.** Four citations, three `tier: state` quoting actual rows: it pulled `04518118`'s rows and
  the control and confirmed the signature. It abstained because the bundle carries no *bodies* for
  `handleCompleteResponse`/`continueExecution` and `orchestration_states` — the only place the
  map-vs-table divergence is observable — is purged for all six. Compatible-with is not
  observation-of. **NOTES §11.**
- **Its challenge to the rv0 finding is REFUTED at source. §9 stands.** It asked whether a retry path
  could reset `retry_version` to 0 on a fresh `request_id`. Only two `INSERT INTO awaited_requests`
  sites exist (`state.go:1611`, `spawn_actions.go:166`), **neither on a retry path**, and all three
  `retry_version` writers are `UPDATE … WHERE request_id`. Its own citation settles it inside one
  orchestration: `04518118`'s `call_handler` is **1 row at rv3**, its `spawn_handler` **2 rows at rv0
  with distinct ids**. Also found: `UpdateAwaitedRequestForRetry` (`state.go:1962`) is **DEAD**; the
  live writer is `UpdateAwaitedRequestRetry` (`coordinator.go:3337`). **NOTES §11a.**
- **C1 survives a test that could have refuted it. NOTES §12.** 17/17 duplicate gaps exceed the 300 s
  `StuckOrchestrationTimeout`, at 414–577 s. C1 and C2 both survive (identical table signature,
  nothing separates them); C3 refuted, C4/C5 disfavoured. Separating C1 from C2 needs logs the 08-17
  pods no longer have → next burst, RSH-011 armed.
- **The fix plan is written and GRADED:** `PLAN_2026-08-19_wedge_fix_park_advance_divergence.md`.
  Fable-drafted, checked against the live system. **Two draft claims did not survive** — a dead
  citation, and a "divergence census 0 and 0 `[MEASURED]`" that is **vacuous** (zero orchestrations
  are `AWAITING_RESPONSES` and zero `awaited_requests` rows are non-terminal fleet-wide, so 0/0 was
  forced). It was headed for a council submission; do not carry it.

## ⏳ IN FLIGHT — read these two first

| | |
|---|---|
| **090 run `d52c3407-14e7-4b9e-be46-c8ee741b2532`** | filed 16:14Z with six frozen `orchestration_id`s + a healthy control seeded into the symptom. **Verdict UNREAD.** `SELECT owner_agent_type,current_step,status FROM orchestration_states WHERE correlation_id::text='d52c3407-14e7-4b9e-be46-c8ee741b2532' ORDER BY created_at;` then `collected_data->'verdict'` on the `diagnose-agent` row |
| **Landmine verifier `a709e01b-afac-48ee-9e49-6aaf43fb8617`** | armed for the new 090-trigger entry |

## DONE 2026-08-19 (this session), all committed

- **`d609cedad`** — the rv0 measurement + NOTES §9 + the LANDMINE.
- **`805eba66b`** — `bugs_open/029` gains the 2026-08-19 section; owner log appended.
- **`7853605ce`** — NOTES §10, the framework gap.
- Council **APPROVED, all reviewers, round 1** on `e03f7122-7895-4b81-8add-5a93f69ed553`
  (the `schemaAlwaysTables` fix, commit `0132a3683`). `098` credits it automatically from the
  `Council-Submitted:` trailer — **do not amend, forward-only.**

## THE FINDING THAT RETIRES A WEEK OF SUSPECTS

`[MEASURED]` from `awaited_requests` (7-day table). 20 instances, all 08-17, none since.
**17/20 register the same `iter_{N+1}_spawn_handler` twice; `retry_version` is 0 on all 37 spawn
rows; 0/20 ever register the next `call_handler`.**

**A retry bumps `retry_version`. So the duplicate is NOT the takeover re-running the step — the step
BODY executed twice.** `handleSpawnRetry` and the whole retry-machinery candidate set are aimed at
the wrong subsystem. The "06:54–09:37 gap consistent with a >5-min takeover" agreed with the
hypothesis and could not have disagreed with it. `WRONG_CALLS.md` 2026-08-19.

## THE CLASS — three sites, all `[VERIFIED at source]`, composition `[UNVERIFIED]`

**The outstanding-awaited set has two representations and nothing reconciles them.**

1. `coordinator.go:2113` `persistAwaitingStateWithRetry` — the "did a reply beat the park?" check is
   keyed on **StepName**, not the request id. On a hit it returns `nil` **without** writing the
   awaited entry or setting `AWAITING_RESPONSES`; the caller reads `nil` as success and still calls
   `InsertAwaitedRequest` (`state.go:1609`). Row in the **table**, nothing in the **map**.
2. `coordinator.go:2671` `handleCompleteResponse` — `allDone := len(freshState.AwaitedRequests)==0`,
   from the **map alone**. That one boolean decides advance-vs-park and whether `continueExecution`
   runs. **The table is never consulted.**
3. `coordinator.go:848` `continueExecution` — silent early `return nil` when the loaded status is
   `AWAITING_RESPONSES`.

(1) creates divergence, (2) turns divergence into a wrong decision, (3) makes it silent.

## ⚠ THE FRAMEWORK GAP — found today, NOT fixed, and it is the best next platform change

`diagnose_load_runtime_action.go` renders row sections for `agent_error_log` (:274),
`site_work_items` (:309) and `orchestration_states` (:344). **There is no `### awaited_requests`
section.** `orchestration_states` retains ~26h; `awaited_requests` retains 7 days. So for any hang
older than a day the bundle renders rows from the table that is empty and describes the columns of
the table that is full. `0132a3683` fixed the schema half only.

**Three 090 stalls, three causes, all ours:** wrong table → right table/no schema → right
table/right schema/**no rows and no ids**. Rule earned: **when a run refutes on absence, establish
WHICH absence from its `needed_evidence` before re-filing.**

⚠ **`0132a3683` is Go and INERT until the next chassis roll.** Pods are on `v1.0.1315` (12:15Z);
the commit is 16:00Z, so it is **not aboard**. Verify after the roll at the build stamp, then by
checking a fresh bundle's Schema section actually describes the table.

## What is left

1. ~~**Read `d52c3407`'s verdict.**~~ **DONE — see above.** Next diagnosis action is NOT another
   re-file: the loop's abstention turns on evidence that no longer exists for this cohort. **Wait for
   the next burst**, which RSH-011 captures automatically with the full `awaited_requests` set.
   ~~ CONFIRMED changes what to build; REFUTED is a result and must be
   recorded as a visible correction here.
2. **Build the `### awaited_requests` rows section** in `diagnose_load_runtime` — council-gate it.
   Ranked above any symptom-text workaround by the close-the-door rule: the workaround must be
   remembered by every future filer, and those who forget get a bundle that looks complete.
3. **Answer "what runs the step body twice at rv0"** — loop expansion / `ErrLoopExpansionHandled`,
   the recursive `continueExecution`, or a second consumer. Do NOT re-enter the retry machinery.
4. **Do not close.** Bar is fixed AND live; nothing about the wedge is fixed. Quiet since 08-17 is a
   baseline (6 of 8 surrounding days are also zero), not evidence.

---

# APPENDED 2026-08-19 ~22:15Z by a SECOND, CONCURRENT SESSION — and an apology in the form of a warning

**I overwrote this file with `cat >` at 22:15Z and have restored it byte-identically from
`9857c0b2e`. Nothing above this line is mine; nothing above it was lost.** I did not run
`scripts/who-owns.py 029` and did not check `git log` on this directory before starting, so I worked
this lane all day without knowing you were on it. **The recovery worked only because the file was
committed** — exactly the case CLAUDE.md makes for the versioned-memory hook. `WRONG_CALLS.md`
2026-08-19.

**Read YOUR sections above as primary.** Where we disagree, you are right and I say so below.

## Where your findings CORRECT mine

- **The duplicate spawn is NOT the takeover — `retry_version` is 0 on all 37 spawn rows.** My
  NOTES §14 called the 06:54–09:37 gap "consistent with the already-established >5-min takeover".
  **That is refuted by your measurement, and worse, I had the refuting data on screen** — my own
  query output shows `retry_version 0` on both spawn rows of every pair. I read past it because the
  takeover story already fitted. Your §"THE FINDING THAT RETIRES A WEEK OF SUSPECTS" is right and my
  line is withdrawn.
- **Your three-site class subsumes my NOTES §6.** I found site (1), the step-name-keyed arrival
  check, independently and stopped there. You have the composition — (2) `allDone` computed from the
  **map alone** at `handleCompleteResponse`, (3) `continueExecution`'s silent early return — which is
  what turns divergence into a wrong decision and then hides it. Mine is a fragment of yours.
- **Your ranking on the framework gap is better than what I did.** You wrote: build the
  `### awaited_requests` **rows** section, and rank it above any symptom-text workaround, because a
  workaround must be remembered by every future filer. **I did the workaround. It failed** (below).

## What is genuinely NEW from my session

- **⚠ YOUR LINE "`0132a3683` … is NOT aboard" IS NOW STALE — the roll happened.** Pods are on
  **`v1.0.1316`** (17:13Z), build point **`07eeba4a1`** present on both replicas with the previous
  build point absent as a control; `0132a3683` **is an ancestor**. So the schema half is live.
- **And it is behaviourally PROVEN, by the check you specified.** A fresh bundle's Schema section
  renders `awaited_requests(request_id varchar, …, retry_version integer, …,
  processing_started_at timestamp, …)`. Control: **four pre-fix bundles render nothing** for it,
  while `orchestration_states(` is present in **all five** — so the section was fine before and the
  gap was specific to this table. NOTES §11 (mine — note our NOTES numbering has collided; see below).
- **The 7-day window is now grounded at SOURCE, not observed:** `cleanup_expired_awaited_requests()`
  runs `DELETE … WHERE status IN ('processed','expired','cancelled','error') AND processed_at <
  NOW() - INTERVAL '7 days'`. Keyed on **`processed_at`**, terminal rows only, enforced per minute.
  **The 08-17 cohort dies ~2026-08-24.** A Go-only `grep "DELETE FROM awaited_requests"` returns
  **nothing**, which is why this was assumed rather than checked.
- **Two more 090s were spent, against your "do not re-file" instruction, which I had not read.**
  `d02a6958` (3 iterations) cited a real 08-17 row and the right code path but stopped one query
  short. `5d1d8f1c` **regressed to 1 iteration** (`stopped_by = scope-not-narrowing`) because I
  changed **three things at once** — so nothing can be attributed. **Your instruction stands and mine
  is withdrawn: do not re-file; wait for the burst.** The artefacts
  (`NEXT_090_single_variable.sh`, `RECONSTRUCTION_QUERY.sql`, `SYMPTOM_d02a6958_baseline.txt`) are
  left in the lane dir for whenever a re-file IS warranted; **they are not a recommendation to fire
  one now.**

## ⚠ HOUSEKEEPING THE NEXT READER MUST KNOW

- **`NOTES_retry_kills_live_child.md` now has TWO §9–§11 sequences**, yours and mine, appended by two
  sessions that could not see each other. Neither is wrong; the numbering is. **Resolve by date and
  content, not by number** — mine are timestamped 2026-08-19 from ~10:45Z onward and concern
  retention, the ticker refutation, the bundle fix and the two runs above.
- **Traps I paid for that are not in your sections:** any check against a diagnosis bundle for a
  string YOU authored is blind (the symptom is quoted into the bundle verbatim — `LIKE
  '%awaited_requests%'` returns true on a *pre-fix* bundle); resolve a council verdict **by
  correlation, never by recency** (`ORDER BY created_at DESC LIMIT 1` returned another lane's
  approval); `row_cap` is **200** and an unfiltered dump `ORDER BY orchestration_id` returns a
  lexicographic slice; and flattening a `.sql` file to one line **without stripping its `--`
  comments** comments out the whole query and returns **zero rows with no error**.
