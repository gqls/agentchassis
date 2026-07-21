-- PATCH 018 — feature-designer: scope expected_symbols to the stage's OWN files
--
-- Answers the delta-2 council-gate objection (submission 5a65ec4c, editquality +
-- guardian, both medium): expected_symbols' verbatim-substring gate can
-- false-reject a CORRECT stage whose named symbol is defined in an EARLIER
-- stage's file (a symbol the stage only CALLS, not one it introduces). The
-- honest fix is at the SOURCE — steer the designer to name only symbols the
-- stage's own produced files introduce — NOT to weaken the deterministic gate
-- (diagnose_prepare_fix_commit's missingExpectedSymbols check stays intact).
--
-- SURGICAL by construction: a single replace() on rule 8's exact text inside the
-- design step's prompt_template. Every other byte of the prompt, and every other
-- step, is untouched. The WHERE guard makes it idempotent (0 rows if already
-- patched or if the row moved). Snapshot the row before applying (below).

-- 1. Snapshot first (discipline: co-edited rows).
SELECT snapshot_agent('feature-designer', 'PATCH_018_expected_symbols_scope');

-- 2. Surgical replace of rule 8 only.
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,design,config,prompt_template}',
    to_jsonb( replace(
        default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template',
        '8. expected_symbols per stage: names (functions, registry keys, SQL step names) that must literally appear in that stage''s produced files — the deterministic gate a reviewer can trust.',
        '8. expected_symbols per stage: names (functions, registry keys, SQL step names) that must literally appear in THIS stage''s OWN produced files — the deterministic gate a reviewer can trust. CRITICAL: list ONLY symbols this stage''s edited files INTRODUCE. NEVER list a symbol your stage merely CALLS or references that is defined in an earlier stage''s file, or in any file this stage does not edit — the gate checks each expected_symbol appears verbatim in a file THIS stage produced, so naming a cross-file symbol false-rejects a correct stage. A stage that only wires prior work and introduces no new named symbol uses an empty expected_symbols list.'
    ) )
)
WHERE type = 'feature-designer'
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template'
      LIKE '%that must literally appear in that stage''s produced files — the deterministic gate a reviewer can trust.%';

-- 3. Verify: old text gone, new guidance present, exactly one active row changed.
SELECT (default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template' LIKE '%INTRODUCE. NEVER list a symbol your stage merely CALLS%') AS has_new_guidance,
       (default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template' LIKE '%appear in that stage''s produced files%') AS still_has_old_text
FROM agent_definitions
WHERE type='feature-designer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
