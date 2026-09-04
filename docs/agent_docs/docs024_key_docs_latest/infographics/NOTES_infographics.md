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

## §7 — 2026-09-04, later: the "use sparingly" line, and why three lanes kept quoting a dead sentence

Owner relayed the finetuning lane's account: *infographics don't get planned so they don't exist; a
line says "use sparingly" and it seems to have been read as none.*

**The first half is right and I said so.** Route-A infographics planned since 718: **0**. And under
the old wording the estate produced exactly **1** in all history — so as an account of how we got
here, it fits the record precisely.

**The second half expired on 2026-09-02.** I ran the check wider than anyone had — *deliberately, to
test whether they might be reading a DIFFERENT live prompt, which would have made them right and the
prompts lane wrong*:

```sql
SELECT type, id, is_active, COALESCE(is_snapshot,false) FROM agent_definitions
 WHERE default_config::text ~* 'sparingly' AND deleted_at IS NULL;   -- → 0 rows
```

**Zero, fleet-wide**, across active/inactive/snapshot/undeleted. There is no such prompt.

**THE MECHANISM, which is the finding.** `docs/agent_docs/sql_for_agents/053_build_site_planner.sql`
— the seed named after the planner — still carries the superseded sentence at lines **1347, 1652,
2058**. Migration 718 edited the **live row, not the seed**. So grepping the repo for the planner's
imagery rules, the obvious first move, returns dead text in triplicate under the canonical filename
with nothing marking it stale. `[MEASURED 2026-09-04]` that one sentence has been quoted as live
evidence **three times in two days by three lanes** and reached the owner **twice**. Banner added to
053 (pure addition, prompt text untouched); landmine filed.

**THE SECOND DOOR, and it is why it reached the owner.** The finetuning lane's NOTES took the
correction on 09-04. Their `README_where_we_are.md` did not — line ~1531 still stated it in plain
prose, with **no correction in the following 74 lines**, and that is the file the owner reads.
**A correction in NOTES does not discharge a claim made in README: different files, different
readers, and the README's reader is the one who acts.** Dated attributed correction appended there;
that lane has since confirmed they want it kept as placed.

> **⚠ MY OWN COUNT WAS THE STALE ONE IN THIS EXCHANGE.** I reported `comparison-table` 7 / route B 45;
> they measured **8 / 46** and were right — a new instance landed at **13:08:47Z**, after my reading.
> Route B moved twice inside one working session. **Any route-B figure in this lane's docs is a
> snapshot with a shelf life measured in hours, not days** — re-run RUNBOOK §1 before quoting one.

## §8 — first real exercise of the selection rule, and it changed a peer's component choice

Owner decision relayed via the finetuning lane, verbatim: *"I still want infographics, we can use them
on the pages and in the cards - maybe simple ones for the cards."* plus *"Three-steps section: a
dedicated section and diagram on the homepage please."* Three asks, all on finetuning.uk/index.html.

**Ask 1 — three steps → `mechanism-flow`. Endorsed as proposed.** Right for the reason it is safe:
VIZ-006 records it has **no numeric field by design** — on an evidence-gated site the absence of the
slot is the control. New section, no approved copy at risk.

**Ask 2 — £99 vs ~$5,000. They proposed `comparison-table`; I recommended `evidence-chart` instead.**
Read both schemas first-hand rather than reasoning from names:

| | `comparison-table` | `evidence-chart` |
|---|---|---|
| where the figure comes from | `rows` are `source: "llm"` **free text** | `facts` ← `site_specs.evidence_base.facts`, points resolve by `fact_id` |
| if the register is missing | renders anyway | `on_missing: skip_section` — *"a chart with no audited series must not be drawn"* |
| what the LLM may write | cells, incl. figures given in-prompt | *"NEVER state, round, summarise or preview a number"* |

comparison-table's own guidance concedes it: *"a figure typed into a text cell publishes a false
claim on a live site just as surely as one in a numeric field"*, and VIZ-017 says the claims risk
there is **"STATED, NOT SOLVED"**. So the two registered figures would have been model-retyped text
that merely matched the register. **This is the selection rule doing the one job the lane exists
for**, and it fired on the first case.

**Three data gaps found, all on their evidence_base, none needing code** `[MEASURED 2026-09-04]`:

1. `facts: 10` but **no `charts` key at all** (`charts: 0`). `evidence-chart` requires it → mounting
   it today **skips the whole section silently**. No error; the section simply is not there.
2. Neither fact carries a **`display`** key (keys: id, kind, claim, value, source, tolerance,
   verified_at, writer_line, context_terms). The template renders `{{if $f.display}}{{$f.display}}`.
3. **`tolerance` is stored and the template NEVER READS IT** — grepped the live template for
   tolerance/approx/circa/band: **zero matches**.

⇒ The owner's *"the approximate side must read as approximate"* is satisfiable in exactly one place:
`ft-market-anchor.display = "~$5,000"`. **The approximation then lives in the register, not in
anyone's memory** — which is the same structural move as VIZ-006's missing numeric field.

**Ask 3 — cards. No fact-resolved card-scale component exists.** `stat-band` (3 placements) is
nearest and its `stats` array is `source: "llm"`; its *"must be a REAL, EVIDENCED number; NEVER
invent"* is a prompt instruction, not a control. **Recommended instead: cards carry NON-NUMERIC
graphics** — and route A is legitimate there, because IMG-046 forbids a generated image *real
numbers*, not decoration. So the two routes divide cleanly by content, not by prestige: **figures go
to the fact-resolved component; a picture with no figures in it may be a picture.** Filling the
card-scale gap properly is `brochure_component_library`'s call; specification offered, not built.

`[UNMEASURED]` Whether `evidence-chart` renders acceptably with only **two** points — every live
instance I sampled has more. Flagged to them, not tested.

## §9 — 2026-09-04: the finetuning lane caught a defect in MY recommendation, and it is a gap in the selection rule itself

They implemented §8's advice, and found the thing neither of us had raised:

> **THE TWO FACTS ARE IN DIFFERENT CURRENCIES.** £99 and $5,000 cannot share a like-for-like bar
> axis, and no exchange rate is a registered fact on that site.

**They are right and this is my miss, not theirs.** I recommended `evidence-chart` because it
resolves **provenance** — each figure comes from the register, the model cannot type a number. That
is true, and I treated it as sufficient. It is not.

**The general form, which is the part worth keeping:**

> **Fact-resolution guarantees PROVENANCE, not COMMENSURABILITY.** Two facts can each be perfectly
> registered, verified, and correctly displayed and still not belong on a shared axis. `evidence-chart`
> has `max`/`max_fact_id` to scale bars against one another and **no notion of whether the two
> quantities are the same kind of thing.** The bars then assert a ratio that no fact in the register
> makes and no gate reads — the claims gate checks the *values*, and the defect lives in the
> *geometry*. `[MEASURED 2026-09-04]` the drawn ratio is ~50:1; at ~0.79 the true one is ~40:1, so the
> picture overstates by about a quarter while every individual figure on it is impeccable.

**And it extends past charts:** the same hole waits wherever a renderer derives a *relationship* —
ratio, rank, share, trend — from independently-registered facts. The inputs are audited; the derived
relationship never was; and the relationship is what the reader takes away. That belongs in this
lane's selection rule, not only in the component's docs.

**Their question, answered at the template** `[MEASURED 2026-09-04]`: per-point fields are exactly
`fact_id`, `label`, `tone`. **There is no per-point unit.** `unit` is chart-level and is appended to
every value —
`{{if $f.display}}{{$f.display}}{{else}}{{printf "%.10g" $f.value}}{{end}}{{$c.unit}}` — so on a
mixed-currency chart whose displays already carry symbols, setting it renders **`£99$`** and
**`~$5,000$`**. ⚠ **The obvious fix corrupts both labels; `unit` must stay empty here.**

So the caption is the only in-component remedy and their call was right. **But a caption is prose and
prose is not a control** — the bars still assert 50:1 to a reader who looks at the picture. Ranked
remedies, recorded for the next case: (1) same unit, proceed; (2) register the conversion as its own
audited fact and chart the converted figure; (3) do not share the axis. They have shipped the honest
version of (3)-by-prose; **(2) has not been put to the owner and should be, once he has seen it
rendered.**

**Their reason for the cards decision is better than mine and I have taken it.** I argued from
IMG-046 (a generated image may be decorative, may not carry real numbers). The uplift lane's reason,
which they relayed: text inside `<svg>` is in `nonAssertionElements`, so **a figure inside a
decorative graphic routes around the claims gate entirely** (VIZ-009). That is a structural argument
where mine was a permissions one.

`[UNMEASURED]`, pre-registered with them before the build lands: (a) whether `evidence-chart` renders
acceptably with only **two** points — every live instance sampled has more; (b) whether the caption
survives into the served bytes.

**Filed:** LANDMINES.md, *"`evidence-chart` guarantees PROVENANCE, not COMMENSURABILITY"*, footprinted
on the chart fields and dispatched to the verifier. Owed: tell `brochure_component_library` — a
per-point unit, or a refusal to share an axis across mixed units, is their component's call and this
is the first evidence for it.

## §10 — 2026-09-04: a boundary honoured, and one observation kept from it

The finetuning lane reported that the owner had said, unprompted and without any diagnosis:
*"Case studies page is missing a hero."* Measured by them: the hero slot renders no image at all
while the site holds a deployed `content-hero-case-studies.jpg` that nothing displays — IMG-077 item
`6db67bde`'s **unwired** state.

**Routed to `bugs_open/114`, not taken.** `who-owns.py` shows that lane active with seven commits on
the bug file on 09-03 alone, and PLAN §4 puts 114 strictly downstream of this lane with "contribute,
do not compete". They have first-hand evidence; a second-hand relay from me would be worse. Offered
to carry it if they prefer.

**Why it is not this lane's even though it is imagery:** a hero is **chrome**. This lane owns the
choice of artefact for something a section must **explain**. A slot whose content was never in
question failing to resolve is a different defect with a different owner. Recording the reasoning
because the boundary is only worth anything if it holds on a case that is adjacent and tempting.

**The observation worth keeping, which IS this lane's kind of thing:** *a human reported from the
outside what a detector had already filed.* IMG-077 had that page in its census. The finding existed.
The route it travelled to attention was the owner looking at the page. That is a fact about detector
**reach**, not detector correctness — and it is the same shape as this session's opening finding,
where a correction sat in a NOTES file while the claim the owner acted on sat in a README.

**Two instances in one day, in unrelated subsystems, of a true finding failing to reach the person
who would act on it.** `[UNMEASURED]` whether that is a pattern or a coincidence — naming it, not
claiming it. It bears on PLAN Phase 3: a detector for "this section should be a diagram" is worth
building only if its findings reach somebody, and neither instance today did so by its own route.
Flagged for `experience_loop` when Phase 3 is specified, since detector reach is their domain.
