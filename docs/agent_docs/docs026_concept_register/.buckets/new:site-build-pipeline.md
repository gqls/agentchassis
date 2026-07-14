
<!-- SOURCE: U21_legacy_docs_b.md -->
### MVP build squad lineage (chief-strategist → architect → content-creator → deployer)
- **category:** NEW:site-build-pipeline
- **status-signal:** superseded
- **status-evidence:** docs009/001 defines the 4 MVP agents with Kafka payloads; docs015/004 shows the mature descendant ("pageflow-builder: ensure_site_record → call_site_planner → sync_pages → populate_nav → image generation → style collection → build_pages_loop").
- **what:** The builder pipeline's evolutionary line: mvp-site-builder (4 agents, single page) → landing-page-builder / content-site-builder (specialist architects per site type) → multipage-website-builder (batching, then sequential loop) → pageflow-builder (DB-backed pages, per-page loop with review and git commit) → site-work-orchestrator (work-item driven). Each generation kept strategist/planner, writer, assembler, deployer roles while moving state from CollectedData into the database.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-Agent-chief-strategist; docs006_workflow_builder/003_current_state_of_agents.sql#SPECIALIST-ARCHITECT-SYSTEM; docs015_data_flow_verification/004_builder_flow.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md
- **relations:** unified work items (final form); loop action; site-planner; deployment-github.
- **verify-later:** which builder agent_definitions still exist and which have traffic.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Sequential page generation (Phase 0 multipage fix)
- **category:** NEW:site-build-pipeline
- **status-signal:** superseded
- **status-evidence:** docs010/019 "Current Problem: Multipage-website-builder tries to generate 4 pages at once — race conditions... Solution: make it work like landing-page-builder: sequential, one page at a time"; two-delay spawn fix noted FIXED in docs009/003.
- **what:** Replacing parallel batch spawning with a strategist-planned page list iterated by the loop action (research → write per page), a wrap_multipage action generating navigation and collecting assets, and spawn timing fixed by double initialization delays. The stabilization step that made multipage builds reliable enough for everything after.
- **sources:** docs010_multitrack_flows_persona_architecture/019_start_here_document.md; docs009_site_interrogation_and_solutions/003_claude_save_point.md#Status
- **relations:** loop action; batched generation (predecessor); pageflow-builder (successor).
- **verify-later:** wrap_multipage action; spawn_actions.go delay logic.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Selective rebuild via build_status
- **category:** NEW:site-build-pipeline
- **status-signal:** deployed
- **status-evidence:** docs018/004 ("get_pages_to_build filters by build_status IN ('planned','needs_rebuild')"); docs015/003 documents the stale-page trap ("If the site planner didn't include use-cases in the new plan... it shows in nav but has stale content").
- **what:** Two orthogonal page state columns — status (active/deleted/needs_attention lifecycle) and build_status (planned/needs_rebuild/deployed) — let rebuilds touch only marked pages. Known failure mode: pages absent from a new plan silently keep old content while remaining in nav. Ancestor of work-item-driven rebuild targeting.
- **sources:** docs018_rerendering/004_trigger_just_pages_that_need_rebuild.md; docs015_data_flow_verification/003_temp_doc_rebuild_flow.md; docs017_legacy_agent_rules_images_design_keydocs/041_page_rebuild_action.md
- **relations:** page-rebuild agent; maintenance_queue (proto-work-items); work items.
- **verify-later:** get_pages_to_build action; build_status usage today.

<!-- SOURCE: U21_legacy_docs_b.md -->
### MVP build squad lineage (chief-strategist → architect → content-creator → deployer)
- **category:** NEW:site-build-pipeline
- **status-signal:** superseded
- **status-evidence:** docs009/001 defines the 4 MVP agents with Kafka payloads; docs015/004 shows the mature descendant ("pageflow-builder: ensure_site_record → call_site_planner → sync_pages → populate_nav → image generation → style collection → build_pages_loop").
- **what:** The builder pipeline's evolutionary line: mvp-site-builder (4 agents, single page) → landing-page-builder / content-site-builder (specialist architects per site type) → multipage-website-builder (batching, then sequential loop) → pageflow-builder (DB-backed pages, per-page loop with review and git commit) → site-work-orchestrator (work-item driven). Each generation kept strategist/planner, writer, assembler, deployer roles while moving state from CollectedData into the database.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-Agent-chief-strategist; docs006_workflow_builder/003_current_state_of_agents.sql#SPECIALIST-ARCHITECT-SYSTEM; docs015_data_flow_verification/004_builder_flow.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md
- **relations:** unified work items (final form); loop action; site-planner; deployment-github.
- **verify-later:** which builder agent_definitions still exist and which have traffic.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Sequential page generation (Phase 0 multipage fix)
- **category:** NEW:site-build-pipeline
- **status-signal:** superseded
- **status-evidence:** docs010/019 "Current Problem: Multipage-website-builder tries to generate 4 pages at once — race conditions... Solution: make it work like landing-page-builder: sequential, one page at a time"; two-delay spawn fix noted FIXED in docs009/003.
- **what:** Replacing parallel batch spawning with a strategist-planned page list iterated by the loop action (research → write per page), a wrap_multipage action generating navigation and collecting assets, and spawn timing fixed by double initialization delays. The stabilization step that made multipage builds reliable enough for everything after.
- **sources:** docs010_multitrack_flows_persona_architecture/019_start_here_document.md; docs009_site_interrogation_and_solutions/003_claude_save_point.md#Status
- **relations:** loop action; batched generation (predecessor); pageflow-builder (successor).
- **verify-later:** wrap_multipage action; spawn_actions.go delay logic.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Selective rebuild via build_status
- **category:** NEW:site-build-pipeline
- **status-signal:** deployed
- **status-evidence:** docs018/004 ("get_pages_to_build filters by build_status IN ('planned','needs_rebuild')"); docs015/003 documents the stale-page trap ("If the site planner didn't include use-cases in the new plan... it shows in nav but has stale content").
- **what:** Two orthogonal page state columns — status (active/deleted/needs_attention lifecycle) and build_status (planned/needs_rebuild/deployed) — let rebuilds touch only marked pages. Known failure mode: pages absent from a new plan silently keep old content while remaining in nav. Ancestor of work-item-driven rebuild targeting.
- **sources:** docs018_rerendering/004_trigger_just_pages_that_need_rebuild.md; docs015_data_flow_verification/003_temp_doc_rebuild_flow.md; docs017_legacy_agent_rules_images_design_keydocs/041_page_rebuild_action.md
- **relations:** page-rebuild agent; maintenance_queue (proto-work-items); work items.
- **verify-later:** get_pages_to_build action; build_status usage today.
