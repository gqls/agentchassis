# FOCUS — Internal Linking (agent-chassis / gamesdesign.co.uk)

**Status:** current as of 2026-06-09. Grounded in `multipage_actions.go`, `site_db_actions.go`, `queryresolve/queryresolve.go`, `plan_sections_action.go`, and `HANDOFF_2026-06-02_hero_resolver_and_section_data_reconciler.md`. Companion to `FOCUS_content_quality.md` — the hero-CTA defect sits on the seam between the two.

---

## What "internal linking" covers
Every link from one built page to another: header/footer navigation, in-body anchor/cross links, list-hub cards (`guides-index` → each guide), "Browse All X" buttons, and the **destination half of hero CTAs**. A CTA's text is content; its destination is a link. The site-wide hero-CTA defect is therefore half this doc, half content-quality.

## The machinery that already exists (reuse — STEP ZERO before building)
1. **Page URLs of record.** `upsertPage` (`site_db_actions.go`) writes each page's `slug`, `url`, `nav_label`, `nav_order`. The `pages` table is the authority for which link targets exist. A valid internal link points at a row here; a phantom link (e.g. `/services.html`) does not.
2. **Navigation.** `SyncPagesToDBAction` builds nav from real pages (`buildNavigationFromPages`) or reads it back via `GetNavigationStructure` (DB, `header`/`footer` types), emitting `nav_data`. `multipage_actions.go`'s `extractNavItemsFromCollectedData` consumes `nav_data`/`db_sync`, then `setActiveNavItems` + `buildNavHTML` render it.
   - **Defect (lines 310–318 of `multipage_actions.go`):** when nav resolution returns empty, `AssembleMultipageSiteAction` injects a **hardcoded fallback nav** — `Home /index.html`, `About /about.html`, `Services /services.html`, `Contact /contact.html`. This generic brochure default is a primary source of the phantom `/services.html` and the stray `/contact.html`.
3. **In-body anchors.** `fixAnchorLinks(html, pageNames)` rewrites `href="#page"` → `/page.html` (index → `/index.html`) for known page names — the single-page-anchor → multipage-link bridge.
4. **List hubs / typed lists.** The `queryresolve` package resolves `query.*`-sourced section fields to page lists: `pages_where_type:<type>` and (added 2026-06-02) `pages_under_section:<area>` (joins `site_areas` via `pages.site_area_id`, with `url_prefix` as a forgiving fallback). This is how a hub like `guides-index` gets its cards and their links. When unresolved at plan time the field is **deferred** as `needs_section_data`; `reconcile_section_data_action.go` (June-02, *not yet wired to a host*) re-triggers the page once the query becomes resolvable; `plan_sections` closes the item on re-render via `closeResolvedDataRequest`.
5. **"Browse All X" buttons.** Sourced from `*_index_url` site_specs — `identity.tool_index_url`/`game_index_url`, but `navigation.*`/`blog.*` for guide/blog (sources are **inconsistent**). **Defect:** these specs are unpopulated → `href=""`.
6. **Link inventory.** `ExtractAndSyncLinksAction` (`site_db_actions.go`) extracts links from rendered HTML (`extractLinksFromHTML`) and syncs them per page into `link_registry` (`syncLinksToDB`). A per-page link inventory already exists — the natural substrate for a broken/phantom-link discovery check.

## Defects → mechanism (CATALOGUE)
- **Hero CTAs wrong site-wide** — both buttons → `/contact.html` & `/services.html`; text↔destination mismatch; `/services.html` is phantom. The destination half is a linking bug; the same hardcoded brochure default that leaks `/services.html` into the fallback nav is the likely family of cause. Whether the hero CTA href is itself a resolvable field or hardcoded in the hero component's `html_template` is the open question below.
- **Empty "Browse All" hrefs** — unpopulated/inconsistent `*_index_url` specs.
- **Phantom targets** — `/services.html` has no `pages` row; `link_registry` should show it as a target with no matching page.

## Resolution model (direction — not yet built)
- Links resolve from **real `pages`/`site_areas`**, never from hardcoded brochure defaults. Remove the 310–318 fallback, or make it derive from actual pages (like `buildNavigationFromPages`) so it can never invent `/services.html`.
- Populate/normalise the `*_index_url` specs from the real hub pages, from one consistent source.
- **Validate every internal link target against `pages`** — a discovery check over `link_registry` (or a hook in `validate_page_content`) that flags any internal href with no matching page row as a broken-link **bug** (→ `content_rewrite` / `component-template-fixer`). This catches `/services.html` and any future phantom generically.
- Hero CTA destinations resolve to the real page that matches the CTA's intent (e.g. "Browse Tools" → `tools-index`), not `/contact.html`.

## Open questions — settle these FIRST in the next chat
1. **Hero CTA: resolvable field or hardcoded template?** Exact analog of June-02 follow-up #2. One row settles it:
   ```sql
   SELECT name, input_schema->'fields' FROM content_components
   WHERE component_level = 'hero' AND is_active = true LIMIT 5;
   ```
   Field with a source → fix the source/resolution. Hardcoded in `html_template` → fix the template and stop emitting the generic CTAs.
2. **Does `link_registry`/`syncLinksToDB` validate targets, or only record them?** Read `syncLinksToDB` + the `link_registry` schema. If it only records, a phantom-link check is purely additive.
3. **Which nav path is live for gamesdesign** — is DB nav (`GetNavigationStructure`) populated, or is the build falling through to the hardcoded default? Check `nav_data` on a build and the rendered header.
4. **`med_url_discovery_action.go` is NOT relevant** — it's a `business_intel` pet-medication price scraper (retailer product URLs via Firecrawl); a keyword false-positive. Ignore for this work.

## Files
- `multipage_actions.go` — `AssembleMultipageSiteAction`, `fixAnchorLinks`, `extractNavItemsFromCollectedData`, `setActiveNavItems`, `buildNavHTML`, the hardcoded fallback nav (310–318).
- `site_db_actions.go` — `upsertPage`, `SyncPagesToDBAction`, `GetNavigationStructure`, `buildNavigationFromPages`, `ExtractAndSyncLinksAction` + `link_registry`.
- `queryresolve/queryresolve.go` — `pages_where_type`, `pages_under_section`.
- `plan_sections_action.go` — `query.*` resolution, `needs_section_data` deferral, `closeResolvedDataRequest`.
- `reconcile_section_data_action.go`, `flag_page_image_rebuild_action.go` — June-02; **registry entries still pending** per that handoff (confirm before relying on them).
- `validate_page_content.go` — where a link-target validation could hook.
