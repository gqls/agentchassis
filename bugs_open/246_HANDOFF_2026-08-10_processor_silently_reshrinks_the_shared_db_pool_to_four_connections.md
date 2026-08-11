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

---

## 2026-08-11 — FIXED IN CODE (candidate 1), awaiting a roll. Taken by the shared-pool-ownership lane.

**Status: fix committed at `039cfce84` (⚠ first recorded here as `039fcce84` — a one-character
c/f transposition that `git show` refuses; corrected 2026-08-11 by the filing lane, which hit
the dead pointer while verifying), NOT yet live** — Go changes are inert until the
next whole-fleet roll. Council submission `c94d73ac-2a15-40cb-98a9-1185a2b7435a`
(`Council-Submitted:` trailer; verdict pending at commit time).
Working docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_246_shared_pool_ownership/`.

### The filing was right. Two things in it were not — both matter to whoever reads next

1. **The code quote in "The mechanism" is wrong.** This file says
   `maxOpen := 12 // production value via CHASSIS_DB_MAX_OPEN_CONNS`. `agent.go:277-289`
   actually reads `maxConns := 4`, with 12 arriving **only** from the environment. This is
   not pedantry: the misquote implies the code hard-codes 12 and that every agent is
   affected. It does not and they are not — and the true version is what makes the fix
   safe (below).
2. **"Anyone reading `CHASSIS_DB_MAX_OPEN_CONNS=12` in the overlay"** — it is not in the
   overlay. `kubectl kustomize …/agent-chassis/overlays/production/uk_001 | grep CHASSIS_`
   returns nothing. The key exists **only on the live Deployment object**. (It is *not*
   at risk from the next `apply -k`: it is absent from `last-applied-configuration`, so
   the three-way merge preserves it. Checked, because the alarming reading is the
   intuitive one.)

### What was confirmed, and how

- **The overwrite, proven in isolation with a control:** 12 → 4 reads 4; setting 9 reads 9,
  so the probe could have come out otherwise.
- **The env var IS set live:** both replicas carry `CHASSIS_DB_MAX_OPEN_CONNS=12` **and**
  `CHASSIS_INTAKE_MODE=worker_pool_all` — the fleet runs the worker-pool workload the raise
  was for, on the size the raise was meant to replace.
- **A third defect in the same constructor, not in this filing:** it also opens its *own*
  pool from `DATABASE_URL` with the error discarded and **no sizing at all**. Go's zero
  value for `MaxOpenConns` is 0 = **unlimited**. Fixed under the same rule — you size what
  you open.
- **A latent panic, removed for free:** `agentbase` leaves `a.db` nil when `DatabaseURL` is
  empty and `(*sql.DB)(nil).SetMaxOpenConns` **panics**. No agent is in that position today.

### Answers to this file's own "What to check before fixing"

- **Does any chassis-adjacent service construct a processor against a DB it does not own?**
  **No.** Exactly one non-test caller (`agent.go:341`), and exactly one binary imports
  `agentbase` (`cmd/agent-chassis`). The override was protecting nothing.
- **pgbouncer sizing:** `pool_mode = transaction`, `max_client_conn = 200`,
  `default_pool_size = 15`, `reserve_pool_size = 5`. 2 replicas × 12 = 24 client
  connections against a 200 ceiling, and transaction pooling is precisely the mechanism
  that absorbs idle clients — so it does **not** just move the queue. Caveat stated rather
  than buried: `default_pool_size = 15` is a **server-side** ceiling shared by all clients
  of `clients_user`/`clients_db`. This change does not create that ceiling; it lets the
  chassis reach it. The configmap's own comment reasons from "3 chassis replicas × 4
  conns" — stale on both numbers, and its owners have been told.
- **Live evidence before changing the pool**, as this file asks: `DISPATCH_LOOKUP_RETRYABLE`
  reads **0**. **Reported as a NON-RESULT.** Demand control: `chassis_intake_events` shows
  60–100 events/hour, i.e. 1–2 messages/minute, which cannot saturate even a 4-connection
  pool. The zero is equally consistent with "4 is fine at this load" and "4 would freeze
  under load". It discriminates nothing, and nothing surfaces
  `db.Stats().WaitCount`/`WaitDuration`, so **no instrument can see pool saturation today.**

### Severity, settled

This file said "Severity: unknown, and that is the point." That is still the right answer.
It is a **latent config-inertness defect, not an observed outage** — the case for fixing it
is ownership correctness, not performance, and any submission arguing impact would be
arguing something the platform cannot evidence.

### Blast radius (why candidate 1 is safe)

`agentbase`'s defaults are 4 / 1 / 10m; the processor's constants were 4 / 1 / 10m —
**identical**. So the deletion is a behavioural no-op wherever the variable is unset.
Measured live: **2 pods of 95** carry it (both static chassis replicas), **1 deployment of
24**, and `personae-prod-config` — what spawned agents inherit — carries **no `CHASSIS_*`
keys at all**. 93 of 95 pods are byte-identical; the 2 that change get the size their
operator already asked for.

### Post-roll verification (there is no way to observe a pool's size)

1. Confirm the commit shipped: `kubectl … logs -l app=agent-chassis | grep -m1 'build provenance'`,
   then `git merge-base --is-ancestor 039fcce84 <stamp>`. **Per service, not per fleet.**
2. The disconfirming observation for the pgbouncer risk: `SHOW POOLS` with `cl_waiting`
   sustained > 0, or `maxwait` climbing.
3. **Do not claim to have "observed the pool at 12".** Nothing reports a pool's size. The
   unit test is the proof of the mechanism; the stamp is the proof it shipped.

### Follow-ups deliberately NOT done here

- **Collapse `p.db` / `p.sqlDB`.** `sqlDB` is nil in production and all eight readers already
  carry a `p.db` fallback. Separate change, own review. Related to `bugs_open/247`'s
  dead-code family.
- **Surface `db.Stats()`** (`WaitCount`, `WaitDuration`) so this class is measurable at all.
  Kept out deliberately so the blast-radius argument above stays checkable.
- **Narrow `db` to a query-only interface** so the defect is structurally unrepresentable.
  Rejected for now: `p.db` flows into `orchestration.NewStateRepository(*sql.DB)` and
  `NewSagaCoordinator(*sql.DB)`, both public and used platform-wide, so it is a large
  refactor of shared signatures for a three-line defect with one caller. Revisit if a
  second caller appears.
