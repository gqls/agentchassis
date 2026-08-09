-- 361_quality_discovery_enables_offer_track_checks.sql
--
-- Enables B3's two offer-track discovery checks (vigilant_designer_offer_analysis
-- lane) by adding their registered names to quality-discovery-agent's run_checks
-- array: premise_incomplete (check_premise_incomplete.go) and revenue_shape
-- (check_revenue_shape.go), both shipped in ad51ca863 + b26fdc81b.
--
-- ORDERING, VERIFIED BEFORE WRITING (2026-08-09): an unregistered name in this
-- array is FATAL since bugfix_149 B4, so image strictly precedes config. Both
-- replicas of v1.0.1276 carry the checks — positive pod-grep
-- (VerifyRevenueShapeCTAResolved = 2, name literals present) AND negative
-- control (the literal b26fdc81b removed greps 0; a sibling SQL literal greps 1,
-- proving the grep can see SQL strings).
--
-- IMP-016 OBSERVE-ONLY: enabling detection dispatches nothing — items are born
-- 'detected' and nothing promotes them until a sweep's triage does; the lane
-- reads what they file before letting one promote. needs_strategy routing is
-- safe post-B2 only (gate amended by 359 to the shipped predicate).
--
-- The pre-assertion pins the EXACT live 7-entry array read 2026-08-09 —
-- decision_guards is the RFC_015 lane's addition since this lane's handoff said
-- 6 names; if the array has moved again, this fails loudly: re-read + recompose.
--
-- ROLLBACK: inverse jsonb_set back to the 7-entry array, same shape.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('quality-discovery-agent',
    'pre-update: 361 — offer-track checks premise_incomplete + revenue_shape enabled (observe-only; image v1.0.1276 pod-verified first)');

DO $pre$
DECLARE v_checks jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,run_checks,config,checks}' INTO v_checks
    FROM agent_definitions
    WHERE type = 'quality-discovery-agent'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_checks IS DISTINCT FROM
       '["broken_nav_links","placeholder_contact","generic_theme","unverified_claims","voice_tells","literal_markdown","decision_guards"]'::jsonb THEN
        RAISE EXCEPTION '361: pre-state mismatch — run_checks.checks is not the 7-entry array this file targets (%). Re-read and recompose.', v_checks;
    END IF;
END $pre$;

UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,run_checks,config,checks}',
    (default_config #> '{workflow,steps,run_checks,config,checks}')
        || '["premise_incomplete","revenue_shape"]'::jsonb
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
          '["broken_nav_links","placeholder_contact","generic_theme","unverified_claims","voice_tells","literal_markdown","decision_guards","premise_incomplete","revenue_shape"]'::jsonb;
    IF n <> 1 THEN
        RAISE EXCEPTION '361: post-state mismatch (% rows carry the 9-entry array)', n;
    END IF;
END $post$;

COMMIT;
