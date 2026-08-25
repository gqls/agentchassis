-- ============================================================================
-- agritec.uk — anchor two patterns, and give the compound-unit rates a value
-- Written 2026-08-25, from a review by the bugs_open/288 lane.
--
-- FIX 1 — TWO OF FOUR PATTERNS WERE UNDER-ANCHORED, and my own test could not
-- have found it. `code:'CSAM3'[^}]*rate:224` also matches `rate:2240`, and
-- `...CSAM2...rate:129` also matches `rate:1290`. CHRW2 and CIPM4 carried `\b`
-- and correctly reject a longer rate; the inconsistency is the tell that it was
-- not a decision. This register already holds four-digit rates (1920, 1072), so
-- it is not hypothetical: a rate gaining a digit would report `fresh` on a tool
-- that had changed.
--
-- ⚠ WHY MY MUTATION TEST PASSED ANYWAY, which is the transferable part. I
-- mutated the PATTERN — changed the rate inside it and confirmed it then matched
-- nothing. That proves the pattern is load-bearing. It says nothing about a
-- mutation of the ARTEFACT, which is where this failure lives. **A MUTATION ONLY
-- PROVES WHAT IT MUTATES.** Re-tested the right way (rewriting the rate in the
-- template and re-running the pattern): both under-anchored patterns still
-- matched, and both anchored ones correctly refused.
--
-- FIX 2 — THE COMPOUND-UNIT RATES HAD NO VALUE, and that was my decision. When
-- generating these facts I set `value` only where the payment string was an
-- unambiguous "£N per hectare|square metre", reasoning that "a compound rate
-- reduced to one number is how 'per 100m for one side' silently becomes 'per
-- 100m', which on a hedgerow is a factor-of-two error." That reasoning was
-- sound about PRESENTATION and wrong about STORAGE: it left five facts asserting
-- a rate with no figure, so the copy path has nothing to substitute and the
-- probe reports `not_probed` for want of a number.
--
-- The resolution keeps both properties: store the value, and carry the qualifier
-- in the `unit` string — "GBP per 100m for one side", not "GBP per 100m". The
-- writer_line already carries the full published wording and remains the control
-- over how it is stated, which is the same conclusion this lane reached about
-- lossy values on the ONS price facts.
--
-- CHRW2 matters most of the five: hedgerow management is one of the two rates
-- the retired calculator had wrong, and it is one of the four fenced facts.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; fenced int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT count(*) INTO fenced FROM _cur, LATERAL jsonb_array_elements(data->'facts') x
   WHERE x ? 'artifact_check';
  IF fenced <> 4 THEN
    RAISE EXCEPTION 'expected the 4 artifact_check entries to be present, found % - the fence is not where I think it is', fenced;
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
         -- anchor the two loose patterns
         WHEN f->>'id' = 'ATT-sfi26-CSAM3' THEN
              jsonb_set(f,'{artifact_check,pattern}', to_jsonb('code:''CSAM3''[^}]*rate:224\b'::text))
         WHEN f->>'id' = 'ATT-sfi26-CSAM2' THEN
              jsonb_set(f,'{artifact_check,pattern}', to_jsonb('code:''CSAM2''[^}]*rate:129\b'::text))
         -- give the compound-unit rates a value, qualifier carried in the unit
         WHEN f->>'id' = 'ATT-sfi26-CHRW2' THEN
              jsonb_set(jsonb_set(f,'{value}','13'::jsonb),'{unit}', to_jsonb('GBP per 100m for one side'::text))
         WHEN f->>'id' = 'ATT-sfi26-BND2' THEN
              jsonb_set(jsonb_set(f,'{value}','11'::jsonb),'{unit}', to_jsonb('GBP per 100m for one side'::text))
         WHEN f->>'id' = 'ATT-sfi26-BND1' THEN
              jsonb_set(jsonb_set(f,'{value}','27'::jsonb),'{unit}', to_jsonb('GBP per 100m for both sides'::text))
         WHEN f->>'id' = 'ATT-sfi26-WBD2' THEN
              jsonb_set(jsonb_set(f,'{value}','4'::jsonb),'{unit}', to_jsonb('GBP per 100m for both sides'::text))
         WHEN f->>'id' = 'ATT-sfi26-WBD1' THEN
              jsonb_set(jsonb_set(f,'{value}','257'::jsonb),'{unit}', to_jsonb('GBP per pond (maximum 3 ponds per hectare)'::text))
         WHEN f->>'id' = 'ATT-sfi26-AHW4' THEN
              jsonb_set(jsonb_set(f,'{value}','11'::jsonb),'{unit}', to_jsonb('GBP per plot (minimum 2 plots)'::text))
         WHEN f->>'id' = 'ATT-sfi26-AHW2' THEN
              jsonb_set(jsonb_set(f,'{value}','732'::jsonb),'{unit}', to_jsonb('GBP per tonne (maximum 1 tonne per 2 hectares of CAHL2)'::text))
         ELSE f END ORDER BY ord)
     FROM jsonb_array_elements(_cur.data->'facts') WITH ORDINALITY t(f,ord))),
  'manual',
  'From the bugs_open/288 review: anchored CSAM3 and CSAM2 with \b (they matched rate:2240 and rate:1290 — my mutation test changed the PATTERN, not the ARTEFACT, so it could not have caught this), and gave the seven compound-unit rates a value with the qualifier carried in the unit string rather than dropped.',
  true, true, 'agritec-workstream-2026-08-25'
FROM _cur;

COMMIT;
