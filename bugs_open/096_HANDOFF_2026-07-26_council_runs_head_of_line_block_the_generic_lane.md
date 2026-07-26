# 096 — a long council run still head-of-line blocks every other generic dispatch

**Filed** 2026-07-26 from the oufe.com workstream.
**Relationship to 030** — `bugs_closed/030` is genuinely fixed and live: **cron got
its own lane**, and publish→start for a scheduled trigger went from ~18 min to
~1 s. This is the part that fix did not address, observed twice in the two days
after it closed. Not a regression; a named residual.
**Severity** medium — latency and diagnosability, not correctness. It costs
sessions time and it makes a correct change look broken.
**Status** OPEN.

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

## How to verify a fix

Publish a council submission and, while it runs, publish a page re-render. The
re-render must start within seconds. Measure with consumer-group lag, and record
both timestamps — a single fast run proves nothing unless a long one was
genuinely in flight at the same time.
