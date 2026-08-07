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
- **status:** **live and exercised** — first real use 2026-07-29, Thames page (`evidence-timeseries-leakage`, migs 265/266)
- **status-evidence:** five-observation series (Thames leakage 2020-25, restated APR figures) served on `/cases/thames-water.html`; claimscan 0 findings across 18 components with the series values resolving through `seriesSupports`; render audit clean (8 pages, 0 firm) the same hour. Scale denominator is a registered fact (the 20.5% final-year commitment), per the component's own rule.
- **first-use notes:** `as_of` renders raw as the axis tick, so year-end months read "2021-03" — honest but terse; the chart footnote explains the 31 March year end. `printf "%.10g"` renders 12.0 as "12" — accepted rather than adding a `display` key that would make content_data diverge from the register.
- **what:** The companion to VIZ-001. That compares magnitudes *between things*; this shows one measurement *over time*, one column per observation, with the `as_of` of each point as the axis tick. Resolves the series through a `fact_id`, and renders **each point's own citation and date beneath the plot**.
- **why it matters:** it is the first renderer whose x-axis is meaningful, and it is deliberately the *second* thing built — the substrate (VIZ-003) landed first, because a time-series component with no legitimate series to plot is an invitation for a writer to fill it from the model.
- **sources:** `sql_for_agents/250_evidence_timeseries_component.sql`; `docs024/oufe/DESIGN_2026-07-28_premise_branching_and_deepthink.md` §3
- **relations:** VIZ-003, VIZ-006, VIZ-007, VIZ-008
- **verify-later:** first live use; whether the column form reads as a time series to a reader, or wants a line

### VIZ-003 — series facts: the substrate a time axis needs
- **status:** deployed (in `v1.0.1185`, pod-verified); **first live series registered 2026-07-29** — `CIT-tw-leakage-series`, five observations each with its own citation (oufe register, mig 265)
- **status-evidence:** three distinctive string literals from `claims_series.go` found in the running binary 2026-07-28, with a positive control. First-series verification 2026-07-29: 5/5 observations carry their own citation (checked per-observation, the way the gate checks it).
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

### VIZ-010 — `scripts/render_audit.py`: the post-deploy render witness
- **status:** built (2026-07-27, brochure workstream), **run by hand only — NOT wired into anything.** SUPERSEDED IN CAPABILITY 2026-07-29: the same measurement is now dispatchable as an orchestration (VIZ-012); the script's retirement is unblocked but is the brochure workstream's call, not made here.
- **status-evidence:** checked 2026-07-28. Nothing in the Makefile, CI, any shell script, or any `agent_definitions` workflow invokes it. Run against oufe.com's Thames page that day: `contrast=0 broken-img=0`.
- **what:** Renders a live page in headless Chromium and measures what a visitor actually sees. For **every element** (`body *`) it takes the computed colour, walks up through transparent ancestors compositing alpha to find the effective background, and applies 4.5:1 / 3.0:1. A background *image* under text is reported as `overImage` so a reader discounts it rather than trusting a number the page cannot justify. It also reports images that failed to load, which the DB-only `image_url_404` check cannot see.
- **why it matters:** it is the only thing that catches 026's **family 3** — a component hard-coding an ink over a themed fill — which `check_palette_contrast` states in its own header it cannot see by construction. On fundamentallyai.com it found 101 AA failures across 5 pages in about two minutes, on a site where every page said `deployed` and none of ~50 discovery checks had objected.
- **the duplication worth recording:** on 2026-07-28 this workstream built `cmd/contrastscan`, a Go tool doing the same job, without finding this one. **The prior-art grep was `--include=*.go` and the prior art is Python.** The Go tool was deleted on discovery; its one arguably-distinct behaviour (refusing to score an unknown backdrop rather than flagging it) is a stricter variant of `overImage` and not worth a second tool. See `WRONG_CALLS.md` 2026-07-28.
- **sources:** `scripts/render_audit.py`; `platform/orchestration/actions/discovery_checks/check_palette_contrast.go:43,108`
- **relations:** VIZ-011, CLM-001

### VIZ-012 — `render-audit-agent`: the render audit as a dispatchable orchestration
- **status:** **live and exercised** (2026-07-29 — three full runs against oufe.com; the third measured a fix the first had found)
- **status-evidence:** publish `{"action":"orchestrate","config":{"agent_type":"render-audit-agent"},"input_data":{"site_id":…,"domain":…}}` on `system.agent.generic.requests`; the orchestration ran `ensure_site_record` → `request_render_audit` → complete, and findings landed in `collected_data->'render_audit'->'response'`. Recipe: oufe `RUNBOOK_oufe.md` §14.
- **what:** The whole chain, callable per site: chassis action `request_render_audit` (enumerates `pages` where `build_status='deployed'`, caps at `max_pages`, reports truncation) → topic `system.adapter.render-audit.requests` → dedicated pod `render-audit-adapter` (browser-runner image, own consumer group/logs) → Chromium renders every deployed page → `ContrastFinding`/`BrokenImage`/`OverflowFinding` + a summary separating **firm** failures from `over_image` approximations. `pages_failed` counts pages with a firm contrast failure; unreachable pages are reported, not skipped.
- **why it matters:** the only *dispatchable* detector for 026 family 3 (component hard-coding an ink over a themed fill). Proven predictive immediately: its first full-site run found a firm 2.61 white-on-gold on oufe's contact form submit — the shared `contact-form` component's hard-coded ink — fixed and re-measured clean the same morning.
- **landmines:** the seed originally wrote `initial_step` for `start_step`, which makes every dispatch fail as `WORKFLOW_INVALID` replied to a topic nobody reads (016b §9, 2026-07-29). It audits **deployed rows only** — a live page whose row says otherwise is invisible to it (oufe's contact page was, for a day). It does not raise work items from findings (deliberate — `bugs_open/122` candidate 2 needs its own design).
- **sources:** `docs/agent_docs/sql_for_agents/256_render_audit_agent.sql`; `platform/orchestration/actions/request_render_audit_action.go`; `internal/adapters/browserrunner/render_audit_action.go`; deployment `render-audit-adapter`
- **relations:** VIZ-010 (the hand-run predecessor), VIZ-011, CLM-001, `bugs_open/122`

### VIZ-013 — `write_render_audit_findings`: the render audit's drain
- **status:** built + tested (2026-08-02, vigilant_designer_offer_analysis workstream); **inert until an image roll AND the config tail step lands on `render-audit-agent`** (both tracked as that lane's A0.3b — a registered action a workflow never names runs nothing).
- **status-evidence:** 5 sqlmock tests pin the effects (`write_render_audit_findings_test.go`); not yet observed live.
- **what:** Chassis action that files a render audit's FIRM findings as routed `site_work_items`: firm contrast failures → `contrast_failure` at `css-patch-agent` (dedup key `contrast_failure:<page-path>#<selector>`; the NEXT render audit is the de-facto verifier); broken images that resolve to an `assets` row → `undeployed_asset` at `asset-deployer`, co-deduping with `check_undeployed_assets` on `undeployed_asset:<asset_id>` (deliberate shared dedup unit). Items are born `detected`, so promotion stays with `improvement-loop.triage_findings` (migration 286's single owner). Deliberately NOT filed, but counted loudly in the result: `over_image` approximations, overflow (no culprit attribution on this path — Tier-4 `no_horizontal_overflow` owns the attributed case), unattributed broken images (source-side planning gap; `check_content_image_missing`'s rail), and findings whose culprit class appears in a LOCKED component's markup. A run filing more than `max_items` (60) reports `findings_capped`.
- **why it matters:** closes the loop VIZ-012's own landmines section left open ("It does not raise work items from findings") — the design `bugs_open/122` candidate 2 asked for. Until this, the fleet's only browser-measured contrast defects stopped in `collected_data`, the bugs_open/115 shape.
- **sources:** `platform/orchestration/actions/write_render_audit_findings_action.go`; registry entry `write_render_audit_findings`
- **relations:** VIZ-012 (the measurement), VIZ-010, `bugs_open/122`, `bugs_open/115`, migration 286

### VIZ-011 — chart furniture is a graphical object, so the 3.0 threshold applies to it
- **status:** deployed (applied in VIZ-002 and VIZ-006)
- **what:** Axis lines, connectors and bars are not decoration: they are required to understand the content, so WCAG's 3.0 non-text contrast minimum applies. On oufe's live stylesheet `--color-border` (#2E3F52) scores **1.66** against the page background and fails it; `--color-accent` scores **6.86**. Both new components therefore draw their furniture in the accent.
- **why it matters:** the intuitive choice (`--color-border`, because it is a border) is the failing one, and nothing on the platform checks it at build time.
- **sources:** measured 2026-07-28 against the live stylesheet; `bugs_open/122`
- **relations:** VIZ-006, VIZ-002, VIZ-010

### VIZ-014 — legible-ink companions: the renderer names BOTH directions of "is this colour readable"
- **status:** **code written, tested, committed 2026-08-06 — NOT live** (needs an image roll; v1.0.1261 predates it). The CONFIG half — the component/layout repoints that consume it, and the render-audit cadence row — is **not yet applied**, so as of this entry the seam is emitted-but-unconsumed by design.
- **status-evidence:** 11 unit tests green against a clean `git archive HEAD` tree (another session's WIP was breaking the package build in place). Three mutations proved the assertions load-bearing, each with a DISTINCT failure: reorder the renderer calls → the AST ordering test fails; `grounds[:1]` → only the two-grounds test fails; delete the source-unchanged branch → only the no-op test fails. Council `c4d9c841-3658-4742-85b5-961e062ecad2` **APPROVED** at round 2 (round 1 REVISE, gated by `editquality`).
- **what:** `buildLegibleInkDefaults` (`palette_specialised_slots.go`) appends a renderer-owned `:root` block as **step 12** of `RenderCSSFromSpecAction`, after `buildSectionDefaults` (10) and `buildTokenAliases` (11). It emits three custom properties, and the reason there are three is the whole point of the entry — a palette colour used as a FILL and the same colour used as an INK are different questions:
  - `--color-<x>-text` — the ink that goes **ON** an `<x>` fill. Already derived by `darkSchemeDerivations`; `accent_text` has existed since 2026-07-27 and **0 of 18 layouts declare it**, so it had never reached one stylesheet. Emitting it here is what turns that dead derivation live.
  - `--color-<x>-ink` — `<x>` **itself, made legible as an ink on the page**. New. Nothing in the platform computed this before. `--color-primary-ink` and `--color-accent-ink`.
  `legibleInkFor(src, grounds, palette, minRatio)` returns `src` unchanged when it clears `minRatio` against **every** ground, else the first palette colour that does, else the achromatic extreme with the better worst case.
- **why it matters:** 17 of 18 layouts colour an ink with `var(--color-primary)` and nothing checked it. `warnUnusablePrimary` had detected the condition since 2026-07-27 and only logged, because — in its own words — "there is no single correct repair". There is one now, and the warning names it.
- **landmines:**
  - **`grounds` is a SLICE and that is load-bearing.** A component may place one ink on the page *and* on a card — dartsonline does exactly that (`__eyebrow` 1.04 on `background`, `.tl-card-link` 1.07 on the derived `card_bg`). One variable is only right for two grounds if it is right for the WORSE of them. Simplify this to a single ground and the fix becomes a half-fix that looks correct on whichever surface you happen to open. `TestLegibleInkFor_TwoGroundsDisagree` is the only guard.
  - **SCHEME-INDEPENDENT, unlike `fillDarkSchemeSpecialisedSlots` — deliberately, and this is the architecture fork the council asked be recorded.** That function is dark-only because a *light literal* is only wrong on a dark site. "This ink is illegible on this ground" is true or false regardless of scheme, and two of the three accent-direction sites are LIGHT (gaswholesalers `#F4F1EB`, finetuning `#F5F3EF`). **If a future edit re-narrows this behind an `isDarkHex` guard, the fix silently stops covering its own measured cases.** `TestBuildLegibleInkDefaults_LightSchemeStillEmits` fails and says so.
  - **Position is load-bearing and is now asserted, not commented.** It skips any name the ASSEMBLED CSS already defines, so running it before step 11 lets a later block define the same name and win — emitting a variable that is present and inert, the dead-config shape the seam exists to avoid. `TestRenderCSS_InkCompanionsComeAfterTokenAliases` reads the **AST**, not the source text: a regex would be satisfied by the order of the *comments*.
  - **Every consumer must be written `var(--color-X-ink, <exactly today's value>)`, never bare.** An empty fallback makes the whole declaration DROP rather than degrade. `TestBuildLegibleInkDefaults_NeverEmitsAnEmptyOrIndirectValue` pins the emission side; the consumer side is the migration's job.
- **sources:** `platform/orchestration/actions/palette_specialised_slots.go` (`legibleInkFor`, `worstRatioAgainst`, `buildLegibleInkDefaults`); `platform/orchestration/actions/render_css_from_spec_action.go` step 12; tests in `palette_specialised_slots_test.go` and `render_css_from_spec_action_test.go`; council `c4d9c841-3658-4742-85b5-961e062ecad2`
- **relations:** VIZ-010, VIZ-012, VIZ-013 (the detector for this defect class, still undispatched), `bugs_open/122`, `bugs_open/211`, `bugs_open/212`
