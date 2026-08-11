-- 388 - council-gate guidelines seat: the nested-contract ruling, mirrored BY
-- HAND in the 383 pattern - deliberately NOT via 099_SYNC_gate_roster.py.
--
-- The gate half of seed 387 (fix-proposer). LANDMINE 2026-08-11:
-- 099_SYNC_gate_roster.py --apply silently REVERTS migration 377 (the cache
-- breakpoint hoist, 68% measured saving) because its transform predates it -
-- the dry run reads "drift: all 17 seats" which is the gate being AHEAD of the
-- mirror, not divergence. Until 099 learns 377, the safe mirror is a surgical
-- insert at a verbatim anchor in ONE seat, guarded so it cannot touch the
-- cached prefix or fire twice (the 381+383 worked pair; this file is the 383
-- of the 387+388 pair). This is the documented exception to CLAUDE.md's
-- "do not hand-patch the gate", not a licence to ignore it - fix 099 and the
-- exception ends.
--
-- Same clause as 387, so the two rosters say the same thing about the same
-- trigger. Anchor measured live 2026-08-11: gate seat anchor at char 2015,
-- breakpoint at 174 (insert is far past the cached prefix), anchor count 1,
-- seat is fix-proposer's twin +37 chars (377's own arithmetic).

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('council-gate',
  '388_council_gate_guidelines_nested_contract_rule.sql: pre-update');

CREATE TABLE IF NOT EXISTS agent_definitions_bak_388 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'council-gate' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $apply$
DECLARE
    v_anchor   text := $a1$A field a workflow step READS still requires the declared contract.$a1$;
    v_marker   text := '<!--CACHE_BREAKPOINT-->';
    v_old      text;
    v_new      text;
    c1         int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_guidelines'->'config'->>'prompt_template'
      INTO v_old
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_old IS NULL THEN
        RAISE EXCEPTION '388: no live council-gate row, or review_guidelines has no prompt_template';
    END IF;
    -- Dual-active-row landmine guard (council round d1e8c36e objection): four
    -- agent types carry TWO active rows and only the higher version loads.
    -- Refuse unless council-gate has EXACTLY ONE active non-snapshot row, so
    -- "UPDATE 1 + verify passed" cannot describe a row the loader never reads.
    IF (SELECT count(*) FROM agent_definitions
         WHERE type = 'council-gate' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL) <> 1 THEN
        RAISE EXCEPTION '388: council-gate does not have exactly one active row - resolve the duplicate before seeding';
    END IF;
    IF position('NESTED-FIELD ADDITIONS' in v_old) > 0 THEN
        RAISE EXCEPTION '388: already applied - the nested-contract rule is present';
    END IF;
    c1 := (length(v_old) - length(replace(v_old, v_anchor, ''))) / length(v_anchor);
    IF c1 <> 1 THEN
        RAISE EXCEPTION '388: anchor count must be 1 on the gate seat, got % - re-read and re-anchor rather than forcing', c1;
    END IF;
    IF position(v_marker in v_old) = 0 THEN
        RAISE EXCEPTION '388: the 377 cache breakpoint is ABSENT from this seat - something has already reverted 377; stop and investigate';
    END IF;
    IF position(v_marker in v_old) >= position(v_anchor in v_old) THEN
        RAISE EXCEPTION '388: the anchor precedes the cache breakpoint - inserting here WOULD change the cached prefix. Refusing';
    END IF;

    v_new := replace(v_old, v_anchor, $r1$A field a workflow step READS still requires the declared contract. NESTED-FIELD ADDITIONS to an ALREADY-DECLARED object input (owner ruling 2026-08-11, resolving this seat's flagged ambiguity on corr a06ff850): adding keys INSIDE an object the contract already passes (worked case: facts_scoped/assigned_writer_block riding inside section_plan's sections_ready entries) does NOT require contract re-declaration - but it MUST be named in the seam's concept-register entry in the commit that ships it (the RFC_010 s1 shape: the register names it so a reader can see the seam's real shape without reading every call site). OBJECT to a nested addition only when it is NEITHER register-named NOR contract-declared; do not ask for re-declaration of a declared object's internals.$r1$);

    UPDATE agent_definitions
       SET default_config = jsonb_set(
               default_config,
               '{workflow,steps,review_guidelines,config,prompt_template}',
               to_jsonb(v_new)),
           updated_at = NOW()
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
END $apply$;

DO $verify$
DECLARE
    t text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'review_guidelines'->'config'->>'prompt_template'
      INTO t
      FROM agent_definitions
     WHERE type = 'council-gate' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF position($v1$NESTED-FIELD ADDITIONS to an ALREADY-DECLARED object input$v1$ in t) = 0 THEN
        RAISE EXCEPTION '388: verify failed - the rule is missing from the gate seat';
    END IF;
    IF position('<!--CACHE_BREAKPOINT-->' in t) <> 174 THEN
        RAISE EXCEPTION '388: verify failed - the cache breakpoint MOVED (position %), the cached prefix was disturbed', position('<!--CACHE_BREAKPOINT-->' in t);
    END IF;
    IF length(t) <> 8732 THEN
        RAISE EXCEPTION '388: verify failed - post length % <> expected 8732', length(t);
    END IF;
END $verify$;

COMMIT;

-- ROLLBACK recipe (hand-run):
--   UPDATE agent_definitions ad SET default_config = b.default_config
--   FROM agent_definitions_bak_388 b WHERE ad.id = b.id;
