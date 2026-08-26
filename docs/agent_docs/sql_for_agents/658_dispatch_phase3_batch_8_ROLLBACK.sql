-- 658_dispatch_phase3_batch_8_ROLLBACK.sql — restore dispatch batch 8 → 5, both knobs.
--
-- ⚠ THIS FILE SETS THE VALUES BACK TO 5 EXPLICITLY. Do NOT "roll back" by deleting the keys
-- ('#-', 633's rollback shape): the Go defaults are 50 (load_work_item_actions.go:699) and
-- 20 (loop_actions.go:52-55), so key deletion would run batch 50/20 — ten times Phase 3.

BEGIN;

-- Refusal BEFORE snapshot (LANDMINES 2026-08-26 ordering rule).
DO $pre$
DECLARE v_mi text; v_it text;
BEGIN
    SELECT default_config#>>'{workflow,steps,load_items,config,max_items}',
           default_config#>>'{workflow,steps,process_item,config,max_iterations}'
      INTO v_mi, v_it
      FROM agent_definitions
     WHERE type='build-dispatch-loop'
       AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v_mi = '5' AND v_it = '5' THEN
        RAISE EXCEPTION '658 ROLLBACK: already rolled back (both knobs read 5)';
    END IF;
    IF v_mi <> '8' OR v_it <> '8' THEN
        RAISE EXCEPTION '658 ROLLBACK: unexpected state (max_items=%, max_iterations=%) — expected 8/8; STOP and diagnose', v_mi, v_it;
    END IF;
END $pre$;

SELECT snapshot_agent('build-dispatch-loop', '658_dispatch_phase3_batch_8_ROLLBACK.sql: pre-rollback');

DO $mig$
DECLARE v_n int;
BEGIN
    UPDATE agent_definitions ad
       SET default_config = jsonb_set(
             jsonb_set(ad.default_config,
                 '{workflow,steps,load_items,config,max_items}', to_jsonb(5)),
                 '{workflow,steps,process_item,config,max_iterations}', to_jsonb(5))
     WHERE ad.type='build-dispatch-loop'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '658 ROLLBACK: expected to update exactly 1 row, updated %', v_n;
    END IF;
END $mig$;

DO $guard$
DECLARE v_mi jsonb; v_it jsonb;
BEGIN
    SELECT default_config#>'{workflow,steps,load_items,config,max_items}',
           default_config#>'{workflow,steps,process_item,config,max_iterations}'
      INTO v_mi, v_it
      FROM agent_definitions
     WHERE type='build-dispatch-loop'
       AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v_mi IS DISTINCT FROM to_jsonb(5) OR v_it IS DISTINCT FROM to_jsonb(5)
       OR jsonb_typeof(v_mi) <> 'number' OR jsonb_typeof(v_it) <> 'number' THEN
        RAISE EXCEPTION '658 ROLLBACK GUARD: knobs read %/% — expected jsonb numbers 5/5', v_mi, v_it;
    END IF;
    RAISE NOTICE '658 ROLLBACK OK: build-dispatch-loop batch restored to 5/5';
END $guard$;

COMMIT;
