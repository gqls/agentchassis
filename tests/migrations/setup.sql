-- test/migrations/setup.sql

-- Create test client schema
CREATE SCHEMA IF NOT EXISTS client_demo_client;

-- Create agent_instances table
CREATE TABLE IF NOT EXISTS client_demo_client.agent_instances (
                                                                  id UUID PRIMARY KEY,
                                                                  template_id UUID NOT NULL,
                                                                  owner_user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
    );

-- Create agent_groups table
CREATE TABLE IF NOT EXISTS agent_groups (
                                            id UUID PRIMARY KEY,
                                            name VARCHAR(255) NOT NULL,
    group_type VARCHAR(100) NOT NULL,
    version VARCHAR(20) DEFAULT '1.0.0',
    parent_id UUID,
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

-- Create orchestrator_state table
CREATE TABLE IF NOT EXISTS orchestrator_state (
                                                  correlation_id VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    current_step VARCHAR(255),
    workflow_plan JSONB NOT NULL,
    collected_data JSONB DEFAULT '{}',
    execution_metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
    );

-- Insert test agent group
INSERT INTO agent_groups (id, name, group_type, agent_configs, orchestration_workflow)
VALUES (
           gen_random_uuid(),
           'Website Builder Team',
           'website-builder',
           '[
               {"role": "architect", "agent_type": "site-architect"},
               {"role": "designer", "agent_type": "visual-designer"},
               {"role": "developer", "agent_type": "html-developer"}
           ]'::jsonb,
           '{
               "start_step": "plan",
               "steps": {
                   "plan": {"action": "create_site_plan", "next_step": "design"},
                   "design": {"action": "create_design", "next_step": "develop"},
                   "develop": {"action": "build_html", "next_step": "complete"},
                   "complete": {"action": "complete_workflow"}
               }
           }'::jsonb
       );