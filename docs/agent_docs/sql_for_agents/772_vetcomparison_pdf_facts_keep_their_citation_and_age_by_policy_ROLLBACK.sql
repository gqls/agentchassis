-- 772_vetcomparison_pdf_facts_keep_their_citation_and_age_by_policy_ROLLBACK.sql
--
-- Reverts the five CMA-DRAFT facts to 759's shape: removes the citation, reverifiable and
-- staleness_days keys and restores a no_citation_because key.
--
-- ⚠ RUNNING THIS RESTORES A FALSE STATEMENT AND LOSES REAL PROVENANCE. The reason 759 recorded
-- ("a citation here would read as citation_lost drift every day for ever") is not true of this
-- platform — the refresher classifies an unreadable source as `error`, never drift
-- (evidence_citations.go:143-148, and that file's own header). Reverting also discards the source
-- URL and the verbatim bracketed quote for all five facts, and removes the staleness clock that
-- makes them ask for re-verification on 2026-09-23 when the substantive Order is due.
-- Prefer correcting a fact forward.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current
     AND f->>'id' LIKE 'CMA-DRAFT-%' AND f ? 'citation_not_refetchable_because';
  IF n IS DISTINCT FROM 5 THEN
    RAISE EXCEPTION '772 ROLLBACK ABORT: expected 5 facts in 772 shape, found % - look before reverting', coalesce(n::text,'NULL');
  END IF;
END $$;

UPDATE site_specs ss
   SET data = jsonb_set(ss.data, '{facts}', (
         SELECT jsonb_agg(
                  CASE WHEN f->>'id' LIKE 'CMA-DRAFT-%'
                       THEN (f - 'reverifiable' - 'staleness_days' - 'citation_not_refetchable_because')
                            || jsonb_build_object(
                                 'no_citation_because', 'REVERTED by 772_ROLLBACK to migration 759 shape. NOTE: 759''s original wording of this key asserted that a citation would read as citation_lost drift every day, which is NOT true of this platform.',
                                 'source', (f->'source') - 'citation')
                       ELSE f END ORDER BY ord)
           FROM jsonb_array_elements(ss.data->'facts') WITH ORDINALITY AS t(f, ord)))
 WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $$
DECLARE nfacts int; ncit int;
BEGIN
  SELECT jsonb_array_length(data->'facts') INTO nfacts FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;
  SELECT count(*) INTO ncit FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current
     AND f->>'id' LIKE 'CMA-DRAFT-%' AND f->'source' ? 'citation';
  IF nfacts IS DISTINCT FROM 21 OR ncit IS DISTINCT FROM 0 THEN
    RAISE EXCEPTION '772 ROLLBACK VERIFY: facts=% citations_remaining=%', coalesce(nfacts::text,'NULL'), coalesce(ncit::text,'NULL');
  END IF;
  RAISE NOTICE '772 ROLLBACK OK: 5 facts reverted to 759 shape - provenance and the 23 Sept clock are gone';
END $$;

COMMIT;
