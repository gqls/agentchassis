-- FILE: platform/database/migrations/003_create_client_schema.sql


-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Global agent definitions table (shared across all clients)
CREATE TABLE IF NOT EXISTS agent_definitions (
                                                 id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL CHECK (category IN ('data-driven', 'code-driven', 'adapter')),
    default_config JSONB NOT NULL DEFAULT '{}',
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

-- Function to create client-specific schema
CREATE OR REPLACE FUNCTION create_client_schema(client_id TEXT)
RETURNS VOID AS $$
DECLARE
schema_name TEXT := 'client_' || client_id;
BEGIN
    -- Create schema
EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);

-- Agent instances table for this client
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.agent_instances (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            template_id UUID NOT NULL,
            owner_user_id VARCHAR(255) NOT NULL,
            name VARCHAR(255) NOT NULL,
            config JSONB NOT NULL DEFAULT ''{}''::jsonb,
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        )', schema_name);

-- Indexes for agent instances
EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_instances_owner
        ON %I.agent_instances(owner_user_id)', schema_name);

EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_instances_template
        ON %I.agent_instances(template_id)', schema_name);

-- Agent memory table with vector support
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.agent_memory (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            agent_instance_id UUID NOT NULL REFERENCES %I.agent_instances(id),
            content TEXT NOT NULL,
            embedding vector(1536) NOT NULL,
            metadata JSONB DEFAULT ''{}''::jsonb,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        )', schema_name, schema_name);

-- Vector index for similarity search
EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_memory_embedding
        ON %I.agent_memory USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)', schema_name);

-- Index for agent memory queries
EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_memory_agent_created
        ON %I.agent_memory(agent_instance_id, created_at DESC)', schema_name);

-- Projects table for this client
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.projects (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            name VARCHAR(255) NOT NULL,
            description TEXT,
            owner_user_id VARCHAR(255) NOT NULL,
            settings JSONB DEFAULT ''{}''::jsonb,
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        )', schema_name);

-- Index for project queries
EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_projects_owner
        ON %I.projects(owner_user_id)', schema_name);

-- Workflow executions table for this client
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.workflow_executions (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            correlation_id UUID NOT NULL,
            project_id UUID REFERENCES %I.projects(id),
            agent_instance_id UUID REFERENCES %I.agent_instances(id),
            status VARCHAR(50) NOT NULL,
            input_data JSONB,
            output_data JSONB,
            error_message TEXT,
            started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            completed_at TIMESTAMPTZ,
            created_by VARCHAR(255) NOT NULL
        )', schema_name, schema_name, schema_name);

-- Indexes for workflow executions
EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_workflow_executions_correlation
        ON %I.workflow_executions(correlation_id)', schema_name);

EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_workflow_executions_status
        ON %I.workflow_executions(status)', schema_name);

EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_workflow_executions_project
        ON %I.workflow_executions(project_id)', schema_name);

-- Usage analytics table for this client
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.usage_analytics (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id VARCHAR(255) NOT NULL,
            agent_type VARCHAR(100) NOT NULL,
            action VARCHAR(100) NOT NULL,
            fuel_consumed INTEGER NOT NULL DEFAULT 0,
            execution_time_ms INTEGER,
            success BOOLEAN NOT NULL,
            metadata JSONB DEFAULT ''{}''::jsonb,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        )', schema_name);

-- Indexes for analytics
EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_usage_analytics_user_date
        ON %I.usage_analytics(user_id, created_at)', schema_name);

EXECUTE format('
        CREATE INDEX IF NOT EXISTS idx_usage_analytics_agent_type
        ON %I.usage_analytics(agent_type, created_at)', schema_name);

END;
$$ LANGUAGE plpgsql;

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

-- Create a demo client schema for testing
SELECT create_client_schema('demo_client');


-- This should be run for each new client
-- Replace {client_id} with actual client ID

CREATE SCHEMA IF NOT EXISTS client_{client_id};

-- Agent instances table
CREATE TABLE IF NOT EXISTS client_{client_id}.agent_instances (
                                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL,
    owner_user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX idx_instances_owner ON client_{client_id}.agent_instances(owner_user_id);
CREATE INDEX idx_instances_template ON client_{client_id}.agent_instances(template_id);

-- Agent memory table with vector support
CREATE TABLE IF NOT EXISTS client_{client_id}.agent_memory (
                                                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_instance_id UUID NOT NULL REFERENCES client_{client_id}.agent_instances(id),
    content TEXT NOT NULL,
    embedding vector(1536) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Create vector index for similarity search
CREATE INDEX ON client_{client_id}.agent_memory USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Orchestrator state table
CREATE TABLE IF NOT EXISTS client_{client_id}.orchestrator_state (
                                                                     correlation_id UUID PRIMARY KEY,
                                                                     status VARCHAR(50) NOT NULL,
    current_step VARCHAR(255) NOT NULL,
    awaited_steps JSONB DEFAULT '[]',
    collected_data JSONB DEFAULT '{}',
    initial_request_data JSONB,
    final_result JSONB,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX idx_orchestrator_status ON client_{client_id}.orchestrator_state(status);

CREATE INDEX idx_memory_agent_created ON client_{client_id}.agent_memory(agent_instance_id, created_at DESC);
