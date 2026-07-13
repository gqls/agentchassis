# FOCUS — Platform reliability: collected_data bloat and stuck-state recovery

Two interacting platform-infrastructure problems surfaced during the
2026-05-19/20 gaswholesalers debugging sessions. They compound: the OOM
bloat (Part 1) is the upstream cause of many stuck claims, and the
reaper gap (Part 2) determines how painful each stuck claim is before it
clears.

> Consolidated from `TODO_orchestration_memory_bloat.md` and
> `reapers_and_stuck_state_recovery.md`. The reaper doc's original "reapers
> are Go code" framing is corrected inline per the 2026-05-21 finding (the
> reapers are scheduled_tasks SQL pre_query entries).

---

## Part 1 — collected_data growth causes OOM-kills and lost work

**Severity:** High — currently causing pod OOM-kills and dropped responses
**Status:** Diagnosed, not yet fixed
**Found:** 2026-05-19 during gaswholesalers news debugging session

## Observation

Component-quality-auditor orchestrations show `collected_data` of **18 MB each**
at step `create_regen_items_iter_8_check_quality`. Six identical-size
orchestrations all stuck on the same iteration step:

```sql
SELECT orchestration_id, current_step, pg_size_pretty(LENGTH(collected_data::text)::bigint)
FROM orchestration_states
WHERE owner_agent_type = 'component-quality-auditor'
  AND status IN ('AWAITING_RESPONSES', 'EXECUTING_STEP');
-- All 6 rows: 18 MB collected_size
```

Pod limits: `memory: 512Mi`. Loading a single 18MB `collected_data` plus the
Go runtime baseline (~80-150MiB) plus working buffers for LLM calls and
JSON encoding/decoding gets very close to the 512Mi cap. Auditor pods
have been observed being OOM-killed mid-processing:

```
Last State:     Terminated
  Reason:       OOMKilled
  Exit Code:    137
```

## Consequence chain

When an agent pod is OOM-killed:

1. Mid-processing work is lost
2. If the kafka request offset was already committed before the OOM, the
   request is NOT redelivered — work is gone permanently
3. If the agent's response wasn't produced before the OOM, the parent
   orchestration waits forever and eventually times out
4. Build-dispatch-loop FAILS its iteration
5. Generic parent FAILS its call_dispatch with "Request X timed out
   after 3 retries"

Today's session showed exactly this pattern: parent orchestrations
timed out at call_dispatch even though the children completed (per DB
state) — but the kafka response message was missing from the topic.

The OOM correlates with the empty topic because both stem from the
agent being killed during the publish phase.

## Suspected contributors to collected_data bloat

Without code-level profiling yet, the most likely contributors:

| Field | Why it grows | Mitigation |
|---|---|---|
| `__raw_message__` | Stores the entire inbound kafka message including the body. Duplicates information that's already structured in `input_data`. | Drop `__raw_message__` once parsing is complete |
| `processing_history` | JSONB array, one entry per status change per step. The auditor's iterative pattern (`iter_0`, `iter_1`...) compounds this. | Cap to last N entries, or move to a separate audit table |
| LLM responses | A single Claude response with tool use can be 50KB+. Across iterations, these accumulate in `collected_data` under each step's output_field. | Strip large fields from collected_data after the step has consumed them; only retain what downstream steps need |
| Component output across iterations | The auditor processes multiple components per run; each iteration adds its result to collected_data without releasing prior iterations' details. | Streaming-style writes (DB updates per iteration with old data discarded) |

## Investigation steps when picking this up

1. Dump one of the 18MB rows to a file and inspect what's actually in
   there:

   ```bash
   psql -h <host> -U <user> -d clients_db -tA -c "
     SELECT jsonb_pretty(collected_data)
     FROM orchestration_states
     WHERE orchestration_id = '914cc2ce-1278-42eb-9adc-174af4a52d54'
   " > /tmp/auditor_collected.json
   wc -c /tmp/auditor_collected.json
   # Then look at top-level keys and their sizes
   jq -r 'to_entries | map({key: .key, size: (.value | tostring | length)}) | sort_by(.size) | reverse' \
     /tmp/auditor_collected.json | head -30
   ```

2. Find which field is dominant. Almost certainly one or two large fields
   are responsible for the bulk.

3. Identify the lifecycle of those fields:
   - When is the data first added to collected_data?
   - Which subsequent steps actually read it?
   - When can it be safely dropped?

4. Patch the relevant action(s) to release the field once no longer needed.

## Short-term mitigations to apply before fixing root cause

While the proper fix is sized:

- **Raise the per-pod memory limit** from 512Mi to 1Gi or 2Gi. Nodes have
  headroom (11%/5%/21%/8%/37% used per `kubectl top nodes`). This
  prevents OOM-kills while the proper fix is developed.
- **Add `GOMEMLIMIT`** to the deployment env vars at ~88% of the k8s limit
  so Go GCs aggressively before the kernel OOMs:
  ```yaml
  env:
  - name: GOMEMLIMIT
    value: "880MiB"     # for a 1Gi limit
  - name: GOGC
    value: "75"
  ```
- **Strip debug symbols** from the binary: `go build -ldflags="-s -w" -trimpath`

These don't address the underlying bloat but stop the cascading failures.

## Why this matters more than it looks

OOM-kills are not just "the pod dies and gets restarted, no big deal".
They cause:

- **Phantom-completed orchestrations** (DB says done, no kafka response)
- **Cascading parent timeouts** that look like response-routing bugs
- **Hours-long debugging sessions** chasing routing issues that are
  actually memory-pressure consequences

This was today's failure mode. Until collected_data is bounded, periodic
OOM-cascades will mask other issues.

## Related issues already documented

- `031_locks.md` — leases on awaited_requests, related to claim race
- `015_batch_processing_architecture_v2.md` — iterator patterns are the
  context where collected_data accumulates fastest
- `016_debugging_guide_v2.md` Section 9 — should include an entry
  pointing to this todo

## Cross-reference: consumer group bug from same session

The consumer group bug (line 3152 of `platform/agentbase/agent.go` using
`a.AgentID` for the response consumer group) is structurally separate
from this memory issue but worth fixing in the same chassis update since
both require a chassis rebuild + deploy.

---

## Part 2 — Reapers and stuck-state recovery

> **Correction (2026-05-21).** An earlier draft of this section assumed the
> reapers were Go code in the coordinator. They are not: the reapers are
> **SQL `pre_query` entries in the `scheduled_tasks` table**, not Go. The
> confirmed scheduled reapers are `stale-orchestration-reaper` (180s — fails
> build-dispatch-loop AWAITING_RESPONSES >30min, anything >90min, expires
> awaited_requests >5min past timeout), `stuck-task-reaper` (300s — resets
> stuck scheduled_tasks), `stale-work-item-reaper` (3600s — triaged >48h →
> unresolved), and `claimed-item-timeout` (pre_query not yet captured). The
> Go 5-minute on-access `StuckOrchestrationTimeout` check below is a
> **secondary** mechanism, not the primary one. The 51-minute stuck claim
> described below was cleared by the scheduled `stale-orchestration-reaper`,
> NOT the Go on-access path. Read the mechanism analysis below with that
> framing: the Go on-access check exists, but the scheduled SQL reapers are
> the workhorses.

The text below documents the mechanisms as originally analysed (the Go
on-access check, the FailWorkItemAction retry paths, the agent-job-cleanup
CronJob, and the claim-reaper gap). The mechanism logic is accurate; only
the "where the reaper lives" framing needed the correction above.


There are three reaper-like mechanisms in the chassis, each operating at a
different layer. Together they recover from most stuck-state scenarios,
but there is a clear gap at the work-item layer that is worth noting.

## 1. Stuck-orchestration reaper

**Where:** chassis coordinator, near `StuckOrchestrationTimeout = 5 * time.Minute`

**What it does:** When ANYTHING accesses an orchestration's state (a request
arrives, a response arrives, a periodic poll runs), the coordinator
checks:

```
if state.CurrentlyExecuting != nil &&
   time.Since(state.LastActivity) > 5 minutes:
    log("Found stuck orchestration, taking over")
    repo.ClearExecutingStep(orchestration_id)
    reload state
    continueExecution(state)
```

So an orchestration stuck in `EXECUTING_STEP` for over 5 minutes gets
its `CurrentlyExecuting` cleared, its state is reloaded, and execution
resumes from where it left off.

**Critical detail — this is NOT a periodic sweep.** It only runs when
something tries to access the orchestration. If nothing accesses it, the
orchestration sits stuck indefinitely. In practice, the kafka-scheduler
runs every minute or so and indirectly causes access through normal
workflow operations, so most stuck orchestrations get reaped within
~5-10 minutes. But there's no background goroutine sweeping for stuck
states.

**Comment from the code about a related scenario:**
> "A deferred recover() converts any panic from inside the handler into
> an error. Without this, a panic kills the processing goroutine and
> leaves the orchestration stuck in EXECUTING_STEP with no log trail
> past the panic — recovery via the reaper only kicks in after 30+
> minutes."

The "30+ minutes" wording suggests that in the worst case (no nearby
activity to trigger the access path), recovery can take much longer
than the 5-minute timeout implies.

## 2. Work-item retry on failure

**Where:** `FailWorkItemAction` in actions.

**What it does:** When a child orchestration explicitly fails its work
item, three paths exist:

1. **AI unavailable** — release back to `triaged` WITHOUT incrementing
   `attempt_count`. The item will be retried when the AI endpoint
   recovers:
   ```sql
   UPDATE site_work_items
   SET error = $2,
       status = 'triaged',
       claimed_by = NULL,
       claimed_at = NULL,
       handled_by = NULL,
       updated_at = NOW()
   WHERE id = $1
   ```

2. **Status override** — caller passes explicit status (e.g. `'blocked'`,
   `'wont_fix'`):
   ```sql
   UPDATE site_work_items
   SET error = $2,
       status = $3,
       handled_by = $4
   WHERE id = $1
   ```

3. **Generic failure** — increment attempt count, release to triaged if
   under max_attempts (default 3), otherwise mark failed:
   ```sql
   UPDATE site_work_items
   SET attempt_count = attempt_count + 1,
       error = $2,
       status = CASE
           WHEN attempt_count + 1 >= max_attempts THEN 'failed'
           ELSE 'triaged'
       END,
       handled_by = $3
   WHERE id = $1
   ```

**The eligibility filter at claim time:**
```sql
SELECT id FROM site_work_items
WHERE status IN ('triaged', 'approved')
  AND attempt_count < max_attempts
```

This is how items with `attempt_count >= 3` (default) are excluded from
future claims. They sit at `status='failed'` and stay there.

**Important caveat:** `FailWorkItemAction` only runs when the child
orchestration EXPLICITLY fails. If the child orchestration gets stuck
(EXECUTING_STEP / AWAITING_RESPONSES) and the stuck-orchestration reaper
kicks in to retry rather than fail, the child orchestration restarts
fresh but the work item's claim is not released until the orchestration
either fully completes (success) or hits its own failure path.

So a stuck orchestration with a long-held work-item claim only gets
released via this path when the orchestration finally fails. The claim
isn't ageing out independently.

## 3. agent-job-cleanup CronJob

**Where:** `services/agent-job-cleanup/agent-job-cleanup-cronjob.yaml`,
runs every 10 minutes.

**What it does:** Operates at the Kubernetes layer:
- Deletes pods in `status.phase=Failed`
- Deletes stale Jobs labelled `spawned-by=orchestrator` for specific
  agent types (vet-practice-verifier, vet-batch-processor,
  area-sweep-orchestrator, area-sweep-discoverer)

This is purely k8s housekeeping. It does NOT touch any database state.
A failed pod gets removed from `kubectl get pods` but the orchestration
it owned (if any) remains in whatever state it was in. The
stuck-orchestration reaper (mechanism 1) handles that side.

## The gap

There is no work-item-level reaper that periodically sweeps for stuck
claims. The schema HAS the index for it:

```
"idx_swi_claimed" btree (status, claimed_at) WHERE status = 'claimed'
```

That index is shaped exactly like a "find old claims" query
(`WHERE status='claimed' AND claimed_at < NOW() - INTERVAL '30 minutes'`),
but no chassis code I've found uses it. The index sits unused.

**Consequence:** if a pod claims a work item, then dies in a way that
doesn't fail the orchestration cleanly (OOM-kill mid-step,
SIGKILL, node eviction, network partition followed by silent loss), the
work item stays `status='claimed'` indefinitely. Other dispatchers see
it as still in progress and don't pick it up. The only release path is
via the stuck-orchestration reaper acting on the OWNING orchestration
— and that path's effectiveness depends on the orchestration being
"poked" by some other activity, which doesn't always happen.

## What happened today (2026-05-19) — diagnostic value

A work item for `how-pricing-works` was claimed at 12:48:51. Observed
51 minutes later it was still claimed. No movement on the queue.

Then ~10 minutes after that, looking again: the queue had advanced,
multiple page-rerender pods had spawned, completed, and exited. The
original claim was either re-claimed by a different pod (same row, new
claim) or had progressed through processing.

Most likely chain:
1. Pod that claimed at 12:48:51 OOM'd or got evicted (no failure
   propagated)
2. Orchestration sat in EXECUTING_STEP for ~51 min, untouched
3. Something accessed the orchestration (probably the scheduler-
   triggered `build-pipeline-trigger` at 13:39 or similar)
4. Stuck-orchestration reaper fired, cleared EXECUTING_STEP
5. continueExecution re-ran the step, possibly with a fresh pod
6. Orchestration eventually completed or failed
7. Work item released/completed

The 51-minute delay is consistent with the "no periodic sweep" gap. The
recovery happened because activity returned to the area, not because a
reaper was actively watching.

## Recommendations (for future work)

### Short-term — add a work-item claim reaper

```go
// Run every 60 seconds
const ClaimTimeoutSeconds = 600  // 10 min — generous for any real work

// SQL — uses the existing idx_swi_claimed index
UPDATE site_work_items
SET status = 'triaged',
    claimed_by = NULL,
    claimed_at = NULL,
    attempt_count = attempt_count + 1,
    error = COALESCE(error, '') || ' [claim_reaper:released_at_' || NOW()::text || ']',
    updated_at = NOW()
WHERE status = 'claimed'
  AND claimed_at < NOW() - INTERVAL '10 minutes'
  AND attempt_count < max_attempts
RETURNING id, item_type, claimed_by;
```

Goes in a background goroutine in the chassis or as a separate CronJob.
The `attempt_count + 1` prevents infinite reclaim loops on broken items.

### Medium-term — periodic stuck-orchestration sweep

A background goroutine that scans `orchestration_states` for rows with
`status='EXECUTING_STEP' AND last_activity < NOW() - INTERVAL '5 min'`
and triggers ClearExecutingStep + continueExecution proactively, rather
than waiting for incidental access.

### Both should write metrics

When the reaper releases something, emit a Prometheus counter. If the
counter is non-zero, that's a signal that something upstream is failing
in a non-cleanly-reported way (likely OOM-kills based on today's
investigation) and deserves attention.

## How this interacts with the OOM issue

The 18MB `collected_data` causing component-quality-auditor OOMs (see
`TODO_orchestration_memory_bloat.md`) is the upstream cause for many
stuck claims today. Fixing the memory bloat reduces the rate at which
the reaper gap manifests. Adding a work-item claim reaper makes the
gap less painful when it does manifest. Both improvements are
complementary, not substitutes.
