-- This file now contains the complete and final version of the function.
CREATE OR REPLACE FUNCTION create_client_schema(client_id TEXT)
RETURNS VOID AS $$
DECLARE
schema_name TEXT := 'client_' || client_id;
BEGIN
    -- Step 1: Create the schema itself (The missing piece)
EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);

-- Step 2: Create the 'projects' table first, as other tables reference it
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

-- Step 3: Create the 'agent_instances' table
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

-- Step 4: Create other tables that depend on the first two
EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.website_projects (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            project_id UUID REFERENCES %I.projects(id),
            domain VARCHAR(255) NOT NULL,
            business_name VARCHAR(255),
            status VARCHAR(50) NOT NULL DEFAULT ''planning'',
            preview_url VARCHAR(500),
            live_url VARCHAR(500)
        )', schema_name, schema_name);

EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.agent_spawn_history (
             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
             parent_agent_id UUID,
             spawned_agent_id UUID NOT NULL,
             spawn_reason TEXT,
             created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        )', schema_name);

END;
$$ LANGUAGE plpgsql;

SELECT create_client_schema('demo_client');