-- ============================================================================
-- agent_error_log — persistent, queryable record of every agent error
--
-- Captures errors from two sources:
--   1. routeToErrorStep — workflow step failed, routing to error handler
--   2. notifyParentOfFailure — entire workflow failed
--
-- This replaces digging through kubectl logs to find error details.
-- ============================================================================

CREATE TABLE IF NOT EXISTS agent_error_log (
                                               id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    occurred_at      timestamptz NOT NULL DEFAULT NOW(),

    -- What failed
    site_id          uuid,
    domain           text,
    work_item_id     uuid,

    -- Where it failed
    orchestration_id text,
    agent_type       text NOT NULL,
    agent_id         text,
    pod_name         text,
    step_name        text,
    action           text,

    -- The error
    error_message    text NOT NULL,
    error_code       text,           -- CHILD_ORCHESTRATION_FAILED, TEMPLATE_ERROR, VALIDATION_FAILED, etc.
    severity         text NOT NULL DEFAULT 'error',  -- error, warning, fatal

-- Context snapshot (spec, input_data, whatever helps debugging)
    context          jsonb DEFAULT '{}'::jsonb,

    -- Resolution tracking
    resolved         boolean NOT NULL DEFAULT false,
    resolved_at      timestamptz,
    resolved_by      text
    );

-- Recent errors (dashboard default view)
CREATE INDEX idx_error_log_time ON agent_error_log (occurred_at DESC);

-- Errors per site
CREATE INDEX idx_error_log_site ON agent_error_log (site_id, occurred_at DESC);

-- Unresolved errors
CREATE INDEX idx_error_log_unresolved ON agent_error_log (resolved, occurred_at DESC)
    WHERE resolved = false;

-- Error frequency by agent type
CREATE INDEX idx_error_log_agent ON agent_error_log (agent_type, occurred_at DESC);