-- 738_loancash_evidence_base_price_cap_complaints_and_breathing_space_ROLLBACK.sql
--
-- Removes the evidence register migration 738 created for loancash.co.uk.
--
-- ⚠ THIS GUARD EXPIRES, BY DESIGN, AT THE FIRST REFRESHER PASS (RUNBOOK_lendzy §8d).
-- The daily evidence-freshness sweep rewrites the register as a NEW row created_by
-- 'evidence-refresher' and supersedes 738's. After that has happened, the row this
-- file targets is no longer is_current and the DELETE below will refuse rather than
-- guess. That refusal is correct: once the refresher has touched it, the register is
-- no longer purely 738's to remove, and a human should look.
--
-- Related landmine: a refreshed spec's created_by names the LAST WRITER, not the
-- author of the values it carries.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74'
     AND aspect = 'evidence_base'
     AND created_by = 'loancash_couk_fca_validation lane (migration 738)'
     AND is_current;
  IF n <> 1 THEN
    RAISE EXCEPTION '738 ROLLBACK ABORT: expected exactly 1 current row created by migration 738, found % - the refresher has probably superseded it; inspect the history before removing anything', n;
  END IF;
END $$;

DELETE FROM site_specs
 WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74'
   AND aspect = 'evidence_base'
   AND created_by = 'loancash_couk_fca_validation lane (migration 738)'
   AND is_current;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = 'ee4a8199-4f5b-4e2e-88ce-01e600721b74' AND aspect = 'evidence_base';
  IF n <> 0 THEN
    RAISE EXCEPTION '738 ROLLBACK VERIFY: expected 0 evidence_base rows after delete, found %', n;
  END IF;
  RAISE NOTICE '738 ROLLBACK OK: loancash evidence register removed';
END $$;

COMMIT;
