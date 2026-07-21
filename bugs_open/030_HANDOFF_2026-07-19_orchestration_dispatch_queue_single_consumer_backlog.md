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

> **WRONG — see the CORRECTION section below before using any number in here.**
> The rate (0.21 msg/min) and the ~6.5h queue estimate are artifacts: my two
> readings were ~69 seconds apart, not 14 minutes, because I *inferred* the first
> reading's timestamp from when I thought I had launched the sampling job instead
> of recording it. Author's note, left at the head of my own section so nobody
> takes the figure from here and misses the correction. The mechanism conclusions
> below stand; the arithmetic does not. — bugfix-036 thread

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

> **MY ~2.4 msg/min IS ALSO WRONG — see "CORRECTION OF THIS CORRECTION" at the end
> of this file before using any number from this section.** The invalidity of the
> 69-second derivation below stands. My *replacement* figure does not: I computed it
> over a window that contained a burst and stopped before the stalls that followed.
> Sustained rate is far lower and the queue was *diverging*, so "nothing had
> degraded" was the wrong conclusion — and the bugfix-036 thread's qualitative
> reading was closer to right than mine. Author's note, left at the head of my own
> section. — bugfix-030 thread

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

## CORRECTION OF THIS CORRECTION — 2026-07-20 (bugfix-030 thread): my ~2.4 msg/min was an artifact too

I corrected the 0.21 msg/min figure above and replaced it with **~2.4 msg/min, "a
queue 86 deep clears in ~36 min", "nothing had degraded by an order of magnitude"**.
**My figure was wrong in the same way theirs was, and I made the error while writing
the section warning about it.** Owning it here because my number is the one that was
left standing.

**What I did.** I computed the rate over 19:10:37 → 19:32:39 — a window that
contained the **+47-message burst** and stopped just after it, before the stalls
that followed. That is the mirror of straddling a stall: I straddled a *burst*. I
had already written "any two samples inside a burst give an arbitrarily high rate"
two paragraphs earlier and then did exactly that.

**What the completed 20-minute run actually shows** (`EVIDENCE_lag_samples_2026-07-20.txt`,
now the full 40 samples rather than the 21 I had when I wrote the correction):

```
19:21:16  96010  lag  68   <- offset pinned
19:29:14  96013             (8.0 min stall)
19:35:51  96024  lag  88
19:44:01  96024  lag 109   <- pinned again
19:51:16  96024  lag 130   <- still pinned, 15.4 min and counting
```

| window | messages | rate | LAG |
|---|---|---|---|
| 19:12:51 → 19:21:16 (the burst I keyed on) | 47 in 8.4 min | **5.6/min** | 96 → 68 |
| 19:21:16 → 19:44:01 (the full sampler run) | 14 in 22.8 min | **0.62/min** | 68 → 109 |
| 19:10:37 → 19:51:16 (everything I sampled) | 61 in 40.7 min | **1.50/min** | 90 → 130 |

**So: LAG grew 82 → 130 across my session. The queue was diverging, not steady.**
"Clears in ~36 min" was false — over the 40 minutes I watched, it never cleared and
got 48 messages deeper. Two stalls of 8.0 and ≥15.4 minutes, one message each.

**The bugfix-036 thread was closer to right than I was.** Their *derivation* was
invalid — that finding stands, and their two readings really were 69 s apart — but
their **conclusion** ("slow, multi-hour possible, variance is large, no stable
answer") matches the sustained behaviour better than my "nothing had degraded". I
corrected their arithmetic and in doing so overturned a conclusion that was sound.
That is the more expensive half of the mistake, and it is the half I got wrong.

**The real lesson, which neither of my previous write-ups reached: an average rate
is the wrong statistic for this queue.** It is not that one of us picked a bad
window — it is that **no single rate exists to be measured**. Throughput here is
`1 / (duration of the orchestration segment currently executing inline)`, and that
duration ranges from milliseconds to ≥15 minutes depending entirely on what work
happens to be at the head. Averaging across it produces a number that describes no
moment and predicts nothing. Three threads have now produced three different
"measured" rates (0.21, 2.4, 0.62) from the same queue on the same afternoon, all
arithmetically defensible, all useless as forecasts.

**What to do instead — this replaces RUNBOOK R2's "take the slope":**

- **For "is my dispatch queued or lost?"** — `LAG > 0` is the answer. Rates are
  irrelevant to that question, which is the one that actually costs sessions time.
- **For "how long will I wait?"** — there is no reliable answer, and the honest
  advice is to say so. Read `LAG` for depth, and check what *kind* of work is ahead
  (`orchestration_states` where status = `EXECUTING_STEP`); a queue of council or
  diagnosis orchestrations is a different proposition from a queue of fast ones.
  Do not quote a figure from this file as a forecast — including mine.
- **Do not report a rate from a window shorter than several full stall/burst
  cycles**, and if you report one at all, publish the raw samples next to it.

The mechanism findings in my correction above are unaffected — those came from
reading `agent.go` and `coordinator.go`, not from the clock, and the divergence
measured here is consistent with them: a 15-minute stall is one message running a
long inline step-run, which is precisely what the code predicts.

---

## Update 2026-07-20 evening — it is not always bursty. Under load it DIVERGES, and the council becomes unusable.

The original measurements (lag 41 → 62 → 24, "bursty rather than permanently diverging") were taken
in a quiet period and **understate the problem**. Same day, evening, with several sessions active:

| time (UTC) | CURRENT-OFFSET | LAG |
|---|---|---|
| 18:13 | 95915 | 21 |
| 18:26 | 95958 | 67 |
| 20:03 | 96029 | **161** |

**Consumption ≈ 0.73 msg/min** (71 messages in 97 minutes). Production over the same window ≈ 1.7
msg/min. **Production is outrunning consumption by ~2.3×, and the backlog grew 21 → 161 in under two
hours.** It is not draining; it is falling behind.

### Measured end-to-end latencies now total four, and the trend is bad

| submission | published (UTC) | orchestration created | latency |
|---|---|---|---|
| council-gate-132453 (2026-07-19) | 12:24:53 | 13:01:16 | 36 min 23 s |
| council-gate-134936 (2026-07-19) | 12:49:36 | 13:14:37 | 25 min 01 s |
| council-gate round 2 (2026-07-20) | 18:07:00 | 18:23:46 | 16 min 46 s |
| council-gate round 3 (2026-07-20) | 19:22 | **still queued at 20:03** | **>40 min, 35 messages still ahead** |

At the observed 0.73 msg/min, the last one starts roughly **48 minutes after** the check above, i.e.
~90 minutes end-to-end.

### The operational consequence, stated plainly

**A council review currently costs an hour or more of wall-clock before it starts.** A REVISE verdict
therefore has a ~2-hour round trip, and the revise→resubmit loop the gate is built around becomes
impractical during busy periods. That is a usability failure of the review mechanism itself, caused
entirely by this queue — nothing about the council is slow (the round that did run took 14 minutes
end to end).

### This sharpens fix candidate 1

Printing the current LAG at publish time is no longer just a diagnosability nicety — at lag 161 the
trigger scripts are telling operators "Submitted. Watch the run:" alongside a query that will return
nothing for an hour and a half. **The script should print the lag and the implied wait**, so the
operator can decide whether to submit at all right now. That is a few lines in
`097_TRIGGER_council_review_v1.sh` / `090_TRIGGER_needs_diagnosis_v1.sh` and needs no topology change:

```bash
LAG=$(kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group 2>/dev/null | awk '/generic\.requests/{print $6}')
echo "Queue depth ahead of you: ${LAG} messages (~$((LAG * 4 / 3)) min at recent throughput)"
```

### Still true, and still the thing not to do

Do **not** re-fire a queued dispatch. At lag 161 the temptation is strongest and the cost is highest:
a duplicate spends the same LLM credits and lands even further back in the same queue.

## NEW FINDING 2026-07-20 (bugfix-030 thread) — the lane is 93% cron, and two 60-second jobs outproduce the consumer

Everything above (including my own corrections) argues about *how fast the consumer
drains*. Nobody had asked **what is actually in the queue**. It is not sessions
competing with each other. It is two scheduled jobs.

### What is in the lane [VERIFIED — message headers, not inference]

`kcat -o -300` on `system.agent.generic.requests`, headers only:

| producer | share |
|---|---|
| `from_agent_type=kafka-scheduler` | **550 / 588 (93%)** |
| `from_agent_type=user` | 38 / 588 (6.5%) |

By workload (`orchestration_name`, suffix stripped):

| job | count | |
|---|---|---|
| `sched-ai-endpoint-health-check` | 258 | **43%** |
| `sched-build-pipeline-trigger` | 244 | **41%** |
| `sched-diagnose-pipeline-trigger` | 36 | 6% |
| `council-gate` | 28 | 5% |
| `sched-stale-orchestration-reaper` | 10 | |
| `needs-diagnosis` | 6 | |

**Two scheduled jobs are 84% of the traffic.** Interactive work — the council and
diagnosis dispatches this bug is *about* — is a rounding error queued behind them.

### Their cadence [VERIFIED — distinct fire-time suffixes]

```
ai-endpoint-health-check : 19:57:06 19:58:05 19:59:05 20:00:06 20:01:05 ...
build-pipeline-trigger   : 19:56:35 19:57:35 19:58:36 19:59:35 20:00:35 ...
```

Both fire **exactly every 60 seconds**. That is a hard floor of **2 msg/min**
injected by cron, continuously, regardless of whether the consumer can keep up.

### The mass balance closes [VERIFIED — offsets, 19:05 → 20:15 UTC, 70 min]

| | offsets | rate |
|---|---|---|
| produced | 96040 → 96222 (+182) | **2.60 /min** |
| consumed | 95958 → 96058 (+100) | **1.43 /min** |
| shortfall | | **1.17 /min** |

Observed `LAG` over the same window: **82 → 164 = +82 in 70 min = +1.17/min.**
The accounting closes to two decimal places, which is the strongest evidence in
this file that nothing else is going on.

> **Note on why quoting rates here is legitimate when I just argued it is not.**
> This is a **mass balance over one closed window with the endpoints stated** — an
> account of where 82 messages came from, not a forecast. It is exactly the use my
> retraction above permits and the use it forbids is the other one: do not carry
> "1.43/min" forward as an ETA. It will be wrong the moment the head of the queue
> changes.

Corroboration from the other side (`orchestration_states` starts per hour — an
**upper bound** on this lane's consumption, since it also counts orchestrations
started from `job.*` topics):

```
15:00  2.52/min      18:00  1.38/min
16:00  2.92/min      19:00  1.55/min
17:00  1.37/min      20:00  0.52/min (partial)
```

Consumption exceeded 2.6/min until ~17:00 and has been below it since. **The
backlog has been growing structurally for roughly three hours**, and will keep
growing until either the head of the queue gets cheap or the cron rate drops.

### Blast radius — 030's "everything funnels through" is too broad [VERIFIED]

Spawned agents do **not** use this lane. `spawnAgentKubernetesJobFromDefinition`
(`platform/orchestration/actions/spawn_actions.go:2320`) launches each spawned agent
as its own **Kubernetes pod** with its own `job.<id>-<name>.requests` topic — 815
topics exist, and 8 `agent-*` job pods were running during this measurement. So the
single lane carries **top-level orchestration starts only**; per-agent work runs
genuinely concurrently in its own pods.

That narrows the bug and makes it *sharper*: the contention is not session-vs-session,
it is **cron-vs-everything**, in the one lane every new dispatch must pass through.

### What this implies for the fix candidates above

- **Candidate 3 (separate lanes) is promoted, and now has a specific shape.** Give
  `kafka-scheduler` its own topic and consumer, and interactive dispatches stop
  queueing behind ~500 health checks. This is reversible, needs no partition-count
  decision, and addresses 84% of the volume.
- **Cheaper still, and worth checking first:** does an AI-endpoint health check need
  to run *every 60 seconds*, and does a build-pipeline trigger? At 1/min each they
  consume the entire serial capacity of the platform's only dispatch lane. Halving
  the cadence would restore headroom today with a config change and no code.
- **Candidate 2 (partitioning) drops further.** It is one-way, and with `replicas: 1`
  it changes nothing — but note it would not help *even with more replicas* while two
  cron jobs are 84% of the volume; it would just spread cron across more consumers.

**[UNVERIFIED]** — I have not checked what either scheduled job *does*, whether the
health check is idempotent/cheap, or whether anything depends on the 60 s cadence.
That is the next thing to establish before touching either.

### Follow-up: the scheduler config, and a landmine in it (same thread, 2026-07-20)

Resolving the `[UNVERIFIED]` above. `scheduled_tasks` (schema via `\d` first —
there is no `schedule` column; the field is `interval_seconds`):

```sql
SELECT name, interval_seconds, enabled, target_topic, concurrency_group, max_concurrent, timeout_seconds
FROM scheduled_tasks ORDER BY enabled DESC, interval_seconds;
```

**Twelve enabled tasks target `system.agent.generic.requests`.** The two that
dominate the lane:

| name | interval_seconds | max_concurrent | timeout_seconds |
|---|---|---|---|
| `ai-endpoint-health-check` | **30** | 1 | 15 |
| `build-pipeline-trigger` | **30** | 8 | 300 |

Nominal configured load across all twelve is **≈6 msg/min** into a lane whose single
consumer was measured draining **1.43/min**. Even the observed 2.6/min is ~1.8× the
consumer.

> **LANDMINE — the queue is currently protected by a scheduling accident, and
> "fixing" it would double the load.** Both jobs are configured at
> `interval_seconds = 30`, and the scheduler's `TICK_INTERVAL_SECONDS` is also
> **30** (`kafka-scheduler` deployment env; `cmd/scheduler/main.go:134`). The due
> check is `last_triggered_at + interval <= NOW()`, so the tick that lands at the
> boundary is marginally *early* and the task is not yet due — it fires on the
> **next** tick instead. Net effect: **a 30 s task actually fires every 60 s.**
> Confirmed against the wire — successive fire times are 59–60 s apart, drifting a
> second at a time (`19:57:06 19:58:05 19:59:05 20:00:06`), which is the signature
> of exactly this aliasing.
>
> So the lane is receiving **half its configured rate**. Anyone who "corrects" the
> off-by-one (or drops the tick to 10 s, or changes the comparison to `<`) will
> **double production to ~4 msg/min against a ~1.4 msg/min consumer** and turn a
> slowly-growing backlog into a fast-growing one. Do not treat the 30 s config as
> the effective rate, and do not tidy this without reading this entry.

**This makes the cheapest fix cheaper still.** `interval_seconds` is a DB column, so
raising it is **live immediately, no image roll** (per `CLAUDE.md`). Setting the two
30 s jobs to something proportionate to what they actually need would restore
headroom today. But note the aliasing when choosing a value: the effective period is
the next tick boundary *after* the interval elapses, so with a 30 s tick the real
options are 60 s, 90 s, 120 s… — asking for 45 s gets you 60 s.

**Two things still [UNVERIFIED]** and worth establishing before changing either:
- What `ai-endpoint-health-check` actually checks, and whether a 15 s timeout on a
  30 s cadence implies anything depends on that frequency. Its handler is
  `platform/orchestration/actions/check_endpoint_health_action.go`.
- Whether `build-pipeline-trigger`'s `max_concurrent = 8` interacts with
  `/bugs_open/029` (hung spawns saturating the dispatch group) — same
  `concurrency_group = 'dispatch'`, and 029 is about that group filling up.

### Follow-up 2: what the two jobs do — and 030 is the mirror image of 029 (same thread, 2026-07-20)

Resolving the two `[UNVERIFIED]` items from the previous section.

**`ai-endpoint-health-check`** (`check_endpoint_health_action.go`) pings each active
row in `ai_endpoint_health` (Ollama `GET /api/tags`; Claude a 1-token haiku,
~$0.000003) and updates the health table. It is a **single local action**, 15 s
timeout, and the *action itself* re-checks a per-endpoint interval, so most fires do
little work. It has **no `pre_query`**, so it fires **unconditionally every cycle**
regardless of whether anything needs checking. Cheap per message, but it still costs
a queue slot and a consumer turn every 60 s. It is **not** the head-of-line blocker —
the 8–15 min stalls are the council/build/diagnosis orchestrations, not this.

**`build-pipeline-trigger`** *is* gated — its `pre_query` fires a message only when
triaged build work exists:

```sql
SELECT COUNT(*)::text FROM sites s WHERE s.locked_at IS NULL
AND EXISTS (SELECT 1 FROM site_work_items wi WHERE wi.site_id=s.id
  AND wi.status='triaged' AND wi.pipeline='build' AND wi.attempt_count < wi.max_attempts)
HAVING COUNT(*) > 0;
```

Right now that returns **3**. So it is not firing blindly — there is a **persistent
build backlog** feeding it, and that is why it is 41% of the lane.

**The coupling, and why this bug is partly self-sustaining.** A page build is itself
a multi-step orchestration that runs on this same lane. If the lane is slow, build
work items stay `triaged` longer, so the `pre_query` keeps returning > 0, so
`build-pipeline-trigger` keeps firing every 60 s. **The slower the lane, the more
persistently the trigger feeds it.** It is capped (`max_concurrent = 8` on the
`dispatch` group, so concurrent build-dispatch orchestrations cannot exceed 8), not
runaway — but the build-trigger contribution will not self-limit while any backlog
exists, which under load it always will.

**030 is the mirror image of `/bugs_open/029`.** Both are failure modes of the same
`dispatch`/generic-lane machinery, at opposite ends:

| | `/bugs_open/029` | this bug (030) |
|---|---|---|
| what saturates | the `dispatch` **concurrency group** (scheduler-side gate) | the generic **requests topic** (the lane itself) |
| trigger | hung spawns (`/bugs_open/003`) fill the group to `max_concurrent=8` | cron out-produces the single consumer |
| symptom | `build-pipeline-trigger` **can't fire** → builds halt fleet-wide, silently | everything **queues** → tens-of-minutes latency |
| when | spawns are hanging | spawns are healthy and the cluster is busy |

They can even alternate on the same job: when spawns hang you get 029 (nothing
fires); when they recover, the backlog that accumulated floods the lane and you get
030 (everything queues). A fix for either should be checked against the other — e.g.
moving `kafka-scheduler` to its own lane (030 candidate 3) must preserve 029's
`dispatch`-group gate, or hung spawns would no longer be capped at 8.

Both `[UNVERIFIED]` items are now resolved. What remains genuinely unverified is only
the core root-cause claim (inline handler vs `PartitionCount`), which is the subject
of diagnosis corr `78470372` — still unstarted, queued in this very backlog.

## CORRECTION 2026-07-21 (bugfix-030 thread) — the "aliasing landmine" was a special case of a UNIVERSAL +1-tick offset; and a config fix is now live

> **This corrects the "tick aliasing" landmine two sections up.** That entry said a
> 30 s task fires every 60 s because `interval == TICK_INTERVAL_SECONDS` is a special
> aliasing case, and implied that setting `interval_seconds` to a clean multiple of
> the tick removes it. **Both halves are wrong.** Verified today by changing the two
> jobs and measuring, and confirmed against the scheduler source.

**The real rule: effective fire period = `interval_seconds + TICK_INTERVAL_SECONDS`,
for any interval that is a multiple of the tick.** Measured, three points:

| interval_seconds | TICK | measured effective period |
|---|---|---|
| 30 | 30 | **60 s** (yesterday) |
| 60 | 30 | **90 s** (today: gaps 90,90,91,89) |
| 120 | 30 | **150 s** (today: gaps 151,149) |

Evidence: `docs024_key_docs_latest/dispatch_queue_serialisation/EVIDENCE_fire_intervals_2026-07-21.txt`.

**Why (from `cmd/scheduler/main.go`, not inferred):** `last_triggered_at` is stamped
`NOW()` at **fire time** — late in `runTick`, after `loadDueTasks`, the concurrency
check, the pre-query and `fireTrigger`. But the due test `last_triggered + interval
<= NOW()` is evaluated by `loadDueTasks` at the **start** of a later tick. At the
exact boundary tick (`last + interval` landing on a grid tick), the stamped-late
timestamp is reliably a hair greater than the checked-early `NOW()`, so `<=` fails
and the task slips to the **next** tick. The extra tick is therefore **not** special
to `interval == tick` — it happens at every multiple. `interval == tick` only *looked*
special because 30→60 is a doubling; 60→90 is ×1.5 and 120→150 is ×1.25, because the
+30 is absolute, not proportional.

**So my earlier advice "pick a clean multiple of the tick" was exactly backwards** —
multiples are the case that lands on the boundary and takes the extra tick. To get a
target effective period **P** (a multiple of 30), set **`interval_seconds = P − 30`**.
(A non-multiple just rounds up to the next grid tick with no extra: `interval = 45`
gives 60 s, not 90.)

### The fix that is now LIVE (config only, no image roll)

```sql
UPDATE scheduled_tasks SET interval_seconds = 60,  updated_at = now() WHERE name = 'ai-endpoint-health-check';  -- was 30
UPDATE scheduled_tasks SET interval_seconds = 120, updated_at = now() WHERE name = 'build-pipeline-trigger';    -- was 30
```

Effective periods measured after: health-check **90 s**, build-pipeline-trigger
**150 s** (both `= interval + 30`). Load effect:

| | before (effective 60 s each) | after |
|---|---|---|
| ai-endpoint-health-check | 1.00/min | 0.67/min |
| build-pipeline-trigger | 1.00/min | 0.40/min |
| two dominant, combined | 2.00/min | **1.07/min** |

Estimated total scheduled production **2.60 → ~1.67/min**. The single consumer was
measured at ~1.43/min *under congestion* — and that figure rises as fewer expensive
build chains are started (build-trigger now fires 2.5× less often), so the two should
now be much closer to balanced, possibly balanced. **First post-change LAG reading:
20** (was 82→168 and diverging yesterday) — but that is **one reading** and overnight
quiet-hours draining is a confounder, so it is consistent-with-improvement, **not**
proof. A proper before/after needs a LAG trend across a busy window (RUNBOOK R7/R2
caveats apply).

**Why these two, and why the values are safe:**
- `ai-endpoint-health-check` is the only high-frequency **unconditional** job. 90 s
  still honours the endpoints' own `check_interval_seconds` (claude 3600 s,
  cpu-ollama 60 s → now checked at 90 s, mildly stale but fine for a health signal;
  gpu-ollama wants 30 s but is `healthy=f` and was already at 60 s). [VERIFIED]
- `build-pipeline-trigger` starts the **expensive** multi-step build chains — the
  ones that hold the single consumer for minutes. Firing them 2.5× less often is the
  lever that matters; build-pickup latency rises to ≤150 s, fine for minute-scale
  builds. Capped by `max_concurrent = 8` (`dispatch` group), unchanged, so no
  interaction with `/bugs_open/029`'s gate.

**Reversible:** `UPDATE scheduled_tasks SET interval_seconds = 30 WHERE name IN
(...)` — but that restores the higher load (and the 60 s effective period).
