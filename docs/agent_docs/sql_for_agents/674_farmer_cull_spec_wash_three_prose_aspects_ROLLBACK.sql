-- 674 ROLLBACK — restore the three predecessor specs, remove the 674 successors.
-- REFUSES if the current rows are not 674's own successors (the analyser or another producer
-- has since superseded them — rolling back would then destroy newer work, not restore ours).
BEGIN;
DO $rb$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0'
     AND aspect IN ('briefing','strategy','vertical_landscape')
     AND is_current AND created_by='copy_quality_two_stage';
  IF n <> 3 THEN RAISE EXCEPTION '674 ROLLBACK: % of 3 current rows are 674 successors — a newer supersede exists; do NOT roll back blind, re-census', n; END IF;

  DELETE FROM site_specs
   WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0'
     AND aspect IN ('briefing','strategy','vertical_landscape')
     AND is_current AND created_by='copy_quality_two_stage';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 3 THEN RAISE EXCEPTION '674 ROLLBACK: deleted % successor rows, want 3', n; END IF;

  UPDATE site_specs SET is_current=true, superseded_at=NULL
   WHERE id IN ('38666a42-e9b5-4a81-aa2f-ab6aa2da9c44','43b01298-b25c-4e6c-abfb-0506bd0e3d78','a7ce9ee2-c485-4f6f-89c9-cab311e13fdb');
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 3 THEN RAISE EXCEPTION '674 ROLLBACK: restored % predecessors, want 3', n; END IF;
END $rb$;
DELETE FROM schema_migrations WHERE filename='674_farmer_cull_spec_wash_three_prose_aspects.sql';
COMMIT;
