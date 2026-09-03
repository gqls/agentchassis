# SUMMARY — 2026-09-03: the calculators nobody checked the arithmetic of

Written to be read aloud. First summary for this lane.

## What we're trying to do

Make a "PASSED" verdict on one of our calculators mean what people take it to mean. At the moment it
does not: the automated check confirms the page loads, doesn't throw errors, fits a phone, and that
*something appears* when you click. It never asks whether the number that appeared is right. So a
mortgage calculator that tells a customer the wrong monthly payment passes every test we own, and its
record says PASSED.

## Where we've come from

The mortgage-calculator team found this the evening before, while answering a direct question from
the owner — "verify the tools". The answer they came back with was that where our verification runs
at all, it isn't checking what the tools are *for*. They wrote it up and stopped for the night.

The platform does own a check that compares actual numbers. It works, and a handful of
hand-written test plans use it. It had simply never been mentioned to the thing that writes the test
plans automatically — whose instructions list four checks it must include and finish with the
sentence "No other check type exists".

## What we've done

Three things, all now running, and one thing deliberately left alone.

**The verdict is now honest about its own scope.** When a calculator passes, the record says whether
any number was compared, and if none was, it says so in plain terms. Nothing changed about which
tools pass — a passing tool still passes. What changed is that the verdict can no longer be quoted as
"the arithmetic is right" when nobody checked the arithmetic. This was chosen first because it fixes
**every existing calculator at once** — there are 187 of these test plans and 116 check no number —
and because it can't rot: the label is worked out from the test plan each time the check runs.

**The single door every automatically-written test plan passes through now leaves a note** when a
plan fills in a calculator's boxes and then checks nothing about what comes out. It records rather
than refuses, because refusing would leave the tool with no test plan at all, and a tool with no plan
is checked by nothing — worse than being checked badly.

**A daily report now runs on a schedule**, because a detector nobody runs is not a system. Its first
real pass graded 241 test plans and found **58** that drive a calculator's inputs and assert nothing.
It leads with how many were written *this week*, not the total, because the old ones don't repair
themselves — a totals-only report would read as "no improvement" for a month after a fix that worked.

The internal review council approved it on the second round; the first round found two real defects
in it, one of which could have made half the work a silent no-op.

## Where we are now

Installed, approved, and **not yet exercised** — and I want to be precise about that gap rather than
paper over it. In the minutes since the new build went out, no calculator has been created and none
has been re-checked, so there is not yet a single real verdict on record carrying its new label. The
handoff says exactly what to look for, and how to tell "nothing has happened yet" from "it's broken",
because both look like an empty result and they want opposite responses.

One thing today taught us cheaply: an earlier build the same afternoon bumped the version, restarted
the servers, and did **not** contain the change. Three facts that each looked like proof, and jointly
proved nothing. We only knew because we asked the running program directly, with a deliberate wrong
answer mixed in to check the question could come back negative.

## Where we're going

The actual cause is untouched: the thing that writes test plans still doesn't know the number-checking
check exists. Teaching it is the obvious next step and it is the one step being deliberately held
back, for two reasons worth stating because they are not obvious.

First, the number check works by recording what a calculator printed when it was known to be working
and defending that figure. A brand-new calculator has never been known to be working — so recording
its day-one output carves today's mistake into stone and then guards it. We have already shipped
that: a stamp-duty calculator using an expired tax threshold, certified green for sixteen months.

Second, a separate open bug means test plans currently point at page elements that have moved, and a
number check that can't find its element *fails*. Switching them on now would not catch wrong
calculators — it would make *right* ones fail loudly and send an automated fixer to rewrite
arithmetic that was never wrong.

So the order is: honest record (done), fix the moved elements (another team, in flight), then teach
the generator — with a firm rule that if it cannot work out the right answer from something other
than the calculator's own code, it writes **no** number check rather than a guessed one. Two teams
have been asked the questions that unblock this. Neither has replied yet, and until they do, going
ahead would be guessing — and a guess here produces the one thing worse than no check: a wrong number
with a green tick beside it.
