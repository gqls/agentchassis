-- 453 ROLLBACK — disable the escalation and return any rows it moved.
--
-- Use when: the 3-day limit proves too aggressive, or escalated rows are
-- unwanted in the review queue. Two independent steps — the first stops future
-- ticks, the second undoes what past ticks did. Run both, or just the first if
-- you want to keep the escalations already made.

BEGIN;

-- 1. stop future ticks (keeps the row so its history and pre_query survive)
UPDATE scheduled_tasks SET enabled = false, updated_at = now()
 WHERE name = 'held-pair-canary-escalation';

-- 2. return every row this task moved, to exactly the state it was in.
--    Keyed on the task's own stamp, so it can touch nothing it did not write.
UPDATE site_work_items
SET status = 'detected',
    resolution_path = NULL,
    result = result - 'held_pair_escalation',
    updated_at = now()
WHERE status = 'needs_human_review'
  AND resolution_path = 'auto:held_pair_escalated'
  AND result ? 'held_pair_escalation';

DO $$
DECLARE n_enabled int; n_left int;
BEGIN
    SELECT count(*) INTO n_enabled FROM scheduled_tasks
     WHERE name='held-pair-canary-escalation' AND enabled;
    IF n_enabled <> 0 THEN
        RAISE EXCEPTION '453 ROLLBACK: the task is still enabled — the disable did not take';
    END IF;
    SELECT count(*) INTO n_left FROM site_work_items
     WHERE resolution_path='auto:held_pair_escalated' AND status='needs_human_review';
    IF n_left <> 0 THEN
        RAISE EXCEPTION '453 ROLLBACK: % escalated rows are still parked — the restore did not take', n_left;
    END IF;
    RAISE NOTICE '453 ROLLBACK: task disabled and all escalated rows returned to detected.';
END $$;

COMMIT;
