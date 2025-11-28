-- ============================================================================
-- UPDATED CHIEF STRATEGIST - Now uses briefing data for enhanced strategy
-- ============================================================================

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = '{
    "workflow": {
      "start_step": "generate_build_plan",
      "steps": {
        "generate_build_plan": {
          "action": "execute_llm_prompt",
          "config": {
            "ai_service": {
              "provider": "anthropic",
              "model": "claude-haiku-4-5-20251001",
              "api_key_env_var": "ANTHROPIC_API_KEY",
              "max_tokens": 3000
            },
            "input_fields": ["input_data", "brief_data"],
            "output_field": "build_plan_json",
            "prompt_template": "You are a website strategist creating a Build Plan based on behavioral psychology and conversion optimization.\n\nWebsite Request:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Behavioral Model: {{.input_data.model}}\n\n{{if .brief_data}}Detailed Brief (from analysis):\n{{.brief_data.structured_brief.result}}\n{{end}}\n\nYour Task:\nCreate a strategic Build Plan that maps behavioral psychology to website sections.\n\nBehavioral Models Available:\n- PAS (Problem-Agitate-Solution): Best for pain-point focused products\n- AIDA (Attention-Interest-Desire-Action): Best for general conversion\n- FAB (Features-Advantages-Benefits): Best for feature-rich products\n- 4Ps (Promise-Picture-Proof-Push): Best for aspirational products\n\nFor each model, define the appropriate sections:\n\nPAS sections: [\"problem_statement\", \"agitation\", \"solution_provider\", \"social_proof\", \"cta\"]\nAIDA sections: [\"attention_hero\", \"interest_features\", \"desire_benefits\", \"action_cta\"]\nFAB sections: [\"features_showcase\", \"advantages_comparison\", \"benefits_outcome\", \"cta\"]\n4Ps sections: [\"promise_hero\", \"picture_vision\", \"proof_testimonials\", \"push_cta\"]\n\n{{if .brief_data}}Use the briefing analysis to enhance your section selection:\n- Recommended theme: Use the theme.recommended from the brief\n- Messaging: Incorporate key_messages and usps into section guidance\n- Audience: Consider the target audience when defining section priorities\n{{end}}\n\nReturn ONLY valid JSON with this structure:\n{\n  \"model\": \"PAS|AIDA|FAB|4Ps\",\n  \"sections\": [\"section1\", \"section2\", ...],\n  \"section_guidance\": {\n    \"section_name\": {\n      \"purpose\": \"what this section should achieve\",\n      \"key_message\": \"primary message for this section\",\n      \"tone\": \"section-specific tone guidance\"\n    }\n  },\n  \"theme_recommendation\": \"recommended theme name\",\n  \"theme_tags\": [\"semantic\", \"tags\"],\n  \"conversion_priority\": [\"most important sections for conversion\"]\n}\n\nDO NOT include any text outside the JSON object."
          },
          "next_step": "complete"
        },
        "complete": {
          "action": "complete_workflow",
          "description": "Return the Build Plan"
        }
      }
    },
    "processing_mode": "task",
    "timeout_seconds": 120
  }'::jsonb
WHERE type = 'chief-strategist';