# Register — navigation

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

13 concepts (NAV-013 added 2026-07-31, post-freeze), consolidated from 28 raw extractions (14 unique blocks, each mechanically duplicated once in the cluster input file — see note in styling-render-pipeline.md) across units U01, U02, U17a, U18, U19, U21, U25.

### NAV-001 — Nav agent family and the three-tier authority model
- **status:** partial
- **status-evidence:** "owner currently populate_nav_tables action within pageflow-builder; tiers described as model" — independently corroborated by a second unit: "core responsibilities are implemented as the populate_nav_tables action... full standalone nav-agent is planned but not yet needed."
- **what:** Navigation is treated as a first-class entity: `site_nav_groups`/`site_nav_items` with typed groups (primary, subsection, content, legal, utility, external, contextual planned). A three-tier authority model governs it: Tier 1, the strategist, owns structure at new-build time (the only tier fully implemented — via `populate_nav_tables`); Tier 2, an autonomous nav agent, would make incremental maintenance decisions (today's `nav-updater`/`nav-link-fixer` cover drift and broken template links but are not the full standalone agent envisioned); Tier 3, periodic drift detection, compares current nav against the original plan ("drift may represent valid evolution") and is not built. A nav dedup guard was recommended after a duplicate-nav-items incident (B-029-1).
- **sources:** 002(4)#Navigation Agent Family; 024; 029 B-029-1; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#1-Navigation-Agent-Family and #Three-Tier-Authority-Model
- **relations:** nav-updater never spawns; populate_nav_tables; NAV-002 (two nav systems); NAV-009 (nav-updater/nav-link-fixer, the actual Tier-2-ish agents that exist)
- **verify-later:** nav drift check + dedup guard status; site_nav_groups/site_nav_items tables; standalone nav-agent existence

### NAV-002 — Two nav systems and the GetNavItems fallback
- **status:** deployed
- **status-evidence:** "Two Nav Systems (and they conflict)" — nav tables intended, pages flags legacy; partial population yields a mix (compiled ~2026-04/05).
- **what:** `site_nav_groups`/`site_nav_items` (populated by `populate_nav_tables`, read by `GetNavItems`) coexist with legacy `pages.in_header`/`in_footer` flags (`GetHeaderNavFromPages` fallback). `GetNavItems` tries tables first, falls back to pages — partial population of the tables mixes the two systems in practice. Nav state is captured in snapshots and restorable via revert.
- **sources:** FOCUS_navigation.md#1, #2, #7
- **relations:** NAV-003 (stale pages problem); nav discovery checks (NAV-004); site-design-planner navigation spec
- **verify-later:** GetNavItems fallback logic; whether Tier 2/3 of NAV-001 exist

### NAV-003 — Stale pages from previous builds polluting nav + config-driven deactivation fix
- **status:** deployed
- **status-evidence:** Initially "SyncPagesToDBAction uses ON CONFLICT (site_id, name) — it only overwrites matching page names" with fixes listed as "needed" and still item 15 on the errors-to-fix list; a later, more specific and dated read states the fix as shipped: nav getters add `AND build_status = 'deployed'` and a new-build flow deactivates pages absent from the current plan, gated by a `deactivate_stale_pages` config flag.
- **what:** Pages from prior builds kept `in_header=true`/`status=active` and appeared in nav even though absent from the current plan, because `SyncPagesToDBAction`'s `ON CONFLICT (site_id, name)` only overwrites matching page names and nav queries didn't filter `build_status`. Fix: nav getters (`GetHeaderNavFromPages`/`GetFooterNavFromPages`) add a `build_status = 'deployed'` filter, and new-build flows deactivate stale pages under a `deactivate_stale_pages: true` flag — while maintenance/adopt flows preserve them, respecting adoption faithfulness.
- **sources:** FOCUS_navigation.md#stale-pages; FOCUS_navigation_errors_to_be_fixed.md#15; WM/001_development_guide(0).md#nav-sync-config-driven-page-deactivation; ED/102_blog_handoff-2026-04-10.md#a-check_orphan_pagesgo-new-routing-logic
- **relations:** NAV-002 (two nav systems); adoption faithfulness (preserve semantics); site plan reconciler nav auditor; LNK (link management)
- **verify-later:** SyncPagesToDBAction current behaviour; GetHeaderNavFromPages/GetFooterNavFromPages; site_nav_items

### NAV-004 — Nav discovery checks and fix agents
- **status:** deployed
- **status-evidence:** check/handler tables in FOCUS_navigation (broken_nav_links→nav-link-fixer; checkNavLayout/checkUnwantedElements→component-template-fixer; checkUnlinkedSiteComponents→site-component-linker; orphan_pages→rerender-pages/content-gap-planner).
- **what:** The nav slice of the improvement loop: quality/design/completeness discovery agents detect anchor-slug links, stacked nav (missing flex), unwanted search icons, unlinked header/footer components, orphan pages, and missing logo images; dedicated fixers repair templates, relink components (clearing rendered_html + needs_rerender), and make orphans reachable. `component-template-fixer`'s idempotency was case-sensitive, injecting responsive CSS 4× (fix: lowercase compare).
- **sources:** FOCUS_navigation.md#3, #4; FOCUS_navigation_HANDOFF_navigation_fix.md#problems-10
- **relations:** NAV-007 (fallback header); NAV-005 (duplicate header/footer)
- **verify-later:** discovery agent checks arrays; fixInjectResponsiveCSS case fix

### NAV-005 — Duplicate header/footer pathology (site-level components in pages.sections)
- **status:** partial
- **status-evidence:** Data fixes applied 2026-04-11 (12 pages.sections rows cleaned, 24 page_components deleted); but 10 dirty rows reappeared by 04-13/14 — "plan_sections filter NOT deployed" (2026-04-20 investigation).
- **what:** `pages.sections` listed site-level component names alongside content sections; rebuilds rendered header/footer as page_components, then InjectHeader/InjectFooter added a second copy. Code fixes designed but pending at doc date: `filterSiteLevelSections` in PlanSectionsAction (prevents recurrence), skip-if-present guards in InjectHeader/InjectFooter. A discovery check for duplicate headers inside `<main>` was also missing.
- **sources:** FOCUS_navigation_HANDOFF_navigation_fix.md (whole); HANDOFF_2026-04-20_error_investigations.md#7; FOCUS_navigation_errors_to_be_fixed.md#1-2
- **relations:** NAV-004 (nav fix agents); page-build-handler
- **verify-later:** plan_sections_action.go for filterSiteLevelSections; component_library.go inject guards

### NAV-006 — Nav quality mechanisms of 2026-04-17 (tiers, child-page exclusion, label trust, quick links)
- **status:** deployed
- **status-evidence:** "What Was Deployed This Session" (2026-04-17): tiered priority, isChildPageURL, navLabelForPage, quick_links_html + footer template SQL.
- **what:** `populate_nav_tables` gained a three-tier page priority (core / hubs+conversion / secondary, overflow to utility) replacing arbitrary nav_order truncation; child-page URL prefixes (/tools/, /blog/ …) excluded from all nav groups; nav labels trust `page.NavLabel` up to 30 chars without truncating to two words; footer Quick Links built from primary+utility groups via a new `quick_links_html` variable.
- **⚠ NARROWED 2026-07-31 — "excluded from all nav groups" is NO LONGER TRUE, and that exclusion turned out to be a defect (`bugs_open/149` A2, chassis v1.0.1215).** A child-path page that has DECLARED nav membership (`in_header`/`in_footer`) now lands in `utility`; only a child page declaring nothing is omitted. The blanket exclusion ran **before** the flags were read, so it silently swallowed the declaration — which made `check_orphan_pages`' `nav_drift` branch unfixable by its own handler (`nav-updater` completed having placed nothing, twice measured, two days apart). The tiering, the label trust and the quick links in this entry are unchanged. See **NAV-013** for the rule that replaced it.
- **sources:** HANDOFF_2026-04-17_nav_empty_sections_footer(1).md#2-5
- **relations:** NAV-002 (two nav systems); NAV-008 (tool nav integration)
- **verify-later:** populate_nav_tables_action.go navPriorityTier/isChildPageURL

### NAV-007 — Hardcoded fallback nav/header defaults inventing structure
- **status:** partial
- **status-evidence:** Defect recorded at "lines 310–318 of multipage_actions.go: … injects a hardcoded fallback nav — Home/About/Services/Contact" (2026-06-09), alongside RenderFallbackHeader's stacked-nav/search-icon behaviour; a later, more specific investigation into a concrete site-wide phantom-link incident (see LNK-004) found that for that incident nav was already real-page-derived and this fallback path was NOT the active cause — the code defect is real and latent, but its blast radius on the observed incident was narrower than first diagnosed.
- **what:** Two brochure-default fallbacks can fabricate structure when resolution fails: `RenderFallbackHeader` (generic header, stacked nav, unwanted search icon) and `AssembleMultipageSiteAction`'s hardcoded 4-item nav. Both were originally flagged as a primary source of phantom `/services.html`-style links; a later, code-grounded investigation traced the actual live-path phantom-link mechanism to `sourceResolver.resolve` fabrication instead (LNK-004), correcting the earlier attribution. Resolution direction for the fallback code itself remains: fallbacks must derive from real pages (`buildNavigationFromPages`) or fail loud, never invent URLs.
- **sources:** FOCUS_internal_linking.md#2; FOCUS_navigation.md#header-footer-rendering; live FOCUS_internal_linking(1).md (2026-06-10) "Correction to the earlier note that blamed 310–318"
- **relations:** phantom-link validation; Tension #1 (silent confident fallbacks); LNK-004 (the corrected root-cause finding)
- **verify-later:** multipage_actions.go lines ~310-318; RenderFallbackHeader callers; whether the fallback-nav code path was ever hardened

### NAV-008 — Tool nav integration
> **SUPERSEDED IN PART, 2026-07-31 (commit `1884f1ee8`, chassis `v1.0.1215`, council
> `4486f1a9` APPROVED). `addToolToNav` — the subject of this entry's `verify-later`
> — IS DELETED.** It was the second writer of a derived table: it hand-wrote a
> `site_nav_items` row into a bespoke `tools` group typed `primary` (the header),
> which `populate_nav_tables`' `DELETE`+rebuild then removed without replacing. Seven
> active rows fleet-wide were in that destroy-and-not-recreate state.
>
> **This entry's open design question is now ANSWERED, and by the platform's own
> stated rule rather than by a new decision.** "Grouping strategy (single Tools entry
> vs individual items)" resolves to: individual tool pages appear **individually, in
> the `utility` group** (footer), never in primary — which is what
> `classifyPagesForNav`'s tier table always said ("Tier 4 (never primary): individual
> tool pages"). The parent `/tools/index.html` listing, being a `section-index`, is
> what represents them in the main menu.
>
> Replacement mechanism: **NAV-013**. Label shortening remains open
> (`navLabelForPage` truncates at 30 chars, and fundamentallyai's live row carries a
> 46-char label written by the deleted function).
- **status:** partial
- **status-evidence:** "Known bug (fixed): addToolToNav used wrong column names … failed silently"; remaining: tools listed individually in primary nav, labels too long (errors-to-fix items 3-5, 18).
- **what:** `create_tool_component` adds a page, page_component and nav entry per tool; a column-name bug was fixed, but grouping strategy (single "Tools" entry vs individual items) and label shortening remain open design work, feeding the site-design-planner's `navigation.tools_strategy` spec.
- **sources:** FOCUS_navigation.md#5; FOCUS_navigation_errors_to_be_fixed.md#3-5
- **relations:** site-design-planner navigation spec; tools pipeline
- **verify-later:** addToolToNav; nav grouping of tool entries on live sites

### NAV-009 — Navigation maintenance: nav-updater and nav-link-fixer
- **status:** deployed
- **status-evidence:** 042 full definition ("Algorithmic only - no LLM calls"); nav-link-fixer in the 075 idle-timeout list; 058 wires it as the fixer for broken_nav_links findings.
- **what:** `nav-updater` refreshes nav tables from current pages (`populate_nav_tables`), re-renders header/footer/head, and reassembles all deployed pages — explicitly distinguished from `rerender-site`, which reuses stale `site_nav_items`. `nav-link-fixer` repairs the `#{{.slug}}` anti-pattern in header/footer component templates (should be `{{.url}}`), then force re-renders site components and pages.
- **sources:** 042_nav_updater_agent.sql; 042b_nav_link_fixer_agent.sql; 058_quality_checks_and_fixers.sql
- **relations:** quality-discovery-agent's broken_nav_links check; orphan_nav finding; rerender pipeline; LNK-016 (nav-link-fixer's scope limitation)
- **verify-later:** populate_nav_tables / fix_nav_link_templates actions

### NAV-010 — Navigation tables (site_nav_groups / site_nav_items)
- **status:** deployed
- **status-evidence:** DDL plus a real-site query result (primary/legal groups for a live site) and the applied global template fix converting anchor links to page URLs.
- **what:** First-class navigation model replacing scattered pages-table queries and the earlier `navigation_structures` cache: groups per site (group_key primary/legal/utility/content, group_type, hierarchy via parent_group_id) containing typed items (page_link/external_link/anchor/section_header, FK to pages with SET NULL, position, status, metadata). Sites without rows fall back to Go logic querying pages directly. Render context supplies both `.slug` and `.url` per item; templates must link `{{.url}}` (a global fix purged `href="#{{.slug}}"` from all header/footer/nav templates).
- **sources:** docs/agent_docs/sql_for_tables/016_nav_tables.sql; docs/agent_docs/sql_for_tables/017_site_nav_groups.sql
- **relations:** site snapshots capture nav; component-based headers consume nav_items; LNK-021 (link-management foundation schema)
- **verify-later:** nav writer agent; fallback path in Go

### NAV-011 — Global context injection for navigation (superseded)
- **status:** superseded
- **status-evidence:** "Context Propagation... any component can access {{.Global.Sitemap}}"; later "reads nav tables directly, falls back to pages table" — superseded by nav tables + GetNavItems.
- **what:** An earlier design treating navigation as data, not structure: the strategist emitted the sitemap first (labels, urls, in_header/in_footer flags) and passed it down as a Global context object so header/footer templates could range over it — pages invented by the strategist automatically appeared in nav. Evolution chain: Global context → sitemap in page_plan → pages-table queries (deployed-only) → NAV-010's site_nav_groups/site_nav_items tables.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#3-Solving-Navigation; docs012_site_maps_and_components/002_site_map_integration.md; docs018_rerendering/003_website_builder_architecture_status_report.md#5
- **relations:** NAV-001 (nav agent family); NAV-010 (successor tables)
- **verify-later:** GetNavItems and populate_nav_tables in component_library.go

### NAV-012 — Header nav from pages.in_header + nav-label hygiene
- **status:** deployed
- **status-evidence:** "Header nav is built from pages.in_header at render time (render_site_components_action.go:550), so setting in_header=false drops a page from the nav without deleting the page" (2026-07-13).
- **what:** Nav membership is data (`pages.in_header`) consumed at header render; decluttering a nav is a data UPDATE, not a template edit. Companion defect: `nav_label` defaults to raw `<title>` strings (e.g. "… | Leopardess Consulting") and needs short, curated labels. Used in practice to cut a ~15-item nav (including a blank 0-section page) down to a business-buyer-relevant set.
- **sources:** docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql (header); docs/leopardessconsulting/AUDIT_verified_facts.md#D3
- **relations:** CTA-graph integrity (LNK-022); link-management
- **verify-later:** render_site_components_action.go:550; pages.in_header usage

### NAV-013 — The nav-membership contract: one declaration, one writer, and a rebuild REQUEST
- **status:** deployed
- **status-evidence:** commit `1884f1ee8` + `8c41e3eaf`, chassis **v1.0.1215**, pod-verified on both replicas. Council `4486f1a9-6d96-4767-9ddd-6ff5e92ba45c` **APPROVED** (12 reviewers, 0 unreadable, 2 medium objections, none high). Diagnosis loop `1d8085f0-b596-4cce-9417-f48227ac67d3` **CONFIRMED** the underlying mechanism, first iteration. `bugs_open/149` A2/A6, A4's recordable half, A5's child half.
- **what:** The stated rule for nav membership, replacing two overlapping implicit ones: **`pages.in_header`/`in_footer` DECLARES nav membership; a page's URL shape may decide WHERE it appears, never WHETHER it appears.** Three parts. (1) `classifyPagesForNav` (`populate_nav_tables_action.go`) collapsed its two notions of "never primary" — one keyed on `page_type` (`blog-post`/`tool`/`entity-page`), one on URL prefix (`/tools/`, `/blog/`, `/guides/`, `/articles/`, `/case-studies/`, `/news/`, `/resources/`, `/insights/`) — into one. The URL-keyed rule used to `continue` out **before either flag was read**, so a nav-flagged child page was dropped entirely; it now demotes to `utility` (footer) exactly as the `page_type`-keyed rule always did. `section-index`/`blog-index`/`news-index`/`entity-directory` keep primary eligibility, keyed on `page_type` not URL. (2) **`site_nav_items` now has exactly ONE writer** — the derivation. `addToolToNav` (NAV-008) is deleted. (3) `RequestNavRebuild` (`nav_rebuild_request.go`) is how a page-creating action asks for its declaration to become real: one work item per site, `item_type` `nav_drift`, `handler_agent` `nav-updater`, `status` `triaged`, `item_key` `nav_rebuild:<site_id>`, `recurrenceExpected: true`. Called by `create_tool_component` and `deploy_tool_to_site`, which also now record `in_header`/`in_footer` explicitly instead of inheriting the schema default `TRUE`.
- **why a REQUEST and not a row (the transferable part):** a nav row is not a link. Chrome is a stored artefact (`bugs_open/117`/`118`), so writing `site_nav_items` changes no served page — while `check_orphan_pages` treats the presence of a nav row as reachability. Writing the row at creation time (which is what `bugs_open/149` A6 literally proposed) would have left the page exactly as unreachable **and silenced the only check that would have noticed**. `nav-updater`'s live workflow (`populate_nav_tables → render_site_components → create_rerender_items → get_pages_for_rerender`) is what actually makes a link ship, so the request invokes that.
- **the evidence, with its control:** a discovery-raised `nav_drift` item for gamesdesign.co.uk naming four `/tools/` pages was `complete` at 17:27:50 on 2026-07-29; all four were still absent from `site_nav_items` on 07-31. The **same** check, handler and action repaired robot-hands.com's `/learning-center.html` and `/news.html` — which differ only in not sitting under a child prefix.
- **LANDMINES for a caller:** (a) **a third caller from OUTSIDE `platform/orchestration/actions` is an RFC moment** — at that point this is a general on-ramp to the work-item queue and has never been reviewed as one (architecture seat, this round). (b) `recurrenceExpected: true` is load-bearing, not decoration: without it `insertWorkItem`'s two-strike rule brands the **third** request on a repeated `item_key` as `unresolved` — terminal, never dispatched — so the third tool added to a site would silently stop reaching the nav. (c) The `item_key` is **site-scoped**, so a second request while one is open writes nothing at all, including its own spec: read `spec.page_name` as "what prompted this", never "what this will fix". (d) The effect is **inert until a nav rebuild runs** — this ships no visible change on its own.
- **sources:** `platform/orchestration/actions/nav_rebuild_request.go` (the argument is in the file header); `populate_nav_tables_action.go:classifyPagesForNav`; `nav_membership_test.go`, `nav_rebuild_request_test.go` (both watched failing on the pre-fix code); `docs024_key_docs_latest/bugfix_149_nav_membership/` (standing five, incl. RUNBOOK R1–R12); `bugs_open/149` A2/A4/A5/A6 banners
- **relations:** NAV-008 (superseded in part — the deleted writer); NAV-002 (two nav systems: this makes the flags authoritative for membership and the tables derived); NAV-010 (the tables); NAV-012 (`pages.in_header` as data); `bugs_open/117`/`118` (chrome is a stored artefact — the reason this is a request); `bugs_open/024` (`recurrenceExpected`'s origin); WRK (work-item integrity)
- **still open, deliberately:** `bugs_open/149` A3 (nothing keeps a parent listing in sync, so an UNFLAGGED tool page is still invisible), A4's schema half (`in_header`/`in_footer` still `DEFAULT TRUE` — architecture scope), A5's non-child half (`buildServicesHTML` still queries `pages` with its own predicate — a ninth query-time nav function `nav_tables.go` was meant to replace)
