# Register — design-composition

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

_Concept count retired 2026-08-09 — derived, not stored; run the drift pair in `000_concept_index.md`, or read `concept-register-drift-check`'s daily row (DOC-074). It said **80** and the file held **81**._ consolidated from 105 unique raw extractions across units
U01, U02, U03, U04, U05, U07, U09, U10, U12, U13, U17a, U18, U19, U20, U21,
U22, U23, U24a, U25.

**Note on the input file:** `.clusters/design-composition.md` contained the
entire raw extraction content duplicated verbatim from line ~6 to ~1057 and
again byte-for-byte from line ~1060 to ~2111 (confirmed via diff — zero
differences). This looks like a file-assembly artifact in the bucketing step,
not independent re-extraction. The 210 `###` blocks in the file therefore
collapse to 105 genuinely distinct raw blocks before any real cross-unit
deduplication begins; the "raw extractions" count above and the per-concept
`sources` below are drawn from that de-duplicated set of 105.

---

### DES-001 — Three-layer design system (content_components / css_themes / style_collections)
- **status:** deployed
- **status-evidence:** 002(4) opening section describes it as live; the early monolithic version is explicitly marked superseded internally once css_themes was split (025).
- **what:** The platform's design system has always had three independently-varying layers: Layer 1 self-contained HTML components (inline style, CSS variables with fallbacks, never hardcoded brand colours); Layer 2 the CSS theme; Layer 3 `style_collections` bundling header/footer components + theme + palette/typography, referenced via `sites.style_collection_id`. Layer 2 evolved: originally `css_themes` was one monolithic row (one full stylesheet per theme); migration 025 split it into `palettes`/`layouts`/`typography_sets` FK-composed rows, with `css_themes.css_content` now populated only at render time from the installed composition. Layer 3 (`style_collections`) survived unchanged as the outer bundle across this evolution.
- **sources:** 002(4)#Design System Layers (U01); 003(8) contracts (U01); old_design_and_styling/FOCUS_design_and_styling.md#"1. The Design System: Three Independent Layers" (U12); 025_palette_layout_typography_migration(3).md#"Splitting css_themes" (U12)
- **relations:** DES-013 (migration 025); DES-002 (style collections bundle); DES-003 (core composition pipeline)
- **verify-later:** content_components, css_themes, style_collections schemas; confirm css_themes legacy columns actually dropped (Phase 7)

### DES-002 — Style collections (data bundle + design bridge ancestry)
- **status:** deployed
- **status-evidence:** Two generations of migration (001 initial + 030_style_collections); sites.style_collection_id FK live; doc017/012 confirm "load style collections" as a standard planner step across generations.
- **what:** A `style_collections` row bundles the components and tokens defining a site's visual identity: header/header-home/footer component ids, css_theme_id, color_palette + typography JSONB, category/industry_tags. Sites link to one collection and may override via `sites.style_overrides` without forking. It is the long-lived "bridge" layer (mix-and-match structure + appearance) that predates and survives the palette/layout/typography decomposition — original motivation was replacing inconsistent LLM-generated headers with tested, reusable templates.
- **sources:** docs/agent_docs/sql_for_components/001_style_collections.sql (U19); 002_styles_documentation.md, 003_styles_implementation.md (U19); docs017_legacy_agent_rules_images_design_keydocs/017_agent_architecture_v2.md, 019b(U21)#Design-System-Layers
- **relations:** DES-001 (three-layer system); DES-004 (component-based headers); DES-013 (migration 025)
- **verify-later:** style_collections rows; GetStyleCollectionForSite; assignment logic in EnsureSiteRecordAction

### DES-003 — Composition pipeline: direction → composition → execution (site-design-planner + webdesign-agent)
- **status:** deployed
- **status-evidence:** 002(4)/027 describe this as the deployed reorder ("Applied"); independently confirmed live end-to-end on gamedesign.uk (2026-04-20, "first successful composition run") and idea.uk (2026-06-20 investigation "confirmed against agent_definitions and the deployed render code").
- **what:** Design is deliberately a two/three-stage pipeline, not one agent. (1) domain-research-classifier writes `design_intent` (structured palette/typography reference_values + style_direction). (2) `site-design-planner` — deterministic, no LLM, triggered by a `needs_composition` work item — resolves layout (weighted scheme-aware tag match), typography (match-or-insert), and a site-specific palette via signal cascades through three resolver actions (`validate_composition_inputs` → `resolve_composition_layout/typography/palette`), then `install_site_composition` atomically writes css_themes+style_collections+`sites.style_collection_id`+a `resolved_composition` spec in one transaction — renders nothing. (3) `webdesign-agent` (`needs_design`, `depends_on` composition) produces an LLM design overlay and renders the layout template over the installed base per a fixed merge-authority rule (LLM wins core palette slots + typography; composition wins layout/structure tokens/specialised slots) — the sole writer/deployer of styles.css. This replaced an earlier conflated "fork+install" step that produced the "first-render-with-wrong-layout" bug (site rendered once, wrong, before composition existed).
- **sources:** 002(4)#Composition, 027 full (U01); HANDOFF_2026-04-18/04-20_design_and_styling…md (U02); idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (U04); HANDOFF_2026-04-19…update4(3).md#"4. Work Plan — Deliverable 4" (U12)
- **relations:** DES-005 (resolved_composition pointer spec); DES-006 (Choice B scope); DES-021 (mandatory-overlay bug); DES-033 (webdesign install/render ordering bug, the predecessor failure this fixed); DES-046 (Design/composition flow gaps A-B-C)
- **verify-later:** fork_theme_composition.go resolvers; install_site_composition; resolve_composition_*.go; needs_composition/needs_design ordering in agent_definitions

### DES-004 — Component-based headers replacing LLM-generated chrome
- **status:** deployed
- **status-evidence:** Plan docs (002/003 styles docs) lay out the founding rule; migration 012 executes population and SQL-side rendering of site_components for header/footer/head.
- **what:** Founding decision that page chrome (header/footer/head) is never LLM-generated per page: tested templates render with a site-derived context (logo from domain, nav from pages/nav tables, colours from collection+overrides) and are injected at assembly. Predates and motivated the style_collections bundle; benefits cited were consistency, instant DB-side updates, A/B-able collections.
- **sources:** docs/agent_docs/sql_for_components/002_styles_documentation.md; 003_styles_implementation.md; docs/agent_docs/sql_for_tables/012_site_components.sql (all U19)
- **relations:** DES-002 (style collections); DES-035 (chrome linkage tangle — where this founding rule later broke down)
- **verify-later:** RenderHeaderForSite / render_site_components action

### DES-005 — resolved_composition pointer spec + install_site_composition semantics
- **status:** deployed
- **status-evidence:** Verified live on idea.uk twice (dark install, then re-resolve, 2026-06); quote: "resolved_composition is a *pointer* — it carries palette_id/name/source, not the colour values."
> **⚠ CORRECTED 2026-08-12 — the "errors rather than overwrites" half is SUPERSEDED and the workaround it implies is UNSAFE.**
> `bugs_open/113`'s fix added an **`allow_reinstall`** step-config flag (default **false**); with it the swap happens
> **inside the action's existing transaction**. Verified live in the running chassis v1.0.1289 (`allow_reinstall` = 6
> occurrences, `replaced_existing` = 1). **Do NOT follow this entry's "clear it manually" recommendation:** nulling
> `style_collection_id` leaves the site uncomposed until the re-resolve lands, and anything rendering in that window
> hits the loader's emergency fallback (`render_css_composition_loader.go:144-158`) and can deploy a
> `standard-brochure` stylesheet over a live site — see `install_site_composition_reinstall_test.go`.
> **Cost of the stale line:** it was repeated verbatim into a design-pass handoff on 2026-08-12 as the recommended
> route, and would have been followed. Caught by the LANDMINES entry *"a concept-register STATUS line is a snapshot
> that outlives its truth"*, whose check is to grep the cited source instead of trusting the status.

- **what:** The composition install contract: a `css_themes` row is created with all three FKs but empty `css_content` (~~webdesign-agent fills it at render~~ — **⚠ CORRECTED 2026-08-21: that fill did NOT exist, from this entry's creation until migration 543.** `render_css_from_spec` returned only `{result, type}` and never wrote `css_themes.css_content`; the only `css_content` it read was `css_snippets.css_content`, a different table. **The empty row was therefore permanent on any site whose stylesheet the design agent owned**, and `css-patch-agent` — which deploys that row wholesale over `assets/css/styles.css` — turned it into nine clobbered live sites across three waves (`bugs_open/198`). This clause is now TRUE: migration 543 adds `persist_css_to_theme` between `generate_css` and `deploy_css`, writing the rendered stylesheet into the row byte-for-byte, guarded on size / `origin <> 'seed'` / exactly-one-linking-site / no-change. **The cost of the wrong version is the point: this sentence is exactly what would reassure a reader that a per-site restore will be maintained, and for months it was not.**); `style_collections` points at the theme; `sites.style_collection_id` is set only if NULL — install **errors rather than overwrites** an existing composition ("re-resolve not supported; clear it manually"). The `resolved_composition` spec aspect is a lineage/decision record (`lineage.{palette_source, typography_source, layout_source}`), not the CSS itself; the old spec is superseded and a new one inserted on re-resolve. Renderer resolution is strict: missing/NULL composition parts hard-error rather than silently default ("migration gaps are audit events, not silent fallbacks"), with a loud emergency fallback to standard-brochure as the last resort.
- **sources:** idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (install + loader sections); idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Stage A) — both U04
- **relations:** DES-003 (composition pipeline); ~~DES-047 (composition re-resolve procedure)~~ **→ CORRECTED 2026-08-12: DES-047 is *Computed-styles extraction*, status aspirational, 0 repo-wide hits. The manual procedure is DES-049 — and DES-049 is itself now superseded, see the status note above.**; DES-032 (renderer theme-resolution cascade / emergency fallback)
- **verify-later:** install_site_composition_action.go; render_css_composition_loader.go

### DES-006 — site-design-planner scope: "Choice B" (composition-only) and its declared spec aspects
- **status:** deployed
- **status-evidence:** "Choice B adopted. The agent's exclusive responsibility is composition resolution... It does NOT write navigation or layout specs" (2026-04-19); doc 103 "Deliverable 2" documents the three candidate spec aspects with per-reader validation, later narrowed in practice.
- **what:** site-design-planner was scoped to write exactly one spec aspect, `resolved_composition` (palette_id/layout_id/typography_set_id + lineage + reasoning), justified by "slim strict responsibilities" — `navigation` and `layout` spec-aspect ownership (nav architecture/CTA/mobile pattern; page-level layout/header-footer style) was deferred to future specialist agents even though doc 103 defined their shapes and readers (populate_nav_tables/InjectHeader/GetNavItems for navigation; AssembleMultipageSiteAction for layout).
- **sources:** old_design_and_styling/HANDOFF_2026-04-19…update4(3).md#"3. Scope Refinement" (U12); 103_site_design_planner.sql (U18)
- **relations:** DES-003 (composition pipeline); DES-054 (site-design-planner agent structure×identity×effects — the rejected broader alternative)
- **verify-later:** agent_definitions row for site-design-planner — confirm workflow only writes resolved_composition

### DES-007 — Superseded design-agent family split (brand-designer / layout-architect / style-generator) — replaced by composition/execution
- **status:** superseded
- **status-evidence:** "The earlier 'one agent generates brand + CSS' shape is superseded by the composition/execution split above" (002_system_architecture(4), live doc, line 596); doc017/019b (an earlier archive generation) already lists the same three names as "Future split"/"Planned" that never shipped under those names; "There's no rush on this split."
- **what:** Across at least two archive generations, the plan was to decompose the monolithic `webdesign-agent` (brand analysis + colour/typography/spacing + CSS generation in one agent) into `brand-designer` (rarely-changing brand_spec), `style-generator` (CSS with theme-library search-and-adapt before generating fresh), and `layout-architect`/`nav-layout-agent` (per-page-type layout definitions). None of these three agents were ever built under those names; the live architecture instead replaced the monolith with the Composition/Execution split — `site-design-planner` (deterministic composition) + `webdesign-agent` (narrowed to render/commit styles.css, "the only writer of styles.css") — with a finer split explicitly deferred until search-and-adapt clearly beats render-from-composition.
- **sources:** old/older1/002c_system_architecture_v3.md#"Design Agent Family" (U12); docs017_legacy…/019b_agent_architecture_v5…md#"3-Design-Agent-Family", 003_design.md (U21); 002_system_architecture(3+4).md (U01)
- **relations:** DES-003 (composition/execution split that replaced it); DES-054 (site-design-planner structure×identity×effects, a different abandoned proposal); DES-008 (brand designer agent, an even earlier precursor)
- **verify-later:** confirm no brand-designer/layout-architect/style-generator agent_definitions rows exist

### DES-008 — Brand designer agent (theme selection) — earliest design decision point (superseded)
- **status:** superseded
- **status-evidence:** Agent SQL + mvp-site-builder workflow insertion (spawn/call_brand_designer feeding brand_theme to the architect); superseded first by content-creator's theme recommendation + semantic tag matching, then by the design-composition system.
- **what:** The very first brand/design decision point in the pipeline's history: an LLM agent analysing domain + objective and picking a CSS theme from a small named library (boxing, bakery, tech, professional-dark, default) with reasoning. Direct ancestor of the later site-design-planner / palette resolution system.
- **sources:** docs004_website_capture_project/website_analysis/README.018.brand_designer_agent.md (U20)
- **relations:** DES-009 (semantic CSS theme system, the next generation); DES-003 (eventual successor architecture)
- **verify-later:** brand-designer agent_definitions row

### DES-009 — Semantic CSS theme and snippet system (theme_tags, css_themes, css_snippets, js_snippets) — superseded
- **status:** partial
- **status-evidence:** Full DDL + seed data across two iterations (text[] then jsonb tags); the design-composition palette/typography system is the taxonomy-named successor.
- **stage2-verified (2026-07-14):** superseded → partial — theme_tags: 0 hits in .go — genuinely gone. But css_snippets/js_snippets NOT superseded: registry.go:734 registers render_js_snippets_for_site (wired), render_js_snippets_for_site_action.go:153 queries js_snippets, render_css_from_spec_action.go:481 queries css_snippets, invoked by 031_webdesign_agent.sql/113_site_a...
- **what:** A semantic tagging vocabulary (mood/style/industry/audience/functional/colour, with related_tags pairing) applied to css_themes (complete `:root` CSS-variable palettes: calm-minimal, bold-conversion, warm-friendly, dark-modern, premium-elegant...), css_snippets (hover/animation/effect/pattern/utility fragments), and js_snippets (nav, scroll, accordion, clipboard, form interaction fragments with trigger metadata). Content-creator recommended theme+tags; the assembler matched snippets by tag. All-CSS-variable theming here is the direct ancestor of the platform's current CSS-variable contract.
- **sources:** docs004_website_capture_project/006semantic_themes/README.020/021 (U20); 007different_types_of_site/027_css_js_schema.sql
- **relations:** DES-008 (brand designer agent, predecessor); DES-001 (three-layer system, successor taxonomy)
- **verify-later:** css_themes/css_snippets/js_snippets/theme_tags tables today

### DES-010 — Site-design-planner agent (structure × identity × effects) — abandoned proposal
- **status:** abandoned
- **status-evidence:** WORK_PLAN_v2 Phase 3 "Site-Design-Planner Agent (not started)" — all sub-items 3a–3g "Not started"; explicitly superseded by the live 027_design_and_site_planner_v2.md architecture.
- **what:** An earlier, more ambitious proposal (Option B) for a dedicated site-design-planner decomposing site design into structure × identity × effects, owning navigation/layout spec schemas outright and driving header/footer selection and hero/nav merging, plus Phase 4 requirement-driven component generation when the library had no match. Never built as specified; the agent that eventually shipped under the same name has the much narrower "Choice B" (composition-only) scope instead.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md#phase-3, #phase-4 (U24a)
- **relations:** DES-006 (the actual, narrower scope that shipped); DES-041 (visual identity library, same phase-4 family of unbuilt ideas)
- **verify-later:** agent_definitions for site-design-planner (confirm this broader shape is absent); 027 live doc

### DES-011 — chief-strategist (build-plan LLM) + component placement dedup rules — superseded
- **status:** superseded
- **status-evidence:** 040 still upgrades its model (haiku→sonnet) but the work-item pipeline planner (build-site-planner, 053) owns planning thereafter; 019 patch injects "COMPONENT PLACEMENT RULES" into its prompt.
- **what:** The v1/v2 planning agent that produced sections/component_details build plans before build-site-planner existed. Its lasting contribution, carried forward, is an anti-repetition rule-set: testimonials/team-grid/faq/contact-form appear on ONE page only, hero variants differ per page, no duplicated services content, merge similar pages.
- **sources:** 019_chief_strategist.sql; sql_for_agents_v1/019, v2/019; 040_optimise_which_llms.sql (all U18)
- **relations:** DES-003 (site-planner/build-site-planner inherit the planning role)
- **verify-later:** is chief-strategist still active or deleted

### DES-012 — Design pipeline guiding principles (mottos)
- **status:** unknown
- **status-evidence:** "Principles Restated" section repeated verbatim across multiple 2026-04-19 handoffs, sourced from 007_adoption_pipeline_v2.md and a FOCUS work-plan doc.
- **what:** A shared decision-shorthand invoked repeatedly to settle scope questions across the design-composition work: "Every build conceptually an adoption," "Design reference is history, design intent is direction," "Adoption is a starting point, not a ceiling," "LLM for reasoning, Go for extraction," "Handlers are self-contained," "Slim strict responsibilities."
- **sources:** old_design_and_styling/HANDOFF_2026-04-19…update4(3).md#"7. Principles Restated" (U12)
- **relations:** DES-006 (Choice B scope, an application of "slim strict responsibilities"); DES-014 (design_reference/design_intent split, source of "history vs direction")
- **verify-later:** none — a documentation/culture artifact, not directly code-verifiable

### DES-013 — Composable theme migration 025 (palettes / layouts / typography_sets split from css_themes)
- **status:** partial
- **status-evidence:** "Phases 1–3 (data model, layouts, seeding) are deployed and verified. Phases 4–5 (renderer cutover, fork action rewrite) were deployed but not end-to-end verified" (2026-04-18), subsequently exercised live in later cascades (idea.uk, 2026-06); legacy css_themes columns retained "until Phase 7"; Phase 4.5 coupling (buildSectionDefaults) still present as of 036.
- **what:** The foundational data-model migration: `css_themes.css_template` conflated palette, typography, and layout concerns in one row behind a silent standard-brochure fallback (one layout, 14 palette skins). Split into `palettes` (colours JSONB, open slot map), `layouts` (Go CSS template + structure_tokens + default header/footer FKs + scheme), `typography_sets` (fonts+scale) — each with an origin/needs_review/fork lineage model; `css_themes`/`style_collections` become FK-composed pointers. Renderer cutover was to a single JOIN loader + FuncMap (`{{palette}}`/`{{typo}}`/`{{token}}`) with hard error on NULL FKs; direct cutover, no shadow mode. Also created ~10 new library layout components (header-with-categories, header-docs, directory-listing, product-grid, etc.).
- **sources:** 025_palette_layout_typography_migration(3).md full; 036 §3 (U01); HANDOFF_2026-04-18_design_and_styling…md#2 (U02); docs/agent_docs/sql_for_tables/038_style_collections.sql (U19)
- **relations:** DES-001 (three-layer system); DES-014 (Layout archetype library); DES-020 (Palette merge rule); DES-025 (Scheme-aware layout matcher)
- **verify-later:** legacy columns still read anywhere; Phase 4.5/7 progress; layouts/palettes/typography_sets row counts; render_css_composition_loader.go

### DES-014 — Layout archetype library (15/17/18 named layouts) — overview
- **status:** deployed
- **status-evidence:** "Phase 1 is next: designing and writing the 15 layout CSS templates" → "Phase 1 — Layouts seeded (15 rows in layouts table)... deployed"; two further layouts (tool-portal-dark, social-lobby) and a light variant (tool-portal-light) added later, bringing the live count to 17–18 by mid-2026.
- **what:** Taxonomy of named structural/visual archetypes (brochure-formal, portfolio-kinetic, utility-tool, media-grid, docs-sidebar, etc.), each with character/structural-trait descriptions, default header/footer/typography, and legacy-theme mappings — the target library for migration 025's `layouts` table. Individual archetypes are catalogued as their own concepts below (DES-015 through DES-031-ish); this entry is the umbrella claim about the library as a whole.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"4. The 15 Layouts" (U12); 025_palette_layout_typography_migration(3).md
- **relations:** DES-013 (migration 025); DES-025 (scheme-aware matcher, the selection mechanism); every DES-Layout-* entry below
- **verify-later:** `layouts` table row count in DB today

### DES-015 — Layout: brochure-formal
- **status:** deployed
- **status-evidence:** "Phase 1 of 025_palette_layout_typography_migration"; satisfies all 5 numbered CONTRACT CHECKS.
- **what:** Structured, understated, CTA-driven brochure layout with corporate restraint. Mapped to themes `default`, `standard-brochure`, `professional-dark`. Suits consultancies, law, finance, B2B. Serves as the canonical reference implementation of all 5 layout contract checks and as the de facto fallback layout when tag resolution fails.
- **sources:** layouts/layout_01_brochure-formal.sql#L1-37,L50-69 (U13)
- **relations:** DES-014; DES-057 (Colour Inheritance Model, reference implementation); DES-025 (scheme matcher fallback target)
- **verify-later:** `layouts` row name='brochure-formal'; confirm still the de facto fallback

### DES-016 — Layout: brochure-bold
- **status:** deployed
- **status-evidence:** "Phase 3 of 025_palette_layout_typography_migration."
- **what:** High-energy conversion variant of brochure-formal — tall hero, gradient accents, display-bold typography, strong CTAs. Suits tech startups, SaaS, fitness brands.
- **sources:** layouts/layout_02_brochure-bold.sql#L1-30,L43-65 (U13)
- **relations:** DES-015 (brochure-formal)
- **verify-later:** `layouts` row name='brochure-bold'

### DES-017 — Layout: portfolio-kinetic
- **status:** deployed
- **status-evidence:** Header explicit "STRUCTURAL DIVERGENCE from brochure-* layouts" list; "Mapped themes: none currently."
- **what:** Asymmetric, motion-forward, display-type-led layout for creative-studio energy — animated underline text-links instead of hero/CTA buttons, 40/60 asymmetric columns, dense-packed work showcase, narrower 1140px container. Suits design studios, creative agencies, photography portfolios.
- **sources:** layouts/layout_03_portfolio-kinetic.sql#L1-33,L46-66 (U13)
- **relations:** DES-015 (contrast case)
- **verify-later:** `layouts` row name='portfolio-kinetic'

### DES-018 — Layout: magazine-grid
- **status:** deployed
- **status-evidence:** "Mapped themes: content-modern."
- **what:** Publication-feel layout: top-level 2/3 main + 1/3 sidebar grid, article cards, featured-article variant, sidebar widgets, serif-editorial typography. Suits news, opinion, long-form blogs.
- **sources:** layouts/layout_04_magazine-grid.sql#L1-35,L37-70 (U13)
- **relations:** DES-021 (soft-editorial), DES-030 (industry-hub)
- **verify-later:** `layouts` row name='magazine-grid'

### DES-019 — Layout: utility-tool
- **status:** deployed
- **status-evidence:** "Mapped themes: none — exists for selector/adoption matching."
- **what:** Minimal-chrome layout where "the tool is the reason" — narrowest container (800px), compact header, single tool card with output region, no card-grids, larger form controls. Suits online calculators, converters, developer utilities.
- **sources:** layouts/layout_05_utility-tool.sql#L1-25,L27-59 (U13)
- **relations:** DES-027 (tool-first-landing, explicit divergence)
- **verify-later:** `layouts` row name='utility-tool'

### DES-020 — Layout: media-grid
- **status:** deployed
- **status-evidence:** "Mapped themes: none"; dark-mode-by-default palette.
- **what:** Thumbnail-dominant, continuous-scroll discovery layout — auto-fill fluid grid, optional featured/pinned item, scrollable chip filter bar, "featured row"/horizontal-scroll shelf variants, fixed aspect-ratio tokens. Suits video platforms, audio libraries, image galleries. Dark theme by default.
- **sources:** layouts/layout_06_media-grid.sql#L1-24,L26-58,L67-90 (U13)
- **relations:** DES-022 (high-energy), DES-031 (tool-portal-dark)
- **verify-later:** `layouts` row name='media-grid'

### DES-021 — Layout: docs-sidebar
- **status:** deployed
- **status-evidence:** "Default typography: mono-technical" matches typography_sets seed row's own note.
- **what:** Reference-grade documentation layout — 3-zone CSS grid (fixed sidebar nav, main reading column, collapsing table-of-contents). Code blocks get accent-border + copy-button; admonitions use `.callout` variants. Suits developer docs, API references, knowledge bases.
- **sources:** layouts/layout_07_docs-sidebar.sql#L1-25,L27-58 (U13)
- **relations:** DES-024 (typography_sets, mono-technical); DES-031 (tool-portal-dark)
- **verify-later:** `layouts` row name='docs-sidebar'

### DES-022 — Layout: soft-editorial
- **status:** deployed
- **status-evidence:** "Mapped themes: bakery, warm-friendly, calm-minimal, soft-editorial" — the only numbered layout with 4 named theme mappings.
- **what:** Warm, reading-first, organic layout — tinted background, pill-shaped buttons, barely-there card borders, serif display headings, transparent floating header, 1.75 line-height. Suits wellness blogs, lifestyle sites, personal essays, bakeries.
- **sources:** layouts/layout_08_soft-editorial.sql#L1-23,L25-57 (U13)
- **relations:** DES-018 (magazine-grid), DES-030 (industry-hub)
- **verify-later:** `layouts` row name='soft-editorial'

### DES-023 — Layout: technical-precise
- **status:** deployed
- **status-evidence:** "Mapped themes: premium-elegant (with serif override), modern-engineering-clean."
- **what:** "Engineered" layout — glass-effect header (backdrop-filter blur) as its signature moment, tight border-radius, bordered/low-shadow cards, flat solid CTAs, light (not dark) footer contrasted against brochure-*'s dark footers. Suits SaaS platforms, infrastructure products, engineering consultancies.
- **sources:** layouts/layout_09_technical-precise.sql#L1-25,L27-58 (U13)
- **relations:** DES-015 (footer contrast case)
- **verify-later:** `layouts` row name='technical-precise'

### DES-024 — Layout: high-energy
- **status:** deployed
- **status-evidence:** "Mapped themes: boxing" (narrowest mapping of all 15).
- **what:** Aggressive, kinetic layout — uppercase headings, 80vh dark hero, diagonal clip-path section separators, zero border-radius, hard offset shadows, numeral-prefixed feature cards. Suits boxing gyms, combat sports, fitness events. Uses display-bold typography.
- **sources:** layouts/layout_10_high-energy.sql#L1-20,L22-53 (U13)
- **relations:** DES-020 (media-grid)
- **verify-later:** `layouts` row name='high-energy'

### DES-025 — Layout: comparison-aggregator
- **status:** deployed
- **status-evidence:** Header distinguishes itself from 3 sibling commerce-adjacent layouts by its defining primitive `.result-card`.
- **what:** Search-first, data-dense, trust-oriented layout — hero IS a search input, sticky filter bar, dense horizontal result-card rows, regulatory info banners, heavy disclaimer footer. First of four deliberately-differentiated "commerce-adjacent" layouts. Suits price/insurance/broadband comparison, trade directories.
- **sources:** layouts/layout_11_comparison-aggregator.sql#L1-24,L26-60 (U13)
- **relations:** DES-026 (affiliate-hub), DES-028 (ecommerce-storefront), DES-030 (industry-hub)
- **verify-later:** `layouts` row name='comparison-aggregator'

### DES-026 — Layout: affiliate-hub
- **status:** deployed
- **status-evidence:** Header's explicit divergence table against comparison-aggregator and ecommerce-storefront.
- **what:** Product-review/buyer-guide layout — persistent disclosure strip, vertical product "picks" cards, pros/cons review blocks, horizontally-scrolling comparison tables, optional sticky "Top Picks" sidebar. Suits product review sites, "best X for Y" guides, deal aggregators.
- **sources:** layouts/layout_12_affiliate-hub.sql#L1-21,L23-56 (U13)
- **relations:** DES-025, DES-028, DES-030
- **verify-later:** `layouts` row name='affiliate-hub'

### DES-027 — Layout: ecommerce-storefront
- **status:** deployed
- **status-evidence:** Header's divergence note vs affiliate-hub (cover-fit lifestyle photography vs contain-fit product-on-white).
- **what:** Retail-clean, product-forward storefront — promo hero, image-overlay category tiles, product grid, add-to-cart CTAs, strike-through sale pricing, CSS-only mini-cart dropdown structure, trust-bar strip. Suits independent shops, small-catalogue retailers.
- **sources:** layouts/layout_13_ecommerce-storefront.sql#L1-24,L26-60,L94-97 (U13)
- **relations:** DES-026, DES-025
- **verify-later:** `layouts` row name='ecommerce-storefront'

### DES-028 — Layout: tool-first-landing
- **status:** deployed
- **status-evidence:** Header's explicit divergence from utility-tool (full-container vs 800px narrow column).
- **what:** Full-container (up to 1400px) tool-dominated landing page where "the tool IS the page" — defining primitive `.split-pane` (50/50 default), dark-mode-friendly, optional tabbed interface. The "loud" counterpart to utility-tool's contained/quiet version. Suits calculators, API playgrounds, demo tools.
- **sources:** layouts/layout_14_tool-first-landing.sql#L1-22,L24-56 (U13)
- **relations:** DES-019 (utility-tool), DES-031 (tool-portal-dark)
- **verify-later:** `layouts` row name='tool-first-landing'

### DES-029 — Layout: industry-hub
- **status:** deployed
- **status-evidence:** Header's 4-way divergence table naming this the only non-commercial member of the "commerce-adjacent" family.
- **what:** Vertical information-authority layout — "About this site" independence-claim banner, `.directory-card`/`.guide-card`/`.news-card`/`.glossary-list` primitives, ordered directory→guides→news→reference, serif-editorial typography for "authority without being corporate." Suits regulatory information hubs, industry explainer sites.
- **sources:** layouts/layout_15_industry-hub.sql#L1-28,L30-61 (U13)
- **relations:** DES-025, DES-026, DES-027, DES-018, DES-022
- **verify-later:** `layouts` row name='industry-hub'

### DES-030 — Layout: tool-portal-dark
- **status:** partial
- **status-evidence:** Seeded by migration "007_seed_layouts_tool_portal_and_social_lobby.sql," explicitly framed as necessary-but-not-sufficient; `needs_review` column present on the INSERT.
- **what:** Dark developer-utility portal layout supporting three page shapes in one template — portal/index, tool pages, article/guide pages (narrow reading column). Dark-mode-first, flat technical aesthetic. Built specifically to close the layout-library gap that caused gamesdesign.co.uk to fall back to brochure-formal, and later selected end-to-end for robot-hands.com via the same class of gap (B7, imagery workstream).
- **sources:** layouts/layout_16_17_vonc_gamesdesign.sql#L1-38,L55-145,L71-94 (U13); PLAN_imagery_best_in_class.md#B7 (U10)
- **relations:** DES-036 (layout-resolution-by-tags gap); DES-031 (social-lobby); DES-048 (No runtime re-compose path — the B7 fix that later re-selected this layout)
- **verify-later:** `layouts` row name='tool-portal-dark', `needs_review` flag value

### DES-031 — Layout: social-lobby
- **status:** partial
- **status-evidence:** Same migration-007 framing as tool-portal-dark; `needs_review` column present.
- **what:** Light, colour-forward social-platform layout built around a room/lobby metaphor. Primary UI unit is the "provocation card"; Arena (competitive) and Stage (creative) rooms differentiated via dedicated palette slots (`arena`, `stage`) rather than component variants. Four page shapes: lobby/homepage, room/topic index, provocation detail, archetype/profile. Reaction-colour slots (`reaction_positive`/`reaction_negative`/`reaction_meta`) are a distinctive palette extension. Named target: vonc.com.
- **sources:** layouts/layout_16_17_vonc_gamesdesign.sql#L21-23,L713-757,L759-810 (U13)
- **relations:** DES-036, DES-030
- **verify-later:** `layouts` row name='social-lobby'; live check against vonc.com

### DES-032 — Renderer theme-resolution cascade and the emergency fallback
- **status:** deployed
- **status-evidence:** 027 §4: theme_name literal cleared "as the cutover moment"; emergency fallback + logger.Error monitoring rule is live doctrine.
- **what:** `render_css_from_spec` resolves theme by `config.theme_id` → `config.theme_name` → `sites.style_collection_id` join (the production path); an all-miss falls to standard-brochure WITH a `logger.Error` — any emergency-fallback log line is treated as a pipeline bug, not a normal path. `resolveThemeIDFromSiteContext` never hard-errors itself, only warns with a distinguishing reason.
- **sources:** 027#Renderer Changes (U01)
- **relations:** DES-003 (composition pipeline); DES-005 (install semantics, strict hard-error philosophy)
- **verify-later:** emergency fallback frequency in logs

### DES-033 — webdesign-agent install/render ordering bug ("first render wrong layout")
- **status:** partial
- **status-evidence:** "9.1 Known ordering issue in webdesign-agent... This is the exact 'first render wrong layout' bug site-design-planner was built to eliminate" (2026-04-19 handoff); deferred fix (reorder install_theme before generate_css) was superseded same-day by outright removal (DES-034), which addressed it a different way.
- **what:** Before site-design-planner existed, webdesign-agent ran `generate_css → deploy_css → ... → install_theme`, so any site without a pre-installed composition hit the emergency fallback and committed it to git before the correct composition was installed a step later — producing two commits, the first knowingly wrong.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19…update4(3).md#"9.1 Known ordering issue" (U12)
- **relations:** DES-003 (the eventual real fix — composition-before-render architecture); DES-034 (the install_theme removal that closed this specific step)
- **verify-later:** webdesign-agent workflow step order today

### DES-034 — Phased belt-and-braces removal plan for webdesign-agent install_theme (abandoned same-day)
- **status:** superseded
- **status-evidence:** v1 doc's changelog (2026-04-19) states the belt-and-braces step "remains... pending the two-phase removal plan"; the live v2 doc's changelog the same day states "Merge applied" (direct removal).
- **what:** `026_design_and_site_planner_v1.md` proposed a cautious two-phase removal of webdesign-agent's defensive `install_theme`/`check_should_install` steps (diagnostic no-op first, delete only after a week of zero firings). Abandoned within hours: the live v2 doc shows both steps deleted outright the same day, routing rewired directly to `generate_css`, relying instead on the renderer's emergency-fallback logging (DES-032) as the sole safety net.
- **sources:** old/older1/026_design_and_site_planner_v1.md#"6..." (U12); docs024_key_docs_latest/027_design_and_site_planner_v2.md#"6...(Applied)", #"12. Change Log"
- **relations:** DES-033 (the bug this step existed to guard against); DES-032 (the safety net that replaced it)
- **verify-later:** confirm install_theme/check_should_install are absent from webdesign-agent's agent_definitions

### DES-035 — webdesign-agent post-merge loop bug and generate_css stuck mystery
- **status:** unknown
- **status-evidence:** "This is a loop bug in my migration... Fix proposal (NOT YET APPLIED)" (2026-04-20); "Even with the loop fixed, we STILL don't know why generate_css didn't execute"; a later cascade (04-23) "proceeded through generate_css and deploy_css to check_should_fork," suggesting recovery but without confirmed root cause.
- **what:** Migration 010 left every non-fork path out of `deploy_css` looping back to `generate_css` (`update_site.next_step` and `check_update_db.else_step` should have pointed at `check_should_fork` instead). Separately and possibly unrelated, one production run sat at `generate_css` (a deterministic action) producing no log line and no heartbeat, with evidence lost to pod rotation — an instrumentation runbook was written for reproduction but the mystery was never confirmed closed.
- **sources:** HANDOFF_2026-04-20_composition_deployed_design_stuck.md#A (U02)
- **relations:** silent-completion failure mode; consumer-group race (candidate explanation, outside this cluster)
- **verify-later:** current webdesign-agent next_step wiring; whether the loop-fix SQL was applied

### DES-036 — Layout-resolution-by-tags gap (classifier not emitting industry_tags) and its migration-008 fix
- **status:** deployed
- **status-evidence:** Root-cause note in layout_16_17 seed header (migration 007): "the classifier doesn't currently emit those two fields... Neither migration alone is sufficient"; migration 008 (dynamic taxonomy, industry_tags array from classifier, `read_layout_taxonomy` action) validated end-to-end 2026-04-23 with tool-portal-dark selected via library_match.
- **what:** The site-design-planner's original layout picker (`resolveLayoutByTags`) intersected a site's classification tags against each layout row's `industry_tags`; the classifier stored only `industry`/`sub_industry` strings, not a tags array, so `tagSet` was always empty and every site fell back to `brochure-formal` regardless of fit (exactly what happened to gamesdesign.co.uk) — and `style_collections.industry_tags` was in turn written empty, breaking future library matching too. Fixed by two coordinated migrations: seeding missing layouts (007) plus migration 008 making the classifier emit a real `industry_tags` array against a dynamic taxonomy read live from the `layouts` table.
- **sources:** layouts/layout_16_17_vonc_gamesdesign.sql#L1-38 (U13); HANDOFF_2026-04-20…md#B, HANDOFF_2026-04-23(1).md#deployed/#validated (U02)
- **relations:** DES-030 (tool-portal-dark, the layout this gap starved); DES-025... (all 15 numbered layouts, as matching candidates); DES-037 (scheme-aware matcher, the next-generation successor)
- **verify-later:** readClassificationFromContext in resolve_composition_helpers.go; classifier output shape post-008

### DES-037 — Scheme-aware weighted layout matcher + needs_new_layout_candidate HITL signal
- **status:** deployed
- **status-evidence:** Matcher code "LIVE — merged... built into the chassis image and site-design-planner rolled. (Confirmed live 2026-06-25.)"; migration applied; idea.uk re-resolve proved tool-portal-light selection end-to-end the same day.
- **what:** Replaced the tags-only, scheme-blind `resolveLayoutByTags` (exact-overlap count, alphabetical ties — the matcher that put light-editorial idea.uk on tool-portal-dark) with a new Go matcher that treats the site's scheme (from `design_intent.style_direction`) as a **near-hard constraint** (a light site won't land on a dark layout while any non-dark layout fits), IDF-weights tag rarity so specific tags beat generic ones, normalises synonyms to a controlled vocabulary, and adds category/description keyword bonuses. On total mismatch it queues `needs_new_layout_candidate` (status needs_human_review, skipped by dispatch) — an honest "library is missing a layout" signal rather than a silent bad pick. > **CORRECTED 2026-09-03 (bugs_open/445):** "on total mismatch" was the whole problem — the category/description/same-scheme bonuses are added to `total` independently of tag matching, so a layout matching NONE of a site's tags still scored above zero and this signal never fired for a library reason (2 items across 63,007 work items ever, both the no-tags arm). Superseded by DES-086, which fires on weak TAG fit. < A paired migration added nullable `layouts.scheme` (light/dark/neutral) and a new `tool-portal-light` layout. Deliberate policy: no auto-layout-generation — a curated, varied library plus scheme-aware matching is the intended lever; LLM-judge/pgvector matching was considered and deferred.
- **sources:** 002(4), 027 §2, 016 §9 (U01); idea.uk/resolveLayoutByTags_weighted.go.patch.txt, migration_layouts_scheme_and_light_tool_portal.sql, HANDOFF(13).md (U04)
- **relations:** DES-036 (the gap this matcher generation replaces); DES-058 (scheme derivation, upstream input); DES-013 (migration 025, where layouts.scheme lives)
- **verify-later:** fork_theme_composition.go current resolveLayoutByTags; remaining NULL layouts.scheme rows (only 3 of 18 curated as of one snapshot)

### DES-038 — Theme/layout library growth: fork-with-review gate + design-asset lineage columns
- **status:** partial
- **status-evidence:** 003(8) forking rules deployed; 038 Part 2(b): lineage columns "required by the Phase 5 fork_theme_from_site action... nothing needs review (fork action hasn't shipped yet)" as of that snapshot; the earlier auto "search→reuse→generate→store" loop was explicitly dropped in favour of a curated-library stance.
- **what:** Layouts are a curated shared grammar — no auto-generated bespoke layout per site. Growth happens via hand-added variants or a HITL route: `ForkThemeFromSiteAction` promotes a rendered design into `css_themes`+`style_collections` with `needs_review=true` and a `needs_theme_review` work item; selectors must exclude `needs_review` rows; rejection only affects future sites. Uniform provenance columns (`origin`, `needs_review`, `forked_from_<entity>_id`, `source_site_id`, `source_domain`, `forked_at`) were added across `palettes`/`layouts`/`typography_sets`/`css_themes`/`style_collections` to support this.
- **sources:** 002(4)#Library growth, 003(8)#CSS Theme Template Contract (U01); docs/agent_docs/sql_for_tables/038_style_collections.sql#PART2 (U19)
- **relations:** DES-013 (migration 025); DES-053 (per-site style fork chain, a concrete instance); DES-062 (fork_theme step double-creation guard)
- **verify-later:** fork_theme_from_site_action.go; needs_review filtering in selectors; any rows with origin != 'seed'

### DES-039 — Early "visual identity poles" layout taxonomy (dropped)
- **status:** superseded
- **status-evidence:** Diff-confirmed only in the earliest of four palette/layout/typography-migration drafts; the final 15-layout table uses different (hyphenated) names, though it keeps several "pole" nicknames.
- **what:** The very first migration draft described layout diversity as nine named "poles" tied to specific reference sites (Brochure/corporate, Magazine/editorial, Portfolio/kinetic "vonc", Commerce/grid, Utility/tool "thunder compute", Media/streaming "youtube", Documentation/reference, High-energy/bold "boxing", Soft/editorial). Dropped in favour of vaguer prose, then crystallised differently as the final 15-layout table (adding six layouts absent from the original nine-pole list).
- **sources:** old/older1/025_palette_layout_typography_migration.md#"2. Scope Decisions"; 025(3)#"7. The Layouts to Build" (both U12)
- **relations:** DES-013 (migration 025); DES-014 (final layout library)
- **verify-later:** final layouts table row count/names vs. the 15-layout plan

### DES-040 — Visual identity library and effects library (composable design assets) — aspirational
- **status:** aspirational
- **status-evidence:** Listed under "Phase 4 (later)" in a 2026-04-11 plan; not confirmed built.
- **what:** Longer-term plan for two accumulating libraries: a visual identity library of palettes/typography/effects searchable by purpose/audience, and an effects library treating elevation/corner radius/animation/density as composable modifiers independent of layout. Likely the precursor idea to the palettes/typography_sets/layouts table split that was actually implemented (structure_tokens is the closest realised fragment).
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Phase 4: Requirement-Driven Components (longer term)" (U12)
- **relations:** DES-013 (migration 025, the mechanism that partly realised this); DES-060 (structure_tokens JSONB convention)
- **verify-later:** whether structure_tokens/effects concepts in the live schema fulfil this idea

### DES-041 — Component-creation via HITL work-item triage — superseded
- **status:** superseded
- **status-evidence:** "migration_025_component_triage.sql — an earlier work-item-based approach that was superseded by the direct insert approach... Do not run this file."
- **what:** Earlier plan for seeding new library components via work items routed through HITL triage. Superseded by a direct SQL insert once components were designed and reviewed.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"2. What's Been Completed" (U12)
- **relations:** DES-014 (layout archetype library, the components it would have seeded)
- **verify-later:** none — historical, file explicitly marked do-not-run

### DES-042 — Palette merge rule: core slots vs specialised slots
- **status:** deployed
- **status-evidence:** "Core slots (spec wins where present)... Specialised slots (theme wins)" — stated as settled rule, later restated identically as the LLM-vs-composition merge-authority rule inside the two-stage pipeline.
- **what:** When a site composes a theme, core palette slots let the site's own spec (or, in the later two-stage pipeline, the LLM overlay) win when present; specialised slots (primary_hover, hero_title, cta_bg, etc.) always take the theme/composition's curated value. This atomic rule is referenced by nearly every later merge-authority discussion in the pipeline.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Palette merge rule" (U12)
- **relations:** DES-003 (composition pipeline, restates this rule as LLM-wins-core); DES-021 (mandatory-full-overlay bug, a violation of this rule in practice); DES-052 (analyze_design requires reference_values)
- **verify-later:** resolve_composition_palette_action.go merge logic

### DES-043 — Palette/typography resolution cascade + the dead-slot bug and fingerprint-fallback hardening
- **status:** partial
- **status-evidence:** Cascade live and proven; dead slot "CONFIRMED why (2026-06-19, from resolve_composition_palette.go)"; hardening "DELIVERED" as code 2026-06-19/20 but "READY... NOT YET APPLIED" pending an image rebuild + roll.
- **what:** The palette source cascade is design_reference → mission → `design_intent.palette.reference_values` → layout seed → archetype default (typography analogous; palettes are always site-specific, layouts a shared curated library). The dead-slot bug: cascade slot 1 read `design_reference.palette.reference_values`, a key the adoption fingerprint extractor never actually writes (it stores `suggested_mapping`/`css_variables`/`colors` instead) — so slot 1 was silently dead and adopted references never drove the composition palette. The delivered (not-yet-rolled) hardening repoints slot 1 at the fingerprint's real keys as a fallback after design_intent. Under the current LLM-wins-core merge rule, the composition palette mostly doesn't visually "paint" the final site anyway — this fix mainly repairs lineage correctness and the rare-gap fallback, not the primary colour lever (which is the LLM overlay, fed by classifier output).
- **sources:** idea.uk/UPDATE_FOCUS_design_adoption_workplan_2026-06-19(1).md#3; idea.uk/HANDOFF(13).md; idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (problem 1) — all U04
- **relations:** DES-042 (merge rule); DES-046 (Gaps A-B-C, Gap C is the sibling producer-side bug); DES-014 (design_reference/design_intent big entry)
- **verify-later:** resolve_composition_reference_helpers.go deployed or not; extractPaletteSignal/extractTypographySignal

### DES-044 — design_reference vs design_intent spec-aspect model: extraction, three-way priority, palette-lock policy
- **status:** deployed
- **status-evidence:** Independently confirmed by at least six extraction units (U01/U02/U04/U05/U12/U17a/U24a) across the archive's full time span (2026-04-12 fingerprint pipeline through 027's live "Related Specs" table); this is one of the most heavily re-derived concepts in the whole design-composition corpus, confirming it as foundational and stable.
- **what:** `design_reference` holds concrete values (hex colours, font stacks, CSS variables, spacing, a `suggested_mapping` source→our-variable-name table) extracted mechanically — no LLM — from a crawled/adopted site's rawHTML via `extract_design_fingerprint` (goquery); it is a historical, immutable record. `design_intent` holds semantic creative direction (e.g. "dark IDE aesthetic... start here"), auto-generated from design_reference by an LLM (`generate_design_intent`) at adoption time or written later by a strategist/human; it is deliberately non-prescriptive by default so the improvement loop and webdesign-agent retain creative room, though it later gained an optional structured `palette.reference_values` block for cases needing exact colour preservation (see DES-052). Together these replace an earlier single, vague, LLM-guessed `design` spec aspect that conflated historical fact with creative direction. The webdesign-agent's `analyze_design` step branches on a three-way priority: `design_intent` present → creative freedom within the described character; only `design_reference` → faithful reproduction, no invented palette; neither → generate from industry/audience/identity. A companion policy locks the palette exactly to the reproduced original until `design_intent` exists, after which the improvement loop may evolve it. Guiding mottos: "design reference is history, design intent is direction" / "every build is conceptually an adoption."
- **sources:** WM/007_adoption_pipeline_v3.md#design-fingerprint-pipeline, #design-evolution-lifecycle (U17a); old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_design_fingerprint_pipeline.md#"Key Decisions Made" (U12); old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-14_v2.md#"Webdesign-Agent Prompt (deployed)" (U12); package_module/FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md (U05, U24a); docs024_key_docs_latest/027_design_and_site_planner_v2.md#"Related Specs" (U01)
- **relations:** DES-045 (design fingerprint extraction pipeline, the mechanical step feeding design_reference); DES-042 (palette merge rule); DES-043 (dead-slot cascade bug, a later-discovered gap in this model); DES-021 (mandatory-full-overlay bug, a later violation of the "creative freedom" clause); DES-052 (analyze_design requires reference_values, the concrete fix instance); DES-012 (guiding principles, source of the mottos)
- **verify-later:** confirm `design` spec aspect is no longer written anywhere; site_specs population rate of design_reference/design_intent across adopted sites; webdesign-agent analyze_design prompt text today

### DES-045 — Design fingerprint extraction pipeline (rawHTML → design_reference)
- **status:** deployed
- **status-evidence:** Went "Not started" → "✅ Deployed, works" (2026-04-14) → "Victory: Design Fingerprint Now Correct" (2026-04-16), verified end-to-end on gamedesign.uk.
- **what:** The Go action (`extract_design_fingerprint`) that parses a crawled site's rawHTML `<style>` blocks, CSS custom properties, Google-Fonts links, and layout signals into a concrete `design_reference` spec — hex values, font stacks, a `suggested_mapping` from source variable names to the platform's own — so adoption rebuilds can reproduce the original's colours/fonts/layout instead of falling back to generic component defaults.
- **sources:** old_design_and_styling/FOCUS_design_and_styling.md#"4. The Adoption Design Gap"; FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"Victory" (U12); FOCUS_design_and_styling_adoption_problems.md (U24a)
- **relations:** DES-044 (design_reference/design_intent model); DES-046 (fpExtractCSSVars fix, a bug within this pipeline); DES-047 (computed-styles extraction, a supplementary step)
- **verify-later:** site_specs rows with aspect='design_reference' for adopted sites

### DES-046 — fpExtractCSSVars regex-based CSS variable extraction (superseded internal bug fix)
- **status:** superseded
- **status-evidence:** BEM selectors like `.btn--primary:hover` were captured as fake variables under the original approach; the replacement uses `:root` block targeting with semicolon-splitting.
- **what:** The original design-fingerprint CSS-variable extractor used one whole-stylesheet regex, producing false positives on BEM class names. Replaced with a multi-strategy extractor that isolates `:root`/body/`[data-theme]` blocks, with a fallback frequency analysis for utility-CSS sites lacking clean custom-property blocks.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"P6 — fpExtractCSSVars BEM False Positives"; FOCUS_design_and_styling_fp_extract_css_vars_integration.md (both U12)
- **relations:** DES-045 (design fingerprint extraction pipeline, the parent mechanism)
- **verify-later:** extract_design_fingerprint_action.go — confirm regex-based extractor removed

### DES-047 — Computed-styles extraction via browser JS injection
- **status:** aspirational
- **status-evidence:** "Computed styles (Phase 2) deferred... Spec written but not implemented" in one record, vs. a complete Go action + workflow SQL described in the Phase 2 doc itself — an unresolved discrepancy in the archive.
- **stage2-verified (2026-07-14):** partial → aspirational — grep -rn 'extract_computed_styles|getComputedStyle|ExtractComputedStyles' --include=*.go . → 0 hits repo-wide; the Go action described as 'fully spec'd' does not exist in the codebase, resolving the archive's ambiguity toward not-implemented.
- **what:** A supplementary fingerprint step: scrape a homepage with injected JS calling `getComputedStyle()`, write the resolved values for a Go action to parse and merge as "ground truth," overriding source-CSS guesses when the two disagree. Fully spec'd but recorded elsewhere as deferred/not implemented — status genuinely unclear from the archive.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_computed_styles_extraction_phase2.md; HANDOFF_2026-04-16_v2(1).md#"Fixes Ready But Not Deployed" (both U12)
- **relations:** DES-045 (design fingerprint extraction pipeline)
- **verify-later:** registry.go for extract_computed_styles; site-adoption-agent workflow steps

### DES-048 — No runtime re-compose path — layout change via the 025 FK-swap pattern
- **status:** partial
- **status-evidence:** "B7 COMPLETED 2026-07-10 evening — via the 025 FK-swap pattern... there is no runtime re-compose path (deliberate deferral). NEW OPEN ITEM: build a proper runtime re-compose mode."
- **what:** Changing an existing site's layout is deliberately unsupported at runtime: `install_site_composition` refuses when a style_collection already exists, and `fork_theme_from_site`'s install mode was removed 2026-04-19. The sanctioned workaround is a targeted `css_themes.layout_id` FK swap (backup + verify) followed by a webdesign-agent CSS re-render + page rerenders. Root cause of the specific case that forced this (robot-hands' brochure fallback): old-format classification lacked `industry_tags`, so the scheme-aware matcher had nothing to score, even though the layout library already held the right answer (tool-portal-dark) — itself grown from a prior instance of the exact same classification-format gap (DES-036). Illustrates that "the library learns" but classification drift can still starve it.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turns-18–22; SQL_2026-07-10_b7_layout_fix.sql/_swap.sql; PLAN_imagery_best_in_class.md#B7 (U10)
- **relations:** DES-036 (classification-tags gap, the recurring root cause); DES-030 (tool-portal-dark, the layout involved); DES-049 (composition re-resolve procedure, the fuller alternative workaround)
- **verify-later:** install_site_composition refusal; robot-hands css_themes.layout_id = tool-portal-dark

### DES-049 — Composition re-resolve procedure (gated, file-based, backup-first)
- **status:** ~~deployed~~ **SUPERSEDED 2026-08-12 by `allow_reinstall` (`bugs_open/113`), and unsafe to follow on a live site**
- **supersession-evidence:** `install_site_composition_action.go:214` reads `allow_reinstall` (default false) and swaps inside the existing transaction; live in v1.0.1289 (pod-grep: 6 occurrences). This procedure's detach-and-clear step opens the window that `render_css_composition_loader.go:144-158`'s emergency fallback fills with `standard-brochure`. Still the correct reference for the BACKUP and UNIQUENESS-CHECK steps, which the flag does not replace.
- **status-evidence:** Steps 1–6 all marked DONE with results (2026-06-22→25); "RE-RESOLVE SUCCEEDED: idea.uk now on tool-portal-light (scheme fix proven end-to-end)."
- **what:** The safe pattern for re-running composition on an already-built site, given that install refuses overwrites (DES-005): ordered SQL FILES — backup+inspect (four uniqueness checks that must all be 0), gated detach+clear (NULL `style_collection_id`; delete the site's own collection→theme→palette→typography chain only where `source_site_id` matches; supersede the old resolved_composition spec), state-check, kcat re-trigger of site-design-planner (`domain` required by `ensure_site_record`), verify. Two learned caveats are now doctrine: run SQL as files, never pasted (pasting mangled `\set`/blank lines and left an open transaction in one incident); a standalone-orchestrated planner run ends at install and emits NO `needs_design` — the styles.css render is a separate, explicit webdesign-agent orchestration. Distinct from the adoption teardown (bulk delete by source_domain), which must NOT be used on a fresh site.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (re-resolve section); reresolve_idea_uk_01_backup_and_inspect.sql (+02/02b/03/04/05 series); running_notes(63).md (all U04)
- **relations:** DES-005 (install semantics this works around); DES-037 (scheme-aware matcher, validated by this procedure); DES-050 (library-row cleanup pattern, the sibling recovery procedure for failed cascades)
- **verify-later:** bak_*_idea_20260625 tables; orchestration_states rows for the re-resolve correlations

### DES-050 — Library-row cleanup pattern for failed composition cascades
- **status:** deployed
- **status-evidence:** Executed 2026-04-23 with recorded counts (4 css_themes, 7 palettes, 4 style_collections cleared for gamesdesign) and a NOT IN guard protecting seeded layouts.
- **what:** A bad composition cascade leaves one set of library rows (css_themes/palettes/style_collections/typography_sets) per resolve attempt behind; if left in place, the matcher can pick these wrong-decision artefacts for future, unrelated sites. The recovery pattern is a reverse-FK-order delete scoped by `source_domain`. A related open item: site deletion should clean up unreferenced library rows too (FKs are currently SET NULL, leaving orphans).
- **sources:** HANDOFF_2026-04-23(1).md#cleanup, item 18 (U02)
- **relations:** DES-049 (composition re-resolve procedure, sibling recovery pattern); DES-038 (fork/library growth, the source of these rows)
- **verify-later:** any delete-site action's library handling

### DES-051 — Design/composition flow gaps A–B–C and the plan-time trigger fixes
- **status:** partial
- **status-evidence:** "UPDATE 2026-05-26: gap (A) is being deployed to production now... Gaps (B) and (C) remain open." Gap A was independently confirmed closed the same day via a live gamesdesign cascade trace ("deployed step order... So Gap A is closed on both fresh-build and adoption paths").
- **what:** Three stacked gaps behind themeless/off-palette built sites, investigated as one thread. (A) Composition/imagery were never triggered — the Phase-1 refactor of build-site-planner lost the `needs_composition`/`needs_design` emission entirely; fixed by restoring `emit_design_items`/`emit_imagery_items` (shared `imageryplan` package) as plan-time steps, `emit_design` guarded on `style_collection_id IS NULL` and `emit_imagery` priority-banded (65 index-hero, 70 site-logo, 75/80 others, 98 clamped section-scope) so imagery lands in the first deploy. (B) Planner design drift — the adopted `design`/`design_reference` aspect is never rendered into the `plan_site` prompt, and `design_intent.style_direction` is a fixed 3-value enum (professional-dark|modern-light|bold-creative) that can't express e.g. "cyberpunk terminal," forcing collapse to the nearest bucket. (C) Colour reaches the resolver only as prose `colour_mood` flattened into directives, not the structured `palette.reference_values` the composition cascade actually reads — so planned colours mostly only reach the render via the webdesign-agent overlay, never the base composition. An outstanding reuse note: extract a shared `emitInitialCompositionAndDesign` helper so `emit_design`'s insert logic can't drift from `WriteBuildItemsAction`'s.
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md; README_difference_between_work_site_orchestrator_and_build_site_planner.md (U09); HANDOFF_2026-05-26_design_imagery_triggers_and_adoption_diagnosis.md#What-deployed (U02)
- **relations:** DES-003 (composition pipeline); DES-043 (palette dead-slot bug, the sibling consumer-side version of Gap C); DES-052 (analyze_design requires reference_values, a later concrete fix for the Gap-C failure mode); DES-014 (design_reference/design_intent, aspect ownership underlying Gap B)
- **verify-later:** build-site-planner v1.0.1047 workflow (reconcile_site_plan → emit_design → emit_imagery → complete); style_direction enum; createPalette core keys

### DES-052 — `analyze_design` requires structured palette.reference_values (else the LLM invents a palette)
- **status:** deployed
- **status-evidence:** "the analyze_design LLM step INVENTED a dark core... Fix: restructured design_intent into palette.reference_values + prescriptive guidance... Re-rendered → all slots now exactly match" (leopardess site, 2026-07-10/12).
- **what:** webdesign-agent's `analyze_design` LLM step reads colours only from `design_intent.palette.reference_values`, never from a looser `color_scheme` field; without prescriptive values there ("these eight values are FIXED, output verbatim") it improvises from mood text under its documented creative-freedom licence. The same pattern was applied to typography reference_values. This is the concrete, proven fix for the general "mandatory-full-overlay" failure mode (DES-021) and Gap C (DES-051) — the leopardess `design_intent.json` is now the worked contract example.
- **sources:** docs/leopardessconsulting/specs/design_intent.json#palette; RUNNING_NOTES.md#Turn-12; HANDOFF.md#4.6 (all U25)
- **relations:** DES-042 (merge rule); DES-021 (mandatory-full-overlay bug, the problem this fixes); DES-051 (Gap C); DES-044 (design_reference/design_intent model)
- **verify-later:** webdesign-agent workflow analyze_design step; render determinism with an empty design_spec

### DES-053 — Per-site style fork chain (palette → css_theme → style_collection)
- **status:** deployed
- **status-evidence:** "Palette — forked... seed 3196d966 untouched, still dresses 3 other sites. Deployed styles.css matches the validated palette exactly" (2026-07-10/12).
- **what:** The safe pattern for restyling one site without affecting others sharing the same seed collection: clone `palettes`+`css_themes`+`style_collections` rows (reusing the seed layout/typography/header/footer), repoint `sites.style_collection_id`, and never edit the shared seed collection. Proven with the leopardess two-tone gold system (bright #C8A951 only on dark chrome at 8.56:1 contrast; bronze #836E32 for links on light backgrounds, since bright gold fails AA at 2.1:1 there). The header component had to be forked too, since `header-professional-dark` hardcodes navy with zero CSS variables across 4 sites — illustrating that a component/collection-wired fork sticks where a mere palette-row fork does not.
- **sources:** docs/leopardessconsulting/scripts/L3_fork_palette.sql; RUNNING_NOTES.md#Turn-10/12; RUNBOOK.md#O10 (all U25)
- **relations:** DES-038 (fork-with-review gate and lineage columns); DES-054 (contrast gate gap, same workstream)
- **verify-later:** style_collections/palettes rows for leopardess; fork_theme_composition.go / install_site_composition

### DES-054 — Deterministic contrast gate missing on specialised palette slots
- **status:** aspirational
- **status-evidence:** "nothing stops a fork shipping an inaccessible palette — the WCAG primitives exist but aren't called at generation/fork/install/render for specialised slots."
- **what:** `color_util.go` has correct WCAG code (`relativeLuminance`, `wcagContrastRatio`, `pickReadableOnBackground`), but it is wired only to loose section-text defaults (3.0/2.0) and forced-text-colour stripping (AA 4.5) — the specialised slots that actually leaked accessibility bugs (card_bg, header_bg, cta_bg/cta_text — white cards, navy chrome, blue CTAs) are never contrast-gated at generation, fork, install, or render time. Validation is currently done by hand. Adding the gate is described as small.
- **sources:** docs/leopardessconsulting/RUNNING_NOTES.md#Turn-10; HANDOFF.md#8; RUNBOOK.md#O10 (all U25)
- **relations:** DES-053 (per-site style fork, where the gap was discovered); DES-063 (layout CTA-pair WCAG curation, a manual version of the same check applied to the shared layout library)
- **verify-later:** color_util.go call sites; whether any generation/fork path calls wcagContrastRatio on specialised slots

### DES-055 — Three-per-row no-orphan grid rule as a content fix
- **status:** convention
- **status-evidence:** "card grids are 3-up (no orphan row), per the brief. That is a CONTENT fix, not a CSS one — the grid components are shared across 5 sites" (2026-07-10).
- **stage2-verified (2026-07-14):** deployed → convention — Three-per-row grid rule is a content-authoring convention encoded in design_intent.layout_preference prose, not a standalone code/db artifact — reclassified as convention.
- **what:** Neither a global `repeat(3,1fr)` nor a per-component `auto-fit,minmax()` avoids orphan/stretched last cards in a shared grid component; the durable fix is a content rule (card counts divisible by three), enforced because the grid CSS itself is shared across sites and untouchable per-site. Some components (e.g. case-studies-grid, hard-wired to five cards) simply cannot be made 3-up. Encoded directly into `design_intent.layout_preference`.
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#L4; L5_homepage.sql, L5_pages.sql headers; design_intent.json#layout_preference (all U25)
- **relations:** DES-002 (style collections / shared component semantics)
- **verify-later:** grid component CSS and shared usage counts

### DES-056 — Section-contrast / dark-section model: the four-generation evolution arc
- **status:** partial
- **status-evidence:** Synthesised across the full archive time span, 2026-04 through 2026-07; the underlying facts of each generation are individually well-evidenced (see the entries it references), but the arc as a whole is not "done" — the final generation's fixes were still rolling out (Go batch unshipped) at the latest snapshot captured.
- **what:** This is a dedicated lineage entry, not a single mechanism — it exists because the same underlying problem (how does a site's light/dark scheme actually reach what the browser paints?) was rediscovered and re-solved at increasing depth across at least four archive generations, each correcting the previous generation's mental model:
  1. **Foundational contract (earliest, still standing).** The Colour Inheritance Model (DES-057) established a two-tier CSS custom-property fallback — `var(--section-*, var(--color-*))` — so a "dark section" component could override just its own container's variable. The Dark Section Variable Contract (DES-058) added the rule that layouts must NOT declare `--section-*` defaults themselves; a Go renderer function (`buildSectionDefaults`) appends them post-render, keyed off palette luminance, with 5 hardcoded class names that must stay in sync on both sides (flagged as "Phase 4.5," a known temporary coupling).
  2. **Scheme selection, then the signal dropped (mid).** `deriveSchemeFromDesignIntent` derives light/dark/'' from `design_intent.style_direction` and feeds the scheme-aware layout matcher (DES-037) as a near-hard constraint — but the derived scheme itself is never recorded in `resolved_composition` and never reaches the component `RenderContext` (DES-058-adjacent finding, "Scheme derivation and drop at render" / "Scheme resolution pipeline and where the signal stops"). Light/dark variety was handled by *paired layouts* (tool-portal-light vs -dark) rather than runtime component flipping; only 3 of 18 layouts had `scheme` populated at one snapshot.
  3. **The gap reframed as structural, not just missing plumbing (later).** A dedicated investigation (REPORT_scheme_does_not_reach_components.md, 2026-06-26) found the scheme reaches styles.css `:root` correctly but never reaches the *components* that render sections/header/footer: the component library is dark-oriented by default (no light hero/CTA/footer exist), components self-style with their own class vocabulary that the layout's section rules don't match, and many hardcode dark treatments outright. `is_dark_section` (an LLM-authored bool) is loaded but never used in section selection, is unreliable, and conflates "intrinsically dark" with "should contrast the page." The corrected understanding that emerged: the scheme *does* reach components implicitly via CSS variables — components defeat it by hardcoding assumptions, so the fix is de-hardcoding existing templates, not new plumbing.
  4. **The fix thesis and its concrete execution (latest).** The "scheme-as-override" thesis (DES-059) reframed scheme as a set of variable *values* — an override layer supplied by composition/renderer and consumed by de-hardcoded components — explicitly rejecting *-light/*-dark component duplication, and separating **base site scheme** from **per-section contrast intent** (a dark hero on a light site is legitimate, intentional contrast, not a bug). This produced concrete work: a library-wide audit classifying the 37 self-declaring components into hazard-class (~18, contradict their own scheme) vs band-class (~19, coherent but block "fully light") (DES-060); a "paired-variable" completion of the *existing* --color-cta-bg/--color-cta-text-style convention rather than a restructure (DES-061); WCAG-gated CTA-pair curation across the shared layout library (DES-063); and a parallel discovery that header/footer chrome resolution has FOUR overlapping, partly-dead default stores feeding a hardcoded-dark `RenderFallbackHeader` (DES-062), plus a scheme-blind, largely bypassed component-scoring path (DES-064).
  The throughline: each generation correctly solved its own layer (CSS contract → layout selection → structural gap diagnosis → override-based fix), but each layer's fix was necessary-not-sufficient for the one below it, so the "site looks the wrong scheme" bug recurred in a different shape at every stage for over three months of archive time.
- **sources:** layouts/layout_01_brochure-formal.sql (contract origin, U13); PLAN/RUNBOOK/running_notes_scheme_to_components (U03, U07); idea.uk/REPORT_scheme_does_not_reach_components.md, HANDOFF_scheme_to_components(1).md (U04); RUNBOOK_scheme_to_components(18).md CHECK 1-4 results (U07)
- **relations:** DES-057, DES-058, DES-059, DES-060, DES-061, DES-062, DES-063, DES-064, DES-065 (every stage-specific entry below); DES-037 (scheme-aware layout matcher, generation 2's mechanism)
- **verify-later:** whether the generation-4 fixes (de-hardcoded components, light footer, update_site_defaults in build path, scheme-aware chrome fallbacks) fully landed in a later thread than the one captured here

### DES-057 — Colour Inheritance Model (two-tier `var(--section-*, var(--color-*))` fallback)
- **status:** deployed
- **status-evidence:** Every layout_NN header lists this as CONTRACT CHECK #1 and the CSS body implements it identically (e.g. layout_01 lines ~160-182).
- **what:** Element-level colour rules (headings, body text, links) resolve via a two-tier CSS custom-property fallback chain: `var(--section-*, var(--color-*))`. This lets a "dark section" override just the `--section-*` variable on its own container without any layout needing to restate rules elsewhere. Applied identically across all 17+ layout templates. Generation-1 foundation of the section-contrast arc (DES-056).
- **sources:** layouts/layout_01_brochure-formal.sql#header+L160-182; layout_02_brochure-bold.sql#header; layout_16_17_vonc_gamesdesign.sql#L832; layout_10_high-energy.sql#header (all U13)
- **relations:** DES-056 (evolution arc); DES-058 (Dark Section Variable Contract); DES-061 (template helper system)
- **verify-later:** render_css_from_spec_action.go; every layout's `:root` and base element rules

### DES-058 — Dark Section Variable Contract / buildSectionDefaults renderer behaviour
- **status:** partial
- **status-evidence:** layout_01 lines 268-289: "TEMPORARY RENDERER COUPLING: these 5 class names must stay in sync with buildSectionDefaults in render_css_from_spec_action.go... Tracked as Phase 4.5."
- **what:** Layout templates must NOT declare `--section-*` defaults on section containers themselves; a Go renderer function `buildSectionDefaults` appends `--section-*` overrides after rendering, chosen by palette luminance. Five renderer-managed surface classes (`.features-section`, `.services-section`, `.differentiators-section`, `.about-section`, `.faq-section`) are hardcoded identically on both the Go side and every layout's SQL comments and must be kept in sync by hand; hero/CTA/testimonials/contact sections are excluded as component-owned. One documented exception: a palette-declared `heading` slot emits a root-level `--section-heading`. Generation-1/1.5 of the section-contrast arc (DES-056) — later found (generation 3) to still leave components with no reliable way to know whether they're "in a dark section."
- **sources:** layout_01_brochure-formal.sql#L14-32,L268-289; layout_02_brochure-bold.sql#header; layout_16_17_vonc_gamesdesign.sql#L85-93 (all U13)
- **relations:** DES-056 (arc); DES-057 (colour inheritance model); DES-060 (hazard/band-class split, which supersedes reliance on this alone)
- **verify-later:** render_css_from_spec_action.go buildSectionDefaults; Phase 4.5 status in docs 025/026/027

### DES-059 — Scheme derivation cascade + the drop-at-render gap
- **status:** deployed
- **status-evidence:** "`deriveSchemeFromDesignIntent(style_direction, suggested_style)` returns light/dark/''... `buildResolvedCompositionSpec` records the layout/palette ids... but not the scheme value" (traced end-to-end 2026-06-30); independently reconfirmed later as "Scheme→variable pipeline verified correct; all 18 layouts carry the four chrome vars" with the same RenderContext gap (2026-07-02).
- **what:** Scheme (light/dark) is derived at composition time from `design_intent.style_direction` by substring matching, used by the layout matcher (DES-037) as a near-hard constraint, then effectively dropped: neither the CSS loader SELECT nor the component `RenderContext` reads `layouts.scheme` afterward (the column is check-constrained to light/dark/neutral). It survives only as the layout's own curated property, recoverable via `sites.style_collection_id → style_collections.css_theme_id → css_themes.layout_id → layouts.scheme`. Light/dark variety is handled by paired layouts (tool-portal-light/-dark), not runtime component flipping; only 3 of 18 active layouts had `scheme` set at one snapshot. The corrected understanding reached later: the scheme *does* reach components implicitly via CSS variables — the real defect is components hardcoding dark assumptions, not missing plumbing.
- **sources:** PLAN_scheme_to_components(1).md#Confirmed-at-code-level; running_notes_scheme_to_components(55).md#Sb/#Sf; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (U03); running_notes_scheme_to_components(22).md Sb/Sc/Sf/Sk; RUNBOOK_scheme_to_components(18).md header + CHECK 1 (U07)
- **relations:** DES-056 (arc); DES-037 (scheme-aware matcher, upstream consumer); DES-065 (the deeper structural gap this feeds into)
- **verify-later:** deriveSchemeFromDesignIntent, resolveLayoutByTags, buildResolvedCompositionSpec; layouts.scheme column + values; RenderContext struct

### DES-060 — Hazard-class vs band-class self-declarer split; is_dark_section demoted to metadata
- **status:** convention
- **status-evidence:** "the 37 self-declarers split into two classes" with named components (audit run 2026-07-02); "6 declarers have is_dark_section=f... never key styling on the LLM-authored flag."
- **stage2-verified (2026-07-14):** deployed → convention — Hazard/band-class split is an audit classification report from a specific run, not a standing code/db object to verify by grep — reclassified as convention/process.
- **what:** Library-wide diagnosis, generation-4 of the section-contrast arc: of 84 active section components, 37 self-declare `--section-*` — roughly 18 "hazard-class" (declare dark context while painting surface vars or nothing, producing white-on-light bugs) vs 19 "band-class" (paint palette bands + white text — internally coherent, but block a site from ever being "fully light"); 15 carry raw hex backgrounds. `is_dark_section` is an LLM-authored component boolean contradicted by 6 of its own declarers and consumed by nothing that actually styles — demoted to selection/imagery metadata only; styling must never key on it. This classification sized every subsequent fix batch in the arc.
- **sources:** RUNBOOK_scheme_to_components(18).md CHECK 2/3 RESULTS; running_notes(22).md Sn, Sh (both U07)
- **relations:** DES-056 (arc); DES-058 (the contract this audit stress-tests); DES-061 (paired-variable direction, the fix)
- **verify-later:** content_components is_dark_section values vs template styling; remaining unconverted declarers (~10 hazard + ~17 band)

### DES-061 — Paired-variable design direction (curated bg+text pairs, completion of the existing standard)
- **status:** partial
- **status-evidence:** "pair convention is ALREADY the standard — 18/18 --color-primary-text, 17/18 --color-cta-text" (audit); W1–W3e template work executed 2026-07-02/03 but "inert until re-render/rebuild"; the Go batch (scheme-aware fallbacks, creator-prompt update, fixer re-aim) was unshipped as of the latest snapshot.
- **what:** The fix direction for the section-contrast arc's final generation: a light scheme must be able to render fully light while legitimately carrying dark hero bands "by choice." Selects layout-curated background+text variable *pairs* (generalising the chrome pattern already used for `--color-cta-bg`/`--color-cta-text`), palette-overridable per site with specialised slots theme-winning, applied per-instance later via plan directives — components consume pairs and never declare `--section-*` themselves; renderer luminance defaults remain the base case. Judged a completion of the existing architecture, not a restructure. Execution at capture: ten templates fixed, seven verified already-clean (footer, CTA via inverse-pair buttons, hero, five hero-* variants, about-content, brief-explanation), idea.uk's chrome repointed; the full rebuild + Go batch was still pending.
- **sources:** running_notes(22).md Sn, So; RUNBOOK_scheme_to_components(18).md CHECK 4 RESULTS + WHERE WE ARE (both U07)
- **relations:** DES-056 (arc); DES-060 (hazard/band split, the diagnosis this fixes); DES-063 (layout CTA-pair WCAG curation, the layout-library-side twin of this work); DES-062 (chrome linkage tangle, same fix family)
- **verify-later:** SPEC_scheme_to_components.md; layouts cta pair coverage; whether the W6 rebuild shipped

### DES-062 — Chrome linkage tangle: four overlapping header/footer default stores and the hardcoded dark fallback
- **status:** partial
- **status-evidence:** "header_component_id is effectively a DEAD column — nothing populates it"; a repoint was executed for one site; the scheme-aware fallback Go batch was still pending as of the latest snapshot. Independently rediscovered as "Header/footer chrome wiring chain (and its live gaps)" by a different investigation thread the same week, describing the identical four-store tangle and hardcoded fallback.
- **what:** Four coexisting default stores for site chrome: `style_collections.header/footer_component_id` (the store `RenderHeader` reads first — installed NULL and never written by any live path), `site_components` slots (a render cache that can pin an inactive component indefinitely), `sites.default_components` JSONB (the `UpdateSiteDefaultsAction` target, unread on the actual render path), and `layouts.default_*_component_id` (all NULL — no layout declares scheme-appropriate defaults, and site-design-planner never calls `update_site_defaults`). `RenderHeader`'s real chain is collection-id → `GetComponentByFunction("site-header")` → `RenderFallbackHeader`, and that fallback hardcodes dark styling (primary-colour background + white text) — so any break anywhere in this chain yields dark chrome regardless of the site's actual scheme. The library also has light headers but no light footer at all. Fix direction: de-hardcode the active chrome components (the header already models this), repoint stale pins, and make fallbacks scheme-aware using the chrome variable pairs all 18 layouts already define.
- **sources:** running_notes(22).md Sg/Sh/Sl; RUNBOOK_scheme_to_components(18).md CHECK 3b, HEAD-SLOT RESOLUTION, W4b (U07); idea.uk/running_notes_2(6).md mmm findings; REPORT_scheme_does_not_reach_components.md Q6/investigation F; 001_component_flow.md (U04)
- **relations:** DES-056 (arc); DES-004 (component-based headers, the founding rule this tangle undermines); DES-061 (paired-variable fix family)
- **verify-later:** style_collections.*_component_id population; RenderFallbackHeader/Footer/Head current CSS; update_site_defaults_action.go call sites; whether the Go batch shipped

### DES-063 — Layout CTA-pair curation with WCAG contrast gates
- **status:** deployed
- **status-evidence:** "W1 complete + verified"; "W1b COMPLETE: five layouts curated"; a five-layout comment batch records the exact hex swaps and expected contrast values; `layouts.updated_at` trigger "observed working in anger."
- **what:** As part of the section-contrast arc's fix work, `tool-portal-light` gained a missing CTA pair (`--color-cta-bg`/`--color-cta-text` = #e9e2d3/#1a1a1a, contrast ≈13.5, mirroring tool-portal-dark's neutral elevated band) added via an anchored `regexp_replace`. A full sweep then computed every layout's CTA-pair contrast; five seed layouts failing the 4.5 AA threshold with white text got same-hue darker fallback swaps (zero live impact — no site used them yet). Several light layouts were confirmed to deliberately curate DARK footer bands — "light site, dark band by choice" is an intentional, pre-existing curated model in the library, not a bug to fix. `layouts.updated_at` also gained a `BEFORE UPDATE` trigger via the shared `set_updated_at` function during this work. Requirement now carried forward into the layout contract: CTA pair contrast must be ≥ 4.5.
- **sources:** w1_01_add_cta_pair.sql; w1b_01_contrast_batch.sql; RUNBOOK_scheme_to_components(50).md#W1/W1b/W2b-RESULTS; SPEC_scheme_to_components.md#W1 (U03, U07)
- **relations:** DES-056 (arc); DES-061 (paired-variable direction); DES-054 (deterministic contrast gate missing, the generalised version of this problem elsewhere in the pipeline)
- **verify-later:** layouts css_template CTA pair values for the touched layouts; trg_layouts_updated_at

### DES-064 — Section→component resolution: direct-function Path 1 vs scoring selector Path 2
- **status:** deployed
- **status-evidence:** Code read 2026-06-26: "Path 1 = components[sectionName] direct lookup... All current sites hit this path"; `component_selector` "SELECTs is_dark_section into the struct but NEVER uses it in scoring."
- **what:** How a planned section becomes an actual rendered component: Path 1 matches the section name directly against `content_components.function` (one active component per function, enforced by a uniqueness index) — every current site hits this path. Path 2, the scoring `component_selector` (weights: suitable_site_types 0.35, page_types 0.15, quality 0.3, plus specificity and usage), only ever runs for section_type names that aren't functions, and it is scheme-blind despite loading `is_dark_section`. Consequence: there is currently no place in the pipeline to pick a scheme-appropriate component variant for any live site, making a scheme-aware selector necessary-but-insufficient on its own — layout-aware section selection is explicitly documented future work. `page-rerender` re-assembles stored HTML without re-selecting components; only `page-build-handler` re-runs `plan_sections`.
- **sources:** idea.uk/running_notes_2(6).md mmm/nnn corrections; REPORT_scheme_does_not_reach_components.md#2; RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md Step 7 findings (all U04)
- **relations:** DES-056 (arc); DES-065 (component selector + creator, the broader/planned version of Path 2); DES-060 (is_dark_section demotion)
- **verify-later:** plan_sections_action.go Path-1 comment; component_selector.go scoring

### DES-065 — Component selector + creator (section_type vs function split)
- **status:** partial
- **status-evidence:** Framed as Phase 3 "planned" work ("Component selector, patterns, and research") in the earlier archive doc; the scoring selector it describes was later confirmed to exist and run in production (DES-064), though scheme-blind — the `component-creator` LLM-generation half is not independently confirmed built.
- **what:** Splits the planner's historically conflated role: the planner decides WHAT section types a page needs; a Go `component-selector` scores candidate templates by metadata with a fallback to a `needs_new_component` work item when nothing scores well enough; a `component-creator` agent LLM-generates a new template from the full component contract when the library has no match. `function` had been doing two jobs at once (page-role identifier and template choice); `section_type` was introduced to separate them.
- **sources:** WM/007_adoption_pipeline_v3.md#component-selector-and-creator, #component-creation-contracts; FOCUS_interactive_content_generation(3).md#components-more-broadly (U17a)
- **relations:** DES-064 (the deployed reality of half this design); DES-056 (arc, this is the intended long-term fix for scheme-aware selection)
- **verify-later:** content_components metadata columns; component-creator agent existence; plan_sections

### DES-066 — Template helper system (`{{palette}}`/`{{typo}}`/`{{token}}` with mandatory fallback)
- **status:** deployed
- **status-evidence:** Every layout file's CONTRACT CHECK #4; 003_palettes_seed.sql: "Key naming... matches the {{palette \"primary_hover\" \"...\"}} template helpers."
- **what:** A Go-template-style substitution convention embedded directly in the `css_template` CSS text: `{{palette "key" "fallback"}}`, `{{typo "key" "fallback"}}`, `{{token "key" "fallback"}}`, each resolving a JSONB slot lookup with a mandatory literal fallback value. A `{{with palette "heading" ""}}...{{end}}` conditional-block variant also appears. This is the mechanical glue that makes migration 025's FK-composed rows renderable as plain CSS.
- **sources:** layouts/layout_01_brochure-formal.sql#L33-34,L89-138; 003_palettes_seed.sql#L14-19; layout_16_17_vonc_gamesdesign.sql#L96-145 (all U13)
- **relations:** DES-057 (colour inheritance model); DES-013 (migration 025); DES-059 (structure_tokens); DES-067/DES-068 (palettes/typography_sets tables)
- **verify-later:** the Go renderer executing these templates; helper lookup precedence (site-adopted vs seed palette)

### DES-067 — structure_tokens JSONB convention
- **status:** deployed
- **status-evidence:** Present as a populated JSONB literal column in the INSERT of all 17 layout seed files.
- **what:** Each `layouts` row carries a `structure_tokens` JSONB column holding non-colour design tokens — spacing, radii, shadows, transitions, and layout-specific one-offs (e.g. `diagonal_slope_top` for high-energy, `split_pane_left/right` for tool-first-landing). The layout-level counterpart to the palette/typography tables; explicitly excluded from palette extraction.
- **sources:** layouts/layout_01_brochure-formal.sql#L55-69; layout_10_high-energy.sql#L38-53; 003_palettes_seed.sql#L39-41 (all U13)
- **relations:** DES-066 (template helper system); DES-040 (visual identity/effects library, an earlier vaguer version of this idea)
- **verify-later:** `layouts` table column `structure_tokens` DDL/constraints in the Phase 2 migration

### DES-068 — Seed-driver transactional load pattern
- **status:** deployed
- **status-evidence:** `003_layouts_seed_driver.sql`: `BEGIN;` ... 15 `\ir` includes ... a verification block asserting `actual_count >= 15` ... `COMMIT;`.
- **what:** A psql driver script wrapping all 15 numbered layout `\ir` includes in a single transaction with `\set ON_ERROR_STOP on`, so any single layout's SQL error rolls back the entire batch. Ends with a `DO $verify$` block raising an exception if the seeded row count falls below expected. Each individual INSERT is itself idempotent (`ON CONFLICT (name) DO UPDATE`).
- **sources:** layouts/003_layouts_seed_driver.sql (full file); 003_palettes_seed.sql#verify block; 003_typography_sets_seed.sql#verify block (all U13)
- **relations:** DES-069 (palettes table/seed); DES-070 (typography_sets table/seed); DES-014 (layout archetype library)
- **verify-later:** Phase 2 migration creating palettes/layouts/typography_sets tables; confirm this driver actually ran against the live DB

### DES-069 — palettes table / seed (CSS-theme-extracted colour slots)
- **status:** deployed
- **status-evidence:** "Diagnostic run (Phase 3 preflight) confirmed... 13 rows have palette data in css_content"; verify block asserts `actual >= 13`.
- **what:** `palettes` stores one row per design palette (`name`, `display_name`, `colours` JSONB slot map, `category`, `industry_tags`, `origin`, `is_active`). The seed migrated 13 legacy `css_themes` rows via a PL/pgSQL helper `_extract_css_palette` that regex-parses `--color-KEY: VALUE;` declarations; non-colour vars are deliberately excluded (they belong to structure_tokens). One theme, `standard-brochure`, had no palette of its own and was mapped to `default` in a later step.
- **sources:** layouts/003_palettes_seed.sql#header, #_extract_css_palette function, #insert+select, #verify+report (U13)
- **relations:** DES-066 (template helper system); DES-067 (structure_tokens); DES-013 (migration 025)
- **verify-later:** css_themes table; confirm the "Phase 3 Step 3" theme-mapping UPDATE actually ran

### DES-070 — typography_sets table / seed (6 named font/scale bundles)
- **status:** deployed
- **status-evidence:** "Seeds the 6 typography sets described in the migration plan section 8"; verify block asserts `actual_count >= 6`.
- **what:** `typography_sets` stores 6 named bundles — sans-modern, serif-editorial, display-bold, mono-technical, serif-classical, sans-friendly — each with `fonts` JSONB and `scale` JSONB, plus `category`/`industry_tags`. Layouts reference these via `{{typo "key" "fallback"}}`. Each set's description names which layout archetypes it pairs with, a documented convention rather than an FK-enforced constraint.
- **sources:** layouts/003_typography_sets_seed.sql#header, #sans-modern, #display-bold, #mono-technical, #serif-classical/sans-friendly, #verify (U13)
- **relations:** DES-069 (palettes table); DES-066 (template helper system)
- **verify-later:** confirm each layout's declared "Default typography" matches typography_sets.name at composition time

### DES-071 — webdesign-agent CSS rendering pipeline (LLM spec → deterministic Go template → git commit)
- **status:** deployed
- **status-evidence:** "render_css_from_spec — 'Render CSS from design spec using Go template (deterministic, no LLM)'"; full chain observed live; a 4,683-line agent definition file confirms it as the heavyweight regeneration path, contrasted with css-patch-agent for targeted fixes.
- **what:** The webdesign flow: `analyze_design` (LLM → design-spec JSON: color_scheme/typography/spacing) → `render_css_from_spec` (deterministic Go template over DB layout templates — `comp.LayoutTemplate` — merged with palette/typography; forkable themes) → `git_commit` styles.css → site-asset-renderer. The defined CSS vocabulary therefore lives in exactly one Go-owned render path — the single home for generic fixes; `storage_actions.go`'s older styles.css writes belong to a separate legacy builder-extract path (DES-074) and must never be patched for this flow. Caution: re-running `analyze_design` mints a fresh LLM spec, so palettes can shift unless pinned — hence a manual bridge-commit option exists for palette-preserving fixes.
- **sources:** NOTES(43).md §9bi/§9bj/§9bm; RUNBOOK(49).md Part D (U07); 031_webdesign_agent.sql; 076_css_patch_agent.sql; 103_site_design_planner.sql (U18)
- **relations:** DES-003 (composition pipeline); DES-032 (renderer theme-resolution cascade); DES-073 (post-025 CSS content flow, a debugging clarification of this same pipeline)
- **verify-later:** render_css_from_spec_action.go; webdesign-agent default_config; patch_01_git_commit_file_path.go

### DES-072 — Legacy monolithic CSS renderer internals (removed)
- **status:** abandoned
- **status-evidence:** "Phase 4.3 already removed... cssTemplateData struct (and its 16 hardcoded fields)... Compile-clean."
- **what:** The original renderer held a flat Go struct populated by `extractDesignColors`/`designColorMaps`, loading one Go template per theme. Deleted wholesale in Phase 4.3 of migration 025 when the renderer switched to composable palette/layout/typography_set rows joined by FK.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"What Phase 4.3 already removed"; FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Template variable system" (both U12)
- **relations:** DES-001 (three-layer system, superseded predecessor); DES-013 (migration 025)
- **verify-later:** grep codebase for loadCSSGoTemplate/extractDesignColors/designColorMaps — should be absent

### DES-073 — css_templating.go theme-forking bridge (known-broken legacy path)
- **status:** partial
- **status-evidence:** "fork_theme_from_site produces rows with NULL palette_id, NULL layout_id, NULL typography_set_id... Adoption-forked themes are unusable by the render path." Flagged for a Phase 5 rewrite.
- **what:** `TemplateCSSFromSpec` converts a rendered CSS snapshot into old flat-field-name placeholders and writes it to the legacy `css_themes.css_template` column, which the post-Phase-4.3 renderer never reads — silently producing unusable NULL-FK theme rows whenever a site's theme was adoption-forked through this path.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"1. css_templating.go"; FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Ready for deployment" (both U12)
- **relations:** DES-072 (legacy renderer internals); DES-074 (parallel legacy HTML-assembly path); DES-038 (fork lineage, which this bridge was meant to serve)
- **verify-later:** confirm fork_theme_from_site_action.go now produces palette/typography_set rows

### DES-074 — Parallel legacy HTML-assembly render path (getThemeByID / GetThemeByName)
- **status:** partial
- **status-evidence:** "css_content is populated for 13 of the 14 themes. standard-brochure has empty css_content... falls through to GetThemeByName('default')."
- **what:** A second, older render path reads `css_themes.css_content` directly into assembled HTML, independent of the spec-driven render path (DES-071). Left untouched by migration 025's Phase 4; its own known gap (standard-brochure's empty css_content) is flagged for resolution only when Phase 7 drops the legacy columns entirely.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"2. getThemeByID / GetThemeByName" (U12)
- **relations:** DES-073 (css_templating.go bridge); DES-013 (migration 025, Phase 7 legacy-column drop)
- **verify-later:** grep for getThemeByID/GetThemeByName call sites

### DES-075 — Post-025 CSS content flow: empty css_content is by design, not a bug
- **status:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-24 citing doc 027 and `install_site_composition_action.go` L210-212: css_content "intentionally empty — post-025 renderer reads composition via FK chain at render time."
- **what:** A debugging-relevant clarification of the deployed pipeline (DES-003/DES-071): the design pipeline runs `needs_composition` (site-design-planner) → gated `needs_design` (webdesign-agent: analyze_design → update_site → generate_css via render_css_from_spec reading composition FKs → deploy_css writes assets/css/styles.css → optional fork_theme). `css_themes.css_content` is intentionally empty post-025; the empty "Theme-specific styles injected here" head-block comment is expected, not evidence of a broken render. webdesign-agent is not deprecated. Practical consequence for debugging: a wrong colour on a page is more likely a component variable-name mismatch than a theme-injection failure.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-24; RUNBOOK_vonc_migrations(14).md#step-6 (U23)
- **relations:** DES-003 (composition pipeline); DES-071 (webdesign-agent CSS rendering pipeline); DES-005 (resolved_composition pointer semantics)
- **verify-later:** install_site_composition_action.go; render_css_from_spec

### DES-076 — Fork_theme step double-creation guard
- **status:** aspirational
- **status-evidence:** "Guard: require both should_fork_theme AND should_promote_to_library flags. Implementation deferred to Deliverable 6."
- **what:** Once site-design-planner runs, the pre-existing `fork_theme` step still present in webdesign-agent risks creating duplicate theme/collection rows for the same site. The documented mitigation requires both `should_fork_theme` and `should_promote_to_library` flags to be true before forking proceeds, but implementation was deferred.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19…update4(3).md#"6. Risks Still Live" (U12)
- **relations:** DES-038 (fork-with-review gate); DES-003 (composition pipeline)
- **verify-later:** fork_theme step config in webdesign-agent — confirm both flags gate execution

### DES-077 — Vertical-specific planner variants
- **status:** aspirational
- **status-evidence:** Phase 3.5 todo item: "Create veterinary/energy/mortgage/seasonal site planner prompt variant" — all sub-items unchecked.
- **what:** Proposed separate agent definitions using the same planner Go code but vertical-tuned prompt templates, so a well-established vertical produces better plans than a generic planner with config injected — e.g. every breed-health page links to "find a vet for this breed"; every mortgage calculator has lead capture below results.
- **sources:** docs021.../026_implementation_todo_vertical_architecture(2).md#3.5 (U22)
- **relations:** DES-003 (composition pipeline, the planner this would specialise)
- **verify-later:** agent_definitions for veterinary/energy/mortgage/seasonal site-planner variants

### DES-078 — Spec ownership / silent-override failure-mode principle (doc 028)
- **status:** deployed
- **status-evidence:** Restated as settled doctrine from doc 028 (2026-05-26).
- **what:** A named failure-mode taxonomy governing all spec-writing agents in the design-composition pipeline: an agent that changes behaviour on information it did not put in the spec is a bug; an agent that overwrites a spec aspect another agent owns is a category error; an agent that produces the right output but doesn't write it to the spec is not helpful. Applied concretely: `design_reference` is owned by site-adoption-agent and records the source site's design — writing a chosen palette into it would misrepresent the source (this corrected an earlier proposal to do exactly that). Producer/consumer schema drift (e.g. `colour_mood` prose vs `reference_values`; `features {title,description}` vs `{icon,name,description}`) is treated as the same failure shape.
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#5, #2 (U09)
- **relations:** DES-044 (design_reference/design_intent, the model this doctrine protects); DES-051 (Gap B/C, instances of this failure shape)
- **verify-later:** doc 028 statement in full; aspect-writer inventory across agent_definitions

### DES-079 — Composition resolution architecture: three resolvers + install action (implementation detail of DES-003)
- **status:** deployed
- **status-evidence:** "validate_composition_inputs_action.go — DONE... install_site_composition_action.go — DONE (~562 lines)" (2026-04-19).
- **what:** The concrete Go action sequence behind site-design-planner's composition stage: `validate_composition_inputs` → three resolvers (`resolve_composition_layout` tag-overlap match, `resolve_composition_typography`, `resolve_composition_palette` fingerprint→mission→design_intent→layout-inherit→default cascade) → `install_site_composition` (one transaction: css_themes+style_collections insert, sites update, resolved_composition spec write). This is the implementation-level detail underlying DES-003's higher-level pipeline description, recorded separately here because the raw source documented it as its own deliverable with its own verification evidence.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19…update4(3).md#"4. Work Plan — Deliverable 4" (U12)
- **relations:** DES-003 (composition pipeline, the parent concept); DES-005 (resolved_composition semantics); DES-043 (palette cascade dead-slot bug, a bug within resolve_composition_palette)
- **verify-later:** confirm resolve_composition_*.go / install_site_composition_action.go in current codebase

### DES-080 — site-design-planner spec aspects as originally scoped (navigation / layout / resolved_composition)
- **status:** partial
- **status-evidence:** Doc 103 is "Deliverable 2: Spec schemas... pre validation" — documents shapes and creates best-effort validation functions, explicitly not table constraints; reader lists name live actions. Superseded in practice by the narrower "Choice B" scope (DES-006), which only ever shipped `resolved_composition`.
- **what:** Doc 103 defined three site_specs aspects site-design-planner was originally slated to write, separated by reader: `navigation` (nav architecture, items, CTA, mobile pattern → populate_nav_tables, InjectHeader, GetNavItems), `layout` (page-level layout, header/footer style → AssembleMultipageSiteAction, templates), and `resolved_composition` (machine-readable pointers to palette/layout/typography rows + reasoning → render_css_from_spec, webdesign-agent, audit agents). Validation functions run at write time; site_specs itself stays open JSONB. In practice only `resolved_composition` shipped (DES-006); `navigation`/`layout` ownership remained deferred.
- **sources:** 103_site_design_planner.sql (U18)
- **relations:** DES-006 (Choice B scope, the actual narrower outcome); DES-003 (composition pipeline)
- **verify-later:** site-design-planner agent existence and writers of these aspects today


### DES-081 — `data-hcc-carousel`: an opt-in carousel hook any section component can emit, driven by the existing hero-card-carousel snippet
- **status:** deployed
- **status-evidence:** Live on `leopardessconsulting.co.uk/services.html` 2026-07-31, both blocks. Verified by real click in headless Chromium with `--force-prefers-reduced-motion`: `icg.NEXT_SCROLLS=true`, `PREV_RETURNS=true`. Mutant-verified twice — with the snippet's `initAll` disabled both carousels die; with the pre-change *narrow* selector restored, only the info-card carousel dies and `teaser-reveal-panel` is unaffected, which attributes the behaviour to this change and nothing else.
- **what:** A two-part opt-in seam so a section component can present a single-row scrolling card track **without a second carousel implementation existing on the platform**. (1) `js_snippets['hero-card-carousel']`'s `initAll` selector was widened from `.hero-card-carousel[data-component='hero-card-carousel']` to add `, [data-hcc-carousel]`; every other line of that snippet was already generic, reading only `data-hcc-track` / `-slide` / `-prev` / `-next` / `-live` / `-autoplay` and no class or component name. Its `applies_to` gained `info-card-grid` so the bundler ships it to sites using that component. (2) `content_components['info-card-grid']` gained a `carousel` boolean in `content_data`: set true, the container becomes a scroll-snap track with overlaid prev/next arrows and each card gains `data-hcc-slide`; absent, **the rendered output is byte-identical to the pre-change template**. Proven, not asserted: all 18 live instances across 9 sites rendered through both templates with `text/template` (the engine `executeGoTemplate` uses) came out byte-identical, 0 errors — and the same 18 rendered *with* the flag all differed, so the comparison was measuring something. 0 of 18 instances carried a `layout`/`display`/`carousel` key beforehand, and `data-hcc-carousel` appeared in 0 `page_components` and 0 `site_components` fleet-wide, so the new arm could not reach any existing page.
- **why this shape and not a fork:** RUNBOOK landmine 14 for the leopardess lane — *a section component fork does NOT survive rerender*, because `save_page_sections` re-links to the canonical component by `function`. A forked carousel component would have been silently reverted on the next rerender. An opt-in flag on the canonical row is the only variant that survives.
- **first-execution note, worth knowing before you build on it:** `hero-card-carousel` had **never rendered anywhere on the fleet** before this — 0 `page_components`, 0 `site_components`, `usage_count` 0 — so although `is_active` was true, its snippet had never been bundled into any site and had **never run in production**. `teaser-reveal-panel` only ever copied its `goTo`/`nearestIndex` *pattern*, not the file. 2026-07-31 is its first execution. Treat its untested corners (auto-advance, the pause control, the `IntersectionObserver` arm) as unexercised: this consumer sets `data-hcc-autoplay="false"` and emits no pause button, so none of that code has run either.
- **cost to other consumers:** the 8 other sites using `info-card-grid` gain ~4.7KB of inert JS in `assets/js/snippets.js` at their next bundle regeneration (they have no `data-hcc-carousel` element, so `initCarousel` is never called). No page markup changes for them.
- **sources:** leopardessconsulting/RUNNING_NOTES.md#2026-07-31; leopardessconsulting/AUDIT_verified_facts.md#Re-measurement 2026-07-31
- **relations:** DES-009 (the css_snippets/js_snippets system this rides on); DES-064 (section→component resolution, the `function`-based re-link that rules out a fork); DES-055 (three-per-row no-orphan grid rule — the wrapping-grid problem a carousel sidesteps entirely)
- **verify-later:** whether any second consumer adopts `data-hcc-carousel`; whether the snippet's auto-advance/pause arms ever get exercised

### DES-082 — `allow_reinstall`: the opt-in that lets a site's composition be REPLACED, and the manual window it closes
- **status:** built 2026-08-10, **not yet exercised in production** — code committed, rides the next chassis roll. No live re-compose has run through it.
- **status-evidence:** three unit tests pass and were proven load-bearing by mutation (`if !allowReinstall` → `if false` fails A and C; hardcoding the refusal fails B), run 2026-08-10 before commit. `go build ./platform/...` clean. **No pod-grep yet — the binary carrying it has not been built.**
- **what:** `install_site_composition` has always loud-failed on a site whose `sites.style_collection_id` is already set ("re-resolve not supported"), and its own log line recommended an operator **null the column by hand**. That manual route is unsafe: between the clear and the re-resolve the site is uncomposed, and anything that renders in that window hits the composition loader's emergency fallback (`render_css_composition_loader.go:144-158`) and can deploy a `standard-brochure` stylesheet over a live site. `allow_reinstall` (step config literal, **default false**) opts into replacing the composition **inside the action's existing transaction**, so the window never opens. Read via `datahelpers.GetBoolFieldLoud`, so a malformed declaration (a string `"true"`) warns and falls back to the SAFE branch rather than switching the unsafe direction on. The link UPDATE's race guard changed from `AND style_collection_id IS NULL` to `AND style_collection_id IS NOT DISTINCT FROM $3::uuid` — it still refuses to clobber a concurrent install, in both modes. Returns `previous_collection_id` (the rollback value, and the only record of it, since the UPDATE overwrites in place) and `replaced_existing`.
- **the landmine:** **the default is OFF and must stay OFF.** The permissive branch re-points a LIVE site's whole stylesheet. Shipped as a flag rather than a doc comment per the owner ruling of 2026-08-02 (RFC_010): new authority on a shared seam ships as an opt-in field with the unsafe default off, because "callers must all be X" is not a control on a tree this many sessions share. **No live workflow sets it yet** — it is reachable by nothing until a step config names it.
- **the open review question:** should a re-compose additionally require HITL approval (a `needs_theme_review`-style gate) rather than a caller-set boolean? The original code called re-resolution "a deliberate future feature behind HITL"; this delivers the mechanism without the HITL half, on the ground that the caller is a workflow author under review, not an end user. Not settled — routed to the council gate with this commit.
- **other consumers to tell (owner ruling 2026-07-29 §3):** `discovery_checks/check_missing_css.go` defers the no-collection case explicitly *because* "routing composition here would fail the install guard" — that constraint is now liftable, and whether missing_css should route composition again is a decision for its owner, not a side effect of this change.
- **sources:** `platform/orchestration/actions/install_site_composition_action.go`; `install_site_composition_reinstall_test.go`; `bugs_open/113` (2026-08-10 sections)
- **relations:** composition resolution architecture (DES-079); fork_theme double-creation guard (DES-076); the composition loader's emergency fallback; `bugs_open/113`

### DES-083 — Discovery's composition pair now emits `triaged`, matching the build path it claims parity with
- **status:** built 2026-08-10, rides the next roll. **Not yet observed emitting** — no discovery run has produced a pair since the change.
- **status-evidence:** the disagreement was measured, not inferred: `emit_design_items_action.go:126,171` emits `needs_composition`/`needs_design` as `triaged`; `check_integrity.go` emitted the same two `item_key`s as `detected`. The dispatch loop claims only `status IN ('triaged','approved')` (`claim_work_item_action.go:102`). Four stranded rows on 2026-08-10, oldest `loancalculator.co.uk` ~33h.
- **what:** `MissingStyleCollectionCheck` (discovery) emits the composition pair with `Status: "triaged"` so the build dispatch loop can claim it. Previously `detected`, which was only ever promotable by `TriageDetectedItemsAction` — whose three callers (`improvement-loop`, `design-audit-agent`, `site-review-agent`) have **no enabled scheduled task between them** (`improvement-sweep` is `enabled=false`). So the identical pair dispatched from the build path and never from discovery.
- **the landmine:** **do NOT "fix" this class by enabling a fleet-wide triage sweep.** `TriageDetectedItemsAction` promotes every `detected` row for a site with no type filter, and there were **448** `detected` build-pipeline rows fleet-wide on 2026-08-10 (193 `page_rerender`, 79 `contrast_failure`, …). Enabling a sweep dispatches all of them at once. The repair belongs at whichever producer disagrees with its twin.
- **the open review question:** the general case is untouched — every other discovery check still emits `detected`, and the promoter is still undriven. Whether the other types should be promoted, and by what, is unanswered; this change deliberately fixes only the pair whose two producers contradicted each other.
- **sources:** `platform/orchestration/actions/discovery_checks/check_integrity.go` (MissingStyleCollectionCheck); `emit_design_items_action.go`; `triage_detect_items_action.go`; `bugs_open/113`
- **relations:** DES-082 (same lane, same commit); dedup index ↔ Go list lockstep (`idx_swi_dedup`, one open row per site/item_key)

> **CORRECTION 2026-08-10 (later), to DES-082 and DES-083 above — read this before citing either.**
>
> **(a) DES-082 is LIVE but NOT SAFELY USABLE, which is worse than "built, not exercised".**
> Pod-grepped on chassis `696d88b4c7`, both replicas: `allow_reinstall` ×4,
> `previous_collection_id` ×1, `re-resolve not requested` ×1, and the string the change
> REMOVED (`re-resolve not supported`) ×0 — a true removal-based negative control, so the
> binary genuinely carries it. **But the flag is read from `StepConfig.Config` only**, and
> `site-design-planner`'s install step config holds nothing but path references. So the only
> way to set it is to edit the agent definition — which turns re-install **ON for every
> composition install fleet-wide**, i.e. exactly the unsafe-default-ON state the flag exists
> to prevent. The council's `editquality` seat called the capability "safe but inert"
> (medium); it is in fact *unusable for a single site*, which is the case it was built for.
> **Revision needed: read the flag per-request (work-item spec / `input_data`) as well as
> from step config.** Until then DES-082 delivers no repair path.
>
> **(b) DES-083's justification cited the wrong work item, and the claim is withdrawn.**
> `47ce091c` is `item_type = needs_design_review` (item_key `shared_style_…`), created
> **2026-04-24**, not the `needs_composition`/`needs_design` pair and not 2026-08-06 — I read
> `updated_at` as the creation date. **The `triaged` change does not unblock it and never
> could.** What DES-083 actually unblocks is three sites' composition pairs
> (`noted.co.uk`, `loanandmortgagecalculator.co.uk`, `loancalculator.co.uk`) — six rows,
> none of them `ai-agent-orchestration.com`. The change itself is still correct (two
> producers of one `item_key` disagreed about status); only the evidence I attached to it
> was wrong. Caught by the council gate, round 1. Full account: `WRONG_CALLS.md`, 2026-08-10.
>
> **(c) The mismatch pattern is bigger than DES-083's landmine says.** That entry sized it at
> 448 `detected` **build**-pipeline rows. The council's own query shows the same
> status/dispatch mismatch across other pipelines too — `undeployed_asset` 86 (design),
> `phantom_internal_link` 18 (content), `unbuilt_internal_link` 17 (content),
> `image_url_404` 16 (design). The "do not enable a fleet-wide sweep" conclusion is
> unchanged and, if anything, stronger.

> **DES-082 UPDATE 2026-08-11 — the round-1 defect above is FIXED in code (not yet rolled).**
> `allow_reinstall` is now read from **two** sources, both defaulting false and both through
> `GetBoolFieldLoud`: the step's own config (an agent-definition edit, therefore fleet-wide)
> **and the dispatching work item's `spec`** (per-request — one dispatch opts in, nobody
> else changes). Step config is checked first; the spec is consulted only if it did not opt
> in. **Prefer the per-request source**; setting it on `site-design-planner`'s step would
> turn re-install on for every install, which is what the flag exists to prevent.
> Helper `requestSpecFromCollected` resolves `input_data.spec` and the `input_data.body.spec`
> wrapper. Two new tests, proven load-bearing by mutation (nil-ing the spec lookup fails the
> per-request test **alone**). **`[UNMEASURED]`: I have not observed a live
> `needs_composition` dispatch's `collected_data` shape — if it produces a third shape the
> flag silently will not arrive, which fails SAFE but looks like a broken flag.**

> **DES-082 / DES-083 UPDATE 2026-08-12 — BOTH ARE NOW EXERCISED IN PRODUCTION, and the
> `[UNMEASURED]` above is settled. This supersedes the "not yet exercised" statuses.**
>
> **DES-082 status: LIVE and USED.** First production re-compose ran 2026-08-12 13:50:13Z —
> work item `57b9b3ff` carrying `spec.allow_reinstall=true` moved
> `ai-agent-orchestration.com` off the shared LIGHT seed collection `3196d966` to its own
> `a0f1ac70`, and the served `--color-card-bg` went `#ffffff` → `#0D1117`
> (`styles.css` last-modified 2026-08-11T16:22:21Z → 2026-08-12T13:56:26Z). Chassis
> `v1.0.1290`, stamp `fa078ab3d`, binary-probed both replicas with a bogus-sha control.
> **The per-request channel is proven BEHAVIOURALLY, which is stronger than the log line:**
> the install refuses by default and `site-design-planner`'s install step config carries no
> `allow_reinstall` key, so a successful install on an already-composed site is reachable
> only via the work item spec.
>
> **The `[UNMEASURED]` shape worry is answered, and the answer removed code.** Over 30 days
> of `orchestration_states` carrying `input_data` (n=6,397): `input_data.spec` present on
> **2,363**, `input_data.body.spec` on **ZERO**. The `body.spec` wrapper branch named above
> was dormant machinery and is **deleted**; `requestSpecFromCollected` now reads the one
> measured path via `input_contracts.GetValueAtExactPath` — deliberately NOT
> `datahelpers.ExtractNestedField`, which auto-unwraps through a `.response` envelope and
> would make an AUTHORITY switch satisfiable by a `true` arriving inside another agent's
> reply. It also returns a REASON on failure, so "no spec arrived" and "a spec arrived that
> is not an object" are distinguishable in the log rather than both reporting the first.
>
> **⚠ THE REPAIR IS TWO WORK ITEMS, NOT ONE — this is the trap this entry most needs to
> carry.** `install_site_composition` completes, changes the DB, and **queues nothing**;
> `styles.css` is rendered by `webdesign-agent` off the *other* half of the pair
> `MissingStyleCollectionCheck` emits. A hand-written repair copying only the
> `needs_composition` half leaves every DB check green while the site serves the old
> stylesheet indefinitely. Fire `needs_design` too (`status='triaged'`), and verify at
> `curl -sI …/styles.css | grep last-modified`, never at the item status. Now in
> `LANDMINES.md` with 12 footprint keys.
>
> **DES-083 status: OBSERVED EMITTING AND DISPATCHING.** Two discovery-produced
> `needs_composition` items (`created_by=design-discovery-agent`, `source=discovery`) were
> claimed and ran to `complete` on 2026-08-11 — `loancash.co.uk` (`fef16250`) and
> `cookly.uk` (`da0b080d`). Before the change they would have sat at `detected`, which
> nothing promotes. **Note the council's `improvement_guardian` seat objected to this change
> at HIGH severity in round 2** (it reads `detected` as a deliberate human-triage gate and
> wanted `emit_design_items` moved to `detected` instead). That objection is recorded and
> not settled by these two runs: they show the change works, not that the gate it removed
> was unwanted. The larger question — whether build-pipeline discovery findings should be
> human-gated at all, and by what — is still open and is bigger than this entry.
>
> **Council trail complete: `b8e341b9-4709-49ad-8b7b-f4c8894ba551` — REVISE, REVISE,
> APPROVED** (round 3, 2026-08-12, 12 reviewers, no high-severity objections). Two of the
> three rounds found real defects. Round 3's own approved verdict carried an advisory that
> was also correct and is fixed in `9d4fbb4f7`. **Still owed: a pod-verify of round 3 after
> the next fleet roll** — it is committed, not rolled.
> **Live consumers, enumerated per RFC_022 rather than asserted (2026-08-11): `0` active
> agent definitions and `0` work items name `allow_reinstall`.** Council round 2 submitted
> under the same trail correlation `b8e341b9-…`.

### DES-084 — Re-compose approval: a recorded approver on every composition replace, defaulting to a GRANT
- **status:** built 2026-08-12, **not rolled**. No live consumer names an approver yet, so every replace so far would record the sentinel.
- **status-evidence:** three tests, proven load-bearing by mutation run before commit — returning a person-looking name instead of the sentinel fails the default test alone; skipping the spec lookup fails the named-approver test alone. `go build ./platform/...` clean.
- **what:** `install_site_composition`'s replace path now always resolves and records **who approved it**, returned as `reinstall_approved_by`. Order: step config `reinstall_approved_by` → work item spec `reinstall_approved_by` → work item spec `approved_by` → the sentinel `reinstallDefaultApprover` (`"default-grant/owner-2026-08-12"`). `approved_by` is in the list on purpose: it is the column a real HITL approval flow already fills, so wiring one needs no change here. The replace Warn log carries `approved_by` and `approval_was_explicit`.
- **the landmine:** **the sentinel is a STORED VALUE, so rewording the constant silently splits the audit population in two.** The whole point of a sentinel distinct from a name is that `SELECT result->>'reinstall_approved_by', count(*) … GROUP BY 1` separates "a human said yes" from "the standing default said yes for them" — which is the only thing that makes tightening the default a measurable change rather than a leap. Do not tidy the string.
- **the open review question:** nothing BLOCKS on approval today — the ruling was "approval needed but for now default that the human approves", so this delivers the *record* and defers the *gate*. Tightening is one line (return `""`, have the caller refuse), but it should not be flipped until the query above shows who would have been refused.
- **sources:** `platform/orchestration/actions/install_site_composition_action.go` (`resolveReinstallApprover`, `reinstallDefaultApprover`); `install_site_composition_reinstall_test.go` (tests G/H/I); owner ruling 2026-08-12
- **relations:** DES-082 (`allow_reinstall`, the flag this approves); `bugs_open/113`

> **DES-084 STATUS CORRECTION 2026-08-15 — "built, not rolled" is STALE. It is LIVE, and it has NEVER RUN.**
> Shipped in chassis stamp `0115f2b4528b0063fd01e7af275ccefe9c5a991d`: `git merge-base
> --is-ancestor 1fa86f5cc 0115f2b4` succeeds, and the chassis binary carries that sha
> (probed on `agent-chassis-7779f5d998-96lpf`, with a `deadbeef` control absent). **Do not
> re-verify with `strings`** — CLAUDE.md retired it 2026-08-11.
>
> **DEPLOYED ≠ EXERCISED, and here the gap is total.** Measured 2026-08-15:
> `SELECT result->>'reinstall_approved_by', count(*) … GROUP BY 1` → **0 rows**;
> `result->>'replaced_existing'='true'` → **0**. No composition replace has run since the
> one that repaired `ai-agent-orchestration.com`, so **not one line of this has executed**
> and the sentinel has never been written. The audit query the whole design rests on has no
> population yet — which also means **the eventual tightening of the default still has
> nothing to measure**, and that was the entry's stated purpose.
>
> **REVIEW CORRECTION, and it is mine to own.** This shipped under
> `Council-Submitted: b8e341b9`, whose trail reached APPROVED at round 3 — **on a plan that
> never contained the approval feature** (its edits were `allow_reinstall`, the race guard,
> `previous_collection_id`, the per-request spec read and the `check_integrity` status
> change; `resolveReinstallApprover` appears in none of them). The 098 report therefore
> credits `1fa86f5cc` as `[b8e341b9, by correlation, via submitted]` for a review that never
> looked at it. **Not a MISMATCH by the report's own rule** (`Council-Submitted` asserts
> nothing, and the run shows `MISMATCH: 0`) — but misleading in effect, and the fix is a
> correlation of its own, not a footnote. Submitted 2026-08-15 as
> **`9767969e-92fa-44d0-b416-d7187c869531`**, with the over-credit named in its rationale.
> **Verdict not yet read — owed.**

### DES-085 — `theme_kits`: a named, listable bundle over the composable-theme system (Phase 1)

- **status:** **LIVE AND UNEXERCISED, verified 2026-09-02/03 — applied AND
  rolled, and adopted by NOTHING.** Phase 1 scope only (registry + apply +
  page structure; nav patterns, voice presets and "create from example"
  designed but not built, see the plan).
  > **CORRECTED 2026-09-03 — this line read "NEVER EXERCISED — not applied,
  > not rolled" for a day after both halves became true.** That is this
  > register's own stale-status landmine firing a second time on the same
  > entry, in the opposite direction: the first correction (below) removed a
  > false "deployed", and then the replacement outlived ITS truth too. Both
  > halves now verified:
  > - **Binary: LIVE.** `agent-chassis` on `v1.0.1355`. The `build provenance`
  >   startup line had already scrolled, so this is a `/proc/1/exe` capability
  >   probe **with both controls**: `apply_theme_kit` PRESENT,
  >   `page_archetypes` PRESENT, `fork_theme_from_site` PRESENT (positive
  >   control, pre-existing action), `zzz_not_a_real_action_zzz` ABSENT
  >   (negative control — the probe discriminates).
  > - **Schema: APPLIED 2026-09-02**, migrations 689 + 691, via a **scoped**
  >   run (`MIGRATIONS_DIR` pointed at a temp dir holding only those two
  >   files, because `--apply` takes EVERY pending file and would have swept a
  >   dozen other lanes' migrations). `to_regclass` returns both tables; 4
  >   kits and 14 fleet archetypes seeded.
  > - **Adoption: 0.** `SELECT count(*) FROM site_specs WHERE
  >   aspect='theme_kit_adoption' AND is_current` → 0 as of 2026-09-03. So
  >   every kit-conditional branch is live, reachable, and has never run.
  >   **Cite this entry as "built and reachable", never as "working".**
- **council verdict — READ THIS BEFORE THE `what:` LINE.** Round 1
  (correlation `bed139b2-f512-436a-9ba8-ff2fbfade8ef`) came back
  **`complete_revise`**, `decided_by: gating objection from guardian`. The
  objection was about the SUBMISSION, not the code, and it was right: the
  rationale claimed a fixed typography cascade asymmetry that the plan's
  sketch never showed, so it was unverifiable from the submission. Resubmitted
  on the same correlation 2026-09-03 with the guard sketched, every
  predicate-dependent figure carrying a runnable query, and one of round 1's
  own evidence claims retracted as false — **that retraction was itself wrong
  and is withdrawn; see the landmine bullet below.**
  **FOUR ROUNDS, ALL READ: `revise` (21:43Z 09-02) → `revise` (15:32Z 09-03,
  which found a REAL DEFECT) → `revise` (15:56Z, gating on my having deferred
  its fix) → `approved` (16:19Z, "approved with 7 advisory objection(s) — none
  high-severity", 3 abstained).** So `Council-Reviewed:` is legitimate on this
  work from 2026-09-03; the earlier commits carry `Council-Submitted:` and 098
  credits them automatically now the correlation has approved.
  **⚠ THE APPROVAL IS NOT A CLEAN BILL, and the architecture seat's two
  objections read as a GATE on adoption rather than a note. Both are quoted
  because a reader of this entry is exactly who they are addressed to:**
  - *"All four seeded kits pin chrome identical to the unpinned default — the
    chrome dimension of a kit is currently a no-op. **Shipping more kits or
    adopters before this is addressed overstates what a kit does.**"*
  - *"Palette cannot reach the served stylesheet under the current
    render-overlay precedence — `theme_kits.palette_id` is **structurally
    decorative**. This is an architecture gap, not a bug in this plan, but it
    should **block further palette-bearing kit adoption** until the precedence
    is fixed or the capability is explicitly dropped from the contract."*
  **So: do not adopt a kit onto a site until the contract states what a kit
  actually delivers.** Other advisory objections worth knowing before extending
  this: `reuse_agent` asks why a NEW bundling table rather than extending
  `style_collections` (which already bundles per site); `constitution` objects
  that `apply_theme_kit` re-implements supersede-then-insert instead of reusing
  `WriteSiteSpecAction`; `bug_historian` notes the round-4 guard detects and
  does not prevent.
- **⚠⚠ THE DEFECT THAT MATTERS MOST, AND IT IS UNFIXED: A KIT APPLIED BEFORE
  CLASSIFICATION SILENTLY LOSES PALETTE AND TYPOGRAPHY, SO KITS ARE DEFEATED ON
  EXACTLY THE PATH THE OWNER ASKED FOR.** Found by the council gate, round 2,
  not by the author. On the FRESH path (`082` with no `--from`)
  `domain-research-classifier` writes `design_intent` AFTER `apply_theme_kit`
  does, and `write_site_spec` supersedes the current row after a deep merge in
  which **scalar keys are overwritten by the incoming value** — so the kit's
  `reference_values` are discarded. `[VERIFIED 2026-09-03 by reading the file]`
  **there is no guard**: `grep -n "classifier\|domain-research"` in
  `apply_theme_kit_action.go` finds only comments about the ruling, never a
  predicate.
  - **layout SURVIVES** — it is read from aspect `theme_kit_adoption`, which
    the classifier does not write.
  - **palette is discarded** — moot for appearance, since no `design_intent`
    palette reaches the 8 core slots anyway.
  - **TYPOGRAPHY IS DISCARDED AND TYPOGRAPHY IS THE DIMENSION THAT RENDERS.**
    This is the one that costs something.
  - **`design_intent.<dim>.locked` does NOT protect against this** — do not
    read this entry's ruling bullet as saying it does. `locked` is read when
    `apply_theme_kit` writes; **nothing makes the CLASSIFIER respect it**, and
    the key survives the deep merge while the values do not, so the row ends up
    **asserting a human pin over a classifier's values.**
  **So a kit works on an ALREADY-CLASSIFIED site and is silently defeated on a
  new one — the inverse of the owner's *"by default it can start with a
  theme"*.** Recorded with three costed remedies as **`bugs_open/438` §6d** (a
  CONTRIB, because 438 §6a-bis already owns the mechanism), and documented in
  the action's own header.
  **PARTLY FIXED 2026-09-03 (`b18091066`) — the loss is now RECORDED, not
  silent.** Council round 3 gated on the fact that I had diagnosed this,
  accepted it and shipped nothing, so the remedy was built rather than deferred
  a third time. `classifierDesignIntentState()` asks whether
  `domain-research-classifier` has ever written this site's `design_intent`; if
  not, the apply writes `design_intent_supersede_risk` into **three** surfaces —
  a WARN naming the mechanism, the `theme_kit_adoption` spec (durable and
  queryable: a reader asking *"why has this themed site not got the kit's
  fonts?"* finds the answer on the adoption row), and the action's result. A
  three-state STRING, never a bool, so a read failure cannot be recorded as "no
  risk". **It REPORTS and does not REFUSE** — layout survives on a different
  aspect and is the only dimension a kit moves, so refusing would throw away the
  working part to protect the broken one.
  **Proven by two mutations** (the guard going blind; a confident wrong answer
  on error), both red, restored green, evidence in the test header; the
  predicate discriminates **38 of 39** live sites, and 39 or 0 would have made
  it decoration.
  ⚠ **The ORDERING is still unfixed and stays `bugs_open/438`'s** — the
  classifier still supersedes and the kit's typography is still lost on the
  fresh path. The guard makes the loss visible, not impossible. The other two
  candidates remain architecture-scope (make the classifier respect `locked`) or
  build on 438's own defect.
  ⚠⚠ **AND THE GUARD IS COMMITTED, NOT LIVE — do not cite this bullet as a
  protection that is in force.** `[VERIFIED 2026-09-03 at the pod, with both
  controls]` `classifierDesignIntentState` and `at_risk_no_classifier_write_yet`
  are **absent** from the running `agent-chassis` binary; `apply_theme_kit` is
  PRESENT (positive control) and a nonsense needle is absent (negative control);
  the pods predate the commit by hours. It rides the next roll. It is therefore
  inert for TWO reasons, not one — adoption is 0 **and** the code is not in the
  image. The council's `debug_historian` seat asked for this check, and this
  entry had recorded only the weaker reason.
  ⚠ **`design_intent.<dim>.locked` is read at exactly THREE call sites, all in
  `apply_theme_kit_action.go`** (464, 465, 469) — `[VERIFIED 2026-09-03]` nothing
  else in `platform/`, `internal/` or `cmd/` reads it. So "nothing makes the
  classifier respect it" now has evidence rather than being asserted, which is
  what the `prior_art_librarian` seat objected to.
  > **CORRECTED 2026-09-02, same day as filing.** This entry first read
  > "deployed … live since commit `0902039c0`". **Both halves were false and
  > the check took one command each**, which is exactly the shape this
  > register's own landmine warns about (a STATUS line outliving its truth,
  > read by council seats as ground truth):
  > - **DB half:** `SELECT to_regclass('public.theme_kits')` → NULL. The
  >   migration was only ever DRY-RUN (`run-migrations.sh` with no `--apply`,
  >   which reports "ran to its own COMMIT without error (everything rolled
  >   back)" — a phrase that reads like success and means the opposite of
  >   applied). `theme_kit_adoption` specs: 0.
  > - **Binary half:** no `make build-*` or roll was ever run this session, so
  >   `apply_theme_kit` is in no running image. `install_site_composition_
  >   action.go`'s kit lookup swallows the error when the table is missing, so
  >   a live chassis would silently skip the kit path and log nothing.
  >
  > Caught by a Fable architecture review, not by the author. Do not cite this
  > entry as evidence the mechanism works until the migration is applied AND a
  > roll carries the binary — then re-verify BOTH and rewrite this block.
- **status-evidence (of the CODE existing, not of it running):**
  `docs/agent_docs/sql_for_agents/689_theme_kits.sql` (schema + 4 seed kits +
  fleet `page_archetypes` rows), `apply_theme_kit_action.go`,
  `theme_kit_defaults.go`, `page_archetypes_resolver.go` — committed in
  `0902039c0`. **The migration number was 686 in that commit and COLLIDED**
  with `686_article_body_hero_image_capability.sql` (another session, filed one
  minute later the same afternoon); renamed to 689 afterwards.
- **what:** a NEW table, `theme_kits`, deliberately NOT named `themes` —
  `css_themes`/`theme_id`/`needs_theme_review`/`forked_from_theme_id` already
  mean "one site's CSS composition record" throughout this codebase (DES-003
  below), and a second table called `themes` would make every `theme_id` in
  the tree ambiguous. A kit is a thin FK registry over the EXISTING composable
  system (`layouts`/`palettes`/`typography_sets`/`content_components` chrome)
  — it adds no new visual-design mechanism, only a named, selectable bundle
  and a materialize-on-apply action. Applying a kit (`apply_theme_kit`)
  writes the kit's resolved palette/typography into `design_intent.{palette,
  typography}.reference_values` (picked up by the EXISTING resolver cascades
  with zero resolver changes for those two dimensions) plus a
  `theme_kit_adoption` site_specs lineage row, and queues `needs_composition`
  — it never installs anything itself (site-design-planner stays the one
  writer of `sites.style_collection_id`, per the "Choice B" precedent this
  platform already settled once: `HANDOFF_2026-04-19_..._update4(3).md`
  narrowed site-design-planner's scope to composition-resolution only, never
  navigation/structure). Layout is the one dimension needing an actual
  resolver change (`resolve_composition_layout_action.go`), since layout
  resolution never consults `design_intent` at all.
- **A KIT IS A STARTING POINT, NEVER A CONSTRAINT (OWNER RULING 2026-09-02):**
  *"by default it can start with a theme and change it if it wishes, but it
  must have full authority to ignore our set of themes if it chooses."* So
  `apply_theme_kit`'s default mode is `start` — it WRITES the kit's values,
  superseding what is there. (It first shipped defaulting to `fill_gaps`, which
  deferred to any existing `design_intent` and was therefore a no-op on the 33
  of 57 sites the classifier had already touched — a theme that never actually
  started anything.) Written values carry `reference_source: "theme_kit:<name>"`
  and `reference_is_default: true`, so a later reader can tell a kit's default
  from a decision, and can override it freely. `fill_gaps` remains as an
  explicit conservative mode; `reapply` also replaces an installed composition.
  The ONE thing no mode overwrites is `design_intent.<dim>.locked: true`, a
  deliberate human pin that nothing sets automatically. **The corollary matters
  as much: nothing anywhere freezes a themed site's values against the
  classifier or the render overlay** — RFC_059 proposed exactly that and was
  WITHDRAWN under this ruling.
- **the fork idiom continues, not diverges**: applying a kit MATERIALIZES
  defaults into a site's own rows — never a live FK the site stays bound to.
  A site can edit any component/palette/section afterward exactly as before;
  nothing checks "is this site themed" on that path. `page_archetypes`
  (replacing the hardcoded `defaultSectionsForPage` Go switch,
  `apply_gap_plan_action.go:995-1042`, kept as a logged last-resort fallback)
  is three-way scoped (site > theme kit > fleet, `CHECK` mutual-exclusive) so
  a site can declare its own durable structure default WITHOUT adopting any
  kit — "sites don't necessarily have to be created from a theme."
- **landmine avoided, and it is CORRECT AS ORIGINALLY WRITTEN. A 2026-09-03
  attempt to retract it was itself wrong and is withdrawn — see the nested
  block; read that before quoting either version.** The CONCLUSION and the
  reasons both stand: the seed (migration **689**, not 686 — renumbered after
  the collision noted above) hardcodes verified chrome UUIDs rather than a
  function-name subquery. `content_components.function` is not unique after
  the canonical predicate, the rows **named** `site-header`/`site-footer` are
  `component_level='section'` and are NOT chrome-eligible despite the matching
  function, and the eligible rows for those functions are the ones **named**
  `header-theme-chrome` / `footer-theme-chrome`.
  > **⚠ WITHDRAWN CORRECTION, 2026-09-03 — I retracted a TRUE claim by
  > querying the wrong column, and the retraction stood in this entry and in a
  > council submission for part of a day.** I ran
  > `WHERE function LIKE '%theme-chrome%'`, got 0 rows, and concluded that
  > `header-theme-chrome`/`footer-theme-chrome` "do not exist in any state".
  > **`content_components` has BOTH `name` AND `function`.** Those two strings
  > are `name` values; their `function` values are `site-header`/`site-footer`,
  > which is exactly the distinction the original entry was drawing. Verified
  > by id — the two UUIDs migration 689 pins resolve to
  > `header-theme-chrome`/`footer-theme-chrome`, `component_level='site'`,
  > `is_active`, unforked. **70 files in this tree name these components and
  > migration 339 has `RAISE EXCEPTION` drift guards on updating them**, which
  > is what made me look again.
  >
  > **The check that settles it, and the one to use — select BOTH columns, never
  > filter on one and conclude about the other:**
  > ```sql
  > SELECT name, function, component_level, is_active,
  >        forked_from IS NULL AS unforked,
  >        (is_active AND component_level IN ('site','header','footer','head')) AS chrome_eligible
  >   FROM content_components
  >  WHERE function IN ('site-header','site-footer')
  >  ORDER BY function, chrome_eligible DESC, name;
  > ```
  > **11 rows as of 2026-09-03. "Eligible" then depends on WHICH of the two
  > predicates you mean, and there are deliberately two** — a third thing I had
  > to get precise, and the code documents it in full at
  > `component_library.go:336-378`:
  > - **`chromeEligibleSQL`** — the POOL SELECTION predicate, `is_active AND
  >   forked_from IS NULL AND component_level IN (…)`. Under it exactly **2**
  >   rows qualify: `header-theme-chrome` and `footer-theme-chrome`.
  > - **`chromePinEligibleSQL`** — the predicate for a
  >   `style_collections` PIN, which **omits `forked_from IS NULL`
  >   deliberately**, because naming a site's own fork is exactly what a pin is
  >   for. Under it **3** rows qualify: those two plus **`header-leopardess`,
  >   an ACTIVE FORK of one client's header.**
  >
  > Both numbers are correct under their own predicate, which is the same trap
  > as the collision figure below arriving for a third time in one entry.
  > `forked_from IS NULL` is load-bearing in the pool predicate and the source
  > comment names this exact row as the reason: *"an ACTIVE fork of one site's
  > header is a candidate to become every other site's header … header-leopardess
  > sorts first among active site-header rows and is what link_site_components
  > would have assigned."* So **no, a client's fork does not win the default** —
  > `ResolveChromeComponent` orders by the pool predicate first. The rows named
  > `site-header`/`site-footer` are `section`-level and ineligible under both,
  > as originally stated.
  >
  > **So the reason to hardcode UUIDs is SHARPER than either version said:** a
  > function-name subquery for `site-header` is ambiguous under the PIN
  > predicate (2 rows) and the extra row is a single client's fork.
  > `bugs_closed/118` had already found `header-leopardess` as exactly this
  > hazard, and `chrome_pin_test.go` pins the asymmetry — make the two
  > predicates equal and it goes red with the reason.
  >
  > The collision figure was true but **unverifiable as written, and its
  > denominator was stale** — that half of the 2026-09-03 correction stands.
  > "3 collisions of 364" holds only under `is_active AND forked_from IS NULL`;
  > raw it is **84**, and distinct `function` values are **425 raw / 410
  > canonical as of 2026-09-03**, not 364. A council reviewer reconstructed the
  > predicate from prose, got a different number, and reviewed that instead.
  > **A figure that is only true under a predicate must carry the predicate as
  > a RUNNABLE query in the same breath.**
  >
  > The collision figure was true but **unverifiable as written and its
  > denominator was stale**: "3 collisions of 364" is right only under
  > `is_active AND forked_from IS NULL`; raw it is **84**, and distinct
  > functions are **425 raw / 410 canonical as of 2026-09-03**, not 364. A
  > council reviewer reconstructed the predicate from prose, got a different
  > number, and reviewed that instead. **A figure that is only true under a
  > predicate must carry the predicate as a RUNNABLE query in the same
  > breath.**
- **⚠ ALL FOUR SEEDED KITS PIN THE CHROME THE DEFAULT ALREADY PICKS, so a kit
  delivers NO chrome differentiation** [MEASURED 2026-09-03, and it is the
  third dimension to fall the same way]. Every kit's
  `header_component_id`/`footer_component_id` points at
  `header-theme-chrome` / `footer-theme-chrome` — **which is exactly the row
  `ResolveChromeComponent` already returns for a site with no pin at all.**
  `ChromeSlotFunction()` (`component_library.go:386`) maps the slot to the
  FUNCTION `site-header`/`site-footer`, and under the POOL predicate
  (`chromeEligibleSQL`, which includes `forked_from IS NULL`) the only eligible
  row for each is `header-theme-chrome`/`footer-theme-chrome` — so
  `ResolveChromeComponent` returns exactly what the kits pin, with no tiebreak
  involved. `bugs_closed/118`
  established the same thing independently from the other direction: after
  118's fleet repoint, `GetComponentByFunction` and `ResolveChromeComponent`
  "already returned the same row for both chrome functions". **So the pins are
  no-ops.**
  > **The reason first written here was sloppy in the same way as the
  > withdrawn correction above** — it said the pins "resolve to
  > `site-header`/`site-footer`", conflating the row's `name` with its
  > `function`. The finding is unchanged; state it as row identity, not as a
  > function string. This is the same
  indistinguishability already recorded for the six pre-existing
  `style_collections.header_component_id` pins (all six point at the
  default's own pick); the kits add four more.
  ```sql
  SELECT tk.name, hc.function, hc.component_level, fc.function, fc.component_level
    FROM theme_kits tk
    LEFT JOIN content_components hc ON hc.id = tk.header_component_id
    LEFT JOIN content_components fc ON fc.id = tk.footer_component_id;
  ```
  **The consequence is the honest summary of this whole entry: of the four
  dimensions a kit bundles, three cannot change what a site looks like.**
  Palette cannot reach the stylesheet (the render overlay is spec-wins on all
  8 core slots — MEASURED at the artefact on gamedesign.uk, which resolved a
  hand-chosen palette and served none of its eight core colours);
  `page_archetypes` governs at most 1 live page in 18 (94.4% of 1,083 live
  pages match no `defaultSectionsForPage` output, so 5.6% is an UPPER bound);
  and chrome is the no-op above. **Layout is the only dimension where
  adopting a kit changes anything at all** — and see the seed-set bullet
  below, where two of the four kits pick a layout the tag matcher would have
  picked anyway.
- **the seed set: two of four kits are REDUNDANT WITH THE MATCHER, and
  `soft-editorial` is a deliberate route to an otherwise unreachable layout.**
  Because `apply_theme_kit` names a layout id and bypasses tag matching
  entirely, **a kit's marginal value is inversely proportional to how
  reachable its layout already is by tags.**
  - `tool-portal-light` — **near-worthless**: 14 sites already reach that
    layout by tags.
  - `brochure-formal` — **near-worthless and worse in kind**: it is the
    resolver's hard-fallback layout, so a kit there dresses the default up as
    a choice.
  - `soft-editorial` — **the most valuable thing in the set, and it must be
    recorded for what it is: a workaround for a tag-vocabulary defect, not a
    design choice.** [MEASURED 2026-09-03, `bugs_open/445`] it scores above
    zero on 27 of 33 sites but only at **0.50, the same-scheme bonus ALONE,
    zero tag hits**, and is one of **nine of eighteen** layouts no site's tags
    reach at all. Its tags (`wellness, lifestyle, bakery, artisan,
    personal-brand, long-form`) are a designer's industry dialect; `long-form`
    is emitted raw by **0** sites. The kit is the only existing route to that
    look.
  - `docs-sidebar` — **pre-positioned, not demanded.** 445 buckets it
    "correctly unused" (never scores above zero for any current site); that
    argument is weaker for kits specifically, since nobody's tags reaching a
    layout does not mean nobody would CHOOSE it. Speculative and cheap (an
    unselected kit is inert) — honest only because this says so.
  **The mechanism underneath is tag DIALECT, not tag count**: layouts tagged
  with form/capability words the classifier emits (`interactive-platform`,
  `tool-portal`, `publication`) get selected; layouts tagged with industry
  words (`law`, `boxing`, `bakery`) do not. `social-lobby` carries the most
  tags of any layout — 15 — and has one site. **DO NOT CURATE BY TASTE**:
  445 is building a fleet scorer as a shared package, and a kit candidate
  should be simulated against the live fleet before it is seeded. Adoption is
  0, so reseeding is free today.
- **a kit-chosen layout records NO candidate and NO fit evidence.** Fixed
  2026-09-03 (`28aeb4ca0`): the layout rung returned
  `candidates: ["<kit layout>"]`, which `install_site_composition_action.go`
  writes through as `lineage.layout_candidates` — reading as "one candidate
  was considered and won" when the matcher never ran. Now an empty slice,
  which the consumer's `len(cands) > 0` guard omits entirely, leaving
  `layout_source: 'theme_kit_default'` to carry the story. Same
  false-structured-fact class as the `layout_source: 'library_match'` defect
  fixed one field over in the same original edit. **The related blind spot is
  deliberate and unfixed:** because this arm returns before the matcher,
  DES-086's compose-time fit measurement never runs for a kit-chosen layout,
  so **a kit site's layout fit is unmeasured**. Recorded by that lane in
  DES-086's relations with no change asked here; fixing it inside the rung
  would mean scoring candidates we then ignore.
- **known gap, not yet closed**: a themed site's exact colours can still
  drift on a LATER `needs_design` pass — `design_intent.palette.
  reference_values` is correctly consulted at composition-install time but
  advisory-only at render-merge time. See RFC_059 (DRAFT,
  `architecture_review/RFC_059_...md`) — filed separately because it changes
  a shared rendering guarantee (DES-042 below), not bundled into this entry.
- **relations:** DES-003 (composition pipeline — the system this extends,
  not replaces), DES-013 (migration 025, the composable-theme lineage),
  DES-042/DES-052 (the merge-authority rule RFC_059 proposes to extend),
  `bugs_open/291` (the correct empty-`handler_agent` HITL idiom — noted for
  Phase 5's "create from example," not yet built, so a future
  `needs_theme_kit_review` item type does not repeat `fork_theme_from_site`'s
  existing phantom `theme-review-handler` mistake).
- **sources:** plan `/home/ant/.claude/plans/please-think-hard-about-starry-
  locket.md`; diagnosed/reviewed collaboratively with `site_delivery_and_
  editor`, `components`, `gap planner`, `calendar`, `webdesign-tool-
  rebuild(s)` the same session (2026-09-02).
- **verify-later:** once Phase 1 has real adoptions, `SELECT count(*) FROM
  site_specs WHERE aspect='theme_kit_adoption' AND is_current` — a durable
  zero would mean the mechanism is built but undriven (a recurring pattern
  elsewhere in this register), not evidence it works.


### DES-086 — Layout FIT evidence: `lineage.layout_match_score` (103's spec, finally written) + a library-gap signal that fires on weak tag fit
- **status:** committed 2026-09-03 (`76db94fc7`), **inert until the next chassis roll**; council `34d57f60` (submitted, verdict unread at entry time)
- **status-evidence:** `TestOneAttractorTagIsAWeakFit` reproduces designblog.co.uk's live 7% (`TagCoverage = 0.072`) and asserts magazine-grid STILL wins; guard proven by three mutations, results in the test header. Post-roll proof owed: the next composition's `resolved_composition` row carries `lineage.layout_match_score`.
- **what:** `resolveLayoutByTags` now returns `layoutFit` — `TagCoverage` = matched IDF weight / total site-term IDF weight (migration 103's "(float 0-1) tag-overlap score", specified April 2026 and never computed: 0 of 33 rows carried it), matched/unmatched terms, runner-up, margin, and the threshold in force. `install_site_composition` writes it to `lineage.layout_match_score` + `lineage.layout_fit`, persists `is_scheme_mismatch` (computed since June, never recorded), and prefers the resolver's REPORTED `source` over inferring it from `is_fallback`. The gap predicate is `LibraryGap() = IsFallback || IsSchemeMismatch || IsWeakFit` (`lmMinTagCoverage = 0.50`, chosen from the measured distribution's widest empty band 38%-62%, NOT invented), and `lineage.layout_source` is promoted to `'needs_new_layout_candidate'` on a gap over a real match. **Selection is unchanged.** No migration: the validator is permissive on unknown keys and already admits the enum value. **Why:** the old predicate could only fire at total==0; four live sites recorded `tags 0.00` WITH `library_match`; exactly TWO gap items exist across 63,007 work items ever written (29,657 live ∪ 33,350 archived), both the degenerate no-tags arm. **Pre-registered disconfirmation:** migration 734 (portfolio_positioning, 11:39:14Z) finally renders the layout taxonomy into the classifier prompt (it had been `null`), so coverage should rise fleet-wide; if compositions land inside 38%-62% the cut was a 33-site artefact — `layout_fit.threshold` per row keeps the re-derivation honest.
- **landmine:** an all-time count of `needs_new_layout_candidate` (or ANY work item) MUST union `site_work_items_archive` — the live table is a rolling window; a peer "independently verified" the wrong figure by making the same omission. `site_specs` does NOT archive (versions in place under `is_current`), so the same check is useless there.
- **sources:** `platform/orchestration/actions/fork_theme_composition.go` (`layoutFit`, `lmBuildFit`, `LibraryGap`, `GapReason`), `resolve_composition_layout_action.go`, `install_site_composition_action.go` (`buildResolvedCompositionSpec`, `readFloatFromContext`), `fork_theme_composition_fit_test.go`; `bugs_open/445` §8; `docs024_key_docs_latest/bugfix_445_layout_fit/`
- **relations:** DES-037 (the predicate this corrects); DES-087 (the archetype the evidence pointed at); DES-085 (theme-kit branch returns before the matcher — a kit site records no fit; a fleet FIT sweep keyed on `sites → style_collections → css_themes → layouts` is planned behind RFC_024's `internal/cronchecks`); RFC_037 (complementary — makes the classifier SAY something different; this makes the difference REACHABLE)
- **verify-later:** first post-roll `resolved_composition` row has `lineage.layout_match_score`; coverage distribution after 734 vs the 38%-62% band

### DES-087 — `content-hub-tools` layout: the archetype for an editorial content hub with EMBEDDED tools — tags chosen by simulation, seeded behind a reachability guard
- **status:** **APPLIED 2026-09-03** (migration 736, verified at the live row: active, light, editorial, 9 tags, 30,303 chars CSS, reachable by 14 current classification specs). **No site composed onto it yet** (owner: fix forward only).
- **status-evidence:** `SELECT is_active, cardinality(industry_tags) FROM layouts WHERE name='content-hub-tools'` → true, 9; the seven `magazine-grid` cluster sites verified UNMOVED after apply.
- **what:** 19th active layout. Category `editorial`, scheme `light`, tags `content-hub, interactive-tools, editorial-publication, long-form, long-form-content, editorial-guides, guides, research-publication, content-platform` — the FORM words the 7-site magazine-grid cluster already emitted and no layout carried. Grammar: one editorial spine (1120px), a 680px reading column narrower than either sibling, full-width TOOL SHELVES that interrupt the prose (`.tool-shelf` / `.tool-list-section`), an inline `.tool-embed` at reading width, a quiet category strip, 6px radius, hairline + lift. Frames both the sibling class contract (`.tool-card`, `.tool-workspace`, `.article-body`) and the classes the live components emit (`tl-*`, `guide-*`, `featured-article__*`, `article-card*`, `faq-*`) — the last three components carry no inline style. **Tag set chosen by SIMULATION** (scorer validated 29/30 against recorded scores): candidate A (445's title words) left designblog.co.uk at 6%; candidate B (adds the emitted form words) rescues 6 of 7 (designblog 7→16%, apis 9→19%) and pulls in gamedesign.uk + farmerinsurance.uk (both judged correct; farmerinsurance sat on industry-hub matching ZERO tags). **Not rescued:** oufe.com — its own tags lead `interactive-platform`. **The migration refuses to seed an unreachable layout** (guard: <2 current specs emit any tag → abort) because 9 of 18 existing layouts have zero sites emitting any of their tags — the inline, first-instance form of Phase 4's `assert_layout_reachable()`, which is owed as a function + `pattern-check.py` rule.
- **landmine:** a new layout is only real if the classifier already emits tags it can match; industry-dialect tags (`wellness, bakery`; `law, consultancy`) are unreachable by construction. And a rollback must DEACTIVATE, never DELETE — `css_themes.layout_id` is `ON DELETE SET NULL`.
- **sources:** `docs/agent_docs/sql_for_agents/736_layout_content_hub_tools_archetype.sql` (+`_ROLLBACK`); `bugs_open/445` §8g; `bugfix_445_layout_fit/COUNCIL_SUBMISSION_445_archetype_736.json`; simulation `scratchpad/simulate2.py` (method recorded in RUNBOOK r9)
- **relations:** DES-086 (the evidence that sized it); DES-014 (the layout library); DES-025/DES-037 (the matcher it is reachable through); DES-085 (theme kits — `soft-editorial` is reachable ONLY via a kit, the opposite failure); RFC_037 (twin-pair differentiation is positioning's job, not the layout's — both twins land here by design)
- **verify-later:** first site to compose onto it (portfolio_positioning's copyonline.co.uk is the canary — their prediction: `tool-portal-light`, or `magazine-grid` if editorial words lead); whether the 17 unbuilt remakes reach it with their REAL tags rather than the [ASSUMED] proxy
