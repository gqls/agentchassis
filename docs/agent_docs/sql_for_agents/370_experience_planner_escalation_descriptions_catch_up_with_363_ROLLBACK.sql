-- 370_experience_planner_escalation_descriptions_catch_up_with_363_ROLLBACK.sql
--
-- Restores the experience-planner row from the snapshot 370 took.
--
-- ⚠ Pick by snapshot_taken_at, NOT created_at: every backup row for one agent
-- shares the SOURCE row's id and created_at, so ORDER BY created_at returns an
-- arbitrary snapshot (LANDMINES 2026-07-30).
--
-- ⚠ This restores the WHOLE default_config as it was before 370, which includes
-- 345's brief-as-data chain and 363's persist-after-approval rewire (both applied
-- earlier and NOT rolled back by this file). Rolling 370 back therefore reinstates
-- three strings that DESCRIBE THE PRE-363 GRAPH while the post-363 graph is still
-- live — which is the drift 370 exists to remove. Only do it if 363 is being
-- rolled back too, and roll 370 back FIRST, then 363, then 345.
--
-- After running, complete_escalated.description should again read
-- "The current (rejected) plan stays is_current ...".

BEGIN;

DO $$
DECLARE
    snap_cfg jsonb; snap_at timestamptz;
BEGIN
    SELECT default_config, snapshot_taken_at
      INTO snap_cfg, snap_at
      FROM agent_definitions_backup
     WHERE type = 'experience-planner'
       AND snapshot_reason LIKE 'pre-update: 227 escalation descriptions%'
     ORDER BY snapshot_taken_at DESC
     LIMIT 1;

    IF snap_cfg IS NULL THEN
        RAISE EXCEPTION '227/370 rollback: no snapshot found with reason LIKE ''pre-update: 227 escalation descriptions%%''';
    END IF;

    UPDATE agent_definitions
       SET default_config = snap_cfg
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    RAISE NOTICE '227/370 rolled back to snapshot taken at %', snap_at;
END $$;

-- Assert the pre-370 strings are back.
DO $$
DECLARE esc_desc text;
BEGIN
    SELECT default_config #>> '{workflow,steps,complete_escalated,description}'
      INTO esc_desc
      FROM agent_definitions
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF esc_desc NOT LIKE '%stays is_current%' THEN
        RAISE EXCEPTION '227/370 rollback verify: complete_escalated.description is not the pre-370 text (%)',
                        left(esc_desc, 80);
    END IF;
    RAISE NOTICE '227/370 rollback verified.';
END $$;

COMMIT;
