# Register — visualisation-and-charts

> **covers-through: 2026-07-28** · written 2026-07-28 from first-hand code/DB reads
> and from rendering the templates, never part of the extraction.
> Everything else dates from the 2026-07-13 extraction freeze — absence
> here is not evidence of absence in the platform. See `bugs_open/106`.

11 concepts. **NOT from the 2026-07-13 extraction.** Two of these components
shipped after the freeze and one shipped the day this entry was written, so none
was ever in the register.

This entry exists because of a specific, recorded failure. On 2026-07-27 a
workstream handoff asserted, in bold, *"there is no chart renderer"*, and that
claim was repeated to the owner twice and used to classify graphs as blocked
work. **Two renderers were live at the time.** The owner corrected it. The cause
was not laziness: the evidence cited was true (`go-echarts` is genuinely absent
from `go.mod`; `report_charts.go` is genuinely narrow) and the conclusion did not
follow, because **capability on this platform is mostly DATA** — `evidence-chart`
is a row in `content_components` and no grep of Go source can ever find it. That
is the reason this register entry is worth its length: the searching habit that
finds code does not find components. See `WRONG_CALLS.md` 2026-07-28.

Status vocabulary per `README.md`. Where a thing is *built but never exercised on
a live site*, that is said explicitly.

### VIZ-001 — `evidence-chart`: magnitudes, resolved through fact ids
- **status:** deployed (live on 1 site)
- **status-evidence:** row read 2026-07-28: `component_level='section'`, `render_mode='agent'`, active; template executed locally against a fixture, parses and renders.
- **what:** Horizontal bar chart comparing quantities. Bars are **CSS**, not SVG: the template writes `style="--v:<value>;--m:<max>"` and the browser does the division. Every plotted point resolves through `{{range $f := $facts}}{{if eq $f.id $p.fact_id}}` — **a chart point cannot carry its own number**. Each row renders `verified {{$f.verified_at}}`; the figure carries a `source_note`. Charts may be filtered per page via `$c.pages`.
- **why it matters:** this is the doctrine ("charts are code-rendered from real figures") existing as a *mechanism* rather than a rule. The unsourced state is unrepresentable, not merely discouraged.
- **sources:** `content_components` row `evidence-chart` (html_template read in full 2026-07-28); `features_open/023`
- **relations:** VIZ-002, VIZ-003, VIZ-005, CLM-001

### VIZ-002 — `evidence-timeseries`: one measurement over time
- **status:** built, **not yet exercised on a live page**
- **status-evidence:** applied 2026-07-28 (`sql_for_agents/250`); template executed against a real two-point series — 2 columns, correct values, dated axis, scale resolved, both sources linked, zero unrendered tags. No live page uses it yet.
- **what:** The companion to VIZ-001. That compares magnitudes *between things*; this shows one measurement *over time*, one column per observation, with the `as_of` of each point as the axis tick. Resolves the series through a `fact_id`, and renders **each point's own citation and date beneath the plot**.
- **why it matters:** it is the first renderer whose x-axis is meaningful, and it is deliberately the *second* thing built — the substrate (VIZ-003) landed first, because a time-series component with no legitimate series to plot is an invitation for a writer to fill it from the model.
- **sources:** `sql_for_agents/250_evidence_timeseries_component.sql`; `docs024/oufe/DESIGN_2026-07-28_premise_branching_and_deepthink.md` §3
- **relations:** VIZ-003, VIZ-006, VIZ-007, VIZ-008
- **verify-later:** first live use; whether the column form reads as a time series to a reader, or wants a line

### VIZ-003 — series facts: the substrate a time axis needs
- **status:** deployed (in `v1.0.1185`, pod-verified) — **no live site has a series fact yet**
- **status-evidence:** three distinctive string literals from `claims_series.go` found in the running binary 2026-07-28, with a positive control.
- **what:** An `EvidenceFact` held one `Value` and three dates — `accessed`, `published`, `verified_at` — **all of which are provenance**. None is the date the value *applies to*, so a time series had no honest shape. `Observation{as_of, value, source, verified_at}` adds one. Two rules are load-bearing: **every observation carries its own source, never inherited from the parent fact**, and `as_of` is distinct from `verified_at` (re-checking a 2021 figure in 2026 moves `verified_at` and must not move the point on the axis).
- **why it matters:** a series where the first point is cited and the rest "continue from the same source" is exactly how interpolation and extrapolation enter looking like data.
- **sources:** `platform/orchestration/datahelpers/claims_series.go`; `claims.go` (`Observations` field, `numberSupported` branch)
- **relations:** VIZ-002, CLM-001, CLM-003

### VIZ-004 — the honesty gate had to learn about series, or it would have fought the chart
- **status:** deployed
- **what:** `numberSupported` skips any fact with `Value == nil`, which is every series fact by design. Without a branch for them, **every value plotted from a series would be reported as an unregistered number**. Matching is deliberately *exact* even when the fact carries a `gte` tolerance, because a `gte` series would blanket-support nearly every number on the page.
- **why it matters:** the round-1 council objection on this change found the sharper form of it — `ValidateSeries` enforced the per-observation source rule but `numberSupported` never called `ValidateSeries`, so an unsourced observation still registered its value. **A rule enforced only in a validator is not enforced**; it has to hold at the gate that decides. Both now share `observationHasResolvableSource`.
- **sources:** `claims.go` `numberSupported`; `claims_series.go` `seriesSupports`; council correlation `da40ddf0` round 1
- **relations:** VIZ-003, CLM-003

### VIZ-005 — the boundary: generated images explain, code-rendered output states
- **status:** designed, not built (the rule is stated; nothing enforces it)
- **what:** Generated (diffusion) imagery is acceptable for *explanatory* graphics — how a process runs, the shape of an architecture. It is the wrong tool for any value that must be **exact**, **selectable** or **translatable**, because a diffusion model draws a bar of approximately the right height and text baked into a JPEG cannot be copied or re-rendered in another language.
- **why it matters:** the routing fix in `bugs_closed/011` made publishable infographics real, and the first one produced was accurate *only because a human hand-wrote audited figures into the prompt*. That is prompt discipline holding a line that should be structural.
- **sources:** `features_open/023_FEATURE_infographic_figures_from_the_evidence_base.md` (R3, R4)
- **relations:** VIZ-001, VIZ-009, CLM-001

### VIZ-006 — `mechanism-flow`: drawing a process, with no numeric field at all
- **status:** deployed (live on oufe.com `/cases/thames-water.html`)
- **status-evidence:** 7 steps and 2 decision branches rendered on the live page, fetched 2026-07-28.
- **what:** A numbered vertical flow with optional decision branches, for explaining a legal or financial *mechanism* — sequence and consequence rather than magnitude. Connectors are CSS; there is no SVG and nothing to load.
- **why it matters:** it **has no numeric field by design.** On an evidence-gated site a number-shaped slot is an invitation for a writer to fill it, so the absence of the slot is the control. It is the answer to "infographics for the harder concepts" that carries no figure risk at all.
- **sources:** `sql_for_agents/247_mechanism_flow_component.sql`; `sql_for_agents/248`
- **relations:** VIZ-007, VIZ-009, VIZ-011

### VIZ-007 — there is NO arithmetic in the render funcmap, and a missing function is a PARSE error
- **status:** deployed (a constraint, not a feature)
- **status-evidence:** the only `template.FuncMap`s are `render_css_from_spec_action.go:238` and `compute_component_quality.go:354`; `executeGoTemplate` (`call_agent.go:1151`) registers `default`, `eq`, `ne`, `lower`, `upper`, `isset` and no arithmetic.
- **what:** No `inc`, `add`, or any numeric helper. A template referencing one **fails to parse**, so the component renders *nothing* — it does not degrade.
- **why it matters:** this single constraint shapes every chart on the platform. It rules out computing SVG polyline coordinates in a template, which is why both chart components pass values to CSS custom properties and let the browser divide. A first draft of VIZ-006 numbered its steps with `{{inc $i}}` and would have rendered an empty section; a CSS counter needs no funcmap at all.
- **sources:** `platform/orchestration/actions/call_agent.go:1149-1175`
- **relations:** VIZ-001, VIZ-002, VIZ-006, VIZ-008

### VIZ-008 — `$facts` is declared by the template, not supplied by the engine
- **status:** deployed
- **what:** `evidence-chart` opens with `{{- $facts := .facts -}}` and `{{- $page := .current_page -}}`. Go templates treat an **undeclared** variable as a parse error, so a component that references `$facts` without declaring it fails to render entirely.
- **why it matters:** it is a contract that is invisible unless you read the top of a working template, and its failure mode is total rather than partial. Caught on `evidence-timeseries` before shipping only by executing the template — the same class as VIZ-007.
- **sources:** `content_components` row `evidence-chart` (template head); `sql_for_agents/250`
- **relations:** VIZ-002, VIZ-007

### VIZ-009 — text inside `<svg>` is invisible to the claims gate
- **status:** deployed (a hazard, not a feature)
- **status-evidence:** confirmed the other way round on 2026-07-28 — the base64 payload `claimscan` actually received for `mechanism-flow` was decoded and found to contain the diagram's own words.
- **what:** `extractAssertions` walks HTML text nodes; SVG text is not reached. **A diagram built from `<svg><text>` could assert anything and scan clean.**
- **why it matters:** it makes "draw it in SVG" the wrong default for any graphic carrying words on an evidence-gated site, and it is the reason both new components use real HTML text with CSS-drawn furniture.
- **sources:** `claims.go` `extractAssertions:165-226`; brochure-workstream landmine
- **relations:** VIZ-001, VIZ-002, VIZ-006, CLM-002

### VIZ-010 — `cmd/contrastscan`: the post-deploy contrast witness
- **status:** deployed (built 2026-07-28), exercised across 10 live sites
- **what:** Measures WCAG contrast on live pages in headless Chromium using computed style and the **actual painted backdrop**, alpha-composited. Exits non-zero on failure.
- **why it matters:** it exists because a stylesheet cannot answer the question — it cannot resolve the cascade, and the result depends on ancestors, alpha and gradients. `bugs_open/122`'s original table was built from a regex and was largely wrong. The tool's three guards are each a false positive it produced before having them, and the general rule is recorded with it: **on live public sites an over-reporting audit is worse than none**, because its findings get "fixed" into real regressions.
- **complements, does not replace:** `platform/colour.AuditPalette` (026 phase 2b) reads the *composed palette* pre-deploy in microseconds; this reads the *painted page* post-deploy. A colour can be legible in the palette and illegible on the page, because chrome carries hardcoded literals that are in no palette.
- **sources:** `cmd/contrastscan/main.go`; `bugs_open/122`; `platform/colour/palette_audit.go:89`
- **relations:** VIZ-011

### VIZ-011 — chart furniture is a graphical object, so the 3.0 threshold applies to it
- **status:** deployed (applied in VIZ-002 and VIZ-006)
- **what:** Axis lines, connectors and bars are not decoration: they are required to understand the content, so WCAG's 3.0 non-text contrast minimum applies. On oufe's live stylesheet `--color-border` (#2E3F52) scores **1.66** against the page background and fails it; `--color-accent` scores **6.86**. Both new components therefore draw their furniture in the accent.
- **why it matters:** the intuitive choice (`--color-border`, because it is a border) is the failing one, and nothing on the platform checks it at build time.
- **sources:** measured 2026-07-28 against the live stylesheet; `bugs_open/122`
- **relations:** VIZ-006, VIZ-002, VIZ-010
