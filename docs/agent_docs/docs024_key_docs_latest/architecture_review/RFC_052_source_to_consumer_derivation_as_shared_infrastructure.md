# RFC_052 — a schema-derived "who consumes this data source" lookup is now shared infrastructure; should the two hard-coded producers converge onto it, and should the declaration generalise beyond page images?

**Status:** OPEN — raised 2026-08-24 by the `bugfix_384_page_list_invalidation` lane, on the council's advisory objections to round `c2873f56` (APPROVED; `architecture` seat: `needs_rfc`; `reuse_agent` and `bug_historian`: advisory). Not a blocker to the shipped fix; the question is forward fitness.

## What shipped (register PBP-048, commit `9a00a1ee9`)

- `queryresolve.pageImageSources` / `SourceReadsPageImages` — a DECLARATION beside the resolver map naming the `query.*` bases whose items carry an image from the page card/hero join (`pages_where_type`, `pages_under_section`, `blog_posts`), pinned to the resolvers by a behavioural lockstep test.
- `queryresolve.PageListConsumerPages` — the estate's first general derivation of "which pages consume source X", read from `content_components.input_schema` (both dialects), filtered to components whose template renders the image, excluding owned pages.
- `actions.requestPageListReresolve` — files one `section_data_resolved` `page_rerender` per consumer; two callers (`derive_card_asset`, `flag_page_image_rebuild`); `discovery_checks.PageRerenderItemKey` exported so the Phase-2 sweep dedups against it.

## The seats' concern, in their words

- **architecture:** *"the platform's first GENERAL 'who consumes this data source' derivation … as exported symbols across three packages … if Phase 2 and future sources build on pageImageSources's shape without an RFC, the estate gets a second ad-hoc dependency-tracking mechanism layered next to the query.* source system rather than a designed one."*
- **reuse_agent:** *"two hard-coded 'producer notifies its own consumer page' mechanisms plus this new generic one … without proposing to retire or migrate the two precedents onto the new generic lookup."*
- **bug_historian:** *"leaves the underlying invariant ('anything that changes page-image-affecting data must invalidate its consumers') unenforced anywhere except at these two call sites."* (Phase 2 — the sweep — was built the same evening; `603_HOLD`.)

## The question for the architecture track

1. **Generalise the declaration, or keep it image-specific?** Today the declaration is a boolean per base ("reads page images"). The general form is a per-base *dependency set* — `pages`, `assets(card|hero)`, `content_feed_items`, `directory_entities` — so that ANY producer names what it changed and the lookup answers who consumes it. `render_news_section` (795 rows live+archive) and `render_directory` (63) each hard-code one consumer page today; they are the obvious second and third callers, and the ones that go stale when a site grows a second listing page.
2. **Where does the declaration live?** Beside `queryHandlers` (as now — one file, one map, one lockstep test) or as a field on a resolver struct (the map's values are bare funcs today; changing that touches `IsKnownQueryName`/`KnownQueryBases`, CLC-018).
3. **Is a re-resolve the right unit?** The seam files a whole-page section re-resolve (no LLM; escalates to the writer if any section lacks a required `source:"llm"` field — 1/25 baseline). A component-scoped re-resolve does not exist; `create_rerender_items`' scoped mode is per component *template*, not per data field.

## What it costs to leave it

Low near-term (the seats agree): the fix is real, the lockstep test guards drift, and the sweep catches producers nobody wired. The cost is the third and fourth hard-coded consumer name — and the day a site's second news page or second directory page goes stale the way dartsonline's listing did.

## Measured, dated

- Consumer pages per site by schema vs by rendering (2026-08-24): loancalculator 26 → 1, fundamentallyai 6 → 1, gamesdesign 6 → 2, robot-hands 3 → 3; `tool-cta` 58 instances, 0 render `.image`.
- `section_data_resolved` `page_rerender` producers live+archive: 1,289 rows / 53 `created_by`; standing: `render_news_section` 795, `rerender-pages` 203, `completeness-discovery-agent` 2.
