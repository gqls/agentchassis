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
