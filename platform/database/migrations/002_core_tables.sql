-- 002_core_tables.sql
-- Global agent definitions table (shared across all clients)
CREATE TABLE IF NOT EXISTS agent_definitions (
                                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL CHECK (category IN ('data-driven', 'code-driven', 'adapter')),
    default_config JSONB NOT NULL DEFAULT '{}',
    capabilities JSONB DEFAULT '[]'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
    );

-- Index for active agent types
CREATE INDEX IF NOT EXISTS idx_agent_definitions_type_active
    ON agent_definitions(type, is_active) WHERE deleted_at IS NULL;

-- Global orchestrator state table (shared across all clients)
CREATE TABLE IF NOT EXISTS orchestrator_state (
  correlation_id UUID PRIMARY KEY,
  client_id VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    workflow_plan JSONB,
    execution_metadata JSONB DEFAULT '{}',
    execution_path JSONB DEFAULT '[]',
    current_step VARCHAR(255) NOT NULL,
    awaited_steps JSONB DEFAULT '[]',
    collected_data JSONB DEFAULT '{}',
    initial_request_data JSONB,
    final_result JSONB,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Indexes for orchestrator state
CREATE INDEX IF NOT EXISTS idx_orchestrator_state_status ON orchestrator_state(status);
CREATE INDEX IF NOT EXISTS idx_orchestrator_state_client ON orchestrator_state(client_id);
CREATE INDEX IF NOT EXISTS idx_orchestrator_state_updated_at ON orchestrator_state(updated_at);
CREATE INDEX IF NOT EXISTS idx_orchestrator_state_client_status ON orchestrator_state(client_id, status);

-- Insert default agent definitions
INSERT INTO agent_definitions (type, display_name, description, category, default_config) VALUES
                                                                                              ('copywriter', 'Copywriter', 'Creates compelling marketing and content copy', 'data-driven', '{"model": "claude-3-sonnet", "temperature": 0.7}'),
                                                                                              ('researcher', 'Research Assistant', 'Conducts thorough research and analysis', 'data-driven', '{"model": "claude-3-opus", "temperature": 0.3}'),
                                                                                              ('reasoning', 'Reasoning Agent', 'Performs logical analysis and decision making', 'code-driven', '{"model": "claude-3-opus", "temperature": 0.2}'),
                                                                                              ('image-generator', 'Image Generator', 'Creates images using AI generation', 'adapter', '{"provider": "stability_ai", "model": "sdxl"}'),
                                                                                              ('web-search', 'Web Search', 'Searches the internet for information', 'adapter', '{"provider": "serpapi", "max_results": 10}')
    ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                              description = EXCLUDED.description,
                              category = EXCLUDED.category,
                              default_config = EXCLUDED.default_config,
                              updated_at = NOW();


