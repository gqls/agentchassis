-- FILE: platform/database/migrations/002_core_tables.sql
-- Global agent definitions table (shared across all clients)
-- This is the single source of truth for all agent configurations
CREATE TABLE IF NOT EXISTS agent_definitions (
                                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic identification
    type VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL CHECK (category IN ('data-driven', 'code-driven', 'adapter', 'orchestrator')),

    -- Deployment configuration
    image_repository VARCHAR(255) NOT NULL DEFAULT 'docker.io/aqls/agent-chassis',
    image_tag VARCHAR(100) NOT NULL DEFAULT 'latest',
    command TEXT[], -- Array of command parts, e.g. ['./agent-chassis', '-config', 'configs/agent-chassis.yaml']

-- Resource requirements (NEW - replaces hardcoded resource maps)
    resources JSONB NOT NULL DEFAULT '{
        "requests": {"cpu": "100m", "memory": "256Mi"},
        "limits": {"cpu": "500m", "memory": "1Gi"}
    }'::jsonb,

    -- Runtime configuration
    default_config JSONB NOT NULL DEFAULT '{}',
    capabilities JSONB DEFAULT '[]'::jsonb,

    -- Kafka topic configuration (NEW - data-driven topic creation)
    topics JSONB DEFAULT '{
        "process": "system.agent.{type}.requests",
        "response": "system.agent.{type}.responses",
        "error": "system.agent.{type}.errors",
        "dlq": "system.agent.{type}.dlq"
    }'::jsonb,

    -- Health check configuration (NEW)
    health_config JSONB DEFAULT '{
        "liveness_path": "/health",
        "readiness_path": "/ready",
        "port": 8080,
        "initial_delay_seconds": 30
    }'::jsonb,

    -- Environment variables specific to this agent type (NEW)
    env_vars JSONB DEFAULT '[]'::jsonb, -- [{"name": "SPECIAL_CONFIG", "value": "xyz"}]

-- Status and lifecycle
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- Versioning support (NEW)
    version INTEGER DEFAULT 1,
    previous_version_id UUID REFERENCES agent_definitions(id),

    CONSTRAINT check_image_repository CHECK (image_repository != ''),
    CONSTRAINT check_image_tag CHECK (image_tag != '')
    );

-- Indexes for agent definitions
CREATE INDEX IF NOT EXISTS idx_agent_definitions_type_active
    ON agent_definitions(type, is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_agent_definitions_category
    ON agent_definitions(category) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_agent_definitions_version
    ON agent_definitions(type, version);

-- Add comments for documentation
COMMENT ON TABLE agent_definitions IS 'Single source of truth for all agent configurations including deployment specs';
COMMENT ON COLUMN agent_definitions.image_repository IS 'Full repository path for the agent Docker image';
COMMENT ON COLUMN agent_definitions.image_tag IS 'Docker image tag/version';
COMMENT ON COLUMN agent_definitions.command IS 'Container entrypoint command as array';
COMMENT ON COLUMN agent_definitions.resources IS 'Kubernetes resource requests and limits';
COMMENT ON COLUMN agent_definitions.topics IS 'Kafka topic patterns for this agent type';
COMMENT ON COLUMN agent_definitions.health_config IS 'Health check endpoints and configuration';
COMMENT ON COLUMN agent_definitions.env_vars IS 'Additional environment variables for this agent type';

-- Global orchestrator state table (remains the same with minor enhancement)
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

    -- Add agent tracking (NEW)
    spawned_agents JSONB DEFAULT '{}', -- {"role": "agent_id"} mapping

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Indexes for orchestrator state
CREATE INDEX IF NOT EXISTS idx_orchestrator_state_status ON orchestrator_state(status);
CREATE INDEX IF NOT EXISTS idx_orchestrator_state_client ON orchestrator_state(client_id);
CREATE INDEX IF NOT EXISTS idx_orchestrator_state_updated_at ON orchestrator_state(updated_at);
CREATE INDEX IF NOT EXISTS idx_orchestrator_state_client_status ON orchestrator_state(client_id, status);