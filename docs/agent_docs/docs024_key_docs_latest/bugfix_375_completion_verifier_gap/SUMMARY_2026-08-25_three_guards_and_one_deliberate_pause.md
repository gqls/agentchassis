# SUMMARY — 2026-08-25 — three guards, and one deliberate pause

Second read-out for this lane. The first was `SUMMARY_2026-08-24_the_unguarded_completion_writer.md`,
written when the fix was submitted and the verdict still running. Plain prose, written to be read
aloud.

---

## What we're trying to do

Stop a particular kind of quiet lie. When an agent finishes trying to fix something on a site, a
record gets stamped "done". We have a mechanism whose only job is to make that stamp honest: before
the record closes, it re-runs the original problem's own test and refuses the stamp if the problem is
still there.

The bug is that the mechanism only guards one of the doors. Three separate pieces of code can stamp a
record done. One always asks. One never asked. A third bypasses both.

## Where we've come from

Filed on 23 August by a different lane, which found it while working out why a router had closed a
real problem as "done" without fixing it. That lane closed and left a handover. Nobody had picked it
up.

Yesterday we measured the damage — four agents, six closing paths, seven kinds of problem, none of
which has a re-check written for it — and fixed the second door: it can now consult the re-check, but
only where a step explicitly opts in, and it writes a note on the record whenever it skips a check
that existed. That went live on this morning's build.

Then we found the part that made it urgent. Our own reference documentation *warns* the next person
about a side effect that cannot happen, precisely because of this bug. They would have braced for one
wrong outcome and got a different one, silently.

## What we've done

Since yesterday, three more pieces of work, all reviewed and approved.

**A build-time guard.** The moment somebody writes a re-check for a problem type whose closing steps
don't consult one, **the build now stops them** and names every step involved. Before this, their
first warning would have been a marker on a record that had already closed wrongly. It has a second
half that matters as much: a command comparing what we've written down against what's actually
running, because a hand-kept list goes stale the day someone adds an agent and then reads as clean.

I built the guard *before* the re-check, so it fired on the real case rather than one invented to
test it.

**A re-check, written and deliberately left switched off.** This is the first item off our own
maintained backlog of "these should have one". It re-uses the detector's own function rather than
reimplementing it, so the two can't drift apart about what the problem even is.

Switching it on is not one line, and that turned out to be the finding. I tried it: five separate
build guards objected. Four were paperwork. The fifth was serious — there's a timeout sweep that
closes stale work by writing to the database directly, bypassing every check, and until a migration
teaches it to skip this problem type it would close records straight past the new re-check. Switching
the guard on without that migration would recreate an old bug *by the act of adding a guard*.

**And the review caught a real defect in my own work.** The reviewers sent the re-check back for
revision, and they were right. A helper we use reports "readable" for a schema whose field list is
*empty*. My test then found nothing missing and concluded "repaired". So a component whose field
declarations had been **deleted** — exactly the silent-loss case we care most about — would have been
certified as fixed by the guard written to catch it. Fixed, proven by deliberately breaking it again,
and approved on the second round.

Two things about that are worth saying plainly. My own testing missed it because I tested the cases I
had thought about rather than everything the helper can return. And the warning for it was already
written down, in a file I never searched for the function I was building on.

**One thing we did not do.** You asked me to file a bug about some unrouted image records. I didn't,
because when I went to write it up the premise fell apart: the missing handler is deliberate and says
so three times in the code, and the handler *had* run — on three records, which I'd been told was
zero because they'd been archived out of view. Better still, on those three it escalated straight
back to "needs a human", which means the obvious fix would spend a job per record to reach the same
answer. That's useful evidence, so it went into the existing bug that already asks this question.

## Where we are now

Three approved changes; nothing pending. The re-check exists, is proven, and is switched off on
purpose — and the reviewers explicitly accepted that reasoning, so the sequence for switching it on
is a reviewed plan rather than my preference.

The bug stays open. Our bar is *fixed and live*. It's live; it isn't fixed. Every closure through
that door is still unverified and no step has opted in. What changed is that the trap now fails the
build instead of only leaving a marker, that a re-check exists for someone to register, and that
three documents which used to mislead now tell the truth.

Worth knowing: **none of today's work is waiting on a deployment.** The build guard is a test, the
comparison command runs from the repo, and the re-check is switched off. The reflex here is to wait
for the next release; that would idle the lane for nothing.

## Where we're going

**First**, switch the re-check on — five steps, migration first. Four of the five are done or
trivial. The only blocker is that the first one edits a file another session has had half-rewritten
all day, and committing a shared file would sweep their unfinished work into our commit. The handoff
says how to tell when that clears.

**Second**, the structural fix: merge the two stamping paths into one, so "which door did this go
through" stops being a question. Everything we've built this week *manages* that asymmetry — five
mechanisms describing one duplication. Merging them makes all five unnecessary. It goes to
architecture review on its own merits, and the honest first step is an afternoon reading the call
sites to find out whether it's even feasible.

**Third**, optionally, put the new comparison command on a nightly schedule. That's a judgement about
cost, and it's yours.
