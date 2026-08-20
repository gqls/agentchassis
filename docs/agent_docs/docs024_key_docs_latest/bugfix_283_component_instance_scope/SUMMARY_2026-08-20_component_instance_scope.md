# SUMMARY — component instance scope (2026-08-20)

**What we're trying to do.** Make every interactive component on the estate safe to place
twice on one page. Historically each calculator and tool hard-coded its element names, so two
copies on a page would silently read each other's inputs. The fix renames every element with a
per-instance prefix and rewires each script to look elements up by the prefixed name — without
changing what any tool computes, ever.

**Where we've come from.** The mechanical three-quarters of the estate was converted last week
by a deterministic rewriter. Building the careful pipeline for the remaining quarter exposed
that the mechanical batch had quietly broken a third of what it converted (names that travel
through variables were never renamed — bug 324), fourteen of them live. We designed a judged
pipeline — machine renames, a language model does only the script surgery, and a gate
re-renders and refuses anything half-done, altered outside its brief, or cut off — and put it
through nine council rounds, two of which found real blind spots in our own gates before they
could ship. Round nine approved it.

**What we've done (the 19th–20th).** The rolls carrying the hardened gates landed, so
everything armed: the repair batch fixed 28 broken components by machine and waved 35 sound
ones through; the judged pipeline then converted the loan-and-mortgage estate — the canary
first, then 20 of the remaining 22 calculators — each one written only after the gate's
eighteen checks passed, delivered to its owner-managed page through the section editor, and
proven live. The independent arithmetic oracle (170 checks, run before and after, with a
deliberate-sabotage control both times) reads identical to the penny across the whole site.
Six components refused by the gate now wait on humans — every refusal inspected and legitimate
(dynamic names no rewrite can be proven safe on, plus one LLM overreach the gate caught).
Along the way we survived three cross-lane traps and wrote each into the shared registers: a
truncating export, a config key armed on the wrong action's contract (another lane's outage,
fixed by them same-day), and a requeue that leaves a job permanently unclaimable.

**Where we are now.** The two-instance defect is closed for everything converted: 88 of 91
census rows carry scoped, collision-free templates (verified by detector, not by status), the
live pages serve them, and no page anywhere serves a broken tool except three
automation-savings placements on the AI sites, which are gate-refused and await one decision.
The last three generic-tool rows are converting tonight through the same pipeline.

**Where we're going.** The generic pair finishes the planned conversion; the six parked rows
need a human (three of them an owner decision: roll back the snapshot or hand-fix the script);
the loan lane recaptures its byte-identity baseline on the seven pages it pins; and the
estate's steady drip of newly-minted tools (18 arrived unconverted this week) becomes a
follow-on: either the birth gate enforces scoping at creation, or a standing sweep converts
arrivals on a schedule.
