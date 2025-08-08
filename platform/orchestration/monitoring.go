// platform/orchestration/monitoring.go

package orchestration

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// WorkflowSummary provides a high-level view of a workflow
type WorkflowSummary struct {
	CorrelationID  string    `json:"correlation_id"`
	ClientID       string    `json:"client_id"`
	WorkflowType   string    `json:"workflow_type"` // from start_step
	CurrentStep    string    `json:"current_step"`
	Status         string    `json:"status"`
	CompletedSteps int       `json:"completed_steps"`
	TotalSteps     int       `json:"total_steps"`
	Progress       float64   `json:"progress"` // percentage
	Duration       string    `json:"duration"` // human readable
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	StuckDuration  string    `json:"stuck_duration,omitempty"`
}

// WorkflowMonitor provides monitoring capabilities
type WorkflowMonitor struct {
	db *sql.DB
}

// NewWorkflowMonitor creates a new monitor instance
func NewWorkflowMonitor(db *sql.DB) *WorkflowMonitor {
	return &WorkflowMonitor{db: db}
}

// GetActiveWorkflows returns all non-completed workflows for a client
func (m *WorkflowMonitor) GetActiveWorkflows(ctx context.Context, clientID string) ([]WorkflowSummary, error) {
	query := `
        SELECT 
            correlation_id,
            client_id,
            workflow_plan->>'start_step' as workflow_type,
            current_step,
            status,
            COALESCE((execution_metadata->>'completed_steps')::int, 0) as completed,
            COALESCE((execution_metadata->>'total_steps')::int, 0) as total,
            created_at,
            updated_at
        FROM orchestrator_state
        WHERE client_id = $1 
        AND status NOT IN ('COMPLETED', 'FAILED')
        ORDER BY updated_at DESC
    `

	rows, err := m.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active workflows: %w", err)
	}
	defer rows.Close()

	var summaries []WorkflowSummary
	for rows.Next() {
		var s WorkflowSummary
		err := rows.Scan(
			&s.CorrelationID,
			&s.ClientID,
			&s.WorkflowType,
			&s.CurrentStep,
			&s.Status,
			&s.CompletedSteps,
			&s.TotalSteps,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}

		// Calculate progress
		if s.TotalSteps > 0 {
			s.Progress = float64(s.CompletedSteps) / float64(s.TotalSteps) * 100
		}

		// Calculate duration
		s.Duration = formatDuration(time.Since(s.CreatedAt))

		summaries = append(summaries, s)
	}

	return summaries, nil
}

// GetStuckWorkflows finds workflows that haven't updated recently
func (m *WorkflowMonitor) GetStuckWorkflows(ctx context.Context, stuckAfter time.Duration) ([]WorkflowSummary, error) {
	cutoffTime := time.Now().Add(-stuckAfter)

	query := `
        SELECT 
            correlation_id,
            client_id,
            workflow_plan->>'start_step' as workflow_type,
            current_step,
            status,
            COALESCE((execution_metadata->>'completed_steps')::int, 0) as completed,
            COALESCE((execution_metadata->>'total_steps')::int, 0) as total,
            created_at,
            updated_at
        FROM orchestrator_state
        WHERE status IN ('RUNNING', 'AWAITING_RESPONSES')
        AND updated_at < $1
        ORDER BY updated_at ASC
    `

	rows, err := m.db.QueryContext(ctx, query, cutoffTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query stuck workflows: %w", err)
	}
	defer rows.Close()

	var summaries []WorkflowSummary
	for rows.Next() {
		var s WorkflowSummary
		err := rows.Scan(
			&s.CorrelationID,
			&s.ClientID,
			&s.WorkflowType,
			&s.CurrentStep,
			&s.Status,
			&s.CompletedSteps,
			&s.TotalSteps,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stuck workflow: %w", err)
		}

		// Calculate how long it's been stuck
		s.StuckDuration = formatDuration(time.Since(s.UpdatedAt))
		s.Duration = formatDuration(time.Since(s.CreatedAt))

		if s.TotalSteps > 0 {
			s.Progress = float64(s.CompletedSteps) / float64(s.TotalSteps) * 100
		}

		summaries = append(summaries, s)
	}

	return summaries, nil
}

// GetWorkflowDetails returns detailed information about a specific workflow
func (m *WorkflowMonitor) GetWorkflowDetails(ctx context.Context, correlationID string) (*OrchestrationState, error) {
	repo := NewStateRepository(m.db, nil)
	return repo.GetState(ctx, correlationID)
}

// GetWorkflowMetrics returns aggregate metrics for workflows
func (m *WorkflowMonitor) GetWorkflowMetrics(ctx context.Context, clientID string, since time.Time) (WorkflowMetrics, error) {
	query := `
        SELECT 
            COUNT(*) as total,
            COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) as completed,
            COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failed,
            COUNT(CASE WHEN status IN ('RUNNING', 'AWAITING_RESPONSES') THEN 1 END) as active,
            COUNT(CASE WHEN status = 'PAUSED_FOR_HUMAN_INPUT' THEN 1 END) as paused,
            COALESCE(AVG(CASE 
                WHEN status = 'COMPLETED' 
                THEN EXTRACT(EPOCH FROM (updated_at - created_at))
            END), 0) as avg_duration_seconds
        FROM orchestrator_state
        WHERE client_id = $1
        AND created_at >= $2
    `

	var metrics WorkflowMetrics
	err := m.db.QueryRowContext(ctx, query, clientID, since).Scan(
		&metrics.TotalWorkflows,
		&metrics.CompletedWorkflows,
		&metrics.FailedWorkflows,
		&metrics.ActiveWorkflows,
		&metrics.PausedWorkflows,
		&metrics.AvgDurationSeconds,
	)

	if err != nil {
		return metrics, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Calculate success rate
	if metrics.TotalWorkflows > 0 {
		metrics.SuccessRate = float64(metrics.CompletedWorkflows) / float64(metrics.TotalWorkflows) * 100
	}

	return metrics, nil
}

// WorkflowMetrics contains aggregate statistics
type WorkflowMetrics struct {
	TotalWorkflows     int     `json:"total_workflows"`
	CompletedWorkflows int     `json:"completed_workflows"`
	FailedWorkflows    int     `json:"failed_workflows"`
	ActiveWorkflows    int     `json:"active_workflows"`
	PausedWorkflows    int     `json:"paused_workflows"`
	SuccessRate        float64 `json:"success_rate"`
	AvgDurationSeconds float64 `json:"avg_duration_seconds"`
}

// Helper function to format duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

// Example usage in your code:
/*
monitor := NewWorkflowMonitor(db)

// Get all active workflows
active, err := monitor.GetActiveWorkflows(ctx, "client_123")

// Find workflows stuck for more than 1 hour
stuck, err := monitor.GetStuckWorkflows(ctx, 1*time.Hour)

// Get metrics for the last 24 hours
metrics, err := monitor.GetWorkflowMetrics(ctx, "client_123", time.Now().Add(-24*time.Hour))
*/
