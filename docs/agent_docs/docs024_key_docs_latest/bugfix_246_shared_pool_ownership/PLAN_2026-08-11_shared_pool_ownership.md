# PLAN — bugfix 246: the shared `*sql.DB` has one owner, and it is not the processor

**Date:** 2026-08-11 · **Bug:** `bugs_open/246` · **Lane:** this session
**Status:** plan settled, council submission next.

---

## The recommendation, in three sentences

Delete the three pool-sizing calls that `NewMessageProcessor` makes on the `*sql.DB`
it is **handed**, keeping the principle "**you size what you open; you never size
what you are given**" — and apply the same principle to the *other* pool in that
constructor, the one it opens itself from `DATABASE_URL`, which today is created
with its error discarded and **no sizing at all** (Go's default is unlimited).
Add a regression test that calls the real constructor and asserts the caller's
sizing survived, because every existing test in this package builds
`&MessageProcessor{…}` as a struct literal and therefore **structurally cannot
catch a defect that lives in the constructor**. Do **not** narrow `db` to a
query-only interface: it is the theoretically strongest fix and it is
disproportionate here, for reasons given under "rejected alternatives".

## 1. The core change — `platform/messaging/processor.go`

Remove, from `NewMessageProcessor`:

```go
db.SetMaxOpenConns(4)
db.SetMaxIdleConns(1)
db.SetConnMaxLifetime(time.Minute * 10)
```

These act on the parameter `db`, which `platform/agentbase/agent.go` opened and
already sized at :277-289 from `CHASSIS_DB_MAX_OPEN_CONNS`. A `*sql.DB` is a pool
object, so this is not a second pool — it is a re-size of the caller's.

A comment goes in their place naming the rule and the bug, so the next reader knows
the absence is deliberate rather than an oversight.

## 2. The second pool in the same constructor — in scope, and consistent

```go
var sqlDB *sql.DB
if connStr := os.Getenv("DATABASE_URL"); connStr != "" {
    sqlDB, _ = sql.Open("pgx", connStr)   // error discarded; pool never sized
}
```

This one the constructor **does** own, so under the stated principle it is the
constructor's job to size it — and it does not. Go's zero value for
`MaxOpenConns` is **0, meaning unlimited**, so on any deployment that sets
`DATABASE_URL` this is an unbounded pool behind a transaction-mode pgbouncer with
`max_client_conn = 200`. Fix: stop discarding the error, and size the pool we open
with the same 4/1/10m the rest of the estate uses.

Including this is not scope creep — it is the *same* judgement applied to the other
half of the same constructor, and leaving it would ship a change whose stated
principle its own file immediately violates.

**Deliberately NOT in scope:** removing `p.sqlDB` altogether. It is read at eight
sites (`processor.go:318, 548, 615-624, 1156, 1570-1581`, `validation_drop.go:79-81`)
and a test documents the intent (`processor_dispatch_resolution_test.go:433-453`).
Ripping it out is a separate, larger change with its own review. **Follow-up bug to
file:** "`p.sqlDB` is nil in production and every reader carries a `p.db` fallback —
collapse the two handles into one." Related to `bugs_open/247`'s dead-code family.

## 3. The regression test — behavioural, not a source scan

`platform/messaging/processor_pool_ownership_test.go`:

1. `sql.Open("pgx", "postgres://u:p@127.0.0.1:1/none")` — builds the pool object
   without connecting, so no database is required.
2. `db.SetMaxOpenConns(12)` — stand in for the operator's configured value.
3. Call **`NewMessageProcessor`** with `zap.NewNop()` and nil for producer,
   orchestrator, validator and initializer. Verified safe: the constructor
   dereferences none of them, and `orchestration.NewStateRepository(nil, logger)`
   only stores its arguments (`state.go:174-176`).
4. Assert `db.Stats().MaxOpenConnections == 12`.
5. **Include a control** in the same test: set it to 9 first and assert 9, so a
   probe that always returns the expected number cannot pass. (This is the discipline
   that made the original proof trustworthy — without it, "it printed 12" is
   compatible with a broken assertion.)
6. `t.Setenv("DATABASE_URL", "")` so the second-pool branch is not taken, and a
   sibling case that DOES set it and asserts the second pool comes out sized rather
   than unlimited.

**Why a source-scanning test was rejected:** this repo has a recorded lesson that
source-scanning tests make comments load-bearing (memory:
`a-source-scanning-test-makes-comments-load-bearing`). Grepping the package for
`SetMaxOpenConns` would fail the moment someone legitimately sizes a pool they own —
which is precisely what item 2 above does.

## 4. Blast radius — measured, not argued

The key property, and it is checkable: **`agentbase`'s defaults and the processor's
constants are identical** — 4 / 1 / 10 minutes on both sides. So for any agent that
does not set `CHASSIS_DB_MAX_OPEN_CONNS`, deleting the processor's calls is a
**behavioural no-op**; behaviour changes *only* where an operator deliberately set
the variable.

Measured 2026-08-11, live:

| check | result |
|---|---|
| pods carrying `CHASSIS_DB_MAX_OPEN_CONNS` | **2 of 95** — both static `agent-chassis` replicas, value `12` |
| deployments carrying it | **1 of 24** — `agent-chassis` |
| `personae-prod-config` (what spawned agents inherit) | **no `CHASSIS_*` keys at all** |
| binaries importing `agentbase` | **1** — `cmd/agent-chassis/main.go` |
| non-test callers of `NewMessageProcessor` | **1** — `agent.go:341` |

So: **93 of 95 pods see byte-identical behaviour**, and the two that change get the
pool size their operator already asked for. This is the "check the no-op case, not
only the damage case" discipline, and here it is the strongest argument for the fix.

**One more thing the change removes, free:** `agentbase` leaves `a.db` nil when
`DatabaseURL` is empty, and `(*sql.DB)(nil).SetMaxOpenConns` **panics** (proven). No
agent is in that position today, so it has never fired — but the crash goes away.

## 5. Other consumers of the seam — to be TOLD, not merely measured

(Owner ruling 2026-07-29 §3: measuring that nothing breaks does not establish that
the other owners would have agreed.)

- **pgbouncer.** `pool_mode = transaction`, `max_client_conn = 200`,
  `default_pool_size = 15`, `reserve_pool_size = 5`. Its configmap comment sizes the
  fleet from *"3 chassis replicas × 4 conns"*. **What changes for them:** the chassis's
  client-side cap becomes the configured 12 instead of a silent 4, so 2 × 12 = 24
  client connections against a 200 ceiling. Transaction pooling absorbs idle clients,
  so this does not simply move the queue — but `default_pool_size = 15` is a
  **server-side ceiling shared by every client of `clients_user`/`clients_db`**, and
  the chassis may now draw a larger share of it. That ceiling is not created by this
  change; the honest statement is that this change lets the chassis reach it.
- **The `chassis_replica_scaling` (CS-2) lane.** **What changes for them:** the
  `CHASSIS_DB_MAX_OPEN_CONNS=12` they set has been inert since they set it. Their
  worker-pool sizing was tuned against a pool they believed was 12 and was 4.
- **The `bugfix_239` lane**, who filed this bug and left the
  `DISPATCH_LOOKUP_RETRYABLE` instrument. **What changes for them:** their transient
  lookup faults had a plausible contributing cause that is now removed, so their
  instrument's future readings are against a different pool.

## 6. Risks and honest unknowns

- `[UNMEASURED]` **Does a 4-connection pool actually throttle the chassis today?**
  Unknown, and unknowable with current instrumentation. `DISPATCH_LOOKUP_RETRYABLE`
  reads zero, but at the observed **1–2 messages/minute** a 4-connection pool cannot
  saturate, so that zero **discriminates nothing**. Reported as a non-result.
- `[UNMEASURED]` **Is 12 safe against `default_pool_size = 15`?** Under transaction
  pooling, server connections are held only for the duration of a transaction, so 24
  mostly-idle clients do not map to 24 server connections. The disconfirming
  observation would be pgbouncer `SHOW POOLS` showing `cl_waiting > 0` sustained, or
  `maxwait` climbing, after the roll. **This is the check to run post-deploy.**
- **No instrument sees pool saturation.** `db.Stats()` carries `WaitCount` and
  `WaitDuration`; nothing surfaces them. Proposed as a **follow-up, not this change** —
  adding metrics is a separate concern from fixing ownership, and bundling them would
  make the blast-radius argument above harder to check.
- **The fix cannot be proven live by observation.** There is no log line or query that
  reports a pool's size. Post-roll verification is therefore: build provenance stamp
  confirms the commit shipped, plus the unit test. Anyone claiming to have "observed
  the pool at 12" in production is claiming something the platform cannot show.

## 7. Ordering

1. Commit the code change with an explicit pathspec (`make build-*` builds from
   committed HEAD, so the commit must precede any build).
2. Submit to the council gate before/alongside the commit; use `Council-Submitted:`
   on the commit if the verdict has not landed, never `Council-Reviewed:` on an
   unread verdict.
3. No image build or roll by this lane — releases are whole-fleet and the owner runs
   `make release`. The change rides the next one.
4. `bugs_open/246` stays in `bugs_open/` (owner ruling 2026-08-06) and is updated
   with status, evidence and the post-roll check.

## Rejected alternatives

- **Narrow `db` to a query-only interface** so the processor *structurally cannot*
  resize the pool. This is the strongest form of "make the bad state
  unrepresentable" and I am rejecting it deliberately: `p.db` is passed onward to
  `orchestration.NewStateRepository(db *sql.DB)` and `NewSagaCoordinator(db *sql.DB)`,
  both **public constructors used across the platform**, so the narrowing ripples
  well beyond this bug — a large refactor of shared signatures to fix a
  three-line defect with exactly one caller. The behavioural test buys most of the
  protection (a re-introduction fails loudly) at a tiny fraction of the risk.
  Worth revisiting if a second caller ever appears.
- **Raise-only (`if current < n`)**, per the bug file's candidate 2. Rejected for the
  bug file's own reason: it leaves two owners of one setting, which is the shape that
  produced the defect.
- **Document it next to the env var** (candidate 3). Rejected: a comment is not a
  control, and this estate has a standing ruling to that effect.
- **Move all pool sizing into a shared helper.** Attractive symmetry, but it would
  add a shared mechanism to fix a local defect, and there is only one opener that
  matters. One caller does not justify a new seam.
