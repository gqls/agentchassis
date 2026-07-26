\set ON_ERROR_STOP on
-- ROLLBACK the chart layer: restore the exact evidence_base row that was current
-- before evidence_base_charts_2026-07-26.sql ran.
--
-- Restores from the snapshot table the seed took, not by un-editing JSON, so
-- anything else the seed touched (the F3/F3b date correction, schema_notes)
-- goes back too. Forward-only in git terms — this writes a NEW current row
-- carrying the old data rather than deleting history.
--
-- NOTE: if refresh_evidence_base has run since the seed, it will have rewritten
-- `value`/`verified_at` on the SQL-sourced facts. Rolling back therefore also
-- reverts those refreshed values to their pre-seed state. Check first:
--   SELECT created_at, source_agent, notes FROM site_specs
--    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
--      AND aspect = 'evidence_base' ORDER BY created_at DESC LIMIT 5;

\set site_id '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM bak_site_specs_fai_evidence_20260726;
  IF n <> 1 THEN
    RAISE EXCEPTION 'snapshot table does not hold exactly 1 row (found %) — refusing to roll back', n;
  END IF;
END $$;

UPDATE site_specs
   SET is_current = false, superseded_at = now()
 WHERE site_id = :'site_id'::uuid AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
SELECT site_id, aspect, data, source, source_agent, true, 'brochure_component_library',
       'ROLLBACK of the 2026-07-26 chart layer — restored from bak_site_specs_fai_evidence_20260726'
  FROM bak_site_specs_fai_evidence_20260726;

COMMIT;

\echo '--- rolled back: charts key should now be absent ---'
SELECT (data ? 'charts') AS has_charts, jsonb_array_length(data->'facts') AS facts
  FROM site_specs
 WHERE site_id = :'site_id'::uuid AND aspect = 'evidence_base' AND is_current;
