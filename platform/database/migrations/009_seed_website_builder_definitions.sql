-- 009_seed_website_builder_definitions.sql
-- This script inserts the version 1 definitions for all agents
-- used in the website-builder group.

-- 1. Domain Analyst Agent Definition
INSERT INTO agent_definitions (type, display_name, description, category, default_config) VALUES
    ('domain-analyst', 'Domain Analyst', 'Analyzes a domain to determine business type and suggest keywords.', 'data-driven', '{
    "ai_service": { "provider": "anthropic", "model": "claude-3-5-sonnet-20241022" },
    "prompt_template": "You are a business domain expert. Analyze the following business name: ''{{.input.business_name}}'' and domain: ''{{.input.domain}}''. Based on this information, provide a structured JSON output with the business ''type'' (e.g., ''e-commerce'', ''portfolio'', ''blog'') and a list of 5 relevant SEO ''keywords''.",
    "workflow": {
        "start_step": "analyze",
        "steps": {
            "analyze": { "action": "execute_llm_prompt", "next_step": "complete" },
            "complete": { "action": "complete_workflow" }
        }
    }
}'::jsonb)
    ON CONFLICT (type) DO UPDATE SET default_config = EXCLUDED.default_config, updated_at = NOW();

-- 2. Site Architect Agent Definition
INSERT INTO agent_definitions (type, display_name, description, category, default_config) VALUES
    ('site-architect', 'Site Architect', 'Plans the website structure, pages, and navigation.', 'data-driven', '{
    "ai_service": { "provider": "anthropic", "model": "claude-3-5-sonnet-20241022" },
    "prompt_template": "You are a website architect. For a ''{{.input.business_type}}'' business named ''{{.input.business_name}}'', design a complete site structure. Provide a structured JSON output with a ''pages'' array. Each object in the array should have a ''name'' (e.g., ''Home'', ''About Us'') and a brief ''description'' of its purpose.",
    "workflow": {
        "start_step": "design_structure",
        "steps": {
            "design_structure": { "action": "execute_llm_prompt", "next_step": "complete" },
            "complete": { "action": "complete_workflow" }
        }
    }
}'::jsonb)
    ON CONFLICT (type) DO UPDATE SET default_config = EXCLUDED.default_config, updated_at = NOW();

-- 3. HTML Developer Agent Definition
INSERT INTO agent_definitions (type, display_name, description, category, default_config) VALUES
    ('html-developer', 'HTML Developer', 'Generates HTML, CSS, and JS code for websites.', 'code-driven', '{
    "ai_service": { "provider": "anthropic", "model": "claude-3-5-sonnet-20241022" },
    "prompt_template": "You are an expert front-end web developer. Generate a single, complete, and responsive HTML file (including inline CSS in a <style> tag) for the ''{{.input.page.name}}'' page of a website for ''{{.input.business_name}}''. The page''s purpose is: {{.input.page.description}}. Use the following content: {{.input.page.content}}.",
    "workflow": {
        "start_step": "generate_code",
        "steps": {
            "generate_code": { "action": "execute_llm_prompt", "next_step": "complete" },
            "complete": { "action": "complete_workflow" }
        }
    }
}'::jsonb)
    ON CONFLICT (type) DO UPDATE SET default_config = EXCLUDED.default_config, updated_at = NOW();

-- 4. Visual Designer Agent Definition (Adapter Agent)
INSERT INTO agent_definitions (type, display_name, description, category, default_config) VALUES
    ('visual-designer', 'Visual Designer', 'Orchestrates image generation and selection.', 'adapter', '{
    "workflow": {
        "start_step": "generate_logo",
        "steps": {
            "generate_logo": {
                "action": "call_agent",
                "config": { "agent_type": "image-generator" },
                "next_step": "complete"
            },
            "complete": { "action": "complete_workflow" }
        }
    }
}'::jsonb)
    ON CONFLICT (type) DO UPDATE SET default_config = EXCLUDED.default_config, updated_at = NOW();

-- 5. Site Publisher Agent Definition (Adapter Agent)
INSERT INTO agent_definitions (type, display_name, description, category, default_config) VALUES
    ('site-publisher', 'Site Publisher', 'Publishes website files to a storage bucket.', 'adapter', '{
    "workflow": {
        "start_step": "upload_files",
        "steps": {
            "upload_files": {
                "action": "s3_upload",
                "next_step": "complete"
            },
            "complete": { "action": "complete_workflow" }
        }
    }
}'::jsonb)
    ON CONFLICT (type) DO UPDATE SET default_config = EXCLUDED.default_config, updated_at = NOW();

-- This would be in the agent_definitions table for the website-builder agent type
INSERT INTO agent_definitions (type, display_name, description, category, default_config) VALUES
    ('website-builder', 'Website Builder Orchestrator', 'Orchestrates complete website creation', 'code-driven', '{
    "workflow": {
        "start_step": "validate_request",
        "steps": {
            "validate_request": {
                "action": "validate_input",
                "next_step": "analyze_domain"
            },
            "analyze_domain": {
                "action": "call_agent",
                "config": {"agent_type": "domain-analyst"},
                "topic": "system.agent.domain-analyst.process",
                "next_step": "architect_site"
            },
            "architect_site": {
                "action": "call_agent",
                "config": {"agent_type": "site-architect"},
                "topic": "system.agent.site-architect.process",
                "next_step": "gather_content"
            },
            "gather_content": {
                "action": "call_agent",
                "config": {"agent_type": "content-creator"},
                "topic": "system.agent.content-creator.process",
                "next_step": "create_visuals"
            },
            "create_visuals": {
                "action": "call_agent",
                "config": {"agent_type": "visual-designer"},
                "topic": "system.agent.visual-designer.process",
                "next_step": "develop_site"
            },
            "develop_site": {
                "action": "call_agent",
                "config": {"agent_type": "html-developer"},
                "topic": "system.agent.html-developer.process",
                "next_step": "publish_site"
            },
            "publish_site": {
                "action": "call_agent",
                "config": {"agent_type": "site-publisher"},
                "topic": "system.agent.site-publisher.process",
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow"
            }
        }
    }
}'::jsonb)
    ON CONFLICT (type) DO UPDATE SET default_config = EXCLUDED.default_config;

-- Just add processing_mode to distinguish orchestrators from workers
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{processing_mode}',
        '"task"'
                     )
WHERE type IN ('domain-analyst', 'site-architect', 'html-developer', 'content-creator');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{processing_mode}',
        '"orchestrator"'
                     )
WHERE type IN ('website-builder', 'visual-designer', 'site-publisher');

-- Ensure ALL agents have processing_mode set
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{processing_mode}',
        CASE
            WHEN category = 'orchestrator' THEN '"orchestrator"'
            WHEN category = 'adapter' THEN '"orchestrator"'
            ELSE '"task"'
            END::jsonb
                     )
WHERE NOT (default_config ? 'processing_mode');