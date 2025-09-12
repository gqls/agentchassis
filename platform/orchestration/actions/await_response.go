// In platform/orchestration/actions/await_response.go (new file)
package actions

import (
	"context"

	"go.uber.org/zap"
)

// AwaitResponseAction is used when a workflow needs to wait for a response
// This is typically used after spawning an agent or calling another service
func AwaitResponseAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Awaiting response",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID))

	// This action doesn't actually do anything - it just signals
	// that the orchestration should wait for a response
	// The actual waiting is handled by the orchestrator based on
	// the await_response flag from the previous action

	// Check if we have a request_id from a previous step
	var requestID string
	for _, data := range params.CollectedData {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if rid, ok := dataMap["request_id"].(string); ok && rid != "" {
				requestID = rid
				break
			}
		}
	}

	return map[string]interface{}{
		"status":         "awaiting",
		"request_id":     requestID,
		"await_response": false, // This action itself doesn't wait
	}, nil
}
