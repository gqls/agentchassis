-- 397 ROLLBACK — remove the two image routers
--
-- Safe to run at any time BEFORE assignment: 397 seeds two brand-new agent
-- types and assigns no work item to either, so removing them restores exactly
-- the prior state (both item types flag-only, as their checks intend).
--
-- ⚠ REFUSES once anything routes to them. If work items have since been
-- assigned (the deferred step described in 397's header), deleting the agent
-- definition would strand those rows pointing at a handler that no longer
-- exists — dispatchable, claimable, and unrunnable. Un-assign first:
--
--   UPDATE site_work_items
--      SET handler_agent = '', status = 'needs_human_review'
--    WHERE handler_agent IN ('image-url-404-handler',
--                            'image-source-unsatisfiable-handler');
--
-- then re-run this file. The refusal is the point: it makes the ordering
-- explicit instead of leaving it to whoever reverts under time pressure.
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 397_…_ROLLBACK.sql

BEGIN;

DO $$
DECLARE
    assigned integer;
    removed  integer;
BEGIN
    SELECT count(*) INTO assigned FROM site_work_items
     WHERE handler_agent IN ('image-url-404-handler', 'image-source-unsatisfiable-handler');

    IF assigned <> 0 THEN
        RAISE EXCEPTION
            '397 ROLLBACK REFUSED: % work item(s) still route to these handlers. Un-assign them first (see this file''s header) or they will be left pointing at a handler that does not exist.',
            assigned;
    END IF;

    DELETE FROM agent_definitions
     WHERE type IN ('image-url-404-handler', 'image-source-unsatisfiable-handler');
    GET DIAGNOSTICS removed = ROW_COUNT;

    -- Not an error: a rollback run twice, or run against a database where 397
    -- never applied, should be a no-op rather than a failure.
    RAISE NOTICE '397 ROLLBACK: removed % agent definition row(s)', removed;

    IF EXISTS (SELECT 1 FROM agent_definitions
                WHERE type IN ('image-url-404-handler', 'image-source-unsatisfiable-handler')) THEN
        RAISE EXCEPTION '397 ROLLBACK: rows survived the DELETE — refusing to commit a partial revert';
    END IF;
END $$;

COMMIT;
