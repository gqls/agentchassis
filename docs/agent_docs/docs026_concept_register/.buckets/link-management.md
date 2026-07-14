
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Link management: link_registry as first-class links + gap to planned links family
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** 024: schema + extract/sync + constraints + validation exist; links-orchestrator family "planned but not implemented"; delete-and-reinsert loses validation history (known)
- **what:** Every anchor in rendered HTML lives in link_registry (scope internal/page/external, type navigation/content/semantic, affiliate fields, validation state); extract_and_sync_links parses post-build (delete+reinsert per page); InjectLinkConstraints feeds valid pages into writer prompts to prevent invented links; validateInternalLinks warns (not blocks) on missing targets; nav structure is separate (site_nav_groups/items; populate_nav_tables classifies primary/legal/utility). Planned: link-crawler/validator/registry-sync/redirect-manager/affiliate-manager under an algorithmic links-orchestrator.
- **sources:** 024 full
- **relations:** orphan_pages/internal-linker; phantom-CTA bug; nav agent family
- **verify-later:** link_registry population; HTTP validation anywhere?

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Internal linking machinery and its defects
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** "current as of 2026-06-09. Grounded in multipage_actions.go, site_db_actions.go, queryresolve…"; defects: hardcoded fallback nav, unpopulated *_index_url specs, phantom /services.html
- **what:** The pages table (via upsertPage slug/url/nav_label) is the authority for link targets; nav built from real pages or DB nav structure; fixAnchorLinks bridges single-page anchors to multipage URLs; queryresolve fills list-hub cards; "Browse All X" buttons read *_index_url site_specs (inconsistent sources, often empty → href=""); ExtractAndSyncLinksAction maintains a per-page link_registry — the natural substrate for a phantom-link discovery check that does not yet exist. Hero CTA destinations are the linking half of the site-wide CTA defect; whether the CTA href is a resolvable field or hardcoded template is the gating open question.
- **sources:** FOCUS_internal_linking.md (whole)
- **relations:** hardcoded fallbacks; content quality catalogue; section-data reconciler
- **verify-later:** syncLinksToDB (records vs validates); link_registry schema; hero component input_schema

<!-- SOURCE: U05_content_quality_linking.md -->
### Hero CTA brochure-default defect (text↔destination mismatch)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Phantom hero CTA — FIX APPLIED 2026-06-26"; NOTES(44) session 2026-06-26: "snapshot 5946a27b… UPDATE 1; readback confirms".
- **what:** The generic hero/call-to-action component schemas carried brochure-site defaults (`cta_url ← pages.contact`, `secondary_cta_url ← pages.services`) while button text is LLM-written — so every hero site-wide linked "Browse Tools" to /contact.html and to the phantom /services.html. Root causes fixed in layers: the `pages` source fabrication (see next), schema/template hardening (Step 1), and finally the writer's select_sections path mismatch that discarded the resolver's correct hubs.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; HANDOFF_2026-06-09(2).md#next-task; NOTES_gamesdesign_silent_norebuild(44).md (2026-06-22/23 sessions); phantom_hero_ctas/001_context
- **relations:** sourceResolver pages fabrication; Step 1 schema hardening; internal-link-resolver agent; select_sections path-mismatch fix.
- **verify-later:** content_components rows hero/call-to-action input_schema; deployed guide-economy-basics hero HTML; page-content-writer default_config select_sections.

<!-- SOURCE: U05_content_quality_linking.md -->
### sourceResolver `pages` fabrication bug (phantom generator)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1: "Layer 1a … DEPLOYED + APPLIED + VERIFIED … `resolve` case `pages` no longer fabricates".
- **what:** `sourceResolver.resolve` (plan_sections_action.go) `case "pages"` fabricated `"/" + path + ".html"` and returned found=true for any non-existent page, so `on_missing` never fired and schema fallbacks were dead code — the machine that minted every hero phantom. Fixed to return the real URL or (nil,false). Blast radius was tight: hero + call-to-action are the only components with a `pages.*` source.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; running_notes_17(21).md#decisive-findings; step1_hero_cta_phantom_fix.sql (header)
- **relations:** hero CTA defect; Step 1 schema hardening; correct-or-absent principle.
- **verify-later:** platform plan_sections_action.go resolve() pages case.

<!-- SOURCE: U05_content_quality_linking.md -->
### Correct-or-absent principle + loud-but-non-blocking phantom policy
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** FOCUS_internal_linking(1) "Policy (settled this round)" 2026-06-10; running_notes_17(21) "Policy settled".
- **what:** The structural rule for all internal links: targets resolve from the real `pages` set, never fabricated or brochure-assumed; an unresolvable destination renders nothing (no button) rather than a broken/empty link, with a build-time signal so the absence isn't silent. Companion policy: a phantom/missing internal link is loud but non-blocking — a deploy-gate warning, not an error; the improvement loop resolves it.
- **sources:** FOCUS_internal_linking(1).md#through-line; running_notes_17(21).md#policy-settled; PLAN_b4_b5_hubs_and_link_resolver(3).md
- **relations:** unresolved_cta signal; validate_page_content gate; every phantom-fix layer.
- **verify-later:** validate_page_content.go warning severities; hero/CTA template gates in content_components.

<!-- SOURCE: U05_content_quality_linking.md -->
### Step 1 / Layer 1a hero+CTA schema/template hardening
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Verified: both components skip_field/fallbacks_gone/has_and_gate all true".
- **what:** SQL that sets hero/call-to-action CTA-url fields to `on_missing: skip_field`, removes phantom fallbacks (/contact.html, #features), and gates each button template on `{{if and .cta_text .cta_url}}` — so an unresolved CTA renders no button. Ships coupled with the Go resolve() fix (order matters: Go first, else the gate still receives a truthy phantom).
- **sources:** step1_hero_cta_phantom_fix.sql; check_linking_sql_applied.sql; RUNBOOK_linking_phantom_fixes(7).md#1
- **relations:** sourceResolver fabrication fix; internal-link-resolver restores destinations.
- **verify-later:** content_components hero/call-to-action html_template + input_schema; content_components_bak_cta0610 snapshot.

<!-- SOURCE: U05_content_quality_linking.md -->
### Layer 1b header/footer phantom fix (shared site components)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Layer 1b … Verified gone (§4 audit: 0 site_component findings)".
- **what:** Header/footer phantoms (/contact.html, /privacy.html, /terms.html) came from hardcoded ContentData in render_site_components_action.go, not templates or nav fallback. Fix at source: header cta_url resolved from the real contact page, footer legal links data-driven from `GetNavItems(NavGroupLegal)`, header CTA gated on cta_url. Being shared components, the edits benefit every site; nav itself was already real-page-derived.
- **sources:** layer1b_header_footer_phantom_fix.sql; FOCUS_internal_linking(1).md#shipped-this-round; NOTES(44) 2026-06-22 "render_site_components shows the phantom was already fixed for site components"
- **relations:** nav-link-fixer (can't reach ContentData literals); deprecated loadNavItems COALESCE(url,'/name.html') phantom source.
- **verify-later:** render_site_components_action.go lines ~141–233; footer-4-column/header-bold-gradient templates.

<!-- SOURCE: U05_content_quality_linking.md -->
### datahelpers/links.go — canonical link classification library
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** FOCUS_internal_linking(1) "Shared machinery (single source of truth)" — "Replaces three previously-divergent normalisers".
- **what:** One shared library (`ExtractHrefs`, `ClassifyLinkScope`, `IsAssetPath`, `NormalizePagePath`, `PageURLSet`) used by both the deploy gate and the post-deploy audit so they agree by construction. Replaced three divergent URL normalisers (validator lowercased+appended .html; audit stripped index.html; inventory ignored assets).
- **sources:** FOCUS_internal_linking(1).md#shared-machinery; running_notes_17(21).md#shipped
- **relations:** validate_page_content gate; check_phantom_internal_links audit.
- **verify-later:** platform/orchestration/datahelpers/links.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### check_phantom_internal_links post-deploy audit + surface routing
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-15(2) §1: "NOT YET ENABLED (deliberate, observe-only later)"; RUNBOOK_linking_phantom_fixes(7) §7a "gate cleared, ready when you choose".
- **what:** A discovery check scanning page_components + site_components rendered_html for phantom/empty internal links, routing per surface: site_component → nav-link-fixer (build), page_component → page-build-handler (content; a rebuild re-runs build-time resolution). Code-confirmed that per-finding pipeline/handler survive insertWorkItem (config check_pipeline is an unused default). Home agent settled as completeness-discovery-agent (content-integrity family). Deliberately inert until enabled; enabling ≠ autonomous remediation because findings land status='detected' (unclaimable). An earlier version routed page_component findings to internal-link-resolver directly — superseded; a stale duplicate z_context copy with that routing is marked for deletion.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#7a; FOCUS_internal_linking(1).md#shared-machinery; running_notes_17(21).md#§7-gate-RESOLVED; README_find_phantom_links.sql
- **relations:** observe-only enablement pattern; improvement-sweep re-enable gating; nav-link-fixer; internal-link-resolver.
- **verify-later:** discovery_checks/check_phantom_internal_links.go routeBySurface; completeness-discovery-agent run_checks.config.checks array (is phantom_internal_links present?).

<!-- SOURCE: U05_content_quality_linking.md -->
### B4/B5 Browse-All hub links via `section_index_for` queryresolve verb
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "B4/B5 (SQL) … Verified"; running_notes_17(21) 2026-06-14 "Browse All Games → /games/index.html … B4/B5 confirmed".
- **what:** The three list components' "Browse All X" buttons rendered href="" because they sourced `cta_url` from unpopulated, inconsistently-named `*_index_url` site_specs. Fix (option c of three considered): a new queryresolve verb `section_index_for:<type>` deriving the hub URL from real page relationships (shared-area lookup, URL-prefix fallback), plus template gates `{{if .cta_url}}`. Options (a) populate specs and (b) `pages.<hub-name>` source were rejected (per-site maintenance / baked naming convention). Notable trap discovered: for query.* fields the field loop never consults on_missing and would apply `fallback` on nil — hence source-only schema changes and gate-in-template.
- **sources:** PLAN_b4_b5_hubs_and_link_resolver(3).md; b4_b5_hub_links_schema.sql; b4_b5_hub_links_template_gate.sql
- **relations:** correct-or-absent; Tier-D list components; queryresolve subsystem.
- **verify-later:** queryresolve/section_index_for.go + Resolve switch; tool-list/game-list_pre_037/guide-list_pre_037 schemas.

<!-- SOURCE: U05_content_quality_linking.md -->
### internal-link-resolver agent
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Step 3 … Wiring confirmed LIVE"; agent row applied 2026-06-11 per RUNBOOK_linking_phantom_fixes(7).
- **what:** A dedicated sub-agent (spawned by page-content-writer, called once per page) whose single responsibility is resolving intent-appropriate internal link destinations from the real pages set — hero/CTA fields augmented in section resolved_data at build time. v1 rules are deterministic (top content hubs by nav_order, excluding about/contact/legal and the page's own hub); the agent boundary deliberately allows an LLM intent-matching upgrade without changing callers. Explicitly "a build-time augmenter, not a rendered-HTML patcher". Thin workflow (resolve_links → complete), logic in the resolve_internal_links Go action, targets validated via PageURLSet so it cannot emit a URL the gate flags.
- **sources:** internal_link_resolver_agent.sql; PLAN_b4_b5_hubs_and_link_resolver(3).md#step-3; running_notes_17(21).md#step-3
- **relations:** page-content-writer wiring; unresolved_cta signal; ctaFieldNames coverage gap; resolver lean-result follow-up.
- **verify-later:** agent_definitions row type='internal-link-resolver'; resolve_internal_links_action.go (chooseCTATargets, ctaFieldNames, setCTAField).

<!-- SOURCE: U05_content_quality_linking.md -->
### unresolved_cta build-time HITL signal
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) watch SQL "B) resolver distress signal — must stay 0"; observed 0 throughout the §5 batch.
- **what:** When a section has CTA text but no resolvable real-page destination, the resolver emits an `unresolved_cta` work item (needs_human_review; one per affected section, mirroring createDeferredItems, ON CONFLICT dedup). Rationale: the deploy gate cannot see a correctly-dropped button — there is no fingerprint in rendered HTML — so resolution time is the only place the absence is detectable pre-deploy.
- **sources:** PLAN_b4_b5_hubs_and_link_resolver(3).md#unresolved_cta; running_notes_17(21).md#step-3-completed; FOCUS_internal_linking(1).md#remaining
- **relations:** correct-or-absent principle; HITL machinery.
- **verify-later:** ResolveInternalLinksAction unresolved_cta emission; site_work_items item_type='unresolved_cta'.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-content-writer ↔ resolver wiring with regression-safe fallback chain
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** page_content_writer_link_resolver_wiring.sql applied 06-11, "all 7 verification columns correct" (running_notes_17(21) Deployment).
- **what:** Workflow-only wiring: spawn_link_resolver, resolve_links (call_agent, error_step falls through), select_sections (extract_fields with a fallback chain: resolver-augmented sections, else the original plan), loop repointed to sections_for_render. Designed so resolver failure is byte-identical to prior behaviour — which later proved double-edged: the fallback silently masked the path mismatch for two weeks.
- **sources:** page_content_writer_link_resolver_wiring.sql; running_notes_17(21).md#step-3-completed
- **relations:** select_sections path-mismatch bug (the fallback's dark side); result-contract work.
- **verify-later:** page-content-writer default_config workflow steps.

<!-- SOURCE: U05_content_quality_linking.md -->
### select_sections path-mismatch bug (resolver output computed then discarded)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-23 "phantom-CTA ROOT CAUSE CONFIRMED (path mismatch)"; fix applied+confirmed 2026-06-26 (snapshot 5946a27b).
- **what:** The resolver ran and returned augmented sections, but the call_agent envelope nests the reply at `resolved_links.response.link_resolution.sections_ready` while select_sections read top-level `resolved_links.sections_ready` → null → silent fallback to the un-augmented plan carrying the schema phantoms. One-line jsonb_set repoint fixed it; takes effect only on a full content build (a bare re-render doesn't re-run the resolver). Two follow-ups remain open: ctaFieldNames matches only exact "hero"/"call-to-action" (variants like hero-about/gauntlet-cta never resolve), and the resolver returns its whole echoed collected_data with empty final_result (should return a lean {sections_ready, unresolved}).
- **sources:** HANDOFF_page_pipeline(11).md#3; NOTES_gamesdesign_silent_norebuild(44).md 2026-06-23/26; phantom_hero_ctas/001_context
- **relations:** result-contract resolution (the `output` mapping form not flattening is the sibling defect); wiring fallback chain.
- **verify-later:** page-content-writer select_sections config; guide-economy-basics hero has_phantom_cta after build e26cd02f.

<!-- SOURCE: U05_content_quality_linking.md -->
### link_registry — records but never validates (dormant substrate)
- **category:** link-management
- **status-signal:** abandoned
- **status-evidence:** FOCUS_internal_linking(1) finding 2: "syncLinksToDB never populates it … wired into no live workflow, so link_registry is empty. It is not a usable substrate today."
- **what:** A per-page link inventory table with a target_page_id column + FK to pages, intended as a phantom-check substrate, but the sync never populates target_page_id and ExtractAndSyncLinksAction is wired into no live workflow. The phantom audit reads rendered_html directly instead.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; running_notes_16_content_quality_and_internal_linking(1).md#part-1
- **relations:** check_phantom_internal_links (the live approach that superseded it).
- **verify-later:** link_registry row counts; ExtractAndSyncLinksAction callers.

<!-- SOURCE: U05_content_quality_linking.md -->
### nav-link-fixer agent (template-anchor scope only)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** README_find_phantom_links.sql output: nav-link-fixer exists, status experimental, is_active=t.
- **what:** The site_component-surface link fixer: find/replaces `#{{.slug}}`/`#{{.name}}` anchors inside html_template (fix_nav_link_templates_action.go). Its scope excludes ContentData values and literal anchors — which is why the B2/B3 header/footer phantoms had to be fixed at source in Go instead.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings-4; README_find_phantom_links.sql
- **relations:** Layer 1b; check_phantom_internal_links routing (site_component surface).
- **verify-later:** fix_nav_link_templates_action.go; nav-link-fixer agent_definitions row.

<!-- SOURCE: U05_content_quality_linking.md -->
### prepare_link_context available_pages gap on the work-item path
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** running_notes_17(21) watch item 2 (2026-06-12): "This path maps NO db_sync to the writer → prepare_link_context/available_pages get nothing … Pre-existing, resolver-independent."
- **what:** The writer's prepare_link_context builds an available-pages constraint for the LLM's in-prose internal linking, but on the work-item rebuild path no db_sync is mapped, so the constraint text is empty — the LLM writes prose links unconstrained. Independent of the resolver (which queries the DB directly). Candidate fixes noted: map db_sync, or make prepare_link_context load pages itself.
- **sources:** running_notes_17(21).md#page-build-handler-contract watch items
- **relations:** internal-link-resolver; page-content-writer.
- **verify-later:** prepare_link_context action; page-build-handler call_content_writer input_mapping.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Semantic linking domain decomposition (5 link types)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** docs012/003 taxonomy table ("Links are not one thing — at least 5 different things") and proposed link-management-group of six agents; docs012/006 concludes "Links live in components, registry is an index"; lifecycle and semantic agents remain unbuilt.
- **what:** Recognition that link work spans navigation (low complexity), content links/CTAs, semantic links (pillar↔cluster topic modelling — AI-heavy), cross-site/network/affiliate links, and technical links (sitemap/canonical/hreflang), each needing different mechanisms and lifecycles (news decays in days, campaign pages expire, products die). Proposed agent group: navigation-agent, seo-agent, lifecycle-agent, cross-site-agent, semantic-link-agent, link-validator.
- **sources:** docs012_site_maps_and_components/003_semantic_linking.md; docs012_site_maps_and_components/004_more_on_links.md; docs012_site_maps_and_components/006_start_concluding_links.md
- **relations:** link_registry; relationships table for semantic pairs; links agent family (docs017/019b, algorithmic-only subset); current link-management docs 024.
- **verify-later:** which of the six proposed agents exist; page relationships in relationships table.

<!-- SOURCE: U21_legacy_docs_b.md -->
### link_registry as derived index (links live in components)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** docs012/006 schema with scope/link_type/affiliate fields; docs012/012 pipeline step "5e. EXTRACT LINKS — Action: extract_and_sync_links; DB Write: link_registry".
- **what:** Links are never stored as primary data — they exist inside rendered components; link_registry is a queryable index derived by extraction after rendering, tracking source component/page/site, resolved internal targets, scope (internal/page/site/network/external), type (navigation/content/semantic/affiliate/reference), anchor text, rel attributes, affiliate provider/tag, and validation health. Enables broken-link detection, orphan detection, and affiliate compliance without duplicating truth.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.5; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-2; docs012_site_maps_and_components/007_link_migration.sql
- **relations:** links agent family heartbeat; validate_page_content; redirect-manager.
- **verify-later:** link_registry table + extract_and_sync_links action.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Links agent family (algorithmic, no-LLM link health)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** docs017/019b family table (link-crawler, link-validator, link-registry-sync, redirect-manager, affiliate-link-manager phase 2 — all "LLM? No") with heartbeat workflow and explicit non-goals.
- **what:** Deliberately judgment-free link maintenance: crawl modified pages' HTML, classify by URL pattern, resolve internals to page records, HEAD-check externals rate-limited, detect broken links and orphan pages, generate redirects on URL changes, track per-page link counts and empty anchors. Explicitly excluded: link placement, nav decisions, SEO strategy, related-content suggestions (LLM territory).
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#2-Links-Agent-Family
- **relations:** link_registry; semantic linking decomposition (the LLM parts deferred); redirect-manager fix agent.
- **verify-later:** links-orchestrator agent; site_redirects table.

<!-- SOURCE: U23_docs_root_vonc.md -->
### site_specs `cta` aspect + CTA graph audit (parked)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** cta aspect inserted 2026-07-02 (primary_url/secondary_url) — un-deferred two sections; CTA-map pass explicitly PARKED (user chose Option B 2026-07-07: leave the circular graph until the real arena exists).
- **what:** A per-site `site_specs` aspect `cta` supplies shared CTA URLs (`cta.primary_url`, `cta.secondary_url`) resolved into component fields (gauntlet-cta.cta_primary_url, system-stats.cta_url) — one populated source fixes all dependants. The vonc CTA graph was then found CIRCULAR (hero→archive, archive→home, gauntlet-cta→archive; only nav/footer reach the Gauntlet tool, and no arena page exists); a deliberate CTA-map pass is queued because CTA URLs are baked into rendered sections, so a proper refresh is a section rebuild, not string surgery.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:35 + #2026-07-02-~19:50; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-step-4-done + #2026-07-07-~16:30; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** plan_sections deferral; phantom CTA bug; unresolved_cta work items (self-resolve when hubs exist)
- **verify-later:** site_specs aspect='cta' rows; retarget SQL parked in notes

<!-- SOURCE: U23_docs_root_vonc.md -->
### Phantom CTA resolution bug (fabricated /{area}.html hero CTAs)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** 016b Part 4 (confirmed 2026-06-22 in deployed HTML, gamesdesign): hero carries two phantom CTAs from schema sources pages.contact/pages.services; "workflow-only fix staged" (select_sections reading resolved_links at the wrong path).
- **what:** Hero CTA resolution can produce constructed/fabricated URLs (`/contact.html`, `/services.html`) while the real hubs live elsewhere, because `select_sections` reads `resolved_links.sections_ready` (null) instead of `resolved_links.response.link_resolution.sections_ready`, falling back to the un-augmented plan; `resolve_internal_links` is a build-time augmenter (writes cta_url into resolved_data for the writer), explicitly not a rendered-HTML patcher, and `check_phantom_internal_links` routes page-link fixes to page-build-handler by design. Distinct from the interactive clobber; `page_rerender` does not re-resolve schema-sourced CTAs (ruled out as a link fix).
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 4 + update); docs/RUNBOOK_vonc_session(1).md#remaining-steps (unresolved_cta parking)
- **relations:** site_specs cta aspect; internal link management (024)
- **verify-later:** select_sections workflow path fix; resolve_internal_links action

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Hero/CTA link fabrication in `sourceResolver.resolve` — the "310–318 hardcoded fallback nav" hypothesis superseded
- **category:** link-management
- **status-signal:** superseded
- **status-evidence:** Archived `FOCUS_internal_linking.md` (2026-06-09): "**Defect (lines 310–318 of `multipage_actions.go`):** when nav resolution returns empty, `AssembleMultipageSiteAction` injects a **hardcoded fallback nav**... This generic brochure default is a primary source of the phantom `/services.html`." Live `FOCUS_internal_linking(1).md` (2026-06-10): "**Nav is already real-page-derived; the brochure fallback was not the live path.**... The header/footer phantoms... came from **hardcoded `ContentData` in `render_site_components_action.go`**... not from the `multipage_actions.go` 310–318 fallback nav... (**Correction to the earlier note that blamed 310–318**.)"
- **what:** The initial (2026-06-09) diagnosis of site-wide phantom links blamed a specific hardcoded fallback-nav code path (`multipage_actions.go:310-318`) as the likely root cause. The next day's investigation, grounded in reads of `render_site_components_action.go`, corrected this: nav was already correctly real-page-derived, and the actual mechanism was (a) `sourceResolver.resolve`'s `"pages"` case *fabricating* a URL (`"/"+path+".html"`) and returning `found=true` for any non-existent page (so schema `on_missing`/`fallback` never fired), plus (b) separately, hardcoded `ContentData` literals for header/footer CTAs and legal links.
- **sources:** content_quality_and_internal_linking/FOCUS_internal_linking.md (archived, 2026-06-09); live FOCUS_internal_linking(1).md (2026-06-10); running_notes_17(16) "Decisive findings"
- **relations:** component-template-fixer CTA-reuse assumption; link_registry hypothesis (below)
- **verify-later:** `plan_sections_action.go` `sourceResolver.resolve` current "pages" case; `render_site_components_action.go` ContentData construction.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `link_registry` as a phantom-link validation substrate — considered, found unusable, abandoned
- **category:** link-management
- **status-signal:** abandoned
- **status-evidence:** Archived `FOCUS_internal_linking.md` (2026-06-09): "**Link inventory.** `ExtractAndSyncLinksAction`... syncs them per page into `link_registry`... A per-page link inventory already exists — the natural substrate for a broken/phantom-link discovery check." Live `FOCUS_internal_linking(1).md` (2026-06-10): "**`link_registry` only records, never validates.** It *has* a `target_page_id` column + FK to `pages`, but `syncLinksToDB` never populates it. And `extract_and_sync_links` is wired into **no live workflow**, so `link_registry` is empty. It is not a usable substrate today."
- **what:** The internal-linking investigation initially proposed reusing the existing `link_registry` table/action as the base for a new phantom-link discovery check. Follow-up code reading found the table permanently empty in practice (the populating column is never written, and the syncing action isn't wired into any live workflow) — so the check that was actually built (`check_phantom_internal_links.go`) instead scans `rendered_html` directly via new shared helpers (`datahelpers/links.go`).
- **sources:** content_quality_and_internal_linking/FOCUS_internal_linking.md (archived); live FOCUS_internal_linking(1).md; running_notes_17(16) "Decisive findings"
- **relations:** hero/CTA link fabrication (above)
- **verify-later:** whether `ExtractAndSyncLinksAction`/`link_registry` were ever wired up subsequently.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Hub "Browse All X" link resolution — rejected design options
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** `PLAN_b4_b5_hubs_and_link_resolver(1).md`: "(a) Populate the `*_index_url` specs from real hubs. **Rejected** — per-site data to maintain; re-introduces the inconsistent-source brittleness. (b) `source: pages.<hub-name>` per component... bakes the `<area>-index` naming convention into each schema... (c) **Recommended.** A new `queryresolve` verb... `query.section_index_for:<type>`." Shipped per running_notes_17(16): "`section_index_for.go` — new `queryresolve` verb... B4/B5 — done."
- **what:** For the empty-href "Browse All Tools/Games/Guides" defect, two options were explicitly weighed and rejected in the design doc before settling on a new `queryresolve` verb: manually populating `*_index_url` site_specs (rejected as brittle, per-site maintenance) and a per-component `pages.<hub-name>` source (rejected as baking a naming convention into every schema). The chosen option — `query.section_index_for:<type>`, resolving the hub via shared `site_area_id`/URL-prefix — shipped and was confirmed working in deployed HTML.
- **sources:** content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md; running_notes_17(16)
- **relations:** hero/CTA link fabrication; internal-link-resolver agent
- **verify-later:** `queryresolve.go` `section_index_for` case in current code.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `internal-link-resolver` agent (Step 3) — dedicated intent-aware internal-link resolution
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** running_notes_17(16): "**Step 3 — completed (2026-06-11), all deliverables written**... Agent row (`internal_link_resolver_agent.sql`)... Writer wiring (`page_content_writer_link_resolver_wiring.sql`)... `unresolved_cta`: emitted in-Go... `status needs_human_review`." Confirmed live end-to-end: "Query D (corrected paths) on both completed rebuilds: `for_render=2`, `plan_count=2`, EQUAL ⇒ resolver augmented sections + writer loop consumed them."
- **what:** A new sub-agent of `page-content-writer` (modelled on `research-agent`, no persistence) that, at build time, resolves hero/CTA link destinations to intent-appropriate real pages (excluding the page's own hub, about/contact/legal) rather than a fixed contact page, validates every candidate against `datahelpers.PageURLSet`, and emits an `unresolved_cta` HITL signal when no destination can be found — the only place a "correctly dropped" (absent) button is detectable, since the deploy gate can't see an absence. Replaces the abandoned assumption that `component-template-fixer` already handled this.
- **sources:** content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md; running_notes_17(16) "Step 3" sections
- **relations:** component-template-fixer CTA-reuse assumption (superseded); identity-advisor/approval_mode (abandoned); hero/CTA fabrication fix
- **verify-later:** `internal_link_resolver_agent.sql`, `resolve_internal_links_action.go` current deployment/image-tag state.

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### link registry, cached navigation structures, and redirects (link-management foundation)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** Live Go code references `link_registry` and `navigation_structures` (e.g. platform/orchestration/actions/html_actions.go, site_db_actions.go, discovery_checks/check_phantom_internal_links.go, platform/orchestration/datahelpers/links.go), and a live doc `024_link_management_v2.md` exists — confirming this MVP schema's core concept shipped and was later versioned.
- **what:** The original link-management schema: a `link_registry` table indexing every link extracted from rendered components (source component/page/site, resolved target page/site, a `scope` of internal/page/site/network/external, a `link_type` of navigation/content/semantic/affiliate/reference, plus validation status for broken-link detection); `navigation_structures` as a **cached, versioned** JSONB nav tree per site+type (header/footer/mobile/sidebar), invalidated by a trigger on any `pages` INSERT/UPDATE/DELETE and rebuilt lazily via `get_current_navigation`/`build_navigation_for_site`; and a `redirects` table (301/302/307/410, hit_count, expiry). Deliberately reuses the existing generic `relationships` table for semantic content relationships (pillar/cluster, related-content, cross-site-reference) rather than inventing a parallel structure.
- **sources:** docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql
- **relations:** core client→network→site→page hierarchy (above, same migration file); link-management (024 anchor, 024_link_management_v2.md)
- **verify-later:** 024_link_management_v2.md — confirm what changed between this v1 schema and "v2"

<!-- SOURCE: U25_leopardess_social.md -->
### CTA-graph integrity (dead-end and circular primary actions)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** NOTES_provocations-index 2026-07-07: "Every primary action on the site dead-ends here" (pre-fix); 2026-07-09 "every primary CTA on the site resolves here"; CTA circularity "parked, Option B" pending a real arena page.
- **what:** The site's call-to-action graph as an auditable object: for two weeks every primary CTA (nav, hero, gauntlet-cta, lobby cards, provocation-card) pointed at an unbuilt page (404), invisible to any check; after the archive shipped, the graph is circular (hero → archive; archive → home; gauntlet-cta → archive) while the only real interactive surface (the Gauntlet tool) is reachable only via nav/footer. Decision Option B: leave until a real take-filing arena exists. Structural note: CTA URLs are baked into rendered sections, so a graph retarget is a section rebuild, not string surgery; brief-explanation's CTAs still carry '#' placeholders.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-index(4).md#2026-07-07; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#9.5; docs/social001_vonc_tiktok_social/tool_docs/NOTES_lobby-grid(6).md#2026-07-04
- **relations:** navigation; silent no-op success (the 404 destination was its product); link_registry
- **verify-later:** link_registry; CTA URLs in deployed vonc HTML

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Link management: link_registry as first-class links + gap to planned links family
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** 024: schema + extract/sync + constraints + validation exist; links-orchestrator family "planned but not implemented"; delete-and-reinsert loses validation history (known)
- **what:** Every anchor in rendered HTML lives in link_registry (scope internal/page/external, type navigation/content/semantic, affiliate fields, validation state); extract_and_sync_links parses post-build (delete+reinsert per page); InjectLinkConstraints feeds valid pages into writer prompts to prevent invented links; validateInternalLinks warns (not blocks) on missing targets; nav structure is separate (site_nav_groups/items; populate_nav_tables classifies primary/legal/utility). Planned: link-crawler/validator/registry-sync/redirect-manager/affiliate-manager under an algorithmic links-orchestrator.
- **sources:** 024 full
- **relations:** orphan_pages/internal-linker; phantom-CTA bug; nav agent family
- **verify-later:** link_registry population; HTTP validation anywhere?

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Internal linking machinery and its defects
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** "current as of 2026-06-09. Grounded in multipage_actions.go, site_db_actions.go, queryresolve…"; defects: hardcoded fallback nav, unpopulated *_index_url specs, phantom /services.html
- **what:** The pages table (via upsertPage slug/url/nav_label) is the authority for link targets; nav built from real pages or DB nav structure; fixAnchorLinks bridges single-page anchors to multipage URLs; queryresolve fills list-hub cards; "Browse All X" buttons read *_index_url site_specs (inconsistent sources, often empty → href=""); ExtractAndSyncLinksAction maintains a per-page link_registry — the natural substrate for a phantom-link discovery check that does not yet exist. Hero CTA destinations are the linking half of the site-wide CTA defect; whether the CTA href is a resolvable field or hardcoded template is the gating open question.
- **sources:** FOCUS_internal_linking.md (whole)
- **relations:** hardcoded fallbacks; content quality catalogue; section-data reconciler
- **verify-later:** syncLinksToDB (records vs validates); link_registry schema; hero component input_schema

<!-- SOURCE: U05_content_quality_linking.md -->
### Hero CTA brochure-default defect (text↔destination mismatch)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Phantom hero CTA — FIX APPLIED 2026-06-26"; NOTES(44) session 2026-06-26: "snapshot 5946a27b… UPDATE 1; readback confirms".
- **what:** The generic hero/call-to-action component schemas carried brochure-site defaults (`cta_url ← pages.contact`, `secondary_cta_url ← pages.services`) while button text is LLM-written — so every hero site-wide linked "Browse Tools" to /contact.html and to the phantom /services.html. Root causes fixed in layers: the `pages` source fabrication (see next), schema/template hardening (Step 1), and finally the writer's select_sections path mismatch that discarded the resolver's correct hubs.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; HANDOFF_2026-06-09(2).md#next-task; NOTES_gamesdesign_silent_norebuild(44).md (2026-06-22/23 sessions); phantom_hero_ctas/001_context
- **relations:** sourceResolver pages fabrication; Step 1 schema hardening; internal-link-resolver agent; select_sections path-mismatch fix.
- **verify-later:** content_components rows hero/call-to-action input_schema; deployed guide-economy-basics hero HTML; page-content-writer default_config select_sections.

<!-- SOURCE: U05_content_quality_linking.md -->
### sourceResolver `pages` fabrication bug (phantom generator)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1: "Layer 1a … DEPLOYED + APPLIED + VERIFIED … `resolve` case `pages` no longer fabricates".
- **what:** `sourceResolver.resolve` (plan_sections_action.go) `case "pages"` fabricated `"/" + path + ".html"` and returned found=true for any non-existent page, so `on_missing` never fired and schema fallbacks were dead code — the machine that minted every hero phantom. Fixed to return the real URL or (nil,false). Blast radius was tight: hero + call-to-action are the only components with a `pages.*` source.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; running_notes_17(21).md#decisive-findings; step1_hero_cta_phantom_fix.sql (header)
- **relations:** hero CTA defect; Step 1 schema hardening; correct-or-absent principle.
- **verify-later:** platform plan_sections_action.go resolve() pages case.

<!-- SOURCE: U05_content_quality_linking.md -->
### Correct-or-absent principle + loud-but-non-blocking phantom policy
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** FOCUS_internal_linking(1) "Policy (settled this round)" 2026-06-10; running_notes_17(21) "Policy settled".
- **what:** The structural rule for all internal links: targets resolve from the real `pages` set, never fabricated or brochure-assumed; an unresolvable destination renders nothing (no button) rather than a broken/empty link, with a build-time signal so the absence isn't silent. Companion policy: a phantom/missing internal link is loud but non-blocking — a deploy-gate warning, not an error; the improvement loop resolves it.
- **sources:** FOCUS_internal_linking(1).md#through-line; running_notes_17(21).md#policy-settled; PLAN_b4_b5_hubs_and_link_resolver(3).md
- **relations:** unresolved_cta signal; validate_page_content gate; every phantom-fix layer.
- **verify-later:** validate_page_content.go warning severities; hero/CTA template gates in content_components.

<!-- SOURCE: U05_content_quality_linking.md -->
### Step 1 / Layer 1a hero+CTA schema/template hardening
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Verified: both components skip_field/fallbacks_gone/has_and_gate all true".
- **what:** SQL that sets hero/call-to-action CTA-url fields to `on_missing: skip_field`, removes phantom fallbacks (/contact.html, #features), and gates each button template on `{{if and .cta_text .cta_url}}` — so an unresolved CTA renders no button. Ships coupled with the Go resolve() fix (order matters: Go first, else the gate still receives a truthy phantom).
- **sources:** step1_hero_cta_phantom_fix.sql; check_linking_sql_applied.sql; RUNBOOK_linking_phantom_fixes(7).md#1
- **relations:** sourceResolver fabrication fix; internal-link-resolver restores destinations.
- **verify-later:** content_components hero/call-to-action html_template + input_schema; content_components_bak_cta0610 snapshot.

<!-- SOURCE: U05_content_quality_linking.md -->
### Layer 1b header/footer phantom fix (shared site components)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Layer 1b … Verified gone (§4 audit: 0 site_component findings)".
- **what:** Header/footer phantoms (/contact.html, /privacy.html, /terms.html) came from hardcoded ContentData in render_site_components_action.go, not templates or nav fallback. Fix at source: header cta_url resolved from the real contact page, footer legal links data-driven from `GetNavItems(NavGroupLegal)`, header CTA gated on cta_url. Being shared components, the edits benefit every site; nav itself was already real-page-derived.
- **sources:** layer1b_header_footer_phantom_fix.sql; FOCUS_internal_linking(1).md#shipped-this-round; NOTES(44) 2026-06-22 "render_site_components shows the phantom was already fixed for site components"
- **relations:** nav-link-fixer (can't reach ContentData literals); deprecated loadNavItems COALESCE(url,'/name.html') phantom source.
- **verify-later:** render_site_components_action.go lines ~141–233; footer-4-column/header-bold-gradient templates.

<!-- SOURCE: U05_content_quality_linking.md -->
### datahelpers/links.go — canonical link classification library
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** FOCUS_internal_linking(1) "Shared machinery (single source of truth)" — "Replaces three previously-divergent normalisers".
- **what:** One shared library (`ExtractHrefs`, `ClassifyLinkScope`, `IsAssetPath`, `NormalizePagePath`, `PageURLSet`) used by both the deploy gate and the post-deploy audit so they agree by construction. Replaced three divergent URL normalisers (validator lowercased+appended .html; audit stripped index.html; inventory ignored assets).
- **sources:** FOCUS_internal_linking(1).md#shared-machinery; running_notes_17(21).md#shipped
- **relations:** validate_page_content gate; check_phantom_internal_links audit.
- **verify-later:** platform/orchestration/datahelpers/links.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### check_phantom_internal_links post-deploy audit + surface routing
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-15(2) §1: "NOT YET ENABLED (deliberate, observe-only later)"; RUNBOOK_linking_phantom_fixes(7) §7a "gate cleared, ready when you choose".
- **what:** A discovery check scanning page_components + site_components rendered_html for phantom/empty internal links, routing per surface: site_component → nav-link-fixer (build), page_component → page-build-handler (content; a rebuild re-runs build-time resolution). Code-confirmed that per-finding pipeline/handler survive insertWorkItem (config check_pipeline is an unused default). Home agent settled as completeness-discovery-agent (content-integrity family). Deliberately inert until enabled; enabling ≠ autonomous remediation because findings land status='detected' (unclaimable). An earlier version routed page_component findings to internal-link-resolver directly — superseded; a stale duplicate z_context copy with that routing is marked for deletion.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#7a; FOCUS_internal_linking(1).md#shared-machinery; running_notes_17(21).md#§7-gate-RESOLVED; README_find_phantom_links.sql
- **relations:** observe-only enablement pattern; improvement-sweep re-enable gating; nav-link-fixer; internal-link-resolver.
- **verify-later:** discovery_checks/check_phantom_internal_links.go routeBySurface; completeness-discovery-agent run_checks.config.checks array (is phantom_internal_links present?).

<!-- SOURCE: U05_content_quality_linking.md -->
### B4/B5 Browse-All hub links via `section_index_for` queryresolve verb
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "B4/B5 (SQL) … Verified"; running_notes_17(21) 2026-06-14 "Browse All Games → /games/index.html … B4/B5 confirmed".
- **what:** The three list components' "Browse All X" buttons rendered href="" because they sourced `cta_url` from unpopulated, inconsistently-named `*_index_url` site_specs. Fix (option c of three considered): a new queryresolve verb `section_index_for:<type>` deriving the hub URL from real page relationships (shared-area lookup, URL-prefix fallback), plus template gates `{{if .cta_url}}`. Options (a) populate specs and (b) `pages.<hub-name>` source were rejected (per-site maintenance / baked naming convention). Notable trap discovered: for query.* fields the field loop never consults on_missing and would apply `fallback` on nil — hence source-only schema changes and gate-in-template.
- **sources:** PLAN_b4_b5_hubs_and_link_resolver(3).md; b4_b5_hub_links_schema.sql; b4_b5_hub_links_template_gate.sql
- **relations:** correct-or-absent; Tier-D list components; queryresolve subsystem.
- **verify-later:** queryresolve/section_index_for.go + Resolve switch; tool-list/game-list_pre_037/guide-list_pre_037 schemas.

<!-- SOURCE: U05_content_quality_linking.md -->
### internal-link-resolver agent
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-15(2) §1 "Step 3 … Wiring confirmed LIVE"; agent row applied 2026-06-11 per RUNBOOK_linking_phantom_fixes(7).
- **what:** A dedicated sub-agent (spawned by page-content-writer, called once per page) whose single responsibility is resolving intent-appropriate internal link destinations from the real pages set — hero/CTA fields augmented in section resolved_data at build time. v1 rules are deterministic (top content hubs by nav_order, excluding about/contact/legal and the page's own hub); the agent boundary deliberately allows an LLM intent-matching upgrade without changing callers. Explicitly "a build-time augmenter, not a rendered-HTML patcher". Thin workflow (resolve_links → complete), logic in the resolve_internal_links Go action, targets validated via PageURLSet so it cannot emit a URL the gate flags.
- **sources:** internal_link_resolver_agent.sql; PLAN_b4_b5_hubs_and_link_resolver(3).md#step-3; running_notes_17(21).md#step-3
- **relations:** page-content-writer wiring; unresolved_cta signal; ctaFieldNames coverage gap; resolver lean-result follow-up.
- **verify-later:** agent_definitions row type='internal-link-resolver'; resolve_internal_links_action.go (chooseCTATargets, ctaFieldNames, setCTAField).

<!-- SOURCE: U05_content_quality_linking.md -->
### unresolved_cta build-time HITL signal
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) watch SQL "B) resolver distress signal — must stay 0"; observed 0 throughout the §5 batch.
- **what:** When a section has CTA text but no resolvable real-page destination, the resolver emits an `unresolved_cta` work item (needs_human_review; one per affected section, mirroring createDeferredItems, ON CONFLICT dedup). Rationale: the deploy gate cannot see a correctly-dropped button — there is no fingerprint in rendered HTML — so resolution time is the only place the absence is detectable pre-deploy.
- **sources:** PLAN_b4_b5_hubs_and_link_resolver(3).md#unresolved_cta; running_notes_17(21).md#step-3-completed; FOCUS_internal_linking(1).md#remaining
- **relations:** correct-or-absent principle; HITL machinery.
- **verify-later:** ResolveInternalLinksAction unresolved_cta emission; site_work_items item_type='unresolved_cta'.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-content-writer ↔ resolver wiring with regression-safe fallback chain
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** page_content_writer_link_resolver_wiring.sql applied 06-11, "all 7 verification columns correct" (running_notes_17(21) Deployment).
- **what:** Workflow-only wiring: spawn_link_resolver, resolve_links (call_agent, error_step falls through), select_sections (extract_fields with a fallback chain: resolver-augmented sections, else the original plan), loop repointed to sections_for_render. Designed so resolver failure is byte-identical to prior behaviour — which later proved double-edged: the fallback silently masked the path mismatch for two weeks.
- **sources:** page_content_writer_link_resolver_wiring.sql; running_notes_17(21).md#step-3-completed
- **relations:** select_sections path-mismatch bug (the fallback's dark side); result-contract work.
- **verify-later:** page-content-writer default_config workflow steps.

<!-- SOURCE: U05_content_quality_linking.md -->
### select_sections path-mismatch bug (resolver output computed then discarded)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-23 "phantom-CTA ROOT CAUSE CONFIRMED (path mismatch)"; fix applied+confirmed 2026-06-26 (snapshot 5946a27b).
- **what:** The resolver ran and returned augmented sections, but the call_agent envelope nests the reply at `resolved_links.response.link_resolution.sections_ready` while select_sections read top-level `resolved_links.sections_ready` → null → silent fallback to the un-augmented plan carrying the schema phantoms. One-line jsonb_set repoint fixed it; takes effect only on a full content build (a bare re-render doesn't re-run the resolver). Two follow-ups remain open: ctaFieldNames matches only exact "hero"/"call-to-action" (variants like hero-about/gauntlet-cta never resolve), and the resolver returns its whole echoed collected_data with empty final_result (should return a lean {sections_ready, unresolved}).
- **sources:** HANDOFF_page_pipeline(11).md#3; NOTES_gamesdesign_silent_norebuild(44).md 2026-06-23/26; phantom_hero_ctas/001_context
- **relations:** result-contract resolution (the `output` mapping form not flattening is the sibling defect); wiring fallback chain.
- **verify-later:** page-content-writer select_sections config; guide-economy-basics hero has_phantom_cta after build e26cd02f.

<!-- SOURCE: U05_content_quality_linking.md -->
### link_registry — records but never validates (dormant substrate)
- **category:** link-management
- **status-signal:** abandoned
- **status-evidence:** FOCUS_internal_linking(1) finding 2: "syncLinksToDB never populates it … wired into no live workflow, so link_registry is empty. It is not a usable substrate today."
- **what:** A per-page link inventory table with a target_page_id column + FK to pages, intended as a phantom-check substrate, but the sync never populates target_page_id and ExtractAndSyncLinksAction is wired into no live workflow. The phantom audit reads rendered_html directly instead.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings; running_notes_16_content_quality_and_internal_linking(1).md#part-1
- **relations:** check_phantom_internal_links (the live approach that superseded it).
- **verify-later:** link_registry row counts; ExtractAndSyncLinksAction callers.

<!-- SOURCE: U05_content_quality_linking.md -->
### nav-link-fixer agent (template-anchor scope only)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** README_find_phantom_links.sql output: nav-link-fixer exists, status experimental, is_active=t.
- **what:** The site_component-surface link fixer: find/replaces `#{{.slug}}`/`#{{.name}}` anchors inside html_template (fix_nav_link_templates_action.go). Its scope excludes ContentData values and literal anchors — which is why the B2/B3 header/footer phantoms had to be fixed at source in Go instead.
- **sources:** FOCUS_internal_linking(1).md#decisive-findings-4; README_find_phantom_links.sql
- **relations:** Layer 1b; check_phantom_internal_links routing (site_component surface).
- **verify-later:** fix_nav_link_templates_action.go; nav-link-fixer agent_definitions row.

<!-- SOURCE: U05_content_quality_linking.md -->
### prepare_link_context available_pages gap on the work-item path
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** running_notes_17(21) watch item 2 (2026-06-12): "This path maps NO db_sync to the writer → prepare_link_context/available_pages get nothing … Pre-existing, resolver-independent."
- **what:** The writer's prepare_link_context builds an available-pages constraint for the LLM's in-prose internal linking, but on the work-item rebuild path no db_sync is mapped, so the constraint text is empty — the LLM writes prose links unconstrained. Independent of the resolver (which queries the DB directly). Candidate fixes noted: map db_sync, or make prepare_link_context load pages itself.
- **sources:** running_notes_17(21).md#page-build-handler-contract watch items
- **relations:** internal-link-resolver; page-content-writer.
- **verify-later:** prepare_link_context action; page-build-handler call_content_writer input_mapping.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Semantic linking domain decomposition (5 link types)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** docs012/003 taxonomy table ("Links are not one thing — at least 5 different things") and proposed link-management-group of six agents; docs012/006 concludes "Links live in components, registry is an index"; lifecycle and semantic agents remain unbuilt.
- **what:** Recognition that link work spans navigation (low complexity), content links/CTAs, semantic links (pillar↔cluster topic modelling — AI-heavy), cross-site/network/affiliate links, and technical links (sitemap/canonical/hreflang), each needing different mechanisms and lifecycles (news decays in days, campaign pages expire, products die). Proposed agent group: navigation-agent, seo-agent, lifecycle-agent, cross-site-agent, semantic-link-agent, link-validator.
- **sources:** docs012_site_maps_and_components/003_semantic_linking.md; docs012_site_maps_and_components/004_more_on_links.md; docs012_site_maps_and_components/006_start_concluding_links.md
- **relations:** link_registry; relationships table for semantic pairs; links agent family (docs017/019b, algorithmic-only subset); current link-management docs 024.
- **verify-later:** which of the six proposed agents exist; page relationships in relationships table.

<!-- SOURCE: U21_legacy_docs_b.md -->
### link_registry as derived index (links live in components)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** docs012/006 schema with scope/link_type/affiliate fields; docs012/012 pipeline step "5e. EXTRACT LINKS — Action: extract_and_sync_links; DB Write: link_registry".
- **what:** Links are never stored as primary data — they exist inside rendered components; link_registry is a queryable index derived by extraction after rendering, tracking source component/page/site, resolved internal targets, scope (internal/page/site/network/external), type (navigation/content/semantic/affiliate/reference), anchor text, rel attributes, affiliate provider/tag, and validation health. Enables broken-link detection, orphan detection, and affiliate compliance without duplicating truth.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.5; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-2; docs012_site_maps_and_components/007_link_migration.sql
- **relations:** links agent family heartbeat; validate_page_content; redirect-manager.
- **verify-later:** link_registry table + extract_and_sync_links action.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Links agent family (algorithmic, no-LLM link health)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** docs017/019b family table (link-crawler, link-validator, link-registry-sync, redirect-manager, affiliate-link-manager phase 2 — all "LLM? No") with heartbeat workflow and explicit non-goals.
- **what:** Deliberately judgment-free link maintenance: crawl modified pages' HTML, classify by URL pattern, resolve internals to page records, HEAD-check externals rate-limited, detect broken links and orphan pages, generate redirects on URL changes, track per-page link counts and empty anchors. Explicitly excluded: link placement, nav decisions, SEO strategy, related-content suggestions (LLM territory).
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#2-Links-Agent-Family
- **relations:** link_registry; semantic linking decomposition (the LLM parts deferred); redirect-manager fix agent.
- **verify-later:** links-orchestrator agent; site_redirects table.

<!-- SOURCE: U23_docs_root_vonc.md -->
### site_specs `cta` aspect + CTA graph audit (parked)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** cta aspect inserted 2026-07-02 (primary_url/secondary_url) — un-deferred two sections; CTA-map pass explicitly PARKED (user chose Option B 2026-07-07: leave the circular graph until the real arena exists).
- **what:** A per-site `site_specs` aspect `cta` supplies shared CTA URLs (`cta.primary_url`, `cta.secondary_url`) resolved into component fields (gauntlet-cta.cta_primary_url, system-stats.cta_url) — one populated source fixes all dependants. The vonc CTA graph was then found CIRCULAR (hero→archive, archive→home, gauntlet-cta→archive; only nav/footer reach the Gauntlet tool, and no arena page exists); a deliberate CTA-map pass is queued because CTA URLs are baked into rendered sections, so a proper refresh is a section rebuild, not string surgery.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:35 + #2026-07-02-~19:50; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-step-4-done + #2026-07-07-~16:30; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** plan_sections deferral; phantom CTA bug; unresolved_cta work items (self-resolve when hubs exist)
- **verify-later:** site_specs aspect='cta' rows; retarget SQL parked in notes

<!-- SOURCE: U23_docs_root_vonc.md -->
### Phantom CTA resolution bug (fabricated /{area}.html hero CTAs)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** 016b Part 4 (confirmed 2026-06-22 in deployed HTML, gamesdesign): hero carries two phantom CTAs from schema sources pages.contact/pages.services; "workflow-only fix staged" (select_sections reading resolved_links at the wrong path).
- **what:** Hero CTA resolution can produce constructed/fabricated URLs (`/contact.html`, `/services.html`) while the real hubs live elsewhere, because `select_sections` reads `resolved_links.sections_ready` (null) instead of `resolved_links.response.link_resolution.sections_ready`, falling back to the un-augmented plan; `resolve_internal_links` is a build-time augmenter (writes cta_url into resolved_data for the writer), explicitly not a rendered-HTML patcher, and `check_phantom_internal_links` routes page-link fixes to page-build-handler by design. Distinct from the interactive clobber; `page_rerender` does not re-resolve schema-sourced CTAs (ruled out as a link fix).
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 4 + update); docs/RUNBOOK_vonc_session(1).md#remaining-steps (unresolved_cta parking)
- **relations:** site_specs cta aspect; internal link management (024)
- **verify-later:** select_sections workflow path fix; resolve_internal_links action

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Hero/CTA link fabrication in `sourceResolver.resolve` — the "310–318 hardcoded fallback nav" hypothesis superseded
- **category:** link-management
- **status-signal:** superseded
- **status-evidence:** Archived `FOCUS_internal_linking.md` (2026-06-09): "**Defect (lines 310–318 of `multipage_actions.go`):** when nav resolution returns empty, `AssembleMultipageSiteAction` injects a **hardcoded fallback nav**... This generic brochure default is a primary source of the phantom `/services.html`." Live `FOCUS_internal_linking(1).md` (2026-06-10): "**Nav is already real-page-derived; the brochure fallback was not the live path.**... The header/footer phantoms... came from **hardcoded `ContentData` in `render_site_components_action.go`**... not from the `multipage_actions.go` 310–318 fallback nav... (**Correction to the earlier note that blamed 310–318**.)"
- **what:** The initial (2026-06-09) diagnosis of site-wide phantom links blamed a specific hardcoded fallback-nav code path (`multipage_actions.go:310-318`) as the likely root cause. The next day's investigation, grounded in reads of `render_site_components_action.go`, corrected this: nav was already correctly real-page-derived, and the actual mechanism was (a) `sourceResolver.resolve`'s `"pages"` case *fabricating* a URL (`"/"+path+".html"`) and returning `found=true` for any non-existent page (so schema `on_missing`/`fallback` never fired), plus (b) separately, hardcoded `ContentData` literals for header/footer CTAs and legal links.
- **sources:** content_quality_and_internal_linking/FOCUS_internal_linking.md (archived, 2026-06-09); live FOCUS_internal_linking(1).md (2026-06-10); running_notes_17(16) "Decisive findings"
- **relations:** component-template-fixer CTA-reuse assumption; link_registry hypothesis (below)
- **verify-later:** `plan_sections_action.go` `sourceResolver.resolve` current "pages" case; `render_site_components_action.go` ContentData construction.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `link_registry` as a phantom-link validation substrate — considered, found unusable, abandoned
- **category:** link-management
- **status-signal:** abandoned
- **status-evidence:** Archived `FOCUS_internal_linking.md` (2026-06-09): "**Link inventory.** `ExtractAndSyncLinksAction`... syncs them per page into `link_registry`... A per-page link inventory already exists — the natural substrate for a broken/phantom-link discovery check." Live `FOCUS_internal_linking(1).md` (2026-06-10): "**`link_registry` only records, never validates.** It *has* a `target_page_id` column + FK to `pages`, but `syncLinksToDB` never populates it. And `extract_and_sync_links` is wired into **no live workflow**, so `link_registry` is empty. It is not a usable substrate today."
- **what:** The internal-linking investigation initially proposed reusing the existing `link_registry` table/action as the base for a new phantom-link discovery check. Follow-up code reading found the table permanently empty in practice (the populating column is never written, and the syncing action isn't wired into any live workflow) — so the check that was actually built (`check_phantom_internal_links.go`) instead scans `rendered_html` directly via new shared helpers (`datahelpers/links.go`).
- **sources:** content_quality_and_internal_linking/FOCUS_internal_linking.md (archived); live FOCUS_internal_linking(1).md; running_notes_17(16) "Decisive findings"
- **relations:** hero/CTA link fabrication (above)
- **verify-later:** whether `ExtractAndSyncLinksAction`/`link_registry` were ever wired up subsequently.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Hub "Browse All X" link resolution — rejected design options
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** `PLAN_b4_b5_hubs_and_link_resolver(1).md`: "(a) Populate the `*_index_url` specs from real hubs. **Rejected** — per-site data to maintain; re-introduces the inconsistent-source brittleness. (b) `source: pages.<hub-name>` per component... bakes the `<area>-index` naming convention into each schema... (c) **Recommended.** A new `queryresolve` verb... `query.section_index_for:<type>`." Shipped per running_notes_17(16): "`section_index_for.go` — new `queryresolve` verb... B4/B5 — done."
- **what:** For the empty-href "Browse All Tools/Games/Guides" defect, two options were explicitly weighed and rejected in the design doc before settling on a new `queryresolve` verb: manually populating `*_index_url` site_specs (rejected as brittle, per-site maintenance) and a per-component `pages.<hub-name>` source (rejected as baking a naming convention into every schema). The chosen option — `query.section_index_for:<type>`, resolving the hub via shared `site_area_id`/URL-prefix — shipped and was confirmed working in deployed HTML.
- **sources:** content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md; running_notes_17(16)
- **relations:** hero/CTA link fabrication; internal-link-resolver agent
- **verify-later:** `queryresolve.go` `section_index_for` case in current code.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `internal-link-resolver` agent (Step 3) — dedicated intent-aware internal-link resolution
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** running_notes_17(16): "**Step 3 — completed (2026-06-11), all deliverables written**... Agent row (`internal_link_resolver_agent.sql`)... Writer wiring (`page_content_writer_link_resolver_wiring.sql`)... `unresolved_cta`: emitted in-Go... `status needs_human_review`." Confirmed live end-to-end: "Query D (corrected paths) on both completed rebuilds: `for_render=2`, `plan_count=2`, EQUAL ⇒ resolver augmented sections + writer loop consumed them."
- **what:** A new sub-agent of `page-content-writer` (modelled on `research-agent`, no persistence) that, at build time, resolves hero/CTA link destinations to intent-appropriate real pages (excluding the page's own hub, about/contact/legal) rather than a fixed contact page, validates every candidate against `datahelpers.PageURLSet`, and emits an `unresolved_cta` HITL signal when no destination can be found — the only place a "correctly dropped" (absent) button is detectable, since the deploy gate can't see an absence. Replaces the abandoned assumption that `component-template-fixer` already handled this.
- **sources:** content_quality_and_internal_linking/PLAN_b4_b5_hubs_and_link_resolver(1).md; running_notes_17(16) "Step 3" sections
- **relations:** component-template-fixer CTA-reuse assumption (superseded); identity-advisor/approval_mode (abandoned); hero/CTA fabrication fix
- **verify-later:** `internal_link_resolver_agent.sql`, `resolve_internal_links_action.go` current deployment/image-tag state.

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### link registry, cached navigation structures, and redirects (link-management foundation)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** Live Go code references `link_registry` and `navigation_structures` (e.g. platform/orchestration/actions/html_actions.go, site_db_actions.go, discovery_checks/check_phantom_internal_links.go, platform/orchestration/datahelpers/links.go), and a live doc `024_link_management_v2.md` exists — confirming this MVP schema's core concept shipped and was later versioned.
- **what:** The original link-management schema: a `link_registry` table indexing every link extracted from rendered components (source component/page/site, resolved target page/site, a `scope` of internal/page/site/network/external, a `link_type` of navigation/content/semantic/affiliate/reference, plus validation status for broken-link detection); `navigation_structures` as a **cached, versioned** JSONB nav tree per site+type (header/footer/mobile/sidebar), invalidated by a trigger on any `pages` INSERT/UPDATE/DELETE and rebuilt lazily via `get_current_navigation`/`build_navigation_for_site`; and a `redirects` table (301/302/307/410, hit_count, expiry). Deliberately reuses the existing generic `relationships` table for semantic content relationships (pillar/cluster, related-content, cross-site-reference) rather than inventing a parallel structure.
- **sources:** docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql
- **relations:** core client→network→site→page hierarchy (above, same migration file); link-management (024 anchor, 024_link_management_v2.md)
- **verify-later:** 024_link_management_v2.md — confirm what changed between this v1 schema and "v2"

<!-- SOURCE: U25_leopardess_social.md -->
### CTA-graph integrity (dead-end and circular primary actions)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** NOTES_provocations-index 2026-07-07: "Every primary action on the site dead-ends here" (pre-fix); 2026-07-09 "every primary CTA on the site resolves here"; CTA circularity "parked, Option B" pending a real arena page.
- **what:** The site's call-to-action graph as an auditable object: for two weeks every primary CTA (nav, hero, gauntlet-cta, lobby cards, provocation-card) pointed at an unbuilt page (404), invisible to any check; after the archive shipped, the graph is circular (hero → archive; archive → home; gauntlet-cta → archive) while the only real interactive surface (the Gauntlet tool) is reachable only via nav/footer. Decision Option B: leave until a real take-filing arena exists. Structural note: CTA URLs are baked into rendered sections, so a graph retarget is a section rebuild, not string surgery; brief-explanation's CTAs still carry '#' placeholders.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-index(4).md#2026-07-07; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#9.5; docs/social001_vonc_tiktok_social/tool_docs/NOTES_lobby-grid(6).md#2026-07-04
- **relations:** navigation; silent no-op success (the 404 destination was its product); link_registry
- **verify-later:** link_registry; CTA URLs in deployed vonc HTML
