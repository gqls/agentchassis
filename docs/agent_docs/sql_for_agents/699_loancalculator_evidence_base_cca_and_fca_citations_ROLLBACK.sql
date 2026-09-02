-- 699_loancalculator_evidence_base_cca_and_fca_citations_ROLLBACK.sql
--
-- Removes loancalculator.co.uk's evidence register created by 699.
--
-- ⚠ WHAT THIS COSTS. Rolling back returns loancalculator.co.uk to having NO
-- registered facts, so its daily citation refresh goes back to passing over an
-- empty set — a clean run that means nothing (RFC_060 §1). It also discards the
-- machine-readable record that two of the site's live claims are wrong (the
-- ten-vs-twelve working days figure and the 10%-of-balance ERC-free attribution);
-- those findings then survive only in the lane NOTES and README.
--
-- Only roll back if the register itself is wrong. If a single fact is wrong, edit
-- that fact — do not delete the register.
--
-- ⚠ Check first whether the refresher has already updated the row, because this
-- deletes ITS work too, not just ours:
--
--   SELECT created_by, created_at, updated_at,
--          jsonb_array_length(data->'facts') AS facts
--     FROM site_specs
--    WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='evidence_base';
--
-- If created_by is no longer 'loancalculator_couk lane (migration 699)', or the
-- fact count is not 12, the register has moved on — the guard below will refuse,
-- and that refusal is correct. Re-state the reason for reverting before forcing it.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
     AND aspect = 'evidence_base'
     AND created_by = 'loancalculator_couk lane (migration 699)'
     AND jsonb_array_length(data->'facts') = 12;
  IF n <> 1 THEN
    RAISE EXCEPTION '699 ROLLBACK ABORT: expected exactly 1 unmodified register written by 699, found % — something else has written or amended it, read before deleting', n;
  END IF;
END $$;

DELETE FROM site_specs
 WHERE site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
   AND aspect = 'evidence_base'
   AND created_by = 'loancalculator_couk lane (migration 699)';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect = 'evidence_base';
  IF n <> 0 THEN
    RAISE EXCEPTION '699 ROLLBACK: % evidence_base row(s) survive', n;
  END IF;
  RAISE NOTICE '699 ROLLBACK OK: register removed — loancalculator is register-less again, and its daily check is vacuous again';
END $$;

COMMIT;
