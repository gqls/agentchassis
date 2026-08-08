# mortgagecalculator.co.uk — the calculators are proven, on identical inputs (2026-08-08)

**What we're trying to do.** Adopt mortgagecalculator.co.uk — a hand-built
twelve-tool site — into the framework, so every page and calculator is
generated, checked and improvable by the platform rather than frozen HTML. The
owner's ruling this week reset the bar: correctness beats fidelity to the
originals, the improvement loops own tool quality, and where two calculators
are right in different ways we supply both, clearly signposted.

**Where we've come from.** All twelve tools were rebuilt through the framework,
nine with a proven id contract. But the arithmetic comparison kept misleading
us: the harness derived its test inputs from each page's own shipped defaults,
so it fed the original and the rebuild different numbers and reported the
different answers as divergence. One night it convicted six calculators; hand
checks dissolved every conviction. The comparator, not the calculators, was
the defect.

**What we've done.** Rebuilt the comparator to replay the golden's recorded
inputs — the exact literal values, by element id, selects by value — into the
rebuilt pages, and judged all nine id-aligned tools on genuinely identical
inputs. Six agree outright (to the penny, where the originals round to the
pound). Two of the apparent disagreements were the harness's own blind spot:
pages with two Calculate buttons, of which the capture only ever pressed the
first, recording zeros the original never shows a real user. Three tools
differ because both sides are right about different things: bridging
(retained-interest quote vs compound interest), rate forecaster (a rate path
over the mortgage's life vs simple what-if rates), and fee analyser (total
cash out the door vs true cost of interest and fees). And one genuine error
surfaced — in an original: the stamp duty tool grants first-time-buyer relief
between £500,000 and £625,000 where the post-April-2025 rules remove it,
under-quoting a £595,000 purchase by £5,000. The rebuild gets it right.

**Where we are now.** Zero rebuilt tools compute a wrong number on identical
inputs. The three model-difference tools are queued for the owner's
both-calculators treatment — primary tool keeps one model, the alternative
becomes its own well-signposted page. Three tools still need their ids aligned
before the same proof can cover them (affordability, fact-finder, portfolio),
and stamp-duty's rebuild needs its select option values pinned to the original's
so the automated replay can drive it without a hand-mapped step.

**Where we're going.** Finish id-alignment on the three stragglers, fold option
values into the id contract, and wire the replay comparator into the
experience/tool improvement loops as their acceptance check — so "results don't
differ on identical inputs, and the answers are independently right" becomes
something the platform enforces on schedule rather than something a session
proves by hand.
