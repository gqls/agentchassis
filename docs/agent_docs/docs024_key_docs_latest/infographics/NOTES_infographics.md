# NOTES — infographics

Full path: `docs/agent_docs/docs024_key_docs_latest/infographics/NOTES_infographics.md`
Append-only, **newest at the bottom**. Evidence, commands, what the system actually said — and every
misstep, dead end and wrong turn, including this lane's own earlier claims that turn out false.

---

## §1 — 2026-09-04, lane opened: the corpus survey

Owner: *"search the code and documentation and find everything we've discussed and done with
infographics … become the main thread … determine where you fit in the responsibility set."*

**Corpus size:** `grep -ril infographic` over the repo → **206 files**. Concentrations:
`docs024/imagery` 17 · `docs/leopardessconsulting` 12 · concept-register extractions 10 ·
`docs024/imagery/old` 8 · `finetuning_uk_service` 7 · `agritec_uk` 7 · `platform/orchestration/actions` 6 ·
`sql_for_agents` 6 · register 6 · `editorial_design_uplift` 5 · `inline_guide_imagery` 5 ·
`brochure_component_library` 5.

**Every layer of route A exists and is wired** (confirmed at the code, not from the CONTRIB that
claimed it):

| layer | file:line |
|---|---|
| generation config block | `generate_image_actions.go:101` |
| style-guide handling (photographic kind) | `imagery_style_guide.go:316` |
| provider routing → Banana | `internal/adapters/imagegenerator/routing.go:63` |
| stability provider entry | `stability/provider.go:68` |
| plan admission (`validImageryKinds`) | `write_site_plan_action.go:213` |
| section consumption | `plan_sections_action.go:563` — `spi.kind IN ('illustration','icon','infographic')` |

`plan_sections_action.go:563` takes all three kinds in **one query** and the scan loop that follows
does **not branch on kind**. So there is no component-capability gate for infographics — a fact the
`framework_prompts_positive_voice` lane hypothesised, then refuted itself, and which is re-confirmed
here at the cited line rather than taken on trust.

## §2 — the measurement that reframed the question

Route A census (RUNBOOK §1, arm A), `[MEASURED 2026-09-04]`:

```
 kind         | rows | sites
 hero         |  436 |    35
 icon         |  219 |    30
 logo         |   54 |    34
 illustration |   32 |     7
 infographic  |    1 |     1
 sprite_sheet |    1 |     1
```

Compared against the same census in the `editorial_design_uplift` CONTRIB of 2026-08-31 (hero 359,
icon 196, illustration 19, infographic 1): **every kind grew except `infographic`, which is
unchanged at 1 across four days and 77 new heroes.** The one row is
`infographic_decision_engine`, `section` scope, `scorecard-simulator:1`, **mortgagecalculator.co.uk**,
created 2026-08-02.

**Then I ran the arm nobody had run** (RUNBOOK §1, arm B):

```
 name                | instances | sites | first_use  | last_use
 mechanism-flow      |        14 |    10 | 2026-07-28 | 2026-09-04
 evidence-chart      |        10 |     5 | 2026-07-28 | 2026-09-03
 checklist           |         9 |     3 | 2026-09-02 | 2026-09-04
 comparison-table    |         7 |     4 | 2026-09-02 | 2026-09-03
 evidence-timeseries |         3 |     3 | 2026-07-29 | 2026-08-20
 period-calendar     |         2 |     2 | 2026-08-31 | 2026-09-04
                      ---- 45 instances / 17 distinct sites ----
```

17 domains: advertise, copyonline, cv1, dartsonline, designblog, farmerinsurance, fundamentallyai,
homegarden, lendzy, leopardessconsulting, loanzy, mortgagecalculator, oufe, remortgagecalculator,
robot-hands, seotools, websitepromotion. **That spread is the control against "one lane seeded
examples by hand."**

Adoption by day — the inflection is recent and sharp:
`≤3/day through August · 4 on 09-02 · 15 on 09-03 · 9 by midday 09-04`.

Page types: `content` 22 · `landing` 12 · **`blog-post` 9** · `entity-directory` 1 · `section-index` 1.

**Served-artefact check with a control** (RUNBOOK §4): websitepromotion.co.uk's launch-promotion
checklist article → HTTP 200, 80,415 B, `checklist__item`/`__body`/`__footnote` markup, 48 `<li>`;
invented sibling path on the same domain → **404**. The probe could have failed and did not.

**Route A : Route B = 1 : 45.** Four sessions searched route A for the owner's answer. Route B is
the answer and carries a different name, so no route-A query could ever have found it.

## §3 — misstep: two schema assumptions, both wrong, both caught by the error rather than by a check

Writing the arm-B query I assumed `content_components.site_id` (to separate global from per-site
components) and `page_components.deleted_at` (to exclude soft-deleted rows). **Neither column
exists.** Postgres errored both times, so the cost was two round trips and nothing worse.

Worth recording because the *silent* version of this is the dangerous one: had I "fixed" the second
by dropping the predicate without reading `\d page_components`, I would have quietly changed the
population and had no way to know. CLAUDE.md's *"schema first: `\d <table>` before writing SQL"*
exists for exactly this and I skipped it. Cheap check that would have caught both: one `\d`.

## §4 — the specification contradiction, and why I am NOT calling it a cause

Read first-hand from the live row (RUNBOOK §3), `[MEASURED 2026-09-04]`, `f263eaa1…`, 39,431 B:

- Sections bullet: *"an `illustration` for a concept, process or scene, an **`infographic` for
  numbers**, comparisons or steps"*.
- Exemplar commentary: both worked entries *"keep all wording out of the image (headings and labels
  are set in HTML beside the graphic)"*; the infographic exemplar `infographic_selection_steps` ends
  *"no text anywhere in the image — headings are set in HTML beside the graphic"*.
- Rule 13 requires *"at least one section-scope `illustration` **or** `infographic`"* — a
  disjunction, `illustration` named first.
- Rule 16: one image per entry.

Against `register/imagery.md` **IMG-046** (design decision D1): *"`infographic` stays
decorative-Banana and **must never carry real numbers**"* — and VIZ-005 / `features_open/023` R4:
diffusion is wrong for values that must be exact, selectable, translatable or accessible.

**So `numbers` — the only trigger unique to `infographic`, since `comparisons`⊂`scene` and
`steps`≡`process` — is assigned to the mechanism two written rules forbid from carrying numbers, in
a form that forbids the wording a number would need.**

> **⚠ I am deliberately NOT presenting this as the explanation of the 1-row count**, and the
> restraint is the point. Three sessions built causal accounts of that count on 2026-09-04 alone —
> "the prompt says sparingly" (stale by two days), "no component can display one" (refuted at
> `plan_sections_action.go:563`), "the sites have no facts to draw" (void: the grouping variable was
> constant, all seven sites had zero facts). All three retracted; `WRONG_CALLS.md` carries them.
>
> The count **cannot** be explained, because the mechanism is undriven: the 21 capable sites and the
> 7 planning sites are **disjoint** (re-verified first-hand, RUNBOOK §2 — 21 and 0). §4 is a defect
> found by **reading the specification**, which is a different kind of claim from one found by
> fitting an observation, and it is disconfirmable independently (PLAN §6, P3).

## §5 — the register entries this lane found stale

- **VIZ-017** (2026-08-24) says the three generic structured components are *"Live, but UNEXERCISED:
  no page has yet been built with any of them."* `[MEASURED 2026-09-04]` `checklist` **9 instances /
  3 sites**, `comparison-table` **7 / 4**, `period-calendar` **2 / 2** — first uses 2026-08-31 →
  09-02, i.e. **after** the entry was written. The entry was true when written and is now the
  register's own documented failure mode (a status line outliving its truth while seats read it as
  ground truth). Correction owed, visibly, in the entry.
- **IMG-046** needs a pointer to the live prompt that contradicts its "never carry real numbers"
  rule. Not a correction — the rule may well be right and the prompt wrong — but a reader of either
  one alone currently cannot see that the other exists.

## §6 — what I did not check, marked so it is not mistaken for a finding

- `[UNMEASURED]` Whether `component_expresses` surfaces route B's six components to the planner in a
  way that makes them reachable for an *explanatory need*, or only when a page type implies them.
  **This is PLAN §7.2 and I have not read the derivation.** Do not assert either way.
- `[UNMEASURED]` What caused the 09-02→09-04 route-B inflection. Three candidates (mig 718; VIZ-017's
  components landing; the 641 planner work) and they are **confounded by date** — all three land in
  the same window. Naming them is not testing them.
- `[UNMEASURED]` Whether route B reaches article **body prose** or only whole sections.
  `editorial_design_uplift` measured 0 of 360 `article-body` pages with a non-chrome section, yet 9
  route-B instances sit on `blog-post` pages. **These are plausibly different populations** — a
  `blog-post` page need not carry `article-body`. Check before quoting either figure against the
  other; I have not.
