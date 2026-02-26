-- 057_image_build_handler.sql
--
-- image-build-handler agent definition
-- Self-contained handler for needs_logo and needs_hero_image work items.
--
-- Flow:
--   1. ensure_site_record — load site record from DB
--   2. spawn image-generator
--   3. check_purpose — branch on spec.purpose (logo vs hero)
--   4. call image-generator with correct prompt from image_prompts
--   5. store_asset — persist to assets table + content_data
--   6. deploy_image_asset — download from S3, optimize, git commit
--   7. complete
--
-- Called by: build-dispatch-loop (handler for needs_logo, needs_hero_image items)
-- Input from dispatch: site_id, domain, work_item_id, item_type, spec
--
-- Work item spec shape:
--   needs_logo:       {"purpose": "logo", "image_prompts": {"logo": "...", "hero_home": "..."}}
--   needs_hero_image: {"purpose": "hero", "image_prompts": {"logo": "...", "hero_home": "..."}}

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics,
    health_config, env_vars, version,
    delegation_preferences, agent_category, status,
    domain_tags, briefing_questionnaire,
    usage_count, is_snapshot, input_contract, output_contract
) VALUES (
             'image-build-handler',
             'Image Build Handler',
             'Self-contained handler for image work items (logo, hero). Calls image-generator, stores asset in DB, deploys optimized image to git. Used by dispatch loop for needs_logo and needs_hero_image items.',
             'orchestrator',
             '{
                 "workflow": {
                     "start_step": "ensure_site_record",
                     "steps": {

                         "ensure_site_record": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "spawn_image_generator",
                             "description": "Load site record from database",
                             "output_field": "site_record"
                         },

                         "spawn_image_generator": {
                             "action": "spawn_agent",
                             "config": {
                                 "role": "image_generator",
                                 "agent_type": "image-generator"
                             },
                             "next_step": "check_purpose",
                             "description": "Spawn image-generator agent",
                             "output_field": "image_generator_info"
                         },

                         "check_purpose": {
                             "action": "conditional",
                             "config": {
                                 "condition": "input_data.spec.purpose == logo",
                                 "then_step": "call_logo_gen",
                                 "else_step": "call_hero_gen"
                             },
                             "description": "Branch on image purpose to select correct prompt key"
                         },

                         "call_logo_gen": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "image-generator",
                                 "target_role": "image_generator",
                                 "input_mapping": {
                                     "prompt":    "input_data.spec.image_prompts.logo",
                                     "site_plan": "input_data.spec"
                                 },
                                 "output_mapping": {
                                     "prompt":       "generate.response.prompt",
                                     "image_uri":    "generate.response.image_uri",
                                     "image_url":    "generate.response.image_url",
                                     "generated_at": "generate.response.generated_at"
                                 },
                                 "timeout_seconds": 120
                             },
                             "next_step": "store_logo_asset",
                             "error_step": "complete_error",
                             "description": "Generate logo image",
                             "output_field": "image_result"
                         },

                         "call_hero_gen": {
                             "action": "call_agent",
                             "config": {
                                 "agent_type": "image-generator",
                                 "target_role": "image_generator",
                                 "input_mapping": {
                                     "prompt":    "input_data.spec.image_prompts.hero_home",
                                     "site_plan": "input_data.spec"
                                 },
                                 "output_mapping": {
                                     "prompt":       "generate.response.prompt",
                                     "image_uri":    "generate.response.image_uri",
                                     "image_url":    "generate.response.image_url",
                                     "generated_at": "generate.response.generated_at"
                                 },
                                 "timeout_seconds": 120
                             },
                             "next_step": "store_hero_asset",
                             "error_step": "complete_error",
                             "description": "Generate hero image",
                             "output_field": "image_result"
                         },

                         "store_logo_asset": {
                             "action": "store_asset",
                             "config": {
                                 "purpose":    "logo",
                                 "asset_type": "logo",
                                 "data_field": "image_result.image_url",
                                 "origin_type": "generated",
                                 "site_id_field": "site_record.site_id",
                                 "origin_prompt_field": "image_result.prompt",
                                 "update_site_brand_assets": true
                             },
                             "next_step": "deploy_image",
                             "error_step": "complete_error",
                             "description": "Store logo in assets table + content_data",
                             "output_field": "asset_stored"
                         },

                         "store_hero_asset": {
                             "action": "store_asset",
                             "config": {
                                 "purpose":    "hero",
                                 "asset_type": "image",
                                 "data_field": "image_result.image_url",
                                 "origin_type": "generated",
                                 "site_id_field": "site_record.site_id",
                                 "origin_prompt_field": "image_result.prompt",
                                 "update_site_brand_assets": true
                             },
                             "next_step": "deploy_image",
                             "error_step": "complete_error",
                             "description": "Store hero image in assets table + content_data",
                             "output_field": "asset_stored"
                         },

                         "deploy_image": {
                             "action": "deploy_image_asset",
                             "config": {
                                 "input_fields": ["s3_uri", "purpose", "domain"],
                                 "domain_field": "site_record.domain"
                             },
                             "next_step": "complete",
                             "error_step": "complete_error",
                             "description": "Download from S3, optimize, git commit. Purpose resolved dynamically from input_data.spec.purpose via ExtractActionInputs. s3_uri found via findStorageURI ({purpose}_uri set by store_asset).",
                             "output_field": "deploy_result"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["image_result", "asset_stored", "deploy_result"]
                             },
                             "description": "Image build complete"
                         },

                         "complete_error": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["image_result", "asset_stored"],
                                 "success_message": "Image build completed with errors"
                             },
                             "description": "Error path — dispatch loop marks work item failed"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 300
             }'::jsonb,
             true,
             '["image-generation", "asset-deploy", "git-commit"]'::jsonb,
             'docker.io/aqls/agent-chassis', 'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb, 1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'coordinator', 'experimental',
             '["build", "image", "asset"]'::jsonb, '{}'::jsonb,
             0, false,
             '{"required": ["site_id", "spec"], "optional": ["domain", "work_item_id", "item_type"], "description": "spec must contain purpose (logo|hero) and image_prompts object"}'::jsonb,
             '{"produces": {"image_result": "image_url + image_uri from generator", "asset_stored": "asset record in DB", "deploy_result": "git commit with optimized image"}}'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       image_tag = EXCLUDED.image_tag,
                                       resources = EXCLUDED.resources,
                                       agent_category = EXCLUDED.agent_category,
                                       domain_tags = EXCLUDED.domain_tags,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = now();

-- Also update WriteBuildItemsAction handler_agent references
-- (Go patch needed — change "image-generator" to "image-build-handler" for both needs_logo and needs_hero_image)
--
-- In write_build_items.go, the two insertWorkItem calls for logo/hero currently use:
--   handlerAgent: "image-generator"
-- Change both to:
--   handlerAgent: "image-build-handler"
--
-- Retroactive fix for any existing work items pointing at image-generator:
UPDATE site_work_items
SET handler_agent = 'image-build-handler'
WHERE item_type IN ('needs_logo', 'needs_hero_image')
  AND handler_agent = 'image-generator'
  AND status IN ('triaged', 'pending');

