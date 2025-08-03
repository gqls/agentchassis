// platform/orchestration/actions/types.go
package actions

import (
	"context"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/kafka"
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
	AgentType     string
	CurrentStep   string
}
