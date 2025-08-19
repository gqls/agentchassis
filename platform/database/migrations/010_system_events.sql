-- 003_system_events.sql
-- System events table for audit logging and tracking agent activities

-- Create enum for event types if you want strict typing (optional)
DO $$ BEGIN
CREATE TYPE event_type_enum AS ENUM (
        'agent_bootstrap',
        'agent_shutdown',
        'agent_error',
        'workflow_started',
        'workflow_completed',
        'workflow_failed',
        'client_created',
        'client_updated',
        'api_call',
        'system_error',
        'configuration_change'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create enum for entity types (optional)
DO $$ BEGIN
CREATE TYPE entity_type_enum AS ENUM (
        'agent',
        'workflow',
        'client',
        'user',
        'system',
        'api'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create the system_events table
CREATE TABLE IF NOT EXISTS system_events (
                                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL, -- or use event_type_enum if you want strict typing
    entity_type VARCHAR(50) NOT NULL, -- or use entity_type_enum if you want strict typing
    entity_id VARCHAR(255) NOT NULL,  -- Can be UUID, string ID, etc.
    client_id VARCHAR(100),            -- Optional: track which client this event belongs to
    user_id VARCHAR(255),              -- Optional: track which user triggered this
    metadata JSONB DEFAULT '{}',       -- Flexible JSON field for event-specific data
    severity VARCHAR(20) DEFAULT 'info' CHECK (severity IN ('debug', 'info', 'warning', 'error', 'critical')),
    source VARCHAR(255),               -- Which service/component generated this event
    ip_address INET,                   -- Optional: track request IP
    user_agent TEXT,                   -- Optional: track user agent
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- No updated_at as events are immutable

    -- Add constraints
    CONSTRAINT check_event_type CHECK (event_type != ''),
    CONSTRAINT check_entity_type CHECK (entity_type != ''),
    CONSTRAINT check_entity_id CHECK (entity_id != '')
    );

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_system_events_event_type
    ON system_events(event_type);

CREATE INDEX IF NOT EXISTS idx_system_events_entity
    ON system_events(entity_type, entity_id);

CREATE INDEX IF NOT EXISTS idx_system_events_client_id
    ON system_events(client_id)
    WHERE client_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_system_events_created_at
    ON system_events(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_system_events_severity
    ON system_events(severity)
    WHERE severity IN ('error', 'critical');

-- Composite index for common query patterns
CREATE INDEX IF NOT EXISTS idx_system_events_entity_time
    ON system_events(entity_type, entity_id, created_at DESC);

-- GIN index for JSONB metadata queries
CREATE INDEX IF NOT EXISTS idx_system_events_metadata
    ON system_events USING GIN (metadata);

-- Create a function to auto-delete old events (optional)
CREATE OR REPLACE FUNCTION cleanup_old_system_events()
RETURNS void AS $$
BEGIN
    -- Delete events older than 90 days, keeping critical events for 1 year
DELETE FROM system_events
WHERE created_at < NOW() - INTERVAL '90 days'
  AND severity NOT IN ('error', 'critical');

DELETE FROM system_events
WHERE created_at < NOW() - INTERVAL '365 days';
END;
$$ LANGUAGE plpgsql;

-- Create a view for recent agent events (helpful for monitoring)
CREATE OR REPLACE VIEW recent_agent_events AS
SELECT
    id,
    event_type,
    entity_id as agent_id,
    client_id,
    metadata,
    severity,
    created_at
FROM system_events
WHERE entity_type = 'agent'
  AND created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;

-- Create a materialized view for event statistics (optional, for dashboards)
CREATE MATERIALIZED VIEW IF NOT EXISTS event_statistics AS
SELECT
    date_trunc('hour', created_at) as hour,
    event_type,
    entity_type,
    severity,
    COUNT(*) as event_count,
    COUNT(DISTINCT entity_id) as unique_entities,
    COUNT(DISTINCT client_id) as unique_clients
FROM system_events
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY 1, 2, 3, 4;

-- Create index on materialized view
CREATE INDEX IF NOT EXISTS idx_event_statistics_hour
    ON event_statistics(hour DESC);

-- Add a comment to the table
COMMENT ON TABLE system_events IS 'Audit log for all system events including agent bootstrapping, workflows, and API calls';
COMMENT ON COLUMN system_events.metadata IS 'Flexible JSON field containing event-specific data like agent_type, request details, etc.';

-- Sample insert statement (for documentation)
/*
INSERT INTO system_events (
    event_type,
    entity_type,
    entity_id,
    client_id,
    metadata,
    severity,
    source
) VALUES (
    'agent_bootstrap',
    'agent',
    'uuid-of-agent',
    'demo_client',
    '{"agent_type": "domain-analyst", "job_name": "agent-domain-analyst-abc123"}',
    'info',
    'core-manager'
);
*/