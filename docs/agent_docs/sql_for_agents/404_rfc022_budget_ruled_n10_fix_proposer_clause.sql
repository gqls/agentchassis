-- 404_rfc022_budget_ruled_n10_fix_proposer_clause.sql
--
-- RFC_022 follow-up to 402. The owner ruled the budget on 2026-08-14: N = 10,
-- closing the RFC. The clause 402 shipped says "The budget N is not yet ruled"
-- — falsified by the ruling — and its fallback instruction ("say so as
-- insufficient") is superseded by the real trigger: growth past 10 on a SHARED
-- action is architecture-scope.
--
-- The ruling's framing correction is part of the clause, deliberately: the
-- owner rejected "a shared action nobody understands" — sharing is estate
-- design (agents are meant to be somewhat independent and reusable across
-- workflows), so the seat is told what is reviewed is the ACCUMULATED SURFACE,
-- never the reuse. A seat prompt that reads sharing as a smell would penalise
-- the estate's own founding design.
--
-- 405 applies the byte-identical replacement to council-gate. Same surgical
-- anchored pattern as 381/383/402/403; 099_SYNC_gate_roster.py remains
-- suspended (it would revert 377).
--
-- ROLLBACK: 404_..._ROLLBACK.sql.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('fix-proposer',
  '404_rfc022_budget_ruled_n10_fix_proposer_clause.sql: pre-update');

DO $apply$
DECLARE
    v_anchor text := E'The budget N\n'
                  || E'  is not yet ruled, so until it is: if the plan shows the action ALREADY\n'
                  || E'  carries several optional keys, say so as "insufficient" with the count you\n'
                  || E'  actually observed, citing the counter, which can give you the exact figure.';
    v_insert text := E'The budget is\n'
                  || E'  RULED (owner, 2026-08-14): N = 10, on SHARED actions (2 or more live\n'
                  || E'  carriers). If the plan grows such an action''s optional-key set past 10, or\n'
                  || E'  grows one already past it, that ACCUMULATION is architecture-scope: signal\n'
                  || E'  needs_rfc with the count you actually observed, citing the counter\n'
                  || E'  (scripts/audit-optional-key-budget.sh). Sharing itself is estate design, not\n'
                  || E'  the defect (owner, same ruling): what is reviewed is the accumulated\n'
                  || E'  surface, never the reuse.';
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
        RAISE EXCEPTION '404: no live fix-proposer row, or review_architecture has no prompt_template';
    END IF;

    IF position('RULED (owner, 2026-08-14)' in v_old) > 0 THEN
        RAISE EXCEPTION '404: already applied — the ruled-budget sentence is present.';
    END IF;

    IF position(v_anchor in v_old) = 0 THEN
        RAISE EXCEPTION '404: 402''s closing sentences not found verbatim — the clause has been edited since. Re-read it and re-anchor rather than forcing this.';
    END IF;

    v_new := replace(v_old, v_anchor, v_insert);
    IF v_new = v_old THEN
        RAISE EXCEPTION '404: replace() was a no-op despite the anchor being present';
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
        RAISE EXCEPTION '404: expected to update exactly 1 fix-proposer row, updated %', v_rows;
    END IF;

    RAISE NOTICE '404: ruled-budget sentence applied to fix-proposer.review_architecture (% -> % chars)',
        length(v_old), length(v_new);
END
$apply$;

-- Verify by RAISE, never by SELECT.
DO $verify$
DECLARE
    v_txt text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_architecture'->'config'->>'prompt_template'
      INTO v_txt
      FROM agent_definitions
     WHERE type = 'fix-proposer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('RULED (owner, 2026-08-14)' in v_txt) = 0 THEN
        RAISE EXCEPTION '404 VERIFY: ruled-budget sentence absent after update';
    END IF;
    IF position('is not yet ruled' in v_txt) > 0 THEN
        RAISE EXCEPTION '404 VERIFY: the stale not-yet-ruled sentence survived the replace';
    END IF;
    IF position('NARROWED BY OWNER RULING 2026-08-11' in v_txt) = 0 THEN
        RAISE EXCEPTION '404 VERIFY: 381''s narrowing clause was lost';
    END IF;

    RAISE NOTICE '404 VERIFY: fix-proposer clause carries the ruled budget; narrowing intact';
END
$verify$;

COMMIT;
