# SUMMARY — 2026-08-01 — site A live and proven

First summary in this lane's series. Written at a real inflection: the
highest-stakes of the three deletes is guarded, live, and demonstrated working in
production.

## What we're trying to do

Stop a class of silent data loss. Several of our build steps reconcile by
deleting everything they previously wrote and re-inserting what they just
produced. That is correct — it stops stale rows accumulating — and it is
catastrophic in exactly one case: when the thing producing the new content
returned only *part* of it. A short answer raises no error, so the delete removes
everything the short run did not replace. The result is not a broken page that
announces itself; it is **absence**, which is indistinguishable from "there was
never anything there".

Four places in the platform do this. The goal is that each one first proves it saw
enough of the corpus to be entitled to delete anything.

## Where we've come from

The code index (`bugs_closed/135`) was fixed last week. That fix deliberately
built the *decision rule* as a shared, reusable thing and deliberately did **not**
convert the three siblings. When it went through the review council, the
`bug_historian` seat objected that stopping there is the platform's single most
repeated mistake — one call site gets the rigorous fix, the siblings stay
exposed — and it was right. Rather than widen the patch, that session filed the
objection as its own case, `bugs_open/165`, and left it unowned. This lane picked
it up.

The three remaining sites, in order of stakes:

- **A — page sections.** The table that holds the actual content of every page on
  every site, and the one that has genuinely lost customer content before: an
  interactive game deleted by a routine rebuild, twice, on two different sites.
- **B — site navigation.**
- **C — the internal link registry.**

## What we've done

**Site A is complete: measured, built, reviewed, shipped, and proven live.**

The hard part was not writing the guard, it was deciding what "enough" means. A
guard that fires on legitimate work is worse than no guard, because the first
person it blocks will remove it. So the thresholds were chosen from the live
distribution rather than from the shape of the code, and two candidate designs
were killed by measurement — including the one the bug file itself proposed, which
would have blocked 89 ordinary edits over four months. One of my own denominators
was wrong in a way that would have refused *perfect* rebuilds of exactly the pages
a human had chosen to protect; that was caught by opening the rows instead of
trusting a count.

It was approved by the council first time, with seven advisory objections. Four
seats independently found a real bug — the durable note would have gone silent
after two occurrences, on precisely the page that keeps failing. Fixed and pinned
by a test.

Then it was **induced in production, both branches**: a deliberately mis-stated
page plan produced a refusal with its numbers, a work item for a human, and — the
point of the exercise — every one of the seven real sections left byte-identical.
Cleared the induction, re-ran, and it rebuilt normally.

Meanwhile another session converted **B and C**, generalising the useful half of
site A's code into a shared helper all four sites now use. Site A's private copy
was retired in favour of it, which also promoted one test from guarding one site to
guarding three.

## Where we are now

| | state |
|---|---|
| **A — page sections** | **DONE.** Live on `v1.0.1223`, both branches induced, council APPROVED |
| **B — navigation** | code committed, council APPROVED, **not yet live, not yet induced** |
| **C — link registry** | code committed, **not yet live**; one round-2 workaround was reverted after the council called it |
| shared rule | four consumers, one decision rule, one durable-refusal helper |

Two findings are open and written down rather than fixed:

1. **The refusal message is misleading for three of the four sites.** It ends by
   telling the operator the leftovers will be tidied by a later run — true for the
   code index, false everywhere else, where the whole operation is refused and
   nothing is tidied later. The sentence lives in the shared rule, so correcting it
   touches all four and wants its own review round.
2. **A refusal on one page currently aborts an entire multi-page build.** Found
   while establishing whether a refusal actually stops the work; it is a real
   disproportion. Another session had filed the underlying cause an hour earlier
   from their end (`bugs_open/173`), and this lane contributed the measurements
   showing it affects four loops rather than the one they found — plus the
   fleet-wide census they had listed as not done, which came back clean.

## Where we're going

- **B and C need what A got**: a roll, then both branches induced. Neither is
  proven until then, and a green run proves nothing — the guard is inert on
  healthy input by design.
- **Two consumers still unmeasured**: `page-build-handler` and
  `tool-recreation-handler` route errors rather than failing, so a refusal there is
  recorded but the pipeline reports success. Content is still protected; only the
  visibility differs. One induction each would settle it.
- **The two open findings above**, each on its own merits rather than riding inside
  a bug fix.
- **`165` stays open** until B and C are live and induced. Both the `bug_historian`
  and `architecture` seats asked specifically that it not be closed the moment the
  highest-stakes site was done, because "partial fix reads as done, residual sites
  forgotten" is the pattern this case exists to interrupt.
