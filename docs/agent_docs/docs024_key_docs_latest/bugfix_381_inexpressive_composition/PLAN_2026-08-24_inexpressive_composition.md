# PLAN 2026-08-24 — `bugs_open/381`: pages composed from components that cannot express what they promise

Lane opened 2026-08-24 by the `bugs_open/381` session. The bug was filed the same day by
`loanzy_uk_example_site` from the greenfield `garden-tools.uk` build and explicitly marked
**UNOWNED** in that lane's HANDOFF 24 §0b. `scripts/who-owns.py 381` named only the filing lane.
**This lane owns the fix; the filing lane owns the account of the site.**

## 1. What the owner saw, and what it actually is

Two complaints — a 300-word wall of text on `how-we-assess`, and a "month by month" seasonal
planner with no months — that the bug file traced to one cause: **the page degrades to fit its
container.** The planner picked four components that cannot render a list, so the writer wrote
prose.

## 2. What the diagnosis added, and where it CORRECTS the bug file

### 2a. It is THREE layers, all prose-biased — not one

| layer | what it does today | evidence |
|---|---|---|
| planner | menu prints `name (display_name): function - description`. **Nothing about markup.** | live `load_components.config.query` on all three planners |
| **field guidance** | `llm_guidance` is the per-field instruction the writer actually reads | `plan_sections_action.go:2328` → prompt's `{{if .description}}` |
| writer prompt | RULE 9: `text` fields are plain strings, no HTML. RULE 10 permits `<p>`, and only for types `rich_text`/`content` | live `page-content-writer` prompt |
| renderer | `text/template`, **no escaping** — whatever the writer emits renders verbatim | `component_library.go:1060` → `call_agent.go:1170` |

**RULE 10 names two types NO COMPONENT DECLARES** `[MEASURED 2026-08-24]`: across active section
components, 940 llm fields are `text` (135 components), 2 are `html`, and **`rich_text` and
`content` are declared zero times.** The rule that permits structure is addressed to nobody.

### 2b. ⚠ THE LEVER IS `llm_guidance`, NOT THE TYPE — this corrects my own first plan

A controlled natural experiment already exists in the live data, with the type held constant
`[MEASURED 2026-08-24, 30d instances]`:

| field | declared type | has `llm_guidance`? | instances rendering a `<ul>/<ol>` |
|---|---|---|---|
| `article-body.content` | **`text`** | **yes** — *"Use h2 for main sections, h3 for subsections, p for paragraphs, ul/ol for lists, blockquote for callouts"* | **116 / 153 = 76%** |
| `generic-text-block.content` | **`text`** | **no guidance at all** | **12 / 173 = 7%** |

Same declared type, same renderer, same RULE 9 forbidding HTML in `text` fields — and an
11-fold difference in outcome. **RULE 9 does not bind; the field's own guidance does.** So:

- The **first** cut of this plan treated the `text` → `html` retype as the fix. That was wrong,
  and `article-body` is the disproof: it is `text` and it already writes lists.
- The retype still earns its place, but for a different and smaller reason: **the type is the
  routing key between RULE 9 and RULE 10** (the prompt prints `` `content` (html, required) ``),
  so retyping is what stops RULE 9 contradicting the guidance. It is honesty, not force.
- ⚠ **The type is otherwise INERT.** `DeclaredTypeSatisfied` is default-TRUE — only `array`/`list`
  are checked (`datahelpers/content_type_violations.go:262`), so `hmtl`, `HTML` or `""` would all
  behave identically to `html` and nothing downstream would ever surface it (flagged by the
  `staged-component-build` lane, 2026-08-24: *"a check that cannot return false"*). The migration
  therefore asserts each field reads back **literally** `'html'` after the write.

### 2c. The "34 list-capable components sat available" premise is true and misleading

Enumerated, the structural components are directories, trackers, calculators, quizzes, spec
sheets, one `pricing` table, two carousels — and `site-footer`, which is chrome. **There is no
generic checklist, steps, comparison-table or calendar component in the library.** A planner told
what each component can express would still have had nothing to compose a seasonal planner from.
The filing lane has recorded this correction in the bug file itself, with the general form:
**counting CAPABILITY without checking SUITABILITY.**

### 2d. `content_shape` was designed as exactly this axis and is dead

`content_components.content_shape` (added with the selector metadata block,
`sql_for_tables/005_content_components.sql:9118`, COMMENT: *"prose, structured_list,
structured_card, key_value_pairs"*) has **zero Go readers**, is **not written at component birth**
(`store_generated_component_action.go:634` omits it), is NULL on 128/151 active section rows, has
drifted to free text (`series`, `sequence`, `mixed`), and **12 rows marked `structured_list` have
no list markup in their template.** A hand-maintained capability column is the failure mode.
**So capability must be DERIVED, never declared.**

### 2e. `content-gap-planner` is the busiest planner, not `build-site-planner`

`[MEASURED 2026-08-24, llm_call_log 30d]` — `content-gap-planner` **749** calls,
`build-site-planner` **27**, `site-planner` **5**. The bug was found on a greenfield build, but
the fix's volume lands on the gap planner.

## 3. Design — four migrations, config-only, live on apply

### A. The planner is told what each component can EXPRESS
`component_expresses(html_template text, input_schema jsonb) RETURNS text[]`, `IMMUTABLE`, called
at **read time** from the planner menu queries. Vocabulary, all derived:
`html-block` (an `llm` field typed `html`) · `list` (`<ul|ol`, or `html-block`) ·
`table` (`<table`, or `html-block`) · `items` (`{{range` over an `llm` array field) · empty = prose only.

No `ALTER TABLE` — the RFC_032/283 lane is rewriting `html_template` fleet-wide, and a stored
column would be stale the moment a template changed. 151 rows; cost is nil.

Menus gain one column and the listing line one token, in the SAME migration (with
`missingkey=zero`, a key the SELECT does not return prints `<no value>`), plus one planning rule
tying structure to promise.

**Dependency stays in the ROW gate, not the vocabulary** (agreed with the `news editorial` lane):
a fact-fed component surfaced as a generic list-expresser would be picked on evidence-less sites
where it can only fail. `component_expresses` stays pure; the menu SQL honours a
`requires-evidence-base` semantic tag exactly as 419 honours `requires-backend`. That lane
measured `data_sources` as **empty on both** `evidence-chart` and `evidence-timeseries` (one
active component uses it at all, fleet-wide), so **the tag is the only mechanism**, not
belt-and-braces. The two-row tagging UPDATE is owed by that lane; an untagged row under the gate
is ungated-as-today, so order does not matter.

⚠ `site-planner`'s menu query takes **no site parameter**, so the evidence gate cannot be
expressed there without inventing one. Recorded as a known asymmetry rather than papered over.

### B. The writer is told to write structure where the slot can hold it
Per-field `llm_guidance` (the measured lever), modelled on `article-body`'s proven wording, plus
the retype that routes the field to RULE 10.

| component.field | today | action |
|---|---|---|
| `generic-text-block.content` | `text`, **no guidance** | add guidance + retype `html` |
| `about-content.content` | `text`, **no guidance** | add guidance + retype `html` |
| `illustrated-text-block.content` | `text`, guidance says *"One or more HTML `<p>` paragraphs"* — **actively restricts to `<p>`** | replace guidance + retype `html` |
| `article-body.content` | `text`, good guidance, 76% lists | retype `html`; append only the forbidden-elements sentence |
| `report-dossier.body` | guidance: *"Pre-rendered … **Never authored by an LLM**"* | **EXCLUDED** — not a writer slot |

Vocabulary (fixed with the `news editorial` lane): `<h3>` subheads, `<p>`, `<ul>/<ol>/<li>`,
`<strong>`, `<table>` when tabular, bare `<blockquote>`. **Explicitly forbidden: `<img>`,
`<figure>`, `<iframe>`** — in-blob imagery is the loss class `inline_guide_imagery` and
`features_open/035` exist to retire, and this change would otherwise be the enabling edit. Also
no forms, inputs, ids, classes, scripts or inline styles: **the writer owns prose STRUCTURE, the
component system owns design.** No furniture class names in guidance — under 035 a pull-quote
becomes a child component instance, and a class name in guidance is a comment, not a control.

RULE 9 is narrowed so it stops contradicting field guidance; RULE 10 is re-addressed to `html`.
Rule 10's enumerable examples are deliberately **not** first-person practice claims ("months of
the year, steps a reader takes, options being compared") — agreed with the `bugs_open/380` lane,
whose writer arm forbids exactly that shape on sites with no operating history.

### C. Later, and only after A+B are measured — the plan's PROMISE travels to the writer
`PLAN-025`, unbuilt: `plan_sections.sectionDescription` fires **only** on the not-found path
(`plan_sections_action.go:1308`), so the writer never sees the plan's intent for a section that
resolved. Optional `"brief"` on object-form section entries, riding the fact-scoping path
(`bugs_open/151`). Go + migration + tests, its own council round. PLAN-025's own caveat is
planner output size, which is why it waits for evidence that A+B were not enough.

## 4. What this deliberately does NOT do

- **No new components.** The library's missing generic structured block is real (§2c) and is a
  separate piece of work — filed as a finding, not smuggled into a bug fix.
- **No `component_selector.go` change.** The `bugs_open/378` lane confirmed the incumbency trap it
  removed is a property of Path 2 (the `section_type` scored contest); this fix operates at the
  planner level on Path 1, resolved by name, which never reaches that score.
- **No restyling.** §5 of the bug file stands: a designer cannot add a `<ul>` to a template that
  has none, and `vigilant_designer` does not exist.
- **No site repair.** `garden-tools.uk` is untouched (owner's standing instruction, and it is
  `bugs_open/206`'s closure canary).

## 5. Sequencing and coordination (all agreed 2026-08-24)

- Migration numbers **591–595 this lane, 596–599 `bugs_open/380`.** Both lanes ship every prompt
  edit as an anchored `replace()` with an exact-count precheck — **never a wholesale base64
  rewrite**, which would silently revert whoever landed first. Anchors are disjoint and listed in
  NOTES.
- `bugs_open/283` (RFC_032): this lane touches `input_schema` on 4 rows and **never**
  `html_template`; `component_expresses` reads at query time, so no ALTER and no trigger.
- `305 negation gate`: probing this change's shape found a REAL defect in their scanner —
  `</th` was missing from the sentence-boundary list while `</td` was present, so a
  define-by-negation construction in a **header cell** produced a markup-bearing "sentence" that a
  repair would have spliced over, breaking the table. Fixed by that lane in `714789d7b`, **Go, so
  inert until the next roll.** ⚠ **Timing consequence for this lane: prefer landing B after that
  roll**, or watch for header-cell copy in the window. Narrow exposure — needs the construction
  inside a `<th>` specifically.
- `staged-component-build`: no gate of theirs enumerates field types; `html` needs to be on no
  list. Their flag about default-true type checking is folded into §2b.
- `news editorial` / `editorial_design_uplift`: adjacent, not duplicate — their Phase B is CSS
  furniture, this is the writer/type seam. Their constraints are adopted verbatim in §3B.

## 6. Falsifiers for this plan

- A post-apply build where a `[expresses: list]` component is chosen for a promise-bearing page,
  yet the page still ships prose → the planner listing is not the lever either, and the cause is
  further downstream.
- `generic-text-block` structure share failing to move toward `article-body`'s 76% after B →
  §2b's controlled comparison was confounded (most likely by WHO writes it: the blog path vs
  `page-content-writer`), and the writer prompt, not the field guidance, is load-bearing after all.
- A council REJECT on scope: the planner menu is a shared seam across three agents, and this
  changes what every one of them sees.
