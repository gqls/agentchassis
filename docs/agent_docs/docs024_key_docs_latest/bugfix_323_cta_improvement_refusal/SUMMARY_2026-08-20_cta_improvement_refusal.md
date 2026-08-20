# SUMMARY 2026-08-20 — bug 323 closed: refusals can no longer complete green

Written to be read aloud. Current state only; chronology in NOTES and README_where_we_are.

## What we're trying to do

Stop a repair job from being stamped complete when its own handler declared it could not do the
work — and stop the class, not the instance (CTA/button findings were the instance: 993 jobs over
five months, 22 sites, none ever done).

## Where we've come from

Yesterday: diagnosis confirmed; the fixer's own workflow was taught to honour its own refusal flag
(live and proven the same evening); the routing and lockstep-test half was committed and approved by
the council first round, but could not take effect until a new chassis build.

## What we've done

The overnight build (v1.0.1317) carries that half, and it was verified rather than assumed: the
running binaries on both replicas contain the new code's strings and have lost the old refusal
arm's string, with controls; and a synthetic CTA finding pushed through the live router filed the
"found work I have no handler for" roadmap row — detail preserved — instead of a doomed job. The
probe cleaned itself up. The bug file has moved to bugs_closed with the evidence.

## Where we are now

Closed, with three named leftovers: the CTA copy-writer itself still does not exist (now a visible
roadmap row per site, and a question owned by the copy-editor lane with three customer lanes); the
roadmap row keeps only the first finding's detail per site per category; the historical jobs stay
as history. No decision is waiting on the owner in this lane.

## Where we're going

This lane is done unless something regresses — the lockstep test guards the routing, the parking
branch guards every other path into the fixer, and the watch queries are in the RUNBOOK. The real
future work — the one-component-one-defect LLM editor — belongs to the copy_quality_two_stage
conversation, built once for its three customers, with the owner ruling on write authority.
