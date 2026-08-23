-- ROLLBACK for 574 — restores required-fields-missing-handler to its 410 (v3) shape:
-- deployed-only component resolution, `stale` reachable from a failed lookup, and the two
-- close-arm evidence strings as 410 wrote them. Removes the two steps 574 added.
--
-- ⚠ READ THIS BEFORE RUNNING IT. Rolling back REINSTATES bugs_open/367: a true finding
-- about a component that is real but not deployed will once again be closed `complete`
-- with no error, reported as "cannot be located on the live site". That failure is SILENT
-- — it is the reason the bug exists. If you are rolling back because something downstream
-- broke, prefer disarming the new route (point route_resolved.else_step back at
-- route_owned, leaving the honest resolution and evidence in place) over restoring the
-- deployed-only predicate.
--
-- Safe to run only while no item is mid-route: an in-flight orchestration that has already
-- classified `target_not_dispatchable` would find no step of that name after this runs.
-- Check first:
--   SELECT count(*) FROM orchestration_states
--    WHERE workflow_plan->>'start_step'='classify' AND status NOT IN ('COMPLETED','FAILED','CANCELLED');

BEGIN;

DO $$
DECLARE
    v_q     text;
    v_steps jsonb;
    v_n     int;
BEGIN
    SELECT default_config -> 'workflow' -> 'steps',
           default_config -> 'workflow' -> 'steps' -> 'classify' -> 'config' ->> 'query'
      INTO v_steps, v_q
      FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_q IS NULL THEN
        RAISE EXCEPTION '574 ROLLBACK: no active required-fields-missing-handler row';
    END IF;
    IF position('target_not_dispatchable' in v_q) = 0 THEN
        RAISE EXCEPTION '574 ROLLBACK: 574 is not applied (no target_not_dispatchable in classify) — nothing to undo';
    END IF;

    PERFORM snapshot_agent('required-fields-missing-handler'::text,
                           '574 ROLLBACK: restoring 410 v3 shape'::text);

    -- Reverse the five query edits, in reverse order, on the same verbatim anchors.
    v_q := replace(v_q,
        'CASE WHEN (SELECT count(*) FROM pg) = 0 THEN ''page_missing'' '
        'WHEN COALESCE((SELECT locked FROM comp), false) THEN ''component_locked'' '
        'WHEN (SELECT count(*) FROM comp) = 0 AND (SELECT retired FROM tomb) THEN ''component_retired'' '
        'WHEN (SELECT count(*) FROM comp) = 0 THEN ''lookup_miss'' '
        'WHEN COALESCE((SELECT bs FROM comp), '''') <> ''deployed'' THEN COALESCE((SELECT bs FROM comp), '''') '
        'ELSE '''' END AS target_state, ', '');
    v_q := replace(v_q,
        'WHEN (SELECT count(*) FROM pg) = 0 '
        'OR COALESCE((SELECT locked FROM comp), false) '
        'OR ((SELECT count(*) FROM comp) = 0 AND (SELECT retired FROM tomb)) THEN ''stale'' '
        'WHEN (SELECT n_still_empty FROM fs) = 0 THEN ''resolved'' '
        'WHEN (SELECT count(*) FROM comp) = 0 '
        'OR COALESCE((SELECT bs FROM comp), '''') <> ''deployed'' THEN ''target_not_dispatchable''',
        'WHEN (SELECT count(*) FROM pg) = 0 OR (SELECT count(*) FROM comp) = 0 OR COALESCE((SELECT locked FROM comp), false) THEN ''stale'' WHEN (SELECT n_still_empty FROM fs) = 0 THEN ''resolved''');
    v_q := replace(v_q,
        'tomb AS (SELECT EXISTS (SELECT 1 FROM page_components pc2 JOIN pg ON pc2.page_id = pg.id CROSS JOIN item '
        'WHERE COALESCE(pc2.build_status, ''pending'') = ''removed'' '
        'AND COALESCE(pc2.slot_name, '''') = COALESCE(item.spec->>''slot_name'', '''')) AS retired), ', '');
    v_q := replace(v_q,
        '(pc.locked_at IS NOT NULL) AS locked, cc.input_schema AS sch, COALESCE(pc.build_status, ''pending'') AS bs',
        '(pc.locked_at IS NOT NULL) AS locked, cc.input_schema AS sch');
    v_q := replace(v_q,
        'COALESCE(pc.build_status, ''pending'') <> ''removed''',
        'pc.build_status = ''deployed''');

    -- Every trace of 574 must be gone from the query, or the reverse was partial.
    IF position('target_not_dispatchable' in v_q) > 0
    OR position('target_state' in v_q) > 0
    OR position('tomb AS (' in v_q) > 0
    OR position('AS bs' in v_q) > 0
    OR position('pc.build_status = ''deployed''' in v_q) = 0 THEN
        RAISE EXCEPTION '574 ROLLBACK: the reverse rewrite was partial — do NOT commit; inspect the query by hand';
    END IF;

    UPDATE agent_definitions
       SET default_config =
           jsonb_set(
               jsonb_set(
                   jsonb_set(
                       jsonb_set(
                           (default_config #- '{workflow,steps,route_not_dispatchable}')
                                            #- '{workflow,steps,park_not_dispatchable}',
                           '{workflow,steps,classify,config,query}', to_jsonb(v_q), false),
                       '{workflow,steps,route_resolved,config,else_step}',
                       to_jsonb('route_owned'::text), false),
                   '{workflow,steps,close_stale,config,result_fields,evidence}',
                   to_jsonb(
                       'page or deployed component no longer exists at (page_name, slot_name), or the component '
                       'is now locked — the producer''s own predicate no longer matches. The dedup key releases; '
                       'discovery rotation (bugs_open/230, fixed 2026-08-09) re-raises within days if the finding '
                       'is still real.'::text), false),
               '{workflow,steps,close_resolved,config,result_fields,evidence}',
               to_jsonb(
                   'every field named in spec.missing_fields is populated on the currently-deployed component — '
                   'the same predicate the review-queue revalidator closes on'::text), false)
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- The classify description is restored separately for readability.
    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config, '{workflow,steps,classify,description}',
               to_jsonb(
                   'One deterministic classification per item, resolved by (page_name, slot_name) — the '
                   'revalidator''s own key, never spec.component_id (unstable across rerenders)'::text), false)
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- Verify the restored shape: back to 20 steps, cascade re-joined, predicate restored.
    SELECT default_config -> 'workflow' -> 'steps' INTO v_steps FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    SELECT count(*) INTO v_n FROM jsonb_object_keys(v_steps);
    IF v_n <> 20 THEN
        RAISE EXCEPTION '574 ROLLBACK: expected 20 steps after removing 2, found %', v_n;
    END IF;
    IF v_steps #>> '{route_resolved,config,else_step}' IS DISTINCT FROM 'route_owned' THEN
        RAISE EXCEPTION '574 ROLLBACK: the else-cascade was not re-joined — items would stop mid-route';
    END IF;
    IF v_steps #> '{route_not_dispatchable}' IS NOT NULL OR v_steps #> '{park_not_dispatchable}' IS NOT NULL THEN
        RAISE EXCEPTION '574 ROLLBACK: an added step survives';
    END IF;
    IF position('pc.build_status = ''deployed''' in (v_steps #>> '{classify,config,query}')) = 0 THEN
        RAISE EXCEPTION '574 ROLLBACK: the deployed-only predicate was not restored';
    END IF;

    RAISE NOTICE '574 ROLLBACK OK: 410 v3 shape restored, 20 steps. bugs_open/367 is REINSTATED and it is silent.';
END $$;

COMMIT;
