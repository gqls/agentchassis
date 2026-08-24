-- 585_bug161_arm_artifact_check_canary_on_gd_trials_ROLLBACK.sql
-- HAND-RUN COMPANION — never applied by the runner (uppercase suffix, SIDECAR_RE).
--
-- Undoes 585: reinstates the register row 585 superseded (the one WITHOUT
-- artifact_check on gd-trials) as current, and marks 585's row superseded.
-- Written per the council's debug_historian advisory on the 585 round
-- (2026-08-24, corr a9e1a0de): the natural rollback should be a written
-- artefact, not an implied one.
--
-- NOTE: if the daily evidence-freshness sweep has rewritten the register
-- SINCE 585 applied (it bumps verified_at on a passing artifact_check), the
-- current row is the SWEEP's, not 585's — this file refuses in that case,
-- because blindly reinstating the pre-585 row would also revert the sweep's
-- legitimate updates. Roll back by hand from history if that happens.

BEGIN;

DO $$
DECLARE cur_created_by text; n int;
BEGIN
    SELECT created_by INTO cur_created_by FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf'
      AND aspect='evidence_base' AND is_current;
    IF cur_created_by IS DISTINCT FROM 'session-2026-08-24-bug161-artifact-check-canary' THEN
        RAISE EXCEPTION 'rollback refused: the current register row was written by %, not by 585 — the sweep or another session has moved the register since. Roll back by hand from history.', coalesce(cur_created_by, '(null)');
    END IF;

    SELECT count(*) INTO n FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf'
      AND aspect='evidence_base' AND is_current = false AND superseded_at IS NOT NULL
      AND NOT (data->'facts' @> '[{"id":"gd-trials","source":{"artifact_check":{}}}]'::jsonb);
    IF n < 1 THEN
        RAISE EXCEPTION 'rollback refused: no superseded row without artifact_check found to reinstate';
    END IF;
END $$;

-- retire 585's row
UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf'
  AND aspect='evidence_base' AND is_current
  AND created_by = 'session-2026-08-24-bug161-artifact-check-canary';

-- reinstate the newest pre-585 row
UPDATE site_specs
SET is_current = true, superseded_at = NULL
WHERE id = (
    SELECT id FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf'
      AND aspect='evidence_base' AND is_current = false
      AND NOT (data->'facts' @> '[{"id":"gd-trials","source":{"artifact_check":{}}}]'::jsonb)
    ORDER BY superseded_at DESC NULLS LAST
    LIMIT 1
);

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf'
      AND aspect='evidence_base' AND is_current;
    IF n IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION 'rollback: expected exactly 1 current row after reinstatement, found %', coalesce(n::text,'(null)');
    END IF;
    SELECT count(*) INTO n FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf'
      AND aspect='evidence_base' AND is_current
      AND data->'facts' @> '[{"id":"gd-trials","source":{"artifact_check":{}}}]'::jsonb;
    IF n IS DISTINCT FROM 0 THEN
        RAISE EXCEPTION 'rollback: the reinstated current row still carries artifact_check';
    END IF;
END $$;

COMMIT;
