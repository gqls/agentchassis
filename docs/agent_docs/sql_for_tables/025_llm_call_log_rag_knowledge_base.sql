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

---
-- fix schema

-- Add the agent_id column the Go code expects
ALTER TABLE llm_call_log ADD COLUMN IF NOT EXISTS agent_id VARCHAR(255);

-- Relax step_name to nullable since Go sends nullIfEmpty
ALTER TABLE llm_call_log ALTER COLUMN step_name DROP NOT NULL;

-- Relax prompt_rendered to nullable since Go sends nullIfEmpty
ALTER TABLE llm_call_log ALTER COLUMN prompt_rendered DROP NOT NULL;

----

-- ============================================================================
-- Training Data Export Queries
-- ============================================================================
-- Use these when you have enough examples to fine-tune open-weight models.
-- Check readiness first, then export in the format your training pipeline needs.
-- ============================================================================


-- ── 1. Check training data readiness ─────────────────────────────────────

-- How many examples per agent type and step?
SELECT agent_type, step_name, model, count(*) as examples,
       count(*) FILTER (WHERE success = true) as successful,
    avg(output_tokens) FILTER (WHERE success = true) as avg_output_tokens
FROM llm_call_log
WHERE created_at > NOW() - INTERVAL '90 days'
GROUP BY agent_type, step_name, model
ORDER BY examples DESC;

-- How many tool recreation training triples?
SELECT
    count(*) as total_triples,
    count(*) FILTER (WHERE data->'metadata'->>'complete' = 'true') as complete,
    count(DISTINCT data->'metadata'->>'domain') as unique_sites,
    count(DISTINCT data->'metadata'->>'vertical') as unique_verticals
FROM research_results
WHERE result_type = 'tool_recreation_training';

-- Training examples linked to work item outcomes
SELECT
    l.agent_type, l.step_name, l.model,
    count(*) as total_calls,
    count(*) FILTER (WHERE w.status = 'complete') as led_to_success,
    count(*) FILTER (WHERE w.status = 'failed') as led_to_failure
FROM llm_call_log l
         JOIN site_work_items w ON w.id = l.work_item_id
WHERE l.success = true
  AND l.work_item_id IS NOT NULL
GROUP BY l.agent_type, l.step_name, l.model
ORDER BY total_calls DESC;


-- ── 2. Export: Tool functional spec generation (analyze_tool) ────────────
-- Input: site context + source HTML → Output: JSON functional spec
-- Target: fine-tune Llama/Qwen for structured analysis

-- As JSONL for Unsloth/HuggingFace:
COPY (
SELECT json_build_object(
               'instruction', 'Analyse this interactive web tool and produce a detailed JSON functional specification.',
               'input', prompt_rendered,
               'output', response_text,
               'metadata', json_build_object(
                       'model', model,
                       'vertical', vertical,
                       'tokens', output_tokens,
                       'work_item_id', work_item_id
                           )
       )
FROM llm_call_log
WHERE agent_type = 'tool-recreation-handler'
  AND step_name = 'analyze_tool'
  AND success = true
  AND response_text IS NOT NULL
  AND LENGTH(response_text) > 500
ORDER BY created_at
    ) TO '/tmp/analyze_tool_training.jsonl';


-- ── 3. Export: Tool code generation (recreate_tool) ──────────────────────
-- Input: functional spec + source → Output: working HTML/CSS/JS
-- Target: fine-tune for code generation (hardest task)
-- Uses the training triples which include quality signals

COPY (
SELECT json_build_object(
               'instruction', 'Recreate this interactive web tool as a self-contained HTML/CSS/JavaScript application.',
               'input', json_build_object(
                       'functional_spec', data->'functional_spec',
                       'source_html_preview', LEFT(data->>'source_html', 2000)
    ),
               'output', data->>'recreated_html',
               'metadata', data->'metadata'
       )
FROM research_results
WHERE result_type = 'tool_recreation_training'
  AND data->'metadata'->>'complete' = 'true'
  AND LENGTH(data->>'recreated_html') > 1000
ORDER BY created_at
    ) TO '/tmp/recreate_tool_training.jsonl';


-- ── 4. Export: Site classification ───────────────────────────────────────
-- High-volume, easy task — good first candidate for model swap

COPY (
SELECT json_build_object(
               'instruction', 'Classify this website based on its domain and crawl data.',
               'input', prompt_rendered,
               'output', response_text
       )
FROM llm_call_log
WHERE step_name IN ('analyze_site', 'classify_archetype')
  AND success = true
  AND response_text IS NOT NULL
  AND LENGTH(response_text) > 100
ORDER BY created_at
    ) TO '/tmp/site_classification_training.jsonl';


-- ── 5. Export: Content writing (with quality filter) ─────────────────────
-- Only export examples where the work item completed without audit failures

COPY (
SELECT json_build_object(
               'instruction', 'Write content for this website section.',
               'input', prompt_rendered,
               'output', response_text,
               'metadata', json_build_object(
                       'vertical', vertical,
                       'model', model
                           )
       )
FROM llm_call_log l
WHERE l.agent_type = 'page-content-writer'
  AND l.step_name = 'generate_content'
  AND l.success = true
  AND l.work_item_id IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM site_work_items w
    WHERE w.id = l.work_item_id AND w.status = 'complete'
)
  AND LENGTH(response_text) > 200
ORDER BY created_at
    ) TO '/tmp/content_writing_training.jsonl';


-- ── 6. Per-vertical counts (decide which verticals have enough data) ────

SELECT
    vertical,
    count(*) as total_calls,
    count(DISTINCT agent_type) as agent_types,
    count(*) FILTER (WHERE step_name = 'generate_content') as content_calls,
    count(*) FILTER (WHERE step_name = 'analyze_tool') as analyze_calls,
    count(*) FILTER (WHERE step_name = 'recreate_tool') as recreate_calls
FROM llm_call_log
WHERE success = true
  AND vertical IS NOT NULL
GROUP BY vertical
ORDER BY total_calls DESC;

