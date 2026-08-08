# SUMMARY 2026-08-08 — bug 205: the poison-pill reaper

## What we're trying to do

Stop a live money-burner and close its class. A vet-data verification step was
failing every single run and being re-dispatched forever — and the platform had no
mechanism anywhere that would ever stop it. We wanted the immediate burn stopped,
the mechanism that allowed it made unable to do it again, and the platform able to
announce the next occurrence itself instead of waiting for a human to notice a
bill.

## Where we've come from

The bug was filed on 2026-08-06 by the token-pressure check (built for bug 183) as
"a step with no configured output limit truncates at a tiny built-in default, and
two records retry forever". Investigation showed the filed mechanism was only half
right. The re-dispatch engine was elsewhere: a housekeeping job — the
stale-orchestration-reaper — "rescues" any verification task stuck in progress for
more than twenty minutes by re-queueing it, unconditionally, with no memory of how
many times it has rescued the same task. A task that fails deterministically
therefore loops forever. Measured: 1,576 dispatches in one day, 1,575 failures,
across just 33 doomed tasks — only one of which was burning paid AI calls; the
other 32 failed earlier, invisibly, fetching practice websites.

## What we've done

Three fixes, all live and proven as of 2026-08-07, all council-APPROVED (round 2,
after answering a REVISE with fresh measurements rather than argument):

1. **The reaper now counts its rescues** and parks a task as failed on the fifth,
   with a note naming this bug and increasing wait between retries. Config change,
   live immediately. All 33 doomed tasks were seeded to park on their next cycle
   and did so in one pass at 01:40 on the 7th. The pipeline has been quiet since —
   quiet because parked, verified, not quiet because dead.
2. **The back door is closed**: the task-creation backfill now refuses to mint a
   fresh task for a business whose task is parked, so re-enabling the currently
   disabled sweep cannot silently restart the loop.
3. **The platform now announces the next instance**: any step whose output limit
   resolves to the hardcoded transport default logs a warning naming the agent and
   step on its first run. Eight of 126 active steps have no configured limit; six
   are dormant landmines that will now identify themselves.

Both code changes are pod-proven live on v1.0.1262. Everything is documented:
plan, notes with all missteps, runbook (including how to un-park), backup and
mechanically-generated rollback for the config surgery, a landmine entry, and a
corrected concept-register entry.

## Where we are now

The owner has ruled on all four open decisions (2026-08-08):

1. The vet step gets a real output limit of **8000** — being applied now, then the
   one parked task that needed it will be un-parked to prove a full cycle
   completes.
2. The other **32 parked tasks are cancelled** for now (they fail fetching
   websites that are likely dead or bot-blocking; no cost while parked, cancelling
   matches the 574-row precedent).
3. **Each task type will declare its own park ceiling** rather than inheriting
   the 5 silently.
4. **The shared reaper-accounting mechanism goes ahead** — this platform's second
   or third hand-written "count and park" logic becomes a common mechanism, RFC'd
   on the architecture track, with the per-task-type ceiling (decision 3) as its
   first feature and the collection-tasks reaper as its first consumer.

## Where we're going

Execute the four decisions: cap and prove the vet step end-to-end; cancel the 32;
then design and ship the shared reap-and-park mechanism (policy table plus one
shared function, the existing reaper rewired to call it) so the next reaper is a
one-line opt-in instead of another hand-rolled copy. After that this workstream
closes; the bug file stays in bugs_open per the owner's 08-06 ruling, marked fixed
and live.
