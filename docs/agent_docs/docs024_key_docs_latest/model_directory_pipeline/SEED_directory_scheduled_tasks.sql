-- SEED — model directory pipeline on a cadence (Phase B)
--
-- APPLY ONLY AFTER a chassis image carrying `refresh_directory_claims` AND
-- `verify_and_register_directory_claims` is deployed and pod-verified.
-- CLAUDE.md: "Image first, then seeds (a seed naming an unregistered action
-- fails at runtime)." Verify first, against the POD:
--
--   kubectl -n ai-persona-system exec <agent-chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c refresh_directory_claims'
--   kubectl -n ai-persona-system exec <agent-chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c verify_and_register_directory_claims'
--
-- Apply SEED_directory_researcher_agent.sql BEFORE this file — the discovery
-- task's inline workflow calls `directory-researcher` via spawn_agent, which
-- must already exist as an agent_definitions row.
--
-- Two tasks, two different jobs:
--
--   model-directory-discovery (weekly, 604800s) — finds NEW models. New-model
--   cadence is slow relative to news, so weekly is ample; the research_query
--   is deliberately broad ("recently released or updated") rather than naming
--   specific vendors, so genuinely new entrants are found, not just the
--   incumbents this seed happens to know about.
--
--   model-directory-freshness (daily, 86400s) — re-verifies EXISTING claims
--   via refresh_directory_claims. Deterministic, no LLM, and selects only
--   claims whose own staleness_days has actually elapsed — see
--   loadDueDirectoryClaims in directory_claims.go — so a daily tick does not
--   re-fetch every citation every day, only the ones due.

INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds
)
SELECT
    'model-directory-discovery',
    'Model directory pipeline: discover new/updated AI models via directory-researcher and register their cited claims.',
    604800,
    'directory-researcher',
    -- NOT a dedicated per-type topic: a custom agent type that runs on the
    -- shared agent-chassis (rather than its own microservice, unlike
    -- business-intel/vet-intel) is dispatched via the generic requests
    -- topic, with the real type carried in the message's config.agent_type
    -- (cmd/scheduler/main.go fireTrigger, line ~430: "agent_type":
    -- task.TargetAgentType, published to task.TargetTopic regardless).
    -- Verified against the live content-feed-refresh row, which targets
    -- target_agent_type='content-feed-trigger' on this SAME topic:
    --   SELECT target_agent_type, target_topic FROM scheduled_tasks
    --   WHERE name = 'content-feed-refresh';
    'system.agent.generic.requests',
    jsonb_build_object(
        'research_query',
        'AI models (text, image, voice, code, or embedding) released or significantly updated in the last 30 days, with pricing and specifications'
    ),
    'model-directory-discovery',
    1,
    true,
    600
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'model-directory-discovery');

INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds
)
SELECT
    'model-directory-freshness',
    'Model directory pipeline: re-verify is_current directory_claims whose staleness_days has elapsed; supersede on any status transition (found/citation_lost/fetch_error); raise stale_directory_claim on any flip away from found.',
    86400,
    'generic',
    'system.agent.generic.requests',
    jsonb_build_object(
        'config', jsonb_build_object(
            'agent_type', 'generic',
            'workflow', jsonb_build_object(
                'start_step', 'refresh_claims',
                'processing_mode', 'orchestrator',
                'timeout_seconds', 600,
                'steps', jsonb_build_object(
                    'refresh_claims', jsonb_build_object(
                        'action', 'refresh_directory_claims',
                        'config', jsonb_build_object(),
                        'description', 'Sweep every due claim across every kind',
                        'next_step', 'complete',
                        'output_field', 'refresh_result'
                    ),
                    'complete', jsonb_build_object(
                        'action', 'complete_workflow',
                        'config', jsonb_build_object('output_fields', jsonb_build_array('refresh_result'))
                    )
                )
            )
        )
    ),
    'model-directory-freshness',
    1,
    true,
    600
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'model-directory-freshness');

-- ── Post-apply verification ────────────────────────────────────────────────
--
-- 1. Both rows exist and are enabled:
--    SELECT name, target_agent_type, interval_seconds, enabled
--    FROM scheduled_tasks WHERE name LIKE 'model-directory-%';
--
-- 2. Bring a task forward for a smoke test rather than waiting out the
--    interval: UPDATE scheduled_tasks SET last_triggered_at = NULL
--    WHERE name = 'model-directory-freshness';
--
-- 3. After discovery fires at least once, claims exist:
--    SELECT de.slug, dc.field, dc.status FROM directory_claims dc
--    JOIN directory_entities de ON de.id = dc.entity_id WHERE dc.is_current;
--
-- 4. After freshness fires, verified_at moves and (on the failing branch —
--    hand-corrupt one claim's citation.quote first) a NEW is_current row
--    appears with status='citation_lost', the old row superseded:
--    SELECT status, is_current, superseded_at, verified_at FROM directory_claims
--    WHERE entity_id = '<id>' AND field = '<field>' ORDER BY created_at;
--
-- 5. Any status flip raised for a human (empty is the healthy steady state):
--    SELECT summary, status, spec->'flipped' FROM site_work_items
--    WHERE item_type = 'stale_directory_claim' ORDER BY created_at DESC;
--
-- ── Rollback ────────────────────────────────────────────────────────────────
--    UPDATE scheduled_tasks SET enabled = false WHERE name LIKE 'model-directory-%';
-- Disabling is sufficient and lossless: every claim revision is superseded,
-- never overwritten, so history survives in directory_claims regardless.
