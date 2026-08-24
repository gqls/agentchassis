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
**34 list-capable and 10 table-capable components were available and none was chosen**, on a site
whose classification is a *buying-guide comparison hub* — a vertical whose entire value is structured
comparison. `build-site-planner`'s prompt lists available components by `name`/`display_name`/
`function`/`description`; **nothing in that listing states what markup a component can produce**, so
the planner is choosing blind on expressiveness.

## 4. A second, separable arm — and it is the cheap one

`generic-text-block` is a **pass-through**: its template is
`<div class="section__content">{{.content}}</div>`, so it will render any HTML it is given. **The
owner's wall-of-text paragraph is in this component** (verified: it is the only one on
`how-we-assess` whose `content_data` contains that text). The writer put
`<p>…</p><p>…</p>` into a slot that would have accepted `<h3>`, `<ul>` and `<strong>` unchanged.

So the defect has two arms with different fixes, and **the string "wall of text" cannot tell them
apart** — the same discipline `bugs_open/206` arrived at for `no sections ready to build`:

- **(a) WRITER-side, cheap.** A pass-through component given only paragraphs. Fixable today by
  telling `page-content-writer` to use `<h3>`/`<ul>`/`<strong>` where the content has natural
  structure, with no template or planner change. **This covers the owner's specific paragraph.**
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
