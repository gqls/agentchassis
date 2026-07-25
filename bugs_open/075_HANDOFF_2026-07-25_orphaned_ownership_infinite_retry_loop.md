# HANDOFF — dead-pod ownership makes responses undeliverable; DB-driven retry then loops forever (with real side effects per cycle)

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
