# SUMMARY — 2026-08-11 — the connection setting that was never in force

Read-out for the owner. Written to be said aloud.

---

## What we're trying to do

Fix a bug where a setting we deliberately configured had no effect, and nothing in the
system would ever have told us. Not a crash, not an outage — a knob that reads as live
and is not.

## Where we've come from

Some weeks ago a piece of work raised the number of simultaneous database connections
each chassis pod is allowed to hold, from four to twelve. There was a good reason,
written down at the time: the chassis had just started running a pool of workers
handling messages in parallel, and four connections shared between four workers, two
background loops, a reply path and a retry driver is a queue — and worse than a queue if
anything is unlucky enough to be holding one connection while it asks for a second.

The raise never took effect. Not once, on any pod, since the day it was made.

The reason is small and completely invisible from either end. One piece of code opens
the database connection and sets the limit from the configuration. A few lines later, a
second piece of code — handed that same connection — sets the limit back to four. It
isn't creating its own connection; it is quietly resizing the one it was given. So the
twelve survives for a few milliseconds and then it's gone, every time.

Another team found this a day ago and wrote it up rather than fixing it, on the grounds
that it wasn't obvious what the right number was or who owned the decision. They were
right to leave it, and their note is what we picked up.

## What we've done

Checked first that nobody else was on it — three different ways, because each way is
blind to something. Then checked the bug was real, which mattered more than expected:
the write-up misquoted the code in a way that made the problem look about four times
bigger than it is. The write-up implied every agent was affected. In fact the built-in
default is four, twelve only ever comes from configuration, and exactly two pods out of
ninety-five are configured that way. That correction is the whole reason the fix is
safe: for ninety-three pods the change does nothing whatsoever, and for the two that
matter it delivers the setting somebody already asked for.

Proved the mechanism rather than reasoning about it — including making the test print a
number that wasn't four, so that "it printed four" couldn't be an artefact of a test that
always prints four.

Fixed it by deleting the offending lines, under a rule now written into the code: **you
size what you open, and never what you were given.** Applied the same rule to the other
half of the same function, which opens a second connection of its own and had never
sized it at all — meaning unlimited. That one is dormant today, so it's a closed trap
rather than a repaired fault.

Added a test that would catch anyone putting it back, and then — this is the part worth
noting — deliberately broke the code again to check the test actually fails. It does.
Every existing test in that area builds its objects by hand in a way that skips the
faulty function entirely, so this whole class of bug was invisible to the test suite by
construction.

Put it through the review council: **approved first time**, with three advisory
comments.

## Where we are now

The fix is committed and will go out with the next fleet release. It is not live yet.

Two things are worth your attention.

**The review earned its keep, on something I'd got wrong.** One reviewer noticed that my
instructions for checking the fix had actually shipped named a technique that is
documented, in our own notes, as not working on this particular service. I'd looked up
the hazards for everything I was *changing*, and none of the hazards for the thing I was
going to *check with*. Rewritten, and logged as a mistake worth remembering. A second
team caught the same thing independently within the hour, along with a failure mode
neither the reviewer nor I had.

**I want to be straight about what we can't show you.** There is no honest way to
demonstrate this fix working in production, because nothing in the platform reports how
many connections a pool is holding or how often it has had to wait for one. The evidence
is the test and the build record. Anyone who tells you they watched the pool go from
four to twelve is describing something the system cannot display. By the same token I
can't tell you the old setting was actually hurting us — traffic during the window was
one or two messages a minute, which four connections handles comfortably, so the
absence of errors proves nothing either way. The case for the fix is that a setting we
chose was being ignored, not that we measured harm.

Also, separately: the configuration that sets these pods to twelve exists only on the
live cluster and appears nowhere in the repository. It's safe — it survives ordinary
deployments — but you cannot learn how the fleet is actually configured by reading our
own code, and a neighbouring config file has already gone stale as a result.

## Where we're going

Three follow-ups, deliberately not bundled in, so that the safety argument for this
change stayed simple enough to check:

1. **Make pool pressure visible at all.** Go already tracks how often a program waited
   for a free connection; nothing surfaces it. Until that exists, this class of problem
   is unmeasurable, which is how it went unnoticed for weeks.
2. **Collapse the two database handles into one.** The second one is dead in production
   and every caller already works around it.
3. **After the next release**, check pgbouncer isn't queueing — the one plausible way
   this change could bite, with a specific reading that would show it.

The bug file records all three, along with the exact checks, so whoever picks them up
doesn't start from scratch.
