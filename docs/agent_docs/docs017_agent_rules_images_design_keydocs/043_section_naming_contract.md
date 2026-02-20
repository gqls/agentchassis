# Section Context Variable Contract — Enforcement Strategy

Companion to `042_component_naming_contract.md`. All function names in this doc
use kebab-case per the naming contract. All `data-component` attributes must
exactly match `content_components.function`.

## The Contract

Any component with a dark background MUST set these CSS custom properties
on its outermost container element:

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

Global `styles.css` uses these variables with light-theme fallbacks:
- `h1-h6 { color: var(--section-heading, var(--color-primary)); }`
- `p { color: var(--section-text, inherit); }`
- `blockquote { background: var(--section-surface, var(--color-surface)); }`

## Dark Components (as of this migration)

| function | is_dark_section | has --section-* |
|---|---|---|
| `hero` | true | yes |
| `hero-about` | true | yes |
| `hero-services` | true | yes |
| `hero-contact` | true | yes |
| `hero-case-studies` | true | yes |
| `hero-use-cases` | true | yes |
| `social-proof` | true | yes |
| `testimonials` | true | yes |
| `call-to-action` | true | yes |
| `portfolio-showcase` | true | needs manual review |

## Enforcement Layers (defense in depth)

### Layer 1: Database Schema (immediate)

`content_components.is_dark_section` boolean column.

- Set to `true` when creating/updating dark components
- Used by validation code to skip checking light components
- Can be queried to find all dark components: `WHERE is_dark_section = true`
- Verification query catches components that are dark but unmarked:

```sql
SELECT name, function, is_dark_section,
       html_template LIKE '%--section-text%' as has_section_vars,
       CASE
         WHEN html_template LIKE '%background: var(--color-primary%' THEN true
         WHEN html_template LIKE '%linear-gradient%1a1a2e%' THEN true
         ELSE false
       END as looks_dark
FROM content_components
WHERE component_level = 'section'
ORDER BY is_dark_section DESC, function;
```

### Layer 2: Go Validation (build time)

`ValidateDarkSectionContract()` in `validate_dark_section.go`.

Hook into these existing actions:

**A. RenderComponentAction** — after rendering, validate the output:
```go
// In RenderComponentAction, after template execution:
rendered := buf.String()
if missing := ValidateDarkSectionContract(rendered, component.IsDarkSection, logger); len(missing) > 0 {
    logger.Warn("Dark section contract violation",
        zap.String("component", component.Function),
        zap.Strings("missing", missing),
    )
    // Don't fail — just warn. The build continues but logs are flagged.
}
```

**B. SavePageSectionsAction** — when persisting, check each section:
```go
// In the insert loop:
if missing := ValidateDarkSectionContract(section.HTML, false, logger); len(missing) > 0 {
    logger.Warn("Saving dark section without context variables",
        zap.String("slot_name", section.ComponentName),
        zap.Strings("missing", missing),
    )
}
```

These are WARNINGS, not errors. The build doesn't fail — but the warnings
surface in logs so problems are caught in testing, not production.

### Layer 3: LLM Prompt Enforcement (generation time)

**A. webdesign-agent prompt** (015_update_webdesign_prompt.sql):
- Includes the full --section-* pattern in "Required Base Styles"
- Explicitly forbids setting color on p, strong, blockquote directly
- Includes the dark section template as a CSS comment at end of output

**B. page-content-writer prompt** — when the LLM generates section HTML
with inline CSS, it should know the contract. Add to its system prompt:

```
## CSS RULES FOR SECTIONS
- Light sections: use var(--color-*) variables, text inherits from body
- Dark sections (dark background): MUST set these on the container:
    --section-text: rgba(255,255,255,0.9);
    --section-text-muted: rgba(255,255,255,0.7);
    --section-heading: #ffffff;
    --section-surface: rgba(255,255,255,0.05);
    --section-border: rgba(255,255,255,0.2);
  Do NOT set color on individual p, h2, blockquote, strong elements.
  The global CSS reads --section-* variables automatically.
```

**C. Component creation prompts** — any workflow that creates new
content_components should include the contract in the LLM prompt.

### Layer 4: Periodic Audit (maintenance)

SQL queries to run periodically or as part of a health check:

```sql
-- Find dark components missing the contract
SELECT name, function, 
       'MISSING --section-* VARIABLES' as issue
FROM content_components
WHERE is_dark_section = true
  AND html_template NOT LIKE '%--section-text%';

-- Find likely-dark components not flagged
-- Filters out components that use #1a1a2e as text color (not background)
SELECT name, function,
       'PROBABLY DARK BUT NOT FLAGGED' as issue
FROM content_components
WHERE is_dark_section = false
  AND component_level = 'section'
  AND (
    html_template LIKE '%background:%#1a1a2e%'
    OR html_template LIKE '%background: #1a1a2e%'
    OR html_template LIKE '%background: var(--color-primary%'
  )
  AND function NOT IN ('head', 'head-seo-standard', 'head-basic',
                        'site-header', 'header-professional-dark',
                        'header-minimal-light', 'header-bold-gradient');

-- Verify data-component matches function (naming contract)
SELECT name, function,
       CASE
         WHEN html_template LIKE '%data-component="' || function || '"%' THEN 'OK'
         ELSE 'MISMATCH'
       END as data_component_check
FROM content_components
WHERE is_dark_section = true
ORDER BY function;
```

## Files Involved

| File | Change | Purpose |
|------|--------|---------|
| `styles.css` | New version | Uses --section-* with fallbacks |
| `014_section_context_migration.sql` | ALTER + UPDATE | is_dark_section column, dark templates |
| `015_update_webdesign_prompt.sql` | Agent config | Teaches LLM the pattern |
| `validate_dark_section.go` | New file | Go validation function |
| `RenderComponentAction` | Add hook | Warn on render |
| `SavePageSectionsAction` | Add hook | Warn on save |

## Deployment Order

1. Run `014_section_context_migration.sql` against clients_db (adds column, flags + updates templates)
2. Run `015_update_webdesign_prompt.sql` against clients_db (updates webdesign-agent)
3. Deploy `validate_dark_section.go` in next Go build
4. Hook validation into RenderComponentAction + SavePageSectionsAction
5. Deploy updated `styles.css` via webdesign-agent or manual git commit
6. Rerender affected pages (index at minimum for social-proof fix)
7. Check `portfolio-showcase` template and add --section-* variables manually


