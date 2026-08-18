# RFC_037 — the classifier cannot see its siblings, so a 140-domain portfolio's differentiation rests entirely on a hand-written brief per site

**Status: FILED 2026-08-18 by the portfolio_positioning lane, at the owner's direction.**
He asked whether the classifier reads the portfolio register — *"maybe it should so no one copies
our briefs"* — and on being shown the measurement below chose **option 2: feed the register
entry to the classifier**. **He has HALTED builds pending this decision**, so this is on the
critical path, not a background improvement.

Filed rather than built because it adds an input to `domain-research-classifier`, a **shared
seam every site in the fleet passes through** — the 2026-07-28 ruling's case for architecture
review, and the accumulation concern RFC_022 exists to catch.

## 1. The defect

`domain-research-classifier` is the first thinking step of a build: it writes the **identity,
classification, content-direction and design-intent** specs that every later agent obeys.
Its `classify_and_extract` step takes exactly four inputs:

```
["input_data", "search_results", "scraped_data", "site_specs"]
```

**All four are single-site.** The classifier has no representation of any sibling domain. Its
prompt never mentions the portfolio (grep for sibling/portfolio/other-sites/neighbour returns
five hits, all false positives — `portfolio` as a `site_type` value, `overlap` as
layout-matcher scoring, `register` as tone-of-voice).

The portfolio register (`portfolio_positioning/REGISTER_positioning.md`) carries exactly the
material that would fix this — every entry has a **`neighbours:`** line and explicit non-overlap
rules, e.g. M3: *"this family shows the market; it never computes a personal number — link to M2
for that"*. **It is markdown in the repo. No agent can read it.**

## 2. The measurement

Seven live finance sites, each classified independently:

| domain | site_type | category | industry |
|---|---|---|---|
| adversecreditmortgage.co.uk | interactive | hub | null |
| loanandmortgagecalculator.co.uk | interactive-platform | interactive | null |
| loancalculator.co.uk | interactive-platform | interactive | null |
| loancash.co.uk | interactive-platform | hub | null |
| loanzy.uk | interactive-platform | interactive | null |
| mortgagecalculator.co.uk | interactive-platform | interactive | null |
| remortgagecalculator.uk | interactive | interactive | null |

**Seven distinct propositions — a loan calculator, a remortgage calculator, an eligibility
guidance site for people refused credit, an FCA rulebook, an example site — collapse to two
values, with `industry` null on every one.** This is not a risk to be modelled; it is the
current behaviour, measured 2026-08-18. At ~140 domains it is the whole differentiation
premise of the portfolio failing silently.

Second-order: because `industry` is null and `site_type` is not in `verticalDirectoryMap`, the
directory recommender is **already** relying solely on the domain string to decide what a site
is (see `bugs_closed/292`). The classifier's thinness is load-bearing in more than one place.

## 3. Prior art — and why it does not cover this

There IS a cross-site-shaped mechanism: the **`vertical_landscape`** spec, written by
`needs_vertical_research`. It is good, and it studies **external competitors** — its rows carry
notes on MoneySavingExpert's framing and HSH's named-expert model. **It points the lens outward
and never at our own portfolio.** So the framework can already reason about a competitive field;
it simply has never been given ours.

## 4. What differentiation depends on today

One channel: **a human writes the register's positioning into each site's mission brief by
hand.** That is what this lane has done for the pilot and for build #1, and it works. It is also
a single point of failure that does not scale to 140 domains, and it fails **silently** — a thin
mission produces a plausible site that quietly duplicates its sibling.

**This couples to the builder-flow decision the owner has paused for.** If sites are specified
by a hand-written brief per domain, register-in-classifier is belt-and-braces. If sites are to
be built from a short prompt (the `loanzy.uk` model — no register entry, *"built just from the
webdesign.uk prompt"*), then **register-in-classifier becomes the only place differentiation
could come from**, and this RFC is a precondition rather than an improvement. The two decisions
should be taken together.

## 5. Proposed change (option 2, the owner's choice)

Give `classify_and_extract` a fifth input carrying **this site's register entry and its
neighbours' one-line propositions**, plus the entry's explicit must-nots. Concretely:

- the site's own proposition, audience, mode and register;
- for each neighbour named in the entry: domain, one-line proposition, and the boundary rule;
- nothing else — this is a differentiation contract, not the whole register.

**Open design questions for the round** (I have deliberately not chosen):
1. **Where does the data live?** The register is markdown. Options: a `site_specs` aspect per
   site (fits existing machinery, needs a writer); a new table (clean, more surface); or a
   generated JSON artefact the classifier reads (cheapest, another sync-drift risk of exactly
   the kind `099` already demonstrates).
2. **Who writes it?** A hand-run sync from the register, or the register stops being the source
   of truth and the DB becomes it. The second is cleaner and a bigger change.
3. **Is it advisory or binding?** A prompt input the LLM may weigh, or a post-classification
   **collision check** (option 3 in the owner's list) that fails a classification which
   duplicates a sibling. Advisory is cheaper; binding is the only version that cannot be
   ignored on a bad day. They compose.
4. **Blast radius.** Every site in the fleet passes through this step, including the ~40 lanes'
   sites that are not finance and have no register entry. **The change must be inert for a site
   with no entry** — the same "no match, no write" discipline `evaluate_directory_features`
   uses, and for the same reason.

## 6. Cost of not changing

Stated plainly, because it is the argument: the register is the portfolio's central asset — the
reasoning about why 140 domains are 140 businesses and not one repeated 140 times. Today that
reasoning reaches production only through a human retyping it per site. **The first thing that
scales without it is duplication**, and duplicated sites in one vertical compete with each
other in the same search results.

## 7. Relations

`bugs_closed/292` (the classifier's null `industry` is why the directory decision fell back to
the domain string) · `DIR-001` · `RFC_022` (accumulated optional surface on shared actions) ·
`vertical_landscape` / `needs_vertical_research` (the outward-facing precedent to mirror) ·
`portfolio_positioning/REGISTER_positioning.md` (the data) · the builder-flow decision (§4).
