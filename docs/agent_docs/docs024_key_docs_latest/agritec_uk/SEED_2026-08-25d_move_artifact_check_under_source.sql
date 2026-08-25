-- ============================================================================
-- agritec.uk — move artifact_check INSIDE source, where the code actually reads it
-- Written 2026-08-25. Applied out of band (psql -f), per-site setup.
--
-- THE DEFECT. I wrote the four artifact_check objects at the TOP LEVEL of each
-- fact. The mechanism reads them from inside `source`. Verified at the source
-- rather than taken on report:
--
--   refresh_evidence_base_action.go:348
--       if src, ok := fact["source"].(map[string]interface{}); ok {
--           if _, hasAC := src["artifact_check"]; hasAC && ...
--   parseArtifactCheck (:767)
--       raw, ok := src["artifact_check"].(map[string]interface{})
--
-- So the fence was **live in the register and invisible to the mechanism**: four
-- entries present, correct contents, correct patterns, correct subject_key, and
-- **never evaluated**. Nothing anywhere would have said so. The way I would
-- eventually have found it is by noticing nothing ever fired, which is the worst
-- way to find anything.
--
-- WHY I PUT IT THERE, because the reasoning was not arbitrary. `artifact_check`
-- describes the FACT — "this figure should appear in that artefact" — not the
-- SOURCE, so the top level read as the natural home. RFC_025 placed it under
-- `source` because it was designed as a sibling of citation/artifact/attested_by,
-- a verification mechanism belonging to a source kind. Both readings are
-- defensible; only one is implemented. **What made it expensive is that the
-- wrong one is SILENT** — a declaration nobody can read looks identical to no
-- declaration at all, which is the same shape as the criteria-fence defect the
-- `bugs_open/288` lane had already fixed one table over.
--
-- HOW IT SURFACED, and the part worth keeping. The 288 lane reported my fence as
-- "never landed" — a false negative from counting `f->'source' ? 'artifact_check'`
-- against a populated row. I checked before replying, found the entries at the
-- top level, and pushed back. Their query was wrong AND their conclusion was
-- righter than either of us knew: the entries existed and were inert. **Neither
-- of us would have got there from the other's answer alone** — I would have
-- reported "it is live", they would have reported "it never landed", and both
-- would have been describing the same broken thing.
--
-- WHAT TO EXPECT AFTER THIS. A dry run should show four entries with
-- tolerance="artifact_check". Three of the four carry BOTH citation and
-- attested_by, so they take the secondary path and their artifact entry will have
-- **no verified_at** — that is correct, not a miss.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; top int; under int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT count(*) INTO top   FROM _cur, LATERAL jsonb_array_elements(data->'facts') x WHERE x ? 'artifact_check';
  SELECT count(*) INTO under FROM _cur, LATERAL jsonb_array_elements(data->'facts') x WHERE x->'source' ? 'artifact_check';
  IF top <> 4 THEN RAISE EXCEPTION 'expected 4 top-level artifact_check entries to move, found %', top; END IF;
  IF under <> 0 THEN RAISE EXCEPTION 'expected 0 already under source, found % - state has moved', under; END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id, 'evidence_base',
  jsonb_set(_cur.data, '{facts}',
    (SELECT jsonb_agg(
       CASE WHEN f ? 'artifact_check'
            THEN (f - 'artifact_check')
                 || jsonb_build_object('source', (f->'source') || jsonb_build_object('artifact_check', f->'artifact_check'))
            ELSE f END ORDER BY ord)
     FROM jsonb_array_elements(_cur.data->'facts') WITH ORDINALITY t(f,ord))),
  'manual',
  'Moved the four artifact_check objects from the top level of the fact into source, where refresh_evidence_base_action.go:348 and parseArtifactCheck:767 actually read them. The fence was live in the register and invisible to the mechanism: correct contents, never evaluated, and nothing would have said so.',
  true, true, 'agritec-workstream-2026-08-25'
FROM _cur;

COMMIT;
