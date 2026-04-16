# Palette, Layout, Typography — Composable Theme Migration

## Splitting `css_themes` into independently-versioned entities

---

## 1. What This Project Is

Today `css_themes.css_template` mixes three distinct concerns into one row: palette, typography, and layout. This has been masked by a silent fallback where only `standard-brochure` has a populated `css_template` and every other theme collapses to it. The 13 non-brochure themes exist as palette metadata only; selecting `bakery` or `dark-modern` or `soft-editorial` today produces the same `standard-brochure` layout with a slightly different `:root` block.

The migration splits these three concerns into three tables (`palettes`, `layouts`, `typography_sets`), each independently versionable with the same lineage model we added to `css_themes` and `style_collections`. A `css_themes` row becomes a composition — a named bundle pointing at one palette, one layout, one typography set.

This enables:

- **Genuinely different visual identities per theme.** A site picking `boxing` gets an aggressive, kinetic layout. A site picking `soft-editorial` gets a calm, reading-first layout with generous line-heights and serif display. These aren't variations of the same layout with different colours — they're structurally different.
- **Adoption forks that capture what's distinctive.** Adopting `vonc.com` (abstract/kinetic portfolio) produces a new layout row the library didn't have. Adopting a small-business site usually just produces a new palette and reuses an existing layout. Forking is granular.
- **Composable selection.** The selector eventually picks "warm palette + content-heavy layout + editorial typography" from independently-scored options rather than one bundled row that approximately matches.

The current theme library is effectively one layout with 14 palette skins. Post-migration, the library holds a diverse set of layouts, each pairable with any palette, producing orders of magnitude more possible visual identities.

---

## 2. Scope Decisions

### Layout diversity

We are **not** producing archetype layouts that each cover a category of themes. We are producing layouts that reflect genuinely different user intents and content types. Examples of the axis of variation we want to cover:

- **Brochure/corporate** — structured sections, clear hierarchy, CTA-driven. Fits consultancies, law firms, professional services.
- **Magazine/editorial** — dense article grids, sidebars, content-first. Fits publications, news sites, long-form blogs.
- **Portfolio/kinetic** — large typography, asymmetric layouts, motion, abstract composition. Fits design studios, agencies, creative work (the "vonc" pole).
- **Commerce/grid** — product-card-dominated, filtering UI, dense information density. Fits retailers and marketplaces.
- **Utility/tool** — minimal chrome, big input/output area, operational feel. Fits calculators, simulators, developer tools (the "thunder compute" pole).
- **Media/streaming** — hero-dominant, continuous scroll, thumbnail grids. Fits video, audio, media (the "youtube" pole).
- **Documentation/reference** — sidebar nav, clean typography, anchored sections, code-friendly. Fits developer docs, knowledge bases.
- **High-energy/bold** — aggressive hero, large type, hard edges, strong accent colours. Fits sports, fitness, combat disciplines (the "boxing" pole).
- **Soft/editorial** — relaxed padding, serif typography, generous line-heights, paper-textured feel. Fits wellness, lifestyle, thoughtful content.
- **Technical/precise** — clean sans-serif, tight radii, subtle shadows, dense information. Fits SaaS, infrastructure, engineering.

This isn't a closed list. Each layout is its own versioned entity. New layouts join the library as adoption captures distinctive structural patterns, or when a human designs one.

### Layouts are not palette-bound

A layout references `{{.Palette.primary}}`, `{{.Palette.hero_title}}`, etc., but doesn't care which palette fills those slots. The same portfolio/kinetic layout works with a muted cool palette, a warm earthy one, or a stark black-and-white one.

### We are not touching style_collections structure

`style_collections` remains a bundle of `css_theme_id` + `header_component_id` + `footer_component_id`. Those relationships stay. The lineage columns we added to `style_collections` stay. Only the entity `style_collections.css_theme_id` points at gets internally restructured.

### We are not changing the selector in this migration

`SelectStyleCollectionAction` still picks a `style_collections` row and follows its `css_theme_id` FK to a `css_themes` row. The renderer now loads palette/layout/typography from that row, but the selection logic is unchanged. Composable selection ("pick a warm palette, pick an editorial layout independently") is a separate project.

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

`structure_tokens` holds container widths, section padding scales, border radii, shadow tokens. These are visible to layouts (and templates) without inflating the palette.

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

Layout templates reference variables via map access:

```css
:root {
  --color-primary: {{.Palette.primary}};
  --color-hero-title: {{.Palette.hero_title}};
  --font-body: {{.Typography.font_family}};
  --shadow: {{.Structure.shadow}};
}
```

### Template helper: safe lookup with defaults

Go templates don't fail on missing map keys; they emit empty strings. A layout referencing `{{.Palette.hero_title}}` against a palette that doesn't declare `hero_title` would produce broken CSS. We add a template function:

```go
funcMap := template.FuncMap{
    "palette": func(key, fallback string) string { ... },
    "typo":    func(key, fallback string) string { ... },
    "token":   func(key, fallback string) string { ... },
}
```

Layouts use `{{palette "primary" "#1a365d"}}` — if the palette has `primary`, it wins; otherwise the fallback applies. Safe, explicit, no silent empty-string bugs.

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
3. Load layout           via css_themes.layout_id → layouts.css_template
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

The initial layout set needs to cover a range wide enough that composing each with each palette produces distinct-feeling sites. The poles referenced: vonc (portfolio/kinetic), youtube (media grid), thunder compute (utility/tool), boxing (high-energy bold), bakery (warm), a soft-editorial, a standard corporate brochure.

### Proposed initial layouts

| Layout name | Character | Use case examples |
|---|---|---|
| `brochure-formal` | Structured, understated, CTA-driven sections | Consultancies, law, finance, B2B services |
| `brochure-bold` | Large hero, gradient accents, strong CTAs, higher motion | Tech startups, fitness, sales-led businesses |
| `portfolio-kinetic` | Asymmetric sections, large display type, motion, negative space | Design studios, agencies, creative portfolios |
| `magazine-grid` | Article cards, sidebar, content-dominant | Publications, news, long-form blogs |
| `utility-tool` | Minimal chrome, centered tool area, dense operational feel | Calculators, simulators, dev tools |
| `media-grid` | Thumbnail grid, continuous scroll, media-first | Video, audio, media libraries |
| `docs-sidebar` | Left sidebar nav, anchored content, code-friendly | Developer docs, knowledge bases |
| `soft-editorial` | Serif display, relaxed padding, paper-textured background, pill buttons | Wellness, lifestyle, thoughtful long-form |
| `technical-precise` | Clean sans-serif, tight radii, subtle shadows, glass headers | SaaS, infrastructure, engineering |
| `high-energy` | Uppercase display type, hard edges, diagonal splits, strong accents | Combat sports, fitness, events |

Ten layouts to start. Each produced as a full CSS template with template variables for palette/typography/structure. Not all 13 existing themes need unique layouts — several can share one — but the library has clear variety.

### What constitutes a "full layout template"

Each layout produces CSS covering:

- `:root` variable block (receives palette + typography + structure via template)
- Body and base typography rules (using the Colour Inheritance Model: `var(--section-*, var(--color-*))` fallbacks)
- Site header
- Site footer
- Section container patterns (`.hero-section`, `.features-section`, `.call-to-action-section`, etc.)
- Button styles (primary, secondary, sizes)
- Form input styles
- Responsive breakpoints
- Any layout-distinctive patterns (glass blur, gradient overlays, asymmetric grids, sidebar structure, etc.)

Each layout is roughly 250-400 lines of CSS. Each is reviewed for correctness against the Colour Inheritance Model and the Dark Section contract.

### Mapping existing themes to initial layouts

This is a placeholder mapping — the final mapping happens after layouts are drafted. Rough intent:

| Theme | Likely layout |
|---|---|
| `default`, `standard-brochure`, `professional-dark` | `brochure-formal` |
| `bold-conversion`, `tech`, `dark-modern` | `brochure-bold` |
| `bakery`, `warm-friendly` | `soft-editorial` |
| `calm-minimal`, `soft-editorial` | `soft-editorial` |
| `boxing` | `high-energy` |
| `premium-elegant` | `technical-precise` (with serif typography) |
| `content-modern` | `magazine-grid` |
| `modern-engineering-clean` | `technical-precise` |

New layouts added over time fill gaps — for example, no existing theme maps to `portfolio-kinetic`, `utility-tool`, `media-grid`, or `docs-sidebar`. Those exist in the library for adoption to fork into, or for new-build sites that specifically need them.

---

## 8. Migration Phases

### Phase 1 — Design and draft the layouts (the large piece)

- Decide final list of layouts to seed (starting from the 10 above)
- Write the CSS template for each, using `{{palette "key" "default"}}` / `{{typo ...}}` / `{{token ...}}` references
- Test each by rendering with a sample palette and eyeballing the output
- Review each against the Colour Inheritance Model and Dark Section contract
- Output: 10 ready CSS templates + structure_tokens + rationale for each

This phase is the bulk of the work. Each layout deserves design attention.

### Phase 2 — Schema migration (additive, reversible)

One migration SQL file:

- Create `palettes`, `layouts`, `typography_sets` tables with lineage columns
- Add FK columns to `css_themes` (nullable)
- Add indexes
- No data migration yet

Nothing reads the new tables; nothing breaks if this goes wrong.

### Phase 3 — Seed the new tables

For each of the 13 `css_themes` rows:

- Extract its `css_content` `:root` block into a `palettes` row, preserving all colour slots
- Extract its distinctive typography (if any) into a `typography_sets` row; otherwise point at a shared default
- Map to an appropriate `layouts` row (the layouts from Phase 1)

Populate the new FK columns on `css_themes` to point at the seeded rows.

After this phase: all `css_themes` rows have three FKs populated; the new tables have 13+ palette rows, 10+ layout rows, 4-6 typography sets; nothing reads any of this yet.

### Phase 4 — Renderer path change (shadow mode first)

- New `cssTemplateData` shape with maps
- Template helper funcs (`palette`, `typo`, `token`)
- `RenderCSSFromSpecAction` new path implemented alongside the old
- Feature flag to select which path runs
- Shadow mode: both paths run, output compared, diffs logged; only the old output is deployed

Verify the new path produces byte-identical output to the old path for existing live sites. Any diff is a regression until investigated.

### Phase 5 — Cutover

- Flip the feature flag to the new path
- Monitor live sites for visual regressions
- Keep old path available for one-command rollback

Old columns (`css_template`, `css_content`, `color_palette`, `typography` on `css_themes`) remain populated but no longer read by the renderer.

### Phase 6 — Adoption-fork action update

`fork_theme_from_site` changes to produce granular forks:

- Always fork a `palettes` row (adopted palettes are almost always distinctive)
- Optionally fork a `layouts` row if the adopted site's structural CSS is genuinely novel
- Optionally fork a `typography_sets` row if the fonts are distinctive
- Create the `css_themes` bundle row pointing at whichever rows were chosen (new or reused)
- HITL review work item becomes richer: review the palette, review the layout if new, review the bundle as a whole

"Distinctive" is initially a flag set by the adoption analysis. LLM-grade judgement about what counts as genuinely novel comes later.

### Phase 7 — Contract doc rewrite

Replace the current "CSS Theme Template Contract" section in `003_contracts_and_standards.md` with subsections for:

- Palette contract (what slots, how slots are used)
- Layout contract (what variables a layout references, responsibility split)
- Typography set contract
- Theme bundle contract (composition row)
- Palette merge rules (core vs specialised slots)
- Forking rules for each entity
- Review gates

### Phase 8 — Legacy column cleanup (deferred)

Once confident nothing reads them, drop from `css_themes`:

- `css_template`
- `css_content`
- `color_palette`
- `typography`

This happens after a stabilisation period (suggest: two weeks post-cutover with no regressions).

---

## 9. Non-Goals for This Migration

- **Composable selection.** The selector continues to pick `style_collections` rows and follow FKs. Picking palette + layout + typography independently is a later project.
- **LLM-driven layout generation.** Layouts are hand-designed in Phase 1. Later, adoption may produce new layouts via the adoption-fork flow, but the renderer doesn't care how a layout was created.
- **Automatic fork-quality scoring.** The review gate is human-driven. Automatic judgement on whether an adopted palette or layout is worth keeping is future work.
- **Theme preview UI.** A dashboard tool that renders a theme with a sample page for review would be useful but isn't in scope here.
- **Retroactive re-selection of live sites.** Sites that already have a `style_collection_id` keep using what they have. They render via the new path but get identical output. A future project may offer "re-theme this site" as an explicit action.

---

## 10. Risks

**Visual regression at cutover.** Mitigated by shadow-mode comparison in Phase 4. Any output diff for a live site is a hard blocker.

**Palette merge producing clashing colours.** Documented constraint; selection quality matters. Surface this in the contracts doc and in the HITL review UI for forked themes.

**Layout template errors producing broken CSS.** Each layout reviewed manually against contracts before seeding. Template helper funcs (`palette`, `typo`, `token`) prevent silent empty-string substitution.

**Adoption forking creating library noise.** Review gate prevents unreviewed forks from being selected. Over time, a periodic audit may prune rarely-selected forks — separate work.

**Old-column reads we missed.** Phase 8 (cleanup) gated on confidence that nothing reads the legacy columns. A grep pass across the codebase before dropping them.

---

## 11. What Needs Deciding Before We Start

1. **Final layout list.** The 10 above are a starting proposal. Does the final set include the `utility-tool`, `media-grid`, `docs-sidebar` layouts that no existing theme maps to? (My suggestion: yes, so the library has variety for adoption to match against.)

2. **Layout-drafting approach.** One layout at a time (design, review, next), or all 10 drafted in parallel with batch review?

3. **Typography set granularity.** A handful of shared typography sets (sans-modern, serif-editorial, display-bold, mono-technical) that most themes reference, or each theme gets its own typography row? Shared is simpler; per-theme is more flexible.

4. **Header/footer component handling during migration.** Forking adopted layouts may or may not want to fork header/footer components. For now the plan is "reuse existing components"; if adopted sites have distinctive headers, those become separate `needs_new_component` items handled by `component-creator`.

5. **Phase 4 shadow-mode duration.** How long do we want the old and new renderer paths running in parallel before cutover? One day? One week?
