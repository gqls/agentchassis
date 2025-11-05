# Human-in-the-Loop (HITL) Implementation Plan

## Executive Summary

This document outlines the implementation plan for adding Human-in-the-Loop approval capabilities to the agent orchestration system. The system will support text and image content approval through a specialized approval agent architecture with pull-based API access.

---

## Key Design Decisions

### 1. **Architecture Pattern: Approval as Specialized Agent**
- **Decision**: Approval is handled by spawning a specialized `human-reviewer` agent, not an inline action
- **Rationale**: Aligns with "every agent is an orchestrator" philosophy; keeps approval complexity isolated; extensible for future approval types (agent reviewers, rule-based)
- **Trade-off**: More complex than inline action, but architecturally consistent

### 2. **Communication Pattern: Kafka-based with State Pause**
- **Decision**: Approval agent creates task → sends to Kafka → pauses workflow → waits for response message
- **Rationale**: Consistent with existing async agent communication; allows for proper state tracking; supports eventual webhook/notification system
- **Implementation**: `StatusPausedForHuman` state already exists in codebase

### 3. **API Pattern: Pull-based (Client Polling)**
- **Decision**: Client polls for pending approvals; no webhooks or push notifications in Phase 1
- **Rationale**: Client controls authentication, rate limiting, and timing; simpler to implement; client can be offline without issues
- **API Endpoints**: RESTful endpoints for listing, getting details, and submitting decisions

### 4. **Content Types Supported**
- **Phase 1**: Text (LLM outputs) and Images (generated images)
- **Phase 2**: Search results, structured data, video
- **Abstraction**: Content-type agnostic approval agent with type-specific handling

### 5. **Image Handling**
- **Storage**: Backblaze S3 with versioned paths
- **Format**: PNG only for Phase 1
- **Replacement**: Pre-signed URL upload
- **Retention**: Keep all versions forever (audit trail)
- **Path Structure**: `/clients/{client_id}/jobs/{orchestration_id}/{generated|user-uploads|approved}/`

### 6. **Data Storage**
- **Orchestration State**: Stored in Chassis Postgres DB (CollectedData with metadata)
- **Approval Tasks**: New `approval_tasks` and `approval_versions` tables in Chassis DB
- **Content Versions**: Kept in Backblaze for images; inline for text
- **Client DB Sync**: Out of scope for Phase 1; client pulls completed orchestrations when ready

### 7. **Version Control**
- **Workflow Versioning**: Out of scope for Phase 1
- **Prompt Library**: Out of scope for Phase 1
- **Content Versioning**: Full audit trail in `approval_versions` table
- **CollectedData Format**: Option 2 (final result + metadata)

### 8. **Scope Boundaries**

**In Scope (Phase 1)**:
- ✅ Text approval (LLM outputs) with direct editing
- ✅ Image approval with replacement via upload
- ✅ Specialized `human-reviewer` agent
- ✅ Pull-based API for approval management
- ✅ Kafka-based approval request/response flow
- ✅ Full audit trail (who, what, when)
- ✅ Content-type abstraction for extensibility

**Out of Scope (Phase 1)**:
- ❌ Prompt regeneration and approval loops
- ❌ Search result selection/filtering
- ❌ Workflow versioning system
- ❌ Multiple reviewers/collaboration
- ❌ Timeout handling and auto-approval
- ❌ Escalation mechanisms
- ❌ Agent-based approval (sense-checking agents)

---

## Implementation Plan

### Phase 1.1: Database Schema & Core Types

**Objective**: Set up database tables and Go types for approval system

#### 1.1.1 Database Migrations

**File**: `platform/migrations/004_approval_system.sql`

```sql
-- Approval tasks table (renamed from human_tasks for generality)
CREATE TABLE IF NOT EXISTS approval_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Orchestration linking
    correlation_id VARCHAR(255) NOT NULL,
    orchestration_id VARCHAR(255) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    step_name VARCHAR(255) NOT NULL,
    
    -- Task configuration
    task_type VARCHAR(100) NOT NULL,  -- 'human_review', 'agent_review', 'rule_based'
    content_type VARCHAR(50) NOT NULL,  -- 'text', 'image', 'search_results'
    assigned_to VARCHAR(255),  -- user_id or agent_id
    assigned_to_type VARCHAR(50) DEFAULT 'human',  -- 'human', 'agent'
    
    -- Task status
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',  -- 'PENDING', 'APPROVED', 'REJECTED', 'SUPERSEDED'
    
    -- Task data (content-type specific)
    task_data JSONB NOT NULL,
    
    -- Response data (content-type specific)
    response_data JSONB,
    
    -- Versioning support
    version INT DEFAULT 0,
    parent_task_id UUID,  -- For future regeneration support
    
    -- Timing
    timeout_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    assigned_at TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Foreign key
    CONSTRAINT fk_correlation
        FOREIGN KEY (correlation_id)
        REFERENCES orchestrator_state(correlation_id)
);

-- Approval version history table
CREATE TABLE IF NOT EXISTS approval_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_task_id UUID NOT NULL,
    version INT NOT NULL,
    
    -- Version data
    action VARCHAR(50) NOT NULL,  -- 'initial', 'edited', 'approved', 'rejected', 'replaced'
    content_data JSONB NOT NULL,
    notes TEXT,
    
    -- Audit
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT fk_approval_task
        FOREIGN KEY (approval_task_id)
        REFERENCES approval_tasks(id)
        ON DELETE CASCADE,
    
    UNIQUE(approval_task_id, version)
);

-- Indexes for performance
CREATE INDEX idx_approval_tasks_correlation ON approval_tasks(correlation_id);
CREATE INDEX idx_approval_tasks_orchestration ON approval_tasks(orchestration_id);
CREATE INDEX idx_approval_tasks_status ON approval_tasks(status);
CREATE INDEX idx_approval_tasks_assigned ON approval_tasks(assigned_to);
CREATE INDEX idx_approval_tasks_client ON approval_tasks(client_id, status);
CREATE INDEX idx_approval_versions_task ON approval_versions(approval_task_id);
```

#### 1.1.2 Go Types

**File**: `platform/orchestration/types/approval_types.go`

```go
package types

import "time"

// ApprovalTaskType defines the type of approval
type ApprovalTaskType string

const (
	ApprovalTypeHumanReview ApprovalTaskType = "human_review"
	ApprovalTypeAgentReview ApprovalTaskType = "agent_review"
	ApprovalTypeRuleBased   ApprovalTaskType = "rule_based"
)

// ContentType defines what kind of content is being reviewed
type ContentType string

const (
	ContentTypeText          ContentType = "text"
	ContentTypeImage         ContentType = "image"
	ContentTypeSearchResults ContentType = "search_results"
	ContentTypeStructured    ContentType = "structured_data"
)

// ApprovalStatus defines the current status of an approval task
type ApprovalStatus string

const (
	ApprovalStatusPending    ApprovalStatus = "PENDING"
	ApprovalStatusApproved   ApprovalStatus = "APPROVED"
	ApprovalStatusRejected   ApprovalStatus = "REJECTED"
	ApprovalStatusSuperseded ApprovalStatus = "SUPERSEDED"
)

// ApprovalAction defines what action the reviewer took
type ApprovalAction string

const (
	ApprovalActionInitial  ApprovalAction = "initial"
	ApprovalActionApprove  ApprovalAction = "approve"
	ApprovalActionEdit     ApprovalAction = "edit"
	ApprovalActionReplace  ApprovalAction = "replace"
	ApprovalActionReject   ApprovalAction = "reject"
)

// ApprovalTask represents a task requiring approval
type ApprovalTask struct {
	ID              string                 `json:"id" db:"id"`
	CorrelationID   string                 `json:"correlation_id" db:"correlation_id"`
	OrchestrationID string                 `json:"orchestration_id" db:"orchestration_id"`
	ClientID        string                 `json:"client_id" db:"client_id"`
	StepName        string                 `json:"step_name" db:"step_name"`
	
	TaskType       ApprovalTaskType       `json:"task_type" db:"task_type"`
	ContentType    ContentType            `json:"content_type" db:"content_type"`
	AssignedTo     string                 `json:"assigned_to" db:"assigned_to"`
	AssignedToType string                 `json:"assigned_to_type" db:"assigned_to_type"`
	
	Status         ApprovalStatus         `json:"status" db:"status"`
	
	TaskData       map[string]interface{} `json:"task_data" db:"task_data"`
	ResponseData   map[string]interface{} `json:"response_data,omitempty" db:"response_data"`
	
	Version        int                    `json:"version" db:"version"`
	ParentTaskID   *string                `json:"parent_task_id,omitempty" db:"parent_task_id"`
	
	TimeoutAt      *time.Time             `json:"timeout_at,omitempty" db:"timeout_at"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	AssignedAt     *time.Time             `json:"assigned_at,omitempty" db:"assigned_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

// ApprovalVersion tracks the history of an approval task
type ApprovalVersion struct {
	ID              string                 `json:"id" db:"id"`
	ApprovalTaskID  string                 `json:"approval_task_id" db:"approval_task_id"`
	Version         int                    `json:"version" db:"version"`
	
	Action          ApprovalAction         `json:"action" db:"action"`
	ContentData     map[string]interface{} `json:"content_data" db:"content_data"`
	Notes           string                 `json:"notes,omitempty" db:"notes"`
	
	CreatedBy       string                 `json:"created_by" db:"created_by"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// ApprovalCapabilities defines what actions are available for this approval
type ApprovalCapabilities struct {
	CanApprove        bool     `json:"can_approve"`
	CanReject         bool     `json:"can_reject"`
	CanEditDirectly   bool     `json:"can_edit_directly"`
	CanReplace        bool     `json:"can_replace"`
	ReplacementMethods []string `json:"replacement_methods,omitempty"`
}

// TextApprovalData is the task_data structure for text content
type TextApprovalData struct {
	ContentType    string                 `json:"content_type"`
	PromptUsed     string                 `json:"prompt_used"`
	LLMOutput      string                 `json:"llm_output"`
	Model          string                 `json:"model"`
	InputData      map[string]interface{} `json:"input_data,omitempty"`
}

// ImageApprovalData is the task_data structure for image content
type ImageApprovalData struct {
	ContentType    string `json:"content_type"`
	PromptUsed     string `json:"prompt_used"`
	ImageURL       string `json:"image_url"`
	Model          string `json:"model"`
	Dimensions     string `json:"dimensions,omitempty"`
	Format         string `json:"format"`
}

// TextApprovalResponse is the response_data structure for text approvals
type TextApprovalResponse struct {
	Action          ApprovalAction `json:"action"`
	OriginalContent string         `json:"original_content"`
	EditedContent   string         `json:"edited_content,omitempty"`
	EditNotes       string         `json:"edit_notes,omitempty"`
	ApprovedBy      string         `json:"approved_by"`
}

// ImageApprovalResponse is the response_data structure for image approvals
type ImageApprovalResponse struct {
	Action            ApprovalAction `json:"action"`
	OriginalURL       string         `json:"original_url"`
	ReplacementURL    string         `json:"replacement_url,omitempty"`
	ReplacementSource string         `json:"replacement_source,omitempty"`
	ApprovedBy        string         `json:"approved_by"`
}
```

#### 1.1.3 Repository Layer

**File**: `platform/orchestration/approval_repository.go`

```go
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

type ApprovalRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewApprovalRepository(db *sql.DB, logger *zap.Logger) *ApprovalRepository {
	return &ApprovalRepository{
		db:     db,
		logger: logger,
	}
}

// CreateApprovalTask creates a new approval task
func (r *ApprovalRepository) CreateApprovalTask(ctx context.Context, task *types.ApprovalTask) error {
	taskDataJSON, err := json.Marshal(task.TaskData)
	if err != nil {
		return fmt.Errorf("failed to marshal task_data: %w", err)
	}
	
	query := `
		INSERT INTO approval_tasks (
			id, correlation_id, orchestration_id, client_id, step_name,
			task_type, content_type, assigned_to, assigned_to_type,
			status, task_data, version, timeout_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at
	`
	
	err = r.db.QueryRowContext(
		ctx, query,
		task.ID, task.CorrelationID, task.OrchestrationID, task.ClientID, task.StepName,
		task.TaskType, task.ContentType, task.AssignedTo, task.AssignedToType,
		task.Status, taskDataJSON, task.Version, task.TimeoutAt,
	).Scan(&task.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create approval task: %w", err)
	}
	
	r.logger.Info("Created approval task",
		zap.String("task_id", task.ID),
		zap.String("orchestration_id", task.OrchestrationID),
		zap.String("content_type", string(task.ContentType)))
	
	return nil
}

// GetApprovalTask retrieves an approval task by ID
func (r *ApprovalRepository) GetApprovalTask(ctx context.Context, taskID string) (*types.ApprovalTask, error) {
	query := `
		SELECT 
			id, correlation_id, orchestration_id, client_id, step_name,
			task_type, content_type, assigned_to, assigned_to_type,
			status, task_data, response_data, version, parent_task_id,
			timeout_at, created_at, assigned_at, completed_at
		FROM approval_tasks
		WHERE id = $1
	`
	
	task := &types.ApprovalTask{}
	var taskDataJSON, responseDataJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(
		&task.ID, &task.CorrelationID, &task.OrchestrationID, &task.ClientID, &task.StepName,
		&task.TaskType, &task.ContentType, &task.AssignedTo, &task.AssignedToType,
		&task.Status, &taskDataJSON, &responseDataJSON, &task.Version, &task.ParentTaskID,
		&task.TimeoutAt, &task.CreatedAt, &task.AssignedAt, &task.CompletedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("approval task not found: %s", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get approval task: %w", err)
	}
	
	if err := json.Unmarshal(taskDataJSON, &task.TaskData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task_data: %w", err)
	}
	
	if len(responseDataJSON) > 0 {
		if err := json.Unmarshal(responseDataJSON, &task.ResponseData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response_data: %w", err)
		}
	}
	
	return task, nil
}

// ListPendingApprovalTasks lists all pending approval tasks for a client
func (r *ApprovalRepository) ListPendingApprovalTasks(ctx context.Context, clientID string) ([]*types.ApprovalTask, error) {
	query := `
		SELECT 
			id, correlation_id, orchestration_id, client_id, step_name,
			task_type, content_type, assigned_to, assigned_to_type,
			status, task_data, version,
			timeout_at, created_at, assigned_at
		FROM approval_tasks
		WHERE client_id = $1 AND status = 'PENDING'
		ORDER BY created_at ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending approvals: %w", err)
	}
	defer rows.Close()
	
	tasks := []*types.ApprovalTask{}
	
	for rows.Next() {
		task := &types.ApprovalTask{}
		var taskDataJSON []byte
		
		err := rows.Scan(
			&task.ID, &task.CorrelationID, &task.OrchestrationID, &task.ClientID, &task.StepName,
			&task.TaskType, &task.ContentType, &task.AssignedTo, &task.AssignedToType,
			&task.Status, &taskDataJSON, &task.Version,
			&task.TimeoutAt, &task.CreatedAt, &task.AssignedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan approval task: %w", err)
		}
		
		if err := json.Unmarshal(taskDataJSON, &task.TaskData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal task_data: %w", err)
		}
		
		tasks = append(tasks, task)
	}
	
	return tasks, nil
}

// UpdateApprovalTaskWithResponse updates task with human response
func (r *ApprovalRepository) UpdateApprovalTaskWithResponse(
	ctx context.Context,
	taskID string,
	status types.ApprovalStatus,
	responseData map[string]interface{},
) error {
	responseDataJSON, err := json.Marshal(responseData)
	if err != nil {
		return fmt.Errorf("failed to marshal response_data: %w", err)
	}
	
	query := `
		UPDATE approval_tasks
		SET status = $1,
		    response_data = $2,
		    completed_at = NOW()
		WHERE id = $3
	`
	
	result, err := r.db.ExecContext(ctx, query, status, responseDataJSON, taskID)
	if err != nil {
		return fmt.Errorf("failed to update approval task: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rows == 0 {
		return fmt.Errorf("approval task not found: %s", taskID)
	}
	
	r.logger.Info("Updated approval task with response",
		zap.String("task_id", taskID),
		zap.String("status", string(status)))
	
	return nil
}

// CreateApprovalVersion creates a version history entry
func (r *ApprovalRepository) CreateApprovalVersion(ctx context.Context, version *types.ApprovalVersion) error {
	contentDataJSON, err := json.Marshal(version.ContentData)
	if err != nil {
		return fmt.Errorf("failed to marshal content_data: %w", err)
	}
	
	query := `
		INSERT INTO approval_versions (
			id, approval_task_id, version, action, content_data, notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`
	
	err = r.db.QueryRowContext(
		ctx, query,
		version.ID, version.ApprovalTaskID, version.Version,
		version.Action, contentDataJSON, version.Notes, version.CreatedBy,
	).Scan(&version.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create approval version: %w", err)
	}
	
	return nil
}

// GetApprovalVersions retrieves all versions for an approval task
func (r *ApprovalRepository) GetApprovalVersions(ctx context.Context, taskID string) ([]*types.ApprovalVersion, error) {
	query := `
		SELECT id, approval_task_id, version, action, content_data, notes, created_by, created_at
		FROM approval_versions
		WHERE approval_task_id = $1
		ORDER BY version ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list approval versions: %w", err)
	}
	defer rows.Close()
	
	versions := []*types.ApprovalVersion{}
	
	for rows.Next() {
		version := &types.ApprovalVersion{}
		var contentDataJSON []byte
		
		err := rows.Scan(
			&version.ID, &version.ApprovalTaskID, &version.Version,
			&version.Action, &contentDataJSON, &version.Notes,
			&version.CreatedBy, &version.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan approval version: %w", err)
		}
		
		if err := json.Unmarshal(contentDataJSON, &version.ContentData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal content_data: %w", err)
		}
		
		versions = append(versions, version)
	}
	
	return versions, nil
}
```

---

### Phase 1.2: Human-Reviewer Agent

**Objective**: Create the specialized `human-reviewer` agent type

#### 1.2.1 Agent Configuration

**File**: `configs/agents/human-reviewer.yaml`

```yaml
agent_type: human-reviewer
agent_name: Human Reviewer Agent
agent_version: "1.0.0"
description: "Specialized agent for handling human approval workflows"

# This agent always requires stateful operation
enable_stateless: false

# Default workflow for human-reviewer agent
default_workflow:
  start_step: create_approval_task
  
  steps:
    create_approval_task:
      action: create_approval_request
      description: "Create approval task and pause"
      next_step: wait_for_approval
    
    wait_for_approval:
      action: wait_for_approval_response
      description: "Wait for human decision via Kafka"
      next_step: process_approval
    
    process_approval:
      action: process_approval_decision
      description: "Process approval decision and update data"
      next_step: complete
    
    complete:
      action: complete_workflow
      description: "Return approved result to parent"

# Kafka topics
kafka:
  requests_topic_pattern: "job.{correlation_id}-{orchestration_id}-{agent_type}-{step_name}.requests"
  responses_topic_pattern: "job.{correlation_id}-{orchestration_id}-{agent_type}-{step_name}.responses"
  
  # Human approval command topics
  approval_commands_topic: "system.human.approval-commands"
  approval_responses_topic: "system.human.approval-responses"
```

#### 1.2.2 Agent Registration

**File**: `platform/agentbase/agent_registry.go` (modify existing)

```go
// Add to RegisterDefaultAgents() or similar initialization

func RegisterHumanReviewerAgent() {
	registry.RegisterAgentType("human-reviewer", AgentTypeConfig{
		SupportsStateless: false,  // Always stateful
		DefaultConfig: map[string]interface{}{
			"max_wait_time": 3600,  // 1 hour default
			"supported_content_types": []string{"text", "image"},
		},
		RequiredActions: []string{
			"create_approval_request",
			"wait_for_approval_response",
			"process_approval_decision",
		},
	})
}
```

#### 1.2.3 Approval Actions

**File**: `platform/orchestration/actions/approval_actions.go`

```go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// CreateApprovalRequestAction creates an approval task and publishes request
type CreateApprovalRequestAction struct {
	approvalRepo *orchestration.ApprovalRepository
	producer     *kafka.KafkaProducer
	logger       *zap.Logger
}

func NewCreateApprovalRequestAction(
	approvalRepo *orchestration.ApprovalRepository,
	producer *kafka.KafkaProducer,
	logger *zap.Logger,
) *CreateApprovalRequestAction {
	return &CreateApprovalRequestAction{
		approvalRepo: approvalRepo,
		producer:     producer,
		logger:       logger,
	}
}

func (a *CreateApprovalRequestAction) Execute(
	ctx context.Context,
	collectedData map[string]interface{},
	config map[string]interface{},
	execCtx *types.ExecutionContext,
) (map[string]interface{}, error) {
	
	a.logger.Info("Creating approval request",
		zap.String("orchestration_id", execCtx.OrchestrationID))
	
	// Extract configuration
	targetStep, ok := config["target_step"].(string)
	if !ok {
		return nil, fmt.Errorf("target_step not specified in config")
	}
	
	// Get the target step's result from CollectedData
	targetData, ok := collectedData[targetStep].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("target step %s data not found in CollectedData", targetStep)
	}
	
	// Infer content type from target step result
	contentType := inferContentType(targetData)
	
	// Build task data based on content type
	taskData, err := buildTaskData(contentType, targetData, collectedData)
	if err != nil {
		return nil, fmt.Errorf("failed to build task data: %w", err)
	}
	
	// Create approval task
	task := &types.ApprovalTask{
		ID:              uuid.New().String(),
		CorrelationID:   execCtx.CorrelationID,
		OrchestrationID: execCtx.ParentOrchestrationID, // Parent's orchestration
		ClientID:        execCtx.ClientID,
		StepName:        targetStep,
		TaskType:        types.ApprovalTypeHumanReview,
		ContentType:     contentType,
		AssignedTo:      execCtx.ClientID, // Default: assign to client
		AssignedToType:  "human",
		Status:          types.ApprovalStatusPending,
		TaskData:        taskData,
		Version:         0,
	}
	
	// Save to database
	if err := a.approvalRepo.CreateApprovalTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create approval task: %w", err)
	}
	
	// Create initial version entry
	initialVersion := &types.ApprovalVersion{
		ID:             uuid.New().String(),
		ApprovalTaskID: task.ID,
		Version:        0,
		Action:         types.ApprovalActionInitial,
		ContentData:    taskData,
		Notes:          "Initial creation",
		CreatedBy:      "system",
	}
	
	if err := a.approvalRepo.CreateApprovalVersion(ctx, initialVersion); err != nil {
		a.logger.Warn("Failed to create initial version", zap.Error(err))
		// Non-fatal, continue
	}
	
	// Publish approval request to Kafka
	approvalRequest := map[string]interface{}{
		"event_type":       "approval_request_created",
		"approval_task_id": task.ID,
		"orchestration_id": task.OrchestrationID,
		"client_id":        task.ClientID,
		"content_type":     string(task.ContentType),
		"step_name":        task.StepName,
		"timestamp":        time.Now().UTC(),
	}
	
	if err := a.publishApprovalRequest(ctx, approvalRequest); err != nil {
		a.logger.Error("Failed to publish approval request", zap.Error(err))
		// Non-fatal, task is in DB so client can still poll
	}
	
	a.logger.Info("Approval request created",
		zap.String("task_id", task.ID),
		zap.String("content_type", string(contentType)))
	
	return map[string]interface{}{
		"approval_task_id": task.ID,
		"status":           "pending",
		"content_type":     string(contentType),
	}, nil
}

// WaitForApprovalResponseAction waits for human response via Kafka
type WaitForApprovalResponseAction struct {
	approvalRepo *orchestration.ApprovalRepository
	stateRepo    *orchestration.StateRepository
	logger       *zap.Logger
}

func NewWaitForApprovalResponseAction(
	approvalRepo *orchestration.ApprovalRepository,
	stateRepo *orchestration.StateRepository,
	logger *zap.Logger,
) *WaitForApprovalResponseAction {
	return &WaitForApprovalResponseAction{
		approvalRepo: approvalRepo,
		stateRepo:    stateRepo,
		logger:       logger,
	}
}

func (a *WaitForApprovalResponseAction) Execute(
	ctx context.Context,
	collectedData map[string]interface{},
	config map[string]interface{},
	execCtx *types.ExecutionContext,
) (map[string]interface{}, error) {
	
	// Get approval task ID from previous step
	createResult, ok := collectedData["create_approval_task"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("create_approval_task result not found")
	}
	
	taskID, ok := createResult["approval_task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("approval_task_id not found in create result")
	}
	
	a.logger.Info("Waiting for approval response",
		zap.String("task_id", taskID),
		zap.String("orchestration_id", execCtx.OrchestrationID))
	
	// Update orchestration state to PAUSED_FOR_HUMAN
	if err := a.stateRepo.UpdateStatus(
		ctx,
		execCtx.OrchestrationID,
		orchestration.StatusPausedForHuman,
	); err != nil {
		a.logger.Error("Failed to update status to PAUSED_FOR_HUMAN", zap.Error(err))
	}
	
	// This action will pause here and resume when response message arrives
	// The coordinator will handle the Kafka message and resume execution
	// For now, return pending status
	
	return map[string]interface{}{
		"approval_task_id": taskID,
		"status":           "waiting",
		"paused_at":        time.Now().UTC(),
	}, nil
}

// ProcessApprovalDecisionAction processes the approval response
type ProcessApprovalDecisionAction struct {
	approvalRepo *orchestration.ApprovalRepository
	logger       *zap.Logger
}

func NewProcessApprovalDecisionAction(
	approvalRepo *orchestration.ApprovalRepository,
	logger *zap.Logger,
) *ProcessApprovalDecisionAction {
	return &ProcessApprovalDecisionAction{
		approvalRepo: approvalRepo,
		logger:       logger,
	}
}

func (a *ProcessApprovalDecisionAction) Execute(
	ctx context.Context,
	collectedData map[string]interface{},
	config map[string]interface{},
	execCtx *types.ExecutionContext,
) (map[string]interface{}, error) {
	
	// Get approval task ID
	waitResult, ok := collectedData["wait_for_approval"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("wait_for_approval result not found")
	}
	
	taskID, ok := waitResult["approval_task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("approval_task_id not found")
	}
	
	// Fetch the completed approval task
	task, err := a.approvalRepo.GetApprovalTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval task: %w", err)
	}
	
	if task.Status != types.ApprovalStatusApproved {
		return nil, fmt.Errorf("approval task not approved: %s", task.Status)
	}
	
	a.logger.Info("Processing approval decision",
		zap.String("task_id", taskID),
		zap.String("status", string(task.Status)))
	
	// Extract the approved/edited content
	approvedContent, err := extractApprovedContent(task)
	if err != nil {
		return nil, fmt.Errorf("failed to extract approved content: %w", err)
	}
	
	// Return the approved content to parent orchestration
	return map[string]interface{}{
		"approval_task_id": task.ID,
		"status":           "approved",
		"content_type":     string(task.ContentType),
		"approved_content": approvedContent,
		"approved_by":      task.ResponseData["approved_by"],
		"approved_at":      task.CompletedAt,
	}, nil
}

// Helper functions

func inferContentType(targetData map[string]interface{}) types.ContentType {
	// Check for image indicators
	if _, ok := targetData["image_url"]; ok {
		return types.ContentTypeImage
	}
	
	// Check for text indicators
	if _, ok := targetData["result"]; ok {
		return types.ContentTypeText
	}
	
	// Default to text
	return types.ContentTypeText
}

func buildTaskData(
	contentType types.ContentType,
	targetData map[string]interface{},
	collectedData map[string]interface{},
) (map[string]interface{}, error) {
	
	switch contentType {
	case types.ContentTypeText:
		return types.TextApprovalData{
			ContentType: "text",
			PromptUsed:  getStringField(targetData, "prompt"),
			LLMOutput:   getStringField(targetData, "result"),
			Model:       getStringField(targetData, "model"),
			InputData:   getMapField(targetData, "input_data"),
		}, nil
		
	case types.ContentTypeImage:
		return types.ImageApprovalData{
			ContentType: "image",
			PromptUsed:  getStringField(targetData, "prompt"),
			ImageURL:    getStringField(targetData, "image_url"),
			Model:       getStringField(targetData, "model"),
			Dimensions:  getStringField(targetData, "dimensions"),
			Format:      getStringField(targetData, "format"),
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
}

func extractApprovedContent(task *types.ApprovalTask) (map[string]interface{}, error) {
	switch task.ContentType {
	case types.ContentTypeText:
		// Check if content was edited
		if editedContent, ok := task.ResponseData["edited_content"].(string); ok {
			return map[string]interface{}{
				"result":  editedContent,
				"edited":  true,
				"original_result": task.TaskData["llm_output"],
			}, nil
		}
		
		// Not edited, return original
		return map[string]interface{}{
			"result": task.TaskData["llm_output"],
			"edited": false,
		}, nil
		
	case types.ContentTypeImage:
		// Check if image was replaced
		if replacementURL, ok := task.ResponseData["replacement_url"].(string); ok {
			return map[string]interface{}{
				"image_url": replacementURL,
				"replaced":  true,
				"original_url": task.TaskData["image_url"],
			}, nil
		}
		
		// Not replaced, return original
		return map[string]interface{}{
			"image_url": task.TaskData["image_url"],
			"replaced":  false,
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported content type: %s", task.ContentType)
	}
}

func (a *CreateApprovalRequestAction) publishApprovalRequest(
	ctx context.Context,
	request map[string]interface{},
) error {
	
	message, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	topic := "system.human.approval-requests"
	
	return a.producer.Produce(ctx, topic, "", message)
}

func getStringField(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getMapField(data map[string]interface{}, key string) map[string]interface{} {
	if val, ok := data[key].(map[string]interface{}); ok {
		return val
	}
	return map[string]interface{}{}
}
```

---

### Phase 1.3: Workflow Integration

**Objective**: Enable workflows to spawn human-reviewer agents

#### 1.3.1 Example Parent Workflow

**File**: `configs/workflows/content-creator-hero-with-approval.yaml`

```yaml
workflow_name: content-creator-hero-with-approval
version: "1.0.0"
description: "Hero content creation with human approval"

workflow:
  start_step: spawn_researcher
  
  steps:
    spawn_researcher:
      action: spawn_agent
      description: "Spawn research agent"
      config:
        agent_type: content-researcher
        role: researcher
      next_step: call_researcher
    
    call_researcher:
      action: call_agent
      description: "Get research data"
      config:
        agent_type: content-researcher
        target_role: researcher
        input_data:
          business_type: "{{.input_data.business_type}}"
        prompt: "Research background information about {{.business_type}} businesses"
      next_step: generate_hero_content
    
    generate_hero_content:
      action: execute_llm_prompt
      description: "Generate hero section"
      config:
        input_fields: ["call_researcher"]
        prompt: "Write a compelling hero section for {{.business_name}}, a {{.business_type}}. Include a powerful headline and subheadline."
      next_step: review_hero_content
    
    # NEW: Approval step
    review_hero_content:
      action: spawn_agent
      description: "Human review of hero content"
      config:
        agent_type: human-reviewer
        role: content_approver
        # Pass configuration to reviewer agent
        approval_config:
          target_step: generate_hero_content
          capabilities:
            can_approve: true
            can_reject: true
            can_edit_directly: true
      next_step: update_with_approved_content
    
    # NEW: Update CollectedData with approved content
    update_with_approved_content:
      action: merge_data
      description: "Merge approved content back"
      config:
        source_step: review_hero_content
        target_step: generate_hero_content
        merge_strategy: replace
      next_step: complete
    
    complete:
      action: complete_workflow
      description: "Return final content"
```

#### 1.3.2 New Action: merge_data

**File**: `platform/orchestration/actions/data_actions.go`

```go
package actions

import (
	"context"
	"fmt"
	
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// MergeDataAction merges data from one step into another
type MergeDataAction struct {
	logger *zap.Logger
}

func NewMergeDataAction(logger *zap.Logger) *MergeDataAction {
	return &MergeDataAction{logger: logger}
}

func (a *MergeDataAction) Execute(
	ctx context.Context,
	collectedData map[string]interface{},
	config map[string]interface{},
	execCtx *types.ExecutionContext,
) (map[string]interface{}, error) {
	
	sourceStep, ok := config["source_step"].(string)
	if !ok {
		return nil, fmt.Errorf("source_step not specified")
	}
	
	targetStep, ok := config["target_step"].(string)
	if !ok {
		return nil, fmt.Errorf("target_step not specified")
	}
	
	mergeStrategy, _ := config["merge_strategy"].(string)
	if mergeStrategy == "" {
		mergeStrategy = "replace"
	}
	
	// Get source data (from approval agent)
	sourceData, ok := collectedData[sourceStep].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("source step %s data not found", sourceStep)
	}
	
	approvedContent, ok := sourceData["approved_content"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("approved_content not found in source data")
	}
	
	// Get target data
	targetData, ok := collectedData[targetStep].(map[string]interface{})
	if !ok {
		targetData = map[string]interface{}{}
	}
	
	// Merge based on strategy
	switch mergeStrategy {
	case "replace":
		// Replace target with approved content
		for k, v := range approvedContent {
			targetData[k] = v
		}
		
	case "merge":
		// Merge approved content into target, keeping non-conflicting fields
		for k, v := range approvedContent {
			targetData[k] = v
		}
		
	default:
		return nil, fmt.Errorf("unknown merge strategy: %s", mergeStrategy)
	}
	
	// Add approval metadata
	targetData["approved"] = true
	targetData["approved_by"] = sourceData["approved_by"]
	targetData["approved_at"] = sourceData["approved_at"]
	
	// Update CollectedData
	collectedData[targetStep] = targetData
	
	a.logger.Info("Merged approved content",
		zap.String("source_step", sourceStep),
		zap.String("target_step", targetStep),
		zap.String("strategy", mergeStrategy))
	
	return map[string]interface{}{
		"merged": true,
		"source_step": sourceStep,
		"target_step": targetStep,
	}, nil
}
```

---

### Phase 1.4: API Layer

**Objective**: RESTful API for clients to manage approvals

#### 1.4.1 API Router

**File**: `platform/api/approval_handler.go`

```go
package api

import (
	"encoding/json"
	"net/http"
	"time"
	
	"github.com/gorilla/mux"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

type ApprovalHandler struct {
	approvalRepo *orchestration.ApprovalRepository
	stateRepo    *orchestration.StateRepository
	logger       *zap.Logger
}

func NewApprovalHandler(
	approvalRepo *orchestration.ApprovalRepository,
	stateRepo *orchestration.StateRepository,
	logger *zap.Logger,
) *ApprovalHandler {
	return &ApprovalHandler{
		approvalRepo: approvalRepo,
		stateRepo:    stateRepo,
		logger:       logger,
	}
}

func (h *ApprovalHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/approvals", h.ListPendingApprovals).Methods("GET")
	router.HandleFunc("/api/v1/approvals/{approval_id}", h.GetApproval).Methods("GET")
	router.HandleFunc("/api/v1/approvals/{approval_id}/approve", h.ApproveTask).Methods("POST")
	router.HandleFunc("/api/v1/approvals/{approval_id}/reject", h.RejectTask).Methods("POST")
	router.HandleFunc("/api/v1/approvals/{approval_id}/upload-url", h.GetUploadURL).Methods("POST")
}

// ListPendingApprovals lists all pending approvals for a client
func (h *ApprovalHandler) ListPendingApprovals(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		http.Error(w, "client_id required", http.StatusBadRequest)
		return
	}
	
	tasks, err := h.approvalRepo.ListPendingApprovalTasks(r.Context(), clientID)
	if err != nil {
		h.logger.Error("Failed to list pending approvals", zap.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	
	// Convert to API response format
	response := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		response[i] = h.taskToAPIResponse(task)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"approvals": response,
		"count":     len(tasks),
	})
}

// GetApproval retrieves a specific approval task
func (h *ApprovalHandler) GetApproval(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	approvalID := vars["approval_id"]
	
	task, err := h.approvalRepo.GetApprovalTask(r.Context(), approvalID)
	if err != nil {
		h.logger.Error("Failed to get approval task", zap.Error(err))
		http.Error(w, "Approval not found", http.StatusNotFound)
		return
	}
	
	response := h.taskToDetailedAPIResponse(task)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ApproveTask handles approval submission
func (h *ApprovalHandler) ApproveTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	approvalID := vars["approval_id"]
	
	var request struct {
		Action         string                 `json:"action"`  // "approve", "edit", "replace"
		EditedContent  string                 `json:"edited_content,omitempty"`
		ReplacementURL string                 `json:"replacement_url,omitempty"`
		UploadedPath   string                 `json:"uploaded_path,omitempty"`
		Notes          string                 `json:"notes,omitempty"`
		ApprovedBy     string                 `json:"approved_by"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Get the task
	task, err := h.approvalRepo.GetApprovalTask(r.Context(), approvalID)
	if err != nil {
		http.Error(w, "Approval not found", http.StatusNotFound)
		return
	}
	
	// Build response data based on content type
	responseData, err := h.buildResponseData(task, request)
	if err != nil {
		h.logger.Error("Failed to build response data", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Update task
	if err := h.approvalRepo.UpdateApprovalTaskWithResponse(
		r.Context(),
		approvalID,
		types.ApprovalStatusApproved,
		responseData,
	); err != nil {
		h.logger.Error("Failed to update approval task", zap.Error(err))
		http.Error(w, "Failed to update approval", http.StatusInternalServerError)
		return
	}
	
	// Create version entry
	version := &types.ApprovalVersion{
		ID:             uuid.New().String(),
		ApprovalTaskID: approvalID,
		Version:        task.Version + 1,
		Action:         types.ApprovalAction(request.Action),
		ContentData:    responseData,
		Notes:          request.Notes,
		CreatedBy:      request.ApprovedBy,
	}
	
	if err := h.approvalRepo.CreateApprovalVersion(r.Context(), version); err != nil {
		h.logger.Warn("Failed to create approval version", zap.Error(err))
	}
	
	// Publish approval response to Kafka to resume workflow
	if err := h.publishApprovalResponse(r.Context(), task, responseData); err != nil {
		h.logger.Error("Failed to publish approval response", zap.Error(err))
		// Non-fatal, state is saved in DB
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "approved",
		"task_id": approvalID,
	})
}

// RejectTask handles rejection
func (h *ApprovalHandler) RejectTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	approvalID := vars["approval_id"]
	
	var request struct {
		Reason     string `json:"reason"`
		RejectedBy string `json:"rejected_by"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	responseData := map[string]interface{}{
		"action":      "reject",
		"reason":      request.Reason,
		"rejected_by": request.RejectedBy,
	}
	
	if err := h.approvalRepo.UpdateApprovalTaskWithResponse(
		r.Context(),
		approvalID,
		types.ApprovalStatusRejected,
		responseData,
	); err != nil {
		h.logger.Error("Failed to reject approval task", zap.Error(err))
		http.Error(w, "Failed to reject approval", http.StatusInternalServerError)
		return
	}
	
	// Get task for publishing
	task, _ := h.approvalRepo.GetApprovalTask(r.Context(), approvalID)
	
	// Publish rejection to resume workflow (with error)
	if err := h.publishApprovalResponse(r.Context(), task, responseData); err != nil {
		h.logger.Error("Failed to publish rejection response", zap.Error(err))
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "rejected",
		"task_id": approvalID,
	})
}

// GetUploadURL generates pre-signed URL for image upload
func (h *ApprovalHandler) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	approvalID := vars["approval_id"]
	
	// Get task to determine paths
	task, err := h.approvalRepo.GetApprovalTask(r.Context(), approvalID)
	if err != nil {
		http.Error(w, "Approval not found", http.StatusNotFound)
		return
	}
	
	if task.ContentType != types.ContentTypeImage {
		http.Error(w, "Upload only supported for images", http.StatusBadRequest)
		return
	}
	
	// Generate upload path
	uploadPath := fmt.Sprintf(
		"clients/%s/jobs/%s/user-uploads/approved-image.png",
		task.ClientID,
		task.OrchestrationID,
	)
	
	// Generate pre-signed URL (implement with Backblaze/S3 SDK)
	presignedURL, err := generatePresignedUploadURL(uploadPath, 15*time.Minute)
	if err != nil {
		h.logger.Error("Failed to generate presigned URL", zap.Error(err))
		http.Error(w, "Failed to generate upload URL", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"upload_url":  presignedURL,
		"upload_path": uploadPath,
		"expires_in":  900, // 15 minutes
	})
}

// Helper functions

func (h *ApprovalHandler) taskToAPIResponse(task *types.ApprovalTask) map[string]interface{} {
	return map[string]interface{}{
		"approval_id":      task.ID,
		"orchestration_id": task.OrchestrationID,
		"step_name":        task.StepName,
		"content_type":     string(task.ContentType),
		"status":           string(task.Status),
		"created_at":       task.CreatedAt,
	}
}

func (h *ApprovalHandler) taskToDetailedAPIResponse(task *types.ApprovalTask) map[string]interface{} {
	capabilities := h.getCapabilities(task.ContentType)
	
	return map[string]interface{}{
		"approval_id":      task.ID,
		"orchestration_id": task.OrchestrationID,
		"step_name":        task.StepName,
		"content_type":     string(task.ContentType),
		"status":           string(task.Status),
		"review_data":      task.TaskData,
		"capabilities":     capabilities,
		"metadata": map[string]interface{}{
			"created_at":  task.CreatedAt,
			"assigned_to": task.AssignedTo,
		},
	}
}

func (h *ApprovalHandler) getCapabilities(contentType types.ContentType) map[string]interface{} {
	switch contentType {
	case types.ContentTypeText:
		return map[string]interface{}{
			"can_approve":      true,
			"can_reject":       true,
			"can_edit_directly": true,
			"can_replace":      false,
		}
	case types.ContentTypeImage:
		return map[string]interface{}{
			"can_approve":         true,
			"can_reject":          true,
			"can_edit_directly":   false,
			"can_replace":         true,
			"replacement_methods": []string{"upload_url", "upload_path"},
		}
	default:
		return map[string]interface{}{
			"can_approve": true,
			"can_reject":  true,
		}
	}
}

func (h *ApprovalHandler) buildResponseData(
	task *types.ApprovalTask,
	request struct {
		Action         string
		EditedContent  string
		ReplacementURL string
		UploadedPath   string
		Notes          string
		ApprovedBy     string
	},
) (map[string]interface{}, error) {
	
	switch task.ContentType {
	case types.ContentTypeText:
		response := map[string]interface{}{
			"action":      request.Action,
			"approved_by": request.ApprovedBy,
		}
		
		if request.Action == "edit" && request.EditedContent != "" {
			response["original_content"] = task.TaskData["llm_output"]
			response["edited_content"] = request.EditedContent
			response["edit_notes"] = request.Notes
		}
		
		return response, nil
		
	case types.ContentTypeImage:
		response := map[string]interface{}{
			"action":       request.Action,
			"original_url": task.TaskData["image_url"],
			"approved_by":  request.ApprovedBy,
		}
		
		if request.Action == "replace" {
			if request.ReplacementURL != "" {
				response["replacement_url"] = request.ReplacementURL
				response["replacement_source"] = "external_url"
			} else if request.UploadedPath != "" {
				// Build full URL from uploaded path
				response["replacement_url"] = buildImageURL(request.UploadedPath)
				response["replacement_source"] = "user_upload"
			} else {
				return nil, fmt.Errorf("replacement_url or uploaded_path required for replace action")
			}
		}
		
		return response, nil
		
	default:
		return nil, fmt.Errorf("unsupported content type: %s", task.ContentType)
	}
}

func (h *ApprovalHandler) publishApprovalResponse(
	ctx context.Context,
	task *types.ApprovalTask,
	responseData map[string]interface{},
) error {
	// Build Kafka message to resume human-reviewer agent
	message := map[string]interface{}{
		"event_type":       "approval_response",
		"approval_task_id": task.ID,
		"orchestration_id": task.OrchestrationID,
		"status":           string(task.Status),
		"response_data":    responseData,
		"timestamp":        time.Now().UTC(),
	}
	
	// Publish to human-reviewer agent's responses topic
	// The topic should be stored in the approval task or derived from orchestration
	responseTopic := fmt.Sprintf(
		"job.%s-%s-human-reviewer-wait_for_approval.responses",
		task.CorrelationID,
		task.OrchestrationID,
	)
	
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// Publish via Kafka producer (inject into handler)
	return h.producer.Produce(ctx, responseTopic, task.ID, messageBytes)
}

// Placeholder functions (implement with actual S3/Backblaze SDK)
func generatePresignedUploadURL(path string, duration time.Duration) (string, error) {
	// TODO: Implement with Backblaze B2 or S3 SDK
	return fmt.Sprintf("https://backblaze.example.com/presigned/%s", path), nil
}

func buildImageURL(path string) string {
	// TODO: Build actual URL from path
	return fmt.Sprintf("https://backblaze.example.com/%s", path)
}
```

---

### Phase 1.5: Coordinator Integration

**Objective**: Modify coordinator to handle approval pause/resume

#### 1.5.1 Coordinator Modifications

**File**: `platform/orchestration/coordinator.go` (modifications)

```go
// Add to Coordinator struct
type Coordinator struct {
	// ... existing fields ...
	approvalRepo *ApprovalRepository
}

// Add approval message handler
func (c *Coordinator) handleApprovalResponse(
	ctx context.Context,
	message *types.ResponseMessage,
) error {
	
	c.logger.Info("Received approval response",
		zap.String("orchestration_id", message.Headers.OrchestrationID))
	
	// Extract approval data from message
	approvalData, ok := message.Body.Body.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid approval response body")
	}
	
	// Get orchestration state
	state, err := c.stateRepo.GetOrchestrationState(
		ctx,
		message.Headers.OrchestrationID,
	)
	if err != nil {
		return fmt.Errorf("failed to get orchestration state: %w", err)
	}
	
	// Verify it's paused for human
	if state.Status != StatusPausedForHuman {
		c.logger.Warn("Received approval response but not paused",
			zap.String("status", string(state.Status)))
		return nil
	}
	
	// Resume execution
	state.Status = StatusRunning
	
	if err := c.stateRepo.UpdateOrchestrationState(ctx, state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}
	
	// Continue workflow execution
	return c.continueExecution(ctx, state, message)
}

// Modify message routing to detect approval responses
func (c *Coordinator) routeMessage(ctx context.Context, message interface{}) error {
	// ... existing routing logic ...
	
	// Check if it's an approval response
	if respMsg, ok := message.(*types.ResponseMessage); ok {
		if respMsg.Headers.MessageType == "approval_response" {
			return c.handleApprovalResponse(ctx, respMsg)
		}
	}
	
	// ... rest of routing ...
}
```

---

### Phase 1.6: Testing & Documentation

**Objective**: Comprehensive testing and developer documentation

#### 1.6.1 Integration Test

**File**: `test/integration/approval_flow_test.go`

```go
package integration

import (
	"context"
	"testing"
	"time"
	
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextApprovalFlow(t *testing.T) {
	// Setup
	ctx := context.Background()
	testEnv := setupTestEnvironment(t)
	defer testEnv.Cleanup()
	
	// Step 1: Trigger workflow with approval
	orchestrationID := triggerContentCreatorWorkflow(t, testEnv, map[string]interface{}{
		"business_name": "Test Business",
		"business_type": "SaaS",
	})
	
	// Step 2: Wait for approval task to be created
	var approvalTask *types.ApprovalTask
	require.Eventually(t, func() bool {
		tasks, err := testEnv.ApprovalRepo.ListPendingApprovalTasks(ctx, "test_client")
		if err != nil {
			t.Logf("Error listing tasks: %v", err)
			return false
		}
		if len(tasks) > 0 {
			approvalTask = tasks[0]
			return true
		}
		return false
	}, 30*time.Second, 1*time.Second, "Approval task not created")
	
	// Step 3: Verify task data
	assert.Equal(t, types.ContentTypeText, approvalTask.ContentType)
	assert.Equal(t, types.ApprovalStatusPending, approvalTask.Status)
	assert.NotEmpty(t, approvalTask.TaskData["llm_output"])
	
	// Step 4: Approve with edits via API
	editedContent := "This is the edited hero section!"
	err := testEnv.APIClient.ApproveTask(approvalTask.ID, map[string]interface{}{
		"action":         "edit",
		"edited_content": editedContent,
		"approved_by":    "test_user",
	})
	require.NoError(t, err)
	
	// Step 5: Wait for workflow completion
	var finalState *orchestration.OrchestrationState
	require.Eventually(t, func() bool {
		state, err := testEnv.StateRepo.GetOrchestrationState(ctx, orchestrationID)
		if err != nil {
			return false
		}
		if state.Status == orchestration.StatusCompleted {
			finalState = state
			return true
		}
		return false
	}, 30*time.Second, 1*time.Second, "Workflow did not complete")
	
	// Step 6: Verify final result contains edited content
	heroData := finalState.CollectedData["generate_hero_content"].(map[string]interface{})
	assert.Equal(t, editedContent, heroData["result"])
	assert.True(t, heroData["edited"].(bool))
	assert.Equal(t, "test_user", heroData["approved_by"])
}

func TestImageApprovalFlow(t *testing.T) {
	// Similar test for image approval with replacement
	// ... implementation ...
}
```

#### 1.6.2 Developer Documentation

**File**: `docs/HUMAN_IN_THE_LOOP.md`

```markdown
# Human-in-the-Loop (HITL) System

## Overview

The HITL system enables workflows to pause for human approval of agent outputs. This is implemented via specialized `human-reviewer` agents that handle the approval orchestration.

## Architecture

```
Parent Workflow
↓
Spawn human-reviewer agent
↓
Create approval task (DB)
↓
Pause workflow (PAUSED_FOR_HUMAN status)
↓
[Human reviews via API]
↓
Submit approval (API → Kafka)
↓
Resume workflow with approved content
↓
Continue to next step
```

## Content Types Supported

### Text (LLM Outputs)
- Direct text editing
- Approve/reject decisions
- Full edit history

### Images (Generated Images)
- View generated image
- Replace with uploaded image
- Pre-signed URL upload

## Usage

### 1. Add Approval to Workflow

```yaml
steps:
  generate_content:
    action: execute_llm_prompt
    next_step: review_content
  
  review_content:
    action: spawn_agent
    config:
      agent_type: human-reviewer
      approval_config:
        target_step: generate_content
        capabilities:
          can_edit_directly: true
    next_step: update_approved
  
  update_approved:
    action: merge_data
    config:
      source_step: review_content
      target_step: generate_content
    next_step: complete
```

### 2. Client Polling

```python
import requests
import time

def poll_for_approvals(client_id):
    while True:
        # Get pending approvals
        response = requests.get(
            f"https://api.example.com/api/v1/approvals",
            params={"client_id": client_id}
        )
        
        approvals = response.json()["approvals"]
        
        for approval in approvals:
            # Display to user
            display_approval(approval)
            
            # Get user decision
            decision = get_user_decision(approval)
            
            # Submit approval
            submit_approval(approval["approval_id"], decision)
        
        time.sleep(5)  # Poll every 5 seconds

def submit_approval(approval_id, decision):
    if decision["action"] == "edit":
        requests.post(
            f"https://api.example.com/api/v1/approvals/{approval_id}/approve",
            json={
                "action": "edit",
                "edited_content": decision["edited_content"],
                "approved_by": "user@example.com"
            }
        )
    elif decision["action"] == "replace_image":
        # Get upload URL
        upload_response = requests.post(
            f"https://api.example.com/api/v1/approvals/{approval_id}/upload-url"
        )
        upload_url = upload_response.json()["upload_url"]
        
        # Upload image
        with open(decision["image_path"], "rb") as f:
            requests.put(upload_url, data=f)
        
        # Approve with uploaded image
        requests.post(
            f"https://api.example.com/api/v1/approvals/{approval_id}/approve",
            json={
                "action": "replace",
                "uploaded_path": upload_response.json()["upload_path"],
                "approved_by": "user@example.com"
            }
        )
```

### 3. API Reference

#### List Pending Approvals
```
GET /api/v1/approvals?client_id={client_id}

Response:
{
  "approvals": [
    {
      "approval_id": "uuid",
      "orchestration_id": "uuid",
      "step_name": "generate_hero",
      "content_type": "text",
      "status": "PENDING",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

#### Get Approval Details
```
GET /api/v1/approvals/{approval_id}

Response:
{
  "approval_id": "uuid",
  "orchestration_id": "uuid",
  "step_name": "generate_hero",
  "content_type": "text",
  "status": "PENDING",
  "review_data": {
    "llm_output": "Amazing hero section!",
    "prompt_used": "Write compelling hero...",
    "model": "claude-3-5-sonnet"
  },
  "capabilities": {
    "can_approve": true,
    "can_reject": true,
    "can_edit_directly": true
  }
}
```

#### Approve Task
```
POST /api/v1/approvals/{approval_id}/approve

Request:
{
  "action": "edit",
  "edited_content": "Incredible hero section!",
  "notes": "Changed tone",
  "approved_by": "user@example.com"
}

Response:
{
  "status": "approved",
  "task_id": "uuid"
}
```

## Database Schema

See `platform/migrations/004_approval_system.sql` for full schema.

Key tables:
- `approval_tasks`: Current state of approval tasks
- `approval_versions`: Complete edit history

## Extending

### Adding New Content Types

1. Add content type to `types/approval_types.go`
2. Implement task data structure (e.g., `VideoApprovalData`)
3. Implement response structure (e.g., `VideoApprovalResponse`)
4. Update `buildTaskData()` and `extractApprovedContent()` in actions
5. Update API handler capabilities

### Adding Agent-Based Approval

Future enhancement: Replace human reviewer with AI agent reviewer.

See Phase 2 plan for details.
```

---

## Summary

This implementation plan provides:

1. **Database schema** for approval tasks and version history
2. **Go types** for type-safe approval handling
3. **Specialized human-reviewer agent** for approval orchestration
4. **Workflow integration** with explicit approval steps
5. **RESTful API** for client polling and submission
6. **Coordinator modifications** to handle pause/resume
7. **Comprehensive tests** and documentation

### Key Files Created/Modified:

**New Files:**
- `platform/migrations/004_approval_system.sql`
- `platform/orchestration/types/approval_types.go`
- `platform/orchestration/approval_repository.go`
- `platform/orchestration/actions/approval_actions.go`
- `platform/orchestration/actions/data_actions.go`
- `platform/api/approval_handler.go`
- `configs/agents/human-reviewer.yaml`
- `docs/HUMAN_IN_THE_LOOP.md`
- `test/integration/approval_flow_test.go`

**Modified Files:**
- `platform/orchestration/coordinator.go`
- `platform/agentbase/agent_registry.go`

### Implementation Order:

1. Database migrations (1.1)
2. Core types and repository (1.1-1.2)
3. Approval actions (1.2.3)
4. Agent configuration (1.2.1-1.2.2)
5. API layer (1.4)
6. Coordinator integration (1.5)
7. Testing (1.6)

### Estimated Effort:

- Phase 1.1: 4 hours
- Phase 1.2: 8 hours
- Phase 1.3: 4 hours
- Phase 1.4: 6 hours
- Phase 1.5: 4 hours
- Phase 1.6: 6 hours

**Total: ~32 hours** (4 development days)