# 096 — a long council run still head-of-line blocks every other generic dispatch

**Filed** 2026-07-26 from the oufe.com workstream.
**Relationship to 030** — `bugs_closed/030` is genuinely fixed and live: **cron got
its own lane**, and publish→start for a scheduled trigger went from ~18 min to
~1 s. This is the part that fix did not address, observed twice in the two days
after it closed. Not a regression; a named residual.
**Severity** medium — latency and diagnosability, not correctness. It costs
sessions time and it makes a correct change look broken.
**Status** OPEN. **Re-checked 2026-07-27 after the `v1.0.1174` fleet roll: still
open, structure unchanged, mechanism reproduced with ms evidence** — see
§"Re-checked 2026-07-27". That section also records a trap that nearly refuted
this bug wrongly: `orchestration_states.requests_topic` is not the lane a message
arrived on.

## Symptom

`system.agent.generic.requests` has one partition and one consumer. A single
long-running submission on it stalls every unrelated message behind it. Twice:

| date | blocking message | offset | effect |
|---|---|---|---|
| 2026-07-25 | `council-gate` submission, 20,178 bytes (robot-hands gripper dossier, round 2) | 104102 | a site submission queued **~28 min**; lag climbed 13 → 28 |
| 2026-07-26 | `council-gate` submission (bugs_open/043 candidates, round 1) | 105214 | three page re-renders queued; lag 2 → 3 and static |

Both cleared the instant the blocking council reached a terminal step. Nothing was
dropped and nothing was malformed in either case.

While blocked, **24 orchestrations from other paths completed normally in the same
15-minute window** — so "the platform is busy" and "your message is stuck" look
identical from every vantage point except consumer-group lag.

## Why a council run in particular

A 16-seat council is a sequence of LLM calls; a single run occupies the consumer
for tens of minutes. It is the longest-running thing routinely published to this
topic, so it is the one that most often becomes the head of the line. The
mechanism is not council-specific — anything slow on this topic does it — but in
practice councils are the observed cause both times.

There is an unpleasant interaction with the council's own norm: CLAUDE.md
encourages putting platform changes through the gate, so the more the estate
follows its own advice, the more often this lane is occupied.

## The diagnostic that settles it, and the two that do not

```bash
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group
```
A committed offset that does not advance while `LOG-END-OFFSET` climbs is the
whole diagnosis. Read the blocking message with
`kcat -C -t system.agent.generic.requests -p 0 -o <committed-offset> -c 1`.

Two things that mislead, both of which cost time here:

- **`processed_messages` is not a was-it-consumed oracle.** Neither the blocked
  message nor the scheduler messages around it had rows, while those schedulers
  were demonstrably running. It records a narrower path and logs
  `DEDUPE_SKIPPED_NO_REQUEST_ID` when it records nothing at all.
- **An absent `orchestration_states` row proves nothing.** It is the expected
  state of a queued message. **Never re-fire on that evidence** — the duplicate
  queues behind the same blockage and then does the work twice.

## Fix candidates, ordered by what closes the door

1. **Give long-running submissions their own lane**, exactly as 030 did for cron:
   an `EXTRA_REQUEST_TOPICS` entry for council/diagnosis dispatches. This is the
   proven shape, it is config, and it makes the interference structurally
   impossible rather than merely rarer.
2. **Partition `system.agent.generic.requests`** and scale the consumer group.
   Larger change; interacts with the chassis-replica-scaling work, where
   `coordinator.go:271` discards non-owner responses, so a shared group alone is
   recorded as unsafe.
3. **Surface the wait**: have publishers print current lag, so a caller sees
   "queued behind N" instead of silence. Does not fix it, but removes the
   misdiagnosis, which is most of the real cost.

Candidate 1 is the smallest change that removes the class.

## Re-checked 2026-07-27 after a fleet roll: STILL OPEN, mechanism reproduced

Checked by the bug-backlog triage sweep the afternoon the fleet rolled to
`v1.0.1174` (all pods restarted 15:11 UTC). **A pod roll rejoins the consumer
group and drains any backlog, so "lag is 0" measured after one proves nothing** —
that is why the structural facts and a pre-roll reproduction are recorded here
instead of a lag reading.

### The structure is unchanged — the fix is not applied

```
kafka-topics.sh --describe --topic system.agent.generic.requests
  -> PartitionCount: 1     ReplicationFactor: 3

kafka-consumer-groups.sh --describe --group generic-requests-group --members
  -> exactly ONE member: agent-chassis@agent-chassis-5994dc6d6c-pt8v9, #PARTITIONS 1

kubectl get deploy agent-chassis -o jsonpath=… env
  -> EXTRA_REQUEST_TOPICS=system.agent.scheduled.requests      (cron only — 030's lane)
```

One partition, one consumer, and **no extra lane for council/diagnosis
dispatches**, so fix candidate 1 has not been done.

And the serialisation is explicit in the code, not inferred —
`platform/agentbase/agent.go:611-619`, the comment on `consumeRequestLane`:

> *"Each lane runs one of these on its own goroutine, so lanes make progress
> independently. **Within a lane, processing stays strictly serial and in
> order** — that is unchanged…"*

`a.processMessage(msg, "request")` is a blocking call in the fetch loop and the
offset is committed after it returns (`:672-675`). One slow message on this topic
is the whole lane.

### Reproduced 2026-07-27, 12:54–13:29 UTC (before the roll)

Six consecutive generic-lane runs, ms-precision, from `orchestration_states`
(`created_at` is when the message was **consumed**, not when it was published):

| orchestration | agent | created | ended | gap from previous end |
|---|---|---|---|---|
| `614f24fd` | council-gate | 12:54:08.905 | 13:00:47.254 | — |
| `e5ad20ea` | council-gate | 13:00:47.367 | 13:08:35.213 | **113 ms** |
| `1b5c172e` | generic | 13:08:35.291 | 13:08:57.730 | **78 ms** |
| `55f8a2a7` | council-gate | 13:09:21.396 | 13:15:35.099 | 24 s |
| `1ebe83ba` | council-gate | 13:15:35.220 | 13:22:54.832 | **121 ms** |
| `2aefeb00` | council-gate | 13:22:54.958 | 13:29:24.722 | **126 ms** |

Four hand-offs at 78–126 ms. A message that starts a tenth of a second after the
previous one finished was **already waiting**; it did not arrive then. The
`generic` run at 13:08:35 is the same shape as the symptom table above — an
unrelated dispatch released the instant a council reached a terminal step. Council
runs on 07-27 took **5–9 minutes** each (not the tens of minutes of the 07-25
observation), so the per-event cost is currently lower than when this was filed;
the mechanism is identical.

### TRAP: `orchestration_states.requests_topic` is NOT the lane the message arrived on

This cost real time here and would have produced a confident refutation of the
whole bug. The column records the agent's own configured requests topic (its
`REQUESTS_TOPIC` env), so `build-pipeline-trigger`, `endpoint-health-checker`,
`content-feed-trigger`, `index-orchestrator`, `model-directory-trigger` and
`work-item-archiver` all read `requests_topic = system.agent.generic.requests`
— and every one of them actually arrives on **`system.agent.scheduled.requests`**,
030's separate lane, on its own goroutine.

The consequence: those agents interleave freely with a running council, which
looks exactly like "the generic lane is not blocked" and is not evidence about
this lane at all. **The authority is `scheduled_tasks.target_topic`:**

```sql
SELECT name, target_agent_type, target_topic FROM scheduled_tasks WHERE enabled;
-- 18 of 23 enabled tasks -> system.agent.scheduled.requests
--  4 -> system.agent.business-intel.requests
--  1 -> system.internal.noop
--  ZERO enabled scheduled task fires onto system.agent.generic.requests
```

So the generic lane carries only what threads and agents publish to it by hand:
council submissions, diagnosis/feature-designer runs, direct page-rerenders
(049b), `content-reviewer`. That is a small population — which is why this bites
sessions specifically and is invisible to the pipeline's own throughput.

### Note for candidate 1

The proven shape is already in the tree: `EXTRA_REQUEST_TOPICS` is read at
`platform/agentbase/agent.go:433`, each entry gets its own goroutine running the
same `consumeRequestLane`, and `:447` carries an explicit warning —
**"NEVER put EXTRA_REQUEST_TOPICS in personae-prod-config"** — because it is
ignored (with a `Warn`) for any agent that is not statically deployed. Read
`:408-460` before adding a lane.

## How to verify a fix

Publish a council submission and, while it runs, publish a page re-render. The
re-render must start within seconds. Measure with consumer-group lag, and record
both timestamps — a single fast run proves nothing unless a long one was
genuinely in flight at the same time.

## Fourth, fifth and sixth instances — 2026-07-27, all in one working session

Contributed from the oufe.com workstream. No fix attempted; this is frequency
evidence, which the file previously lacked.

Blocked on 2026-07-27 while doing ordinary site work: three page re-renders, a
tool-page render (twice), and a `grounded-explainer` research run. Each sat with
**no `orchestration_states` row at all** while the lane ran councils, and lag
climbed 5 → 8 → 9 → 11 across the afternoon. Head of line, read from the topic:

```
offset 105214  council-gate   (bugs_open/043 candidates)
offset 105335  council-gate   ("HONEST FRAMING FIRST: this change is already committed…")
in flight      9cd19dba  EXECUTING_STEP  review_debug_historian
```

**What this adds to the case.** The file already establishes the mechanism and
the latency. What it did not have is how often a normal working session runs into
it: **six blocked dispatches in one afternoon, on a day when nothing unusual was
happening.** The estate simply followed its own norm of putting changes through
the council, and ordinary content work queued behind it each time.

The practical cost is not the wait. It is that a queued dispatch and a dropped one
look identical for the first several minutes, so the temptation each time is to
re-fire — which is the one action that makes it worse, and which this file already
warns against. On this day the warning held only because the same person had
written it the day before.

**This also lands on a user-visible outcome**, which the earlier instances did
not. The owner asked why he could not find any tools on oufe.com. The tool page
existed, its render was correct, and it stayed 404 for the rest of the session
because the lane was busy. "Nothing is broken and the page is not there" is the
shape of this bug seen from outside.

Candidate 1 (an `EXTRA_REQUEST_TOPICS` lane for council/diagnosis dispatches,
exactly as 030 did for cron) remains unapplied. It is config.
