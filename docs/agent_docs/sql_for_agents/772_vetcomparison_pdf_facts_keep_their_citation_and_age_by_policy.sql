-- 772_vetcomparison_pdf_facts_keep_their_citation_and_age_by_policy.sql
--
-- Corrects migration 759's handling of the five CMA facts whose primary source is a PDF, and the
-- FALSE REASON it recorded for that handling.
--
-- ─── WHAT 759 GOT WRONG ──────────────────────────────────────────────────────────────────────
--
-- 759 gave those facts `source.attested_by` and NO citation, and recorded the reason in a
-- `no_citation_because` key: *"a citation here would fail the quote match EVERY DAY and read as
-- citation_lost drift for ever"*. **That reason is false.** The nightly refresher never treats an
-- unreadable source as drift: `refreshCitationFact` -> `verifyCitationLiveForRule` ->
-- `fetchCitationDocument`, which refuses a non-html/xml/text content type
-- (`evidence_citations.go:143-148`) and returns an error classified `fetch_error` -> outcome
-- **`error`**. That file's header has always said so: *"fetch failed (network, 403, 5xx, unsupported
-- content type) -> UNKNOWN, not drift ... Reported as an error, never as loss"*, and *"PDFs and other
-- non-text content are refused rather than half-read: a fact whose source the platform cannot
-- re-fetch as text should carry `reverifiable: false` and a human attestation of having read it."*
--
-- **That last sentence is the platform's own answer and 759 did not take it.** The claim came from
-- `cmd/fcaquotecheck`, which did its own bare fetch rather than production's - fixed the same day,
-- and every site of the false claim is corrected rather than edited away (WRONG_CALLS 2026-09-04).
--
-- ─── WHAT THIS CHANGES, AND WHY IT IS STRICTLY BETTER THAN A DISCARD ─────────────────────────
--
-- Each of the five facts KEEPS what it had and GAINS what 759 threw away:
--
--   + `source.citation` {url, quote, title, publisher, accessed, published} - the URL and the
--     VERBATIM bracketed text are back in the record, so a human can check the figure at its source
--     instead of taking the attestation's word for it. 759 discarded both for no benefit.
--   + `reverifiable: false` - the documented path. `refreshCitationFact` checks it BEFORE fetching
--     and returns early, so the PDF is never requested and no error is ever logged.
--   + `staleness_days: 64`, anchored on `published: 2026-07-21` (the consultation publication date).
--     `citationDateStale` ages from `published`, so these facts go stale on **2026-09-23** - the CMA's
--     statutory deadline to make the substantive Order - and the refresher then reports
--     *"past its staleness_days policy and is marked reverifiable:false - re-attest it by hand"*.
--     **The 23 September re-verification stops depending on someone reading a handoff.**
--   ~ `no_citation_because` is REPLACED by `citation_not_refetchable_because`, carrying the true
--     reason. The key is renamed, not just reworded, because the old NAME is now wrong: there IS a
--     citation.
--   = `source.attested_by` is KEPT. Who read the document and when is still the load-bearing
--     provenance for a fact the machine cannot re-check.
--
-- The three non-PDF attested facts (VC-CLAIMING-IS-FREE, VC-NO-INVENTED-FIGURES,
-- VC-DIRECTORY-LISTINGS-60) are NOT touched: they are attestations about the site itself with no
-- external URL, which is what `attested_by` is for.
--
-- Rollback: 772_..._ROLLBACK.sql

BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM sites WHERE id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND domain = 'vetcomparison.uk') THEN
    RAISE EXCEPTION '772 ABORT: site_id does not resolve to vetcomparison.uk';
  END IF;
END $$;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current
     AND f->>'id' LIKE 'CMA-DRAFT-%' AND f ? 'no_citation_because';
  IF n IS DISTINCT FROM 5 THEN
    RAISE EXCEPTION '772 ABORT: expected 5 CMA-DRAFT facts carrying the old no_citation_because key, found % - the register has moved on', coalesce(n::text,'NULL');
  END IF;
END $$;

UPDATE site_specs ss
   SET data = jsonb_set(ss.data, '{facts}', (
         SELECT jsonb_agg(
                  CASE
                    WHEN f->>'id' = 'CMA-DRAFT-PRIMARY-PRESCRIPTION-FEE-CAP-21' THEN
                      (f - 'no_citation_because')
                        || jsonb_build_object(
                             'reverifiable', false,
                             'staleness_days', 64,
                             'citation_not_refetchable_because', 'This source is a PDF. The nightly refresher''s fetch refuses a non-text content type (evidence_citations.go:143-148) and classifies it fetch_error -> outcome `error`, never drift, so the citation below is recorded for PROVENANCE and is deliberately never re-fetched: `reverifiable: false` short-circuits refreshCitationFact before the fetch. The quote was verified by a human reading the primary document with pdftotext -layout on 2026-09-03. `staleness_days: 64` is anchored on the draft''s publication date (2026-07-21), so this fact ages out on 2026-09-23 - the CMA''s statutory deadline to MAKE the substantive Order - and the refresher then reports it as needing a hand re-attestation, which is exactly when the bracketed figures become real ones.',
                             'source', (f->'source') || jsonb_build_object('citation', '{"url": "https://connect.cma.gov.uk/consultations/draft-substantive-order-and-undertakings/supporting_documents/draft-substantive-orderpdf", "quote": "‘Initial Primary Prescription Fee Cap’ means [£21 inclusive of VAT. This will be adjusted for inflation between the date of the Final Decision Report and the latest monthly CPI figure available before the Order is made].", "title": "Draft Veterinary Services Market Investigation Order 2026 (consultation draft)", "publisher": "Competition and Markets Authority", "accessed": "2026-09-03", "published": "2026-07-21"}'::jsonb))
                    WHEN f->>'id' = 'CMA-DRAFT-ADDITIONAL-PRESCRIPTION-FEE-CAP-12-50' THEN
                      (f - 'no_citation_because')
                        || jsonb_build_object(
                             'reverifiable', false,
                             'staleness_days', 64,
                             'citation_not_refetchable_because', 'This source is a PDF. The nightly refresher''s fetch refuses a non-text content type (evidence_citations.go:143-148) and classifies it fetch_error -> outcome `error`, never drift, so the citation below is recorded for PROVENANCE and is deliberately never re-fetched: `reverifiable: false` short-circuits refreshCitationFact before the fetch. The quote was verified by a human reading the primary document with pdftotext -layout on 2026-09-03. `staleness_days: 64` is anchored on the draft''s publication date (2026-07-21), so this fact ages out on 2026-09-23 - the CMA''s statutory deadline to MAKE the substantive Order - and the refresher then reports it as needing a hand re-attestation, which is exactly when the bracketed figures become real ones.',
                             'source', (f->'source') || jsonb_build_object('citation', '{"url": "https://connect.cma.gov.uk/consultations/draft-substantive-order-and-undertakings/supporting_documents/draft-substantive-orderpdf", "quote": "‘Initial Additional Prescription Fee Cap’ means [£12.50 inclusive of VAT. This will be adjusted for inflation between the date of the Final Decision Report and the latest monthly CPI figure available before the Order is made].", "title": "Draft Veterinary Services Market Investigation Order 2026 (consultation draft)", "publisher": "Competition and Markets Authority", "accessed": "2026-09-03", "published": "2026-07-21"}'::jsonb))
                    WHEN f->>'id' = 'CMA-DRAFT-LARGE-BUSINESS-THRESHOLD-15' THEN
                      (f - 'no_citation_because')
                        || jsonb_build_object(
                             'reverifiable', false,
                             'staleness_days', 64,
                             'citation_not_refetchable_because', 'This source is a PDF. The nightly refresher''s fetch refuses a non-text content type (evidence_citations.go:143-148) and classifies it fetch_error -> outcome `error`, never drift, so the citation below is recorded for PROVENANCE and is deliberately never re-fetched: `reverifiable: false` short-circuits refreshCitationFact before the fetch. The quote was verified by a human reading the primary document with pdftotext -layout on 2026-09-03. `staleness_days: 64` is anchored on the draft''s publication date (2026-07-21), so this fact ages out on 2026-09-23 - the CMA''s statutory deadline to MAKE the substantive Order - and the refresher then reports it as needing a hand re-attestation, which is exactly when the bracketed figures become real ones.',
                             'source', (f->'source') || jsonb_build_object('citation', '{"url": "https://connect.cma.gov.uk/consultations/draft-substantive-order-and-undertakings/supporting_documents/draft-substantive-orderpdf", "quote": "‘Large Veterinary Business’ means a Veterinary Business with 15 or more FOPs and/or OOH Centres.", "title": "Draft Veterinary Services Market Investigation Order 2026 (consultation draft)", "publisher": "Competition and Markets Authority", "accessed": "2026-09-03", "published": "2026-07-21"}'::jsonb))
                    WHEN f->>'id' = 'CMA-DRAFT-PRICE-LIST-SERVICES-36' THEN
                      (f - 'no_citation_because')
                        || jsonb_build_object(
                             'reverifiable', false,
                             'staleness_days', 64,
                             'citation_not_refetchable_because', 'This source is a PDF. The nightly refresher''s fetch refuses a non-text content type (evidence_citations.go:143-148) and classifies it fetch_error -> outcome `error`, never drift, so the citation below is recorded for PROVENANCE and is deliberately never re-fetched: `reverifiable: false` short-circuits refreshCitationFact before the fetch. The quote was verified by a human reading the primary document with pdftotext -layout on 2026-09-03. `staleness_days: 64` is anchored on the draft''s publication date (2026-07-21), so this fact ages out on 2026-09-23 - the CMA''s statutory deadline to MAKE the substantive Order - and the refresher then reports it as needing a hand re-attestation, which is exactly when the bracketed figures become real ones.',
                             'source', (f->'source') || jsonb_build_object('citation', '{"url": "https://connect.cma.gov.uk/consultations/draft-substantive-order-and-undertakings/supporting_documents/draft-schedule-1-price-list-schedulepdf", "quote": "Service, product, treatment or procedure (36 total)", "title": "Draft Schedule 1: Price List Schedule (consultation draft)", "publisher": "Competition and Markets Authority", "accessed": "2026-09-03", "published": "2026-07-21"}'::jsonb))
                    WHEN f->>'id' = 'CMA-DRAFT-PRICE-LIST-CATEGORIES-5' THEN
                      (f - 'no_citation_because')
                        || jsonb_build_object(
                             'reverifiable', false,
                             'staleness_days', 64,
                             'citation_not_refetchable_because', 'This source is a PDF. The nightly refresher''s fetch refuses a non-text content type (evidence_citations.go:143-148) and classifies it fetch_error -> outcome `error`, never drift, so the citation below is recorded for PROVENANCE and is deliberately never re-fetched: `reverifiable: false` short-circuits refreshCitationFact before the fetch. The quote was verified by a human reading the primary document with pdftotext -layout on 2026-09-03. `staleness_days: 64` is anchored on the draft''s publication date (2026-07-21), so this fact ages out on 2026-09-23 - the CMA''s statutory deadline to MAKE the substantive Order - and the refresher then reports it as needing a hand re-attestation, which is exactly when the bracketed figures become real ones.',
                             'source', (f->'source') || jsonb_build_object('citation', '{"url": "https://connect.cma.gov.uk/consultations/draft-substantive-order-and-undertakings/supporting_documents/draft-schedule-1-price-list-schedulepdf", "quote": "Service, product, treatment or procedure (36 total)", "title": "Draft Schedule 1: Price List Schedule (consultation draft)", "publisher": "Competition and Markets Authority", "accessed": "2026-09-03", "published": "2026-07-21"}'::jsonb))
                    ELSE f
                  END ORDER BY ord)
           FROM jsonb_array_elements(ss.data->'facts') WITH ORDINALITY AS t(f, ord)))
 WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $$
DECLARE nfacts int; nban int; ncit int; nrev int; nstale int; nold int; posture text; nq int;
BEGIN
  SELECT jsonb_array_length(data->'facts'), jsonb_array_length(data->'banned_claims'), data->'posture'->>'rung'
    INTO nfacts, nban, posture
    FROM site_specs WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;

  IF nfacts IS DISTINCT FROM 21 OR nban IS DISTINCT FROM 6 OR posture IS DISTINCT FROM 'relied_upon' THEN
    RAISE EXCEPTION '772 VERIFY: register damaged - facts=% banned=% posture=%',
      coalesce(nfacts::text,'NULL'), coalesce(nban::text,'NULL'), coalesce(posture,'NULL');
  END IF;

  SELECT count(*) FILTER (WHERE f->'source' ? 'citation'),
         count(*) FILTER (WHERE (f->>'reverifiable') = 'false'),
         count(*) FILTER (WHERE (f->>'staleness_days') = '64'),
         count(*) FILTER (WHERE f ? 'no_citation_because'),
         count(*) FILTER (WHERE length(f->'source'->'citation'->>'quote') > 20)
    INTO ncit, nrev, nstale, nold, nq
    FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current
     AND f->>'id' LIKE 'CMA-DRAFT-%';

  IF ncit IS DISTINCT FROM 5 OR nq IS DISTINCT FROM 5 THEN
    RAISE EXCEPTION '772 VERIFY: expected 5 CMA-DRAFT facts with a citation carrying a substantive quote, found % with citation / % with quote', coalesce(ncit::text,'NULL'), coalesce(nq::text,'NULL');
  END IF;
  -- reverifiable:false is what stops the refresher ever FETCHING the PDF. Without it the citation
  -- would be requested nightly and logged as an error for ever - noise, not drift, but still noise.
  IF nrev IS DISTINCT FROM 5 THEN
    RAISE EXCEPTION '772 VERIFY: expected 5 facts marked reverifiable:false, found % - a citation without it WOULD be fetched', coalesce(nrev::text,'NULL');
  END IF;
  IF nstale IS DISTINCT FROM 5 THEN
    RAISE EXCEPTION '772 VERIFY: expected 5 facts with staleness_days=64 (ages out on the 2026-09-23 statutory deadline), found %', coalesce(nstale::text,'NULL');
  END IF;
  IF nold IS DISTINCT FROM 0 THEN
    RAISE EXCEPTION '772 VERIFY: % fact(s) still carry the old no_citation_because key with its false reason', coalesce(nold::text,'NULL');
  END IF;

  RAISE NOTICE '772 OK: 5 PDF-sourced facts now keep their citation, are marked reverifiable:false so the refresher never fetches them, and age out on 2026-09-23';
END $$;

COMMIT;
