-- FILE: platform/database/migrations/011_fix_agent_configs.sql
-- Fix agent configurations to ensure they can work in hybrid mode

BEGIN;

-- Ensure all data-driven and code-driven agents have proper AI configuration
UPDATE agent_definitions
SET default_config = default_config || '{
    "processing_mode": "orchestrator",
    "ai_service": {
        "provider": "anthropic",
        "model": "claude-3-5-sonnet-20241022"
    },
    "temperature": 0.3,
    "max_tokens": 2000
}'::jsonb
WHERE category IN ('data-driven', 'code-driven')
  AND (
    default_config->'ai_service' IS NULL
   OR default_config->>'ai_service' = 'null'
   OR default_config->'ai_service'->>'provider' IS NULL
    );

-- Ensure all agents have a prompt_template if they use execute_llm_prompt
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{prompt_template}',
        '"You are a {{.Type}} agent. Process this request: {{.input.message}}"'::jsonb
                     )
WHERE default_config->>'prompt_template' IS NULL
  AND category IN ('data-driven', 'code-driven');

-- Fix domain-analyst specifically
UPDATE agent_definitions
SET default_config = default_config || '{
    "prompt_template": "You are a business domain expert. Analyze the following business name: ''{{.input.business_name}}'' and domain: ''{{.input.domain}}''. Based on this information, provide a structured JSON output with the business ''type'' (e.g., ''e-commerce'', ''portfolio'', ''blog'') and a list of 5 relevant SEO ''keywords''."
}'::jsonb
WHERE type = 'domain-analyst';

-- Fix site-architect
UPDATE agent_definitions
SET default_config = default_config || '{
    "prompt_template": "You are a website architect. For a ''{{.input.business_type}}'' business named ''{{.input.business_name}}'', design a complete site structure. Provide a structured JSON output with a ''pages'' array. Each object in the array should have a ''name'' (e.g., ''Home'', ''About Us'') and a brief ''description'' of its purpose."
}'::jsonb
WHERE type = 'site-architect';

-- Fix html-developer
UPDATE agent_definitions
SET default_config = default_config || '{
    "prompt_template": "Generate a complete HTML page with inline CSS for {{.input.business_name}}. Page: {{.input.page.name}}. Description: {{.input.page.description}}. Content: {{.input.page.content}}"
}'::jsonb
WHERE type = 'html-developer';

-- Fix content-creator
UPDATE agent_definitions
SET default_config = default_config || '{
    "prompt_template": "Create high-quality content for {{.input.business_name}}. Type: {{.input.content_type}}. Topic: {{.input.topic}}. Generate engaging, SEO-optimized content."
}'::jsonb
WHERE type = 'content-creator';

-- Ensure all orchestrators are in orchestrator mode
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{processing_mode}',
        '"orchestrator"'::jsonb
                     )
WHERE category = 'orchestrator';

-- Update timestamps
UPDATE agent_definitions
SET updated_at = NOW()
WHERE updated_at < NOW() - INTERVAL '1 second';

COMMIT;