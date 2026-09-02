-- 658_dispatch_phase3_batch_8_HOLD.sql — Phase 3: dispatch batch 5 → 8, BOTH knobs in one file.
--
-- Raises build-dispatch-loop's per-turn batch from 5 to 8:
--   {workflow,steps,load_items,config,max_items}              5 → 8
--   {workflow,steps,process_item,config,max_iterations}       5 → 8
--
-- WHY _HOLD: ordering-critical. Committed 2026-08-26; applied BY HAND ~09:30Z 2026-08-27,
-- only after the 24h post-B read (owner-confirmed sequencing) passes its go/no-go gate
-- (cadence p50 <= ~65s, lost claims well below pre-B 58-60%, 0 true double-handles, 584
-- VERIFY green). SIDECAR_RE excludes _HOLD from the runner while still listing it.
--
-- WHY BOTH KNOBS TOGETHER (verified at the code, 2026-08-26): the loop action truncates
-- its collection at max_iterations (platform/orchestration/actions/loop_actions.go:197-203),
-- so max_items alone loads 3 surplus rows that are dropped from the slice (harmless — the
-- loader is a pure SELECT and claims nothing; they stay 'triaged') with only a pod-log Warn;
-- max_iterations alone is completely silent and inert. Either half alone is a THROUGHPUT
-- no-op, not a correctness hazard — but ship both.
--
-- ⚠ VALUES MUST BE BARE JSON NUMBERS. Both readers fall back to their Go defaults on a type
-- mismatch — load_work_item_actions.go:699 GetIntField default **50**, loop_actions.go:52-55
-- bare .(float64) default **20** — so a quoted "8" would silently run batch 50/20, not 8.
-- to_jsonb(8) below writes a jsonb number; the guard asserts jsonb_typeof = 'number'.
--
-- ⚠ ROLLBACK IS 658_dispatch_phase3_batch_8_ROLLBACK.sql — explicit jsonb_set back to 5.
-- NEVER remove the keys with '#-' (633's rollback shape): with these Go defaults, deleting
-- the keys sets the batch to 50/20, ten times the intended state.
--
-- Consumers checked 2026-08-26: nothing else reads either knob — no 584 VERIFY edit, no
-- optional-key-budget change (max_items is a config literal outside the ActionInputSpec by
-- design, load_work_item_actions.go:653-661), no test pins the values. The 413 fix thread
-- (657, selector rework) reads max_items LIVE from this row by agreement, so 658 needs no
-- lockstep there either. Do NOT re-run 051_build_dispatch_loop.sql — it would replace this
-- whole workflow with the ancient one-item config, not merely reset a knob.

BEGIN;

-- ALREADY-APPLIED REFUSAL **BEFORE** snapshot_agent (LANDMINES 2026-08-26: the house
-- template's order is backwards — a replay must refuse before it takes a decoy 'pre-update'
-- snapshot whose content is actually post-change; 526/541/633 all have this inverted).
DO $pre$
DECLARE v_mi text; v_it text;
BEGIN
    SELECT default_config#>>'{workflow,steps,load_items,config,max_items}',
           default_config#>>'{workflow,steps,process_item,config,max_iterations}'
      INTO v_mi, v_it
      FROM agent_definitions
     WHERE type='build-dispatch-loop'
       AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v_mi IS NULL OR v_it IS NULL THEN
        RAISE EXCEPTION '658: knob path missing (max_items=%, max_iterations=%) — the step shape has drifted since 2026-08-26; STOP and re-read the row', v_mi, v_it;
    END IF;
    IF v_mi = '8' AND v_it = '8' THEN
        RAISE EXCEPTION '658: already applied (both knobs read 8)';
    END IF;
    IF v_mi <> '5' OR v_it <> '5' THEN
        RAISE EXCEPTION '658: unexpected pre-state (max_items=%, max_iterations=%) — expected 5/5; concurrent edit or half-applied state; STOP and diagnose', v_mi, v_it;
    END IF;
END $pre$;

-- Negative-control baseline: NO other live agent row's config may change in this transaction.
CREATE TEMP TABLE _658_control AS
    SELECT id, md5(default_config::text) AS h
      FROM agent_definitions
     WHERE type <> 'build-dispatch-loop';

SELECT snapshot_agent('build-dispatch-loop', '658_dispatch_phase3_batch_8_HOLD.sql: pre-update');

DO $mig$
DECLARE v_n int;
BEGIN
    UPDATE agent_definitions ad
       SET default_config = jsonb_set(
             jsonb_set(ad.default_config,
                 '{workflow,steps,load_items,config,max_items}', to_jsonb(8)),
                 '{workflow,steps,process_item,config,max_iterations}', to_jsonb(8))
     WHERE ad.type='build-dispatch-loop'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '658: expected to update exactly 1 build-dispatch-loop row, updated %', v_n;
    END IF;
END $mig$;

-- ---------------------------------------------------------------------------
-- GUARDS — read back with queries this file's writer does not use, assert values,
-- types, config shapes, the untouched sub_workflow, and the fleet negative control.
-- ⚠ Never verify by agent_definitions.updated_at (degenerate — bulk touches) and never
-- by joining snapshot_agent()'s return value (it returns the SOURCE row id; LANDMINES).
DO $guard$
DECLARE v jsonb; v_keys int; v_ctrl int;
BEGIN
    SELECT s.value INTO v
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-dispatch-loop' AND s.key='load_items'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    IF v->'config'->'max_items' IS DISTINCT FROM to_jsonb(8)
       OR jsonb_typeof(v->'config'->'max_items') <> 'number' THEN
        RAISE EXCEPTION '658 GUARD: load_items.max_items reads % (type %) — expected jsonb number 8',
            v->'config'->'max_items', jsonb_typeof(v->'config'->'max_items');
    END IF;
    SELECT count(*) INTO v_keys FROM jsonb_object_keys(v->'config') k;
    IF v_keys <> 3 THEN
        RAISE EXCEPTION '658 GUARD: load_items.config has % keys, expected 3 (site_id, max_items, honour_site_lock) — the write disturbed a sibling key', v_keys;
    END IF;

    SELECT s.value INTO v
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-dispatch-loop' AND s.key='process_item'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    IF v->'config'->'max_iterations' IS DISTINCT FROM to_jsonb(8)
       OR jsonb_typeof(v->'config'->'max_iterations') <> 'number' THEN
        RAISE EXCEPTION '658 GUARD: process_item.max_iterations reads % (type %) — expected jsonb number 8',
            v->'config'->'max_iterations', jsonb_typeof(v->'config'->'max_iterations');
    END IF;
    SELECT count(*) INTO v_keys FROM jsonb_object_keys(v->'config') k;
    IF v_keys <> 5 THEN
        RAISE EXCEPTION '658 GUARD: process_item.config has % keys, expected 5 (items_field, item_variable, max_iterations, continue_on_error, sub_workflow) — the write disturbed a sibling key', v_keys;
    END IF;
    IF v->'config'->'sub_workflow'->>'start_step' IS DISTINCT FROM 'claim' THEN
        RAISE EXCEPTION '658 GUARD: process_item.sub_workflow.start_step reads % — expected claim; the nested write corrupted the sub_workflow', v->'config'->'sub_workflow'->>'start_step';
    END IF;

    -- NEGATIVE CONTROL: no other live row's config changed. (READ COMMITTED means another
    -- session's concurrent commit could false-RAISE here — that aborts THIS transaction
    -- harmlessly; re-run the file.)
    SELECT count(*) INTO v_ctrl
      FROM agent_definitions ad JOIN _658_control c ON c.id = ad.id
     WHERE md5(ad.default_config::text) <> c.h;
    IF v_ctrl <> 0 THEN
        RAISE EXCEPTION '658 GUARD (negative control): % other agent row(s) changed in this transaction — the UPDATE escaped its WHERE', v_ctrl;
    END IF;

    RAISE NOTICE '658 OK: build-dispatch-loop batch is 8/8 (max_items + max_iterations, both jsonb numbers), config shapes intact, sub_workflow untouched, no other row changed';
END $guard$;

COMMIT;
