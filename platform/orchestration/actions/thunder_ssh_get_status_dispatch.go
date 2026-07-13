// FILE: platform/orchestration/actions/thunder_ssh_get_status_dispatch.go
//
// Dispatches an ssh_get_status request to thunder-adapter. Used by the
// thunder-training-monitor (the periodic poller that releases a detached
// training box once its run finishes and reconciles the run row). Probes
// reachability and runs an optional status_command on a provisioned Thunder
// Compute instance, identified by provisioning_id.
//
// REUSE: the adapter side already exists — internal/adapters/thunder/
// ssh_exec_actions.go handleSSHGetStatus → SSHExecAction.GetStatus, which
// re-marshals body into SSHGetStatusRequest{ProvisioningID, StatusCommand}
// and returns {provisioning_id, exit_code, stdout, stderr, reachable}. When the
// box is unreachable it returns reachable:false as a VALID answer (not an error),
// which is exactly the signal the monitor needs ("unreachable → candidate lost").
// There was no chassis-side dispatcher for it (only dispatch_thunder_ssh_exec),
// so this adds one. It is a near-clone of DispatchThunderSSHExecAction
// (thunder_ssh_exec_dispatch.go): same envelope, same publish to
// thunderAdapterTopic, same AwaitResponse:true so the saga coordinator pauses
// until the adapter replies. Only the action name and body shape differ
// (status_command instead of command, and it is OPTIONAL — the adapter applies a
// default reachability probe when empty).
//
// REUSE NOTE: thunderAdapterTopic (const), defaultIfEmpty, configOrInput and
// interpolateCommandTemplate are already declared in this `actions` package
// (thunder_decommission_dispatch.go / thunder_ssh_exec_dispatch.go). This file
// MUST NOT redeclare them — it uses them directly.
//
// Input (StepConfig.Config constants and/or CollectedData["input_data"]):
//   - provisioning_id: DB row UUID of the thunder_instances row (required —
//     the adapter resolves ip/port/ssh_user/key from it)
//   - status_command: the probe command to run (optional). Supply directly, or
//     as a command_template (config) with {token} interpolation. If absent, the
//     adapter runs its default reachability probe (`echo ready`).

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

const (
	// thunderSSHGetStatusAction is the adapter-side action handled by
	// internal/adapters/thunder/ssh_exec_actions.go handleSSHGetStatus.
	thunderSSHGetStatusAction = "ssh_get_status"
)

// DispatchThunderSSHGetStatusAction publishes an ssh_get_status request to
// thunder-adapter and returns AwaitResponse:true so the chassis state machine
// waits for the reply (reachable, exit_code, stdout, stderr) before advancing.
func DispatchThunderSSHGetStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"initialized": true}, nil
	}

	params.Logger.Info("DispatchThunderSSHGetStatusAction starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// ── Extract input fields ──
	// provisioning_id: runtime (from input_mapping). status_command: a config
	// constant (the probe script) or runtime; OR a command_template (config) with
	// {placeholder} tokens substituted from input_data — because input_mapping
	// can't interpolate, the action does it here (same helper as ssh_exec).
	provisioningID := configOrInput(params, "provisioning_id")
	if provisioningID == "" {
		return nil, fmt.Errorf("dispatch_thunder_ssh_get_status: provisioning_id is required (input_mapping -> input_data.provisioning_id)")
	}

	// status_command is OPTIONAL (unlike ssh_exec's command): an empty value lets
	// the adapter apply its default reachability probe. Prefer a direct value,
	// then a command_template.
	statusCommand := configOrInput(params, "status_command")
	if statusCommand == "" {
		if tmpl := configOrInput(params, "command_template"); tmpl != "" {
			statusCommand = interpolateCommandTemplate(tmpl, params)
		}
	}

	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "system"
	}

	// ── Pick the responses_topic the adapter reply must route to ──
	// This MUST equal the topic the saga coordinator registers the awaited request
	// against, or the adapter's reply lands where the chassis has no consumer and the
	// await never resolves. coordinator.determineResponsesTopic resolves to
	// (env RESPONSES_TOPIC | execCtx.ResponsesTopic) — i.e. the AWAITING pod's OWN
	// responses topic, not its parent's. Observed 2026-06-04: firing this worker via
	// the generic entry, the coordinator awaited on system.agent.generic.responses
	// (own/env) while __parent_responses_topic__ was system.generic.responses (the
	// CLI's reply topic, which nothing consumes) → orphaned await, row never updated.
	// DIVERGES (intentionally) from the other thunder dispatch actions, which prefer
	// __parent_responses_topic__; those are only ever called from spawned children
	// where the two coincide. Preferring ExecutionContext.ResponsesTopic makes the
	// envelope match the coordinator in BOTH the spawned and generic-entry paths.
	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		if parentTopic, ok := params.CollectedData["__parent_responses_topic__"].(string); ok && parentTopic != "" {
			myResponsesTopic = parentTopic
			params.Logger.Warn(
				"dispatch_thunder_ssh_get_status: ExecutionContext.ResponsesTopic empty; falling back to __parent_responses_topic__ — verify the chassis awaits on this topic",
				zap.String("fallback_topic", myResponsesTopic),
			)
		}
	}
	if myResponsesTopic == "" {
		return nil, fmt.Errorf("dispatch_thunder_ssh_get_status: could not determine a responses topic — neither ExecutionContext.ResponsesTopic nor __parent_responses_topic__ is set; cannot route adapter reply")
	}

	newRequestID := uuid.NewString()

	// ── Build the request body thunder-adapter expects ──
	// adapter.go handleSSHGetStatus re-marshals body into SSHGetStatusRequest.
	// status_command is only sent when non-empty so the adapter default applies.
	requestBody := map[string]interface{}{
		"action":          thunderSSHGetStatusAction,
		"reply_to_topic":  myResponsesTopic,
		"provisioning_id": provisioningID,
	}
	if statusCommand != "" {
		requestBody["status_command"] = statusCommand
	}

	// ── Build the chassis envelope (headers + body) ──
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
		"action":                  thunderSSHGetStatusAction,

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
		return nil, fmt.Errorf("dispatch_thunder_ssh_get_status: marshal envelope: %w", err)
	}

	// ── Publish to thunder-adapter's requests topic ──
	key := []byte(params.ExecutionContext.CorrelationID)

	if err := params.Producer.ProduceWithValidation(
		ctx, thunderAdapterTopic, requestHeaders, key, envelopeBytes,
	); err != nil {
		params.Logger.Error("Failed to publish ssh_get_status to thunder-adapter",
			zap.String("topic", thunderAdapterTopic),
			zap.String("provisioning_id", provisioningID),
			zap.Error(err))
		return nil, fmt.Errorf("dispatch_thunder_ssh_get_status: produce: %w", err)
	}

	params.Logger.Info("Dispatched ssh_get_status to thunder-adapter",
		zap.String("topic", thunderAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("provisioning_id", provisioningID),
		zap.Bool("has_status_command", statusCommand != ""),
		zap.String("await_responses_topic", myResponsesTopic),
	)

	// ── Return AwaitResponse so the chassis waits for the response ──
	return map[string]interface{}{
		"request_id":        newRequestID,
		"topic_sent_to":     thunderAdapterTopic,
		"requests_topic":    thunderAdapterTopic,
		"responses_topic":   myResponsesTopic,
		"await_response":    true,
		"target_agent_type": "thunder-adapter",
		"action_sent":       thunderSSHGetStatusAction,
	}, nil
}
