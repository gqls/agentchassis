# HANDOFF — 2026-08-19b — `bugs_open/029`, continue here

Supersedes `HANDOFF_2026-08-19_continue_here.md` (still accurate on Part A, RSH-011 and the
baseline trap — read it for those). Then `NOTES_retry_kills_live_child.md` §9 and §10.
`README_where_we_are.md` is the owner's plain-prose log — append, never rewrite.

## State in one line

**029 stays OPEN.** Part A fixed/approved/live/proven, unchanged. The wedge now has a **named class
with three verified source sites** and a **retired wrong subsystem**; a fourth 090 is in flight; and
one real framework gap in the diagnosis harness was found and is **not yet fixed**.

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

1. **Read `d52c3407`'s verdict.** CONFIRMED changes what to build; REFUTED is a result and must be
   recorded as a visible correction here.
2. **Build the `### awaited_requests` rows section** in `diagnose_load_runtime` — council-gate it.
   Ranked above any symptom-text workaround by the close-the-door rule: the workaround must be
   remembered by every future filer, and those who forget get a bundle that looks complete.
3. **Answer "what runs the step body twice at rv0"** — loop expansion / `ErrLoopExpansionHandled`,
   the recursive `continueExecution`, or a second consumer. Do NOT re-enter the retry machinery.
4. **Do not close.** Bar is fixed AND live; nothing about the wedge is fixed. Quiet since 08-17 is a
   baseline (6 of 8 surrounding days are also zero), not evidence.
