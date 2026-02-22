-- 023_site_work_orchestrator.sql
-- Agent definition for site-work-orchestrator
--
-- Uses actual agent_definitions column names:
--   default_config (not config), is_active (not enabled),
--   image_repository (not docker_image), image_tag (not docker_tag),
--   health_config (not health_check), input_contract (not input_spec),
--   output_contract (not output_spec), domain_tags (not semantic_tags)

INSERT INTO agent_definitions (
    type, display_name, description, category, default_config, is_active,
    capabilities, image_repository, image_tag,
    resources, topics, health_config, env_vars,
    delegation_preferences, status, domain_tags,
    briefing_questionnaire, input_contract, output_contract
) VALUES (
             'site-work-orchestrator',
             'Site Work Orchestrator',
             'Unified build/maintenance orchestrator. Builds sites from work items in site_work_items table. Processes items by priority, calling appropriate handler agents. Compatible with pageflow-builder''s planner and content writer.',
             'orchestrator',
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 1800,
                 "workflow": {
                     "start_step": "spawn_planner",
                     "steps": {

                         "spawn_planner": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "planner",
                                 "agent_type": "site-planner"
                             },
                             "next_step": "spawn_content_writer",
                             "description": "Spawn site planner agent",
                             "output_field": "planner_agent"
                         },
                         "spawn_content_writer": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "content_writer",
                                 "agent_type": "page-content-writer"
                             },
                             "next_step": "spawn_reviewer",
                             "description": "Spawn content writer agent",
                             "output_field": "content_writer_agent"
                         },
                         "spawn_reviewer": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "reviewer",
                                 "agent_type": "content-reviewer"
                             },
                             "next_step": "ensure_site_record",
                             "description": "Spawn content reviewer agent",
                             "output_field": "reviewer_agent"
                         },

                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {
                                 "store_brief_in_content_data": true
                             },
                             "next_step": "call_site_planner",
                             "description": "Create or update site record in database",
                             "output_field": "site_record"
                         },

                         "call_site_planner": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "site-planner",
                                 "target_role": "planner",
                                 "input_mapping": {
                                     "input_data": "input_data",
                                     "site_record": "site_record",
                                     "reviewed_brief": "input_data.reviewed_brief"
                                 },
                                 "timeout_seconds": 120
                             },
                             "next_step": "store_reviewed_brief",
                             "description": "Plan pages, select components, identify asset needs",
                             "output_field": "site_plan"
                         },

                         "store_reviewed_brief": {
                             "action": "update_site_content",
                             "config": {
                                 "merge": true,
                                 "content_field": "input_data.reviewed_brief",
                                 "site_id_field": "site_record.site_id"
                             },
                             "next_step": "store_site_plan",
                             "description": "Store the reviewed brief in sites.content_data",
                             "output_field": "brief_stored"
                         },
                         "store_site_plan": {
                             "action": "update_site_content",
                             "config": {
                                 "merge": true,
                                 "content_field": "site_plan",
                                 "site_id_field": "site_record.site_id"
                             },
                             "next_step": "sync_pages_to_db",
                             "description": "Store the site plan in sites.content_data",
                             "output_field": "content_stored"
                         },
                         "sync_pages_to_db": {
                             "action": "sync_pages_to_db",
                             "config": {
                                 "input_fields": ["site_record", "site_plan"]
                             },
                             "next_step": "write_build_items",
                             "description": "Create page records from site plan",
                             "output_field": "db_sync"
                         },

                         "write_build_items": {
                             "action": "write_build_items",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "site_plan": "site_plan"
                             },
                             "next_step": "populate_nav",
                             "description": "Convert site plan into work items in site_work_items table",
                             "output_field": "build_items_written"
                         },

                         "populate_nav": {
                             "action": "populate_nav_tables",
                             "config": {
                                 "input_fields": ["site_id"],
                                 "max_header_items": 8
                             },
                             "next_step": "check_assets_needed",
                             "description": "Populate navigation tables from page records",
                             "output_field": "nav_data"
                         },

                         "check_assets_needed": {
                             "action": "conditional",
                             "config": {
                                 "condition": "site_plan.needs_logo == true OR site_plan.needs_images == true OR site_plan.response.needs_logo == true OR site_plan.response.needs_images == true",
                                 "then_step": "spawn_image_generator",
                                 "else_step": "select_style_collection"
                             },
                             "description": "Check if logo or images need to be generated"
                         },
                         "spawn_image_generator": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "image_generator",
                                 "agent_type": "image-generator"
                             },
                             "next_step": "spawn_webdesign_agent",
                             "description": "Spawn image generator agent",
                             "output_field": "image_generator_info"
                         },
                         "spawn_webdesign_agent": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "webdesigner",
                                 "agent_type": "webdesign-agent"
                             },
                             "next_step": "generate_logo",
                             "description": "Spawn webdesign agent",
                             "output_field": "webdesign_agent"
                         },
                         "generate_logo": {
                             "action": "conditional",
                             "config": {
                                 "condition": "site_plan.needs_logo == true",
                                 "then_step": "call_logo_generation",
                                 "else_step": "check_hero_images"
                             },
                             "description": "Check if logo needs to be generated"
                         },
                         "call_logo_generation": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "image-generator",
                                 "target_role": "image_generator",
                                 "input_mapping": {
                                     "prompt": "site_plan.response.image_prompts.logo",
                                     "site_plan": "site_plan",
                                     "reviewed_brief": "input_data.reviewed_brief"
                                 },
                                 "output_mapping": {
                                     "prompt": "generate.response.prompt",
                                     "image_uri": "generate.response.image_uri",
                                     "image_url": "generate.response.image_url",
                                     "generated_at": "generate.response.generated_at"
                                 },
                                 "timeout_seconds": 120
                             },
                             "next_step": "store_logo_asset",
                             "description": "Generate logo using image-generator agent",
                             "output_field": "logo_result"
                         },
                         "store_logo_asset": {
                             "action": "store_asset",
                             "config": {
                                 "purpose": "logo",
                                 "asset_type": "logo",
                                 "data_field": "logo_result.image_url",
                                 "origin_type": "generated",
                                 "site_id_field": "site_record.site_id",
                                 "brand_asset_key": "logo.primary",
                                 "origin_prompt_field": "site_plan.image_prompts.logo",
                                 "update_site_brand_assets": true
                             },
                             "next_step": "deploy_logo_image",
                             "description": "Store generated logo",
                             "output_field": "logo_stored"
                         },
                         "deploy_logo_image": {
                             "action": "deploy_image_asset",
                             "config": {
                                 "purpose": "logo",
                                 "uri_field": "logo_result.image_uri",
                                 "domain_field": "site_record.domain"
                             },
                             "next_step": "check_hero_images",
                             "description": "Deploy logo image to git",
                             "output_field": "logo_deployed"
                         },
                         "check_hero_images": {
                             "action": "conditional",
                             "config": {
                                 "condition": "site_plan.needs_images == true OR site_plan.response.needs_images == true",
                                 "then_step": "generate_hero_image",
                                 "else_step": "select_style_collection"
                             },
                             "description": "Check if hero images need to be generated"
                         },
                         "generate_hero_image": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "image-generator",
                                 "target_role": "image_generator",
                                 "input_mapping": {
                                     "prompt": "site_plan.response.image_prompts.hero_home",
                                     "site_plan": "site_plan",
                                     "reviewed_brief": "input_data.reviewed_brief"
                                 },
                                 "output_mapping": {
                                     "prompt": "generate.response.prompt",
                                     "image_uri": "generate.response.image_uri",
                                     "image_url": "generate.response.image_url",
                                     "generated_at": "generate.response.generated_at"
                                 },
                                 "timeout_seconds": 120
                             },
                             "next_step": "store_hero_asset",
                             "description": "Generate hero image",
                             "output_field": "hero_result"
                         },
                         "store_hero_asset": {
                             "action": "store_asset",
                             "config": {
                                 "purpose": "hero",
                                 "asset_type": "image",
                                 "data_field": "hero_result.image_url",
                                 "origin_type": "generated",
                                 "site_id_field": "site_record.site_id",
                                 "brand_asset_key": "hero_images.home",
                                 "origin_prompt_field": "site_plan.image_prompts.hero_home",
                                 "update_site_brand_assets": true
                             },
                             "next_step": "deploy_hero_image",
                             "description": "Store generated hero image",
                             "output_field": "hero_stored"
                         },
                         "deploy_hero_image": {
                             "action": "deploy_image_asset",
                             "config": {
                                 "purpose": "hero",
                                 "uri_field": "hero_result.image_uri",
                                 "domain_field": "site_record.domain",
                                 "output_mapping": {
                                     "files": "response.data.files",
                                     "domain": "response.data.domain",
                                     "deployed": "response.data.success",
                                     "repo_url": "response.data.repo_url",
                                     "image_url": "response.data.file_path"
                                 }
                             },
                             "next_step": "select_style_collection",
                             "description": "Deploy hero image to git",
                             "output_field": "hero_deployed"
                         },

                         "select_style_collection": {
                             "action": "select_style_collection",
                             "config": {
                                 "style_from": "site_plan.style_collection",
                                 "site_id_field": "site_record.site_id",
                                 "fallback_by_domain": true
                             },
                             "next_step": "set_default_components",
                             "description": "Choose style collection based on site plan",
                             "output_field": "style_collection"
                         },
                         "set_default_components": {
                             "action": "update_site_defaults",
                             "config": {
                                 "defaults": {
                                     "head": "head-seo-standard",
                                     "footer_from": "style_collection.footer_component_name",
                                     "header_from": "style_collection.header_component_name"
                                 },
                                 "site_id_field": "site_record.site_id"
                             },
                             "next_step": "render_site_components",
                             "description": "Set default head/header/footer components",
                             "output_field": "defaults_set"
                         },
                         "render_site_components": {
                             "action": "render_site_components",
                             "config": {
                                 "slots": ["header", "footer", "head"],
                                 "domain_field": "site_record.domain",
                                 "site_id_field": "site_record.site_id",
                                 "force_rerender": false
                             },
                             "next_step": "load_work_items",
                             "description": "Render and store site-level components",
                             "output_field": "site_components_rendered"
                         },

                         "load_work_items": {
                             "action": "load_work_items",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "item_domain": "build",
                                 "handler_agent": "page-content-writer",
                                 "max_items": 20
                             },
                             "next_step": "check_has_items",
                             "description": "Load content page work items from queue",
                             "output_field": "work_items"
                         },
                         "check_has_items": {
                             "action": "conditional",
                             "config": {
                                 "condition": "work_items.has_items == true",
                                 "then_step": "build_items_loop",
                                 "else_step": "apply_site_design"
                             },
                             "description": "Check if there are work items to process"
                         },

                         "build_items_loop": {
                             "action": "loop",
                             "config": {
                                 "mode": "sequential",
                                 "items_field": "work_items.items",
                                 "item_variable": "current_item",
                                 "max_iterations": 20,
                                 "sub_workflow": {
                                     "start_step": "write_page_content",
                                     "steps": {
                                         "write_page_content": {
                                             "action": "call_agent",
                                             "config": {
                                                 "agent_type": "page-content-writer",
                                                 "target_role": "content_writer",
                                                 "input_mapping": {
                                                     "db_sync": "db_sync",
                                                     "hero_url?": "hero_url",
                                                     "logo_url?": "logo_url",
                                                     "site_plan": "site_plan",
                                                     "site_record": "site_record",
                                                     "current_page": "current_item.spec",
                                                     "reviewed_brief": "input_data.reviewed_brief",
                                                     "brand_logo_url?": "brand_logo_url",
                                                     "style_collection": "style_collection"
                                                 },
                                                 "timeout_seconds": 300
                                             },
                                             "next_step": "review_page_content",
                                             "description": "Write content for this page",
                                             "output_field": "page_content"
                                         },
                                         "review_page_content": {
                                             "action": "call_agent",
                                             "config": {
                                                 "agent_type": "content-reviewer",
                                                 "target_role": "reviewer",
                                                 "input_mapping": {
                                                     "site_record": "site_record",
                                                     "current_page": "current_item.spec",
                                                     "page_content": "page_content",
                                                     "reviewed_brief": "input_data.reviewed_brief"
                                                 },
                                                 "timeout_seconds": 3900
                                             },
                                             "next_step": "check_review_approved",
                                             "description": "Review page content (HITL or auto-eval)",
                                             "output_field": "reviewed_content"
                                         },
                                         "check_review_approved": {
                                             "action": "conditional",
                                             "config": {
                                                 "condition": "reviewed_content.review_result.approved == true OR reviewed_content.approved == true",
                                                 "then_step": "assemble_page",
                                                 "else_step": "fail_item"
                                             },
                                             "description": "Check if content was approved"
                                         },
                                         "assemble_page": {
                                             "action": "assemble_page",
                                             "config": {
                                                 "inject_head": false,
                                                 "content_field": "page_content.response.page_html",
                                                 "add_navigation": false
                                             },
                                             "next_step": "deploy_page",
                                             "description": "Assemble full page HTML from components",
                                             "output_field": "assembled_page"
                                         },
                                         "deploy_page": {
                                             "action": "git_commit",
                                             "config": {
                                                 "page_field": "current_item.spec",
                                                 "domain_field": "site_record.domain",
                                                 "content_field": "assembled_page.html"
                                             },
                                             "next_step": "save_sections",
                                             "description": "Commit page to git (triggers deploy via GitHub Action)",
                                             "output_field": "page_deployed"
                                         },
                                         "save_sections": {
                                             "action": "save_page_sections",
                                             "config": {
                                                 "html_field": "assembled_page.html",
                                                 "site_id_field": "site_record.site_id",
                                                 "page_name_field": "current_item.spec.name"
                                             },
                                             "next_step": "update_page_status",
                                             "description": "Save rendered sections to page_components",
                                             "output_field": "save_result"
                                         },
                                         "update_page_status": {
                                             "action": "update_page_status",
                                             "config": {
                                                 "status": "deployed",
                                                 "commit_from": "page_deployed.commit_sha",
                                                 "page_id_field": "current_item.spec.id"
                                             },
                                             "next_step": "complete_work_item",
                                             "description": "Mark page as deployed in database"
                                         },
                                         "complete_work_item": {
                                             "action": "complete_work_item",
                                             "config": {
                                                 "work_item_id": "current_item.id",
                                                 "commit_sha": "page_deployed.commit_sha"
                                             },
                                             "next_step": "complete_page",
                                             "description": "Mark work item as complete in queue",
                                             "output_field": "item_completed"
                                         },
                                         "fail_item": {
                                             "action": "fail_work_item",
                                             "config": {
                                                 "work_item_id": "current_item.id",
                                                 "error_message": "Content review not approved"
                                             },
                                             "next_step": "complete_page",
                                             "description": "Mark work item as failed (review rejected)"
                                         },
                                         "complete_page": {
                                             "action": "loop_complete",
                                             "description": "Work item processing complete"
                                         }
                                     }
                                 }
                             },
                             "next_step": "apply_site_design",
                             "description": "Process each work item: write, review, deploy, mark complete",
                             "output_field": "items_built"
                         },

                         "apply_site_design": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "webdesign-agent",
                                 "target_role": "webdesigner",
                                 "input_mapping": {
                                     "domain": "site_record.domain",
                                     "site_id": "site_record.site_id"
                                 },
                                 "timeout_seconds": 300
                             },
                             "next_step": "update_site_status",
                             "description": "Generate and deploy site stylesheet",
                             "output_field": "design_result"
                         },

                         "update_site_status": {
                             "action": "update_site_status",
                             "config": {
                                 "status": "deployed",
                                 "deployed_at": "now",
                                 "site_id_field": "site_record.site_id"
                             },
                             "next_step": "complete",
                             "description": "Mark site as deployed",
                             "output_field": "site_updated"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["site_record", "items_built", "build_items_written", "design_result"]
                             },
                             "description": "Site build complete"
                         }
                     }
                 }
             }'::jsonb,
             true,
             '["orchestration", "website-builder", "work-items", "unified"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.791',
             '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             '[]'::jsonb,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'experimental',
             '["website", "build", "maintenance", "work-items"]'::jsonb,
             '{}'::jsonb,
             '{"expects": {"input_data.domain": "string", "input_data.objective": "string", "reviewed_brief": "object"}, "required": ["input_data", "reviewed_brief"]}'::jsonb,
             '{"produces": {"site_id": "uuid", "items_built": "object", "build_items_written": "object"}}'::jsonb
         )
    ON CONFLICT (type,version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              capabilities = EXCLUDED.capabilities,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();