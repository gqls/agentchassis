-- 458_aiao_imagery_policy_people_allowed_not_as_staff_ROLLBACK.sql
--
-- Restores the pre-ruling design_intent for ai-agent-orchestration.com by promoting
-- the row 458 superseded, and closing 458's successor.
--
-- ⚠ THIS REINSTATES A POLICY THE OWNER SUPERSEDED on 2026-08-18. The restored text
-- says "Technical illustrations and architectural diagrams ONLY … never staged
-- corporate photography", which forbids the people-at-work photography the owner
-- explicitly permitted, and restores an `avoid` line that bans a CAROUSEL rather
-- than the impersonation. Roll back only to undo a bad write, never to re-litigate
-- the ruling.
--
-- Identifies 458's successor by created_by, so a later legitimate supersession by
-- another author is not clobbered.

BEGIN;

-- 1. Close 458's row.
UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND aspect = 'design_intent'
  AND is_current
  AND created_by = 'session:site_ai_agent_orchestration';

-- 2. Promote the row it had superseded (the newest one NOT written by 458).
UPDATE site_specs
SET is_current = true, superseded_at = NULL
WHERE id = (
  SELECT id FROM site_specs
  WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
    AND aspect = 'design_intent'
    AND created_by <> 'session:site_ai_agent_orchestration'
    AND superseded_at IS NOT NULL
  ORDER BY superseded_at DESC
  LIMIT 1);

DO $$
DECLARE
  current_rows int;
  has_old      int;
BEGIN
  SELECT count(*) INTO current_rows FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='design_intent' AND is_current;
  IF current_rows <> 1 THEN
    RAISE EXCEPTION 'rollback 458: expected exactly 1 current design_intent row, found % — the partial unique index would already have refused a second, so 0 means nothing was promoted', current_rows;
  END IF;

  SELECT count(*) INTO has_old FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='design_intent' AND is_current
     AND data->'avoid' @> '["Testimonial carousels with headshots of fake people"]'::jsonb;
  IF has_old <> 1 THEN
    RAISE EXCEPTION 'rollback 458: the promoted row is not the pre-458 one (its avoid list lacks the superseded carousel line)';
  END IF;

  RAISE NOTICE 'rollback 458 OK: pre-ruling design_intent restored. THE OWNER RULING OF 2026-08-18 IS NO LONGER IN THE SPEC.';
END $$;

COMMIT;
