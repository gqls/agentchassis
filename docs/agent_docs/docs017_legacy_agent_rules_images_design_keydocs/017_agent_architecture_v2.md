https://claude.ai/chat/fbdaef1b-bb4c-45dd-88e5-34349bfe27bf
# Design System Architecture

## The Three Independent Layers

### Layer 1: HTML Components (structure & layout)
**Table:** `content_components`
**What they are:** Self-contained HTML blocks — a hero section, a testimonial grid, a FAQ accordion, a services card layout, a CTA banner. Each has its own inline `<style>` for layout (grid columns, flexbox, padding, dark section overrides).
**What varies:** A "hero" might be full-bleed background image, or split-panel with text left/image right, or minimal centered text. Multiple components exist for the same function (e.g. `hero-split`, `hero-fullwidth`, `hero-minimal`).
**How they reference colours:** Via CSS variables with fallbacks: `var(--color-primary, #1a1a2e)`. They never hardcode brand colours.
**Dark sections:** Components that need light text (testimonials on dark bg, CTA sections, footers) set `color: #fff` on their container. All children inherit. This is why the base CSS must not force `color: var(--color-text)` on child elements.

### Layer 2: CSS Theme (appearance)
**Table:** `css_themes`
**What it is:** A complete base stylesheet. Sets `:root` variables (colours, fonts, spacing), base resets (box-sizing, margin), typography scale (h1-h6 sizes), button styles, responsive breakpoints, accessibility focus states.
**What varies:** A "finance-dark" theme has navy/gold palette, tight spacing, system fonts. A "creative-bold" theme has vibrant colours, generous spacing, display fonts. A "minimal-light" theme has muted tones, lots of whitespace, thin type.
**Key rule:** Sets `body { color: var(--color-text) }` as the single source of default text colour. All other elements inherit. Headings use `color: inherit`, not `color: var(--color-primary)`.
**Deployed as:** `/assets/css/styles.css` — one per site, committed to the site's git repo by the webdesign agent.

### Layer 3: Style Collection (the bridge)
**Table:** `style_collections`
**What it is:** A named grouping that ties together a header component + footer component + CSS theme + colour palette. Think of it as a "design kit."
**Examples:** `professional-dark`, `modern-light`, `bold-creative`, `clean-minimal`.
**What it references:**
- `header_component_id` → which header HTML component to use
- `footer_component_id` → which footer HTML component to use
- `css_theme_id` → which CSS theme to use
- `color_palette` → JSON with primary, secondary, accent, background, text colours

## How They Connect

```
site (leopardessconsulting.co.uk)
  └── style_collection_id → style_collections (professional-dark)
        ├── header_component_id → content_components (header-professional-dark)
        ├── footer_component_id → content_components (footer-4-column)
        ├── css_theme_id → css_themes (base stylesheet → /assets/css/styles.css)
        └── color_palette → {primary: "#1a1a2e", accent: "#0f3460", ...}

  └── pages (from site plan)
        ├── index.html
        │     ├── hero (content_components: hero-fullwidth)
        │     ├── differentiators (content_components: differentiators-grid)
        │     ├── services (content_components: services-grid)
        │     ├── testimonials (content_components: social-proof)
        │     └── cta (content_components: call-to-action)
        ├── about.html
        │     ├── hero (same or different hero component)
        │     ├── team-section (content_components: team-grid)
        │     └── cta (content_components: call-to-action)
        └── services.html
              └── ...different body components...

  All pages share:
    - Same header (from style collection)
    - Same footer (from style collection)
    - Same /assets/css/styles.css (from CSS theme)
    - Same :root variables (from colour palette)
```

## What Each Agent Does

**Site Planner** — decides which pages exist and which body components go on each page. Picks the style collection. Outputs a site plan with page list and component assignments.

**Webdesign Agent** — analyses the brand/industry, generates a design spec (colour scheme, typography, spacing), then generates the CSS theme. Stores it in `css_themes` and deploys as `styles.css`. Over time, reuses existing themes instead of regenerating.

**Content Writer** — fills component templates with actual copy. Receives the page's component list and render context (company name, nav items, etc). Outputs populated HTML.

**Pageflow Builder** — orchestrates the build. Runs sync_pages_to_db → populate_nav → render components → assemble pages → deploy. Calls InjectHeader/InjectFooter which read from nav tables and render the header/footer components.

## The Inheritance Chain (why it matters)

```
/assets/css/styles.css (from css_themes)
  body { color: #333; }              ← default dark text for light sections
    ↓ inherits
    h1, h2, p, li, blockquote        ← all dark text (no explicit color set)

Component inline <style>
  .social-proof-section { color: #fff; }  ← dark section overrides to light
    ↓ inherits
    h2, blockquote, cite, p, strong       ← all light text automatically

If styles.css forces: p { color: #333; }
  → p inside dark sections stays dark    ← BROKEN, unreadable
```

## How the Theme Library Grows

1. Build a site → webdesign agent generates CSS → stored as new `css_themes` row
2. Tag it with industry, style category, colour characteristics
3. Next similar brief → search existing themes → reuse if match found
4. Collect inspiration from external sites → extract their patterns → inform design spec → generate or select theme
5. Library grows from "generate every time" to "curate and select"

## Mix and Match Examples

Same components, different themes:
- Hero + services-grid + testimonials on `professional-dark` → navy bg, system fonts, tight
- Hero + services-grid + testimonials on `creative-bold` → coral accent, display fonts, generous

Same theme, different components:
- `professional-dark` + hero-fullwidth + testimonials-grid → big hero, card layout
- `professional-dark` + hero-minimal + testimonials-carousel → text hero, scrolling quotes

Same theme, different pages:
- Site A: 5 pages (home, about, services, case-studies, contact)
- Site B: 3 pages (home, services, contact) with different body components on each