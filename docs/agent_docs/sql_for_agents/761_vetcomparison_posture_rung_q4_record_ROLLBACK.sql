-- 761_vetcomparison_posture_rung_q4_record_ROLLBACK.sql
--
-- Removes ONLY the `posture` key migration 761 added to vetcomparison's evidence register.
-- It does not touch facts or banned_claims.
--
-- ⚠ Removing it re-opens the second half of the `missing_evidence_register` acceptance test
-- ("the posture rung is recorded with who declared it and when"), so the item becomes
-- unsatisfied again. That is the correct consequence, not a side effect - but know it before
-- running this.
--
-- Unlike 759's rollback, this one does NOT expire at the first refresher pass: it keys on the
-- key's presence rather than on created_by, and unknown keys round-trip through the daily
-- writer (RUNBOOK §8d), so it stays valid after a refresh.

BEGIN;

DO $$
DECLARE has boolean;
BEGIN
  SELECT data ? 'posture' INTO has FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;
  IF has IS NULL THEN
    RAISE EXCEPTION '761 ROLLBACK ABORT: no current evidence_base row for vetcomparison';
  END IF;
  IF NOT has THEN
    RAISE EXCEPTION '761 ROLLBACK ABORT: the current register carries no posture key - nothing to remove';
  END IF;
END $$;

UPDATE site_specs SET data = data - 'posture'
 WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;

DO $$
DECLARE has boolean; nfacts int;
BEGIN
  SELECT data ? 'posture', jsonb_array_length(data->'facts') INTO has, nfacts FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;
  IF has THEN
    RAISE EXCEPTION '761 ROLLBACK VERIFY: posture key still present';
  END IF;
  IF nfacts <> 21 THEN
    RAISE EXCEPTION '761 ROLLBACK VERIFY: the register lost facts - expected 21, found %', nfacts;
  END IF;
  RAISE NOTICE '761 ROLLBACK OK: posture key removed, register intact at % facts', nfacts;
END $$;

COMMIT;
