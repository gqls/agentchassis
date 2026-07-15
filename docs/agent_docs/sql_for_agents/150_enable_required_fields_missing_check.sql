-- 150: enable the required_fields_missing discovery check
--
-- Context (2026-07-14, robot-hands empty-product-pages investigation):
-- deployed component instances can carry chrome-only content_data while every
-- schema-required, LLM-sourced value field is absent — the template renders
-- them as empty strings and the page ships as a hollow shell. empty_sections
-- misses these (the markup is full of static label text) and
-- sectionHasVisibleContent keeps them (>10 chars of text). The new check
-- (check_required_fields_missing.go) flags them at needs_human_review,
-- flag-only, capped at 25 per pass.
--
-- Runs under completeness-discovery-agent alongside empty_sections.
-- Requires the chassis image that registers "required_fields_missing".
--
-- Verify after applying:
--   SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
--   FROM agent_definitions WHERE type='completeness-discovery-agent' AND is_active;
--   -- expect the array to end with "required_fields_missing"

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
        || '"required_fields_missing"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent'
  AND is_active
  AND deleted_at IS NULL
  -- idempotent: skip if already enabled
  AND NOT (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
          ? 'required_fields_missing';

COMMIT;
