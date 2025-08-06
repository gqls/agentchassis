-- test/scripts/verify_results.sql
-- Comprehensive test verification queries

\echo '=== Test Verification Report ==='
\echo ''

\echo '1. Recent Test Workflows (Last Hour)'
SELECT
    correlation_id,
    status,
    current_step,
    EXTRACT(EPOCH FROM (NOW() - created_at)) as seconds_ago,
    execution_metadata->>'completed_steps' as completed,
    execution_metadata->>'total_steps' as total
FROM orchestrator_state
WHERE correlation_id LIKE 'test-%'
  AND created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC
    LIMIT 10;

\echo ''
\echo '2. Test Agent Instances'
SELECT
    id,
    name,
    config->>'agent_type' as type,
    is_active,
    created_at
FROM client_demo_client.agent_instances
WHERE name LIKE 'test-%'
   OR id LIKE '00000000-0000-0000-0000-%'
ORDER BY created_at DESC
    LIMIT 10;

\echo ''
\echo '3. Agent Group Usage'
SELECT
    name,
    group_type,
    usage_count,
    last_used_at,
    performance_metrics->>'success_rate' as success_rate
FROM agent_groups
WHERE usage_count > 0
ORDER BY last_used_at DESC NULLS LAST;

\echo ''
\echo '4. Test Summary'
SELECT
    COUNT(*) FILTER (WHERE status = 'COMPLETED') as completed,
    COUNT(*) FILTER (WHERE status = 'FAILED') as failed,
    COUNT(*) FILTER (WHERE status IN ('RUNNING', 'AWAITING_RESPONSES')) as active,
    COUNT(*) as total
FROM orchestrator_state
WHERE correlation_id LIKE 'test-%'
  AND created_at > NOW() - INTERVAL '24 hours';