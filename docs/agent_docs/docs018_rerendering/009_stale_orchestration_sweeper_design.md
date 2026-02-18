# Stale Orchestration Sweeper — Design

## Problem

Timeout handling uses in-process goroutines (`go handleRequestTimeout(...)`).
These die when pods restart, leaving orchestrations stuck in AWAITING_RESPONSES
forever. This is the #1 cause of pipeline stalls.

Three failure modes we've hit repeatedly:
1. **Response lost** — child completed but Kafka message never received (topic
   issue, message too large, consumer not listening)
2. **Timeout goroutine lost** — pod that started the goroutine restarted
3. **Child stuck** — child itself is in AWAITING_RESPONSES (cascading stall)

## Approach

A periodic DB sweep running on existing agent-chassis pods. No new service.
Uses `FOR UPDATE SKIP LOCKED` so multiple pods can run the sweep safely —
each pod claims different expired requests.

## Where It Runs

Add to the agent-chassis startup as a background goroutine, alongside the
existing Kafka consumer loop. Every chassis pod runs it. The `FOR UPDATE
SKIP LOCKED` pattern ensures no double-processing.

```go
// In agent startup (agentbase/agent.go), after consumer starts:
go sweeper.RunStaleRequestSweeper(ctx, db, producer, logger)
```

## Sweep Logic

Runs every 60 seconds. Each cycle:

### Step 1: Find expired awaited requests

```sql
SELECT ar.request_id, ar.orchestration_id, ar.correlation_id,
       ar.step_id, ar.step_name, ar.retry_version,
       ar.responses_topic, ar.requests_topic,
       ar.timeout_at, ar.target_agent_type
FROM awaited_requests ar
WHERE ar.status = 'waiting'
  AND ar.timeout_at < NOW() - INTERVAL '30 seconds'
ORDER BY ar.timeout_at ASC
LIMIT 20
FOR UPDATE SKIP LOCKED
```

The 30-second grace period avoids racing with in-process goroutines that
might still be about to fire. LIMIT 20 prevents one sweep from taking too
long.

### Step 2: For each expired request, classify the situation

```sql
-- Check if the child orchestration completed
SELECT os.orchestration_id, os.status, os.final_result
FROM orchestration_states os
WHERE os.orchestration_id = (
    SELECT child_os.orchestration_id
    FROM orchestration_states child_os
    WHERE child_os.parent_orchestration_id = $parent_orch_id
      AND child_os.status IN ('COMPLETED', 'FAILED')
    ORDER BY child_os.updated_at DESC
    LIMIT 1
)
```

Three outcomes:

**A. Child COMPLETED — response was lost**
The child did its work but the parent never got the message.
→ Synthesize a completion response from the child's `final_result`
→ Produce it to the parent's `responses_topic`
→ Mark awaited_request as `status = 'processed'`

**B. Child FAILED**
→ Forward the failure to the parent
→ Parent's existing error handling (continue_on_error, fail workflow) takes over

**C. No child found, or child still running**
→ Increment `retry_version`
→ If `retry_version < 3`: re-send the original request (retry)
→ If `retry_version >= 3`: mark as expired, fail the parent orchestration
(or skip loop iteration if continue_on_error)

### Step 3: Handle the response

For case A (most common — the response-lost scenario):

```go
func synthesizeCompletionResponse(
    awaited *AwaitedRequest,
    childState *OrchestrationState,
) *kafka.Message {
    // Build a response message that looks like a normal child completion
    response := map[string]interface{}{
        "headers": map[string]interface{}{
            "correlation_id":            awaited.CorrelationID,
            "orchestration_id":          childState.OrchestrationID,
            "in_response_to_request_id": awaited.RequestID,
            "in_response_to_step_id":    awaited.StepID,
            "in_response_to_step_name":  awaited.StepName,
            "message_type":              "response",
            "status":                    "complete",
            "sender_agent_type":         childState.OwnerAgentType,
            "sender": map[string]interface{}{
                "agent_type": childState.OwnerAgentType,
                "agent_id":   childState.OwnerAgentID,
                "pod_name":   "stale-sweeper",
            },
        },
        "body": childState.FinalResult,
    }
    // Produce to awaited.ResponsesTopic
    return buildKafkaMessage(awaited.ResponsesTopic, response)
}
```

For case C (retry):

```go
// Reuse existing handleRequestTimeout logic but triggered from DB sweep
// instead of in-process goroutine
func retryExpiredRequest(awaited *AwaitedRequest, state *OrchestrationState) {
    // This is essentially the same as the existing handleRequestTimeout
    // but driven by DB state instead of a goroutine timer
    
    // Increment retry_version in DB
    // Re-execute the step (re-send the request to the target)
    // Set new timeout_at
}
```

## Concurrency Safety

- `FOR UPDATE SKIP LOCKED` on awaited_requests ensures only one pod
  handles each expired request
- Version check on orchestration_states before updating (existing pattern)
- The synthesized response goes through normal Kafka message handling,
  which already handles deduplication via request_id matching

## What Changes

### New file: `platform/orchestration/sweeper.go`

```go
package orchestration

type StaleRequestSweeper struct {
    db       *sql.DB
    producer kafka.Producer
    logger   *zap.Logger
    repo     *StateRepository
    interval time.Duration  // default 60s
    podName  string
}

func NewStaleRequestSweeper(db *sql.DB, producer kafka.Producer, logger *zap.Logger) *StaleRequestSweeper

func (s *StaleRequestSweeper) Run(ctx context.Context)
// Main loop: tick every s.interval, call s.sweep()

func (s *StaleRequestSweeper) sweep(ctx context.Context) error
// Single sweep cycle: find expired, classify, handle

func (s *StaleRequestSweeper) handleExpiredRequest(ctx context.Context, tx *sql.Tx, awaited *AwaitedRequest) error
// Classify and handle one expired request

func (s *StaleRequestSweeper) synthesizeResponse(awaited *AwaitedRequest, childState *OrchestrationState) error
// Build and produce a synthetic response for case A

func (s *StaleRequestSweeper) retryRequest(ctx context.Context, awaited *AwaitedRequest) error
// Re-execute the step for case C (reuses existing retry logic)

func (s *StaleRequestSweeper) failOrchestration(ctx context.Context, awaited *AwaitedRequest) error
// Max retries exceeded for case C
```

### Modified: `agentbase/agent.go`

Add sweeper startup in the Run() method:

```go
// Start stale request sweeper
sweeper := orchestration.NewStaleRequestSweeper(a.db, a.producer, a.logger)
go sweeper.Run(ctx)
```

### No schema changes needed

`awaited_requests` already has everything:
- `status = 'waiting'` + `timeout_at` for finding expired requests
- `retry_version` for tracking retries
- `idx_awaited_requests_cleanup` index for the sweep query
- `processing_pod` column for tracking which pod is handling it

## Logging

Each sweep cycle logs:
- Number of expired requests found
- For each: action taken (synthesize/retry/fail), orchestration_id, step_name
- Any errors

```
INFO  "Stale sweep: found 3 expired requests"
INFO  "Stale sweep: synthesized response for run_sweeps" orch_id=34c15d37 child_status=COMPLETED
WARN  "Stale sweep: retrying request (attempt 2)" orch_id=549c42d3 step=scrape_website
ERROR "Stale sweep: max retries exceeded, failing orchestration" orch_id=... 
```

## Edge Cases

1. **Race with normal response delivery**: Sweeper synthesizes a response,
   but the real response also arrives. The parent's normal dedup logic
   (checking if request_id is still in awaited_requests) handles this —
   second response is ignored.

2. **Parent already failed/completed**: Check parent status before
   synthesizing. If parent is already COMPLETED/FAILED, just mark the
   awaited_request as processed and move on.

3. **Cascading stalls**: Parent waiting for child, child waiting for
   grandchild. The sweeper handles this naturally — it processes the
   deepest (oldest timeout_at first) expired request first. Next cycle
   picks up the parent.

4. **Job topics expired**: Job-specific Kafka topics have 1hr retention.
   If the parent's response topic is gone, the synthesized message fails.
   In this case, directly update the parent's orchestration_state to
   advance it past the awaited step (write the child's result into
   collected_data and set current_step to next_step).

## Metrics (future)

- `stale_sweep_expired_found` — counter per cycle
- `stale_sweep_synthesized` — counter of case A
- `stale_sweep_retried` — counter of case C
- `stale_sweep_failed` — counter of permanent failures
- `stale_sweep_duration_ms` — histogram of sweep cycle time