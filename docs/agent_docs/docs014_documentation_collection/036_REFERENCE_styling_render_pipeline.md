# REFERENCE — how the chassis renders CSS, layouts and components

A reconstructed mental model of the chassis styling/render pipeline, built from reading the render-path code, the schema, and a full pull of the eighteen active layouts plus the hero/CTA/footer components (bundles 1–2 and the D1 layout dump). It exists to stop us re-deriving the same facts each session.

Two tags are used throughout: **FINDING** = verified directly from code or live data this thread; **THEORY** = inferred and still to confirm (flagged so we don't treat it as settled). Companions: `PLAN_scheme_to_components.md`, `RUNBOOK_scheme_to_components.md`, `running_notes_scheme_to_components.md`.

---

## 1. The pipeline at a glance

Two deploy paths, both ending git → GH Actions → Backblaze B2:

- **Stylesheet:** site-design-planner composes a theme → writes a `resolved_composition` pointer → `RenderCSSFromSpecAction` renders `styles.css` → `deploy_css` commits it.
- **Pages:** page-build-handler → `plan_sections` (resolve each section to a component, render it) → `CompilePageSectionsAction` (concatenate the section HTML, then inject head/header/footer) → deploy the page HTML.

The stylesheet and the page components are rendered by **separate** code paths and only meet in the browser, through shared CSS class names and CSS custom properties. Most of the light/dark trouble lives in that gap.

## 2. Composition — how scheme, palette and layout are chosen

`deriveSchemeFromDesignIntent(style_direction, suggested_style)` returns `light` / `dark` / `""`. `resolveLayoutByTags` then treats that scheme as a near-hard constraint when picking the layout. **FINDING:** `buildResolvedCompositionSpec` records `css_theme_id`, `palette_id`, `layout_id`, `typography_set_id` (plus lineage and reasoning) but **not** the scheme value — so downstream, scheme survives only on `layouts.scheme` (constrained to `light`/`dark`/`neutral`).

A `css_themes` row ties a palette + layout + typography_set together. A site reaches its theme via `sites.style_collection_id → style_collections.css_theme_id → css_themes` (→ `palettes`, `layouts`, `typography_sets`). That join, extended by one column to `layouts.scheme`, is the recovery path for "what scheme is this site".

## 3. styles.css — the three-part assembly (`RenderCSSFromSpecAction`)

1. **Layout template.** The layout's `css_template` is a Go `text/template` rendered with three FuncMap helpers — `{{palette "key" "fallback"}}`, `{{typo …}}`, `{{token …}}` — over merged maps. Merge rules (`render_css_composition_helpers.go`): palette **core** slots (`primary, secondary, accent, background, surface, text, text_muted, border`) → spec wins; palette **specialised** slots (heading, cta_bg, footer_bg, …) → theme wins; typography → spec wins; structure tokens → layout only.
2. **Component snippets.** `loadComponentCSSSnippets` appends CSS from the `css_snippets` table whose `applies_to` overlaps the site's component list.
3. **Section defaults.** `buildSectionDefaults` appends a luminance-driven `--section-*` block (see §4).

Return shape `{result: "<css>", type: "text"}` → `deploy_css`. **FINDING:** the template data map also carries `Components`, `SectionStyles`, `BackgroundIsDark`, `SurfaceIsDark`, `Spacing`, but no active layout reads any of them (see §7).

## 4. The colour system (the heart of it)

- **`:root` palette variables** (`--color-*`) are emitted by the layout template from the merged palette.
- **The Colour Inheritance Model.** Element rules are written `var(--section-*, var(--color-*))` — e.g. `color: var(--section-heading, var(--color-text))`. So a colour resolves: per-section override first, then the palette, then a literal fallback. This is the contract that lets one set of element rules serve both light and dark sections.
- **`--section-*` is a dark-context override, not a general theme.** It is only ever *set* when a background is dark. `buildSectionDefaults` (`color_util.go`): returns an empty string unless `bgIsDark || surfaceIsDark`; when the whole-site background is dark it emits a `body { --section-text/-muted/-heading/-surface/-border }` block; when the surface is dark it emits the same block scoped to five fixed classes (`.features-section, .services-section, .differentiators-section, .about-section, .faq-section`). **FINDING:** on a fully light site it emits nothing, so every element falls back to `--color-*` — dark ink on light ground. That is why `tool-portal-light` renders correctly with no special handling.
- **Palette-aware text picking.** `pickReadableOnBackground` walks a preference order (`background, text_muted, text, accent, secondary, primary`) and returns the first colour clearing a WCAG contrast ratio against the section background (`isDarkHex`, `wcagContrastRatio`, loose thresholds — body 3.0, heading 2.0 — to preserve palette character over clinical white). `--section-surface`/`--section-border` are emitted as fixed `rgba(255,255,255,…)` values, not palette-picked.

## 5. The component model (page sections)

- A section component is an `html_template` (Go template) carrying an inline `<style>`. `content_components.component_level` ∈ `section / header / footer / element / tool`.
- **Resolution.** `plan_sections` resolves a section to a component by **direct function lookup** — one active component per function. **FINDING:** C4 showed no section function has more than one active component, so the convention holds even though the unique index only covers `component_level='tool'`. The scoring `component_selector` exists but is not on this path (§7).
- **The `{function}-section` class contract.** The stylesheet (the layout's structural rules, and the five surface classes in both the layout and `buildSectionDefaults`) targets `.{function}-section` — `hero → .hero-section`, `call-to-action → .call-to-action-section`, etc. **FINDING:** the five surface sections and `footer-with-disclaimer` (`.footer-with-disclaimer-section`) honour it; `hero` emits `.hero` and `call-to-action` emits `.cta-section`, so the layout's structural rules and any per-section override miss them.
- **The Dark Section Variable Contract (003).** Hero / CTA / testimonials backgrounds are **component-owned** — the layouts deliberately do not paint them. The five surface classes are **layout/renderer-owned** (`background: var(--color-surface)` in every layout). A section that is intrinsically dark is expected to set `--section-*` on its own container; well-behaved components then read `var(--section-*)` for their children.
- **Self-styling vs consuming — the bug.** **FINDING:** `hero` and `call-to-action` hardcode a dark background (a gradient/solid built from `--color-primary`) plus a dark `--section-*` block plus literal white text, unconditionally. `footer-with-disclaimer` consumes `var(--section-*)` for its children (good) but still self-declares the dark *values* at the top (plus `--color-footer-bg: #1a1a2e`). None of these adapt to a light site. C2: ~40 active sections carry `is_dark_section = t` and most self-declare `--section-*`.

## 6. Header / footer / head (the site-component path)

Separate from page sections. Rendered by `RenderHeader` / `RenderFooter` / `RenderFallbackHeader` / `RenderFallbackFooter` (`component_library.go`) and injected into the page HTML by `InjectHeader` / `InjectFooter` / `InjectHead`, called from `CompilePageSectionsAction`.

- **Selection (`RenderHeader`/`RenderFooter`).** **FINDING:** fetch the site's `StyleCollection` (by site id, else by domain); if `coll.HeaderComponentID != nil` use `GetComponentByID`; else `GetComponentByFunction("site-header")`; else `RenderFallbackHeader`. So the operative store is **`style_collections.header_component_id` / `footer_component_id`**.
- **The fallback is hardcoded dark.** **FINDING:** `RenderFallbackHeader`/`Footer` emit `.site-header`/`.site-footer` with `background:<primary>` and white text (`color:#fff`, nav `rgba(255,255,255,.9)`), no scheme awareness, using the legacy `renderCtx.PrimaryColor`/`AccentColor` fields. Any site that reaches the fallback gets dark chrome.
- **`InjectHeader`** skips injection if the page HTML already contains `class="site-header"`, refreshes nav items from deployed pages, renders via `RenderHeader`, then regex-replaces any existing header (and its trailing `<style>`/`<script>`) and inserts after `<body>`.
- **Four overlapping default stores** (the F tangle). **FINDING (they exist):** `style_collections.header_component_id`/`footer_component_id` (read by `RenderHeader`); the `site_components` table slots `header`/`footer`/`head` (idea.uk's point to **inactive** rows, distinct from the active `site-header`/`site-footer`); `sites.default_components` JSONB (written by `UpdateSiteDefaultsAction`); `layouts.default_header_component_id`/`default_footer_component_id` (exist, all NULL). **THEORY:** the intended design is *layout declares a default → composition copies it into the style collection / `sites.default_components` via `update_site_defaults` → `RenderHeader` reads it*; today the layout defaults are NULL, `RenderHeader` reads `style_collections`, and it is unconfirmed whether the composition path runs `update_site_defaults` at all. `site_components` for header/footer may be a superseded store. Confirm in F.

## 7. What is and isn't still used

**Live (on the current render path):**
- The layout `css_template` — `:root` palette, element rules, structure, and the five-surface-class painting.
- `buildSectionDefaults` — the renderer-owned `--section-*` emitter (body + the five surface classes).
- `css_snippets` append — though **FINDING:** of 21 snippets, none set `--section-*`; they are animation/button/card utilities plus two news-grid templates (a couple hardcode hex). Not a factor in the scheme problem.
- Direct-function component resolution in `plan_sections`.
- The `{function}-section` contract — for the five surface sections and `footer-with-disclaimer`.
- `RenderHeader`/`Footer` via `style_collections.*_component_id`, with the function-name and hardcoded fallbacks.

**Dead, vestigial, or not on the current path:**
- **`SectionStyles` / `buildCSSsectionStyles`.** **FINDING:** computed and passed in the template data, but no active layout references `{{range .SectionStyles}}`; likewise `.Components`/`.BackgroundIsDark`/`.SurfaceIsDark`/`.Spacing` are unused by active layouts. So the per-section style list is dead for current sites — the live per-section mechanism is `buildSectionDefaults` plus component inline CSS. The fix extends `buildSectionDefaults`; it does not wire `SectionStyles`.
- **`component_selector` scoring.** Exists, but current sections resolve by direct function lookup, so the scorer is not exercised. **THEORY:** kept for a future many-candidates case; not relevant to the scheme fix.
- **Legacy variable names** (`--primary-color`/`--accent-color`, which fall back to navy because the renderer emits `--color-primary`/`--color-accent`). **FINDING:** present in the *inactive* hero twin. **THEORY:** a few active components (social-proof, testimonials, portfolio-showcase, contact-form, tool-cta) trip the coarse C2 flag, but at least the CTA trips it only on the benign `--color-white`; whether any active component truly uses the navy-fallback names needs a per-template check.
- **`site_components` slots for header/footer.** **THEORY:** an older/overlapping store; `RenderHeader` reads `style_collections` instead, and idea.uk's slots point to inactive rows — likely superseded for header/footer selection. Confirm in F.
- **`sites.default_components` + `layouts.default_*_component_id`.** **THEORY:** an intended default-wiring mechanism that is currently unwired (layout columns NULL; `RenderHeader` doesn't read `sites.default_components` on the path read so far). Confirm in F.
- **The "TEMPORARY RENDERER COUPLING".** The five surface class names are duplicated in `buildSectionDefaults` and hardcoded in every layout. Live, but explicitly slated for removal by **Phase 4.5** of `025_palette_layout_typography_migration` (move surface painting into components; switch the renderer to a `data-section-bg` attribute selector). The scheme fix is the natural place to generalise Phase 4.5 to cover scheme and the hero/CTA.

## 8. Why a light site still renders dark chrome (the B/G/Q4 summary)

Scheme is decided, used to pick the layout, then dropped at the component render context (`RenderContext` has no scheme field). It still reaches `styles.css` as the `:root` palette and, implicitly, through the luminance-driven `buildSectionDefaults`. **So the scheme already reaches components implicitly, via the palette and luminance — the components defeat it by hardcoding.** On `tool-portal-light` the palette resolves `--color-primary` to `#1a1a1a`, so the hero and CTA (which paint their backgrounds from primary) come out dark, and the header/footer come from a pinned dark component or the hardcoded-dark fallback. The result is dark chrome over light content. The fix (Q4 = option a): components adopt `{function}-section` and become strictly variable-driven (consume `var(--section-*, var(--color-*))` and the palette; self-declare nothing), and the renderer (extended `buildSectionDefaults`) owns the per-section `--section-*` for sections that are dark by intent — no `*-light`/`*-dark` duplication.

## 9. Open, pointers to the next bundle

- **E (section-contrast model):** where a per-section dark/contrast intent is stored (the site plan? composition?), how `is_dark_section` is set at component creation, and whether the renderer should set each dark section's *background* as well as its `--section-*` text (else background-from-component and text-from-renderer can desync).
- **F (header/footer):** untangle the four default stores, confirm whether composition runs `update_site_defaults`, and make the fallback and the header/footer components scheme-aware.
