# 003 — Contracts and Standards

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