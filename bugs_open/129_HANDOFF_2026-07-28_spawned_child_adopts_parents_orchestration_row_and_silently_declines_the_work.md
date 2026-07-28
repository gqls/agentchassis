# 129 — the spawned child ADOPTS the parent's orchestration row and silently declines the work

**Filed 2026-07-28 (bugs thread 2).** Status: OPEN, unowned — but read the family notes
below before routing: the *measurement* discipline and prior refuted theory live in the
handshake record, and `platform/orchestration/coordinator.go` is being actively worked by
the work-item-parallelisation workstream (CS-1 landed there hours before this filing).

This is the **child-side mechanism** behind (at least part of) the known spawn→call
handshake failure — the family measured in `agent_error_log` as
`"Request <id> timed out after 3 retries"` (dominant sources `generic/call_dispatch`,
`build-pipeline-trigger/spawn_dispatch`; the 07-27 ordering-race theory for it was
loop-REFUTED, corr `eb8df254`, leaving "genuine non-response" as the cause). This file
shows what "genuine non-response" actually is on the child, with logs: **the request
arrives, the child processes it "successfully", and chooses to do nothing.**

## Symptom

`index-orchestrator` (spawn_indexer → call_indexer → complete) FAILED at `spawn_indexer`
twice consecutively, ~11:01 and ~11:10 on 2026-07-28, chassis `v1.0.1184`:

```
orchestration_states: FAILED / spawn_indexer
error: Request 91f55361-3ef3-457d-a1bd-5c583d9130c0 timed out after 3 retries
```

Both spawned pods came up healthy (1/1 Running, image v1.0.1184) and stayed idle.

## Evidence — the child's own log (captured before pod reaping; orchestrations 11a6e647, 58a53a6a)

The child received **all three retries** (`retry_version` 1, 2, 3 all logged), and for
each one walked the same path (pod `agent-code-indexer-25dc09ae-sjvbh`, 11:16:34, the
same lines for retries 1 and 2):

```
processor.go:285   "Workflow validated successfully, executing workflow"
processor.go:1748  "Executing workflow ... start_step: request_analysis, total_steps: 3"
coordinator.go:615 "Found existing state by orchestration_id"  orchestration_id=58a53a6a   <-- THE PARENT'S ROW
coordinator.go:153 "Orchestration state retrieved" is_new=false status=AWAITING_RESPONSES
coordinator.go:729 "Handling orchestration status" state.CurrentStep=spawn_indexer
coordinator.go:780 "Orchestration is awaiting responses in handleOrchestrationStatus" awaited_count=1
processor.go:293   "Workflow started and is now waiting for a response"                   <-- IT IS NOT
processor.go:1612  "ProcessMessage completed successfully"
```

`grep -c "fetched repo source"` on both pods' full logs: **0**. The child never ran
`request_analysis`, never fetched, never replied. The parent times out after 3 retries.

**The mechanism in one sentence:** the spawn request carries the PARENT's
`orchestration_id` (58a53a6a); the child's `SagaCoordinator.ExecuteWorkflow` loads state
by that id, finds the parent's row at `AWAITING_RESPONSES / spawn_indexer`,
and `handleOrchestrationStatus` concludes "already awaiting — nothing to execute" —
about a workflow (`request_analysis → index_symbols → complete`) that has never started.
The child's "success" log line and the parent's timeout describe the same event.

## What this is NOT (checked, same window)

- **Not a fleet-wide spawn break on v1.0.1184:** a spawned council (orchestration
  `32a1e2bc`, plan contains `spawn_agent`) completed with approval 11:12→11:17, between
  the two failures.
- **Not proven to be CS-1** (`afbd005f9`, staleness guard in the claim-recovery reset,
  same file, landed in the same roll). Attractive theory — pre-CS-1 CLAIM_RECOVERY
  stealing the parent's "live" claim could have been the accident that let children run —
  **but the hourly `agent_error_log` counts refute the clean version of it**: the 10:00
  hour's 15 `spawn_dispatch` timeouts ran on `v1.0.1182`, built BEFORE CS-1 was
  committed (09:52:37 UTC), and the family long predates CS-1 (07-27: 25/12, 23/13).
  `[HYPOTHESIS, UNTESTED]`: CS-1 may still have *changed the odds* on this branch;
  nothing here measures that.
- **Not the persist-ordering race** — that theory was already refuted by the loop on
  07-27 (`persistAwaitingStateWithRetry` re-loads and returns early when a response
  already arrived; verified citations in the handshake record).

## Why it matters

- The 24h `code-index-refresh` cadence goes through this exact lane — when this fires,
  the index refresh **fails silently from the cadence's point of view** (the scheduled
  task's `last_completed_at` still updates on trigger, not on outcome).
- The diagnosis loop and feature-builder spawn through the same coordinator path — the
  two prior `needs_diagnosis` filings for this family (SpawnAgentAction 07-27,
  ProcessResponse 07-20) are both `status='failed'`: **the defect eats the diagnosis
  runs sent to diagnose it.**
- It blocked the live verification of `bugs_open/108`'s fix for ~an hour (three manual
  dispatches to get one reindex through).

## Fix candidates — ordered by what makes the bad state unrepresentable

1. **A spawned child executing its own workflow must never satisfy itself with the
   PARENT's state row.** Key the child's state by its own orchestration identity (or
   namespace the row by executing agent), so "found existing state" can only ever find
   the child's own progress. Makes the swallow structurally impossible.
2. **`handleOrchestrationStatus` must not treat `AWAITING_RESPONSES` as "nothing to do"
   when the awaited request is THE ONE IT IS CURRENTLY PROCESSING.** The child is
   holding request `052f24bb` — the very request the parent's row is awaiting; declining
   it because the row says "awaiting" is self-referential. Narrower than 1.
3. **Reply something.** Even if the child declines, an explicit decline-response would
   convert a 6-minute-3-retry timeout into an immediate, diagnosable failure. Does not
   fix the defect; ends the silence.

## Burstiness confirmed on the same image, same hour

The THIRD consecutive dispatch (corr `7e89536a`, ~11:20) went straight through:
`request_analysis` fetched and analysed, `index_symbols` ran. Same image
(`v1.0.1184`), same lane, same input — so the swallow is **bursty, not
deterministic**, consistent with the family's history (25/12 → 24/1 → 23/13 windows).
2-of-3 failure rate for this lane today. Whatever gates the
`Found existing state → decline` branch varies per run; that variance is a diagnosis
lead in itself (timing of the parent's state write vs the child's first consume?
which pod consumed the earlier retries?).

## How to verify a fix

- Induced: dispatch `index-orchestrator` (TRIGGER_code_indexer_v2.sh) — currently fails
  ~consistently on v1.0.1184 within ~8 min; after a fix, `spawn_indexer → call_indexer`
  must advance and the child's log must show `fetched repo source`.
- The free reproducer for the wider family remains `build-pipeline-trigger` (every 30s,
  bursty): measure in `agent_error_log`, never `orchestration_states` (which showed
  166 COMPLETED / 0 FAILED while `agent_error_log` held 79 timeouts).

## Evidence preservation

Full pod logs from both failures are ephemeral (job cleanup). Salient excerpts above are
the durable copy; the failing orchestration rows `11a6e647-37c1-4d22-b081-e790c22abbb3`
and `58a53a6a-18f0-4aa0-878b-4ecc727c3f94` were **deliberately not cancelled** so a
diagnosis run can read them (the 07-27 lesson: cancel destroyed the evidence).
