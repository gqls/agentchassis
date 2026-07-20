# RUNBOOK — dispatch queue serialisation (bugs_open/030)

Every command that was hard to get right, with its gotcha attached. When one
changes, change it **here**.

---

## R1 — Is my dispatch queued or lost? (the decisive check)

```bash
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group
```

`LAG > 0` → queued, **not** lost. Do not re-submit (see 030's landmine: retrying a
queued dispatch duplicates the work and spends credits twice).

Gotchas, each of which cost time:
- Broker pod is `personae-kafka-cluster-combined-pool-prod-0`.
  `personae-kafka-cluster-dual-role-0` does **not** exist here and fails quietly
  enough to look like an empty result.
- The Kafka CLI is at `/opt/kafka/bin/`, **not** on `$PATH`.
- `--describe --all-groups` iterates every group and takes **>120 s** — always name
  the one group.

## R2 — Sampling the queue over time

> **CORRECTED 2026-07-20, same day, by the thread that wrote it.** This section
> originally ended "sample for ≥20 minutes and take the slope". **That advice is
> withdrawn.** I followed it and still got a figure ~4× off, because a longer window
> makes a rate *stabler*, not *truer* — and there is no true rate here to converge
> on. Three threads produced 0.21, 2.4 and 0.62 msg/min from this queue in one
> afternoon, all correctly computed. Throughput is
> `1 / (duration of the orchestration segment currently running inline on the
> consumer goroutine)`, which ranges from milliseconds to ≥15 minutes depending on
> what is at the head — a non-stationary signal, so its average describes no moment
> and forecasts nothing. See `WRONG_CALLS.md` (7)(8)(9).

**Sample to see the SHAPE, never to derive a rate.** The drain is a sawtooth: the
offset pins for 8–15+ minutes while one long message is processed, then bursts by
tens of messages. Two minutes of a stall looks identical to a dead consumer; two
minutes of a burst looks identical to a healthy queue; and twenty minutes of either
looks like a confident number that is wrong.

**Do not publish a rate from this.** If you must quote one, publish the raw samples
beside it and state the window. What the samples legitimately tell you: whether the
consumer is moving at all, and how long the current head message has been running.

```bash
for i in $(seq 1 40); do
  kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
    /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
    --describe --group generic-requests-group 2>/dev/null \
  | grep "system.agent.generic.requests" \
  | awk -v t="$(date -u +%H:%M:%S)" '{print t, $4, $5, $6}' >> lag_samples.txt
  sleep 30
done
```

Columns are `time CURRENT-OFFSET LOG-END-OFFSET LAG`.

Gotcha: **`grep` for the topic line, don't use `awk 'NR==2'`.** The output has a
header row and the row offset is not stable — `NR==2` silently prints the header
(`CUR=CURRENT-OFFSET`), which looks like a parse success and wastes a sampling run.
That happened here first time.

## R3 — Is the chassis alive? (never infer this from the offset)

A static `CURRENT-OFFSET` proves **nothing** about liveness — it holds still for the
entire duration of the longest message on the topic. Check the log instead:

```bash
kubectl -n ai-persona-system logs <chassis-pod> --tail=5 --timestamps
kubectl -n ai-persona-system logs <chassis-pod> --since=2m | wc -l
```

Gotcha: comparing `--since=2m`, `--since=10m` and `--since=30m` and getting the
**same count** means everything arrived in one recent burst after a long silence —
that is the normal sawtooth here, not a fault. Reading the silence as death is the
documented bugfix-028 error.

## R4 — What is currently hogging the consumer?

```sql
SELECT orchestration_id, orchestration_name, status, current_step,
       created_at, now()-created_at AS age
FROM orchestration_states
WHERE status IN ('EXECUTING_STEP','RUNNING','AWAITING_RESPONSES')
ORDER BY created_at DESC LIMIT 15;
```

Read it this way:
- `AWAITING_RESPONSES` orchestrations are **idle** — they are parked waiting on an
  external reply and are not holding the consumer.
- `EXECUTING_STEP` with a recent `updated_at` is the one running inline steps and
  therefore the one blocking the queue.
- `EXECUTING_STEP` with an age of **hours** is a zombie, not a blocker — that is the
  bugfix-003 reaper's target (`EXECUTING_STEP > 4h`), a different defect.

## R5 — Measuring how long one step actually takes

```bash
kubectl -n ai-persona-system logs <chassis-pod> --since=3m --timestamps \
| python3 -c "
import sys,json
for line in sys.stdin:
    ts,_,rest=line.partition(' ')
    try: d=json.loads(rest)
    except: continue
    print(ts[11:19], d.get('level','')[:4], (d.get('caller','') or '')[-40:], (d.get('msg','') or '')[:95])
"
```

The span that matters is `ai_actions.go:234 Rendered prompt template` →
`ai_actions.go:479 LLM response received` (one LLM step, ~28 s measured), then
`coordinator.go:923 Transitioning to next step` immediately after — that adjacency
is the inline-loop signature.

## R6 — Clock trap

Trigger scripts name orchestrations `...-$(date +%H%M%S)` in **local time (BST)**;
the DB is **UTC**. A run named `-132453` was created at `12:24:53` UTC. This will
make you think a run is an hour old or an hour in the future.

## R7 — Answering "how long will my dispatch wait?" (added 2026-07-20)

**There is no reliable answer. Say so.** This is the honest replacement for the
rate-based estimate R2 used to encourage.

What you *can* establish, and what each thing is good for:

```bash
# 1. Depth — is it queued at all? This is the question that actually matters.
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group
```

`LAG > 0` → queued, not lost. **Do not resubmit.** That is the whole diagnostic
value of this queue and it needs no rate at all.

```sql
-- 2. What kind of work is at the head — the only forecast input worth anything.
SELECT orchestration_id, orchestration_name, status, current_step,
       now()-updated_at AS since_update
FROM orchestration_states
WHERE status = 'EXECUTING_STEP' ORDER BY updated_at DESC LIMIT 5;
```

A head of council/diagnosis orchestrations is a wholly different proposition from a
head of fast ones. `since_update` on the top row tells you how long the *current*
blocking message has been running — measured at 8.0 and ≥15.4 min on 2026-07-20.

**What not to do:** do not quote a msg/min figure from any doc in this repo,
including this one, as an ETA. Three exist, they disagree by 12×, and all three were
computed correctly. If someone needs a number for planning, give them `LAG` and the
head-of-queue work type, and tell them the variance is the finding.
