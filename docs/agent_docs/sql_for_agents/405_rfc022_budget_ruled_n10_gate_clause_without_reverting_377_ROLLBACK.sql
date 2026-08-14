-- 405_rfc022_budget_ruled_n10_gate_clause_without_reverting_377_ROLLBACK.sql
-- Restores 403's "not yet ruled" closing sentences in council-gate. Inverse of
-- 405; breakpoint guarded, since a rollback can corrupt the cached prefix
-- exactly as easily as an apply.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('council-gate',
  '405_rfc022_budget_ruled_n10_gate_clause_without_reverting_377_ROLLBACK.sql: pre-rollback');

DO $rollback$
DECLARE
    v_new text := E'The budget N\n'
               || E'  is not yet ruled, so until it is: if the plan shows the action ALREADY\n'
               || E'  carries several optional keys, say so as "insufficient" with the count you\n'
               || E'  actually observed, citing the counter, which can give you the exact figure.';
    v_cur text := E'The budget is\n'
               || E'  RULED (owner, 2026-08-14): N = 10, on SHARED actions (2 or more live\n'
               || E'  carriers). If the plan grows such an action''s optional-key set past 10, or\n'
               || E'  grows one already past it, that ACCUMULATION is architecture-scope: signal\n'
               || E'  needs_rfc with the count you actually observed, citing the counter\n'
               || E'  (scripts/audit-optional-key-budget.sh). Sharing itself is estate design, not\n'
               || E'  the defect (owner, same ruling): what is reviewed is the accumulated\n'
               || E'  surface, never the reuse.';
    v_marker text := '<!--CACHE_BREAKPOINT-->';
    v_old    text;
    v_out    text;
    v_rows   int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_old
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_old IS NULL OR position(v_cur in v_old) = 0 THEN
        RAISE EXCEPTION '405 ROLLBACK: 405''s sentence not found verbatim — nothing safe to invert';
    END IF;

    v_out := replace(v_old, v_cur, v_new);
    IF position(v_marker in v_out) <> position(v_marker in v_old) THEN
        RAISE EXCEPTION '405 ROLLBACK: the cache breakpoint would move. Refusing.';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','review_architecture','config','prompt_template'],
             to_jsonb(v_out), false),
           updated_at = now()
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '405 ROLLBACK: expected 1 row, updated %', v_rows;
    END IF;
END
$rollback$;

COMMIT;
