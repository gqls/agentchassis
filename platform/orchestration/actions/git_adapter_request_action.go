// FILE: platform/orchestration/actions/git_adapter_request_action.go
//
// F1.1b(c): ONE generic chassis-side caller for the git-adapter, instead of a
// near-duplicate of GitCommitAction per adapter verb. The fix-implementer uses
// it three times — create_branch, commit (branch-targeted), and
// create_pull_request — with the adapter action name and the data mapping in
// step config:
//
//	config:
//	  adapter_action: "create_branch"
//	  data_fields:   {"branch": "commit_prep.branch"}   # resolved from CollectedData
//	  data_literals: {"repo_name": "agentchassis"}      # passed through verbatim
//
// The envelope (headers/body/reply-topic wiring) mirrors GitCommitAction —
// the adapter's contract — and the step awaits the adapter's response
// (AwaitResponse: true), so the response payload (success, sha, pr_url, …)
// lands as the step's output. The write credential never leaves the adapter.
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var GitAdapterRequestInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"adapter_action"},
	Optional:   []string{"data_fields", "data_literals"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("git_adapter_request", GitAdapterRequestInputSpec)
}

// gitAdapterActions is the allowlist of adapter verbs this generic caller may
// invoke. Deliberate: a workflow typo must fail here, loudly, not at the
// adapter's unknown-action branch; and destructive verbs (delete_repo) are
// NOT reachable through the generic caller at all.
var gitAdapterActions = map[string]bool{
	"commit":              true,
	"create_branch":       true,
	"create_pull_request": true,
}

// GitAdapterRequestAction sends one request to the git-adapter and awaits it.
func GitAdapterRequestAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "git_adapter_request"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	adapterAction, _ := config["adapter_action"].(string)
	if !gitAdapterActions[adapterAction] {
		return nil, fmt.Errorf("adapter_action %q is not allowed (want one of: commit, create_branch, create_pull_request)", adapterAction)
	}

	// Assemble the data payload: literals verbatim + fields resolved from
	// CollectedData. A field that resolves to nil is a wiring error — the
	// adapter must never receive a silently-missing key.
	data := map[string]interface{}{}
	if lits, ok := config["data_literals"].(map[string]interface{}); ok {
		for k, v := range lits {
			data[k] = v
		}
	}
	if fields, ok := config["data_fields"].(map[string]interface{}); ok {
		for k, pathRaw := range fields {
			path, _ := pathRaw.(string)
			v := datahelpers.ExtractNestedField(params.CollectedData, path)
			if v == nil {
				return nil, fmt.Errorf("data_fields[%q] = %q resolved to nothing in collected data", k, path)
			}
			data[k] = v
		}
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data payload — configure data_fields / data_literals")
	}

	correlationID := getExecutionContextValue(params, "correlation_id", params.ExecutionContext.CorrelationID)
	orchestrationID := getExecutionContextValue(params, "orchestration_id", params.ExecutionContext.OrchestrationID)
	clientID := getExecutionContextValue(params, "client_id", params.ExecutionContext.ClientID)
	myResponsesTopic := getExecutionContextValue(params, "responses_topic", params.ExecutionContext.ResponsesTopic)
	if myResponsesTopic == "" {
		if alt, ok := params.CollectedData["__my_responses_topic__"].(string); ok && alt != "" {
			myResponsesTopic = alt
		}
	}
	if myResponsesTopic == "" {
		if parent, ok := params.CollectedData["__parent_responses_topic__"].(string); ok && parent != "" {
			myResponsesTopic = parent
		}
	}

	adapterTopic := "system.adapter.git.requests"
	newRequestID := uuid.New().String()

	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":         correlationID,
			"orchestration_id":       orchestrationID,
			"client_id":              clientID,
			"step_name":              params.ExecutionContext.StepName,
			"step_id":                params.ExecutionContext.StepID,
			"request_id":             newRequestID,
			"message_type":           "request",
			"sender_agent_type":      params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":        orchestrationID,
			"responses_topic":        myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,
			"timestamp":              time.Now().UTC().Format(time.RFC3339),
			"action":                 adapterAction,
		},
		"body": map[string]interface{}{
			"action":                 adapterAction,
			"data":                   data,
			"reply_to_topic":         myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,
			"metadata": map[string]interface{}{
				"requesting_agent_id":   orchestrationID,
				"requesting_agent_type": params.ExecutionContext.Sender.AgentType,
				"requesting_step":       params.ExecutionContext.StepName,
				"client_id":             clientID,
			},
			"request_context": map[string]interface{}{
				"correlation_id":   correlationID,
				"orchestration_id": orchestrationID,
				"request_id":       newRequestID,
			},
		},
	}

	messageBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("marshal adapter request: %w", err)
	}
	rawHeaders := adapterRequest["headers"].(map[string]interface{})
	headers := make(map[string]string, len(rawHeaders))
	for k, v := range rawHeaders {
		headers[k] = fmt.Sprintf("%v", v)
	}

	if err := params.Producer.ProduceWithValidation(ctx, adapterTopic, headers, []byte(correlationID), messageBytes); err != nil {
		return nil, fmt.Errorf("send to git-adapter: %w", err)
	}

	logger.Info("git_adapter_request: sent, awaiting response",
		zap.String("adapter_action", adapterAction),
		zap.String("request_id", newRequestID),
		zap.String("correlation_id", correlationID),
		zap.String("orchestration_id", orchestrationID))

	return GitCommitResult{
		Success:       true,
		RequestID:     newRequestID,
		TopicSentTo:   adapterTopic,
		RequestsTopic: adapterTopic,
		AwaitResponse: true,
		Metadata: map[string]interface{}{
			"adapter_action": adapterAction,
		},
	}, nil
}
