-- 008_add_discovery_functions.sql
-- Functions for agent discovery and capability matching

-- Function to find agents by capability
CREATE OR REPLACE FUNCTION find_agents_by_capability(
    required_capabilities TEXT[],
    client_id TEXT DEFAULT 'demo_client'
)
RETURNS TABLE (
    agent_id UUID,
    agent_type VARCHAR,
    agent_name VARCHAR,
    capabilities JSONB,
    performance_score DECIMAL
) AS $$
DECLARE
schema_name TEXT := 'client_' || client_id;
BEGIN
    -- Check if schema exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = schema_name) THEN
        RAISE EXCEPTION 'Client schema % does not exist', schema_name;
END IF;

    -- Dynamic query to search in the correct client schema
RETURN QUERY EXECUTE format('
        SELECT
            ai.id,
            ai.config->>''agent_type'',
            ai.name,
            ai.config->''capabilities'',
            COALESCE(am.success_rate, 0.5) as performance_score
        FROM %I.agent_instances ai
        LEFT JOIN agent_metrics am ON ai.id = am.agent_id
        WHERE ai.config->''capabilities'' ?| $1
          AND ai.is_active = true
        ORDER BY performance_score DESC',
        schema_name
    ) USING required_capabilities;
END;
$$ LANGUAGE plpgsql;

-- Function to find agent groups by capability
CREATE OR REPLACE FUNCTION find_agent_groups_by_capability(
    required_capabilities TEXT[]
)
RETURNS TABLE (
    group_id UUID,
    group_name VARCHAR,
    group_type VARCHAR,
    version VARCHAR,
    capabilities JSONB,
    usage_count INTEGER,
    last_used_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
RETURN QUERY
SELECT
    ag.id,
    ag.name,
    ag.group_type,
    ag.version,
    ag.capabilities,
    ag.usage_count,
    ag.last_used_at
FROM agent_groups ag
WHERE ag.capabilities ?| required_capabilities
ORDER BY ag.usage_count DESC NULLS LAST, ag.updated_at DESC;
END;
$$ LANGUAGE plpgsql;

-- Function to get agent performance summary
CREATE OR REPLACE FUNCTION get_agent_performance_summary(
    p_agent_type VARCHAR DEFAULT NULL,
    p_limit INTEGER DEFAULT 10
)
RETURNS TABLE (
    agent_id UUID,
    agent_type VARCHAR,
    total_tasks INTEGER,
    success_rate DECIMAL,
    avg_response_time_ms INTEGER,
    avg_fuel_per_task INTEGER,
    avg_quality_score DECIMAL
) AS $$
BEGIN
RETURN QUERY
SELECT
    am.agent_id,
    am.agent_type,
    am.total_tasks,
    am.success_rate,
    am.avg_response_time_ms,
    am.avg_fuel_per_task,
    am.avg_quality_score
FROM agent_metrics am
WHERE (p_agent_type IS NULL OR am.agent_type = p_agent_type)
  AND am.total_tasks > 0
ORDER BY am.success_rate DESC, am.avg_response_time_ms ASC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Function to recommend agents for a task
CREATE OR REPLACE FUNCTION recommend_agents_for_task(
    task_type VARCHAR,
    required_capabilities TEXT[] DEFAULT NULL
)
RETURNS TABLE (
    agent_type VARCHAR,
    display_name VARCHAR,
    category VARCHAR,
    capabilities JSONB,
    performance_score DECIMAL,
    recommendation_reason TEXT
) AS $$
BEGIN
RETURN QUERY
    WITH agent_scores AS (
        SELECT
            ad.type,
            ad.display_name,
            ad.category,
            ad.capabilities,
            COALESCE(am.success_rate, 0.5) as performance_score,
            COALESCE(am.total_tasks, 0) as experience,
            CASE
                -- Direct capability match
                WHEN required_capabilities IS NOT NULL
                     AND ad.capabilities ?| required_capabilities THEN 3
                -- Partial capability match
                WHEN required_capabilities IS NOT NULL
                     AND ad.capabilities ?| ARRAY(
                         SELECT unnest(required_capabilities)
                         INTERSECT
                         SELECT jsonb_array_elements_text(ad.capabilities)
                     ) THEN 2
                -- Category match
                WHEN ad.category = task_type THEN 1
                ELSE 0
            END as match_score
        FROM agent_definitions ad
        LEFT JOIN (
            SELECT agent_type,
                   AVG(success_rate) as success_rate,
                   SUM(total_tasks) as total_tasks
            FROM agent_metrics
            GROUP BY agent_type
        ) am ON ad.type = am.agent_type
        WHERE ad.is_active = true
          AND ad.deleted_at IS NULL
    )
SELECT
    type,
    display_name,
    category,
    capabilities,
    performance_score,
    CASE
        WHEN match_score = 3 THEN 'Perfect capability match'
        WHEN match_score = 2 THEN 'Partial capability match'
        WHEN match_score = 1 THEN 'Category match'
        WHEN experience > 100 THEN 'High experience with ' || experience || ' tasks'
        WHEN performance_score > 0.8 THEN 'High success rate: ' || ROUND(performance_score * 100) || '%'
        ELSE 'Available agent'
        END as recommendation_reason
FROM agent_scores
WHERE match_score > 0 OR performance_score > 0.7
ORDER BY match_score DESC, performance_score DESC, experience DESC
    LIMIT 5;
END;
$$ LANGUAGE plpgsql;

-- Add helpful comments
COMMENT ON FUNCTION find_agents_by_capability IS 'Find agent instances by required capabilities for a specific client';
COMMENT ON FUNCTION find_agent_groups_by_capability IS 'Find agent groups that have the required capabilities';
COMMENT ON FUNCTION get_agent_performance_summary IS 'Get performance metrics summary for agents';
COMMENT ON FUNCTION recommend_agents_for_task IS 'Recommend best agents for a specific task based on capabilities and performance';