// FILE: platform/orchestration/types/orchestration_state.go
package types

import (
	"time"

	"github.com/gqls/agentchassis/pkg/models"
)

// Move OrchestrationState and related types here
type OrchestrationState struct {
	// All the fields from the original
	OrchestrationID       string
	OrchestrationName     string
	CorrelationID         string
	OwnerAgentID          string
	OwnerAgentType        string
	ParentOrchestrationID string
	ClientID              string
	Status                string
	CurrentStep           string
	AwaitedSteps          []string
	AwaitedRequests       map[string]*AwaitedRequest
	CurrentlyExecuting    *string
	LastActivity          time.Time
	ProcessingNode        string
	ExecutionStartedAt    *time.Time
	CollectedData         map[string]interface{}
	InitialRequestData    []byte
	FinalResult           []byte
	WorkflowPlan          models.WorkflowPlan
	ExecutionPath         []ExecutionRecord
	ExecutionMetadata     ExecutionMetadata
	ProcessingHistory     []ProcessingRecord
	SubtreeAgents         map[string]*SubtreeInfo
	FuelBudget            int
	Error                 string
	Version               int
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// Also move related types
type AwaitedRequest struct {
	RequestID       string
	StepID          string
	StepName        string
	RetryVersion    int
	TargetAgentID   string
	TargetAgentType string
	ResponseTopic   string
	SentAt          time.Time
	TimeoutAt       time.Time
	ParentRequestID string
}

type ExecutionRecord struct {
	Step      string
	Action    string
	StartTime time.Time
	EndTime   *time.Time
	Result    string
	Error     string
}

type ExecutionMetadata struct {
	TotalSteps     int
	CompletedSteps int
	SkippedSteps   int
	FailedSteps    int
	RetryCount     map[string]int
	Checkpoints    map[string]time.Time
	StartTime      time.Time
	EndTime        *time.Time
}

type ProcessingRecord struct {
	PodName   string
	StepID    string
	StepName  string
	Action    string
	Timestamp time.Time
	Details   string
}
