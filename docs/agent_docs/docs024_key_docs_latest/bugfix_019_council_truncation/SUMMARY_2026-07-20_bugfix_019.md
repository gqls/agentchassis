# SUMMARY — council truncation bug (019), 2026-07-20

**What we're trying to do.** Stop one over-long council reviewer from destroying
an entire review round. When a reviewer's answer hits its output limit, the
platform was throwing away the whole round — every other reviewer's finished
work, the verdict, the lot — and presenting it in a way that made submitters
believe their own submission was broken.

**Where we've come from.** Four separate threads hit this over two days and
built an unusually good case file between them: reproductions, measurements,
corrections of each other's claims. The file blamed the code that aggregates the
reviewers' verdicts, and a queued diagnosis request pointed the same way. The
open argument was whether to raise the output limit or change the
round-destroying behaviour.

**What we've done.** Counted where dead rounds actually die: nine of eleven
never reach the aggregation code — the round dies the moment the overrunning
reviewer's own call returns, because the provider-level code discards what the
model wrote and reports a hard failure, and the reviewers run in sequence so
everything after stops. Fixed all three layers: the provider code now keeps the
partial answer; a reviewer step can (by explicit configuration, applied to the
35 council review seats only) accept the cut answer and let the round continue;
and the aggregator recovers the verdict from a cut review where possible, or
records the seat as unreadable — which can downgrade an approval to
"send it back" but can never soften an objection. Also found that the councils
log under a generic name (which blinded the automated diagnosis), and that a
step with a limit four times larger truncated the same week — settling the
raise-the-limit argument in the negative. Code and configuration are committed;
the configuration is applied live.

**Where we are now.** The fix is inert until the next image build and roll —
the deployed build predates it — so the bug file stays open and the bug remains
reproducible in production today. The fix itself went through the council gate
for advisory review; verdict pending as this is written.

**Where we're going.** Roll an image, verify with the pod-grep and the
council-report check in the RUNBOOK, then move the case to closed. Two things
deliberately left on the table for their owners: whether to also raise the
reviewer output limit (we recommend not, as a fix — sized breathing room at
most), and whether experience-planner's compose step should get the same
tolerance (its consumers can't handle partial answers today, so it was excluded
from the narrow scope on purpose).
