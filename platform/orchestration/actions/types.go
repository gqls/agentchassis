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
	ID              string                 `db:"id"`
	Type            string                 `db:"type"`
	DisplayName     string                 `db:"display_name"`
	Description     string                 `db:"description"`
	Category        string                 `db:"category"`
	ImageRepository string                 `db:"image_repository"`
	ImageTag        string                 `db:"image_tag"`
	Command         []string               `db:"command"`
	Resources       json.RawMessage        `db:"resources"`
	DefaultConfig   map[string]interface{} `db:"default_config"`
	Capabilities    json.RawMessage        `db:"capabilities"`
	Topics          json.RawMessage        `db:"topics"`
	HealthConfig    json.RawMessage        `db:"health_config"`
	EnvVars         json.RawMessage        `db:"env_vars"`
	IsActive        bool                   `db:"is_active"`
}

// AgentMatch represents a discovered agent
type AgentMatch struct {
	AgentID          string
	AgentType        string
	AgentName        string
	Capabilities     []string
	PerformanceScore float64
}

// AgentRecommendation represents a recommended agent for a task
type AgentRecommendation struct {
	AgentType            string
	DisplayName          string
	Category             string
	Capabilities         []string
	PerformanceScore     float64
	RecommendationReason string
}

// ResourceSpec represents Kubernetes resource requirements
type ResourceSpec struct {
	Requests map[string]string `json:"requests"`
	Limits   map[string]string `json:"limits"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	LivenessPath        string `json:"liveness_path"`
	ReadinessPath       string `json:"readiness_path"`
	Port                int    `json:"port"`
	InitialDelaySeconds int    `json:"initial_delay_seconds"`
}

// TopicConfig represents Kafka topic configuration
type TopicConfig struct {
	Process  string `json:"process"`
	Response string `json:"response"`
	Error    string `json:"error"`
	DLQ      string `json:"dlq"`
}

// EnvVar represents an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// EnhancedDiscovery wraps database connections and provides discovery methods
type EnhancedDiscovery struct {
	db interface{} // Can be *sql.DB or *pgxpool.Pool
}

// AgentInfo represents basic agent information
type AgentInfo struct {
	ID string
}

// ComponentTemplate is a helper struct for our DB query.
type ComponentTemplate struct {
	ID           string          `db:"id"`
	Name         string          `json:"name"`
	Category     string          `json:"category"`
	HTMLTemplate string          `db:"html_template"`
	InputSchema  json.RawMessage `db:"input_schema"`
	Function     string          `db:"function"`
}
