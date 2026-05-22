# FOCUS — Visual pipeline: component-list resolution and CSS-snippet matching

Why deployed sites render visually bare even though the CSS-snippet library
is rich. Two stacked causes, both surfaced on gaswholesalers.com in 2026-05:
the component-list that drives CSS selection was a hardcoded fallback rather
than the site's real components, and even with the right list the snippet
`applies_to` matching is too literal to match real component names.

> Consolidated from `002_how_webdesign_handles_snippets.md`,
> `design_actions_status_filter_fix.md`,
> `design_actions_observability_patch.md`, and
> `css_snippets_matching_known_issue.md`. The diagnostic versions of these
> (the "css_snippets missing from styles.css" and "assumed-status-values
> trap" entries) live in `016_debugging_guide_addenda.md`; this doc is the
> mechanism and the fixes.

## The pipeline in one line

`webdesign-agent` runs `load_site_context` → stashes
`site_context.all_component_functions` → `render_css_from_spec_action.go`
calls `extractCSSComponents(collectedData)` → that list is the input to
`loadComponentCSSSnippets`, whose `WHERE applies_to && <component list>`
selects which `css_snippets` get concatenated into the site's
`assets/css/styles.css`. So the deployed CSS is only as good as that
component list and that match.

## Cause 1 — the component list was a hardcoded fallback (the "fake 5-item list")

### What was observed

Querying `collected_data -> 'site_context' -> 'all_component_functions'` on
the three most recent webdesign-agent orchestrations across two different
sites returned the **same 5 items** every time, only the order varying:

```
["hero", "services-grid", "differentiators", "social-proof", "call-to-action"]
```

That is exactly the hardcoded fallback list in `extractCSSComponents`
(`render_css_from_spec_action.go` ~line 363) — the value the function returns
when `site_context.all_component_functions` is missing. For gaswholesalers the
*real* list is 25 components (about-content, call-to-action, case-studies-list,
contact-form, contact-info, differentiators, faq, features,
generic-text-block, hero, hero-about, hero-case-studies, hero-contact,
hero-services, info-card-grid, latest-news, pricing, services-grid,
social-proof, testimonials, four tool components, use-cases-list). The same 5
fake items came back for a different site too — conclusive that it was the
fallback, not the real inventory.

### Why it matters

Every `css_snippet` whose `applies_to` contains only functions outside those
5 never reaches any site. Of 21 snippets seen, only `fade-in-up` (matches
"hero") ever shipped; the other 20 (button styles, card styles, hover effects,
news grids, section-specific styling) were dead weight. This is why every site
had the same surprisingly-bare aesthetic.

### Root cause and fix — the status filter

The fallback fired because `load_site_for_design`'s `loadPagesWithComponents`
returned an empty slice, so `allComponents` stayed empty. The query filtered:

```go
WHERE p.site_id = $1
  AND p.status IN ('deployed', 'published', 'draft', 'planned')
```

But every page on every site has `status = 'active'` — those four assumed
status names don't exist in the data. The filter excluded everything. The fix:

```go
WHERE p.site_id = $1
  AND p.status = 'active'
```

(matching the pattern `LoadSitePagesAction` already uses). An older, larger
bug — `loadPagesWithComponents` reading from `pages.sections` instead of
`page_components` — had masked this status filter for months; when that was
fixed, the status filter became the next layer of the onion. Verify before
applying that `'active'` is the dominant value platform-wide:
`SELECT DISTINCT status, COUNT(*) FROM pages GROUP BY status`.

### Observability so this can't hide again

A patch to `LoadSiteForDesignAction` makes the fallback loud and the code path
externally visible. A `Warn` fires when the fallback is used on a site that
has built pages (which should never happen), and four signature fields are
written into the result map (and thus into `orchestration_states.collected_data`)
so an external observer can confirm which code path ran:

```go
usingFallback := false
if len(allComponents) == 0 {
    usingFallback = true
    params.Logger.Warn("LoadSiteForDesignAction: NO COMPONENTS FOUND — using hardcoded 5-item fallback list. "+
        "For a site with built pages this indicates page_components is empty or the query is broken.",
        zap.String("site_id", id.String()),
        zap.Bool("queried_pages", includePages),
        zap.Int("pages_loaded", pagesLoaded),
        zap.Int("built_components_total", builtTotal),
        zap.Int("planned_components_total", plannedTotal))
    for _, f := range []string{"hero", "services-grid", "differentiators", "social-proof", "call-to-action"} {
        allComponents[f] = true
    }
}
// ... build funcSlice ...
result["all_component_functions"]   = funcSlice
result["pages_loaded"]              = pagesLoaded
result["built_components_total"]    = builtTotal
result["planned_components_total"]  = plannedTotal
result["used_fallback_components"]  = usingFallback
```

`pages_loaded`/`built_components_total`/`planned_components_total`/
`used_fallback_components` are absent in the old code, so their presence in
`collected_data` is proof the new code is live, and `used_fallback_components`
tells you instantly whether a given run degraded to the fallback.

## Cause 2 — applies_to granularity mismatch (known issue, not yet fixed)

Even once `all_component_functions` returns the real list, many snippets still
won't match. `loadComponentCSSSnippets` does exact-text overlap between
`css_snippets.applies_to` and the component functions. `applies_to` uses
**generic categorical terms** (`card`, `feature`, `button`, `cta`); the system
reports **specific component names** (`features`, `testimonials`,
`differentiators`, `call-to-action`). No exact overlap → no match:

| css_snippet | applies_to | Should match? | Matched? |
|---|---|---|---|
| `hover-lift` | `["card","feature","testimonial"]` | yes (`features`,`testimonials`) | NO |
| `card-bordered` | `["card","feature"]` | yes | NO |
| `hover-glow` | `["button","card","cta"]` | yes (`call-to-action`) | NO |
| `Latest News Grid` | `["latest-news"]` | yes | YES |
| `News Listing Page` | `["news-listing"]` | yes | YES |

Singular/plural is the simplest mismatch; lemma families (`button` vs
`cta-button` vs `primary-button`) are wider.

### Two fix paths

1. **Update `applies_to` to real component names.** Audit every snippet, change
   `["card","feature"]` → `["features","differentiators","info-card-grid",...]`.
   Tightly coupled, breaks when new component names appear, manual — but keeps
   the match query simple.
2. **Make the match lemma/slug-aware.** Strip plurals, accept word-stem overlap
   and hyphenated subsets (`card` matches `*-card` and `card-*`). Loose
   coupling, accommodates future components without `applies_to` edits. Risk of
   false positives (`card` matching `scorecard`). Implement in
   `loadComponentCSSSnippets` (expand the EXISTS subquery to suffix/prefix
   patterns, or preprocess names through a lemmatizer; Postgres has `unaccent`
   and basic stemming).

Path 2 is the right long-term answer. A sizing query that shows what would
match under three loosening rules (exact / plural-of-singular / substring),
useful for catching false positives before deploying loose matching:

```sql
WITH site_components AS (
  SELECT DISTINCT cc.function
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '<site-id>' AND cc.function IS NOT NULL
)
SELECT s.name, s.applies_to::text,
  ARRAY(
    SELECT comp.function
    FROM site_components comp,
         jsonb_array_elements_text(s.applies_to) AS applies_elem
    WHERE comp.function = applies_elem
       OR comp.function = applies_elem || 's'
       OR comp.function LIKE '%' || applies_elem || '%'
  ) AS would_match_components
FROM css_snippets s
WHERE COALESCE(s.is_active, true) = true
ORDER BY s.name;
```

## How the two causes stack

Cause 1 (fallback fires) was fixed 2026-05-16, but Cause 2 only becomes
*visible* after Cause 1 is fixed: while the fallback was firing, the 5 fake
names had their own mismatch with the snippet library, so the granularity
problem was masked. With the real component list now flowing, the `applies_to`
granularity is the remaining reason snippets don't match. Until it's addressed,
visual effects on cards/buttons/hovers won't propagate even when the snippets
exist — sites stay plain despite a rich library.

## References

- Code: `render_css_from_spec_action.go` (`extractCSSComponents`,
  `loadComponentCSSSnippets`), `design_actions.go`
  (`LoadSiteForDesignAction`, `loadPagesWithComponents`)
- Diagnostic entries: `016_debugging_guide_addenda.md`
- The same `load_site_context`/component-list plumbing also feeds JS-snippet
  selection — see `FOCUS_news_rendering_and_component_assets.md`
