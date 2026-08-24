# Inventory — round-trip writers (the architecture seat's ask, 198 council round `5249320e`)

Owed since **2026-08-05**. First complete population pass: **2026-08-24**.

The seat's objection, verbatim: *"No inventory of other `agent_definitions` workflows that
round-trip a full artifact through an LLM and write it straight to a DB column/file with no size
or shrink guard — that inventory would turn the class-cost claim from qualitative to
quantified."*

**Answer, in one line:** the population is **9 writer steps across 6 active definitions**
(as of 2026-08-24), of which **3** are the shape the seat means, **3** replace an incumbent
artefact, and **all 3 replace-paths carry a live size guard** — so the class is currently
**closed on the measured population**, with two named blind spots that make 9 a **floor**.

---

## 1. Method — and why the one in `HANDOFF_2026-08-10` under-reports

Ground truth for "what is LLM output" (never inferred from path spellings):
`execute_llm_prompt` **140** steps / **66** active definitions, plus `generate_html` 4/2 and one
each of `execute_vision_prompt`, `fetch_llm_news`, `generate_provocations`, `ch_llm_review`.

References are `output_field.subfield` against `collected_data`, so the graph is over
**`output_field` names**, not step names. The 08-10 method's join is **one hop**, and that is
why it returned 1 row. **Transitive closure** over the reference graph returns 9 — and
**6 of the 9 are multi-hop**, i.e. structurally invisible to the documented join.

| hops | writer steps |
|---|---|
| 1 | 3 |
| 2 | 4 |
| 3 | 2 |

⚠ **The one-hop join cannot see `bugs_open/198` itself.** `css-patch-agent/deploy_css` — the git
writer that gutted nine stylesheets — has `content_field = css_saved.css_content`, the **save**
step's output. The artefact travels `LLM → DB row → git commit`; the join sees neither hop
together, and reports **0** exposed `git_commit` writers.

### Two blind spots that remain — so 9 is a FLOOR, not a count
1. **Implicit Go-side inputs.** The graph is built from config *text*. `webdesign-agent`'s
   `generate_css` has an **empty `config`** — `render_css_from_spec` reads `design_spec` in Go —
   so the edge `generated_css → design_spec` does not exist textually and no closure over config
   can find it. (Here it changes no verdict: the renderer is deterministic. Elsewhere it might.)
2. **`files_field` map-of-files** payloads are matched, but their per-file provenance is not
   traced; a map assembled in Go has the same invisibility as (1).

---

## 2. The class needs SPLITTING — and only one half has 198's failure mode

The seat's phrase conflates two shapes. The distinction is not cosmetic: they fail differently.

**A — the LLM's returned bytes ARE the persisted artefact** (3 steps, all 1-hop). This is the
012/198 class. The live risk is **prompt non-compliance that still parses** — truncation already
fails loud (`ai_actions.go:427`, `aiservice.IsTruncated`) and `output_format: json` parse-fails
closed, so the dangerous output is well-formed and gutted.

**B — the LLM output is an INPUT to a deterministic Go composer/renderer that produces the
bytes** (6 steps). A non-compliant reply fails spec validation or composition, not silently. Two
of these six (`component-template-fixer`) write `site_work_items` — a work item, **not an
artefact** — so they are out of class entirely.

| definition | step | action | hops | path | class |
|---|---|---|---|---|---|
| css-patch-agent | `save_css_to_db` | query_database (UPDATE) | 1 | `css_fix` | **A** |
| component-creator | `store_component` | store_generated_component | 1 | `generate_template` | **A** |
| tool-generator | `save_tool` | create_tool_component | 1 | `generated_html` | **A** |
| css-patch-agent | `deploy_css` | git_commit | 2 | `css_fix → css_saved` | B |
| report-builder | `publish_ready` | git_commit | 2 | `report_prose → composed` | B |
| report-builder | `deploy_page` | git_commit | 3 | `report_prose → composed → rendered_page` | B |
| tool-recreation-handler | `save_sections` | save_page_sections | 3 | `tool_recreation → completeness_check → validation_result` | B |
| component-template-fixer | `create_rerender` | query_database (INSERT) | 2 | `scoped_script → fix_result` | out (work item) |
| component-template-fixer | `create_section_edit_delivery` | query_database (INSERT) | 2 | `scoped_script → fix_result` | out (work item) |

---

## 3. The guard story, cross-cut by REPLACE vs CREATE

**This is the load-bearing distinction and the 08-10 method does not draw it: a shrink floor is
meaningless without an incumbent.** 198 was the replacement of a 17–26 KB file. A writer creating
a component has nothing to compare against, so "no shrink guard" is not a defect there — the
right guard is structural.

| writer | shape | guard | live? |
|---|---|---|---|
| `save_css_to_db` | REPLACE `css_themes.css_content` | migration 542 base-integrity refusal (<4096 B base, shared theme row) + monotonic append (318) | yes |
| `deploy_css` | REPLACE the stylesheet FILE | **DGH-016** `file_shrink_floor = 0.5` | yes — both halves on `70fd163c2` |
| `save_sections` | REPLACE page sections | `enforceSectionShrinkFloor` (`save_page_sections_action.go:681`) | yes — present in the running build `70fd163c2`, not just the dirty tree |
| `store_component` | CREATE | structural truncation gates: no `<section>`/`<div>` → reject; unclosed `<style>` in markup context → reject; empty `input_schema` → reject; plus field-name contract on regeneration | yes |
| `save_tool` | CREATE | field-contract enforcement point, hard fail (`create_tool_component_action.go:115`) | yes |

**So: 3 replace-paths, 3 guards, all live.** The quantified class-cost answer is not "N unguarded
writers" — it is *"three direct writers; three replace-paths; every one guarded; population a
floor for two named structural reasons."*

---

## 4. ⚠ A number in here that will be misread — do NOT file 19 bugs

**20 `git_commit` steps exist fleet-wide and exactly ONE carries `file_shrink_floor`**
(`css-patch-agent/deploy_css`, at 0.5). That reads like 19 exposed writers. It is not:
- most write `files_field` maps of **Go-rendered** pages (`page-rerender`, `sitemap-refresh`,
  `content-feed-orchestrator`, the seven `deploy_js_snippets` steps) — class B;
- several write a **fresh** path with no incumbent, where a ratio has nothing to divide by;
- `webdesign-agent/deploy_css` **looks** like the 198 case (`content_field=generated_css.result`,
  no floor) and is not: `generate_css` is `render_css_from_spec`, a deterministic Go renderer,
  explicitly documented as "no LLM".

Naming this because the neighbouring lane hit the same trap this week from the other side (a
census that looked like a fleet-wide gap and was a house norm already met, `WRONG_CALLS.md`
2026-08-24). **The census that matters is `REPLACE ∧ LLM-returned-bytes`, which is 3.**

---

## 5. What is NOT claimed
- No fleet-wide assertion that nothing is exposed. The floor caveat in §1 is real and unbuilt
  work: closing it needs the Go-side input edges, which no config query can supply.
- Step 4 of the 08-10 method (read each **prompt** for a whole-artefact vs fragment contract) is
  done only for the three class-A steps by inference from their guards, **not** by reading all
  nine prompts.
- No `090` run: this is a survey, not a root-cause claim. Per the 2026-07-31 ruling, the moment
  anyone asserts a cross-cutting defect from this table, that ruling applies.
