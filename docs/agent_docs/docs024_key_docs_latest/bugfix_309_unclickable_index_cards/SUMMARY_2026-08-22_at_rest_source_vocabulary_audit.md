# SUMMARY — 2026-08-22 — the second door: watching the components we already have

*Second read-out for this lane. The first (`SUMMARY_2026-08-19`) closed it. This one
reopens and finishes it, and the reason it is a separate file rather than an edit is that
the two together show what the first one could not see.*

## What we're trying to do

Stop the platform serving pages that look complete and carry no data. The original symptom
was one page — an article index with six cards and not a single link on any of them — but
the fault underneath it is a shape, not a page: **a component can ask for data from a place
that does not exist, and when the data does not arrive, the page renders without it,
silently.** No error, no gap, no broken layout. Just an absence that reads as a design
choice. That is the thing we are trying to make impossible.

## Where we've come from

In August the lane found the mechanism and shut the obvious door. When the system generates
a new component, it declares where each field's data comes from; we added a lock that
refuses a component whose declarations point at nothing. That lock went live on 19 August
and it has worked — not one bad component has been created since.

The lane then closed itself, and the closing read-out said so. **What that read-out could
not tell you is that it had shut one of two doors.**

## What we've done

Picked the bug up again, checked that everything previously declared done was still true —
it was, including the original page, still serving its eight linked cards three days on —
and then asked the question the closing summary had not: *what about the components that
were already there?*

They had never been looked at. The lock only fires when the system generates a component,
and components do not only arrive that way — they are routinely written straight into the
database by hand or by a migration, which is exactly how the original broken one got there.
So we counted: **sixty-nine fields, across seventeen components, asking for data that does
not exist and never has. Six of those components are live on forty-six real pages today.**

We built the second door: a nightly check that asks the same question of everything already
in the database. The design decision that matters is that it **calls the very same code the
lock uses** rather than re-implementing the rule. Two copies of a rule drift apart, and then
you have two answers and no way to know which is right. One rule, two doors.

The hard part was making it useful rather than merely present. Sixty-nine existing problems
cannot be fixed in one go — each needs a judgement about that specific component, and fixing
*one* of them last week took an owner decision and hit a safety guard on the way. But a
check that is red every morning is a check people stop reading. So the sixty-nine sit in a
frozen list the check treats as already known — written precisely enough that it excuses
only those exact sixty-nine, refusing new entries outright, and affecting only the pass/fail
light and never the report. We proved each of those three properties by deliberately
breaking them and confirming a test caught it.

## Where we are now

Written, reviewed, committed. The sixty-nine have their own file so they can be worked
through deliberately instead of rediscovered by accident.

**And it has not run yet** — that is the honest state, and it matters more than it sounds.
The check ships as an image with the next fleet release. Until then it cannot execute, and
on this cluster a job whose image is missing displays as *still running*, not failed. So if
it were applied early it would sit there looking healthy having never done anything. The
bug is deliberately left open on that one step rather than marked done.

The review council asked for one thing we genuinely owed — a written procedure for proving
the check really ran after it is deployed — and that is now in the runbook. It also raised
three concerns we could answer with measurements rather than argument, and one that was our
own fault: we compressed the submission to fit a size limit and left one entry describing
the wrong file, which made reviewers report several shipped files as missing.

## Where we're going

Three steps, in order, and none of them needs a decision from you.

**Build and ship the image with the next release, then verify at the artefact** — not at the
make target, which reports success either way. The recipe is written down.

**Then the sixty-nine.** Not urgent — they have been like this for months and nothing is
breaking — but they are real missing content on live sites. The six live ones first. Each
repair deletes its line from the frozen list, so **that file shrinking is the progress
bar**; there is no separate tracking to keep in step.

**And one thing to watch rather than do:** the first time this check goes red for a reason
we did not put there ourselves will be the first proof that the second door was worth
building. Until then its silence is expected, not reassuring.
