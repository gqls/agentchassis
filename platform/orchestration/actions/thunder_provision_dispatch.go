// FILE: platform/orchestration/actions/thunder_provision_dispatch.go
//
// Dispatches a provision_instance request to thunder-adapter. Used by
// the gpu-provisioner agent (Phase 3.6) — the real implementation that
// replaces the migration-022 stub.
//
// Mirrors thunder_decommission_dispatch.go in shape: build the chassis
// envelope, publish to system.adapter.thunder.requests, return
// AwaitResponse: true so the chassis state machine pauses the workflow
// until the adapter's response arrives (typically 3-5 min for a real
// provision: Thunder API create + WaitForRunning poll).
//
// Input (from CollectedData["input_data"]):
//   - training_run_id   (optional) — links the instance to a training run row
//   - gpu               (optional) — "a100", "h100", "t4" (default in adapter)
//   - num_gpus          (optional) — int, default 1
//   - vcpus             (optional) — int, default 4
//   - disk_size_gb      (optional) — int, default 100
//   - mode              (optional) — "prototyping" or "production"
//   - template          (optional) — Thunder image template
//
// Response shape (matches what model-trainer's call_launcher reads):
//   instance_ip, ssh_user, ssh_key_secret_name, provisioning_id,
//   thunder_identifier, provisioned_at

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
	// Same fixed topic as the decommission dispatch action.
	// (thunderAdapterTopic is declared in thunder_decommission_dispatch.go.)

	// thunderProvisionAction is the adapter-side action handled by
	// internal/adapters/thunder/adapter.go handleProvisionInstance.
	thunderProvisionAction = "provision_instance"
)

// DispatchThunderProvisionAction publishes a provision_instance request
// to thunder-adapter. Returns immediately with AwaitResponse: true; the
// chassis state machine waits for the response before advancing.
//
// The waiting step's timeout_seconds should be at least 360 (6 min):
// the adapter's WaitForRunning has a 5-min internal deadline, and we
// want a small buffer above that.
func DispatchThunderProvisionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Early-return for initialize calls — deterministic actions no-op.
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"initialized": true}, nil
	}

	params.Logger.Info("DispatchThunderProvisionAction starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// ── Extract input fields ──
	// All optional. The thunder-adapter applies defaults for any missing
	// field (see internal/adapters/thunder/provision_action.go resolveDefaults).
	trainingRunID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.training_run_id")
	gpu := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.gpu")
	mode := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.mode")
	template := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.template")
	numGPUs := datahelpers.ExtractNestedFieldInt(params.CollectedData, "input_data.num_gpus")
	vcpus := datahelpers.ExtractNestedFieldInt(params.CollectedData, "input_data.vcpus")
	diskSizeGB := datahelpers.ExtractNestedFieldInt(params.CollectedData, "input_data.disk_size_gb")

	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "system"
	}

	// ── Determine which responses topic to await the response on ──
	// Same logic as dispatch_thunder_decommission.
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
	// adapter.go handleProvisionInstance re-marshals body into
	// ProvisionInstanceRequest. Field names must match.
	requestBody := map[string]interface{}{
		"action":         thunderProvisionAction,
		"reply_to_topic": myResponsesTopic,
	}
	// Only set fields that have non-zero values — let the adapter apply
	// defaults for the rest.
	if trainingRunID != "" {
		requestBody["training_run_id"] = trainingRunID
	}
	if gpu != "" {
		requestBody["gpu"] = gpu
	}
	if numGPUs > 0 {
		requestBody["num_gpus"] = numGPUs
	}
	if vcpus > 0 {
		requestBody["vcpus"] = vcpus
	}
	if diskSizeGB > 0 {
		requestBody["disk_size_gb"] = diskSizeGB
	}
	if mode != "" {
		requestBody["mode"] = mode
	}
	if template != "" {
		requestBody["template"] = template
	}

	// ── Build the chassis envelope ──
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
		"action":                  thunderProvisionAction,

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
		return nil, fmt.Errorf("dispatch_thunder_provision: marshal envelope: %w", err)
	}

	// ── Publish ──
	key := []byte(params.ExecutionContext.CorrelationID)

	if err := params.Producer.ProduceWithValidation(
		ctx, thunderAdapterTopic, requestHeaders, key, envelopeBytes,
	); err != nil {
		params.Logger.Error("Failed to publish provision to thunder-adapter",
			zap.String("topic", thunderAdapterTopic),
			zap.String("training_run_id", trainingRunID),
			zap.Error(err))
		return nil, fmt.Errorf("dispatch_thunder_provision: produce: %w", err)
	}

	params.Logger.Info("Dispatched provision_instance to thunder-adapter",
		zap.String("topic", thunderAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("training_run_id", trainingRunID),
		zap.String("gpu", gpu),
		zap.String("mode", mode),
		zap.String("await_responses_topic", myResponsesTopic),
	)

	// ── Return AwaitResponse so the chassis pauses the workflow ──
	return map[string]interface{}{
		"request_id":        newRequestID,
		"topic_sent_to":     thunderAdapterTopic,
		"requests_topic":    thunderAdapterTopic,
		"responses_topic":   myResponsesTopic,
		"await_response":    true,
		"target_agent_type": "thunder-adapter",
		"action_sent":       thunderProvisionAction,
	}, nil
}
