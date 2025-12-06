UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "spawn_domain_analyst",
                "steps": {
                    "spawn_domain_analyst": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "domain-analyst",
                            "role": "analyst"
                        },
                        "output_field": "spawned_analyst",
                        "next_step": "analyze_domain",
                        "description": "Spawn domain analyst agent"
                    },
                    "analyze_domain": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "analyst",
                            "input_fields": ["input_data"],
                            "timeout_seconds": 120
                        },
                        "output_field": "domain_analysis",
                        "next_step": "spawn_site_architect",
                        "description": "Analyze the domain name and objective"
                    },
                    "spawn_site_architect": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "site-architect",
                            "role": "architect"
                        },
                        "output_field": "spawned_architect",
                        "next_step": "design_architecture",
                        "description": "Spawn site architect agent"
                    },
                    "design_architecture": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "architect",
                            "input_fields": ["input_data", "domain_analysis"],
                            "timeout_seconds": 180
                        },
                        "output_field": "site_architecture",
                        "next_step": "spawn_content_creator",
                        "description": "Design the site structure and components"
                    },
                    "spawn_content_creator": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "content-creator",
                            "role": "content"
                        },
                        "output_field": "spawned_content",
                        "next_step": "create_content",
                        "description": "Spawn content creator agent"
                    },
                    "create_content": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "content",
                            "input_fields": ["input_data", "domain_analysis", "site_architecture"],
                            "timeout_seconds": 300
                        },
                        "output_field": "site_content",
                        "next_step": "spawn_html_developer",
                        "description": "Create content for all site sections"
                    },
                    "spawn_html_developer": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "html-developer",
                            "role": "developer"
                        },
                        "output_field": "spawned_developer",
                        "next_step": "develop_site",
                        "description": "Spawn HTML developer agent"
                    },
                    "develop_site": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "developer",
                            "input_fields": ["input_data", "site_architecture", "site_content"],
                            "timeout_seconds": 300
                        },
                        "output_field": "developed_site",
                        "next_step": "complete",
                        "description": "Develop the HTML/CSS for the site"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Website build complete"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'website-builder';


-----
spawn all agents first then call them

-- ============================================================================
-- FIX: Update website-builder workflow to spawn agents before calling them
-- ============================================================================
--
-- Current workflow: spawn/call interleaved (race condition)
-- Fixed workflow: spawn ALL agents first, then call them in sequence
--
-- Flow: spawn_analyst → spawn_architect → spawn_content → spawn_developer →
--       call_analyst → call_architect → call_content → call_developer → complete
-- ============================================================================

BEGIN;

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "spawn_domain_analyst",
                "steps": {
                    "spawn_domain_analyst": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "domain-analyst",
                            "role": "analyst"
                        },
                        "output_field": "spawned_analyst",
                        "next_step": "spawn_site_architect",
                        "description": "Spawn domain analyst agent"
                    },
                    "spawn_site_architect": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "site-architect",
                            "role": "architect"
                        },
                        "output_field": "spawned_architect",
                        "next_step": "spawn_content_creator",
                        "description": "Spawn site architect agent"
                    },
                    "spawn_content_creator": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "content-creator",
                            "role": "content"
                        },
                        "output_field": "spawned_content",
                        "next_step": "spawn_html_developer",
                        "description": "Spawn content creator agent"
                    },
                    "spawn_html_developer": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "html-developer",
                            "role": "developer"
                        },
                        "output_field": "spawned_developer",
                        "next_step": "analyze_domain",
                        "description": "Spawn HTML developer agent"
                    },
                    "analyze_domain": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "analyst",
                            "input_fields": ["input_data"],
                            "timeout_seconds": 120
                        },
                        "output_field": "domain_analysis",
                        "next_step": "design_architecture",
                        "description": "Analyze the domain name and objective"
                    },
                    "design_architecture": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "architect",
                            "input_fields": ["input_data", "domain_analysis"],
                            "timeout_seconds": 180
                        },
                        "output_field": "site_architecture",
                        "next_step": "create_content",
                        "description": "Design the site structure and components"
                    },
                    "create_content": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "content",
                            "input_fields": ["input_data", "domain_analysis", "site_architecture"],
                            "timeout_seconds": 300
                        },
                        "output_field": "site_content",
                        "next_step": "develop_site",
                        "description": "Create content for all site sections"
                    },
                    "develop_site": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "developer",
                            "input_fields": ["input_data", "site_architecture", "site_content"],
                            "timeout_seconds": 300
                        },
                        "output_field": "developed_site",
                        "next_step": "complete",
                        "description": "Develop the HTML/CSS for the site"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Website build complete"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'website-builder';

-- Verify the update
DO $$
DECLARE
start_step TEXT;
    spawn_analyst_next TEXT;
    spawn_developer_next TEXT;
BEGIN
SELECT
    default_config->'workflow'->>'start_step',
    default_config->'workflow'->'steps'->'spawn_domain_analyst'->>'next_step',
    default_config->'workflow'->'steps'->'spawn_html_developer'->>'next_step'
INTO start_step, spawn_analyst_next, spawn_developer_next
FROM agent_definitions
WHERE type = 'website-builder';

IF start_step != 'spawn_domain_analyst' THEN
        RAISE EXCEPTION 'start_step not updated. Current: %', start_step;
END IF;

    -- Verify spawn_analyst goes to spawn_architect (not analyze_domain)
    IF spawn_analyst_next != 'spawn_site_architect' THEN
        RAISE EXCEPTION 'spawn_domain_analyst should go to spawn_site_architect, got: %', spawn_analyst_next;
END IF;

    -- Verify spawn_developer goes to analyze_domain (first call step)
    IF spawn_developer_next != 'analyze_domain' THEN
        RAISE EXCEPTION 'spawn_html_developer should go to analyze_domain, got: %', spawn_developer_next;
END IF;

    RAISE NOTICE 'website-builder workflow updated: spawn ALL first, then call in sequence';
END $$;

COMMIT;

--
add multipage-wrapper and site deployer
UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
                "start_step": "spawn_domain_analyst",
                "steps": {
                    "spawn_domain_analyst": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "domain-analyst",
                            "role": "analyst"
                        },
                        "output_field": "spawned_analyst",
                        "next_step": "spawn_site_architect",
                        "description": "Spawn domain analyst agent"
                    },
                    "spawn_site_architect": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "site-architect",
                            "role": "architect"
                        },
                        "output_field": "spawned_architect",
                        "next_step": "spawn_content_creator",
                        "description": "Spawn site architect agent"
                    },
                    "spawn_content_creator": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "content-creator",
                            "role": "content"
                        },
                        "output_field": "spawned_content",
                        "next_step": "spawn_html_developer",
                        "description": "Spawn content creator agent"
                    },
                    "spawn_html_developer": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "html-developer",
                            "role": "developer"
                        },
                        "output_field": "spawned_developer",
                        "next_step": "spawn_wrapper",
                        "description": "Spawn HTML developer agent"
                    },
                    "spawn_wrapper": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "multipage-wrapper",
                            "role": "wrapper"
                        },
                        "output_field": "spawned_wrapper",
                        "next_step": "spawn_deployer",
                        "description": "Spawn multipage wrapper agent"
                    },
                    "spawn_deployer": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "site-deployer",
                            "role": "deployer"
                        },
                        "output_field": "spawned_deployer",
                        "next_step": "analyze_domain",
                        "description": "Spawn site deployer agent"
                    },
                    "analyze_domain": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "analyst",
                            "input_fields": ["input_data"],
                            "timeout_seconds": 120
                        },
                        "output_field": "domain_analysis",
                        "next_step": "design_architecture",
                        "description": "Analyze the domain name and objective"
                    },
                    "design_architecture": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "architect",
                            "input_fields": ["input_data", "domain_analysis"],
                            "timeout_seconds": 180
                        },
                        "output_field": "site_architecture",
                        "next_step": "create_content",
                        "description": "Design the site structure and components"
                    },
                    "create_content": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "content",
                            "input_fields": ["input_data", "domain_analysis", "site_architecture"],
                            "timeout_seconds": 300
                        },
                        "output_field": "site_content",
                        "next_step": "develop_site",
                        "description": "Create content for all site sections"
                    },
                    "develop_site": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "developer",
                            "input_fields": ["input_data", "site_architecture", "site_content"],
                            "timeout_seconds": 300
                        },
                        "output_field": "developed_html",
                        "next_step": "wrap_multipage",
                        "description": "Develop the HTML/CSS for the site"
                    },
                    "wrap_multipage": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "wrapper",
                            "input_fields": ["developed_html", "input_data"],
                            "timeout_seconds": 60
                        },
                        "output_field": "site_files",
                        "next_step": "deploy_site",
                        "description": "Create about and contact pages, package as files map"
                    },
                    "deploy_site": {
                        "action": "call_agent",
                        "config": {
                            "target_role": "deployer",
                            "input_fields": ["site_files", "input_data"],
                            "timeout_seconds": 180
                        },
                        "output_field": "deployment_result",
                        "next_step": "complete",
                        "description": "Deploy site to hosting"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Website build complete"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'website-builder';


--

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,develop_site}',
            '{
                "action": "call_agent",
                "config": {
                    "target_role": "developer",
                    "input_fields": ["input_data", "site_architecture", "site_content"],
                    "timeout_seconds": 300
                },
                "output_field": "final_html",
                "next_step": "wrap_multipage",
                "description": "Develop the HTML/CSS for the site"
            }'::jsonb
                     )
WHERE type = 'website-builder';


UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,wrap_multipage}',
            '{
                "action": "call_agent",
                "config": {
                    "target_role": "wrapper",
                    "input_fields": ["final_html", "input_data"],
                    "timeout_seconds": 60
                },
                "output_field": "site_files",
                "next_step": "deploy_site",
                "description": "Create about and contact pages, package as files map"
            }'::jsonb
                     )
WHERE type = 'website-builder';


## Suggested Data Contract Standards

Create a convention for common outputs:

Standard Output Fields:
- domain_analysis    → Analysis of domain/objective
- site_architecture  → Site structure/components
- site_content       → Content JSON/data
- final_html         → Assembled HTML (single page)
- site_files         → Map of files (multipage)
- deployment_result  → Deployment status/URLs

                     --

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,wrap_multipage}',
            '{
                "action": "call_agent",
                "config": {
                    "target_role": "wrapper",
                    "input_fields": ["final_html", "input_data"],
                    "timeout_seconds": 60
                },
                "output_field": "site_files",
                "next_step": "deploy_site",
                "description": "Create about and contact pages, package as files map"
            }'::jsonb
                     )
WHERE type = 'website-builder';

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,wrap_multipage,config,index_html_field}',
            '"developed_html.html_result"'::jsonb
                     )
WHERE type = 'website-builder';


----
changing it back again
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,wrap_multipage,config,index_html_field}',
        '"final_html.html_result"'::jsonb
                     )
WHERE type = 'website-builder';