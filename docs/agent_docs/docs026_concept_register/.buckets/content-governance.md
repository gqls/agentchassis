
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
