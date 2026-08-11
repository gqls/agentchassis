-- 383_rfc022_narrow_the_gate_architecture_seat_without_reverting_377_ROLLBACK.sql
--
-- Removes the RFC_022 narrowing clause from council-gate.review_architecture,
-- restoring the seat's pre-2026-08-11 trigger. Asserts that 377's cache
-- breakpoint does not move, exactly as the forward file does — the whole point
-- of doing this by hand rather than by mirror is that 377 must survive both
-- directions.
--
-- Pair with 381_..._ROLLBACK.sql if you are reversing the ruling entirely;
-- running only one of the two leaves the rosters saying different things.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('council-gate',
  '383_..._ROLLBACK.sql: pre-rollback');

DO $rollback$
DECLARE
    v_anchor   text := 'relocate it. Say so via ARCHITECTURE_SIGNAL: needs_rfc. Two live precedents:';
    v_marker   text := '<!--CACHE_BREAKPOINT-->';
    v_old      text;
    v_new      text;
    v_start    int;
    v_end      int;
    v_rows     int;
    v_mark_old int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_old
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_old IS NULL THEN
        RAISE EXCEPTION '383 ROLLBACK: no live council-gate row';
    END IF;

    v_mark_old := position(v_marker in v_old);
    IF v_mark_old = 0 THEN
        RAISE EXCEPTION '383 ROLLBACK: 377 marker already absent — stop and investigate rather than cutting further';
    END IF;

    v_start := position(E'\n\n  NARROWED BY OWNER RULING 2026-08-11' in v_old);
    IF v_start = 0 THEN
        RAISE EXCEPTION '383 ROLLBACK: narrowing clause not present — nothing to reverse';
    END IF;

    v_end := position('bugs_closed/124' in v_old);
    IF v_end = 0 OR v_end <= v_start THEN
        RAISE EXCEPTION '383 ROLLBACK: cannot locate the clause end marker; refusing to cut blind';
    END IF;

    v_new := substring(v_old from 1 for v_start - 1)
          || E'\n'
          || substring(v_old from v_end);

    IF position(v_anchor in v_new) = 0 THEN
        RAISE EXCEPTION '383 ROLLBACK: the anchor line did not survive the cut';
    END IF;
    IF position('NARROWED BY OWNER RULING 2026-08-11' in v_new) > 0 THEN
        RAISE EXCEPTION '383 ROLLBACK: clause still present after the cut';
    END IF;
    IF position(v_marker in v_new) <> v_mark_old THEN
        RAISE EXCEPTION '383 ROLLBACK: the cache breakpoint MOVED — refusing';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','review_architecture','config','prompt_template'],
             to_jsonb(v_new), false),
           updated_at = now()
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '383 ROLLBACK: expected 1 row, updated %', v_rows;
    END IF;

    RAISE NOTICE '383 ROLLBACK: clause removed (% -> % chars); breakpoint unmoved', length(v_old), length(v_new);
END
$rollback$;

COMMIT;
