currently:

         read_specs → ensure_site → load_components → load_styles → plan_site (LLM)
→ validate_plan → write_plan_spec → write_design_intent → write_content_direction
→ store_in_content_data → sync_pages → populate_nav → write_build_items → complete
--- changed to:
read_specs → ensure_site → load_components → load_styles → plan_site
→ validate_plan → write_site_plan → sync_pages → populate_nav → reconcile_site_plan → complete

--------------------------------------------------------------------------------------
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

--

-- path fix

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
                                 "site_id": "input_data.site_id"
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
                                 "prompt_template": "Plan a website for {{.input_data.domain}}.\n\n## Research Data\nIdentity: {{.site_specs.specs.identity}}\nClassification: {{.site_specs.specs.classification}}\n\n## Briefing Answers\n{{.site_specs.specs.briefing}}\n\n## Available Section Components\nThe following components are available in our component library. You MUST use ONLY these exact component names in the \"sections\" arrays:\n\n{{range .available_components}}\n- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}\n{{end}}\n\n## Available Style Collections\n{{.available_styles}}\n\n## Task\nCreate a comprehensive site plan using ONLY the components listed above.\n\nReturn JSON in this format:\n```json\n{\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Site Name\",\n      \"nav_label\": \"Home\",\n      \"nav_order\": 1,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"sections\": [\"hero\", \"features\", \"testimonials\", \"call_to_action\"]\n    }\n  ],\n  \"style_collection\": \"style-name\",\n  \"needs_logo\": true,\n  \"needs_images\": true,\n  \"image_prompts\": {\n    \"logo\": \"Description for logo generation\",\n    \"hero_home\": \"Description for home hero image\"\n  }\n}\n```\n\nSTRICT RULES:\n1. ONLY use component names from the \"Available Section Components\" list above\n2. DO NOT invent new component names\n3. Use these standard mappings:\n   - Hero/banner at page top: \"hero\" or page-specific variants like \"contact-hero\", \"services-hero\"\n   - Feature lists: \"features\"\n   - Service listings: \"services-grid\"\n   - Testimonials: \"testimonials\" or \"social_proof\"\n   - Calls to action: \"call_to_action\"\n   - Contact forms: \"contact-form\"\n   - Contact details: \"contact-info\"\n   - Team sections: \"leadership-team\"\n   - About content: \"about-content\"\n   - Differentiators: \"differentiators-section\"\n4. Choose style_collection based on industry and tone from the briefing\n5. Keep header navigation to 5-8 items maximum\n6. Always include: index (home) and contact pages\n7. Set needs_logo: true and needs_images: true (always)\n8. Provide image_prompts with \"logo\" and \"hero_home\" keys\n\nReturn ONLY valid JSON."
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


-- 064c_patch_build_site_planner_prompt.sql
--
-- Patch the EXISTING build-site-planner's LLM prompt to:
--   1. Read strategy aspect from site_specs (already passed via input_fields)
--   2. Use canonical page_type values
--   3. Respect strategist's recommendations (but have final say)
--
-- The build-site-planner already has the right workflow:
--   read_specs → ensure_site → load_components → load_styles → plan_site
--   → validate_plan → write_plan_spec → store_in_content_data
--   → sync_pages → populate_nav → write_build_items → complete
--
-- Only the prompt_template in the plan_site step needs updating.
-- The existing input_fields already include site_specs.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(E'Plan a website for {{.input_data.domain}}.\n\n## Research Data\nIdentity: {{.site_specs.specs.identity}}\nClassification: {{.site_specs.specs.classification}}\n\n## Domain Strategy\n{{if .site_specs.specs.strategy}}{{.site_specs.specs.strategy}}{{else}}No strategy data available — use the briefing and classification to determine site structure.{{end}}\n\n## Briefing Answers\n{{.site_specs.specs.briefing}}\n\n## Available Section Components\nYou MUST use ONLY these exact component names in the \"sections\" arrays:\n\n{{range .available_components}}\n- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}\n{{end}}\n\n## Available Style Collections\n{{.available_styles}}\n\n## Canonical Page Types\n\nEvery page MUST have a page_type from this list:\n\n| page_type | Description | Use for |\n|-----------|-------------|----------|\n| `index` | Home page | Always exactly one |\n| `content` | Standard content page | About, services, contact, FAQ, etc |\n| `landing` | Conversion-focused page | Lead capture, specific offers |\n| `entity-directory` | Searchable directory of entities | Business listings, provider directories |\n| `entity-page` | Individual entity profile | Single business/provider detail page |\n| `tool` | Interactive tool or calculator | Cost calculators, comparison tools |\n| `blog-index` | Blog/news listing page | Article index, news feed |\n| `blog-post` | Individual blog article | Editorial content, guides |\n\nNot all page types have builders available yet. Plan the IDEAL site regardless — the build system handles which pages can be built now vs later.\n\n## Strategy Guidance\n\nIf a domain strategy is available above, use it as strong input:\n- The recommended site_type should guide the overall structure\n- The recommended page_types should inform which pages you plan\n- The revenue model should shape what conversion/lead-capture mechanisms you include\n- The tone should influence your style_collection choice\n\nYou have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment but note why in strategy_notes.\n\nReturn JSON:\n```json\n{\n  \"site_type\": \"from the strategy or your own assessment\",\n  \"strategy_notes\": \"any notes on how you used or diverged from the strategy\",\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Site Name\",\n      \"page_type\": \"index\",\n      \"nav_label\": \"Home\",\n      \"nav_order\": 1,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"sections\": [\"hero\", \"features\", \"testimonials\", \"call-to-action\"]\n    },\n    {\n      \"name\": \"directory\",\n      \"title\": \"Find Gas Wholesalers | Gas Wholesalers\",\n      \"page_type\": \"entity-directory\",\n      \"nav_label\": \"Directory\",\n      \"nav_order\": 2,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"sections\": []\n    }\n  ],\n  \"style_collection\": \"style-name\",\n  \"needs_logo\": true,\n  \"needs_images\": true,\n  \"image_prompts\": {\n    \"logo\": \"Description for logo generation\",\n    \"hero_home\": \"Description for home hero image\"\n  }\n}\n```\n\nRULES:\n1. ONLY use component names from the Available Section Components list for sections arrays\n2. Every page MUST have a page_type from the canonical list\n3. Pages with page_type entity-directory, entity-page, tool, blog-index, blog-post may have empty sections arrays (their builders handle structure internally)\n4. Content and index pages need sections arrays populated with available components\n5. Always include: index (home) and contact pages\n6. Keep header navigation to 5-8 items maximum\n7. For authority-portal and local-directory site types: include directory pages even though their builder may not exist yet\n8. Set needs_logo: true and needs_images: true (always)\n9. Provide detailed image_prompts for logo and hero_home\n\nReturn ONLY valid JSON.'::text)
                     )
WHERE type = 'build-site-planner';

---


-- ============================================================================
-- Step 4: Add write_site_spec steps to pageflow-builder and site-work-orchestrator
--
-- Both workflows currently write planning data to content_data only.
-- Add steps to also write to site_specs for versioned storage.
--
-- In both workflows, after store_site_plan, add sync_plan_to_specs.
-- After store_reviewed_brief, add sync_brief_to_specs.
-- ============================================================================

-- ============================================================================
-- pageflow-builder: add spec sync steps
-- ============================================================================

-- Change store_site_plan.next_step from sync_pages_to_db to sync_plan_to_specs
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_site_plan,next_step}',
        '"sync_plan_to_specs"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder' AND deleted_at IS NULL;

-- Add sync_plan_to_specs step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,sync_plan_to_specs}',
        '{
            "action": "write_site_spec",
            "config": {
                "site_id": "site_record.site_id",
                "spec_data": "site_plan",
                "aspect": "site_plan",
                "source": "pageflow-builder",
                "source_agent": "site-planner"
            },
            "next_step": "sync_pages_to_db",
            "description": "Persist site plan to site_specs for versioned storage",
            "output_field": "plan_spec_written"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder' AND deleted_at IS NULL;

-- Change store_reviewed_brief.next_step from store_site_plan to sync_brief_to_specs
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_reviewed_brief,next_step}',
        '"sync_brief_to_specs"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder' AND deleted_at IS NULL;

-- Add sync_brief_to_specs step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,sync_brief_to_specs}',
        '{
            "action": "write_site_spec",
            "config": {
                "site_id": "site_record.site_id",
                "spec_data": "input_data.reviewed_brief",
                "aspect": "briefing",
                "source": "pageflow-builder",
                "source_agent": "briefing-agent"
            },
            "next_step": "store_site_plan",
            "description": "Persist briefing to site_specs for versioned storage",
            "output_field": "brief_spec_written"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'pageflow-builder' AND deleted_at IS NULL;


-- ============================================================================
-- site-work-orchestrator: same changes
-- ============================================================================

-- Change store_site_plan.next_step from sync_pages_to_db to sync_plan_to_specs
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_site_plan,next_step}',
        '"sync_plan_to_specs"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'site-work-orchestrator' AND deleted_at IS NULL;

-- Add sync_plan_to_specs step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,sync_plan_to_specs}',
        '{
            "action": "write_site_spec",
            "config": {
                "site_id": "site_record.site_id",
                "spec_data": "site_plan",
                "aspect": "site_plan",
                "source": "site-work-orchestrator",
                "source_agent": "site-planner"
            },
            "next_step": "sync_pages_to_db",
            "description": "Persist site plan to site_specs for versioned storage",
            "output_field": "plan_spec_written"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'site-work-orchestrator' AND deleted_at IS NULL;

-- Change store_reviewed_brief.next_step from store_site_plan to sync_brief_to_specs
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_reviewed_brief,next_step}',
        '"sync_brief_to_specs"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'site-work-orchestrator' AND deleted_at IS NULL;

-- Add sync_brief_to_specs step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,sync_brief_to_specs}',
        '{
            "action": "write_site_spec",
            "config": {
                "site_id": "site_record.site_id",
                "spec_data": "input_data.reviewed_brief",
                "aspect": "briefing",
                "source": "site-work-orchestrator",
                "source_agent": "briefing-agent"
            },
            "next_step": "store_site_plan",
            "description": "Persist briefing to site_specs for versioned storage",
            "output_field": "brief_spec_written"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'site-work-orchestrator' AND deleted_at IS NULL;


-- ============================================================================
-- Step 6: Add design_intent and content_direction to build-site-planner
--
-- After write_plan_spec, add steps to extract design intent and content
-- direction from the LLM plan and write them as separate spec aspects.
-- ============================================================================

-- Change write_plan_spec.next_step from store_in_content_data to write_design_intent
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,write_plan_spec,next_step}',
        '"write_design_intent"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL;

-- Add write_design_intent step
-- The planner's LLM prompt should produce design_intent in its output.
-- For now, we derive a basic one from the style_collection choice.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,write_design_intent}',
        '{
            "action": "write_site_spec",
            "config": {
                "site_id": "input_data.site_id",
                "spec_data": "site_plan.design_intent",
                "aspect": "design_intent",
                "source": "build-site-planner",
                "source_agent": "build-site-planner",
                "source_item_id": "input_data.work_item_id"
            },
            "next_step": "write_content_direction",
            "error_step": "store_in_content_data",
            "description": "Persist design intent to site_specs",
            "output_field": "design_intent_written"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL;

-- Add write_content_direction step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,write_content_direction}',
        '{
            "action": "write_site_spec",
            "config": {
                "site_id": "input_data.site_id",
                "spec_data": "site_plan.content_direction",
                "aspect": "content_direction",
                "source": "build-site-planner",
                "source_agent": "build-site-planner",
                "source_item_id": "input_data.work_item_id"
            },
            "next_step": "store_in_content_data",
            "error_step": "store_in_content_data",
            "description": "Persist content direction to site_specs",
            "output_field": "content_direction_written"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL;


-- ============================================================================
-- Verify the workflow chain changes
-- ============================================================================

-- pageflow-builder chain: store_reviewed_brief → sync_brief_to_specs → store_site_plan → sync_plan_to_specs → sync_pages_to_db
SELECT
    default_config->'workflow'->'steps'->'store_reviewed_brief'->>'next_step' as brief_next,
    default_config->'workflow'->'steps'->'sync_brief_to_specs'->>'next_step' as brief_sync_next,
    default_config->'workflow'->'steps'->'store_site_plan'->>'next_step' as plan_next,
    default_config->'workflow'->'steps'->'sync_plan_to_specs'->>'next_step' as plan_sync_next
FROM agent_definitions
WHERE type = 'pageflow-builder' AND deleted_at IS NULL;

-- site-work-orchestrator chain: same
SELECT
    default_config->'workflow'->'steps'->'store_reviewed_brief'->>'next_step' as brief_next,
    default_config->'workflow'->'steps'->'sync_brief_to_specs'->>'next_step' as brief_sync_next,
    default_config->'workflow'->'steps'->'store_site_plan'->>'next_step' as plan_next,
    default_config->'workflow'->'steps'->'sync_plan_to_specs'->>'next_step' as plan_sync_next
FROM agent_definitions
WHERE type = 'site-work-orchestrator' AND deleted_at IS NULL;

-- build-site-planner chain: write_plan_spec → write_design_intent → write_content_direction → store_in_content_data
SELECT
    default_config->'workflow'->'steps'->'write_plan_spec'->>'next_step' as plan_next,
    default_config->'workflow'->'steps'->'write_design_intent'->>'next_step' as design_next,
    default_config->'workflow'->'steps'->'write_content_direction'->>'next_step' as content_next
FROM agent_definitions
WHERE type = 'build-site-planner' AND deleted_at IS NULL;

---
-- disambiguation

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(REPLACE(
                default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
                'Build ONLY what is in the current phase described below. When pages list section_types, use those in the sections arrays. Section types that do not match existing component names will be resolved by the component selector — output them as-is. Do NOT invent pages beyond what the current phase specifies.',
                'IMPORTANT — ROADMAP OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the current phase below. For each page, use EXACTLY the section_types listed — even if they do not appear in the Available Section Components list above. Unknown section types are handled by the component selector downstream. Do NOT replace roadmap section_types with standard components. Do NOT invent additional pages. The roadmap is the authority for this site.'
                 )::text)
                     ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active = true;

---------------
-- backup = agent_def_build_site_planner_backup_20260505

-- ============================================================================
-- Migration 032: Phase 1 build-site-planner workflow change (doc 030 step 4)
-- ============================================================================
-- Replaces the build-site-planner default_config with the Phase 1 form.
--
-- Old terminal stretch:
--   validate_plan → write_plan_spec → write_design_intent →
--   write_content_direction → store_in_content_data → sync_pages →
--   populate_nav → write_build_items → complete
--
-- New terminal stretch:
--   validate_plan → write_site_plan → sync_pages → populate_nav →
--   reconcile_site_plan → complete
--
-- What changes:
--   - Replaces 4 site_specs writes with one write_site_plan call.
--     Plan and plan-time directives now live in the
--     site_plans / site_plan_pages / site_plan_sections /
--     site_plan_directives schema (migration 031).
--   - Replaces write_build_items (planner emits page work items
--     directly) with reconcile_site_plan (separate action that diffs
--     intended vs realised and emits work items for the delta).
--
-- What stays:
--   - read_specs, ensure_site, load_components, load_styles unchanged.
--   - plan_site (the LLM call) — prompt and output schema unchanged.
--     ValidateSitePlanAction strips site-chrome components and adds
--     ensure_pages defaults; output keys design_intent and
--     content_direction continue under the validated `site_plan` map.
--   - validate_plan unchanged except next_step now points to
--     write_site_plan.
--   - sync_pages_to_db (transitional): keeps the pages table populated
--     so page-build-handler can find rows. The reconciler reads pages
--     to do its diff.
--   - populate_nav_tables (transitional): nav tables still feed
--     page-build-handler.
--
-- The older site-planner agent (UUID f7c8bee1-...) used by
-- pageflow-builder and site-work-orchestrator is NOT touched by this
-- migration.
--
-- Backup mechanism: per-agent dated snapshot table, matching the
-- established convention (agent_def_page_build_handler_backup_20260331,
-- agent_def_webdesign_backup_20260416, etc).
--
-- To roll back:
--   UPDATE agent_definitions
--   SET default_config = (SELECT default_config FROM agent_def_build_site_planner_backup_20260505 WHERE id = 'f263eaa1-...'),
--       updated_at = NOW()
--   WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';
-- ============================================================================

BEGIN;

-- 1. Snapshot the row to a per-agent dated backup table (codebase convention).
--    CREATE TABLE … AS SELECT preserves the row's exact state; the table is
--    kept indefinitely as audit/rollback material.
CREATE TABLE IF NOT EXISTS agent_def_build_site_planner_backup_20260505 AS
SELECT * FROM agent_definitions
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
  AND type = 'build-site-planner';

-- Fail loudly if the snapshot is empty (no matching row) — better than
-- silently UPDATE'ing zero rows.
DO $$
DECLARE
snap_count INT;
BEGIN
SELECT COUNT(*) INTO snap_count FROM agent_def_build_site_planner_backup_20260505;
IF snap_count = 0 THEN
        RAISE EXCEPTION 'Snapshot table empty — no row found for build-site-planner UUID f263eaa1-61e1-446e-9410-648e12b7875b. Aborting.';
END IF;
END$$;

-- 2. Apply the new default_config (the column is named default_config,
--    NOT workflow_definition — verified against pg_catalog).
UPDATE agent_definitions
SET default_config = '{"workflow":{"steps":{"read_specs":{"action":"read_site_spec","config":{"site_id":"input_data.site_id"},"next_step":"ensure_site","description":"Load identity, classification, and briefing from site_specs","output_field":"site_specs"},"ensure_site":{"action":"ensure_site_record","config":{},"next_step":"load_components","description":"Ensure site record exists (needed for content_data storage and page sync)","output_field":"site_record"},"load_components":{"action":"query_database","config":{"query":"SELECT name, display_name, \"function\", category, description FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true ORDER BY category, name","output_format":"array"},"next_step":"load_styles","description":"Load available section components from database","output_field":"available_components"},"load_styles":{"action":"query_database","config":{"query":"SELECT name, display_name, category, description FROM style_collections WHERE is_active = true ORDER BY name","output_format":"array"},"next_step":"plan_site","description":"Load available style collections","output_field":"available_styles"},"plan_site":{"action":"execute_llm_prompt","config":{"ai_service":{"model":"claude-opus-4-6","provider":"anthropic","max_tokens":4000,"api_key_env_var":"ANTHROPIC_API_KEY"},"input_fields":["input_data","site_specs","available_components","available_styles"],"output_format":"json","prompt_template":"Plan a website for {{.input_data.domain}}.\n\n## Research Data\nIdentity: {{.site_specs.specs.identity}}\nClassification: {{.site_specs.specs.classification}}\n\n## Domain Strategy\n{{if .site_specs.specs.strategy}}{{.site_specs.specs.strategy}}{{else}}No strategy data available \u2014 use the briefing and classification to determine site structure.{{end}}\n\n## Briefing Answers\n{{.site_specs.specs.briefing}}\n\n{{if .site_specs.specs.mission_brief}}## Mission\n{{.site_specs.specs.mission_brief.text}}\n{{end}}\n{{if .site_specs.specs.roadmap_brief}}## Roadmap\nIMPORTANT \u2014 ROADMAP OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the current phase below. For each page, use EXACTLY the section_types listed \u2014 even if they do not appear in the Available Section Components list above. Unknown section types are handled by the component selector downstream. Do NOT replace roadmap section_types with standard components. Do NOT invent additional pages. The roadmap is the authority for this site.\n\n{{.site_specs.specs.roadmap_brief.text}}\n{{end}}\n## Available Section Components\nYou MUST use ONLY these exact component names in the \"sections\" arrays (unless the roadmap specifies section_types \u2014 those override):\n\n{{range .available_components}}\n- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}\n{{end}}\n\n## Available Style Collections\n{{.available_styles}}\n\n## Canonical Page Types\n\nEvery page MUST have a page_type from this list:\n\n| page_type | Description | Use for |\n|-----------|-------------|----------|\n| index | Home page | Always exactly one |\n| content | Standard content page | About, services, contact, FAQ, etc |\n| landing | Conversion-focused page | Lead capture, specific offers |\n| entity-directory | Searchable directory of entities | Business listings, provider directories |\n| entity-page | Individual entity profile | Single business/provider detail page |\n| tool | Interactive tool or calculator | Cost calculators, comparison tools |\n| blog-index | Blog/news listing page | Article index, news feed |\n| blog-post | Individual blog article | Editorial content, guides |\n\nNot all page types have builders available yet. Plan the IDEAL site regardless \u2014 the build system handles which pages can be built now vs later.\n\n## Strategy Guidance\n\nIf a domain strategy is available above, use it as strong input:\n- The recommended site_type should guide the overall structure\n- The recommended page_types should inform which pages you plan\n- The revenue model should shape what conversion/lead-capture mechanisms you include\n- The tone should influence your style_collection choice\n\nYou have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment but note why in strategy_notes.\n\nReturn JSON:\n{\n  \"site_type\": \"from the strategy, roadmap, or your own assessment\",\n  \"strategy_notes\": \"any notes on how you used or diverged from the strategy/roadmap\",\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Site Name\",\n      \"page_type\": \"index\",\n      \"nav_label\": \"Home\",\n      \"nav_order\": 1,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"sections\": [\"hero\", \"features\", \"testimonials\", \"call-to-action\"]\n    }\n  ],\n  \"style_collection\": \"style-name\",\n  \"needs_logo\": true,\n  \"needs_images\": true,\n  \"image_prompts\": {\n    \"logo\": \"Description for logo generation\",\n    \"hero_home\": \"Description for home hero image\"\n  },\n  \"design_intent\": {\n    \"style_direction\": \"professional-dark or modern-light or bold-creative\",\n    \"colour_mood\": \"Description of colour feeling and why it fits the industry\",\n    \"typography_mood\": \"Description of font personality\",\n    \"imagery_direction\": \"What images should show\",\n    \"layout_preference\": \"Layout style description\",\n    \"avoid\": [\"Things to avoid in design\"]\n  },\n  \"content_direction\": {\n    \"voice\": \"How the site should sound\",\n    \"emphasis\": \"What to emphasise in content\",\n    \"avoid_phrases\": [\"Phrases to never use\"],\n    \"social_proof_style\": \"How to handle testimonials and proof\",\n    \"blog_strategy\": \"Content strategy for blog if applicable\"\n  }\n}\n\nRULES:\n1. Use component names from the Available Section Components list for sections arrays \u2014 UNLESS the roadmap specifies section_types, in which case use those\n2. Every page MUST have a page_type from the canonical list\n3. Pages with page_type entity-directory, entity-page, tool, blog-index, blog-post may have empty sections arrays\n4. Content and index pages need sections arrays populated\n5. Always include: index (home) and contact pages (unless the roadmap says otherwise)\n6. Keep header navigation to 5-8 items maximum\n7. Set needs_logo: true and needs_images: true (always)\n8. Provide detailed image_prompts for logo and hero_home\n9. Include design_intent with explicit colour mood, typography direction, and layout preferences\n10. Include content_direction with voice, emphasis, and avoid_phrases tailored to the target audience\n11. If the classification data includes content_features.news_feed.recommended = true, add \"latest-news\" to the homepage sections\n12. When a roadmap is present, the pages and section_types from the current phase take precedence over your own page planning\n\nReturn ONLY valid JSON."},"next_step":"validate_plan","description":"LLM creates site plan from research and briefing data","output_field":"llm_plan"},"validate_plan":{"action":"validate_site_plan","config":{"max_pages":20,"plan_field":"llm_plan.result","ensure_pages":["index","contact"],"default_style":"professional-dark","validate_components":true},"next_step":"write_site_plan","description":"Validate and normalize the site plan","output_field":"site_plan"},"write_site_plan":{"action":"write_site_plan","config":{"target_site_id":"site_record.site_id"},"next_step":"sync_pages","description":"Write validated plan to site_plans + site_plan_pages + site_plan_sections + site_plan_directives; transfer HITL locks from previous current plan","output_field":"plan_written"},"sync_pages":{"action":"sync_pages_to_db","config":{"input_fields":["site_record","site_plan"]},"next_step":"populate_nav","description":"Create page records from site plan","output_field":"db_sync"},"populate_nav":{"action":"populate_nav_tables","config":{"input_fields":["site_id"],"max_header_items":8},"next_step":"reconcile_site_plan","description":"Populate navigation tables from page records","output_field":"nav_data"},"reconcile_site_plan":{"action":"reconcile_site_plan","config":{"target_site_id":"site_record.site_id","plan_id":"plan_written.plan_id"},"next_step":"complete","description":"Diff site_plan_pages vs realised pages; emit needs_page work items for the delta; emit terminal needs_rerender if any","output_field":"reconcile_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["site_plan","plan_written","db_sync","reconcile_result"]},"description":"Site planning complete \u2014 plan written, pages synced, reconciler emitted work items"}},"start_step":"read_specs"},"processing_mode":"orchestrator","timeout_seconds":300}'::jsonb,
    updated_at     = NOW()
WHERE id   = 'f263eaa1-61e1-446e-9410-648e12b7875b'
  AND type = 'build-site-planner';

-- 3. Verification — should show exactly one row with the new step list.
SELECT id, type,
       default_config->'workflow'->>'start_step' AS start_step,
    (SELECT count(*) FROM jsonb_object_keys(default_config->'workflow'->'steps')) AS step_count
FROM agent_definitions
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

COMMIT;

-- After commit, this query lists the new step names for visual confirmation:
-- SELECT jsonb_object_keys(default_config->'workflow'->'steps') AS step
-- FROM agent_definitions
-- WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
-- ORDER BY step;


-- new workflow for reference:
{
  "workflow": {
    "steps": {
      "read_specs": {
        "action": "read_site_spec",
        "config": {
          "site_id": "input_data.site_id"
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
          "query": "SELECT name, display_name, \"function\", category, description FROM content_components WHERE component_level IN ('section', 'element') AND is_active = true ORDER BY category, name",
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
            "model": "claude-opus-4-6",
            "provider": "anthropic",
            "max_tokens": 4000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "input_fields": [
            "input_data",
            "site_specs",
            "available_components",
            "available_styles"
          ],
          "output_format": "json",
          "prompt_template": "Plan a website for {{.input_data.domain}}.\n\n## Research Data\nIdentity: {{.site_specs.specs.identity}}\nClassification: {{.site_specs.specs.classification}}\n\n## Domain Strategy\n{{if .site_specs.specs.strategy}}{{.site_specs.specs.strategy}}{{else}}No strategy data available \u2014 use the briefing and classification to determine site structure.{{end}}\n\n## Briefing Answers\n{{.site_specs.specs.briefing}}\n\n{{if .site_specs.specs.mission_brief}}## Mission\n{{.site_specs.specs.mission_brief.text}}\n{{end}}\n{{if .site_specs.specs.roadmap_brief}}## Roadmap\nIMPORTANT \u2014 ROADMAP OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the current phase below. For each page, use EXACTLY the section_types listed \u2014 even if they do not appear in the Available Section Components list above. Unknown section types are handled by the component selector downstream. Do NOT replace roadmap section_types with standard components. Do NOT invent additional pages. The roadmap is the authority for this site.\n\n{{.site_specs.specs.roadmap_brief.text}}\n{{end}}\n## Available Section Components\nYou MUST use ONLY these exact component names in the \"sections\" arrays (unless the roadmap specifies section_types \u2014 those override):\n\n{{range .available_components}}\n- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}\n{{end}}\n\n## Available Style Collections\n{{.available_styles}}\n\n## Canonical Page Types\n\nEvery page MUST have a page_type from this list:\n\n| page_type | Description | Use for |\n|-----------|-------------|----------|\n| index | Home page | Always exactly one |\n| content | Standard content page | About, services, contact, FAQ, etc |\n| landing | Conversion-focused page | Lead capture, specific offers |\n| entity-directory | Searchable directory of entities | Business listings, provider directories |\n| entity-page | Individual entity profile | Single business/provider detail page |\n| tool | Interactive tool or calculator | Cost calculators, comparison tools |\n| blog-index | Blog/news listing page | Article index, news feed |\n| blog-post | Individual blog article | Editorial content, guides |\n\nNot all page types have builders available yet. Plan the IDEAL site regardless \u2014 the build system handles which pages can be built now vs later.\n\n## Strategy Guidance\n\nIf a domain strategy is available above, use it as strong input:\n- The recommended site_type should guide the overall structure\n- The recommended page_types should inform which pages you plan\n- The revenue model should shape what conversion/lead-capture mechanisms you include\n- The tone should influence your style_collection choice\n\nYou have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment but note why in strategy_notes.\n\nReturn JSON:\n{\n  \"site_type\": \"from the strategy, roadmap, or your own assessment\",\n  \"strategy_notes\": \"any notes on how you used or diverged from the strategy/roadmap\",\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Site Name\",\n      \"page_type\": \"index\",\n      \"nav_label\": \"Home\",\n      \"nav_order\": 1,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"sections\": [\"hero\", \"features\", \"testimonials\", \"call-to-action\"]\n    }\n  ],\n  \"style_collection\": \"style-name\",\n  \"needs_logo\": true,\n  \"needs_images\": true,\n  \"image_prompts\": {\n    \"logo\": \"Description for logo generation\",\n    \"hero_home\": \"Description for home hero image\"\n  },\n  \"design_intent\": {\n    \"style_direction\": \"professional-dark or modern-light or bold-creative\",\n    \"colour_mood\": \"Description of colour feeling and why it fits the industry\",\n    \"typography_mood\": \"Description of font personality\",\n    \"imagery_direction\": \"What images should show\",\n    \"layout_preference\": \"Layout style description\",\n    \"avoid\": [\"Things to avoid in design\"]\n  },\n  \"content_direction\": {\n    \"voice\": \"How the site should sound\",\n    \"emphasis\": \"What to emphasise in content\",\n    \"avoid_phrases\": [\"Phrases to never use\"],\n    \"social_proof_style\": \"How to handle testimonials and proof\",\n    \"blog_strategy\": \"Content strategy for blog if applicable\"\n  }\n}\n\nRULES:\n1. Use component names from the Available Section Components list for sections arrays \u2014 UNLESS the roadmap specifies section_types, in which case use those\n2. Every page MUST have a page_type from the canonical list\n3. Pages with page_type entity-directory, entity-page, tool, blog-index, blog-post may have empty sections arrays\n4. Content and index pages need sections arrays populated\n5. Always include: index (home) and contact pages (unless the roadmap says otherwise)\n6. Keep header navigation to 5-8 items maximum\n7. Set needs_logo: true and needs_images: true (always)\n8. Provide detailed image_prompts for logo and hero_home\n9. Include design_intent with explicit colour mood, typography direction, and layout preferences\n10. Include content_direction with voice, emphasis, and avoid_phrases tailored to the target audience\n11. If the classification data includes content_features.news_feed.recommended = true, add \"latest-news\" to the homepage sections\n12. When a roadmap is present, the pages and section_types from the current phase take precedence over your own page planning\n\nReturn ONLY valid JSON."
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
          "ensure_pages": [
            "index",
            "contact"
          ],
          "default_style": "professional-dark",
          "validate_components": true
        },
        "next_step": "write_site_plan",
        "description": "Validate and normalize the site plan",
        "output_field": "site_plan"
      },
      "write_site_plan": {
        "action": "write_site_plan",
        "config": {
          "target_site_id": "site_record.site_id"
        },
        "next_step": "sync_pages",
        "description": "Write validated plan to site_plans + site_plan_pages + site_plan_sections + site_plan_directives; transfer HITL locks from previous current plan",
        "output_field": "plan_written"
      },
      "sync_pages": {
        "action": "sync_pages_to_db",
        "config": {
          "input_fields": [
            "site_record",
            "site_plan"
          ]
        },
        "next_step": "populate_nav",
        "description": "Create page records from site plan",
        "output_field": "db_sync"
      },
      "populate_nav": {
        "action": "populate_nav_tables",
        "config": {
          "input_fields": [
            "site_id"
          ],
          "max_header_items": 8
        },
        "next_step": "reconcile_site_plan",
        "description": "Populate navigation tables from page records",
        "output_field": "nav_data"
      },
      "reconcile_site_plan": {
        "action": "reconcile_site_plan",
        "config": {
          "target_site_id": "site_record.site_id",
          "plan_id": "plan_written.plan_id"
        },
        "next_step": "complete",
        "description": "Diff site_plan_pages vs realised pages; emit needs_page work items for the delta; emit terminal needs_rerender if any",
        "output_field": "reconcile_result"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {
          "output_fields": [
            "site_plan",
            "plan_written",
            "db_sync",
            "reconcile_result"
          ]
        },
        "description": "Site planning complete \u2014 plan written, pages synced, reconciler emitted work items"
      }
    },
    "start_step": "read_specs"
  },
  "processing_mode": "orchestrator",
  "timeout_seconds": 300
}

