-- 766 ROLLBACK — put the task back on directory-researcher and disable it (its pre-766 state).
-- Note this restores a task that CANNOT produce copywriter entries; it is a rollback, not a repair.
BEGIN;
UPDATE scheduled_tasks
   SET target_agent_type='directory-researcher', enabled=false, updated_at=NOW()
 WHERE name='copywriter-directory-discovery';
DO $v$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name='copywriter-directory-discovery'
                   AND target_agent_type='directory-researcher' AND NOT enabled)
  THEN RAISE EXCEPTION '766 ROLLBACK VERIFY: task not restored'; END IF;
END $v$;
COMMIT;
