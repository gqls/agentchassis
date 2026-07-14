# Cluster: content-quality-news-traffic
Categories included: content-governance, content-quality, news-feed-pipeline, traffic-analytics, new:topic-intelligence


<!-- SOURCE: U01_docs024_numbered_core.md -->
### Section editor and buildRenderContextFromDB
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 002(4) section; edit paths live in 013
- **what:** Granular edits without the full pipeline: content_edit (field_updates merge or full replacement) and component_swap; every edit updates content_data first then re-renders (edits survive re-renders). buildRenderContextFromDB reconstructs RenderContext purely from DB (site data, style collection, theme, nav, page meta, content_data) — no collected_data needed.
- **sources:** 002(4)#Section Editor; 003(8)#Source of truth principle
- **relations:** content_data source of truth; admin dashboard edit paths
- **verify-later:** section-editor definition; buildRenderContextFromDB

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Human direction channels and the pinned direction spec
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** 007 channels table; 004 v4 integration deployed for auditors; dashboard direction panel "pending" in 007 Phase 1 then ✅ in 013 Phase 4
- **what:** Three channels: work-item request (until completed), direction update (permanent, site_specs `direction` aspect, pinned — only humans change), reference suggestion (feeds next planning). Auditors must not flag must_have/should_have features for removal; strategist may add nice_to_have but can't remove must_have; direction change resets the audit pass counter.
- **sources:** 007#Human Direction; 004#Human Direction Integration; 013 Direction tab
- **relations:** audit pass reset; spec propagation
- **verify-later:** direction aspect reads in auditor prompts

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Page growth budget (three-tier weekly limits)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 006/013 ✅; the news-index-blocked-by-content-budget bug drove the third tier
- **what:** CheckPageGrowthBudget shared by apply_gap_plan and create_blog_posts: free under initial_target (12), then rolling 7-day caps per type — content 3/wk, blog 2/wk, structural (news-index, privacy…) 3/wk — absolute_max 60. Over-budget items become `blocked` (retryable), blog posts skip-and-continue. Config in site_specs aspect growth_config (pinnable).
- **sources:** 013#Page Growth Budget; 006 growth budget rows
- **relations:** content-gap-planner; blog planner
- **verify-later:** page_growth_budget.go tiers

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Three edit paths (direct edit / brief regenerate / direction propagate)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 012/013 shared table; all phases marked deployed
- **what:** Typos/fields: direct edit + auto-lock + rerender (seconds, no LLM). Section direction: edit content_brief → Regenerate → content_rewrite item, writer prompts from the brief (briefs control regeneration). Page: purpose edit + Regenerate Page (all unlocked sections). Site: Direction tab spec edit + explicit Propagate creating per-page items (skips fully-locked pages). Section suppression (page-scoped remove/restore) completes the set.
- **sources:** 013#Three Edit Paths, #Content Briefs & Regeneration
- **relations:** lock semantics; page_components.content_brief
- **verify-later:** content_brief column; regenerate endpoints

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Two sources of truth for site contact email
- **category:** content-governance
- **status-signal:** unknown
- **status-evidence:** "sites.email vs site_specs.identity.email can drift. loadSiteContactEmail uses COALESCE across both. Content writers may use either. Needs consolidation." (April 2026)
- **what:** Contact email lives in two places with no single owner; drift produces placeholder/incorrect contact details on pages (a recurring audit finding and false-positive source).
- **sources:** HANDOFF-pipeline-triage-april-2026.md#patterns-1
- **relations:** identity-advisor specialist; content quality catalogue (empty footer contact)
- **verify-later:** loadSiteContactEmail; identity aspect writers

<!-- SOURCE: U04_idea_uk.md -->
### Standing-ambition default in the mission aspect (aspiration has no generated home)
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** "Go action carrying a default in the mission aspect… owner to finalise the principle wording (draft in notes)" — still an open backlog item in every later list.
- **what:** Diagnosis: the classifier is a current-state tool; mission_brief/roadmap_brief are the designated aspirational slots but owner-supplied (nothing generates them), and strategy runs after the classifier — so with the slots empty, fresh builds describe what exists instead of leading the field. Fix (framework, not hand-seeding): domain-submitter always writes a mission_brief = a fixed platform standing-ambition principle (lead the vertical; most useful forward-looking content; build around the site's distinctive tools; surpass don't mirror), merged with any owner mission, in a Go action — no new aspect (aspect list checked: no vision/ambition; free-text), no reorder. Design deliberately excluded: ambition lifts content via the LLM readers, but site-design-planner is deterministic and doesn't read mission prose — design leadership is its own track.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Setting direction before the build); idea.uk/HANDOFF_chassis_site.md (the two decisions); idea.uk/running_notes(63).md (w/x checkpoints)
- **relations:** build-standard migration (the classifier-side sibling); mission-file submission; Phase 0 read.
- **verify-later:** domain-submitter def for any standing-ambition merge (expect absent).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Optimistic-lock co-management of shared rows across parallel chats
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "Optimistic-lock pattern for co-managed writes: WHERE updated_at = '<last-known>' — UPDATE 0 = stop, re-read, coordinate" (RUNBOOK(49) Constants); lock held across 3 idle days on the Step-7 write (NOTES §9bd).
- **what:** Multiple concurrent chats/agents co-manage the same shared components and sites, so every write to co-managed rows is guarded by a freshness check plus an optimistic-lock UPDATE on the last-known updated_at; UPDATE 0 means the row moved underneath — stop and coordinate, never blind-write. Includes proactive notification of the other chat and re-reading fleet/workflow state after parallel deploys (image bumps invalidate cached workflow knowledge).
- **sources:** RUNBOOK(49).md Constants; NOTES(43).md §3, §9ao, §9ap, §9bd; HANDOFF(7).md §Platform operating model; RUNBOOK_pre_cleanup_backup.md Step R7
- **relations:** snapshot-before-change; F8 remediation; locks category (031) more broadly.
- **verify-later:** whether any systematic lock/lease exists beyond the convention (relates to locks/031 concepts).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Snapshot-before-change backup conventions (snapshot_agent, manual version inserts, CTAS bak tables)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "Backups: snapshot_agent('<type>','<reason>') for agents; CTAS *_bak_* tables for data (two exist, see Cleanup)" (HANDOFF(7)); every mutating step in both threads shows a backup first.
- **what:** The layered backup discipline observed throughout: agents → snapshot_agent before config migrations (revert_agent to restore); shared component schema/template → manual component_versions INSERT mirroring the working insert paths; data rows → CREATE TABLE … AS SELECT `*_bak_*` tables (dropped only at closeout); templates → shell-redirected full-column dumps before anchored UPDATEs. Backups are named with dates and tracked as explicit cleanup debt.
- **sources:** HANDOFF(7).md §operating model; RUNBOOK(49).md Constants + Step 12(c); RUNBOOK_scheme_to_components(18).md W1/W2a backup steps
- **relations:** optimistic-lock co-management; component versioning; site-snapshots-and-revert (014) more broadly.
- **verify-later:** snapshot_agent/revert_agent SQL functions; outstanding bak tables.

<!-- SOURCE: U09_adoption.md -->
### Silent-fallback link family (phantom /contact.html, /services.html; hero CTA resolution)
- **category:** content-governance
- **status-signal:** unknown
- **status-evidence:** "Hero CTAs wrong site-wide (highest value — every page)… /services.html is a phantom page… Investigate hero-component CTA resolution" — still the named NEXT TASK as of HANDOFF_2026-06-09; no fix recorded in this corpus.
- **what:** Components emit links to pages that don't exist rather than resolving real targets or degrading gracefully: hero/bottom-CTA to /contact.html + /services.html on every page with text↔destination mismatch, header "Get Started", footer legal /terms.html, empty `href=""` per-card CTAs — several distinct mechanisms (schema fallback defaults like `cta_url use_fallback=/contact.html`, unresolved per-item fields, hardcoded legal links). Proposed direction: data-driven CTA resolution from realised pages plus a broad deployed-href vs realised-page-set audit distinguishing dropped source pages from component hallucination.
- **sources:** CATALOGUE(9)#family-b, HANDOFF_2026-06-09#next-task, HANDOFF_2026-05-25#parked-2
- **relations:** cta_url required/fallback fixes (the list-component subset, fixed); prepare_link_context/build_render_context; content quality polish batch (brand-suffix titles, empty footer contact/tagline, hero H1 reuse, empty meta descriptions)
- **verify-later:** hero content_components input_schema CTA fields; prepare_link_context available-pages resolution

<!-- SOURCE: U14_docs019_runbooks.md -->
### Content-regression guard on section save
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) worked example "NOTE the content-regression guard (L227) PROTECTS the content-rich page from being overwritten by an empty shell"; gamesdesign_index_rebuild Stage B "content regression blocked: new content has N chars vs M existing → … a loud, correct block, not a silent no-op. … Do not disable the guard to force it through."
- **what:** save_page_sections refuses to overwrite substantially richer live content with a much thinner regeneration (~3k vs ~13–15k chars observed) — so a stale page can be the guard PROTECTING content, presenting as success-no-change. Doctrine: a block means investigate why the writer produced thin content upstream, never disable the guard. Also the designed skip path: empty sections → skipped→complete via the workflow's check_skipped conditional.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#worked-example; docs019/RUNBOOK_gamesdesign_index_rebuild.md#5 (Stage B); docs019/RUNBOOK(31)_diagnosis_loop.md#update-5537ffdb
- **relations:** workflow result contract (the upstream thin-content cause); site quality LEG 4
- **verify-later:** save_page_sections_action.go regression check; check_skipped conditional

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Standards curation & governance — concern curators
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §2 "Ownership maps to the concern, not to a node in the agent tree, and the set of owners is flat (one per top-level concern, ~8–10)"
- **what:** One curator agent per top-level concern owns that concern's atoms — reusing the auditor pattern and doubling as the routing advisor. A curator does vigilance + drafting + mechanical health but holds no authority over a rule's *meaning*. Ownership is flat and horizontal, deliberately not tied to the volatile agent tree.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#2, ED/FOCUS_standards_curation_and_governance(1).md#3, ED/FOCUS_standards_curation_and_governance(1).md#6
- **relations:** atomic standard; confirm-not-initiate; coordinator role; user-representative advocate
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Coordinator role (arbitrates and frames)
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §7 "A thin coordinator layer above the curators … Resolved: the coordinator both arbitrates and frames"
- **what:** A thin layer above the peer curators owning what belongs to no single concern: the concern taxonomy, the `applies_to` vocabulary, cross-concern conflicts, and packaging cross-concern decisions for human confirmation. Both arbitrates and frames, checked by a user-aligned advocate inside the framing process.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#7, ED/FOCUS_standards_curation_and_governance(1).md#8
- **relations:** concern curators; user-representative advocate; decision authority
- **verify-later:** none

<!-- SOURCE: U17b_docs019_gofiles.md -->
### vonc.com mini-lobby content-edit re-render architecture
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "docs 003/002 say HTML patching was REJECTED as an edit mechanism ('content_data is always the source of truth … if we only patched rendered_html, the edit would be lost on the next re-render')" (bundle_minilobby_trim2.sh / bundle_minilobby_trim3.sh header)
- **what:** The established (not this-unit's-invention) rule that content edits must go through `content_data`, never by patching `rendered_html` directly, because the next re-render would discard a raw HTML patch. Two re-render paths exist: the full path (`needs_page` → page-content-writer, LLM) and the light path (`rerender_page_sections` behind a `page_rerender` item, no LLM, re-renders each section from stored `content_data` via `RenderComponentAction`) — and neither is `rerender_single_page`, which only assembles already-rendered components. `fix_component_template_action`'s `remove_element` fix_type explicitly does NOT touch `page_components.rendered_html` content, because "content changes go through the section-editor workflow" — leaving the correct edit path for a structural trim (like the provocation-card mini-lobby) genuinely unclear without a bundle to check.
- **sources:** contextkit/bundle_minilobby_trim2.sh#header, contextkit/bundle_minilobby_trim3.sh#header
- **relations:** vonc (site-case-studies), section-editor workflow, content-governance (013), styling-render-pipeline (036)
- **verify-later:** platform/orchestration/actions/fix_component_template_action.go and rerender_page_sections/rerender_single_page to confirm the scope boundaries described are still accurate; whether the mini-lobby trim task itself was ever completed

<!-- SOURCE: U18_sql_for_agents.md -->
### section-editor (granular edits that survive re-renders)
- **category:** content-governance
- **status-signal:** unknown
- **status-evidence:** Full definition with example trigger messages in 043; no later patches or timeout entries reference it.
- **what:** Edits a single page section without the full rebuild pipeline. Core invariant: always update content_data first, then re-render from template + DB context, so edits survive future re-renders (nav updates, theme changes). Supports content_edit (field merge or full replace) and component_swap (new template, same content_data). Target addressed by page_component_id or (page_name + slot_name).
- **sources:** 043_section_editor.sql
- **relations:** inline editing / content governance concepts; page_components
- **verify-later:** whether section-editor exists live and is used by admin UI

<!-- SOURCE: U18_sql_for_agents.md -->
### Blog/content planning agents (blog-content-planner, content-gap-planner, internal-linker)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 069 documents the full empty_blog loop; 070/071/101 definitions; content-gap-planner in 075's timeout list.
- **what:** LLM planners that turn detected content gaps into concrete work: blog-content-planner (needs_blog_posts) plans 3–4 posts from specs and reuses write_build_items to create pages + needs_content_page items + blog-index rerender; content-gap-planner (needs_content_planning) decides per gap between add-section (content_rewrite), new page, spec update (needs_spec_update), or wont_fix — "The LLM here is the PLANNER, not the auditor"; internal-linker finds pages that should contextually link to an orphaned sub-page and creates content_rewrite items for natural placements. 070 records the reuse-over-new-Go deliberation verbatim.
- **sources:** 069_blog_posts.sql; 070_blog_content_planner.sql; 071_content_gap_planner.sql; 101_internal_linker.sql
- **relations:** empty_blog / orphan page checks; page-build-handler executes their items; spec-updater
- **verify-later:** create_blog_posts action; empty_blog check

<!-- SOURCE: U19_sql_tables_components.md -->
### page_components.build_status CHECK constraint
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "APPLIED 2026-07-11" header; pre-flight survey documented (deployed 597, pending 20; writers enumerated); constraint proved via pg_get_constraintdef.
- **what:** build_status was free text, which let apply_section_edit invent 'approved' and silently remove a live section from every discovery check's audit surface (all filter build_status='deployed'). CHECK now restricts to deployed/pending/approved/removed/needs_rebuild — turning invented values into loud write failures. 'removed' and 'needs_rebuild' retained without writers so future writers need no migration; residual legitimate-'approved'-stuck case covered by the page_component_status_drift check.
- **sources:** docs/agent_docs/sql_for_tables/049_page_components_build_status_check.sql
- **relations:** improvement-loop discovery checks; PLAN_generalise_fixes_to_fleet §4; evidence-based claimed-item timeout (deployed-flag trust).
- **verify-later:** page_components_build_status_check constraint; check_page_component_status_drift.

<!-- SOURCE: U19_sql_tables_components.md -->
### Schema-mode strict/flexible subsystem (abandoned)
- **category:** content-governance
- **status-signal:** abandoned
- **status-evidence:** 009 drop migration (2026-07-09): "only partially applied and then abandoned... snapshot columns were never created in production... no Go code reads schema_mode... auto_lock_on_deploy fired exactly once in the system's history"; trigger and function dropped, single strict row normalised.
- **what:** A designed governance regime where approved sections lock to 'strict' schema mode (schema_snapshot + content_snapshot captured; lock_section_to_strict / unlock_section_for_redesign functions; auto-lock on first deploy per sites.strict_mode_trigger). It became an active liability when the apply_section_edit build_status fix would have made every edited section the only locked row on a site for a feature nothing consumed. Orphan functions and columns deliberately left in place; function body preserved as backup.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql; docs/agent_docs/sql_for_tables/009_drop_auto_lock_on_deploy.sql; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#item6
- **relations:** superseded in spirit by Pattern A locks + page_component_history; component_versions.
- **verify-later:** absence of trigger; orphan columns schema_mode/strict_mode_trigger still present.

<!-- SOURCE: U19_sql_tables_components.md -->
### Growth budget spec (growth_config)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** Seed migration inserts default growth_config for existing active sites: '{"initial_target": 12, "weekly_content_pages_max": 3, "weekly_blog_posts_max": 2, "absolute_max": 60}'.
- **what:** Per-site growth limits stored as a site_specs aspect: initial page target, weekly content/blog caps, absolute page maximum. Admin-overridable via the dashboard Direction tab; consumed by growth/budget calculations in planning.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#growth_config
- **relations:** site_specs store; page_growth_budget.go (referenced from 042 thunder step-zero).
- **verify-later:** budget enforcement in planner/discovery.

<!-- SOURCE: U19_sql_tables_components.md -->
### content_items reusable content layer
- **category:** content-governance
- **status-signal:** unknown
- **status-evidence:** Full DDL + helper get_component_content + v_content_usage view exist and page_components.content_item_id survives into the live dump, but no later file shows content_items being written.
- **what:** Separates "what to say" from "how to show it": typed reusable content rows (headline, tagline, service_description, testimonial, bio, cta, faq...) with semantic content_key, plain_text search, library sharing (site_id NULL + is_library + industry_vertical + library_tags), assets-style origin tracking, and status workflow. page_components reference a content_item with content_data acting as shallow-merge override (get_component_content). Would let one tagline appear in hero, footer and meta without duplication and let library content seed new sites.
- **sources:** docs/agent_docs/sql_for_tables/004_content_items.sql; docs/agent_docs/sql_for_tables/004b_content_items.md; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql#content_item_id
- **relations:** pages/page_components split; assets origin pattern.
- **verify-later:** content_items row count; any writer using content_item_id.

<!-- SOURCE: U19_sql_tables_components.md -->
### page_component_history full-snapshot content history
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** Created in Phase-0 Block A (019) explicitly as the replacement for the dropped content_snapshot/schema_snapshot columns.
- **what:** Before any content_data write to a page_component, the current value is copied here as a complete snapshot (not a diff) with source ('content-writer', 'section-editor', 'rollback'...) and triggering work item id — the rollback/audit substrate for section edits.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#3
- **relations:** replaces schema-mode snapshots; section-editor; site snapshots (page-level vs site-level revert).
- **verify-later:** writers copying into history before UPDATE.

<!-- SOURCE: U19_sql_tables_components.md -->
### Section governance columns: content_brief and suppressed_sections
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "content_brief: records the instructions that generated each component's content. Enables admins to see, edit, and regenerate" (008 tail); "suppressed_sections... prevents discovery checks from recreating sections that were intentionally removed. The DELETE component endpoint writes to this column" (003).
- **what:** Two small governance mechanisms: page_components.content_brief JSONB preserves the generation instructions per section for admin-editable regeneration; pages.suppressed_sections lists intentionally removed section functions so discovery does not resurrect them (component-removal flow Phases 2/5).
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#content-brief; docs/agent_docs/sql_for_tables/003_pages.sql#suppressed_sections
- **relations:** improvement-loop discovery checks; inline editing / regeneration.
- **verify-later:** DELETE component endpoint; discovery filters on suppressed_sections.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Page-content-writer + admin content brief regeneration flow
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** 100_content_page_build_handler_flow.md documents the live flow (page-build-handler → page-content-writer → load_site_specs → prepare_link_context → load_page_components → process_sections_loop) then states "What's missing: The prompt has no awareness of page_components.content_brief" and specifies the new content_rewrite flow.
- **what:** The bridge document into the modern era: the current work-item pipeline's content generation path, and the gap it fixes — admin edits a brief in the dashboard (page_components.content_brief), clicks Regenerate creating a content_rewrite work item, and the writer's generate_content prompt gains an "## Admin Content Brief" block for briefed sections while unbriefed sections behave as before.
- **sources:** docs001_flow_general/100_content_page_build_handler_flow.md
- **relations:** content-governance (briefs, regeneration — anchor doc 013); development-guide work-item lifecycle; content_components definitions.
- **verify-later:** page_components.content_brief column; content_rewrite work item type; load_page_section_components step.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Spatial addressing for natural-language editing
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** Vision in docs009/001 ("edit the paragraph on the left of the blue call to action... data-uuid and data-path attributes"); partially realized per docs018/008: "Component labeling (new) — injects data-pc-id, data-slot, data-position into each <section>".
- **what:** Every visible element carries a unique ID and genealogy path so an editing agent can resolve fuzzy human instructions spatially ("3rd paragraph on the left", "the one under the yellow button") by highlighting candidates iteratively. Shipped at section granularity (data-pc-id/data-slot/data-position mapping sections to page_components rows); element-level addressing and the conversational disambiguation loop remain unfulfilled.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-The-Spatial-Address-System; docs018_rerendering/008_granular_editing.md#What-Exists-Today; docs012_site_maps_and_components/006_start_concluding_links.md#2.4
- **relations:** section-editor agent; page_components.data_uuid/data_path columns; granular editing spectrum.
- **verify-later:** data_uuid/data_path columns on page_components; labeling injection code.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Flexible vs strict schema mode (approval snapshot)
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** docs017/001_flexible proposes sites.schema_mode + page_components.schema_snapshot/content_snapshot/component_version_id; docs018/008 lists "page_components.content_snapshot stores approved values... schema_mode can be strict or flexible per section" under "What Exists Today".
- **what:** Two-phase content lifecycle: initial builds run flexible (best-effort substitution, warnings on missing fields, creative freedom); at approval the section's input_schema and content values are snapshotted and the section moves to strict mode (edits validated against locked schema, unsubstituted placeholders fail, template upgrades can't break approved pages, rollback via content_snapshot). Open questions recorded: granularity, versioning vs snapshot, transition trigger, unlock capability.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/001_flexible_schema_enforcement.md; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-5; docs018_rerendering/008_granular_editing.md
- **relations:** locks (current descendant of the freeze idea); content-reviewer approval; section editor.
- **verify-later:** schema_mode/schema_snapshot/content_snapshot columns and whether any transition code exists.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Legal content agent + legal constraint rules
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** docs017/019b "legal-content-agent | Template + minimal LLM | New" with legal_rules JSON (required_disclaimers by trigger, forbidden_phrases, required_pages by jurisdiction template); compliance-discovery-agent phased for maintenance.
- **what:** Jurisdiction-aware legal pages from vetted templates (privacy-gdpr-uk etc.), plus machine-readable constraints exported to the content writer: disclaimers triggered by content conditions with placement rules, forbidden phrases per industry ("guaranteed returns"), required pages routed to the legal nav group. Compliance discovery monitors regulatory changes in maintenance.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Legal-Content-Agent; docs017_legacy_agent_rules_images_design_keydocs/022_maintenance_architecture_plan_v2.md#Discovery-Agents
- **relations:** policy-as-filters (org framework ancestor); finance site stress test; brand DNA forbidden phrases.
- **verify-later:** legal-content-agent definition; legal_rules key in sites.content_data.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Content review flow with rejection → needs_attention
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** docs011/003 flow diagram (validate → auto-eval vs HITL → approve/reject → finalize_hitl or mark_page needs_attention); docs018/003: "Rejected pages are picked up by maintenance workflow."
- **what:** The content-reviewer agent's dual-mode gate: algorithmic validation feeds either auto-evaluation (errors escalate) or HITL review (human sees issues, can edit HTML inline); outcomes are approve (deploy), approve-with-edits (use edited HTML), or reject (page marked needs_attention, skipped by the build loop, queued for maintenance). Established review as a first-class pipeline stage rather than an afterthought.
- **sources:** docs011_api_hitl/003_hitl_new_plan.md; docs018_rerendering/003_website_builder_architecture_status_report.md#4; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Active-Agents
- **relations:** HITL protocol; flexible/strict approval snapshot; maintenance pickup of needs_attention.
- **verify-later:** content-reviewer workflow JSON; needs_attention status handling in current loop.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Section editor agent (content_data is the source of truth)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** docs018/010 records decisions as implemented ("Decision: content_data is always the source of truth... We considered a lightweight html_patch edit type... decided against"; reused-code inventory; ActionInputSpec pattern used).
- **what:** Granular post-deploy editing without the full pipeline: two edit types — content_edit (field_updates merge or full content_data replace) and component_swap (new template, same content) — both updating page_components.content_data first, then re-rendering via buildRenderContextFromDB (reconstructs the full RenderContext purely from DB: site data, style collection, theme, nav, page meta, section content), reassembling the page, committing, and deploying. HTML patching was explicitly rejected because edits would vanish on the next rerender. Targets identified by page_component_id or page_name+slot_name (normalized). Future: LLM section rewrite via content_direction; bulk edits.
- **sources:** docs018_rerendering/010_section_editor_architecture.md; docs018_rerendering/008_granular_editing.md
- **relations:** granular editing spectrum; spatial addressing labels; rerender assemblePage; locks/inline editing (current descendants).
- **verify-later:** load_edit_context/apply_section_edit actions; section-editor agent definition.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Granular editing spectrum (word → multi-page)
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** docs018/008 table mapping six edit scopes to mechanisms and LLM needs; "What Exists Today" vs "What's Needed: One New Agent, Two New Actions"; content_direction JSONB "(new)".
- **what:** A routing model for edit requests by scope: word/phrase (direct patch — later rejected in favour of content_data edits), field value (template re-render), section rewrite (content-writer on one section), component swap, page rewrite (page-rebuild with content_direction instructions flowing into prompts), multi-page (maintenance-triage → page-rebuild). All routed through the same maintenance infrastructure.
- **sources:** docs018_rerendering/008_granular_editing.md; docs018_rerendering/010_section_editor_architecture.md#Future-Extensions
- **relations:** section editor; work items; content_direction column on pages.
- **verify-later:** content_direction column and its prompt integration.

<!-- SOURCE: U22_recent_small_docs.md -->
### maintenance_queue + claim/complete/fail functions
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** "MAINTENANCE QUEUE TABLE (for future use) ... For now, pages are flagged manually and the agent is triggered via generic-agent." Table + PL/pgSQL claim/complete/fail functions defined.
- **what:** A generic site-maintenance work queue (`maintenance_queue`) keyed by site_id with task_type (page_rebuild/css_update/nav_fix/link_repair/content_refresh), priority, JSONB payload, retry logic, and atomic `claim_maintenance_task`/`complete_maintenance_task`/`fail_maintenance_task` SQL functions using SKIP LOCKED. Later reused as the trigger surface for the chatbot install (`install_chat` task).
- **sources:** docs019_business/016_maintenance_queue_table.sql, docs019_business/017_maintenance_triage_agent.sql, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path
- **relations:** maintenance-triage agent, site-chat-installer, improvement-loop
- **verify-later:** maintenance_queue table + claim/complete/fail functions

<!-- SOURCE: U22_recent_small_docs.md -->
### maintenance_queue as generic install/uninstall trigger surface
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** Chatbot design reuses it: "Installation is requested by enqueuing a maintenance task — task_type='install_chat' on the existing maintenance_queue (which already has site_id, payload, status, retries)."
- **what:** The recognition that the existing `maintenance_queue` (built for page rebuilds) is a reusable, generic trigger surface for opt-in site add-ons — chat install/uninstall being the first — without touching the build pipeline. Establishes a pattern: new per-site capabilities enqueue a maintenance task rather than becoming a build-pipeline stage. (Cross-cut between docs019 infra and docs025 chatbot.)
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path, docs019_business/016_maintenance_queue_table.sql
- **relations:** maintenance_queue, site-chat-installer, maintenance-triage
- **verify-later:** maintenance_queue task_type values in use

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Recommendation-specialist architecture (bug vs recommendation vs gap)
- **category:** content-governance
- **status-signal:** abandoned
- **status-evidence:** PLAN_design-note-recommendation-specialists (April 2026): "Status: Proposed — not yet implemented"; "Deferred Until HITL queue becomes a bottleneck"; interim workaround is a manual `UPDATE site_work_items SET status='wont_fix'`.
- **what:** Proposal to make auditors emit an explicit `finding_type` (bug | recommendation | gap) so the pipeline stops auto-rewriting subjective opinions as if they were bugs (which caused false-positive rewrites like inventing a fake email). Bugs → `content_rewrite`; gaps → `needs_content_page` (P9, already deployed); recommendations → new specialist agents (identity-advisor, tone-shift-agent, content-strategist) returning apply/dismiss/escalate, gated by a per-site `sites.approval_mode` (auto | review). Never implemented.
- **sources:** old_design_and_styling/PLAN_design-note-recommendation-specialists.md#proposed-architecture, #per-site-approval-mode
- **relations:** HITL review flow; approval_mode; write_audit_findings_action.go; content-quality/visual-design auditors
- **verify-later:** write_audit_findings_action.go auditFinding struct; sites.approval_mode column existence

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Privacy posture (no cookies/JS/IP; UK GDPR/PECR)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** running_notes standing observations "Privacy posture (UK GDPR/PECR, low risk appetite): no cookies, no JS, no IP stored, referer reduced to host, country only from a coarse CDN header".
- **what:** A deliberate low-risk privacy stance baked into the engine and page: no cookies, no JavaScript, no IP stored, referer reduced to host only, country only from a coarse CDN header (CF-IPCountry). Open ingest choice: redact email/phone patterns at ingest vs rely on the 90-day prune.
- **sources:** traffic_probe_running_notes(27).md#standing-observations, traffic_probe_plan(11).md#p4, traffic_probe_runbook(12).md#6
- **relations:** shapes intent-probe (no-JS form), sanitisation, retention timer
- **verify-later:** engine handlers (no cookie/IP); site-engine-prune.timer RETENTION_DAYS=90

<!-- SOURCE: U25_leopardess_social.md -->
### site_specs `pinned` flag not honoured by the write path
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** REPLICATION §2: "WriteSiteSpec does not honour pinned … today the only thing protecting these specs is that improvement-sweep is disabled" (2026-07-10).
- **what:** Operator-corrected specs are marked pinned=true, but WriteSiteSpec supersedes unconditionally — pinned is an admin-display flag only. A pinned check in the spec write path is a named platform gap; the current protection is incidental (the improvement-sweep scheduler that fires content-gap-planner has been disabled since 2026-05-02).
- **sources:** docs/leopardessconsulting/REPLICATION_in_chassis.md#2; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-5
- **relations:** LLM fabrication classes; discovery/improvement-sweep disabled state
- **verify-later:** platform/orchestration/actions/site_spec_actions.go; scheduled_tasks row for improvement-sweep

<!-- SOURCE: U25_leopardess_social.md -->
### Section-editor: single-section edit propagation path
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** VERDICT §7: "the section-editor's first ever production run (3 orchestrations, all COMPLETED)" 2026-07-09; fix verified end-to-end on v1.0.1102 2026-07-10.
- **what:** The sanctioned route for a template change to reach one live section: apply_section_edit (edit_type content_edit) re-renders exactly one page_component from its (changed) template, rewrites rendered_html + content_data, reassembles the page from the other stored sections and commits — blast radius one section, unlike rerender_page_sections (all sections, gated on reason=image_landed/section_data_resolved) or assemble-only rerender_single_page. Doc 003's "HTML patching was rejected as an edit mechanism" is honoured by editing the source template (with a hand-written component_versions snapshot, since direct SQL is the only writer for arbitrary template edits) and letting the action re-render. Quirk: field_updates must be non-empty (a no-op merge re-supplying a current value works).
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#0, #1, #5; docs/social001_vonc_tiktok_social/minilobby_task/086_section_edit_provocation-card_vonc.sh (header)
- **relations:** page-rerender mode contract; build_status approved defect; mini-lobby trim
- **verify-later:** section_editor_actions.go apply_section_edit; orchestration_states for section-editor runs

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Section editor and buildRenderContextFromDB
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 002(4) section; edit paths live in 013
- **what:** Granular edits without the full pipeline: content_edit (field_updates merge or full replacement) and component_swap; every edit updates content_data first then re-renders (edits survive re-renders). buildRenderContextFromDB reconstructs RenderContext purely from DB (site data, style collection, theme, nav, page meta, content_data) — no collected_data needed.
- **sources:** 002(4)#Section Editor; 003(8)#Source of truth principle
- **relations:** content_data source of truth; admin dashboard edit paths
- **verify-later:** section-editor definition; buildRenderContextFromDB

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Human direction channels and the pinned direction spec
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** 007 channels table; 004 v4 integration deployed for auditors; dashboard direction panel "pending" in 007 Phase 1 then ✅ in 013 Phase 4
- **what:** Three channels: work-item request (until completed), direction update (permanent, site_specs `direction` aspect, pinned — only humans change), reference suggestion (feeds next planning). Auditors must not flag must_have/should_have features for removal; strategist may add nice_to_have but can't remove must_have; direction change resets the audit pass counter.
- **sources:** 007#Human Direction; 004#Human Direction Integration; 013 Direction tab
- **relations:** audit pass reset; spec propagation
- **verify-later:** direction aspect reads in auditor prompts

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Page growth budget (three-tier weekly limits)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 006/013 ✅; the news-index-blocked-by-content-budget bug drove the third tier
- **what:** CheckPageGrowthBudget shared by apply_gap_plan and create_blog_posts: free under initial_target (12), then rolling 7-day caps per type — content 3/wk, blog 2/wk, structural (news-index, privacy…) 3/wk — absolute_max 60. Over-budget items become `blocked` (retryable), blog posts skip-and-continue. Config in site_specs aspect growth_config (pinnable).
- **sources:** 013#Page Growth Budget; 006 growth budget rows
- **relations:** content-gap-planner; blog planner
- **verify-later:** page_growth_budget.go tiers

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Three edit paths (direct edit / brief regenerate / direction propagate)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 012/013 shared table; all phases marked deployed
- **what:** Typos/fields: direct edit + auto-lock + rerender (seconds, no LLM). Section direction: edit content_brief → Regenerate → content_rewrite item, writer prompts from the brief (briefs control regeneration). Page: purpose edit + Regenerate Page (all unlocked sections). Site: Direction tab spec edit + explicit Propagate creating per-page items (skips fully-locked pages). Section suppression (page-scoped remove/restore) completes the set.
- **sources:** 013#Three Edit Paths, #Content Briefs & Regeneration
- **relations:** lock semantics; page_components.content_brief
- **verify-later:** content_brief column; regenerate endpoints

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Two sources of truth for site contact email
- **category:** content-governance
- **status-signal:** unknown
- **status-evidence:** "sites.email vs site_specs.identity.email can drift. loadSiteContactEmail uses COALESCE across both. Content writers may use either. Needs consolidation." (April 2026)
- **what:** Contact email lives in two places with no single owner; drift produces placeholder/incorrect contact details on pages (a recurring audit finding and false-positive source).
- **sources:** HANDOFF-pipeline-triage-april-2026.md#patterns-1
- **relations:** identity-advisor specialist; content quality catalogue (empty footer contact)
- **verify-later:** loadSiteContactEmail; identity aspect writers

<!-- SOURCE: U04_idea_uk.md -->
### Standing-ambition default in the mission aspect (aspiration has no generated home)
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** "Go action carrying a default in the mission aspect… owner to finalise the principle wording (draft in notes)" — still an open backlog item in every later list.
- **what:** Diagnosis: the classifier is a current-state tool; mission_brief/roadmap_brief are the designated aspirational slots but owner-supplied (nothing generates them), and strategy runs after the classifier — so with the slots empty, fresh builds describe what exists instead of leading the field. Fix (framework, not hand-seeding): domain-submitter always writes a mission_brief = a fixed platform standing-ambition principle (lead the vertical; most useful forward-looking content; build around the site's distinctive tools; surpass don't mirror), merged with any owner mission, in a Go action — no new aspect (aspect list checked: no vision/ambition; free-text), no reorder. Design deliberately excluded: ambition lifts content via the LLM readers, but site-design-planner is deterministic and doesn't read mission prose — design leadership is its own track.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Setting direction before the build); idea.uk/HANDOFF_chassis_site.md (the two decisions); idea.uk/running_notes(63).md (w/x checkpoints)
- **relations:** build-standard migration (the classifier-side sibling); mission-file submission; Phase 0 read.
- **verify-later:** domain-submitter def for any standing-ambition merge (expect absent).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Optimistic-lock co-management of shared rows across parallel chats
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "Optimistic-lock pattern for co-managed writes: WHERE updated_at = '<last-known>' — UPDATE 0 = stop, re-read, coordinate" (RUNBOOK(49) Constants); lock held across 3 idle days on the Step-7 write (NOTES §9bd).
- **what:** Multiple concurrent chats/agents co-manage the same shared components and sites, so every write to co-managed rows is guarded by a freshness check plus an optimistic-lock UPDATE on the last-known updated_at; UPDATE 0 means the row moved underneath — stop and coordinate, never blind-write. Includes proactive notification of the other chat and re-reading fleet/workflow state after parallel deploys (image bumps invalidate cached workflow knowledge).
- **sources:** RUNBOOK(49).md Constants; NOTES(43).md §3, §9ao, §9ap, §9bd; HANDOFF(7).md §Platform operating model; RUNBOOK_pre_cleanup_backup.md Step R7
- **relations:** snapshot-before-change; F8 remediation; locks category (031) more broadly.
- **verify-later:** whether any systematic lock/lease exists beyond the convention (relates to locks/031 concepts).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Snapshot-before-change backup conventions (snapshot_agent, manual version inserts, CTAS bak tables)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "Backups: snapshot_agent('<type>','<reason>') for agents; CTAS *_bak_* tables for data (two exist, see Cleanup)" (HANDOFF(7)); every mutating step in both threads shows a backup first.
- **what:** The layered backup discipline observed throughout: agents → snapshot_agent before config migrations (revert_agent to restore); shared component schema/template → manual component_versions INSERT mirroring the working insert paths; data rows → CREATE TABLE … AS SELECT `*_bak_*` tables (dropped only at closeout); templates → shell-redirected full-column dumps before anchored UPDATEs. Backups are named with dates and tracked as explicit cleanup debt.
- **sources:** HANDOFF(7).md §operating model; RUNBOOK(49).md Constants + Step 12(c); RUNBOOK_scheme_to_components(18).md W1/W2a backup steps
- **relations:** optimistic-lock co-management; component versioning; site-snapshots-and-revert (014) more broadly.
- **verify-later:** snapshot_agent/revert_agent SQL functions; outstanding bak tables.

<!-- SOURCE: U09_adoption.md -->
### Silent-fallback link family (phantom /contact.html, /services.html; hero CTA resolution)
- **category:** content-governance
- **status-signal:** unknown
- **status-evidence:** "Hero CTAs wrong site-wide (highest value — every page)… /services.html is a phantom page… Investigate hero-component CTA resolution" — still the named NEXT TASK as of HANDOFF_2026-06-09; no fix recorded in this corpus.
- **what:** Components emit links to pages that don't exist rather than resolving real targets or degrading gracefully: hero/bottom-CTA to /contact.html + /services.html on every page with text↔destination mismatch, header "Get Started", footer legal /terms.html, empty `href=""` per-card CTAs — several distinct mechanisms (schema fallback defaults like `cta_url use_fallback=/contact.html`, unresolved per-item fields, hardcoded legal links). Proposed direction: data-driven CTA resolution from realised pages plus a broad deployed-href vs realised-page-set audit distinguishing dropped source pages from component hallucination.
- **sources:** CATALOGUE(9)#family-b, HANDOFF_2026-06-09#next-task, HANDOFF_2026-05-25#parked-2
- **relations:** cta_url required/fallback fixes (the list-component subset, fixed); prepare_link_context/build_render_context; content quality polish batch (brand-suffix titles, empty footer contact/tagline, hero H1 reuse, empty meta descriptions)
- **verify-later:** hero content_components input_schema CTA fields; prepare_link_context available-pages resolution

<!-- SOURCE: U14_docs019_runbooks.md -->
### Content-regression guard on section save
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** thin_slice(27) worked example "NOTE the content-regression guard (L227) PROTECTS the content-rich page from being overwritten by an empty shell"; gamesdesign_index_rebuild Stage B "content regression blocked: new content has N chars vs M existing → … a loud, correct block, not a silent no-op. … Do not disable the guard to force it through."
- **what:** save_page_sections refuses to overwrite substantially richer live content with a much thinner regeneration (~3k vs ~13–15k chars observed) — so a stale page can be the guard PROTECTING content, presenting as success-no-change. Doctrine: a block means investigate why the writer produced thin content upstream, never disable the guard. Also the designed skip path: empty sections → skipped→complete via the workflow's check_skipped conditional.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#worked-example; docs019/RUNBOOK_gamesdesign_index_rebuild.md#5 (Stage B); docs019/RUNBOOK(31)_diagnosis_loop.md#update-5537ffdb
- **relations:** workflow result contract (the upstream thin-content cause); site quality LEG 4
- **verify-later:** save_page_sections_action.go regression check; check_skipped conditional

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Standards curation & governance — concern curators
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §2 "Ownership maps to the concern, not to a node in the agent tree, and the set of owners is flat (one per top-level concern, ~8–10)"
- **what:** One curator agent per top-level concern owns that concern's atoms — reusing the auditor pattern and doubling as the routing advisor. A curator does vigilance + drafting + mechanical health but holds no authority over a rule's *meaning*. Ownership is flat and horizontal, deliberately not tied to the volatile agent tree.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#2, ED/FOCUS_standards_curation_and_governance(1).md#3, ED/FOCUS_standards_curation_and_governance(1).md#6
- **relations:** atomic standard; confirm-not-initiate; coordinator role; user-representative advocate
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Coordinator role (arbitrates and frames)
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §7 "A thin coordinator layer above the curators … Resolved: the coordinator both arbitrates and frames"
- **what:** A thin layer above the peer curators owning what belongs to no single concern: the concern taxonomy, the `applies_to` vocabulary, cross-concern conflicts, and packaging cross-concern decisions for human confirmation. Both arbitrates and frames, checked by a user-aligned advocate inside the framing process.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#7, ED/FOCUS_standards_curation_and_governance(1).md#8
- **relations:** concern curators; user-representative advocate; decision authority
- **verify-later:** none

<!-- SOURCE: U17b_docs019_gofiles.md -->
### vonc.com mini-lobby content-edit re-render architecture
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "docs 003/002 say HTML patching was REJECTED as an edit mechanism ('content_data is always the source of truth … if we only patched rendered_html, the edit would be lost on the next re-render')" (bundle_minilobby_trim2.sh / bundle_minilobby_trim3.sh header)
- **what:** The established (not this-unit's-invention) rule that content edits must go through `content_data`, never by patching `rendered_html` directly, because the next re-render would discard a raw HTML patch. Two re-render paths exist: the full path (`needs_page` → page-content-writer, LLM) and the light path (`rerender_page_sections` behind a `page_rerender` item, no LLM, re-renders each section from stored `content_data` via `RenderComponentAction`) — and neither is `rerender_single_page`, which only assembles already-rendered components. `fix_component_template_action`'s `remove_element` fix_type explicitly does NOT touch `page_components.rendered_html` content, because "content changes go through the section-editor workflow" — leaving the correct edit path for a structural trim (like the provocation-card mini-lobby) genuinely unclear without a bundle to check.
- **sources:** contextkit/bundle_minilobby_trim2.sh#header, contextkit/bundle_minilobby_trim3.sh#header
- **relations:** vonc (site-case-studies), section-editor workflow, content-governance (013), styling-render-pipeline (036)
- **verify-later:** platform/orchestration/actions/fix_component_template_action.go and rerender_page_sections/rerender_single_page to confirm the scope boundaries described are still accurate; whether the mini-lobby trim task itself was ever completed

<!-- SOURCE: U18_sql_for_agents.md -->
### section-editor (granular edits that survive re-renders)
- **category:** content-governance
- **status-signal:** unknown
- **status-evidence:** Full definition with example trigger messages in 043; no later patches or timeout entries reference it.
- **what:** Edits a single page section without the full rebuild pipeline. Core invariant: always update content_data first, then re-render from template + DB context, so edits survive future re-renders (nav updates, theme changes). Supports content_edit (field merge or full replace) and component_swap (new template, same content_data). Target addressed by page_component_id or (page_name + slot_name).
- **sources:** 043_section_editor.sql
- **relations:** inline editing / content governance concepts; page_components
- **verify-later:** whether section-editor exists live and is used by admin UI

<!-- SOURCE: U18_sql_for_agents.md -->
### Blog/content planning agents (blog-content-planner, content-gap-planner, internal-linker)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 069 documents the full empty_blog loop; 070/071/101 definitions; content-gap-planner in 075's timeout list.
- **what:** LLM planners that turn detected content gaps into concrete work: blog-content-planner (needs_blog_posts) plans 3–4 posts from specs and reuses write_build_items to create pages + needs_content_page items + blog-index rerender; content-gap-planner (needs_content_planning) decides per gap between add-section (content_rewrite), new page, spec update (needs_spec_update), or wont_fix — "The LLM here is the PLANNER, not the auditor"; internal-linker finds pages that should contextually link to an orphaned sub-page and creates content_rewrite items for natural placements. 070 records the reuse-over-new-Go deliberation verbatim.
- **sources:** 069_blog_posts.sql; 070_blog_content_planner.sql; 071_content_gap_planner.sql; 101_internal_linker.sql
- **relations:** empty_blog / orphan page checks; page-build-handler executes their items; spec-updater
- **verify-later:** create_blog_posts action; empty_blog check

<!-- SOURCE: U19_sql_tables_components.md -->
### page_components.build_status CHECK constraint
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "APPLIED 2026-07-11" header; pre-flight survey documented (deployed 597, pending 20; writers enumerated); constraint proved via pg_get_constraintdef.
- **what:** build_status was free text, which let apply_section_edit invent 'approved' and silently remove a live section from every discovery check's audit surface (all filter build_status='deployed'). CHECK now restricts to deployed/pending/approved/removed/needs_rebuild — turning invented values into loud write failures. 'removed' and 'needs_rebuild' retained without writers so future writers need no migration; residual legitimate-'approved'-stuck case covered by the page_component_status_drift check.
- **sources:** docs/agent_docs/sql_for_tables/049_page_components_build_status_check.sql
- **relations:** improvement-loop discovery checks; PLAN_generalise_fixes_to_fleet §4; evidence-based claimed-item timeout (deployed-flag trust).
- **verify-later:** page_components_build_status_check constraint; check_page_component_status_drift.

<!-- SOURCE: U19_sql_tables_components.md -->
### Schema-mode strict/flexible subsystem (abandoned)
- **category:** content-governance
- **status-signal:** abandoned
- **status-evidence:** 009 drop migration (2026-07-09): "only partially applied and then abandoned... snapshot columns were never created in production... no Go code reads schema_mode... auto_lock_on_deploy fired exactly once in the system's history"; trigger and function dropped, single strict row normalised.
- **what:** A designed governance regime where approved sections lock to 'strict' schema mode (schema_snapshot + content_snapshot captured; lock_section_to_strict / unlock_section_for_redesign functions; auto-lock on first deploy per sites.strict_mode_trigger). It became an active liability when the apply_section_edit build_status fix would have made every edited section the only locked row on a site for a feature nothing consumed. Orphan functions and columns deliberately left in place; function body preserved as backup.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql; docs/agent_docs/sql_for_tables/009_drop_auto_lock_on_deploy.sql; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#item6
- **relations:** superseded in spirit by Pattern A locks + page_component_history; component_versions.
- **verify-later:** absence of trigger; orphan columns schema_mode/strict_mode_trigger still present.

<!-- SOURCE: U19_sql_tables_components.md -->
### Growth budget spec (growth_config)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** Seed migration inserts default growth_config for existing active sites: '{"initial_target": 12, "weekly_content_pages_max": 3, "weekly_blog_posts_max": 2, "absolute_max": 60}'.
- **what:** Per-site growth limits stored as a site_specs aspect: initial page target, weekly content/blog caps, absolute page maximum. Admin-overridable via the dashboard Direction tab; consumed by growth/budget calculations in planning.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#growth_config
- **relations:** site_specs store; page_growth_budget.go (referenced from 042 thunder step-zero).
- **verify-later:** budget enforcement in planner/discovery.

<!-- SOURCE: U19_sql_tables_components.md -->
### content_items reusable content layer
- **category:** content-governance
- **status-signal:** unknown
- **status-evidence:** Full DDL + helper get_component_content + v_content_usage view exist and page_components.content_item_id survives into the live dump, but no later file shows content_items being written.
- **what:** Separates "what to say" from "how to show it": typed reusable content rows (headline, tagline, service_description, testimonial, bio, cta, faq...) with semantic content_key, plain_text search, library sharing (site_id NULL + is_library + industry_vertical + library_tags), assets-style origin tracking, and status workflow. page_components reference a content_item with content_data acting as shallow-merge override (get_component_content). Would let one tagline appear in hero, footer and meta without duplication and let library content seed new sites.
- **sources:** docs/agent_docs/sql_for_tables/004_content_items.sql; docs/agent_docs/sql_for_tables/004b_content_items.md; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql#content_item_id
- **relations:** pages/page_components split; assets origin pattern.
- **verify-later:** content_items row count; any writer using content_item_id.

<!-- SOURCE: U19_sql_tables_components.md -->
### page_component_history full-snapshot content history
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** Created in Phase-0 Block A (019) explicitly as the replacement for the dropped content_snapshot/schema_snapshot columns.
- **what:** Before any content_data write to a page_component, the current value is copied here as a complete snapshot (not a diff) with source ('content-writer', 'section-editor', 'rollback'...) and triggering work item id — the rollback/audit substrate for section edits.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#3
- **relations:** replaces schema-mode snapshots; section-editor; site snapshots (page-level vs site-level revert).
- **verify-later:** writers copying into history before UPDATE.

<!-- SOURCE: U19_sql_tables_components.md -->
### Section governance columns: content_brief and suppressed_sections
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "content_brief: records the instructions that generated each component's content. Enables admins to see, edit, and regenerate" (008 tail); "suppressed_sections... prevents discovery checks from recreating sections that were intentionally removed. The DELETE component endpoint writes to this column" (003).
- **what:** Two small governance mechanisms: page_components.content_brief JSONB preserves the generation instructions per section for admin-editable regeneration; pages.suppressed_sections lists intentionally removed section functions so discovery does not resurrect them (component-removal flow Phases 2/5).
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#content-brief; docs/agent_docs/sql_for_tables/003_pages.sql#suppressed_sections
- **relations:** improvement-loop discovery checks; inline editing / regeneration.
- **verify-later:** DELETE component endpoint; discovery filters on suppressed_sections.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Page-content-writer + admin content brief regeneration flow
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** 100_content_page_build_handler_flow.md documents the live flow (page-build-handler → page-content-writer → load_site_specs → prepare_link_context → load_page_components → process_sections_loop) then states "What's missing: The prompt has no awareness of page_components.content_brief" and specifies the new content_rewrite flow.
- **what:** The bridge document into the modern era: the current work-item pipeline's content generation path, and the gap it fixes — admin edits a brief in the dashboard (page_components.content_brief), clicks Regenerate creating a content_rewrite work item, and the writer's generate_content prompt gains an "## Admin Content Brief" block for briefed sections while unbriefed sections behave as before.
- **sources:** docs001_flow_general/100_content_page_build_handler_flow.md
- **relations:** content-governance (briefs, regeneration — anchor doc 013); development-guide work-item lifecycle; content_components definitions.
- **verify-later:** page_components.content_brief column; content_rewrite work item type; load_page_section_components step.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Spatial addressing for natural-language editing
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** Vision in docs009/001 ("edit the paragraph on the left of the blue call to action... data-uuid and data-path attributes"); partially realized per docs018/008: "Component labeling (new) — injects data-pc-id, data-slot, data-position into each <section>".
- **what:** Every visible element carries a unique ID and genealogy path so an editing agent can resolve fuzzy human instructions spatially ("3rd paragraph on the left", "the one under the yellow button") by highlighting candidates iteratively. Shipped at section granularity (data-pc-id/data-slot/data-position mapping sections to page_components rows); element-level addressing and the conversational disambiguation loop remain unfulfilled.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-The-Spatial-Address-System; docs018_rerendering/008_granular_editing.md#What-Exists-Today; docs012_site_maps_and_components/006_start_concluding_links.md#2.4
- **relations:** section-editor agent; page_components.data_uuid/data_path columns; granular editing spectrum.
- **verify-later:** data_uuid/data_path columns on page_components; labeling injection code.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Flexible vs strict schema mode (approval snapshot)
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** docs017/001_flexible proposes sites.schema_mode + page_components.schema_snapshot/content_snapshot/component_version_id; docs018/008 lists "page_components.content_snapshot stores approved values... schema_mode can be strict or flexible per section" under "What Exists Today".
- **what:** Two-phase content lifecycle: initial builds run flexible (best-effort substitution, warnings on missing fields, creative freedom); at approval the section's input_schema and content values are snapshotted and the section moves to strict mode (edits validated against locked schema, unsubstituted placeholders fail, template upgrades can't break approved pages, rollback via content_snapshot). Open questions recorded: granularity, versioning vs snapshot, transition trigger, unlock capability.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/001_flexible_schema_enforcement.md; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-5; docs018_rerendering/008_granular_editing.md
- **relations:** locks (current descendant of the freeze idea); content-reviewer approval; section editor.
- **verify-later:** schema_mode/schema_snapshot/content_snapshot columns and whether any transition code exists.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Legal content agent + legal constraint rules
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** docs017/019b "legal-content-agent | Template + minimal LLM | New" with legal_rules JSON (required_disclaimers by trigger, forbidden_phrases, required_pages by jurisdiction template); compliance-discovery-agent phased for maintenance.
- **what:** Jurisdiction-aware legal pages from vetted templates (privacy-gdpr-uk etc.), plus machine-readable constraints exported to the content writer: disclaimers triggered by content conditions with placement rules, forbidden phrases per industry ("guaranteed returns"), required pages routed to the legal nav group. Compliance discovery monitors regulatory changes in maintenance.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Legal-Content-Agent; docs017_legacy_agent_rules_images_design_keydocs/022_maintenance_architecture_plan_v2.md#Discovery-Agents
- **relations:** policy-as-filters (org framework ancestor); finance site stress test; brand DNA forbidden phrases.
- **verify-later:** legal-content-agent definition; legal_rules key in sites.content_data.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Content review flow with rejection → needs_attention
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** docs011/003 flow diagram (validate → auto-eval vs HITL → approve/reject → finalize_hitl or mark_page needs_attention); docs018/003: "Rejected pages are picked up by maintenance workflow."
- **what:** The content-reviewer agent's dual-mode gate: algorithmic validation feeds either auto-evaluation (errors escalate) or HITL review (human sees issues, can edit HTML inline); outcomes are approve (deploy), approve-with-edits (use edited HTML), or reject (page marked needs_attention, skipped by the build loop, queued for maintenance). Established review as a first-class pipeline stage rather than an afterthought.
- **sources:** docs011_api_hitl/003_hitl_new_plan.md; docs018_rerendering/003_website_builder_architecture_status_report.md#4; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Active-Agents
- **relations:** HITL protocol; flexible/strict approval snapshot; maintenance pickup of needs_attention.
- **verify-later:** content-reviewer workflow JSON; needs_attention status handling in current loop.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Section editor agent (content_data is the source of truth)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** docs018/010 records decisions as implemented ("Decision: content_data is always the source of truth... We considered a lightweight html_patch edit type... decided against"; reused-code inventory; ActionInputSpec pattern used).
- **what:** Granular post-deploy editing without the full pipeline: two edit types — content_edit (field_updates merge or full content_data replace) and component_swap (new template, same content) — both updating page_components.content_data first, then re-rendering via buildRenderContextFromDB (reconstructs the full RenderContext purely from DB: site data, style collection, theme, nav, page meta, section content), reassembling the page, committing, and deploying. HTML patching was explicitly rejected because edits would vanish on the next rerender. Targets identified by page_component_id or page_name+slot_name (normalized). Future: LLM section rewrite via content_direction; bulk edits.
- **sources:** docs018_rerendering/010_section_editor_architecture.md; docs018_rerendering/008_granular_editing.md
- **relations:** granular editing spectrum; spatial addressing labels; rerender assemblePage; locks/inline editing (current descendants).
- **verify-later:** load_edit_context/apply_section_edit actions; section-editor agent definition.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Granular editing spectrum (word → multi-page)
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** docs018/008 table mapping six edit scopes to mechanisms and LLM needs; "What Exists Today" vs "What's Needed: One New Agent, Two New Actions"; content_direction JSONB "(new)".
- **what:** A routing model for edit requests by scope: word/phrase (direct patch — later rejected in favour of content_data edits), field value (template re-render), section rewrite (content-writer on one section), component swap, page rewrite (page-rebuild with content_direction instructions flowing into prompts), multi-page (maintenance-triage → page-rebuild). All routed through the same maintenance infrastructure.
- **sources:** docs018_rerendering/008_granular_editing.md; docs018_rerendering/010_section_editor_architecture.md#Future-Extensions
- **relations:** section editor; work items; content_direction column on pages.
- **verify-later:** content_direction column and its prompt integration.

<!-- SOURCE: U22_recent_small_docs.md -->
### maintenance_queue + claim/complete/fail functions
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** "MAINTENANCE QUEUE TABLE (for future use) ... For now, pages are flagged manually and the agent is triggered via generic-agent." Table + PL/pgSQL claim/complete/fail functions defined.
- **what:** A generic site-maintenance work queue (`maintenance_queue`) keyed by site_id with task_type (page_rebuild/css_update/nav_fix/link_repair/content_refresh), priority, JSONB payload, retry logic, and atomic `claim_maintenance_task`/`complete_maintenance_task`/`fail_maintenance_task` SQL functions using SKIP LOCKED. Later reused as the trigger surface for the chatbot install (`install_chat` task).
- **sources:** docs019_business/016_maintenance_queue_table.sql, docs019_business/017_maintenance_triage_agent.sql, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path
- **relations:** maintenance-triage agent, site-chat-installer, improvement-loop
- **verify-later:** maintenance_queue table + claim/complete/fail functions

<!-- SOURCE: U22_recent_small_docs.md -->
### maintenance_queue as generic install/uninstall trigger surface
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** Chatbot design reuses it: "Installation is requested by enqueuing a maintenance task — task_type='install_chat' on the existing maintenance_queue (which already has site_id, payload, status, retries)."
- **what:** The recognition that the existing `maintenance_queue` (built for page rebuilds) is a reusable, generic trigger surface for opt-in site add-ons — chat install/uninstall being the first — without touching the build pipeline. Establishes a pattern: new per-site capabilities enqueue a maintenance task rather than becoming a build-pipeline stage. (Cross-cut between docs019 infra and docs025 chatbot.)
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path, docs019_business/016_maintenance_queue_table.sql
- **relations:** maintenance_queue, site-chat-installer, maintenance-triage
- **verify-later:** maintenance_queue task_type values in use

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Recommendation-specialist architecture (bug vs recommendation vs gap)
- **category:** content-governance
- **status-signal:** abandoned
- **status-evidence:** PLAN_design-note-recommendation-specialists (April 2026): "Status: Proposed — not yet implemented"; "Deferred Until HITL queue becomes a bottleneck"; interim workaround is a manual `UPDATE site_work_items SET status='wont_fix'`.
- **what:** Proposal to make auditors emit an explicit `finding_type` (bug | recommendation | gap) so the pipeline stops auto-rewriting subjective opinions as if they were bugs (which caused false-positive rewrites like inventing a fake email). Bugs → `content_rewrite`; gaps → `needs_content_page` (P9, already deployed); recommendations → new specialist agents (identity-advisor, tone-shift-agent, content-strategist) returning apply/dismiss/escalate, gated by a per-site `sites.approval_mode` (auto | review). Never implemented.
- **sources:** old_design_and_styling/PLAN_design-note-recommendation-specialists.md#proposed-architecture, #per-site-approval-mode
- **relations:** HITL review flow; approval_mode; write_audit_findings_action.go; content-quality/visual-design auditors
- **verify-later:** write_audit_findings_action.go auditFinding struct; sites.approval_mode column existence

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Privacy posture (no cookies/JS/IP; UK GDPR/PECR)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** running_notes standing observations "Privacy posture (UK GDPR/PECR, low risk appetite): no cookies, no JS, no IP stored, referer reduced to host, country only from a coarse CDN header".
- **what:** A deliberate low-risk privacy stance baked into the engine and page: no cookies, no JavaScript, no IP stored, referer reduced to host only, country only from a coarse CDN header (CF-IPCountry). Open ingest choice: redact email/phone patterns at ingest vs rely on the 90-day prune.
- **sources:** traffic_probe_running_notes(27).md#standing-observations, traffic_probe_plan(11).md#p4, traffic_probe_runbook(12).md#6
- **relations:** shapes intent-probe (no-JS form), sanitisation, retention timer
- **verify-later:** engine handlers (no cookie/IP); site-engine-prune.timer RETENTION_DAYS=90

<!-- SOURCE: U25_leopardess_social.md -->
### site_specs `pinned` flag not honoured by the write path
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** REPLICATION §2: "WriteSiteSpec does not honour pinned … today the only thing protecting these specs is that improvement-sweep is disabled" (2026-07-10).
- **what:** Operator-corrected specs are marked pinned=true, but WriteSiteSpec supersedes unconditionally — pinned is an admin-display flag only. A pinned check in the spec write path is a named platform gap; the current protection is incidental (the improvement-sweep scheduler that fires content-gap-planner has been disabled since 2026-05-02).
- **sources:** docs/leopardessconsulting/REPLICATION_in_chassis.md#2; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-5
- **relations:** LLM fabrication classes; discovery/improvement-sweep disabled state
- **verify-later:** platform/orchestration/actions/site_spec_actions.go; scheduled_tasks row for improvement-sweep

<!-- SOURCE: U25_leopardess_social.md -->
### Section-editor: single-section edit propagation path
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** VERDICT §7: "the section-editor's first ever production run (3 orchestrations, all COMPLETED)" 2026-07-09; fix verified end-to-end on v1.0.1102 2026-07-10.
- **what:** The sanctioned route for a template change to reach one live section: apply_section_edit (edit_type content_edit) re-renders exactly one page_component from its (changed) template, rewrites rendered_html + content_data, reassembles the page from the other stored sections and commits — blast radius one section, unlike rerender_page_sections (all sections, gated on reason=image_landed/section_data_resolved) or assemble-only rerender_single_page. Doc 003's "HTML patching was rejected as an edit mechanism" is honoured by editing the source template (with a hand-written component_versions snapshot, since direct SQL is the only writer for arbitrary template edits) and letting the action re-render. Quirk: field_updates must be non-empty (a no-op merge re-supplying a current value works).
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#0, #1, #5; docs/social001_vonc_tiktok_social/minilobby_task/086_section_edit_provocation-card_vonc.sh (header)
- **relations:** page-rerender mode contract; build_status approved defect; mini-lobby trim
- **verify-later:** section_editor_actions.go apply_section_edit; orchestration_states for section-editor runs

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Content quality defect catalogue (gamesdesign) and work order
- **category:** content-quality
- **status-signal:** unknown
- **status-evidence:** "current as of 2026-06-09. Source of record: CATALOGUE_gamesdesign_post_sync_fix_defects(9).md"
- **what:** Five live defect classes on built pages: hero CTA text↔destination mismatch site-wide (lead item, spans content+linking); guide copy tool-flavoured; brand suffix leaking into card titles; empty footer brand/contact; empty tool descriptions. Work order: settle CTA field-vs-template, reuse component-template-fixer's CTA handling; then footer/titles/descriptions batch; then guide re-flavouring. Routing reality check flagged: the three-way finding classification and specialist agents are PROPOSED, not confirmed built.
- **sources:** FOCUS_content_quality.md (whole)
- **relations:** recommendation specialist architecture; internal linking; validate_page_content
- **verify-later:** whether identity-advisor/component-template-fixer CTA handling/sites.approval_mode exist

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### validate_page_content gate
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "the content validator and the gate … routes validate_content error_step → mark_needs_review → needs_human_review. (This is Mode 2 of the silent-completion work, already confirmed fixed.)" (2026-06-09)
- **what:** Blocker-detecting validator (placeholder text, unrendered templates, empty required sections, cross-site contamination) that any content fix must pass; failures now route consistently to human review. Known false-positive class: adopted content referencing the source domain trips the contamination heuristic (Bug 7 — needs an adopted-from whitelist for mode=recreate); legitimate emails (contactforsales.com) also flagged.
- **sources:** FOCUS_content_quality.md#machinery; HANDOFF_2026-04-23(1).md Bug 7; HANDOFF-pipeline-triage-april-2026.md#queue
- **relations:** silent completion mode 2; phantom-link check hook candidate
- **verify-later:** validate_page_content.go blocker classes; adopted-domain whitelist

<!-- SOURCE: U05_content_quality_linking.md -->
### Content-quality defect catalogue (gamesdesign.co.uk)
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** FOCUS_content_quality(2) 2026-06-10 status table: defect 1 "addressed (Step 1)", defects 2–5 "open"; HANDOFF_2026-06-15 §7 lists the parked items "next package, EXPECTED to recur on readopt".
- **what:** A maintained catalogue of content-correctness defects on built pages (CATALOGUE_gamesdesign_post_sync_fix_defects as source of record): hero CTA text↔destination mismatch, tool-flavoured guide copy, brand-suffix in card titles, empty footer brand-tagline/contact, empty tool descriptions, empty meta descriptions. Content quality is explicitly scoped as "words and per-component data" — distinct from design fidelity and from link destinations. Defects are worked as separate threads, each read-the-code-first.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality(2).md; HANDOFF_2026-06-15_index_stale_rebuild(2).md#7; running_notes_17_internal_linking_phantom_fixes(21).md#content-quality-observations
- **relations:** internal-linking through-line; brand-suffix leakage; site metadata fixer; guide copy re-flavouring; readopt as acceptance test.
- **verify-later:** CATALOGUE_gamesdesign_post_sync_fix_defects(9).md; live gamesdesign.co.uk pages; site_work_items for content_rewrite items.

<!-- SOURCE: U05_content_quality_linking.md -->
### validate_page_content deploy gate
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** FOCUS_content_quality(2) machinery §1 "verified this round"; running_notes_15(12) Part 10 "Mode 2 … ALREADY FIXED".
- **what:** The pre-deploy content validator: placeholder/template/contamination/email checks remain blockers (error → mark_needs_review → needs_human_review), while `validateInternalLinks` (now on datahelpers) flags `phantom_link` and `empty_internal_href` as non-blocking warnings, tolerating planned-but-unbuilt pages. Known gap: it does not flag brand-suffix titles, empty contact, or empty descriptions (content/spec issues, not link/placeholder issues).
- **sources:** FOCUS_content_quality(2).md#the-machinery; FOCUS_internal_linking(1).md#shared-machinery; running_notes_15(12).md#part-10
- **relations:** phantom policy; mark_needs_review; content-quality catalogue gaps.
- **verify-later:** validate_page_content.go; page-build-handler validate_content error_step.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Shared-component regen clobber failure mode (silent overwrite of dependent pages)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "RESOLVED + RECOVERED (verified)" (HANDOFF(7), 2026-07-04); "R6b pass 2026-07-03: distinct md5s, needles true"; root cause section dated + confirmed in NOTES §4.
- **what:** Regenerating a shared component overwrites its `input_schema`/`html_template` field contract in place without migrating dependent pages' `content_data` to the new field names; rendering binds by exact field name and silently empties misses, so every dependent renders a content-free shell that the assembler silently drops — fanning out across every page/site sharing the component. Confirmed on `system-stats` (`fdd92ad4`, regen 2026-06-24 15:06): 24 old keys vs 22 new, five live pages on three sites byte-identical empty. `content_data` stayed intact and per-page, so the breakage was recoverable without an LLM.
- **sources:** NOTES_component_regen_clobber(43).md §1, §4, §8; HANDOFF_component_regen_clobber(7).md §Incident 1; RUNBOOK_pre_cleanup_backup.md §The problem
- **relations:** F1 field-contract guard (fix); F3 scoped rerender (repair path); F5/F8 (sibling facets); RenderTemplate silent-empty mechanism; visible-content filter.
- **verify-later:** platform/orchestration/actions/store_generated_component_action.go (regen branch ~L354–432); content_components/component_versions/page_components tables; component `fdd92ad4-521a-4602-89cf-7ee1a66c10f1`.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Recovery playbook for stranded dependents (Route A rebuild vs Route B re-key + scoped re-render)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "R5b step 8 PASS — all five RECOVERED and verified" (NOTES §9ad, 2026-07-03); leopardess confirmed live by screenshot (§9ae).
- **what:** Two recovery routes for pages stranded by a shared-contract change: Route A — full `needs_page` writer rebuild (regenerates content_data under the new schema; simplest, costs LLM); Route B — re-key each page's `content_data` old→new (explicit reviewable jsonb_build_object mapping, dry-run first, CTAS backup, non-1:1 fields handled explicitly) then trigger the F3-scoped section re-render (no LLM, preserves per-page values). Route B executed for the five, doubling as F3's end-to-end proof; gated on fleet image, freshness check, and a cta-schema decision.
- **sources:** NOTES(43).md §6, §9q, §9s–§9t, §9ad; RUNBOOK(49).md Part A; PLAN(1).md Phase 4
- **relations:** F3; section readiness model (the cta_url blocker it hit); optimistic-lock co-management; snapshot-before-change.
- **verify-later:** page_components content_data keys for the five; backup tables page_components_bak_sysstats_20260702 / _briefexp_20260703 (may be dropped).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F8 — shared-component contamination: site-specific copy baked into shared machinery (three carriers)
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** "STEP-12 SWEEP PASSES — contamination cleared board-wide" (NOTES §9bg, 2026-07-06) — incident remediated; "F8 mitigation… WHAT: a guard/lint so shared-component fallbacks and llm_guidance must be site-neutral" still an open Part E flag.
- **what:** A pre-guard regen (2026-07-01) baked vonc's product pitch ("Spark", the daily Gauntlet) into the shared `brief-explanation` component via three carriers invisible to the name-only F1 guard: (1) static-field fallback values; (2) those values merged into dependents' content_data by the stored⊕resolved merge; (3) per-field `llm_guidance` — the strongest, actively instructing every future writer pass on any site to write vonc's product (reproduced verbatim on robot-hands and idea.uk; contamination also migrated into generated LLM copy on pages built pre-fix — the knock-on). Remediation playbook executed: snapshot v2/v3 → neutralize fallbacks (stats→llm optional; CTAs→neutral statics) → strip merged keys with CTAS backup → scoped F3 re-renders → writer rebuilds under cleaned guidance → board-wide strpos sweep (clean except vonc's own legitimate copy). Falsified along the way: field-description carrier, content_brief column, restore-v1 option (old-architecture contract). Proposed structural mitigation (unbuilt): store-time site-neutrality lint over fallbacks + llm_guidance.
- **sources:** NOTES(43).md §9an–§9bb, §9bg; RUNBOOK(49).md Part C + Part E F8; HANDOFF(7).md §Incident 2
- **relations:** F1 guard (its blind spot); llm_guidance surface; stored⊕resolved merge; neutralize-in-place remediation; D2b lint (same detection-net shape).
- **verify-later:** brief-explanation input_schema (neutral guidance ×11, stats source=llm no fallback); component_versions v1–v3; store-side lint absence.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Neutralize-in-place remediation pattern for contaminated shared components
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "v1 not a restore candidate… NEUTRALIZE-IN-PLACE chosen" (NOTES §9ao); Steps 1–2 landed with optimistic lock held (§9ap); Steps 6–7 landed §9bd.
- **what:** When a shared component's history offers no clean restore point (v1 predated the current field contract; restoring would regress dependents on the new architecture), the fix is surgical in-place neutralization: manual snapshot first, then targeted jsonb patches replacing only the offending attributes (fallbacks, guidance) under an optimistic lock, preserving names/types/structure — followed by per-dependent cleanup (strip merged keys, scoped re-renders, writer rebuilds) mapped per consumer (vonc's own copy untouched; robot-hands stripped; old-architecture pages escalated to rebuilds).
- **sources:** NOTES(43).md §9ao–§9aq, §9bb, §9bd; RUNBOOK(49).md Part C Steps 1–9
- **relations:** F8; optimistic-lock co-management; component versioning; recovery playbook.
- **verify-later:** the CTE jsonb_object_agg patch shape in RUNBOOK(49) Step 7 as reusable SQL.

<!-- SOURCE: U09_adoption.md -->
### Adoption content-quality defect families (polish batch)
- **category:** content-quality
- **status-signal:** unknown
- **status-evidence:** Open Groups 3 items as of HANDOFF_2026-06-09: "- GameDesign.uk brand-suffix in card titles; footer mailto/tagline empty; one empty tool description; guide tables render poorly; guides should cross-link to tools"; hero H1 reuse and empty meta descriptions from the catalogue remain untracked-as-fixed.
- **what:** The residual content-quality class after build mechanics were fixed: source-brand `<title>` suffixes used as display names (preserving the source brand, not the destination), empty footer contact/tagline (no graceful no-data path — components render empty structure instead of hiding), hero H1 duplicated across hubs, meta_description populated in DB but emitted empty, tool-flavoured guide copy (user open to real embedded interactive demos in guides), guide→tool cross-linking as enhancement.
- **sources:** CATALOGUE(9)#family-e, HANDOFF_2026-06-09#next-task, running_notes_14(25)#part-14n
- **relations:** silent-fallback link family; page-content-writer prompts; internal linking (024)
- **verify-later:** current gamesdesign deployed HTML; hero/footer component schemas' no-data paths

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Post-build validation of structured components (Fix D, unimplemented)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "Post-build validation (Fix D): assert a component whose input_schema declares a required structured field ... actually has it populated before deploy" — listed under open/structural, not done
- **what:** Proposed check that runs after a build, asserting that any component whose `input_schema` declares a required structured field actually has that field populated in `content_data`; if empty, flags the page instead of deploying a silently-empty component. Catches the bug class regardless of which planner or writer path produced the empty result.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#D-Post-build-validation, js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** FAQ duplicate content-surface bug; per-section briefs gap
- **verify-later:** grep/inspect `input_schema`; `content_data`

<!-- SOURCE: U15_docs019_running_notes.md -->
### Site-quality programme handoff
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** "site-quality programme HANDED OFF to its own runbook... 0 nav / 0 img / 0 svg / 0 script on ALL pages" (NOTES_running_synthesis_v4(39).md, 2026-07-06).
- **what:** Following the platform's first recorded domain→deployed-site milestone (dartsonline.com), a measured baseline (four rendered pages, all missing nav/images/svg/script, thin CSS variable usage, near-zero internal links) triggered a dedicated handoff (`RUNBOOK_site_quality.md`) splitting remaining work into stuck-dispatch (chrome/design/imagery), delivered-but-poor (content depth, links), and never-in-scope (feeds/RSS/graphics/games, disabled improvement-sweep) — with a live hypothesis that the relay path lacks the monolith's `render_site_components` chrome step, explaining nav-zero across every page.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-06 "MILESTONE recorded" and "site-quality programme HANDED OFF" entries.
- **relations:** Work-item relay / builder-generations architecture; diagnosis→fix loop workstream founding (the same "unresolved_cta" defect class recurs across both threads).

<!-- SOURCE: U19_sql_tables_components.md -->
### Placeholder-content suppression sweep
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** Executed SQL in 018: find deployed sections containing 'NEEDS HUMAN REVIEW'/'Lorem ipsum'/'[INSERT'/'<no value>', replace with hidden comment, create per-page placeholder_content items (handler 'human-review', status needs_human_review) plus per-site needs_rerender items.
- **what:** A validation pattern: placeholder or unreviewed text must never stay live — offending sections are hidden behind an HTML comment, a needs_human_review work item requests the real data (team names, photos...), and a rerender item republishes. Companion flows later resolve needs_section_data items as wont_fix when data arrives via site_specs (team, departments) or the section is dropped (pricing → engagement process).
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#placeholder-sweep and #075b-075e
- **relations:** work-item queue; site_specs identity enrichment; hitl approval.
- **verify-later:** validation agent producing these; recurrence of placeholder text.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Audited content pipeline (persona → research → draft → veracity/copyright audits)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** "Content Pipeline cannot be a single agent… Step 4 (Audit - Veracity)… Step 5 (Audit - Copyright)" (001); "Purifier Agent" and "Copywriters with Character" in the phase summary (014); site_persona step defined in 011.
- **what:** Content generation as an orchestrated sub-system: define a site persona/style guide, research via search/scrape adapters, persona-driven drafting, fact-check against research (separate agent, possible HITL), plagiarism/copyright audit (images only from licensed/free sources), then inject into template slots found by parsing data-function attributes. Motivated by veracity/copyright being "mission-critical legal and reputational risks".
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md#content-bottleneck; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** live successors: content-quality docs (content_quality_and_internal_linking), research-agents; persona idea → persona architecture across the platform.
- **verify-later:** whether any veracity/plagiarism audit step exists in the current content pipeline.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Content validation before review (validate_page_content)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** docs018/003: "validate_page_content action runs before review mode determination... Validation errors force HITL review, blocking auto-approval."
- **what:** Deterministic pre-review checks on generated pages: extract all hrefs and verify internal links against the pages table, verify emails against site contact data; errors (broken links) force human review while warnings flow through. Companion mechanisms: prepare_link_context injects an only-link-to-these-pages allowlist into writer prompts, and rerender-time contact injection replaces hallucinated phone/email with DB truth.
- **sources:** docs018_rerendering/003_website_builder_architecture_status_report.md#3; docs018_rerendering/002_summary_link_constraints.md; docs018_rerendering/003_website_builder_architecture_status_report.md#6
- **relations:** content-reviewer workflow; link_registry; content-quality internal linking (successor).
- **verify-later:** validate_page_content + prepare_link_context in registry; prompt inclusion of link_constraint_text ("Not Yet Done" at the time).

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Prompt composition asymmetry (text cascade vs image)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** loop_closure(9) decisions: "Step 5 image prompt cascade — Defer. Keep the single-prepend `imagery_direction` cascade for v1"; references `FOCUS_prompt_composition_pattern.md` — "the text pattern itself is fragile and shouldn't be copied wholesale … a composer step that produces a parameter envelope (for both text and images) is the strongest candidate."
- **what:** Deliberate design opinion that image prompts use only a single-prepend `imagery_direction` cascade, not the richer page-content-writer text composition — because the text cascade is considered fragile and a better target is a unified composer producing a parameter envelope for both text and images (likely landing in 2H, not a step-5 extension).
- **sources:** imagery/old/PLAN_imagery_loop_closure(9).md#decisions-taken, #image-prompt-cascade-deferred
- **relations:** live FOCUS_prompt_composition_pattern.md; image request shape (2H); directive cascade
- **verify-later:** composeImagePromptWithDirection; getImageryDirectionForSite

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Input sanitisation (sanitizeValue, Cc/Cf stripping, NFD survives)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "Sanitisation v2 … now strips Cc AND Cf … Real bug found by the new tests: checking IsControl before IsSpace silently JOINED words".
- **what:** The engine's `sanitizeValue()` strips control (Cc) and format (Cf: zero-widths, bidi overrides incl. U+202E, BOM, soft hyphen) chars, collapses whitespace runs (IsSpace checked FIRST to avoid joining words like `gmt\t\tmaster`→`gmtmaster`), and caps by RUNES not bytes (MaxValueLen semantic changed). NFD combining marks deliberately survive; NFC normalisation + lowercasing are deferred to the P4 collector (needs x/text; engine stays stdlib-only).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_running_notes(27).md#2026-06-12-debug-guide, traffic_probe_plan(11).md#p4
- **relations:** pairs with P4 ingest validation contract (NFC there)
- **verify-later:** service.go sanitizeValue; MAX_VALUE_LEN handling

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `component-template-fixer` CTA-handling reuse assumption — corrected, replaced by dedicated agent
- **category:** content-quality
- **status-signal:** superseded
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09): "The plan notes `component-template-fixer` 'already handles CTA fixes' — verify and extend rather than build new." Live `FOCUS_content_quality(2).md` (2026-06-10): "`component-template-fixer` exists but **explicitly punts on CTAs** (`cta_improvement`/`cta` → `fixed:false, action:'needs_review'`)... So the PLAN's 'already handles CTA fixes' was wrong; there was no CTA resolver to reuse — hence the dedicated `internal-link-resolver` (Step 3)."
- **what:** `PLAN_design-note-recommendation-specialists.md` asserted `component-template-fixer` already had CTA-fix handling that could be reused/extended for the hero-CTA phantom-link defect. Verification against the live agent's actual routing table found it explicitly declines CTA improvements, routing them to `needs_review` instead of fixing them. This wrong assumption, once corrected, directly motivated building a new dedicated agent (`internal-link-resolver`, see below) rather than extending the wrong one.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived); live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** internal-link-resolver agent (below); identity-advisor/sites.approval_mode (below)
- **verify-later:** `component-template-fixer`'s current action set.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `identity-advisor` agent and `sites.approval_mode` gate — proposed, confirmed never built
- **category:** content-quality
- **status-signal:** abandoned
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09) lists them as PROPOSED pieces needing verification ("Before relying on `identity-advisor` / `component-template-fixer` / `sites.approval_mode`, confirm each exists"). Live `FOCUS_content_quality(2).md`/`FOCUS_internal_linking(1).md` (2026-06-10) confirm: "`identity-advisor` does **not** exist. `sites.approval_mode` does **not** exist. The three-way `finding_type` classification and those specialists are PROPOSED, not built."
- **what:** `PLAN_design-note-recommendation-specialists.md`'s three-way finding-routing design (bug / gap / recommendation) named `identity-advisor` as the specialist for contact/email findings and `sites.approval_mode` as the gate for whether recommendation-type findings auto-apply. Neither was ever implemented — a clean case of a documented plan whose specific pieces were checked against the live schema/agent_definitions and found absent.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived) and live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** component-template-fixer CTA-reuse assumption (above)
- **verify-later:** re-check `agent_definitions` and the `sites` table for these names in case they were built later.

<!-- SOURCE: U25_leopardess_social.md -->
### LLM fabrication classes in self-built site content
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** AUDIT U1–U11 with removal dates ("Section DELETED 2026-07-10", "FAQ replaced 2026-07-10"); "Fabrication sweep, 2026-07-10 … CLEAN".
- **what:** Catalogue of what unconstrained content agents invented on a live site: fictional staff ("Peter Grenfell"), a nonexistent "8 departments" taxonomy, platform subsystems dressed as client case studies, AI agents listed as human team members with 404 portraits, capabilities that don't exist (Playwright scraping, proxy pools, circuit breakers, Helm/IAM), and misaligned stat suffixes ("99.9x uptime"). Removal required both spec rewrites and component deletion because some copy was baked into rendered_html.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#2; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-5, #Turn-11; docs/leopardessconsulting/scripts/L5_pages.sql (header)
- **relations:** claim-evidence audit rule; site_specs pinned gap (fabrications regenerate while specs are wrong)
- **verify-later:** page_components.content_data pattern sweep on all sites; content-gap-planner rewrite history in site_specs

<!-- SOURCE: U25_leopardess_social.md -->
### Anti-hype voice and claim-discipline spec
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** HANDOFF §3 "Specs — identity/voice/design_intent/portfolio rewritten (source_agent operator-rebuild, pinned)".
- **what:** A reusable voice contract for LLM-written site copy: positive framing (no strawmen, no competitor swipes), prefer the smaller exactly-true claim, plain language over compression, a banned-language list ("digital transformation", "leverage", "seamless"…), an LLM-tells-to-avoid list (reflexive triads, em-dash rhythm, summarising flourishes), CTA governance (name the next thing that happens; vary per page — repetition "signals content shallowness"), and honest uncertainty framing ("we have not done that one yet").
- **sources:** docs/leopardessconsulting/specs/voice.json; docs/leopardessconsulting/specs/identity.json#content_posture; docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql (header)
- **relations:** LLM fabrication classes; portfolio honest-labelling pattern ("Not yet done for a client")
- **verify-later:** site_specs aspect 'voice' for leopardess; whether content writers consume voice spec

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Content quality defect catalogue (gamesdesign) and work order
- **category:** content-quality
- **status-signal:** unknown
- **status-evidence:** "current as of 2026-06-09. Source of record: CATALOGUE_gamesdesign_post_sync_fix_defects(9).md"
- **what:** Five live defect classes on built pages: hero CTA text↔destination mismatch site-wide (lead item, spans content+linking); guide copy tool-flavoured; brand suffix leaking into card titles; empty footer brand/contact; empty tool descriptions. Work order: settle CTA field-vs-template, reuse component-template-fixer's CTA handling; then footer/titles/descriptions batch; then guide re-flavouring. Routing reality check flagged: the three-way finding classification and specialist agents are PROPOSED, not confirmed built.
- **sources:** FOCUS_content_quality.md (whole)
- **relations:** recommendation specialist architecture; internal linking; validate_page_content
- **verify-later:** whether identity-advisor/component-template-fixer CTA handling/sites.approval_mode exist

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### validate_page_content gate
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "the content validator and the gate … routes validate_content error_step → mark_needs_review → needs_human_review. (This is Mode 2 of the silent-completion work, already confirmed fixed.)" (2026-06-09)
- **what:** Blocker-detecting validator (placeholder text, unrendered templates, empty required sections, cross-site contamination) that any content fix must pass; failures now route consistently to human review. Known false-positive class: adopted content referencing the source domain trips the contamination heuristic (Bug 7 — needs an adopted-from whitelist for mode=recreate); legitimate emails (contactforsales.com) also flagged.
- **sources:** FOCUS_content_quality.md#machinery; HANDOFF_2026-04-23(1).md Bug 7; HANDOFF-pipeline-triage-april-2026.md#queue
- **relations:** silent completion mode 2; phantom-link check hook candidate
- **verify-later:** validate_page_content.go blocker classes; adopted-domain whitelist

<!-- SOURCE: U05_content_quality_linking.md -->
### Content-quality defect catalogue (gamesdesign.co.uk)
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** FOCUS_content_quality(2) 2026-06-10 status table: defect 1 "addressed (Step 1)", defects 2–5 "open"; HANDOFF_2026-06-15 §7 lists the parked items "next package, EXPECTED to recur on readopt".
- **what:** A maintained catalogue of content-correctness defects on built pages (CATALOGUE_gamesdesign_post_sync_fix_defects as source of record): hero CTA text↔destination mismatch, tool-flavoured guide copy, brand-suffix in card titles, empty footer brand-tagline/contact, empty tool descriptions, empty meta descriptions. Content quality is explicitly scoped as "words and per-component data" — distinct from design fidelity and from link destinations. Defects are worked as separate threads, each read-the-code-first.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality(2).md; HANDOFF_2026-06-15_index_stale_rebuild(2).md#7; running_notes_17_internal_linking_phantom_fixes(21).md#content-quality-observations
- **relations:** internal-linking through-line; brand-suffix leakage; site metadata fixer; guide copy re-flavouring; readopt as acceptance test.
- **verify-later:** CATALOGUE_gamesdesign_post_sync_fix_defects(9).md; live gamesdesign.co.uk pages; site_work_items for content_rewrite items.

<!-- SOURCE: U05_content_quality_linking.md -->
### validate_page_content deploy gate
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** FOCUS_content_quality(2) machinery §1 "verified this round"; running_notes_15(12) Part 10 "Mode 2 … ALREADY FIXED".
- **what:** The pre-deploy content validator: placeholder/template/contamination/email checks remain blockers (error → mark_needs_review → needs_human_review), while `validateInternalLinks` (now on datahelpers) flags `phantom_link` and `empty_internal_href` as non-blocking warnings, tolerating planned-but-unbuilt pages. Known gap: it does not flag brand-suffix titles, empty contact, or empty descriptions (content/spec issues, not link/placeholder issues).
- **sources:** FOCUS_content_quality(2).md#the-machinery; FOCUS_internal_linking(1).md#shared-machinery; running_notes_15(12).md#part-10
- **relations:** phantom policy; mark_needs_review; content-quality catalogue gaps.
- **verify-later:** validate_page_content.go; page-build-handler validate_content error_step.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Shared-component regen clobber failure mode (silent overwrite of dependent pages)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "RESOLVED + RECOVERED (verified)" (HANDOFF(7), 2026-07-04); "R6b pass 2026-07-03: distinct md5s, needles true"; root cause section dated + confirmed in NOTES §4.
- **what:** Regenerating a shared component overwrites its `input_schema`/`html_template` field contract in place without migrating dependent pages' `content_data` to the new field names; rendering binds by exact field name and silently empties misses, so every dependent renders a content-free shell that the assembler silently drops — fanning out across every page/site sharing the component. Confirmed on `system-stats` (`fdd92ad4`, regen 2026-06-24 15:06): 24 old keys vs 22 new, five live pages on three sites byte-identical empty. `content_data` stayed intact and per-page, so the breakage was recoverable without an LLM.
- **sources:** NOTES_component_regen_clobber(43).md §1, §4, §8; HANDOFF_component_regen_clobber(7).md §Incident 1; RUNBOOK_pre_cleanup_backup.md §The problem
- **relations:** F1 field-contract guard (fix); F3 scoped rerender (repair path); F5/F8 (sibling facets); RenderTemplate silent-empty mechanism; visible-content filter.
- **verify-later:** platform/orchestration/actions/store_generated_component_action.go (regen branch ~L354–432); content_components/component_versions/page_components tables; component `fdd92ad4-521a-4602-89cf-7ee1a66c10f1`.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Recovery playbook for stranded dependents (Route A rebuild vs Route B re-key + scoped re-render)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "R5b step 8 PASS — all five RECOVERED and verified" (NOTES §9ad, 2026-07-03); leopardess confirmed live by screenshot (§9ae).
- **what:** Two recovery routes for pages stranded by a shared-contract change: Route A — full `needs_page` writer rebuild (regenerates content_data under the new schema; simplest, costs LLM); Route B — re-key each page's `content_data` old→new (explicit reviewable jsonb_build_object mapping, dry-run first, CTAS backup, non-1:1 fields handled explicitly) then trigger the F3-scoped section re-render (no LLM, preserves per-page values). Route B executed for the five, doubling as F3's end-to-end proof; gated on fleet image, freshness check, and a cta-schema decision.
- **sources:** NOTES(43).md §6, §9q, §9s–§9t, §9ad; RUNBOOK(49).md Part A; PLAN(1).md Phase 4
- **relations:** F3; section readiness model (the cta_url blocker it hit); optimistic-lock co-management; snapshot-before-change.
- **verify-later:** page_components content_data keys for the five; backup tables page_components_bak_sysstats_20260702 / _briefexp_20260703 (may be dropped).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F8 — shared-component contamination: site-specific copy baked into shared machinery (three carriers)
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** "STEP-12 SWEEP PASSES — contamination cleared board-wide" (NOTES §9bg, 2026-07-06) — incident remediated; "F8 mitigation… WHAT: a guard/lint so shared-component fallbacks and llm_guidance must be site-neutral" still an open Part E flag.
- **what:** A pre-guard regen (2026-07-01) baked vonc's product pitch ("Spark", the daily Gauntlet) into the shared `brief-explanation` component via three carriers invisible to the name-only F1 guard: (1) static-field fallback values; (2) those values merged into dependents' content_data by the stored⊕resolved merge; (3) per-field `llm_guidance` — the strongest, actively instructing every future writer pass on any site to write vonc's product (reproduced verbatim on robot-hands and idea.uk; contamination also migrated into generated LLM copy on pages built pre-fix — the knock-on). Remediation playbook executed: snapshot v2/v3 → neutralize fallbacks (stats→llm optional; CTAs→neutral statics) → strip merged keys with CTAS backup → scoped F3 re-renders → writer rebuilds under cleaned guidance → board-wide strpos sweep (clean except vonc's own legitimate copy). Falsified along the way: field-description carrier, content_brief column, restore-v1 option (old-architecture contract). Proposed structural mitigation (unbuilt): store-time site-neutrality lint over fallbacks + llm_guidance.
- **sources:** NOTES(43).md §9an–§9bb, §9bg; RUNBOOK(49).md Part C + Part E F8; HANDOFF(7).md §Incident 2
- **relations:** F1 guard (its blind spot); llm_guidance surface; stored⊕resolved merge; neutralize-in-place remediation; D2b lint (same detection-net shape).
- **verify-later:** brief-explanation input_schema (neutral guidance ×11, stats source=llm no fallback); component_versions v1–v3; store-side lint absence.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Neutralize-in-place remediation pattern for contaminated shared components
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "v1 not a restore candidate… NEUTRALIZE-IN-PLACE chosen" (NOTES §9ao); Steps 1–2 landed with optimistic lock held (§9ap); Steps 6–7 landed §9bd.
- **what:** When a shared component's history offers no clean restore point (v1 predated the current field contract; restoring would regress dependents on the new architecture), the fix is surgical in-place neutralization: manual snapshot first, then targeted jsonb patches replacing only the offending attributes (fallbacks, guidance) under an optimistic lock, preserving names/types/structure — followed by per-dependent cleanup (strip merged keys, scoped re-renders, writer rebuilds) mapped per consumer (vonc's own copy untouched; robot-hands stripped; old-architecture pages escalated to rebuilds).
- **sources:** NOTES(43).md §9ao–§9aq, §9bb, §9bd; RUNBOOK(49).md Part C Steps 1–9
- **relations:** F8; optimistic-lock co-management; component versioning; recovery playbook.
- **verify-later:** the CTE jsonb_object_agg patch shape in RUNBOOK(49) Step 7 as reusable SQL.

<!-- SOURCE: U09_adoption.md -->
### Adoption content-quality defect families (polish batch)
- **category:** content-quality
- **status-signal:** unknown
- **status-evidence:** Open Groups 3 items as of HANDOFF_2026-06-09: "- GameDesign.uk brand-suffix in card titles; footer mailto/tagline empty; one empty tool description; guide tables render poorly; guides should cross-link to tools"; hero H1 reuse and empty meta descriptions from the catalogue remain untracked-as-fixed.
- **what:** The residual content-quality class after build mechanics were fixed: source-brand `<title>` suffixes used as display names (preserving the source brand, not the destination), empty footer contact/tagline (no graceful no-data path — components render empty structure instead of hiding), hero H1 duplicated across hubs, meta_description populated in DB but emitted empty, tool-flavoured guide copy (user open to real embedded interactive demos in guides), guide→tool cross-linking as enhancement.
- **sources:** CATALOGUE(9)#family-e, HANDOFF_2026-06-09#next-task, running_notes_14(25)#part-14n
- **relations:** silent-fallback link family; page-content-writer prompts; internal linking (024)
- **verify-later:** current gamesdesign deployed HTML; hero/footer component schemas' no-data paths

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Post-build validation of structured components (Fix D, unimplemented)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "Post-build validation (Fix D): assert a component whose input_schema declares a required structured field ... actually has it populated before deploy" — listed under open/structural, not done
- **what:** Proposed check that runs after a build, asserting that any component whose `input_schema` declares a required structured field actually has that field populated in `content_data`; if empty, flags the page instead of deploying a silently-empty component. Catches the bug class regardless of which planner or writer path produced the empty result.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#D-Post-build-validation, js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** FAQ duplicate content-surface bug; per-section briefs gap
- **verify-later:** grep/inspect `input_schema`; `content_data`

<!-- SOURCE: U15_docs019_running_notes.md -->
### Site-quality programme handoff
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** "site-quality programme HANDED OFF to its own runbook... 0 nav / 0 img / 0 svg / 0 script on ALL pages" (NOTES_running_synthesis_v4(39).md, 2026-07-06).
- **what:** Following the platform's first recorded domain→deployed-site milestone (dartsonline.com), a measured baseline (four rendered pages, all missing nav/images/svg/script, thin CSS variable usage, near-zero internal links) triggered a dedicated handoff (`RUNBOOK_site_quality.md`) splitting remaining work into stuck-dispatch (chrome/design/imagery), delivered-but-poor (content depth, links), and never-in-scope (feeds/RSS/graphics/games, disabled improvement-sweep) — with a live hypothesis that the relay path lacks the monolith's `render_site_components` chrome step, explaining nav-zero across every page.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-06 "MILESTONE recorded" and "site-quality programme HANDED OFF" entries.
- **relations:** Work-item relay / builder-generations architecture; diagnosis→fix loop workstream founding (the same "unresolved_cta" defect class recurs across both threads).

<!-- SOURCE: U19_sql_tables_components.md -->
### Placeholder-content suppression sweep
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** Executed SQL in 018: find deployed sections containing 'NEEDS HUMAN REVIEW'/'Lorem ipsum'/'[INSERT'/'<no value>', replace with hidden comment, create per-page placeholder_content items (handler 'human-review', status needs_human_review) plus per-site needs_rerender items.
- **what:** A validation pattern: placeholder or unreviewed text must never stay live — offending sections are hidden behind an HTML comment, a needs_human_review work item requests the real data (team names, photos...), and a rerender item republishes. Companion flows later resolve needs_section_data items as wont_fix when data arrives via site_specs (team, departments) or the section is dropped (pricing → engagement process).
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#placeholder-sweep and #075b-075e
- **relations:** work-item queue; site_specs identity enrichment; hitl approval.
- **verify-later:** validation agent producing these; recurrence of placeholder text.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Audited content pipeline (persona → research → draft → veracity/copyright audits)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** "Content Pipeline cannot be a single agent… Step 4 (Audit - Veracity)… Step 5 (Audit - Copyright)" (001); "Purifier Agent" and "Copywriters with Character" in the phase summary (014); site_persona step defined in 011.
- **what:** Content generation as an orchestrated sub-system: define a site persona/style guide, research via search/scrape adapters, persona-driven drafting, fact-check against research (separate agent, possible HITL), plagiarism/copyright audit (images only from licensed/free sources), then inject into template slots found by parsing data-function attributes. Motivated by veracity/copyright being "mission-critical legal and reputational risks".
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md#content-bottleneck; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** live successors: content-quality docs (content_quality_and_internal_linking), research-agents; persona idea → persona architecture across the platform.
- **verify-later:** whether any veracity/plagiarism audit step exists in the current content pipeline.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Content validation before review (validate_page_content)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** docs018/003: "validate_page_content action runs before review mode determination... Validation errors force HITL review, blocking auto-approval."
- **what:** Deterministic pre-review checks on generated pages: extract all hrefs and verify internal links against the pages table, verify emails against site contact data; errors (broken links) force human review while warnings flow through. Companion mechanisms: prepare_link_context injects an only-link-to-these-pages allowlist into writer prompts, and rerender-time contact injection replaces hallucinated phone/email with DB truth.
- **sources:** docs018_rerendering/003_website_builder_architecture_status_report.md#3; docs018_rerendering/002_summary_link_constraints.md; docs018_rerendering/003_website_builder_architecture_status_report.md#6
- **relations:** content-reviewer workflow; link_registry; content-quality internal linking (successor).
- **verify-later:** validate_page_content + prepare_link_context in registry; prompt inclusion of link_constraint_text ("Not Yet Done" at the time).

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Prompt composition asymmetry (text cascade vs image)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** loop_closure(9) decisions: "Step 5 image prompt cascade — Defer. Keep the single-prepend `imagery_direction` cascade for v1"; references `FOCUS_prompt_composition_pattern.md` — "the text pattern itself is fragile and shouldn't be copied wholesale … a composer step that produces a parameter envelope (for both text and images) is the strongest candidate."
- **what:** Deliberate design opinion that image prompts use only a single-prepend `imagery_direction` cascade, not the richer page-content-writer text composition — because the text cascade is considered fragile and a better target is a unified composer producing a parameter envelope for both text and images (likely landing in 2H, not a step-5 extension).
- **sources:** imagery/old/PLAN_imagery_loop_closure(9).md#decisions-taken, #image-prompt-cascade-deferred
- **relations:** live FOCUS_prompt_composition_pattern.md; image request shape (2H); directive cascade
- **verify-later:** composeImagePromptWithDirection; getImageryDirectionForSite

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Input sanitisation (sanitizeValue, Cc/Cf stripping, NFD survives)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "Sanitisation v2 … now strips Cc AND Cf … Real bug found by the new tests: checking IsControl before IsSpace silently JOINED words".
- **what:** The engine's `sanitizeValue()` strips control (Cc) and format (Cf: zero-widths, bidi overrides incl. U+202E, BOM, soft hyphen) chars, collapses whitespace runs (IsSpace checked FIRST to avoid joining words like `gmt\t\tmaster`→`gmtmaster`), and caps by RUNES not bytes (MaxValueLen semantic changed). NFD combining marks deliberately survive; NFC normalisation + lowercasing are deferred to the P4 collector (needs x/text; engine stays stdlib-only).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_running_notes(27).md#2026-06-12-debug-guide, traffic_probe_plan(11).md#p4
- **relations:** pairs with P4 ingest validation contract (NFC there)
- **verify-later:** service.go sanitizeValue; MAX_VALUE_LEN handling

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `component-template-fixer` CTA-handling reuse assumption — corrected, replaced by dedicated agent
- **category:** content-quality
- **status-signal:** superseded
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09): "The plan notes `component-template-fixer` 'already handles CTA fixes' — verify and extend rather than build new." Live `FOCUS_content_quality(2).md` (2026-06-10): "`component-template-fixer` exists but **explicitly punts on CTAs** (`cta_improvement`/`cta` → `fixed:false, action:'needs_review'`)... So the PLAN's 'already handles CTA fixes' was wrong; there was no CTA resolver to reuse — hence the dedicated `internal-link-resolver` (Step 3)."
- **what:** `PLAN_design-note-recommendation-specialists.md` asserted `component-template-fixer` already had CTA-fix handling that could be reused/extended for the hero-CTA phantom-link defect. Verification against the live agent's actual routing table found it explicitly declines CTA improvements, routing them to `needs_review` instead of fixing them. This wrong assumption, once corrected, directly motivated building a new dedicated agent (`internal-link-resolver`, see below) rather than extending the wrong one.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived); live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** internal-link-resolver agent (below); identity-advisor/sites.approval_mode (below)
- **verify-later:** `component-template-fixer`'s current action set.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `identity-advisor` agent and `sites.approval_mode` gate — proposed, confirmed never built
- **category:** content-quality
- **status-signal:** abandoned
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09) lists them as PROPOSED pieces needing verification ("Before relying on `identity-advisor` / `component-template-fixer` / `sites.approval_mode`, confirm each exists"). Live `FOCUS_content_quality(2).md`/`FOCUS_internal_linking(1).md` (2026-06-10) confirm: "`identity-advisor` does **not** exist. `sites.approval_mode` does **not** exist. The three-way `finding_type` classification and those specialists are PROPOSED, not built."
- **what:** `PLAN_design-note-recommendation-specialists.md`'s three-way finding-routing design (bug / gap / recommendation) named `identity-advisor` as the specialist for contact/email findings and `sites.approval_mode` as the gate for whether recommendation-type findings auto-apply. Neither was ever implemented — a clean case of a documented plan whose specific pieces were checked against the live schema/agent_definitions and found absent.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived) and live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** component-template-fixer CTA-reuse assumption (above)
- **verify-later:** re-check `agent_definitions` and the `sites` table for these names in case they were built later.

<!-- SOURCE: U25_leopardess_social.md -->
### LLM fabrication classes in self-built site content
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** AUDIT U1–U11 with removal dates ("Section DELETED 2026-07-10", "FAQ replaced 2026-07-10"); "Fabrication sweep, 2026-07-10 … CLEAN".
- **what:** Catalogue of what unconstrained content agents invented on a live site: fictional staff ("Peter Grenfell"), a nonexistent "8 departments" taxonomy, platform subsystems dressed as client case studies, AI agents listed as human team members with 404 portraits, capabilities that don't exist (Playwright scraping, proxy pools, circuit breakers, Helm/IAM), and misaligned stat suffixes ("99.9x uptime"). Removal required both spec rewrites and component deletion because some copy was baked into rendered_html.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#2; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-5, #Turn-11; docs/leopardessconsulting/scripts/L5_pages.sql (header)
- **relations:** claim-evidence audit rule; site_specs pinned gap (fabrications regenerate while specs are wrong)
- **verify-later:** page_components.content_data pattern sweep on all sites; content-gap-planner rewrite history in site_specs

<!-- SOURCE: U25_leopardess_social.md -->
### Anti-hype voice and claim-discipline spec
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** HANDOFF §3 "Specs — identity/voice/design_intent/portfolio rewritten (source_agent operator-rebuild, pinned)".
- **what:** A reusable voice contract for LLM-written site copy: positive framing (no strawmen, no competitor swipes), prefer the smaller exactly-true claim, plain language over compression, a banned-language list ("digital transformation", "leverage", "seamless"…), an LLM-tells-to-avoid list (reflexive triads, em-dash rhythm, summarising flourishes), CTA governance (name the next thing that happens; vary per page — repetition "signals content shallowness"), and honest uncertainty framing ("we have not done that one yet").
- **sources:** docs/leopardessconsulting/specs/voice.json; docs/leopardessconsulting/specs/identity.json#content_posture; docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql (header)
- **relations:** LLM fabrication classes; portfolio honest-labelling pattern ("Not yet done for a client")
- **verify-later:** site_specs aspect 'voice' for leopardess; whether content writers consume voice spec

<!-- SOURCE: U01_docs024_numbered_core.md -->
### News feed pipeline (sources → async ingest → triage → JSON render → commit)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 006 status table all ✅ with gaswholesalers evidence; scheduled 6-hourly
- **what:** content-feed-trigger finds recommended sites → content-feed-orchestrator: seed_sources (from classification spec) → dispatch ingesters async per due source (rss/news_search/api_news/scrape) → feed-triage scores PRIOR runs' items → render latest-news.json (6) + news-archive.json (20) → git commit. Two-pass by design (ingest now, triage next run). Homepage snippet + /news.html listing page both client-fetch the JSON — news updates decoupled from page rerender.
- **sources:** 006 full
- **relations:** growth budget; content-gap-planner chain; source diversity
- **verify-later:** content_sources/content_feed_items; content-feed-refresh task

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Feed triage: relevance + credibility + source-attribution provenance
- **category:** news-feed-pipeline
- **status-signal:** partial
- **status-evidence:** 006 deployed, but Known Open Issues: "credibility always 0 … fields exist but aren't being populated"
- **what:** LLM triage scores relevance 0-100 and credibility high/medium/low with attribution chain {original_source, found_via, source_tier} across a 6-tier source taxonomy; rejects fabricated URLs, nav links, uncorroborated low-credibility claims. Status lifecycle ingested→relevant/review/rejected→expired(30d).
- **sources:** 006#feed-triage, #Issues; #Resolved Decisions 47
- **relations:** diversity scoring plan; Grok provider choice
- **verify-later:** credibility population bug fixed?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Real-time-search news providers (Grok Responses API decision)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 006 resolved decision 48; hallucinated-URL bug table entry
- **what:** api_news sources route to xAI grok-4-1-fast via Responses API with web_search+x_search, OpenAI Responses (gpt-4.1-mini) or Perplexity sonar — all real-time search; chat-completions grok-3-mini hallucinated 2023 URLs and was dropped.
- **sources:** 006#fetch_llm_news provider routing
- **relations:** feed triage credibility
- **verify-later:** provider keys in personae-default-secrets

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Render source-diversity interleaving
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 006 status table ✅; decision recorded after single-source domination
- **what:** loadNewsItems uses ROW_NUMBER() OVER (PARTITION BY source_id) ordered by source_rank then recency so each source contributes at most ~2 of 6 display slots; with topic-focused sources this also yields topical diversity.
- **sources:** 006#Render action source diversity, #Content Diversity §6
- **relations:** topic-focused source splitting (planned)
- **verify-later:** render_news_section_action.go query

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Content diversity & original research pipeline (readership segments, timelines, scenario analysis, engagement)
- **category:** news-feed-pipeline
- **status-signal:** aspirational
- **status-evidence:** 006 "What's Not Built Yet" lists every piece (topic splitting ready-to-implement; article-rewriter/feed-publisher/feed-lifecycle blocked/unbuilt)
- **what:** Planned evolution: topic-focused source splitting (SQL-only), coverage-gap pre-fetch step, multi-language regional discovery with triage translation, triage diversity scoring, research-agent multi-step investigations (fact/history/quotes/numbers) → writer targeted per readership segment (procurement/ops/trading/strategy) → eval agent quality gate → publish; continuous annotated timelines with pattern recognition; if/then scenario analysis (no predictions); client-side engagement measurement feeding content planning.
- **sources:** 006#Expansion Roadmap, #Content Diversity & Research Pipeline
- **relations:** research-agents; batch API integration
- **verify-later:** none built — check for article-rewriter definition

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### News publishing gap (curation → deployed posts)
- **category:** news-feed-pipeline
- **status-signal:** aspirational
- **status-evidence:** "News — pipeline exists, publishing doesn't … The pipeline ends at curation" (2026-05)
- **what:** Ingestion/triage/diversity produce latest-news.json per site but nothing turns curated items into deployed blog posts; Path B connects news ingestion to page deployment via page-content-writer with a news-feed input, passing the site's deployed tool list for cross-linking.
- **sources:** FOCUS_interactive_content_generation(4).md#News, #Path-B
- **relations:** feed triage fixes; topic splitting
- **verify-later:** whether an article-publishing step now exists in news pipeline

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Feed triage scoring repair (config reads + wrapper unwrap)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** "Triage is working. As of last check: 41 relevant, 23 rejected, 232 ingested (backlog clearing at 15 items per cycle)" (2026-04-17)
- **what:** 200+ items unscored since April 2nd due to three stacked bugs: LLM output truncation (max_items 50 → 15; max_tokens → 8192), config literal invisible to inputs.GetInt (use GetIntField on StepConfig.Config), and the execute_llm_prompt wrapper map ({type,result}) never unwrapped. Topic splitting of the single Grok source into topic-focused sources planned (SQL-only).
- **sources:** HANDOFF_2026-04-17_triage_and_component_linking.md#1, #4
- **relations:** chassis input conventions; content-feed-trigger workflow bug
- **verify-later:** feed_triage_actions.go; content_feed_items backlog state

<!-- SOURCE: U10_imagery.md -->
### News pipeline replication and the news enrichment pattern
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** "News pipeline: LIVE and healthy (replicate gas)" (2026-05-20 recon); robot-hands "9 content_sources seeded, 0 erroring" (2026-07-10).
- **what:** The live chain content-feed-trigger → content-feed-orchestrator → feed-ingester → feed-triage → render_news_section (→ /data/latest-news.json + news-archive.json), with content_sources rows of four parallel types (rss, news_search, api_news with grok/web-search tools, scrape) as the replication template — pure data rows, no new code. Adding news to an existing site is enrichment, not re-plan: evaluate_news_feed writes classification.content_features.news_feed, news-section-addition amends the plan (RULE 11 places latest-news on the homepage). Two distinct components serve it (latest-news card grid on index; news-listing full page). Item expiry happens via status transition; the expires_at column exists but is unpopulated.
- **sources:** HANDOFF_robot_hands_rebuild.md#PIPELINE-RECON, old/README_news_pipeline.md, PLAN_imagery_best_in_class.md#Phase-I0-status
- **relations:** deploy_page files_field dependency (news JS silently dropped otherwise); news imagery (I5) builds on it.
- **verify-later:** robot-hands content_sources rows; content_feed_items lifecycle counts.

<!-- SOURCE: U10_imagery.md -->
### Price-news TTL and news→infographic enhancements
- **category:** news-feed-pipeline
- **status-signal:** aspirational
- **status-evidence:** "Both (a) and (b) are NICE-TO-HAVE… backlog" (2026-05-20, user-stated not urgent).
- **what:** (a) Price-aware filtering with short expiry: classify fetched news for price-movement items and expire them after 1–2 days via the existing-but-unused expires_at column plus a topics-based triage tag; per-site vertical. (b) News→infographic: pick 1–2 items, research the subject, generate an infographic — ties into the imagery infographic kind, research adapters, and the data-graph pipeline when data-driven.
- **sources:** HANDOFF_robot_hands_rebuild.md#NEWS-ENHANCEMENTS, old/README_news_pipeline.md
- **relations:** data-graph pipeline; news imagery I5 partially absorbs (b).
- **verify-later:** expires_at population; any price-tagging triage rule.

<!-- SOURCE: U12_docs024_archives.md -->
### "Insights section" as the Tier-2 news-feed expansion target
- **category:** news-feed-pipeline
- **status-signal:** superseded
- **status-evidence:** Archive Tier 2 = "Insights section... Future"; live Tier 2 = "News listing page... ✅ Working," curated/rewritten-article idea folded into Tier 3.
- **what:** The original three-tier roadmap treated a dedicated `/insights/` section of rewritten, curated articles as the second expansion tier after homepage snippets. When the archive-first news-index/listing page was actually built, it took the Tier-2 slot instead, and the "curated rewritten articles" idea was pushed down into Tier 3, where `article-rewriter` and `feed-publisher` remain listed as not-yet-built in both versions.
- **sources:** old/older1/006_news_feed_pipeline.md#"Expansion Roadmap"; docs024_key_docs_latest/006_news_feed_pipeline_v2.md#"Expansion Roadmap"
- **relations:** article-rewriter/feed-publisher agents (still unbuilt)
- **verify-later:** check whether a `/insights/` route or `article-rewriter` agent definition exists anywhere.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### News rendering three-layer architecture (data / behaviour / structure+style)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** "Discovered and fixed 2026-05-19/20 during the gaswholesalers.com news rollout"; FOCUS doc header "Consolidated from ... the fix-plan half of findings_and_plan_news_visual.md"
- **what:** News (and any data-driven component) rendering splits into three independently produced/deployed layers: Data (content_feed_items → /data/*.json via render_news_section), Behaviour (content_components.js_content → /tools/assets/{function}.js via rerender_single_page), and Structure+style (html_template + css_snippets, inlined per page). They connect only at runtime in the browser via fetch. This separation is deliberate and is what allows multiple independent news views per site.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Rendering-layer, js_snippets_news_gaswholesalers/old/006_news_feed_pipeline_addendum_rendering.md
- **relations:** files_field deploy mechanism; two-news-components pattern; component asset coupling gap
- **verify-later:** render_news_section action, rerender_single_page action, content_components.js_content/html_template columns, css_snippets table

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two distinct news components as a multi-view pattern (latest-news vs news-listing)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** table of both components' function/JS/data pairing verified live on gaswholesalers.com
- **what:** `latest-news` (homepage card grid, curated top 6) and `news-listing` (full archive list) are two separate content_components rows, each with its own template, JS, and data file — not duplicates. The architecture generalizes: adding a new filtered/styled news view requires only a new content_components row + CSS + a data-producing step; the deploy mechanism is generic over component function name, no workflow change needed.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#The-two-news-components-and-their-pairings, js_snippets_news_gaswholesalers/old/002_how_webdesign_handles_snippets.md
- **relations:** News rendering three-layer architecture; component asset coupling gap
- **verify-later:** content_components rows id 77dafa26 (latest-news), 11d4dc21 (news-listing)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### files_field vs content_field git_commit deploy bug (component JS assets silently dropped)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** "Verification: Single-page rerenders of index and news both returned files_count: 2" — fixed via jsonb_set, no code change, structural (site-wide) fix
- **what:** The `page-rerender` workflow's `deploy_page` step was configured with `content_field: "rendered_page.html"` (HTML only) instead of `files_field: "rendered_page.files"` (HTML + all component JS). `git_commit`'s `extractFilesForGit` had three extraction methods and the wrong one was selected, so every component's `js_content` was computed but discarded before ever reaching git — for every site, since inception. Fixed by a config-only jsonb_set edit; applies structurally to all components/sites, present and future.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Resolved, js_snippets_news_gaswholesalers/old/006_news_feed_pipeline_addendum_rendering.md, js_snippets_news_gaswholesalers/TODO_remaining_work.md#Done
- **relations:** News rendering three-layer architecture; rendered_html snapshot-not-view pattern
- **verify-later:** page-rerender agent_definition deploy_page step config, git_commit action extractFilesForGit

<!-- SOURCE: U13_docs024_small_dirs.md -->
### rerender-pages refresh-flag coupling (three concerns behind one flag)
- **category:** news-feed-pipeline
- **status-signal:** aspirational
- **status-evidence:** "Future improvement (note for backlog)... Low effort, modest value, do it next time we touch rerender-pages versions" — not implemented as of TODO_remaining_work.md
- **what:** The `rerender-pages` workflow ties three conceptually-independent refresh operations (site components re-render, JS snippets rebuild+deploy, blog-listing rebuild) behind a single `refresh_site_components` boolean. Proposed fix: split into three independent flags.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#rerender-pages-workflow-findings, js_snippets_news_gaswholesalers/old/rerender_pages_workflow_findings.md
- **relations:** rebuild_blog_listing news-index gap; two rerender paths
- **verify-later:** grep/inspect `rerender-pages`; `refresh_site_components`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### rebuild_blog_listing does not handle news-index pages (silent no-op gap)
- **category:** news-feed-pipeline
- **status-signal:** partial
- **status-evidence:** "the step is a silent no-op. It logs 'No blog page found, skipping' ... news visuals do get updated [via later page_rerender]. But there is no equivalent news-listing rebuild"
- **what:** `RebuildBlogListingAction`'s `findBlogPage` only matches `page_type='blog-index'` or `name='blog' AND page_type='content'` — never `page_type='news-index'` or `name='news'`. On news-only sites the step silently no-ops; would need a parallel `rebuild_news_listing`/`findNewsPage`.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Finding-1, js_snippets_news_gaswholesalers/old/rerender_pages_workflow_findings.md
- **relations:** rerender-pages refresh-flag coupling
- **verify-later:** grep/inspect `RebuildBlogListingAction`; `findBlogPage`; `page_type='blog-index'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two rerender trigger paths (site-wide work-item batch vs single-page orchestration-only)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF doc "Key facts to carry forward: Two rerender paths: site-wide rerender-pages creates site_work_items (item_type=page_rerender); single-page page-rerender is an orchestration only (no work item)"
- **what:** Site-wide `rerender-pages` creates `site_work_items` rows (batch, dispatched over time, load-bearing on reaper/OOM fragility); single-page `page-rerender` is triggered as a direct orchestration with no work-item row, used for quick manual/test verification of a fix.
- **sources:** js_snippets_news_gaswholesalers/old/HANDOFF_2026-05-21_faq_prevention_and_news.md
- **relations:** rerender-pages refresh-flag coupling; reaper mechanisms and gap
- **verify-later:** grep/inspect `rerender-pages`; `site_work_items`; `page-rerender`

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Blog-listing / orphan-page routing session handoff
- **category:** news-feed-pipeline
- **status-signal:** partial
- **status-evidence:** 102_blog_handoff header "Session Handoff — April 10 2026"; "Ready to Deploy (files generated, not yet applied)"
- **what:** A dated operational handoff fixing blog-listing rendering (slot-name mismatch, empty-schema CSS-only template, missing article links) and reclassifying orphan pages into three routes (blog-post→rerender, nav-flags→nav-drift→nav-updater, no-nav→needs_internal_links→internal-linker). Documents self-hosted GitHub Actions runner deploy, the page-build-handler `error_step`-placement fix (46 validation crashes), the dedup pattern, and a future Mistral-Small-on-CPU internal-linker.
- **sources:** ED/102_blog_handoff-2026-04-10.md#completed-this-session, ED/102_blog_handoff-2026-04-10.md#ready-to-deploy-files-generated-not-yet-applied, ED/102_blog_handoff-2026-04-10.md#remaining-unresolved-groups-not-yet-addressed
- **relations:** work item lifecycle/unresolved; deployment-github; nav sync; link management
- **verify-later:** rebuild_blog_listing_action.go; check_orphan_pages.go; github-actions-runner; nav-updater/internal-linker

<!-- SOURCE: U18_sql_for_agents.md -->
### News feed pipeline (feed-ingester, content-feed-orchestrator, feed-triage, content-feed-trigger, latest-news component)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 087–090 definitions; 100 portfolio: "Four source types operational. Credibility scoring... Six-hour refresh cycles... Live news sections deployed on production sites."
- **what:** Per-site news: content-feed-trigger is a 6-hour heartbeat that finds sites whose classification spec recommends a news feed (content_features.news_feed.recommended) and needing refresh, dispatching content-feed-orchestrator per site; feed-ingester fetches one source (RSS / news search / LLM news / scrape, routed by source_type) into content_feed_items; feed-triage (initially a stub) scores relevance/credibility; the latest-news content component is data-driven — rendered by the render_news_section Go action, not the LLM writer — with CSS from theme variables; 113's redesign migrated news components to contract-003 without regex (split_part/position/substring surgery).
- **sources:** 087_feeds_triage_ingester_orchestrator_etc.sql; 089_latest_news.sql; 090_b_content_feed_trigger.sql; 090_content_feed_orchestrator.sql; 113_site_asset_renderer.sql
- **relations:** content_sources/content_feed_items tables; scheduler-and-tasks; site_specs classification aspect
- **verify-later:** feed-triage real implementation vs stub; render_news_section action

<!-- SOURCE: U19_sql_tables_components.md -->
### News feed pipeline: content_sources and feed-item lifecycle
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** content_sources DDL (twice-iterated) + seed_boxing_sources function; content_feed_items in 018; applied handler-routing fixes and content-feed-refresh task (6h); live Grok config update to grok-4-1-fast with search_tools for gaswholesalers.
- **what:** Per-site content sources with typed configs — news_search (web search adapter), rss, api_news (LLM news via xAI/Grok incl. prompt_template, hours_lookback, search_tools), scrape (Firecrawl), api_data (structured APIs like BoE rates) — scheduled by fetch_interval/next_fetch_at with error tracking. Fetched items flow through content_feed_items' separate lifecycle (ingested→filtered→relevant→queued→published/rejected/expired/duplicate) with per-site relevance scoring, entity cross-referencing and dedup, becoming a site_work_items row only at publish time. Routing contract: missing_news_sources / stale_news_section / all_sources_erroring → content-feed-orchestrator; missing_news_section → content-gap-planner.
- **sources:** docs/agent_docs/sql_for_tables/027_content_sources_table.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#028_news_feed_handler_routing_fixes
- **relations:** work queue; latest-news client rendering; scheduler.
- **verify-later:** content-feed-orchestrator workflow; feed item volumes.

<!-- SOURCE: U19_sql_tables_components.md -->
### Client-side latest-news JSON rendering
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** Applied rendered_html update installing the IIFE that fetches /data/latest-news.json (headline, subheadline, items[title,url,summary,source,date], insights_url) on gaswholesalers' index; 044 adds formatNewsDate and the redesigned news CSS.
- **what:** News sections render client-side from a static JSON artefact deployed alongside the site (/data/latest-news.json), so news refresh is a data commit, not a page rebuild. Component ships noscript fallback, date humanisation (formatNewsDate expanding "2d ago"), and canonical CSS in css_snippets picked up on the next webdesign run.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#news-feed-js; docs/agent_docs/sql_for_tables/044_css_snippets.sql
- **relations:** med JSON export (same static-JSON publishing pattern); css/js snippets.
- **verify-later:** JSON writer for /data/latest-news.json; per-site adoption.

<!-- SOURCE: U21_legacy_docs_b.md -->
### News & content feed pipeline (mid-era design)
- **category:** news-feed-pipeline
- **status-signal:** superseded
- **status-evidence:** v1 in docs017/030, refined in 019b (sub-agents feed-ingester/deduplicator/triage/article-rewriter/publisher/lifecycle), restructured in 023 ("Feed items go through ingestion, filtering, deduplication and relevance scoring before they become publishable" with work_item linkage); today's news-feed-pipeline (docs024 006) is the deployed descendant.
- **what:** Per-site content sources (RSS/API/scrape/entity_event) polled on configurable intervals → raw content_feed_items → dedup (near-duplicate headline detection) → LLM triage (relevance, urgency, angle for THIS site) → article-rewriter producing original articles in site voice with entity cross-links and required disclaimers → publication as pages → time-based lifecycle decay (featured 0-24h → current → aging → archive → prune, with per-site-type pacing and event-calendar coupling). Later revision: publishable items become site_work_items (handler article-writer) and rewritten articles become entities; news display is a design concern owned by the component/theme system, not the feed.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/030_news_feeds_v1.md; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#7-News-and-Content-Feed-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Content-Feed-Items
- **relations:** entity news_triggers; work items; current news-feed-pipeline (successor, incl. diversity concerns).
- **verify-later:** content_sources/content_feed_items current schema vs these designs.

<!-- SOURCE: U23_docs_root_vonc.md -->
### News feed pipeline as the proven data-layer template
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-25 ~18:00: "all confirmed active v1.0.1078" — read from agent definitions and action source.
- **what:** content-feed-trigger (scheduled heartbeat 6h via scheduled_tasks name='content-feed-refresh'; finds news-recommended sites, spawns content-feed-orchestrator per site, max 5) → orchestrator (seed_content_sources → dispatch_feed_sources → feed-ingester per due source [rss/scrape/news_search/api_news] → feed-triage LLM relevance+credibility scoring) → render_news_section (loads items, expires stale, builds JSON from a Go struct, produces an archive JSON if a news-index page exists) → git_commit `/data/latest-news.json`. The latest-news component fetches the JSON via its own extracted component JS (Path 1); the news-date-formatter snippet is only a helper. This is the platform's model for any static-site runtime data feed.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00 + #2026-06-29-~17:20 (PD-1); docs/PLAN_spark_provocation_pipeline.md#architecture
- **relations:** Phase-3 provocation pipeline (clone); scheduler-and-tasks (scheduled_tasks.name)
- **verify-later:** render_news_section_action.go; content-feed-* agent_definitions; scheduled_tasks row content-feed-refresh

<!-- SOURCE: U01_docs024_numbered_core.md -->
### News feed pipeline (sources → async ingest → triage → JSON render → commit)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 006 status table all ✅ with gaswholesalers evidence; scheduled 6-hourly
- **what:** content-feed-trigger finds recommended sites → content-feed-orchestrator: seed_sources (from classification spec) → dispatch ingesters async per due source (rss/news_search/api_news/scrape) → feed-triage scores PRIOR runs' items → render latest-news.json (6) + news-archive.json (20) → git commit. Two-pass by design (ingest now, triage next run). Homepage snippet + /news.html listing page both client-fetch the JSON — news updates decoupled from page rerender.
- **sources:** 006 full
- **relations:** growth budget; content-gap-planner chain; source diversity
- **verify-later:** content_sources/content_feed_items; content-feed-refresh task

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Feed triage: relevance + credibility + source-attribution provenance
- **category:** news-feed-pipeline
- **status-signal:** partial
- **status-evidence:** 006 deployed, but Known Open Issues: "credibility always 0 … fields exist but aren't being populated"
- **what:** LLM triage scores relevance 0-100 and credibility high/medium/low with attribution chain {original_source, found_via, source_tier} across a 6-tier source taxonomy; rejects fabricated URLs, nav links, uncorroborated low-credibility claims. Status lifecycle ingested→relevant/review/rejected→expired(30d).
- **sources:** 006#feed-triage, #Issues; #Resolved Decisions 47
- **relations:** diversity scoring plan; Grok provider choice
- **verify-later:** credibility population bug fixed?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Real-time-search news providers (Grok Responses API decision)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 006 resolved decision 48; hallucinated-URL bug table entry
- **what:** api_news sources route to xAI grok-4-1-fast via Responses API with web_search+x_search, OpenAI Responses (gpt-4.1-mini) or Perplexity sonar — all real-time search; chat-completions grok-3-mini hallucinated 2023 URLs and was dropped.
- **sources:** 006#fetch_llm_news provider routing
- **relations:** feed triage credibility
- **verify-later:** provider keys in personae-default-secrets

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Render source-diversity interleaving
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 006 status table ✅; decision recorded after single-source domination
- **what:** loadNewsItems uses ROW_NUMBER() OVER (PARTITION BY source_id) ordered by source_rank then recency so each source contributes at most ~2 of 6 display slots; with topic-focused sources this also yields topical diversity.
- **sources:** 006#Render action source diversity, #Content Diversity §6
- **relations:** topic-focused source splitting (planned)
- **verify-later:** render_news_section_action.go query

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Content diversity & original research pipeline (readership segments, timelines, scenario analysis, engagement)
- **category:** news-feed-pipeline
- **status-signal:** aspirational
- **status-evidence:** 006 "What's Not Built Yet" lists every piece (topic splitting ready-to-implement; article-rewriter/feed-publisher/feed-lifecycle blocked/unbuilt)
- **what:** Planned evolution: topic-focused source splitting (SQL-only), coverage-gap pre-fetch step, multi-language regional discovery with triage translation, triage diversity scoring, research-agent multi-step investigations (fact/history/quotes/numbers) → writer targeted per readership segment (procurement/ops/trading/strategy) → eval agent quality gate → publish; continuous annotated timelines with pattern recognition; if/then scenario analysis (no predictions); client-side engagement measurement feeding content planning.
- **sources:** 006#Expansion Roadmap, #Content Diversity & Research Pipeline
- **relations:** research-agents; batch API integration
- **verify-later:** none built — check for article-rewriter definition

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### News publishing gap (curation → deployed posts)
- **category:** news-feed-pipeline
- **status-signal:** aspirational
- **status-evidence:** "News — pipeline exists, publishing doesn't … The pipeline ends at curation" (2026-05)
- **what:** Ingestion/triage/diversity produce latest-news.json per site but nothing turns curated items into deployed blog posts; Path B connects news ingestion to page deployment via page-content-writer with a news-feed input, passing the site's deployed tool list for cross-linking.
- **sources:** FOCUS_interactive_content_generation(4).md#News, #Path-B
- **relations:** feed triage fixes; topic splitting
- **verify-later:** whether an article-publishing step now exists in news pipeline

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Feed triage scoring repair (config reads + wrapper unwrap)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** "Triage is working. As of last check: 41 relevant, 23 rejected, 232 ingested (backlog clearing at 15 items per cycle)" (2026-04-17)
- **what:** 200+ items unscored since April 2nd due to three stacked bugs: LLM output truncation (max_items 50 → 15; max_tokens → 8192), config literal invisible to inputs.GetInt (use GetIntField on StepConfig.Config), and the execute_llm_prompt wrapper map ({type,result}) never unwrapped. Topic splitting of the single Grok source into topic-focused sources planned (SQL-only).
- **sources:** HANDOFF_2026-04-17_triage_and_component_linking.md#1, #4
- **relations:** chassis input conventions; content-feed-trigger workflow bug
- **verify-later:** feed_triage_actions.go; content_feed_items backlog state

<!-- SOURCE: U10_imagery.md -->
### News pipeline replication and the news enrichment pattern
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** "News pipeline: LIVE and healthy (replicate gas)" (2026-05-20 recon); robot-hands "9 content_sources seeded, 0 erroring" (2026-07-10).
- **what:** The live chain content-feed-trigger → content-feed-orchestrator → feed-ingester → feed-triage → render_news_section (→ /data/latest-news.json + news-archive.json), with content_sources rows of four parallel types (rss, news_search, api_news with grok/web-search tools, scrape) as the replication template — pure data rows, no new code. Adding news to an existing site is enrichment, not re-plan: evaluate_news_feed writes classification.content_features.news_feed, news-section-addition amends the plan (RULE 11 places latest-news on the homepage). Two distinct components serve it (latest-news card grid on index; news-listing full page). Item expiry happens via status transition; the expires_at column exists but is unpopulated.
- **sources:** HANDOFF_robot_hands_rebuild.md#PIPELINE-RECON, old/README_news_pipeline.md, PLAN_imagery_best_in_class.md#Phase-I0-status
- **relations:** deploy_page files_field dependency (news JS silently dropped otherwise); news imagery (I5) builds on it.
- **verify-later:** robot-hands content_sources rows; content_feed_items lifecycle counts.

<!-- SOURCE: U10_imagery.md -->
### Price-news TTL and news→infographic enhancements
- **category:** news-feed-pipeline
- **status-signal:** aspirational
- **status-evidence:** "Both (a) and (b) are NICE-TO-HAVE… backlog" (2026-05-20, user-stated not urgent).
- **what:** (a) Price-aware filtering with short expiry: classify fetched news for price-movement items and expire them after 1–2 days via the existing-but-unused expires_at column plus a topics-based triage tag; per-site vertical. (b) News→infographic: pick 1–2 items, research the subject, generate an infographic — ties into the imagery infographic kind, research adapters, and the data-graph pipeline when data-driven.
- **sources:** HANDOFF_robot_hands_rebuild.md#NEWS-ENHANCEMENTS, old/README_news_pipeline.md
- **relations:** data-graph pipeline; news imagery I5 partially absorbs (b).
- **verify-later:** expires_at population; any price-tagging triage rule.

<!-- SOURCE: U12_docs024_archives.md -->
### "Insights section" as the Tier-2 news-feed expansion target
- **category:** news-feed-pipeline
- **status-signal:** superseded
- **status-evidence:** Archive Tier 2 = "Insights section... Future"; live Tier 2 = "News listing page... ✅ Working," curated/rewritten-article idea folded into Tier 3.
- **what:** The original three-tier roadmap treated a dedicated `/insights/` section of rewritten, curated articles as the second expansion tier after homepage snippets. When the archive-first news-index/listing page was actually built, it took the Tier-2 slot instead, and the "curated rewritten articles" idea was pushed down into Tier 3, where `article-rewriter` and `feed-publisher` remain listed as not-yet-built in both versions.
- **sources:** old/older1/006_news_feed_pipeline.md#"Expansion Roadmap"; docs024_key_docs_latest/006_news_feed_pipeline_v2.md#"Expansion Roadmap"
- **relations:** article-rewriter/feed-publisher agents (still unbuilt)
- **verify-later:** check whether a `/insights/` route or `article-rewriter` agent definition exists anywhere.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### News rendering three-layer architecture (data / behaviour / structure+style)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** "Discovered and fixed 2026-05-19/20 during the gaswholesalers.com news rollout"; FOCUS doc header "Consolidated from ... the fix-plan half of findings_and_plan_news_visual.md"
- **what:** News (and any data-driven component) rendering splits into three independently produced/deployed layers: Data (content_feed_items → /data/*.json via render_news_section), Behaviour (content_components.js_content → /tools/assets/{function}.js via rerender_single_page), and Structure+style (html_template + css_snippets, inlined per page). They connect only at runtime in the browser via fetch. This separation is deliberate and is what allows multiple independent news views per site.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Rendering-layer, js_snippets_news_gaswholesalers/old/006_news_feed_pipeline_addendum_rendering.md
- **relations:** files_field deploy mechanism; two-news-components pattern; component asset coupling gap
- **verify-later:** render_news_section action, rerender_single_page action, content_components.js_content/html_template columns, css_snippets table

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two distinct news components as a multi-view pattern (latest-news vs news-listing)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** table of both components' function/JS/data pairing verified live on gaswholesalers.com
- **what:** `latest-news` (homepage card grid, curated top 6) and `news-listing` (full archive list) are two separate content_components rows, each with its own template, JS, and data file — not duplicates. The architecture generalizes: adding a new filtered/styled news view requires only a new content_components row + CSS + a data-producing step; the deploy mechanism is generic over component function name, no workflow change needed.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#The-two-news-components-and-their-pairings, js_snippets_news_gaswholesalers/old/002_how_webdesign_handles_snippets.md
- **relations:** News rendering three-layer architecture; component asset coupling gap
- **verify-later:** content_components rows id 77dafa26 (latest-news), 11d4dc21 (news-listing)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### files_field vs content_field git_commit deploy bug (component JS assets silently dropped)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** "Verification: Single-page rerenders of index and news both returned files_count: 2" — fixed via jsonb_set, no code change, structural (site-wide) fix
- **what:** The `page-rerender` workflow's `deploy_page` step was configured with `content_field: "rendered_page.html"` (HTML only) instead of `files_field: "rendered_page.files"` (HTML + all component JS). `git_commit`'s `extractFilesForGit` had three extraction methods and the wrong one was selected, so every component's `js_content` was computed but discarded before ever reaching git — for every site, since inception. Fixed by a config-only jsonb_set edit; applies structurally to all components/sites, present and future.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Resolved, js_snippets_news_gaswholesalers/old/006_news_feed_pipeline_addendum_rendering.md, js_snippets_news_gaswholesalers/TODO_remaining_work.md#Done
- **relations:** News rendering three-layer architecture; rendered_html snapshot-not-view pattern
- **verify-later:** page-rerender agent_definition deploy_page step config, git_commit action extractFilesForGit

<!-- SOURCE: U13_docs024_small_dirs.md -->
### rerender-pages refresh-flag coupling (three concerns behind one flag)
- **category:** news-feed-pipeline
- **status-signal:** aspirational
- **status-evidence:** "Future improvement (note for backlog)... Low effort, modest value, do it next time we touch rerender-pages versions" — not implemented as of TODO_remaining_work.md
- **what:** The `rerender-pages` workflow ties three conceptually-independent refresh operations (site components re-render, JS snippets rebuild+deploy, blog-listing rebuild) behind a single `refresh_site_components` boolean. Proposed fix: split into three independent flags.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#rerender-pages-workflow-findings, js_snippets_news_gaswholesalers/old/rerender_pages_workflow_findings.md
- **relations:** rebuild_blog_listing news-index gap; two rerender paths
- **verify-later:** grep/inspect `rerender-pages`; `refresh_site_components`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### rebuild_blog_listing does not handle news-index pages (silent no-op gap)
- **category:** news-feed-pipeline
- **status-signal:** partial
- **status-evidence:** "the step is a silent no-op. It logs 'No blog page found, skipping' ... news visuals do get updated [via later page_rerender]. But there is no equivalent news-listing rebuild"
- **what:** `RebuildBlogListingAction`'s `findBlogPage` only matches `page_type='blog-index'` or `name='blog' AND page_type='content'` — never `page_type='news-index'` or `name='news'`. On news-only sites the step silently no-ops; would need a parallel `rebuild_news_listing`/`findNewsPage`.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Finding-1, js_snippets_news_gaswholesalers/old/rerender_pages_workflow_findings.md
- **relations:** rerender-pages refresh-flag coupling
- **verify-later:** grep/inspect `RebuildBlogListingAction`; `findBlogPage`; `page_type='blog-index'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two rerender trigger paths (site-wide work-item batch vs single-page orchestration-only)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF doc "Key facts to carry forward: Two rerender paths: site-wide rerender-pages creates site_work_items (item_type=page_rerender); single-page page-rerender is an orchestration only (no work item)"
- **what:** Site-wide `rerender-pages` creates `site_work_items` rows (batch, dispatched over time, load-bearing on reaper/OOM fragility); single-page `page-rerender` is triggered as a direct orchestration with no work-item row, used for quick manual/test verification of a fix.
- **sources:** js_snippets_news_gaswholesalers/old/HANDOFF_2026-05-21_faq_prevention_and_news.md
- **relations:** rerender-pages refresh-flag coupling; reaper mechanisms and gap
- **verify-later:** grep/inspect `rerender-pages`; `site_work_items`; `page-rerender`

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Blog-listing / orphan-page routing session handoff
- **category:** news-feed-pipeline
- **status-signal:** partial
- **status-evidence:** 102_blog_handoff header "Session Handoff — April 10 2026"; "Ready to Deploy (files generated, not yet applied)"
- **what:** A dated operational handoff fixing blog-listing rendering (slot-name mismatch, empty-schema CSS-only template, missing article links) and reclassifying orphan pages into three routes (blog-post→rerender, nav-flags→nav-drift→nav-updater, no-nav→needs_internal_links→internal-linker). Documents self-hosted GitHub Actions runner deploy, the page-build-handler `error_step`-placement fix (46 validation crashes), the dedup pattern, and a future Mistral-Small-on-CPU internal-linker.
- **sources:** ED/102_blog_handoff-2026-04-10.md#completed-this-session, ED/102_blog_handoff-2026-04-10.md#ready-to-deploy-files-generated-not-yet-applied, ED/102_blog_handoff-2026-04-10.md#remaining-unresolved-groups-not-yet-addressed
- **relations:** work item lifecycle/unresolved; deployment-github; nav sync; link management
- **verify-later:** rebuild_blog_listing_action.go; check_orphan_pages.go; github-actions-runner; nav-updater/internal-linker

<!-- SOURCE: U18_sql_for_agents.md -->
### News feed pipeline (feed-ingester, content-feed-orchestrator, feed-triage, content-feed-trigger, latest-news component)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 087–090 definitions; 100 portfolio: "Four source types operational. Credibility scoring... Six-hour refresh cycles... Live news sections deployed on production sites."
- **what:** Per-site news: content-feed-trigger is a 6-hour heartbeat that finds sites whose classification spec recommends a news feed (content_features.news_feed.recommended) and needing refresh, dispatching content-feed-orchestrator per site; feed-ingester fetches one source (RSS / news search / LLM news / scrape, routed by source_type) into content_feed_items; feed-triage (initially a stub) scores relevance/credibility; the latest-news content component is data-driven — rendered by the render_news_section Go action, not the LLM writer — with CSS from theme variables; 113's redesign migrated news components to contract-003 without regex (split_part/position/substring surgery).
- **sources:** 087_feeds_triage_ingester_orchestrator_etc.sql; 089_latest_news.sql; 090_b_content_feed_trigger.sql; 090_content_feed_orchestrator.sql; 113_site_asset_renderer.sql
- **relations:** content_sources/content_feed_items tables; scheduler-and-tasks; site_specs classification aspect
- **verify-later:** feed-triage real implementation vs stub; render_news_section action

<!-- SOURCE: U19_sql_tables_components.md -->
### News feed pipeline: content_sources and feed-item lifecycle
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** content_sources DDL (twice-iterated) + seed_boxing_sources function; content_feed_items in 018; applied handler-routing fixes and content-feed-refresh task (6h); live Grok config update to grok-4-1-fast with search_tools for gaswholesalers.
- **what:** Per-site content sources with typed configs — news_search (web search adapter), rss, api_news (LLM news via xAI/Grok incl. prompt_template, hours_lookback, search_tools), scrape (Firecrawl), api_data (structured APIs like BoE rates) — scheduled by fetch_interval/next_fetch_at with error tracking. Fetched items flow through content_feed_items' separate lifecycle (ingested→filtered→relevant→queued→published/rejected/expired/duplicate) with per-site relevance scoring, entity cross-referencing and dedup, becoming a site_work_items row only at publish time. Routing contract: missing_news_sources / stale_news_section / all_sources_erroring → content-feed-orchestrator; missing_news_section → content-gap-planner.
- **sources:** docs/agent_docs/sql_for_tables/027_content_sources_table.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#028_news_feed_handler_routing_fixes
- **relations:** work queue; latest-news client rendering; scheduler.
- **verify-later:** content-feed-orchestrator workflow; feed item volumes.

<!-- SOURCE: U19_sql_tables_components.md -->
### Client-side latest-news JSON rendering
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** Applied rendered_html update installing the IIFE that fetches /data/latest-news.json (headline, subheadline, items[title,url,summary,source,date], insights_url) on gaswholesalers' index; 044 adds formatNewsDate and the redesigned news CSS.
- **what:** News sections render client-side from a static JSON artefact deployed alongside the site (/data/latest-news.json), so news refresh is a data commit, not a page rebuild. Component ships noscript fallback, date humanisation (formatNewsDate expanding "2d ago"), and canonical CSS in css_snippets picked up on the next webdesign run.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#news-feed-js; docs/agent_docs/sql_for_tables/044_css_snippets.sql
- **relations:** med JSON export (same static-JSON publishing pattern); css/js snippets.
- **verify-later:** JSON writer for /data/latest-news.json; per-site adoption.

<!-- SOURCE: U21_legacy_docs_b.md -->
### News & content feed pipeline (mid-era design)
- **category:** news-feed-pipeline
- **status-signal:** superseded
- **status-evidence:** v1 in docs017/030, refined in 019b (sub-agents feed-ingester/deduplicator/triage/article-rewriter/publisher/lifecycle), restructured in 023 ("Feed items go through ingestion, filtering, deduplication and relevance scoring before they become publishable" with work_item linkage); today's news-feed-pipeline (docs024 006) is the deployed descendant.
- **what:** Per-site content sources (RSS/API/scrape/entity_event) polled on configurable intervals → raw content_feed_items → dedup (near-duplicate headline detection) → LLM triage (relevance, urgency, angle for THIS site) → article-rewriter producing original articles in site voice with entity cross-links and required disclaimers → publication as pages → time-based lifecycle decay (featured 0-24h → current → aging → archive → prune, with per-site-type pacing and event-calendar coupling). Later revision: publishable items become site_work_items (handler article-writer) and rewritten articles become entities; news display is a design concern owned by the component/theme system, not the feed.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/030_news_feeds_v1.md; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#7-News-and-Content-Feed-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Content-Feed-Items
- **relations:** entity news_triggers; work items; current news-feed-pipeline (successor, incl. diversity concerns).
- **verify-later:** content_sources/content_feed_items current schema vs these designs.

<!-- SOURCE: U23_docs_root_vonc.md -->
### News feed pipeline as the proven data-layer template
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-25 ~18:00: "all confirmed active v1.0.1078" — read from agent definitions and action source.
- **what:** content-feed-trigger (scheduled heartbeat 6h via scheduled_tasks name='content-feed-refresh'; finds news-recommended sites, spawns content-feed-orchestrator per site, max 5) → orchestrator (seed_content_sources → dispatch_feed_sources → feed-ingester per due source [rss/scrape/news_search/api_news] → feed-triage LLM relevance+credibility scoring) → render_news_section (loads items, expires stale, builds JSON from a Go struct, produces an archive JSON if a news-index page exists) → git_commit `/data/latest-news.json`. The latest-news component fetches the JSON via its own extracted component JS (Path 1); the news-date-formatter snippet is only a helper. This is the platform's model for any static-site runtime data feed.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00 + #2026-06-29-~17:20 (PD-1); docs/PLAN_spark_provocation_pipeline.md#architecture
- **relations:** Phase-3 provocation pipeline (clone); scheduler-and-tasks (scheduled_tasks.name)
- **verify-later:** render_news_section_action.go; content-feed-* agent_definitions; scheduled_tasks row content-feed-refresh

<!-- SOURCE: U05_content_quality_linking.md -->
### Traffic-probe capture backend (misfiled in this unit)
- **category:** traffic-analytics
- **status-signal:** unknown
- **status-evidence:** Code present with reconciled architecture comments ("Division of labour after the chassis reconciliation"); no deployment claim in this unit.
- **what:** A stdlib-only Go service capturing visitor intent on probe domains: POST /intent (search/categories/freetext events), GET /api/hit (1x1 no-cookie visit beacon), key-gated /stats, host-keyed JSON store forked from idea.uk; no IPs stored, referer reduced to host (UK GDPR/PECR posture). The chassis builds/serves the static probe pages; nginx proxies only the capture paths. Sits in this unit's golang_code/ by accident of filing — belongs to the traffic-probe concept area.
- **sources:** golang_code/service.go, store.go, main.go (headers)
- **relations:** docs024 traffic_probe unit; idea.uk store pattern.
- **verify-later:** deployment on probe VMs; overlap with docs024/traffic_probe docs (canonical home).

<!-- SOURCE: U11_traffic_probe.md -->
### Traffic-probe mission — intent discovery on parked domains
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** HANDOFF (≈2026-06-13): "Engine (site-engine, stdlib Go) live on a dedicated EU box for relojistas.com … HTTPS, capturing"; plan: P0–P2 done, P3/P4 in progress, P5 not started.
- **what:** Domains that still receive residual visitors but serve only a parking lander are put on a minimal "probe" page that plausibly reflects the old vertical and invites ONE action (search box / category links / free-text). The stated intent is captured server-side; after 2–4 weeks the terms rank which domains have real demand worth building an idea.uk-style site for. Explicit scope boundary: capture what visitors *say they want* on our own page, never recover anyone's old gated content.
- **sources:** TASK_traffic_probe_brief.md#1-2, traffic_probe_plan(12).md#how-it-all-fits, traffic_probe_runbook(13).md#0
- **relations:** probe page pattern, ranking queries + graduation criteria, VM-hosted backend sites class
- **verify-later:** live relojistas.com/stats; intent_events table row counts; sites rows with deploy_config.target='vm'

<!-- SOURCE: U11_traffic_probe.md -->
### Wayback grounding of probe pages
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-11: "Relojistas grounded in the snapshot: it was a Spanish watch FORUM"; 2026-06-13(b): grounding constraint recorded.
- **what:** Before building a probe page, look up what the domain used to be via archive.org (CDX path list, availability API, snapshot view). The old path list (/login, /members, /forum…) signals what was gated and what visitors still want; the snapshot fixes language, vertical, and the invited action. Operational constraint discovered: Claude can web_fetch archive pages only when a search surfaces the exact URL and cannot enumerate CDX on demand — so the operator supplies Wayback URLs/snapshots, or grounding falls back to web search + the domain name.
- **sources:** TASK_traffic_probe_brief.md#2-method, traffic_probe_running_notes(28).md#2026-06-13-b, HANDOFF#thread-c
- **relations:** per-domain notes convention; adoption-pipeline (site recreation from crawl) is the platform cousin
- **verify-later:** archive.org.results/ snapshots exist for both live domains (they do, in this unit)

<!-- SOURCE: U11_traffic_probe.md -->
### Probe page pattern — one invited action, plausible framing
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas live 2026-06-12 (first capture 13:03:44 UTC); wayfaringlondoner page "built + grounded … not yet deployed" (HANDOFF).
- **what:** The page looks intentional, not parked: a one-line tagline matching the old vertical, exactly one invited action (v1: a single text input, kind=search or freetext), a plain privacy line, a 1×1 beacon, no JS, no cookies. Framing follows the domain's heritage — relojistas is a Spanish marketplace/search posture (marca/modelo/reparación/compraventa, thanks at /gracias.html); wayfaringlondoner is an English BLOG posture asking for a destination/story. Hand-made pages for the first domains were explicitly a go-live unblocker; chassis-built pages take over under P3.
- **sources:** TASK_traffic_probe_brief.md#2, relojistas_notes(8).md#decisions, wayfaringlondoner_notes.md#decisions, relojistas_golive/index.html
- **relations:** intent-probe component (the library form of the same pattern), probe content restraint
- **verify-later:** live page HTML vs intent-probe component render

<!-- SOURCE: U11_traffic_probe.md -->
### Minimal-data privacy posture (UK GDPR/PECR)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes "Standing observations": "no cookies, no JS, no IP stored, referer reduced to host, country only from a coarse CDN header"; holds "regardless of volume" (relojistas_notes).
- **what:** Server-side-only logging, no third-party trackers, no non-essential cookies (nothing stored on the device → no consent banner needed), no names/emails collected, free-text treated as potentially personal and not retained longer than needed, plain privacy line on every page. Explicitly declared load-invariant: under traffic pressure the project will not add client-side JS, third-party analytics, or IP logging. Open choice: redact email/phone patterns at ingest vs rely on the 90-day prune.
- **sources:** TASK_traffic_probe_brief.md#4, relojistas_notes(8).md#traffic-handling, traffic_probe_running_notes(28).md#standing-observations
- **relations:** intent event record, ingest validation contract, content-governance (platform-wide posture cousin)
- **verify-later:** intent-probe component privacy_text fallback; engine code stores no IP/UA

<!-- SOURCE: U11_traffic_probe.md -->
### Intent event record (fields and deliberate omissions)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas_notes "What we record" + FIRST LIVE CAPTURE log entry 2026-06-12 13:03:44 UTC.
- **what:** One event per submission: id, host, kind (search|categories|freetext), value (typed text ≤500 runes), ref_host (referer reduced to bare host, blank if same-site), country (coarse CDN header or empty), created_at (UTC), plus landing_query (inbound ?q=/?utm= that survived to the form page — added 2026-06-13 so the structured export carries it without a log join). Deliberately NOT recorded: IP addresses, user agents, cookies, full referer URLs, names/emails. There is no results page: the probe performs no search; the submission itself is the product (303 → thanks page).
- **sources:** relojistas_notes(8).md#what-we-record, traffic_probe_running_notes(28).md#2026-06-13 (landing_query), intent_events_migration(1).sql
- **relations:** minimal-data privacy posture, /events export, intent_events table
- **verify-later:** IntentEvent struct in site-engine repo; events-*.jsonl line shape on box

<!-- SOURCE: U11_traffic_probe.md -->
### Visit beacon and events-per-1k metric
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** runbook §2 "the page must include the visit beacon"; running notes self-correction 2026-06-11 (beacon removed from gracias page).
- **what:** A no-JS, no-cookie 1×1 `<img src="/api/hit">` counts human-with-browser visits per host — the denominator for the project's core metric, intent events per 1,000 visits. The thanks page deliberately carries no beacon so submissions don't inflate the denominator. Because the beacon counts humans only, nginx access logs remain the bot-inclusive ground truth for traffic-claim comparisons.
- **sources:** traffic_probe_runbook(13).md#2, relojistas_notes(8).md#what-we-record, traffic_probe_running_notes(28).md#2026-06-11
- **relations:** intent_site_stats snapshot, traffic-claim verification, access-log passive harvest
- **verify-later:** /api/hit handler in service.go; counters.json per-host visits

<!-- SOURCE: U11_traffic_probe.md -->
### Capture-side input sanitisation with deferred normalisation
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-12 "Sanitisation v2 … Tests green"; a real bug (tab both Cc and whitespace silently joining words, `gmt\t\tmaster` → `gmtmaster`) found and fixed.
- **what:** The engine's sanitizeValue strips Unicode Cc AND Cf (zero-widths, bidi overrides incl. U+202E, BOM, soft hyphen), collapses whitespace runs (IsSpace checked FIRST), caps values by runes not bytes (multibyte-safe), drops junk-only submissions. Deliberate division of labour: NFC normalisation + lowercasing happen at the P4 collector, not the engine — the engine is stdlib-only (no x/text), so NFD combining marks pass through and two byte-forms of "ñ" count as separate terms until ingest normalises.
- **sources:** traffic_probe_running_notes(28).md#2026-06-12 (sanitisation v2), traffic_probe_plan(12).md#P4 ingest contract, relojistas_notes(8).md#decisions
- **relations:** ingest validation contract, ranking queries (lower() caveat)
- **verify-later:** sanitizeValue in site-engine service.go; NFC step in collector action

<!-- SOURCE: U11_traffic_probe.md -->
### /events export endpoint and checkpoint contract
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-12: "GET /events built + tested … Tests green ×6"; HANDOFF lists "/events export endpoint" among live capabilities.
- **what:** Key-gated NDJSON stream of stored events, oldest first, original line bytes preserved; params since (RFC3339, strictly-after), host, limit (default 5000); final `_meta` line {count, truncated, server_time}. Checkpoint contract: collector stores max created_at received; strictly-after semantics + the engine event id make pulls duplicate-free. Lock-free by design so a large export can never block live captures — a torn mid-append tail line is skipped and arrives next pull. Day-file skip by filename date.
- **sources:** traffic_probe_runbook(13).md#6, traffic_probe_running_notes(28).md#2026-06-12 (events built), relojistas_notes(8).md#how-we-see
- **relations:** intent_events table (consumer), pull-not-push collection topology
- **verify-later:** Store.StreamEvents + App.events in site-engine; nginx /events location on box

<!-- SOURCE: U11_traffic_probe.md -->
### Access-log passive harvest and /access-digest
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** passive_harvest_spec(2) 2026-06-13: "Part 2 — access-log digest: DONE" (endpoint built + tested) but "STILL TO DO on the collector side: pull /access-digest per site into a rollup table".
- **what:** The signals the structured event stream can never see on a static page load — external referer, landing path+query, the dead-forum 404 paths (themselves an intent signal: what surviving inbound links point at), and user-agent for bot classification — already sit in nginx's combined access log. Option A (chosen over B: defer to P5 ssh; C: Cloudflare analytics): the engine reads its own box's per-domain log and exposes key-gated `GET /access-digest?host=&since=&top=` returning status mix, top referers (canonicalHost-reduced, self excluded), top paths, top 404 paths, UA buckets (known_search_bot / seo_or_scraper_bot / other_bot / browser_like / empty / other), top real client IPs. Requires setup.sh support: per-domain access_log files, engine user in adm group, CF real_ip conf when proxied.
- **sources:** passive_harvest_spec(2).md, traffic_probe_running_notes(28).md#2026-06-13-g, deploy_setup/working_dir/accessdigest.go (header)
- **relations:** global bot-IP blocklist (same rollup source), traffic-claim verification, Cloudflare-proxied option
- **verify-later:** accessdigest.go in site-engine repo; whether the collector rollup table was ever built

<!-- SOURCE: U11_traffic_probe.md -->
### intent_events table with structural idempotency
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-13(d): "Migration applied (operator: CREATE TABLE + 3 indexes + INSERT 0 1 task)".
- **what:** Cluster-side landing table for pulled events: engine_event_id UNIQUE makes re-pulling overlapping windows a no-op via ON CONFLICT DO NOTHING, so the collector can use a safely-overlapping since. Checkpoint needs no extra storage — next since = max(event_created_at) per host. CHECK constraints on kind enum and value length; host resolved to site_id (nullable FK to sites). Collected_at vs event_created_at kept separate.
- **sources:** intent_events_migration(1).sql, traffic_probe_running_notes(28).md#2026-06-13-b/d
- **relations:** /events checkpoint contract, intent collection topology, ranking queries
- **verify-later:** \d intent_events in clients_db; uq_intent_events_engine_id

<!-- SOURCE: U11_traffic_probe.md -->
### Intent collection topology — collector action under a wrapper-orchestrator
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** intent_collector_registration.sql enable-order: migration applied (done), action/agents/enable steps still pending; scheduled_tasks row "INSERTED DISABLED".
- **what:** Collection needs NO adapter and NO SSH: one Go action (`collect_intent_events`, Category "data", IsLocal) self-queries all sites with deploy_config.target='vm', pulls /events + /stats per box over key-gated HTTPS, and upserts; per-site failures caught and skipped. Because it is scheduler-reached AND does substantive unbounded work, guideline 001's wrapper rule applies: a thin `intent-collection-orchestrator` (spawn→call→complete, med-export pair mirrored verbatim incl. image v1.0.1063) spawns the `intent-collector` task worker in its own pod. The box's INTERNAL_API_KEY lives in sites.deploy_config.engine.stats_key (low-sensitivity read-only export key; movable to a secrets table later). agent_definitions is UNIQUE(type,version), so idempotency uses ON CONFLICT (type, version).
- **sources:** intent_collector_registration.sql, intent_collector_agents(2).sql, intent_events_migration(1).sql#scheduled-collector, traffic_probe_running_notes(28).md#2026-06-13-c/d
- **relations:** scheduler single-fire semantics (design correction), pull-not-push topology, scheduler-and-tasks, development-guide wrapper rule
- **verify-later:** collect_intent_events in GlobalActionRegistry; agent_definitions rows intent-collection-orchestrator/intent-collector; scheduled_tasks 'intent-collection' enabled flag and target_agent_type

<!-- SOURCE: U11_traffic_probe.md -->
### Ingest validation contract
- **category:** traffic-analytics
- **status-signal:** aspirational
- **status-evidence:** plan P4 section (2026-06-12/13 additions) specifies the contract; collector enablement itself still pending.
- **what:** Everything the collector must enforce when pulling engine lines into the DB: parameterised SQL only (values are data, never concatenated — injection structurally impossible per house rule); per-line shape checks (JSON parses, kind ∈ enum, value ≤500 runes, host ∈ accepted set, timestamp sane); burst dedupe of identical (host,value) within a minute as bot noise (raw JSONL stays source of truth); Unicode NFC normalisation + lowercasing HERE (deferred from the stdlib-only engine); DB CHECK constraints; values escaped at every display surface. Open choice: redact email/phone patterns at ingest vs rely on the 90-day prune.
- **sources:** traffic_probe_plan(12).md#P4, relojistas_notes(8).md#decisions (input hygiene), passive_harvest_spec(2).md
- **relations:** capture-side sanitisation (the other half), intent_events table, minimal-data privacy posture
- **verify-later:** validation body of collect_intent_events action

<!-- SOURCE: U11_traffic_probe.md -->
### intent_site_stats visit-count snapshot
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** passive_harvest_spec(2): "Part 1 — visit counts: DONE" (built/validated); enablement rides on the disabled collector.
- **what:** The events-per-1k denominator (visits) lives only in the engine's counters.json exposed at /stats — not in intent_events. A one-row-per-host table holds the latest cumulative /stats snapshot (visits, events, observed_at); the collector's collectSiteStats pulls it non-fatally each run; ranking query 1 LEFT JOINs it for the true rate. History table explicitly deferred until a visits-over-time trend is wanted.
- **sources:** intent_site_stats_migration.sql, passive_harvest_spec(2).md#part-1, intent_ranking_queries(1).sql#1
- **relations:** visit beacon, ranking queries, intent collection topology
- **verify-later:** \d intent_site_stats; collectSiteStats in collector action

<!-- SOURCE: U11_traffic_probe.md -->
### Ranking queries and graduation criteria
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** running notes 2026-06-13(e): "ranking ✓ … Works TODAY on absolute signal"; graduation numbers are an explicit "proposal — tune once data exists" (relojistas_notes).
- **what:** Six read-only queries over intent_events answer "is there demand here?": per-domain summary (with events-per-1k via intent_site_stats), top terms, dominant-cluster share (crude single-term proxy; real clustering a later refinement), referer breakdown, landing-query breakdown, recent raw submissions. Proposed graduation criterion (probe → real build): sustained events-per-1k ≥ 20 AND a dominant intent cluster covering ≥ 30% of terms over 2–4 weeks.
- **sources:** intent_ranking_queries(1).sql, relojistas_notes(8).md#open-choices, passive_harvest_spec(2).md#whats-not-blocked
- **relations:** intent_site_stats, traffic-probe mission (the ranking is the mission's output)
- **verify-later:** whether any report/dashboard consumes these queries

<!-- SOURCE: U11_traffic_probe.md -->
### intent-probe content component
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-11: "intent-probe INSERTED into the live library (INSERT 0 1 …); the second run's INSERT 0 0 is the ON CONFLICT idempotency working."
- **what:** New `content_components` section (STEP ZERO verdict: nothing in the 83-section library captures anonymous intent server-side; contact-form collects PII — the opposite posture). Kebab function `intent-probe`, v2 input schema (tagline/action_label/placeholder/submit_label llm-sourced; probe_kind and privacy_text from config with fallbacks; contact_email from site_specs.identity, skip-if-missing), plain HTML form POST to /intent + beacon img (js_content NULL — JS Content Separation trivially satisfied), CSS-var theming scoped to .intent-probe-section. Deliberate v1 limit: single text-input action only; the {{range}}-based category-buttons variant is deferred until the renderer's array handling is verified ("arrays are where templates fail").
- **sources:** intent_probe_component(1).sql, traffic_probe_running_notes(28).md#2026-06-10 (STEP ZERO) and #2026-06-11
- **relations:** requires-backend capability gate (carries the tag), probe page pattern, contracts-and-standards, tool-library
- **verify-later:** SELECT … FROM content_components WHERE name='intent-probe'; renderer array handling for the categories variant

<!-- SOURCE: U11_traffic_probe.md -->
### Probe content restraint — no results, no imagery, no anchoring
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas_notes decisions dated 2026-06-11 ("No results page in v1", "Imagery: v1 ships text-only").
- **what:** Three linked restraint decisions that protect the signal: (1) no results page — the probe performs no search and returns nothing; revisit only if repeated same-term re-submissions show visitors expect an answer; (2) v1 text-only — no manufacturer/press photos (rights, shop-implication, and any displayed list ANCHORS what visitors search for); v1.1 option is ONE brand-free generated hero via the chassis image pipeline; (3) the "novedades" category-buttons idea would turn the latest-models display into measurement itself (kind=categories) but must run as an A/B against the plain box, with top-terms read before choosing the button set. Status of (2)-hero and (3): aspirational.
- **sources:** relojistas_notes(8).md#decisions (imagery, no-results), traffic_probe_running_notes(28).md#2026-06-11 (imagery)
- **relations:** intent-probe component (deferred categories variant), imagery (platform pipeline)
- **verify-later:** whether any probe page ever gained a hero image or category buttons

<!-- SOURCE: U11_traffic_probe.md -->
### Traffic-claim verification and the bot-vs-human verdict method
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas_notes Log 2026-06-12: "VERDICT (access log, 14,961 reqs): overwhelmingly bots/ghosts, human intent ≈ 0 … a clean probe result, not a measurement failure."
- **what:** Marketplace visit estimates are treated as unverified relative rankings (relojistas' claimed ~1.2M/mo was the outlier test case). Method: convert the claim to expected visits/hour; compare beacon (humans-only) vs nginx access log (bot-inclusive ground truth); enumerate confounds before concluding (DNS propagation window, humans-only beacon, the invisible www gap); set a dated verdict criterion (48h, UA-split requests/day, www share). Relojistas outcome: 83% 404s on dead vBulletin paths; UA mix Chrome-spoof crawler / Claude-SearchBot / SemrushBot / YandexBot; Cloudflare's "unique visitors" an upper bound dominated by bots. A negative verdict is a successful probe result. By-product: the 404 paths ARE intent and feed the passive harvest.
- **sources:** relojistas_notes(8).md#log (verdict + traffic-claim assessment), traffic_probe_running_notes(28).md#2026-06-13, README_stats_internal_key.md (the settle-it commands)
- **relations:** visit beacon, access-log passive harvest, WWW_ALIAS (closes the www confound), debugging (don't-jump-to-conclusions rule applied)
- **verify-later:** relojistas access-log digests over a longer window; whether any other domain got the same treatment

<!-- SOURCE: U11_traffic_probe.md -->
### Global bot-IP blocklist (Thread D)
- **category:** traffic-analytics
- **status-signal:** aspirational
- **status-evidence:** HANDOFF Thread D: "Design sketch for this thread … This is separate from intent capture but shares the log source" — no build claimed anywhere.
- **what:** Operator idea: relojistas' bot storm makes it a harvesting ground for illegitimate-crawler IPs (high-volume, spoofed-UA, 404-storming, robots.txt-ignoring) to block GLOBALLY across all boxes/sites via a shared denylist applied at the edge (nginx geo/map deny, or Cloudflare where proxied), with legitimate crawlers (Googlebot, Bing, real Claude-SearchBot) allow-listed. Consumes the same UA/IP rollup the access-digest produces.
- **sources:** HANDOFF_vm_sites_permanent_thread.md#thread-d, passive_harvest_spec(2).md#if-option-a
- **relations:** access-log passive harvest (shared source), Cloudflare-proxied option
- **verify-later:** any denylist artifact on the boxes or in vm-sites/site-engine repos

<!-- SOURCE: U11_traffic_probe.md -->
### Relojistas static-rebuild manifest (Thread A)
- **category:** traffic-analytics
- **status-signal:** aspirational
- **status-evidence:** HANDOFF Thread A: "do first; concrete … Open: build now from heritage alone, or wait ~1–2 weeks for P4 intent data? (Lean: scaffold now, enrich from data.)"
- **what:** Despite the bot verdict, relojistas keeps value: an RSS feed real aggregators still pull (populate with OUR content), heavy crawler presence already indexing the domain, and the 404/referer log revealing what inbound links want. Plan: package provenance (Spanish watch forum, boards), language, vertical, an RSS/news section (news-feed pipeline), top inbound 404 paths + referer clusters, and roadmap-pinned section_types into a manifest handed to the framework for a multi-page static build deployed via the same vm-sites Action — optionally retaining intent-probe (capability=backend) or going pure-static.
- **sources:** HANDOFF_vm_sites_permanent_thread.md#thread-a, traffic_probe_running_notes(28).md#2026-06-13-b
- **relations:** news-feed-pipeline, site-plan-and-reconciler (roadmap section_types pinning), VM-hosted backend sites class
- **verify-later:** any relojistas manifest/site_specs/roadmap rows; whether the static build happened

<!-- SOURCE: U11_traffic_probe.md -->
### Domain shortlist and selection policy
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** TASK brief §5 table + §4 "Start with 3–5 high-traffic, clearly generic domains you fully control"; two domains actioned by 2026-06-13.
- **what:** A parked-marketplace export (traffic_probe_domains.tsv, 388 lines) ranked by the marketplace's own estimated visits, with name-based vertical guesses and per-domain probe ideas. Policy: eligibility statuses concern the parking program's monetisation, NOT DNS control; repointing DNS stops parking revenue — choose deliberately; start with a few controlled generic domains; health-adjacent names (healthscare.*, overpronation.com…) need careful non-clinical framing; verify estimates against own logs before committing effort.
- **sources:** TASK_traffic_probe_brief.md#5-7, traffic_probe_domains.tsv (header), traffic_probe_plan(12).md#risks
- **relations:** traffic-claim verification, Wayback grounding
- **verify-later:** which domains beyond relojistas/wayfaringlondoner were ever probed

<!-- SOURCE: U11_traffic_probe.md -->
### Per-domain notes and living-docs convention
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** HANDOFF Thread C: "Each domain gets its own <domain>_notes.md … per the relojistas/wayfaringlondoner template"; cross-thread rule "append, don't fork".
- **what:** Every probe domain gets a living `<domain>_notes.md` holding provenance (what it was, evidence snapshot), dated decisions, open choices, coordinates (box/IP/repos/paths/key location), and a dated log. Project-level knowledge lives in three living docs (plan = decisions + phases; runbook = operational how-to; running notes = per-session reasoning journal with a rename map and "new names per the standing rule" discipline). These are the single source of truth across parallel chats.
- **sources:** relojistas_notes(8).md (the template instance), wayfaringlondoner_notes.md, HANDOFF#cross-thread, traffic_probe_running_notes(28).md#conventions
- **relations:** documentation-system (travelling/living doc conventions)
- **verify-later:** n/a (documentary convention)

<!-- SOURCE: U18_sql_for_agents.md -->
### Intent-event collection from VM-hosted backend sites (P4)
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** 119 "Pattern... mirrors... the thunder-training-monitor convention of INSERTING DISABLED until the action is deployed"; tables created; agents mirror the med-export pair.
- **what:** Off-box collection of visitor intent: VM-hosted sites expose key-gated GET /events (NDJSON) and /stats; a scheduled intent-collection-orchestrator/intent-collector pair pulls events into intent_events (engine_event_id UNIQUE gives structural idempotency — safe overlapping `since` windows, checkpoint derived from max(event_created_at)) and cumulative visit counters into intent_site_stats (one row per host) so ranking can compute true events-per-1k-visits. kind constrained to search/categories/freetext.
- **sources:** 119_intent_events_for_vms.sql; 120_intent_site_stats.sql; 121_intent_collector_agents.sql
- **relations:** intent capture engine on the VM side (vonc/backend sites); scheduler pre_query dispatch
- **verify-later:** collector action deployment; scheduled task enabled flag; ranking queries

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Traffic-probe program (residual-traffic intent capture)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** "FIRST LIVE CAPTURE" 2026-06-12 13:03:44 UTC (kind=search, "correa Omega Seamaster"); relojistas.com live behind Cloudflare 2026-06-13.
- **what:** Put dormant-but-still-visited domains on the chassis as first-class sites whose page plausibly reflects the old vertical and offers ONE invited action ("what are you looking for?"). Captured intent ranks which domains are worth building out. End-to-end: VM (nginx + site-engine) serves + captures, cluster pulls data on schedule, framework treats each as a normal `sites` row.
- **sources:** traffic_probe_plan(11).md#how-it-all-fits, traffic_probe_running_notes(27).md#2026-06-12-first-live-capture, traffic_probe_runbook(12).md#0
- **relations:** parent of site-engine, intent-probe component, P4 collection, VM-hosted backend sites class
- **verify-later:** `sites` rows with deploy_config.target='vm'; intent_events table

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Visit beacon + events-per-1k ranking metric
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** service(24).go / main.go headers describe `GET /api/hit` 1×1 beacon; runbook(12) §6 "Metric: intent events per 1,000 visits"; ranking query 1 LEFT JOINs for events-per-1k.
- **what:** A no-JS/no-cookie 1×1 image (`<img src="/api/hit">`) on the page counts human visits as the denominator for an "intent events per 1,000 visits" ranking metric. Visits live in the engine's counters.json (/stats), not in intent_events, so the rate metric requires joining the intent_site_stats snapshot. The gracias/thanks page deliberately omits the beacon (would inflate the denominator).
- **sources:** deploy_setup/working_dir/service(24).go#header, traffic_probe_runbook(12).md#6, traffic_probe_running_notes(27).md#2026-06-13-e
- **relations:** feeds intent_ranking_queries; depends on intent_site_stats
- **verify-later:** counters.json; /stats visit counter; intent_ranking_queries.sql query 1

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### /access-digest endpoint (passive nginx-log harvest)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** accessdigest(1).go header "parse this box's nginx combined access log into a compact, key-gated digest"; running_notes 2026-06-13(g) "/access-digest endpoint BUILT + tested … Builds clean".
- **what:** `GET /access-digest?host=&since=&top=` returns a key-gated JSON rollup of one domain's nginx combined access log: status mix, top referers (canonicalHost-reduced), top paths, top 404 paths, UA buckets, top REAL client IPs. Captures the referer/landing-path/404-intent/UA signals the engine can't see on a static page load. Needs per-domain logs + engine in the `adm` group (both from setup.sh); needs `CLOUDFLARE=true` (nginx real_ip) on proxied boxes so IPs are the real client, not Cloudflare's.
- **sources:** deploy_setup/working_dir/accessdigest(1).go#header, traffic_probe_running_notes(27).md#2026-06-13-g, traffic_probe_runbook(12).md#6
- **relations:** implements passive_harvest_spec Option A part 2; shares source with Thread-D bot blocklist
- **verify-later:** accessdigest.go buildAccessDigest/classifyUA/safeHost; NGINX_LOG_DIR config (main(19).go)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### intent_site_stats + intent_ranking_queries
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(f) "Option A part 1 (visits) BUILT: intent_site_stats table … ranking query 1 LEFT JOINs for events-per-1k"; 2026-06-13(e) "intent_ranking_queries.sql — 6 read-only queries".
- **what:** `intent_site_stats` stores the latest /stats snapshot per host (PK host); the collector's collectSiteStats pulls /stats and upserts (non-fatal). `intent_ranking_queries.sql` is 6 read-only queries over intent_events: per-domain summary, top terms, dominant-cluster share (the graduation signal), referer breakdown, landing-query breakdown, recent raw submissions — working today on absolute signal, with events-per-1k once visits join.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-e, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** consumes /stats; ranking is the domain-graduation decision input
- **verify-later:** intent_site_stats_migration.sql; intent_ranking_queries.sql

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### passive_harvest_spec (3 options, A recommended)
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(e) "passive_harvest_spec.md lays out 3 options … RECOMMENDS A … DECISION NEEDED from operator before building"; parts built in (f)/(g).
- **what:** Spec for getting the visit rate + passive signals (referer/404/UA, which live in nginx's combined log, not visible to the engine on static loads). Option A: engine reads its own box's nginx log + /stats → key-gated digest, preserving the pull model (new intent_site_stats table + /access-digest). Option B: defer to the P5 vmhost SSH adapter. Option C: Cloudflare analytics if proxied. A was chosen and built.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-e, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** realised by /access-digest + intent_site_stats
- **verify-later:** passive_harvest_spec.md options A/B/C

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### landing_query enrichment on IntentEvent
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-13 "Small legitimate engine enrichment shipped: landing_query field on IntentEvent … Tested … Additive, no breaking change".
- **what:** IntentEvent gained a `landing_query` field populated from the submission's Referer query (the inbound ?q=/?utm= that survives into the form page), so the structured /events export carries inbound-query intent without a log-join. omitempty when absent; external ref_host still recorded separately.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict
- **relations:** complements the access-log harvest (referer host)
- **verify-later:** service.go IntentEvent.LandingQuery / landingQuery() helper

<!-- SOURCE: U05_content_quality_linking.md -->
### Traffic-probe capture backend (misfiled in this unit)
- **category:** traffic-analytics
- **status-signal:** unknown
- **status-evidence:** Code present with reconciled architecture comments ("Division of labour after the chassis reconciliation"); no deployment claim in this unit.
- **what:** A stdlib-only Go service capturing visitor intent on probe domains: POST /intent (search/categories/freetext events), GET /api/hit (1x1 no-cookie visit beacon), key-gated /stats, host-keyed JSON store forked from idea.uk; no IPs stored, referer reduced to host (UK GDPR/PECR posture). The chassis builds/serves the static probe pages; nginx proxies only the capture paths. Sits in this unit's golang_code/ by accident of filing — belongs to the traffic-probe concept area.
- **sources:** golang_code/service.go, store.go, main.go (headers)
- **relations:** docs024 traffic_probe unit; idea.uk store pattern.
- **verify-later:** deployment on probe VMs; overlap with docs024/traffic_probe docs (canonical home).

<!-- SOURCE: U11_traffic_probe.md -->
### Traffic-probe mission — intent discovery on parked domains
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** HANDOFF (≈2026-06-13): "Engine (site-engine, stdlib Go) live on a dedicated EU box for relojistas.com … HTTPS, capturing"; plan: P0–P2 done, P3/P4 in progress, P5 not started.
- **what:** Domains that still receive residual visitors but serve only a parking lander are put on a minimal "probe" page that plausibly reflects the old vertical and invites ONE action (search box / category links / free-text). The stated intent is captured server-side; after 2–4 weeks the terms rank which domains have real demand worth building an idea.uk-style site for. Explicit scope boundary: capture what visitors *say they want* on our own page, never recover anyone's old gated content.
- **sources:** TASK_traffic_probe_brief.md#1-2, traffic_probe_plan(12).md#how-it-all-fits, traffic_probe_runbook(13).md#0
- **relations:** probe page pattern, ranking queries + graduation criteria, VM-hosted backend sites class
- **verify-later:** live relojistas.com/stats; intent_events table row counts; sites rows with deploy_config.target='vm'

<!-- SOURCE: U11_traffic_probe.md -->
### Wayback grounding of probe pages
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-11: "Relojistas grounded in the snapshot: it was a Spanish watch FORUM"; 2026-06-13(b): grounding constraint recorded.
- **what:** Before building a probe page, look up what the domain used to be via archive.org (CDX path list, availability API, snapshot view). The old path list (/login, /members, /forum…) signals what was gated and what visitors still want; the snapshot fixes language, vertical, and the invited action. Operational constraint discovered: Claude can web_fetch archive pages only when a search surfaces the exact URL and cannot enumerate CDX on demand — so the operator supplies Wayback URLs/snapshots, or grounding falls back to web search + the domain name.
- **sources:** TASK_traffic_probe_brief.md#2-method, traffic_probe_running_notes(28).md#2026-06-13-b, HANDOFF#thread-c
- **relations:** per-domain notes convention; adoption-pipeline (site recreation from crawl) is the platform cousin
- **verify-later:** archive.org.results/ snapshots exist for both live domains (they do, in this unit)

<!-- SOURCE: U11_traffic_probe.md -->
### Probe page pattern — one invited action, plausible framing
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas live 2026-06-12 (first capture 13:03:44 UTC); wayfaringlondoner page "built + grounded … not yet deployed" (HANDOFF).
- **what:** The page looks intentional, not parked: a one-line tagline matching the old vertical, exactly one invited action (v1: a single text input, kind=search or freetext), a plain privacy line, a 1×1 beacon, no JS, no cookies. Framing follows the domain's heritage — relojistas is a Spanish marketplace/search posture (marca/modelo/reparación/compraventa, thanks at /gracias.html); wayfaringlondoner is an English BLOG posture asking for a destination/story. Hand-made pages for the first domains were explicitly a go-live unblocker; chassis-built pages take over under P3.
- **sources:** TASK_traffic_probe_brief.md#2, relojistas_notes(8).md#decisions, wayfaringlondoner_notes.md#decisions, relojistas_golive/index.html
- **relations:** intent-probe component (the library form of the same pattern), probe content restraint
- **verify-later:** live page HTML vs intent-probe component render

<!-- SOURCE: U11_traffic_probe.md -->
### Minimal-data privacy posture (UK GDPR/PECR)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes "Standing observations": "no cookies, no JS, no IP stored, referer reduced to host, country only from a coarse CDN header"; holds "regardless of volume" (relojistas_notes).
- **what:** Server-side-only logging, no third-party trackers, no non-essential cookies (nothing stored on the device → no consent banner needed), no names/emails collected, free-text treated as potentially personal and not retained longer than needed, plain privacy line on every page. Explicitly declared load-invariant: under traffic pressure the project will not add client-side JS, third-party analytics, or IP logging. Open choice: redact email/phone patterns at ingest vs rely on the 90-day prune.
- **sources:** TASK_traffic_probe_brief.md#4, relojistas_notes(8).md#traffic-handling, traffic_probe_running_notes(28).md#standing-observations
- **relations:** intent event record, ingest validation contract, content-governance (platform-wide posture cousin)
- **verify-later:** intent-probe component privacy_text fallback; engine code stores no IP/UA

<!-- SOURCE: U11_traffic_probe.md -->
### Intent event record (fields and deliberate omissions)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas_notes "What we record" + FIRST LIVE CAPTURE log entry 2026-06-12 13:03:44 UTC.
- **what:** One event per submission: id, host, kind (search|categories|freetext), value (typed text ≤500 runes), ref_host (referer reduced to bare host, blank if same-site), country (coarse CDN header or empty), created_at (UTC), plus landing_query (inbound ?q=/?utm= that survived to the form page — added 2026-06-13 so the structured export carries it without a log join). Deliberately NOT recorded: IP addresses, user agents, cookies, full referer URLs, names/emails. There is no results page: the probe performs no search; the submission itself is the product (303 → thanks page).
- **sources:** relojistas_notes(8).md#what-we-record, traffic_probe_running_notes(28).md#2026-06-13 (landing_query), intent_events_migration(1).sql
- **relations:** minimal-data privacy posture, /events export, intent_events table
- **verify-later:** IntentEvent struct in site-engine repo; events-*.jsonl line shape on box

<!-- SOURCE: U11_traffic_probe.md -->
### Visit beacon and events-per-1k metric
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** runbook §2 "the page must include the visit beacon"; running notes self-correction 2026-06-11 (beacon removed from gracias page).
- **what:** A no-JS, no-cookie 1×1 `<img src="/api/hit">` counts human-with-browser visits per host — the denominator for the project's core metric, intent events per 1,000 visits. The thanks page deliberately carries no beacon so submissions don't inflate the denominator. Because the beacon counts humans only, nginx access logs remain the bot-inclusive ground truth for traffic-claim comparisons.
- **sources:** traffic_probe_runbook(13).md#2, relojistas_notes(8).md#what-we-record, traffic_probe_running_notes(28).md#2026-06-11
- **relations:** intent_site_stats snapshot, traffic-claim verification, access-log passive harvest
- **verify-later:** /api/hit handler in service.go; counters.json per-host visits

<!-- SOURCE: U11_traffic_probe.md -->
### Capture-side input sanitisation with deferred normalisation
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-12 "Sanitisation v2 … Tests green"; a real bug (tab both Cc and whitespace silently joining words, `gmt\t\tmaster` → `gmtmaster`) found and fixed.
- **what:** The engine's sanitizeValue strips Unicode Cc AND Cf (zero-widths, bidi overrides incl. U+202E, BOM, soft hyphen), collapses whitespace runs (IsSpace checked FIRST), caps values by runes not bytes (multibyte-safe), drops junk-only submissions. Deliberate division of labour: NFC normalisation + lowercasing happen at the P4 collector, not the engine — the engine is stdlib-only (no x/text), so NFD combining marks pass through and two byte-forms of "ñ" count as separate terms until ingest normalises.
- **sources:** traffic_probe_running_notes(28).md#2026-06-12 (sanitisation v2), traffic_probe_plan(12).md#P4 ingest contract, relojistas_notes(8).md#decisions
- **relations:** ingest validation contract, ranking queries (lower() caveat)
- **verify-later:** sanitizeValue in site-engine service.go; NFC step in collector action

<!-- SOURCE: U11_traffic_probe.md -->
### /events export endpoint and checkpoint contract
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-12: "GET /events built + tested … Tests green ×6"; HANDOFF lists "/events export endpoint" among live capabilities.
- **what:** Key-gated NDJSON stream of stored events, oldest first, original line bytes preserved; params since (RFC3339, strictly-after), host, limit (default 5000); final `_meta` line {count, truncated, server_time}. Checkpoint contract: collector stores max created_at received; strictly-after semantics + the engine event id make pulls duplicate-free. Lock-free by design so a large export can never block live captures — a torn mid-append tail line is skipped and arrives next pull. Day-file skip by filename date.
- **sources:** traffic_probe_runbook(13).md#6, traffic_probe_running_notes(28).md#2026-06-12 (events built), relojistas_notes(8).md#how-we-see
- **relations:** intent_events table (consumer), pull-not-push collection topology
- **verify-later:** Store.StreamEvents + App.events in site-engine; nginx /events location on box

<!-- SOURCE: U11_traffic_probe.md -->
### Access-log passive harvest and /access-digest
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** passive_harvest_spec(2) 2026-06-13: "Part 2 — access-log digest: DONE" (endpoint built + tested) but "STILL TO DO on the collector side: pull /access-digest per site into a rollup table".
- **what:** The signals the structured event stream can never see on a static page load — external referer, landing path+query, the dead-forum 404 paths (themselves an intent signal: what surviving inbound links point at), and user-agent for bot classification — already sit in nginx's combined access log. Option A (chosen over B: defer to P5 ssh; C: Cloudflare analytics): the engine reads its own box's per-domain log and exposes key-gated `GET /access-digest?host=&since=&top=` returning status mix, top referers (canonicalHost-reduced, self excluded), top paths, top 404 paths, UA buckets (known_search_bot / seo_or_scraper_bot / other_bot / browser_like / empty / other), top real client IPs. Requires setup.sh support: per-domain access_log files, engine user in adm group, CF real_ip conf when proxied.
- **sources:** passive_harvest_spec(2).md, traffic_probe_running_notes(28).md#2026-06-13-g, deploy_setup/working_dir/accessdigest.go (header)
- **relations:** global bot-IP blocklist (same rollup source), traffic-claim verification, Cloudflare-proxied option
- **verify-later:** accessdigest.go in site-engine repo; whether the collector rollup table was ever built

<!-- SOURCE: U11_traffic_probe.md -->
### intent_events table with structural idempotency
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-13(d): "Migration applied (operator: CREATE TABLE + 3 indexes + INSERT 0 1 task)".
- **what:** Cluster-side landing table for pulled events: engine_event_id UNIQUE makes re-pulling overlapping windows a no-op via ON CONFLICT DO NOTHING, so the collector can use a safely-overlapping since. Checkpoint needs no extra storage — next since = max(event_created_at) per host. CHECK constraints on kind enum and value length; host resolved to site_id (nullable FK to sites). Collected_at vs event_created_at kept separate.
- **sources:** intent_events_migration(1).sql, traffic_probe_running_notes(28).md#2026-06-13-b/d
- **relations:** /events checkpoint contract, intent collection topology, ranking queries
- **verify-later:** \d intent_events in clients_db; uq_intent_events_engine_id

<!-- SOURCE: U11_traffic_probe.md -->
### Intent collection topology — collector action under a wrapper-orchestrator
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** intent_collector_registration.sql enable-order: migration applied (done), action/agents/enable steps still pending; scheduled_tasks row "INSERTED DISABLED".
- **what:** Collection needs NO adapter and NO SSH: one Go action (`collect_intent_events`, Category "data", IsLocal) self-queries all sites with deploy_config.target='vm', pulls /events + /stats per box over key-gated HTTPS, and upserts; per-site failures caught and skipped. Because it is scheduler-reached AND does substantive unbounded work, guideline 001's wrapper rule applies: a thin `intent-collection-orchestrator` (spawn→call→complete, med-export pair mirrored verbatim incl. image v1.0.1063) spawns the `intent-collector` task worker in its own pod. The box's INTERNAL_API_KEY lives in sites.deploy_config.engine.stats_key (low-sensitivity read-only export key; movable to a secrets table later). agent_definitions is UNIQUE(type,version), so idempotency uses ON CONFLICT (type, version).
- **sources:** intent_collector_registration.sql, intent_collector_agents(2).sql, intent_events_migration(1).sql#scheduled-collector, traffic_probe_running_notes(28).md#2026-06-13-c/d
- **relations:** scheduler single-fire semantics (design correction), pull-not-push topology, scheduler-and-tasks, development-guide wrapper rule
- **verify-later:** collect_intent_events in GlobalActionRegistry; agent_definitions rows intent-collection-orchestrator/intent-collector; scheduled_tasks 'intent-collection' enabled flag and target_agent_type

<!-- SOURCE: U11_traffic_probe.md -->
### Ingest validation contract
- **category:** traffic-analytics
- **status-signal:** aspirational
- **status-evidence:** plan P4 section (2026-06-12/13 additions) specifies the contract; collector enablement itself still pending.
- **what:** Everything the collector must enforce when pulling engine lines into the DB: parameterised SQL only (values are data, never concatenated — injection structurally impossible per house rule); per-line shape checks (JSON parses, kind ∈ enum, value ≤500 runes, host ∈ accepted set, timestamp sane); burst dedupe of identical (host,value) within a minute as bot noise (raw JSONL stays source of truth); Unicode NFC normalisation + lowercasing HERE (deferred from the stdlib-only engine); DB CHECK constraints; values escaped at every display surface. Open choice: redact email/phone patterns at ingest vs rely on the 90-day prune.
- **sources:** traffic_probe_plan(12).md#P4, relojistas_notes(8).md#decisions (input hygiene), passive_harvest_spec(2).md
- **relations:** capture-side sanitisation (the other half), intent_events table, minimal-data privacy posture
- **verify-later:** validation body of collect_intent_events action

<!-- SOURCE: U11_traffic_probe.md -->
### intent_site_stats visit-count snapshot
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** passive_harvest_spec(2): "Part 1 — visit counts: DONE" (built/validated); enablement rides on the disabled collector.
- **what:** The events-per-1k denominator (visits) lives only in the engine's counters.json exposed at /stats — not in intent_events. A one-row-per-host table holds the latest cumulative /stats snapshot (visits, events, observed_at); the collector's collectSiteStats pulls it non-fatally each run; ranking query 1 LEFT JOINs it for the true rate. History table explicitly deferred until a visits-over-time trend is wanted.
- **sources:** intent_site_stats_migration.sql, passive_harvest_spec(2).md#part-1, intent_ranking_queries(1).sql#1
- **relations:** visit beacon, ranking queries, intent collection topology
- **verify-later:** \d intent_site_stats; collectSiteStats in collector action

<!-- SOURCE: U11_traffic_probe.md -->
### Ranking queries and graduation criteria
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** running notes 2026-06-13(e): "ranking ✓ … Works TODAY on absolute signal"; graduation numbers are an explicit "proposal — tune once data exists" (relojistas_notes).
- **what:** Six read-only queries over intent_events answer "is there demand here?": per-domain summary (with events-per-1k via intent_site_stats), top terms, dominant-cluster share (crude single-term proxy; real clustering a later refinement), referer breakdown, landing-query breakdown, recent raw submissions. Proposed graduation criterion (probe → real build): sustained events-per-1k ≥ 20 AND a dominant intent cluster covering ≥ 30% of terms over 2–4 weeks.
- **sources:** intent_ranking_queries(1).sql, relojistas_notes(8).md#open-choices, passive_harvest_spec(2).md#whats-not-blocked
- **relations:** intent_site_stats, traffic-probe mission (the ranking is the mission's output)
- **verify-later:** whether any report/dashboard consumes these queries

<!-- SOURCE: U11_traffic_probe.md -->
### intent-probe content component
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-11: "intent-probe INSERTED into the live library (INSERT 0 1 …); the second run's INSERT 0 0 is the ON CONFLICT idempotency working."
- **what:** New `content_components` section (STEP ZERO verdict: nothing in the 83-section library captures anonymous intent server-side; contact-form collects PII — the opposite posture). Kebab function `intent-probe`, v2 input schema (tagline/action_label/placeholder/submit_label llm-sourced; probe_kind and privacy_text from config with fallbacks; contact_email from site_specs.identity, skip-if-missing), plain HTML form POST to /intent + beacon img (js_content NULL — JS Content Separation trivially satisfied), CSS-var theming scoped to .intent-probe-section. Deliberate v1 limit: single text-input action only; the {{range}}-based category-buttons variant is deferred until the renderer's array handling is verified ("arrays are where templates fail").
- **sources:** intent_probe_component(1).sql, traffic_probe_running_notes(28).md#2026-06-10 (STEP ZERO) and #2026-06-11
- **relations:** requires-backend capability gate (carries the tag), probe page pattern, contracts-and-standards, tool-library
- **verify-later:** SELECT … FROM content_components WHERE name='intent-probe'; renderer array handling for the categories variant

<!-- SOURCE: U11_traffic_probe.md -->
### Probe content restraint — no results, no imagery, no anchoring
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas_notes decisions dated 2026-06-11 ("No results page in v1", "Imagery: v1 ships text-only").
- **what:** Three linked restraint decisions that protect the signal: (1) no results page — the probe performs no search and returns nothing; revisit only if repeated same-term re-submissions show visitors expect an answer; (2) v1 text-only — no manufacturer/press photos (rights, shop-implication, and any displayed list ANCHORS what visitors search for); v1.1 option is ONE brand-free generated hero via the chassis image pipeline; (3) the "novedades" category-buttons idea would turn the latest-models display into measurement itself (kind=categories) but must run as an A/B against the plain box, with top-terms read before choosing the button set. Status of (2)-hero and (3): aspirational.
- **sources:** relojistas_notes(8).md#decisions (imagery, no-results), traffic_probe_running_notes(28).md#2026-06-11 (imagery)
- **relations:** intent-probe component (deferred categories variant), imagery (platform pipeline)
- **verify-later:** whether any probe page ever gained a hero image or category buttons

<!-- SOURCE: U11_traffic_probe.md -->
### Traffic-claim verification and the bot-vs-human verdict method
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas_notes Log 2026-06-12: "VERDICT (access log, 14,961 reqs): overwhelmingly bots/ghosts, human intent ≈ 0 … a clean probe result, not a measurement failure."
- **what:** Marketplace visit estimates are treated as unverified relative rankings (relojistas' claimed ~1.2M/mo was the outlier test case). Method: convert the claim to expected visits/hour; compare beacon (humans-only) vs nginx access log (bot-inclusive ground truth); enumerate confounds before concluding (DNS propagation window, humans-only beacon, the invisible www gap); set a dated verdict criterion (48h, UA-split requests/day, www share). Relojistas outcome: 83% 404s on dead vBulletin paths; UA mix Chrome-spoof crawler / Claude-SearchBot / SemrushBot / YandexBot; Cloudflare's "unique visitors" an upper bound dominated by bots. A negative verdict is a successful probe result. By-product: the 404 paths ARE intent and feed the passive harvest.
- **sources:** relojistas_notes(8).md#log (verdict + traffic-claim assessment), traffic_probe_running_notes(28).md#2026-06-13, README_stats_internal_key.md (the settle-it commands)
- **relations:** visit beacon, access-log passive harvest, WWW_ALIAS (closes the www confound), debugging (don't-jump-to-conclusions rule applied)
- **verify-later:** relojistas access-log digests over a longer window; whether any other domain got the same treatment

<!-- SOURCE: U11_traffic_probe.md -->
### Global bot-IP blocklist (Thread D)
- **category:** traffic-analytics
- **status-signal:** aspirational
- **status-evidence:** HANDOFF Thread D: "Design sketch for this thread … This is separate from intent capture but shares the log source" — no build claimed anywhere.
- **what:** Operator idea: relojistas' bot storm makes it a harvesting ground for illegitimate-crawler IPs (high-volume, spoofed-UA, 404-storming, robots.txt-ignoring) to block GLOBALLY across all boxes/sites via a shared denylist applied at the edge (nginx geo/map deny, or Cloudflare where proxied), with legitimate crawlers (Googlebot, Bing, real Claude-SearchBot) allow-listed. Consumes the same UA/IP rollup the access-digest produces.
- **sources:** HANDOFF_vm_sites_permanent_thread.md#thread-d, passive_harvest_spec(2).md#if-option-a
- **relations:** access-log passive harvest (shared source), Cloudflare-proxied option
- **verify-later:** any denylist artifact on the boxes or in vm-sites/site-engine repos

<!-- SOURCE: U11_traffic_probe.md -->
### Relojistas static-rebuild manifest (Thread A)
- **category:** traffic-analytics
- **status-signal:** aspirational
- **status-evidence:** HANDOFF Thread A: "do first; concrete … Open: build now from heritage alone, or wait ~1–2 weeks for P4 intent data? (Lean: scaffold now, enrich from data.)"
- **what:** Despite the bot verdict, relojistas keeps value: an RSS feed real aggregators still pull (populate with OUR content), heavy crawler presence already indexing the domain, and the 404/referer log revealing what inbound links want. Plan: package provenance (Spanish watch forum, boards), language, vertical, an RSS/news section (news-feed pipeline), top inbound 404 paths + referer clusters, and roadmap-pinned section_types into a manifest handed to the framework for a multi-page static build deployed via the same vm-sites Action — optionally retaining intent-probe (capability=backend) or going pure-static.
- **sources:** HANDOFF_vm_sites_permanent_thread.md#thread-a, traffic_probe_running_notes(28).md#2026-06-13-b
- **relations:** news-feed-pipeline, site-plan-and-reconciler (roadmap section_types pinning), VM-hosted backend sites class
- **verify-later:** any relojistas manifest/site_specs/roadmap rows; whether the static build happened

<!-- SOURCE: U11_traffic_probe.md -->
### Domain shortlist and selection policy
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** TASK brief §5 table + §4 "Start with 3–5 high-traffic, clearly generic domains you fully control"; two domains actioned by 2026-06-13.
- **what:** A parked-marketplace export (traffic_probe_domains.tsv, 388 lines) ranked by the marketplace's own estimated visits, with name-based vertical guesses and per-domain probe ideas. Policy: eligibility statuses concern the parking program's monetisation, NOT DNS control; repointing DNS stops parking revenue — choose deliberately; start with a few controlled generic domains; health-adjacent names (healthscare.*, overpronation.com…) need careful non-clinical framing; verify estimates against own logs before committing effort.
- **sources:** TASK_traffic_probe_brief.md#5-7, traffic_probe_domains.tsv (header), traffic_probe_plan(12).md#risks
- **relations:** traffic-claim verification, Wayback grounding
- **verify-later:** which domains beyond relojistas/wayfaringlondoner were ever probed

<!-- SOURCE: U11_traffic_probe.md -->
### Per-domain notes and living-docs convention
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** HANDOFF Thread C: "Each domain gets its own <domain>_notes.md … per the relojistas/wayfaringlondoner template"; cross-thread rule "append, don't fork".
- **what:** Every probe domain gets a living `<domain>_notes.md` holding provenance (what it was, evidence snapshot), dated decisions, open choices, coordinates (box/IP/repos/paths/key location), and a dated log. Project-level knowledge lives in three living docs (plan = decisions + phases; runbook = operational how-to; running notes = per-session reasoning journal with a rename map and "new names per the standing rule" discipline). These are the single source of truth across parallel chats.
- **sources:** relojistas_notes(8).md (the template instance), wayfaringlondoner_notes.md, HANDOFF#cross-thread, traffic_probe_running_notes(28).md#conventions
- **relations:** documentation-system (travelling/living doc conventions)
- **verify-later:** n/a (documentary convention)

<!-- SOURCE: U18_sql_for_agents.md -->
### Intent-event collection from VM-hosted backend sites (P4)
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** 119 "Pattern... mirrors... the thunder-training-monitor convention of INSERTING DISABLED until the action is deployed"; tables created; agents mirror the med-export pair.
- **what:** Off-box collection of visitor intent: VM-hosted sites expose key-gated GET /events (NDJSON) and /stats; a scheduled intent-collection-orchestrator/intent-collector pair pulls events into intent_events (engine_event_id UNIQUE gives structural idempotency — safe overlapping `since` windows, checkpoint derived from max(event_created_at)) and cumulative visit counters into intent_site_stats (one row per host) so ranking can compute true events-per-1k-visits. kind constrained to search/categories/freetext.
- **sources:** 119_intent_events_for_vms.sql; 120_intent_site_stats.sql; 121_intent_collector_agents.sql
- **relations:** intent capture engine on the VM side (vonc/backend sites); scheduler pre_query dispatch
- **verify-later:** collector action deployment; scheduled task enabled flag; ranking queries

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Traffic-probe program (residual-traffic intent capture)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** "FIRST LIVE CAPTURE" 2026-06-12 13:03:44 UTC (kind=search, "correa Omega Seamaster"); relojistas.com live behind Cloudflare 2026-06-13.
- **what:** Put dormant-but-still-visited domains on the chassis as first-class sites whose page plausibly reflects the old vertical and offers ONE invited action ("what are you looking for?"). Captured intent ranks which domains are worth building out. End-to-end: VM (nginx + site-engine) serves + captures, cluster pulls data on schedule, framework treats each as a normal `sites` row.
- **sources:** traffic_probe_plan(11).md#how-it-all-fits, traffic_probe_running_notes(27).md#2026-06-12-first-live-capture, traffic_probe_runbook(12).md#0
- **relations:** parent of site-engine, intent-probe component, P4 collection, VM-hosted backend sites class
- **verify-later:** `sites` rows with deploy_config.target='vm'; intent_events table

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Visit beacon + events-per-1k ranking metric
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** service(24).go / main.go headers describe `GET /api/hit` 1×1 beacon; runbook(12) §6 "Metric: intent events per 1,000 visits"; ranking query 1 LEFT JOINs for events-per-1k.
- **what:** A no-JS/no-cookie 1×1 image (`<img src="/api/hit">`) on the page counts human visits as the denominator for an "intent events per 1,000 visits" ranking metric. Visits live in the engine's counters.json (/stats), not in intent_events, so the rate metric requires joining the intent_site_stats snapshot. The gracias/thanks page deliberately omits the beacon (would inflate the denominator).
- **sources:** deploy_setup/working_dir/service(24).go#header, traffic_probe_runbook(12).md#6, traffic_probe_running_notes(27).md#2026-06-13-e
- **relations:** feeds intent_ranking_queries; depends on intent_site_stats
- **verify-later:** counters.json; /stats visit counter; intent_ranking_queries.sql query 1

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### /access-digest endpoint (passive nginx-log harvest)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** accessdigest(1).go header "parse this box's nginx combined access log into a compact, key-gated digest"; running_notes 2026-06-13(g) "/access-digest endpoint BUILT + tested … Builds clean".
- **what:** `GET /access-digest?host=&since=&top=` returns a key-gated JSON rollup of one domain's nginx combined access log: status mix, top referers (canonicalHost-reduced), top paths, top 404 paths, UA buckets, top REAL client IPs. Captures the referer/landing-path/404-intent/UA signals the engine can't see on a static page load. Needs per-domain logs + engine in the `adm` group (both from setup.sh); needs `CLOUDFLARE=true` (nginx real_ip) on proxied boxes so IPs are the real client, not Cloudflare's.
- **sources:** deploy_setup/working_dir/accessdigest(1).go#header, traffic_probe_running_notes(27).md#2026-06-13-g, traffic_probe_runbook(12).md#6
- **relations:** implements passive_harvest_spec Option A part 2; shares source with Thread-D bot blocklist
- **verify-later:** accessdigest.go buildAccessDigest/classifyUA/safeHost; NGINX_LOG_DIR config (main(19).go)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### intent_site_stats + intent_ranking_queries
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(f) "Option A part 1 (visits) BUILT: intent_site_stats table … ranking query 1 LEFT JOINs for events-per-1k"; 2026-06-13(e) "intent_ranking_queries.sql — 6 read-only queries".
- **what:** `intent_site_stats` stores the latest /stats snapshot per host (PK host); the collector's collectSiteStats pulls /stats and upserts (non-fatal). `intent_ranking_queries.sql` is 6 read-only queries over intent_events: per-domain summary, top terms, dominant-cluster share (the graduation signal), referer breakdown, landing-query breakdown, recent raw submissions — working today on absolute signal, with events-per-1k once visits join.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-e, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** consumes /stats; ranking is the domain-graduation decision input
- **verify-later:** intent_site_stats_migration.sql; intent_ranking_queries.sql

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### passive_harvest_spec (3 options, A recommended)
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(e) "passive_harvest_spec.md lays out 3 options … RECOMMENDS A … DECISION NEEDED from operator before building"; parts built in (f)/(g).
- **what:** Spec for getting the visit rate + passive signals (referer/404/UA, which live in nginx's combined log, not visible to the engine on static loads). Option A: engine reads its own box's nginx log + /stats → key-gated digest, preserving the pull model (new intent_site_stats table + /access-digest). Option B: defer to the P5 vmhost SSH adapter. Option C: Cloudflare analytics if proxied. A was chosen and built.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-e, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** realised by /access-digest + intent_site_stats
- **verify-later:** passive_harvest_spec.md options A/B/C

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### landing_query enrichment on IntentEvent
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-13 "Small legitimate engine enrichment shipped: landing_query field on IntentEvent … Tested … Additive, no breaking change".
- **what:** IntentEvent gained a `landing_query` field populated from the submission's Referer query (the inbound ?q=/?utm= that survives into the form page), so the structured /events export carries inbound-query intent without a log-join. omitempty when absent; external ref_host still recorded separately.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict
- **relations:** complements the access-log harvest (referer host)
- **verify-later:** service.go IntentEvent.LandingQuery / landingQuery() helper

<!-- SOURCE: U26_misc_dirs.md -->
### Audio-monitoring topic discovery with auto-spawned topic agents
- **category:** NEW:topic-intelligence
- **status-signal:** abandoned
- **status-evidence:** 017/018 fully design the pipeline (Bloomberg/podcast transcription via Whisper → topic extraction → novelty check → spawn agent) with a phased plan starting "Week 1: financial podcasts"; nothing downstream ever references it.
- **what:** A self-expanding intelligence network: audio streams/podcasts are transcribed, novel topic clusters detected (novel-phrase and frequency-spike detection against a 30-day corpus), and a specialised monitoring agent is automatically spawned per new topic (sources, sentiment, players, trajectory, content generation, subscriber alerts) — "Bloomberg mentions topic at 9:00 AM, your system publishes analysis by 9:30". Included a Domain Intelligence Orchestrator (DIO) deciding which intelligence strategy fits each domain.
- **sources:** docs/architecture/017-audio-monitoring-discussion; docs/architecture/018-audio-monitoring-tech.md#realistic-implementation-path
- **relations:** topic amplifier engine; agent spawning; cross-domain intelligence network
- **verify-later:** n/a

<!-- SOURCE: U26_misc_dirs.md -->
### Topic amplifier / deep digger engine
- **category:** NEW:topic-intelligence
- **status-signal:** abandoned
- **status-evidence:** 019/020 catalogue the hard problems and Python component designs (MinHash LSH dedup, spaCy extraction, verification engine, source discovery, PG+Elasticsearch+Redis storage) with a 6-week plan; no implementation trace exists.
- **what:** The engineering backbone for topic intelligence: data collection (news/social/RSS/scraping), temporal tracking with velocity/anomaly detection, structured extraction (dates, money, entities), claim verification against trusted sources, source discovery (link following, social-graph expansion, citation mining), scalable near-duplicate detection, and a hybrid division of labour — LLMs for context understanding/relevance/noise-filtering (rated "very strong"), traditional code for collection, temporal/quantitative analysis, dedup and storage. Honest bootstrap/noise/evolution problem analysis included.
- **sources:** docs/architecture/020-topic-amplifier-deep-digger.md; docs/architecture/019-information-discovery-agent-spawning#the-honest-assessment; docs/architecture/019-information-discovery-agent-spawning#llms-in-the-loop
- **relations:** audio-monitoring topic discovery; deep-research domain insight agent
- **verify-later:** n/a

<!-- SOURCE: U26_misc_dirs.md -->
### Cross-domain intelligence network and subscription tiers
- **category:** NEW:topic-intelligence
- **status-signal:** abandoned
- **status-evidence:** 016's "Hidden Superpowers" section (living knowledge graphs, insight arbitrage between domains, $10/$99/$999/$9,999 subscription tiers, "Organizational OS") is pure vision with no follow-through in later documentation.
- **what:** Developed domains share intelligence: patterns detected on one site alert sibling sites to opportunities ("vehicle-hire.com notices courier demand spike → couriervans.com gets alert"); accumulated contextual memory, relationship mapping and time-series pattern recognition become sellable subscriptions (industry intelligence, trend prediction, competitive clusters) and ultimately an org-wide agent deployment ("every employee gets a personal agent dashboard").
- **sources:** docs/architecture/016-competitive-advantge.md#the-hidden-superpowers-of-your-system; docs/architecture/016-competitive-advantge.md#the-organizational-os-concept
- **relations:** EBORG; audio-monitoring topic discovery; business-strategy subscription models
- **verify-later:** n/a

<!-- SOURCE: U26_misc_dirs.md -->
### Audio-monitoring topic discovery with auto-spawned topic agents
- **category:** NEW:topic-intelligence
- **status-signal:** abandoned
- **status-evidence:** 017/018 fully design the pipeline (Bloomberg/podcast transcription via Whisper → topic extraction → novelty check → spawn agent) with a phased plan starting "Week 1: financial podcasts"; nothing downstream ever references it.
- **what:** A self-expanding intelligence network: audio streams/podcasts are transcribed, novel topic clusters detected (novel-phrase and frequency-spike detection against a 30-day corpus), and a specialised monitoring agent is automatically spawned per new topic (sources, sentiment, players, trajectory, content generation, subscriber alerts) — "Bloomberg mentions topic at 9:00 AM, your system publishes analysis by 9:30". Included a Domain Intelligence Orchestrator (DIO) deciding which intelligence strategy fits each domain.
- **sources:** docs/architecture/017-audio-monitoring-discussion; docs/architecture/018-audio-monitoring-tech.md#realistic-implementation-path
- **relations:** topic amplifier engine; agent spawning; cross-domain intelligence network
- **verify-later:** n/a

<!-- SOURCE: U26_misc_dirs.md -->
### Topic amplifier / deep digger engine
- **category:** NEW:topic-intelligence
- **status-signal:** abandoned
- **status-evidence:** 019/020 catalogue the hard problems and Python component designs (MinHash LSH dedup, spaCy extraction, verification engine, source discovery, PG+Elasticsearch+Redis storage) with a 6-week plan; no implementation trace exists.
- **what:** The engineering backbone for topic intelligence: data collection (news/social/RSS/scraping), temporal tracking with velocity/anomaly detection, structured extraction (dates, money, entities), claim verification against trusted sources, source discovery (link following, social-graph expansion, citation mining), scalable near-duplicate detection, and a hybrid division of labour — LLMs for context understanding/relevance/noise-filtering (rated "very strong"), traditional code for collection, temporal/quantitative analysis, dedup and storage. Honest bootstrap/noise/evolution problem analysis included.
- **sources:** docs/architecture/020-topic-amplifier-deep-digger.md; docs/architecture/019-information-discovery-agent-spawning#the-honest-assessment; docs/architecture/019-information-discovery-agent-spawning#llms-in-the-loop
- **relations:** audio-monitoring topic discovery; deep-research domain insight agent
- **verify-later:** n/a

<!-- SOURCE: U26_misc_dirs.md -->
### Cross-domain intelligence network and subscription tiers
- **category:** NEW:topic-intelligence
- **status-signal:** abandoned
- **status-evidence:** 016's "Hidden Superpowers" section (living knowledge graphs, insight arbitrage between domains, $10/$99/$999/$9,999 subscription tiers, "Organizational OS") is pure vision with no follow-through in later documentation.
- **what:** Developed domains share intelligence: patterns detected on one site alert sibling sites to opportunities ("vehicle-hire.com notices courier demand spike → couriervans.com gets alert"); accumulated contextual memory, relationship mapping and time-series pattern recognition become sellable subscriptions (industry intelligence, trend prediction, competitive clusters) and ultimately an org-wide agent deployment ("every employee gets a personal agent dashboard").
- **sources:** docs/architecture/016-competitive-advantge.md#the-hidden-superpowers-of-your-system; docs/architecture/016-competitive-advantge.md#the-organizational-os-concept
- **relations:** EBORG; audio-monitoring topic discovery; business-strategy subscription models
- **verify-later:** n/a
