kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db << 'SQL'
CREATE SCHEMA IF NOT EXISTS client_vetcomparison;

CREATE TABLE IF NOT EXISTS client_vetcomparison.projects (
                                                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_user_id VARCHAR(255) NOT NULL,
    settings JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
    );

CREATE TABLE IF NOT EXISTS client_vetcomparison.agent_instances (
                                                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID,  -- No FK, agent_definitions is in templates_db
    project_id UUID,
    owner_user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_project FOREIGN KEY (project_id)
    REFERENCES client_vetcomparison.projects(id) ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS client_vetcomparison.agent_spawn_history (
                                                                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_agent_id UUID,
    spawned_agent_id UUID NOT NULL,
    spawn_reason TEXT,
    spawn_config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_vetcomparison_agent_instances_active
    ON client_vetcomparison.agent_instances(is_active);
SQL