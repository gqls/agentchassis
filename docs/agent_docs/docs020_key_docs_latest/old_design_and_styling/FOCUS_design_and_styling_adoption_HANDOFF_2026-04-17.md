# HANDOFF: Design & Styling — Composable Theme System Migration

## Date: 2026-04-17
## Previous session transcript: `/mnt/transcripts/2026-04-16-15-30-08-theme-system-migration-plan.txt`

---

## 1. What We're Doing

Splitting `css_themes` (which today conflates palette, typography, and layout into one row) into three independently-versioned tables: `palettes`, `layouts`, `typography_sets`. A `css_themes` row becomes a composition — a named bundle pointing at one palette, one layout, one typography set.

The full plan is in `025_palette_layout_typography_migration.md` (in project knowledge and `/mnt/user-data/outputs/`). 15 layouts, 6 typography sets, 14 palettes extracted from existing themes. The plan has 8 phases. **Phase 1 is next: designing and writing the 15 layout CSS templates.**

---

## 2. What's Been Completed

### Deployed to production
- **Deactivated `content-block-leadership` component** (was fabricating team member names)
- **Expanded placeholder patterns in `validate_page_content.go`** — added "human review required", "X needed" variants
- **`color_util.go`** — extracted WCAG helpers (`parseHexColor`, `sRGBToLinear`, `relativeLuminance`, `wcagContrastRatio`, `isDarkHex`, `hexToRGBA`, `pickReadableOnBackground`)
- **`render_css_from_spec_action.go`** — `buildSectionDefaults` appends palette-aware `--section-*` defaults to rendered CSS based on background/surface luminance; themes no longer declare these themselves
- **`css_templating.go`** — `TemplateCSSFromSpec` converts rendered CSS to Go-templated form by replacing hex values in `:root` with `{{.Primary}}` etc.; `cssRepl` type at package scope; `sortReplacementsByLengthDesc` sorts by length descending
- **Lineage columns added** to `css_themes` and `style_collections`: `forked_from_theme_id`, `forked_from_collection_id`, `source_site_id`, `source_domain`, `forked_at`, `origin` enum, `needs_review` boolean

### Ready for deployment (files in outputs / discussed but not yet committed)
- **`fork_theme_from_site_action.go`** — forks an adopted site's CSS into the theme library; runs inside the webdesign-agent workflow (not the orchestrator); never fails the parent workflow; creates theme + collection + HITL work item in one transaction
- **Registry entry** for `fork_theme_from_site` (category: site, IsLocal: true)
- **Webdesign-agent workflow SQL** — adds `check_should_fork` conditional + `fork_theme` step after `update_site`. Uses `default_config` (not `config`), column `type` (not `agent_type`), `AND deleted_at IS NULL` filter, `updated_at = NOW()`. Backup table: `agent_def_webdesign_backup_20260416`.

### Deployed today (this session)
- **`system.internal` site row** — placeholder for library-level work items, `status='system'`, `skip_deploy`, `skip_build`
- **10 library components** inserted directly into `content_components`:
  - 5 headers: `header-with-categories`, `header-minimal-tool`, `header-with-search`, `header-docs`, `header-with-cart-or-nav`
  - 1 footer: `footer-with-disclaimer`
  - 4 section components: `directory-listing`, `filtered-result-grid`, `product-card-with-cta`, `product-grid`
  - All follow contracts: kebab-case function, `data-component` attr, scoped CSS, Colour Inheritance Model, Input Schema v2, JS in IIFE, `created_from='manual'`
  - `suitable_site_types` populated with meaningful values (`["comparison", "directory"]`, `["ecommerce", "marketplace"]`, etc.)

### Documents produced
- **`003_contracts_and_standards.md`** — updated with "CSS Theme Template Contract" section (between Section Context Variable Contract and Query Parameterisation Contract). Covers responsibility split (renderer vs theme template), template variables table, theme storage columns, lineage columns, review gate, forking rules. **This section will need rewriting after Phase 4** to reflect the new three-entity model.
- **`025_palette_layout_typography_migration.md`** — the full migration plan. 15 layouts, 6 typography sets, 8 phases, palette merge rules, renderer flow, downstream dependencies. **This is the authoritative reference for the next session.**
- **`migration_025_component_library.sql`** — the 10 component inserts (already run)
- **`migration_025_component_triage.sql`** — an earlier work-item-based approach that was superseded by the direct insert approach. **Do not run this file.** It's been replaced by `migration_025_component_library.sql`.

---

## 3. What's Next: Phase 1 — Design the 15 Layout CSS Templates

This is the bulk of the remaining work. Each layout is a full CSS template (~250-400 lines) that uses `{{palette "key" "fallback"}}`, `{{typo "key" "fallback"}}`, `{{token "key" "fallback"}}` template helper functions for safe map-based variable access.

All 15 layouts are being produced in one batch so the structural vocabulary stays consistent: same section class naming, same responsive breakpoint philosophy, same accessibility patterns.

### Template variable system (new — replaces the current flat struct)

Current `cssTemplateData` has hardcoded fields (`Primary string`, `Secondary string`, etc.). The new shape is:

```go
type cssTemplateData struct {
    Palette      map[string]string  // all colour variables
    Typography   map[string]string  // font stacks, sizes
    Structure    map[string]string  // container widths, padding, radii, shadows

    Components       []string
    SectionStyles    []sectionStyleEntry
    BackgroundIsDark bool
    SurfaceIsDark    bool
}
```

Layout templates reference variables via helper funcs, not direct map access:

```css
:root {
  --color-primary: {{palette "primary" "#1a365d"}};
  --color-hero-title: {{palette "hero_title" "#0f172a"}};
  --font-body: {{typo "font_family" "system-ui, sans-serif"}};
  --shadow: {{token "shadow" "0 2px 4px rgba(0,0,0,0.1)"}};
}
```

Every `{{palette ...}}` call includes a fallback value. Missing keys produce the fallback, not empty strings.

### Palette merge rule

When a site selects a theme, two palette sources merge:
- **Core slots (spec wins where present):** `primary`, `secondary`, `accent`, `background`, `surface`, `text`, `text_muted`, `border`
- **Specialised slots (theme wins):** `primary_hover`, `heading`, `header_bg`, `hero_title`, `cta_bg`, `footer_text`, and any other theme-declared slots

### Contracts each layout must follow

1. **Colour Inheritance Model** (003): base typography uses `var(--section-*, var(--color-*))` fallback pattern
2. **Dark Section Variable Contract** (003): renderer owns `--section-*` defaults, layouts MUST NOT declare them
3. **Template helper fallback** (025): every `{{palette ...}}` / `{{typo ...}}` / `{{token ...}}` has a default
4. **Responsive**: `@media (max-width: 768px)` at minimum; touch targets >= 44px

### What each layout CSS template covers

- `:root` variable block (palette + typography + structure via helpers)
- Body and base typography rules
- Site header structure (basic — layout-specific headers are separate components)
- Site footer structure (basic)
- Section container patterns (`.hero-section`, `.features-section`, `.services-section`, `.call-to-action-section`, `.faq-section`, `.about-section`, `.differentiators-section`, etc.)
- Button styles (primary, secondary, sizes)
- Form input styles
- Card styling
- Grid patterns
- Responsive breakpoints
- Layout-distinctive patterns (glass blur, diagonal splits, sidebar structure, etc.)

---

## 4. The 15 Layouts — Detailed Descriptions

### 1. `brochure-formal`

**Character:** Structured, understated, CTA-driven. Corporate restraint.  
**Example sites:** Consultancies, law firms, finance, B2B professional services.  
**Structural traits:**
- Container max-width 1200px, centered
- Sections alternate white/surface backgrounds for visual rhythm
- Hero: text-focused with subtle background tint, centered or left-aligned, clear CTA button below
- Features/services: 3-column grid of icon+text cards with subtle borders, no heavy shadows
- Call-to-action: full-width band with primary background, white text, single CTA
- Testimonials: centered quote with attribution
- Footer: 4-column link grid, dark background
- Buttons: standard rectangle, subtle radius (0.375rem), no uppercase, medium weight
- Typography: body 16px, headings use weight and size for hierarchy (not colour or decoration)
- Hover states: subtle — border colour change on cards, slight shadow lift
- Responsive: 3-col → 2-col → 1-col; hero stacks at 768px

**Default header/footer:** `site-header` (existing), `site-footer` (existing)  
**Default typography:** `sans-modern` (Inter)  
**Mapped themes:** `default`, `standard-brochure`, `professional-dark`

---

### 2. `brochure-bold`

**Character:** High-energy conversion. Large hero, gradient accents, strong CTAs, more visual motion.  
**Example sites:** Tech startups, fitness brands, sales-led SaaS landing pages.  
**Structural traits:**
- Hero: tall (70-80vh), gradient overlay or large background, oversized heading (3.5-4rem), prominent dual CTAs
- Features: larger cards with gradient accent borders or subtle gradient backgrounds
- CTA sections: gradient backgrounds using primary→accent or primary→secondary, larger text
- Buttons: slightly larger padding, font-weight 700, optional uppercase on primary CTAs, larger border-radius
- Cards: more shadow (the `--shadow-lg` token), more hover motion (translateY(-4px))
- Section padding: generous (5rem top/bottom minimum)
- Gradient accents: thin gradient bars under headings, gradient hover effects on buttons
- Responsive: hero heading drops to 2rem on mobile; gradient bars scale down

**Default header/footer:** `site-header` (existing), `site-footer` (existing)  
**Default typography:** `display-bold` (Archivo Black / Impact)  
**Mapped themes:** `bold-conversion`, `tech`, `dark-modern`

---

### 3. `portfolio-kinetic`

**Character:** Asymmetric, motion-forward, large display type, negative space. Creative energy.  
**Example sites:** Design studios, creative agencies, photographer portfolios (the "vonc" pole).  
**Structural traits:**
- Hero: full viewport height, oversized display type (5-6rem heading), minimal content, text may be offset to one side with empty space intentionally preserved
- Layout: asymmetric — alternating sections break from centred grid, using CSS grid with offset columns (e.g. 40%/60% splits, then 60%/40%)
- Project/work showcase: large image tiles with hover reveal of project name, masonry or staggered grid
- Typography-led: headings are the primary visual element, not images or colour blocks
- Transitions: transform and opacity transitions on scroll-triggered elements (the CSS provides the animation classes; JS would be a component concern)
- Minimal decoration: no card borders, no shadows in default state; whitespace is the separator
- CTAs: text-link style with animated underline, not button-block style
- Footer: minimal — single line copyright, maybe social links
- Responsive: asymmetric grids collapse to single-column stacked; type scales down but stays relatively large

**Default header/footer:** existing minimal header, existing minimal footer  
**Default typography:** `sans-modern` (Inter — clean canvas for the type to dominate)  
**Mapped themes:** none currently — exists in library for adoption/new-build matching

---

### 4. `magazine-grid`

**Character:** Content-dense, article-card-dominant, sidebar-optional. Publishing feel.  
**Example sites:** News sites, opinion publications, long-form blog networks.  
**Structural traits:**
- Hero: featured article with large image, overlaid or adjacent headline, author/date meta
- Main content area: 2/3 + 1/3 grid (main article grid + sidebar) on desktop, full-width on mobile
- Article cards: image top, category badge, title, excerpt (2-line clamp), author/date meta at bottom
- Sidebar: widgets — popular posts (numbered list), categories, newsletter signup, ad zone placeholder
- Section headers: title + "View all →" link aligned right
- Featured article: larger card spanning full width of main column, image + text side-by-side on desktop
- Compact article cards: smaller image, tighter padding, for "more stories" sections
- Pagination or "Load more" at bottom (CSS only — behaviour is component JS)
- Footer: multi-column with newsletter form
- Responsive: sidebar drops below main content at 1024px; article grid goes from 3→2→1 column

**Default header/footer:** new `header-with-categories`, `footer-4-column` (existing)  
**Default typography:** `serif-editorial` (Merriweather + Georgia)  
**Mapped themes:** `content-modern`

---

### 5. `utility-tool`

**Character:** Minimal chrome, tool-first. The page exists to serve a function, not to sell.  
**Example sites:** Online calculators, unit converters, developer utilities, data validators.  
**Structural traits:**
- Hero: almost nothing — a one-line title, maybe a one-line description, then straight into the tool area
- Tool area: centred, max-width 800px (narrower than typical), generous padding, clear form controls
- Form elements: larger inputs (48px height), clear labels, visible focus states, grouped with fieldsets
- Output/result area: visually distinct from input (surface background, or bordered card), clear typography for results
- Supporting content below the fold: "How to use", "About this tool", "Related tools" — present but secondary
- Minimal navigation: the header is compact (52px), just brand + sparse links
- No hero images, no testimonials, no social proof sections — just the tool
- Footer: minimal — copyright, maybe privacy/terms links
- Responsive: tool area is already narrow so it adapts naturally; inputs stack vertically

**Default header/footer:** new `header-minimal-tool`, existing minimal footer  
**Default typography:** `sans-modern` (Inter)  
**Mapped themes:** none currently

---

### 6. `media-grid`

**Character:** Thumbnail-dominant, continuous scroll, media-forward. Discovery through browsing.  
**Example sites:** Video platforms, audio libraries, image galleries (the "youtube" pole).  
**Structural traits:**
- Hero: minimal or absent — the grid IS the content. Optionally a featured/pinned item at the top spanning full width
- Primary grid: auto-fill thumbnails in 4-col desktop → 3-col → 2-col → 1-col responsive, uniform aspect ratio (16:9 for video, 1:1 for audio)
- Thumbnail cards: image dominant (70-80% of card), title below, channel/author, view count, duration overlay on image
- Hover state: thumbnail scales slightly (transform: scale(1.03)), preview info becomes more prominent
- Filter bar: horizontal category chips or tabs above the grid, scrollable on mobile
- Infinite scroll / load more pattern (CSS provides the loading state styles)
- Sidebar: optional, only on desktop, for "trending" or "recommended" — similar to `magazine-grid` sidebar but tuned for media
- Header: search-dominant (uses `header-with-search`)
- Responsive: grid adapts column count; thumbnail aspect ratio stays consistent; filter chips scroll horizontally

**Default header/footer:** new `header-with-search`, existing minimal footer  
**Default typography:** `sans-modern` (Inter)  
**Mapped themes:** none currently

---

### 7. `docs-sidebar`

**Character:** Reference-grade, sidebar-navigated, code-friendly. Knowledge architecture.  
**Example sites:** Developer documentation, API references, knowledge bases, technical guides.  
**Structural traits:**
- Layout: 3-zone CSS grid — fixed left sidebar (260px), main content (flex), optional right-side "on this page" TOC (200px)
- Sidebar: persistent on desktop, collapsible on mobile (hamburger in header), scrollable independently, current-page indicator
- Sidebar items: nested tree structure (section → pages → subpages), active item highlighted
- Main content: max-width 780px within its zone, generous line-height (1.7-1.8), code blocks with distinct styling
- Code blocks: monospace, surface background, left border accent, horizontal scroll for long lines, copy button affordance (CSS only)
- Headings: anchored (padding-top to offset sticky header), subtle left border or accent for visual hierarchy
- Tables: bordered, alternating row backgrounds, responsive horizontal scroll wrapper
- Callouts/admonitions: info/warning/danger boxes with left accent border and icon
- Footer: minimal, within main content zone only (sidebar doesn't have a footer)
- Responsive: sidebar collapses at 1024px; right TOC hides at 1280px; main content becomes full-width

**Default header/footer:** new `header-docs`, existing minimal footer  
**Default typography:** `mono-technical` (IBM Plex Mono for code, system-ui for body)  
**Mapped themes:** none currently

---

### 8. `soft-editorial`

**Character:** Relaxed, reading-first, organic. Paper-textured feel with warm typography.  
**Example sites:** Wellness blogs, lifestyle sites, personal essays, bakeries, artisan businesses.  
**Structural traits:**
- Background: not pure white — tinted slightly warm (the palette's `background` slot, typically `#fafaf9` or `#fffbeb`)
- Sections: generous padding (6rem), wider line-height (1.7), content areas narrower than typical (max-width 1000px)
- Hero: text-centred with a gentle gradient fade (not a hard background — a transparent wash using the primary at 3% opacity)
- Cards: barely-there borders (`1px solid rgba(0,0,0,0.03)`), soft shadows (`0 4px 6px rgba(0,0,0,0.05)`), generous border-radius (12px)
- Buttons: pill-shaped (border-radius: 50px), generous horizontal padding (2rem), lighter feel
- Headings: serif display font (font-family: var(--font-display)), letter-spacing: -0.02em
- Header: transparent background, minimal — just floats on the page
- Section breaks: no hard borders — just spacing and gentle background shifts
- Footer: light surface background (not dark), warm text
- Responsive: narrower content area means less dramatic reflow; padding reduces but stays generous

**Default header/footer:** `site-header` (existing), `site-footer` (existing)  
**Default typography:** `serif-editorial` (Merriweather + Lato body)  
**Mapped themes:** `bakery`, `warm-friendly`, `calm-minimal`, `soft-editorial`

---

### 9. `technical-precise`

**Character:** Clean, dense, engineered. Communicates precision and competence.  
**Example sites:** SaaS platforms, infrastructure companies, engineering consultancies.  
**Structural traits:**
- Tight border-radius (6px) — not rounded, not sharp, precisely engineered
- Cards: bordered (`1px solid var(--color-border)`), minimal shadow, hover adds `box-shadow` and border-colour change to secondary
- Header: glass effect (`backdrop-filter: blur(8px)`, semi-transparent background), border-bottom
- Typography: clean sans-serif, `-webkit-font-smoothing: antialiased`, letter-spacing: -0.01em on headings
- Buttons: medium weight (font-weight 500), tight letter-spacing, standard radius
- Grid: clean 3-column, consistent gap, no playfulness
- Section backgrounds: white alternating with a very subtle grey (`--color-background-alt`, typically slate-50)
- Hero: clean text hierarchy, no gradients, no overlays — text and CTA on a clean background
- Data/metrics: if the page has stats, display them in a 4-column stat row with large numbers
- CTA sections: solid primary background (no gradient), white text, clean and flat
- Footer: light background (slate-50), muted text, no drama
- Responsive: standard 3→2→1 breakpoints; glass header loses blur on mobile (performance)

**Default header/footer:** `site-header` (existing), `site-footer` (existing)  
**Default typography:** `sans-modern` (Inter)  
**Mapped themes:** `premium-elegant` (with serif typography override), `modern-engineering-clean`

---

### 10. `high-energy`

**Character:** Aggressive, kinetic, bold. Impact over subtlety.  
**Example sites:** Boxing gyms, combat sports, fitness events, extreme sports (the "boxing" pole).  
**Structural traits:**
- Hero: large (80vh minimum), dark background, oversized uppercase heading (4-5rem, text-transform: uppercase, letter-spacing: 2px), aggressive CTA
- Diagonal/angled section separators: `clip-path` or `transform: skewY(-3deg)` on section backgrounds, content counter-rotated to stay level
- Buttons: uppercase, letter-spacing 0.05em, font-weight 700, slightly larger padding, hard edges (small or zero radius)
- Colour usage: high contrast — dark sections with bright accent highlights, alternating dark/light sections
- Typography: impact/condensed display font for headings, standard sans for body
- Cards: dark background cards on light sections, light on dark — always contrasting with parent
- Borders: none or very sharp (1px solid). No soft shadows — hard shadows if any
- Section padding: aggressive (5-6rem)
- Feature callouts: large numbers or stats, bold, potentially with accent colour underline
- Footer: dark, dense, minimal decoration
- Responsive: hero heading drops to 2.5rem; diagonal separators flatten on mobile (clip-path simplifies); cards go single-column

**Default header/footer:** existing minimal header, existing minimal footer  
**Default typography:** `display-bold` (Impact / Archivo Black)  
**Mapped themes:** `boxing`

---

### 11. `comparison-aggregator`

**Character:** Search-first, data-dense, trustworthy. Users arrive to compare options.  
**Example sites:** VetComparison.uk, insurance comparison, broadband comparison, trade directories.  
**Structural traits:**
- Hero: prominent search input as the centrepiece (not just a title — the search IS the hero), brief positioning statement above, optional disclaimer banner below
- Filter/controls bar: sticky below header on scroll, contains search input, sort dropdown, filter toggles (category buttons), price range slider
- Results grid: responsive card grid (auto-fill, min 320px), each card shows provider name, key metric (prominently sized/coloured), location, type tag, description excerpt, CTA button
- Info/alert banners: styled callout boxes for regulatory information, CMA updates, etc.
- Calculator/tool sections: dedicated area below results for cost calculators or comparison tools
- Guide cards: 2-column grid of linked guide summaries (title, excerpt, "Read guide →")
- Footer: uses `footer-with-disclaimer` — standard columns plus heavy disclaimer block
- Data source: all listing data from `query.*` sources, never LLM-fabricated
- Responsive: filter controls stack vertically on mobile; result cards go single-column; search input retains prominence

**Reference:** VetComparison.uk (`/mnt/user-data/uploads/index.html` in this session)  
**Default header/footer:** new `header-with-search`, new `footer-with-disclaimer`  
**Default typography:** `sans-modern` (Inter)  
**Mapped themes:** none currently

---

### 12. `affiliate-hub`

**Character:** Commercial but trustworthy. Product-focused content with affiliate monetisation.  
**Example sites:** Product review sites, "best X for Y" buyer's guides, deal aggregators.  
**Structural traits:**
- Hero: category-led — "Best [Category] for [Audience]" pattern, possibly with a featured product highlight
- Category index: horizontal or grid of category cards linking to category pages
- Product cards: image-dominant, rating (stars + count), tagline, price, prominent CTA button ("Check price on X"), secondary review link, required "Ad" or "Affiliate" disclosure label
- Review content blocks: long-form review sections with pros/cons, feature comparison tables
- Comparison tables: responsive horizontal-scroll tables with feature rows and product columns
- Disclosure banners: subtle but persistent — "This site contains affiliate links" in a muted banner
- Sidebar: optional — "Top Picks", "Editor's Choice", category links
- Footer: 4-column with disclosure in a distinct block
- Responsive: product cards go 2-col then 1-col; comparison tables scroll horizontally; disclosure remains visible

**Default header/footer:** new `header-with-cart-or-nav`, `footer-4-column` (existing)  
**Default typography:** `sans-modern` (Inter)  
**Mapped themes:** none currently

---

### 13. `ecommerce-storefront`

**Character:** Retail-clean, product-forward. Commercial trust signals.  
**Example sites:** Independent shops, small-catalogue retailers, marketplace sellers.  
**Structural traits:**
- Hero: promotional banner / featured collection with large image, overlay text, shop-now CTA
- Category grid: large image cards for top-level categories (4-column desktop, 2-column mobile)
- Product grid: uses `product-grid` component — image-dominant cards, clean pricing, sale badges, add-to-cart buttons
- Product cards: white/neutral background regardless of page palette (product images demand it), consistent square aspect ratio
- Trust bar: thin horizontal strip with trust signals (free shipping, returns policy, secure checkout icons)
- Featured/sale section: highlighted background, "Sale" badging, strike-through original prices
- Newsletter: prominent signup section (email capture for marketing)
- Footer: multi-column with payment method icons, trust badges, customer service links
- Cart patterns: cart icon with count badge in header, mini-cart dropdown (CSS for structure; JS for behaviour)
- Responsive: category grid 4→2; product grid 4→3→2→1; trust bar wraps; hero text stacks

**Default header/footer:** new `header-with-cart-or-nav`, `footer-4-column` (existing)  
**Default typography:** `sans-modern` (Inter)  
**Mapped themes:** none currently

---

### 14. `tool-first-landing`

**Character:** The tool IS the page. Everything else is supporting material.  
**Example sites:** Single-purpose calculators, API playgrounds, demo tools, configuration generators (the "thunder compute" pole).  
**Structural traits:**
- Above the fold: tool occupies 80%+ of viewport. Minimal heading (one line), possibly no heading at all — just the tool interface
- Tool area: full container width (not the narrow 800px of `utility-tool`), designed for tools with wide layouts (dashboards, multi-column inputs, preview panes)
- Split-pane option: left input / right output layout using CSS grid (50/50 or 40/60 split)
- Below the fold: "How it works" (3-step horizontal), "Why use this" feature list, documentation-style details
- Minimal decoration: no testimonials, no social proof, no team section — the tool speaks for itself
- Dark-mode friendly: many tool-first sites prefer dark UI for readability of output
- Code output: if the tool generates code, monospace output with syntax-highlighting-ready classes
- Footer: minimal, functional
- Responsive: split panes stack vertically; tool area takes full width; below-fold sections are standard single-column

**Distinction from `utility-tool`:** `utility-tool` is a contained calculator in a page; `tool-first-landing` is a page that IS a tool. `utility-tool` has narrower max-width (800px) and more chrome. `tool-first-landing` goes full container width and minimises everything that isn't the tool.

**Default header/footer:** new `header-minimal-tool`, existing minimal footer  
**Default typography:** `sans-modern` (Inter)  
**Mapped themes:** none currently

---

### 15. `industry-hub`

**Character:** Independent authority on a vertical. Not a participant in the industry — an information resource about it.  
**Example sites:** Gas wholesaler directories, industry explainer sites, regulatory information hubs (the "gaswholesalers.co.uk" pole).  
**Structural traits:**
- Hero: positioning statement that clearly establishes the site's role — "Your independent guide to [industry]", prominent search or browse CTA
- "About this site" banner: a subtle callout near the top explaining the site is an independent resource, not an industry participant. Border-left accent, muted background.
- Directory section: primary content element — a grid of provider/supplier cards (name, description, region, category tag), filter bar above. Uses `directory-listing` component.
- Guide index: not a blog-chronological feed but a topic-organised index. Cards grouped by topic area (e.g. "Regulations", "How It Works", "Buyers' Guide"), each with title and description.
- News/updates section: smaller, chronological, separate from the guides — for regulatory changes, industry developments. Compact cards, date-prominent.
- Glossary/reference section: optional alphabetical or categorised term list for industry jargon
- Footer: uses `footer-with-disclaimer` — heavy disclaimer block explaining independence, data sourcing, liability limitations
- Typography: slightly more formal than magazine — serif display headings signal authority without being corporate
- Section rhythm: directory first (primary value), then guides, then news, then reference — ordered by user intent frequency
- Responsive: directory grid 3→2→1; guide index cards stack; news section becomes a simple list

**Default header/footer:** new `header-with-categories`, new `footer-with-disclaimer`  
**Default typography:** `serif-editorial` (Merriweather)  
**Mapped themes:** none currently

---

## 5. Schema Changes Pending (Phase 2)

Three new tables (`palettes`, `layouts`, `typography_sets`), three new FK columns on `css_themes`. Full DDL is in section 3 of `025_palette_layout_typography_migration.md`. Key details:

- `layouts` has `default_header_component_id` and `default_footer_component_id` FKs to `content_components`
- `palettes.colours` is a flat JSONB map — any key names, no fixed struct
- `css_themes` gets nullable `palette_id`, `layout_id`, `typography_set_id` columns
- Legacy columns (`css_template`, `css_content`, `color_palette`, `typography`) stay until Phase 7

## 6. Renderer Changes Pending (Phase 4)

`render_css_from_spec_action.go` gets rewritten to:
1. Load palette via `css_themes.palette_id → palettes.colours`
2. Load layout via `css_themes.layout_id → layouts.css_template`
3. Load typography via `css_themes.typography_set_id → typography_sets.fonts + scale`
4. Merge palette (spec wins for core slots, theme wins for specialised)
5. Build map-based `cssTemplateData`
6. Parse layout template with `palette`, `typo`, `token` template funcs
7. Render
8. Append component snippets (unchanged)
9. Append renderer-enforced `--section-*` defaults (unchanged)

Cutover is direct (no shadow mode). Rollback is a code revert.

## 7. Existing Theme → Layout Mapping (for Phase 3 seeding)

| Theme | Layout | Typography |
|---|---|---|
| `default`, `standard-brochure`, `professional-dark` | `brochure-formal` | `sans-modern` |
| `bold-conversion`, `tech`, `dark-modern` | `brochure-bold` | `display-bold` |
| `bakery`, `warm-friendly`, `calm-minimal`, `soft-editorial` | `soft-editorial` | `serif-editorial` |
| `boxing` | `high-energy` | `display-bold` |
| `premium-elegant` | `technical-precise` | `serif-classical` |
| `content-modern` | `magazine-grid` | `serif-editorial` |
| `modern-engineering-clean` | `technical-precise` | `sans-modern` |

## 8. Typography Sets (for Phase 3 seeding)

| Name | Fonts | Scale |
|---|---|---|
| `sans-modern` | Inter + system fallbacks | 16px base, 1.6 line-height |
| `serif-editorial` | Merriweather headings + Lato/Georgia body | 16px base, 1.7 line-height |
| `display-bold` | Archivo Black / Impact headings + sans body | 16px base, 1.5 line-height |
| `mono-technical` | IBM Plex Mono code + system-ui body | 15px base, 1.6 line-height |
| `serif-classical` | Cormorant Garamond + Georgia | 17px base, 1.7 line-height |
| `sans-friendly` | Nunito + Segoe UI | 16px base, 1.6 line-height |

## 9. Key File Paths

| File | Status | Purpose |
|---|---|---|
| `/mnt/user-data/uploads/render_css_from_spec_action.go` | Current production code | The renderer to be rewritten in Phase 4 |
| `/mnt/user-data/uploads/003_contracts_and_standards.md` | Current + CSS Theme Template Contract section | Contracts doc to update in Phase 6 |
| `/mnt/user-data/outputs/025_palette_layout_typography_migration.md` | Plan doc | Authoritative migration plan |
| `/mnt/user-data/outputs/003_contracts_and_standards.md` | Updated contracts doc | Has CSS Theme Template Contract section |
| `/mnt/user-data/outputs/css_templating.go` | Deployed | TemplateCSSFromSpec helper |
| `/mnt/user-data/outputs/migration_025_component_library.sql` | Deployed | 10 library component inserts |
| `/mnt/project/production_agent-chassis-full_context.txt` | Reference (115K lines) | Full codebase context |
| `/mnt/project/bk_agent_definitions_backup.sql` | Reference | Agent definitions including component-creator workflow |
| `/mnt/project/some_schemas` | Reference | DB schemas (sites, site_work_items, content_components, etc.) |

## 10. Things to Watch Out For

- Column is `type` not `agent_type` on `agent_definitions`; workflow in `default_config` not `config`
- Always `AND deleted_at IS NULL` on agent_definitions queries
- `bakery` theme has Comic Sans on `.section__title` — should be removed during palette extraction
- `content-modern` theme has extensive article-card CSS that's really layout, not palette — most of it goes into `magazine-grid` layout
- `modern-engineering-clean` and `soft-editorial` themes have comments in their CSS (/* Palette: Organic & Calm */ etc.) — strip these during palette extraction
- The existing `standard-brochure` template in `css_themes.css_template` is the only one that actually renders today. Its layout structure should be the baseline for `brochure-formal`.
- Some themes have inconsistent indentation (4-space in dark-modern/calm-minimal vs 2-space in bakery/default). Normalise during extraction.
- `suitable_site_types` values on the 10 new components use terms like `comparison`, `industry-hub`, `ecommerce`, `affiliate`. These must align with whatever the classifier produces. If the classifier uses different vocabulary, update either the classifier or the component arrays.
- Don't use `logger.Debug` — it won't show in logs. Use `logger.Info` or `logger.Warn`.
- Kubernetes namespace is `-n ai-persona-system` for main pods, `-n kafka` for Kafka.
