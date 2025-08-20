-- 080_clientsdb_content_creator_agent_definition_with_memory.sql
-- Additional agent definitions with enhanced configurations
-- Note: This complements the definitions in 005 and 009

-- ============================================================================
-- 1. CONTENT CREATOR AGENT WITH MEMORY
-- ============================================================================
-- Enhanced agent definition with memory configuration for advanced content generation
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
             "start_step": "retrieve_context",
             "steps": {
                 "retrieve_context": {
                     "action": "retrieve_memory",
                     "description": "Retrieve relevant memories and context",
                     "config": {
                         "memory_types": ["style_preferences", "user_feedback"],
                         "max_results": 5
                     },
                     "next_step": "generate_content"
                 },
                 "generate_content": {
                     "action": "ai_text_generate_anthropic",
                     "description": "Generate text content using Anthropic LLM with memory context",
                     "store_memory": true,
                     "next_step": "quality_check"
                 },
                 "quality_check": {
                     "action": "validate_output",
                     "description": "Check content quality and compliance",
                     "config": {
                         "check_grammar": true,
                         "check_tone": true,
                         "check_length": true
                     },
                     "next_step": "store_memory"
                 },
                 "store_memory": {
                     "action": "store_memory",
                     "description": "Store generated content in memory",
                     "config": {
                         "memory_type": "generated_content",
                         "include_metadata": true
                     },
                     "next_step": "complete_workflow"
                 },
                 "complete_workflow": {
                     "action": "complete_workflow",
                     "description": "Mark workflow as complete and return results"
                 }
             }
         },
         "supported_content_types": [
             "blog_post", "product_description", "social_media",
             "email", "landing_page", "press_release", "technical_doc",
             "ad_copy", "newsletter", "case_study", "whitepaper"
         ],
         "style_options": ["informative", "persuasive", "casual", "professional", "creative", "technical", "conversational"],
         "tone_options": ["friendly", "formal", "conversational", "authoritative", "enthusiastic", "empathetic", "neutral"],
         "length_options": ["micro", "short", "medium", "long", "comprehensive"],
         "platform_support": ["generic", "twitter", "linkedin", "facebook", "instagram", "medium", "substack"],
         "seo_optimization": {
             "enabled": true,
             "keyword_density": 0.02,
             "meta_description": true,
             "heading_optimization": true
         }
     }'::jsonb,
     '["content-generation", "writing", "copywriting", "seo", "memory-enabled", "style-adaptive"]'::jsonb)
    ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              category = EXCLUDED.category,
                              default_config = EXCLUDED.default_config,
                              capabilities = EXCLUDED.capabilities,
                              updated_at = NOW();

-- ============================================================================
-- 2. GENERIC ORCHESTRATOR AGENT
-- ============================================================================
-- Generic agent that can spawn groups and orchestrate workflows
INSERT INTO agent_definitions (type, display_name, description, category, default_config, capabilities) VALUES
    ('generic', 'Generic Orchestrator', 'Generic agent that can spawn groups and orchestrate workflows dynamically', 'orchestrator',
     '{
         "processing_mode": "orchestrator",
         "max_concurrent_agents": 5,
         "timeout_seconds": 300,
         "retry_config": {
             "max_retries": 3,
             "backoff_multiplier": 2,
             "initial_delay_ms": 1000
         },
         "workflow": {
             "start_step": "analyze_request",
             "steps": {
                 "analyze_request": {
                     "action": "analyze_input",
                     "description": "Analyze the request to determine which group to spawn",
                     "next_step": "determine_group"
                 },
                 "determine_group": {
                     "action": "select_group",
                     "description": "Select the appropriate agent group based on request",
                     "config": {
                         "default_group": "website-builder",
                         "selection_strategy": "capability_match"
                     },
                     "next_step": "spawn_website_team"
                 },
                 "spawn_website_team": {
                     "action": "spawn_group",
                     "description": "Spawn the selected agent group",
                     "config": {
                         "group_type": "website-builder",
                         "pass_context": true,
                         "inherit_config": true
                     },
                     "next_step": "monitor_execution"
                 },
                 "monitor_execution": {
                     "action": "monitor_agents",
                     "description": "Monitor spawned agents execution",
                     "config": {
                         "log_level": "info",
                         "collect_metrics": true
                     },
                     "next_step": "complete"
                 },
                 "complete": {
                     "action": "complete_workflow",
                     "description": "Collect results and complete workflow"
                 }
             }
         },
         "supported_groups": ["website-builder", "content-team", "research-team", "development-team"],
         "fallback_behavior": "graceful_degradation"
     }'::jsonb,
     '["orchestration", "spawn_group", "workflow-management", "dynamic-routing", "monitoring"]'::jsonb)
    ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              category = EXCLUDED.category,
                              default_config = EXCLUDED.default_config,
                              capabilities = EXCLUDED.capabilities,
                              updated_at = NOW();

-- ============================================================================
-- 3. DOMAIN ANALYST BEHAVIOR UPDATE
-- ============================================================================
-- Update domain-analyst with enhanced behavior configuration
UPDATE agent_definitions
SET
    default_config = jsonb_set(
            default_config,
            '{behavior}',
            '{
                "processing_mode": "task",
                "capabilities": ["analyze", "extract", "summarize", "categorize"],
                "analysis_depth": "comprehensive",
                "output_format": "structured_json",
                "workflow": {
                    "start_step": "validate_domain",
                    "steps": {
                        "validate_domain": {
                            "action": "validate_input",
                            "description": "Validate domain format and accessibility",
                            "config": {
                                "check_dns": true,
                                "check_format": true
                            },
                            "next_step": "process_input"
                        },
                        "process_input": {
                            "action": "execute_llm_prompt",
                            "description": "Analyze domain and business information",
                            "next_step": "enrich_data"
                        },
                        "enrich_data": {
                            "action": "call_agent",
                            "description": "Enrich with additional research if needed",
                            "config": {
                                "agent_type": "content-researcher",
                                "optional": true
                            },
                            "next_step": "format_output"
                        },
                        "format_output": {
                            "action": "transform_data",
                            "description": "Format analysis results",
                            "config": {
                                "transformation": "json_structure",
                                "schema": "domain_analysis_v1"
                            },
                            "next_step": "send_response"
                        },
                        "send_response": {
                            "action": "send_notification",
                            "description": "Send analysis results",
                            "config": {
                                "include_confidence_scores": true
                            },
                            "next_step": "complete"
                        },
                        "complete": {
                            "action": "complete_workflow",
                            "description": "Complete the analysis workflow"
                        }
                    }
                }
            }'::jsonb
                     ),
    capabilities = capabilities || '["domain-validation", "business-categorization", "keyword-extraction"]'::jsonb,
    updated_at = NOW()
WHERE type = 'domain-analyst';

-- ============================================================================
-- 4. ENSURE ALL AGENTS HAVE PROCESSING MODE
-- ============================================================================
-- Final consistency check for all agent definitions
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{processing_mode}',
        CASE
            WHEN category = 'orchestrator' THEN '"orchestrator"'
            WHEN category = 'adapter' THEN '"adapter"'
            WHEN category = 'code-driven' THEN '"task"'
            WHEN category = 'data-driven' THEN '"task"'
            ELSE '"task"'
            END::jsonb
                     )
WHERE NOT (default_config ? 'processing_mode')
   OR default_config->>'processing_mode' IS NULL
   OR default_config->>'processing_mode' = '';

-- ============================================================================
-- 5. ADD PERFORMANCE TRACKING
-- ============================================================================
-- Insert initial metrics records for new agents
INSERT INTO agent_metrics (agent_id, agent_type)
SELECT
    gen_random_uuid(),
    type
FROM agent_definitions
WHERE type IN ('content-creator', 'generic')
  AND NOT EXISTS (
    SELECT 1 FROM agent_metrics WHERE agent_type = agent_definitions.type
);

-- Add comment for documentation
COMMENT ON TABLE agent_definitions IS 'Core agent definitions including content-creator with memory and generic orchestrator - updated by migration 080';