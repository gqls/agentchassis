-- 603_enable_page_list_stale_HOLD_ROLLBACK.sql
--
-- Un-enables the page_list_stale sweep: completeness-discovery-agent stops
-- naming it, and the check becomes registered-but-undriven — its state before
-- 603 was applied. Does NOT delete filed items: they are ordinary page_rerender
-- requests (section_data_resolved) and complete on their own; if any were
-- wrong, they are a no-LLM re-render of a page that already renders.
--
-- WHEN: page_list_stale items appear against listings whose stored arrays
-- demonstrably match a fresh resolve (re-run the comparison by hand first), or
-- the completeness sweep's run_checks step starts failing after 603.

BEGIN;

SELECT snapshot_agent('completeness-discovery-agent', '603_ROLLBACK: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (SELECT COALESCE(jsonb_agg(c), '[]'::jsonb)
            FROM jsonb_array_elements(
                   default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') AS c
           WHERE c <> '"page_list_stale"'::jsonb)),
       updated_at = NOW()
 WHERE type = 'completeness-discovery-agent'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

DO $$
DECLARE still int; n_checks int;
BEGIN
    SELECT count(*) INTO still FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'completeness-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["page_list_stale"]'::jsonb;
    IF still <> 0 THEN
        RAISE EXCEPTION '603 ROLLBACK verify: page_list_stale is still in the checks array';
    END IF;
    SELECT jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
      INTO n_checks FROM agent_definitions
     WHERE type='completeness-discovery-agent'
       AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;
    IF n_checks < 44 THEN
        RAISE EXCEPTION '603 ROLLBACK verify: checks array is % long — the other checks must survive (44 on 2026-08-24)', n_checks;
    END IF;
END $$;

COMMIT;
