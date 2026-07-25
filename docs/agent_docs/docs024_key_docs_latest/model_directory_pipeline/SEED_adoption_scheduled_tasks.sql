-- SEED — adoption tracker scheduled tasks (Phase E)
--
-- NO IMAGE ROLL REQUIRED: adoption-researcher (SEED_adoption_researcher_agent.sql,
-- applied 2026-07-25) uses only actions that already ship, and the freshness
-- sweep below calls refresh_directory_claims with a kind filter — a parameter
-- that action has accepted since Phase B (directory_claims.go:508,
-- "AND ($1 = '' OR de.kind = $1)").
--
-- WHY A SCHEDULED TASK AND NOT A MANUAL DISPATCH: on 2026-07-25 five direct
-- kcat `orchestrate` dispatches to system.agent.generic.requests produced zero
-- orchestration rows while the chassis was healthy and consuming that topic
-- from other producers (RUNBOOK, "Direct kcat dispatch — DID NOT WORK"). The
-- scheduler's own fireTrigger path was working throughout the same window.
-- So the acquisition lane is wired the way the model-directory lane already
-- is — a scheduled_tasks row — rather than depending on a trigger mechanism
-- that is currently unreliable for reasons nobody has established.
--
-- TOPIC: NOT a per-type topic. A custom agent type running on the shared
-- agent-chassis is dispatched via the generic requests topic with the real
-- type carried in config.agent_type (cmd/scheduler/main.go fireTrigger).
-- A dedicated system.agent.adoption-researcher.requests has no consumer and
-- the task would fire into the void, silently.
--
-- CADENCE, and why it differs from the model directory's: model prices change
-- and need a weekly sweep. Adoption announcements ACCUMULATE — a case study
-- published in March is still true in July — so discovery runs weekly to catch
-- new ones but the freshness re-verification is a fortnightly formality rather
-- than the model directory's daily price check. The one thing that does drift
-- is rollout_scope and agent_count, which is why those carry staleness_days
-- 200 in the researcher's prompt rather than 400.

BEGIN;

-- 1) Discovery — find new adoption announcements weekly.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds
)
SELECT
    'adoption-tracker-discovery',
    'Adoption tracker (Phase E): discover organisations deploying AI agents and register their cited claims (framework, rollout scope, claimed result, and how it was measured) via adoption-researcher.',
    604800,
    'adoption-researcher',
    'system.agent.generic.requests',
    jsonb_build_object(
        'research_query',
        'named companies deploying AI agents in production with reported results, rollout scope and how the result was measured'
    ),
    'adoption-tracker-discovery',
    1,
    true,
    900
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'adoption-tracker-discovery');

-- 2) Protocol discovery — MCP and successors. A separate task rather than a
--    second query on the same one, because the two searches want different
--    words and a single weekly slot would starve one of them.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds
)
SELECT
    'protocol-tracker-discovery',
    'Adoption tracker (Phase E): discover agent communication protocols (Model Context Protocol and successors) and register cited facts about their specification, stewardship and uptake.',
    604800,
    'adoption-researcher',
    'system.agent.generic.requests',
    jsonb_build_object(
        'research_query',
        'Model Context Protocol MCP and other agent communication protocols: specification versions, governance, and which organisations have announced adoption'
    ),
    'adoption-tracker-discovery',
    1,
    true,
    900
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'protocol-tracker-discovery');

-- 3) Freshness — re-verify company/protocol claims whose staleness has
--    elapsed. Deterministic: re-fetch each cited URL, re-check the quote, and
--    supersede on any status transition. No LLM in this path at all, which is
--    the point — the model proposed once; from here on a string comparison
--    disposes.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds
)
SELECT
    'adoption-tracker-freshness',
    'Adoption tracker (Phase E): re-verify is_current company/protocol directory_claims whose staleness_days has elapsed; supersede on any status transition; raise stale_directory_claim on a flip away from found.',
    1209600,
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
                        'config', jsonb_build_object('kind', 'company'),
                        'next_step', 'refresh_protocols',
                        'output_field', 'refresh_result',
                        'description', 'Re-verify due company claims'
                    ),
                    'refresh_protocols', jsonb_build_object(
                        'action', 'refresh_directory_claims',
                        'config', jsonb_build_object('kind', 'protocol'),
                        'next_step', 'complete',
                        'output_field', 'refresh_protocol_result',
                        'description', 'Re-verify due protocol claims'
                    ),
                    'complete', jsonb_build_object(
                        'action', 'complete_workflow',
                        'config', jsonb_build_object(
                            'output_fields', jsonb_build_array('refresh_result', 'refresh_protocol_result')
                        )
                    )
                )
            )
        )
    ),
    'adoption-tracker-freshness',
    1,
    true,
    900
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'adoption-tracker-freshness');

SELECT name, interval_seconds, target_agent_type, enabled
FROM scheduled_tasks WHERE name LIKE 'adoption-%' OR name LIKE 'protocol-%' ORDER BY name;

COMMIT;

-- ── Post-apply verification ────────────────────────────────────────────────
-- 1) The rows fired at all (last_triggered_at stops being NULL):
--      SELECT name, last_triggered_at, last_completed_at FROM scheduled_tasks
--      WHERE name IN ('adoption-tracker-discovery','protocol-tracker-discovery','adoption-tracker-freshness');
-- 2) To make discovery run NOW rather than waiting a week (the same trick the
--    model-directory lane uses):
--      UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = 'adoption-tracker-discovery';
-- 3) The register gained a second kind:
--      SELECT kind, count(*) FROM directory_entities GROUP BY 1;
-- 4) What good looks like — and what to be SUSPICIOUS of. A first run over
--    marketing material that verifies 100% is less believable than a mixed
--    one; the verifier is supposed to reject paraphrase:
--      SELECT dc.status, count(*) FROM directory_claims dc
--      JOIN directory_entities de ON de.id = dc.entity_id
--      WHERE de.kind IN ('company','protocol') AND dc.is_current GROUP BY 1;
-- 5) Nothing is published to any site yet: the resolver, publish profiles and
--    discovery checks for kind='company'/'protocol' are in commit f1dafb6e4
--    (council-approved, corr c67ecb24) and INERT until the next image roll.
--    Until then this lane fills the register and shows nowhere, deliberately.
