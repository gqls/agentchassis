-- 009_seed_website_builder_definitions.sql
-- Seed additional agent definitions that weren't included in the initial setup
-- Note: Some of these update existing definitions from 005_website_builder_agents.sql

-- 1. Update Domain Analyst with complete workflow
UPDATE agent_definitions
SET default_config = '{
    "model": "claude-3-5-sonnet-20241022",
    "temperature": 0.3,
    "processing_mode": "task",
    "ai_service": {
        "provider": "anthropic",
        "model": "claude-3-5-sonnet-20241022"
    },
    "prompt_template": "You are a business domain expert. Analyze the following business name: ''{{.input.business_name}}'' and domain: ''{{.input.domain}}''. Based on this information, provide a structured JSON output with the business ''type'' (e.g., ''e-commerce'', ''portfolio'', ''blog'') and a list of 5 relevant SEO ''keywords''.",
    "workflow": {
        "start_step": "analyze",
        "steps": {
            "analyze": {
                "action": "execute_llm_prompt",
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'domain-analyst';

-- 2. Update Site Architect with complete workflow
UPDATE agent_definitions
SET default_config = '{
    "model": "claude-3-5-sonnet-20241022",
    "temperature": 0.5,
    "processing_mode": "task",
    "ai_service": {
        "provider": "anthropic",
        "model": "claude-3-5-sonnet-20241022"
    },
    "prompt_template": "You are a website architect. For a ''{{.input.business_type}}'' business named ''{{.input.business_name}}'', design a complete site structure. Provide a structured JSON output with a ''pages'' array. Each object in the array should have a ''name'' (e.g., ''Home'', ''About Us'') and a brief ''description'' of its purpose.",
    "workflow": {
        "start_step": "design_structure",
        "steps": {
            "design_structure": {
                "action": "execute_llm_prompt",
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'site-architect';

-- 3. Update HTML Developer with complete workflow
UPDATE agent_definitions
SET default_config = '{
    "model": "claude-3-5-sonnet-20241022",
    "temperature": 0.2,
    "processing_mode": "task",
    "ai_service": {
        "provider": "anthropic",
        "model": "claude-3-5-sonnet-20241022"
    },
    "prompt_template": "You are an expert front-end web developer. Generate a single, complete, and responsive HTML file (including inline CSS in a <style> tag) for the ''{{.input.page.name}}'' page of a website for ''{{.input.business_name}}''. The page''s purpose is: {{.input.page.description}}. Use the following content: {{.input.page.content}}.",
    "workflow": {
        "start_step": "generate_code",
        "steps": {
            "generate_code": {
                "action": "execute_llm_prompt",
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow"
            }
        }
    }
}'::jsonb,
updated_at = NOW()
WHERE type = 'html-developer';

-- 4. Content Creator Agent (if not already exists)
INSERT INTO agent_definitions (type, display_name, description, category, default_config, capabilities) VALUES
    ('content-creator', 'Content Creator', 'Advanced AI-powered content generation with memory and style adaptation', 'data-driven',
     '{
         "model": "claude-3-5-sonnet-20241022",
         "temperature": 0.7,
         "max_tokens": 2000,
         "processing_mode": "task",
         "memory_config": {
             "enabled": true,
             "auto_store": true,
             "auto_store_threshold": 0.7,
             "max_memories": 100,
             "retrieval_count": 5,
             "embedding_model": "text-embedding-ada-002",
             "include_types": ["generated_content", "user_feedback", "style_preferences"]
         },
         "metrics_config": {
             "enabled": true,
             "fail_silently": true,
             "detailed_errors": false,
             "record_token_usage": true,
             "record_latency": true,
             "record_errors": true
         },
         "workflow": {
             "start_step": "generate_content",
             "steps": {
                 "generate_content": {
                     "action": "ai_text_generate_anthropic",
                     "description": "Generate text content using Anthropic LLM with memory context",
                     "store_memory": true,
                     "next_step": "complete_workflow"
                 },
                 "complete_workflow": {
                     "action": "complete_workflow",
                     "description": "Mark workflow as complete and store results"
                 }
             }
         },
         "supported_content_types": [
             "blog_post", "product_description", "social_media",
             "email", "landing_page", "press_release", "technical_doc"
         ],
         "style_options": ["informative", "persuasive", "casual", "professional", "creative"],
         "tone_options": ["friendly", "formal", "conversational", "authoritative", "enthusiastic"],
         "length_options": ["short", "medium", "long"],
         "platform_support": ["generic", "twitter", "linkedin", "facebook", "instagram"]
     }'::jsonb,
     '["content-generation", "writing", "copywriting", "seo", "memory-enabled"]'::jsonb)
    ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              default_config = EXCLUDED.default_config,
                              capabilities = EXCLUDED.capabilities,
                              updated_at = NOW();

-- 5. Generic Orchestrator Agent (if not already exists)
INSERT INTO agent_definitions (type, display_name, description, category, default_config, capabilities) VALUES
    ('generic', 'Generic Orchestrator', 'Generic agent that can spawn groups and orchestrate workflows', 'orchestrator',
     '{
         "processing_mode": "orchestrator",
         "workflow": {
             "start_step": "spawn_website_team",
             "steps": {
                 "spawn_website_team": {
                     "action": "spawn_group",
                     "config": {"group_type": "website-builder"},
                     "next_step": "complete"
                 },
                 "complete": {
                     "action": "complete_workflow"
                 }
             }
         }
     }'::jsonb,
     '["orchestration", "spawn_group", "workflow-management"]'::jsonb)
    ON CONFLICT (type) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              capabilities = EXCLUDED.capabilities,
                              updated_at = NOW();

-- 6. Ensure all agents have proper processing_mode set
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
WHERE NOT (default_config ? 'processing_mode') OR default_config->>'processing_mode' IS NULL;

-- 7. Add topics to orchestrator agents that need them
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,analyze_domain,topic}',
        '"system.agent.domain-analyst.process"'::jsonb
                     )
WHERE type = 'website-builder'
  AND default_config->'workflow'->'steps'->'analyze_domain' IS NOT NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,architect_site,topic}',
        '"system.agent.site-architect.process"'::jsonb
                     )
WHERE type = 'website-builder'
  AND default_config->'workflow'->'steps'->'architect_site' IS NOT NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,gather_content,topic}',
        '"system.agent.content-creator.process"'::jsonb
                     )
WHERE type = 'website-builder'
  AND default_config->'workflow'->'steps'->'gather_content' IS NOT NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,develop_site,topic}',
        '"system.agent.html-developer.process"'::jsonb
                     )
WHERE type = 'website-builder'
  AND default_config->'workflow'->'steps'->'develop_site' IS NOT NULL;

-- Add a summary comment
COMMENT ON TABLE agent_definitions IS 'Core agent definitions with workflows, updated by migration 009';