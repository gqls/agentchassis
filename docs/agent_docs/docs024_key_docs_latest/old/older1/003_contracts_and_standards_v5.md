# 003 — Contracts & Standards

Rules that all agents, actions, templates, and migrations must follow. Breaking these contracts causes subtle bugs that surface days later in production.

---

## Component Naming Contract

`content_components.function` is the canonical identifier. Everything else derives from it or must match it exactly.

### Format: kebab-case

All `function` values use lowercase letters, digits, and hyphens only.

```
Good:  hero, social-proof, call-to-action, case-studies-hero, footer-4-column
Bad:   social_proof, call_to_action, SocialProof, Hero, HERO
```

**Regex:** `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`

**DB constraint:**
```sql
ALTER TABLE content_components
ADD CONSTRAINT chk_function_kebab_case
CHECK (function = '' OR function ~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$');
```

### Uniqueness: one function, one active component

Enforced by partial unique index:

```sql
CREATE UNIQUE INDEX idx_content_components_unique_active_function
ON content_components (function) WHERE is_active = true AND function != '';
```

For variants of the same concept, give each a distinct function: `hero`, `hero-split`, `hero-minimal`, `hero-fullwidth`.

### The data-component attribute

Every section-type component template must include `data-component` on its root element, matching the component's `function` exactly:

```html
<!-- function = "social-proof" -->
<section class="social-proof-section" data-component="social-proof">
    ...
</section>
```

Used by: SavePageSectionsAction, page-rerender, maintenance agents, front-end tooling.

### Naming patterns

| Scope | Pattern | Examples |
|---|---|---|
| General section | `{purpose}` | `hero`, `social-proof`, `call-to-action` |
| Page-specific variant | `{page}-{purpose}` | `about-hero`, `services-hero`, `contact-hero` |
| Site-level header | `header-{variant}` | `header-professional-dark`, `header-minimal-light` |
| Site-level footer | `footer-{variant}` | `footer-4-column`, `footer-standard` |
| Site-level head | `head-{variant}` | `head-seo-standard` |

Page-specific heroes get their own function because they have different CSS (height, gradient). Sharing `function = 'hero'` for all of them breaks uniqueness and makes lookup ambiguous.

### Where function appears

| Location | Column/Field | Must match |
|---|---|---|
| `content_components` | `function` | Source of truth |
| `content_components` | `html_template` → `data-component` | Exactly |
| `page_components` | `slot_name` | Yes — how rerenders find the template |
| `page_components` | `component_id` | Points to the `content_components.id` with this function |
| Workflow config | `component_function` | Yes |
| Site plan | `sections[].function` | Yes |

### The data flow

```
Planner assigns function "social-proof" to a page section
  → RenderComponentAction looks up content_components WHERE function = 'social-proof'
  → CompilePageSectionsAction joins HTML, preserves sections_metadata
  → SavePageSectionsAction stores to page_components: slot_name = "social-proof"
  → page-rerender reads page_components, concatenates rendered_html in order
```

### Lookup safety net

`GetComponentWithFallback` tries: exact match → normalized form (underscore→hyphen) → generic fallback (`generic-text-block`). This prevents silent failures from legacy data but should not be relied upon.

### Go validation functions (component_validation.go)

- `ValidateComponentFunction(function)` — checks kebab-case format
- `ValidateComponentTemplate(function, htmlTemplate)` — checks data-component matches
- `NormalizeComponentFunction(function)` — converts `social_proof` → `social-proof`

### New component checklist

Before adding any new component to `content_components`:

1. **Function name** — kebab-case, matches one of the naming patterns above
2. **`data-component`** — root element has `data-component="{function}"` matching exactly
3. **Template variables** — use `{{.field_name}}` from the render context, NOT hardcoded values
4. **Logo support** (header components only) — template MUST include `{{if .logo_url}}<img>{{end}}`
5. **Nav rendering** (header components) — use `{{range .nav_items}}` or `{{.nav_items_html}}`, NOT hardcoded links
6. **Dark section contract** — if `is_dark_section = true`, template MUST set `--section-*` CSS variables on container
7. **CSS scoping** — component CSS MUST be scoped to the component's class. Never set global element rules (`h1 { }`) in a component
8. **Style collection linkage** (header/footer only) — component ID must be added to at least one `style_collections` row (`header_component_id` or `footer_component_id`)
9. **No search icon** — header templates should NOT include search toggles unless the site has search functionality enabled
10. **Responsive** — include `@media (max-width: 768px)` rules for mobile layout

### Template variable reference

Site-level components (header, footer, head) receive these variables from `contextToInterfaceMap`:

| Variable | Source | Example |
|----------|--------|---------|
| `{{.company_name}}` | `sites.company_name` | "FineTuning" |
| `{{.logo_text}}` | `sites.logo_text` | "FineTuning" |
| `{{.logo_url}}` | `sites.logo_url` via ContentData | "/assets/images/logo.png" |
| `{{.tagline}}` | `sites.tagline` | "AI for the Rest of Us" |
| `{{.email}}` | `sites.email` | "hello@finetuning.uk" |
| `{{.phone}}` | `sites.phone` | "+44 ..." |
| `{{.nav_items}}` | pages table via GetNavItems | slice of NavItem |
| `{{.nav_items_html}}` | pre-rendered `<li>` HTML | `<li><a href="/about.html">About</a></li>` |
| `{{.primary_color}}` | style collection color_palette | "#1a1a2e" |
| `{{.theme_css}}` | css_themes.css_content | `:root { --color-primary: ... }` |
| `{{.year}}` | current year | "2026" |
| `{{.title}}` | page title (head only) | "About Us \| FineTuning" |
| `{{.description}}` | page meta_description (head only) | "..." |

**Note:** `logo_url` is NOT a direct field on `RenderContext` — it flows through `ContentData["logo_url"]` which is merged into the template data by `contextToInterfaceMap`. If the merge doesn't include it, the template sees an empty value. This is a known gap — `logo_url` should be added as a direct field on `RenderContext` for reliability.

---

## Site Component Linkage Contract

Site-level components (header, footer, head) are stored in `site_components` and rendered from `content_components` templates. The linkage between them is the most common source of rendering failures.

### The rule

Every `site_components` row MUST have `component_id` pointing to a valid `content_components.id`. Without this, `renderAndStoreSiteComponent` falls through to a generic lookup, then to a hardcoded fallback that ignores the style collection's templates entirely.

### How it should work

```
style_collections.header_component_id → content_components.id (e.g. header-professional-dark)
                                       ↓
update_site_defaults action copies to:
                                       ↓
site_components.component_id → same content_components.id
                                       ↓
renderAndStoreSiteComponent joins:
  site_components.component_id → content_components.html_template
                                       ↓
RenderTemplate(html_template, renderCtx) → stored as site_components.rendered_html
```

### What breaks this

1. `update_site_defaults` doesn't run (e.g. skipped in workflow, or style collection not selected)
2. `style_collections.header_component_id` is NULL (no header assigned to collection)
3. `site_components.component_id` is NULL after initial build (legacy data)

Any of these causes `renderAndStoreSiteComponent` to fall through to `RenderFallbackHeader` — a hardcoded Go function that produces generic HTML with no logo support, stacked nav, and a search icon.

### Slot name ↔ function mapping

`site_components.slot_name` uses short names. `content_components.function` uses prefixed names. The mapping is:

| slot_name | function |
|-----------|----------|
| `header` | `header-{variant}` (e.g. `header-professional-dark`) |
| `footer` | `footer-{variant}` (e.g. `footer-standard`) |
| `head` | `head-{variant}` (e.g. `head-seo-standard`) |

The generic fallback query `WHERE function = $1` with `slot_name = 'header'` will NOT find a component because the function is `site-header` or `header-professional-dark`, not `header`. This is why the fallback fails silently.

### Validation

After `update_site_defaults` or any site setup:

```sql
-- All slots must have component_id linked
SELECT sc.slot_name, sc.component_id IS NOT NULL as linked
FROM site_components sc
WHERE sc.site_id = $site_id
  AND sc.slot_name IN ('header', 'footer', 'head');
```

If any row has `linked = false`, the site will render with fallback HTML.

### Discovery check: `unlinked_site_components`

```sql
SELECT s.domain, sc.slot_name
FROM site_components sc
JOIN sites s ON s.id = sc.site_id
WHERE sc.component_id IS NULL
  AND sc.slot_name IN ('header', 'footer', 'head');
```

Handler: `site-metadata-fixer` — resolves component_id from style collection.

---

## CSS Colour Inheritance Model

This is the single most important rule in the design system. Getting it wrong causes unreadable text (light text on light backgrounds).

### The rule

Text elements reference `--section-*` CSS custom properties with light-theme fallbacks. Dark-section components override these variables on their container. Everything adapts automatically.

```
/assets/css/styles.css (from css_themes)
  body { color: var(--color-text); }
  h1-h6 { color: var(--section-heading, var(--color-primary)); }
  p, li, blockquote { color: var(--section-text, inherit); }
  strong, em, cite, span — do NOT set color (they inherit from parent)

Light section (no overrides):
  h2 gets var(--section-heading) → not set → fallback var(--color-primary) → #1a1a2e
  p gets var(--section-text) → not set → fallback inherit → body's #333333

Dark section component sets --section-* on container:
  .social-proof-section {
    --section-heading: #ffffff;
    --section-text: rgba(255,255,255,0.9);
  }
  h2 gets var(--section-heading) → #ffffff
  p gets var(--section-text) → rgba(255,255,255,0.9)
```

### Rules for styles.css

- `body` sets `color: var(--color-text)` — the base default
- `h1-h6` use `color: var(--section-heading, var(--color-primary))` — prominent in light, white in dark
- `p`, `li`, `blockquote` use `color: var(--section-text, inherit)` — adapts to section context
- `strong`, `em`, `cite`, `span` — do NOT set `color` at all (inherit from parent element)
- `a` — `color: var(--color-accent)` is the one exception (links are always explicit)
- `blockquote` — do NOT set `background-color` (components handle this contextually)

### What breaks this

If the base CSS sets `color: var(--color-text)` directly on `p` or `h1`, the `--section-*` override is bypassed. The element gets `#333333` regardless of what the dark section container sets. This is the bug that caused light-on-light text in testimonial sections.

Similarly, if the base CSS sets `color: var(--color-primary)` on `h1` instead of `var(--section-heading, var(--color-primary))`, dark sections can't override headings to white.

---

## Section Context Variable Contract (Dark Sections)

Any component with a dark background MUST set these CSS custom properties on its outermost container:

```css
.my-dark-section {
    background: var(--color-primary, #1a1a2e);
    color: var(--color-white, #fff);

    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.7);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);
}
```

The global `styles.css` base element rules consume these variables (see CSS Colour Inheritance Model above). In light sections where `--section-*` variables are not set, the fallbacks provide correct light-theme values. In dark sections, the overrides apply automatically to all children.

### Dark components

| function | is_dark_section |
|---|---|
| `hero`, `hero-about`, `hero-services`, `hero-contact`, `hero-case-studies`, `hero-use-cases` | true |
| `social-proof`, `testimonials`, `call-to-action` | true |

### Enforcement layers

1. **Database:** `content_components.is_dark_section` boolean column
2. **Go validation:** `ValidateDarkSectionContract()` warns in RenderComponentAction and SavePageSectionsAction (warnings, not errors)
3. **LLM prompt:** webdesign-agent and page-content-writer prompts include the pattern
4. **Periodic audit:** SQL queries to find dark components missing `--section-*` variables

### Audit queries

```sql
-- Dark components missing the contract
SELECT name, function FROM content_components
WHERE is_dark_section = true
  AND html_template NOT LIKE '%--section-text%';

-- Likely-dark components not flagged
SELECT name, function FROM content_components
WHERE is_dark_section = false AND component_level = 'section'
  AND (html_template LIKE '%background:%#1a1a2e%'
       OR html_template LIKE '%background: var(--color-primary%');
```

---

## Query Database Parameterisation Contract

### The rule

All new `query_database` usage in workflow configs MUST use parameterised queries with `$1`, `$2` placeholders and a `"params"` array. Never embed values into SQL strings via Go template interpolation (`{{.field}}`).

### Correct (parameterised):

```json
{
    "action": "query_database",
    "config": {
        "query": "SELECT domain, company_name FROM sites WHERE id = $1",
        "params": ["site_record.site_id"],
        "output_format": "object"
    }
}
```

The `params` array contains dot-paths resolved from collected_data via `ExtractNestedField`. Values are passed as query arguments to `QueryContext`, preventing SQL injection.

### Incorrect (template interpolation — legacy pattern):

```json
{
    "action": "query_database",
    "config": {
        "query": "SELECT domain FROM sites WHERE id = '{{.input_data.site_id}}'",
        "output_format": "object"
    }
}
```

This embeds the value directly into the SQL string. It works but is a SQL injection risk — if `input_data.site_id` contained malicious content, it would be executed as SQL.

### Why this matters

The system processes external input (domain names, briefing answers, scraped content). While current inputs come from trusted sources (CLI, planner), future inputs may come from APIs, webhooks, or user-facing forms. Parameterised queries protect against this from the start.

### Legacy migration (TODO)

The following existing agents use template interpolation in `query_database` and should be migrated to parameterised queries:

| Agent | Steps using `{{.field}}` in SQL |
|-------|-------------------------------|
| `tool-suggester` | `load_pages`, `load_existing_tools` |
| `tool-improver` | `load_tool` |

Migration: replace `'{{.input_data.site_id}}'` with `$1` and add `"params": ["input_data.site_id"]` to the config. Test that the query returns the same results.

Agents already using parameterised queries (no migration needed):
- `visual-design-auditor` — all queries
- `content-quality-auditor` — all queries
- `site-review-agent` — all queries
- `build-pipeline-trigger` — `find_dispatchable_site` (no site-specific params)
- `build-site-planner` — `load_components`, `load_styles` (no params needed)

---

## Schema Enforcement (Flexible vs Strict Mode)

### Two modes

| Mode | Use Case | Behavior |
|---|---|---|
| **Flexible** | Initial build, creative exploration | LLM can generate any structure; renderer does best-effort substitution |
| **Strict** | Editing, maintenance, approved designs | Must match component's `input_schema` exactly; fail if mismatch |

### How it works

During initial build, `schema_mode = 'flexible'`. Missing fields become empty strings, unsubstituted placeholders are logged as warnings.

At approval time, the structure is locked:
- `page_components.schema_snapshot` = the component's `input_schema` at approval
- `page_components.content_snapshot` = the actual content values used
- `sites.schema_mode` switches to `'strict'`

In strict mode: validation errors on mismatch, all required fields must be present, component template upgrades don't break existing pages because the snapshot preserves the contract.

### Source of truth principle

Every page section has two representations:
- **`content_data`** (structured JSON in `page_components`) — produced by content writer
- **`rendered_html`** (final HTML in `page_components`) — produced by rendering template with content_data + site context

**content_data is always the source of truth.** Every edit updates content_data first, then re-renders. If we only patched rendered_html, the edit would be lost on the next re-render (nav update, theme change, page rebuild).

This is why HTML patching was rejected as an edit mechanism — edits vanish on re-render, and content_data/rendered_html drift apart.

---

## Component Input Schema v2 Contract

`content_components.input_schema` declares what data each component needs and where it comes from. The `plan_sections` action reads this schema to determine which sections can be generated vs which need human input.

### Format

```json
{
    "fields": {
        "headline": {
            "type": "text",
            "source": "llm",
            "required": true
        },
        "team_members": {
            "type": "array",
            "source": "site_specs.identity.team",
            "required": true,
            "on_missing": "needs_human_review",
            "min_items": 1,
            "missing_reason": "Team member names, titles, and bios are needed"
        },
        "hero_image": {
            "type": "image",
            "source": "site_assets.hero",
            "required": false,
            "on_missing": "use_fallback",
            "fallback": "/assets/images/hero.jpg"
        }
    }
}
```

### Field properties

| Property | Required | Description |
|----------|----------|-------------|
| `type` | yes | `text`, `array`, `image`, `url`, `boolean` |
| `source` | yes | Where the data comes from (see source prefixes) |
| `required` | yes | Whether the section can render without this field |
| `on_missing` | no | What to do when data unavailable (default: `skip_field`) |
| `fallback` | no | Value to use when `on_missing` is `use_fallback` |
| `min_items` | no | Minimum array length for array types |
| `missing_reason` | no | Human-readable explanation for HITL work items |
| `items` | no | Schema for array element structure |
| `llm_guidance` | no | Hint for the content writer when generating this field |

### Source prefixes

| Prefix | Resolution | When resolved |
|--------|-----------|--------------|
| `llm` | Content writer generates | At content generation time |
| `site_specs.{aspect}.{path}` | Lookup in `site_specs` table | At plan_sections time |
| `site_assets.{type}` | Check site asset exists | At plan_sections time |
| `pages.{name}` | Check page exists, resolve URL | At plan_sections time |
| `config.{path}` | Read from site config/specs | At plan_sections time |
| `renderer` | Assembled by page-rerender agent | At render time |
| `static` | Use fallback value always | At plan_sections time |
| `query.{name}` | DB query at render time (blog posts, categories) | At render time |

### on_missing values

| Value | Effect | Use case |
|-------|--------|----------|
| `use_fallback` | Use the fallback value, continue | Hero image, CTA URL, default text |
| `skip_field` | Omit field from template data | Optional phone, social links |
| `skip_section` | Don't generate this section at all | Testimonials with no data |
| `needs_human_review` | Create HITL work item, omit section | Team members, pricing, case studies |
| `block` | Fail the entire page build | Company name missing |

### Component categories

| Category | Source pattern | Examples |
|----------|---------------|---------|
| Heroes, CTAs, features | `source: "llm"` for all fields | LLM generates from site identity |
| Team, testimonials, pricing | `source: "site_specs.*"` with `needs_human_review` | Needs real data |
| Contact info | `source: "site_specs.identity.*"` | Email, phone, address |
| Headers, footers | `source: "renderer"` | Assembled at render time |
| Blog/content | `source: "query.*"` | Dynamic DB queries |
| Tools | Empty `fields: {}` | Self-contained |

### Rule: content writers never fabricate data they don't have

If a field's source is not `llm` and the data is not available, the content writer never sees that section. The `plan_sections` action prevents it from reaching the LLM. This eliminates placeholder text at the source rather than catching it after the fact.

---

## Content Validation Contract

`validate_page_content` runs after the content writer and before `save_sections`. It is the safety net — `plan_sections` prevents most problems upstream, but the LLM can still produce bad content in sections it was given.

### Severity model

| Severity | Blocks deployment | Examples |
|----------|------------------|---------|
| `blocker` | Yes | Placeholder text, unrendered templates, cross-site contamination, placeholder emails |
| `error` | Yes | Broken internal links, hallucinated emails |
| `warning` | No (logged) | Short content |

### What it returns

On pass: `{valid: true, clean_html: "...", issues: [...]}` — save_sections reads `validation_result.clean_html`.

On fail: returns error → workflow routes to `mark_needs_review` → `fail_work_item` with `status_override: needs_human_review`.

### Cross-site contamination check

The validator maintains a list of known site domains and company names. If any domain or company name appears in content for a different site, it's a blocker. This catches the case where LLM context bleed causes one site's identity to appear in another site's pages.

---

## Orchestrator Boundaries

Orchestrators and child agents have distinct responsibilities.

**Orchestrator's job:**
- Know what needs doing and in what order
- Load batches of work from the database
- Dispatch child agents with raw domain identifiers
- Track overall progress

**Agent's job:**
- Know how to do the work
- Own its search strategy, data transformation, and domain logic
- Be independently callable and testable

**Warning sign:** If changing an agent's internal approach requires updating its parent orchestrator, the boundary is wrong.

---

## site_context Schema

Standardized format for design data from any source (DB, scrape, manual):

```json
{
  "site_id": "uuid",
  "domain": "example.com",
  "company_name": "Example Corp",
  "tagline": "Making examples simple",
  "industry": "technology",
  "color_palette": {
    "primary": "#1a1a2e", "secondary": "#16213e",
    "accent": "#0f3460", "background": "#ffffff", "text": "#333333"
  },
  "typography": {
    "font_family": "-apple-system, sans-serif",
    "heading_font": "inherit", "base_size": "16px", "line_height": "1.6"
  },
  "pages": [
    {"title": "Home", "slug": "/", "component_functions": ["hero", "services-grid"]}
  ],
  "all_component_functions": ["hero", "services-grid", "cta"],
  "source": "database|scrape|manual"
}
```

---

## Cross-Domain Coordination Rules

Domains are independent. No agent calls another agent for coordination. They communicate through the `site_work_items` table.

When a fix agent changes something that affects another domain:

```
section-rewriter removes a paragraph containing a link
  → new item: domain='links', item_type='link_removed', source='side_effect'

page-content-writer creates a new page
  → new item: domain='navigation', item_type='nav_update_needed', source='side_effect'
  → new item: domain='seo', item_type='needs_sitemap_update', source='side_effect'
```

Side-effect items get triaged and processed in the next cycle.

### Cross-site operations

Same change across multiple sites: work items created per site via SQL INSERT. Each site's orchestrator processes independently, respecting each site's design and tone.

Cross-site pattern detection: the catch-all agent (daily cron) detects same broken URL across 50 sites → fix once, apply to all.

---

## Agent Definition SQL Conventions

### Required fields

Every agent definition includes: `type`, `display_name`, `description`, `agent_category` (must be one of: strategist, executor, analyst, integrator, coordinator, specialist), `status`, `default_config` (with workflow), `input_contract`, `output_contract`.

### docker_image / docker_tag

All existing agent definitions include these. The table schema requires them. `spawn_actions.go` overrides at runtime. Include them in SQL even though the checklist historically said not to.

### Handler agent contract

All handler agents follow the same contract:

```
Input:  work item (from site_work_items row — spec field is primary input)
        site context (site record, style collection, nav data, brief)
Output: result JSONB (includes commit_sha if page was deployed)
        status: 'complete' or 'failed'
```

---

## content_direction (Page-Level Edit Instructions)

JSONB column on `pages` table. Flows to content-writer's prompt when present.

```json
{
  "format": "problem-agitate-solution",
  "instruction": "Each use case: present a real problem, agitate the pain, suggest solutions.",
  "use_cases": ["automated website generation", "content refresh"]
}
```

Used when: page-level rewrite needed (set `build_status = 'needs_rebuild'`), then trigger page-rebuild.

---

## Legal Rules Schema

Per-site, stored in `sites.content_data.legal_rules`:

```json
{
  "legal_rules": {
    "industry": "finance",
    "required_disclaimers": [
      { "trigger": "any_financial_content", "text": "...", "placement": "section_footer" }
    ],
    "forbidden_phrases": ["guaranteed returns", "risk-free"],
    "required_pages": [
      { "name": "privacy", "template": "privacy-gdpr-uk", "nav_group": "legal" }
    ]
  }
}
```

Templates seed common rules per industry. Live rules belong to the site and may be customized.