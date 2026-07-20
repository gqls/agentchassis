# PLAN — dispatch queue serialisation (bugs_open/030)

**Started:** 2026-07-20 · **Branch:** `085_debug_and_feature_loops`
**Bug:** `bugs_open/030_HANDOFF_2026-07-19_orchestration_dispatch_queue_single_consumer_backlog.md`

---

## What this workstream owns

Why a dispatch fired at the cluster takes tens of minutes to start, and why that
delay is indistinguishable from a drop. 030 filed the symptom and the topology on
2026-07-19; this workstream is about establishing the **binding constraint** before
anyone spends a one-way change on it.

## Where I think 030's root cause is under-specified

030 names the root cause as: `system.agent.generic.requests` has one partition, so
Kafka can deliver to only one consumer. That is true but, I believe, **not the
constraint that binds**:

- `agent-chassis` runs `replicas: 1` with no HPA **[VERIFIED 2026-07-20]**. So even
  at `PartitionCount: 8`, one consumer would be assigned all eight partitions.
- Each message is processed **synchronously** on the consuming goroutine
  (`agent.go processRequests` → `processMessage`), and `coordinator.go
  continueExecution` walks consecutive local steps in an inline `for` loop, spawning
  no goroutines anywhere in that file **[VERIFIED — grep returns nothing]**.
- Measured: one message held the consumer ~8 minutes **[VERIFIED — see NOTES]**.

So the ordering constraint is not what serialises the fleet; **one replica running
one blocking goroutine** is. This matters because it changes which fix is correct:
raising `PartitionCount` is **one-way** (Kafka cannot reduce it) and on its own
would buy nothing.

**Status of that claim: UNCONFIRMED.** Filed for diagnosis 2026-07-20 as corr
`78470372-7617-40e4-888c-66cac94006bf` rather than asserted, because it is a
structural claim that would redirect the fix — exactly the case CLAUDE.md says to
file before committing to. A REFUTED verdict here is a good outcome and gets
recorded as a correction in NOTES.

## Open question that changes the severity

030 says "every trigger from every concurrent session" funnels through this lane.
But spawned agents take a job-specific `REQUESTS_TOPIC` from env and only fall back
to the generic topic when it is unset (`agent.go setupConsumers`). If most fleet
work runs on job topics with their own consumers, the blast radius is much smaller
than 030 implies. **[UNVERIFIED]** — included in the diagnosis submission.

## Phasing

1. **Characterise** (done). Re-verify topology, measure the drain shape properly,
   read the consume→process→coordinate path. → NOTES, RUNBOOK.
2. **Establish the binding constraint** (in flight). Diagnosis run above.
3. **Correct 030 in place.** Append a dated correction covering (a) the
   under-specified root cause, (b) the landmine's wrong commit mechanism. Append —
   do not rewrite: other threads may be editing that file.
4. **Then** choose a fix, and put it through the council gate before committing.

## Fix preference, pending step 2

030's own candidate 1 — **acknowledge the request on publish** — remains the best
first move regardless of how step 2 lands, and I would sequence it first:

- It is the cheapest change and needs no topology decision.
- It fixes the part that actually costs sessions time. The latency is tolerable;
  being unable to distinguish "queued" from "dropped" is not. Both recorded
  incidents (a duplicate paid council run; an abandoned investigation) were
  misdiagnoses, not timeouts.
- It is not one-way, unlike raising `PartitionCount`.

Concretely: have the trigger scripts print the current `LAG` and the head-of-queue
age after publishing, so the operator sees "queued behind N messages" instead of a
query that returns 0 rows for half an hour.

**A throughput fix is a second, separate decision** and should not be bundled with
it. If step 2 confirms the constraint is the inline loop, the candidate set is
"scale replicas" and/or "stop running multi-step segments on the consumer
goroutine" — both far larger changes than partitioning, and both needing the
council gate.

## Landmines carried from 030

- **Do not fix this by making triggers retry.** A queued dispatch that gets retried
  duplicates paid LLM work.
- **Do not raise partition count without deciding the key.** Unkeyed messages across
  partitions lose ordering the state machine may rely on. And it cannot be undone.
- The ~300 s post-restart drop documented in `CLAUDE.md` is a **separate, real**
  failure; fixing this does not retire it.
