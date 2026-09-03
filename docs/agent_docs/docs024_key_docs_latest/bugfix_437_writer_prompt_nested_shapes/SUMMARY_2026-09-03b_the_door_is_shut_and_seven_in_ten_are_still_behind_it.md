# SUMMARY 2026-09-03b — the door is shut, and seven in ten are still behind it

*Second summary of the day. The first was written before the release; it ended with "after
the next release, confirm at the artefact". This is that confirmation, and one finding the
first summary could not have contained.*

## What we're trying to do

Stop a class of failure where a page is planned, approved, and then never appears — while
other live pages go on linking to it. Six sites had pages in that state, some for weeks.
The goal was to find why 119 builds in a fortnight failed identically, and to shut the door
rather than patch the symptom.

## Where we've come from

The failures all said the same thing: the writer had produced a sentence where the page
design called for a list of decision points. That reads as an unreliable AI writer, and the
bug was filed along those lines.

It was not the writer. The instructions we give the writer include a worked example of the
shape we want, and that example is **generated** from the page design rather than copied
from it. The generator could only describe a list of simple values; asked to describe a list
of *structured* items, it flattened them to a bare name and produced an example that showed
a sentence. The writer copied the example it was given, every single time. The safety check
that refused the result was right every time too. So the 119 failures measured our
instructions, not the model — which is why there were no lucky passes.

The fix teaches the example-generator to show the real nested shape. It is small, and it
turned out to affect exactly one component in the whole library.

## What we've done

The change went in as two halves that are safe in either order, both now live: a code half
that works out the real shape, and a database half that uses it when present. The reviewer
council approved it.

Then, this afternoon, we confirmed it — properly, rather than on the first encouraging sign.
Six pages have been written since the new code started running; all six were given the
corrected instructions and none the old broken ones. Four of those have gone all the way to
being built and published, across three different sites, and every one stored its content in
the right shape.

We also looked at one of the finished pages on the live web. The advertising regulation map
on advertise.co.uk — a page this bug had left stranded — now loads and shows seven properly
laid-out decision branches where there had been an unusable blob of text. Nobody fired
anything at it; it repaired itself on the routine sweep, exactly as predicted.

Along the way we nearly recorded the opposite conclusion. The obvious way to check whether a
fix worked is to count the jobs still carrying the error, and that count said three had
failed *after* the fix. None had. The queue keeps the last error message on a job after the
job moves on, and any later touch of that job re-dates it, so an old failure resurfaces
looking new. The giveaway was that two of the three jobs were marked *complete* while still
carrying a failure message. That trap is now written down for everyone, because it fails in
the direction that makes you distrust a fix that is working.

## Where we are now

The door is shut. New pages of this kind build correctly, and we have watched them do it end
to end, on live sites, with the checks arranged so they could have told us otherwise.

The finding the earlier summary could not contain is how much of the existing damage the fix
reaches, and the answer is: less than a third.

Seventy-three pages' worth of work was hit by this. **Fifty-two of them — just over seven in
ten — cannot recover on their own, ever.** When a job fails twice it gets branded
"unresolved". That brand deliberately keeps the job open, and an open job is one the system
will not re-create, so the brand permanently blocks the very retry that would now succeed.
There are 251 such records across four sites.

This is worth being precise about, because a related number *is* improving and could easily
be mistaken for this one. A separate counter tracks pages drifting *towards* the brand, and
it decays on its own as old failures age out — one site dropped from six at-risk pages to two
in the three hours we were working. The branded records do not decay. They are permanent
until a person clears them.

So: about three in ten of the damaged pages will heal themselves, and are doing so. The other
seven in ten are waiting on a decision, not on a repair.

## Where we're going

**The immediate question is for the owner, and it is a judgement call rather than a technical
one.** Clearing those 251 records would let the system retry 52 pages that would now succeed.
We deliberately did not do it. The reason for waiting was that we wanted to see a real page
build successfully first — that has now happened four times over, so the evidence is settled.
What remains is that this means changing records in bulk across four sites that other people
are actively working on, and that is the owner's call to make.

Beyond that, two pieces of work stay open, and the second is the more valuable:

1. **Nothing repairs a build that failed this way.** It fails, retries identically, and is
   marked terminal. There is no path that regenerates the one bad field with the error in
   hand. The 52 blocked pages are what that gap costs, now measured rather than described.
2. **Nothing notices an active page that live pages link to and that has never been built.**
   That silence is why these sat for weeks: the failure was loud in the queue and completely
   invisible anywhere a person would look.

One thing to keep watching: we were worried that showing the writer a filled-in example would
make it invent decision points where the source material had none. On the first real evidence
it is being conservative rather than over-producing — seven of twenty-one steps carry a
filled list, and the rest are correctly left empty. That is a first reading on a small sample,
not a settled answer, and it wants another look as volume grows.
