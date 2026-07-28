# Consolidation and divergence — where we are, 2026-07-28

*Second in the series. The first was `SUMMARY_2026-07-27_consolidation_and_divergence.md`.
Cold-start for the work itself: `HANDOFF_2026-07-28_continue_here.md`.*

## What we're trying to do

We intend to run thousands of domains from one platform, and we are at fifteen or
so live sites. The thing that stops that scaling is not hosting and not the
cluster — it is building the same thing repeatedly in slightly different ways.
This work exists to find where that is already happening, fix what is worth
fixing, and put something in place that notices the next one early enough to
matter.

## Where we've come from

It started with a near-miss. A design here specified a new public service for the
island machine; a different thread had shipped one to that same machine the day
before, built from the start to serve many tools across many sites. The owner
caught it by asking how the two would fit together. No machine caught it, and the
reason turned out to be more interesting than the mistake: the prior-art search
was done, and was *correct* at the time. Nothing in the platform ever looks again.

Yesterday produced the diagnosis and a consolidation programme. Today was meant
to be execution.

## What we've done

Two shared packages are built, tested and council-approved. `platform/mailer` is
the first email sender anywhere in the code we actually build and deploy — before
today there was none, and the only working one lived in idea.uk's box outside the
build. `platform/httpguard` carries the client-identification hardening from a
proven production incident, which the public API still lacks.

The divergence check is live in the pre-commit hook. A document that proposes a
new program now gets told which programs already exist, with recently-added ones
marked. Measured against real history it fires about once in seventy-five commits
— and auditing the commit from the near-miss now prints the very service that
arrived after the original search. It fired on this session's own handoff, too,
correctly.

Then two things happened that were not on the plan.

**The audit that started all this turned out to be wrong three times over.** Its
"eight byte-identical health servers" were eight different bodies. Its "clear
duplicate" exporters shared a purpose and none of their sixteen functions. Its
"generic" Firecrawl action had no callers anywhere, while the "bespoke" one it
was supposed to replace was live. And the headline number funding the whole
programme — nine single-site actions out of nearly three hundred — recounts to
one. Most of what it flagged was already generic.

So the biggest planned item, generalising the gripper scorer into a configurable
engine, is now recorded as a **won't-do**, on evidence rather than on timing. The
exemplar it was modelled on turns out to be ordinary Go code, not configuration.
And a second scoring operation already exists on idea.uk — an AI rubric scoring
business ideas one to five — whose settings have *nothing whatsoever* in common
with gripper physics. We would have built the abstraction and then discovered
site two could not use it.

**And a test found a live defect in a product we sell.** I wrote a test that
asserts the guarantee rather than examples of it: removing a published
manufacturer figure must never make a gripper look like a better match. On its
first run it failed. A gripper whose maker never published its cup-size range was
being scored as a **match** — because the code, reasonably-looking, treated "we
don't know the size window" as "there is no size window". Six of seven checks in
that file handled a missing figure correctly; this one didn't. A paying customer
would have been told a part fits a cup nobody has published the size of.

That is fixed and confirmed running in production. Two smaller faults went with
it, including one that would have crashed the honesty check at exactly the moment
it mattered most.

## Where we are now

The report pipeline is honest again, verified against the running system rather
than assumed from the fact that we committed something.

The review council looked at that fix and asked for a revision — fairly. It
noticed that the two new pieces of information I pass between steps aren't
formally declared anywhere. It is a small change and it is the next task.

The two shared packages are in an awkward state worth naming plainly: **built,
approved, and used by nothing.** That is the worst position shared code can
occupy — it carries all the maintenance cost and delivers none of the benefit,
while the four different copies it was written to replace carry on drifting.
Connecting them up is now the most valuable open item, and it needs a
conversation with the thread that owns that service rather than a commit.

The through-line of the whole day, and the thing I would want remembered: **every
substantial claim that failed today failed because someone believed a summary
instead of opening the file.** That includes three of my own. The checks that
caught them were all cheap — a grep, a recount, a test that asserts a rule
instead of an example.

## Where we're going

Resubmit the revision, then connect the two shared packages up. After that, the
dossier's public half, built as a room inside the service that already exists
rather than a second building beside it.

The larger question the audit was really pointing at is still open, and better
posed now: not "how do we make the scorer configurable" but "how does a second
site acquire one of these at all". That is the maturity-ladder work — named rungs
and worked examples — which remains the stated method for lifting the whole
portfolio and still has no owner. It is a better use of the next design session
than the thing I nearly built today.
