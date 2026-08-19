# SUMMARY 2026-08-19 — bug 323: the fixer said "I can't", and we finally listened

Written to be read aloud. Current state only; the chronology is in NOTES and README_where_we_are.

## What we're trying to do

Stop a class of false green: a repair agent that declares, in its own reply, that it cannot do the
work — and whose job is then stamped complete anyway, which also mutes the next audit's re-report.
The instance is `cta_improvement` (button copy and destinations), but the fix has to be one that
stops the class, not the instance.

## Where we've come from

For five months every CTA/navigation finding from five different auditors was routed to the
component-template-fixer, whose code has always replied "needs LLM-driven changes, not programmatic
edits — mark for review". Nothing read that reply. 993 jobs, 22 sites, zero ever fixed; the code
comment beside the flag claimed it stopped the dispatch loop, and it did not. The lane that found
it (302) filed the measurement and, correctly, did not chase it; it also believed the refusal and a
genuine "already done" were indistinguishable in the data — they are not: the refusal carries a
machine-readable flag, the no-op never does (470 vs 299 rows lifetime, no overlap).

## What we've done

Three layers, each an existing estate pattern, all today:
1. **Live and proven** — the fixer's own workflow now honours its own flag: a refusal parks the job
   visibly for a human (the page-builder's pattern) instead of completing. Proven by sending the
   real fixer a throw-away job and watching it park; throw-away deleted.
2. **Committed, waiting for the next chassis build** — CTA/nav findings are no longer sent to the
   fixer at all; they become the estate's standard "found work I have no handler for" roadmap row
   (the owner's 077 ruling), with the auditor's suggestion and acceptance test preserved.
3. **Committed, same build** — a build-time test that refuses any routing of a category at the fixer
   for a fix type the fixer's own code declines (mutation-proven).
Plus: diagnosis loop CONFIRMED; council round 1 submitted; landmine, 016b pattern, register entry
WII-023, three WRONG_CALLS entries; the copy-editor lane told it has a third customer.

## Where we are now

Half live, half inert until the roll. Council verdict pending at write time. Nothing is waiting on
the owner, except an optional preference: until the roll, CTA findings land in the human-review
queue (~980 rows) rather than the roadmap — a one-line config change would route them to the
roadmap now if preferred.

## Where we're going

Read and act on the council verdict; after the roll, confirm at the binary and watch for the first
real parked refusal. The real fix for the *copy* class — a small LLM editor that turns "this
component is wrong in this way" into a one-to-three-field edit for the section-editor — is now
wanted by three lanes (277, 301/083, 323) and should be built once, with the owner's ruling on
whether it may write to live pages without a human; this lane will not build it alone.
