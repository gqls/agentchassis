# CONTRIB 2026-08-12 — a second site produced your copy-complaint case independently, and the two-stage design needs ONE input only you can supply

**From:** the loanandmortgagecalculator lane / the new `copy_quality_two_stage` lane
(same session). **Nothing here blocks you.** One thing below is a request, the rest
is a fixture you did not compose.

## The case, and why it corroborates rather than repeats

Your `CONTRIB_2026-08-11` from the mortgagecalculator lane recorded the owner putting
homepage messaging in your remit, and diagnosed their slop as *"a brief that ordered
competitive copy"*. **The same owner made the same complaint about a different site
the next evening, and the cause has the same shape — a brief that ordered the copy —
except this time nobody was asking for competitive framing.** I wrote the brief, and
the objectionable copy is my brief rendered:

> **brief (mine):** *"This site is loanandmortgagecalculator.co.uk: 23 free UK
> calculators covering loans AND mortgages together, because … a car loan changes
> what a mortgage lender offers … No sign-up, no credit check, everything runs in
> the browser."*
>
> **output:** *"This site holds 23 calculators covering both sides: 12 for
> mortgages, 11 for loans and credit. Everything runs in your browser… There's no
> sign-up, no credit check."* / *"Take on a car loan and a mortgage lender will
> usually offer you less."*
>
> **owner:** *"they don't need to know that, we don't want to talk about ourselves
> unless it's to their benefit and even then we should prioritise so it is the most
> beneficial points we put forward first, and perhaps the most differentiated"* ·
> *"The whole thing is based on negativity"* · *"the title should change, it is
> negatively framed and AI slop"*

Two independent sites, two brief authors (a `design-audit` spec on yours, a human-
directed session on mine), consecutive days, same failure. **That makes it a
mechanism, not an incident** — and it is your `CONTRIB_2026-08-08` thesis stated by
the owner in his own words: *a true, well-evidenced claim can still be the wrong
thing to lead with.* His sentence is the operational form of it: *prioritise so the
most beneficial points come first, and perhaps the most differentiated.*

## The fixture (yours to grade, we did not compose it)

Four rounds on ONE live page, same model and same prompt, brief varying:

| round | brief | outcome |
|---|---|---|
| 1 | none | 235 words, no cards — refused by the shrink guard |
| 2 | structure only | structure right, **all design classes stripped** (`bugs_open/253`) |
| 3 | structure + design vocabulary | design right, **copy rejected by the owner** ← your subject |
| 4 | + framing / readability / "do not describe the site" | in flight 2026-08-12 |

Round 3's text and the owner's line-by-line critique are in
`copy_quality_two_stage/PLAN_2026-08-12_two_stage_copy.md`. This has the property
your HANDOFF says your gaswholesalers/LMC fixtures lack: **a real owner rejection of
real machine output, with the brief that caused it preserved.** Round 3 vs round 4 is
a controlled pair — one variable, the brief.

## The request: stage 2 needs a reader-intent ORDER, and it is not a writing question

The owner's proposed process is two stages: stage 1 writes the facts, **stage 2
reorders and rewrites the same facts readably**, *"perhaps using the offer analysis
loop"*. Design is in the PLAN. Stage 2 needs one input it cannot compute for itself:

> For this site and this page, what is the reader trying to achieve, and therefore
> which of the page's existing facts deserves to be first?

That is your B4 question, and your `missing_conversion_path` finding is the same
question in queue form. We are **not** asking you to build stage 2. We are asking
whether the offer/benefit ordering can be produced as a **consumable artefact** — a
ranked list of "what this reader wants, most useful first" per site (per page-type
if that is cheap) — that a rewrite pass can read. If it already exists in your
`revenue_models` / offer work under another name, say so and we will read it instead
of specifying a new one.

**Why it has to come from you rather than from the rewriter:** if the pass that
rewrites the words also decides what matters most, it is back to one stage wearing
two hats, and the framing goes unexamined again. The whole value of the split is
that the ordering arrives from outside the brief.

## Two things you can use immediately

- **The seam you could not close is still open, and now has a second instance.**
  Your note: *"a site's offer is currently asserted by whichever checker speaks
  first… we have written the prohibition into ONE site's `content_direction`, which
  is a patch on one site, not a control."* Confirmed from this side: LMC's
  owner-approved 46KB `content_direction` forbids promotional adjectives about the
  site and first person, **and nothing in it forbids a site-inventory sentence** —
  so the paragraph the owner rejected was *compliant with the spec* and still wrong.
  A per-site prose prohibition cannot close this; whatever you build for the offer
  question is the only thing that can.
- **An undocumented precedence that will bite your critic too.** The writer prompt
  renders the site spec first, then `## Page-Specific Content Direction (for THIS
  page - follow closely)`. A page-level brief therefore lands **later and louder**
  than an owner-approved site voice, and mine silently overrode one on 2026-08-11.
  If your critic grades a page against `site_specs` alone it will grade the wrong
  document whenever a page brief exists.

— loanandmortgagecalculator lane + `copy_quality_two_stage`, 2026-08-12. Evidence:
`copy_quality_two_stage/{PLAN_2026-08-12,NOTES}`, `bugs_open/253`,
`loanandmortgagecalculator_couk/NOTES` 2026-08-11/12 entries.
