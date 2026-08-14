-- 409_improvement_loop_calls_the_offer_analyser_HOLD_ROLLBACK.sql
--
-- Un-enrols the offer analyser from the improvement sweep: re-points
-- call_site_review straight back at record_audit_pass and drops the two steps
-- 409 added. Leaves the `offer-analyser` agent row alone (that is 408's
-- rollback) — after this the analyser still exists and still works when
-- hand-dispatched, it simply is not in the automatic sweep.
--
-- It does NOT restore from `bak_ad_improvementloop_20260814` wholesale, and that
-- is deliberate: other sessions edit this loop, so a wholesale restore of an
-- hours-old snapshot would silently revert THEIR work too. Two named pointers
-- and two named keys is the surgical undo. The snapshot is there for a human
-- who needs to see the before-state, not as the undo mechanism.

BEGIN;

DO $guard$
DECLARE
  steps jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps' INTO steps
  FROM agent_definitions
  WHERE type = 'improvement-loop' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF steps IS NULL THEN
    RAISE EXCEPTION 'guard: no live improvement-loop row';
  END IF;
  IF NOT (steps ? 'call_offer_analyser') THEN
    RAISE EXCEPTION 'guard: already rolled back (call_offer_analyser absent)';
  END IF;
  IF (steps->'call_offer_analyser'->>'next_step') IS DISTINCT FROM 'record_audit_pass' THEN
    RAISE EXCEPTION 'guard: call_offer_analyser now hands on to % — something was spliced AFTER it, so removing it would break that chain. Re-read the loop before rolling back', (steps->'call_offer_analyser'->>'next_step');
  END IF;
END $guard$;

UPDATE agent_definitions SET
  default_config = jsonb_set(
    default_config,
    '{workflow,steps}',
    ((default_config->'workflow'->'steps')
       - 'spawn_offer_analyser'
       - 'call_offer_analyser')
      || jsonb_build_object(
           'call_site_review',
           jsonb_set(
             jsonb_set(default_config->'workflow'->'steps'->'call_site_review',
                       '{next_step}', '"record_audit_pass"'),
             '{error_step}', '"record_audit_pass"')
         )
  ),
  updated_at = now()
WHERE type = 'improvement-loop' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $verify$
DECLARE
  steps jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps' INTO steps
  FROM agent_definitions
  WHERE type = 'improvement-loop' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF steps ? 'spawn_offer_analyser' OR steps ? 'call_offer_analyser' THEN
    RAISE EXCEPTION 'rollback verify: an offer-analyser step survives';
  END IF;
  IF (steps->'call_site_review'->>'next_step')  IS DISTINCT FROM 'record_audit_pass'
  OR (steps->'call_site_review'->>'error_step') IS DISTINCT FROM 'record_audit_pass' THEN
    RAISE EXCEPTION 'rollback verify: call_site_review does not rejoin record_audit_pass on both arms — the audit branch is now broken, which is worse than the enrolment was';
  END IF;
END $verify$;

COMMIT;
