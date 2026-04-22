# Palette, Layout, Typography — Composable Theme Migration

## Splitting `css_themes` into independently-versioned entities

---

## 1. What This Project Is

Today `css_themes.css_template` mixes three distinct concerns into one row: palette, typography, and layout. This has been masked by a silent fallback where only `standard-brochure` has a populated `css_template` and every other theme collapses to it. The 13 non-brochure themes exist as palette metadata only; selecting `bakery` or `dark-modern` or `soft-editorial` today produces the same `standard-brochure` layout with a slightly different `:root` block.

The migration splits these three concerns into three tables (`palettes`, `layouts`, `typography_sets`), each independently versionable with the same lineage model we added to `css_themes` and `style_collections`. A `css_themes` row becomes a composition — a named bundle pointing at one palette, one layout, one typography set.

This enables:

- **Genuinely different visual identities per theme.** A site picking `boxing` gets an aggressive, kinetic layout. A site picking `soft-editorial` gets a calm, reading-first layout with generous line-heights and serif display. A site picking `industry-hub` gets a directory-style vertical information layout. These aren't variations of the same layout with different colours — they're structurally different.
- **Adoption forks that capture what's distinctive.** Adopting a design studio portfolio produces a new layout the library didn't have. Adopting a small-business site usually just produces a new palette and reuses an existing layout. Forking is granular.
- **Composable selection.** The selector eventually picks "warm palette + content-heavy layout + editorial typography" from independently-scored options rather than one bundled row that approximately matches.

The current theme library is effectively one layout with 14 palette skins. Post-migration, the library holds a diverse set of layouts, each pairable with any palette, producing orders of magnitude more possible visual identities.

---

## 2. Scope Decisions

### Layout diversity

We are **not** producing archetype layouts that each cover a category of themes. We are producing layouts that reflect genuinely different user intents and content types. The axis of variation spans visual character, content-type dominance, and commercial intent.

### Layouts are not palette-bound

A layout references `{{palette "primary"}}`, `{{palette "hero_title"}}`, etc., but doesn't care which palette fills those slots. The same portfolio-kinetic layout works with a muted cool palette, a warm earthy one, or a stark black-and-white one.

### Layouts may have preferred header/footer components

Different layouts often need structurally different headers and footers: a comparison-aggregator needs a header with a prominent search input; an ecommerce-storefront needs a header with a cart icon; a docs-sidebar needs a fixed left nav. These are structural differences, not stylistic variations — they need different HTML, which means different `content_components`.

`layouts` carries `default_header_component_id` and `default_footer_component_id` columns as nullable FKs into `content_components`. When a `style_collections` row is created for a layout without specifying its own header/footer, the layout's defaults apply. Style_collections can still override (any header can pair with any layout) — the layout's defaults are recommendations, not constraints.

### We are not touching style_collections structure

`style_collections` remains a bundle of `css_theme_id` + `header_component_id` + `footer_component_id`. Those relationships stay. The lineage columns we added to `style_collections` stay. Only the entity `style_collections.css_theme_id` points at gets internally restructured.

### We are not changing the selector in this migration

`SelectStyleCollectionAction` still picks a `style_collections` row and follows its `css_theme_id` FK to a `css_themes` row. The renderer now loads palette/layout/typography from that row, but the selection logic is unchanged. Composable selection ("pick a warm palette, pick an editorial layout independently") is a separate project.

### Cutover is direct, not shadow-mode

The 12 non-`standard-brochure` themes silently fall back to `standard-brochure` today. Currently-live sites are all on `professional-dark` → `standard-brochure` layout in practice. With seeded data in place, the new renderer should produce identical output for live sites (same layout, same palette, just sourced differently). Shadow mode would be ceremony without safety value. Rollback if needed is a code revert — the legacy columns stay populated until Phase 7.

---

## 3. Schema Shape

### New table: `palettes`

```sql
CREATE TABLE palettes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text UNIQUE NOT NULL,
    display_name text NOT NULL,
    colours jsonb NOT NULL,
    category text,
    industry_tags jsonb DEFAULT '[]'::jsonb,
    is_active boolean NOT NULL DEFAULT true,
    origin text NOT NULL DEFAULT 'seed'
        CHECK (origin IN ('seed', 'handcrafted', 'adopted', 'fork_of_adopted')),
    needs_review boolean NOT NULL DEFAULT false,
    forked_from_palette_id uuid REFERENCES palettes(id) ON DELETE SET NULL,
    source_site_id uuid REFERENCES sites(id) ON DELETE SET NULL,
    source_domain text,
    forked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);
```

The `colours` JSONB is the full palette map. Any slot names a layout template needs — `primary`, `hero_title`, `cta_bg`, `footer_text`, arbitrary future additions — live here. No struct field needs to be added when a palette exposes a new slot.

### New table: `layouts`

```sql
CREATE TABLE layouts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text UNIQUE NOT NULL,
    display_name text NOT NULL,
    css_template text NOT NULL,
    structure_tokens jsonb NOT NULL DEFAULT '{}'::jsonb,
    category text,
    industry_tags jsonb DEFAULT '[]'::jsonb,
    description text,
    default_header_component_id uuid REFERENCES content_components(id) ON DELETE SET NULL,
    default_footer_component_id uuid REFERENCES content_components(id) ON DELETE SET NULL,
    is_active boolean NOT NULL DEFAULT true,
    origin text NOT NULL DEFAULT 'seed'
        CHECK (origin IN ('seed', 'handcrafted', 'adopted', 'fork_of_adopted')),
    needs_review boolean NOT NULL DEFAULT false,
    forked_from_layout_id uuid REFERENCES layouts(id) ON DELETE SET NULL,
    source_site_id uuid REFERENCES sites(id) ON DELETE SET NULL,
    source_domain text,
    forked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);
```

`structure_tokens` holds container widths, section padding scales, border radii, shadow tokens. These are visible to layout templates without inflating the palette.

### New table: `typography_sets`

```sql
CREATE TABLE typography_sets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text UNIQUE NOT NULL,
    display_name text NOT NULL,
    fonts jsonb NOT NULL,
    scale jsonb NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    origin text NOT NULL DEFAULT 'seed'
        CHECK (origin IN ('seed', 'handcrafted', 'adopted', 'fork_of_adopted')),
    forked_from_typography_set_id uuid REFERENCES typography_sets(id) ON DELETE SET NULL,
    source_site_id uuid REFERENCES sites(id) ON DELETE SET NULL,
    source_domain text,
    forked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);
```

`fonts` — body/heading/display font stacks. `scale` — base size, line heights, heading size ratios.

### Changes to `css_themes`

Three new FK columns; no columns dropped in this phase:

```sql
ALTER TABLE css_themes
    ADD COLUMN palette_id uuid REFERENCES palettes(id),
    ADD COLUMN layout_id uuid REFERENCES layouts(id),
    ADD COLUMN typography_set_id uuid REFERENCES typography_sets(id);
```

After migration every row has the three FKs populated. `css_template`, `css_content`, `color_palette`, `typography` remain as legacy columns until a later cleanup phase confirms nothing reads them.

### Indexes

Partial indexes on the lineage and review columns, matching the pattern already established:

```sql
CREATE INDEX idx_palettes_forked_from ON palettes(forked_from_palette_id) WHERE forked_from_palette_id IS NOT NULL;
CREATE INDEX idx_palettes_needs_review ON palettes(needs_review) WHERE needs_review = true;
CREATE INDEX idx_layouts_forked_from ON layouts(forked_from_layout_id) WHERE forked_from_layout_id IS NOT NULL;
CREATE INDEX idx_layouts_needs_review ON layouts(needs_review) WHERE needs_review = true;
CREATE INDEX idx_typography_sets_forked_from ON typography_sets(forked_from_typography_set_id) WHERE forked_from_typography_set_id IS NOT NULL;
```

---

## 4. Template Data Shape (the map-based redesign)

### The problem with the current struct

`cssTemplateData` today has hardcoded fields: `Primary string`, `Secondary string`, etc. Adding a new palette slot requires a Go code change. The palette tables in each of the 13 existing themes already contain slots the struct doesn't know about (`hero_title`, `cta_bg`, `footer_text`, and more). Rendering them today would produce empty strings in those slots.

### The new shape

```go
type cssTemplateData struct {
    Palette      map[string]string
    Typography   map[string]string
    Structure    map[string]string

    // Non-palette derived data stays as before
    Components       []string
    SectionStyles    []sectionStyleEntry
    BackgroundIsDark bool
    SurfaceIsDark    bool
}
```

Layout templates reference variables via map access with a fallback:

```css
:root {
  --color-primary: {{palette "primary" "#1a365d"}};
  --color-hero-title: {{palette "hero_title" "#0f172a"}};
  --font-body: {{typo "font_family" "system-ui, sans-serif"}};
  --shadow: {{token "shadow" "0 2px 4px rgba(0,0,0,0.1)"}};
}
```

### Template helper: safe lookup with defaults

Go templates don't fail on missing map keys; they emit empty strings. A layout referencing `{{.Palette.hero_title}}` against a palette that doesn't declare `hero_title` would produce broken CSS. Template helper funcs enforce safe lookup:

```go
funcMap := template.FuncMap{
    "palette": func(key, fallback string) string { ... },
    "typo":    func(key, fallback string) string { ... },
    "token":   func(key, fallback string) string { ... },
}
```

Layouts use `{{palette "primary" "#1a365d"}}` — if the palette has `primary`, it wins; otherwise the fallback applies. Safe, explicit, no silent empty-string bugs. Every layout template is expected to pass a fallback for every lookup.

---

## 5. The Palette Merge Rule

This is the single most important piece of rendering logic in the new system.

When a site selects a theme, the renderer has two palette sources:

- **The theme's palette** (from `css_themes.palette_id → palettes.colours`) — curated, complete, carries the theme's character
- **The design spec's palette** (from the webdesign-agent's `analyze_design` output) — may be partial, reflects the site's identity

The rule: **spec wins for core palette slots; theme wins for specialised slots.**

**Core palette slots** (spec wins where present):
`primary`, `secondary`, `accent`, `background`, `surface`, `text`, `text_muted`, `border`

These are the slots a site's identity owns. If a site has a brand primary of `#6366f1`, that primary applies regardless of which theme is selected.

**Specialised slots** (theme wins):
`primary_hover`, `primary_text`, `secondary_hover`, `secondary_text`, `heading`, `header_bg`, `header_text`, `hero_title`, `hero_subtitle`, `card_bg`, `cta_bg`, `cta_text`, `footer_bg`, `footer_text`, and any future additions the theme exposes.

These are the slots the theme owns. A theme like `bakery` has curated them to work together — the cream header on chocolate, the warm gradient CTA. Site-level overrides of these would require the site to have curated equivalents, which it rarely does.

**Missing slots** fall through to renderer defaults (same as today's hardcoded fallbacks).

### Why this shape

When a site adopts a theme without a strong palette identity of its own (new builds, most cases), the theme fills everything and the site looks coherent. When a site has a specific brand identity, the core palette injects cleanly while the theme's specialised character preserves. Selection quality matters — a dramatically mismatched spec against a theme still produces visible clash in the specialised slots — but that's a selection problem, not a rendering one.

### Documented constraint

The contract doc will note: "a site selecting a theme is committing to that theme's specialised character. If the site's identity palette clashes with the theme's specialised slots, either select a different theme or ask the webdesign-agent to generate a fresh theme." This is an honest statement of the tradeoff.

---

## 6. Renderer Flow (new path)

```
RenderCSSFromSpecAction:

1. Resolve css_themes row
   (via sites.style_collection_id → style_collections.css_theme_id → css_themes)

2. Load palette          via css_themes.palette_id → palettes.colours
3. Load layout           via css_themes.layout_id → layouts.css_template + structure_tokens
4. Load typography_set   via css_themes.typography_set_id → typography_sets.fonts + scale

5. Merge palette (theme + spec)
   - Start with theme.colours
   - Overlay design_spec.color_scheme values for core slots
   - Leave specialised slots from theme

6. Build cssTemplateData with Palette, Typography, Structure maps

7. Parse layout.css_template with template funcs (palette, typo, token)

8. Render

9. Append component snippets (unchanged)

10. Append renderer-enforced --section-* defaults (unchanged,
    using pickReadableOnBackground against the merged palette)
```

Steps 1-5 are new; steps 7-10 are modified template data; steps 9-10 unchanged.

---

## 7. The Layouts to Build

Fifteen initial layouts, characterised below. Each produced as a full CSS template with palette/typography/structure references via template helper funcs.

| Layout name | Character | Example site types |
|---|---|---|
| `brochure-formal` | Structured, understated, CTA-driven sections | Consultancies, law, finance, B2B services |
| `brochure-bold` | Large hero, gradient accents, strong CTAs, higher motion | Tech startups, fitness businesses, sales-led |
| `portfolio-kinetic` | Asymmetric sections, large display type, motion, negative space | Design studios, agencies, creative portfolios (the "vonc" pole) |
| `magazine-grid` | Article cards, sidebar, content-dominant | Publications, news sites, opinion long-form |
| `utility-tool` | Minimal chrome, centered tool area, dense operational feel | Dev tools, calculators, SaaS product utilities |
| `media-grid` | Thumbnail grid, continuous scroll, media-first | Video, audio, media libraries (the "youtube" pole) |
| `docs-sidebar` | Left sidebar nav, anchored content, code-friendly | Developer docs, knowledge bases, API references |
| `soft-editorial` | Serif display, relaxed padding, paper-textured background, pill buttons | Wellness, lifestyle, thoughtful long-form |
| `technical-precise` | Clean sans-serif, tight radii, subtle shadows, glass headers | SaaS, infrastructure, engineering products |
| `high-energy` | Uppercase display, hard edges, diagonal splits, strong accents | Combat sports, fitness, events (the "boxing" pole) |
| `comparison-aggregator` | Hero with prominent search, filter bar, result grid, guide cards, heavy disclosure footer | Price/service comparison sites (the "vetcomparison" pole) |
| `affiliate-hub` | Hero, category index, product cards with affiliate CTAs, review blocks, disclosure banners | Product review sites, buyer's guides |
| `ecommerce-storefront` | Hero, category grid, featured products, cart-ready patterns | Retail storefronts, marketplaces |
| `tool-first-landing` | Minimal preamble, tool dominates the fold, supporting content below | Calculators, simulators, single-purpose tools (the "thunder compute" pole) |
| `industry-hub` | Industry authority positioning, directory/listing as primary element, guide index, news area, heavy disclaimer footer | Vertical information hubs not operated by an industry participant (the "gas wholesalers" pole) |

### What constitutes a "full layout template"

Each layout produces CSS covering:

- `:root` variable block (receives palette + typography + structure via helpers)
- Body and base typography rules (using the Colour Inheritance Model: `var(--section-*, var(--color-*))` fallbacks)
- Site header structure
- Site footer structure
- Section container patterns (`.hero-section`, `.features-section`, `.call-to-action-section`, etc.)
- Button styles (primary, secondary, sizes)
- Form input styles
- Responsive breakpoints
- Any layout-distinctive patterns (glass blur, gradient overlays, asymmetric grids, sidebar structure, directory grids, etc.)

Each layout is roughly 250-400 lines of CSS. Each reviewed against:
- Colour Inheritance Model (base typography uses `var(--section-*, var(--color-*))` fallbacks)
- Dark Section Variable Contract (section defaults come from the renderer, not the template)
- Template helper fallback pattern (every `{{palette ...}}` has a default)

### Default header/footer components per layout

Some layouts work with the existing generic header/footer components; others need structurally different ones. The initial mapping:

| Layout | Default header | Default footer |
|---|---|---|
| `brochure-formal`, `brochure-bold`, `soft-editorial`, `technical-precise` | `header-standard` (existing) | `footer-4-column` (existing) |
| `portfolio-kinetic`, `high-energy` | Existing minimal header | Existing minimal footer |
| `magazine-grid`, `industry-hub` | New: `header-with-categories` (top strip of category links) | `footer-with-disclaimer` (new for industry-hub) |
| `utility-tool`, `tool-first-landing` | New: `header-minimal-tool` (just logo + optional menu) | Existing minimal footer |
| `media-grid` | New: `header-with-search` (prominent search) | Existing minimal footer |
| `docs-sidebar` | New: `header-docs` (sidebar-aware, mobile hamburger) | Existing minimal footer |
| `comparison-aggregator` | New: `header-with-search` | `footer-with-disclaimer` |
| `affiliate-hub`, `ecommerce-storefront` | New: `header-with-cart-or-nav` | `footer-4-column` |

New header/footer components are noted but not in scope for this migration's Phase 1 — they land as component work alongside the layouts they belong to. Layouts that need new components will reference them as `default_header_component_id = NULL` initially and be updated when the components land.

### Mapping existing themes to initial layouts

| Theme | Initial layout assignment |
|---|---|
| `default`, `standard-brochure`, `professional-dark` | `brochure-formal` |
| `bold-conversion`, `tech`, `dark-modern` | `brochure-bold` |
| `bakery`, `warm-friendly` | `soft-editorial` |
| `calm-minimal` | `soft-editorial` |
| `soft-editorial` | `soft-editorial` |
| `boxing` | `high-energy` |
| `premium-elegant` | `technical-precise` (with serif typography) |
| `content-modern` | `magazine-grid` |
| `modern-engineering-clean` | `technical-precise` |

The theme name and the layout name don't have to match. A theme is a bundle; the layout inside the bundle is just one of three ingredients.

### Layouts with no current theme mapping

`portfolio-kinetic`, `utility-tool`, `media-grid`, `docs-sidebar`, `comparison-aggregator`, `affiliate-hub`, `ecommerce-storefront`, `tool-first-landing`, `industry-hub` — all ship in the library with no existing theme assigned. They're there for the selector to reach when appropriate, and for adoption to fork into when an adopted site matches their pattern.

---

## 8. Typography Sets

Six shared typography sets cover the palette of needs across the 15 layouts. Each is referenced by multiple themes. Forking only happens when an adopted site uses a font stack genuinely outside this set.

| Name | Character | Default fonts |
|---|---|---|
| `sans-modern` | Clean, neutral, reads well at any size | Inter + system fallbacks |
| `serif-editorial` | Warm, reading-first, magazine feel | Merriweather + Georgia fallbacks |
| `display-bold` | High-impact, condensed, uppercase-friendly | Impact / Archivo Black + sans fallbacks |
| `mono-technical` | Code-friendly, docs, utilitarian | IBM Plex Mono + system-ui for body |
| `serif-classical` | Formal, elegant, luxury feel | Cormorant Garamond + Georgia fallbacks |
| `sans-friendly` | Rounded, approachable, conversational | Nunito + Segoe UI fallbacks |

Each has a `scale` JSONB with base size, line height, and heading ratio (e.g. `{"base_size": "16px", "line_height": "1.6", "h1_ratio": "2.5"}`).

### Mapping layouts to typography sets

Most layouts map cleanly to one typography set by default. Themes override when appropriate.

| Layout | Default typography |
|---|---|
| `brochure-formal`, `technical-precise`, `utility-tool`, `tool-first-landing` | `sans-modern` |
| `magazine-grid`, `soft-editorial`, `industry-hub` | `serif-editorial` |
| `high-energy`, `brochure-bold` | `display-bold` |
| `docs-sidebar` | `mono-technical` |
| `portfolio-kinetic` | `sans-modern` (may pair with `serif-classical` in specific themes) |
| `media-grid` | `sans-modern` |
| `comparison-aggregator`, `affiliate-hub`, `ecommerce-storefront` | `sans-modern` |

---

## 9. Migration Phases

### Phase 1 — Design and draft the 15 layouts

Single batch, all 15 layouts produced together so the shared structural vocabulary (section class names, responsive breakpoint philosophy, accessibility behaviour) stays consistent.

Per layout:
- CSS template written against palette/typography/structure helpers
- Matched against Colour Inheritance Model and Dark Section contract
- Sample render against a known-good palette to verify output
- Default header/footer component decisions captured (even if some components land later)

Output: 15 reviewed CSS templates with structure_tokens definitions.

### Phase 2 — Schema migration (additive, reversible)

One migration SQL file:
- Create `palettes`, `layouts`, `typography_sets` tables with lineage columns
- Add `default_header_component_id`, `default_footer_component_id` FK columns to `layouts`
- Add `palette_id`, `layout_id`, `typography_set_id` columns to `css_themes` (nullable initially)
- Add the indexes listed in section 3

Nothing reads the new tables; nothing breaks if this goes wrong.

### Phase 3 — Seed the new tables

- Extract each of the 14 existing `css_themes` rows' palette data into `palettes` rows
- Seed the 6 `typography_sets` rows
- Seed the 15 `layouts` rows from Phase 1 output
- Populate `css_themes.palette_id`, `css_themes.layout_id`, `css_themes.typography_set_id` per the mapping in section 7

After this phase: all `css_themes` rows have three FKs populated; new tables are populated; nothing reads any of this yet.

### Phase 4 — Renderer path change and cutover

- Implement new `cssTemplateData` map shape
- Add template helper funcs (`palette`, `typo`, `token`)
- Rewrite `RenderCSSFromSpecAction` to load palette + layout + typography, run palette merge, render
- Deploy
- Trigger re-render on live sites; verify no visual regression
- If regression: revert code (legacy columns still populated, old path still works)

### Phase 5 — Adoption-fork action update

`fork_theme_from_site` changes to produce granular forks:
- Always fork a `palettes` row
- Optionally fork a `layouts` row if the adopted site's structural CSS is genuinely novel (LLM or human judgement, initially a flag)
- Optionally fork a `typography_sets` row if the fonts are distinctive
- Optionally fork header/footer `content_components` if the layout fork's header/footer don't match any existing component
- Create the `css_themes` bundle row pointing at whichever rows were chosen (new or reused)
- HITL review work items become richer — separate review for palette, for layout if new, and for the bundle as a whole

### Phase 6 — Contract doc rewrite

Replace the current "CSS Theme Template Contract" section in `003_contracts_and_standards.md` with:
- Palette contract (what slots, naming conventions)
- Layout contract (what variables a layout references, responsibility split, default header/footer FKs)
- Typography set contract
- Theme bundle contract (composition row)
- Palette merge rules (core vs specialised slots)
- Forking rules for each entity
- Review gates

### Phase 7 — Legacy column cleanup (deferred)

Once confident nothing reads them, drop from `css_themes`:
- `css_template`
- `css_content`
- `color_palette`
- `typography`

Happens after a stabilisation period (two weeks post-cutover with no regressions). Preceded by a grep pass across the codebase for any remaining references.

---

## 10. Downstream Dependencies

Some layouts reference directory/listing content patterns that don't exist as components yet. Specifically:

- `industry-hub` wants a `directory-listing` section component (supplier/provider grid)
- `comparison-aggregator` wants a `filtered-result-grid` component (sortable, filterable cards)
- `affiliate-hub` wants a `product-card-with-cta` component
- `ecommerce-storefront` wants a `product-grid` component

These are not blocking the layout work — layouts can host whatever components the existing library offers, and the missing components become `needs_new_component` work items naturally. The layouts ship first; the specific components they'd ideally pair with follow.

---

## 11. Non-Goals for This Migration

- **Composable selection.** The selector continues to pick `style_collections` rows and follow FKs. Picking palette + layout + typography independently is a later project.
- **LLM-driven layout generation.** Layouts are hand-designed in Phase 1.
- **Automatic fork-quality scoring.** The review gate is human-driven.
- **Theme preview UI.** A dashboard tool that renders a theme with a sample page for review would be useful but isn't in scope.
- **Retroactive re-selection of live sites.** Sites keep what they have. A future project may offer "re-theme this site" as an explicit action.
- **The new directory/listing components** that `industry-hub`, `comparison-aggregator`, `affiliate-hub`, `ecommerce-storefront` want. They ship separately.

---

## 12. Risks

**Visual regression at cutover.** Live sites all render through `standard-brochure` layout today. The new renderer produces identical output for them after seeding. Rollback is a code revert — the legacy columns stay populated.

**Palette merge producing clashing colours.** Documented constraint; selection quality matters. Surface this in the contracts doc and in the HITL review UI for forked themes.

**Layout template errors producing broken CSS.** Each layout reviewed manually against contracts before seeding. Template helper funcs (`palette`, `typo`, `token`) prevent silent empty-string substitution.

**Adoption forking creating library noise.** Review gate prevents unreviewed forks from being selected. Over time, a periodic audit may prune rarely-selected forks — separate work.

**Old-column reads we missed.** Phase 7 (cleanup) gated on confidence that nothing reads the legacy columns. A grep pass across the codebase before dropping them.

**Missing directory/listing components for some layouts.** `industry-hub` and `comparison-aggregator` will render with placeholder sections initially. Acceptable — the layouts are visually complete even without the specific components, and sites on those layouts can use existing grid/card components in the interim.

---

## 13. What Needs Deciding Before We Start

All resolved. Proceeding in order:

1. ~~Final layout list~~ — 15 layouts, listed in section 7 with `industry-hub` added for vertical-content sites.
2. ~~Layout-drafting approach~~ — all 15 drafted in one batch for consistent structural vocabulary.
3. ~~Typography granularity~~ — 6 shared typography sets, listed in section 8.
4. ~~Header/footer handling~~ — `layouts` has `default_header_component_id` and `default_footer_component_id` FKs. Layouts ship with existing components as defaults where possible; new header/footer components land alongside their layouts in follow-on work.
5. ~~Shadow-mode duration~~ — no shadow mode. Direct cutover with code-revert rollback.

Starting on Phase 1.
