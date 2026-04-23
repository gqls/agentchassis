-- ============================================================================
-- Migration 085: AI Endpoint Health Table + LLM Call Log Flywheel Columns
-- ============================================================================
-- Part A: ai_endpoint_health — the GPU/model scheduler.
--   Healthy endpoint → items flow. Unhealthy → items wait.
--   Checked by claim_work_item before claiming.
--   Updated by scheduler pings (active) and failed calls (reactive).
--
-- Part B: llm_call_log flywheel columns — link calls to work items,
--   track prompt variants, verticals, and RAG usage for training data.
-- ============================================================================


-- ============================================================================
-- Part A: ai_endpoint_health
-- ============================================================================

CREATE TABLE IF NOT EXISTS ai_endpoint_health (
                                                  endpoint_url  TEXT PRIMARY KEY,
                                                  name          TEXT NOT NULL,
                                                  healthy       BOOLEAN NOT NULL DEFAULT false,
                                                  last_checked  TIMESTAMPTZ,
                                                  last_healthy  TIMESTAMPTZ,
                                                  error         TEXT,
                                                  check_interval_seconds INT DEFAULT 60,
                                                  check_mode    TEXT NOT NULL DEFAULT 'active',  -- 'active' (scheduler pings) or 'reactive' (only on failure)
                                                  ping_path     TEXT DEFAULT '/api/tags',          -- path to ping for health, or 'claude_ping' for Anthropic
                                                  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Index for the dispatch loop query (check health before claiming)
CREATE INDEX IF NOT EXISTS idx_endpoint_health_name ON ai_endpoint_health(name);

-- Seed data: the three endpoints we know about
INSERT INTO ai_endpoint_health (endpoint_url, name, healthy, last_checked, last_healthy, check_interval_seconds, check_mode, ping_path)
VALUES
    ('https://api.anthropic.com/v1/messages', 'claude', true, NOW(), NOW(), 3600, 'active', 'claude_ping'),
    ('http://ollama-adapter.ai-persona-system.svc.cluster.local:11434', 'cpu-ollama', true, NOW(), NOW(), 60, 'active', '/api/tags'),
    ('http://ollama-gpu.ai-persona-system.svc.cluster.local:11434', 'gpu-ollama', false, NOW(), NULL, 30, 'active', '/api/tags')
    ON CONFLICT (endpoint_url) DO NOTHING;


-- Operator-friendly view
CREATE OR REPLACE VIEW ai_endpoint_status AS
SELECT
    name,
    healthy,
    CASE WHEN healthy THEN 'UP' ELSE 'DOWN' END as status,
    error,
    last_checked,
    last_healthy,
    CASE
        WHEN last_healthy IS NULL THEN 'never healthy'
        WHEN healthy THEN 'n/a'
        ELSE age(now(), last_healthy)::text
        END as down_since,
    check_interval_seconds,
    check_mode,
    endpoint_url
FROM ai_endpoint_health
ORDER BY name;


-- Helper function for reactive health updates (called from Go on failure)
CREATE OR REPLACE FUNCTION update_endpoint_health(
    p_endpoint_url TEXT,
    p_healthy BOOLEAN,
    p_error TEXT DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
UPDATE ai_endpoint_health
SET healthy = p_healthy,
    last_checked = NOW(),
    last_healthy = CASE WHEN p_healthy THEN NOW() ELSE last_healthy END,
    error = p_error,
    updated_at = NOW()
WHERE endpoint_url = p_endpoint_url;

-- If no row existed, this is a new endpoint — insert it
IF NOT FOUND THEN
        INSERT INTO ai_endpoint_health (endpoint_url, name, healthy, last_checked, last_healthy, error)
        VALUES (p_endpoint_url, p_endpoint_url, p_healthy, NOW(),
                CASE WHEN p_healthy THEN NOW() ELSE NULL END, p_error);
END IF;
END;
$$ LANGUAGE plpgsql;


-- ============================================================================
-- Part B: llm_call_log flywheel columns
-- ============================================================================
-- These columns enable training data collection as a byproduct of normal ops.
-- work_item_id links each LLM call to its outcome (did the fix work?).
-- prompt_variant enables A/B testing of prompt versions.
-- vertical enables per-industry analysis.
-- rag_context_used tracks whether RAG was involved.

ALTER TABLE llm_call_log
    ADD COLUMN IF NOT EXISTS work_item_id UUID,
    ADD COLUMN IF NOT EXISTS prompt_variant TEXT DEFAULT 'default',
    ADD COLUMN IF NOT EXISTS vertical TEXT,
    ADD COLUMN IF NOT EXISTS rag_context_used BOOLEAN DEFAULT false;

-- Index to join calls back to work item outcomes
CREATE INDEX IF NOT EXISTS idx_llm_call_log_work_item
    ON llm_call_log(work_item_id) WHERE work_item_id IS NOT NULL;

-- Index for per-vertical analysis
CREATE INDEX IF NOT EXISTS idx_llm_call_log_vertical
    ON llm_call_log(vertical) WHERE vertical IS NOT NULL;


-- ============================================================================
-- Verification
-- ============================================================================

SELECT 'ai_endpoint_health' as table_name, COUNT(*) as rows FROM ai_endpoint_health
UNION ALL
SELECT 'ai_endpoint_status view', COUNT(*) FROM ai_endpoint_status;

SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'llm_call_log'
  AND column_name IN ('work_item_id', 'prompt_variant', 'vertical', 'rag_context_used')
ORDER BY column_name;