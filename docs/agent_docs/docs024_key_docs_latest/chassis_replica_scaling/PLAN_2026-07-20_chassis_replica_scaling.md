# PROBLEM STATEMENT — the chassis cannot safely run more than one replica

**Status: PROBLEM STATEMENT, not yet a plan.** Written 2026-07-20 (bugfix 003
thread) at the owner's request: state the problem clearly first, work it up to a
plan afterwards. §5 sketches candidate directions deliberately as *options with
open questions*, not as a chosen design. Nothing here has been implemented.

> **UPDATED 2026-07-20, later (fixing-throughput thread): §§1–7 stand as the
> problem statement; §§8–13 below work it up into the plan.** The owner's steer
> arrived the same day — the target is **thousands of domains** — so this is a
> throughput problem as much as a safety one, and §7's budget questions are
> partly answered. Where the plan contradicts §5's sketches it says so
> explicitly (notably: option A alone is UNSAFE — §8.1). Still nothing
> implemented.

Related, do not fork: `bugs_open/003` (lost child responses — §4.4d is the
recommendation this problem blocks), `bugs_open/030` (dispatch queue latency),
`ANALYSIS_chassis_response_consumer_group_race.md` (the 2026-05-10 discovery,
still unremediated), `001_development_guide(5).md:466–505` (topic architecture).

---

## 1. The problem in one paragraph

The `agent-chassis` Deployment — the pod that consumes the platform's generic
entry point and runs every orchestration that is not a spawned Job — runs at
**`replicas: 1`**, and cannot currently be raised. Raising it activates a
documented, unfixed defect: chassis pods each join a **per-pod consumer group**
on the shared responses topic, so every response is delivered to *every* pod and
processed more than once. Meanwhile staying at one replica has its own standing
cost: every deploy leaves a window with **no consumer at all**, and because the
consume loop is at-most-once (`bugs_open/003` §3d), messages in flight during
that window are destroyed rather than delayed. So the system is pinned between
two failure modes, and the pin is invisible — nothing alerts on either.

## 2. Why replicas cannot simply be raised (the blocker)

**Evidence, code:** `platform/agentbase/agent.go:376–381` builds the response
consumer with `a.AgentID` as its group ID:

```go
responseConsumer, err := kafka.NewConsumer(
    a.config.KafkaBrokers,
    responsesTopic,
    a.AgentID,      // <-- per-pod UUID, NOT a shared group name
    a.logger,
)
```

The request consumer, two dozen lines earlier, correctly uses the shared
`KAFKA_CONSUMER_GROUP` (`generic-requests-group`). The asymmetry is the bug: one
topic behaves as a work pool, the other as a broadcast.

**Evidence, observed:** `ANALYSIS_chassis_response_consumer_group_race.md`
recorded two chassis pods running `ProcessResponse` on the same spawn-response
**215 ms apart**, the loser failing its optimistic-concurrency check. Its own
account of why it surfaced: *"the generic chassis Deployment has 3 replicas,
which is why we hit it here."* It is explicitly marked *"Discovery, not yet
remediated"* and *"Nothing in this document changes anything."*

**Evidence, live 2026-07-20:** 101 consumer groups on the cluster, **76 of them
per-pod UUID-shaped** — the sprawl that doc described has grown, not decayed.
Each abandoned group is dead weight in `__consumer_offsets`.

> **A correction to how this is usually stated.** "Only at 3 replicas" is too
> comfortable. The Deployment's strategy is `RollingUpdate` with
> `maxSurge: 25%` / `maxUnavailable: 25%`; at `replicas: 1` that rounds to
> **surge first, then terminate** — so *every rolling update briefly runs two
> chassis pods*, and during that overlap the duplicate-response path is live.
> The race is not dormant today; it is merely rare and unlogged. **[INFERRED
> from the strategy arithmetic — not yet confirmed by catching a duplicate
> `ProcessResponse` in a rollout window. Cheap to check: two pods, same
> response, timestamps within a second.]**

## 3. Why staying at one replica is not a resting state either

- **Every deploy has a no-consumer gap.** One pod, one consumer. Until the
  replacement joins the group and is assigned the partition, nothing is reading
  the entry point.
- **The gap destroys work rather than delaying it**, because `Consume()` commits
  the offset before the handler runs (`bugs_open/003` §3d). This is the
  mechanism behind CLAUDE.md's folklore rule that a dispatch within ~300s of a
  chassis restart is silently dropped.
- **No failover.** Node drain, OOM kill or eviction stops the entire generic
  entry point until the pod is rescheduled. Scheduled tasks keep firing into a
  topic nobody is reading.
- **It masks the defect in §2**, which is why the constraint has never been
  forced into the open: single-replica *looks* stable precisely because it
  suppresses the symptom.

## 4. The trap: more replicas would not buy throughput anyway

Both generic topics are **`PartitionCount: 1`** (`ReplicationFactor: 3`,
verified 2026-07-20). Kafka assigns a partition to at most one consumer per
group, so within `generic-requests-group` a second chassis pod is an **idle
standby** — it buys failover and rollout continuity, and **zero** additional
throughput.

This matters because it is easy to reach for replicas as the fix for
`bugs_open/030` (every dispatch queueing behind every other, ~25–36 min
latency). It would not work. **030 is a partition-count problem wearing a
replica-count costume**, and raising replicas without repartitioning would add
the race in §2 while changing throughput not at all.

## 5. Candidate directions (NOT a chosen plan — open questions attached)

**A. Give the responses consumer a stable shared group name.** The fix
`ANALYSIS` proposes and nobody applied. Cheapest, and unblocks everything else.
- *Open:* what should the group name be keyed on — agent **type**, so all
  chassis pods share one and spawned Jobs keep their own isolation? A spawned
  Job is a single pod, so a per-pod group is harmless there; the defect is
  specific to multi-replica Deployments. Any change must not collapse Jobs into
  a shared group by accident.
- *Open:* does a response need to land on the pod that sent the request? Belief
  is **no** — the coordinator reloads state from the DB, so any pod can process
  any response — but the in-memory timeout goroutine (`bugs_open/003`'s third
  root cause) does live in the sending pod. **[UNVERIFIED — this belief must be
  checked before it is designed on.]**
  > **ANSWERED 2026-07-20 (fixing-throughput thread), and the answer changes
  > option A's shape: the belief was half right.** The DB layer is fully
  > pod-agnostic, but `ProcessResponse` then *deliberately refuses* responses
  > for orchestrations stamped by another pod — after claiming them. See §8.1:
  > A is only safe shipped together with the removal of that check.
- *Open:* migrating group names resets offsets. With `StartOffset: FirstOffset`
  a fresh group **replays the topic from the beginning** — observed on 2026-07-20
  when a test pod replayed 11 days of messages. A one-off offset seed is needed,
  or the cutover replays history.

**B. Repartition the entry-point topics** (for `bugs_open/030`, not for this).
- *Open:* partition key. `orchestration_id` preserves per-orchestration ordering
  while allowing parallelism — but does anything rely on **global** ordering
  across orchestrations today? Unknown.
- *Open:* Kafka cannot reduce partitions, and repartitioning an existing topic
  redistributes keys. Sequencing with A matters.

**C. Make a rollout survivable** — already scoped as `bugs_open/003` F3/F4:
at-least-once consume with completion-time dedupe, honest readiness (shipped),
`preStop`, grace > drain. **Independent of A and B, and worth doing regardless**
— it is what turns the deploy gap from destructive into merely slow.

**D. Static group membership** (`group.instance.id`) to damp rebalance churn once
more than one consumer exists — the architecture doc lists it under the future
shared-topic strategy. Probably later, not now.

**Provisional sequencing, for argument:** C is independent and already in
flight. A must precede any replica increase. B is a separate problem that should
not be conflated, but shares a cutover window with A.

## 6. What would tell us this is fixed

- Two or more chassis pods running, and a given response processed **exactly
  once** across them (log a response id per pod and diff).
- A rolling update with **no** dispatch loss and no duplicate-response
  transitions — the deliberate test is a rollout mid-orchestration.
- The per-pod group count stops growing.
- `bugs_open/003` §4.4d can be applied without the warning that currently sits
  on it.

## 7. Open questions for the owner

1. **Is multi-replica actually wanted, or is the goal just a safe deploy?**
   C alone delivers safe deploys at `replicas: 1`. A + raised replicas is only
   needed for failover and node-drain resilience. These are different budgets.
2. **Is the ~25–36 min dispatch latency (`bugs_open/030`) a problem worth
   repartitioning for**, or is it tolerable and merely surprising? That decides
   whether B is in scope at all.
3. **How much replay is acceptable at cutover** (option A's offset question) —
   seed offsets to "now" and accept a small blind window, or accept a replay?

---

# THE PLAN (added 2026-07-20, fixing-throughput thread)

## 8. Evidence gathered while working this up

### 8.1 §5A's open question is answered — and option A alone is UNSAFE

Any pod *can* process any response at the DB layer: `ProcessResponse` loads
state fresh from Postgres, and duplicate suppression is an atomic row claim
(`ClaimAwaitedRequest`, `platform/orchestration/coordinator.go:303–390`), not
pod memory. But immediately **after** a successful claim, `coordinator.go:271–277`
compares `state.ProcessingNode` with the pod's own name and returns without
processing on mismatch ("Response for orchestration owned by different pod,
ignoring"). `ProcessingNode` is stamped with the executing pod's `HOSTNAME` by
`SetExecutingStep` (`platform/orchestration/state.go:1109`; callers
`coordinator.go:698` and `:820`) and is never cleared or re-stamped on
takeover. The claim is **not released** on this path; the only reset
(`processResponseClaimWithRetry`'s CLAIM_RECOVERY) fires only when a *second*
response arrives for the same request_id — which a child never sends.

Consequences:

- **Shared group + N replicas without deleting this check = systematic response
  destruction.** ~(N−1)/N of responses land on a non-stamping pod, get claimed,
  get dropped, and wedge their `awaited_requests` row (`claimed`,
  `processed_at NULL`) until the reaper fails the orchestration. **Option A
  must ship in the same change as the removal of this check.** The check's
  protective purpose — duplicate suppression — is already served, properly and
  atomically, by the claim itself.
- At `replicas: 1` **today**, the same path is armed across every rollout: the
  new pod replays the responses topic under its fresh per-pod UUID group
  (`StartOffset: FirstOffset` — §5A's 11-day replay), and any response for an
  orchestration stamped by the *old* pod hits the mismatch. Exposure per roll =
  orchestrations in `AWAITING_RESPONSES` at roll time.
- Live snapshot 2026-07-20 ~21:30Z: 0 "owned by different pod" lines in the
  current pod's log (pod 115 min old), and only **one** orchestration in flight
  (`EXECUTING_STEP`, stamped by the current pod) — so no live incident evidence
  in this window, but the at-risk population was also ~zero, which makes the 0
  uninformative (RUNBOOK R2). **[Mechanism: code-verified this session.
  Live-loss claim: FILED to the diagnosis loop rather than asserted — corr
  `2d02d62a-7d96-41f0-a82b-e1ebd7ef5d6b`, verdict pending. Read it before
  building P2.]**

### 8.2 The "response must home to its sender" residue is exactly F2's job

The one thing that genuinely lives in the sending pod is the timeout timer
(`bugs_open/003` third root cause: `go s.handleRequestTimeout(...)` sleeping
until `timeout_at`). `bugs_open/003` **F2** — DB-driven retry of expired
`awaited_requests`, driven from the existing per-pod ticker with an atomic
claim — removes the last reason a response would need to land anywhere in
particular. **F2 is therefore a prerequisite of multi-replica**, not just a
spawn-loss fix.

### 8.3 Throughput arithmetic (measured, not estimated)

- Sustained drain of `system.agent.generic.requests`: **~2.4 msg/min**
  (`bugs_open/030`, corrected 22-min continuous sample), set by mean inline
  segment duration — the consume loop is synchronous (`processRequests` →
  `processMessage` inline) and `continueExecution` advances consecutive steps
  in an unbounded for-loop spawning no goroutines. **[FILED as corr
  `78470372-7617-40e4-888c-66cac94006bf`; still `awaiting_diagnosis` at
  21:30Z — sitting in the very queue it describes.]**
- Volume this week: **1,918–3,872 orchestrations/day** (`orchestration_states`
  7-day group-by, 2026-07-20), at **11 deployed sites** (+17 inert pool, 1
  system) plus heavy concurrent-session traffic.
- So the current ceiling is ~0.04 dispatch/s, and the topic already queues
  40–90 deep routinely.

### 8.4 `orchestration_requests` is a dead table

Intake-shaped schema (request_id, orchestration_id, status, timeout_at, FK to
states, useful indexes) and **zero Go references** (tree-wide grep,
2026-07-20). Recorded so nobody assumes it is live. P1 may adopt it or ignore
it; adopting a dead table into a new contract needs a migration either way.

## 9. Sizing for thousands of domains

Site-driven work (feeds, builds, discovery checks, imagery, claims, rerenders)
scales ~linearly with domains; session-driven work does not. Scaling 11 →
1,000–5,000 domains multiplies site-driven volume ~100–500×: order
**10⁵–10⁶ orchestrations/day ≈ 2–10 sustained dispatches/s**, bursty. Gap to
the measured ceiling: **~2.5 orders of magnitude.**

Closing that gap with Kafka-native parallelism alone means
`partitions × (1 / mean-segment-duration)`: at ~25 s mean segments, 5/s needs
**≥125 partitions and 125 blocking consumers** — pre-provisioned (partition
count is a one-way door), rebalance-stalled at every roll, still
head-of-line-blocked *within* each partition (a council step still delays a
page render sharing its partition), and still SQL-invisible for diagnosis.
That is the wrong tool shape for this workload. **[Target rates are ARITHMETIC
on measured figures, but the scaling assumption is linear — owner input on
per-domain cadence wanted, §13. It sets P3's envelope, nothing earlier.]**

## 10. The design: Kafka delivers, Postgres decides

Every correctness mechanism that actually works in this platform is already in
Postgres: atomic response claims, `site_work_items` claim/lease,
processed-message dedupe, optimistic state locking, the reaper, F2's timer
rows. Kafka currently provides transport *and scheduling*; scheduling is what
it does badly for long, variable-duration, per-orchestration-ordered work —
and every fix this platform has shipped for a Kafka-semantics defect has moved
truth into the DB. The long-term design completes that convergence: **move
scheduling into Postgres; leave transport in Kafka.**

1. **Thin ingest.** The chassis request consumer does, per message: validate →
   persist an intake event row (idempotent on request_id/message_id — F3's
   dedupe contract) → commit offset. Milliseconds per message, so one partition
   carries hundreds/s and **`PartitionCount` ceases to be a throughput
   parameter at all**. The intake row is 030's missing acknowledgement: a
   SQL-visible `received_at` the moment the message is consumed. The entire
   "was my dispatch dropped?" failure class — two WRONG_CALLS rows and
   counting — dies.
2. **Claim-workers.** A worker pool inside the same chassis binary claims
   runnable work with `SELECT … FOR UPDATE SKIP LOCKED` + lease/heartbeat (the
   `site_work_items` shape, reused not reinvented): new intake events, response
   events, expired `awaited_requests` (F2's driver, generalised). **Per-
   orchestration serialization is structural** — a worker claims the
   orchestration, drains its pending events in order, releases. Cross-
   orchestration head-of-line blocking is gone: a 5-minute council run occupies
   one worker, not the lane. Concurrency = replicas × workers-per-pod,
   adjustable live; poison work poisons one claim, not a partition.
3. **Responses join the same path.** One stable shared group
   (`generic-responses-group` — keyed by *role*, not pod). The consumer
   persists response events; workers run them through the existing
   `ProcessResponse` **minus the deleted ownership check** (§8.1). Spawned Jobs
   keep their per-job topics and get a group name derived from the stable job
   identity instead of the pod UUID, so a Job restart stops replaying its
   topic (same `FirstOffset` trap, observed 2026-07-20).
4. **Timers**: in-memory goroutines stay as the fast path; DB expiry (F2) is
   the guarantee; the reaper (F1, live) the backstop.
5. **Deploys become boring**: thin consumers drain in seconds (F4's
   preStop/grace); leases expire; surviving pods resume the claims;
   at-least-once + dedupe means the window loses nothing. `replicas: 2+`
   becomes safe AND meaningful, and the ~300 s post-restart folklore rule
   retires.

**What does not change:** every producer contract (trigger scripts,
kafka-scheduler, the spawner), the coordinator state machine and actions,
per-spawn topics, adapters, the generic topics themselves. The "current door"
stays the door; what changes is that nothing substantial happens behind it on
the consumer goroutine.

**Why not Kafka-native scaling** (the rejected alternative): §9's partition
arithmetic; competing consumers with out-of-order completion need manual
offset management (commit only up to contiguous-done) which kafka-go makes
genuinely hard; hot keys re-serialize; rebalances stall the whole group at
every roll; and once F2+F3 land, Kafka would provide nothing Postgres isn't
already obligated to provide — except the bottleneck and the blindness.

**Postgres headroom:** target event rates (≤10–20 writes/s for intake+claims)
are orders below where SKIP-LOCKED queues degrade. The real next constraint is
`orchestration_states` write amplification (full-JSONB row rewrite per step,
large `collected_data`, TOAST churn) — a **P4 watch item with a measurement
gate**, not built now. Connection budget: workers × replicas through pooling.
**[UNMEASURED: states row-size distribution — measure before P3 tuning.]**

## 11. Phasing (each phase shippable and valuable alone)

**P0 — prerequisites (owned by the bugfix-003 thread; mostly in flight).**
F1 ✓ live. F2+F3 ship together (migration 180 + image roll). F4 rollout
survivability. **Also: read the two diagnosis verdicts (corr `78470372`,
`2d02d62a`) before writing P1/P2 code** — this plan is deliberately filed
ahead of its two structural premises, and a REFUTED verdict revises §10, not
the other way round.

**P1 — decouple consumption from execution** (fixes 030's latency AND buys the
throughput headroom; no topology change, no group rename, replicas stay 1).
Intake persist + commit in the request consumer; the worker pool executes.
Verify: a trigger fired under load has a SQL row within seconds; dispatch
latency drops from ~25–36 min to seconds; drain scales with worker count
(measure via the dispatch-queue workstream's runbook, not two-point samples).

**P2 — responses + multi-replica** (fixes this file's headline problem).
Shared response group with offset seeding at cutover (seed to latest; sequence
behind F3 so the window loses nothing); **delete the §8.1 ownership check in
the same commit**; response events through workers; `replicas: 2–3`.
Verify: §6's tests — a response processed exactly once across pods (per-pod
response-id logs, diffed); a rolling update mid-orchestration completes; the
per-pod group count stops growing.

**P3 — scale-out + hygiene.** HPA on a queue-depth SQL metric; tune
workers-per-pod; sweep the 76 dead UUID groups; stable per-job group names;
revisit partition count ONLY for ingest redundancy (keyed by
`orchestration_id` if ever raised — it is no longer a throughput decision).

**P4 — watch items, gated on measurement.** States write amplification;
Postgres capacity; terminal-state archival. Collect the numbers first
(row-size distribution, writes/step, TOAST churn at P1 load); design only if
the gate trips.

## 12. What would tell us the plan is working (extends §6)

- P1: publish→row latency in seconds under concurrent load; topic `LAG` near
  zero in steady state; a council run in flight no longer delays the page
  build behind it.
- P2: §6's exactly-once and rollout tests pass; a week with zero
  `reaper:`-failed orchestrations attributable to deploy windows.
- Both filed diagnosis verdicts read and reconciled — CONFIRMED hardens §10;
  REFUTED revises it (a REFUTED verdict is a success; it is the cheapest place
  to be wrong).

## 13. Questions for the owner (supersedes §7)

§7 Q1 is answered by the steer that prompted this plan (thousands of domains →
both safety and throughput, phased). §7 Q2 dissolves — repartitioning is no
longer the mechanism. §7 Q3 narrows to the P2 cutover: recommend seed-to-now,
accepting a blind window of seconds, sequenced behind F3's at-least-once.

1. **Per-domain cadence at target scale.** §9 assumes linear scaling from
   today's 11 sites. What does one domain generate per day at 1,000+ (feed
   polls hourly? builds weekly?)? Sets P3's HPA envelope only — P1/P2 are
   unconditional.
2. **Sequencing consent:** P1 before P2? (Recommended: yes — P1 removes the
   daily 25–36 min pain with no topology risk, and proves the worker machinery
   at `replicas: 1` before P2 depends on it.)
3. **Archival policy** for terminal orchestrations (P4): how long must
   completed states stay hot?
