# EXTRACTION U21 — mid-era legacy docs (docs005–docs018, excluding docs008/013)
Extracted 2026-07-13. Files in scope: 102. Concepts found: 58.

Directories: docs005_briefing_agent_domain_authority, docs006_workflow_builder,
docs007_brochure_builder, docs009_site_interrogation_and_solutions,
docs010_multitrack_flows_persona_architecture, docs011_api_hitl,
docs012_site_maps_and_components, docs014_research_agent,
docs015_data_flow_verification, docs016_dogs_medicine_pathways,
docs017_legacy_agent_rules_images_design_keydocs, docs018_rerendering.

Era: roughly Dec 2025 (docs006–012, image tags v1.0.478–510) through Feb–Mar 2026
(docs018 dated 2026-02-06; docs016 dated 2026-03-02/03). Many concepts here are the
direct ancestors of the current build/improvement-loop system; others were designed
in detail and silently dropped.

## Coverage
| file | treatment |
|---|---|
| docs005_briefing_agent_domain_authority/README.0130.briefing_agent.md | full |
| docs006_workflow_builder/001_workflow_builder.md | full |
| docs006_workflow_builder/002_removing_agent_group_definitions.md | full |
| docs006_workflow_builder/003_current_state_of_agents.sql | header-scan |
| docs006_workflow_builder/004_agent_groups_or_not.md | full |
| docs006_workflow_builder/005_acme_corp_org_chart.md | full |
| docs006_workflow_builder/006_conclude_role_entity_strategy.md | full |
| docs006_workflow_builder/006b_evolution_design_discussion.md | family-delta (duplicate of 004) |
| docs006_workflow_builder/006c_org_framework_discussion.md | family-delta (duplicate of 006 part 1) |
| docs006_workflow_builder/007_new_tables_entity_state_log.sql | header-scan |
| docs006_workflow_builder/008_20_plus_pages.md | full |
| docs006_workflow_builder/009_massive_multipage_sites.md | full |
| docs006_workflow_builder/010_debugging.md | full |
| docs006_workflow_builder/011_working_landing_page_builder.md | full |
| docs007_brochure_builder/001_brochure_builder_plan.md | full |
| docs007_brochure_builder/003_original_message_copy | full |
| docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md | full |
| docs009_site_interrogation_and_solutions/002_claude_discussion | full |
| docs009_site_interrogation_and_solutions/003_claude_save_point.md | full |
| docs009_site_interrogation_and_solutions/004_multitrack_sitemap_architecture_different_flows.md | full |
| docs010_multitrack_flows_persona_architecture/001_implement_multi_track_flow.md | full |
| docs010_multitrack_flows_persona_architecture/002_multi_track_schema.sql | header-scan |
| docs010_multitrack_flows_persona_architecture/003_multi_track_configuration_guide.md | header-scan |
| docs010_multitrack_flows_persona_architecture/004_multipage_workflow_with_flows.md | header-scan |
| docs010_multitrack_flows_persona_architecture/005_implementation_summary.md | full |
| docs010_multitrack_flows_persona_architecture/006_examples.md | header-scan |
| docs010_multitrack_flows_persona_architecture/007_personas_discussion.md | full |
| docs010_multitrack_flows_persona_architecture/008_example_personas.md | header-scan |
| docs010_multitrack_flows_persona_architecture/009_persona_system_schema.sql | header-scan |
| docs010_multitrack_flows_persona_architecture/010_persona_system_schema.sql | header-scan (persona roster seed data) |
| docs010_multitrack_flows_persona_architecture/011_persona_cognitive_architecture.sql | header-scan |
| docs010_multitrack_flows_persona_architecture/012_persona_cognitive_actions.sql | header-scan |
| docs010_multitrack_flows_persona_architecture/014_drBimpton_setup_example.sql | header-scan |
| docs010_multitrack_flows_persona_architecture/015_persona_README_architecture.md | full |
| docs010_multitrack_flows_persona_architecture/016_persona_deployment_post | header-scan (duplicate of 015 material) |
| docs010_multitrack_flows_persona_architecture/017_implementation_plan_from_here | header-scan |
| docs010_multitrack_flows_persona_architecture/018_priority_matrix.md | header-scan |
| docs010_multitrack_flows_persona_architecture/019_start_here_document.md | full |
| docs010_multitrack_flows_persona_architecture/020_revised_consolidated_action_plan.md | header-scan |
| docs010_multitrack_flows_persona_architecture/021_loop_action_discussion.md | header-scan |
| docs010_multitrack_flows_persona_architecture/022_loop_actions_guide.md | header-scan |
| docs010_multitrack_flows_persona_architecture/023_loop_explanation.md | full |
| docs011_api_hitl/001_hitl_api_analysis.md | full |
| docs011_api_hitl/002_implementation.md | full |
| docs011_api_hitl/003_hitl_new_plan.md | full |
| docs012_site_maps_and_components/001_plan_from_here.md | header-scan |
| docs012_site_maps_and_components/002_site_map_integration.md | header-scan |
| docs012_site_maps_and_components/003_semantic_linking.md | header-scan |
| docs012_site_maps_and_components/004_more_on_links.md | header-scan |
| docs012_site_maps_and_components/005_more_on_links | skipped-generated (zero-length file) |
| docs012_site_maps_and_components/006_start_concluding_links.md | full |
| docs012_site_maps_and_components/007_link_migration.sql | header-scan |
| docs012_site_maps_and_components/008_link_integration_guide.md | header-scan |
| docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md | full |
| docs012_site_maps_and_components/010_component_and_site_architecture.md | header-scan |
| docs012_site_maps_and_components/011_updated_flow.md | full |
| docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md | full |
| docs014_research_agent/001_human_in_the_loop_response_flow.md | full |
| docs015_data_flow_verification/001_data_flow_verification.md | full |
| docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md | full |
| docs015_data_flow_verification/003_temp_doc_rebuild_flow.md | full |
| docs015_data_flow_verification/004_builder_flow.md | full |
| docs016_dogs_medicine_pathways/002_project_outline.md | full |
| docs016_dogs_medicine_pathways/003_canine_biology_project_baseline.md | family-delta |
| docs016_dogs_medicine_pathways/003b_canine_biology_project_baseline_v2.md | family-delta |
| docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md | family-latest |
| docs016_dogs_medicine_pathways/004_medical_business_reality_assessment.md | full |
| docs017_legacy_agent_rules_images_design_keydocs/001_changes_needed.md | full |
| docs017_legacy_agent_rules_images_design_keydocs/001_flexible_schema_enforcement.md | full |
| docs017_legacy_agent_rules_images_design_keydocs/002_full_new_agent_architecture.md | family-delta (v1 of 019b) |
| docs017_legacy_agent_rules_images_design_keydocs/002_pageflow_image_changes.md | full |
| docs017_legacy_agent_rules_images_design_keydocs/002_standardising_deployment_implementation_plan.md | header-scan |
| docs017_legacy_agent_rules_images_design_keydocs/003_design.md | full |
| docs017_legacy_agent_rules_images_design_keydocs/004_checklist_for_new_specialist_agent_v1.md | family-delta |
| docs017_legacy_agent_rules_images_design_keydocs/005_checklist_for_new_specialist_agent_v2.md | family-delta (dropped: "Use input_fields, Not Explicit Paths") |
| docs017_legacy_agent_rules_images_design_keydocs/006_checklist_for_new_specialist_agent_v3.md | family-delta |
| docs017_legacy_agent_rules_images_design_keydocs/007_checklist_for_new_specialist_agent_v4.md | family-delta |
| docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md | family-latest |
| docs017_legacy_agent_rules_images_design_keydocs/017_agent_architecture_v2.md | family-delta (design layers only) |
| docs017_legacy_agent_rules_images_design_keydocs/018_agent_architecture_v3.md | family-delta |
| docs017_legacy_agent_rules_images_design_keydocs/019_agent_architecture_v4.md | family-delta |
| docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md | family-latest |
| docs017_legacy_agent_rules_images_design_keydocs/021_maintenance_architecture_plan_v1.md | family-delta |
| docs017_legacy_agent_rules_images_design_keydocs/022_maintenance_architecture_plan_v2.md | family-delta (spawn chain, vet-batch precedent, budget mgmt) |
| docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md | family-latest |
| docs017_legacy_agent_rules_images_design_keydocs/030_news_feeds_v1.md | family-delta (v1 of 019b feed section) |
| docs017_legacy_agent_rules_images_design_keydocs/041_page_rebuild_action.md | full |
| docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md | full |
| docs017_legacy_agent_rules_images_design_keydocs/043_section_naming_contract.md | full |
| docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md | full |
| docs018_rerendering/001_rerender_pages_summary.md | full |
| docs018_rerendering/002_summary_link_constraints.md | header-scan |
| docs018_rerendering/003_website_builder_architecture_status_report.md | full |
| docs018_rerendering/004_trigger_just_pages_that_need_rebuild.md | header-scan |
| docs018_rerendering/005_triggering_agent_from_kafka.md | header-scan |
| docs018_rerendering/006_build_path_rerender_path.md | full |
| docs018_rerendering/007_proposed_modular_rerendering.md | full |
| docs018_rerendering/008_granular_editing.md | header-scan |
| docs018_rerendering/008b_my_notes_what_I_can_do | full |
| docs018_rerendering/009_agent_initial_message_structure.md | header-scan |
| docs018_rerendering/009_stale_orchestration_sweeper_design.md | full |
| docs018_rerendering/010_section_editor_architecture.md | full |

## Concepts

### Briefing agent (pre-strategist brief enrichment)
- **category:** onboarding-config
- **status-signal:** superseded
- **status-evidence:** docs005 sketch ("A new agent type briefing-agent... Sits before chief-strategist"); docs006/011 shows it live in the intake workflow ("call_briefer → Briefer fills questionnaire (HITL or LLM)").
- **what:** An agent inserted before the strategist that takes raw user input (domain, rough objective), asks clarifying questions, and outputs a structured brief (audience, tone, USPs, competitors, key messages) with a human approval pause. Evolved into the briefing-agent + per-builder `briefing_questionnaire` with interactive (HITL) and auto (LLM-infer) modes.
- **sources:** docs005_briefing_agent_domain_authority/README.0130.briefing_agent.md; docs006_workflow_builder/011_working_landing_page_builder.md#Briefing-Agent; docs006_workflow_builder/003_current_state_of_agents.sql#3-BRIEFING-AGENT
- **relations:** builder questionnaires per site type; intake orchestrator; HITL pauses; successor: reviewed_brief in current build pipeline.
- **verify-later:** agent_definitions row 'briefing-agent'; briefing_questionnaire column on agent_definitions; current intake workflow.

### Per-builder briefing questionnaires
- **category:** onboarding-config
- **status-signal:** superseded
- **status-evidence:** docs006/002 full questionnaire JSON on landing-page-builder and content-site-builder definitions; docs007/001 contrasts landing (10 conversion fields) vs brochure (15+ corporate fields).
- **what:** Each builder agent definition carries a `briefing_questionnaire` JSONB (sections of typed questions — brand, value proposition, conversion, social proof for landing; company, services, leadership, case studies for brochure). `fetch_agent_questionnaire` retrieves the correct questionnaire for the chosen builder, and the briefing agent fills it via HITL or LLM inference.
- **sources:** docs006_workflow_builder/002_removing_agent_group_definitions.md#Step-2; docs007_brochure_builder/001_brochure_builder_plan.md#Questionnaire-Differences; docs006_workflow_builder/003_current_state_of_agents.sql
- **relations:** briefing agent; site classifier; reviewed_brief.
- **verify-later:** briefing_questionnaire values in agent_definitions; fetch_agent_questionnaire action in Go.

### Workflow Builder & Validator (YAML DSL)
- **category:** NEW:workflow-authoring
- **status-signal:** abandoned
- **status-evidence:** docs006/001 full design with roadmap claiming "[x] Phase 1: Core parser & validator, [x] Phase 2: Path resolution, [x] Phase 3: JSON generation"; no later doc references the tool; workflows continued to be hand-written SQL.
- **what:** A validation-first system for authoring orchestration workflows in human-readable YAML instead of raw JSON: parses a DSL, validates agent types exist in agent_definitions, detects circular dependencies and invalid input references, auto-computes CollectedData paths (agent call vs local action nesting), generates the orchestration_workflow JSON, test cases, and docs, then inserts into the DB. CLI (`workflow-builder build/validate/test/list/show/docs`), planned HTTP API, web UI, and git-based CI/CD workflow deployment.
- **sources:** docs006_workflow_builder/001_workflow_builder.md#Architecture; docs006_workflow_builder/001_workflow_builder.md#Path-Resolution; docs006_workflow_builder/001_workflow_builder.md#Roadmap
- **relations:** data-path resolution problem; workflow validator tool (docs017/002_standardising); superseded in spirit by input_mapping/ActionInputSpec conventions.
- **verify-later:** platform/workflowbuilder/ directory existence in repo history; any workflow YAML files.

### Data-path resolution problem (agent vs local action nesting)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** docs006/001 ("Local Action: CollectedData[\"wrap_multipage\"] ... input_data.site_files.wrap_multipage ← Extra layer!"); docs009/002 ("collected_data.spawn_x.call_x.spawn_y...result"); resolved later by input_mapping + ActionInputSpec (docs017/008).
- **what:** The recurring class of runtime failures where workflow config referenced CollectedData paths that didn't match where actions actually stored results — agent calls store flat, local actions add a step-name layer, and each spawn/call deepens nesting. Drove multiple generations of mitigation: workflow builder path computation, explicit output_field conventions, data-flow verification matrices, and finally standardized input extraction.
- **sources:** docs006_workflow_builder/001_workflow_builder.md#The-Problem; docs009_site_interrogation_and_solutions/002_claude_discussion#C; docs015_data_flow_verification/001_data_flow_verification.md
- **relations:** ActionInputSpec/ExtractActionInputs; workflow builder; data-flow verification practice.
- **verify-later:** datahelpers.ResolveInputMapping and FindByPath in platform code.

### "Every agent is an orchestrator" — elimination of agent_group_definitions
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** docs006/006 implementation doc: "GroupDiscovery is aliased to AgentDefinitionDiscovery... FindBestGroup now queries agent_definitions"; backward-compat table showing both group_type and agent_type message formats work.
- **what:** Architectural unification: a "group" is just an agent whose workflow spawns and calls other agents, so agent_group_definitions was eliminated in favour of agent_definitions carrying the orchestration workflow in default_config. spawn_group became a thin wrapper delegating to spawn_agent; discovery, message processor, and metadata all gained aliases for backward compatibility. This is the foundational premise of the current hierarchical agent tree.
- **sources:** docs006_workflow_builder/002_removing_agent_group_definitions.md#Part-2; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Implementation; docs006_workflow_builder/004_agent_groups_or_not.md#Key-Decisions
- **relations:** spawn-before-call pattern; intake orchestrator; agent families.
- **verify-later:** platform/discovery/agent_discovery.go; absence/deprecation of agent_group_definitions table; spawn_group action code.

### entity_state_log — append-only cross-orchestration memory
- **category:** NEW:agent-memory-and-evolution
- **status-signal:** abandoned
- **status-evidence:** Full schema + five Go actions (append_entity_state, read_latest_entity_state, read_entity_history, read_my_state, write_my_state) in docs006/002 and migration SQL in docs006/007; no later documents reference entity_state_log in the build pipeline.
- **what:** Persistent data that survives across orchestrations: an append-only log keyed by entity_id/namespace/path with accumulation patterns (additive, evolutionary, versioned, singleton), agent-namespaced storage ("read_my_state"/"write_my_state" use AGENT_TYPE as namespace), supersession pointers for future compaction, and LLM-based consolidation as a future enhancement. Intended for accumulating research, brand learnings, and build history per domain.
- **sources:** docs006_workflow_builder/002_removing_agent_group_definitions.md#Part-5; docs006_workflow_builder/007_new_tables_entity_state_log.sql; docs006_workflow_builder/004_agent_groups_or_not.md#Where-Learnings-Live
- **relations:** four-level learnings model; relationships table; improvement_proposals; conceptual ancestor of per-site content_data accumulation.
- **verify-later:** entity_state_log table existence in clients_db; entity_state_actions.go in repo.

### Agent variants + snapshot versioning
- **category:** NEW:agent-memory-and-evolution
- **status-signal:** abandoned
- **status-evidence:** docs006/004: "Snapshot model (preferred): variants explicitly reference a snapshot version"; is_snapshot column added in docs006/007 migration; agent_variants table proposed but never seen again.
- **what:** A controlled-evolution model for agent definitions: base agents are versioned and can be frozen as snapshots (is_snapshot flag); task variants (agent_variants) reference a specific base version with config_overrides, metrics, and lineage, so the base can evolve without breaking variants. Three evolution types (bug fix / improvement / innovation) with escalating oversight; promotion of successful variants to new bases left as an open question.
- **sources:** docs006_workflow_builder/004_agent_groups_or_not.md#The-Fragility-Problem; docs006_workflow_builder/006b_evolution_design_discussion.md; docs006_workflow_builder/007_new_tables_entity_state_log.sql
- **relations:** improvement_proposals; four-level learnings model; agent_definitions versioning today.
- **verify-later:** is_snapshot/usage_count columns on agent_definitions; whether agent_variants table was ever created.

### improvement_proposals — HITL-gated agent evolution queue
- **category:** NEW:agent-memory-and-evolution
- **status-signal:** abandoned
- **status-evidence:** docs006/004: "The system proposes changes but requires HITL approval before applying"; docs006/006 lists ReviewPerformanceAction → improvement_proposals and ApproveImprovementAction; not referenced by later architectures.
- **what:** A review queue where proposed changes to agent definitions, variants, or entity knowledge — sourced from metrics regressions, agent observations, or humans — wait as pending proposals until a human approves, rejects, or applies them. Included review_performance action recording execution metrics to entity_state_log and generating proposals.
- **sources:** docs006_workflow_builder/004_agent_groups_or_not.md#What-Triggers-Evolution; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Discovery-Actions-Changes
- **relations:** conceptual ancestor of the improvement-loop's suggest/flag resolution paths and HITL approvals.
- **verify-later:** improvement_proposals table; discovery_actions.go history.

### relationships table — first-class entity relationships
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** Created in docs006/007 migration; docs012/006 (later era) says "Relationships (existing, empty)... This is PERFECT for semantic links between pages!" — table existed but unused, then earmarked for semantic page links.
- **what:** A generic first-class relationship entity (source/target entity id+type, relationship_type, direction, properties JSONB, status) modelled explicitly on website links ("relationships are like links — first-class objects with their own identity and state"), with relationship-scoped entity_state for learned communication preferences. Designed for org-framework roles, reused conceptually for pillar↔cluster semantic page relationships.
- **sources:** docs006_workflow_builder/006_conclude_role_entity_strategy.md#Relationships-as-First-Class-Objects; docs006_workflow_builder/007_new_tables_entity_state_log.sql#7; docs012_site_maps_and_components/006_start_concluding_links.md#Part-1
- **relations:** link-management (semantic links); org framework.
- **verify-later:** relationships table in clients_db and whether any rows exist; link_registry vs relationships usage.

### Organizational framework (roles, listeners, policy-as-filters)
- **category:** NEW:org-framework
- **status-signal:** abandoned
- **status-evidence:** Extended thought experiment across docs006/005, /006, /006c ("Acme Corp", "Sarah, Marketing Content Writer"); "Open Items for Later" never picked up; no later doc builds on roles/listeners.
- **what:** A design showing the framework is domain-agnostic by modelling a whole company: roles typed as identity/function/composite/position (only identity roles get schemas, like client_X); employees as clients with personal agent_instances; always-on shared listeners (like adapters) that spawn discrete orchestrations per task ("Sarah isn't running, she's ready"); authority as conditional filters (policy-owner agents like legal-review-agent injected by trigger conditions) rather than hierarchy; strategy flowing down as intake, decomposing at each level. Cross-cutting agents concluded to be "just agents many workflows call."
- **sources:** docs006_workflow_builder/005_acme_corp_org_chart.md; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Role-vs-Agent; docs006_workflow_builder/006c_org_framework_discussion.md
- **relations:** entity_state_log; relationships; policy filters prefigure legal-content-agent constraints; "strategy as intake" prefigures autonomous mission decomposition.
- **verify-later:** roles/role_assignments tables (expected absent).

### Site classifier agent
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** docs006/003 SQL defines 'site-classifier' with a Haiku prompt classifying landing/content/portfolio/brochure; docs015/004 confirms "Single LLM call → outputs ONE site_type... ONE recommended_builder"; current system uses the multi-aspect site-spec classifier (docs024 021).
- **what:** A lightweight LLM agent that classifies a domain+objective into a site type (landing/content/portfolio/brochure) with confidence, reasoning, detected industry and signals, and recommends the corresponding builder group. Its single-label output was later superseded by the richer site-spec aspect classification.
- **sources:** docs006_workflow_builder/003_current_state_of_agents.sql#2-SITE-CLASSIFIER; docs007_brochure_builder/001_brochure_builder_plan.md#Classification-Signals; docs015_data_flow_verification/004_builder_flow.md
- **relations:** intake orchestrator; HITL type confirmation; successor: site-spec-and-classifier architecture.
- **verify-later:** agent_definitions 'site-classifier' vs current classifier agents.

### Intake orchestrator workflow (classify → brief → spawn builder)
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** docs006/011: "✅ WORKING SYSTEM: HITL Orchestration with Multi-Agent Workflow" listing the 11-step intake workflow with two HITL pauses; docs017/023 later introduces intake-orchestrator-v2 routing.
- **what:** The entry-point orchestration: spawn/call site-classifier → fetch_available_builders from DB → HITL confirm site type (human can override classifier and builder choice) → fetch builder questionnaire → briefing agent fills it → HITL review brief → spawn and call the chosen builder. Established the pattern of human quality-gates before expensive generation.
- **sources:** docs006_workflow_builder/011_working_landing_page_builder.md#Working-Agents; docs015_data_flow_verification/004_builder_flow.md; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Naming-and-Coexistence
- **relations:** site classifier; briefing agent; intake-orchestrator-v2; HITL protocol.
- **verify-later:** intake-orchestrator workflow JSON in agent_definitions; whether v2 routing exists.

### HITL request/response protocol (awaited requests, three IDs)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** docs006/011 "Key Discovery: DO NOT use orchestration's original request_id; DO use HITL action's generated request_id from logs"; docs014/001 documents the same protocol with reply_to_topic and cleanup behaviour.
- **what:** The mechanism by which workflows pause for human input: request_human_input registers an entry in AwaitedRequests with a freshly generated request token; the human response (Kafka message with correlation_id, orchestration_id, in_response_to_request_id, in_response_to_step_name headers) is matched by the coordinator, which removes the awaited entry, stores the data, and resumes. Multiple sequential HITL pauses supported. Notable operational pain: request IDs had to be grepped from pod logs.
- **sources:** docs006_workflow_builder/011_working_landing_page_builder.md#HITL-Message-Requirements; docs014_research_agent/001_human_in_the_loop_response_flow.md; docs007_brochure_builder/003_original_message_copy
- **relations:** HITL API endpoint; parent-timeout race; system.notifications.ui.
- **verify-later:** awaited_requests table; request_human_input action; coordinator response matching code.

### HITL API endpoint (/api/v1/hitl/respond)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** docs011/001 recommends "quick fix now, API endpoint later (2-3 hours)"; docs011/002 is the complete handler implementation guide "For Future Implementation"; the admin dashboard later exposes HITL response UI (per docs024 012 era).
- **what:** An HTTP gateway endpoint that accepts a JSON HITL response (correlation_id, orchestration_id, request_id, step_name, data) and constructs the correctly-headed Kafka response message, replacing fragile hand-built kcat commands (whose immediate bug was unsubstituted `${VAR}` template strings inside single-quoted heredocs).
- **sources:** docs011_api_hitl/001_hitl_api_analysis.md; docs011_api_hitl/002_implementation.md
- **relations:** HITL protocol; admin-dashboard-and-api; system.notifications.ui consumer gap.
- **verify-later:** internal/gateway/hitl_handler.go; route registration in production-api.

### system.notifications.ui topic and the missing HITL UI service
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** docs014/001: "It appears there is no service consuming system.notifications.ui to present HITL requests to humans" with input_requests/awaited_requests SQL to check pending items.
- **what:** HITL escalations publish a rich notification (request_type, reply_to_topic, ui_config with title/description/issues_field, timeout, editable flag) to a dedicated UI topic; a consumer service was required to display requests and collect responses. At the time nothing consumed it — the documented alternatives were building the UI service or raising auto-approval thresholds. Later matured into the admin dashboard's HITL surface.
- **sources:** docs014_research_agent/001_human_in_the_loop_response_flow.md#What-Needs-to-Consume
- **relations:** HITL API endpoint; content-reviewer escalate_to_human; admin-dashboard-and-api.
- **verify-later:** consumers of system.notifications.ui; input_requests table.

### Parent-timeout vs child-HITL race
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** docs014/001 log trace: "pageflow-builder times out (5 min)... content-reviewer: Cleaned up expired awaited requests count=1. The fix is to increase the parent's timeout."
- **what:** A failure class where a parent's call_agent timeout fires before the child's HITL request can be answered; the parent retries with null body and the child's awaited request is cleaned up as expired, losing the pause. Fix: parent timeouts must exceed child HITL timeout windows.
- **sources:** docs014_research_agent/001_human_in_the_loop_response_flow.md#Why-There-Were-No-Awaited-Requests
- **relations:** stale orchestration sweeper; HITL protocol; timeout heuristics in debugging docs.
- **verify-later:** current call_agent timeout_seconds vs HITL timeout defaults.

### HTML action architecture (generate → process → validate)
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** docs006/008: "ALWAYS use the HTML actions instead of raw LLM calls... The architecture is already there — use it!"; replaced wholesale by component-template rendering in docs012+.
- **what:** A three-action pipeline for LLM page generation: `generate_html` (auto-gathers context from analyze_domain/architect_site/create_content/input_data, builds optimized prompt, extracts clean HTML), `process_html` (goquery parsing, meta tags, OG tags, responsive checks, lazy loading, minification), `validate_html` (structure, required elements, image alts, links, accessibility). Plus `assemble_html_parts` for chunking one huge page into structure/styles/content generations.
- **sources:** docs006_workflow_builder/008_20_plus_pages.md#The-HTML-Actions; docs006_workflow_builder/009_massive_multipage_sites.md#The-Actions-Available
- **relations:** superseded by content_components template rendering + render_mode matrix; chunked generation.
- **verify-later:** html_actions.go survival/usage in current action registry.

### Batched multipage generation (assemble_multipage_site)
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** docs006/009: "for 20+ pages you need assemble_multipage_site... 5 batches × 4 pages = 80k tokens = WORKS"; docs010/019 then replaces batching: "Current (broken): spawn_multiple_writers ❌ Spawns 4 at once → New: loop".
- **what:** Handling 6–200+ page sites within LLM output limits by generating pages in batches of 3–5 per call, generating shared CSS once, injecting navigation with active states, and streaming files to S3 to avoid memory/Kafka-size limits (auto_store threshold pattern). Superseded by sequential per-page generation with the loop action after race conditions and quality problems.
- **sources:** docs006_workflow_builder/009_massive_multipage_sites.md#Quick-Decision-Tree; docs010_multitrack_flows_persona_architecture/019_start_here_document.md#Week-1
- **relations:** loop action; Kafka message size limits; stream_to_s3/auto_store (ancestor of storage-architecture S3 result offloading).
- **verify-later:** assemble_multipage_site action current form; auto_store config in agent chassis.

### Orchestration debug log taxonomy
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** docs006/010 raw notes listing grep targets ("DEBUGaa: What have I done with CollectedData", "The Golden Search: grep -B 5 -A 30 generate_html") plus a real database lock incident (idle-in-transaction blocking INSERT INTO sites).
- **what:** The early debugging playbook: canonical log messages for action execution flow, LLM calls, data extraction and CollectedData tracking, with kubectl grep recipes; plus pg_stat_activity lock triage and pg_terminate_backend for idle-in-transaction blockers. Ancestor of the formal debugging guides.
- **sources:** docs006_workflow_builder/010_debugging.md
- **relations:** debugging category docs 016/016b; data-path problem.
- **verify-later:** whether DEBUGaa markers remain in code.

### Pragmatic Evolution Engine (portfolio build/learn/test/optimize)
- **category:** NEW:portfolio-evolution
- **status-signal:** abandoned
- **status-evidence:** docs009/001 full 4-phase plan ("Internal Library of Effectiveness", "Controlled Evolutionary Cohorts"); no subsequent doc implements cohort testing, manifests, or the Librarian.
- **what:** The founding mission statement for a large-scale website portfolio: Phase 1 pragmatic-first MVP builds from behavioural models (AIDA/PAS) with intelligent component fallback; Phase 2 "Idea Generator" evidence gathering from winner sites (Prospector via Ahrefs-type metrics, Capture Bot producing dom+screenshot+layout_map "Rosetta Stone", Pattern Deconstructor VLM scoring components against behavioural models, Librarian producing a Hypothesis Priority List); Phase 3 large-scale single-variable A/B cohort tests turning correlation into causation, with content and layout evolved on separate tracks for SEO stability; Phase 4 site-specific optimization (winners applied only where they won — no monoculture), manifest.json component "genes" per site, git_hook_adapter flagging human-edited repos as desynchronized for HITL review, and exporter agents (WordPress XML/SQL) for client handoff.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Core-Mission; docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Phase-2; docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Phase-4
- **relations:** site interrogation/pattern library; adoption-pipeline; improvement-loop is the maintenance-shaped descendant; llm-quality-testing.
- **verify-later:** any manifest.json in site repos; git_hook_adapter; cohort/experiment tables (expected absent).

### Site interrogation & pattern library
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** docs009/003: "'Interrogate' successful sites to extract... Store extracted patterns"; docs012/012 Part 3 details the 5-phase pipeline (discover → firecrawl capture → LLM structure analysis → pattern extraction → component creation) with pattern_sources table marked "(future)".
- **what:** Learning from successful sites without copying: capture HTML+screenshot, LLM-analyse section types, visual hierarchy, content strategy and psychological principles, extract reusable patterns tagged by industry/funnel-stage/audience with "why it works" notes, and mint content_components (origin_type='extracted') from them. Patterns become queryable ("for finance trust-building, use X") and feed component selection. The most persistent unfulfilled idea of this era — restated in docs009, docs010 roadmaps (Phase 4), and docs012.
- **sources:** docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-3; docs009_site_interrogation_and_solutions/003_claude_save_point.md#2; docs010_multitrack_flows_persona_architecture/018_priority_matrix.md
- **relations:** Pragmatic Evolution Engine phase 2; adoption-pipeline site crawling (current descendant); pattern_sources/captured_sites tables; component library.
- **verify-later:** pattern_sources table; origin_type/industry_tags/funnel_stages columns on content_components; website-capture-firecrawl agent.

### data-function contract + intelligent component fallback (P1/P2/P3)
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** docs009/001: "A data-function attribute in the HTML acts as a 'shared contract'... P1 perfect match, P2 good match, P3 generic-text-block — the site always gets built"; superseded by the function/kebab-case + data-component contract (docs017/042) whose GetComponentWithFallback keeps the 3-step fallback.
- **what:** The original decoupling of structure from content: the architect assembles empty containers tagged by function (data-function="problem_statement") so the content pipeline can independently fill them; component lookup degrades gracefully (exact function → similar purpose → generic-text-block) so a build never fails for lack of a component.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Phase-1; docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md#Lookup-Safety-Net
- **relations:** component naming contract (successor); content_components.function; AssembleFromLibraryAction.
- **verify-later:** GetComponentWithFallback in component_library.go; generic-text-block component row.

### MVP build squad lineage (chief-strategist → architect → content-creator → deployer)
- **category:** NEW:site-build-pipeline
- **status-signal:** superseded
- **status-evidence:** docs009/001 defines the 4 MVP agents with Kafka payloads; docs015/004 shows the mature descendant ("pageflow-builder: ensure_site_record → call_site_planner → sync_pages → populate_nav → image generation → style collection → build_pages_loop").
- **what:** The builder pipeline's evolutionary line: mvp-site-builder (4 agents, single page) → landing-page-builder / content-site-builder (specialist architects per site type) → multipage-website-builder (batching, then sequential loop) → pageflow-builder (DB-backed pages, per-page loop with review and git commit) → site-work-orchestrator (work-item driven). Each generation kept strategist/planner, writer, assembler, deployer roles while moving state from CollectedData into the database.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-Agent-chief-strategist; docs006_workflow_builder/003_current_state_of_agents.sql#SPECIALIST-ARCHITECT-SYSTEM; docs015_data_flow_verification/004_builder_flow.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md
- **relations:** unified work items (final form); loop action; site-planner; deployment-github.
- **verify-later:** which builder agent_definitions still exist and which have traffic.

### Recursive component tree ("everything is a component")
- **category:** contracts-and-standards
- **status-signal:** abandoned
- **status-evidence:** docs009/001: "We remove is_container... If the HTML template contains {{.Slot_main}}, it IS a container"; RenderNode recursive algorithm, ghost components (wrapper_tag NULL), slot merging; the shipped system instead uses a flat section list per page with header/footer injection.
- **what:** A radically simplified component model where structure is defined entirely by template placeholders: components declare defined_slots and data_schema; the build plan is itself a component tree the architect walks recursively (RenderNode), handling any nesting depth; themes are just root components; "ghost" components (no wrapper tag) reduce div nesting. Content generation is decoupled by flattening the tree into a content_map of UUID→field requirements. The flat-sections production system never adopted the recursion, though slots re-surface in docs018's slot-based assembly proposal.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-The-Simplification; docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#2-The-Recursion-Logic; docs009_site_interrogation_and_solutions/003_claude_save_point.md#Key-Architectural-Principles
- **relations:** slot-based modular assembly (docs018/007); asset bubble-up; content injector pattern.
- **verify-later:** defined_slots column on content_components (expected absent or unused).

### Asset bubble-up deduplication
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** docs009/001 "Return Value Bubble-Up... use 100 buttons, button.css included once"; production instead uses a single global styles.css plus inline component <style> blocks.
- **what:** During recursive rendering, each component returns its HTML plus its CSS/JS dependency list; parents merge children's assets upward, and the root injects the deduplicated set once into the head. Tied to js_dependencies column proposals on content_components.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#4-Solving-Assets
- **relations:** recursive component tree; CSS responsibility barrier (what actually shipped); JavaScript management section in docs017/023.
- **verify-later:** js_dependencies column existence.

### Global context injection for navigation
- **category:** navigation
- **status-signal:** superseded
- **status-evidence:** docs009/001 "Context Propagation... any component can access {{.Global.Sitemap}}"; docs012/002 adds explicit sitemap JSON to strategist output; superseded by nav tables + GetNavItems (docs017/019b "reads nav tables directly, falls back to pages table").
- **what:** Navigation treated as data, not structure: the strategist emits the sitemap first (labels, urls, in_header/in_footer flags), and it is passed down as a Global context object so header/footer templates range over it — pages invented by the strategist automatically appear in nav. Evolution chain: Global context → sitemap in page_plan → pages-table queries (deployed-only) → site_nav_groups/site_nav_items tables.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#3-Solving-Navigation; docs012_site_maps_and_components/002_site_map_integration.md; docs018_rerendering/003_website_builder_architecture_status_report.md#5
- **relations:** nav agent family; navigation-from-pages; three-tier authority model.
- **verify-later:** GetNavItems and populate_nav_tables in component_library.go.

### Spatial addressing for natural-language editing
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** Vision in docs009/001 ("edit the paragraph on the left of the blue call to action... data-uuid and data-path attributes"); partially realized per docs018/008: "Component labeling (new) — injects data-pc-id, data-slot, data-position into each <section>".
- **what:** Every visible element carries a unique ID and genealogy path so an editing agent can resolve fuzzy human instructions spatially ("3rd paragraph on the left", "the one under the yellow button") by highlighting candidates iteratively. Shipped at section granularity (data-pc-id/data-slot/data-position mapping sections to page_components rows); element-level addressing and the conversational disambiguation loop remain unfulfilled.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-The-Spatial-Address-System; docs018_rerendering/008_granular_editing.md#What-Exists-Today; docs012_site_maps_and_components/006_start_concluding_links.md#2.4
- **relations:** section-editor agent; page_components.data_uuid/data_path columns; granular editing spectrum.
- **verify-later:** data_uuid/data_path columns on page_components; labeling injection code.

### Multi-track flows (journeys, narrative arcs, layered context)
- **category:** NEW:flows-and-narrative
- **status-signal:** abandoned
- **status-evidence:** Full schema (site_flows, flow_pages, page_transitions, site_brand_dna) in docs010/002; docs010/005 "Configuration: Single-flow (production)... build for complexity, configure for simplicity"; docs012/007 MVP migration re-lists site_flows as still-to-create; no later doc shows flows populated.
- **what:** Model a site as choreographed audience journeys rather than a flat page list: each flow has an audience segment, entry points, a narrative arc of stages with per-stage voice parameters, and ordered pages with context_overrides; context inherits hierarchically SITE (immutable brand DNA) → FLOW (narrative) → PAGE (objective/overrides) → COMPONENT (paragraph tactics); navigation becomes flow-aware (different next-step CTAs per track); shared pages get per-flow variants; page_transitions support A/B weighting. "Stop thinking pages → start thinking journeys."
- **sources:** docs009_site_interrogation_and_solutions/004_multitrack_sitemap_architecture_different_flows.md; docs010_multitrack_flows_persona_architecture/002_multi_track_schema.sql; docs010_multitrack_flows_persona_architecture/005_implementation_summary.md
- **relations:** brand DNA; voice parameters; persona assignment per stage; pattern library flow-stage tagging; conceptual ancestor of content strategy in site plans.
- **verify-later:** site_flows/flow_pages/page_transitions/site_brand_dna tables — created? populated?

### Brand DNA invariants with bounded variance
- **category:** NEW:flows-and-narrative
- **status-signal:** abandoned
- **status-evidence:** docs009/004 "brand_dna.invariants: core_message, forbidden_phrases, required_elements; variance_allowed: voice_formality [0.4,1.0]"; site_brand_dna table in docs010/002; later brand data lives in sites.content_data.brand_spec instead (docs017/019b).
- **what:** A site-level immutable identity layer — core message, values, visual system, forbidden phrases, required elements — plus explicit allowed ranges for voice variance, enforced by an evaluator check before content is accepted (vocabulary, contradiction, variance bounds, visual consistency). Solves coherence-vs-variation across multiple flows/voices.
- **sources:** docs009_site_interrogation_and_solutions/004_multitrack_sitemap_architecture_different_flows.md#Q3; docs010_multitrack_flows_persona_architecture/002_multi_track_schema.sql#BRAND-DNA
- **relations:** brand_spec in sites.content_data (descendant); content-reviewer coherence checks; design-composition brand decisions.
- **verify-later:** site_brand_dna table vs sites.content_data.brand_spec usage.

### Voice parameters (numeric stage-tuned voice)
- **category:** NEW:flows-and-narrative
- **status-signal:** superseded
- **status-evidence:** docs010/019 Week 2 plan (get_voice_for_page SQL, formality 0.5 home / 0.7 elsewhere); docs010/007: "Instead of trying to tune voice parameters numerically (formality 0.7 → 0.8), we select the right copywriter persona."
- **what:** Continuous voice dials (formality, technical_depth, sales_pressure, urgency, data_density, emotional_appeal 0–1) attached to flow stages and page context_overrides, injected into content prompts so voice progresses through the journey (awareness casual → conversion formal). Explicitly superseded within its own directory by persona selection, which embodies the parameters naturally.
- **sources:** docs010_multitrack_flows_persona_architecture/019_start_here_document.md#Week-2; docs010_multitrack_flows_persona_architecture/007_personas_discussion.md#The-Key-Insight
- **relations:** copywriter persona roster (successor); multi-track flows.
- **verify-later:** get_voice_for_page function existence.

### Copywriter persona roster
- **category:** NEW:persona-architecture
- **status-signal:** abandoned
- **status-evidence:** docs010/010 SQL seeds six personas (Elena Martinez B2B, James Chen technical, Marcus Williams conversion, Aisha Okonkwo thought-leadership, Raj Patel data, Sophie Dubois premium) with style agents; persona_assignments schema in docs010/009; no later builder references personas.
- **what:** A roster of copywriter personas — each a personality profile (biography, Big Five psychology, expertise weights, voice traits) with attached specialized style agents — assigned to flow stages or content types ("assign Marcus to all conversion pages") via personas / specialized_agents / persona_assignments tables and get_persona_for_page lookup (page → stage → default). Voice emerges from persona choice rather than parameter tuning; maps to real agency roles.
- **sources:** docs010_multitrack_flows_persona_architecture/008_example_personas.md; docs010_multitrack_flows_persona_architecture/009_persona_system_schema.sql; docs010_multitrack_flows_persona_architecture/007_personas_discussion.md
- **relations:** persona cognitive architecture; multi-track flows; page-content-writer.
- **verify-later:** personas/specialized_agents/persona_assignments tables in clients_db.

### Persona cognitive architecture (swappable cognitive components)
- **category:** NEW:persona-architecture
- **status-signal:** abandoned
- **status-evidence:** docs010/015 "The architecture is ready for the full vision while starting simple today"; full schema + 8 cognitive actions + Dr Bimpton example SQL delivered; nothing downstream implements the Go actions.
- **what:** Personas as complete cognitive entities: immutable personality DNA plus swappable subsystems (perception, working/episodic/semantic memory, knowledge retrieval, reasoning engine, response generator, style applicator, learning system), each with pluggable implementations evolving Phase 1 all-LLM → vector-DB memory → fine-tuned persona models → multi-model per task → custom reasoning services, switchable via is_default without workflow changes. Running instances persist memory and emotional state; persona_knowledge holds facts/beliefs/opinions with confidence and future embeddings; task executions log full cognitive traces. Eight-step cognitive workflow per task (initialize→perceive→retrieve→reason→generate→style→learn→complete).
- **sources:** docs010_multitrack_flows_persona_architecture/015_persona_README_architecture.md; docs010_multitrack_flows_persona_architecture/011_persona_cognitive_architecture.sql; docs010_multitrack_flows_persona_architecture/014_drBimpton_setup_example.sql
- **relations:** finetuning-flywheel (fine-tuned persona models); reasoning; entity_state_log (parallel memory design); copywriter roster.
- **verify-later:** personas/persona_cognitive_components/persona_instances/persona_knowledge/persona_task_executions tables; load_cognitive_system etc. in action registry (expected absent).

### Loop action via dynamic workflow expansion
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** docs010/022 integration guide with concrete file placements; docs010/023 walkthrough ("transforms the loop into individual workflow steps... Async-compatible... Recoverable... resume from exact step"); every later builder uses build_pages_loop.
- **what:** The `loop` action doesn't execute iterations itself — a coordinator-side expansion handler injects one workflow step per iteration×substep (generate_pages_loop_iter_0_research …), chained by NextStep, with loop_metadata in CollectedData, setLoopVariable placing the current item under the loop_var before each step, and a loop_complete step aggregating results into output_field. Design chosen (over in-process execution) because steps can await async agent responses and survive crashes/restarts as ordinary persisted workflow steps.
- **sources:** docs010_multitrack_flows_persona_architecture/021_loop_action_discussion.md; docs010_multitrack_flows_persona_architecture/022_loop_actions_guide.md; docs010_multitrack_flows_persona_architecture/023_loop_explanation.md
- **relations:** sequential page generation; work-item loops; orchestration state persistence.
- **verify-later:** loop_action.go, loop_expansion_handler.go, loop_complete_action.go in platform.

### Sequential page generation (Phase 0 multipage fix)
- **category:** NEW:site-build-pipeline
- **status-signal:** superseded
- **status-evidence:** docs010/019 "Current Problem: Multipage-website-builder tries to generate 4 pages at once — race conditions... Solution: make it work like landing-page-builder: sequential, one page at a time"; two-delay spawn fix noted FIXED in docs009/003.
- **what:** Replacing parallel batch spawning with a strategist-planned page list iterated by the loop action (research → write per page), a wrap_multipage action generating navigation and collecting assets, and spawn timing fixed by double initialization delays. The stabilization step that made multipage builds reliable enough for everything after.
- **sources:** docs010_multitrack_flows_persona_architecture/019_start_here_document.md; docs009_site_interrogation_and_solutions/003_claude_save_point.md#Status
- **relations:** loop action; batched generation (predecessor); pageflow-builder (successor).
- **verify-later:** wrap_multipage action; spawn_actions.go delay logic.

### Assembly action consolidation (3 clear actions)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** docs010/020: "You have [6 overlapping assembly actions]... Too much overlap. Proposed: 3 clear actions (assemble_page ...)"; later flows use assemble_page (docs011/003, docs015).
- **what:** Rationalizing the accumulated assembly actions (assemble_from_library, assemble_full_page, AssembleHTMLParts, AssembleMultipageSite, WrapMultipage, html_actions) into a minimal set: assemble_page (one page from structure+styles+content), plus multipage assembly and library assembly with shared code. A recurring theme: action proliferation followed by consolidation.
- **sources:** docs010_multitrack_flows_persona_architecture/020_revised_consolidated_action_plan.md; docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md
- **relations:** component library unification; slot-based assembly proposal.
- **verify-later:** current action registry entries for assembly actions.

### Component library unification (component_library.go)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs012/009: "one source of truth for all component operations... RenderTemplate handles both Go-style {{.field}} and Handlebars-style {{field}}, {{#each}}, {{#if}}"; later docs (015, 017, 018) treat component_library.go functions as load-bearing infrastructure.
- **what:** A shared Go module consolidating duplicated component code: component queries (by function, by ID, with fallback), style collection resolution (per-site with domain-keyword fallback), theme loading, dual-syntax template rendering, and high-level RenderHeader/RenderFooter/InjectHeader/InjectFooter/InjectHead used by both full-page assembly (assemble_from_library) and header/footer injection into LLM-generated pages (multipage path).
- **sources:** docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md#Summary; docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md
- **relations:** style collections; InjectHead bug; GetNavItems; rerender pipeline.
- **verify-later:** platform/orchestration/actions/component_library.go current contents.

### Style collections as the design bridge
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** docs017/017 three-layer model (HTML components / CSS theme / style collection "the bridge"); docs012/009 migration 030_style_collections_migration.sql; per docs015/004 "load style collections" is a standard planner step.
- **what:** Layer 3 of the design system: a style_collection binds a site to specific header/footer/head component choices plus a CSS theme (colors, typography), selected per site (stored on sites, or chosen by domain keywords as fallback). Enables mix-and-match of structure and appearance and consistent chrome across the multipage path. Ancestor of the current palette/typography/layout resolution system.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/017_agent_architecture_v2.md; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Design-System-Layers; docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md
- **relations:** css_themes; webdesign-agent; design agent family split; current design-composition docs 025-027.
- **verify-later:** style_collections table shape and GetStyleCollectionForSite.

### Semantic linking domain decomposition (5 link types)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** docs012/003 taxonomy table ("Links are not one thing — at least 5 different things") and proposed link-management-group of six agents; docs012/006 concludes "Links live in components, registry is an index"; lifecycle and semantic agents remain unbuilt.
- **what:** Recognition that link work spans navigation (low complexity), content links/CTAs, semantic links (pillar↔cluster topic modelling — AI-heavy), cross-site/network/affiliate links, and technical links (sitemap/canonical/hreflang), each needing different mechanisms and lifecycles (news decays in days, campaign pages expire, products die). Proposed agent group: navigation-agent, seo-agent, lifecycle-agent, cross-site-agent, semantic-link-agent, link-validator.
- **sources:** docs012_site_maps_and_components/003_semantic_linking.md; docs012_site_maps_and_components/004_more_on_links.md; docs012_site_maps_and_components/006_start_concluding_links.md
- **relations:** link_registry; relationships table for semantic pairs; links agent family (docs017/019b, algorithmic-only subset); current link-management docs 024.
- **verify-later:** which of the six proposed agents exist; page relationships in relationships table.

### link_registry as derived index (links live in components)
- **category:** link-management
- **status-signal:** deployed
- **status-evidence:** docs012/006 schema with scope/link_type/affiliate fields; docs012/012 pipeline step "5e. EXTRACT LINKS — Action: extract_and_sync_links; DB Write: link_registry".
- **what:** Links are never stored as primary data — they exist inside rendered components; link_registry is a queryable index derived by extraction after rendering, tracking source component/page/site, resolved internal targets, scope (internal/page/site/network/external), type (navigation/content/semantic/affiliate/reference), anchor text, rel attributes, affiliate provider/tag, and validation health. Enables broken-link detection, orphan detection, and affiliate compliance without duplicating truth.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.5; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-2; docs012_site_maps_and_components/007_link_migration.sql
- **relations:** links agent family heartbeat; validate_page_content; redirect-manager.
- **verify-later:** link_registry table + extract_and_sync_links action.

### Clients → networks → sites hierarchy
- **category:** database-and-infrastructure
- **status-signal:** partial
- **status-evidence:** docs012/006 and /007 CREATE TABLE clients/networks/sites "designed for 1000s of sites, 10000s+ pages... networks of sites"; sites is heavily used later, networks/clients rarely referenced again.
- **what:** The multi-tenancy spine: clients (linked to auth-service external_id) own networks (with network-wide settings such as affiliate config), networks own sites (domain, brand_dna, github repo/branch, settings, build/deploy timestamps). Motivated by cross-site linking within a client's networks and component-level bulk updates across many pages.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.1; docs012_site_maps_and_components/004_more_on_links.md#Part-1
- **relations:** cross-site link scope; multicluster scaling; client schemas in database-and-infrastructure.
- **verify-later:** networks/clients tables — created and populated?

### page_components — component instances as the page's stored form
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Schema introduced docs012/006 ("the bridge between content_components (templates) and actual page content"); docs018/010 treats it as established core ("Each section on a page maps to a page_components row").
- **what:** Every section of every page is a row: template reference (component_id), position/slot_name, nesting支持 (parent_component_instance_id), the rendered_html actually deployed, the content_data that produced it, content_hash for change detection, and semantic addressing fields (data_path, data_uuid). This is the storage foundation that makes rerendering, section editing, locking, and maintenance possible — the single most consequential schema decision of this era.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.4; docs018_rerendering/010_section_editor_architecture.md#Component-Architecture; docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md#The-Data-Flow
- **relations:** content_data source-of-truth principle; component naming contract (slot_name = function); rerender; locks (asset locking mirrors page_components).
- **verify-later:** page_components current columns incl. schema_snapshot/content_snapshot/build_status.

### render_mode decision matrix (DB template vs agent vs research)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** docs012/012 decision matrix ("Headers/Footers: template, never generated; FAQ: agent + research; Pricing/Contact: template + brief data") and "The render_mode field on content_components is the key differentiator".
- **what:** Each component declares how its content is produced: render_mode='template' renders directly from brief/render_context data (LLM only fills missing schema fields); render_mode='agent' spawns LLM generation, optionally preceded by the research-agent when needs_research=true. Pure-structure components never touch an LLM; research-backed components always cite. Governs the per-section branch inside the page build loop.
- **sources:** docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-1; docs012_site_maps_and_components/010_component_and_site_architecture.md
- **relations:** research agent; page-content-writer section loop; "LLM = Agent" principle (every LLM call gets its own agent with research/draft/review).
- **verify-later:** render_mode/needs_research/agent_type/data_sources columns on content_components.

### Agent input/output contracts (expects/required/produces)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** docs012/012 Part 4 contract JSON for site-planner/page-content-writer/research-agent; "input_contract/output_contract fields in agent_definitions" cited as the tracking mechanism; docs017/002_standardising exports them for validation.
- **what:** Formal per-agent declarations of expected input fields (with required subset) and produced output shapes, stored on agent_definitions, intended to make cross-agent data flow checkable (workflow validator) and self-documenting. Partially realized; the enforced end of it became ActionInputSpec at the action level.
- **sources:** docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-4; docs017_legacy_agent_rules_images_design_keydocs/002_standardising_deployment_implementation_plan.md
- **relations:** ActionInputSpec; workflow builder/validator; contracts-and-standards doc 003 (current descendant).
- **verify-later:** input_contract/output_contract columns populated?

### Flexible vs strict schema mode (approval snapshot)
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** docs017/001_flexible proposes sites.schema_mode + page_components.schema_snapshot/content_snapshot/component_version_id; docs018/008 lists "page_components.content_snapshot stores approved values... schema_mode can be strict or flexible per section" under "What Exists Today".
- **what:** Two-phase content lifecycle: initial builds run flexible (best-effort substitution, warnings on missing fields, creative freedom); at approval the section's input_schema and content values are snapshotted and the section moves to strict mode (edits validated against locked schema, unsubstituted placeholders fail, template upgrades can't break approved pages, rollback via content_snapshot). Open questions recorded: granularity, versioning vs snapshot, transition trigger, unlock capability.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/001_flexible_schema_enforcement.md; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-5; docs018_rerendering/008_granular_editing.md
- **relations:** locks (current descendant of the freeze idea); content-reviewer approval; section editor.
- **verify-later:** schema_mode/schema_snapshot/content_snapshot columns and whether any transition code exists.

### Research agent with cited sources
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** docs015/001 step-by-step verified pipeline (extract_topic → build_search_query → web_search → prepare_urls → batch_webscrape → format_research_content → synthesize → insert_research_result); docs012/010 principle "Research is cited — all LLM-generated content must cite sources, which are stored."
- **what:** A self-contained research agent: composes a search query from raw inputs, searches, selects top URLs, batch-scrapes them via the webscrape adapter, formats findings with snippet context, synthesizes a JSON summary (key points, recommendations, confidence), and persists to research_results with full source list — returning a research_id that content sections reference. Backing store for research-driven components (FAQ, long-form).
- **sources:** docs015_data_flow_verification/001_data_flow_verification.md; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#research-agent; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Active-Agents
- **relations:** render_mode needs_research; batch_webscrape action; adapters (webscrape); current research-agents category.
- **verify-later:** research_results table; prepare_urls/batch_webscrape/format_research_content in registry.

### Asset & product provenance tables
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** docs012/010 new-table list: "assets — track all images/videos with full provenance; products; product_assets; affiliate_programs; affiliate_products"; docs017/002_pageflow stores hero assets in "assets table (existing)".
- **what:** All media (generated, uploaded, scraped) tracked with provenance in an assets table; product catalog and affiliate product caching designed alongside for e-commerce/review sites. The assets side shipped (used by image generation flow); products/affiliate tables remained design.
- **sources:** docs012_site_maps_and_components/010_component_and_site_architecture.md#New-Tables; docs017_legacy_agent_rules_images_design_keydocs/002_pageflow_image_changes.md
- **relations:** image generation pipeline; entity-data (products as entities superseded product tables); link_registry affiliate fields.
- **verify-later:** assets table columns; products/affiliate_programs existence.

### Data-flow verification matrix practice
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** docs015/001 is a complete per-step verification table ("Config | Value | Verified ✓") including output structures and registration checklist; docs017/044 repeats the practice for the site-work-orchestrator.
- **what:** A documentation/QA practice: before deploying a workflow, trace every step's config paths against the action implementations — where each output lands in collected_data, its structure, and each input's exact path — plus response-header compliance and action-registration checklists. The manual ancestor of automated contract validation.
- **sources:** docs015_data_flow_verification/001_data_flow_verification.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md
- **relations:** input contracts; workflow validator; ActionInputSpec.
- **verify-later:** n/a (practice, not code).

### Head-inside-body bug and positional injection fixes
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs015/002 trace: "cleanHTMLStructure keeps the LARGER <head> — wrong heuristic... InjectHead does in-place replacement — preserves wrong position" with concrete fixes and deployment order ending "re-run rerender_pages".
- **what:** Two compounding rendering bugs: LLM sections sometimes emit full HTML documents, and the dedup heuristic kept the larger (misplaced) head while in-place head replacement preserved the wrong position. Fixes: remove all head blocks then always insert before <body>; dedup by position (remove heads after <body>) not size. Exemplifies the fragility of regex injection that motivated slot-based assembly.
- **sources:** docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md#Bug-1
- **relations:** slot-based assembly proposal; component_library.go InjectHead; rerender pipeline.
- **verify-later:** current InjectHead/cleanHTMLStructure implementations.

### Colour inheritance model + CSS responsibility barrier
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs015/002 Bug 2 ("body sets color — the ONLY default text colour; h1-h6 color: inherit; dark sections set color:#fff on container"); docs018/003: "Global CSS: all colors/fonts; Component CSS (inline): layout, positioning, structure only."
- **what:** The design-system rule set that fixed light-text-on-light-background failures: exactly one place sets default text colour (body); headings and text elements inherit; components never force colours or backgrounds on text elements; dark sections override at container level so children inherit white. Paired with the responsibility barrier: global styles.css owns colour/typography, component inline CSS owns layout only. Enforced through the webdesign-agent CSS prompt.
- **sources:** docs015_data_flow_verification/002_temp_doc_flow_of_html_and_css_creation.md#Bug-2; docs018_rerendering/003_website_builder_architecture_status_report.md#1; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#The-Colour-Inheritance-Model
- **relations:** dark-section --section-* contract (refinement); section-contrast model (current descendant); webdesign-agent.
- **verify-later:** current webdesign/CSS prompts; styles.css conventions.

### Dark-section context variable contract (--section-*)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** docs017/043 lists 10 dark components migrated with is_dark_section=true and --section-* vars; four enforcement layers specified with deployment order ("Run 014_section_context_migration.sql...").
- **what:** Any dark-background component must set --section-text/-text-muted/-heading/-surface/-border custom properties on its container; global CSS reads them with light-theme fallbacks. Enforced in depth: DB flag (is_dark_section) + audit queries, Go warnings in RenderComponentAction/SavePageSectionsAction, LLM prompt rules in webdesign-agent and page-content-writer, and periodic SQL audits. Direct ancestor of the current section-contrast model.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/043_section_naming_contract.md
- **relations:** colour inheritance model; component naming contract (companion doc); maintenance audits.
- **verify-later:** is_dark_section column; validate_dark_section.go; current section-contrast implementation.

### Component naming contract (function = canonical kebab-case ID)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** docs017/042: "content_components.function is the canonical identifier... DB constraint chk_function_kebab_case... partial unique index on active function"; migration table of renames (social_proof → social-proof, 5 heroes disambiguated).
- **what:** One rule ending a class of chain-breaking bugs: `function` (kebab-case, regex-constrained, unique among active components) identifies a component everywhere — the template's data-component attribute must equal it, page_components.slot_name stores it, planners assign by it, rerenders match by it. NormalizeComponentFunction + 3-step fallback tolerate legacy data; adoption pipelines must translate external names, never import them.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md
- **relations:** data-function contract (ancestor); page_components; adoption pipeline mapping; SavePageSections/page-rerender.
- **verify-later:** chk_function_kebab_case constraint and unique index in DB; component_validation.go.

### Specialist agent design doctrine (agents own their domain)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Five versions of the checklist culminating in docs017/008; audited against real code in docs017/044 ("Audit Notes vs 008_checklist..."), including accepted divergences.
- **what:** The core agent-design rulebook: agents are self-contained and independently callable, with dedicated load_* actions gathering their own data; callers pass raw domain identifiers, never derived values ("if changing the child requires updating the caller, you've leaked responsibility"); reuse/patch existing actions before creating new ones; workflows stay declarative (templates/config = intent OK; loops/branching = Go); orchestrator vs agent boundary (what/order vs how); standalone + integrated dual modes; spawn before call; agents reply to the caller's topic; no container config in definitions. v2's interim "use input_fields not explicit paths" rule was replaced by the ActionInputSpec regime in v3.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md; docs017_legacy_agent_rules_images_design_keydocs/007_checklist_for_new_specialist_agent_v4.md#Orchestrator-Boundaries; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md#Audit-Notes
- **relations:** ActionInputSpec; webdesign-agent (first exemplar, docs017/003); current development-guide doc 001.
- **verify-later:** how closely current agents follow load_* pattern.

### ActionInputSpec / ExtractActionInputs standardized extraction
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** docs017/008 full spec ("No more boilerplate — 40+ lines of extraction code per action eliminated"; deprecation warnings for *_field patterns); real bug documented: "the site plan contamination bug — ExtractActionInputs found site_record.content_data via nested lookup... overwriting the hero section with the site plan."
- **what:** Every action declares an ActionInputSpec (Required/Optional/Defaults/Deprecated) and calls one extraction function that tries input_fields, falls back to deprecated *_field keys with warnings, checks nested parents (current_page/site_record/input_data/rerender_pages), validates and defaults. Includes the hazard doctrine — never name optional fields after common nested keys (content_data, domain, status...), prefix when in doubt — and the `?` suffix for optional input_mapping fields (skip silently if source path missing) supporting multi-mode agents. Literal config values must be read directly from config, not through path resolution.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md#Decision-Standardized-Input-Extraction; docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md#Avoid-Field-Names; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md#Fixed-during-audit
- **relations:** data-path problem (root cause); input contracts; workflow validation.
- **verify-later:** datahelpers.ActionInputSpec/ExtractActionInputs/RegisterActionInputSpec in platform.

### Agent families architecture (nav/links/design/content/entity/tools/feed/maintenance)
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** docs017/019b (v5) with per-family status columns ("populate_nav_tables — Deployed"; layout-architect — New; brand-designer — Future split) and a Data Ownership Summary table mapping every table to an owner agent; phased plan 1→4.
- **what:** The master blueprint of the specialist-agent era: eight agent families each owning a data domain — navigation (nav tables), links (algorithmic health), design (brand/layout/CSS split), content (marketing/legal/SEO/product writers + reviewer + researcher), entity data, tool builder tiers, news/content feed, and maintenance — with explicit "does NOT do" boundaries, a component-builder-v2 workflow sketch, site-type stress tests (brochure/e-commerce/finance/events/platform), and single-owner-per-table data governance. Much became real (nav actions, webdesign, feeds, maintenance→work items); some never did (layout-architect, nav-layout-agent, product-content-writer).
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md; docs017_legacy_agent_rules_images_design_keydocs/002_full_new_agent_architecture.md; docs017_legacy_agent_rules_images_design_keydocs/018_agent_architecture_v3.md
- **relations:** nearly every other concept in this unit; data ownership prefigures council-agent domain ownership.
- **verify-later:** which family agents exist in agent_definitions today.

### Navigation agent family + three-tier authority
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** docs017/019b: "core responsibilities are implemented as the populate_nav_tables action... full standalone nav-agent is planned but not yet needed"; utility classification list and nav data flow marked (implemented).
- **what:** Navigation as a first-class entity: site_nav_groups/site_nav_items with typed groups (primary, subsection, content, legal, utility, external, contextual); populate_nav_tables classifies pages (FAQ/Blog/Careers etc. routed to utility even if in_header); GetNavItems serves header (primary, deployed-only) and footer (primary+utility+legal) rendering with pages-table fallback. Authority tiers: strategist owns structure at build; nav agent makes incremental decisions in maintenance; periodic drift detection compares current nav against the original plan ("drift may represent valid evolution").
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#1-Navigation-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Three-Tier-Authority-Model
- **relations:** navigation-from-pages (predecessor); nav-updater fix agent; current navigation FOCUS docs.
- **verify-later:** site_nav_groups/site_nav_items tables; populate_nav_tables action; standalone nav-agent existence.

### Links agent family (algorithmic, no-LLM link health)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** docs017/019b family table (link-crawler, link-validator, link-registry-sync, redirect-manager, affiliate-link-manager phase 2 — all "LLM? No") with heartbeat workflow and explicit non-goals.
- **what:** Deliberately judgment-free link maintenance: crawl modified pages' HTML, classify by URL pattern, resolve internals to page records, HEAD-check externals rate-limited, detect broken links and orphan pages, generate redirects on URL changes, track per-page link counts and empty anchors. Explicitly excluded: link placement, nav decisions, SEO strategy, related-content suggestions (LLM territory).
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#2-Links-Agent-Family
- **relations:** link_registry; semantic linking decomposition (the LLM parts deferred); redirect-manager fix agent.
- **verify-later:** links-orchestrator agent; site_redirects table.

### Design agent family split (brand-designer / style-generator / layout-architect)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** docs017/019b: webdesign-agent "Exists, prompt updated for colour inheritance"; brand-designer/style-generator "Future split"; layout-architect/nav-layout-agent "New" (never appear later); "There's no rush on this split."
- **what:** Decompose the monolithic webdesign-agent (analyse_design → generate_css → update_site, deploying /assets/css/styles.css) into: brand-designer producing a rarely-changing brand_spec (palette, type scale, spacing, tone, image direction) in sites.content_data; style-generator producing CSS with theme-library search-and-adapt before generating fresh (feeding css_themes for reuse); layout-architect producing per-page-type layout definitions (nav placement, content zones, max components) with rendering fallbacks. Direct ancestor of the current site-design-planner / design composition system.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#3-Design-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/003_design.md
- **relations:** style collections; colour inheritance; current design-composition docs 025-027 (successor).
- **verify-later:** brand_spec/layout_definitions keys in sites.content_data; whether split agents exist.

### site-scraper companion agent (design context from live URLs)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** docs017/003: site-scraper "firecrawl_scrape → analyze_design → returns site_context for webdesign-agent"; standardized site_context schema with source: "database|scrape|manual".
- **what:** A standardized site_context interface (domain, company, industry, palette, typography, component functions, source) produced by either DB load, live-site scraping, or manual input, so the webdesign-agent can restyle from any source — enabling "scrape competitor → feed to design agent → apply to your site" pipelines. The schema-standardization idea matured; the scraper flow folded into the adoption pipeline.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/003_design.md#Architecture; docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md#Standardized-Interface-Schemas
- **relations:** adoption-pipeline capture; webdesign-agent; standardized interface schemas doctrine.
- **verify-later:** site-scraper agent definition; load_site_for_design action.

### Entity data agent family (structured data drives pages)
- **category:** NEW:entity-data
- **status-signal:** partial
- **status-evidence:** docs017/019b: site_entities/site_entity_relationships "(exist)", entity_sources/entity_sync_log "(planned)"; "First implementation target: boxing ticket/events site, then football tickets, then finance"; no later doc confirms the sync pipeline ran.
- **what:** Real-world entities (events, performers, venues, ticket tiers, products, articles) stored as typed JSONB rows with relationships, synced from configured sources (API/scrape/feed with field_mapping, poll intervals, rate limits), change-logged, and driving template-rendered pages with minimal LLM. Entity lifecycle is state-based, not time-based (announced → on_sale → selling_fast → sold_out → event_day → past → historical/cancelled) with per-state page and nav behaviour; status transitions auto-detected from source data. entity_sources.news_triggers defines which changes are newsworthy, bridging to the feed pipeline. Three of four stress-tested site types need it.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#5-Entity-Data-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Entity-Lifecycle; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Entity-Data
- **relations:** news feed pipeline (entity_event triggers); tickets vertical; products tables (superseded by entities); dogs-medicine entities unrelated.
- **verify-later:** site_entities/site_entity_relationships rows; entity_sources/entity_sync_log existence; entity-data-agent definition.

### Events/tickets vertical (boxing first target)
- **category:** NEW:entity-data
- **status-signal:** abandoned
- **status-evidence:** docs017/019b "Events / Tickets Site (first target — boxing, then football)... API sources: Ticketmaster, SeatGeek, BoxRec"; entity examples "Fury vs Joshua"; no boxing/tickets site appears in later portfolio lists.
- **what:** The planned first entity-driven site type: dense entity relationships (event↔performer↔venue↔ticket_tier), frequently-updating ticket tier data (price/availability) flowing to pages quickly, state-transition-driven news (fight announced, on sale, sold out, results), contextual per-event/per-performer navigation, and past events retained as permanent SEO assets.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Entity-Types-for-Events-Tickets; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Site-Type-Stress-Tests
- **relations:** entity data family; news feed pipeline.
- **verify-later:** any boxing/tickets site records.

### News & content feed pipeline (mid-era design)
- **category:** news-feed-pipeline
- **status-signal:** superseded
- **status-evidence:** v1 in docs017/030, refined in 019b (sub-agents feed-ingester/deduplicator/triage/article-rewriter/publisher/lifecycle), restructured in 023 ("Feed items go through ingestion, filtering, deduplication and relevance scoring before they become publishable" with work_item linkage); today's news-feed-pipeline (docs024 006) is the deployed descendant.
- **what:** Per-site content sources (RSS/API/scrape/entity_event) polled on configurable intervals → raw content_feed_items → dedup (near-duplicate headline detection) → LLM triage (relevance, urgency, angle for THIS site) → article-rewriter producing original articles in site voice with entity cross-links and required disclaimers → publication as pages → time-based lifecycle decay (featured 0-24h → current → aging → archive → prune, with per-site-type pacing and event-calendar coupling). Later revision: publishable items become site_work_items (handler article-writer) and rewritten articles become entities; news display is a design concern owned by the component/theme system, not the feed.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/030_news_feeds_v1.md; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#7-News-and-Content-Feed-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Content-Feed-Items
- **relations:** entity news_triggers; work items; current news-feed-pipeline (successor, incl. diversity concerns).
- **verify-later:** content_sources/content_feed_items current schema vs these designs.

### Tool builder tiers (static / dynamic / application)
- **category:** dynamic-applications
- **status-signal:** partial
- **status-evidence:** docs017/019b tier table ("Static: component library; Dynamic: self-contained JS, LLM-generated or pre-built; Application: engineer-built only") and platform stress test ("Agent-as-API pattern. User IS the HITL"); mortgagecalculator + website-design.com cited as early instances in docs018/008b.
- **what:** Interactive functionality classified by creation risk: static HTML components from the library; dynamic self-contained JS applications (calculators, visualisations) that LLMs may generate; full applications with API integration reserved for engineers. The agent-as-API pattern for platform sites treats the end user as the HITL. Matured into the tool-pipeline/tool-library/tool-lifecycle systems.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#6-Tool-Builder-Agent; docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** tool-pipeline (successor); JavaScript management (docs017/023); finance/tools stress test.
- **verify-later:** current tool generation pipeline lineage.

### Legal content agent + legal constraint rules
- **category:** content-governance
- **status-signal:** aspirational
- **status-evidence:** docs017/019b "legal-content-agent | Template + minimal LLM | New" with legal_rules JSON (required_disclaimers by trigger, forbidden_phrases, required_pages by jurisdiction template); compliance-discovery-agent phased for maintenance.
- **what:** Jurisdiction-aware legal pages from vetted templates (privacy-gdpr-uk etc.), plus machine-readable constraints exported to the content writer: disclaimers triggered by content conditions with placement rules, forbidden phrases per industry ("guaranteed returns"), required pages routed to the legal nav group. Compliance discovery monitors regulatory changes in maintenance.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Legal-Content-Agent; docs017_legacy_agent_rules_images_design_keydocs/022_maintenance_architecture_plan_v2.md#Discovery-Agents
- **relations:** policy-as-filters (org framework ancestor); finance site stress test; brand DNA forbidden phrases.
- **verify-later:** legal-content-agent definition; legal_rules key in sites.content_data.

### SEO content agent
- **category:** NEW:seo
- **status-signal:** aspirational
- **status-evidence:** docs017/019b "seo-content-agent | LLM for generation, algorithmic for validation | New — runs after page content is written"; seo-discovery-agent in maintenance Phase 0; slot exists in component-builder-v2 sketch.
- **what:** A post-content sweep owning meta titles/descriptions, structured data/JSON-LD, robots directives, canonical URLs and Open Graph across all pages, with algorithmic validation and LLM generation; complemented in maintenance by sitemap-sync, schema validation, and meta-freshness discovery plus sitemap-regenerator and schema-fixer fix agents. No dedicated SEO category exists in the current taxonomy despite recurring SEO responsibilities across eras.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#SEO-Content-Agent; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Fix-Agents
- **relations:** meta-manager (docs018/007); link technical types; site-finalizer sitemap generation.
- **verify-later:** any seo agent definitions; sitemap.xml generation code path.

### Heartbeat maintenance model (findings-based, pre-work-items)
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** v1 (docs017/021) findings table + domain orchestrators; v2 (docs017/022) full spawn chain modeled on "the vet-batch pattern" with budget management; explicitly replaced by 023: "maintenance-triage + page-rebuild → maintenance-batch-scheduler + site-work-orchestrator".
- **what:** The first full maintenance architecture: K8s CronJob (8h) → agent-chassis spawns maintenance-batch-scheduler → claims batch (FOR UPDATE SKIP LOCKED, batch_size controls concurrency) → per-site site-maintenance-orchestrator runs fix-pending → verify-previous → discover-due → triage cycle; discovery agents per domain (content/links/seo/compliance/structural) write maintenance_findings; triage (a step, not an agent) enriches with impact reads and classifies resolution path (auto_fix/suggest/flag/monitor/ignore); narrow fix agents resolve; cross-domain coordination only via side-effect findings with parent_finding_id — "no agent calls another agent for coordination." Daily maintenance-catch-all handles stale findings, HITL reminders, cross-site patterns, stuck recovery.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/022_maintenance_architecture_plan_v2.md; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#8-Maintenance-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/021_maintenance_architecture_plan_v1.md
- **relations:** vet-batch-processor precedent (vet-med-pricing); unified work items (successor); scheduler-and-tasks; maintenance profile.
- **verify-later:** maintenance_findings/maintenance_tasks tables; maintenance-batch-scheduler agent history.

### Unified build & maintenance work items (site_work_items)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** docs017/023: "Build and maintenance are the same process. A new site is a set of findings that need fixing"; full site_work_items DDL with item_key dedup, depends_on, parent_item_id, batch claiming; docs017/044 traces the working site-work-orchestrator step by step.
- **what:** The pivotal unification: every piece of work — building a page, fixing stale content, adding a tool, publishing an article — is one work item with source (planner/discovery/content_feed/manual/improvement/side_effect), domain, item_type, severity, spec JSONB, triage enrichment (impact, resolution_path, priority, handler_agent), lifecycle statuses detected→triaged→approved→claimed→in_progress→complete→pending_verify→verified, dependencies, dedup keys, attempt limits, and archival. The planner becomes a discovery agent writing 'needs_content_page' items; the same orchestrator/fix agents process build and maintenance; sites start minimal and improve incrementally, always left in a working state; per-page git commits. Old and new systems coexist (v2 intake routes between pageflow-builder and site-work-orchestrator). This is the direct ancestor of the current work-item lifecycle and improvement loop.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md; docs017_legacy_agent_rules_images_design_keydocs/041_page_rebuild_action.md
- **relations:** heartbeat model (predecessor); maintenance_queue/page-rebuild (earlier still); work-item lifecycle in development-guide (current form); news feed → work items.
- **verify-later:** site_work_items vs current work_items table naming/shape; site-work-orchestrator vs current orchestrators.

### Per-site maintenance profile with budgets
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** docs017/019b maintenance_profile JSON in sites.settings ("content every 7d, links every 8h... budget: llm_calls_per_cycle: 20, max_auto_fixes_per_cycle: 5"); 023 extends with content_feed cadence and time_sensitivity.
- **what:** Each site declares which maintenance domains run, at what cadence, with which sub-agents and regulatory bodies, plus hard budgets on LLM calls and auto-fixes per cycle — a finance site gets hourly high-sensitivity feeds and FCA compliance, a brochure site gets links+freshness only. Ancestor of the growth budget concept.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Per-Site-Configuration; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Per-Site-Configuration
- **relations:** content-governance growth budget (descendant); scheduler cadence.
- **verify-later:** maintenance_profile key in sites.settings rows.

### Image generation in the build pipeline
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** docs017/001_changes + 002_pageflow give exact patches ("generate_hero_image → store_hero_asset → deploy_hero_image (NEW) → templates use {{.hero_url}} → /assets/images/hero.jpg"); the current imagery system (site_plan_imagery, kind enums) replaced this.
- **what:** First-generation site imagery: image-generator agent produces logo/hero via adapter; store_generated_image/StoreAssetAction persists S3 URI into assets table and sites.content_data by purpose; deploy_image_asset downloads from S3, optimizes for web (resize per purpose), base64-commits via git-adapter; hero/logo URLs flow through render context into templates as background images.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/002_pageflow_image_changes.md; docs017_legacy_agent_rules_images_design_keydocs/001_changes_needed.md
- **relations:** assets table; imagery pipeline (successor); image-optimiser fix agent.
- **verify-later:** deploy_image_asset action; storage/image_processing.go; current imagery pipeline contrast.

### Standardized per-page git deployment
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** docs017/002_standardising problem table ("Inconsistent deploy patterns → Standardize to per-page commits"); docs017/023 "Individual git commits per page → each goes live via GitHub Action → S3"; work items store commit_sha in result.
- **what:** Deployment converges on one mechanism: each page is committed individually via git_commit (with file_path override enabling CSS/asset commits), GitHub Actions deploy to hosting (Cloudflare, later S3), and commit SHAs are recorded on pages and work items for traceability. Removed redundant deployer steps whose data paths kept breaking.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/002_standardising_deployment_implementation_plan.md; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Git-Commit-Strategy
- **relations:** git-adapter; deployment-github category; per-page loop.
- **verify-later:** git_commit action file_path config; GitHub Action workflows in site repos.

### Rerender pipeline (reassemble without regenerating)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** docs018/001 rerender_site_pages action doc; docs018/006 dual paths ("rerender-site: ensure_site_record → render_site_components [force] → loop(call page-rerender) → trigger_deploy"); rerender is a pillar of the current improvement loop.
- **what:** Re-assemble deployed pages from stored page_components.rendered_html with current site-level components (head/header/footer, CSS links, nav) without touching content: strip old wrappers, apply current chrome, commit. Split into page-rerender (single page) and rerender-pages orchestrator (batch), used after component/theme/nav changes. Includes contact-info injection from DB during rerender to overwrite hallucinated details.
- **sources:** docs018_rerendering/001_rerender_pages_summary.md; docs018_rerendering/006_build_path_rerender_path.md; docs018_rerendering/003_website_builder_architecture_status_report.md#2
- **relations:** page_components storage; improvement-loop rerender stage; section editor assemblePage reuse.
- **verify-later:** rerender_single_page_action.go, rerender-pages agent, trigger_deploy.

### Content validation before review (validate_page_content)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** docs018/003: "validate_page_content action runs before review mode determination... Validation errors force HITL review, blocking auto-approval."
- **what:** Deterministic pre-review checks on generated pages: extract all hrefs and verify internal links against the pages table, verify emails against site contact data; errors (broken links) force human review while warnings flow through. Companion mechanisms: prepare_link_context injects an only-link-to-these-pages allowlist into writer prompts, and rerender-time contact injection replaces hallucinated phone/email with DB truth.
- **sources:** docs018_rerendering/003_website_builder_architecture_status_report.md#3; docs018_rerendering/002_summary_link_constraints.md; docs018_rerendering/003_website_builder_architecture_status_report.md#6
- **relations:** content-reviewer workflow; link_registry; content-quality internal linking (successor).
- **verify-later:** validate_page_content + prepare_link_context in registry; prompt inclusion of link_constraint_text ("Not Yet Done" at the time).

### Content review flow with rejection → needs_attention
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** docs011/003 flow diagram (validate → auto-eval vs HITL → approve/reject → finalize_hitl or mark_page needs_attention); docs018/003: "Rejected pages are picked up by maintenance workflow."
- **what:** The content-reviewer agent's dual-mode gate: algorithmic validation feeds either auto-evaluation (errors escalate) or HITL review (human sees issues, can edit HTML inline); outcomes are approve (deploy), approve-with-edits (use edited HTML), or reject (page marked needs_attention, skipped by the build loop, queued for maintenance). Established review as a first-class pipeline stage rather than an afterthought.
- **sources:** docs011_api_hitl/003_hitl_new_plan.md; docs018_rerendering/003_website_builder_architecture_status_report.md#4; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Active-Agents
- **relations:** HITL protocol; flexible/strict approval snapshot; maintenance pickup of needs_attention.
- **verify-later:** content-reviewer workflow JSON; needs_attention status handling in current loop.

### Selective rebuild via build_status
- **category:** NEW:site-build-pipeline
- **status-signal:** deployed
- **status-evidence:** docs018/004 ("get_pages_to_build filters by build_status IN ('planned','needs_rebuild')"); docs015/003 documents the stale-page trap ("If the site planner didn't include use-cases in the new plan... it shows in nav but has stale content").
- **what:** Two orthogonal page state columns — status (active/deleted/needs_attention lifecycle) and build_status (planned/needs_rebuild/deployed) — let rebuilds touch only marked pages. Known failure mode: pages absent from a new plan silently keep old content while remaining in nav. Ancestor of work-item-driven rebuild targeting.
- **sources:** docs018_rerendering/004_trigger_just_pages_that_need_rebuild.md; docs015_data_flow_verification/003_temp_doc_rebuild_flow.md; docs017_legacy_agent_rules_images_design_keydocs/041_page_rebuild_action.md
- **relations:** page-rebuild agent; maintenance_queue (proto-work-items); work items.
- **verify-later:** get_pages_to_build action; build_status usage today.

### Slot-based modular page assembly (proposal)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** docs018/007 "Status: Draft for discussion, Created 2026-02-06"; user's inline answers ("agents should spawn their own dependencies", "no, we don't need migration"); site_components + render_site_components subsequently appear in the build path (docs018/006) but page_sections-as-JSON did not fully replace rendered_html storage.
- **what:** Replace regex header/footer injection with pure concatenation of slots (doctype/head/header/sections/footer); pre-render site-level components once into a site_components table; store section content as schema-validated JSON (page_sections) and render only at assembly so template changes never require content regeneration; explicit invalidation rules per change type; seven single-responsibility agents (site-planner, site-component-renderer, section-content-writer, link-manager, page-assembler, meta-manager, site-finalizer). Partially adopted: site_components and render_site_components shipped; JSON-first storage arrived instead as page_components.content_data source-of-truth.
- **sources:** docs018_rerendering/007_proposed_modular_rerendering.md; docs018_rerendering/006_build_path_rerender_path.md
- **relations:** recursive component tree (same instinct, earlier); InjectHead bug (motivation); section editor content_data principle.
- **verify-later:** site_components table + render_site_components action; page_sections existence.

### Stale orchestration sweeper
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** docs018/009_stale full design ("This is the #1 cause of pipeline stalls"; synthesize/retry/fail classification; "No schema changes needed"); no deployment confirmation in this unit.
- **what:** Replace lossy in-process timeout goroutines with a periodic DB sweep on every chassis pod: claim expired awaited_requests (FOR UPDATE SKIP LOCKED, 30s grace, LIMIT 20), classify — child COMPLETED means the response was lost, so synthesize a completion message from the child's final_result to the parent's topic; child FAILED forwards failure; no/running child retries up to retry_version 3 then fails the orchestration. Handles cascading stalls oldest-first and dead job topics by directly advancing parent state.
- **sources:** docs018_rerendering/009_stale_orchestration_sweeper_design.md
- **relations:** parent-timeout race; awaited_requests; debugging pipeline stalls; idle timeout/cleanup in system-architecture.
- **verify-later:** platform/orchestration/sweeper.go existence; sweeper startup in agentbase.

### Section editor agent (content_data is the source of truth)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** docs018/010 records decisions as implemented ("Decision: content_data is always the source of truth... We considered a lightweight html_patch edit type... decided against"; reused-code inventory; ActionInputSpec pattern used).
- **what:** Granular post-deploy editing without the full pipeline: two edit types — content_edit (field_updates merge or full content_data replace) and component_swap (new template, same content) — both updating page_components.content_data first, then re-rendering via buildRenderContextFromDB (reconstructs the full RenderContext purely from DB: site data, style collection, theme, nav, page meta, section content), reassembling the page, committing, and deploying. HTML patching was explicitly rejected because edits would vanish on the next rerender. Targets identified by page_component_id or page_name+slot_name (normalized). Future: LLM section rewrite via content_direction; bulk edits.
- **sources:** docs018_rerendering/010_section_editor_architecture.md; docs018_rerendering/008_granular_editing.md
- **relations:** granular editing spectrum; spatial addressing labels; rerender assemblePage; locks/inline editing (current descendants).
- **verify-later:** load_edit_context/apply_section_edit actions; section-editor agent definition.

### Granular editing spectrum (word → multi-page)
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** docs018/008 table mapping six edit scopes to mechanisms and LLM needs; "What Exists Today" vs "What's Needed: One New Agent, Two New Actions"; content_direction JSONB "(new)".
- **what:** A routing model for edit requests by scope: word/phrase (direct patch — later rejected in favour of content_data edits), field value (template re-render), section rewrite (content-writer on one section), component swap, page rewrite (page-rebuild with content_direction instructions flowing into prompts), multi-page (maintenance-triage → page-rebuild). All routed through the same maintenance infrastructure.
- **sources:** docs018_rerendering/008_granular_editing.md; docs018_rerendering/010_section_editor_architecture.md#Future-Extensions
- **relations:** section editor; work items; content_direction column on pages.
- **verify-later:** content_direction column and its prompt integration.

### Agent message structure & spawn+call pattern (external triggering)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** docs018/009_agent and docs018/005 document the working format ("Agents don't exist as running processes until spawned... You cannot call an agent that hasn't been spawned in the current workflow"; responses_topic always the caller's).
- **what:** The canonical three-layer message (Kafka headers, mirrored JSON headers, config.workflow + input_data payload) for driving agents from CLI or external systems, with inline workflow support on the generic agent, mandatory spawn-before-call, and reply routing to the sender's responses_topic enabling parent-child orchestration. The operational lingua franca of the whole system.
- **sources:** docs018_rerendering/009_agent_initial_message_structure.md; docs018_rerendering/005_triggering_agent_from_kafka.md; docs007_brochure_builder/003_original_message_copy
- **relations:** HITL protocol; generic agent as thin launcher; kafka reset/cleanup runbooks (docs007/003).
- **verify-later:** current message types in platform/orchestration/types.

### Canine biology knowledge tree (1M-agent demo)
- **category:** canine-biology
- **status-signal:** aspirational
- **status-evidence:** docs016/003c dated 2026-03-02, "Status: Working draft for further iteration"; docs016/004 (2026-03-03) demotes it: "best treated as marketing spend, not as a product... Build one branch (cardiovascular) as a polished showcase."
- **what:** A hierarchical agent swarm building a citable Labrador-reference knowledge tree: 7 levels of decomposition (root → body systems → aspects → subtopics → specific topics → mechanisms → molecular detail, branching 8–12), ~800K–1M agents across nine roles (Opus decomposers/synthesisers at top levels; BioMistral 7B research and finding-synthesis; non-LLM paper fetchers hitting PubMed; SciSpacy NER entity extractors; embedded-3B relevance filters; mermaid/FLUX diagram agents; 7B validators flagging cross-branch contradictions). Design priorities: accuracy over completeness; no reader-visible text from 3B models; phased rollout (125K live agents on five priority branches, background fill, then continuous PubMed-monitoring updates ~500-1000 agents/week); every node auditable (agent, prompt, sources, model); correction/discussion layer with versioning; pathway/mechanism cross-layer. Honest-risk section: credibility vs Plumb's/Merck, theatrical agent count, hallucination persistence, front-end decisive, costs 2-3x estimates ($2.2K-8.5K full run).
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md; docs016_dogs_medicine_pathways/002_project_outline.md; docs016_dogs_medicine_pathways/004_medical_business_reality_assessment.md
- **relations:** canine-biology category (docs018 feature plans); multicluster worker pools; model-tiering; business strategy demotion.
- **verify-later:** any decomposer/leaf agent definitions; knowledge tree tables (expected absent).

### Model-tiering by task ("the 3B problem")
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** docs016/003c "The 3B Problem" section: "A 3B model gets ~60-70% of this right. Errors at the leaf level propagate upward... Use the 3B model only for classification"; allocation table routing tasks across Opus/7B/BioMistral/NER/3B.
- **what:** A principled task-to-model allocation doctrine: frontier models only for structure-shaping decisions and top-level synthesis; domain-fine-tuned 7B for analysis; specialised tiny models (biomedical NER) beat general LLMs for structured extraction; embedded 3B only for binary classification; no LLM at all for retrieval. Pipeline design separates cheap structured extraction from semantic interpretation so the strong model gets one focused call. Generalizable beyond the canine project to any large-scale agent workload.
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md#The-3B-Problem; docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md#The-Paper-Analysis-Pipeline; docs016_dogs_medicine_pathways/002_project_outline.md
- **relations:** model-infrastructure (Ollama/GPU hosting); finetuning-flywheel; embedded worker-pod models.
- **verify-later:** inference cluster configs; any vLLM/BioMistral deployment manifests.

### Shared Kafka topic pools + worker pools for 1M agents
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** docs016/003c: "system.work.pool.{00-63} — 64 topics × 16 partitions = 1,024 partitions... Blast radius of a single bad agent is limited to ~1,000 co-located agents"; worker pod spec (embedded 3B via llama.cpp, SciSpacy, 5,000-10,000 goroutine workflows).
- **what:** Scaling architecture replacing per-agent topics with hashed shared pools carrying target_agent_id headers; long-running worker pods execute thousands of agent workflows as goroutines with local small models, routing bigger calls to shared inference servers; Redis/Valkey holds hot orchestration state (100K+ writes/sec) with Postgres persisting on completion; multi-cluster worker fleets via remote-job-spawner.
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md#Infrastructure-Design; docs016_dogs_medicine_pathways/002_project_outline.md#Infrastructure
- **relations:** multicluster docs021 (remote-job-spawner "proven working" per docs016/004); scheduler-and-tasks; canine project.
- **verify-later:** work pool topics in Kafka config; Redis state layer existence.

### Two-tier commercialisation model (sell output → sell setup)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** docs016/004 dated 2026-03-03 "Working notes for strategic direction": three-tier model trimmed to two ("drop the domain selling tier"); practical next step "produce 10 websites in a niche... validate with real money".
- **what:** Frank commercial assessment: framework differentiators are real infrastructure (K8s/Kafka/Postgres, data-driven workflows, multi-cluster, chassis pattern) but lack docs/community; revenue paths ranked (website service most mature; SEO content; document processing needs domain partner; framework sales longest); recommended model — run the service in a chosen niche to accumulate live outputs, then sell the whole setup as a business-in-a-box (£5-25K) once 20-50 outputs prove it, repeating per product; canine project reframed as portfolio/demo spend. Open decisions: niche, sellable quality, solo vs partner, runway.
- **sources:** docs016_dogs_medicine_pathways/004_medical_business_reality_assessment.md; docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** business-strategy category (pitch, domain strategy); early portfolio inventory; canine biology demotion.
- **verify-later:** n/a (strategy).

### Early portfolio inventory (honest capability notes)
- **category:** business-strategy
- **status-signal:** unknown
- **status-evidence:** docs018/008b raw notes: "None of our sites get leads at the moment so we can't say they do... we'd rather sell the sites achievement at the moment"; lists leopardessconsulting, vetcomparison.uk, wykefarm.co.uk, mortgagecalculator, website-design.com.
- **what:** A candid snapshot of what existed circa Feb 2026: leopardessconsulting built and evolving over days; a veterinary price-comparison site plus vet search/scrape/data-collection service; wykefarm.co.uk farm site (biodiversity content); a quickly-built but rough mortgage calculator site; website-design.com with functional tools (paste boards, mind maps, colour tools) but poor polish; framework scaling claim "several thousand agents". Useful ground truth for verifying which case-study sites actually functioned.
- **sources:** docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** site-case-studies (leopardess, vet, wyke); vet-med-pricing; dynamic-applications (website-design.com tools).
- **verify-later:** sites table rows for each named domain.

### "Database is source of truth, Git is the deployment artifact"
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** docs012/010 Core Principles ("Content lives in the database; Git is the deployment artifact"); reversal of docs012/004's "GitHub as current source of truth, database for metadata/links"; everything after (rerender, section editor, work items) depends on it.
- **what:** The pivotal data-ownership decision of the era: page content, sections, nav, entities, and design specs live in Postgres; git repos hold only rendered deployment artifacts, rebuilt from DB at will. Enables rerendering, granular editing, locking, and maintenance — and makes external git edits an anomaly to detect (git_hook_adapter desync idea) rather than a normal input.
- **sources:** docs012_site_maps_and_components/010_component_and_site_architecture.md#Core-Principles; docs012_site_maps_and_components/004_more_on_links.md#Context; docs018_rerendering/010_section_editor_architecture.md#The-Source-of-Truth-Principle
- **relations:** page_components; rerender; site-snapshots-and-revert (later formalization); deployment-github.
- **verify-later:** n/a (doctrine, observable in every pipeline).
