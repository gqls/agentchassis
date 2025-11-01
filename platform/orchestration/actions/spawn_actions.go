// FILE: platform/orchestration/actions/spawn_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/discovery"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	// Kubernetes imports
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// SpawnAgentAction spawns a single agent - supports both K8s Job creation and message sending with hierarchy tracking
// Main entry point, now orchestrates smaller functions
func SpawnAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Log entry
	logSpawnStart(params)

	params.Logger.Info("in SpawnAgentAction check request ids",
		zap.Any("DEBUGaa: params.ExecCtx with requestid and replyto request id, which is which and what is right", params.ExecutionContext))

	// 1. Validate and extract configuration
	agentType, role, clientID, _, err := extractSpawnConfiguration(params)
	if err != nil {
		return nil, fmt.Errorf("configuration extraction failed in spawn actions: %w", err)
	}

	// 2. Generate agent identities - use simple values
	agentID := uuid.New().String()
	agentName := GenerateAgentName(agentType, role)

	// 3. Create agent hierarchy tracking
	subtreeInfo := createSubtreeInfo(agentID, agentName, agentType, params.ExecutionContext.FromAgentID)

	// 4. Get agent definition from database - use existing function
	agentDef, err := getAgentDefinition(ctx, params.DB, agentType, params.Logger)
	if err != nil {
		params.Logger.Error("Failed to get agent definition",
			zap.String("agent_type", agentType),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get agent definition: %w", err)
	}

	// 5. Create agent in database - use existing function
	if err := createAgentInDBFromDefinition(ctx, params, agentID, agentDef, clientID); err != nil {
		params.Logger.Error("Failed to create agent in database",
			zap.String("agent_id", agentID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create agent in DB: %w", err)
	}

	// 6. Setup topic configuration
	childRequestsTopic, childResponsesTopic, parentResponsesTopic, stableIdentity, err := setupAgentTopics(ctx, params, agentType, agentID)
	if err != nil {
		params.Logger.Error("Failed to setup topics",
			zap.String("agent_id", agentID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to setup topics: %w", err)
	}

	// 7. Spawn Kubernetes job - use existing function
	jobName, err := spawnAgentKubernetesJobFromDefinition(
		ctx,
		agentID,
		agentDef,
		clientID,
		childRequestsTopic,
		childResponsesTopic,
		parentResponsesTopic,
		params.Logger,
	)
	if err != nil {
		params.Logger.Error("Failed to spawn K8s job",
			zap.String("agent_id", agentID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to spawn K8s job: %w", err)
	}

	params.Logger.Info("Successfully spawned K8s job",
		zap.String("job_name", jobName),
		zap.String("agent_id", agentID))

	// 8. Always send an initialization message to let the agent know it's been spawned
	// This is separate from the actual task data which CallAgentAction will send later

	// Generate a NEW request ID for the initialization message
	initRequestID := uuid.New().String()
	if err := sendInitializationMessage(ctx, params, agentID, agentName, agentType, role,
		initRequestID, childRequestsTopic, parentResponsesTopic); err != nil {
		params.Logger.Error("Failed to send initialization message",
			zap.String("agent_id", agentID),
			zap.Error(err))
		// Don't fail the spawn if message send fails - agent is already created
	}

	// 11. Build and return comprehensive result
	return buildSpawnResult(agentID, agentName, agentType, role, initRequestID,
		childRequestsTopic, childResponsesTopic, parentResponsesTopic,
		stableIdentity, subtreeInfo), nil
}

// Configuration extraction - using existing structs, no new types
func extractSpawnConfiguration(params ActionParams) (agentType, role, clientID string, sendInitData bool, err error) {
	config := params.StepConfig.Config

	// Extract agent type (required)
	agentType, ok := config["agent_type"].(string)
	if !ok || agentType == "" {
		return "", "", "", false, fmt.Errorf("agent_type is required")
	}

	// Extract role (optional, defaults to step name)
	role, hasRole := config["role"].(string)
	if !hasRole || role == "" {
		role = params.ExecutionContext.StepName
	}

	// Extract client ID
	clientID = params.Headers["client_id"]
	if clientID == "" {
		params.Logger.Warn("No client_id in headers, using default")
		clientID = "demo_client"
	}

	// Check if we should send initialization data
	sendInitData = false
	if send, ok := config["send_init_data"].(bool); ok {
		sendInitData = send
	}

	return agentType, role, clientID, sendInitData, nil
}

// Topic setup - setting up topics for the spawned agent so e.g. parentResponsesTopic should be _this_ agent's RESPONSES_TOPIC
func setupAgentTopics(ctx context.Context, params ActionParams, agentType string, agentID string) (childRequestsTopic, childResponsesTopic, parentResponsesTopic, stableIdentity string, err error) {
	// Get parent's response topic
	parentResponsesTopic = os.Getenv("RESPONSES_TOPIC")
	if parentResponsesTopic == "" {
		//setting up topics for the spawned agent so e.g. parentResponsesTopic should be _this_ agent's RESPONSES_TOPIC
		parentResponsesTopic = params.ExecutionContext.ResponsesTopic
	}

	// Create stable identity for topics
	stableIdentity = kafka.CreateStableIdentity(
		params.ExecutionContext.CorrelationID[:8],
		params.ExecutionContext.OrchestrationID,
		agentType,
		params.CurrentStep,
	)

	// Generate topic names
	childRequestsTopic = fmt.Sprintf("job.%s.requests", stableIdentity)
	childResponsesTopic = fmt.Sprintf("job.%s.responses", stableIdentity)

	params.Logger.Info("Creating child agent topics",
		zap.String("stable_identity", stableIdentity),
		zap.String("requests_topic", childRequestsTopic),
		zap.String("responses_topic", childResponsesTopic),
		zap.String("parent_responses_topic", parentResponsesTopic))

	// Create topics using TopicManager
	if err = createTopics(ctx, params.Logger, childRequestsTopic, childResponsesTopic); err != nil {
		return "", "", "", "", fmt.Errorf("failed to create topics: %w", err)
	}

	return childRequestsTopic, childResponsesTopic, parentResponsesTopic, stableIdentity, nil
}

func createTopics(ctx context.Context, logger *zap.Logger, requestsTopic, responsesTopic string) error {
	kafkaBrokers := getConfiguredKafkaBrokers(logger)
	topicManager := kafka.NewTopicManager(kafkaBrokers, logger)

	// Create requests topic
	requestTopicDef := kafka.TopicDefinition{
		Name:              requestsTopic,
		Partitions:        2,
		ReplicationFactor: 1,
	}

	if err := topicManager.CreateTopic(ctx, requestTopicDef); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create requests topic: %w", err)
		}
	}

	// Wait for requests topic
	if err := topicManager.WaitForTopic(ctx, requestsTopic, logger); err != nil {
		logger.Warn("Requests topic not ready after waiting", zap.Error(err))
	}

	// Create responses topic
	responseTopicDef := kafka.TopicDefinition{
		Name:              responsesTopic,
		Partitions:        2,
		ReplicationFactor: 1,
	}

	if err := topicManager.CreateTopic(ctx, responseTopicDef); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create responses topic: %w", err)
		}
	}

	// Wait for responses topic
	if err := topicManager.WaitForTopic(ctx, responsesTopic, logger); err != nil {
		logger.Warn("Responses topic not ready after waiting", zap.Error(err))
	}

	return nil
}

// Decision functions
func shouldSendInitializationMessage(sendInitData bool) bool {
	// By default, don't send init data during spawn
	// Let CallAgentAction handle sending the actual task
	return sendInitData
}

// Data preparation
func prepareInitializationData(params ActionParams, sendInitData bool) map[string]interface{} {
	// Start with empty data
	initData := make(map[string]interface{})

	// Only include data if explicitly configured to do so
	if sendInitData {
		// Look for input_data in collected data
		if inputData, ok := params.CollectedData["input_data"]; ok {
			initData["input_data"] = inputData
		}

		// Add any agent-specific configuration
		if agentConfig, ok := params.CollectedData["agent_config"]; ok {
			initData["config"] = agentConfig
		}
	} else {
		// Send empty structures to avoid nil issues
		initData["input_data"] = make(map[string]interface{})
		initData["config"] = make(map[string]interface{})
	}

	params.Logger.Info("Prepared initialization data",
		zap.Bool("send_init_data", sendInitData),
		zap.Int("data_fields", len(initData)))

	return initData
}

// Message building and sending
func sendInitializationMessage(ctx context.Context, params ActionParams, agentID, agentName, agentType, role, requestID, childRequestsTopic, parentResponsesTopic string) error {

	params.Logger.Info("Sending initialization message",
		zap.String("I am on agent:", os.Getenv("AGENT_TYPE")),
		zap.String("agent_type", agentType),
		zap.String("agent_name", agentName),
		zap.String("role", role),
		zap.String("request_id", requestID),
		zap.String("reply_to_request_id", params.ExecutionContext.ReplyToRequestID),
		zap.String("reply_to_topic", params.ExecutionContext.ReplyToTopic),
	)
	// Prepare clean initialization data (empty for spawn)
	initData := make(map[string]interface{})

	// Prepare agent configuration
	agentConfig := map[string]interface{}{
		"agent_id":   agentID,
		"agent_type": agentType,
		"agent_name": agentName,
		"role":       role,
	}

	// Use the BuildInitializationRequest helper
	message := datahelpers.BuildInitializationRequest(
		params.ExecutionContext,
		agentType,
		role,
		initData,
		agentConfig,
		params.Logger,
	)

	// Build sender identity - this is the SPAWNER (parent agent)
	sender := types.AgentIdentity{
		AgentID:      params.ExecutionContext.FromAgentID,   // Parent's AgentID
		AgentType:    params.ExecutionContext.FromAgentType, // Parent's type
		Role:         params.ExecutionContext.Sender.Role,
		AgentVersion: params.ExecutionContext.Version,
		PodName:      params.ExecutionContext.ProcessingNode,
	}

	// If we don't have FromAgentType, use environment
	if sender.AgentType == "" {
		sender.AgentType = os.Getenv("AGENT_TYPE")
		if sender.AgentType == "" {
			sender.AgentType = "generic"
		}
	}

	// Override specific fields for this spawn initialization
	message.Headers.Sender = sender

	// Child's identity
	message.Headers.OrchestrationID = agentID
	message.Headers.OrchestrationName = agentName
	message.Headers.ToAgent = agentID
	message.Headers.ToAgentType = agentType

	// Parent context
	message.Headers.ParentOrchestrationID = params.ExecutionContext.OrchestrationID
	message.Headers.ParentOrchestrationName = params.ExecutionContext.OrchestrationName

	params.Logger.Info("Sending initialization message resetting the reply to request id, do I really need to?",
		zap.String("DEBUGaa: do I really need to set replytorequest id here. Headers.ReplyToRequestID before are:", message.Headers.ReplyToRequestID),
		zap.Any("DEBUGaa: in exec ctx look at what requestID (this is now replyto) and also what replyto request id was", params.ExecutionContext),
	)

	// maybe we try and get it from execution context
	if message.Headers.ReplyToRequestID == "" {
		message.Headers.ReplyToRequestID = params.ExecutionContext.RequestID
	}

	// Step context - CRITICAL: Must be populated
	message.Headers.StepID = params.ExecutionContext.StepID
	message.Headers.StepName = params.ExecutionContext.StepName // e.g., "spawn_hero_writer"
	message.Headers.RequestID = requestID
	message.Headers.Action = "initialize"

	// Routing
	message.Headers.ResponsesTopic = parentResponsesTopic
	message.Headers.RequestsTopic = "" // Child will set from environment

	// Set functional role for the child
	message.Headers.FunctionalRole = role

	params.Logger.Info("Sending initialization message to spawned agent",
		zap.String("agent_id", agentID),
		zap.String("agent_name", agentName),
		zap.String("topic", childRequestsTopic),
		zap.String("action", "initialize"),
		zap.String("step_name", message.Headers.StepName),
		zap.String("step_id", message.Headers.StepID),
		zap.String("sender_type", sender.AgentType),
		zap.String("parent_orchestration_id", message.Headers.ParentOrchestrationID),
		zap.String("reply_to_request_id", message.Headers.ReplyToRequestID),
		zap.String("parent responses_topic", message.Headers.ParentResponsesTopic),
	)

	// Marshal message
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	params.Logger.Info("Sending initialization message to spawned agent",
		zap.String("agent_id", agentID),
		zap.String("topic", childRequestsTopic),
		zap.String("reply to request id", message.Headers.ReplyToRequestID),
		zap.String("action", "initialize"))

	// Send with retries
	return sendMessageWithRetries(ctx, params, childRequestsTopic, message.Headers.ToMap(), requestID, messageBytes)
}

func buildInitializationMessage(params ActionParams, agentID, agentName, agentType, role, requestID, parentResponsesTopic, senderType string, initData map[string]interface{}) *types.RequestMessage {
	return &types.RequestMessage{
		Headers: types.RequestHeaders{
			Sender: types.AgentIdentity{
				AgentType: senderType,
				AgentID:   params.ExecutionContext.FromAgentID,
				PodName:   params.ExecutionContext.ProcessingNode,
				Role:      params.ExecutionContext.Sender.Role,
			},
			CorrelationID:           params.ExecutionContext.CorrelationID,
			CorrelationName:         params.ExecutionContext.CorrelationName,
			ClientID:                params.ExecutionContext.ClientID,
			OrchestrationID:         agentID,
			OrchestrationName:       agentName,
			ParentOrchestrationID:   params.ExecutionContext.OrchestrationID,
			ParentOrchestrationName: params.ExecutionContext.OrchestrationName,
			ReplyToRequestID:        params.ExecutionContext.RequestID,
			StepID:                  params.ExecutionContext.StepID,
			StepName:                "spawn_agent",
			RequestID:               requestID,
			RetryVersion:            0,
			FunctionalRole:          role,
			MessageID:               uuid.New().String(),
			MessageType:             "request",
			ToAgent:                 agentID,
			ToAgentType:             agentType,
			Action:                  "initialize",
			Timestamp:               time.Now(),
			FuelBudget:              params.ExecutionContext.FuelBudget - 100,
			TimeoutSeconds:          30,
			ResponsesTopic:          parentResponsesTopic,
			RequestsTopic:           "", // Child will set from environment
		},
		Body: initData,
	}
}

func sendMessageWithRetries(ctx context.Context, params ActionParams, topic string, headers map[string]string, requestID string, messageBytes []byte) error {
	const maxRetries = 10
	const retryDelay = 8 * time.Second

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		err := params.Producer.ProduceWithValidation(
			ctx,
			topic,
			headers,
			[]byte(requestID),
			messageBytes,
		)

		if err == nil {
			params.Logger.Info("Message sent successfully",
				zap.String("topic", topic),
				zap.String("request_id", requestID))
			return nil
		}

		lastErr = err
		params.Logger.Warn("Failed to send message, retrying",
			zap.Int("attempt", i+1),
			zap.Int("max_attempts", maxRetries),
			zap.Error(err))

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// Helper functions
func determineSenderType(params ActionParams) string {
	// Priority order for determining sender type

	// I am the sender I think, so:
	// 1.0 Check env vars
	agentType := os.Getenv("AGENT_TYPE")
	if agentType != "" {
		return agentType
	}

	// 1.1 Check ExecutionContext
	if params.ExecutionContext.FromAgentType != "" {
		return params.ExecutionContext.FromAgentType
	}

	// 2. Check Sender in ExecutionContext
	if params.ExecutionContext.Sender.AgentType != "" {
		return params.ExecutionContext.Sender.AgentType
	}

	// 3. Check agent_config in CollectedData
	if agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{}); ok {
		if agentType, ok = agentConfig["agent_type"].(string); ok && agentType != "" {
			return agentType
		}
	}

	// 4. Default
	return "orchestrator"
}

func createSubtreeInfo(agentID, agentName, agentType, parentAgentID string) *types.SubtreeInfo {
	return &types.SubtreeInfo{
		AgentID:       agentID,
		AgentType:     agentType,
		AgentName:     agentName,
		ParentAgentID: parentAgentID,
		Children:      make(map[string]*types.SubtreeInfo),
		CreatedAt:     time.Now(),
		LastActiveAt:  time.Now(),
		Performance: &types.PerformanceMetrics{
			TasksCompleted:   0,
			TasksFailed:      0,
			AverageLatencyMs: 0,
			FuelConsumed:     0,
			LastUpdated:      time.Now(),
		},
	}
}

func buildSpawnResult(agentID, agentName, agentType, role, requestID,
	childRequestsTopic, childResponsesTopic, parentResponsesTopic,
	stableIdentity string, subtree *types.SubtreeInfo) map[string]interface{} {
	return map[string]interface{}{
		"agent_id":          agentID,
		"agent_name":        agentName,
		"agent_type":        agentType,
		"status":            "initialized",
		"role":              role,
		"request_id":        requestID,
		"await_response":    true,
		"target_agent_type": agentType,
		"subtree_info":      subtree,

		// Topic information
		"requests_topic":         childRequestsTopic,
		"responses_topic":        parentResponsesTopic,
		"parent_responses_topic": parentResponsesTopic,

		// For backward compatibility
		"stable_identity": stableIdentity,
		"topic_sent_to":   childRequestsTopic,

		// Nested result for consistency
		"result": map[string]interface{}{
			"agent_id":   agentID,
			"agent_type": agentType,
			"role":       role,
			"topics": map[string]string{
				"requests":         childRequestsTopic,
				"responses":        parentResponsesTopic,
				"parent_responses": parentResponsesTopic,
			},
		},
	}
}

func logSpawnStart(params ActionParams) {
	current, caller := getFuncInfo(1)
	params.Logger.Info("SpawnAgentAction starting",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.Any("config", params.StepConfig),
		zap.Any("headers", params.Headers))
}

func SpawnAgentActionOld2(ctx context.Context, params ActionParams) (interface{}, error) {
	/*params.Logger.Info("SpawnAgentAction starting",
	zap.Any("config", params.StepConfig),
	zap.Any("params.Headers look for client_id", params.Headers))*/

	// Keep stack trace for debugging
	current, caller := getFuncInfo(1)
	params.Logger.Info("SpawnAgentAction starting",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.Any("config", params.StepConfig),
		zap.Any("params.Headers", params.Headers),
	)

	config := params.StepConfig.Config

	// Validate required fields - KEEP
	agentType, ok := config["agent_type"].(string)
	if !ok || agentType == "" {
		return nil, fmt.Errorf("agent_type is required")
	}

	role, hasRole := config["role"].(string)
	if !hasRole || role == "" {
		role = params.ExecutionContext.StepName // Use step name as fallback
	}

	// Generate unique agent ID and name - KEEP
	agentID := uuid.New().String()
	agentName := GenerateAgentName(agentType, role)

	// Pre-generate request ID
	requestID := params.Headers["request_id"]
	if requestID == "" {
		requestID = uuid.New().String()
		params.Headers["request_id"] = requestID
	}

	params.Logger.Info("Spawning a new agent",
		zap.String("new agent_id", agentID),
		zap.String("new agent_name", agentName),
		zap.String("agent_type", agentType),
		zap.String("role", role),
		zap.String("request_id", requestID))

	// Create subtree info for hierarchy tracking - KEEP
	subtreeInfo := &types.SubtreeInfo{
		AgentID:       agentID,
		AgentType:     agentType,
		AgentName:     agentName,
		ParentAgentID: params.ExecutionContext.FromAgentID,
		Children:      make(map[string]*types.SubtreeInfo),
		CreatedAt:     time.Now(),
		LastActiveAt:  time.Now(),
		Performance: &types.PerformanceMetrics{
			TasksCompleted:   0,
			TasksFailed:      0,
			AverageLatencyMs: 0,
			FuelConsumed:     0,
			LastUpdated:      time.Now(),
		},
	}

	clientID := params.Headers["client_id"]
	params.Logger.Info("looking for client id in SpawnAgentAction",
		zap.String("client_id", clientID))
	if clientID == "" {
		params.Logger.Error("Failed to get client id in SpawnAgentAction reverting to default",
			zap.String("clientID", clientID),
		)
		clientID = "demo_client"
	}

	// Get agent definition
	agentDef, err := getAgentDefinition(ctx, params.DB, agentType, params.Logger)
	if err != nil {
		params.Logger.Error("Failed to get agent definition in SpawnAgentAction",
			zap.String("agent_type", agentType),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get agent definition: %w", err)
	}

	params.Logger.Info("Got agent definition",
		zap.String("agent_type", agentType),
		zap.String("is the client id the original? if its just demo_client then no its been changed", agentType),
		zap.String("image", fmt.Sprintf("%s:%s", agentDef.ImageRepository, agentDef.ImageTag)))

	// Create agent in DB
	if err := createAgentInDBFromDefinition(ctx, params, agentID, agentDef, clientID); err != nil {
		params.Logger.Error("Failed to create agent in database",
			zap.String("agent_id", agentID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create agent in DB: %w", err)
	}

	// this is the new child's parentResponseTopic or the current agents normal response topic
	parentResponsesTopic := params.ExecutionContext.ResponsesTopic
	if parentResponsesTopic == "" {
		// about to spawn child so it needs this agent's normal response topic
		parentResponsesTopic = os.Getenv("RESPONSES_TOPIC")
	}

	// Create stable identity for child's topics
	// Using agentID for uniqueness since it's stable for this spawn
	stableIdentity := kafka.CreateStableIdentity(
		params.ExecutionContext.CorrelationID[:8],
		params.ExecutionContext.OrchestrationID,
		agentType,
		params.CurrentStep) // The step that's spawning this agent

	// Create job-specific topics
	childRequestsTopic := fmt.Sprintf("job.%s.requests", stableIdentity)
	childResponsesTopic := fmt.Sprintf("job.%s.responses", stableIdentity)

	params.Logger.Info("Creating child agent topics",
		zap.String("stable_identity", stableIdentity),
		zap.String("requests_topic", childRequestsTopic),
		zap.String("responses_topic", childResponsesTopic))

	// Get Kafka brokers with proper fallback
	kafkaBrokers := getConfiguredKafkaBrokers(params.Logger)

	// Use the TopicManager to create topics properly
	topicManager := kafka.NewTopicManager(kafkaBrokers, params.Logger)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create the request topic
	requestTopicDef := kafka.TopicDefinition{
		Name:              childRequestsTopic,
		Partitions:        2,
		ReplicationFactor: 1, // Use 1 for quick creation
	}

	err = topicManager.CreateTopic(ctx, requestTopicDef)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		params.Logger.Error("Failed to create requests topic",
			zap.Error(err),
			zap.String("topic", childRequestsTopic))
		// Don't fail here - try to continue anyway
	} else {
		params.Logger.Info("Created or verified requests topic",
			zap.String("topic", childRequestsTopic))
	}

	// wait 10 or so seconds to determine if requests topic is ready
	err = topicManager.WaitForTopic(ctx, childRequestsTopic, params.Logger)
	if err != nil {
		params.Logger.Error("Requests topic never got ready in 10 seconds",
			zap.Error(err),
			zap.String("topic", childRequestsTopic),
		)
	}

	// Create the response topic
	responseTopicDef := kafka.TopicDefinition{
		Name:              childResponsesTopic,
		Partitions:        2,
		ReplicationFactor: 1,
	}

	err = topicManager.CreateTopic(ctx, responseTopicDef)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		params.Logger.Error("Failed to create responses topic.",
			zap.Error(err),
			zap.String("topic", childResponsesTopic))
	} else {
		params.Logger.Info("Created or verified responses topic.",
			zap.String("topic", childResponsesTopic))
	}

	// wait 10 or so seconds to determine if response topic is ready
	err = topicManager.WaitForTopic(ctx, childResponsesTopic, params.Logger)
	if err != nil {
		params.Logger.Error("Responses topic never got ready in 10 seconds",
			zap.Error(err),
			zap.String("topic", childResponsesTopic),
		)
	}

	params.Logger.Info("Topic creation completed, waiting for propagation",
		zap.String("requests_topic", childRequestsTopic),
		zap.String("responses_topic", childResponsesTopic))

	// Spawn K8s job with new topic names
	jobName, err := spawnAgentKubernetesJobFromDefinition(
		ctx,
		agentID,
		agentDef,
		clientID,
		childRequestsTopic,  // Pass as REQUESTS_TOPIC env var
		childResponsesTopic, // Pass as RESPONSES_TOPIC env var
		parentResponsesTopic,
		params.Logger)

	if err != nil {
		params.Logger.Error("Failed to spawn K8s job",
			zap.String("agent_id", agentID),
			zap.String("agent_type", agentType),
			zap.Error(err))
		return nil, fmt.Errorf("failed to spawn K8s job: %w", err)
	}

	params.Logger.Info("Successfully spawned K8s job",
		zap.String("job_name", jobName),
		zap.String("agent_id", agentID))

	// Determine sender type
	senderAgentType := "generic"
	if params.ExecutionContext.FromAgentType != "" {
		senderAgentType = params.ExecutionContext.FromAgentType
	} else if params.ExecutionContext.Sender.AgentType != "" {
		senderAgentType = params.ExecutionContext.Sender.AgentType
	}

	// Check agent_config for override
	if agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{}); ok {
		if configAgentType, ok := agentConfig["agent_type"].(string); ok && configAgentType != "" {
			senderAgentType = configAgentType
		}
	}

	// The initial request often nests the user's data inside {"input_data": {"input_data": ...}}.
	// We prioritize finding this deeper, more specific data structure.
	//var dataToSend interface{}
	/*if nestedData, ok := getNestedInputDataValue(params.CollectedData, "input_data", "input_data"); ok {
		params.Logger.Info("Found nested input_data for spawned agent.", zap.String("path", "input_data.input_data"))
		dataToSend = nestedData
	} else if topLevelData, ok := params.CollectedData["input_data"]; ok {
		// If the deeper structure isn't found, fall back to the top-level input_data.
		params.Logger.Info("Using top-level input_data for spawned agent.", zap.String("path", "input_data"))
		dataToSend = topLevelData
	} else {
		// If no input_data is found at all, send an empty map to avoid errors.
		params.Logger.Warn("No input_data found in CollectedData for spawned agent.")
		dataToSend = make(map[string]interface{})
	}*/

	params.Logger.Info("Determined sender agent type",
		zap.String("sender_type", senderAgentType),
		zap.String("parent_responses_topic, i.e. the response topic of me, the sender", parentResponsesTopic))

	// Create spawn initialization message
	spawnMessage := &types.RequestMessage{
		Headers: types.RequestHeaders{
			Sender: types.AgentIdentity{
				AgentType: senderAgentType,
				AgentID:   params.ExecutionContext.FromAgentID,
				PodName:   params.ExecutionContext.ProcessingNode,
				Role:      params.ExecutionContext.Sender.Role,
			},
			CorrelationID:   params.ExecutionContext.CorrelationID,
			CorrelationName: params.ExecutionContext.CorrelationName,
			ClientID:        params.ExecutionContext.ClientID,

			// Child's orchestration info
			OrchestrationID:         agentID,
			OrchestrationName:       agentName,
			ParentOrchestrationID:   params.ExecutionContext.OrchestrationID,
			ParentOrchestrationName: params.ExecutionContext.OrchestrationName,
			ReplyToRequestID:        params.ExecutionContext.RequestID,

			StepID:         params.ExecutionContext.StepID,
			StepName:       "spawn_agent",
			RequestID:      requestID,
			RetryVersion:   0,
			FunctionalRole: role,

			MessageID:   uuid.New().String(),
			MessageType: "request",
			ToAgent:     agentID,
			ToAgentType: agentType,
			Action:      "initialize",
			Timestamp:   time.Now(),

			FuelBudget:     params.ExecutionContext.FuelBudget - 100,
			TimeoutSeconds: 30,

			// Tell child where parent listens for responses
			ResponsesTopic: parentResponsesTopic, // i.e. the response topic of the current me, the sender
			// Child's own topics will be set from environment
			RequestsTopic: "",
		},
		Body: map[string]interface{}{
			//"input_data": dataToSend,
			// Send an empty map. The agent's task will be sent later by CallAgentAction.
			"input_data": make(map[string]interface{}),
			"action":     "initialize",
			//"config":     params.CollectedData["agent_config"],
			"config": make(map[string]interface{}),
			"role":   role,
		},
	}

	params.Logger.Info("DEBUGaa: the params.CollectedData input_data in SpawnAgentActions - we should extract",
		zap.Any("Collected Data input_data, spawn agent action", params.CollectedData["input_data"]),
	)

	params.Logger.Info("Sending spawn initialization message",
		zap.String("agent_id", agentID),
		zap.String("to_topic", childRequestsTopic),
		zap.String("parent_expects_responses_on", parentResponsesTopic))

	// Send to child's requests topic
	messageBytes, err := json.Marshal(spawnMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spawn message: %w", err)
	}

	if produceErr := params.Producer.ProduceWithValidation(
		ctx,
		childRequestsTopic, // Send to child's topic
		spawnMessage.Headers.ToMap(),
		[]byte(requestID),
		messageBytes); produceErr != nil {
		return nil, fmt.Errorf("failed to produce message SpawnAgentAction: %w", produceErr)
	}

	var produceErr error
	const maxRetries = 3
	const retryDelay = 5 * time.Second
	for i := 0; i < maxRetries; i++ {
		produceErr = params.Producer.ProduceWithValidation(
			ctx,
			childRequestsTopic, // Send to child's topic
			spawnMessage.Headers.ToMap(),
			[]byte(requestID),
			messageBytes)

		if produceErr == nil {
			break
		}

		if strings.Contains(produceErr.Error(), "Unknown Topic Or Partition") {
			params.Logger.Warn("Failed to send spawn message due to unknown topic:",
				zap.Int("attempt:", i+1),
				zap.Int("max_attempts", maxRetries),
				zap.Duration("delay", retryDelay),
				zap.Error(produceErr),
			)
		}
	}

	if produceErr != nil {
		return nil, fmt.Errorf("failed to send spawn message after %d attempts: %w", maxRetries, err)
	}

	params.Logger.Info("Agent spawn message sent successfully",
		zap.String("agent_id", agentID),
		zap.String("sent_to_topic", childRequestsTopic),
		zap.String("request_id", requestID))

	// Return comprehensive result - KEEP all this info
	return map[string]interface{}{
		"agent_id":          agentID,
		"agent_name":        agentName,
		"agent_type":        agentType,
		"status":            "initialized",
		"role":              role,
		"request_id":        requestID,
		"await_response":    true,
		"target_agent_type": agentType,
		"subtree_info":      subtreeInfo,

		// Use consistent naming
		"requests_topic":         childRequestsTopic,
		"responses_topic":        parentResponsesTopic, // I think this determines where the executeLocalAction will direct next response
		"parent_responses_topic": parentResponsesTopic, //? how do we use this

		// For backward compatibility and debugging
		"stable_identity": stableIdentity,
		"topic_sent_to":   childRequestsTopic,

		// Nested result for consistency
		"result": map[string]interface{}{
			"agent_id":   agentID,
			"agent_type": agentType,
			"role":       role,
			"topics": map[string]string{
				"requests":         childRequestsTopic,
				"responses":        parentResponsesTopic,
				"parent_responses": parentResponsesTopic,
			},
		},
	}, nil
}

// Helper function to get Kafka brokers with fallback
func getConfiguredKafkaBrokers(logger *zap.Logger) []string {
	// Try multiple environment variables
	brokersEnv := os.Getenv("SERVICE_INFRASTRUCTURE_KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = os.Getenv("KAFKA_BROKERS")
	}
	if brokersEnv == "" {
		brokersEnv = os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	}

	if brokersEnv != "" {
		brokers := strings.Split(brokersEnv, ",")
		validBrokers := []string{}
		for _, broker := range brokers {
			broker = strings.TrimSpace(broker)
			if broker != "" && broker != ":9092" {
				validBrokers = append(validBrokers, broker)
			}
		}
		if len(validBrokers) > 0 {
			logger.Info("Using configured Kafka brokers",
				zap.Strings("brokers", validBrokers))
			return validBrokers
		}
	}

	// Fallback to default brokers
	defaultBrokers := []string{
		"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092",
		"personae-kafka-cluster-kafka-bootstrap.kafka:9092",
		"kafka-0.kafka-headless.kafka:9092",
	}

	logger.Warn("No Kafka brokers configured, using defaults",
		zap.Strings("brokers", defaultBrokers))

	return defaultBrokers
}

func SpawnAgentActionOld(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("In SpawnAgentAction")

	params.Logger.Info("SpawnAgentAction starting",
		zap.Any("config", params.StepConfig))

	current, caller := getFuncInfo(1)

	params.Logger.Info("In file spawn_actions.go SpawnAgentAction",
		zap.String("function FUNCTION", current),
		zap.String("called_by CALLED_BY", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	config := params.StepConfig.Config

	agentType, ok := config["agent_type"].(string)
	if !ok || agentType == "" {
		return nil, fmt.Errorf("agent_type is required")
	}

	role, hasRole := config["role"].(string)
	if !hasRole || role == "" {
		// Use step name as fallback
		role = params.ExecutionContext.StepName
	}

	// Generate unique agent ID and name
	agentID := uuid.New().String()
	agentName := GenerateAgentName(agentType, role)

	// Pre-generate request ID (critical for response matching)
	requestID := params.Headers["request_id"]
	if requestID == "" {
		requestID = uuid.New().String()
		params.Headers["request_id"] = requestID
	}

	params.Logger.Info("Spawning agent in spawn_actions SpawnAgentAction:",
		zap.String("agent_id", agentID),
		zap.String("agent_name", agentName),
		zap.String("agent_type", agentType),
		zap.String("role", role),
		zap.String("request_id", requestID),
		zap.String("orchestration_id", requestID),
		zap.String("orchestration_name", requestID),
	)

	// Create subtree info for hierarchy tracking
	subtreeInfo := &types.SubtreeInfo{
		AgentID:       agentID,
		AgentType:     agentType,
		AgentName:     agentName,
		ParentAgentID: params.ExecutionContext.FromAgentID,
		Children:      make(map[string]*types.SubtreeInfo),
		CreatedAt:     time.Now(),
		LastActiveAt:  time.Now(),
		Performance: &types.PerformanceMetrics{
			TasksCompleted:   0,
			TasksFailed:      0,
			AverageLatencyMs: 0,
			FuelConsumed:     0,
			LastUpdated:      time.Now(),
		},
	}

	// Log what mode we're using
	params.Logger.Info("SpawnAgentAction mode",
		zap.String("agent_type", agentType))

	// ALWAYS create K8s job when spawning
	clientID := params.Headers["client_id"]
	if clientID == "" {
		clientID = "demo_client"
	}

	params.Logger.Info("Attempting to create K8s job",
		zap.String("client_id", clientID),
		zap.String("agent_type", agentType))

	// Get agent definition and create K8s job
	agentDef, err := getAgentDefinition(ctx, params.DB, agentType, params.Logger)
	if err != nil {
		params.Logger.Error("Failed to get agent definition",
			zap.String("agent_type", agentType),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get agent definition: %w", err)
	}

	params.Logger.Info("Got agent definition",
		zap.String("agent_type", agentType),
		zap.String("image", fmt.Sprintf("%s:%s", agentDef.ImageRepository, agentDef.ImageTag)),
		zap.Bool("is_active", agentDef.IsActive))

	// Create agent in DB
	if err := createAgentInDBFromDefinition(ctx, params, agentID, agentDef, clientID); err != nil {
		params.Logger.Error("Failed to create agent in database",
			zap.String("agent_id", agentID),
			zap.Error(err))
		// Don't continue if DB creation fails
		return nil, fmt.Errorf("failed to create agent in DB: %w", err)
	}

	params.Logger.Info("Created agent in database",
		zap.String("agent_id", agentID))

	// Get the actual step name
	stepName := params.ExecutionContext.StepName
	if stepName == "" {
		stepName = "spawn_agent" // fallback
	}

	jobRequestsTopic := fmt.Sprintf("%s.requests", kafka.GenerateJobTopic(
		params.ExecutionContext.CorrelationID,
		params.ExecutionContext.OrchestrationID,
		stepName,
		agentType))

	jobResponsesTopic := fmt.Sprintf("%s.responses", kafka.GenerateJobTopic(
		params.ExecutionContext.CorrelationID,
		params.ExecutionContext.OrchestrationID,
		stepName,
		agentType))

	params.Logger.Info("Topic generated in SpawnAgentAction",
		zap.String("jobResponseTopic", jobResponsesTopic),
		zap.String("jobRequestTopic", jobRequestsTopic),
	)

	kafkaBrokers := strings.Split(os.Getenv("SERVICE_INFRASTRUCTURE_KAFKA_BROKERS"), ",")
	if len(kafkaBrokers) == 0 || kafkaBrokers[0] == "" {
		// Try to get from database config if available
		kafkaBrokers = []string{"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"}
	}

	if err := kafka.CreateJobTopic(kafkaBrokers, jobRequestsTopic, 2); err != nil {
		params.Logger.Warn("Failed to create job requests topic (may already exist)",
			zap.String("topic", jobRequestsTopic),
			zap.Error(err))
		// Don't fail - topic might already exist from a retry
	}

	if err := kafka.CreateJobTopic(kafkaBrokers, jobResponsesTopic, 2); err != nil {
		params.Logger.Warn("Failed to create job responses topic (may already exist)",
			zap.String("topic", jobResponsesTopic),
			zap.Error(err))
		// Don't fail - topic might already exist from a retry
	}

	parentResponsesTopic := params.ExecutionContext.ResponsesTopic
	// Always spawn K8s job
	jobName, err := spawnAgentKubernetesJobFromDefinition(ctx, agentID, agentDef, clientID, jobRequestsTopic, jobResponsesTopic, parentResponsesTopic, params.Logger)
	if err != nil {
		params.Logger.Error("Failed to spawn K8s job",
			zap.String("agent_id", agentID),
			zap.String("agent_type", agentType),
			zap.Error(err))
		// Return error instead of continuing
		return nil, fmt.Errorf("failed to spawn K8s job: %w", err)
	}

	params.Logger.Info("Successfully spawned K8s job",
		zap.String("job_name", jobName),
		zap.String("agent_id", agentID))

	// Determine what action the spawned agent should perform after initializing
	targetAction := "process" // A sensible default
	if action, ok := params.StepConfig.Config["target_action"].(string); ok {
		targetAction = action
	}

	params.Logger.Info("The message body and spawn info",
		zap.String("agent_id", agentID),
		zap.String("agent_name", agentName),
		zap.String("role", role),
		zap.Any("config", params.StepConfig.Config["config"]),
		zap.Any("params inputdata", params.CollectedData["input_data"]),
		zap.Any("params input action", params.CollectedData["input_action"]),
		zap.Any("targetAction", targetAction),
	)

	// Debug what's in CollectedData
	params.Logger.Info("DEBUGaa: CollectedData contents before spawn",
		zap.Any("all_collected_data", params.CollectedData),
		zap.Any("input_data", params.CollectedData["input_data"]),
		zap.Any("EXECUTION CONTEXTAA", params.ExecutionContext),
	)

	senderAgentType := "genericdefault" // Default

	// Try ExecutionContext first
	if params.ExecutionContext.FromAgentType != "" {
		senderAgentType = params.ExecutionContext.FromAgentType
	} else if params.ExecutionContext.Sender.AgentType != "" {
		senderAgentType = params.ExecutionContext.Sender.AgentType
	}

	// Override with agent_config if available and valid
	if agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{}); ok {
		if configAgentType, ok := agentConfig["agent_type"].(string); ok && configAgentType != "" {
			senderAgentType = configAgentType
			params.Logger.Info("Using agent type from config",
				zap.String("agent_type", configAgentType))
		}
	}

	params.Logger.Info("Determined sender agent type",
		zap.String("sender_type", senderAgentType),
		zap.String("responses_topic", fmt.Sprintf("system.agent.%s.responses", senderAgentType)))

	params.Logger.Info("Determining sender type for response topic",
		zap.String("ExecutionContext.FromAgentType", params.ExecutionContext.FromAgentType),
		zap.String("ExecutionContext.Sender.AgentType", params.ExecutionContext.Sender.AgentType),
	)

	// LOGGING only Log agent_config.agent_type safely
	if agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{}); ok {
		if agentTypeFromConfig, ok := agentConfig["agent_type"].(string); ok {
			params.Logger.Info("Agent type from config",
				zap.String("agent_config.agent_type", agentTypeFromConfig))
		}
	}

	// The most reliable source of the CURRENT agent's type is its own agent_config
	// that was loaded by the processor.
	if agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{}); ok {
		if currentAgentType, ok := agentConfig["agent_type"].(string); ok && currentAgentType != "" {
			senderAgentType = currentAgentType
		}
	}

	if senderAgentType == "" {
		senderAgentType = "generic" // fallback
	}

	params.Logger.Info("Using sender agent type",
		zap.String("sender_type", senderAgentType))

	params.Logger.Info("DEBUGaa: SpawnAgentAction - setting orchestration ID to agent ID",
		zap.String("orchestration ID", agentID),
		zap.String("orchestration Name", agentName),
	)

	// Create spawn message
	spawnMessage := &types.RequestMessage{
		Headers: types.RequestHeaders{
			Sender: types.AgentIdentity{
				AgentType: senderAgentType,
				AgentID:   params.ExecutionContext.FromAgentID,
				PodName:   params.ExecutionContext.ProcessingNode,
				Role:      params.ExecutionContext.Sender.Role,
			},
			CorrelationID:         params.ExecutionContext.CorrelationID,
			CorrelationName:       params.ExecutionContext.CorrelationName,
			ClientID:              params.ExecutionContext.ClientID,
			OrchestrationID:       agentID,
			OrchestrationName:     agentName,
			ParentOrchestrationID: params.ExecutionContext.OrchestrationID,
			StepID:                params.ExecutionContext.StepID,
			StepName:              "spawn_agent",
			RequestID:             requestID,
			RetryVersion:          0,
			FunctionalRole:        role, // This is the role for the NEW agent being spawned
			MessageID:             uuid.New().String(),
			MessageType:           "request",
			ToAgent:               agentID,
			ToAgentType:           agentType,
			Action:                "initialize",
			Timestamp:             time.Now(),
			FuelBudget:            params.ExecutionContext.FuelBudget - 100,
			TimeoutSeconds:        30,
			ResponsesTopic:        jobResponsesTopic,
		},
		Body: map[string]interface{}{
			// Always wrap the payload in the standard key
			"input_data": params.CollectedData["input_data"],
			"action":     "initialize", // What the spawned agent should do
			"config":     params.CollectedData["agent_config"],
			"role":       role,
		},
	}

	params.Logger.Info("DEBUGaa: 1 SpawnAgentAction Request Message - spawnMessage - for spawn",
		zap.Any("spawn Message aa", spawnMessage),
	)

	// Send spawn message to agent's request topic
	//targetTopic := fmt.Sprintf("system.agent.%s.requests", agentType)
	targetTopic := jobRequestsTopic

	messageBytes, err := json.Marshal(spawnMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spawn message: %w", err)
	}

	if err := params.Producer.Produce(ctx, targetTopic, spawnMessage.Headers.ToMap(), []byte(requestID), messageBytes); err != nil {
		return nil, fmt.Errorf("failed to send spawn message: %w", err)
	}

	params.Logger.Info("Agent spawn message sent",
		zap.String("agent_id", agentID),
		zap.String("topic", targetTopic),
		zap.String("request_id", requestID))

	// Return result with await_response flag
	return map[string]interface{}{
		"agent_id":          agentID,
		"agent_name":        agentName,
		"agent_type":        agentType,
		"status":            "initialized",
		"role":              role,
		"request_id":        requestID,
		"await_response":    true,
		"target_agent_type": agentType,
		"subtree_info":      subtreeInfo,
		"requests_topic":    jobRequestsTopic,
		"responses_topic":   jobResponsesTopic,
		"topic_sent_to":     targetTopic,
		"result": map[string]interface{}{
			"agent_id":   agentID,
			"agent_type": agentType,
			"role":       role,
			"topics": map[string]string{
				"requests":  jobRequestsTopic,
				"responses": jobResponsesTopic,
			},
		},
	}, nil
}

// SpawnGroupAction spawns a group of agents with hierarchy tracking
func SpawnGroupAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("SpawnGroupAction starting",
		zap.Any("config", params.StepConfig))

	current, caller := getFuncInfo(1)

	params.Logger.Info("In file spawn_actions.go SpawnGroupAction",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	// Extract group configuration
	groupType, ok := params.StepConfig.Config["group_type"].(string)
	if !ok || groupType == "" {
		return nil, fmt.Errorf("group_type is required")
	}

	agents, ok := params.StepConfig.Config["agents"].(map[string]interface{})
	if !ok || len(agents) == 0 {
		return nil, fmt.Errorf("agents configuration is required")
	}

	// Generate group ID and name
	groupID := uuid.New().String()
	groupName := GenerateGroupName(groupType)

	// Pre-generate request ID for the group
	requestID := params.Headers["request_id"]
	if params.Headers["reply_to_request_id"] == "" {
		params.Headers["reply_to_request_id"] = params.Headers["request_id"]
	}
	if params.Headers["parent_responses_topic"] == "" {
		params.Headers["parent_responses_topic"] = params.Headers["responsesTopic"]
	}
	if requestID == "" {
		requestID = uuid.New().String()
		params.Headers["request_id"] = requestID
	}

	params.Logger.Info("Spawning agent group",
		zap.String("group_id", groupID),
		zap.String("group_name", groupName),
		zap.String("group_type", groupType),
		zap.Int("agent_count", len(agents)),
		zap.String("request_id", requestID),
		zap.String("reply_to_request_id", params.Headers["reply_to_request_id"]),
		zap.String("reply_to_/parents responses topic", params.Headers["parent_responses_topic"]),
	)

	// Track spawned agents
	spawnedAgents := make(map[string]string)               // role -> agentID
	spawnedSubtrees := make(map[string]*types.SubtreeInfo) // Use types.SubtreeInfo

	// Spawn each agent in the group
	for role, agentConfig := range agents {
		agentConfigMap, ok := agentConfig.(map[string]interface{})
		if !ok {
			params.Logger.Error("Invalid agent config",
				zap.String("role", role))
			continue
		}

		agentType, ok := agentConfigMap["type"].(string)
		if !ok {
			params.Logger.Error("Agent type not specified",
				zap.String("role", role))
			continue
		}

		// Create spawn config for this agent
		spawnConfig := models.Step{
			Action: "spawn_agent",
			Config: map[string]interface{}{
				"agent_type": agentType,
				"role":       role,
				"config":     agentConfigMap["config"],
				"group_id":   groupID,
				"group_name": groupName,
			},
		}

		// Use SpawnAgentAction to spawn individual agent
		spawnParams := params // Copy params
		spawnParams.StepConfig = spawnConfig

		result, err := SpawnAgentAction(ctx, spawnParams)
		if err != nil {
			params.Logger.Error("Failed to spawn agent",
				zap.String("role", role),
				zap.Error(err))
			continue
		}

		agentResult, ok := result.(map[string]interface{})
		if !ok {
			params.Logger.Error("Unexpected result type from SpawnAgentAction",
				zap.String("role", role))
			continue
		}

		agentID, ok := agentResult["agent_id"].(string)
		if !ok {
			params.Logger.Error("No agent_id in result",
				zap.String("role", role))
			continue
		}

		spawnedAgents[role] = agentID

		// Track subtree info
		if subtreeInfo, ok := agentResult["subtree_info"].(*types.SubtreeInfo); ok {
			spawnedSubtrees[agentID] = subtreeInfo
		}

		params.Logger.Info("Successfully spawned agent in group",
			zap.String("role", role),
			zap.String("agent_id", agentID),
			zap.String("agent_type", agentType),
			zap.String("group_id", groupID))
	}

	params.Logger.Info("All agents in group spawned",
		zap.String("group_id", groupID),
		zap.Int("total_spawned", len(spawnedAgents)),
		zap.Any("spawned_agents", spawnedAgents))

	// Get or create workflow for the group
	workflow := extractGroupWorkflow(params.StepConfig.Config)
	if workflow == nil {
		// Create default workflow
		workflow = createDefaultGroupWorkflow()
	}

	// Validate and marshal workflow
	workflowBytes, err := json.Marshal(workflow)
	if err != nil {
		params.Logger.Error("Failed to marshal workflow",
			zap.Error(err))
		workflowBytes = []byte(`{"start_step": "execute", "steps": {"execute": {"action": "calculate"}}}`)
	}

	// Create group subtree info using types.SubtreeInfo
	groupSubtree := &types.SubtreeInfo{
		AgentID:       groupID,
		AgentType:     "group",
		AgentName:     groupName,
		ParentAgentID: params.ExecutionContext.FromAgentID,
		Children:      spawnedSubtrees,
		CreatedAt:     time.Now(),
		LastActiveAt:  time.Now(),
		Performance: &types.PerformanceMetrics{
			TasksCompleted: 0,
			TasksFailed:    0,
			LastUpdated:    time.Now(),
		},
	}

	params.Logger.Info("SpawnGroupAction completed",
		zap.String("group_id", groupID),
		zap.String("request_id", requestID),
		zap.Int("agents_spawned", len(spawnedAgents)))

	// Return result
	return map[string]interface{}{
		"group_id":          groupID,
		"group_name":        groupName,
		"agents":            spawnedAgents,
		"workflow":          workflowBytes,
		"request_id":        requestID,
		"await_response":    true, // Wait for group initialization
		"target_agent_type": groupType,
		"subtree_info":      groupSubtree, // Return for coordinator to handle
		"spawn_count":       len(spawnedAgents),
	}, nil
}

// Helper functions

func GenerateAgentName(agentType, role string) string {
	timestamp := time.Now().Format("0102-1504")
	if role != "" && role != agentType {
		return fmt.Sprintf("%s-%s-%s", agentType, role, timestamp)
	}
	return fmt.Sprintf("%s-%s", agentType, timestamp)
}

func GenerateGroupName(groupType string) string {
	timestamp := time.Now().Format("0102-1504")
	return fmt.Sprintf("%s-group-%s", groupType, timestamp)
}

func extractGroupWorkflow(config map[string]interface{}) map[string]interface{} {
	if workflow, ok := config["workflow"].(map[string]interface{}); ok {
		return workflow
	}

	// Try to extract from JSON string
	if workflowStr, ok := config["workflow"].(string); ok {
		var workflow map[string]interface{}
		if err := json.Unmarshal([]byte(workflowStr), &workflow); err == nil {
			return workflow
		}
	}

	return nil
}

func createDefaultGroupWorkflow() map[string]interface{} {
	return map[string]interface{}{
		"start_step": "execute",
		"steps": map[string]interface{}{
			"execute": map[string]interface{}{
				"action":      "execute_llm_prompt",
				"description": "Execute group task",
				"next_step":   "complete",
			},
			"complete": map[string]interface{}{
				"action":      "complete_workflow",
				"description": "Complete group workflow",
			},
		},
	}
}

// DiscoverAgentsAction discovers available agents
func DiscoverAgentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	agentType, ok := config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_type not specified")
	}

	discover := NewAgentDiscovery(params.DB)
	if discover == nil {
		return nil, fmt.Errorf("failed to create discovery service")
	}

	// Get capabilities if specified
	var capabilities []string
	if caps, ok := config["capabilities"].([]string); ok {
		capabilities = caps
	} else if caps, ok := config["capabilities"].([]interface{}); ok {
		for _, cap := range caps {
			if c, ok := cap.(string); ok {
				capabilities = append(capabilities, c)
			}
		}
	}

	// Try recommendations first
	recommendations, err := discover.RecommendAgentsForTask(ctx, agentType, capabilities)
	if err == nil && len(recommendations) > 0 {
		params.Logger.Info("Found agent recommendations",
			zap.Int("count", len(recommendations)))
	}

	// Also get existing agents
	requirements := discovery.Requirements{
		AgentType:    agentType,
		ClientID:     params.Headers["client_id"],
		Capabilities: capabilities,
	}

	matches, err := discover.DiscoverAgents(ctx, requirements)
	if err != nil {
		return nil, err
	}

	// Get performance summary if requested
	var performanceSummary []map[string]interface{}
	if includePerf, ok := config["include_performance"].(bool); ok && includePerf {
		performanceSummary, _ = discover.GetAgentPerformanceSummary(ctx, agentType, 10)
	}

	return map[string]interface{}{
		"found_agents":     len(matches),
		"agents":           matches,
		"recommendations":  recommendations,
		"performance_data": performanceSummary,
	}, nil
}

// StartOrchestrationAction starts a child orchestration
func StartOrchestrationAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("StartOrchestrationAction starting",
		zap.Any("params into StartOrchestrationAction", params),
	)

	// Check if we're resuming from a previous attempt
	if existingChild := checkExistingChild(params); existingChild != nil {
		params.Logger.Info("Found existing child orchestration, not spawning new one")
		return existingChild, nil
	}

	// Find spawn data from previous steps
	spawnData, err := findSpawnData(params)
	if err != nil {
		return nil, fmt.Errorf("failed to find spawn data: %w", err)
	}

	// Extract workflow
	workflow, err := extractWorkflow(spawnData, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to extract workflow: %w", err)
	}

	// Generate child orchestration ID
	childOrchID := uuid.New().String()
	childOrchName := fmt.Sprintf("child-%s-%s", params.ExecutionContext.OrchestrationName, time.Now().Format("1504"))

	// Pre-generate request ID
	requestID := params.Headers["request_id"]
	if requestID == "" {
		requestID = uuid.New().String()
		params.Headers["request_id"] = requestID
	}

	parentResponsesTopicForChild := os.Getenv("RESPONSES_TOPIC")

	// Extract role from spawn data if available
	childRole := ""
	if role, ok := spawnData["role"].(string); ok {
		childRole = role
	}

	// Create start orchestration message
	startMessage := &types.RequestMessage{
		Headers: types.RequestHeaders{
			Sender: types.AgentIdentity{
				AgentType: params.ExecutionContext.FromAgentType,
				AgentID:   params.ExecutionContext.FromAgentID,
				PodName:   params.ExecutionContext.ProcessingNode,
				Role:      params.ExecutionContext.Sender.Role,
			},
			CorrelationID:           params.ExecutionContext.CorrelationID,
			CorrelationName:         params.ExecutionContext.CorrelationName,
			OrchestrationID:         childOrchID,
			OrchestrationName:       childOrchName,
			ParentOrchestrationID:   params.ExecutionContext.OrchestrationID,
			ParentOrchestrationName: params.ExecutionContext.OrchestrationName,
			FunctionalRole:          childRole,
			ReplyToRequestID:        requestID,
			StepID:                  uuid.New().String(),
			StepName:                "start_child_orchestration",
			RequestID:               requestID,
			MessageID:               uuid.New().String(),
			MessageType:             "request",
			Action:                  "start_orchestration",
			Timestamp:               time.Now(),
			FuelBudget:              params.ExecutionContext.FuelBudget - 200,
			TimeoutSeconds:          300,
		},
		Body: map[string]interface{}{
			"workflow":   workflow,
			"spawn_data": spawnData,
			"parent_info": map[string]interface{}{
				"parent_orchestration_id": params.ExecutionContext.OrchestrationID,
				"reply_to_request_id":     requestID,
				"parent_responses_topic":  parentResponsesTopicForChild,
			},
		},
	}

	// Determine target agent for orchestration
	targetAgentType := "orchestrator"
	if agentType, ok := spawnData["agent_type"].(string); ok {
		targetAgentType = agentType
	}

	// Send to target agent's request topic
	targetTopic := fmt.Sprintf("system.agent.%s.requests", targetAgentType)

	messageBytes, err := json.Marshal(startMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal start message: %w", err)
	}

	if err := params.Producer.Produce(ctx, targetTopic, startMessage.Headers.ToMap(), []byte(requestID), messageBytes); err != nil {
		return nil, fmt.Errorf("failed to send start orchestration message: %w", err)
	}

	params.Logger.Info("Child orchestration started",
		zap.String("child_orchestration_id", childOrchID),
		zap.String("child_orchestration_name", childOrchName),
		zap.String("request_id", requestID),
		zap.String("target_topic", targetTopic))

	return map[string]interface{}{
		"child_orchestration_id":   childOrchID,
		"child_orchestration_name": childOrchName,
		"request_id":               requestID,
		"await_response":           true,
		"target_agent_type":        targetAgentType,
		"workflow":                 workflow,
	}, nil
}

// Helper functions for StartOrchestrationAction

func checkExistingChild(params ActionParams) interface{} {
	// Check if we already started a child orchestration
	if params.CollectedData["start_orchestration"] != nil {
		if childMap, ok := params.CollectedData["start_orchestration"].(map[string]interface{}); ok {
			if childID, ok := childMap["child_orchestration_id"].(string); ok && childID != "" {
				params.Logger.Info("Found existing child",
					zap.String("child_orchestration_id", childID))
				return childMap
			}
		}
	}
	return nil
}

func findSpawnData(params ActionParams) (map[string]interface{}, error) {
	// Look for spawn results in collected data
	for stepName, data := range params.CollectedData {
		if dataMap, ok := data.(map[string]interface{}); ok {
			// Check for spawn result indicators
			if dataMap["workflow"] != nil || dataMap["agents"] != nil || dataMap["group_id"] != nil {
				params.Logger.Debug("Found spawn result",
					zap.String("from_step", stepName))
				return dataMap, nil
			}
		}
	}
	return nil, fmt.Errorf("no spawn result found in collected data")
}

func extractWorkflow(spawnData map[string]interface{}, logger *zap.Logger) (json.RawMessage, error) {
	workflow, ok := spawnData["workflow"]
	if !ok {
		return nil, fmt.Errorf("workflow not found in spawn result")
	}

	// Convert to JSON based on type
	switch wf := workflow.(type) {
	case json.RawMessage:
		return wf, nil
	case []byte:
		return json.RawMessage(wf), nil
	case string:
		return json.RawMessage(wf), nil
	case map[string]interface{}:
		return json.Marshal(wf)
	default:
		return nil, fmt.Errorf("unexpected workflow type: %T", workflow)
	}
}

// NewAgentDiscovery creates a discovery service from the database connection
func NewAgentDiscovery(db interface{}) *EnhancedDiscovery {
	return &EnhancedDiscovery{db: db}
}

// FindAgentsByCapability uses the SQL function to find agents
func (d *EnhancedDiscovery) FindAgentsByCapability(ctx context.Context, capabilities []string, clientID string) ([]AgentMatch, error) {
	// Convert capabilities to PostgreSQL array format
	capArray := "{" + strings.Join(capabilities, ",") + "}"

	query := `SELECT * FROM find_agents_by_capability($1::text[], $2)`

	var matches []AgentMatch

	// Handle both database types
	switch db := d.db.(type) {
	case *sql.DB:
		rows, err := db.QueryContext(ctx, query, capArray, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to find agents by capability: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var match AgentMatch
			var capabilitiesJSON []byte

			err := rows.Scan(
				&match.AgentID,
				&match.AgentType,
				&match.AgentName,
				&capabilitiesJSON,
				&match.PerformanceScore,
			)
			if err != nil {
				return nil, err
			}

			// Parse capabilities JSON
			if err := json.Unmarshal(capabilitiesJSON, &match.Capabilities); err != nil {
				return nil, err
			}

			matches = append(matches, match)
		}

		return matches, rows.Err()

	case *pgxpool.Pool:
		rows, err := db.Query(ctx, query, capArray, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to find agents by capability: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var match AgentMatch
			var capabilitiesJSON []byte

			err := rows.Scan(
				&match.AgentID,
				&match.AgentType,
				&match.AgentName,
				&capabilitiesJSON,
				&match.PerformanceScore,
			)
			if err != nil {
				return nil, err
			}

			// Parse capabilities JSON
			if err := json.Unmarshal(capabilitiesJSON, &match.Capabilities); err != nil {
				return nil, err
			}

			matches = append(matches, match)
		}

		return matches, rows.Err()

	default:
		return nil, fmt.Errorf("unsupported database type: %T", d.db)
	}
}

// RecommendAgentsForTask uses the SQL function to get agent recommendations
func (d *EnhancedDiscovery) RecommendAgentsForTask(ctx context.Context, taskType string, capabilities []string) ([]AgentRecommendation, error) {
	var capArray interface{}
	if len(capabilities) > 0 {
		capArray = "{" + strings.Join(capabilities, ",") + "}"
	} else {
		capArray = nil
	}

	query := `SELECT * FROM recommend_agents_for_task($1, $2::text[])`

	var recommendations []AgentRecommendation

	switch db := d.db.(type) {
	case *sql.DB:
		rows, err := db.QueryContext(ctx, query, taskType, capArray)
		if err != nil {
			return nil, fmt.Errorf("failed to get recommendations: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var rec AgentRecommendation
			var capabilitiesJSON []byte

			err := rows.Scan(
				&rec.AgentType,
				&rec.DisplayName,
				&rec.Category,
				&capabilitiesJSON,
				&rec.PerformanceScore,
				&rec.RecommendationReason,
			)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(capabilitiesJSON, &rec.Capabilities); err != nil {
				return nil, err
			}

			recommendations = append(recommendations, rec)
		}

		return recommendations, rows.Err()

	case *pgxpool.Pool:
		rows, err := db.Query(ctx, query, taskType, capArray)
		if err != nil {
			return nil, fmt.Errorf("failed to get recommendations: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var rec AgentRecommendation
			var capabilitiesJSON []byte

			err := rows.Scan(
				&rec.AgentType,
				&rec.DisplayName,
				&rec.Category,
				&capabilitiesJSON,
				&rec.PerformanceScore,
				&rec.RecommendationReason,
			)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(capabilitiesJSON, &rec.Capabilities); err != nil {
				return nil, err
			}

			recommendations = append(recommendations, rec)
		}

		return recommendations, rows.Err()

	default:
		return nil, fmt.Errorf("unsupported database type: %T", d.db)
	}
}

// DiscoverAgents provides backward compatibility with existing code
func (d *EnhancedDiscovery) DiscoverAgents(ctx context.Context, requirements discovery.Requirements) ([]discovery.AgentMatch, error) {
	// Use the new FindAgentsByCapability function
	matches, err := d.FindAgentsByCapability(ctx, requirements.Capabilities, requirements.ClientID)
	if err != nil {
		return nil, err
	}

	// Convert to the expected return type
	var result []discovery.AgentMatch
	for _, m := range matches {
		result = append(result, discovery.AgentMatch{
			AgentID:     m.AgentID,
			AgentType:   m.AgentType,
			Performance: m.PerformanceScore,
		})
	}

	// Filter by agent type if specified
	if requirements.AgentType != "" {
		var filtered []discovery.AgentMatch
		for _, m := range result {
			if m.AgentType == requirements.AgentType {
				filtered = append(filtered, m)
			}
		}
		result = filtered
	}

	return result, nil
}

// GetAgentPerformanceSummary gets performance metrics for agents
func (d *EnhancedDiscovery) GetAgentPerformanceSummary(ctx context.Context, agentType string, limit int) ([]map[string]interface{}, error) {
	query := `SELECT * FROM get_agent_performance_summary($1, $2)`

	var results []map[string]interface{}

	switch db := d.db.(type) {
	case *sql.DB:
		rows, err := db.QueryContext(ctx, query, agentType, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get performance summary: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var agentID string
			var agentType string
			var totalTasks int
			var successRate float64
			var avgResponseTime int
			var avgFuelPerTask int
			var avgQualityScore *float64

			err := rows.Scan(
				&agentID,
				&agentType,
				&totalTasks,
				&successRate,
				&avgResponseTime,
				&avgFuelPerTask,
				&avgQualityScore,
			)
			if err != nil {
				return nil, err
			}

			result := map[string]interface{}{
				"agent_id":          agentID,
				"agent_type":        agentType,
				"total_tasks":       totalTasks,
				"success_rate":      successRate,
				"avg_response_time": avgResponseTime,
				"avg_fuel_per_task": avgFuelPerTask,
			}

			if avgQualityScore != nil {
				result["avg_quality_score"] = *avgQualityScore
			}

			results = append(results, result)
		}

		return results, rows.Err()

	case *pgxpool.Pool:
		rows, err := db.Query(ctx, query, agentType, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get performance summary: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var agentID string
			var agentType string
			var totalTasks int
			var successRate float64
			var avgResponseTime int
			var avgFuelPerTask int
			var avgQualityScore *float64

			err := rows.Scan(
				&agentID,
				&agentType,
				&totalTasks,
				&successRate,
				&avgResponseTime,
				&avgFuelPerTask,
				&avgQualityScore,
			)
			if err != nil {
				return nil, err
			}

			result := map[string]interface{}{
				"agent_id":          agentID,
				"agent_type":        agentType,
				"total_tasks":       totalTasks,
				"success_rate":      successRate,
				"avg_response_time": avgResponseTime,
				"avg_fuel_per_task": avgFuelPerTask,
			}

			if avgQualityScore != nil {
				result["avg_quality_score"] = *avgQualityScore
			}

			results = append(results, result)
		}

		return results, rows.Err()

	default:
		return nil, fmt.Errorf("unsupported database type: %T", d.db)
	}
}

// Helper functions continued...

// getAgentDefinition retrieves agent definition from database
func getAgentDefinition(ctx context.Context, db interface{}, agentType string, logger *zap.Logger) (*AgentDefinition, error) {
	query := `
        SELECT id, type, display_name, description, category,
               image_repository, image_tag, command,
               resources, default_config, capabilities, topics,
               health_config, env_vars, is_active
        FROM agent_definitions
        WHERE type = $1 AND deleted_at IS NULL
		ORDER BY version DESC
        LIMIT 1
    `

	var def AgentDefinition
	var command sql.NullString
	var defaultConfigJSON, capabilitiesJSON, topicsJSON, healthConfigJSON, envVarsJSON []byte

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, agentType).Scan(
			&def.ID, &def.Type, &def.DisplayName, &def.Description, &def.Category,
			&def.ImageRepository, &def.ImageTag, &command,
			&def.Resources,     // This is already json.RawMessage
			&defaultConfigJSON, // Scan as bytes first
			&capabilitiesJSON,  // Scan as bytes first
			&topicsJSON,        // Already json.RawMessage
			&healthConfigJSON,  // Scan as bytes first
			&envVarsJSON,       // Already json.RawMessage
			&def.IsActive,
		)
		if err != nil {
			return nil, err
		}

		// Parse the JSON columns
		if defaultConfigJSON != nil {
			if err := json.Unmarshal(defaultConfigJSON, &def.DefaultConfig); err != nil {
				logger.Warn("Failed to unmarshal default_config", zap.Error(err))
				def.DefaultConfig = make(map[string]interface{})
			}
		}

		if capabilitiesJSON != nil {
			json.Unmarshal(capabilitiesJSON, &def.Capabilities)
		}

		def.Topics = json.RawMessage(topicsJSON)

		if healthConfigJSON != nil {
			json.Unmarshal(healthConfigJSON, &def.HealthConfig)
		}

		def.EnvVars = json.RawMessage(envVarsJSON)

		if command.Valid {
			def.Command = parsePostgresArray(command.String)
		}

	case *pgxpool.Pool:
		// pgx handles JSONB differently, it can scan directly
		err := d.QueryRow(ctx, query, agentType).Scan(
			&def.ID, &def.Type, &def.DisplayName, &def.Description, &def.Category,
			&def.ImageRepository, &def.ImageTag, &def.Command,
			&def.Resources, &def.DefaultConfig, &def.Capabilities, &def.Topics,
			&def.HealthConfig, &def.EnvVars, &def.IsActive,
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	logger.Info("Agent definition loaded for spawn",
		zap.String("agent_type", agentType),
		zap.String("image", fmt.Sprintf("%s:%s", def.ImageRepository, def.ImageTag)))

	return &def, nil
}

// createAgentInDBFromDefinition creates agent instance using definition from database
func createAgentInDBFromDefinition(ctx context.Context, params ActionParams, agentID string, agentDef *AgentDefinition, clientID string) error {
	defaultConfig := agentDef.DefaultConfig
	if defaultConfig == nil {
		defaultConfig = make(map[string]interface{})
		params.Logger.Warn("No default config, using empty config",
			zap.String("agent_type", agentDef.Type))
	}

	// Ensure we have a workflow
	if _, hasWorkflow := defaultConfig["workflow"]; !hasWorkflow {
		params.Logger.Info("No workflow in default config, using minimal workflow",
			zap.String("agent_type", agentDef.Type))
		defaultConfig["workflow"] = buildMinimalWorkflow(agentDef.Type)
	}

	// Parse and set capabilities
	var capabilities []string
	if err := json.Unmarshal(agentDef.Capabilities, &capabilities); err == nil {
		defaultConfig["capabilities"] = capabilities
	} else {
		defaultConfig["capabilities"] = []string{agentDef.Type}
	}

	// Parse topics configuration
	topics := parseTopicConfig(agentDef.Topics)
	processTopic := strings.ReplaceAll(topics.Process, "{type}", agentDef.Type)

	// Add runtime configuration
	runtimeConfig := map[string]interface{}{
		"agent_id":     agentID,
		"agent_type":   agentDef.Type,
		"display_name": agentDef.DisplayName,
		"category":     agentDef.Category,
		"topic":        processTopic,
		"topics": map[string]string{
			"process":  strings.ReplaceAll(topics.Process, "{type}", agentDef.Type),
			"response": strings.ReplaceAll(topics.Response, "{type}", agentDef.Type),
			"error":    strings.ReplaceAll(topics.Error, "{type}", agentDef.Type),
			"dlq":      strings.ReplaceAll(topics.DLQ, "{type}", agentDef.Type),
		},
	}

	// Merge runtime config with default config
	for k, v := range runtimeConfig {
		defaultConfig[k] = v
	}

	// Add any overrides from the spawn request
	if overrides, ok := params.StepConfig.Config["config_overrides"].(map[string]interface{}); ok {
		for k, v := range overrides {
			defaultConfig[k] = v
		}
		params.Logger.Info("Applied config overrides",
			zap.String("agent_id", agentID),
			zap.Int("override_count", len(overrides)))
	}

	// Marshal the final configuration
	configJSON, err := json.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal agent config: %w", err)
	}

	// Prepare the insert query for the client-specific schema
	insertQuery := fmt.Sprintf(`
		INSERT INTO client_%s.agent_instances 
		(id, template_id, owner_user_id, name, config, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			config = EXCLUDED.config,
			updated_at = NOW()
	`, clientID)

	// Get user ID from headers or default to system
	userID := params.Headers["user_id"]
	if userID == "" {
		userID = "system"
	}

	// Generate a descriptive name for the instance
	instanceName := fmt.Sprintf("%s-%s", agentDef.DisplayName, time.Now().Format("20060102-150405"))
	if customName, ok := params.StepConfig.Config["instance_name"].(string); ok && customName != "" {
		instanceName = customName
	}

	// Execute the insert
	_, err = params.DB.ExecContext(ctx, insertQuery,
		agentID,
		agentDef.ID, // Reference to the agent_definitions table
		userID,
		instanceName,
		configJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to insert agent instance: %w", err)
	}

	params.Logger.Info("Agent instance created in database",
		zap.String("agent_id", agentID),
		zap.String("agent_definition id", agentDef.ID),
		zap.String("agent_id", agentID),
		zap.String("agent_type", agentDef.Type),
		zap.String("instance_name", instanceName),
		zap.String("client_id", clientID))

	return nil
}

// spawnAgentKubernetesJobFromDefinition spawns job using database definition
func spawnAgentKubernetesJobFromDefinition(ctx context.Context, agentID string, agentDef *AgentDefinition, clientID, jobRequestsTopic, jobResponsesTopic, parentResponsesTopic string, logger *zap.Logger) (string, error) {
	logger.Info("In spawnAgentKubernetesJobFromDefinition")

	current, caller := getFuncInfo(1)

	logger.Info("In file spawn_actions.go spawnAgentKubernetesJobFromDefinition",
		zap.String("function", current),
		zap.String("called_by", caller),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	// Get in-cluster config
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Generate unique job name - use only first 8 chars for K8s naming constraints
	jobName := fmt.Sprintf("agent-%s-%s", agentDef.Type, agentID[:8])

	// Check if job already exists and handle accordingly
	existingJob, err := clientset.BatchV1().Jobs("ai-persona-system").Get(ctx, jobName, metav1.GetOptions{})
	if err == nil {
		if existingJob.Status.Failed > 0 || existingJob.Status.Succeeded > 0 {
			logger.Info("Deleting completed/failed job before recreating",
				zap.String("job_name", jobName))

			deletePolicy := metav1.DeletePropagationForeground
			err = clientset.BatchV1().Jobs("ai-persona-system").Delete(ctx, jobName, metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			})
			if err != nil {
				logger.Warn("Failed to delete old job", zap.Error(err))
			}
			time.Sleep(2 * time.Second)
		} else if existingJob.Status.Active > 0 {
			logger.Info("Job is already running",
				zap.String("job_name", jobName))
			return jobName, nil
		}
	}

	// Parse configurations
	resources := parseResourceSpec(agentDef.Resources)
	healthConfig := parseHealthConfig(agentDef.HealthConfig)
	envVars := parseEnvVars(agentDef.EnvVars)

	// Parse topics from json.RawMessage
	// topics := parseTopicConfig(agentDef.Topics)

	// Build topic name
	//processTopic := strings.ReplaceAll(topics.Process, "{type}", agentDef.Type)

	// Build environment variables
	envList := []corev1.EnvVar{
		// Core configuration
		{Name: "AGENT_TYPE", Value: agentDef.Type},
		{Name: "AGENT_ID", Value: agentID},
		{Name: "CLIENT_ID", Value: clientID},

		// Dynamic topic configuration
		{Name: "KAFKA_TOPIC", Value: jobRequestsTopic},  // legacy?
		{Name: "KAFKA_TOPICS", Value: jobRequestsTopic}, // what to listen to
		{Name: "RESPONSES_TOPIC", Value: jobResponsesTopic},
		{Name: "REQUESTS_TOPIC", Value: jobRequestsTopic},
		{Name: "PARENT_RESPONSES_TOPIC", Value: parentResponsesTopic}, // this should set env var in agent container
		{Name: "KAFKA_CONSUMER_GROUP", Value: fmt.Sprintf("%s-group-%s", agentDef.Type, agentID[:8])},

		// Health server ports
		{Name: "HEALTH_PORT", Value: fmt.Sprintf("%d", healthConfig.Port)},
		{Name: "METRICS_PORT", Value: "9090"},

		// Core Manager URL
		{Name: "CORE_MANAGER_URL", Value: "http://core-manager.ai-persona-system.svc.cluster.local:8088"},
	}

	// Add custom env vars from definition
	for _, envVar := range envVars {
		envList = append(envList, corev1.EnvVar{
			Name:  envVar.Name,
			Value: envVar.Value,
		})
	}

	// Add infrastructure configuration from orchestrator environment
	envList = append(envList, []corev1.EnvVar{
		{Name: "SERVICE_INFRASTRUCTURE_KAFKA_BROKERS", Value: os.Getenv("KAFKA_BROKERS")},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_HOST", Value: os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_HOST")},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_PORT", Value: os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_PORT")},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_USER", Value: os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_USER")},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_DB_NAME", Value: os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_DB_NAME")},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_HOST", Value: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_HOST")},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_PORT", Value: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_PORT")},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_USER", Value: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_USER")},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_DB_NAME", Value: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_DB_NAME")},
	}...)

	// Add database passwords and other secrets
	envList = append(envList, []corev1.EnvVar{
		{
			Name: "CLIENTS_DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "personae-platform-secrets",
					},
					Key: "CLIENTS_DB_PASSWORD",
				},
			},
		},
		{
			Name: "PGPASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "personae-platform-secrets",
					},
					Key: "CLIENTS_DB_PASSWORD",
				},
			},
		},
		{
			Name: "TEMPLATES_DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "personae-platform-secrets",
					},
					Key: "TEMPLATES_DB_PASSWORD",
				},
			},
		},
		{
			Name: "AUTH_DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "personae-platform-secrets",
					},
					Key: "AUTH_DB_PASSWORD",
				},
			},
		},
		{
			Name: "ANTHROPIC_API_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "personae-default-secrets",
					},
					Key: "ANTHROPIC_API_KEY",
				},
			},
		},
		{
			Name: "AGENT_BOOTSTRAP_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "personae-platform-secrets",
					},
					Key: "agent-bootstrap-key",
				},
			},
		},
	}...)

	// Add storage configuration for storage-enabled agents
	if isStorageEnabledAgent(agentDef.Type) || agentDef.Category == "orchestrator" || agentDef.Category == "code-driven" {
		// Get storage credentials from orchestrator's environment
		awsKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
		awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		b2KeyId := os.Getenv("B2_APPLICATION_KEY_ID")
		b2Key := os.Getenv("B2_APPLICATION_KEY")

		logger.Info("Injecting storage credentials",
			zap.String("agent_type", agentDef.Type),
			zap.Bool("has_aws", awsKeyId != ""),
			zap.Bool("has_b2", b2KeyId != ""))

		// Add storage credentials as direct values
		if awsKeyId != "" {
			envList = append(envList, corev1.EnvVar{Name: "AWS_ACCESS_KEY_ID", Value: awsKeyId})
		}
		if awsSecretKey != "" {
			envList = append(envList, corev1.EnvVar{Name: "AWS_SECRET_ACCESS_KEY", Value: awsSecretKey})
		}
		if b2KeyId != "" {
			envList = append(envList, corev1.EnvVar{Name: "B2_APPLICATION_KEY_ID", Value: b2KeyId})
		}
		if b2Key != "" {
			envList = append(envList, corev1.EnvVar{Name: "B2_APPLICATION_KEY", Value: b2Key})
		}

		// Storage configuration from ConfigMap
		storageConfigMap := os.Getenv("AGENT_STORAGE_CONFIGMAP")
		if storageConfigMap == "" {
			storageConfigMap = "storage-config"
		}

		envList = append(envList, []corev1.EnvVar{
			{
				Name: "S3_ENDPOINT",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: storageConfigMap},
						Key:                  "S3-ENDPOINT",
					},
				},
			},
			{
				Name: "S3_REGION",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: storageConfigMap},
						Key:                  "S3-REGION",
					},
				},
			},
			{
				Name: "IMAGE_BUCKET",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: storageConfigMap},
						Key:                  "image_bucket",
					},
				},
			},
			{
				Name: "ASSETS_BUCKET",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: storageConfigMap},
						Key:                  "assets_bucket",
					},
				},
			},
			{
				Name: "S3_USE_PATH_STYLE",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: storageConfigMap},
						Key:                  "S3_USE_PATH_STYLE",
					},
				},
			},
		}...)
	}

	// Read the config from the mounted ConfigMap
	var dbHost, dbPort, dbUser, dbName string

	// Try to read from the mounted ConfigMap first
	configPath := "/app/configs/agent-chassis.yaml"
	if configData, err := ioutil.ReadFile(configPath); err == nil {
		var configMap map[string]interface{}
		if err := yaml.Unmarshal(configData, &configMap); err == nil {
			if infra, ok := configMap["infrastructure"].(map[interface{}]interface{}); ok {
				if clientsDB, ok := infra["clients_database"].(map[interface{}]interface{}); ok {
					dbHost, _ = clientsDB["host"].(string)
					if port, ok := clientsDB["port"].(int); ok {
						dbPort = fmt.Sprintf("%d", port)
					}
					dbUser, _ = clientsDB["user"].(string)
					dbName, _ = clientsDB["db_name"].(string)

					logger.Info("Loaded DB config from config file",
						zap.String("host", dbHost),
						zap.String("port", dbPort),
						zap.String("user", dbUser),
						zap.String("dbname", dbName))
				}
			}
		}
	} else {
		logger.Warn("Could not read config file, using fallback values",
			zap.String("path", configPath),
			zap.Error(err))
	}

	// Fallback to environment variables if ConfigMap read failed
	if dbHost == "" {
		dbHost = os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_HOST")
		if dbHost == "" {
			// Final fallback to hardcoded values
			dbHost = "postgres-clients.ai-persona-system.svc.cluster.local"
		}
	}
	if dbPort == "" {
		dbPort = os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_PORT")
		if dbPort == "" {
			dbPort = "5432"
		}
	}
	if dbUser == "" {
		dbUser = os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_USER")
		if dbUser == "" {
			dbUser = "clients_user"
		}
	}
	if dbName == "" {
		dbName = os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_DB_NAME")
		if dbName == "" {
			dbName = "clients_db"
		}
	}

	// Build DATABASE_URL with the values from ConfigMap or fallbacks
	dbURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbName)

	logger.Info("Constructed DATABASE_URL for spawned agent",
		zap.String("dbURL", dbURL),
		zap.String("agent_type", agentDef.Type))

	envList = append(envList, corev1.EnvVar{Name: "DATABASE_URL", Value: dbURL})

	logger.Info("DEBUGaa: In spawnAgentKubernetesJobFromDefinition in spawn_actions.go",
		zap.String("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_USER", os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_USER")),
		zap.String("CLIENTS_DB_PASSWORD", os.Getenv("CLIENTS_DB_PASSWORD")),
		zap.String("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_HOST", os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_HOST")),
		zap.String("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_PORT", os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_PORT")),
		zap.String("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_DB_NAME", os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_DB_NAME")),
		zap.String("CLIENTS_DATABASE_USER", os.Getenv("CLIENTS_DATABASE_USER")),
		zap.String("CLIENTS_DATABASE_HOST", os.Getenv("CLIENTS_DATABASE_HOST")),
		zap.String("CLIENTS_DATABASE_PORT", os.Getenv("CLIENTS_DATABASE_PORT")),
		zap.String("CLIENTS_DATABASE_DB_NAME", os.Getenv("CLIENTS_DATABASE_DB_NAME")),
	)

	// Define the Job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "ai-persona-system",
			Labels: map[string]string{
				"app":        "dynamic-agent",
				"agent-type": agentDef.Type,
				"agent-id":   agentID,
				"client-id":  clientID,
				"spawned-by": "orchestrator",
				"component":  "agent",
				"category":   agentDef.Category,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: int32Ptr(3600),
			BackoffLimit:            int32Ptr(3),
			ActiveDeadlineSeconds:   int64Ptr(86400),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        "dynamic-agent",
						"agent-type": agentDef.Type,
						"agent-id":   agentID,
						"client-id":  clientID,
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9090",
						"prometheus.io/path":   "/metrics",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: "ai-persona-app",
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: "docker-hub-creds"},
					},
					Containers: []corev1.Container{
						{
							Name:    "agent",
							Image:   fmt.Sprintf("%s:%s", agentDef.ImageRepository, agentDef.ImageTag),
							Command: agentDef.Command,
							Ports: []corev1.ContainerPort{
								{
									Name:          "health",
									ContainerPort: int32(healthConfig.Port),
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "metrics",
									ContainerPort: 9090,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: envList,
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "personae-prod-config",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(resources.Requests["cpu"]),
									corev1.ResourceMemory: resource.MustParse(resources.Requests["memory"]),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(resources.Limits["cpu"]),
									corev1.ResourceMemory: resource.MustParse(resources.Limits["memory"]),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   healthConfig.LivenessPath,
										Port:   intstr.FromString("health"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: int32(healthConfig.InitialDelaySeconds),
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   healthConfig.ReadinessPath,
										Port:   intstr.FromString("health"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       5,
								TimeoutSeconds:      3,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
						},
					},
				},
			},
		},
	}

	// Create the Job
	createdJob, err := clientset.BatchV1().Jobs("ai-persona-system").Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			logger.Warn("Job already exists (race condition), continuing",
				zap.String("job_name", jobName))
			return jobName, nil
		}
		return "", fmt.Errorf("failed to create kubernetes job: %w", err)
	}

	logger.Info("Created Kubernetes Job",
		zap.String("job_name", createdJob.Name),
		zap.String("namespace", createdJob.Namespace),
		zap.String("uid", string(createdJob.UID)))

	return createdJob.Name, nil
}

// checkExistingAgent checks if an agent already exists in the database
func checkExistingAgent(ctx context.Context, db interface{}, clientID, agentType string) *AgentInfo {
	var id string
	query := fmt.Sprintf(`
		SELECT id FROM client_%s.agent_instances 
		WHERE config->>'agent_type' = $1 
		AND is_active = true
		LIMIT 1
	`, clientID)

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, agentType).Scan(&id)
		if err != nil {
			return nil
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, agentType).Scan(&id)
		if err != nil {
			return nil
		}
	default:
		return nil
	}

	return &AgentInfo{ID: id}
}

// isAgentJobRunning checks if an agent job is currently running
func isAgentJobRunning(ctx context.Context, agentID string, logger *zap.Logger) bool {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Warn("Failed to get k8s config", zap.Error(err))
		return false
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		logger.Warn("Failed to create k8s client", zap.Error(err))
		return false
	}

	// Check for jobs with this agent ID
	jobs, err := clientset.BatchV1().Jobs("ai-persona-system").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("agent-id=%s", agentID),
	})

	if err != nil || len(jobs.Items) == 0 {
		return false
	}

	// Check if any job is active
	for _, job := range jobs.Items {
		if job.Status.Active > 0 {
			logger.Info("Found active job for agent",
				zap.String("job_name", job.Name),
				zap.Int32("active_pods", job.Status.Active))
			return true
		}
	}

	return false
}

// parseTopicConfig parses the topic configuration from JSON
func parseTopicConfig(topicsJSON json.RawMessage) TopicConfig {
	var topics TopicConfig
	if err := json.Unmarshal(topicsJSON, &topics); err != nil {
		// Return defaults
		return TopicConfig{
			Process:  "system.agent.{type}.requests",
			Response: "system.agent.{type}.responses",
			Error:    "system.agent.{type}.errors",
			DLQ:      "system.agent.{type}.dlq",
		}
	}
	return topics
}

// parseResourceSpec parses resource configuration
func parseResourceSpec(resourcesJSON json.RawMessage) ResourceSpec {
	var spec ResourceSpec
	if err := json.Unmarshal(resourcesJSON, &spec); err != nil {
		// Return defaults
		return ResourceSpec{
			Requests: map[string]string{"cpu": "100m", "memory": "256Mi"},
			Limits:   map[string]string{"cpu": "500m", "memory": "1Gi"},
		}
	}
	return spec
}

// parseHealthConfig parses health check configuration
func parseHealthConfig(healthJSON json.RawMessage) HealthCheckConfig {
	var config HealthCheckConfig
	if err := json.Unmarshal(healthJSON, &config); err != nil {
		// Return defaults
		return HealthCheckConfig{
			LivenessPath:        "/health",
			ReadinessPath:       "/ready",
			Port:                8080,
			InitialDelaySeconds: 30,
		}
	}
	return config
}

// parseEnvVars parses environment variables configuration
func parseEnvVars(envJSON json.RawMessage) []EnvVar {
	var envVars []EnvVar
	json.Unmarshal(envJSON, &envVars)
	return envVars
}

// parsePostgresArray parses PostgreSQL array format
func parsePostgresArray(s string) []string {
	// PostgreSQL arrays look like: {value1,value2,value3}
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

// buildMinimalWorkflow creates a minimal fallback workflow
func buildMinimalWorkflow(agentType string) map[string]interface{} {
	// Determine the primary action based on agent type
	primaryAction := "execute_llm_prompt"

	switch {
	case strings.Contains(agentType, "generic"):
		primaryAction = "validate_input"
	case strings.Contains(agentType, "adapt"):
		primaryAction = "http_request"
	case strings.Contains(agentType, "publish"):
		primaryAction = "deploy_to_hosting"
	default:
		primaryAction = "execute_llm_prompt"
	}

	return map[string]interface{}{
		"start_step": "process",
		"steps": map[string]interface{}{
			"process": map[string]interface{}{
				"action":      primaryAction,
				"description": fmt.Sprintf("Process %s task", agentType),
				"next_step":   "complete",
			},
			"complete": map[string]interface{}{
				"action":      "complete_workflow",
				"description": "Complete the workflow",
			},
		},
	}
}

// isStorageEnabledAgent checks if an agent type requires storage access
func isStorageEnabledAgent(agentType string) bool {
	storageAgents := []string{
		"site-publisher",
		"html-developer",
		"visual-designer",
		"image-generator",
		"content-creator",
		"content-researcher",
		"domain-analyst",
		"site-architect",
		"website-builder",
	}

	for _, t := range storageAgents {
		if t == agentType {
			return true
		}
	}
	return false
}

// Helper functions for runtime
func getFuncInfo(skip int) (current, caller string) {
	if pc, _, _, ok := runtime.Caller(skip); ok {
		current = runtime.FuncForPC(pc).Name()
	}
	if pc, _, _, ok := runtime.Caller(skip + 1); ok {
		caller = runtime.FuncForPC(pc).Name()
	}
	return
}

// Pointer helper functions
func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

// parseHealthConfigFromMap parses health check configuration from a map
func parseHealthConfigFromMap(configMap map[string]interface{}) HealthCheckConfig {
	config := HealthCheckConfig{
		LivenessPath:        "/health",
		ReadinessPath:       "/ready",
		Port:                8080,
		InitialDelaySeconds: 30,
	}

	if configMap == nil {
		return config
	}

	if path, ok := configMap["liveness_path"].(string); ok {
		config.LivenessPath = path
	}
	if path, ok := configMap["readiness_path"].(string); ok {
		config.ReadinessPath = path
	}
	if port, ok := configMap["port"].(float64); ok {
		config.Port = int(port)
	}
	if delay, ok := configMap["initial_delay_seconds"].(float64); ok {
		config.InitialDelaySeconds = int(delay)
	}

	return config
}

// getNestedValue safely traverses a map[string]interface{} to find a value.
func getNestedInputDataValue(data map[string]interface{}, path ...string) (interface{}, bool) {
	var current interface{} = data
	for _, key := range path {
		m, ok := current.(map[string]interface{})
		if !ok {
			// This level is not a map, so we can't go deeper.
			return nil, false
		}
		current, ok = m[key]
		if !ok {
			// The key doesn't exist at this level.
			return nil, false
		}
	}
	return current, true
}
