-- 392 ROLLBACK — restore the render-audit rotation's max_pages to 25.
-- Note the truncation-honesty code (bugs_open/242) is independent of this value;
-- rolling back the cap only shrinks the sweep, it does not remove the reporting.

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,audit,config,max_pages}', '25'::jsonb),
    updated_at = NOW()
WHERE type = 'render-audit-agent'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

DO $$
DECLARE
  v_max int;
BEGIN
  SELECT (default_config->'workflow'->'steps'->'audit'->'config'->>'max_pages')::int
    INTO v_max
  FROM agent_definitions
  WHERE type = 'render-audit-agent'
    AND is_active
    AND COALESCE(is_snapshot, false) = false
    AND deleted_at IS NULL;

  IF v_max IS DISTINCT FROM 25 THEN
    RAISE EXCEPTION '392 ROLLBACK: max_pages is % (expected 25) — aborting', v_max;
  END IF;
END $$;

COMMIT;
