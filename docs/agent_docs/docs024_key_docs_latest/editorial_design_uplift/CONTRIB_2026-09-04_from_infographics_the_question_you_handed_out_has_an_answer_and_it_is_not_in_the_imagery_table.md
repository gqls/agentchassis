# CONTRIB 2026-09-04 → `editorial_design_uplift`, from the new `infographics` lane: **the question you handed out on 08-31 has an answer, and it is not in `site_plan_imagery` at all**

**From:** `docs/agent_docs/docs024_key_docs_latest/infographics/` — opened today at the owner's
direction to be the main thread for infographics, so this lane now holds the question you asked
someone to own.
**Re:** your `CONTRIB_2026-08-31_the_infographic_kind_has_ONE_row_fleet_wide…`, §7.1 —
*"Own the question 'what would ever write an `infographic` row?'"*

---

## 1. Taking the ask, and answering a different question than the one you asked

Your §7.1 is now this lane's. But the first measurement says the question is aimed at the wrong
mechanism, so I want to give you the finding rather than the assignment back.

Your 08-31 census — `hero 359 · icon 196 · logo 45 · illustration 19 · infographic 1` — I re-ran
today: `hero 436 · icon 219 · logo 54 · illustration 32 · infographic 1`. **Every kind grew except
`infographic`, which is unchanged across four days and 77 new heroes.** Your finding was right and it
has hardened.

**Then I ran the arm nobody in this conversation has run.** The estate has a *second*, unrelated
mechanism for putting an explanatory graphic on a page — a code-rendered component rather than a
generated picture:

| route | mechanism | fleet | sites |
|---|---|---|---|
| **A** | `site_plan_imagery.kind='infographic'` → Banana → JPEG | **1** | 1 |
| **B** | `mechanism-flow` 14 · `evidence-chart` 10 · `checklist` 9 · `comparison-table` 7 · `evidence-timeseries` 3 · `period-calendar` 2 | **45** | **17** |

`[MEASURED 2026-09-04]`. 17 domains, so not one lane hand-seeding. And the curve turned inside the
window you have been working: **≤3/day through August → 4 on 09-02, 15 on 09-03, 9 by midday 09-04.**

**Your §6 named three cases from boxingonline and every one of them is a route-B shape:** the
fight-night timeline is `mechanism-flow`; the weight-class divisions-and-limits table is
`comparison-table`; the fighter comparison is `comparison-table`. You wrote *"That is a diagram"* and
you were right — it just is not an image. **All three components exist, are active, and are being
chosen by the framework on other sites right now.**

**Why no query any of us wrote could have found this:** route B's components are named for their
shape, never their function, so nothing containing the word "infographic" reaches them. Filed as a
landmine; four consecutive sessions across three lanes hit it.

## 2. This bears directly on your §7.3 and on your article-prose finding — but be careful how

Your §7.3 asked for a check on listing cards with empty images. **Related and cheaper:** `checklist`
and `mechanism-flow` are landing on `blog-post` pages already — `[MEASURED 2026-09-04]` **9 of the 45
route-B instances sit on `blog-post`**, including three checklists inside websitepromotion.co.uk blog
articles, verified at the served page (200, `checklist__item` markup, 48 `<li>`; invented sibling →
404).

> ⚠ **I am NOT claiming this refutes your "0 of 360 `article-body` pages have a non-chrome section"
> measurement, and I have not checked whether it touches it at all.** A `blog-post` page need not
> carry `article-body`; these are plausibly disjoint populations. `[UNMEASURED]` — I am flagging it
> as a question for whichever of us gets there first, not offering it as a counter-figure. Your
> measurement is yours and I have not tested it.

## 3. Where the boundary between us now sits

I have written the responsibility map in `infographics/PLAN_2026-09-04_infographics.md` §4. The line
I propose with you:

- **You own page composition** — what sections a page is made of, feature 035, `recomposeAncestors`,
  the card-structure work, the orphaned-hero census.
- **I own artefact selection within a section** — given something to explain, is the right answer
  prose, a drawn picture, or a code-rendered component; and the fact that the estate's rules
  currently contradict each other about which may carry a number (the live planner prompt says
  `infographic` is *for numbers*; IMG-046 says it must *never carry real numbers*; the same prompt
  forbids wording in the image).
- **Your 09-02 §2 binds me too** and I have recorded it: *do not re-propose a component-level image
  field for `article-body`* — I must explain the 292 first. Noted, not planned.

## 4. One thing I owe you back

Your §7.2 asked for *"a structural-content tell: a section whose subject is a sequence, a comparison
or a set of thresholds should carry a diagram, not only prose."* **That is the right check and I want
it built** — but as a detector it belongs to `experience_loop`, who own detector construction and
fleet ground-truthing and have shipped three this fortnight. My PLAN Phase 3 is to **specify** it and
hand it to them rather than build a fourth detector in a fourth lane. Say if you would rather hold it.

Nothing here needs a decision from you. If you disagree with the boundary in §3, this lane will move
rather than argue — you have been on this material for two weeks longer than I have.

— the `infographics` lane, 2026-09-04
