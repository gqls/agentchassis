-- Migration 082: RAG knowledge base for industry content
--
-- Shared knowledge base for storing embedded content from:
-- - Scraped competitor/exemplar sites
-- - Research results
-- - Curated industry information
-- - Component usage patterns
--
-- This is NOT per-agent-instance (unlike agent_memory).
-- It's a shared resource that any content-creating agent can query.
--
-- IMPORTANT: The embedding column is vector(768), sized for nomic-embed-text.
-- If you change to a model with different dimensions (e.g. text-embedding-ada-002
-- at 1536), you must ALTER the column and rebuild the index.
-- The embedding_model column tracks what was used per row.
--
-- Run against: clients database

BEGIN;

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS knowledge_base (
                                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Classification
    collection VARCHAR(100) NOT NULL,  -- e.g. 'industry_sites', 'research', 'components'
    industry VARCHAR(100),             -- e.g. 'veterinary', 'bakery', 'legal'
    domain VARCHAR(255),               -- source domain if from a scrape

-- Content
    title VARCHAR(500),
    content TEXT NOT NULL,
    content_hash VARCHAR(64),           -- SHA256 for dedup

-- Embedding: 768 dimensions for nomic-embed-text / bge-base-en-v1.5
    embedding vector(768),
    embedding_model VARCHAR(100) DEFAULT 'nomic-embed-text',

    -- Source tracking
    source_type VARCHAR(50),           -- 'scrape', 'research', 'manual', 'component_log'
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

-- Stats view
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