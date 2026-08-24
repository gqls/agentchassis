# 381 — the site planner composes pages from components that CANNOT EXPRESS the page it just planned, so a "month by month" page ships as four prose blocks and the whole site has zero lists and zero tables

**Filed 2026-08-24** by the `loanzy_uk_example_site` lane from the completed greenfield
`garden-tools.uk` build, on the owner's review of the served pages. **Status: OPEN, unowned.
Severity: MEDIUM-HIGH** — it degrades every page of every site quietly, and the result reads as
"the writer is verbose" rather than as a defect.

> **2026-07-31 ruling — first-hand verification substituted for `090`.** The cause is read from live
> `content_components.html_template` and the live `page_components` join for the affected pages,
> quoted below, plus a fleet-wide count. The claim is structural but not inferential.

## 1. The symptom the owner named

Two complaints that turn out to be one defect:

- **"A wall of text."** The 300-word paragraph *"What we check before a tool earns a recommendation…"*
  on `how-we-assess`, with no subheads, no list, no emphasis.
- **"The seasonal planner says *'What your shed needs, month by month'* but there is no calendar and
  no month by month list."**

## 2. Measured on the served site `[MEASURED 2026-08-24]`

Across **all seven** pages that built, the *content* contains:

| | count |
|---|---|
| `<table>` | **0** |
| content `<ul>`/`<ol>` | **0** (the single `<ul>` per page is the nav — verified on the 56-word `contact` page) |
| `<strong>` | **0** |
| `<p>` | 1–15 per page, longest single paragraph **104 words** |

`how-we-assess` is 1,486 words, 14 paragraphs, zero lists, zero tables, zero emphasis.
`seasonal-planner` promises *month by month* under an `<h2>` and delivers four `<h3>` prose blocks
(Spring/Summer/Autumn/Winter); **three month names appear on the whole page**, all incidental.

## 3. The cause, and it is COMPOSITION — not the writer, and not a designer

`seasonal-planner`'s four components, and what their templates can express:

| component | `<ul>`/`<ol>`? | `<table>`? |
|---|---|---|
| `hero` | no | no |
| `generic-text-block` | no | no |
| `info-card-grid` | no | no |
| `call-to-action` | no | no |

**A page whose own heading promises a month-by-month calendar was composed entirely from components
that cannot render one.** The writer had nowhere to put twelve months, so it wrote four seasons in
prose. That is the page degrading to fit its container, and it is invisible downstream: the sections
rendered, the page deployed, every check passes.

**The vocabulary EXISTS — the planner simply did not reach for it** `[MEASURED 2026-08-24]`:
```sql
SELECT count(*) AS active_section_components,
       count(*) FILTER (WHERE html_template LIKE '%<ul%' OR html_template LIKE '%<ol%') AS can_list,
       count(*) FILTER (WHERE html_template LIKE '%<table%') AS can_table
FROM content_components WHERE is_active AND component_level='section';
-- 151 total | 34 can_list (23%) | 10 can_table (7%)
```
~~**34 list-capable and 10 table-capable components were available and none was chosen**~~
`build-site-planner`'s prompt lists available components by `name`/`display_name`/`function`/
`description`; **nothing in that listing states what markup a component can produce**, so the planner
is choosing blind on expressiveness. That part stands.

> **⚠ CORRECTED 2026-08-24 by the `bugs_open/381` session, verified here. The "44 sat available"
> premise is TRUE AND MISLEADING, and the misleading half is mine.** Enumerated, the 34 list-capable
> components are **all special-purpose**: `model-directory`, `mortgage-lender-directory`,
> `savings-provider-directory`, `health-insurer-directory` (+ their `-listing` twins),
> `adoption-tracker`, `protocol-tracker`, `loans-application-tracker`, `ai-readiness-quiz`,
> `archetype-taster-quiz`, `mortgages-repayment`, `mortgages-simple`, `pricing`, `product-details`,
> `hero-card-carousel`, `swipeable-insight-carousel`, `content-sidebar` — and **`site-footer`**,
> which is chrome and should never have been in a count of composable page vocabulary.
>
> **There is NO generic checklist, steps, table, or calendar component.** So a planner told what each
> component can express would still have **nothing to compose a seasonal planner from**. My §6
> candidate 1 ("tell the planner what each can express") is therefore **necessary but not
> sufficient** — on its own it converts a blind choice into an informed refusal. The missing piece is
> a generic structured-content component to choose. **Counting capability without checking
> SUITABILITY is the error**: 44 is a real number that answers a question nobody asked.

## 4. A second, separable arm — and it is the cheap one

`generic-text-block` is a **pass-through**: its template is
`<div class="section__content">{{.content}}</div>`, so it will render any HTML it is given. **The
owner's wall-of-text paragraph is in this component** (verified: it is the only one on
`how-we-assess` whose `content_data` contains that text). The writer put
`<p>…</p><p>…</p>` into a slot that would have accepted `<h3>`, `<ul>` and `<strong>` unchanged.

So the defect has two arms with different fixes, and **the string "wall of text" cannot tell them
apart** — the same discipline `bugs_open/206` arrived at for `no sections ready to build`:

- ~~**(a) WRITER-side, cheap.** A pass-through component given only paragraphs. Fixable today by
  telling `page-content-writer` to use `<h3>`/`<ul>`/`<strong>`…~~
  > **⚠ CORRECTED 2026-08-24 by the `bugs_open/381` session, schema verified here.** The writer was
  > **not free to choose**. `generic-text-block`'s `input_schema` types `content` as
  > **`{"type": "text", "source": "llm", "required": true}`**, and the writer's RULE 9 forbids HTML
  > in `text` fields. So "the writer simply chose prose" is wrong — **it was instructed not to emit
  > markup**, and the pass-through template downstream is irrelevant to that decision. Their fix
  > retypes five prose slots to `html` and rewrites RULE 10 (migrations 594/595); the `rich_text`/
  > `content` types RULE 10 already describes are **declared by zero components**, so the rule
  > documents a capability nothing uses.
  >
  > **One empirical wrinkle worth having, which cuts slightly against the tidy version:** the field
  > is `type: text` and its stored value **already contains HTML** — `{"content": "<p>Some of the
  > links on this site…</p><p>…"}`. So `<p>` is passing through a text-typed field today. Either the
  > type is unenforced, or the rule is understood as "no *structural* markup". **Worth confirming
  > before assuming the retype alone changes writer behaviour** — if the constraint is a prompt rule
  > rather than an enforced type, the prompt is the load-bearing half and the retype is permission
  > the writer may not notice it has been given.
- **(b) COMPOSITION-side, structural.** `differentiators`, `faq`, `hero`, `info-card-grid` hard-wrap
  in `<p>` and contain no list or table markup at all. **No writer instruction and no designer pass
  can add a list to these** — the markup is not in the template. Fixing this means either the
  planner picking expressive components, or those components gaining the markup.

## 5. ⚠ On the owner's `vigilant_designer` hypothesis — wrong layer, and the agent does not exist

The owner suggested this *"might be a job for the vigilant_designer thread/agent (through the
framework) or it might be a missing step in the workflow."*

**There is no `vigilant_designer`** `[MEASURED 2026-08-24]`. The live design agents are
`brand-designer`, `feature-designer`, `visual-designer` (all active, no snapshots). A name recalled
from memory is an `[UNVERIFIED]` claim wearing settled clothes — `LANDMINES.md` carries this exact
trap for `cmd/contrastscan`.

**And more importantly: a designer is the wrong layer for arm (b).** Design agents work on visual
presentation — palette, type scale, spacing, CSS. **They cannot add a `<ul>` to a page whose
components emit no `<ul>`.** Routing this to a designer would produce a better-looking wall of text.
The owner's second hypothesis — *"a missing step in the workflow"* — is the right one, and §4 says
which of two steps is missing depending on the arm.

## 6. Fix candidates, ordered by what closes the door

1. **Tell the planner what each component can EXPRESS (closes (b) at source).** Add markup capability
   to the component listing in `plan_site`'s prompt — even a crude `can_list`/`can_table` flag
   derived from the template. A planner that knows `seasonal-planner` needs a month list can pick one
   of the 34. Cheapest structural fix; no new components, no template edits.
2. **Writer guidance for pass-through slots (closes (a) today).** `page-content-writer` should use
   real structure in the 11 pass-through components. Independent of 1, ships immediately, and covers
   the paragraph the owner quoted.
3. **A compositional check: does any component on this page match the page's own promise?** A page
   headed *"month by month"* composed of four prose components is detectable. This is the general
   form and the most work.
4. **Not a fix: restyling.** See §5. Also not a fix: adding markup to `generic-text-block` — it is
   already a pass-through and is not the constrained one.

## 7. Related, and how this is distinct

- `bugs_open/206` — those pages did not build at all. **These built and shipped hollow**, which is
  worse in one respect: nothing anywhere reports a problem.
- `bugs_open/380` (same lane, same build) — ungated claims. Distinct defect, same root cause shape:
  the machinery that would catch it was never engaged.
- The owner's third point, **cards stacking on mobile**, is measured but NOT filed here: there is no
  carousel on this site (`scroll-snap` count **0**, no `.carousel`/`.slider` class). It is a CSS card
  grid that collapses to one column — `index` carries **14** cards in markup, `seasonal-planner` 12,
  `care` 8. "Restrict to 3 per component" is already true of the one identified grid; the wall comes
  from the NUMBER OF CARD SECTIONS per page, not cards per section. Needs its own decision about
  page composition before it is worth filing.

---

## 8. FIX BUILT, COUNCIL-APPROVED, NOT YET APPLIED — and what the diagnosis changed (`bugs_open/381` lane, 2026-08-24)

**Owner: the `bugs_open/381` session** (this file's filing lane marked it UNOWNED and keeps the
account of the site; the fix lane is
`docs/agent_docs/docs024_key_docs_latest/bugfix_381_inexpressive_composition/`).
**Status: LIVE. All five migrations applied and recorded 2026-08-24 16:53–16:58Z** (council
`ca400ba6`, round 1, APPROVED, three advisory objections, none high — all acted on). Config-only,
so no chassis roll was needed. Verified at the artefact, not at the exit code: every menu query
executes bound as the chassis binds it; the evidence-base gate discriminates in BOTH directions
(excludes exactly 2 rows on an evidence-less site, 0 on an evidence-bearing one); all four target
fields read back the LITERAL `html`; 304's markdown ban survived the rule-9 replace; and
`generic-text-block` moved from `[prose only]` to `{html-block,list,table}` in the planner's own
listing. Arm B was briefly held as `_HOLD.sql` pending another lane's Go fix and released the same
afternoon once `714789d7b` was proven an ancestor of the running chassis `70fd163c2`.
**What is NOT yet evidenced: no page has been BUILT under the new instructions** — the mechanism is
verified, the outcome is not. See §8f.

### 8a. Still valid, and it is not site-specific `[MEASURED 2026-08-24, live DB]`

Re-validated on live rows: 22 sections on `garden-tools.uk`, not one from a list/table-capable
template. **Fleet 30d: 741 pages / 29 sites; 327 (44%) contain no list, table or `<strong>`
anywhere in their content; 1,863 of 1,980 section placements (94%) used a prose-only template.**

### 8b. ⚠ THERE ARE THREE LAYERS, AND §4's ARM (a) NAMED THE WRONG LEVER

This file's amended §4 says the writer "was instructed not to emit markup" by RULE 9 and its
`text` type. That is closer than the original, and **it is still not the lever.** The disproof was
in the same table the whole time `[MEASURED 2026-08-24, 30d instances]`:

| field | declared type | `llm_guidance`? | instances rendering `<ul>`/`<ol>` |
|---|---|---|---|
| `article-body.content` | **`text`** | **yes** — *"…h3 for subsections, p for paragraphs, ul/ol for lists…"* | **116 / 153 = 76%** |
| `generic-text-block.content` | **`text`** | **none at all** | **12 / 173 = 7%** |

Same declared type, same `text/template` renderer (no escaping), same RULE 9 — **eleven-fold
difference.** So RULE 9 does not bind, and **the field's own `llm_guidance` is what produces
structure.** The retype to `html` keeps a smaller, honest job: the prompt prints each field as
`` `content` (TYPE, required) `` and addresses its rules BY TYPE, so the type is the **routing key**
between RULE 9 and RULE 10 — which matters because **RULE 10 names `rich_text`/`content`, and
`[MEASURED 2026-08-24]` ZERO components declare either** (940 llm fields are `text`, 2 are `html`).
The one rule permitting structure has been addressed to nobody for its whole life.

### 8c. `content_shape` was built for exactly §6-candidate-1 and is DEAD

Zero Go readers repo-wide; **omitted from the birth INSERT** (`store_generated_component_action.go:634`
writes 19 columns and not this one, so every generated component is permanently NULL); NULL on
128/151 active section rows; no CHECK, so the vocabulary has drifted (`series`, `sequence`,
`mixed`); and **12 rows declare `structured_list` while their template contains no list markup** —
reading it would tell you a prose component can render a list. **This is why the fix derives
capability instead of declaring it.** (Answers `TLIB-016`'s long-open verify-later; the register
entry is amended.)

### 8d. What shipped, and what it deliberately does not do

`component_expresses(html_template, input_schema)` — IMMUTABLE, read at query time by all three
planner menus (591–593), returning `html-block` / `list` / `table` / `items` / `{}`; plus a
planning rule tying page promise to section capability. `594` gives the prose slots real
`llm_guidance` (and retypes them); `595` re-addresses RULE 10 to `html` and narrows RULE 9 without
touching 304's markdown ban. **`content-gap-planner` is the busiest planner by 27× (749 calls/30d
vs 27), so most of the benefit lands there, not on the greenfield path this bug was found on.**

**It adds no component, and §7's promise-vs-delivery gap is therefore NOT closed.** Confirming this
file's own amended §3: the 44 structural components are all special-purpose, so an informed planner
still has nothing generic to compose a month-by-month page from. **The missing generic
checklist/steps/comparison-table/calendar component is open work and is the next thing anyone
picking this up should weigh.**

### 8e. The objection that nearly got through, recorded because the class recurs

The council's `bug_historian` seat asked whether the four templates already wrap `{{.content}}` in
a literal `<p>` — because if so, a writer emitting `<h3>`/`<ul>` would nest block elements inside a
paragraph: invalid HTML, browser-repaired inconsistently, **and invisible at every point we look**
(the DB row is schema-valid, no check fails, only the served page is wrong). I had verified the
renderer and the type checker and never asked what the *container* was.
**CLEARED `[MEASURED 2026-08-24]`: all four use a `<div>`.** The re-check query is in `594`'s header
because the RFC_032 lane is rewriting `html_template` fleet-wide and this fact has a shelf life.
⚠ `about-content` contains `<p>` elsewhere, so the predicate must test the **container of the
slot**, not the presence of `<p>` anywhere in the template.

### 8f. How to tell whether it worked, and what would refute it

Acceptance measure and its pre-fix baseline are in the lane's `RUNBOOK`; the promise-vs-delivery
check is the filing lane's `after_test.sh` (HANDOFF 24 §3), which fires on `seasonal-planner` and
stays silent on the other 11. **The falsifier: if `generic-text-block`'s structure share does not
move toward `article-body`'s 76%, then §8b's comparison was confounded by WHICH AGENT writes each
field** (`article-body` on the blog path, `generic-text-block` by `page-content-writer`) and the
prompt, not the guidance, is load-bearing. `595` hedges that by fixing the prompt half too, so the
two arms fail independently rather than together.
