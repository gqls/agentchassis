# Register — content-governance

28 concepts, consolidated from 68 raw extractions (34 unique blocks, each duplicated
once in the source cluster file) across units U01, U02, U04, U07, U09, U14, U17a,
U17b, U18, U19, U20, U21, U22, U24a, U24c, U25.

### CGV-001 — Section-editor architecture (content_data as source of truth)
- **status:** deployed
- **status-evidence:** Decision recorded as implemented in docs018/010 ("content_data is always the source of truth... HTML patching considered and rejected"); SQL definition live (043_section_editor.sql); reconfirmed live in docs013/002; first production run 2026-07-09 (3 orchestrations, all COMPLETED), fix verified end-to-end on v1.0.1102 (2026-07-10).
- **what:** Granular post-deploy editing of a single page section without the full rebuild pipeline. Two edit types — content_edit (field_updates merge or full content_data replace) and component_swap (new template, same content) — both update page_components.content_data first, then re-render via buildRenderContextFromDB (reconstructs the full RenderContext purely from DB: site data, style collection, theme, nav, page meta, section content — no collected_data needed), reassemble the page, commit, and deploy. HTML patching was explicitly rejected because a raw rendered_html patch would be lost on the next re-render. The sanctioned production route (apply_section_edit, edit_type content_edit) re-renders exactly one page_component — blast radius one section, unlike rerender_page_sections (all sections) or the assemble-only rerender_single_page. Targets addressed by page_component_id or (page_name + slot_name). Quirk: field_updates must be non-empty (a no-op merge re-supplying a current value works).
- **sources:** docs018_rerendering/010_section_editor_architecture.md; 043_section_editor.sql; 002(4)#Section Editor / 003(8)#Source of truth principle; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#0,#1,#5
- **relations:** granular editing spectrum / three edit paths (CGV-002); spatial addressing labels (CGV-003); page_component_history (CGV-019); build_status CHECK constraint (CGV-016); locks/inline editing descendants
- **verify-later:** load_edit_context/apply_section_edit actions; section_editor_actions.go; orchestration_states for section-editor runs; buildRenderContextFromDB

### CGV-002 — Granular editing spectrum & three edit paths (routing model)
- **status:** deployed
- **status-evidence:** docs013 (current) documents all four paths as deployed ("012/013 shared table; all phases marked deployed"); docs018/008 (older, docs018_rerendering) is the ancestor design proposing the six-scope routing table, itself "partial" at the time.
- **what:** A routing model for content edits by scope, evolved from an original six-scope design (word/phrase, field value, section rewrite, component swap, page rewrite, multi-page — docs018/008) into the deployed four-path model: typos/fields get a direct edit + auto-lock + rerender (seconds, no LLM); section direction goes through edit content_brief → Regenerate → content_rewrite work item (writer prompts from the brief); page-level purpose edits trigger Regenerate Page (all unlocked sections); site-level Direction-tab spec edits require an explicit Propagate creating per-page items (skipping fully-locked pages). Section suppression (page-scoped remove/restore) completes the set. content_direction (JSONB) was the proposed carrier for page-rewrite instructions.
- **sources:** 013#Three Edit Paths, #Content Briefs & Regeneration; docs018_rerendering/008_granular_editing.md; docs018_rerendering/010_section_editor_architecture.md#Future-Extensions
- **relations:** section-editor architecture (CGV-001); lock semantics; page_components.content_brief; content_direction column on pages
- **verify-later:** content_brief column; regenerate endpoints; content_direction column and its prompt integration

### CGV-003 — Spatial addressing for natural-language editing
- **status:** partial
- **status-evidence:** Vision in docs009/001 ("edit the paragraph on the left of the blue call to action... data-uuid and data-path attributes"); partially realized per docs018/008: "Component labeling (new) — injects data-pc-id, data-slot, data-position into each <section>".
- **what:** Every visible element was meant to carry a unique ID and genealogy path so an editing agent could resolve fuzzy human instructions spatially ("3rd paragraph on the left", "the one under the yellow button") by highlighting candidates iteratively. Shipped only at section granularity (data-pc-id/data-slot/data-position mapping sections to page_components rows); element-level addressing and the conversational disambiguation loop remain unfulfilled.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-The-Spatial-Address-System; docs018_rerendering/008_granular_editing.md#What-Exists-Today; docs012_site_maps_and_components/006_start_concluding_links.md#2.4
- **relations:** section-editor architecture (CGV-001); page_components.data_uuid/data_path columns; granular editing spectrum (CGV-002)
- **verify-later:** data_uuid/data_path columns on page_components; labeling injection code

### CGV-004 — Page growth budget (growth_config, three-tier weekly limits)
- **status:** deployed
- **status-evidence:** "006/013 ✅; the news-index-blocked-by-content-budget bug drove the third tier"; seed migration inserts default growth_config for all existing active sites.
- **what:** CheckPageGrowthBudget, shared by apply_gap_plan and create_blog_posts: free under initial_target (12), then rolling 7-day caps per type — content 3/wk, blog 2/wk, structural (news-index, privacy...) 3/wk — absolute_max 60. Over-budget items become `blocked` (retryable); blog posts skip-and-continue. Config stored as a site_specs aspect (growth_config: initial_target, weekly_content_pages_max, weekly_blog_posts_max, absolute_max), admin-overridable via the dashboard Direction tab, and pinnable.
- **sources:** 013#Page Growth Budget; 006 growth budget rows; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#growth_config
- **relations:** content-gap-planner; blog planner; site_specs store; page_growth_budget.go
- **verify-later:** page_growth_budget.go tiers; budget enforcement in planner/discovery

### CGV-005 — Human direction channels and the pinned direction spec
- **status:** partial
- **status-evidence:** 007 channels table; 004 v4 integration deployed for auditors; dashboard direction panel "pending" in 007 Phase 1 then ✅ in 013 Phase 4.
- **what:** Three channels for human input: work-item request (until completed), direction update (permanent, site_specs `direction` aspect, pinned — only humans change it), and reference suggestion (feeds next planning). Auditors must not flag must_have/should_have features for removal; the strategist may add nice_to_have items but cannot remove must_have ones; a direction change resets the audit pass counter.
- **sources:** 007#Human Direction; 004#Human Direction Integration; 013 Direction tab
- **relations:** audit pass reset; spec propagation; site_specs `pinned` flag not honoured (CGV-028)
- **verify-later:** direction aspect reads in auditor prompts

### CGV-006 — Two sources of truth for site contact email
- **status:** partial
- **status-evidence:** "sites.email vs site_specs.identity.email can drift. loadSiteContactEmail uses COALESCE across both. Content writers may use either. Needs consolidation." (April 2026)
- **stage2-verified (2026-07-14):** unknown → partial — Dual-source pattern confirmed live via COALESCE(email,...) in site_db_actions.go:1019, maintenance_actions.go:160, render_site_components_action.go:328. A consolidation action exists — sync_site_identity_action.go (registered registry.go:740) copies site_specs.identity into sites columns — but its header states 'Sho...
- **what:** Contact email lives in two places with no single owner; drift produces placeholder/incorrect contact details on pages — a recurring audit finding and false-positive source.
- **sources:** HANDOFF-pipeline-triage-april-2026.md#patterns-1
- **relations:** identity-advisor specialist; content quality catalogue (empty footer contact)
- **verify-later:** loadSiteContactEmail; identity aspect writers

### CGV-007 — Standing-ambition default in the mission aspect
- **status:** aspirational
- **status-evidence:** "Go action carrying a default in the mission aspect… owner to finalise the principle wording (draft in notes)" — still an open backlog item in every later list.
- **what:** Diagnosis: the classifier is a current-state tool; mission_brief/roadmap_brief are the designated aspirational slots but are owner-supplied (nothing generates them), and strategy runs after the classifier — so with the slots empty, fresh builds describe what exists instead of leading the field. Proposed fix (framework, not hand-seeding): domain-submitter always writes a mission_brief = a fixed platform standing-ambition principle (lead the vertical; most useful forward-looking content; build around the site's distinctive tools; surpass don't mirror), merged with any owner mission, in a Go action — no new aspect, no reorder. Design leadership deliberately excluded (site-design-planner is deterministic and doesn't read mission prose).
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Setting direction before the build); idea.uk/HANDOFF_chassis_site.md (the two decisions); idea.uk/running_notes(63).md (w/x checkpoints)
- **relations:** build-standard migration (the classifier-side sibling); mission-file submission; Phase 0 read
- **verify-later:** domain-submitter def for any standing-ambition merge (expect absent)

### CGV-008 — Optimistic-lock co-management of shared rows across parallel chats
- **status:** deployed
- **status-evidence:** "Optimistic-lock pattern for co-managed writes: WHERE updated_at = '<last-known>' — UPDATE 0 = stop, re-read, coordinate" (RUNBOOK(49) Constants); lock held across 3 idle days on the Step-7 write (NOTES §9bd).
- **what:** Multiple concurrent chats/agents co-manage the same shared components and sites, so every write to a co-managed row is guarded by a freshness check plus an optimistic-lock UPDATE on the last-known updated_at; UPDATE 0 means the row moved underneath — stop and coordinate, never blind-write. Includes proactive notification of the other chat and re-reading fleet/workflow state after parallel deploys (image bumps invalidate cached workflow knowledge).
- **sources:** RUNBOOK(49).md Constants; NOTES(43).md §3, §9ao, §9ap, §9bd; HANDOFF(7).md §Platform operating model; RUNBOOK_pre_cleanup_backup.md Step R7
- **relations:** snapshot-before-change backup conventions (CGV-009); F8 remediation; locks category more broadly
- **verify-later:** whether any systematic lock/lease exists beyond the convention

### CGV-009 — Snapshot-before-change backup conventions
- **status:** deployed
- **status-evidence:** "Backups: snapshot_agent('<type>','<reason>') for agents; CTAS *_bak_* tables for data (two exist, see Cleanup)" (HANDOFF(7)); every mutating step in both threads shows a backup first.
- **what:** A layered backup discipline: agents → snapshot_agent before config migrations (revert_agent to restore); shared component schema/template → manual component_versions INSERT mirroring the working insert paths; data rows → CREATE TABLE … AS SELECT `*_bak_*` tables (dropped only at closeout); templates → shell-redirected full-column dumps before anchored UPDATEs. Backups are named with dates and tracked as explicit cleanup debt.
- **sources:** HANDOFF(7).md §operating model; RUNBOOK(49).md Constants + Step 12(c); RUNBOOK_scheme_to_components(18).md W1/W2a backup steps
- **relations:** optimistic-lock co-management (CGV-008); component versioning; site-snapshots-and-revert more broadly
- **verify-later:** snapshot_agent/revert_agent SQL functions; outstanding bak tables

### CGV-010 — Silent-fallback link family (phantom /contact.html, /services.html)
- **status:** unknown
- **status-evidence:** "Hero CTAs wrong site-wide (highest value — every page)… /services.html is a phantom page… Investigate hero-component CTA resolution" — still the named NEXT TASK as of HANDOFF_2026-06-09; no fix recorded in this corpus.
- **what:** Components emit links to pages that don't exist rather than resolving real targets or degrading gracefully: hero/bottom-CTA to /contact.html + /services.html on every page with text↔destination mismatch, header "Get Started", footer legal /terms.html, empty `href=""` per-card CTAs — several distinct mechanisms (schema fallback defaults like `cta_url use_fallback=/contact.html`, unresolved per-item fields, hardcoded legal links). Proposed direction: data-driven CTA resolution from realised pages plus a broad deployed-href vs realised-page-set audit distinguishing dropped source pages from component hallucination.
- **sources:** CATALOGUE(9)#family-b, HANDOFF_2026-06-09#next-task, HANDOFF_2026-05-25#parked-2
- **relations:** cta_url required/fallback fixes (the list-component subset, fixed); prepare_link_context/build_render_context; content quality polish batch
- **verify-later:** hero content_components input_schema CTA fields; prepare_link_context available-pages resolution

### CGV-011 — Content-regression guard on section save
- **status:** deployed
- **status-evidence:** thin_slice(27) worked example "NOTE the content-regression guard (L227) PROTECTS the content-rich page from being overwritten by an empty shell"; gamesdesign_index_rebuild Stage B "content regression blocked... a loud, correct block, not a silent no-op. Do not disable the guard to force it through."
- **what:** save_page_sections refuses to overwrite substantially richer live content with a much thinner regeneration (~3k vs ~13-15k chars observed) — so a stale page can be the guard PROTECTING content, presenting as success-no-change. Doctrine: a block means investigate why the writer produced thin content upstream, never disable the guard. Also the designed skip path: empty sections → skipped→complete via the workflow's check_skipped conditional.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#worked-example; docs019/RUNBOOK_gamesdesign_index_rebuild.md#5 (Stage B); docs019/RUNBOOK(31)_diagnosis_loop.md#update-5537ffdb
- **relations:** workflow result contract (the upstream thin-content cause); site quality LEG 4
- **verify-later:** save_page_sections_action.go regression check; check_skipped conditional

### CGV-012 — Standards curation & governance — concern curators
- **status:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §2 "Ownership maps to the concern, not to a node in the agent tree, and the set of owners is flat (one per top-level concern, ~8-10)"
- **what:** One curator agent per top-level concern would own that concern's atoms — reusing the auditor pattern and doubling as the routing advisor. A curator does vigilance + drafting + mechanical health but holds no authority over a rule's *meaning*. Ownership is flat and horizontal, deliberately not tied to the volatile agent tree. (Note: this is a general platform-governance proposal, not strictly content-specific; filed here as the closest assigned category.)
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#2, #3, #6
- **relations:** coordinator role (CGV-013); atomic standard; confirm-not-initiate; user-representative advocate
- **verify-later:** none

### CGV-013 — Coordinator role (arbitrates and frames)
- **status:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §7 "A thin coordinator layer above the curators … Resolved: the coordinator both arbitrates and frames"
- **what:** A thin layer above the peer curators (CGV-012) owning what belongs to no single concern: the concern taxonomy, the `applies_to` vocabulary, cross-concern conflicts, and packaging cross-concern decisions for human confirmation. Both arbitrates and frames, checked by a user-aligned advocate inside the framing process.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#7, #8
- **relations:** concern curators (CGV-012); user-representative advocate; decision authority
- **verify-later:** none

### CGV-014 — vonc.com mini-lobby content-edit re-render scope-boundary question
- **status:** deployed
- **status-evidence:** "docs 003/002 say HTML patching was REJECTED as an edit mechanism ('content_data is always the source of truth… if we only patched rendered_html, the edit would be lost on the next re-render')" (bundle_minilobby_trim2.sh / bundle_minilobby_trim3.sh header) — the general rule is established and deployed; the specific structural-trim task it applies to was left unresolved.
- **what:** The established rule that content edits must go through `content_data`, never by patching `rendered_html` directly, because the next re-render would discard a raw HTML patch. Two re-render paths exist: the full path (`needs_page` → page-content-writer, LLM) and the light path (`rerender_page_sections` behind a `page_rerender` item, no LLM, re-renders each section from stored `content_data` via `RenderComponentAction`) — and neither is `rerender_single_page`, which only assembles already-rendered components. `fix_component_template_action`'s `remove_element` fix_type explicitly does NOT touch `page_components.rendered_html` content, because "content changes go through the section-editor workflow" — leaving the correct edit path for a structural trim (like the provocation-card mini-lobby) genuinely unclear without a bundle to check.
- **sources:** contextkit/bundle_minilobby_trim2.sh#header, contextkit/bundle_minilobby_trim3.sh#header
- **relations:** section-editor architecture (CGV-001); vonc (site-case-studies); styling-render-pipeline
- **verify-later:** platform/orchestration/actions/fix_component_template_action.go and rerender_page_sections/rerender_single_page scope boundaries; whether the mini-lobby trim task itself was ever completed

### CGV-015 — Blog/content planning agents (blog-content-planner, content-gap-planner, internal-linker)
- **status:** deployed
- **status-evidence:** 069 documents the full empty_blog loop; 070/071/101 definitions; content-gap-planner in 075's timeout list.
- **what:** LLM planners that turn detected content gaps into concrete work: blog-content-planner (needs_blog_posts) plans 3-4 posts from specs and reuses write_build_items to create pages + needs_content_page items + blog-index rerender; content-gap-planner (needs_content_planning) decides per gap between add-section (content_rewrite), new page, spec update (needs_spec_update), or wont_fix — "The LLM here is the PLANNER, not the auditor"; internal-linker finds pages that should contextually link to an orphaned sub-page and creates content_rewrite items for natural placements.
- **sources:** 069_blog_posts.sql; 070_blog_content_planner.sql; 071_content_gap_planner.sql; 101_internal_linker.sql
- **relations:** empty_blog / orphan page checks; page-build-handler executes their items; spec-updater; news feed pipeline (content-gap-planner routing)
- **verify-later:** create_blog_posts action; empty_blog check

### CGV-016 — page_components.build_status CHECK constraint
- **status:** deployed
- **status-evidence:** "APPLIED 2026-07-11" header; pre-flight survey documented (deployed 597, pending 20; writers enumerated); constraint proved via pg_get_constraintdef.
- **what:** build_status was free text, which let apply_section_edit invent 'approved' and silently remove a live section from every discovery check's audit surface (all filter build_status='deployed'). CHECK now restricts to deployed/pending/approved/removed/needs_rebuild — turning invented values into loud write failures. 'removed' and 'needs_rebuild' retained without writers so future writers need no migration; residual legitimate-'approved'-stuck case covered by the page_component_status_drift check.
- **sources:** docs/agent_docs/sql_for_tables/049_page_components_build_status_check.sql
- **relations:** improvement-loop discovery checks; section-editor architecture (CGV-001); evidence-based claimed-item timeout
- **verify-later:** page_components_build_status_check constraint; check_page_component_status_drift

### CGV-017 — Schema-mode strict/flexible subsystem (abandoned)
- **status:** abandoned
- **status-evidence:** Originally proposed in docs017/001 (sites.schema_mode + page_components.schema_snapshot/content_snapshot/component_version_id) and listed as partially existing in docs018/008 ("What Exists Today"); the 2026-07-09 drop migration (009) states it was "only partially applied and then abandoned... snapshot columns were never created in production... no Go code reads schema_mode... auto_lock_on_deploy fired exactly once in the system's history" — trigger and function dropped, single strict row normalised. The abandonment supersedes the earlier partial/proposed status.
- **what:** A designed two-phase content lifecycle where initial builds run flexible (best-effort substitution, warnings on missing fields) and at approval a section's input_schema and content values are snapshotted and it moves to strict mode (edits validated against locked schema, template upgrades can't break approved pages, rollback via content_snapshot; lock_section_to_strict / unlock_section_for_redesign functions; auto-lock on first deploy per sites.strict_mode_trigger). It became an active liability when the apply_section_edit build_status fix would have made every edited section the only locked row on a site for a feature nothing consumed. Orphan functions and columns deliberately left in place; function body preserved as backup.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql; docs/agent_docs/sql_for_tables/009_drop_auto_lock_on_deploy.sql; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#item6; docs017_legacy_agent_rules_images_design_keydocs/001_flexible_schema_enforcement.md; docs018_rerendering/008_granular_editing.md
- **relations:** superseded in spirit by Pattern A locks + page_component_history (CGV-019); component_versions; content-reviewer approval flow (CGV-023)
- **verify-later:** absence of trigger; orphan columns schema_mode/strict_mode_trigger still present

### CGV-018 — content_items reusable content layer
- **status:** aspirational
- **status-evidence:** Full DDL + helper get_component_content + v_content_usage view exist and page_components.content_item_id survives into the live dump, but no later file shows content_items being written.
- **stage2-verified (2026-07-14):** unknown → aspirational — grep -rn "content_item_id|content_items" platform/orchestration/actions/*.go: only one unrelated hit (storage_actions.go:49, log field name). 0 Go writers/readers of content_items/content_item_id. DDL exists (004_content_items.sql) but nothing in the live action layer uses it.
- **what:** Separates "what to say" from "how to show it": typed reusable content rows (headline, tagline, service_description, testimonial, bio, cta, faq...) with semantic content_key, plain_text search, library sharing (site_id NULL + is_library + industry_vertical + library_tags), assets-style origin tracking, and status workflow. page_components reference a content_item with content_data acting as shallow-merge override (get_component_content). Would let one tagline appear in hero, footer and meta without duplication and let library content seed new sites.
- **sources:** docs/agent_docs/sql_for_tables/004_content_items.sql; docs/agent_docs/sql_for_tables/004b_content_items.md; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql#content_item_id
- **relations:** pages/page_components split; assets origin pattern
- **verify-later:** content_items row count; any writer using content_item_id

### CGV-019 — page_component_history full-snapshot content history
- **status:** deployed
- **status-evidence:** Created in Phase-0 Block A (019) explicitly as the replacement for the dropped content_snapshot/schema_snapshot columns.
- **what:** Before any content_data write to a page_component, the current value is copied here as a complete snapshot (not a diff) with source ('content-writer', 'section-editor', 'rollback'...) and triggering work item id — the rollback/audit substrate for section edits.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#3
- **relations:** replaces schema-mode snapshots (CGV-017); section-editor architecture (CGV-001); site snapshots (page-level vs site-level revert)
- **verify-later:** writers copying into history before UPDATE

### CGV-020 — Section governance columns: content_brief and suppressed_sections
- **status:** deployed
- **status-evidence:** "content_brief: records the instructions that generated each component's content. Enables admins to see, edit, and regenerate" (008 tail); "suppressed_sections... prevents discovery checks from recreating sections that were intentionally removed. The DELETE component endpoint writes to this column" (003).
- **what:** Two small governance mechanisms: page_components.content_brief JSONB preserves the generation instructions per section for admin-editable regeneration; pages.suppressed_sections lists intentionally removed section functions so discovery does not resurrect them (component-removal flow Phases 2/5).
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#content-brief; docs/agent_docs/sql_for_tables/003_pages.sql#suppressed_sections
- **relations:** improvement-loop discovery checks; granular editing spectrum (CGV-002); page-content-writer brief flow (CGV-021)
- **verify-later:** DELETE component endpoint; discovery filters on suppressed_sections

### CGV-021 — Page-content-writer + admin content brief regeneration flow
- **status:** partial
- **status-evidence:** 100_content_page_build_handler_flow.md documents the live flow (page-build-handler → page-content-writer → load_site_specs → prepare_link_context → load_page_components → process_sections_loop) then states "What's missing: The prompt has no awareness of page_components.content_brief" and specifies the new content_rewrite flow.
- **what:** The bridge document into the modern era: the current work-item pipeline's content generation path, and the gap it fixes — admin edits a brief in the dashboard (page_components.content_brief), clicks Regenerate creating a content_rewrite work item, and the writer's generate_content prompt gains an "## Admin Content Brief" block for briefed sections while unbriefed sections behave as before.
- **sources:** docs001_flow_general/100_content_page_build_handler_flow.md
- **relations:** three edit paths (CGV-002); section governance columns (CGV-020); development-guide work-item lifecycle
- **verify-later:** page_components.content_brief column; content_rewrite work item type; load_page_section_components step

### CGV-022 — Legal content agent + legal constraint rules
- **status:** aspirational
- **status-evidence:** docs017/019b "legal-content-agent | Template + minimal LLM | New" with legal_rules JSON (required_disclaimers by trigger, forbidden_phrases, required_pages by jurisdiction template); compliance-discovery-agent phased for maintenance.
- **what:** Jurisdiction-aware legal pages from vetted templates (privacy-gdpr-uk etc.), plus machine-readable constraints exported to the content writer: disclaimers triggered by content conditions with placement rules, forbidden phrases per industry ("guaranteed returns"), required pages routed to the legal nav group. Compliance discovery monitors regulatory changes in maintenance.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Legal-Content-Agent; docs017_legacy_agent_rules_images_design_keydocs/022_maintenance_architecture_plan_v2.md#Discovery-Agents
- **relations:** policy-as-filters (org framework ancestor); finance site stress test; brand DNA forbidden phrases
- **verify-later:** legal-content-agent definition; legal_rules key in sites.content_data

### CGV-023 — Content review flow with rejection → needs_attention
- **status:** deployed
- **status-evidence:** docs011/003 flow diagram (validate → auto-eval vs HITL → approve/reject → finalize_hitl or mark_page needs_attention); docs018/003: "Rejected pages are picked up by maintenance workflow."
- **what:** The content-reviewer agent's dual-mode gate: algorithmic validation feeds either auto-evaluation (errors escalate) or HITL review (human sees issues, can edit HTML inline); outcomes are approve (deploy), approve-with-edits (use edited HTML), or reject (page marked needs_attention, skipped by the build loop, queued for maintenance). Established review as a first-class pipeline stage rather than an afterthought.
- **sources:** docs011_api_hitl/003_hitl_new_plan.md; docs018_rerendering/003_website_builder_architecture_status_report.md#4; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Active-Agents
- **relations:** HITL protocol; flexible/strict approval snapshot (CGV-017); maintenance pickup of needs_attention
- **verify-later:** content-reviewer workflow JSON; needs_attention status handling in current loop

### CGV-024 — maintenance_queue + claim/complete/fail functions
- **status:** partial
- **status-evidence:** "MAINTENANCE QUEUE TABLE (for future use) ... For now, pages are flagged manually and the agent is triggered via generic-agent." Table + PL/pgSQL claim/complete/fail functions defined.
- **what:** A generic site-maintenance work queue (`maintenance_queue`) keyed by site_id with task_type (page_rebuild/css_update/nav_fix/link_repair/content_refresh), priority, JSONB payload, retry logic, and atomic `claim_maintenance_task`/`complete_maintenance_task`/`fail_maintenance_task` SQL functions using SKIP LOCKED. Later reused as the trigger surface for the chatbot install (`install_chat` task).
- **sources:** docs019_business/016_maintenance_queue_table.sql, docs019_business/017_maintenance_triage_agent.sql, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path
- **relations:** maintenance-triage agent, site-chat-installer, improvement-loop, generic install/uninstall trigger surface (CGV-025)
- **verify-later:** maintenance_queue table + claim/complete/fail functions

### CGV-025 — maintenance_queue as generic install/uninstall trigger surface
- **status:** aspirational
- **status-evidence:** Chatbot design reuses it: "Installation is requested by enqueuing a maintenance task — task_type='install_chat' on the existing maintenance_queue (which already has site_id, payload, status, retries)."
- **stage2-verified (2026-07-14):** partial → aspirational — grep -rn "install_chat" across platform/ returns 0 hits (only in docs025 FOCUS design doc and the register). No Go code or agent SQL def enqueues/handles an install_chat/uninstall task_type — generic-trigger-surface reuse for chat is unimplemented.
- **what:** The recognition that the existing `maintenance_queue` (built for page rebuilds) is a reusable, generic trigger surface for opt-in site add-ons — chat install/uninstall being the first — without touching the build pipeline. Establishes a pattern: new per-site capabilities enqueue a maintenance task rather than becoming a build-pipeline stage.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path, docs019_business/016_maintenance_queue_table.sql
- **relations:** maintenance_queue (CGV-024), site-chat-installer, maintenance-triage
- **verify-later:** maintenance_queue task_type values in use

### CGV-026 — Recommendation-specialist architecture (bug vs recommendation vs gap)
- **status:** abandoned
- **status-evidence:** PLAN_design-note-recommendation-specialists (April 2026): "Status: Proposed — not yet implemented"; "Deferred Until HITL queue becomes a bottleneck"; interim workaround is a manual `UPDATE site_work_items SET status='wont_fix'`.
- **what:** Proposal to make auditors emit an explicit `finding_type` (bug | recommendation | gap) so the pipeline stops auto-rewriting subjective opinions as if they were bugs (which caused false-positive rewrites like inventing a fake email). Bugs → `content_rewrite`; gaps → `needs_content_page` (already deployed); recommendations → new specialist agents (identity-advisor, tone-shift-agent, content-strategist) returning apply/dismiss/escalate, gated by a per-site `sites.approval_mode` (auto | review). Never implemented — confirmed absent in later live docs.
- **sources:** old_design_and_styling/PLAN_design-note-recommendation-specialists.md#proposed-architecture, #per-site-approval-mode
- **relations:** HITL review flow (CGV-023); approval_mode; write_audit_findings_action.go; content-quality/visual-design auditors
- **verify-later:** write_audit_findings_action.go auditFinding struct; sites.approval_mode column existence

### CGV-027 — Privacy posture (no cookies/JS/IP; UK GDPR/PECR)
- **status:** deployed
- **status-evidence:** running_notes standing observations "Privacy posture (UK GDPR/PECR, low risk appetite): no cookies, no JS, no IP stored, referer reduced to host, country only from a coarse CDN header".
- **what:** A deliberate low-risk privacy stance baked into the traffic-probe engine and page: no cookies, no JavaScript, no IP stored, referer reduced to host only, country only from a coarse CDN header (CF-IPCountry). Open ingest choice: redact email/phone patterns at ingest vs rely on the 90-day prune. (Filed here per source extraction; the same posture also recurs, described independently, under traffic-analytics — see TRF-004.)
- **sources:** traffic_probe_running_notes(27).md#standing-observations, traffic_probe_plan(11).md#p4, traffic_probe_runbook(12).md#6
- **relations:** intent-probe (no-JS form; TRF-015); sanitisation; retention timer; minimal-data privacy posture (traffic-analytics TRF-004)
- **verify-later:** engine handlers (no cookie/IP); site-engine-prune.timer RETENTION_DAYS=90

### CGV-028 — site_specs `pinned` flag not honoured by the write path
- **status:** partial
- **status-evidence:** REPLICATION §2: "WriteSiteSpec does not honour pinned … today the only thing protecting these specs is that improvement-sweep is disabled" (2026-07-10).
- **what:** Operator-corrected specs are marked pinned=true, but WriteSiteSpec supersedes unconditionally — pinned is an admin-display flag only. A pinned check in the spec write path is a named platform gap; the current protection is incidental (the improvement-sweep scheduler that fires content-gap-planner has been disabled since 2026-05-02).
- **sources:** docs/leopardessconsulting/REPLICATION_in_chassis.md#2; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-5
- **relations:** human direction channels (CGV-005); LLM fabrication classes; discovery/improvement-sweep disabled state
- **verify-later:** platform/orchestration/actions/site_spec_actions.go; scheduled_tasks row for improvement-sweep
