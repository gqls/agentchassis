-- 708_enable_unrendered_page_imagery_HOLD_ROLLBACK.sql — removes
-- "unrendered_page_imagery" from design-discovery-agent's checks array,
-- restoring pre-708 behaviour exactly. Removing a name is always safe — an
-- absent check is simply not run — so this needs no hold of its own.
-- Sidecar, hand-run only (SIDECAR_RE excludes it from --apply). Added per the
-- round-2 debug_historian note: 709 ships a rollback sibling, 708 should too.

BEGIN;

DO $$
DECLARE present int;
BEGIN
    SELECT count(*) INTO present FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'design-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["unrendered_page_imagery"]'::jsonb;
    IF present = 0 THEN
        RAISE EXCEPTION '708 rollback: the name is not in the array — 708 was never applied (or already rolled back); nothing to do';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (SELECT COALESCE(jsonb_agg(e), '[]'::jsonb)
            FROM jsonb_array_elements(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') AS e
           WHERE e <> '"unrendered_page_imagery"'::jsonb)),
       updated_at = NOW()
 WHERE type = 'design-discovery-agent'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

DO $$
DECLARE still int;
BEGIN
    SELECT count(*) INTO still FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'design-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["unrendered_page_imagery"]'::jsonb;
    IF still <> 0 THEN
        RAISE EXCEPTION '708 rollback verify: the name survived the removal';
    END IF;
    RAISE NOTICE '708 rollback OK: unrendered_page_imagery removed from the checks array';
END $$;

COMMIT;
