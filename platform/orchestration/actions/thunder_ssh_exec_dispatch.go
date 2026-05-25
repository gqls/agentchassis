// FILE: platform/orchestration/actions/thunder_ssh_exec_dispatch.go
//
// Dispatches an ssh_exec request to thunder-adapter. Used by the
// training-launcher agent (Phase 5) to run commands on a provisioned
// Thunder Compute instance — fetch the dataset/scripts and launch the
// backgrounded training process.
//
// Clone of DispatchThunderDecommissionAction (thunder_decommission_dispatch.go):
// same envelope construction, same publish, same AwaitResponse:true return so
// the saga coordinator pauses the workflow until thunder-adapter replies. Only
// the action name, the input fields read, and the body shape differ.
//
// REUSE NOTE: thunderAdapterTopic (const) and defaultIfEmpty (func) are already
// declared in thunder_decommission_dispatch.go in this same `actions` package.
// This file MUST NOT redeclare them — it uses them directly.
//
// Input (from CollectedData["input_data"]):
//   - provisioning_id: DB row UUID of the thunder_instances row (required —
//     the adapter resolves ip/port/ssh_user/key from it)
//   - command: the shell command to run on the instance (required)
//   - timeout_seconds: optional override for the adapter's command timeout
//
// The adapter (ssh_exec_actions.go handleSSHExec) re-marshals body into
// SSHExecRequest{ProvisioningID, Command, TimeoutSeconds}.

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
	// thunderSSHExecAction is the adapter-side action handled by
	// internal/adapters/thunder/ssh_exec_actions.go handleSSHExec.
	thunderSSHExecAction = "ssh_exec"
)

// DispatchThunderSSHExecAction publishes an ssh_exec request to thunder-adapter
// and returns AwaitResponse:true so the chassis state machine waits for the
// response (which carries exit_code, stdout, stderr, reachable) before advancing.
func DispatchThunderSSHExecAction(ctx context.Context, params ActionParams) (interface{}, error) {
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"initialized": true}, nil
	}

	params.Logger.Info("DispatchThunderSSHExecAction starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// ── Extract input fields ──
	// provisioning_id: runtime (from input_mapping). timeout_seconds: constant
	// (config) or runtime. command: either supplied directly (input_data.command)
	// OR built from command_template (a config constant) with {placeholder}
	// tokens substituted from runtime values in input_data — because
	// input_mapping can't interpolate, the action does it here.
	provisioningID := configOrInput(params, "provisioning_id")
	timeoutSeconds := configOrInput(params, "timeout_seconds")

	if provisioningID == "" {
		return nil, fmt.Errorf("dispatch_thunder_ssh_exec: provisioning_id is required (input_mapping -> input_data.provisioning_id)")
	}

	command := configOrInput(params, "command")
	if command == "" {
		// Build from a template: config holds command_template with {tokens};
		// the tokens name runtime fields resolved into input_data by input_mapping.
		tmpl := configOrInput(params, "command_template")
		if tmpl == "" {
			return nil, fmt.Errorf("dispatch_thunder_ssh_exec: need either a command (config/input_data) or a command_template in config")
		}
		command = interpolateCommandTemplate(tmpl, params)
	}
	if command == "" {
		return nil, fmt.Errorf("dispatch_thunder_ssh_exec: resolved command is empty")
	}

	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "system"
	}

	// ── Pick the responses_topic the adapter reply must route to ──
	// MUST be the PARENT (caller's) responses topic — the one the saga
	// coordinator subscribes to for matching awaited_requests rows. Using the
	// agent's OWN topic (the earlier bug) routes the reply where the chassis
	// has no consumer, so the await never resolves. Matches the corrected
	// derivation in thunder_provision_dispatch.go / thunder_decommission_dispatch.go.
	//   1. CollectedData["__parent_responses_topic__"]  ← canonical v2 envelope field
	//   2. Last-resort fallback: ExecutionContext.ResponsesTopic with a warning log
	myResponsesTopic := ""
	if parentTopic, ok := params.CollectedData["__parent_responses_topic__"].(string); ok && parentTopic != "" {
		myResponsesTopic = parentTopic
	}
	if myResponsesTopic == "" {
		myResponsesTopic = params.ExecutionContext.ResponsesTopic
		params.Logger.Warn(
			"dispatch_thunder_ssh_exec: no __parent_responses_topic__ found, falling back to own ResponsesTopic — adapter response may not be matched by the chassis",
			zap.String("fallback_topic", myResponsesTopic),
		)
	}
	if myResponsesTopic == "" {
		return nil, fmt.Errorf("dispatch_thunder_ssh_exec: could not determine a responses topic — neither __parent_responses_topic__ nor ExecutionContext.ResponsesTopic is set; cannot route adapter reply")
	}

	newRequestID := uuid.NewString()

	// ── Build the request body thunder-adapter expects ──
	// adapter.go handleSSHExec re-marshals body into SSHExecRequest.
	requestBody := map[string]interface{}{
		"action":          thunderSSHExecAction,
		"reply_to_topic":  myResponsesTopic,
		"provisioning_id": provisioningID,
		"command":         command,
	}
	// timeout_seconds is an int field on SSHExecRequest; only send it if the
	// caller specified one (parsed from the input string). Leaving it out lets
	// the adapter apply its default command timeout.
	if timeoutSeconds != "" {
		if n, err := parsePositiveInt(timeoutSeconds); err == nil && n > 0 {
			requestBody["timeout_seconds"] = n
		}
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
		"action":                  thunderSSHExecAction,

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
		return nil, fmt.Errorf("dispatch_thunder_ssh_exec: marshal envelope: %w", err)
	}

	// ── Publish to thunder-adapter's requests topic ──
	key := []byte(params.ExecutionContext.CorrelationID)

	if err := params.Producer.ProduceWithValidation(
		ctx, thunderAdapterTopic, requestHeaders, key, envelopeBytes,
	); err != nil {
		params.Logger.Error("Failed to publish ssh_exec to thunder-adapter",
			zap.String("topic", thunderAdapterTopic),
			zap.String("provisioning_id", provisioningID),
			zap.Error(err))
		return nil, fmt.Errorf("dispatch_thunder_ssh_exec: produce: %w", err)
	}

	params.Logger.Info("Dispatched ssh_exec to thunder-adapter",
		zap.String("topic", thunderAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("provisioning_id", provisioningID),
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
		"action_sent":       thunderSSHExecAction,
	}, nil
}

// parsePositiveInt parses a decimal string into an int. Kept package-local and
// tiny; returns an error for anything non-numeric so the caller can skip the
// optional field rather than send a bad value.
func parsePositiveInt(s string) (int, error) {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("not a positive integer: %q", s)
		}
		n = n*10 + int(r-'0')
	}
	if n == 0 {
		return 0, fmt.Errorf("empty or zero: %q", s)
	}
	return n, nil
}

// configOrInput returns a string value for `name`, preferring the step's
// static config (constants like a literal key, method, or command template)
// and falling back to input_data in CollectedData (runtime values resolved by
// input_mapping). Shared by the thunder dispatch actions in this package.
//
// Rationale: ResolveInputMapping only does field references — it cannot express
// a literal constant. So constants live in step config (the proven preparer
// pattern, e.g. s3_key_template), while runtime references arrive via
// input_mapping under input_data.*.
func configOrInput(params ActionParams, name string) string {
	if params.StepConfig.Config != nil {
		if v, ok := params.StepConfig.Config[name].(string); ok && v != "" {
			return v
		}
	}
	return datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data."+name)
}

// interpolateCommandTemplate replaces {token} placeholders in a command
// template with runtime values. Each token is resolved by resolveTemplateToken:
// first a config-declared source dot-path (resolved from CollectedData — this
// is how cross-step values such as a prior presign step's presigned_url reach
// the command), then input_data.<token>. The launcher keeps the static template
// and the source paths in step config (e.g. "scripts_url":
// "scripts_url_result.presigned_url"); this stitches them together. We do NOT
// use a local-step input_mapping — the coordinator doesn't resolve it for plain
// local actions.
//
// A {token} that resolves to empty is substituted with empty string and logged
// — better a visibly-broken command in launch.log than a literal "{scripts_url}"
// silently curl'd. Tokens are [a-zA-Z0-9_].
func interpolateCommandTemplate(tmpl string, params ActionParams) string {
	var b []byte
	i := 0
	for i < len(tmpl) {
		if tmpl[i] == '{' {
			// find closing brace
			j := i + 1
			for j < len(tmpl) && tmpl[j] != '}' {
				c := tmpl[j]
				if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
					(c >= '0' && c <= '9') || c == '_') {
					break // not a token char — treat '{' as literal
				}
				j++
			}
			if j < len(tmpl) && tmpl[j] == '}' && j > i+1 {
				token := tmpl[i+1 : j]
				val := resolveTemplateToken(params, token)
				if val == "" {
					params.Logger.Warn("command_template token resolved to empty; substituting empty",
						zap.String("token", token))
				}
				b = append(b, val...)
				i = j + 1
				continue
			}
		}
		b = append(b, tmpl[i])
		i++
	}
	return string(b)
}

// resolveTemplateToken resolves a {token} used in a command_template. A token
// names EITHER a config-declared source dot-path (resolved from CollectedData)
// or, failing that, a field already in input_data. Tokens are paths by
// convention, so a config value for a token is always treated as a dot-path
// reference into CollectedData — never a literal. This is what distinguishes
// tokens from key/method/command_template, which configOrInput reads as
// literals. It is how a prior step's output (e.g.
// "scripts_url_result.presigned_url") reaches the command, since the
// coordinator does not resolve input_mapping for local action steps.
//
// Uses datahelpers.ExtractNestedFieldString (the canonical resolver, with
// .response auto-unwrap) — no new path-resolution logic.
func resolveTemplateToken(params ActionParams, token string) string {
	if params.StepConfig.Config != nil {
		if path, ok := params.StepConfig.Config[token].(string); ok && path != "" {
			if v := datahelpers.ExtractNestedFieldString(params.CollectedData, path); v != "" {
				return v
			}
		}
	}
	return datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data."+token)
}
