-- 546_css_patch_agent_refused_append_stops_minting_complete_ROLLBACK.sql
--
-- Reverses 546: removes mark_append_refused and points check_saved's else arm back at
-- complete_error.
--
-- ⚠ WHAT ROLLING BACK COSTS: a refused append — the model returning an empty or
-- oversized css_added, i.e. the founding 2026-08-04 failure mode — goes back to reading
-- `complete` on the work item. Nothing is written and nothing is deployed either way;
-- what returns is the false ledger entry, and with it the census problem (an item
-- counted as repaired that was never touched).

BEGIN;

SELECT snapshot_agent('css-patch-agent', '546_ROLLBACK: pre-revert');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,check_saved,config,else_step}',
         to_jsonb('complete_error'::text)
       ),
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,mark_append_refused}',
       updated_at = NOW()
 WHERE type = 'css-patch-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    v_steps jsonb;
    v_dangling int;
BEGIN
    SELECT default_config #> '{workflow,steps}'
      INTO v_steps
      FROM agent_definitions
     WHERE type = 'css-patch-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_steps ? 'mark_append_refused' THEN
        RAISE EXCEPTION '198/546 ROLLBACK: mark_append_refused survives';
    END IF;
    IF v_steps #>> '{check_saved,config,else_step}' <> 'complete_error' THEN
        RAISE EXCEPTION '198/546 ROLLBACK: check_saved.else_step not restored';
    END IF;

    SELECT count(*) INTO v_dangling
      FROM (
        SELECT e.v->>'next_step' AS tgt FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'next_step'
        UNION ALL
        SELECT e.v->>'error_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'error_step'
        UNION ALL
        SELECT e.v->'config'->>'then_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'then_step'
        UNION ALL
        SELECT e.v->'config'->>'else_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'else_step'
      ) AS edges
     WHERE tgt IS NOT NULL AND NOT (v_steps ? tgt);

    IF v_dangling > 0 THEN
        RAISE EXCEPTION '198/546 ROLLBACK: % dangling edge(s) left behind', v_dangling;
    END IF;

    RAISE NOTICE '198/546 ROLLBACK: verified — pre-546 shape restored (a refused append reads complete again)';
END $$;

COMMIT;
