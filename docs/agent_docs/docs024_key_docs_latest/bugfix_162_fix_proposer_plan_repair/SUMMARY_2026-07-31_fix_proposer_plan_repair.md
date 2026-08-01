# SUMMARY — 2026-07-31 — bug 162, fix-proposer's plan repair loop

## What we're trying to do

Stop the fix-proposer agent throwing away good work. When it drafts a plan for fixing a
bug, that plan passes a structural checker before any reviewer sees it — does every edit
name a file, is the same file edited twice by accident, and so on. Until today, a plan
that failed that checker was **discarded**: not judged and rejected, but binned before
review, over a rule the drafting prompt was never told. And binned quietly, with the
column a dashboard reads left empty, so the run reported as clean.

## Where we've come from

We already had the cure. Another lane, working `bugs_open/099`, built a bounded repair
loop: hand the plan back to the agent with the exact problems, let it fix them, once.
That shipped in the chassis a couple of days ago and is live.

They switched it on for one agent, `feature-designer`, and deliberately left it off for
`fix-proposer` — a different lane's live agent, and reaching into one from outside is
how concurrent sessions tread on each other here. So they wrote the switch-on as a
ready-to-run, guarded migration and filed bug 162 so it could not be forgotten. That
worked exactly as intended: the file was waiting with instructions, and this lane picked
it up. The mechanism is deliberately **opt-in**, which is what let it ship without
changing any other consumer's guarantee — and is also why `fix-proposer` was still
exposed.

## What we've done

Confirmed the bug was still real, that nobody else was on it, and that the compiled half
was genuinely running on both chassis replicas — then applied the migration. `fix-proposer`
now routes a structurally-refused plan to a repair step and back, capped at one round,
instead of discarding it.

Two things were worth more than the switch-on itself.

**First, we checked the router against real data, not just against the code.** The new
router asks "is this plan valid?" and we had satisfied ourselves by reading the parser
that it would work. That reasoning was sound and it answered the wrong question. The
real hazard ran the other way: if step results were stored wrapped rather than flat, the
router would never find its answer, and **every good plan would be sent for repair** in a
loop nothing bounds. Only live rows could say which shape we had. They said flat. Hazard
excluded — by measurement, which reading could not have done.

**Second, we found that four tests were decorative.** While verifying, we noticed the
shared action had five ways to abandon a plan and only one left any record an operator
could find — while a comment in that same file assured the reader that finding nothing in
the log meant nothing had been refused. Correcting that meant changing behaviour four
tests protect, so it went to the review council rather than being decided alone.
Approved, nine reviewers, three advisory objections, all discharged. One asked us to
prove by test that opted-out agents could not be affected. We wrote that test, it passed
— and then we broke the code on purpose and it **still** passed, along with the three
older "must not touch the database" assertions. All four were guards that could not fail,
because the assertion everyone had used reports "everything I asked for happened", and
if you ask for nothing that is always true. All four are now real and verified to fail
when they should.

## Where we are now

The bug is **fixed and live**. That half was configuration, which takes effect
immediately, and it is verified at the live row rather than at the migration's own
say-so. The ticket is closed and moved to `bugs_closed`.

The extra hardening — recording a refusal on the two paths that used to give up silently,
plus the corrected comment and the four repaired tests — is **committed but not yet
running**, and will not be until someone next builds the chassis image. We have said so
plainly rather than counting it as done.

One thing is deliberately unproven: we did not induce a live refusal on `fix-proposer` to
watch the repair happen end to end. The method is written down, but running it means
briefly changing a setting on a shared agent, and three other lanes had jobs running
through it at that moment. Proving our own wiring at their expense is not a trade this
lane gets to make for them. It is recorded as a gap rather than dressed up as a pass —
the same mechanism is proven on the sibling agent, which exercised it three times today.

## Where we're going

Nothing here needs a decision. Two loose threads, both cheap and neither urgent:

- **Induce the refusal on `fix-proposer` in a quiet window** to close the one
  unproven step. The runbook has the arm/disarm commands and the check-the-queue-first
  warning.
- **The decorative-test pattern may not be confined to this file.** Thirty-three files
  use the same mocking library; none is wholly inert, so the flaw is per-test rather than
  per-file, and we have audited only the one file we were in. A sweep would be a
  reasonable small job, and the search to start from is in the landmine entry.

The larger point, which is not about this bug: a test that has never been watched to fail
is not evidence. It cost about a minute to find that out here, and it would have been
expensive to learn later — the review that asked for the guarantee would have been
answered, on the record, by a test that never checked it.
