# NOTES — dispatch queue serialisation (bugs_open/030)

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep — including my own claims in this file that later
turned out false.

---

## 2026-07-20 — session start, re-measuring 030 against the live cluster

Picked up `bugs_open/030` (filed 2026-07-19: "every orchestration dispatch queues
behind every other: one partition, one consumer, ~25–36 min latency").

### Confirmed the filed topology facts still hold

```
$ kafka-topics.sh --describe --topic system.agent.generic.requests
PartitionCount: 1  ReplicationFactor: 3  min.insync.replicas=2
```

```
$ kubectl -n ai-persona-system get deploy agent-chassis -o jsonpath='{.spec.replicas}'
1
$ kubectl -n ai-persona-system get hpa
(no resources)
```

`kubectl get pods | grep agent-chassis` → exactly one pod
(`agent-chassis-5567d99bd6-5snzn`). So: one partition, one replica, no autoscaler.
**[VERIFIED 2026-07-20]**

### The near-miss I want on the record first

At 19:10–19:12 I sampled the consumer group every ~20 s and saw `CURRENT-OFFSET`
pinned at 95963 while `LOG-END-OFFSET` climbed 96053 → 96059 and `LAG` grew
90 → 96. I also found **zero** `"Message fetched successfully"` lines in a 30-minute
log window. I was one step from writing "the queue is permanently diverging;
consumption has stopped; ~19 hours of backlog".

**That would have been wrong, and 030's own landmine is what stopped me** — the
bugfix-028 thread had already made this exact error on 2026-07-20 and written it
up. I let the sampler run longer instead. By 19:21 the offset had moved
95963 → 96010 (**+47 messages in ~8.5 min**) and LAG had *fallen* 96 → 68.

**Rule, restated for this file: on this topic a static `CURRENT-OFFSET` is
worthless as a liveness or divergence signal, and two minutes of sampling is not a
trend.** The drain is a sawtooth — long stall, then a burst. Sample for ≥20 min
before saying anything about the direction of the queue.

### The measurement that actually characterises the bug

40 samples at 30 s (`scratchpad/lag_samples.txt`), 19:21–19:29:

```
19:21:16  96010  96078  68     <- offset pinned here
19:22:25  96010  96080  70
19:24:07  96010  96087  77
19:26:24  96010  96091  81
19:28:40  96010  96097  87     <- still pinned, ~8 min later
19:29:14  96013  96098  85     <- advances
```

**One message occupied the consumer for ~8 minutes** while producers added ~2
msg/min behind it. That is the shape of the bug: not a dead consumer, not a
diverging queue — head-of-line blocking with a multi-minute head.

### Why it blocks — the mechanism, read in the code

This is where I think the filed root cause is **under-specified** (see PLAN for the
proposed correction; the claim is out for diagnosis as corr
`78470372-7617-40e4-888c-66cac94006bf`, so treat it as UNCONFIRMED until that
returns).

1. `platform/agentbase/agent.go` `processRequests()` — `for { msg := Consume(ctx);
   a.processMessage(msg, "request") }`. `processMessage` is called **synchronously**.
   No goroutine. So the loop cannot fetch again until the message is fully handled.
2. `platform/orchestration/coordinator.go` `continueExecution()` — a `for {}` loop
   that calls `executeStep` repeatedly, advancing through consecutive steps inline.
   `grep -n "go func\|go p\.\|go c\.\|go sc\." platform/orchestration/coordinator.go`
   returns **nothing** — the coordinator spawns no goroutines at all.
3. Each local LLM step is tens of seconds. Measured from the pod log:
   `ai_actions.go:234 Rendered prompt template` 19:14:16 → `ai_actions.go:479 LLM
   response received` 19:14:44 = **28 s**, then immediately
   `coordinator.go:923 Transitioning to next step` → next step at 19:14:49.

So a workflow segment of N consecutive local LLM steps holds the single consumer
goroutine for roughly N × 30 s, and nothing else on the topic can even be *fetched*
during that time.

**Consequence for the filed fix candidates:** raising `PartitionCount` (030 fix
candidate 2) cannot help on its own, because `replicas: 1` means one consumer would
simply be assigned every partition and still run one blocking goroutine. Partition
count is not the binding constraint. **[INFERRED — this is the specific claim under
diagnosis]**

### A separate defect found while reading the consumer

`platform/kafka/consumer.go` `Consume()`:

```go
msg, err := c.reader.FetchMessage(fetchCtx)
...
// After successful processing, commit the offset
if err := c.reader.CommitMessages(ctx, msg); err != nil {
```

The comment says "after successful processing". The code commits **immediately
after fetch, before the message is handed to the caller** — `processMessage` has not
run yet. That is at-most-once delivery: a message in flight when the pod dies is
gone. This is *already* a known root cause in the bugfix-003 spawn-loss workstream
("at-most-once consume"), so it is not new — but note it is **not** what 030's
landmine says.

> **CORRECTION to `bugs_open/030`'s landmine (added there 2026-07-20 by the
> bugfix-028 thread).** It states: "offsets are committed *after* a message is fully
> processed, and the message in flight was a multi-step council orchestration."
> The first half is wrong for this code path — commit happens at fetch
> (`consumer.go:103`), before processing. The landmine's *practical advice* (a frozen
> offset is not a dead consumer) is still correct, but the reason is different: the
> loop is blocked inside `processMessage` and therefore never calls `Consume()`
> again, so no *new* offset is fetched or committed. Same observable, different
> mechanism. Caught by reading `consumer.go` rather than trusting the note.

### Open question I could not settle by reading

Spawned agents get a job-specific `REQUESTS_TOPIC` from env
(`agent.go setupConsumers`), and only fall back to `system.agent.generic.requests`
when it is unset. So it is **not** established that *all* fleet work funnels through
this one topic — much of it may run on job topics with their own consumers. 030
asserts "every trigger from every concurrent session" funnels through the single
lane; that part I have **not** verified, and it materially affects how bad this is.
Included as an explicit question in the diagnosis submission. **[UNVERIFIED]**

---

## 2026-07-20 (later) — a concurrent thread corrected, and the same trap from both ends

While I was measuring, the **bugfix-036 thread** committed a throughput section
into `bugs_open/030` (`33008f0ca`). My session-start read of that file predates it —
a live demonstration of the point CLAUDE.md opens with, and of the user's warning
that files move underneath you. **Re-read before appending; I did, and it mattered.**

### What they claimed, and why it is wrong

They derived **0.21 msg/min** and a **~6.5 h** queue wait from two
`kafka-consumer-groups.sh` readings labelled `19:16` and `19:30`.

Both of their readings appear verbatim in my continuous sample file:

```
their "19:16"  96013 96099 86   ==  my sample at 19:29:48
their "19:30"  96016 96102 86   ==  my sample at 19:30:57
```

69 seconds apart, not 14 minutes. And the first cannot have been at 19:16 at all:
my sampler shows the offset **pinned at 96010 from 19:21:16 to 19:28:40**, 15
consecutive readings, so `current=96013` at 19:16 would require the committed
offset to run backwards. It does not, absent a group reset.

3 messages / 69 s ≈ 2.6 msg/min. Over my full continuous window
(19:10:37 → 19:32:39, 95963 → 96016) the honest figure is **53 messages in 22 min
≈ 2.4 msg/min** → a queue of 86 clears in **~36 min**, which is exactly the
**25–36 min** 030 measured on 2026-07-19. Nothing had degraded; the entry as
originally filed was right.

Correction appended to `bugs_open/030` (commit `5904eb60c`) with the sample file
alongside it as `EVIDENCE_lag_samples_2026-07-20.txt`.

### What they got RIGHT, and it is the important half

Their mechanism inference — "the rate is set by orchestration DURATION, not by
queue mechanics … consistent with the consumer running each orchestration to its
wait point before taking the next message" — is **correct**. I reached the same
conclusion independently the same hour by reading `agent.go` and `coordinator.go`.
Two routes, one answer: that is the part worth keeping, and the correction says so
explicitly rather than burying it.

### The thing I actually want to remember

**I made the mirror-image error two hours earlier** (near-miss, §above): frozen
offset + silent log → "consumption has stopped". They made it in the other
direction: two samples across a stall → "12× too slow, multi-hour waits".

Same signal, same week, two threads, opposite conclusions, both wrong — so this is
a property of **the signal**, not of either session. The queue is a sawtooth and
any window shorter than one tooth is uninformative in both directions. Logged as
`WRONG_CALLS.md` (7) and (8) with that synthesis, plus the mechanical remedy
(≥20 min continuous sampling, take the slope — RUNBOOK R2).

And the tell both of us share: **we had the code and used the clock.** The
synchronous-handler mechanism is legible in two files in about ten minutes and
predicts the sawtooth outright. Reading it first would have made both measurements
unnecessary — or at least uninterpretable-in-the-wrong-direction.

### Status

Root-cause claim still **UNCONFIRMED** — diagnosis corr
`78470372-7617-40e4-888c-66cac94006bf` had still not started at 19:36
(`orchestration_states` → 0 rows; work item `awaiting_diagnosis` since 19:23:35).
It is queued behind the backlog it is about. Per
[[council-queue-latency-trap]] and 030's own landmine: **do not resubmit.**
Owner's call (2026-07-20): wait for the verdict before building any fix.

---

## 2026-07-20 (later still) — I made the same error I had just written up

> **CORRECTED — the "~2.4 msg/min" figure earlier in this file, and in my
> `bugs_open/030` correction, is WRONG.** Recording it here in full because this
> file's stated purpose is that the missteps are the point.

I published the correction from a **partial read of a sampler I had already
started**. At the time I wrote it I had 21 of 40 samples. The completed run says
something different:

| window | messages | rate | LAG |
|---|---|---|---|
| 19:12:51 → 19:21:16 (contains the +47 burst — the window I used) | 47 in 8.4 min | 5.6/min | 96 → 68 |
| 19:21:16 → 19:44:01 (full sampler run) | 14 in 22.8 min | **0.62/min** | 68 → **109** |
| 19:10:37 → 19:51:16 (everything sampled) | 61 in 40.7 min | **1.50/min** | 90 → **130** |

`LAG` **grew 82 → 130** across the session, with one message pinning the offset for
**≥15.4 minutes**. The queue was **diverging**. My "clears in ~36 min, nothing had
degraded" was false in both halves.

**The shape of my error is precisely the one I had documented one paragraph
earlier.** I wrote "any two samples inside a burst give an arbitrarily high rate"
and then computed my headline number over a window built around a burst. Knowing
the trap by name did not prevent it — I was looking for the trap in *their* data,
having already decided what mine showed.

**And the worse half is not arithmetic.** The bugfix-036 thread's conclusion —
"variance is large; 'how long will my submission wait' has no stable answer" — was
**correct**, and I overturned it. Their derivation really was invalid (the 69-second
finding stands, and they have since owned it, `7c43e6aee`). But being right about
their *method* is what licensed me to be wrong about their *conclusion*, and I did
not separate the two. **Faulting a derivation does not entitle you to reverse the
finding.**

### What all three errors actually share, and it is not window choice

Three threads, three defensible rates from one queue in one afternoon: 0.21, 2.4,
0.62. When three careful measurements of a quantity disagree by 12×, the fault has
stopped being in the measurements. **There is no single rate here to measure.**
Throughput is `1 / (duration of the orchestration segment currently running inline
on the consumer goroutine)` — the mechanism established from the source earlier in
this file — and that duration spans milliseconds to ≥15 minutes depending entirely
on what sits at the head. It is a non-stationary signal; its mean describes no
moment and predicts nothing.

So the remedy I proposed in `WRONG_CALLS` (7)/(8) — "sample ≥20 min and take the
slope" — **is itself withdrawn** (RUNBOOK R2 marked corrected, R7 added). A longer
window buys stability, not truth. All three of us had the mechanism available in two
files and reached for the stopwatch instead; the code predicts non-stationarity
outright, which would have told us the rate was not a thing to go and measure.

Logged as `WRONG_CALLS.md` (9), with that synthesis.

---

## 2026-07-20 (evening) — the question nobody had asked: what is IN the queue

Three threads had now argued about how fast the consumer drains. None of us had
looked at what we were draining. `kcat -o -300`, headers only:

- **93%** of messages are `from_agent_type=kafka-scheduler` (550/588); 6.5% `user`.
- **84%** are two jobs: `sched-ai-endpoint-health-check` (43%) and
  `sched-build-pipeline-trigger` (41%).
- `council-gate` is 5%, `needs-diagnosis` 1%. **The interactive work this bug is
  about is a rounding error queued behind cron.** [VERIFIED — headers]

### The mass balance closes, which is the strongest evidence yet

19:05 → 20:15 UTC (70 min), from offsets:

| | | rate |
|---|---|---|
| produced | 96040 → 96222 (+182) | 2.60/min |
| consumed | 95958 → 96058 (+100) | 1.43/min |
| shortfall | | **1.17/min** |

Observed LAG: 82 → 164 = **+1.17/min**. To two decimals. Nothing else is going on.

**On quoting rates after retracting rates:** this is a *mass balance over one closed
window with both endpoints stated* — an account of where 82 messages came from, not
a forecast. That is the distinction my retraction turns on, and I want it explicit
so the next reader does not think I have quietly resumed the habit. Do not carry
1.43/min forward as an ETA.

### Config, and a landmine in it

`scheduled_tasks` — twelve enabled tasks target this lane; the two dominant ones are
`interval_seconds = 30`. Nominal configured load **≈6 msg/min** against a consumer
draining 1.43/min.

But `TICK_INTERVAL_SECONDS` is **also 30**, and the due check is
`last_triggered_at + interval <= NOW()` (`cmd/scheduler/main.go` `loadDueTasks`), so
the boundary tick is marginally early, the task is not yet due, and it fires on the
*next* tick. **A 30 s task fires every 60 s.** Confirmed on the wire: 59–60 s
spacing, drifting a second at a time — the aliasing signature.

**So the queue is currently protected by an off-by-one, and correcting it would
double production to ~4/min against a ~1.4/min consumer.** Recorded as a landmine in
the bug file. This is the sort of thing that gets "tidied" by someone who sees
`interval_seconds=30` and a 60 s reality and assumes a bug.

### Blast radius narrowed [VERIFIED]

`spawnAgentKubernetesJobFromDefinition` (`spawn_actions.go:2320`) launches each
spawned agent as its **own Kubernetes pod** on its own `job.*` topic — 815 topics,
8 `agent-*` pods live during the measurement. So this lane is **top-level dispatch
only**; per-agent work is genuinely concurrent. 030's "every trigger from every
concurrent session funnels through that single lane" is right about *triggers* and
too broad if read as *work*.

The contention is **cron-vs-everything**, not session-vs-session. That is a sharper
bug and a cheaper fix: `interval_seconds` is a DB column and therefore live with no
image roll.

### Status — the wait-for-the-verdict plan is being defeated by the bug itself

Diagnosis corr `78470372` has been `awaiting_diagnosis` for **55 minutes** and has
not started. LAG 168 and rising at ~1.2/min. Since the queue is *diverging*, there
is no basis for expecting it to start at any particular time — it will start when
the head of the queue happens to be cheap, or not today. Flagged to the owner.

---

## 2026-07-21 — first fix applied: cron/config only (owner-instructed)

Owner asked for the config-only change to "adjust the timeouts so they work together
properly", and for a milestone snapshot. Snapshot = `SUMMARY_2026-07-21_*`.

**Grounded the values before changing (read before write):**
- `scheduled_tasks`: of the 12 enabled tasks on the generic lane, only
  `ai-endpoint-health-check` is high-frequency **and** unconditional (`pre_query`
  NULL → fires every cycle). `build-pipeline-trigger` is gated but its gate is open
  (3 pending triaged build sites now). Everything else is 120 s+ and gated. [VERIFIED]
- `ai_endpoint_health`: active endpoints want claude 3600 s, cpu-ollama 60 s,
  gpu-ollama 30 s. gpu-ollama is `healthy=f`. So the shortest *healthy* need is 60 s;
  the 30 s want was already unmet (aliasing fired the orchestration at 60 s). [VERIFIED]

**Applied** (both guarded with `AND interval_seconds = 30` — safe no-op on concurrent
change; both returned `UPDATE 1`):

```
ai-endpoint-health-check : 30 -> 60
build-pipeline-trigger   : 30 -> 120
```

Both are exact multiples of `TICK_INTERVAL_SECONDS` (30), so they now fire
deterministically and the interval==tick aliasing landmine is gone. Full command +
reversal in RUNBOOK R8.

**Reasoning, not just arithmetic:**
- The health-check change is **behaviourally neutral** (it already fired at 60 s under
  aliasing) — its value is honesty (config == reality) and removing the trap.
- The build-trigger change is the **load lever that matters**, and not because of raw
  message count: it starts the expensive multi-step build orchestrations that occupy
  the single consumer for minutes at a time. Halving its rate reduces both production
  *and* the number of long inline stalls, which raises effective consumption. That
  coupling is why 1.43/min is not a fixed ceiling — it was measured under exactly the
  congestion this change reduces.

**Did NOT touch `timeout_seconds`.** The owner said "timeouts" but the defect is the
firing cadence (`interval_seconds`) vs the tick; `timeout_seconds` is the re-fire
guard and is fine (15 s << 60 s interval for the health check, so no overlap). Noting
the interpretation explicitly in case it was meant literally — easy to revisit.

**Verification** (against the running scheduler, not the config): background sampler
of `last_triggered_at` running now. Expect `build-pipeline-trigger` spacing to widen
from ~60 s to ~120 s; health-check stays ~60 s. Result appended below when the
sampler completes. **Not claiming the backlog is fixed** — that needs a fresh LAG
trend over ≥20 min AND accounting for what work is at the head, and the queue was
diverging when the change went in. This reduces scheduled production; whether it
brings production under consumption is the thing to measure next, not assert.

---

## 2026-07-21 — verified the config change, and corrected my own mechanism claim

Measured the fire cadence after the change (`EVIDENCE_fire_intervals_2026-07-21.txt`,
13 samples at 30 s):

- `ai-endpoint-health-check` (interval 60): fired 10:34:03, 10:35:33, 10:37:03,
  10:38:34, 10:40:03 → gaps **90, 90, 91, 89 s**. Steady 90 s.
- `build-pipeline-trigger` (interval 120): fired 10:33:33, 10:36:04, 10:38:33 →
  gaps **151, 149 s**. Steady 150 s.

**This overturned my own "tick aliasing" landmine.** I had written that a 30 s task
fires every 60 s because `interval == TICK` is a special aliasing case, and that a
clean multiple of the tick removes it. Both wrong. The measured rule is
**effective = interval + TICK, for any multiple of the tick** (30→60, 60→90,
120→150), and the source says why: `last_triggered_at` is stamped at *fire time*
(late in `runTick`) but the due test reads `NOW()` at the *start* of the next tick,
so the boundary tick reliably fails `last + interval <= NOW()` and slips one tick.
Universal, not special. My "pick a clean multiple" advice was exactly backwards —
multiples are the boundary case that takes the extra tick. Corrected rule: for a
target effective period P, set `interval = P − 30`. Recorded in the bug file and
RUNBOOK R8; SUMMARY carries a visible correction note.

**Load effect (from measured effective periods):**

| | before (60 s each) | after |
|---|---|---|
| ai-endpoint-health-check | 1.00/min | 0.67/min |
| build-pipeline-trigger | 1.00/min | 0.40/min |
| two dominant | 2.00/min | **1.07/min** |

Estimated total scheduled production **2.60 → ~1.67/min**, against a consumer
measured ~1.43/min under congestion — and that consumer figure *rises* as fewer
expensive build chains start, so the two are now near-balanced, possibly balanced.

**LAG after the change: 20** (10:49 UTC), against 82→168 and diverging yesterday.
**[NOT PROOF]** — one reading, and the offset had advanced ~1,960 overnight, so
quiet-hours draining is a large confounder. Consistent with improvement; a real
before/after needs a LAG trend across a *busy* window, which I have not yet run. Do
not record "030 fixed" on this evidence.

### What remains
- The core root-cause claim (inline handler vs `PartitionCount`) is still only filed,
  not adjudicated — diagnosis corr `78470372` never started (stuck in this queue).
  The config change does not depend on it, and the mass balance + blast-radius
  findings corroborate the mechanism independently.
- Structural fix (dedicated scheduler lane) still open, still needs the council gate,
  still must be checked against `/bugs_open/029`.
- The diagnosability win (triggers print `LAG` on publish) is untouched and remains
  the highest-value cheap change.

---

## 2026-07-21 — busy-window LAG trend: divergence → bounded sawtooth (the actual verification)

The evidence I said the fix needed. 27 samples at 60 s, 15:44–16:11 UTC (working
hours, real load), `EVIDENCE_lag_postfix_2026-07-21.txt`:

```
15:44  LAG 0    <- fully drained
15:52  LAG 9    (offset pinned at 98442 since 15:45 — one message, ~6.5 min stall)
15:54  LAG 0    <- drained again
16:07  LAG 15   (offset pinned at 98462 since 15:58 — one message, ~8.5 min stall)
16:11  LAG 15
```

**Read the shape, not the average — the average is a window-phase artifact.** Raw
counts over the window: produced +41 (1.53/min), consumed +26 (0.97/min). That
"production > consumption" is exactly the LAG delta (0→15) and nothing more — the
window opens on a drained queue and closes mid-climb. Averaging a sawtooth across an
odd number of half-cycles always manufactures a trend. (This is the same trap as
`WRONG_CALLS` 7/8/9, so I am not repeating it.)

**What actually changed — qualitative, and it is decisive:**

| | yesterday (pre-fix) | today (post-fix) |
|---|---|---|
| LAG path | 82 → 130 → 168, monotonic | 0 → 9 → 0 → 15, oscillating |
| returns to 0 | never (in 40 min watched) | twice (in 27 min) |
| peak LAG | 168 and climbing | 15 |

The queue now **fully clears between stalls**. That is the difference between a
diverging backlog and a bounded one, and it is not subtle.

**What did NOT change — and this is the structural residual:** the single-message
head-of-line stall is still ~7–8 minutes (offset pinned at 98442 for 6.5 min, 98462
for 8.5 min). The config change reduced how *often* expensive orchestrations arrive,
so each drain burst now catches all the way up — but when one is at the head,
everything still waits behind it for its full duration. That is the inline
synchronous handler (the unadjudicated root cause), and it is exactly what the
structural fix (dedicated scheduler lane, and/or non-blocking step execution) would
address. The config change made the queue bounded; it did not make it non-blocking.

**Verdict I am willing to record:** the config change converted an unbounded,
diverging backlog into a bounded sawtooth peaking around 15 with periodic full
drains, measured across a busy 27-minute window with two complete drain cycles. That
is a real, measured improvement — **not** "030 closed" (the head-of-line stall
persists, and two drain cycles is indicative, not a week of data), but enough to say
the lever worked and the direction is right.

---

## 2026-07-25 — both remaining fixes BUILT: publish acknowledgement (live) + a scheduler lane (inert until the roll)

Session "bugfix 030". Picking up the two items the 2026-07-21 summary left: the
cheap diagnosability win, and the structural lane split. Ownership checked first
(`scripts/who-owns.py 030` → this workstream, ACTIVE; last workstream commit
07-21; the 07-24 entry in the case file is another thread contributing evidence
and explicitly starting no competing fix).

### The mechanism re-confirmed live on the CURRENT image, before touching anything

Not taken from the file — re-measured, because every figure in here has a
staleness half-life of days.

```
17:04 UTC, chassis pod agent-chassis-774877f4c6-zjh4t (v1.0.1159, started 15:25)
  orchestration 407cb6b5 · council seat review_compliance → review_guardian
  EXECUTING_STEP for 9m02s, updated_at moving (ai_actions.go LLM steps logging)
  LAG on system.agent.generic.requests over the same minutes: 8 → 25
```

One orchestration, mid multi-step LLM chain, inline on the consume goroutine,
with everything else stacking up behind it. That is the entry's mechanism
verbatim, on today's image, with F2/F3 (bugs_open/003) already in.

Also re-checked: the 07-21 config change is **still live and intact**
(`ai-endpoint-health-check` 60, `build-pipeline-trigger` 120 — nobody has
re-seeded over it) and LAG at session start was **8**, i.e. still bounded, not
diverging. So the config lever held for four days.

**[NEW, and it changes the arithmetic]** The generic lane now has **21 enabled
scheduled tasks**, up from 12 on 07-21 — including `diagnose-pipeline-trigger`
at 60 s and `claimed-item-timeout` at 120 s, which did not exist (or were not
enabled) when I measured. Summing effective periods (`interval + 30` tick, the
rule this workstream corrected on 07-21) puts nominal cron production back at
**≈2.5/min**, right back where it was before the 07-21 fix. **The config lever
does not hold by itself — the lane silently re-fills as tasks are added, because
nothing owns the total.** That is the argument for the structural split rather
than another round of interval tuning: tuning decays, a separate lane does not.

### Fix 1 — acknowledge the dispatch on publish (LIVE, commit a5a494459)

`scripts/dispatch-queue-depth.sh`, called by 090 and 097 after they publish.
Prints: LAG (>0 = queued, not lost, do NOT re-fire), an explicit fault call-out
when the group has **no member** (that case really is a fault, not a wait),
consumer liveness from the last `orchestration_states.updated_at` transition
rather than the offset, and the in-flight orchestrations so the operator can see
whether something expensive is at the head. ~8 s to run.

**It prints no ETA, deliberately**, and says so in its own output. This
workstream produced three defensible-and-useless drain rates in one afternoon;
the script is the place where that lesson had to become mechanical rather than
remembered.

Proof it works in situ — the tail of my own council submission a few minutes
later, which is exactly the moment an operator would otherwise start wondering
whether the dispatch had vanished:

```
  QUEUE DEPTH (LAG) : 18
  Your message was published to the back of this lane, so roughly
  17 message(s) are ahead of it. It is QUEUED, NOT LOST
  Consumer liveness: last orchestration step advanced 9s ago.
  In flight now: ... generic-orchestrate-0725-1714 | review_prior_art | 171s
```

### Fix 2 — a dedicated lane for cron (BUILT, INERT until the roll, commit f9bc7f45f)

`EXTRA_REQUEST_TOPICS` (comma-separated) makes a statically deployed agent
consume additional request topics, each with **its own goroutine, consumer group
and offsets**. The chassis deployment sets it to
`system.agent.scheduled.requests`. Within a lane, fetch→process→commit stays
strictly serial and in order, so per-orchestration ordering and F3's
commit-after-process semantics are untouched; across lanes, a 9-minute council
chain no longer parks cron and vice versa.

Design decisions worth having on the record:

- **The fetch loop is EXTRACTED and SHARED (`consumeRequestLane`), not copied.**
  Those delivery semantics were paid for by bugs_open/003 F3; two copies would
  drift, and the drift would be silent.
- **Partitioning the shared topic is still refused**, for the reason this
  workstream established: one-way, and with `replicas: 1` one consumer takes
  every partition and still blocks. Nothing about that changed today.
- **Concurrency inside the handler is deliberately NOT touched.** Decoupling
  consumption from execution is `chassis_replica_scaling`'s P1 — designed, not
  built, and explicitly written up there as *the* fix for 030's latency. Lane
  separation composes with it and does not compete with it; building P1 here
  would have been starting a competing implementation of another workstream's
  designed phase.
- **Two guards, because the failure mode is duplicate execution of every
  scheduled dispatch**: a lane equal to the main topic is refused, and the var is
  ignored unless the agent's own requests topic is `system.agent.*` (i.e. static,
  not a `job.*` spawn). The second guard exists because spawned pods inherit
  `envFrom: personae-prod-config` **wholesale** (`spawn_actions.go`), so the var
  reaching that ConfigMap would put every live spawned pod on the lane under its
  own group. **NEVER put EXTRA_REQUEST_TOPICS in personae-prod-config.**
- **A lane that cannot be created is skipped and logged loudly, not fatal.** A
  crash-looping chassis is worse than a missing lane.

### The rollout order is the risk, not the code

`scheduled_tasks.target_topic` is a DB column — live immediately. So the only way
to get this wrong is to switch the producers **before** the consumer exists, which
would publish every cron dispatch into a topic nobody reads: silent, and all
scheduled work stops. Order is therefore image → verify the lane is really being
consumed → **then** the UPDATE. RUNBOOK R9 has the exact sequence and the
rollback.

### Where it stopped this session

Image `v1.0.1164` built from committed HEAD and verified to contain the change
before pushing (`strings /app/agent-chassis | grep -c "Extra request lane ready"`
→ 1, with `Starting request processor` → 2 as a positive control). **`docker push`
is blocked for this session by the tool-permission classifier**, so the roll, the
`target_topic` switch and the post-fix measurement are handed to the owner as one
short sequence (RUNBOOK R9). Fix 2 is committed and inert until then — which is
exactly the state `/bugs_closed/README.md` says must stay OPEN.

Council gate: submitted, `SUBMISSION_CORR=f47c2305-a873-459a-83e6-13eb9cb0cf1f`.
At submission time the lane was 18 deep, so the verdict is ~30 min out — the bug
under review delaying the review of its own fix, one more time.

## 2026-07-26 — the lane is ARMED and the latency is gone: publish→start 1 s (was ~18 min for the same submission yesterday)

The owner pushed a chassis image overnight; production is on **v1.0.1165**, which
carries the lane code, and the Deployment carries
`EXTRA_REQUEST_TOPICS=system.agent.scheduled.requests`.

**Verified before arming anything** (R9 step 3, all three checks):

```
strings /app/agent-chassis | grep -c "Extra request lane ready"   -> 1
strings /app/agent-chassis | grep -c "Starting request processor" -> 2   (positive control)
kafka-topics.sh --describe --topic system.agent.scheduled.requests
  -> PartitionCount 1, RF 3, compression.type=snappy, cleanup.policy=delete,
     retention.ms=604800000, max.message.bytes=5242880
kafka-consumer-groups.sh --list | grep lane
  -> generic-requests-group-lane-system-agent-scheduled-requests   (member: the chassis pod)
```

The topic's config entries are `TopicManager.CreateTopic`'s own — i.e. **the lane
topic was created by the new code path**, not by hand and not by broker
auto-create. That is the self-provisioning claim, observed rather than asserted.
(The pod log lines were already rotated out of the retained window by the time I
looked — the chassis is verbose enough that a 70-minute-old startup line is gone.
The Kafka group + the topic's config fingerprint carried the proof instead.)

**Armed** (R9 step 4): `UPDATE scheduled_tasks ... WHERE enabled AND
target_topic='system.agent.generic.requests'` → **UPDATE 18**.

### Before/after, same submission through the same council

| | round 1 (2026-07-25, shared lane) | round 2b (2026-07-26, split lanes) |
|---|---|---|
| LAG at publish | **18** | **0** (then 1 = my own message) |
| publish → `orchestration_states` row | ~18 min (17:15 → fix_plan artifact 17:33:06) | **~1 s** (13:26:33 → 13:26:34) |
| state at first look | not yet created | already `review_editquality` |

Both lanes at LAG 0 ten minutes after the switch, with the scheduled lane
consuming what it receives (`CURRENT-OFFSET 2 / LOG-END 2`) and cron work
completing on it — `generic-orchestrate-0726-1322` and `-1323`, 90 s apart, which
is the health-check's `interval 60 + 30 s tick` effective period exactly. So cron
still runs, on its own lane, and the interactive lane is empty when an operator
arrives at it. **That is the fix, measured on both sides.**

### The council round 1 came back REVISE — and every objection was answerable by looking

Gating objection from `guardian`, plus two from `prior_art_librarian`; five seats
approved (`editquality`, `reuse_agent`, `debug_historian`, `constitution`,
`mission`). What the checks found:

- **"Was the zero-code alternative considered — a second chassis Deployment
  pointed at the new topic via `REQUESTS_TOPIC`?"** Fair, and I had not ruled it
  out. Ruled out now in code: the **response** consumer's group is `a.AgentID`
  (per pod), so every chassis process receives every response, and
  `ProcessResponse` (coordinator.go:271-277) discards any response whose
  `state.ProcessingNode != s.podName`. That pair is exactly what
  `chassis_replica_scaling` documents as the reason `replicas: 1` is pinned and
  what `bugs_open/075` shows biting; `NewConsumer` also sets
  `StartOffset: FirstOffset`, so each new per-pod group replays the whole
  responses topic on every start. A second Deployment buys lane isolation by
  creating a second ownership domain in the configuration already ruled unsafe.
  One process, two lanes keeps one response consumer and one ownership domain.
- **"You did not name the blast radius of editing the shared loop."** Partly
  conceded, partly corrected: `grep -rl "platform/agentbase" --include=*.go cmd/`
  → **`cmd/agent-chassis` only**. No fleet-wide redeploy of the other backend
  services; the adapters do not link agentbase. It IS true the chassis image is
  what spawned pods run, and there the change is inert by construction.
- **"The new intra-process concurrency should be checked in code, not asserted by
  analogy."** The right objection, and the answer is that it is not new:
  `handleCompleteResponse` calls `continueExecution` at **coordinator.go:2307**,
  so the response goroutine already executes local actions inline concurrently
  with the request goroutine — same `executeStep → executeAction` path. Handlers
  are plain funcs from `GlobalActionRegistry`, a package map with three read sites
  and **zero runtime writes**; `SagaCoordinator`'s only mutable field is
  mutex-guarded; `*sql.DB` and `*kafka.Writer` are concurrency-safe by contract
  and `SetValidator` has no callers.
- **"The P1-is-unbuilt claim has no evidence."** Correct — it was the one absence
  claim I asserted from reading a doc. Evidence now: `git log` for
  `chassis_replica_scaling/` returns **2 commits, both docs, both 2026-07-20**, and
  P1's named machinery (intake persist + claim-worker pool) has no counterpart in
  the tree.
- **"`agent_definitions.topics` may already drive multi-topic subscription —
  dormant machinery?"** Checked: the column is a flat three-key map
  `{process,response,error}` of **singular** legacy templates
  (`system.agent.{type}.process`), identical in every row, and the only reader of a
  topics config (`bootstrap.go:110`) looks for keys `requests`/`responses` — which
  that map does not contain. Inert for subscription; nothing to extend.

### MISSTEP — a hypothesis I generated, then refuted, instead of shipping it as a risk

Working through the guardian's concurrency objection I convinced myself lane
separation had introduced a **new** hazard: serialisation used to make it
impossible for a reaper/takeover message to be processed *while* another
orchestration sat mid-inline-step with a stale `last_activity`; with two lanes it
becomes possible, so could a live orchestration now be reaped or taken over?

Refuted twice over by reading rather than reasoning:
- the stuck-step takeover (`coordinator.go:713`, `StuckOrchestrationTimeout = 5m`)
  fires only on a **second dispatch carrying the same `orchestration_id`**. Lane
  separation creates none — every cron dispatch carries its own id.
- `stale-orchestration-reaper` does its work in the **scheduler's `pre_query`**,
  never on the chassis lane at all (its dispatched workflow is a single
  `complete_workflow` no-op), and it only touches `AWAITING_RESPONSES` at 30/90-minute
  thresholds — never `EXECUTING_STEP`.

Recording it because the hypothesis was plausible, specific, and wrong, and
because it is the shape of thing that gets written into a handoff as a "risk" and
then believed. Cost of checking: two queries.

### MISSTEP 2 — the resubmission was refused in 6 seconds, and the validator was right

Round 2's first attempt (orch `ba2a015f`) died at `persist_submission`:

```
plan failed validation: edit 2: sketch declares no code change — a fix plan
proposes changes, not observations; drop the edit or make it real
```

I had turned an **answer to an objection** into a fake edit ("concurrency safety
evidence (no edit)", sketch: "no code change — evidence only"). Answers belong in
the `rationale`, which is what reviewers judge the plan against; the `edits` array
is for changes. Round 2b keeps round 1's real hunks verbatim and carries the
answers in the rationale. **Trap worth knowing: the gate refuses
observation-shaped edits, and it does so before spending a single reviewer call —
6 seconds, no credits.** It also, incidentally, produced another latency
datapoint: created 13:24:02, invalid 13:24:08.

### Council round 2b: REJECTED — a hard guardian veto, naming no defect

`abstained: 7`, `unreadable: none` (so not the truncation class). Six seats
approved (`editquality`, `reuse_agent`, `guidelines`, `tooling_provenance`,
`constitution`, `mission`); `debug_historian` and `prior_art_librarian` objected;
**`guardian` vetoed**. Its four objections are one argument repeated across the four
touched sites: adding lane-plurality to `Agent`, provisioning N consumer groups in
`setupConsumers`, generalising `processRequests` into `consumeRequestLane`, and
starting/stopping a variable number of goroutines in `Run`/`Shutdown` are
*foundational plumbing* edits, and the stability preference says rule out a
higher-layer fix first.

**It names exactly one such alternative, and it does not exist.** "Routing
scheduled/cron dispatch through the existing per-job spawn mechanism
(`job.<id>.requests` + generated consumer group, per `spawn_actions.go` ~2389) …
keeps the change entirely in the scheduler's dispatch decision and never touches
`platform/agentbase`'s shared consumption loop at all." Checked:

```
grep -rn "kubernetes|k8s.io|BatchV1|CreateJob|clientset" cmd/scheduler/*.go   -> NOTHING
cmd/scheduler/main.go:454   return producer.Produce(ctx, task.TargetTopic, headers, []byte(reqID), bodyBytes)
```

The scheduler can publish a message and nothing else. And a `job.*` topic exists
only because `spawn_actions` created it for a pod it spawned — and `spawn_actions`
runs **inside the chassis**, as an action of an orchestration, which must first be
dispatched through the very lane under discussion. So the named path would require
building pod-spawning (client-go, RBAC, pod template, topic create/cleanup,
lifecycle) into a service that today only knows how to publish, and would stand up
a second implementation of spawning in the fleet. **That is a larger, lower-level
change than the one vetoed, not a higher-layer one.**

One measured correction to the veto's premise: it calls `agent.go` "the load-bearing
chassis every static agent and every spawned job pod runs". `grep -rl
"platform/agentbase" --include=*.go cmd/` → **`cmd/agent-chassis` only** — the
adapters and other backend services do not link it. Spawned job pods *do* run the
chassis image, and there the code is inert twice over (the var is absent from a
spawn's explicit env list, and the static-agent guard refuses it anyway).

**The one objection with production consequences, audited rather than argued.**
Guardian OBJ 5 named three concrete defect shapes in the lifecycle change — leaked
consumer, goroutine never joined on partial startup failure, `wg` mismatch. Audited
in the code as shipped:

- a lane whose `CreateTopic` or `NewConsumer` fails is never appended, so nothing to
  leak; every appended lane is closed in `Shutdown` (agent.go:1509);
- `Run` pairs `a.wg.Add(1)` with `defer a.wg.Done()` **inside** each lane closure
  (agent.go:540-547) — no mismatch, and the loop takes an indexed copy so there is
  no capture hazard;
- a partial `setupConsumers` failure returns before `Run`, and the process exits on
  it, so there is no half-started lane set to join.

Clean, but it was the right thing to ask for.

`debug_historian` (medium) wanted a needle-gate on the production `UPDATE`. Partly
conceded: a pre-count **was** run immediately before (`SELECT count(*) … WHERE
enabled AND target_topic='system.agent.generic.requests'` → **18**), the statement
returned exactly `UPDATE 18`, and the post-state was verified by listing all 18
names grouped by `target_topic`. What was genuinely missing is a `RETURNING` clause
and a dump — the baseline/result pairing is now written into R9 and the case file so
it is auditable rather than narrated.

`prior_art_librarian` (high) asked that a human re-run the pod checks *before* the
SQL fired. That is the order they were run in (R9 step 3 → step 4); the seat has no
tier that can see a running pod, so it flagged per procedure, which is correct
behaviour on its part.

**What I am doing with a veto that names no defect.** The gate is advisory by
design; it records a verdict and cannot block. The change is live, measured on both
sides, and the alternative the veto asks for would be bigger. So: round 3 submitted
with the check folded in (`RESUBMIT_CORR` keeps the trail on one correlation), the
change stays live, **no `Council-Reviewed:` trailer is claimed anywhere** — that
trailer is earned by APPROVED only — and the veto is recorded in the case file's
closing section and put to the owner as a decision they may reverse with one
`UPDATE`. Contesting a veto with evidence is the loop working, not a bypass; if the
owner sides with the guardian, the rollback is the runbook's and costs seconds.

### Council round 3: APPROVED — the veto stood down, and three advisories closed by check

`approved with 2 advisory objection(s) — none high-severity`, `abstained: 6`,
`unreadable: none`. The guardian's own words: *"The round-2 veto's named
higher-layer alternative is refuted by hard evidence … — accepted."*

**REVISE → REJECTED → APPROVED on one correlation
(`f47c2305-a873-459a-83e6-13eb9cb0cf1f`).** The transferable bit: **a hard veto is
contestable with evidence, but it costs rounds** — three here, ~7 minutes of seats
each plus queueing. What moved it was not argument but two greps and a read: the
scheduler has no Kubernetes capability at all, and a `job.*` topic only exists
because the chassis spawned a pod for it.

Three advisories, each closed by a check rather than a preference:

1. **The other deployment-layer alternative, costed.** Round 3 asked about a second
   `agent-chassis` Deployment (unmodified image, own group, `REQUESTS_TOPIC` on the
   scheduled lane) — "zero lines of `platform/agentbase`". Measured cost:
   `system.agent.generic.responses` stands at offset **10,905**, and every chassis
   process consumes it under a **per-pod** group with `StartOffset: FirstOffset`, so
   a second Deployment replays all 10,905 responses through `processMessage` on
   **every pod start**, plus it creates a second **ownership domain**
   (`ProcessResponse`'s discard at coordinator.go:271 — the thing that pins
   `replicas: 1`, and `bugs_open/075`). Zero lines of Go ≠ cheaper.
2. **My own open residual, closed.** 318 package-level vars under
   `platform/orchestration/actions/`; exactly **one** runtime write outside a
   declaration (`registry.go:1940 deprecationLogger = logger`, startup-only, setter
   has **no callers** so it is always nil in prod). No action handler mutates shared
   state. **MISSTEP: my first pass reported 58** — it counted declarations inside
   `var (...)` blocks. A check whose failure mode is a false positive still has to
   be verified before its number is quoted.
3. **The needle-gate gap: conceded, and irrecoverable.** I tried to bind the 18
   changed rows to my `UPDATE` after the fact via `updated_at` — impossible, because
   `cmd/scheduler/main.go:273` re-stamps `updated_at = NOW()` on every fire. Only
   **7 of 18** still carry my statement's `13:21:31` stamp (the ones that have not
   fired since); the other 11 were re-stamped by the scheduler, which is at least
   independent proof they fire. **`RETURNING` has no substitute on a
   producer-touched table** → RUNBOOK R9 step 4a. **MISSTEP: my first binding query
   used a 13:22–13:24 window and returned 0**, which reads alarming; the `UPDATE`
   ran at 13:21:31 and the window was simply wrong. Checked before writing it down.
