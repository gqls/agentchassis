# PROBLEM STATEMENT — the chassis cannot safely run more than one replica

**Status: PROBLEM STATEMENT, not yet a plan.** Written 2026-07-20 (bugfix 003
thread) at the owner's request: state the problem clearly first, work it up to a
plan afterwards. §5 sketches candidate directions deliberately as *options with
open questions*, not as a chosen design. Nothing here has been implemented.

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
