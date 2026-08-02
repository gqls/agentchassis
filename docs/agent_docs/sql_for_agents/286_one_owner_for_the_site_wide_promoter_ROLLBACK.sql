-- 286 ROLLBACK — restore the `triage` step to design-audit-agent and
--                site-review-agent (undo RFC 006 option (a))
--
-- Hand-run only. The UPPERCASE suffix puts this under the migration runner's
-- SIDECAR_RE, so `--apply` lists it and never executes it — which is the whole
-- point of a rollback file (a rollback that can be swept up by a bulk apply is
-- a loaded gun, see migration-runner practice).
--
-- Written as a SEPARATE FILE at the council's request: `debug_historian`
-- objected (medium, corr 60f4b425) that 286 bundled snapshot + guarded update +
-- verify in one file with the rollback only as commented-out SQL, and that the
-- needle-gate discipline wants verify and rollback as their own artefacts. The
-- seat was right — a rollback you have to uncomment is one you cannot run under
-- pressure without editing it first.
--
-- ══════════════════════════════════════════════════════════════════════════
-- PREFER THE SNAPSHOT. This file is the fallback.
-- ══════════════════════════════════════════════════════════════════════════
-- 286 called the 2-arg `snapshot_agent(text, text)` overload, which writes a
-- true point-in-time copy into `agent_definitions_backup` (it does NOT write an
-- `is_snapshot` row in agent_definitions — the two overloads write to two
-- different tables, which is its own landmine). Confirmed retrievable
-- 2026-08-02, and confirmed to contain the PRE-change shape rather than merely
-- to exist:
--
--   SELECT type, snapshot_taken_at, left(snapshot_reason,45) AS reason,
--          (default_config #> '{workflow,steps}') ? 'triage' AS has_triage_step
--   FROM agent_definitions_backup
--   WHERE type IN ('design-audit-agent','site-review-agent')
--   ORDER BY snapshot_taken_at DESC LIMIT 2;
--   -- 2026-08-02 10:27:07.357419+00, has_triage_step = t on BOTH rows.
--
-- "A snapshot exists" and "a snapshot holds what you need" are different
-- claims. The `has_triage_step` column is the second one; check it before
-- relying on either path.
--
-- WHEN YOU WOULD RUN THIS. Only if putting the promoter back into both children
-- is genuinely wanted — i.e. the owner ruling is reversed. Note that restoring
-- the fan-out RE-OPENS bugs_closed/150's mechanism unless
-- improvement-loop.check_has_findings still reads `site_dispatchable` (migration
-- 281). Check that first:
--
--   SELECT default_config #>> '{workflow,steps,check_has_findings,config,condition}'
--   FROM agent_definitions WHERE type='improvement-loop' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- must read 'triage_result.site_dispatchable == true'. If it reads
--   -- has_items, do NOT run this file — you would restore the exact bug.

BEGIN;

-- ── STEP 1 — PRE-FLIGHT ───────────────────────────────────────────────────
-- Expect 2: both children present and currently WITHOUT the step. If this is
-- not 2, the rows are not in the state 286 left them in — stop and read them.
SELECT count(*) AS rows_to_restore_expect_2
FROM agent_definitions
WHERE type IN ('design-audit-agent', 'site-review-agent')
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND NOT (default_config #> '{workflow,steps}') ? 'triage';

-- ── STEP 2 — SNAPSHOT (yes, even on the way back) ─────────────────────────
SELECT snapshot_agent('design-audit-agent',  'rollback of 286 — restoring the triage step');
SELECT snapshot_agent('site-review-agent',   'rollback of 286 — restoring the triage step');

-- ── STEP 3a — design-audit-agent ──────────────────────────────────────────
-- Restores: the step, the success edge INTO it, the error edge into it (which
-- 286 missed on the way out and 288 repointed — see WRONG_CALLS.md 2026-08-02),
-- and `triage_result` in the published output_fields.
UPDATE agent_definitions
SET default_config =
      jsonb_set(
        jsonb_set(
          jsonb_set(
            jsonb_set(default_config,
              '{workflow,steps,triage}',
              '{"action":"triage_detected_items","config":{"site_id":"site_record.site_id","target_domain":"build"},"next_step":"complete","output_field":"triage_result"}'::jsonb,
              true),
            '{workflow,steps,call_content_auditor,next_step}', '"triage"'::jsonb, false),
          '{workflow,steps,call_content_auditor,error_step}', '"triage"'::jsonb, false),
        '{workflow,steps,complete,config,output_fields}',
        (default_config #> '{workflow,steps,complete,config,output_fields}') || '"triage_result"'::jsonb,
        false),
    updated_at = now()
WHERE type = 'design-audit-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND NOT (default_config #> '{workflow,steps}') ? 'triage'   -- idempotent
  AND NOT (default_config #> '{workflow,steps,complete,config,output_fields}') @> '"triage_result"'::jsonb;

-- ── STEP 3b — site-review-agent ───────────────────────────────────────────
-- No error edge here: write_strategic_findings never carried one.
UPDATE agent_definitions
SET default_config =
      jsonb_set(
        jsonb_set(
          jsonb_set(default_config,
            '{workflow,steps,triage}',
            '{"action":"triage_detected_items","config":{"site_id":"site_record.site_id","target_domain":"build"},"next_step":"complete","output_field":"triage_result"}'::jsonb,
            true),
          '{workflow,steps,write_strategic_findings,next_step}', '"triage"'::jsonb, false),
        '{workflow,steps,complete,config,output_fields}',
        (default_config #> '{workflow,steps,complete,config,output_fields}') || '"triage_result"'::jsonb,
        false),
    updated_at = now()
WHERE type = 'site-review-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND NOT (default_config #> '{workflow,steps}') ? 'triage'
  AND NOT (default_config #> '{workflow,steps,complete,config,output_fields}') @> '"triage_result"'::jsonb;

-- ── STEP 4 — ENFORCING VERIFY ─────────────────────────────────────────────
-- A DO block, not a SELECT. 286's verify was SELECTs; it found its own defect,
-- printed it, and COMMITted anyway, because a non-empty result set is not an
-- error and ON_ERROR_STOP does not fire on one. Do not repeat that here.
DO $$
DECLARE
    carriers int;
    dangling text;
BEGIN
    SELECT count(DISTINCT ad.type) INTO carriers
    FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS step
    WHERE ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
      AND step.value->>'action' = 'triage_detected_items';

    IF carriers <> 3 THEN
        RAISE EXCEPTION 'expected 3 carriers after rollback (the pre-286 fan-out), found %', carriers;
    END IF;

    SELECT string_agg(t.agent || '.' || t.step_name || ' -> ' || t.target, ', ') INTO dangling
    FROM (
        SELECT ad.type AS agent, step.key AS step_name, tgt.target AS target
        FROM agent_definitions ad,
             jsonb_each(ad.default_config->'workflow'->'steps') AS step,
             LATERAL (VALUES (step.value->>'next_step'), (step.value->>'error_step'),
                             (step.value->'config'->>'then_step'),
                             (step.value->'config'->>'else_step')) AS tgt(target)
        WHERE ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
          AND tgt.target IS NOT NULL AND tgt.target <> ''
          AND NOT (ad.default_config #> '{workflow,steps}') ? tgt.target
    ) t;

    IF dangling IS NOT NULL THEN
        RAISE EXCEPTION 'rollback left dangling step reference(s): %', dangling;
    END IF;

    RAISE NOTICE 'rollback verified: 3 carriers, no dangling step references fleet-wide';
END $$;

COMMIT;

-- ── AFTER RUNNING THIS ──
-- Re-run the fleet detector; it SHOULD now report the violation again, and that
-- is the correct outcome — it is the state RFC 006 ruled against:
--   ./scripts/audit-single-owner-actions.sh    # expect 1 finding, 3 carriers, exit 1
-- If it reports none, the rollback did not take effect and the DO block above
-- was bypassed somehow. Also drop `SingleOwner: true` from
-- TriageDetectedItemsInputSpec in the same change, or the detector will keep
-- failing a state the owner has (by hypothesis) chosen to return to.
