// FILE: platform/orchestration/actions/thunder_decommission_dispatch.go
//
// Dispatches a decommission_instance request to thunder-adapter. Used by
// the thunder-reaper agent (Phase 3.5) and by any other workflow that
// wants to terminate a Thunder Compute instance.
//
// Mirrors the existing dedicated-action pattern used for adapter dispatch
// (git_commit, web_search, web_scrape — each is a hand-rolled action that
// builds the chassis envelope and produces directly to the adapter's
// topic). Returns AwaitResponse: true so the chassis state machine waits
// for thunder-adapter's success/error response before advancing.
//
// Input (from CollectedData["input_data"]):
//   - provisioning_id: DB row UUID of the thunder_instances row (preferred)
//   - thunder_identifier: Thunder API numeric ID as string (fallback)
//   - reason: optional audit string, e.g. "reaper:max_uptime_exceeded"
//
// At least one of provisioning_id or thunder_identifier must be present.

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

const (
	// thunderAdapterTopic is the fixed adapter topic — matches what's
	// configured in configs/thunder-adapter.yaml (custom.adapter_settings.main_topic).
	thunderAdapterTopic = "system.adapter.thunder.requests"

	// thunderDispatchAction is the adapter-side action handled by
	// internal/adapters/thunder/adapter.go handleDecommissionInstance.
	thunderDispatchAction = "decommission_instance"
)

// DispatchThunderDecommissionAction publishes a decommission_instance request
// to thunder-adapter. Returns immediately with AwaitResponse: true; the
// chassis state machine waits for the response before advancing the workflow.
func DispatchThunderDecommissionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Standard early-return for "initialize" calls (the chassis sends an
	// initialize action when starting a workflow; deterministic actions
	// should no-op).
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"initialized": true}, nil
	}

	params.Logger.Info("DispatchThunderDecommissionAction starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// ── Extract input fields ──
	provisioningID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.provisioning_id")
	thunderIdentifier := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.thunder_identifier")
	reason := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.reason")

	if provisioningID == "" && thunderIdentifier == "" {
		return nil, fmt.Errorf("dispatch_thunder_decommission: must supply input_data.provisioning_id or input_data.thunder_identifier")
	}

	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "system"
	}

	// ── Pick the responses_topic this agent listens on ──
	// Same logic as other adapter-dispatch actions: prefer the agent's
	// own configured responses_topic, fall back to the standard generic one.
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

	// ── Build the request body that thunder-adapter expects ──
	// adapter.go handleDecommissionInstance re-marshals body into
	// DecommissionInstanceRequest. Fields match decommission_action.go.
	requestBody := map[string]interface{}{
		"action":         thunderDispatchAction,
		"reply_to_topic": myResponsesTopic,
	}
	if provisioningID != "" {
		requestBody["provisioning_id"] = provisioningID
	}
	if thunderIdentifier != "" {
		requestBody["thunder_identifier"] = thunderIdentifier
	}
	if reason != "" {
		requestBody["reason"] = reason
	}

	// ── Build the chassis envelope (headers + body) ──
	// Headers carry Tier 1 + Tier 2 correlation fields so the adapter's
	// validator and the chassis-side response router both work.
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
		"action":                  thunderDispatchAction,

		// Sender identity — required by the adapter's validator.
		"sender_agent_type":    defaultIfEmpty(params.ExecutionContext.Sender.AgentType, os.Getenv("AGENT_TYPE")),
		"sender_agent_id":      params.ExecutionContext.OrchestrationID,
		"sender_pod_name":      defaultIfEmpty(params.ExecutionContext.Sender.PodName, os.Getenv("POD_NAME")),
		"sender_agent_version": params.ExecutionContext.Sender.AgentVersion,

		// Response routing — adapter sends success/error responses here.
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
		return nil, fmt.Errorf("dispatch_thunder_decommission: marshal envelope: %w", err)
	}

	// ── Publish to thunder-adapter's requests topic ──
	key := []byte(params.ExecutionContext.CorrelationID)

	if err := params.Producer.ProduceWithValidation(
		ctx, thunderAdapterTopic, requestHeaders, key, envelopeBytes,
	); err != nil {
		params.Logger.Error("Failed to publish decommission to thunder-adapter",
			zap.String("topic", thunderAdapterTopic),
			zap.String("provisioning_id", provisioningID),
			zap.String("thunder_identifier", thunderIdentifier),
			zap.Error(err))
		return nil, fmt.Errorf("dispatch_thunder_decommission: produce: %w", err)
	}

	params.Logger.Info("Dispatched decommission_instance to thunder-adapter",
		zap.String("topic", thunderAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("provisioning_id", provisioningID),
		zap.String("thunder_identifier", thunderIdentifier),
		zap.String("reason", reason),
		zap.String("await_responses_topic", myResponsesTopic),
	)

	// ── Return AwaitResponse so the chassis waits for the response ──
	// The same field shape as call_agent / git_commit etc. — the saga
	// coordinator recognises these fields and pauses the workflow until
	// a response arrives on responses_topic with in_response_to_request_id
	// matching newRequestID.
	return map[string]interface{}{
		"request_id":        newRequestID,
		"topic_sent_to":     thunderAdapterTopic,
		"requests_topic":    thunderAdapterTopic,
		"responses_topic":   myResponsesTopic,
		"await_response":    true,
		"target_agent_type": "thunder-adapter",
		"action_sent":       thunderDispatchAction,
	}, nil
}

// NOTE: defaultIfEmpty is already declared elsewhere in the actions package
// (used by web_scrape, web_search, etc.). Don't redeclare it here.
