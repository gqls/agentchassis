
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Styling render pipeline reference: two assembly paths and the scheme gap
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** 036 FINDING/THEORY-tagged reconstruction from code + live data
- **what:** Stylesheet and page-section renders are separate code paths meeting only in the browser via class names/custom properties. Key FINDINGS: resolved_composition doesn't record scheme (survives only on layouts.scheme); buildSectionDefaults emits --section-* only for dark bg/surface (light sites correctly get nothing); five surface classes are duplicated renderer+layouts (Phase 4.5 debt); hero/CTA components hardcode dark backgrounds + literal white text defeating the scheme; .{function}-section class contract broken by hero (.hero) and CTA (.cta-section); four overlapping chrome default stores (style_collections ids [live read], site_components slots [likely superseded], sites.default_components, layouts.default_* [all NULL]); RenderFallbackHeader is hardcoded dark; SectionStyles/component_selector are dead on the current path. Fix direction (Q4a): strictly variable-driven components + renderer-owned per-section --section-*.
- **sources:** 036 full; 016b light-site-dark-chrome entry
- **relations:** section painting contract; site component linkage; scheme-to-components runbook
- **verify-later:** F-thread confirmations (update_site_defaults on composition path)

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### CSS assembly pipeline (composable theme → styles.css)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "fully built path" (2026-05-12); render_css_from_spec_action.go deterministic, verified live schema for css_snippets
- **what:** webdesign-agent: analyze_design (LLM) → render_css_from_spec (deterministic Go: theme composition from palettes/layouts/typography_sets FKs, css_snippets matched via applies_to JSONB overlap against the site's component functions, dark-section variants) → deploy_css git commit to assets/css/styles.css → B2 CDN sync. css-patch-agent is the bypass path for one-off fixes (patches the deployed file directly, not the snippet library).
- **sources:** FOCUS-css_js_mechanisms.md#1; HANDOFF_2026-04-18_design_and_styling…md
- **relations:** composable theme migration 025; site-design-planner
- **verify-later:** render_css_from_spec_action.go, render_css_composition_helpers.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### JS three-path model (js_content deployed, js_snippets loader missing, inline legacy)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "declared in contracts, table populated, BUT NO LOADER IS WIRED UP" (verified 2026-05-12: 9 js_snippets rows, no reference in head templates or RenderHead)
- **what:** Path A (deployed): per-component JS in content_components.js_content, extracted at store time by separateInlineJS(), deployed as /tools/assets/{function}.js via collectJSAssets() multi-file git commits. Path B (aspirational): js_snippets shared utility table with applies_to scoping — a registry of intentions with no runtime loader; contracts' "loaded via head component" claim is aspirational. Path C (legacy anti-pattern): inline <script> baked into html_template, violating contract 003 — news components still there. Path D: html-assembler's inject_js flag has no visible reader. Interim tactic: insert the snippet row AND duplicate inline until the loader exists.
- **sources:** FOCUS-css_js_mechanisms.md#2, #3, #4; HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#2
- **relations:** component contract 003; JS separation deployment (2026-04-17)
- **verify-later:** RenderHead in component_library.go; js_snippets rows; whether a loader was ever built

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Component quality tracking (quality_score et al.)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "migration_component_quality.sql (applied) … compute_component_quality_action.go (pending Go deploy)" (2026-04-17); quality scoring described as working in the 04-18 handoff status table
- **what:** content_components gains template_variable_count, schema_field_count, template_closed, schema_template_synced, has_data_component, quality_score (100 minus deductions), quality_issues; scored inline on store and by a component-quality-auditor agent; planner prefers high scores, auditor targets low ones for regeneration. Backfill via system.internal work item. 43 pre-existing components had 0 template variables (content baked in) — regeneration targets, not mass-deletable.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#6, #Architecture-decisions
- **relations:** pre-store validation gates; component-creator prompt tiers
- **verify-later:** compute_component_quality registry entry; quality_score population in DB

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Pre-store component validation gates + planning deferrals + empty-section filter
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** deployed 2026-04-17 (three checks before INSERT; sectionHasVisibleContent; empty-schema deferral); root incident: max_tokens=4000 truncation left unclosed <section>, CSS rendered as page text on vonc.com
- **what:** Three layers preventing broken components/sections reaching pages: store-time rejection (template must contain <section>/<div>, balanced <style> tags, non-empty input_schema), plan-time deferral of content-type components with empty schemas, and render-time skipping of sections with <10 chars visible text. Component-creator max_tokens raised 4000→16000 and prompt made context-aware.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#1, #6; HANDOFF_2026-04-17_nav_empty_sections_footer(1).md#1, #6, #7
- **relations:** LLM reliability tracks; quality tracking
- **verify-later:** store_generated_component_action.go validation block; rerender_single_page_action.go sectionHasVisibleContent

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Scheme-to-components P0: light-resolved site renders dark
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** PLAN "## CLOSED (2026-07-03) … closed on deployed evidence (RUNBOOK §SCHEME CLOSE: all nine grep checks pass; the stale-section fossil `var(--accent-color` is gone)"; RUNBOOK §SCHEME CLOSE lists the nine counts (site-header-section 32 / gradient 0 / footer 37 / --hero-ink 13 / color-mix 14 / cta pair consumed / white rgba 0 / fossil 0 / brief-explanation 0-expected).
- **what:** The defining P0 of this thread: the chassis resolves each site to a light or dark scheme and the scheme travels correctly through layout and palette variables, but the component library was written dark-first — components hardcode white text and dark `--section-*` context inline, so a light-resolved site (idea.uk, `tool-portal-light`) deployed dark chrome and dark sections. The winning mechanism was completion of the existing paired-variable standard rather than restructure: one layout patched, ten templates de-hardcoded, chrome repointed + force-rerendered, then a full page-build-handler rebuild.
- **sources:** PLAN_scheme_to_components(1).md#CLOSED; RUNBOOK_scheme_to_components(50).md#SCHEME-CLOSE; HANDOFF_scheme_to_components_for_claude_code(1).md#The-problem; running_notes_scheme_to_components(55).md#Tk
- **relations:** paired-variable standard; hero ink model; hazard/band declarer taxonomy; rebuild vs rerender semantics; chrome selection path.
- **verify-later:** deployed idea.uk B2 index.html greps; `content_components.html_template` for site-footer/call-to-action/hero; `layouts.css_template` for tool-portal-light.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Three-part styles.css assembly and palette merge rules
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sc), read from `render_css_from_spec_action.go` bodies in full, 2026-06-30; HANDOFF §Established mechanism restates it as verified.
- **what:** `RenderCSSFromSpecAction` builds styles.css in three appended parts: (1) the layout `css_template` rendered as a Go text/template with `{{palette}}`/`{{typo}}`/`{{token}}` FuncMap helpers over merged maps — merge rules: 8 core palette slots spec-wins, specialised slots theme-wins, typography spec-wins, structure layout-only; (2) component CSS from the `css_snippets` table (name, css_content, applies_to jsonb) where applies_to overlaps the site's components — a third CSS surface distinct from inline `<style>` (C3 cleared it of dark-section treatment: all 21 snippets are utilities); (3) the `buildSectionDefaults` luminance block. Theme composition loads via style_collections → css_themes joined to palettes/layouts/typography_sets, hard-erroring on NULL FKs.
- **sources:** running_notes_scheme_to_components(55).md#Sc #Se; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; PLAN_scheme_to_components(1).md#Confirmed-at-code-level
- **relations:** buildSectionDefaults; layout CTA pair curation; scheme derivation.
- **verify-later:** render_css_from_spec_action.go, render_css_composition_loader.go/_helpers.go; `css_snippets`, `palettes.colours`, `typography_sets` tables.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### buildSectionDefaults: luminance-keyed dark-only --section-* defaults
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sf) 2026-07-01: "`--section-*` is a DARK-ONLY override; light is the fallback. `buildSectionDefaults` returns '' unless bg or surface is dark."
- **what:** The renderer's only live per-section adaptation: `buildSectionDefaults` (color_util.go, WCAG `isDarkHex`/`pickReadableOnBackground`) emits a `body { --section-* }` block only when the merged palette's background or surface is dark, plus a dark-surface variant on 5 hardcoded surface classes (`.features/.services/.differentiators/.about/.faq-section`). On a light palette it emits nothing and element rules fall through to `var(--color-*)`. Retained unchanged under the paired-variable decision as the whole-palette-darkness base/safety net.
- **sources:** running_notes_scheme_to_components(55).md#Sf; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; SPEC_scheme_to_components.md#Decision-record
- **relations:** Colour Inheritance Model; Phase 4.5 deferral (generalises the 5-class list); paired-variable standard.
- **verify-later:** color_util.go buildSectionDefaults/isDarkHex; emitted styles.css tail on a dark-palette site.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### SectionStyles per-section CSS mechanism (built, disconnected, retired)
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** Notes (Sf): "`SectionStyles` is DEAD for current sites. None of the 18 active layouts reference `{{range .SectionStyles}}` … computed-but-unused"; SPEC: "`SectionStyles` stays retired."
- **what:** A fully-built but never-connected renderer mechanism: `queryDarkSectionsForCSS` + `buildCSSsectionStyles` compute per-component `{Function, ClassName: function+"-section", IsDark}` entries from `content_components.is_dark_section` (fallback dark list hero/social-proof/call-to-action/testimonials) and pass them to the layout template — which no active layout consumes. Considered as the cheap renderer-owns vehicle (Alt B) and explicitly retired by the paired-variable decision. A textbook infrastructure-orphan: ~80% built, deliberately not revived.
- **sources:** running_notes_scheme_to_components(55).md#Sf #Si; SPEC_scheme_to_components.md#Decision-record; HANDOFF_scheme_to_components_for_claude_code(1).md#Established
- **relations:** superseded by paired-variable standard; related Phase 4.5 (the other renderer-owns design).
- **verify-later:** render_css_from_spec_action.go buildCSSsectionStyles/queryDarkSectionsForCSS still present and uncalled from layouts.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Hero ink model and the structural-dark exception
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sv) "W3b COMPLETE (UPDATE 1; … ink in both inline branches, layered solid+single-hue color-mix gradient, five ink-referencing section vars…)"; W3d extended it to the five hero-* variants (UPDATE 5).
- **what:** Image/layered sections define a per-branch `--hero-ink` custom property and derive all text/context from it: the image branch sets `--hero-ink:#fff` under the structural-dark exception (an `rgba(0,0,0,x)` overlay guarantees darkness, so white text is always safe); the no-image branch sets `--hero-ink: var(--color-primary-text)` over a layered `var(--color-primary)` solid plus a single-hue gradient mixing 15% toward the ink (depth on both dark and light primaries; the solid layer doubles as the color-mix-less fallback). Buttons become the inverse pair (ink background, primary label). Chosen after data showed imageless heroes are the common case (80/114 hero, 26/26 hero-*), and it fixed a latent white-on-cyan failure on tool-portal-dark.
- **sources:** running_notes_scheme_to_components(55).md#St #Su #Sv #Sw; w3b_01_hero_conversion.sql; RUNBOOK_scheme_to_components(50).md#HERO-(c)-DESIGN
- **relations:** section painting contract model (c); paired-variable standard.
- **verify-later:** hero + hero-* `content_components.html_template` current bytes; rendered index hero.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Hazard-class vs band-class declarer taxonomy (library blast radius)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** CHECK 3 RESULTS (2026-07-02): 84 active sections — 15 hex backgrounds, 37 self-declare `--section-*`, split ~18 hazard vs ~19 band; SCHEME CLOSE remaining work item 4: "~10 remaining surface-painting declarers + ~17 band-class components (non-idea.uk)" still open.
- **what:** The diagnostic taxonomy that sized every fix decision: hazard-class components declare dark `--section-*` while painting surface variables or nothing — live white-on-light bugs today (the footer, site-head, the five hero-* variants, brief-explanation etc.); band-class components paint from primary/secondary/accent with white text — coherent today but blocking "fully light" (CTA, hero, social-proof, testimonials…). Ten templates (the idea.uk-visible set) were fixed by hand-needles; the non-idea.uk tail awaits the re-aimed fixer (Step D decision).
- **sources:** RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS #SCHEME-CLOSE; SPEC_scheme_to_components.md#W2 #W3; running_notes_scheme_to_components(55).md#Sn
- **relations:** fix_forced_text_colours re-aim (the tail vehicle); supervised fixer first-run.
- **verify-later:** re-run the 3c split query; count remaining literal declarers among active sections.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Chrome selection path and the dead header_component_id column
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sl): "`install_site_composition` sets `style_collections.header_component_id`/`footer_component_id` = NULL … grep finds NO code that writes them non-NULL … effectively a DEAD column"; HANDOFF §Established restates the chain with line numbers.
- **what:** Page-compile chrome resolution: `CompilePageSectionsAction` → `InjectHeader/InjectFooter/InjectHead` → `RenderHeader/RenderFooter` reads `style_collections.header_component_id` (always NULL — inserted NULL with a "webdesign-agent populates these later" comment and never written) → falls to `GetComponentByFunction("site-header")`, the single library-wide active component per function → else the hardcoded-dark fallback. `RenderHead` looks up function `head` (the only head component is inactive, so builds always used the fallback head); `site-head` is section-level and unreachable as chrome. Five other active header/footer functions (`*_pre_037`) are unreachable on this path. The one-active-component-per-function convention holds for sections by data (C4: no function has >1 active) though the UNIQUE index only covers tools.
- **sources:** running_notes_scheme_to_components(55).md#Sl #Sq #Se(C4); HANDOFF_scheme_to_components_for_claude_code(1).md#Established; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3d)
- **relations:** four chrome default stores; scheme-aware fallback chrome; dual chrome render paths.
- **verify-later:** component_library.go RenderHeader/RenderFooter/RenderHead/GetComponentByFunction; install_site_composition_action.go NULL insert + comment fate (W4c chose deleting the comment).

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Scheme-aware fallback chrome (RenderFallbackHeader/Footer consume the pairs)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Ud) 2026-07-06: "slice 2+F live"; RUNBOOK 07-06-night State: "Deployed: slices 1 …, 2 (fallback chrome C/D + Debug tidy E)".
- **what:** The safety-net chrome functions hardcoded dark (`background: ctx.PrimaryColor` default `#1a1a2e`, literal white text) — so any site whose chain broke got dark chrome regardless of scheme. Edits C/D replace the whole functions: backgrounds become `var(--color-header-bg, var(--color-surface))`/footer equivalent, text `var(--color-header-text, var(--color-text))`, muted/borders via `color-mix` — safe library-wide because Check 3e proved all 18 layouts set all four chrome vars. `RenderFallbackHead` deliberately unchanged (its only colour use is a `<meta theme-color>` value where `var()` cannot work). Edit E swapped the file's eight `logger.Debug` calls to Info per the no-Debug rule.
- **sources:** gobatch_02_component_library.md; running_notes_scheme_to_components(55).md#Sl #Tq #Ud; SPEC_scheme_to_components.md#W4(a)
- **relations:** chrome selection path; paired-variable standard; no-logger.Debug convention.
- **verify-later:** component_library.go RenderFallbackHeader/RenderFallbackFooter current bodies; deployed image tag containing them.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Dual chrome render paths and repoint-before-force_rerender ordering
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sy): "render_site_components_action.go:345–430: pinned-component join has NO is_active filter; non-force runs SKIP non-empty slots → repoint BEFORE force_rerender or the old dark chrome re-renders"; W4b executed and verified 2026-07-02 (header 3750→6258B, footer color-mix in).
- **what:** Chrome has two render paths: the page-compile path renders fresh via RenderHeader/Footer, while `render_site_components` writes `site_components.rendered_html`, which the RERENDER handler injects into pages. The pinned join ignores `is_active`, and without `force_rerender` non-empty slots are skipped — so stale renders of deactivated components persist indefinitely. The W4b remedy pattern: repoint `site_components.component_id` to the active components (guarded on the known old ids; `rendered_html` deliberately left in place so there is no chrome-less window), then trigger rerender-pages v6 with `spec.refresh_site_components: true`. A deliberate side-effect became a staging technique: the chrome-refresh deploys the whole site as an intermediate visual checkpoint (light chrome over old sections) before the full rebuild.
- **sources:** running_notes_scheme_to_components(55).md#Sy #Sz #Ta #Td; w4b_01_repoint.sql; w4b_04_trigger_item.sql; RUNBOOK_scheme_to_components(50).md#W4b-RESULTS
- **relations:** rerender-pages v6 workflow; rebuild vs rerender semantics; four chrome default stores.
- **verify-later:** render_site_components_action.go force_rerender/skip logic; idea.uk site_components rows point at active components.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Phase 4.5 data-section-bg surface generalisation (deferred)
- **category:** styling-render-pipeline
- **status-signal:** aspirational
- **status-evidence:** SPEC consequences: "025 Phase 4.5 (`data-section-bg` surface generalisation) is deferred as a separate dark-site concern."
- **what:** Doc 025's already-designed decouple: components carry a `data-section-bg="surface"` attribute; the renderer replaces its hardcoded 5-surface-class list with an attribute selector; dual-write migration. The thread audited it seriously (Si prior-art pass) and then argued it down: it solves a dark-site generalisation idea.uk never hits, its blanket "never self-declare" conflates hazardous surface declarations with load-bearing band declarations, and renderer ownership reintroduces component intent one hop away. Remains the designed answer for dark sites with surface sections outside the hardcoded 5, if that ever bites.
- **sources:** running_notes_scheme_to_components(55).md#Si #Sk #Sm; SPEC_scheme_to_components.md#Decision-record; HANDOFF_scheme_to_components_for_claude_code(1).md#Questioning-025
- **relations:** buildSectionDefaults (the 5-class list it generalises); paired-variable standard (chosen instead).
- **verify-later:** docs 025 §427–505; any data-section-bg attribute usage in components.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Explicit RenderContext.Scheme signal (Q1)
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** Notes (Sf): "explicit `RenderContext.Scheme` is SECONDARY … This revises the Q1 emphasis in the PLAN"; never implemented anywhere in the executed fix.
- **what:** The original leading design (Q1/Q3): plumb the resolved scheme explicitly into both render entry points — `l.scheme` in the CSS loader SELECT + `themeComposition.Scheme`, and a `Scheme` field on `RenderContext` exposed via `contextToInterfaceMap` — so component templates receive a light/dark signal. Overtaken when Check 1 showed the scheme already reaches components implicitly through the palette `:root` values and luminance defaults; the components were the only thing defeating an already-working system. No scheme field was ever added.
- **sources:** PLAN_scheme_to_components(1).md#Q1; running_notes_scheme_to_components(55).md#Sb #Sf #Sk
- **relations:** superseded by paired-variable standard + implicit palette mechanism.
- **verify-later:** RenderContext struct (component_library.go) — confirm no Scheme field exists.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Exact-field-name template binding with silent empty on miss (RenderTemplate `<no value>` strip)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "Correction 2 — the silent-empty mechanism is in RenderTemplate" (NOTES §2, confirmed from uploaded component_library.go).
- **what:** `RenderTemplate` (component_library.go) binds a page's `content_data` into a component's `html_template` by exact field name via Go `text/template`, then strips the `<no value>` tokens of unmatched placeholders to empty string, logging only a warning — no error. This is *why* a renamed or missing field fails silently rather than loudly; the entire clobber class rests on it.
- **sources:** NOTES(43).md §2 Correction 2, §8; BUNDLE(3).md §1
- **relations:** clobber failure mode; F1 guard (compensating control); fail-loud guard route (never built as such).
- **verify-later:** platform/orchestration/actions/component_library.go:RenderTemplate; the `<no value>` cleanup.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Section readiness model (planSection source tiers, required/fallback semantics, spec resolver)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "planSection required semantics CONFIRMED from plan_sections_action.go… 'Required with no fallback — defer' → deterministic defer → carry" (NOTES §9w).
- **what:** Section fields declare a source (static with fallback; llm; Tier-C spec paths like `site_specs.cta.primary_url`) plus required/fallback attributes. planSection resolves each non-LLM field; a required field with no resolvable source and no fallback defers the section (→ carry). The resolver reads site_specs per-aspect rows (`aspect`,`data` jsonb, is_current; `resolveSpecPath("cta.primary_url") = specs["cta"]["primary_url"]`), checks presence not validity, and the stored⊕resolved merge persists resolved values into content_data at render time. Tier-C fields are by design never content_data keys.
- **sources:** NOTES(43).md §9s, §9u–§9x; RUNBOOK(49).md Part A
- **relations:** carry-forward path; F5; stored⊕resolved merge; phantom-CTA lesson (spec presence ≠ URL validity).
- **verify-later:** plan_sections_action.go (ensureSpecs, resolveSpecPath, on_missing switch); site_specs schema.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Stored⊕resolved merge writes resolved values back into content_data
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "resolver persists cta_url into content_data (stored⊕resolved merge — expected)" (NOTES §9w); robot-hands cd_keys gained merged fallback keys (§9an).
- **what:** When a section re-render resolves fields (spec values, static fallbacks), the merged result is persisted back into the page's `content_data` as well as baked into `rendered_html`. Double-edged: it makes recoveries durable (cta_url landed in content_data), but it is also F8 carrier 2 — contaminated fallback values were merged into dependents' content_data, surviving the later schema fix and needing an explicit key-strip.
- **sources:** NOTES(43).md §9w, §9an, §9ao; HANDOFF(7).md §Incident 2
- **relations:** F8 contamination; section readiness model; recovery playbook.
- **verify-later:** rerender_page_sections persist-merged-content step.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Visible-content filter drops near-empty bands at assembly
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "sections = page_components by page_id ORDER BY position; visible-text filter drops bands with ≤10 chars of stripped text (logger.Warn per skip)" (NOTES §9ai, from rerender_single_page_action.go).
- **what:** Page assembly filters out any band whose rendered HTML strips to ≤10 visible characters, logging only a warning. It is the final silencer in the clobber chain (an emptied section vanishes rather than erroring) and produces counter-intuitive interims — the F8 "neutral shell" bands survived the filter because two neutral CTA labels exceeded 10 chars.
- **sources:** NOTES(43).md §1, §9ai, §9ar; BUNDLE(3).md §1
- **relations:** clobber failure mode; assembly membership model; fail-loud guard route (unbuilt alternative).
- **verify-later:** rerender_single_page_action.go visible-text filter.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Assembly membership and chrome model (page_components by position; pages.sections is metadata)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "assembly membership = page_components by position… pages.sections jsonb is NOT assembly membership" (NOTES §9ai — corrected a wrong sections_listed=0 inference); three head shapes identified.
- **what:** A page artifact is assembled from `page_components` rows ordered by position (not from `pages.sections`, which is planning metadata); head/header/footer come from site-scoped `site_components` rows; with no stored head, `buildDefaultHead` emits a 5-line head linking /assets/css/styles.css. A third, legacy builder (`assemble_from_library`, theme CSS from css_themes) produced older artifacts with big inline heads — three coexisting head shapes that repeatedly confused artifact forensics. Also: `data-component` attributes exist on only some templates, so artifact greps on them undercount bands (owned metric artifact).
- **sources:** NOTES(43).md §9ah–§9al; RUNBOOK(49).md Part B
- **relations:** visible-content filter; R6f vocabulary drift (missing :root head); chrome refresh gating.
- **verify-later:** rerender_single_page_action.go; assemble_from_library (registry L493); site_components schema.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### R6f — theming vocabulary drift (defined vs consumed CSS custom properties)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "R6f confirmed as the 'renders badly' mechanism; fresh rebuilds render WORSE" (NOTES §9bh); mechanism narrowed to defined-vs-consumed drift §9am–§9bi.
- **what:** Component templates consume CSS custom properties (`var(--x, fallback)`) whose names drift from what the site's generated styles.css `:root` defines — 11 gap names in two patterns (synonyms like --border-radius vs --radius, and orphans like --hero-ink). Sections on undefined vocabulary render via per-component fallback values — a "fallback lottery" that goes dark-on-dark invisible on dark canvases (gripper-detail's blank page). Newer generators put :root in styles.css (rootless heads), older sites carry inline :root heads (why leopardess was immune). Every fresh rebuild worsened it as new templates minted new names.
- **sources:** NOTES(43).md §9al, §9am, §9bh, §9bi; RUNBOOK(49).md Part D; HANDOFF(7).md §R6f
- **relations:** D2a token aliases (fix); D2b prevention; deterministic styles.css rendering; assembly head model.
- **verify-later:** site styles.css :root contents vs template var() usage; robot-hands/vonc pre-fix stylesheets.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### D2a — buildTokenAliases renderer-enforced compatibility bridge
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "D2a VERIFIED in production 2026-07-06 — step 11 emits the alias block on a real gamesdesign pass" (NOTES §9bn); curl shows the trailing compatibility-aliases :root block.
- **what:** A step-11 post-pass in RenderCSSFromSpecAction (mirroring step 10's buildSectionDefaults "renderer-enforced" pattern): after rendering, append a trailing `:root` block defining ONLY the missing names from a package-level 11-entry alias table (synonyms → canonical var() references, orphans → palette-safe literals). Definition detection by `name+":"` so var() usages and sibling names don't count; idempotent; layout-defined values always win; one zap log field (token_alias_length). Sites self-heal on their next design pass; verified live via an adapted 076 webdesign trigger on gamesdesign.
- **sources:** NOTES(43).md §9bj–§9bn; RUNBOOK(49).md Part D D2a
- **relations:** R6f (the drift it bridges); D2b (prevention side); buildSectionDefaults (pattern precedent).
- **verify-later:** render_css_from_spec_action.go buildTokenAliases + tokenAliases table; render_css_from_spec_alias_test.go.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### D2b — canonical-token prevention (contract rule 11 + AuditTemplateTokens lint + prompt rule)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "D2b in progress: lint coded… contract rule 11 drafted; prompt edit pending agent-identify" (NOTES §9bq, 2026-07-06 — thread ends here).
- **what:** Stops new orphan tokens at the source, reuse-first: (1) contract rule 11 in 003's New Component Checklist — templates reference only canonical tokens + sanctioned aliases, invent no new var(--…) (drafted as a paste-in patch); (2) the generating agent's prompt enforces the rule in place (agent identification via default_config ILIKE still pending — 21 candidates); (3) `AuditTemplateTokens` warn-only lint appended to component_validation.go (canonicalCSSTokens allowlist = 39 theme names + 11 aliases, first-seen dedup, never rejects — detection net not gate, since vocabulary evolves), pending call-site wiring where ValidateComponentTemplate already runs. Notable design subtlety: rule 11 is the reciprocal of checklist item 6 (dark sections must SET --section-*) — consume-side vs set-side.
- **sources:** NOTES(43).md §9bo–§9bq; RUNBOOK(49).md Part D D2b
- **relations:** D2a (defines the sanctioned alias set); F8 lint (sibling detection-net); contracts-and-standards checklist.
- **verify-later:** component_validation.go AuditTemplateTokens; whether wired into StoreGeneratedComponentAction; 003 doc rule 11 presence.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### buildSectionDefaults — luminance-keyed dark-only section context
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "--section-* is a DARK-ONLY override; light is the fallback. buildSectionDefaults returns '' unless bg or surface is dark" (running_notes Sf, from color_util.go).
- **what:** The renderer-owned `--section-*` emitter: on a dark background/surface palette it emits a body-level (and 5 hardcoded surface-class) block of readable section text vars via WCAG helpers (isDarkHex, pickReadableOnBackground); a fully light site gets nothing and element rules fall back to `--color-*`. It is the live half of the section-context mechanism and the pattern precedent for D2a's step-11 alias block.
- **sources:** running_notes(22).md Sc, Sf; RUNBOOK_scheme_to_components(18).md D1 resolution
- **relations:** SectionStyles (dead sibling); paired-variable direction (keeps this as base); D2a.
- **verify-later:** color_util.go buildSectionDefaults/isDarkHex; the 5-class list vs 025's data-section-bg plan.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### SectionStyles — the dead per-section styling mechanism and the uneven {function}-section contract
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** "SectionStyles is DEAD for current sites — none of the 18 active layouts reference {{range .SectionStyles}}… computed-but-unused" (running_notes Sf); "SectionStyles stays retired" (Sn).
- **what:** A built-but-disconnected mechanism: queryDarkSectionsForCSS + buildCSSsectionStyles compute per-section entries (ClassName = function + "-section", IsDark from is_dark_section) for layout templates that no active layout consumes. The `{function}-section` class contract it assumes is real but honoured unevenly (hero emits `.hero`, CTA `.cta-section`). Decision: do not reconnect it — Phase 4.5's data-section-bg attribute keying and then the paired-variable direction supersede it.
- **sources:** running_notes(22).md Sc, Sf, Si, Sn; RUNBOOK_scheme_to_components(18).md D1
- **relations:** buildSectionDefaults (the live half); is_dark_section demotion; paired-variable direction (superseding).
- **verify-later:** render_css_from_spec_action.go buildCSSsectionStyles — still computed-and-unused?

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Hero ink model (per-branch --hero-ink with structural-dark exception)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "W3b COMPLETE… ink in both inline branches, layered solid+single-hue color-mix gradient" (running_notes Sv, run 2026-07-02 16:43); W3d extended it to the five variants.
- **what:** The hero's contrast contract after conversion: each branch sets an ink variable — image branch `--hero-ink:#fff` (the rgba overlay guarantees darkness: the sanctioned structural-dark exception); no-image branch `--hero-ink: var(--color-primary-text)` over a layered solid + single-hue color-mix gradient (15% toward the ink; solid layer is the color-mix-less fallback). Section vars derive from the ink at preserved alphas; buttons become the inverse pair. Fixed a latent white-on-cyan failure on the dark portal, not just the light-site problem. Imageless heroes turned out to be the COMMON case (80/114 + 26/26), reversing an assumption.
- **sources:** running_notes(22).md St–Sw; RUNBOOK_scheme_to_components(18).md HERO (c) DESIGN, W3b/W3d RESULTS
- **relations:** paired-variable direction; ambient pass-through (sibling pattern); D2a (--hero-ink later an alias orphan).
- **verify-later:** hero + hero-* html_template current state; whether rebuilds shipped the converted renders.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Ambient pass-through pattern for surface painters with fallback-less consumers
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "Sanctioned pattern recorded: page/surface painters with fallback-less consumers pass the ambient context through" (RUNBOOK_scheme(18) W3e, executed 2026-07-02 17:22).
- **what:** For components that paint the page/surface colour but whose internal rules consume `var(--section-*)` without fallbacks, the safe conversion is declaring `--section-x: var(--color-x)` pass-throughs rather than deleting the declarations (deletion would fall to currentColor/transparent). Scheme-correct on both light and dark by definition since the core vars ARE the scheme. Companion finding: `rgba(var(--hex-var), α)` is invalid CSS that never rendered — color-mix is the working replacement.
- **sources:** RUNBOOK_scheme_to_components(18).md W3e RESULTS; running_notes(22).md Sx
- **relations:** paired-variable direction; creator-prompt fallback mandate (future components shouldn't need this).
- **verify-later:** brief-explanation template pass-through block (NB: later regenerated in the F8 saga — check current state).

<!-- SOURCE: U08_travelling_docs.md -->
### Interactive-section clobber + interactivity-aware save guard (preserve-sections)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** 016b Part 4 "CAUSE CONFIRMED; FIX PENDING" → 2026-06-24 "fix WRITTEN (un-deployed)"; still listed as the pending multi-page prerequisite in RUNBOOK §5.4 at unit close.
- **what:** Any full rebuild regenerates a page from plan_sections, which knows nothing of an interactive tool stored only as a section's rendered_html — the game is silently discarded (detool-on-rebuild). Layered fix, both layers in `save_page_sections` (the only place holding the markup): (1) interactivity guard blocking a non-interactive set replacing a deployed interactive one; (2) carry-forward of existing interactive sections; plus source_item_id stamping. The unstated invariant "interactive sections survive every rebuild route" is the canonical example of what pipeline PLAN invariants should record.
- **sources:** 016b_debugging_guide_7_3_(7).md#open-threads-part-4; RUNBOOK_travelling_docs(38).md#§5.4; PLAN_travelling_docs(6).md#tool-assurance
- **relations:** multi-page prerequisites; pipeline documentation model; page build/rerender pipeline threads (Parts 1–5).
- **verify-later:** whether the patched save_page_sections_action.go deployed; page_component_history source_item_id.

<!-- SOURCE: U09_adoption.md -->
### Theme-layer render resolution (style_collection → css_theme, specs not read)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "Confirmed in the chassis actions: render context sets PrimaryColor/AccentColor/SecondaryColor from the resolved collection… Nothing in that path reads a site_specs design aspect" (FOCUS_design_composition, 2026-05-26).
- **what:** The live render path resolves colour/typography exclusively from `sites.style_collection_id → style_collections → css_themes`; site_specs design aspects influence it only upstream via the composition resolver. A NULL style_collection_id means no palette at all in render — the expected outcome for build-site-planner sites before the emit_design fix. Related earlier symptoms: dead BEM design-system CSS shipped in every head and no theme CSS variables defined (visual coherence coming from LLM-picked fallbacks).
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#1, old2/HANDOFF_2026-05-07(1)#8
- **relations:** emit_design_items; site-design-planner (doc 027, no-LLM thin wrapper over createPalette)
- **verify-later:** render context construction reads (~chassis 13548, 17464, 27934); theme variable injection status

<!-- SOURCE: U09_adoption.md -->
### Render-off-build_status debt (planned-vs-rendered diff)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "Render decision off build_status — the page-rerender agent appears to skip rebuilding planned-but-unrendered sections on a deployed page. Proper fix: drive render off a planned-vs-rendered diff… Until then the build_status='needs_rebuild' reset is the workaround" (HANDOFF_2026-06-09, parked).
- **what:** A `needs_page` rebuild of an already-deployed page completes without rebuilding planned-but-missing components; the render short-circuits on `build_status='deployed'` instead of diffing planned sections against current page_components. Workaround (reset to needs_rebuild) is proven; the structural diff-driven render is open debt.
- **sources:** HANDOFF_2026-06-06#resolved (14q), HANDOFF_2026-06-09#later-parked
- **relations:** positive-evidence principle (same "derived flag can lie" family); pages.sections vs page_components distinction
- **verify-later:** page-rerender agent def + assemble/deploy action short-circuit

<!-- SOURCE: U13_docs024_small_dirs.md -->
### js_snippets site-level rendering pipeline (site-asset-renderer)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** guidelines_compliance_check(1).md test plan table + "Migration A/B/C applied" checklist; walked against dev-guide/architecture/contracts docs as a deliberate review before shipping
- **what:** `render_js_snippets_for_site_action.go` + `site-asset-renderer` agent (4-step workflow) implements the previously-missing JS-snippet deploy step: `js_snippets` table (global, is_active flag) → matched against a site's component functions via `applies_to` → concatenated → `assets/js/snippets.js` per site, loaded via a `<script>` tag injected into the head template. Mirrors `render_css_from_spec`'s existing pattern exactly.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#The-JS-snippet-renderer-deliverable, js_snippets_news_gaswholesalers/old/guidelines_compliance_check(1).md
- **relations:** CSS component-list fallback bug; CSS applies_to granularity mismatch; component contract 003 (JS split)
- **verify-later:** js_snippets table (9 rows, 6 dormant is_active=false), site-asset-renderer agent_definition, head template script tag

<!-- SOURCE: U13_docs024_small_dirs.md -->
### CSS component-list fallback bug (fake 5-item list masking real component inventory)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "Cause 1 (fallback fires) was fixed 2026-05-16" — status filter fix applied and verified across two sites
- **what:** `extractCSSComponents` falls back to a hardcoded 5-item list whenever `site_context.all_component_functions` is empty. That field was empty because `loadPagesWithComponents`'s status filter matched nothing — every page's actual `status` value is `'active'`, never in that list. Fixed to `WHERE p.status = 'active'`.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_visual_pipeline_css_and_component_lists.md#Cause-1, js_snippets_news_gaswholesalers/old/design_actions_status_filter_fix.md
- **relations:** Assumed-status-values trap; CSS applies_to granularity mismatch
- **verify-later:** render_css_from_spec_action.go extractCSSComponents, design_actions.go loadPagesWithComponents

<!-- SOURCE: U13_docs024_small_dirs.md -->
### CSS applies_to granularity mismatch (known issue, unfixed)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "Cause 2 ... known issue, not yet fixed" — only 2 of ~21 snippets actually match real sites
- **what:** Even after the fallback-list bug is fixed, `loadComponentCSSSnippets` does exact-text overlap between `css_snippets.applies_to` (generic terms) and real component functions — no exact overlap, no match, so most visual snippets never ship. Two proposed fixes: manually update every `applies_to`, or make matching lemma/slug-aware.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_visual_pipeline_css_and_component_lists.md#Cause-2, js_snippets_news_gaswholesalers/old/css_snippets_matching_known_issue.md
- **relations:** CSS component-list fallback bug
- **verify-later:** loadComponentCSSSnippets in render_css_from_spec_action.go

<!-- SOURCE: U18_sql_for_agents.md -->
### Rerender pipeline (rerender-pages, page-rerender, render-site-components, rerender-site)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** 033–036 sequence; idle timeouts for rerender-pages/page-rerender in 075; needs_rerender items created by many fixers (056, 064) with handler rerender-pages.
- **what:** The assembly/deployment half of the system, separated from content generation: page_components store rendered sections; render_site_components renders header/footer/head into site_components; page-rerender re-assembles a single page from stored sections (with skip detection) and deploys; rerender-site orchestrates site-wide re-render (components → per-page loop → Cloudflare deploy). Design principle stated in 036: the loop sub_workflow is minimal, all per-page logic lives in the page-rerender agent. needs_rerender work items (priority 99, run last) are the standard "make fixes visible" side-effect.
- **sources:** 033_rerender_pages_action.sql; 034_page_rerender_agent.sql; 035_render_site_components.sql; 036_rerender_site_agent.sql; 064_site_component_linker_and_fixer.sql
- **relations:** nav-updater (adds nav refresh first); every fixer agent that returns needs_rerender
- **verify-later:** rerender_single_page / render_site_components actions; needs_rerender dedup guard (NOT EXISTS insert in 064)

<!-- SOURCE: U18_sql_for_agents.md -->
### site-asset-renderer (deterministic /assets/js/snippets.js)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** 113 INSERT with verification queries; description "Deterministic — no LLM. Triggered when js_snippets or component set changes".
- **what:** Renders a site's shared JS snippet bundle (e.g. relative-time expansion for news feeds) from the js_snippets table and commits it to git; components load it via a single `<script src="/assets/js/snippets.js">` injected into templates. Establishes the site-level shared-asset mechanism distinct from per-tool inline JS.
- **sources:** 113_site_asset_renderer.sql
- **relations:** js_snippets table; latest-news component; contrasts with the never-built per-tool asset extraction (143)
- **verify-later:** render path and trigger wiring for snippets.js

<!-- SOURCE: U19_sql_tables_components.md -->
### CSS responsibility barrier and CSS variable contract
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "CSS Responsibility Barrier Implementation — Global CSS handles all appearance... Components should NOT re-declare colors" plus the component CSS-variables migration (var(--variable-name, fallback)) applied across all seeded components; hardcoded-colour discovery audit (063b).
- **what:** Global styles.css (from webdesign-agent) owns colours/fonts; component CSS owns only layout/spacing, consuming CSS custom properties with fallbacks (var(--color-primary, #...)). Components must not re-declare colours global CSS styles, with an explicit exception protocol for dark/inverted sections. Audit queries exist to find violators.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#global-vs-local-css and #component-css-variables and #063b_hardcoded_colors_discovery
- **relations:** section-contrast model; style collections; webdesign-agent.
- **verify-later:** styles.css generation; remaining hardcoded colours in component templates.

<!-- SOURCE: U19_sql_tables_components.md -->
### Section-contrast model (is_dark_section + --section-* variables)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Live COMMENT: "is_dark_section ... MUST set --section-text, --section-text-muted, --section-heading, --section-surface, --section-border on container"; 014 section-context variable migration in 005.
- **what:** Components with dark backgrounds are flagged is_dark_section=true and must define the --section-* variable set on their container so text/heading/surface colours invert correctly regardless of the global palette. Migration audited false positives (components using #1a1a2e as text colour, not background) and back-filled the variables per naming contract.
- **sources:** docs/agent_docs/sql_for_tables/005b_bk_content_components.sql#is_dark_section-comment; docs/agent_docs/sql_for_tables/005_content_components.sql#014-section-context
- **relations:** CSS responsibility barrier; component naming contract.
- **verify-later:** is_dark_section rows vs presence of --section-* in their templates.

<!-- SOURCE: U19_sql_tables_components.md -->
### css_snippets / js_snippets with missing JS loader
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** 044: js_snippets row news-date-formatter inserted but "THIS ROW IS NOT CURRENTLY LOADED ANYWHERE — the head component template has no snippet-loading mechanism... A small half-day piece of work to mirror loadComponentCSSSnippets" (TODO).
- **what:** Per-component CSS lives in css_snippets (canonical; picked up when webdesign-agent runs) and is loaded via loadComponentCSSSnippets. A parallel js_snippets table exists but no loader; shared JS (e.g. formatNewsDate) is therefore duplicated inline in component IIFEs and page_components.rendered_html as a documented temporary violation of contract 003.
- **sources:** docs/agent_docs/sql_for_tables/044_css_snippets.sql
- **relations:** inline-JS extraction contract; news feed rendering; contracts doc 003.
- **verify-later:** js_snippets loader in RenderHead; duplication of formatNewsDate inline.

<!-- SOURCE: U19_sql_tables_components.md -->
### Inline JS extraction contract (js_content / separateInlineJS)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "add js_content column for interactive components — Add the column for future use" (005 ~9779); 044 notes the news component's inline <script> "violates contract 003. Properly extracting it via separateInlineJS() would make js_content the source of truth, with /tools/assets/latest-news.js as the served file."
- **what:** Interactive components should store scripts in content_components.js_content and serve them as external files under /tools/assets/, not as inline <script> in html_template. Column added; extraction not consistently done.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#js_content; docs/agent_docs/sql_for_tables/044_css_snippets.sql#why-temporary
- **relations:** css/js snippets; standalone tools (which embed script by design).
- **verify-later:** separateInlineJS usage; js_content population.

<!-- SOURCE: U20_legacy_docs_a.md -->
### aggregate_webpage HTML assembly action
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** Used in robot-hands-complete-website workflows (html_head/html_foot wrapper + section_order + response_fields, add_section_tags, page_name); replaced within docs004 by assemble_full_page/html-assembler and later by the current render pipeline.
- **what:** First-generation page renderer: wraps LLM-generated section content in a hard-coded HTML head (embedded CSS, nav) and footer, stitching named step outputs in a declared order into a complete page file. One action call per page.
- **sources:** docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md
- **relations:** successor: assemble_full_page + html-assembler agent, then the current CSS/render pipeline (styling-render-pipeline docs 036).
- **verify-later:** does aggregate_webpage still exist in the action registry.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Content/structure separation: JSON content + html-assembler (assemble_full_page)
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** 021/022: content-creator refactored to "structured JSON, not full HTML"; html-assembler agent with assemble_full_page action (template render → theme query → snippet queries → document assembly); the current render pipeline is the taxonomy successor.
- **what:** Separation of concerns that defines the modern pipeline: architect emits an empty {{placeholder}} template + content_requirements; content-creator emits pure content JSON (meta, theme recommendation, per-component sections); html-assembler merges template+content via Go templates then injects the CSS theme, tag-matched CSS snippets, and JS snippets into a complete document. Deployer receives finished HTML.
- **sources:** docs004_website_capture_project/006semantic_themes/README.022.description.md; docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md; docs004_website_capture_project/007different_types_of_site/031_about_page_multipage_site.md
- **relations:** successor: styling-render-pipeline (docs 036) + component render contracts (docs 003); content_components input_schema.
- **verify-later:** assemble_full_page in registry; html-assembler agent row.

<!-- SOURCE: U21_legacy_docs_b.md -->
### HTML action architecture (generate → process → validate)
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** docs006/008: "ALWAYS use the HTML actions instead of raw LLM calls... The architecture is already there — use it!"; replaced wholesale by component-template rendering in docs012+.
- **what:** A three-action pipeline for LLM page generation: `generate_html` (auto-gathers context from analyze_domain/architect_site/create_content/input_data, builds optimized prompt, extracts clean HTML), `process_html` (goquery parsing, meta tags, OG tags, responsive checks, lazy loading, minification), `validate_html` (structure, required elements, image alts, links, accessibility). Plus `assemble_html_parts` for chunking one huge page into structure/styles/content generations.
- **sources:** docs006_workflow_builder/008_20_plus_pages.md#The-HTML-Actions; docs006_workflow_builder/009_massive_multipage_sites.md#The-Actions-Available
- **relations:** superseded by content_components template rendering + render_mode matrix; chunked generation.
- **verify-later:** html_actions.go survival/usage in current action registry.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Batched multipage generation (assemble_multipage_site)
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** docs006/009: "for 20+ pages you need assemble_multipage_site... 5 batches × 4 pages = 80k tokens = WORKS"; docs010/019 then replaces batching: "Current (broken): spawn_multiple_writers ❌ Spawns 4 at once → New: loop".
- **what:** Handling 6–200+ page sites within LLM output limits by generating pages in batches of 3–5 per call, generating shared CSS once, injecting navigation with active states, and streaming files to S3 to avoid memory/Kafka-size limits (auto_store threshold pattern). Superseded by sequential per-page generation with the loop action after race conditions and quality problems.
- **sources:** docs006_workflow_builder/009_massive_multipage_sites.md#Quick-Decision-Tree; docs010_multitrack_flows_persona_architecture/019_start_here_document.md#Week-1
- **relations:** loop action; Kafka message size limits; stream_to_s3/auto_store (ancestor of storage-architecture S3 result offloading).
- **verify-later:** assemble_multipage_site action current form; auto_store config in agent chassis.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Asset bubble-up deduplication
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** docs009/001 "Return Value Bubble-Up... use 100 buttons, button.css included once"; production instead uses a single global styles.css plus inline component <style> blocks.
- **what:** During recursive rendering, each component returns its HTML plus its CSS/JS dependency list; parents merge children's assets upward, and the root injects the deduplicated set once into the head. Tied to js_dependencies column proposals on content_components.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#4-Solving-Assets
- **relations:** recursive component tree; CSS responsibility barrier (what actually shipped); JavaScript management section in docs017/023.
- **verify-later:** js_dependencies column existence.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Assembly action consolidation (3 clear actions)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** docs010/020: "You have [6 overlapping assembly actions]... Too much overlap. Proposed: 3 clear actions (assemble_page ...)"; later flows use assemble_page (docs011/003, docs015).
- **what:** Rationalizing the accumulated assembly actions (assemble_from_library, assemble_full_page, AssembleHTMLParts, AssembleMultipageSite, WrapMultipage, html_actions) into a minimal set: assemble_page (one page from structure+styles+content), plus multipage assembly and library assembly with shared code. A recurring theme: action proliferation followed by consolidation.
- **sources:** docs010_multitrack_flows_persona_architecture/020_revised_consolidated_action_plan.md; docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md
- **relations:** component library unification; slot-based assembly proposal.
- **verify-later:** current action registry entries for assembly actions.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Component library unification (component_library.go)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs012/009: "one source of truth for all component operations... RenderTemplate handles both Go-style {{.field}} and Handlebars-style {{field}}, {{#each}}, {{#if}}"; later docs (015, 017, 018) treat component_library.go functions as load-bearing infrastructure.
- **what:** A shared Go module consolidating duplicated component code: component queries (by function, by ID, with fallback), style collection resolution (per-site with domain-keyword fallback), theme loading, dual-syntax template rendering, and high-level RenderHeader/RenderFooter/InjectHeader/InjectFooter/InjectHead used by both full-page assembly (assemble_from_library) and header/footer injection into LLM-generated pages (multipage path).
- **sources:** docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md#Summary; docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md
- **relations:** style collections; InjectHead bug; GetNavItems; rerender pipeline.
- **verify-later:** platform/orchestration/actions/component_library.go current contents.

<!-- SOURCE: U21_legacy_docs_b.md -->
### page_components — component instances as the page's stored form
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Schema introduced docs012/006 ("the bridge between content_components (templates) and actual page content"); docs018/010 treats it as established core ("Each section on a page maps to a page_components row").
- **what:** Every section of every page is a row: template reference (component_id), position/slot_name, nesting支持 (parent_component_instance_id), the rendered_html actually deployed, the content_data that produced it, content_hash for change detection, and semantic addressing fields (data_path, data_uuid). This is the storage foundation that makes rerendering, section editing, locking, and maintenance possible — the single most consequential schema decision of this era.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.4; docs018_rerendering/010_section_editor_architecture.md#Component-Architecture; docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md#The-Data-Flow
- **relations:** content_data source-of-truth principle; component naming contract (slot_name = function); rerender; locks (asset locking mirrors page_components).
- **verify-later:** page_components current columns incl. schema_snapshot/content_snapshot/build_status.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Head-inside-body bug and positional injection fixes
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs015/002 trace: "cleanHTMLStructure keeps the LARGER <head> — wrong heuristic... InjectHead does in-place replacement — preserves wrong position" with concrete fixes and deployment order ending "re-run rerender_pages".
- **what:** Two compounding rendering bugs: LLM sections sometimes emit full HTML documents, and the dedup heuristic kept the larger (misplaced) head while in-place head replacement preserved the wrong position. Fixes: remove all head blocks then always insert before <body>; dedup by position (remove heads after <body>) not size. Exemplifies the fragility of regex injection that motivated slot-based assembly.
- **sources:** docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md#Bug-1
- **relations:** slot-based assembly proposal; component_library.go InjectHead; rerender pipeline.
- **verify-later:** current InjectHead/cleanHTMLStructure implementations.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Colour inheritance model + CSS responsibility barrier
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs015/002 Bug 2 ("body sets color — the ONLY default text colour; h1-h6 color: inherit; dark sections set color:#fff on container"); docs018/003: "Global CSS: all colors/fonts; Component CSS (inline): layout, positioning, structure only."
- **what:** The design-system rule set that fixed light-text-on-light-background failures: exactly one place sets default text colour (body); headings and text elements inherit; components never force colours or backgrounds on text elements; dark sections override at container level so children inherit white. Paired with the responsibility barrier: global styles.css owns colour/typography, component inline CSS owns layout only. Enforced through the webdesign-agent CSS prompt.
- **sources:** docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md#Bug-2; docs018_rerendering/003_website_builder_architecture_status_report.md#1; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#The-Colour-Inheritance-Model
- **relations:** dark-section --section-* contract (refinement); section-contrast model (current descendant); webdesign-agent.
- **verify-later:** current webdesign/CSS prompts; styles.css conventions.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Rerender pipeline (reassemble without regenerating)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs018/001 rerender_site_pages action doc; docs018/006 dual paths ("rerender-site: ensure_site_record → render_site_components [force] → loop(call page-rerender) → trigger_deploy"); rerender is a pillar of the current improvement loop.
- **what:** Re-assemble deployed pages from stored page_components.rendered_html with current site-level components (head/header/footer, CSS links, nav) without touching content: strip old wrappers, apply current chrome, commit. Split into page-rerender (single page) and rerender-pages orchestrator (batch), used after component/theme/nav changes. Includes contact-info injection from DB during rerender to overwrite hallucinated details.
- **sources:** docs018_rerendering/001_rerender_pages_summary.md; docs018_rerendering/006_build_path_rerender_path.md; docs018_rerendering/003_website_builder_architecture_status_report.md#2
- **relations:** page_components storage; improvement-loop rerender stage; section editor assemblePage reuse.
- **verify-later:** rerender_single_page_action.go, rerender-pages agent, trigger_deploy.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Slot-based modular page assembly (proposal)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** docs018/007 "Status: Draft for discussion, Created 2026-02-06"; user's inline answers ("agents should spawn their own dependencies", "no, we don't need migration"); site_components + render_site_components subsequently appear in the build path (docs018/006) but page_sections-as-JSON did not fully replace rendered_html storage.
- **what:** Replace regex header/footer injection with pure concatenation of slots (doctype/head/header/sections/footer); pre-render site-level components once into a site_components table; store section content as schema-validated JSON (page_sections) and render only at assembly so template changes never require content regeneration; explicit invalidation rules per change type; seven single-responsibility agents (site-planner, site-component-renderer, section-content-writer, link-manager, page-assembler, meta-manager, site-finalizer). Partially adopted: site_components and render_site_components shipped; JSON-first storage arrived instead as page_components.content_data source-of-truth.
- **sources:** docs018_rerendering/007_proposed_modular_rerendering.md; docs018_rerendering/006_build_path_rerender_path.md
- **relations:** recursive component tree (same instinct, earlier); InjectHead bug (motivation); section editor content_data principle.
- **verify-later:** site_components table + render_site_components action; page_sections existence.

<!-- SOURCE: U22_recent_small_docs.md -->
### CSS generation bug (webdesign-agent design_spec not applied)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "the webdesign-agent reported css_deployed: success:true ... But the deployed styles.css in git still contains the default blue template — the design_spec colors were never applied" (unsolved, 2026-03-02).
- **what:** A documented production defect: the webdesign-agent generates a correct `design_spec` (industry colours/fonts) but the generated/deployed CSS reverts to the default blue template. Three suspected causes: design_spec not reaching the template in structured form, an over-long prompt reproducing literal template CSS, or `content_field` resolution (`generated_css.result`) losing the CSS in the response envelope. Flagged for stage-2 debugging.
- **sources:** docs021.../024_handoff_summary_2026_03_02.md#the-css-bug
- **relations:** webdesign-agent, git_commit content_field resolution, unified site spec design_intent
- **verify-later:** webdesign-agent generate_css/deploy_css steps; extractFilesForGit content_field handling

<!-- SOURCE: U23_docs_root_vonc.md -->
### separateInlineJS inline-script extraction (+ collectJSAssets reader)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-29 ~19:30: "CODE IS CORRECT (not the bug)"; 2026-07-07: extraction pattern confirmed live on gauntlet-interface/latest-news/archetype-quiz (js_content + `<script src=` refs, no raw inline).
- **what:** On component store, `separateInlineJS` extracts bare `<script>` blocks (regex requires a closing tag; deliberately skips attributed tags — `src`, `type="application/ld+json"`, `type="module"` must stay inline) into `content_components.js_content`, replacing them with a `<script src="/tools/assets/{function}.js">` ref; multiple blocks are lazily matched and joined. `collectJSAssets` at rerender emits the per-component JS files. Known soft gaps: silent empty return on an unterminated `<script>` (warning proposed) and no log when an attributed script is left inline.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~19:30 + #2026-06-29-~20:00; docs/RUNBOOK_phase2_provocation_js(29).md#extraction-bug
- **relations:** two JS delivery paths; store-path validation hardening; legacy un-extracted components
- **verify-later:** store_generated_component_action.go separateInlineJS (~line 105); rerender_single_page_action.go collectJSAssets

<!-- SOURCE: U23_docs_root_vonc.md -->
### Visible-content filter + data-runtime-fill assembler exemption
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_section_assembly_drop(3) RESULT 2026-07-03: "FIX VERIFIED... provocation-card RESTORED to the live page; lobby-grid correctly ABSENT"; carried-forward state: "DEPLOYED + verified".
- **what:** `rerender_single_page`'s getPageSections drops any section whose rendered_html has ≤10 chars of visible text after stripping style/script/tags/entities — correct for genuinely empty shells, wrong for intentionally-empty runtime-filled ones. PATCH_section_visible_content adds one regexp + early return: a section carrying `data-runtime-fill` is kept regardless of build-time text; unmarked sections filter exactly as before (so unbuilt shells like the then-empty lobby-grid stay correctly dropped). The investigation is also a model correction arc: the raw-inline-script hypothesis was proven WRONG by reading the action (script is stripped before measuring).
- **sources:** docs/RUNBOOK_section_assembly_drop(3).md#d4-result + #result-fix-verified; docs/PATCH_section_visible_content(1).go.txt; docs/RUNNING_NOTES_vonc(36).md#2026-07-03-~14:15
- **relations:** runtime-fill mechanism; assemble-only rerender; marker REPLACE anchoring
- **verify-later:** rerender_single_page_action.go sectionHasVisibleContent + reRuntimeFill

<!-- SOURCE: U23_docs_root_vonc.md -->
### Two page-assembly paths with different chrome sources (stale site_components)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** 016b merged entry (from the parallel scheme-to-components thread): mechanism confirmed with provenance greps; paired-variable fix + workstreams referenced in SPEC_scheme_to_components (fix not claimed deployed in this corpus).
- **what:** Build path (page-build-handler → CompilePageSections → InjectHeader/Footer → RenderHeader/Footer) renders chrome FRESH — via style_collections.header_component_id (a dead, never-written column) falling through to GetComponentByFunction('site-header') or a dark RenderFallbackHeader. Rerender path reassembles stored page_components.rendered_html and injects STORED site_components.rendered_html, which can carry long-deactivated dark chrome — nothing refreshes site_components on deactivation, and stored section renders "fossilise" old templates (legacy `--accent-color` vars are the tell; needs_rerender re-fossilises; only a full rebuild re-renders templates). Provenance greps distinguish the three header origins; InjectHeader skips when a site-header class already exists.
- **sources:** docs/016b_debugging_guide_merged(3).md#light-site-renders-dark-chrome
- **relations:** two re-render paths; CSS variable naming (legacy-tell); post-025 theme flow
- **verify-later:** site_components refresh logic; style_collections.header_component_id writes; RUNBOOK/SPEC_scheme_to_components outcome

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### http2 deprecation fix at the generator
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "nginx 1.28.3 warns on `listen ... http2` (deprecated since 1.25) … the generator now emits version-neutral `listen 443 ssl;`".
- **what:** A field finding: newer nginx deprecates `listen ... http2` while the local container lacks the replacement `http2 on;`, so setup.sh's conf generator emits version-neutral `listen 443 ssl;` (with an opt-in comment for ≥1.25.1). Caught by fixing at the generator rather than per-box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-12-cert-issued
- **relations:** part of setup.sh nginx conf generation
- **verify-later:** setup.sh nginx server block listen directive

## Additional notes
The four `.go` entrypoint files are a version family: `main.go` ≡ `main(15).go` ≡ `main(17).go` (no NginxLogDir); `main(19).go` adds `NginxLogDir` config — the change that enabled the /access-digest endpoint. `env.go` header states its tiny env/envInt helpers are "copied verbatim from idea.uk's engine.go" — the trace of the idea.uk fork lineage. The two archived SQL files are byte-identical to each other and to live `intent_collector_agents(2).sql` (only cross-version change was `ON CONFLICT (type)` → `ON CONFLICT (type, version)`, i.e. debug-guide #28). This unit overlaps U11 (traffic_probe, the live tree) — consolidation should de-duplicate against U11.

<!-- SOURCE: U25_leopardess_social.md -->
### Core vs specialised palette slot merge semantics (buildPaletteMap leak)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES turn 10: "buildPaletteMap lets design_spec win the 8 core slots but the theme palette win every specialised slot" — root cause "fully code+data verified".
- **what:** The CSS renderer merges LLM design_spec over the theme palette for core slots only; specialised slots always come from the theme palette. A site that overrides core to dark/gold while sharing a seed palette serves mixed output (white cards, navy header, blue gradient CTA on a black-and-gold page). With no design_spec present the palette fully determines output (deterministic re-render).
- **sources:** docs/leopardessconsulting/RUNNING_NOTES.md#Turn-3, #Turn-10; docs/leopardessconsulting/scripts/L3_fork_palette.sql (header)
- **relations:** per-site style fork; analyze_design reference_values; specialised-slot contrast gap
- **verify-later:** render_css_from_spec / buildPaletteMap in the styling pipeline

<!-- SOURCE: U25_leopardess_social.md -->
### page-rerender mode contract and site-uniformity reconcile pattern
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES turn 14: "Assemble mode (page_id, NO spec.reason) deploys unconditionally … the holdouts flipped immediately. Result: 27/30 active pages carry the gold header."
- **what:** page-rerender has two modes: scoped (spec.reason='section_data_resolved' or 'image_landed' + spec.page_name) re-renders each stored section from html_template + stored content_data — NO content-hash skip exists [CORRECTED 2026-07-20, bugs_closed/031: the earlier hash-skip claim was never true of the code; the real bail-outs are page-level skipped (no stored components) / escalated (incomplete content_data), and section-level carried (missing component / unready plan / empty html_template) — see register/styling-render-pipeline.md STY-048]; assemble mode (page_id, no reason) re-embeds current header/footer unconditionally. rerender-site's sequential page loop stalls on a lost child response — drive pages individually. Throughput is a platform constraint: one chassis replica consumes page-rerenders serially (~45–60s each). reconcile_headers.sh/leo_reconcile_bg.sh implement the idempotent pattern: each round re-fire only pages whose deployed HTML still shows the old artifact.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#O8, #O9, #landmines-13/15; docs/leopardessconsulting/scripts/reassemble_pages.sh, reconcile_headers.sh, rerender_pages.sh (headers); docs/leopardessconsulting/RUNNING_NOTES.md#Turn-12–14
- **relations:** section-editor single-section path (vonc unit found the third route); assembler visible-content filter
- **verify-later:** check_rerender_mode in page-rerender workflow; rerender_single_page_action.go; chassis replica count

<!-- SOURCE: U25_leopardess_social.md -->
### Assembler visible-content filter with runtime-fill exemption
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES_provocation-card 2026-07-03 (end of day): "Assembler patch (data-runtime-fill exemption in sectionHasVisibleContent) deployed … runtime-filled shells are now first-class in the assembler."
- **what:** rerender_single_page's assembly drops any section with ≤10 chars of visible text after stripping style/script/tags — which silently removed the two Mode-B interactive sections from the deployed index while page_components said deployed (diagnosed via the H1–H4 hypothesis/decision-gate method in PLAN/RUNBOOK_section_assembly_drop). Fix: sections marked data-runtime-fill are exempt; genuinely empty unmarked shells (lobby-grid pre-loader) stay correctly dropped. assemblePage/sectionHasVisibleContent are shared by rerender and section-editor paths, so the exemption holds on both.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/PLAN_section_assembly_drop.md; docs/social001_vonc_tiktok_social/tool_docs/RUNBOOK_section_assembly_drop.md; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#2026-07-03
- **relations:** runtime-fill mechanism; page-rerender mode contract; verify-by-artifact discipline (diagnose before fix)
- **verify-later:** rerender_single_page_action.go:429-452 sectionHasVisibleContent

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Styling render pipeline reference: two assembly paths and the scheme gap
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** 036 FINDING/THEORY-tagged reconstruction from code + live data
- **what:** Stylesheet and page-section renders are separate code paths meeting only in the browser via class names/custom properties. Key FINDINGS: resolved_composition doesn't record scheme (survives only on layouts.scheme); buildSectionDefaults emits --section-* only for dark bg/surface (light sites correctly get nothing); five surface classes are duplicated renderer+layouts (Phase 4.5 debt); hero/CTA components hardcode dark backgrounds + literal white text defeating the scheme; .{function}-section class contract broken by hero (.hero) and CTA (.cta-section); four overlapping chrome default stores (style_collections ids [live read], site_components slots [likely superseded], sites.default_components, layouts.default_* [all NULL]); RenderFallbackHeader is hardcoded dark; SectionStyles/component_selector are dead on the current path. Fix direction (Q4a): strictly variable-driven components + renderer-owned per-section --section-*.
- **sources:** 036 full; 016b light-site-dark-chrome entry
- **relations:** section painting contract; site component linkage; scheme-to-components runbook
- **verify-later:** F-thread confirmations (update_site_defaults on composition path)

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### CSS assembly pipeline (composable theme → styles.css)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "fully built path" (2026-05-12); render_css_from_spec_action.go deterministic, verified live schema for css_snippets
- **what:** webdesign-agent: analyze_design (LLM) → render_css_from_spec (deterministic Go: theme composition from palettes/layouts/typography_sets FKs, css_snippets matched via applies_to JSONB overlap against the site's component functions, dark-section variants) → deploy_css git commit to assets/css/styles.css → B2 CDN sync. css-patch-agent is the bypass path for one-off fixes (patches the deployed file directly, not the snippet library).
- **sources:** FOCUS-css_js_mechanisms.md#1; HANDOFF_2026-04-18_design_and_styling…md
- **relations:** composable theme migration 025; site-design-planner
- **verify-later:** render_css_from_spec_action.go, render_css_composition_helpers.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### JS three-path model (js_content deployed, js_snippets loader missing, inline legacy)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "declared in contracts, table populated, BUT NO LOADER IS WIRED UP" (verified 2026-05-12: 9 js_snippets rows, no reference in head templates or RenderHead)
- **what:** Path A (deployed): per-component JS in content_components.js_content, extracted at store time by separateInlineJS(), deployed as /tools/assets/{function}.js via collectJSAssets() multi-file git commits. Path B (aspirational): js_snippets shared utility table with applies_to scoping — a registry of intentions with no runtime loader; contracts' "loaded via head component" claim is aspirational. Path C (legacy anti-pattern): inline <script> baked into html_template, violating contract 003 — news components still there. Path D: html-assembler's inject_js flag has no visible reader. Interim tactic: insert the snippet row AND duplicate inline until the loader exists.
- **sources:** FOCUS-css_js_mechanisms.md#2, #3, #4; HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#2
- **relations:** component contract 003; JS separation deployment (2026-04-17)
- **verify-later:** RenderHead in component_library.go; js_snippets rows; whether a loader was ever built

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Component quality tracking (quality_score et al.)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "migration_component_quality.sql (applied) … compute_component_quality_action.go (pending Go deploy)" (2026-04-17); quality scoring described as working in the 04-18 handoff status table
- **what:** content_components gains template_variable_count, schema_field_count, template_closed, schema_template_synced, has_data_component, quality_score (100 minus deductions), quality_issues; scored inline on store and by a component-quality-auditor agent; planner prefers high scores, auditor targets low ones for regeneration. Backfill via system.internal work item. 43 pre-existing components had 0 template variables (content baked in) — regeneration targets, not mass-deletable.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#6, #Architecture-decisions
- **relations:** pre-store validation gates; component-creator prompt tiers
- **verify-later:** compute_component_quality registry entry; quality_score population in DB

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Pre-store component validation gates + planning deferrals + empty-section filter
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** deployed 2026-04-17 (three checks before INSERT; sectionHasVisibleContent; empty-schema deferral); root incident: max_tokens=4000 truncation left unclosed <section>, CSS rendered as page text on vonc.com
- **what:** Three layers preventing broken components/sections reaching pages: store-time rejection (template must contain <section>/<div>, balanced <style> tags, non-empty input_schema), plan-time deferral of content-type components with empty schemas, and render-time skipping of sections with <10 chars visible text. Component-creator max_tokens raised 4000→16000 and prompt made context-aware.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#1, #6; HANDOFF_2026-04-17_nav_empty_sections_footer(1).md#1, #6, #7
- **relations:** LLM reliability tracks; quality tracking
- **verify-later:** store_generated_component_action.go validation block; rerender_single_page_action.go sectionHasVisibleContent

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Scheme-to-components P0: light-resolved site renders dark
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** PLAN "## CLOSED (2026-07-03) … closed on deployed evidence (RUNBOOK §SCHEME CLOSE: all nine grep checks pass; the stale-section fossil `var(--accent-color` is gone)"; RUNBOOK §SCHEME CLOSE lists the nine counts (site-header-section 32 / gradient 0 / footer 37 / --hero-ink 13 / color-mix 14 / cta pair consumed / white rgba 0 / fossil 0 / brief-explanation 0-expected).
- **what:** The defining P0 of this thread: the chassis resolves each site to a light or dark scheme and the scheme travels correctly through layout and palette variables, but the component library was written dark-first — components hardcode white text and dark `--section-*` context inline, so a light-resolved site (idea.uk, `tool-portal-light`) deployed dark chrome and dark sections. The winning mechanism was completion of the existing paired-variable standard rather than restructure: one layout patched, ten templates de-hardcoded, chrome repointed + force-rerendered, then a full page-build-handler rebuild.
- **sources:** PLAN_scheme_to_components(1).md#CLOSED; RUNBOOK_scheme_to_components(50).md#SCHEME-CLOSE; HANDOFF_scheme_to_components_for_claude_code(1).md#The-problem; running_notes_scheme_to_components(55).md#Tk
- **relations:** paired-variable standard; hero ink model; hazard/band declarer taxonomy; rebuild vs rerender semantics; chrome selection path.
- **verify-later:** deployed idea.uk B2 index.html greps; `content_components.html_template` for site-footer/call-to-action/hero; `layouts.css_template` for tool-portal-light.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Three-part styles.css assembly and palette merge rules
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sc), read from `render_css_from_spec_action.go` bodies in full, 2026-06-30; HANDOFF §Established mechanism restates it as verified.
- **what:** `RenderCSSFromSpecAction` builds styles.css in three appended parts: (1) the layout `css_template` rendered as a Go text/template with `{{palette}}`/`{{typo}}`/`{{token}}` FuncMap helpers over merged maps — merge rules: 8 core palette slots spec-wins, specialised slots theme-wins, typography spec-wins, structure layout-only; (2) component CSS from the `css_snippets` table (name, css_content, applies_to jsonb) where applies_to overlaps the site's components — a third CSS surface distinct from inline `<style>` (C3 cleared it of dark-section treatment: all 21 snippets are utilities); (3) the `buildSectionDefaults` luminance block. Theme composition loads via style_collections → css_themes joined to palettes/layouts/typography_sets, hard-erroring on NULL FKs.
- **sources:** running_notes_scheme_to_components(55).md#Sc #Se; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; PLAN_scheme_to_components(1).md#Confirmed-at-code-level
- **relations:** buildSectionDefaults; layout CTA pair curation; scheme derivation.
- **verify-later:** render_css_from_spec_action.go, render_css_composition_loader.go/_helpers.go; `css_snippets`, `palettes.colours`, `typography_sets` tables.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### buildSectionDefaults: luminance-keyed dark-only --section-* defaults
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sf) 2026-07-01: "`--section-*` is a DARK-ONLY override; light is the fallback. `buildSectionDefaults` returns '' unless bg or surface is dark."
- **what:** The renderer's only live per-section adaptation: `buildSectionDefaults` (color_util.go, WCAG `isDarkHex`/`pickReadableOnBackground`) emits a `body { --section-* }` block only when the merged palette's background or surface is dark, plus a dark-surface variant on 5 hardcoded surface classes (`.features/.services/.differentiators/.about/.faq-section`). On a light palette it emits nothing and element rules fall through to `var(--color-*)`. Retained unchanged under the paired-variable decision as the whole-palette-darkness base/safety net.
- **sources:** running_notes_scheme_to_components(55).md#Sf; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; SPEC_scheme_to_components.md#Decision-record
- **relations:** Colour Inheritance Model; Phase 4.5 deferral (generalises the 5-class list); paired-variable standard.
- **verify-later:** color_util.go buildSectionDefaults/isDarkHex; emitted styles.css tail on a dark-palette site.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### SectionStyles per-section CSS mechanism (built, disconnected, retired)
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** Notes (Sf): "`SectionStyles` is DEAD for current sites. None of the 18 active layouts reference `{{range .SectionStyles}}` … computed-but-unused"; SPEC: "`SectionStyles` stays retired."
- **what:** A fully-built but never-connected renderer mechanism: `queryDarkSectionsForCSS` + `buildCSSsectionStyles` compute per-component `{Function, ClassName: function+"-section", IsDark}` entries from `content_components.is_dark_section` (fallback dark list hero/social-proof/call-to-action/testimonials) and pass them to the layout template — which no active layout consumes. Considered as the cheap renderer-owns vehicle (Alt B) and explicitly retired by the paired-variable decision. A textbook infrastructure-orphan: ~80% built, deliberately not revived.
- **sources:** running_notes_scheme_to_components(55).md#Sf #Si; SPEC_scheme_to_components.md#Decision-record; HANDOFF_scheme_to_components_for_claude_code(1).md#Established
- **relations:** superseded by paired-variable standard; related Phase 4.5 (the other renderer-owns design).
- **verify-later:** render_css_from_spec_action.go buildCSSsectionStyles/queryDarkSectionsForCSS still present and uncalled from layouts.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Hero ink model and the structural-dark exception
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sv) "W3b COMPLETE (UPDATE 1; … ink in both inline branches, layered solid+single-hue color-mix gradient, five ink-referencing section vars…)"; W3d extended it to the five hero-* variants (UPDATE 5).
- **what:** Image/layered sections define a per-branch `--hero-ink` custom property and derive all text/context from it: the image branch sets `--hero-ink:#fff` under the structural-dark exception (an `rgba(0,0,0,x)` overlay guarantees darkness, so white text is always safe); the no-image branch sets `--hero-ink: var(--color-primary-text)` over a layered `var(--color-primary)` solid plus a single-hue gradient mixing 15% toward the ink (depth on both dark and light primaries; the solid layer doubles as the color-mix-less fallback). Buttons become the inverse pair (ink background, primary label). Chosen after data showed imageless heroes are the common case (80/114 hero, 26/26 hero-*), and it fixed a latent white-on-cyan failure on tool-portal-dark.
- **sources:** running_notes_scheme_to_components(55).md#St #Su #Sv #Sw; w3b_01_hero_conversion.sql; RUNBOOK_scheme_to_components(50).md#HERO-(c)-DESIGN
- **relations:** section painting contract model (c); paired-variable standard.
- **verify-later:** hero + hero-* `content_components.html_template` current bytes; rendered index hero.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Hazard-class vs band-class declarer taxonomy (library blast radius)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** CHECK 3 RESULTS (2026-07-02): 84 active sections — 15 hex backgrounds, 37 self-declare `--section-*`, split ~18 hazard vs ~19 band; SCHEME CLOSE remaining work item 4: "~10 remaining surface-painting declarers + ~17 band-class components (non-idea.uk)" still open.
- **what:** The diagnostic taxonomy that sized every fix decision: hazard-class components declare dark `--section-*` while painting surface variables or nothing — live white-on-light bugs today (the footer, site-head, the five hero-* variants, brief-explanation etc.); band-class components paint from primary/secondary/accent with white text — coherent today but blocking "fully light" (CTA, hero, social-proof, testimonials…). Ten templates (the idea.uk-visible set) were fixed by hand-needles; the non-idea.uk tail awaits the re-aimed fixer (Step D decision).
- **sources:** RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS #SCHEME-CLOSE; SPEC_scheme_to_components.md#W2 #W3; running_notes_scheme_to_components(55).md#Sn
- **relations:** fix_forced_text_colours re-aim (the tail vehicle); supervised fixer first-run.
- **verify-later:** re-run the 3c split query; count remaining literal declarers among active sections.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Chrome selection path and the dead header_component_id column
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sl): "`install_site_composition` sets `style_collections.header_component_id`/`footer_component_id` = NULL … grep finds NO code that writes them non-NULL … effectively a DEAD column"; HANDOFF §Established restates the chain with line numbers.
- **what:** Page-compile chrome resolution: `CompilePageSectionsAction` → `InjectHeader/InjectFooter/InjectHead` → `RenderHeader/RenderFooter` reads `style_collections.header_component_id` (always NULL — inserted NULL with a "webdesign-agent populates these later" comment and never written) → falls to `GetComponentByFunction("site-header")`, the single library-wide active component per function → else the hardcoded-dark fallback. `RenderHead` looks up function `head` (the only head component is inactive, so builds always used the fallback head); `site-head` is section-level and unreachable as chrome. Five other active header/footer functions (`*_pre_037`) are unreachable on this path. The one-active-component-per-function convention holds for sections by data (C4: no function has >1 active) though the UNIQUE index only covers tools.
- **sources:** running_notes_scheme_to_components(55).md#Sl #Sq #Se(C4); HANDOFF_scheme_to_components_for_claude_code(1).md#Established; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3d)
- **relations:** four chrome default stores; scheme-aware fallback chrome; dual chrome render paths.
- **verify-later:** component_library.go RenderHeader/RenderFooter/RenderHead/GetComponentByFunction; install_site_composition_action.go NULL insert + comment fate (W4c chose deleting the comment).

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Scheme-aware fallback chrome (RenderFallbackHeader/Footer consume the pairs)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Ud) 2026-07-06: "slice 2+F live"; RUNBOOK 07-06-night State: "Deployed: slices 1 …, 2 (fallback chrome C/D + Debug tidy E)".
- **what:** The safety-net chrome functions hardcoded dark (`background: ctx.PrimaryColor` default `#1a1a2e`, literal white text) — so any site whose chain broke got dark chrome regardless of scheme. Edits C/D replace the whole functions: backgrounds become `var(--color-header-bg, var(--color-surface))`/footer equivalent, text `var(--color-header-text, var(--color-text))`, muted/borders via `color-mix` — safe library-wide because Check 3e proved all 18 layouts set all four chrome vars. `RenderFallbackHead` deliberately unchanged (its only colour use is a `<meta theme-color>` value where `var()` cannot work). Edit E swapped the file's eight `logger.Debug` calls to Info per the no-Debug rule.
- **sources:** gobatch_02_component_library.md; running_notes_scheme_to_components(55).md#Sl #Tq #Ud; SPEC_scheme_to_components.md#W4(a)
- **relations:** chrome selection path; paired-variable standard; no-logger.Debug convention.
- **verify-later:** component_library.go RenderFallbackHeader/RenderFallbackFooter current bodies; deployed image tag containing them.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Dual chrome render paths and repoint-before-force_rerender ordering
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Sy): "render_site_components_action.go:345–430: pinned-component join has NO is_active filter; non-force runs SKIP non-empty slots → repoint BEFORE force_rerender or the old dark chrome re-renders"; W4b executed and verified 2026-07-02 (header 3750→6258B, footer color-mix in).
- **what:** Chrome has two render paths: the page-compile path renders fresh via RenderHeader/Footer, while `render_site_components` writes `site_components.rendered_html`, which the RERENDER handler injects into pages. The pinned join ignores `is_active`, and without `force_rerender` non-empty slots are skipped — so stale renders of deactivated components persist indefinitely. The W4b remedy pattern: repoint `site_components.component_id` to the active components (guarded on the known old ids; `rendered_html` deliberately left in place so there is no chrome-less window), then trigger rerender-pages v6 with `spec.refresh_site_components: true`. A deliberate side-effect became a staging technique: the chrome-refresh deploys the whole site as an intermediate visual checkpoint (light chrome over old sections) before the full rebuild.
- **sources:** running_notes_scheme_to_components(55).md#Sy #Sz #Ta #Td; w4b_01_repoint.sql; w4b_04_trigger_item.sql; RUNBOOK_scheme_to_components(50).md#W4b-RESULTS
- **relations:** rerender-pages v6 workflow; rebuild vs rerender semantics; four chrome default stores.
- **verify-later:** render_site_components_action.go force_rerender/skip logic; idea.uk site_components rows point at active components.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Phase 4.5 data-section-bg surface generalisation (deferred)
- **category:** styling-render-pipeline
- **status-signal:** aspirational
- **status-evidence:** SPEC consequences: "025 Phase 4.5 (`data-section-bg` surface generalisation) is deferred as a separate dark-site concern."
- **what:** Doc 025's already-designed decouple: components carry a `data-section-bg="surface"` attribute; the renderer replaces its hardcoded 5-surface-class list with an attribute selector; dual-write migration. The thread audited it seriously (Si prior-art pass) and then argued it down: it solves a dark-site generalisation idea.uk never hits, its blanket "never self-declare" conflates hazardous surface declarations with load-bearing band declarations, and renderer ownership reintroduces component intent one hop away. Remains the designed answer for dark sites with surface sections outside the hardcoded 5, if that ever bites.
- **sources:** running_notes_scheme_to_components(55).md#Si #Sk #Sm; SPEC_scheme_to_components.md#Decision-record; HANDOFF_scheme_to_components_for_claude_code(1).md#Questioning-025
- **relations:** buildSectionDefaults (the 5-class list it generalises); paired-variable standard (chosen instead).
- **verify-later:** docs 025 §427–505; any data-section-bg attribute usage in components.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Explicit RenderContext.Scheme signal (Q1)
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** Notes (Sf): "explicit `RenderContext.Scheme` is SECONDARY … This revises the Q1 emphasis in the PLAN"; never implemented anywhere in the executed fix.
- **what:** The original leading design (Q1/Q3): plumb the resolved scheme explicitly into both render entry points — `l.scheme` in the CSS loader SELECT + `themeComposition.Scheme`, and a `Scheme` field on `RenderContext` exposed via `contextToInterfaceMap` — so component templates receive a light/dark signal. Overtaken when Check 1 showed the scheme already reaches components implicitly through the palette `:root` values and luminance defaults; the components were the only thing defeating an already-working system. No scheme field was ever added.
- **sources:** PLAN_scheme_to_components(1).md#Q1; running_notes_scheme_to_components(55).md#Sb #Sf #Sk
- **relations:** superseded by paired-variable standard + implicit palette mechanism.
- **verify-later:** RenderContext struct (component_library.go) — confirm no Scheme field exists.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Exact-field-name template binding with silent empty on miss (RenderTemplate `<no value>` strip)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "Correction 2 — the silent-empty mechanism is in RenderTemplate" (NOTES §2, confirmed from uploaded component_library.go).
- **what:** `RenderTemplate` (component_library.go) binds a page's `content_data` into a component's `html_template` by exact field name via Go `text/template`, then strips the `<no value>` tokens of unmatched placeholders to empty string, logging only a warning — no error. This is *why* a renamed or missing field fails silently rather than loudly; the entire clobber class rests on it.
- **sources:** NOTES(43).md §2 Correction 2, §8; BUNDLE(3).md §1
- **relations:** clobber failure mode; F1 guard (compensating control); fail-loud guard route (never built as such).
- **verify-later:** platform/orchestration/actions/component_library.go:RenderTemplate; the `<no value>` cleanup.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Section readiness model (planSection source tiers, required/fallback semantics, spec resolver)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "planSection required semantics CONFIRMED from plan_sections_action.go… 'Required with no fallback — defer' → deterministic defer → carry" (NOTES §9w).
- **what:** Section fields declare a source (static with fallback; llm; Tier-C spec paths like `site_specs.cta.primary_url`) plus required/fallback attributes. planSection resolves each non-LLM field; a required field with no resolvable source and no fallback defers the section (→ carry). The resolver reads site_specs per-aspect rows (`aspect`,`data` jsonb, is_current; `resolveSpecPath("cta.primary_url") = specs["cta"]["primary_url"]`), checks presence not validity, and the stored⊕resolved merge persists resolved values into content_data at render time. Tier-C fields are by design never content_data keys.
- **sources:** NOTES(43).md §9s, §9u–§9x; RUNBOOK(49).md Part A
- **relations:** carry-forward path; F5; stored⊕resolved merge; phantom-CTA lesson (spec presence ≠ URL validity).
- **verify-later:** plan_sections_action.go (ensureSpecs, resolveSpecPath, on_missing switch); site_specs schema.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Stored⊕resolved merge writes resolved values back into content_data
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "resolver persists cta_url into content_data (stored⊕resolved merge — expected)" (NOTES §9w); robot-hands cd_keys gained merged fallback keys (§9an).
- **what:** When a section re-render resolves fields (spec values, static fallbacks), the merged result is persisted back into the page's `content_data` as well as baked into `rendered_html`. Double-edged: it makes recoveries durable (cta_url landed in content_data), but it is also F8 carrier 2 — contaminated fallback values were merged into dependents' content_data, surviving the later schema fix and needing an explicit key-strip.
- **sources:** NOTES(43).md §9w, §9an, §9ao; HANDOFF(7).md §Incident 2
- **relations:** F8 contamination; section readiness model; recovery playbook.
- **verify-later:** rerender_page_sections persist-merged-content step.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Visible-content filter drops near-empty bands at assembly
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "sections = page_components by page_id ORDER BY position; visible-text filter drops bands with ≤10 chars of stripped text (logger.Warn per skip)" (NOTES §9ai, from rerender_single_page_action.go).
- **what:** Page assembly filters out any band whose rendered HTML strips to ≤10 visible characters, logging only a warning. It is the final silencer in the clobber chain (an emptied section vanishes rather than erroring) and produces counter-intuitive interims — the F8 "neutral shell" bands survived the filter because two neutral CTA labels exceeded 10 chars.
- **sources:** NOTES(43).md §1, §9ai, §9ar; BUNDLE(3).md §1
- **relations:** clobber failure mode; assembly membership model; fail-loud guard route (unbuilt alternative).
- **verify-later:** rerender_single_page_action.go visible-text filter.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Assembly membership and chrome model (page_components by position; pages.sections is metadata)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "assembly membership = page_components by position… pages.sections jsonb is NOT assembly membership" (NOTES §9ai — corrected a wrong sections_listed=0 inference); three head shapes identified.
- **what:** A page artifact is assembled from `page_components` rows ordered by position (not from `pages.sections`, which is planning metadata); head/header/footer come from site-scoped `site_components` rows; with no stored head, `buildDefaultHead` emits a 5-line head linking /assets/css/styles.css. A third, legacy builder (`assemble_from_library`, theme CSS from css_themes) produced older artifacts with big inline heads — three coexisting head shapes that repeatedly confused artifact forensics. Also: `data-component` attributes exist on only some templates, so artifact greps on them undercount bands (owned metric artifact).
- **sources:** NOTES(43).md §9ah–§9al; RUNBOOK(49).md Part B
- **relations:** visible-content filter; R6f vocabulary drift (missing :root head); chrome refresh gating.
- **verify-later:** rerender_single_page_action.go; assemble_from_library (registry L493); site_components schema.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### R6f — theming vocabulary drift (defined vs consumed CSS custom properties)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "R6f confirmed as the 'renders badly' mechanism; fresh rebuilds render WORSE" (NOTES §9bh); mechanism narrowed to defined-vs-consumed drift §9am–§9bi.
- **what:** Component templates consume CSS custom properties (`var(--x, fallback)`) whose names drift from what the site's generated styles.css `:root` defines — 11 gap names in two patterns (synonyms like --border-radius vs --radius, and orphans like --hero-ink). Sections on undefined vocabulary render via per-component fallback values — a "fallback lottery" that goes dark-on-dark invisible on dark canvases (gripper-detail's blank page). Newer generators put :root in styles.css (rootless heads), older sites carry inline :root heads (why leopardess was immune). Every fresh rebuild worsened it as new templates minted new names.
- **sources:** NOTES(43).md §9al, §9am, §9bh, §9bi; RUNBOOK(49).md Part D; HANDOFF(7).md §R6f
- **relations:** D2a token aliases (fix); D2b prevention; deterministic styles.css rendering; assembly head model.
- **verify-later:** site styles.css :root contents vs template var() usage; robot-hands/vonc pre-fix stylesheets.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### D2a — buildTokenAliases renderer-enforced compatibility bridge
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "D2a VERIFIED in production 2026-07-06 — step 11 emits the alias block on a real gamesdesign pass" (NOTES §9bn); curl shows the trailing compatibility-aliases :root block.
- **what:** A step-11 post-pass in RenderCSSFromSpecAction (mirroring step 10's buildSectionDefaults "renderer-enforced" pattern): after rendering, append a trailing `:root` block defining ONLY the missing names from a package-level 11-entry alias table (synonyms → canonical var() references, orphans → palette-safe literals). Definition detection by `name+":"` so var() usages and sibling names don't count; idempotent; layout-defined values always win; one zap log field (token_alias_length). Sites self-heal on their next design pass; verified live via an adapted 076 webdesign trigger on gamesdesign.
- **sources:** NOTES(43).md §9bj–§9bn; RUNBOOK(49).md Part D D2a
- **relations:** R6f (the drift it bridges); D2b (prevention side); buildSectionDefaults (pattern precedent).
- **verify-later:** render_css_from_spec_action.go buildTokenAliases + tokenAliases table; render_css_from_spec_alias_test.go.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### D2b — canonical-token prevention (contract rule 11 + AuditTemplateTokens lint + prompt rule)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "D2b in progress: lint coded… contract rule 11 drafted; prompt edit pending agent-identify" (NOTES §9bq, 2026-07-06 — thread ends here).
- **what:** Stops new orphan tokens at the source, reuse-first: (1) contract rule 11 in 003's New Component Checklist — templates reference only canonical tokens + sanctioned aliases, invent no new var(--…) (drafted as a paste-in patch); (2) the generating agent's prompt enforces the rule in place (agent identification via default_config ILIKE still pending — 21 candidates); (3) `AuditTemplateTokens` warn-only lint appended to component_validation.go (canonicalCSSTokens allowlist = 39 theme names + 11 aliases, first-seen dedup, never rejects — detection net not gate, since vocabulary evolves), pending call-site wiring where ValidateComponentTemplate already runs. Notable design subtlety: rule 11 is the reciprocal of checklist item 6 (dark sections must SET --section-*) — consume-side vs set-side.
- **sources:** NOTES(43).md §9bo–§9bq; RUNBOOK(49).md Part D D2b
- **relations:** D2a (defines the sanctioned alias set); F8 lint (sibling detection-net); contracts-and-standards checklist.
- **verify-later:** component_validation.go AuditTemplateTokens; whether wired into StoreGeneratedComponentAction; 003 doc rule 11 presence.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### buildSectionDefaults — luminance-keyed dark-only section context
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "--section-* is a DARK-ONLY override; light is the fallback. buildSectionDefaults returns '' unless bg or surface is dark" (running_notes Sf, from color_util.go).
- **what:** The renderer-owned `--section-*` emitter: on a dark background/surface palette it emits a body-level (and 5 hardcoded surface-class) block of readable section text vars via WCAG helpers (isDarkHex, pickReadableOnBackground); a fully light site gets nothing and element rules fall back to `--color-*`. It is the live half of the section-context mechanism and the pattern precedent for D2a's step-11 alias block.
- **sources:** running_notes(22).md Sc, Sf; RUNBOOK_scheme_to_components(18).md D1 resolution
- **relations:** SectionStyles (dead sibling); paired-variable direction (keeps this as base); D2a.
- **verify-later:** color_util.go buildSectionDefaults/isDarkHex; the 5-class list vs 025's data-section-bg plan.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### SectionStyles — the dead per-section styling mechanism and the uneven {function}-section contract
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** "SectionStyles is DEAD for current sites — none of the 18 active layouts reference {{range .SectionStyles}}… computed-but-unused" (running_notes Sf); "SectionStyles stays retired" (Sn).
- **what:** A built-but-disconnected mechanism: queryDarkSectionsForCSS + buildCSSsectionStyles compute per-section entries (ClassName = function + "-section", IsDark from is_dark_section) for layout templates that no active layout consumes. The `{function}-section` class contract it assumes is real but honoured unevenly (hero emits `.hero`, CTA `.cta-section`). Decision: do not reconnect it — Phase 4.5's data-section-bg attribute keying and then the paired-variable direction supersede it.
- **sources:** running_notes(22).md Sc, Sf, Si, Sn; RUNBOOK_scheme_to_components(18).md D1
- **relations:** buildSectionDefaults (the live half); is_dark_section demotion; paired-variable direction (superseding).
- **verify-later:** render_css_from_spec_action.go buildCSSsectionStyles — still computed-and-unused?

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Hero ink model (per-branch --hero-ink with structural-dark exception)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "W3b COMPLETE… ink in both inline branches, layered solid+single-hue color-mix gradient" (running_notes Sv, run 2026-07-02 16:43); W3d extended it to the five variants.
- **what:** The hero's contrast contract after conversion: each branch sets an ink variable — image branch `--hero-ink:#fff` (the rgba overlay guarantees darkness: the sanctioned structural-dark exception); no-image branch `--hero-ink: var(--color-primary-text)` over a layered solid + single-hue color-mix gradient (15% toward the ink; solid layer is the color-mix-less fallback). Section vars derive from the ink at preserved alphas; buttons become the inverse pair. Fixed a latent white-on-cyan failure on the dark portal, not just the light-site problem. Imageless heroes turned out to be the COMMON case (80/114 + 26/26), reversing an assumption.
- **sources:** running_notes(22).md St–Sw; RUNBOOK_scheme_to_components(18).md HERO (c) DESIGN, W3b/W3d RESULTS
- **relations:** paired-variable direction; ambient pass-through (sibling pattern); D2a (--hero-ink later an alias orphan).
- **verify-later:** hero + hero-* html_template current state; whether rebuilds shipped the converted renders.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Ambient pass-through pattern for surface painters with fallback-less consumers
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "Sanctioned pattern recorded: page/surface painters with fallback-less consumers pass the ambient context through" (RUNBOOK_scheme(18) W3e, executed 2026-07-02 17:22).
- **what:** For components that paint the page/surface colour but whose internal rules consume `var(--section-*)` without fallbacks, the safe conversion is declaring `--section-x: var(--color-x)` pass-throughs rather than deleting the declarations (deletion would fall to currentColor/transparent). Scheme-correct on both light and dark by definition since the core vars ARE the scheme. Companion finding: `rgba(var(--hex-var), α)` is invalid CSS that never rendered — color-mix is the working replacement.
- **sources:** RUNBOOK_scheme_to_components(18).md W3e RESULTS; running_notes(22).md Sx
- **relations:** paired-variable direction; creator-prompt fallback mandate (future components shouldn't need this).
- **verify-later:** brief-explanation template pass-through block (NB: later regenerated in the F8 saga — check current state).

<!-- SOURCE: U08_travelling_docs.md -->
### Interactive-section clobber + interactivity-aware save guard (preserve-sections)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** 016b Part 4 "CAUSE CONFIRMED; FIX PENDING" → 2026-06-24 "fix WRITTEN (un-deployed)"; still listed as the pending multi-page prerequisite in RUNBOOK §5.4 at unit close.
- **what:** Any full rebuild regenerates a page from plan_sections, which knows nothing of an interactive tool stored only as a section's rendered_html — the game is silently discarded (detool-on-rebuild). Layered fix, both layers in `save_page_sections` (the only place holding the markup): (1) interactivity guard blocking a non-interactive set replacing a deployed interactive one; (2) carry-forward of existing interactive sections; plus source_item_id stamping. The unstated invariant "interactive sections survive every rebuild route" is the canonical example of what pipeline PLAN invariants should record.
- **sources:** 016b_debugging_guide_7_3_(7).md#open-threads-part-4; RUNBOOK_travelling_docs(38).md#§5.4; PLAN_travelling_docs(6).md#tool-assurance
- **relations:** multi-page prerequisites; pipeline documentation model; page build/rerender pipeline threads (Parts 1–5).
- **verify-later:** whether the patched save_page_sections_action.go deployed; page_component_history source_item_id.

<!-- SOURCE: U09_adoption.md -->
### Theme-layer render resolution (style_collection → css_theme, specs not read)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "Confirmed in the chassis actions: render context sets PrimaryColor/AccentColor/SecondaryColor from the resolved collection… Nothing in that path reads a site_specs design aspect" (FOCUS_design_composition, 2026-05-26).
- **what:** The live render path resolves colour/typography exclusively from `sites.style_collection_id → style_collections → css_themes`; site_specs design aspects influence it only upstream via the composition resolver. A NULL style_collection_id means no palette at all in render — the expected outcome for build-site-planner sites before the emit_design fix. Related earlier symptoms: dead BEM design-system CSS shipped in every head and no theme CSS variables defined (visual coherence coming from LLM-picked fallbacks).
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#1, old2/HANDOFF_2026-05-07(1)#8
- **relations:** emit_design_items; site-design-planner (doc 027, no-LLM thin wrapper over createPalette)
- **verify-later:** render context construction reads (~chassis 13548, 17464, 27934); theme variable injection status

<!-- SOURCE: U09_adoption.md -->
### Render-off-build_status debt (planned-vs-rendered diff)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "Render decision off build_status — the page-rerender agent appears to skip rebuilding planned-but-unrendered sections on a deployed page. Proper fix: drive render off a planned-vs-rendered diff… Until then the build_status='needs_rebuild' reset is the workaround" (HANDOFF_2026-06-09, parked).
- **what:** A `needs_page` rebuild of an already-deployed page completes without rebuilding planned-but-missing components; the render short-circuits on `build_status='deployed'` instead of diffing planned sections against current page_components. Workaround (reset to needs_rebuild) is proven; the structural diff-driven render is open debt.
- **sources:** HANDOFF_2026-06-06#resolved (14q), HANDOFF_2026-06-09#later-parked
- **relations:** positive-evidence principle (same "derived flag can lie" family); pages.sections vs page_components distinction
- **verify-later:** page-rerender agent def + assemble/deploy action short-circuit

<!-- SOURCE: U13_docs024_small_dirs.md -->
### js_snippets site-level rendering pipeline (site-asset-renderer)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** guidelines_compliance_check(1).md test plan table + "Migration A/B/C applied" checklist; walked against dev-guide/architecture/contracts docs as a deliberate review before shipping
- **what:** `render_js_snippets_for_site_action.go` + `site-asset-renderer` agent (4-step workflow) implements the previously-missing JS-snippet deploy step: `js_snippets` table (global, is_active flag) → matched against a site's component functions via `applies_to` → concatenated → `assets/js/snippets.js` per site, loaded via a `<script>` tag injected into the head template. Mirrors `render_css_from_spec`'s existing pattern exactly.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#The-JS-snippet-renderer-deliverable, js_snippets_news_gaswholesalers/old/guidelines_compliance_check(1).md
- **relations:** CSS component-list fallback bug; CSS applies_to granularity mismatch; component contract 003 (JS split)
- **verify-later:** js_snippets table (9 rows, 6 dormant is_active=false), site-asset-renderer agent_definition, head template script tag

<!-- SOURCE: U13_docs024_small_dirs.md -->
### CSS component-list fallback bug (fake 5-item list masking real component inventory)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "Cause 1 (fallback fires) was fixed 2026-05-16" — status filter fix applied and verified across two sites
- **what:** `extractCSSComponents` falls back to a hardcoded 5-item list whenever `site_context.all_component_functions` is empty. That field was empty because `loadPagesWithComponents`'s status filter matched nothing — every page's actual `status` value is `'active'`, never in that list. Fixed to `WHERE p.status = 'active'`.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_visual_pipeline_css_and_component_lists.md#Cause-1, js_snippets_news_gaswholesalers/old/design_actions_status_filter_fix.md
- **relations:** Assumed-status-values trap; CSS applies_to granularity mismatch
- **verify-later:** render_css_from_spec_action.go extractCSSComponents, design_actions.go loadPagesWithComponents

<!-- SOURCE: U13_docs024_small_dirs.md -->
### CSS applies_to granularity mismatch (known issue, unfixed)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "Cause 2 ... known issue, not yet fixed" — only 2 of ~21 snippets actually match real sites
- **what:** Even after the fallback-list bug is fixed, `loadComponentCSSSnippets` does exact-text overlap between `css_snippets.applies_to` (generic terms) and real component functions — no exact overlap, no match, so most visual snippets never ship. Two proposed fixes: manually update every `applies_to`, or make matching lemma/slug-aware.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_visual_pipeline_css_and_component_lists.md#Cause-2, js_snippets_news_gaswholesalers/old/css_snippets_matching_known_issue.md
- **relations:** CSS component-list fallback bug
- **verify-later:** loadComponentCSSSnippets in render_css_from_spec_action.go

<!-- SOURCE: U18_sql_for_agents.md -->
### Rerender pipeline (rerender-pages, page-rerender, render-site-components, rerender-site)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** 033–036 sequence; idle timeouts for rerender-pages/page-rerender in 075; needs_rerender items created by many fixers (056, 064) with handler rerender-pages.
- **what:** The assembly/deployment half of the system, separated from content generation: page_components store rendered sections; render_site_components renders header/footer/head into site_components; page-rerender re-assembles a single page from stored sections (with skip detection) and deploys; rerender-site orchestrates site-wide re-render (components → per-page loop → Cloudflare deploy). Design principle stated in 036: the loop sub_workflow is minimal, all per-page logic lives in the page-rerender agent. needs_rerender work items (priority 99, run last) are the standard "make fixes visible" side-effect.
- **sources:** 033_rerender_pages_action.sql; 034_page_rerender_agent.sql; 035_render_site_components.sql; 036_rerender_site_agent.sql; 064_site_component_linker_and_fixer.sql
- **relations:** nav-updater (adds nav refresh first); every fixer agent that returns needs_rerender
- **verify-later:** rerender_single_page / render_site_components actions; needs_rerender dedup guard (NOT EXISTS insert in 064)

<!-- SOURCE: U18_sql_for_agents.md -->
### site-asset-renderer (deterministic /assets/js/snippets.js)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** 113 INSERT with verification queries; description "Deterministic — no LLM. Triggered when js_snippets or component set changes".
- **what:** Renders a site's shared JS snippet bundle (e.g. relative-time expansion for news feeds) from the js_snippets table and commits it to git; components load it via a single `<script src="/assets/js/snippets.js">` injected into templates. Establishes the site-level shared-asset mechanism distinct from per-tool inline JS.
- **sources:** 113_site_asset_renderer.sql
- **relations:** js_snippets table; latest-news component; contrasts with the never-built per-tool asset extraction (143)
- **verify-later:** render path and trigger wiring for snippets.js

<!-- SOURCE: U19_sql_tables_components.md -->
### CSS responsibility barrier and CSS variable contract
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "CSS Responsibility Barrier Implementation — Global CSS handles all appearance... Components should NOT re-declare colors" plus the component CSS-variables migration (var(--variable-name, fallback)) applied across all seeded components; hardcoded-colour discovery audit (063b).
- **what:** Global styles.css (from webdesign-agent) owns colours/fonts; component CSS owns only layout/spacing, consuming CSS custom properties with fallbacks (var(--color-primary, #...)). Components must not re-declare colours global CSS styles, with an explicit exception protocol for dark/inverted sections. Audit queries exist to find violators.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#global-vs-local-css and #component-css-variables and #063b_hardcoded_colors_discovery
- **relations:** section-contrast model; style collections; webdesign-agent.
- **verify-later:** styles.css generation; remaining hardcoded colours in component templates.

<!-- SOURCE: U19_sql_tables_components.md -->
### Section-contrast model (is_dark_section + --section-* variables)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Live COMMENT: "is_dark_section ... MUST set --section-text, --section-text-muted, --section-heading, --section-surface, --section-border on container"; 014 section-context variable migration in 005.
- **what:** Components with dark backgrounds are flagged is_dark_section=true and must define the --section-* variable set on their container so text/heading/surface colours invert correctly regardless of the global palette. Migration audited false positives (components using #1a1a2e as text colour, not background) and back-filled the variables per naming contract.
- **sources:** docs/agent_docs/sql_for_tables/005b_bk_content_components.sql#is_dark_section-comment; docs/agent_docs/sql_for_tables/005_content_components.sql#014-section-context
- **relations:** CSS responsibility barrier; component naming contract.
- **verify-later:** is_dark_section rows vs presence of --section-* in their templates.

<!-- SOURCE: U19_sql_tables_components.md -->
### css_snippets / js_snippets with missing JS loader
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** 044: js_snippets row news-date-formatter inserted but "THIS ROW IS NOT CURRENTLY LOADED ANYWHERE — the head component template has no snippet-loading mechanism... A small half-day piece of work to mirror loadComponentCSSSnippets" (TODO).
- **what:** Per-component CSS lives in css_snippets (canonical; picked up when webdesign-agent runs) and is loaded via loadComponentCSSSnippets. A parallel js_snippets table exists but no loader; shared JS (e.g. formatNewsDate) is therefore duplicated inline in component IIFEs and page_components.rendered_html as a documented temporary violation of contract 003.
- **sources:** docs/agent_docs/sql_for_tables/044_css_snippets.sql
- **relations:** inline-JS extraction contract; news feed rendering; contracts doc 003.
- **verify-later:** js_snippets loader in RenderHead; duplication of formatNewsDate inline.

<!-- SOURCE: U19_sql_tables_components.md -->
### Inline JS extraction contract (js_content / separateInlineJS)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "add js_content column for interactive components — Add the column for future use" (005 ~9779); 044 notes the news component's inline <script> "violates contract 003. Properly extracting it via separateInlineJS() would make js_content the source of truth, with /tools/assets/latest-news.js as the served file."
- **what:** Interactive components should store scripts in content_components.js_content and serve them as external files under /tools/assets/, not as inline <script> in html_template. Column added; extraction not consistently done.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#js_content; docs/agent_docs/sql_for_tables/044_css_snippets.sql#why-temporary
- **relations:** css/js snippets; standalone tools (which embed script by design).
- **verify-later:** separateInlineJS usage; js_content population.

<!-- SOURCE: U20_legacy_docs_a.md -->
### aggregate_webpage HTML assembly action
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** Used in robot-hands-complete-website workflows (html_head/html_foot wrapper + section_order + response_fields, add_section_tags, page_name); replaced within docs004 by assemble_full_page/html-assembler and later by the current render pipeline.
- **what:** First-generation page renderer: wraps LLM-generated section content in a hard-coded HTML head (embedded CSS, nav) and footer, stitching named step outputs in a declared order into a complete page file. One action call per page.
- **sources:** docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md
- **relations:** successor: assemble_full_page + html-assembler agent, then the current CSS/render pipeline (styling-render-pipeline docs 036).
- **verify-later:** does aggregate_webpage still exist in the action registry.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Content/structure separation: JSON content + html-assembler (assemble_full_page)
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** 021/022: content-creator refactored to "structured JSON, not full HTML"; html-assembler agent with assemble_full_page action (template render → theme query → snippet queries → document assembly); the current render pipeline is the taxonomy successor.
- **what:** Separation of concerns that defines the modern pipeline: architect emits an empty {{placeholder}} template + content_requirements; content-creator emits pure content JSON (meta, theme recommendation, per-component sections); html-assembler merges template+content via Go templates then injects the CSS theme, tag-matched CSS snippets, and JS snippets into a complete document. Deployer receives finished HTML.
- **sources:** docs004_website_capture_project/006semantic_themes/README.022.description.md; docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md; docs004_website_capture_project/007different_types_of_site/031_about_page_multipage_site.md
- **relations:** successor: styling-render-pipeline (docs 036) + component render contracts (docs 003); content_components input_schema.
- **verify-later:** assemble_full_page in registry; html-assembler agent row.

<!-- SOURCE: U21_legacy_docs_b.md -->
### HTML action architecture (generate → process → validate)
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** docs006/008: "ALWAYS use the HTML actions instead of raw LLM calls... The architecture is already there — use it!"; replaced wholesale by component-template rendering in docs012+.
- **what:** A three-action pipeline for LLM page generation: `generate_html` (auto-gathers context from analyze_domain/architect_site/create_content/input_data, builds optimized prompt, extracts clean HTML), `process_html` (goquery parsing, meta tags, OG tags, responsive checks, lazy loading, minification), `validate_html` (structure, required elements, image alts, links, accessibility). Plus `assemble_html_parts` for chunking one huge page into structure/styles/content generations.
- **sources:** docs006_workflow_builder/008_20_plus_pages.md#The-HTML-Actions; docs006_workflow_builder/009_massive_multipage_sites.md#The-Actions-Available
- **relations:** superseded by content_components template rendering + render_mode matrix; chunked generation.
- **verify-later:** html_actions.go survival/usage in current action registry.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Batched multipage generation (assemble_multipage_site)
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** docs006/009: "for 20+ pages you need assemble_multipage_site... 5 batches × 4 pages = 80k tokens = WORKS"; docs010/019 then replaces batching: "Current (broken): spawn_multiple_writers ❌ Spawns 4 at once → New: loop".
- **what:** Handling 6–200+ page sites within LLM output limits by generating pages in batches of 3–5 per call, generating shared CSS once, injecting navigation with active states, and streaming files to S3 to avoid memory/Kafka-size limits (auto_store threshold pattern). Superseded by sequential per-page generation with the loop action after race conditions and quality problems.
- **sources:** docs006_workflow_builder/009_massive_multipage_sites.md#Quick-Decision-Tree; docs010_multitrack_flows_persona_architecture/019_start_here_document.md#Week-1
- **relations:** loop action; Kafka message size limits; stream_to_s3/auto_store (ancestor of storage-architecture S3 result offloading).
- **verify-later:** assemble_multipage_site action current form; auto_store config in agent chassis.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Asset bubble-up deduplication
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** docs009/001 "Return Value Bubble-Up... use 100 buttons, button.css included once"; production instead uses a single global styles.css plus inline component <style> blocks.
- **what:** During recursive rendering, each component returns its HTML plus its CSS/JS dependency list; parents merge children's assets upward, and the root injects the deduplicated set once into the head. Tied to js_dependencies column proposals on content_components.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#4-Solving-Assets
- **relations:** recursive component tree; CSS responsibility barrier (what actually shipped); JavaScript management section in docs017/023.
- **verify-later:** js_dependencies column existence.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Assembly action consolidation (3 clear actions)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** docs010/020: "You have [6 overlapping assembly actions]... Too much overlap. Proposed: 3 clear actions (assemble_page ...)"; later flows use assemble_page (docs011/003, docs015).
- **what:** Rationalizing the accumulated assembly actions (assemble_from_library, assemble_full_page, AssembleHTMLParts, AssembleMultipageSite, WrapMultipage, html_actions) into a minimal set: assemble_page (one page from structure+styles+content), plus multipage assembly and library assembly with shared code. A recurring theme: action proliferation followed by consolidation.
- **sources:** docs010_multitrack_flows_persona_architecture/020_revised_consolidated_action_plan.md; docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md
- **relations:** component library unification; slot-based assembly proposal.
- **verify-later:** current action registry entries for assembly actions.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Component library unification (component_library.go)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs012/009: "one source of truth for all component operations... RenderTemplate handles both Go-style {{.field}} and Handlebars-style {{field}}, {{#each}}, {{#if}}"; later docs (015, 017, 018) treat component_library.go functions as load-bearing infrastructure.
- **what:** A shared Go module consolidating duplicated component code: component queries (by function, by ID, with fallback), style collection resolution (per-site with domain-keyword fallback), theme loading, dual-syntax template rendering, and high-level RenderHeader/RenderFooter/InjectHeader/InjectFooter/InjectHead used by both full-page assembly (assemble_from_library) and header/footer injection into LLM-generated pages (multipage path).
- **sources:** docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md#Summary; docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md
- **relations:** style collections; InjectHead bug; GetNavItems; rerender pipeline.
- **verify-later:** platform/orchestration/actions/component_library.go current contents.

<!-- SOURCE: U21_legacy_docs_b.md -->
### page_components — component instances as the page's stored form
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Schema introduced docs012/006 ("the bridge between content_components (templates) and actual page content"); docs018/010 treats it as established core ("Each section on a page maps to a page_components row").
- **what:** Every section of every page is a row: template reference (component_id), position/slot_name, nesting支持 (parent_component_instance_id), the rendered_html actually deployed, the content_data that produced it, content_hash for change detection, and semantic addressing fields (data_path, data_uuid). This is the storage foundation that makes rerendering, section editing, locking, and maintenance possible — the single most consequential schema decision of this era.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.4; docs018_rerendering/010_section_editor_architecture.md#Component-Architecture; docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md#The-Data-Flow
- **relations:** content_data source-of-truth principle; component naming contract (slot_name = function); rerender; locks (asset locking mirrors page_components).
- **verify-later:** page_components current columns incl. schema_snapshot/content_snapshot/build_status.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Head-inside-body bug and positional injection fixes
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs015/002 trace: "cleanHTMLStructure keeps the LARGER <head> — wrong heuristic... InjectHead does in-place replacement — preserves wrong position" with concrete fixes and deployment order ending "re-run rerender_pages".
- **what:** Two compounding rendering bugs: LLM sections sometimes emit full HTML documents, and the dedup heuristic kept the larger (misplaced) head while in-place head replacement preserved the wrong position. Fixes: remove all head blocks then always insert before <body>; dedup by position (remove heads after <body>) not size. Exemplifies the fragility of regex injection that motivated slot-based assembly.
- **sources:** docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md#Bug-1
- **relations:** slot-based assembly proposal; component_library.go InjectHead; rerender pipeline.
- **verify-later:** current InjectHead/cleanHTMLStructure implementations.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Colour inheritance model + CSS responsibility barrier
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs015/002 Bug 2 ("body sets color — the ONLY default text colour; h1-h6 color: inherit; dark sections set color:#fff on container"); docs018/003: "Global CSS: all colors/fonts; Component CSS (inline): layout, positioning, structure only."
- **what:** The design-system rule set that fixed light-text-on-light-background failures: exactly one place sets default text colour (body); headings and text elements inherit; components never force colours or backgrounds on text elements; dark sections override at container level so children inherit white. Paired with the responsibility barrier: global styles.css owns colour/typography, component inline CSS owns layout only. Enforced through the webdesign-agent CSS prompt.
- **sources:** docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md#Bug-2; docs018_rerendering/003_website_builder_architecture_status_report.md#1; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#The-Colour-Inheritance-Model
- **relations:** dark-section --section-* contract (refinement); section-contrast model (current descendant); webdesign-agent.
- **verify-later:** current webdesign/CSS prompts; styles.css conventions.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Rerender pipeline (reassemble without regenerating)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs018/001 rerender_site_pages action doc; docs018/006 dual paths ("rerender-site: ensure_site_record → render_site_components [force] → loop(call page-rerender) → trigger_deploy"); rerender is a pillar of the current improvement loop.
- **what:** Re-assemble deployed pages from stored page_components.rendered_html with current site-level components (head/header/footer, CSS links, nav) without touching content: strip old wrappers, apply current chrome, commit. Split into page-rerender (single page) and rerender-pages orchestrator (batch), used after component/theme/nav changes. Includes contact-info injection from DB during rerender to overwrite hallucinated details.
- **sources:** docs018_rerendering/001_rerender_pages_summary.md; docs018_rerendering/006_build_path_rerender_path.md; docs018_rerendering/003_website_builder_architecture_status_report.md#2
- **relations:** page_components storage; improvement-loop rerender stage; section editor assemblePage reuse.
- **verify-later:** rerender_single_page_action.go, rerender-pages agent, trigger_deploy.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Slot-based modular page assembly (proposal)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** docs018/007 "Status: Draft for discussion, Created 2026-02-06"; user's inline answers ("agents should spawn their own dependencies", "no, we don't need migration"); site_components + render_site_components subsequently appear in the build path (docs018/006) but page_sections-as-JSON did not fully replace rendered_html storage.
- **what:** Replace regex header/footer injection with pure concatenation of slots (doctype/head/header/sections/footer); pre-render site-level components once into a site_components table; store section content as schema-validated JSON (page_sections) and render only at assembly so template changes never require content regeneration; explicit invalidation rules per change type; seven single-responsibility agents (site-planner, site-component-renderer, section-content-writer, link-manager, page-assembler, meta-manager, site-finalizer). Partially adopted: site_components and render_site_components shipped; JSON-first storage arrived instead as page_components.content_data source-of-truth.
- **sources:** docs018_rerendering/007_proposed_modular_rerendering.md; docs018_rerendering/006_build_path_rerender_path.md
- **relations:** recursive component tree (same instinct, earlier); InjectHead bug (motivation); section editor content_data principle.
- **verify-later:** site_components table + render_site_components action; page_sections existence.

<!-- SOURCE: U22_recent_small_docs.md -->
### CSS generation bug (webdesign-agent design_spec not applied)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "the webdesign-agent reported css_deployed: success:true ... But the deployed styles.css in git still contains the default blue template — the design_spec colors were never applied" (unsolved, 2026-03-02).
- **what:** A documented production defect: the webdesign-agent generates a correct `design_spec` (industry colours/fonts) but the generated/deployed CSS reverts to the default blue template. Three suspected causes: design_spec not reaching the template in structured form, an over-long prompt reproducing literal template CSS, or `content_field` resolution (`generated_css.result`) losing the CSS in the response envelope. Flagged for stage-2 debugging.
- **sources:** docs021.../024_handoff_summary_2026_03_02.md#the-css-bug
- **relations:** webdesign-agent, git_commit content_field resolution, unified site spec design_intent
- **verify-later:** webdesign-agent generate_css/deploy_css steps; extractFilesForGit content_field handling

<!-- SOURCE: U23_docs_root_vonc.md -->
### separateInlineJS inline-script extraction (+ collectJSAssets reader)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-29 ~19:30: "CODE IS CORRECT (not the bug)"; 2026-07-07: extraction pattern confirmed live on gauntlet-interface/latest-news/archetype-quiz (js_content + `<script src=` refs, no raw inline).
- **what:** On component store, `separateInlineJS` extracts bare `<script>` blocks (regex requires a closing tag; deliberately skips attributed tags — `src`, `type="application/ld+json"`, `type="module"` must stay inline) into `content_components.js_content`, replacing them with a `<script src="/tools/assets/{function}.js">` ref; multiple blocks are lazily matched and joined. `collectJSAssets` at rerender emits the per-component JS files. Known soft gaps: silent empty return on an unterminated `<script>` (warning proposed) and no log when an attributed script is left inline.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~19:30 + #2026-06-29-~20:00; docs/RUNBOOK_phase2_provocation_js(29).md#extraction-bug
- **relations:** two JS delivery paths; store-path validation hardening; legacy un-extracted components
- **verify-later:** store_generated_component_action.go separateInlineJS (~line 105); rerender_single_page_action.go collectJSAssets

<!-- SOURCE: U23_docs_root_vonc.md -->
### Visible-content filter + data-runtime-fill assembler exemption
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_section_assembly_drop(3) RESULT 2026-07-03: "FIX VERIFIED... provocation-card RESTORED to the live page; lobby-grid correctly ABSENT"; carried-forward state: "DEPLOYED + verified".
- **what:** `rerender_single_page`'s getPageSections drops any section whose rendered_html has ≤10 chars of visible text after stripping style/script/tags/entities — correct for genuinely empty shells, wrong for intentionally-empty runtime-filled ones. PATCH_section_visible_content adds one regexp + early return: a section carrying `data-runtime-fill` is kept regardless of build-time text; unmarked sections filter exactly as before (so unbuilt shells like the then-empty lobby-grid stay correctly dropped). The investigation is also a model correction arc: the raw-inline-script hypothesis was proven WRONG by reading the action (script is stripped before measuring).
- **sources:** docs/RUNBOOK_section_assembly_drop(3).md#d4-result + #result-fix-verified; docs/PATCH_section_visible_content(1).go.txt; docs/RUNNING_NOTES_vonc(36).md#2026-07-03-~14:15
- **relations:** runtime-fill mechanism; assemble-only rerender; marker REPLACE anchoring
- **verify-later:** rerender_single_page_action.go sectionHasVisibleContent + reRuntimeFill

<!-- SOURCE: U23_docs_root_vonc.md -->
### Two page-assembly paths with different chrome sources (stale site_components)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** 016b merged entry (from the parallel scheme-to-components thread): mechanism confirmed with provenance greps; paired-variable fix + workstreams referenced in SPEC_scheme_to_components (fix not claimed deployed in this corpus).
- **what:** Build path (page-build-handler → CompilePageSections → InjectHeader/Footer → RenderHeader/Footer) renders chrome FRESH — via style_collections.header_component_id (a dead, never-written column) falling through to GetComponentByFunction('site-header') or a dark RenderFallbackHeader. Rerender path reassembles stored page_components.rendered_html and injects STORED site_components.rendered_html, which can carry long-deactivated dark chrome — nothing refreshes site_components on deactivation, and stored section renders "fossilise" old templates (legacy `--accent-color` vars are the tell; needs_rerender re-fossilises; only a full rebuild re-renders templates). Provenance greps distinguish the three header origins; InjectHeader skips when a site-header class already exists.
- **sources:** docs/016b_debugging_guide_merged(3).md#light-site-renders-dark-chrome
- **relations:** two re-render paths; CSS variable naming (legacy-tell); post-025 theme flow
- **verify-later:** site_components refresh logic; style_collections.header_component_id writes; RUNBOOK/SPEC_scheme_to_components outcome

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### http2 deprecation fix at the generator
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "nginx 1.28.3 warns on `listen ... http2` (deprecated since 1.25) … the generator now emits version-neutral `listen 443 ssl;`".
- **what:** A field finding: newer nginx deprecates `listen ... http2` while the local container lacks the replacement `http2 on;`, so setup.sh's conf generator emits version-neutral `listen 443 ssl;` (with an opt-in comment for ≥1.25.1). Caught by fixing at the generator rather than per-box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-12-cert-issued
- **relations:** part of setup.sh nginx conf generation
- **verify-later:** setup.sh nginx server block listen directive

## Additional notes
The four `.go` entrypoint files are a version family: `main.go` ≡ `main(15).go` ≡ `main(17).go` (no NginxLogDir); `main(19).go` adds `NginxLogDir` config — the change that enabled the /access-digest endpoint. `env.go` header states its tiny env/envInt helpers are "copied verbatim from idea.uk's engine.go" — the trace of the idea.uk fork lineage. The two archived SQL files are byte-identical to each other and to live `intent_collector_agents(2).sql` (only cross-version change was `ON CONFLICT (type)` → `ON CONFLICT (type, version)`, i.e. debug-guide #28). This unit overlaps U11 (traffic_probe, the live tree) — consolidation should de-duplicate against U11.

<!-- SOURCE: U25_leopardess_social.md -->
### Core vs specialised palette slot merge semantics (buildPaletteMap leak)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES turn 10: "buildPaletteMap lets design_spec win the 8 core slots but the theme palette win every specialised slot" — root cause "fully code+data verified".
- **what:** The CSS renderer merges LLM design_spec over the theme palette for core slots only; specialised slots always come from the theme palette. A site that overrides core to dark/gold while sharing a seed palette serves mixed output (white cards, navy header, blue gradient CTA on a black-and-gold page). With no design_spec present the palette fully determines output (deterministic re-render).
- **sources:** docs/leopardessconsulting/RUNNING_NOTES.md#Turn-3, #Turn-10; docs/leopardessconsulting/scripts/L3_fork_palette.sql (header)
- **relations:** per-site style fork; analyze_design reference_values; specialised-slot contrast gap
- **verify-later:** render_css_from_spec / buildPaletteMap in the styling pipeline

<!-- SOURCE: U25_leopardess_social.md -->
### page-rerender mode contract and site-uniformity reconcile pattern
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES turn 14: "Assemble mode (page_id, NO spec.reason) deploys unconditionally … the holdouts flipped immediately. Result: 27/30 active pages carry the gold header."
- **what:** page-rerender has two modes: scoped (spec.reason='section_data_resolved' or 'image_landed' + spec.page_name) re-renders each stored section from html_template + stored content_data — NO content-hash skip exists [CORRECTED 2026-07-20, bugs_closed/031: the earlier hash-skip claim was never true of the code; the real bail-outs are page-level skipped (no stored components) / escalated (incomplete content_data), and section-level carried (missing component / unready plan / empty html_template) — see register/styling-render-pipeline.md STY-048]; assemble mode (page_id, no reason) re-embeds current header/footer unconditionally. rerender-site's sequential page loop stalls on a lost child response — drive pages individually. Throughput is a platform constraint: one chassis replica consumes page-rerenders serially (~45–60s each). reconcile_headers.sh/leo_reconcile_bg.sh implement the idempotent pattern: each round re-fire only pages whose deployed HTML still shows the old artifact.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#O8, #O9, #landmines-13/15; docs/leopardessconsulting/scripts/reassemble_pages.sh, reconcile_headers.sh, rerender_pages.sh (headers); docs/leopardessconsulting/RUNNING_NOTES.md#Turn-12–14
- **relations:** section-editor single-section path (vonc unit found the third route); assembler visible-content filter
- **verify-later:** check_rerender_mode in page-rerender workflow; rerender_single_page_action.go; chassis replica count

<!-- SOURCE: U25_leopardess_social.md -->
### Assembler visible-content filter with runtime-fill exemption
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES_provocation-card 2026-07-03 (end of day): "Assembler patch (data-runtime-fill exemption in sectionHasVisibleContent) deployed … runtime-filled shells are now first-class in the assembler."
- **what:** rerender_single_page's assembly drops any section with ≤10 chars of visible text after stripping style/script/tags — which silently removed the two Mode-B interactive sections from the deployed index while page_components said deployed (diagnosed via the H1–H4 hypothesis/decision-gate method in PLAN/RUNBOOK_section_assembly_drop). Fix: sections marked data-runtime-fill are exempt; genuinely empty unmarked shells (lobby-grid pre-loader) stay correctly dropped. assemblePage/sectionHasVisibleContent are shared by rerender and section-editor paths, so the exemption holds on both.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/PLAN_section_assembly_drop.md; docs/social001_vonc_tiktok_social/tool_docs/RUNBOOK_section_assembly_drop.md; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#2026-07-03
- **relations:** runtime-fill mechanism; page-rerender mode contract; verify-by-artifact discipline (diagnose before fix)
- **verify-later:** rerender_single_page_action.go:429-452 sectionHasVisibleContent
