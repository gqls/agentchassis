# SUMMARY — 2026-08-10 · a plan the council rejects now leaves nothing behind

*(the `bugs_open/227` platform job; the site's own copy/voice thread is separate and finished)*

## What we're trying to do

There is an agent on this platform whose job is to plan an *experience* for a site — not a
page or a tool, but the whole journey a visitor takes, what each button promises, where the
numbers on the screen come from, and whether any of it can honestly be built. It writes a plan;
a panel of five critics then reviews it, with two of them holding a veto. The approved plan
becomes that site's plan of record, and a later build round executes it.

The aim of this job was to make that machinery trustworthy on *any* site, and to make sure the
document it leaves behind is one the critics actually approved.

## Where we've come from

We found the fault by accident, on 8 August. We ran the planner to settle a question about
this site's page ordering, and it came back with a detailed, confident plan about a completely
different site — describing pages that do not exist here.

The cause turned out to be that one site's diagnosis had been written into the shared prompt
back in July, in the imperative, immediately after the agent is told which site it is actually
planning for. It had never bitten because nobody had ever run the agent anywhere else: of 61
plans in the system's history, 59 were for that one site.

Reading the run that exposed it turned up a second, separate fault. The plan was being saved as
the site's official plan *before* the critics voted — and nothing demoted it when they voted
against it. On the run we examined, the panel vetoed a plan at 18:21 and a rejected, fabricated
document was the plan of record eight seconds before the run reported failure.

## What we've done

**The first fault is fixed by making the brief data instead of prompt text.** Each site's
diagnosis, decisions and data contracts now live in a per-experience record the agent reads at
the start of its run; the other site's material was moved into that channel verbatim, so nothing
was lost for the site that had been relying on it. Every trace of the hardcoded diagnosis is
gone from the shared configuration — 48 occurrences across five prompts, down to zero.

**The second fault is fixed by moving the save.** The plan is now written only from the branch
the panel's approval leads to. There is no route from a rejection, an escalation or a failure to
a write — a property the change asserts about itself rather than merely intending.

**And we proved it, on both halves, which took longer than the fix.** A vetoed plan is never
written: watched live, with the count of stored plans staying at nothing while a plan was
drafted, vetoed, and redrafted, and moving only when a second attempt was approved. And a run
that is vetoed and then gives up leaves nothing behind: a full ten-and-a-half-thousand-character
plan, vetoed for needing a server this platform does not have, and not one row written — with
the finished plan demonstrably still in hand, in the exact place the saving step reads from, at
the moment the run stopped.

To get the second half we had to force it, because the system is deliberately built to avoid
giving up: after a veto it is told to shrink the idea to something honest and try again, and it
only escalates on a second refusal. We told the panel it had one round instead of five, ran a
deliberately impossible experience past it, and put the setting back.

## Where we are now

Both faults are fixed, live and verified. All of it was configuration, so none of it waited for
a build, and we have now watched all three changes survive three separate rebuilds.

We also corrected three descriptions inside the agent that still described the old behaviour —
one of them stating flatly that a rejected plan stays official, which is now the opposite of the
truth and precisely the kind of stale note a future reader would have trusted.

The honest record of the job includes three near misses, all of the same shape and all logged:
each time, the check we had written down would have returned the answer we expected *whether or
not the fix worked*. A phrase that was also in the template we were testing for. A row count
that only distinguishes anything on a multi-round run. A "nothing was written" reading from a
run that had died before it could write anything. None of them reached a conclusion, but only
because something else caught them — which is the argument for the habit we have now written
down: before reading a result, say what the failing version of *this* run would look like.

## Where we're going

Nothing is outstanding on this job. Two things are worth carrying forward.

The first is a known gap we chose not to close: running the planner to check something *changes
that site's official plan*, because there is no dry-run mode. That is how a test displaced
another site's plan on Saturday, which we restored by hand. Closing it needs the larger of the
two routes that were costed — an opt-in switch on the shared write helper — and that is a
platform change owing an architecture round, not a bug fix.

The second is that the planner has still only ever run on two sites. Everything above says the
next site will get its own brief and its own plan; the first time somebody actually does it is
still the real test.
