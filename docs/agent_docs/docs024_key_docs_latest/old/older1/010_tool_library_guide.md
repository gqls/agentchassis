# 009 — Tool Library Guide

How tools are stored, deployed, owned, and maintained. Read this before creating, deploying, or modifying tools.

---

## Core Concept: Fork-on-Deploy

Tools live in two places:

**Library tools** — canonical templates in `content_components` with `component_level = 'tool'` and `forked_from IS NULL`. These are never directly referenced by any site page. They are blueprints.

**Site forks** — copies created when a tool is deployed to a site. `forked_from` points back to the library tool. The site's `page_components` references the fork, not the library original.

Once a tool is deployed to a site, it belongs to that site. Changes to the library tool do not cascade. If you improve a library tool and want it on existing sites, that means per-site work items — the same triage process as any other content change.

This is deliberate. A bad library change shouldn't break ten sites simultaneously.

---

## The Tool Library

### Where tools are stored

```sql
SELECT id, function, display_name, category,
       semantic_tags::text,
    CASE WHEN html_template = '' THEN 'NEEDS_TEMPLATE' ELSE 'has_template' END as status
FROM content_components
WHERE component_level = 'tool'
  AND forked_from IS NULL
  AND is_active = true
ORDER BY category, function;
```

### Content_components columns used by tools

| Column | Purpose |
|--------|---------|
| `name` | Internal identifier, e.g. `tool-ab-test-calculator` |
| `display_name` | Human-readable, e.g. `A/B Test Significance Calculator` |
| `function` | Unique key, same as `name` for tools |
| `category` | Grouping: `tool-calculator`, `tool-generator`, `tool-converter`, `tool-analyzer` |
| `component_level` | Always `'tool'` |
| `render_mode` | Always `'standalone'` — no template substitution, HTML is the product |
| `semantic_tags` | JSONB array for matching tools to sites, e.g. `["calculator", "marketing", "statistics"]` |
| `html_template` | The full tool markup: `<style>` + `<main>` + `<script>` |
| `input_schema` | Unused for tools (tools are self-contained), set to `'{}'` |
| `forked_from` | `NULL` = library tool, UUID = site fork |
| `is_active` | Soft delete |

### Template structure

Every tool template follows the same pattern. The site's `<head>` (with global.css), header, and footer are injected by `compile_page_sections` during rendering — the tool only provides the middle:

```html
<style>
    /* Tool-specific styles. Use CSS variables for brand colours: */
    .input-card {
        background: var(--color-surface, #fff);
        border: 1px solid var(--color-border, #ddd);
    }
    .result-box {
        background: var(--color-primary, #1e1e1e);
        color: var(--color-white, #fff);
    }
</style>

<main class="container" style="padding-top: var(--space-lg, 3rem);">
    <h1 style="font-size: var(--text-h2, 2rem);">Tool Title</h1>

    <!-- Guide box — always present, explains the tool to users -->
    <div class="guide-box" style="background: var(--color-surface, #fff); border: ...">
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
            <div>
                <h3>Concept 1</h3>
                <p>Why this matters and how to use it.</p>
            </div>
            <div>
                <h3>Concept 2</h3>
                <p>The second thing users need to know.</p>
            </div>
        </div>
    </div>

    <!-- Tool interface -->
    <div class="tool-layout">
        <aside><!-- Controls --></aside>
        <section><!-- Output / Canvas / Results --></section>
    </div>
</main>

<script>
    // Self-contained. No external API calls. No imports.
    // Use IIFEs or unique variable names to avoid collisions
    // if multiple tools appear on one page.
    (function() {
        // tool logic
    })();
</script>
```

### Design conventions

These patterns recur across all existing tools and should be maintained:

**Guide box** — two-column grid at the top explaining what the tool does and why. Uses `var(--color-accent)` for subheadings. This is educational content that also helps with SEO.

**Tool layout** — sidebar (controls) + main area (output). Responsive via `@media (min-width: 900px)` grid switch. Controls on the left, visual output on the right.

**Code output boxes** — dark background (`#1e1e1e`), monospace font, green text (`#a9ff68`) for generated code/output. Copy button included.

**CSS variables with fallbacks** — every `var(--color-*)` reference includes a fallback value so the tool works standalone during development: `var(--color-accent, #3b82f6)`.

**No external dependencies** — tools run entirely client-side. No API calls, no CDN imports, no server roundtrips. User data stays in the browser.

**Unique element IDs** — prefix IDs with something tool-specific (e.g. `entropyInput` not `input`) to prevent collisions if tools share a page.

---

## Current Tool Inventory

### Calculators (`tool-calculator`)

| Function | Display Name | Tags | Description |
|----------|-------------|------|-------------|
| `tool-ab-test-calculator` | A/B Test Significance Calculator | calculator, statistics, marketing, ab-testing, conversion | Z-score test for conversion rate differences. 95% confidence interval. |
| `tool-bayesian-ranking` | Bayesian Ranking Calculator | calculator, statistics, rating, ranking, ecommerce, reviews | Demonstrates why naive star averages mislead. Bayesian average with adjustable confidence constant. Two-product comparison. |

### Generators (`tool-generator`)

| Function | Display Name | Tags | Description |
|----------|-------------|------|-------------|
| `tool-favicon-generator` | Smart Favicon Generator | generator, favicon, design, icon, webdev | Emoji or upload → favicon.ico + apple-touch-icon.png. Browser tab preview, ICO binary construction in JS. |
| `tool-clip-path-builder` | CSS Clip-Path Builder | generator, css, clip-path, design, webdev, visual | Drag-to-edit polygon clip-path. Presets (triangle, hexagon, pentagon, ribbon). CSS output with copy. |
| `tool-meme-generator` | Meme Studio | generator, meme, image, social-media, fun | Canvas-based meme creator. Impact font, black/white text toggle, size slider, JPEG download. |
| `tool-prompt-architect` | AI Prompt Architect | generator, ai, prompt, midjourney, creative, photography | Midjourney prompt builder. Lighting, camera, style tag toggles. Aspect ratio and stylize params. |
| `tool-bg-remover` | Magic Background Eraser | generator, image, background-removal, photo-editing, design | Flood-fill magic wand + manual eraser brush. Tolerance slider, undo history, checkerboard transparency, PNG export. |

### Analyzers (`tool-analyzer`)

| Function | Display Name | Tags | Description |
|----------|-------------|------|-------------|
| `tool-password-entropy` | Password Strength Physics | calculator, security, password, entropy, privacy | Shannon entropy calculation, GPU crack time estimation, dictionary attack heuristic warning. |

---

## Deploying a Tool to a Site

### What happens during deployment

The `deploy_tool_to_site` action does four things:

1. **Loads** the library tool from `content_components`
2. **Checks** if this tool is already deployed (fork exists for the site)
3. **Forks** the tool — new `content_components` row with `forked_from` set
4. **Creates page** — new `pages` row at `/tools/{tool-name}.html` + `page_components` row linking the fork

After deployment, the page has `build_status = 'planned'` and needs the normal render/deploy pipeline to assemble and push it.

### Deploying manually via triage

To add a tool to a specific site:

```sql
-- 1. Find the library tool ID
SELECT id, function, display_name
FROM content_components
WHERE component_level = 'tool'
  AND forked_from IS NULL
  AND is_active = true;

-- 2. Create a triaged work item
INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
) VALUES (
             '<site_uuid>',
             'manual',
             'build',
             'add_tool',
             'low',
             'Add A/B Test Calculator tool',
             jsonb_build_object('tool_component_id', '<tool_uuid>'),
             80,
             'tool-deployer',
             'triaged',
             'admin',
             'add_tool:tool-ab-test-calculator'
         );
```

The `build-dispatch-loop` picks this up and spawns `tool-deployer`.

### Deploying automatically via discovery

The `missing_tools` check in `design-discovery-agent` suggests tools based on:

**Site type affinity** — `tools` sites get all tool categories, `ecommerce` sites get calculators and stats tools, `portfolio` sites get design tools.

**Industry affinity** — marketing sites get A/B testing and conversion tools, tech sites get CSS and design tools, photography sites get image editing tools.

**Universal tools** — security/privacy tools (password entropy) are suggested for all site types.

The check creates `add_tool` work items with `status = 'detected'`. These go through normal triage before deployment.

### Checking what's deployed where

```sql
-- Tools deployed to a specific site
SELECT
    cc.function,
    cc.display_name,
    cc.forked_from::text as library_tool_id,
    p.url as page_url,
    p.build_status
FROM content_components cc
         JOIN page_components pc ON pc.component_id = cc.id
         JOIN pages p ON pc.page_id = p.id
WHERE cc.component_level = 'tool'
  AND cc.forked_from IS NOT NULL
  AND p.site_id = '<site_uuid>';

-- All sites with a specific tool
SELECT
    s.domain,
    cc.id as fork_id,
    p.url,
    p.build_status
FROM content_components cc
         JOIN page_components pc ON pc.component_id = cc.id
         JOIN pages p ON pc.page_id = p.id
         JOIN sites s ON p.site_id = s.id
WHERE cc.forked_from = '<library_tool_uuid>';
```

---

## Creating a New Library Tool

### Step 1: Write the HTML

Follow the template structure above. Verify:

- Uses `var(--color-*)` with fallbacks for all brand colours
- No hardcoded hex colours on text elements (the forced_text_colors check will flag these)
- Has a guide-box explaining the tool
- Responsive layout that works on mobile (test at 375px width)
- Self-contained JS — no external APIs or CDN imports
- Unique element IDs (prefixed)
- IIFE or scoped variables to avoid global namespace pollution

### Step 2: Insert into the library

```sql
INSERT INTO content_components (
    name, display_name, function, category, component_level, render_mode,
    is_dark_section, is_active, description,
    semantic_tags, html_template, input_schema, forked_from
) VALUES (
             'tool-my-new-tool',
             'My New Tool',
             'tool-my-new-tool',
             'tool-calculator',         -- or tool-generator, tool-converter, tool-analyzer
             'tool',
             'standalone',
             false,
             true,
             'One-sentence description for meta/SEO.',
             '["tag1", "tag2", "tag3"]'::jsonb,
             '<style>...</style><main>...</main><script>...</script>',
             '{}'::jsonb,
             NULL                        -- NULL = library tool
         );
```

### Step 3: Tag it for discovery

Semantic tags control which sites the tool gets suggested for. Use a mix of:

- **Category tags**: `calculator`, `generator`, `converter`, `analyzer`
- **Domain tags**: `marketing`, `statistics`, `design`, `webdev`, `security`
- **Industry tags**: `ecommerce`, `photography`, `finance`, `education`
- **Feature tags**: specific to what the tool does, e.g. `ab-testing`, `clip-path`, `favicon`

The `matchToolToSite` function in the discovery check uses these tags against site type and industry to decide relevance.

---

## Modifying Tools

### Modifying a site's fork (normal case)

The site fork is a regular `content_components` row. Edit its `html_template` directly, then rerender the page:

```sql
-- Update the fork's template
UPDATE content_components
SET html_template = '<style>...</style><main>...</main><script>...</script>',
    updated_at = NOW()
WHERE id = '<fork_uuid>';

-- Mark the page for rerender
UPDATE page_components
SET build_status = 'pending',
    updated_at = NOW()
WHERE component_id = '<fork_uuid>';
```

Or create a work item for the improvement pipeline to handle it.

### Modifying the library tool (does NOT affect sites)

```sql
UPDATE content_components
SET html_template = '...',
    updated_at = NOW()
WHERE function = 'tool-ab-test-calculator'
  AND forked_from IS NULL;
```

This only changes the blueprint. New deployments get the updated version. Existing site forks are untouched.

### Pushing a library change to all sites

Create a work item per site. There is no cascade mechanism and this is intentional:

```sql
-- Create update work items for all sites that have this tool
INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
)
SELECT
    p.site_id,
    'manual',
    'build',
    'update_tool',
    'low',
    'Update ' || cc.display_name || ' to latest library version',
    jsonb_build_object(
            'fork_id', cc.id,
            'library_tool_id', cc.forked_from,
            'tool_function', cc.function
    ),
    90,
    'tool-deployer',
    'triaged',
    'admin',
    'update_tool:' || cc.function
FROM content_components cc
         JOIN page_components pc ON pc.component_id = cc.id
         JOIN pages p ON pc.page_id = p.id
WHERE cc.forked_from = '<library_tool_uuid>'
  AND cc.is_active = true;
```

Note: the `update_tool` item type would need a handler step added to `tool-deployer` — this isn't built yet. For now, update forks manually per site.

---

## Storage and Query Patterns

### How PostgreSQL handles large templates

Tool templates live in the `html_template` TEXT column. PostgreSQL automatically TOAST-compresses values over ~2KB and stores them out-of-line. A complex tool at 50-100KB is compressed to perhaps 15-30KB on disk. Storage is not a concern.

The concern is query I/O. When a query selects `html_template`, PostgreSQL must decompress and fetch the TOAST data for every matching row. For a query loading 8 components to check their `semantic_tags`, that's 8 unnecessary decompressions if templates aren't needed.

### Rule: never load html_template in listing or discovery queries

**Listing queries** (what tools exist, what's deployed where) should select only metadata:

```sql
-- CORRECT: metadata-only query for discovery/listing
SELECT id, function, display_name, category, semantic_tags::text
FROM content_components
WHERE component_level = 'tool'
  AND forked_from IS NULL
  AND is_active = true;
```

**Render queries** (building a page, fixing CSS) need the template and should load it:

```sql
-- CORRECT: full load for rendering
SELECT id, name, function, html_template, input_schema
FROM content_components
WHERE id = $1;
```

### Current query audit

| Query location | Loads template? | Justified? |
|----------------|----------------|------------|
| `queryComponentLibrary` | No | Listing — correct |
| `findMissingTools` | No | Discovery — correct |
| `loadComponentForEdit` | Yes | Editing — needs it |
| `loadComponentByFunction` | Yes | Rendering — needs it |
| `fixTemplateColors` | Yes | CSS fixing — needs it |
| `InjectHead/Header/Footer` | Yes | Page assembly — needs it |
| `LoadPageSectionComponentsAction` | Yes | Page building — needs it |

Note: `LoadPageSectionComponentsAction` has `include_templates` and `include_input_schema` config flags in its workflow step, but these are dead code — the query always loads both columns regardless. This is fine because the only caller is the page-content-writer which needs the full data. If a future caller needs metadata-only section loading, honour those flags or create a separate action.

### When to split template from component

For most tools (under 200KB), keep everything in `html_template`. If a tool grows beyond that — large embedded datasets, bundled libraries — split the JS into a separate file stored as an `asset`:

```html
<!-- Template stays small -->
<style>/* ... */</style>
<main class="container"><!-- markup --></main>
<script src="/tools/assets/mortgage-calculator.js"></script>
```

The JS file goes through the normal asset pipeline (assets table → S3/git deploy). The deploy step pushes both files. This keeps the content_component row lean while the asset system handles the heavy payload.

This isn't built yet and shouldn't be needed for current tools.

---

## Fork Naming

When a tool is forked for a site, the fork's `name` is constructed as:

```
{tool-name}-{full-domain-slug}
```

The full domain is preserved because sites may share a base name with different TLDs — `example.co.uk` and `example.uk` are different sites. Dots are replaced with hyphens:

| Domain | Fork name |
|--------|-----------|
| `website-design.com` | `tool-ab-test-calculator-website-design-com` |
| `website-design.co.uk` | `tool-ab-test-calculator-website-design-co-uk` |
| `website-design.uk` | `tool-ab-test-calculator-website-design-uk` |

The `domainSlug` function in `deploy_tool_action.go` handles this:

```go
func domainSlug(domain string) string {
    return strings.ReplaceAll(domain, ".", "-")
}
```

The `function` column is NOT slugged — forks keep the same function value as the library tool (`tool-ab-test-calculator`). The function identifies what the tool does; the name identifies whose copy it is.

---

## Tool Quality Standards

The improvement pipeline runs the same checks on tool pages as regular pages:

| Check | What it catches |
|-------|----------------|
| `hardcoded_section_colors` | Background hex values that should be CSS variables |
| `forced_text_colors` | Text colour declarations that fight the inheritance model |
| `missing_css` | Pages missing the site's global stylesheet link |
| `undeployed_assets` | Referenced images not yet pushed to storage |

Additional checks worth building (not yet implemented):

- **Mobile responsiveness** — does the tool have a `@media` breakpoint? Are touch targets ≥ 44px?
- **Accessibility** — do inputs have labels? Are colour contrasts AA compliant?
- **JS isolation** — are variables scoped? (IIFE or unique names)
- **Offline capability** — does the tool work without network? (should always be yes)

---

## Agent and Action Reference

### tool-deployer agent

Handles `add_tool` work items from the build-dispatch-loop.

**Workflow:** `load_item → check_has_item → deploy_tool → complete_item → complete`

**Input:** `site_id` (from dispatch), tool details from work item spec.

**Output:** `deploy_result` containing `fork_id`, `page_id`, `page_url`, `needs_rerender`.

### deploy_tool_to_site action

The Go action that does the actual work.

**Inputs (ActionInputSpec):**
- `site_id` (required) — target site
- `tool_component_id` (required) — library tool to fork
- `page_name` (optional) — override auto-generated page name
- `page_title` (optional) — override auto-generated page title

**Config:**
- `nav_section` (default `"Tools"`) — nav group label
- `in_header` (default `true`)
- `in_footer` (default `false`)

**Behaviour:**
1. Loads library tool from `content_components`
2. Checks if already deployed (returns early with `already_deployed: true`)
3. Forks: `INSERT INTO content_components` with `forked_from` set
4. Creates page at `/tools/{name}.html` with `page_type = 'tool'`
5. Creates `page_components` record linking fork to page
6. Returns result for work item completion

### findMissingTools discovery check

Added to `design-discovery-agent` as the `"missing_tools"` check.

**Logic:**
1. Loads site's industry and type from `site_specs` (falls back to `sites` table)
2. Queries library tools not already forked for this site
3. Matches tools using `matchToolToSite` (semantic tags vs site type/industry affinity)
4. Creates one `add_tool` work item per matched tool, `status = 'detected'`, `priority = 120`

---

## Pipeline Flow

```
                          ┌─────────────────────┐
                          │   Library Tool       │
                          │   (content_component │
                          │   forked_from=NULL)  │
                          └─────────┬───────────┘
                                    │
                     ┌──────────────┼──────────────┐
                     │              │              │
               Manual triage   Discovery check   API/Admin
                     │              │              │
                     ▼              ▼              ▼
              ┌──────────────────────────────────────┐
              │   site_work_items                     │
              │   item_type = 'add_tool'              │
              │   handler_agent = 'tool-deployer'     │
              │   spec.tool_component_id = <uuid>     │
              └─────────────────┬────────────────────┘
                                │
                          triage promotes
                                │
                                ▼
              ┌──────────────────────────────────────┐
              │   build-dispatch-loop                  │
              │   spawns tool-deployer                 │
              └─────────────────┬────────────────────┘
                                │
                                ▼
              ┌──────────────────────────────────────┐
              │   tool-deployer                       │
              │   1. Fork content_component           │
              │   2. Create page                      │
              │   3. Create page_component            │
              │   4. Complete work item               │
              └─────────────────┬────────────────────┘
                                │
                          needs_rerender
                                │
                                ▼
              ┌──────────────────────────────────────┐
              │   rerender-pages                      │
              │   compile_page_sections injects:      │
              │   - <head> with global.css            │
              │   - site header                       │
              │   - site footer                       │
              │   Tool template provides:             │
              │   - <style> (with CSS var refs)       │
              │   - <main> (tool UI)                  │
              │   - <script> (tool logic)             │
              └─────────────────┬────────────────────┘
                                │
                                ▼
              ┌──────────────────────────────────────┐
              │   git commit → GitHub Actions → S3    │
              │   Tool live at /tools/{name}.html     │
              └──────────────────────────────────────┘
```

---

## Ideas for Future Tools

Organised by likely site-type affinity. These are starting points — the tool-creator agent (not yet built) would generate these from descriptions.

### Broadly useful
- Colour contrast checker (WCAG AA/AAA)
- Unit converter (px/rem/em, weight, temperature)
- Lorem ipsum generator (with paragraph/word count)
- QR code generator
- JSON formatter / validator
- Regex tester with match highlighting
- Countdown timer / event date calculator
- Aspect ratio calculator

### Marketing / Business sites
- ROI calculator
- Headline analyser (word count, power words, emotional score)
- Email subject line tester
- UTM link builder
- Social media image size reference

### Ecommerce sites
- Shipping cost estimator (weight × zone)
- Markup / margin calculator
- Product comparison table builder
- Discount / sale price calculator

### Finance sites
- Mortgage calculator
- Compound interest calculator
- Loan amortisation table
- Currency converter (static rates, or with API if extended)
- Tax bracket estimator

### Health / Fitness sites
- BMI calculator
- Calorie calculator (TDEE)
- Pace calculator (running / cycling)
- Water intake calculator

### Property / Construction sites
- Tile calculator (area → tiles needed, with waste %)
- Paint calculator (wall area → litres needed)
- Room area / volume calculator
- Flooring calculator

### Education sites
- Grade calculator (weighted average)
- GPA calculator
- Study time planner
- Citation formatter

### Design / Dev sites
- Box shadow generator
- Gradient generator
- Border radius previewer
- SVG path editor
- Base64 encoder/decoder
- Colour palette generator (complementary, analogous, triadic)

Each of these follows the same pattern: self-contained HTML+CSS+JS, guide-box explaining the concept, sidebar controls + main output area, CSS variables for branding.

