// FILE: platform/orchestration/actions/thunder_prepare_object_urls_dispatch.go
//
// Dispatches a BATCH prepare_object_urls request to thunder-adapter. This is the
// plural sibling of DispatchThunderPrepareObjectURLAction: instead of presigning
// one key per awaited round-trip (the per-checkpoint loop, which re-persisted the
// whole expanded workflow each iteration → O(K^2) and crawled to a halt by iter_9
// in the 2026-06-09 run), it sends the WHOLE key array in one request and awaits
// one reply carrying all the presigned URLs. One round-trip, one state persist,
// no loop, no flatten step.
//
// REUSE NOTE: thunderAdapterTopic (const), defaultIfEmpty / configOrInput /
// parsePositiveInt (funcs), preRegisterAwaitedRequest (the race-fix helper), and
// ownResponsesTopic (the own-responses-topic derivation, shared with the singular
// dispatch) are all already declared in this `actions` package. This file MUST
// NOT redeclare them.
//
// Input:
//   - keys: a collected-data dot-path to the ordered key array (e.g. the compute
//     step's checkpoint_keys). Resolved via ExtractActionInputs exactly like
//     assemble_upload_manifest resolves its checkpoint_keys — the config VALUE is
//     a dot-path against CollectedData, not a literal. (Local steps get no
//     input_mapping, so the step declares WHERE to read the array.)
//   - method: "GET" or "PUT" (literal, via configOrInput) — checkpoints are PUT.
//   - expiry_minutes: optional literal override (via configOrInput).
//
// The adapter (data_url_actions.go handlePrepareObjectURLs) presigns each key with
// the same method/expiry and replies with {presigned_urls:[...], keys:[...], count}.
// presigned_urls is ordered 1:1 with the input keys; the launcher's
// assemble_upload_manifest reads it as checkpoint_urls (paired by index with the
// same compute checkpoint_keys), so flatten_checkpoint_urls is no longer needed.

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
	// thunderPrepareObjectURLsAction is the adapter-side action handled by
	// internal/adapters/thunder/data_url_actions.go handlePrepareObjectURLs.
	thunderPrepareObjectURLsAction = "prepare_object_urls"
)

// prepareObjectURLsSpec resolves the "keys" config dot-path to the ordered key
// array against CollectedData (same mechanism assemble_upload_manifest uses for
// checkpoint_keys, which also declares its list inputs Optional and enforces
// presence separately). method/expiry_minutes are read as literals via
// configOrInput, NOT through the spec (ExtractActionInputs resolves paths, not
// bare literals). Presence of keys is enforced by the len()==0 guard below.
var prepareObjectURLsSpec = datahelpers.ActionInputSpec{
	Optional: []string{"keys"},
}

func init() {
	datahelpers.RegisterActionInputSpec("dispatch_thunder_prepare_object_urls", prepareObjectURLsSpec)
}

// DispatchThunderPrepareObjectURLsAction publishes a batch prepare_object_urls
// request to thunder-adapter and returns AwaitResponse:true so the chassis state
// machine waits for the full list of presigned URLs before advancing.
func DispatchThunderPrepareObjectURLsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"initialized": true}, nil
	}

	params.Logger.Info("DispatchThunderPrepareObjectURLsAction starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// ── Resolve the key array (the proven ExtractActionInputs path) ──
	// The "keys" config value is a dot-path (e.g. "ckpt_keys.checkpoint_keys");
	// ExtractActionInputs resolves it against CollectedData and GetRaw returns the
	// raw list, which ExtractStringListHelper coerces to []string (dropping any
	// non-string element silently — we fail loudly on empty below).
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, prepareObjectURLsSpec, params.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("dispatch_thunder_prepare_object_urls: %w", err)
	}
	keys := datahelpers.ExtractStringListHelper(inputs.GetRaw("keys"))
	if len(keys) == 0 {
		return nil, fmt.Errorf("dispatch_thunder_prepare_object_urls: 'keys' path resolved to no keys")
	}

	method := configOrInput(params, "method")
	expiryMinutes := configOrInput(params, "expiry_minutes")

	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "system"
	}

	// ── Pick the responses_topic the adapter reply must route to ──
	// Shared derivation with the singular dispatch (the agent's OWN responses
	// topic, seeded from __my_responses_topic__; NOT the parent topic). See
	// ownResponsesTopic + the rationale in thunder_prepare_object_url_dispatch.go.
	myResponsesTopic := ownResponsesTopic(params)

	newRequestID := uuid.NewString()

	// ── Build the request body thunder-adapter expects ──
	requestBody := map[string]interface{}{
		"action":         thunderPrepareObjectURLsAction,
		"reply_to_topic": myResponsesTopic,
		"keys":           keys,
	}
	if method != "" {
		requestBody["method"] = method
	}
	if expiryMinutes != "" {
		if n, err := parsePositiveInt(expiryMinutes); err == nil && n > 0 {
			requestBody["expiry_minutes"] = n
		}
	}

	// ── Build the chassis envelope (headers + body) ── identical to the singular
	// dispatch; only action/body shape differ.
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
		"action":                  thunderPrepareObjectURLsAction,

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
		return nil, fmt.Errorf("dispatch_thunder_prepare_object_urls: marshal envelope: %w", err)
	}

	// ── Pre-register the awaited request BEFORE the send (loop-await race fix) ──
	// Same posture as the singular dispatch: newRequestID is the id used in the
	// envelope header AND returned in result["request_id"], so the coordinator's
	// later InsertAwaitedRequest reuses it and no-ops via ON CONFLICT. Closes the
	// send-before-register window. (For the batch there is only ONE awaited
	// request per launch, not K, so the race surface is tiny — but we keep the
	// register-before-send invariant uniform across all adapter dispatches.)
	if params.DB != nil {
		if err := preRegisterAwaitedRequest(ctx, params, newRequestID,
			"",                  // target_agent_id (unknown; matches createAwaitedRequest)
			"thunder-adapter",   // target_agent_type
			thunderAdapterTopic, // requestsTopic
			myResponsesTopic,    // responsesTopic
		); err != nil {
			params.Logger.Warn("Failed to pre-register awaited request before prepare_object_urls send - response matching may race",
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
		params.Logger.Error("Failed to publish prepare_object_urls to thunder-adapter",
			zap.String("topic", thunderAdapterTopic),
			zap.Int("key_count", len(keys)),
			zap.Error(err))
		return nil, fmt.Errorf("dispatch_thunder_prepare_object_urls: produce: %w", err)
	}

	params.Logger.Info("Dispatched prepare_object_urls to thunder-adapter",
		zap.String("topic", thunderAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.Int("key_count", len(keys)),
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
		"action_sent":       thunderPrepareObjectURLsAction,
	}, nil
}
