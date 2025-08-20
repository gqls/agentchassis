-- Enhanced agent definition with memory configuration
INSERT INTO agent_definitions (type, display_name, description, category, default_config) VALUES
    ('content-creator', 'Content Creator', 'Advanced AI-powered content generation with memory and style adaptation', 'data-driven', '{
  "model": "claude-3-5-sonnet-20241022",
  "temperature": 0.7,
  "max_tokens": 2000,
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
    "blog_post",
    "product_description",
    "social_media",
    "email",
    "landing_page",
    "press_release",
    "technical_doc"
  ],
  "style_options": ["informative", "persuasive", "casual", "professional", "creative"],
  "tone_options": ["friendly", "formal", "conversational", "authoritative", "enthusiastic"],
  "length_options": ["short", "medium", "long"],
  "platform_support": ["generic", "twitter", "linkedin", "facebook", "instagram"]
}')
    ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
      description = EXCLUDED.description,
      category = EXCLUDED.category,
      default_config = EXCLUDED.default_config,
      updated_at = NOW();


-- Update agent_definitions to include behavior configuration
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{behavior}',
        '{
            "processing_mode": "task",
            "capabilities": ["analyze", "extract", "summarize"],
            "workflow": {
                "start_step": "process_input",
                "steps": {
                    "process_input": {
                        "action": "execute_llm_prompt",
                        "next_step": "format_output"
                    },
                    "format_output": {
                        "action": "transform_data",
                        "config": {"transformation": "json_structure"},
                        "next_step": "send_response"
                    },
                    "send_response": {
                        "action": "send_notification",
                        "next_step": "complete"
                    },
                    "complete": {
                        "action": "complete_workflow"
                    }
                }
            }
        }'::jsonb
                     )
WHERE type = 'domain-analyst';

-- Insert the generic agent definition with spawn_group capability
INSERT INTO agent_definitions (type, display_name, description, category, default_config, capabilities)
VALUES (
           'generic',
           'Generic Orchestrator',
           'Generic agent that can spawn groups and orchestrate workflows',
           'code-driven',
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
           '["orchestration", "spawn_group"]'::jsonb
       )
    ON CONFLICT (type) DO UPDATE
                              SET default_config = EXCLUDED.default_config,
                              capabilities = EXCLUDED.capabilities,
                              updated_at = NOW();
EOF