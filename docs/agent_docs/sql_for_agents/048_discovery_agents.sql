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


-----

-- fix hardcoded hero agents

-- 063b: Add hardcoded_section_colors to design-discovery-agent's check list
--
-- The design-discovery-agent runs: undeployed_assets, missing_css, duplicate_palette
-- This adds hardcoded_section_colors so the improvement loop detects hero/section
-- components with inline hardcoded hex backgrounds.
--
-- When detected, the dispatch loop routes to color-variable-fixer agent
-- which does find/replace in <style> blocks only — no content changes.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["undeployed_assets", "missing_css", "duplicate_palette", "hardcoded_section_colors"]'::jsonb
                     )
WHERE type = 'design-discovery-agent';

--

-- 064_forced_text_colors_check_fixer.sql
--
-- Wires the forced_text_colors discovery check and fixer into the existing
-- improvement pipeline:
--
-- Flow:
--   improvement-loop → design-discovery-agent (runs "forced_text_colors" check)
--     → writes site_work_item with handler_agent='color-variable-fixer'
--     → triage promotes to 'triaged'
--     → build-dispatch-loop spawns color-variable-fixer
--     → color-variable-fixer runs fix_hardcoded_colors THEN fix_forced_text_colors
--     → improvement-loop inserts needs_rerender → pages redeployed
--
-- Prerequisites:
--   - fix_forced_text_colors_action.go compiled and deployed
--   - Action registered in registry.go:
--       "fix_forced_text_colors": {
--           Handler:     FixForcedTextColorsAction,
--           Category:    "maintenance",
--           Description: "Remove forced text colors from child elements, validate WCAG contrast",
--       },
--   - Discovery check addition in run_discovery_checks_action.go:
--       findForcedTextColors function + containsCheck block
--
-- Both Go changes are in the companion .go files.

-- ============================================================
-- 1. Add "forced_text_colors" to design-discovery-agent checks
-- ============================================================
-- Current checks: ["undeployed_assets", "missing_css", "duplicate_palette", "hardcoded_section_colors"]
-- New:            ["undeployed_assets", "missing_css", "duplicate_palette", "hardcoded_section_colors", "forced_text_colors"]

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        '["undeployed_assets", "missing_css", "duplicate_palette", "hardcoded_section_colors", "forced_text_colors"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'design-discovery-agent';

-- ============================================================
-- 2. Extend color-variable-fixer workflow
-- ============================================================
-- Current workflow:
--   fix_colors → complete
--
-- New workflow:
--   fix_bg_colors → fix_text_colors → complete
--
-- Both steps are idempotent. Running the full workflow regardless
-- of which item type triggered it is safe — each step only
-- changes components that match its criteria.

-- Rename existing step and re-route
-- Step 1: Update the existing fix_colors step to point to fix_text_colors
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,fix_colors,next_step}',
        '"fix_text_colors"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'color-variable-fixer';

-- Step 2: Add the new fix_text_colors step
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,fix_text_colors}',
        '{
          "action": "fix_forced_text_colors",
          "config": {
            "fix_rendered": true,
            "fix_templates": true,
            "min_contrast": 4.5
          },
          "next_step": "complete",
          "description": "Remove forced text colors from child elements, validate WCAG AA contrast",
          "output_field": "text_color_result"
        }'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'color-variable-fixer';

-- Step 3: Update complete step to include both result fields
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,complete,config,output_fields}',
        '["fix_result", "text_color_result"]'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'color-variable-fixer';

-- ============================================================
-- 3. Verify the changes
-- ============================================================
-- Run these after applying to confirm:

-- Check design-discovery-agent has the new check
SELECT
    type,
    default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as checks
FROM agent_definitions
WHERE type = 'design-discovery-agent';

-- Check color-variable-fixer workflow flow
SELECT
    type,
    key as step_name,
    value->>'action' as action,
    value->>'next_step' as next_step,
    value->>'output_field' as output_field
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE type = 'color-variable-fixer'
ORDER BY
    CASE key
    WHEN 'fix_colors' THEN 1
    WHEN 'fix_text_colors' THEN 2
    WHEN 'complete' THEN 3
END;

---


-- check tools work - basic algorithmic tier 1

-- ============================================================
-- Add tool_health check to design-discovery-agent
-- ============================================================

-- Get current checks list first
SELECT type,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as current_checks
FROM agent_definitions
WHERE type = 'design-discovery-agent';

-- Add tool_health to the checks array
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        (
            SELECT jsonb_agg(elem)
            FROM (
                     SELECT elem FROM jsonb_array_elements(
                                              default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
                                      ) AS elem
                     UNION
                     SELECT '"tool_health"'::jsonb
                 ) sub
        )
                     ),
    updated_at = NOW()
WHERE type = 'design-discovery-agent'
  AND NOT (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' @> '"tool_health"'::jsonb);

-- Verify
SELECT type,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' as checks
FROM agent_definitions
WHERE type = 'design-discovery-agent';

---
-- phase_2g_step4_register_check.sql
--
-- Phase 2G step 4 wiring — register `unfulfilled_imagery_plan` in
-- design-discovery-agent's run_checks list so the new check actually
-- fires during discovery runs.
--
-- Without this, the Go check self-registers via init() but is never
-- invoked, because RunDiscoveryChecksAction iterates only the names
-- listed in the workflow step's config.
--
-- Idempotent: re-running is a no-op if the check is already registered.
--
-- The existing legacy check `unfulfilled_image_prompt` is intentionally
-- left in place. Both checks run in parallel during the transition;
-- the legacy check naturally stops finding gaps once all sites' planner
-- runs are post-2G.3. Deregistration is an operational decision for
-- some future date, not part of this migration.

\set ON_ERROR_STOP on

-- ── Backup ──

CREATE TABLE agent_def_design_discovery_agent_backup_20260513_pre_register_imagery_check AS
SELECT * FROM agent_definitions
WHERE type = 'design-discovery-agent' AND is_active = true;

SELECT
    (SELECT COUNT(*) FROM agent_definitions
     WHERE type = 'design-discovery-agent' AND is_active = true) AS live,
    (SELECT COUNT(*) FROM agent_def_design_discovery_agent_backup_20260513_pre_register_imagery_check) AS backup;

-- ── Migration ──

BEGIN;

DO $register$
DECLARE
v_checks jsonb;
BEGIN
    -- Read current list
SELECT default_config #> '{workflow,steps,run_checks,config,checks}'
INTO v_checks
FROM agent_definitions
WHERE type = 'design-discovery-agent' AND is_active = true;

IF v_checks IS NULL THEN
        RAISE EXCEPTION 'design-discovery-agent: run_checks.config.checks not found';
END IF;

    -- Idempotency: skip if already registered
    IF v_checks ? 'unfulfilled_imagery_plan' THEN
        RAISE NOTICE 'unfulfilled_imagery_plan already registered; no change';
        RETURN;
END IF;

    -- Append the new check name
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        v_checks || '["unfulfilled_imagery_plan"]'::jsonb
                     ),
    updated_at = now()
WHERE type = 'design-discovery-agent'
  AND is_active = true;

RAISE NOTICE 'unfulfilled_imagery_plan registered in design-discovery-agent run_checks list';
END
$register$;

-- ── Verification ──

DO $verify$
DECLARE
v_checks jsonb;
    v_check_count int;
BEGIN
SELECT default_config #> '{workflow,steps,run_checks,config,checks}'
INTO v_checks
FROM agent_definitions
WHERE type = 'design-discovery-agent' AND is_active = true;

IF NOT (v_checks ? 'unfulfilled_imagery_plan') THEN
        RAISE EXCEPTION 'unfulfilled_imagery_plan not in run_checks list after migration';
END IF;

    IF NOT (v_checks ? 'unfulfilled_image_prompt') THEN
        RAISE EXCEPTION 'legacy unfulfilled_image_prompt missing — migration may have clobbered the list';
END IF;

SELECT jsonb_array_length(v_checks) INTO v_check_count;
RAISE NOTICE 'design-discovery-agent now runs % checks (was 14, expected 15)', v_check_count;

    IF v_check_count < 15 THEN
        RAISE EXCEPTION 'check count is % after migration, expected ≥ 15', v_check_count;
END IF;
END
$verify$;

COMMIT;