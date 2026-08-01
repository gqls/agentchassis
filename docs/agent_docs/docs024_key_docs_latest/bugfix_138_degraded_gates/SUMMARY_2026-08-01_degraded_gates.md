# SUMMARY — 2026-08-01 — the fix worked well enough to hide its own last unproven part

## What we're trying to do

A council of automated reviewers reads every code change before it lands. Each reviewer
has a fixed budget of words. Run past it and the reply is cut off mid-sentence; we keep
the fragment, mark it damaged, and treat a damaged objection as blocking — because a
serious objection might have been in the part we lost.

The bug was that the record then named the reviewer, as though it had judged. It hadn't;
it ran out of room. And a reviewer that blocks a lot is what we look for when deciding a
reviewer is too noisy to keep. **The failure hid inside its own evidence: the remedy is
to give the reviewer more room, and the invitation is to sack it.**

## Where we've come from

Four fixes were written down: say *why* a round was blocked; measure how often it happens;
give the reviewers that need it more room; and put the important fields first, since it is
the end that gets cut.

The first has been live since 29 July. The third you completed yesterday. The second and
fourth were this week's work — and the fourth was largely disproved rather than built,
which turned out to be the cheaper outcome.

## What we've done

**The record now says which it was**, and has survived three separate rebuilds — worth
checking each time, because a rebuild from an older starting point would quietly remove a
working fix and nothing else would say so.

**We measure the pressure before it bites**, with a report anyone can run and a check that
runs itself every six hours for no cost. It has found three problems on its own, including
one that mattered: **a reviewer we had given double the room had already grown back into
it in three days.** More room buys time, not safety.

**The fourth idea was mostly wrong.** The important fields are already first — that is why
our recovery works at all — and the one case I was certain about has never occurred in
2,713 records. The change I was about to make would have pushed the objections themselves
off the end, which is worse. What survived was the other half: telling reviewers to be
brief, and why. Measured properly, on one reviewer with its room limit unchanged, its
longest reply went from ninety-eight per cent of the limit to fifty-five. You approved
rolling that out to all of them, and it now covers forty-eight of fifty-one.

**Then I found a defect in my own instrument.** The truncation counter I shipped could not
count a truncation — a cut-off call records no token count at all, so the obvious query can
never match one. It reported four across the fleet where the true number was ninety-four,
and six reviewers read "no truncations" while having truncated, one of them fifty-one
times. Worse, the same omission quietly removed those calls from the pressure statistics,
so the instrument was blindest exactly where it was meant to look. Both halves are fixed.

The part worth keeping is not the mistake but its shape: **I had used that broken counter's
zero to argue that the other indicator was necessary.** The conclusion was right and the
measurement was false. That is the most durable kind of error, because every check you run
on the conclusion passes.

## Where we are now

All four are answered, the rollout has been measured, and the coverage risk I flagged has
not appeared: since the change, replies are about fourteen per cent shorter and objections
have gone *up*, not down.

One thing remains unproven, and today it got more interesting rather than less. The message
we show when a round *is* blocked by a cut-off **has still never appeared in production** —
sixty-eight rounds, none of them. But we now know why, which is different from not knowing.
Three reviews have come back damaged since the fix went live, and each was correctly
excluded: two had a serious objection survive the cut, so the round genuinely blocked on
merits and was correctly attributed to the reviewer that raised it; the third was damaged
but still approved. The part of the code that suppresses the wrong label is demonstrably
working, even though the label it guards has never printed.

**And the reason it is getting rarer is that we fixed the cause.** No reviewer has come back
damaged at all in the last 399 reviews — every one of the three predates that reviewer
getting its brevity instruction.

Which leaves an awkward position, and it is the reason for today's decision: **the fix works
well enough that its own failure message may never appear on its own.** We removed the
pressure that produces it. Waiting is therefore a weaker plan than it looked, because we are
waiting to observe the thing we have been suppressing.

## Where we're going

You have decided to induce it deliberately rather than wait, which I agree with: it is the
last unverified thing in this work, it costs one round, and "proven by test" is exactly the
kind of claim this project treats with suspicion. Immediately after, the change is reverted
and the reversion verified — an induced fault left in place is a worse bug than the one it
proves.

If it renders correctly, this closes. If it does not, we have found something a unit test
could not see, which is the whole point of doing it.
