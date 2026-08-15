-- 422: the publish reconciler — scheduled drift sweep + site-publisher
-- repurposed onto the publish seam (site_delivery_and_editor Phase 2).
-- REVISED 2026-08-15 after council round 1 (corr 21aba3f5, gating objection
-- from debug_historian): the snapshot now uses the sanctioned two-arg
-- snapshot_agent() into agent_definitions_backup (my hand-rolled partial
-- INSERT would have dropped topics/capabilities/image fields from any
-- restore), re-application is a graceful no-op, and the post-update state is
-- read back rather than assumed.
--
-- ⚠ WHY THIS IS A _HOLD FILE (ordering-critical, excluded from --apply by
-- SIDECAR_RE): it seeds workflows naming the `publish_site` action, which
-- exists only in images built from commit 71e4d9736 onward. A seed naming an
-- unregistered action fails at runtime (CLAUDE.md: image first, then seeds).
--
-- BEFORE APPLYING, VERIFY THE RUNNING POD — not git, not the image tag
-- (per-SERVICE, both replicas if you doubt one):
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor 71e4d9736 <the stamp sha>   # must exit 0
-- The stamp is a STARTUP line and scrolls on a busy service; an empty grep
-- means "not in range", not "unstamped". Fallback binary probe, WITH BOTH
-- controls in the same breath (a lone absent-grep on a Go binary reads
-- absent even when the commit shipped — LANDMINES; never `strings`, never a
-- marker containing an em dash or any non-ASCII byte):
--   P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
--   kubectl -n ai-persona-system exec ${P#pod/} -- grep -aq "<full sha of the STAMP>" /proc/1/exe && echo stamp-present
--   kubectl -n ai-persona-system exec ${P#pod/} -- grep -aq "7f3a9c2e8b1d4f6a0c5e9b3d7a1f8c4e2b6d0a9f" /proc/1/exe && echo CONTROL-BROKEN-random-hex-matches
-- (positive control: the stamp sha must match; negative control: the random
-- hex value must NOT — if it does, the probe is not discriminating and
-- proves nothing. ⚠ Do NOT use the all-zeros sha as the negative control:
-- it is git's null-sha CONSTANT and legitimately embedded in git-aware
-- binaries — it matched on v1.0.1303, a false CONTROL-BROKEN, measured
-- 2026-08-15 while the random-hex control stayed clean and the stamp probe
-- correctly discriminated 1-of-4 candidate shas.)
--
-- ALSO RE-VERIFY THE ZERO-CONSUMER CLAIM AT APPLY TIME (council round 2,
-- prior_art seat: two of its four legs read tables outside some reviewers'
-- reach, and days may have passed — an enumeration is evidence only at the
-- moment it ran). All four must still be 0:
--   SELECT count(*) FROM agent_definitions WHERE default_config::text LIKE '%site-publisher%'
--     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL AND type <> 'site-publisher';
--   SELECT count(*) FROM scheduled_tasks WHERE target_agent_type='site-publisher';
--   SELECT count(*) FROM site_work_items WHERE handler_agent='site-publisher';
--   SELECT count(*) FROM orchestration_states
--     WHERE collected_data->'__execution_context__'->>'run_agent_type'='site-publisher';  -- slow (~2 min), full scan
--
-- THEN apply with exactly this command from the repo root:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--     -f - < docs/agent_docs/sql_for_agents/422_site_publish_reconciler_HOLD.sql
-- ⚠ There is NO --record-only step: the runner REFUSES to record
-- UPPERCASE-suffixed sidecars (by design — they are never in its apply
-- model), measured 2026-08-15. The apply record is the lane NOTES entry +
-- this header. STATUS: APPLIED 2026-08-15 ~22:00Z on v1.0.1303 (stamp
-- 5e075a6f9, 71e4d9736 ancestry verified, 4-leg enumeration re-run = 0).
-- Re-running after a successful apply is SAFE: the repurpose block detects
-- the applied state and no-ops with a NOTICE (it will not stack snapshots).
--
-- WHAT IT DOES (all inert until a site opts in — sites.publish_target is
-- NULL fleet-wide, migration 412):
--   1. site_publish_checks — the reconciler's selection stamp (records
--      SELECTION not completion, the migration-346 rotation pattern, so a
--      failed run cannot pin the rotation head).
--   2. Repurposes the LIVE site-publisher definition (a pre-070-refactor
--      fossil: object-format workflow, upload_to_s3 into a bucket "websites"
--      that does not exist; ZERO consumers by query — 0 workflows, 0
--      schedules, 0 work items, 0 orchestrations all-history) onto the seam.
--      Guarded on the fossil's exact pre-state; full-row snapshot via
--      snapshot_agent(type, reason) FIRST, and the snapshot is verified to
--      hold the PRE-change config before anything mutates (a snapshot
--      carrying the post-change value restores nothing — LANDMINES).
--      site-publisher rather than a new type because the spawner's storage
--      allow-list and topic_manager already carry the name — a spawned
--      site-publisher pod is exactly the credentialed environment
--      publish_site requires (the standing chassis carries no B2
--      credentials by owner ruling 2026-08-08).
--   3. publish-reconciler — a small orchestrator the scheduler dispatches:
--      spawn site-publisher, call it with {site_id, domain}, complete. The
--      spawn→call pair is the estate's standing (and known-racy) handshake;
--      a failed handshake is retried by the next tick, not by this workflow.
--   4. scheduled_tasks row site-publish-reconciler: every 600s, pre_query
--      picks the least-recently-checked opted-in site and stamps it; zero
--      opted-in sites -> zero rows -> the gate skips and nothing fires.
--
-- Note (editquality seat, same round): nothing prevents an operator setting
-- publish_target='cfpages' before that backend is armed — by design. The
-- refusal is loud, names the working backend, and recurs each tick until
-- the target is corrected or cfpages is armed. Do not "fix" this with a
-- CHECK constraint: arming cfpages would then need migration churn.
--
-- ROLLBACK RECIPE: 422_site_publish_reconciler_HOLD_ROLLBACK.sql (restores
-- from agent_definitions_backup by snapshot_reason, newest snapshot_taken_at).

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
  snap_action text;
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

  -- Idempotent re-application: already carrying the seam workflow means a
  -- previous apply succeeded — no-op, and take no second snapshot.
  IF cur_action = 'publish_site' THEN
    RAISE NOTICE '422: site-publisher already carries publish_site — previously applied, skipping repurpose';
    RETURN;
  END IF;

  -- Guard on the fossil's exact pre-state; anything else means another
  -- session got here first or the row is not what 2026-08-15 measured.
  IF cur_action IS DISTINCT FROM 'upload_to_s3' THEN
    RAISE EXCEPTION '422: site-publisher pre-state mismatch — steps.publish.action is %, expected the upload_to_s3 fossil. Read the row before re-running.', cur_action;
  END IF;

  -- Full-row snapshot via the SANCTIONED mechanism. Two-arg form ->
  -- agent_definitions_backup (the one-arg form writes an is_snapshot row to
  -- a DIFFERENT table — the documented dual-overload trap). The distinctive
  -- reason is what makes OUR snapshot findable: backup rows copy id AND
  -- created_at from the source, so only snapshot_reason + snapshot_taken_at
  -- identify a specific snapshot (LANDMINES).
  PERFORM snapshot_agent('site-publisher', '422 pre-repurpose: upload_to_s3 fossil');

  -- A snapshot EXISTING is not the check — it must hold the PRE-change
  -- config, or the restore path is already broken. Abort before mutating.
  SELECT b.default_config#>>'{workflow,steps,publish,action}' INTO snap_action
    FROM agent_definitions_backup b
   WHERE b.type = 'site-publisher'
     AND b.snapshot_reason = '422 pre-repurpose: upload_to_s3 fossil'
   ORDER BY b.snapshot_taken_at DESC LIMIT 1;
  IF snap_action IS DISTINCT FROM 'upload_to_s3' THEN
    RAISE EXCEPTION '422: snapshot does not hold the pre-change config (found %) — restore would be impossible, aborting with nothing mutated', snap_action;
  END IF;

  UPDATE agent_definitions
     SET description = 'Publishes one site''s built artefact tree to its opted-in hosting backend via the publish seam (platform/publish). Runs publish_site, which no-ops with a recorded reason when sites.publish_target is NULL (the default), when the tree has no drift against sites.published_hash, or when nothing is built. Must run as a SPAWNED pod: it is on the spawner''s storage allow-list, and the standing chassis carries no B2 credentials. Dispatch with input_data {site_id, domain}. Pre-422 fossil config in agent_definitions_backup, snapshot_reason ''422 pre-repurpose: upload_to_s3 fossil''.',
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

  -- Post-condition: read back the row just written, never assume the write.
  SELECT default_config#>>'{workflow,steps,publish,action}' INTO cur_action
    FROM agent_definitions WHERE id = live_id;
  IF cur_action IS DISTINCT FROM 'publish_site' THEN
    RAISE EXCEPTION '422: post-update read-back shows steps.publish.action = %, want publish_site', cur_action;
  END IF;
END $repurpose$;

-- 3. The dispatching orchestrator ----------------------------------------
INSERT INTO agent_definitions (type, display_name, description, category, agent_category, status, is_active, default_config)
SELECT
  'publish-reconciler',
  'Publish Reconciler',
  'Scheduler-dispatched shim: spawns a site-publisher pod (the storage-credentialed environment) and calls it with one site''s {site_id, domain}. All publish logic and all no-op decisions live in site-publisher''s publish_site action; this workflow only crosses the credential boundary. Fired by scheduled_tasks row site-publish-reconciler.',
  'orchestrator',
  -- agent_category is CHECK-constrained (check_ad_category): only
  -- strategist/executor/analyst/integrator/coordinator/specialist or NULL.
  -- 'orchestrator' lives in the UNconstrained category column above.
  'coordinator',
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

  -- The restore path must actually exist: a backup row under OUR reason
  -- holding the PRE-change config (not merely any snapshot).
  SELECT count(*) INTO n FROM agent_definitions_backup
   WHERE type = 'site-publisher'
     AND snapshot_reason = '422 pre-repurpose: upload_to_s3 fossil'
     AND default_config#>>'{workflow,steps,publish,action}' = 'upload_to_s3';
  IF n < 1 THEN
    RAISE EXCEPTION '422 verify: no restorable pre-repurpose snapshot in agent_definitions_backup';
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
