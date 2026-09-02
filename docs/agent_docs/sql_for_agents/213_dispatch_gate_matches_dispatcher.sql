-- 213_dispatch_gate_matches_dispatcher.sql
--
-- bugs_open/029 residual — make the build trigger's GATE ask the same question
-- as the DISPATCHER, in both directions.
--
-- WHY
-- ---
-- Three predicates sit in one chain and none of them agreed:
--
--   A  scheduled_tasks.build-pipeline-trigger.pre_query   "should we fire?"
--        status='triaged'  pipeline='build'  s.locked_at IS NULL
--   B  agent_definitions.build-pipeline-trigger
--        .workflow.steps.find_dispatchable_site           "which site?"
--        status IN ('triaged','approved')  any pipeline  no lock check
--        AND NOT EXISTS (a 'claimed' item on the site)
--   C  load_work_items (Go, configured with no item_pipeline)  "which items?"
--        status IN ('triaged','approved')  any pipeline
--
-- Consequences, measured live 2026-07-25/26 (evidence in
-- docs024_key_docs_latest/bugfix_029_dispatch_gate/NOTES_dispatch_gate.md):
--
--   * A ignores the claimed-mutex that B enforces, so the gate said "2 pending
--     sites" while the dispatcher could dispatch NOTHING. The trigger fired
--     every 120s and landed on complete_idle — ~360 wasted orchestrations/day,
--     and a false heartbeat that made a stalled pipeline look alive. That false
--     heartbeat is what misread 029 three reproductions running.
--   * A only sees pipeline='build', while B and C dispatch any pipeline. There
--     is no trigger for 'content'/'design'/'experience'/'maintenance' items, so
--     they are dispatched only opportunistically — whenever build work happens
--     to co-exist. With no build work anywhere, they wait indefinitely.
--   * A only sees status='triaged' while B and C accept 'approved'. Inert today
--     (0 approved rows ever) but it is exactly the path HITL approval
--     (bugs_open/033) would write to.
--   * A honours sites.locked_at and B does not, so B can dispatch a locked site.
--     Inert today (0 of 32 sites locked, ever).
--
-- WHAT THIS DOES
-- --------------
-- Rewrites A to be the existence-test of B, and adds the lock clause to B, so
-- the two are literally the same predicate rather than accidentally equivalent.
-- After this, "the trigger is not firing" means "there is genuinely nothing
-- dispatchable" — which is the diagnostic 029 needed and never had.
--
-- NOT a fix for hung spawns; that is bugs_open/003 (F2/F3 live v1.0.1159).
-- Deliberately does NOT add another orchestration reaper: stale-orchestration-
-- reaper already exists and is not the lever.
--
-- ROLLBACK
-- --------
--   The pre-image of the agent row is taken by snapshot_agent() below; restore
--   with the usual snapshot restore. To revert the scheduled task by hand:
--
--   UPDATE scheduled_tasks SET pre_query = $old$
--   SELECT COUNT(*)::text as pending_sites
--   FROM sites s
--   WHERE s.locked_at IS NULL
--     AND EXISTS (
--       SELECT 1 FROM site_work_items wi
--       WHERE wi.site_id = s.id
--         AND wi.status = 'triaged'
--         AND wi.pipeline = 'build'
--         AND wi.attempt_count < wi.max_attempts
--   )
--   HAVING COUNT(*) > 0
--   $old$ WHERE name = 'build-pipeline-trigger';
--   -- and drop " AND s.locked_at IS NULL" from find_dispatchable_site's query.

BEGIN;

SELECT snapshot_agent('build-pipeline-trigger',
                      '213_dispatch_gate_matches_dispatcher.sql: pre-update');

-- ---------------------------------------------------------------------------
-- A — the gate becomes the existence-test of the dispatcher
-- ---------------------------------------------------------------------------
UPDATE scheduled_tasks
SET pre_query = $pq$
SELECT COUNT(*)::text AS pending_sites
FROM (
    SELECT DISTINCT wi.site_id
      FROM site_work_items wi
      JOIN sites s ON s.id = wi.site_id
     WHERE wi.status IN ('triaged', 'approved')
       AND wi.attempt_count < wi.max_attempts
       AND s.locked_at IS NULL
       AND NOT EXISTS (
           SELECT 1 FROM site_work_items active
            WHERE active.site_id = wi.site_id
              AND active.status = 'claimed'
       )
) dispatchable
HAVING COUNT(*) > 0
$pq$,
    updated_at = NOW()
WHERE name = 'build-pipeline-trigger';

-- ---------------------------------------------------------------------------
-- B — the dispatcher honours the site lock the gate has always honoured.
--     Site-selection order (DISTINCT ON site_id, priority ASC, LIMIT 1) is
--     preserved exactly: changing dispatch fairness is not this fix's remit.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,find_dispatchable_site,config,query}',
        to_jsonb($q$SELECT DISTINCT ON (wi.site_id) wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.status IN ('triaged', 'approved') AND wi.attempt_count < wi.max_attempts AND s.locked_at IS NULL AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = 'claimed') ORDER BY wi.site_id, wi.priority ASC LIMIT 1$q$::text)
    ),
    updated_at = NOW()
WHERE type = 'build-pipeline-trigger'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- Guards — assert the exact post-conditions, inside the transaction so a
-- failure rolls the whole file back.
-- ---------------------------------------------------------------------------
DO $guard$
DECLARE
    v_pre   text;
    v_query text;
    v_n     int;
BEGIN
    SELECT pre_query INTO v_pre
      FROM scheduled_tasks WHERE name = 'build-pipeline-trigger';

    IF v_pre IS NULL THEN
        RAISE EXCEPTION '213: build-pipeline-trigger has no pre_query';
    END IF;
    IF v_pre LIKE '%wi.domain%' THEN
        RAISE EXCEPTION '213: pre_query still references the dropped column wi.domain';
    END IF;
    IF v_pre NOT LIKE '%active.status%' THEN
        RAISE EXCEPTION '213: pre_query does not enforce the claimed-item mutex';
    END IF;
    IF v_pre NOT LIKE '%approved%' THEN
        RAISE EXCEPTION '213: pre_query does not accept approved items';
    END IF;
    IF v_pre LIKE '%pipeline = %' OR v_pre LIKE '%pipeline=%' THEN
        RAISE EXCEPTION '213: pre_query still filters on a single pipeline';
    END IF;
    IF v_pre NOT LIKE '%s.locked_at IS NULL%' THEN
        RAISE EXCEPTION '213: pre_query dropped the site-lock check';
    END IF;

    SELECT default_config->'workflow'->'steps'->'find_dispatchable_site'->'config'->>'query'
      INTO v_query
      FROM agent_definitions
     WHERE type = 'build-pipeline-trigger'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_query IS NULL THEN
        RAISE EXCEPTION '213: find_dispatchable_site query missing after update';
    END IF;
    IF v_query NOT LIKE '%s.locked_at IS NULL%' THEN
        RAISE EXCEPTION '213: find_dispatchable_site did not get the site-lock check';
    END IF;
    IF v_query NOT LIKE '%active.status%' THEN
        RAISE EXCEPTION '213: find_dispatchable_site lost the claimed-item mutex';
    END IF;
    IF v_query NOT LIKE '%DISTINCT ON (wi.site_id)%' THEN
        RAISE EXCEPTION '213: find_dispatchable_site lost its site-selection shape';
    END IF;

    -- The point of the change: gate and dispatcher now select the same set.
    -- Report it (0 on both sides is a valid agreement, not a failure).
    SELECT count(*) INTO v_n
      FROM (
        SELECT DISTINCT wi.site_id
          FROM site_work_items wi
          JOIN sites s ON s.id = wi.site_id
         WHERE wi.status IN ('triaged', 'approved')
           AND wi.attempt_count < wi.max_attempts
           AND s.locked_at IS NULL
           AND NOT EXISTS (SELECT 1 FROM site_work_items active
                            WHERE active.site_id = wi.site_id
                              AND active.status = 'claimed')
      ) d;

    RAISE NOTICE '213: gate and dispatcher aligned; % dispatchable site(s) right now', v_n;
END
$guard$;

COMMIT;
