# =====================================================================
# FOCUSED AGENT CHASSIS CONTEXT (trimmed ~55% from full)
# Keeps: orchestration core, key actions, discovery, kafka, configs
# Drops: component_library, multipage, business_intel, companies_house,
#        html_actions, section_editor, generate_image, fix_forced_text,
#        storage, loop_actions, migration READMEs, migration SQL,
#        manual seeding, tests, makefile, topicflow.png
# =====================================================================
agent-chassis-focused)
MODULE_DIRS=(
# Agent core
"platform/agentbase/"
"platform/messaging/"

# Orchestration core (coordinator, state, helpers, types)
"platform/orchestration/types/"
"platform/orchestration/input_contracts/"
"platform/orchestration/actioncheck/"
"platform/orchestration/datahelpers/"

# Discovery checks
"platform/orchestration/actions/discovery_checks/"

# Kafka
"platform/kafka/"

# AI service
"platform/aiservice/"

# Database (just the Go code, not migrations)
# Storage interface only
)
MODULE_FILES=(
# Orchestration core files (not in a subdirectory)
"platform/orchestration/coordinator.go"
"platform/orchestration/state.go"
"platform/orchestration/helpers.go"
"platform/orchestration/agent_error_log.go"
"platform/orchestration/monitoring.go"
"platform/orchestration/loop_expansion_handler.go"
"platform/orchestration/loop_error_handler.go"
"platform/orchestration/coordinator_hitl_extraction_helper.go"
"platform/orchestration/ui_notifications.go"

# Action registry
"platform/orchestration/actions/registry.go"
"platform/orchestration/actions/types.go"
"platform/orchestration/actions/helpers.go"

# Core actions (spawning, calling, workflow)
"platform/orchestration/actions/spawn_actions.go"
"platform/orchestration/actions/spawn_group.go"
"platform/orchestration/actions/call_agent.go"
"platform/orchestration/actions/workflow_actions.go"
"platform/orchestration/actions/basic_actions.go"
"platform/orchestration/actions/conditional_branch_action.go"
"platform/orchestration/actions/generic_actions.go"
"platform/orchestration/actions/await_response.go"

# Site building actions
"platform/orchestration/actions/v3_site_actions.go"
"platform/orchestration/actions/site_db_actions.go"
"platform/orchestration/actions/site_spec_actions.go"
"platform/orchestration/actions/maintenance_actions.go"
"platform/orchestration/actions/validate_page_content.go"
"platform/orchestration/actions/render_site_components_action.go"
"platform/orchestration/actions/rerender_pages_actions.go"
"platform/orchestration/actions/rerender_single_page_action.go"
"platform/orchestration/actions/save_page_sections_action.go"
"platform/orchestration/actions/render_css_from_spec_action.go"
"platform/orchestration/actions/plan_sections_action.go"

# Work item + dispatch actions
"platform/orchestration/actions/load_work_item_actions.go"
"platform/orchestration/actions/claim_work_item_action.go"
"platform/orchestration/actions/create_work_item_action.go"
"platform/orchestration/actions/dispatch_actions.go"
"platform/orchestration/actions/dispatch_area_discoverers.go"
"platform/orchestration/actions/triage_detect_items_action.go"
"platform/orchestration/actions/seed_build_queue_action.go"

# Discovery + audit actions
"platform/orchestration/actions/discovery_actions.go"
"platform/orchestration/actions/discovery_checks.go"
"platform/orchestration/actions/write_audit_findings_action.go"

# Fix actions (component template, nav, colours)
"platform/orchestration/actions/fix_component_template_action.go"
"platform/orchestration/actions/fix_harcoded_colours_action.go"
"platform/orchestration/actions/fix_nav_link_templates_action.go"
"platform/orchestration/actions/update_component_html_action.go"
"platform/orchestration/actions/update_site_spec_from_item_action.go"

# Blog + content actions
"platform/orchestration/actions/create_blog_posts_action.go"
"platform/orchestration/actions/apply_gap_plan_action.go"
"platform/orchestration/actions/create_tool_component_action.go"

# LLM + AI actions
"platform/orchestration/actions/ai_actions.go"

# Git deploy
"platform/orchestration/actions/git_deployer_actions.go"

# Navigation
"platform/orchestration/actions/nav_tables.go"
"platform/orchestration/actions/populate_nav_tables_action.go"
"platform/orchestration/actions/link_constraints.go"
"platform/orchestration/actions/link_site_components_action.go"
"platform/orchestration/actions/prepare_link_context_action.go"

# Page loading
"platform/orchestration/actions/load_page_record_action.go"
"platform/orchestration/actions/load_site_pages_action.go"
"platform/orchestration/actions/get_pages_to_build_actions.go"
"platform/orchestration/actions/get_pages_for_rerender_action.go"

# Database actions
"platform/orchestration/actions/database_actions.go"
"platform/orchestration/actions/query_agent_definitions_actions.go"

# Design actions
"platform/orchestration/actions/design_actions.go"
"platform/orchestration/actions/assemble_from_library.go"
"platform/orchestration/actions/load_component_library_actions.go"
"platform/orchestration/actions/deploy_tool_action.go"
"platform/orchestration/actions/deploy_image_asset_action.go"
"platform/orchestration/actions/validate_dark_sections.go"
"platform/orchestration/actions/component_validation.go"
"platform/orchestration/actions/save_component_history_action.go"

# HITL
"platform/orchestration/actions/hitl_actions.go"
"platform/orchestration/actions/hitl_persistence.go"
"platform/orchestration/actions/hitl_request_human_input.go"
"platform/orchestration/actions/fetch_agent_questionnaire.go"

# Research + web
"platform/orchestration/actions/research_actions.go"
"platform/orchestration/actions/web_search_action.go"
"platform/orchestration/actions/webscrape_actions.go"
"platform/orchestration/actions/batch_webscrape_action.go"

# Sync + transform
"platform/orchestration/actions/sync_site_identity_action.go"
"platform/orchestration/actions/transform_actions.go"
"platform/orchestration/actions/entity_state_actions.go"

# Vet pipeline
"platform/orchestration/actions/promote_candidates_action.go"
"platform/orchestration/actions/scan_discovery_candidates.go"
"platform/orchestration/actions/load_pending_verifications_action.go"
"platform/orchestration/actions/load_unswept_areas_action.go"
"platform/orchestration/actions/dispatch_verifiers.go"
"platform/orchestration/actions/process_area_vet_sweep.go"
"platform/orchestration/actions/prepare_extraction_context.go"

# Database Go code
"platform/database/postgres.go"
"platform/database/mysql.go"

# Storage interface
"platform/storage/interface.go"
"platform/storage/s3.go"

# Agent implementations
"cmd/agent-chassis/main.go"
"internal/agents/contentcreator/agent.go"
"internal/agents/reasoning/agent.go"

# Configs
"configs/agent-chassis.yaml"
"configs/core-manager.yaml"
"configs/git-adapter.yaml"

# Deployment
"deployments/kustomize/services/agent-chassis/overlays/production/uk_001/patch-deployment.yaml"
"deployments/kustomize/services/agent-chassis/overlays/production/uk_001/kustomization.yaml"
"deployments/kustomize/services/agent-chassis/base/deployment.yaml"
)
;;