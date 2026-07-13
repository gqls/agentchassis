-- ============================================================================
-- 0NN_create_intent_events.sql — P4 collection target + scheduled collector
-- (numbering: operator's next free migration number)
-- ============================================================================
-- Off-box collection: a scheduled agent pulls each VM-hosted backend site's
-- key-gated GET /events (NDJSON) and upserts new lines here. Ranking is then a
-- query over this table. Idempotency is structural: engine_event_id (the
-- ev_... the engine assigns) is UNIQUE, so re-pulling overlapping windows is a
-- no-op via ON CONFLICT DO NOTHING — which lets the collector use a SAFE
-- overlapping `since` without risking duplicates.
--
-- Checkpoint needs NO extra storage: the collector's next `since` is
--   SELECT max(event_created_at) FROM intent_events WHERE host = $1
-- (minus a small safety margin). Strictly-after + the unique key make this
-- exactly-once in effect.
--
-- Schema basis: checked against live \d sites (has deploy_config jsonb,
-- domain unique) and \d scheduled_tasks (interval_seconds, target_agent_type,
-- target_topic, pre_query, enabled, concurrency_group, max_concurrent,
-- timeout_seconds, fire_message). Pattern for the task row mirrors the live
-- improvement-sweep (per-row dispatch via pre_query) and the thunder-training-
-- monitor convention of INSERTING DISABLED until the action is deployed.
-- ============================================================================

CREATE TABLE IF NOT EXISTS intent_events (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    engine_event_id  text NOT NULL,                       -- ev_… from the engine; idempotency key
    site_id          uuid REFERENCES sites(id) ON DELETE CASCADE,  -- resolved host→site; NULL if host not a known site yet
    host             text NOT NULL,
    kind             text NOT NULL,
    value            text NOT NULL,                       -- the captured term (already sanitised at the engine)
    ref_host         text NOT NULL DEFAULT '',            -- external referer host ("" if direct/same-site)
    landing_query    text NOT NULL DEFAULT '',            -- inbound ?q=/?utm=… that survived to the form page
    country          text NOT NULL DEFAULT '',            -- coarse CDN country, or ""
    event_created_at timestamptz NOT NULL,                -- when the visitor submitted (engine clock)
    collected_at     timestamptz NOT NULL DEFAULT now(),  -- when we pulled it
    CONSTRAINT chk_intent_kind      CHECK (kind IN ('search', 'categories', 'freetext')),
    CONSTRAINT chk_intent_value_len CHECK (char_length(value) <= 1000)
);

-- Idempotency + the indexes the collector and ranking need.
CREATE UNIQUE INDEX IF NOT EXISTS uq_intent_events_engine_id
    ON intent_events (engine_event_id);
CREATE INDEX IF NOT EXISTS idx_intent_events_host_time
    ON intent_events (host, event_created_at DESC);
CREATE INDEX IF NOT EXISTS idx_intent_events_site
    ON intent_events (site_id) WHERE site_id IS NOT NULL;

-- ---------------------------------------------------------------------------
-- Scheduled collector — INSERTED DISABLED (enable after the intent-collector
-- action is deployed AND each backend site carries deploy_config.engine.*).
-- pre_query returns one row per VM-hosted site; dispatch fans out per row
-- (improvement-sweep pattern), each invocation collecting that site.
-- ---------------------------------------------------------------------------
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    concurrency_group, max_concurrent, pre_query, enabled, timeout_seconds, fire_message
) VALUES (
    'intent-collection',
    'Every 10 min: pull new intent events from each VM-hosted backend site''s /events endpoint into intent_events. Gated to sites with deploy_config target=vm. INSERTED DISABLED — enable after the intent-collector action is deployed and per-site deploy_config.engine.{base_url,stats_key} is populated.',
    600,
    'intent-collector',
    'system.agent.generic.requests',
    'dispatch',
    2,
    $q$SELECT s.id::text AS site_id, s.domain
       FROM sites s
       WHERE s.status IN ('active', 'deployed')
         AND s.deploy_config->>'target' = 'vm'$q$,
    false,
    300,
    true
)
ON CONFLICT (name) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Per-site onboarding (the designation UPDATE, extended for collection):
-- stores the box base URL + the engine's INTERNAL_API_KEY so the collector can
-- authenticate to /events. The stats key is a read-only capture-export key for
-- anonymised intent data on that one box — low sensitivity — and lives in
-- deploy_config alongside the other VM specifics. (If a stronger posture is
-- wanted later, move it to a secrets table; the action reads one accessor
-- either way.)
--   UPDATE sites SET
--     github_repo  = 'vm-sites',
--     deploy_config = COALESCE(deploy_config, '{}'::jsonb) || jsonb_build_object(
--       'target', 'vm',
--       'capabilities', jsonb_build_array('backend'),
--       'engine', jsonb_build_object(
--         'base_url',  'https://relojistas.com',
--         'stats_key', '<the INTERNAL_API_KEY from the box env>'
--       ))
--   WHERE domain = 'relojistas.com';
-- ---------------------------------------------------------------------------
