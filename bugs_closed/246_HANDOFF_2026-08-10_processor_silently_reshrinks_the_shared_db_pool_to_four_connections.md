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

### Post-roll verification

> **CORRECTED 2026-08-11, same day, twice over.** What stood here was:
> *"Confirm the commit shipped: `kubectl … logs -l app=agent-chassis | grep -m1 'build
> provenance'`, then `git merge-base --is-ancestor <sha> <stamp>`."* **That recipe is
> documented as INOPERATIVE on this exact service** and I should not have written it.
> Caught independently by two reviewers within an hour: the council's `debug_historian`
> seat (objection, medium — *"substitutes a verification mechanism documented as broken
> on this exact service"*) and the `bugfix_239` lane by cross-session message. Both cited
> the same landmine, which I had not grepped for `agent-chassis` before writing a
> verification step for `agent-chassis`.

**Why the obvious recipe fails here, and why the obvious fallback is worse:**

- `build provenance` is logged **once, at startup**. The chassis is the noisiest service
  in the fleet and its container log rotates on size, so the line is out of range within
  the hour. An empty grep means *"rotated"*, not *"unstamped"*. A naive grep over full
  logs can also **false-match a council-gate payload that quotes the phrase** (239 lane).
- The tempting fallback — probe `/proc/1/exe` for **your own** commit sha — is **the wrong
  test**, not merely a weak one. The binary is stamped with **ONE** commit (the build
  point), not with every ancestor, so a binary that genuinely contains this change reports
  your sha as absent. Three absents in a row read as "my fix did not ship".

**So, in order:**

1. **Get the stamp from the IMAGE, not the pod**, then test ANCESTRY:
   ```bash
   docker image inspect docker.io/aqls/agent-chassis:$IMAGE_TAG \
     --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'
   git merge-base --is-ancestor 039cfce84 <that-revision>   # exit 0 = shipped
   ```
   **Per service, not per fleet** (`bugs_open/249`: one tag shipped three revisions).
2. **Accept that there is NO behavioural witness for this change, and do not invent one.**
   The landmine's standing advice for the chassis is "prefer behaviour over provenance —
   fire the smallest input only the new code can answer". **That advice cannot be followed
   here**, because nothing in the platform reports a pool's size or its wait counters. This
   is not a gap in the verification plan; it is the same gap the change itself is about, and
   it is the argument for the deferred `db.Stats()` follow-up below.
3. The disconfirming observation for the pgbouncer risk: `SHOW POOLS` with `cl_waiting`
   sustained > 0, or `maxwait` climbing.
4. **Do not claim to have "observed the pool at 12".** Nothing reports a pool's size. The
   mutation-proven unit test is the evidence for the mechanism; image-label ancestry is the
   evidence that it shipped. Those are the only two available, and neither is a live reading.

**Note on which roll:** `v1.0.1286` (stamp `c3b424c8e`) rolled at 12:03 UTC on 2026-08-11 and
does **not** carry this fix — `039cfce84` is later. 246 awaits the NEXT roll. (Contributed by
the `bugfix_239` lane, whose own fix that roll did carry.)

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

---

## 2026-08-11 evening — LIVE on `v1.0.1288`. Fix shipped and proven by ancestry.

**Status: FIXED AND LIVE.** Kept in `bugs_open/` per the owner ruling of 2026-08-06.

### Proof it shipped (the corrected method, run as written above)

```
image label (docker image inspect …:v1.0.1288 org.opencontainers.image.revision)
  = bb534864249117003ac758e50adc0df9176ef370

git merge-base --is-ancestor 039cfce84 bb5348642   -> exit 0   FIX SHIPPED
git merge-base --is-ancestor 6ba3fca28 bb5348642   -> exit 0   council follow-up shipped
git merge-base --is-ancestor 2194df2cf bb5348642   -> exit 1   NEGATIVE CONTROL holds
```

Pods `agent-chassis-596d84f6b-{kmc2t,tb8gd}`, image `v1.0.1288`, started 17:13/17:14Z,
**0 restarts** — so the constructor change did not break startup.

> **The landmine was confirmed live, by accident, in the same breath.** The pods started
> at 17:13:36Z; their logs' FIRST AVAILABLE LINE is already **17:43:51Z**. Thirty minutes
> of startup log had rotated away within half an hour, so `grep 'build provenance'`
> returned nothing on a pod that is unambiguously stamped. **An empty grep here means
> "rotated", never "unstamped"** — exactly as the landmine says, now with a second
> measurement behind it.

> **MISSTEP, recorded because the check nearly passed for the wrong reason.** My first
> negative control was my own summary commit `1a72d08f9`, which I assumed post-dated the
> build. It did not — the stamp `bb5348642` is 17:52:33, later than my commit — so the
> control reported "is an ancestor" and looked like a broken test. The control was wrong,
> not the test. **A negative control must be chosen by ASKING the repo which commits are
> outside the stamp, never by assuming your own work is the newest thing on a shared
> branch:** `git rev-list <stamp>..HEAD | head -1`. On this tree, "my commit is the latest"
> is false within minutes.

### Post-roll measurements — one useful, one still MISSING

**pgbouncer server-side pool (Postgres view), now vs earlier today:**

| | pre-roll (~13:30Z) | post-roll (~18:10Z) |
|---|---|---|
| server conns to `clients_db` | 6 | **10** (of `default_pool_size = 15`) |
| active / idle | — | 2 / 8 |
| chassis intake events/hour | 60–100 | **182–228** |

**Do NOT read the 6 → 10 rise as the effect of this fix.** It is **confounded**: the
client-side cap went 4 → 12 *and* load roughly tripled in the same window. Two variables,
one observation. What can be said: 10 of 15 leaves less headroom than before, nothing is
waiting on a lock, and this is the number to watch.

**THE DISCONFIRMING OBSERVATION IS STILL UNMEASURED — and it is blocked on a credential.**
The real saturation signal is pgbouncer's own `SHOW POOLS` (`cl_waiting` sustained > 0, or
`maxwait` climbing). `pg_stat_activity` **cannot** show it: every row's `client_addr` is
pgbouncer itself, so the Postgres side sees the server pool and never the client queue.
The admin console requires `pgbouncer_admin` (`admin_users`/`stats_users`,
`pgbouncer-configmap.yaml:73-74`); that user exists in
`/etc/pgbouncer/userlist.txt` but **its password is not in `personae-platform-secrets`**,
so no session can currently run the one query that would settle this risk.
**This is an owner decision, not a task** — see the lane's handoff.

### What is still true, and is not going to be settled by watching

There is **no behavioural witness for this change** and there never was going to be:
nothing in the platform reports a pool's size, and nothing surfaces
`db.Stats().WaitCount`/`WaitDuration`. The evidence for the mechanism is the
mutation-proven unit test; the evidence that it shipped is the ancestry check above.
Those are the only two, and neither is a live reading of the pool. **Anyone who later
writes "the pool was observed at 12" is describing something the platform cannot show.**

---

## CLOSED 2026-08-12 — fixed, council-approved, live and post-roll verified

Moved to `bugs_closed/` under the owner direction of **2026-08-12**: *"if it is fixed and
live it should be moved"* (commit `2aa3014a3`), which **supersedes the 2026-08-06
stay-in-`bugs_open` direction** this file was previously kept under.

**Closure evidence, all in this file above:** the mechanism proven in isolation with a
control; the fix and its council follow-up both council-APPROVED at round 1
(`c94d73ac-2a15-40cb-98a9-1185a2b7435a`); a mutation-proven regression test; and shipment
proven by image-label ancestry on `v1.0.1288`, re-verified still shipped on `v1.0.1290`
(revision `fa078ab3d`) with a valid negative control.

**What is deliberately NOT claimed by this closure**, so nobody inherits a false sense of
completeness:

- **The pgbouncer risk remains UNMEASURED.** `SHOW POOLS` (`cl_waiting`, `maxwait`) needs
  the `pgbouncer_admin` console user; the Terraform half was wired 2026-08-12
  (`aee444a35`) but the `pgbouncer-userlist` half is not Terraform-managed and still needs
  the credential holder. Runbook §9.
- **There is no behavioural witness for this fix and there never could be** — nothing in
  the platform reports a pool's size or its wait counters. Anyone later writing "the pool
  was observed at 12" is describing something the platform cannot show.
- **The follow-ups are live work, not loose ends**, and they are owned elsewhere now:
  `db.Stats()` instrumentation (D2) is with the `bugfix_239` lane; the `p.sqlDB` collapse
  (D5) turned out to be a defect and is **`bugs_open/259`** (the
  `three_processor_paths...` one — the number is ambiguous), also with that lane.

Remaining owner decisions live in
`docs/agent_docs/docs024_key_docs_latest/bugfix_246_shared_pool_ownership/HANDOFF_2026-08-11_continue_here.md`
— closing the bug does not close those.

---

## 2026-08-14 — THE RISK IS MEASURED AT LAST. It is not materialising.

This file closed on 2026-08-12 saying plainly that its one real risk — whether the chassis,
now holding its configured 12 client connections instead of a silent 4, queues at pgbouncer
— was **UNMEASURED**, and that the instrument (`SHOW POOLS`) was unreachable for want of a
credential. The credential is now recorded (owner decision D1, closed 2026-08-14) and the
measurement has been taken:

| database | user | cl_active | **cl_waiting** | sv_active | sv_idle | **maxwait** | pool_mode |
|---|---|---|---|---|---|---|---|
| clients_db | clients_user | 17 | **0** | 3 | 2 | **0** | transaction |
| pgbouncer | pgbouncer | 1 | 0 | 0 | 0 | 0 | statement |

**No client is queued for a server connection, and none has waited.** 17 client connections
are multiplexed onto **5 server connections of `default_pool_size = 15`** — which is
transaction pooling behaving exactly as the submission argued it would when it claimed the
change "does not simply move the queue". That argument is now evidenced rather than
reasoned.

**Consequences:**
- The 2026-08-12 closure's outstanding caveat is **discharged**.
- **Owner decision D3 is answered: `default_pool_size = 15` does not need raising.**
- The pgbouncer configmap's stale rationale (*"3 chassis replicas × 4 conns"*) can be
  corrected to reflect 2 replicas × 12 whenever someone is in that file; it is now
  demonstrably not a capacity problem.

> **What this reading does NOT establish, stated so it is not over-quoted later.** It is
> **one sample**, at ~17 client connections. `cl_waiting` is instantaneous, and pgbouncer's
> `maxwait` is **the current longest wait, not a high-water mark** — it returns to 0 as soon
> as the waiting client is served. So a zero cannot rule out queueing *between* samples, and
> a burst is precisely when the answer could differ. To turn this into a result rather than
> a snapshot, sample across a busy period (`SHOW POOLS` on a loop, or `SHOW STATS`).
> **The honest sentence is "pgbouncer was not queueing at this moment, at this load" — not
> "pgbouncer is fine under load".**

> **And the other limit still stands, permanently:** there is still **no behavioural witness
> for the fix itself**. Nothing reports a pool's size or its wait counters, so this measures
> the *consequence* at pgbouncer, not the chassis-side pool. That gap is what the `db.Stats()`
> follow-up (D2, owned by the `bugfix_239` lane) exists to close.
