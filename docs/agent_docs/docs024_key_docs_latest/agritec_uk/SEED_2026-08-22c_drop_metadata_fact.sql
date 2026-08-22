-- ============================================================================
-- agritec.uk — drop one content-free fact from the register (2nd occurrence)
-- Written 2026-08-22. Applied out of band (psql -f), per-site setup.
--
-- Phase 2 run 3 (correlation 3503c85d-5df8-4641-85b9-3648ef3e939e) was a good
-- run: 6 facts, 5 of them from ONS, all NON-domestic (the market this site's
-- readers actually buy in), all real sentences rather than table cells, and all
-- carrying a writer_line that names the period and the source — e.g.
--   "was {value} pence per kWh in Q4 2024, sourced from DESNZ QEP table 3.4.1
--    (as cited by ONS, May 2025)"
-- which is what stops a 20-month-old figure being written as today's price.
--
-- One fact is dropped, on RUNBOOK check 3 (content-free): CIT-1595249487b06487,
-- "DESNZ QEP table 3.4.1 was last updated on 30 June 2026". True, checkable,
-- carries no value and licenses no claim. The register is the writer's
-- whitelist, and an entry that permits nothing only dilutes it.
--
-- ⚠ SECOND OCCURRENCE. The identical claim was registered and quarantined
-- earlier today as CIT-3782d8ebbf0a652d. The extractor re-finds the "Last
-- updated:" line every time it visits that GOV.UK page, so expect it again on
-- any future run that touches DESNZ QEP and drop it again. It is NOT banned:
-- the phrase is page furniture rather than a claim anyone would publish, and a
-- ban on "last updated" would fire on legitimate prose about data currency —
-- which this site, whose whole pitch is dated sourcing, will want to write.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; f int; hit int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT jsonb_array_length(data->'facts') INTO f FROM _cur;
  IF f IS DISTINCT FROM 17 THEN
    RAISE EXCEPTION 'expected 17 facts, found % - another session has written here, refusing', f;
  END IF;
  SELECT count(*) INTO hit FROM _cur, LATERAL jsonb_array_elements(data->'facts') x
   WHERE x->>'id' = 'CIT-1595249487b06487';
  IF hit <> 1 THEN
    RAISE EXCEPTION 'target fact CIT-1595249487b06487 not present exactly once (found %)', hit;
  END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id, 'evidence_base',
  jsonb_set(_cur.data, '{facts}',
    (SELECT jsonb_agg(f ORDER BY ord)
       FROM jsonb_array_elements(_cur.data->'facts') WITH ORDINALITY t(f,ord)
      WHERE f->>'id' <> 'CIT-1595249487b06487')),
  'manual',
  'Phase 2 run 3: kept 5 ONS non-domestic price facts (all writer_line-scoped with period and source); dropped CIT-1595249487b06487 as content-free (RUNBOOK check 3). Second occurrence of this metadata claim today.',
  true, true, 'agritec-workstream-2026-08-22'
FROM _cur;

COMMIT;

-- Verify: facts 17 -> 16, bans unchanged at 29, and the five ONS price facts present.
