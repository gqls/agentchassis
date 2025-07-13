// FILE: platform/contracts/contracts.go
// This package defines the core data structures used for agent configuration
// and communication throughout the entire system. It is the "shared language"
// that all services and agents will use.
package contracts

// AgentConfig is the master configuration for a single agent instance.
// This is stored as a JSONB object in the `agent_instances` table and
// loaded by the Agent Chassis at runtime to determine its behavior.
type AgentConfig struct {
	AgentID   string          `json:"agent_id"`
	AgentType string          `json:"agent_type"` // e.g., "copywriter", "orchestrator", "reasoning-agent"
	Version   int             `json:"version"`
	CoreLogic CoreLogicConfig `json:"core_logic"` // The "Who" - parameters for its main skill
	Workflow  WorkflowPlan    `json:"workflow"`   // The "How" - the plan it follows to do its job
}

// CoreLogicConfig is a generic map to hold the specific parameters
// for an agent's primary function.
// For an AI agent, this would contain prompts and model info.
// For an adapter, it might contain API endpoints and credentials.
type CoreLogicConfig map[string]interface{}

// WorkflowPlan defines the orchestration logic for an agent.
// It is a declarative, directed graph of steps.
type WorkflowPlan struct {
	StartStep string          `json:"start_step"`
	Steps     map[string]Step `json:"steps"`
}

// Step represents a single node in the workflow graph. It can be either
// a direct action for the agent to perform, or a call to another agent.
type Step struct {
	// Action is the specific function this agent should perform for this step.
	// e.g., "generate_text", "fan_out", "pause_for_human_input".
	Action string `json:"action"`

	// Description provides a human-readable explanation of the step's purpose.
	Description string `json:"description"`

	// Topic is the Kafka topic to send a message to if this step involves
	// calling another agent or service.
	Topic string `json:"topic,omitempty"`

	// Dependencies lists the `step_name`s that must be completed before
	// this step can begin. The orchestrator will not execute this step
	// until it has received responses from all dependencies.
	Dependencies []string `json:"dependencies,omitempty"`

	// NextStep defines the name of the next step to execute upon successful
	// completion of this one, for simple linear workflows.
	NextStep string `json:"next_step,omitempty"`

	// SubTasks is used for "fan_out" actions, defining a list of parallel
	// tasks to be executed.
	SubTasks []SubTask `json:"sub_tasks,omitempty"`
}

// SubTask defines a single task to be executed in parallel within a "fan_out" step.
type SubTask struct {
	// StepName is the logical name for this sub-task, used for dependency tracking.
	StepName string `json:"step_name"`
	// Topic is the Kafka topic to which the request for this sub-task will be sent.
	Topic string `json:"topic"`
}
