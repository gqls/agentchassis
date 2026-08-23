-- ROLLBACK for 576 — removes the tomb CTE's empty-slot guard, restoring the 574 shape.
--
-- ⚠ Rolling back re-opens the edge the council flagged: an item carrying no `slot_name`
-- could be matched against a removed row that also carries no `slot_name`, anywhere on the
-- same page, and CLOSED as `stale` on it. [MEASURED 2026-08-23] not reachable — 0 of 38
-- removed rows and 0 page_components rows anywhere have an empty slot_name — so this is a
-- guard rather than a live fix, and rolling it back is safe TODAY. Re-measure before
-- assuming that still holds.

BEGIN;

DO $$
DECLARE
    v_q text; v_new_q text;
    c_new CONSTANT text :=
        'AND COALESCE(pc2.slot_name, '''') = COALESCE(item.spec->>''slot_name'', '''') '
        'AND COALESCE(item.spec->>''slot_name'', '''') <> '''') AS retired)';
    c_old CONSTANT text :=
        'AND COALESCE(pc2.slot_name, '''') = COALESCE(item.spec->>''slot_name'', '''')) AS retired)';
BEGIN
    SELECT default_config -> 'workflow' -> 'steps' -> 'classify' -> 'config' ->> 'query'
      INTO v_q FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_q IS NULL THEN RAISE EXCEPTION '576 ROLLBACK: no active row'; END IF;
    IF position(c_new in v_q) = 0 THEN
        RAISE EXCEPTION '576 ROLLBACK: 576 is not applied — nothing to undo';
    END IF;

    PERFORM snapshot_agent('required-fields-missing-handler'::text,
                           '576 ROLLBACK: removing the tomb empty-slot guard'::text);

    v_new_q := replace(v_q, c_new, c_old);
    IF v_new_q = v_q THEN RAISE EXCEPTION '576 ROLLBACK: no-op rewrite'; END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
               '{workflow,steps,classify,config,query}', to_jsonb(v_new_q), false)
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    SELECT default_config -> 'workflow' -> 'steps' -> 'classify' -> 'config' ->> 'query'
      INTO v_q FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF position('AND COALESCE(item.spec->>''slot_name'', '''') <> ''''' in v_q) > 0 THEN
        RAISE EXCEPTION '576 ROLLBACK: the guard survives — the reverse was partial';
    END IF;
    IF position('tomb AS (SELECT EXISTS' in v_q) = 0 OR position('target_not_dispatchable' in v_q) = 0 THEN
        RAISE EXCEPTION '576 ROLLBACK: 574''s changes were damaged — do NOT commit';
    END IF;
    RAISE NOTICE '576 ROLLBACK OK: back to the 574 tomb shape.';
END $$;

COMMIT;
