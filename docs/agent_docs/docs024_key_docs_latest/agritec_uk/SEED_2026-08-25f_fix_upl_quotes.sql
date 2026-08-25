-- ============================================================================
-- agritec.uk — UPL5/UPL6 quotes: drop the space the RENDERED page has and the
-- EXTRACTOR does not
-- Written 2026-08-25. Applied out of band (psql -f), per-site setup.
--
-- Two of the 83 quotes fail, and I had already written down the reason before
-- making the mistake: in the same message where I flagged these two as needing
-- hand-work I said "a generated quote is a hypothesis about the extractor's
-- whitespace, not about the source" — and then hand-wrote them from the
-- RENDERED page anyway, which is a hypothesis about the extractor's whitespace.
--
-- THE MECHANISM. Stored: "... on moorland (minimum 70% GLU ) £18 per hectare",
-- with a space before the closing paren, which is how my own regex text-strip
-- rendered it. NormalizeForQuoteMatch (datahelpers/citations.go:116) unescapes
-- entities, normalises punctuation and thousands separators, lowercases and
-- collapses whitespace RUNS to one space. It does not DELETE a single space. So
-- "glu )" can never match text extracting as "glu)", on either side, for ever.
-- Re-fetching cannot help: the strings differ by a character the normaliser
-- preserves.
--
-- HOW THE CORRECTION WAS CHECKED, and it is the part worth copying. Rather than
-- re-implement the extractor a third time, I called the real one — a throwaway
-- Go program importing datahelpers.VisibleTextFromHTML and
-- datahelpers.QuoteFoundInText, run over a fresh fetch of the source. Then I ran
-- ALL 104 stored quotes through it, not just the two:
--
--     pass 89   FAIL 2 (UPL5, UPL6)   not-checked 13 (cite other sources)
--
-- matching the 288 lane's independent dry run exactly.
--
-- ⚠ MY FIRST HARNESS SAID 15 FAILING AND WAS WRONG. It mapped every non-vt.edu
-- URL to the SFI page, so facts citing ONS, Carbon Brief, GPN and Purdue were
-- compared against a document that does not contain them — guaranteed failures
-- carrying no information. The same defect I have been catching all week, in the
-- tool I built to catch it: **a checker that answers a question about the wrong
-- document.** Fixed by matching each citation's host to its cached source and
-- reporting the unmatched ones as NOT CHECKED rather than as failures.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; hits int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT count(*) INTO hits FROM _cur, LATERAL jsonb_array_elements(data->'facts') x
   WHERE x->>'id' IN ('ATT-sfi26-UPL5','ATT-sfi26-UPL6')
     AND x->'source'->'citation'->>'quote' LIKE '%GLU ) %';
  IF hits <> 2 THEN RAISE EXCEPTION 'expected 2 spaced-paren quotes to fix, found %', hits; END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id, 'evidence_base',
  jsonb_set(_cur.data, '{facts}',
    (SELECT jsonb_agg(
       CASE f->>'id'
         WHEN 'ATT-sfi26-UPL5' THEN jsonb_set(f,'{source,citation,quote}',
              to_jsonb('UPL5 Supplement: Keep cattle and ponies on moorland (minimum 70% GLU) £18 per hectare'::text))
         WHEN 'ATT-sfi26-UPL6' THEN jsonb_set(f,'{source,citation,quote}',
              to_jsonb('UPL6 Supplement: Keep cattle and ponies on moorland (100% GLU) £23 per hectare'::text))
         ELSE f END ORDER BY ord)
     FROM jsonb_array_elements(_cur.data->'facts') WITH ORDINALITY t(f,ord))),
  'manual',
  'Removed the stray space before the closing paren in the UPL5 and UPL6 quotes. The rendered page shows "(minimum 70% GLU )"; the extractor produces "(minimum 70% GLU)", and NormalizeForQuoteMatch collapses whitespace runs but never deletes a single space, so the stored form could never match. Corrected form verified by calling datahelpers.QuoteFoundInText directly against a fresh fetch.',
  true, true, 'agritec-workstream-2026-08-25'
FROM _cur;

COMMIT;
