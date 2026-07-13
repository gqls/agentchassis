# Focus: Design & Styling — Adoption, Colour Choices, Implementation, Discovery & Fix

This document pulls together everything related to site visual design from across the system. Use it as a single reference when debugging or improving how sites look.

---

## Current State: What's Wrong

The design pipeline is the weakest part of the system right now. The symptoms:

- **Adopted sites get generic styles.** The adoption pipeline correctly captures identity and content direction but the build pipeline ignores crawled CSS and applies generic brochure styles instead.
- **Hardcoded hex colours** appear in components that should use CSS variables, causing light-on-light or dark-on-dark text.
- **Dark sections break** because base CSS sets `color: var(--color-text)` directly on elements, bypassing the `--section-*` override mechanism.
- **Missing CSS entirely** — sites deploy with no `/assets/css/styles.css`, rendering with browser defaults.
- **Generic themes** — default style collection applied with no industry or brand customisation.
- **Fallback headers** — when component linkage fails, a hardcoded Go function produces ugly generic HTML.
- **Design intent not flowing downstream** — the classifier produces `design_intent` and `visual_direction` specs but the webdesign-agent sometimes doesn't read them, generating CSS from scratch based on industry name alone.

---

## 1. The Design System: Three Independent Layers

### Layer 1: HTML Components (structure)

**Table:** `content_components`

Self-contained HTML blocks (hero, testimonials, FAQ, etc.). Each has inline `<style>` for layout (grid, flexbox, spacing) and dark section overrides. Multiple variants exist per function (e.g. `hero-split`, `hero-fullwidth`, `hero-minimal`).

Components reference CSS variables with fallbacks: `var(--color-primary, #1a1a2e)`. They **never** hardcode brand colours. Dark sections set `color: #fff` on their container.

### Layer 2: CSS Theme (appearance)

**Table:** `css_themes`

A complete base stylesheet setting `:root` variables (colours, fonts, spacing), base resets, typography scale, button styles, responsive breakpoints, and accessibility focus states.

Deployed as `/assets/css/styles.css` — one per site, committed to git by the webdesign-agent.

### Layer 3: Style Collection (the bridge)

**Table:** `style_collections`

A named grouping tying together header component + footer component + CSS theme + colour palette + typography. Think of it as a "design kit."

```
sites.style_collection_id → style_collections (e.g. "professional-dark")
    ├── header_component_id → content_components (header-professional-dark)
    ├── footer_component_id → content_components (footer-professional-dark)
    ├── css_theme_id → css_themes (→ /assets/css/styles.css)
    ├── color_palette → {primary: "#1a1a2e", accent: "#0f3460", ...}
    └── typography → {font_family: "Inter, sans-serif", heading_font: "Playfair Display", ...}
```

### How they vary independently

- Change the **theme** (Layer 2) → same components, different colours/fonts
- Change a **component** (Layer 1) → same look, different layout
- Change the **collection** (Layer 3) → different header/footer, different theme, different palette

---

## 2. CSS Colour Inheritance Model (the contract)

This is the single most important design contract. When it breaks, text becomes unreadable.

### How it should work

```
/assets/css/styles.css (from css_themes)
  body { color: var(--color-text); }
  h1-h6 { color: var(--section-heading, var(--color-primary)); }
  p, li, blockquote { color: var(--section-text, inherit); }
  strong, em, cite, span — do NOT set color (they inherit from parent)

Light section (default):
  h2 gets var(--section-heading) → not set → fallback var(--color-primary) → #1a1a2e
  p gets var(--section-text) → not set → inherit → var(--color-text) → #333

Dark section component sets --section-* on container:
    --section-heading: #ffffff;
    --section-text: rgba(255,255,255,0.9);
  h2 gets var(--section-heading) → #ffffff
  p gets var(--section-text) → rgba(255,255,255,0.9)
```

### The rules

- `body` sets `color: var(--color-text)` — the base default
- `h1-h6` use `color: var(--section-heading, var(--color-primary))` — prominent in light, white in dark
- `p`, `li`, `blockquote` use `color: var(--section-text, inherit)` — adapts to section context
- `strong`, `em`, `cite`, `span` — do NOT set `color` at all (inherit from parent)

### How it breaks

If the base CSS sets `color: var(--color-text)` directly on `p` or `h1`, the `--section-*` override is bypassed. The element gets `#333333` regardless of what the dark section container sets. **This is the bug that causes light-on-light text in testimonial sections.**

Similarly, if the base CSS sets `color: var(--color-primary)` on `h1` instead of `var(--section-heading, var(--color-primary))`, dark sections can't override headings to white.

### Dark section contract

Components with `is_dark_section = true` MUST set these CSS variables on their root container:

```css
.my-dark-section {
    background: var(--color-primary, #1a1a2e);
    color: #fff;

    /* Section context overrides */
    --section-heading: #ffffff;
    --section-text: rgba(255, 255, 255, 0.9);
    --section-link: var(--color-accent, #e94560);
    --section-muted: rgba(255, 255, 255, 0.7);
}
```

### CSS variables available to components

```
Colours: --color-primary, --color-primary-hover, --color-primary-text,
  --color-secondary, --color-accent, --color-text, --color-text-muted,
  --color-heading, --color-background, --color-surface, --color-card-bg,
  --color-border, --color-header-bg, --color-header-text,
  --color-footer-bg, --color-footer-text, --color-white

Section context: --section-heading, --section-text, --section-link, --section-muted

Spacing: --spacing-sm, --spacing-md, --spacing-lg, --spacing-xl
Typography: --font-family, --heading-font, --base-size, --line-height
```

---

## 3. How Design Gets Created

### New build flow

```
domain-research-classifier
  → writes site_specs: identity, classification, content_direction, design_intent
  → design_intent contains: style direction, colour mood, typography, imagery, layout

site-planner
  → reads design_intent + identity
  → selects style collection (or creates work item for custom design)
  → creates needs_design work item (priority 8)

webdesign-agent (handles needs_design)
  → load_site_for_design: site info, pages, components, colours, typography
  → analyze_design: LLM generates design spec from identity + design_intent
  → generate_css: LLM produces CSS theme
  → deploy_css: git commit → GitHub Actions → S3 → live
  → writes CSS to site_components and css_themes
```

### What the classifier should produce for design_intent

```json
{
  "style": "professional, modern, trustworthy",
  "colour_mood": "deep blues and warm accents, not clinical",
  "typography": "serif headings for authority, sans-serif body for readability",
  "imagery": "professional photography, no stock photos of handshakes",
  "layout": "content-heavy with comparison tables, prioritise readability"
}
```

### What the webdesign-agent produces

A CSS theme deployed as `/assets/css/styles.css` containing:
- `:root` variables for all colours, fonts, spacing
- Base resets and typography scale
- Button styles
- Responsive breakpoints
- Accessibility focus states
- Dark section variable declarations

---

## 4. The Adoption Design Gap

### What should happen

When adopting an existing site, the crawler captures the site's visual identity. The build pipeline should reproduce (or improve on) the original design.

### What actually happens

The adoption pipeline (`analyze_site` LLM step) produces:
- Identity spec (who the company is)
- Design spec (palette, typography, layout notes)
- Content direction (writing style)
- Page structure

But the build pipeline:

1. **Ignores crawled CSS and applies generic styles** — the webdesign-agent generates CSS from scratch based on identity, not from the captured design data
2. **Can't reproduce JavaScript applications** via the content writer (mitigated by tool-recreation-handler)
3. **Uses generic brochure components** for custom layouts
4. **Improvement loop audits against generic standards** rather than the adopted site's character

### What needs building

| Priority | Item | Status |
|---|---|---|
| 1 | **Design fingerprint extraction** — extract CSS variables, colour palette, font stacks from crawled HTML and pass to webdesign-agent as `adopt_from` context | Not started |
| 2 | **Archetype-aware design** — the site archetype (character, visual density, polish level) should constrain the webdesign-agent's CSS generation | Archetype classification exists but isn't persisted as a spec (Go patch needed) |
| 3 | **Adoption-aware improvement loop** — respect the archetype's `constraints` field (things the improvement loop must never change) | Not started |

### The design fingerprint idea

When crawling a site, extract:
- Computed colour palette (most-used background colours, text colours, accent colours)
- Font families from `<link>` tags and `font-family` declarations
- Layout patterns (grid vs flex, max-width, spacing rhythm)
- Dark/light section detection

Pass this as `adopt_from.design` in the `needs_design` work item spec. The webdesign-agent receives it and generates CSS that matches the original palette rather than inventing one.

---

## 5. Discovery Checks for Design

### design-discovery-agent (algorithmic)

| Check | What it detects | Handler | How it works |
|---|---|---|---|
| `hardcoded_section_colors` | Inline hex colours that should use CSS variables | `color-variable-fixer` | Scans `rendered_html` for `color: #hex` and `background: #hex` patterns |
| `forced_text_colors` | Child text elements with hardcoded `color: #hex` | `color-variable-fixer` | Specifically catches `<p style="color:#...">` and similar |
| `missing_css` | Site with no `/assets/css/styles.css` deployed | `webdesign-agent` | Checks git repo for the file |
| `generic_theme` | Default theme with no customisation | `webdesign-agent` | Detects when style collection is the system default |
| `validate_component_standards` | Unlinked site components, slot mismatches | `site-component-linker` / `component-template-fixer` | SQL queries against site_components |

### design-audit-agent (LLM-based orchestrator)

Groups checks by shared context (CSS theme, colour palette, rendered HTML) into one LLM call per group. UP TO 5 findings per pass.

**visual-design-auditor:**
- Loads: style collection, CSS theme, colour palette, rendered HTML samples (locked/unexpired excluded)
- LLM checks: colour consistency, spacing, typography, dark sections, responsive
- Respects direction spec — doesn't flag requested visual features

### How algorithmic and LLM checks interact

The algorithmic check catches obvious cases: hardcoded hex `#1a1a2e` that should be `var(--color-primary)`. The LLM catches subtle cases: "this shade of blue doesn't match the overall warm tone of the palette."

**Ordering matters:** `validate_component_standards` runs BEFORE `hardcoded_section_colors` and `forced_text_colors` because those operate on rendered HTML which may be wrong if components aren't linked properly. Fix structural issues first, then audit the content.

---

## 6. Fix Agents for Design

### webdesign-agent
Handles `needs_design`, `needs_design_review`, `missing_css`, `generic_theme`. Self-contained: receives `site_id` + `domain`, loads its own context, generates CSS, deploys to git. Writes CSS to `site_components` directly — can be used as a dispatch handler.

Workflow: `check_site_context → load_site_for_design → analyze_design → generate_css → deploy_css → complete`

### color-variable-fixer
Handles `hardcoded_section_colors`, `forced_text_colors`. Replaces hex values with CSS variable references in `rendered_html`.

### component-template-fixer
Routes on `spec.fix_type`:
- `inject_nav_flex_css` — adds flex CSS to nav
- `remove_element` — removes unwanted elements
- Various other CSS/HTML patches

### site-component-linker
Fixes header/footer/head components not linked to style collection. The root cause of most fallback rendering issues.

---

## 7. The Design Intent Chain

The intended flow from classification to deployed CSS:

```
Classifier → design_intent spec (what the site should feel like)
  ↓
Planner → selects style collection + creates needs_design item
  ↓
Webdesign-agent → reads design_intent + identity + style collection
  → generates CSS theme matching the intent
  → deploys to git
  ↓
Audit agent → checks if deployed CSS matches design_intent
  → creates work items for drift (colour_fix, typography_fix, etc.)
  ↓
Same webdesign-agent → handles the fix items
```

**Where it breaks:**
- Classifier doesn't produce `design_intent` (older sites, adoption pipeline)
- Webdesign-agent doesn't read `design_intent` from specs (falls back to industry-name guessing)
- Audit agent has no design_intent to audit against → proposes its own direction

**Exception for adopted sites:** When no design_intent exists, the audit agent switches to "propose" mode rather than "enforce" mode. Work items get flagged for HITL review.

---

## 8. The Spec-to-CSS Pipeline

### What specs drive design

| Spec aspect | Written by | Read by | Contains |
|---|---|---|---|
| `identity` | classifier | webdesign-agent, auditors | company info, industry, tone |
| `design_intent` | classifier | webdesign-agent | style direction, colour mood, typography, imagery |
| `visual_direction` | classifier, HITL | webdesign-agent, planner | palette preferences, layout direction |
| `content_direction` | classifier, strategist | content writers (not design) | voice, emphasis — NOT visual |

### Style collection selection

The planner selects a style collection based on design_intent:

```
sites.style_collection_id → style_collections (e.g. "professional-dark")
```

Available collections are matched by industry, style tags, and colour characteristics. If none match, the planner creates a `needs_design` work item for the webdesign-agent to generate custom CSS.

### CSS theme generation

The webdesign-agent's LLM prompt receives:
- Site identity (company, industry, tone)
- Design intent (colours, typography, imagery direction)
- Page list (what types of content need styling)
- Available components (what CSS classes need to exist)

It produces a complete `:root` variable block plus all element and component styles. This gets stored in `css_themes` and deployed as `/assets/css/styles.css`.

---

## 9. Diagnostic Queries

### Does the site have CSS deployed?

```sql
SELECT sc.slot_name, LENGTH(sc.rendered_html) as size_bytes,
       sc.component_id IS NOT NULL as has_template,
       sc.build_status
FROM site_components sc
WHERE sc.site_id = '<site_id>'
  AND sc.slot_name IN ('header', 'footer', 'head')
ORDER BY sc.slot_name;
```

### What style collection is the site using?

```sql
SELECT s.domain, sc.name as collection_name,
       sc.color_palette, sc.typography,
       ct.id as css_theme_id
FROM sites s
LEFT JOIN style_collections sc ON sc.id = s.style_collection_id
LEFT JOIN css_themes ct ON ct.id = sc.css_theme_id
WHERE s.id = '<site_id>';
```

### Are there hardcoded colours in components?

```sql
SELECT pc.slot_name, p.name as page_name,
       (regexp_matches(pc.rendered_html, 'color:\s*#[0-9a-fA-F]{3,8}', 'g'))[1] as hardcoded_color,
       COUNT(*) as occurrences
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '<site_id>'
  AND pc.rendered_html ~ 'color:\s*#[0-9a-fA-F]{3,8}'
GROUP BY pc.slot_name, p.name, 3
ORDER BY occurrences DESC;
```

### Does the site have design-related specs?

```sql
SELECT aspect, source, LEFT(data::text, 120) as preview,
       created_at, is_current
FROM site_specs
WHERE site_id = '<site_id>'
  AND aspect IN ('design_intent', 'visual_direction', 'identity', 'design')
ORDER BY aspect, created_at DESC;
```

### Dark section components missing variables?

```sql
SELECT cc.function, cc.is_dark_section,
       CASE WHEN cc.html_template LIKE '%--section-heading%' THEN 'has vars' ELSE 'MISSING VARS' END as section_vars
FROM content_components cc
WHERE cc.is_dark_section = true
ORDER BY cc.function;
```

### Design-related work items?

```sql
SELECT item_type, status, handler_agent, LEFT(summary, 80), attempt_count, created_at
FROM site_work_items
WHERE site_id = '<site_id>'
  AND (handler_agent IN ('webdesign-agent', 'color-variable-fixer', 'component-template-fixer', 'site-component-linker')
       OR item_type IN ('needs_design', 'needs_design_review', 'missing_css', 'generic_theme',
                        'hardcoded_section_colors', 'forced_text_colors'))
ORDER BY created_at DESC;
```

### Components using fallback rendering?

```sql
SELECT sc.slot_name,
       CASE
           WHEN sc.rendered_html LIKE '%search-toggle%' THEN 'FALLBACK (has search icon)'
           WHEN sc.rendered_html LIKE '%RenderFallback%' THEN 'FALLBACK (explicit)'
           WHEN sc.component_id IS NULL THEN 'UNLINKED (will use fallback)'
           ELSE 'template-rendered'
       END as render_source,
       LENGTH(sc.rendered_html) as size
FROM site_components sc
WHERE sc.site_id = '<site_id>';
```

---

## 10. Fix Sequence for a Site with Bad Design

### Phase 1: Structural (fix first — everything else depends on this)

1. **Check component linkage** — are header/footer/head linked to the style collection?
2. **Check style collection** — does it have `css_theme_id`, `color_palette`, `typography` set?
3. **Fix unlinked components** — trigger `site-component-linker` or manually set `component_id`
4. **Verify no fallback rendering** — check for search icons, stacked nav

### Phase 2: CSS Theme

5. **Check design_intent spec exists** — if not, either run the classifier or manually create one
6. **Check if CSS is deployed** — look for `/assets/css/styles.css` in the git repo
7. **Trigger webdesign-agent** — create a `needs_design_review` work item
8. **Verify CSS variable declarations** — `:root` block has all expected variables

### Phase 3: Component Compliance

9. **Check hardcoded colours** — run `hardcoded_section_colors` discovery check
10. **Check dark sections** — verify `--section-*` variables set on all `is_dark_section` components
11. **Check forced text colours** — run `forced_text_colors` discovery check
12. **Fix or re-render components** — create work items for the fixes

### Phase 4: Verify

13. **Re-render all pages** — trigger `needs_rerender` with `refresh_site_components: true`
14. **Visual check** — look at the deployed site
15. **Run improvement loop** — verify audit findings reduce

### Quick health check query

```sql
SELECT
    s.domain,
    CASE WHEN s.style_collection_id IS NOT NULL THEN 'set' ELSE 'MISSING' END as style_collection,
    (SELECT COUNT(*) FROM site_specs WHERE site_id = s.id AND aspect = 'design_intent' AND is_current = true) as has_design_intent,
    (SELECT COUNT(*) FROM site_components WHERE site_id = s.id AND slot_name = 'head' AND rendered_html LIKE '%:root%') as has_css_vars,
    (SELECT COUNT(*) FROM site_components WHERE site_id = s.id AND component_id IS NULL) as unlinked_components,
    (SELECT COUNT(*) FROM site_work_items WHERE site_id = s.id AND handler_agent = 'webdesign-agent' AND status = 'failed') as failed_design_items
FROM sites s
WHERE s.id = '<site_id>';
```
