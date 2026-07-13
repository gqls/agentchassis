
-- Track request-response pairs
CREATE TABLE orchestration_requests (
                                        request_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                        orchestration_id UUID NOT NULL,
                                        from_agent_id UUID NOT NULL,
                                        to_agent_id UUID NOT NULL,
                                        message_type VARCHAR(50) DEFAULT 'request',
                                        in_response_to UUID,
                                        status VARCHAR(50) DEFAULT 'pending',
                                        timeout_at TIMESTAMP,
                                        created_at TIMESTAMP DEFAULT NOW(),
                                        completed_at TIMESTAMP,

                                        CONSTRAINT fk_orch FOREIGN KEY (orchestration_id)
                                            REFERENCES orchestration_states(orchestration_id)
);

-- Indexes for orchestration_requests
CREATE INDEX idx_orch_requests ON orchestration_requests (orchestration_id, status);
CREATE INDEX idx_response_tracking ON orchestration_requests (in_response_to);

-- Create new table with proper ownership model
CREATE TABLE IF NOT EXISTS orchestration_states (
                                                    orchestration_id UUID PRIMARY KEY,
                                                    correlation_id UUID NOT NULL,
                                                    owner_agent_id UUID NOT NULL,
                                                    parent_orch_id UUID,
                                                    client_id VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    current_step VARCHAR(255) NOT NULL,
    awaited_steps JSONB DEFAULT '[]'::jsonb,
    collected_data JSONB DEFAULT '{}'::jsonb,
    initial_request_data JSONB,
    final_result JSONB,
    error TEXT,
    workflow_plan JSONB NOT NULL,
    execution_metadata JSONB DEFAULT '{}'::jsonb,
    execution_path JSONB DEFAULT '[]'::jsonb,
    version INT DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure ownership integrity
    CONSTRAINT uk_orchestration_owner UNIQUE (orchestration_id, owner_agent_id)
    );

-- Indexes for orchestration_states
CREATE INDEX idx_correlation ON orchestration_states (correlation_id);
CREATE INDEX idx_owner ON orchestration_states (owner_agent_id, status);
CREATE INDEX idx_parent ON orchestration_states (parent_orch_id);
CREATE INDEX idx_client ON orchestration_states (client_id);
CREATE INDEX idx_status ON orchestration_states (status);
CREATE INDEX idx_updated ON orchestration_states (updated_at);

-- Create pending requests table for tracking async operations
CREATE TABLE IF NOT EXISTS pending_requests (
                                                request_id UUID PRIMARY KEY,
                                                orchestration_id UUID NOT NULL,
                                                to_agent_id UUID NOT NULL,
                                                status VARCHAR(50) DEFAULT 'pending',
    timeout_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,

    CONSTRAINT fk_pending_orch FOREIGN KEY (orchestration_id)
    REFERENCES orchestration_states(orchestration_id) ON DELETE CASCADE
    );

-- Indexes for pending_requests
CREATE INDEX idx_orchestration ON pending_requests (orchestration_id);
CREATE INDEX idx_timeout ON pending_requests (timeout_at);

-- Create agent groups table for tracking agent groups
CREATE TABLE IF NOT EXISTS agent_groups (
                                            group_id UUID PRIMARY KEY,
                                            parent_orch_id UUID NOT NULL,
                                            group_type VARCHAR(100),
    members JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_agent_groups_parent FOREIGN KEY (parent_orch_id)
    REFERENCES orchestration_states(orchestration_id) ON DELETE CASCADE
    );

CREATE TABLE agent_group_definitions (
                                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                         name VARCHAR(255) NOT NULL,
                                         group_type VARCHAR(100) NOT NULL,
                                         agent_configs JSONB NOT NULL,
                                         orchestration_workflow JSONB,
                                         usage_count INTEGER DEFAULT 0,
                                         version INTEGER DEFAULT 1,
                                         created_at TIMESTAMP DEFAULT NOW(),
                                         updated_at TIMESTAMP DEFAULT NOW()
);

-- Add some sample data
INSERT INTO agent_group_definitions (name, group_type, agent_configs, orchestration_workflow)
VALUES (
           'Website Builder Team',
           'website-builder',
           '[
             {"role": "orchestrator", "agent_type": "website-builder"},
             {"role": "architect", "agent_type": "site-architect"},
             {"role": "developer", "agent_type": "html-developer"}
           ]'::jsonb,
           '{
             "start_step": "init",
             "steps": {
               "init": {
                 "action": "validate_input",
                 "next_step": "complete"
               },
               "complete": {
                 "action": "complete_workflow"
               }
             }
           }'::jsonb
       );