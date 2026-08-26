-- 660 ROLLBACK — remove the order-intake collector agent + schedule.
--
-- Safe at any time: the task ships disabled, and a mid-flight collector run
-- is idempotent end to end (paid gate, ON CONFLICT insert, dedup-keyed work
-- items, idempotent ack), so removal strands nothing. Uncollected briefs
-- simply stay on the box.

BEGIN;
DELETE FROM scheduled_tasks WHERE name = 'order-intake-collect';
UPDATE agent_definitions
   SET is_active = false, deleted_at = now()
 WHERE type = 'order-intake-collector' AND deleted_at IS NULL;
COMMIT;
