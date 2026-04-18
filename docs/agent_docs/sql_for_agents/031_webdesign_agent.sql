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

--

-- bugfix missing '

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_has_site_id,config,condition}',
        '"site_context.site_id != null"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'webdesign-agent';

-- fix css generation, with easier template for llm to read (in datahelpers)

-- Migration 069: webdesign-agent CSS generation fixes
--
-- Problem 1: check_site_context has truncated condition string
--   "input_data.site_context.domain != null AND input_data.site_context.domain != '"
--   The trailing quote was eaten by SQL escaping. Same bug as check_has_site_id (fixed in prior session).
--
-- Problem 2: generate_css prompt template uses {{.design_spec.result}} which renders
--   Go maps as "map[key:val ...]" format instead of JSON. The LLM can't reliably parse
--   the colors from this format and falls back to reproducing the template's default CSS.
--
-- Fix 1: Simplify check_site_context condition to just null check.
-- Fix 2: Change {{.design_spec.result}} to {{.design_spec.result | toJSON}} in the prompt.
--   Requires the corresponding Go patch (toJSON template function in RenderPromptTemplate).
--
-- NOTE: The | toJSON pipe depends on the Go code change being deployed first.
--   If deployed in wrong order, the template will fail to parse and the step will error.
--   Deploy Go code first, then apply this migration.

BEGIN;

-- Fix 1: check_site_context truncated condition
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_site_context,config,condition}',
        '"input_data.site_context.domain != null"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'webdesign-agent';

-- Fix 2: update generate_css prompt to use | toJSON for design_spec rendering
--
-- We replace the prompt_template value. The old template has:
--   "## Design Spec\\n{{.design_spec.result}}\\n"
-- We change it to:
--   "## Design Spec\\n{{.design_spec.result | toJSON}}\\n"
--
-- Since the prompt_template is a very long string in JSONB, we use a targeted
-- text replacement via a DO block rather than replacing the entire value.

DO $$
DECLARE
current_prompt TEXT;
    new_prompt TEXT;
BEGIN
    -- Extract the current prompt template
SELECT default_config #>> '{workflow,steps,generate_css,config,prompt_template}'
INTO current_prompt
FROM agent_definitions
WHERE type = 'webdesign-agent';

IF current_prompt IS NULL THEN
        RAISE NOTICE 'webdesign-agent generate_css prompt_template not found, skipping';
        RETURN;
END IF;

    -- Replace the template expression
    -- The double-backslash in the stored JSON becomes literal \n in the extracted text
    new_prompt := replace(
        current_prompt,
        '{{.design_spec.result}}',
        '{{.design_spec.result | toJSON}}'
    );

    -- Only update if something actually changed
    IF new_prompt = current_prompt THEN
        RAISE NOTICE 'No change needed — {{.design_spec.result}} not found in prompt (may already be fixed)';
        RETURN;
END IF;

    -- Write it back
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_css,config,prompt_template}',
        to_jsonb(new_prompt)
                     ),
    updated_at = NOW()
WHERE type = 'webdesign-agent';

RAISE NOTICE 'Updated generate_css prompt: {{.design_spec.result}} → {{.design_spec.result | toJSON}}';
END;
$$;

-- Verify
SELECT
    type,
    default_config #>> '{workflow,steps,check_site_context,config,condition}' AS check_site_context_cond,
    CASE
        WHEN default_config #>> '{workflow,steps,generate_css,config,prompt_template}'
             LIKE '%toJSON%' THEN 'HAS toJSON'
        ELSE 'MISSING toJSON'
END AS prompt_has_tojson,
    updated_at
FROM agent_definitions
WHERE type = 'webdesign-agent';

COMMIT;


---

-- Migration 070: Deterministic CSS generation via Go template
--
-- Replaces the LLM-based generate_css step in webdesign-agent with a
-- deterministic render_css_from_spec action that uses a Go text/template
-- stored in css_themes.css_template.
--
-- Changes:
--   1. Add css_template column to css_themes table
--   2. Insert "standard-brochure" template (the default Go template)
--   3. Update webdesign-agent workflow: generate_css step uses render_css_from_spec action
--
-- The analyze_design step (LLM) still picks industry-appropriate colors/fonts.
-- This migration removes only the second LLM call that was doing mechanical substitution.
--
-- Rollback: UPDATE agent_definitions SET default_config = <old config> WHERE type = 'webdesign-agent';
--           DELETE FROM css_themes WHERE name = 'standard-brochure';
--           ALTER TABLE css_themes DROP COLUMN css_template;

BEGIN;

-- ============================================================================
-- 1. Add css_template column to css_themes
-- ============================================================================
ALTER TABLE css_themes ADD COLUMN IF NOT EXISTS css_template text;

COMMENT ON COLUMN css_themes.css_template IS
  'Go text/template for CSS generation. Rendered by render_css_from_spec action. '
  'If NULL, this theme uses static css_content only (backward compatible).';

-- ============================================================================
-- 2. Insert standard-brochure Go template
-- ============================================================================
INSERT INTO css_themes (name, display_name, description, category, css_content, css_template, semantic_tags)
VALUES (
           'standard-brochure',
           'Standard Brochure',
           'Default CSS template for multi-page business/brochure sites. Provides :root variables, base typography, section context model, header/footer layout, buttons, forms, responsive breakpoints.',
           'brochure',
           '', -- css_content left empty; the template generates CSS at runtime
           -- ↓↓↓ The Go text/template ↓↓↓
           $TMPL$:root {
  --color-primary: {{.Primary}};
  --color-secondary: {{.Secondary}};
  --color-accent: {{.Accent}};
  --color-background: {{.Background}};
  --color-surface: {{.Surface}};
  --color-text: {{.Text}};
  --color-text-muted: {{.TextMuted}};
  --color-border: {{.Border}};
  --color-white: #ffffff;
  --font-family: {{.FontFamily}};
  --spacing-section: {{.SectionPadding}};
  --container-max-width: {{.ContainerMaxWidth}};
  --transition-speed: 0.3s;
  --border-radius: 8px;
  --shadow-sm: 0 2px 4px rgba(0,0,0,0.1);
  --shadow-md: 0 4px 6px rgba(0,0,0,0.1);
  --shadow-lg: 0 10px 20px rgba(0,0,0,0.15);
}

*, *::before, *::after {
  box-sizing: border-box;
}

html {
  scroll-behavior: smooth;
  margin: 0;
  padding: 0;
}

body {
  margin: 0;
  padding: 0;
  font-family: var(--font-family);
  color: var(--color-text);
  line-height: {{.LineHeight}};
  background-color: var(--color-background);
  -webkit-font-smoothing: antialiased;
}

.container {
  max-width: var(--container-max-width);
  margin: 0 auto;
  padding: 0 2rem;
}

/* ── Typography ── */

h1 {
  font-size: clamp(2rem, 5vw, 3rem);
  color: var(--section-heading, var(--color-primary));
  font-weight: 700;
  line-height: 1.2;
  margin: 0 0 1rem;
}

h2 {
  font-size: clamp(1.75rem, 4vw, 2.5rem);
  color: var(--section-heading, var(--color-primary));
  font-weight: 700;
  line-height: 1.2;
  margin: 0 0 1rem;
}

h3 {
  font-size: clamp(1.25rem, 3vw, 1.5rem);
  color: var(--section-heading, var(--color-primary));
  font-weight: 600;
  line-height: 1.3;
  margin: 0 0 0.75rem;
}

h4, h5, h6 {
  color: var(--section-heading, var(--color-primary));
  font-weight: 600;
  margin: 0 0 0.5rem;
}

p {
  color: var(--section-text, inherit);
  margin: 0 0 1rem;
  line-height: {{.LineHeight}};
}

li {
  color: var(--section-text, inherit);
}

blockquote {
  color: var(--section-text, inherit);
  font-style: italic;
  margin: 1rem 0;
  padding: 1rem 1.5rem;
  border-left: 3px solid var(--section-border, var(--color-border));
}

a {
  color: var(--color-accent);
  text-decoration: none;
  transition: color var(--transition-speed);
}

a:hover {
  text-decoration: underline;
}

/* ── Site Header ── */

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

.logo {
  text-decoration: none;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-white);
}

.logo-text {
  color: var(--color-white);
}

.logo-accent {
  color: var(--color-accent);
}

.logo-img {
  max-height: 40px;
  width: auto;
  display: block;
}

.main-nav ul {
  display: flex;
  list-style: none;
  margin: 0;
  padding: 0;
  gap: 2rem;
}

.main-nav a {
  color: rgba(255,255,255,0.9);
  text-decoration: none;
  font-weight: 500;
  padding: 0.5rem 0;
  transition: color var(--transition-speed);
}

.main-nav a:hover, .main-nav a.active {
  color: var(--color-accent);
}

.mobile-menu-toggle {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
}

.mobile-menu-toggle span {
  display: block;
  width: 24px;
  height: 2px;
  background: white;
  margin: 5px 0;
}

/* ── Site Footer ── */

.site-footer {
  background: var(--color-primary);
  color: rgba(255,255,255,0.8);
  padding: 3rem 0 1.5rem;
}

.footer-container {
  max-width: var(--container-max-width);
  margin: 0 auto;
  padding: 0 2rem;
}

.footer-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 2rem;
  margin-bottom: 2rem;
}

.footer-column h3, .footer-column h4 {
  color: var(--color-white);
  margin-bottom: 1rem;
}

.footer-column ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.footer-column li {
  margin-bottom: 0.5rem;
}

.footer-column a {
  color: rgba(255,255,255,0.7);
  text-decoration: none;
  transition: color var(--transition-speed);
}

.footer-column a:hover {
  color: var(--color-white);
}

.footer-bottom {
  border-top: 1px solid rgba(255,255,255,0.2);
  padding-top: 1.5rem;
  text-align: center;
  font-size: 0.875rem;
  color: rgba(255,255,255,0.5);
}

/* ── Section Containers ── */
{{range .SectionStyles}}
.{{.ClassName}} {
  padding: var(--spacing-section);{{if .IsDark}}
  background: var(--color-primary);
  color: var(--color-white);
  --section-text: rgba(255,255,255,0.9);
  --section-text-muted: rgba(255,255,255,0.7);
  --section-heading: #ffffff;
  --section-surface: rgba(255,255,255,0.05);
  --section-border: rgba(255,255,255,0.2);{{end}}
}
{{end}}
/* Alternating light sections get surface background for visual rhythm */
.differentiators-section,
.features-section,
.faq-section {
  background: var(--color-surface);
}

/* ── Buttons ── */

.btn {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
  font-weight: 600;
  text-align: center;
  text-decoration: none;
  border-radius: var(--border-radius);
  border: 2px solid transparent;
  cursor: pointer;
  transition: all var(--transition-speed);
  font-family: var(--font-family);
}

.btn-primary {
  background: var(--color-accent);
  color: var(--color-white);
  border-color: var(--color-accent);
}

.btn-primary:hover {
  background: var(--color-secondary);
  border-color: var(--color-secondary);
  text-decoration: none;
}

.btn-secondary {
  background: transparent;
  color: var(--color-accent);
  border-color: var(--color-accent);
}

.btn-secondary:hover {
  background: var(--color-accent);
  color: var(--color-white);
  text-decoration: none;
}

.btn-large {
  padding: 1rem 2rem;
  font-size: 1.125rem;
}

.btn-small {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
}

/* ── Form Elements ── */

input, textarea, select {
  font-family: var(--font-family);
  font-size: 1rem;
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius);
  background: var(--color-white);
  color: var(--color-text);
  transition: border-color var(--transition-speed), box-shadow var(--transition-speed);
}

input:focus, textarea:focus, select:focus {
  outline: none;
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px rgba(49, 130, 206, 0.1);
}

/* ── Accessibility ── */

button:focus-visible, a:focus-visible, input:focus-visible, textarea:focus-visible, select:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}

/* ── Responsive ── */

@media (max-width: 768px) {
  .mobile-menu-toggle {
    display: block;
  }

  .main-nav {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background: var(--color-primary);
    padding: 1rem;
    display: none;
  }

  .main-nav.active {
    display: block;
  }

  .main-nav ul {
    flex-direction: column;
    gap: 0;
  }

  .main-nav a {
    display: block;
    padding: 0.75rem 0;
    border-bottom: 1px solid rgba(255,255,255,0.1);
  }

  .footer-grid {
    grid-template-columns: 1fr;
  }

  :root {
    --spacing-section: 3rem 0;
  }
}

@media (max-width: 1024px) {
  .container {
    padding: 0 1.5rem;
  }

  .header-container {
    padding: 0 1.5rem;
  }

  .footer-container {
    padding: 0 1.5rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }

  html {
    scroll-behavior: auto;
  }
}

/* DARK SECTION TEMPLATE — components with dark backgrounds set these:
   .my-dark-section {
     --section-text: rgba(255,255,255,0.9);
     --section-text-muted: rgba(255,255,255,0.7);
     --section-heading: #ffffff;
     --section-surface: rgba(255,255,255,0.05);
     --section-border: rgba(255,255,255,0.2);
   }
*/$TMPL$,
  ARRAY['brochure', 'business', 'default']
)
ON CONFLICT (name) DO UPDATE
SET css_template = EXCLUDED.css_template,
    description = EXCLUDED.description,
    updated_at = NOW();

-- ============================================================================
-- 3. Update webdesign-agent: replace generate_css LLM step with render_css_from_spec
-- ============================================================================
-- Change only the generate_css step. The workflow wiring stays the same:
--   analyze_design → generate_css → deploy_css
-- The step name stays "generate_css" to avoid changing any step references.
-- Only the action and config change.

UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,generate_css}',
    '{
      "action": "render_css_from_spec",
      "config": {
        "theme_name": "standard-brochure"
      },
      "next_step": "deploy_css",
      "description": "Render CSS from design spec using Go template (deterministic, no LLM)",
      "output_field": "generated_css"
    }'::jsonb
),
updated_at = NOW()
WHERE type = 'webdesign-agent';

-- ============================================================================
-- Verify
-- ============================================================================
SELECT
    type,
    default_config #>> '{workflow,steps,generate_css,action}' AS generate_css_action,
    default_config #>> '{workflow,steps,generate_css,config,theme_name}' AS theme_name,
    updated_at
FROM agent_definitions
WHERE type = 'webdesign-agent';

SELECT name, category,
    CASE WHEN css_template IS NOT NULL AND css_template != '' THEN 'has template (' || length(css_template) || ' chars)'
         ELSE 'static only'
    END AS template_status
FROM css_themes
WHERE name = 'standard-brochure';

COMMIT;

      --

      -- services is an array of strings like ["consulting", "AI strategy"] instead of [{"name": "Consulting", "description": "..."}], that's the mismatch.
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,analyze_design,config,prompt_template}',
    to_jsonb(
        replace(
            replace(
                default_config->'workflow'->'steps'->'analyze_design'->'config'->>'prompt_template',
                '{{range .site_specs.specs.identity.services}}- {{.name}}: {{.description}}',
                '{{range .site_specs.specs.identity.services}}- {{if .name}}{{.name}}: {{.description}}{{else}}{{.}}{{end}}'
            ),
            '{{range .site_context.services}}- {{.name}}: {{.description}}',
            '{{range .site_context.services}}- {{if .name}}{{.name}}: {{.description}}{{else}}{{.}}{{end}}'
        )
    )
)
WHERE type = 'webdesign-agent' AND deleted_at IS NULL;

-- Verify
SELECT substring(
    default_config->'workflow'->'steps'->'analyze_design'->'config'->>'prompt_template'
    FROM 'range .site_context.services.*?\n'
) FROM agent_definitions
WHERE type = 'webdesign-agent' AND deleted_at IS NULL;


---
      --

      -- backup
 9dc5f47a-7c1f-461a-80f1-8ae37feca24e | webdesign-agent | Web Design Agent | Generates CSS stylesheets for sites. Accepts site_context or loads from DB. Analyzes design requirements and generates production CSS. | specialist | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_fields": ["design_spec", "css_deployed", "site_context"]}}, "deploy_css": {"action": "git_commit", "config": {"file_path": "assets/css/styles.css", "domain_field": "site_context.domain", "content_field": "generated_css.result", "commit_message": "Update stylesheet via webdesign-agent"}, "next_step": "check_update_db", "description": "Deploy CSS to git", "output_field": "css_deployed"}, "update_site": {"action": "update_site_content", "config": {"merge": true, "content_field": "design_spec.result", "site_id_field": "site_context.site_id"}, "next_step": "complete", "description": "Store design spec", "output_field": "site_updated"}, "generate_css": {"action": "render_css_from_spec", "config": {"theme_name": "standard-brochure"}, "next_step": "deploy_css", "description": "Render CSS from design spec using Go template (deterministic, no LLM)", "output_field": "generated_css"}, "analyze_design": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 2000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_context", "site_specs"], "output_format": "json", "prompt_template": "You are a web design expert. Analyze the site and output a design specification.\n\n## Site\nDomain: {{.site_context.domain}}\nCompany: {{if .site_specs.specs.identity.company_name}}{{.site_specs.specs.identity.company_name}}{{else}}{{.site_context.company_name}}{{end}}\nIndustry: {{if .site_specs.specs.identity.industry}}{{.site_specs.specs.identity.industry}}{{else if .site_context.industry}}{{.site_context.industry}}{{else}}professional services{{end}}\nSub-industry: {{if .site_specs.specs.identity.sub_industry}}{{.site_specs.specs.identity.sub_industry}}{{end}}\nTagline: {{if .site_specs.specs.identity.tagline}}{{.site_specs.specs.identity.tagline}}{{else}}{{.site_context.tagline}}{{end}}\nSite Type: {{if .site_specs.specs.strategy.site_type}}{{.site_specs.specs.strategy.site_type}}{{else if .site_specs.specs.classification.site_type}}{{.site_specs.specs.classification.site_type}}{{else}}{{.site_context.site_type}}{{end}}\nTone: {{if .site_specs.specs.strategy.tone}}{{.site_specs.specs.strategy.tone}}{{else if .site_specs.specs.classification.tone_suggestion}}{{.site_specs.specs.classification.tone_suggestion}}{{else}}{{.site_context.brand_tone}}{{end}}\nTarget Audience: {{if .site_specs.specs.identity.target_audience}}{{.site_specs.specs.identity.target_audience}}{{end}}\nValue Proposition: {{if .site_specs.specs.strategy.value_proposition}}{{.site_specs.specs.strategy.value_proposition}}{{end}}\n\n{{if .site_specs.specs.identity.about_summary}}## About the Business\n{{.site_specs.specs.identity.about_summary}}{{else if .site_context.about_us}}## About the Business\n{{.site_context.about_us}}{{end}}\n\n{{if .site_specs.specs.identity.services}}## Services Offered\n{{range .site_specs.specs.identity.services}}- {{if .name}}{{.name}}: {{.description}}{{else}}{{.}}{{end}}\n{{end}}{{else if .site_context.services}}## Services Offered\n{{range .site_context.services}}- {{if .name}}{{.name}}: {{.description}}{{else}}{{.}}{{end}}\n{{end}}{{end}}\n\n{{if .site_specs.specs.identity.unique_selling_points}}## Unique Selling Points\n{{range .site_specs.specs.identity.unique_selling_points}}- {{.}}\n{{end}}{{end}}\n\n{{if .site_specs.specs.strategy.content_strategy}}## Content Strategy\n{{.site_specs.specs.strategy.content_strategy}}{{end}}\n\n## Components Used\n{{range .site_context.all_component_functions}}- {{.}}\n{{end}}\n\nUsing the industry, business description, tone, target audience, and services above,\nchoose colors and typography that are appropriate and distinctive for this specific\nindustry and brand. Do NOT use generic defaults. A fuel distribution company should\nlook different from a consulting firm. A veterinary practice should feel different\nfrom a law firm.\n\nReturn ONLY valid JSON:\n{\n  \"color_scheme\": {\n    \"primary\": \"#hex (industry-appropriate, NOT #2c3e50)\",\n    \"secondary\": \"#hex\",\n    \"accent\": \"#hex\",\n    \"background\": \"#ffffff\",\n    \"surface\": \"#hex\",\n    \"text\": \"#333333\",\n    \"text_muted\": \"#666666\",\n    \"border\": \"#hex\"\n  },\n  \"typography\": {\n    \"font_family\": \"appropriate font stack\",\n    \"heading_font\": \"inherit or specific\",\n    \"base_size\": \"16px\",\n    \"line_height\": \"1.6\"\n  },\n  \"spacing\": {\n    \"section_padding\": \"5rem 2rem\",\n    \"container_max_width\": \"1200px\"\n  },\n  \"design_notes\": \"explain why these colors/fonts suit this industry and brand tone\"\n}"}, "next_step": "generate_css", "description": "Generate design spec", "output_field": "design_spec"}, "check_update_db": {"action": "conditional", "config": {"condition": "site_context.site_id != null", "else_step": "complete", "then_step": "update_site"}, "description": "Check if we should update DB"}, "read_site_specs": {"action": "read_site_spec", "config": {"site_id": "site_context.site_id"}, "next_step": "analyze_design", "description": "Load identity, classification, strategy from site_specs", "output_field": "site_specs"}, "check_has_site_id": {"action": "conditional", "config": {"condition": "site_context.site_id != null", "else_step": "analyze_design", "then_step": "read_site_specs"}, "description": "Skip site_specs read when no DB record (e.g. scrape-only path)"}, "load_site_context": {"action": "load_site_for_design", "config": {"domain_field": "input_data.domain", "include_pages": true, "site_id_field": "input_data.site_id", "include_style_collection": true}, "next_step": "check_has_site_id", "description": "Load site from database", "output_field": "site_context"}, "check_site_context": {"action": "conditional", "config": {"condition": "input_data.site_context.domain != null", "else_step": "load_site_context", "then_step": "use_provided_context"}, "description": "Check if site_context was provided"}, "use_provided_context": {"action": "transform_data", "config": {"output_key": "site_context", "source_field": "input_data.site_context"}, "next_step": "check_has_site_id", "description": "Use provided site_context", "output_field": "site_context"}}, "start_step": "check_site_context"}, "processing_mode": "task", "timeout_seconds": 300} | t         | 2026-01-29 15:39:53.19592+00 | 2026-04-11 15:39:41.731404+00 |            | ["design", "css", "styling", "specialist"] | docker.io/aqls/agent-chassis | v1.0.951  |         | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} |                | experimental | []          | {}                     |           0 | f           | {"optional": ["site_id", "domain", "site_context"], "required": []} | {"produces": {"design_spec": "design spec", "css_deployed": "git result"}} |                  180
(1 row)

      ---
      -- The key change in the prompt: it now has a conditional branch. When design_reference exists in site_specs (adopted sites), the prompt says "here are the concrete values from the original site, use them as your starting point, do NOT invent a new palette." When it doesn't exist (new builds), it falls back to the existing instruction: "choose colours appropriate for this industry."
-- ============================================================================
-- Phase 2b: Update webdesign-agent analyze_design prompt
-- ============================================================================
-- Three-way priority for design direction:
--   1. design_intent exists → use it (semantic, creative freedom, evolution)
--   2. design_reference exists, no design_intent → reproduce reference (first adoption build)
--   3. neither → generate from industry (new build)
--
-- This determines WHEN the palette can change:
--   - First adoption: only design_reference exists → locked to reference values
--   - Strategist/human writes design_intent → agent has creative freedom within intent
--   - Improvement loop proposes design_intent update → palette evolves
-- ============================================================================

-- Verify current state
SELECT LEFT(
    default_config->'workflow'->'steps'->'analyze_design'->'config'->>'prompt_template',
    80
) as prompt_start
FROM agent_definitions
WHERE type = 'webdesign-agent';

-- Update the prompt_template
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
{{range .site_specs.specs.identity.services}}- {{if .name}}{{.name}}: {{.description}}{{else}}{{.}}{{end}}
{{end}}{{else if .site_context.services}}## Services Offered
{{range .site_context.services}}- {{if .name}}{{.name}}: {{.description}}{{else}}{{.}}{{end}}
{{end}}{{end}}

{{if .site_specs.specs.identity.unique_selling_points}}## Unique Selling Points
{{range .site_specs.specs.identity.unique_selling_points}}- {{.}}
{{end}}{{end}}

{{if .site_specs.specs.strategy.content_strategy}}## Content Strategy
{{.site_specs.specs.strategy.content_strategy}}{{end}}

## Components Used
{{range .site_context.all_component_functions}}- {{.}}
{{end}}

{{if .site_specs.specs.design_intent}}## Design Intent

A design direction has been set for this site. You have creative freedom to interpret this intent — the values below are guidance, not exact specifications. Your design should express the described character while fitting our CSS variable conventions.

{{if .site_specs.specs.design_intent.palette}}### Palette
{{if .site_specs.specs.design_intent.palette.character}}Character: {{.site_specs.specs.design_intent.palette.character}}{{end}}
{{if .site_specs.specs.design_intent.palette.reference_values}}Reference values (use as starting points, not exact targets):
{{range $key, $value := .site_specs.specs.design_intent.palette.reference_values}}  --color-{{$key}}: {{$value}}
{{end}}{{end}}
{{if .site_specs.specs.design_intent.palette.guidance}}Guidance: {{.site_specs.specs.design_intent.palette.guidance}}{{end}}
{{end}}

{{if .site_specs.specs.design_intent.typography}}### Typography
{{if .site_specs.specs.design_intent.typography.character}}Character: {{.site_specs.specs.design_intent.typography.character}}{{end}}
{{if .site_specs.specs.design_intent.typography.reference_values}}Reference values:
{{range $key, $value := .site_specs.specs.design_intent.typography.reference_values}}  {{$key}}: {{$value}}
{{end}}{{end}}
{{if .site_specs.specs.design_intent.typography.guidance}}Guidance: {{.site_specs.specs.design_intent.typography.guidance}}{{end}}
{{end}}

{{if .site_specs.specs.design_intent.spacing}}### Spacing
{{if .site_specs.specs.design_intent.spacing.character}}Character: {{.site_specs.specs.design_intent.spacing.character}}{{end}}
{{if .site_specs.specs.design_intent.spacing.reference_values}}Reference values:
{{range $key, $value := .site_specs.specs.design_intent.spacing.reference_values}}  {{$key}}: {{$value}}
{{end}}{{end}}
{{end}}

Generate a design that expresses the described character. You may adjust the reference values to better achieve the intent — explain your choices in design_notes.

{{else if .site_specs.specs.design_reference}}## Design Reference (Adopted Site — No Design Intent Yet)

This site was adopted from an existing design but no design direction has been set yet. The values below were extracted from the original site''s actual CSS. Use them directly — do NOT invent a new palette or font stack.

{{if .site_specs.specs.design_reference.suggested_mapping}}### Reference Values (mapped to our CSS variables)
Use these values directly:
{{range $key, $value := .site_specs.specs.design_reference.suggested_mapping}}  --color-{{$key}}: {{$value}}
{{end}}{{end}}

{{if .site_specs.specs.design_reference.css_variables}}### Original CSS Variables
The original site defined these custom properties:
{{range $key, $value := .site_specs.specs.design_reference.css_variables}}  {{$key}}: {{$value}}
{{end}}{{end}}

{{if .site_specs.specs.design_reference.dark_sections}}### Dark/Light Scheme
Predominant scheme: {{.site_specs.specs.design_reference.dark_sections.predominant_scheme}}
Has dark sections: {{.site_specs.specs.design_reference.dark_sections.has_dark_sections}}
{{end}}

{{if .site_specs.specs.design_reference.llm_description}}### Design Character (from analysis)
{{if .site_specs.specs.design_reference.llm_description.visual_tone}}Visual tone: {{.site_specs.specs.design_reference.llm_description.visual_tone}}{{end}}
{{if .site_specs.specs.design_reference.llm_description.palette}}Palette description: {{.site_specs.specs.design_reference.llm_description.palette}}{{end}}
{{if .site_specs.specs.design_reference.llm_description.typography}}Typography: {{.site_specs.specs.design_reference.llm_description.typography}}{{end}}
{{end}}

IMPORTANT: Without a design_intent spec, your job is to reproduce the original design faithfully. Use the reference values above as-is. Do not override them with generic industry defaults.

{{else}}
Using the industry, business description, tone, target audience, and services above,
choose colors and typography that are appropriate and distinctive for this specific
industry and brand. Do NOT use generic defaults. A fuel distribution company should
look different from a consulting firm. A veterinary practice should feel different
from a law firm.
{{end}}

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
    ),
    updated_at = now()
WHERE type = 'webdesign-agent';

-- Verify
SELECT
    LENGTH(default_config->'workflow'->'steps'->'analyze_design'->'config'->>'prompt_template') as prompt_length,
    CASE
        WHEN default_config->'workflow'->'steps'->'analyze_design'->'config'->>'prompt_template'
             LIKE '%design_intent%' THEN 'HAS_INTENT'
        ELSE 'NO_INTENT'
    END as has_intent,
    CASE
        WHEN default_config->'workflow'->'steps'->'analyze_design'->'config'->>'prompt_template'
             LIKE '%design_reference%' THEN 'HAS_REFERENCE'
        ELSE 'NO_REFERENCE'
    END as has_reference
FROM agent_definitions
WHERE type = 'webdesign-agent';

--  fix for above

-- ============================================================================
-- Phase 2e: Auto-generate design_intent from design_reference
-- ============================================================================
-- Adds two steps to the adoption workflow after apply_plan:
--   generate_design_intent (LLM) → produces rich semantic design_intent
--   write_design_intent (write_site_spec) → persists to site_specs
--
-- This unlocks prompt path 1 in the webdesign-agent — creative freedom
-- within the described character, rather than locked reproduction of
-- reference values.
--
-- Current flow:  ... → apply_plan → complete
-- New flow:      ... → apply_plan → generate_design_intent
--                                  → write_design_intent → complete
-- ============================================================================

-- Step 1: Add generate_design_intent step
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,generate_design_intent}',
    $$
    {
        "action": "execute_llm_prompt",
        "config": {
            "ai_service": {
                "model": "claude-sonnet-4-6",
                "provider": "anthropic",
                "api_key_env_var": "ANTHROPIC_API_KEY"
            },
            "max_tokens": 4000,
            "temperature": 0.3,
            "input_fields": ["site_record", "design_fingerprint", "adoption_analysis"],
            "prompt_template": "You are a senior brand designer writing a design brief for a web designer. You have concrete CSS data extracted from an existing site and you need to describe what the design IS — its character, its intent, its visual personality — so that a designer can reproduce and eventually evolve it.\n\nDomain: {{if .site_record}}{{.site_record.domain}}{{end}}\n\n== EXTRACTED DESIGN DATA ==\n{{if .design_fingerprint}}{{if .design_fingerprint.suggested_mapping}}Suggested CSS mapping:\n{{range $key, $value := .design_fingerprint.suggested_mapping}}  {{$key}}: {{$value}}\n{{end}}{{end}}\n{{if .design_fingerprint.css_variables}}Original CSS variables:\n{{range $key, $value := .design_fingerprint.css_variables}}  {{$key}}: {{$value}}\n{{end}}{{end}}\n{{if .design_fingerprint.typography}}Typography:\n{{range $key, $value := .design_fingerprint.typography}}  {{$key}}: {{$value}}\n{{end}}{{end}}\n{{if .design_fingerprint.dark_sections}}Dark sections: predominant scheme is {{.design_fingerprint.dark_sections.predominant_scheme}}{{end}}{{end}}\n\n== SITE IDENTITY ==\n{{if .adoption_analysis}}{{if .adoption_analysis.identity}}Company: {{.adoption_analysis.identity.company_name}}\nIndustry: {{.adoption_analysis.identity.industry}}\nTone: {{.adoption_analysis.identity.tone}}\nAudience: {{.adoption_analysis.identity.target_audience}}{{end}}\n{{if .adoption_analysis.design}}LLM design description: {{.adoption_analysis.design.visual_tone}}{{end}}{{end}}\n\nProduce a design intent specification that describes the character of this design — not just the values, but WHY those values work and what they achieve. A designer reading this should understand the visual personality well enough to make good decisions about things not explicitly specified.\n\nRespond with ONLY valid JSON:\n{\n  \"source\": \"design_reference\",\n  \"palette\": {\n    \"character\": \"A detailed description of the colour approach — what mood it creates, what it communicates about the brand, why these specific colours work for this industry and audience\",\n    \"reference_values\": {\n      \"primary\": \"#hex from the extracted data\",\n      \"secondary\": \"#hex\",\n      \"accent\": \"#hex\",\n      \"background\": \"#hex\",\n      \"surface\": \"#hex\",\n      \"text\": \"#hex\",\n      \"text_muted\": \"#hex\"\n    },\n    \"guidance\": \"What to preserve about the palette and what constraints to respect when evolving it\"\n  },\n  \"typography\": {\n    \"character\": \"A description of what the font choices communicate — why these fonts suit this type of site and audience\",\n    \"reference_values\": {\n      \"font_family\": \"the extracted font stack\",\n      \"heading_font\": \"heading font if different\"\n    },\n    \"guidance\": \"What to preserve about the typography when evolving\"\n  },\n  \"spacing\": {\n    \"character\": \"A description of the spacing approach — dense or spacious, why that suits the content type\",\n    \"reference_values\": {\n      \"section_padding\": \"extracted or inferred\",\n      \"container_max_width\": \"extracted or inferred\"\n    }\n  },\n  \"dark_light\": {\n    \"scheme\": \"dark|light|mixed\",\n    \"rationale\": \"Why this scheme works for this site\"\n  },\n  \"overall_character\": \"A 2-3 sentence summary of the entire visual identity that captures its essence\"\n}\n\nRules:\n- The reference_values MUST come from the extracted design data above, not invented\n- The character descriptions should explain WHY the values work, not just restate them\n- The guidance fields should help a designer know what to preserve vs what can evolve\n- If extracted data is missing for a field, use a reasonable inference and note it"
        },
        "next_step": "write_design_intent",
        "error_step": "complete",
        "description": "LLM generates rich semantic design_intent from extracted fingerprint data",
        "output_field": "design_intent_generated"
    }
    $$::jsonb
),
updated_at = now()
WHERE type = 'site-adoption-agent';

-- Step 2: Add write_design_intent step
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,write_design_intent}',
    $$
    {
        "action": "write_site_spec",
        "config": {
            "aspect": "design_intent",
            "source": "site-adoption-agent",
            "site_id": "site_record.site_id",
            "spec_data": "design_intent_generated",
            "source_agent": "site-adoption-agent"
        },
        "next_step": "complete",
        "error_step": "complete",
        "description": "Write design_intent spec from LLM-generated design brief",
        "output_field": "design_intent_written"
    }
    $$::jsonb
),
updated_at = now()
WHERE type = 'site-adoption-agent';

-- Step 3: Re-point apply_plan to go to generate_design_intent instead of complete
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,apply_plan,next_step}',
    '"generate_design_intent"'::jsonb
),
updated_at = now()
WHERE type = 'site-adoption-agent';

-- Step 4: Update complete step to include new output fields
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,complete,config,output_fields}',
    '["site_record", "formatted_crawl", "adoption_analysis", "adoption_result", "design_fingerprint", "design_intent_generated"]'::jsonb
),
updated_at = now()
WHERE type = 'site-adoption-agent';

-- Verify the new flow
SELECT
    default_config->'workflow'->'steps'->'apply_plan'->>'next_step' as apply_plan_goes_to,
    default_config->'workflow'->'steps'->'generate_design_intent'->>'next_step' as gen_intent_goes_to,
    default_config->'workflow'->'steps'->'write_design_intent'->>'next_step' as write_intent_goes_to,
    default_config->'workflow'->'steps'->'generate_design_intent'->>'action' as gen_intent_action,
    default_config->'workflow'->'steps'->'write_design_intent'->>'action' as write_intent_action
FROM agent_definitions
WHERE type = 'site-adoption-agent';
-- Expected: generate_design_intent, write_design_intent, complete, execute_llm_prompt, write_site_spec

---
      -- richer prompts

      -- ============================================================================
-- Phase 2b: Update webdesign-agent analyze_design prompt
-- ============================================================================
-- Uses $$prompt$$ dollar-quoting to avoid single-quote escaping issues.
-- Three-way priority: design_intent → design_reference → generate from industry
-- Richer system prompt that establishes design expertise and constraints.
-- ============================================================================

-- ============================================================================
-- Phase 2b: Update webdesign-agent analyze_design prompt
-- ============================================================================
-- Uses DO block with nested dollar-quoting ($do$ outer, $prompt$ inner)
-- to avoid all escaping issues with long multi-line prompts.
-- ============================================================================

DO $do$
DECLARE
    new_prompt text;
BEGIN
    new_prompt := $prompt$You are a senior web designer specialising in brand-appropriate visual systems. You produce CSS design specifications that express a company's identity, industry, and audience through colour, typography, and spacing. Your designs are distinctive — you never fall back to generic blue-and-grey palettes. You understand that a game design utility platform should feel completely different from a veterinary practice, which should feel completely different from a fuel distribution company. Every design decision you make should be traceable to something specific about the business, its audience, or its industry.

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
{{range .site_specs.specs.identity.services}}- {{if .name}}{{.name}}: {{.description}}{{else}}{{.}}{{end}}
{{end}}{{else if .site_context.services}}## Services Offered
{{range .site_context.services}}- {{if .name}}{{.name}}: {{.description}}{{else}}{{.}}{{end}}
{{end}}{{end}}

{{if .site_specs.specs.identity.unique_selling_points}}## Unique Selling Points
{{range .site_specs.specs.identity.unique_selling_points}}- {{.}}
{{end}}{{end}}

{{if .site_specs.specs.strategy.content_strategy}}## Content Strategy
{{.site_specs.specs.strategy.content_strategy}}{{end}}

## Components Used
{{range .site_context.all_component_functions}}- {{.}}
{{end}}

{{if .site_specs.specs.design_intent}}## Design Intent

A design direction has been set for this site. You have creative freedom to interpret this intent. The reference values are starting points — you may adjust them to better express the described character. Explain your choices in design_notes.

{{if .site_specs.specs.design_intent.palette}}### Palette
{{if .site_specs.specs.design_intent.palette.character}}Character: {{.site_specs.specs.design_intent.palette.character}}{{end}}
{{if .site_specs.specs.design_intent.palette.reference_values}}Reference values (starting points, not exact targets):
{{range $key, $value := .site_specs.specs.design_intent.palette.reference_values}}  --color-{{$key}}: {{$value}}
{{end}}{{end}}
{{if .site_specs.specs.design_intent.palette.guidance}}Guidance: {{.site_specs.specs.design_intent.palette.guidance}}{{end}}
{{end}}

{{if .site_specs.specs.design_intent.typography}}### Typography
{{if .site_specs.specs.design_intent.typography.character}}Character: {{.site_specs.specs.design_intent.typography.character}}{{end}}
{{if .site_specs.specs.design_intent.typography.reference_values}}Reference values:
{{range $key, $value := .site_specs.specs.design_intent.typography.reference_values}}  {{$key}}: {{$value}}
{{end}}{{end}}
{{if .site_specs.specs.design_intent.typography.guidance}}Guidance: {{.site_specs.specs.design_intent.typography.guidance}}{{end}}
{{end}}

{{if .site_specs.specs.design_intent.spacing}}### Spacing
{{if .site_specs.specs.design_intent.spacing.character}}Character: {{.site_specs.specs.design_intent.spacing.character}}{{end}}
{{if .site_specs.specs.design_intent.spacing.reference_values}}Reference values:
{{range $key, $value := .site_specs.specs.design_intent.spacing.reference_values}}  {{$key}}: {{$value}}
{{end}}{{end}}
{{end}}

{{else if .site_specs.specs.design_reference}}## Design Reference (Adopted Site — No Design Intent Set)

This site was adopted from an existing design. No design direction has been set yet, so your job is to faithfully reproduce the original visual identity using our CSS variable conventions. Do NOT invent a new palette or font stack — use the reference values directly.

{{if .site_specs.specs.design_reference.suggested_mapping}}### Reference Values (mapped to our CSS variables)
Use these values directly:
{{range $key, $value := .site_specs.specs.design_reference.suggested_mapping}}  --color-{{$key}}: {{$value}}
{{end}}{{end}}

{{if .site_specs.specs.design_reference.css_variables}}### Original CSS Variables
The original site defined these custom properties:
{{range $key, $value := .site_specs.specs.design_reference.css_variables}}  {{$key}}: {{$value}}
{{end}}{{end}}

{{if .site_specs.specs.design_reference.dark_sections}}### Dark/Light Scheme
Predominant scheme: {{.site_specs.specs.design_reference.dark_sections.predominant_scheme}}
Has dark sections: {{.site_specs.specs.design_reference.dark_sections.has_dark_sections}}
{{end}}

{{if .site_specs.specs.design_reference.llm_description}}### Design Character (from analysis)
{{if .site_specs.specs.design_reference.llm_description.visual_tone}}Visual tone: {{.site_specs.specs.design_reference.llm_description.visual_tone}}{{end}}
{{if .site_specs.specs.design_reference.llm_description.palette}}Palette: {{.site_specs.specs.design_reference.llm_description.palette}}{{end}}
{{if .site_specs.specs.design_reference.llm_description.typography}}Typography: {{.site_specs.specs.design_reference.llm_description.typography}}{{end}}
{{end}}

{{else}}
## Design Direction

No design reference or intent exists for this site. Using the industry, business description, tone, target audience, and services above, design a colour palette and typography system that is appropriate and distinctive for this specific business. Think about what colours and fonts the target audience would expect and trust. Think about what competitors in this industry typically use — then find a way to be distinctive without being inappropriate.

Do NOT use generic defaults. Do NOT default to #2c3e50 or #3498db.
{{end}}

Return ONLY valid JSON:
{
  "color_scheme": {
    "primary": "#hex",
    "secondary": "#hex",
    "accent": "#hex",
    "background": "#hex",
    "surface": "#hex",
    "text": "#hex",
    "text_muted": "#hex",
    "border": "#hex"
  },
  "typography": {
    "font_family": "appropriate font stack",
    "heading_font": "inherit or specific heading font",
    "base_size": "16px",
    "line_height": "1.6"
  },
  "spacing": {
    "section_padding": "appropriate section padding",
    "container_max_width": "appropriate max width"
  },
  "design_notes": "explain specifically why each colour and font choice suits THIS business and audience — generic explanations are not acceptable"
}$prompt$;

    UPDATE agent_definitions
    SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_design,config,prompt_template}',
        to_jsonb(new_prompt)
    ),
    updated_at = now()
    WHERE type = 'webdesign-agent';

    RAISE NOTICE 'Prompt updated, length: %', length(new_prompt);
END $do$;

-- Verify
SELECT
    LENGTH(default_config->'workflow'->'steps'->'analyze_design'->'config'->>'prompt_template') as prompt_length,
    CASE
        WHEN default_config->'workflow'->'steps'->'analyze_design'->'config'->>'prompt_template'
             LIKE '%design_intent%' THEN 'HAS_INTENT'
        ELSE 'NO_INTENT'
    END as has_intent,
    CASE
        WHEN default_config->'workflow'->'steps'->'analyze_design'->'config'->>'prompt_template'
             LIKE '%design_reference%' THEN 'HAS_REFERENCE'
        ELSE 'NO_REFERENCE'
    END as has_reference,
    CASE
        WHEN default_config->'workflow'->'steps'->'analyze_design'->'config'->>'prompt_template'
             LIKE '%senior web designer%' THEN 'UPDATED'
        ELSE 'OLD_PROMPT'
    END as prompt_version
FROM agent_definitions
WHERE type = 'webdesign-agent';


--
      -- Back up the webdesign-agent row into a dated snapshot table
CREATE TABLE IF NOT EXISTS agent_def_webdesign_backup_20260416 AS
SELECT * FROM agent_definitions
WHERE type = 'webdesign-agent' AND deleted_at IS NULL;

-- Apply the workflow changes: add check_should_fork and fork_theme steps,
-- re-point update_site.next_step, and re-point check_update_db.else_step
UPDATE agent_definitions
SET default_config = jsonb_set(
    jsonb_set(
        jsonb_set(
            jsonb_set(
                default_config,
                '{workflow,steps,check_should_fork}',
                '{
                    "action": "conditional",
                    "config": {
                        "condition": "input_data.should_fork_theme == true",
                        "then_step": "fork_theme",
                        "else_step": "complete"
                    },
                    "description": "Check if this design run should be forked into the theme library"
                }'::jsonb
            ),
            '{workflow,steps,fork_theme}',
            '{
                "action": "fork_theme_from_site",
                "config": {
                    "site_id_field": "site_context.site_id",
                    "domain_field": "site_context.domain",
                    "design_spec_field": "design_spec.result",
                    "rendered_css_field": "generated_css.result",
                    "current_collection_id_field": "site_context.style_collection_id"
                },
                "next_step": "complete",
                "error_step": "complete",
                "description": "Fork adopted site theme into reusable library",
                "output_field": "fork_result"
            }'::jsonb
        ),
        '{workflow,steps,update_site,next_step}',
        '"check_should_fork"'::jsonb
    ),
    '{workflow,steps,check_update_db,config,else_step}',
    '"check_should_fork"'::jsonb
),
updated_at = NOW()
WHERE type = 'webdesign-agent' AND deleted_at IS NULL;

-- Verify the four changes landed
SELECT
    default_config -> 'workflow' -> 'steps' -> 'update_site' -> 'next_step' AS update_site_next,
    default_config -> 'workflow' -> 'steps' -> 'check_should_fork' IS NOT NULL AS has_check_should_fork,
    default_config -> 'workflow' -> 'steps' -> 'fork_theme' IS NOT NULL AS has_fork_theme,
    default_config -> 'workflow' -> 'steps' -> 'check_update_db' -> 'config' -> 'else_step' AS check_update_db_else
FROM agent_definitions
WHERE type = 'webdesign-agent' AND deleted_at IS NULL;

---
      -- add fork for css theme

-- Patch: webdesign-agent workflow — route to install_theme when site has no collection
--
-- Adds two workflow steps to the existing webdesign-agent:
--   1. `check_should_install` conditional: if the site has no style_collection_id,
--      route to `install_theme`; otherwise go straight to `complete`.
--   2. `install_theme`: calls `fork_theme_from_site` with `install_on_site: true`,
--      which inserts css_themes + style_collections and sets sites.style_collection_id.
--
-- Also changes the `check_should_fork.else_step` from "complete" to
-- "check_should_install" so sites without explicit fork requests still get
-- their theme installed.
--
-- Existing behaviour preserved:
--   - `fork_theme` (library contribution) still runs when input_data.should_fork_theme == true
--   - Direct API callers who neither fork nor install still complete cleanly
--     (they will hit install_theme only if the site has no collection — which is
--      the correct behaviour: a site with no theme should get one)
--
-- Prerequisite:
--   Go patch to fork_theme_from_site_action.go adding `install_on_site` flag must
--   be DEPLOYED before this SQL is applied. Without the flag, the install_theme
--   step would run in library-contribution mode and fail to link the site.

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
    jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,check_should_fork,config,else_step}',
            '"check_should_install"'::jsonb
        ),
        '{workflow,steps,check_should_install}',
        '{
            "action": "conditional",
            "config": {
                "condition": "site_context.style_collection_id == null",
                "then_step": "install_theme",
                "else_step": "complete"
            },
            "description": "Install generated theme on site if no collection is linked yet"
        }'::jsonb
    ),
    '{workflow,steps,install_theme}',
    '{
        "action": "fork_theme_from_site",
        "config": {
            "domain_field": "site_context.domain",
            "site_id_field": "site_context.site_id",
            "design_spec_field": "design_spec.result",
            "rendered_css_field": "generated_css.result",
            "current_collection_id_field": "site_context.style_collection_id",
            "install_on_site": true
        },
        "next_step": "complete",
        "error_step": "complete",
        "description": "Persist theme + collection and link to site (sets sites.style_collection_id)",
        "output_field": "install_result"
    }'::jsonb
)
WHERE type = 'webdesign-agent'
  AND is_active = true;

-- Also extend the complete step's output_fields so the install result is
-- surfaced to callers (e.g. the work item result jsonb).
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,complete,config,output_fields}',
    '["design_spec", "css_deployed", "site_context", "install_result", "fork_result"]'::jsonb
)
WHERE type = 'webdesign-agent'
  AND is_active = true;

-- Verify structure
SELECT
    default_config #>> '{workflow,steps,check_should_fork,config,else_step}'        AS fork_else,
    default_config #>> '{workflow,steps,check_should_install,action}'               AS install_cond_action,
    default_config #>> '{workflow,steps,check_should_install,config,condition}'     AS install_condition,
    default_config #>> '{workflow,steps,check_should_install,config,then_step}'     AS install_then,
    default_config #>> '{workflow,steps,install_theme,action}'                      AS install_action,
    default_config #>> '{workflow,steps,install_theme,config,install_on_site}'      AS install_flag,
    default_config #>> '{workflow,steps,install_theme,next_step}'                   AS install_next,
    default_config #>> '{workflow,steps,complete,config,output_fields}'             AS complete_outputs
FROM agent_definitions
WHERE type = 'webdesign-agent' AND is_active = true;

COMMIT;


--

      -- fork theme
      -- Fix: install_theme and fork_theme step configs use wrong key names
--
-- ForkThemeFromSiteAction calls resolveConfigString(config, "site_id", ...) and
-- resolveConfigString(config, "domain", ...). These look up literal keys
-- "site_id" and "domain" in the step config. Our step configs were using
-- "site_id_field" and "domain_field" (matching the ...Field naming convention
-- used for design_spec_field, rendered_css_field etc), so the lookups returned
-- empty and the site_id fell through to the site_record.site_id fallback,
-- which doesn't exist in the webdesign-agent's collected_data (it has
-- site_context, not site_record). Result: "invalid site_id" skip.
--
-- This patch renames the keys. The VALUES (e.g. "site_context.site_id") are
-- unchanged — resolveConfigString treats dotted values as paths to resolve
-- against collected_data.
--
-- Why fork_theme (the library path) didn't hit this: its gate
-- input_data.should_fork_theme == true has never been triggered in practice,
-- so the broken path wasn't exercised. Fixing it now so the gate works when
-- someone does flip it.

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
    jsonb_set(
        default_config,
        '{workflow,steps,install_theme,config}',
        '{
            "domain": "site_context.domain",
            "site_id": "site_context.site_id",
            "design_spec_field": "design_spec.result",
            "rendered_css_field": "generated_css.result",
            "current_collection_id_field": "site_context.style_collection_id",
            "install_on_site": true
        }'::jsonb
    ),
    '{workflow,steps,fork_theme,config}',
    '{
        "domain": "site_context.domain",
        "site_id": "site_context.site_id",
        "design_spec_field": "design_spec.result",
        "rendered_css_field": "generated_css.result",
        "current_collection_id_field": "site_context.style_collection_id"
    }'::jsonb
)
WHERE type = 'webdesign-agent'
  AND is_active = true;

-- Verify
SELECT
    jsonb_pretty(default_config #> '{workflow,steps,install_theme,config}') AS install_theme_config,
    jsonb_pretty(default_config #> '{workflow,steps,fork_theme,config}')    AS fork_theme_config
FROM agent_definitions
WHERE type = 'webdesign-agent' AND is_active = true;

COMMIT;

      