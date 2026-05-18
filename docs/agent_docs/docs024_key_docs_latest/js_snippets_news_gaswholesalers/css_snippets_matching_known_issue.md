# css_snippets matching needs slug/lemma flexibility OR applies_to needs updating

**Status:** Known issue, not yet fixed. Add to debugging guide Section 9.

**Symptom:** Sites have many `css_snippets` rows whose `applies_to` arrays
look like they should match the site's components, but they never reach
the generated styles.css. For gaswholesalers (verified 2026-05-17, orch
`0c1b3dbf-53e6-41c2-9a11-032d737d1136`):

| css_snippet | applies_to | Should match? | Actually matched? |
|---|---|---|---|
| `hover-lift` | `["card", "feature", "testimonial"]` | yes (site has `features`, `testimonials`) | NO |
| `card-bordered` | `["card", "feature"]` | yes | NO |
| `hover-glow` | `["button", "card", "cta"]` | yes (site has `call-to-action`) | NO |
| `Latest News Grid` | `["latest-news"]` | yes (site has `latest-news`) | YES |
| `News Listing Page` | `["news-listing"]` | yes (site has `news-listing`) | YES |

**Root cause:** The matching query does exact-text overlap between
`css_snippets.applies_to` and the site's reported component functions.
The applies_to uses **generic categorical terms** (`card`, `feature`,
`button`, `cta`) but the system reports **specific component names**
(`features`, `testimonials`, `differentiators`, `call-to-action`). No
exact-text overlap → no match.

So `hover-lift` with applies_to `["card", "feature", "testimonial"]` will
NOT match a site whose components include `features` (plural) or
`testimonials` (plural). Singular/plural is the simplest mismatch; lemma
families (`button` vs `cta-button` vs `primary-button`) are wider.

**Two paths to fix, both legitimate:**

1. **Update applies_to values to match actual component names.** Audit
   every css_snippets row, change e.g. `["card", "feature"]` to
   `["features", "differentiators", "info-card-grid", ...]`. Tightly
   coupled, breaks when new component names appear, manual work, but
   keeps the matching query simple.

2. **Make the matching query lemma-aware / slug-aware.** Strip plurals,
   accept word-stem overlap, accept hyphenated subsets (`card` matches
   `*-card` and `card-*`). Loose coupling, accommodates future
   components without changes to applies_to. Risk of false-positive
   matches (e.g. `card` matching `scorecard` accidentally).

Path 2 is the right long-term answer. Implementation in
`loadComponentCSSSnippets` in `render_css_from_spec_action.go` — would
expand the EXISTS subquery to also match suffix/prefix patterns, or
preprocess component names through a lemmatizer (Postgres has
`unaccent` and basic stemming).

**Diagnostic:**

```sql
-- For a site, which css_snippets WOULD match if we used loose word matching?
WITH site_components AS (
  SELECT DISTINCT cc.function
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '<site-id>'
    AND cc.function IS NOT NULL
)
SELECT
  s.name,
  s.applies_to::text,
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

That query shows which snippets would match under three loosening rules
(exact, plural-of-singular, substring). Useful for sizing the fix and
catching false positives before deploying loose matching.

**Until this is fixed:** Visual effects on cards, buttons, hovers etc.
won't propagate even when css_snippets exist for them. Sites will be
visually plain even though the snippet library is rich.

**Cross-references:**
- `load_site_for_design` 5-item fallback list — fixed 2026-05-16, but
  this css_snippets matching issue surfaces only AFTER the load fix
  starts returning real component lists, because before that the
  fallback list (`hero`, `services-grid`, ...) had its own mismatch
  with the snippet library.
