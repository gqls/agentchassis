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
//   - key: the B2 object key to presign (required unless key_path/s3_uri given).
//     The launcher supplies the concrete key; for the dataset it comes from the
//     agent_def step config's key_template with {export_id} substituted, so the
//     template lives in ONE declarative place (the migration), not hardcoded in Go.
//   - key_path: a collected-data dot-path to a dynamic key (e.g. "ckpt_keys.final_key"),
//     for plain local steps where input_mapping is dead. Consulted only if no key.
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

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
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
	// key/method/expiry are literals read from STEP CONFIG. configOrInput reads
	// config first, then input_data.<name>. We do NOT use a local-step
	// input_mapping: the coordinator only resolves input_mapping for call_agent
	// and loop, so on a plain local action it would be dead config. Runtime
	// values instead arrive in input_data (the parent threads them in via
	// call_agent's input_mapping when it calls us).
	//
	// Source for the object, in order of precedence:
	//   1. explicit `key` (config or input_data)
	//   2. explicit `s3_uri` (an s3://bucket/key passed directly)
	//   3. the preparer's `dataset_uri` in input_data — the launcher's standard
	//      case: call_launcher threads preparation_result.dataset_uri into the
	//      launcher's input_data, and we derive the bucket-relative key from it.
	// When given an s3:// URI we strip the scheme+bucket to recover the key.
	key := configOrInput(params, "key")
	// key_path (ADDED Phase B): resolve a dynamic, bucket-relative key from a
	// collected-data dot-path (e.g. "ckpt_keys.final_key"). Needed for PLAIN LOCAL
	// steps, where input_mapping is dead so a cross-step value can't arrive via
	// input_data.key — the step declares WHERE to read the key, exactly as
	// ssh_exec_launch declares scripts_url/dataset_url. Backward-compatible: only
	// consulted when no explicit key was supplied. No existing source changed.
	if key == "" {
		if kp, ok := params.StepConfig.Config["key_path"].(string); ok && kp != "" {
			key = datahelpers.ExtractNestedFieldString(params.CollectedData, kp)
		}
	}
	s3URI := configOrInput(params, "s3_uri")
	method := configOrInput(params, "method")
	expiryMinutes := configOrInput(params, "expiry_minutes")

	if key == "" && s3URI == "" {
		// Launcher path: derive from the preparer's dataset_uri in input_data.
		s3URI = configOrInput(params, "dataset_uri")
	}
	if key == "" && s3URI != "" {
		key = keyFromS3URI(s3URI)
	}
	if key == "" {
		return nil, fmt.Errorf("dispatch_thunder_prepare_object_url: need a key (config/input_data), an s3_uri, or input_data.dataset_uri")
	}

	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "system"
	}

	// ── Pick the responses_topic the adapter reply must route to ──
	// SAME derivation as the proven dispatch actions (provision/decommission):
	// the agent's OWN responses topic. This agent's saga coordinator awaits the
	// adapter reply on this topic and matches it against the awaited_request in
	// THIS orchestration's state. ExecutionContext.ResponsesTopic is populated by
	// the coordinator from __my_responses_topic__ first (coordinator.go ~L1352).
	//
	// NOT __parent_responses_topic__: that is the topic the child uses to notify
	// its PARENT of its FINAL result (notifyParentOfSuccess/Failure). Using it for
	// an intermediate adapter reply routes the reply to the parent, where no
	// awaited_request matches — the launcher would publish fine but hang on the
	// await. (provision/decommission, which work in production, use this
	// own-topic derivation.)
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

	// ── Pre-register the awaited request BEFORE the send (loop-await race fix) ──
	// The adapter can reply in ~1s. Without this, the send happens first and the
	// coordinator only inserts the awaited_requests row afterwards (in
	// processAwaitResponse → InsertAwaitedRequest); a fast reply can land before
	// the row exists, ClaimAwaitedRequest (WHERE status='waiting') finds nothing,
	// the reply is dropped, the request times out and re-dispatches forever. We
	// register first — exactly as spawn_agent does via preRegisterAwaitedRequest
	// (spawn_actions.go, same `actions` package so no import needed).
	//
	// Consistency with the coordinator's later insert (so that insert cleanly
	// no-ops instead of creating a second row):
	//   - newRequestID is the SAME id we put in the envelope header and return in
	//     result["request_id"], so processAwaitResponse → InsertAwaitedRequest
	//     reuses it and hits ON CONFLICT (request_id) DO NOTHING ("already
	//     exists" → treated as success, no duplicate timeout goroutine).
	//   - step_name = params.CurrentStep, which buildActionParams sets to
	//     state.CurrentStep — identical to what createAwaitedRequest writes, so
	//     the response resumes the same (expanded loop) step.
	//   - target_agent_id "" matches createAwaitedRequest's extractTargetAgentID
	//     (our result carries no agent_id key).
	// Arg order is (requestsTopic, responsesTopic) — requests first.
	//
	// NOTE: preRegisterAwaitedRequest hardcodes a 120s await timeout; via the
	// ON CONFLICT path that pins every presign dispatch's await to 120s
	// regardless of the step's configured timeout. Fine for ~1s presigns.
	if params.DB != nil {
		if err := preRegisterAwaitedRequest(ctx, params, newRequestID,
			"",                  // target_agent_id (unknown; matches createAwaitedRequest)
			"thunder-adapter",   // target_agent_type
			thunderAdapterTopic, // requestsTopic
			myResponsesTopic,    // responsesTopic
		); err != nil {
			params.Logger.Warn("Failed to pre-register awaited request before prepare_object_url send - response matching may race",
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
