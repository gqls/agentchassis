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