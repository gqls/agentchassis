// platform/orchestration/actions/types.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// ActionHandler is the function signature for action handlers
type ActionHandler func(context.Context, ActionParams) (interface{}, error)

// Producer interface to avoid importing kafka or messaging directly
type Producer interface {
	Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error
}

// MessageTracer interface
type MessageTracer interface {
	TraceMessage(execCtx interface{}, direction, topic string, payload interface{})
}

// ActionFunc is the signature for action handlers
type ActionFunc func(context.Context, ActionParams) (interface{}, error)

// ActionParams contains all parameters an action might need
type ActionParams struct {
	Context          context.Context
	ExecutionContext *types.ExecutionContext
	Headers          map[string]string
	StepConfig       models.Step
	InputData        []byte
	CollectedData    map[string]interface{}
	SagaCoordinator  interface{}
	Producer         kafka.Producer
	DB               *sql.DB
	Logger           *zap.Logger
	Tracer           types.MessageTracer // interface
	AgentType        string
	CurrentStep      string
}

// AgentDefinition represents an agent's configuration from the database
type AgentDefinition struct {
	ID              int                    `json:"id"`
	Type            string                 `json:"type"`
	DisplayName     string                 `json:"display_name"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category"`
	ImageRepository string                 `json:"image_repository"`
	ImageTag        string                 `json:"image_tag"`
	Resources       json.RawMessage        `json:"resources"`
	DefaultConfig   map[string]interface{} `json:"default_config"`
	Capabilities    json.RawMessage        `json:"capabilities"`
	Topics          json.RawMessage        `json:"topics"`
	HealthConfig    json.RawMessage        `json:"health_config"`
	EnvVars         json.RawMessage        `json:"env_vars"`
	IsActive        bool                   `json:"is_active"`
}
