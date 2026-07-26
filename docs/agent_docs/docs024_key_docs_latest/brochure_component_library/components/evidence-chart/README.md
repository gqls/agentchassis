# `evidence-chart` — the shared, evidence-sourced chart component

Owner green-light 2026-07-26. One shared chassis component for every site,
**values sourced from `site_specs.evidence_base`** so a chart cannot structurally
display an unverified figure. Code-rendered (CSS bars from the real value), no
runtime chart library, no generated image of a chart, no JS at all.

Source of truth for the DB row: `template.html` + `input_schema.json`.
Regenerate the installer after editing either — never hand-edit `register.sql`:

```bash
python3 ../../scripts/gen_component_register_sql.py .
```

## The guarantee, and why it holds

```
charts  ← site_specs.evidence_base.charts   ids, labels, order, unit, max, pages
facts   ← site_specs.evidence_base.facts    THE VALUES — nothing else holds them
eyebrow / title / intro ← llm               framing only
```

A chart definition **names fact ids and never restates a value**. The template
joins `point.fact_id` → `facts[].id`, so a number that is not a fact row has
nowhere to live. Both data fields are system-resolved in
`plan_sections_action.go`, and resolved data beats LLM content at render time —
so the writer cannot supply a figure even if it tries. This is a structural
property, not a prompt instruction, which is the whole point:
`features_open/023` R3 exists because prompt discipline was holding this line.

**Fail closed.** Both fields are `required` with `on_missing: skip_section`. A
site with no evidence base gets no chart, not an invented one.

## Traps found while building it (all reproduced, all fixed)

1. **A round million renders as `1e+06`.** JSONB numbers arrive as `float64` and
   `{{.value}}` prints via `%v` (i.e. `%g`), so `1000000` becomes `1e+06` —
   invalid CSS in the bar geometry *and* nonsense as visible text. Geometry uses
   `printf "%.4f"`; the visible fallback uses `printf "%.10g"`, which keeps
   decimals but never goes exponential in any range we would chart.
2. **`html/template` neutralises a hostile value in a `style` attribute**
   (`ZgotmplZ`) — proven, so the data layer cannot inject CSS. The same filter
   rejects a *string*-typed value under `printf "%.4f"`, so **a charted fact's
   `value` must be a JSON number**; the VERIFY script asserts it.
3. **The claims gate reads a ±70-character window** around a number and needs one
   of the fact's `context_terms` inside it (`datahelpers/claims.go:493`), and
   block elements delimit those windows. So the point label and its figure must
   sit in ONE block element — they do, in `.evidence-chart__row` — and a label
   should echo the fact's own wording.
4. **`<svg>` is in `nonAssertionElements`** (`claims.go:137`): text inside an SVG
   is invisible to the claims gate. This component keeps every figure in real
   HTML text, so the gate still sees it. **Anyone lifting the geometry into the
   Go SVG emitter later must add a check for this** — the numbers would silently
   leave the gate's view.
5. **Dangling `fact_id` draws nothing** — no bar, no row, no invented value.
   The VERIFY script makes it impossible rather than relying on the silence.
6. **Half the CSS variable names in the obvious vocabulary do not exist.**
   Checked against live `css_themes`: `--color-surface`, `--spacing-section` and
   `--container-max-width` are defined by **no** theme, so anything resting on
   them always renders its fallback — which is how a light card lands on a dark
   page. The names the themes really define are `--color-background`,
   `--color-text`, `--color-text-muted`, `--color-primary`, `--color-secondary`,
   `--color-accent`, `--color-card-bg`, `--color-border`, `--border-radius`,
   `--shadow`, `--spacing-xs…xl`. This component uses those, and where no
   variable exists its fallback is a neutral translucent grey that reads on
   light and dark alike rather than a light literal.

## Per-page selection

A chart may declare `pages: ["index"]`. The template filters on `current_page`
(present in the render data map, `component_library.go:756`/`:873`). A chart with
no `pages` key shows everywhere, and an empty `current_page` degrades to showing
all charts — so a render path that does not set it under-filters rather than
rendering an empty section. Both `index` and `index.html` forms match.

## Data contract

```jsonc
"charts": [{
  "id": "relojistas-feed-restoration",   // stable; becomes data-chart="…"
  "title": "…", "caption": "…",          // prose, no figures
  "unit": "%",                           // suffix appended to every value
  "max": 100,                            // scale constant only (e.g. 100 for a %)
  "max_fact_id": "F9-…",                 // PREFERRED: denominator from a fact row,
                                         // so no business figure sits in a definition
  "pages": ["index"],                    // optional
  "source_note": "…",                    // prose, no figures
  "points": [{ "fact_id": "F3-…", "label": "…", "tone": "primary|muted|accent" }]
}]
```

A charted fact may carry an optional `"display": "97"` — how the figure should
read. The fact's `value` stays the canonical number and the VERIFY script checks
`display` against it, so display can never drift into a different figure.

**Only exact-tolerance facts may be charted.** `F1` and `F2` carry
`tolerance: gte` and "state a FLOOR, never the exact number" — a bar labelled
with the exact value would break the fact's own rule.

## Validate before installing

```bash
go run ../../scripts/render_component_template.go template.html sample_data.json index
go run ../../scripts/render_component_template.go template.html sample_data.json capabilities
go run ../../scripts/render_component_template.go template.html sample_data.json ""
```

The sample exercises the failing branches on purpose: an excluded page, a
dangling fact id, a zero value, a round million, and an unreferenced fact.

## Acceptance (PLAN's 7 items)

- [x] 1. Row in `content_components`, `component_level='section'`.
- [ ] 2. Named in the planner prompt with a selection rule (`site-architect`).
- [x] 3. CSS on `var(--color-*, <literal>)` throughout, no hardcoded colour.
- [x] 4. No JS asset — nothing to publish, nothing to bundle (deliberate for v1).
- [x] 5. No `site_plan_imagery` kind.
- [ ] 6. Copy path is content-writer + `validate_page_content`.
- [ ] 7. Links verified as served (this component emits no `href` at all).
