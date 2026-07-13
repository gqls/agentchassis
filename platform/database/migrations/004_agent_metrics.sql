-- Agent performance metrics for evolution decisions
CREATE TABLE IF NOT EXISTS agent_metrics (
                                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL,
    agent_type VARCHAR(100) NOT NULL,

    -- Performance metrics
    total_tasks INTEGER DEFAULT 0,
    successful_tasks INTEGER DEFAULT 0,
    failed_tasks INTEGER DEFAULT 0,
    success_rate DECIMAL(5,4) GENERATED ALWAYS AS
(CASE WHEN total_tasks > 0 THEN successful_tasks::DECIMAL / total_tasks ELSE 0 END) STORED,

    -- Timing metrics
    total_execution_time_ms BIGINT DEFAULT 0,
    avg_response_time_ms INTEGER GENERATED ALWAYS AS
(CASE WHEN total_tasks > 0 THEN total_execution_time_ms / total_tasks ELSE 0 END) STORED,

    -- Resource usage
    total_fuel_consumed INTEGER DEFAULT 0,
    avg_fuel_per_task INTEGER GENERATED ALWAYS AS
(CASE WHEN total_tasks > 0 THEN total_fuel_consumed / total_tasks ELSE 0 END) STORED,

    -- Quality metrics (from human feedback)
    quality_ratings JSONB DEFAULT '[]'::jsonb,
    avg_quality_score DECIMAL(3,2),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

CREATE INDEX idx_agent_metrics_agent ON agent_metrics(agent_id);
CREATE INDEX idx_agent_metrics_type ON agent_metrics(agent_type);
CREATE INDEX idx_agent_metrics_performance ON agent_metrics(success_rate DESC, avg_response_time_ms ASC);