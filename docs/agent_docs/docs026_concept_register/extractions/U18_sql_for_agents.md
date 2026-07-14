# EXTRACTION U18 — docs/agent_docs/sql_for_agents/
Extracted 2026-07-13. Files in scope: 170. Concepts found: 62.

Chronology note: this directory is the de-facto history of the agent fleet.
`sql_for_agents_v1/` is the earliest era (monolithic LLM-chain site builders,
intake-orchestrator, ~2025-11 → 2025-12). `sql_for_agents_v2/` is the
component-architecture era (~2026-01). Root files 001–123 are the work-item
pipeline era (applied ad hoc); files 124–146 are the migration-runner era
(2026-07-04 onward, tracked in `schema_migrations`, baseline 124). File numbers
are used below as rough chronology evidence.

## Coverage
| file | treatment |
|---|---|
| 000_agent_definitions_backup_070_refactor.sql | family-delta |
| 001_remove_loops_in_workflow.md | full |
| 001_validator_sql.sql | full |
| 001b_implementation_plan.md | family-latest |
| 002_intake_orchestrator.sql | full |
| 003_site_classifier.sql | full |
| 006_site_architect.sql | header-scan |
| 011_site_deployer.sql | full |
| 017_multipage_website_builder.sql | header-scan |
| 019_chief_strategist.sql | full |
| 022_site_planner.sql | full |
| 023_page_content_writer_agent.sql | full |
| 024_research_agent.sql | full |
| 025_content_reviewer_agent.sql | full |
| 026_pageflow_builder.sql | full |
| 029_image_generator.sql | full |
| 030_input_mapping_changes.sql | full |
| 030b_remaining_agents_needing_input_mapping | full |
| 031_webdesign_agent.sql | full |
| 032_site_scraper_agent.sql | full |
| 033_rerender_pages_action.sql | full |
| 033_rerender_pages_trigger.sh | header-scan |
| 034_page_rerender_agent.sql | full |
| 035_render_site_components.sql | full |
| 036_rerender_site_agent.sql | full |
| 037_area_sweep_discoverer.sql | full |
| 038_area_sweep_orchestrator.sql | full |
| 039_page_rebuild_agent.sql | full |
| 040_optimise_which_llms.sql | full |
| 041_rag_knowledge_base.sql | full |
| 042_nav_updater_agent.sql | full |
| 042b_nav_link_fixer_agent.sql | full |
| 043_section_editor.sql | full |
| 044_asset_deployer.sql | full |
| 045_site_work_orchestrator.sql | full |
| 047_discovery_checks.sql | full |
| 048_discovery_agents.sql | family-delta (same content as 047, extended) |
| 049_domain_research_classifier.sql | full |
| 050_build_briefing_agent.sql | full |
| 051_build_dispatch_loop.sql | full |
| 052_build_pipeline_trigger.sql | full |
| 053_build_site_planner.sql | full |
| 054_improvement_loop.sql | full |
| 055_page_build_handler.sql | full |
| 056_colour_variable_fixer.sql | full |
| 057_image_build_handler.sql | full |
| 058_quality_checks_and_fixers.sql | full |
| 059_quality_discovery_agent.sql | family-delta (subset of 058) |
| 060_domain_strategist.sql | full |
| 061_tool_deployer_and_discovery_agent.sql | full |
| 062_tool_suggester_and_improver.sql | full |
| 062b_tool_deployer_and_generator_agent.sql | full |
| 063_vet_batch_processor.sql | header-scan |
| 063b_vet_practice_verifier.sql | full |
| 064_site_component_linker_and_fixer.sql | full |
| 065_page_build_handler_wrapper.sql | full |
| 066_audit_agent_definitions.sql | full |
| 067_implement_extended_thinking_not_yet_implemented.sql | full |
| 068_domain_submitter_agent.sql | full |
| 069_blog_posts.sql | full |
| 070_blog_content_planner.sql | full |
| 071_content_gap_planner.sql | full |
| 072_spec_updater_agent.sql | full |
| 073_create_new_client_schema.sql | header-scan |
| 074_completeness_discovery_agent.sql | full |
| 075_various_timeout_column.sql | full |
| 076_css_patch_agent.sql | full |
| 077_business_intel_companies_house.sql | full |
| 078_companies_house_ch_collector.sql | full |
| 079_companies_house_ch_matcher.sql | full |
| 080_companies_house_ch_llm_reviewer.sql | full |
| 081_companies_house_ch_detail-fetcher.sql | full |
| 082_company_number_scraper_ch_company_scraper.sql | full |
| 083_companies_house_ch_accounts_fetcher.sql | full |
| 084_site_review_agents.sql | full |
| 085_ai_endpoint_health_checker.sql | full |
| 086_visual_design_auditor.sql | full |
| 087_feeds_triage_ingester_orchestrator_etc.sql | full |
| 088_tool_auditor_agent.sql | full |
| 089_latest_news.sql | full |
| 090_b_content_feed_trigger.sql | full |
| 090_content_feed_orchestrator.sql | full |
| 091_site_adoption_agent.sql | full |
| 092_vet_med_pricing_agent.sql | full |
| 093_component_creator.sql | full |
| 094_vet_med_url_discoverer.sql | full |
| 095_vet_med_firecrawl_url_agent.sql | full |
| 096_vet_med_url_discover_orchestrator.sql | full |
| 098_tool_suggester_cross_linking.sql | full |
| 099_tool_recreation_handler.sql | full |
| 100_portfolio_use_cases_etc.sql | full |
| 101_internal_linker.sql | full |
| 102_component_quality_auditor.sql | full |
| 103_site_design_planner.sql | full |
| 104_site_adoption_orchestrator.sql | full |
| 105_rag_test_agent.sql | full |
| 106_training_data_exporter.sql | full |
| 107_image_build_handler.sql | full |
| 108_site_plan_pages.sql | full |
| 109_model_trainer_orchestrator.sql | full |
| 110_training_data_preparer.sql | full |
| 111_gpu_provisioner_thunder.sql | full |
| 112_training_launcher.sql | full |
| 113_site_asset_renderer.sql | full |
| 114_thunder_reaper.sql | full |
| 115_locks.sql | full |
| 116_thunder_training_monitor_worker.sql | full |
| 117_thunder_training_monitor_orchestrator.sql | full |
| 118_code_indexer_for_analyser.sql | full |
| 119_intent_events_for_vms.sql | full |
| 120_intent_site_stats.sql | full |
| 121_intent_collector_agents.sql | full |
| 122_diagnose_agents.sql | full |
| 123_component_creator.sql | full |
| 124_schema_migrations.sql | full |
| 125_doc_plans_and_notes.sql | full |
| 126_wire_persist_diagnosis_note.sql | full |
| 127_diagnose_load_runtime_error_step.sql | full |
| 128_fix_load_runtime_error_step_target.sql | full |
| 129_wire_diagnosis_subject_threading.sql | full |
| 130_pilot_plan_tool_archetype_taster_quiz.sql | full |
| 131_tool_generator_plan_writing.sql | full |
| 132_fix_agents_note_writing.sql | full |
| 133_add_component_provenance.sql | full |
| 134_fix_prompt_template_field_paths.sql | full |
| 135_bypass_index_plan_until_embed_timeout.sql | full |
| 136_supersede_xp_curve_plan_selectors.sql | full |
| 137_recreation_spec_and_note_subject.sql | full |
| 138_recreate_tool_carries_spec_features.sql | full |
| 139_reenable_index_plan.sql | full |
| 140_rebypass_index_plan_chunk_loop.sql | full |
| 141_reenable_index_plan_after_chunk_fix.sql | full |
| 142_enable_tool_acceptance_check.sql | full |
| 143_supersede_plans_inline_delivery.sql | full |
| 144_composer_inline_delivery.sql | full |
| 145_tool_acceptance_agent.sql | full |
| 146_enable_tool_acceptance_due.sql | full |
| sql_for_agents_v1/001_agent_definitions_etc.sql | family-delta |
| sql_for_agents_v1/002_intake_orchestrator.sql | family-delta |
| sql_for_agents_v1/003_site_classifier.sql | family-delta |
| sql_for_agents_v1/004_website_builder.sql | family-delta |
| sql_for_agents_v1/005_domain_analyst.sql | family-delta |
| sql_for_agents_v1/006_site_architect.sql | family-delta |
| sql_for_agents_v1/007_content_creator.sql | family-delta |
| sql_for_agents_v1/008_html_developer.sql | family-delta |
| sql_for_agents_v1/009_all | family-delta |
| sql_for_agents_v1/010_contract_validation.sql | family-delta |
| sql_for_agents_v1/011_site_deployer.sql | family-delta |
| sql_for_agents_v1/012_multipage_wrapper.sql | family-delta |
| sql_for_agents_v1/014_html_developer_chunked.sql | family-delta |
| sql_for_agents_v1/015_example_20_page_workflow.sql | family-delta |
| sql_for_agents_v1/016_landing_page_builder.sql | family-delta |
| sql_for_agents_v1/017_multipage_website_builder.sql | family-delta |
| sql_for_agents_v1/019_chief_strategist.sql | family-delta |
| sql_for_agents_v1/020_where_we_are_now | header-scan (schema snapshot) |
| sql_for_agents_v2/000_backup.sql | family-delta |
| sql_for_agents_v2/000_backup_agents.sql | family-delta |
| sql_for_agents_v2/001_general_rule.md | full |
| sql_for_agents_v2/002_intake_orchestrator.sql | family-delta |
| sql_for_agents_v2/006_site_architect.sql | family-delta |
| sql_for_agents_v2/007_content_creator.sql | family-delta |
| sql_for_agents_v2/017_multipage_website_builder.sql | family-delta |
| sql_for_agents_v2/019_chief_strategist.sql | family-delta |
| sql_for_agents_v2/022_site_planner.sql | family-delta |
| sql_for_agents_v2/023_page_content_writer_agent.sql | family-delta |
| sql_for_agents_v2/024_research_agent.sql | family-delta |
| sql_for_agents_v2/025_content_reviewer_agent.sql | family-delta |
| sql_for_agents_v2/026_pageflow_builder.sql | family-delta |
| sql_for_agents_v2/027_old_agent_definitions.sql | family-delta |
| sql_for_agents_v2/027_replace_claude_model_names.sql | full |

## Concepts

### v1 monolithic LLM-chain site builders (website-builder, domain-analyst, site-architect, content-creator, html-developer, multipage-wrapper, html-assembler, site-deployer)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** v2/026_pageflow_builder.sql renames multipage-website-builder v3 → pageflow-builder ("Component-based website builder... uses DB components for structure, LLM only for content"); v2/027_old_agent_definitions.sql captures the whole v1 fleet as "old"; root files never patch these agents again.
- **what:** The first architecture (2025-11/12): a website-builder orchestrator spawns a chain of one-LLM-call specialists — domain-analyst (audience/tone JSON), site-architect (page structure + colours), content-creator (copy JSON), html-developer (whole-page HTML), multipage-wrapper (file map), site-deployer (git commit). Everything is free-form LLM output; no component library, no DB page records.
- **sources:** sql_for_agents_v1/004_website_builder.sql; sql_for_agents_v1/005_domain_analyst.sql; sql_for_agents_v1/008_html_developer.sql; sql_for_agents_v2/027_old_agent_definitions.sql
- **relations:** superseded by pageflow-builder (component-based); site-deployer contract survives in 011_site_deployer.sql
- **verify-later:** agent_definitions rows for these types (is_active, deleted_at); whether any workflow still references them

### Batched multi-page generation (multipage-website-builder) and chunked HTML generation
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** v2/026 renames its v3 to pageflow-builder; the batch-of-4-pages prompts (generate_batch_1..5) appear only in v1/015 and v1/017 snapshots.
- **what:** Anti-token-limit strategies from the v1 era: build 20-page sites by generating pages in five batches of four ("Return as JSON map of filename to HTML"), with shared CSS generated once and injected at assembly; html-developer-chunked generated structure/styles/sections in separate calls. Both are ideas the component architecture made unnecessary.
- **sources:** sql_for_agents_v1/015_example_20_page_workflow.sql; sql_for_agents_v1/017_multipage_website_builder.sql; sql_for_agents_v1/014_html_developer_chunked.sql
- **relations:** replaced by pageflow-builder per-page loop and later the one-item-per-run dispatch loop
- **verify-later:** none needed — historical

### intake-orchestrator (classify → brief → HITL confirm → spawn builder → rerender)
- **category:** hitl
- **status-signal:** superseded
- **status-evidence:** 068_domain_submitter_agent.sql creates a new entry point ("Entry point for new domain submissions... creates needs_domain_research work item"); intake files stop being patched after 030-era; 002 header shows HITL steps with `skip_if: input_data.hitl_mode == auto`.
- **what:** The v1/v2 entry pipeline: discovers available `%-builder` agents, spawns site-classifier and briefing-agent, runs two human-in-the-loop gates (confirm site type; review brief), spawns the recommended builder, then (added later) a rerender pass for nav consistency. Notable mechanisms: `dynamic_select` HITL fields fed from a live agent query; `skip_if` auto mode making HITL optional per run.
- **sources:** 002_intake_orchestrator.sql; sql_for_agents_v1/001_agent_definitions_etc.sql; sql_for_agents_v2/002_intake_orchestrator.sql
- **relations:** superseded by domain-submitter + build-dispatch-loop; HITL gate pattern survives in content-reviewer HITL mode
- **verify-later:** agent_definitions 'intake-orchestrator' status; whether any trigger path still uses it

### site-classifier → research-backed classification with domain_profile
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** 003 header: "Changes site-classifier from a single Haiku LLM guess into a research-backed orchestrator"; later 049 creates domain-research-classifier for the work-item pipeline, which takes over first-stage classification.
- **what:** Evolution of classification: v1 was one Haiku call mapping domain+objective to {landing, content, portfolio, brochure} + recommended_builder. v2 (file 003) made it an orchestrator: Haiku research brief → research-agent web investigation → Sonnet synthesis producing backward-compatible site_type plus a rich domain_profile (business identity, tone, visual_direction, image_guidance, strategic analysis). Explicit responsibility fences: does NOT pick pages or style_collection (planner's job) but DOES provide design inputs consumed by planner, image-generator, webdesign-agent, page-content-writer.
- **sources:** 003_site_classifier.sql; sql_for_agents_v1/003_site_classifier.sql; sql_for_agents_v2/000_backup_agents.sql
- **relations:** succeeded by domain-research-classifier (work-item pipeline); domain_profile is ancestor of site_specs aspects
- **verify-later:** live site-classifier definition vs domain-research-classifier; who consumes domain_profile today

### pageflow-builder (component-based site build orchestration)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** Still being patched in Phase 0 imagery work 2026-05-05 (107 backs up its row before migration); 026 documents the full live step chain.
- **what:** The central v2-era builder, renamed from multipage-website-builder v3. Spawns planner/content-writer/reviewer/deployer, then: ensure_site_record → call_site_planner → store brief+plan → sync_pages_to_db → populate_nav → asset steps → select_style_collection → set_default_components → render_site_components → get_pages_to_build (filters by build_status) → build_pages_loop (write → review → assemble → deploy per page) → apply_site_design (CSS) → trigger_site_deploy (Cloudflare). The known hazard that sync_pages_to_db can reset page statuses is documented in-file.
- **sources:** 026_pageflow_builder.sql; sql_for_agents_v2/026_pageflow_builder.sql; 107_image_build_handler.sql (backup section)
- **relations:** parallel/legacy path beside site-work-orchestrator and build-dispatch-loop; uses site-planner, page-content-writer, content-reviewer, deployer-agent
- **verify-later:** whether new sites still route through pageflow-builder or only via the work-item pipeline

### Remove-loops plan: input_mapping, contract validation, sequential_fan_out, page-builder worker
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** input_mapping conversion executed in 030/030b/023; contracts added across many files ("These contracts define what each agent expects... fail fast"); but build_pages_loop still present in 026 and sequential_fan_out/page-builder never appear in any later agent file — that half is effectively abandoned in favour of the dispatch loop.
- **what:** A four-phase plan to replace loop/substep injection: (1) explicit `input_mapping` instead of `input_fields` path-hunting plus runtime input/output contract validation with hard fails and `__raw_message__` deprecation; (2) a `sequential_fan_out` action spawning one child orchestration per page; (3) a page-builder worker agent; (4) rewire pageflow-builder. Phases 1 landed; phases 2–4 were superseded by the site_work_items dispatch-loop architecture, which achieves the same "one visible orchestration per unit of work" goal differently. 001_validator_sql.sql is a jsonb_path_query audit extracting every field path referenced in workflows.
- **sources:** 001_remove_loops_in_workflow.md; 001b_implementation_plan.md; 030_input_mapping_changes.sql; 030b_remaining_agents_needing_input_mapping; 001_validator_sql.sql
- **relations:** input contracts appear in nearly every agent file (002, 011, 022, 024, 025, 029...); dispatch loop (051) is the spiritual successor of sequential_fan_out
- **verify-later:** chassis code: contract validation enforcement; whether `sequential_fan_out` action exists in the registry; `__raw_message__` fallback removal status

### Input/output contracts on agent definitions
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Contract UPDATE statements across 011/022/024/025/029 etc.; 129 extends contracts as live behaviour ("input_mapping must satisfy the input_contract — 016b §9 spawn+call rule").
- **what:** Every agent row carries `input_contract` (required/optional fields) and `output_contract` (produces). Contracts are both documentation and runtime validation hooks; the 2026-07 diagnosis work established the durable rule that an input the workflow reads must be declared in the contract (137's "spec is UNDECLARED" fix) and that call-site input_mapping must satisfy the callee's contract.
- **sources:** 011_site_deployer.sql; 022_site_planner.sql; 129_wire_diagnosis_subject_threading.sql; 137_recreation_spec_and_note_subject.sql; sql_for_agents_v1/009_all
- **relations:** remove-loops plan; workflow_contract_chain view (v1/010)
- **verify-later:** chassis contract-validation code path; how strictly contracts fail fast in production

### Call metadata vs response-data convention (output_field.response)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** v2/001_general_rule.md states it as "The general rule going forward"; 116 confirms it verified in coordinator.go ("a step result is stored under BOTH the step name AND its output_field, adapter body under .response").
- **what:** Workflow data-shape convention: when a step calls another agent, call metadata (agent_id, request_id, topics) is stored directly at the step's output_field while the called agent's response payload lands at `output_field.response`. Many prompt-template and field-path bugs in this directory trace to violating this shape.
- **sources:** sql_for_agents_v2/001_general_rule.md; 116_thunder_training_monitor_worker.sql; 003_site_classifier.sql (classification.response.result paths)
- **relations:** template field-path rules (134); input_mapping
- **verify-later:** coordinator.go result-storage code (~L1636/L2408 per 116)

### chief-strategist (build-plan LLM) + component placement dedup rules
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** 040 still upgrades its model (haiku→sonnet) but the work-item pipeline planner (build-site-planner, 053) owns planning thereafter; 019 patch injects "COMPONENT PLACEMENT RULES" into its prompt.
- **what:** The v1/v2 planning agent producing sections/component_details build plans. Its lasting contribution is the component placement rule-set injected by 019: testimonials/team-grid/faq/contact-form on ONE page only, per-page hero variants, no duplicated services content, merge similar pages — an early anti-repetition contract for planners.
- **sources:** 019_chief_strategist.sql; sql_for_agents_v1/019_chief_strategist.sql; sql_for_agents_v2/019_chief_strategist.sql; 040_optimise_which_llms.sql
- **relations:** site-planner, build-site-planner inherit the planning role; parse_json_field/unwrapDeep pattern (v1/019)
- **verify-later:** is chief-strategist still active or deleted

### site-planner (single-LLM-call site plan)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** 022 shows model flip-flops (sonnet→haiku for cost, 040 haiku→sonnet because planning is "high-leverage"); 053 build-site-planner is the successor for work-item builds.
- **what:** v2 planner: one LLM call over brief + component library + style collections producing validated_plan, pages, style_collection, needs_logo/needs_images. The model-choice oscillation (cost vs quality on high-leverage decisions) is documented reasoning worth keeping.
- **sources:** 022_site_planner.sql; sql_for_agents_v2/022_site_planner.sql; 040_optimise_which_llms.sql
- **relations:** chief-strategist (predecessor), build-site-planner (successor), pageflow-builder (caller)
- **verify-later:** which planner the live pipelines invoke

### page-content-writer (section-by-section content generation)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** Continuously patched from v2 era through 069 (reads site_specs), 107-era imagery direction; 075 gives it idle_timeout 180.
- **what:** Writes one page section-by-section: spawn_research_agent → load_page_components → build_render_context → process_sections_loop (per-section LLM call constrained to that component's `llm_field_specs`) → compile_page. The prompt is a major behavioural contract: official-contact-only rule, internal-link constraint to listed pages, content_direction/imagery_direction from site_specs, admin content briefs, "Recreate Mode" for adopted sites (adapt original page markdown), and an 18-rule anti-fabrication list (no invented people/testimonials/statistics/case studies; "ALWAYS better to be honest and general than specific and fabricated").
- **sources:** 023_page_content_writer_agent.sql; sql_for_agents_v2/023_page_content_writer_agent.sql; 069_blog_posts.sql
- **relations:** called by pageflow-builder, page-build-handler; feeds save_page_sections/page_components; anti-fabrication rules relate to content-governance
- **verify-later:** live prompt_template vs the 023 copies; llm_field_specs source in content_components

### content-reviewer (HITL + auto-eval dual mode with pre-validation)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** 025 adds validate_content step "runs BEFORE review mode determination"; agent defined active in v2/025.
- **what:** Reviews generated page content in either human (request_human_input, editable) or auto-eval (LLM approve/flag) mode, selected at runtime. 025 adds algorithmic pre-validation: internal links must point to existing site pages; email addresses must match the site's contact_email — findings are handed to whichever review mode runs.
- **sources:** 025_content_reviewer_agent.sql; sql_for_agents_v2/025_content_reviewer_agent.sql
- **relations:** page-content-writer output consumer; HITL pattern from intake-orchestrator
- **verify-later:** validate_page_content action; current review-mode default

### research-agent (cited web research into research_results)
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** Defined active v1.0.575 (v2/024); idle timeout set in 075; classifier v2 (003) depends on it.
- **what:** Web-search research specialist that extracts relevant quotes, synthesises findings with full source attribution and stores in a research_results table for citation ([0], [1] markers consumed by page-content-writer prompts).
- **sources:** sql_for_agents_v2/024_research_agent.sql; 024_research_agent.sql; 003_site_classifier.sql
- **relations:** spawned by page-content-writer and site-classifier v2
- **verify-later:** research_results table usage

### image-generator + image prompt plumbing
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Contract in 029; Phase 0.1 (107) wires site_id through six call sites "so it can read design_intent.imagery_direction from site_specs".
- **what:** AI image generation specialist taking prompt/image_prompts (logo, hero_home) and producing image_url/image_data. Phase 0.1 made it site-aware: callers pass site_id so the generator composes design_intent.imagery_direction into prompts.
- **sources:** 029_image_generator.sql; 107_image_build_handler.sql (Section 1)
- **relations:** image-build-handler, site-work-orchestrator, pageflow-builder call it; asset-deployer deploys results
- **verify-later:** image generation adapter/action; imagery_direction composition code

### image-build-handler + needs_imagery kind branch (Phase 2G)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 057 defines the handler; 107's "phase_2g_step5" section adds check_item_type_imagery / spawn_image_gen_imagery / call_imagery_gen / check_imagery_brand_update / store_imagery_brand_asset|store_imagery_asset steps ("teach image-build-handler to process needs_imagery work items (emitted by step 4's check_unfulfilled_imagery_plan)").
- **what:** Self-contained dispatch-loop handler for image work items: originally needs_logo/needs_hero_image (branch on spec.purpose → call image-generator → store_asset → deploy_image_asset via S3/optimize/git). Phase 2G extends it to generic needs_imagery items carrying kind-specific behaviour (icon transparency, logo variants), routed by item_type, with a spec.brand_update boolean deciding whether the stored asset also updates site brand assets.
- **sources:** 057_image_build_handler.sql; 107_image_build_handler.sql
- **relations:** build-dispatch-loop (caller); check_unfulfilled_imagery_plan discovery (imagery plan reconciliation); asset-deployer
- **verify-later:** live image-build-handler workflow steps; needs_imagery item emission in discovery checks

### Imagery provenance: origin_model + origin_prompt on assets
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 107 combined migration Sections 2–3 (2026-05-05): "set origin_model literal on store_asset steps so the assets table records what produced each image"; origin_prompt_field normalised to record "the actual composed prompt sent to the model... not the un-composed plan prompt".
- **what:** Asset provenance discipline: every stored image records the generating model and the exact post-composition prompt. Required coordinated Go+SQL shipping (three concerns in one transaction) across image-build-handler, site-work-orchestrator, pageflow-builder.
- **sources:** 107_image_build_handler.sql (Sections 2–3 + backup preamble)
- **relations:** imagery audit work (file says provenance is "better for future iterations of the imagery audit work")
- **verify-later:** assets table columns origin_model/origin_prompt population

### webdesign-agent (full CSS stylesheet generation)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 4,683-line definition file; referenced as the "full theme" path by 076 ("Unlike webdesign-agent (which regenerates everything from scratch)"); idle timeout in 075.
- **what:** Generates production CSS for a site. Accepts a provided site_context or loads context from DB (conditional first step), analyzes design requirements, writes stylesheet via git_commit with file_path config. It is the heavyweight regeneration path, contrasted with css-patch-agent for targeted fixes.
- **sources:** 031_webdesign_agent.sql; 076_css_patch_agent.sql; 103_site_design_planner.sql (resolved_composition reader list)
- **relations:** site-scraper feeds it site_context; css themes/style_collections; site-design-planner
- **verify-later:** current webdesign-agent workflow vs 031 copy; patch_01_git_commit_file_path.go

### site-scraper (Firecrawl scrape → site_context)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 032 definition ("Uses firecrawl_scrape action (requires patch_02_webscrape_url_field.go)").
- **what:** Scrapes a live site's homepage via the webscrape adapter (Firecrawl), then an LLM step transforms results into the site_context format webdesign-agent consumes — the original design-transfer mechanism.
- **sources:** 032_site_scraper_agent.sql
- **relations:** webdesign-agent; ancestor of site-adoption-agent's full crawl
- **verify-later:** whether site-scraper is still used vs site-adoption-agent

### Rerender pipeline (rerender-pages, page-rerender, render-site-components, rerender-site)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** 033–036 sequence; idle timeouts for rerender-pages/page-rerender in 075; needs_rerender items created by many fixers (056, 064) with handler rerender-pages.
- **what:** The assembly/deployment half of the system, separated from content generation: page_components store rendered sections; render_site_components renders header/footer/head into site_components; page-rerender re-assembles a single page from stored sections (with skip detection) and deploys; rerender-site orchestrates site-wide re-render (components → per-page loop → Cloudflare deploy). Design principle stated in 036: the loop sub_workflow is minimal, all per-page logic lives in the page-rerender agent. needs_rerender work items (priority 99, run last) are the standard "make fixes visible" side-effect.
- **sources:** 033_rerender_pages_action.sql; 034_page_rerender_agent.sql; 035_render_site_components.sql; 036_rerender_site_agent.sql; 064_site_component_linker_and_fixer.sql
- **relations:** nav-updater (adds nav refresh first); every fixer agent that returns needs_rerender
- **verify-later:** rerender_single_page / render_site_components actions; needs_rerender dedup guard (NOT EXISTS insert in 064)

### Navigation maintenance: nav-updater and nav-link-fixer
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** 042 full definition ("Algorithmic only - no LLM calls"); nav-link-fixer in 075 idle-timeout list; 058 wires it as fixer for broken_nav_links findings.
- **what:** nav-updater refreshes nav tables from current pages (populate_nav_tables), re-renders header/footer/head and reassembles all deployed pages — explicitly distinguished from rerender-site, which reuses stale site_nav_items. nav-link-fixer repairs the `#{{.slug}}` anti-pattern in header/footer component templates (should be `{{.url}}`), then force re-renders site components and pages.
- **sources:** 042_nav_updater_agent.sql; 042b_nav_link_fixer_agent.sql; 058_quality_checks_and_fixers.sql
- **relations:** quality-discovery-agent's broken_nav_links check; orphan_nav finding; rerender pipeline
- **verify-later:** populate_nav_tables / fix_nav_link_templates actions

### section-editor (granular edits that survive re-renders)
- **category:** content-governance
- **status-signal:** unknown
- **status-evidence:** Full definition with example trigger messages in 043; no later patches or timeout entries reference it.
- **what:** Edits a single page section without the full rebuild pipeline. Core invariant: always update content_data first, then re-render from template + DB context, so edits survive future re-renders (nav updates, theme changes). Supports content_edit (field merge or full replace) and component_swap (new template, same content_data). Target addressed by page_component_id or (page_name + slot_name).
- **sources:** 043_section_editor.sql
- **relations:** inline editing / content governance concepts; page_components
- **verify-later:** whether section-editor exists live and is used by admin UI

### asset-deployer (S3 → optimize-by-purpose → git)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** 044 definition; idle timeout in 075; called from image-build-handler flows.
- **what:** Single-purpose specialist wrapping deploy_image_asset: downloads an asset from S3, optimizes it according to purpose (logo vs hero), commits to git. Reusable for any image deploy task.
- **sources:** 044_asset_deployer.sql; 057_image_build_handler.sql
- **relations:** image-build-handler, undeployed_assets discovery check
- **verify-later:** deploy_image_asset action; optimization rules per purpose

### site-work-orchestrator (unified build/maintenance over site_work_items)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 045 definition; row backed up and patched in Phase 0 imagery (107, 2026-05-05).
- **what:** Orchestrator that builds sites from prioritized site_work_items rows, calling appropriate handler agents per item, "compatible with pageflow-builder's planner and content writer". The first expression of the unified build/maintenance queue idea, later refined into the one-item-at-a-time build-dispatch-loop.
- **sources:** 045_site_work_orchestrator.sql; 107_image_build_handler.sql
- **relations:** site_work_items table; build-dispatch-loop (leaner successor/sibling); discovery agents write into its queue
- **verify-later:** which orchestrator the live triggers use; site_work_items schema

### Work-item build pipeline: domain-submitter → dispatch loop → handler agents
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 051/052 definitions; 075 sets idle timeouts across the whole handler fleet; 146 still adding items into the same queue in 2026-07.
- **what:** The current architecture. domain-submitter (068) creates a site record + needs_domain_research item from just a domain. build-pipeline-trigger (052) is a 30-min heartbeat: seeds the build queue, finds one site with pending items, fires build-dispatch-loop (051), which loads the highest-priority claimable item, claims it, spawns+calls the handler agent, marks complete, and if items remain spawns a FRESH dispatch loop (separate orchestration, clean logs). Item chain for a new site: needs_domain_research → needs_strategy → needs_briefing → needs_site_plan → needs_content_page (per page) → images → needs_rerender. Concurrency safety via claim_work_item; health-gating via ai_endpoint_health before claiming.
- **sources:** 051_build_dispatch_loop.sql; 052_build_pipeline_trigger.sql; 068_domain_submitter_agent.sql; 085_ai_endpoint_health_checker.sql
- **relations:** every handler agent below; scheduler-and-tasks (CronJob trigger); replaces intake-orchestrator
- **verify-later:** LoadWorkItemsAction first_item patch; claim semantics; current item_type → handler_agent routing table

### domain-research-classifier (work-item first stage)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 049 header documents pipeline position "first agent after seed_build_queue"; 067 adds extended-thinking budget to its classify_and_extract step (conditional on patch deploy).
- **what:** Handler for needs_domain_research: researches a domain via web search and scrape, classifies site type, extracts identity signals, writes site_specs aspects "identity" and "classification", creates the next work item (needs_briefing; later needs_strategy per 060).
- **sources:** 049_domain_research_classifier.sql; 060_domain_strategist.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** successor of site-classifier v2; site_specs aspect model
- **verify-later:** current next-item wiring (strategy vs briefing)

### domain-strategist (strategy vs architecture separation)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** 060 definition with explicit responsibility statement; work item chain needs_strategy → needs_briefing.
- **what:** Handler for needs_strategy items. Determines the strategy for a domain — canonical site_type, revenue model, content strategy, page_type recommendations, tone/positioning — and writes site_specs aspect "strategy". Explicit contract: does NOT design page architecture; "The planner has final say... may agree, adjust, or override"; does NOT overwrite the researcher's "classification" aspect.
- **sources:** 060_domain_strategist.sql
- **relations:** build-site-planner reads strategy; domain-research-classifier upstream
- **verify-later:** strategy aspect consumption in plan_site prompt

### build-briefing-agent (spec-reading briefing)
- **category:** onboarding-config
- **status-signal:** deployed
- **status-evidence:** 050 definition, "Distinct from existing briefing-agent (v1) which... receives questionnaire directly as input. This version reads from site_specs."
- **what:** Handler for needs_briefing: answers the briefing questionnaire autonomously from site_specs identity + classification (no human), writes aspect "briefing", creates needs_site_plan. Marks the shift from HITL-driven briefing to spec-derived config.
- **sources:** 050_build_briefing_agent.sql
- **relations:** v1 briefing-agent (superseded for this path); build-site-planner downstream
- **verify-later:** briefing aspect shape

### build-site-planner + roadmap-overrides-components rule
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 053 shows the workflow rewired to the site_plans domain ("changed to: ... write_site_plan → sync_pages → populate_nav → reconcile_site_plan → complete"); plan_site runs on claude-opus-4-6; 067 adds thinking budget.
- **what:** Handler for needs_site_plan. Reads site_specs (identity/classification/briefing/strategy), loads component library and style collections, plans via LLM, validates, then writes into the site_plans domain and reconciles. Carries the ROADMAP OVERRIDE rule verbatim: "ROADMAP OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the current phase... use EXACTLY the section_types listed — even if they do not appear in the Available Section Components list... Unknown section types are handled by the component selector downstream. Do NOT replace roadmap section_types with standard components... The roadmap is the authority for this site." Earlier form wrote plan/design_intent/content_direction specs + write_build_items (one needs_content_write per page).
- **sources:** 053_build_site_planner.sql; 108_site_plan_pages.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** site_plans/reconciler domain (docs 029/030); component selector creating needs_new_component items; roadmap spec aspect
- **verify-later:** write_site_plan + reconcile_site_plan actions; roadmap aspect producer

### site_plan_pages schema repair (plan-domain drift)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 108 "Migration 033: Reconcile site_plan_pages columns + drop orphan site_plan_partials... every write_site_plan call to date has failed at the title-column error."
- **what:** Repairs drift between two drafts of the site-plan schema: adds title/meta_description/nav_label columns, drops page_data and the unused site_plan_partials table (directives are row-per-directive in site_plan_directives). Documents the CREATE TABLE IF NOT EXISTS silent-skip failure mode when a rewritten migration follows an applied earlier draft.
- **sources:** 108_site_plan_pages.sql
- **relations:** build-site-planner; migration-discipline concepts (124)
- **verify-later:** live \d site_plan_pages / site_plan_directives

### page-build-handler (content-page handler with section planning and validation gates)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 065 documents the evolved workflow with plan_sections/validate_content and error paths; 070 notes empty_sections handler switched to it.
- **what:** Wrapper solving "specialist vs handler": page-content-writer generates but doesn't persist, so this handler loads page + specs, plan_sections resolves data sources per section (creating deferred items when sections aren't ready), calls the writer, validate_content checks placeholders/templates/cross-site contamination (blockers → mark_needs_review), then save_page_sections, update_page_status, and deploys via page-rerender. Earlier version (055) was simpler (no plan/validate steps, no deploy).
- **sources:** 065_page_build_handler_wrapper.sql; 055_page_build_handler.sql; 070_blog_content_planner.sql
- **relations:** page-content-writer, page-rerender; content_rewrite items route here; needs_new_component items from plan_sections
- **verify-later:** plan_sections + validate_page_content actions; deferred-item creation

### page-rebuild (rebuild pages without re-planning)
- **category:** NEW:build-pipeline
- **status-signal:** unknown
- **status-evidence:** 039 full definition with detailed reuse/skip lists; no later references found in this unit.
- **what:** Rebuilds specific pages (build_status='needs_rebuild') on an existing site loading all context from DB given a domain, explicitly skipping planner, sync_pages_to_db, asset generation, component rendering, CSS and nav (all already done) while reusing the standard build-loop agents. Documents design principles: agent owns its domain; spawnable not standalone; reuse before creating; complexity in Go.
- **sources:** 039_page_rebuild_agent.sql
- **relations:** pageflow-builder (same loop, different input_mapping via rebuild_context); load_site_for_rebuild action
- **verify-later:** whether page-rebuild survived the dispatch-loop refactor

### Discovery agents (design / quality / completeness) and the check registry
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 047/048 define them; 074 expands completeness checks; 142/146 still adding checks to design-discovery in 2026-07 ("run_discovery_checks warns... and skips unknown names" — the safe-rollout pattern).
- **what:** Read-only detectors that "find problems. They do not fix anything. They do not call other agents." Each runs run_discovery_checks with a named check list, writing findings to site_work_items (source='discovery', status='detected'). design: undeployed_assets, missing_css, duplicate_palette, missing_tools, tool_health, tool_acceptance, tool_acceptance_due. quality: broken_nav_links, placeholder_contact, generic_theme. completeness: empty_sections plus integrity checks — cross_site_contamination, unrendered_templates, missing_style_collection, deactivated_site_components. All algorithmic, no LLM budget. Unknown check names warn-and-skip, so SQL can enable a check before the Go ships.
- **sources:** 047_discovery_checks.sql; 048_discovery_agents.sql; 058_quality_checks_and_fixers.sql; 074_completeness_discovery_agent.sql; 142_enable_tool_acceptance_check.sql; 146_enable_tool_acceptance_due.sql
- **relations:** improvement-loop orchestrates them; fixer agents consume their items; check registry in discovery_checks.go
- **verify-later:** registered checks in run_discovery_checks_action.go / discovery_checks/*.go

### improvement-loop (post-build discovery → triage → fix → rerender cycle)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 054 definition; 086 Part B adds audit_pass_count guard "stops after 3 passes"; 100 portfolio claims sites "receive autonomous content audits... on rolling schedules".
- **what:** Runs after initial build (or on schedule/manual trigger): spawns the three discovery agents, triage_detected_items promotes detected → triaged, and if anything was promoted inserts needs_rerender at priority 99 and fires build-dispatch-loop to process all fixes then rerender. 086's audit-pass cap plus section locking provide the loop's termination condition ("the triage drain").
- **sources:** 054_improvement_loop.sql; 086_visual_design_auditor.sql; 061_tool_deployer_and_discovery_agent.sql (flow diagram)
- **relations:** discovery agents, fixers, audit agents, locks
- **verify-later:** improvement-sweep scheduled task; triage_detected_items action; audit_pass_count in sites.settings

### Fixer agents: color-variable-fixer, site-component-linker, component-template-fixer, css-patch-agent
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 056/064/076 definitions; all in 075's idle-timeout list; component-template-fixer gains note-writing in 132.
- **what:** Narrow algorithmic/LLM fixers dispatched from the queue: color-variable-fixer replaces hardcoded hex in component inline styles with CSS variables (fixes both templates permanently and rendered_html immediately); site-component-linker fixes NULL component_id causing fallback rendering; component-template-fixer applies targeted template surgery (nav flex CSS injection, element removal, slot_name alignment) routed on spec.fix_type; css-patch-agent LLM-patches the current stylesheet for spacing/responsive/layout issues without full regeneration (explicitly NOT theme redesign — that's webdesign-agent). All create deduplicated needs_rerender items only when they changed something.
- **sources:** 056_colour_variable_fixer.sql; 064_site_component_linker_and_fixer.sql; 076_css_patch_agent.sql; 058_quality_checks_and_fixers.sql
- **relations:** discovery checks (hardcoded_section_colors, unlinked components); rerender pipeline; audit findings
- **verify-later:** fix_hardcoded_colors, link_site_components, fix_component_template actions

### Audit agent hierarchy (visual-design-auditor, content-quality-auditor, design-audit-agent, site-review-agent)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 066 defines the hierarchy; 084 patches prompts due to a real cost incident ("845 design-audit work items across 4 domains in ~10 days... cost explosion"); 086 excludes locked components from audit queries.
- **what:** LLM auditors layered above discovery: pattern is "algorithmic checks first, then ONE LLM call for subjective assessment, then write findings" (write_audit_findings). 084 makes findings structured and bounded: TOP 5 only, every finding must carry current_value, a concrete `acceptance_test` "that a DIFFERENT agent could verify without re-auditing", max_fix_attempts, and must skip what algorithmic checks already caught. site-review-agent adds strategic alignment review; unclassifiable gaps become needs_content_planning items for content-gap-planner.
- **sources:** 066_audit_agent_definitions.sql; 084_site_review_agents.sql; 086_visual_design_auditor.sql; 071_content_gap_planner.sql
- **relations:** locks (locked_at exclusion); improvement-loop pass cap; fixers consume findings
- **verify-later:** write_audit_findings action; current audit prompts vs 084 text

### Section/component locking with timed expiry
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 115 header: "the project doc 004 v4 designed and docs 031_locks... approved (2026-05-08). Implemented now (Option A)... This migration is SCHEMA + BACKFILL only. The Go follow-on... lands as separate code changes."
- **what:** Locking is the improvement loop's termination and protection mechanism: verified/human-edited rows get locked_at set; auditors exclude locked rows (086). 115 adds lock_type + lock_expires_at to all four Pattern A lock-bearing tables (page_components, site_components, site_plan_directives, +1) in one transaction for coherence. Policy: admin/manual/checkpoint = permanent; deploy = timed +30d; visual-design-auditor / imagery-quality-auditor / adoption (new, faithful-first-pass) = timed +90d. Unlock predicate: `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())`. Go-side sweep of 11 callsites still pending at write time.
- **sources:** 115_locks.sql; 086_visual_design_auditor.sql
- **relations:** adoption faithfulness (FOCUS_adoption_faithfulness_via_locks.md); expired_review_locks discovery check (planned)
- **verify-later:** the 11 `locked_at IS NULL` callsites; CheckComponentLock extension; whether expiry sweep landed

### Blog/content planning agents (blog-content-planner, content-gap-planner, internal-linker)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 069 documents the full empty_blog loop; 070/071/101 definitions; content-gap-planner in 075's timeout list.
- **what:** LLM planners that turn detected content gaps into concrete work: blog-content-planner (needs_blog_posts) plans 3–4 posts from specs and reuses write_build_items to create pages + needs_content_page items + blog-index rerender; content-gap-planner (needs_content_planning) decides per gap between add-section (content_rewrite), new page, spec update (needs_spec_update), or wont_fix — "The LLM here is the PLANNER, not the auditor"; internal-linker finds pages that should contextually link to an orphaned sub-page and creates content_rewrite items for natural placements. 070 records the reuse-over-new-Go deliberation verbatim.
- **sources:** 069_blog_posts.sql; 070_blog_content_planner.sql; 071_content_gap_planner.sql; 101_internal_linker.sql
- **relations:** empty_blog / orphan page checks; page-build-handler executes their items; spec-updater
- **verify-later:** create_blog_posts action; empty_blog check

### spec-updater (mechanical site_specs merge from findings)
- **category:** site-spec-and-classifier
- **status-signal:** unknown
- **status-evidence:** 072 definition; no later patches in this unit.
- **what:** Handler for needs_spec_update items: applies {aspect, field, suggested_value} to site_specs with the WriteSiteSpecAction versioning pattern. No LLM. Description-only items complete as "needs human review". "The complexity is in the Go action, not in the workflow."
- **sources:** 072_spec_updater_agent.sql
- **relations:** content-gap-planner and audits emit its items; site_specs supersede-versioning
- **verify-later:** update_site_spec_from_item action

### component-creator (LLM component template generation) + CSS variable naming contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 093 definition (needs_new_component handler); 123 patches its prompt with the STRICT RULE on variable names, showing live iteration.
- **what:** Generates reusable HTML component templates from section-type descriptions, storing them in content_components with selection metadata (section_type etc.) for reuse. 123 hardens the prompt: only `--color-{role}` variables from the enumerated list are permitted; invented names like --primary-color "are WRONG and will produce broken output because they are undefined in every deployed stylesheet". Closes the loop with build-site-planner's roadmap rule: unknown roadmap section_types become needs_new_component items handled here.
- **sources:** 093_component_creator.sql; 123_component_creator.sql; 053_build_site_planner.sql
- **relations:** component contracts (003 docs); component-quality-auditor; component selector
- **verify-later:** store_generated_component action; component_selector.go; selection metadata columns

### component-quality-auditor (library health scoring)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 102 definition + one-shot backfill item scoring every existing component.
- **what:** Periodically scores content_components via compute_component_quality and creates needs_component_regeneration items for low scorers — keeps the shared component library healthy rather than only fixing per-site instances.
- **sources:** 102_component_quality_auditor.sql
- **relations:** component-creator (regeneration handler); improvement loop
- **verify-later:** compute_component_quality scoring criteria

### Tool pipeline: tool-suggester → tool-generator/tool-deployer → cross-linking
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** 062/062b/098 definitions and patches; 100 portfolio: "Six industry-specific tools deployed... Tool references automatically woven into 18 content pages. Full pipeline... runs autonomously."
- **what:** tool-suggester (evaluate_tools handler) uses LLM judgment over specs+pages to decide which interactive tools would genuinely help a site (not limited to library catalogue), creating add_tool items; tool-deployer forks a library tool to the site (component fork + tool page + page_component link, then normal render/deploy); tool-generator creates new tool HTML from brand context (and since 131 writes a travelling PLAN); 098 adds cross-linking — suggestions carry related_pages, and create_tool_cross_link_items generates content_rewrite items so page-build-handler weaves tool references into existing copy. missing_tools discovery check auto-seeds add_tool items.
- **sources:** 062_tool_suggester_and_improver.sql; 062b_tool_deployer_and_generator_agent.sql; 098_tool_suggester_cross_linking.sql; 061_tool_deployer_and_discovery_agent.sql
- **relations:** tool-library; tool acceptance tiers; travelling docs
- **verify-later:** deploy_tool_to_site action; create_tool_cross_link_items

### Tool quality tiers: tool-auditor (Tier 2 LLM review), tool-improver, acceptance checks (Tier 2 static) and tool-acceptance-agent (Tier 4 browser runs)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 088 (tool-auditor); 142 enables tool_acceptance check (2026-07-10, doc_notes entry "unit tests... green"); 145 inserts tool-acceptance-agent; 146 makes Tier 4 continuous via tool_acceptance_due sweep.
- **what:** Layered tool verification. Tier 1: check_tool_health structural checks. Tier 2 (LLM): tool-auditor reads full HTML/CSS/JS and reasons through logic/mobile/UX/accessibility, creating improve_tool or needs_human_review items. Tier 2 (static): check_tool_acceptance asserts the PLAN's criteria fence against the deployed page under the ANCHOR RULE ("validate a selector's leftmost id/class token, never the whole path; confirm, never refute; -EDIT ids skipped"). Tier 4: tool-acceptance-agent drives the deployed tool in headless Chromium via the browser-runner adapter against PLAN criteria — "the tier that turns 'deployed' into 'works'" — pass → acceptance-run note; fail → acceptance-fail note + one improve_tool item carrying criteria as acceptance_test. tool-improver executes improve_tool fixes. 7-day cooldowns; cancelled items excluded from cooldown (146).
- **sources:** 088_tool_auditor_agent.sql; 142_enable_tool_acceptance_check.sql; 145_tool_acceptance_agent.sql; 146_enable_tool_acceptance_due.sql; 062_tool_suggester_and_improver.sql
- **relations:** travelling PLAN criteria fences; design-discovery-agent hosts the checks; browser-runner adapter
- **verify-later:** request_browser_run / judge_acceptance_results actions; check_tool_acceptance.go anchor rule; browser-runner adapter deployment

### Acceptance-criteria honesty: invented selectors and inline-delivery decisions
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 136 (2026-07-09) repairs the first machine-written PLAN's invented ids (#xpTableBody→#tableWrap tbody, #statsStrip→#statRow); 143/144 (2026-07-10) "PLANs surrender to delivered reality" — asset extraction "was designed but never built", so criteria drop asset_loads and the composer prompt is corrected.
- **what:** Two recurring failure classes in machine-written acceptance criteria, and their durable remedies: (1) composers invent selectors they ASSERT on even while obeying never-invent for controls they ACT on — remedy is Tier-2 static validation of criteria selectors against html_template (anchor rule), not sterner prompts; (2) criteria must describe what the system DELIVERS, not aspirations — the /tools/assets/<fn>.js extraction path was never built, all JS ships inline, so PLANs and the composer prompt were superseded to inline delivery ("born honest"). Also note the abandoned mechanism: Path-1 tool asset extraction on rerender.
- **sources:** 136_supersede_xp_curve_plan_selectors.sql; 143_supersede_plans_inline_delivery.sql; 144_composer_inline_delivery.sql; 113_site_asset_renderer.sql (the extraction design it contradicts)
- **relations:** travelling docs supersede pattern; tool acceptance tiers
- **verify-later:** whether asset extraction ever ships (would trigger forward supersede)

### tool-recreation-handler (recreate interactive tools from crawled source)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 099 definition; live-run evidence in 138 (run e1018366 recreated bugs faithfully → prompt fix) and 132/137 (note-writing wired, subject corrected).
- **what:** Two-stage recreation of JS-heavy pages during site adoption: analyze_tool (LLM functional spec from source + context) then recreate_tool (Opus generates working replacement HTML/CSS/JS), with completeness/truncation checks, validation, save/deploy. 138 adds the "Mandatory Behaviour Requirements" prompt section rendered from spec.interactive_features which OVERRIDES the original source — fixing the observed failure where explicit spec fixes were buried in analysis JSON and Opus faithfully recreated the original bugs.
- **sources:** 099_tool_recreation_handler.sql; 138_recreate_tool_carries_spec_features.sql; 137_recreation_spec_and_note_subject.sql
- **relations:** site-adoption-agent creates its items; tool acceptance verifies results
- **verify-later:** current recreate_tool prompt; spec.interactive_features producers

### Site adoption pipeline (site-adoption-agent + wrapper orchestrator)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 091 definition; 104 adds the wrapper (2026, "Pattern copied verbatim from med-export-orchestrator"); 115 adds the 'adoption' lock source for "faithful first pass".
- **what:** Adopts an existing live site: firecrawl_crawl via the webscrape adapter returns per-page markdown; an LLM analyze step classifies pages and extracts identity/design/content structure into a JSON plan; apply_adoption_plan creates site_specs, page records, and work items to recreate the site in-platform. 104 wraps it in a spawn→call orchestrator so the long crawl runs in its own Job pod with clean correlation logs.
- **sources:** 091_site_adoption_agent.sql; 104_site_adoption_orchestrator.sql; 115_locks.sql
- **relations:** page-content-writer Recreate Mode; tool-recreation-handler; adoption locks
- **verify-later:** apply_adoption_plan action; adoption directive writer (pending per 115)

### News feed pipeline (feed-ingester, content-feed-orchestrator, feed-triage, content-feed-trigger, latest-news component)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 087–090 definitions; 100 portfolio: "Four source types operational. Credibility scoring... Six-hour refresh cycles... Live news sections deployed on production sites."
- **what:** Per-site news: content-feed-trigger is a 6-hour heartbeat that finds sites whose classification spec recommends a news feed (content_features.news_feed.recommended) and needing refresh, dispatching content-feed-orchestrator per site; feed-ingester fetches one source (RSS / news search / LLM news / scrape, routed by source_type) into content_feed_items; feed-triage (initially a stub) scores relevance/credibility; the latest-news content component is data-driven — rendered by the render_news_section Go action, not the LLM writer — with CSS from theme variables; 113's redesign migrated news components to contract-003 without regex (split_part/position/substring surgery).
- **sources:** 087_feeds_triage_ingester_orchestrator_etc.sql; 089_latest_news.sql; 090_b_content_feed_trigger.sql; 090_content_feed_orchestrator.sql; 113_site_asset_renderer.sql
- **relations:** content_sources/content_feed_items tables; scheduler-and-tasks; site_specs classification aspect
- **verify-later:** feed-triage real implementation vs stub; render_news_section action

### site-asset-renderer (deterministic /assets/js/snippets.js)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** 113 INSERT with verification queries; description "Deterministic — no LLM. Triggered when js_snippets or component set changes".
- **what:** Renders a site's shared JS snippet bundle (e.g. relative-time expansion for news feeds) from the js_snippets table and commits it to git; components load it via a single `<script src="/assets/js/snippets.js">` injected into templates. Establishes the site-level shared-asset mechanism distinct from per-tool inline JS.
- **sources:** 113_site_asset_renderer.sql
- **relations:** js_snippets table; latest-news component; contrasts with the never-built per-tool asset extraction (143)
- **verify-later:** render path and trigger wiring for snippets.js

### Companies House enrichment chain (business-intel / ch-* agents)
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** 077–083 sequential build-out with scheduled tasks; 100 portfolio: "Thousands of veterinary practices collected, verified against Companies House records, and enriched with financial data."
- **what:** Multi-stage enrichment on the business-intel pod: ch-collector bulk-mirrors all SIC 75000 companies into ch_vet_companies (paginated, rate-limited); ch-matcher matches verified businesses against the mirror by postcode + name similarity (pure SQL/Go scoring, threshold 0.40, no API); ch-llm-reviewer classifies ambiguous matches (Haiku, 15 pairs/batch) as confirmed/rejected/uncertain; ch-detail-fetcher pulls profile/officers/PSC for confirmed matches and derives succession-risk signals; ch-company-scraper regex-extracts registration numbers from business website footers (generic across verticals); ch-accounts-fetcher parses filed iXBRL accounts into financial columns (net assets, turnover, employees). ch-enricher (077, renamed business-intel) was the original combined agent.
- **sources:** 077_business_intel_companies_house.sql; 079_companies_house_ch_matcher.sql; 080_companies_house_ch_llm_reviewer.sql; 082_company_number_scraper_ch_company_scraper.sql; 083_companies_house_ch_accounts_fetcher.sql
- **relations:** vet vertical pipeline (verified businesses input); scheduled_tasks entries per agent
- **verify-later:** ch_* actions; match-rate stats views; scheduled task cadence

### Vet vertical data pipeline (area sweep, batch processor, practice verifier, med pricing)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** Tuning migrations show live operation (063 "up the max iterations to clear the backlog" → 1700; 063b prompt refined to extract registration_number).
- **what:** The veterinary vertical's collection stack: area-sweep-orchestrator loads un-swept UK postcode districts and dispatches area-sweep-discoverer per district (web search → discovery candidates for unknown businesses); vet-batch-processor works candidate batches; vet-practice-verifier web-searches each business (postcode/town fallback query template) and LLM-extracts/reconciles structured practice data including Companies House number. Med pricing: med-url-discoverer scrapes retailer category pages for product URLs, med-url-mapper uses Firecrawl /map site-wide, med-price-collector scrapes prices; each has a thin spawn→call orchestrator wrapper and scheduled task.
- **sources:** 037_area_sweep_discoverer.sql; 038_area_sweep_orchestrator.sql; 063_vet_batch_processor.sql; 063b_vet_practice_verifier.sql; 092_vet_med_pricing_agent.sql; 095_vet_med_firecrawl_url_agent.sql; 096_vet_med_url_discover_orchestrator.sql
- **relations:** companies-house chain consumes verified businesses; business-intel pod/topics
- **verify-later:** search_areas / businesses tables; scheduled task states

### Wrapper-orchestrator pattern ("spawns a temporary pod to do X")
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Named as the canonical pattern in 104 ("Pattern copied verbatim from med-export-orchestrator..."), 121 ("mirror the LIVE med-export-orchestrator / med-json-exporter pair verbatim — the canonical scheduler-triggered wrapper + task-worker"), 122 (dev guide "does this agent need a wrapper?" test).
- **what:** Convention: substantive in-chassis work (long LLM loops, crawls, collections) must not run in shared generic pods; instead a thin orchestrator (spawn_agent → call_agent → complete) creates a dedicated K8s Job pod for the worker, giving clean per-correlation logs, isolation, and idle-timeout cleanup. Spawn-before-call ordering is required for target_role lookups (109/111/112).
- **sources:** 096_vet_med_url_discover_orchestrator.sql; 104_site_adoption_orchestrator.sql; 121_intent_collector_agents.sql; 122_diagnose_agents.sql
- **relations:** idle_timeout_seconds; scheduler-and-tasks; K8s Job lifecycle
- **verify-later:** dev guide §wrapper test; spawn_actions.go

### idle_timeout_seconds (Job pod auto-exit)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 075 ALTER TABLE + fleet-wide backfill (180s) with rationale ("timer resets on every message... multi-step workflow... stays alive as long as responses keep arriving").
- **what:** Column on agent_definitions controlling how long a spawned Job pod waits with no messages before exiting cleanly (0 = no timeout for Deployment agents). Paired with TTLSecondsAfterFinished for cleanup. The 075 list doubles as a census of the then-live spawnable fleet.
- **sources:** 075_various_timeout_column.sql
- **relations:** wrapper pattern; K8s cleanup; debugging (timeouts)
- **verify-later:** chassis idle-timer implementation

### LLM model governance: aliases, per-step model choice, llm_call_log
- **category:** llm-quality-testing
- **status-signal:** deployed
- **status-evidence:** v2/027 regex-replaces all dated model names with aliases ("only the alias resolver in code needs updating"); 040 upgrades planners to sonnet with rationale and creates llm_call_log.
- **what:** Conventions for model management across ~90 agent definitions: model aliases (claude-sonnet-4-5 not dated strings) resolved in code; deliberate per-step model tiering (haiku for cheap classification, sonnet for high-leverage planning, opus for plan_site and tool recreation); llm_call_log capturing calls for cost analysis and training data. 067 (filename: "not_yet_implemented") prepared extended-thinking budget_tokens for classifier/planner gated on a Go patch.
- **sources:** sql_for_agents_v2/027_replace_claude_model_names.sql; 040_optimise_which_llms.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** finetuning flywheel (llm_call_log flywheel columns in 085); ai_endpoint_health
- **verify-later:** alias resolver; whether extended thinking was ever enabled

### ai_endpoint_health (GPU/model availability gating)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 085 creates table + seeds claude/cpu-ollama/gpu-ollama endpoints; "Checked by claim_work_item before claiming."
- **what:** Health registry for AI endpoints: healthy → work items flow, unhealthy → items wait. Active mode (scheduler pings, per-endpoint interval and ping path incl. 'claude_ping') and reactive mode (failure-driven). Integrates model availability into the dispatch loop's claim decision. Part B adds flywheel columns to llm_call_log (work-item link, prompt variants, verticals, RAG usage).
- **sources:** 085_ai_endpoint_health_checker.sql
- **relations:** build-dispatch-loop claim; Ollama/GPU infrastructure; finetuning flywheel
- **verify-later:** claim_work_item health check; scheduler ping task

### RAG knowledge base (shared pgvector store) and rag_index/rag_lookup
- **category:** NEW:rag-retrieval
- **status-signal:** deployed
- **status-evidence:** 041 creates knowledge_base (vector(768), nomic-embed-text, content_hash dedup); 105 rag-test-agent verifies chassis registration; 141 finally lands first tool_docs rows after the chunk-loop saga.
- **what:** Shared (not per-agent) embedded knowledge store for scraped exemplar sites, research, curated industry info and component usage patterns, queryable by any content-creating agent. Collections partition use-cases (industry_sites, research, components, tool_docs, flywheel_b_chassis_test). Embedding-model column tracks provenance; changing dimensions requires column ALTER + reindex.
- **sources:** 041_rag_knowledge_base.sql; 105_rag_test_agent.sql; 141_reenable_index_plan_after_chunk_fix.sql
- **relations:** travelling docs (doc_plans indexed into tool_docs); rag_actions.go; code-indexer (separate code_symbols store)
- **verify-later:** knowledge_base row counts per collection; rag_lookup consumers

### rag_index chunkContent OOM saga (bypass → reenable → rebypass → fix)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Four-migration arc 135/139/140/141 (2026-07-09/10) with root cause CONFIRMED in 140: "chunkContent() never terminated on content longer than chunk_size... ~2Gi of duplicate chunks in seconds. Both chassis OOMKills were this loop."
- **what:** A model incident record: tool creation hung/OOMed at index_plan. First hypothesis (no embedding deadline) produced 135's bypass + a hygiene deadline (139); reoccurrence disproved it; the real bug was a non-terminating chunk loop (start = end - overlap re-entering forever), fixed in Go with regression tests, then re-enabled by 141. Durable practices demonstrated: reversible SQL bypasses that keep truth in Postgres (write_plan) while sacrificing only derived indexing; explicit preconditions in re-enable migrations; superseding one's own root-cause statements on record.
- **sources:** 135_bypass_index_plan_until_embed_timeout.sql; 139_reenable_index_plan.sql; 140_rebypass_index_plan_chunk_loop.sql; 141_reenable_index_plan_after_chunk_fix.sql
- **relations:** rag knowledge base; travelling docs pipeline notes; 016b debugging lessons
- **verify-later:** rag_actions_chunk_test.go presence; deployed image ≥ fix commit

### Finetuning flywheel Phase 5: training kickoff orchestration (model-trainer, training-data-preparer, gpu-provisioner, training-launcher)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** 109/110 real; 111/112 explicitly STUBs "to unblock end-to-end testing... will be replaced by a real implementation in a future migration"; 116/117 monitor real running instances, implying provisioning later became real.
- **what:** model-trainer owns the KICKOFF phase: training-data-preparer exports a training_exports snapshot as JSONL to S3 and INSERTs the model_lifecycle.training_runs row (pending); gpu-provisioner calls Thunder Compute API for an A100, stores the SSH key as a k8s secret; training-launcher SCPs scripts/dataset and nohup-launches training, returning the pid. The workflow exits immediately — completion is deliberately handled by a separate scheduled monitor so no orchestration holds open for ~9 hours. Full hyperparameter set captured for reproducibility.
- **sources:** 109_model_trainer_orchestrator.sql; 110_training_data_preparer.sql; 111_gpu_provisioner_thunder.sql; 112_training_launcher.sql
- **relations:** training-data-exporter (106) upstream; thunder monitor/reaper downstream; thunder-adapter
- **verify-later:** real gpu-provisioner/training-launcher implementations vs stubs

### Thunder instance lifecycle: reaper + training monitor (orchestrator/worker)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 114 (reaper, every 15 min, idempotent decommission); 116/117 with verified coordinator internals and the insert-DISABLED-until-actions-deploy discipline.
- **what:** Cost/safety controls for rented GPUs: thunder-reaper decommissions instances past max_uptime_hours (one per tick, pre_query LIMIT 1); thunder-training-monitor orchestrator finds every running training instance each tick and spawns a per-instance worker that probes via SSH, classifies (alive / unreachable-streak / done_ok / done_fail), reconciles training_runs and decommissions. 117 records WHY orchestrator-with-loop beats the reaper's scheduler-pre_query shape (must visit every instance, not just the top row) and why the loop must stay sequential (topic reuse safety).
- **sources:** 114_thunder_reaper.sql; 116_thunder_training_monitor_worker.sql; 117_thunder_training_monitor_orchestrator.sql
- **relations:** scheduler-and-tasks (pre_query dispatch patterns); thunder adapter; model-trainer
- **verify-later:** scheduled_tasks rows enabled; probe/classify actions

### training-data-exporter (llm_call_log → ChatML JSONL)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 106 definition with concrete kubectl retrieval instructions and input payload example.
- **what:** Deterministic single-action agent exporting successful LLM calls from llm_call_log as NDJSON training data in ChatML + metadata format, filterable by agent_type/step/model, with fenced-output and strict-JSON options.
- **sources:** 106_training_data_exporter.sql; 040_optimise_which_llms.sql (llm_call_log)
- **relations:** training-data-preparer consumes exports; flywheel columns from 085
- **verify-later:** training_data_export action; training_exports schema

### Intent-event collection from VM-hosted backend sites (P4)
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** 119 "Pattern... mirrors... the thunder-training-monitor convention of INSERTING DISABLED until the action is deployed"; tables created; agents mirror the med-export pair.
- **what:** Off-box collection of visitor intent: VM-hosted sites expose key-gated GET /events (NDJSON) and /stats; a scheduled intent-collection-orchestrator/intent-collector pair pulls events into intent_events (engine_event_id UNIQUE gives structural idempotency — safe overlapping `since` windows, checkpoint derived from max(event_created_at)) and cumulative visit counters into intent_site_stats (one row per host) so ranking can compute true events-per-1k-visits. kind constrained to search/categories/freetext.
- **sources:** 119_intent_events_for_vms.sql; 120_intent_site_stats.sql; 121_intent_collector_agents.sql
- **relations:** intent capture engine on the VM side (vonc/backend sites); scheduler pre_query dispatch
- **verify-later:** collector action deployment; scheduled task enabled flag; ranking queries

### Diagnosis loop agents (diagnose-orchestrator / diagnose-agent)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** 122 seeds both as 'experimental' ("until the real-bug evaluation gate passes; promote to 'active' after"); 126–129 wire persistence and subject threading; incidents in 127/128 show live runs 2026-07-06/10.
- **what:** Read-only diagnosis: hypothesise → gather scoped evidence (code + runtime) → cite-or-abstain verdict → re-scope by following evidence; emits a diagnosis + evidence trail for a human, never changes code. Loop CONTROL lives in the Go engine (diagnose_run), not workflow conditionals; gather steps stay explicit for log visibility. Wrapper-mandated (substantive in-chassis LLM work). Runtime evidence is an optional bundle tier — error routing makes anchorless (code-only) runs survive.
- **sources:** 122_diagnose_agents.sql; 126_wire_persist_diagnosis_note.sql; 127_diagnose_load_runtime_error_step.sql; 129_wire_diagnosis_subject_threading.sql
- **relations:** code-indexer supplies code_symbols retrieval; travelling docs receive diagnosis notes; docs019 diagnosis programme
- **verify-later:** diagnose_run engine; promotion to active; evaluation gate results

### error_step-inside-config routing rule
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 128 "Durable rules this incident banked (016b): error_step lives INSIDE step.Config and must name an EXISTING step; derive convergence targets from the step's own next_step, never guess"; effect verified live 2026-07-10; 131/132 retro-move ten inert step-level error_steps into config.
- **what:** Chassis workflow convention discovered through failures: the coordinator reads step.Config["error_step"] only — step-LEVEL error_step keys are silently ignored; a routing target that names a non-existent step fails the whole workflow. Correct-while-touching policy migrates old inert keys whenever a workflow is edited.
- **sources:** 128_fix_load_runtime_error_step_target.sql; 127_diagnose_load_runtime_error_step.sql; 131_tool_generator_plan_writing.sql; 132_fix_agents_note_writing.sql
- **relations:** 016b debugging heuristics; template field-path rule (134)
- **verify-later:** coordinator error-routing code

### Prompt-template field-path rule (text vs json output shapes)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 134 (2026-07-09) "THE RULE (proven by this run, not assumed)": text-format steps pass the bare string to downstream templates ({{.generated_html}}, not .result); json-format steps pass a map (use `| toJSON`); action-config field paths are a DIFFERENT resolver and keep .result.
- **what:** A durable rendering contract distinguishing three resolvers: Go template rendering of LLM text results (bare string), of JSON results (map, dump with toJSON rather than guessing keys), and action-config field paths (keep .result suffix). Applied as one blocker fix plus three pre-emptive corrections of the same bug class.
- **sources:** 134_fix_prompt_template_field_paths.sql
- **relations:** call metadata/response convention; error containment via config.error_step (docs steps can never fail tool creation, 131)
- **verify-later:** ExtractActionInputs / template renderer code

### code-indexer (repo → code_symbols for the analyser)
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** 118 marked DRAFT/experimental but with a checked label convention "[checked 2026-06-11: composition implemented in IndexCodeSymbolsAction]".
- **what:** Orchestrator that asks the analyser adapter to parse a repo at a ref into symbols, then index_code_symbols upserts them (embedding changed symbols, pruning absent ones). repo label is composed as "owner/repo" from the analyser reply so labels always match what was fetched; retrieval side is lookup_code_symbols used by diagnosis agents. Non-git corpora may override repo (e.g. 'domain:kruste.com').
- **sources:** 118_code_indexer_for_analyser.sql
- **relations:** diagnose-agent evidence gathering; analyser adapter; docs019 contextkit
- **verify-later:** code_symbols table; agent status live

### Travelling docs: doc_plans / doc_notes with automated writing
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** 125 tables (design "PLAN_travelling_docs.md rev 4. Truth in Postgres; knowledge_base is the derived RAG index"); 130 first hand-seeded PLAN; 131 tool-generator writes PLANs automatically after save_tool; 132 three fix agents append NOTES; first machine-written PLAN 2026-07-09 (136).
- **what:** Per-subject living documentation keyed by (subject_type, subject_key) — tool → content_components.function, pipeline → site_work_items.pipeline. doc_plans holds versioned intent (supersede pattern, one is_current row, may embed a ```criteria fence consumed by acceptance tiers); doc_notes is append-only history written by agents on every fix (Observed/Root cause/Fix/Verified/Categories format) and by diagnosis persist_note. Doc-writing steps always carry config.error_step so documentation failure can never fail the substantive work.
- **sources:** 125_doc_plans_and_notes.sql; 130_pilot_plan_tool_archetype_taster_quiz.sql; 131_tool_generator_plan_writing.sql; 132_fix_agents_note_writing.sql
- **relations:** tool acceptance criteria fences; diagnosis subject threading (129); rag tool_docs indexing
- **verify-later:** write_doc_plan / append_doc_note / persist_diagnosis_note actions; doc_notes row growth

### Migration discipline: schema_migrations ledger, snapshot_agent, migration_backups
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 124 "APPLIED 2026-07-10... From this file onward, every numbered file... is applied via scripts/migration/run-migrations.sh... Files 001–123... are HISTORY, not pending work; the runner's baseline (124) excludes them." Backfill rows include a lost-file reconstruction (128).
- **what:** The operational regime for this directory: schema_migrations records WHAT ran WHEN (filename PK, checksum, notes); snapshot_agent(type, reason) is the standing rule opening every agent-updating transaction (MVCC before-image); migration_backups holds manual before-values; 107's backup preamble adds the no-DROP rule ("The collision IS the safety net"). Workflow-altering migrations must leave a pipeline doc_note (runbook §3, seen in 141/142/144/146).
- **sources:** 124_schema_migrations.sql; 107_image_build_handler.sql; 131_tool_generator_plan_writing.sql; 128_fix_load_runtime_error_step_target.sql
- **relations:** travelling docs; versioned agent_definitions (UNIQUE on type+version, 121)
- **verify-later:** run-migrations.sh; snapshot_agent function; schema_migrations contents

### site-design-planner spec aspects (navigation / layout / resolved_composition)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 103 is "Deliverable 2: Spec schemas... pre validation" — documents shapes and creates best-effort validation functions, explicitly not table constraints; reader lists name live actions.
- **what:** Defines three site_specs aspects the site-design-planner writes, separated by reader: navigation (nav architecture, items, CTA, mobile pattern → populate_nav_tables, InjectHeader, GetNavItems), layout (page-level layout, header/footer style → AssembleMultipageSiteAction, templates), resolved_composition (machine-readable pointers to palette/layout/typography rows + reasoning → render_css_from_spec, webdesign-agent, audit agents). Validation functions run at write time; site_specs stays open JSONB.
- **sources:** 103_site_design_planner.sql
- **relations:** design-composition docs 025/026/027; webdesign-agent; nav-updater
- **verify-later:** site-design-planner agent existence and writers of these aspects

### Portfolio/use-case spec seeds (ai-agent-orchestration.com)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** 100 INSERTs site_specs 'portfolio' aspect with five dated case studies claiming operational metrics ("Six production sites deployed and self-maintaining... under 4 hours" domain-to-live).
- **what:** Marketing-facing data seed whose case studies double as a platform capability inventory circa file-100: autonomous multi-site pipeline (30+ agents), tool generation + cross-linking, vet data platform, news aggregation with credibility scoring, and the orchestration layer itself (Kafka/Postgres/K8s, hot-swappable SQL workflow definitions, fuel budgets). Useful as documentary status evidence for many other concepts, not ground truth.
- **sources:** 100_portfolio_use_cases_etc.sql
- **relations:** nearly every pipeline concept above; site-case-studies
- **verify-later:** claims vs stage-2 code/DB verification

### client_system schema for agent instances
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 073 drops wrong tables and recreates agent_instances/agent_spawn_history "matching what create_client_schema and spawn_agent expect".
- **what:** Per-client schema tables backing spawn_agent: agent_instances (template_id → agent_definitions, project FK, config) and agent_spawn_history (parent/spawned lineage). Documents the column contract spawn_agent expects.
- **sources:** 073_create_new_client_schema.sql
- **relations:** spawn_agent action; create_client_schema; multitenancy/client schemas (docs 011)
- **verify-later:** create_client_schema function vs 073 shapes

## Proposed NEW categories
- **NEW:build-pipeline** — the site_work_items work-item build pipeline and its builder/handler agents (pageflow-builder, site-work-orchestrator, dispatch loop, page-build-handler, page-content-writer, page-rebuild). Distinct from improvement-loop (post-build) and site-plan-and-reconciler (planning domain); large enough to back a council agent.
- **NEW:rag-retrieval** — shared knowledge_base pgvector store, rag_index/rag_lookup actions, collections (tool_docs, industry_sites), embedding-model management. Not covered by model-infrastructure (endpoints/GPUs) or documentation-system.
