-- 422: the publish reconciler — scheduled drift sweep + site-publisher
-- repurposed onto the publish seam (site_delivery_and_editor Phase 2).
--
-- ⚠ WHY THIS IS A _HOLD FILE (ordering-critical, excluded from --apply by
-- SIDECAR_RE): it seeds workflows naming the `publish_site` action, which
-- exists only in images built from the commit that ships platform/publish.
-- A seed naming an unregistered action fails at runtime (CLAUDE.md: image
-- first, then seeds). APPLY BY HAND, AFTER the chassis roll that carries
-- publish_site, verified at the binary (grep the build-provenance stamp or
-- ask the pod), with exactly these two commands from the repo root:
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--     -f - < docs/agent_docs/sql_for_agents/422_site_publish_reconciler_HOLD.sql
--   ./scripts/migration/run-migrations.sh --record-only \
--     docs/agent_docs/sql_for_agents/422_site_publish_reconciler_HOLD.sql \
--     --note "hand-applied post-roll; HOLD was for image-before-seed ordering"
--
-- WHAT IT DOES (all inert until a site opts in — sites.publish_target is
-- NULL fleet-wide, migration 412):
--   1. site_publish_checks — the reconciler's selection stamp (records
--      SELECTION not completion, the migration-346 rotation pattern, so a
--      failed run cannot pin the rotation head).
--   2. Repurposes the LIVE site-publisher definition (a pre-070-refactor
--      fossil: object-format workflow, upload_to_s3 into a bucket "websites"
--      that does not exist) onto the seam. Guarded on the fossil's exact
--      pre-state; snapshots the row first. site-publisher rather than a new
--      type because the spawner's storage allow-list and topic_manager
--      already carry the name — a spawned site-publisher pod is exactly the
--      credentialed environment publish_site requires (the standing chassis
--      carries no B2 credentials by owner ruling 2026-08-08).
--   3. publish-reconciler — a small orchestrator the scheduler dispatches:
--      spawn site-publisher, call it with {site_id, domain}, complete. The
--      spawn→call pair is the estate's standing (and known-racy) handshake;
--      a failed handshake is retried by the next tick, not by this workflow.
--   4. scheduled_tasks row site-publish-reconciler: every 600s, pre_query
--      picks the least-recently-checked opted-in site and stamps it; zero
--      opted-in sites -> zero rows -> the gate skips and nothing fires.
--
-- ROLLBACK RECIPE: 422_site_publish_reconciler_HOLD_ROLLBACK.sql.

BEGIN;

-- 1. Selection stamp -----------------------------------------------------
CREATE TABLE IF NOT EXISTS site_publish_checks (
  site_id         uuid PRIMARY KEY REFERENCES sites(id) ON DELETE CASCADE,
  last_checked_at timestamptz NOT NULL DEFAULT now()
);
COMMENT ON TABLE site_publish_checks IS
  'Publish-reconciler rotation stamp: when each opted-in site was last SELECTED for a drift check (not when one completed). Written by the site-publish-reconciler pre_query (migration 422).';

-- 2. Repurpose site-publisher onto the seam ------------------------------
DO $repurpose$
DECLARE
  n int;
  live_id uuid;
  cur_action text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'site-publisher'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '422: expected exactly 1 live site-publisher row, found % (concurrent edit?)', n;
  END IF;

  SELECT id, default_config#>>'{workflow,steps,publish,action}'
    INTO live_id, cur_action
    FROM agent_definitions
   WHERE type = 'site-publisher'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  -- Guard on the fossil's exact pre-state; anything else means another
  -- session got here first or the row is not what 2026-08-15 measured.
  IF cur_action IS DISTINCT FROM 'upload_to_s3' THEN
    RAISE EXCEPTION '422: site-publisher pre-state mismatch — steps.publish.action is %, expected the upload_to_s3 fossil. Read the row before re-running.', cur_action;
  END IF;

  -- Snapshot the fossil before touching it.
  INSERT INTO agent_definitions
        (type, display_name, description, category, agent_category, status,
         is_active, is_snapshot, default_config, created_at)
  SELECT type, display_name,
         COALESCE(description, '') || ' [snapshot before migration 422 repurposed the live row onto platform/publish]',
         category, agent_category, status,
         false, true, default_config, now()
    FROM agent_definitions WHERE id = live_id;

  UPDATE agent_definitions
     SET description = 'Publishes one site''s built artefact tree to its opted-in hosting backend via the publish seam (platform/publish). Runs publish_site, which no-ops with a recorded reason when sites.publish_target is NULL (the default), when the tree has no drift against sites.published_hash, or when nothing is built. Must run as a SPAWNED pod: it is on the spawner''s storage allow-list, and the standing chassis carries no B2 credentials. Dispatch with input_data {site_id, domain}.',
         default_config = jsonb_set(default_config, '{workflow}', jsonb_build_object(
           'start_step', 'publish',
           'timeout_seconds', 600,
           'steps', jsonb_build_object(
             'publish', jsonb_build_object(
               'action', 'publish_site',
               'config', jsonb_build_object(
                 'domain',  'input_data.domain',
                 'site_id', 'input_data.site_id'
               ),
               'next_step', 'complete',
               'description', 'One reconciliation pass: hash the built tree, publish on drift, accept on served bytes',
               'output_field', 'publish_result'
             ),
             'complete', jsonb_build_object(
               'action', 'complete_workflow',
               'config', jsonb_build_object('output_fields', jsonb_build_array('publish_result')),
               'description', 'Return the publish result to the caller'
             )
           )
         )),
         updated_at = now()
   WHERE id = live_id;
END $repurpose$;

-- 3. The dispatching orchestrator ----------------------------------------
INSERT INTO agent_definitions (type, display_name, description, category, agent_category, status, is_active, default_config)
SELECT
  'publish-reconciler',
  'Publish Reconciler',
  'Scheduler-dispatched shim: spawns a site-publisher pod (the storage-credentialed environment) and calls it with one site''s {site_id, domain}. All publish logic and all no-op decisions live in site-publisher''s publish_site action; this workflow only crosses the credential boundary. Fired by scheduled_tasks row site-publish-reconciler.',
  'orchestrator',
  'orchestrator',
  'experimental',
  true,
  jsonb_build_object('workflow', jsonb_build_object(
    'start_step', 'spawn_publisher',
    'processing_mode', 'orchestrator',
    'timeout_seconds', 900,
    'steps', jsonb_build_object(
      'spawn_publisher', jsonb_build_object(
        'action', 'spawn_agent',
        'config', jsonb_build_object('role', 'publisher', 'agent_type', 'site-publisher'),
        'next_step', 'call_publisher',
        'description', 'Spawn the storage-credentialed publisher pod',
        'output_field', 'publisher_spawned'
      ),
      'call_publisher', jsonb_build_object(
        'action', 'call_agent',
        'config', jsonb_build_object(
          'target_role', 'publisher',
          'input_mapping', jsonb_build_object(
            'domain',  'input_data.domain',
            'site_id', 'input_data.site_id'
          ),
          'timeout_seconds', 600
        ),
        'next_step', 'complete',
        'description', 'One reconciliation pass for the selected site',
        'output_field', 'publish_result'
      ),
      'complete', jsonb_build_object(
        'action', 'complete_workflow',
        'config', jsonb_build_object('output_fields', jsonb_build_array('publish_result')),
        'description', 'Done — the next scheduler tick handles the next site'
      )
    )
  ))
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions WHERE type = 'publish-reconciler' AND deleted_at IS NULL
);

-- 4. The schedule ---------------------------------------------------------
INSERT INTO scheduled_tasks
      (name, description, interval_seconds, target_agent_type, target_topic,
       input_data, concurrency_group, max_concurrent, pre_query,
       enabled, timeout_seconds, fire_message)
SELECT
  'site-publish-reconciler',
  'Publish-seam drift sweep: one opted-in site per tick, least recently checked first. Gate skips (zero rows) while no site has publish_target set — the seam is opt-in default OFF (migration 412).',
  600,
  'publish-reconciler',
  'system.agent.scheduled.requests',
  '{}'::jsonb,
  'site-publish',
  1,
  $pq$WITH due AS (
  SELECT s.id AS sid, s.domain
    FROM sites s
    LEFT JOIN site_publish_checks c ON c.site_id = s.id
   WHERE s.publish_target IS NOT NULL
   ORDER BY c.last_checked_at ASC NULLS FIRST
   LIMIT 1
), stamped AS (
  INSERT INTO site_publish_checks (site_id, last_checked_at)
  SELECT sid, now() FROM due
  ON CONFLICT (site_id) DO UPDATE SET last_checked_at = now()
  RETURNING site_id
)
SELECT due.sid::text AS site_id, due.domain FROM due$pq$,
  true,
  900,
  true
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'site-publish-reconciler');

-- 5. Verify ---------------------------------------------------------------
DO $verify$
DECLARE
  n int;
  a text;
BEGIN
  SELECT default_config#>>'{workflow,steps,publish,action}' INTO a
    FROM agent_definitions
   WHERE type = 'site-publisher'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF a IS DISTINCT FROM 'publish_site' THEN
    RAISE EXCEPTION '422 verify: site-publisher publish step action is %, want publish_site', a;
  END IF;

  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'site-publisher' AND COALESCE(is_snapshot, false) = true;
  IF n < 1 THEN
    RAISE EXCEPTION '422 verify: no snapshot of the pre-422 site-publisher exists';
  END IF;

  SELECT count(*) INTO n FROM jsonb_object_keys((
    SELECT default_config->'workflow'->'steps' FROM agent_definitions
     WHERE type = 'publish-reconciler'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL));
  IF n <> 3 THEN
    RAISE EXCEPTION '422 verify: publish-reconciler should have 3 steps, found %', n;
  END IF;

  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'site-publish-reconciler' AND enabled = true AND fire_message = true;
  IF n <> 1 THEN
    RAISE EXCEPTION '422 verify: scheduled row site-publish-reconciler absent or disabled';
  END IF;
END $verify$;

COMMIT;
