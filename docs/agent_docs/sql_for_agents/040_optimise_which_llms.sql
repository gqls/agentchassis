-- Migration 081: Model upgrades and LLM call logging
--
-- Part 1: Upgrade chief-strategist and site-planner to claude-sonnet-4-5
-- Part 2: Update stale model references to current aliases
-- Part 3: Create llm_call_log table for training data collection
--
-- Run against: clients database

BEGIN;

-- ============================================================================
-- Part 1: Upgrade planning agents to Sonnet 4.5
-- These agents make high-leverage decisions (site structure, page hierarchy)
-- that determine quality of everything downstream
-- ============================================================================

-- chief-strategist: haiku-4-5 → sonnet-4-5 (all occurrences in config)
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '"model": "claude-haiku-4-5"',
        '"model": "claude-sonnet-4-5"'
                     )::jsonb,
updated_at = NOW()
WHERE type = 'chief-strategist'
  AND default_config::text LIKE '%claude-haiku-4-5%';

-- site-planner: haiku-4-5 → sonnet-4-5
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

-- reasoning agent: claude-3-opus and claude-3-5-sonnet → claude-sonnet-4-5
UPDATE agent_definitions
SET default_config = replace(
        replace(
                default_config::text,
                '"model": "claude-3-opus"',
                '"model": "claude-sonnet-4-5"'
        ),
        '"model": "claude-3-5-sonnet"',
        '"model": "claude-sonnet-4-5"'
                     )::jsonb,
updated_at = NOW()
WHERE type = 'reasoning'
  AND (default_config::text LIKE '%claude-3-opus%' OR default_config::text LIKE '%claude-3-5-sonnet%');

-- researcher: same treatment
UPDATE agent_definitions
SET default_config = replace(
        replace(
                default_config::text,
                '"model": "claude-3-opus"',
                '"model": "claude-sonnet-4-5"'
        ),
        '"model": "claude-3-5-sonnet"',
        '"model": "claude-sonnet-4-5"'
                     )::jsonb,
updated_at = NOW()
WHERE type = 'researcher'
  AND (default_config::text LIKE '%claude-3-opus%' OR default_config::text LIKE '%claude-3-5-sonnet%');

-- copywriter: claude-3-sonnet and claude-3-5-sonnet → haiku-4-5 (templated copy)
UPDATE agent_definitions
SET default_config = replace(
        replace(
                default_config::text,
                '"model": "claude-3-sonnet"',
                '"model": "claude-haiku-4-5"'
        ),
        '"model": "claude-3-5-sonnet"',
        '"model": "claude-haiku-4-5"'
                     )::jsonb,
updated_at = NOW()
WHERE type = 'copywriter'
  AND (default_config::text LIKE '%claude-3-sonnet%' OR default_config::text LIKE '%claude-3-5-sonnet%');

-- Section-specific content creators: 3-5-sonnet → haiku-4-5
-- Short, constrained outputs (hero, CTA, contact, features, testimonials)
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '"model": "claude-3-5-sonnet"',
        '"model": "claude-haiku-4-5"'
                     )::jsonb,
updated_at = NOW()
WHERE type IN (
    'content-creator-hero',
    'content-creator-cta',
    'content-creator-contact',
    'content-creator-features',
    'content-creator-testimonials',
    'simple-content-writer-with-approval'
    )
  AND default_config::text LIKE '%claude-3-5-sonnet%';

-- website-builder: 3-5-sonnet → haiku-4-5 (orchestration agent, minimal LLM use)
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '"model": "claude-3-5-sonnet"',
        '"model": "claude-haiku-4-5"'
                     )::jsonb,
updated_at = NOW()
WHERE type = 'website-builder'
  AND default_config::text LIKE '%claude-3-5-sonnet%';


-- ============================================================================
-- Part 3: LLM call logging table for training data collection
-- Every execute_llm_prompt call gets logged here
-- ============================================================================

CREATE TABLE IF NOT EXISTS llm_call_log (
                                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Context: who called this and why
    agent_type VARCHAR(100) NOT NULL,
    step_name VARCHAR(100) NOT NULL,
    orchestration_id VARCHAR(255),
    correlation_id VARCHAR(255),
    client_id VARCHAR(100),

    -- LLM configuration
    provider VARCHAR(50) NOT NULL DEFAULT 'anthropic',
    model VARCHAR(100) NOT NULL,
    model_resolved VARCHAR(100),  -- actual API model string after alias resolution
    temperature REAL,
    max_tokens INTEGER,

    -- The actual I/O (this is the training data)
    prompt_template TEXT,          -- raw template before rendering
    prompt_rendered TEXT NOT NULL,  -- rendered prompt sent to LLM
    response_text TEXT,            -- raw LLM response
    response_type VARCHAR(20),     -- 'json' or 'text'

-- Performance
    input_tokens INTEGER,
    output_tokens INTEGER,
    latency_ms INTEGER,

    -- Outcome
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Indexes for common query patterns
CREATE INDEX idx_llm_log_agent_step ON llm_call_log(agent_type, step_name);
CREATE INDEX idx_llm_log_model ON llm_call_log(model);
CREATE INDEX idx_llm_log_created ON llm_call_log(created_at DESC);
CREATE INDEX idx_llm_log_orch ON llm_call_log(orchestration_id) WHERE orchestration_id IS NOT NULL;
CREATE INDEX idx_llm_log_errors ON llm_call_log(success) WHERE NOT success;

-- Composite for training data export: "all successful calls for agent X"
CREATE INDEX idx_llm_log_training ON llm_call_log(agent_type, step_name, success)
    WHERE success = true;

-- Cleanup function — call from pg_cron or maintenance-catch-all
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