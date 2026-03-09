see document 017_creating_new_client_schemas.md

-- Drop the wrong tables
DROP TABLE IF EXISTS client_system.agent_spawn_history CASCADE;
DROP TABLE IF EXISTS client_system.agent_instances CASCADE;

-- Recreate matching what create_client_schema and spawn_agent expect
CREATE TABLE client_system.agent_instances (
                                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                               template_id UUID REFERENCES agent_definitions(id),
                                               project_id UUID,
                                               owner_user_id VARCHAR(255) NOT NULL DEFAULT 'system',
                                               name VARCHAR(255) NOT NULL,
                                               config JSONB NOT NULL DEFAULT '{}'::jsonb,
                                               is_active BOOLEAN DEFAULT true,
                                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                               updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                               deleted_at TIMESTAMPTZ,
                                               CONSTRAINT fk_project FOREIGN KEY (project_id)
                                                   REFERENCES client_system.projects(id) ON DELETE CASCADE
);

CREATE TABLE client_system.agent_spawn_history (
                                                   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                                   parent_agent_id UUID,
                                                   spawned_agent_id UUID NOT NULL,
                                                   spawn_reason TEXT,
                                                   spawn_config JSONB DEFAULT '{}'::jsonb,
                                                   created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_system_agent_instances_active
    ON client_system.agent_instances(is_active) WHERE deleted_at IS NULL;

-- Verify columns match what spawn_agent expects:
-- (id, template_id, owner_user_id, name, config, is_active, created_at, updated_at)
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'client_system' AND table_name = 'agent_instances'
ORDER BY ordinal_position;


