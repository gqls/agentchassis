-- ============================================================================
-- NEW: html-developer-chunked - Generates HTML in manageable chunks
-- Handles enormous sites by breaking generation into smaller LLM calls
-- ============================================================================

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
    version
) VALUES (
             'html-developer-chunked',
             'Chunked HTML Developer',
             'Generates HTML in smaller chunks to handle large sites without token limits',
             'code-driven',
             '{
                 "workflow": {
                     "start_step": "generate_structure",
                     "steps": {
                         "generate_structure": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-haiku-4-5-20251001",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 4000
                                 },
                                 "input_fields": ["input_data", "site_architecture"],
                                 "prompt_template": "Generate the HTML structure (DOCTYPE, head, meta tags, basic body structure) for a website.\n\nDomain: {{.input_data.domain}}\nArchitecture: {{.site_architecture.architecture_result.result}}\n\nReturn ONLY:\n<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n  <!-- meta tags, title, CSS framework here -->\n</head>\n<body>\n  <!-- Empty body with proper semantic structure sections commented -->\n</body>\n</html>"
                             },
                             "output_field": "base_structure",
                             "next_step": "generate_styles",
                             "description": "Generate base HTML structure"
                         },
                         "generate_styles": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-haiku-4-5-20251001",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 8000
                                 },
                                 "input_fields": ["input_data", "site_architecture"],
                                 "prompt_template": "Generate complete CSS for the website.\n\nArchitecture: {{.site_architecture.architecture_result.result}}\n\nReturn ONLY the <style> tag with complete CSS:\n- Modern, clean design\n- Mobile-first responsive\n- Color scheme from architecture\n- Typography from architecture\n\nReturn: <style>/* your CSS */</style>"
                             },
                             "output_field": "styles",
                             "next_step": "generate_content_sections",
                             "description": "Generate CSS styles"
                         },
                         "generate_content_sections": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5-20250514",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 12000
                                 },
                                 "input_fields": ["site_content"],
                                 "prompt_template": "Generate HTML for ALL content sections.\n\nContent: {{.site_content.content_result.result}}\n\nReturn ONLY the HTML body content (all sections):\n<header>...</header>\n<main>\n  <section class=\"hero\">...</section>\n  <section class=\"features\">...</section>\n  <!-- all other sections -->\n</main>\n<footer>...</footer>"
                             },
                             "output_field": "content_html",
                             "next_step": "assemble_html",
                             "description": "Generate content sections"
                         },
                         "assemble_html": {
                             "action": "assemble_html_parts",
                             "config": {
                                 "structure_field": "base_structure.result",
                                 "styles_field": "styles.result",
                                 "content_field": "content_html.result"
                             },
                             "output_field": "final_html",
                             "next_step": "complete",
                             "description": "Assemble all parts into complete HTML"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Return complete HTML"
                         }
                     }
                 },
                 "storage": {
                     "type": "s3",
                     "enabled": true,
                     "auto_store": true,
                     "bucket_env": "ASSETS_BUCKET",
                     "public_access": true
                 },
                 "processing_mode": "task",
                 "timeout_seconds": 300
             }'::jsonb,
             true,
             '["html", "css", "chunked-generation", "frontend"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.508',
             '{"limits": {"cpu": "800m", "memory": "1Gi"}, "requests": {"cpu": "150m", "memory": "384Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             1
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       updated_at = now();

-- Note: The assemble_html_parts action needs to be implemented
-- For now, let's use a simpler approach with just better chunking

-- ============================================================================
-- NEW: html-developer-chunked - Generates HTML in manageable chunks
-- Handles enormous sites by breaking generation into smaller LLM calls
-- ============================================================================
--
-- PREREQUISITES:
-- 1. The assemble_html_parts action must be implemented in the codebase
-- 2. Action must be registered in the action registry
-- 3. See html_assembly_actions.go and assemble_html_parts_documentation.md
--
-- DEPLOYMENT:
-- 1. Add html_assembly_actions.go to platform/orchestration/actions/
-- 2. Register action in action_registry.go
-- 3. Rebuild agent-chassis with new action
-- 4. Update image_tag below to your new version
-- 5. Run this SQL to create the agent
-- ============================================================================

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
    version
) VALUES (
             'html-developer-chunked',
             'Chunked HTML Developer',
             'Generates HTML in smaller chunks to handle large sites without token limits',
             'code-driven',
             '{
                 "workflow": {
                     "start_step": "generate_structure",
                     "steps": {
                         "generate_structure": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-haiku-4-5-20251001",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 4000
                                 },
                                 "input_fields": ["input_data", "site_architecture"],
                                 "prompt_template": "Generate the HTML structure skeleton for a website.\n\nDomain: {{.input_data.domain}}\n\nCreate a basic HTML5 structure with:\n- DOCTYPE declaration\n- HTML with lang attribute\n- Head with charset, viewport, title\n- Empty body with semantic structure comments\n\nIMPORTANT: Include a comment placeholder <!-- CONTENT_HERE --> in the body.\n\nReturn ONLY the HTML code, no explanations."
                             },
                             "output_field": "base_structure",
                             "next_step": "generate_styles",
                             "description": "Generate base HTML structure"
                         },
                         "generate_styles": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-haiku-4-5-20251001",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 8000
                                 },
                                 "input_fields": ["input_data", "site_architecture"],
                                 "prompt_template": "Generate complete CSS for a website.\n\nDomain: {{.input_data.domain}}\n\nCreate modern, responsive CSS with:\n- CSS reset and box-sizing\n- Mobile-first responsive design\n- Modern font stack\n- Clean, professional styling\n- Utility classes\n\nReturn ONLY CSS wrapped in <style> tags."
                             },
                             "output_field": "styles",
                             "next_step": "generate_content_sections",
                             "description": "Generate CSS styles"
                         },
                         "generate_content_sections": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "ai_service": {
                                     "model": "claude-sonnet-4-5-20250514",
                                     "provider": "anthropic",
                                     "api_key_env_var": "ANTHROPIC_API_KEY",
                                     "max_tokens": 12000
                                 },
                                 "input_fields": ["input_data", "site_content"],
                                 "prompt_template": "Generate HTML body content for a website.\n\nDomain: {{.input_data.domain}}\n\nCreate semantic HTML5 content including:\n- Header with navigation\n- Main content area with sections\n- Footer\n\nUse the provided site content to populate the sections.\n\nReturn ONLY the HTML body content (no <html>, <head>, or <body> tags)."
                             },
                             "output_field": "content_html",
                             "next_step": "assemble_html",
                             "description": "Generate content sections"
                         },
                         "assemble_html": {
                             "action": "assemble_html_parts",
                             "config": {
                                 "structure_field": "base_structure.result",
                                 "styles_field": "styles.result",
                                 "content_field": "content_html.result"
                             },
                             "output_field": "html_result",
                             "next_step": "complete",
                             "description": "Assemble all parts into complete HTML"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Return complete HTML"
                         }
                     }
                 },
                 "storage": {
                     "type": "s3",
                     "enabled": true,
                     "auto_store": true,
                     "bucket_env": "ASSETS_BUCKET",
                     "public_access": true
                 },
                 "processing_mode": "task",
                 "timeout_seconds": 300
             }'::jsonb,
             true,
             '["html", "css", "chunked-generation", "frontend"]'::jsonb,
             'docker.io/aqls/agent-chassis',
             'v1.0.509',  -- UPDATE THIS to your new image version with assemble_html_parts action
             '{"limits": {"cpu": "800m", "memory": "1Gi"}, "requests": {"cpu": "150m", "memory": "384Mi"}}'::jsonb,
             '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
             '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
             1
         )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       image_tag = EXCLUDED.image_tag,
                                       updated_at = now();

---

-- ============================================================================
-- Add output_type to Agent Configs
-- This tells ai_actions.go whether to append JSON output instructions
-- ============================================================================

-- 7. HTML DEVELOPER CHUNKED - Multiple steps, add to relevant ones
-- Structure step doesn't need JSON instructions
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_structure,config,output_type}',
        '"html"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'html-developer-chunked'
  AND is_active = true;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_styles,config,output_type}',
        '"html"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'html-developer-chunked'
  AND is_active = true;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_content_sections,config,output_type}',
        '"html"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'html-developer-chunked'
  AND is_active = true;


