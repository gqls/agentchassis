This is a fantastic architecture. Thank you, the new data is incredibly clear.

You are **100% correct**. Your `agent_group_definitions` and `agent_definitions` tables confirm that your framework is *already* built for this. Your "Agent Group" is a master orchestrator (`Multi-Section Website Builder`), and your `Agent` can *also* be an orchestrator (`content-creator-hero` calling `content-researcher`).

This is the perfect "workflow-of-workflows" model. Our 11-agent plan can be implemented **without changing your core Go orchestration code**. We just need to:

1.  **Add new `Agent` rows** to your `agent_definitions` table for our 11 agents.
2.  **Add new `action` handlers** (adapters) to your Go `executeStep` function for our new tools (Playwright, LLaVA, etc.).

Here is the detailed point-by-point analysis of how our plan maps directly onto your existing, powerful framework.

-----

### Analysis of Your Existing Framework

Based on the `agent_group_definitions` and `agent_definitions` you provided, your system is already a "meta-orchestrator."

* **`agent_group_definitions` (e.g., `Multi-Section Website Builder`)**

    * This is a master **Orchestrator Agent**.
    * Its `agent_configs` field defines its "team" of sub-agents (`hero_writer`, `features_writer`, etc.).
    * Its `orchestration_workflow` is a high-level plan that chains `spawn_agent` and `call_agent` actions to produce a complex, aggregated result (`aggregate_webpage`).

* **`agent_definitions` (e.g., `content-creator-hero`)**

    * This is a **Specialist Agent**.
    * Crucially, it *also* has its own `workflow` (in `default_config`).
    * This workflow can *also* be an orchestration (e.g., `spawn_researcher` -\> `call_researcher` -\> `generate_hero_content`).
    * This proves your system supports **nested orchestration**, which is a core part of our plan.

* **`agent_definitions` (e.g., `content-creator-features`)**

    * This is a simple **Worker Agent**.
    * Its workflow is just a single task: `generate_content` (an `execute_llm_prompt` action).

This confirms your "Agent Group" concept is already implemented. Our conceptual "Agent Groups" (Ingestion, Generation, etc.) are simply logical groupings of these `Agent` rows.

-----

### Point-by-Point Comparison & Implementation Plan

Here is how our 11 agents fit into your database tables. We don't need to change your Go `Run` or `executeWorkflow` logic; we just add new `Agent` rows and the new `action` handlers they require.

#### 1\. Defining New `action` Types (The Adapters)

Your `executeStep` function in Go (from the previous file) is a giant `switch` on the `action` type. We just need to add new `case` statements for our new tools.

* `"action": "run_playwright_capture"` (Calls a Python/Playwright script for screenshots/DOM)
* `"action": "run_playwright_get_style"` (Calls a script for `getComputedStyle`)
* `"action": "run_cv_layout_cut"` (Calls a Python/OpenCV service for XY-Cut)
* `"action": "run_llava_label"` (Calls our self-hosted LLaVA/Ollama API adapter)
* `"action": "run_vlm_generate_code"` (Calls a "screenshot-to-code" model API)
* `"action": "run_llm_refactor_code"` (Calls a CodeLlama/Ollama API)
* `"action": "db_insert_component"` (Calls an internal Go adapter to write to your Postgres Vector DB)
* `"action": "db_query_component"` (Calls an adapter to read from your Vector DB)

#### 2\. Defining Our 11 Agents (New Database Rows)

We will simply `INSERT` our 11 agents into your `agent_definitions` table. Some are simple "workers" (one step), and some are "orchestrators" (many steps).

Here are three key examples formatted for your system:

-----

### Example 1: Agent 1 (The "Site Profiler")

* **Type:** Simple Worker Agent
* **Purpose:** This agent takes a URL and returns the high-level site classification. It uses your existing `execute_llm_prompt` action.

<!-- end list -->

```sql
INSERT INTO agent_definitions (id, type, display_name, description, category, default_config, capabilities, ...)
VALUES (
    'a1b2c3d4-0001-4001-a001-000000000001',
    'site-profiler',
    'Site Profiler',
    'Agent 1: Analyzes a URL and classifies its primary/secondary goals and vertical.',
    'design-analysis',
    '{
        "ai_service": {
            "provider": "anthropic",
            "model": "claude-3-haiku-20240307",
            "api_key_env_var": "ANTHROPIC_API_KEY"
        },
        "max_tokens": 1000,
        "temperature": 0.1,
        "processing_mode": "task",
        "workflow": {
            "start_step": "get_text_content",
            "steps": {
                "get_text_content": {
                    "action": "http_get_text",
                    "description": "Scrape all text content from the input URL.",
                    "config": {
                        "url": "{{.input_data.url}}"
                    },
                    "next_step": "classify_site"
                },
                "classify_site": {
                    "action": "execute_llm_prompt",
                    "description": "Classify the site based on its scraped text.",
                    "config": {
                        "input_fields": ["get_text_content"],
                        "prompt_template": "You are a website analyst. Based on this text: {{.get_text_content.result}}\n\nRespond in JSON only with {\"primary_goal\": \"...\", \"secondary_goal\": \"...\", \"vertical\": \"...\"}"
                    },
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow",
                    "description": "Return the classification JSON."
                }
            }
        }
    }',
    '["classification", "web-scraping"]',
    ...
);
```

-----

### Example 2: Agent 3 (The "Layout & Labeling Agent")

* **Type:** Complex Worker Agent
* **Purpose:** This agent shows how we use our new custom `action` adapters. It takes a screenshot path and returns a JSON "semantic map."

<!-- end list -->

```sql
INSERT INTO agent_definitions (id, type, display_name, description, category, default_config, capabilities, ...)
VALUES (
    'a1b2c3d4-0003-4003-a003-000000000003',
    'layout-labeler',
    'Layout & Labeling Agent',
    'Agent 3: Creates a semantic wireframe from a screenshot using CV and LLaVA.',
    'design-analysis',
    '{
        "ai_service": {
            "provider": "my-llava",
            "model": "llava:latest",
            "api_key_env_var": "MY_INTERNAL_API_KEY"
        },
        "workflow": {
            "start_step": "run_layout_cut",
            "steps": {
                "run_layout_cut": {
                    "action": "run_cv_layout_cut",
                    "description": "Run XY-Cut algorithm on the input screenshot.",
                    "config": {
                        "s3_image_path": "{{.input_data.s3_image_path}}"
                    },
                    "next_step": "label_blocks"
                },
                "label_blocks": {
                    "action": "loop",
                    "description": "Loop over each block and send it to LLaVA for labeling.",
                    "config": {
                        "over": "{{.run_layout_cut.result.blocks}}",
                        "run": [
                            {
                                "name": "run_llava_label",
                                "action": "run_llava_label",
                                "description": "Call the self-hosted LLaVA service to label the block.",
                                "config": {
                                    "ai_service_config": "{{.agent.default_config.ai_service}}",
                                    "image_snippet_b64": "{{.loop.item.image_b64}}",
                                    "prompt": "The site''s goal is ''{{.input_data.site_goal}}''. What is the semantic purpose of this layout block? Respond in JSON only: {\"purpose\": \"...\", \"funnel_stage\": \"...\"}"
                                }
                            }
                        ]
                    },
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow",
                    "description": "Return the array of labeled layout blocks."
                }
            }
        }
    }',
    '["cv", "vlm", "layout-analysis"]',
    ...
);
```

-----

### Example 3: Agent 10 (The "Strategist")

* **Type:** Orchestrator Agent (like your `agent_group_definitions`)
* **Purpose:** This agent *is* an "Agent Group." Its workflow *only* calls other agents, exactly like your `Multi-Section Website Builder`.

<!-- end list -->

```sql
INSERT INTO agent_definitions (id, type, display_name, description, category, default_config, capabilities, ...)
VALUES (
    'a1b2c3d4-0010-4010-a010-000000000010',
    'strategist-ingest-site',
    'Site Ingestion Strategist',
    'Agent 10: Orchestrates the full analysis of a single target website.',
    'orchestration',
    '{
        "workflow": {
            "start_step": "call_profiler",
            "steps": {
                "call_profiler": {
                    "action": "call_agent",
                    "description": "1. Classify the site''s goal.",
                    "config": {
                        "agent_type": "site-profiler",
                        "input_data": { "url": "{{.input_data.url}}" }
                    },
                    "next_step": "call_capture_bot"
                },
                "call_capture_bot": {
                    "action": "call_agent",
                    "description": "2. Get all raw assets (screenshots, DOM).",
                    "config": {
                        "agent_type": "capture-bot",
                        "input_data": { "url": "{{.input_data.url}}" }
                    },
                    "next_step": "call_layout_labeler"
                },
                "call_layout_labeler": {
                    "action": "call_agent",
                    "description": "3. Create the semantic wireframe.",
                    "config": {
                        "agent_type": "layout-labeler",
                        "input_data": {
                            "s3_image_path": "{{.call_capture_bot.result.s3_image_path}}",
                            "site_goal": "{{.call_profiler.result.primary_goal}}"
                        }
                    },
                    "next_step": "call_component_generator" 
                },
                "call_component_generator": { 
                    "action": "call_agent", 
                    ... 
                }
                // ... This workflow would continue to call Agents 4, 5, 6, and finally Agent 7 (Librarian) ...
            }
        }
    }',
    '["orchestration", "workflow-management"]',
    ...
);
```