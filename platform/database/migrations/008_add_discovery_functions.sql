-- Function to find agents by capability
CREATE OR REPLACE FUNCTION find_agents_by_capability(
    required_capabilities TEXT[],
    client_id TEXT
)
RETURNS TABLE (
    agent_id UUID,
    agent_type VARCHAR,
    agent_name VARCHAR,
    capabilities JSONB,
    performance_score DECIMAL
) AS $$
BEGIN
RETURN QUERY
SELECT
    ai.id,
    ai.config->>'agent_type',
    ai.name,
    ai.config->'capabilities',
    COALESCE(am.success_rate, 0.5) as performance_score
FROM client_demo_client.agent_instances ai
    LEFT JOIN agent_metrics am ON ai.id = am.agent_id
WHERE ai.config->'capabilities' ?| required_capabilities
  AND ai.is_active = true
ORDER BY performance_score DESC;
END;
$$ LANGUAGE plpgsql;