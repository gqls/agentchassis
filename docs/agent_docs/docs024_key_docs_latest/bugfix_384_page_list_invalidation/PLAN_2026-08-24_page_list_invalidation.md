# PLAN 2026-08-24 — bugs_open/384: page-list consumer invalidation (approved by the owner 2026-08-24 ~19:30Z; copied verbatim from the session plan file)

## Context

**The bug.** A page's listing card (`assets.purpose='card'`, entity-linked to the page) is derived by
`derive_card_asset` (run by `asset-deployer`, step → `complete`). The listing page that shows it holds the
card path inside a **stored array** — `page_components.content_data->'articles'` (or `items`), filled once
from a `query.*` source at section-resolve time. Only a re-render carrying `spec.reason='section_data_resolved'`
(or `image_landed`) re-runs the query; a plain assemble-mode re-render re-ships the stored array verbatim.
**Nothing in the card-landing chain requests the re-resolving mode**, so the listing shows text-only cards
until something unrelated happens to re-resolve it. `derive_card_asset_action.go` emits nothing
(`[MEASURED 2026-08-24]` no `insertWorkItem`/`insertPageRerenderItem`/produce after the upsert).

**Bug still valid, instance repaired.** The dartsonline.com instance was hand-repaired by the filing session at
18:17Z (served `/`: 12/12 cards with `<img>`, re-verified 19:17Z). The mechanism is untouched. Fleet census
`[MEASURED 2026-08-24 19:2xZ]` over every `query.*` array field (43 fields, 25 components): (card asset, stored
entry) pairs — `content-listing`/`blog_posts` 32 pairs, 0 stale (post-repair); `tool-list` 37/0;
`tool-cta` 62 pairs, **14 stale** (5 written *after* the card landed) — invisible only because `tool-cta`'s
template does not render `image`. Demand: **41 card landings / 14 days across 8 sites** (~3/day).

**The class is wider than cards.** `queryresolve.pageImageProjection` (card → plan-hero fallback → "") is
spliced into `resolvePagesWhereType` (bases `pages_where_type:*`, `blog_posts`) and `resolvePagesUnderSection`.
`flag_page_image_rebuild` re-renders only the ARTICLE when a hero lands, so a hero landing leaves listings stale
the same way. Two existing producers already solve this shape for *their* data — `render_news_section_html.go:139-152`
(`insertPageRerenderItem`, key `pageRerenderItemKey(name, site, "section_data_resolved")`) and
`render_directory_action.go:420-441` — each hard-coding its consumer page. **There is no helper anywhere that
answers "which pages consume query source X"** (only a COUNT probe in `check_content_image_missing.go:174-189`).

**Ownership.** Filing session (`dartsonline_traffic`, transcript `fe285621…`) is alive but moved on (last 384 work
18:21Z; since then a Cloudflare question). No commits, no transcript, no dirty file on the fix site. Resuming here.
Corrections to 384's premises found while researching: `bugs_open/083` and `bugs_open/052` are both in
`bugs_closed/` (083 closed 2026-08-22 — the promoter drains handler-bearing `detected` items since 08-15).

## Design — one seam, two event callers, one sweep backstop

**Rule:** *a producer that changes the data behind a page-list query source files a `section_data_resolved`
re-render for every page on the site that consumes one.* Producers name the CAUSE, not the pages.

### New shared pieces

1. **`queryresolve.SourceReadsPageImages(name string) bool`** + `pageImageSources` set
   (`platform/orchestration/actions/queryresolve/queryresolve.go`), declared next to `queryHandlers`:
   `{"pages_where_type", "pages_under_section", "blog_posts"}` (`section_index_for` does NOT splice
   `pageImageJoins` — checked). Same `parseQueryName` normalisation as `Resolve`.
   *Lockstep test* (`queryresolve/page_image_sources_test.go`): drive EVERY `queryHandlers` entry against
   sqlmock with a `QueryMatcherFunc` that records SQL and returns empty rows; assert the set of handlers whose
   recorded SQL contains `ca.purpose = 'card'` equals `pageImageSources` in both directions. Behavioural, not a
   source scan — a new resolver that splices the join without declaring itself fails.

2. **`queryresolve.PageListConsumerPages(ctx, q, siteID) ([]ConsumerPage, error)`** (new file
   `queryresolve/consumers.go`; `q` = the `QueryContext` interface so it takes `*sql.DB` or `*sql.Tx`):
   ```sql
   SELECT DISTINCT p.id, p.name, COALESCE(p.url,''), s.domain, cc.input_schema
     FROM page_components pc
     JOIN pages p  ON p.id = pc.page_id
     JOIN sites s  ON s.id = p.site_id
     JOIN content_components cc ON cc.id = pc.component_id
    WHERE p.site_id = $1
      AND p.status IN ('active','deployed')
      AND pc.build_status <> 'removed'
      AND COALESCE(p.rebuild_policy,'generic') <> 'owned'   -- mirrors ownedPageExclusionSQL; owned pages fail save_sections (OWNED_PAGE_GUARD, 12 FAILED runs/14d)
      AND cc.input_schema::text LIKE '%query.%'
   ```
   then in Go: `datahelpers.SchemaContentFields(schema)` → keep the page if any field's `source` has prefix
   `query.` and `SourceReadsPageImages(strings.TrimPrefix(source,"query."))`. Returns page id/name/url/domain
   + the consuming (component, field, source) triples (the sweep needs them).

3. **`discovery_checks.PageRerenderItemKey(pageName, siteID, keyReason)`** — exported, ONE spelling;
   `actions.pageRerenderItemKey` (`create_rerender_items_action.go:113`) becomes a one-line delegate. Needed so
   the sweep (package `discovery_checks`, which cannot import `actions`) dedups against the event emitter.

4. **`actions.requestPageListReresolve(ctx, exec, siteID, source, cause, batchID, logger) pageListReresolve`**
   (new file `platform/orchestration/actions/page_list_reresolve.go`). For each consumer page:
   `insertPageRerenderItem(ctx, exec, siteID, pageID, source, "low",
   "Re-render <page> — page-list data changed (<cause>)",
   {"reason":"section_data_resolved","page_name","page_id","domain","cause":cause},
   pageRerenderItemKey(name, siteID, "section_data_resolved"), batchID)`.
   Returns `{consumers, queued, deduped, disposition}`; **never fails the caller** (the artefact side effects
   have already happened — same posture as `emitContentCardDerive`), logs at Error on failure and surfaces the
   disposition in the caller's result map. `exec` is `rerenderItemExec` + `QueryContext` so it runs inside
   `flag_page_image_rebuild`'s tx or on `derive_card_asset`'s bare DB.
   No anti-churn brake: `insertPageRerenderItem` is a raw canonical INSERT (targetless `ON CONFLICT DO NOTHING`
   on `idx_swi_dedup`) — consistent with 326's ruling that an action request is not braked; `recurrenceExpected`
   does not apply. Not added to the 326 ratchet list (that list is for `insertWorkItem` sites).

### Callers

5. **`derive_card_asset_action.go`** — after the provenance upsert, `if provenanceRecorded` →
   `requestPageListReresolve(ctx, params.DB, siteID, "derive_card_asset", "card_landed:"+pageName, uuid.New(), logger)`;
   add `listing_reresolve_consumers` / `listing_reresolve_queued` / `listing_reresolve` to the result map.
   Skip when the row was lock-suppressed (no DB change → projection unchanged). Ordering note: the card bytes
   go to the git adapter BEFORE the item is inserted, and page-rerender dispatch is minutes behind, so the
   listing never references a not-yet-committed file in practice; recorded in NOTES, not guarded.

6. **`flag_page_image_rebuild_action.go`** — after `cardEmit := emitContentCardDerive(...)`, inside the same tx:
   `if cardEmit != "raised" { requestPageListReresolve(ctx, tx, siteID, "image-build-handler", "page_image_landed:"+pageName, batchID, logger) }`.
   Rule stated in the comment: *invalidate now unless a card derive was raised — the derive invalidates when it
   lands (caller 5).* Over-invalidation costs one no-LLM re-render per listing, dedup'd by key.
   Its sqlmock tests (`page_section_satisfiability_test.go:392-430`, helper `expectNeedsPageInsertActionRequest`)
   must gain the consumer-lookup query + INSERT expectations — this file was touched today by the 326 lane
   (`e4d20d97a`); clean now; message them before editing (see Coordination).

### Backstop (Phase 2) — sweep sharing the same lookup

7. **`discovery_checks/check_page_list_stale.go`**, name `page_list_stale`, enabled on
   `completeness-discovery-agent` (hosts `contact_form_undeliverable`, `section_source_drift`, `orphan_pages`).
   For each `PageListConsumerPages` page and each consuming field: fresh `queryresolve.Resolve` vs the stored
   array — compare per `url` on `image` (and title/url membership as a cheap extra). Stale → `WorkItemSpec`
   `page_rerender`/`page-rerender`, `Status:"detected"`, spec `{"reason":"section_data_resolved", "check":"page_list_stale", "page_name","page_id", "stale":[…]}`,
   `ItemKey: PageRerenderItemKey(name, site, "section_data_resolved")`. Current → `CheckResult.Resolved`
   (RFC_010 seam) so a previously filed row closes. Precedent shape: `check_contact_form_undeliverable.go:167-194`.
   Gates it must pass: `verifier_coverage_test.go` (`page_rerender` already an acknowledged gap), no
   `fmt.Sprintf` item_type, no hand-rolled INSERT. Enablement = migration `docs/agent_docs/sql_for_agents/NNN_page_list_stale_check.sql`
   appending the name to that agent's `checks` array (unknown names warn-and-skip, so safe in either order vs the roll;
   **it is inert until the binary rolls** — say so in the migration header). Drain: `(page_rerender, page-rerender)`
   pair has 1,323 completes/14d, so the promoter (SCH-026, 900s) dispatches it without a hand canary.

### Out of scope, recorded not fixed
- `rebuild_blog_listing_action.go:212-220` writes `"image": ""` for every listed post (bypasses the projection).
  `[MEASURED 2026-08-24]` latent: 0 of 3 `blog-index` pages list a post that has a card. Filed as a residual in
  384; the Phase-2 sweep would catch it when it fires.
- `tool-cta`'s 14 stale entries — no served defect (template renders no image); the event seam fixes them forward.
- `needs_page`/`page-build-handler` path (the LLM chain) is deliberately not used — `page_rerender` is no-LLM.

## Files

| file | change |
|---|---|
| `platform/orchestration/actions/queryresolve/queryresolve.go` | `pageImageSources`, `SourceReadsPageImages` |
| `platform/orchestration/actions/queryresolve/consumers.go` (new) | `PageListConsumerPages`, `ConsumerPage` |
| `platform/orchestration/actions/queryresolve/page_image_sources_test.go` (new) | lockstep test (sqlmock, records SQL) |
| `platform/orchestration/actions/queryresolve/consumers_test.go` (new) | filter/owned-exclusion/dialect tests |
| `platform/orchestration/actions/discovery_checks/content_image_helpers.go` | `PageRerenderItemKey` (exported) |
| `platform/orchestration/actions/create_rerender_items_action.go` | `pageRerenderItemKey` delegates |
| `platform/orchestration/actions/page_list_reresolve.go` (new) + `_test.go` | the seam; sqlmock: N consumers → N INSERTs; dedup; owned excluded; failure non-fatal |
| `platform/orchestration/actions/derive_card_asset_action.go` | caller 5 |
| `platform/orchestration/actions/flag_page_image_rebuild_action.go` | caller 6 |
| `platform/orchestration/actions/page_section_satisfiability_test.go` | extend expectations for caller 6 |
| Phase 2: `discovery_checks/check_page_list_stale.go` + `_test.go`, `docs/agent_docs/sql_for_agents/NNN_…sql` | backstop |
| `docs/agent_docs/docs026_concept_register/register/rebuild-cascade.md` | **REB-008** "page-list consumer invalidation" (+ index count in `002_PLAN_extraction.md` / ratchet) |
| `bugs_open/384_…md` | status block: what shipped, the class widening, stale pointers corrected, residuals |
| `docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/` (new) | PLAN, RUNBOOK, NOTES, README_where_we_are (created at START) |
| `docs/agent_docs/docs024_key_docs_latest/WRONG_CALLS.md` | any misstep as it happens (so far: none durable — three query-shape slips caught before assertion) |
| `LANDMINES.md` | one entry: *a `query.*`-fed array is a SNAPSHOT — assemble-mode re-renders re-affirm it; a producer that changes page-list data must call `requestPageListReresolve` or its change is invisible until an unrelated section re-resolve* — footprint `queryresolve.pageImageJoins`, `derive_card_asset_action.go`, `assets.entity_id`; then `./scripts/landmines-verify-dispatch.sh` |

Tests to keep green: `go test ./platform/orchestration/actions/... ./platform/orchestration/actions/queryresolve/... ./platform/orchestration/actions/discovery_checks/...`
(`findingcodes_scan_test.go` is known-red from other lanes — record, don't chase). Mutation proof before commit:
delete the call in caller 5 → its test must fail (a mock's bookkeeping cannot assert a negative).

## Sequence

1. **Ownership + coordination first** (all lagging checks re-run at each phase boundary): `git log --since=90min`
   on the bug file + fix-site files; transcript grep for `derive_card_asset|insertPageRerenderItem|PageListConsumerPages`.
   Create the lane dir + standing docs. Append a dated "taken up here" block to `bugs_open/384` (the filer's shared
   account). Message peers via SendMessage: **`bugs_open/326`** (new `page_rerender` emitter, edit to
   `flag_page_image_rebuild_action.go` + its test file), **`bugs_open/357`** (neighbourhood of
   `create_rerender_items_action.go`), **`bugs_open/352`** (checks emitting `page_rerender` are their
   population — Phase 2 adds one), **`bugs_open/333`** (owned-page door: I exclude owned pages at the lookup).
   114's lane and 309's lane have no live session → dated note in their NOTES tails. dartsonline_traffic → bug file.
2. **Phase 1 code** (items 1–6) + tests; `go build ./...`; targeted `go test`; mutation check.
3. **Council submission** (`097_TRIGGER_council_review_v1.sh`, rationale + ≤8 edits + `grounded_in` quoting
   `derive_card_asset_action.go:198-274`, `render_news_section_html.go:130-152`, the census numbers). Register
   REB-008 **in the same commit** (owner ruling 2026-07-28 cond. 2). Commit with explicit pathspec and
   `Council-Submitted: <corr>`; re-run `git log` immediately before committing.
4. **Phase 2** (item 7 + migration) — separate commit, separate council round; migration is in council scope.
5. Do **not** run `make release` (whole-fleet, owner's). Commit is what puts it in the next chassis build.

## Verification (post-roll, at the artefact)

- Prove the roll per SERVICE: `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'`
  → `git merge-base --is-ancestor <my-commit> <stamp>`.
- **Induced case with a demand control.** Pick a site with a known consumer count N (`PageListConsumerPages`
  via SQL, e.g. dartsonline: `index`, `guides-index` → N=2). Dispatch `asset-deployer` in `content_card` mode
  for one listed page (re-derive its existing card; command shape from `imagery/SQL_2026-07-16_asset_deployer_content_card_mode.sql`
  + the kcat lib `scripts/kafka-publish-lib.sh` with its receipt asserted). Expect within the run: **exactly N**
  `site_work_items` rows `item_type='page_rerender'`, `spec->>'cause'='card_landed:<page>'`, key
  `page_rerender_<name>_<site>_section_data_resolved`; then N `page-rerender` orchestrations COMPLETED with
  `rerender_sections.escalated=false`; `pages.deployed_at` advanced on both listing pages **because of** those items
  (join `source_item_id` in `page_component_history`). Disconfirming result: 0 rows, or rows on owned pages, or
  N≠consumers. Served page re-read with `curl` (12/12 stays 12/12 — that check is NOT discriminating on
  dartsonline, the item rows and `deployed_at` causation are).
- **Visible-delta variant** if a candidate exists at roll time: a listed page with a plan hero and no card on a
  site with a rendering consumer (query in RUNBOOK) — the served listing's `<img src>` must switch from
  `hero-…` to `card-…` without anyone touching the listing.
- Hero flavour: an `image-build-handler` run whose `cardEmit != "raised"` must leave one item per consumer with
  `cause='page_image_landed:<page>'`.
- Escalation exposure: count `rerender_sections.escalated=true` among the new items' runs over the first week
  (baseline 1/25 for `section_data_resolved`); record in NOTES with the date.
- Phase 2: run `page_list_stale` on the site with the 14 stale `tool-cta` entries → expect items; after they
  complete, re-run → `Resolved` closes them.
