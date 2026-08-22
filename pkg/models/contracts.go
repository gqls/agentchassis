// FILE: pkg/models/contracts.go (updated)
package models

import "time"

// AgentConfig defines the "mind" of an agent, loaded from the database
type AgentConfig struct {
	AgentID      string                 `json:"agent_id"`
	AgentType    string                 `json:"agent_type"`
	Version      int                    `json:"version"`
	CoreLogic    map[string]interface{} `json:"core_logic"`
	Workflow     WorkflowPlan           `json:"workflow"`
	MemoryConfig MemoryConfiguration    `json:"memory_config,omitempty"`
	Actions      []ActionConfig         `json:"actions,omitempty"`
}

// ActionConfig represents a single action configuration
type ActionConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Handler     string                 `json:"handler,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// MemoryConfiguration controls how the agent uses long-term memory
type MemoryConfiguration struct {
	Enabled            bool     `json:"enabled"`
	AutoStore          bool     `json:"auto_store"`
	AutoStoreThreshold float64  `json:"auto_store_threshold"`
	MaxMemories        int      `json:"max_memories"`
	RetrievalCount     int      `json:"retrieval_count"`
	EmbeddingModel     string   `json:"embedding_model"`
	IncludeTypes       []string `json:"include_types"`
}

// MemoryEntry represents a single memory to be stored
type MemoryEntry struct {
	Content   string                 `json:"content"`
	Type      string                 `json:"type"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// WorkflowPlan defines the orchestration steps for an agent
type WorkflowPlan struct {
	StartStep      string          `json:"start_step"`
	Steps          map[string]Step `json:"steps"`
	TimeoutSeconds int             `json:"timeout_seconds,omitempty"`

	// AwaitReconcileEnforce makes the advance decision ADOPT the awaited_requests
	// table's view when it disagrees with the in-memory AwaitedRequests map:
	// instead of advancing on "the map is empty", the orchestration re-parks on the
	// rows the table still shows outstanding.
	//
	// Default false = today's behaviour exactly. Detection of the divergence is
	// UNCONDITIONAL and logs either way; only the decision is gated. bug 343 (silent post-abandonment freeze):
	// two representations of "what is outstanding" with nothing reconciling them is
	// the wedge's substrate, and enforcement is the half that changes what runs, so
	// it ships opt-in with the unsafe side off (owner ruling 2026-08-02 §2).
	//
	// Two hazards to weigh before switching it on for a workflow: adopting a row
	// means waiting out its timeout, and the retry driver then REPLAYS the request,
	// re-running real side effects; and the table must stay a cross-check, never a
	// naive replacement for the map's optimistic-lock CAS, which is what serialises
	// two pods racing to advance.
	AwaitReconcileEnforce bool `json:"await_reconcile_enforce,omitempty"`
}

// Step represents a single action or sub-workflow within a plan
type Step struct {
	Action          string                 `json:"action"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Topic           string                 `json:"topic,omitempty"`
	TargetAgentType string                 `json:"target_agent_type"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	NextStep        string                 `json:"next_step,omitempty"`
	ErrorStep       string                 `json:"error_step,omitempty"`
	OutputField     string                 `json:"output_field,omitempty"`
	SubTasks        []SubTask              `json:"sub_tasks,omitempty"`
	StoreMemory     bool                   `json:"store_memory,omitempty"`
	Config          map[string]interface{} `json:"config,omitempty"`
	Timeout         time.Duration          `json:"timeout,omitempty"`
}

// SubTask for fan-out operations
type SubTask struct {
	StepName string `json:"step_name"`
	Topic    string `json:"topic"`
}

// Standard message payloads
type TaskRequest struct {
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data"`
}

type TaskResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error,omitempty"`
}
