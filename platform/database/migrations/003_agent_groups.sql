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
CREATE INDEX idx_agent_groups_type ON agent_groups(group_type);
CREATE INDEX idx_agent_groups_capabilities ON agent_groups USING gin(capabilities);
CREATE INDEX idx_agent_groups_tags ON agent_groups USING gin(tags);
CREATE INDEX idx_agent_groups_usage ON agent_groups(usage_count DESC);