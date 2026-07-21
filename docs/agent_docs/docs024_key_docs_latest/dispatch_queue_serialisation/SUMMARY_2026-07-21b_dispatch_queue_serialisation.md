# SUMMARY — dispatch queue serialisation (bugs_open/030)

**2026-07-21 (b)** · second snapshot today. The morning one
(`SUMMARY_2026-07-21_*`) closed on "config lever pulled, being verified, not yet
proven." This one exists because that verification is now in: the fix measurably
worked. Written to be read aloud. Five parts: what we're trying to do · where we've
come from · what we've done · where we are now · where we're going.

---

## What we're trying to do

Make it so that when anyone fires a job at the cluster — a review, a diagnosis, a
site build — it starts promptly, and if it's going to be delayed, they can tell
"waiting in a queue" apart from "silently failed." The whole problem behind bug `030`
is that those two look identical: a job can sit invisibly for tens of minutes with
no acknowledgement, which reads as failure and has repeatedly caused people to give
up or pay to run the same thing twice.

## Where we've come from

The bug was filed with the right symptom and the wrong lever. It said: there's one
queue with one reader, so everything serialises — fix it by splitting the queue into
more lanes. That framing turned out to point at a one-way, expensive change that
wouldn't actually have helped, because there's only one worker to read those lanes.

Getting from that framing to the real cause took a genuinely messy stretch, which is
worth being honest about: three separate sessions each tried to measure how fast the
queue drains and got three different answers, twelve-fold apart, all arithmetically
correct. (I contributed one of the wrong ones.) The lesson from that mess became the
most useful measurement rule we have: this queue moves in fits — it stalls for
minutes on a single slow job, then bursts — so any short measurement gives a
confident, wrong number. There is no single "drain rate" to quote; asking for one is
the mistake.

Once we stopped asking "how fast does it drain" and asked "what's actually in it,"
the real cause fell out cleanly: the queue is 93% our own housekeeping. Two routine
scheduled jobs — an AI-endpoint health check and a build-pipeline trigger — are 84%
of everything in it. The reviews and diagnoses the bug is *about* are a few percent,
queuing behind the chores. So this was never people competing with each other; it
was two chores running faster than the single worker could keep up, and the backlog
growing without end.

## What we've done

**We made the safe, config-only change — and this time we proved it worked.**

We slowed the two dominant chores: the health check and the build trigger. It's a
database setting the scheduler reads live, so it took effect immediately, with no
rebuild and no deploy, and it's reversible with one line.

Then we watched it, properly, through a busy half-hour of the working day rather than
trusting the setting or reading tea-leaves from two data points. Two things came out
of that watching:

- **The runaway backlog is gone.** Yesterday the queue only ever grew — 82 waiting,
  then 130, then 168, never coming back down. Today, over the same kind of window, it
  bounced between empty and about fifteen, and hit *empty twice*. It clears itself
  between busy patches now. That is exactly the change we wanted.
- **We caught ourselves getting the mechanism wrong, again, and corrected it.** We
  measured the jobs actually firing and found they fire one scheduler-heartbeat later
  than their setting — always, not as a special coincidence we'd claimed the day
  before. That's a small thing, but we'd written the wrong explanation into the bug
  file, so we corrected it everywhere. It also means the change gave us slightly more
  breathing room than intended, which is a happy accident.

## Where we are now

The change is live, reversible, and **measurably working**: the diverging backlog has
become a bounded, self-clearing one. That is a real, verified improvement, and it is
the thing the owner asked for.

It is **not** the whole bug fixed, and it's important not to oversell it. The queue
still has a hard limitation underneath: it has one worker that handles one job to
completion before picking up the next, so a single slow job still parks everything
behind it for seven or eight minutes. Our change didn't remove those stalls — it made
the slow jobs arrive rarely enough that the queue fully recovers between them. So the
symptom people feel (an occasional multi-minute wait) can still happen; what's gone is
the version where it never recovered and just got worse all day.

The deeper cause of those stalls — the one-job-at-a-time worker — is still only a
strong theory, not an independently-confirmed one. We sent it for a formal check, but
that check is itself stuck in the very queue it was meant to investigate and never
ran. It doesn't block anything, because two other lines of evidence already point to
the same cause.

## Where we're going

Three things, in rough order of value, all left for the owner to steer:

- **The cheap win we haven't done yet is the most valuable one for day-to-day pain:**
  make the trigger scripts say "you're queued behind N, expect roughly X" instead of
  exiting silently. That's what would actually stop people mistaking a wait for a
  failure — which is the original complaint — and it doesn't depend on anything else.
- **The structural fix** — giving the scheduler its own separate lane, and/or letting
  the worker take the next job without waiting for the slow one to finish — is what
  would remove the stalls themselves. It's a real code change, so it goes through the
  review gate, and it has to be checked against a sibling bug (`029`) that is the
  opposite failure of the same machinery, so a fix for one mustn't reopen the other.
- **Keep an eye on the config change over a longer stretch.** Half an hour with two
  clean recoveries is convincing but not a week of data; if the backlog ever starts
  growing again, the next lever (slow the health check further, or the build trigger)
  is understood and one line away.
