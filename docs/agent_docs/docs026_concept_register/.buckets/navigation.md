
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Nav agent family and the three-tier authority model
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** 002(4): owner "currently populate_nav_tables action within pageflow-builder"; tiers described as model
- **what:** Navigation as first-class entity (groups: primary/subsection/content/legal/utility/external; contextual groups planned). Tier 1 strategist authority (new builds), Tier 2 nav-agent autonomous maintenance, Tier 3 drift detection vs original plan. nav-updater/nav-link-fixer handle drift and broken template links today; nav dedup guard recommended after B-029-1 duplicate nav items.
- **sources:** 002(4)#Navigation Agent Family; 024; 029 B-029-1
- **relations:** nav-updater never spawns; populate_nav_tables
- **verify-later:** nav drift check + dedup guard status

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Two nav systems and the GetNavItems fallback
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "Two Nav Systems (and they conflict)" — nav tables intended, pages flags legacy; partial population yields a mix (undated FOCUS, compiled ~2026-04/05)
- **what:** site_nav_groups/site_nav_items (populated by populate_nav_tables, read by GetNavItems) versus pages.in_header/in_footer legacy flags (GetHeaderNavFromPages fallback). GetNavItems tries tables first, falls back to pages — partial population mixes the two. Nav authority tiers designed (Tier 1 planner rebuild — only tier implemented; Tier 2 autonomous nav agent; Tier 3 drift detection). Nav state captured in snapshots and restorable via revert.
- **sources:** FOCUS_navigation.md#1, #2, #7
- **relations:** stale pages problem; nav discovery checks; site-design-planner navigation spec
- **verify-later:** GetNavItems fallback logic; whether Tier 2/3 exist

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Stale pages from previous builds polluting nav
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** "SyncPagesToDBAction uses ON CONFLICT (site_id, name) — it only overwrites matching page names" with fixes listed as "needed" (FOCUS); still item 15 in the errors-to-fix list
- **what:** Pages from prior builds keep in_header=true/status=active and appear in nav though absent from the current plan. Fix design: build_status='deployed' filters on the pages-table nav readers; SyncPagesToDB deactivates stale pages gated by a deactivate_stale_pages flag (new builds deactivate; maintenance/adopt flows preserve).
- **sources:** FOCUS_navigation.md#stale-pages; FOCUS_navigation_errors_to_be_fixed.md#15
- **relations:** two nav systems; adoption faithfulness (preserve semantics)
- **verify-later:** SyncPagesToDBAction current behaviour

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Nav discovery checks and fix agents
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** check/handler tables in FOCUS_navigation (broken_nav_links→nav-link-fixer; checkNavLayout/checkUnwantedElements→component-template-fixer; checkUnlinkedSiteComponents→site-component-linker; orphan_pages→rerender-pages/content-gap-planner)
- **what:** The nav slice of the improvement loop: quality/design/completeness discovery agents detect anchor-slug links, stacked nav (missing flex), unwanted search icons, unlinked header/footer components, orphan pages, missing logo img; dedicated fixers repair templates, relink components (clearing rendered_html + needs_rerender), and make orphans reachable. component-template-fixer's idempotency was case-sensitive, injecting responsive CSS 4× (fix: lowercase compare).
- **sources:** FOCUS_navigation.md#3, #4; FOCUS_navigation_HANDOFF_navigation_fix.md#problems-10
- **relations:** fallback header; duplicate header/footer
- **verify-later:** discovery agent checks arrays; fixInjectResponsiveCSS case fix

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Duplicate header/footer pathology (site-level components in pages.sections)
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** Data fixes applied 2026-04-11 (12 pages.sections rows cleaned, 24 page_components deleted); but 10 dirty rows reappeared by 04-13/14 — "plan_sections filter NOT deployed" (2026-04-20 investigation 7)
- **what:** pages.sections listed site-level component names alongside content sections; rebuilds rendered header/footer as page_components, then InjectHeader/InjectFooter added a second copy. Code fixes designed but pending at doc date: filterSiteLevelSections in PlanSectionsAction (prevents recurrence), skip-if-present guards in InjectHeader/InjectFooter. A discovery check for duplicate headers inside <main> also missing.
- **sources:** FOCUS_navigation_HANDOFF_navigation_fix.md (whole); HANDOFF_2026-04-20_error_investigations.md#7; FOCUS_navigation_errors_to_be_fixed.md#1-2
- **relations:** nav fix agents; page-build-handler
- **verify-later:** plan_sections_action.go for filterSiteLevelSections; component_library.go inject guards

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Nav quality mechanisms of 2026-04-17 (tiers, child-page exclusion, label trust, quick links)
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "What Was Deployed This Session" (2026-04-17): tiered priority, isChildPageURL, navLabelForPage, quick_links_html + footer template SQL
- **what:** populate_nav_tables gained a three-tier page priority (core / hubs+conversion / secondary, overflow to utility) replacing arbitrary nav_order truncation; child-page URL prefixes (/tools/, /blog/ …) excluded from all nav groups; nav labels trust page.NavLabel ≤30 chars and rendering no longer truncates to two words; footer Quick Links built from primary+utility groups via a new quick_links_html variable.
- **sources:** HANDOFF_2026-04-17_nav_empty_sections_footer(1).md#2-5
- **relations:** two nav systems; tool nav integration
- **verify-later:** populate_nav_tables_action.go navPriorityTier/isChildPageURL

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Hardcoded fallback nav/header defaults inventing structure
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "Defect (lines 310–318 of multipage_actions.go): … injects a hardcoded fallback nav — Home/About/Services/Contact" (2026-06-09); RenderFallbackHeader stacked-nav/search-icon behaviour in FOCUS_navigation
- **what:** Two brochure-default fallbacks fabricate structure when resolution fails: RenderFallbackHeader (generic header, stacked nav, unwanted search icon) and AssembleMultipageSiteAction's hardcoded 4-item nav — the primary source of phantom /services.html links. Resolution direction: fallbacks must derive from real pages (buildNavigationFromPages) or fail loud, never invent URLs.
- **sources:** FOCUS_internal_linking.md#2; FOCUS_navigation.md#header-footer-rendering
- **relations:** phantom-link validation; Tension #1 (silent confident fallbacks)
- **verify-later:** multipage_actions.go lines ~310-318; RenderFallbackHeader callers

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Tool nav integration
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** "Known bug (fixed): addToolToNav used wrong column names … failed silently"; remaining: tools listed individually in primary nav, labels too long (errors-to-fix items 3-5, 18)
- **what:** create_tool_component adds a page, page_component and nav entry per tool; column-name bug fixed, but grouping strategy (single "Tools" entry vs individual items) and label shortening remain open design work — feeding the site-design-planner navigation.tools_strategy spec.
- **sources:** FOCUS_navigation.md#5; FOCUS_navigation_errors_to_be_fixed.md#3-5
- **relations:** site-design-planner navigation spec; tools pipeline
- **verify-later:** addToolToNav; nav grouping of tool entries on live sites

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Nav sync & config-driven page deactivation
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** 001(0) "Nav Sync: Config-Driven Page Deactivation … deactivate pages not in the current plan"; deactivate_stale_pages config flag
- **what:** Header/footer nav displayed stale pages because `SyncPagesToDBAction`'s `ON CONFLICT` only overwrote matching names and nav queries didn't filter `build_status`. Fix: nav getters add `AND build_status = 'deployed'`, and a new-build flow deactivates pages absent from the current plan gated by `deactivate_stale_pages: true`.
- **sources:** WM/001_development_guide(0).md#nav-sync-config-driven-page-deactivation, ED/102_blog_handoff-2026-04-10.md#a-check_orphan_pagesgo-new-routing-logic
- **relations:** site plan reconciler nav auditor; link management; blog-listing handoff
- **verify-later:** SyncPagesToDBAction; GetHeaderNavFromPages/GetFooterNavFromPages; site_nav_items

<!-- SOURCE: U18_sql_for_agents.md -->
### Navigation maintenance: nav-updater and nav-link-fixer
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** 042 full definition ("Algorithmic only - no LLM calls"); nav-link-fixer in 075 idle-timeout list; 058 wires it as fixer for broken_nav_links findings.
- **what:** nav-updater refreshes nav tables from current pages (populate_nav_tables), re-renders header/footer/head and reassembles all deployed pages — explicitly distinguished from rerender-site, which reuses stale site_nav_items. nav-link-fixer repairs the `#{{.slug}}` anti-pattern in header/footer component templates (should be `{{.url}}`), then force re-renders site components and pages.
- **sources:** 042_nav_updater_agent.sql; 042b_nav_link_fixer_agent.sql; 058_quality_checks_and_fixers.sql
- **relations:** quality-discovery-agent's broken_nav_links check; orphan_nav finding; rerender pipeline
- **verify-later:** populate_nav_tables / fix_nav_link_templates actions

<!-- SOURCE: U19_sql_tables_components.md -->
### Navigation tables (site_nav_groups / site_nav_items)
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** DDL plus a real-site query result (primary/legal groups for a live site) and the applied global template fix converting anchor links to page URLs.
- **what:** First-class navigation model replacing scattered pages-table queries and the navigation_structures cache: groups per site (group_key primary/legal/utility/content, group_type, hierarchy via parent_group_id) containing typed items (page_link/external_link/anchor/section_header, FK to pages with SET NULL, position, status, metadata). Sites without rows fall back to Go logic querying pages directly. Render context supplies both .slug and .url per item; templates must link {{.url}} (061 fix purged href="#{{.slug}}" from all header/footer/nav templates).
- **sources:** docs/agent_docs/sql_for_tables/016_nav_tables.sql; docs/agent_docs/sql_for_tables/017_site_nav_groups.sql
- **relations:** site snapshots capture nav; component-based headers consume nav_items.
- **verify-later:** nav writer agent; fallback path in Go.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Global context injection for navigation
- **category:** navigation
- **status-signal:** superseded
- **status-evidence:** docs009/001 "Context Propagation... any component can access {{.Global.Sitemap}}"; docs012/002 adds explicit sitemap JSON to strategist output; superseded by nav tables + GetNavItems (docs017/019b "reads nav tables directly, falls back to pages table").
- **what:** Navigation treated as data, not structure: the strategist emits the sitemap first (labels, urls, in_header/in_footer flags), and it is passed down as a Global context object so header/footer templates range over it — pages invented by the strategist automatically appear in nav. Evolution chain: Global context → sitemap in page_plan → pages-table queries (deployed-only) → site_nav_groups/site_nav_items tables.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#3-Solving-Navigation; docs012_site_maps_and_components/002_site_map_integration.md; docs018_rerendering/003_website_builder_architecture_status_report.md#5
- **relations:** nav agent family; navigation-from-pages; three-tier authority model.
- **verify-later:** GetNavItems and populate_nav_tables in component_library.go.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Navigation agent family + three-tier authority
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** docs017/019b: "core responsibilities are implemented as the populate_nav_tables action... full standalone nav-agent is planned but not yet needed"; utility classification list and nav data flow marked (implemented).
- **what:** Navigation as a first-class entity: site_nav_groups/site_nav_items with typed groups (primary, subsection, content, legal, utility, external, contextual); populate_nav_tables classifies pages (FAQ/Blog/Careers etc. routed to utility even if in_header); GetNavItems serves header (primary, deployed-only) and footer (primary+utility+legal) rendering with pages-table fallback. Authority tiers: strategist owns structure at build; nav agent makes incremental decisions in maintenance; periodic drift detection compares current nav against the original plan ("drift may represent valid evolution").
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#1-Navigation-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Three-Tier-Authority-Model
- **relations:** navigation-from-pages (predecessor); nav-updater fix agent; current navigation FOCUS docs.
- **verify-later:** site_nav_groups/site_nav_items tables; populate_nav_tables action; standalone nav-agent existence.

<!-- SOURCE: U25_leopardess_social.md -->
### Header nav from pages.in_header + nav-label hygiene
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** L5_nav_and_ctas.sql header (2026-07-13): "Header nav is built from pages.in_header at render time (render_site_components_action.go:550), so setting in_header=false drops a page from the nav without deleting the page."
- **what:** Nav membership is data (pages.in_header) consumed at header render; decluttering is an UPDATE, not a template edit. Companion defect: nav_label defaults to raw `<title>` strings ("… | Leopardess Consulting") and needs short labels (AUDIT D3). Used to cut a ~15-item nav (including a blank 0-section page) to a business-buyer set.
- **sources:** docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql (header); docs/leopardessconsulting/AUDIT_verified_facts.md#D3
- **relations:** CTA-graph integrity (vonc); link-management
- **verify-later:** render_site_components_action.go:550; pages.in_header usage

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Nav agent family and the three-tier authority model
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** 002(4): owner "currently populate_nav_tables action within pageflow-builder"; tiers described as model
- **what:** Navigation as first-class entity (groups: primary/subsection/content/legal/utility/external; contextual groups planned). Tier 1 strategist authority (new builds), Tier 2 nav-agent autonomous maintenance, Tier 3 drift detection vs original plan. nav-updater/nav-link-fixer handle drift and broken template links today; nav dedup guard recommended after B-029-1 duplicate nav items.
- **sources:** 002(4)#Navigation Agent Family; 024; 029 B-029-1
- **relations:** nav-updater never spawns; populate_nav_tables
- **verify-later:** nav drift check + dedup guard status

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Two nav systems and the GetNavItems fallback
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "Two Nav Systems (and they conflict)" — nav tables intended, pages flags legacy; partial population yields a mix (undated FOCUS, compiled ~2026-04/05)
- **what:** site_nav_groups/site_nav_items (populated by populate_nav_tables, read by GetNavItems) versus pages.in_header/in_footer legacy flags (GetHeaderNavFromPages fallback). GetNavItems tries tables first, falls back to pages — partial population mixes the two. Nav authority tiers designed (Tier 1 planner rebuild — only tier implemented; Tier 2 autonomous nav agent; Tier 3 drift detection). Nav state captured in snapshots and restorable via revert.
- **sources:** FOCUS_navigation.md#1, #2, #7
- **relations:** stale pages problem; nav discovery checks; site-design-planner navigation spec
- **verify-later:** GetNavItems fallback logic; whether Tier 2/3 exist

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Stale pages from previous builds polluting nav
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** "SyncPagesToDBAction uses ON CONFLICT (site_id, name) — it only overwrites matching page names" with fixes listed as "needed" (FOCUS); still item 15 in the errors-to-fix list
- **what:** Pages from prior builds keep in_header=true/status=active and appear in nav though absent from the current plan. Fix design: build_status='deployed' filters on the pages-table nav readers; SyncPagesToDB deactivates stale pages gated by a deactivate_stale_pages flag (new builds deactivate; maintenance/adopt flows preserve).
- **sources:** FOCUS_navigation.md#stale-pages; FOCUS_navigation_errors_to_be_fixed.md#15
- **relations:** two nav systems; adoption faithfulness (preserve semantics)
- **verify-later:** SyncPagesToDBAction current behaviour

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Nav discovery checks and fix agents
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** check/handler tables in FOCUS_navigation (broken_nav_links→nav-link-fixer; checkNavLayout/checkUnwantedElements→component-template-fixer; checkUnlinkedSiteComponents→site-component-linker; orphan_pages→rerender-pages/content-gap-planner)
- **what:** The nav slice of the improvement loop: quality/design/completeness discovery agents detect anchor-slug links, stacked nav (missing flex), unwanted search icons, unlinked header/footer components, orphan pages, missing logo img; dedicated fixers repair templates, relink components (clearing rendered_html + needs_rerender), and make orphans reachable. component-template-fixer's idempotency was case-sensitive, injecting responsive CSS 4× (fix: lowercase compare).
- **sources:** FOCUS_navigation.md#3, #4; FOCUS_navigation_HANDOFF_navigation_fix.md#problems-10
- **relations:** fallback header; duplicate header/footer
- **verify-later:** discovery agent checks arrays; fixInjectResponsiveCSS case fix

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Duplicate header/footer pathology (site-level components in pages.sections)
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** Data fixes applied 2026-04-11 (12 pages.sections rows cleaned, 24 page_components deleted); but 10 dirty rows reappeared by 04-13/14 — "plan_sections filter NOT deployed" (2026-04-20 investigation 7)
- **what:** pages.sections listed site-level component names alongside content sections; rebuilds rendered header/footer as page_components, then InjectHeader/InjectFooter added a second copy. Code fixes designed but pending at doc date: filterSiteLevelSections in PlanSectionsAction (prevents recurrence), skip-if-present guards in InjectHeader/InjectFooter. A discovery check for duplicate headers inside <main> also missing.
- **sources:** FOCUS_navigation_HANDOFF_navigation_fix.md (whole); HANDOFF_2026-04-20_error_investigations.md#7; FOCUS_navigation_errors_to_be_fixed.md#1-2
- **relations:** nav fix agents; page-build-handler
- **verify-later:** plan_sections_action.go for filterSiteLevelSections; component_library.go inject guards

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Nav quality mechanisms of 2026-04-17 (tiers, child-page exclusion, label trust, quick links)
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "What Was Deployed This Session" (2026-04-17): tiered priority, isChildPageURL, navLabelForPage, quick_links_html + footer template SQL
- **what:** populate_nav_tables gained a three-tier page priority (core / hubs+conversion / secondary, overflow to utility) replacing arbitrary nav_order truncation; child-page URL prefixes (/tools/, /blog/ …) excluded from all nav groups; nav labels trust page.NavLabel ≤30 chars and rendering no longer truncates to two words; footer Quick Links built from primary+utility groups via a new quick_links_html variable.
- **sources:** HANDOFF_2026-04-17_nav_empty_sections_footer(1).md#2-5
- **relations:** two nav systems; tool nav integration
- **verify-later:** populate_nav_tables_action.go navPriorityTier/isChildPageURL

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Hardcoded fallback nav/header defaults inventing structure
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "Defect (lines 310–318 of multipage_actions.go): … injects a hardcoded fallback nav — Home/About/Services/Contact" (2026-06-09); RenderFallbackHeader stacked-nav/search-icon behaviour in FOCUS_navigation
- **what:** Two brochure-default fallbacks fabricate structure when resolution fails: RenderFallbackHeader (generic header, stacked nav, unwanted search icon) and AssembleMultipageSiteAction's hardcoded 4-item nav — the primary source of phantom /services.html links. Resolution direction: fallbacks must derive from real pages (buildNavigationFromPages) or fail loud, never invent URLs.
- **sources:** FOCUS_internal_linking.md#2; FOCUS_navigation.md#header-footer-rendering
- **relations:** phantom-link validation; Tension #1 (silent confident fallbacks)
- **verify-later:** multipage_actions.go lines ~310-318; RenderFallbackHeader callers

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Tool nav integration
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** "Known bug (fixed): addToolToNav used wrong column names … failed silently"; remaining: tools listed individually in primary nav, labels too long (errors-to-fix items 3-5, 18)
- **what:** create_tool_component adds a page, page_component and nav entry per tool; column-name bug fixed, but grouping strategy (single "Tools" entry vs individual items) and label shortening remain open design work — feeding the site-design-planner navigation.tools_strategy spec.
- **sources:** FOCUS_navigation.md#5; FOCUS_navigation_errors_to_be_fixed.md#3-5
- **relations:** site-design-planner navigation spec; tools pipeline
- **verify-later:** addToolToNav; nav grouping of tool entries on live sites

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Nav sync & config-driven page deactivation
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** 001(0) "Nav Sync: Config-Driven Page Deactivation … deactivate pages not in the current plan"; deactivate_stale_pages config flag
- **what:** Header/footer nav displayed stale pages because `SyncPagesToDBAction`'s `ON CONFLICT` only overwrote matching names and nav queries didn't filter `build_status`. Fix: nav getters add `AND build_status = 'deployed'`, and a new-build flow deactivates pages absent from the current plan gated by `deactivate_stale_pages: true`.
- **sources:** WM/001_development_guide(0).md#nav-sync-config-driven-page-deactivation, ED/102_blog_handoff-2026-04-10.md#a-check_orphan_pagesgo-new-routing-logic
- **relations:** site plan reconciler nav auditor; link management; blog-listing handoff
- **verify-later:** SyncPagesToDBAction; GetHeaderNavFromPages/GetFooterNavFromPages; site_nav_items

<!-- SOURCE: U18_sql_for_agents.md -->
### Navigation maintenance: nav-updater and nav-link-fixer
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** 042 full definition ("Algorithmic only - no LLM calls"); nav-link-fixer in 075 idle-timeout list; 058 wires it as fixer for broken_nav_links findings.
- **what:** nav-updater refreshes nav tables from current pages (populate_nav_tables), re-renders header/footer/head and reassembles all deployed pages — explicitly distinguished from rerender-site, which reuses stale site_nav_items. nav-link-fixer repairs the `#{{.slug}}` anti-pattern in header/footer component templates (should be `{{.url}}`), then force re-renders site components and pages.
- **sources:** 042_nav_updater_agent.sql; 042b_nav_link_fixer_agent.sql; 058_quality_checks_and_fixers.sql
- **relations:** quality-discovery-agent's broken_nav_links check; orphan_nav finding; rerender pipeline
- **verify-later:** populate_nav_tables / fix_nav_link_templates actions

<!-- SOURCE: U19_sql_tables_components.md -->
### Navigation tables (site_nav_groups / site_nav_items)
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** DDL plus a real-site query result (primary/legal groups for a live site) and the applied global template fix converting anchor links to page URLs.
- **what:** First-class navigation model replacing scattered pages-table queries and the navigation_structures cache: groups per site (group_key primary/legal/utility/content, group_type, hierarchy via parent_group_id) containing typed items (page_link/external_link/anchor/section_header, FK to pages with SET NULL, position, status, metadata). Sites without rows fall back to Go logic querying pages directly. Render context supplies both .slug and .url per item; templates must link {{.url}} (061 fix purged href="#{{.slug}}" from all header/footer/nav templates).
- **sources:** docs/agent_docs/sql_for_tables/016_nav_tables.sql; docs/agent_docs/sql_for_tables/017_site_nav_groups.sql
- **relations:** site snapshots capture nav; component-based headers consume nav_items.
- **verify-later:** nav writer agent; fallback path in Go.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Global context injection for navigation
- **category:** navigation
- **status-signal:** superseded
- **status-evidence:** docs009/001 "Context Propagation... any component can access {{.Global.Sitemap}}"; docs012/002 adds explicit sitemap JSON to strategist output; superseded by nav tables + GetNavItems (docs017/019b "reads nav tables directly, falls back to pages table").
- **what:** Navigation treated as data, not structure: the strategist emits the sitemap first (labels, urls, in_header/in_footer flags), and it is passed down as a Global context object so header/footer templates range over it — pages invented by the strategist automatically appear in nav. Evolution chain: Global context → sitemap in page_plan → pages-table queries (deployed-only) → site_nav_groups/site_nav_items tables.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#3-Solving-Navigation; docs012_site_maps_and_components/002_site_map_integration.md; docs018_rerendering/003_website_builder_architecture_status_report.md#5
- **relations:** nav agent family; navigation-from-pages; three-tier authority model.
- **verify-later:** GetNavItems and populate_nav_tables in component_library.go.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Navigation agent family + three-tier authority
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** docs017/019b: "core responsibilities are implemented as the populate_nav_tables action... full standalone nav-agent is planned but not yet needed"; utility classification list and nav data flow marked (implemented).
- **what:** Navigation as a first-class entity: site_nav_groups/site_nav_items with typed groups (primary, subsection, content, legal, utility, external, contextual); populate_nav_tables classifies pages (FAQ/Blog/Careers etc. routed to utility even if in_header); GetNavItems serves header (primary, deployed-only) and footer (primary+utility+legal) rendering with pages-table fallback. Authority tiers: strategist owns structure at build; nav agent makes incremental decisions in maintenance; periodic drift detection compares current nav against the original plan ("drift may represent valid evolution").
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#1-Navigation-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Three-Tier-Authority-Model
- **relations:** navigation-from-pages (predecessor); nav-updater fix agent; current navigation FOCUS docs.
- **verify-later:** site_nav_groups/site_nav_items tables; populate_nav_tables action; standalone nav-agent existence.

<!-- SOURCE: U25_leopardess_social.md -->
### Header nav from pages.in_header + nav-label hygiene
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** L5_nav_and_ctas.sql header (2026-07-13): "Header nav is built from pages.in_header at render time (render_site_components_action.go:550), so setting in_header=false drops a page from the nav without deleting the page."
- **what:** Nav membership is data (pages.in_header) consumed at header render; decluttering is an UPDATE, not a template edit. Companion defect: nav_label defaults to raw `<title>` strings ("… | Leopardess Consulting") and needs short labels (AUDIT D3). Used to cut a ~15-item nav (including a blank 0-section page) to a business-buyer set.
- **sources:** docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql (header); docs/leopardessconsulting/AUDIT_verified_facts.md#D3
- **relations:** CTA-graph integrity (vonc); link-management
- **verify-later:** render_site_components_action.go:550; pages.in_header usage
