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

## Related

- `/bugs_open/019` — one truncated reviewer voids a whole council round. **Not the same bug, but it
  compounds this one**: both of this session's submissions waited ~30 min in this queue and then
  died on 019, so the round cost two waits and two sets of reviewer LLM calls for zero verdicts.
- `/bugs_open/006` — idea.uk infra errors, including claim-timeout churn (a different dispatch-layer
  problem; do not conflate).
- `docs024_key_docs_latest/idea_uk_vm_site/RUNNING_NOTES_idea_uk_vm_site.md` §X.1 — the full
  misdiagnosis, including the reasoning error ("first in a `kcat -o -60` window" means *oldest
  unprocessed*, not *skipped*) and the transferable rule.
