-- 764 ROLLBACK — restore both agents' default_config from the pre-image snapshot_agent() took
-- (agent_definitions_backup, snapshot_reason LIKE '764_%'). Restores the WHOLE default_config of each
-- agent to its 764 pre-image, so anything applied to these two rows AFTER 764 is reverted too — check
-- `SELECT type, updated_at FROM agent_definitions WHERE type IN (…)` against the backup's
-- snapshot_taken_at before running, and prefer a forward fix if anything landed in between.
BEGIN;
DO $$
DECLARE rec record; n int;
BEGIN
  FOR rec IN SELECT DISTINCT ON (type) type, default_config, snapshot_taken_at
               FROM agent_definitions_backup
              WHERE snapshot_reason LIKE '764_%' AND type IN ('domain-research-classifier','build-site-planner')
              ORDER BY type, snapshot_taken_at DESC LOOP
    UPDATE agent_definitions SET default_config = rec.default_config, updated_at = now()
     WHERE type = rec.type AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    GET DIAGNOSTICS n = ROW_COUNT;
    IF n <> 1 THEN RAISE EXCEPTION '764 ROLLBACK: % restored % rows, expected 1', rec.type, n; END IF;
    RAISE NOTICE '764 ROLLBACK: % restored from snapshot taken %', rec.type, rec.snapshot_taken_at;
  END LOOP;
  IF (SELECT count(*) FROM agent_definitions WHERE type IN ('domain-research-classifier','build-site-planner') AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL AND default_config::text LIKE '%{{toJSON .site_specs.specs.mission_brief}}%') <> 0 THEN
    RAISE EXCEPTION '764 ROLLBACK VERIFY: fallback expression still present'; END IF;
END $$;
COMMIT;
