-- ============================================================================
-- agritec.uk — register the Virginia Tech DLI table by ATTESTATION
-- Written 2026-08-22. Applied out of band (psql -f), per-site setup.
--
-- WHY THIS EXISTS, AND WHY IT CORRECTS AN EARLIER CONCLUSION OF THIS LANE
--
-- Phase 2 run 6 raised a `citation_unverified` work item: six candidate DLI
-- claims REJECTED as `citation_lost`, whose own advice reads "cited a source
-- that does not contain its quote (citation_lost / possible hallucination —
-- discard or re-research)". Two of the rejected claims were:
--     "The recommended DLI range for lettuce is 12-17 mol/m2/d"
--     "The recommended DLI range for tomato is 20-30 mol/m2/d"
-- which are EXACTLY the figures in the retired site's own crop-dli-table.json.
--
-- The obvious reading — the model reproduced the old site's unsourced numbers
-- and invented a citation for them — IS WRONG, and I nearly recorded it.
--
-- MEASURED 2026-08-22 by fetching https://pubs.ext.vt.edu/SPES/spes-720/spes-720.html
-- (HTTP 200) and reading it: its Table 3 says, verbatim,
--     "Micro-greens 9−12  Lettuce 12−17  Spinach 14−20  Parsley 10−15
--      Cilantro 15−20  Basil 15−25  Tomato 20−30  Cucumber 20−30  Zucchini 20−30"
-- attributed in the page to Dou et al. 2018, Faust et al. 2005, and Pramuk.
-- So the claims are TRUE and the source really does contain them.
--
-- THE REJECTION IS A FORMAT ARTEFACT, TWICE OVER:
--   1. The figure lives in a TABLE. The extractor had to turn "Lettuce 12−17"
--      into a sentence, and a sentence it composed is not a verbatim quote, so
--      re-matching necessarily fails.
--   2. The separator is U+2212 MINUS SIGN, not a hyphen or en-dash. Any
--      normalisation on the way through breaks a byte-exact match even if the
--      words survive.
--
-- CONSEQUENCE, and it is the dangerous part: `citation_lost` says "possible
-- hallucination — discard". A session following that advice discards a true
-- figure from a university extension publication AND concludes the retired
-- site's data was invented. Both wrong, from a check behaving as designed.
--
-- SO: the retired crop-dli-table.json is UNCITED, not fabricated — at least for
-- lettuce and tomato, which match this table exactly. That is a correction to
-- this lane's earlier framing, recorded in NOTES and the ledger too.
--
-- REMEDY USED: `source.attested_by`, the documented path for a figure this
-- pipeline cannot reach (48 attested facts exist fleet-wide). A human opened the
-- page and read the cell. Ranges carry NO `value` — the range goes in the
-- writer_line, which is the correct handling of a range and what the extractor
-- itself did for the Purdue ranges.
--
-- SCOPE NOTE: the ornamentals in Table 3 (Impatiens, Begonia, Geranium, Petunia)
-- are deliberately NOT registered — wrong audience for this site. Seedlings and
-- cuttings ARE registered because they are the transplant stage, which is what
-- reconciles this table with the Purdue transplant figures already in the
-- register. And "general DLI recommendations" is the page's own framing: these
-- are guidance ranges, not cultivar-specific optima.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; f int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT jsonb_array_length(data->'facts') INTO f FROM _cur;
  IF f IS DISTINCT FROM 22 THEN
    RAISE EXCEPTION 'expected 22 facts, found % - another session has written here, refusing', f;
  END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id, 'evidence_base',
  jsonb_set(_cur.data, '{facts}', (_cur.data->'facts') || $facts$[
    {"id":"ATT-dli-vt-table3","kind":"attestation",
     "claim":"Virginia Cooperative Extension publication SPES-720, Table 3, gives general recommended daily light integral ranges for common horticultural crops, attributed in the page to Dou et al. (2018), Faust et al. (2005) and Pramuk.",
     "source":{"attested_by":"agritec_uk lane, 2026-08-22: fetched https://pubs.ext.vt.edu/SPES/spes-720/spes-720.html (HTTP 200) and read Table 3 directly. Registered by attestation because the figures are TABLE CELLS using U+2212 MINUS SIGN, which no verbatim-quote check can match from a composed sentence - see SEED_2026-08-22f header."},
     "verified_at":"2026-08-22",
     "writer_line":"Virginia Cooperative Extension, SPES-720, Table 3"},

    {"id":"ATT-dli-seedlings","kind":"metric","claim":"The recommended DLI range for seedlings is 5 to 10 mol per square metre per day","unit":"mol/m2/d",
     "source":{"attested_by":"agritec_uk lane, 2026-08-22, from VT SPES-720 Table 3 (see ATT-dli-vt-table3)"},"verified_at":"2026-08-22",
     "writer_line":"seedlings 5-10 mol/m2/d (Virginia Cooperative Extension, SPES-720, Table 3)"},
    {"id":"ATT-dli-microgreens","kind":"metric","claim":"The recommended DLI range for micro-greens is 9 to 12 mol per square metre per day","unit":"mol/m2/d",
     "source":{"attested_by":"agritec_uk lane, 2026-08-22, from VT SPES-720 Table 3 (see ATT-dli-vt-table3)"},"verified_at":"2026-08-22",
     "writer_line":"micro-greens 9-12 mol/m2/d (Virginia Cooperative Extension, SPES-720, Table 3)"},
    {"id":"ATT-dli-lettuce","kind":"metric","claim":"The recommended DLI range for lettuce is 12 to 17 mol per square metre per day","unit":"mol/m2/d",
     "source":{"attested_by":"agritec_uk lane, 2026-08-22, from VT SPES-720 Table 3 (see ATT-dli-vt-table3)"},"verified_at":"2026-08-22",
     "writer_line":"lettuce 12-17 mol/m2/d (Virginia Cooperative Extension, SPES-720, Table 3)"},
    {"id":"ATT-dli-spinach","kind":"metric","claim":"The recommended DLI range for spinach is 14 to 20 mol per square metre per day","unit":"mol/m2/d",
     "source":{"attested_by":"agritec_uk lane, 2026-08-22, from VT SPES-720 Table 3 (see ATT-dli-vt-table3)"},"verified_at":"2026-08-22",
     "writer_line":"spinach 14-20 mol/m2/d (Virginia Cooperative Extension, SPES-720, Table 3)"},
    {"id":"ATT-dli-parsley","kind":"metric","claim":"The recommended DLI range for parsley is 10 to 15 mol per square metre per day","unit":"mol/m2/d",
     "source":{"attested_by":"agritec_uk lane, 2026-08-22, from VT SPES-720 Table 3 (see ATT-dli-vt-table3)"},"verified_at":"2026-08-22",
     "writer_line":"parsley 10-15 mol/m2/d (Virginia Cooperative Extension, SPES-720, Table 3)"},
    {"id":"ATT-dli-cilantro","kind":"metric","claim":"The recommended DLI range for cilantro (coriander) is 15 to 20 mol per square metre per day","unit":"mol/m2/d",
     "source":{"attested_by":"agritec_uk lane, 2026-08-22, from VT SPES-720 Table 3 (see ATT-dli-vt-table3)"},"verified_at":"2026-08-22",
     "writer_line":"cilantro (coriander) 15-20 mol/m2/d (Virginia Cooperative Extension, SPES-720, Table 3)"},
    {"id":"ATT-dli-basil","kind":"metric","claim":"The recommended DLI range for basil is 15 to 25 mol per square metre per day","unit":"mol/m2/d",
     "source":{"attested_by":"agritec_uk lane, 2026-08-22, from VT SPES-720 Table 3 (see ATT-dli-vt-table3)"},"verified_at":"2026-08-22",
     "writer_line":"basil 15-25 mol/m2/d (Virginia Cooperative Extension, SPES-720, Table 3)"},
    {"id":"ATT-dli-tomato","kind":"metric","claim":"The recommended DLI range for tomato is 20 to 30 mol per square metre per day","unit":"mol/m2/d",
     "source":{"attested_by":"agritec_uk lane, 2026-08-22, from VT SPES-720 Table 3 (see ATT-dli-vt-table3)"},"verified_at":"2026-08-22",
     "writer_line":"tomato 20-30 mol/m2/d (Virginia Cooperative Extension, SPES-720, Table 3)"},
    {"id":"ATT-dli-cucumber","kind":"metric","claim":"The recommended DLI range for cucumber is 20 to 30 mol per square metre per day","unit":"mol/m2/d",
     "source":{"attested_by":"agritec_uk lane, 2026-08-22, from VT SPES-720 Table 3 (see ATT-dli-vt-table3)"},"verified_at":"2026-08-22",
     "writer_line":"cucumber 20-30 mol/m2/d (Virginia Cooperative Extension, SPES-720, Table 3)"},
    {"id":"ATT-dli-zucchini","kind":"metric","claim":"The recommended DLI range for zucchini (courgette) is 20 to 30 mol per square metre per day","unit":"mol/m2/d",
     "source":{"attested_by":"agritec_uk lane, 2026-08-22, from VT SPES-720 Table 3 (see ATT-dli-vt-table3)"},"verified_at":"2026-08-22",
     "writer_line":"zucchini (courgette) 20-30 mol/m2/d (Virginia Cooperative Extension, SPES-720, Table 3)"}
  ]$facts$::jsonb),
  'manual',
  'Registered the VT SPES-720 Table 3 DLI ranges by attestation after run 6 rejected them as citation_lost. The rejection was a FORMAT artefact (table cells, U+2212 separator), not a fabrication: the page really does contain these figures, read first-hand 2026-08-22. Ornamentals deliberately omitted (wrong audience). Ranges carry no value; the range lives in writer_line.',
  true, true, 'agritec-workstream-2026-08-22'
FROM _cur;

COMMIT;
