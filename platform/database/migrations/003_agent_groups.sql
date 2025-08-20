-- 003_agent_groups.sql
-- Agent groups table for reusable team configurations

CREATE TABLE IF NOT EXISTS agent_groups (
                                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    group_type VARCHAR(100) NOT NULL,
    version VARCHAR(50) DEFAULT '1.0.0',
    parent_id UUID REFERENCES agent_groups(id), -- For evolution tracking

-- Configuration
    agent_configs JSONB NOT NULL,
    orchestration_workflow JSONB NOT NULL,

    -- Discovery metadata
    capabilities JSONB DEFAULT '[]'::jsonb,
    tags JSONB DEFAULT '[]'::jsonb,

    -- Evolution tracking
    mutation_history JSONB DEFAULT '[]'::jsonb, -- Track all changes from parent

-- Performance tracking
    performance_metrics JSONB DEFAULT '{}'::jsonb,
    usage_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,

                               -- Metadata
                               created_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

-- Indexes for discovery
CREATE INDEX IF NOT EXISTS idx_agent_groups_type ON agent_groups(group_type);
CREATE INDEX IF NOT EXISTS idx_agent_groups_capabilities ON agent_groups USING gin(capabilities);
CREATE INDEX IF NOT EXISTS idx_agent_groups_tags ON agent_groups USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_agent_groups_usage ON agent_groups(usage_count DESC);
CREATE INDEX IF NOT EXISTS idx_agent_groups_parent ON agent_groups(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_agent_groups_mutation_history ON agent_groups USING gin(mutation_history);

-- Comments
COMMENT ON TABLE agent_groups IS 'Reusable agent team configurations with evolution tracking';
COMMENT ON COLUMN agent_groups.mutation_history IS 'JSON array tracking all changes/mutations from parent groups';
COMMENT ON COLUMN agent_groups.parent_id IS 'Reference to parent group for version evolution tracking';
COMMENT ON COLUMN agent_groups.performance_metrics IS 'JSON object with execution metrics and performance data';