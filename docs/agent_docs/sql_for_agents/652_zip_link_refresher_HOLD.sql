-- 652 (_HOLD): the ZIP link REFRESHER — the half DGH-018 named as unbuilt.
--
-- WHAT IT CLOSES: a /d/<token> link carries a stored presign capped at 7 days
-- by SigV4, while the token lives LiveLinkWindow (42 days served). Without this,
-- every download link quietly enters the "being refreshed" state a week after
-- its last stamp, and refresh is a manual operator act armed by ZIP_LINK_STALE
-- rows. With it, live links are re-stamped ~48h BEFORE their presign dies, so a
-- customer should never see the stale page in normal operation.
--
-- ⚠ _HOLD, image-before-seeds: names `refresh_zip_link`, which exists only from
-- the roll carrying it (same roll as `send_delivery_email`). Apply AFTER:
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor <the commit adding refresh_zip_link> <the stamp>
--
-- SHAPE: the scheduler's pre_query selects (site_id, domain) for every site
-- holding a LIVE zip_download token whose stored presign dies within 48h; each
-- row dispatches one zip-link-refresher run: the proven spawn->call pair from
-- 459 (zip-deliverer is the ONLY presign minter, bugs_open/245) followed by
-- refresh_zip_link, which re-stamps stored_url on that site's live tokens. The
-- action's WHERE refuses revoked and expired tokens — a refresh can never
-- resurrect a killed link or extend a closed window (mutation-proven).
--
-- ⚠ INHERITED: the spawn->call handshake fails ~half the time fleet-wide
-- (MEMORY spawn-call-handshake-races): a FAILED refresher run self-heals at the
-- next 6h tick (the pre_query re-selects anything still unstamped) — never
-- cancel the failing row pre-diagnosis, and do not read one FAILED run as the
-- refresher being broken. The honest health check is the OUTCOME:
--   SELECT count(*) FROM customer_access_tokens
--    WHERE purpose='zip_download' AND revoked_at IS NULL AND expires_at > now()
--      AND stored_url_expires_at < now();          -- want 0 in steady state
-- and the customer-facing backstop stays armed either way: a stale hit renders
-- the honest page and files a ZIP_LINK_STALE row.

BEGIN;

INSERT INTO agent_definitions (type, display_name, description, category, agent_category, status, is_active, input_contract, default_config)
VALUES (
  'zip-link-refresher',
  'ZIP Link Refresher',
  'Re-stamps the stored presign on one site''s live zip_download tokens before the 7-day SigV4 ceiling kills it: spawns zip-deliverer (the only presign minter, bugs_open/245), calls it for a fresh cut, then refresh_zip_link writes the URL onto the live tokens (revoked and expired tokens untouchable by the action''s WHERE). Dispatched by the zip-link-refresh scheduled task ~48h ahead of expiry. Register DGH-018.',
  'executor',
  'executor',
  'active',
  true,
  jsonb_build_object(
    'required', jsonb_build_array('site_id', 'domain'),
    'notes', 'Both cross the spawn->call boundary via input_data; the pre_query supplies them.'
  ),
  jsonb_build_object('workflow', jsonb_build_object(
    'start_step', 'spawn_zipper',
    'processing_mode', 'orchestrator',
    'timeout_seconds', 1200,
    'steps', jsonb_build_object(
      'spawn_zipper', jsonb_build_object(
        'action', 'spawn_agent',
        'config', jsonb_build_object('role', 'zipper', 'agent_type', 'zip-deliverer'),
        'next_step', 'call_zipper',
        'description', 'Spawn the storage-credentialed zip pod (459''s proven pair)',
        'output_field', 'zipper_spawned'
      ),
      'call_zipper', jsonb_build_object(
        'action', 'call_agent',
        'config', jsonb_build_object(
          'target_role', 'zipper',
          'input_mapping', jsonb_build_object('domain', 'input_data.domain'),
          'timeout_seconds', 900
        ),
        'next_step', 'refresh',
        'description', 'One fresh ZIP cut + presign for the site',
        'output_field', 'zip_result'
      ),
      'refresh', jsonb_build_object(
        'action', 'refresh_zip_link',
        'config', jsonb_build_object(
          'site_id',        'input_data.site_id',
          'presigned_url',  'zip_result.presigned_url',
          'expiry_minutes', 'zip_result.expiry_minutes'
        ),
        'next_step', 'complete',
        'description', 'Write the fresh presign onto the site''s live zip_download tokens',
        'output_field', 'refresh_result'
      ),
      'complete', jsonb_build_object(
        'action', 'complete_workflow',
        'config', jsonb_build_object('output_fields', jsonb_build_array('refresh_result')),
        'description', 'Report how many tokens were re-stamped'
      )
    )
  ))
)
ON CONFLICT (type) DO NOTHING;

-- The schedule: every 6h, refresh anything dying within 48h. Zero live tokens
-- (today's state) -> pre_query returns nothing -> no dispatch, no cost.
INSERT INTO scheduled_tasks (name, description, interval_seconds, target_agent_type, target_topic, concurrency_group, max_concurrent, pre_query, enabled)
SELECT
  'zip-link-refresh',
  'Re-stamp stored presigns on live zip_download tokens ~48h before the 7-day SigV4 ceiling kills them (DGH-018''s refresher half). One dispatch per site holding a due token; the action''s WHERE cannot touch revoked or expired tokens.',
  21600,
  'zip-link-refresher',
  'system.agent.generic.requests',
  'zip-link-refresh',
  1,
  $Q$SELECT DISTINCT t.site_id::text AS site_id, s.domain
       FROM customer_access_tokens t
       JOIN sites s ON s.id = t.site_id
      WHERE t.purpose = 'zip_download'
        AND t.revoked_at IS NULL
        AND t.expires_at > now()
        AND (t.stored_url_expires_at IS NULL
             OR t.stored_url_expires_at < now() + interval '48 hours')$Q$,
  true
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'zip-link-refresh');

-- Verify: agent active; schedule present, enabled, pointed at the agent; and
-- the pre_query still carries the three token predicates (dropping one here
-- would out-run the action's own WHERE for selection, wasting dispatches or
-- missing NULL-stamped rows).
DO $$
DECLARE n int; q text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'zip-link-refresher' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '652 verify failed: zip-link-refresher not active (found %)', n; END IF;

  SELECT pre_query INTO q FROM scheduled_tasks
   WHERE name = 'zip-link-refresh' AND enabled AND target_agent_type = 'zip-link-refresher';
  IF q IS NULL THEN RAISE EXCEPTION '652 verify failed: schedule missing/disabled/mistargeted'; END IF;
  IF q NOT LIKE '%zip_download%' OR q NOT LIKE '%revoked_at IS NULL%' OR q NOT LIKE '%stored_url_expires_at%' THEN
    RAISE EXCEPTION '652 verify failed: the pre_query lost a protective predicate';
  END IF;
END $$;

COMMIT;
