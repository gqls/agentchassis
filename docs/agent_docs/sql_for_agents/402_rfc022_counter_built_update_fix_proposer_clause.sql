-- 402_rfc022_counter_built_update_fix_proposer_clause.sql
--
-- RFC_022 follow-up to 381. The clause 381 inserted into
-- fix-proposer.review_architecture states "that counter is not built yet" —
-- true when written, FALSE once the optional-key-budget counter merged
-- (cmd/config-key-audit --optional-key-budget + scripts/audit-optional-key-budget.sh,
-- register WFA-013, bugfix_223_index_answerability lane, 2026-08-13). A seat
-- reading a prompt that asserts a mechanism does not exist, when it does, is
-- the stale-status class this estate keeps paying for — so the sentence is
-- updated by the same surgical anchored pattern that inserted it, NOT by
-- 099_SYNC_gate_roster.py, which remains suspended (it would revert 377).
--
-- WHAT CHANGES: only the clause's closing sentences. The exception's three
-- conditions, the accumulation warning, and the reduced "insufficient" signal
-- all stand; the seat is now told the counter exists and can supply the exact
-- figure. The budget N itself is deliberately NOT stated: it awaits the owner's
-- ruling, and a prompt asserting an unruled threshold would be inventing policy.
--
-- 403 applies the byte-identical replacement to council-gate, so the two
-- rosters keep saying the same thing about the same trigger.
--
-- ROLLBACK: 402_..._ROLLBACK.sql.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('fix-proposer',
  '402_rfc022_counter_built_update_fix_proposer_clause.sql: pre-update');

DO $apply$
DECLARE
    v_anchor text := E'  ruling''s destination is a BUDGET on the accumulated optional-key COUNT, but\n'
                  || E'  that counter is not built yet. So: if the plan shows the action ALREADY\n'
                  || E'  carries several optional keys, say so as "insufficient" with the count you\n'
                  || E'  actually observed. That reduced signal is the part still worth having.';
    v_insert text := E'  ruling''s destination is a BUDGET on the accumulated optional-key COUNT. The\n'
                  || E'  counter is BUILT (2026-08-13, register WFA-013): scripts/audit-optional-key-budget.sh,\n'
                  || E'  the --optional-key-budget mode of cmd/config-key-audit, reports each shared\n'
                  || E'  action''s declared optional-key count beside its live carriers. The budget N\n'
                  || E'  is not yet ruled, so until it is: if the plan shows the action ALREADY\n'
                  || E'  carries several optional keys, say so as "insufficient" with the count you\n'
                  || E'  actually observed, citing the counter, which can give you the exact figure.';
    v_old  text;
    v_new  text;
    v_rows int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_old
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_old IS NULL THEN
        RAISE EXCEPTION '402: no live fix-proposer row, or review_architecture has no prompt_template';
    END IF;

    IF position('counter is BUILT (2026-08-13' in v_old) > 0 THEN
        RAISE EXCEPTION '402: already applied — the counter-built sentence is present.';
    END IF;

    IF position(v_anchor in v_old) = 0 THEN
        RAISE EXCEPTION '402: 381''s closing sentences not found verbatim — the clause has been edited since. Re-read it and re-anchor rather than forcing this.';
    END IF;

    v_new := replace(v_old, v_anchor, v_insert);
    IF v_new = v_old THEN
        RAISE EXCEPTION '402: replace() was a no-op despite the anchor being present';
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
        RAISE EXCEPTION '402: expected to update exactly 1 fix-proposer row, updated %', v_rows;
    END IF;

    RAISE NOTICE '402: counter-built sentence applied to fix-proposer.review_architecture (% -> % chars)',
        length(v_old), length(v_new);
END
$apply$;

-- Verify by RAISE, never by SELECT (a SELECT-shaped verify block lets a failed
-- migration COMMIT under ON_ERROR_STOP).
DO $verify$
DECLARE
    v_txt text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_txt
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('counter is BUILT (2026-08-13' in v_txt) = 0 THEN
        RAISE EXCEPTION '402 VERIFY: counter-built sentence absent after update';
    END IF;
    IF position('NARROWED BY OWNER RULING 2026-08-11' in v_txt) = 0 THEN
        RAISE EXCEPTION '402 VERIFY: 381''s narrowing clause was lost';
    END IF;
    IF position('that counter is not built yet' in v_txt) > 0 THEN
        RAISE EXCEPTION '402 VERIFY: the stale sentence survived the replace';
    END IF;

    RAISE NOTICE '402 VERIFY: fix-proposer clause now names the counter; narrowing intact';
END
$verify$;

COMMIT;
