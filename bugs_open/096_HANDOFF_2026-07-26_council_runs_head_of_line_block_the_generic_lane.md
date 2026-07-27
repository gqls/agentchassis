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


## Applying the fix — two hazards found before doing it (2026-07-27)

The manifest change is committed (`e88852825`) and deliberately **not applied**.
Two things to get right, both discovered by checking rather than by trying.

**1. Do NOT `kubectl apply -k` the overlay.** Sixteen files under `deployments/`
are modified and uncommitted by other sessions right now, mostly `kustomization.yaml`
image-tag bumps. Applying the overlay would ship every one of them to production
alongside this env var. Use a targeted edit instead:

```bash
kubectl -n ai-persona-system set env deploy/agent-chassis \
  EXTRA_REQUEST_TOPICS=system.agent.scheduled.requests,system.agent.council-gate.requests
```
Current live value, for the rollback: `system.agent.scheduled.requests`.

**2. Wait on WHAT is running, not on how much.** "Wait for an idle chassis" is
circular — the congestion this bug describes is what keeps it busy, so it is never
idle. The distinction that matters is cost of loss: a page render or dispatch loop
killed by the roll re-fires for nothing, while a council or diagnosis round costs
money and an afternoon. Gate on no `review_*`/`gate_*`/`verdict`/`route` step being
in flight, and cheap work can be running.

**Ordering, restated because it is easy to get backwards** (`agent.go:429-432`):
roll first, confirm `system.agent.council-gate.requests` has a consumer, and only
then point `097_TRIGGER_council_review_v1.sh` at it. A producer aimed at an
unconsumed topic piles messages up where nothing will ever run them.


## Why the fix has not been applied yet — measured, and it needs an owner decision

Watched the cluster for ~90 minutes on 2026-07-27 looking for a safe moment.
**It never came.** Council and diagnosis runs were in flight on every single
check: 3, 3, 2, 3, 2, 1, 2, 1, 2, 2, 2, 1, 2. Never zero.

That is the bug describing itself. The lane is busy because councils run
back-to-back, and the fix for that congestion needs a gap the congestion prevents.

**A roll does cost an in-flight run, and this is not a guess.**
`Agent.Shutdown()` (`agentbase/agent.go:1485-1505`) signals shutdown, waits
`30 seconds` for goroutines, logs `"Agent shutdown timeout"` and closes the
consumers regardless. A council step is a single LLM call that routinely runs for
minutes, so it is inside that timeout and gets cut. The pod's grace period is 60s,
which does not help: shutdown gives up at 30.

The deployment itself is safe (`RollingUpdate`, `replicas=1`, `maxSurge=25%`), so
the new pod starts before the old one is terminated and there is no window with no
consumer. The loss is the in-flight work, not availability.

### The decision

Three options, none of which a session should take unilaterally:

1. **Apply it at a genuinely quiet hour** (overnight). Costs nothing, delays the
   fix by a day. Recommended, and the change is already committed and ready.
2. **Apply now and accept one lost council round.** One session loses a review and
   re-submits; every session stops queueing behind councils from then on. A
   defensible trade, but the cost lands on somebody who did not choose it.
3. Leave it. The congestion is real and measured: seven ordinary dispatches
   blocked in one afternoon, and a live site missing a page all session because of
   it.

**Applying it is one command and does not need the overlay:**
```bash
kubectl -n ai-persona-system set env deploy/agent-chassis \
  EXTRA_REQUEST_TOPICS=system.agent.scheduled.requests,system.agent.council-gate.requests
```
Then confirm the new lane has a consumer before pointing `097` at it.

> **UPDATE 2026-07-27 ~19:40 UTC (different thread): this was applied.** Verified
> live — the deployment now carries both topics, the pod is `v1.0.1177` (started
> 19:22 UTC), and `kafka-consumer-groups.sh --list` shows
> `generic-requests-group-lane-system-agent-council-gate-requests`. `097` is still
> pointed at `system.agent.generic.requests`, i.e. consumer shipped, producer not
> yet switched — the deliberate rollout order. One council (`f849afaf`) was
> in flight at 19:18 and was indeed killed by that roll, exactly as the section
> above predicted; it sat wedged at `review_guardian` for 27+ minutes afterwards.

## Candidate 4 — run the council in a SPAWNED POD. Built and tested 2026-07-27; the fix WORKS and the spawn handshake under it DOES NOT.

Owner direction 2026-07-27: coerce the council into the standard agent workflow
framework — a thin orchestrator wrapper spawning the real workflow — rather than
changing the chassis. Built as
`docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_council_gate_orchestrator.sql`
(applied; the `council-gate-orchestrator` row is live) and exercised end-to-end.

**This supersedes the "give councils N lanes" idea. Do not build that.** With a
wrapper the lane is held only for spawn+call, so N councils run concurrently
through ONE lane — no chassis roll, no static N, DB config only.

### What the test PROVED (the lane problem is solved)

Submission `f5da8f65-a3ec-4d16-8254-3dbfcb76953c`, published 19:50:05 UTC:

- the wrapper reached `AWAITING_RESPONSES` **24 s** after publish;
- a dedicated pod `agent-council-gate-31da8db4-ktgmj` ran the council;
- **`QUEUE DEPTH (LAG) : 0` on `system.agent.generic.requests` while that pod was
  running.** This is the decisive measurement and it is structural rather than a
  race-test: the offset is committed only *after* `processMessage` returns
  (`agent.go:672-675`), so under the old inline path LAG is necessarily >= 1 for
  the whole 4-9 minutes. LAG 0 means the consumer is back in `FetchMessage` with
  nothing behind it.

### What the test REFUTED (do not repeat my mistake)

I described the wrapper pattern as "proven", on the strength of the archetype
being in daily use and of feature-builder's merged PR. **Measured instead, over
the whole retained history of `orchestration_states` (oldest row 2026-07-13):**

```
diagnose-orchestrator | COMPLETED |  2 | avg  527s
diagnose-orchestrator | FAILED    |  2 | avg 2796s   <- both "timed out after 3 retries" at call_diagnoser
```

**Two of four.** The archetype this candidate copies fails about half the time,
and my run failed too — stuck at `spawn_council` / `AWAITING_RESPONSES`, never
reaching `call_council`. "In daily use" is not a reliability measurement, and the
query that settles it takes one minute. Logged in `WRONG_CALLS.md`.

### The race, with ms evidence — `[INFERRED]`, not diagnosed

```
19:50:11.113  parent  spawn_council -> EXECUTING_STEP     (spawn action begins)
19:50:24.189  child   sends its initialisation response
19:50:26.109  parent  spawn_council -> AWAITING_RESPONSES (parent starts listening)
```

**The child answered 1.92 s before the parent began listening.** `SpawnAgentAction`
pre-registers the awaited request (`spawn_actions.go:104-126`), then sleeps 5 s,
sends init, sleeps 5 s more (`:130-151`), and only then returns a result carrying
`await_response: true` (`buildSpawnResult`, `:565`) — after which the coordinator
persists `AWAITING_RESPONSES`. A child that boots fast enough to reply inside that
10 s window replies into a window the parent is not yet in.

Consistent with this, the archetype's successful runs never show `spawn_*` in
`AWAITING_RESPONSES` at all — `spawn_diagnoser -> call_diagnoser` in 13.7 s and
12.0 s, no awaiting state recorded — i.e. their responses were applied while the
step was still `EXECUTING_STEP`. `council-gate` boots in ~3 s and appears to lose
the race consistently where `diagnose-agent` loses it sometimes.

**This is a theory with good timing evidence, NOT a diagnosis.** It is a durable
claim about shared spawn infrastructure, the cause may not be where the symptom
is, and this thread was already wrong once today about this same subsystem — so
it belongs in the diagnosis loop before anyone acts on it. It is very likely the
same family as `bugs_open/003` (spawn lost child response) and `029` (what those
hangs do to the fleet); read both before filing anything new.

### State left behind — safe, and deliberately not the default

- `council-gate-orchestrator` **is live in `agent_definitions`, but nothing
  targets it.** `097_TRIGGER_council_review_v1.sh` still defaults to
  `council-gate`; the wrapper is reachable only via an explicit
  `TARGET_AGENT_TYPE=council-gate-orchestrator`. **Do not flip that default until
  the spawn race is fixed** — it would move every session onto a path that fails
  about half the time.
- `council-gate.idle_timeout_seconds` 0 -> 900. Independent small fix (0 falls
  through to the 3600 s default at `spawn_actions.go:2691`, so a finished council
  pod would loiter for an hour). Inert while nothing spawns council-gate.
- The test run was cancelled and its Job reaped, so it is not sitting in
  `AWAITING_RESPONSES` feeding the `029` saturation class.

### What is actually blocking, restated

The lane fix is correct and measured. **The blocker is one level down: the
spawn→call handshake.** Fixing that unblocks this candidate, the diagnosis loop's
own reliability (half its runs die the same way), and the feature-builder chain —
so it is worth more than this bug alone.
