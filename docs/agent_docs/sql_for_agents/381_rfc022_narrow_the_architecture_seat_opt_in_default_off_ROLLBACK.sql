-- 381_rfc022_narrow_the_architecture_seat_opt_in_default_off_ROLLBACK.sql
--
-- Reverses 381 by removing the inserted narrowing clause from
-- fix-proposer.review_architecture's prompt_template, restoring the seat's
-- pre-2026-08-11 trigger (needs_rfc on the SHAPE of a new reserved key on a
-- shared action, with no opt-in exception).
--
-- Anchored on the inserted text and RAISE-guarded in both directions, so a
-- rollback against a prompt that has since been rewritten aborts rather than
-- half-applying.
--
-- AFTER THIS FILE, RE-RUN THE MIRROR so council-gate follows fix-proposer back:
--   python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/099_SYNC_gate_roster.py --apply

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('fix-proposer',
  '381_..._ROLLBACK.sql: pre-rollback');

DO $rollback$
DECLARE
    v_anchor text := 'relocate it. Say so via ARCHITECTURE_SIGNAL: needs_rfc. Two live precedents:';
    v_old    text;
    v_new    text;
    v_start  int;
    v_end    int;
    v_rows   int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_old
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_old IS NULL THEN
        RAISE EXCEPTION '381 ROLLBACK: no live fix-proposer row';
    END IF;

    v_start := position(E'\n\n  NARROWED BY OWNER RULING 2026-08-11' in v_old);
    IF v_start = 0 THEN
        RAISE EXCEPTION '381 ROLLBACK: narrowing clause not present — nothing to reverse (already rolled back, or the prompt was rewritten).';
    END IF;

    -- The clause runs to the end of its final sentence; everything after that
    -- is the original prompt continuing at "bugs_closed/124".
    v_end := position('bugs_closed/124' in v_old);
    IF v_end = 0 OR v_end <= v_start THEN
        RAISE EXCEPTION '381 ROLLBACK: cannot locate the clause end marker; refusing to cut blind';
    END IF;

    v_new := substring(v_old from 1 for v_start - 1)
          || E'\n'
          || substring(v_old from v_end);

    IF position(v_anchor in v_new) = 0 THEN
        RAISE EXCEPTION '381 ROLLBACK: the anchor line did not survive the cut';
    END IF;
    IF position('NARROWED BY OWNER RULING 2026-08-11' in v_new) > 0 THEN
        RAISE EXCEPTION '381 ROLLBACK: clause still present after the cut';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','review_architecture','config','prompt_template'],
             to_jsonb(v_new), false),
           updated_at = now()
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '381 ROLLBACK: expected 1 row, updated %', v_rows;
    END IF;

    RAISE NOTICE '381 ROLLBACK: clause removed (% -> % chars)', length(v_old), length(v_new);
END
$rollback$;

COMMIT;
