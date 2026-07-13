# Reapers and Stuck-State Recovery in the Chassis

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
