# Component Naming Contract

Authoritative reference for component identification across the system. All agents, actions, templates, and migrations must follow these rules.

## The Problem This Solves

Components flow through multiple stages: template selection → rendering → page assembly → storage → rerender. At each stage, the component needs to be identifiable. When naming is inconsistent (`social_proof` in the DB but `social-proof` in the HTML attribute), the chain breaks — components can't be matched back to their templates, CSS gets lost, and rerenders produce unstyled pages.

## The Single Rule

**`content_components.function` is the canonical identifier.** Everything else derives from it or must match it exactly.

## Format: Kebab-Case

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

## Uniqueness: One Function, One Active Component

No two active components may share the same `function` value. This is enforced by a partial unique index:

```sql
CREATE UNIQUE INDEX idx_content_components_unique_active_function
ON content_components (function)
WHERE is_active = true AND function != '';
```

If you need multiple variants of the same concept (e.g. different hero layouts), give each a distinct function: `hero`, `hero-split`, `hero-minimal`, `hero-fullwidth`.

## The `data-component` Attribute

Every section-type component template must include a `data-component` attribute on its root element. Its value **must exactly equal** the component's `function`.

```html
<!-- function = "social-proof" -->
<section class="social-proof-section" data-component="social-proof">
    ...
</section>
<style>
.social-proof-section { ... }
</style>
```

This attribute is used by:
- `SavePageSectionsAction` (fallback HTML parsing path) to identify stored sections
- `page-rerender` to match sections back to templates
- Maintenance agents to diagnose and repair component links
- Front-end tooling for section-level editing (future)

**Validation in Go:**
```go
if err := ValidateComponentTemplate(function, htmlTemplate); err != nil {
    // data-component doesn't match function — reject
}
```

## Naming Patterns

| Component scope | Pattern | Examples |
|---|---|---|
| General section | `{purpose}` | `hero`, `social-proof`, `call-to-action`, `features` |
| Page-specific variant | `{page}-{purpose}` | `about-hero`, `services-hero`, `contact-hero`, `case-studies-hero` |
| Site-level (header) | `header-{variant}` | `header-professional-dark`, `header-minimal-light`, `header-bold-gradient` |
| Site-level (footer) | `footer-{variant}` | `footer-4-column`, `footer-standard`, `footer-simple` |
| Site-level (head) | `head-{variant}` | `head-seo-standard` |

**Why page-specific heroes get their own function:** A `services-hero` has different CSS (shorter height, different gradient) from the homepage `hero`. They're distinct templates. Sharing `function = 'hero'` for all of them breaks the uniqueness contract and makes lookup ambiguous.

## Where `function` Appears in the System

| Location | Column/Field | Must match `function` |
|---|---|---|
| `content_components` | `function` | — (this IS the source) |
| `content_components` | `html_template` → `data-component="..."` | Yes, exactly |
| `page_components` | `slot_name` | Yes — this is how rerenders find the template |
| `page_components` | `component_id` | Points to the `content_components.id` with this function |
| Workflow config | `component_function` in step config | Yes — used by `RenderComponentAction` |
| Site plan | `sections[].function` or `sections[].component` | Yes — how planner assigns components to pages |

## The Data Flow

```
Site planner assigns function "social-proof" to a page section
    ↓
RenderComponentAction looks up content_components WHERE function = 'social-proof'
    → returns {rendered_html (with <style>), component_id, component_function: "social-proof"}
    ↓
CompilePageSectionsAction joins HTML, preserves sections_metadata array
    → each entry has {rendered_html, component_id, component_function}
    ↓
SavePageSectionsAction stores to page_components:
    slot_name = "social-proof"  (from component_function)
    component_id = uuid          (from component_id)
    rendered_html = "..."        (includes <style> block)
    ↓
page-rerender reads page_components, concatenates rendered_html in order
    → CSS intact, component linkage intact
```

## Adopted Sites

When importing/adopting an external site, the adoption pipeline must:

1. Parse the external HTML to identify sections
2. Map each section to the closest match in our component library
3. Use `NormalizeComponentFunction()` to clean up any names
4. Either match an existing component or create a new one with a conforming function name
5. **Never import external naming conventions** — always translate to our standard

Example: an adopted site has `<div class="testimonial-area" id="reviews">`. The adoption agent maps this to our `social-proof` component (or `testimonials` if that's a better fit), creates the `page_components` row with `slot_name = 'social-proof'`, and links `component_id`.

## Adding a New Component

1. Choose a `function` name following the naming patterns above
2. Validate it: `ValidateComponentFunction(function)` must return nil
3. Ensure uniqueness: no other active component has this function
4. Write the template with `data-component="{function}"` on the root element
5. Validate the template: `ValidateComponentTemplate(function, template)` must return nil
6. Insert into `content_components` — the DB constraints will reject violations

## Go Validation Functions

Located in `component_validation.go`:

- **`ValidateComponentFunction(function string) error`** — checks kebab-case format
- **`ValidateComponentTemplate(function, htmlTemplate string) error`** — checks `data-component` matches `function`
- **`NormalizeComponentFunction(function string) string`** — converts `social_proof` → `social-proof`, `SocialProof` → `social-proof`

## Lookup Safety Net

`GetComponentWithFallback` in `component_library.go` tries three steps:

1. Exact match on `function`
2. Normalized form (underscore→hyphen, lowercase)
3. Generic fallback (`generic-text-block`)

This prevents silent failures from legacy data but should not be relied upon — callers should pass correct kebab-case names.

## Migration From Legacy Names

The `component_naming_standardization.sql` migration handles the one-time cleanup:

| Before | After | Reason |
|---|---|---|
| `social_proof` | `social-proof` | Underscore → hyphen |
| `call_to_action` | `call-to-action` | Underscore → hyphen |
| `hero` (5 components) | `hero`, `about-hero`, `services-hero`, `contact-hero`, `case-studies-hero` | Uniqueness |
| `site-header` (3 components) | `header-professional-dark`, `header-minimal-light`, `header-bold-gradient` | Uniqueness |
| `site-footer` (3 components) | `footer-4-column`, `footer-standard`, `footer-simple` | Uniqueness |
| `head` (2 components) | `head-seo-standard`, `head-basic` | Uniqueness |
| `testimonials` template with `data-component="social-proof"` | `data-component="testimonials"` | Attribute sync |

After migration, the DB constraints prevent regression.