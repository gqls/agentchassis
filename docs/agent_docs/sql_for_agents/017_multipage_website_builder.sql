-- ============================================================================
-- VERSION 2 AGENTS - Unified Site Builder Architecture
-- ============================================================================
-- These are v2 agents that work alongside existing v1 agents.
-- v1 agents continue to work as before.
-- v2 agents use the new pages/components structure.
--
-- To use v2: reference agent_type with version, or update workflows to use v2
-- ============================================================================

-- ============================================================================
-- 3. MULTIPAGE-WEBSITE-BUILDER-V2 - Uses pages array loop
-- ============================================================================
INSERT INTO agent_definitions (
    id,
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    version,
    input_contract,
    output_contract
)
SELECT
    gen_random_uuid(),
    'multipage-website-builder',
    'Multipage Website Builder V2',
    'Builds websites using unified pages/components architecture (v2)',
    category,
    jsonb_build_object(
            'workflow', jsonb_build_object(
            'start_step', 'spawn_strategist',
            'steps', jsonb_build_object(
                    'spawn_strategist', jsonb_build_object(
                            'action', 'spawn_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'chief-strategist',
                                    'agent_version', 2,
                                    'role', 'strategist'
                                      ),
                            'next_step', 'spawn_content_creator',
                            'output_field', 'strategist_info',
                            'description', 'Spawn v2 strategist'
                                        ),
                    'spawn_content_creator', jsonb_build_object(
                            'action', 'spawn_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'content-creator',
                                    'agent_version', 2,
                                    'role', 'writer'
                                      ),
                            'next_step', 'spawn_html_developer',
                            'output_field', 'writer_info',
                            'description', 'Spawn v2 content creator'
                                             ),
                    'spawn_html_developer', jsonb_build_object(
                            'action', 'spawn_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'html-developer',
                                    'role', 'developer'
                                      ),
                            'next_step', 'spawn_deployer',
                            'output_field', 'developer_info',
                            'description', 'Spawn HTML developer'
                                            ),
                    'spawn_deployer', jsonb_build_object(
                            'action', 'spawn_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'deployer-agent',
                                    'role', 'deployer'
                                      ),
                            'next_step', 'call_strategist',
                            'output_field', 'deployer_info',
                            'description', 'Spawn deployer'
                                      ),
                    'call_strategist', jsonb_build_object(
                            'action', 'call_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'chief-strategist',
                                    'target_role', 'strategist',
                                    'timeout_seconds', 120
                                      ),
                            'next_step', 'generate_pages_loop',
                            'output_field', 'page_plan',
                            'description', 'Get page plan from v2 strategist'
                                       ),
                    'generate_pages_loop', jsonb_build_object(
                            'action', 'loop',
                            'config', jsonb_build_object(
                                    'iterate_over', 'page_plan.plan_data.pages',
                                    'loop_var', 'current_page',
                                    'max_iterations', 10,
                                    'substeps', jsonb_build_object(
                                            'generate_content', jsonb_build_object(
                                                    'action', 'call_agent',
                                                    'config', jsonb_build_object(
                                                            'agent_type', 'content-creator',
                                                            'target_role', 'writer',
                                                            'input_fields', jsonb_build_array('current_page', 'input_data', 'page_plan'),
                                                            'timeout_seconds', 180
                                                              ),
                                                    'next_step', 'create_html',
                                                    'output_field', 'page_content',
                                                    'description', 'Generate content for page'
                                                                ),
                                            'create_html', jsonb_build_object(
                                                    'action', 'call_agent',
                                                    'config', jsonb_build_object(
                                                            'agent_type', 'html-developer',
                                                            'target_role', 'developer',
                                                            'input_fields', jsonb_build_array('page_content', 'current_page', 'input_data', 'page_plan'),
                                                            'timeout_seconds', 180
                                                              ),
                                                    'output_field', 'page_html',
                                                    'description', 'Convert content to HTML'
                                                           )
                                                )
                                      ),
                            'next_step', 'assemble_site',
                            'output_field', 'all_pages',
                            'description', 'Generate all pages'
                                           ),
                    'assemble_site', jsonb_build_object(
                            'action', 'assemble_multipage_site',
                            'config', jsonb_build_object(
                                    'pages_field', 'all_pages',
                                    'add_navigation', true,
                                    'generate_standard_pages', false
                                      ),
                            'next_step', 'deploy',
                            'output_field', 'site_files',
                            'description', 'Assemble pages with navigation'
                                     ),
                    'deploy', jsonb_build_object(
                            'action', 'call_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'deployer-agent',
                                    'target_role', 'deployer',
                                    'input_fields', jsonb_build_array('site_files', 'input_data'),
                                    'timeout_seconds', 180
                                      ),
                            'next_step', 'complete',
                            'output_field', 'deployment_result',
                            'description', 'Deploy to repository'
                              ),
                    'complete', jsonb_build_object(
                            'action', 'complete_workflow',
                            'description', 'Build complete'
                                )
                     )
                        ),
            'processing_mode', 'orchestration',
            'timeout_seconds', 900
    ),
    true,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    2,  -- VERSION 2
    '{"required": ["input_data"], "expects": {"input_data.domain": "string", "input_data.objective": "string"}}'::jsonb,
    '{"produces": "deployment_result", "format": {"type": "object"}}'::jsonb
FROM agent_definitions
WHERE type = 'multipage-website-builder' AND version = 1
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       display_name = EXCLUDED.display_name,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();


-- ============================================================================
-- VERIFICATION
-- ============================================================================
SELECT
    type,
    version,
    display_name,
    substring(description from 1 for 60) as desc_preview
FROM agent_definitions
WHERE type IN ('chief-strategist', 'content-creator', 'multipage-website-builder')
ORDER BY type, version;


-- ============================================================================
-- Update multipage-website-builder to iterate over PAGES not SECTIONS
-- ============================================================================
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_pages_loop,config,iterate_over}',
        '"page_plan.plan_data.pages"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'multipage-website-builder'
  AND is_active = true;


-- ============================================================================
-- SITEMAP-ENABLED NAVIGATION: Complete Update
-- ============================================================================
-- 1. chief-strategist: Output pages + sitemap structure
-- 2. multipage-website-builder: Pass page_plan to html-developer
-- ============================================================================

-- ============================================================================
-- 2. UPDATE MULTIPAGE-WEBSITE-BUILDER
--    - iterate_over: pages (not sections)
--    - Pass page_plan to html-developer for sitemap access
-- ============================================================================

UPDATE agent_definitions
SET
    updated_at = NOW(),
    default_config = jsonb_set(
            default_config,
            '{workflow}',
            '{
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
                        "description": "Spawn strategist for page planning"
                    },
                    "spawn_content_creator": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "content-creator",
                            "role": "writer"
                        },
                        "next_step": "spawn_html_developer",
                        "output_field": "writer_info",
                        "description": "Spawn content creator"
                    },
                    "spawn_html_developer": {
                        "action": "spawn_agent",
                        "config": {
                            "agent_type": "html-developer",
                            "role": "developer"
                        },
                        "next_step": "spawn_deployer",
                        "output_field": "developer_info",
                        "description": "Spawn HTML developer"
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
                            "input_fields": ["input_data"],
                            "timeout_seconds": 120
                        },
                        "next_step": "generate_pages_loop",
                        "output_field": "page_plan",
                        "description": "Get page plan with sitemap from strategist"
                    },
                    "generate_pages_loop": {
                        "action": "loop",
                        "config": {
                            "iterate_over": "page_plan.plan_data.pages",
                            "loop_var": "current_page",
                            "max_iterations": 10,
                            "substeps": {
                                "generate_content": {
                                    "action": "call_agent",
                                    "config": {
                                        "agent_type": "content-creator",
                                        "target_role": "writer",
                                        "input_fields": ["current_page", "input_data", "page_plan"],
                                        "timeout_seconds": 180
                                    },
                                    "next_step": "create_html",
                                    "output_field": "page_content",
                                    "description": "Generate content for page"
                                },
                                "create_html": {
                                    "action": "call_agent",
                                    "config": {
                                        "agent_type": "html-developer",
                                        "target_role": "developer",
                                        "input_fields": ["page_content", "current_page", "input_data", "page_plan"],
                                        "timeout_seconds": 180
                                    },
                                    "output_field": "page_html",
                                    "description": "Convert content to HTML with navigation from sitemap"
                                }
                            }
                        },
                        "next_step": "assemble_site",
                        "output_field": "all_pages",
                        "description": "Generate all pages with content and HTML"
                    },
                    "assemble_site": {
                        "action": "assemble_multipage_site",
                        "config": {
                            "pages_field": "all_pages",
                            "sitemap_field": "page_plan.plan_data.sitemap",
                            "add_navigation": true,
                            "generate_standard_pages": false
                        },
                        "next_step": "deploy",
                        "output_field": "site_files",
                        "description": "Assemble pages with sitemap navigation"
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
                        "description": "Deploy site to repository"
                    },
                    "complete": {
                        "action": "complete_workflow",
                        "description": "Multipage site build complete"
                    }
                }
            }'::jsonb
                     )
WHERE type = 'multipage-website-builder'
  AND is_active = true;


----

adding links

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow}',
        '{
            "start_step": "ensure_site",
            "steps": {
                "ensure_site": {
                    "action": "ensure_site_record",
                    "config": {},
                    "output_field": "site_record",
                    "next_step": "spawn_strategist",
                    "description": "Create or get site record in database"
                },
                "spawn_strategist": {
                    "action": "spawn_agent",
                    "config": {"agent_type": "chief-strategist", "role": "strategist"},
                    "next_step": "spawn_content_creator",
                    "output_field": "strategist_info"
                },
                "spawn_content_creator": {
                    "action": "spawn_agent",
                    "config": {"agent_type": "content-creator", "role": "writer"},
                    "next_step": "spawn_html_developer",
                    "output_field": "writer_info"
                },
                "spawn_html_developer": {
                    "action": "spawn_agent",
                    "config": {"agent_type": "html-developer", "role": "developer"},
                    "next_step": "spawn_deployer",
                    "output_field": "developer_info"
                },
                "spawn_deployer": {
                    "action": "spawn_agent",
                    "config": {"agent_type": "deployer-agent", "role": "deployer"},
                    "next_step": "call_strategist",
                    "output_field": "deployer_info"
                },
                "call_strategist": {
                    "action": "call_agent",
                    "config": {
                        "agent_type": "chief-strategist",
                        "target_role": "strategist",
                        "input_fields": ["input_data", "site_record"],
                        "timeout_seconds": 120
                    },
                    "next_step": "sync_pages",
                    "output_field": "page_plan"
                },
                "sync_pages": {
                    "action": "sync_pages_to_db",
                    "config": {},
                    "output_field": "db_sync",
                    "next_step": "generate_pages_loop",
                    "description": "Sync pages to database, build navigation"
                },
                "generate_pages_loop": {
                    "action": "loop",
                    "config": {
                        "iterate_over": "page_plan.plan_data.pages",
                        "loop_var": "current_page",
                        "max_iterations": 10,
                        "substeps": {
                            "generate_content": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "content-creator",
                                    "target_role": "writer",
                                    "input_fields": ["current_page", "input_data", "page_plan"],
                                    "timeout_seconds": 180
                                },
                                "next_step": "create_html",
                                "output_field": "page_content"
                            },
                            "create_html": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "html-developer",
                                    "target_role": "developer",
                                    "input_fields": ["page_content", "current_page", "input_data", "db_sync"],
                                    "timeout_seconds": 180
                                },
                                "next_step": "extract_links",
                                "output_field": "page_html"
                            },
                            "extract_links": {
                                "action": "extract_and_sync_links",
                                "config": {},
                                "output_field": "link_sync",
                                "description": "Extract links from HTML, sync to registry"
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
                        "include_sitemap_xml": true,
                        "include_robots_txt": true
                    },
                    "next_step": "deploy",
                    "output_field": "site_files"
                },
                "deploy": {
                    "action": "call_agent",
                    "config": {
                        "agent_type": "deployer-agent",
                        "target_role": "deployer",
                        "input_fields": ["site_files", "input_data", "site_record"],
                        "timeout_seconds": 180
                    },
                    "next_step": "update_timestamps",
                    "output_field": "deployment_result"
                },
                "update_timestamps": {
                    "action": "update_site_timestamps",
                    "config": {},
                    "next_step": "complete",
                    "description": "Update last_built_at, last_deployed_at"
                },
                "complete": {
                    "action": "complete_workflow"
                }
            }
        }'::jsonb
                     )
WHERE type = 'multipage-website-builder';

===

before change:

 d8c6adcc-7ce7-4b18-af1b-2f34db616c06 | multipage-website-builder | Multi-Page Website Builder | Builds large websites (20+ pages) using batched generation to avoid token limits | orchestrator | {"workflow": {"steps": {"deploy": {"action": "call_agent", "config": {"agent_type": "deployer-agent", "target_role": "deployer", "input_fields": ["site_files", "input_data"], "timeout_seconds": 180}, "next_step": "complete", "description": "Deploy site to repository", "output_field": "deployment_result"}, "complete": {"action": "complete_workflow", "description": "Multipage site build complete"}, "assemble_site": {"action": "assemble_multipage_site", "config": {"pages_field": "all_pages", "sitemap_field": "page_plan.plan_data.sitemap", "add_navigation": true, "generate_standard_pages": false}, "next_step": "deploy", "description": "Assemble pages with sitemap navigation", "output_field": "site_files"}, "spawn_deployer": {"action": "spawn_agent", "config": {"role": "deployer", "agent_type": "deployer-agent"}, "next_step": "call_strategist", "description": "Spawn deployer agent", "output_field": "deployer_info"}, "call_strategist": {"action": "call_agent", "config": {"agent_type": "chief-strategist", "target_role": "strategist", "input_fields": ["input_data"], "timeout_seconds": 120}, "next_step": "generate_pages_loop", "description": "Get page plan with sitemap from strategist", "output_field": "page_plan"}, "spawn_strategist": {"action": "spawn_agent", "config": {"role": "strategist", "agent_type": "chief-strategist"}, "next_step": "spawn_content_creator", "description": "Spawn strategist for page planning", "output_field": "strategist_info"}, "generate_pages_loop": {"action": "loop", "config": {"loop_var": "current_page", "substeps": {"create_html": {"action": "call_agent", "config": {"agent_type": "html-developer", "target_role": "developer", "input_fields": ["page_content", "current_page", "input_data", "page_plan"], "timeout_seconds": 180}, "description": "Convert content to HTML with navigation from sitemap", "output_field": "page_html"}, "generate_content": {"action": "call_agent", "config": {"agent_type": "content-creator", "target_role": "writer", "input_fields": ["current_page", "input_data", "page_plan"], "timeout_seconds": 180}, "next_step": "create_html", "description": "Generate content for page", "output_field": "page_content"}}, "iterate_over": "page_plan.plan_data.pages", "max_iterations": 10}, "next_step": "assemble_site", "description": "Generate all pages with content and HTML", "output_field": "all_pages"}, "spawn_html_developer": {"action": "spawn_agent", "config": {"role": "developer", "agent_type": "html-developer"}, "next_step": "spawn_deployer", "description": "Spawn HTML developer", "output_field": "developer_info"}, "spawn_content_creator": {"action": "spawn_agent", "config": {"role": "writer", "agent_type": "content-creator"}, "next_step": "spawn_html_developer", "description": "Spawn content creator", "output_field": "writer_info"}}, "start_step": "spawn_strategist"}, "timeout_seconds": 600} | t         | 2025-12-06 19:02:26.358753+00 | 2025-12-19 10:58:24.866539+00 |            | ["orchestration", "website-builder", "multi-page", "batched-generation"] | docker.io/aqls/agent-chassis | v1.0.566  |         | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true} |                | experimental | []          | {"sections": [{"name": "company_info", "title": "Company Information", "questions": [{"type": "text", "field": "company_name", "label": "Company Name", "required": true}, {"type": "textarea", "field": "about_us", "label": "About Us", "required": true}, {"type": "text", "field": "tagline", "label": "Tagline", "required": false}]}, {"name": "services", "title": "Services & Offerings", "questions": [{"type": "json_array", "field": "services", "label": "Services (list of {name, description})", "required": true}]}, {"name": "team", "title": "Team & Leadership", "questions": [{"type": "json_array", "field": "leadership_team", "label": "Leadership Team (list of {name, title, bio})", "required": false}]}, {"name": "portfolio", "title": "Case Studies & Portfolio", "questions": [{"type": "json_array", "field": "case_studies", "label": "Case Studies", "required": false}]}, {"name": "contact", "title": "Contact Information", "questions": [{"type": "text", "field": "contact_email", "label": "Email", "required": true}, {"type": "text", "field": "contact_phone", "label": "Phone", "required": false}, {"type": "text", "field": "headquarters", "label": "Location", "required": false}]}, {"name": "features", "title": "Site Features", "questions": [{"type": "boolean", "field": "has_blog", "label": "Include Blog?", "default": false}, {"type": "boolean", "field": "has_careers", "label": "Include Careers?", "default": false}]}]} |           0 | f           |                |
(1 row)

-- ============================================================================
-- MULTIPAGE WEBSITE BUILDER - Updated Workflow with Database Integration
-- ============================================================================
-- This update adds database persistence for sites, pages, and links.
--
-- New steps added:
--   1. ensure_site_record - First step, creates/gets site in database
--   2. sync_pages_to_db - After strategist, syncs pages and builds navigation
--   3. extract_and_sync_links - After html creation in loop, extracts links
--   4. update_site_timestamps - After deployment, updates timestamps
--
-- Key changes:
--   - html-developer now receives db_sync.navigation for accurate URLs
--   - Links are extracted and stored in link_registry
--   - Site/page records enable incremental updates in future
-- ============================================================================

-- First, ensure the required tables exist (run link_management_migration.sql first)

-- Update the multipage-website-builder workflow
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow}',
        '{
            "start_step": "ensure_site_record",
            "steps": {
                "ensure_site_record": {
                    "action": "ensure_site_record",
                    "config": {},
                    "next_step": "spawn_strategist",
                    "output_field": "site_record",
                    "description": "Create or retrieve site record from database"
                },
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
                        "input_fields": ["input_data", "site_record"],
                        "timeout_seconds": 120
                    },
                    "next_step": "sync_pages_to_db",
                    "output_field": "page_plan",
                    "description": "Get page plan from chief strategist"
                },
                "sync_pages_to_db": {
                    "action": "sync_pages_to_db",
                    "config": {},
                    "next_step": "generate_pages_loop",
                    "output_field": "db_sync",
                    "description": "Sync pages to database and build navigation structure"
                },
                "generate_pages_loop": {
                    "action": "loop",
                    "config": {
                        "iterate_over": "page_plan.plan_data.pages",
                        "loop_var": "current_page",
                        "max_iterations": 20,
                        "substeps": {
                            "generate_content": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "content-creator",
                                    "target_role": "writer",
                                    "input_fields": ["current_page", "input_data", "page_plan"],
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
                                    "input_fields": ["page_content", "current_page", "input_data", "db_sync", "page_plan"],
                                    "timeout_seconds": 180
                                },
                                "next_step": "extract_links",
                                "output_field": "page_html",
                                "description": "Convert content to professional HTML page"
                            },
                            "extract_links": {
                                "action": "extract_and_sync_links",
                                "config": {},
                                "output_field": "link_sync",
                                "description": "Extract links from HTML and sync to link registry"
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
                        "generate_standard_pages": true,
                        "include_sitemap_xml": true,
                        "include_robots_txt": true
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
                        "input_fields": ["site_files", "input_data", "site_record"],
                        "timeout_seconds": 180
                    },
                    "next_step": "update_timestamps",
                    "output_field": "deployment_result",
                    "description": "Deploy site to git repository"
                },
                "update_timestamps": {
                    "action": "update_site_timestamps",
                    "config": {},
                    "next_step": "complete",
                    "output_field": "timestamp_update",
                    "description": "Update site last_built_at and last_deployed_at"
                },
                "complete": {
                    "action": "complete_workflow",
                    "description": "Multipage site build complete"
                }
            }
        }'::jsonb
                     )
WHERE type = 'multipage-website-builder';

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================

-- Check the new workflow structure
SELECT
    type,
    jsonb_pretty(default_config->'workflow'->'start_step') as start_step
FROM agent_definitions
WHERE type = 'multipage-website-builder';

-- List all steps in order (approximately)
SELECT
    type,
    step_name,
    step_config->>'action' as action,
    step_config->>'next_step' as next_step,
    step_config->>'description' as description
FROM agent_definitions,
    jsonb_each(default_config->'workflow'->'steps') as steps(step_name, step_config)
WHERE type = 'multipage-website-builder'
ORDER BY
    CASE step_name
    WHEN 'ensure_site_record' THEN 1
    WHEN 'spawn_strategist' THEN 2
    WHEN 'spawn_content_creator' THEN 3
    WHEN 'spawn_html_developer' THEN 4
    WHEN 'spawn_deployer' THEN 5
    WHEN 'call_strategist' THEN 6
    WHEN 'sync_pages_to_db' THEN 7
    WHEN 'generate_pages_loop' THEN 8
    WHEN 'assemble_site' THEN 9
    WHEN 'deploy' THEN 10
    WHEN 'update_timestamps' THEN 11
    WHEN 'complete' THEN 12
    ELSE 99
END;

-- ============================================================================
-- DATA FLOW DOCUMENTATION
-- ============================================================================
/*
Data flow through the updated workflow:

1. ensure_site_record
   Input:  input_data.domain
   Output: site_record {site_id, domain, network_id, status}
   DB:     INSERT/UPDATE sites table

2. spawn_* steps
   Input:  (none)
   Output: *_info with agent reference
   DB:     (none)

3. call_strategist
   Input:  input_data, site_record
   Output: page_plan {plan_data: {pages: [...], sitemap: [...]}}
   DB:     (none, LLM only)

4. sync_pages_to_db
   Input:  site_record.site_id, page_plan
   Output: db_sync {pages_synced, navigation: {items: [...]}}
   DB:     INSERT/UPDATE pages table, navigation_structures cache

5. generate_pages_loop (per page)
   5a. generate_content
       Input:  current_page, input_data, page_plan
       Output: page_content
       DB:     (none, LLM only)

   5b. create_html
       Input:  page_content, current_page, input_data, db_sync, page_plan
       Output: page_html (uses db_sync.navigation for URLs)
       DB:     (none, LLM only)

   5c. extract_links
       Input:  page_html, site_record.site_id, current_page
       Output: link_sync {links_extracted, links_persisted}
       DB:     DELETE old links, INSERT new links to link_registry

6. assemble_site
   Input:  all_pages (array of page results)
   Output: site_files {files: {filename: html_content}}
   DB:     Can read navigation_structures for sitemap.xml

7. deploy
   Input:  site_files, input_data, site_record
   Output: deployment_result
   DB:     (none, GitHub only)

8. update_timestamps
   Input:  site_record.site_id
   Output: timestamp_update {updated: true, last_built_at, last_deployed_at}
   DB:     UPDATE sites SET last_built_at, last_deployed_at

Key change: html-developer receives db_sync containing navigation built from
the database, ensuring consistent and correct relative URLs across all pages.
*/
