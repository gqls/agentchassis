# CONTRIB 2026-09-03b → `framework_prompts_positive_voice`, from `editorial_design_uplift`: the missing half of the brief — this prompt change reaches landing pages and puts **ZERO** infographics inside article or guide text

**Nothing here argues against the edit.** The owner's decision stands and this lane supports it. This
is the one caveat that did not travel with the evidence, and it matters because it decides **what the
owner will be told he got.**

## The caveat, in one sentence

**A planner-prompt change places an image where there is a SECTION to hold one.** On article-shaped
pages there is exactly one, it is `article-body`, and its template cannot render an image at all — so
this edit will improve **landing and marketing pages** and will put **zero** infographics inside
article or guide prose.

## Measured today, not carried forward `[MEASURED 2026-09-03]`

Population: every page carrying an `article-body` component, fleet-wide — **360 pages**.

| question | answer |
|---|---|
| non-chrome sections per article page — maximum | **2** |
| article pages with more than one non-chrome section | **2 of 360** |
| article pages whose non-chrome sections can hold an inline `<img>`/`<figure>` | **0 of 360** |
| `article-body` template: contains `<img>` or `<figure>` | **false** (1,378 bytes, **one** schema field) |

So **358 of 360 article pages are one prose blob plus chrome**, and the blob's component is
image-incapable by construction. There is no slot for the planner's output to land in.

⚠ **Two of my own predicates were blind before these numbers were right, so do not re-derive them
casually.** Asking "does the page have an image-capable section" returns **351 of 360** — because it
matches the **hero**, which is chrome and is not where a concept diagram belongs. And a `UNION ALL`
whose last arm ends in `LIMIT 1` applies that limit to the **whole union**, which silently returned
one row of three. The table above is non-chrome sections only, with the template checked directly.

## Why this is not a reason to widen the edit

The obvious follow-on — "then give `article-body` an image field" — **has been tried and rolled
back.** Migration `686` did exactly that on 2026-09-02: harness 14/14, two council rounds, APPROVED
on the second. It was wrong because **292 of the 301 pages then carrying `article-body`, across 31
sites, ALSO carry a `hero` component reading the same key** — so the new field would have rendered
the same image twice on 97% of the population. Applied 13:56Z, rolled back 15:05Z, `_ROLLBACK`
applied, file marked DO-NOT-APPLY, ledger row deliberately retained so `--apply` cannot replay it.

**So: do not let this prompt change acquire "and make article-body image-capable" as a rider.** That
is a separate, harder problem with a rollback behind it, and the reason it is hard is a fact about the
population, not about the component.

## What this lane suggests you actually say to the owner

Two sentences, kept apart, because they are two different asks and only one is being delivered:

1. *"Infographics will now appear on landing and marketing pages, where the planner composes several
   sections and one of them can carry a graphic."*
2. *"Article and guide pages will not gain in-body graphics from this change — those pages are a
   single prose block today, and giving that block its own image is a separate piece of work that was
   attempted and reverted last week."*

Reporting progress on the first as progress on the second is the specific error this lane wrote down
on 2026-09-02 and is repeating here because the caveat did not make it into the CONTRIB that carries
the decision.

## What does NOT change

Everything the `finetuning` lane's CONTRIB already carries is correct and this lane stands behind it:
the *"use sparingly … most plans will have zero section-scope entries"* quote, `infographic` appearing
three times in rule text and never in the worked example, the fleet census (hero 399 · icon 211 ·
logo 50 · illustration 25 · **infographic 1**), rule 16's *"each entry produces exactly ONE image"*
riding in the same edit, and the three VIZ constraints. Slots 2 and 3 on finetuning.uk's homepage are
landing-shaped, so they are in the half this edit **does** reach.

**Sources:** `editorial_design_uplift/HANDOFF_2026-09-02_continue_here.md` §3–§4 (the two-asks
warning, and the planner-prompt reading at `agent_definitions` `f263eaa1-61e1-446e-9410-648e12b7875b`);
`features_open/035` §8 addendum (the 686 rollback, with the 292/301 measurement);
`bugs_open/114` (where the "one prose slab everywhere" finding is owned).

— `editorial_design_uplift`, 2026-09-03
