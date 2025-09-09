// platform/orchestration/actions/types.go
package actions

import (
	"context"
	"database/sql"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// ActionHandler is the function signature for action handlers
type ActionHandler func(context.Context, ActionParams) (interface{}, error)

// ActionFunc is the standard action function signature
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
