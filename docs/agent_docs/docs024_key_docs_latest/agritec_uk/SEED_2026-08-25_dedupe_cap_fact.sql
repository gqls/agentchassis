-- ============================================================================
-- agritec.uk — de-duplicate the £100,000 SFI26 cap: two facts, one fact
-- Written 2026-08-25. Applied out of band (psql -f), per-site setup.
--
-- WHY. The first real evidence-freshness sweep (09:06Z) filed a
-- `fact_binding_suggested` note against tool-sfi26-revenue-stacker proposing
-- TWO fact ids for the single `100000` in its script:
--   CIT-3f1b219f15ec6a39  "SFI26 agreements have an annual agreement value cap
--                          of £100,000 per year"
--   CIT-86c4010f7cdf820d  "The SFI26 annual agreement value cap is £100,000 per
--                          agreement year, and the application service will
--                          prevent submission of applications exceeding this limit"
-- Both are true, both cite the same GOV.UK page, and both assert the SAME FACT
-- from two different sentences of it. Declaring both against one constant would
-- owe a reconciliation nobody can perform.
--
-- I SAW THIS AND CALLED IT HARMLESS. On 2026-08-22, reviewing the SFI run, I
-- wrote: "Two near-duplicates (#2/#10) - harmless." It was not harmless; it was
-- latent. What made it visible was a consumer that had to CHOOSE between them —
-- and until something had to choose, "two true facts" and "one fact twice" were
-- indistinguishable. **A duplicate is harmless exactly until something reads it.**
--
-- KEEPING CIT-86c4010f7cdf820d, dropping CIT-3f1b219f15ec6a39:
--   - its quote is the operative rule ("You can only apply for a maximum SFI26
--     agreement value of £100,000 per agreement year") rather than a restatement;
--   - "per agreement year" is the scheme's own wording and more precise than
--     "per year";
--   - it additionally records the enforcement (the application service refuses
--     submissions above the limit), which the calculator's cap message can use.
--
-- ⚠ NOT TOUCHING THE OTHER NINE SHARED VALUES, and they are not duplicates.
-- The register carries nine further pairs of facts with equal values —
-- (UPL10, CNUM2) at 102, (WBD6, WBD7) at 115, (GRH8, OFC1) at 187, (OFC5, OFM6)
-- at 1920, (CAHL4, CIGL2) at 515, (HEF6, CIPM3) at 55, (BFS1, OFM5) at 707,
-- (AHW5, WBD3) at 765, and (CIT-f88b5cd, OFM1) at 20. Every one is a pair of
-- DIFFERENT SFI26 ACTIONS that happen to be paid the same rate. Collapsing them
-- would destroy real facts. Equal value is not duplication, and the test is
-- whether the two rows assert the same thing — not whether they carry the same
-- number.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; found_cnt int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT count(*) INTO found_cnt FROM _cur, LATERAL jsonb_array_elements(data->'facts') x
   WHERE x->>'id' IN ('CIT-3f1b219f15ec6a39','CIT-86c4010f7cdf820d');
  IF found_cnt <> 2 THEN RAISE EXCEPTION 'expected both cap facts present, found %', found_cnt; END IF;
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
      WHERE f->>'id' <> 'CIT-3f1b219f15ec6a39')),
  'manual',
  'De-duplicated the £100,000 SFI26 annual agreement cap: two facts asserted the same rule from two sentences of the same page, and the first real evidence sweep proposed BOTH against the calculator''s single 100000 constant. Kept CIT-86c4010f7cdf820d (operative quote, scheme''s own "per agreement year" wording, records the enforcement). The nine other equal-value pairs are DIFFERENT ACTIONS sharing a rate and are deliberately untouched.',
  true, true, 'agritec-workstream-2026-08-25'
FROM _cur;

COMMIT;
