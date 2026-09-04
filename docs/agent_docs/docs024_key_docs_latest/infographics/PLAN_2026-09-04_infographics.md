# PLAN 2026-09-04 — infographics: the artefact-selection lane

Full path: `docs/agent_docs/docs024_key_docs_latest/infographics/PLAN_2026-09-04_infographics.md`
Opened 2026-09-04 at the owner's direction: *"become the main thread for everything related to
infographics … determine with the other experience loops, components, imagery, graphics threads
where you fit in in the responsibility set."*

---

## 1. What this lane is for, in one sentence

**For a given explanatory need on a page, decide which artefact answers it — prose, a generated
picture, or a code-rendered component — and make the framework's rules say that consistently.**

> ### THE RULE, in one line (arrived at 2026-09-04 with `news_editorial_features`, their formulation)
>
> **If the thing is made of the page's own words, a RULE re-derives it. If it is an ASSET, something
> must own its persistence.**
>
> That is the top-level fork, and it decides the route before any component is chosen:
>
> | the thing is… | route | durability | needs building |
> |---|---|---|---|
> | **structure** made of the page's own words (table, checklist, ordered steps) | **C** — a `content_direction` rule; or **B** — a mounted component | **C: by REGENERATION** (the rewrite re-reads the rule). **B: by separation** (a sibling section is never rewritten) | C: nothing. B: nothing — the components exist |
> | an **asset** (a picture, a chart of registered figures) | **A** (diffusion) or **B** (`evidence-chart`) | **an owner of persistence is required** — a spec rule can ASK for an image and cannot produce one; the writer emits prose, not a JPEG | `inline_guide_imagery`'s binding (IMG-075) for in-body; a sibling section otherwise |
>
> **Why the fork is load-bearing:** this lane stated the regeneration property too broadly on
> 2026-09-04 (*"the artefact is not durable, the RULE is"*) and `news_editorial_features` bounded it
> the same day. It is true of structure and **false of assets**, and the whole difference is that a
> writer emits words. Their narrowed scoreboard follows and this lane endorses it: **composition's
> justification is removed for structure-inside-prose and SURVIVES for imagery-inside-prose.**

Nobody owns that question today. Every neighbouring lane owns a *mechanism* and each correctly
stopped at its own boundary (§4). The selection rule between mechanisms fell between them, and the
consequence is measured in §2: the estate's own written rules currently disagree about which
mechanism is allowed to carry a number.

## 2. The finding this lane opens on — TWO ROUTES, and the fleet has been counting one

`[MEASURED 2026-09-04, live DB, queries in RUNBOOK §1]`

| route | mechanism | fleet instances | sites |
|---|---|---|---|
| **A** | diffusion image, `site_plan_imagery.kind='infographic'` → Banana → JPEG | **1** | 1 |
| **B** | code-rendered component (`mechanism-flow` 14, `evidence-chart` 10, `checklist` 9, `comparison-table` 7, `evidence-timeseries` 3, `period-calendar` 2) | **45** | **17** |

Route A's single row: `infographic_decision_engine`, mortgagecalculator.co.uk, planned
2026-08-02, the only one in fleet history.

**Route B is not one lane hand-seeding examples.** 17 distinct domains, and the adoption curve
turned in the last three days: ≤3/day through August, then **4 on 09-02, 15 on 09-03, 9 by midday
09-04**. Page-type split: 22 `content`, 12 `landing`, **9 `blog-post`**, 1 `entity-directory`,
1 `section-index`.

**Verified at the served artefact, not at the row** (estate rule: trust the rendered artefact):
`https://websitepromotion.co.uk/blog/website-launch-promotion-checklist.html` → HTTP 200, 80,415 B,
carries `checklist__item` / `checklist__body` / `checklist__footnote` markup and 48 `<li>`.
**Control:** an invented path on the same domain → 404, so the probe could have come out otherwise.

### Why this matters more than the count

The owner's 2026-08-31 question was *"why we didn't use infographics to take the place of much of
the explanatory copy"*. **Route B is that, and it is happening**: a tools comparison as a real
`<table>` on seotools.co.uk, a regulation process as a numbered flow on advertise.co.uk, a launch
checklist inside a blog article on websitepromotion.co.uk. Four consecutive sessions searched for
the answer in route A and found silence, because the word "infographic" names route A only.

⚠ **This is a landmine, filed as one** (LANDMINES.md, footprint `site_plan_imagery`): querying
`kind='infographic'` measures one of two routes and returns a confident, correct, useless number.

> ### ⚠ UPDATED 2026-09-04 (evening) — THERE IS A **ROUTE C**, and §2's two-route model is incomplete
>
> **Route C: structured markup inside the prose blob, driven by a `content_direction` spec rule.**
> `[MEASURED 2026-09-04]` gamesdesign.co.uk: `<table>` in **13 of 13** article bodies (100%) vs **3
> across the other 368** (0.8%); cause is that site's own spec — *"never describe a sequence of steps
> purely in prose when a table would make it scannable"* + a `list_usage` field. **It needs no
> component, no plan row, no pipeline and no composition — it is a spec field**, which makes it the
> cheapest lever in the estate for the owner's ask.
>
> **It also inverts the durability premise this lane and two others reason from:** route C's artefact
> does not survive a wholesale rewrite — it is **RE-DERIVED**, because the rule is still in the spec.
> Durability by regeneration, not preservation.
>
> **And §2's own arm-B query is blind to it** (no `page_components` row), exactly as route-A queries
> are blind to route B. Held at **n=2**: a second instructing site gets 1 of 8, unexplained, and the
> disconfirming read (are its 7 tableless articles *correctly* tableless?) is **unmade**.
> Full account: NOTES §11. Landmine corrected the same day.

## 3. The contradiction, stated as a specification defect and NOT as a cause

Three of the estate's own rules disagree about who may carry a number:

| source | rule |
|---|---|
| **live planner prompt** (`agent_definitions` `f263eaa1…`, mig 718, 2026-09-02) | *"an `illustration` for a concept, process or scene, an **`infographic` for numbers**, comparisons or steps"* |
| **same prompt**, exemplar commentary | both worked entries *"keep all wording out of the image (headings and labels are set in HTML beside the graphic)"*; the infographic exemplar's own prompt ends *"no text anywhere in the image"* |
| **register IMG-046** (design decision D1, imagery lane) | *"`infographic` stays decorative-Banana and **must never carry real numbers**"* |
| **VIZ-005 / `features_open/023` R4** | diffusion is the wrong tool for any value that must be exact, selectable, translatable or screen-reader accessible; *"Go emits SVG from real values"* |

So the live instruction assigns **numbers** — the only trigger unique to `infographic`, the other
two (`comparisons`, `steps`) overlapping `illustration`'s `scene`/`process` — to the mechanism that
two written rules forbid from carrying numbers, in a form (no wording in the image) that could not
state a number even if permitted.

> **⚠ THIS IS NOT OFFERED AS THE CAUSE OF THE 1-ROW COUNT, and the distinction is load-bearing.**
> Three sessions in a row built a causal account of that count and each was retracted
> (`WRONG_CALLS.md` 2026-09-04, three entries). The count has **no** explanatory power: the
> `framework_prompts_positive_voice` lane established, and this lane re-verified first-hand
> (RUNBOOK §2), that the **21** sites holding a current plan *and* non-empty `evidence_base.facts`
> and the **7** sites that have planned imagery since 718 are **disjoint sets**. No site capable of
> producing an infographic has been planned since the instruction landed. **The instruction has
> never been exercised.** What §3 describes is a defect in the *specification*, discoverable by
> reading, which will be waiting when the test is finally run — not an explanation of the silence.

## 4. THE RESPONSIBILITY SET — where this lane fits

Nine lanes touch infographics. **The rule of thumb: they own MECHANISMS, this lane owns the
SELECTION RULE between them.** Every boundary below was read from that lane's own current docs.

| lane | owns | boundary with this lane |
|---|---|---|
| **`imagery`** (`docs024/imagery/`, IMG-001…080) | the generation machine: the `kind` enum, provider routing (`routing.go:63` → Banana), style guides, asset deploy. **The five-place new-kind checklist (IMG-031) is theirs.** | They own *how a picture gets drawn*. This lane owns *whether a picture is the right artefact and what it may contain*. **Any change to the `kind` enum goes through them, not around them.** |
| **`inline_guide_imagery`** | per-section binding (IMG-075): a figure attaches to one section and survives a wholesale rewrite. Proved end-to-end 09-03 on grip-styles. Their own summary: *"the ask has three layers and this lane owns only the top one."* | They make a figure **stick**. This lane decides **which figure**. Their durability property is a precondition for anything this lane recommends inside an article. |
| **`editorial_design_uplift`** | page composition and structure — feature 035, `recomposeAncestors`, card structures; the 189-page orphaned-hero census; the "one prose slab plus chrome" finding. | They own **how a page is composed of sections**. This lane owns **what fills an explanatory section**. Their 09-02 handoff §2 (`DO NOT re-propose a component-level image field for article-body`) binds this lane too. |
| **`framework_prompts_positive_voice`** | the planner/writer **prompt bytes** (migrations 641, 718). RFC_016 §5.2 — the owner reads the exact bytes before they apply. Their standing recommendation on this question is *"change nothing until one of the 21 has been planned"* and this lane **agrees**. | **This lane writes the specification; that lane cuts the migration.** This lane must never cut a planner migration. Note mig 729 (`bugfix_450`) has anchors pinned on this same prompt awaiting an owner permission decision — a real cost to any edit. |
| **`brochure_component_library`** + the **VIZ register** | the code-rendered components themselves — `evidence-chart`, `evidence-timeseries`, `mechanism-flow`, and (2026-08-24) `checklist` / `period-calendar` / `comparison-table`. VIZ-007/009/011 constraints. | They **build** route B's components. This lane **routes work to them** and owns the rule that says when a need is route B rather than route A. §5 owes them a correction: VIZ-017's *"UNEXERCISED"* is stale. |
| **`bugs_open/114`** (owned, active) | generated imagery that is deployed and never referenced — the resolution/render gap; 189 pages / 21 sites. | Strictly **downstream**. They repair pictures that exist and do not render. This lane is upstream of whether the picture should have been asked for. **Contribute into 114, never file a competing bug.** |
| **`experience_loop`** | detectors over the built site (listing-class, experience-promise, rule D), each with a fleet ground-truth pass. | The natural home for *"this section's subject is a sequence/comparison/threshold and it shipped as prose"*. **This lane would specify such a check; that lane owns building and running detectors.** Do not build a competing detector here. |
| **`dartsonline_traffic`, `finetuning_uk_service`, `agritec_uk`, `site_delivery_and_editor`, site lanes** | their own sites and pipelines. | **Consumers.** finetuning.uk is the estate's only site with 10 registered facts and is the named canary — but it has **0 `site_plans` rows** and owner-approved copy, so a run there is a materially bigger act than "just plan it". Not this lane's call. |
| **`experience_register`** | approved experience patterns. | Adjacent; no live overlap found. Revisit if a graphic-vs-prose rule becomes a registered pattern. |

**What this lane will NOT do**, stated so peers can hold it to this:
cut a planner-prompt migration · change the `site_plan_imagery` kind enum · build a component ·
build a detector · dispatch a build at another lane's site · file a bug that duplicates 114.

## 5. Phasing

**Phase 0 — establish and correct the record (this session).** Standing five created. One landmine
filed (route A ≠ the whole question). Register corrections owed: **VIZ-017**'s *"Live, but
UNEXERCISED: no page has yet been built with any of them"* is stale — `checklist` 9 / 3 sites,
`comparison-table` 7 / 4, `period-calendar` 2 / 2, all first used **after** that entry was written;
and **IMG-046**'s "never carry real numbers" needs a pointer to the live prompt that contradicts it.
Correct the entries visibly, per the register-status landmine.

**Phase 1 — the one observation everyone agrees is next, and nobody has made.** Plan **one of the
21** eligible sites (current plan + non-empty `evidence_base.facts`) and record which artefact the
planner reaches for on a genuinely numeric section. This is the only measurement that separates the
live accounts. It needs a site whose copy is not owner-approved — the selection is the work.
**Pre-register the prediction before running it** (§6).

> **PHASE 1 UPDATE 2026-09-04:** `dartsonline.com` **is** one of the 21 (current plan + 9 facts) and now
> carries a live owner ask — but its 9 facts are **PDC tour calendars and news events, none about
> equipment**, so it cannot serve as the *numeric-comparison* test even though it is eligible.
> **Eligibility and suitability are different filters and the 21 was only the first.** A Phase 1 site
> needs a current plan, non-empty facts, facts ABOUT something a section would compare, and copy that is
> not owner-approved. Assessment: `ASSESSMENT_2026-09-04_dartsonline_guide_family.md`.

**Phase 2 — the specification, if Phase 1 warrants it.** A written rule assigning *quantity /
comparison-of-quantities / ordered-sequence-to-follow* to route B and *concept / process / scene* to
route A, with a worked exemplar for each that demonstrates the **unique** trigger rather than the
overlap. Handed to `framework_prompts_positive_voice` as bytes for the owner to read. Coordinated
with `bugfix_450`'s pinned anchors.

**Phase 3 — the detector, specified here and built by `experience_loop`:** a section whose subject
is a sequence, a comparison or a set of thresholds that shipped as prose.

## 6. Pre-registered prediction for Phase 1

Recorded now so it cannot be fitted afterwards. On the first eligible site planned:

- **P1:** the planner emits **at least one** section-scope entry for a numeric section. *(If it
  emits none at all, the overlap/wording accounts are both irrelevant and the gate is elsewhere.)*
- **P2:** that entry's `kind` is **`illustration`**, not `infographic` — the disjunction reading.
- **P3:** if it *is* `infographic`, its `prompt` will **either** name no figure (obeying the
  no-wording rule, hence not actually a numbers graphic) **or** bake a figure into the image
  (violating IMG-046). **Both outcomes confirm §3's contradiction; a prompt that does neither
  refutes it.**

## 7. Open questions this lane holds

1. Which of the 21 eligible sites is safe to plan? (Copy provenance is the constraint, not capability.)
2. Does `component_expresses` surface route B's six components to the planner in a way that makes
   them reachable for an explanatory need — or are they chosen only when a page type already implies
   them? **Not yet read. Do not assert either way.**
3. Is the 09-02→09-04 route-B inflection caused by migration 718, by VIZ-017's three components
   landing, or by the 641/planner work? Three candidates, none tested, and they are confounded by
   date. **Named as a question, not a finding.**
4. Does route B reach *article body prose*, or only whole sections? `editorial_design_uplift`
   measured 0 of 360 `article-body` pages with a non-chrome section — yet 9 route-B instances sit on
   `blog-post` pages. **These may be different populations; check before quoting either.**
