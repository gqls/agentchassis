# HANDOFF 2026-06-02 — hero resolver and section-data reconciler

## Status

Two related rendering gaps were diagnosed end to end and the fixes are in production: per-page hero/logo images that the pipeline produced but pages didn't render, and `needs_section_data` items that sat at `needs_human_review` even when the data they needed was mechanically resolvable. The build error that surfaced mid-deploy (`check_tool_recreation_needed.go` PageID pointer mismatch) is also fixed.

Deployed:

- `plan_sections_action.go` — page-aware `ensureAssets`
- `flag_page_image_rebuild_action.go` — new terminal step in the image-build-handler workflow
- `reconcile_section_data_action.go` — new, **not yet wired to a host** (see open items)
- `queryresolve.go` — `pages_under_section` query added
- `check_tool_recreation_needed.go` — `PageID *uuid.UUID` pointer fix
- image-build-handler agent definition — `flag_rebuild` step wired in, verified one row updated
- Three doc updates applied: `016_debugging_guide_v2_26.md`, `FOCUS_imagery_assessment_1_.md` §5.1, `FUTURE_section_data_handler_1_.md`

Two pieces are still required to close the section-data loop end to end: a host for the reconciler (a loop check or a post-build finalize step), and the registry entries for both new local actions.

---

## The two problems and their fixes

### 1. Per-page hero/logo images deployed but pages render the fallback

**Symptom** (observed on gamesdesign.co.uk after a clean adoption build, site `a79c77ec-208e-4436-ab44-2b54a1ab442c`). All four `needs_imagery` items completed, the assets table contained `hero-home`, `hero-games`, `hero-about`, `hero-tools`, `logo`, and section icons keyed correctly, the deploy committed them to git — and every page rendered `background-image: linear-gradient(...), url('/assets/images/hero.jpg')`, a file that was never produced. Identical static URL on every page. The header showed a text mark, not `logo.jpg`.

**Diagnosis.** The asset layer and the section-render layer disagreed. Three things stacked:

1. `plan_sections`' `sourceResolver.ensureAssets` mapped a single site-wide `content_data["hero_url"]` to `assets["hero"]`. `store_asset` writes `content_data["<purpose>_url"]` keyed by purpose, and every page hero has purpose `hero`, so they overwrote each other last-write-wins. Even when resolved, every page got one hero.
2. `needs_imagery` runs asynchronously after the first render (heroes completed 21:00–22:03; the pages had built around 16:00). At first render `site_assets.hero` was unresolved, the field's `on_missing: use_fallback` fired, and `/assets/images/hero.jpg` baked into `page_components.rendered_html`.
3. The 22:05 terminal `needs_rerender` reassembled stored HTML (regex `UPDATE page_components SET rendered_html` in the colour/CSS fixers) — it did not re-run `plan_sections`, so the baked fallback survived even though the assets existed by then.

The trace queries are in `016_debugging_guide_v2_26.md` §9 under *Deployed hero/logo images exist in git but the page renders the fallback*.

**Fix.** `ensureAssets` now resolves the page's hero from the current plan's imagery joined to the deployed asset (`site_plan_imagery.key = assets.asset_key`, `assets.url` is the web path), scoped to this page (`scope='page' AND scope_ref=pageName AND kind='hero'`), and the site logo from the site-scope row. `content_data` stays as gap-fill for legacy/adopted sites. Coupled with `flag_page_image_rebuild`: a terminal step in image-build-handler that, for page-scoped imagery, flags the page `needs_rebuild` (reusing `flagPagesForRebuild`) and emits `needs_page` → `page-build-handler` at priority 99 so the page re-resolves *through* `plan_sections` after its asset lands. The shared dedup key `page_rerender:<page>` collapses concurrent triggers (one re-render no matter how many of a page's images complete).

The logo is deliberately out of scope of that fix — the header is a site component rendered by `render_site_components`, not by `plan_sections`, so a page rebuild won't touch it.

### 2. `needs_section_data` items sitting at `needs_human_review` for query-resolvable data

**Symptom.** Two `needs_section_data` items at `needs_human_review`, one each for the guide-list on `index` and `guides-index` — duplicating the same deferred field.

**Diagnosis.** `plan_sections` resolves `query.*`-sourced fields via the `queryresolve` package. The dispatch switch implemented only `pages_where_type:<type>`; the vocabulary comment in the same file named `pages_under_section:<section>` but it wasn't in the switch — unknown query → fell through → defer. `loadOpenSectionDataRequests` + `closeResolvedDataRequest` only close an item on a *later* `plan_sections` run, and nothing re-plans those pages after the guides build. So the items sit indefinitely. There is no `needs_section_data` handler agent, and the original `directory-builder` agent (per `FUTURE_section_data_handler_1_.md`) was never built.

**Decision** (recorded in `FUTURE_section_data_handler_1_.md`, dated update 2026-05-27). A full LLM handler agent is not needed for the query-resolvable cases. Two pieces:

- Implement `pages_under_section:<area>` in `queryresolve` (a near-clone of `resolvePagesWhereType` that joins `site_areas` via `pages.site_area_id`, matching `<area>` against `site_areas.name` with `url_prefix` as a forgiving fallback). Done.
- Add a lightweight reconciler — a resolver, not an LLM agent — that scans open `needs_section_data` items whose missing fields are all `query.*`-sourced, re-runs the query, and emits `needs_page` for pages where the data now exists (shared dedup key `page_rerender:<page>`). `plan_sections` closes the open item on re-render via `closeResolvedDataRequest`. Items with any non-query (human) missing field — team, pricing, case studies — stay HITL untouched. Reconciler distinguishes by `spec.missing[].source`.

The reconciler is built but not yet wired to a host.

---

## Files

### Go — in `platform/orchestration/actions/`

- `plan_sections_action.go` — `sourceResolver` gained a `pageName` field; `newSourceResolver` takes `pageName` (one caller, threaded from the existing `pageName` in `PlanSectionsAction`); `ensureAssets` rewritten to resolve per-page hero + site logo from `site_plan_imagery` JOIN `assets`, with `content_data` as gap-fill. Empty `pageName` degrades to the fallback.
- `flag_page_image_rebuild_action.go` — new. Reads `scope`/`scope_ref` from the work item spec; for `scope='page'` flags the page `needs_rebuild` and emits `needs_page` → `page-build-handler` at priority 99 with item_key `page_rerender:<page>`. No-ops for non-page-scoped (legacy `needs_logo`/`needs_hero_image` items pass through harmlessly).
- `reconcile_section_data_action.go` — new. Inputs `site_id` (required) + optional `page_name`. Scans open `needs_section_data`, parses `spec.missing[]`, re-runs `query.*` sources via `queryresolve.Resolve`, emits `needs_page` for pages where every missing field is `query.*` and every query now returns non-empty. Does not close items itself — re-rendering through `plan_sections` is the single source of truth.
- `discovery_checks/check_tool_recreation_needed.go` — `PageID t.ID` → `pid := t.ID; PageID: &pid`. The original took the address of a `range`-loop variable, which is unsafe pre-Go 1.22; the per-iteration local is version-independent.

### Go — in `platform/orchestration/actions/queryresolve/`

- `queryresolve.go` — added `case "pages_under_section"` to the `Resolve` switch and the `resolvePagesUnderSection` function (joins `site_areas`, filters `status IN ('active','deployed')`, returns the same standard list-item shape as `pages_where_type`).

### Workflow — `agent_definitions` row `04b10d94-11ee-447c-9ff9-7924b8e9897c` (`image-build-handler`)

- New step `flag_rebuild` with `action: flag_page_image_rebuild`, config paths `site_record.site_id` / `input_data.spec.scope` / `input_data.spec.scope_ref`, `next_step: complete`, `error_step: complete` (a flag failure must not fail the asset workflow — the asset is already deployed).
- `mark_work_item_complete.next_step` redirected from `complete` → `flag_rebuild`. Every path (`needs_imagery`, logo, hero, variant) converges at `mark_work_item_complete`, so all flow through.

`UPDATE 1` returned; verify showed `mwc_next = flag_rebuild`, `fr_action = flag_page_image_rebuild`, `fr_next = complete`, `fr_scope_ref = input_data.spec.scope_ref`.

### Docs

- `016_debugging_guide_v2_26.md` §9 — new entry *Deployed hero/logo images exist in git but the page renders the fallback*, with the three diagnostic SQL queries. Existing `needs_section_data → wont_fix` entry augmented to distinguish query-resolvable list data from human-only data and to point at the reconciler direction.
- `FOCUS_imagery_assessment_1_.md` §5.1 — gap and decision, with the two open items it leaves (logo path, second-resolution-path reconciliation, field-vs-template verification).
- `FUTURE_section_data_handler_1_.md` — dated update 2026-05-27 added after the 2026-05-06 supersession block (refines, doesn't contradict): `queryresolve` state today, why the guide-list duplication happens, the reconciler decision, what stays HITL.

---

## Registry entries still required

Both new local actions need their registry entry in `registry.go` before they're callable from a workflow:

```go
"flag_page_image_rebuild": {
    Handler:     FlagPageImageRebuildAction,
    Category:    "site",
    Description: "Re-render a page after its image asset lands so the hero resolves",
    IsLocal:     true,
},
"reconcile_section_data": {
    Handler:     ReconcileSectionDataAction,
    Category:    "site",
    Description: "Re-trigger pages whose deferred section data is now query-resolvable",
    IsLocal:     true,
},
```

`flag_page_image_rebuild` is referenced by the live image-build-handler workflow — it must be registered before the next image-build runs, or the workflow will error at the `flag_rebuild` step.

---

## Open follow-ups

1. **Host for the section-data reconciler.** It is invocation-agnostic — given `site_id` it scans and re-triggers. Plan-time is too early (the listed pages don't exist yet). Two reasonable hosts: a periodic loop discovery check, or a post-build finalize step in the rerender path. Pick one and add the step.
2. **Verify the hero is a `site_assets.hero` field, not a hardcoded template path.** The resolver fix assumes the hero component declares its background as a field with source `site_assets.hero` and `/assets/images/hero.jpg` as its `use_fallback` — the inference from Part A of the diagnosis. If the path is hardcoded in the component's `html_template` instead, the resolver has nothing to fill and the fix moves to the template. One row from `content_components` will confirm or refute this:
   ```sql
   SELECT name, input_schema->'fields' FROM content_components
   WHERE component_level = 'hero' AND is_active = true LIMIT 5;
   ```
3. **Logo/header path.** The header is rendered by `render_site_components`, separate from `plan_sections`. The logo URL ultimately needs the same kind of resolution there (or the header component needs swapping for a logo-image variant). Not done.
4. **Reconcile the two image-resolution paths.** `FOCUS_imagery_assessment` §5 already documents that `BuildRenderContextAction` recognises per-variant keys (`hero_home_url`, `hero_about_url`, …) but only `hero_home` is generated. The new `plan_sections` ensureAssets resolves through a different mechanism. Confirm which path the deployed hero actually comes through; reconcile or document the relationship between the two.
5. **Sweep other `discovery_checks/check_*.go` for the same `PageID` pattern.** The mismatch fixed in `check_tool_recreation_needed.go` would surface anywhere else a `WorkItemSpec.PageID *uuid.UUID` is assigned a value-type UUID directly:
   ```bash
   grep -n "PageID:" platform/orchestration/actions/discovery_checks/*.go
   ```

---

## Verification — first build to watch

After the next clean adoption or build-pipeline run, four queries confirm the fixes are doing what they should.

```sql
-- A. Per-page heroes are deployed and active (asset layer correct)
SELECT spi.scope_ref, spi.kind, spi.key, a.url, a.status
FROM site_plan_imagery spi
JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
LEFT JOIN assets a ON a.site_id = sp.site_id
                  AND a.asset_key = spi.key AND a.status = 'active'
WHERE sp.site_id = '<site>'
ORDER BY spi.scope, spi.scope_ref, spi.kind;

-- B. Stored page HTML now references the per-page hero, not /assets/images/hero.jpg
SELECT p.name, substring(pc.rendered_html from 'url\(([^)]*)\)') AS bg_ref
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE pc.site_id = '<site>' AND pc.rendered_html LIKE '%hero%';
-- Expect: hero-home.jpg / hero-games.jpg / hero-about.jpg / hero-tools.jpg
-- per page; not a uniform /assets/images/hero.jpg.

-- C. page_rerender re-render items emitted and consumed
SELECT item_key, status, priority, created_by, created_at, completed_at
FROM site_work_items
WHERE site_id = '<site>'
  AND item_key LIKE 'page_rerender:%'
ORDER BY created_at;
-- Expect: one per page that has imagery; status transitions to complete;
-- created_by either 'image-build-handler' or 'section-data-reconciler'.

-- D. needs_section_data items: closed for query-resolvable cases,
--    still open at needs_human_review only for genuinely-human data
SELECT item_key, status,
       jsonb_path_query_array(spec, '$.missing[*].source') AS missing_sources
FROM site_work_items
WHERE site_id = '<site>' AND item_type = 'needs_section_data'
ORDER BY status, created_at;
-- Expect: items with all-query sources end up complete (closed by
-- closeResolvedDataRequest on re-plan). Items with non-query sources stay
-- at needs_human_review.
```

If query B still shows `/assets/images/hero.jpg` on a freshly built page, the resolver isn't being hit on render — re-check the hero component's `input_schema` per follow-up 2.

---

## Context pointers

Previous handoff in the same thread: `HANDOFF_2026-05-25_part_a_deployed_pending_clean_test.md`. The build-side design trigger, build-time imagery emitter (`emit_imagery_items` + the `imageryplan` shared package), and site-status flip (`mark_site_deployed`) were verified live on gamesdesign earlier in this thread; that state has not changed. The hero/section-data work in this handoff sits on top of those.

Relevant existing files in the repo that the new code reuses:
- `flagPagesForRebuild` (line ~40974 in `production_agent-chassis-actions-current_context.txt`) — reused unchanged by both new actions.
- `insertWorkItem` + `workItem` struct (top of the actions package) — reused for the two `needs_page` emits.
- `queryresolve.Resolve` + `QueryRequest` — reused by the reconciler.
- `closeResolvedDataRequest`, `loadOpenSectionDataRequests`, `createDeferredItems` (in `plan_sections_action.go`) — left untouched; the reconciler relies on `closeResolvedDataRequest` running inside `plan_sections` on re-plan.
- `ExtractActionInputs` Strategy 0 — both new actions use it; config dot-paths resolve from collected data automatically.

The original transcript covering this work (Part A/B diagnosis and the build of the resolver, reconciler, and queryresolve addition) is in this session's history.
