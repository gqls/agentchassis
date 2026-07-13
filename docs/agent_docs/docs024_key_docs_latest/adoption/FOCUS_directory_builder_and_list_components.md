# FOCUS — Components: Resolving List/Directory Section Data

**Date:** 2026-05-06
**Status:** Architecture decision documented; implementation not started.
**Related decisions already in the system:** Doc 002 line 776 names `directory-builder` as a Phase 2 agent; doc 003 lines 660–690 define `query.{name}` as a source type and assign it to "Blog/content" components; doc 024 establishes the populate_nav_tables → render-context pattern (which is the structural model for this work); doc 006's news-pipeline JSON-file pattern is a deliberate non-model — see below.

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
| `query.{name}` | DB query at plan time, projected to field-shape | At `plan_sections` time |

Listing/directory/grid components are explicitly assigned `source: "query.*"` per the same doc:

> Blog/content | `source: "query.*"` | Dynamic DB queries

### From doc 002 (architecture) line 776

`directory-builder` is listed as a Phase 2 agent (alongside `tool-builder`, which already exists). It's not been built yet but the architectural place is reserved.

### From doc 002 lines 547–562 — Entity Data Agent Family

The pattern used for entity directories: `entity-data-agent` runs in setup mode (configure source → fetch initial data → render pages → build directory) and discovery mode (check sources for changes). Real-time data uses client-side JS that fetches at view time.

### From doc 006 (news feed) and doc 022 (dynamic apps) — explicitly NOT the model

A working pattern exists for dynamic content: write data to git-tracked JSON files (`/data/latest-news.json`, `/data/news-archive.json`), have components fetch client-side. Doc 022 lists this as Tier 1 capability. Production: gaswholesalers.com has a working `news-listing` component using this pattern.

This pattern is **deliberately not the model for directory-builder.** News has an independent ingestion lifecycle (every 6 hours, async RSS/API sources, two-pass triage and scoring) that's decoupled from site builds. List/directory components don't have that lifecycle — a site's tool list changes when the site's tool pages change, which is itself a build-pipeline event. Build-time resolution is the right fit; client-side JSON fetch would be over-engineering. The news pattern stays as it is; directory-builder doesn't extend it.

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

**There is nowhere in this flow that `query.*` gets resolved.** The contract originally said "at render time," but `rerender_single_page_action.go` does pure HTML concatenation — no template engine, no field substitution. The actual resolver doesn't exist.

### Contract change: `query.*` resolves at plan time, not render time

Doc 003's source-prefix table originally specified "At render time" for `query.{name}`. That stage was aspirational — a render-time resolver was envisaged but never implemented, and the renderer doesn't have a template engine that could consume one.

This work pulls resolution forward to `plan_sections` time. The data flow already routes through `plan_sections → resolved_data → content writer → rendered_html`. Query results travel naturally on this path. Build/render is triggered by the reconciler when the underlying data (pages table) changes, so plan-time data is fresh enough.

The trade-off: query results are fresh as of last build, not as of last view. Acceptable for v1 because list contents change with the site's structure (new tool pages added, removed, or reclassified), and those changes already trigger page rebuilds via the reconciler. There is no facility for render-time resolution and building one would duplicate the resolution logic that already runs at plan time.

This is a real contract change. **Doc 003 line 669 must be updated** to say "At plan time" rather than "At render time" for `query.{name}`. The change is consistent with how the other source types (`site_specs.*`, `pages.*`, `config.*`) work.

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

For v1, no workflow change is needed. `plan_sections` already runs as a step inside the page-build-handler workflow. The query resolution happens inside that step — when plan_sections walks a section's input_schema fields and encounters a `query.*` source, it calls `queryresolve.Resolve` and puts the result into `resolvedData` for that field. The downstream content writer sees the pre-resolved items and uses them rather than fabricating.

```
plan_sections (v1)
  → for each section's input_schema fields:
      llm fields              → mark for content writer
      site_specs.* fields     → resolve via existing path
      query.* fields          → call queryresolve.Resolve, put result in resolvedData  ← NEW
      static / fallback       → existing path
  → spawn_content_writer (LLM produces HTML, sees pre-resolved query data)
    → ...
```

For v2, the dedicated `directory-builder` agent gets a new workflow step inserted:

```
plan_sections (v2)
  → check_has_query_sections      (NEW: any sections with query.* fields?)
    → spawn_directory_builder      (NEW: only if yes — agent calls queryresolve.Resolve)
      → directory_builder writes content_data, signals ready
        → spawn_content_writer
          → ...
```

The migration v1 → v2 is well-defined: `plan_sections` flips back to deferring `query.*` (returning them as needs-section-data work items targeted at directory-builder), and the new agent calls the same resolver package. The Go function `queryresolve.Resolve` is unchanged; only its caller moves.

---

## Implementation strategy: hybrid (resolver package now, agent later)

Three options were considered:

- **(2a) Inline step inside `plan_sections`.** Smallest change. Query resolution becomes part of the field-resolution loop. No new agent, no new workflow. But ties query logic to plan_sections forever; harder to add re-trigger surface later.
- **(2b) Dedicated `directory-builder` agent** per doc 002 line 776. Cleanest separation. Re-triggerable. Most code.
- **Hybrid: build the resolver as a standalone package, called inline from plan_sections initially, with a future directory-builder agent calling the same package later.** The package is the durable artifact; how it's invoked can change without rewriting the resolution logic.

We chose hybrid.

The resolver is `platform/orchestration/actions/queryresolve/`. It exposes one entry point:

```go
queryresolve.Resolve(ctx, db, queryresolve.QueryRequest{
    Name:   "pages_where_type:tool",
    SiteID: siteID,
    Limit:  6,
}, logger) (interface{}, error)
```

For v1, `plan_sections_action.go` calls this directly when it encounters a `query.*` source in the field loop (the new code path replaces what was previously the "skip query.* fields" stub branch). The resolver runs SQL, returns `[]map[string]interface{}` shaped for the field, and plan_sections puts the result into `resolvedData` exactly like any other resolved field.

For a future v2 (per doc 002's roadmap), the `directory-builder` agent gets created. Its workflow has a `resolve_queries` step that calls the **same** `queryresolve.Resolve` function, then writes results to `page_components.content_data`. `plan_sections` switches back to deferring `query.*` and emits a `needs_directory_section` work item; directory-builder consumes that item and runs the resolver out-of-band. This adds re-triggerability and parallelism.

The hybrid keeps both options open and avoids speculative agent construction. It also matches the user's guidance: *"keep workflows simple, put complexity in golang action code."*

### What the focus doc previously called "path 2b" — the dedicated agent — is the v2 we will reach later

The reasons for it are still valid:

1. **Doc 002 already named `directory-builder`.** When we build the agent, that's the canonical name.
2. **Reusability across query types.** The agent will serve tool lists, blog post grids, entity directories, case studies — all using the same resolver package.
3. **Independent re-triggering.** A scheduled or signal-driven run can refresh a single section without full page rebuild.
4. **Cache invalidation.** A dedicated agent + reaper can detect underlying-data changes and re-resolve.
5. **Pattern parity.** Mirrors `component-creator`, `feed-ingester`, `entity-data-agent`, `tool-improver`.

For v1 these benefits don't matter yet — full page rebuilds when pages change is fine — and the resolver-only slice gets us a working `tool-list` end-to-end with significantly less code. v2 is then a thin wrapper around the resolver that already exists.

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
4. **Real-time vs build-time.** Directory-builder produces build-time data — items are resolved when the page is built or rebuilt. The data is fresh as of the last build. This is intentional: list contents change with the site's structure (new tool pages, removed guides, page-type changes), and those changes already trigger page rebuilds via the reconciler. There's no need for client-side fetch or scheduled re-resolution; tying list data to the build cycle is the natural fit.

   The news-pipeline JSON-file pattern (doc 006) is a deliberate exception — news has its own ingestion lifecycle (every 6 hours, async sources, scoring) that's decoupled from site builds. List/directory components don't have that lifecycle and shouldn't borrow the pattern.
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
- `006_news_feed_pipeline_v2.md` — news pipeline (deliberately a different pattern; do not borrow from this for directory-builder)
- `024_link_management_v2.md` — populate_nav_tables pattern (read pages → produce structured nav data)
- `026_component_regeneration_flow.md` — the regeneration pipeline that step 6 above will drive
- `FUTURE_section_data_handler.md` — earlier draft, fully superseded; its content (the needs_section_data origin) is folded into the Implementation history below
- Phase 1 deployment + the build sessions are folded into the Implementation history below

---

# Implementation history and current state

> Consolidated from the section-data origin (`FUTURE_section_data_handler.md`,
> `HANDOFF_2026-05-07_phase1_deployed_section_data_handler_pending.md`), the two
> directory-builder build sessions
> (`HANDOFF_2026-05-08_directory_builder_v1_validator_pending.md`,
> `HANDOFF_2026-05-11_..._page_content_writer_finding.md`), and the Step 2 / Step 3
> change sets that closed the fabrication gap. Session narrative is compressed;
> decisions, migrations, identifiers, gotchas, and open follow-ups are kept.

## Arc in one paragraph

The list-data problem began as `needs_section_data` work items with no handler
(Phase 1 deployment, 2026-05-06): list sections couldn't be populated, so they
were either stuck at `needs_human_review` or silently fabricated. The
architectural fix was decided as this focus doc (resolver package, hybrid v1/v2).
A Tier D component-schema shape (`items` array + `query.*` source) was built and
two components regenerated to it. **Then the key finding (2026-05-11): the
resolver wasn't actually in the live content path — `page-content-writer` was
fabricating the `items` arrays itself.** Step 2 and Step 3 (2026-05-12) closed
that gap: `plan_sections` now resolves `query.*` before the LLM call and carries
the result on `section_plan.sections_ready`; `page-content-writer` was rewritten
to receive only the `source: llm` fields and to merge the resolved data in as
authoritative.

## Origin: needs_section_data had no handler (Phase 1, 2026-05-06)

Phase 1 (declarative site-plan + reconciler, doc 030) deployed and verified
end-to-end on gamesdesign.co.uk (re-adopt correlation `4a06d19f-...`). Migrations
`031`–`037` (site_plans schema, planner workflow, schema repair, prompt tightening,
component-registry hygiene incl. the 42-broken-template deactivation, component-
creator structural self-check). Chassis hash `8584b77648` carried
`write_site_plan_action.go`, `reconcile_site_plan_action.go`, `page_canonical.go`,
`validate_page_content.go`, `store_generated_component_action.go`.

The blocker that opened this whole thread: **41 `needs_section_data` items stuck
at `needs_human_review` across six sites** (leopardessconsulting 17, gaswholesalers
6, robot-hands 6, ai-agent-orchestration 5, finetuning 4, vonc 3), because nothing
dispatched them. For gamesdesign the six were manually completed
(`claimed_by='manual_acknowledge_no_handler'`) to unblock parent pages — so those
sections rendered with empty/fabricated data. `FUTURE_section_data_handler.md`
(now superseded by this doc) first scoped a handler; the correct architecture
turned out to be `query.*` source resolution per doc 003, which is what this doc
specifies.

Also surfaced at Phase 1 deployment (Phase 2 territory, tracked elsewhere): dead
BEM design system in every page head, no theme CSS variables landing,
adoption-shape vs planner-shape nav duplication, and the LLM-fabricated list
sections — see `FOCUS_chrome_templates_and_page_shape.md` and
`FOCUS_adoption_fidelity_and_variants.md`.

## Tier D build (Session 4, 2026-05-08) — chassis v1.0.994

Goal: make list components draw items from a real DB query instead of fabricating.
Smallest slice = `tool-list`. Landed:

- Migrations `038`–`040b`: backups; component-creator prompt Tier D block (with
  the funcMap-helper rewrite after several `replace()` anchor failures — see
  Gotchas); `tool-list` rename + regen work item.
- Go (deployed v1.0.994): new `queryresolve/queryresolve.go` package (single entry
  `Resolve(ctx, db, QueryRequest, logger)`, handles `pages_where_type:<type>`, hard
  cap 24 / default 12); `plan_sections_action.go` query.* handling block;
  `data_helpers.go` funcMap helpers (`rangeStart`/`rangeEnd`).
- Go (pending at the time): `compute_component_quality.go` — `extractSchemaFields`
  array-sub-schema traversal, so the validator stops rejecting Tier D templates
  whose `{{range .items}}{{.title}}` references live in the array sub-schema rather
  than top-level fields.

## Tier D converge + the fabrication finding (Session 5, 2026-05-11)

- `tool-list` hand-written as the canonical Tier D reference (migration `041`,
  quality_score 100). `guide-list` LLM-regenerated (migration `042`): attempt 1
  failed validation on one orphan label (`read_guide_label`), attempt 2 passed —
  the same retry-and-converge pattern as tool-list (which had missed
  `card_link_label`). Retry budget of 3 is calibrated for this single-bookkeeping-
  orphan failure class; central label registry would prevent it entirely.
- **Validation observability deployed**: `store_generated_component_action.go`
  `recordValidationRejection` writes a structured `agent_error_log` row on every
  pre-store rejection (severity warning for bookkeeping, error for structural;
  orphan/unknown field names parsed into typed JSONB arrays). guide-list's
  attempt-1 failure was captured exactly — one SQL row, no pod-log forensics.
- **The finding**: the deployed `tools.html` showed 6 tool cards that *looked*
  like queryresolve output but were produced by `page-content-writer`'s LLM at
  build time. `guide-list` on the index showed 6 *fabricated* guides
  (`/guides/pseudo-random-distribution.html` etc.) matching no real pages.
  llm_call_log confirmed `page-content-writer` step
  `process_sections_loop_iter_N_generate_content` returning the full section JSON
  including the `items` array. It does NOT consult queryresolve or the pages table;
  it generates from LLM context. The tool items "happened to" match real pages
  because adoption context contained the real tool names — coincidence, not
  architecture. Also note: gamesdesign guides are `page_type=blog_post`, so the
  regenerated guide-list's `query.pages_where_type:guide` returns zero rows anyway
  (the missing-`game`/`guide`-vocabulary issue — see
  `FOCUS_component_schema_patterns.md`).

This established that queryresolve was built but never wired into the live content
path — page-content-writer was the de-facto items producer. That is what Step 2/3
fixed.

## Step 2 (2026-05-12) — plan_sections enrichment, shared loader (additive)

The "A1c" sequence: resolve queryresolve before the LLM call and route
pre-resolved items into the prompt. Additive only — no workflow change, no caller
behaviour change.

- `v3_site_actions.go`: shared `loadSectionComponents(ctx, db, sectionNames,
  pageID, activeOnly, logger)` helper; `LoadPageSectionComponentsAction` refactored
  to a thin wrapper over it (passes `activeOnly=false` to preserve its behaviour;
  `plan_sections` passes `true`). Support fns: `scanSectionComponentRow`,
  `buildStubSectionComponents`, `enrichSectionComponentsWithBriefs`.
- `plan_sections_action.go`: adds `sectionPlanItem.Component` (full per-section
  component map — input_schema, html_template, render_mode, description, category,
  content_brief) and `componentInfo.Raw`; `sectionTemplateValid` guard mirrored.
  Behaviour preserved exactly (by-name→by-function fallback, order preservation,
  stub generation, content_brief enrichment, template-truncation guard).
- Diffs: `step2_plan_sections_action.diff`, `step2_v3_site_actions.diff`.

**Verification (PASS, 2026-05-12)** on gamesdesign (note: site re-adopted again,
site_id `9a8baddf-77c2-4486-a56b-1b7dde9c1e9e`): `sections_ready` carried
`Component` end-to-end (2/2 with component+html_template+input_schema; 1/2 with
resolved_data — the all-LLM hero used fallbacks). Caveat carried to Step 3: the
verified run's resolved_data was *fallback*-driven, not *query*-driven — the
query.* path itself still needed verifying on a `tool-list`/`guide-list` section.

## Step 3 (2026-05-12) — targeted prompt, section_plan loop, merge_with

Closes the fabrication gap. Three sub-changes:

- **3a (additive)** `plan_sections_action.go`: `llmFieldSpec` struct +
  `LLMFieldSpecs` on `sectionPlanItem`, populated from each schema field's
  `llm_guidance` (the actual field name, verified against production — not
  `description`). Carries name/type/required/description/on_missing/fallback.
- **3b (substantive)** `page-content-writer` workflow rewrite (SQL `step3b_apply.sql`
  with snapshot guard; standalone `step3b_workflow.json` + `step3b_prompt_template.txt`):
  removed `load_page_components` (its data now arrives on
  `section_plan.sections_ready[*].component`); `process_sections_loop.iterate_over`
  changed to `input_data.section_plan.sections_ready`; `render_section` /
  `render_from_template` read `current_section.component` and gained
  `merge_with: current_section.resolved_data`; conditionals read
  `current_section.component.{render_mode,needs_llm,needs_research}`. The prompt
  was rewritten to iterate `llm_field_specs` (only the `source: llm` fields) and
  show a JSON example built from exactly those keys — the *absence* of `items`,
  `cta_url`, etc. from the example is the boundary. Per the anti-"pink-elephant"
  principle, the prompt does NOT enumerate forbidden field names; it states once
  that lists/URLs/images/labels/DB-resolved data are handled separately. All 17
  non-fabrication strict rules and all upstream context blocks retained.
- **3c (additive)** `v3_site_actions.go` `RenderComponentAction` honours
  `merge_with`: extracts the merge map (e.g. `current_section.resolved_data`) and
  overlays it onto `sectionContentData` AND `renderCtx` after the `content_from`
  LLM-output extraction. Order is deliberate — resolved_data wins on conflict
  (queryresolve output and static fallbacks are authoritative over LLM invention).
  Diff `step3c_v3_site_actions.diff`. Revert of 3b via `revert_agent('page-content-writer')`;
  3c needs no revert (merge_with is optional).

**Verification surface (post-deploy):** (a) `sections_ready[*].llm_field_specs`
present with `spec_count` matching the schema's `source: llm` field count;
(b) `page_components.content_data->'items'` on tool-list/guide-list contains real
`pages` rows (title/url/meta_description), not `Calculator A` → `/tools/calc-a.html`;
(c) the LLM response from `generate_content` contains only the llm_field keys, not
a sprawling items array. Queries in the Step 3 changelog.

## Staged code artifacts (not inlined — they are code, kept as references)

- `queryresolve/queryresolve.go` — the resolver package (v1.0.994).
- `plan_sections_action.go` + `step2_plan_sections_action.diff`, `step3a_plan_sections_action.diff`.
- `v3_site_actions.go` + `step2_v3_site_actions.diff`, `step3c_v3_site_actions.diff`.
- `compute_component_quality.go` + `.diff` — array-sub-schema validator traversal.
- `store_generated_component_action.go` + `.diff` — rejection logger.
- `data_helpers.go` + `.diff` — funcMap helpers.
- `step3b_workflow.json`, `step3b_prompt_template.txt`, `step3b_apply.sql` — the page-content-writer rewrite.
- Migrations `038`–`042`.

## Gotchas worth keeping (hard-won during the build)

- **`replace()` and prompt anchors.** The component-creator prompt has irregular
  indentation (3+5 spaces). Postgres `replace()` is literal-byte and silently
  no-ops on a missed anchor while still reporting `UPDATE 1`. Verify anchor bytes
  via hex dump; test presence with COUNT not LIKE; prefer short distinctive
  anchors. A re-entrant replace (replacement contains the anchor) appends on every
  run — `039b` ran 3× and produced 3 duplicate paragraphs.
- **CASE doesn't short-circuit sub-SELECTs.** `CASE WHEN COUNT(*)=0 THEN (SELECT 1/0) ...`
  evaluates the sub-SELECT eagerly at planning. Use DO blocks with `RAISE EXCEPTION`.
- **Postgres NAMEDATALEN = 63.** Backup table names truncate (`_pre_directory_builder`
  → `_pre_directory_buil`). Keep names short or reference the truncated form.
- **Prompt templates go through Go `text/template`.** Any literal `{{...}}` in a
  prompt is parsed as an action. Use funcMap helpers: `{{placeholder "X"}}` →
  `{{.X}}`, `{{rangeStart "X"}}` → `{{range .X}}`, `{{rangeEnd}}` → `{{end}}`.
- **Prompt self-check and Go validator are two halves.** The prompt's pre-output
  field-count check and the Go `scoreComponent` / `extractSchemaFields` validator
  must change together; updating one without the other breaks Tier D.

## Schema quirks noted during verification

- `orchestration_states` is keyed by `orchestration_id` (uuid); there is no
  `workflow_states` table.
- `agent_error_log.orchestration_id` is `text`, not `uuid` — cast for joins.
- `build_queue` has no `site_id` (keyed by domain + its own id uuid).
- `page_components` has no `page_name` — join to `pages` (`page_components.page_id → pages.id`).
- `pages.build_status` ≠ `pages.status`: `status='active'` means the row exists;
  `build_status='deployed'` means content rendered/pushed; `'planned'` means awaiting build.
- `site_work_items`: `priority` INTEGER (lower = higher priority); `severity` TEXT
  (low/medium/high); `summary`, `created_by`, `pipeline` all NOT NULL; dedup index
  `idx_swi_dedup` partial on (site_id, item_key) WHERE status NOT IN
  (complete, verified, rejected, wont_fix, failed) — use DELETE+INSERT not ON CONFLICT.

## Open follow-ups (not closed by Step 3)

- **`cta_url` silently dropped** when its site_specs path returns nothing and no
  fallback is declared — template renders an empty href. Schema authors should
  declare fallbacks for required URLs.
- **Tier B "soft static" fields.** Fields with `llm_guidance` like "override if site
  tone differs" always use the static fallback today; the author's intent suggests
  the LLM could override when context warrants. Out of scope for Step 3.
- **work_item completion lag.** When page-content-writer FAILs (e.g. Anthropic 529),
  the work item still gets marked complete ("Auto-completed: work verified done
  despite lost response"), leaving pages at `build_status='planned'` despite a
  `complete` work item. This is the `complete-with-error` chassis bug; needs the
  state machine to mark failed-not-complete. (Also the `error` column isn't cleared
  when a retry succeeds — cosmetic but confusing.)
- **Retry budget on LLM overload.** 4 retries on `execute_llm_prompt` don't ride
  through sustained 529s. Backoff/retry tuning is separate.
- **Broaden Tier D** to `game-list`, `blog-listing`, `case-studies-grid` etc. — the
  resolver accepts any `pages_where_type:<X>`, so no resolver code change; each is a
  regen + prompt convergence. Blocked until the above content path was trusted —
  now that Step 3 has landed, this is the natural next expansion. Note the missing
  `game` page_type vocabulary (`FOCUS_component_schema_patterns.md`) blocks
  game-list resolving even after rewrite.
- **The 41 stuck `needs_section_data` items** across the six sites — mostly
  "component not found" cases that should have been `needs_component:*`. Triage is
  separate cleanup, enabled (not done) by this work. `plan_sections` should split
  "component missing" (emit `needs_component`) from "required field unsourceable"
  (genuine HITL) rather than conflating both as `needs_section_data`.
- **v2 (dedicated directory-builder agent)** remains the deferred evolution: a thin
  wrapper around the same `queryresolve.Resolve`, adding re-triggerability and
  parallelism, per the hybrid decision above. v1 (inline resolution in
  plan_sections + merge in page-content-writer) is what's deployed.

## Key identifiers

- Component-creator agent: `23720180-7a39-4e3d-92e1-ebdbf95b57f4`
- system.internal site: `eac60db8-b032-432b-b36d-76f37632045d`
- gamesdesign.co.uk site_id has changed across re-adoptions: `2524997b-...`
  (Session 4), `859d7ad5-0f22-4ba1-8efd-cd59e8fb042f` (Session 5),
  `9a8baddf-77c2-4486-a56b-1b7dde9c1e9e` (Step 2/3 verification). Always resolve via
  `SELECT id FROM sites WHERE domain='gamesdesign.co.uk'`.
- tool-list component: `a68b52b7-61c5-4797-a701-8e8643684f75` (hand-written, migration 041)
- guide-list component: `9d5e461a-8981-4ecc-b236-05895edfc15d` (regenerated, migration 042)
- Chassis images: `8584b77648` (Phase 1), `v1.0.994` (Tier D Go), validator-rejection-logger build (Session 5)
