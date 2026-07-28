// FILE: platform/orchestration/actions/web_search_options_test.go
//
// Regression tests for bugs_open/127: the adapter request payload must carry
// search_type and time_range, and FetchNewsSearchAction must pass an optional
// time_range from source_config through to WebSearchAction.
package actions

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

type capturingProducer struct {
	topic string
	value []byte
}

func (p *capturingProducer) Produce(_ context.Context, topic string, _ map[string]string, _, value []byte) error {
	p.topic = topic
	p.value = value
	return nil
}

func (p *capturingProducer) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	return p.Produce(ctx, topic, headers, key, value)
}

func (p *capturingProducer) Close() error { return nil }

func producedSearchData(t *testing.T, producer *capturingProducer) map[string]interface{} {
	t.Helper()
	if producer.value == nil {
		t.Fatal("no message was produced to the adapter")
	}
	var msg struct {
		Body struct {
			Data map[string]interface{} `json:"data"`
		} `json:"body"`
	}
	if err := json.Unmarshal(producer.value, &msg); err != nil {
		t.Fatalf("produced payload is not JSON: %v", err)
	}
	return msg.Body.Data
}

func searchTestParams(producer *capturingProducer) ActionParams {
	return ActionParams{
		Logger:   zap.NewNop(),
		Producer: producer,
		ExecutionContext: &types.ExecutionContext{
			CorrelationID:   "corr-1",
			OrchestrationID: "orch-1",
			ClientID:        "client-1",
			ResponsesTopic:  "system.agent.test.responses",
			Sender:          types.AgentIdentity{AgentType: "test-agent"},
		},
	}
}

func TestWebSearchActionPayloadCarriesSearchTypeAndTimeRange(t *testing.T) {
	producer := &capturingProducer{}
	params := searchTestParams(producer)
	params.StepConfig = models.Step{Config: map[string]interface{}{
		"query":       "web platform news",
		"search_type": "news",
		"time_range":  "week",
	}}

	if _, err := WebSearchAction(context.Background(), params); err != nil {
		t.Fatalf("WebSearchAction returned error: %v", err)
	}

	data := producedSearchData(t, producer)
	if data["search_type"] != "news" {
		t.Fatalf("payload search_type = %v, want news", data["search_type"])
	}
	if data["time_range"] != "week" {
		t.Fatalf("payload time_range = %v, want week — the recency control never reached the adapter", data["time_range"])
	}
}

func TestFetchNewsSearchPassesTimeRangeThrough(t *testing.T) {
	producer := &capturingProducer{}
	params := searchTestParams(producer)
	params.StepConfig = models.Step{}
	params.CollectedData = map[string]interface{}{
		"input_data": map[string]interface{}{
			"source_config": map[string]interface{}{
				"query":      "boxing news",
				"time_range": "week",
			},
		},
	}

	if _, err := FetchNewsSearchAction(context.Background(), params); err != nil {
		t.Fatalf("FetchNewsSearchAction returned error: %v", err)
	}

	data := producedSearchData(t, producer)
	if data["query"] != "boxing news" {
		t.Fatalf("payload query = %v, want the source_config query", data["query"])
	}
	if data["search_type"] != "news" {
		t.Fatalf("payload search_type = %v, want news — the action must force it", data["search_type"])
	}
	if data["time_range"] != "week" {
		t.Fatalf("payload time_range = %v, want week — source_config time_range was not passed through", data["time_range"])
	}
}
