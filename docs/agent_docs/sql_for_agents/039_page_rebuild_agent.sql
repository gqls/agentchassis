-- ensure_site_record → site_record
-- load_site_for_rebuild → rebuild_context (NEW — the supplementary data)
-- select_style_collection → style_collection
-- get_pages_to_build → pages_to_build (filtered to needs_rebuild only)
-- check_has_pages → spawn agents → build_pages_loop → deploy → complete

The only change in the build loop vs pageflow-builder is the input_mapping:

"reviewed_brief": "rebuild_context.reviewed_brief" (instead of "input_data.reviewed_brief")
"db_sync": "rebuild_context.db_sync" (instead of "db_sync")
"site_plan": "rebuild_context.site_plan" (instead of "site_plan")


-- ============================================================================
-- AGENT: page-rebuild
-- PURPOSE: Rebuild specific pages on an existing site without re-planning.
--
-- DESIGN PRINCIPLES (from 007_checklist):
--   - Agent owns its domain: loads all context from DB given just a domain
--   - Spawnable, not standalone-messaged: called via generic-agent or future triage
--   - Reuse before creating: uses ensure_site_record, select_style_collection,
--     get_pages_to_build as-is. One new action (load_site_for_rebuild) fills gaps.
--   - Workflows simple, complexity in Go: load action does the heavy lifting
--   - Triage-queue-ready: accepts optional task_id for future maintenance queue
--
-- WHAT IT SKIPS (vs pageflow-builder):
--   - No site planner (avoids regenerating page plan)
--   - No sync_pages_to_db (avoids flipping deployed → needs_rebuild)
--   - No asset generation (logo, hero already deployed)
--   - No set_default_components (already set during original build)
--   - No render_site_components (header/footer/head already in site_components)
--   - No apply_site_design/CSS (stylesheet already deployed)
--   - No populate_nav (nav tables already populated)
--
-- WHAT IT REUSES:
--   - ensure_site_record → loads site record + content_data from DB
--   - load_site_for_rebuild (NEW) → extracts brief, nav, pages for link context
--   - select_style_collection → loads existing style collection
--   - get_pages_to_build → queries pages with build_status = 'needs_rebuild'
--   - page-content-writer → writes content (same agent as pageflow-builder uses)
--   - content-reviewer → reviews content
--   - git_commit, save_page_sections, update_page_status → deploy pipeline
--   - deployer-agent → Cloudflare deployment
--
-- USAGE:
--   1. Flag pages:
--      UPDATE pages SET build_status = 'needs_rebuild'
--      WHERE name IN ('use-cases','privacy','terms')
--        AND site_id = (SELECT id FROM sites WHERE domain = 'example.com');
--
--   2. Trigger via generic-agent (see caller workflow below) or future triage agent
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category, default_config,
    is_active, capabilities, image_repository, image_tag,
    topics, health_config, env_vars, version,
    delegation_preferences, agent_category, status, domain_tags,
    input_contract, output_contract
) VALUES (
    'page-rebuild',
    'Page Rebuild Agent',
    'Rebuilds specific pages flagged as needs_rebuild on an existing site. Loads site context from DB, generates fresh content, deploys. Skips planning, asset generation, and CSS. Used for maintenance: fixing stale pages, adding new pages, refreshing content.',
    'specialist',
    '{
        "workflow": {
            "start_step": "ensure_site_record",
            "processing_mode": "orchestrator",
            "timeout_seconds": 900,
            "steps": {

                "ensure_site_record": {
                    "action": "ensure_site_record",
                    "config": {},
                    "output_field": "site_record",
                    "next_step": "load_rebuild_context",
                    "description": "Load existing site record from database"
                },

                "load_rebuild_context": {
                    "action": "load_site_for_rebuild",
                    "config": {
                        "site_id_field": "site_record.site_id"
                    },
                    "output_field": "rebuild_context",
                    "next_step": "select_style_collection",
                    "description": "Load reviewed brief, navigation, pages list, and brand assets from DB"
                },

                "select_style_collection": {
                    "action": "select_style_collection",
                    "config": {
                        "site_id_field": "site_record.site_id",
                        "fallback_by_domain": true
                    },
                    "output_field": "style_collection",
                    "next_step": "get_pages_to_rebuild",
                    "description": "Load existing style collection for site"
                },

                "get_pages_to_rebuild": {
                    "action": "get_pages_to_build",
                    "config": {
                        "include_all": false,
                        "build_statuses": ["needs_rebuild"]
                    },
                    "output_field": "pages_to_build",
                    "next_step": "check_has_pages",
                    "description": "Get only pages flagged as needs_rebuild"
                },

                "check_has_pages": {
                    "action": "conditional",
                    "config": {
                        "condition": "pages_to_build.page_count > 0",
                        "then_step": "spawn_content_writer",
                        "else_step": "complete_no_pages"
                    },
                    "description": "Skip if no pages need rebuilding"
                },

                "complete_no_pages": {
                    "action": "complete_workflow",
                    "config": {
                        "output_fields": ["pages_to_build"],
                        "success_message": "No pages flagged as needs_rebuild"
                    },
                    "description": "Complete early — nothing to rebuild"
                },

                "spawn_content_writer": {
                    "action": "spawn_agent",
                    "config": {
                        "role": "content_writer",
                        "agent_type": "page-content-writer"
                    },
                    "output_field": "content_writer_agent",
                    "next_step": "spawn_reviewer",
                    "description": "Spawn content writer agent"
                },

                "spawn_reviewer": {
                    "action": "spawn_agent",
                    "config": {
                        "role": "reviewer",
                        "agent_type": "content-reviewer"
                    },
                    "output_field": "reviewer_agent",
                    "next_step": "spawn_deployer",
                    "description": "Spawn content reviewer agent"
                },

                "spawn_deployer": {
                    "action": "spawn_agent",
                    "config": {
                        "role": "deployer",
                        "agent_type": "deployer-agent"
                    },
                    "output_field": "deployer_agent",
                    "next_step": "build_pages_loop",
                    "description": "Spawn deployer agent"
                },

                "build_pages_loop": {
                    "action": "loop",
                    "config": {
                        "mode": "sequential",
                        "items_field": "pages_to_build.pages",
                        "item_variable": "current_page",
                        "max_iterations": 20,
                        "sub_workflow": {
                            "start_step": "write_page_content",
                            "steps": {
                                "write_page_content": {
                                    "action": "call_agent",
                                    "config": {
                                        "agent_type": "page-content-writer",
                                        "target_role": "content_writer",
                                        "input_mapping": {
                                            "current_page": "current_page",
                                            "site_record": "site_record",
                                            "reviewed_brief": "rebuild_context.reviewed_brief",
                                            "style_collection": "style_collection",
                                            "db_sync": "rebuild_context.db_sync",
                                            "site_plan": "rebuild_context.site_plan",
                                            "logo_url": "rebuild_context.logo_url",
                                            "hero_url": "rebuild_context.hero_url"
                                        },
                                        "timeout_seconds": 300
                                    },
                                    "output_field": "page_content",
                                    "next_step": "review_page_content",
                                    "description": "Write content for this page"
                                },

                                "review_page_content": {
                                    "action": "call_agent",
                                    "config": {
                                        "agent_type": "content-reviewer",
                                        "target_role": "reviewer",
                                        "input_mapping": {
                                            "current_page": "current_page",
                                            "site_record": "site_record",
                                            "page_content": "page_content",
                                            "reviewed_brief": "rebuild_context.reviewed_brief"
                                        },
                                        "timeout_seconds": 3900
                                    },
                                    "output_field": "reviewed_content",
                                    "next_step": "check_review_approved",
                                    "description": "Review page content"
                                },

                                "check_review_approved": {
                                    "action": "conditional",
                                    "config": {
                                        "condition": "reviewed_content.review_result.approved == true OR reviewed_content.approved == true",
                                        "then_step": "assemble_page",
                                        "else_step": "complete_page"
                                    },
                                    "description": "Check if content was approved"
                                },

                                "assemble_page": {
                                    "action": "assemble_page",
                                    "config": {
                                        "inject_head": true,
                                        "content_field": "page_content.response.page_html",
                                        "add_navigation": false
                                    },
                                    "output_field": "assembled_page",
                                    "next_step": "deploy_page",
                                    "description": "Assemble full page HTML from components"
                                },

                                "deploy_page": {
                                    "action": "git_commit",
                                    "config": {
                                        "page_field": "current_page",
                                        "domain_field": "site_record.domain",
                                        "content_field": "assembled_page.html"
                                    },
                                    "output_field": "page_deployed",
                                    "next_step": "save_sections",
                                    "description": "Commit page to git"
                                },

                                "save_sections": {
                                    "action": "save_page_sections",
                                    "config": {
                                        "html_field": "assembled_page.html",
                                        "site_id_field": "site_record.site_id",
                                        "page_name_field": "current_page.name"
                                    },
                                    "output_field": "save_result",
                                    "next_step": "update_page_status",
                                    "description": "Save sections to page_components for future rerender"
                                },

                                "update_page_status": {
                                    "action": "update_page_status",
                                    "config": {
                                        "status": "deployed",
                                        "commit_from": "page_deployed.commit_sha",
                                        "page_id_field": "current_page.id"
                                    },
                                    "next_step": "complete_page",
                                    "description": "Mark page as deployed"
                                },

                                "complete_page": {
                                    "action": "loop_complete",
                                    "description": "Page rebuild complete"
                                }
                            }
                        }
                    },
                    "output_field": "pages_built",
                    "next_step": "trigger_site_deploy",
                    "description": "Rebuild each flagged page: write → review → assemble → deploy"
                },

                "trigger_site_deploy": {
                    "action": "call_agent",
                    "config": {
                        "agent_type": "deployer-agent",
                        "target_role": "deployer",
                        "input_mapping": {
                            "pages_built": "pages_built",
                            "site_record": "site_record"
                        },
                        "timeout_seconds": 180
                    },
                    "output_field": "deployment_result",
                    "next_step": "complete",
                    "description": "Trigger Cloudflare deployment"
                },

                "complete": {
                    "action": "complete_workflow",
                    "config": {
                        "output_fields": ["site_record", "pages_built", "deployment_result"]
                    },
                    "description": "Page rebuild complete"
                }
            }
        }
    }'::jsonb,
    true,                       -- is_active
    '[]'::jsonb,               -- capabilities
    'docker.io/aqls/agent-chassis',
    'v1.0.770',
    '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
    '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
    '[]'::jsonb,               -- env_vars
    1,                         -- version
    '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
    'specialist',              -- agent_category (spawnable specialist)
    'active',                  -- status
    '["maintenance", "rebuild", "pages"]'::jsonb,
    '{"required": ["domain"], "optional": ["site_id", "task_id"], "description": "Provide domain (or site_id). Pages must be pre-flagged as needs_rebuild in the pages table. Optional task_id for maintenance queue tracking."}'::jsonb,
    '{"produces": {"pages_built": "Pages rebuilt and deployed", "deployment_result": "Cloudflare deployment result"}}'::jsonb
)
ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       default_config = EXCLUDED.default_config,
                                       image_tag = EXCLUDED.image_tag,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       status = EXCLUDED.status,
                                       updated_at = NOW();

-- Verify the agent definition
SELECT type, display_name, status, agent_category
FROM agent_definitions
WHERE type = 'page-rebuild';


-- ============================================================================
-- CALLER WORKFLOW for generic-agent
-- Use this to spawn and call page-rebuild from the generic agent.
--
-- Send to: system.agent.generic-agent.process
-- With workflow_override containing this workflow.
--
-- Example Kafka message:
-- {
--   "message_type": "request",
--   "action": "process",
--   "client_id": "demo_client",
--   "input_data": {
--     "domain": "leopardessconsulting.co.uk"
--   },
--   "workflow_override": { <this workflow> }
-- }
-- ============================================================================

-- For reference / copy-paste, the caller workflow is:
/*
{
    "start_step": "spawn_rebuilder",
    "processing_mode": "orchestrator",
    "timeout_seconds": 1200,
    "steps": {
        "spawn_rebuilder": {
            "action": "spawn_agent",
            "config": {
                "role": "rebuilder",
                "agent_type": "page-rebuild"
            },
            "output_field": "rebuilder_agent",
            "next_step": "call_rebuilder",
            "description": "Spawn page-rebuild agent"
        },
        "call_rebuilder": {
            "action": "call_agent",
            "config": {
                "agent_type": "page-rebuild",
                "target_role": "rebuilder",
                "input_mapping": {
                    "domain": "input_data.domain",
                    "site_id": "input_data.site_id",
                    "task_id": "input_data.task_id"
                },
                "timeout_seconds": 900
            },
            "output_field": "rebuild_result",
            "next_step": "complete",
            "description": "Call page-rebuild with domain"
        },
        "complete": {
            "action": "complete_workflow",
            "config": {
                "output_fields": ["rebuild_result"]
            },
            "description": "Rebuild dispatch complete"
        }
    }
}
*/

