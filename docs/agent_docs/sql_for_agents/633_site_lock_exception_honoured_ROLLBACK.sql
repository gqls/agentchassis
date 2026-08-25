-- 633_site_lock_exception_honoured_ROLLBACK.sql — reverses 633_..._HOLD.sql
--
-- Returns the site gate to the plain `WHERE s.locked_at IS NULL` and removes the
-- loader opt-in. Safe to run at any time: the result is the pre-396 behaviour,
-- which is a FULL hold on any locked site. Nothing is stranded either way,
-- because neither this file nor 633 ever touches a work-item row.
--
-- ⚠ Run this BEFORE 632's rollback if you are removing the column too, or the
-- gate is left referencing a column that no longer exists. 632's rollback
-- refuses while any live config still names it, so the order is enforced.

BEGIN;

SELECT snapshot_agent('build-pipeline-trigger', '633_ROLLBACK: pre-revert');
SELECT snapshot_agent('build-dispatch-loop',    '633_ROLLBACK: pre-revert');

DO $rb$
DECLARE
    v_new CONSTANT text := 'WHERE (s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[]))) AND wi.status IN';
    v_old CONSTANT text := 'WHERE s.locked_at IS NULL AND wi.status IN';
    v_q   text;
    v_n   int;
BEGIN
    SELECT s.value->'config'->>'query' INTO v_q
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-pipeline-trigger' AND s.key='find_dispatchable_site'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

    IF v_q IS NULL THEN
        RAISE EXCEPTION '633 ROLLBACK: find_dispatchable_site has no config.query';
    END IF;
    IF position(v_new in v_q) = 0 THEN
        RAISE EXCEPTION '633 ROLLBACK: the 633 clause is not present — already rolled back, or the step drifted; STOP';
    END IF;

    UPDATE agent_definitions ad
       SET default_config = jsonb_set(ad.default_config,
             '{workflow,steps,find_dispatchable_site,config,query}',
             to_jsonb(replace(v_q, v_new, v_old)))
     WHERE ad.type='build-pipeline-trigger'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 1 THEN RAISE EXCEPTION '633 ROLLBACK: updated % trigger rows, expected 1', v_n; END IF;

    UPDATE agent_definitions ad
       SET default_config = #- '{workflow,steps,load_items,config,honour_site_lock}'
     WHERE ad.type='build-dispatch-loop'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 1 THEN RAISE EXCEPTION '633 ROLLBACK: updated % dispatch-loop rows, expected 1', v_n; END IF;
END;
$rb$;

DO $guard$
DECLARE
    v_gate text;
    v_optin jsonb;
BEGIN
    SELECT s.value->'config'->>'query' INTO v_gate
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-pipeline-trigger' AND s.key='find_dispatchable_site'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    IF position('lock_except_item_ids' in COALESCE(v_gate,'')) > 0 THEN
        RAISE EXCEPTION '633 ROLLBACK GUARD: the gate still names lock_except_item_ids';
    END IF;
    IF position('s.locked_at IS NULL' in COALESCE(v_gate,'')) = 0 THEN
        RAISE EXCEPTION '633 ROLLBACK GUARD: the gate has no locked_at test at all — the lock is OFF; restore from the snapshot taken above';
    END IF;

    SELECT s.value->'config'->'honour_site_lock' INTO v_optin
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-dispatch-loop' AND s.key='load_items'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    IF v_optin IS NOT NULL THEN
        RAISE EXCEPTION '633 ROLLBACK GUARD: honour_site_lock is still present (%)', v_optin;
    END IF;

    RAISE NOTICE '633 ROLLBACK OK: gate back to a plain locked_at test, loader opt-in removed.';
END;
$guard$;

COMMIT;
