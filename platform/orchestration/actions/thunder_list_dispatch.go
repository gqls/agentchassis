// FILE: platform/orchestration/actions/thunder_list_dispatch.go
//
// Dispatches a list_instances request to thunder-adapter and awaits the
// response. Used by the thunder-orphan-scan agent as step 1 of the orphan
// sweep: the adapter answers with Thunder's own view of the account
// (every instance the authenticated token can see), which the
// reconcile_thunder_instances step then compares against thunder_instances.
//
// Mirrors thunder_decommission_dispatch.go — same envelope, same topic,
// same AwaitResponse contract. The request carries no parameters: the
// adapter-side action is a read-only passthrough to GET /instances/list.
//
// The awaited response body lands in CollectedData under this step's
// output_field as {"response": {"success": true, "count": N,
// "instances": [...]}} — instances in Thunder's own field shapes
// (camelCase, id as a JSON string, UPPERCASE statuses).

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

// thunderListAction is the adapter-side action handled by
// internal/adapters/thunder/adapter.go handleListInstances.
const thunderListAction = "list_instances"

// DispatchThunderListAction publishes a list_instances request to
// thunder-adapter. Returns immediately with AwaitResponse: true; the
// chassis state machine waits for the response before advancing.
func DispatchThunderListAction(ctx context.Context, params ActionParams) (interface{}, error) {
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"initialized": true}, nil
	}

	params.Logger.Info("DispatchThunderListAction starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "system"
	}

	// Pick the responses_topic this agent listens on — same logic as the
	// other adapter-dispatch actions.
	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		agentType := params.ExecutionContext.Sender.AgentType
		if agentType == "" {
			agentType = os.Getenv("AGENT_TYPE")
		}
		if agentType != "" {
			myResponsesTopic = fmt.Sprintf("system.agent.%s.responses", agentType)
		} else {
			myResponsesTopic = "system.agent.generic.responses"
		}
	}

	newRequestID := uuid.NewString()

	requestBody := map[string]interface{}{
		"action":         thunderListAction,
		"reply_to_topic": myResponsesTopic,
	}

	requestHeaders := map[string]string{
		"correlation_id":          params.ExecutionContext.CorrelationID,
		"orchestration_id":        params.ExecutionContext.OrchestrationID,
		"orchestration_name":      params.ExecutionContext.OrchestrationName,
		"parent_orchestration_id": params.ExecutionContext.ParentOrchestrationID,
		"client_id":               clientID,
		"step_name":               params.ExecutionContext.StepName,
		"step_id":                 params.ExecutionContext.StepID,
		"request_id":              newRequestID,
		"message_type":            "request",
		"action":                  thunderListAction,

		"sender_agent_type":    defaultIfEmpty(params.ExecutionContext.Sender.AgentType, os.Getenv("AGENT_TYPE")),
		"sender_agent_id":      params.ExecutionContext.OrchestrationID,
		"sender_pod_name":      defaultIfEmpty(params.ExecutionContext.Sender.PodName, os.Getenv("POD_NAME")),
		"sender_agent_version": params.ExecutionContext.Sender.AgentVersion,

		"responses_topic":        myResponsesTopic,
		"reply_to_topic":         myResponsesTopic,
		"parent_responses_topic": myResponsesTopic,

		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	envelope := map[string]interface{}{
		"headers": requestHeaders,
		"body":    requestBody,
	}

	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("dispatch_thunder_list: marshal envelope: %w", err)
	}

	key := []byte(params.ExecutionContext.CorrelationID)

	if err := params.Producer.ProduceWithValidation(
		ctx, thunderAdapterTopic, requestHeaders, key, envelopeBytes,
	); err != nil {
		params.Logger.Error("Failed to publish list_instances to thunder-adapter",
			zap.String("topic", thunderAdapterTopic),
			zap.Error(err))
		return nil, fmt.Errorf("dispatch_thunder_list: produce: %w", err)
	}

	params.Logger.Info("Dispatched list_instances to thunder-adapter",
		zap.String("topic", thunderAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("await_responses_topic", myResponsesTopic),
	)

	return map[string]interface{}{
		"request_id":        newRequestID,
		"topic_sent_to":     thunderAdapterTopic,
		"requests_topic":    thunderAdapterTopic,
		"responses_topic":   myResponsesTopic,
		"await_response":    true,
		"target_agent_type": "thunder-adapter",
		"action_sent":       thunderListAction,
	}, nil
}
