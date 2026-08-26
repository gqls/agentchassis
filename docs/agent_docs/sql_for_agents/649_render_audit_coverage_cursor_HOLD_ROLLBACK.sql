-- 649 ROLLBACK — return request_render_audit to the deterministic prefix.
--
-- ORDER MATTERS AND IT IS THE ONLY SUBTLE THING HERE: remove the KEY first,
-- then the TABLE. Between the two statements the code must never be in the state
-- "rotation on, table gone" — it degrades and logs rather than failing, but it
-- would log a defect that is really a half-finished rollback, and a warning that
-- means nothing is worse than no warning. Both are in one transaction, so a
-- reader outside it sees neither state.
--
-- Rolling back restores today's behaviour EXACTLY: with the key absent the
-- action takes the first max_pages rows, which is what it did before 394.
-- Nothing else reads this table, so dropping it strands no consumer.

BEGIN;

UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,audit,config,rotate_coverage}',
    updated_at = NOW()
WHERE type = 'render-audit-agent'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

DROP TABLE IF EXISTS render_audit_page_cursor;

DO $$
DECLARE v_has int;
BEGIN
  SELECT count(*) INTO v_has
  FROM agent_definitions
  WHERE type = 'render-audit-agent'
    AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
    AND default_config->'workflow'->'steps'->'audit'->'config' ? 'rotate_coverage';
  IF v_has <> 0 THEN
    RAISE EXCEPTION '649 rollback: rotate_coverage is still present — aborting rather than leaving the flag on with no table';
  END IF;
END $$;

COMMIT;
