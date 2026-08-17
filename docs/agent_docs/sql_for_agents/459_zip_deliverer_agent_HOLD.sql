-- 459: zip-deliverer + its dispatch shim — the ZIP deliverable's agents
-- (site_delivery_and_editor Phase 3, register DGH-011).
--
-- ⚠ WHY THIS IS A _HOLD FILE (ordering-critical, excluded from --apply by
-- SIDECAR_RE): it seeds workflows naming the `zip_deliverable` action, which
-- exists only in images built from THE PHASE 3 COMMIT onward (the commit that
-- adds platform/orchestration/actions/zip_deliverable_action.go — take its
-- sha from `git log --oneline -- platform/orchestration/actions/zip_deliverable_action.go`).
-- A seed naming an unregistered action fails at runtime (CLAUDE.md: image
-- first, then seeds).
--
-- BEFORE APPLYING, VERIFY THE RUNNING POD — not git, not the image tag
-- (per-SERVICE; both replicas if you doubt one):
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor <phase-3 commit sha> <the stamp sha>   # must exit 0
-- The stamp is a STARTUP line and scrolls on a busy service; an empty grep
-- means "not in range", not "unstamped". Fallback binary probe, WITH BOTH
-- controls in the same breath (never `strings`; never a non-ASCII marker):
--   P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
--   kubectl -n ai-persona-system exec ${P#pod/} -- grep -aq "<full stamp sha>" /proc/1/exe && echo stamp-present
--   kubectl -n ai-persona-system exec ${P#pod/} -- grep -aq "7f3a9c2e8b1d4f6a0c5e9b3d7a1f8c4e2b6d0a9f" /proc/1/exe && echo CONTROL-BROKEN-random-hex-matches
-- (positive control: the stamp sha must match; negative control: the random
-- hex must NOT. ⚠ NEVER the all-zeros sha as the negative control — it is
-- git's null-sha constant and matches legitimately, measured 2026-08-15.)
--
-- THEN apply with exactly this command from the repo root:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--     -f - < docs/agent_docs/sql_for_agents/459_zip_deliverer_agent_HOLD.sql
-- ⚠ There is NO --record-only step: the runner REFUSES to record
-- UPPERCASE-suffixed sidecars (by design). The apply record is the lane
-- NOTES entry + this header. STATUS: NOT YET APPLIED.
-- Re-running after a successful apply is SAFE: both inserts are guarded by
-- WHERE NOT EXISTS and no-op on the second pass.
--
-- WHAT IT DOES (all inert until something dispatches the shim — there is NO
-- scheduled task; the ZIP is cut ON DEMAND, unlike the publish reconciler):
--   1. zip-deliverer — the spawned executor. Its workflow runs the
--      zip_deliverable action: list portfolio-sites/<domain>/, compose the
--      archive to a temp file (seekable body — B2 411s bare streams), verify
--      at the artefact (entry count == listing; index.html byte-equal;
--      remote size == local), upload to deliverables/<domain>/, presign.
--      MUST run spawned: it is on the spawner's storage allow-list
--      (isStorageEnabledAgent, same Phase 3 commit); the standing chassis
--      carries no B2 credentials (owner ruling 2026-08-08).
--      The step config maps size_alert_bytes and expiry_minutes as explicit
--      dot-paths because a field WITH a spec default can ONLY be overridden
--      by a Strategy-0 path (bugs_open/248 finding (b)); unsupplied, the
--      paths do not resolve and the defaults stand (512 MiB alert; 7-day
--      expiry).
--   2. zip-deliverable-dispatch — the orchestrator shim (422's proven
--      spawn→call shape): spawn zip-deliverer, call it with {domain [,
--      size_alert_bytes, expiry_minutes]}, complete. The spawn→call pair is
--      the estate's standing (and known-racy) handshake — a failed handshake
--      is re-dispatched by the caller, not retried by this workflow.
--
-- DISPATCH RECIPE (Phase 3 acceptance; the rerender_page_safe.sh publish
-- pattern — payload in the container COMMAND, PUBLISH_OK receipt, because
-- `kubectl run -i | kcat -P` drops ~4 in 5 publishes silently):
--   msg = {"action":"orchestrate","config":{"agent_type":"zip-deliverable-dispatch"},
--          "input_data":{"domain":"noted.co.uk"}}
--   → base64 → kubectl -n kafka run kcat-zip-$RANDOM --rm --restart=Never \
--       --image=edenhill/kcat:1.7.1 --attach=true --quiet --command -- sh -c \
--       "echo '<b64>' | base64 -d | kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
--        -t system.agent.generic.requests -H correlation_id=<uuid> \
--        -H orchestration_id=<uuid> -H request_id=<uuid> -H message_id=<uuid> \
--        -H orchestration_name=zip-cut -H step_name=start -H client_id=demo_client \
--        -H message_type=request -H action=orchestrate -H from_agent_type=user \
--        -H from_agent_id=cli -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"
--   No PUBLISH_OK means NOTHING was published — re-run. No dispatch within
--   ~300s of a chassis pod (re)start. Follow by correlation_id in
--   orchestration_states (a missing row is latency, not a drop — budget 30m).
--   For the size-alert demand control, add "size_alert_bytes":1 to input_data
--   and expect size_alert=true WITH a completed cut in the result.
--
-- ROLLBACK RECIPE: 459_zip_deliverer_agent_HOLD_ROLLBACK.sql (soft-deletes
-- both definitions; they are new rows, nothing is repurposed here).

BEGIN;

-- 1. The spawned executor -------------------------------------------------
INSERT INTO agent_definitions (type, display_name, description, category, agent_category, status, is_active, default_config)
SELECT
  'zip-deliverer',
  'ZIP Deliverer',
  'Cuts the ZIP ownership artefact for one site: lists portfolio-sites/<domain>/, composes the archive to a seekable temp file, verifies at the artefact (entry count, index.html bytes, remote size), uploads under deliverables/<domain>/ and returns a presigned URL. Alerts (never truncates) past size_alert_bytes. MUST run as a SPAWNED pod: on the spawner''s storage allow-list; the standing chassis carries no B2 credentials. Dispatch via zip-deliverable-dispatch with input_data {domain}. Register DGH-011.',
  'executor',
  'executor',
  'experimental',
  true,
  jsonb_build_object('workflow', jsonb_build_object(
    'start_step', 'zip',
    'timeout_seconds', 900,
    'steps', jsonb_build_object(
      'zip', jsonb_build_object(
        'action', 'zip_deliverable',
        'config', jsonb_build_object(
          'domain',           'input_data.domain',
          'size_alert_bytes', 'input_data.size_alert_bytes',
          'expiry_minutes',   'input_data.expiry_minutes'
        ),
        'next_step', 'complete',
        'description', 'Cut, verify, upload and presign the site ZIP',
        'output_field', 'zip_result'
      ),
      'complete', jsonb_build_object(
        'action', 'complete_workflow',
        'config', jsonb_build_object('output_fields', jsonb_build_array('zip_result')),
        'description', 'Return the ZIP result to the caller'
      )
    )
  ))
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions WHERE type = 'zip-deliverer' AND deleted_at IS NULL
);

-- 2. The dispatching shim -------------------------------------------------
INSERT INTO agent_definitions (type, display_name, description, category, agent_category, status, is_active, default_config)
SELECT
  'zip-deliverable-dispatch',
  'ZIP Deliverable Dispatch',
  'On-demand shim: spawns a zip-deliverer pod (the storage-credentialed environment) and calls it with one site''s {domain}. All archive logic lives in zip-deliverer''s zip_deliverable action; this workflow only crosses the credential boundary. NO scheduled task fires this — the ZIP is cut on demand (Phase 4''s handover, or a manual dispatch for acceptance). Register DGH-011.',
  'orchestrator',
  'coordinator',
  'experimental',
  true,
  jsonb_build_object('workflow', jsonb_build_object(
    'start_step', 'spawn_zipper',
    'processing_mode', 'orchestrator',
    'timeout_seconds', 1200,
    'steps', jsonb_build_object(
      'spawn_zipper', jsonb_build_object(
        'action', 'spawn_agent',
        'config', jsonb_build_object('role', 'zipper', 'agent_type', 'zip-deliverer'),
        'next_step', 'call_zipper',
        'description', 'Spawn the storage-credentialed zip pod',
        'output_field', 'zipper_spawned'
      ),
      'call_zipper', jsonb_build_object(
        'action', 'call_agent',
        'config', jsonb_build_object(
          'target_role', 'zipper',
          'input_mapping', jsonb_build_object(
            'domain',           'input_data.domain',
            'size_alert_bytes', 'input_data.size_alert_bytes',
            'expiry_minutes',   'input_data.expiry_minutes'
          ),
          'timeout_seconds', 900
        ),
        'next_step', 'complete',
        'description', 'One ZIP cut for the named site',
        'output_field', 'zip_result'
      ),
      'complete', jsonb_build_object(
        'action', 'complete_workflow',
        'config', jsonb_build_object('output_fields', jsonb_build_array('zip_result')),
        'description', 'Return the ZIP result (zip_key, presigned_url, sizes) to the dispatcher'
      )
    )
  ))
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions WHERE type = 'zip-deliverable-dispatch' AND deleted_at IS NULL
);

-- 3. Verify ---------------------------------------------------------------
DO $verify$
DECLARE
  n int;
  a text;
BEGIN
  SELECT default_config#>>'{workflow,steps,zip,action}' INTO a
    FROM agent_definitions
   WHERE type = 'zip-deliverer'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF a IS DISTINCT FROM 'zip_deliverable' THEN
    RAISE EXCEPTION '459 verify: zip-deliverer zip step action is %, want zip_deliverable', a;
  END IF;

  SELECT count(*) INTO n FROM jsonb_object_keys((
    SELECT default_config->'workflow'->'steps' FROM agent_definitions
     WHERE type = 'zip-deliverable-dispatch'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL));
  IF n <> 3 THEN
    RAISE EXCEPTION '459 verify: zip-deliverable-dispatch should have 3 steps, found %', n;
  END IF;

  -- No schedule must exist: the ZIP is on-demand by design. A row here means
  -- another session armed one — read its intent before proceeding.
  SELECT count(*) INTO n FROM scheduled_tasks WHERE target_agent_type IN ('zip-deliverer', 'zip-deliverable-dispatch');
  IF n <> 0 THEN
    RAISE EXCEPTION '459 verify: found % scheduled_tasks targeting the zip agents — this seed arms NO schedule', n;
  END IF;
END $verify$;

COMMIT;
