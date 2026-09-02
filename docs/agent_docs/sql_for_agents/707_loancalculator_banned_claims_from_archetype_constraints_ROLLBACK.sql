-- 707_loancalculator_banned_claims_from_archetype_constraints_ROLLBACK.sql
--
-- Empties loancalculator.co.uk's banned_claims back to [] (facts untouched).
--
-- ⚠ WHAT THIS COSTS. The eight patterns are the ONLY enforcement of the site's
-- archetype constraints ("never appear to give regulated advice", "never
-- reposition as a lender/broker", "never recommend lenders") — rolling back
-- returns those to unenforced prose. The five adapted predatory-language
-- patterns (guaranteed acceptance etc.) also stop refusing saves.
--
-- Only roll back if the PATTERN SET is wrong as a whole. A single bad pattern
-- (e.g. one that starts refusing legitimate copy) is removed by editing the
-- array — via the same supersede-and-merge shape — not by emptying it.
--
-- Supersede-and-merge, same reason as 707 itself: the daily refresher's
-- write-back is a CAS keyed on the row id it read, so an in-place edit between
-- its read and write is silently lost; a supersede makes it skip. The guard
-- keys on the SHAPE (12 facts + exactly the 8 patterns 707 wrote), not on
-- created_by, so it tolerates refresher passes in between — and refuses if the
-- pattern set has been edited since, which is the correct refusal.

\set ON_ERROR_STOP on
BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='evidence_base' AND is_current
     AND jsonb_array_length(data->'facts') = 12
     AND jsonb_array_length(COALESCE(data->'banned_claims','[]'::jsonb)) = 8;
  IF n <> 1 THEN
    RAISE EXCEPTION '707 ROLLBACK ABORT: expected exactly 1 current register with 12 facts and the 8 patterns 707 wrote, found % - the register has moved on, read before deleting', n;
  END IF;
END $$;

WITH cur AS (
  UPDATE site_specs SET is_current=false, superseded_at=now()
   WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='evidence_base' AND is_current
   RETURNING data, pinned
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT
  '0162cde4-633e-45e9-8ca6-87a6b2fe1d26',
  'evidence_base',
  jsonb_set(cur.data, '{banned_claims}', '[]'::jsonb),
  'manual', NULL, 'loancalculator_couk lane (707 ROLLBACK)', true, cur.pinned,
  '707 rolled back: banned_claims emptied, facts carried forward unchanged. The archetype constraints are unenforced prose again.'
FROM cur;

DO $$
DECLARE nbc int; nfacts int; ncur int;
BEGIN
  SELECT count(*) INTO ncur FROM site_specs
   WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='evidence_base' AND is_current;
  IF ncur <> 1 THEN RAISE EXCEPTION '707 ROLLBACK VERIFY: expected exactly 1 current row, found %', ncur; END IF;
  SELECT jsonb_array_length(data->'banned_claims'), jsonb_array_length(data->'facts')
    INTO nbc, nfacts FROM site_specs
   WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='evidence_base' AND is_current;
  IF nbc <> 0 OR nfacts <> 12 THEN
    RAISE EXCEPTION '707 ROLLBACK VERIFY: expected 0 banned_claims and 12 untouched facts, found % / %', nbc, nfacts;
  END IF;
  RAISE NOTICE '707 ROLLBACK OK: banned_claims emptied - the archetype constraints are unenforced prose again';
END $$;
COMMIT;
