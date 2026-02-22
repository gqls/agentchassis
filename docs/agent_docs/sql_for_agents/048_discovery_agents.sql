-- ============================================================================
-- 023_discovery_agents.sql
-- Discovery agent definitions for the unified build/maintenance system
-- ============================================================================
-- These agents write findings to site_work_items (source='discovery',
-- status='detected'). They run as part of site-work-orchestrator step 4
-- (run_discovery) or standalone via maintenance-batch-scheduler.
--
-- Uses: run_discovery_checks action (discovery_checks.go)
-- Depends on: site_work_items table (023_site_work_items.sql)
--
-- Discovery agents find problems. They do not fix anything.
-- They do not call other agents.
-- ============================================================================


-- ============================================================================
-- AGENT: design-discovery-agent
-- ============================================================================
-- Checks: undeployed_assets, missing_css, duplicate_palette
-- Domain: design
-- Schedule: weekly (via maintenance profile)
-- All checks are algorithmic (no LLM calls)

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active,
    capabilities, domain_tags, agent_category, status,
    input_contract, output_contract
) VALUES (
             'design-discovery-agent',
             'Design Discovery Agent',
             'Scans sites for design-domain issues: undeployed assets, missing CSS themes, duplicate colour palettes. Writes findings to site_work_items as detected items. All checks are algorithmic — no LLM budget needed.',
             'specialist',
             jsonb_build_object(
                     'processing_mode', 'task',
                     'timeout_seconds', 120,
                     'workflow', jsonb_build_object(
                             'start_step', 'ensure_site_record',
                             'steps', jsonb_build_object(
                                     'ensure_site_record', jsonb_build_object(
                                     'action', 'ensure_site_record',
                                     'config', jsonb_build_object(
                                             'input_fields', jsonb_build_array('site_id', 'domain')
                                               ),
                                     'next_step', 'run_checks',
                                     'description', 'Load site record from domain or site_id',
                                     'output_field', 'site_record'
                                                           ),
                                     'run_checks', jsonb_build_object(
                                             'action', 'run_discovery_checks',
                                             'config', jsonb_build_object(
                                                     'site_id', 'site_record.site_id',
                                                     'check_domain', 'design',
                                                     'checks', jsonb_build_array(
                                                             'undeployed_assets', 'missing_css', 'duplicate_palette'
                                                               )
                                                       ),
                                             'next_step', 'complete',
                                             'description', 'Run design checks and write findings to site_work_items',
                                             'output_field', 'discovery_result'
                                                   ),
                                     'complete', jsonb_build_object(
                                             'action', 'complete_workflow',
                                             'config', jsonb_build_object(
                                                     'output_fields', jsonb_build_array('discovery_result')
                                                       ),
                                             'description', 'Design discovery complete'
                                                 )
                                      )
                                 )
             ),
             true,
             '["discovery", "design_audit", "css_check", "asset_check"]'::jsonb,
             '["maintenance", "discovery", "design"]'::jsonb,
             'analyst',
             'active',
             jsonb_build_object(
                     'required', jsonb_build_array('site_id'),
                     'optional', jsonb_build_array('domain'),
                     'description', 'Pass site_id or domain to scan a single site.'
             ),
             jsonb_build_object(
                     'produces', jsonb_build_object(
                     'discovery_result', 'items_inserted count, findings details, batch_id'
                                 )
             )
         )
    ON CONFLICT (type,version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              default_config = EXCLUDED.default_config,
                              capabilities = EXCLUDED.capabilities,
                              domain_tags = EXCLUDED.domain_tags,
                              agent_category = EXCLUDED.agent_category,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();


-- ============================================================================
-- AGENT: completeness-discovery-agent
-- ============================================================================
-- Checks: empty_sections
-- Domain: content
-- Schedule: every 3 days (via maintenance profile)
-- All checks are algorithmic (no LLM calls)

INSERT INTO agent_definitions (
    type, display_name, description, category,
    default_config, is_active,
    capabilities, domain_tags, agent_category, status,
    input_contract, output_contract
) VALUES (
             'completeness-discovery-agent',
             'Completeness Discovery Agent',
             'Scans sites for content completeness issues: empty deployed sections, placeholder content. Writes findings to site_work_items as detected items. All checks are algorithmic — no LLM budget needed.',
             'specialist',
             jsonb_build_object(
                     'processing_mode', 'task',
                     'timeout_seconds', 120,
                     'workflow', jsonb_build_object(
                             'start_step', 'ensure_site_record',
                             'steps', jsonb_build_object(
                                     'ensure_site_record', jsonb_build_object(
                                     'action', 'ensure_site_record',
                                     'config', jsonb_build_object(
                                             'input_fields', jsonb_build_array('site_id', 'domain')
                                               ),
                                     'next_step', 'run_checks',
                                     'description', 'Load site record from domain or site_id',
                                     'output_field', 'site_record'
                                                           ),
                                     'run_checks', jsonb_build_object(
                                             'action', 'run_discovery_checks',
                                             'config', jsonb_build_object(
                                                     'site_id', 'site_record.site_id',
                                                     'check_domain', 'content',
                                                     'checks', jsonb_build_array('empty_sections')
                                                       ),
                                             'next_step', 'complete',
                                             'description', 'Run completeness checks and write findings to site_work_items',
                                             'output_field', 'discovery_result'
                                                   ),
                                     'complete', jsonb_build_object(
                                             'action', 'complete_workflow',
                                             'config', jsonb_build_object(
                                                     'output_fields', jsonb_build_array('discovery_result')
                                                       ),
                                             'description', 'Completeness discovery complete'
                                                 )
                                      )
                                 )
             ),
             true,
             '["discovery", "content_audit", "completeness_check"]'::jsonb,
             '["maintenance", "discovery", "content"]'::jsonb,
             'analyst',
             'active',
             jsonb_build_object(
                     'required', jsonb_build_array('site_id'),
                     'optional', jsonb_build_array('domain'),
                     'description', 'Pass site_id or domain to scan a single site.'
             ),
             jsonb_build_object(
                     'produces', jsonb_build_object(
                     'discovery_result', 'items_inserted count, findings details, batch_id'
                                 )
             )
         )
    ON CONFLICT (type,version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              default_config = EXCLUDED.default_config,
                              capabilities = EXCLUDED.capabilities,
                              domain_tags = EXCLUDED.domain_tags,
                              agent_category = EXCLUDED.agent_category,
                              input_contract = EXCLUDED.input_contract,
                              output_contract = EXCLUDED.output_contract,
                              updated_at = NOW();


-- ============================================================================
-- NOTES: How discovery fits into the orchestrator
-- ============================================================================
--
-- site-work-orchestrator step 4 (run_discovery, phase 1):
--
--   "run_discovery": {
--       "action": "conditional",
--       "config": {
--           "condition": "site_record.status == 'deployed'",
--           "then_step": "spawn_design_discovery",
--           "else_step": "triage_new_items"
--       },
--       "description": "Skip discovery on new sites (nothing to discover yet)"
--   },
--   "spawn_design_discovery": {
--       "action": "spawn_agent",
--       "config": { "role": "design_discoverer", "agent_type": "design-discovery-agent" },
--       "next_step": "call_design_discovery",
--       "output_field": "design_discoverer"
--   },
--   "call_design_discovery": {
--       "action": "call_agent",
--       "config": {
--           "agent_type": "design-discovery-agent",
--           "target_role": "design_discoverer",
--           "input_mapping": { "site_id": "site_record.site_id" },
--           "timeout_seconds": 120
--       },
--       "next_step": "spawn_completeness_discovery",
--       "output_field": "design_findings"
--   },
--   ... (repeat for completeness-discovery-agent)
--   ... → triage_new_items
--
-- The findings land in site_work_items as status='detected'.
-- The triage step (step 5) then:
--   1. Reads items WHERE status='detected' AND site_id=$1
--   2. Sets resolution_path, adjusts priority, confirms handler_agent
--   3. Updates status to 'triaged'
-- Next orchestrator cycle processes them.
--
--
-- NOTES: Standalone execution via maintenance-batch-scheduler
-- ============================================================================
--
-- The batch scheduler can also trigger discovery agents directly:
--
--   FOR each site WHERE maintenance_profile.design.due:
--     spawn design-discovery-agent with { site_id: site.id }
--
-- This is the phase 1 pattern from the v6 doc. The batch scheduler
-- replaces maintenance-triage for the new system over time.
--
--
-- NOTES: Coexistence with existing system
-- ============================================================================
--
-- maintenance-triage + maintenance_queue continue working unchanged.
-- These discovery agents write to site_work_items (different table).
-- Both systems can scan the same site without conflict.
--
-- The existing scan_sites_for_maintenance checks (stale_pages, missing_content,
-- orphan_nav) still write to maintenance_queue. They can be ported to the
-- new system later by adding check functions to discovery_checks.go.