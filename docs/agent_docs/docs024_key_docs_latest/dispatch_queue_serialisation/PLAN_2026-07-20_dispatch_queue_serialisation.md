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

---

## Decisions taken 2026-07-25 (and the one deliberately NOT taken)

Phasing above said: candidate 1 first regardless of how step 2 lands, throughput
as a separate decision. Both halves now resolved.

**Candidate 1 — built and live** (`scripts/dispatch-queue-depth.sh`, wired into
090/097). One correction to the plan's own wording: it said "print the current
`LAG` and the head-of-queue age". The head-of-queue age was dropped on purpose —
reading it costs a kcat pod spawn (~10 s) and it is a proxy for the ETA this
workstream has retracted twice. What replaced it is strictly better and cheaper:
the in-flight orchestration list, which shows *what* you are behind rather than
implying *how long*.

**Candidate 3 — built, inert until the roll** (`EXTRA_REQUEST_TOPICS`). Promoted
over candidate 2 for the reasons already established (partitioning is one-way and
buys nothing at `replicas: 1`), and chosen over another round of interval tuning
because of a new finding: the lane's enabled scheduled tasks went **12 → 21**
between 07-21 and 07-25, putting nominal cron production back at ≈2.5/min. **A
tuning fix decays because nothing owns the total; a lane split does not.**

**Step 2 (the filed diagnosis, corr `78470372`) is now moot as a gate.** It never
started — it is still queued in the lane it was filed about — but the claim it was
filed to test (inline handler, not partition count) is confirmed twice over from
the source and from the live pod, and neither change made today depends on the
verdict. Left filed rather than cancelled: if it ever runs, a REFUTED verdict is
still worth reading.

**NOT taken: concurrency inside the handler.** `chassis_replica_scaling`'s **P1**
("decouple consumption from execution") is written up in that workstream as *the*
fix for this bug's latency, is designed and unbuilt, and its plan explicitly gates
P1/P2 code on reading two diagnosis verdicts. Building it here would be a
competing implementation of another workstream's designed phase — the exact
failure `scripts/who-owns.py` exists to prevent. Lane separation composes with P1
(more thin lanes, same worker pool) and stands alone without it.

**Consequence for closing 030:** the filed defect — dispatches indistinguishable
from drops, and a diverging backlog producing 25–36 min waits — is addressed by
the 07-21 config fix plus these two changes. The **architectural residual** (one
orchestration at a time per lane) is P1's, and belongs to that workstream's ticket
rather than to this one. 030 closes on the roll, naming the residual and its
owner.
