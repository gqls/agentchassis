-- 372_site_availability_driver.sql
--
-- bugs_open/236 (522 half): a fully built site served HTTP 522 to every visitor
-- indefinitely and nothing noticed. The Go half (check_site_unreachable.go,
-- committed alongside this file) probes https://<domain>/ the way a visitor
-- would and files/self-clears a `site_unreachable` work item. This migration is
-- the driver: a lightweight agent that runs ONLY that check, and a rotation
-- task that gives every active/deployed site a probe every ~4 hours.
--
-- HOLD RELEASED 2026-08-10 22:02Z. The gate this file described was satisfied
-- before the rename: chassis v1.0.1283 (pods 696d88b4c7-95mgb / -wnbs8, both
-- started 21:43Z) greps `site_unreachable` 7 and `+site_unreachable` 1 on BOTH
-- replicas, invented negative control `site_unreachabl3` 0, pipeline control
-- `asset_reference_404` 13 — one exec each. The ordering constraint below was
-- real and is now discharged; it is kept because it applies again to anyone
-- re-applying this against an older image.
--
-- ORIGINAL HOLD NOTE: run_discovery_checks HARD-FAILS on a check name the binary
-- does not register (bugs_open/149 B4), so applying this against an old binary
-- makes every availability run fail loudly until the roll. The code half without
-- this config is inert and harmless.
--
-- Register entry IMP-053 (improvement-loop.md) ships in the same commit as the
-- Go check, per the seam-registration rule (CLAUDE.md 2026-07-28/29).
--
-- Design (bugfix_236_site_availability/PLAN, decisions D1-D5):
--  * agent: 3-step clone of quality-discovery-agent with checks=[site_unreachable].
--    No LLM steps, no spawned children (so: no per-job topics — bugs_open/240).
--  * driver: the 230 lane's rotation pre_query verbatim, with agent_type
--    'availability-discovery-agent' and cooldown '4 hours' (theirs: '7 days').
--    Reuses their site_discovery_rotation table — a new agent_type value, no
--    schema change; their staleness watchdog keys on the three content agents
--    and is unaffected.
--  * cadence arithmetic (2026-08-10): 21 eligible sites / 4h = 5.25 dispatch/h
--    needed; tick 300s = 12/h capacity → 2.3x headroom. Detection latency for a
--    total outage drops from unbounded to <= ~4-8h.
--  * own concurrency_group 'site-availability': the scheduler's in-memory
--    inFlight head-of-queue coupling (bugs_open/048) cannot then touch the
--    content rotations.
--
-- ROLLBACK RECIPE (also in 372_site_availability_driver_ROLLBACK.sql):
--   UPDATE scheduled_tasks SET enabled=false WHERE name='site-discovery-rotation-availability';
--   UPDATE agent_definitions SET is_active=false, deleted_at=now()
--    WHERE type='availability-discovery-agent' AND deleted_at IS NULL;
--   DELETE FROM site_discovery_rotation WHERE agent_type='availability-discovery-agent';

BEGIN;

INSERT INTO agent_definitions (type, display_name, description, category, status, is_active, default_config)
SELECT
  'availability-discovery-agent',
  'Availability Discovery (does the site serve?)',
  'Runs exactly one discovery check, site_unreachable (bugs_open/236): fetch https://<domain>/ as a visitor would; file a high-severity alert item when the site does not serve, self-clear it when it does. Dispatch with input_data {site_id, domain}. Deliberately owns no LLM steps and spawns nothing.',
  'specialist',
  'experimental',
  true,
  jsonb_build_object('workflow', jsonb_build_object(
    'start_step', 'ensure_site_record',
    'steps', jsonb_build_object(
      'ensure_site_record', jsonb_build_object(
        'action', 'ensure_site_record',
        'config', jsonb_build_object('input_fields', jsonb_build_array('site_id', 'domain')),
        'next_step', 'run_checks',
        'description', 'Load site record from domain or site_id',
        'output_field', 'site_record'
      ),
      'run_checks', jsonb_build_object(
        'action', 'run_discovery_checks',
        'config', jsonb_build_object(
          'checks', jsonb_build_array('site_unreachable'),
          'site_id', 'site_record.site_id',
          'check_pipeline', 'build'
        ),
        'next_step', 'complete',
        'description', 'Probe the public apex and file/clear the site_unreachable alert',
        'output_field', 'discovery_result'
      ),
      'complete', jsonb_build_object(
        'action', 'complete_workflow',
        'config', jsonb_build_object('output_fields', jsonb_build_array('discovery_result')),
        'description', 'Availability probe complete'
      )
    )
  ))
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions
  WHERE type = 'availability-discovery-agent' AND deleted_at IS NULL
);

INSERT INTO scheduled_tasks
  (name, description, interval_seconds, target_agent_type, target_topic,
   input_data, concurrency_group, max_concurrent, timeout_seconds, fire_message, enabled, pre_query)
SELECT
  'site-discovery-rotation-availability',
  'bugs_open/236: fair-rotation availability probe. Picks the least-recently-probed active/deployed site whose stamp is older than 4 hours, stamps it, dispatches availability-discovery-agent (checks=[site_unreachable]). Same pattern as the three site-discovery-rotation-* content tasks (bugs_open/230), faster clock because an outage is not a content defect.',
  300,
  'availability-discovery-agent',
  'system.agent.scheduled.requests',
  '{}'::jsonb,
  'site-availability',
  1,
  600,
  true,
  true,
  $PRE$WITH due AS (
  SELECT s.id AS sid, s.domain
  FROM sites s
  LEFT JOIN site_discovery_rotation r
    ON r.site_id = s.id AND r.agent_type = 'availability-discovery-agent'
  WHERE s.status IN ('active', 'deployed')
    AND COALESCE(r.last_selected_at, '-infinity'::timestamptz) < now() - interval '4 hours'
    AND NOT EXISTS (
      SELECT 1 FROM site_work_items wi
      WHERE wi.site_id = s.id AND wi.status = 'claimed' AND wi.pipeline = 'build')
  ORDER BY r.last_selected_at ASC NULLS FIRST, s.id
  LIMIT 1), stamped AS (
  INSERT INTO site_discovery_rotation (site_id, agent_type, last_selected_at)
  SELECT sid, 'availability-discovery-agent', now() FROM due
  ON CONFLICT (site_id, agent_type) DO UPDATE SET last_selected_at = EXCLUDED.last_selected_at)
SELECT sid::text AS site_id, domain FROM due$PRE$
WHERE NOT EXISTS (
  SELECT 1 FROM scheduled_tasks WHERE name = 'site-discovery-rotation-availability'
);

-- Verify as DO/RAISE, not bare SELECTs: ON_ERROR_STOP ignores a non-empty
-- result set, so only an exception can stop the COMMIT (WRONG_CALLS/mig-verify
-- lesson, memory 'a migration verify block of SELECTs cannot stop the COMMIT').
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'availability-discovery-agent' AND is_active AND deleted_at IS NULL
     AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' ? 'site_unreachable';
  IF n <> 1 THEN
    RAISE EXCEPTION 'availability-discovery-agent row wrong (count=%, or checks array lacks site_unreachable)', n;
  END IF;
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'site-discovery-rotation-availability' AND enabled
     AND pre_query LIKE '%availability-discovery-agent%'
     AND pre_query LIKE '%interval ''4 hours''%';
  IF n <> 1 THEN
    RAISE EXCEPTION 'site-discovery-rotation-availability task wrong (count=%)', n;
  END IF;
END $$;

COMMIT;
