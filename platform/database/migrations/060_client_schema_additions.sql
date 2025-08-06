-- Update the create_client_schema function to include new tables
CREATE OR REPLACE FUNCTION create_client_schema(client_id TEXT)
RETURNS VOID AS $$
DECLARE
schema_name TEXT := 'client_' || client_id;
BEGIN
    -- ... existing tables ...

    -- Website projects table
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.website_projects (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            project_id UUID REFERENCES %I.projects(id),
            domain VARCHAR(255) NOT NULL,
            business_name VARCHAR(255),
            business_type VARCHAR(100),

            -- Site configuration
            site_config JSONB NOT NULL DEFAULT ''{}''::jsonb,
            site_structure JSONB,

            -- Storage info
            bucket_path VARCHAR(500),
            preview_url VARCHAR(500),
            live_url VARCHAR(500),

            -- Status tracking
            status VARCHAR(50) NOT NULL DEFAULT ''planning'',
            build_progress JSONB DEFAULT ''{}''::jsonb,

            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            published_at TIMESTAMP WITH TIME ZONE
        )', schema_name, schema_name);

-- Agent spawn history
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.agent_spawn_history (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            parent_agent_id UUID,
            spawned_agent_id UUID NOT NULL,
            spawn_reason TEXT,
            configuration_diff JSONB,
            performance_metrics JSONB,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )', schema_name);

END;
$$ LANGUAGE plpgsql;