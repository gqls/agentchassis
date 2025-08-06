-- Check spawned agents
SELECT id, name, config->>'agent_type' as type, is_active, created_at
FROM client_demo_client.agent_instances
WHERE created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC
LIMIT 10;

-- Check workflow states
SELECT correlation_id, status, current_step,
       execution_metadata->>'completed_steps' as completed
FROM orchestrator_state
WHERE created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC
LIMIT 10;

-- Check agent groups
SELECT id, name, group_type, usage_count, last_used_at
FROM agent_groups
WHERE last_used_at > NOW() - INTERVAL '1 hour'
ORDER BY last_used_at DESC NULLS LAST;