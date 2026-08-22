-- ============================================================================
-- agritec.uk — drop four content-free "carbon intensity" facts
-- Written 2026-08-22. Applied out of band (psql -f), per-site setup.
--
-- Phase 2 run 4 (correlation 0a50c075-e25c-4f7e-a279-5dc9800a4bbf) asked for the
-- UK grid average carbon intensity in gCO2e/kWh. It COMPLETED and registered
-- four facts. NONE OF THEM CONTAINS A CARBON INTENSITY FIGURE — all four have
-- value: (none). What landed was: a publication date, two descriptions of a
-- methodology revision (from a third-party consultancy blog), and a GOV.UK
-- correction notice about well-to-tank emissions for hybrid cars and hotel
-- stays, which has nothing to do with agriculture.
--
-- THE CAUSE IS STRUCTURAL, NOT A BAD QUERY. Measured 2026-08-22 by fetching the
-- DESNZ publication page and searching it: ZERO kgCO2e/kWh figures appear in the
-- HTML. The conversion factors are published only as
--   ghg-conversion-factors-2025-condensed-set.xlsx
--   ghg-conversion-factors-2025-full-set.xlsx
--   ghg-conversion-factors-2025-flat-format.xlsx
--   2025-GHG-CF-methodology-paper.pdf
-- The HTML page is a landing page describing the downloads. evidence-researcher
-- scrapes HTML and requires a VERBATIM QUOTE from the page text, so it can only
-- ever find prose ABOUT the spreadsheet, never a number inside it. No amount of
-- re-phrasing the question reaches a figure that is not in any HTML.
--
-- THE GENERAL RULE, which shapes the rest of Phase 2: THE REGISTER CAN ONLY HOLD
-- WHAT A SOURCE STATES IN HTML PROSE. A figure published only in a spreadsheet,
-- a PDF table, or behind an API is unreachable by this pipeline. That is a
-- TOOLING gap, not a knowledge gap, and the distinction matters when reporting
-- it: "we could not verify this" and "we could not verify this WITH THIS TOOL"
-- are different statements. The supported remedy is the `attested_by` source
-- kind - a human opens the file, reads the cell, and registers the fact with the
-- file URL and an attestation - which the register's own schema_notes already
-- allow ("source: EXACTLY ONE of {sql|artifact|attested_by|citation}").
--
-- CARBON INTENSITY IS DEFERRED, not failed. No Phase 1 calculator needs it: the
-- vertical-farm energy tool takes DLI, photoperiod, area, efficacy and
-- electricity cost, and returns money, not emissions. The only places carbon
-- intensity appeared on the retired site were the fabricated ticker (gone) and
-- the dead data layer (never read). Spending further runs on a figure no Phase 1
-- artefact consumes is not a good use of them.
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
  IF f IS DISTINCT FROM 20 THEN
    RAISE EXCEPTION 'expected 20 facts, found % - another session has written here, refusing', f;
  END IF;
  SELECT count(*) INTO hit FROM _cur, LATERAL jsonb_array_elements(data->'facts') x
   WHERE x->>'id' IN ('CIT-aabc8109512d57f4','CIT-293c8c4a7f554eed','CIT-7dc344e6469dc616','CIT-a402dc079865382b');
  IF hit <> 4 THEN RAISE EXCEPTION 'expected the 4 target facts, found %', hit; END IF;
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
      WHERE f->>'id' NOT IN ('CIT-aabc8109512d57f4','CIT-293c8c4a7f554eed','CIT-7dc344e6469dc616','CIT-a402dc079865382b'))),
  'manual',
  'Phase 2 run 4 returned zero usable facts: the DESNZ GHG conversion factors exist only as .xlsx and .pdf, so the HTML-scraping pipeline can never reach a gCO2e/kWh figure. Removed all four (content-free; one about hotel stays and hybrid cars). Carbon intensity DEFERRED - no Phase 1 calculator consumes it.',
  true, true, 'agritec-workstream-2026-08-22'
FROM _cur;

COMMIT;

-- Verify: facts 20 -> 16, bans unchanged at 29, the 5 ONS price facts still present.
