# RFC_052 — a schema-derived "who consumes this data source" lookup is now shared infrastructure; should the two hard-coded producers converge onto it, and should the declaration generalise beyond page images?

**Status:** **CLOSED 2026-08-25 — OWNER RULING: "generalise it now". Built, tested and committed the same day; see §Resolution.** Raised 2026-08-24 by the `bugfix_384_page_list_invalidation` lane, on the council's advisory objections to round `c2873f56` (APPROVED; `architecture` seat: `needs_rfc`; `reuse_agent` and `bug_historian`: advisory). Not a blocker to the shipped fix; the question is forward fitness.

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


---

## Resolution — OWNER RULING 2026-08-25: "generalise it now"

Put to the owner alongside the other three open decisions on `bugs_open/384`. The
answer was to generalise, not to defer. What shipped answers question 1 in full,
answers question 2 by leaving the declaration where it was, and leaves question 3
open (see below).

### ⚠ First, a correction to this RFC's own premise

§What shipped and the question in §1 both say `render_news_section` and
`render_directory` **"each hard-code their one consumer page"**. **They do not, and
they never did.** Both select consumer pages by COMPONENT FUNCTION:

- `queueNewsPageRerenders`: `WHERE cc.function IN ('latest-news', 'news-listing')`
- `queueDirectoryPageRerenders`: `WHERE cc.function IN ($2, $3)` from
  `profile.SnippetComponent` / `profile.ListingComponent`

So both already handled a site with a second news or directory page — the failure
this RFC predicted for them could not have happened in the form described. What
IS hard-coded is the **component set**, which goes stale the day a component is
renamed, or a second component starts consuming the same source. That is a real
defect and a slower one; it is not the one this document argued from. The claim
was checked only when the migration was actually attempted.

**The check that would have caught it: read the function, do not paraphrase its
comment.** `queueNewsPageRerenders`' own doc comment is about routing, not
selection, and the "hard-coded consumer page" reading came from the summary in
PBP-048 rather than from the twelve lines of SQL underneath it.

### 1. Generalise the declaration — YES, done

`pageImageSources map[string]bool` became:

```go
type SourceDependency string
var sourceDependencies map[string]map[SourceDependency][]string
```

per base: **which dependency classes it reads, and which ITEM KEYS each class
feeds**. Five classes today: `page_card_images` (keys `{image}`),
`content_feed_items`, `directory_entities`, `business_intel`, `products`.

**The item-key list is the load-bearing generalisation, and it is what made one
lookup able to serve every producer.** A class that feeds NAMED keys only matters
to a consumer whose template renders one of them — that is the round-2 bound the
council added, preserved exactly. A class with NO key list governs the whole item
SET (its membership and order), so EVERY consumer is affected and no template
filter applies at all. Read backwards, the filter would silently return nothing
for news and directories, whose templates render no `.image`. That asymmetry is
asserted directly by `TestConsumerSQLAppliesTheTemplateFilterOnlyForNamedKeys`.

`business_directory` is its own class: it reads `business_intel`, while the
model/adoption/protocol/lender trackers read `directory_entities`. Merging them
on the strength of the word "directory" would notify the wrong consumers on
every publish.

### 2. Where the declaration lives — UNCHANGED, beside `queryHandlers`

Not moved onto a resolver struct. The map's values are still bare funcs, so
`IsKnownQueryName`/`KnownQueryBases` (CLC-018) are untouched. What changed is that
the lockstep test now covers every class, not just cards, and gained a
COMPLETENESS half: `TestEveryRegisteredBaseDeclaresItsDependencies` requires every
registered base to appear in the declaration. A base with nothing to declare says
so with an **empty map** (`section_index_for` does), so "nobody thought about it"
and "there is nothing to think about" stay distinguishable — which the old boolean
map could not express at all.

⚠ **A gated resolver defeats the lockstep harness.** `resolveBusinessDirectory`
looks up its exporter config first and returns an error when there is none
(deliberately — bugs_open/206), so under a mock that answers every query empty it
never issues its `business_intel` query and records as "reads nothing". The
generalised test failed on exactly that on its first run and reported a CORRECT
declaration as stale. The fix is to feed the gate, not to add an exclusion list:
an exclusion would hide the one thing the test exists to catch.

### 3. Is a re-resolve the right unit — STILL OPEN

Unchanged by this work. The seam still files a whole-page section re-resolve, and
a component-scoped re-resolve still does not exist. Nothing here made that worse
or better; it is the remaining question if anyone reopens this track.

### The producers migrated — SELECTION only, never the route

Both now call `queryresolve.ConsumerPages`. **Routes and item shapes are
deliberately unchanged**: news files `page_rerender` via `insertPageRerenderItem`;
directory files `needs_page` → `page-build-handler` with its own
`page_rerender:<page>` key. Changing a route is what made the news emitter's first
version LLM-regenerate live pages four times a day for four days, and this
migration had no reason to touch it.

`render_directory` publishes ONE KIND out of a shared store, so it narrows the
shared answer with `ConsumerPage.ConsumesAny(profile.SnippetSource,
profile.ListingSource)` rather than the shared lookup learning a per-caller special
case. Its profiles gained the two `query.*` base names, verified 1:1 against the
live schema.

The shared lookup also **gained `datahelpers.PageHasShippedPredicateFor`**, which
both producers already carried. Without it the lookup would have been the weaker of
the two and a migration meant to remove drift would have introduced some.

### Measured before migrating — all three are no-ops on today's fleet

| comparison | both | producer-only | schema-only |
|---|---|---|---|
| news consumer pages (fleet) | **16** | 0 | 0 |
| directory consumer pages (fleet) | **5** | 0 | 0 |
| directory per kind (model / company / protocol / mortgage-lender) | 1 / 1 / 1 / 2 | 0 | 0 |
| has-shipped floor on the page-image lookup | 62 → **62** | — | 0 excluded |

`[MEASURED 2026-08-25]`. That the sets are identical is what makes this reviewable
rather than a leap: the migration cannot change behaviour today, only what happens
the next time a component is renamed.

⚠ **One of these numbers was wrong on the first run.** The news comparison first
reported "2 schema-only" pages. That was an operator-precedence bug in my own SQL
(`AND … LIKE … OR … LIKE …` — the `OR` escapes the `AND` chain). Parenthesised, it
is 0. Recorded in `WRONG_CALLS.md`; the cheap check is to parenthesise every mixed
AND/OR before reading the result.

### What is NOT claimed

The generalisation is **built, tested and committed**; it is **not yet rolled**.
Go changes are inert until a chassis image is built and rolled, so nothing in this
section is live behaviour until then — verify at the binary
(`service_binary_capabilities`, or the `build provenance` line plus
`git merge-base --is-ancestor`) before quoting it as live. The migrations that ship
with the same lane (`614`/`615`) ARE live, because DB config is.
