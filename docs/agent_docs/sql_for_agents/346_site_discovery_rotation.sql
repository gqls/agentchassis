-- 346_site_discovery_rotation.sql
--
-- bugs_open/230: site discovery has no recurring driver — every scheduled entry
-- targeting the three discovery agents is a disabled one-off, so a site stops
-- being examined the moment a human stops looking at it.
--
-- This migration gives detection a clock, and nothing else:
--
--   1. `site_discovery_rotation` — one stamp per (site, agent): when the site
--      was last SELECTED for examination by that agent. Selection, not
--      completion, deliberately: a site whose run fails must not pin the
--      rotation head and starve the fleet (the bugs_open/048 shape the
--      scheduler's own comments warn about). Whether selected sites actually
--      RAN is the site-discovery-staleness-check CronJob's question, answered
--      daily against the 24h orchestration_states retention window.
--
--   2. Three scheduled_tasks rows, one per discovery agent. Each hourly tick,
--      the pre_query picks the least-recently-selected active/deployed site
--      whose stamp is older than 7 days (or absent), skips it THIS TICK ONLY if
--      the site has a claimed build item in flight (the skip costs the site
--      nothing — its stamp stays oldest, so it is re-picked as soon as it is
--      free), stamps it, and fires the agent with {site_id, domain} — the same
--      envelope the proven oneshot rows used.
--
-- Why not re-enable improvement-sweep instead: that row is the whole
-- detect→triage→fix loop, whose re-enable is an owner decision recorded as
-- PENDING in bugs_open/083; its site selection starves (register IMP-010,
-- ORDER BY sites.updated_at — nothing it does advances its own sort key); and
-- its <50-open-items cap permanently excludes exactly the most-worked sites
-- (measured 2026-08-09: webdesign.co.uk 85, dartsonline.com 79). This rotation
-- is detection-only — findings land wherever each check already puts them, and
-- insertion is dedup-guarded (idx_swi_dedup), so re-examination is idempotent
-- for open findings. Observe-only is a designed mode, not an improvisation
-- (register improvement-loop.md: findings insert status='detected' so checks
-- can run while the triager stays disabled).
--
-- ROLLBACK RECIPE (also in 346_site_discovery_rotation_ROLLBACK.sql):
--   DELETE FROM scheduled_tasks WHERE name LIKE 'site-discovery-rotation-%';
--   DROP TABLE IF EXISTS site_discovery_rotation;
-- Pause without rollback:
--   UPDATE scheduled_tasks SET enabled=false WHERE name LIKE 'site-discovery-rotation-%';

BEGIN;

CREATE TABLE site_discovery_rotation (
    site_id          uuid        NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    agent_type       text        NOT NULL,
    last_selected_at timestamptz NOT NULL,
    PRIMARY KEY (site_id, agent_type)
);

COMMENT ON TABLE site_discovery_rotation IS
    'bugs_open/230: fair-rotation stamps for the site-discovery scheduled tasks. '
    'last_selected_at is when the scheduler last PICKED the site for this agent '
    '(not proof the run completed — the staleness CronJob checks that). '
    'Written by the site-discovery-rotation-* pre_queries; safe to TRUNCATE '
    '(every site becomes immediately due, once each, in site_id order).';

INSERT INTO scheduled_tasks
    (name, description, target_topic, target_agent_type,
     interval_seconds, concurrency_group, max_concurrent, timeout_seconds,
     enabled, fire_message, input_data, pre_query)
SELECT
    'site-discovery-rotation-' || short_name,
    'bugs_open/230: recurring fair-rotation driver for ' || agent_type ||
    '. Each hourly tick examines the least-recently-selected active/deployed site '
    'whose stamp is older than 7 days; detection only (no triage, no fixes). '
    'Period/pause knobs: bugfix_230_discovery_driver/RUNBOOK_discovery_driver.md.',
    'system.agent.scheduled.requests',
    agent_type,
    3600,
    'site-discovery',
    1,
    600,
    true,
    true,
    '{}'::jsonb,
    'WITH due AS (' ||
    '  SELECT s.id AS sid, s.domain' ||
    '  FROM sites s' ||
    '  LEFT JOIN site_discovery_rotation r' ||
    '    ON r.site_id = s.id AND r.agent_type = ''' || agent_type || '''' ||
    '  WHERE s.status IN (''active'', ''deployed'')' ||
    '    AND COALESCE(r.last_selected_at, ''-infinity''::timestamptz) < now() - interval ''7 days''' ||
    '    AND NOT EXISTS (' ||
    '      SELECT 1 FROM site_work_items wi' ||
    '      WHERE wi.site_id = s.id AND wi.status = ''claimed'' AND wi.pipeline = ''build'')' ||
    '  ORDER BY r.last_selected_at ASC NULLS FIRST, s.id' ||
    '  LIMIT 1' ||
    '), stamped AS (' ||
    '  INSERT INTO site_discovery_rotation (site_id, agent_type, last_selected_at)' ||
    '  SELECT sid, ''' || agent_type || ''', now() FROM due' ||
    '  ON CONFLICT (site_id, agent_type) DO UPDATE SET last_selected_at = EXCLUDED.last_selected_at' ||
    ') ' ||
    'SELECT sid::text AS site_id, domain FROM due'
FROM (VALUES
    ('quality',      'quality-discovery-agent'),
    ('design',       'design-discovery-agent'),
    ('completeness', 'completeness-discovery-agent')
) AS agents(short_name, agent_type);

-- One doc_notes row establishing the new pipeline as a queryable subject, so the
-- next fix touching this mechanism finds a subject_key-keyed record instead of
-- rediscovering it from scattered bug files (council 2281fc48, tooling_provenance
-- seat, medium).
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
VALUES ('pipeline', 'site-discovery-rotation',
        'site-discovery-rotation: fair recurring driver for the three site-discovery '
        'agents (bugs_open/230, migration 346, register SCH-025). One stamp table '
        'site_discovery_rotation (site_id, agent_type, last_selected_at = SELECTION, '
        'not completion) + three hourly scheduled_tasks (site-discovery-rotation-'
        '{quality,design,completeness}), each picking the least-recently-selected '
        'active/deployed site older than the 7-day period, busy-skip deferring (never '
        'excluding) sites with a claimed build item. Observe-only: findings land where '
        'each check already writes them; triage/drain stays bugs_open/083''s pending '
        'owner decision. Watched daily by the site-discovery-staleness-check CronJob '
        '(doc_notes subject_key site-discovery-staleness). Plan/notes/runbook: '
        'docs024_key_docs_latest/bugfix_230_discovery_driver/. Council-Reviewed: '
        '2281fc48-f0c5-4842-88c7-8391d0098944.',
        '["site-discovery-rotation"]'::jsonb, 'migration-346');

-- Guard: assert the exact post-conditions; any failure rolls the whole file back.
DO $$
DECLARE
    task_count integer;
    bad_preq   integer;
    live_agents integer;
BEGIN
    IF to_regclass('public.site_discovery_rotation') IS NULL THEN
        RAISE EXCEPTION 'guard: site_discovery_rotation table missing';
    END IF;

    -- The three target_agent_type strings are free text; a typo would make a
    -- rotation task silently never fire its agent (council 2281fc48, editquality
    -- seat, medium). Assert each names a live, active, non-snapshot agent.
    SELECT count(*) INTO live_agents
    FROM scheduled_tasks st
    WHERE st.name LIKE 'site-discovery-rotation-%'
      AND EXISTS (SELECT 1 FROM agent_definitions ad
                  WHERE ad.type = st.target_agent_type
                    AND ad.is_active
                    AND COALESCE(ad.is_snapshot, false) = false
                    AND ad.deleted_at IS NULL);
    IF live_agents <> 3 THEN
        RAISE EXCEPTION 'guard: only % of 3 rotation tasks name a live active agent_definitions.type', live_agents;
    END IF;

    SELECT count(*) INTO task_count
    FROM scheduled_tasks
    WHERE name LIKE 'site-discovery-rotation-%'
      AND enabled AND fire_message
      AND target_topic = 'system.agent.scheduled.requests'
      AND concurrency_group = 'site-discovery';
    IF task_count <> 3 THEN
        RAISE EXCEPTION 'guard: expected 3 enabled site-discovery-rotation tasks, found %', task_count;
    END IF;

    -- each pre_query must name its own agent (the stamp key) and the rotation table
    SELECT count(*) INTO bad_preq
    FROM scheduled_tasks
    WHERE name LIKE 'site-discovery-rotation-%'
      AND (pre_query IS NULL
           OR position(target_agent_type IN pre_query) = 0
           OR position('site_discovery_rotation' IN pre_query) = 0);
    IF bad_preq <> 0 THEN
        RAISE EXCEPTION 'guard: % rotation task(s) whose pre_query does not reference its agent and the stamp table', bad_preq;
    END IF;
END $$;

COMMIT;
