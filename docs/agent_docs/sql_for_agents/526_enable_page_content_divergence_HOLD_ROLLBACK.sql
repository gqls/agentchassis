-- 526_enable_page_content_divergence_HOLD_ROLLBACK.sql
--
-- Un-enables the divergence sweep. Restores today's behaviour byte for byte:
-- availability-discovery-agent goes back to running site_unreachable alone, and
-- page_content_divergence becomes registered-but-undriven, which is exactly its
-- state before 526 was applied.
--
-- WHEN TO RUN THIS. Either symptom is sufficient and neither needs a diagnosis
-- first — restore service, then investigate:
--   * availability-discovery-agent runs start FAILING at run_checks (the image
--     does not register the name after all — this takes site_unreachable down
--     with it, which is the damage that matters and is not visible in this
--     check's own output);
--   * page_content_divergence items appear against pages that are demonstrably
--     healthy (re-run the comparison by hand before concluding this).
--
-- It does NOT delete any work items already filed. They are flag-only and carry
-- both hashes, so they remain readable evidence; close them by hand if they
-- were false.

BEGIN;

SELECT snapshot_agent('availability-discovery-agent', '526_ROLLBACK: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (SELECT COALESCE(jsonb_agg(c), '[]'::jsonb)
            FROM jsonb_array_elements(
                   default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') AS c
           WHERE c <> '"page_content_divergence"'::jsonb)),
       updated_at = NOW()
 WHERE type = 'availability-discovery-agent'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

DO $$
DECLARE still int; n_checks int;
BEGIN
    SELECT count(*) INTO still FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'availability-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["page_content_divergence"]'::jsonb;
    IF still <> 0 THEN
        RAISE EXCEPTION '526 ROLLBACK verify: page_content_divergence is still in the checks array';
    END IF;

    -- site_unreachable MUST survive. A rollback that restores the fault it was
    -- run to fix, by emptying the array, is worse than the fault.
    SELECT jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
      INTO n_checks FROM agent_definitions
     WHERE type='availability-discovery-agent'
       AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;
    IF n_checks <> 1 THEN
        RAISE EXCEPTION '526 ROLLBACK verify: checks array is % long, expected exactly 1 (site_unreachable)', n_checks;
    END IF;
END $$;

COMMIT;
