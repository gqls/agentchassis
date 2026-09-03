-- 759_vetcomparison_evidence_base_cma_draft_order_and_govuk_pet_travel_ROLLBACK.sql
--
-- Removes the evidence register migration 759 created for vetcomparison.uk.
--
-- ⚠ THIS GUARD EXPIRES, BY DESIGN, AT THE FIRST REFRESHER PASS (RUNBOOK_lendzy §8d).
-- The daily evidence-freshness sweep rewrites the register as a NEW row created_by
-- 'evidence-refresher' and supersedes 759's. After that has happened the row this file
-- targets is no longer is_current and the DELETE below will REFUSE rather than guess.
-- That refusal is correct: once the refresher has touched it, the register carries work
-- that is not 759's to remove, and a human should look before forcing anything.
--
-- ⚠ AND ROLLING BACK IS NOT NEUTRAL. This register is the only place three live errors on
-- vetcomparison.uk are recorded (the November-2024 final-report date, the £21/£12.50 caps
-- served as settled, and "36 service categories" for 36 services in 5 categories). Deleting
-- it does not restore a safe prior state - it restores the state in which nothing on the
-- platform knows those errors exist, and in which the site's `missing_evidence_register`
-- work item becomes true again. It also disarms the six banned-claim patterns, including the
-- one guarding the phrase this site was remediated for in July 2026. If you are removing this
-- because a FACT is wrong, prefer correcting that fact forward.
--
-- Related landmine: a refreshed spec's created_by names the LAST WRITER, not the author of
-- the values it carries.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d'
     AND aspect = 'evidence_base'
     AND created_by = 'bugfix_414 register-programme lane (migration 759)'
     AND is_current;
  IF n <> 1 THEN
    RAISE EXCEPTION '759 ROLLBACK ABORT: expected exactly 1 current row created by migration 759, found % - the refresher has probably superseded it; inspect the history before removing anything', n;
  END IF;
END $$;

DELETE FROM site_specs
 WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d'
   AND aspect = 'evidence_base'
   AND created_by = 'bugfix_414 register-programme lane (migration 759)'
   AND is_current;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;
  IF n <> 0 THEN
    RAISE EXCEPTION '759 ROLLBACK VERIFY: expected 0 current evidence_base rows after delete, found %', n;
  END IF;
  RAISE NOTICE '759 ROLLBACK OK: vetcomparison evidence register removed - the three recorded live errors are no longer recorded anywhere on the platform';
END $$;

COMMIT;
