-- Webdesign Agent Definition (Final)
--
-- Generates CSS stylesheets for sites.
-- Uses file_path config in git_commit (requires patch_01_git_commit_file_path.go)
-- No container config (handled by spawn_actions.go)

/*INSERT INTO agent_definitions (
    type,
    name,
    description,
    role,
    default_config,
    can_delegate,
    tags,
    input_contract,
    output_contract
) VALUES (
             'webdesign-agent',
             'Web Design Agent',
             'Generates CSS stylesheets for sites. Accepts site_context or loads from DB. Analyzes design requirements and generates production CSS.',
             'specialist',
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "check_site_context",
                     "steps": {
                         "check_site_context": {
                             "action": "conditional",
                             "config": {
                                 "condition": "input_data.site_context.domain != null AND input_data.site_context.domain != ''",
                                 "then_step": "use_provided_context",
                                 "else_step": "load_site_context"
                             },
                             "description": "Check if site_context was provided"
                         },
                         "use_provided_context": {
                             "action": "transform_data",
                             "config": {
                                 "source_field": "input_data.site_context",
                                 "output_key": "site_context"
                             },
                             "description": "Use provided site_context",
                             "next_step": "analyze_design",
                             "output_field": "site_context"
                         },
                         "load_site_context": {
                             "action": "load_site_for_design",
                             "config": {
                                 "site_id_field": "input_data.site_id",
                                 "domain_field": "input_data.domain",
                                 "include_pages": true,
                                 "include_style_collection": true
                             },
                             "description": "Load site from database",
                             "next_step": "analyze_design",
                             "output_field": "site_context"
                         },
                         "analyze_design": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 2000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_context"],
                                 "output_format": "json",
                                 "prompt_template": "You are a web design expert. Analyze the site and output a design specification.\n\n## Site\nDomain: {{.site_context.domain}}\nCompany: {{.site_context.company_name}}\nIndustry: {{if .site_context.industry}}{{.site_context.industry}}{{else}}professional services{{end}}\nTagline: {{.site_context.tagline}}\n\n## Existing Style\n{{if .site_context.color_palette}}Colors: {{.site_context.color_palette | tojson}}{{end}}\n{{if .site_context.typography}}Typography: {{.site_context.typography | tojson}}{{end}}\n\n## Components\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\nReturn ONLY valid JSON:\n{\n  \"color_scheme\": {\n    \"primary\": \"#hex\",\n    \"secondary\": \"#hex\",\n    \"accent\": \"#hex\",\n    \"background\": \"#fff\",\n    \"surface\": \"#f8f9fa\",\n    \"text\": \"#333\",\n    \"text_muted\": \"#666\",\n    \"border\": \"#e2e8f0\"\n  },\n  \"typography\": {\n    \"font_family\": \"-apple-system, sans-serif\",\n    \"heading_font\": \"inherit\",\n    \"base_size\": \"16px\",\n    \"line_height\": \"1.6\"\n  },\n  \"spacing\": {\n    \"section_padding\": \"4rem\",\n    \"container_max_width\": \"1200px\"\n  },\n  \"components_to_style\": [\"hero\", \"services-grid\"],\n  \"design_notes\": \"notes\"\n}"
                             },
                             "description": "Generate design spec",
                             "next_step": "generate_css",
                             "output_field": "design_spec"
                         },
                         "generate_css": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 8000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_context", "design_spec"],
                                 "output_format": "text",
                                 "prompt_template": "Generate a complete production CSS stylesheet.\n\n## Design Spec\n{{.design_spec.result}}\n\n## Components\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\n## Requirements\n1. CSS variables in :root based on color_scheme\n2. Minimal reset (box-sizing, margin/padding)\n3. Typography (body, h1-h6, p, a)\n4. Layout (.container max-width centered, .section padding)\n5. Buttons (.btn, .btn-primary, .btn-secondary, .btn-large)\n6. Hero (.hero min-70vh, centered, white text with shadow)\n7. Grids (.services-grid responsive 1-2-3 cols, .team-grid, .stats-grid)\n8. Cards (.service-item, .team-member with shadow and hover)\n9. CTA (.cta-section with background)\n10. Mobile-first responsive (768px, 1024px)\n11. Focus states, smooth transitions\n\nOutput ONLY CSS. No markdown. Start with :root {"
                             },
                             "description": "Generate CSS",
                             "next_step": "deploy_css",
                             "output_field": "generated_css"
                         },
                         "deploy_css": {
                             "action": "git_commit",
                             "config": {
                                 "domain_field": "site_context.domain",
                                 "content_field": "generated_css.result",
                                 "file_path": "assets/css/styles.css",
                                 "commit_message": "Update stylesheet via webdesign-agent"
                             },
                             "description": "Deploy CSS to git",
                             "next_step": "check_update_db",
                             "output_field": "css_deployed"
                         },
                         "check_update_db": {
                             "action": "conditional",
                             "config": {
                                 "condition": "site_context.site_id != null AND site_context.site_id != ''",
                                 "then_step": "update_site",
                                 "else_step": "complete"
                             },
                             "description": "Check if we should update DB"
                         },
                         "update_site": {
                             "action": "update_site_content",
                             "config": {
                                 "site_id_field": "site_context.site_id",
                                 "merge": true,
                                 "content_field": "design_spec.result"
                             },
                             "description": "Store design spec",
                             "next_step": "complete",
                             "output_field": "site_updated"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["design_spec", "css_deployed", "site_context"]
                             }
                         }
                     }
                 }
             }',
             false,
             ARRAY['design', 'css', 'styling', 'specialist'],
             '{"required": [], "optional": ["site_id", "domain", "site_context"]}',
             '{"produces": {"css_deployed": "git result", "design_spec": "design spec"}}'
         )
    ON CONFLICT (type) DO UPDATE SET
    name = EXCLUDED.name,
                              description = EXCLUDED.description,
                              default_config = EXCLUDED.default_config,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();*/

--- fix for above:
-- Webdesign Agent Definition (Final)
--
-- Generates CSS stylesheets for sites.
-- Uses file_path config in git_commit (requires patch_01_git_commit_file_path.go)
-- No container config (handled by spawn_actions.go)

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    input_contract,
    output_contract
) VALUES (
             'webdesign-agent',
             'Web Design Agent',
             'Generates CSS stylesheets for sites. Accepts site_context or loads from DB. Analyzes design requirements and generates production CSS.',
             'specialist',
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "check_site_context",
                     "steps": {
                         "check_site_context": {
                             "action": "conditional",
                             "config": {
                                 "condition": "input_data.site_context.domain != null AND input_data.site_context.domain != ''",
                                 "then_step": "use_provided_context",
                                 "else_step": "load_site_context"
                             },
                             "description": "Check if site_context was provided"
                         },
                         "use_provided_context": {
                             "action": "transform_data",
                             "config": {
                                 "source_field": "input_data.site_context",
                                 "output_key": "site_context"
                             },
                             "description": "Use provided site_context",
                             "next_step": "analyze_design",
                             "output_field": "site_context"
                         },
                         "load_site_context": {
                             "action": "load_site_for_design",
                             "config": {
                                 "site_id_field": "input_data.site_id",
                                 "domain_field": "input_data.domain",
                                 "include_pages": true,
                                 "include_style_collection": true
                             },
                             "description": "Load site from database",
                             "next_step": "analyze_design",
                             "output_field": "site_context"
                         },
                         "analyze_design": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 2000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_context"],
                                 "output_format": "json",
                                 "prompt_template": "You are a web design expert. Analyze the site and output a design specification.\n\n## Site\nDomain: {{.site_context.domain}}\nCompany: {{.site_context.company_name}}\nIndustry: {{if .site_context.industry}}{{.site_context.industry}}{{else}}professional services{{end}}\nTagline: {{.site_context.tagline}}\n\n## Existing Style\n{{if .site_context.color_palette}}Colors: {{.site_context.color_palette | tojson}}{{end}}\n{{if .site_context.typography}}Typography: {{.site_context.typography | tojson}}{{end}}\n\n## Components\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\nReturn ONLY valid JSON:\n{\n  \"color_scheme\": {\n    \"primary\": \"#hex\",\n    \"secondary\": \"#hex\",\n    \"accent\": \"#hex\",\n    \"background\": \"#fff\",\n    \"surface\": \"#f8f9fa\",\n    \"text\": \"#333\",\n    \"text_muted\": \"#666\",\n    \"border\": \"#e2e8f0\"\n  },\n  \"typography\": {\n    \"font_family\": \"-apple-system, sans-serif\",\n    \"heading_font\": \"inherit\",\n    \"base_size\": \"16px\",\n    \"line_height\": \"1.6\"\n  },\n  \"spacing\": {\n    \"section_padding\": \"4rem\",\n    \"container_max_width\": \"1200px\"\n  },\n  \"components_to_style\": [\"hero\", \"services-grid\"],\n  \"design_notes\": \"notes\"\n}"
                             },
                             "description": "Generate design spec",
                             "next_step": "generate_css",
                             "output_field": "design_spec"
                         },
                         "generate_css": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 8000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_context", "design_spec"],
                                 "output_format": "text",
                                 "prompt_template": "Generate a complete production CSS stylesheet.\n\n## Design Spec\n{{.design_spec.result}}\n\n## Components\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\n## Requirements\n1. CSS variables in :root based on color_scheme\n2. Minimal reset (box-sizing, margin/padding)\n3. Typography (body, h1-h6, p, a)\n4. Layout (.container max-width centered, .section padding)\n5. Buttons (.btn, .btn-primary, .btn-secondary, .btn-large)\n6. Hero (.hero min-70vh, centered, white text with shadow)\n7. Grids (.services-grid responsive 1-2-3 cols, .team-grid, .stats-grid)\n8. Cards (.service-item, .team-member with shadow and hover)\n9. CTA (.cta-section with background)\n10. Mobile-first responsive (768px, 1024px)\n11. Focus states, smooth transitions\n\nOutput ONLY CSS. No markdown. Start with :root {"
                             },
                             "description": "Generate CSS",
                             "next_step": "deploy_css",
                             "output_field": "generated_css"
                         },
                         "deploy_css": {
                             "action": "git_commit",
                             "config": {
                                 "domain_field": "site_context.domain",
                                 "content_field": "generated_css.result",
                                 "file_path": "assets/css/styles.css",
                                 "commit_message": "Update stylesheet via webdesign-agent"
                             },
                             "description": "Deploy CSS to git",
                             "next_step": "check_update_db",
                             "output_field": "css_deployed"
                         },
                         "check_update_db": {
                             "action": "conditional",
                             "config": {
                                 "condition": "site_context.site_id != null AND site_context.site_id != ''",
                                 "then_step": "update_site",
                                 "else_step": "complete"
                             },
                             "description": "Check if we should update DB"
                         },
                         "update_site": {
                             "action": "update_site_content",
                             "config": {
                                 "site_id_field": "site_context.site_id",
                                 "merge": true,
                                 "content_field": "design_spec.result"
                             },
                             "description": "Store design spec",
                             "next_step": "complete",
                             "output_field": "site_updated"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["design_spec", "css_deployed", "site_context"]
                             }
                         }
                     }
                 }
             }',
             true,
             '["design", "css", "styling", "specialist"]',
             '{"required": [], "optional": ["site_id", "domain", "site_context"]}',
             '{"produces": {"css_deployed": "git result", "design_spec": "design spec"}}'
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();


---

-- Webdesign Agent Definition (Final)
--
-- Generates CSS stylesheets for sites.
-- Uses file_path config in git_commit (requires patch_01_git_commit_file_path.go)
-- No container config (handled by spawn_actions.go)

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    input_contract,
    output_contract
) VALUES (
             'webdesign-agent',
             'Web Design Agent',
             'Generates CSS stylesheets for sites. Accepts site_context or loads from DB. Analyzes design requirements and generates production CSS.',
             'specialist',
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "check_site_context",
                     "steps": {
                         "check_site_context": {
                             "action": "conditional",
                             "config": {
                                 "condition": "input_data.site_context.domain != null AND input_data.site_context.domain != ''",
                                 "then_step": "use_provided_context",
                                 "else_step": "load_site_context"
                             },
                             "description": "Check if site_context was provided"
                         },
                         "use_provided_context": {
                             "action": "transform_data",
                             "config": {
                                 "source_field": "input_data.site_context",
                                 "output_key": "site_context"
                             },
                             "description": "Use provided site_context",
                             "next_step": "analyze_design",
                             "output_field": "site_context"
                         },
                         "load_site_context": {
                             "action": "load_site_for_design",
                             "config": {
                                 "site_id_field": "input_data.site_id",
                                 "domain_field": "input_data.domain",
                                 "include_pages": true,
                                 "include_style_collection": true
                             },
                             "description": "Load site from database",
                             "next_step": "analyze_design",
                             "output_field": "site_context"
                         },
                         "analyze_design": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 2000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_context"],
                                 "output_format": "json",
                                 "prompt_template": "You are a web design expert. Analyze the site and output a design specification.\n\n## Site\nDomain: {{.site_context.domain}}\nCompany: {{.site_context.company_name}}\nIndustry: {{if .site_context.industry}}{{.site_context.industry}}{{else}}professional services{{end}}\nTagline: {{.site_context.tagline}}\n\n## Components Used\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\nReturn ONLY valid JSON:\n{\n  \"color_scheme\": {\n    \"primary\": \"#1a1a2e\",\n    \"secondary\": \"#16213e\",\n    \"accent\": \"#0f3460\",\n    \"background\": \"#ffffff\",\n    \"surface\": \"#f8f9fa\",\n    \"text\": \"#333333\",\n    \"text_muted\": \"#666666\",\n    \"border\": \"#e2e8f0\"\n  },\n  \"typography\": {\n    \"font_family\": \"-apple-system, BlinkMacSystemFont, sans-serif\",\n    \"heading_font\": \"inherit\",\n    \"base_size\": \"16px\",\n    \"line_height\": \"1.6\"\n  },\n  \"spacing\": {\n    \"section_padding\": \"5rem 2rem\",\n    \"container_max_width\": \"1200px\"\n  },\n  \"design_notes\": \"brief notes about design choices\"\n}"
                             },
                             "description": "Generate design spec",
                             "next_step": "generate_css",
                             "output_field": "design_spec"
                         },
                         "generate_css": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 8000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_context", "design_spec"],
                                 "output_format": "text",
                                 "prompt_template": "Generate a complete production CSS stylesheet.\n\n## Design Spec\n{{.design_spec.result}}\n\n## Components\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\n## Requirements\n1. CSS variables in :root based on color_scheme\n2. Minimal reset (box-sizing, margin/padding)\n3. Typography (body, h1-h6, p, a)\n4. Layout (.container max-width centered, .section padding)\n5. Buttons (.btn, .btn-primary, .btn-secondary, .btn-large)\n6. Hero (.hero min-70vh, centered, white text with shadow)\n7. Grids (.services-grid responsive 1-2-3 cols, .team-grid, .stats-grid)\n8. Cards (.service-item, .team-member with shadow and hover)\n9. CTA (.cta-section with background)\n10. Mobile-first responsive (768px, 1024px)\n11. Focus states, smooth transitions\n\nOutput ONLY CSS. No markdown. Start with :root {"
                             },
                             "description": "Generate CSS",
                             "next_step": "deploy_css",
                             "output_field": "generated_css"
                         },
                         "deploy_css": {
                             "action": "git_commit",
                             "config": {
                                 "domain_field": "site_context.domain",
                                 "content_field": "generated_css.result",
                                 "file_path": "assets/css/styles.css",
                                 "commit_message": "Update stylesheet via webdesign-agent"
                             },
                             "description": "Deploy CSS to git",
                             "next_step": "check_update_db",
                             "output_field": "css_deployed"
                         },
                         "check_update_db": {
                             "action": "conditional",
                             "config": {
                                 "condition": "site_context.site_id != null AND site_context.site_id != ''",
                                 "then_step": "update_site",
                                 "else_step": "complete"
                             },
                             "description": "Check if we should update DB"
                         },
                         "update_site": {
                             "action": "update_site_content",
                             "config": {
                                 "site_id_field": "site_context.site_id",
                                 "merge": true,
                                 "content_field": "design_spec.result"
                             },
                             "description": "Store design spec",
                             "next_step": "complete",
                             "output_field": "site_updated"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["design_spec", "css_deployed", "site_context"]
                             }
                         }
                     }
                 }
             }',
             true,
             '["design", "css", "styling", "specialist"]',
             '{"required": [], "optional": ["site_id", "domain", "site_context"]}',
             '{"produces": {"css_deployed": "git result", "design_spec": "design spec"}}'
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();


---


-- Fix webdesign-agent conditional check
-- The '' (empty string) becomes ' (single quote) due to PostgreSQL escaping
-- Solution: just check != null since UUIDs are never empty strings

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_update_db,config,condition}',
        '"site_context.site_id != null"'
                     )
WHERE type = 'webdesign-agent';

-- Verify
SELECT type,
       default_config->'workflow'->'steps'->'check_update_db'->'config'->>'condition' as condition
FROM agent_definitions
WHERE type = 'webdesign-agent';


---

-- Webdesign Agent Definition (Final)
--
-- Generates CSS stylesheets for sites.
-- Uses file_path config in git_commit (requires patch_01_git_commit_file_path.go)
-- No container config (handled by spawn_actions.go)

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    input_contract,
    output_contract
) VALUES (
             'webdesign-agent',
             'Web Design Agent',
             'Generates CSS stylesheets for sites. Accepts site_context or loads from DB. Analyzes design requirements and generates production CSS.',
             'specialist',
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "check_site_context",
                     "steps": {
                         "check_site_context": {
                             "action": "conditional",
                             "config": {
                                 "condition": "input_data.site_context.domain != null AND input_data.site_context.domain != ''",
                                 "then_step": "use_provided_context",
                                 "else_step": "load_site_context"
                             },
                             "description": "Check if site_context was provided"
                         },
                         "use_provided_context": {
                             "action": "transform_data",
                             "config": {
                                 "source_field": "input_data.site_context",
                                 "output_key": "site_context"
                             },
                             "description": "Use provided site_context",
                             "next_step": "analyze_design",
                             "output_field": "site_context"
                         },
                         "load_site_context": {
                             "action": "load_site_for_design",
                             "config": {
                                 "site_id_field": "input_data.site_id",
                                 "domain_field": "input_data.domain",
                                 "include_pages": true,
                                 "include_style_collection": true
                             },
                             "description": "Load site from database",
                             "next_step": "analyze_design",
                             "output_field": "site_context"
                         },
                         "analyze_design": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 2000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_context"],
                                 "output_format": "json",
                                 "prompt_template": "You are a web design expert. Analyze the site and output a design specification.\n\n## Site\nDomain: {{.site_context.domain}}\nCompany: {{.site_context.company_name}}\nIndustry: {{if .site_context.industry}}{{.site_context.industry}}{{else}}professional services{{end}}\nTagline: {{.site_context.tagline}}\n\n## Components Used\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\nReturn ONLY valid JSON:\n{\n  \"color_scheme\": {\n    \"primary\": \"#1a1a2e\",\n    \"secondary\": \"#16213e\",\n    \"accent\": \"#0f3460\",\n    \"background\": \"#ffffff\",\n    \"surface\": \"#f8f9fa\",\n    \"text\": \"#333333\",\n    \"text_muted\": \"#666666\",\n    \"border\": \"#e2e8f0\"\n  },\n  \"typography\": {\n    \"font_family\": \"-apple-system, BlinkMacSystemFont, sans-serif\",\n    \"heading_font\": \"inherit\",\n    \"base_size\": \"16px\",\n    \"line_height\": \"1.6\"\n  },\n  \"spacing\": {\n    \"section_padding\": \"5rem 2rem\",\n    \"container_max_width\": \"1200px\"\n  },\n  \"design_notes\": \"brief notes about design choices\"\n}"
                             },
                             "description": "Generate design spec",
                             "next_step": "generate_css",
                             "output_field": "design_spec"
                         },
                         "generate_css": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 8000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_context", "design_spec"],
                                 "output_format": "text",
                                 "prompt_template": "Generate a complete production CSS stylesheet.\n\n## Design Spec\n{{.design_spec.result}}\n\n## Components\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\n## Requirements\n1. MUST start with :root containing these EXACT variable names (use design_spec colors):\n   --color-primary: (from color_scheme.primary)\n   --color-secondary: (from color_scheme.secondary)\n   --color-accent: (from color_scheme.accent)\n   --color-background: (from color_scheme.background)\n   --color-surface: (from color_scheme.surface)\n   --color-text: (from color_scheme.text)\n   --color-text-muted: (from color_scheme.text_muted)\n   --color-border: (from color_scheme.border)\n   --color-white: #ffffff\n   --font-family: (from typography.font_family)\n   --spacing-section: (from spacing.section_padding)\n   --container-max-width: (from spacing.container_max_width)\n\n2. Minimal reset (box-sizing, margin/padding for html, body)\n3. Base typography using var(--font-family), var(--color-text)\n4. Headings h1-h6 using var(--color-primary)\n5. Links using var(--color-accent)\n6. Layout helpers (.container using var(--container-max-width))\n7. Button styles using CSS variables\n8. Hero section (min-height: 70vh, centered content)\n9. Responsive breakpoints (768px, 1024px)\n10. Smooth transitions, focus states\n\nIMPORTANT: Components have their own CSS that references these variables. Do NOT duplicate component-specific styles - just provide the :root variables and base styles.\n\nOutput ONLY CSS. No markdown. No explanations. Start with :root {"
                             },
                             "description": "Generate CSS",
                             "next_step": "deploy_css",
                             "output_field": "generated_css"
                         },
                         "deploy_css": {
                             "action": "git_commit",
                             "config": {
                                 "domain_field": "site_context.domain",
                                 "content_field": "generated_css.result",
                                 "file_path": "assets/css/styles.css",
                                 "commit_message": "Update stylesheet via webdesign-agent"
                             },
                             "description": "Deploy CSS to git",
                             "next_step": "check_update_db",
                             "output_field": "css_deployed"
                         },
                         "check_update_db": {
                             "action": "conditional",
                             "config": {
                                 "condition": "site_context.site_id != null",
                                 "then_step": "update_site",
                                 "else_step": "complete"
                             },
                             "description": "Check if we should update DB"
                         },
                         "update_site": {
                             "action": "update_site_content",
                             "config": {
                                 "site_id_field": "site_context.site_id",
                                 "merge": true,
                                 "content_field": "design_spec.result"
                             },
                             "description": "Store design spec",
                             "next_step": "complete",
                             "output_field": "site_updated"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["design_spec", "css_deployed", "site_context"]
                             }
                         }
                     }
                 }
             }',
             true,
             '["design", "css", "styling", "specialist"]',
             '{"required": [], "optional": ["site_id", "domain", "site_context"]}',
             '{"produces": {"css_deployed": "git result", "design_spec": "design spec"}}'
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();


-- global vs local css
-- Webdesign Agent Definition (Final)
--
-- Generates CSS stylesheets for sites.
-- Uses file_path config in git_commit (requires patch_01_git_commit_file_path.go)
-- No container config (handled by spawn_actions.go)

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    input_contract,
    output_contract
) VALUES (
             'webdesign-agent',
             'Web Design Agent',
             'Generates CSS stylesheets for sites. Accepts site_context or loads from DB. Analyzes design requirements and generates production CSS.',
             'specialist',
             '{
                 "processing_mode": "task",
                 "timeout_seconds": 300,
                 "workflow": {
                     "start_step": "check_site_context",
                     "steps": {
                         "check_site_context": {
                             "action": "conditional",
                             "config": {
                                 "condition": "input_data.site_context.domain != null AND input_data.site_context.domain != ''",
                                 "then_step": "use_provided_context",
                                 "else_step": "load_site_context"
                             },
                             "description": "Check if site_context was provided"
                         },
                         "use_provided_context": {
                             "action": "transform_data",
                             "config": {
                                 "source_field": "input_data.site_context",
                                 "output_key": "site_context"
                             },
                             "description": "Use provided site_context",
                             "next_step": "analyze_design",
                             "output_field": "site_context"
                         },
                         "load_site_context": {
                             "action": "load_site_for_design",
                             "config": {
                                 "site_id_field": "input_data.site_id",
                                 "domain_field": "input_data.domain",
                                 "include_pages": true,
                                 "include_style_collection": true
                             },
                             "description": "Load site from database",
                             "next_step": "analyze_design",
                             "output_field": "site_context"
                         },
                         "analyze_design": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 2000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_context"],
                                 "output_format": "json",
                                 "prompt_template": "You are a web design expert. Analyze the site and output a design specification.\n\n## Site\nDomain: {{.site_context.domain}}\nCompany: {{.site_context.company_name}}\nIndustry: {{if .site_context.industry}}{{.site_context.industry}}{{else}}professional services{{end}}\nTagline: {{.site_context.tagline}}\n\n## Components Used\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\nReturn ONLY valid JSON:\n{\n  \"color_scheme\": {\n    \"primary\": \"#1a1a2e\",\n    \"secondary\": \"#16213e\",\n    \"accent\": \"#0f3460\",\n    \"background\": \"#ffffff\",\n    \"surface\": \"#f8f9fa\",\n    \"text\": \"#333333\",\n    \"text_muted\": \"#666666\",\n    \"border\": \"#e2e8f0\"\n  },\n  \"typography\": {\n    \"font_family\": \"-apple-system, BlinkMacSystemFont, sans-serif\",\n    \"heading_font\": \"inherit\",\n    \"base_size\": \"16px\",\n    \"line_height\": \"1.6\"\n  },\n  \"spacing\": {\n    \"section_padding\": \"5rem 2rem\",\n    \"container_max_width\": \"1200px\"\n  },\n  \"design_notes\": \"brief notes about design choices\"\n}"
                             },
                             "description": "Generate design spec",
                             "next_step": "generate_css",
                             "output_field": "design_spec"
                         },
                         "generate_css": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "provider": "anthropic",
                                     "model": "claude-sonnet-4-5",
                                     "max_tokens": 8000,
                                     "api_key_env_var": "ANTHROPIC_API_KEY"
                                 },
                                 "input_fields": ["site_context", "design_spec"],
                                 "output_format": "text",
                                 "prompt_template": "Generate a complete production CSS stylesheet.\n\n## Design Spec\n{{.design_spec.result}}\n\n## Components\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\n## CSS RESPONSIBILITY RULES\nGlobal CSS handles ALL appearance. Component CSS handles only layout.\n\nYou MUST provide:\n1. :root with EXACT variable names (use design_spec colors)\n2. Base element styling that components inherit\n\nComponents will NOT re-declare colors on h1-h6, p, a - they inherit from you.\n\n## Required :root Variables\n:root {\n  --color-primary: (from color_scheme.primary);\n  --color-secondary: (from color_scheme.secondary);\n  --color-accent: (from color_scheme.accent);\n  --color-background: (from color_scheme.background);\n  --color-surface: (from color_scheme.surface);\n  --color-text: (from color_scheme.text);\n  --color-text-muted: (from color_scheme.text_muted);\n  --color-border: (from color_scheme.border);\n  --color-white: #ffffff;\n  --font-family: (from typography.font_family);\n  --spacing-section: (from spacing.section_padding);\n  --container-max-width: (from spacing.container_max_width);\n}\n\n## Required Base Styles\n- *, *::before, *::after { box-sizing: border-box; }\n- html, body { margin: 0; padding: 0; }\n- body { font-family: var(--font-family); color: var(--color-text); line-height: 1.6; }\n- h1, h2, h3, h4, h5, h6 { color: var(--color-primary); line-height: 1.2; margin: 0 0 1rem; }\n- h1 { font-size: clamp(2rem, 5vw, 3rem); }\n- h2 { font-size: clamp(1.75rem, 4vw, 2.5rem); }\n- h3 { font-size: clamp(1.25rem, 3vw, 1.5rem); }\n- p { margin: 0 0 1rem; color: var(--color-text); }\n- a { color: var(--color-accent); }\n- .container { max-width: var(--container-max-width); margin: 0 auto; padding: 0 2rem; }\n\n## Also Include\n- Button base styles (.btn, .btn-primary, .btn-secondary)\n- Focus states for accessibility\n- Smooth transitions\n- Responsive adjustments at 768px and 1024px\n\n## DO NOT Include\n- Component-specific selectors (.services-grid, .testimonial-item, etc.)\n- Components have their own CSS that inherits from your base styles\n\nOutput ONLY CSS. No markdown. No explanations. Start with :root {"
                             },
                             "description": "Generate CSS",
                             "next_step": "deploy_css",
                             "output_field": "generated_css"
                         },
                         "deploy_css": {
                             "action": "git_commit",
                             "config": {
                                 "domain_field": "site_context.domain",
                                 "content_field": "generated_css.result",
                                 "file_path": "assets/css/styles.css",
                                 "commit_message": "Update stylesheet via webdesign-agent"
                             },
                             "description": "Deploy CSS to git",
                             "next_step": "check_update_db",
                             "output_field": "css_deployed"
                         },
                         "check_update_db": {
                             "action": "conditional",
                             "config": {
                                 "condition": "site_context.site_id != null",
                                 "then_step": "update_site",
                                 "else_step": "complete"
                             },
                             "description": "Check if we should update DB"
                         },
                         "update_site": {
                             "action": "update_site_content",
                             "config": {
                                 "site_id_field": "site_context.site_id",
                                 "merge": true,
                                 "content_field": "design_spec.result"
                             },
                             "description": "Store design spec",
                             "next_step": "complete",
                             "output_field": "site_updated"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "config": {
                                 "output_fields": ["design_spec", "css_deployed", "site_context"]
                             }
                         }
                     }
                 }
             }',
             true,
             '["design", "css", "styling", "specialist"]',
             '{"required": [], "optional": ["site_id", "domain", "site_context"]}',
             '{"produces": {"css_deployed": "git result", "design_spec": "design spec"}}'
         )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();

--

-- Fix webdesign-agent CSS generation prompt.
-- Problem: the prompt instructs the LLM to set explicit color on p, h1-h6,
-- blockquote, li, strong etc. This prevents colour inheritance in dark sections
-- where components set color: #fff on a parent container.
--
-- Fix: instruct LLM to use color: inherit on elements, let body set the
-- default, and add explicit guidance about dark section inheritance.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_css,config,prompt_template}',
        to_jsonb(
                'Generate a complete production CSS stylesheet.

                ## Design Spec
                {{.design_spec.result}}

                ## Components
                {{range .site_context.all_component_functions}}- {{.}}
                {{end}}

                ## CSS RESPONSIBILITY RULES
                Global CSS handles base appearance and resets. Component inline CSS handles layout AND section-specific colours (dark sections, CTAs etc.).

                You MUST provide:
                1. :root with EXACT variable names (use design_spec colors)
                2. Base element styling using INHERITANCE (not forced colors)
                3. Button styles, focus states, responsive breakpoints

                ## Required :root Variables
                :root {
                  --color-primary: (from color_scheme.primary);
                  --color-secondary: (from color_scheme.secondary);
                  --color-accent: (from color_scheme.accent);
                  --color-background: (from color_scheme.background);
                  --color-surface: (from color_scheme.surface);
                  --color-text: (from color_scheme.text);
                  --color-text-muted: (from color_scheme.text_muted);
                  --color-border: (from color_scheme.border);
                  --color-white: #ffffff;
                  --font-family: (from typography.font_family);
                  --spacing-section: (from spacing.section_padding);
                  --container-max-width: (from spacing.container_max_width);
                  --transition-speed: 0.3s;
                  --border-radius: 0.5rem;
                  --box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
                  --box-shadow-hover: 0 4px 16px rgba(0, 0, 0, 0.15);
                }

                ## COLOUR INHERITANCE RULES (IMPORTANT - READ CAREFULLY)
                body sets color: var(--color-text) which ALL elements inherit by default.
                Components have dark sections (testimonials, CTAs, footers) that set color: #fff on a parent container, and all children MUST inherit that light colour.

                Therefore you MUST follow these rules:
                - body: set color: var(--color-text) - this is the ONLY place default text color is set
                - h1, h2, h3, h4, h5, h6: use color: inherit (NOT var(--color-primary))
                - p: do NOT set color at all (it inherits from parent)
                - li: do NOT set color at all
                - blockquote: do NOT set background-color or color (components handle this)
                - strong, b: do NOT set color
                - em, i: do NOT set color
                - span: do NOT set color
                - cite: do NOT set color
                - a: color: var(--color-accent) is OK (links are an exception)

                If you force color: var(--color-text) on p, blockquote, li, or h1-h6, dark sections will have dark text on dark backgrounds and be unreadable. This is the single most important rule.

                ## Required Base Styles
                - *, *::before, *::after { box-sizing: border-box; }
                - html, body { margin: 0; padding: 0; }
                - body { font-family: var(--font-family); color: var(--color-text); line-height: 1.6; background-color: var(--color-background); -webkit-font-smoothing: antialiased; }
                - h1, h2, h3, h4, h5, h6 { color: inherit; line-height: 1.2; margin: 0 0 1rem; font-weight: 700; }
                - h1 { font-size: clamp(2rem, 5vw, 3rem); }
                - h2 { font-size: clamp(1.75rem, 4vw, 2.5rem); }
                - h3 { font-size: clamp(1.25rem, 3vw, 1.5rem); }
                - h4 { font-size: clamp(1.1rem, 2.5vw, 1.25rem); }
                - p { margin: 0 0 1rem; }
                - a { color: var(--color-accent); text-decoration: none; transition: color var(--transition-speed) ease; }
                - a:hover { color: var(--color-primary); }
                - a:focus { outline: 2px solid var(--color-accent); outline-offset: 2px; }
                - img { max-width: 100%; height: auto; display: block; }
                - ul, ol { margin: 0 0 1rem; padding-left: 1.5rem; }
                - blockquote { margin: 0 0 1rem; padding: 1rem 1.5rem; border-left: 4px solid var(--color-accent); font-style: italic; }
                - hr { border: 0; border-top: 1px solid var(--color-border); margin: 2rem 0; }
                - .container { max-width: var(--container-max-width); margin: 0 auto; padding: 0 2rem; }

                ## Also Include
                - Button base styles (.btn, .btn-primary, .btn-secondary) with hover/active/focus states
                - Focus-visible states for accessibility
                - Smooth transitions using var(--transition-speed)
                - Responsive adjustments at 480px, 768px, and 1024px
                - prefers-reduced-motion media query

                ## DO NOT Include
                - Component-specific selectors (.services-grid, .testimonial-item, .case-study-item, .differentiator-item, .social-proof-section, .cta-section, etc.)
                - Components have their own inline CSS that handles their layout and dark section colours
                - Do NOT set background-color on blockquote (components handle this contextually)
                - Do NOT set color on p, li, h1-h6, blockquote, strong, cite, span (they inherit)

                Output ONLY CSS. No markdown. No explanations. Start with :root {'
        )
                     )
WHERE type = 'webdesign-agent' AND is_active = true;

--

-- styles fixes

-- Fix webdesign-agent generate_css prompt to enforce colour inheritance model
--
-- Changes from current prompt:
--   1. h1-h6: color: var(--color-primary) → color: inherit
--   2. p: removed "color: var(--color-text)"
--   3. Added explicit DO NOT rules for blockquote background, strong color, etc.
--   4. Added COLOUR INHERITANCE MODEL section explaining why
--
-- The current prompt contradicts the design system architecture by telling the LLM
-- to set explicit colors on elements that should inherit. This causes white/light
-- text on white/light backgrounds in dark sections (testimonials, CTAs, footers)
-- because component CSS sets color: #fff on the container but global CSS forces
-- colors back on children.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_css,config,prompt_template}',
        to_jsonb(
                E'Generate a complete production CSS stylesheet.\n\n## Design Spec\n{{.design_spec.result}}\n\n## Components\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\n## CSS RESPONSIBILITY RULES\nGlobal CSS handles ALL appearance. Component CSS handles only layout.\n\nYou MUST provide:\n1. :root with EXACT variable names (use design_spec colors)\n2. Base element styling that components inherit\n\nComponents will NOT re-declare colors on h1-h6, p, a - they inherit from you.\n\n## Required :root Variables\n:root {\n  --color-primary: (from color_scheme.primary);\n  --color-secondary: (from color_scheme.secondary);\n  --color-accent: (from color_scheme.accent);\n  --color-background: (from color_scheme.background);\n  --color-surface: (from color_scheme.surface);\n  --color-text: (from color_scheme.text);\n  --color-text-muted: (from color_scheme.text_muted);\n  --color-border: (from color_scheme.border);\n  --color-white: #ffffff;\n  --font-family: (from typography.font_family);\n  --spacing-section: (from spacing.section_padding);\n  --container-max-width: (from spacing.container_max_width);\n}\n\n## COLOUR INHERITANCE MODEL (this is essential)\nbody sets color ONCE via var(--color-text). ALL other elements inherit.\nDark sections (testimonials, CTAs, footers) override with color: #fff on their container.\nChildren inherit light text automatically — but ONLY if you do not force colors on them.\n\nIf you set color: var(--color-primary) on h2, a dark section cannot override it.\nIf you set color: var(--color-text) on p, a dark section cannot override it.\nIf you set background-color on blockquote, dark sections get light boxes inside dark containers.\n\n## Required Base Styles\n- *, *::before, *::after { box-sizing: border-box; }\n- html, body { margin: 0; padding: 0; }\n- body { font-family: var(--font-family); color: var(--color-text); line-height: 1.6; background-color: var(--color-background); }\n- h1, h2, h3, h4, h5, h6 { color: inherit; line-height: 1.2; margin: 0 0 1rem; font-weight: 700; }\n- h1 { font-size: clamp(2rem, 5vw, 3rem); }\n- h2 { font-size: clamp(1.75rem, 4vw, 2.5rem); }\n- h3 { font-size: clamp(1.25rem, 3vw, 1.5rem); }\n- p { margin: 0 0 1rem; }\n- a { color: var(--color-accent); }\n- .container { max-width: var(--container-max-width); margin: 0 auto; padding: 0 2rem; }\n\n## DO NOT set color on these elements (they must inherit):\n- p — NO color property at all\n- h1, h2, h3, h4, h5, h6 — use color: inherit, NOT var(--color-primary)\n- strong, b, em, i — NO color property\n- li — NO color property\n- blockquote — NO color property, NO background-color property\n- cite, span — NO color property\n\n## Also Include\n- Button base styles (.btn, .btn-primary, .btn-secondary)\n- Focus states for accessibility\n- Smooth transitions\n- Responsive adjustments at 768px and 1024px\n- blockquote: margin, padding, border-left only (NO background-color, NO color)\n- strong: font-weight only (NO color)\n\n## DO NOT Include\n- Component-specific selectors (.services-grid, .testimonial-item, etc.)\n- Components have their own CSS that inherits from your base styles\n\nOutput ONLY CSS. No markdown. No explanations. Start with :root {'
        )
             )
WHERE type = 'webdesign-agent';

--

-- don't put light text on light backgrounds

UPDATE agent_definitions
SET default_config = jsonb_set(
  default_config,
  '{workflow,steps,generate_css,config,prompt_template}',
  to_jsonb(
    'Generate a complete production CSS stylesheet.\n\n## Design Spec\n{{.design_spec.result}}\n\n## Components\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\n## CSS RESPONSIBILITY RULES\nGlobal CSS handles ALL appearance. Component CSS handles only layout.\n\nYou MUST provide:\n1. :root with EXACT variable names (use design_spec colors)\n2. Base element styling that components inherit\n\nComponents will NOT re-declare colors on h1-h6, p, a - they inherit from you.\n\n## Required :root Variables\n:root {\n  --color-primary: (from color_scheme.primary);\n  --color-secondary: (from color_scheme.secondary);\n  --color-accent: (from color_scheme.accent);\n  --color-background: (from color_scheme.background);\n  --color-surface: (from color_scheme.surface);\n  --color-text: (from color_scheme.text);\n  --color-text-muted: (from color_scheme.text_muted);\n  --color-border: (from color_scheme.border);\n  --color-white: #ffffff;\n  --font-family: (from typography.font_family);\n  --spacing-section: (from spacing.section_padding);\n  --container-max-width: (from spacing.container_max_width);\n}\n\n## Required Base Styles\n- *, *::before, *::after { box-sizing: border-box; }\n- html, body { margin: 0; padding: 0; }\n- body { font-family: var(--font-family); color: var(--color-text); line-height: 1.6; }\n- h1, h2, h3, h4, h5, h6 { color: var(--color-primary); line-height: 1.2; margin: 0 0 1rem; }\n- h1 { font-size: clamp(2rem, 5vw, 3rem); }\n- h2 { font-size: clamp(1.75rem, 4vw, 2.5rem); }\n- h3 { font-size: clamp(1.25rem, 3vw, 1.5rem); }\n- p { margin: 0 0 1rem; } /* NO color — inherits from body or dark-section parent */\n- a { color: var(--color-accent); }\n- .container { max-width: var(--container-max-width); margin: 0 auto; padding: 0 2rem; }\n\n## INHERITANCE RULES — READ CAREFULLY\nText elements (p, span, li, blockquote, strong, b, em, cite) must NOT have explicit color set.\nBody sets color: var(--color-text) and children inherit.\nDark-section components set color: #fff on their container — children must inherit, not fight.\nIf you set p { color: var(--color-text); } it BREAKS every dark section.\nblockquote must NOT have background-color set (components handle their own backgrounds).\nstrong, b must NOT have color set (they inherit from parent).\n\n## Also Include\n- Button base styles (.btn, .btn-primary, .btn-secondary)\n- Focus states for accessibility\n- Smooth transitions\n- Responsive adjustments at 768px and 1024px\n\n## DO NOT Include\n- Component-specific selectors (.services-grid, .testimonial-item, etc.)\n- Components have their own CSS that inherits from your base styles\n- Explicit color on p, blockquote, strong, b, span, li, cite elements\n- background-color on blockquote\n\nOutput ONLY CSS. No markdown. No explanations. Start with :root {'::text
  )
),
updated_at = NOW()
WHERE type = 'webdesign-agent';

--

-- ============================================================================
-- UPDATE WEBDESIGN-AGENT PROMPT — SECTION CONTEXT VARIABLE CONTRACT
-- ============================================================================
-- The generate_css step prompt must teach the LLM the --section-* pattern
-- so that regenerated stylesheets always follow the contract.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_css,config,prompt_template}',
        to_jsonb(
                'Generate a complete production CSS stylesheet.

            ## Design Spec
            {{.design_spec.result}}

            ## Components
            {{range .site_context.all_component_functions}}- {{.}}
            {{end}}

            ## ARCHITECTURE: SECTION CONTEXT VARIABLES
            This site uses a section-context variable pattern for dark/light section support.
            Body sets light-theme defaults. Text elements use --section-* variables with fallbacks.
            Dark-section components override --section-* on their container — children adapt automatically.

            ## Required :root Variables
            :root {
              --color-primary: (from color_scheme.primary);
              --color-secondary: (from color_scheme.secondary);
              --color-accent: (from color_scheme.accent);
              --color-background: (from color_scheme.background);
              --color-surface: (from color_scheme.surface);
              --color-text: (from color_scheme.text);
              --color-text-muted: (from color_scheme.text_muted);
              --color-border: (from color_scheme.border);
              --color-white: #ffffff;
              --font-family: (from typography.font_family);
              --spacing-section: (from spacing.section_padding);
              --container-max-width: (from spacing.container_max_width);
              --transition-speed: 0.3s;
              --border-radius: 8px;
              --shadow-sm: 0 2px 4px rgba(0,0,0,0.1);
              --shadow-md: 0 4px 6px rgba(0,0,0,0.1);
              --shadow-lg: 0 10px 20px rgba(0,0,0,0.15);
            }

            ## Required Base Styles
            - *, *::before, *::after { box-sizing: border-box; }
            - html { scroll-behavior: smooth; }
            - html, body { margin: 0; padding: 0; }
            - body { font-family: var(--font-family); color: var(--color-text); line-height: 1.6; background-color: var(--color-background); }

            ## SECTION CONTEXT PATTERN — FOLLOW EXACTLY
            Headings use --section-heading with fallback:
              h1-h6 { color: var(--section-heading, var(--color-primary)); }

            Text elements use --section-* with inherit fallback:
              p { margin: 0 0 1rem; color: var(--section-text, inherit); }
              strong, b { font-weight: 700; color: var(--section-heading, inherit); }

            Styled elements use --section-* with theme fallback:
              blockquote { background-color: var(--section-surface, var(--color-surface)); color: var(--section-text-muted, var(--color-text-muted)); border-left: 4px solid var(--section-border, var(--color-border)); }
              cite { color: var(--section-text-muted, var(--color-text-muted)); }
              .text-muted { color: var(--section-text-muted, var(--color-text-muted)); }

            DO NOT set explicit color on p, strong, b, blockquote, cite, span, li.
            Always use var(--section-*, <fallback>) pattern instead.
            This ensures dark-section components work by overriding --section-* on their container.

            ## Typography
            - h1 { font-size: clamp(2rem, 5vw, 3rem); }
            - h2 { font-size: clamp(1.75rem, 4vw, 2.5rem); }
            - h3 { font-size: clamp(1.25rem, 3vw, 1.5rem); }
            - a { color: var(--color-accent); }
            - .container { max-width: var(--container-max-width); margin: 0 auto; padding: 0 2rem; }

            ## Also Include
            - Button base styles (.btn, .btn-primary, .btn-secondary, .btn-large, .btn-small)
            - Focus-visible states for accessibility
            - Smooth transitions
            - Responsive adjustments at 768px and 1024px
            - prefers-reduced-motion media query
            - Form input styling

            ## DO NOT Include
            - Component-specific selectors (.services-grid, .testimonial-item, etc.)
            - Components have their own inline CSS that inherits from these base styles

            ## Include at end as CSS comment: Dark Section Template
            /* DARK SECTION TEMPLATE — components with dark backgrounds set these:
               .my-dark-section {
                 --section-text: rgba(255,255,255,0.9);
                 --section-text-muted: rgba(255,255,255,0.7);
                 --section-heading: #ffffff;
                 --section-surface: rgba(255,255,255,0.05);
                 --section-border: rgba(255,255,255,0.2);
               }
            */

            Output ONLY CSS. No markdown fences. No explanations. Start with :root {'::text
        )
                     ),
    updated_at = NOW()
WHERE type = 'webdesign-agent';

--

-- 060_patch_webdesign_agent_prompt.sql
--
-- Updates the webdesign-agent's analyze_design prompt to use the new fields
-- that LoadSiteForDesignAction now populates from site_specs.
--
-- New fields available in site_context after the Go patch:
--   .site_context.services    (from identity.services)
--   .site_context.brand_tone  (from briefing.tone)
--   .site_context.about_us    (from briefing.about_us)
--   .site_context.site_type   (from classification.site_type)
--
-- The existing fields remain unchanged:
--   .site_context.domain, .company_name, .industry, .tagline,
--   .all_component_functions, .color_palette, .typography

-- Step 1: Update the analyze_design prompt to include richer context
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_design,config,prompt_template}',
        to_jsonb(
                'You are a web design expert. Analyze the site and output a design specification.' ||
                E'\n\n## Site' ||
        E'\nDomain: {{.site_context.domain}}' ||
        E'\nCompany: {{.site_context.company_name}}' ||
        E'\nIndustry: {{if .site_context.industry}}{{.site_context.industry}}{{else}}professional services{{end}}' ||
        E'\nTagline: {{.site_context.tagline}}' ||
        E'\nSite Type: {{.site_context.site_type}}' ||
        E'\nBrand Tone: {{.site_context.brand_tone}}' ||
        E'\n\n{{if .site_context.about_us}}## About the Business' ||
        E'\n{{.site_context.about_us}}{{end}}' ||
        E'\n\n{{if .site_context.services}}## Services Offered' ||
        E'\n{{range .site_context.services}}- {{.name}}: {{.description}}' ||
        E'\n{{end}}{{end}}' ||
        E'\n\n## Components Used' ||
        E'\n{{range .site_context.all_component_functions}}- {{.}}' ||
        E'\n{{end}}' ||
        E'\n\nUsing the industry, business description, brand tone, and services above, ' ||
        'choose colors and typography that are appropriate and distinctive for this specific industry. ' ||
        'Do NOT use generic defaults. A fuel distribution company should look different from a consulting firm.' ||
        E'\n\nReturn ONLY valid JSON:' ||
        E'\n{' ||
        E'\n  "color_scheme": {' ||
        E'\n    "primary": "#hex (industry-appropriate, NOT #2c3e50)",' ||
        E'\n    "secondary": "#hex",' ||
        E'\n    "accent": "#hex",' ||
        E'\n    "background": "#ffffff",' ||
        E'\n    "surface": "#hex",' ||
        E'\n    "text": "#333333",' ||
        E'\n    "text_muted": "#666666",' ||
        E'\n    "border": "#hex"' ||
        E'\n  },' ||
        E'\n  "typography": {' ||
        E'\n    "font_family": "appropriate font stack",' ||
        E'\n    "heading_font": "inherit or specific",' ||
        E'\n    "base_size": "16px",' ||
        E'\n    "line_height": "1.6"' ||
        E'\n  },' ||
        E'\n  "spacing": {' ||
        E'\n    "section_padding": "5rem 2rem",' ||
        E'\n    "container_max_width": "1200px"' ||
        E'\n  },' ||
        E'\n  "design_notes": "explain why these colors/fonts suit this industry"' ||
        E'\n}'
        )
                     )
WHERE type = 'webdesign-agent';

-- Verify the update
SELECT
    type,
    default_config->'workflow'->'steps'->'analyze_design'->'config'->>'prompt_template' as prompt_preview
FROM agent_definitions
WHERE type = 'webdesign-agent';

--

-- 062: Fix webdesign-agent generate_css prompt to match CSS Colour Inheritance Model
--
-- Problems in current prompt:
--   1. "DO NOT include component-specific selectors" — wrong. Global CSS should provide
--      base styles for ALL components. Components' inline <style> handles layout-only
--      overrides using var(--variable, fallback). Colors inherited from global cascade.
--   2. h1-h6 { color: var(--section-heading, var(--color-primary)) } — contracts say
--      h1-h6 must use color: inherit, NOT var(--color-primary)
--   3. p { color: var(--section-text, inherit) } — contracts say do NOT set color on p
--   4. blockquote gets explicit color/background — contracts say don't
--   5. No header/footer/nav structural styles — renders as bullet list
--
-- The correct cascade (from 003_contracts_and_standards.md):
--   styles.css:  body { color: var(--color-text); }  ← only place default text set
--                h1-h6 { color: inherit; }            ← NOT var(--color-primary)
--                p, li, strong, blockquote             ← do NOT set color at all
--   Component:   .dark-section { color: #fff; }       ← overrides, children inherit
--   Inline CSS:  layout-only overrides using var(--variable, fallback)


UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_css,config,prompt_template}',
        to_jsonb(
                E'Generate a complete production CSS stylesheet.\n\n## Design Spec\n{{.design_spec.result}}\n\n## Components on this site\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\n## CSS ARCHITECTURE\n\nThis stylesheet is the global base. It provides:\n1. CSS custom properties (:root variables)\n2. Base typography and body styles\n3. Base styles for EVERY component class listed above\n4. Site structural layout (header, footer, nav)\n\nComponents have their own inline <style> blocks that handle layout-only overrides\nusing var(--variable, fallback). Colors are INHERITED from this global stylesheet,\nnot set in component inline styles (except dark-section overrides).\n\n## Required :root Variables\n:root {\n  --color-primary: (from color_scheme.primary);\n  --color-secondary: (from color_scheme.secondary);\n  --color-accent: (from color_scheme.accent);\n  --color-background: (from color_scheme.background);\n  --color-surface: (from color_scheme.surface);\n  --color-text: (from color_scheme.text);\n  --color-text-muted: (from color_scheme.text_muted);\n  --color-border: (from color_scheme.border);\n  --color-white: #ffffff;\n  --font-family: (from typography.font_family);\n  --spacing-section: (from spacing.section_padding);\n  --container-max-width: (from spacing.container_max_width);\n  --transition-speed: 0.3s;\n  --border-radius: 8px;\n  --shadow-sm: 0 2px 4px rgba(0,0,0,0.1);\n  --shadow-md: 0 4px 6px rgba(0,0,0,0.1);\n  --shadow-lg: 0 10px 20px rgba(0,0,0,0.15);\n}\n\n## CSS COLOUR INHERITANCE MODEL — FOLLOW EXACTLY\n\nThis is the single most important rule. Getting it wrong causes unreadable text.\n\n body sets color: var(--color-text) — the ONLY place default text colour is set.\n\n h1, h2, h3, h4, h5, h6 { color: inherit; }\n   NOT color: var(--color-primary). Headings MUST inherit so dark sections work.\n\n p, li, blockquote, strong, em, cite, span — do NOT set color at all.\n   They inherit from their parent container.\n\n a { color: var(--color-accent); } — links are the ONE exception (explicit colour).\n\n blockquote — do NOT set background-color (components handle this contextually).\n\nWhy: If base CSS forces color on p or blockquote, dark sections break because\nchildren cannot inherit color: #fff from their parent dark container.\n\n## Required Base Styles\n- *, *::before, *::after { box-sizing: border-box; }\n- html { scroll-behavior: smooth; }\n- html, body { margin: 0; padding: 0; }\n- body { font-family: var(--font-family); color: var(--color-text); line-height: 1.6; background-color: var(--color-background); }\n- .container { max-width: var(--container-max-width); margin: 0 auto; padding: 0 2rem; }\n\n## Typography\n- h1 { font-size: clamp(2rem, 5vw, 3rem); color: inherit; }\n- h2 { font-size: clamp(1.75rem, 4vw, 2.5rem); color: inherit; }\n- h3 { font-size: clamp(1.25rem, 3vw, 1.5rem); color: inherit; }\n- a { color: var(--color-accent); text-decoration: none; }\n- a:hover { text-decoration: underline; }\n\n## REQUIRED: Site Structural Layout\n\nThe header and footer templates use these exact class names. Include ALL of these:\n\n.site-header {\n  background: var(--color-primary);\n  padding: 1rem 0;\n  position: sticky;\n  top: 0;\n  z-index: 1000;\n  box-shadow: var(--shadow-sm);\n}\n.header-container {\n  max-width: var(--container-max-width);\n  margin: 0 auto;\n  padding: 0 2rem;\n  display: flex;\n  align-items: center;\n  justify-content: space-between;\n}\n.logo { text-decoration: none; font-size: 1.5rem; font-weight: 700; color: var(--color-white); }\n.logo-text { color: var(--color-white); }\n.logo-accent { color: var(--color-accent); }\n.logo-img { max-height: 40px; width: auto; display: block; }\n.main-nav ul { display: flex; list-style: none; margin: 0; padding: 0; gap: 2rem; }\n.main-nav a { color: rgba(255,255,255,0.9); text-decoration: none; font-weight: 500; padding: 0.5rem 0; transition: color var(--transition-speed); }\n.main-nav a:hover, .main-nav a.active { color: var(--color-accent); }\n.mobile-menu-toggle { display: none; background: none; border: none; cursor: pointer; padding: 0.5rem; }\n.mobile-menu-toggle span { display: block; width: 24px; height: 2px; background: white; margin: 5px 0; }\n\n.site-footer { background: var(--color-primary); color: rgba(255,255,255,0.8); padding: 3rem 0 1.5rem; }\n.footer-container { max-width: var(--container-max-width); margin: 0 auto; padding: 0 2rem; }\n.footer-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 2rem; margin-bottom: 2rem; }\n.footer-column h3, .footer-column h4 { color: var(--color-white); margin-bottom: 1rem; }\n.footer-column ul { list-style: none; padding: 0; margin: 0; }\n.footer-column li { margin-bottom: 0.5rem; }\n.footer-column a { color: rgba(255,255,255,0.7); text-decoration: none; transition: color var(--transition-speed); }\n.footer-column a:hover { color: var(--color-white); }\n.footer-bottom { border-top: 1px solid rgba(255,255,255,0.2); padding-top: 1.5rem; text-align: center; font-size: 0.875rem; color: rgba(255,255,255,0.5); }\n\n@media (max-width: 768px) {\n  .mobile-menu-toggle { display: block; }\n  .main-nav { position: absolute; top: 100%; left: 0; right: 0; background: var(--color-primary); padding: 1rem; display: none; }\n  .main-nav.active { display: block; }\n  .main-nav ul { flex-direction: column; gap: 0; }\n  .main-nav a { display: block; padding: 0.75rem 0; border-bottom: 1px solid rgba(255,255,255,0.1); }\n  .footer-grid { grid-template-columns: 1fr; }\n}\n\n## REQUIRED: Component Base Styles\n\nGenerate base styles for EVERY component listed in the Components section above.\nThese provide the default look. Components inline <style> handles layout-only\noverrides using var(--variable, fallback) — colors inherited from here.\n\nPattern for each component:\n  .{component}-section { padding: var(--spacing-section); }\n  .{component}-container { max-width: var(--container-max-width); margin: 0 auto; }\n  Then appropriate grid/flex layout, card styles, spacing for that component type.\n\nCommon component patterns:\n- hero sections: full-width, min-height, centered text, dark background with --section-* vars\n- features/services grids: CSS grid, responsive columns, card styling\n- testimonials: card layout, quote styling, attribution\n- call-to-action: centered text, prominent button\n- contact forms: form field styling, label layout\n- FAQ accordions: expandable items, borders\n- team grids: photo + name + title cards\n- pricing tiers: comparison columns\n\nDark-background components (hero, social-proof, testimonials, call-to-action) need:\n  .hero-section {\n    background: var(--color-primary);\n    color: var(--color-white);\n    --section-text: rgba(255,255,255,0.9);\n    --section-text-muted: rgba(255,255,255,0.7);\n    --section-heading: #ffffff;\n    --section-surface: rgba(255,255,255,0.05);\n    --section-border: rgba(255,255,255,0.2);\n  }\n\n## Also Include\n- Button base styles (.btn, .btn-primary, .btn-secondary, .btn-large, .btn-small)\n- Focus-visible states for accessibility\n- Smooth transitions\n- Responsive adjustments at 768px and 1024px\n- prefers-reduced-motion media query\n- Form input styling\n\n## Include at end as CSS comment: Dark Section Template\n/* DARK SECTION TEMPLATE — components with dark backgrounds set these:\n   .my-dark-section {\n     --section-text: rgba(255,255,255,0.9);\n     --section-text-muted: rgba(255,255,255,0.7);\n     --section-heading: #ffffff;\n     --section-surface: rgba(255,255,255,0.05);\n     --section-border: rgba(255,255,255,0.2);\n   }\n*/\n\nOutput ONLY CSS. No markdown fences. No explanations. Start with :root {'::text
        )
                     )
WHERE type = 'webdesign-agent';

---

-- patch

-- 060_patch_webdesign_read_site_specs.sql
--
-- Problem: webdesign-agent's analyze_design prompt uses site_context fields
-- (industry, brand_tone, services, about_us) but these come from the OLD
-- load_site_for_design action which reads sites.content_data. In the new flow,
-- domain-research-classifier and domain-strategist write rich data to site_specs
-- (aspects: identity, classification, strategy). The webdesign agent never sees this.
--
-- Fix: Insert a read_site_specs step before analyze_design. Update the
-- analyze_design prompt to use site_specs data (identity, classification,
-- strategy) alongside whatever site_context provides.
--
-- NOTES:
-- - read_site_spec with NO "aspect" key in config returns ALL aspects mode.
--   The Go code (ReadSiteSpecAction) checks `if aspect != ""` — setting
--   "aspect": "all" would query WHERE aspect='all' which returns nothing.
--
-- - All-aspects mode returns: { "specs": { "identity": {...}, ... }, "aspects": [...] }
--   So template paths must go through .site_specs.specs.identity.* etc.
--
-- - When called with externally-provided site_context (scrape, no DB record),
--   site_context.site_id may be null. ReadSiteSpecInputSpec requires site_id.
--   A conditional step skips the read when there's no site_id.
--
-- Run order: 060 before 062.

-- Add check_has_site_id conditional
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_has_site_id}',
        '{
          "action": "conditional",
          "config": {
            "condition": "site_context.site_id != null AND site_context.site_id != ''",
            "then_step": "read_site_specs",
            "else_step": "analyze_design"
          },
          "description": "Skip site_specs read when no DB record (e.g. scrape-only path)"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'webdesign-agent';

-- Add read_site_specs step (NO "aspect" key — omitting it triggers all-aspects mode)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,read_site_specs}',
        '{
          "action": "read_site_spec",
          "config": {
            "site_id": "site_context.site_id"
          },
          "description": "Load identity, classification, strategy from site_specs",
          "next_step": "analyze_design",
          "output_field": "site_specs"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'webdesign-agent';

-- Re-route load_site_context → check_has_site_id (was → analyze_design)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_site_context,next_step}',
        '"check_has_site_id"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'webdesign-agent';

-- Re-route use_provided_context → check_has_site_id (was → analyze_design)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,use_provided_context,next_step}',
        '"check_has_site_id"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'webdesign-agent';

-- Update analyze_design input_fields to include site_specs
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_design,config,input_fields}',
        '["site_context", "site_specs"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'webdesign-agent';

-- Update analyze_design prompt to reference site_specs.specs.* paths
-- (all-aspects mode returns data under a "specs" key)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_design,config,prompt_template}',
        to_jsonb(
                'You are a web design expert. Analyze the site and output a design specification.

                ## Site
                Domain: {{.site_context.domain}}
                Company: {{if .site_specs.specs.identity.company_name}}{{.site_specs.specs.identity.company_name}}{{else}}{{.site_context.company_name}}{{end}}
                Industry: {{if .site_specs.specs.identity.industry}}{{.site_specs.specs.identity.industry}}{{else if .site_context.industry}}{{.site_context.industry}}{{else}}professional services{{end}}
                Sub-industry: {{if .site_specs.specs.identity.sub_industry}}{{.site_specs.specs.identity.sub_industry}}{{end}}
                Tagline: {{if .site_specs.specs.identity.tagline}}{{.site_specs.specs.identity.tagline}}{{else}}{{.site_context.tagline}}{{end}}
                Site Type: {{if .site_specs.specs.strategy.site_type}}{{.site_specs.specs.strategy.site_type}}{{else if .site_specs.specs.classification.site_type}}{{.site_specs.specs.classification.site_type}}{{else}}{{.site_context.site_type}}{{end}}
                Tone: {{if .site_specs.specs.strategy.tone}}{{.site_specs.specs.strategy.tone}}{{else if .site_specs.specs.classification.tone_suggestion}}{{.site_specs.specs.classification.tone_suggestion}}{{else}}{{.site_context.brand_tone}}{{end}}
                Target Audience: {{if .site_specs.specs.identity.target_audience}}{{.site_specs.specs.identity.target_audience}}{{end}}
                Value Proposition: {{if .site_specs.specs.strategy.value_proposition}}{{.site_specs.specs.strategy.value_proposition}}{{end}}

                {{if .site_specs.specs.identity.about_summary}}## About the Business
                {{.site_specs.specs.identity.about_summary}}{{else if .site_context.about_us}}## About the Business
                {{.site_context.about_us}}{{end}}

                {{if .site_specs.specs.identity.services}}## Services Offered
                {{range .site_specs.specs.identity.services}}- {{.name}}: {{.description}}
                {{end}}{{else if .site_context.services}}## Services Offered
                {{range .site_context.services}}- {{.name}}: {{.description}}
                {{end}}{{end}}

                {{if .site_specs.specs.identity.unique_selling_points}}## Unique Selling Points
                {{range .site_specs.specs.identity.unique_selling_points}}- {{.}}
                {{end}}{{end}}

                {{if .site_specs.specs.strategy.content_strategy}}## Content Strategy
                {{.site_specs.specs.strategy.content_strategy}}{{end}}

                ## Components Used
                {{range .site_context.all_component_functions}}- {{.}}
                {{end}}

                Using the industry, business description, tone, target audience, and services above,
                choose colors and typography that are appropriate and distinctive for this specific
                industry and brand. Do NOT use generic defaults. A fuel distribution company should
                look different from a consulting firm. A veterinary practice should feel different
                from a law firm.

                Return ONLY valid JSON:
                {
                  "color_scheme": {
                    "primary": "#hex (industry-appropriate, NOT #2c3e50)",
                    "secondary": "#hex",
                    "accent": "#hex",
                    "background": "#ffffff",
                    "surface": "#hex",
                    "text": "#333333",
                    "text_muted": "#666666",
                    "border": "#hex"
                  },
                  "typography": {
                    "font_family": "appropriate font stack",
                    "heading_font": "inherit or specific",
                    "base_size": "16px",
                    "line_height": "1.6"
                  },
                  "spacing": {
                    "section_padding": "5rem 2rem",
                    "container_max_width": "1200px"
                  },
                  "design_notes": "explain why these colors/fonts suit this industry and brand tone"
                }'::text
        )
                     ),
    updated_at = NOW()
WHERE type = 'webdesign-agent';

---

-- css inheritance

-- 062_patch_webdesign_css_inheritance.sql
--
-- Problem: The generate_css prompt tells the LLM to set explicit color
-- on p, h1-h6, blockquote, li, strong. This breaks dark sections because
-- children cannot override when the global CSS bypasses --section-* variables.
--
-- Fix: Replace the generate_css prompt_template with one that uses the
-- --section-* variable pattern from 003_contracts_and_standards:
--   - body sets color: var(--color-text) — the base default
--   - h1-h6 use color: var(--section-heading, var(--color-primary))
--   - p, li, blockquote use color: var(--section-text, inherit)
--   - strong, em, cite, span — do NOT set color at all
--   - a uses color: var(--color-accent) — the one exception
--   - Dark sections override --section-* on their container
--
-- NOTE: Run AFTER 060 (which adds the read_site_specs step).
-- This patch only touches the generate_css step.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_css,config,prompt_template}',
        to_jsonb(
                'Generate a complete production CSS stylesheet.

                ## Design Spec
                {{.design_spec.result}}

                ## Components on this site
                {{range .site_context.all_component_functions}}- {{.}}
                {{end}}

                ## CSS ARCHITECTURE

                This stylesheet is the global base. It provides:
                1. CSS custom properties (:root variables)
                2. Base typography and body styles
                3. Section context variables (--section-*) consumed by base element rules
                4. Site structural layout (header, footer, nav)

                Components have their own inline <style> blocks that handle layout-only concerns
                (grid, flexbox, spacing). Colors flow through CSS variables — components do NOT
                re-declare colors on h1-h6, p, a.

                ## Required :root Variables
                :root {
                  --color-primary: (from color_scheme.primary);
                  --color-secondary: (from color_scheme.secondary);
                  --color-accent: (from color_scheme.accent);
                  --color-background: (from color_scheme.background);
                  --color-surface: (from color_scheme.surface);
                  --color-text: (from color_scheme.text);
                  --color-text-muted: (from color_scheme.text_muted);
                  --color-border: (from color_scheme.border);
                  --color-white: #ffffff;
                  --font-family: (from typography.font_family);
                  --spacing-section: (from spacing.section_padding);
                  --container-max-width: (from spacing.container_max_width);
                  --transition-speed: 0.3s;
                  --border-radius: 8px;
                  --shadow-sm: 0 2px 4px rgba(0,0,0,0.1);
                  --shadow-md: 0 4px 6px rgba(0,0,0,0.1);
                  --shadow-lg: 0 10px 20px rgba(0,0,0,0.15);
                }

                ## SECTION CONTEXT VARIABLE MODEL — FOLLOW EXACTLY

                This is the single most important rule. Getting it wrong causes unreadable text.

                Text elements reference --section-* variables with light-theme fallbacks.
                Dark-section components override --section-* on their container.
                Everything adapts automatically — no specificity wars.

                body { color: var(--color-text); } — the base default.

                h1, h2, h3, h4, h5, h6 { color: var(--section-heading, var(--color-primary)); }
                  In light sections: --section-heading is not set, fallback gives --color-primary.
                  In dark sections: --section-heading is #ffffff.
                  Do NOT use color: inherit or color: var(--color-primary) directly.

                p, li, blockquote { color: var(--section-text, inherit); }
                  In light sections: --section-text is not set, fallback is inherit (from body).
                  In dark sections: --section-text is rgba(255,255,255,0.9).
                  Do NOT use color: var(--color-text) on these elements.

                strong, em, cite, span — do NOT set color at all. They inherit from their parent.

                a { color: var(--color-accent); } — links are the ONE exception (always explicit).

                blockquote — do NOT set background-color (components handle this contextually).

                WHY: If base CSS sets color: var(--color-text) on p, the --section-text override
                is bypassed. The element gets #333333 regardless of the dark section context.

                ## Required Base Styles
                - *, *::before, *::after { box-sizing: border-box; }
                - html { scroll-behavior: smooth; }
                - html, body { margin: 0; padding: 0; }
                - body { font-family: var(--font-family); color: var(--color-text); line-height: 1.6; background-color: var(--color-background); -webkit-font-smoothing: antialiased; }
                - .container { max-width: var(--container-max-width); margin: 0 auto; padding: 0 2rem; }

                ## Typography (using --section-* pattern)
                - h1 { font-size: clamp(2rem, 5vw, 3rem); color: var(--section-heading, var(--color-primary)); font-weight: 700; line-height: 1.2; margin: 0 0 1rem; }
                - h2 { font-size: clamp(1.75rem, 4vw, 2.5rem); color: var(--section-heading, var(--color-primary)); font-weight: 700; line-height: 1.2; margin: 0 0 1rem; }
                - h3 { font-size: clamp(1.25rem, 3vw, 1.5rem); color: var(--section-heading, var(--color-primary)); font-weight: 600; line-height: 1.3; margin: 0 0 0.75rem; }
                - h4, h5, h6 { color: var(--section-heading, var(--color-primary)); font-weight: 600; margin: 0 0 0.5rem; }
                - p { color: var(--section-text, inherit); margin: 0 0 1rem; line-height: 1.6; }
                - li { color: var(--section-text, inherit); }
                - blockquote { color: var(--section-text, inherit); font-style: italic; margin: 1rem 0; padding: 1rem 1.5rem; border-left: 3px solid var(--section-border, var(--color-border)); }
                - a { color: var(--color-accent); text-decoration: none; }
                - a:hover { text-decoration: underline; }

                ## REQUIRED: Site Structural Layout

                The header and footer templates use these exact class names. Include ALL of these:

                .site-header {
                  background: var(--color-primary);
                  padding: 1rem 0;
                  position: sticky;
                  top: 0;
                  z-index: 1000;
                  box-shadow: var(--shadow-sm);
                }
                .header-container {
                  max-width: var(--container-max-width);
                  margin: 0 auto;
                  padding: 0 2rem;
                  display: flex;
                  align-items: center;
                  justify-content: space-between;
                }
                .logo { text-decoration: none; font-size: 1.5rem; font-weight: 700; color: var(--color-white); }
                .logo-text { color: var(--color-white); }
                .logo-accent { color: var(--color-accent); }
                .logo-img { max-height: 40px; width: auto; display: block; }
                .main-nav ul { display: flex; list-style: none; margin: 0; padding: 0; gap: 2rem; }
                .main-nav a { color: rgba(255,255,255,0.9); text-decoration: none; font-weight: 500; padding: 0.5rem 0; transition: color var(--transition-speed); }
                .main-nav a:hover, .main-nav a.active { color: var(--color-accent); }
                .mobile-menu-toggle { display: none; background: none; border: none; cursor: pointer; padding: 0.5rem; }
                .mobile-menu-toggle span { display: block; width: 24px; height: 2px; background: white; margin: 5px 0; }

                .site-footer { background: var(--color-primary); color: rgba(255,255,255,0.8); padding: 3rem 0 1.5rem; }
                .footer-container { max-width: var(--container-max-width); margin: 0 auto; padding: 0 2rem; }
                .footer-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 2rem; margin-bottom: 2rem; }
                .footer-column h3, .footer-column h4 { color: var(--color-white); margin-bottom: 1rem; }
                .footer-column ul { list-style: none; padding: 0; margin: 0; }
                .footer-column li { margin-bottom: 0.5rem; }
                .footer-column a { color: rgba(255,255,255,0.7); text-decoration: none; transition: color var(--transition-speed); }
                .footer-column a:hover { color: var(--color-white); }
                .footer-bottom { border-top: 1px solid rgba(255,255,255,0.2); padding-top: 1.5rem; text-align: center; font-size: 0.875rem; color: rgba(255,255,255,0.5); }

                @media (max-width: 768px) {
                  .mobile-menu-toggle { display: block; }
                  .main-nav { position: absolute; top: 100%; left: 0; right: 0; background: var(--color-primary); padding: 1rem; display: none; }
                  .main-nav.active { display: block; }
                  .main-nav ul { flex-direction: column; gap: 0; }
                  .main-nav a { display: block; padding: 0.75rem 0; border-bottom: 1px solid rgba(255,255,255,0.1); }
                  .footer-grid { grid-template-columns: 1fr; }
                }

                ## REQUIRED: Section Container Styles

                For each component listed above, provide a section container with padding and
                a container div for max-width. This is NOT component-specific layout — it is
                the section-level wrapper that the global stylesheet owns.

                Pattern:
                  .{component}-section { padding: var(--spacing-section); }

                Dark-background sections (hero, social-proof, testimonials, call-to-action) MUST
                set --section-* overrides on their container:

                  .hero-section {
                    background: var(--color-primary);
                    color: var(--color-white);
                    --section-text: rgba(255,255,255,0.9);
                    --section-text-muted: rgba(255,255,255,0.7);
                    --section-heading: #ffffff;
                    --section-surface: rgba(255,255,255,0.05);
                    --section-border: rgba(255,255,255,0.2);
                  }

                Light sections need no overrides — the fallback values in base element rules apply.

                ## Also Include
                - Button base styles (.btn, .btn-primary, .btn-secondary, .btn-large, .btn-small)
                - Focus-visible states for accessibility
                - Smooth transitions on interactive elements
                - Responsive adjustments at 768px and 1024px
                - prefers-reduced-motion media query
                - Form input styling (border, padding, focus state)

                ## DO NOT Include
                - Component-internal layout (grid columns, card layouts, flexbox arrangements)
                - Components have their own inline CSS for layout

                ## Include at end as CSS comment: Dark Section Template
                /* DARK SECTION TEMPLATE — components with dark backgrounds set these:
                   .my-dark-section {
                     --section-text: rgba(255,255,255,0.9);
                     --section-text-muted: rgba(255,255,255,0.7);
                     --section-heading: #ffffff;
                     --section-surface: rgba(255,255,255,0.05);
                     --section-border: rgba(255,255,255,0.2);
                   }
                */

                Output ONLY CSS. No markdown fences. No explanations. Start with :root {'::text
        )
                     ),
    updated_at = NOW()
WHERE type = 'webdesign-agent';