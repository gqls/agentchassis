-- 060_client_schema_additions.sql
-- Reusable function for creating client-specific schemas with all necessary tables

-- Drop function if exists to avoid conflicts
DROP FUNCTION IF EXISTS create_client_schema(TEXT);

-- Function to create client-specific schemas
CREATE OR REPLACE FUNCTION create_client_schema(client_id TEXT)
RETURNS VOID AS $$
DECLARE
schema_name TEXT := 'client_' || client_id;
BEGIN
    -- Validate client_id
    IF client_id IS NULL OR client_id = '' THEN
        RAISE EXCEPTION 'Client ID cannot be null or empty';
END IF;

    -- Create the schema
EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);
RAISE NOTICE 'Schema % created or already exists', schema_name;

    -- Create projects table
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.projects (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(255) NOT NULL,
            description TEXT,
            owner_user_id VARCHAR(255) NOT NULL,
            settings JSONB DEFAULT ''{}''::jsonb,
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            deleted_at TIMESTAMPTZ
        )', schema_name);

-- Create agent_instances table
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.agent_instances (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            template_id UUID REFERENCES agent_definitions(id),
            project_id UUID,
            owner_user_id VARCHAR(255) NOT NULL,
            name VARCHAR(255) NOT NULL,
            config JSONB NOT NULL DEFAULT ''{}''::jsonb,
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            deleted_at TIMESTAMPTZ,
            CONSTRAINT fk_project FOREIGN KEY (project_id)
                REFERENCES %I.projects(id) ON DELETE CASCADE
        )', schema_name, schema_name);

-- Create website_projects table
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.website_projects (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            project_id UUID NOT NULL REFERENCES %I.projects(id) ON DELETE CASCADE,
            domain VARCHAR(255) NOT NULL,
            business_name VARCHAR(255),
            business_type VARCHAR(100),
            status VARCHAR(50) NOT NULL DEFAULT ''planning'',
            site_structure JSONB,
            content_data JSONB,
            visual_assets JSONB,
            preview_url VARCHAR(500),
            live_url VARCHAR(500),
            s3_bucket VARCHAR(255),
            cloudfront_distribution_id VARCHAR(255),
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            published_at TIMESTAMPTZ,
            CONSTRAINT valid_status CHECK (status IN (
                ''planning'', ''researching'', ''designing'',
                ''developing'', ''reviewing'', ''published'', ''archived''
            ))
        )', schema_name, schema_name);

-- Create agent_spawn_history table
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.agent_spawn_history (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            parent_agent_id UUID,
            spawned_agent_id UUID NOT NULL,
            group_id UUID REFERENCES agent_groups(id),
            spawn_reason TEXT,
            spawn_config JSONB DEFAULT ''{}''::jsonb,
            created_at TIMESTAMPTZ DEFAULT NOW()
        )', schema_name);

-- Create agent_memory table (for vector embeddings)
-- Note: Requires pgvector extension
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.agent_memory (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            agent_instance_id UUID NOT NULL,
            memory_type VARCHAR(50) DEFAULT ''general'',
            content TEXT NOT NULL,
            embedding vector(1536),
            metadata JSONB DEFAULT ''{}''::jsonb,
            relevance_score FLOAT,
            access_count INTEGER DEFAULT 0,
            last_accessed_at TIMESTAMPTZ,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            expires_at TIMESTAMPTZ,
            CONSTRAINT fk_agent_instance FOREIGN KEY (agent_instance_id)
                REFERENCES %I.agent_instances(id) ON DELETE CASCADE
        )', schema_name, schema_name);

-- Create workflow_executions table for tracking
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.workflow_executions (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            correlation_id UUID NOT NULL,
            agent_instance_id UUID,
            workflow_name VARCHAR(255),
            status VARCHAR(50) DEFAULT ''running'',
            input_data JSONB,
            output_data JSONB,
            error_message TEXT,
            started_at TIMESTAMPTZ DEFAULT NOW(),
            completed_at TIMESTAMPTZ,
            duration_ms INTEGER GENERATED ALWAYS AS (
                CASE
                    WHEN completed_at IS NOT NULL
                    THEN EXTRACT(EPOCH FROM (completed_at - started_at)) * 1000
                    ELSE NULL
                END
            ) STORED,
            CONSTRAINT fk_agent_instance FOREIGN KEY (agent_instance_id)
                REFERENCES %I.agent_instances(id) ON DELETE SET NULL
        )', schema_name, schema_name);

-- Create indexes for better performance
EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_agent_instances_active
        ON %I.agent_instances(is_active) WHERE deleted_at IS NULL', schema_name, schema_name);

EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_agent_instances_owner
        ON %I.agent_instances(owner_user_id)', schema_name, schema_name);

EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_projects_active
        ON %I.projects(is_active) WHERE deleted_at IS NULL', schema_name, schema_name);

EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_projects_owner
        ON %I.projects(owner_user_id)', schema_name, schema_name);

EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_website_projects_status
        ON %I.website_projects(status)', schema_name, schema_name);

EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_website_projects_domain
        ON %I.website_projects(domain)', schema_name, schema_name);

EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_agent_spawn_history_parent
        ON %I.agent_spawn_history(parent_agent_id)', schema_name, schema_name);

EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_agent_spawn_history_spawned
        ON %I.agent_spawn_history(spawned_agent_id)', schema_name, schema_name);

EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_agent_memory_instance
        ON %I.agent_memory(agent_instance_id)', schema_name, schema_name);

EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_agent_memory_type
        ON %I.agent_memory(memory_type)', schema_name, schema_name);

EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_workflow_executions_correlation
        ON %I.workflow_executions(correlation_id)', schema_name, schema_name);

EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_workflow_executions_status
        ON %I.workflow_executions(status)', schema_name, schema_name);

-- Add table comments
EXECUTE format('COMMENT ON TABLE %I.projects IS ''Projects for client %s''', schema_name, client_id);
EXECUTE format('COMMENT ON TABLE %I.agent_instances IS ''Agent instances for client %s''', schema_name, client_id);
EXECUTE format('COMMENT ON TABLE %I.website_projects IS ''Website projects for client %s''', schema_name, client_id);
EXECUTE format('COMMENT ON TABLE %I.agent_memory IS ''Agent memory storage for client %s''', schema_name, client_id);

RAISE NOTICE 'All tables and indexes created successfully for client %', client_id;

EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Failed to create schema for client %: %', client_id, SQLERRM;
END;
$$ LANGUAGE plpgsql;

-- Add function comment
COMMENT ON FUNCTION create_client_schema(TEXT) IS 'Creates a complete schema with all necessary tables for a new client';

-- Note: The demo_client schema was already created in 006_client_schema.sql
-- This function is for creating additional client schemas dynamically