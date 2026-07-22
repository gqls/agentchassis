-- 194: enable the missing_model_directory_section / missing_model_directory_page
-- discovery checks (model_directory_pipeline Phase D)
--
-- Context: two new Layer 1 discovery checks
-- (platform/orchestration/actions/discovery_checks/check_model_directory.go)
-- detect a site that opted into the model directory
-- (site_specs.classification.content_features.model_directory) but has no
-- homepage section / dedicated page for it yet, once the global registry
-- (directory_entities/directory_claims) actually has publishable data.
--
-- SAFE TO APPLY AHEAD OF THE IMAGE, matching the precedent set by 190
-- (contact_form_undeliverable): an unregistered check NAME is skipped, not an
-- error (discovery_checks.go:122-127) — but this is applied on the same
-- image-first-then-seed schedule as the rest of Phase D regardless, so the
-- check actually starts finding anything the moment it's enabled rather than
-- silently no-opping for an unknown stretch of time. Requires the chassis
-- image that registers "missing_model_directory_section" /
-- "missing_model_directory_page" (pod-verify before relying on this):
--   kubectl -n ai-persona-system exec <agent-chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c missing_model_directory_section'
--
-- Applied the same statement shape as 150_/188_/189_/190_.
--
-- Verify after applying:
--   SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
--   FROM agent_definitions WHERE type='completeness-discovery-agent' AND is_active;
--   -- expect the array to end with "missing_model_directory_section", "missing_model_directory_page"
--   SELECT status, count(*) FROM site_work_items
--   WHERE item_type IN ('missing_model_directory_section', 'missing_model_directory_page')
--   GROUP BY status;
--   -- items appear only for sites that BOTH opted in AND have registry data
--   -- (see check_model_directory.go's modelDirectoryHasPublishableData gate)

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
        || '["missing_model_directory_section", "missing_model_directory_page"]'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent'
  AND is_active
  AND deleted_at IS NULL
  AND COALESCE(is_snapshot, false) = false   -- never touch a snapshot row
  -- idempotent: skip if already enabled
  AND NOT (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
          ? 'missing_model_directory_section';

COMMIT;
