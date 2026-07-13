-- https://claude.ai/chat/3606e9df-4929-4b45-a12d-8ead94f9f65d
-- Create audit table for tracking state changes
CREATE TABLE IF NOT EXISTS orchestration_state_audit (
                                                         id SERIAL PRIMARY KEY,
                                                         orchestration_id UUID NOT NULL,
                                                         old_version INT,
                                                         new_version INT,
                                                         old_status TEXT,
                                                         new_status TEXT,
                                                         old_current_step TEXT,
                                                         new_current_step TEXT,
                                                         changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    pg_backend_pid INT DEFAULT pg_backend_pid(),
    application_name TEXT DEFAULT current_setting('application_name', true)
    );

-- Index for querying by orchestration_id and time
CREATE INDEX IF NOT EXISTS idx_audit_orch_id_time
    ON orchestration_state_audit(orchestration_id, changed_at DESC);

-- Create trigger function
CREATE OR REPLACE FUNCTION log_orchestration_state_changes()
RETURNS TRIGGER AS $$
BEGIN
INSERT INTO orchestration_state_audit (
    orchestration_id,
    old_version,
    new_version,
    old_status,
    new_status,
    old_current_step,
    new_current_step
) VALUES (
             NEW.orchestration_id,
             OLD.version,
             NEW.version,
             OLD.status,
             NEW.status,
             OLD.current_step,
             NEW.current_step
         );
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if present (for re-runs)
DROP TRIGGER IF EXISTS orchestration_state_audit_trigger ON orchestration_states;

-- Attach trigger
CREATE TRIGGER orchestration_state_audit_trigger
    AFTER UPDATE ON orchestration_states
    FOR EACH ROW
    EXECUTE FUNCTION log_orchestration_state_changes();

--
=======
-- to analyse afterwards


-- See all version changes for a specific orchestration
SELECT
    old_version,
    new_version,
    old_status,
    new_status,
    old_current_step,
    new_current_step,
    changed_at,
    pg_backend_pid,
    application_name,
    -- Time since previous update
    changed_at - LAG(changed_at) OVER (ORDER BY changed_at) as time_since_prev
FROM orchestration_state_audit
WHERE orchestration_id = 'YOUR-ORCHESTRATION-ID-HERE'
ORDER BY changed_at ASC;


====

to clean it update -- Remove trigger when done investigating


DROP TRIGGER IF EXISTS orchestration_state_audit_trigger ON orchestration_states;

-- Optionally keep the audit table for reference, or drop it:
-- DROP TABLE orchestration_state_audit;