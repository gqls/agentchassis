// FILE: platform/orchestration/actions/dispatch_actions.go
//
// DispatchAgentAction provides a remote-cluster alternative to SpawnAgentAction.
// It reuses the same helper functions (extractSpawnConfiguration, setupAgentTopics,
// createAgentInDBFromDefinition, sendInitializationMessage, buildSpawnResult, etc.)
// but instead of calling spawnAgentKubernetesJobFromDefinition (which talks to the
// local K8s API), it publishes a dispatch request to a Kafka topic.
//
// A lightweight dispatcher service running in the target cluster consumes these
// messages and creates local K8s Jobs. The parent agent doesn't know or care
// where the child runs — it communicates through the same Kafka topics either way.
//
// Usage in workflows:
//
//	"spawn_researcher": {
//	    "action": "dispatch_agent",
//	    "config": {
//	        "agent_type": "content-researcher",
//	        "role": "researcher",
//	        "target_cluster": "cluster-b"
//	    },
//	    "next_step": "call_researcher"
//	}
//
// If "target_cluster" is omitted, it defaults to "default".
// All other config fields (agent_type, role, etc.) are identical to spawn_agent.

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DispatchRequest is the message payload sent to the dispatch topic.
// The remote dispatcher uses this to create a K8s Job in its local cluster.
type DispatchRequest struct {
	// Agent identity
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	AgentName string `json:"agent_name"`
	Role      string `json:"role"`
	ClientID  string `json:"client_id"`

	// Container spec — from agent_definitions
	ImageRepository string          `json:"image_repository"`
	ImageTag        string          `json:"image_tag"`
	Command         []string        `json:"command"`
	Resources       json.RawMessage `json:"resources"`
	HealthConfig    json.RawMessage `json:"health_config"`
	EnvVars         json.RawMessage `json:"env_vars"`
	Category        string          `json:"category"`

	// Kafka topics for the spawned agent
	RequestsTopic        string `json:"requests_topic"`
	ResponsesTopic       string `json:"responses_topic"`
	ParentResponsesTopic string `json:"parent_responses_topic"`

	// Target cluster identifier
	TargetCluster string `json:"target_cluster"`

	// Infrastructure config — the dispatcher will use these to build
	// the DATABASE_URL and KAFKA_BROKERS env vars for the spawned agent.
	// If empty, the dispatcher uses its own local defaults.
	KafkaBrokers    string `json:"kafka_brokers,omitempty"`
	DatabaseHost    string `json:"database_host,omitempty"`
	DatabasePort    string `json:"database_port,omitempty"`
	DatabaseUser    string `json:"database_user,omitempty"`
	DatabaseName    string `json:"database_name,omitempty"`
	TemplatesDBHost string `json:"templates_db_host,omitempty"`
	TemplatesDBPort string `json:"templates_db_port,omitempty"`
	TemplatesDBUser string `json:"templates_db_user,omitempty"`
	TemplatesDBName string `json:"templates_db_name,omitempty"`

	// Timestamp for dispatch tracking
	DispatchedAt string `json:"dispatched_at"`
}

const (
	// DispatchRequestsTopic is the topic the dispatcher listens on
	DispatchRequestsTopic = "system.dispatch.requests"

	// DispatchResponsesTopic is where the dispatcher confirms job creation
	DispatchResponsesTopic = "system.dispatch.responses"

	// DefaultRemoteStartupWait is longer than local (5s) to account for
	// cross-cluster Kafka latency + remote K8s scheduling
	DefaultRemoteStartupWait = 12 * time.Second

	// DefaultRemoteConsumerWait is the second wait after init message
	DefaultRemoteConsumerWait = 8 * time.Second
)

// DispatchAgentAction dispatches agent creation to a remote cluster via Kafka.
//
// Steps 1-6 and 8-11 are identical to SpawnAgentAction.
// Step 7 replaces the local K8s Job call with a Kafka dispatch message.
//
// NOTE: variable names are intentionally kept identical to SpawnAgentAction
// so the two functions stay in sync during maintenance.
func DispatchAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("DispatchAgentAction starting",
		zap.Any("config", params.StepConfig),
		zap.Any("headers", params.Headers))

	// --- Steps 1-6: identical to SpawnAgentAction ---

	// 1. Validate and extract configuration
	agentType, role, clientID, _, err := extractSpawnConfiguration(params)
	if err != nil {
		return nil, fmt.Errorf("configuration extraction failed in dispatch actions: %w", err)
	}

	// 2. Generate agent identities
	agentID := uuid.New().String()
	agentName := GenerateAgentName(agentType, role)

	// 3. Create agent hierarchy tracking
	subtreeInfo := createSubtreeInfo(agentID, agentName, agentType, params.ExecutionContext.FromAgentID)

	// 4. Get agent definition from database
	agentDef, err := getAgentDefinition(ctx, params.DB, agentType, params.Logger)
	if err != nil {
		params.Logger.Error("Failed to get agent definition",
			zap.String("agent_type", agentType),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get agent definition: %w", err)
	}

	// 5. Create agent in database (in the local/primary DB — same as spawn)
	if err := createAgentInDBFromDefinition(ctx, params, agentID, agentDef, clientID); err != nil {
		params.Logger.Error("Failed to create agent in database",
			zap.String("agent_id", agentID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create agent in DB: %w", err)
	}

	// 6. Setup topic configuration (creates Kafka topics on the shared cluster)
	childRequestsTopic, childResponsesTopic, parentResponsesTopic, stableIdentity, err := setupAgentTopics(ctx, params, agentType, agentID)
	if err != nil {
		params.Logger.Error("Failed to setup topics",
			zap.String("agent_id", agentID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to setup topics: %w", err)
	}

	// --- Step 7: DISPATCH instead of local K8s Job ---

	targetCluster := "default"
	if cluster, ok := params.StepConfig.Config["target_cluster"].(string); ok && cluster != "" {
		targetCluster = cluster
	}

	// bugs_open/066: the remote spawner builds its pod from these two fields,
	// so the correction has to happen HERE — on the far side there is no
	// agent_definitions row to consult and no chassis image to compare against.
	dispatchImage := resolveAgentImage(ctx, agentDef, params.Logger)

	dispatchReq := DispatchRequest{
		AgentID:              agentID,
		AgentType:            agentType,
		AgentName:            agentName,
		Role:                 role,
		ClientID:             clientID,
		ImageRepository:      dispatchImage.Repository,
		ImageTag:             dispatchImage.Tag,
		Command:              agentDef.Command,
		Resources:            agentDef.Resources,
		HealthConfig:         agentDef.HealthConfig,
		EnvVars:              agentDef.EnvVars,
		Category:             agentDef.Category,
		RequestsTopic:        childRequestsTopic,
		ResponsesTopic:       childResponsesTopic,
		ParentResponsesTopic: parentResponsesTopic,
		TargetCluster:        targetCluster,
		// Pass infra config so the dispatcher can build env vars for the agent.
		// The dispatcher may override these with its own local values.
		KafkaBrokers:    os.Getenv("KAFKA_BROKERS"),
		DatabaseHost:    os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_HOST"),
		DatabasePort:    os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_PORT"),
		DatabaseUser:    os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_USER"),
		DatabaseName:    os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_DB_NAME"),
		TemplatesDBHost: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_HOST"),
		TemplatesDBPort: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_PORT"),
		TemplatesDBUser: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_USER"),
		TemplatesDBName: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_DB_NAME"),
		DispatchedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	dispatchJSON, err := json.Marshal(dispatchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dispatch request: %w", err)
	}

	headers := map[string]string{
		"correlation_id":   params.ExecutionContext.CorrelationID,
		"orchestration_id": params.ExecutionContext.OrchestrationID,
		"agent_id":         agentID,
		"agent_type":       agentType,
		"target_cluster":   targetCluster,
		"message_type":     "dispatch_request",
	}

	params.Producer.Produce(ctx, DispatchRequestsTopic, headers,
		[]byte(agentID), dispatchJSON)

	params.Logger.Info("Dispatched agent to remote cluster",
		zap.String("agent_id", agentID),
		zap.String("agent_type", agentType),
		zap.String("target_cluster", targetCluster),
		zap.String("dispatch_topic", DispatchRequestsTopic),
		zap.String("requests_topic", childRequestsTopic),
		zap.String("responses_topic", childResponsesTopic),
		zap.String("parent_responses_topic", parentResponsesTopic))

	// --- Steps 8-11: identical to SpawnAgentAction (with longer waits) ---

	// 8. Generate the initialization request ID
	initRequestID := uuid.New().String()

	// 9. Pre-register the awaited request
	if params.DB != nil {
		if err := preRegisterAwaitedRequest(ctx, params, initRequestID, agentID, agentType,
			childRequestsTopic, parentResponsesTopic); err != nil {
			params.Logger.Warn("Failed to pre-register awaited request - response matching may fail",
				zap.String("request_id", initRequestID),
				zap.Error(err))
		} else {
			params.Logger.Info("Pre-registered awaited request before sending init message",
				zap.String("request_id", initRequestID),
				zap.String("orchestration_id", params.ExecutionContext.OrchestrationID))
		}
	}

	// 10. Wait for remote pod startup
	// Longer than local (5s) because: Kafka cross-cluster latency + dispatcher
	// processing + remote K8s scheduling + image pull (if not cached)
	startupWait := DefaultRemoteStartupWait
	if customWait, ok := params.StepConfig.Config["startup_wait_seconds"].(float64); ok && customWait > 0 {
		startupWait = time.Duration(customWait) * time.Second
	}

	params.Logger.Info("Waiting for remote pod startup...",
		zap.String("agent_id", agentID),
		zap.String("agent_type", agentType),
		zap.String("target_cluster", targetCluster),
		zap.Duration("wait", startupWait))

	time.Sleep(startupWait)

	// 11. Send initialization message
	if err := sendInitializationMessage(ctx, params, agentID, agentName, agentType, role,
		initRequestID, childRequestsTopic, parentResponsesTopic); err != nil {
		params.Logger.Error("Failed to send initialization message",
			zap.String("agent_id", agentID),
			zap.Error(err))
		// Don't fail — agent might still come up and process it
	}

	// 12. Wait for Kafka consumers to start on the remote agent
	consumerWait := DefaultRemoteConsumerWait
	if customWait, ok := params.StepConfig.Config["consumer_wait_seconds"].(float64); ok && customWait > 0 {
		consumerWait = time.Duration(customWait) * time.Second
	}

	params.Logger.Info("Initialization message sent, waiting for remote Kafka consumers...",
		zap.String("agent_id", agentID),
		zap.String("agent_type", agentType),
		zap.Duration("wait", consumerWait))

	time.Sleep(consumerWait)

	// 13. Build and return result — same structure as SpawnAgentAction
	return buildSpawnResult(agentID, agentName, agentType, role, initRequestID,
		childRequestsTopic, childResponsesTopic, parentResponsesTopic,
		stableIdentity, subtreeInfo), nil
}
