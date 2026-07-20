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
