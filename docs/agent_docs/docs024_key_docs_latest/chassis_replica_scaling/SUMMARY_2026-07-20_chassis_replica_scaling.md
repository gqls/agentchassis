# SUMMARY — chassis replica scaling — 2026-07-20

**What we're trying to do.** Make the platform's central coordinator — the
single pod that runs every workflow that isn't a spawned worker — able to run
as many copies as the load needs. The target is thousands of domains, which
means hundreds of thousands of workflow dispatches a day; today the design
tops out at a few thousand, in single file, with half-hour queues.

**Where we've come from.** A fault found in May and never fixed means a second
copy of the coordinator doesn't share the work — it duplicates it, because
each copy listens for replies under a private identity and so every copy
receives every reply. That pinned us at one copy. One copy has its own costs:
every deploy has a window where arriving work is destroyed rather than
delayed, and there is no failover. A separate investigation measured the
throughput ceiling: one dispatch at a time, about one every 25 seconds,
because the consumer runs each workflow segment to completion before reading
the next message. The problem statement for all of this was written earlier
today; the owner then set the scale target, which turned it into a design
question.

**What we've done.** Settled the one question the problem statement said the
design hinged on: replies do not need to return to the pod that sent the
request — the database layer handles any-pod processing correctly — but a
guard in the code actively discards replies that arrive at the "wrong" pod,
after marking them taken. So the cheap fix (a shared listening identity) is
only safe if that guard is deleted in the same change. That reading has been
filed with the diagnosis loop for independent verification rather than
asserted. We also measured current volumes, confirmed the throughput
arithmetic can never reach the target by adding copies or lanes alone, and
wrote the plan.

**Where we are now.** The plan is written and waiting on two things: the
diagnosis loop's verdict on the discarding guard (and on the earlier
throughput claim, filed by a sibling thread), and the owner's answers to three
non-blocking questions — per-domain volume at target scale, consent to
sequencing, and how long finished records must stay queryable. The
prerequisite fixes from the spawn-loss workstream (durable timers,
at-least-once delivery, survivable rollouts) are already built or staged by
that thread and are unchanged by this plan.

**Where we're going.** The design is "Kafka delivers, Postgres decides": the
Kafka consumer shrinks to a letterbox that records each arriving message as a
database row in milliseconds, and a pool of workers claims jobs from the
database — per-workflow ordering enforced by the claim itself, not by queue
lanes. Phase one ships the letterbox-and-workers change alone (no topology
change, ends the half-hour queues). Phase two gives replies a shared listening
identity, deletes the discarding guard, and raises the coordinator to two or
three copies, ending deploy losses. Phase three tunes for scale; phase four
watches the database's own growth, which becomes the next constraint once the
queue stops being one.
