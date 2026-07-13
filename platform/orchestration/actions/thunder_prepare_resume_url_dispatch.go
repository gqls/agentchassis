// FILE: platform/orchestration/actions/thunder_prepare_resume_url_dispatch.go
//
// Dispatches a prepare_resume_url request to thunder-adapter (Phase D resume).
// Asks the adapter whether the run already has checkpoints in B2 and, if so, to
// presign a GET for the LATEST one so the launch can resume from it. The adapter
// (data_url_actions.go handlePrepareResumeURL → DataURLAction.ResumeURL) lists
// finetuning/checkpoints/<training_run_id>/, picks the latest ckpt-N, and replies
// {found, presigned_url, key, index, expires_at}.
//
// found=false is a VALID answer (not an error): the run has no checkpoints yet,
// so the launcher treats it as a FRESH start. assemble_upload_manifest only emits
// a manifest "resume" block when resume_url is non-empty, so wiring this step in
// unconditionally lets ONE launcher workflow serve both fresh and resume launches
// — a re-run with the same training_run_id (on a fresh box) auto-resumes from the
// latest checkpoint, a first run finds none and starts clean.
//
// Clone of DispatchThunderPrepareObjectURLAction: identical envelope, the same
// preRegisterAwaitedRequest race fix, the same ownResponsesTopic derivation, and
// AwaitResponse:true. Only the action name and body differ (training_run_id
// instead of key/method).
//
// REUSE NOTE: thunderAdapterTopic (const), defaultIfEmpty / configOrInput /
// parsePositiveInt (funcs), preRegisterAwaitedRequest (race-fix helper) and
// ownResponsesTopic (shared topic derivation) are all already declared in this
// `actions` package. This file MUST NOT redeclare them.
//
// Input:
//   - training_run_id (required) — config or input_data; the launcher's standard
//     case is input_data.training_run_id (the same source assemble_upload_manifest
//     reads), threaded in by the parent's call_agent input_mapping.
//   - expiry_minutes (optional) — literal via configOrInput (numeric-safe since
//     configOrInput coerces config scalars).

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
	// thunderPrepareResumeURLAction is the adapter-side action handled by
	// internal/adapters/thunder/data_url_actions.go handlePrepareResumeURL.
	thunderPrepareResumeURLAction = "prepare_resume_url"
)

// DispatchThunderPrepareResumeURLAction publishes a prepare_resume_url request to
// thunder-adapter and returns AwaitResponse:true so the chassis state machine
// waits for the {found, presigned_url, key, index} reply before advancing.
func DispatchThunderPrepareResumeURLAction(ctx context.Context, params ActionParams) (interface{}, error) {
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"initialized": true}, nil
	}

	params.Logger.Info("DispatchThunderPrepareResumeURLAction starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// ── Extract input fields ──
	// training_run_id: config or input_data (configOrInput). The adapter derives
	// the checkpoint prefix from it. expiry_minutes: optional literal override.
	trainingRunID := configOrInput(params, "training_run_id")
	if trainingRunID == "" {
		return nil, fmt.Errorf("dispatch_thunder_prepare_resume_url: training_run_id is required (config or input_data.training_run_id)")
	}
	expiryMinutes := configOrInput(params, "expiry_minutes")

	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "system"
	}

	// ── Pick the responses_topic the adapter reply must route to ──
	// Shared own-responses-topic derivation (NOT the parent topic). See
	// ownResponsesTopic + rationale in thunder_prepare_object_url_dispatch.go.
	myResponsesTopic := ownResponsesTopic(params)

	newRequestID := uuid.NewString()

	// ── Build the request body thunder-adapter expects ──
	requestBody := map[string]interface{}{
		"action":          thunderPrepareResumeURLAction,
		"reply_to_topic":  myResponsesTopic,
		"training_run_id": trainingRunID,
	}
	if expiryMinutes != "" {
		if n, err := parsePositiveInt(expiryMinutes); err == nil && n > 0 {
			requestBody["expiry_minutes"] = n
		}
	}

	// ── Build the chassis envelope (headers + body) ── identical to the singular
	// prepare_object_url dispatch; only action/body differ.
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
		"action":                  thunderPrepareResumeURLAction,

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
		return nil, fmt.Errorf("dispatch_thunder_prepare_resume_url: marshal envelope: %w", err)
	}

	// ── Pre-register the awaited request BEFORE the send (loop-await race fix) ──
	// Same posture as the other thunder dispatches: newRequestID is reused in the
	// header AND returned in result["request_id"], so the coordinator's later
	// InsertAwaitedRequest no-ops via ON CONFLICT. The resume probe can reply in
	// ~1s (a B2 list + local presign), so the register-before-send window matters.
	if params.DB != nil {
		if err := preRegisterAwaitedRequest(ctx, params, newRequestID,
			"",                  // target_agent_id (unknown; matches createAwaitedRequest)
			"thunder-adapter",   // target_agent_type
			thunderAdapterTopic, // requestsTopic
			myResponsesTopic,    // responsesTopic
		); err != nil {
			params.Logger.Warn("Failed to pre-register awaited request before prepare_resume_url send - response matching may race",
				zap.String("request_id", newRequestID),
				zap.Error(err))
			// Continue: race mitigation, not a hard requirement (same posture as spawn).
		}
	}

	// ── Publish to thunder-adapter's requests topic ──
	key2 := []byte(params.ExecutionContext.CorrelationID)

	if err := params.Producer.ProduceWithValidation(
		ctx, thunderAdapterTopic, requestHeaders, key2, envelopeBytes,
	); err != nil {
		params.Logger.Error("Failed to publish prepare_resume_url to thunder-adapter",
			zap.String("topic", thunderAdapterTopic),
			zap.String("training_run_id", trainingRunID),
			zap.Error(err))
		return nil, fmt.Errorf("dispatch_thunder_prepare_resume_url: produce: %w", err)
	}

	params.Logger.Info("Dispatched prepare_resume_url to thunder-adapter",
		zap.String("topic", thunderAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("training_run_id", trainingRunID),
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
		"action_sent":       thunderPrepareResumeURLAction,
	}, nil
}
