-- 755_d4b_gate_composes_the_withheld_body_only_when_withholding_ROLLBACK.sql — restores the
-- post-754 gate query (md5 77336786e5c169402211ee886d998f60), whose body is composed
-- UNCONDITIONALLY. ⚠ After this, every ADMITTED council run again carries a sentence saying it
-- was WITHHELD (the 2026-09-04 misreading). Roll back only if the conditional body itself
-- misbehaves.

BEGIN;

DO $$
DECLARE q text;
BEGIN
  SELECT default_config#>>'{workflow,steps,gate_spend_governor,config,query}' INTO q
  FROM agent_definitions WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF q IS NULL THEN RAISE EXCEPTION '755 ROLLBACK REFUSED: no gate step.'; END IF;
  IF position('CASE WHEN COALESCE(governor_admits_agent' in q) = 0 THEN RAISE EXCEPTION '755 ROLLBACK REFUSED: body is not conditional — 755 not applied (or already rolled back).'; END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,gate_spend_governor,config,query}',
  to_jsonb($Q$SELECT COALESCE(governor_admits_agent('council-gate'), true) AS admitted, COALESCE(gs.shed_level, 0) AS shed_level, format('spend-governor: council-gate run for submission %s WITHHELD at shed level %s (%s%% of budget spent) - NOT queued; do not retry; re-trigger when governor_state.shed_level drops. RFC_065.', $1::text, COALESCE(gs.shed_level, 0), COALESCE(round(100*gs.mtd_usd/NULLIF(gc.monthly_budget_usd,0)), 0)) AS body FROM (SELECT 1) always_one_row LEFT JOIN governor_state gs ON gs.id=1 LEFT JOIN governor_config gc ON gc.id=1$Q$::text))
WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE q text;
BEGIN
  SELECT default_config#>>'{workflow,steps,gate_spend_governor,config,query}' INTO q
  FROM agent_definitions WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF md5(q) <> '77336786e5c169402211ee886d998f60' THEN RAISE EXCEPTION '755 ROLLBACK VERIFY: gate query md5 % is not the post-754 text', md5(q); END IF;
  EXECUTE 'PREPARE p755rb AS ' || q; EXECUTE 'DEALLOCATE p755rb';
  RAISE NOTICE '755 ROLLBACK OK: post-754 gate query restored (md5 77336786…). REMINDER: its body says WITHHELD on every admitted run.';
END $$;

COMMIT;
