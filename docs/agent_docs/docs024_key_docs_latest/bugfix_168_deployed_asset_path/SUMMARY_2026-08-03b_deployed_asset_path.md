# SUMMARY — 2026-08-03b — the queue can now change its mind, and it has

*Fourth in the series. The previous one (`SUMMARY_2026-08-03`) closed with a mechanism that was
built, approved and had never once fired. That is the sentence this summary exists to replace.*

## What we're trying to do

Make the work queue capable of learning that a problem it reported has stopped being a problem.
Until today every one of the platform's fifty discovery checks could only ever *add* to the
queue. Each one works out, on every run, what is currently wrong with a site — and in doing so it
also works out what is currently *right*, and throws that half away. So findings outlived the
conditions that produced them, sat in the queue for months, and stayed dispatchable. What makes
that dangerous rather than merely untidy is that the code which eventually acts on an old item
may have changed underneath it: a stale instruction that was harmless when written can become
destructive later.

## Where we've come from

This started as a narrow bug about one image path. Fixing it properly meant changing what a
queued instruction would *do*, and that turned up eleven items that had been false for three days
and would have overwritten live social cards. The review board caught it, twice telling us we
were wrong when we were sure we were right. The lesson was compressed into a sentence we now
repeat: **a predicate change stops the tap, it does not empty the bath.**

That became a design paper, two owner rulings, and — yesterday — a mechanism: a check may now
declare what it has *positively observed to be healthy*, and the system closes those items for
it. Deliberately opt-in, with the dangerous setting off by default. It shipped correct, approved,
and completely inert, because the only check using it is one nothing runs. We said so plainly at
the time rather than calling it working.

## What we've done

We found it a real user. The obvious candidate was the check with the biggest backlog — and it
turned out to be a trap, which is the most useful thing we learned today. Two different parts of
the system deliberately file that same kind of item into a shared slot. One is the check. The
other loads a real page in a browser and reports "this image is broken on screen". An image can
only be broken on screen if the page refers to it — which is the exact condition the first check
treats as proof that all is well. Letting it close items would have silently wiped out every
genuine broken-image finding, across every site. Nothing about that is visible from the code you
would be editing; we only found it by counting who else files that kind of item before touching
anything.

So we used a different check — the one that finds page sections rendering empty. It has a single
producer, and it already contained a tested function answering "does this section actually have
content?", written for another purpose, so the closing logic reuses the answer we already had
instead of inventing a second one that could drift.

Then we measured before writing code, sent it to the review board, and it was approved first time
with fifteen seats. Four of the advisory objections were checkable, so we checked them rather
than filing them; one produced a small hardening. Seven deliberate breakages were required to
prove each safety guard actually fails when it should.

## Where we are now

It is live, and this afternoon it did the thing. A routine sweep of one site **closed four
findings that had been raised in April** — over three months old, which nothing in this platform
could previously close.

The number that matters, though, is the other one: **the same sweep left six of that site's ten
findings open.** Three were still genuinely empty. Two are flagged for a human. One names a
section whose component has vanished entirely — which reads like a fix, but reads identically to
a rebuild having silently deleted it, so the check refuses to guess. A mechanism that had closed
all ten would look better in a headline and would be exactly the disaster the owner's condition
was written to prevent. It closed what it had evidence for and nothing else.

Fourteen more are queued to close on five other sites as they get swept.

One question is deliberately left open and has been ruled on rather than quietly accepted: closing
an item counts toward a rule that shelves problems reported three times, and closing-by-observation
now counts the same as closing-by-repair. It affects nothing today — measured, none of the affected
items are recent enough to count — and changing it would touch the insert path of every work item
in the estate, so it is recorded as a tracked question rather than fixed in passing.

## Where we're going

More adopters, one or two at a time, each measured on the real queue before any code is written.
The pattern is now cheap to copy and the three traps that would have caught us are written down
where the next person will hit them. Beyond that, the harder half of the original ruling still
waits: making the "unresolved" state properly de-duplicated, which needs a coupled database and
binary change, eighty-seven duplicate rows cleaned up first, and someone watching the roll.

The honest one-line status: **the queue can now change its mind — on one check, on evidence, and
it has done so four times.** Not "the problem is solved".
