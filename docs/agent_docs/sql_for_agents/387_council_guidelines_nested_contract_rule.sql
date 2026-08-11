-- 387 - council guidelines seat: the nested-contract ruling becomes seat-visible
-- (fix-proposer seat; council-gate follows via SEED 388, not the 099 mirror; DECISIONS_2026-08-11 ruling 3's "still owed" half).
--
-- Why: the 2026-08-11 owner ruling resolved the guidelines seat's own flagged
-- ambiguity (corr a06ff850) - nested additions to a declared object input are
-- register-named, not re-declared. Until the seat can READ that ruling, it will
-- keep flagging the same shape. The 247 pattern put rulings in this prompt's
-- load-bearing rules list; this appends the new ruling to the DECLARED
-- CONTRACTS rule it refines, by anchored replacement (never a whole-prompt
-- rewrite - that is one drift behind live the moment it is written).
--
-- COUNCIL-GATE HALF: seed 388, a surgical anchored insert in the 383 pattern -
-- NOT 099_SYNC_gate_roster.py. LANDMINE 2026-08-11: 099 --apply silently
-- REVERTS migration 377 (its transform predates the cache-breakpoint hoist),
-- so the CLAUDE.md mirror rule's documented exception applies until 099 is
-- taught 377. Apply 387 then 388 in the same session; their clauses are
-- byte-identical, which is what keeps the rosters in agreement this round.

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_387_fix_proposer AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'fix-proposer' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $do$
DECLARE
    t text; c1 int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_guidelines'->'config'->>'prompt_template'
    INTO t FROM agent_definitions
    WHERE type = 'fix-proposer' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF t IS NULL THEN
        RAISE EXCEPTION '387: fix-proposer review_guidelines prompt_template not found';
    END IF;
    -- Dual-active-row landmine guard (council round d1e8c36e objection): four
    -- agent types carry TWO active rows and only the higher version loads.
    -- Refuse unless fix-proposer has EXACTLY ONE active non-snapshot row, so
    -- "UPDATE 1 + verify passed" cannot describe a row the loader never reads.
    IF (SELECT count(*) FROM agent_definitions
         WHERE type = 'fix-proposer' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL) <> 1 THEN
        RAISE EXCEPTION '387: fix-proposer does not have exactly one active row - resolve the duplicate before seeding';
    END IF;
    IF position('NESTED-FIELD ADDITIONS' in t) > 0 THEN
        RAISE EXCEPTION '387: already applied on fix-proposer';
    END IF;
    c1 := (length(t) - length(replace(t, $a1$A field a workflow step READS still requires the declared contract.$a1$, ''))) / length($a1$A field a workflow step READS still requires the declared contract.$a1$);
    IF c1 <> 1 THEN
        RAISE EXCEPTION '387: anchor count must be 1 on fix-proposer, got % - prompt drifted; regenerate from a fresh dump', c1;
    END IF;
END $do$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,review_guidelines,config,prompt_template}',
        to_jsonb(
            replace(
                default_config->'workflow'->'steps'->'review_guidelines'->'config'->>'prompt_template',
                $a1$A field a workflow step READS still requires the declared contract.$a1$,
                $r1$A field a workflow step READS still requires the declared contract. NESTED-FIELD ADDITIONS to an ALREADY-DECLARED object input (owner ruling 2026-08-11, resolving this seat's flagged ambiguity on corr a06ff850): adding keys INSIDE an object the contract already passes (worked case: facts_scoped/assigned_writer_block riding inside section_plan's sections_ready entries) does NOT require contract re-declaration - but it MUST be named in the seam's concept-register entry in the commit that ships it (the RFC_010 s1 shape: the register names it so a reader can see the seam's real shape without reading every call site). OBJECT to a nested addition only when it is NEITHER register-named NOR contract-declared; do not ask for re-declaration of a declared object's internals.$r1$)
        )
    ),
    updated_at = NOW()
WHERE type = 'fix-proposer' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $do$
DECLARE
    t text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_guidelines'->'config'->>'prompt_template'
    INTO t FROM agent_definitions
    WHERE type = 'fix-proposer' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF position($v1$NESTED-FIELD ADDITIONS to an ALREADY-DECLARED object input$v1$ in t) = 0 THEN
        RAISE EXCEPTION '387: verify failed on fix-proposer - the nested-contract rule is missing';
    END IF;
    IF length(t) <> 8695 THEN
        RAISE EXCEPTION '387: verify failed on fix-proposer - post length % <> expected 8695', length(t);
    END IF;
END $do$;

COMMIT;

-- ROLLBACK recipe (hand-run):
--   UPDATE agent_definitions ad SET default_config = b.default_config
--   FROM agent_definitions_bak_387_fix_proposer b WHERE ad.id = b.id;
