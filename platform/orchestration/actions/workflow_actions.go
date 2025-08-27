// platform/orchestration/actions/workflow_actions.go

package actions

import (
	"context"
	"strings"
	"time"
)

func CompleteWorkflowAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Completing workflow")

	// Clean collected data - remove internal tracking keys
	result := make(map[string]interface{})
	for k, v := range params.CollectedData {
		// Skip internal keys and avoid duplicating correlation IDs
		if k == params.CurrentStep ||
			k == "complete" ||
			strings.HasSuffix(k, "_started") ||
			k == params.Headers["correlation_id"] ||
			k == params.Headers["parent_correlation_id"] {
			continue
		}
		result[k] = v
	}

	return map[string]interface{}{
		"status":    "completed",
		"result":    result,
		"timestamp": time.Now().UTC(),
	}, nil
}
