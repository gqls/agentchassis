-- ============================================================================
-- agritec.uk — artifact_check on the four SFI26 rates that matter most
-- Written 2026-08-25. Applied out of band (psql -f), per-site setup.
--
-- WHY THESE FOUR. They are the retired calculator's surviving actions, and two
-- of them were WRONG on it: herbal leys ran £382 against an actual £224, and
-- hedgerow management ran £10 per 100m against £13 per 100m FOR ONE SIDE. The
-- other two (CSAM2 £129, CIPM4 £45) kept their rate and changed their code. If
-- any of the four silently drifts again, this is what says so.
--
-- WHY artifact_check RATHER THAN THE AUTOMATED PROBE. The `bugs_open/288` lane's
-- code probe has a MEASURED distinctiveness floor of 1000 (false-positive rate
-- 3.79% at two digits, 32.75% at one). Every rate here — 224, 129, 45, 13 — is
-- below it and correctly REFUSED rather than guessed at. Their sweep of this
-- site on 2026-08-25 found exactly one value it would probe, the £100,000 cap.
-- So an author-written, context-bearing pattern is the only route for these, and
-- it is admissible precisely because the ACTION CODE does the discriminating
-- where a bare `224` could not.
--
-- THE PATTERNS WERE READ OFF THE BUILT ARTEFACT, NOT GUESSED. The tool stores
-- its rates as `{code:'CSAM3', name:'Herbal leys', rate:224, unit:'per_hectare',
-- display:'£224 per hectare'}`, so `code:'CSAM3'[^}]*rate:224` binds the code to
-- its rate. Tested on both arms against the live template: all four match, and
-- all four find NOTHING when the rate in the pattern is mutated. A pattern that
-- cannot fail is not a check.
--
-- ⚠ THE PINNED VALUE IS DELIBERATE, AND IT WILL GO RED ON PURPOSE. When DEFRA
-- moves a rate, the tool's constant and this pattern are meant to change in the
-- SAME edit. The window between them is the check firing, not a fault. Budget
-- for updating both together, or it sits permanently red and someone learns to
-- ignore it — which is worse than not having it.
--
-- ⚠ AND DO NOT READ A ZERO YET. The `bugs_open/288` machinery had its first real
-- sweep at 09:06Z today. Until a sweep has run AFTER this migration, a zero from
-- it means "it has not run", not "nothing to find".
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
   WHERE x->>'id' IN ('ATT-sfi26-CSAM3','ATT-sfi26-CHRW2','ATT-sfi26-CSAM2','ATT-sfi26-CIPM4');
  IF hits <> 4 THEN RAISE EXCEPTION 'expected the 4 target rate facts, found %', hits; END IF;
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
         WHEN 'ATT-sfi26-CSAM3' THEN jsonb_set(f,'{artifact_check}', jsonb_build_object(
              'subject_key','tool-sfi26-revenue-stacker','must_be_present',true,
              'pattern','code:''CSAM3''[^}]*rate:224'))
         WHEN 'ATT-sfi26-CHRW2' THEN jsonb_set(f,'{artifact_check}', jsonb_build_object(
              'subject_key','tool-sfi26-revenue-stacker','must_be_present',true,
              'pattern','code:''CHRW2''[^}]*rate:13\b'))
         WHEN 'ATT-sfi26-CSAM2' THEN jsonb_set(f,'{artifact_check}', jsonb_build_object(
              'subject_key','tool-sfi26-revenue-stacker','must_be_present',true,
              'pattern','code:''CSAM2''[^}]*rate:129'))
         WHEN 'ATT-sfi26-CIPM4' THEN jsonb_set(f,'{artifact_check}', jsonb_build_object(
              'subject_key','tool-sfi26-revenue-stacker','must_be_present',true,
              'pattern','code:''CIPM4''[^}]*rate:45\b'))
         ELSE f END ORDER BY ord)
     FROM jsonb_array_elements(_cur.data->'facts') WITH ORDINALITY t(f,ord))),
  'manual',
  'artifact_check on the four SFI26 rates the retired calculator got wrong or recoded (CSAM3 was £382 vs £224; CHRW2 was £10 vs £13 for one side; CSAM2 and CIPM4 kept their rate, changed code). Patterns read off the built artefact and tested on both arms: all four match, all four find nothing when mutated. These four are below the 288 probe''s measured floor of 1000, so an author-written context-bearing pattern is the only route.',
  true, true, 'agritec-workstream-2026-08-25'
FROM _cur;

COMMIT;
