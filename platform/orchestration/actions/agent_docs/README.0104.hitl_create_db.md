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