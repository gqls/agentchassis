-- 277_feature_designer_seat_caps_to_16000.sql
-- OWNER CALL, 2026-07-31: raise review_architecture / review_editquality /
-- review_guidelines from 8000 to 16000 on feature-designer, matching fix-proposer
-- and council-gate.
--
-- WHY, and why it needed an owner call rather than being an obvious repair.
-- The 2026-07-29 owner ruling raised these seats' caps on the strength of measured
-- truncations, with the stated criterion "leave the others until they actually
-- truncate". It was applied to fix-proposer and mirrored to council-gate by
-- 099_SYNC_gate_roster.py — which mirrors those two councils and ONLY those two.
-- feature-designer holds the same three seat NAMES and was reached by nothing, so
-- the ruling landed on two of the three councils that hold them.
--
-- That gap was invisible to both existing checks, which is the interesting part:
--   * 099 compares fix-proposer against council-gate, so a value missing from BOTH
--     — or from a third council entirely — reads as "in sync".
--   * 102_LINT_council_seat_parity compares each seat against its OWN council's
--     family and deliberately declines cross-council comparison, because councils
--     legitimately differ (different remits, different seat sets, an owner ruling
--     that experience-planner omits tolerate_truncation). Within feature-designer
--     all six seats sat at 8000 — perfectly uniform, nothing to flag.
-- Found by bugs_open/138 candidate 2's report (FIX-058), whose section 4 prints
-- cross-council divergence as INFORMATION next to the truncation evidence.
--
-- ON THE EVIDENCE, STATED HONESTLY. feature-designer's OWN calls have never
-- truncated: 4 per seat, 20 review calls total, all 2026-07-26/27. So the 07-29
-- criterion is NOT met by this council's own data, and the case for raising is a
-- transfer from the sibling councils at the same cap — where review_editquality and
-- review_guidelines both truncated at 8000 and review_architecture truncated 2 of
-- its first 3 reviews. Same seat names, largely shared prompts. The owner made that
-- call on 2026-07-31 after the divergence and its evidence were put in front of him.
--
-- WHAT THIS DOES NOT DO. The 07-29 fix to review_architecture had THREE parts: the
-- cap, `notes` moved ahead of `objections` (so the mandated ARCHITECTURE_SIGNAL
-- survives a cut), and a length budget in the prompt. This propagates the CAP only.
-- feature-designer's review_architecture prompt is still the pre-fix one — verified
-- by diff: it lacks both the length block and the notes-first output order, while
-- legitimately differing elsewhere (it judges a design, not a fix, and renders
-- {{.spec_row.spec_json}}). Those two halves change what a reviewer is asked to do
-- rather than how much room it has, so they are a separate decision, deliberately
-- not bundled here. See bugs_open/138, the 2026-07-31 entry.
--
-- SAFETY
--   * snapshot_agent() first — lands in agent_definitions_backup (NOT an
--     is_snapshot row in agent_definitions; see LANDMINES, footprint snapshot_agent).
--   * jsonb_set with create_if_missing := false, so a wrong path is a SILENT no-op
--     rather than an error. The path is config.ai_service.max_tokens — max_tokens is
--     NESTED inside ai_service, unlike prompt_template which is its sibling.
--   * Guarded on the current value being 8000, so this is idempotent and cannot
--     stomp a different value another session has since set.
--   * The RETURNING count is the check. Expect exactly 3 rows.
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < 277_feature_designer_seat_caps_to_16000.sql

BEGIN;

SELECT snapshot_agent('feature-designer',
  'pre-update: 277 raise architecture/editquality/guidelines caps 8000->16000 (owner call 2026-07-31, bugs_open/138)') AS snapshot_id;

WITH upd AS (
  UPDATE agent_definitions a
  SET default_config = jsonb_set(
        jsonb_set(
          jsonb_set(a.default_config,
            '{workflow,steps,review_architecture,config,ai_service,max_tokens}', '16000'::jsonb, false),
          '{workflow,steps,review_editquality,config,ai_service,max_tokens}',   '16000'::jsonb, false),
        '{workflow,steps,review_guidelines,config,ai_service,max_tokens}',      '16000'::jsonb, false),
      updated_at = now()
  WHERE a.type = 'feature-designer'
    AND a.is_active AND COALESCE(a.is_snapshot,false) = false AND a.deleted_at IS NULL
    -- Guard: only move a seat that is still at the old value. Idempotent, and it
    -- cannot overwrite a deliberate different setting made since this was written.
    AND a.default_config #>> '{workflow,steps,review_architecture,config,ai_service,max_tokens}' = '8000'
    AND a.default_config #>> '{workflow,steps,review_editquality,config,ai_service,max_tokens}'  = '8000'
    AND a.default_config #>> '{workflow,steps,review_guidelines,config,ai_service,max_tokens}'   = '8000'
  RETURNING a.id
)
SELECT count(*) AS rows_updated_expect_1 FROM upd;

-- Verification, inside the transaction: all three must read 16000 on all three
-- councils, i.e. nine rows and no 8000 among them.
SELECT a.type, s.key, (s.value->'config'->'ai_service'->>'max_tokens')::int AS cap
FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
WHERE a.is_active AND COALESCE(a.is_snapshot,false) = false AND a.deleted_at IS NULL
  AND s.key IN ('review_architecture','review_editquality','review_guidelines')
ORDER BY s.key, a.type;

COMMIT;
