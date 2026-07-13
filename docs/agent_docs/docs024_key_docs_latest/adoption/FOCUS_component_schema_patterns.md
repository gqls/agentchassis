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

---

## Appendix — baseline verification queries and the Path C interactive-fingerprint plan

_Folded in from a working-session `README.md`. The six baseline queries verify the before/after of the schema fixes on gamesdesign.co.uk; the lower half is the Path C plan to capture interactive (tool/game) machinery during adoption — the fix for "tool pages have prose, not tools" (problem 3 in `FOCUS_adoption_fidelity_and_variants.md`)._
1. Page inventory — what adoption decided each page is
   sqlSELECT name, page_type, status, build_status, created_at
   FROM pages
   WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
   ORDER BY page_type, name;
   Key thing to check: any pages that should be game but got classified as content/tool/entity_page. This is the "missing vocabulary" signal we'll want to fix as part of Path A (games), but also worth knowing whether adoption noticed any game-like pages on the source.
2. Classification aspect content — what the classifier said the site contains
   sqlSELECT data
   FROM site_specs
   WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
   AND aspect = 'classification'
   AND is_current = true;
   Look for: any mention of games as content type, any page_types enumeration, any signal that the classifier saw games even if it couldn't label them.
3. Strategy aspect — what the strategist proposed
   sqlSELECT data
   FROM site_specs
   WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
   AND aspect = 'strategy'
   AND is_current = true;
   This tells us what the platform intends the site to be, beyond what was adopted directly. Useful contrast with what Path C will surface (parsed source interactivity).
4. Tool-list resolved vs game-list fabricated — side by side
   sqlSELECT p.name, pc.slot_name,
   pc.content_data->>'section_heading' AS heading,
   jsonb_array_length(COALESCE(pc.content_data->'items','[]'::jsonb)) AS resolved_items,
   pc.content_data->>'game1_title' AS fab_g1,
   pc.content_data->>'game2_title' AS fab_g2,
   pc.content_data->>'game3_title' AS fab_g3,
   pc.content_data->'items'->0->>'title' AS resolved_first_title,
   pc.content_data->'items'->0->>'url'   AS resolved_first_url
   FROM page_components pc
   JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
   AND pc.slot_name IN ('tool-list','guide-list','game-list')
   ORDER BY pc.slot_name, p.name;
   The clearest before/after frame. After Path C + Path A, the fab_g* columns should be NULL (schema rewritten away) and resolved_items should be >0 for game-list too.
5. Components in use on this site — which of the 25 anti-pattern components are deployed
   sqlSELECT DISTINCT pc.slot_name, COUNT(*) AS used_count
   FROM page_components pc
   JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
   GROUP BY pc.slot_name
   ORDER BY pc.slot_name;
   Tells us which other numbered-pattern components might be silently fabricating today. From the earlier audit: blog-listing, case-studies-grid, archetype-combinations, featured-inventory, product-details, provocation-feed all share the shape. If any are in use on this site, they're invisible fabrications waiting for Path E.
6. Corrected orchestration query — for future use
   sqlSELECT orchestration_id, owner_agent_type, current_step, status, updated_at
   FROM orchestration_states
   WHERE correlation_id = '1400b739-40b4-4627-80eb-9b76324c572a'
   ORDER BY created_at;
   Use orchestration_id not id everywhere we query this table.

Run those, paste the output, and I'll capture the baseline in a short note alongside the focus doc. Then we start on Path C.
Path C starts with me reading the two existing Go files closely:

extract_design_fingerprint_action.go (line ~62021, 250+ lines)
enrich_fingerprint_with_css_action.go (line ~38055, ~120 lines)

And the surrounding workflow that calls them (likely in the adoption orchestrator workflow JSON). Once I have those clear, I'll propose the concrete changes: new selectors, new fetch step, new aspect names, new LLM prompt. We'll work in step sizes that keep things reversible.


-------------------------

The agent is site-adoption-agent and the design step is named extract_fingerprint. Let me trace through the workflow carefully before writing the patch.
Current workflow flow (relevant section)
crawl_site (firecrawl_crawl, formats: ["markdown", "rawHtml"], limit 30)
→ format_crawl
→ check_crawl_content
├── then: extract_fingerprint   ← design extractor lives here
│           output_field: design_fingerprint
│           next_step: check_has_external_css
│
│      check_has_external_css
│      ├── then: fetch_primary_css → enrich_fingerprint → analyze_site
│      └── else: analyze_site
│
└── else: crawl_failed
crawl_site already requests rawHtml in its formats. The crawled data with rawHtml is in crawl_result. The new action will read from the same crawl_result that the design extractor uses.
Where to insert extract_interactive_fingerprint
Cleanest insertion point: between extract_fingerprint and check_has_external_css.
extract_fingerprint (UNCHANGED action+config, only next_step changes)
next_step:  extract_interactive_fingerprint   ← was "check_has_external_css"

extract_interactive_fingerprint (NEW)
action:     extract_interactive_fingerprint
config:     { crawl_field: "crawl_result" }
output_field: interactive_fingerprint
next_step:  check_has_external_css

check_has_external_css (UNCHANGED)
... unchanged from here on
Why this position:

Both fingerprints run against the same crawl_result
Existing CSS branching downstream is untouched
If the new action returns an empty fingerprint (no pages with rawHtml — edge case), nothing breaks; it just stores an empty result and moves on
No existing workflow variables are renamed

Collision check on variable names
The new output_field: "interactive_fingerprint" doesn't collide with any existing variable in the workflow. Existing names: crawl_result, formatted_crawl, design_fingerprint, css_scrape_result, adoption_analysis, site_archetype_analysis, representative_content, content_direction_analysis, adoption_result, nav_data, design_intent_generated, design_intent_written, site_record. Clean.
