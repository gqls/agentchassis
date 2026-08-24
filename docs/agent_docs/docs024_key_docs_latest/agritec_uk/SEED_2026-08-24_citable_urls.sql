-- ============================================================================
-- agritec.uk — give every attested fact a STRUCTURED, linkable source URL
-- Written 2026-08-24. Applied out of band (psql -f), per-site setup.
--
-- WHY. Owner instruction 2026-08-24: link out to the facts so readers can have
-- confidence in our figures. That requires a URL a renderer can EXTRACT, and
-- measured before writing this: of 105 facts, only 22 had one. The other 83 are
-- the ones this lane attested, and their URLs were buried inside the free-text
-- `attested_by` string — readable by a human, useless to a template.
--
-- So this is a defect in my own earlier migrations, not a new requirement
-- landing on sound work. The attestations were auditable but not renderable.
--
-- WHAT IT DOES. Adds `source.citation.{url,title}` to all 83 attested facts, so
-- `f->'source'->'citation'->>'url'` resolves uniformly across ALL 105 - the same
-- path the 22 citation-verified facts already use. One accessor, every fact.
-- `attested_by` is left exactly as it is: it records WHO checked and HOW, which
-- the URL does not, and the two are not substitutes.
--
-- ON THE STRUCT. `EvidenceSource` (claims.go:65) models sql/artifact/attested_by
-- and NOT citation - yet the 22 verified facts carry `source.citation` in their
-- JSON, and TestRegulatedOnlyBaseIsNotSafeToWriteBack pins that a struct round
-- trip destroys it. So citation already lives in the JSON outside the struct.
-- This is consistent with how the register actually works, and it is safe for
-- exactly the reason that test states: no ParseEvidenceBase caller may persist
-- the struct, so nothing round-trips it away.
--
-- HOW THE LINK MUST BE RENDERED, because getting this wrong is a known defect
-- class: as an HTML anchor in a markup-bearing field. NOT as markdown. There is
-- NO markdown renderer anywhere in the platform (nothing in go.mod), and
-- `check_literal_markdown` treats `[text](url)` in a text-typed field as a
-- defect to strip - 79 such work items exist fleet-wide. Its own letter-guards
-- exempt "values carrying HTML markup", which is the seam a citation link must
-- use: an <a href> inside a markup-bearing field renders; a markdown link in a
-- text field gets stripped and the citation silently disappears.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; f int; nourl int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT jsonb_array_length(data->'facts') INTO f FROM _cur;
  IF f IS DISTINCT FROM 105 THEN
    RAISE EXCEPTION 'expected 105 facts, found % - another session has written here, refusing', f;
  END IF;
  SELECT count(*) INTO nourl FROM _cur, LATERAL jsonb_array_elements(data->'facts') x
   WHERE x->'source'->'citation'->>'url' IS NULL;
  IF nourl <> 83 THEN
    RAISE EXCEPTION 'expected 83 facts lacking a structured url, found % - state has moved, refusing', nourl;
  END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id, 'evidence_base',
  jsonb_set(_cur.data, '{facts}',
    (SELECT jsonb_agg(
       CASE
         WHEN f->'source'->'citation'->>'url' IS NOT NULL THEN f          -- already linkable
         WHEN f->>'id' LIKE 'ATT-sfi26-%' THEN
           jsonb_set(f, '{source,citation}', jsonb_build_object(
             'url',   'https://www.gov.uk/government/publications/sustainable-farming-incentive-2026-sfi26/sfi26-scheme-rules-and-guidance',
             'title', 'SFI26 scheme rules and guidance',
             'publisher', 'Department for Environment, Food and Rural Affairs (GOV.UK)',
             'captured', '2026-08-22'))
         WHEN f->>'id' LIKE 'ATT-dli-%' THEN
           jsonb_set(f, '{source,citation}', jsonb_build_object(
             'url',   'https://pubs.ext.vt.edu/SPES/spes-720/spes-720.html',
             'title', 'Lighting for greenhouse and indoor production, SPES-720, Table 3',
             'publisher', 'Virginia Cooperative Extension',
             'captured', '2026-08-22'))
         ELSE f
       END ORDER BY ord)
     FROM jsonb_array_elements(_cur.data->'facts') WITH ORDINALITY t(f,ord))),
  'manual',
  'Owner instruction 2026-08-24: cite the figures so readers can check them. Added source.citation.{url,title,publisher,captured} to all 83 attested facts so f->source->citation->>url resolves uniformly across all 105 - previously only the 22 citation-verified facts had an extractable URL and the attested ones buried it in free text. attested_by left untouched: it records who checked and how, which a URL does not.',
  true, true, 'agritec-workstream-2026-08-24'
FROM _cur;

COMMIT;
