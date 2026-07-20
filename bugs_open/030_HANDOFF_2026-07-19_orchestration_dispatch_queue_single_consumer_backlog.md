# 030 — Every orchestration dispatch queues behind every other: one partition, one consumer, ~25–36 min latency

**Filed:** 2026-07-19 · **Branch:** `085_debug_and_feature_loops` · **Status:** OPEN, not started
**Severity:** medium-high — not a data-corruption bug; it is a **latency and diagnosability** bug that
wastes sessions' time, has already caused at least two threads to misdiagnose a delay as a failure,
and gets worse the more sessions work concurrently.
**Class:** platform infrastructure (Kafka topic + consumer topology), not a per-site defect.

> **Read this first if you are about to conclude "my dispatch was dropped".** It probably was not.
> Check the consumer-group lag before concluding anything (§ "The one command" below).

---

## Symptom

You fire a trigger (`090_TRIGGER_needs_diagnosis`, `097_TRIGGER_council_review`, a discovery run,
any `action=orchestrate` publish). The script prints an orchestration id and exits cleanly. Then:

- `SELECT … FROM orchestration_states WHERE orchestration_id=…` → **0 rows**
- `SELECT … FROM orchestration_requests WHERE orchestration_id=…` → **0 rows**
- the chassis log has **nothing** for your correlation id
- meanwhile **other** council/diagnosis runs visibly complete and write their verdict notes

Everything about that picture says "my message was dropped". It was not. It is sitting in an
in-order queue behind other threads' messages, and it will run — in **25 to 36 minutes**.

## Measured, not estimated (2026-07-19)

**Two end-to-end measurements from this session**, publish timestamp → `orchestration_states` row:

| submission | published (UTC) | orchestration created (UTC) | latency |
|---|---|---|---|
| council-gate-132453 | 12:24:53 | 13:01:16 | **36 min 23 s** |
| council-gate-134936 | 12:49:36 | 13:14:37 | **25 min 01 s** |

**Topology** — this is the root of it:

```
$ kafka-topics.sh --describe --topic system.agent.generic.requests
Topic: system.agent.generic.requests  PartitionCount: 1  ReplicationFactor: 3
                                      Configs: min.insync.replicas=2
```

```
$ kafka-consumer-groups.sh --describe --group generic-requests-group
GROUP                  TOPIC                          PARTITION CURRENT-OFFSET LOG-END-OFFSET LAG CONSUMER-ID
generic-requests-group system.agent.generic.requests  0         93402          93443          41  agent-chassis-5c568b8c74-2f4qv
                                                                                                  (github.com/segmentio/kafka-go)
```

**One partition. One consumer instance. Strict in-order, serial processing.** Every trigger from
every concurrent session — and there are many — funnels through that single lane.

Observed lag over ~35 minutes of sampling: **41 → 62 → 24**. So the consumer *does* drain; the queue
is bursty rather than permanently diverging. But head-of-queue age was measured at **25.9 minutes**:

```
head-of-queue message timestamp (offset 93403): 12:24:43 UTC
wall clock at time of check:                    12:50:39 UTC
```

That 26-minute head age *is* the user-visible latency, and it is set by how many other sessions
fired recently — not by anything about your own request.

## Why this is worth fixing rather than documenting

The latency alone is tolerable. **The diagnosability is not.** The failure mode is
indistinguishable from a drop from every observable surface a session has:

1. Nothing anywhere records "your message was received and queued". There is no row, no log line,
   no acknowledgement between `kcat` exiting 0 and an orchestration appearing half an hour later.
2. The trigger scripts print "Submitted. Watch the run:" with a query that returns 0 rows for the
   next ~30 minutes, which reads as failure.
3. **It has already cost real work, at least twice:**
   - This session concluded the message was dropped, re-submitted, and paid for a duplicate run
     (both copies later died on `/bugs_open/019`, so the duplicate spent reviewer LLM calls for
     nothing). Recorded in `docs024_key_docs_latest/idea_uk_vm_site/RUNNING_NOTES §X.1`.
   - An earlier session in the same workstream recorded that on-demand discovery dispatches
     "produced no orchestration row at all", noted that the documented 300 s post-restart drop
     could not explain it (the pod was six hours old), and abandoned the investigation with a TODO.
     That was almost certainly this. See `idea_uk_vm_site/README_where_we_are.md` §S.
4. It actively **misleads against a real, documented rule.** `CLAUDE.md` warns that dispatches within
   ~300 s of a chassis restart are silently dropped. That rule is true, but it is not the common
   cause — queue depth is — and a session that knows the rule will check pod age, find it healthy,
   and be left with no explanation.

## The one command that settles it

Before concluding a dispatch was dropped, run this. It is decisive and takes seconds:

```bash
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group
```

`LAG` > 0 means queued, not lost. To see how stale the head of the queue is (i.e. your real wait),
read the timestamp of the message at `CURRENT-OFFSET`:

```bash
kubectl -n kafka run -i --rm kcat-head-$$ --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -C -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests -p 0 -o <CURRENT-OFFSET> -c 1 -e -f '%T\n'
```

To prove your own message reached the broker at all:

```bash
kcat -C … -o -60 -e -f '%T %h\n' | grep <your correlation_id>
```

**Gotchas, each of which cost time here:**
- The broker pod is `personae-kafka-cluster-combined-pool-prod-0`. `personae-kafka-cluster-dual-role-0`
  does not exist in this cluster and `kubectl exec` fails silently enough to look like an empty result.
- The Kafka CLI is at `/opt/kafka/bin/`, **not** on `$PATH`.
- `--describe --all-groups` iterates every group and takes **>120 s** — always name the one group.
- The trigger scripts name orchestrations `council-gate-$(date +%H%M%S)` in **local time (BST)**,
  while the DB is UTC. A run named `-132453` was created at `12:24:53` UTC. This will make you think
  a run is an hour old, or an hour in the future.

## Root cause

`system.agent.generic.requests` has **one partition**, so Kafka can only ever deliver it to one
consumer in the group, no matter how many chassis replicas run. Ordering is therefore total across
*all* work types and *all* sessions, and throughput is capped at one orchestration step-dispatch at
a time. Council and diagnosis runs are minutes-long multi-step orchestrations, so a handful of them
in flight is enough to put half an hour of head-of-line blocking in front of everyone else.

Nothing here is misconfigured *per se* — it is a topology that was fine for one operator and does
not survive many concurrent sessions, which is exactly the situation `CLAUDE.md` opens by describing.

## Fix candidates (in rough order of value/risk)

1. **Acknowledge the request on publish.** Cheapest, biggest diagnosability win, no topology change:
   have the trigger scripts (or a tiny consumer) write a row the moment a request is accepted, or
   simply have the scripts print the current `LAG` and estimated wait after publishing. A session
   that sees "queued behind 41 messages, ~25 min" does not misdiagnose a drop and does not
   re-submit. **This alone would have prevented both recorded incidents.**
2. **Partition the topic and scale the consumer group.** Raise `PartitionCount` above 1 and run
   matching chassis consumers. ⚠️ Check first whether anything depends on total ordering across the
   topic — partitioning by `orchestration_id` (as the key) preserves per-orchestration order while
   allowing parallelism, which is almost certainly the semantics wanted. Note Kafka **cannot reduce**
   partition count later, so this is one-way.
3. **Separate lanes for long-running work.** Council/diagnosis orchestrations are minutes long and
   are what create the head-of-line blocking. A dedicated topic for them would stop a council round
   delaying every page build in the fleet.
4. **Surface lag as a health signal** — if `LAG` on this topic is a first-class metric, "the cluster
   is busy" stops being something each session rediscovers by hand.

## How to verify a fix

- Re-run the two measurements above: publish → `orchestration_states` row should be seconds, not
  tens of minutes, while other sessions are active.
- `LAG` should stay near zero under normal concurrent load.
- The negative test that matters: fire a trigger while several other orchestrations are running and
  confirm it starts promptly rather than after them.

## Landmines

- **Do not "fix" this by making triggers retry.** Retrying a queued-but-not-yet-processed dispatch
  duplicates the work and spends credits twice — that is precisely the mistake this session made by
  hand, and automating it would make it systematic.
- **Do not raise partition count without deciding the key.** Unkeyed messages across multiple
  partitions lose ordering guarantees that the orchestration state machine may rely on.
- The 300 s post-restart drop documented in `CLAUDE.md` is a **separate, real** failure. Fixing this
  one does not retire that rule.
- **A FROZEN CONSUMER OFFSET IS NOT A DEAD CONSUMER — it is this bug, mid-message** (added
  2026-07-20, bugfix-028 thread, after I got it wrong). Chasing a stuck dispatch I sampled
  `CURRENT-OFFSET` three times over a minute and found it pinned at **95919** while
  `LOG-END-OFFSET` climbed (95970 → 95972) and `LAG` grew 51 → 61. I concluded the chassis had
  **stopped consuming fleet-wide** and was about to report a live outage. It had not: offsets are
  committed *after* a message is fully processed, and the message in flight was a multi-step
  council orchestration. The pod was healthy and logging LLM steps the whole time
  (`Rendered prompt template`, 18:33:52) — I had simply queried a log window that returned
  nothing and read the silence as death. **So: `LAG` is a trustworthy queue-depth signal; a
  static `CURRENT-OFFSET` is worthless as a liveness signal**, and will hold still for the entire
  duration of the longest orchestration on the topic. To check liveness, look at the chassis log
  for *any* recent line, or watch whether the offset eventually jumps by several messages at once
  — which is what draining looks like here, not a smooth advance.
- **Clearing bug 029's hung dispatch slots does NOT unblock this.** Same session, same hour: I
  found two `build-pipeline-trigger` orchestrations hung at `spawn_dispatch` and applied 029's
  documented recovery (2 cancelled, slots freed, group down to 1 of 8). The work item still did
  not move, because the bottleneck was **upstream** — the trigger cannot even be *delivered* while
  the single consumer is busy. 029 and 030 present identically (a healthy-looking `triaged` item
  that never dispatches) and the 029 recovery is cheap and harmless, so it is a reasonable first
  move — but if the item still does not budge, check `LAG` before concluding the recovery failed.

## Related

- `/bugs_open/019` — one truncated reviewer voids a whole council round. **Not the same bug, but it
  compounds this one**: both of this session's submissions waited ~30 min in this queue and then
  died on 019, so the round cost two waits and two sets of reviewer LLM calls for zero verdicts.
- `/bugs_open/006` — idea.uk infra errors, including claim-timeout churn (a different dispatch-layer
  problem; do not conflate).
- `docs024_key_docs_latest/idea_uk_vm_site/RUNNING_NOTES_idea_uk_vm_site.md` §X.1 — the full
  misdiagnosis, including the reasoning error ("first in a `kcat -o -60` window" means *oldest
  unprocessed*, not *skipped*) and the transferable rule.

## Measured throughput, 2026-07-20 19:16–19:31 UTC (bugfix-036 thread)

A hard number for the drain rate, since the file so far describes the shape but not
the speed. Two `kafka-consumer-groups.sh --describe --group generic-requests-group`
readings 14 minutes apart, on `system.agent.generic.requests`:

```
19:16   current=96013  end=96099  lag=86
19:30   current=96016  end=96102  lag=86
```

**3 messages consumed in 14 minutes ≈ 0.21 msg/min ≈ one message every 4.7
minutes.** Production was adding messages at the same rate, so `LAG` sat flat at
86 rather than falling — a steady state, not a drain. At that rate a message
arriving at the back of the queue waits **~6.5 hours**.

Two things this pins down that the entry above leaves open:

- **The rate is set by orchestration DURATION, not by queue mechanics.** ~4.7 min
  per message is about what one council/diagnosis orchestration takes end to end,
  which is consistent with the consumer running each orchestration to its wait
  point before taking the next message. So throughput on this topic is
  `1 / (mean orchestration duration)` — adding work of a *slower kind* lowers the
  ceiling for everything else on the topic, including fast work.
- **It is not the hung-spawn class.** At the time of these readings only **6**
  orchestrations were in flight (4 `AWAITING_RESPONSES`, 2 `EXECUTING_STEP`), and
  the single >4h straggler was reaped during the window. So the queue was 86 deep
  with an almost-idle cluster — this is the consumer's own serialisation, exactly
  as the entry above says, and 029 recovery would have changed nothing.

**Variance is large and worth expecting.** Earlier the same day this queue drained
85 → 5 in roughly 50 minutes (~1.6 msg/min, ~8× faster) when the queued work was
short. A post-deploy burst of council-shaped work collapsed it to 0.21. So "how
long will my submission wait" has no stable answer — read `LAG` *and* consider what
kind of work is in front of you.

**Practical consequence for any thread dispatching at the cluster:** a live
verification that needs a round trip through this topic can be a **multi-hour**
proposition with no failure signal at all — the run simply does not exist yet.
Check `LAG` before you start, and do not read an absent `orchestration_states` row
as a dropped spawn (that misreading is already a `WRONG_CALLS.md` row, twice).

## CORRECTION 2026-07-20 (bugfix-030 thread) — the 0.21 msg/min figure is an artifact; ~2.4 msg/min measured

> **This corrects the "Measured throughput, 2026-07-20 19:16–19:31 UTC" section
> immediately above, not the entry as a whole.** That section's *mechanism*
> conclusion is right and is independently confirmed below. Its *rate* — and the
> "~6.5 hours" wait that follows from it — comes from a mistimed first reading.

**What is wrong.** That section's two readings are, exactly and to the digit, two
samples from a continuous 30-second sampling run this thread had going at the same
time:

```
their "19:16"  current=96013  end=96099  lag=86   <- my sample at 19:29:48 UTC
their "19:30"  current=96016  end=96102  lag=86   <- my sample at 19:30:57 UTC
```

So the two readings are **69 seconds apart, not 14 minutes.** The first one cannot
have been taken at 19:16, because the offset was pinned at **96010** for the whole
of 19:21:16–19:28:40 (15 consecutive samples) — for `current=96013` to be true at
19:16 the committed offset would have had to run *backwards*, which it does not do
absent a group reset. Full sample file:
`docs024_key_docs_latest/dispatch_queue_serialisation/EVIDENCE_lag_samples_2026-07-20.txt`.

3 messages / 69 s ≈ **2.6 msg/min**, not 0.21. The stated figure is ~12× too slow
and the ~6.5-hour queue estimate is wrong by the same factor.

**What the rate actually is.** Continuous sampling, 19:10:37 → 19:32:39 UTC
(22 minutes, offset 95963 → 96016): **53 messages ≈ 2.4 msg/min.** At that rate a
queue 86 deep clears in **~36 minutes** — which lands squarely on the **25–36 min**
this entry measured on 2026-07-19. The original figure was right; nothing had
degraded by an order of magnitude.

**Why the error is easy to make, and the rule that prevents it.** Two point
readings cannot measure this queue, because the drain is a **sawtooth**: it pins at
one offset for minutes while a single long message is processed, then bursts.
Measured here: pinned at 95963 (≥2 min), then **+47 messages in ~8.5 min**, then
pinned at 96010 for **~8 min**, then moving again. Any two samples that happen to
straddle a stall give an arbitrarily low rate; two inside a burst give an
arbitrarily high one. This is the same trap as the frozen-offset landmine above,
one level up: **a static offset misleads about liveness, and two static offsets
mislead about throughput.** Sample continuously for ≥20 min and take the slope —
see RUNBOOK R2 in the workstream dir.

**What that section gets right, and this thread confirms from the source.** Its
inference that "the rate is set by orchestration DURATION, not by queue mechanics …
consistent with the consumer running each orchestration to its wait point before
taking the next message" is **correct**, and was reached independently here by
reading the code rather than the clock:

- `platform/agentbase/agent.go` `processRequests()` calls `processMessage()`
  **synchronously** — the loop cannot fetch again until the message is fully handled.
- `platform/orchestration/coordinator.go` `continueExecution()` is a `for {}` loop
  calling `executeStep` repeatedly, advancing consecutive steps **inline**;
  `grep -n "go func\|go p\.\|go c\.\|go sc\." platform/orchestration/coordinator.go`
  returns **nothing** — that file spawns no goroutines at all.
- One local LLM step measured at ~28 s in the pod log
  (`ai_actions.go:234 Rendered prompt template` 19:14:16 → `ai_actions.go:479 LLM
  response received` 19:14:44), with `coordinator.go:923 Transitioning to next step`
  immediately after.

So throughput really is ≈ `1 / (mean duration of an inline step-run)`, and the
8-minute stalls are single messages executing multi-step segments.

**Consequence for "Fix candidates" above — candidate 2 is not sufficient alone.**
`agent-chassis` runs `replicas: 1` with **no HPA** (verified 2026-07-20:
`kubectl -n ai-persona-system get deploy agent-chassis -o jsonpath='{.spec.replicas}'`
→ `1`; `get hpa` → no resources). Raising `PartitionCount` would hand every
partition to that single consumer, which would still run one blocking goroutine —
**no throughput gain, and partition count cannot be reduced later.** Whatever else
is done, the replica count and the synchronous handler have to be part of it.
**[This specific claim is UNCONFIRMED — filed for diagnosis 2026-07-20, corr
`78470372-7617-40e4-888c-66cac94006bf`, rather than asserted.]**

**Also: the frozen-offset landmine above explains the right symptom with the wrong
mechanism.** It states offsets "are committed *after* a message is fully
processed". They are not — `platform/kafka/consumer.go` `Consume()` calls
`CommitMessages` **immediately after `FetchMessage`, before the message is returned
to the caller**, so the commit happens before `processMessage` runs (the adjacent
comment, "After successful processing, commit the offset", is inaccurate). The
landmine's practical advice is still correct, for a different reason: the loop is
blocked inside `processMessage` and therefore never calls `Consume()` again, so no
new offset is fetched or committed. (The commit-on-fetch behaviour is at-most-once
delivery and is already recorded as a root cause in the bugfix-003 spawn-loss
workstream — not new, but not what this entry says.)
