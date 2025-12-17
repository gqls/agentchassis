id                     | d8c6adcc-7ce7-4b18-af1b-2f34db616c06
type                   | multipage-website-builder
display_name           | Multi-Page Website Builder
description            | Builds large websites (20+ pages) using batched generation to avoid token limits
category               | orchestrator
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Site assembly complete"}, "generate_batch_1": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 16000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture", "input_data"], "prompt_template": "Generate complete HTML for pages 1-4 of {{.input_data.domain}}:\n\n1. index.html (Home)\n2. about.html (About Us)\n3. services.html (Services)\n4. team.html (Team)\n\nArchitecture: {{.site_architecture.result}}\n\nFor each page, return complete HTML from <!DOCTYPE> to </html>.\n\nRETURN AS JSON:\n{\n  \"index.html\": \"<!DOCTYPE html>...\",\n  \"about.html\": \"<!DOCTYPE html>...\",\n  \"services.html\": \"<!DOCTYPE html>...\",\n  \"team.html\": \"<!DOCTYPE html>...\"\n}\n\nIMPORTANT: DO NOT include shared CSS - it will be injected automatically."}, "next_step": "generate_batch_2", "description": "Generate pages 1-4", "output_field": "batch_1_pages"}, "generate_batch_2": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 16000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture", "input_data"], "prompt_template": "Generate complete HTML for pages 5-8:\n\n5. products.html (Products)\n6. pricing.html (Pricing)\n7. features.html (Features)\n8. testimonials.html (Testimonials)\n\nReturn as JSON map of filename to HTML."}, "next_step": "generate_batch_3", "description": "Generate pages 5-8", "output_field": "batch_2_pages"}, "generate_batch_3": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 16000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture", "input_data"], "prompt_template": "Generate complete HTML for pages 9-12:\n\n9. case-studies.html (Case Studies)\n10. blog.html (Blog)\n11. resources.html (Resources)\n12. faq.html (FAQ)\n\nReturn as JSON map."}, "next_step": "generate_batch_4", "description": "Generate pages 9-12", "output_field": "batch_3_pages"}, "generate_batch_4": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 16000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture", "input_data"], "prompt_template": "Generate complete HTML for pages 13-16:\n\n13. support.html (Support)\n14. contact.html (Contact)\n15. careers.html (Careers)\n16. partners.html (Partners)\n\nReturn as JSON map."}, "next_step": "generate_batch_5", "description": "Generate pages 13-16", "output_field": "batch_4_pages"}, "generate_batch_5": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 16000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture", "input_data"], "prompt_template": "Generate complete HTML for pages 17-20:\n\n17. press.html (Press)\n18. legal.html (Legal)\n19. privacy.html (Privacy Policy)\n20. sitemap.html (Sitemap)\n\nReturn as JSON map."}, "next_step": "assemble_multipage_site", "description": "Generate pages 17-20", "output_field": "batch_5_pages"}, "analyze_requirements": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5-20250514", "provider": "anthropic", "max_tokens": 4000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["input_data.domain", "input_data.page_list"], "prompt_template": "Analyze requirements for a {{.input_data.domain}} website with these pages: {{.input_data.page_list}}.\n\nCreate a site architecture including:\n- Navigation structure\n- Shared design elements\n- Color scheme and typography\n- Content themes for each page\n\nReturn as JSON."}, "next_step": "generate_shared_styles", "description": "Analyze site requirements and create architecture", "output_field": "site_architecture"}, "generate_shared_styles": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-haiku-4-5-20251001", "provider": "anthropic", "max_tokens": 8000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["site_architecture"], "prompt_template": "Generate shared CSS for entire site based on:\n{{.site_architecture.result}}\n\nInclude:\n- CSS reset and base styles\n- Typography system\n- Color variables\n- Responsive layout utilities\n- Navigation styles\n- Footer styles\n\nReturn ONLY CSS (no HTML, no markdown)."}, "next_step": "generate_batch_1", "description": "Generate shared CSS for all pages", "output_field": "shared_styles"}, "assemble_multipage_site": {"action": "assemble_multipage_site", "config": {"batch_fields": ["batch_1_pages", "batch_2_pages", "batch_3_pages", "batch_4_pages", "batch_5_pages"], "stream_to_s3": true, "index_html_field": "batch_1_pages.index.html", "navigation_field": "site_architecture.navigation", "shared_styles_field": "shared_styles.result"}, "next_step": "complete", "description": "Assemble all pages with shared styles and navigation", "output_field": "assembled_site"}}, "start_step": "analyze_requirements"}, "processing_mode": "orchestrator", "timeout_seconds": 600}
is_active              | t
created_at             | 2025-12-06 19:02:26.358753+00
updated_at             | 2025-12-11 11:15:23.041575+00
deleted_at             |
capabilities           | ["orchestration", "website-builder", "multi-page", "batched-generation"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.528
command                |
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    |
task_workflow          |
orchestrator_workflow  |
orchestration_workflow |
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         |
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         |
output_contract        |


Add briefing_questionnaire to multipage-website-builder

UPDATE agent_definitions
SET briefing_questionnaire = '{
  "sections": [
    {
      "name": "company_info",
      "title": "Company Information",
      "questions": [
        {"type": "text", "field": "company_name", "label": "Company Name", "required": true},
        {"type": "textarea", "field": "about_us", "label": "About Us", "required": true},
        {"type": "text", "field": "tagline", "label": "Tagline", "required": false}
      ]
    },
    {
      "name": "services",
      "title": "Services & Offerings",
      "questions": [
        {"type": "json_array", "field": "services", "label": "Services (list of {name, description})", "required": true}
      ]
    },
    {
      "name": "team",
      "title": "Team & Leadership",
      "questions": [
        {"type": "json_array", "field": "leadership_team", "label": "Leadership Team (list of {name, title, bio})", "required": false}
      ]
    },
    {
      "name": "portfolio",
      "title": "Case Studies & Portfolio",
      "questions": [
        {"type": "json_array", "field": "case_studies", "label": "Case Studies", "required": false}
      ]
    },
    {
      "name": "contact",
      "title": "Contact Information",
      "questions": [
        {"type": "text", "field": "contact_email", "label": "Email", "required": true},
        {"type": "text", "field": "contact_phone", "label": "Phone", "required": false},
        {"type": "text", "field": "headquarters", "label": "Location", "required": false}
      ]
    },
    {
      "name": "features",
      "title": "Site Features",
      "questions": [
        {"type": "boolean", "field": "has_blog", "label": "Include Blog?", "default": false},
        {"type": "boolean", "field": "has_careers", "label": "Include Careers?", "default": false}
      ]
    }
  ]
}'::jsonb
WHERE type = 'multipage-website-builder';


{
  "workflow": {
    "start_step": "analyze_requirements",
    "steps": {
      "analyze_requirements": {
        "action": "execute_llm_prompt",
        "description": "Analyze site requirements and create architecture",
        "input_fields": [
          "input_data.domain",
          "input_data.page_list"
        ],
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-5-20250514",
            "provider": "anthropic",
            "max_tokens": 4000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "prompt_template": "Analyze requirements for a {{.input_data.domain}} website with these pages: {{.input_data.page_list}}.\n\nCreate a site architecture including:\n- Navigation structure\n- Shared design elements\n- Color scheme and typography\n- Content themes for each page\n\nReturn as JSON."
        },
        "next_step": "generate_shared_styles",
        "output_field": "site_architecture"
      },

      "generate_shared_styles": {
        "action": "execute_llm_prompt",
        "description": "Generate shared CSS for all pages",
        "input_fields": [
          "site_architecture"
        ],
        "config": {
          "ai_service": {
            "model": "claude-haiku-4-5-20251001",
            "provider": "anthropic",
            "max_tokens": 8000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "prompt_template": "Generate shared CSS for entire site based on:\n{{.site_architecture.result}}\n\nInclude:\n- CSS reset and base styles\n- Typography system\n- Color variables\n- Responsive layout utilities\n- Navigation styles\n- Footer styles\n\nReturn ONLY CSS (no HTML, no markdown)."
        },
        "next_step": "generate_batch_1",
        "output_field": "shared_styles"
      },

      "generate_batch_1": {
        "action": "execute_llm_prompt",
        "description": "Generate pages 1–4",
        "input_fields": [
          "site_architecture",
          "input_data"
        ],
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-5-20250514",
            "provider": "anthropic",
            "max_tokens": 16000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "prompt_template": "Generate complete HTML for pages 1-4 of {{.input_data.domain}}:\n\n1. index.html (Home)\n2. about.html (About Us)\n3. services.html (Services)\n4. team.html (Team)\n\nArchitecture: {{.site_architecture.result}}\n\nFor each page, return complete HTML from <!DOCTYPE> to </html>.\n\nRETURN AS JSON:\n{\n  \"index.html\": \"<!DOCTYPE html>...\",\n  \"about.html\": \"<!DOCTYPE html>...\",\n  \"services.html\": \"<!DOCTYPE html>...\",\n  \"team.html\": \"<!DOCTYPE html>...\"\n}\n\nIMPORTANT: DO NOT include shared CSS - it will be injected automatically."
        },
        "next_step": "generate_batch_2",
        "output_field": "batch_1_pages"
      },

      "generate_batch_2": {
        "action": "execute_llm_prompt",
        "description": "Generate pages 5–8",
        "input_fields": [
          "site_architecture",
          "input_data"
        ],
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-5-20250514",
            "provider": "anthropic",
            "max_tokens": 16000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "prompt_template": "Generate complete HTML for pages 5-8:\n\n5. products.html (Products)\n6. pricing.html (Pricing)\n7. features.html (Features)\n8. testimonials.html (Testimonials)\n\nReturn as JSON map of filename to HTML."
        },
        "next_step": "generate_batch_3",
        "output_field": "batch_2_pages"
      },

      "generate_batch_3": {
        "action": "execute_llm_prompt",
        "description": "Generate pages 9–12",
        "input_fields": [
          "site_architecture",
          "input_data"
        ],
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-5-20250514",
            "provider": "anthropic",
            "max_tokens": 16000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "prompt_template": "Generate complete HTML for pages 9-12:\n\n9. case-studies.html (Case Studies)\n10. blog.html (Blog)\n11. resources.html (Resources)\n12. faq.html (FAQ)\n\nReturn as JSON map."
        },
        "next_step": "generate_batch_4",
        "output_field": "batch_3_pages"
      },

      "generate_batch_4": {
        "action": "execute_llm_prompt",
        "description": "Generate pages 13–16",
        "input_fields": [
          "site_architecture",
          "input_data"
        ],
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-5-20250514",
            "provider": "anthropic",
            "max_tokens": 16000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "prompt_template": "Generate complete HTML for pages 13-16:\n\n13. support.html (Support)\n14. contact.html (Contact)\n15. careers.html (Careers)\n16. partners.html (Partners)\n\nReturn as JSON map."
        },
        "next_step": "generate_batch_5",
        "output_field": "batch_4_pages"
      },

      "generate_batch_5": {
        "action": "execute_llm_prompt",
        "description": "Generate pages 17–20",
        "input_fields": [
          "site_architecture",
          "input_data"
        ],
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-5-20250514",
            "provider": "anthropic",
            "max_tokens": 16000,
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "prompt_template": "Generate complete HTML for pages 17-20:\n\n17. press.html (Press)\n18. legal.html (Legal)\n19. privacy.html (Privacy Policy)\n20. sitemap.html (Sitemap)\n\nReturn as JSON map."
        },
        "next_step": "assemble_multipage_site",
        "output_field": "batch_5_pages"
      },

      "assemble_multipage_site": {
        "action": "assemble_multipage_site",
        "description": "Assemble all pages with shared styles and navigation",
        "config": {
          "batch_fields": [
            "batch_1_pages",
            "batch_2_pages",
            "batch_3_pages",
            "batch_4_pages",
            "batch_5_pages"
          ],
          "stream_to_s3": true,
          "index_html_field": "batch_1_pages.index.html",
          "navigation_field": "site_architecture.navigation",
          "shared_styles_field": "shared_styles.result"
        },
        "next_step": "complete",
        "output_field": "assembled_site"
      },

      "complete": {
        "action": "complete_workflow",
        "description": "Site assembly complete"
      }
    }
  },

  "processing_mode": "orchestrator",
  "timeout_seconds": 600
}


===========================================================================================
===========================================================================================
===========================================================================================

Revised: simplified. adds loop. one page at a time

{
    "workflow": {
        "start_step": "call_strategist",
        "steps": {
            "call_strategist": {
                "action": "call_agent",
                "config": {
                    "agent_type": "chief-strategist",
                    "timeout_seconds": 120
                },
                "next_step": "generate_pages_loop",
                "output_field": "page_plan"
            },

            "generate_pages_loop": {
                "action": "loop",
                "config": {
                    "iterate_over": "page_plan.pages",
                    "loop_var": "current_page",
                    "max_iterations": 10,
                    "substeps": {
                        "generate_page": {
                            "action": "call_agent",
                            "config": {
                                "agent_type": "content-creator",
                                "input_fields": ["current_page"],
                                "timeout_seconds": 180
                            },
                            "output_field": "page_html"
                        }
                    }
                },
                "next_step": "assemble_site",
                "output_field": "all_pages"
            },

            "assemble_site": {
                "action": "assemble_multipage_site",
                "config": {
                    "pages_field": "all_pages",
                    "add_navigation": true
                },
                "next_step": "deploy",
                "output_field": "site_files"
            },

            "deploy": {
                "action": "call_agent",
                "config": {
                    "agent_type": "deployer-agent",
                    "input_fields": ["site_files"]
                },
                "next_step": "complete"
            },

            "complete": {
                "action": "complete_workflow"
            }
        }
    }
}

---
update statement for above

-- ============================================================================
-- UPDATE: multipage-website-builder Agent Definition
-- New simplified workflow using consolidated actions
-- ============================================================================

-- First, let's see the current config (for reference)
-- SELECT type, display_name, default_config FROM agent_definitions
-- WHERE type = 'multipage-website-builder';

-- Update the multipage-website-builder with new sequential workflow
UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "task",
    "timeout_seconds": 600,
    "workflow": {
        "start_step": "call_strategist",
        "steps": {
            "call_strategist": {
                "action": "call_agent",
                "config": {
                    "agent_type": "chief-strategist",
                    "target_role": "strategist",
                    "timeout_seconds": 120
                },
                "next_step": "generate_pages_loop",
                "output_field": "page_plan",
                "description": "Get page plan from chief strategist"
            },

            "generate_pages_loop": {
                "action": "loop",
                "config": {
                    "iterate_over": "page_plan.pages",
                    "loop_var": "current_page",
                    "max_iterations": 10,
                    "substeps": {
                        "generate_page": {
                            "action": "call_agent",
                            "config": {
                                "agent_type": "content-creator",
                                "target_role": "writer",
                                "input_fields": ["current_page", "input_data"],
                                "timeout_seconds": 180
                            },
                            "output_field": "page_html",
                            "description": "Generate content for each page"
                        }
                    }
                },
                "next_step": "assemble_site",
                "output_field": "all_pages",
                "description": "Generate all pages sequentially"
            },

            "assemble_site": {
                "action": "assemble_multipage_site",
                "config": {
                    "pages_field": "all_pages",
                    "add_navigation": true,
                    "generate_standard_pages": true
                },
                "next_step": "deploy",
                "output_field": "site_files",
                "description": "Assemble pages into complete site with navigation"
            },

            "deploy": {
                "action": "call_agent",
                "config": {
                    "agent_type": "deployer-agent",
                    "target_role": "deployer",
                    "input_fields": ["site_files", "input_data"],
                    "timeout_seconds": 180
                },
                "next_step": "complete",
                "output_field": "deployment_result",
                "description": "Deploy site to git repository"
            },

            "complete": {
                "action": "complete_workflow",
                "description": "Multipage site build complete"
            }
        }
    }
}'::jsonb,
updated_at = now()
WHERE type = 'multipage-website-builder';

-- Verify the update
SELECT
    type,
    display_name,
    jsonb_pretty(default_config->'workflow'->'steps') as workflow_steps,
    updated_at
FROM agent_definitions
WHERE type = 'multipage-website-builder';

-- ============================================================================
-- NOTES
-- ============================================================================

/*
This workflow is now:

1. SEQUENTIAL (not parallel)
   - Strategist creates page plan
   - Loop generates pages one at a time
   - Assembly happens after all pages done
   - Deployment happens last

2. SIMPLE (like landing-page-builder)
   - Clear step progression
   - No complex nested data structures
   - Easy to debug

3. USES CONSOLIDATED ACTIONS
   - assemble_multipage_site (single clear purpose)
   - No SQL in config
   - All complexity hidden in actions

Expected flow:
- Input: {"domain": "example.com", "objective": "consulting site"}
- Strategist returns: {"pages": ["index", "services", "about"]}
- Loop generates: {"index": "...", "services": "...", "about": "..."}
- Assembly adds: navigation + contact page
- Deployer commits: all files to git
- Output: {"deployment_result": {...}}

If this completes successfully, you have a working multipage builder.
*/

-- ============================================================================
-- TROUBLESHOOTING QUERIES
-- ============================================================================

-- Check if agent definition exists
SELECT COUNT(*) as exists
FROM agent_definitions
WHERE type = 'multipage-website-builder';

-- If count is 0, the agent doesn't exist and you'll need to create it first:
/*
INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    capabilities,
    is_active
)
VALUES (
    'multipage-website-builder',
    'Multipage Website Builder',
    'Builds complete multi-page websites with navigation',
    'website_builder',
    '{...workflow config from above...}'::jsonb,
    '["website_building", "multipage", "orchestration"]'::jsonb,
    true
);
*/

-- View current orchestrations using this agent
SELECT
    o.id,
    o.status,
    o.current_step,
    o.created_at,
    o.updated_at
FROM orchestrator_state o
WHERE o.workflow_type = 'multipage-website-builder'
ORDER BY o.created_at DESC
    LIMIT 10;

-- Check for stuck workflows
SELECT
    id,
    status,
    current_step,
    EXTRACT(EPOCH FROM (NOW() - updated_at))/3600 as hours_since_update
FROM orchestrator_state
WHERE workflow_type = 'multipage-website-builder'
  AND status IN ('RUNNING', 'AWAITING_RESPONSES')
  AND updated_at < NOW() - INTERVAL '1 hour';

==

add spawning

-- Fixed multipage-website-builder workflow with spawn steps

-- Fixed multipage-website-builder workflow with all agents spawned first

UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "task",
    "timeout_seconds": 600,
    "workflow": {
        "start_step": "spawn_strategist",
        "steps": {
            "spawn_strategist": {
                "action": "spawn_agent",
                "config": {
                    "agent_type": "chief-strategist",
                    "role": "strategist"
                },
                "next_step": "spawn_content_creator",
                "output_field": "strategist_info",
                "description": "Spawn chief strategist agent"
            },

            "spawn_content_creator": {
                "action": "spawn_agent",
                "config": {
                    "agent_type": "content-creator",
                    "role": "writer"
                },
                "next_step": "spawn_deployer",
                "output_field": "writer_info",
                "description": "Spawn content creator agent for loop iterations"
            },

            "spawn_deployer": {
                "action": "spawn_agent",
                "config": {
                    "agent_type": "deployer-agent",
                    "role": "deployer"
                },
                "next_step": "call_strategist",
                "output_field": "deployer_info",
                "description": "Spawn deployer agent"
            },

            "call_strategist": {
                "action": "call_agent",
                "config": {
                    "agent_type": "chief-strategist",
                    "target_role": "strategist",
                    "timeout_seconds": 120
                },
                "next_step": "generate_pages_loop",
                "output_field": "page_plan",
                "description": "Get page plan from chief strategist"
            },

            "generate_pages_loop": {
                "action": "loop",
                "config": {
                    "iterate_over": "page_plan.pages",
                    "loop_var": "current_page",
                    "max_iterations": 10,
                    "substeps": {
                        "generate_page": {
                            "action": "call_agent",
                            "config": {
                                "agent_type": "content-creator",
                                "target_role": "writer",
                                "input_fields": ["current_page", "input_data"],
                                "timeout_seconds": 180
                            },
                            "output_field": "page_html",
                            "description": "Generate content for each page"
                        }
                    }
                },
                "next_step": "assemble_site",
                "output_field": "all_pages",
                "description": "Generate all pages sequentially"
            },

            "assemble_site": {
                "action": "assemble_multipage_site",
                "config": {
                    "pages_field": "all_pages",
                    "add_navigation": true,
                    "generate_standard_pages": true
                },
                "next_step": "deploy",
                "output_field": "site_files",
                "description": "Assemble pages into complete site with navigation"
            },

            "deploy": {
                "action": "call_agent",
                "config": {
                    "agent_type": "deployer-agent",
                    "target_role": "deployer",
                    "input_fields": ["site_files", "input_data"],
                    "timeout_seconds": 180
                },
                "next_step": "complete",
                "output_field": "deployment_result",
                "description": "Deploy site to git repository"
            },

            "complete": {
                "action": "complete_workflow",
                "description": "Multipage site build complete"
            }
        }
    }
}'::jsonb,
updated_at = now()
WHERE type = 'multipage-website-builder';


---
-- change from looking for pages to looking for sections

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_pages_loop,config,iterate_over}',
        '"page_plan.sections"'
                     )
WHERE type = 'multipage-website-builder';

--

added transform step

      -- Update multipage-website-builder to use corrected strategist output
-- After fix, strategist returns {sections: [...], component_details: {...}}
-- This is stored in page_plan, so loop accesses page_plan.sections

UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "task",
    "timeout_seconds": 600,
    "workflow": {
        "start_step": "spawn_strategist",
        "steps": {
            "spawn_strategist": {
                "action": "spawn_agent",
                "config": {
                    "agent_type": "chief-strategist",
                    "role": "strategist"
                },
                "next_step": "spawn_content_creator",
                "output_field": "strategist_info",
                "description": "Spawn chief strategist agent"
            },

            "spawn_content_creator": {
                "action": "spawn_agent",
                "config": {
                    "agent_type": "content-creator",
                    "role": "writer"
                },
                "next_step": "spawn_deployer",
                "output_field": "writer_info",
                "description": "Spawn content creator agent for loop iterations"
            },

            "spawn_deployer": {
                "action": "spawn_agent",
                "config": {
                    "agent_type": "deployer-agent",
                    "role": "deployer"
                },
                "next_step": "call_strategist",
                "output_field": "deployer_info",
                "description": "Spawn deployer agent"
            },

            "call_strategist": {
                "action": "call_agent",
                "config": {
                    "agent_type": "chief-strategist",
                    "target_role": "strategist",
                    "timeout_seconds": 120
                },
                "next_step": "generate_pages_loop",
                "output_field": "page_plan",
                "description": "Get page plan from chief strategist"
            },

            "generate_pages_loop": {
                "action": "loop",
                "config": {
                    "iterate_over": "page_plan.plan_data.sections",
                    "loop_var": "current_page",
                    "max_iterations": 10,
                    "substeps": {
                        "generate_page": {
                            "action": "call_agent",
                            "config": {
                                "agent_type": "content-creator",
                                "target_role": "writer",
                                "input_fields": ["current_page", "input_data"],
                                "timeout_seconds": 180
                            },
                            "output_field": "page_html",
                            "description": "Generate content for each page"
                        }
                    }
                },
                "next_step": "assemble_site",
                "output_field": "all_pages",
                "description": "Generate all pages sequentially"
            },

            "assemble_site": {
                "action": "assemble_multipage_site",
                "config": {
                    "pages_field": "all_pages",
                    "add_navigation": true,
                    "generate_standard_pages": true
                },
                "next_step": "deploy",
                "output_field": "site_files",
                "description": "Assemble pages into complete site with navigation"
            },

            "deploy": {
                "action": "call_agent",
                "config": {
                    "agent_type": "deployer-agent",
                    "target_role": "deployer",
                    "input_fields": ["site_files", "input_data"],
                    "timeout_seconds": 180
                },
                "next_step": "complete",
                "output_field": "deployment_result",
                "description": "Deploy site to git repository"
            },

            "complete": {
                "action": "complete_workflow",
                "description": "Multipage site build complete"
            }
        }
    }
}'::jsonb,
updated_at = now()
WHERE type = 'multipage-website-builder';

-- Verify the update
SELECT
    type,
    default_config->'workflow'->'steps'->'call_strategist'->>'output_field' as strategist_output,
    default_config->'workflow'->'steps'->'generate_pages_loop'->'config'->>'iterate_over' as loop_path
FROM agent_definitions
WHERE type = 'multipage-website-builder';


---

separate html builder step



-- Updated multipage-website-builder workflow with html-developer
-- This adds HTML conversion step in the loop to fix the "loop_name.html" issue

UPDATE agent_definitions
SET default_config = '{
    "processing_mode": "task",
    "timeout_seconds": 600,
    "workflow": {
            "start_step": "spawn_strategist",
            "steps": {
                "spawn_strategist": {
                    "action": "spawn_agent",
                    "config": {
                        "agent_type": "chief-strategist",
                        "role": "strategist"
                    },
                    "next_step": "spawn_content_creator",
                    "output_field": "strategist_info",
                    "description": "Spawn chief strategist agent"
                },
                "spawn_content_creator": {
                    "action": "spawn_agent",
                    "config": {
                        "agent_type": "content-creator",
                        "role": "writer"
                    },
                    "next_step": "spawn_html_developer",
                    "output_field": "writer_info",
                    "description": "Spawn content creator agent for loop iterations"
                },
                "spawn_html_developer": {
                    "action": "spawn_agent",
                    "config": {
                        "agent_type": "html-developer",
                        "role": "developer"
                    },
                    "next_step": "spawn_deployer",
                    "output_field": "developer_info",
                    "description": "Spawn HTML developer to convert content to pages"
                },
                "spawn_deployer": {
                    "action": "spawn_agent",
                    "config": {
                        "agent_type": "deployer-agent",
                        "role": "deployer"
                    },
                    "next_step": "call_strategist",
                    "output_field": "deployer_info",
                    "description": "Spawn deployer agent"
                },
                "call_strategist": {
                    "action": "call_agent",
                    "config": {
                        "agent_type": "chief-strategist",
                        "target_role": "strategist",
                        "timeout_seconds": 120
                    },
                    "next_step": "generate_pages_loop",
                    "output_field": "page_plan",
                    "description": "Get page plan from chief strategist"
                },
                "generate_pages_loop": {
                    "action": "loop",
                    "config": {
                        "iterate_over": "page_plan.plan_data.sections",
                        "loop_var": "current_page",
                        "max_iterations": 10,
                        "substeps": {
                            "generate_content": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "content-creator",
                                    "target_role": "writer",
                                    "input_fields": ["current_page", "input_data"],
                                    "timeout_seconds": 180
                                },
                                "next_step": "create_html",
                                "output_field": "page_content",
                                "description": "Generate content strategy/copy for page"
                            },
                            "create_html": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "html-developer",
                                    "target_role": "developer",
                                    "input_fields": ["page_content", "current_page", "input_data"],
                                    "timeout_seconds": 180
                                },
                                "output_field": "page_html",
                                "description": "Convert content to professional HTML page"
                            }
                        }
                    },
                    "next_step": "assemble_site",
                    "output_field": "all_pages",
                    "description": "Generate all pages with content and HTML conversion"
                },
                "assemble_site": {
                    "action": "assemble_multipage_site",
                    "config": {
                        "pages_field": "all_pages",
                        "add_navigation": true,
                        "generate_standard_pages": true
                    },
                    "next_step": "deploy",
                    "output_field": "site_files",
                    "description": "Assemble pages into complete site with navigation"
                },
                "deploy": {
                    "action": "call_agent",
                    "config": {
                        "agent_type": "deployer-agent",
                        "target_role": "deployer",
                        "input_fields": ["site_files", "input_data"],
                        "timeout_seconds": 180
                    },
                    "next_step": "complete",
                    "output_field": "deployment_result",
                    "description": "Deploy site to git repository"
                },
                "complete": {
                    "action": "complete_workflow",
                    "description": "Multipage site build complete"
                }
            }
        }
    }'::jsonb,
    updated_at = NOW()
WHERE type = 'multipage-website-builder';

-- Verify the update
SELECT
    type,
    jsonb_pretty(default_config->'workflow'->'steps'->'generate_pages_loop') as loop_config
FROM agent_definitions
WHERE type = 'multipage-website-builder';