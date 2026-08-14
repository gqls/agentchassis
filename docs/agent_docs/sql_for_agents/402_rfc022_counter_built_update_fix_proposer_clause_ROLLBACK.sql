-- 402_rfc022_counter_built_update_fix_proposer_clause_ROLLBACK.sql
-- Restores 381's original closing sentences ("that counter is not built yet")
-- in fix-proposer.review_architecture. Inverse of 402: same anchor discipline.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('fix-proposer',
  '402_rfc022_counter_built_update_fix_proposer_clause_ROLLBACK.sql: pre-rollback');

DO $rollback$
DECLARE
    v_new text := E'  ruling''s destination is a BUDGET on the accumulated optional-key COUNT, but\n'
               || E'  that counter is not built yet. So: if the plan shows the action ALREADY\n'
               || E'  carries several optional keys, say so as "insufficient" with the count you\n'
               || E'  actually observed. That reduced signal is the part still worth having.';
    v_cur text := E'  ruling''s destination is a BUDGET on the accumulated optional-key COUNT. The\n'
               || E'  counter is BUILT (2026-08-13, register WFA-013): scripts/audit-optional-key-budget.sh,\n'
               || E'  the --optional-key-budget mode of cmd/config-key-audit, reports each shared\n'
               || E'  action''s declared optional-key count beside its live carriers. The budget N\n'
               || E'  is not yet ruled, so until it is: if the plan shows the action ALREADY\n'
               || E'  carries several optional keys, say so as "insufficient" with the count you\n'
               || E'  actually observed, citing the counter, which can give you the exact figure.';
    v_old  text;
    v_rows int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_old
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_old IS NULL OR position(v_cur in v_old) = 0 THEN
        RAISE EXCEPTION '402 ROLLBACK: 402''s sentence not found verbatim — nothing safe to invert';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
             ARRAY['workflow','steps','review_architecture','config','prompt_template'],
             to_jsonb(replace(v_old, v_cur, v_new)), false),
           updated_at = now()
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '402 ROLLBACK: expected 1 row, updated %', v_rows;
    END IF;
END
$rollback$;

COMMIT;
