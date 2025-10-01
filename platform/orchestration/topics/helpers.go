package topics

import "fmt"

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
