-- test/migrations/003_human_tasks.sql
-- Create human tasks table for testing human-in-the-loop

CREATE TABLE IF NOT EXISTS human_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    correlation_id VARCHAR(255) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    task_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    assigned_to VARCHAR(255),
    required_role VARCHAR(100),
    task_data JSONB NOT NULL,
    response_data JSONB,
    timeout_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    assigned_at TIMESTAMP,
    completed_at TIMESTAMP,

    CONSTRAINT fk_correlation
    FOREIGN KEY (correlation_id)
    REFERENCES orchestrator_state(correlation_id)
    );

CREATE INDEX idx_human_tasks_correlation ON human_tasks(correlation_id);
CREATE INDEX idx_human_tasks_status ON human_tasks(status);
CREATE INDEX idx_human_tasks_assigned ON human_tasks(assigned_to);
CREATE INDEX idx_human_tasks_timeout ON human_tasks(timeout_at);

-- Create notifications table
CREATE TABLE IF NOT EXISTS human_task_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL,
    notification_type VARCHAR(50) NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    sent_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_task
    FOREIGN KEY (task_id)
    REFERENCES human_tasks(id)
    );


CREATE TABLE approval_requests (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       request_id VARCHAR(255) UNIQUE NOT NULL,
       orchestration_id VARCHAR(255) NOT NULL,
       correlation_id VARCHAR(255) NOT NULL,
       agent_type VARCHAR(100),
       agent_id VARCHAR(255),
       step_name VARCHAR(255),
       data JSONB,
       status VARCHAR(50) DEFAULT 'pending',
       approved_by VARCHAR(255),
       comments TEXT,
       created_at TIMESTAMP DEFAULT NOW(),
       updated_at TIMESTAMP DEFAULT NOW(),
       approved_at TIMESTAMP
);