-- ============================================================================
-- ⚠ STALE PROMPT TEXT — THIS SEED IS HISTORY, THE LIVE ROW IS FACT (added 2026-09-04)
--
-- This file still contains the pre-2026-09-02 imagery instruction, THREE TIMES
-- (lines ~1347, ~1652, ~2058):
--
--     "Use sparingly in v1 - most plans will have zero section-scope entries."
--
-- Migration 718 (2026-09-02, 718_planner_imagery_content_expected_prompt_and_
-- exemplar.sql) REPLACED that sentence in the live agent_definitions row with:
--
--     "Content-carrying imagery is EXPECTED here, not exceptional..."
--
-- 718 edited the LIVE ROW, not this seed. The text below was never updated and
-- is left as written, because a seed is a record of what the agent WAS.
--
-- ⚠ WHY THIS BANNER EXISTS RATHER THAN A QUIET EDIT: the superseded sentence has
-- now been quoted as live evidence THREE TIMES in two days, by three different
-- lanes, and it reached the owner twice - once in a decision brief that routed a
-- fleet-wide ruling onto a cause that had already been fixed, and once via a lane's
-- owner-facing README that carried it 24h after the correction landed in that same
-- lane's NOTES. Grepping the repo for the planner's imagery rules is the OBVIOUS
-- move and it returns this file, named after the agent, in triplicate, with nothing
-- marking it superseded.
--
-- [MEASURED 2026-09-04] `agent_definitions` rows fleet-wide containing "sparingly",
-- across active/inactive/snapshot/undeleted: ZERO. The line is dead in the system
-- and alive only here.
--
-- BEFORE QUOTING ANY PROMPT TEXT FROM THIS FILE, read the live row:
--   SELECT default_config::text FROM agent_definitions
--    WHERE id='f263eaa1-61e1-446e-9410-648e12b7875b';
-- (build-site-planner; 39,431 B as of 2026-09-04. Confirm by CONTENT, not id:
--  `grep -c sparingly` must be 0 and "Content-carrying imagery is EXPECTED" present.)
--
-- Owned by: docs/agent_docs/docs024_key_docs_latest/infographics/ (the selection
-- rule) and framework_prompts_positive_voice (the prompt bytes). See MEMORY
-- [[seed-sql-is-history-live-row-is-fact]] and LANDMINES.md.
-- ============================================================================

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

-- ============================================================================
-- Migration 034: Tighten build-site-planner prompt — strict role vocabulary,
--                no speculative pages
-- ============================================================================
-- Two surgical changes to the plan_site step's prompt_template:
--
-- 1. Replace "Every page MUST have a page_type from this list" with
--    "Use ONLY these page_type values, verbatim, lowercase,
--    dash-separated". The original wording allowed the LLM to think
--    "from this list" was a guideline; the new wording forbids
--    invented values and tells the LLM to default to `content` if no
--    listed type fits cleanly.
--
-- 2. Replace "Plan the IDEAL site regardless..." with explicit
--    instruction to plan ONLY pages directly justified by inputs and
--    NOT to add speculative/demo/example pages. The original encouraged
--    the LLM to pad — gamesdesign's first Phase 1 plan included
--    `prototype-page` and `post` neither of which had any justification
--    in the briefing or strategy.
--
-- The change is applied via jsonb_set + replace() on the live prompt
-- text, so it works regardless of whatever the current prompt looks
-- like (idempotent against the specific old strings — running this
-- migration a second time is a no-op because the old strings are gone).
--
-- Rollback: agent_def_build_site_planner_backup_20260505 was captured
-- before migration 032 and contains the pre-Phase-1 prompt; for prompt-
-- only rollback, copy default_config -> 'workflow' -> 'steps' ->
-- 'plan_site' -> 'config' -> 'prompt_template' from there.
-- ============================================================================

BEGIN;

-- Capture the current default_config before the prompt change. Per the
-- per-agent dated backup convention.
CREATE TABLE IF NOT EXISTS agent_def_build_site_planner_backup_20260505_promptchange AS
SELECT * FROM agent_definitions
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
  AND type = 'build-site-planner';

DO $$
DECLARE
snap_count INT;
BEGIN
SELECT COUNT(*) INTO snap_count
FROM agent_def_build_site_planner_backup_20260505_promptchange;
IF snap_count = 0 THEN
        RAISE EXCEPTION 'Snapshot empty — no row found for build-site-planner. Aborting.';
END IF;
END$$;

-- Apply the two replacements as a single jsonb_set call. We extract the
-- current prompt as text, run replace() twice, then set it back. This
-- is idempotent: running again finds no matching old strings and
-- replace() returns the unchanged value.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
                replace(
                        replace(
                                default_config #>> '{workflow,steps,plan_site,config,prompt_template}',
                                'Every page MUST have a page_type from this list:',
                                'Use ONLY these page_type values, verbatim, lowercase, dash-separated:'
                        ),
                        'Not all page types have builders available yet. Plan the IDEAL site regardless — the build system handles which pages can be built now vs later.',
                        E'Plan ONLY pages that are directly justified by the briefing, strategy, classification, or roadmap. Do NOT add speculative, demo, or example pages. Every page in your output must serve an explicit need surfaced by one of those inputs. If you don''t have evidence for a page, leave it out — fewer well-justified pages are better than padding the count.\n\nIf a page doesn''t fit any category cleanly, use `content` as the default. Do not invent new page_type values.'
                )
        )
                     ),
    updated_at = NOW()
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

-- Verification: show that the new strings are present in the prompt.
SELECT
    'has_new_role_constraint' AS check,
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%Use ONLY these page_type values%' AS pass
FROM agent_definitions
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
UNION ALL
SELECT
    'has_no_ideal_site_phrase',
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        NOT LIKE '%Plan the IDEAL site%'
FROM agent_definitions
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

COMMIT;


-- remove ensure contact, about pages validation

-- ============================================================================
-- Migration 035: Remove `contact` from build-site-planner's ensure_pages
-- ============================================================================
-- The validate_plan step on build-site-planner has historically forced
-- `index` and `contact` to be present on every plan via:
--
--   "ensure_pages": ["index", "contact"]
--
-- This created a `contact` stub for sites that didn't justify one
-- (e.g. gamesdesign, where the briefing pointed at tools/guides as the
-- product, with no justification for a contact page). The stub had no
-- role on it, which then propagated empty role values to
-- site_plan_pages and pages.
--
-- The Phase 1 canonicaliser now defaults empty roles to "content", so
-- the stub no longer breaks the persist step — but the principle
-- remains: ensure_pages is a workflow-level forced add that should be
-- driven by domain need (e.g. the strategist or briefing recommending
-- a contact page) rather than baked into every site's planner config.
--
-- This migration narrows ensure_pages to ["index"] only. The home page
-- is genuinely required for any web product and is a reasonable
-- universal default. Other pages (contact, about, terms, privacy)
-- should arrive when justified by upstream specs.
--
-- Rollback: previous full default_config is in
--   agent_def_build_site_planner_backup_20260505 (from migration 032).
-- For ensure_pages-only rollback, jsonb_set the array back to
--   ["index", "contact"].
-- ============================================================================

BEGIN;

CREATE TABLE IF NOT EXISTS agent_def_build_site_planner_backup_20260505_ensurepages AS
SELECT * FROM agent_definitions
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,validate_plan,config,ensure_pages}',
        '["index"]'::jsonb
                     ),
    updated_at = NOW()
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

-- Verify the change took.
SELECT default_config #> '{workflow,steps,validate_plan,config,ensure_pages}' AS ensure_pages
FROM agent_definitions
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

COMMIT;

----

-- backup
-- phase_2g_step3_pre_migration_backup.sql
--
-- Snapshots build-site-planner's current default_config as a single UPDATE
-- statement, written to phase_2g_step3_rollback.sql. To roll back the
-- prompt-template change later, run:
--
--   psql -d clients_db -f phase_2g_step3_rollback.sql
--
-- Safe to run any time before the main migration. Idempotent — re-running
-- overwrites the rollback file with whatever the current state is.

\set ON_ERROR_STOP on

-- Sanity: confirm the target row exists and is active.
DO $check$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
          AND is_active = true
    ) THEN
        RAISE EXCEPTION 'build-site-planner row not found or inactive (id f263eaa1-...)';
END IF;
END
$check$;

-- Emit a self-contained rollback UPDATE.
\pset format unaligned
\pset tuples_only on
\o phase_2g_step3_rollback.sql

SELECT format(
               E'-- Rollback for phase_2g_step3 — restores build-site-planner default_config\n'
    '-- Generated %s\n'
    '\\set ON_ERROR_STOP on\n'
    'BEGIN;\n'
    'UPDATE agent_definitions\n'
    '   SET default_config = %L::jsonb,\n'
    '       updated_at = now()\n'
    ' WHERE id = %L\n'
    '   AND is_active = true;\n'
    'COMMIT;\n',
               now()::text,
               default_config::text,
               id::text
       )
FROM agent_definitions
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

\o
\pset tuples_only off
\pset format aligned

\echo ''
\echo 'Rollback SQL written to phase_2g_step3_rollback.sql'
\echo 'To roll back if needed: psql -d clients_db -f phase_2g_step3_rollback.sql'

----

-- imagery prompt

-- phase_2g_step3_planner_imagery_prompt.sql
--
-- Phase 2G step 3 — teach build-site-planner to emit a structured `imagery`
-- block alongside the legacy `image_prompts` map. Downstream
-- write_site_plan_action.flattenImageryBlock has been deployed dormant
-- since step 2; this migration is the first behavioural change that
-- populates site_plan_imagery rows.
--
-- Idempotent in effect: replaces the entire prompt_template; safe to re-run.
-- Wrapped in BEGIN/COMMIT with a post-write verification that aborts the
-- transaction if either the new "## Imagery Block" section or the
-- `"imagery":` key is missing from the resulting template.
--
-- Run AFTER phase_2g_step3_pre_migration_backup.sql.

\set ON_ERROR_STOP on

BEGIN;

-- Sanity: target row still exists.
DO $check$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
          AND is_active = true
    ) THEN
        RAISE EXCEPTION 'build-site-planner row not found or inactive';
END IF;
END
$check$;

-- Replace the prompt_template field. Dollar-quoted to preserve quotes,
-- newlines, and braces in the template body. Tag $bsp_prompt$ chosen
-- to avoid collision with any text inside the prompt.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb($bsp_prompt$Plan a website for {{.input_data.domain}}.

## Research Data
Identity: {{.site_specs.specs.identity}}
Classification: {{.site_specs.specs.classification}}

## Domain Strategy
{{if .site_specs.specs.strategy}}{{.site_specs.specs.strategy}}{{else}}No strategy data available — use the briefing and classification to determine site structure.{{end}}

## Briefing Answers
{{.site_specs.specs.briefing}}

{{if .site_specs.specs.mission_brief}}## Mission
{{.site_specs.specs.mission_brief.text}}
{{end}}
{{if .site_specs.specs.roadmap_brief}}## Roadmap
IMPORTANT — ROADMAP OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the current phase below. For each page, use EXACTLY the section_types listed — even if they do not appear in the Available Section Components list above. Unknown section types are handled by the component selector downstream. Do NOT replace roadmap section_types with standard components. Do NOT invent additional pages. The roadmap is the authority for this site.

{{.site_specs.specs.roadmap_brief.text}}
{{end}}
## Available Section Components
You MUST use ONLY these exact component names in the "sections" arrays (unless the roadmap specifies section_types — those override):

{{range .available_components}}
- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}
{{end}}

## Available Style Collections
{{.available_styles}}

## Canonical Page Types

Use ONLY these page_type values, verbatim, lowercase, dash-separated:

| page_type | Description | Use for |
|-----------|-------------|----------|
| index | Home page | Always exactly one |
| content | Standard content page | About, services, contact, FAQ, etc |
| landing | Conversion-focused page | Lead capture, specific offers |
| entity-directory | Searchable directory of entities | Business listings, provider directories |
| entity-page | Individual entity profile | Single business/provider detail page |
| tool | Interactive tool or calculator | Cost calculators, comparison tools |
| blog-index | Blog/news listing page | Article index, news feed |
| blog-post | Individual blog article | Editorial content, guides |

Plan ONLY pages that are directly justified by the briefing, strategy, classification, or roadmap. Do NOT add speculative, demo, or example pages. Every page in your output must serve an explicit need surfaced by one of those inputs. If you don't have evidence for a page, leave it out — fewer well-justified pages are better than padding the count.

If a page doesn't fit any category cleanly, use `content` as the default. Do not invent new page_type values.

## Imagery Block

Return a structured `imagery` block alongside the legacy `image_prompts` map. The downstream image pipeline reads `imagery` as the source of truth; `image_prompts` is retained for backward compatibility during transition.

### Shape

`imagery` is a nested object with three optional sub-keys: `site`, `pages`, `sections`. Scope is implied by position in the nesting — entries do NOT carry `scope` or `scope_ref` fields themselves.

```
"imagery": {
  "site":     [ entry, entry, ... ],
  "pages":    { "<page_name>": [ entry, ... ], ... },
  "sections": { "<page_name>:<section_ordering>": [ entry, ... ], ... }
}
```

The `<section_ordering>` is the zero-indexed position of the section within that page's `sections` array (so `"index:0"` is the first section of the home page).

### Entry fields

| Field | Required | Notes |
|---|---|---|
| key | yes | short identifier, lowercase, underscore-separated (e.g. `logo`, `hero_about`, `illustration_team_values`). Unique within its scope. |
| kind | yes | one of: `logo`, `hero`, `illustration`, `icon`, `infographic`. No other values permitted. |
| prompt | yes | full generation prompt, one sentence to one paragraph. Reflect the site's `imagery_direction` and `colour_mood`. |
| style_hints | optional | JSON object, e.g. {"medium": "line drawing", "mood": "collaborative"} |
| constraints | optional | JSON object, e.g. {"aspect": "1:1", "transparent_background": true} |

### What to populate

- **`site`** — imagery that appears across pages or is brand-level. Always include one `logo` entry. Optionally include one site-wide brand `hero` or `illustration` entry. Two to three entries is typical.
- **`pages`** — one entry per page-specific image. Always include a `hero` entry for each page whose `sections` array contains a hero-class component. The map key is the page's `name` field. Skip pages that have no hero section.
- **`sections`** — for icons, illustrations, or infographics attached to a specific section. Use sparingly in v1 — most plans will have zero section-scope entries. Only emit a section entry when a specific section's imagery need is not covered by the page hero.

### Per-row prompt construction

For each entry, write a `prompt` that:
- Names the subject concretely (what is in the image)
- Reflects the site's `design_intent.imagery_direction` and `colour_mood`
- Avoids brand markings, logos in the subject (unless the entry IS a logo), and text-on-image unless explicit
- Is self-contained — the image generator sees only this prompt, not the surrounding site context

### Worked example

```
"imagery": {
  "site": [
    {
      "key": "logo",
      "kind": "logo",
      "prompt": "A precise, technical logomark — geometric, restrained, no human figures, no text outside the wordmark itself"
    },
    {
      "key": "hero_canonical",
      "kind": "hero",
      "prompt": "A dramatic, high-contrast close-up of an industrial robotic gripper in soft directional lighting, neutral background"
    }
  ],
  "pages": {
    "about": [
      {
        "key": "hero_about",
        "kind": "hero",
        "prompt": "A wide-angle view of a modern automated production line, calm and orderly, blue and grey palette"
      },
      {
        "key": "illustration_team_values",
        "kind": "illustration",
        "prompt": "Stylised group of engineers collaborating around a workbench, no faces visible, mid-century technical illustration feel",
        "style_hints": {"medium": "line drawing", "mood": "collaborative"}
      }
    ],
    "tools": [
      {
        "key": "hero_tools",
        "kind": "hero",
        "prompt": "An engineering workspace abstraction — measurement instruments, technical drawings, soft daylight"
      }
    ]
  },
  "sections": {
    "index:2": [
      {
        "key": "icon_precision",
        "kind": "icon",
        "prompt": "Geometric icon representing precision engineering — single colour, sharp corners",
        "constraints": {"aspect": "1:1", "transparent_background": true}
      }
    ]
  }
}
```

## Strategy Guidance

If a domain strategy is available above, use it as strong input:
- The recommended site_type should guide the overall structure
- The recommended page_types should inform which pages you plan
- The revenue model should shape what conversion/lead-capture mechanisms you include
- The tone should influence your style_collection choice

You have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment but note why in strategy_notes.

Return JSON:
{
  "site_type": "from the strategy, roadmap, or your own assessment",
  "strategy_notes": "any notes on how you used or diverged from the strategy/roadmap",
  "pages": [
    {
      "name": "index",
      "title": "Page Title | Site Name",
      "page_type": "index",
      "nav_label": "Home",
      "nav_order": 1,
      "in_header": true,
      "in_footer": true,
      "sections": ["hero", "features", "testimonials", "call-to-action"]
    }
  ],
  "style_collection": "style-name",
  "needs_logo": true,
  "needs_images": true,
  "image_prompts": {
    "logo": "Description for logo generation",
    "hero_home": "Description for home hero image"
  },
  "imagery": {
    "site": [
      {"key": "logo", "kind": "logo", "prompt": "..."}
    ],
    "pages": {
      "index": [
        {"key": "hero_home", "kind": "hero", "prompt": "..."}
      ]
    },
    "sections": {}
  },
  "design_intent": {
    "style_direction": "professional-dark or modern-light or bold-creative",
    "colour_mood": "Description of colour feeling and why it fits the industry",
    "typography_mood": "Description of font personality",
    "imagery_direction": "What images should show",
    "layout_preference": "Layout style description",
    "avoid": ["Things to avoid in design"]
  },
  "content_direction": {
    "voice": "How the site should sound",
    "emphasis": "What to emphasise in content",
    "avoid_phrases": ["Phrases to never use"],
    "social_proof_style": "How to handle testimonials and proof",
    "blog_strategy": "Content strategy for blog if applicable"
  }
}

RULES:
1. Use component names from the Available Section Components list for sections arrays — UNLESS the roadmap specifies section_types, in which case use those
2. Every page MUST have a page_type from the canonical list
3. Pages with page_type entity-directory, entity-page, tool, blog-index, blog-post may have empty sections arrays
4. Content and index pages need sections arrays populated
5. Always include: index (home) and contact pages (unless the roadmap says otherwise)
6. Keep header navigation to 5-8 items maximum
7. Set needs_logo: true and needs_images: true (always)
8. Provide detailed image_prompts for logo and hero_home
9. Include design_intent with explicit colour mood, typography direction, and layout preferences
10. Include content_direction with voice, emphasis, and avoid_phrases tailored to the target audience
11. If the classification data includes content_features.news_feed.recommended = true, add "latest-news" to the homepage sections
12. When a roadmap is present, the pages and section_types from the current phase take precedence over your own page planning
13. Populate the `imagery` block per the Imagery Block rules above. At minimum: one site-scope `logo` entry, and one page-scope `hero` entry for each page whose `sections` array contains a hero-class component
14. `imagery` and `image_prompts` must be consistent: the site-scope `logo` entry's prompt should match `image_prompts.logo`; the page-scope `hero` entry for the index page should match `image_prompts.hero_home`
15. Use ONLY the allowed values for `kind` (logo|hero|illustration|icon|infographic). Section scope keys MUST follow `"<page_name>:<ordering>"` format with a colon. Wrong values fail DB CHECK constraints and the plan is rejected

Return ONLY valid JSON.$bsp_prompt$::text)
       ),
       updated_at = now()
 WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
   AND is_active = true;

-- Verify the new prompt contains both the new section marker and the
-- "imagery" key in the JSON skeleton. Abort the transaction if either
-- is missing — means the jsonb_set didn't land cleanly.
DO $verify$
DECLARE
new_prompt text;
BEGIN
SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}'
INTO new_prompt
FROM agent_definitions
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

IF new_prompt IS NULL THEN
        RAISE EXCEPTION 'prompt_template is NULL after update';
END IF;

    IF position('## Imagery Block' IN new_prompt) = 0 THEN
        RAISE EXCEPTION 'Imagery Block section not found in updated prompt_template';
END IF;

    IF position('"imagery":' IN new_prompt) = 0 THEN
        RAISE EXCEPTION '"imagery": key not found in updated prompt_template';
END IF;

    IF position('Rule 15' IN new_prompt) > 0 THEN
        RAISE EXCEPTION 'Unexpected literal "Rule 15" found — template body may have been double-processed';
END IF;

    RAISE NOTICE 'phase_2g_step3: prompt_template updated successfully (length: % chars)', length(new_prompt);
END
$verify$;

COMMIT;

---
-- correction to above

-- phase_2g_step3_planner_imagery_prompt.sql
--
-- Phase 2G step 3 — teach build-site-planner to emit a structured `imagery`
-- block alongside the legacy `image_prompts` map. Downstream
-- write_site_plan_action.flattenImageryBlock has been deployed dormant
-- since step 2; this migration is the first behavioural change that
-- populates site_plan_imagery rows.
--
-- Idempotent in effect: replaces the entire prompt_template; safe to re-run.
-- Wrapped in BEGIN/COMMIT with a post-write verification that aborts the
-- transaction if either the new "## Imagery Block" section or the
-- `"imagery":` key is missing from the resulting template.
--
-- Run AFTER taking the standard backup per doc 009:
--   CREATE TABLE agent_definitions_backup_20260512_pre_phase2g_step3_planner AS
--   SELECT * FROM agent_definitions
--   WHERE type = 'build-site-planner' AND is_active = true;

\set ON_ERROR_STOP on

BEGIN;

-- Sanity: target row still exists.
DO $check$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
          AND is_active = true
    ) THEN
        RAISE EXCEPTION 'build-site-planner row not found or inactive';
END IF;
END
$check$;

-- Replace the prompt_template field. Dollar-quoted to preserve quotes,
-- newlines, and braces in the template body. Tag $bsp_prompt$ chosen
-- to avoid collision with any text inside the prompt.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb($bsp_prompt$Plan a website for {{.input_data.domain}}.

## Research Data
Identity: {{.site_specs.specs.identity}}
Classification: {{.site_specs.specs.classification}}

## Domain Strategy
{{if .site_specs.specs.strategy}}{{.site_specs.specs.strategy}}{{else}}No strategy data available — use the briefing and classification to determine site structure.{{end}}

## Briefing Answers
{{.site_specs.specs.briefing}}

{{if .site_specs.specs.mission_brief}}## Mission
{{.site_specs.specs.mission_brief.text}}
{{end}}
{{if .site_specs.specs.roadmap_brief}}## Roadmap
IMPORTANT — ROADMAP OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the current phase below. For each page, use EXACTLY the section_types listed — even if they do not appear in the Available Section Components list above. Unknown section types are handled by the component selector downstream. Do NOT replace roadmap section_types with standard components. Do NOT invent additional pages. The roadmap is the authority for this site.

{{.site_specs.specs.roadmap_brief.text}}
{{end}}
## Available Section Components
You MUST use ONLY these exact component names in the "sections" arrays (unless the roadmap specifies section_types — those override):

{{range .available_components}}
- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}
{{end}}

## Available Style Collections
{{.available_styles}}

## Canonical Page Types

Use ONLY these page_type values, verbatim, lowercase, dash-separated:

| page_type | Description | Use for |
|-----------|-------------|----------|
| index | Home page | Always exactly one |
| content | Standard content page | About, services, contact, FAQ, etc |
| landing | Conversion-focused page | Lead capture, specific offers |
| entity-directory | Searchable directory of entities | Business listings, provider directories |
| entity-page | Individual entity profile | Single business/provider detail page |
| tool | Interactive tool or calculator | Cost calculators, comparison tools |
| blog-index | Blog/news listing page | Article index, news feed |
| blog-post | Individual blog article | Editorial content, guides |

Plan ONLY pages that are directly justified by the briefing, strategy, classification, or roadmap. Do NOT add speculative, demo, or example pages. Every page in your output must serve an explicit need surfaced by one of those inputs. If you don't have evidence for a page, leave it out — fewer well-justified pages are better than padding the count.

If a page doesn't fit any category cleanly, use `content` as the default. Do not invent new page_type values.

## Imagery Block

Return a structured `imagery` block alongside the legacy `image_prompts` map. The downstream image pipeline reads `imagery` as the source of truth; `image_prompts` is retained for backward compatibility during transition.

### Shape

`imagery` is a nested object with three optional sub-keys: `site`, `pages`, `sections`. Scope is implied by position in the nesting — entries do NOT carry `scope` or `scope_ref` fields themselves.

```
"imagery": {
  "site":     [ entry, entry, ... ],
  "pages":    { "<page_name>": [ entry, ... ], ... },
  "sections": { "<page_name>:<section_ordering>": [ entry, ... ], ... }
}
```

The `<section_ordering>` is the zero-indexed position of the section within that page's `sections` array (so `"index:0"` is the first section of the home page).

### Entry fields

| Field | Required | Notes |
|---|---|---|
| key | yes | short identifier, lowercase, underscore-separated (e.g. `logo`, `hero_about`, `illustration_team_values`). Unique within its scope. |
| kind | yes | one of: `logo`, `hero`, `illustration`, `icon`, `infographic`. No other values permitted. |
| prompt | yes | full generation prompt, one sentence to one paragraph. Reflect the site's `imagery_direction` and `colour_mood`. |
| style_hints | optional | JSON object, e.g. {"medium": "line drawing", "mood": "collaborative"} |
| constraints | optional | JSON object, e.g. {"aspect": "1:1", "transparent_background": true} |

### What to populate

- **`site`** — imagery that appears across pages or is brand-level. Always include one `logo` entry. Optionally include one site-wide brand `hero` or `illustration` entry. Two to three entries is typical.
- **`pages`** — one entry per page-specific image. Always include a `hero` entry for the `index` page, and a `hero` entry for every other page whose `sections` array contains a hero-class component. The map key is the page's `name` field. Skip pages that have no hero section.
- **`sections`** — for icons, illustrations, or infographics attached to a specific section. Use sparingly in v1 — most plans will have zero section-scope entries. Only emit a section entry when a specific section's imagery need is not covered by the page hero.

### Per-row prompt construction

For each entry, write a `prompt` that:
- Names the subject concretely (what is in the image)
- Reflects the site's `design_intent.imagery_direction` and `colour_mood`
- Avoids brand markings, logos in the subject (unless the entry IS a logo), and text-on-image unless explicit
- Is self-contained — the image generator sees only this prompt, not the surrounding site context

### Worked example

```
"imagery": {
  "site": [
    {
      "key": "logo",
      "kind": "logo",
      "prompt": "A precise, technical logomark — geometric, restrained, no human figures, no text outside the wordmark itself"
    },
    {
      "key": "hero_canonical",
      "kind": "hero",
      "prompt": "A dramatic, high-contrast close-up of an industrial robotic gripper in soft directional lighting, neutral background"
    }
  ],
  "pages": {
    "index": [
      {
        "key": "hero_home",
        "kind": "hero",
        "prompt": "A dramatic, high-contrast close-up of an industrial robotic gripper in soft directional lighting, neutral background"
      }
    ],
    "about": [
      {
        "key": "hero_about",
        "kind": "hero",
        "prompt": "A wide-angle view of a modern automated production line, calm and orderly, blue and grey palette"
      },
      {
        "key": "illustration_team_values",
        "kind": "illustration",
        "prompt": "Stylised group of engineers collaborating around a workbench, no faces visible, mid-century technical illustration feel",
        "style_hints": {"medium": "line drawing", "mood": "collaborative"}
      }
    ],
    "tools": [
      {
        "key": "hero_tools",
        "kind": "hero",
        "prompt": "An engineering workspace abstraction — measurement instruments, technical drawings, soft daylight"
      }
    ]
  },
  "sections": {
    "index:2": [
      {
        "key": "icon_precision",
        "kind": "icon",
        "prompt": "Geometric icon representing precision engineering — single colour, sharp corners",
        "constraints": {"aspect": "1:1", "transparent_background": true}
      }
    ]
  }
}
```

Note in the example above: `image_prompts.hero_home` would carry the same subject as the `pages.index[0]` entry's prompt, satisfying rule 14. The `pages.index` hero is REQUIRED whenever the index page has a hero section (which is nearly always).

## Strategy Guidance

If a domain strategy is available above, use it as strong input:
- The recommended site_type should guide the overall structure
- The recommended page_types should inform which pages you plan
- The revenue model should shape what conversion/lead-capture mechanisms you include
- The tone should influence your style_collection choice

You have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment but note why in strategy_notes.

Return JSON:
{
  "site_type": "from the strategy, roadmap, or your own assessment",
  "strategy_notes": "any notes on how you used or diverged from the strategy/roadmap",
  "pages": [
    {
      "name": "index",
      "title": "Page Title | Site Name",
      "page_type": "index",
      "nav_label": "Home",
      "nav_order": 1,
      "in_header": true,
      "in_footer": true,
      "sections": ["hero", "features", "testimonials", "call-to-action"]
    }
  ],
  "style_collection": "style-name",
  "needs_logo": true,
  "needs_images": true,
  "image_prompts": {
    "logo": "Description for logo generation",
    "hero_home": "Description for home hero image"
  },
  "imagery": {
    "site": [
      {"key": "logo", "kind": "logo", "prompt": "..."}
    ],
    "pages": {
      "index": [
        {"key": "hero_home", "kind": "hero", "prompt": "..."}
      ]
    },
    "sections": {}
  },
  "design_intent": {
    "style_direction": "professional-dark or modern-light or bold-creative",
    "colour_mood": "Description of colour feeling and why it fits the industry",
    "typography_mood": "Description of font personality",
    "imagery_direction": "What images should show",
    "layout_preference": "Layout style description",
    "avoid": ["Things to avoid in design"]
  },
  "content_direction": {
    "voice": "How the site should sound",
    "emphasis": "What to emphasise in content",
    "avoid_phrases": ["Phrases to never use"],
    "social_proof_style": "How to handle testimonials and proof",
    "blog_strategy": "Content strategy for blog if applicable"
  }
}

RULES:
1. Use component names from the Available Section Components list for sections arrays — UNLESS the roadmap specifies section_types, in which case use those
2. Every page MUST have a page_type from the canonical list
3. Pages with page_type entity-directory, entity-page, tool, blog-index, blog-post may have empty sections arrays
4. Content and index pages need sections arrays populated
5. Always include: index (home) and contact pages (unless the roadmap says otherwise)
6. Keep header navigation to 5-8 items maximum
7. Set needs_logo: true and needs_images: true (always)
8. Provide detailed image_prompts for logo and hero_home
9. Include design_intent with explicit colour mood, typography direction, and layout preferences
10. Include content_direction with voice, emphasis, and avoid_phrases tailored to the target audience
11. If the classification data includes content_features.news_feed.recommended = true, add "latest-news" to the homepage sections
12. When a roadmap is present, the pages and section_types from the current phase take precedence over your own page planning
13. Populate the `imagery` block per the Imagery Block rules above. At minimum: one site-scope `logo` entry, one page-scope `hero` entry under `pages.index`, and one page-scope `hero` entry for every other page whose `sections` array contains a hero-class component
14. `imagery` and `image_prompts` must be consistent: the site-scope `logo` entry's prompt should match `image_prompts.logo`; the `pages.index` `hero` entry's prompt should match `image_prompts.hero_home`
15. Use ONLY the allowed values for `kind` (logo|hero|illustration|icon|infographic). Section scope keys MUST follow `"<page_name>:<ordering>"` format with a colon. Wrong values fail DB CHECK constraints and the plan is rejected

Return ONLY valid JSON.$bsp_prompt$::text)
       ),
       updated_at = now()
 WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
   AND is_active = true;

-- Verify the new prompt contains both the new section marker and the
-- "imagery" key in the JSON skeleton. Abort the transaction if either
-- is missing — means the jsonb_set didn't land cleanly.
DO $verify$
DECLARE
new_prompt text;
BEGIN
SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}'
INTO new_prompt
FROM agent_definitions
WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

IF new_prompt IS NULL THEN
        RAISE EXCEPTION 'prompt_template is NULL after update';
END IF;

    IF position('## Imagery Block' IN new_prompt) = 0 THEN
        RAISE EXCEPTION 'Imagery Block section not found in updated prompt_template';
END IF;

    IF position('"imagery":' IN new_prompt) = 0 THEN
        RAISE EXCEPTION '"imagery": key not found in updated prompt_template';
END IF;

    IF position('"hero_home"' IN new_prompt) = 0 THEN
        RAISE EXCEPTION '"hero_home" reference not found — worked example may be malformed';
END IF;

    RAISE NOTICE 'phase_2g_step3: prompt_template updated successfully (length: % chars)', length(new_prompt);
END
$verify$;

COMMIT;

---
-- more tokens

-- phase_2g_step3_planner_maxtokens_bump.sql
--
-- Raise build-site-planner's plan_site step max_tokens from 4000 to 8000.
-- Phase 2G step 3 (the imagery block) enlarged required output: site-scope
-- logo + page-scope hero per page, on top of the existing pages/sections/
-- directives/image_prompts/design_intent/content_direction structure.
-- The prior 4000 cap truncated the LLM response on multi-page roadmaps
-- (15-page site for robot-hands.com), causing validate_site_plan to fail
-- with "unexpected end of JSON input" at the strategy_notes field.
--
-- Reversible. Backup taken per doc 009 convention.

\set ON_ERROR_STOP on

-- ── Backup (outside transaction) ──

CREATE TABLE agent_def_build_site_planner_backup_20260513_pre_phase2g_maxtokens AS
SELECT * FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active = true;

SELECT
    (SELECT COUNT(*) FROM agent_definitions
     WHERE type = 'build-site-planner' AND is_active = true) AS live,
    (SELECT COUNT(*) FROM agent_def_build_site_planner_backup_20260513_pre_phase2g_maxtokens) AS backup;

-- ── Migration ──

BEGIN;

-- Sanity: target row exists and has the field we're about to update
DO $check$
DECLARE
v_current_max int;
BEGIN
SELECT (default_config #>> '{workflow,steps,plan_site,config,ai_service,max_tokens}')::int
INTO v_current_max
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active = true;

IF v_current_max IS NULL THEN
        RAISE EXCEPTION 'plan_site.ai_service.max_tokens not found in default_config';
END IF;

    RAISE NOTICE 'Current max_tokens: %', v_current_max;
END
$check$;

-- Apply the bump. jsonb_set with create_missing=false (last arg) is what
-- we want here — we're updating an existing field, not creating one.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,ai_service,max_tokens}',
        to_jsonb(8000),
        false
                     ),
    updated_at = now()
WHERE type = 'build-site-planner'
  AND is_active = true;

-- Verify
DO $verify$
DECLARE
v_new_max int;
BEGIN
SELECT (default_config #>> '{workflow,steps,plan_site,config,ai_service,max_tokens}')::int
INTO v_new_max
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active = true;

IF v_new_max IS NULL THEN
        RAISE EXCEPTION 'max_tokens is NULL after update';
END IF;

    IF v_new_max <> 8000 THEN
        RAISE EXCEPTION 'max_tokens not 8000 after update; got %', v_new_max;
END IF;

    RAISE NOTICE 'phase_2g_step3 max_tokens: 4000 -> %', v_new_max;
END
$verify$;

COMMIT;

---

-- phase_2g_planner_imagery_prompt_decomposition.sql
--
-- Updates the build-site-planner prompt_template to:
--   - rename constraints.aspect → style_hints.aspect_ratio (item 4)
--   - add per-image decomposition guidance (item 25)
--   - strengthen icon-style prompt construction (item 23)
--   - demonstrate multi-entry sections in the worked example (items 23, 25)
--   - add explicit Rule 16 about one-entry-equals-one-image
--
-- Reference: planner_prompt_patch_imagery.md
--
-- Safe to re-apply (idempotent in effect — overwrites with the same content).
-- Old prompt value is saved to migration_backups for rollback.

BEGIN;

-- ----------------------------------------------------------------------------
-- Backup infrastructure (created once, reused by all migrations)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS migration_backups (
                                                 id serial PRIMARY KEY,
                                                 migration_name text NOT NULL,
                                                 applied_at timestamptz NOT NULL DEFAULT now(),
    target_table text,
    target_id text,
    old_value jsonb,
    notes text
    );

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT
    'phase_2g_planner_imagery_prompt_decomposition',
    'agent_definitions',
    id::text,
    jsonb_build_object(
            'prompt_template', default_config #> '{workflow,steps,plan_site,config,prompt_template}'
    ),
    'Pre-update value of build-site-planner prompt_template (Imagery Block changes for items 4, 23, 25)'
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active = true;

-- ----------------------------------------------------------------------------
-- Apply: replace prompt_template with the updated version
-- ----------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb($PROMPT$Plan a website for {{.input_data.domain}}.

## Research Data
Identity: {{.site_specs.specs.identity}}
Classification: {{.site_specs.specs.classification}}

## Domain Strategy
{{if .site_specs.specs.strategy}}{{.site_specs.specs.strategy}}{{else}}No strategy data available — use the briefing and classification to determine site structure.{{end}}

## Briefing Answers
{{.site_specs.specs.briefing}}

{{if .site_specs.specs.mission_brief}}## Mission
{{.site_specs.specs.mission_brief.text}}
{{end}}
{{if .site_specs.specs.roadmap_brief}}## Roadmap
IMPORTANT — ROADMAP OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the current phase below. For each page, use EXACTLY the section_types listed — even if they do not appear in the Available Section Components list above. Unknown section types are handled by the component selector downstream. Do NOT replace roadmap section_types with standard components. Do NOT invent additional pages. The roadmap is the authority for this site.

{{.site_specs.specs.roadmap_brief.text}}
{{end}}
## Available Section Components
You MUST use ONLY these exact component names in the "sections" arrays (unless the roadmap specifies section_types — those override):

{{range .available_components}}
- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}
{{end}}

## Available Style Collections
{{.available_styles}}

## Canonical Page Types

Use ONLY these page_type values, verbatim, lowercase, dash-separated:

| page_type | Description | Use for |
|-----------|-------------|----------|
| index | Home page | Always exactly one |
| content | Standard content page | About, services, contact, FAQ, etc |
| landing | Conversion-focused page | Lead capture, specific offers |
| entity-directory | Searchable directory of entities | Business listings, provider directories |
| entity-page | Individual entity profile | Single business/provider detail page |
| tool | Interactive tool or calculator | Cost calculators, comparison tools |
| blog-index | Blog/news listing page | Article index, news feed |
| blog-post | Individual blog article | Editorial content, guides |

Plan ONLY pages that are directly justified by the briefing, strategy, classification, or roadmap. Do NOT add speculative, demo, or example pages. Every page in your output must serve an explicit need surfaced by one of those inputs. If you don't have evidence for a page, leave it out — fewer well-justified pages are better than padding the count.

If a page doesn't fit any category cleanly, use `content` as the default. Do not invent new page_type values.

## Imagery Block

Return a structured `imagery` block alongside the legacy `image_prompts` map. The downstream image pipeline reads `imagery` as the source of truth; `image_prompts` is retained for backward compatibility during transition.

### Shape

`imagery` is a nested object with three optional sub-keys: `site`, `pages`, `sections`. Scope is implied by position in the nesting — entries do NOT carry `scope` or `scope_ref` fields themselves.

```
"imagery": {
  "site":     [ entry, entry, ... ],
  "pages":    { "<page_name>": [ entry, ... ], ... },
  "sections": { "<page_name>:<section_ordering>": [ entry, ... ], ... }
}
```

The `<section_ordering>` is the zero-indexed position of the section within that page's `sections` array (so `"index:0"` is the first section of the home page).

### Entry fields

| Field | Required | Notes |
|---|---|---|
| key | yes | short identifier, lowercase, underscore-separated (e.g. `logo`, `hero_about`, `illustration_team_values`). Unique within its scope. |
| kind | yes | one of: `logo`, `hero`, `illustration`, `icon`, `infographic`. No other values permitted. |
| prompt | yes | full generation prompt, one sentence to one paragraph. Reflect the site's `imagery_direction` and `colour_mood`. |
| style_hints | optional | JSON object. Use `aspect_ratio` (e.g. "1:1", "16:9", "4:3") to control image proportions. Other keys (`medium`, `mood`, `palette`) are advisory and may be ignored. |
| constraints | optional | JSON object. Currently informational only — does not influence generation. Reserved for future use (anticipated: per-provider validation, content safety modes). |

### What to populate

- **`site`** — imagery that appears across pages or is brand-level. Always include one `logo` entry. Optionally include one site-wide brand `hero` or `illustration` entry. Two to three entries is typical.
- **`pages`** — one entry per page-specific image. Always include a `hero` entry for the `index` page, and a `hero` entry for every other page whose `sections` array contains a hero-class component. The map key is the page's `name` field. Skip pages that have no hero section.
- **`sections`** — for icons, illustrations, or infographics attached to a specific section. Use sparingly in v1 — most plans will have zero section-scope entries. Only emit a section entry when a specific section's imagery need is not covered by the page hero.

**Each entry produces exactly ONE image.** If a concept requires multiple images (e.g., six icons representing six gripper actuation types, or three illustrations representing three values), emit one entry per image with a distinct `key` for each. The image model interprets a single prompt as a single image — describing multiple images in one prompt produces a multi-panel composition that is unusable. Err toward over-decomposing rather than under-decomposing: a few unused icons are cheaper than one botched multi-panel image.

### Per-row prompt construction

For each entry, write a `prompt` that:
- **Describes ONE image.** Never use "set of", "multiple", "various", "a series of", or counting words ("six", "three") that imply a composition. Each entry is its own image.
- Names the subject concretely (what is in the image)
- Reflects the site's `design_intent.imagery_direction` and `colour_mood`
- For icons specifically: emphasise "single", "flat", "minimal", "line illustration", "plain background" — these style words help the model produce icon-appropriate output rather than photorealistic renders
- Avoids brand markings, logos in the subject (unless the entry IS a logo), and text-on-image unless explicit
- Is self-contained — the image generator sees only this prompt, not the surrounding site context

### Worked example

```
"imagery": {
  "site": [
    {
      "key": "logo",
      "kind": "logo",
      "prompt": "A precise, technical logomark — geometric, restrained, no human figures, no text outside the wordmark itself"
    },
    {
      "key": "hero_canonical",
      "kind": "hero",
      "prompt": "A dramatic, high-contrast close-up of an industrial robotic gripper in soft directional lighting, neutral background"
    }
  ],
  "pages": {
    "index": [
      {
        "key": "hero_home",
        "kind": "hero",
        "prompt": "A dramatic, high-contrast close-up of an industrial robotic gripper in soft directional lighting, neutral background"
      }
    ],
    "about": [
      {
        "key": "hero_about",
        "kind": "hero",
        "prompt": "A wide-angle view of a modern automated production line, calm and orderly, blue and grey palette"
      },
      {
        "key": "illustration_team_values",
        "kind": "illustration",
        "prompt": "Stylised group of engineers collaborating around a workbench, no faces visible, mid-century technical illustration feel",
        "style_hints": {"medium": "line drawing", "mood": "collaborative"}
      }
    ],
    "tools": [
      {
        "key": "hero_tools",
        "kind": "hero",
        "prompt": "An engineering workspace abstraction — measurement instruments, technical drawings, soft daylight"
      }
    ]
  },
  "sections": {
    "index:2": [
      {
        "key": "icon_precision",
        "kind": "icon",
        "prompt": "A single minimalist flat icon representing precision engineering — line illustration, geometric, sharp corners, single dark colour on plain background, no shadows, no photorealism",
        "style_hints": {"aspect_ratio": "1:1"}
      },
      {
        "key": "icon_speed",
        "kind": "icon",
        "prompt": "A single minimalist flat icon representing fast cycle speed — line illustration, dynamic geometric form, single dark colour on plain background, no shadows, no photorealism",
        "style_hints": {"aspect_ratio": "1:1"}
      },
      {
        "key": "icon_reliability",
        "kind": "icon",
        "prompt": "A single minimalist flat icon representing process reliability — line illustration, balanced geometric form, single dark colour on plain background, no shadows, no photorealism",
        "style_hints": {"aspect_ratio": "1:1"}
      }
    ]
  }
}
```

Note in the example above: `image_prompts.hero_home` would carry the same subject as the `pages.index[0]` entry's prompt, satisfying rule 14. The `pages.index` hero is REQUIRED whenever the index page has a hero section (which is nearly always). The three icon entries in `sections."index:2"` demonstrate the key decomposition principle: each conceptually-distinct image gets its own entry — never describe multiple images in a single prompt.

## Strategy Guidance

If a domain strategy is available above, use it as strong input:
- The recommended site_type should guide the overall structure
- The recommended page_types should inform which pages you plan
- The revenue model should shape what conversion/lead-capture mechanisms you include
- The tone should influence your style_collection choice

You have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment but note why in strategy_notes.

Return JSON:
{
  "site_type": "from the strategy, roadmap, or your own assessment",
  "strategy_notes": "any notes on how you used or diverged from the strategy/roadmap",
  "pages": [
    {
      "name": "index",
      "title": "Page Title | Site Name",
      "page_type": "index",
      "nav_label": "Home",
      "nav_order": 1,
      "in_header": true,
      "in_footer": true,
      "sections": ["hero", "features", "testimonials", "call-to-action"]
    }
  ],
  "style_collection": "style-name",
  "needs_logo": true,
  "needs_images": true,
  "image_prompts": {
    "logo": "Description for logo generation",
    "hero_home": "Description for home hero image"
  },
  "imagery": {
    "site": [
      {"key": "logo", "kind": "logo", "prompt": "..."}
    ],
    "pages": {
      "index": [
        {"key": "hero_home", "kind": "hero", "prompt": "..."}
      ]
    },
    "sections": {}
  },
  "design_intent": {
    "style_direction": "professional-dark or modern-light or bold-creative",
    "colour_mood": "Description of colour feeling and why it fits the industry",
    "typography_mood": "Description of font personality",
    "imagery_direction": "What images should show",
    "layout_preference": "Layout style description",
    "avoid": ["Things to avoid in design"]
  },
  "content_direction": {
    "voice": "How the site should sound",
    "emphasis": "What to emphasise in content",
    "avoid_phrases": ["Phrases to never use"],
    "social_proof_style": "How to handle testimonials and proof",
    "blog_strategy": "Content strategy for blog if applicable"
  }
}

RULES:
1. Use component names from the Available Section Components list for sections arrays — UNLESS the roadmap specifies section_types, in which case use those
2. Every page MUST have a page_type from the canonical list
3. Pages with page_type entity-directory, entity-page, tool, blog-index, blog-post may have empty sections arrays
4. Content and index pages need sections arrays populated
5. Always include: index (home) and contact pages (unless the roadmap says otherwise)
6. Keep header navigation to 5-8 items maximum
7. Set needs_logo: true and needs_images: true (always)
8. Provide detailed image_prompts for logo and hero_home
9. Include design_intent with explicit colour mood, typography direction, and layout preferences
10. Include content_direction with voice, emphasis, and avoid_phrases tailored to the target audience
11. If the classification data includes content_features.news_feed.recommended = true, add "latest-news" to the homepage sections
12. When a roadmap is present, the pages and section_types from the current phase take precedence over your own page planning
13. Populate the `imagery` block per the Imagery Block rules above. At minimum: one site-scope `logo` entry, one page-scope `hero` entry under `pages.index`, and one page-scope `hero` entry for every other page whose `sections` array contains a hero-class component
14. `imagery` and `image_prompts` must be consistent: the site-scope `logo` entry's prompt should match `image_prompts.logo`; the `pages.index` `hero` entry's prompt should match `image_prompts.hero_home`
15. Use ONLY the allowed values for `kind` (logo|hero|illustration|icon|infographic). Section scope keys MUST follow `"<page_name>:<ordering>"` format with a colon. Wrong values fail DB CHECK constraints and the plan is rejected
16. Each entry in `imagery` produces exactly ONE image. NEVER describe multiple images in a single prompt. If a section conceptually needs N icons (e.g., six gripper types, three pricing tiers), emit N separate entries with distinct `key` values. Phrases like "set of", "multiple", "various", "a series of", or counting words ("six", "three") cause the image model to produce a multi-panel composition that is unusable. Over-decomposing is cheap (a few unused icons); under-decomposing is expensive (manual cleanup of botched output).

Return ONLY valid JSON.$PROMPT$::text),
    false
  )
WHERE type = 'build-site-planner' AND is_active = true;

-- ----------------------------------------------------------------------------
-- Verify: confirm key strings are present in the updated prompt
-- ----------------------------------------------------------------------------
SELECT
    type,
    length(default_config #>> '{workflow,steps,plan_site,config,prompt_template}') AS prompt_length,
    position('Each entry produces exactly ONE image' IN (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')) > 0 AS has_decomposition_rule,
    position('style_hints.aspect_ratio' IN (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')) > 0 OR
    position('"aspect_ratio"' IN (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')) > 0 AS has_aspect_ratio,
    position('16. Each entry' IN (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')) > 0 AS has_rule_16,
    position('"constraints": {"aspect"' IN (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')) = 0 AS no_old_constraints_aspect
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active = true;

COMMIT;

-- Expected verification output:
--   prompt_length: ~7500-8500 chars
--   has_decomposition_rule: t
--   has_aspect_ratio: t
--   has_rule_16: t
--   no_old_constraints_aspect: t
--
-- If any of these are wrong, ROLLBACK and investigate before re-applying.
-- The previous prompt is recoverable from migration_backups.

-- ----------------------------------------------------------------------------
-- Rollback procedure (if needed)
-- ----------------------------------------------------------------------------
-- BEGIN;
-- UPDATE agent_definitions
-- SET default_config = jsonb_set(
--     default_config,
--     '{workflow,steps,plan_site,config,prompt_template}',
--     (SELECT old_value->'prompt_template'
--      FROM migration_backups
--      WHERE migration_name = 'phase_2g_planner_imagery_prompt_decomposition'
--      ORDER BY applied_at DESC LIMIT 1)
--   )
-- WHERE type = 'build-site-planner' AND is_active = true;
-- COMMIT;

---
-- icon background transparency workaround

-- planner_icon_background_fix.sql
--
-- Stop icons being planned with "transparent background" / "plain background".
-- Image models can't produce true alpha — they paint a transparency
-- checkerboard into the RGB pixels instead (confirmed on icon_cycle_time:
-- mode=RGB, has_alpha=False, 1024x1024, painted grey/white checker).
--
-- DECISION (option 2 — embrace the box): icons are generated on a flat, solid,
-- selectable grey background and sit inside a styled container ("chip") on the
-- page. No keying, no alpha, nothing fragile. Dark-grey line on light-grey
-- background so icon and background never merge; explicit hex so the CSS chip
-- can match or deliberately contrast.
--
-- Edits the build-site-planner LLM prompt_template (two substrings):
--   1. The "Per-row prompt construction" icon-guidance bullet.
--   2. The three worked-example icon entries (one shared substring → all three).
--
-- METHOD: read prompt_template as text (#>>), replace() the two fragments,
-- write back via to_jsonb(). Only the changed fragments are specified — the
-- ~8KB template is otherwise untouched.
--
-- Schema: column default_config, key type='build-site-planner',
--         path {workflow,steps,plan_site,config,prompt_template}.

BEGIN;

-- 0. Snapshot before mutating.
SELECT snapshot_agent('build-site-planner', 'icon background: transparent/plain -> flat selectable grey (embrace the chip)');

-- 1. Sanity: confirm the two target fragments are present (expect both true).
SELECT
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%"line illustration", "plain background" — these style words%' AS has_guidance_fragment,
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%single dark colour on plain background, no shadows, no photorealism%' AS has_worked_example_fragment
FROM agent_definitions
WHERE type = 'build-site-planner' AND deleted_at IS NULL;

-- 2. Apply both replacements.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
                replace(
                        replace(
                                default_config #>> '{workflow,steps,plan_site,config,prompt_template}',
                            -- (1) guidance bullet tail
                                ', "plain background" — these style words help the model produce icon-appropriate output rather than photorealistic renders',
                                '; specify a flat solid light grey (#EEEEEE) background with a darker grey (#4A4A4A) line icon — one single uniform background colour, no gradients, no shadows, no checkerboard pattern, no transparency, no photorealism. Icons are placed inside a styled container ("chip") on the page, so an opaque flat light-grey background is correct and expected — do NOT request or imply transparency'
                        ),
                    -- (2) worked-example shared substring (matches all three icon entries)
                        'single dark colour on plain background, no shadows, no photorealism',
                        'a darker grey (#4A4A4A) line on a flat solid light grey (#EEEEEE) background, one single uniform background colour, no gradients, no shadows, no checkerboard, no transparency, no photorealism'
                )
        )
                     )
WHERE type = 'build-site-planner' AND deleted_at IS NULL;

-- 3. Verify: old fragments gone, new guidance present.
SELECT
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%plain background%'   AS still_has_plain_background_should_be_f,
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%transparent%'        AS still_mentions_transparent,
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%#EEEEEE%'            AS has_grey_hex_should_be_t,
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%no checkerboard%'    AS has_no_checkerboard_should_be_t
FROM agent_definitions
WHERE type = 'build-site-planner' AND deleted_at IS NULL;
-- Expect: still_has_plain_background = f
--         still_mentions_transparent = f  (the only "transparent" refs were the icon ones;
--                                          if other parts of the template legitimately use
--                                          the word this may be t — eyeball if so)
--         has_grey_hex = t
--         has_no_checkerboard = t

-- 4. Spot-check the actual edited region reads correctly.
SELECT substring(
               default_config #>> '{workflow,steps,plan_site,config,prompt_template}'
  FROM '#EEEEEE.{0,180}'
) AS edited_region_preview
FROM agent_definitions
WHERE type = 'build-site-planner' AND deleted_at IS NULL;

-- COMMIT;  -- uncomment when verify output looks right
ROLLBACK;   -- safe default

---
-- adoption page similarity

-- 052_planner_reads_realised_state.sql
-- Doc 029 Phase 1: make build-site-planner read the realised (adopted) pages
-- and converge on them, instead of planning a generic skeleton from identity alone.
--
-- Root cause (diagnosed 2026-05-19, FOCUS_planner_ignores_adopted_state.md):
--   plan_site's input_fields are [input_data, site_specs, available_components,
--   available_styles] — the realised pages are NOT among them. The planner had
--   no idea adoption already built 20 pages, so it invented a 9-page generic
--   skeleton (flat games/tools pages, renamed tool dups, a "post" placeholder).
--
-- Four changes, all to the build-site-planner definition:
--   1. New step load_existing_pages (query_database) reads pages for the site
--   2. plan_site gains existing_pages input + a "preserve exactly" prompt block
--   3. plan_site max_tokens 4000 -> 8000 (room for ~30 pages of output)
--   4. validate_plan max_pages 20 -> 40 (don't truncate a faithful plan)
--
-- No new Go code: load_existing_pages uses the same query_database action that
-- load_components / load_styles already use. params resolves site_record.site_id
-- to $1.
--
-- Unaffected: from-scratch (non-adopted) sites. existing_pages is empty on first
-- plan, the {{else}} branch tells the planner to plan from the brief, and
-- sync_pages creates the pages as before. On a re-plan it converges on the
-- previously-planned pages, which is also desirable (no churn).

BEGIN;

-- Snapshot first (revertable via revert_agent('build-site-planner'))
SELECT snapshot_agent('build-site-planner',
                      'Phase 1: planner reads realised pages (052) — converge instead of re-propose');

-- ---------------------------------------------------------------------------
-- Change 1a: rewire ensure_site -> load_existing_pages (was -> load_components)
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,ensure_site,next_step}',
        '"load_existing_pages"'::jsonb,
        true
                     )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Change 1b: add the load_existing_pages step (next_step -> load_components)
-- pages liveness column is `status = 'active'` (not deleted_at).
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_existing_pages}',
        $STEP$
            {
            "action": "query_database",
        "config": {
                "query": "SELECT name, page_type, url, title, nav_label, in_header, in_footer FROM pages WHERE site_id = $1 AND status = 'active' ORDER BY name",
        "params": ["site_record.site_id"],
        "output_format": "array"
    },
            "next_step": "load_components",
            "description": "Load already-realised pages (e.g. from adoption) so the planner converges on them rather than re-proposing a generic skeleton",
            "output_field": "existing_pages"
        }
        $STEP$::jsonb,
        true
    )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Change 2a: add existing_pages to plan_site.input_fields
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,input_fields}',
        (default_config->'workflow'->'steps'->'plan_site'->'config'->'input_fields')
            || '["existing_pages"]'::jsonb,
        true
                     )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Change 2b: inject the "preserve exactly" block into the prompt, immediately
-- before the "## Available Section Components" marker (appears exactly once).
-- Uses replace() so we only author the new block, not the whole 4KB prompt.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
                replace(
                        default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
                        '## Available Section Components',
                        $INJECT$## Existing Pages — ALREADY BUILT, PRESERVE EXACTLY

{{if .existing_pages}}This site has already been built or adopted. The pages listed below ALREADY EXIST. You MUST include every one of them in your output with the EXACT same name, page_type, and url shown. Do NOT rename them. Do NOT change their page_type. Do NOT emit alternative or sibling versions of them. For example: if "games-index" exists, do not add a separate flat "games" page; if "tool-lanchester-sim" exists, do not rename it to "tool-lanchester-combat-calculator". You MAY add genuinely new pages if the briefing, strategy, or roadmap justifies them, but never duplicate, replace, or rename an existing page. Do not invent placeholder or example pages such as "post".

Existing pages:
{{range .existing_pages}}- name: {{.name}} | page_type: {{.page_type}} | url: {{.url}}
{{end}}
{{else}}No existing pages on this site yet — plan the structure from the brief below.
{{end}}

## Available Section Components$INJECT$
            )
        ),
        true
    )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Change 3: raise plan_site max_tokens 4000 -> 8000
-- build-site-planner has NO top-level ai_service, so the step-level
-- ai_service.max_tokens IS the value the chassis reads (verified via the
-- config-location audit). Setting it here takes effect.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,ai_service,max_tokens}',
        '8000'::jsonb,
        true
                     )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Change 4: raise validate_plan max_pages 20 -> 40
-- A faithful adopted plan can exceed 20 pages (gamesdesign has 27). Without
-- this, validate_plan would truncate and drop real adopted pages.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,validate_plan,config,max_pages}',
        '40'::jsonb,
        true
                     )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Verification (run after COMMIT)
-- ---------------------------------------------------------------------------
-- 1. New step exists and is wired
--    SELECT default_config->'workflow'->'steps'->'ensure_site'->>'next_step' AS ensure_next,
--           default_config->'workflow'->'steps'->'load_existing_pages'->>'next_step' AS lep_next,
--           default_config->'workflow'->'steps'->'load_existing_pages'->'config'->>'query' AS lep_query
--    FROM agent_definitions WHERE type = 'build-site-planner' AND is_active = true;
--    -- expect: ensure_next=load_existing_pages, lep_next=load_components, query present
--
-- 2. plan_site input + max_tokens
--    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->'input_fields' AS inputs,
--           default_config->'workflow'->'steps'->'plan_site'->'config'->'ai_service'->>'max_tokens' AS max_tok
--    FROM agent_definitions WHERE type = 'build-site-planner' AND is_active = true;
--    -- expect: inputs includes "existing_pages", max_tok=8000
--
-- 3. prompt injection landed exactly once
--    SELECT (length(default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template')
--            - length(replace(default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
--                              'PRESERVE EXACTLY', ''))) / length('PRESERVE EXACTLY') AS inject_count
--    FROM agent_definitions WHERE type = 'build-site-planner' AND is_active = true;
--    -- expect: inject_count = 1
--
-- 4. validate_plan max_pages
--    SELECT default_config->'workflow'->'steps'->'validate_plan'->'config'->>'max_pages'
--    FROM agent_definitions WHERE type = 'build-site-planner' AND is_active = true;
--    -- expect: 40

COMMIT;

--

-- adoption faithfulness for 90 days using new lock
-- CORRECTED 2026-07-21 (bugs_open/051): the 90-day "new lock" (branch (b) of the
-- 054 design below) was NEVER built — it is absent from the live
-- load_existing_pages query and has zero rows behind it fleet-wide. What is live
-- is the minimal first-plan-only variant further down; there is NO 90-day window
-- and no per-page timed lock. See bugs_open/051.
-- 052_planner_reads_realised_state.sql
-- Doc 029 Phase 1: make build-site-planner read the realised (adopted) pages
-- and converge on them, instead of planning a generic skeleton from identity alone.
--
-- Root cause (diagnosed 2026-05-19, FOCUS_planner_ignores_adopted_state.md):
--   plan_site's input_fields are [input_data, site_specs, available_components,
--   available_styles] — the realised pages are NOT among them. The planner had
--   no idea adoption already built 20 pages, so it invented a 9-page generic
--   skeleton (flat games/tools pages, renamed tool dups, a "post" placeholder).
--
-- Four changes, all to the build-site-planner definition:
--   1. New step load_existing_pages (query_database) reads pages for the site
--   2. plan_site gains existing_pages input + a "preserve exactly" prompt block
--   3. plan_site max_tokens 4000 -> 16000 (room for ~80-100 pages of output)
--   4. validate_plan max_pages 20 -> 80 (don't truncate; 80-page ceiling)
--
-- No new Go code: load_existing_pages uses the same query_database action that
-- load_components / load_styles already use. params resolves site_record.site_id
-- to $1.
--
-- Unaffected: from-scratch (non-adopted) sites. existing_pages is empty on first
-- plan, the {{else}} branch tells the planner to plan from the brief, and
-- sync_pages creates the pages as before. On a re-plan it converges on the
-- previously-planned pages, which is also desirable (no churn).

BEGIN;

-- Snapshot first (revertable via revert_agent('build-site-planner'))
SELECT snapshot_agent('build-site-planner',
                      'Phase 1: planner reads realised pages (052) — converge instead of re-propose');

-- ---------------------------------------------------------------------------
-- Change 1a: rewire ensure_site -> load_existing_pages (was -> load_components)
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,ensure_site,next_step}',
        '"load_existing_pages"'::jsonb,
        true
                     )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Change 1b: add the load_existing_pages step (next_step -> load_components)
-- pages liveness column is `status = 'active'` (not deleted_at).
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_existing_pages}',
        $STEP$
            {
            "action": "query_database",
        "config": {
                "query": "SELECT name, page_type, url, title, nav_label, in_header, in_footer FROM pages WHERE site_id = $1 AND status = 'active' ORDER BY name",
        "params": ["site_record.site_id"],
        "output_format": "array"
    },
            "next_step": "load_components",
            "description": "Load already-realised pages (e.g. from adoption) so the planner converges on them rather than re-proposing a generic skeleton",
            "output_field": "existing_pages"
        }
        $STEP$::jsonb,
        true
    )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Change 2a: add existing_pages to plan_site.input_fields
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,input_fields}',
        (default_config->'workflow'->'steps'->'plan_site'->'config'->'input_fields')
            || '["existing_pages"]'::jsonb,
        true
                     )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Change 2b: inject the "preserve exactly" block into the prompt, immediately
-- before the "## Available Section Components" marker (appears exactly once).
-- Uses replace() so we only author the new block, not the whole 4KB prompt.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
                replace(
                        default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
                        '## Available Section Components',
                        $INJECT$## Existing Pages — ALREADY BUILT, PRESERVE EXACTLY

{{if .existing_pages}}This site has already been built or adopted. The pages listed below ALREADY EXIST. You MUST include every one of them in your output with the EXACT same name, page_type, and url shown. Do NOT rename them. Do NOT change their page_type. Do NOT emit alternative or sibling versions of them. For example: if "games-index" exists, do not add a separate flat "games" page; if "tool-lanchester-sim" exists, do not rename it to "tool-lanchester-combat-calculator". You MAY add genuinely new pages if the briefing, strategy, or roadmap justifies them, but never duplicate, replace, or rename an existing page. Do not invent placeholder or example pages such as "post".

Existing pages:
{{range .existing_pages}}- name: {{.name}} | page_type: {{.page_type}} | url: {{.url}}
{{end}}
{{else}}No existing pages on this site yet — plan the structure from the brief below.
{{end}}

## Available Section Components$INJECT$
            )
        ),
        true
    )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Change 3: raise plan_site max_tokens 4000 -> 16000
-- build-site-planner has NO top-level ai_service, so the step-level
-- ai_service.max_tokens IS the value the chassis reads (verified via the
-- config-location audit). Setting it here takes effect.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,ai_service,max_tokens}',
        '16000'::jsonb,
        true
                     )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Change 4: raise validate_plan max_pages 20 -> 80
-- A faithful adopted plan can exceed 20 pages (gamesdesign has 27); larger
-- builds target up to 80 pages in a single plan_site call. Without
-- this, validate_plan would truncate and drop real adopted pages.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,validate_plan,config,max_pages}',
        '80'::jsonb,
        true
                     )
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

-- ---------------------------------------------------------------------------
-- Verification (run after COMMIT)
-- ---------------------------------------------------------------------------
-- 1. New step exists and is wired
--    SELECT default_config->'workflow'->'steps'->'ensure_site'->>'next_step' AS ensure_next,
--           default_config->'workflow'->'steps'->'load_existing_pages'->>'next_step' AS lep_next,
--           default_config->'workflow'->'steps'->'load_existing_pages'->'config'->>'query' AS lep_query
--    FROM agent_definitions WHERE type = 'build-site-planner' AND is_active = true;
--    -- expect: ensure_next=load_existing_pages, lep_next=load_components, query present
--
-- 2. plan_site input + max_tokens
--    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->'input_fields' AS inputs,
--           default_config->'workflow'->'steps'->'plan_site'->'config'->'ai_service'->>'max_tokens' AS max_tok
--    FROM agent_definitions WHERE type = 'build-site-planner' AND is_active = true;
--    -- expect: inputs includes "existing_pages", max_tok=8000
--
-- 3. prompt injection landed exactly once
--    SELECT (length(default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template')
--            - length(replace(default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template',
--                              'PRESERVE EXACTLY', ''))) / length('PRESERVE EXACTLY') AS inject_count
--    FROM agent_definitions WHERE type = 'build-site-planner' AND is_active = true;
--    -- expect: inject_count = 1
--
-- 4. validate_plan max_pages
--    SELECT default_config->'workflow'->'steps'->'validate_plan'->'config'->>'max_pages'
--    FROM agent_definitions WHERE type = 'build-site-planner' AND is_active = true;
--    -- expect: 80

COMMIT;

---
-- adoption locks
--
-- CORRECTED 2026-07-21 (bugs_open/051): the FULL 054 immediately below (branches
-- (a)+(b), with the "90-day window" note) is a DESIGN that was never built.
-- Branch (b) — the per-page timed preserve-directive — is absent from the live
-- load_existing_pages query and has zero rows behind it (462 site_plan_directives,
-- locked_by NULL on every one; nothing writes them). What is LIVE is the minimal
-- first-plan-only variant ("054 minimal", further down): adoption_locked = true
-- ONLY on a site's first plan, false on every re-plan after. There is no 90-day
-- lock, no per-page lock, no expiry. Read the aspirational text below as design
-- history, not as live behaviour.

-- 054_load_existing_pages_adoption_locked.sql
-- Update build-site-planner.load_existing_pages so each returned page carries
-- an `adoption_locked` boolean. The validate_site_plan convergence layer
-- (v3_site_actions.go) force-preserves only pages where adoption_locked = true.
--
-- adoption_locked is true when EITHER:
--   (a) there is NO current plan for the site yet — the first plan after
--       adoption. (Only adopted sites have pages before their first plan;
--       from-scratch sites have none until sync_pages runs, so this branch
--       only marks genuinely-adopted pages.) OR
--   (b) the current plan has a LIVE adoption preserve-directive for this page
--       (locked_by='adoption', lock_type='timed', not yet expired).
--
-- After the 90-day window, (b) goes false (expiry), (a) no longer applies
-- (a plan exists), so adoption_locked = false and the site develops normally.
--
-- Depends on: 053 (lock_type/lock_expires_at columns) and the write_site_plan
-- patch (emits + locks the preserve-directives). Until those land, branch (b)
-- finds no rows and branch (a) still correctly protects the first plan.

BEGIN;

SELECT snapshot_agent('build-site-planner',
                      'load_existing_pages exposes adoption_locked (054)');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_existing_pages,config,query}',
        to_jsonb($Q$
                     SELECT p.name, p.page_type, p.url, p.title, p.nav_label, p.in_header, p.in_footer,
                 CASE
                     WHEN NOT EXISTS (
                         SELECT 1 FROM site_plans sp
                         WHERE sp.site_id = p.site_id AND sp.is_current = true
                     ) THEN true
                     WHEN EXISTS (
                         SELECT 1
                         FROM site_plans sp
                                  JOIN site_plan_directives d ON d.plan_id = sp.id
                         WHERE sp.site_id = p.site_id
                           AND sp.is_current = true
                           AND d.scope = 'page'
                           AND d.scope_ref = p.name
                           AND d.category = 'preserve'
                           AND d.locked_by = 'adoption'
                           AND d.lock_type = 'timed'
                           AND d.lock_expires_at IS NOT NULL
                           AND d.lock_expires_at > NOW()
                     ) THEN true
                     ELSE false
                     END AS adoption_locked
FROM pages p
WHERE p.site_id = $1 AND p.status = 'active'
ORDER BY p.name
$Q$::text),
        true
                     ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

COMMIT;

-- ---------------------------------------------------------------------------
-- Verification (after COMMIT)
-- ---------------------------------------------------------------------------
-- SELECT default_config->'workflow'->'steps'->'load_existing_pages'->'config'->>'query'
-- FROM agent_definitions WHERE type = 'build-site-planner' AND is_active = true;
-- -- confirm the query contains adoption_locked and the two CASE branches.
--
-- Functional check after a fresh adoption + first plan run: every adopted page
-- should come back adoption_locked = true. Re-run after 90+ days (or with a
-- back-dated lock_expires_at) and they should be false.

    -- backup
snapshot_agent
--------------------------------------
f263eaa1-61e1-446e-9410-648e12b7875b
(1 row)


---
-- amend to above split into parts

-- 054_load_existing_pages_first_plan_only.sql
--
-- MINIMAL, LOCK-FREE variant of load_existing_pages. Gets faithful first-pass
-- adoption working WITHOUT any of the timed-lock machinery (no 053 columns, no
-- write_site_plan locking, no transferDirectiveLocks change).
--
-- Use this INSTEAD OF 054_load_existing_pages_adoption_locked.sql for now. The
-- full version references d.lock_type / d.lock_expires_at on site_plan_directives,
-- which only exist after 053 — applying it before 053 would make
-- load_existing_pages fail with "column lock_type does not exist".
--
-- BEHAVIOUR with this minimal version:
--   adoption_locked = true  ONLY when the site has no current plan yet.
--   That is uniquely the first plan after adoption (from-scratch sites have no
--   pages until the planner's own sync_pages runs, so "no current plan + pages
--   exist" only happens for adopted sites).
--
--   => First plan after adoption: every existing (adopted) page is preserved
--      faithfully by the convergence layer.
--   => Every subsequent plan: a current plan exists, so adoption_locked = false
--      for all pages, convergence is a no-op, and the site develops normally
--      (the planner / improvement loop may add, edit, restructure, delete).
--
-- This fixes the doc-029 spurious/duplicate-pages problem (which occurred on the
-- first plan after adoption). It does NOT give a 90-day window — after the first
-- plan the site is immediately free to evolve. The 90-day window is added later
-- by swapping in 054_load_existing_pages_adoption_locked.sql once 053 + the
-- write_site_plan lock patch are applied.
--
-- PREREQUISITE: 052 must be applied (it creates the load_existing_pages step and
-- raises max_pages -> 80, max_tokens -> 16000). If 052 is not applied, apply it
-- first; this migration only rewrites the step's query.

BEGIN;

SELECT snapshot_agent('build-site-planner',
                      'load_existing_pages adoption_locked, first-plan-only / lock-free (054 minimal)');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_existing_pages,config,query}',
        to_jsonb($Q$
                     SELECT p.name, p.page_type, p.url, p.title, p.nav_label, p.in_header, p.in_footer,
                 CASE
                     WHEN NOT EXISTS (
                         SELECT 1 FROM site_plans sp
                         WHERE sp.site_id = p.site_id AND sp.is_current = true
                     ) THEN true
                     ELSE false
                     END AS adoption_locked
FROM pages p
WHERE p.site_id = $1 AND p.status = 'active'
ORDER BY p.name
$Q$::text),
        true
                     ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL AND is_active = true;

COMMIT;

-- ---------------------------------------------------------------------------
-- Verification
-- ---------------------------------------------------------------------------
-- 1. The step query updated:
--    SELECT default_config->'workflow'->'steps'->'load_existing_pages'->'config'->>'query'
--    FROM agent_definitions WHERE type='build-site-planner' AND is_active=true;
--    -- contains adoption_locked, the NOT EXISTS site_plans branch, no lock_type.
--
-- 2. Functional: trigger a fresh adoption (e.g. gamesdesign.co.uk) and let the
--    planner run its first plan. Expect every adopted page returned with
--    adoption_locked = true, and the final plan to contain all adopted pages
--    with no spurious/duplicate pages (no flat games vs games-index, etc).
--
-- 3. Re-plan the same site (any later planner run): a current plan now exists,
--    so adoption_locked = false for all pages and convergence is a no-op — the
--    planner is free to change the structure. This is expected and correct for
--    the lock-free variant.
--
-- ROLLBACK if needed: SELECT revert_agent('build-site-planner');

---
-- add design and composition back in

-- ============================================================================
-- build_pipeline_wiring.sql
--
-- Wires the three build-path fixes into agent workflows:
--   1. build-site-planner: add emit_design  (emit_design_items)  after reconcile
--   2. build-site-planner: add emit_imagery (emit_imagery_items) after emit_design
--   3. rerender-pages:     add mark_site_deployed (update_site_status=deployed)
--
-- These edit agent_definitions.default_config (jsonb). Apply per your normal
-- agent-update / versioning process (snapshot if you version defs).
--
-- IMPORTANT — INSPECT FIRST. These paths assume the backup snapshot's step
-- names. Confirm the LIVE structure before applying:
--
--   SELECT jsonb_object_keys(default_config->'workflow'->'steps')
--   FROM agent_definitions WHERE type = 'build-site-planner';
--   -- expect: ..., reconcile_site_plan, complete  (NOT already emit_design)
--
--   SELECT default_config->'workflow'->'steps'->'reconcile_site_plan'->>'next_step'
--   FROM agent_definitions WHERE type = 'build-site-planner';   -- expect: complete
--
--   SELECT jsonb_object_keys(default_config->'workflow'->'steps')
--   FROM agent_definitions WHERE type = 'rerender-pages';
--   -- expect: ..., create_rerender_items, complete
--
--   SELECT default_config->'workflow'->'steps'->'create_rerender_items'->>'next_step'
--   FROM agent_definitions WHERE type = 'rerender-pages';        -- expect: complete
--
-- If default_config is `json` not `jsonb`, cast: default_config::jsonb ... ::json.
-- Prereq: emit_design_items + emit_imagery_items registered in the action
-- registry (see registry_entries.txt) and the chassis rebuilt.
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1 + 2. build-site-planner: reconcile_site_plan -> emit_design -> emit_imagery -> complete
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                jsonb_set(
                        default_config,
                        '{workflow,steps,emit_design}',
                        $$ {
                    "action": "emit_design_items",
                        "config": { "site_id": "input_data.site_id" },
                    "next_step": "emit_imagery",
                    "description": "Plan-time design trigger: queue needs_composition + needs_design (guarded on style_collection_id IS NULL)",
                    "output_field": "design_items"
                } $$::jsonb,
                true
            ),
                '{workflow,steps,emit_imagery}',
                $$ {
                "action": "emit_imagery_items",
                "config": { "site_id": "input_data.site_id" },
                "next_step": "complete",
                "description": "Plan-time imagery trigger: queue needs_imagery from the current plan's site_plan_imagery rows",
                "output_field": "imagery_items"
            } $$::jsonb,
            true
        ),
        '{workflow,steps,reconcile_site_plan,next_step}',
        '"emit_design"'::jsonb,
        false
                     ),
    updated_at = now()
WHERE type = 'build-site-planner';

-- ---------------------------------------------------------------------------
-- 3. rerender-pages: create_rerender_items -> mark_site_deployed -> complete
--    Marks the site deployed once the terminal rerender has committed. Pages
--    were already built+deployed by page-build-handler, so the site is live.
--    Idempotent: re-stamps last_deployed_at on every rerender.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,mark_site_deployed}',
                $$ {
                "action": "update_site_status",
                "config": {
                    "status": "deployed",
                "deployed_at": "now",
                "site_id_field": "input_data.site_id"
            },
                "next_step": "complete",
                "description": "Mark site deployed after rerender + commit",
                "output_field": "site_updated"
            } $$::jsonb,
            true
        ),
        '{workflow,steps,create_rerender_items,next_step}',
        '"mark_site_deployed"'::jsonb,
        false
                     ),
    updated_at = now()
WHERE type = 'rerender-pages';

-- ---------------------------------------------------------------------------
-- POST-APPLY VERIFICATION (run after a fresh submit-domain build)
-- ---------------------------------------------------------------------------
-- Design + imagery items should appear for the new site:
--   SELECT item_type, handler_agent, status, source, created_by, priority
--   FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
--   WHERE s.domain = '<new-domain>'
--     AND item_type IN ('needs_composition','needs_design','needs_imagery')
--   ORDER BY priority;
--
-- After the terminal rerender, the site should be deployed:
--   SELECT domain, status, style_collection_id, last_deployed_at
--   FROM sites WHERE domain = '<new-domain>';
-- ============================================================================

---
-- same as above
-- ============================================================================
-- build_pipeline_wiring.sql
--
-- Wires the three build-path fixes into agent workflows:
--   1. build-site-planner: add emit_design  (emit_design_items)  after reconcile
--   2. build-site-planner: add emit_imagery (emit_imagery_items) after emit_design
--   3. rerender-pages:     add mark_site_deployed (update_site_status=deployed)
--
-- These edit agent_definitions.default_config (jsonb). Apply per your normal
-- agent-update / versioning process (snapshot if you version defs).
--
-- IMPORTANT — INSPECT FIRST. These paths assume the backup snapshot's step
-- names. Confirm the LIVE structure before applying:
--
--   SELECT jsonb_object_keys(default_config->'workflow'->'steps')
--   FROM agent_definitions WHERE type = 'build-site-planner';
--   -- expect: ..., reconcile_site_plan, complete  (NOT already emit_design)
--
--   SELECT default_config->'workflow'->'steps'->'reconcile_site_plan'->>'next_step'
--   FROM agent_definitions WHERE type = 'build-site-planner';   -- expect: complete
--
--   SELECT jsonb_object_keys(default_config->'workflow'->'steps')
--   FROM agent_definitions WHERE type = 'rerender-pages';
--   -- expect: ..., create_rerender_items, complete
--
--   SELECT default_config->'workflow'->'steps'->'create_rerender_items'->>'next_step'
--   FROM agent_definitions WHERE type = 'rerender-pages';        -- expect: complete
--
-- If default_config is `json` not `jsonb`, cast: default_config::jsonb ... ::json.
-- Prereq: emit_design_items + emit_imagery_items registered in the action
-- registry (see registry_entries.txt) and the chassis rebuilt.
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 1 + 2. build-site-planner: reconcile_site_plan -> emit_design -> emit_imagery -> complete
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                jsonb_set(
                        default_config,
                        '{workflow,steps,emit_design}',
                        $$ {
                    "action": "emit_design_items",
                        "config": { "site_id": "input_data.site_id" },
                    "next_step": "emit_imagery",
                    "description": "Plan-time design trigger: queue needs_composition + needs_design (guarded on style_collection_id IS NULL)",
                    "output_field": "design_items"
                } $$::jsonb,
                true
            ),
                '{workflow,steps,emit_imagery}',
                $$ {
                "action": "emit_imagery_items",
                "config": { "site_id": "input_data.site_id" },
                "next_step": "complete",
                "description": "Plan-time imagery trigger: queue needs_imagery from the current plan's site_plan_imagery rows",
                "output_field": "imagery_items"
            } $$::jsonb,
            true
        ),
        '{workflow,steps,reconcile_site_plan,next_step}',
        '"emit_design"'::jsonb,
        false
                     ),
    updated_at = now()
WHERE type = 'build-site-planner';

-- ---------------------------------------------------------------------------
-- 3. rerender-pages: create_rerender_items -> mark_site_deployed -> complete
--    Marks the site deployed once the terminal rerender has committed. Pages
--    were already built+deployed by page-build-handler, so the site is live.
--    Idempotent: re-stamps last_deployed_at on every rerender.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,mark_site_deployed}',
                $$ {
                "action": "update_site_status",
                "config": {
                    "status": "deployed",
                "deployed_at": "now",
                "site_id_field": "input_data.site_id"
            },
                "next_step": "complete",
                "description": "Mark site deployed after rerender + commit",
                "output_field": "site_updated"
            } $$::jsonb,
            true
        ),
        '{workflow,steps,create_rerender_items,next_step}',
        '"mark_site_deployed"'::jsonb,
        false
                     ),
    updated_at = now()
WHERE type = 'rerender-pages';

-- ---------------------------------------------------------------------------
-- POST-APPLY VERIFICATION (run after a fresh submit-domain build)
-- ---------------------------------------------------------------------------
-- Design + imagery items should appear for the new site:
--   SELECT item_type, handler_agent, status, source, created_by, priority
--   FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
--   WHERE s.domain = '<new-domain>'
--     AND item_type IN ('needs_composition','needs_design','needs_imagery')
--   ORDER BY priority;
--
-- After the terminal rerender, the site should be deployed:
--   SELECT domain, status, style_collection_id, last_deployed_at
--   FROM sites WHERE domain = '<new-domain>';
-- ============================================================================

-- ============================================================================
-- 4. build-site-planner output_contract: drop stale "build_items", reflect the
--    real terminal outputs (this workflow uses reconcile_site_plan, not
--    write_build_items, and now also emits design + imagery items).
-- ============================================================================
UPDATE agent_definitions
SET output_contract = jsonb_set(
        output_contract,
        '{produces}',
        $$ {
            "db_sync":          "Pages synced to database",
        "site_plan":        "Validated site plan",
        "plan_written":     "Plan written to site_plans + pages/sections/directives/imagery",
        "reconcile_result": "needs_page work items + terminal needs_rerender emitted for the delta",
        "design_items":     "needs_composition + needs_design emitted when no composition is installed",
        "imagery_items":    "needs_imagery emitted from the plan's site_plan_imagery rows"
    } $$::jsonb,
        true
    ),
    updated_at = now()
WHERE type = 'build-site-planner';

-- Optional: surface the new step outputs in the workflow result alongside the
-- existing ones (cosmetic — they already run; this just returns their counts).
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,complete,config,output_fields}',
        $$ ["site_plan","plan_written","db_sync","reconcile_result","design_items","imagery_items"] $$::jsonb,
        false
                     ),
    updated_at = now()
WHERE type = 'build-site-planner';
-- ============================================================================

The wiring: reconcile_site_plan.next_step → emit_design → emit_imagery → complete; emit_design calls emit_design_items and emit_imagery calls emit_imagery_items, both with site_id: input_data.site_id; output_contract.produces dropped build_items and now lists reconcile_result/design_items/imagery_items; complete.output_fields includes the two new ones. Updated 2026-05-26 11:06. The flow is right.
---

                                                                                                                                                                                                                                                                                                                                        clients_db=# SELECT snapshot_agent('build-site-planner', 'write_site_plan step description: add site_plan_imagery');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,write_site_plan,description}',
        to_jsonb('Write validated plan to site_plans + site_plan_pages + site_plan_sections + site_plan_directives + site_plan_imagery; transfer HITL locks (directives + imagery) from previous current plan'::text),
        false)
WHERE type = 'build-site-planner'
  AND is_active = true;
NOTICE:  Snapshot captured: type=build-site-planner, source_version=1, source_id=f263eaa1-61e1-446e-9410-648e12b7875b, reason=write_site_plan step description: add site_plan_imagery
            snapshot_agent
--------------------------------------
 f263eaa1-61e1-446e-9410-648e12b7875b
(1 row)


---
      -- migration_planner_topic_sibling_rule.sql
-- ----------------------------------------------------------------------------
-- OPTIONAL STOPGAP — build-site-planner prompt hardening.
--
-- Diagnosis (llm_call_log, plan_site @ 2026-06-03 20:25:22): the planner WAS
-- handed the adopted guide pages in existing_pages (e.g.
-- "guide-economy-basics | page_type: blog-post | url: /blog/guide-economy-basics.html")
-- and still emitted a parallel "economy-basics" blog-post. The existing
-- "do NOT emit alternative or sibling versions" rule only illustrates a hub/flat
-- pair (games/games-index) and a rename (tool-lanchester-sim); the LLM did not
-- generalise it to a guide-prefix sibling.
--
-- This adds an explicit topic/prefix/role-duplicate rule with the concrete
-- guide-economy-basics example.
--
-- NOTE ON LAYER: this is a PROMPT nudge, not a guarantee — the LLM can still
-- ignore it. The durable fix is a deterministic guard in validate_site_plan /
-- write_site_plan that drops a planned page whose topic STEM (role-prefix
-- stripped, as CanonicalisePage already computes via TrimPrefix) collides with
-- an existing page. Treat this migration as a bridge until that lands.
--
-- One quote-free, newline-free replace() on default_config::text -> jsonb
-- (the cast validates the JSON; a malformed result aborts the txn). Anchor
-- verified unique in the live build-site-planner config (pre-check below).
-- ----------------------------------------------------------------------------

BEGIN;

-- SNAPSHOT (rollback safety): full backup of the agent row before the edit.
CREATE TABLE IF NOT EXISTS agent_definitions_bak_planner_sibling AS
SELECT * FROM agent_definitions WHERE type = 'build-site-planner';

-- Pre-check: the anchor must be present exactly once (column = 1). If not, STOP.
SELECT
    (length(default_config::text)
        - length(replace(default_config::text,
                         'but never duplicate, replace, or rename an existing page.', '')))
        / length('but never duplicate, replace, or rename an existing page.')
        AS anchor_count
FROM agent_definitions
WHERE type = 'build-site-planner';

-- Apply: append the topic/prefix/role-duplicate rule after the anchor sentence.
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        'but never duplicate, replace, or rename an existing page.',
        'but never duplicate, replace, or rename an existing page. This also bars TOPIC duplicates across different names, slugs, prefixes, or page_types: if an existing page already covers a subject (including an existing guide- or blog- page), do NOT plan another page on that same subject. For example, if guide-economy-basics exists, do NOT add an economy-basics blog-post and do NOT add an economy-basics guide; reuse the existing page. A genuinely new page must be a NEW subject not covered by ANY existing page, never a re-slugged, re-prefixed, or re-typed version of one.'
                     )::jsonb,
    updated_at = NOW()
WHERE type = 'build-site-planner';

-- Verify: the new rule is present (expect t).
SELECT (default_config::text LIKE '%This also bars TOPIC duplicates%') AS rule_present
FROM agent_definitions
WHERE type = 'build-site-planner';

COMMIT;

-- ----------------------------------------------------------------------------
-- ROLLBACK (if needed): restore the whole config from the snapshot.
--   UPDATE agent_definitions a
--   SET default_config = b.default_config, updated_at = NOW()
--   FROM agent_definitions_bak_planner_sibling b WHERE a.type = b.type;
-- Drop the snapshot once satisfied: DROP TABLE agent_definitions_bak_planner_sibling;
-- ----------------------------------------------------------------------------

-- migration_load_existing_pages_carry_fields.sql
-- ----------------------------------------------------------------------------
-- PROPER FIX (a of 2) — make load_existing_pages expose the fields the
-- adoption-union must carry, so activating the convergence does not regress
-- adopted content.
--
-- CONTEXT: 054 is live — load_existing_pages already emits adoption_locked via
-- the first-plan branch (no current plan), so reconcilePlanWithRealised runs on
-- the first post-adoption pass. Pass A unions LLM-omitted adopted pages via
-- normaliseRealisedToPlanPage; sync_pages -> upsertPage then does
-- "sections = EXCLUDED.sections", "meta_description = EXCLUDED.meta_description",
-- "nav_order = EXCLUDED.nav_order" (nav_label is COALESCE-preserved, so safe).
-- The current query does NOT select sections/meta_description/nav_order, so the
-- union has nothing to carry and the upsert clobbers the adopted page's real
-- values to empty. This adds those three columns to the SELECT.
--
-- Pairs with the Go change (b): normaliseRealisedToPlanPage now reads
-- rm["sections"] (jsonb arrives as a JSON string via query_database),
-- rm["meta_description"], rm["nav_order"]. Both must land together.
--
-- One quote-free, value-preserving replace() on default_config::text -> jsonb.
-- Anchor verified against the live query (the load_existing_pages SELECT is the
-- only one aliased `p.`). Pre-check below must report anchor_count = 1.
-- ----------------------------------------------------------------------------

BEGIN;

-- SNAPSHOT (rollback safety).
CREATE TABLE IF NOT EXISTS agent_definitions_bak_lep_carry AS
SELECT * FROM agent_definitions WHERE type = 'build-site-planner';

-- Pre-check: the anchor must appear exactly once (column = 1). If not, STOP.
SELECT
    (length(default_config::text)
        - length(replace(default_config::text,
                         'p.nav_label, p.in_header, p.in_footer,', '')))
        / length('p.nav_label, p.in_header, p.in_footer,')
        AS anchor_count
FROM agent_definitions
WHERE type = 'build-site-planner';

-- Apply: add sections, meta_description, nav_order to the SELECT list.
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        'p.nav_label, p.in_header, p.in_footer,',
        'p.nav_label, p.in_header, p.in_footer, p.sections, p.meta_description, p.nav_order,'
                     )::jsonb,
    updated_at = NOW()
WHERE type = 'build-site-planner';

-- Verify: the query now selects the three fields (expect t), and still emits
-- adoption_locked (expect t).
SELECT
    (default_config::text LIKE '%p.sections, p.meta_description, p.nav_order%') AS carries_fields,
    (default_config::text LIKE '%adoption_locked%')                            AS still_emits_lock
FROM agent_definitions
WHERE type = 'build-site-planner';

COMMIT;

-- ----------------------------------------------------------------------------
-- ROLLBACK:
--   UPDATE agent_definitions a SET default_config = b.default_config,
--          updated_at = NOW()
--   FROM agent_definitions_bak_lep_carry b WHERE a.type = b.type;
--   DROP TABLE agent_definitions_bak_lep_carry;
-- ----------------------------------------------------------------------------