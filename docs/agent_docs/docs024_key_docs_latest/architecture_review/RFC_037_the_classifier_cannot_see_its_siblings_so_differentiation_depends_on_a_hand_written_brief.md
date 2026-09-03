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

### OWNER RULING 2026-08-19 — all four open questions ANSWERED, and the scope widened

Recorded verbatim in intent, at the place the design decision is made.

1. **Where does the data live? → A DATABASE.** Not markdown, not a generated JSON artefact.
2. **Who writes it / what is authoritative? → THE DATABASE IS THE SOURCE OF TRUTH.** This is the
   bigger of the two options offered and the owner took it deliberately:
   `REGISTER_positioning.md` **stops being the source of truth** and becomes a rendering of, or
   an input to, the database. Whoever builds this owes a migration path for the 44 entries that
   live in markdown today, and a decision about what the markdown file becomes afterwards (a
   generated view, or retired). **Do not leave both writable — two hand-maintained copies of one
   roster is precisely the drift class `099` demonstrates and this estate has already paid for.**
3. **Advisory or binding? → ADVISORY.** A prompt input the classifier may weigh; **not** a
   post-classification collision check that fails a duplicate. The binding check stays available
   as a later addition (§5 note: "they compose") but is NOT in scope now.
4. **Blast radius / the ~40 non-register sites? → SUPERSEDED BY A WIDER RULING: give them a
   registry too.** The owner's instruction is *"For the 40 non finance sites add a registry, also
   for the rest of the 2000 .uk domains."* So the answer is not "be inert for sites with no
   entry" — it is **there should not be sites with no entry.** Inertness is still required as the
   *engineering* default (a site whose entry has not been written yet must be unaffected, and the
   change must never fail closed on a missing row), but it is now a transitional state, not the
   permanent design.

**⚠ What this ruling turns RFC_037 into.** It was scoped as "feed the finance register to the
classifier". It is now **"the portfolio register becomes a database-backed asset covering the
whole domain estate, and the classifier reads it."** The mechanism is unchanged; the corpus is
roughly an order of magnitude larger. Consequences a builder must not discover late:

- **The register today is 44 entries covering 153 domains** (`REGISTER_positioning.md`,
  `PORTFOLIO_domains.txt`, counted 2026-08-19). The `sites` table holds **43** rows.
- **There is no inventory of the ~2,000 domains anywhere in this repo or the database**
  [MEASURED 2026-08-19 — searched the repo for a domain list, and every `information_schema`
  column named `domain`/`domain_name`]. `z_bundles/old/domainsubmit1.txt` is a log dump, not a
  list. **The inventory is a prerequisite and it must come from the owner** — this is recorded as
  an open ask, not an assumption.
- **The collision invariant does not obviously scale.** The register's rule is "no two entries
  may share (family × audience × mode)", checked by hand and by `check_register.py`. At 44
  entries that is tractable; at ~2,000 it is not a document rule any more, which is an argument
  *for* the database and possibly for reviving the binding check that question 3 just deferred.
- **Neighbour selection becomes a real design problem.** §5 says the input carries "each
  neighbour named in the entry". With 44 entries neighbours are hand-named. With 2,000 they
  cannot be, so the mechanism needs a rule for *which* siblings a site is told about — nearest by
  family, by vertical, by TLD twin — and that rule is now load-bearing rather than incidental.

**Open design questions for the round** (superseded above where the owner has ruled; the
remainder stand):
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

## ADDENDUM 2026-09-03 — the layer BELOW this RFC, and why neither change substitutes for the other

Contributed by the `bugs_open/445` lane (their measurement, their framing, cited with permission).

**RFC_037's fix alone would not deliver differentiation.** This RFC measures seven finance sites
collapsing to two `category` values because the classifier cannot see its siblings. 445 measures the
layer below: **of 216 distinct terms the classifier emits across 33 sites, only 28 can match any
layout — 188 (87%) match nothing.** Four attractor terms decide every site's layout.

So a classifier that reads the register and writes beautifully differentiated tags still funnels
every site onto the same handful of layouts, because differentiation that lands in the unmatchable
87% is invisible to layout selection. **The two changes are complementary and neither is sufficient:
this RFC makes the classifier SAY something different; 445's tag-vocabulary work makes the
difference REACHABLE.** Worth stating in the round, because "we fed it the register and the sites
still look the same" is the failure this addendum exists to predict.

**Status of this RFC as of 2026-09-03, measured, because a peer lane needed a binding answer:**
- **The database half is BUILT.** Migration `511_positioning_register_table.sql` created
  `positioning_register`; it holds **194 rows / 194 domains** (newest 09:27Z today). Ruling
  question 1 ("a database") and 2 ("the DB is the source of truth") are satisfied in structure.
  ⚠ The markdown/DB duality the ruling warned about is still live — `REGISTER_positioning.md` is
  still hand-edited by some lanes and the two-copies question is unresolved.
- **The classifier half is UNSTARTED.** `classify_and_extract` still declares exactly the four
  single-site inputs this RFC measured — `["input_data","search_results","scraped_data","site_specs"]`
  — with no register input and no sibling vocabulary in the prompt. No migration exists in any tree.
- **It is blocked on an owner deliverable**, not on engineering: the widened ruling requires a
  registry for the whole estate, and this RFC records that **no ~2,000-domain inventory exists
  anywhere** and must come from the owner. That ask is still open.

## ~~BUILT 2026-09-03~~ — **ATTEMPTED AND ROLLED BACK. THE CLASSIFIER STILL DOES NOT READ THE REGISTER.**

> **⚠ CORRECTION 2026-09-03 16:01:39Z, before anyone builds on this section.** Migration `734` was
> applied at 11:39:14Z and **never worked**. Its `read_positioning_register` step used `$1` in the
> query while the step config carried **no `params` array**, which is how `query_database` binds
> parameters (`offer-analyser.load_premise`: `"params": ["site_record.site_id"]`). Every classifier
> run after 11:39Z failed with `query failed: expected 1 arguments, got 0`. It was invisible for four
> hours because no site attempted a classification until copyonline was released at 15:49Z; two of
> its runs then failed. **Rolled back 16:01:39Z.**
>
> **What survives, and it is the more valuable half:** `layout_taxonomy` remains in
> `classify_and_extract`'s `input_fields`. That fixes a defect predating this RFC — the classifier
> was shown `null` where the library tag list should be while being told to match it, which is the
> mechanical cause of `bugs_open/445`'s finding that 188 of 216 emitted terms match no layout.
>
> **What is still NOT built:** the register input itself. **RFC_037's classifier half remains
> unstarted in effect**, and the owner's instruction of 2026-09-03 is not yet satisfied.
>
> **To fix forward:** add `"params": ["<path to site_id>"]` to the step config — but the path must be
> PROVEN in this agent first. The classifier has no `ensure_site_record` step, so it does not carry
> offer-analyser's `site_record.site_id`; the candidate is `input_data.site_id`, unverified.
> **And then MAKE IT RUN ONCE before claiming it works:** fire one classification and read
> `orchestration_states.current_step/status`. Every guard in 734 asserted the config was well-formed;
> not one asserted the step could execute, and a `COMMIT`→`ROLLBACK` dry run proves only that the
> migration applies. That is what turned a four-hour outage into something nobody noticed.

## What the attempt did establish (the design below is sound; only the wiring was wrong)

Owner instruction the same day: *"please fix the classifier to read the register."* Built against
the register that EXISTS (194 rows) rather than waiting for the estate-wide inventory, because the
ruling's own engineering default — inert for a site with no entry, never fail closed — makes the
missing inventory a COVERAGE limit, not a blocker.

- **Mechanism**: `read_positioning_register` (`query_database`, no Go change) between
  `read_layout_taxonomy` and `classify_and_extract`; `output_field: positioning_register`; the
  prompt gains one variable. Advisory prose carrying position, audience, mode, stance, must-nots
  and each recorded sibling boundary, deferring explicitly to a Pre-Defined Mission (ruling 3).
- **Inertness** (ruling 4) tested per site before apply: 2 blocks, 4 empties. Also excludes a row
  that self-declares "direction unassigned" — `L9` sits on 6 domains including `webdesign.uk`.
- **It also fixed a defect that made this RFC's problem worse**: `layout_taxonomy` was never in
  `classify_and_extract`'s `input_fields` allow-list, so the classifier was shown a `null` library
  tag list while being told to match it. Found by this lane, measured at the rendered prompt,
  verified independently by `bugs_open/445` whose 87%-unmatchable finding it explains.
- **Council**: `Council-Submitted: f0ad8366-d489-440d-8a3b-59b000de0ff2` — round 1 REVISE (the sketch omitted the prompt edit), round 2 submitted. ⚠ **Round 2 describes a change that has since been rolled back; whoever reads the verdict must not treat approval as evidence it worked.**
- **Still open on this RFC**: the ~2,000-domain inventory (owner); what `REGISTER_positioning.md`
  becomes now the DB is authoritative; and neighbour SELECTION at scale — today the block uses the
  entry's own hand-named neighbours, which does not generalise past a few hundred entries.

