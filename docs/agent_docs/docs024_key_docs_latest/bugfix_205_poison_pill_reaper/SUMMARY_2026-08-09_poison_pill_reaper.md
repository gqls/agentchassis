# SUMMARY 2026-08-09 — bug 205: the poison-pill reaper, and what it taught us about our own measuring

## What we're trying to do

Stop a live money-burner, close its class, and leave the platform able to announce the
next occurrence itself rather than waiting for a human to notice a bill. That was the
brief on 6 August. By today the money-burner has been off for two days, and the work has
turned into something more useful: making sure the checks we rely on to say "this can't
happen any more" are actually capable of finding it.

## Where we've come from

A vet-data verification step was failing every single run and being re-dispatched for
ever. The filed diagnosis blamed a step with no output limit truncating at a tiny
built-in default. That was half right. The re-dispatch engine was somewhere else
entirely: a housekeeping job that rescues any task stuck in progress for twenty minutes
by re-queueing it, unconditionally, with no memory of how many times it has rescued the
same task. A task that fails deterministically therefore loops for ever. In one day, 33
doomed tasks were re-run about fifty times each — over 1,500 failures, only one of them
burning paid AI calls, which is why nobody noticed the other 32.

Three fixes went live on the 7th and were proven the same night: the housekeeping job now
counts its rescues and parks a task after five; the task-creation path refuses to re-mint
a parked task; and the platform logs a warning whenever a step runs without anyone having
chosen its output limit. All 33 doomed tasks parked in a single pass at 01:40 and the
pipeline has been quiet ever since — quiet because parked, which we checked, not quiet
because dead. The council reviewed it and approved on the second round, after we answered
its objections with fresh measurements rather than argument.

On the 8th you ruled on the four decisions the work had raised, and all four were carried
out: the vet step got a real limit of 8000 and the record that had failed roughly 64
times verified on its first attempt; the other 32 parked tasks were cancelled; and the
"count the rescues and park" logic became a shared mechanism, with each kind of task able
to declare its own ceiling, rather than a third hand-written copy.

## What we've done

Since that milestone, three things.

**We chose output limits for every step that lacked one.** Not by guesswork: we measured
what the comparable steps actually produce, and set limits at roughly double the biggest
plausible output — thirty-two thousand for the site-design step, whose class of output we
have measured at up to twenty thousand tokens; sixteen thousand for the plan-writing and
content-writing steps; eight thousand for the analysis and checking ones. A limit costs
nothing unless a call actually runs that long, and being too mean is what caused this bug,
so we erred generous.

**We fixed the truncations that were still happening elsewhere in the fleet.** Three steps
were being cut off mid-answer at limits chosen too small — a different fault from this
bug, which was about limits missing altogether. The tool-auditing step had been cut off
nine times and is raised from four to sixteen thousand. The page-content writer is raised
from eight to sixteen thousand. Six seats on the review panel go from eight to sixteen
thousand, matching four of their siblings that had already been raised for the same
reason; that change went through the panel's own roster-copying tool rather than by hand,
because editing its two copies separately is how they drift apart.

**And we found that our own check was half-blind.** On the morning of the 9th we reported
that no step in the fleet could fall back to the tiny default any more, and the count read
zero. It was wrong. The query — inherited from the bug file, where it had been quoted for
three days by several people without anyone re-reading it — only counted steps that
already had a settings block to read a limit from. A step with no settings block at all,
which is uncapped by definition and exactly what we were hunting, was never in the
population. The same query also looked only at the top level of each agent's workflow, so
anything running inside a loop was invisible — and that is precisely where the
page-content writer's truncating step lives. Counting properly: 134 steps that call a
model, of which eleven were uncapped, against the 126 and zero we had been quoting.

## Where we are now

Everything is applied, verified and reversible. The vet pipeline holds 606 cancelled and
2,528 completed tasks with nothing pending, in progress or failed. The housekeeping job
runs the new shared logic on schedule. All eleven previously-invisible steps are capped,
and the fleet count of unsized steps is genuinely zero on a check that walks every depth
and keys on what a step *does* rather than on what settings it happens to carry. Three
migrations are applied and recorded in the ledger, each with a backup and a tested undo
script.

The two teams whose agents we changed have been told in their own working notes, dated,
without editing anything they had written. The traps we hit are written down where the
next person will hit them, including the one worth repeating out loud: on the current
models an output limit covers the model's private thinking as well as its visible words,
so a step with eight thousand to spend produced only about four thousand characters before
being cut. Any limit on these models needs to be roughly double what the visible output
needs.

The honest note is that we published three wrong claims in two days — that the new warning
had fired when it never had, that a duplicate configuration row made the system's choice
unpredictable when it is in fact deterministic, and that the fleet count was zero when it
was eleven. Each was measured, and each measurement was real; what was false was the
sentence built on top of it. All three are corrected where they were made, and the shared
pattern is recorded: the check that catches this is asking what *else* would produce the
result you are looking at.

## Where we're going

This workstream is finished. The bug file stays where it is, marked fixed and live, per
the standing ruling.

Three things are left for someone, none urgent. The new warning is now purely a tripwire
for any future step added without a limit — when it fires, that step needs a chosen
number. Four agents carry two live copies of their definition, where only the newer is
ever used, so a change applied to the wrong copy would look right and do nothing; that
belongs to whoever owns definition hygiene, and it has no detector today. And the standing
check that watches for token pressure across the fleet runs on schedule, but we could not
locate the notes it is supposed to write — worth someone confirming, because a detector
whose output nobody can find is the same failure that started this whole case, one level
up.
