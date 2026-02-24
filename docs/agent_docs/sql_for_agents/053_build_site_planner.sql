-- build-site-planner agent definition
-- Handler for: needs_site_plan work items
-- Pipeline position: after build-briefing-agent, before page content writing
--
-- Receives from dispatch loop:
--   input_data.site_id       — UUID of the site
--   input_data.domain        — domain name
--   input_data.work_item_id  — the work item being processed
--
-- Reads from site_specs:
--   identity       — company info from domain research
--   classification — site type, recommended builder
--   briefing       — answered questionnaire from build-briefing-agent
--
-- Outputs:
--   site_specs aspect "site_plan"  — the validated plan
--   sites.content_data             — plan merged into site record
--   pages table                    — page records synced from plan
--   nav_items table                — navigation populated
--   site_work_items                — one work item per page for content writing
--
-- Creates next work items via write_build_items:
--   needs_content_write → page-content-writer (one per page)

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    env_vars,
    version,
    delegation_preferences,
    agent_category,
    status,
    domain_tags,
    briefing_questionnaire,
    usage_count,
    is_snapshot,
    input_contract,
    output_contract
) VALUES (
             'build-site-planner',
             'Build Site Planner',
             'Handler-mode site planner for the build dispatch pipeline. Reads research and briefing from site_specs, plans site structure via LLM, validates plan, syncs pages to DB, creates page content work items.',
             'specialist',
             '{
                 "workflow": {
                     "start_step": "read_specs",
                     "steps": {

                         "read_specs": {
                             "action": "read_site_spec",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "aspect": "all"
                             },
                             "next_step": "ensure_site",
                             "description": "Load identity, classification, and briefing from site_specs",
                             "output_field": "site_specs"
                         },

                         "ensure_site": {
                             "action": "ensure_site_record",
                             "config": {},
                             "next_step": "load_components",
                             "description": "Ensure site record exists (needed for content_data storage and page sync)",
                             "output_field": "site_record"
                         },

                         "load_components": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT name, display_name, \"function\", category, description FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true ORDER BY category, name",
                                 "output_format": "array"
                             },
                             "next_step": "load_styles",
                             "description": "Load available section components from database",
                             "output_field": "available_components"
                         },

                         "load_styles": {
                             "action": "query_database",
                             "config": {
                                 "query": "SELECT name, display_name, category, description FROM style_collections WHERE is_active = true ORDER BY name",
                                 "output_format": "array"
                             },
                             "next_step": "plan_site",
                             "description": "Load available style collections",
                             "output_field": "available_styles"
                         },

                         "plan_site": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5",
                                     "provider": "anthropic",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["input_data", "site_specs", "available_components", "available_styles"],
                                 "output_format": "json",
                                 "prompt_template": "Plan a website for {{.input_data.domain}}.\n\n## Research Data\nIdentity: {{.site_specs.identity}}\nClassification: {{.site_specs.classification}}\n\n## Briefing Answers\n{{.site_specs.briefing}}\n\n## Available Section Components\nThe following components are available in our component library. You MUST use ONLY these exact component names in the \"sections\" arrays:\n\n{{range .available_components}}\n- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}\n{{end}}\n\n## Available Style Collections\n{{.available_styles}}\n\n## Task\nCreate a comprehensive site plan using ONLY the components listed above.\n\nReturn JSON in this format:\n```json\n{\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Site Name\",\n      \"nav_label\": \"Home\",\n      \"nav_order\": 1,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"sections\": [\"hero\", \"features\", \"testimonials\", \"call_to_action\"]\n    }\n  ],\n  \"style_collection\": \"style-name\",\n  \"needs_logo\": true,\n  \"needs_images\": true,\n  \"image_prompts\": {\n    \"logo\": \"Description for logo generation\",\n    \"hero_home\": \"Description for home hero image\"\n  }\n}\n```\n\nSTRICT RULES:\n1. ONLY use component names from the \"Available Section Components\" list above\n2. DO NOT invent new component names\n3. Use these standard mappings:\n   - Hero/banner at page top: \"hero\" or page-specific variants like \"contact-hero\", \"services-hero\"\n   - Feature lists: \"features\"\n   - Service listings: \"services-grid\"\n   - Testimonials: \"testimonials\" or \"social_proof\"\n   - Calls to action: \"call_to_action\"\n   - Contact forms: \"contact-form\"\n   - Contact details: \"contact-info\"\n   - Team sections: \"leadership-team\"\n   - About content: \"about-content\"\n   - Differentiators: \"differentiators-section\"\n4. Choose style_collection based on industry and tone from the briefing\n5. Keep header navigation to 5-8 items maximum\n6. Always include: index (home) and contact pages\n7. Set needs_logo: true and needs_images: true (always)\n8. Provide image_prompts with \"logo\" and \"hero_home\" keys\n\nReturn ONLY valid JSON."
                             },
                             "next_step": "validate_plan",
                             "description": "LLM creates site plan from research and briefing data",
                             "output_field": "llm_plan"
                         },

                         "validate_plan": {
                             "action": "validate_site_plan",
                             "config": {
                                 "max_pages": 20,
                                 "plan_field": "llm_plan.result",
                                 "ensure_pages": ["index", "contact"],
                                 "default_style": "professional-dark",
                                 "validate_components": true
                             },
                             "next_step": "write_plan_spec",
                             "description": "Validate and normalize the site plan",
                             "output_field": "site_plan"
                         },

                         "write_plan_spec": {
                             "action": "write_site_spec",
                             "config": {
                                 "site_id": "input_data.site_id",
                                 "spec_data": "site_plan",
                                 "aspect": "site_plan",
                                 "source": "build-site-planner",
                                 "source_agent": "build-site-planner",
                                 "source_item_id": "input_data.work_item_id"
                             },
                             "next_step": "store_in_content_data",
                             "description": "Persist site plan to site_specs",
                             "output_field": "plan_written"
                         },

                         "store_in_content_data": {
                             "action": "update_site_content",
                             "config": {
                                 "merge": true,
                                 "content_field": "site_plan",
                                 "site_id_field": "site_record.site_id"
                             },
                             "next_step": "sync_pages",
                             "description": "Store plan in sites.content_data for downstream actions",
                             "output_field": "content_stored"
                         },

                         "sync_pages": {
                             "action": "sync_pages_to_db",
                             "config": {
                                 "input_fields": ["site_record", "site_plan"]
                             },
                             "next_step": "populate_nav",
                             "description": "Create page records from site plan",
                             "output_field": "db_sync"
                         },

                         "populate_nav": {
                             "action": "populate_nav_tables",
                             "config": {
                                 "input_fields": ["site_id"],
                                 "max_header_items": 8
                             },
                             "next_step": "write_build_items",
                             "description": "Populate navigation tables from page records",
                             "output_field": "nav_data"
                         },

                         "write_build_items": {
                             "action": "write_build_items",
                             "config": {
                                 "site_id": "site_record.site_id",
                                 "site_plan": "site_plan"
                             },
                             "next_step": "complete",
                             "description": "Create work items for each page — dispatch loop picks these up next",
                             "output_field": "build_items"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["site_plan", "plan_written", "db_sync", "build_items"]
                             },
                             "description": "Site planning complete — page content items queued for dispatch"
                         }
                     }
                 },
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 300
             }'::jsonb,
             true,
             '["planning", "site-architecture", "component-selection", "llm"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.803',
             '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
             '[]'::jsonb,
             1,
             '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
             'specialist',
             'experimental',
             '["site-planning", "build-pipeline"]'::jsonb,
             '{}'::jsonb,
             0,
             false,
             '{"required": ["site_id"], "optional": ["domain", "work_item_id"], "description": "Receives site_id from dispatch loop. Reads all specs from site_specs."}'::jsonb,
             '{"produces": {"site_plan": "Validated site plan", "plan_written": "site_spec write result", "db_sync": "Pages synced to database", "build_items": "Page content work items created"}}'::jsonb
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