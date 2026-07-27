# Register — build-pipeline

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

15 concepts, consolidated from 17 raw extractions across units U14, U15, U18, U21.
Absorbed categories: new:site-build-pipeline, new:site-build-orchestration-generations
(their raw material described the same builder-lineage territory as new:build-pipeline
and is folded in below rather than kept as separate near-empty register files).

### BLD-001 — Builder route method: map what exists before building (§B0 census)
- **status:** convention
- **status-evidence:** builder_route(21) "Rule honoured: map what EXISTS against what the problem statement wants BEFORE creating anything. Sources: the 147-row agent_definitions census (2026-07-03)."
- **stage2-verified (2026-07-14):** deployed → convention — This is a documented methodology/finding (census route method), not a built artifact claim; no code to check — status-evidence is a doc citation of a one-off analysis, correctly re-classed as process not code-deployed.
- **what:** The builder route's opening method: an inventory matrix of problem-statement capabilities (intake, research, planning, design, content, tools, feeds, infographics, build/deploy, improvement, observability) against the ~147 existing agent types. Findings: every section except infographics has agents; the real defect is ~8 overlapping top-tier "build the site" orchestrators; the per-section content family is already prototyped; genuine gaps are the infographics owner and a success-factor synthesis step. Liveness is judged from pump + handler references, not the status column.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B0, #B0-findings
- **relations:** three builder generations + work-item relay spine (BLD-002); vertical-exemplar researcher (the gap filled)
- **verify-later:** agent_definitions census queries; duplicate-row Q1

### BLD-002 — Three coexisting builder generations + the work-item relay spine (baton/hop model)
- **status:** deployed
- **status-evidence:** builder_route(21) §B1: "Three generations coexist: GEN-1 (template era) … GEN-2 (in-memory multipage) … GEN-3 (component/spec/DB era — the LIVE architecture): pageflow-builder v20 (ACTIVE)"; §B3 "DECISION: the relay reaches page-build-handler natively ⇒ THE SPINE = THE WORK-ITEM RELAY"; MILESTONE 2026-07-06 "first end-to-end domain→deployed site through the relay" (dartsonline.com). Independently re-derived and confirmed by a second unit the same week ("THREE generations coexist... §B3 CLOSED: spine = the work-item relay").
- **what:** The archaeology of site building, and the settled current architecture. Three generations: GEN-1 template chains (strategist→architect→writer→html-assembler→site-deployer); GEN-2 in-memory multipage (chief-strategist→content loop→assemble→deployer-agent, no components/specs/review); GEN-3 component/spec/DB era — `pageflow-builder` v20's full inline build, with `site-work-orchestrator` as its queue-native sibling. Explains duplicate deployers (site-deployer serves GEN-1; deployer-agent GEN-2/3). The settled spine: work moves as a relay of `site_work_items` batons — each names a handler_agent; the 30s pump claims unclaimed batons and spawns the named agent; the agent does one job, writes findings to `site_specs` (the site's shared notebook — spec-not-message), creates the next baton, stops. Full chain: domain-submitter/adoption → classifier → (vertical research) → strategist → briefing → build-site-planner (emits needs_page/design/imagery/rerender) → page-build-handler per page → rerender/deploy. `pageflow-builder` survives as intake's initial-build convenience, demoted to "intake convenience" rather than the spine. A commented-out `"tool"` route in the relay's routing table is the mechanism gap blocking tool/infographics pages from the relay.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1, #B3, #milestone; NOTES_running_synthesis_v4(39).md 2026-07-04 "§B0" through "§B3 CLOSED"
- **relations:** builder route method (BLD-001); build pump (BLD-003); roadmap-phases scope decision gap (site-plan-and-reconciler register, PLAN-033); MVP build squad lineage (BLD-014, predecessor history)
- **verify-later:** workflow dumps of the nine builders; pageflow-builder v20 definition; load_work_item_actions.go routing table; the 37-row dartsonline item chain

### BLD-003 — Build pump and the queue immune system
- **status:** deployed
- **status-evidence:** builder_route(21) §B2: "the scheduler fires build-pipeline-trigger EVERY 30s … The queue's immune system is all ENABLED: claimed-item-timeout (evidence-based auto-complete), feasibility-recheck, stale-orchestration-reaper, stale-work-item-reaper (48h), work-item-archiver, database-cleanup. FLAG: improvement-sweep is DISABLED."
- **what:** What drives the relay: scheduled `build-pipeline-trigger` (30s, pre_query gated, concurrency dispatch/8) → `build-dispatch-loop` → atomic claim → spawn dynamic handler → complete/fail → touch scheduled_tasks. The immune system self-heals the queue (claimed-item-timeout does evidence-based auto-complete; feasibility-recheck unblocks when handlers appear; reapers and archiver bound staleness). Standing flag: improvement-sweep is disabled platform-wide, so the improvement loop is not running; content-feed-refresh is enabled 6-hourly.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2
- **relations:** work-item relay spine (BLD-002); dispatch chain + NOT-EXISTS blocker (work-dispatch register)
- **verify-later:** scheduled_tasks rows (build-pipeline-trigger, improvement-sweep enabled flags); claimed-item-timeout SQL

### BLD-004 — Two front doors and duplicate classifiers (Q5)
- **status:** partial
- **status-evidence:** builder_route(21) "Two front doors, two classifiers (overlap)"; queue item 2 "[MAIN] Q5 front-door consolidation — two classifiers, one responsibility" (queued, undecided).
- **what:** Intake exists twice: the queue door (domain-submitter → work-item relay with domain-research-classifier) and intake-orchestrator v3 (HITL: site-classifier → confirm type → questionnaire → briefing-agent → spawn dynamic builder). site-classifier and domain-research-classifier hold the same responsibility; the classifier prompt hardcodes `recommended_builder="pageflow-builder"`; intake carries orphaned rerender steps. Consolidation direction (deprecate the intake door vs align contracts) is an open decision.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1, #queue (item 2)
- **relations:** work-item relay spine (BLD-002); site_type taxonomy drift
- **verify-later:** intake-orchestrator usage evidence; site-classifier workflow

### BLD-005 — image_tag 'latest' stale-default trap
- **status:** partial
- **status-evidence:** builder_route(21): "INCIDENT 2026-07-06 — first claim STALLED … THE ONE REAL DIFFERENCE: image_tag='latest' (column default) … the registry's latest is an ANCIENT chassis build … FIX APPLIED … NEW PARKED TRAP (systemic): agent_definitions.image_tag DEFAULTS to 'latest' — every future seeded agent inherits it."
- **what:** Seeded agents inherit `image_tag='latest'`, which points at an ancient pre-architecture chassis build (boots the retired generic.process consumer regardless of env) — a newly seeded researcher stalled on it. Immediate fix: copy image columns from a live donor in every seed. Systemic options parked: repoint/retire `latest`, ALTER the column default, or a New Agent checklist line. Rollback convention is the same lever inverted: revert by repointing image_tag to the prior tag.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4 (incident), #queue (item 1)
- **relations:** stale-corpus class; standing evidence rules (seed hygiene)
- **verify-later:** agent_definitions image_tag column default; whether redeploy-agents bumps rows

### BLD-006 — Coverage baseline: guides, tools, news, curated top-N on most sites
- **status:** aspirational
- **status-evidence:** builder_route(21) queue item 7: "standing expectation going forward is most sites should carry guides + tools + news + a curated (LLM-picked, non-affiliate) top-N list … the curated-list mechanism, which IS new"; "STANDING EXPECTATION HOME: 001_development_guide."
- **what:** A platform content-coverage policy: most sites should carry guides, tools, news, and a curated non-affiliate top-N list of the vertical's best products/services with outbound links; "pages need not be original to be best-in-class." Enforcement points are the strategist/planner prompts; the curated-list mechanism is the one genuinely new build needed (reuse candidates: research-agent or the exemplar-researcher crawl pattern feeding a curation step). The mechanism for guides/tools/news already EXISTS (proven on other sites) — an absent one elsewhere is a broken route, not a missing feature.
- **sources:** docs019/RUNBOOK_builder_route(21).md#queue (item 7)
- **relations:** roadmap-phases scope decision gap (site-plan-and-reconciler register, PLAN-033); F0 guides pilot
- **verify-later:** 001 guideline amendment; strategist/planner prompt coverage clauses

### BLD-007 — pageflow-builder (component-based site build orchestration)
- **status:** deployed
- **status-evidence:** Still being patched in Phase 0 imagery work 2026-05-05 (its row backed up before a migration); a full live step chain is documented from later units too, confirming it as GEN-3's active monolith.
- **what:** The central v2-era builder, renamed from multipage-website-builder v3. Spawns planner/content-writer/reviewer/deployer, then: ensure_site_record → call_site_planner → store brief+plan → sync_pages_to_db → populate_nav → asset steps → select_style_collection → set_default_components → render_site_components → get_pages_to_build (filters by build_status) → build_pages_loop (write → review → assemble → deploy per page) → apply_site_design (CSS) → trigger_site_deploy (Cloudflare). The known hazard that sync_pages_to_db can reset page statuses is documented in-file.
- **sources:** 026_pageflow_builder.sql; sql_for_agents_v2/026_pageflow_builder.sql; 107_image_build_handler.sql (backup section)
- **relations:** three builder generations (BLD-002); site-planner (site-plan-and-reconciler register, PLAN-034); page-content-writer (BLD-008)
- **verify-later:** whether new sites still route through pageflow-builder or only via the work-item pipeline

### BLD-008 — page-content-writer (section-by-section content generation)
- **status:** deployed
- **status-evidence:** Continuously patched from v2 era through migration 069 (reads site_specs) and the 107-era imagery direction; 075 gives it idle_timeout 180.
- **what:** Writes one page section-by-section: spawn_research_agent → load_page_components → build_render_context → process_sections_loop (per-section LLM call constrained to that component's llm_field_specs) → compile_page. The prompt is a major behavioural contract: official-contact-only rule, internal-link constraint to listed pages, content_direction/imagery_direction from site_specs, admin content briefs, "Recreate Mode" for adopted sites, and an 18-rule anti-fabrication list (no invented people/testimonials/statistics/case studies). The same agent is documented in the page-build-pipeline register from the no-persistence architectural angle (PBP-009) — this entry covers its SQL/prompt-migration history and behavioural contract.
- **sources:** 023_page_content_writer_agent.sql; sql_for_agents_v2/023_page_content_writer_agent.sql; 069_blog_posts.sql
- **relations:** page-content-writer task-specialist fact (page-build-pipeline register, PBP-009); pageflow-builder (BLD-007); page-build-handler (BLD-011)
- **verify-later:** live prompt_template vs the 023 copies; llm_field_specs source in content_components

### BLD-009 — site-work-orchestrator (unified build/maintenance over site_work_items)
- **status:** deployed
- **status-evidence:** 045 definition; row backed up and patched in Phase 0 imagery work (107, 2026-05-05).
- **what:** Orchestrator that builds sites from prioritized site_work_items rows, calling appropriate handler agents per item, "compatible with pageflow-builder's planner and content writer." The first expression of the unified build/maintenance queue idea, later refined into the one-item-at-a-time build-dispatch-loop.
- **sources:** 045_site_work_orchestrator.sql; 107_image_build_handler.sql
- **relations:** site_work_items table; build-dispatch-loop / work-item build pipeline (BLD-010); unified build & maintenance (work-dispatch register)
- **verify-later:** which orchestrator the live triggers use; site_work_items schema

### BLD-010 — Work-item build pipeline: domain-submitter → dispatch loop → handler agents
- **status:** deployed
- **status-evidence:** 051/052 definitions; migration 075 sets idle timeouts across the whole handler fleet; migration 146 still adding items into the same queue in 2026-07.
- **what:** The current architecture, as reconstructed from the SQL migration history. domain-submitter (068) creates a site record + needs_domain_research item from just a domain. build-pipeline-trigger (052) is a 30-min heartbeat: seeds the build queue, finds one site with pending items, fires build-dispatch-loop (051), which loads the highest-priority claimable item, claims it, spawns+calls the handler agent, marks complete, and if items remain spawns a FRESH dispatch loop (separate orchestration, clean logs). Item chain for a new site: needs_domain_research → needs_strategy → needs_briefing → needs_site_plan → needs_content_page (per page) → images → needs_rerender. Concurrency safety via claim_work_item; health-gating via ai_endpoint_health before claiming.
- **sources:** 051_build_dispatch_loop.sql; 052_build_pipeline_trigger.sql; 068_domain_submitter_agent.sql; 085_ai_endpoint_health_checker.sql
- **relations:** every handler agent in this file; dispatch chain (work-dispatch register); replaces intake-orchestrator
- **verify-later:** LoadWorkItemsAction first_item patch; claim semantics; current item_type → handler_agent routing table

### BLD-011 — page-build-handler (content-page handler with section planning and validation gates)
- **status:** deployed
- **status-evidence:** Migration 065 documents the evolved workflow with plan_sections/validate_content and error paths; migration 070 notes the empty_sections handler switched to it.
- **what:** Wrapper solving "specialist vs handler": page-content-writer generates but doesn't persist, so this handler loads page + specs, plan_sections resolves data sources per section (creating deferred items when sections aren't ready), calls the writer, validate_content checks placeholders/templates/cross-site contamination (blockers → mark_needs_review), then save_page_sections, update_page_status, and deploys via page-rerender. An earlier version (055) was simpler (no plan/validate steps, no deploy).
- **sources:** 065_page_build_handler_wrapper.sql; 055_page_build_handler.sql; 070_blog_content_planner.sql
- **relations:** page-build-handler build path (page-build-pipeline register, PBP-008); page-content-writer (BLD-008); page-rebuild (BLD-013)
- **verify-later:** plan_sections + validate_page_content actions; deferred-item creation

### BLD-012 — MVP build squad lineage (chief-strategist → architect → content-creator → deployer)
- **status:** superseded
- **status-evidence:** docs009/001 defines the 4 MVP agents with Kafka payloads; docs015/004 shows the mature descendant ("pageflow-builder: ensure_site_record → call_site_planner → sync_pages → populate_nav → image generation → style collection → build_pages_loop").
- **what:** The builder pipeline's earliest evolutionary line, predating the GEN-1/2/3 framing in BLD-002: mvp-site-builder (4 agents, single page) → landing-page-builder / content-site-builder (specialist architects per site type) → multipage-website-builder (batching, then sequential loop) → pageflow-builder (DB-backed pages, per-page loop with review and git commit) → site-work-orchestrator (work-item driven). Each generation kept strategist/planner, writer, assembler, deployer roles while moving state from CollectedData into the database.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-Agent-chief-strategist; docs006_workflow_builder/003_current_state_of_agents.sql; docs015_data_flow_verification/004_builder_flow.md
- **relations:** three builder generations (BLD-002, later-era continuation of this lineage); sequential page generation (BLD-013); pageflow-builder (BLD-007)
- **verify-later:** which builder agent_definitions still exist and which have traffic

### BLD-013 — Sequential page generation (Phase 0 multipage fix)
- **status:** superseded
- **status-evidence:** docs010/019 "Current Problem: Multipage-website-builder tries to generate 4 pages at once — race conditions... Solution: make it work like landing-page-builder: sequential, one page at a time"; two-delay spawn fix noted FIXED in docs009/003.
- **what:** Replacing parallel batch spawning with a strategist-planned page list iterated by the loop action (research → write per page), a `wrap_multipage` action generating navigation and collecting assets, and spawn timing fixed by double initialization delays. The stabilization step that made multipage builds reliable enough for everything after.
- **sources:** docs010_multitrack_flows_persona_architecture/019_start_here_document.md; docs009_site_interrogation_and_solutions/003_claude_save_point.md#Status
- **relations:** MVP build squad lineage (BLD-012); multi-page site support (site-plan-and-reconciler register, PLAN-037, the following generation); pageflow-builder (BLD-007, successor)
- **verify-later:** wrap_multipage action; spawn_actions.go delay logic

### BLD-014 — Selective rebuild via build_status
- **status:** deployed
- **status-evidence:** docs018/004 ("get_pages_to_build filters by build_status IN ('planned','needs_rebuild')"); docs015/003 documents the stale-page trap ("If the site planner didn't include use-cases in the new plan... it shows in nav but has stale content").
- **what:** Two orthogonal page state columns — status (active/deleted/needs_attention lifecycle) and build_status (planned/needs_rebuild/deployed) — let rebuilds touch only marked pages. Known failure mode: pages absent from a new plan silently keep old content while remaining in nav. Ancestor of work-item-driven rebuild targeting (BLD-010).
- **sources:** docs018_rerendering/004_trigger_just_pages_that_need_rebuild.md; docs015_data_flow_verification/003_temp_doc_rebuild_flow.md; docs017_legacy_agent_rules_images_design_keydocs/041_page_rebuild_action.md
- **relations:** page-rebuild (BLD-015); maintenance_queue (proto-work-items); work items (work-dispatch register)
- **verify-later:** get_pages_to_build action; build_status usage today

### BLD-015 — page-rebuild (rebuild pages without re-planning)
- **status:** deployed
- **status-evidence:** 039 full definition with detailed reuse/skip lists; no later references found in the source unit.
- **stage2-verified (2026-07-14):** unknown → deployed — page-rebuild is actively wired post-dispatch-loop-refactor: maintenance-triage agent (k8s/bk_agent_definitions_backup.sql:183, status='active') workflow calls agent_type 'page-rebuild' via call_agent in a rebuild_loop step; live Go actions back this end-to-end — platform/orchestration/actions/maintenance_actions.go ...
- **what:** Rebuilds specific pages (build_status='needs_rebuild') on an existing site, loading all context from DB given a domain, explicitly skipping planner, sync_pages_to_db, asset generation, component rendering, CSS and nav (all already done) while reusing the standard build-loop agents. Documents design principles: agent owns its domain; spawnable not standalone; reuse before creating; complexity in Go.
- **sources:** 039_page_rebuild_agent.sql
- **relations:** pageflow-builder (BLD-007, same loop, different input_mapping via rebuild_context); selective rebuild via build_status (BLD-014)
- **verify-later:** whether page-rebuild survived the dispatch-loop refactor
