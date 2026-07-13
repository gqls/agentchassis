# 019 — Tool Library

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

The `missing_tools` check in `design-discovery-agent` does not evaluate tool fit — it only checks whether the site has any tools deployed and whether a recent evaluation exists (7-day cooldown). If both are no, it creates a single `evaluate_tools` work item for `tool-suggester`.

Tool-suggester then uses LLM judgment to evaluate what tools would genuinely help the site's visitors, considering industry, audience, services, and existing pages. It can suggest library tools (forked by `tool-deployer`) or novel tools built from scratch (generated by `tool-generator`). It can also suggest zero tools if none are appropriate.

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

Semantic tags are stored in `content_components.semantic_tags` as a JSONB array. They are loaded by tool-suggester's `load_library_tools` step and included in the LLM prompt so the model can see what each library tool does. Tags are not used for automated matching — tool selection is an LLM judgment call.

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

### Three-tier quality pipeline

Tool quality is checked automatically on the improvement sweep through `design-discovery-agent`:

**Tier 1 — Structural checks** (`tool_health` discovery check, Go, zero cost):

| Check | Severity | What it catches |
|-------|----------|----------------|
| `not_deployed` | blocker | Page `build_status` not `deployed`/`active` |
| `no_rendered_html` | blocker | Empty `page_component.rendered_html` |
| `empty_template` | blocker | Fork `html_template` is empty (forked from empty library tool) |
| `no_script` | error | No `<script>` block — tool can't be interactive |
| `no_style` | warning | No `<style>` block |
| `no_responsive` | warning | No `@media` breakpoint — may break on mobile |
| `hardcoded_colors` | warning | >3 bare hex values outside `var()` fallbacks |
| `external_fetch` | warning | Uses `fetch()` — should be self-contained |
| `external_cdn` | warning | References external CDN |

Blockers create `improve_tool` → `tool-improver`. Other issues are bundled into the LLM audit.

**Tier 2 — LLM code review** (`tool-auditor` agent, Sonnet, ~1 call per tool):

| Category | What the LLM checks |
|----------|---------------------|
| `js_bug` | Uninitialised variables, missing event listeners, division by zero, DOM reference mismatches, dead code paths |
| `mobile` | Layout at 375px, touch targets ≥44px, clipped/overflowing elements, hover-only interactions |
| `ux` | Clear first action, interaction feedback, visible labels, working copy/download |
| `css` | Hardcoded colours, missing CSS variables, `!important` conflicts |
| `accessibility` | Input labels, alt text, keyboard operability |
| `dependency` | External APIs, CDN references, implicit dependencies on parent page, ID collisions |

Certain/likely findings → `improve_tool` (auto-fix). Possible findings → `needs_human_review` (HITL).

**Tier 3 — Visual testing** (planned, `tool-visual-tester` agent, headless browser):

Would render the tool page at multiple viewports (375px, 768px, 1280px), take screenshots, and check for JS console errors. Requires its own Kubernetes pod with Chromium. Not yet built.

### Existing checks from design-discovery-agent

The regular design checks also apply to tool pages:

| Check | What it catches |
|-------|----------------|
| `hardcoded_section_colors` | Background hex values that should be CSS variables |
| `forced_text_colors` | Text colour declarations that fight the inheritance model |
| `missing_css` | Pages missing the site's global stylesheet link |
| `undeployed_assets` | Referenced images not yet pushed to storage |

---

## Agent and Action Reference

### tool-deployer agent

Handles `add_tool` work items where `tool_component_id` is present (library fork).

**Workflow:** `ensure_site_record → deploy_tool → complete`

**Input:** `site_id`, `domain` (from dispatch), tool details from work item spec.

**Output:** `deploy_result` containing `fork_id`, `page_id`, `page_url`, `guide_url`, `needs_rerender`.

### tool-generator agent

Handles `add_tool` work items where `tool_component_id` is null (novel tool, no library match).

**Workflow:** `ensure_site_record → load_brand_context → generate_tool_html → save_tool → complete`

**Input:** `site_id`, `domain`, `spec` (with `name`, `function`, `description`, `complexity`).

**Output:** `create_result` containing `component_id`, `page_id`, `page_url`, `guide_url`, `generated: true`.

The `generate_tool_html` step sends the tool description and site context to Sonnet, which produces a self-contained HTML fragment. The `save_tool` step calls `create_tool_component` which creates the component, page, page_component (with rendered_html), nav entry, content work items, and companion guide.

### deploy_tool_to_site action

The Go action for library fork deployment.

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
5. Creates `page_components` record linking fork to page (position 2, with `rendered_html`)
6. Creates `needs_content_page` work item for tool page (hero, intro, CTA)
7. Creates companion guide page + `needs_content_page` work item
8. Adds nav entry under "Tools" group

### create_tool_component action

The Go action for novel tool creation (called by tool-generator).

**Inputs (ActionInputSpec):**
- `site_id` (required)
- `html_content` (required) — LLM-generated HTML
- `function` (required) — kebab-case tool identifier
- `display_name` (required)
- `description` (optional)
- `category` (optional, default `"interactive"`)

**Behaviour:** Same outputs as `deploy_tool_to_site` but creates the component from scratch instead of forking. Strips markdown code fences via `datahelpers.StripCodeFences`. Sets `created_from = 'generated'`.

### missing_tools discovery check

Added to `design-discovery-agent` as the `"missing_tools"` check.

**Logic:**
1. Counts deployed tools for the site (via `content_components` + `page_components`)
2. Checks for recent `evaluate_tools` work items (7-day cooldown)
3. If no tools and no recent evaluation → creates single `evaluate_tools` item with `handler_agent: tool-suggester`

Does not evaluate tool fit, does not look at library tools, does not match by tags. All evaluation is delegated to tool-suggester's LLM step.

---

## Pipeline Flow

```
                          ┌─────────────────────┐
                          │   tool-suggester     │
                          │   (LLM evaluation)   │
                          └─────────┬───────────┘
                                    │
                     ┌──────────────┴──────────────┐
                     │                             │
              tool_component_id              tool_component_id
                 != null                        == null
                     │                             │
                     ▼                             ▼
              ┌──────────────┐          ┌──────────────────┐
              │  add_tool     │          │  add_tool         │
              │  handler:     │          │  handler:         │
              │  tool-deployer│          │  tool-generator   │
              └──────┬───────┘          └────────┬─────────┘
                     │                           │
                     ▼                           ▼
              ┌──────────────┐          ┌──────────────────┐
              │ tool-deployer │          │ tool-generator    │
              │ 1. Fork       │          │ 1. LLM generates  │
              │    component  │          │    HTML/CSS/JS    │
              │ 2. Create page│          │ 2. Create         │
              │ 3. Content    │          │    component      │
              │    sections   │          │ 3. Create page    │
              │ 4. Companion  │          │ 4. Content items  │
              │    guide      │          │ 5. Companion      │
              │ 5. Nav entry  │          │    guide          │
              └──────┬───────┘          │ 6. Nav entry      │
                     │                  └────────┬─────────┘
                     │                           │
                     └──────────┬────────────────┘
                                │
                     page-build-handler
                     (writes hero/intro/CTA + guide)
                                │
                          needs_rerender
                                │
                                ▼
              ┌──────────────────────────────────────┐
              │   rerender-pages → page-rerender      │
              │   compile_page_sections injects:      │
              │   - <head> with global.css            │
              │   - site header                       │
              │   - site footer                       │
              │   Tool template provides:             │
              │   - <style> (with CSS var refs)       │
              │   - tool UI                           │
              │   - <script> (tool logic)             │
              └─────────────────┬────────────────────┘
                                │
                                ▼
              ┌──────────────────────────────────────┐
              │   git commit → GitHub Actions → S3    │
              │   Tool live at /tools/{name}.html     │
              │   Guide at /guides/{name}-guide.html  │
              └──────────────────────────────────────┘
```

---

## Ideas for Future Tools

Organised by likely site-type affinity. The tool-generator agent can create these from descriptions — tool-suggester evaluates which ones would genuinely help a site and creates `add_tool` items with `handler_agent: tool-generator`.

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

