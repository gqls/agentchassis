{
  "workflow": {
    "steps": {
      "complete": {
        "action": "complete_workflow",
        "description": "Return hero content"
      },
      "call_researcher": {
        "action": "call_agent",
        "config": {
          "prompt": "Research background information about {{.business_type}} businesses",
          "agent_type": "content-researcher",
          "input_data": {
            "business_type": "{{.input_data.business_type}}"
          },
          "target_role": "researcher"
        },
        "next_step": "generate_hero_content",
        "description": "Get research data"
      },
      "spawn_researcher": {
        "action": "spawn_agent",
        "config": {
          "role": "researcher",
          "agent_type": "content-researcher"
        },
        "next_step": "call_researcher",
        "description": "Spawn research agent"
      },
      "generate_hero_content": {
        "action": "execute_llm_prompt",
        "config": {
          "input_fields": ["call_researcher", "input_data"],
          "prompt_template": "Using this research: {{.call_researcher.result}}\n\nWrite a compelling hero section for {{.business_name}}, a {{.business_type}}. Include a powerful headline and engaging subheadline that captures attention and communicates the core value proposition."
        },
        "next_step": "complete",
        "description": "Generate hero section with research"
      }
    },
    "start_step": "spawn_researcher"
  },
  "ai_service": {
    "model": "claude-3-5-sonnet-20241022",
    "provider": "anthropic",
    "api_key_env_var": "ANTHROPIC_API_KEY"
  },
  "max_tokens": 2000,
  "temperature": 0.7,
  "processing_mode": "task"
}


--

new

--

INSERT INTO agent_definitions (
    id,
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    created_at,
    updated_at,
    capabilities,
    image_repository,
    image_tag,
    command,
    resources,
    topics,
    health_config,
    env_vars,
    version,
    delegation_preferences
) VALUES (
    gen_random_uuid(),
    'content-creator-hero-without-research',
    'Content Creator (Hero - No Research)',
    'Generates hero sections for websites without performing research; uses direct input only.',
    'adapter',
    '{
      "workflow": {
        "steps": {
          "complete": {
            "action": "complete_workflow",
            "description": "Return hero content"
          },
          "generate_hero_content": {
            "action": "execute_llm_prompt",
            "config": {
              "input_fields": ["input_data"],
              "prompt_template": "Write a compelling hero section for {{.business_name}}, a {{.business_type}}. Include a powerful headline and an engaging subheadline that captures attention and communicates the core value proposition."
            },
            "next_step": "complete",
            "description": "Generate hero section"
          }
        },
        "start_step": "generate_hero_content"
      },
      "ai_service": {
        "model": "claude-3-5-sonnet-20241022",
        "provider": "anthropic",
        "api_key_env_var": "ANTHROPIC_API_KEY"
      },
      "max_tokens": 1500,
      "temperature": 0.7,
      "processing_mode": "task"
    }'::jsonb,
    true,
    now(),
    now(),
    '["content-creation", "text-generation"]'::jsonb,
    'docker.io/aqls/agent-chassis',
    'v1.0.400',
    NULL,
    '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
    '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
    '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
    '[]'::jsonb,
    1,
    '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb
);
