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

// Regression tests for the UK-news-default rider on bugs_open/316 (owner ask
// 2026-08-31): a source's "region" config key must reach the adapter
// request, and .uk/.co.uk domains must derive "uk" at seed time.

func TestWebSearchActionPayloadCarriesRegion(t *testing.T) {
	producer := &capturingProducer{}
	params := searchTestParams(producer)
	params.StepConfig = models.Step{Config: map[string]interface{}{
		"query":  "web platform news",
		"region": "uk",
	}}

	if _, err := WebSearchAction(context.Background(), params); err != nil {
		t.Fatalf("WebSearchAction returned error: %v", err)
	}

	data := producedSearchData(t, producer)
	if data["region"] != "uk" {
		t.Fatalf("payload region = %v, want uk — the geo-target never reached the adapter", data["region"])
	}
}

func TestWebSearchActionPayloadOmitsRegionWhenNotConfigured(t *testing.T) {
	producer := &capturingProducer{}
	params := searchTestParams(producer)
	params.StepConfig = models.Step{Config: map[string]interface{}{"query": "web platform news"}}

	if _, err := WebSearchAction(context.Background(), params); err != nil {
		t.Fatalf("WebSearchAction returned error: %v", err)
	}

	data := producedSearchData(t, producer)
	if data["region"] != "" {
		t.Fatalf("payload region = %v, want empty — a site with no region config must not fabricate one", data["region"])
	}
}

func TestFetchNewsSearchPassesRegionThrough(t *testing.T) {
	producer := &capturingProducer{}
	params := searchTestParams(producer)
	params.StepConfig = models.Step{}
	params.CollectedData = map[string]interface{}{
		"input_data": map[string]interface{}{
			"source_config": map[string]interface{}{
				"query":  "boxing news",
				"region": "uk",
			},
		},
	}

	if _, err := FetchNewsSearchAction(context.Background(), params); err != nil {
		t.Fatalf("FetchNewsSearchAction returned error: %v", err)
	}

	data := producedSearchData(t, producer)
	if data["region"] != "uk" {
		t.Fatalf("payload region = %v, want uk — source_config region was not passed through", data["region"])
	}
}

func TestFetchNewsSearchOmitsRegionWhenSourceConfigHasNone(t *testing.T) {
	producer := &capturingProducer{}
	params := searchTestParams(producer)
	params.StepConfig = models.Step{}
	params.CollectedData = map[string]interface{}{
		"input_data": map[string]interface{}{
			"source_config": map[string]interface{}{"query": "boxing news"},
		},
	}

	if _, err := FetchNewsSearchAction(context.Background(), params); err != nil {
		t.Fatalf("FetchNewsSearchAction returned error: %v", err)
	}

	data := producedSearchData(t, producer)
	if data["region"] != "" {
		t.Fatalf("payload region = %v, want empty — a source with no region key must not fabricate one", data["region"])
	}
}

func TestRegionForDomain(t *testing.T) {
	cases := []struct{ domain, want string }{
		{"boxingonline.com", ""},
		{"webdesign.co.uk", "uk"},
		{"idea.uk", "uk"},
		{"WEBDESIGN.CO.UK", "uk"},
		{"notreally.uk.com", ""}, // .uk.com is a different TLD, not a .uk suffix
		{"", ""},
	}
	for _, tc := range cases {
		if got := regionForDomain(tc.domain); got != tc.want {
			t.Errorf("regionForDomain(%q) = %q, want %q", tc.domain, got, tc.want)
		}
	}
}
