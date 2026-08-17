# SUMMARY 2026-08-17 — the tool suggester saw half the library, and the shape behind it

## What we're trying to do

Stop a class of bug where a query quietly returns part of a list, hands it to a language model, and the
model answers confidently as though it had seen everything. `bugs_open/275` is one instance: the agent
that decides which interactive tools a website should have was choosing from **30 of 74**.

## Where we've come from

The query fetching our tool library ended with an instruction to return the first 30 entries
alphabetically. The library has grown to 74, so 44 tools could never be suggested for any site, and
which 44 was decided by the alphabet rather than by relevance.

Nothing ever looked wrong, and that is the defining feature rather than a detail. A model returns
sensible-looking suggestions whether it saw 30 tools or 74. There is no error and no missing output —
only a quieter kind of wrongness that can be found solely by reading the query. It was reported on
14 August with 68 tools and 38 hidden; three days later it was 74 and 44, because the library grows and
the cap does not.

## What we've done

**Fixed the instance, but not by the obvious route.** Simply deleting the limit would have tripled the
prompt. Measuring first showed that one field — the description — was 80% of the payload, with a median
of under 400 characters and a worst case of 2,500. So we bounded the descriptions instead of the
coverage: **the whole library for about a quarter more prompt**. Live and verified — 44 previously
unreachable tools are now selectable.

**Fixed the class, which is the larger half.** This was never really one bad query. Twenty-six such
limits exist across our live configuration, and all of them run through a single piece of code. That
code now notices when a result comes back exactly as full as its own limit — the signature of a
truncated view — and says so, naming the step. It changes nothing about what any query returns.

**Checked the rest, and the first count was wrong.** Of the seven limits I initially called suspicious,
two turned out not to bound their result at all and two more are work queues where the remainder simply
arrives on the next run. **Three genuinely bite** — and the two we have not fixed are worse than
expected: one agent picks internal links from at most 15 of 68 candidates, and another gets 10 of up to
107 pages as context. Both are recorded, neither is filed as its own ticket, and that is a decision for
the owner.

**Then the review sent it back, and improved it.** The sharpest objection was that our fix had
reproduced the bug one level down: descriptions were being shortened *silently*, so a tool whose
distinguishing feature appears late now reads as generic with nothing to signal the loss. Shortened
descriptions now carry a visible marker, costing about 3%. Two further objections found a guard that
could not have caught the problem it guarded against, and a migration that had been applied without
being recorded anywhere.

## Where we are now

Both configuration changes are **live and verified**. The detection code is committed and waits for the
next build like any code change. The review is on its second round.

One claim was withdrawn before it was ever made. Tracing what happens to suggestions the model cannot
match, we found they get built from scratch — so the cap should have been causing us to rebuild tools we
already had, and 18 of 19 such requests named an existing tool. Checking the dates refuted it: those
library entries were created *by* those builds. **There is no measured waste**, and the harm is exactly
what the original report described.

## Where we're going

Two things are owed and both are written where the next person looks. The report's own end-to-end proof
needs the suggester to actually run, which has not happened yet — the checks done so far are necessary
but not sufficient, and the record says so. And the new warning will, once the code ships, answer a
question nobody has ever asked in production: which of these caps are silently biting right now.

The two unfixed instances are the open decision. One of them shows a model 9% of its available context.
That is worse than the bug this work started from.
