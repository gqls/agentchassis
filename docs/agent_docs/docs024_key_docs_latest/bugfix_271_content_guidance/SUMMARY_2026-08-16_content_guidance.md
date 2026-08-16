# SUMMARY — the rewrite briefs nothing read (bug 271), 2026-08-16

Written to be read aloud. Plain prose, current state only.

## What we're trying to do

Make the platform actually follow the instructions we give it. When someone —
an operator, or one of the platform's own planning agents — tells the system
"rewrite this page and, specifically, state these six service names, drop that
claim, add a link here", those instructions should reach the model that does the
writing. That is the whole of it.

## Where we've come from

They didn't reach it. The instructions were written into a field on the work
item called `content_guidance`, and nothing anywhere in the system read that
field. The rewrite went ahead regardless, steered only by the site's general
facts and whatever was already on the page, and reported success.

The reason this survived for months is that it did not look broken. It looked
*unreliable*. If your instruction happened to say something the site's own fact
sheet already said, the writer would produce it anyway and the instruction
appeared to work. So the natural conclusion was "the model is being flaky, let
me word it more firmly" — and one lane spent four full rewrite rounds doing
exactly that, three of them chasing incidental faults the failures seemed to
expose. The instructions never reached any prompt in any round.

## What we've done

We found that the platform already had a working channel for this, under a
different name — a field called `suggestion`, which travels all the way through
to the writer's prompt and appears there under a heading reading "Rewrite
Guidance (IMPORTANT: incorporate this into the content)". So the real problem
was never a missing pipe. It was **two names for one thing, with only one of
them plumbed in** — and the unplumbed one was what the platform's own
gap-planner had been using.

Rather than build a second pipe, we connected the old name to the working one at
the single point every piece of queued work passes through on its way to being
done. Nothing is rewritten in the database; the substitution happens in memory,
on the way past, so no historical record is altered. The four pieces of code
that had been writing the dead name now write the live one, and a test now fails
the build if anyone reintroduces it.

The change went through the platform's review council, which approved it first
time and raised three advisory objections. The most useful said that a safety
rule written as a code comment is not a rule — the next person reads "the one
permitted exception" as permission for a second. We replaced it with a test that
enforces the boundary mechanically, then deliberately broke the code to confirm
the test catches it.

## Where we are now

**Live and proven.** The fix shipped in chassis v1.0.1304 and we verified it on
the running system rather than inferring it: a deliberately absurd sentinel
phrase, carried only by the dead field, appeared in both of the writer's prompts
and in the stored page content. A control item carrying no instructions at all
produced no guidance heading, which is what proves the first result was caused by
the fix rather than by something always present.

At the owner's instruction, all twenty-five outstanding work items whose
instructions had been lost this way have been put back into the queue, so they
will be redone with their briefs actually reaching the writer. Ten of those had
previously reported failure at the final publishing step; fifteen were parked
waiting for a human, and re-queueing those was a deliberate owner decision taken
with the trade-off stated.

The bug is closed. There were four recorded missteps along the way, including
one that nearly had us report the fix missing from a build that contained it.

## Where we're going

Two things remain, and neither is urgent.

The first is honest scope: we fixed the one wrong name, not the reason a wrong
name goes unnoticed. The system still has no list of which fields on a work item
mean something, so a future piece of code inventing a *third* name would
reproduce this bug exactly, in silence. Closing that needs a declared vocabulary
that does not exist yet — and the same gap is already recorded against another
open bug, so it is a known, shared piece of unfinished business rather than a
surprise.

The second is simply to watch the twenty-five re-queued items land, and confirm
the pages they touch come out carrying the instructions they were given.
