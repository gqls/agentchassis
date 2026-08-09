# mortgagecalculator.co.uk — the improvement loop is engaged, and the law is watched (2026-08-09)

**What we're trying to do.** Adopt mortgagecalculator.co.uk — a hand-built
twelve-tool site — into the framework, so every page and calculator is
generated, checked and improvable by the platform rather than frozen HTML.
The owner's standing rulings: correctness beats fidelity to the originals;
the improvement loops own tool quality; where two calculators are right in
different ways, supply both, clearly signposted.

**Where we've come from.** Yesterday the rebuilt comparator finally fed both
sides identical inputs and settled the arithmetic question: zero rebuilt
tools compute a wrong number; six of nine agree with the originals outright;
three differ because both sides are right about different things; and one
genuine error sits in an original — the stamp duty tool grants first-time-buyer
relief between £500,000 and £625,000, which the rules removed in April 2025.

**What we've done.** The three both-ways tools are now in the improvement
queue, approved and armed. The bridging page switches to the retained-interest
structure lenders actually quote, with a new companion page for the
compound-interest model. The rate forecaster adopts the original's cleverer
over-time model, with a new companion page for simple what-if comparisons. The
fee analyser will show both cost figures side by side, each explained in a
line. Every job embeds its formula and a worked example the new calculator
must reproduce exactly, so a wrong implementation fails loudly instead of
looking plausible. And the owner's legislation question turned out to have the
best kind of answer: the platform already runs a daily fact-freshness sweep
that re-fetches cited sources and checks the quoted wording still stands — the
site just had no facts registered for it to watch. It now does: the stamp duty
bands, the first-time-buyer thresholds, the £500,000 relief cliff the original
misses, and the additional-property surcharge, each quoting GOV.UK verbatim.

**Where we are now.** Five work items are queued and will build through the
normal pipeline. The evidence register is live and enrolled in the daily
sweep from its next pass. Still open: stamp-duty's rebuild needs its dropdown
option values pinned (the id contract missed them, which also blocks automated
fences there); three tools still need id alignment (affordability,
fact-finder, portfolio); and nothing yet connects a registered fact to the
constants inside a calculator's code — a tool encoding a stale threshold
would still pass its checks today.

**Where we're going.** When the five builds land, the replay comparator runs
again — bridging and forecaster should then match the originals exactly.
Then: the "current stamp duty rates" page the owner suggested, written by the
framework so its numbers can only come from the registered facts; and the
bigger piece, designed deliberately and council-gated, an acceptance check
that computes expected answers from the fact register itself — so the next
time legislation moves, the calculators fail their checks the same day the
facts do.
