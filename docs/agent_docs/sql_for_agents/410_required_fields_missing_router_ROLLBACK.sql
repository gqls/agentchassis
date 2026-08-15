-- 410 ROLLBACK — remove the required_fields_missing router
--
-- Safe to run at any time BEFORE assignment: 410 seeds one brand-new agent type
-- and assigns no work item to it, so removing it restores exactly the prior
-- state (the type flag-only, as its check intended before the 2026-08-15 owner
-- ruling).
--
-- ⚠ REFUSES once anything routes to it. If work items have since been assigned
-- (the canary/fleet step described in 410's header), deleting the definition
-- would strand those rows pointing at a handler that no longer exists —
-- dispatchable, claimable, and unrunnable. Un-assign first:
--
--   UPDATE site_work_items
--      SET handler_agent = '', status = 'needs_human_review'
--    WHERE handler_agent = 'required-fields-missing-handler'
--      AND status NOT IN ('complete','verified','rejected','wont_fix',
--                         'cancelled','failed','unresolved');
--
-- (terminal rows keep their handler_agent as history — the refusal below only
-- counts non-terminal rows), then re-run this file.
--
-- NOTE the Go half separately: if check_required_fields_missing.go has already
-- shipped naming this handler, rolling back only the definition re-creates the
-- bugs_closed/077 shape (a producer filing items for a handler that does not
-- exist — items go 'blocked' at claim). Roll the producer back too, or leave
-- the definition in place.
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 410_…_ROLLBACK.sql

BEGIN;

DO $$
DECLARE
    assigned integer;
    removed  integer;
BEGIN
    SELECT count(*) INTO assigned FROM site_work_items
     WHERE handler_agent = 'required-fields-missing-handler'
       AND status NOT IN ('complete','verified','rejected','wont_fix',
                          'cancelled','failed','unresolved');

    IF assigned <> 0 THEN
        RAISE EXCEPTION
            '410 ROLLBACK REFUSED: % non-terminal work item(s) still route to required-fields-missing-handler. Un-assign them first (see this file''s header) or they will be left pointing at a handler that does not exist.',
            assigned;
    END IF;

    DELETE FROM agent_definitions
     WHERE type = 'required-fields-missing-handler';
    GET DIAGNOSTICS removed = ROW_COUNT;

    -- Not an error: a rollback run twice, or run where 410 never applied,
    -- should be a no-op rather than a failure.
    RAISE NOTICE '410 ROLLBACK: removed % agent definition row(s)', removed;

    IF EXISTS (SELECT 1 FROM agent_definitions
                WHERE type = 'required-fields-missing-handler') THEN
        RAISE EXCEPTION '410 ROLLBACK: rows survived the DELETE — refusing to commit a partial revert';
    END IF;
END $$;

COMMIT;
