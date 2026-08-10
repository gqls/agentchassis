# 246 — `NewMessageProcessor` unconditionally re-shrinks the SHARED `*sql.DB` to 4 connections, silently undoing `CHASSIS_DB_MAX_OPEN_CONNS`

**Filed 2026-08-10** by the `bugfix_239` lane, found while tracing why an agent-definition
lookup could fault transiently (`bugs_open/239`'s transient arm). **Status: OPEN, not
fixed — this lane changed no pool settings.** Severity: unknown, and that is the point —
it is a config key that reads as live and is not.

## The mechanism

`platform/agentbase/agent.go:277-289` sizes the database pool from the environment:

```go
maxOpen := 12 // production value via CHASSIS_DB_MAX_OPEN_CONNS
...
db.SetMaxOpenConns(maxOpen)
```

with a comment that explains exactly why it matters: *"4 workers + two consume loops + the
response path + the retry driver against 4 connections is a queue, and a freeze if anything
holds a transaction while acquiring a second conn."*

Then `platform/messaging/processor.go:66-74`, constructed LATER (`agent.go:341`) against
**the same `*sql.DB` handle**, does:

```go
db.SetMaxOpenConns(4)
db.SetMaxIdleConns(1)
db.SetConnMaxLifetime(time.Minute * 10)
```

unconditionally. `*sql.DB` is a pool object, not a connection, so the second call is not a
second pool — it re-sizes the first. **The env-driven value survives for the few
milliseconds between the two constructors.**

## Why it is worth a bug rather than a one-line fix by this lane

1. **The chassis_replica_scaling (CS-2) work raised this value deliberately**, and its
   raise has been inert since. Anyone reading `CHASSIS_DB_MAX_OPEN_CONNS=12` in the
   overlay, or the comment above it, will believe the fleet runs 12.
2. **`ConnMaxLifetime(10m)` forces a connection recycle on a ten-minute clock**, which is
   the same order as the "byte-identical dispatches diverged ~15 minutes apart" observation
   in `bugs_open/239`. That turned out not to be 239's cause, but a 4-connection pool with
   forced recycling under a spawn-heavy workload is a plausible source of the transient
   lookup faults 239's fix now (correctly) retries rather than swallows.
3. Which value is RIGHT is not obvious and is not this lane's call: the processor's 4 may
   have been chosen for a reason nobody wrote down, and raising it fleet-wide touches
   pgbouncer's own limits (`patch-deployment.yaml:49`).

## What to check before fixing

- Whether any chassis-adjacent service constructs a processor against a DB it does NOT own
  (i.e. whether the override was protecting something).
- `pgbouncer`'s pool sizing, so raising the client-side cap does not just move the queue.
- Live evidence of contention before and after: `pg_stat_activity` waiting counts, and
  chassis logs for lookup faults now that 239's fix names them
  (`DISPATCH_LOOKUP_RETRYABLE`) instead of silently substituting a no-op. **That log line
  is the first instrument this class has ever had** — measure with it before changing the
  pool, or the change cannot be shown to have done anything.

## Fix candidates, ordered by what closes the door

1. **Delete the processor's three calls.** The pool belongs to whoever opened it; a
   constructor that mutates a handed-in shared resource is the defect, independent of the
   numbers. Sizing then has exactly one home.
2. If the processor genuinely needs a floor, have it RAISE only (`if current < n`), which
   needs a getter it does not have — `db.Stats().MaxOpenConnections` — and is strictly
   worse than (1) because two owners of one setting is the shape that produced this.
3. Do nothing but document it in the overlay next to the env var. Rejected as a candidate,
   listed so the reader knows it was considered: a comment is not a control.
