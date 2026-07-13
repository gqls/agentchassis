# FOCUS — Components: Resolving List/Directory Section Data

**Date:** 2026-05-06
**Status:** Architecture decision documented; implementation not started.
**Related decisions already in the system:** Doc 002 line 776 names `directory-builder` as a Phase 2 agent; doc 003 lines 660–690 define `query.{name}` as a source type and assign it to "Blog/content" components; doc 006 establishes the news-pipeline JSON-file pattern; doc 024 establishes the populate_nav_tables → render-context pattern.

---

## What this is for

To produce a working `tool-list`, `guide-list`, `game-list`, `news-listing`, `directory-listing`, `product-grid`, `services-grid`, `case-studies-grid`, etc. — components that show a list of *things* — the system needs a way to resolve which things go in the list and where to source their data. Today there is no such mechanism. The LLM is asked to fabricate items because the input_schema declares fields like `post1_title` with `source: "llm"`. That's the fabrication problem we observed on the deployed gamesdesign site.

This focus doc captures the architectural choice (already half-decided in prior docs but never implemented) and describes what needs to be built.

---

## What's already decided in the system

These are NOT proposals. These are decisions the project has already made and encoded in docs.

### From doc 003 (contracts) lines 658–690

The contract defines exactly seven source types for component fields:

| Prefix | Resolution | When resolved |
|---|---|---|
| `llm` | Content writer generates | At content generation time |
| `site_specs.{aspect}.{path}` | Lookup in `site_specs` | At `plan_sections` time |
| `site_assets.{type}` | Check site asset exists | At `plan_sections` time |
| `pages.{name}` | Check page exists, resolve URL | At `plan_sections` time |
| `config.{path}` | Read from site config/specs | At `plan_sections` time |
| `renderer` | Assembled by page-rerender agent | At render time |
| `static` | Use fallback value always | At `plan_sections` time |
| `query.{name}` | DB query at render time (blog posts, categories) | At render time |

Listing/directory/grid components are explicitly assigned `source: "query.*"` per the same doc:

> Blog/content | `source: "query.*"` | Dynamic DB queries

### From doc 002 (architecture) line 776

`directory-builder` is listed as a Phase 2 agent (alongside `tool-builder`, which already exists). It's not been built yet but the architectural place is reserved.

### From doc 002 lines 547–562 — Entity Data Agent Family

The pattern used for entity directories: `entity-data-agent` runs in setup mode (configure source → fetch initial data → render pages → build directory) and discovery mode (check sources for changes). Real-time data uses client-side JS that fetches at view time.

### From doc 006 (news feed) and doc 022 (dynamic apps)

A working pattern exists for dynamic content: write data to git-tracked JSON files (`/data/latest-news.json`, `/data/news-archive.json`), have components fetch client-side. Doc 022 explicitly lists this as Tier 1 capability. Production: gaswholesalers.com has a working `news-listing` component using this pattern.

### From doc 024 (link management) — populate_nav_tables pattern

`populate_nav_tables` reads pages → classifies them into nav groups → populates `site_nav_groups` and `site_nav_items` → `render_site_components` injects `{{.nav_items_html}}` into header/footer templates at render-context build time. This is also a working pattern.

---

## The actual problem in concrete terms

A page in the `pages` table has a `sections` array. Each section names a `function` (e.g. `tool-list`). When `plan_sections` runs, it:

1. Looks up the component in `content_components` by function.
2. Reads the component's `input_schema.fields`.
3. For each field, resolves the `source` to a concrete value. If `source == "query.*"`, returns `nil, true` and skips.
4. Returns a `resolvedData` map for the section.
5. Hands the resolved data to the content writer (LLM), which produces HTML using the fields.
6. The HTML is stored in `page_components.rendered_html`.
7. The renderer concatenates `rendered_html` into the page.

**There is nowhere in this flow that `query.*` gets resolved.** The contract says "at render time," but `rerender_single_page_action.go` does pure HTML concatenation — no template engine, no field substitution. The actual resolver doesn't exist.

Today, components like `blog-listing` work around this by declaring 6+ individual fields with `source: "llm"`:

```
post1_title, post1_excerpt, post1_url, post1_image_url, post1_date, post1_category, post1_author_name, post1_author_avatar, post1_read_time, post1_image_alt
post2_title, ... (same shape repeated for posts 2-6)
```

The LLM fabricates these fields. Result: list pages render with invented content; URLs point at pages that don't exist; same problem this focus doc is here to fix.

---

## What needs to be built

A `directory-builder` agent (per doc 002's Phase 2 plan), with these responsibilities:

### Inputs

A work item with:
- `site_id`
- `page_name`
- `section_name` (e.g. `tool-list`)
- `component_id` (resolved by plan_sections)
- The component's `input_schema` (read by the agent itself)
- The site's `pages` and `site_specs` rows (read by the agent)

### Behaviour

For each field in the component's `input_schema` whose `source` starts with `query.`:
1. Parse the query name (e.g. `query.tools_on_site`, `query.blog_posts_recent`, `query.guides_under_section:guides`).
2. Execute the corresponding query against the database (NOT against an LLM).
3. Project results to the field's expected shape (`type: "array"` with `items: { ... }` describing each element).
4. Write results to `page_components.content_data` keyed by section/slot.

### Outputs

- `page_components.content_data` updated with resolved data.
- A signal back to the parent page-build orchestration that the section is ready.

### Query DSL (concrete vocabulary)

Define a small set of query types up front. Don't try to make this open-ended:

| Query name | Semantics | SQL shape |
|---|---|---|
| `query.pages_where_type:tool` | All tool pages on this site | `SELECT name, url, title, meta_description FROM pages WHERE site_id=? AND page_type='tool' AND in_nav=true ORDER BY nav_order` |
| `query.pages_where_type:blog_post` | All blog posts | `WHERE page_type='blog_post' ORDER BY published_at DESC` |
| `query.pages_under_section:<name>` | Pages with parent_section=<name> | `WHERE parent_section=$<name>` |
| `query.pages_where_type:entity_page&entity_type:<type>` | Entity pages of a given type | `WHERE page_type='entity_page' AND ...` |
| `query.case_studies` | Site_specs.case_studies projected to array | `SELECT data FROM site_specs WHERE aspect='case_studies'` (then unwrap JSON) |
| `query.team_members` | Site_specs.identity.team | similar |

Each query resolves to a concrete SQL statement. New query types are added by extending a registry in the agent — not by the LLM inventing query names.

### How `query.*` field schemas should look

Today (broken):

```json
"post1_title": { "source": "llm", "required": true, "llm_guidance": "Title of post 1" },
"post2_title": { "source": "llm", ... },
... (12+ more fields per item)
```

After (proposed):

```json
"items": {
  "type": "array",
  "source": "query.pages_where_type:blog_post",
  "min_items": 1,
  "items": {
    "title":           { "type": "text", "source": "field.title" },
    "url":             { "type": "url",  "source": "field.url" },
    "excerpt":         { "type": "text", "source": "field.meta_description" },
    "published_date":  { "type": "text", "source": "field.published_at" }
  },
  "limit": 6,
  "order_by": "published_at DESC"
}
```

The component's html_template uses `{{range .items}}...{{end}}` instead of `post1_*`/`post2_*`/`post3_*` repetition.

This means **component-creator's prompt also needs updating** to generate this shape for list/directory/grid components. That's a follow-up but the agent design and the component-creator change must land together (a directory-builder that resolves `items.source: query.*` is useless if no components declare `items.source: query.*`).

---

## Where the agent slots into the pipeline

`page-build-handler` workflow today:

```
ensure_site_record
  → load_page_record
    → check_page_found
      → plan_sections
        → check_has_ready_sections
          → spawn_content_writer
            → call_content_writer
              → check_content_produced
                → validate_content
                  → save_sections
                    → update_status
                      → spawn_rerender → deploy_page → complete
```

The directory-builder fits between `plan_sections` and `spawn_content_writer`. Today, plan_sections punts on `query.*` and the field arrives at the content writer empty (so the LLM fabricates). With a directory-builder step:

```
plan_sections
  → check_has_query_sections      (NEW: any sections with query.* fields?)
    → spawn_directory_builder      (NEW: only if yes)
      → directory_builder writes content_data, signals ready
        → spawn_content_writer (LLM only fills llm-source fields; query.* are pre-populated)
          → ...
```

This matches the orchestrator-spawns-handler pattern used for `component-creator`, `feed-ingester`, etc.

Alternative: `directory_builder` writes content_data BEFORE call_content_writer. The content writer sees pre-filled items and only generates the LLM-source fields (heading, eyebrow, etc.). HTML template at render time uses both.

---

## Why path 2b (dedicated agent), not path 2a (inline step)

Both options were considered. Going with the dedicated agent (per the user's choice and per doc 002):

**Reasons for dedicated agent:**

1. **Doc 002 already named it.** `directory-builder` is the canonical name. Reinventing it as an inline step would create a parallel codepath against existing architectural intent.
2. **Reusability.** The same agent serves multiple use cases: tool lists on tool indexes, blog post grids on blog indexes, entity directories on entity-directory pages, case studies on case-study indexes. These all need the same query→array projection, just with different query names.
3. **Independent triggering.** When a new tool page is added to a site, the tool-list section on the index goes stale. A scheduled or signal-driven re-run of directory-builder for that section refreshes it without requiring a full page rebuild. An inline step doesn't have this surface.
4. **Cache invalidation.** A dedicated agent owns the "list of things" logic and can be paired with a reaper/scheduler to detect when underlying data has changed (new page added, page removed, page type changed) and re-resolve. Inline steps would miss this.
5. **Matches existing patterns.** `component-creator`, `feed-ingester`, `entity-data-agent`, `tool-improver` are all dedicated handlers consuming specific work-item types. `directory-builder` consuming `needs_directory_section` items (or similar) follows the same shape.

**The tradeoff:** more code than option 2a. One agent definition, one workflow definition, one Go action. But the cost is one-time and the value is structural.

---

## What about `needs_section_data` items

We saw 41 of these stuck at `needs_human_review` across the platform, plus 6 we manually completed for gamesdesign.

Looking at `plan_sections_action.go` line 1140, these are emitted with **`status='needs_human_review'` directly** — they're HITL by design, not "fetch this data async." The trigger is "couldn't resolve component or required field." Two distinct sub-cases get conflated:

1. **Component missing** — `function = ""`, `component_id = ""` in spec. Today's case. Should emit `needs_component:<function>` to component-creator instead.
2. **Required field can't be sourced** — `missing` array populated. Genuine HITL case (e.g. team members not configured).

`needs_section_data` is NOT the same as `needs_directory_section`. Once directory-builder exists:

- Sections with `query.*` fields go through the new path, not via `needs_section_data`.
- `plan_sections` should split case 1 (emit `needs_component:*` and let regen pipeline resolve) from case 2 (keep `needs_section_data` for genuine HITL).
- The 41 stuck items can be triaged: most are case 1 (emit a `needs_component` instead), some are case 2 (real HITL).

This cleanup is separate from but enabled by the directory-builder work.

---

## Implementation order

1. **Define the query DSL.** Concrete list of query names and their SQL semantics. Stored in code, not in the database — these are part of the platform contract, not site config.
2. **Create the agent definition.** `directory-builder` with workflow `read_spec → resolve_queries → write_content_data → mark_ready → complete`.
3. **Write the Go action(s).**
   - `resolve_query_sources_action.go` — given a section's input_schema, find query.* fields, run queries, return data.
   - `write_section_data_action.go` — UPDATE `page_components.content_data` with resolved data.
4. **Wire into page-build-handler.** New step `check_has_query_sections` between `plan_sections` and `spawn_content_writer`. New step `spawn_directory_builder` if any present.
5. **Update component-creator's prompt** to produce `items.source: query.*` for list/directory/grid components instead of `post1_*`/`post2_*` field repetition. (This is a SEPARATE migration but the two should land together — a directory-builder with no components using `items.source: query.*` does nothing.)
6. **Migrate existing list components.** For each affected component (`blog-listing`, `tool-list`, `guide-list`, `game-list`, `news-listing`, `case-studies-grid`, `directory-listing`, `services-grid`, `product-grid`, `info-card-grid`, `archetype-grid`, `departments-grid`, `featured-inventory`, `lobby-grid`), update `input_schema` to use `items.source: query.*` and update `html_template` to use `{{range .items}}...{{end}}`. This is per-component manual work.
7. **Triage the 41 stuck `needs_section_data` items.** Most should be re-emitted as `needs_component:*`. A few will be genuine HITL.

Steps 1–4 are the core build. Steps 5–6 are the rollout (and can happen incrementally — one component at a time). Step 7 is cleanup.

---

## Open questions to resolve before implementation

1. **Where does limit/sort/offset go?** Component schema (per-section override) or query DSL (per-query default)? Probably both — query DSL has sensible defaults, schema can override.
2. **Empty-list handling.** Three options: (a) skip the section entirely, (b) render with placeholder text from a fallback, (c) render an explicit "no items yet" state. Probably component-controlled — schema declares `empty_state_text` and `on_empty: skip|fallback|render`.
3. **Scoping.** A `query.pages_where_type:tool` runs on the *current site*. Is there ever a cross-site case? Probably not in the foreseeable future — defer.
4. **Real-time vs build-time.** The news pattern fetches client-side. The nav pattern is build-time. Directory-builder produces build-time data. If a tool is added between builds, the index page is stale until next rebuild. Acceptable for v1; future enhancement could write results to `/data/<page>-<section>.json` and have the component fetch client-side (mirror of news pattern). Defer.
5. **Reactive rebuilds.** When a new tool page is added, the tool-list section on the index becomes stale. Does directory-builder need to be re-triggered automatically, or is that a reconciler concern? Probably reconciler — when a page is added/removed/type-changed, emit `needs_rerender:<page-with-list-section>`. Defer to reconciler enhancement.

---

## What to build first (smallest viable slice)

Bottom-up: get one list component working end-to-end before generalising.

**Pick `tool-list`** (simplest, internal data, our test site has it):

1. Add `query.pages_where_type:tool` to a query registry stub.
2. Update `tool-list`'s `input_schema` to declare `items.source: query.pages_where_type:tool` plus a small set of LLM-source fields (heading, eyebrow, subtitle).
3. Update `tool-list`'s `html_template` to use `{{range .items}}...{{end}}` (this is a regen, see migration 036 pattern).
4. Build a minimal `resolve_query_sources_action.go` that handles only `query.pages_where_type:tool`.
5. Wire into page-build-handler as a new step before content writer.
6. Test on gamesdesign.co.uk.

If that works, generalise to other query types and other components.

---

## Anti-goals

- Do NOT create a new item type called `needs_directory_data` distinct from `needs_section_data`. Use the existing item-type taxonomy.
- Do NOT make the query DSL LLM-driven. Concrete query names with concrete SQL. The LLM picks WHICH query, but doesn't write SQL.
- Do NOT solve cache invalidation in v1. Defer to reconciler enhancement.
- Do NOT solve the 41 stuck `needs_section_data` items in this work. Separate cleanup.
- Do NOT touch the news pipeline. It works; leave it alone.

---

## See also

- `002_system_architecture.md` line 776 — directory-builder named as Phase 2 agent
- `003_contracts_and_standards_v7.md` lines 658–690 — source prefix table including `query.*`
- `006_news_feed_pipeline_v2.md` — working pattern for dynamic content via JSON files
- `024_link_management_v2.md` — populate_nav_tables pattern (read pages → produce structured nav data)
- `026_component_regeneration_flow.md` — the regeneration pipeline that step 6 above will drive
- `FUTURE_section_data_handler.md` (this session) — earlier draft superseded by this doc
- `HANDOFF_2026-05-06_phase1_deployed_section_data_handler_pending.md` — current state and how this fits in
