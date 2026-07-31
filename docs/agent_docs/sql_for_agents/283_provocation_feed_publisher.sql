-- 283_provocation_feed_publisher.sql — the agent and the schedule that finally
-- make the daily provocation rotate.
--
-- DEPENDS ON: 282_provocation_pool.sql (the pool), and a chassis image carrying
-- the `render_provocation_feed` action (commit 572ae8dc6).
--
-- ============================================================================
-- THE SCHEDULED TASK IS SEEDED **DISABLED**. THAT IS DELIBERATE.
-- ============================================================================
-- Go changes are inert until an image is rebuilt and rolled; DB config is live
-- immediately. So a scheduled row enabled here would start firing at a chassis
-- that does not yet know the action, and "a seed naming an unregistered action
-- fails at runtime" — a failure that looks like a broken feature rather than a
-- missing deploy.
--
-- Seeding it disabled makes this migration safe to apply at any time, and turns
-- the ordering constraint into a single reversible flip once the image is proven
-- ON THE POD:
--
--   kubectl exec -n ai-persona-system <chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c render_provocation_feed'
--
-- A roll is NOT evidence the action shipped: the image may predate the commit and
-- carries no provenance. Grep a string the change ADDED, with a positive control
-- in the same exec. Then:
--
--   UPDATE scheduled_tasks SET enabled = true WHERE name = 'provocation-feed-refresh';
--
-- ============================================================================
-- WHY DAILY, AND WHY THAT IS THE ONLY OPTION
-- ============================================================================
-- The cheap design was to publish the whole schedule once and let both readers
-- select by date — no job at all, and rotation that cannot silently stop. The
-- owner's seal ruling (2026-07-31) forecloses it: publishing ahead would put every
-- future provocation in a world-readable file when even today's is meant to be
-- hidden until you step into the round. The seal is what makes this job mandatory.
--
-- The interval is 6h rather than 24h on purpose. The action is idempotent and
-- SKIPS the commit when nothing has changed, so extra runs are free; what they buy
-- is that a failed or missed run is retried within hours instead of leaving the
-- site a day stale. Rotation still happens exactly once per day, because that is a
-- property of the pool's dates, not of how often this fires.

BEGIN;

DO $guard$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables
                   WHERE table_schema = 'public' AND table_name = 'provocations') THEN
        RAISE EXCEPTION '283: the provocations table does not exist — apply 282 first';
    END IF;
    IF EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'provocation-feed-refresh') THEN
        RAISE EXCEPTION '283: provocation-feed-refresh already exists — migration already applied';
    END IF;
END
$guard$;

-- ---------------------------------------------------------------------------
-- The agent. One action step and a completion, the same shape as
-- directory-json-exporter, which does the same job for a different artefact.
-- ---------------------------------------------------------------------------

INSERT INTO agent_definitions
    (type, display_name, category, description, is_active, status,
     topics, default_config)
VALUES (
    'provocation-feed-publisher',
    'Provocation Feed Publisher',
    'content',
    'Selects today''s provocation for a site from the provocations pool by publish '
    'date, builds the feed (today/seal/sample/arena/archive) and commits it to the '
    'site repo. Fails closed: an empty pool, a failed seal or engine invariant, or '
    'a shrinking archive all leave the served file untouched. Skips the commit when '
    'nothing has changed, so the file''s history stays an honest record of rotation.',
    true,
    'active',
    '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
    jsonb_build_object(
        'processing_mode', 'task',
        'workflow', jsonb_build_object(
            'start_step', 'publish_feed',
            'steps', jsonb_build_object(
                'publish_feed', jsonb_build_object(
                    'action', 'render_provocation_feed',
                    'config', '{}'::jsonb,
                    'next_step', 'complete',
                    'description', 'Select today''s provocation, build the feed, verify it, commit it'
                ),
                'complete', jsonb_build_object(
                    'action', 'complete_workflow',
                    'description', 'Provocation feed published or unchanged'
                )
            )
        )
    )
);

-- ---------------------------------------------------------------------------
-- The schedule. DISABLED — see the header.
--
-- Domain is explicit and has no default, matching the action, which refuses to
-- publish without one. A second site becomes a second row, not a code change.
-- ---------------------------------------------------------------------------

INSERT INTO scheduled_tasks
    (name, description, interval_seconds, target_agent_type, target_topic,
     input_data, concurrency_group, max_concurrent, enabled, timeout_seconds)
VALUES (
    'provocation-feed-refresh',
    'Rebuild and republish vonc.com''s daily provocation feed from the provocations '
    'pool. Idempotent: commits only when the selected provocation or the archive '
    'actually changes. DISABLED until a chassis image carrying render_provocation_feed '
    'is proven on the pod.',
    21600,
    'provocation-feed-publisher',
    'system.agent.scheduled.requests',
    jsonb_build_object(
        'domain', 'vonc.com',
        'repo_name', 'sites',
        'data_path', 'data',
        'filename', 'provocations.json',
        'commit_message_prefix', 'Update daily provocation',
        'task_name', 'provocation-feed-refresh'
    ),
    'vonc-com-provocations',
    1,
    false,
    600
);

-- Assert what was actually created, and assert the safety property rather than
-- assuming it: a row that arrived enabled would fire at a chassis that does not
-- know the action yet.
DO $verify$
DECLARE
    n_agent int;
    is_on   boolean;
BEGIN
    SELECT count(*) INTO n_agent FROM agent_definitions
        WHERE type = 'provocation-feed-publisher' AND is_active
          AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n_agent <> 1 THEN
        RAISE EXCEPTION '283: expected exactly 1 active provocation-feed-publisher, found %', n_agent;
    END IF;

    SELECT enabled INTO is_on FROM scheduled_tasks WHERE name = 'provocation-feed-refresh';
    IF is_on IS DISTINCT FROM false THEN
        RAISE EXCEPTION '283: provocation-feed-refresh must be seeded DISABLED (enabled=%)', is_on;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'provocation-feed-publisher'
          AND default_config #>> '{workflow,steps,publish_feed,action}' = 'render_provocation_feed'
    ) THEN
        RAISE EXCEPTION '283: the publish step does not name render_provocation_feed';
    END IF;
END
$verify$;

COMMIT;
