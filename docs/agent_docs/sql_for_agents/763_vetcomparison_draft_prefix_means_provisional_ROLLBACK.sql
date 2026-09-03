-- 763_vetcomparison_draft_prefix_means_provisional_ROLLBACK.sql
--
-- Renames the fact id back. ⚠ Doing so RESTORES the defect the council's medium objection named:
-- six ids matching CMA-DRAFT-% against five carrying draft_status, so a reader grepping the
-- prefix gets a set that disagrees with the tag. Only run this if the rename itself caused a
-- problem, and say what it was.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current
     AND f->>'id' = 'CMA-ORDER-CONSULTATION-PUBLISHED-2026-07-21';
  IF n <> 1 THEN
    RAISE EXCEPTION '763 ROLLBACK ABORT: expected 1 fact with the renamed id, found %', n;
  END IF;
END $$;

UPDATE site_specs ss
   SET data = jsonb_set(ss.data, '{facts}', (
         SELECT jsonb_agg(
                  CASE WHEN f->>'id' = 'CMA-ORDER-CONSULTATION-PUBLISHED-2026-07-21'
                       THEN jsonb_set(f, '{id}', '"CMA-DRAFT-ORDER-CONSULTATION-2026-07-21"'::jsonb)
                       ELSE f END
                  ORDER BY ord)
           FROM jsonb_array_elements(ss.data->'facts') WITH ORDINALITY AS t(f, ord)))
 WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $$
DECLARE nfacts int;
BEGIN
  SELECT jsonb_array_length(data->'facts') INTO nfacts FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;
  IF nfacts IS DISTINCT FROM 21 THEN
    RAISE EXCEPTION '763 ROLLBACK VERIFY: expected 21 facts, found %', coalesce(nfacts::text,'NULL');
  END IF;
  RAISE NOTICE '763 ROLLBACK OK: id renamed back - the prefix/tag mismatch is restored';
END $$;

COMMIT;
