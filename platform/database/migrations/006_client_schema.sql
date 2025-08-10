
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
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
