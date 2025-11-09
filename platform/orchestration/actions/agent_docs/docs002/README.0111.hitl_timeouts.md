# HITL Approval Timeout Analysis and Long-term Configuration

## Current Architecture Overview

### How HITL Timeout Works

1. **Workflow Configuration Level**
    - Timeouts are specified in the workflow step config as `timeout_seconds` (integer)
    - Example from logs: `"timeout_seconds": 86400` (24 hours)

2. **Flow Through System**
   ```
   Workflow Config (timeout_seconds: 86400)
     ↓
   AwaitApprovalAction (buildApprovalNotification)
     ↓
   Notification sent to UI with timeout_seconds field
     ↓
   processAwaitResponse (creates AwaitedRequest)
     ↓
   AwaitedRequest.TimeoutAt = Now + getTimeout(step)
     ↓
   handleRequestTimeout goroutine launched
   ```

3. **Current Timeout Sources**
    - **Step.Timeout field** (time.Duration): Used by `getTimeout(step)` to set AwaitedRequest.TimeoutAt
    - **Step.Config["timeout_seconds"]**: Currently sent in notification but NOT mapped to Step.Timeout
    - **DefaultRequestTimeout**: 180 seconds (3 minutes) - fallback when Step.Timeout is 0

### Current Problem

The `timeout_seconds` from the workflow config is:
- ✅ Sent in the UI notification (so the UI knows about it)
- ❌ NOT mapped to the `Step.Timeout` field
- ❌ NOT stored in the database `approval_requests` table (table doesn't have this field)
- ❌ NOT used when setting `AwaitedRequest.TimeoutAt`

**Result**: All approval requests currently timeout after 180 seconds (3 minutes) regardless of the workflow config, because `Step.Timeout` is 0 and defaults to `DefaultRequestTimeout`.

## Required Changes for Long-term Solution

### 1. Database Schema Change

Add `timeout_seconds` field to the `approval_requests` table:

```sql
ALTER TABLE approval_requests 
ADD COLUMN timeout_seconds INTEGER DEFAULT 180,
ADD COLUMN timeout_at TIMESTAMP;
```

This allows:
- Recording the configured timeout for each approval request
- Querying for approvals approaching timeout
- Historical analysis of approval times vs configured timeouts

### 2. Workflow Parsing Enhancement

The system needs to map `config.timeout_seconds` to `Step.Timeout` when loading workflows.

**Location**: Where workflow steps are unmarshalled from JSON (need to add UnmarshalJSON custom logic)

```go
// When parsing step config, convert timeout_seconds to Step.Timeout
if timeoutSecs, ok := config["timeout_seconds"].(float64); ok {
    step.Timeout = time.Duration(timeoutSecs) * time.Second
}
```

**Alternative Approach**: Add conversion in the coordinator when loading steps:
```go
// In coordinator.go, before executing step
if step.Timeout == 0 && step.Config != nil {
    if timeoutSecs, ok := step.Config["timeout_seconds"].(float64); ok {
        step.Timeout = time.Duration(timeoutSecs) * time.Second
    }
}
```

### 3. Approval Request Storage Implementation

Currently `storeApprovalRequest` is a stub. Implement it to store:

```go
func storeApprovalRequest(
    ctx context.Context, 
    db interface{}, 
    requestID string, 
    execCtx *types.ExecutionContext, 
    data map[string]interface{},
    timeoutSeconds int,  // ADD THIS PARAMETER
    logger *zap.Logger
) error {
    // Get DB connection
    pgDB, ok := db.(*sqlx.DB)
    if !ok {
        return fmt.Errorf("invalid database type")
    }
    
    // Calculate timeout_at
    timeoutAt := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
    
    // Insert into approval_requests
    query := `
        INSERT INTO approval_requests (
            request_id, orchestration_id, correlation_id, 
            agent_type, agent_id, step_name, data, 
            status, timeout_seconds, timeout_at, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
    `
    
    dataJSON, _ := json.Marshal(data)
    
    _, err := pgDB.ExecContext(ctx, query,
        requestID,
        execCtx.OrchestrationID,
        execCtx.CorrelationID,
        execCtx.Sender.AgentType,
        execCtx.Sender.AgentID,
        execCtx.StepName,
        dataJSON,
        "pending",
        timeoutSeconds,
        timeoutAt,
    )
    
    return err
}
```

### 4. Update AwaitApprovalAction to Pass Timeout

Modify `AwaitApprovalAction` to extract and pass timeout:

```go
// In AwaitApprovalAction function, after line 50616
timeoutSeconds := 86400 // Default 24 hours
if timeout, ok := params.StepConfig.Config["timeout_seconds"].(float64); ok {
    timeoutSeconds = int(timeout)
}

// When calling storeApprovalRequest (line 50663)
err = storeApprovalRequest(ctx, params.DB, approvalToken, 
    params.ExecutionContext, dataForApproval, timeoutSeconds, params.Logger)
```

### 5. Timeout Monitoring Enhancement

The `handleRequestTimeout` function already works correctly using `AwaitedRequest.TimeoutAt`.
Once Step.Timeout is properly set from config, the timeout handling will work for longer periods.

**Important**: The goroutine `time.Sleep(time.Until(timeoutAt))` can handle multi-day waits.

### 6. Configuration Validation

Add validation to prevent unreasonable timeouts:

```go
const (
    MinApprovalTimeout = 60        // 1 minute
    MaxApprovalTimeout = 604800    // 7 days
    DefaultApprovalTimeout = 86400 // 24 hours
)

func validateApprovalTimeout(timeout int) int {
    if timeout < MinApprovalTimeout {
        return MinApprovalTimeout
    }
    if timeout > MaxApprovalTimeout {
        return MaxApprovalTimeout
    }
    return timeout
}
```

## Implementation Priority

### Phase 1 (Immediate - Fixes Current Issue)
1. ✅ Add timeout conversion when loading/executing steps
    - Convert `step.Config["timeout_seconds"]` to `step.Timeout`
    - Ensures existing workflow configs work correctly

2. ✅ Test with multi-day timeouts
    - Verify goroutine handles long sleeps
    - Verify state persistence across pod restarts

### Phase 2 (Near-term - Database Support)
1. ✅ Add database schema changes
2. ✅ Implement `storeApprovalRequest` properly
3. ✅ Implement `updateApprovalRequest` properly

### Phase 3 (Future - Enhanced Features)
1. Timeout warning notifications (e.g., "approval needed in 2 hours")
2. Timeout escalation (notify additional approvers)
3. Approval request dashboard with timeout visualization

## Testing Recommendations

### Test Cases
1. **Short timeout (3 minutes)**: Verify default behavior unchanged
2. **Medium timeout (1 hour)**: Verify proper waiting
3. **Long timeout (2 days)**: Verify multi-day waits work
4. **Pod restart during wait**: Verify state persistence and timeout recovery
5. **Approval before timeout**: Verify normal flow
6. **Approval after timeout**: Verify timeout handling

### Test Workflow Example
```json
{
  "start_step": "generate",
  "steps": {
    "generate": {
      "action": "execute_llm_prompt",
      "config": {
        "prompt_template": "Generate content for {{.topic}}"
      },
      "next_step": "await_approval"
    },
    "await_approval": {
      "action": "await_approval",
      "config": {
        "timeout_seconds": 172800,
        "approval_type": "content_review",
        "ui_config": {
          "title": "48-Hour Content Review",
          "description": "Please review within 48 hours"
        }
      },
      "next_step": "process_approval"
    },
    "process_approval": {
      "action": "process_approval_decision",
      "next_step": "done"
    },
    "done": {
      "action": "complete_workflow"
    }
  }
}
```

## State Persistence Considerations

For multi-day waits, consider:

1. **Timeout Goroutines and Pod Restarts**:
    - Current: Goroutines are lost on pod restart
    - Solution: On agent startup, check for existing AwaitedRequests and restart timeout goroutines

```go
// In agent initialization
func (c *SagaCoordinator) recoverPendingTimeouts(ctx context.Context) {
    states, _ := c.repo.GetActiveOrchestrations(ctx)
    for _, state := range states {
        for reqID, awaited := range state.AwaitedRequests {
            if time.Now().Before(awaited.TimeoutAt) {
                go c.handleRequestTimeout(ctx, state.OrchestrationID, reqID, awaited.TimeoutAt)
            }
        }
    }
}
```

2. **Database vs In-Memory State**:
    - AwaitedRequests are in OrchestrationState (stored in DB)
    - Timeout goroutines are in-memory only
    - Need timeout recovery on startup

## Summary

To support 1-2 day approval timeouts:

**Critical Path**:
1. Map `config.timeout_seconds` to `step.Timeout` when executing steps
2. Test with long timeouts (existing infrastructure should work)
3. Implement timeout recovery on agent restart

**Good to Have**:
1. Database schema update for approval_requests
2. Full implementation of storage functions
3. Validation and monitoring enhancements

The architecture is mostly ready for long timeouts - the main missing piece is converting the workflow config's `timeout_seconds` to the `Step.Timeout` field that the system actually uses.