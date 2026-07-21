# SUMMARY — dispatch queue serialisation (bugs_open/030)

**2026-07-21** · first milestone snapshot for this workstream. Written to be read
aloud. Five parts: what we're trying to do · where we've come from · what we've
done · where we are now · where we're going.

---

## What we're trying to do

Make it so that when a session fires a trigger at the cluster — a council review, a
diagnosis run, a discovery pass — it starts promptly, and if it doesn't, the session
can tell *queued* from *dropped*. Today a dispatch can sit invisibly for tens of
minutes with no acknowledgement, which looks exactly like failure. That has already
cost real work: sessions have re-submitted (paying twice) or abandoned
investigations, each time reasonably concluding a message was lost when it was only
waiting. Bug `030` is that problem.

## Where we've come from

`030` was filed on 2026-07-19 with a clear symptom and a plausible root cause: the
work topic has one partition, so only one consumer can ever read it, so everything
serialises. The suggested fix was to split the topic into more partitions.

That framing turned out to be right about the *symptom* and wrong about the *lever*,
and getting from one to the other took an unusually messy afternoon that is worth
being honest about. Three separate sessions, including this one, tried to measure how
fast the queue drains, and produced three different "measured" rates — 0.21, 2.4 and
0.62 messages a minute — each arithmetically correct. I contributed the middle one,
by correcting another session's figure and then making the mirror-image error myself:
publishing from the first half of a measurement I had already set running. All three
mistakes have the same cause, and it is the most useful thing we learned about
*measuring* this system: the queue moves in fits — it stalls for eight to fifteen
minutes on a single slow message, then bursts — so any short measurement gives a
confident, wrong number. There is, in fact, no single drain rate to measure; the
speed is just however long the job currently at the front happens to take.

## What we've done

Once we stopped asking "how fast does it drain" and asked "what is actually in it",
the picture resolved quickly and firmly:

- **The queue is 93% our own housekeeping.** Two scheduled jobs — an AI-endpoint
  health check and a build-pipeline trigger — are 84% of everything in the lane. The
  council and diagnosis dispatches the bug is *about* are about 6%. So this was never
  sessions competing with each other; it is routine cron work running faster than the
  single consumer can service it.
- **The arithmetic closes exactly.** Over 70 minutes the scheduler added messages at
  2.6/min and the consumer removed them at 1.4/min; the backlog grew by precisely the
  difference. Nothing else is going on.
- **The blast radius is narrower than the bug claimed.** Each spawned worker gets its
  own private queue and its own pod — over 800 of them — so only top-level "start
  this job" messages funnel through the single lane. The bug's "everything funnels
  through" is true of triggers, too broad for work.
- **We found a trap guarding the whole thing.** Both heavy jobs are configured to run
  every 30 seconds but actually run every 60, because of a timing coincidence between
  the job interval and the scheduler's own tick. So the lane was receiving *half* its
  configured rate — and anyone who "corrected" that discrepancy would have doubled the
  load overnight.

**And today we made the first fix — config only, no code, live immediately.** We
raised the two dominant jobs to reduce the load: the health check from a 30 to a 60
second setting, and the build trigger from 30 to 120. The build trigger is the one
that matters, because it halves the rate at which the *expensive* build chains are
started — the ones that actually clog the single consumer.

> **CORRECTED same day:** an earlier draft of this paragraph said the new values were
> "clean multiples of the scheduler tick, so the every-30s-means-every-60s trap is
> gone." I then measured it and found I had the mechanism backwards. Every task
> actually fires *one tick later* than its setting — the trap is universal, not
> special to the 30-equals-30 case — so the jobs now fire every 90 and 150 seconds,
> not 60 and 120. That is *more* headroom, so the fix is if anything better than
> intended; but the tidy "clean multiples" story was wrong, and the honest rule is
> "effective period = your setting plus one tick." Full detail in the bug file.

## Where we are now

The change is live in the database and being verified against the running scheduler
as this is written. Nothing was rebuilt or redeployed; `interval_seconds` is a column
the scheduler reads every tick, so it took effect at once and is trivially
reversible (old values were both 30).

We have deliberately **not** touched the deeper question — whether the real
constraint is the single partition or the fact that the one consumer processes each
message to completion before taking the next. The evidence points hard at the latter
(one replica, no autoscaler, and the consumer visibly running a workflow's steps
back-to-back for fifteen minutes at a time), but that is a claim about platform
structure, so it was filed for independent diagnosis rather than asserted. Awkwardly,
that diagnosis run is itself stuck in the very queue it is meant to investigate, and
has not started after an hour — so we may end up deciding without it.

## Where we're going

- **Watch whether the config change is enough.** If scheduled production now sits
  under the consumer, the backlog should stop growing and the tens-of-minutes latency
  should ease without any code change. If it merely slows the divergence, the next
  config lever (slower health check, or a slower build trigger) is understood and
  documented.
- **Decide the structural fix separately, and carefully.** The strong candidate is a
  dedicated lane for the scheduler so cron can never again queue in front of
  interactive work — reversible, and it addresses 84% of the volume. Splitting the
  partition (the bug's original suggestion) is one-way and, with a single replica,
  buys nothing on its own; it has been demoted. Whatever we choose must be checked
  against `bugs_open/029`, which is the *opposite* failure of the same machinery, so a
  fix for one must not reopen the other.
- **Fix the diagnosability regardless.** The cheapest win in the whole bug is
  unchanged by any of this: have triggers print the queue depth on publish, so a
  waiting dispatch reads as "queued behind N", not as silence. That is what would have
  prevented every recorded incident, and it does not depend on the structural verdict.
