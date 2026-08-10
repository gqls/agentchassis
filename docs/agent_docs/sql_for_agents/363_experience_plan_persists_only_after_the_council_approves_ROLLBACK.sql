-- 363_experience_plan_persists_only_after_the_council_approves_ROLLBACK.sql
--
-- Restores the experience-planner row from the snapshot 363 took.
--
-- ⚠ Pick by snapshot_taken_at, NOT created_at: every backup row for one agent
-- shares the SOURCE row's id and created_at, so ORDER BY created_at returns an
-- arbitrary snapshot (LANDMINES 2026-07-30).
--
-- ⚠ This restores the WHOLE default_config as it was before 363, which includes
-- 345's brief-as-data chain (345 was applied first and is NOT rolled back by
-- this file). If both need reverting, roll 363 back first, then 345.
--
-- After running, the graph should read compose/recompose/reframe -> persist_plan
-- and check_approved.then_step = complete.

BEGIN;

DO $$
DECLARE
    snap_cfg jsonb; snap_at timestamptz;
BEGIN
    SELECT default_config, snapshot_taken_at
      INTO snap_cfg, snap_at
      FROM agent_definitions_backup
     WHERE type = 'experience-planner'
       AND snapshot_reason LIKE 'pre-update: 227 persist%'
     ORDER BY snapshot_taken_at DESC
     LIMIT 1;

    IF snap_cfg IS NULL THEN
        RAISE EXCEPTION '227/363 rollback: no snapshot found with reason LIKE ''pre-update: 227 persist%%''';
    END IF;

    UPDATE agent_definitions
       SET default_config = snap_cfg
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    RAISE NOTICE '227/363 rolled back to snapshot taken at %', snap_at;
END $$;

-- Assert the pre-363 shape is back.
DO $$
DECLARE c_next text; a_then text;
BEGIN
    SELECT default_config #>> '{workflow,steps,compose,next_step}',
           default_config #>> '{workflow,steps,check_approved,config,then_step}'
      INTO c_next, a_then
      FROM agent_definitions
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF c_next <> 'persist_plan' OR a_then <> 'complete' THEN
        RAISE EXCEPTION '227/363 rollback verify: graph is not pre-363 (compose->%, approved_then=%)',
                        c_next, a_then;
    END IF;
    RAISE NOTICE '227/363 rollback verified.';
END $$;

COMMIT;
