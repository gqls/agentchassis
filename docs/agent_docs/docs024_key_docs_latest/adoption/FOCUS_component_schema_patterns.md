# FOCUS — Component Schema Patterns

A reference covering three related issues discovered during Step 3 testing:
the numbered-field anti-pattern in component schemas, the missing `game`
page_type in adoption, and the hyphen-vs-underscore drift on page_type
values. Not a plan — a map of what's known and where the questions live.

## The canonical pattern: items-array

The platform's queryresolve system pulls lists of pages from the database
and merges them into rendered sections. Components that display lists
(tool cards, guide cards, blog post cards, navigation links) should
declare a single `items` array field with per-item sub-fields:

```json
"items": {
  "type": "array",
  "source": "query.pages_where_type:tool",
  "limit": 6,
  "min_items": 1,
  "required": true,
  "items": {
    "title":            { "type": "text" },
    "url":              { "type": "url"  },
    "meta_description": { "type": "text" },
    "nav_label":        { "type": "text" }
  }
}
```

The template then renders the list with `{{range .items}}`, producing
however many cards the resolver returned. The component scales with the
data. No hardcoded count. No LLM invention.

**tool-list** is the canonical working example. Step 3 verified it: 6
real tool pages render correctly, with section heading and intro written
by the LLM and items pulled from the database.

## The anti-pattern: numbered flat fields

Some older components declare list data as numbered flat fields:

```json
"game1_title": { "type": "text", "source": "llm", "required": true,
                 "llm_guidance": "Title of the first featured game..." },
"game1_genre": { "type": "text", "source": "llm", "required": true, ... },
"game2_title": { "type": "text", "source": "llm", "required": true, ... },
...
"game6_platform": { ... }
```

This hardcodes the count (always exactly N cards), forces the LLM to
invent each one when source is `llm`, can't be sourced from queryresolve
(queryresolve returns arrays — the schema's flat fields don't match),
and locks the template into N hardcoded card blocks.

**game-list** is the clearest example. 51 fields, 6 cards × ~7 LLM-required
fields each, plus structural section fields. The LLM is *required* to
invent 6 game titles, genres, ratings, platforms, and descriptions.
Step 3's prompt rule against inventing named entities cannot save it
because the schema demands invention by its own structure.

## Numbered-pattern scope (May 2026 audit)

Twenty-five active components match the numbered-field signature
(`<prefix>1_<field>` `<prefix>2_<field>` `<prefix>3_<field>` all present).
They cluster into recognisable groups:

**Navigation (8 components)**: `site-header`, `site-footer`,
`header-docs`, `header-with-cart-or-nav`, `header-with-categories`,
`header-with-search`, `site-head`, `footer-with-disclaimer`. All use
`nav_link_N_*` or `col_N_*` patterns. Nav links are curated by admin,
not query-resolvable; needs a different source (`nav.header_items` or
similar) or per-site spec data.

**Card/grid (7 components)**: `game-list`, `case-studies-grid`,
`archetype-combinations`, `featured-inventory`, `blog-listing`,
`product-details`, `provocation-feed`. Each is logically a list of N
entities — straight candidates for the items-array rewrite, sourced
from `query.pages_where_type:X` or equivalent.

**Tier/stat/spec (5 components)**: `pricing`, `content-block-about`,
`system-stats`, `platform-comparison`, `archetype-result-card`. Curated
content, bounded count, typically authored per-site. May fit
`source: site_specs.<aspect>.items` once site_specs supports array data.

**Tool-internal field clusters (5 components)**: `ai-readiness-quiz`,
`tool-agent-complexity-estimator`, `tool-equity-release`,
`tool-gripper-payload-calculator`, `tool-prompt-architect`. Different
shape — these are heterogeneous fields for a single form or tool
configuration, not lists of similar items. Items-array rewrite may not
fit; review case-by-case.

## Missing page_type: `game`

The platform-mission framing identifies three primary content types:
**tools, guides, and games**. Adoption supports the first two:

- `tool` — first-class, used for individual calculator/simulator pages
- `blog_post` — used for guides (named via the blog convention)

But there is no `game` page_type. The distinct values currently used:

```
blog-index, blog_index, blog-post, blog_post, content, entity_directory,
entity_page, index, news-index, tool
```

When a source site has playable game pages, the adoption-time classifier
has no `game` label to assign. They get classified as `content`, `tool`,
or `entity_page`. As a result, `query.pages_where_type:game` always
returns zero items even when the source content exists.

This affects more than game-list: any component declaring
`source: query.pages_where_type:game` won't resolve until the classifier
recognises the type.

## Page-type normalisation drift

The distinct list contains both `blog-index` and `blog_index`, and both
`blog-post` and `blog_post`. The hyphenated and underscored variants
coexist. This is a normalisation gap in the classifier or page-creation
path. Any code filtering by page_type currently has to handle both.

The right fix is a one-time canonicalisation (decide on underscores,
update existing rows) and a normalising step at write time so the drift
doesn't return. Not done; flagged here for visibility.

## What the audit query catches and misses

The regex `"<prefix>1_<field>"` / `"<prefix>2_<field>"` / `"<prefix>3_<field>"`
catches genuine numbered patterns. It misses:

- Components with array-shaped data but only 2 items (`primary_*` and
  `secondary_*`). These have the same conceptual issue but no `_3` to
  match. Probably rare; worth a secondary sweep if we tackle the
  migration.
- Components that fake items via a `count` field plus shared structure
  (e.g. `card_count: 6` plus a single `card_template_html`). Doesn't
  appear to exist in the corpus.

The 25 matches are likely the full set. Worth one re-run if patterns
expand.

## Relationship to Step 3

Step 3 closed the LLM-vs-resolved boundary for components that already
declare items correctly. tool-list works because:

1. Its schema declares `items` as an array with `source: query.X`
2. plan_sections resolves the items array to real page rows
3. The new `merge_with` config injects resolved items into the render
4. The LLM is asked only for section-level fields (heading, intro, CTA text)
5. The template uses `{{range .items}}` to render whatever count was resolved

For the 25 numbered-pattern components, Step 3 cannot help. Their
schemas don't have an items array to resolve; their templates don't
have a range loop to render through; their LLM contracts demand
invention by structure. The fix is per-component schema and template
rewrite, not workflow or prompt.

## Recommendations to defer

These follow naturally from the findings; calling them out so they're
visible without committing to immediate action:

1. **Add `game` to the classifier vocabulary** so games-as-pages can be
   recognised at adoption time. Without this, game-list can't resolve
   even after rewriting.

2. **Canonicalise page_type values** (drop the hyphen variants). One-time
   migration plus a normalising step at write.

3. **Document the items-array contract** in `003_contracts_and_standards`
   or `019_tool_library`. The component-creator agent's prompt should
   refuse the numbered-flat shape and produce items-array shapes for any
   "list of N similar things" component.

4. **Migrate the 7 card/grid components** to items-array shape in
   priority order. Each is a per-component rewrite (schema + template)
   plus a regeneration sweep for any site using them.

5. **Decide on a curated-list source vocabulary** (`nav.X`,
   `site_specs.X.items`, `entity.X`) for the navigation and tier/stat
   components that aren't query-resolvable from pages. This is a
   queryresolve extension, not a per-component fix.

6. **Component-creator-level guard**: when component-creator generates a
   new schema for a "list of N items" intent, it must produce an items
   array, not numbered flat fields. Worth a unit test on the agent's
   output, not just a prompt rule.

7. **Site-planner-level guard**: when a section's items-array resolves
   to zero items at plan time, the planner should either defer the
   section or substitute a no-items variant (e.g. a "coming soon"
   state). Today the section renders with whatever the template's
   fallback produces, which may be hardcoded sample content.

## What this document is for

A starting reference. When game-list (or any other numbered-pattern
component) comes up for fixing, this is where to look for the
list-of-affected-components, the canonical replacement shape, and the
adjacent issues (game page_type, normalisation drift) that need to be
solved together for a clean fix.

Not a roadmap, not a commitment to fix order, not a prescription. The
work is real but it's separate from Step 3's contract — Step 3 fixed
fabrication where the schema enabled it, and surfaced the schema-level
work that remains.
