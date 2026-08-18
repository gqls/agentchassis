# SUMMARY — 2026-08-18b — bug 029: the fix is proven in production, and the real bug finally has a shape

## What we're trying to do

Bug 029 is the oldest open entry in the estate's "builds mysteriously stop" family, filed
2026-07-19. Site builds quietly stop happening and then quietly resume, with nothing failing and
nothing alerting. This lane's job is to find what actually causes it and fix it at the framework
level rather than for one pipeline.

## Where we've come from

The bug has been through three explanations and all three were wrong. Its own title — that hung
jobs saturate a concurrency pool — was refuted in-file in July. The replacement, that a retry was
killing work still in progress, was this lane's own and was withdrawn the same day it was made.
That history set the standard of proof here: everything gets a control, and a claim that could
not have come out false is not evidence.

Earlier today we found and fixed a real, separate defect. Each step declares how long it should
be allowed; the build dispatcher asks for fifteen minutes. That declaration was honoured on the
first attempt only — every retry was silently given five minutes, and anything asking for more
than thirty minutes was given three. The longer a step asked for, the less it got. It shipped,
council-approved, and stood at "live but not proven to work".

## What we've done

**We proved it works, in production, on a real job.** At 18:28 a build dispatch step retried and
was granted the full fifteen minutes it declares, and the job it was waiting on answered inside
that window instead of being abandoned. Both halves matter: the number shows the arithmetic is
right, and the successful answer shows the extra time was actually used.

One case is enough here, which is normally not true. Every retry of this kind in the whole
recorded history before the fix — two hundred and nineteen of them — got five minutes. All of
them, with no exceptions. Fifteen minutes is not a lucky draw; it is a result the old code could
not produce. We also confirmed nothing else moved: first attempts still get exactly what they
always got.

It took two hours and forty-three minutes for a suitable case to appear, and during that wait
there were hundreds of jobs running — just none of the kind that could tell the two versions
apart. Calling it verified on that traffic would have been true in every number and unearned in
its conclusion.

**We closed one open question by measurement.** A suspicion carried into this session — that the
*first* wait, not just the retries, was also failing to read its declared timeout — turns out to
be wrong. The step it was based on honours its declaration in every one of the twenty-nine cases
on record, and every short reading belongs to a retry, which is the thing already fixed. Checked
more widely: eighteen agent-and-step pairs, all honoured, none short.

**And the real bug now has a shape.** Looking at every frozen build job we still hold records
for — eighteen — they are not eighteen problems but one, repeated with unexpected consistency.
Every one froze after an earlier step had already failed, never during normal running. Every one
had successfully started its next job and then stopped about twelve seconds later without ever
recording that it was waiting for anything. Seventeen of the eighteen had asked for that next job
twice. And they all died at almost exactly twenty-five minutes old, within a few seconds of one
another. Things that die at the same age are killed by a clock, not by bad luck.

## Where we are now

One contributing defect is fixed, approved and now proven. **The bug is not closed.** What
freezes a build job is still unexplained — but the question has moved from "why do jobs sometimes
freeze" to something far narrower and testable: what kills the parent a few seconds after it has
successfully started its next child, on the path taken when the previous step had failed.

We also caught ourselves twice today, both before anything was written down as fact: a query that
compared two different agents' settings and reported a fault that did not exist, and a reading of
"this stopped happening yesterday" that turned out to be the age of our records rather than a
fact about the system. Both are logged where the estate keeps its wrong calls.

## Where we're going

The investigation of the frozen jobs is written, targeted and ready, and it is **waiting on one
decision.** Our automated diagnosis reads the code from the shared server, and the fix we just
proved is not on that server — nor is the exact version running in production. Two hundred and
thirty-three commits from today, belonging to a dozen parallel pieces of work, are sitting
unpublished. Sent off as things stand, the diagnosis would read the old code and very likely
report, with confident citations, that the cause is the thing we fixed this afternoon.

That decision is with the owner, and it has a clock on it: we keep about a day of these records,
so the eighteen frozen jobs disappear tomorrow, and after that the investigation would be looking
at an empty table.
