# HANDOFF — dead-pod ownership makes responses undeliverable; DB-driven retry then loops forever (with real side effects per cycle)

> **STATUS 2026-07-26: FIXED IN CODE, INERT — the case stays OPEN.** Fixes 1 and 2
> below are implemented and committed; the Go is dead weight until a chassis image
> is built past that commit and rolled. The standing bar for `/bugs_closed/` is
> **fixed AND live**, and "live" here means induced-fault-verified, not
> pod-grepped. Everything the roll owes is scripted:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_003_spawn_loss/VERIFY_075_post_roll.sh`
> — it refuses to run against an image that lacks the fix. See **Fix record
> (2026-07-26)** at the bottom.

**Filed:** 2026-07-25, by the bugfix-003 thread, from a live incident its own
induced-fault test caused and caught the same hour.
**Severity:** High — any coordinator pod death strands its active
orchestrations in a state where every response is silently discarded; since
bug-003's F2 went live (v1.0.1159), that stranding no longer terminates: the
retry driver re-executes the step every ~3 minutes indefinitely, and each
cycle can carry REAL external side effects (observed: a GitHub commit to
vm-sites per cycle). The reaper cannot catch it because every cycle refreshes
`last_activity`.

## One-paragraph version

When a response arrives, the consuming coordinator loads the orchestration
state and, if `state.ProcessingNode` names a different pod, logs "Response for
orchestration owned by different pod, ignoring" and returns nil
(`coordinator.go:269–277`) — under at-least-once consume the offset is then
committed, so the response is gone permanently. A dead pod's name never
matches any living consumer, so an orchestration whose owner died can NEVER
receive a response again. Pre-F2 that meant a silent strand until the 90-min
reaper. Post-F2 the DB retry driver dutifully rescues the expired request and
re-executes the step — the adapter does the work, responds in ~4 s, the
response is discarded by the ownership check, the request expires, repeat.
Two auxiliary defects keep the loop unbounded: (i) the adapter-action retry
branch (`coordinator.go:2710–2757`) creates a FRESH request with
`retry_version=0` each cycle and does
`RetryCount[stepName] = awaited.RetryVersion + 1` — assignment of 0+1=1 every
time, not an increment — so the `>=3` cap is unreachable; (ii) the reaper's
stale-AWAITING clause keys on `last_activity`, which every cycle refreshes.

## Evidence (all from 2026-07-25, UTC)

- 15:25:06 — chassis pod `agent-chassis-774877f4c6-7hpxp` deliberately deleted
  (bug-003 induced-fault test) while three `generic-orchestrate-0725-1524`
  orchestrations it owned were in flight.
- 15:28:01 — `RETRY_TICKER_CLAIMED` on **vet-intel** pod for request
  `6779e7a6` (orch `dc853e38`): the F2 cross-pod rescue worked exactly as
  designed, took the adapter-action branch, re-executed `deploy_page`.
- Then the loop, ~3-min period, per `awaited_requests` history: requests sent
  15:28:06 / 15:31:17 / 15:34:21 / 15:37:28 / 15:40:30 (orch `dc853e38`),
  every one `processed rv=0` moments after its timeout; mirror series on
  `14bedb94`. git-adapter logs show each cycle fully succeeding:
  "Committing to repo vm-sites" → "Successfully committed" →
  "Success response sent" — 4 s end-to-end, response produced, never applied.
- `orchestration_states.processing_node` for both: the dead pod's name.
- Consumer-group `git.adapter.group` lag 0 throughout — the adapter was never
  the problem.
- Third casualty: `026e9fab` frozen `EXECUTING_STEP/update_status` at
  15:25:10 (mid-step when the pod died; nothing re-drives an EXECUTING_STEP
  orphan until F1's >4h reaper).
- The class predates today: three `agent-tool-auditor-*` pods own active
  EXECUTING_STEP orchestrations and zero such pods exist; plus one
  INITIALIZED row owned by a replicaset gone since **July 13** (`0dcdd076`).

## Containment applied (2026-07-25, same hour)

`UPDATE orchestration_states SET processing_node=''` for `dc853e38`,
`14bedb94` (guarded on the dead pod's name), and `026e9fab`. The check admits
empty owner, so the next cycle's response applied: `dc853e38` moved to
`EXECUTING_STEP/complete` within a minute. The tool-auditor and July-13
orphans were left for their owning workstreams (grep before touching).

## Root causes (three, distinct)

1. **Ownership-by-pod-name with no liveness or takeover** — a design for
   in-memory locality that predates DB-backed state. The state is entirely in
   Postgres; any coordinator can drive any orchestration (the F2 ticker
   proves this cross-service). Discarding a response because a DEAD pod's
   name is stamped on the row protects nothing and loses real work.
2. **Adapter-action retry cap is unreachable** — fresh request each cycle
   (rv=0) + `RetryCount[step] = rv+1` (assignment, not increment). Even with
   ownership fixed, any genuinely-failing adapter step would loop forever.
3. **Reaper blindness to healthy-looking loops** — `last_activity`-keyed
   sweeps cannot see a loop that makes progress-shaped writes every cycle.

## Fix candidates

1. **Ownership takeover instead of discard** (the structural fix): on the
   ownership mismatch, attempt an atomic CAS —
   `UPDATE orchestration_states SET processing_node=$me WHERE
   orchestration_id=$1 AND processing_node=$stale` — and proceed if won;
   discard only if a LIVING owner holds it (loser of the CAS). Removes the
   dead-owner black hole entirely; the SKIP LOCKED claim discipline from F2
   is the in-tree pattern to copy.
2. **Fix the adapter retry accounting**:
   `RetryCount[step] = max(existing, 0) + 1` and CHECK it against the cap in
   the adapter branch before re-executing (route to
   `handleUnrecoverableError` at the cap, same as the message path).
3. **Reaper clause for loops**: sweep orchestrations whose
   `ExecutionMetadata.RetryCount[step]` exceeds N, or that have re-entered
   the same step more than N times within an hour (needs the count from fix
   2 to be real first).
4. **(Related, pre-existing)** sweep for active orchestrations owned by
   nonexistent pods — the tool-auditor/July-13 rows show these accumulate.

Fixes 1–3 are Go (platform/orchestration) — chassis image roll; fix 1 changes
response-path semantics and should cite RFC_001's precedent on review track
choice. Fix 4 could be a scheduled_tasks clause (config-only).

## How to verify

- Re-run the bug-003 kill test (delete the chassis pod mid-AWAITING): the
  orchestration must COMPLETE via takeover, with at most ONE retry cycle and
  no repeated GitHub commits.
- Induce a genuinely-failing adapter step (bad repo name): the loop must stop
  at the cap with a loud terminal error, not run forever.
- `SELECT processing_node ... WHERE status NOT IN (terminal) AND
  processing_node NOT IN (<live pods>)` returns 0 after the sweep exists.

## Fix record (2026-07-26, the bugfix-003 thread)

**What shipped (committed, inert until a chassis roll).** Fixes 1 and 2 only.
Fixes 3 and 4 are deferred with reasons and triggers, below.

### Fix 1 — takeover instead of discard

`SagaCoordinator.ProcessResponse` no longer returns nil on a pod-name mismatch.
It calls a new `StateRepository.TakeOverOrchestration(ctx, orchID, me, previous)`
— `UPDATE orchestration_states SET processing_node=$2, updated_at=NOW() WHERE
orchestration_id=$1 AND processing_node=$3` — logs
`ORCHESTRATION_TAKEN_OVER` (or `..._RACED` / `..._FAILED`) naming the previous
holder, and **processes the response either way**. The CAS deliberately does not
touch `version`, so it cannot collide with `UpdateStateWithVersion`'s optimistic
lock.

Two findings from this pass justify processing unconditionally, and both are new
to this file:

- **`processing_node` is write-once at row creation.** `SetExecutingStep`
  (`state.go`) assigns `state.ProcessingNode = os.Getenv("HOSTNAME")`, but
  `UpdateStateWithVersion`'s UPDATE column list — `… last_activity = $15,
  updated_at = $16, owner_agent_id = $17, owner_agent_type = $18,
  owner_agent_role = $19, version = $20` — **omits `processing_node`**, so that
  assignment never reaches the database. The gate was therefore comparing
  against "the pod that created the row", not "the pod driving it", which is why
  no amount of liveness reasoning could rescue it. A comment now sits at the
  inert line so the next reader does not trust it.
- **Discarding is unconditional loss in either consumer-group regime.** A
  response goes to exactly one member of the group that holds it. Responses are
  consumed with `GroupID = a.AgentID` (`agentbase/agent.go`) and `AGENT_ID` is
  `metadata.uid` in every production overlay, i.e. per-pod; under a shared group
  it is one member of many. Either way the pod holding the message is the only
  actor that can apply it, and `AgentClient.processResponse` commits the offset
  the moment `ProcessResponse` returns nil.

Single-processor safety was never coming from the stamp: `ClaimAwaitedRequest`
is an atomic `UPDATE … WHERE status='waiting' RETURNING` (exactly one actor per
response) and state writes are version-CAS'd. Deployment fact checked, not
assumed: **no multi-replica service owns any orchestration row** — all 1,284
`agent-chassis` rows (replicas 1), single-pod spawned Job agents, and
business-intel (replicas 1); core-manager (2) and reasoning-agent (3) own none.

### Fix 2 — the adapter retry cap made reachable

New pure helper `nextAdapterRetryAttempt(retryCounts, stepName, retryVersion)`
takes the HIGHER of the durable per-step counter and the request's own
`retry_version`, caps at `maxStepRetries = 3`, and returns the attempt number.
The adapter branch of `handleRecoverableError` now checks the cap **before**
re-executing, releases the claim terminally (`MarkAwaitedRequestFailed`, a
SQL-guarded no-op unless the row is `retrying`), and routes through
`handleUnrecoverableError` so `error_step` handling matches the message path.
The old `RetryCount[step] = awaited.RetryVersion + 1` becomes
`RetryCount[step] = attempt` — an increment, not an assignment.

Unit tests: `platform/orchestration/adapter_retry_cap_test.go`, four tests, no
DB. One of them, `TestOldAssignmentRuleNeverCaps`, is a **negative control**: it
encodes the replaced rule and asserts it never reaches the cap, so the suite
cannot quietly pass against code that never worked. All four pass (run in a
`git archive HEAD` + overlay scratch tree, because the package's pre-existing
external test file `orchestration_test.go` does not compile at HEAD — a stale
`NewSagaCoordinator` signature, nothing to do with this change and not edited).

Behaviour changes worth naming: the cap is per step name per orchestration for
the orchestration's whole life, not per burst; and a step that used to loop
invisibly now FAILS or routes to `error_step`, so failure counts will rise.

### Council

`SUBMISSION_CORR=4a227ed9-2a99-471b-8329-d0aceb63f28c` (6 edits).

**Round 1 = REVISE, and the decision was mechanical**: `decided_by` reads
*"unreadable reviewer(s): review_editquality.result"* — one seat's output was
lost, which forces REVISE regardless of content (the `bugs_closed/019` class).
Of the seats that were read: **4 approved** (constitution, mission,
debug_historian, prior_art_librarian), **2 objected** (guardian ×2 medium + 1
low, reuse_agent ×1 low), **no veto**. The guardian's note says explicitly it
"clears to approve next round" if the two questions are answered.

**Round 2 answers each objection with a check that was run, not an argument:**

- *Guardian: "no multi-replica coordinator" is in tension with "F2 drives
  orchestrations across services".* Both true — they are different verbs, and
  the round-1 submission conflated them. **Owns/creates rows:** agent-chassis
  (replicas 1), single-pod spawned Job agents, business-intel (replicas 1) —
  the census over all 1,644 rows returns only those (1,282 agent-chassis, 5 business-intel, the rest spawned single-pod Job agents; 378 distinct pod names). **May drive a row:** any
  pod running the per-minute ticker, in any service, because
  `ClaimExpiredAwaitedRequestsForRetry` claims by request, not by owner —
  proven cross-service on 07-25 (vet-intel's ticker, `RETRY_TICKER_CLAIMED`
  15:28:01). So pod-name ownership is **already** not respected by the retry
  driver; this change makes the response path agree with the driver.
- *Guardian: was a higher-layer fix rejected or never considered?* Rejected,
  three variants: quarantining stale-stamped responses upstream still destroys
  them (the loss IS the defect); fixing it in the retry driver cannot work
  because the driver neither created the stamp nor can repair it (pre-F2 the
  same discard was a silent 90-minute strand); the F1 reaper is structurally
  blind because every loop cycle refreshes `last_activity`.
- *Prior-art / reuse: does a takeover or re-stamp helper already exist?* No.
  Every Go occurrence of `processing_node` is a struct tag, two INSERT column
  lists, two SELECT column lists, the inert `SetExecutingStep` assignment, and
  a header round-trip. **No UPDATE writes it anywhere** in `platform/`,
  `internal/`, `pkg/`, `cmd/` — `TakeOverOrchestration` is the first. The live
  evidence agrees: after F2's cross-service rescue the column still read the
  dead pod's name.
- *Prior-art: verify the absence of a liveness mechanism before deferring fix
  4.* Verified both sides. Go has `sendHeartbeats()`, but it is a spawned
  CHILD messaging its PARENT over Kafka, returns immediately when
  `ParentAgentID` is empty (a Deployment pod like the chassis never sends
  one), and nothing persists it. In the DB the only pod-shaped columns in the
  whole public schema are `agent_error_log.pod_name` and
  `awaited_requests.processing_pod`; neither answers "is pod X alive now". The
  fix-4 deferral rests on a **checked** absence.
- *Reuse: note in code that the two locks must never govern one field.* Added
  as an INVARIANT paragraph on `TakeOverOrchestration`: the version-CAS owns
  every workflow field, the pod-name CAS owns `processing_node` alone, and
  fields go to one side, never both.

**Round 3 = APPROVED** (2026-07-26 21:41:21Z, 8 seats, all approve, no veto,
no unreadable seat). The trail on one correlation reads
`revise → revise → approved`. Four LOW advisory objections survive into the
approval and are answered here rather than dropped:

- *Guardian: the cap-then-fail path makes a previously-non-terminating case
  terminate, so failures will appear where loops were.* Intended and named in
  this file; the 07-25 loop is what it replaces. Watch for a rise in
  `ADAPTER_RETRY_CAP_REACHED` and FAILED orchestrations after the roll — that
  is the fix working, not a regression.
- *Guardian: the takeover is unconditional, which is fine only while
  replicas=1.* Correct. The residual is the CLAIM_RECOVERY hazard, recorded in
  `chassis_replica_scaling/NOTES` where the owning PLAN already documents that
  path.
- *Prior-art ×2: the "no other writer of processing_node" and "no multi-replica
  owner" premises cannot be verified from a review seat.* Also correct — so
  both are now re-run by `VERIFY_075_post_roll.sh` §4b on every invocation
  (a live replica listing plus the owner census, and the repo-side grep spelled
  out), and the disjoint-columns half is enforced by `state_locks_test.go` in
  `go test` rather than by a comment.

**No `Council-Reviewed:` trailer exists or can exist for this work** — the
trailer is earned by an APPROVED verdict only, and every verdict here
post-dates commit `5bbfe6a3a`. Expect 098 to list it as un-reviewed; that is a
permanent, correct false negative. Read the trail with:
`SELECT created_at, metadata->>'decision', metadata->>'decided_by' FROM diagnosis_artifacts WHERE correlation_id='4a227ed9-2a99-471b-8329-d0aceb63f28c' AND kind='council_report' ORDER BY created_at;`

### Deferred, with triggers

- **Fix 3 (reaper clause for loops)** — this file already sequences it after fix
  2 ("needs the count from fix 2 to be real first"), and fixes 1+2 remove the
  driver of the observed loop. **Trigger to file it:** any
  `execution_metadata->'retry_count'` value above 3 seen in the wild after the
  roll, or a loop observed with a shape the cap does not cover.
- **Fix 4 (sweep orchestrations owned by nonexistent pods)** — after fix 1,
  `AWAITING_RESPONSES` orphans heal on their next response. What remains is a
  different class: `EXECUTING_STEP` (covered by F1's >4h reaper) and
  `INITIALIZED` (covered by nothing — the 2026-07-13 row named above is still
  there). SQL cannot tell a live pod from a dead one without a heartbeat this
  system does not record, so a real fix needs new infrastructure and should not
  be guessed at inside the shared `stale-orchestration-reaper` pre_query.
  **Trigger:** an `INITIALIZED` orphan that matters, or the replicas≥2 work
  needing a liveness signal anyway.
- **The CLAIM_RECOVERY hazard is NOT filed as a new bug — it is already owned.**
  `processResponseClaimWithRetry` resets *any* claimed-but-unprocessed request
  back to `waiting`, including one claimed milliseconds ago by a live pod. With
  the ownership gate gone, that path is what would let two pods double-apply one
  response under a shared group with replicas≥2. It is harmless today (no
  multi-replica coordinator) and the account for it lives in
  `docs/agent_docs/docs024_key_docs_latest/chassis_replica_scaling/PLAN_2026-07-20…`,
  which already documents this exact path. A dated note has been added to that
  workstream's NOTES rather than forking a second account here.

## Cross-references

- `bugs_open/003` — the retry driver that made this loud (working as
  designed; do NOT revert it — before F2 this same defect was a silent 90-min
  strand).
- `docs/agent_docs/docs024_key_docs_latest/ANALYSIS_chassis_response_consumer_group_race.md`
  — the same ownership check blocks replicas≥2; fix 1 here is a prerequisite
  step of that work.
- `bugs_open/043` — prior processing_node appearance (diagnosis route hang).
- RFC_001 (`docs/.../architecture_review/`) — where the delivery-guarantee
  context lives.
