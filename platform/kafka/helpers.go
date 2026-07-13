package kafka

import (
	"fmt"
	"os"
	"strings"
)

// GetBrokers returns the Kafka broker addresses from environment variables.
// Checks SERVICE_INFRASTRUCTURE_KAFKA_BROKERS first, falls back to KAFKA_BROKERS.
func GetBrokers() []string {
	brokersEnv := os.Getenv("SERVICE_INFRASTRUCTURE_KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = os.Getenv("KAFKA_BROKERS")
	}
	if brokersEnv == "" {
		return nil
	}
	return strings.Split(brokersEnv, ",")
}

func CreateStableIdentity(correlationID, orchestrationID, agentType, stepName string) string {
	if correlationID == "" {
		correlationID = "unknowncorrelationid"
	}
	if orchestrationID == "" {
		orchestrationID = "unknownorchestrationid"
	}
	if agentType == "" {
		agentType = "unknownagent"
	}
	if stepName == "" {
		stepName = "unknownstep"
	}
	return fmt.Sprintf("%s-%s-%s-%s",
		correlationID[:8],
		orchestrationID[:8],
		agentType,
		stepName)
}
