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
	StartStep string          `json:"start_step"`
	Steps     map[string]Step `json:"steps"`
}

// Step represents a single action or sub-workflow within a plan
type Step struct {
	Action       string    `json:"action"`
	Description  string    `json:"description"`
	Topic        string    `json:"topic,omitempty"`
	Dependencies []string  `json:"dependencies,omitempty"`
	NextStep     string    `json:"next_step,omitempty"`
	SubTasks     []SubTask `json:"sub_tasks,omitempty"`
	StoreMemory  bool      `json:"store_memory,omitempty"` // New field
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
