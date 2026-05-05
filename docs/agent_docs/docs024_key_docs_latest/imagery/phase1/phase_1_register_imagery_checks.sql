-- ============================================================================
-- Phase 1 imagery work — register the three new discovery checks
--
-- Appends three check names to design-discovery-agent.run_checks.config.checks:
--   - unfulfilled_image_prompt   (planner asked for image, no asset exists)
--   - placeholder_image_in_use   (rendered HTML uses fallback path, no asset)
--   - image_url_404              (rendered HTML references unknown image path)
--
-- All three are implemented in
-- platform/orchestration/actions/discovery_checks/check_*.go and register
-- themselves via init(). The chassis must be rebuilt and redeployed for the
-- registry to pick them up — until then this SQL change is dormant
-- (RunDiscoveryChecksAction warns about unregistered names but doesn't fail).
--
-- Pattern note (per HANDOFF_2026-04-19):
-- We use jsonb || array-append rather than jsonb_set with a literal array
-- replacement. The append pattern can never accidentally drop existing
-- entries — replacement risks losing checks added since the SQL was drafted.
--
-- Idempotency:
-- Re-running this is safe because we use ?| (any-of) to test for the new
-- names being already present. Each check is appended only if missing.
-- ============================================================================

BEGIN;

-- Helper: a small CTE that lists the new check names so the UPDATE below
-- stays declarative and easy to review.

WITH new_checks AS (
    SELECT 'unfulfilled_image_prompt'::text AS name
    UNION ALL SELECT 'placeholder_image_in_use'
    UNION ALL SELECT 'image_url_404'
),
to_add AS (
    -- Filter to only the names not already present
    SELECT name FROM new_checks
    WHERE NOT (
        (SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
         FROM agent_definitions
         WHERE type = 'design-discovery-agent')
        @> to_jsonb(name)
    )
),
new_array AS (
    SELECT jsonb_agg(to_jsonb(name)) AS arr FROM to_add
)
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
            || COALESCE((SELECT arr FROM new_array), '[]'::jsonb)
    ),
    updated_at = NOW()
WHERE type = 'design-discovery-agent'
  AND EXISTS (SELECT 1 FROM to_add);

-- ----------------------------------------------------------------------------
-- Verification: show the resulting array; expect 14 entries with the three
-- new names at the end.
-- ----------------------------------------------------------------------------
SELECT type,
       jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') AS count,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' AS checks
FROM agent_definitions
WHERE type = 'design-discovery-agent';

COMMIT;


-- ============================================================================
-- ROLLBACK — removes the three new check names (preserves the 11 originals)
-- ============================================================================
-- BEGIN;
-- UPDATE agent_definitions
-- SET default_config = jsonb_set(
--     default_config,
--     '{workflow,steps,run_checks,config,checks}',
--     (
--         SELECT jsonb_agg(elem)
--         FROM jsonb_array_elements_text(
--             default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
--         ) AS elem
--         WHERE elem NOT IN ('unfulfilled_image_prompt','placeholder_image_in_use','image_url_404')
--     )
-- ),
--     updated_at = NOW()
-- WHERE type = 'design-discovery-agent';
-- COMMIT;
