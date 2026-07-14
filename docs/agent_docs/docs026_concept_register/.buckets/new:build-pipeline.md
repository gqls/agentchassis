
<!-- SOURCE: U14_docs019_runbooks.md -->
### Builder route method — map what exists before building (§B0 census)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "Rule honoured: map what EXISTS against what the problem statement wants BEFORE creating anything. Sources: the 147-row agent_definitions census (2026-07-03)"; §B0 findings enumerated.
- **what:** The builder route's opening method: an inventory matrix of problem-statement capabilities (intake, research, planning, design, content, tools, feeds, infographics, build/deploy, improvement, observability) against the ~147 existing agent types. Findings: every section except infographics has agents; the real defect is ~8 overlapping top-tier "build the site" orchestrators; the per-section content family is already prototyped; genuine gaps are the infographics owner and the success-factor synthesis step. Liveness comes from pump + handler references, not the status column.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B0; docs019/RUNBOOK_builder_route(21).md#B0-findings
- **relations:** three builder generations; work-item relay spine; vertical-exemplar researcher (the gap filled)
- **verify-later:** agent_definitions census queries; duplicate-row Q1

<!-- SOURCE: U14_docs019_runbooks.md -->
### Three coexisting builder generations
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B1 "Three generations coexist: GEN-1 (template era) … GEN-2 (in-memory multipage) … GEN-3 (component/spec/DB era — the LIVE architecture): pageflow-builder v20 (ACTIVE)" (dumps read 2026-07-04).
- **what:** The archaeology of site building: GEN-1 template chains (strategist→architect→writer→html-assembler→site-deployer), GEN-2 in-memory multipage (chief-strategist→content loop→assemble→deployer-agent, no components/specs/review), GEN-3 component/spec/DB (pageflow-builder v20's full inline build; site-work-orchestrator as its queue-native sibling with dynamic per-item handlers and maintenance mode). Explains duplicate deployers (Q3: site-deployer serves GEN-1; deployer-agent GEN-2/3) and frames consolidation.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1; docs019/RUNBOOK_builder_route(21).md#open-questions (Q3)
- **relations:** builder census; work-item relay spine (the decision among them)
- **verify-later:** workflow dumps of the nine builders; pageflow-builder v20 definition

<!-- SOURCE: U14_docs019_runbooks.md -->
### The work-item relay spine (baton/hop model)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B3 "DECISION (pre-stated rule fires): the relay reaches page-build-handler natively ⇒ THE SPINE = THE WORK-ITEM RELAY"; MILESTONE 2026-07-06 "first end-to-end domain→deployed site through the relay" (dartsonline.com).
- **what:** The settled build architecture: work moves as a relay of site_work_items batons — each names a handler_agent; the 30s pump claims unclaimed batons and spawns the named agent; the agent does one job, writes findings to site_specs (the site's shared notebook — spec-not-message, the 1.27MB lesson), creates the next baton, stops. Full chain: domain-submitter/adoption → classifier → (vertical research) → strategist → briefing → build-site-planner (emits needs_page/design/imagery/rerender items) → page-build-handler per page → rerender/deploy. Observed extra hops: needs_composition→site-design-planner, needs_design→webdesign-agent, needs_imagery→image-build-handler, needs_rerender→rerender-pages; page items are item_type needs_page. pageflow-builder survives as intake's initial-build convenience.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B3; docs019/RUNBOOK_builder_route(21).md#B4 (plain-language explainer, map corrections); docs019/RUNBOOK_builder_route(21).md#milestone
- **relations:** build pump + immune system; builder generations; roadmap scope-decision gap; site quality programme (first output's gaps)
- **verify-later:** load_work_item_actions.go routing; the 37-row dartsonline item chain

<!-- SOURCE: U14_docs019_runbooks.md -->
### Build pump and the queue immune system
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B2 "the scheduler fires build-pipeline-trigger EVERY 30s … The queue's immune system is all ENABLED: claimed-item-timeout (evidence-based auto-complete …), feasibility-recheck, stale-orchestration-reaper, stale-work-item-reaper (48h), work-item-archiver, database-cleanup. FLAG: improvement-sweep is DISABLED."
- **what:** What drives the relay: scheduled build-pipeline-trigger (30s, pre_query gated, concurrency dispatch/8) → build-dispatch-loop → atomic claim → spawn dynamic handler → complete/fail → touch scheduled_tasks. The immune system self-heals the queue (claimed-item-timeout does evidence-based auto-complete, its SQL documenting the gamesdesign false-positive lesson; feasibility-recheck unblocks when handlers appear; reapers and archiver bound staleness). Standing flag: improvement-sweep is disabled platform-wide, so the improvement loop is not running; content-feed-refresh is enabled 6-hourly.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2
- **relations:** work-item relay spine; needs_diagnosis intake (rides the same machinery); site quality LEG 6
- **verify-later:** scheduled_tasks rows (build-pipeline-trigger, improvement-sweep enabled flags); claimed-item-timeout SQL

<!-- SOURCE: U14_docs019_runbooks.md -->
### Two front doors and duplicate classifiers (Q5)
- **category:** NEW:build-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) "Two front doors, two classifiers (overlap)"; queue item 2 "[MAIN] Q5 front-door consolidation — two classifiers, one responsibility" (queued, undecided).
- **what:** Intake exists twice: the queue door (domain-submitter → work-item relay with domain-research-classifier) and intake-orchestrator v3 (HITL: site-classifier → confirm type → questionnaire → briefing-agent → spawn dynamic builder). site-classifier and domain-research-classifier hold the same responsibility; the classifier prompt hardcodes recommended_builder="pageflow-builder"; intake carries orphaned rerender steps. Consolidation direction (deprecate the intake door vs align contracts) is an open decision.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1; docs019/RUNBOOK_builder_route(21).md#queue (item 2)
- **relations:** work-item relay spine; site_type taxonomy drift; adoption fidelity inversion
- **verify-later:** intake-orchestrator usage evidence (orchestration_names ILIKE intake); site-classifier workflow

<!-- SOURCE: U14_docs019_runbooks.md -->
### image_tag 'latest' stale-default trap
- **category:** NEW:build-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) "INCIDENT 2026-07-06 — first claim STALLED … THE ONE REAL DIFFERENCE: image_tag='latest' (column default) … the registry's latest is an ANCIENT chassis build … FIX APPLIED … NEW PARKED TRAP (systemic): agent_definitions.image_tag DEFAULTS to 'latest' — every future seeded agent inherits it."
- **what:** Seeded agents inherit image_tag='latest', which points at an ancient pre-architecture chassis build (boots the retired generic.process consumer regardless of env) — the newly seeded researcher stalled on it. Immediate fix: copy image columns from a live donor in every seed. Systemic options parked: repoint/retire `latest`, ALTER the column default, or a New Agent checklist line. Rollback convention is the same lever inverted: revert by repointing image_tag to the prior tag. Same staleness class as the HEAD-pinned index. Follow-up question: does deploy bulk-bump pinned tags (all five tool rows updated at once suggests yes)?
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4 (incident); docs019/RUNBOOK_builder_route(21).md#queue (item 1); docs019/RUNBOOK_gamesdesign_index_rebuild.md#8 (rollback)
- **relations:** stale-corpus class; standing evidence rules (seed hygiene)
- **verify-later:** agent_definitions image_tag column default; whether redeploy-agents bumps rows

<!-- SOURCE: U14_docs019_runbooks.md -->
### Coverage baseline — guides, tools, news, curated top-N on most sites
- **category:** NEW:build-pipeline
- **status-signal:** aspirational
- **status-evidence:** builder_route(21) queue item 7 "standing expectation going forward is most sites should carry guides + tools + news + a curated (LLM-picked, non-affiliate) top-N list … the curated-list mechanism, which IS new"; "STANDING EXPECTATION HOME: 001_development_guide … NOT the per-message prompt (decays), NOT the constitution (dev method)."
- **what:** A platform content-coverage policy: most sites should carry guides, tools, news, and a curated non-affiliate top-N list of the vertical's best products/services with outbound links; "pages need not be original to be best-in-class — genuinely useful common content counts". Enforcement points are the strategist/planner prompts (relay-wide-fixes-every-site logic); the curated-list mechanism is the one genuinely new build (reuse candidates: research-agent or the exemplar-researcher crawl pattern feeding a curation step). The mechanism for guides/tools/news EXISTS (gamesdesign, gaswholesalers prove it) — dartsonline's absence is a broken route, not a missing feature.
- **sources:** docs019/RUNBOOK_builder_route(21).md#queue (item 7)
- **relations:** F0 guides pilot (the broken route); roadmap gap (same enforcement points); site quality LEG 5
- **verify-later:** 001 guideline amendment; strategist/planner prompt coverage clauses

<!-- SOURCE: U18_sql_for_agents.md -->
### pageflow-builder (component-based site build orchestration)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** Still being patched in Phase 0 imagery work 2026-05-05 (107 backs up its row before migration); 026 documents the full live step chain.
- **what:** The central v2-era builder, renamed from multipage-website-builder v3. Spawns planner/content-writer/reviewer/deployer, then: ensure_site_record → call_site_planner → store brief+plan → sync_pages_to_db → populate_nav → asset steps → select_style_collection → set_default_components → render_site_components → get_pages_to_build (filters by build_status) → build_pages_loop (write → review → assemble → deploy per page) → apply_site_design (CSS) → trigger_site_deploy (Cloudflare). The known hazard that sync_pages_to_db can reset page statuses is documented in-file.
- **sources:** 026_pageflow_builder.sql; sql_for_agents_v2/026_pageflow_builder.sql; 107_image_build_handler.sql (backup section)
- **relations:** parallel/legacy path beside site-work-orchestrator and build-dispatch-loop; uses site-planner, page-content-writer, content-reviewer, deployer-agent
- **verify-later:** whether new sites still route through pageflow-builder or only via the work-item pipeline

<!-- SOURCE: U18_sql_for_agents.md -->
### page-content-writer (section-by-section content generation)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** Continuously patched from v2 era through 069 (reads site_specs), 107-era imagery direction; 075 gives it idle_timeout 180.
- **what:** Writes one page section-by-section: spawn_research_agent → load_page_components → build_render_context → process_sections_loop (per-section LLM call constrained to that component's `llm_field_specs`) → compile_page. The prompt is a major behavioural contract: official-contact-only rule, internal-link constraint to listed pages, content_direction/imagery_direction from site_specs, admin content briefs, "Recreate Mode" for adopted sites (adapt original page markdown), and an 18-rule anti-fabrication list (no invented people/testimonials/statistics/case studies; "ALWAYS better to be honest and general than specific and fabricated").
- **sources:** 023_page_content_writer_agent.sql; sql_for_agents_v2/023_page_content_writer_agent.sql; 069_blog_posts.sql
- **relations:** called by pageflow-builder, page-build-handler; feeds save_page_sections/page_components; anti-fabrication rules relate to content-governance
- **verify-later:** live prompt_template vs the 023 copies; llm_field_specs source in content_components

<!-- SOURCE: U18_sql_for_agents.md -->
### site-work-orchestrator (unified build/maintenance over site_work_items)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 045 definition; row backed up and patched in Phase 0 imagery (107, 2026-05-05).
- **what:** Orchestrator that builds sites from prioritized site_work_items rows, calling appropriate handler agents per item, "compatible with pageflow-builder's planner and content writer". The first expression of the unified build/maintenance queue idea, later refined into the one-item-at-a-time build-dispatch-loop.
- **sources:** 045_site_work_orchestrator.sql; 107_image_build_handler.sql
- **relations:** site_work_items table; build-dispatch-loop (leaner successor/sibling); discovery agents write into its queue
- **verify-later:** which orchestrator the live triggers use; site_work_items schema

<!-- SOURCE: U18_sql_for_agents.md -->
### Work-item build pipeline: domain-submitter → dispatch loop → handler agents
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 051/052 definitions; 075 sets idle timeouts across the whole handler fleet; 146 still adding items into the same queue in 2026-07.
- **what:** The current architecture. domain-submitter (068) creates a site record + needs_domain_research item from just a domain. build-pipeline-trigger (052) is a 30-min heartbeat: seeds the build queue, finds one site with pending items, fires build-dispatch-loop (051), which loads the highest-priority claimable item, claims it, spawns+calls the handler agent, marks complete, and if items remain spawns a FRESH dispatch loop (separate orchestration, clean logs). Item chain for a new site: needs_domain_research → needs_strategy → needs_briefing → needs_site_plan → needs_content_page (per page) → images → needs_rerender. Concurrency safety via claim_work_item; health-gating via ai_endpoint_health before claiming.
- **sources:** 051_build_dispatch_loop.sql; 052_build_pipeline_trigger.sql; 068_domain_submitter_agent.sql; 085_ai_endpoint_health_checker.sql
- **relations:** every handler agent below; scheduler-and-tasks (CronJob trigger); replaces intake-orchestrator
- **verify-later:** LoadWorkItemsAction first_item patch; claim semantics; current item_type → handler_agent routing table

<!-- SOURCE: U18_sql_for_agents.md -->
### page-build-handler (content-page handler with section planning and validation gates)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 065 documents the evolved workflow with plan_sections/validate_content and error paths; 070 notes empty_sections handler switched to it.
- **what:** Wrapper solving "specialist vs handler": page-content-writer generates but doesn't persist, so this handler loads page + specs, plan_sections resolves data sources per section (creating deferred items when sections aren't ready), calls the writer, validate_content checks placeholders/templates/cross-site contamination (blockers → mark_needs_review), then save_page_sections, update_page_status, and deploys via page-rerender. Earlier version (055) was simpler (no plan/validate steps, no deploy).
- **sources:** 065_page_build_handler_wrapper.sql; 055_page_build_handler.sql; 070_blog_content_planner.sql
- **relations:** page-content-writer, page-rerender; content_rewrite items route here; needs_new_component items from plan_sections
- **verify-later:** plan_sections + validate_page_content actions; deferred-item creation

<!-- SOURCE: U18_sql_for_agents.md -->
### page-rebuild (rebuild pages without re-planning)
- **category:** NEW:build-pipeline
- **status-signal:** unknown
- **status-evidence:** 039 full definition with detailed reuse/skip lists; no later references found in this unit.
- **what:** Rebuilds specific pages (build_status='needs_rebuild') on an existing site loading all context from DB given a domain, explicitly skipping planner, sync_pages_to_db, asset generation, component rendering, CSS and nav (all already done) while reusing the standard build-loop agents. Documents design principles: agent owns its domain; spawnable not standalone; reuse before creating; complexity in Go.
- **sources:** 039_page_rebuild_agent.sql
- **relations:** pageflow-builder (same loop, different input_mapping via rebuild_context); load_site_for_rebuild action
- **verify-later:** whether page-rebuild survived the dispatch-loop refactor

<!-- SOURCE: U14_docs019_runbooks.md -->
### Builder route method — map what exists before building (§B0 census)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "Rule honoured: map what EXISTS against what the problem statement wants BEFORE creating anything. Sources: the 147-row agent_definitions census (2026-07-03)"; §B0 findings enumerated.
- **what:** The builder route's opening method: an inventory matrix of problem-statement capabilities (intake, research, planning, design, content, tools, feeds, infographics, build/deploy, improvement, observability) against the ~147 existing agent types. Findings: every section except infographics has agents; the real defect is ~8 overlapping top-tier "build the site" orchestrators; the per-section content family is already prototyped; genuine gaps are the infographics owner and the success-factor synthesis step. Liveness comes from pump + handler references, not the status column.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B0; docs019/RUNBOOK_builder_route(21).md#B0-findings
- **relations:** three builder generations; work-item relay spine; vertical-exemplar researcher (the gap filled)
- **verify-later:** agent_definitions census queries; duplicate-row Q1

<!-- SOURCE: U14_docs019_runbooks.md -->
### Three coexisting builder generations
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B1 "Three generations coexist: GEN-1 (template era) … GEN-2 (in-memory multipage) … GEN-3 (component/spec/DB era — the LIVE architecture): pageflow-builder v20 (ACTIVE)" (dumps read 2026-07-04).
- **what:** The archaeology of site building: GEN-1 template chains (strategist→architect→writer→html-assembler→site-deployer), GEN-2 in-memory multipage (chief-strategist→content loop→assemble→deployer-agent, no components/specs/review), GEN-3 component/spec/DB (pageflow-builder v20's full inline build; site-work-orchestrator as its queue-native sibling with dynamic per-item handlers and maintenance mode). Explains duplicate deployers (Q3: site-deployer serves GEN-1; deployer-agent GEN-2/3) and frames consolidation.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1; docs019/RUNBOOK_builder_route(21).md#open-questions (Q3)
- **relations:** builder census; work-item relay spine (the decision among them)
- **verify-later:** workflow dumps of the nine builders; pageflow-builder v20 definition

<!-- SOURCE: U14_docs019_runbooks.md -->
### The work-item relay spine (baton/hop model)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B3 "DECISION (pre-stated rule fires): the relay reaches page-build-handler natively ⇒ THE SPINE = THE WORK-ITEM RELAY"; MILESTONE 2026-07-06 "first end-to-end domain→deployed site through the relay" (dartsonline.com).
- **what:** The settled build architecture: work moves as a relay of site_work_items batons — each names a handler_agent; the 30s pump claims unclaimed batons and spawns the named agent; the agent does one job, writes findings to site_specs (the site's shared notebook — spec-not-message, the 1.27MB lesson), creates the next baton, stops. Full chain: domain-submitter/adoption → classifier → (vertical research) → strategist → briefing → build-site-planner (emits needs_page/design/imagery/rerender items) → page-build-handler per page → rerender/deploy. Observed extra hops: needs_composition→site-design-planner, needs_design→webdesign-agent, needs_imagery→image-build-handler, needs_rerender→rerender-pages; page items are item_type needs_page. pageflow-builder survives as intake's initial-build convenience.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B3; docs019/RUNBOOK_builder_route(21).md#B4 (plain-language explainer, map corrections); docs019/RUNBOOK_builder_route(21).md#milestone
- **relations:** build pump + immune system; builder generations; roadmap scope-decision gap; site quality programme (first output's gaps)
- **verify-later:** load_work_item_actions.go routing; the 37-row dartsonline item chain

<!-- SOURCE: U14_docs019_runbooks.md -->
### Build pump and the queue immune system
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B2 "the scheduler fires build-pipeline-trigger EVERY 30s … The queue's immune system is all ENABLED: claimed-item-timeout (evidence-based auto-complete …), feasibility-recheck, stale-orchestration-reaper, stale-work-item-reaper (48h), work-item-archiver, database-cleanup. FLAG: improvement-sweep is DISABLED."
- **what:** What drives the relay: scheduled build-pipeline-trigger (30s, pre_query gated, concurrency dispatch/8) → build-dispatch-loop → atomic claim → spawn dynamic handler → complete/fail → touch scheduled_tasks. The immune system self-heals the queue (claimed-item-timeout does evidence-based auto-complete, its SQL documenting the gamesdesign false-positive lesson; feasibility-recheck unblocks when handlers appear; reapers and archiver bound staleness). Standing flag: improvement-sweep is disabled platform-wide, so the improvement loop is not running; content-feed-refresh is enabled 6-hourly.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2
- **relations:** work-item relay spine; needs_diagnosis intake (rides the same machinery); site quality LEG 6
- **verify-later:** scheduled_tasks rows (build-pipeline-trigger, improvement-sweep enabled flags); claimed-item-timeout SQL

<!-- SOURCE: U14_docs019_runbooks.md -->
### Two front doors and duplicate classifiers (Q5)
- **category:** NEW:build-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) "Two front doors, two classifiers (overlap)"; queue item 2 "[MAIN] Q5 front-door consolidation — two classifiers, one responsibility" (queued, undecided).
- **what:** Intake exists twice: the queue door (domain-submitter → work-item relay with domain-research-classifier) and intake-orchestrator v3 (HITL: site-classifier → confirm type → questionnaire → briefing-agent → spawn dynamic builder). site-classifier and domain-research-classifier hold the same responsibility; the classifier prompt hardcodes recommended_builder="pageflow-builder"; intake carries orphaned rerender steps. Consolidation direction (deprecate the intake door vs align contracts) is an open decision.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1; docs019/RUNBOOK_builder_route(21).md#queue (item 2)
- **relations:** work-item relay spine; site_type taxonomy drift; adoption fidelity inversion
- **verify-later:** intake-orchestrator usage evidence (orchestration_names ILIKE intake); site-classifier workflow

<!-- SOURCE: U14_docs019_runbooks.md -->
### image_tag 'latest' stale-default trap
- **category:** NEW:build-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) "INCIDENT 2026-07-06 — first claim STALLED … THE ONE REAL DIFFERENCE: image_tag='latest' (column default) … the registry's latest is an ANCIENT chassis build … FIX APPLIED … NEW PARKED TRAP (systemic): agent_definitions.image_tag DEFAULTS to 'latest' — every future seeded agent inherits it."
- **what:** Seeded agents inherit image_tag='latest', which points at an ancient pre-architecture chassis build (boots the retired generic.process consumer regardless of env) — the newly seeded researcher stalled on it. Immediate fix: copy image columns from a live donor in every seed. Systemic options parked: repoint/retire `latest`, ALTER the column default, or a New Agent checklist line. Rollback convention is the same lever inverted: revert by repointing image_tag to the prior tag. Same staleness class as the HEAD-pinned index. Follow-up question: does deploy bulk-bump pinned tags (all five tool rows updated at once suggests yes)?
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4 (incident); docs019/RUNBOOK_builder_route(21).md#queue (item 1); docs019/RUNBOOK_gamesdesign_index_rebuild.md#8 (rollback)
- **relations:** stale-corpus class; standing evidence rules (seed hygiene)
- **verify-later:** agent_definitions image_tag column default; whether redeploy-agents bumps rows

<!-- SOURCE: U14_docs019_runbooks.md -->
### Coverage baseline — guides, tools, news, curated top-N on most sites
- **category:** NEW:build-pipeline
- **status-signal:** aspirational
- **status-evidence:** builder_route(21) queue item 7 "standing expectation going forward is most sites should carry guides + tools + news + a curated (LLM-picked, non-affiliate) top-N list … the curated-list mechanism, which IS new"; "STANDING EXPECTATION HOME: 001_development_guide … NOT the per-message prompt (decays), NOT the constitution (dev method)."
- **what:** A platform content-coverage policy: most sites should carry guides, tools, news, and a curated non-affiliate top-N list of the vertical's best products/services with outbound links; "pages need not be original to be best-in-class — genuinely useful common content counts". Enforcement points are the strategist/planner prompts (relay-wide-fixes-every-site logic); the curated-list mechanism is the one genuinely new build (reuse candidates: research-agent or the exemplar-researcher crawl pattern feeding a curation step). The mechanism for guides/tools/news EXISTS (gamesdesign, gaswholesalers prove it) — dartsonline's absence is a broken route, not a missing feature.
- **sources:** docs019/RUNBOOK_builder_route(21).md#queue (item 7)
- **relations:** F0 guides pilot (the broken route); roadmap gap (same enforcement points); site quality LEG 5
- **verify-later:** 001 guideline amendment; strategist/planner prompt coverage clauses

<!-- SOURCE: U18_sql_for_agents.md -->
### pageflow-builder (component-based site build orchestration)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** Still being patched in Phase 0 imagery work 2026-05-05 (107 backs up its row before migration); 026 documents the full live step chain.
- **what:** The central v2-era builder, renamed from multipage-website-builder v3. Spawns planner/content-writer/reviewer/deployer, then: ensure_site_record → call_site_planner → store brief+plan → sync_pages_to_db → populate_nav → asset steps → select_style_collection → set_default_components → render_site_components → get_pages_to_build (filters by build_status) → build_pages_loop (write → review → assemble → deploy per page) → apply_site_design (CSS) → trigger_site_deploy (Cloudflare). The known hazard that sync_pages_to_db can reset page statuses is documented in-file.
- **sources:** 026_pageflow_builder.sql; sql_for_agents_v2/026_pageflow_builder.sql; 107_image_build_handler.sql (backup section)
- **relations:** parallel/legacy path beside site-work-orchestrator and build-dispatch-loop; uses site-planner, page-content-writer, content-reviewer, deployer-agent
- **verify-later:** whether new sites still route through pageflow-builder or only via the work-item pipeline

<!-- SOURCE: U18_sql_for_agents.md -->
### page-content-writer (section-by-section content generation)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** Continuously patched from v2 era through 069 (reads site_specs), 107-era imagery direction; 075 gives it idle_timeout 180.
- **what:** Writes one page section-by-section: spawn_research_agent → load_page_components → build_render_context → process_sections_loop (per-section LLM call constrained to that component's `llm_field_specs`) → compile_page. The prompt is a major behavioural contract: official-contact-only rule, internal-link constraint to listed pages, content_direction/imagery_direction from site_specs, admin content briefs, "Recreate Mode" for adopted sites (adapt original page markdown), and an 18-rule anti-fabrication list (no invented people/testimonials/statistics/case studies; "ALWAYS better to be honest and general than specific and fabricated").
- **sources:** 023_page_content_writer_agent.sql; sql_for_agents_v2/023_page_content_writer_agent.sql; 069_blog_posts.sql
- **relations:** called by pageflow-builder, page-build-handler; feeds save_page_sections/page_components; anti-fabrication rules relate to content-governance
- **verify-later:** live prompt_template vs the 023 copies; llm_field_specs source in content_components

<!-- SOURCE: U18_sql_for_agents.md -->
### site-work-orchestrator (unified build/maintenance over site_work_items)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 045 definition; row backed up and patched in Phase 0 imagery (107, 2026-05-05).
- **what:** Orchestrator that builds sites from prioritized site_work_items rows, calling appropriate handler agents per item, "compatible with pageflow-builder's planner and content writer". The first expression of the unified build/maintenance queue idea, later refined into the one-item-at-a-time build-dispatch-loop.
- **sources:** 045_site_work_orchestrator.sql; 107_image_build_handler.sql
- **relations:** site_work_items table; build-dispatch-loop (leaner successor/sibling); discovery agents write into its queue
- **verify-later:** which orchestrator the live triggers use; site_work_items schema

<!-- SOURCE: U18_sql_for_agents.md -->
### Work-item build pipeline: domain-submitter → dispatch loop → handler agents
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 051/052 definitions; 075 sets idle timeouts across the whole handler fleet; 146 still adding items into the same queue in 2026-07.
- **what:** The current architecture. domain-submitter (068) creates a site record + needs_domain_research item from just a domain. build-pipeline-trigger (052) is a 30-min heartbeat: seeds the build queue, finds one site with pending items, fires build-dispatch-loop (051), which loads the highest-priority claimable item, claims it, spawns+calls the handler agent, marks complete, and if items remain spawns a FRESH dispatch loop (separate orchestration, clean logs). Item chain for a new site: needs_domain_research → needs_strategy → needs_briefing → needs_site_plan → needs_content_page (per page) → images → needs_rerender. Concurrency safety via claim_work_item; health-gating via ai_endpoint_health before claiming.
- **sources:** 051_build_dispatch_loop.sql; 052_build_pipeline_trigger.sql; 068_domain_submitter_agent.sql; 085_ai_endpoint_health_checker.sql
- **relations:** every handler agent below; scheduler-and-tasks (CronJob trigger); replaces intake-orchestrator
- **verify-later:** LoadWorkItemsAction first_item patch; claim semantics; current item_type → handler_agent routing table

<!-- SOURCE: U18_sql_for_agents.md -->
### page-build-handler (content-page handler with section planning and validation gates)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 065 documents the evolved workflow with plan_sections/validate_content and error paths; 070 notes empty_sections handler switched to it.
- **what:** Wrapper solving "specialist vs handler": page-content-writer generates but doesn't persist, so this handler loads page + specs, plan_sections resolves data sources per section (creating deferred items when sections aren't ready), calls the writer, validate_content checks placeholders/templates/cross-site contamination (blockers → mark_needs_review), then save_page_sections, update_page_status, and deploys via page-rerender. Earlier version (055) was simpler (no plan/validate steps, no deploy).
- **sources:** 065_page_build_handler_wrapper.sql; 055_page_build_handler.sql; 070_blog_content_planner.sql
- **relations:** page-content-writer, page-rerender; content_rewrite items route here; needs_new_component items from plan_sections
- **verify-later:** plan_sections + validate_page_content actions; deferred-item creation

<!-- SOURCE: U18_sql_for_agents.md -->
### page-rebuild (rebuild pages without re-planning)
- **category:** NEW:build-pipeline
- **status-signal:** unknown
- **status-evidence:** 039 full definition with detailed reuse/skip lists; no later references found in this unit.
- **what:** Rebuilds specific pages (build_status='needs_rebuild') on an existing site loading all context from DB given a domain, explicitly skipping planner, sync_pages_to_db, asset generation, component rendering, CSS and nav (all already done) while reusing the standard build-loop agents. Documents design principles: agent owns its domain; spawnable not standalone; reuse before creating; complexity in Go.
- **sources:** 039_page_rebuild_agent.sql
- **relations:** pageflow-builder (same loop, different input_mapping via rebuild_context); load_site_for_rebuild action
- **verify-later:** whether page-rebuild survived the dispatch-loop refactor
