// FILE: platform/orchestration/types/subtree.go
// Move SubtreeInfo and related types here
package types

import "time"

// SubtreeInfo tracks agents spawned in a subtree, tracks hierarchical agent relationships
type SubtreeInfo struct {
	AgentID       string                  `json:"agent_id"`
	AgentType     string                  `json:"agent_type"`
	AgentName     string                  `json:"agent_name"`
	ParentAgentID string                  `json:"parent_agent_id"`
	Children      map[string]*SubtreeInfo `json:"children"`
	CreatedAt     time.Time               `json:"created_at"`
	LastActiveAt  time.Time               `json:"last_active_at"`
	Performance   *PerformanceMetrics     `json:"performance"`
}

// PerformanceMetrics tracks performance of agents
type PerformanceMetrics struct {
	TasksCompleted   int       `json:"tasks_completed"`
	TasksFailed      int       `json:"tasks_failed"`
	AverageLatencyMs int64     `json:"average_latency_ms"`
	FuelConsumed     int       `json:"fuel_consumed"`
	LastUpdated      time.Time `json:"last_updated"`
}
