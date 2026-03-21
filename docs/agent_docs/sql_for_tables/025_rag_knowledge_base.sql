-- Migration 081: Model upgrades and LLM call logging
--
-- Part 1: Upgrade classifiers and planners to claude-sonnet-4-6 / claude-opus-4-6
-- Part 2: Update stale model references
-- Part 3: Create llm_call_log table for training data collection
--
-- Run against: clients database

BEGIN;

-- ============================================================================
-- Part 1: Upgrade planning and classification agents
-- These make high-leverage structural decisions. Worth using the best models.
-- ============================================================================

-- chief-strategist: → opus-4-6 (most important structural decisions)
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '"model": "claude-sonnet-4-5"',
        '"model": "claude-opus-4-6"'
                     )::jsonb,
updated_at = NOW()
WHERE type = 'chief-strategist'
  AND default_config::text LIKE '%claude-sonnet-4-5%';

-- site-planner: → sonnet-4-6 (page structure, component selection)
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '"model": "claude-sonnet-4-5"',
        '"model": "claude-sonnet-4-6"'
                     )::jsonb,
updated_at = NOW()
WHERE type IN ('site-planner', 'build-site-planner')
  AND default_config::text LIKE '%claude-sonnet-4-5%';

-- domain-research-classifier: → sonnet-4-6 (domain analysis, vertical classification)
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '"model": "claude-sonnet-4-5"',
        '"model": "claude-sonnet-4-6"'
                     )::jsonb,
updated_at = NOW()
WHERE type = 'domain-research-classifier'
  AND default_config::text LIKE '%claude-sonnet-4-5%';

-- domain-strategist: → sonnet-4-6 (revenue model, competitive positioning)
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '"model": "claude-sonnet-4-5"',
        '"model": "claude-sonnet-4-6"'
                     )::jsonb,
updated_at = NOW()
WHERE type = 'domain-strategist'
  AND default_config::text LIKE '%claude-sonnet-4-5%';

-- site-classifier: → sonnet-4-6 (vertical classification is now more complex)
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '"model": "claude-haiku-4-5"',
        '"model": "claude-sonnet-4-6"'
                     )::jsonb,
updated_at = NOW()
WHERE type = 'site-classifier'
  AND default_config::text LIKE '%claude-haiku-4-5%';

-- ============================================================================
-- Part 2: Update stale model references across all agents
-- ============================================================================

-- Any remaining claude-3-5-sonnet-20241022 refs → sonnet-4-6
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '"model": "claude-3-5-sonnet-20241022"',
        '"model": "claude-sonnet-4-6"'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text LIKE '%claude-3-5-sonnet-20241022%';

-- Any remaining claude-3-opus refs → sonnet-4-6
UPDATE agent_definitions
SET default_config = replace(
        default_config::text,
        '"model": "claude-3-opus-20240229"',
        '"model": "claude-sonnet-4-6"'
                     )::jsonb,
updated_at = NOW()
WHERE default_config::text LIKE '%claude-3-opus-20240229%';

-- ============================================================================
-- Part 3: LLM call logging table (idempotent)
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
    model_resolved VARCHAR(100),
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

-- Drop and recreate indexes (idempotent)
DROP INDEX IF EXISTS idx_llm_log_agent;
DROP INDEX IF EXISTS idx_llm_log_model;
DROP INDEX IF EXISTS idx_llm_log_created;
DROP INDEX IF EXISTS idx_llm_log_orch;
DROP INDEX IF EXISTS idx_llm_log_errors;
DROP INDEX IF EXISTS idx_llm_log_training;

CREATE INDEX idx_llm_log_agent ON llm_call_log(agent_type, step_name);
CREATE INDEX idx_llm_log_model ON llm_call_log(model);
CREATE INDEX idx_llm_log_created ON llm_call_log(created_at DESC);
CREATE INDEX idx_llm_log_orch ON llm_call_log(orchestration_id) WHERE orchestration_id IS NOT NULL;
CREATE INDEX idx_llm_log_errors ON llm_call_log(success) WHERE NOT success;
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


----------
-- rag knowledge base

-- Migration 082: RAG knowledge base for industry content (IDEMPOTENT)
--
-- Safe to re-run — drops existing indexes before recreating.
-- The table uses CREATE TABLE IF NOT EXISTS so it won't fail if already created.
--
-- Run against: clients database

BEGIN;

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS knowledge_base (
                                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Classification
    collection VARCHAR(100) NOT NULL,
    industry VARCHAR(100),
    domain VARCHAR(255),

    -- Content
    title VARCHAR(500),
    content TEXT NOT NULL,
    content_hash VARCHAR(64),

    -- Embedding: 768 dimensions for nomic-embed-text
    embedding vector(768),
    embedding_model VARCHAR(100) DEFAULT 'nomic-embed-text',

    -- Source tracking
    source_type VARCHAR(50),
    source_url TEXT,
    source_agent_type VARCHAR(100),
    source_orchestration_id VARCHAR(255),

    -- Metadata
    metadata JSONB DEFAULT '{}',
    quality_score REAL,
    usage_count INTEGER DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Drop existing indexes before recreating (makes this migration idempotent)
DROP INDEX IF EXISTS idx_kb_embedding;
DROP INDEX IF EXISTS idx_kb_collection;
DROP INDEX IF EXISTS idx_kb_industry;
DROP INDEX IF EXISTS idx_kb_collection_industry;
DROP INDEX IF EXISTS idx_kb_content_hash;
DROP INDEX IF EXISTS idx_kb_created;
DROP INDEX IF EXISTS idx_kb_dedup;
DROP INDEX IF EXISTS idx_kb_content_trgm;

-- Vector similarity search within a collection
CREATE INDEX idx_kb_embedding ON knowledge_base
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Filter indexes
CREATE INDEX idx_kb_collection ON knowledge_base(collection);
CREATE INDEX idx_kb_industry ON knowledge_base(industry) WHERE industry IS NOT NULL;
CREATE INDEX idx_kb_collection_industry ON knowledge_base(collection, industry);
CREATE INDEX idx_kb_content_hash ON knowledge_base(content_hash);
CREATE INDEX idx_kb_created ON knowledge_base(created_at DESC);

-- Dedup: same collection + same content = skip
CREATE UNIQUE INDEX idx_kb_dedup ON knowledge_base(collection, content_hash)
    WHERE content_hash IS NOT NULL;

-- Trigram index for keyword fallback when embedding is unavailable
CREATE INDEX idx_kb_content_trgm ON knowledge_base USING gin (content gin_trgm_ops);

-- Stats view (CREATE OR REPLACE is already idempotent)
CREATE OR REPLACE VIEW knowledge_base_stats AS
SELECT
    collection,
    industry,
    COUNT(*) as chunk_count,
    COUNT(*) FILTER (WHERE embedding IS NOT NULL) as embedded_count,
    ROUND(AVG(length(content))) as avg_chunk_length,
    SUM(usage_count) as total_retrievals,
    MIN(created_at) as oldest,
    MAX(created_at) as newest
FROM knowledge_base
GROUP BY collection, industry
ORDER BY chunk_count DESC;

COMMIT;

