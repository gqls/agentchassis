-- 288 — repoint the error_step that 286 left dangling
--
-- CORRECTION TO 286, applied 2026-08-02 minutes after it. Forward-only.
--
-- WHAT WENT WRONG. 286 deleted the `triage` step from design-audit-agent after
-- repointing `call_content_auditor.next_step` -> `complete`. It repointed the
-- SUCCESS edge and missed the ERROR edge: `call_content_auditor` also carries
-- `error_step: "triage"`, so a failing content-auditor call now routes to a step
-- that no longer exists.
--
-- HOW IT WAS FOUND, and the real lesson. 286's own verify block asked exactly
-- the right question — check (iii), "does anything still POINT at the deleted
-- step", covering next_step, error_step, then_step and else_step — and it
-- ANSWERED CORRECTLY, returning one row naming this exact step. It did not stop
-- the migration, because a `SELECT` is not an assertion: psql prints the row and
-- carries on to `COMMIT`, and `-v ON_ERROR_STOP=1` does not help because a
-- non-empty result set is not an error. **A verify block made of SELECTs is a
-- report you have to read, not a guard.** Where the check must actually hold,
-- it has to RAISE — which is what STEP 2 below does, so this file cannot repeat
-- the mistake it is fixing.
--
-- WHY `complete` IS THE RIGHT TARGET. The edge meant "if the content-auditor
-- call fails, do not abandon the run — carry on to triage, then finish". Triage
-- is now improvement-loop's job (RFC 006), so the surviving intent is just "do
-- not abandon the run": go to `complete`, which is where the success edge now
-- goes too. The alternative — leaving the run to strand — would be a behaviour
-- change 286 never intended and never stated.
--
-- SCOPE. One row, one key. site-review-agent was checked in the same query and
-- has no dangling reference: its `write_strategic_findings` carries no
-- error_step. The check below is written fleet-wide anyway, so it would catch a
-- third case if one existed.

BEGIN;

-- ── STEP 1 — THE FIX ──────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,call_content_auditor,error_step}',
      '"complete"'::jsonb,
      false),
    updated_at = now()
WHERE type = 'design-audit-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,call_content_auditor,error_step}' = 'triage';

-- ── STEP 2 — ENFORCING CHECK (this is the point of the file) ──────────────
-- Not a SELECT. If any live step still routes to a step its own workflow does
-- not define, this RAISEs and the transaction aborts. 286's advisory version of
-- this same question is why 288 exists.
DO $$
DECLARE
    dangling text;
BEGIN
    SELECT string_agg(t.agent || '.' || t.step_name || ' -> ' || t.target, ', ')
    INTO dangling
    FROM (
        SELECT ad.type AS agent,
               step.key AS step_name,
               tgt.target AS target
        FROM agent_definitions ad,
             jsonb_each(ad.default_config->'workflow'->'steps') AS step,
             LATERAL (VALUES (step.value->>'next_step'),
                             (step.value->>'error_step'),
                             (step.value->'config'->>'then_step'),
                             (step.value->'config'->>'else_step')) AS tgt(target)
        WHERE ad.is_active
          AND COALESCE(ad.is_snapshot, false) = false
          AND ad.deleted_at IS NULL
          AND ad.type IN ('design-audit-agent', 'site-review-agent', 'improvement-loop')
          AND tgt.target IS NOT NULL
          AND tgt.target <> ''
          -- the target must be a key of the SAME workflow's step map
          AND NOT (ad.default_config #> '{workflow,steps}') ? tgt.target
    ) t;

    IF dangling IS NOT NULL THEN
        RAISE EXCEPTION 'dangling step reference(s) remain: %', dangling;
    END IF;

    RAISE NOTICE 'no dangling step references across the three RFC-006 agents';
END $$;

COMMIT;

-- ── ROLLBACK ──
-- UPDATE agent_definitions SET default_config = jsonb_set(default_config,
--   '{workflow,steps,call_content_auditor,error_step}', '"triage"'::jsonb, false)
-- WHERE type='design-audit-agent' AND is_active
--   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- (Only meaningful alongside a rollback of 286, which restores the triage step.)
--
-- ── THE REUSABLE CHECK ──
-- The DO block above is worth running against the WHOLE fleet, not just these
-- three agents — a dangling edge is invisible until the run that takes it, and
-- most edges are error edges that almost never fire. Drop the `ad.type IN (...)`
-- line to widen it. Read-only in that form; run it before believing any
-- step-deleting migration is finished.
