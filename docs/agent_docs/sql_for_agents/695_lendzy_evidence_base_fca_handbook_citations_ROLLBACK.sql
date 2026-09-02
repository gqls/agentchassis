-- 695_lendzy_evidence_base_fca_handbook_citations_ROLLBACK.sql
--
-- Removes lendzy.co.uk's evidence register created by 695.
--
-- ⚠ WHAT THIS COSTS. Rolling back returns lendzy to having NO registered facts,
-- which means its daily citation refresh goes back to passing over an empty set —
-- a clean run that means nothing (RFC_060 §1). It also discards the record that
-- two of the site's live rule citations are wrong; that finding then survives only
-- in the lane's NOTES and README, not anywhere a machine reads.
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
--    WHERE site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND aspect='evidence_base';
--
-- If created_by is no longer 'lendzy_co_uk lane (migration 695)', or the fact
-- count is not 8, the register has moved on — the guard below will refuse, and
-- that refusal is correct. Re-state the reason for reverting before forcing it.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
     AND aspect = 'evidence_base'
     AND created_by = 'lendzy_co_uk lane (migration 695)'
     AND jsonb_array_length(data->'facts') = 8;
  IF n <> 1 THEN
    RAISE EXCEPTION '695 ROLLBACK ABORT: expected exactly 1 unmodified register written by 695, found % — something else has written or amended it, read before deleting', n;
  END IF;
END $$;

DELETE FROM site_specs
 WHERE site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76'
   AND aspect = 'evidence_base'
   AND created_by = 'lendzy_co_uk lane (migration 695)';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id = '8ff093d5-1f19-453b-9439-a10379bbcd76' AND aspect = 'evidence_base';
  IF n <> 0 THEN
    RAISE EXCEPTION '695 ROLLBACK: % evidence_base row(s) survive', n;
  END IF;
  RAISE NOTICE '695 ROLLBACK OK: register removed — lendzy is register-less again, and its daily check is vacuous again';
END $$;

COMMIT;
