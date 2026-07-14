# Cluster: styling-nav-links
Categories included: styling-render-pipeline, navigation, link-management


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
- **what:** page-rerender has two modes: with spec.reason='section_data_resolved' (or 'image_landed') + spec.page_name it regenerates section HTML from content_data, but SKIPS pages whose content hash is unchanged — silently wrong for header/footer changes; assemble mode (page_id, no reason) re-embeds current header/footer unconditionally. rerender-site's sequential page loop stalls on a lost child response — drive pages individually. Throughput is a platform constraint: one chassis replica consumes page-rerenders serially (~45–60s each). reconcile_headers.sh/leo_reconcile_bg.sh implement the idempotent pattern: each round re-fire only pages whose deployed HTML still shows the old artifact.
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
- **what:** page-rerender has two modes: with spec.reason='section_data_resolved' (or 'image_landed') + spec.page_name it regenerates section HTML from content_data, but SKIPS pages whose content hash is unchanged — silently wrong for header/footer changes; assemble mode (page_id, no reason) re-embeds current header/footer unconditionally. rerender-site's sequential page loop stalls on a lost child response — drive pages individually. Throughput is a platform constraint: one chassis replica consumes page-rerenders serially (~45–60s each). reconcile_headers.sh/leo_reconcile_bg.sh implement the idempotent pattern: each round re-fire only pages whose deployed HTML still shows the old artifact.
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
### Nav agent family and the three-tier authority model
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** 002(4): owner "currently populate_nav_tables action within pageflow-builder"; tiers described as model
- **what:** Navigation as first-class entity (groups: primary/subsection/content/legal/utility/external; contextual groups planned). Tier 1 strategist authority (new builds), Tier 2 nav-agent autonomous maintenance, Tier 3 drift detection vs original plan. nav-updater/nav-link-fixer handle drift and broken template links today; nav dedup guard recommended after B-029-1 duplicate nav items.
- **sources:** 002(4)#Navigation Agent Family; 024; 029 B-029-1
- **relations:** nav-updater never spawns; populate_nav_tables
- **verify-later:** nav drift check + dedup guard status

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Two nav systems and the GetNavItems fallback
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "Two Nav Systems (and they conflict)" — nav tables intended, pages flags legacy; partial population yields a mix (undated FOCUS, compiled ~2026-04/05)
- **what:** site_nav_groups/site_nav_items (populated by populate_nav_tables, read by GetNavItems) versus pages.in_header/in_footer legacy flags (GetHeaderNavFromPages fallback). GetNavItems tries tables first, falls back to pages — partial population mixes the two. Nav authority tiers designed (Tier 1 planner rebuild — only tier implemented; Tier 2 autonomous nav agent; Tier 3 drift detection). Nav state captured in snapshots and restorable via revert.
- **sources:** FOCUS_navigation.md#1, #2, #7
- **relations:** stale pages problem; nav discovery checks; site-design-planner navigation spec
- **verify-later:** GetNavItems fallback logic; whether Tier 2/3 exist

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Stale pages from previous builds polluting nav
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** "SyncPagesToDBAction uses ON CONFLICT (site_id, name) — it only overwrites matching page names" with fixes listed as "needed" (FOCUS); still item 15 in the errors-to-fix list
- **what:** Pages from prior builds keep in_header=true/status=active and appear in nav though absent from the current plan. Fix design: build_status='deployed' filters on the pages-table nav readers; SyncPagesToDB deactivates stale pages gated by a deactivate_stale_pages flag (new builds deactivate; maintenance/adopt flows preserve).
- **sources:** FOCUS_navigation.md#stale-pages; FOCUS_navigation_errors_to_be_fixed.md#15
- **relations:** two nav systems; adoption faithfulness (preserve semantics)
- **verify-later:** SyncPagesToDBAction current behaviour

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Nav discovery checks and fix agents
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** check/handler tables in FOCUS_navigation (broken_nav_links→nav-link-fixer; checkNavLayout/checkUnwantedElements→component-template-fixer; checkUnlinkedSiteComponents→site-component-linker; orphan_pages→rerender-pages/content-gap-planner)
- **what:** The nav slice of the improvement loop: quality/design/completeness discovery agents detect anchor-slug links, stacked nav (missing flex), unwanted search icons, unlinked header/footer components, orphan pages, missing logo img; dedicated fixers repair templates, relink components (clearing rendered_html + needs_rerender), and make orphans reachable. component-template-fixer's idempotency was case-sensitive, injecting responsive CSS 4× (fix: lowercase compare).
- **sources:** FOCUS_navigation.md#3, #4; FOCUS_navigation_HANDOFF_navigation_fix.md#problems-10
- **relations:** fallback header; duplicate header/footer
- **verify-later:** discovery agent checks arrays; fixInjectResponsiveCSS case fix

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Duplicate header/footer pathology (site-level components in pages.sections)
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** Data fixes applied 2026-04-11 (12 pages.sections rows cleaned, 24 page_components deleted); but 10 dirty rows reappeared by 04-13/14 — "plan_sections filter NOT deployed" (2026-04-20 investigation 7)
- **what:** pages.sections listed site-level component names alongside content sections; rebuilds rendered header/footer as page_components, then InjectHeader/InjectFooter added a second copy. Code fixes designed but pending at doc date: filterSiteLevelSections in PlanSectionsAction (prevents recurrence), skip-if-present guards in InjectHeader/InjectFooter. A discovery check for duplicate headers inside <main> also missing.
- **sources:** FOCUS_navigation_HANDOFF_navigation_fix.md (whole); HANDOFF_2026-04-20_error_investigations.md#7; FOCUS_navigation_errors_to_be_fixed.md#1-2
- **relations:** nav fix agents; page-build-handler
- **verify-later:** plan_sections_action.go for filterSiteLevelSections; component_library.go inject guards

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Nav quality mechanisms of 2026-04-17 (tiers, child-page exclusion, label trust, quick links)
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "What Was Deployed This Session" (2026-04-17): tiered priority, isChildPageURL, navLabelForPage, quick_links_html + footer template SQL
- **what:** populate_nav_tables gained a three-tier page priority (core / hubs+conversion / secondary, overflow to utility) replacing arbitrary nav_order truncation; child-page URL prefixes (/tools/, /blog/ …) excluded from all nav groups; nav labels trust page.NavLabel ≤30 chars and rendering no longer truncates to two words; footer Quick Links built from primary+utility groups via a new quick_links_html variable.
- **sources:** HANDOFF_2026-04-17_nav_empty_sections_footer(1).md#2-5
- **relations:** two nav systems; tool nav integration
- **verify-later:** populate_nav_tables_action.go navPriorityTier/isChildPageURL

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Hardcoded fallback nav/header defaults inventing structure
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "Defect (lines 310–318 of multipage_actions.go): … injects a hardcoded fallback nav — Home/About/Services/Contact" (2026-06-09); RenderFallbackHeader stacked-nav/search-icon behaviour in FOCUS_navigation
- **what:** Two brochure-default fallbacks fabricate structure when resolution fails: RenderFallbackHeader (generic header, stacked nav, unwanted search icon) and AssembleMultipageSiteAction's hardcoded 4-item nav — the primary source of phantom /services.html links. Resolution direction: fallbacks must derive from real pages (buildNavigationFromPages) or fail loud, never invent URLs.
- **sources:** FOCUS_internal_linking.md#2; FOCUS_navigation.md#header-footer-rendering
- **relations:** phantom-link validation; Tension #1 (silent confident fallbacks)
- **verify-later:** multipage_actions.go lines ~310-318; RenderFallbackHeader callers

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Tool nav integration
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** "Known bug (fixed): addToolToNav used wrong column names … failed silently"; remaining: tools listed individually in primary nav, labels too long (errors-to-fix items 3-5, 18)
- **what:** create_tool_component adds a page, page_component and nav entry per tool; column-name bug fixed, but grouping strategy (single "Tools" entry vs individual items) and label shortening remain open design work — feeding the site-design-planner navigation.tools_strategy spec.
- **sources:** FOCUS_navigation.md#5; FOCUS_navigation_errors_to_be_fixed.md#3-5
- **relations:** site-design-planner navigation spec; tools pipeline
- **verify-later:** addToolToNav; nav grouping of tool entries on live sites

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Nav sync & config-driven page deactivation
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** 001(0) "Nav Sync: Config-Driven Page Deactivation … deactivate pages not in the current plan"; deactivate_stale_pages config flag
- **what:** Header/footer nav displayed stale pages because `SyncPagesToDBAction`'s `ON CONFLICT` only overwrote matching names and nav queries didn't filter `build_status`. Fix: nav getters add `AND build_status = 'deployed'`, and a new-build flow deactivates pages absent from the current plan gated by `deactivate_stale_pages: true`.
- **sources:** WM/001_development_guide(0).md#nav-sync-config-driven-page-deactivation, ED/102_blog_handoff-2026-04-10.md#a-check_orphan_pagesgo-new-routing-logic
- **relations:** site plan reconciler nav auditor; link management; blog-listing handoff
- **verify-later:** SyncPagesToDBAction; GetHeaderNavFromPages/GetFooterNavFromPages; site_nav_items

<!-- SOURCE: U18_sql_for_agents.md -->
### Navigation maintenance: nav-updater and nav-link-fixer
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** 042 full definition ("Algorithmic only - no LLM calls"); nav-link-fixer in 075 idle-timeout list; 058 wires it as fixer for broken_nav_links findings.
- **what:** nav-updater refreshes nav tables from current pages (populate_nav_tables), re-renders header/footer/head and reassembles all deployed pages — explicitly distinguished from rerender-site, which reuses stale site_nav_items. nav-link-fixer repairs the `#{{.slug}}` anti-pattern in header/footer component templates (should be `{{.url}}`), then force re-renders site components and pages.
- **sources:** 042_nav_updater_agent.sql; 042b_nav_link_fixer_agent.sql; 058_quality_checks_and_fixers.sql
- **relations:** quality-discovery-agent's broken_nav_links check; orphan_nav finding; rerender pipeline
- **verify-later:** populate_nav_tables / fix_nav_link_templates actions

<!-- SOURCE: U19_sql_tables_components.md -->
### Navigation tables (site_nav_groups / site_nav_items)
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** DDL plus a real-site query result (primary/legal groups for a live site) and the applied global template fix converting anchor links to page URLs.
- **what:** First-class navigation model replacing scattered pages-table queries and the navigation_structures cache: groups per site (group_key primary/legal/utility/content, group_type, hierarchy via parent_group_id) containing typed items (page_link/external_link/anchor/section_header, FK to pages with SET NULL, position, status, metadata). Sites without rows fall back to Go logic querying pages directly. Render context supplies both .slug and .url per item; templates must link {{.url}} (061 fix purged href="#{{.slug}}" from all header/footer/nav templates).
- **sources:** docs/agent_docs/sql_for_tables/016_nav_tables.sql; docs/agent_docs/sql_for_tables/017_site_nav_groups.sql
- **relations:** site snapshots capture nav; component-based headers consume nav_items.
- **verify-later:** nav writer agent; fallback path in Go.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Global context injection for navigation
- **category:** navigation
- **status-signal:** superseded
- **status-evidence:** docs009/001 "Context Propagation... any component can access {{.Global.Sitemap}}"; docs012/002 adds explicit sitemap JSON to strategist output; superseded by nav tables + GetNavItems (docs017/019b "reads nav tables directly, falls back to pages table").
- **what:** Navigation treated as data, not structure: the strategist emits the sitemap first (labels, urls, in_header/in_footer flags), and it is passed down as a Global context object so header/footer templates range over it — pages invented by the strategist automatically appear in nav. Evolution chain: Global context → sitemap in page_plan → pages-table queries (deployed-only) → site_nav_groups/site_nav_items tables.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#3-Solving-Navigation; docs012_site_maps_and_components/002_site_map_integration.md; docs018_rerendering/003_website_builder_architecture_status_report.md#5
- **relations:** nav agent family; navigation-from-pages; three-tier authority model.
- **verify-later:** GetNavItems and populate_nav_tables in component_library.go.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Navigation agent family + three-tier authority
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** docs017/019b: "core responsibilities are implemented as the populate_nav_tables action... full standalone nav-agent is planned but not yet needed"; utility classification list and nav data flow marked (implemented).
- **what:** Navigation as a first-class entity: site_nav_groups/site_nav_items with typed groups (primary, subsection, content, legal, utility, external, contextual); populate_nav_tables classifies pages (FAQ/Blog/Careers etc. routed to utility even if in_header); GetNavItems serves header (primary, deployed-only) and footer (primary+utility+legal) rendering with pages-table fallback. Authority tiers: strategist owns structure at build; nav agent makes incremental decisions in maintenance; periodic drift detection compares current nav against the original plan ("drift may represent valid evolution").
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#1-Navigation-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Three-Tier-Authority-Model
- **relations:** navigation-from-pages (predecessor); nav-updater fix agent; current navigation FOCUS docs.
- **verify-later:** site_nav_groups/site_nav_items tables; populate_nav_tables action; standalone nav-agent existence.

<!-- SOURCE: U25_leopardess_social.md -->
### Header nav from pages.in_header + nav-label hygiene
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** L5_nav_and_ctas.sql header (2026-07-13): "Header nav is built from pages.in_header at render time (render_site_components_action.go:550), so setting in_header=false drops a page from the nav without deleting the page."
- **what:** Nav membership is data (pages.in_header) consumed at header render; decluttering is an UPDATE, not a template edit. Companion defect: nav_label defaults to raw `<title>` strings ("… | Leopardess Consulting") and needs short labels (AUDIT D3). Used to cut a ~15-item nav (including a blank 0-section page) to a business-buyer set.
- **sources:** docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql (header); docs/leopardessconsulting/AUDIT_verified_facts.md#D3
- **relations:** CTA-graph integrity (vonc); link-management
- **verify-later:** render_site_components_action.go:550; pages.in_header usage

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Nav agent family and the three-tier authority model
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** 002(4): owner "currently populate_nav_tables action within pageflow-builder"; tiers described as model
- **what:** Navigation as first-class entity (groups: primary/subsection/content/legal/utility/external; contextual groups planned). Tier 1 strategist authority (new builds), Tier 2 nav-agent autonomous maintenance, Tier 3 drift detection vs original plan. nav-updater/nav-link-fixer handle drift and broken template links today; nav dedup guard recommended after B-029-1 duplicate nav items.
- **sources:** 002(4)#Navigation Agent Family; 024; 029 B-029-1
- **relations:** nav-updater never spawns; populate_nav_tables
- **verify-later:** nav drift check + dedup guard status

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Two nav systems and the GetNavItems fallback
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "Two Nav Systems (and they conflict)" — nav tables intended, pages flags legacy; partial population yields a mix (undated FOCUS, compiled ~2026-04/05)
- **what:** site_nav_groups/site_nav_items (populated by populate_nav_tables, read by GetNavItems) versus pages.in_header/in_footer legacy flags (GetHeaderNavFromPages fallback). GetNavItems tries tables first, falls back to pages — partial population mixes the two. Nav authority tiers designed (Tier 1 planner rebuild — only tier implemented; Tier 2 autonomous nav agent; Tier 3 drift detection). Nav state captured in snapshots and restorable via revert.
- **sources:** FOCUS_navigation.md#1, #2, #7
- **relations:** stale pages problem; nav discovery checks; site-design-planner navigation spec
- **verify-later:** GetNavItems fallback logic; whether Tier 2/3 exist

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Stale pages from previous builds polluting nav
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** "SyncPagesToDBAction uses ON CONFLICT (site_id, name) — it only overwrites matching page names" with fixes listed as "needed" (FOCUS); still item 15 in the errors-to-fix list
- **what:** Pages from prior builds keep in_header=true/status=active and appear in nav though absent from the current plan. Fix design: build_status='deployed' filters on the pages-table nav readers; SyncPagesToDB deactivates stale pages gated by a deactivate_stale_pages flag (new builds deactivate; maintenance/adopt flows preserve).
- **sources:** FOCUS_navigation.md#stale-pages; FOCUS_navigation_errors_to_be_fixed.md#15
- **relations:** two nav systems; adoption faithfulness (preserve semantics)
- **verify-later:** SyncPagesToDBAction current behaviour

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Nav discovery checks and fix agents
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** check/handler tables in FOCUS_navigation (broken_nav_links→nav-link-fixer; checkNavLayout/checkUnwantedElements→component-template-fixer; checkUnlinkedSiteComponents→site-component-linker; orphan_pages→rerender-pages/content-gap-planner)
- **what:** The nav slice of the improvement loop: quality/design/completeness discovery agents detect anchor-slug links, stacked nav (missing flex), unwanted search icons, unlinked header/footer components, orphan pages, missing logo img; dedicated fixers repair templates, relink components (clearing rendered_html + needs_rerender), and make orphans reachable. component-template-fixer's idempotency was case-sensitive, injecting responsive CSS 4× (fix: lowercase compare).
- **sources:** FOCUS_navigation.md#3, #4; FOCUS_navigation_HANDOFF_navigation_fix.md#problems-10
- **relations:** fallback header; duplicate header/footer
- **verify-later:** discovery agent checks arrays; fixInjectResponsiveCSS case fix

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Duplicate header/footer pathology (site-level components in pages.sections)
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** Data fixes applied 2026-04-11 (12 pages.sections rows cleaned, 24 page_components deleted); but 10 dirty rows reappeared by 04-13/14 — "plan_sections filter NOT deployed" (2026-04-20 investigation 7)
- **what:** pages.sections listed site-level component names alongside content sections; rebuilds rendered header/footer as page_components, then InjectHeader/InjectFooter added a second copy. Code fixes designed but pending at doc date: filterSiteLevelSections in PlanSectionsAction (prevents recurrence), skip-if-present guards in InjectHeader/InjectFooter. A discovery check for duplicate headers inside <main> also missing.
- **sources:** FOCUS_navigation_HANDOFF_navigation_fix.md (whole); HANDOFF_2026-04-20_error_investigations.md#7; FOCUS_navigation_errors_to_be_fixed.md#1-2
- **relations:** nav fix agents; page-build-handler
- **verify-later:** plan_sections_action.go for filterSiteLevelSections; component_library.go inject guards

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Nav quality mechanisms of 2026-04-17 (tiers, child-page exclusion, label trust, quick links)
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "What Was Deployed This Session" (2026-04-17): tiered priority, isChildPageURL, navLabelForPage, quick_links_html + footer template SQL
- **what:** populate_nav_tables gained a three-tier page priority (core / hubs+conversion / secondary, overflow to utility) replacing arbitrary nav_order truncation; child-page URL prefixes (/tools/, /blog/ …) excluded from all nav groups; nav labels trust page.NavLabel ≤30 chars and rendering no longer truncates to two words; footer Quick Links built from primary+utility groups via a new quick_links_html variable.
- **sources:** HANDOFF_2026-04-17_nav_empty_sections_footer(1).md#2-5
- **relations:** two nav systems; tool nav integration
- **verify-later:** populate_nav_tables_action.go navPriorityTier/isChildPageURL

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Hardcoded fallback nav/header defaults inventing structure
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "Defect (lines 310–318 of multipage_actions.go): … injects a hardcoded fallback nav — Home/About/Services/Contact" (2026-06-09); RenderFallbackHeader stacked-nav/search-icon behaviour in FOCUS_navigation
- **what:** Two brochure-default fallbacks fabricate structure when resolution fails: RenderFallbackHeader (generic header, stacked nav, unwanted search icon) and AssembleMultipageSiteAction's hardcoded 4-item nav — the primary source of phantom /services.html links. Resolution direction: fallbacks must derive from real pages (buildNavigationFromPages) or fail loud, never invent URLs.
- **sources:** FOCUS_internal_linking.md#2; FOCUS_navigation.md#header-footer-rendering
- **relations:** phantom-link validation; Tension #1 (silent confident fallbacks)
- **verify-later:** multipage_actions.go lines ~310-318; RenderFallbackHeader callers

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Tool nav integration
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** "Known bug (fixed): addToolToNav used wrong column names … failed silently"; remaining: tools listed individually in primary nav, labels too long (errors-to-fix items 3-5, 18)
- **what:** create_tool_component adds a page, page_component and nav entry per tool; column-name bug fixed, but grouping strategy (single "Tools" entry vs individual items) and label shortening remain open design work — feeding the site-design-planner navigation.tools_strategy spec.
- **sources:** FOCUS_navigation.md#5; FOCUS_navigation_errors_to_be_fixed.md#3-5
- **relations:** site-design-planner navigation spec; tools pipeline
- **verify-later:** addToolToNav; nav grouping of tool entries on live sites

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Nav sync & config-driven page deactivation
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** 001(0) "Nav Sync: Config-Driven Page Deactivation … deactivate pages not in the current plan"; deactivate_stale_pages config flag
- **what:** Header/footer nav displayed stale pages because `SyncPagesToDBAction`'s `ON CONFLICT` only overwrote matching names and nav queries didn't filter `build_status`. Fix: nav getters add `AND build_status = 'deployed'`, and a new-build flow deactivates pages absent from the current plan gated by `deactivate_stale_pages: true`.
- **sources:** WM/001_development_guide(0).md#nav-sync-config-driven-page-deactivation, ED/102_blog_handoff-2026-04-10.md#a-check_orphan_pagesgo-new-routing-logic
- **relations:** site plan reconciler nav auditor; link management; blog-listing handoff
- **verify-later:** SyncPagesToDBAction; GetHeaderNavFromPages/GetFooterNavFromPages; site_nav_items

<!-- SOURCE: U18_sql_for_agents.md -->
### Navigation maintenance: nav-updater and nav-link-fixer
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** 042 full definition ("Algorithmic only - no LLM calls"); nav-link-fixer in 075 idle-timeout list; 058 wires it as fixer for broken_nav_links findings.
- **what:** nav-updater refreshes nav tables from current pages (populate_nav_tables), re-renders header/footer/head and reassembles all deployed pages — explicitly distinguished from rerender-site, which reuses stale site_nav_items. nav-link-fixer repairs the `#{{.slug}}` anti-pattern in header/footer component templates (should be `{{.url}}`), then force re-renders site components and pages.
- **sources:** 042_nav_updater_agent.sql; 042b_nav_link_fixer_agent.sql; 058_quality_checks_and_fixers.sql
- **relations:** quality-discovery-agent's broken_nav_links check; orphan_nav finding; rerender pipeline
- **verify-later:** populate_nav_tables / fix_nav_link_templates actions

<!-- SOURCE: U19_sql_tables_components.md -->
### Navigation tables (site_nav_groups / site_nav_items)
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** DDL plus a real-site query result (primary/legal groups for a live site) and the applied global template fix converting anchor links to page URLs.
- **what:** First-class navigation model replacing scattered pages-table queries and the navigation_structures cache: groups per site (group_key primary/legal/utility/content, group_type, hierarchy via parent_group_id) containing typed items (page_link/external_link/anchor/section_header, FK to pages with SET NULL, position, status, metadata). Sites without rows fall back to Go logic querying pages directly. Render context supplies both .slug and .url per item; templates must link {{.url}} (061 fix purged href="#{{.slug}}" from all header/footer/nav templates).
- **sources:** docs/agent_docs/sql_for_tables/016_nav_tables.sql; docs/agent_docs/sql_for_tables/017_site_nav_groups.sql
- **relations:** site snapshots capture nav; component-based headers consume nav_items.
- **verify-later:** nav writer agent; fallback path in Go.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Global context injection for navigation
- **category:** navigation
- **status-signal:** superseded
- **status-evidence:** docs009/001 "Context Propagation... any component can access {{.Global.Sitemap}}"; docs012/002 adds explicit sitemap JSON to strategist output; superseded by nav tables + GetNavItems (docs017/019b "reads nav tables directly, falls back to pages table").
- **what:** Navigation treated as data, not structure: the strategist emits the sitemap first (labels, urls, in_header/in_footer flags), and it is passed down as a Global context object so header/footer templates range over it — pages invented by the strategist automatically appear in nav. Evolution chain: Global context → sitemap in page_plan → pages-table queries (deployed-only) → site_nav_groups/site_nav_items tables.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#3-Solving-Navigation; docs012_site_maps_and_components/002_site_map_integration.md; docs018_rerendering/003_website_builder_architecture_status_report.md#5
- **relations:** nav agent family; navigation-from-pages; three-tier authority model.
- **verify-later:** GetNavItems and populate_nav_tables in component_library.go.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Navigation agent family + three-tier authority
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** docs017/019b: "core responsibilities are implemented as the populate_nav_tables action... full standalone nav-agent is planned but not yet needed"; utility classification list and nav data flow marked (implemented).
- **what:** Navigation as a first-class entity: site_nav_groups/site_nav_items with typed groups (primary, subsection, content, legal, utility, external, contextual); populate_nav_tables classifies pages (FAQ/Blog/Careers etc. routed to utility even if in_header); GetNavItems serves header (primary, deployed-only) and footer (primary+utility+legal) rendering with pages-table fallback. Authority tiers: strategist owns structure at build; nav agent makes incremental decisions in maintenance; periodic drift detection compares current nav against the original plan ("drift may represent valid evolution").
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#1-Navigation-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Three-Tier-Authority-Model
- **relations:** navigation-from-pages (predecessor); nav-updater fix agent; current navigation FOCUS docs.
- **verify-later:** site_nav_groups/site_nav_items tables; populate_nav_tables action; standalone nav-agent existence.

<!-- SOURCE: U25_leopardess_social.md -->
### Header nav from pages.in_header + nav-label hygiene
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** L5_nav_and_ctas.sql header (2026-07-13): "Header nav is built from pages.in_header at render time (render_site_components_action.go:550), so setting in_header=false drops a page from the nav without deleting the page."
- **what:** Nav membership is data (pages.in_header) consumed at header render; decluttering is an UPDATE, not a template edit. Companion defect: nav_label defaults to raw `<title>` strings ("… | Leopardess Consulting") and needs short labels (AUDIT D3). Used to cut a ~15-item nav (including a blank 0-section page) to a business-buyer set.
- **sources:** docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql (header); docs/leopardessconsulting/AUDIT_verified_facts.md#D3
- **relations:** CTA-graph integrity (vonc); link-management
- **verify-later:** render_site_components_action.go:550; pages.in_header usage

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Link management: link_registry as first-class links + gap to planned links family
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** 024: schema + extract/sync + constraints + validation exist; links-orchestrator family "planned but not implemented"; delete-and-reinsert loses validation history (known)
- **what:** Every anchor in rendered HTML lives in link_registry (scope internal/page/external, type navigation/content/semantic, affiliate fields, validation state); extract_and_sync_links parses post-build (delete+reinsert per page); InjectLinkConstraints feeds valid pages into writer prompts to prevent invented links; validateInternalLinks warns (not blocks) on missing targets; nav structure is separate (site_nav_groups/items; populate_nav_tables classifies primary/legal/utility). Planned: link-crawler/validator/registry-sync/redirect-manager/affiliate-manager under an algorithmic links-orchestrator.
- **sources:** 024 full
- **relations:** orphan_pages/internal-linker; phantom-CTA bug; nav agent family
- **verify-later:** link_registry population; HTTP validation anywhere?

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Internal linking machinery and its defects
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** "current as of 2026-06-09. Grounded in multipage_actions.go, site_db_actions.go, queryresolve…"; defects: hardcoded fallback nav, unpopulated *_index_url specs, phantom /services.html
- **what:** The pages table (via upsertPage slug/url/nav_label) is the authority for link targets; nav built from real pages or DB nav structure; fixAnchorLinks bridges single-page anchors to multipage URLs; queryresolve fills list-hub cards; "Browse All X" buttons read *_index_url site_specs (inconsistent sources, often empty → href=""); ExtractAndSyncLinksAction maintains a per-page link_registry — the natural substrate for a phantom-link discovery check that does not yet exist. Hero CTA destinations are the linking half of the site-wide CTA defect; whether the CTA href is a resolvable field or hardcoded template is the gating open question.
- **sources:** FOCUS_internal_linking.md (whole)
- **relations:** hardcoded fallbacks; content quality catalogue; section-data reconciler
- **verify-later:** syncLinksToDB (records vs validates); link_registry schema; hero component input_schema

<!-- SOURCE: U05_content_quality_linking.md -->
### Hero CTA brochure-default defect (text↔destination mismatch)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Phantom hero CTA — FIX APPLIED 2026-06-26"; NOTES(44) session 2026-06-26: "snapshot 5946a27b… UPDATE 1; readback confirms".
- **what:** The generic hero/call-to-action component schemas carried brochure-site defaults (`cta_url ← pages.contact`, `secondary_cta_url ← pages.services`) while button text is LLM-written — so every hero site-wide linked "Browse Tools" to /contact.html and to the phantom /services.html. Root causes fixed in layers: the `pages` source fabrication (see next), schema/template hardening (Step 1), and finally the writer's select_sections path mismatch that discarded the resolver's correct hubs.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; HANDOFF_2026-06-09(2).md#next-task; NOTES_gamesdesign_silent_norebuild(44).md (2026-06-22/23 sessions); phantom_hero_ctas/001_context
- **relations:** sourceResolver pages fabrication; Step 1 schema hardening; internal-link-resolver agent; select_sections path-mismatch fix.
- **verify-later:** content_components rows hero/call-to-action input_schema; deployed guide-economy-basics hero HTML; page-content-writer default_config select_sections.

<!-- SOURCE: U05_content_quality_linking.md -->
### sourceResolver `pages` fabrication bug (phantom generator)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1: "Layer 1a … DEPLOYED + APPLIED + VERIFIED … `resolve` case `pages` no longer fabricates".
- **what:** `sourceResolver.resolve` (plan_sections_action.go) `case "pages"` fabricated `"/" + path + ".html"` and returned found=true for any non-existent page, so `on_missing` never fired and schema fallbacks were dead code — the machine that minted every hero phantom. Fixed to return the real URL or (nil,false). Blast radius was tight: hero + call-to-action are the only components with a `pages.*` source.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; running_notes_17(21).md#decisive-findings; step1_hero_cta_phantom_fix.sql (header)
- **relations:** hero CTA defect; Step 1 schema hardening; correct-or-absent principle.
- **verify-later:** platform plan_sections_action.go resolve() pages case.

<!-- SOURCE: U05_content_quality_linking.md -->
### Correct-or-absent principle + loud-but-non-blocking phantom policy
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** FOCUS_internal_linking(1) "Policy (settled this round)" 2026-06-10; running_notes_17(21) "Policy settled".
- **what:** The structural rule for all internal links: targets resolve from the real `pages` set, never fabricated or brochure-assumed; an unresolvable destination renders nothing (no button) rather than a broken/empty link, with a build-time signal so the absence isn't silent. Companion policy: a phantom/missing internal link is loud but non-blocking — a deploy-gate warning, not an error; the improvement loop resolves it.
- **sources:** FOCUS_internal_linking(1).md#through-line; running_notes_17(21).md#policy-settled; PLAN_b4_b5_hubs_and_link_resolver(3).md
- **relations:** unresolved_cta signal; validate_page_content gate; every phantom-fix layer.
- **verify-later:** validate_page_content.go warning severities; hero/CTA template gates in content_components.

<!-- SOURCE: U05_content_quality_linking.md -->
### Step 1 / Layer 1a hero+CTA schema/template hardening
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Verified: both components skip_field/fallbacks_gone/has_and_gate all true".
- **what:** SQL that sets hero/call-to-action CTA-url fields to `on_missing: skip_field`, removes phantom fallbacks (/contact.html, #features), and gates each button template on `{{if and .cta_text .cta_url}}` — so an unresolved CTA renders no button. Ships coupled with the Go resolve() fix (order matters: Go first, else the gate still receives a truthy phantom).
- **sources:** step1_hero_cta_phantom_fix.sql; check_linking_sql_applied.sql; RUNBOOK_linking_phantom_fixes(7).md#1
- **relations:** sourceResolver fabrication fix; internal-link-resolver restores destinations.
- **verify-later:** content_components hero/call-to-action html_template + input_schema; content_components_bak_cta0610 snapshot.

<!-- SOURCE: U05_content_quality_linking.md -->
### Layer 1b header/footer phantom fix (shared site components)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Layer 1b … Verified gone (§4 audit: 0 site_component findings)".
- **what:** Header/footer phantoms (/contact.html, /privacy.html, /terms.html) came from hardcoded ContentData in render_site_components_action.go, not templates or nav fallback. Fix at source: header cta_url resolved from the real contact page, footer legal links data-driven from `GetNavItems(NavGroupLegal)`, header CTA gated on cta_url. Being shared components, the edits benefit every site; nav itself was already real-page-derived.
- **sources:** layer1b_header_footer_phantom_fix.sql; FOCUS_internal_linking(1).md#shipped-this-round; NOTES(44) 2026-06-22 "render_site_components shows the phantom was already fixed for site components"
- **relations:** nav-link-fixer (can't reach ContentData literals); deprecated loadNavItems COALESCE(url,'/name.html') phantom source.
- **verify-later:** render_site_components_action.go lines ~141–233; footer-4-column/header-bold-gradient templates.

<!-- SOURCE: U05_content_quality_linking.md -->
### datahelpers/links.go — canonical link classification library
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** FOCUS_internal_linking(1) "Shared machinery (single source of truth)" — "Replaces three previously-divergent normalisers".
- **what:** One shared library (`ExtractHrefs`, `ClassifyLinkScope`, `IsAssetPath`, `NormalizePagePath`, `PageURLSet`) used by both the deploy gate and the post-deploy audit so they agree by construction. Replaced three divergent URL normalisers (validator lowercased+appended .html; audit stripped index.html; inventory ignored assets).
- **sources:** FOCUS_internal_linking(1).md#shared-machinery; running_notes_17(21).md#shipped
- **relations:** validate_page_content gate; check_phantom_internal_links audit.
- **verify-later:** platform/orchestration/datahelpers/links.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### check_phantom_internal_links post-deploy audit + surface routing
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-15(2) §1: "NOT YET ENABLED (deliberate, observe-only later)"; RUNBOOK_linking_phantom_fixes(7) §7a "gate cleared, ready when you choose".
- **what:** A discovery check scanning page_components + site_components rendered_html for phantom/empty internal links, routing per surface: site_component → nav-link-fixer (build), page_component → page-build-handler (content; a rebuild re-runs build-time resolution). Code-confirmed that per-finding pipeline/handler survive insertWorkItem (config check_pipeline is an unused default). Home agent settled as completeness-discovery-agent (content-integrity family). Deliberately inert until enabled; enabling ≠ autonomous remediation because findings land status='detected' (unclaimable). An earlier version routed page_component findings to internal-link-resolver directly — superseded; a stale duplicate z_context copy with that routing is marked for deletion.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#7a; FOCUS_internal_linking(1).md#shared-machinery; running_notes_17(21).md#§7-gate-RESOLVED; README_find_phantom_links.sql
- **relations:** observe-only enablement pattern; improvement-sweep re-enable gating; nav-link-fixer; internal-link-resolver.
- **verify-later:** discovery_checks/check_phantom_internal_links.go routeBySurface; completeness-discovery-agent run_checks.config.checks array (is phantom_internal_links present?).

<!-- SOURCE: U05_content_quality_linking.md -->
### B4/B5 Browse-All hub links via `section_index_for` queryresolve verb
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "B4/B5 (SQL) … Verified"; running_notes_17(21) 2026-06-14 "Browse All Games → /games/index.html … B4/B5 confirmed".
- **what:** The three list components' "Browse All X" buttons rendered href="" because they sourced `cta_url` from unpopulated, inconsistently-named `*_index_url` site_specs. Fix (option c of three considered): a new queryresolve verb `section_index_for:<type>` deriving the hub URL from real page relationships (shared-area lookup, URL-prefix fallback), plus template gates `{{if .cta_url}}`. Options (a) populate specs and (b) `pages.<hub-name>` source were rejected (per-site maintenance / baked naming convention). Notable trap discovered: for query.* fields the field loop never consults on_missing and would apply `fallback` on nil — hence source-only schema changes and gate-in-template.
- **sources:** PLAN_b4_b5_hubs_and_link_resolver(3).md; b4_b5_hub_links_schema.sql; b4_b5_hub_links_template_gate.sql
- **relations:** correct-or-absent; Tier-D list components; queryresolve subsystem.
- **verify-later:** queryresolve/section_index_for.go + Resolve switch; tool-list/game-list_pre_037/guide-list_pre_037 schemas.

<!-- SOURCE: U05_content_quality_linking.md -->
### internal-link-resolver agent
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Step 3 … Wiring confirmed LIVE"; agent row applied 2026-06-11 per RUNBOOK_linking_phantom_fixes(7).
- **what:** A dedicated sub-agent (spawned by page-content-writer, called once per page) whose single responsibility is resolving intent-appropriate internal link destinations from the real pages set — hero/CTA fields augmented in section resolved_data at build time. v1 rules are deterministic (top content hubs by nav_order, excluding about/contact/legal and the page's own hub); the agent boundary deliberately allows an LLM intent-matching upgrade without changing callers. Explicitly "a build-time augmenter, not a rendered-HTML patcher". Thin workflow (resolve_links → complete), logic in the resolve_internal_links Go action, targets validated via PageURLSet so it cannot emit a URL the gate flags.
- **sources:** internal_link_resolver_agent.sql; PLAN_b4_b5_hubs_and_link_resolver(3).md#step-3; running_notes_17(21).md#step-3
- **relations:** page-content-writer wiring; unresolved_cta signal; ctaFieldNames coverage gap; resolver lean-result follow-up.
- **verify-later:** agent_definitions row type='internal-link-resolver'; resolve_internal_links_action.go (chooseCTATargets, ctaFieldNames, setCTAField).

<!-- SOURCE: U05_content_quality_linking.md -->
### unresolved_cta build-time HITL signal
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) watch SQL "B) resolver distress signal — must stay 0"; observed 0 throughout the §5 batch.
- **what:** When a section has CTA text but no resolvable real-page destination, the resolver emits an `unresolved_cta` work item (needs_human_review; one per affected section, mirroring createDeferredItems, ON CONFLICT dedup). Rationale: the deploy gate cannot see a correctly-dropped button — there is no fingerprint in rendered HTML — so resolution time is the only place the absence is detectable pre-deploy.
- **sources:** PLAN_b4_b5_hubs_and_link_resolver(3).md#unresolved_cta; running_notes_17(21).md#step-3-completed; FOCUS_internal_linking(1).md#remaining
- **relations:** correct-or-absent principle; HITL machinery.
- **verify-later:** ResolveInternalLinksAction unresolved_cta emission; site_work_items item_type='unresolved_cta'.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-content-writer ↔ resolver wiring with regression-safe fallback chain
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** page_content_writer_link_resolver_wiring.sql applied 06-11, "all 7 verification columns correct" (running_notes_17(21) Deployment).
- **what:** Workflow-only wiring: spawn_link_resolver, resolve_links (call_agent, error_step falls through), select_sections (extract_fields with a fallback chain: resolver-augmented sections, else the original plan), loop repointed to sections_for_render. Designed so resolver failure is byte-identical to prior behaviour — which later proved double-edged: the fallback silently masked the path mismatch for two weeks.
- **sources:** page_content_writer_link_resolver_wiring.sql; running_notes_17(21).md#step-3-completed
- **relations:** select_sections path-mismatch bug (the fallback's dark side); result-contract work.
- **verify-later:** page-content-writer default_config workflow steps.

<!-- SOURCE: U05_content_quality_linking.md -->
### select_sections path-mismatch bug (resolver output computed then discarded)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-23 "phantom-CTA ROOT CAUSE CONFIRMED (path mismatch)"; fix applied+confirmed 2026-06-26 (snapshot 5946a27b).
- **what:** The resolver ran and returned augmented sections, but the call_agent envelope nests the reply at `resolved_links.response.link_resolution.sections_ready` while select_sections read top-level `resolved_links.sections_ready` → null → silent fallback to the un-augmented plan carrying the schema phantoms. One-line jsonb_set repoint fixed it; takes effect only on a full content build (a bare re-render doesn't re-run the resolver). Two follow-ups remain open: ctaFieldNames matches only exact "hero"/"call-to-action" (variants like hero-about/gauntlet-cta never resolve), and the resolver returns its whole echoed collected_data with empty final_result (should return a lean {sections_ready, unresolved}).
- **sources:** HANDOFF_page_pipeline(11).md#3; NOTES_gamesdesign_silent_norebuild(44).md 2026-06-23/26; phantom_hero_ctas/001_context
- **relations:** result-contract resolution (the `output` mapping form not flattening is the sibling defect); wiring fallback chain.
- **verify-later:** page-content-writer select_sections config; guide-economy-basics hero has_phantom_cta after build e26cd02f.

<!-- SOURCE: U05_content_quality_linking.md -->
### link_registry — records but never validates (dormant substrate)
- **category:** link-management
- **status-signal:** abandoned
- **status-evidence:** FOCUS_internal_linking(1) finding 2: "syncLinksToDB never populates it … wired into no live workflow, so link_registry is empty. It is not a usable substrate today."
- **what:** A per-page link inventory table with a target_page_id column + FK to pages, intended as a phantom-check substrate, but the sync never populates target_page_id and ExtractAndSyncLinksAction is wired into no live workflow. The phantom audit reads rendered_html directly instead.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; running_notes_16_content_quality_and_internal_linking(1).md#part-1
- **relations:** check_phantom_internal_links (the live approach that superseded it).
- **verify-later:** link_registry row counts; ExtractAndSyncLinksAction callers.

<!-- SOURCE: U05_content_quality_linking.md -->
### nav-link-fixer agent (template-anchor scope only)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** README_find_phantom_links.sql output: nav-link-fixer exists, status experimental, is_active=t.
- **what:** The site_component-surface link fixer: find/replaces `#{{.slug}}`/`#{{.name}}` anchors inside html_template (fix_nav_link_templates_action.go). Its scope excludes ContentData values and literal anchors — which is why the B2/B3 header/footer phantoms had to be fixed at source in Go instead.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings-4; README_find_phantom_links.sql
- **relations:** Layer 1b; check_phantom_internal_links routing (site_component surface).
- **verify-later:** fix_nav_link_templates_action.go; nav-link-fixer agent_definitions row.

<!-- SOURCE: U05_content_quality_linking.md -->
### prepare_link_context available_pages gap on the work-item path
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** running_notes_17(21) watch item 2 (2026-06-12): "This path maps NO db_sync to the writer → prepare_link_context/available_pages get nothing … Pre-existing, resolver-independent."
- **what:** The writer's prepare_link_context builds an available-pages constraint for the LLM's in-prose internal linking, but on the work-item rebuild path no db_sync is mapped, so the constraint text is empty — the LLM writes prose links unconstrained. Independent of the resolver (which queries the DB directly). Candidate fixes noted: map db_sync, or make prepare_link_context load pages itself.
- **sources:** running_notes_17(21).md#page-build-handler-contract watch items
- **relations:** internal-link-resolver; page-content-writer.
- **verify-later:** prepare_link_context action; page-build-handler call_content_writer input_mapping.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Semantic linking domain decomposition (5 link types)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** docs012/003 taxonomy table ("Links are not one thing — at least 5 different things") and proposed link-management-group of six agents; docs012/006 concludes "Links live in components, registry is an index"; lifecycle and semantic agents remain unbuilt.
- **what:** Recognition that link work spans navigation (low complexity), content links/CTAs, semantic links (pillar↔cluster topic modelling — AI-heavy), cross-site/network/affiliate links, and technical links (sitemap/canonical/hreflang), each needing different mechanisms and lifecycles (news decays in days, campaign pages expire, products die). Proposed agent group: navigation-agent, seo-agent, lifecycle-agent, cross-site-agent, semantic-link-agent, link-validator.
- **sources:** docs012_site_maps_and_components/003_semantic_linking.md; docs012_site_maps_and_components/004_more_on_links.md; docs012_site_maps_and_components/006_start_concluding_links.md
- **relations:** link_registry; relationships table for semantic pairs; links agent family (docs017/019b, algorithmic-only subset); current link-management docs 024.
- **verify-later:** which of the six proposed agents exist; page relationships in relationships table.

<!-- SOURCE: U21_legacy_docs_b.md -->
### link_registry as derived index (links live in components)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** docs012/006 schema with scope/link_type/affiliate fields; docs012/012 pipeline step "5e. EXTRACT LINKS — Action: extract_and_sync_links; DB Write: link_registry".
- **what:** Links are never stored as primary data — they exist inside rendered components; link_registry is a queryable index derived by extraction after rendering, tracking source component/page/site, resolved internal targets, scope (internal/page/site/network/external), type (navigation/content/semantic/affiliate/reference), anchor text, rel attributes, affiliate provider/tag, and validation health. Enables broken-link detection, orphan detection, and affiliate compliance without duplicating truth.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.5; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-2; docs012_site_maps_and_components/007_link_migration.sql
- **relations:** links agent family heartbeat; validate_page_content; redirect-manager.
- **verify-later:** link_registry table + extract_and_sync_links action.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Links agent family (algorithmic, no-LLM link health)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** docs017/019b family table (link-crawler, link-validator, link-registry-sync, redirect-manager, affiliate-link-manager phase 2 — all "LLM? No") with heartbeat workflow and explicit non-goals.
- **what:** Deliberately judgment-free link maintenance: crawl modified pages' HTML, classify by URL pattern, resolve internals to page records, HEAD-check externals rate-limited, detect broken links and orphan pages, generate redirects on URL changes, track per-page link counts and empty anchors. Explicitly excluded: link placement, nav decisions, SEO strategy, related-content suggestions (LLM territory).
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#2-Links-Agent-Family
- **relations:** link_registry; semantic linking decomposition (the LLM parts deferred); redirect-manager fix agent.
- **verify-later:** links-orchestrator agent; site_redirects table.

<!-- SOURCE: U23_docs_root_vonc.md -->
### site_specs `cta` aspect + CTA graph audit (parked)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** cta aspect inserted 2026-07-02 (primary_url/secondary_url) — un-deferred two sections; CTA-map pass explicitly PARKED (user chose Option B 2026-07-07: leave the circular graph until the real arena exists).
- **what:** A per-site `site_specs` aspect `cta` supplies shared CTA URLs (`cta.primary_url`, `cta.secondary_url`) resolved into component fields (gauntlet-cta.cta_primary_url, system-stats.cta_url) — one populated source fixes all dependants. The vonc CTA graph was then found CIRCULAR (hero→archive, archive→home, gauntlet-cta→archive; only nav/footer reach the Gauntlet tool, and no arena page exists); a deliberate CTA-map pass is queued because CTA URLs are baked into rendered sections, so a proper refresh is a section rebuild, not string surgery.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:35 + #2026-07-02-~19:50; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-step-4-done + #2026-07-07-~16:30; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** plan_sections deferral; phantom CTA bug; unresolved_cta work items (self-resolve when hubs exist)
- **verify-later:** site_specs aspect='cta' rows; retarget SQL parked in notes

<!-- SOURCE: U23_docs_root_vonc.md -->
### Phantom CTA resolution bug (fabricated /{area}.html hero CTAs)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** 016b Part 4 (confirmed 2026-06-22 in deployed HTML, gamesdesign): hero carries two phantom CTAs from schema sources pages.contact/pages.services; "workflow-only fix staged" (select_sections reading resolved_links at the wrong path).
- **what:** Hero CTA resolution can produce constructed/fabricated URLs (`/contact.html`, `/services.html`) while the real hubs live elsewhere, because `select_sections` reads `resolved_links.sections_ready` (null) instead of `resolved_links.response.link_resolution.sections_ready`, falling back to the un-augmented plan; `resolve_internal_links` is a build-time augmenter (writes cta_url into resolved_data for the writer), explicitly not a rendered-HTML patcher, and `check_phantom_internal_links` routes page-link fixes to page-build-handler by design. Distinct from the interactive clobber; `page_rerender` does not re-resolve schema-sourced CTAs (ruled out as a link fix).
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 4 + update); docs/RUNBOOK_vonc_session(1).md#remaining-steps (unresolved_cta parking)
- **relations:** site_specs cta aspect; internal link management (024)
- **verify-later:** select_sections workflow path fix; resolve_internal_links action

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Hero/CTA link fabrication in `sourceResolver.resolve` — the "310–318 hardcoded fallback nav" hypothesis superseded
- **category:** link-management
- **status-signal:** superseded
- **status-evidence:** Archived `FOCUS_internal_linking.md` (2026-06-09): "**Defect (lines 310–318 of `multipage_actions.go`):** when nav resolution returns empty, `AssembleMultipageSiteAction` injects a **hardcoded fallback nav**... This generic brochure default is a primary source of the phantom `/services.html`." Live `FOCUS_internal_linking(1).md` (2026-06-10): "**Nav is already real-page-derived; the brochure fallback was not the live path.**... The header/footer phantoms... came from **hardcoded `ContentData` in `render_site_components_action.go`**... not from the `multipage_actions.go` 310–318 fallback nav... (**Correction to the earlier note that blamed 310–318**.)"
- **what:** The initial (2026-06-09) diagnosis of site-wide phantom links blamed a specific hardcoded fallback-nav code path (`multipage_actions.go:310-318`) as the likely root cause. The next day's investigation, grounded in reads of `render_site_components_action.go`, corrected this: nav was already correctly real-page-derived, and the actual mechanism was (a) `sourceResolver.resolve`'s `"pages"` case *fabricating* a URL (`"/"+path+".html"`) and returning `found=true` for any non-existent page (so schema `on_missing`/`fallback` never fired), plus (b) separately, hardcoded `ContentData` literals for header/footer CTAs and legal links.
- **sources:** content_quality_and_internal_linking/FOCUS_internal_linking.md (archived, 2026-06-09); live FOCUS_internal_linking(1).md (2026-06-10); running_notes_17(16) "Decisive findings"
- **relations:** component-template-fixer CTA-reuse assumption; link_registry hypothesis (below)
- **verify-later:** `plan_sections_action.go` `sourceResolver.resolve` current "pages" case; `render_site_components_action.go` ContentData construction.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `link_registry` as a phantom-link validation substrate — considered, found unusable, abandoned
- **category:** link-management
- **status-signal:** abandoned
- **status-evidence:** Archived `FOCUS_internal_linking.md` (2026-06-09): "**Link inventory.** `ExtractAndSyncLinksAction`... syncs them per page into `link_registry`... A per-page link inventory already exists — the natural substrate for a broken/phantom-link discovery check." Live `FOCUS_internal_linking(1).md` (2026-06-10): "**`link_registry` only records, never validates.** It *has* a `target_page_id` column + FK to `pages`, but `syncLinksToDB` never populates it. And `extract_and_sync_links` is wired into **no live workflow**, so `link_registry` is empty. It is not a usable substrate today."
- **what:** The internal-linking investigation initially proposed reusing the existing `link_registry` table/action as the base for a new phantom-link discovery check. Follow-up code reading found the table permanently empty in practice (the populating column is never written, and the syncing action isn't wired into any live workflow) — so the check that was actually built (`check_phantom_internal_links.go`) instead scans `rendered_html` directly via new shared helpers (`datahelpers/links.go`).
- **sources:** content_quality_and_internal_linking/FOCUS_internal_linking.md (archived); live FOCUS_internal_linking(1).md; running_notes_17(16) "Decisive findings"
- **relations:** hero/CTA link fabrication (above)
- **verify-later:** whether `ExtractAndSyncLinksAction`/`link_registry` were ever wired up subsequently.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Hub "Browse All X" link resolution — rejected design options
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** `PLAN_b4_b5_hubs_and_link_resolver(1).md`: "(a) Populate the `*_index_url` specs from real hubs. **Rejected** — per-site data to maintain; re-introduces the inconsistent-source brittleness. (b) `source: pages.<hub-name>` per component... bakes the `<area>-index` naming convention into each schema... (c) **Recommended.** A new `queryresolve` verb... `query.section_index_for:<type>`." Shipped per running_notes_17(16): "`section_index_for.go` — new `queryresolve` verb... B4/B5 — done."
- **what:** For the empty-href "Browse All Tools/Games/Guides" defect, two options were explicitly weighed and rejected in the design doc before settling on a new `queryresolve` verb: manually populating `*_index_url` site_specs (rejected as brittle, per-site maintenance) and a per-component `pages.<hub-name>` source (rejected as baking a naming convention into every schema). The chosen option — `query.section_index_for:<type>`, resolving the hub via shared `site_area_id`/URL-prefix — shipped and was confirmed working in deployed HTML.
- **sources:** content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md; running_notes_17(16)
- **relations:** hero/CTA link fabrication; internal-link-resolver agent
- **verify-later:** `queryresolve.go` `section_index_for` case in current code.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `internal-link-resolver` agent (Step 3) — dedicated intent-aware internal-link resolution
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** running_notes_17(16): "**Step 3 — completed (2026-06-11), all deliverables written**... Agent row (`internal_link_resolver_agent.sql`)... Writer wiring (`page_content_writer_link_resolver_wiring.sql`)... `unresolved_cta`: emitted in-Go... `status needs_human_review`." Confirmed live end-to-end: "Query D (corrected paths) on both completed rebuilds: `for_render=2`, `plan_count=2`, EQUAL ⇒ resolver augmented sections + writer loop consumed them."
- **what:** A new sub-agent of `page-content-writer` (modelled on `research-agent`, no persistence) that, at build time, resolves hero/CTA link destinations to intent-appropriate real pages (excluding the page's own hub, about/contact/legal) rather than a fixed contact page, validates every candidate against `datahelpers.PageURLSet`, and emits an `unresolved_cta` HITL signal when no destination can be found — the only place a "correctly dropped" (absent) button is detectable, since the deploy gate can't see an absence. Replaces the abandoned assumption that `component-template-fixer` already handled this.
- **sources:** content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md; running_notes_17(16) "Step 3" sections
- **relations:** component-template-fixer CTA-reuse assumption (superseded); identity-advisor/approval_mode (abandoned); hero/CTA fabrication fix
- **verify-later:** `internal_link_resolver_agent.sql`, `resolve_internal_links_action.go` current deployment/image-tag state.

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### link registry, cached navigation structures, and redirects (link-management foundation)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** Live Go code references `link_registry` and `navigation_structures` (e.g. platform/orchestration/actions/html_actions.go, site_db_actions.go, discovery_checks/check_phantom_internal_links.go, platform/orchestration/datahelpers/links.go), and a live doc `024_link_management_v2.md` exists — confirming this MVP schema's core concept shipped and was later versioned.
- **what:** The original link-management schema: a `link_registry` table indexing every link extracted from rendered components (source component/page/site, resolved target page/site, a `scope` of internal/page/site/network/external, a `link_type` of navigation/content/semantic/affiliate/reference, plus validation status for broken-link detection); `navigation_structures` as a **cached, versioned** JSONB nav tree per site+type (header/footer/mobile/sidebar), invalidated by a trigger on any `pages` INSERT/UPDATE/DELETE and rebuilt lazily via `get_current_navigation`/`build_navigation_for_site`; and a `redirects` table (301/302/307/410, hit_count, expiry). Deliberately reuses the existing generic `relationships` table for semantic content relationships (pillar/cluster, related-content, cross-site-reference) rather than inventing a parallel structure.
- **sources:** docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql
- **relations:** core client→network→site→page hierarchy (above, same migration file); link-management (024 anchor, 024_link_management_v2.md)
- **verify-later:** 024_link_management_v2.md — confirm what changed between this v1 schema and "v2"

<!-- SOURCE: U25_leopardess_social.md -->
### CTA-graph integrity (dead-end and circular primary actions)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** NOTES_provocations-index 2026-07-07: "Every primary action on the site dead-ends here" (pre-fix); 2026-07-09 "every primary CTA on the site resolves here"; CTA circularity "parked, Option B" pending a real arena page.
- **what:** The site's call-to-action graph as an auditable object: for two weeks every primary CTA (nav, hero, gauntlet-cta, lobby cards, provocation-card) pointed at an unbuilt page (404), invisible to any check; after the archive shipped, the graph is circular (hero → archive; archive → home; gauntlet-cta → archive) while the only real interactive surface (the Gauntlet tool) is reachable only via nav/footer. Decision Option B: leave until a real take-filing arena exists. Structural note: CTA URLs are baked into rendered sections, so a graph retarget is a section rebuild, not string surgery; brief-explanation's CTAs still carry '#' placeholders.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-index(4).md#2026-07-07; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#9.5; docs/social001_vonc_tiktok_social/tool_docs/NOTES_lobby-grid(6).md#2026-07-04
- **relations:** navigation; silent no-op success (the 404 destination was its product); link_registry
- **verify-later:** link_registry; CTA URLs in deployed vonc HTML

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Link management: link_registry as first-class links + gap to planned links family
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** 024: schema + extract/sync + constraints + validation exist; links-orchestrator family "planned but not implemented"; delete-and-reinsert loses validation history (known)
- **what:** Every anchor in rendered HTML lives in link_registry (scope internal/page/external, type navigation/content/semantic, affiliate fields, validation state); extract_and_sync_links parses post-build (delete+reinsert per page); InjectLinkConstraints feeds valid pages into writer prompts to prevent invented links; validateInternalLinks warns (not blocks) on missing targets; nav structure is separate (site_nav_groups/items; populate_nav_tables classifies primary/legal/utility). Planned: link-crawler/validator/registry-sync/redirect-manager/affiliate-manager under an algorithmic links-orchestrator.
- **sources:** 024 full
- **relations:** orphan_pages/internal-linker; phantom-CTA bug; nav agent family
- **verify-later:** link_registry population; HTTP validation anywhere?

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Internal linking machinery and its defects
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** "current as of 2026-06-09. Grounded in multipage_actions.go, site_db_actions.go, queryresolve…"; defects: hardcoded fallback nav, unpopulated *_index_url specs, phantom /services.html
- **what:** The pages table (via upsertPage slug/url/nav_label) is the authority for link targets; nav built from real pages or DB nav structure; fixAnchorLinks bridges single-page anchors to multipage URLs; queryresolve fills list-hub cards; "Browse All X" buttons read *_index_url site_specs (inconsistent sources, often empty → href=""); ExtractAndSyncLinksAction maintains a per-page link_registry — the natural substrate for a phantom-link discovery check that does not yet exist. Hero CTA destinations are the linking half of the site-wide CTA defect; whether the CTA href is a resolvable field or hardcoded template is the gating open question.
- **sources:** FOCUS_internal_linking.md (whole)
- **relations:** hardcoded fallbacks; content quality catalogue; section-data reconciler
- **verify-later:** syncLinksToDB (records vs validates); link_registry schema; hero component input_schema

<!-- SOURCE: U05_content_quality_linking.md -->
### Hero CTA brochure-default defect (text↔destination mismatch)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Phantom hero CTA — FIX APPLIED 2026-06-26"; NOTES(44) session 2026-06-26: "snapshot 5946a27b… UPDATE 1; readback confirms".
- **what:** The generic hero/call-to-action component schemas carried brochure-site defaults (`cta_url ← pages.contact`, `secondary_cta_url ← pages.services`) while button text is LLM-written — so every hero site-wide linked "Browse Tools" to /contact.html and to the phantom /services.html. Root causes fixed in layers: the `pages` source fabrication (see next), schema/template hardening (Step 1), and finally the writer's select_sections path mismatch that discarded the resolver's correct hubs.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; HANDOFF_2026-06-09(2).md#next-task; NOTES_gamesdesign_silent_norebuild(44).md (2026-06-22/23 sessions); phantom_hero_ctas/001_context
- **relations:** sourceResolver pages fabrication; Step 1 schema hardening; internal-link-resolver agent; select_sections path-mismatch fix.
- **verify-later:** content_components rows hero/call-to-action input_schema; deployed guide-economy-basics hero HTML; page-content-writer default_config select_sections.

<!-- SOURCE: U05_content_quality_linking.md -->
### sourceResolver `pages` fabrication bug (phantom generator)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1: "Layer 1a … DEPLOYED + APPLIED + VERIFIED … `resolve` case `pages` no longer fabricates".
- **what:** `sourceResolver.resolve` (plan_sections_action.go) `case "pages"` fabricated `"/" + path + ".html"` and returned found=true for any non-existent page, so `on_missing` never fired and schema fallbacks were dead code — the machine that minted every hero phantom. Fixed to return the real URL or (nil,false). Blast radius was tight: hero + call-to-action are the only components with a `pages.*` source.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; running_notes_17(21).md#decisive-findings; step1_hero_cta_phantom_fix.sql (header)
- **relations:** hero CTA defect; Step 1 schema hardening; correct-or-absent principle.
- **verify-later:** platform plan_sections_action.go resolve() pages case.

<!-- SOURCE: U05_content_quality_linking.md -->
### Correct-or-absent principle + loud-but-non-blocking phantom policy
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** FOCUS_internal_linking(1) "Policy (settled this round)" 2026-06-10; running_notes_17(21) "Policy settled".
- **what:** The structural rule for all internal links: targets resolve from the real `pages` set, never fabricated or brochure-assumed; an unresolvable destination renders nothing (no button) rather than a broken/empty link, with a build-time signal so the absence isn't silent. Companion policy: a phantom/missing internal link is loud but non-blocking — a deploy-gate warning, not an error; the improvement loop resolves it.
- **sources:** FOCUS_internal_linking(1).md#through-line; running_notes_17(21).md#policy-settled; PLAN_b4_b5_hubs_and_link_resolver(3).md
- **relations:** unresolved_cta signal; validate_page_content gate; every phantom-fix layer.
- **verify-later:** validate_page_content.go warning severities; hero/CTA template gates in content_components.

<!-- SOURCE: U05_content_quality_linking.md -->
### Step 1 / Layer 1a hero+CTA schema/template hardening
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Verified: both components skip_field/fallbacks_gone/has_and_gate all true".
- **what:** SQL that sets hero/call-to-action CTA-url fields to `on_missing: skip_field`, removes phantom fallbacks (/contact.html, #features), and gates each button template on `{{if and .cta_text .cta_url}}` — so an unresolved CTA renders no button. Ships coupled with the Go resolve() fix (order matters: Go first, else the gate still receives a truthy phantom).
- **sources:** step1_hero_cta_phantom_fix.sql; check_linking_sql_applied.sql; RUNBOOK_linking_phantom_fixes(7).md#1
- **relations:** sourceResolver fabrication fix; internal-link-resolver restores destinations.
- **verify-later:** content_components hero/call-to-action html_template + input_schema; content_components_bak_cta0610 snapshot.

<!-- SOURCE: U05_content_quality_linking.md -->
### Layer 1b header/footer phantom fix (shared site components)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Layer 1b … Verified gone (§4 audit: 0 site_component findings)".
- **what:** Header/footer phantoms (/contact.html, /privacy.html, /terms.html) came from hardcoded ContentData in render_site_components_action.go, not templates or nav fallback. Fix at source: header cta_url resolved from the real contact page, footer legal links data-driven from `GetNavItems(NavGroupLegal)`, header CTA gated on cta_url. Being shared components, the edits benefit every site; nav itself was already real-page-derived.
- **sources:** layer1b_header_footer_phantom_fix.sql; FOCUS_internal_linking(1).md#shipped-this-round; NOTES(44) 2026-06-22 "render_site_components shows the phantom was already fixed for site components"
- **relations:** nav-link-fixer (can't reach ContentData literals); deprecated loadNavItems COALESCE(url,'/name.html') phantom source.
- **verify-later:** render_site_components_action.go lines ~141–233; footer-4-column/header-bold-gradient templates.

<!-- SOURCE: U05_content_quality_linking.md -->
### datahelpers/links.go — canonical link classification library
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** FOCUS_internal_linking(1) "Shared machinery (single source of truth)" — "Replaces three previously-divergent normalisers".
- **what:** One shared library (`ExtractHrefs`, `ClassifyLinkScope`, `IsAssetPath`, `NormalizePagePath`, `PageURLSet`) used by both the deploy gate and the post-deploy audit so they agree by construction. Replaced three divergent URL normalisers (validator lowercased+appended .html; audit stripped index.html; inventory ignored assets).
- **sources:** FOCUS_internal_linking(1).md#shared-machinery; running_notes_17(21).md#shipped
- **relations:** validate_page_content gate; check_phantom_internal_links audit.
- **verify-later:** platform/orchestration/datahelpers/links.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### check_phantom_internal_links post-deploy audit + surface routing
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-15(2) §1: "NOT YET ENABLED (deliberate, observe-only later)"; RUNBOOK_linking_phantom_fixes(7) §7a "gate cleared, ready when you choose".
- **what:** A discovery check scanning page_components + site_components rendered_html for phantom/empty internal links, routing per surface: site_component → nav-link-fixer (build), page_component → page-build-handler (content; a rebuild re-runs build-time resolution). Code-confirmed that per-finding pipeline/handler survive insertWorkItem (config check_pipeline is an unused default). Home agent settled as completeness-discovery-agent (content-integrity family). Deliberately inert until enabled; enabling ≠ autonomous remediation because findings land status='detected' (unclaimable). An earlier version routed page_component findings to internal-link-resolver directly — superseded; a stale duplicate z_context copy with that routing is marked for deletion.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#7a; FOCUS_internal_linking(1).md#shared-machinery; running_notes_17(21).md#§7-gate-RESOLVED; README_find_phantom_links.sql
- **relations:** observe-only enablement pattern; improvement-sweep re-enable gating; nav-link-fixer; internal-link-resolver.
- **verify-later:** discovery_checks/check_phantom_internal_links.go routeBySurface; completeness-discovery-agent run_checks.config.checks array (is phantom_internal_links present?).

<!-- SOURCE: U05_content_quality_linking.md -->
### B4/B5 Browse-All hub links via `section_index_for` queryresolve verb
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "B4/B5 (SQL) … Verified"; running_notes_17(21) 2026-06-14 "Browse All Games → /games/index.html … B4/B5 confirmed".
- **what:** The three list components' "Browse All X" buttons rendered href="" because they sourced `cta_url` from unpopulated, inconsistently-named `*_index_url` site_specs. Fix (option c of three considered): a new queryresolve verb `section_index_for:<type>` deriving the hub URL from real page relationships (shared-area lookup, URL-prefix fallback), plus template gates `{{if .cta_url}}`. Options (a) populate specs and (b) `pages.<hub-name>` source were rejected (per-site maintenance / baked naming convention). Notable trap discovered: for query.* fields the field loop never consults on_missing and would apply `fallback` on nil — hence source-only schema changes and gate-in-template.
- **sources:** PLAN_b4_b5_hubs_and_link_resolver(3).md; b4_b5_hub_links_schema.sql; b4_b5_hub_links_template_gate.sql
- **relations:** correct-or-absent; Tier-D list components; queryresolve subsystem.
- **verify-later:** queryresolve/section_index_for.go + Resolve switch; tool-list/game-list_pre_037/guide-list_pre_037 schemas.

<!-- SOURCE: U05_content_quality_linking.md -->
### internal-link-resolver agent
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Step 3 … Wiring confirmed LIVE"; agent row applied 2026-06-11 per RUNBOOK_linking_phantom_fixes(7).
- **what:** A dedicated sub-agent (spawned by page-content-writer, called once per page) whose single responsibility is resolving intent-appropriate internal link destinations from the real pages set — hero/CTA fields augmented in section resolved_data at build time. v1 rules are deterministic (top content hubs by nav_order, excluding about/contact/legal and the page's own hub); the agent boundary deliberately allows an LLM intent-matching upgrade without changing callers. Explicitly "a build-time augmenter, not a rendered-HTML patcher". Thin workflow (resolve_links → complete), logic in the resolve_internal_links Go action, targets validated via PageURLSet so it cannot emit a URL the gate flags.
- **sources:** internal_link_resolver_agent.sql; PLAN_b4_b5_hubs_and_link_resolver(3).md#step-3; running_notes_17(21).md#step-3
- **relations:** page-content-writer wiring; unresolved_cta signal; ctaFieldNames coverage gap; resolver lean-result follow-up.
- **verify-later:** agent_definitions row type='internal-link-resolver'; resolve_internal_links_action.go (chooseCTATargets, ctaFieldNames, setCTAField).

<!-- SOURCE: U05_content_quality_linking.md -->
### unresolved_cta build-time HITL signal
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) watch SQL "B) resolver distress signal — must stay 0"; observed 0 throughout the §5 batch.
- **what:** When a section has CTA text but no resolvable real-page destination, the resolver emits an `unresolved_cta` work item (needs_human_review; one per affected section, mirroring createDeferredItems, ON CONFLICT dedup). Rationale: the deploy gate cannot see a correctly-dropped button — there is no fingerprint in rendered HTML — so resolution time is the only place the absence is detectable pre-deploy.
- **sources:** PLAN_b4_b5_hubs_and_link_resolver(3).md#unresolved_cta; running_notes_17(21).md#step-3-completed; FOCUS_internal_linking(1).md#remaining
- **relations:** correct-or-absent principle; HITL machinery.
- **verify-later:** ResolveInternalLinksAction unresolved_cta emission; site_work_items item_type='unresolved_cta'.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-content-writer ↔ resolver wiring with regression-safe fallback chain
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** page_content_writer_link_resolver_wiring.sql applied 06-11, "all 7 verification columns correct" (running_notes_17(21) Deployment).
- **what:** Workflow-only wiring: spawn_link_resolver, resolve_links (call_agent, error_step falls through), select_sections (extract_fields with a fallback chain: resolver-augmented sections, else the original plan), loop repointed to sections_for_render. Designed so resolver failure is byte-identical to prior behaviour — which later proved double-edged: the fallback silently masked the path mismatch for two weeks.
- **sources:** page_content_writer_link_resolver_wiring.sql; running_notes_17(21).md#step-3-completed
- **relations:** select_sections path-mismatch bug (the fallback's dark side); result-contract work.
- **verify-later:** page-content-writer default_config workflow steps.

<!-- SOURCE: U05_content_quality_linking.md -->
### select_sections path-mismatch bug (resolver output computed then discarded)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-23 "phantom-CTA ROOT CAUSE CONFIRMED (path mismatch)"; fix applied+confirmed 2026-06-26 (snapshot 5946a27b).
- **what:** The resolver ran and returned augmented sections, but the call_agent envelope nests the reply at `resolved_links.response.link_resolution.sections_ready` while select_sections read top-level `resolved_links.sections_ready` → null → silent fallback to the un-augmented plan carrying the schema phantoms. One-line jsonb_set repoint fixed it; takes effect only on a full content build (a bare re-render doesn't re-run the resolver). Two follow-ups remain open: ctaFieldNames matches only exact "hero"/"call-to-action" (variants like hero-about/gauntlet-cta never resolve), and the resolver returns its whole echoed collected_data with empty final_result (should return a lean {sections_ready, unresolved}).
- **sources:** HANDOFF_page_pipeline(11).md#3; NOTES_gamesdesign_silent_norebuild(44).md 2026-06-23/26; phantom_hero_ctas/001_context
- **relations:** result-contract resolution (the `output` mapping form not flattening is the sibling defect); wiring fallback chain.
- **verify-later:** page-content-writer select_sections config; guide-economy-basics hero has_phantom_cta after build e26cd02f.

<!-- SOURCE: U05_content_quality_linking.md -->
### link_registry — records but never validates (dormant substrate)
- **category:** link-management
- **status-signal:** abandoned
- **status-evidence:** FOCUS_internal_linking(1) finding 2: "syncLinksToDB never populates it … wired into no live workflow, so link_registry is empty. It is not a usable substrate today."
- **what:** A per-page link inventory table with a target_page_id column + FK to pages, intended as a phantom-check substrate, but the sync never populates target_page_id and ExtractAndSyncLinksAction is wired into no live workflow. The phantom audit reads rendered_html directly instead.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; running_notes_16_content_quality_and_internal_linking(1).md#part-1
- **relations:** check_phantom_internal_links (the live approach that superseded it).
- **verify-later:** link_registry row counts; ExtractAndSyncLinksAction callers.

<!-- SOURCE: U05_content_quality_linking.md -->
### nav-link-fixer agent (template-anchor scope only)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** README_find_phantom_links.sql output: nav-link-fixer exists, status experimental, is_active=t.
- **what:** The site_component-surface link fixer: find/replaces `#{{.slug}}`/`#{{.name}}` anchors inside html_template (fix_nav_link_templates_action.go). Its scope excludes ContentData values and literal anchors — which is why the B2/B3 header/footer phantoms had to be fixed at source in Go instead.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings-4; README_find_phantom_links.sql
- **relations:** Layer 1b; check_phantom_internal_links routing (site_component surface).
- **verify-later:** fix_nav_link_templates_action.go; nav-link-fixer agent_definitions row.

<!-- SOURCE: U05_content_quality_linking.md -->
### prepare_link_context available_pages gap on the work-item path
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** running_notes_17(21) watch item 2 (2026-06-12): "This path maps NO db_sync to the writer → prepare_link_context/available_pages get nothing … Pre-existing, resolver-independent."
- **what:** The writer's prepare_link_context builds an available-pages constraint for the LLM's in-prose internal linking, but on the work-item rebuild path no db_sync is mapped, so the constraint text is empty — the LLM writes prose links unconstrained. Independent of the resolver (which queries the DB directly). Candidate fixes noted: map db_sync, or make prepare_link_context load pages itself.
- **sources:** running_notes_17(21).md#page-build-handler-contract watch items
- **relations:** internal-link-resolver; page-content-writer.
- **verify-later:** prepare_link_context action; page-build-handler call_content_writer input_mapping.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Semantic linking domain decomposition (5 link types)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** docs012/003 taxonomy table ("Links are not one thing — at least 5 different things") and proposed link-management-group of six agents; docs012/006 concludes "Links live in components, registry is an index"; lifecycle and semantic agents remain unbuilt.
- **what:** Recognition that link work spans navigation (low complexity), content links/CTAs, semantic links (pillar↔cluster topic modelling — AI-heavy), cross-site/network/affiliate links, and technical links (sitemap/canonical/hreflang), each needing different mechanisms and lifecycles (news decays in days, campaign pages expire, products die). Proposed agent group: navigation-agent, seo-agent, lifecycle-agent, cross-site-agent, semantic-link-agent, link-validator.
- **sources:** docs012_site_maps_and_components/003_semantic_linking.md; docs012_site_maps_and_components/004_more_on_links.md; docs012_site_maps_and_components/006_start_concluding_links.md
- **relations:** link_registry; relationships table for semantic pairs; links agent family (docs017/019b, algorithmic-only subset); current link-management docs 024.
- **verify-later:** which of the six proposed agents exist; page relationships in relationships table.

<!-- SOURCE: U21_legacy_docs_b.md -->
### link_registry as derived index (links live in components)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** docs012/006 schema with scope/link_type/affiliate fields; docs012/012 pipeline step "5e. EXTRACT LINKS — Action: extract_and_sync_links; DB Write: link_registry".
- **what:** Links are never stored as primary data — they exist inside rendered components; link_registry is a queryable index derived by extraction after rendering, tracking source component/page/site, resolved internal targets, scope (internal/page/site/network/external), type (navigation/content/semantic/affiliate/reference), anchor text, rel attributes, affiliate provider/tag, and validation health. Enables broken-link detection, orphan detection, and affiliate compliance without duplicating truth.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.5; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-2; docs012_site_maps_and_components/007_link_migration.sql
- **relations:** links agent family heartbeat; validate_page_content; redirect-manager.
- **verify-later:** link_registry table + extract_and_sync_links action.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Links agent family (algorithmic, no-LLM link health)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** docs017/019b family table (link-crawler, link-validator, link-registry-sync, redirect-manager, affiliate-link-manager phase 2 — all "LLM? No") with heartbeat workflow and explicit non-goals.
- **what:** Deliberately judgment-free link maintenance: crawl modified pages' HTML, classify by URL pattern, resolve internals to page records, HEAD-check externals rate-limited, detect broken links and orphan pages, generate redirects on URL changes, track per-page link counts and empty anchors. Explicitly excluded: link placement, nav decisions, SEO strategy, related-content suggestions (LLM territory).
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#2-Links-Agent-Family
- **relations:** link_registry; semantic linking decomposition (the LLM parts deferred); redirect-manager fix agent.
- **verify-later:** links-orchestrator agent; site_redirects table.

<!-- SOURCE: U23_docs_root_vonc.md -->
### site_specs `cta` aspect + CTA graph audit (parked)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** cta aspect inserted 2026-07-02 (primary_url/secondary_url) — un-deferred two sections; CTA-map pass explicitly PARKED (user chose Option B 2026-07-07: leave the circular graph until the real arena exists).
- **what:** A per-site `site_specs` aspect `cta` supplies shared CTA URLs (`cta.primary_url`, `cta.secondary_url`) resolved into component fields (gauntlet-cta.cta_primary_url, system-stats.cta_url) — one populated source fixes all dependants. The vonc CTA graph was then found CIRCULAR (hero→archive, archive→home, gauntlet-cta→archive; only nav/footer reach the Gauntlet tool, and no arena page exists); a deliberate CTA-map pass is queued because CTA URLs are baked into rendered sections, so a proper refresh is a section rebuild, not string surgery.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:35 + #2026-07-02-~19:50; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-step-4-done + #2026-07-07-~16:30; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** plan_sections deferral; phantom CTA bug; unresolved_cta work items (self-resolve when hubs exist)
- **verify-later:** site_specs aspect='cta' rows; retarget SQL parked in notes

<!-- SOURCE: U23_docs_root_vonc.md -->
### Phantom CTA resolution bug (fabricated /{area}.html hero CTAs)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** 016b Part 4 (confirmed 2026-06-22 in deployed HTML, gamesdesign): hero carries two phantom CTAs from schema sources pages.contact/pages.services; "workflow-only fix staged" (select_sections reading resolved_links at the wrong path).
- **what:** Hero CTA resolution can produce constructed/fabricated URLs (`/contact.html`, `/services.html`) while the real hubs live elsewhere, because `select_sections` reads `resolved_links.sections_ready` (null) instead of `resolved_links.response.link_resolution.sections_ready`, falling back to the un-augmented plan; `resolve_internal_links` is a build-time augmenter (writes cta_url into resolved_data for the writer), explicitly not a rendered-HTML patcher, and `check_phantom_internal_links` routes page-link fixes to page-build-handler by design. Distinct from the interactive clobber; `page_rerender` does not re-resolve schema-sourced CTAs (ruled out as a link fix).
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 4 + update); docs/RUNBOOK_vonc_session(1).md#remaining-steps (unresolved_cta parking)
- **relations:** site_specs cta aspect; internal link management (024)
- **verify-later:** select_sections workflow path fix; resolve_internal_links action

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Hero/CTA link fabrication in `sourceResolver.resolve` — the "310–318 hardcoded fallback nav" hypothesis superseded
- **category:** link-management
- **status-signal:** superseded
- **status-evidence:** Archived `FOCUS_internal_linking.md` (2026-06-09): "**Defect (lines 310–318 of `multipage_actions.go`):** when nav resolution returns empty, `AssembleMultipageSiteAction` injects a **hardcoded fallback nav**... This generic brochure default is a primary source of the phantom `/services.html`." Live `FOCUS_internal_linking(1).md` (2026-06-10): "**Nav is already real-page-derived; the brochure fallback was not the live path.**... The header/footer phantoms... came from **hardcoded `ContentData` in `render_site_components_action.go`**... not from the `multipage_actions.go` 310–318 fallback nav... (**Correction to the earlier note that blamed 310–318**.)"
- **what:** The initial (2026-06-09) diagnosis of site-wide phantom links blamed a specific hardcoded fallback-nav code path (`multipage_actions.go:310-318`) as the likely root cause. The next day's investigation, grounded in reads of `render_site_components_action.go`, corrected this: nav was already correctly real-page-derived, and the actual mechanism was (a) `sourceResolver.resolve`'s `"pages"` case *fabricating* a URL (`"/"+path+".html"`) and returning `found=true` for any non-existent page (so schema `on_missing`/`fallback` never fired), plus (b) separately, hardcoded `ContentData` literals for header/footer CTAs and legal links.
- **sources:** content_quality_and_internal_linking/FOCUS_internal_linking.md (archived, 2026-06-09); live FOCUS_internal_linking(1).md (2026-06-10); running_notes_17(16) "Decisive findings"
- **relations:** component-template-fixer CTA-reuse assumption; link_registry hypothesis (below)
- **verify-later:** `plan_sections_action.go` `sourceResolver.resolve` current "pages" case; `render_site_components_action.go` ContentData construction.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `link_registry` as a phantom-link validation substrate — considered, found unusable, abandoned
- **category:** link-management
- **status-signal:** abandoned
- **status-evidence:** Archived `FOCUS_internal_linking.md` (2026-06-09): "**Link inventory.** `ExtractAndSyncLinksAction`... syncs them per page into `link_registry`... A per-page link inventory already exists — the natural substrate for a broken/phantom-link discovery check." Live `FOCUS_internal_linking(1).md` (2026-06-10): "**`link_registry` only records, never validates.** It *has* a `target_page_id` column + FK to `pages`, but `syncLinksToDB` never populates it. And `extract_and_sync_links` is wired into **no live workflow**, so `link_registry` is empty. It is not a usable substrate today."
- **what:** The internal-linking investigation initially proposed reusing the existing `link_registry` table/action as the base for a new phantom-link discovery check. Follow-up code reading found the table permanently empty in practice (the populating column is never written, and the syncing action isn't wired into any live workflow) — so the check that was actually built (`check_phantom_internal_links.go`) instead scans `rendered_html` directly via new shared helpers (`datahelpers/links.go`).
- **sources:** content_quality_and_internal_linking/FOCUS_internal_linking.md (archived); live FOCUS_internal_linking(1).md; running_notes_17(16) "Decisive findings"
- **relations:** hero/CTA link fabrication (above)
- **verify-later:** whether `ExtractAndSyncLinksAction`/`link_registry` were ever wired up subsequently.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Hub "Browse All X" link resolution — rejected design options
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** `PLAN_b4_b5_hubs_and_link_resolver(1).md`: "(a) Populate the `*_index_url` specs from real hubs. **Rejected** — per-site data to maintain; re-introduces the inconsistent-source brittleness. (b) `source: pages.<hub-name>` per component... bakes the `<area>-index` naming convention into each schema... (c) **Recommended.** A new `queryresolve` verb... `query.section_index_for:<type>`." Shipped per running_notes_17(16): "`section_index_for.go` — new `queryresolve` verb... B4/B5 — done."
- **what:** For the empty-href "Browse All Tools/Games/Guides" defect, two options were explicitly weighed and rejected in the design doc before settling on a new `queryresolve` verb: manually populating `*_index_url` site_specs (rejected as brittle, per-site maintenance) and a per-component `pages.<hub-name>` source (rejected as baking a naming convention into every schema). The chosen option — `query.section_index_for:<type>`, resolving the hub via shared `site_area_id`/URL-prefix — shipped and was confirmed working in deployed HTML.
- **sources:** content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md; running_notes_17(16)
- **relations:** hero/CTA link fabrication; internal-link-resolver agent
- **verify-later:** `queryresolve.go` `section_index_for` case in current code.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `internal-link-resolver` agent (Step 3) — dedicated intent-aware internal-link resolution
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** running_notes_17(16): "**Step 3 — completed (2026-06-11), all deliverables written**... Agent row (`internal_link_resolver_agent.sql`)... Writer wiring (`page_content_writer_link_resolver_wiring.sql`)... `unresolved_cta`: emitted in-Go... `status needs_human_review`." Confirmed live end-to-end: "Query D (corrected paths) on both completed rebuilds: `for_render=2`, `plan_count=2`, EQUAL ⇒ resolver augmented sections + writer loop consumed them."
- **what:** A new sub-agent of `page-content-writer` (modelled on `research-agent`, no persistence) that, at build time, resolves hero/CTA link destinations to intent-appropriate real pages (excluding the page's own hub, about/contact/legal) rather than a fixed contact page, validates every candidate against `datahelpers.PageURLSet`, and emits an `unresolved_cta` HITL signal when no destination can be found — the only place a "correctly dropped" (absent) button is detectable, since the deploy gate can't see an absence. Replaces the abandoned assumption that `component-template-fixer` already handled this.
- **sources:** content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md; running_notes_17(16) "Step 3" sections
- **relations:** component-template-fixer CTA-reuse assumption (superseded); identity-advisor/approval_mode (abandoned); hero/CTA fabrication fix
- **verify-later:** `internal_link_resolver_agent.sql`, `resolve_internal_links_action.go` current deployment/image-tag state.

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### link registry, cached navigation structures, and redirects (link-management foundation)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** Live Go code references `link_registry` and `navigation_structures` (e.g. platform/orchestration/actions/html_actions.go, site_db_actions.go, discovery_checks/check_phantom_internal_links.go, platform/orchestration/datahelpers/links.go), and a live doc `024_link_management_v2.md` exists — confirming this MVP schema's core concept shipped and was later versioned.
- **what:** The original link-management schema: a `link_registry` table indexing every link extracted from rendered components (source component/page/site, resolved target page/site, a `scope` of internal/page/site/network/external, a `link_type` of navigation/content/semantic/affiliate/reference, plus validation status for broken-link detection); `navigation_structures` as a **cached, versioned** JSONB nav tree per site+type (header/footer/mobile/sidebar), invalidated by a trigger on any `pages` INSERT/UPDATE/DELETE and rebuilt lazily via `get_current_navigation`/`build_navigation_for_site`; and a `redirects` table (301/302/307/410, hit_count, expiry). Deliberately reuses the existing generic `relationships` table for semantic content relationships (pillar/cluster, related-content, cross-site-reference) rather than inventing a parallel structure.
- **sources:** docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql
- **relations:** core client→network→site→page hierarchy (above, same migration file); link-management (024 anchor, 024_link_management_v2.md)
- **verify-later:** 024_link_management_v2.md — confirm what changed between this v1 schema and "v2"

<!-- SOURCE: U25_leopardess_social.md -->
### CTA-graph integrity (dead-end and circular primary actions)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** NOTES_provocations-index 2026-07-07: "Every primary action on the site dead-ends here" (pre-fix); 2026-07-09 "every primary CTA on the site resolves here"; CTA circularity "parked, Option B" pending a real arena page.
- **what:** The site's call-to-action graph as an auditable object: for two weeks every primary CTA (nav, hero, gauntlet-cta, lobby cards, provocation-card) pointed at an unbuilt page (404), invisible to any check; after the archive shipped, the graph is circular (hero → archive; archive → home; gauntlet-cta → archive) while the only real interactive surface (the Gauntlet tool) is reachable only via nav/footer. Decision Option B: leave until a real take-filing arena exists. Structural note: CTA URLs are baked into rendered sections, so a graph retarget is a section rebuild, not string surgery; brief-explanation's CTAs still carry '#' placeholders.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-index(4).md#2026-07-07; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#9.5; docs/social001_vonc_tiktok_social/tool_docs/NOTES_lobby-grid(6).md#2026-07-04
- **relations:** navigation; silent no-op success (the 404 destination was its product); link_registry
- **verify-later:** link_registry; CTA URLs in deployed vonc HTML
