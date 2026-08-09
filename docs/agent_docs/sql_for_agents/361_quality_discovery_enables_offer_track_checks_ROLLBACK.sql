-- 361_quality_discovery_enables_offer_track_checks_ROLLBACK.sql
--
-- Inverse of 361: removes premise_incomplete + revenue_shape from
-- quality-discovery-agent's run_checks array, returning it to the 7-entry
-- form (the state 361's pre-assertion pinned).
--
-- Safe to run alone: detection simply stops; already-filed items stay where
-- they are (born 'detected', nothing dispatches them). This is the FIRST half
-- of a full B3 rollback — pair with 358_ROLLBACK only after this one, so the
-- timeout exclusions never uncover live verifiers.
--
-- The pre-assertion pins the exact 9-entry array 361 produced; if another lane
-- has appended a 10th name by rollback time, this fails loudly — re-read and
-- recompose rather than force it.
--
-- Written 2026-08-09 (round-2 council objection: needle-gate discipline wants a
-- separate rollback file, not just a fail-closed forward file).

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('quality-discovery-agent',
    'pre-update: 361_ROLLBACK — offer-track checks premise_incomplete + revenue_shape removed from the checks array');

DO $pre$
DECLARE v_checks jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,run_checks,config,checks}' INTO v_checks
    FROM agent_definitions
    WHERE type = 'quality-discovery-agent'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_checks IS DISTINCT FROM
       '["broken_nav_links","placeholder_contact","generic_theme","unverified_claims","voice_tells","literal_markdown","decision_guards","premise_incomplete","revenue_shape"]'::jsonb THEN
        RAISE EXCEPTION '361_ROLLBACK: pre-state mismatch — run_checks.checks is not the 9-entry array 361 produced (%). Re-read and recompose.', v_checks;
    END IF;
END $pre$;

UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,run_checks,config,checks}',
    '["broken_nav_links","placeholder_contact","generic_theme","unverified_claims","voice_tells","literal_markdown","decision_guards"]'::jsonb
)
WHERE type = 'quality-discovery-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $post$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'quality-discovery-agent'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,run_checks,config,checks}' =
          '["broken_nav_links","placeholder_contact","generic_theme","unverified_claims","voice_tells","literal_markdown","decision_guards"]'::jsonb;
    IF n <> 1 THEN
        RAISE EXCEPTION '361_ROLLBACK: post-state mismatch (% rows carry the 7-entry array)', n;
    END IF;
END $post$;

COMMIT;
