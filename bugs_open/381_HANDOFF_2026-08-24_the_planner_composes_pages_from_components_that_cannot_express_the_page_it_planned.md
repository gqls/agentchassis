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

---

## 9. THE VOCABULARY GAP IS CLOSED — three generic components built and LIVE (2026-08-24, on the owner's instruction)

§8d recorded that the fix "adds no component, and §7's promise-vs-delivery gap is therefore NOT
closed", and named the missing generic vocabulary as open work. **The owner then asked for those
components. They are built, council-approved (`c134b0e9`) and LIVE.**

| component | expresses | for |
|---|---|---|
| `checklist` (604) | `items, list` | criteria, standards, what-we-check — real `<ul>` |
| `period-calendar` (605) | `items, list` | month-by-month, quarters, seasons — real `<ol>` |
| `comparison-table` (606) | `items, table` | options × criteria — real `<table>`, stacks on mobile |

`[VERIFIED 2026-08-24]` All three carry the `section_type` the 581 birth gate demands, and **all
three appear in the live `build-site-planner` menu for `garden-tools.uk`** — the site that could not
compose a month-by-month page this morning. Generic structural vocabulary moved 39→42 list-capable,
15→17 table-capable.

**A fourth was NOT built, and that is the useful part.** `mechanism-flow` (VIZ-006) already draws an
ordered process with decision branches, so a "steps" component would have been a near-duplicate —
the thing that makes a library harder to choose from. A checklist is unordered; a calendar's periods
do not cause one another. Reading the nearest neighbour first turned four components into three.

**⚠ §7's card-wall point is NOT closed either, and is not addressed by these.** The owner's third
complaint was the NUMBER of card sections per page. `period-calendar` renders compact rows rather
than cards precisely so twelve periods do not become twelve cards — but that is one component
declining to make it worse, not a fix for page composition.

**⚠ THE CLAIMS RISK ON `comparison-table` IS STATED, NOT SOLVED, and the figures are worse than the
council's own seat believed.** `[MEASURED 2026-08-24]` `banned_claims` is a KEY INSIDE the
`evidence_base` spec, not a separate aspect — so the two are one opt-in. **48 sites; 19 have an
evidence base; 15 of those have a non-empty `banned_claims[]`; 29 of 48 have NEITHER.** No
price/rating/score/rank field exists in the component, but cells are free text and on those 29 sites
nothing audits them. The gating machinery exists (591's `requires-evidence-base` tag) and was
deliberately NOT applied, because it would withhold the component from exactly the sites that need
it most — including this bug's own site. That is a judgement, it is one line to reverse, and it is
recorded in 606's header. **The real fix is `bugs_open/380`.**

**Still not evidenced: no page has been built with any of them.** Everything above is verified at
the seed, the schema, the rendered template (11 cases, engine-exact harness) and the live planner
menu. The outcome remains unproven until a build runs — see §8f.


---

## 10. STATUS 2026-08-25 — writer arm PROVEN, planner arm never run. **Not closeable; one build away.**

**Continue here:** `docs/agent_docs/docs024_key_docs_latest/bugfix_381_inexpressive_composition/HANDOFF_2026-08-25_continue_here.md`

Re-verified after the fresh chassis roll (`635f2d32f`), by needle in the live text rather than by
`updated_at` (that column is degenerate — see `LANDMINES.md`): all eight migrations recorded, both
arms live, all three components expressing.

**THE WRITER ARM IS EVIDENCED** `[MEASURED 2026-08-25, llm_call_log]`. Of the writer calls that
actually offered a retyped prose slot: **21 of 29 (72%) produced a list** against a **10%** baseline,
and **29 of 29 (100%) produced an `<h3>`**. 268 of 396 writer calls since the apply carry the new
RULE 10. This is §8f's falsifier NOT firing.

⚠ **THE PAGE-LEVEL NUMBER READS FLAT (5 of 48) AND IS THE WRONG INSTRUMENT.** Three dilutions, all
of which must be known before quoting it: most rows are **re-renders** from `content_data` written
under the old prompt (a rerender never calls the writer); most writer output does not reach a page in
the window (of 12 sampled list-bearing responses, **3 reached a page and all 3 carry the list**); and
the correlation probe is confounded because `rewrite_negations` can rewrite a sentence between
response and row, so a phrase miss is **not** evidence of loss.

**THE OPEN ITEM: `build-site-planner` has run ZERO times since the menus changed**, so `checklist`,
`period-calendar` and `comparison-table` have **0 placements**. Not broken — unexercised. All three
appear in the live planner menu for `garden-tools.uk` (3 of 3 verified), so the wiring is proven and
only the trigger is absent. **It needs one greenfield build** —
`scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh <domain> --email … --mission-file …`
— on a subject that is genuinely structured (a buying guide, a how-to, anything seasonal); a
two-page brochure exercises none of them.

**CLOSURE BAR:** `/bugs_open/` closes on **fixed AND live**. Half is live and evidenced; the half
that was the point of the bug has never run. **Do not close on the writer evidence alone.** Close
when one build places at least one of the three components, or demonstrably declines to and the
reason is understood.

---

## 11. ⭐ PROVEN ON A LIVE GREENFIELD BUILD, 2026-08-25 — `homegarden.uk`

**Continue here:** `docs/agent_docs/docs024_key_docs_latest/bugfix_381_inexpressive_composition/HANDOFF_2026-08-25_continue_here.md`
**Evidence:** that lane's `evidence/` — the captured planner prompt (114KB) and the mint snapshot.

Build dispatched 2026-08-25 **10:21:49Z**, planner ran **11:30:52Z**, first page built ~11:38Z.

### 11a. The three claims, each measured at the right layer

| claim | measured where | result |
|---|---|---|
| the planner is TOLD what components express | captured `prompt_rendered` | **YES** — `expresses` token, `[prose only]` token, rule 19, and all three new components present in the menu |
| the planner CHOOSES a capability-appropriate component | `pages.sections` + `site_plan_sections` | **YES** — `period-calendar` filed on the landing page. A component that did not exist two days earlier. |
| the writer USES the structure it is given | `page_components.rendered_html` on a served page | **YES** — see 11b |

### 11b. What the framework actually wrote, verbatim from `rendered_html`

```html
…keep some fleece or an old sheet to hand for cold nights.</p>
<h3>Garden jobs that genuinely need doing in April</h3>
<ul>
  <li>Mowing regularly, as grass growth picks up noticeably from early April in most of England…</li>
  <li><strong>Deadheading</strong> spring bulbs once flowers fade, leaving the foliage to die back
      naturally so the bulb can store energy for next year</li>
  <li>Checking <strong>soil pH</strong> before feeding beds heavily, since a lot of feed is wasted
      on soil that's too acid or alkaline for what's growing in it</li>
```

A real subhead, a real content list, `<strong>` on precisely the terms a reader scans for.
**`garden-tools.uk` had ZERO of all three across all seven served pages** — §2 of this file. That is
the before and after, on the same kind of subject, three weeks apart in effort and two days apart in
calendar time.

### 11c. ⚠ What this does NOT establish, stated because the temptation is to round up

- **One page of twenty-one.** 18 were still `triaged` when this was written. A single page is a
  single page.
- **The `bugs_open/206` risk is NOT disproven.** `april-index` is a `section-index` page and it built
  — 2 sections, no error — so the no-op did not fire *on that page*. That is much weaker than "the
  pairing is safe": 17 of 21 pages carry `section-index` at `page-build-handler`, the exact pairing
  that parked `garden-tools.uk/buying-guides-index`, and the routing fix (`efec862f4`, 09:58:33Z) is
  **not** in the running pods (`v1.0.1337`, started 09:27:48Z).
- **`checklist` 0 placements** — the planner was offered it and did not use it. A real negative,
  worth understanding, and NOT explained by anything measured yet.
- **`comparison-table` 0 placements is UNEXERCISED, not failed** — `[MEASURED by the
  loanzy_uk_example_site lane]` the vertical landscape was built from `rhs.org.uk` and
  `gardenersworld.com` only; the one comparison publisher in the draw (`which.co.uk`) crawled
  "successfully" and returned **0 sources**. The subject never gave the planner a reason to compare.
- **⚠ AND AN INTERACTION NOBODY PREDICTED, recorded not diagnosed.** The planner expressed
  "month by month" as **SEVENTEEN `section-index` pages** rather than as one page carrying
  `period-calendar`. So a structural promise satisfied at SITE level routes straight into the one
  page role with no builder — and if those 17 no-op, the site keeps the component and loses the
  calendar. Whether that is a 381 concern, a 206 concern or a third thing is open.

### 11d. Closure

**Not closed by this alone.** The bar is fixed AND live, and the planner arm is now demonstrably
both — but the honest position is "proven on one page of a build still running". Close when the
build settles and the picture holds. If the 17 `section-index` pages no-op, **that is `206` in this
window and must not be recorded as a 381 failure** — the symptom sentence is identical and the
acceptance script says so in bold.

### 11e. ⚠ RETRACTION of the "seventeen thin index pages" claim in 11c — my own data refutes it

§11c recorded, as the interaction "nobody predicted", that the planner had expressed *"month by
month"* as seventeen `section-index` pages **"so that no single page has to express it"**, and that
the site was "about to serve seventeen thin index pages". **That is withdrawn.**

`[MEASURED 2026-08-25 11:45Z]` the pages are not thin:

| page | component | chars | `<li>` | `<h3>` |
|---|---|---|---|---|
| `april-index` | generic-text-block | 2,822 | 4 | 4 |
| `august-index` | generic-text-block | 2,994 | 3 | 4 |
| `comparisons-index` | generic-text-block | 2,149 | 4 | 3 |

**4 of 4 bodies distinct** — not boilerplate. `august-index` opens *"August is usually the driest
month of the growing season, so the jobs that matter most this month are about protecting what you
already have rather than starting anything new."* Substantive, month-specific, structured content on
pages I had written off.

**And `n_sections = 3` is a planner CHOICE, not a default** — the discriminator the
`loanzy_uk_example_site` lane demanded before letting me publish the claim. Other sites'
`section-index` pages carry 1–2 sections (`["hero","blog-listing"]`, `["hero","guide-list"]`,
`["hero","category-listing"]`). **homegarden is the only site anywhere with
`["hero","generic-text-block","content-listing"]`, and the extra member is the prose block.** The
planner chose to give every month page prose; the writer filled each with a real list.

**The honest reading, which is better news than the claim I retracted:** the planner answered
"month by month" with a per-month PAGE architecture rather than one page carrying `period-calendar`,
and **each page keeps its own promise with real structure**. That is the fix working. What it
legitimately raises is whether `period-calendar` is NECESSARY for this shape — a finding about the
component, not about the planner dodging anything.

⚠ **How I got it wrong, because it is the fifth time in this lane and the same shape every time:**
I had the disconfirming data before I wrote the claim — `april-index` already displayed
`generic-text-block[LIST][H3]` in output I had read minutes earlier — and I wrote "thin" from the
**shape of the plan** (`n_sections=3`, a repeated layout) instead of from the **content of the
pages**. **A section count is not a content measure.** The peer challenge that made me look was
"is 3 a choice or a default?"; the answer mattered less than the side effect — it sent me to the
artefact.

### 11f. ⭐ THE CALENDAR SHIPPED — twelve months, on a served page, 2026-08-25 12:05Z

`homegarden.uk/index.html`, `build_status=deployed`, carrying a `period-calendar` instance of
**9,471 chars, one `<ol>`, twelve `<li>`** — labels `January … December`, all twelve, in order.

> **A month-by-month garden and home checklist**
> *Find your month below to see what's worth doing now, and how much it matters if you don't get to
> it straight away.*
>
> **January** — *Structural jobs while the garden rests* — "With little growing, this is the month
> for jobs that don't depend on the weather being mild: pruning apple and pear trees, checking shed
> roofs for storm damage, and clearing gutters before the wet months properly set in. None of this
> is urgent to the day, but leaving gutter clearing much longer risks water finding its way into
> brickwork."
>
> **February** — *First sowing and pre-spring exterior checks* — …

**Against §1, the owner's original complaint:** `garden-tools.uk`'s seasonal planner promised
*"What your shed needs, month by month"* under its own `<h2>` and delivered **four prose blocks with
three incidental month names**. This delivers **twelve months, each with a focus line and practical
detail** — and it hedges honestly ("none of this is urgent to the day"), which the field guidance
asked for and which no part of this fix could have forced.

**The whole chain, each link measured at the layer where it applies:**

| link | evidence |
|---|---|
| planner TOLD | captured `prompt_rendered`: `[expresses:]` / `[prose only]` tokens, rule 19, all three components in the menu |
| planner CHOSE | `period-calendar` in `pages.sections` and `site_plan_sections` for `index` |
| writer FILLED | twelve distinct month entries, not boilerplate |
| renderer SHIPPED | `build_status=deployed`, `/index.html`, one `<ol>`, twelve `<li>` |

**Site-wide writer arm at the same moment:** 31 sections rendered, **14 with a list, 17 with an
`<h3>`, 7 with `<strong>`** — against `garden-tools.uk`'s **0 / 0 / 0** across all seven of its
served pages (§2).

⚠ **Not yet the closure.** 12 of 21 pages deployed and the build is still running; `checklist` and
`comparison-table` remain unplaced (the latter unexercised by this vertical, §11e). And the final
page count must be taken **over HTTP** via the `loanzy_uk_example_site` lane's `after_test.sh`, not
from `build_status` — that column is wrong in both directions (their `contact` case: a page serving
57,753 bytes with `deployed_at` NULL).
