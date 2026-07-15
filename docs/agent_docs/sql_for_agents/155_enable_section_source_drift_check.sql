-- 155: enable the section_source_drift discovery check
--
-- Context (2026-07-15): a page's section list lives in three stores read in
-- priority order by load_page_sections_from_spec (site_plan_sections table >
-- site_specs.site_plan aspect > pages.sections), and the winner is synced
-- down over pages.sections on every rebuild. Editing only a lower store lets
-- the next rebuild silently revert the edit (bit the robot-hands
-- product-detail swap; migration 154 fixed it). This check flags pages whose
-- authoritative source disagrees with pages.sections, so the drift is caught
-- before a rebuild acts on it. Flag-only, needs_human_review, capped 25/pass.
--
-- Runs under completeness-discovery-agent alongside the other checks.
-- Requires the chassis image that registers "section_source_drift".
--
-- Verify after applying:
--   SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
--   FROM agent_definitions WHERE type='completeness-discovery-agent' AND is_active;
--   -- expect the array to contain "section_source_drift"

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
        || '"section_source_drift"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent'
  AND is_active
  AND deleted_at IS NULL
  AND NOT (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
          ? 'section_source_drift';

COMMIT;
