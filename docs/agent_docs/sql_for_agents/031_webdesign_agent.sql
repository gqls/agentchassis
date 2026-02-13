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