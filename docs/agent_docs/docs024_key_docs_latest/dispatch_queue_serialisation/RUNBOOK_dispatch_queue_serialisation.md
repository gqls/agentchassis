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

## R8 — The config change made 2026-07-21 (and how to reverse it)

Cron/config only, no image roll — `interval_seconds` is a DB column the scheduler
reads every tick, so this is **live immediately** and trivially reversible.

**Before** (captured): both `interval_seconds = 30`.

```sql
-- APPLIED 2026-07-21. Reduces the two dominant jobs' firing rate.
-- NOTE: measured effective periods are interval + 30 (see caveat below), so these
-- fire at 90 s and 150 s, not 60 s and 120 s.
UPDATE scheduled_tasks SET interval_seconds = 60,  updated_at = now()
  WHERE name = 'ai-endpoint-health-check' AND interval_seconds = 30;
UPDATE scheduled_tasks SET interval_seconds = 120, updated_at = now()
  WHERE name = 'build-pipeline-trigger'   AND interval_seconds = 30;
```

The `AND interval_seconds = 30` guard makes each UPDATE a safe no-op if another
session has already changed the value — check `UPDATE 1` vs `UPDATE 0`.

**Why these two, and these values:**
- `ai-endpoint-health-check` is the only high-frequency **unconditional** job (no
  `pre_query` — fires every cycle regardless of work). 30→60 makes the config equal
  to the 60 s reality it already had under aliasing, and honours the endpoints' own
  `check_interval_seconds` (claude 3600 s, cpu-ollama 60 s; gpu-ollama wants 30 s but
  is unhealthy and was already at 60 s). No behavioural change, landmine removed.
- `build-pipeline-trigger` is gated (`pre_query` on triaged build work) but its gate
  is currently open, and it starts the **expensive** multi-step build chains. 30→120
  halves both its lane contribution and, more importantly, the rate at which those
  long inline orchestrations begin — which is what actually stalls the consumer.

**To reverse:**
```sql
UPDATE scheduled_tasks SET interval_seconds = 30, updated_at = now()
  WHERE name IN ('ai-endpoint-health-check','build-pipeline-trigger');
```
(But if you reverse, note you are restoring the aliasing landmine — the tasks will
again read as 30 s while firing at 60 s.)

**Choosing any future interval (CORRECTED 2026-07-21 — I had this backwards):**
the effective fire period is **`interval_seconds + TICK_INTERVAL_SECONDS`** whenever
`interval_seconds` is a multiple of the 30 s tick. Measured: 30→60, 60→90, 120→150.
The +30 is because `last_triggered_at` is stamped late in the tick (at fire time)
while the due-check reads `NOW()` early in the next tick, so the boundary tick always
fails `last + interval <= NOW()` and slips one tick — this is universal, **not** a
special `interval == tick` aliasing. So **to get a target effective period P (a
multiple of 30), set `interval_seconds = P − 30`.** A non-multiple just rounds up to
the next grid tick with no extra (45 → 60). The two live values (60, 120) are exact
multiples, so they fire at 90 s and 150 s respectively — deliberately, for headroom.

**Verify against the running scheduler, not the config:**
```sql
SELECT name, interval_seconds, last_triggered_at FROM scheduled_tasks
WHERE name IN ('ai-endpoint-health-check','build-pipeline-trigger');
```
Sample `last_triggered_at` a few times — `build-pipeline-trigger` should now advance
by ~120 s, not ~60 s.

---

## R9 — Rolling out the scheduler lane (added 2026-07-25) — ORDER IS LOAD-BEARING

The Go side (`EXTRA_REQUEST_TOPICS`, commit `f9bc7f45f`) is inert until an image
carrying it runs. `scheduled_tasks.target_topic` is a DB column and takes effect
**immediately**, so switching the producers first would publish every cron
dispatch into a topic nobody consumes — silent, and all scheduled work stops.
Image first, verify, then the UPDATE.

**1. Build and push** (`v1.0.1164` is already built locally and verified to
contain the change; the push was blocked for the session that built it):

```bash
docker push aqls/agent-chassis:v1.0.1164
```

**2. Roll the chassis ONLY** — `make deploy-agents` is fleet-wide, do not use it:

```bash
sed -i 's/newTag:.*/newTag: v1.0.1164/' \
  deployments/kustomize/services/agent-chassis/overlays/production/uk_001/kustomization.yaml
kubectl apply -k deployments/kustomize/services/agent-chassis/overlays/production/uk_001
kubectl -n ai-persona-system rollout status deployment/agent-chassis --timeout=180s
```

**3. Verify the lane is REALLY being consumed.** Three checks, and all three
matter — the pod-grep only proves the binary shipped, not that the lane came up:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')

# a) the change is in the running binary (a string this change CREATED, plus a
#    positive control so a zero means absence rather than a broken grep)
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "Extra request lane ready"; strings /app/agent-chassis | grep -c "Starting request processor"'
# expect 1 then 2

# b) the lane consumer actually started in THIS pod
kubectl -n ai-persona-system logs $POD | grep -E "Extra request lane ready|Starting request processor for extra lane"

# c) Kafka agrees the group exists (it appears once the reader joins)
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --list \
  | grep generic-requests-group-lane-system-agent-scheduled-requests
```

If (b) or (c) is empty, **stop** — do not switch any producer. The lane topic
creation is idempotent but skips-and-logs on failure by design, so an absent lane
is a logged Error in the pod, not a crash.

**4. Only then, move the cron producers** (live, no roll, reversible):

```sql
UPDATE scheduled_tasks SET target_topic='system.agent.scheduled.requests', updated_at=now()
 WHERE enabled AND target_topic='system.agent.generic.requests';
-- 18 rows as of 2026-07-25; SELECT first, another session may have added tasks
```

**5. Confirm the split is real** (both lanes, ~10 minutes apart is enough):

```bash
scripts/dispatch-queue-depth.sh                                        # interactive lane
scripts/dispatch-queue-depth.sh --topic system.agent.scheduled.requests \
  --group generic-requests-group-lane-system-agent-scheduled-requests  # cron lane
```
Expect: the cron lane carries the traffic and may show depth; the **interactive**
lane stays near zero, and a dispatch fired while a long cron orchestration runs
starts promptly instead of behind it. That last one is the test that matters —
`bugs_open/030` "How to verify a fix", negative test.

**Rollback** (either half, independently):

```sql
UPDATE scheduled_tasks SET target_topic='system.agent.generic.requests', updated_at=now()
 WHERE target_topic='system.agent.scheduled.requests';
```
and/or remove `EXTRA_REQUEST_TOPICS` from the chassis Deployment and re-apply. The
lane topic can be left in place empty; it costs nothing.

**Landmine:** `EXTRA_REQUEST_TOPICS` must stay a **direct env var on the chassis
Deployment**. Spawned agent pods inherit `envFrom: personae-prod-config`
wholesale, so putting it there would have every live spawned pod join the lane
under its own consumer group and re-run every scheduled dispatch. `agent.go`
refuses the var on non-static agents as a backstop; do not lean on the backstop.

---

## R10 — The publish acknowledgement (added 2026-07-25)

```bash
scripts/dispatch-queue-depth.sh [--topic T] [--group G] [--brief]
```

Runs in ~8 s, exits 0 always, timeout-wraps every probe: it is called from the
publish path of `090_TRIGGER_needs_diagnosis_v1.sh` and
`097_TRIGGER_council_review_v1.sh`, so it must never be able to fail a
submission. Any other trigger can call it the same way.

It answers **"queued or lost?"** and refuses to answer "how long?" — see R7 and
the retraction in `bugs_open/030`: this lane has no stable drain rate to quote,
and printing one would re-import the exact error three threads have already made.

The one case it flags as a genuine fault rather than a wait: **no member in the
consumer group**. Then nothing is draining the lane at all, and queued work will
sit there until a consumer joins.

### R9 step 6 — the starvation check (run it after ANY change to a lane's producers)

Two messages arriving proves the lane works; it does not prove nothing was
stranded. Judge every task against **its own** interval:

```sql
SELECT name, interval_seconds AS iv,
       EXTRACT(EPOCH FROM (now()-last_triggered_at))::int AS since_s,
       CASE WHEN EXTRACT(EPOCH FROM (now()-last_triggered_at)) > interval_seconds + 90
            THEN 'OVERDUE' ELSE 'ok' END AS state
  FROM scheduled_tasks WHERE enabled AND target_topic='system.agent.scheduled.requests'
 ORDER BY interval_seconds;
```

`+ 90` allows the universal `interval + TICK` offset plus a tick of slack. Measured
2026-07-26, ~15 min after the switch: **18 rows, 18 `ok`**. Anything `OVERDUE` on a
short-interval task means the lane is not being drained — check the group has a
member before anything else (`scripts/dispatch-queue-depth.sh` calls that out
explicitly, because an empty group is a fault while a non-zero LAG is not).

### R9 step 4a — use `RETURNING` on the producer-touched table (learned the hard way, 2026-07-26)

When you fire the `target_topic` switch, write it as:

```sql
UPDATE scheduled_tasks SET target_topic='system.agent.scheduled.requests', updated_at=now()
 WHERE enabled AND target_topic='system.agent.generic.requests'
 RETURNING id, name, interval_seconds;
```

I ran it without `RETURNING` and tried to reconstruct which rows changed afterwards
from `updated_at`. **You cannot**: `cmd/scheduler/main.go:273` re-stamps
`updated_at = NOW()` on **every fire** (`UPDATE scheduled_tasks SET
last_triggered_at = NOW(), last_completed_at = NOW(), updated_at = NOW()`), so the
witness column belongs to the producer, not to you. Fifteen minutes later only 7 of
18 rows still carried my statement's timestamp — the long-interval ones that had not
fired yet.

Generalises beyond this table: **on any table a live producer writes to, a
count-before plus a select-after is not decisive, and `RETURNING` has no
substitute.** A concurrent writer can hide behind a matching row count. Take the
count as a gate, but bind the rows with `RETURNING` in the same statement.
