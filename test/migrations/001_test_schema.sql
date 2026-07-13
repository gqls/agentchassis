-- test/migrations/001_test_schema.sql
-- Test schema setup for agent spawning tests

-- Ensure test client schema exists
CREATE SCHEMA IF NOT EXISTS client_test_client;
CREATE SCHEMA IF NOT EXISTS client_demo_client;

-- Create agent_instances table
CREATE TABLE IF NOT EXISTS client_demo_client.agent_instances (
                                                                  id UUID PRIMARY KEY,
                                                                  template_id UUID NOT NULL,
                                                                  owner_user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL,
    is_active BOOLEAN DEFAULT true,
    last_heartbeat TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
    );

-- Create agent_groups table
CREATE TABLE IF NOT EXISTS agent_groups (
                                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    group_type VARCHAR(100) NOT NULL,
    version VARCHAR(20) DEFAULT '1.0.0',
    parent_id UUID,
    version VARCHAR(20) DEFAULT '1.0.0',
    agent_configs JSONB NOT NULL,
    orchestration_workflow JSONB NOT NULL,
    capabilities JSONB DEFAULT '[]',
    usage_count INTEGER DEFAULT 0,
    performance_metrics JSONB DEFAULT '{}',
    mutation_history JSONB DEFAULT '[]',
    last_used_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
    );

-- Create orchestrator_state table if not exists
CREATE TABLE IF NOT EXISTS orchestrator_state (
                                                  correlation_id VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    current_step VARCHAR(255),
    workflow_plan JSONB NOT NULL,
    collected_data JSONB DEFAULT '{}',
    execution_metadata JSONB DEFAULT '{}',
    execution_path JSONB DEFAULT '[]',
    awaited_steps TEXT[],
    initial_request_data JSONB,
    final_result JSONB,
    error TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
    );

-- Create agent_metrics table
CREATE TABLE IF NOT EXISTS agent_metrics (
                                             agent_id UUID NOT NULL,
                                             success_rate FLOAT DEFAULT 0.5,
                                             avg_response_time INTEGER,
                                             total_requests INTEGER DEFAULT 0,
                                             failed_requests INTEGER DEFAULT 0,
                                             last_updated TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (agent_id)
    );

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_agent_instances_type ON client_demo_client.agent_instances((config->>'agent_type'));
CREATE INDEX IF NOT EXISTS idx_agent_instances_active ON client_demo_client.agent_instances(is_active);
CREATE INDEX IF NOT EXISTS idx_orchestrator_state_client ON orchestrator_state(client_id);
CREATE INDEX IF NOT EXISTS idx_orchestrator_state_status ON orchestrator_state(status);
CREATE INDEX IF NOT EXISTS idx_agent_groups_type ON agent_groups(group_type);