// FILE: platform/orchestration/actions/thunder_prepare_object_url_dispatch.go
//
// Dispatches a prepare_object_url request to thunder-adapter. Used by the
// training-launcher (Phase 5) to presign B2 objects the training VM must
// fetch (the dataset JSONL) or push — by EXPLICIT key. The adapter is the
// credential boundary; it holds the B2 keys and returns a time-limited
// presigned URL. This action just relays the request and awaits the reply.
//
// Clone of DispatchThunderDecommissionAction (thunder_decommission_dispatch.go):
// identical envelope construction, publish, and AwaitResponse:true return.
// Only the action name, input fields, and body shape differ.
//
// REUSE NOTE: thunderAdapterTopic (const) and defaultIfEmpty (func) are already
// declared in thunder_decommission_dispatch.go in this same `actions` package.
// This file MUST NOT redeclare them.
//
// Input (from CollectedData["input_data"]):
//   - key: the B2 object key to presign (required). The launcher supplies the
//     concrete key; for the dataset it comes from the agent_def step config's
//     key_template with {export_id} substituted, so the template lives in ONE
//     declarative place (the migration), not hardcoded in Go.
//   - method: "GET" (default) or "PUT"
//   - expiry_minutes: optional override
//
// The adapter (data_url_actions.go handlePrepareObjectURL) re-marshals body
// into PrepareObjectURLRequest{Key, Method, ExpiryMinutes} and replies with
// {presigned_url, key, expires_at, method}.

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
	// thunderPrepareObjectURLAction is the adapter-side action handled by
	// internal/adapters/thunder/data_url_actions.go handlePrepareObjectURL.
	thunderPrepareObjectURLAction = "prepare_object_url"
)

// DispatchThunderPrepareObjectURLAction publishes a prepare_object_url request
// to thunder-adapter and returns AwaitResponse:true so the chassis state
// machine waits for the presigned URL before advancing.
func DispatchThunderPrepareObjectURLAction(ctx context.Context, params ActionParams) (interface{}, error) {
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"initialized": true}, nil
	}

	params.Logger.Info("DispatchThunderPrepareObjectURLAction starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// ── Extract input fields ──
	// Constants (a literal key, the method, an expiry) come from STEP CONFIG —
	// input_mapping only does field references, never literals (verified against
	// ResolveInputMapping). Runtime references (the dataset's resolved s3_uri,
	// or a key passed dynamically) come from input_data via input_mapping.
	// Precedence: config first (explicit constant), then input_data.
	//
	// Caller supplies EITHER an explicit key, OR an s3://bucket/key URI (e.g. the
	// preparer's dataset_uri). When given a URI we strip the scheme+bucket to
	// recover the bucket-relative key the presigner needs.
	key := configOrInput(params, "key")
	s3URI := configOrInput(params, "s3_uri")
	method := configOrInput(params, "method")
	expiryMinutes := configOrInput(params, "expiry_minutes")

	if key == "" && s3URI != "" {
		key = keyFromS3URI(s3URI)
	}
	if key == "" {
		return nil, fmt.Errorf("dispatch_thunder_prepare_object_url: a key (config or input_data) or s3_uri is required")
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
			"dispatch_thunder_prepare_object_url: no __parent_responses_topic__ found, falling back to own ResponsesTopic — adapter response may not be matched by the chassis",
			zap.String("fallback_topic", myResponsesTopic),
		)
	}
	if myResponsesTopic == "" {
		return nil, fmt.Errorf("dispatch_thunder_prepare_object_url: could not determine a responses topic — neither __parent_responses_topic__ nor ExecutionContext.ResponsesTopic is set; cannot route adapter reply")
	}

	newRequestID := uuid.NewString()

	// ── Build the request body thunder-adapter expects ──
	requestBody := map[string]interface{}{
		"action":         thunderPrepareObjectURLAction,
		"reply_to_topic": myResponsesTopic,
		"key":            key,
	}
	if method != "" {
		requestBody["method"] = method
	}
	if expiryMinutes != "" {
		if n, err := parsePositiveInt(expiryMinutes); err == nil && n > 0 {
			requestBody["expiry_minutes"] = n
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
		"action":                  thunderPrepareObjectURLAction,

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
		return nil, fmt.Errorf("dispatch_thunder_prepare_object_url: marshal envelope: %w", err)
	}

	// ── Publish to thunder-adapter's requests topic ──
	key2 := []byte(params.ExecutionContext.CorrelationID)

	if err := params.Producer.ProduceWithValidation(
		ctx, thunderAdapterTopic, requestHeaders, key2, envelopeBytes,
	); err != nil {
		params.Logger.Error("Failed to publish prepare_object_url to thunder-adapter",
			zap.String("topic", thunderAdapterTopic),
			zap.String("key", key),
			zap.Error(err))
		return nil, fmt.Errorf("dispatch_thunder_prepare_object_url: produce: %w", err)
	}

	params.Logger.Info("Dispatched prepare_object_url to thunder-adapter",
		zap.String("topic", thunderAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("object_key", key),
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
		"action_sent":       thunderPrepareObjectURLAction,
	}, nil
}

// keyFromS3URI converts an "s3://bucket/path/to/object" URI into the
// bucket-relative key "path/to/object" that the presigner expects. If the
// input isn't an s3:// URI it's returned unchanged (treated as already a key).
// Kept dependency-free (no strings import) — a single manual scan.
func keyFromS3URI(uri string) string {
	const scheme = "s3://"
	if len(uri) < len(scheme) || uri[:len(scheme)] != scheme {
		return uri // not an s3:// URI; assume it's already a key
	}
	rest := uri[len(scheme):] // "bucket/path/to/object"
	// The key is everything after the first '/' (the bucket name).
	for i := 0; i < len(rest); i++ {
		if rest[i] == '/' {
			return rest[i+1:]
		}
	}
	// No '/' after the bucket — no key component.
	return ""
}
