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

// ActionParams contains all parameters an action might need
type ActionParams struct {
	Context       context.Context
	Headers       map[string]string
	StepConfig    models.Step
	InputData     []byte
	CollectedData map[string]interface{}
	Producer      kafka.Producer
	DB            *sql.DB
	Logger        *zap.Logger
	AgentType     string
	CurrentStep   string
}
