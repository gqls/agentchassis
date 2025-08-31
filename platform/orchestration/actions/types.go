// platform/orchestration/actions/types.go
package actions

import (
	"context"
	"database/sql"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

// ActionHandler is the function signature for action handlers
type ActionHandler func(context.Context, ActionParams) (interface{}, error)

// MessageTracer interface to avoid cyclic import
type MessageTracer interface {
	TraceMessage(execCtx interface{}, direction, topic string, payload interface{})
	TraceAwaitedSteps(execCtx interface{}, awaitedSteps []string, action string)
	TraceError(execCtx interface{}, err error, context string)
	DumpTrace(correlationID string)
}

// ActionParams contains all parameters an action might need
type ActionParams struct {
	Context         context.Context
	Headers         map[string]string
	StepConfig      models.Step
	InputData       []byte
	CollectedData   map[string]interface{}
	SagaCoordinator interface{}
	Producer        kafka.Producer
	DB              *sql.DB
	Logger          *zap.Logger
	Tracer          MessageTracer // interface
	AgentType       string
	CurrentStep     string
}
