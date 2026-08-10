# SUMMARY — 2026-08-10 — the scheduler that could not stay alive

Written for the owner to read aloud. Full technical case: `bugs_open/240`.
Running record: `NOTES_240_kafka_metadata_storm.md`.

---

## What we're trying to do

Keep the scheduled half of the platform running. Almost nothing here is
triggered by a person — the build pipeline, the three site-discovery rotations,
the reapers that unstick claimed work, the health checks all fire on timers held
in one small service called `kafka-scheduler`. If that service stops, nothing
announces it. The site pages stay up, the database stays up, no error appears
anywhere. Work simply stops arriving, and everything downstream looks idle
rather than broken.

## Where we've come from

This did not start as an investigation into the scheduler. It started as a piece
of ordinary work on the logo lane: eleven live sites are serving their logo as a
photograph-shaped JPEG instead of a small transparent PNG, because a
configuration bug had been storing every logo as though it were a hero image.
That bug was fixed at source yesterday, and today's job was to prove the fix
works and then re-make the eleven bad logos.

The proof worked. I sent one logo rebuild through the repaired path on a
throwaway site, and it came back correct — a PNG, at logo size, stamped with the
right purpose, through the exact branch that used to get it wrong. That fix is
now demonstrated end to end rather than merely believed.

What made this a different day was the twenty minutes in the middle. The job I
filed sat in the queue, untouched, when it should have been picked up within two
minutes. I went looking for why.

## What we've done

The scheduler had been dying and restarting **132 times in thirteen hours**. It
would come up, run for about a minute, exhaust its memory allowance, be killed,
and wait five minutes before trying again. Roughly one minute alive in every six.
It was never *down* — which is exactly why nobody noticed. Across the fleet, work
that should be running at 150 to 330 orchestrations an hour had fallen to between
six and seventeen.

The cause turned out not to be in the scheduler at all. It is in a shared piece
of plumbing that every service here uses to talk to Kafka. When we set that
plumbing up, we left one setting blank. The library's own documentation says
that a blank value means *fetch information about every topic in the cluster* —
and it does this quietly in the background, roughly every three seconds, for the
lifetime of the process, whether or not the service ever sends a message.

That was harmless when the cluster had a few hundred topics. The cluster now has
**over twenty-five thousand**, and twenty-four thousand of them are disposable
ones created for a single step of a single job and then never cleaned up. So
every service in the estate has been repeatedly loading a very large amount of
information it has no use for. The scheduler is simply the service with the
smallest memory allowance, so it is the one that dies. The others are all sitting
noticeably heavier than they should — that they are heavy *and not restarting* is
what proves the problem is shared rather than the scheduler's own fault.

The disposable topics accumulate because the thing meant to clear them will only
act when the entire fleet is simultaneously idle, and then only on its second
consecutive attempt. A working platform is essentially never simultaneously idle.
That guard is itself the fix for an earlier bug, in July, where the same cleanup
was too aggressive and deleted topics out from under running agents, killing
them. We traded a loud failure for a silent one.

With your authorisation I have run a cleanup that removes the disposable topics
using a per-topic safety test instead of a fleet-wide one, so it can run while
the platform is busy: a topic is only removed if the job it belonged to is
finished and nothing has touched it for six hours. It refuses to run at all if
that safety check returns nothing, because "the check found nothing" must never
be allowed to mean "delete everything" — which is precisely how the July bug
behaved.

**Two things I got wrong today, both caught before they did harm.** I wrote a
measurement into the bug file that the probe producing it had not finished taking
— the conclusion held, but I had invented a stronger number than I had. And more
importantly, I discovered late that the command I had been using to count topics
**silently returns a short answer** when there are this many: three readings
eighteen seconds apart gave 21,409, then 23,017, then 5,809. No error, no
warning. I had already written a "correction" into the bug file based on reading
a downward trend into that noise. Both are now retracted in place, and the
counting method that actually works — asking the machine to write the list to a
file and then reading the file — gives the same answer every time. The cleanup
script now refuses to run if two consecutive counts disagree.

## Where we are now

The diagnosis is filed with its evidence, along with a landmine entry so the next
person who reads that Kafka setting knows what a blank value costs at this scale.
The cleanup is running. The logo fix is proven.

The scheduler will stop dying once the topic count is down, but **nothing yet
stops this happening again**. Today's cleanup is a one-off. The underlying
setting is still blank, and the disposable topics will start accumulating again
from tomorrow.

I should be plain about one limit of the diagnosis: I could not take a memory
profile, because the service does not expose one. The attribution rests on the
memory growing hard during a window in which the process was demonstrably running
nothing else and writing nothing to its log. That is strong, and it is not the
same as having watched the allocation happen.

## Where we're going

Three things remain, in the order I would do them.

**Stop the accumulation.** Give the topic cleanup a rule that works per-topic
rather than requiring the whole fleet to be asleep — the safety test the sweep
script already uses is the candidate, and it does not reintroduce the July bug
because it asks about one topic at a time rather than about the fleet.

**Close the door properly.** Set the Kafka setting so that services ask about the
topics they use rather than all of them. This is a change to shared plumbing that
every service depends on, so it wants a review round and a careful look at how
the library handles the job-specific topics that are created on the fly — I have
deliberately not guessed at that.

**Give the scheduler some headroom and a memory ceiling**, so that the next time
something grows underneath it, it degrades instead of dying. This one is worth
doing but worth doing *last*: on its own it would hide the problem rather than
fix it, and the cost would keep growing invisibly.

Then back to the logos. That work is unblocked and the list has been corrected
along the way — one site on the original list of eleven turned out to be fine,
and a different one that was missing from it turned out to be affected.
