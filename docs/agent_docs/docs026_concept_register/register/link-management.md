# Register — link-management

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

23 concepts (LNK-023 added 2026-07-28, post-freeze), consolidated from 56 raw extractions (28 unique blocks, each mechanically duplicated once in the cluster input file — see note in styling-render-pipeline.md) across units U01, U02, U05, U21, U23, U24d, U24f, U25.

### LNK-001 — link_registry as first-class link index + planned links-orchestrator family
- **status:** partial
- **status-evidence:** Current doc 024: schema + extract/sync + constraints + validation exist; links-orchestrator agent family "planned but not implemented"; delete-and-reinsert loses validation history (known limitation). An earlier design doc (docs012) describes the same registry-as-index concept as already implemented at that time.
- **what:** Every anchor in rendered HTML is meant to live in `link_registry` (scope internal/page/external; type navigation/content/semantic/affiliate/reference; affiliate fields; validation state) as a derived index — links are never stored as primary data, they exist inside rendered components, and the registry is populated by post-build extraction (`extract_and_sync_links`, delete+reinsert per page). `InjectLinkConstraints` feeds valid pages into writer prompts to prevent invented links; `validateInternalLinks` warns (not blocks) on missing targets. Nav structure is a separate concern (site_nav_groups/items). The originally proposed agent family sitting on top of this index — link-crawler/validator/registry-sync/redirect-manager/affiliate-manager — remains planned but unbuilt (see LNK-019).
- **sources:** 024 full; docs012_site_maps_and_components/006_start_concluding_links.md#2.5; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-2; docs012_site_maps_and_components/007_link_migration.sql
- **relations:** orphan_pages/internal-linker; LNK-014 (phantom-CTA bug); LNK-019 (the planned agent family); NAV-001 (nav agent family)
- **verify-later:** link_registry population; HTTP validation anywhere; extract_and_sync_links wiring status

### LNK-002 — Internal linking machinery and its defects
- **status:** partial
- **status-evidence:** "current as of 2026-06-09. Grounded in multipage_actions.go, site_db_actions.go, queryresolve…"; defects catalogued: hardcoded fallback nav, unpopulated `*_index_url` specs, phantom /services.html.
- **what:** The `pages` table (via `upsertPage` slug/url/nav_label) is the authority for link targets; nav is built from real pages or DB nav structure; `fixAnchorLinks` bridges single-page anchors to multipage URLs; `queryresolve` fills list-hub cards; "Browse All X" buttons read `*_index_url` site_specs (inconsistent sources, often empty → `href=""`, see LNK-010); `ExtractAndSyncLinksAction` maintains a per-page link_registry — the natural substrate for a phantom-link discovery check that didn't yet exist at this point (later built as LNK-009). Hero CTA destinations are the linking half of the site-wide CTA defect (LNK-003).
- **sources:** FOCUS_internal_linking.md (whole)
- **relations:** hardcoded fallbacks (NAV-007); content quality catalogue; LNK-003
- **verify-later:** syncLinksToDB (records vs validates); link_registry schema; hero component input_schema

### LNK-003 — Hero CTA brochure-default defect (text↔destination mismatch)
- **status:** deployed
- **status-evidence:** "Phantom hero CTA — FIX APPLIED 2026-06-26"; readback confirmed against a live snapshot.
- **what:** The generic hero/call-to-action component schemas carried brochure-site defaults (`cta_url ← pages.contact`, `secondary_cta_url ← pages.services`) while button text was LLM-written — so every hero site-wide linked "Browse Tools" to /contact.html and to the phantom /services.html. Root causes were fixed in layers: the `pages` source fabrication (LNK-004), schema/template hardening (LNK-006), and finally the writer's select_sections path mismatch that discarded the resolver's correct hubs (LNK-014).
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; HANDOFF_2026-06-09(2).md#next-task; NOTES_gamesdesign_silent_norebuild(44).md (2026-06-22/23 sessions)
- **relations:** LNK-004; LNK-006; LNK-011 (internal-link-resolver agent); LNK-014
- **verify-later:** content_components rows hero/call-to-action input_schema; deployed guide-economy-basics hero HTML

### LNK-004 — sourceResolver `pages` fabrication bug — the phantom generator
- **status:** deployed
- **status-evidence:** "Layer 1a … DEPLOYED + APPLIED + VERIFIED … `resolve` case `pages` no longer fabricates." This also corrected an earlier (2026-06-09) diagnosis that had blamed a hardcoded fallback-nav code path (multipage_actions.go:310–318, see NAV-007) as the root cause: a code-grounded re-investigation the next day found nav was already real-page-derived, and the true mechanism was this resolver bug plus a separate hardcoded-ContentData issue (LNK-007).
- **what:** `sourceResolver.resolve` (plan_sections_action.go) `case "pages"` fabricated `"/" + path + ".html"` and returned `found=true` for any non-existent page, so `on_missing` never fired and schema fallbacks were dead code — the single machine that minted every hero phantom link. Fixed to return the real URL or `(nil, false)`. Blast radius was tight: hero + call-to-action are the only components with a `pages.*` source.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; running_notes_17(21).md#decisive-findings; step1_hero_cta_phantom_fix.sql; content_quality_and_internal_linking/FOCUS_internal_linking.md (archived, 2026-06-09, the superseded hypothesis)
- **relations:** LNK-003 (hero CTA defect); LNK-006 (schema hardening); LNK-005 (correct-or-absent principle); NAV-007 (the corrected earlier hypothesis)
- **verify-later:** platform plan_sections_action.go resolve() pages case

### LNK-005 — Correct-or-absent principle + loud-but-non-blocking phantom policy
- **status:** deployed
- **status-evidence:** "Policy (settled this round)" 2026-06-10.
- **what:** The structural rule for all internal links: targets resolve from the real `pages` set, never fabricated or brochure-assumed; an unresolvable destination renders nothing (no button) rather than a broken/empty link, with a build-time signal so the absence isn't silent. Companion policy: a phantom/missing internal link is loud but non-blocking — a deploy-gate warning, not an error; the improvement loop resolves it over time.
- **sources:** FOCUS_internal_linking(1).md#through-line; running_notes_17(21).md#policy-settled; PLAN_b4_b5_hubs_and_link_resolver(3).md
- **relations:** LNK-012 (unresolved_cta signal); validate_page_content gate; every phantom-fix layer
- **verify-later:** validate_page_content.go warning severities; hero/CTA template gates in content_components

### LNK-006 — Step 1 / Layer 1a hero+CTA schema/template hardening
- **status:** deployed
- **status-evidence:** "Verified: both components skip_field/fallbacks_gone/has_and_gate all true."
- **what:** SQL that sets hero/call-to-action CTA-url fields to `on_missing: skip_field`, removes phantom fallbacks (/contact.html, #features), and gates each button template on `{{if and .cta_text .cta_url}}` — so an unresolved CTA renders no button. Shipped coupled with the Go `resolve()` fix (LNK-004): order matters, since without it the gate would still receive a truthy phantom.
- **sources:** step1_hero_cta_phantom_fix.sql; check_linking_sql_applied.sql; RUNBOOK_linking_phantom_fixes(7).md#1
- **relations:** LNK-004 (resolver fabrication fix); LNK-011 (internal-link-resolver restores real destinations)
- **verify-later:** content_components hero/call-to-action html_template + input_schema

### LNK-007 — Layer 1b header/footer phantom fix (shared site components)
- **status:** deployed
- **status-evidence:** "Layer 1b … Verified gone (§4 audit: 0 site_component findings)."
- **what:** Header/footer phantoms (/contact.html, /privacy.html, /terms.html) came from hardcoded ContentData in `render_site_components_action.go`, not templates or nav fallback. Fixed at source: header cta_url resolved from the real contact page, footer legal links data-driven from `GetNavItems(NavGroupLegal)`, header CTA gated on cta_url. Being shared components, the edit benefited every site immediately.
- **sources:** layer1b_header_footer_phantom_fix.sql; FOCUS_internal_linking(1).md#shipped-this-round; NOTES(44) 2026-06-22
- **relations:** LNK-016 (nav-link-fixer can't reach ContentData literals); LNK-004
- **verify-later:** render_site_components_action.go lines ~141–233; footer-4-column/header-bold-gradient templates

### LNK-008 — datahelpers/links.go — canonical link classification library
- **status:** deployed
- **status-evidence:** "Shared machinery (single source of truth)" — "Replaces three previously-divergent normalisers."
- **what:** One shared library (`ExtractHrefs`, `ClassifyLinkScope`, `IsAssetPath`, `NormalizePagePath`, `PageURLSet`) used by both the deploy gate and the post-deploy audit so they agree by construction. Replaced three divergent URL normalisers (validator lowercased+appended .html; audit stripped index.html; inventory ignored assets).
- **sources:** FOCUS_internal_linking(1).md#shared-machinery; running_notes_17(21).md#shipped
- **relations:** LNK-009 (validate_page_content gate; check_phantom_internal_links audit)
- **verify-later:** platform/orchestration/datahelpers/links.go

### LNK-009 — check_phantom_internal_links post-deploy audit + surface routing
- **status:** partial
- **status-evidence:** "NOT YET ENABLED (deliberate, observe-only later)"; "gate cleared, ready when you choose."
- **what:** A discovery check scanning page_components + site_components rendered_html for phantom/empty internal links, routing per surface: site_component → nav-link-fixer (build), page_component → page-build-handler (content; a rebuild re-runs build-time resolution). Home agent settled as completeness-discovery-agent (content-integrity family). Deliberately inert until enabled; enabling ≠ autonomous remediation because findings land `status='detected'` (unclaimable). An earlier version routed page_component findings to internal-link-resolver directly — superseded; a stale duplicate copy with that routing is marked for deletion.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#7a; FOCUS_internal_linking(1).md#shared-machinery; running_notes_17(21).md#§7-gate-RESOLVED; README_find_phantom_links.sql
- **relations:** observe-only enablement pattern; NAV-009 (nav-link-fixer); LNK-011 (internal-link-resolver)
- **verify-later:** discovery_checks/check_phantom_internal_links.go routeBySurface; completeness-discovery-agent run_checks.config.checks array

### LNK-010 — B4/B5 Browse-All hub links via `section_index_for` queryresolve verb
- **status:** deployed
- **status-evidence:** "B4/B5 (SQL) … Verified"; "Browse All Games → /games/index.html … B4/B5 confirmed."
- **what:** The three list components' "Browse All X" buttons rendered `href=""` because they sourced `cta_url` from unpopulated, inconsistently-named `*_index_url` site_specs. Three options were weighed: (a) populate the specs from real hubs — rejected as brittle per-site maintenance; (b) a per-component `pages.<hub-name>` source — rejected for baking a naming convention into every schema; (c) recommended and shipped — a new `queryresolve` verb `section_index_for:<type>` deriving the hub URL from real page relationships (shared-area lookup, URL-prefix fallback), plus template gates `{{if .cta_url}}`. Notable trap discovered along the way: for `query.*` fields the field loop never consults `on_missing` and would apply `fallback` on nil — hence source-only schema changes and gate-in-template.
- **sources:** PLAN_b4_b5_hubs_and_link_resolver(3).md; b4_b5_hub_links_schema.sql; b4_b5_hub_links_template_gate.sql; content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md
- **relations:** LNK-005 (correct-or-absent); Tier-D list components; LNK-011
- **verify-later:** queryresolve/section_index_for.go + Resolve switch; tool-list/game-list_pre_037/guide-list_pre_037 schemas

### LNK-011 — internal-link-resolver agent
- **status:** deployed
- **status-evidence:** "Step 3 … Wiring confirmed LIVE"; agent row applied 2026-06-11; confirmed live end-to-end via matched for_render/plan_count queries on completed rebuilds.
- **what:** A dedicated sub-agent (spawned by page-content-writer, called once per page; modelled on research-agent, no persistence) whose single responsibility is resolving intent-appropriate internal link destinations from the real pages set — hero/CTA fields augmented in section resolved_data at build time. v1 rules are deterministic (top content hubs by nav_order, excluding about/contact/legal and the page's own hub); the agent boundary deliberately allows a future LLM intent-matching upgrade without changing callers. Explicitly "a build-time augmenter, not a rendered-HTML patcher." Targets are validated via `PageURLSet` so it cannot emit a URL the deploy gate would flag. Replaces the abandoned assumption that component-template-fixer already handled this.
- **sources:** internal_link_resolver_agent.sql; PLAN_b4_b5_hubs_and_link_resolver(3).md#step-3; running_notes_17(21).md#step-3; content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md
- **relations:** page-content-writer wiring (LNK-013); LNK-012 (unresolved_cta signal); ctaFieldNames coverage gap
- **verify-later:** agent_definitions row type='internal-link-resolver'; resolve_internal_links_action.go (chooseCTATargets, ctaFieldNames, setCTAField)

### LNK-012 — unresolved_cta build-time HITL signal
- **status:** deployed
- **status-evidence:** watch SQL "B) resolver distress signal — must stay 0"; observed 0 throughout a batch run.
- **what:** When a section has CTA text but no resolvable real-page destination, the resolver emits an `unresolved_cta` work item (needs_human_review; one per affected section, mirroring createDeferredItems, ON CONFLICT dedup). Rationale: the deploy gate cannot see a correctly-dropped button — there is no fingerprint in rendered HTML — so resolution time is the only place the absence is detectable pre-deploy.
- **sources:** PLAN_b4_b5_hubs_and_link_resolver(3).md#unresolved_cta; running_notes_17(21).md#step-3-completed; FOCUS_internal_linking(1).md#remaining
- **relations:** LNK-005 (correct-or-absent principle); HITL machinery
- **verify-later:** ResolveInternalLinksAction unresolved_cta emission; site_work_items item_type='unresolved_cta'

### LNK-013 — page-content-writer ↔ resolver wiring with regression-safe fallback chain
- **status:** deployed
- **status-evidence:** "all 7 verification columns correct" (applied 2026-06-11).
- **what:** Workflow-only wiring: spawn_link_resolver, resolve_links (call_agent, error_step falls through), select_sections (extract_fields with a fallback chain: resolver-augmented sections, else the original plan), loop repointed to sections_for_render. Designed so resolver failure is byte-identical to prior behaviour — which later proved double-edged: the fallback silently masked a path mismatch for two weeks (LNK-014).
- **sources:** page_content_writer_link_resolver_wiring.sql; running_notes_17(21).md#step-3-completed
- **relations:** LNK-014 (the fallback's dark side); result-contract work
- **verify-later:** page-content-writer default_config workflow steps

### LNK-014 — select_sections path-mismatch bug (phantom CTA root cause)
- **status:** deployed
- **status-evidence:** First surfaced 2026-06-22 on the vonc/gamesdesign site ("hero carries two phantom CTAs from schema sources pages.contact/pages.services... workflow-only fix staged"); root cause confirmed 2026-06-23 and the fix applied+confirmed 2026-06-26 (snapshot 5946a27b).
- **what:** The resolver ran and returned augmented sections, but the `call_agent` envelope nests the reply at `resolved_links.response.link_resolution.sections_ready` while `select_sections` read the top-level `resolved_links.sections_ready` → null → silent fallback to the un-augmented plan carrying schema phantoms. A one-line jsonb_set repoint fixed it; takes effect only on a full content build (a bare re-render doesn't re-run the resolver). Distinct from the interactive-section clobber; `page_rerender` does not re-resolve schema-sourced CTAs. Two follow-ups remain open: `ctaFieldNames` matches only exact "hero"/"call-to-action" (variants like hero-about/gauntlet-cta never resolve), and the resolver returns its whole echoed collected_data with empty final_result (should return a lean `{sections_ready, unresolved}`).
- **sources:** HANDOFF_page_pipeline(11).md#3; NOTES_gamesdesign_silent_norebuild(44).md 2026-06-23/26; docs/016b_debugging_guide_merged(3).md#open-threads (Part 4); docs/RUNBOOK_vonc_session(1).md#remaining-steps
- **relations:** LNK-013 (wiring fallback chain); LNK-003 (hero CTA defect); LNK-020 (site_specs cta aspect)
- **verify-later:** page-content-writer select_sections config; guide-economy-basics hero has_phantom_cta after build e26cd02f

### LNK-015 — link_registry — records but never validates (dormant substrate, abandoned)
- **status:** abandoned
- **status-evidence:** "syncLinksToDB never populates it … wired into no live workflow, so link_registry is empty. It is not a usable substrate today." Independently reconsidered later as a candidate substrate for a new phantom-link check and re-confirmed unusable for the same reason.
- **what:** A per-page link inventory table with a `target_page_id` column + FK to `pages`, intended as a phantom-check substrate, but the sync never populates `target_page_id` and `ExtractAndSyncLinksAction` is wired into no live workflow. The phantom audit that was actually built (`check_phantom_internal_links.go`, LNK-009) reads `rendered_html` directly instead, via the shared `datahelpers/links.go` (LNK-008).
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; running_notes_16_content_quality_and_internal_linking(1).md#part-1; content_quality_and_internal_linking/FOCUS_internal_linking.md (archived)
- **relations:** LNK-009 (the live approach that superseded it)
- **verify-later:** link_registry row counts; ExtractAndSyncLinksAction callers

### LNK-016 — nav-link-fixer agent (template-anchor scope only)
- **status:** deployed
- **status-evidence:** README_find_phantom_links.sql output: nav-link-fixer exists, status experimental, is_active=t.
- **what:** The site_component-surface link fixer: find/replaces `#{{.slug}}`/`#{{.name}}` anchors inside html_template (`fix_nav_link_templates_action.go`). Its scope excludes ContentData values and literal anchors — which is why the header/footer phantoms (LNK-007) had to be fixed at source in Go instead.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings-4; README_find_phantom_links.sql
- **relations:** LNK-007 (Layer 1b); LNK-009 (site_component surface routing); NAV-009 (nav-updater/nav-link-fixer pair)
- **verify-later:** fix_nav_link_templates_action.go; nav-link-fixer agent_definitions row

### LNK-017 — prepare_link_context available_pages gap on the work-item path
- **status:** partial
- **status-evidence:** "This path maps NO db_sync to the writer → prepare_link_context/available_pages get nothing … Pre-existing, resolver-independent" (2026-06-12).
- **what:** The writer's `prepare_link_context` builds an available-pages constraint for the LLM's in-prose internal linking, but on the work-item rebuild path no db_sync is mapped, so the constraint text is empty — the LLM writes prose links unconstrained. Independent of the resolver (which queries the DB directly). Candidate fixes: map db_sync, or make prepare_link_context load pages itself.
- **sources:** running_notes_17(21).md#page-build-handler-contract watch items
- **relations:** LNK-011 (internal-link-resolver); page-content-writer
- **verify-later:** prepare_link_context action; page-build-handler call_content_writer input_mapping

### LNK-018 — Semantic linking domain decomposition (5 link types)
- **status:** partial
- **status-evidence:** Taxonomy table ("Links are not one thing — at least 5 different things") and proposed link-management-group of six agents; "Links live in components, registry is an index" conclusion; lifecycle and semantic agents remain unbuilt.
- **what:** Recognition that link work spans navigation (low complexity), content links/CTAs, semantic links (pillar↔cluster topic modelling — AI-heavy), cross-site/network/affiliate links, and technical links (sitemap/canonical/hreflang), each needing different mechanisms and lifecycles (news decays in days, campaign pages expire, products die). Proposed agent group: navigation-agent, seo-agent, lifecycle-agent, cross-site-agent, semantic-link-agent, link-validator.
- **sources:** docs012_site_maps_and_components/003_semantic_linking.md; docs012_site_maps_and_components/004_more_on_links.md; docs012_site_maps_and_components/006_start_concluding_links.md
- **relations:** LNK-001 (link_registry); relationships table for semantic pairs; LNK-019 (the algorithmic-only subset that was actually specced)
- **verify-later:** which of the six proposed agents exist; page relationships in relationships table

### LNK-019 — Links agent family (algorithmic, no-LLM link health)
- **status:** aspirational
- **status-evidence:** Family table (link-crawler, link-validator, link-registry-sync, redirect-manager, affiliate-link-manager phase 2 — all "LLM? No") with heartbeat workflow and explicit non-goals; still "planned but not implemented" per the current doc 024 (LNK-001).
- **stage2-verified (2026-07-14):** partial → aspirational — grep across .go/.sql for link-crawler, link-validator, link-registry-sync, redirect-manager, affiliate-link-manager, links-orchestrator: 0 hits outside docs/ (all references are in docs024/docs026 notes only). grep for site_redirects table: 0 hits in any .go/.sql. Entire proposed agent family has zero code footprint...
- **what:** Deliberately judgment-free link maintenance: crawl modified pages' HTML, classify by URL pattern, resolve internals to page records, HEAD-check externals rate-limited, detect broken links and orphan pages, generate redirects on URL changes, track per-page link counts and empty anchors. Explicitly excluded: link placement, nav decisions, SEO strategy, related-content suggestions (LLM territory).
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#2-Links-Agent-Family
- **relations:** LNK-001 (link_registry, planned family); LNK-018 (semantic decomposition, the LLM parts deferred); redirect-manager
- **verify-later:** links-orchestrator agent; site_redirects table

### LNK-020 — site_specs `cta` aspect + CTA graph audit (parked)
- **status:** partial
- **status-evidence:** cta aspect inserted 2026-07-02 (primary_url/secondary_url) — un-deferred two sections; CTA-map pass explicitly parked (user chose to leave the circular graph 2026-07-07 until a real arena page exists).
- **what:** A per-site `site_specs` aspect `cta` supplies shared CTA URLs (`cta.primary_url`, `cta.secondary_url`) resolved into component fields (gauntlet-cta.cta_primary_url, system-stats.cta_url) — one populated source fixes all dependants. The vonc CTA graph was then found circular (hero→archive, archive→home, gauntlet-cta→archive; only nav/footer reach the Gauntlet tool, no arena page exists); a deliberate CTA-map pass is queued because CTA URLs are baked into rendered sections, so a proper refresh is a section rebuild, not string surgery.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** LNK-014 (plan_sections deferral, phantom CTA bug); LNK-012 (unresolved_cta, self-resolves when hubs exist); LNK-022 (CTA-graph integrity)
- **verify-later:** site_specs aspect='cta' rows; retarget SQL parked in notes

### LNK-021 — link registry, cached navigation structures, and redirects (link-management foundation)
- **status:** deployed
- **status-evidence:** Live Go code references `link_registry` and `navigation_structures` (html_actions.go, site_db_actions.go, check_phantom_internal_links.go, datahelpers/links.go), and a live doc `024_link_management_v2.md` exists — confirming this MVP schema's core concept shipped and was later versioned.
- **what:** The original link-management schema: `link_registry` indexing every link extracted from rendered components (source component/page/site, resolved target page/site, scope internal/page/site/network/external, link_type navigation/content/semantic/affiliate/reference, validation status); `navigation_structures` as a cached, versioned JSONB nav tree per site+type (header/footer/mobile/sidebar), invalidated by a trigger on any `pages` change and rebuilt lazily via `get_current_navigation`/`build_navigation_for_site`; and a `redirects` table (301/302/307/410, hit_count, expiry). Deliberately reuses the generic `relationships` table for semantic content relationships rather than inventing a parallel structure.
- **sources:** docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql
- **relations:** core client→network→site→page hierarchy (same migration file); LNK-001/024_link_management_v2.md; NAV-010 (nav tables, the successor to navigation_structures)
- **verify-later:** 024_link_management_v2.md — confirm what changed between this v1 schema and "v2"

### LNK-022 — CTA-graph integrity (dead-end and circular primary actions)
- **status:** partial
- **status-evidence:** "Every primary action on the site dead-ends here" (pre-fix, 2026-07-07); "every primary CTA on the site resolves here" (2026-07-09); CTA circularity explicitly "parked, Option B" pending a real arena page.
- **what:** The site's call-to-action graph as an auditable object: for two weeks every primary CTA (nav, hero, gauntlet-cta, lobby cards, provocation-card) pointed at an unbuilt page (404), invisible to any check; after the archive shipped, the graph became circular (hero → archive; archive → home; gauntlet-cta → archive) while the only real interactive surface (the Gauntlet tool) is reachable only via nav/footer. Decision: leave until a real take-filing arena exists. Structural note: CTA URLs are baked into rendered sections, so a graph retarget is a section rebuild, not string surgery; brief-explanation's CTAs still carry '#' placeholders.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-index(4).md#2026-07-07; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#9.5; docs/social001_vonc_tiktok_social/tool_docs/NOTES_lobby-grid(6).md#2026-07-04
- **relations:** NAV-012 (header nav from pages.in_header); LNK-020 (site_specs cta aspect); silent no-op success pattern
- **verify-later:** link_registry; CTA URLs in deployed vonc HTML

### LNK-023 — repairOutboundPageLinks: the shared rerender-path link repair
- **status:** deployed
- **status-evidence:** Live chassis v1.0.1187+ (pod-grep markers); first production run 2026-07-28 repaired 23 dead links across five robot-hands.com pages (agent_error_log rows, `agent_type='page-rerender'`, code CONTENT_LINK_REPAIR_DETAIL); served pages re-probed clean.
- **what:** One shared function (`platform/orchestration/actions/rerender_link_repair.go`) applying the build gate's dead-internal-link repair (`datahelpers.RepairPageLinks` against the `loadValidPagePaths` index) at the last step before a RERENDERED page's HTML leaves for deploy — called by both rerender paths (`RerenderSinglePageAction` per page; the bulk collection loop with one index load per run). Exists because `RepairPageLinks` had exactly one call site (the initial-build validate gate), so pages were protected on first build and unprotected on every rebuild (`bugs_open/097`, diagnosis 9543aaf1). Fail-open on index-load failure, loudly. Outbound-only: DB `rendered_html` keeps the unrepaired form (same philosophy as `StripToolDocHeader`); the durable record is the origin-stamped `writeLinkRepairLog` row — `linkRepairOrigin` names WHICH path repaired, so a drifting sibling is detectable from the log. Note: `RerenderSitePagesAction` (the file's "bulk action") is unregistered dead code; the live bulk mechanism is per-page work items through the single-page action.
- **sources:** bugs_open/097 (2026-07-28 sections); platform/orchestration/actions/rerender_link_repair.go; commit c18f6f430
- **relations:** LNK-008 (shared links.go); LNK-009 (post-deploy audit — detects what this repairs); CQ-002 (the gate whose repair this extends to rerenders)
- **verify-later:** rerender_link_repair.go; agent_error_log rows with agent_type='page-rerender' AND error_code='CONTENT_LINK_REPAIR_DETAIL'

### LNK-024 — repairSectionsBeforePersist: dead-link repair at the PERSISTENCE point
- **status:** built (committed 5083124e3, 2026-07-28; NOT deployed — inert until a chassis image carrying it rolls, and "deployed" would overstate it)
- **status-evidence:** Unit tests pass (`go test ./platform/orchestration/actions/...`, three seam tests); council submission `7c24776e-07f8-4c2e-b1b6-ad3e73c6023c`. Blast radius MEASURED against live config 2026-07-28 before submitting: 6 active non-snapshot agent definitions carry a `save_page_sections` step (page-build-handler, pageflow-builder, page-rebuild, page-rerender, site-work-orchestrator, tool-recreation-handler) and only 2 of the 6 carry any `validate_page_content` step, so 4 of 6 persistence paths had no dead-link repair by any route. `repair_internal_links` appeared in ZERO live `agent_definitions` rows, so the step-config key collides with nothing.
- **what:** `SavePageSectionsAction` now applies `datahelpers.RepairPageLinks` to every section's HTML in place before the insert (`platform/orchestration/actions/save_sections_link_repair.go`: `repairSectionLinks` is the pure seam, `repairSectionsBeforePersist` the DB-touching wrapper). Placed AFTER the interactive-tool preservation block — so stored DB markup carried forward is repaired too — and BEFORE the content-regression/interactivity guards, so those guards measure the bytes actually persisted (unlinking keeps the anchor text, so their tag-stripped totals do not move). **THE LANDMINE THIS EXISTS FOR:** the build gate's repair lands in `clean_html`, and `save_page_sections` reads `html_field` ONLY when the structured `sections_metadata_field` yields zero sections — with `require_sections_metadata: true` on the live page-build plan that branch is unreachable, so the gate's repair was computed, durably logged, and structurally discarded on every build (`bugs_open/079`, REOPENED 2026-07-28; proven on three sites). **Any transformation applied only to `clean_html` has the same defect** — the gate's comment-stripping included. Fail-open throughout: an untrustworthy page index means sections persist untouched plus a `CONTENT_LINK_REPAIR_SKIPPED` record; never a blocked save. Reversal lever `repair_internal_links` (default true, DB config so live-immediately). Error code stays `CONTENT_LINK_REPAIR_DETAIL` — `linkRepairOrigin` discriminates the three paths.
- **open review question:** this is a shared-mechanism change (every pipeline that persists body sections), submitted to the council gate rather than routed to architecture review; whether persistence-point content transforms are the right general pattern, or whether each transform belongs at its own gate, is the question a reviewer should press on.
- **sources:** bugs_open/079 (REOPENED 2026-07-28 banner); docs/agent_docs/docs024_key_docs_latest/bugfix_079_phantom_link_gate/HANDOFF_2026-07-28_platform_fix_candidate1.md; platform/orchestration/actions/save_sections_link_repair.go; commit 5083124e3
- **relations:** LNK-023 (the rerender OUTBOUND seam — deliberately untouched; note its "DB `rendered_html` keeps the unrepaired form" clause now holds only for the outbound assembled page, not for body sections); CQ-002 (the build gate whose repair this rescues from discard); LNK-008 (shared links.go); bugs_open/092 (the upstream cause — the writer never receives its link constraints — still open); bugs_open/117 (stored chrome is never rebuilt, which is why stored markup needed a repair point at all)
- **verify-later:** pod-grep `"SavePageSectionsAction: repaired dead internal links before persist"` after the roll; `agent_error_log` rows with `action='save_page_sections'` AND `error_code='CONTENT_LINK_REPAIR_DETAIL'`
