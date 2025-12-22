-- ===========================================================================
-- SITE PLANNER AGENT
-- File: 044_site_planner_agent.sql
-- ===========================================================================
-- Plans the site structure: which pages, which components, style choices.
-- Uses LLM to analyze brief and match to available components.
-- ===========================================================================

BEGIN;

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    is_active,
    status,
    version,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    input_contract,
    output_contract,
    default_config
) VALUES (
             'site-planner',
             'Site Planner',
             'Analyzes brief and plans site structure: pages, components, style collection, asset needs. Single LLM call to create comprehensive plan.',
             'specialist',
             true,
             'active',
             1,
             '["planning", "site-architecture", "component-selection"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.575',
             '{
                 "requests": {"cpu": "100m", "memory": "256Mi"},
                 "limits": {"cpu": "500m", "memory": "512Mi"}
             }'::jsonb,
             '{
                 "error": "system.errors.{type}",
                 "process": "system.agent.{type}.process",
                 "response": "system.responses.{type}"
             }'::jsonb,
             '{
                 "port": 8080,
                 "liveness_path": "/health",
                 "readiness_path": "/ready",
                 "initial_delay_seconds": 15
             }'::jsonb,
             -- Input contract
             '{
                 "expects": {
                     "input_data": {
                         "domain": "string - the domain name",
                         "objective": "string - what the site should achieve"
                     },
                     "reviewed_brief": "object with company_name, services, about_us, etc",
                     "site_record": "object with site_id, domain"
                 },
                 "required": ["input_data.domain", "reviewed_brief"]
             }'::jsonb,
             -- Output contract
             '{
                 "produces": {
                     "pages": "array of {name, title, nav_label, nav_order, sections[], in_header, in_footer}",
                     "style_collection": "string - name of style collection to use",
                     "needs_logo": "boolean - whether to generate a logo",
                     "needs_images": "boolean - whether to generate hero/background images",
                     "image_prompts": "object with prompts for each needed image"
                 }
             }'::jsonb,
             -- Workflow
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 120,
                 "workflow": {
                     "start_step": "load_available_components",
                     "steps": {
                         "load_available_components": {
                             "action": "query_database",
                             "description": "Load available section components from database",
                             "config": {
                                 "query": "SELECT name, display_name, function, category, semantic_tags, description FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true ORDER BY category, name",
                                 "output_format": "array"
                             },
                             "next_step": "load_style_collections",
                             "output_field": "available_components"
                         },

                         "load_style_collections": {
                             "action": "query_database",
                             "description": "Load available style collections",
                             "config": {
                                 "query": "SELECT name, display_name, category, description FROM style_collections WHERE is_active = true ORDER BY name",
                                 "output_format": "array"
                             },
                             "next_step": "plan_site",
                             "output_field": "available_styles"
                         },

                         "plan_site": {
                             "action": "execute_llm_prompt",
                             "description": "LLM creates comprehensive site plan",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5-20250514",
                                     "max_tokens": 4000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["input_data", "reviewed_brief", "available_components", "available_styles"],
                                 "output_format": "json",
                                 "prompt_template": "You are a website architect. Plan a website for {{input_data.domain}}.\n\n## Brief\nCompany: {{reviewed_brief.company_name}}\nTagline: {{reviewed_brief.tagline}}\nAbout: {{reviewed_brief.about_us}}\nServices: {{reviewed_brief.services}}\nLeadership Team: {{reviewed_brief.leadership_team}}\nCase Studies: {{reviewed_brief.case_studies}}\nContact Email: {{reviewed_brief.contact_email}}\nContact Phone: {{reviewed_brief.contact_phone}}\nHas Blog: {{reviewed_brief.has_blog}}\nHas Careers: {{reviewed_brief.has_careers}}\nTarget Audience: {{reviewed_brief.target_audience}}\nTone: {{reviewed_brief.tone}}\n\n## Available Section Components\n{{available_components}}\n\n## Available Style Collections\n{{available_styles}}\n\n## Task\nCreate a site plan. Return JSON:\n\n```json\n{\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Home | Company Name\",\n      \"nav_label\": \"Home\",\n      \"nav_order\": 1,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"sections\": [\"hero-gradient\", \"services-grid\", \"testimonials-carousel\", \"cta-contact\"]\n    },\n    {\n      \"name\": \"about\",\n      \"title\": \"About Us | Company Name\",\n      \"nav_label\": \"About\",\n      \"nav_order\": 2,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"sections\": [\"hero-simple\", \"about-content\", \"team-grid\", \"values-section\"]\n    }\n  ],\n  \"style_collection\": \"professional-dark\",\n  \"needs_logo\": true,\n  \"needs_images\": true,\n  \"image_prompts\": {\n    \"logo\": \"Modern minimalist logo for [industry] company, abstract geometric, professional\",\n    \"hero_home\": \"Professional [industry] imagery, modern office environment, abstract\",\n    \"hero_about\": \"Team collaboration, diverse professionals, warm lighting\"\n  }\n}\n```\n\nGuidelines:\n- Use exact component names from the available list where possible\n- If no matching component exists, use a generic name like \"content-block\" or \"custom-section\"\n- Match style to the industry (consulting = professional-dark, creative = minimal-light)\n- Keep navigation to 5-8 items in header\n- Always include: index, about, services/products, contact\n- Add blog/insights page if has_blog is true\n- Add careers page if has_careers is true\n- For image prompts, be specific to the industry and company"
                             },
                             "next_step": "validate_plan",
                             "output_field": "llm_plan"
                         },

                         "validate_plan": {
                             "action": "validate_site_plan",
                             "description": "Validate and normalize the site plan",
                             "config": {
                                 "plan_field": "llm_plan",
                                 "ensure_pages": ["index", "contact"],
                                 "max_pages": 20,
                                 "validate_components": true,
                                 "default_style": "professional-dark"
                             },
                             "next_step": "complete",
                             "output_field": "validated_plan"
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Return the validated site plan",
                             "config": {
                                 "output_field": "validated_plan"
                             }
                         }
                     }
                 }
             }'::jsonb
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              version = EXCLUDED.version,
                              default_config = EXCLUDED.default_config,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = now();

COMMIT;