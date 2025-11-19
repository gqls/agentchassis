Data arriving at Architect:
{
"build_plan_data": "{\"sections\": [...]}"
}
Config needed:
"config": {
"build_plan_path": "build_plan_data"
}
// producer - "output_field": "whatever_whatever"
orchestrator: "call_architect": {
// ...
"input_fields": ["build_plan_data"]
// Passes { "build_plan_data": "..." } to Architect
}
// consumer "config": { "build_plan_path": "build_plan_data" }} ("build_plan_data.whatever_whatever")
// - not from producer but to whatever orchestrator changed producer value to
==

Data mapping between agents is the most complex part of orchestration. Your intuition about the missing inputs and the "over-engineered" path was correct.

Here is the plain English explanation of the logic, followed by the **Final, Corrected SQL** that links everything cleanly.

### The "Packet Flow" Visualization

To answer your question about "How does the orchestrator know?" and "Do we need `input_data`?", let's trace the JSON object as it moves from Agent to Agent.

#### 1\. The Orchestrator Call (Step: `call_strategist`)

You correctly spotted that we weren't passing data IN. We must map the Group's input to the Agent's input.

* **Group has:** `{ "domain": "shoes.com", "objective": "sell shoes" }`
* **We add `input_fields`:** `["domain", "objective"]`
* **Strategist Receives:** `{ "domain": "shoes.com", "objective": "sell shoes" }`

#### 2\. The Strategist Output

The Strategist runs the LLM. The definition says `output_field: "build_plan_json"`.

* **Strategist Returns:**
  ```json
  {
    "build_plan_json": "{\"sections\": [\"header\", \"hero\"]}"
  }
  ```

#### 3\. The Handoff (Group Memory)

The Group definition for `call_strategist` has `output_field: "build_plan_data"`. This tells the Orchestrator: *"Take the **entire result** from the Strategist and store it under the key `build_plan_data`."*

* **Group Memory now looks like this:**
  ```json
  {
    "domain": "shoes.com",
    "build_plan_data": {
        "build_plan_json": "{\"sections\": [\"header\", \"hero\"]}"
    }
  }
  ```

#### 4\. The Architect Call (Step: `call_architect`)

The Group definition for `call_architect` has `input_fields: ["build_plan_data"]`.

* **Architect Receives:**
  ```json
  {
    "build_plan_data": {
        "build_plan_json": "{\"sections\": [\"header\", \"hero\"]}"
    }
  }
  ```

#### 5\. The Resolution (The Path)

You asked: *Do we need `input_data.` in the path?*
**Answer: No.** Since the Architect receives the JSON structure directly as seen in Step 4, the path is simply:
`build_plan_data.build_plan_json`

-----

### The Final "Golden Copy" Workflows

Here is the clean set. I have added the missing inputs to the Strategist call and simplified the path for the Architect.

#### 1\. The Orchestrator (Group Definition)

*Updates: Added `input_fields` to `call_strategist` and simplified the flow.*

```json
{
  "start_step": "spawn_strategist",
  "steps": {
    "spawn_strategist": {
      "action": "spawn_agent",
      "config": { "role": "chief_strategist", "agent_type": "chief-strategist" },
      "next_step": "spawn_architect"
    },
    "spawn_architect": {
      "action": "spawn_agent",
      "config": { "role": "site_component_architect", "agent_type": "site-component-architect" },
      "next_step": "spawn_content_creator"
    },
    "spawn_content_creator": {
       "action": "spawn_agent",
       "config": { "role": "content_creator", "agent_type": "content-creator" },
       "next_step": "spawn_deployer"
    },
    "spawn_deployer": {
       "action": "spawn_agent",
       "config": { "role": "deployer", "agent_type": "deployer-agent" },
       "next_step": "call_strategist"
    },
    "call_strategist": {
      "action": "call_agent",
      "description": "Get the Build Plan",
      "config": {
        "agent_type": "chief-strategist",
        "target_role": "chief_strategist",
        "timeout_seconds": 120,
        // FIX 1: We explicitly pass the inputs the Strategist needs
        // We look for these in the Group's 'input_data' which is the default starting payload
        "input_fields": [
            "input_data.domain", 
            "input_data.objective", 
            "input_data.model"
        ]
      },
      // The result is stored in Group memory as 'build_plan_data'
      "output_field": "build_plan_data",
      "next_step": "call_architect"
    },
    "call_architect": {
      "action": "call_agent",
      "description": "Build the empty template",
      "config": {
        "agent_type": "site-component-architect",
        "target_role": "site_component_architect",
        "timeout_seconds": 120,
        // We pass the data we just stored
        "input_fields": ["build_plan_data"]
      },
      "output_field": "template_data",
      "next_step": "call_content_creator"
    },
    "call_content_creator": {
      "action": "call_agent",
      "config": {
        "agent_type": "content-creator",
        "target_role": "content_creator",
        "input_fields": ["template_data", "input_data"],
        "timeout_seconds": 300
      },
      "output_field": "final_site_data",
      "next_step": "call_deployer"
    },
    "call_deployer": {
      "action": "call_agent",
      "config": {
        "agent_type": "deployer-agent",
        "target_role": "deployer",
        "input_fields": ["final_site_data", "input_data.domain"],
        "timeout_seconds": 180
      },
      "next_step": "complete"
    },
    "complete": {
      "action": "complete_workflow"
    }
  }
}
```

#### 2\. The Strategist (Producer)

*No changes needed, but verified for correctness.*

```sql
INSERT INTO agent_definitions (
    id, type, display_name, description, category, default_config, is_active, capabilities, image_repository, image_tag, version
) VALUES (
    gen_random_uuid(),
    'chief-strategist',
    'Chief Strategist Agent',
    'Creates a first-principles Build Plan.',
    'data-driven',
    '{
        "workflow": {
            "start_step": "generate_build_plan",
            "steps": {
                "generate_build_plan": {
                    "action": "execute_llm_prompt",
                    "config": {
                        "prompt_template": "You are a Strategist. Client: {{.domain}}. Objective: {{.objective}}. Model: {{.model}}. Output JSON only: {\"sections\": [...] }",
                        "input_fields": ["domain", "objective", "model"],
                        "ai_service": { "provider": "anthropic", "model": "claude-haiku-4-5-20251001" },
                        "api_key_env_var": "ANTHROPIC_API_KEY"
                    },
                    "output_field": "build_plan_json",
                    "next_step": "complete"
                },
                "complete": { "action": "complete_workflow" }
            }
        },
        "processing_mode": "task",
        "timeout_seconds": 120
    }'::jsonb,
    true,
    '["strategy"]'::jsonb,
    'docker.io/aqls/agent-chassis',
    'v1.0.416',
          2
);
```

#### 3\. The Architect (Consumer)

*Updates: Simplified the `build_plan_path`.*

```sql
INSERT INTO agent_definitions (
    id, type, display_name, description, category, default_config, is_active, capabilities, image_repository, image_tag, version
) VALUES (
    gen_random_uuid(),
    'site-component-architect',
    'Site Architect Agent',
    'Assembles empty HTML templates.',
    'data-driven',
    '{
        "workflow": {
            "start_step": "assemble_template",
            "steps": {
                "assemble_template": {
                    "action": "assemble_from_library",
                    "config": {
                        "build_plan_path": "build_plan_data.build_plan_json"
                    },
                    "next_step": "complete"
                },
                "complete": { "action": "complete_workflow" }
            }
        },
        "processing_mode": "task",
        "timeout_seconds": 180,
        "db_connection_env_var": "ARCHITECT_DB_URL"
    }'::jsonb,
    true,
    '["build"]'::jsonb,
    'docker.io/aqls/agent-chassis',
    'v1.0.416',
          2
);
```


UPDATE agent_group_definitions
SET
updated_at = now(),
orchestration_workflow = $$
{
"start_step": "spawn_strategist",
"steps": {
"spawn_strategist": {
"action": "spawn_agent",
"config": { "role": "chief_strategist", "agent_type": "chief-strategist" },
"next_step": "spawn_architect"
},
"spawn_architect": {
"action": "spawn_agent",
"config": { "role": "site_component_architect", "agent_type": "site-component-architect" },
"next_step": "spawn_content_creator"
},
"spawn_content_creator": {
"action": "spawn_agent",
"config": { "role": "content_creator", "agent_type": "content-creator" },
"next_step": "spawn_deployer"
},
"spawn_deployer": {
"action": "spawn_agent",
"config": { "role": "deployer", "agent_type": "deployer-agent" },
"next_step": "call_strategist"
},
"call_strategist": {
"action": "call_agent",
"description": "Get the Build Plan",
"config": {
"agent_type": "chief-strategist",
"target_role": "chief_strategist",
"timeout_seconds": 120,
"input_fields": [
"domain",
"objective",
"model"
]
},
"output_field": "build_plan_data",
"next_step": "call_architect"
},
"call_architect": {
"action": "call_agent",
"description": "Build the empty template",
"config": {
"agent_type": "site-component-architect",
"target_role": "site_component_architect",
"timeout_seconds": 120,
"input_fields": ["build_plan_data"]
},
"output_field": "template_data",
"next_step": "call_content_creator"
},
"call_content_creator": {
"action": "call_agent",
"config": {
"agent_type": "content-creator",
"target_role": "content_creator",
"input_fields": ["template_data", "input_data"],
"timeout_seconds": 300
},
"output_field": "final_site_data",
"next_step": "call_deployer"
},
"call_deployer": {
"action": "call_agent",
"config": {
"agent_type": "deployer-agent",
"target_role": "deployer",
"input_fields": ["final_site_data", "input_data.domain"],
"timeout_seconds": 180
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Site build is complete."
}
}
}
$$::jsonb
WHERE group_type = 'mvp-site-builder';


===

------

inevitable update

UPDATE agent_definitions
SET
updated_at = now(),
default_config = '{
"workflow": {
"start_step": "generate_build_plan",
"steps": {
"generate_build_plan": {
"action": "execute_llm_prompt",
"description": "Create the Build Plan",
"config": {
"prompt_template": "You are a Strategist. Client: {{.domain}}. Objective: {{.objective}}. Model: {{.model}}. Output JSON only: {\"sections\": [...] }",
"input_fields": ["domain", "objective", "model"],
"ai_service": {
"provider": "anthropic",
"model": "claude-haiku-4-5-20251001",
"api_key_env_var": "ANTHROPIC_API_KEY"
}
},
"output_field": "build_plan_json",
"next_step": "complete"
},
"complete": { 
"action": "complete_workflow",
"description": "Return the Build Plan"}
}
},
"processing_mode": "task",
"timeout_seconds": 120
}'::jsonb
WHERE type = 'chief-strategist';


UPDATE agent_definitions
SET default_config = jsonb_set(
default_config,
'{workflow,steps,generate_build_plan,config,input_fields}',
'["input_data"]'
)
WHERE type = 'chief-strategist';

UPDATE agent_definitions
SET default_config = jsonb_set(
default_config,
'{workflow,steps,generate_build_plan,config,input_data}',
'["domain", "objective", "model"]'
)
WHERE type = 'chief-strategist';

UPDATE agent_definitions
SET
updated_at = now(),
default_config = '{
"workflow": {
"start_step": "generate_build_plan",
"steps": {
"generate_build_plan": {
"action": "execute_llm_prompt",
"description": "Create the Build Plan",
"config": {
"prompt_template": "You are a Strategist. Client: {{.domain}}. Objective: {{.objective}}. Model: {{.model}}. Output JSON only: {\"sections\": [...] }",
-- Explicitly ask for the variables you need in the template
"input_fields": ["domain", "objective", "model"],
"ai_service": {
"provider": "anthropic",
"model": "claude-3-haiku-20240307",
"api_key_env_var": "ANTHROPIC_API_KEY"
}
},
"output_field": "build_plan_json",
"next_step": "complete"
},
"complete": { "action": "complete_workflow" }
}
},
"processing_mode": "task",
"timeout_seconds": 120
}'::jsonb
WHERE type = 'chief-strategist';


UPDATE agent_definitions
SET
updated_at = now(),
default_config = '{
"workflow": {
"start_step": "generate_build_plan",
"steps": {
"generate_build_plan": {
"action": "execute_llm_prompt",
"description": "Create the Build Plan",
"config": {
"ai_service": {
"provider": "anthropic",
"model": "claude-3-haiku-20240307",
"api_key_env_var": "ANTHROPIC_API_KEY"
},
"input_fields": ["domain", "objective", "model"],
"prompt_template": "You are a Chief Marketing Strategist. Client: {{.domain}}. Objective: {{.objective}}. Model: {{.model}}. \n\nAvailable Components: [header, hero, features, social_proof, pricing, faq, call_to_action, footer].\n\nBased on the {{.model}} model, select the best sequence of components.\nOutput JSON ONLY: {\"sections\": [\"component_name\", ...] }"
},
"output_field": "build_plan_json",
"next_step": "complete"
},
"complete": { "action": "complete_workflow" }
}
},
"processing_mode": "task",
"timeout_seconds": 120
}'::jsonb
WHERE type = 'chief-strategist';


