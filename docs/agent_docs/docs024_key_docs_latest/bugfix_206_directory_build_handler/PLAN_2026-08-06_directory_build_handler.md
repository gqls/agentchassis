# PLAN 2026-08-06 — implementing `bugs_open/206`: a real `directory-build-handler`

**Ask:** owner asked to build `vetcomparison.uk`'s practice-directory page through the
framework, and to fix whatever the framework is missing rather than hand-build. `bugs_open/206`
diagnosed the gap: no builder ever turns a plan-scaffolded `entity-directory` page into a real,
data-backed page. This is that fix.

## What's already proven, reused rather than rebuilt

- `directory_export_action.go` / `directory-export-json` — already exports vetcomparison's 2,337
  veterinary businesses to the site's repo every 48h. Untouched by this work.
- `queryresolve` package — the generic `query.*` resolver already used by `guide-list`/`tool-list`
  (both PROVEN live: `mortgagecalculator.co.uk`, `idea.uk`, `gamesdesign.co.uk`, `relojistas.com`
  all deploy `["hero","guide-list"]` guides-index pages backed by
  `query.pages_where_type:guide`). This confirms the query-resolution mechanism itself is sound;
  only a business-directory-shaped query is missing.
- `load_page_sections_from_spec_action.go` — the page-build pipeline's own priority order for a
  page's section layout (site_plan_sections → site_specs.site_plan → pages.sections →
  same-role-sibling synthesis). Confirms `directory-index` gets nothing from any of the four
  because it's the only `entity-directory` page on this site and nothing ever wrote it a plan row.
- `apply_gap_plan_action.go`'s `defaultSectionsForPage` — the existing type/name-keyed default
  layout chooser. Extending this in place (not duplicating it) keeps one source of truth for
  "what's the sensible default layout for a page like this."
- `page-build-handler` — the existing generic content-writer/deploy pipeline. The new handler
  agent DELEGATES to it rather than re-implementing content generation.

## What's genuinely new

1. **`queryresolve.resolveBusinessDirectory`** (new file `business_directory.go`) — given a
   site_id, looks up that site's OWN `directory-export-json`-shaped `scheduled_tasks` config (by
   domain match) to find its vertical + optional business_type filter — the SAME config
   `directory_export_action.go` reads — then runs the identical filter `loadDirectoryEntries`
   uses (verified, website_url + postcode present, matching vertical), capped for build-time
   rendering. Registered as `query.business_directory` (no arg — vertical comes from the site's
   own export config, so the component schema stays generic across any vertical, not hardcoded to
   veterinary).
   - **Why look up scheduled_tasks rather than take vertical as a static arg** (contrast
     `query.pages_where_type:tool`, which DOES take a static arg): `pages_where_type` args are
     genuinely site-agnostic vocabulary (every site's tool pages are `page_type='tool'`).
     Vertical is not — hardcoding `query.business_directory:veterinary` into the SHARED
     `directory-listing` component schema would make it unusable for a future non-veterinary
     comparison site. Deriving it from the site's own already-live export config means the
     component works for any site once IT has an export config, and the rendered listing and the
     exported JSON can never name a different vertical by accident.
2. **`ensure_page_section_layout` action** (new file) — given `site_id` + `page_name`: refuses
   (no-op) if the page already has ANY plan sections from `site_plan_sections` (current plan) or
   a non-empty `pages.sections` — this is the guard that makes it safe under "never re-plan this
   site" (`bugs_closed/001`): it can only ever fill a genuinely empty page, never touch an
   existing plan. Otherwise resolves the default layout via `defaultSectionsForPage` and inserts
   `site_plan_sections` rows for that one page under the site's current plan — the exact INSERT
   shape `write_site_plan_action.go` already uses (`plan_id, page_name, ordering, component_name`,
   resolved IDs left NULL for the downstream resolver), so this action produces a row
   indistinguishable from one the real planner wrote.
3. **`defaultSectionsForPage` extended** (in place, `apply_gap_plan_action.go`) — two new cases:
   - `page_type == "entity-directory"` → `["hero", "directory-listing"]`.
   - name-matches `guides-index`/`guide-index` → `["hero", "guide-list"]`; `tools-index`/
     `tool-index` → `["hero", "tool-list"]` — mirrors the ALREADY-PROVEN fleet-wide pattern
     exactly (queried live: 4 other sites use this exact layout for this exact page shape). This
     also fixes `vetcomparison.uk/guides-index`, which needed no new capability at all — its gap
     was identical (empty plan), and the fix is the same action, same guard, zero new resolver
     code.
4. **`entity-directory` moved from `unavailableBuilders` to `availableBuilders`**
   (`load_work_item_actions.go`) — uncomments the line the original authors already reserved for
   this, pointing at a new `directory-build-handler` agent type / `needs_directory` item type.
5. **`directory-build-handler` agent** (DB seed, not Go) — workflow `ensure_layout` (the new
   action) → `call_page_build_handler` (delegates to the existing generic builder) → `complete`.
   No content-writing logic of its own.
6. **`directory-listing` component schema + template fixed** (DB seed) — its `entries.source`
   currently points at `query.directory_entries`, which is not a registered query name in
   `queryresolve.go` (checked: the actual registered names are `model_directory`/
   `model_directory_full`/`adoption_tracker`/`protocol_tracker`/etc. — this component's schema was
   never actually wired to a live query, consistent with its 0 live callers). Fields also don't
   match business data (`region`/`category_slug` vs. what business_intel actually has). Rewritten
   to `source: "query.business_directory"` with fields `name, postcode, location, website,
   is_claimed` — exactly `directoryEntry`'s own shape, so the rendered listing and the exported
   JSON can never disagree about a business's facts.

## Deliberately NOT in this pass

- **`entity-page` stays in `unavailableBuilders`.** `practice` (the per-business page type) is
  correctly on HOLD per the 07-26 review — P1's company-number crawl is 10/~2,109 done, so a
  practice page would carry almost nothing today. Building its handler now would be building
  ahead of a capability nothing can use yet.
- **No client-side search/pagination across the full 2,337 businesses.** The rendered
  `directory-listing` section is capped (matching every other listing component's own cap
  discipline) for build-time SSR. Full browse/search of the complete set already exists via the
  client-side-fetchable exported JSON (`vet-full-index.json`) — building a search UI against that
  file is a separate, clearly-scoped follow-on, not silently bundled into this fix.
- **`tool-compliance-deadline-calculator`** — separate mechanism (tool pipeline), separate legal
  constraints (no calendar dates). Not touched.

## Platform-seams compliance

This is new, shared, opt-in capability (`query.business_directory`, `directory-build-handler`) —
reachable by nothing until a page's plan names it. Per OWNER RULING 2026-07-29 §1, this is normal
council-gate scope, not an RFC (it changes no existing guarantee). Registering in the concept
register in the same commit that ships it, and submitting to the council gate before/alongside
committing, per the platform-seams ruling.

## Verification plan

- Unit tests for `resolveBusinessDirectory` (mock DB: no export config → empty+no error; export
  config present → correct filter applied) and `ensure_page_section_layout` (guard refuses on
  existing sections; applies default layout when genuinely empty; never touches another page).
- `go build`/`go test` clean against `git archive HEAD` + the change.
- Council submission before commit.
- After roll: pod-grep the new symbols on the running pod (never trust the tag).
- Re-triage `bugs_open/206`'s two named `needs_page` rows (`715ec305` directory-index,
  `2f50bfda` guides-index) to their new/existing handler and `status='triaged'`, then let
  `build-pipeline-trigger` (already live, 120s cadence) pick them up — NOT a direct kcat dispatch
  this time, since the whole point is these are ORDINARY work items now, indistinguishable from
  any other page build.
- Verify the DEPLOYED pages directly (curl, not status) — real business names/postcodes on
  directory-index, real guide titles on guides-index.
