-- FILE: platform/database/migrations/003_create_client_schema.sql
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
