-- 751_d4b_governor_agent_scope_stage_a_ROLLBACK.sql — undo D4b stage A.
-- Restores the standalone 675 body of governor_admits, drops the agent namespace and the
-- withheld-runs record. REFUSES while anything live references governor_admits_agent: a
-- rollback that removes a function under a live caller is an outage, not a rollback.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname='governor_admits_agent') THEN
    RAISE EXCEPTION '751 ROLLBACK REFUSED: governor_admits_agent() absent — 751 not applied (or already rolled back).';
  END IF;
  SELECT count(*) INTO n FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config::text LIKE '%governor_admits_agent%';
  IF n <> 0 THEN RAISE EXCEPTION '751 ROLLBACK REFUSED: % live agent rows reference governor_admits_agent — remove the callers first.', n; END IF;
  SELECT count(*) INTO n FROM scheduled_tasks WHERE pre_query LIKE '%governor_admits_agent%';
  IF n <> 0 THEN RAISE EXCEPTION '751 ROLLBACK REFUSED: % scheduled tasks reference governor_admits_agent — remove the callers first.', n; END IF;
  -- Stage B's Go gate is opt-in per agent config under a key of its own; refuse if any row has it.
  SELECT count(*) INTO n FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config::text LIKE '%honour_spend_governor_run%';
  IF n <> 0 THEN RAISE EXCEPTION '751 ROLLBACK REFUSED: % live agent rows carry the stage-B run flag — roll stage B back first.', n; END IF;
END $$;

-- Restore governor_admits to the standalone 675 body (verbatim), then drop the factoring.
CREATE OR REPLACE FUNCTION governor_admits(p_item_type text) RETURNS boolean
LANGUAGE sql STABLE AS $FN$
  SELECT NOT COALESCE((
    SELECT gc.enabled
       AND COALESCE(m.llm_bearing, true)
       AND gs.shed_level >= CASE COALESCE(m.class, 'maintenance')
             WHEN 'maintenance' THEN 1
             WHEN 'build'       THEN 2
             ELSE                    3
           END
    FROM governor_config gc
    JOIN governor_state gs ON gs.id = 1
    LEFT JOIN governor_work_class_map m ON m.item_type = p_item_type
    WHERE gc.id = 1
  ), false)
$FN$;

COMMENT ON FUNCTION governor_admits(text) IS
'D4 spend governor (AGOV-013): TRUE unless the governor currently withholds this item_type.
The ONE canonical shed predicate — the Go loader/claim and the dispatch selector all call it;
do not re-spell the logic anywhere (council corr 8f4bb57d r1, architecture seat). Fail-open:
missing config/state rows admit everything. Unmapped types = maintenance+llm_bearing.';

DROP VIEW governor_withheld_runs_recent;
DROP TABLE governor_withheld_runs;
DROP FUNCTION governor_admits_agent(text);
DROP TABLE governor_agent_class_map;
DROP FUNCTION governor_admits_class(text, boolean);

-- Verify: the restored predicate still executes and still sheds the right cell.
DO $$
DECLARE saved_level int; saved_enabled boolean;
BEGIN
  SELECT shed_level INTO saved_level FROM governor_state WHERE id=1;
  SELECT enabled INTO saved_enabled FROM governor_config WHERE id=1;
  UPDATE governor_config SET enabled = true WHERE id=1;
  UPDATE governor_state SET shed_level = 1 WHERE id=1;
  INSERT INTO governor_work_class_map (item_type, class, llm_bearing, note)
  VALUES ('__751_rb_probe__', 'maintenance', true, 'transient')
  ON CONFLICT (item_type) DO UPDATE SET class='maintenance', llm_bearing=true;
  IF governor_admits('__751_rb_probe__') IS DISTINCT FROM false THEN
    RAISE EXCEPTION '751 ROLLBACK VERIFY: restored governor_admits does not shed a maintenance/bearing item at L1';
  END IF;
  DELETE FROM governor_work_class_map WHERE item_type='__751_rb_probe__';
  UPDATE governor_config SET enabled = saved_enabled WHERE id=1;
  UPDATE governor_state SET shed_level = saved_level WHERE id=1;
  IF EXISTS (SELECT 1 FROM pg_proc WHERE proname IN ('governor_admits_agent','governor_admits_class')) THEN
    RAISE EXCEPTION '751 ROLLBACK VERIFY: a D4b function survived the drop';
  END IF;
  RAISE NOTICE '751 ROLLBACK OK: governor_admits restored to the 675 body and proven to shed; D4b objects gone.';
END $$;

COMMIT;
