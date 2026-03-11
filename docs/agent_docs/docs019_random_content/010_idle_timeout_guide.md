# Idle Timeout for Spawned Agents

## Problem

Every time the build-dispatch-loop (or any orchestrator) delegates work, it spawns a K8s Job. The Job creates a pod running the agent-chassis binary. The agent processes its workflow — loading data, calling LLMs, writing to the database — then completes its orchestration and sends a response back to its parent.

At that point the workflow is done. But the pod doesn't exit. It sits in its Kafka consumer loop calling `Consume()` with a 5-second timeout, getting `DeadlineExceeded`, and looping again. Forever.

The K8s Job has `TTLSecondsAfterFinished: 3600` which would clean up the Job and pod — but only after the Job *completes*. A Job completes when its pod exits. The pod never exits. So the TTL never fires.

The result: every scheduled trigger spawns new pods that never die. Over 13 hours, `build-pipeline-trigger` firing every 120 seconds created 390+ dispatch loop pods, each of which spawned handler pods, content writer pods, and research agent pods. The cluster ran out of CPU.

## Solution

An idle monitor goroutine inside the agent that shuts down the process after a configurable period of inactivity.

### How it works

```
Agent starts
  ├── processRequests()   ← updates lastActivity on every message
  ├── processResponses()  ← updates lastActivity on every message
  ├── healthCheck()
  ├── sendHeartbeats()    ← if spawned
  └── idleMonitor()       ← NEW: if IDLE_TIMEOUT_SECONDS > 0
        │
        ├── checks time.Since(lastActivity) every 10s
        │
        └── if idle >= timeout:
              close(shutdownChan) → Run() returns → main() exits → pod exits
                                                      → Job completes → TTL cleans up
```

The idle timer resets on every real Kafka message (request or response). A multi-step workflow that takes 2 minutes with gaps between steps stays alive because responses keep arriving. The 120-second timeout only triggers after the *last* message — when the workflow is done and no more messages will arrive.

### Configuration

The timeout is stored in `agent_definitions.idle_timeout_seconds`:

| Value | Meaning | Used by |
|-------|---------|---------|
| 0 | No timeout — run forever | Deployment agents (chassis, core-manager, adapters) |
| 120 | Exit after 2 minutes idle | All Job-spawned agents |

The value flows through three layers:

1. **Database**: `agent_definitions.idle_timeout_seconds` column
2. **Job spawner**: reads the column, passes `IDLE_TIMEOUT_SECONDS` env var on the container
3. **Agent process**: reads the env var, starts the idle monitor goroutine

Deployment agents (which run as K8s Deployments, not Jobs) have `idle_timeout_seconds = 0`. The spawner never sets the env var, the monitor never starts, and the agent runs forever — same as before.

### Shutdown safety

Two things can trigger shutdown: the idle monitor and a SIGTERM from K8s. Both call `close(shutdownChan)`. Closing an already-closed channel panics in Go. A `sync.Once` on the Agent struct protects against this:

```go
a.shutdownOnce.Do(func() {
    close(a.shutdownChan)
})
```

Both `Shutdown()` (called from signal handler) and `idleMonitor()` use this wrapper.

### main.go fix

The existing `main.go` only sends to `errCh` when `Run()` returns an error. A clean idle shutdown returns `nil`, so nothing is sent and `main()` blocks on `select` forever — the pod stays alive even though Run returned. The fix sends the return value unconditionally:

```go
// Before
go func() {
    if err := agent.Run(); err != nil {
        errCh <- err
    }
}()

// After
go func() {
    errCh <- agent.Run()
}()
```

## Files changed

| File | Changes |
|------|---------|
| `agent_definitions` table | New column `idle_timeout_seconds int NOT NULL DEFAULT 0` |
| `platform/orchestration/actions/spawn_actions.go` | `AgentDefinition` struct gets `IdleTimeoutSeconds` field; `loadAgentDefinitionFromDB` query and scan updated; env var injected into Job container |
| `platform/agentbase/agent.go` | `shutdownOnce` field on Agent struct; `lastActivity` updated in both consumer loops; new `idleMonitor()` method; `Shutdown()` uses `sync.Once`; monitor started in `Run()` |
| `cmd/agent-chassis/main.go` | `errCh` always receives from `Run()`; clean exit handled without error log |

## Deploy order

1. Run SQL migration (`add_idle_timeout_column.sql`) — adds column, sets values
2. Build and deploy Go changes — new binary reads the column and honours the env var
3. Kill existing lingering Jobs — they're running the old binary without the monitor

Existing pods won't gain the idle timeout (they don't have the env var). They'll age out naturally via the TTL once they're eventually killed or the node recycles. New pods spawned after the deploy will auto-exit.

## Tuning

120 seconds is the initial value. If agents have legitimate gaps longer than 120s between messages (e.g. waiting for a slow LLM call routed through another agent), increase per agent type:

```sql
UPDATE agent_definitions
SET idle_timeout_seconds = 300
WHERE type = 'research-agent';
```

The change takes effect on the next spawn — no binary redeploy needed.

To disable for a specific agent type (make it run forever like a Deployment):

```sql
UPDATE agent_definitions
SET idle_timeout_seconds = 0
WHERE type = 'some-long-running-agent';
```

## Observability

The idle monitor logs at startup and shutdown:

```json
{"level":"info","msg":"Idle monitor started","timeout":"2m0s","check_interval":"10s"}
```

```json
{"level":"info","msg":"Idle timeout reached — shutting down","idle_duration":"2m10s","timeout":"2m0s","last_activity":"2026-03-11T10:45:00Z"}
```

To check for pods that should have exited but didn't (running old binary):

```bash
kubectl -n ai-persona-system get pods --sort-by=.metadata.creationTimestamp | grep agent- | awk '$5 ~ /[0-9]+h/ {print $1, $5}'
```

Any agent pod older than a few hours is likely an orphan from before the deploy.
