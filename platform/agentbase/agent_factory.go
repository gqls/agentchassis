// FILE: platform/agentbase/agent_factory.go
package agentbase

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// NewAgentFromSpawnRequest creates an agent from a spawn request
func NewAgentFromSpawnRequest(ctx context.Context, spawnRequest *types.RequestMessage) (*Agent, error) {
	// Extract configuration from spawn request
	body, ok := spawnRequest.Body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spawn request body")
	}

	// Extract fields with safe type assertions
	agentID := getStringFromMap(body, "agent_id")
	if agentID == "" {
		agentID = uuid.New().String()
	}

	agentName := getStringFromMap(body, "agent_name")
	if agentName == "" {
		agentName = fmt.Sprintf("%s-%s", spawnRequest.Headers.ToAgentType, time.Now().Format("0102-1504"))
	}

	agentType := spawnRequest.Headers.ToAgentType
	role := getStringFromMap(body, "role")

	// Extract parent info
	parentInfo, _ := body["parent"].(map[string]interface{})
	parentAgentID := getStringFromMap(parentInfo, "agent_id")
	parentAgentType := getStringFromMap(parentInfo, "agent_type")

	// Extract dynamic config
	dynamicConfig, _ := body["config"].(map[string]interface{})

	// Get Kafka brokers from environment
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}
	if kafkaBrokers[0] == "" {
		kafkaBrokers = []string{"localhost:9092"}
	}

	// Create agent config
	config := AgentConfig{
		AgentType:             agentType,
		AgentName:             agentName,
		AgentVersion:          os.Getenv("AGENT_VERSION"),
		Role:                  role,
		ParentAgentID:         parentAgentID,
		ParentAgentType:       parentAgentType,
		ParentOrchestrationID: spawnRequest.Headers.ParentOrchestrationID,
		EnableStateless:       true,
		KafkaBrokers:          kafkaBrokers,
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		DynamicConfig:         dynamicConfig,
	}

	// Create the agent
	agent, err := NewAgent(config)
	if err != nil {
		return nil, err
	}

	// Mark as spawned with spawn time
	agent.spawned = true
	agent.spawnTime = time.Now()

	// Send initialization response
	if err := agent.SendInitializationResponse(spawnRequest); err != nil {
		agent.logger.Error("Failed to send initialization response", zap.Error(err))
	}

	// Register with parent if applicable
	if agent.ParentOrchestrationID != "" {
		agent.registerWithParent(spawnRequest)
	}

	return agent, nil
}

// Helper function to safely extract strings from maps
func getStringFromMap(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
