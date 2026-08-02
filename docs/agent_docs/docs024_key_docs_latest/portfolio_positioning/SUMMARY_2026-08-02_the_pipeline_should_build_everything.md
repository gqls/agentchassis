# SUMMARY — the error: the pipeline should be building everything

**2026-08-02.** This summary exists to record an error, at the owner's direction, so it
leads with it.

## The error

**loancash.co.uk was built by hand when the pipeline should be building everything —
and yesterday's summary described that hand-build as the machine running its full
cycle.** The claim in `SUMMARY_2026-08-01c` — "the machine has run its full cycle on a
brand-new site without a person touching the middle" — is false in the way that
matters: the generative pipeline wrote **nothing**. A person (this thread) hand-wrote
every guide, every tool, every line of styling and structured data; the platform's only
involvement was *adopting* the finished site byte-preservingly afterwards — a mode whose
entire design purpose is to generate nothing, and whose success we ourselves measured as
"zero AI writing tasks". The celebrated 16 minutes measured typing speed, not the
production line.

Two mistakes, distinct: the **method** (hand-building, when the standing intent is that
the platform builds these sites) and the **description** (attributing the work to "the
machine" in a milestone document). The second is the worse one, because a summary is
written to be repeated to other people, and this one would have propagated a false
belief about what the platform can currently do. What caught it was the owner asking
one plain question: *did we use the submit-domain trigger, or was there substantial
help from this thread?* The cheap check that would have caught it at writing time: a
sentence crediting "the machine" must name which mechanism did the work — and we had
already counted zero generative work items ourselves.

## What we are trying to do

~150 finance and insurance domains as substantial sites in deliberately different
directions — **built by the pipeline, from the submit-domain trigger, to the best they
can possibly be.** Hand-building is not the method; at most it is a fallback for what
the pipeline provably cannot do, and even that must be proven, not assumed. The
register still governs *what* each site is; the pipeline must become how each site
gets made.

## Where we have come from

Yesterday closed with the register decision-complete (152 domains, 43 propositions,
seats doctrine) and loancash live, adopted, gated and positioned — a genuinely useful
result, but the wrong kind of proof. What it proved end-to-end is the **preservation
path**: adopt a finished site without corrupting a byte. What the programme needs is
the **production path**, and on that, this programme's score so far is: sites built by
the pipeline, zero.

## What we have done

Audited the trigger's fresh-build path against the quality bar loancash set. The path
exists end to end — research → strategy → briefing → site plan → composition → design →
a written page per planned page → deploy, emitting sitemaps, canonicals and structured
data. Six gaps stand between it and "the best these sites can possibly be", in order of
leverage (full detail in `README_where_we_are`, 2026-08-02 entry):

1. **Positioning goes in as a suggestion, not a setting** — the mission brief only
   "weights" the classifier; the register entry must be pre-seeded into the spec fields
   the writer actually reads, protected from being superseded — and *proven* to reach
   the prompt (the planted-marker acceptance test, already queued, gates everything).
2. **No live-origin validity gate** — the checks that caught our own three defects
   (links resolve, sitemap honest, canonicals correct, structured data parses) exist
   only as a workstream script; three fleet sites 404 today on their own links.
3. **Tool correctness is unaudited** — for calculator/rates/quote domains the tools are
   the product; a generated tool needs correctness fixtures (assert the £15 and the
   0.8%/day), not just a responds check.
4. **No fact/citation enforcement** — constraint text can reach the writer's prompt,
   but enforcement means arming the banned-claims detector per vertical; a wrong fact
   in specs gets written into pages and then defended (bug 161).
5. **Truncation is not gated** — a cut-off completion can persist half a page and
   report success; the structure check is doctrine, not a pipeline gate.
6. **The fidelity dial is unbuilt** — high/medium/low modulate nothing. Acceptable for
   now.

## Where we are now

Four sites live — every one of them hand-built or hand-adopted. The register is the
programme's brain and it is done; the pipeline is the programme's hands and it is
untested on this portfolio. The honest position: we do not yet know whether the six
gaps are polish or foundations, because no domain has been put through the fresh path.

## Where we are going

The owner has sharpened the experiment into the right shape: **run the same loancash
proposition through the pipeline, and fix the pipeline until its output matches the
hand-built site** — the new loancash nav (not the pipeline's standard one), the
research depth (every figure carrying its rule name), the article quality, the copy
quality. The hand-built site stops being an embarrassment and becomes the benchmark:
for the first time we have a pipeline target that is fully specified, because we wrote
every byte of it and documented why.

Mechanics, with one safety rule: **the fresh build never runs at loancash.co.uk
itself** — that site row is live, adopted and locked, and a second submission at the
same domain risks the pipeline writing over a serving site. The benchmark runs at a
shadow domain (`loancash.uk` — same name signal, unclaimed in the register, no
Cloudflare zone, so its output can never reach the public), with the L10 register entry
supplied both as the mission brief and pre-seeded into its specs, and a marker sentence
planted. Then diff the pipeline's output against the hand-built site dimension by
dimension — nav model, footer disclaimers, fact-with-rule-name density, copy register,
tools — fix the highest-leverage seam, and re-run. The six gaps above are the candidate
seams; the diff decides the order. Hand-building is over as a default; if the loop
shows the pipeline genuinely cannot produce exact-constant tools, the locked-component
route can carry hand-built tools into pipeline-built sites — a measured fallback to
earn, not a plan to start from. The owner's queue is unchanged: build order across the
43 propositions, and the two residual insurance twins.
