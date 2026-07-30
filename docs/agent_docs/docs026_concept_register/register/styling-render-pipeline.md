# Register — styling-render-pipeline

> **covers-through: 2026-07-30** · STY-049 added 2026-07-16 and STY-050 added
> 2026-07-30 (post-freeze hand-patches).
> Everything else dates from the 2026-07-13 extraction freeze — absence
> here is not evidence of absence in the platform. See `bugs_open/106`.

48 concepts, consolidated from 120 raw extractions (60 unique blocks, each mechanically duplicated once in the cluster input — see consolidator note below) across units U01, U02, U03, U07, U08, U09, U13, U18, U19, U20, U21, U22, U23, U24c, U25.

> Consolidator note: the cluster input file `styling-nav-links.md` contains its entire content twice back-to-back (every one of 102 raw concept blocks across all three categories appears byte-identically twice — a mechanical file-doubling, not independent-unit re-extraction). That duplication was collapsed first without comment. The 48 entries below reflect a *second*, genuine layer of dedup: real cases where two different extraction units (e.g. U03 and U07, both reading the scheme-to-components thread from different doc snapshots) independently produced blocks describing the same mechanism.

### STY-001 — Styling render pipeline reference: two assembly paths and the scheme gap
- **status:** deployed
- **status-evidence:** 036 FINDING/THEORY-tagged reconstruction from code + live data.
- **what:** Umbrella architectural finding: stylesheet rendering and page-section rendering are separate code paths that only meet in the browser via CSS class names/custom properties. Catalogued findings include: `resolved_composition` doesn't record scheme (survives only on `layouts.scheme`); `buildSectionDefaults` emits `--section-*` only for dark bg/surface; five surface classes are duplicated between renderer and layouts (Phase 4.5 debt); hero/CTA components hardcode dark backgrounds defeating the scheme; the `.{function}-section` class contract is broken by hero/CTA; four overlapping chrome default stores exist, three dead; `RenderFallbackHeader` is hardcoded dark; `SectionStyles`/`component_selector` are dead code on the current path. This entry is the synthesis that the rest of the register's scheme/chrome entries (STY-005 through STY-015) resolve piece by piece.
- **sources:** 036 full; 016b light-site-dark-chrome entry
- **relations:** every scheme-to-components entry below; site component linkage
- **verify-later:** F-thread confirmations (update_site_defaults on composition path)

### STY-002 — CSS assembly pipeline (composable theme → styles.css)
- **status:** deployed
- **status-evidence:** "fully built path" (2026-05-12); render_css_from_spec_action.go deterministic, verified live schema for css_snippets.
- **what:** webdesign-agent flow: `analyze_design` (LLM) → `render_css_from_spec` (deterministic Go: theme composition from palettes/layouts/typography_sets FKs, css_snippets matched via `applies_to` JSONB overlap against the site's component functions, dark-section variants) → `deploy_css` git commit to `assets/css/styles.css` → B2 CDN sync. `css-patch-agent` is the bypass path for one-off fixes, patching the deployed file directly rather than the snippet library.
- **sources:** FOCUS-css_js_mechanisms.md#1; HANDOFF_2026-04-18_design_and_styling…md
- **relations:** composable theme migration 025; site-design-planner; STY-006 (palette merge rules)
- **verify-later:** render_css_from_spec_action.go, render_css_composition_helpers.go

### STY-003 — Component quality tracking (quality_score et al.)
- **status:** deployed
- **status-evidence:** "migration_component_quality.sql (applied) … compute_component_quality_action.go (pending Go deploy)" (2026-04-17); quality scoring described as working in the 04-18 handoff status table.
- **what:** `content_components` gains `template_variable_count`, `schema_field_count`, `template_closed`, `schema_template_synced`, `has_data_component`, `quality_score` (100 minus deductions), `quality_issues`; scored inline on store and by a component-quality-auditor agent. Planner prefers high scores, auditor targets low ones for regeneration; 43 pre-existing components had 0 template variables (content baked in) — regeneration targets, not mass-deletable.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#6, #Architecture-decisions
- **relations:** pre-store validation gates (STY-004); component-creator prompt tiers
- **verify-later:** compute_component_quality registry entry; quality_score population in DB

### STY-004 — Pre-store component validation gates + planning deferrals + empty-section filter
- **status:** deployed
- **status-evidence:** deployed 2026-04-17 (three checks before INSERT; sectionHasVisibleContent; empty-schema deferral); root incident: max_tokens=4000 truncation left unclosed `<section>`, CSS rendered as page text on vonc.com.
- **what:** Three layers preventing broken components/sections reaching pages: store-time rejection (template must contain `<section>`/`<div>`, balanced `<style>` tags, non-empty input_schema), plan-time deferral of content-type components with empty schemas, and render-time skipping of sections with under 10 chars visible text. Component-creator `max_tokens` raised 4000→16000 and prompt made context-aware.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#1, #6; HANDOFF_2026-04-17_nav_empty_sections_footer(1).md#1, #6, #7
- **relations:** LLM reliability tracks; STY-003 (quality tracking); STY-019 (visible-content filter, the render-time layer of this same defence)
- **verify-later:** store_generated_component_action.go validation block; rerender_single_page_action.go sectionHasVisibleContent

### STY-005 — Scheme-to-components P0: light-resolved site renders dark
- **status:** deployed
- **status-evidence:** PLAN "## CLOSED (2026-07-03) … closed on deployed evidence (RUNBOOK §SCHEME CLOSE: all nine grep checks pass; the stale-section fossil `var(--accent-color` is gone)."
- **what:** The defining P0 of the scheme-to-components thread: the chassis resolves each site to a light or dark scheme and the scheme travels correctly through layout/palette variables, but the component library was written dark-first, so a light-resolved site (idea.uk, tool-portal-light) deployed dark chrome and dark sections. The winning fix was completing the existing paired-variable standard rather than restructuring: one layout patched, ten templates de-hardcoded, chrome repointed + force-rerendered, then a full page-build-handler rebuild.
- **sources:** PLAN_scheme_to_components(1).md#CLOSED; RUNBOOK_scheme_to_components(50).md#SCHEME-CLOSE; HANDOFF_scheme_to_components_for_claude_code(1).md#The-problem; running_notes_scheme_to_components(55).md#Tk
- **relations:** paired-variable standard; STY-009 (hero ink model); STY-010 (hazard/band taxonomy); STY-013 (dual chrome paths)
- **verify-later:** deployed idea.uk B2 index.html greps; content_components.html_template for site-footer/call-to-action/hero; layouts.css_template for tool-portal-light

### STY-006 — Three-part styles.css assembly and core/specialised palette merge rules
- **status:** deployed
- **status-evidence:** Notes (Sc), read from render_css_from_spec_action.go bodies in full, 2026-06-30; corroborated independently by leopardess-thread code+data verification (RUNNING_NOTES turn 10).
- **what:** `RenderCSSFromSpecAction` builds styles.css in three appended parts: (1) the layout `css_template` rendered via Go text/template with palette/typo/token FuncMap helpers over merged maps — 8 core palette slots are spec-wins, specialised slots are theme-wins, typography is spec-wins, structure is layout-only; (2) component CSS from `css_snippets` matched via `applies_to` overlap; (3) the `buildSectionDefaults` luminance block. The core-vs-specialised split has a sharp edge: a site overriding core colours to dark/gold while sharing a seed palette serves visibly mixed output (white cards, navy header, blue gradient CTA on a black-and-gold page) because specialised slots never take the override — observed live on the leopardess site.
- **sources:** running_notes_scheme_to_components(55).md#Sc #Se; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-3, #Turn-10; docs/leopardessconsulting/scripts/L3_fork_palette.sql
- **relations:** STY-007 (buildSectionDefaults); layout CTA pair curation; per-site style fork
- **verify-later:** render_css_from_spec_action.go buildPaletteMap; css_snippets, palettes.colours, typography_sets tables

### STY-007 — buildSectionDefaults: luminance-keyed dark-only --section-* defaults
- **status:** deployed
- **status-evidence:** "`--section-*` is a DARK-ONLY override; light is the fallback. `buildSectionDefaults` returns '' unless bg or surface is dark" (running_notes, from color_util.go) — corroborated identically by a second unit.
- **what:** The renderer's only live per-section adaptation: `buildSectionDefaults` (color_util.go, WCAG `isDarkHex`/`pickReadableOnBackground`) emits a `body { --section-* }` block only when the merged palette's background or surface is dark, plus a dark-surface variant on 5 hardcoded surface classes (`.features/.services/.differentiators/.about/.faq-section`). A fully light site gets nothing and element rules fall through to `var(--color-*)`. Retained unchanged under the paired-variable decision as the whole-palette-darkness base/safety net, and is the pattern precedent for STY-022's alias block.
- **sources:** running_notes_scheme_to_components(55).md#Sf; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; SPEC_scheme_to_components.md#Decision-record; running_notes(22).md Sc, Sf
- **relations:** Colour Inheritance Model (STY-032); Phase 4.5 deferral (STY-014); STY-033 (section-contrast model, the component-side contract this renderer-side mechanism backstops)
- **verify-later:** color_util.go buildSectionDefaults/isDarkHex; emitted styles.css tail on a dark-palette site

### STY-008 — SectionStyles: built-but-disconnected per-section CSS mechanism, retired
- **status:** abandoned
- **status-evidence:** "`SectionStyles` is DEAD for current sites. None of the 18 active layouts reference `{{range .SectionStyles}}` … computed-but-unused"; "`SectionStyles` stays retired" (decision record) — independently confirmed by a second unit reading the same code.
- **what:** A fully-built but never-connected renderer mechanism: `queryDarkSectionsForCSS` + `buildCSSsectionStyles` compute per-component `{Function, ClassName: function+"-section", IsDark}` entries from `content_components.is_dark_section` and pass them to the layout template — which no active layout consumes. The `{function}-section` class contract it assumes is honoured unevenly (hero emits `.hero`, CTA emits `.cta-section`). Explicitly retired by the paired-variable decision; a textbook infrastructure-orphan, ~80% built and deliberately not revived.
- **sources:** running_notes_scheme_to_components(55).md#Sf #Si; SPEC_scheme_to_components.md#Decision-record; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; running_notes(22).md Sc, Sf, Si, Sn
- **relations:** superseded by paired-variable standard; STY-014 (Phase 4.5, the other renderer-owns design); STY-007 (buildSectionDefaults, the live sibling)
- **verify-later:** render_css_from_spec_action.go buildCSSsectionStyles/queryDarkSectionsForCSS still present and uncalled from layouts

### STY-009 — Hero ink model and the structural-dark exception
- **status:** deployed
- **status-evidence:** "W3b COMPLETE … ink in both inline branches, layered solid+single-hue color-mix gradient, five ink-referencing section vars…"; W3d extended it to the five hero-* variants — confirmed by a second unit's independent read of the same run.
- **what:** Image/layered hero sections define a per-branch `--hero-ink` custom property: the image branch sets `--hero-ink:#fff` under the structural-dark exception (an `rgba(0,0,0,x)` overlay guarantees darkness, so white text is always safe); the no-image branch sets `--hero-ink: var(--color-primary-text)` over a layered `var(--color-primary)` solid plus a single-hue color-mix gradient (15% toward the ink). Buttons become the inverse pair. Chosen after data showed imageless heroes are the common case (80/114 hero, 26/26 hero-*, reversing an assumption), and it also fixed a latent white-on-cyan failure on tool-portal-dark, not just the light-site problem.
- **sources:** running_notes_scheme_to_components(55).md#St #Su #Sv #Sw; w3b_01_hero_conversion.sql; RUNBOOK_scheme_to_components(50).md#HERO-(c)-DESIGN; running_notes(22).md St–Sw
- **relations:** section painting contract model; paired-variable standard; STY-024 (ambient pass-through, sibling pattern); STY-022 (D2a — `--hero-ink` later became an alias orphan)
- **verify-later:** hero + hero-* content_components.html_template current bytes; rendered index hero

### STY-010 — Hazard-class vs band-class declarer taxonomy (library blast radius)
- **status:** partial
- **status-evidence:** CHECK 3 RESULTS (2026-07-02): 84 active sections — 15 hex backgrounds, 37 self-declare `--section-*`, split ~18 hazard vs ~19 band; SCHEME CLOSE remaining work item 4 explicitly still open: "~10 remaining surface-painting declarers + ~17 band-class components (non-idea.uk)".
- **what:** The diagnostic taxonomy that sized every scheme-fix decision: hazard-class components declare dark `--section-*` while painting surface variables or nothing — live white-on-light bugs (footer, site-head, the five hero-* variants, brief-explanation); band-class components paint from primary/secondary/accent with white text — coherent today but blocking "fully light" (CTA, hero, social-proof, testimonials). Ten templates (the idea.uk-visible set) were hand-fixed; the non-idea.uk tail awaits a re-aimed fixer.
- **sources:** RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS #SCHEME-CLOSE; SPEC_scheme_to_components.md#W2 #W3; running_notes_scheme_to_components(55).md#Sn
- **relations:** fix_forced_text_colours re-aim (the tail vehicle); supervised fixer first-run
- **verify-later:** re-run the 3c split query; count remaining literal declarers among active sections

### STY-011 — Chrome selection path and the dead header_component_id column
- **status:** deployed
- **status-evidence:** "`install_site_composition` sets `style_collections.header_component_id`/`footer_component_id` = NULL … grep finds NO code that writes them non-NULL … effectively a DEAD column."
- **what:** Page-compile chrome resolution: `CompilePageSectionsAction` → `InjectHeader/InjectFooter/InjectHead` → `RenderHeader/RenderFooter` reads `style_collections.header_component_id` (always NULL, inserted with a "webdesign-agent populates these later" comment never honoured) → falls to `GetComponentByFunction("site-header")`, the single library-wide active component per function → else the hardcoded-dark fallback (STY-012). `RenderHead` looks up function `head` (the only head component is inactive, so builds always used the fallback head). Five other active header/footer functions (`*_pre_037`) are unreachable on this path.
- **sources:** running_notes_scheme_to_components(55).md#Sl #Sq #Se(C4); HANDOFF_scheme_to_components_for_claude_code(1).md#Established; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3d)
- **relations:** four chrome default stores; STY-012 (scheme-aware fallback chrome); STY-013 (dual chrome render paths)
- **verify-later:** component_library.go RenderHeader/RenderFooter/RenderHead/GetComponentByFunction; install_site_composition_action.go NULL insert + comment fate

### STY-012 — Scheme-aware fallback chrome (RenderFallbackHeader/Footer consume the pairs)
- **status:** deployed
- **status-evidence:** "slice 2+F live"; RUNBOOK 07-06-night State: "Deployed: slices 1 …, 2 (fallback chrome C/D + Debug tidy E)."
- **what:** The safety-net chrome functions were hardcoded dark (`background: ctx.PrimaryColor` default `#1a1a2e`, literal white text), so any site whose resolution chain broke got dark chrome regardless of scheme. Edits replaced the whole functions with `var(--color-header-bg, var(--color-surface))`/footer equivalent, text via `var(--color-header-text, var(--color-text))`, muted/borders via `color-mix` — safe library-wide because all 18 layouts were checked to set all four chrome vars. `RenderFallbackHead` was deliberately left unchanged (its only colour use is a `<meta theme-color>` where `var()` cannot work).
- **sources:** gobatch_02_component_library.md; running_notes_scheme_to_components(55).md#Sl #Tq #Ud; SPEC_scheme_to_components.md#W4(a)
- **relations:** STY-011 (chrome selection path); paired-variable standard; no-logger.Debug convention
- **verify-later:** component_library.go RenderFallbackHeader/RenderFallbackFooter current bodies; deployed image tag containing them

### STY-013 — Dual chrome render paths (build-fresh vs stale rerender-injected)
- **status:** deployed
- **status-evidence:** W4b executed and verified 2026-07-02 (header 3750→6258B, footer color-mix in); an earlier, less specific read of the same mechanism (from a parallel vonc-site thread) had flagged the paired-variable fix as not yet confirmed deployed — superseded by this dated, code-verified remedy.
- **what:** Chrome has two render paths: the page-compile path (page-build-handler → CompilePageSections → InjectHeader/Footer → RenderHeader/Footer) renders fresh, while `render_site_components` writes `site_components.rendered_html`, which the rerender path injects into pages unconditionally. The pinned-component join ignores `is_active`, and non-force runs skip non-empty slots — so stale renders of deactivated components ("fossilised" chrome, tagged by legacy `--accent-color` vars) persist indefinitely; nothing refreshes `site_components` on deactivation and `needs_rerender` just re-fossilises. Remedy pattern: repoint `site_components.component_id` to the active components first (rendered_html left in place so there's no chrome-less window), then trigger rerender-pages with `spec.refresh_site_components: true` — repoint-before-force_rerender ordering matters, or the old dark chrome re-renders.
- **sources:** running_notes_scheme_to_components(55).md#Sy #Sz #Ta #Td; w4b_01_repoint.sql; w4b_04_trigger_item.sql; RUNBOOK_scheme_to_components(50).md#W4b-RESULTS; docs/016b_debugging_guide_merged(3).md#light-site-renders-dark-chrome
- **relations:** rerender-pages v6 workflow; STY-031 (rerender pipeline); STY-011 (chrome selection path)
- **verify-later:** render_site_components_action.go force_rerender/skip logic; idea.uk site_components rows point at active components; site_components refresh-on-deactivation logic

### STY-014 — Phase 4.5 data-section-bg surface generalisation (deferred)
- **status:** aspirational
- **status-evidence:** SPEC consequences: "025 Phase 4.5 (`data-section-bg` surface generalisation) is deferred as a separate dark-site concern."
- **what:** Doc 025's already-designed decouple: components carry a `data-section-bg="surface"` attribute; the renderer replaces its hardcoded 5-surface-class list with an attribute selector; dual-write migration. Audited seriously and argued down: it solves a dark-site generalisation idea.uk never hits, its blanket "never self-declare" conflates hazardous surface declarations with load-bearing band declarations, and renderer ownership reintroduces component intent one hop away. Remains the designed answer if a dark site with surface sections outside the hardcoded 5 ever bites.
- **sources:** running_notes_scheme_to_components(55).md#Si #Sk #Sm; SPEC_scheme_to_components.md#Decision-record; HANDOFF_scheme_to_components_for_claude_code(1).md#Questioning-025
- **relations:** STY-007 (buildSectionDefaults, the 5-class list it generalises); paired-variable standard (chosen instead)
- **verify-later:** docs 025 §427–505; any data-section-bg attribute usage in components

### STY-015 — Explicit RenderContext.Scheme signal (Q1) — abandoned design
- **status:** abandoned
- **status-evidence:** "explicit `RenderContext.Scheme` is SECONDARY … This revises the Q1 emphasis in the PLAN"; never implemented anywhere in the executed fix.
- **what:** The original leading design: plumb the resolved scheme explicitly into both render entry points (`l.scheme` in the CSS loader SELECT + `themeComposition.Scheme`, and a `Scheme` field on `RenderContext`) so component templates receive an explicit light/dark signal. Overtaken when Check 1 showed the scheme already reaches components implicitly through palette `:root` values and luminance defaults — the components were the only thing defeating an already-working system. No scheme field was ever added.
- **sources:** PLAN_scheme_to_components(1).md#Q1; running_notes_scheme_to_components(55).md#Sb #Sf #Sk
- **relations:** superseded by paired-variable standard + implicit palette mechanism
- **verify-later:** RenderContext struct (component_library.go) — confirm no Scheme field exists

### STY-016 — Exact-field-name template binding with silent empty on miss (RenderTemplate)
- **status:** deployed
- **status-evidence:** "Correction 2 — the silent-empty mechanism is in RenderTemplate" (confirmed from uploaded component_library.go).
- **what:** `RenderTemplate` binds a page's `content_data` into a component's `html_template` by exact field name via Go text/template, then strips the `<no value>` tokens of unmatched placeholders to empty string, logging only a warning — no error. This is why a renamed or missing field fails silently rather than loudly; the entire "clobber" failure class in the content-quality thread rests on it.
- **sources:** NOTES(43).md §2 Correction 2, §8; BUNDLE(3).md §1
- **relations:** clobber failure mode; F1 guard (compensating control); STY-021 (R6f drift, a sibling silent-failure pattern)
- **verify-later:** platform/orchestration/actions/component_library.go:RenderTemplate; the `<no value>` cleanup

### STY-017 — Section readiness model (planSection source tiers, spec resolver)
- **status:** deployed
- **status-evidence:** "planSection required semantics CONFIRMED from plan_sections_action.go… 'Required with no fallback — defer' → deterministic defer → carry."
- **what:** Section fields declare a source (static with fallback; llm; Tier-C spec paths like `site_specs.cta.primary_url`) plus required/fallback attributes. planSection resolves each non-LLM field; a required field with no resolvable source and no fallback defers the section. The resolver reads site_specs per-aspect rows, checks presence not validity, and the stored⊕resolved merge (STY-018) persists resolved values into content_data at render time. Tier-C fields are, by design, never content_data keys.
- **sources:** NOTES(43).md §9s, §9u–§9x; RUNBOOK(49).md Part A
- **relations:** carry-forward path; STY-018; phantom-CTA lesson (spec presence ≠ URL validity, see LNK-003)
- **verify-later:** plan_sections_action.go (ensureSpecs, resolveSpecPath, on_missing switch); site_specs schema

### STY-018 — Stored⊕resolved merge writes resolved values back into content_data
- **status:** deployed
- **status-evidence:** "resolver persists cta_url into content_data (stored⊕resolved merge — expected)"; robot-hands cd_keys gained merged fallback keys.
- **what:** When a section re-render resolves fields (spec values, static fallbacks), the merged result is persisted back into the page's content_data as well as baked into rendered_html. Double-edged: it makes recoveries durable, but it is also a contamination carrier — bad fallback values can merge into dependents' content_data and survive a later schema fix, needing an explicit key-strip to clear.
- **sources:** NOTES(43).md §9w, §9an, §9ao; HANDOFF(7).md §Incident 2
- **relations:** contamination carrier; STY-017 (section readiness model); recovery playbook
- **verify-later:** rerender_page_sections persist-merged-content step

### STY-019 — Visible-content filter (≤10 chars) + data-runtime-fill assembler exemption
- **status:** deployed
- **status-evidence:** Base mechanism confirmed from rerender_single_page_action.go (content-quality thread); exemption fix independently confirmed "FIX VERIFIED" 2026-07-03 on the vonc site and again "deployed + verified" the same day from a second doc tree covering the same site.
- **what:** Page assembly drops any section whose rendered_html strips to ≤10 visible characters after removing style/script/tags/entities, logging only a warning — correct for genuinely empty shells but wrong for intentionally-empty runtime-filled ones (e.g. interactive Mode-B sections silently vanishing from a deployed index while page_components still said "deployed"). Fix: a `data-runtime-fill` marker on a section exempts it from the check regardless of build-time text; unmarked sections filter exactly as before. `assemblePage`/`sectionHasVisibleContent` are shared by the rerender and section-editor paths, so the exemption holds on both.
- **sources:** NOTES(43).md §1, §9ai, §9ar; RUNBOOK_section_assembly_drop(3).md#d4-result; PATCH_section_visible_content(1).go.txt; NOTES_provocation-card(12).md#2026-07-03
- **relations:** runtime-fill mechanism; STY-004 (the render-time layer of the same defence); clobber failure mode; verify-by-artifact discipline
- **verify-later:** rerender_single_page_action.go sectionHasVisibleContent + reRuntimeFill (lines ~429-452)

### STY-020 — Assembly membership and chrome model (page_components by position; pages.sections is metadata)
- **status:** deployed
- **status-evidence:** "assembly membership = page_components by position… pages.sections jsonb is NOT assembly membership" (corrected a wrong sections_listed=0 inference); three head shapes identified.
- **what:** A page artifact is assembled from `page_components` rows ordered by position, not from `pages.sections` (planning metadata only); head/header/footer come from site-scoped `site_components` rows, with `buildDefaultHead` emitting a minimal 5-line head when nothing is stored. A third, legacy builder (`assemble_from_library`) produced older artifacts with big inline heads — three coexisting head shapes repeatedly confused artifact forensics. `data-component` attributes exist on only some templates, so artifact greps on them undercount bands.
- **sources:** NOTES(43).md §9ah–§9al; RUNBOOK(49).md Part B
- **relations:** STY-019 (visible-content filter); STY-021 (R6f, missing :root head); chrome refresh gating
- **verify-later:** rerender_single_page_action.go; assemble_from_library (registry L493); site_components schema

### STY-021 — R6f — theming vocabulary drift (defined vs consumed CSS custom properties)
- **status:** deployed
- **status-evidence:** "R6f confirmed as the 'renders badly' mechanism; fresh rebuilds render WORSE"; mechanism narrowed to defined-vs-consumed drift.
- **what:** Component templates consume CSS custom properties (`var(--x, fallback)`) whose names drift from what the site's generated styles.css `:root` defines — 11 gap names in two patterns (synonyms like `--border-radius` vs `--radius`, and orphans like `--hero-ink`). Sections on undefined vocabulary render via per-component fallback values — a "fallback lottery" that goes dark-on-dark invisible on dark canvases. Newer generators put `:root` in styles.css (rootless heads); older sites carry inline `:root` heads. Every fresh rebuild worsened it as new templates minted new names.
- **sources:** NOTES(43).md §9al, §9am, §9bh, §9bi; RUNBOOK(49).md Part D; HANDOFF(7).md §R6f
- **relations:** STY-022 (D2a token aliases, the fix); STY-023 (D2b, prevention); STY-020 (assembly head model)
- **verify-later:** site styles.css :root contents vs template var() usage; robot-hands/vonc pre-fix stylesheets

### STY-022 — D2a — buildTokenAliases renderer-enforced compatibility bridge
- **status:** deployed
- **status-evidence:** "D2a VERIFIED in production 2026-07-06 — step 11 emits the alias block on a real gamesdesign pass"; curl shows the trailing compatibility-aliases `:root` block.
- **what:** A step-11 post-pass in RenderCSSFromSpecAction (mirroring STY-007's "renderer-enforced" pattern): after rendering, append a trailing `:root` block defining only the missing names from an 11-entry alias table (synonyms → canonical var() references, orphans → palette-safe literals). Idempotent, layout-defined values always win. Sites self-heal on their next design pass.
- **sources:** NOTES(43).md §9bj–§9bn; RUNBOOK(49).md Part D D2a
- **relations:** STY-021 (R6f, the drift it bridges); STY-023 (D2b, prevention side); STY-007 (pattern precedent)
- **verify-later:** render_css_from_spec_action.go buildTokenAliases + tokenAliases table; render_css_from_spec_alias_test.go

### STY-023 — D2b — canonical-token prevention (contract rule 11 + AuditTemplateTokens lint + prompt rule)
- **status:** partial
- **status-evidence:** "D2b in progress: lint coded… contract rule 11 drafted; prompt edit pending agent-identify" (2026-07-06, thread ends here).
- **what:** Stops new orphan tokens at the source: (1) contract rule 11 in 003's New Component Checklist restricting templates to canonical tokens + sanctioned aliases; (2) the generating agent's prompt to enforce the rule (agent identification still pending); (3) `AuditTemplateTokens` warn-only lint (canonicalCSSTokens allowlist = 39 theme names + 11 aliases, first-seen dedup, never rejects), pending call-site wiring. Rule 11 is the reciprocal of checklist item 6 (dark sections must SET `--section-*`) — consume-side vs set-side.
- **sources:** NOTES(43).md §9bo–§9bq; RUNBOOK(49).md Part D D2b
- **relations:** STY-022 (defines the sanctioned alias set); F8 lint; contracts-and-standards checklist
- **verify-later:** component_validation.go AuditTemplateTokens; whether wired into StoreGeneratedComponentAction; 003 doc rule 11 presence

### STY-024 — Ambient pass-through pattern for surface painters with fallback-less consumers
- **status:** deployed
- **status-evidence:** "Sanctioned pattern recorded: page/surface painters with fallback-less consumers pass the ambient context through" (executed 2026-07-02 17:22).
- **what:** For components that paint the page/surface colour but whose internal rules consume `var(--section-*)` without fallbacks, the safe conversion declares `--section-x: var(--color-x)` pass-throughs rather than deleting the declarations (deletion would fall to currentColor/transparent). Scheme-correct on both light and dark by definition since the core vars ARE the scheme. Companion finding: `rgba(var(--hex-var), α)` is invalid CSS that never rendered — color-mix is the working replacement.
- **sources:** RUNBOOK_scheme_to_components(18).md W3e RESULTS; running_notes(22).md Sx
- **relations:** paired-variable direction; creator-prompt fallback mandate
- **verify-later:** brief-explanation template pass-through block (later regenerated in the F8 saga — check current state)

### STY-025 — Interactive-section clobber + interactivity-aware save guard (preserve-sections)
- **status:** partial
- **status-evidence:** "CAUSE CONFIRMED; FIX PENDING" → 2026-06-24 "fix WRITTEN (un-deployed)"; still listed as a pending multi-page prerequisite at unit close.
- **what:** Any full rebuild regenerates a page from plan_sections, which knows nothing of an interactive tool stored only as a section's rendered_html — the game is silently discarded on rebuild. Layered fix, both in `save_page_sections` (the only place holding the markup): an interactivity guard blocking a non-interactive set replacing a deployed interactive one, plus carry-forward of existing interactive sections and source_item_id stamping.
- **sources:** 016b_debugging_guide_7_3_(7).md#open-threads-part-4; RUNBOOK_travelling_docs(38).md#§5.4; PLAN_travelling_docs(6).md#tool-assurance
- **relations:** multi-page prerequisites; pipeline documentation model
- **verify-later:** whether the patched save_page_sections_action.go deployed; page_component_history source_item_id

### STY-026 — Theme-layer render resolution (style_collection → css_theme, specs not read)
- **status:** deployed
- **status-evidence:** "Confirmed in the chassis actions: render context sets PrimaryColor/AccentColor/SecondaryColor from the resolved collection… Nothing in that path reads a site_specs design aspect" (2026-05-26).
- **what:** The live render path resolves colour/typography exclusively from `sites.style_collection_id → style_collections → css_themes`; `site_specs` design aspects influence it only upstream via the composition resolver. A NULL `style_collection_id` means no palette at all in render — the expected outcome for build-site-planner sites before the emit_design fix.
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#1, old2/HANDOFF_2026-05-07(1)#8
- **relations:** emit_design_items; site-design-planner
- **verify-later:** render context construction reads; theme variable injection status

### STY-027 — Render-off-build_status debt (planned-vs-rendered diff)
- **status:** partial
- **status-evidence:** "Render decision off build_status — the page-rerender agent appears to skip rebuilding planned-but-unrendered sections on a deployed page… Until then the build_status='needs_rebuild' reset is the workaround" (parked).
- **what:** A `needs_page` rebuild of an already-deployed page completes without rebuilding planned-but-missing components; the render short-circuits on `build_status='deployed'` instead of diffing planned sections against current page_components. The reset-to-needs_rebuild workaround is proven; the structural diff-driven render is open debt.
- **sources:** HANDOFF_2026-06-06#resolved (14q), HANDOFF_2026-06-09#later-parked
- **relations:** positive-evidence principle; pages.sections vs page_components distinction (STY-020)
- **verify-later:** page-rerender agent def + assemble/deploy action short-circuit

### STY-028 — site-asset-renderer: deterministic per-site JS snippet bundling
- **status:** deployed
- **status-evidence:** guidelines_compliance_check(1).md test plan + "Migration A/B/C applied" checklist; 113_site_asset_renderer.sql INSERT with verification queries; description "Deterministic — no LLM. Triggered when js_snippets or component set changes."
- **what:** `render_js_snippets_for_site_action.go` + the `site-asset-renderer` agent implements the JS-snippet deploy step: the global `js_snippets` table (is_active flag) is matched against a site's component functions via `applies_to`, concatenated into `assets/js/snippets.js` per site, committed to git, and loaded via a single `<script>` tag injected into the head template — mirroring `render_css_from_spec`'s existing pattern, and establishing the site-level shared-asset mechanism distinct from per-component inline JS. This closed the loader gap described in STY-034 for the js_snippets aspect specifically.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#The-JS-snippet-renderer-deliverable; 113_site_asset_renderer.sql
- **relations:** STY-034 (the gap this closed); latest-news component; contrasts with the never-built per-tool asset extraction (143)
- **verify-later:** js_snippets table (9 rows, 6 dormant is_active=false), site-asset-renderer agent_definition, head template script tag

### STY-029 — CSS component-list fallback bug (fake 5-item list masking real component inventory)
- **status:** deployed
- **status-evidence:** "Cause 1 (fallback fires) was fixed 2026-05-16" — status filter fix applied and verified across two sites.
- **what:** `extractCSSComponents` falls back to a hardcoded 5-item list whenever `site_context.all_component_functions` is empty. That field was empty because `loadPagesWithComponents`'s status filter matched nothing — every page's actual `status` value is `'active'`, never in the filter's list. Fixed to `WHERE p.status = 'active'`.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_visual_pipeline_css_and_component_lists.md#Cause-1
- **relations:** Assumed-status-values trap; STY-030 (applies_to granularity mismatch)
- **verify-later:** render_css_from_spec_action.go extractCSSComponents, design_actions.go loadPagesWithComponents

### STY-030 — CSS applies_to granularity mismatch (known issue, unfixed)
- **status:** partial
- **status-evidence:** "Cause 2 ... known issue, not yet fixed" — only 2 of ~21 snippets actually match real sites.
- **what:** Even after STY-029's fallback-list bug is fixed, `loadComponentCSSSnippets` does exact-text overlap between `css_snippets.applies_to` (generic terms) and real component functions — no exact overlap, no match, so most visual snippets never ship. Two proposed fixes: manually update every applies_to, or make matching lemma/slug-aware.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_visual_pipeline_css_and_component_lists.md#Cause-2
- **relations:** STY-029
- **verify-later:** loadComponentCSSSnippets in render_css_from_spec_action.go

### STY-031 — Rerender pipeline (rerender-pages, page-rerender, render-site-components, rerender-site)
- **status:** deployed
- **status-evidence:** 033–036 SQL migration sequence; idle timeouts for rerender-pages/page-rerender in 075; corroborated by the independently-documented docs018 rerender architecture ("ensure_site_record → render_site_components [force] → loop(call page-rerender) → trigger_deploy").
- **what:** The assembly/deployment half of the system, separated from content generation: `page_components` store rendered sections; `render_site_components` renders header/footer/head into `site_components`; `page-rerender` re-assembles a single page from stored sections (with skip detection) and deploys; `rerender-site` orchestrates site-wide re-render. Reassembles from stored `page_components.rendered_html` with current site-level chrome without touching content — strip old wrappers, apply current chrome, commit; includes contact-info injection from DB to overwrite hallucinated details. `needs_rerender` work items (priority 99, run last) are the standard "make fixes visible" side-effect from every fixer agent.
- **sources:** 033_rerender_pages_action.sql; 034_page_rerender_agent.sql; 035_render_site_components.sql; 036_rerender_site_agent.sql; docs018_rerendering/001_rerender_pages_summary.md; docs018_rerendering/006_build_path_rerender_path.md
- **relations:** nav-updater (adds nav refresh first); STY-013 (dual chrome paths); every fixer agent that returns needs_rerender
- **verify-later:** rerender_single_page / render_site_components actions; needs_rerender dedup guard

### STY-032 — CSS responsibility barrier + colour inheritance model
- **status:** deployed
- **status-evidence:** "CSS Responsibility Barrier Implementation — Global CSS handles all appearance... Components should NOT re-declare colors" plus component CSS-variables migration applied across all seeded components; independently stated as "Global CSS: all colors/fonts; Component CSS (inline): layout, positioning, structure only" in the architecture status report, tracing back to the original colour-inheritance bugfix (body sets the one default text colour; headings inherit; dark sections override at container level).
- **what:** Global styles.css (from webdesign-agent) owns colours/fonts; component CSS owns only layout/spacing, consuming CSS custom properties with fallbacks (`var(--color-primary, #...)`). This design-system rule fixed light-text-on-light-background failures: exactly one place sets default text colour, elements inherit, and dark sections override at container level so children inherit correctly. Components must not re-declare colours, with an explicit exception protocol for dark/inverted sections; audit queries exist to find violators.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#global-vs-local-css and #063b_hardcoded_colors_discovery; docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md#Bug-2; docs018_rerendering/003_website_builder_architecture_status_report.md#1
- **relations:** STY-033 (section-contrast model, current descendant); webdesign-agent
- **verify-later:** styles.css generation; remaining hardcoded colours in component templates; current webdesign/CSS prompts

### STY-033 — Section-contrast model (is_dark_section + --section-* variables)
- **status:** deployed
- **status-evidence:** Live COMMENT: "is_dark_section ... MUST set --section-text, --section-text-muted, --section-heading, --section-surface, --section-border on container"; 014 section-context variable migration in schema 005.
- **what:** Components with dark backgrounds are flagged `is_dark_section=true` and must define the `--section-*` variable set on their container so text/heading/surface colours invert correctly regardless of the global palette. Migration audited false positives (components using `#1a1a2e` as text colour, not background) and back-filled the variables per naming contract.
- **sources:** docs/agent_docs/sql_for_tables/005b_bk_content_components.sql#is_dark_section-comment; docs/agent_docs/sql_for_tables/005_content_components.sql#014-section-context
- **relations:** STY-032 (CSS responsibility barrier); STY-007 (buildSectionDefaults, the renderer-side backstop); component naming contract
- **verify-later:** is_dark_section rows vs presence of --section-* in their templates

### STY-034 — JS delivery paths & the js_snippets loader gap (historical)
- **status:** partial
- **status-evidence:** "declared in contracts, table populated, BUT NO LOADER IS WIRED UP" (verified 2026-05-12: 9 js_snippets rows, no reference in head templates or RenderHead); a later migration (113, see STY-028) subsequently shipped a site-level loader for this exact table, though the wider three/four-path picture below was never fully consolidated.
- **what:** Four coexisting JS delivery paths were catalogued: Path A (deployed) — per-component JS in `content_components.js_content`, extracted at store time and deployed as `/tools/assets/{function}.js`; Path B (aspirational at the time) — the `js_snippets` shared utility table with `applies_to` scoping, a registry of intentions with no runtime loader (later closed for js_snippets specifically by STY-028's site-asset-renderer); Path C (legacy anti-pattern) — inline `<script>` baked into html_template, violating contract 003 (news components still had this); Path D — html-assembler's `inject_js` flag with no visible reader. Interim tactic before the loader shipped: insert the snippet row AND duplicate inline.
- **sources:** FOCUS-css_js_mechanisms.md#2, #3, #4; HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#2; docs/agent_docs/sql_for_tables/044_css_snippets.sql
- **relations:** component contract 003; STY-028 (the eventual fix for path B); STY-035 (inline JS extraction contract, path A/C)
- **verify-later:** RenderHead in component_library.go; whether Path C/D were ever closed

### STY-035 — Inline JS extraction contract (separateInlineJS / js_content)
- **status:** deployed
- **status-evidence:** Early state (schema 044) noted "column added for future use... not consistently done"; later confirmed "CODE IS CORRECT (not the bug)" and, on 2026-07-07, "extraction pattern confirmed live on gauntlet-interface/latest-news/archetype-quiz (js_content + `<script src=` refs, no raw inline)."
- **what:** On component store, `separateInlineJS` extracts bare `<script>` blocks (regex requires a closing tag; deliberately skips attributed tags — `src`, `type="application/ld+json"`, `type="module"` stay inline) into `content_components.js_content`, replacing them with a `<script src="/tools/assets/{function}.js">` ref; `collectJSAssets` at rerender emits the per-component JS files. Interactive components should use this path rather than inline `<script>` in html_template; the column existed for a while before extraction was consistently applied (an early news component was a known violator). Known soft gaps: silent empty return on an unterminated `<script>`, no log when an attributed script is left inline.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#js_content; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~19:30 + #2026-07-07; docs/RUNBOOK_phase2_provocation_js(29).md#extraction-bug
- **relations:** STY-034 (two JS delivery paths); store-path validation hardening
- **verify-later:** store_generated_component_action.go separateInlineJS (~line 105); rerender_single_page_action.go collectJSAssets

### STY-036 — aggregate_webpage HTML assembly action (first-gen renderer)
- **status:** partial
- **status-evidence:** Used in robot-hands-complete-website workflows; replaced within docs004 by assemble_full_page/html-assembler and later by the current render pipeline.
- **stage2-verified (2026-07-14):** superseded → partial — aggregate_webpage is still registered live: registry.go:371 (Handler: AggregateWebpageAction, IsLocal:true), local_actions.go:124 enabled=true, and used in docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql (not a backup file, modified Nov 2025) — the claimed successor (STY-037/assemb...
- **what:** First-generation page renderer: wraps LLM-generated section content in a hard-coded HTML head (embedded CSS, nav) and footer, stitching named step outputs in a declared order into a complete page file. One action call per page.
- **sources:** docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md
- **relations:** successor: STY-037 (assemble_full_page + html-assembler), then STY-001 (current render pipeline)
- **verify-later:** does aggregate_webpage still exist in the action registry

### STY-037 — Content/structure separation: JSON content + html-assembler (assemble_full_page)
- **status:** superseded
- **status-evidence:** "structured JSON, not full HTML"; html-assembler agent with assemble_full_page action; the current render pipeline is the taxonomy successor.
- **what:** Separation of concerns that defines the modern pipeline: architect emits an empty `{{placeholder}}` template + content_requirements; content-creator emits pure content JSON; html-assembler merges template+content via Go templates then injects the CSS theme, tag-matched CSS snippets, and JS snippets into a complete document.
- **sources:** docs004_website_capture_project/006semantic_themes/README.022.description.md; docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md
- **relations:** successor: STY-001 (current pipeline) + component render contracts (doc 003)
- **verify-later:** assemble_full_page in registry; html-assembler agent row

### STY-038 — HTML action architecture (generate → process → validate)
- **status:** superseded
- **status-evidence:** "ALWAYS use the HTML actions instead of raw LLM calls... The architecture is already there — use it!"; replaced wholesale by component-template rendering in docs012+.
- **what:** A three-action pipeline for LLM page generation: `generate_html` (auto-gathers context, builds optimized prompt, extracts clean HTML), `process_html` (goquery parsing, meta/OG tags, responsive checks, minification), `validate_html` (structure, required elements, accessibility). Plus `assemble_html_parts` for chunking one huge page into structure/styles/content generations.
- **sources:** docs006_workflow_builder/008_20_plus_pages.md#The-HTML-Actions; docs006_workflow_builder/009_massive_multipage_sites.md#The-Actions-Available
- **relations:** superseded by content_components template rendering + render_mode matrix
- **verify-later:** html_actions.go survival/usage in current action registry

### STY-039 — Batched multipage generation (assemble_multipage_site)
- **status:** partial
- **status-evidence:** "for 20+ pages you need assemble_multipage_site... 5 batches × 4 pages = 80k tokens = WORKS"; later replaced: "Current (broken): spawn_multiple_writers ❌ Spawns 4 at once → New: loop."
- **stage2-verified (2026-07-14):** superseded → partial — registry.go:525 assemble_multipage_site still registered (Handler: AssembleMultipageSiteAction, IsLocal:true), local_actions.go:57 enabled=true, and used live in sql_for_agents_v1/v2 017_multipage_website_builder.sql (non-backup). loop action also registered (registry.go:47) and used in generate_pages_loop in 000_ba...
- **what:** Handling 6–200+ page sites within LLM output limits by generating pages in batches of 3–5 per call, generating shared CSS once, injecting navigation with active states, and streaming files to S3 (auto_store threshold pattern). Superseded by sequential per-page generation with the loop action after race conditions and quality problems.
- **sources:** docs006_workflow_builder/009_massive_multipage_sites.md#Quick-Decision-Tree; docs010_multitrack_flows_persona_architecture/019_start_here_document.md#Week-1
- **relations:** loop action; Kafka message size limits; stream_to_s3/auto_store
- **verify-later:** assemble_multipage_site action current form; auto_store config

### STY-040 — Asset bubble-up deduplication (proposal, never shipped)
- **status:** abandoned
- **status-evidence:** "Return Value Bubble-Up... use 100 buttons, button.css included once"; production instead uses a single global styles.css plus inline component `<style>` blocks.
- **what:** During recursive rendering, each component would return its HTML plus its CSS/JS dependency list; parents merge children's assets upward, and the root injects the deduplicated set once into the head. Tied to js_dependencies column proposals on content_components; never how the system actually shipped (STY-032's responsibility barrier is what shipped instead).
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#4-Solving-Assets
- **relations:** recursive component tree; STY-032 (what actually shipped)
- **verify-later:** js_dependencies column existence

### STY-041 — Assembly action consolidation (3 clear actions)
- **status:** deployed
- **status-evidence:** "You have [6 overlapping assembly actions]... Too much overlap. Proposed: 3 clear actions (assemble_page ...)"; later flows use assemble_page.
- **stage2-verified (2026-07-14):** partial → deployed — registry.go:505-528 now has exactly assemble_from_library, assemble_page, assemble_multipage_site (3 actions matching the proposal); grep for assemble_full_page/AssembleHTMLParts/WrapMultipage/html_actions in registry.go returns 0 hits — consolidation is complete, not merely proposed/partial.
- **what:** Rationalizing the accumulated assembly actions (assemble_from_library, assemble_full_page, AssembleHTMLParts, AssembleMultipageSite, WrapMultipage, html_actions) into a minimal set: assemble_page (one page from structure+styles+content), plus multipage and library assembly sharing code. A recurring theme in this codebase: action proliferation followed by consolidation.
- **sources:** docs010_multitrack_flows_persona_architecture/020_revised_consolidated_action_plan.md; docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md
- **relations:** STY-042 (component library unification); STY-045 (slot-based assembly proposal)
- **verify-later:** current action registry entries for assembly actions

### STY-042 — Component library unification (component_library.go)
- **status:** deployed
- **status-evidence:** "one source of truth for all component operations... RenderTemplate handles both Go-style {{.field}} and Handlebars-style {{field}}, {{#each}}, {{#if}}"; treated as load-bearing infrastructure in later docs.
- **what:** A shared Go module consolidating duplicated component code: component queries (by function, by ID, with fallback), style collection resolution, theme loading, dual-syntax template rendering, and high-level RenderHeader/RenderFooter/InjectHeader/InjectFooter/InjectHead used by both full-page assembly and header/footer injection into LLM-generated pages.
- **sources:** docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md#Summary; docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md
- **relations:** style collections; STY-044 (InjectHead bug); GetNavItems; STY-031 (rerender pipeline)
- **verify-later:** platform/orchestration/actions/component_library.go current contents

### STY-043 — page_components: component instances as the page's stored form
- **status:** deployed
- **status-evidence:** Schema introduced as "the bridge between content_components (templates) and actual page content"; later treated as established core ("Each section on a page maps to a page_components row").
- **what:** Every section of every page is a row: template reference (component_id), position/slot_name, nesting support (parent_component_instance_id), the rendered_html actually deployed, the content_data that produced it, content_hash for change detection, and semantic addressing fields. This is the storage foundation that makes rerendering, section editing, locking, and maintenance possible — arguably the single most consequential schema decision of this era.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.4; docs018_rerendering/010_section_editor_architecture.md#Component-Architecture; docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md#The-Data-Flow
- **relations:** content_data source-of-truth principle; STY-020 (assembly membership model); locks
- **verify-later:** page_components current columns incl. schema_snapshot/content_snapshot/build_status

### STY-044 — Head-inside-body bug and positional injection fixes
- **status:** deployed
- **status-evidence:** "cleanHTMLStructure keeps the LARGER <head> — wrong heuristic... InjectHead does in-place replacement — preserves wrong position" with concrete fixes ending "re-run rerender_pages."
- **what:** Two compounding rendering bugs: LLM sections sometimes emit full HTML documents, and the dedup heuristic kept the larger (misplaced) head while in-place head replacement preserved the wrong position. Fixes: remove all head blocks then always insert before `<body>`; dedup by position, not size. Exemplifies the fragility of regex injection that motivated the slot-based assembly proposal (STY-045).
- **sources:** docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md#Bug-1
- **relations:** STY-045 (slot-based assembly proposal); component_library.go InjectHead
- **verify-later:** current InjectHead/cleanHTMLStructure implementations

### STY-045 — Slot-based modular page assembly (proposal, partially adopted)
- **status:** partial
- **status-evidence:** "Status: Draft for discussion, Created 2026-02-06"; site_components + render_site_components subsequently appear in the build path, but page_sections-as-JSON did not fully replace rendered_html storage.
- **what:** Proposal to replace regex header/footer injection with pure concatenation of slots (doctype/head/header/sections/footer); pre-render site-level components once into a site_components table; store section content as schema-validated JSON and render only at assembly; seven single-responsibility agents proposed. Partially adopted: site_components and render_site_components shipped; JSON-first storage arrived instead as page_components.content_data source-of-truth (STY-043).
- **sources:** docs018_rerendering/007_proposed_modular_rerendering.md; docs018_rerendering/006_build_path_rerender_path.md
- **relations:** STY-044 (motivation); STY-043 (what actually landed instead)
- **verify-later:** site_components table + render_site_components action; page_sections existence

### STY-046 — CSS generation bug (webdesign-agent design_spec not applied)
- **status:** superseded
- **status-evidence:** "the webdesign-agent reported css_deployed: success:true ... But the deployed styles.css in git still contains the default blue template — the design_spec colors were never applied" (unsolved, 2026-03-02).
- **stage2-verified (2026-07-14):** partial → superseded — 031_webdesign_agent.sql line ~2105/3005: generate_css step's action was changed from execute_llm_prompt to render_css_from_spec (deterministic Go, 'no LLM') — the exact class of bug (LLM CSS not applying design_spec colours) is structurally closed by removing the LLM CSS-generation step; agent row status='active', v...
- **what:** A documented production defect: the webdesign-agent generates a correct design_spec (industry colours/fonts) but the generated/deployed CSS reverts to the default blue template. Three suspected causes: design_spec not reaching the template in structured form, an over-long prompt reproducing literal template CSS, or content_field resolution losing the CSS in the response envelope. Flagged for stage-2 debugging.
- **sources:** docs021.../024_handoff_summary_2026_03_02.md#the-css-bug
- **relations:** webdesign-agent; git_commit content_field resolution
- **verify-later:** webdesign-agent generate_css/deploy_css steps; extractFilesForGit content_field handling

### STY-047 — http2 deprecation fix at the nginx conf generator
- **status:** deployed
- **status-evidence:** "nginx 1.28.3 warns on `listen ... http2` (deprecated since 1.25) … the generator now emits version-neutral `listen 443 ssl;`" (2026-06-12).
- **what:** A field finding: newer nginx deprecates `listen ... http2` while the local container lacks the replacement `http2 on;`, so setup.sh's conf generator emits version-neutral `listen 443 ssl;` (with an opt-in comment for ≥1.25.1). Caught by fixing at the generator rather than per-box. Note: the source unit for this entry (traffic_probe/access-digest tooling) explicitly overlaps another extraction unit (U11, the live traffic_probe tree) — treat as the same underlying material if reconciling against U11 in a later stage.
- **sources:** traffic_probe_running_notes(27).md#2026-06-12-cert-issued
- **relations:** part of setup.sh nginx conf generation
- **verify-later:** setup.sh nginx server block listen directive; reconcile against U11 traffic_probe unit

### STY-048 — page-rerender mode contract and site-uniformity reconcile pattern
- **status:** deployed
- **status-evidence:** "Assemble mode (page_id, NO spec.reason) deploys unconditionally … the holdouts flipped immediately. Result: 27/30 active pages carry the gold header."
- **what:** page-rerender has two modes. Scoped mode (`spec.reason='section_data_resolved'` or `'image_landed'` + `spec.page_name`) re-renders each stored section from its `html_template` + stored `content_data` + re-resolved source fields — there is **no content-hash comparison anywhere in this path**. Its real bail-outs (`rerender_page_sections_action.go`, lines as of 2026-07-20): page-level — `skipped` when the page has no stored components (:157), `escalated` to the writer when any section's stored content_data is absent or missing a required `source:"llm"` field (:186; nothing is written or deployed in either case); section-level — stored HTML `carried` when the component can't be loaded (:229), `planSection` status != `"ready"` (:239), or `html_template` is empty (:251). Assemble mode (page_id, no reason) concatenates stored section HTML and re-embeds current header/footer unconditionally. Chrome-only changes still belong in assemble mode — not because of any hash skip, but because scoped mode's page-level bail-outs leave zero-section and incomplete-content pages undeployed. rerender-site's sequential page loop can stall on a lost child response, so driving pages individually is safer. Throughput is a platform constraint: one chassis replica consumes page-reenders serially (~45–60s each). Idempotent reconcile scripts re-fire only pages whose deployed HTML still shows the old artifact, round after round.
- > **CORRECTED 2026-07-20 (`bugs_closed/031`):** this entry previously asserted scoped mode "SKIPS pages whose content hash is unchanged". That was **never true of the code** — `grep -rn content_hash --include=*.go platform/ internal/` finds nothing in the rerender path, and `git log -S "content_hash"` over those files returns no commits. The claim was an inference from observed behaviour (pages not updating during the leopardess header reconcile, Turn 14) written in contract voice with no file:line; the observed skips were the page-level bail-outs above. A council seat quoted it as "the pipeline's own contract" and blocked a correct plan at HIGH severity (submission `7ef4de4e`, round 3 — refuted in round 4, seat approved). Caught by the travelling-docs thread's code-grep + live probe `478c44c9`.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#O8, #O9, #landmines-13/15; docs/leopardessconsulting/scripts/reassemble_pages.sh, reconcile_headers.sh, rerender_pages.sh
- **relations:** section-editor single-section path; STY-019 (assembler visible-content filter)
- **verify-later:** check_rerender_mode in page-rerender workflow; rerender_single_page_action.go; chassis replica count

### STY-049 — missingkey=zero silent-empty-render root pattern + escalate-not-blank guard (2026-07-16)
- **status:** partial
- **status-evidence:** Guard confirmed LIVE in the running pod as of 2026-07-16 (`v1.0.1123`, commit `9752bc68d` lands strictly between `b869469c8`/v1.0.1122 and `aec780f8e`/v1.0.1123 per `git merge-base --is-ancestor`) — grepped binary symbols `missingRequiredLLMFields`=2, `"escalating page to writer instead of blanking"`=1, `escalateRerenderToWriter`=4 (`aaa_fails_to_mend/004_HANDOFF_image_landing_blanks_article_body.md`). **004 itself is stale**: it still frames "recover 13 pages" and "fix ParseLLMJSON's 14 fixtures" as open (last edited 2026-07-16 11:58), but the later same-day `aaa_fails_to_mend/005_HANDOFF_2026-07-15_article_body_root_cause_is_truncation_FIXED.md` (mtime 17:52) shows both resolved — root cause was LLM output-token truncation (`max_tokens` 2000→8000 on the writer, unrelated to `missingkey=zero` itself), all 17 article-body instances regenerated healthy, and a live `go test ./platform/orchestration/actions/... -run TestParseLLMJSON` run (2026-07-16) confirms every ParseLLMJSON test green, including `TestParseLLMJSON_LiveEnvelopeDistribution`'s 1-repairable/13-unparseable split. The broader structural fail-loud guard (`RenderComponentAction`, `v3_site_actions.go:1632`, calling the same `missingRequiredLLMFields`) exists in current source but whether it's shipped in a *deployed* chassis image is unconfirmed (source-verified only, not pod-verified the way the first guard was).
- **what:** Go's `text/template` runs with `Option("missingkey=zero")` (`call_agent.go:1152`, unchanged by any of the recent fix commits), so any template referencing a key absent from its data map — e.g. `{{.content}}` when a section's `content_data` holds a raw, never-unwrapped LLM JSON envelope instead of a top-level `content` key — renders **empty silently, no error**. Two independent guards exist for two different call paths, both calling the same `json_envelope.go:192` `missingRequiredLLMFields` check: (1) `rerender_page_sections_action.go:191,491` (`escalateRerenderToWriter`) escalates the whole page to the writer instead of overwriting good HTML with a blank shell — this is what's live in prod (v1.0.1123); (2) `RenderComponentAction` in `v3_site_actions.go:1632` refuses to render at all when a component's `input_schema`-required content field is missing, logging "required content field(s) missing — refusing to render an empty section" — a broader, path-independent fail-loud, present in current source. **The root behaviour at `call_agent.go:1152` itself remains completely generic and unpatched** — any *other*, still-unidentified call site rendering a `required:true` field through this same template path has the identical silent-blanking exposure; only two call sites are known to be guarded so far. This is the same failure family as PBP-019/STY-004/STY-019's `sectionHasVisibleContent` empty-shell drop, TL-001's tool-widget-clobber hazard (interactive content silently destroyed by a content rebuild — the same "silent content loss during rerender" shape, different trigger), and the CLC-003 field-contract guard (tool-library.md) — a recurring, cross-category platform pattern (schema says required, renderer says silently empty) rather than a one-off bug, and explicitly named as sharing a class with a "product-page defect" (`HANDOFF_2026-07-14_empty_product_sections.md`).
- **sources:** aaa_fails_to_mend/004_HANDOFF_image_landing_blanks_article_body.md (full, esp. §3, §7 code map — status framing now superseded by 005 for items 1-2); aaa_fails_to_mend/005_HANDOFF_2026-07-15_article_body_root_cause_is_truncation_FIXED.md (the resolution); call_agent.go:1152; rerender_page_sections_action.go:191,491; v3_site_actions.go:1632; json_envelope.go:192
- **relations:** PBP-019 (sectionHasVisibleContent assembler filter); STY-004 (pre-store validation + empty-section filter); STY-019 (visible-content filter, same defence family); TL-001 (tool-widget-clobber hazard, tool-lifecycle.md — same "silent content loss during rerender" family, different trigger); CLC-003 (component-lifecycle field-contract guard, tool-library.md — same class, different mechanism); the article-body-json-envelope incident (now CLOSED — see article-body-json-envelope-workstream memory / the 005 doc for the data-recovery side, a separate and resolved issue from this structural rendering pattern)
- **verify-later:** confirm `RenderComponentAction`'s fail-loud path (`v3_site_actions.go:1632`) ships in a deployed chassis image and is pod-verified live, the same way the first guard was; whether any OTHER template-rendering call site besides `rerender_page_sections` and `RenderComponentAction` still has an unguarded `missingkey=zero` path — no fleet-wide audit of `content_components.input_schema` `required:true`/`source:"llm"` fields has been run yet to scope this

### STY-050 — Per-site chrome config via a gated `input_schema` field (`config.*` → `site_specs`) (2026-07-30)
- **status:** deployed
- **status-evidence:** Live on idea.uk 2026-07-30 — `GTM-PQ3WCTBD` verified in the rendered artefact of 20/20 re-assembled pages and on 19/19 fetchable live URLs (`curl -s https://idea.uk/<path> | grep -c googletagmanager` = 2). The eight other sites sharing `Document Head` render byte-identically (gated block, no key set). First use of the mechanism; the resolver path it rides on pre-existed.
- **what:** How to put a **per-site value into SHARED site chrome** without forking the component. The `head` slot resolves to just **3** components across the 14 live domains (`Document Head` ×9, `head-seo-standard` ×4, one fork) and `header` to 6, so a literal written into a chrome template ships to every site that shares it. Instead: (1) store the value in `site_specs`, aspect `site_config`, at a dotted path; (2) declare a **map-valued** `input_schema` field with `source: "config.<dotted.path>"`; (3) **gate** the template block with `{{if .field}}…{{end}}` so unset sites render byte-identically. `render_site_components_action.go:585-645` gap-fills the component's own `input_schema` through `newSourceResolver`, and `config.*` → `resolveConfigPath` (`plan_sections_action.go:516-527`) searches `site_specs` aspects `site_config`/`identity`/`design_intent`. **No Go change is required** — this is the intended use of the bugs_open/018 schema-driven fill, which until now had no per-site consumer. Worked example: Google Tag Manager on idea.uk (`idea_uk_vm_site/sql/p4_34_gtm_container.sql`).
- **three traps that make it silently not work:** (a) an `input_schema` entry whose value is a **scalar** is skipped as "not a field descriptor" (`:612-615`) — which is why `Document Head`'s own flat `title`/`description` scalars have never resolved; (b) an **ungated** placeholder renders empty on unset sites and, if it is a URL attribute, gets dropped and filed as a dead control (`DropDeadURLControls`, bugs_open/054); (c) writing only `html_template` changes **nothing on any page** — chrome is a stored artefact (`bugs_open/117`) and pages assemble from `site_components.rendered_html`, so **both** must be written or the change is inert (template-only) or temporary (artefact-only).
- **sources:** `analytics_gtm/PLAN_2026-07-30_analytics_gtm.md` §2-3; `analytics_gtm/RUNBOOK_analytics_gtm.md` §0-3; `idea_uk_vm_site/sql/p4_34_gtm_container.sql`; render_site_components_action.go:585-645, :604-607, :612-615; plan_sections_action.go:433-494, :516-527; LANDMINES.md ("Editing a site's head/header chrome edits every site that shares the component")
- **relations:** CLC-001 (shared content-component reuse — the same cross-site blast radius, here handled by gating rather than forking); STY-049 (`missingkey=zero` silent-empty render — the gate is what stops an unresolved field becoming a silent blank); bugs_open/018 (the schema-driven fill this consumes); bugs_open/117 (chrome as a stored artefact); bugs_open/054 (dead URL controls)
- **verify-later:** whether `site_config` should become the conventional aspect for all per-site render config (only `analytics.gtm_container_id` lives there today, on one site); whether the `noscript`-after-`<body>` case argues for a chassis-level slot, since `<body>` is a Go string literal in `assemblePage:577` and the top of the `header` slot is currently the only reachable seam
