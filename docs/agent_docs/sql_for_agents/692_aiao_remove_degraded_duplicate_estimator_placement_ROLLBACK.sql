-- 692_..._ROLLBACK.sql
--
-- Re-inserts the degraded duplicate estimator placement removed by 692, from the
-- full-row backup that migration wrote.
--
-- ⚠ WHAT THIS RESTORES IS THE DAMAGE, NOT A WORKING STATE. The row is a REDUCED
-- regeneration of the complexity estimator — 1 fieldset, 1 legend, **1 input**, against
-- the surviving tool's 4 / 4 / **12** — and putting it back makes
-- /tools/agent-complexity-estimator.html serve the same calculator TWICE (two
-- `<h2>Agent Architecture Complexity Estimator` headings), with the second copy also
-- re-opening the 1.04:1 button contrast defect that migration 625 closed.
-- Restore it only to hand the artefact to someone diagnosing the producer, and say so.
--
-- ⚠ The page must be re-assembled for this to reach the site, and it is
-- `rebuild_policy='owned'` — use `refresh_owned_page_chrome.sh` (assemble mode), never
-- a generic rebuild. See 625, and see the LANDMINE about running that script under a
-- command timeout.

BEGIN;

INSERT INTO page_components
SELECT (jsonb_populate_record(NULL::page_components, b.old_value)).*
FROM migration_backups b
WHERE b.migration_name='692_aiao_remove_degraded_duplicate_estimator_placement'
  AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.id = (b.old_value->>'id')::uuid);

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM page_components WHERE id='9aa63fc0-5edf-4768-8e90-cade95e6cf34';
  IF n <> 1 THEN RAISE EXCEPTION 'rollback 692: the removed placement was not restored (found %)', n; END IF;
  RAISE NOTICE 'rollback 692 OK: the DEGRADED duplicate is back. The page will serve two estimators once re-assembled.';
END $$;

COMMIT;
