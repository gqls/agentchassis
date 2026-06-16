// FILE: platform/orchestration/actions/analyser_request_action.go
//
// DRAFT for the agent-chassis repo. Does not compile in the contextkit
// container — built in your env.
//
// request_repo_analysis sends an "analyse" request to the analyser adapter on
// its fixed requests topic and returns AwaitResponse=true, so the orchestration
// engine registers an awaited request and resumes the workflow when the adapter
// replies (the same async-adapter pattern as WebscrapeAction → webscrape
// adapter). The reply lands in collected_data under this step's output_field
// (e.g. "repo_analysis"), where index_code_symbols then reads it.
//
// The request envelope matches what the analyser adapter parses: the JSON value
// is {headers:{...}, body:{action:"analyse", reply_to_topic, data:{owner,repo,
// ref,language}}} — the adapter reads body.action and unmarshals body.data into
// AnalyseRequest. Kafka message headers mirror the JSON headers (stringified),
// like WebscrapeAction.

package analyser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const analyserAdapterTopic = "system.adapter.analyser.requests"

// RequestRepoAnalysisResult mirrors WebscrapeResult: it signals the engine to
// await the adapter's response.
type RequestRepoAnalysisResult struct {
	Success       bool                   `json:"success"`
	RequestID     string                 `json:"request_id"`
	TopicSentTo   string                 `json:"topic_sent_to"`
	AwaitResponse bool                   `json:"await_response"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

func RequestRepoAnalysisAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	owner := resolveRAGConfigField(config, "owner_field", "owner", params.CollectedData)
	repo := resolveRAGConfigField(config, "repo_field", "repo", params.CollectedData)
	ref := resolveRAGConfigField(config, "ref_field", "ref", params.CollectedData)
	language := resolveRAGConfigField(config, "language_field", "language", params.CollectedData)
	if language == "" {
		language = "go"
	}
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("request_repo_analysis: owner and repo are required (set config.owner/repo or *_field paths)")
	}

	newRequestID := uuid.NewString()

	// Where the adapter should reply — the caller's responses topic.
	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		myResponsesTopic = os.Getenv("RESPONSES_TOPIC")
	}
	if myResponsesTopic == "" {
		myResponsesTopic = fmt.Sprintf("system.agent.%s.responses", params.ExecutionContext.Sender.AgentType)
	}

	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "default"
	}

	// JSON value: headers carry routing/identity + responses_topic (git reads
	// the reply topic there); the BODY carries action + reply_to_topic + the
	// AnalyseRequest payload under data. The adapter reads body.action and
	// unmarshals body.data — matching thunder/git/image-generator. action and
	// reply_to_topic are also left in the headers so header-reading adapters
	// still work (belt-and-braces, as WebscrapeAction does).
	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":          params.ExecutionContext.CorrelationID,
			"orchestration_id":        params.ExecutionContext.OrchestrationID,
			"orchestration_name":      params.ExecutionContext.OrchestrationName,
			"parent_orchestration_id": params.ExecutionContext.ParentOrchestrationID,
			"client_id":               clientID,
			"step_name":               params.ExecutionContext.StepName,
			"step_id":                 params.ExecutionContext.StepID,
			"request_id":              newRequestID,
			"message_type":            "request",
			"action":                  "analyse",
			"sender_agent_type":       params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":         params.ExecutionContext.OrchestrationID,
			"sender_pod_name":         params.ExecutionContext.Sender.PodName,
			"responses_topic":         myResponsesTopic,
			"parent_responses_topic":  myResponsesTopic,
			"reply_to_topic":          myResponsesTopic,
			"timestamp":               time.Now().UTC().Format(time.RFC3339),
		},
		"body": map[string]interface{}{
			"action":         "analyse",
			"reply_to_topic": myResponsesTopic,
			"data": map[string]interface{}{
				"owner":    owner,
				"repo":     repo,
				"ref":      ref,
				"language": language,
			},
		},
	}

	// Kafka message headers (stringified) — mirror the JSON headers, like webscrape.
	rawHeaders := adapterRequest["headers"].(map[string]interface{})
	headers := make(map[string]string, len(rawHeaders))
	for k, v := range rawHeaders {
		if s, ok := v.(string); ok {
			headers[k] = s
		} else {
			headers[k] = fmt.Sprintf("%v", v)
		}
	}

	messageBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("request_repo_analysis: marshal request: %w", err)
	}
	key := []byte(params.ExecutionContext.CorrelationID)

	logger.Info("request_repo_analysis: sending to analyser adapter",
		zap.String("topic", analyserAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("reply_to_topic", myResponsesTopic),
		zap.String("owner", owner),
		zap.String("repo", repo),
		zap.String("ref", ref),
		zap.String("language", language))

	if err := params.Producer.ProduceWithValidation(ctx, analyserAdapterTopic, headers, key, messageBytes); err != nil {
		return nil, fmt.Errorf("request_repo_analysis: send to analyser adapter: %w", err)
	}

	return &RequestRepoAnalysisResult{
		Success:       true,
		RequestID:     newRequestID,
		TopicSentTo:   analyserAdapterTopic,
		AwaitResponse: true,
		Metadata: map[string]interface{}{
			"owner":           owner,
			"repo":            repo,
			"ref":             ref,
			"language":        language,
			"responses_topic": myResponsesTopic,
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
		},
	}, nil
}
