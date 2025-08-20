-- 006_client_schema.sql
-- Creates the initial client schema and inserts the generic orchestrator instance

-- Ensure the client schema exists
CREATE SCHEMA IF NOT EXISTS client_demo_client;

-- Create the agent_instances table for demo client
CREATE TABLE IF NOT EXISTS client_demo_client.agent_instances (
                                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID,
    owner_user_id VARCHAR(255),
    name VARCHAR(255),
    config JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
    );

-- Create index for active instances
CREATE INDEX IF NOT EXISTS idx_demo_client_agent_instances_active
    ON client_demo_client.agent_instances(is_active);

-- Insert the generic orchestrator instance for demo_client
INSERT INTO client_demo_client.agent_instances (
    id,
    template_id,
    owner_user_id,
    name,
    config,
    is_active
) VALUES (
             '00000000-0000-0000-0000-000000000001'::uuid,
             '00000000-0000-0000-0000-000000000001'::uuid,
             'system',
             'Generic Orchestrator Instance',
             '{
                 "agent_type": "generic",
                 "topic": "system.agent.generic.process",
                 "capabilities": ["orchestration", "spawn_group"]
             }'::jsonb,
             true
         )
    ON CONFLICT (id) DO UPDATE
                            SET config = EXCLUDED.config,
                            updated_at = NOW();