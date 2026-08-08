-- 345_experience_planner_site_brief_becomes_data_ROLLBACK.sql
--
-- Restores experience-planner's default_config to the pre-345 state from
-- agent_definitions_backup. After this, the agent is contaminated again exactly
-- as bugs_open/227 describes — that is what a rollback of this change MEANS.
--
-- THE SNAPSHOT PICK IS THE WHOLE RISK HERE. Every agent_definitions_backup row
-- for one agent carries the SOURCE row's id and created_at, so `ORDER BY
-- created_at DESC LIMIT 1` returns an ARBITRARY snapshot — for council-gate on
-- 2026-07-30 it returned a 17 July one (LANDMINES.md). Order by
-- snapshot_taken_at, and pin the reason. Both are done below, and the guard
-- refuses if the chosen row does not actually carry the pre-change text.
--
-- The doc_notes brief for vonc-spark-game is DELIBERATELY LEFT IN PLACE. Once
-- default_config is restored, no step reads it (load_brief is gone with the rest
-- of the config), so it is inert; and it is the only copy of that brief outside
-- the prompt. Deleting it would be the one irreversible thing in this file. If
-- you truly want it gone, do it by hand, eyes open:
--   DELETE FROM doc_notes WHERE subject_type='experience'
--     AND subject_key='vonc-spark-game' AND categories @> '["experience-brief"]'::jsonb;

-- Guard: the snapshot we are about to restore must be the PRE-change one.
-- A row carrying the post-change value restores nothing and silently succeeds.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions_backup
         WHERE type = 'experience-planner'
           AND snapshot_reason LIKE 'pre-update: 227%'
           AND default_config::text ~* 'provocation'
           AND default_config #> '{workflow,steps,load_brief}' IS NULL
    ) THEN
        RAISE EXCEPTION '227/345 ROLLBACK: no backup row that carries the PRE-345 config (contaminated text present, load_brief absent). Refusing to restore something that is not the old state.';
    END IF;
END $$;

BEGIN;

UPDATE agent_definitions a
   SET default_config = b.default_config,
       updated_at     = now()
  FROM (
        SELECT default_config
          FROM agent_definitions_backup
         WHERE type = 'experience-planner'
           AND snapshot_reason LIKE 'pre-update: 227%'
           AND default_config::text ~* 'provocation'
           AND default_config #> '{workflow,steps,load_brief}' IS NULL
         ORDER BY snapshot_taken_at DESC
         LIMIT 1
       ) b
 WHERE a.type = 'experience-planner'
   AND a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL;

-- VERIFY: a DO block that RAISEs. A verify block of bare SELECTs cannot stop the
-- COMMIT — ON_ERROR_STOP ignores a non-empty result.
DO $$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n
      FROM agent_definitions
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #> '{workflow,steps,load_brief}' IS NULL
       AND default_config #>> '{workflow,steps,load_schema_hint,next_step}' = 'compose'
       AND default_config #>> '{workflow,steps,compose,config,prompt_template}' LIKE '%The diagnosis you are fixing%'
       AND NOT (default_config #> '{workflow,steps,compose,config,input_fields}' @> '["experience_brief"]'::jsonb);
    IF n <> 1 THEN
        RAISE EXCEPTION '227/345 ROLLBACK VERIFY FAILED: expected exactly 1 restored row, found %', n;
    END IF;
END $$;

COMMIT;
