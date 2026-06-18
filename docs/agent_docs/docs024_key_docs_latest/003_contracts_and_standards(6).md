# 003 — Contracts & Standards

> **Consolidation note.** Canonical `003 — Contracts & Standards`, superseding
> v8–v10. The progression was cleanly additive: v9 added the String-Value Naming
> Convention, v10 added the Adapter Response Envelope Contract, and v11 expanded
> that contract with the "Headers MUST be a typed struct (the bool trap)"
> subsection — the typed-struct requirement *strengthens* the prior anti-pattern
> guidance (the v10 checklist's wording was reworded and extended here, not
> dropped). No earlier content was lost.


Rules that all agents, actions, templates, and migrations must follow. Breaking these contracts causes subtle bugs that surface days later in production.

---

## Workflow Result Contract (complete_workflow output)

The `complete` step declares what a workflow returns to its parent. The coordinator
(`resolveResultSpec` → `extractWorkflowResult`, platform/orchestration) reads **one**
of three keys; pick the one that matches how the parent reads the result.

| Mode | Preferred key | Deprecated alias | Result shape |
|------|---------------|------------------|--------------|
| Flatten | `result_from: "field"` | `output_field` | the named field's **contents become the body** |
| Fields  | `multiple_output_fields: ["a","b"]` | `output_fields` | each field **nested under its own key** |
| Mapping | `result_mapping: {"t":"src.path"}` | `output` | body built from explicit target←source pairs |

A parent reads a child's result at `<call_output_field>.response.<key>`. So **flatten**
puts the field's keys directly at `.response.<key>` (e.g. page-content-writer's
`result_from: "page_content"` makes `page_content.response.page_html` and
`page_content.response.sections_metadata` resolve for page-build-handler); **fields**
puts them one level deeper at `.response.<field>.<key>`. Choosing the wrong mode is a
silent runtime mismatch (null reads), not a compile error — match the mode to the
consumer.

Rules:

- **Use the preferred names.** `output_field` and `output_fields` differ by one letter
  but mean different shapes — a foot-gun. The preferred names (`result_from`,
  `multiple_output_fields`, `result_mapping`) are non-confusable. Deprecated aliases
  still resolve but log a deprecation `Warn`; migrate an agent's key when you next
  touch it.
- **Flatten is for single-object results**, nest is for multi-output handlers. If a
  consumer reads `x.response.foo` (flat), the producer must flatten; if it reads
  `x.response.field.foo`, the producer must nest.
- **No key set → fallback dump** of `collected_data` (large, unbounded — discouraged;
  it can breach the ~900k result cap and collapse to a stub). Always declare a contract
  on a workflow whose result a parent consumes.
- **Completion metadata** (`orchestration_id`, `completed_steps`, `completed_at`) is
  appended automatically, only-if-absent, so it can't clobber a flattened/mapped field.

See `016_debugging_guide` §9 for the failure mode this contract fixed (a singular
`output_field` silently ignored, the result dumped and stubbed).

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

---

## String-Value Naming Convention

Beyond `content_components.function`, the codebase uses string values in many other columns and identifiers: `pages.page_type`, `site_work_items.item_type`, `agent_definitions.type`, action registry keys, status enums, and so on. These don't all follow the same convention — and they shouldn't, because they serve different purposes. This section documents the rule.

### The two-and-a-half conventions

| Category | Convention | Examples |
|---|---|---|
| **Identifier-shaped values** — used as keys in Go `map[string]X` lookups, `switch case` constants, action registry names, work-item dispatch routing | `snake_case` | `site_work_items.item_type` (`needs_blog_posts`, `orphan_blog_posts`), action names (`create_blog_posts`, `apply_adoption_plan`), `research_results.result_type` (`adoption_crawl`, `adoption_page`) |
| **Data-shaped values** — describe what a thing *is*, never used as identifiers in code; pure metadata that flows through JSON and SQL columns | `kebab-case` | `content_components.function` (`hero`, `social-proof`), `pages.page_type` (`blog-post`, `section-index`, `landing`), `agent_definitions.type` (`site-adoption-agent`, `page-build-handler`), CSS class fragments, URL path segments |
| **Single-word values** — no separator needed | lowercase word | `pages.build_status` (`planned`, `deployed`), `site_work_items.status` (`detected`, `triaged`, `claimed`), `agent_definitions.status` (`active`, `experimental`) |

### The rule for choosing

When introducing a new string-typed column or enum-like value, ask: **is this value ever used as an identifier in code?** Specifically:

- Does any Go file have `case "the-value":` or `case "the_value":` switching on it?
- Is it ever a map key in a registry-shaped data structure?
- Does it appear in a Kubernetes label, Kafka topic segment, or work-item dispatch route?

If **yes** to any → `snake_case`. The value will appear in Go code as a string literal and snake_case avoids visual confusion with subtraction in surrounding hyphen-using contexts.

If **no** → `kebab-case`. The value is data, will end up in CSS / URLs / HTML eventually, and kebab is the prevailing convention there.

### Why two conventions

The honest answer is that the codebase has two different consumers of these values:

1. **Code consumers** (switch statements, registries, dispatch routers) prefer snake because the values look like identifiers in surrounding Go code. `case "blog_post":` reads as "the constant blog_post"; `case "blog-post":` reads visually as "blog minus post" (it is, syntactically, a perfectly legal string literal — but the brain still does the subtraction lookup).
2. **Data consumers** (CSS templates, URLs, HTML attributes, the analyze_site LLM prompts) prefer kebab because kebab is the convention there. URLs are lowercase-kebab. CSS classes are lowercase-kebab. HTML attributes (`data-component`, `aria-label`) are lowercase-kebab.

A single convention would force one consumer to translate. The split prevents the translation by aligning each column with its dominant consumer.

### Enforcement

Both conventions are enforced via `CHECK` constraints where the column is enum-like:

```sql
-- kebab-case constraints
ALTER TABLE content_components
    ADD CONSTRAINT chk_function_kebab_case
        CHECK (function = '' OR function ~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$');

ALTER TABLE pages
    ADD CONSTRAINT chk_page_type_kebab_case
        CHECK (page_type IS NULL OR page_type = ''
            OR page_type ~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$');
```

`snake_case` columns have not yet had explicit constraints added — that's a follow-up. The regex would be `^[a-z][a-z0-9]*(_[a-z0-9]+)*$`.

### page_type vocabulary

The canonical `pages.page_type` values:

| Value | Meaning | Where created |
|---|---|---|
| `landing` | The homepage (with `name = 'index'`), or a conversion-focused flat page | `CanonicalisePage` for the `index` role and the `home`/`index` slug collapse; the validator's rule 1 |
| `content` | Generic content page | Default for the `content` role; fallback for unknown roles |
| `tool` | Interactive utility page (a tool the user uses) | The `tool` role branch in `CanonicalisePage` |
| `guide` | Tutorial / explanatory long-form page | The `guide` role branch |
| `game` | Interactive game page | The `game` role branch |
| `blog-post` | Individual blog/article page | The `blog-post` role branch |
| `blog-index` | Index page listing blog posts | The `blog-index` branch (a section-index flavour) |
| `section-index` | Generic index page for a section directory | The `section-index` branch |
| `entity-page` | Individual entity within an entity directory | The `entity-page` role branch |
| `entity-directory` | Index page listing entities | The `entity-directory` branch (a section-index flavour) |

The `CanonicalisePage` helper accepts the legacy snake-form inputs for backward compatibility with older planner prompts (`blog_post`, `blog_index`, `section_index`, `entity_directory`, `entity_page`) and normalises them to kebab on output. New prompts should emit kebab directly.

### Why the homepage's page_type is "landing", not "index"

A common pattern was to use `page_type = 'index'` for the homepage. Doing so conflates the page's **name** (which is "index" by storage convention) with its **type** (which describes what kind of page it is). The kind-of-page is "landing" — the same kind as any other conversion-focused entry page. Using "landing" for the type lets:

- The homepage and other landing pages share rendering / CTA / nav-suppression logic by checking `page_type = 'landing'`
- Future analytics distinguish landing pages from content pages without inferring from the name
- The same row act as the homepage *and* as a particular type of page, without one fact shadowing the other

The page's `name` column still holds `"index"`; that's the storage convention for the homepage, independent of type.

---

## JS Content Separation Contract

Interactive components (games, feeds, explorers) often need substantial JavaScript. Storing large `<script>` blocks inside `html_template` causes two problems: LLM token-limit truncation during component generation, and slow page loads from inline JS.

### The split

`content_components` has two columns for a single component:

| Column | Contents | Served as |
|---|---|---|
| `html_template` | `<style>...</style><section>...</section><script src="/tools/assets/{function}.js"></script>` | inline in the page HTML |
| `js_content` | The raw JS body (no `<script>` tags) | `/tools/assets/{function}.js` asset file |

Components without JS have `js_content = NULL` — no asset file is generated, no script tag is added.

### The flow

```
LLM generates component (may include inline <script> blocks)
  ↓
store_generated_component_action.go
  → separateInlineJS() extracts <script>...</script> bodies
  → replaces them with <script src="/tools/assets/{function}.js"></script>
  → INSERT both columns
  ↓
page-content-writer renders html_template with content data
  → rendered_html contains the <script src> tag (not inline JS)
  ↓
RerenderSinglePageAction assembles the page
  → collectJSAssets() queries js_content for all page components
  → returns files map: {"{page}.html": html, "tools/assets/{function}.js": js, ...}
  ↓
page-rerender workflow's git_commit step
  → uses files_field: rendered_page.files (multi-file commit)
  → git-adapter writes HTML + all JS assets in one commit
  ↓
deployer-agent triggers CDN deploy
  → CDN serves both the HTML page and the JS asset files
```

### Asset path convention

Component JS files live at `/tools/assets/{function}.js` — the `function` value from `content_components`, not the `section_type` or any other field. This keeps the path deterministic and collision-safe (functions are globally unique among active components).

### Relationship to js_snippets

`js_snippets` is a separate table for **shared design effects** (scroll reveals, parallax, lazy loading) that apply across many components. It is loaded via the head component's snippet-loading mechanism. It is NOT used for component-specific behaviour.

| Purpose | Storage | Scoping |
|---|---|---|
| Component-specific JS (game logic, interactive widgets) | `content_components.js_content` | 1:1 with the component |
| Shared design effects (animations, hydration helpers) | `js_snippets` table | Applied across many components via `applies_to` |

### What must not happen

- Do NOT put inline `<script>` blocks in `html_template`. `separateInlineJS()` removes them, but LLM prompts should generate the contract-compliant form from the start.
- Do NOT commit component JS to `js_snippets` unless it's genuinely shared across multiple components.
- Do NOT hardcode the `<script src>` path — the component-creator pipeline sets it based on `function`.

### For component-creator prompts

Tell the LLM: "If the component needs JavaScript, include one or more `<script>` blocks with only the JS body (no attributes on the `<script>` tag). The pipeline will separate these into a distinct asset file automatically."

---


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

## Component Creation & Regeneration Contract

`StoreGeneratedComponentAction` persists LLM-generated component templates to `content_components`. It has two primary paths that downstream workflows may need to distinguish:

### Return status values

| `status` | Meaning | Primary key behaviour | Side effects |
|---|---|---|---|
| `created` | No existing active component with this `function`. Row INSERTed. | New UUID generated | Version 1 snapshot written to `component_versions` |
| `regenerated` | Existing component found. Old state snapshotted, new template UPDATE in place. | UUID preserved — foreign keys in `page_components`, `site_components`, etc. remain valid | Snapshot written to `component_versions` with `MAX(version_number)+1`; one `needs_rerender` work item created per affected site |

The `already_exists` status from earlier versions is REMOVED. Every call now either creates or regenerates — the LLM's output is never silently discarded just because a row with the same function already exists.

### Rejection path

Layer 1 pre-store validation runs BEFORE the create/regenerate branch. If the new template fails validation (zero template placeholders despite a populated schema, malformed structure, etc.), the action returns an error and does NOT touch the existing row. An existing broken component stays in place — rejection never makes things worse.

### Return payload (regenerated)

```json
{
  "component_id": "<existing-uuid>",
  "function": "provocation-feed",
  "status": "regenerated",
  "previous_version": 2,
  "new_version": 3,
  "pages_marked_rebuild": 4,
  "affected_sites": 2,
  "rerender_items_created": 2,
  "quality_score": 82,
  "quality_issues": []
}
```

### Return payload (created)

```json
{
  "component_id": "<new-uuid>",
  "function": "foo-section",
  "status": "created",
  "new_version": 1,
  "quality_score": 85,
  "quality_issues": []
}
```

### Downstream contract

Workflow steps consuming `StoreGeneratedComponentAction` output:

- MAY branch on `status` to distinguish the cases
- MUST NOT assume `component_id` is newly minted — it might point to a long-lived component whose template was just replaced
- SHOULD NOT create their own `needs_rerender` items when `status = "regenerated"` — the action has already done so (one per affected site), deduped via `item_key = component_regen_rerender:<uuid>`

### Version history preservation

Every create and regenerate writes a row to `component_versions` with `change_source` recording the originating work item's `source` field. See `014_site_snapshots_and_revert.md` for population patterns across writers, and `026_component_regeneration_flow.md` for the full flow.

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

## CSS Theme Template Contract

`css_themes.css_template` is a Go template rendered by `render_css_from_spec_action` to produce each site's deployed `/assets/css/styles.css`. This section defines what a theme template owns, what the renderer owns, and how theme lineage flows through `css_themes` and `style_collections`.

### Responsibility split

**The renderer owns:**
- Injecting palette values into `{{.Primary}}`, `{{.Surface}}`, etc. placeholders
- Appending a block of `--section-*` defaults based on palette luminance, using `pickReadableOnBackground` to choose colours that preserve palette character (see `color_util.go`)
- Appending component-specific CSS snippets from `content_components.css_snippet`

**The theme template owns:**
- Layout structure: container widths, grid systems, section padding patterns
- Typography scale: heading sizes, line heights, font weights
- Component-level styling: buttons, forms, cards, navigation
- Base element rules using the fallback pattern from the CSS Colour Inheritance Model
- Per-section backgrounds (painting `.differentiators-section` with `var(--color-surface)` for visual rhythm, for example)

**Themes MUST NOT:**
- Declare `--section-text`, `--section-heading`, `--section-text-muted`, `--section-surface`, or `--section-border` as defaults anywhere in the template. The renderer does this based on palette luminance. Duplicating it causes double-emission and freezes colour choices that should adapt to each site's palette.
- Hardcode hex colours on text elements. Always use the `var(--section-*, var(--color-*))` fallback pattern from the CSS Colour Inheritance Model.

### Template variables

| Variable | Source | Example |
|----------|--------|---------|
| `{{.Primary}}` | `design_spec.color_scheme.primary` | `#1a365d` |
| `{{.Secondary}}` | `design_spec.color_scheme.secondary` | `#2c5282` |
| `{{.Accent}}` | `design_spec.color_scheme.accent` | `#3182ce` |
| `{{.Background}}` | `design_spec.color_scheme.background` | `#ffffff` |
| `{{.Surface}}` | `design_spec.color_scheme.surface` | `#f7fafc` |
| `{{.Text}}` | `design_spec.color_scheme.text` | `#2d3748` |
| `{{.TextMuted}}` | `design_spec.color_scheme.text_muted` | `#718096` |
| `{{.Border}}` | `design_spec.color_scheme.border` | `#e2e8f0` |
| `{{.FontFamily}}` | `design_spec.typography.font_family` | font stack |
| `{{.HeadingFont}}` | `design_spec.typography.heading_font` | font stack |
| `{{.BaseSize}}` | `design_spec.typography.base_size` | `16px` |
| `{{.LineHeight}}` | `design_spec.typography.line_height` | `1.6` |
| `{{.SectionPadding}}` | `design_spec.spacing.section_padding` | `4rem 0` |
| `{{.ContainerMaxWidth}}` | `design_spec.spacing.container_max_width` | `1200px` |
| `{{.SectionStyles}}` | Derived from site components + `is_dark_section` | Array for `{{range}}` |
| `{{.Components}}` | `site_context.all_component_functions` | Array of function names |
| `{{.BackgroundIsDark}}` | Computed luminance flag — available but renderer owns section defaults | boolean |
| `{{.SurfaceIsDark}}` | Computed luminance flag — available but renderer owns section defaults | boolean |

### Theme storage columns

| Column | Purpose |
|--------|---------|
| `css_template` | Go template with `{{.Primary}}` etc. placeholders. Rendered per-site by `render_css_from_spec_action`. |
| `css_content` | Frozen snapshot of rendered CSS at fork time. Reference-only — not used by the renderer. Preserves what the adopted site looked like at the moment of fork. |
| `color_palette` | JSON palette stored on the theme row. Surfaced to selectors and the review UI. |
| `typography` | JSON typography stored on the theme row. Same role. |

Every adopted theme stores both `css_template` and `css_content`. Seed themes may populate only one (legacy themes have `css_content` only, and the renderer falls back to `standard-brochure`'s `css_template` — to be addressed by converting legacy themes to template form).

### Lineage columns

Present on both `css_themes` and `style_collections`:

| Column | Meaning |
|--------|---------|
| `origin` | `seed`, `handcrafted`, `adopted`, or `fork_of_adopted` |
| `forked_from_theme_id` / `forked_from_collection_id` | Parent theme/collection this was derived from. Null for seed themes. |
| `source_site_id` | FK to `sites.id` — the site this theme was forked from. `ON DELETE SET NULL`. |
| `source_domain` | Domain string preserved even if the source site is deleted. |
| `forked_at` | When the fork happened. |
| `needs_review` | True for newly-forked themes until HITL approves them. Selectors MUST filter out `needs_review = true` rows. |

### Review gate

Forked themes enter the library with `needs_review = true` and `is_active = true`. A `needs_theme_review` work item is created in the same transaction. They are not selectable by the style collection selector until a human approves (setting `needs_review = false`) or rejects (setting `is_active = false`).

Rejection is safe and reversible: the source site retains its own deployed CSS; rejection only affects whether the theme is offered to *future* sites.

### Forking rules

When `webdesign-agent` runs with `should_fork_theme: true` in its input:

1. After `deploy_css` and `update_site` succeed, the workflow calls `fork_theme_from_site`
2. The action inserts `css_themes` + `style_collections` + `site_work_items` in one transaction
3. If the site has an existing `style_collection_id`, its theme becomes `forked_from_theme_id` and `origin` is `fork_of_adopted`
4. If not, `forked_from_theme_id` is null and `origin` is `adopted`
5. Header/footer components from the source site's collection are reused as-is — forking components is a separate decision handled through `needs_new_component` work items when layout genuinely differs
6. The theme is named `adopted-{domain-slug}`, with a timestamp suffix appended on name collision

The fork never fails the parent workflow — a failed fork logs a warning and returns `{forked: false, reason: "..."}`. The adopted site already has its CSS deployed; the library contribution is best-effort.

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
| `query.{name}` | DB query at plan time, projected to the field's shape | At `plan_sections` time |

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

## Adapter Response Envelope Contract

Moved to `035_adapter_guide.md` §1 ("The message envelope contract"), now the
single source for it. The contract is normative for any component that replies to
a chassis request — adapters, spawned agents, and inline actions.

In short: a reply must be recognised as an *awaited response* (`HandleResponse →
ProcessResponse → ClaimAwaitedRequest`), or the chassis falls through to the
process-as-work path and the `awaited_requests` row sits `waiting` until timeout
with no error logged. The load-bearing field is `in_response_to_request_id` (the
incoming `request_id`); the body `headers` object must be a typed struct with
real bool `is_complete`/`is_error`; send via `ProduceWithValidation`. See 035 §1
for the full contract, the request-parse rules, the header tiers, and the audit
checklist.

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

#### Input data paths

The dispatch loop (`build-dispatch-loop`) passes work item data to handlers with a **consistent nested structure**. Handlers MUST use these paths in their workflow step configs:

| What you want | Path in workflow config | Where it comes from |
|---|---|---|
| The work item spec (primary input) | `input_data.spec` | `site_work_items.spec` (the JSONB column) |
| A field from the spec | `input_data.spec.{field_name}` | e.g. `input_data.spec.page_name` |
| The site_id | `input_data.site_id` or `site_record.site_id` after `ensure_site_record` | Set by the dispatch loop |
| The domain | `input_data.domain` or `site_record.domain` after `ensure_site_record` | Set by the dispatch loop |
| The work_item_id | `input_data.work_item_id` | Set by the dispatch loop |
| The item_type | `input_data.item_type` | Set by the dispatch loop |

**Contract rule:** All fields originating from the work item's `spec` JSONB live under `input_data.spec.*`. Do NOT rely on top-level flattened paths like `input_data.page_name` — these only exist when the dispatch loop's `input_mapping` has an explicit `page_name?` entry for that field, and the `?` makes them silently nil when absent. Using the nested path is the one contract-compliant way.

**Good (contract-compliant):**
```json
"load_page_record": {
"action": "load_page_record",
"config": {
"site_id":   "site_record.site_id",
"page_name": "input_data.spec.page_name",
"page_id":   "input_data.spec.page_id"
}
}
```

**Bad (relies on optional flattening):**
```json
"load_page_record": {
"config": {
"page_name": "input_data.page_name"
}
}
```

Within a single workflow, ALL steps that read the same spec field MUST use the same path. Mixing `input_data.page_name` and `input_data.spec.page_name` in the same workflow is a bug — it will silently break when the dispatch loop's `input_mapping` changes.

#### Action-level defense

Go actions that read common fields (`page_name`, `page_id`, `site_id`) should implement a fallback chain: first the explicit config path, then `input_data.spec.{field}`, then other well-known locations. This protects the system against accidental config drift. See `load_page_record_action.go` for the pattern.

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

