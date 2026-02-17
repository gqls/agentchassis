-- =============================================================================
-- NAV-UPDATER AGENT DEFINITION
-- =============================================================================
-- Purpose: Refreshes navigation structure and propagates updated header/footer
--          to all deployed pages. Algorithmic only - no LLM calls needed.
--
-- What it does:
--   1. Loads site record
--   2. Rebuilds nav tables from current pages (populate_nav_tables)
--   3. Re-renders header/footer/head with fresh nav data
--   4. Reassembles all deployed pages from stored sections + new header/footer
--   5. Deploys via git + Cloudflare
--
-- When to use:
--   - After new pages are deployed but nav doesn't reflect them (orphan_nav)
--   - After pages are removed or renamed
--   - After nav_order or nav_label changes
--   - After header/footer component template changes
--
-- Difference from rerender-site:
--   rerender-site skips nav table refresh — it re-renders header from whatever
--   is already in site_nav_items. nav-updater adds populate_nav_tables first,
--   so the nav tables reflect the CURRENT state of the pages table.
--
-- Input:   { "domain": "example.com" }  (standalone)
--          or via input_mapping from maintenance-triage
-- Output:  { site_record, nav_refreshed, site_components_rendered,
--            pages_rerendered, deployment_result }
--
-- Reuses existing actions — NO new Go code required:
--   - ensure_site_record
--   - populate_nav_tables
--   - render_site_components
--   - get_pages_for_rerender
--   - page-rerender agent (spawned)
--   - deployer-agent (spawned)
-- =============================================================================

-- =============================================================================
-- NAV-UPDATER AGENT DEFINITION
-- =============================================================================
-- Corrected to match actual agent_definitions schema.
--
-- Column mapping from what I had wrong → actual:
--   name → display_name
--   workflow → default_config (contains the workflow JSON)
--   is_orchestrator → (not a column, implied by default_config.processing_mode)
--   semantic_tags → capabilities
--   agent_tags → domain_tags
--   resource_requirements → resources
--   health_check → health_config
--   scaling → (not a column)

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    domain_tags,
    resources,
    health_config,
    agent_category,
    status,
    input_contract
)
VALUES (
           'nav-updater',
           'Nav Updater Agent',
           'Refreshes navigation tables from current pages, re-renders header/footer components with updated nav, then reassembles and deploys all pages. Algorithmic only - no LLM calls. Used when pages are added/removed/renamed and nav needs updating across the site.',
           'maintenance',
           '{
               "processing_mode": "orchestrator",
               "timeout_seconds": 600,
               "workflow": {
                   "start_step": "ensure_site_record",
                   "steps": {
                       "ensure_site_record": {
                           "action": "ensure_site_record",
                           "config": {
                               "store_brief_in_content_data": false
                           },
                           "description": "Load existing site record from database",
                           "output_field": "site_record",
                           "next_step": "refresh_nav_tables"
                       },

                       "refresh_nav_tables": {
                           "action": "populate_nav_tables",
                           "config": {
                               "input_fields": ["site_id"],
                               "max_header_items": 8
                           },
                           "description": "Rebuild site_nav_groups and site_nav_items from current pages",
                           "output_field": "nav_refreshed",
                           "next_step": "render_site_components"
                       },

                       "render_site_components": {
                           "action": "render_site_components",
                           "config": {
                               "slots": ["header", "footer", "head"],
                               "force_rerender": true
                           },
                           "description": "Re-render header/footer/head with fresh nav data",
                           "output_field": "site_components_rendered",
                           "next_step": "get_pages"
                       },

                       "get_pages": {
                           "action": "get_pages_for_rerender",
                           "config": {
                               "include_statuses": ["deployed", "active"]
                           },
                           "description": "Get all deployed pages for reassembly",
                           "output_field": "pages_for_rerender",
                           "next_step": "check_has_pages"
                       },

                       "check_has_pages": {
                           "action": "conditional",
                           "config": {
                               "condition": "pages_for_rerender.has_pages == true",
                               "then_step": "spawn_rerenderer",
                               "else_step": "complete_no_pages"
                           },
                           "description": "Skip if no pages to process"
                       },

                       "spawn_rerenderer": {
                           "action": "spawn_agent",
                           "config": {
                               "role": "rerenderer",
                               "agent_type": "page-rerender"
                           },
                           "description": "Spawn page-rerender agent for per-page work",
                           "output_field": "rerenderer_agent",
                           "next_step": "spawn_deployer"
                       },

                       "spawn_deployer": {
                           "action": "spawn_agent",
                           "config": {
                               "role": "deployer",
                               "agent_type": "deployer-agent"
                           },
                           "description": "Spawn deployer for git + Cloudflare",
                           "output_field": "deployer_agent",
                           "next_step": "rerender_loop"
                       },

                       "rerender_loop": {
                           "action": "loop",
                           "config": {
                               "items_field": "pages_for_rerender.pages",
                               "item_variable": "current_page",
                               "mode": "sequential",
                               "max_iterations": 50,
                               "sub_workflow": {
                                   "start_step": "call_rerender",
                                   "steps": {
                                       "call_rerender": {
                                           "action": "call_agent",
                                           "config": {
                                               "agent_type": "page-rerender",
                                               "target_role": "rerenderer",
                                               "input_mapping": {
                                                   "page_id": "current_page.page_id",
                                                   "site_id": "site_record.site_id",
                                                   "domain": "site_record.domain"
                                               },
                                               "timeout_seconds": 120
                                           },
                                           "description": "Reassemble page with updated header/footer",
                                           "output_field": "page_result",
                                           "next_step": "complete_page"
                                       },
                                       "complete_page": {
                                           "action": "loop_complete",
                                           "description": "Page rerender iteration complete"
                                       }
                                   }
                               }
                           },
                           "description": "Loop through all pages, reassemble each with new header/footer",
                           "output_field": "pages_rerendered",
                           "next_step": "trigger_deploy"
                       },

                       "trigger_deploy": {
                           "action": "call_agent",
                           "config": {
                               "agent_type": "deployer-agent",
                               "target_role": "deployer",
                               "input_mapping": {
                                   "site_record": "site_record",
                                   "pages_built": "pages_rerendered"
                               },
                               "timeout_seconds": 180
                           },
                           "description": "Trigger Cloudflare deployment",
                           "output_field": "deployment_result",
                           "next_step": "complete"
                       },

                       "complete": {
                           "action": "complete_workflow",
                           "config": {
                               "output_fields": [
                                   "site_record",
                                   "nav_refreshed",
                                   "site_components_rendered",
                                   "pages_rerendered",
                                   "deployment_result"
                               ]
                           },
                           "description": "Nav update complete"
                       },

                       "complete_no_pages": {
                           "action": "complete_workflow",
                           "config": {
                               "output_fields": ["nav_refreshed", "site_components_rendered"],
                               "success_message": "Nav tables refreshed but no deployed pages to update"
                           },
                           "description": "Complete early - no pages to reassemble"
                       }
                   }
               }
           }'::jsonb,
           true,
           '["navigation", "header", "footer", "rerender"]'::jsonb,
           '["maintenance", "nav", "header", "footer"]'::jsonb,
           '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "128Mi"}}',
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}',
           'specialist',
           'active',
           '{"required": ["domain"], "optional": ["site_id"], "description": "Provide domain. Agent refreshes nav tables, re-renders header/footer, reassembles all pages, and deploys."}'
       )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       default_config = EXCLUDED.default_config,
                                       is_active = EXCLUDED.is_active,
                                       capabilities = EXCLUDED.capabilities,
                                       domain_tags = EXCLUDED.domain_tags,
                                       resources = EXCLUDED.resources,
                                       health_config = EXCLUDED.health_config,
                                       agent_category = EXCLUDED.agent_category,
                                       status = EXCLUDED.status,
                                       input_contract = EXCLUDED.input_contract,
                                       updated_at = NOW();


-- =============================================================================
-- STEP 1: VERIFY NAV TABLES BEFORE RUNNING ANYTHING
-- =============================================================================
-- Run these queries to check the current state of navigation for the site.
-- If careers and insights are NOT in site_nav_items, populate_nav_tables
-- needs to run first (which the nav-updater does automatically).

-- Check site_nav_items:
-- SELECT ni.label, ni.url, ng.group_type, ni.position, ni.status
-- FROM site_nav_items ni
-- JOIN site_nav_groups ng ON ni.group_id = ng.id
-- WHERE ni.site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'
-- ORDER BY ng.position, ni.position;

-- Check pages table for careers/insights:
-- SELECT name, title, url, in_header, in_footer, status, build_status, nav_order
-- FROM pages
-- WHERE site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'
--   AND name IN ('careers', 'insights')
-- ORDER BY name;

-- Check ALL pages and their header/footer flags:
-- SELECT name, in_header, in_footer, status, build_status
-- FROM pages
-- WHERE site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'
--   AND status = 'active'
-- ORDER BY nav_order, name;


-- =============================================================================
-- STEP 2: TEST WITH ONE PAGE REBUILD (FAQ - genuinely stale)
-- =============================================================================
-- Flag FAQ as needs_rebuild so page-rebuild picks it up.
-- This tests the full flow: content regeneration + fresh header from nav tables.

-- UPDATE pages
-- SET build_status = 'needs_rebuild', updated_at = NOW()
-- WHERE site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'
--   AND name = 'faq'
--   AND status = 'active';

-- Reset the stuck maintenance_queue task:
-- UPDATE maintenance_queue
-- SET status = 'pending', claimed_at = NULL, claimed_by = NULL
-- WHERE id = '5a20f87f-b780-457e-bd5f-0ef17bbba3f4';

-- Then trigger via generic agent with:
-- { "action": "process", "agent_type": "page-rebuild", "data": { "domain": "leopardessconsulting.co.uk" } }

-- OR manually trigger maintenance-triage:
-- { "action": "process", "agent_type": "maintenance-triage", "data": { "domain": "leopardessconsulting.co.uk", "dry_run": false, "stale_threshold_days": 4 } }


-- =============================================================================
-- STEP 3: RUN NAV-UPDATER FOR ALL PAGES (after verifying one page works)
-- =============================================================================
-- Trigger via generic agent:
-- { "action": "process", "agent_type": "nav-updater", "data": { "domain": "leopardessconsulting.co.uk" } }
--
-- This will:
--   1. Load site record
--   2. Rebuild nav tables (populate_nav_tables) - adds careers/insights to nav
--   3. Re-render header/footer with new nav items
--   4. Reassemble ALL pages with updated header/footer
--   5. Deploy to git + Cloudflare