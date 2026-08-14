-- 408_offer_analyser_agent_ROLLBACK.sql
--
-- Soft-deletes the offer-analyser agent definition. The row is config only.
--
-- What this does NOT undo, deliberately: any `site_specs` rows the agent has
-- already written under aspect `offer_ordering`. Those are data, not config;
-- nothing else reads them yet, and destroying a written analysis to undo a
-- config change is the wrong trade. To retire them too, demote rather than
-- delete, so the history survives:
--
--   UPDATE site_specs SET is_current = false, superseded_at = now()
--    WHERE aspect = 'offer_ordering' AND is_current = true;
--
-- If the improvement-loop wiring migration has been applied, roll THAT back
-- first — an improvement sweep that calls a soft-deleted agent fails the step.

BEGIN;

UPDATE agent_definitions
   SET is_active = false, deleted_at = now()
 WHERE type = 'offer-analyser' AND deleted_at IS NULL;

DO $verify$
DECLARE
  n integer;
BEGIN
  SELECT count(*) INTO n
  FROM agent_definitions
  WHERE type = 'offer-analyser' AND is_active AND deleted_at IS NULL;

  IF n <> 0 THEN
    RAISE EXCEPTION 'rollback verify: % active offer-analyser row(s) remain', n;
  END IF;
END $verify$;

COMMIT;
