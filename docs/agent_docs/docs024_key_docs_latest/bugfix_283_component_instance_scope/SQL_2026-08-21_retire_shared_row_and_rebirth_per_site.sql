-- 283: retire the shared automation-savings row and give each site a fresh birth
-- (OWNER RULING 2026-08-21 evening, decision 1). Row 795c34e6 is placed on
-- ai-agent-orchestration.com AND fundamentallyai.com, serving composition-broken
-- bytes on both; site-scoped regeneration refuses cross-site rows and the armed
-- fork guard (correctly) refuses forking a broken template — so: deactivate the
-- row (it remains its own pre-image; component_versions also holds snapshots),
-- tombstone its two slots (build_status='removed', the estate's retire-a-slot
-- verb; assembly excludes it, the archive trigger preserved served bytes), and
-- let two fresh add_tool births adopt the existing pages (adopt_existing_page,
-- the bugs_open/286 ported-tool route) — on v1.0.1322 the armed birth guard
-- means both are born CONVERTED, which is also the guard's live demand check.
BEGIN;
DO $$
DECLARE n int;
BEGIN
  SELECT count(DISTINCT p.site_id) INTO n
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE pc.component_id='795c34e6-dd41-4984-a4ee-e547f29a0643'
    AND pc.build_status IS DISTINCT FROM 'removed';
  IF n <> 2 THEN
    RAISE EXCEPTION 'retire-shared-row: expected the row live on 2 sites, found % — re-census before applying', n;
  END IF;
END $$;

UPDATE content_components SET is_active=false, updated_at=now()
WHERE id='795c34e6-dd41-4984-a4ee-e547f29a0643' AND is_active=true;

UPDATE page_components SET build_status='removed', updated_at=now()
WHERE component_id='795c34e6-dd41-4984-a4ee-e547f29a0643'
  AND build_status IS DISTINCT FROM 'removed';

-- Arm page adoption for the two births (temporary; removed after they land —
-- the un-arm statement lives beside this file's application record in NOTES):
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,save_tool,config,adopt_existing_page}', 'true'::jsonb, true),
    updated_at = now()
WHERE type='tool-generator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE a boolean; nrow int; nslot int;
BEGIN
  SELECT count(*) INTO nrow FROM content_components
   WHERE id='795c34e6-dd41-4984-a4ee-e547f29a0643' AND is_active=false;
  SELECT count(*) INTO nslot FROM page_components
   WHERE component_id='795c34e6-dd41-4984-a4ee-e547f29a0643' AND build_status='removed';
  SELECT (default_config#>>'{workflow,steps,save_tool,config,adopt_existing_page}')::boolean INTO a
   FROM agent_definitions WHERE type='tool-generator' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF nrow <> 1 OR nslot <> 2 OR NOT a THEN
    RAISE EXCEPTION 'retire-shared-row verify: row_inactive=% slots_removed=% adopt=%', nrow, nslot, a;
  END IF;
END $$;
COMMIT;
