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
`"ai_service": {
"provider": "anthropic",
"model": "claude-haiku-4-5-20251001",
"api_key_env_var": "ANTHROPIC_API_KEY"
},`
"input_data": ["domain", "objective", "model"],
"prompt_template": "You are a Chief Marketing Strategist. Client: {{.domain}}. Objective: {{.objective}}. Model: {{.model}}. \n\nAvailable Components: [header, hero, features, social_proof, pricing, faq, call_to_action, footer].\n\nBased on the {{.model}} model, select the best sequence of components. Then for each component devise a plan for the copy structure, suggested copy and suggested graphics style that suits the objective {{ .objective }} and the marketing model {{ .model }}.\nOutput JSON ONLY: {\"sections\": [\"component_name\", ...] }"
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
"input_data",
"template_data"
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

architect

{
"workflow": {
"start_step": "assemble_template",
"steps": {
"assemble_template": {
"action": "assemble_from_library",
"config": {
"build_plan_path": "input_data.call_strategist.generate_build_plan.result"
},
"next_step": "complete"
},
"complete": {"action": "complete_workflow"}
}
}
}

UPDATE agent_definitions
SET
updated_at = now(),
default_config = '{
"workflow": {
"start_step": "assemble_template",
"steps": {
"assemble_template": {
"action": "assemble_from_library",
"config": {
"input_fields": ["build_plan_data"]
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow"
}
}
},
"processing_mode": "task",
"timeout_seconds": 120
}'::jsonb
WHERE type = 'site-component-architect';

UPDATE agent_definitions
SET default_config = '{
"workflow": {
"start_step": "generate_content",
"steps": {
"generate_content": {
"action": "execute_llm_prompt",
"config": {
"ai_service": {
"provider": "anthropic",
"model": "claude-haiku-4-5-20251001",
"api_key_env_var": "ANTHROPIC_API_KEY"
},
"input_fields": ["input_data"],
"output_field": "generated_content",
"prompt_template": "You are creating website content for {{.input_data.input_data.domain}} with objective: {{.input_data.input_data.objective}}.\n\nHere is the HTML template with placeholders:\n{{.input_data.template_data.assemble_template.stitched_html_template}}\n\nHere are the content requirements (default values to customize):\n{{.input_data.template_data.assemble_template.content_requirements}}\n\nGenerate customized, compelling content for this boxing ticket sales website. Replace the generic placeholder values with boxing-specific, subtle sales-focused copy.\n\nReturn ONLY the complete HTML with all placeholders replaced with actual content. Do not include any explanation."
},
"next_step": "complete",
"description": "Generate customized website content"
},
"complete": {
"action": "complete_workflow",
"description": "Return final HTML"
}
}
},
"processing_mode": "task",
"timeout_seconds": 300
}'::jsonb
WHERE type = 'content-creator';


UPDATE agent_definitions
SET default_config = '{
"workflow": {
"start_step": "prepare_commit_data",
"steps": {
"prepare_commit_data": {
"action": "json_transform",
"config": {
"input_field": "input_data.final_site_data.final_html",
"jq_expression": "{\"index.html\": .}",
"output_field": "files_map"
},
"description": "Transform final_html into a files object for the adapter",
"next_step": "commit_to_git"
},
"commit_to_git": {
"action": "git_commit",
"config": {
"commit_message": "MVP v1: Initial site build.",
"files": "{{.files_map}}",
"repo_name": "{{.input_data.domain}}"
},
"description": "Send commit request to the git-adapter",
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return the Git repo URL"
}
}
},
"processing_mode": "task",
"timeout_seconds": 180
}'::jsonb
WHERE type = 'deployer-agent';


===
revised
===

-- Update the orchestration workflow to properly pass data between agents
UPDATE agent_group_definitions
SET
updated_at = now(),
orchestration_workflow = $$
{
"start_step": "spawn_strategist",
"steps": {
"spawn_strategist": {
"action": "spawn_agent",
"config": {
"role": "chief_strategist",
"agent_type": "chief-strategist"
},
"next_step": "spawn_architect",
"description": "Spawn Chief Strategist"
},
"spawn_architect": {
"action": "spawn_agent",
"config": {
"role": "site_component_architect",
"agent_type": "site-component-architect"
},
"next_step": "spawn_content_creator",
"description": "Spawn Site Component Architect"
},
"spawn_content_creator": {
"action": "spawn_agent",
"config": {
"role": "content_creator",
"agent_type": "content-creator"
},
"next_step": "spawn_deployer",
"description": "Spawn Content Creator"
},
"spawn_deployer": {
"action": "spawn_agent",
"config": {
"role": "deployer",
"agent_type": "deployer-agent"
},
"next_step": "call_strategist",
"description": "Spawn Deployer"
},
"call_strategist": {
"action": "call_agent",
"description": "Get the Build Plan from the Strategist",
"config": {
"agent_type": "chief-strategist",
"target_role": "chief_strategist",
"timeout_seconds": 120
},
"output_field": "build_plan_data",
"next_step": "call_architect"
},
"call_architect": {
"action": "call_agent",
"description": "Build the empty template from the Build Plan",
"config": {
"agent_type": "site-component-architect",
"target_role": "site_component_architect",
"timeout_seconds": 120,
"input_fields": ["build_plan_data", "input_data"]
},
"output_field": "template_data",
"next_step": "call_content_creator"
},
"call_content_creator": {
"action": "call_agent",
"description": "Fill the template with content",
"config": {
"agent_type": "content-creator",
"target_role": "content_creator",
"input_fields": ["template_data", "build_plan_data", "input_data"],
"timeout_seconds": 300
},
"output_field": "final_site_data",
"next_step": "call_deployer"
},
"call_deployer": {
"action": "call_agent",
"description": "Push the final site to Git",
"config": {
"agent_type": "deployer-agent",
"target_role": "deployer",
"input_fields": ["final_site_data", "input_data"],
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

-- Update the site-component-architect agent definition
UPDATE agent_definitions
SET
updated_at = now(),
default_config = '{
"workflow": {
"start_step": "assemble_template",
"steps": {
"assemble_template": {
"action": "assemble_from_library",
"config": {
-- "build_plan_field": "build_plan_data",
"input_fields": ["build_plan_data"]
},
"next_step": "complete",
"description": "Assemble HTML template from component library"
},
"complete": {
"action": "complete_workflow",
"description": "Return the assembled template"
}
}
},
"processing_mode": "task",
"timeout_seconds": 180,
}'::jsonb
WHERE type = 'site-component-architect';

-- Update the content-creator agent definition with corrected prompt template
UPDATE agent_definitions
SET
updated_at = now(),
default_config = '{
"workflow": {
"start_step": "generate_content",
"steps": {
"generate_content": {
"action": "execute_llm_prompt",
"config": {
"ai_service": {
"provider": "anthropic",
"model": "claude-haiku-4-5-20251001",
"api_key_env_var": "ANTHROPIC_API_KEY",
"max_tokens": 4000
},
"input_fields": ["template_data", "build_plan_data", "input_data"],
"output_field": "filled_html",
"prompt_template": "You are a professional website content creator specialising in creating compelling, industry-specific content.\n\nWebsite Details:\n- Domain: {{.domain}}\n- Objective: {{.objective}}\n- Model: {{.model}}\n\nBuild Strategy (from strategist):\n{{.build_plan_data.generate_build_plan.result}}\n\nHTML Template to Fill:\n{{.template_data.assemble_template.stitched_html_template}}\n\nContent Requirements (placeholders to replace):\n{{.template_data.assemble_template.content_requirements}}\n\nYour Task:\n1. Parse the HTML template and identify ALL placeholder variables (format: {{.placeholder_name}})\n2. For each placeholder, create compelling, industry-specific content that:\n   - Matches the domain ({{.domain}})\n   - Aligns with the objective ({{.objective}})\n   - Uses the suggested copy from the build strategy where available\n   - Is subtle, sales-focused and conversion-optimised\n3. Replace EVERY placeholder with actual, high quality, well written, subtle, copy\n\nGuidelines:\n- We do not have real testimonials and we do not want to give fake content so please fill this area with copy that expounds the benefits of our products and services and is domain specific but the people credited to the quotes or statements are like: [Future You], or [Soon to be Delighted Customer] -, or other non-fake endorsements like that. (happy, pleased, delighted, will/would make you want to return).\n- Use action-oriented language for CTAs\n- Include trust signals and security messaging\n- Keep brand consistency throughout\n\nReturn ONLY the complete HTML with all placeholders replaced. No explanations or markdown.\nPlease double check that all placeholders have been replaced with copy."
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return the filled HTML template",
"output_field": ["final_html"]
}
}
},
"processing_mode": "task",
"timeout_seconds": 300
}'::jsonb
WHERE type = 'content-creator';