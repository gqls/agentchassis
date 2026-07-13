-- Migration 081: Model upgrades and LLM call logging
-- 
-- Part 1: Upgrade planning agents to claude-sonnet-4-5
-- Part 2: Update stale model references to current aliases  
-- Part 3: Create llm_call_log table for training data collection
--
-- Run against: clients database

BEGIN;

-- ============================================================================
-- Part 1: Upgrade planning agents to Sonnet 4.5
-- ============================================================================

-- chief-strategist: all model refs → sonnet-4-5
UPDATE agent_definitions
SET default_config = replace(
    default_config::text,
    '"model": "claude-haiku-4-5"',
    '"model": "claude-sonnet-4-5"'
)::jsonb,
updated_at = NOW()
WHERE type = 'chief-strategist'
  AND default_config::text LIKE '%claude-haiku-4-5%';

-- site-planner: all model refs → sonnet-4-5
UPDATE agent_definitions
SET default_config = replace(
    default_config::text,
    '"model": "claude-haiku-4-5"',
    '"model": "claude-sonnet-4-5"'
)::jsonb,
updated_at = NOW()
WHERE type = 'site-planner'
  AND default_config::text LIKE '%claude-haiku-4-5%';

-- ============================================================================
-- Part 2: Update stale model references
-- ============================================================================

-- reasoning agent: old refs → sonnet-4-5
UPDATE agent_definitions
SET default_config = replace(
    replace(
        default_config::text,
        '"model": "claude-3-opus-20240229"',
        '"model": "claude-sonnet-4-5"'
    ),
    '"model": "claude-3-5-sonnet-20241022"',
    '"model": "claude-sonnet-4-5"'
)::jsonb,
updated_at = NOW()
WHERE type = 'reasoning'
  AND (default_config::text LIKE '%claude-3-opus%' 
    OR default_config::text LIKE '%claude-3-5-sonnet-20241022%');

-- domain-analyst: old ref → sonnet-4-5
UPDATE agent_definitions
SET default_config = replace(
    default_config::text,
    '"model": "claude-3-5-sonnet-20241022"',
    '"model": "claude-sonnet-4-5"'
)::jsonb,
updated_at = NOW()
WHERE type = 'domain-analyst'
  AND default_config::text LIKE '%claude-3-5-sonnet-20241022%';

-- researcher: old ref → sonnet-4-5
UPDATE agent_definitions
SET default_config = replace(
    default_config::text,
    '"model": "claude-3-5-sonnet-20241022"',
    '"model": "claude-sonnet-4-5"'
)::jsonb,
updated_at = NOW()
WHERE type IN ('research-agent', 'content-researcher')
  AND default_config::text LIKE '%claude-3-5-sonnet-20241022%';

-- ============================================================================
-- Part 3: LLM call logging table
-- ============================================================================

CREATE TABLE IF NOT EXISTS llm_call_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Who called
    agent_type VARCHAR(100) NOT NULL,
    agent_id VARCHAR(255),
    step_name VARCHAR(255),
    orchestration_id VARCHAR(255),
    correlation_id VARCHAR(255),
    
    -- What was called
    model VARCHAR(100) NOT NULL,
    model_resolved VARCHAR(100),   -- actual API string after alias resolution
    provider VARCHAR(50) DEFAULT 'anthropic',
    
    -- Prompt data (for training export)
    prompt_template TEXT,
    prompt_rendered TEXT,
    response_text TEXT,
    
    -- Usage
    input_tokens INTEGER,
    output_tokens INTEGER,
    latency_ms INTEGER,
    
    -- Outcome
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    
    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_llm_log_agent ON llm_call_log(agent_type, step_name);
CREATE INDEX idx_llm_log_model ON llm_call_log(model);
CREATE INDEX idx_llm_log_created ON llm_call_log(created_at DESC);
CREATE INDEX idx_llm_log_orch ON llm_call_log(orchestration_id) WHERE orchestration_id IS NOT NULL;
CREATE INDEX idx_llm_log_errors ON llm_call_log(success) WHERE NOT success;

-- Composite for training data export
CREATE INDEX idx_llm_log_training ON llm_call_log(agent_type, step_name, success)
    WHERE success = true;

-- Cleanup function
CREATE OR REPLACE FUNCTION cleanup_old_llm_logs()
RETURNS void AS $$
BEGIN
    DELETE FROM llm_call_log
    WHERE created_at < NOW() - INTERVAL '180 days' AND NOT success;
    DELETE FROM llm_call_log
    WHERE created_at < NOW() - INTERVAL '90 days' AND success;
END;
$$ LANGUAGE plpgsql;

-- Stats view
CREATE OR REPLACE VIEW llm_call_stats AS
SELECT 
    agent_type,
    step_name,
    model,
    COUNT(*) as call_count,
    COUNT(*) FILTER (WHERE success) as success_count,
    ROUND(AVG(latency_ms)) as avg_latency_ms,
    ROUND(AVG(input_tokens)) as avg_input_tokens,
    ROUND(AVG(output_tokens)) as avg_output_tokens,
    MIN(created_at) as first_call,
    MAX(created_at) as last_call
FROM llm_call_log
GROUP BY agent_type, step_name, model;

COMMIT;
