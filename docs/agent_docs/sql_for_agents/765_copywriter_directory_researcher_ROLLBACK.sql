-- 765 ROLLBACK — remove the copywriter-directory-researcher row. Safe: nothing else references
-- it until 766 retargets the scheduled task, and 766's own rollback restores that first.
BEGIN;
DO $r$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE target_agent_type='copywriter-directory-researcher' AND enabled;
  IF n > 0 THEN RAISE EXCEPTION '765 ROLLBACK REFUSED: % enabled scheduled task(s) still target this agent — roll back 766 first', n; END IF;
  DELETE FROM agent_definitions WHERE type='copywriter-directory-researcher' AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  RAISE NOTICE '765 ROLLBACK: removed % row(s)', n;
END $r$;
COMMIT;
