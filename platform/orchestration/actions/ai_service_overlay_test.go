package actions

import (
	"reflect"
	"testing"
)

// TestResolveAIServiceConfig covers bugs_open/009: a root ai_service block used
// to shadow the step-level block entirely (first-found-wins), so per-step
// max_tokens/model overrides were dead config fleet-wide. The overlay must let
// the step win PER KEY while root keys the step omits survive as defaults.
//
// The root+step fixture mirrors the live shape that proved the bug: feed-triage
// carries root max_tokens 4000 with score_relevance declaring 8192, and
// site-adoption-agent's step blocks omit max_tokens and must inherit root's.
func TestResolveAIServiceConfig(t *testing.T) {
	agentWith := func(root, stepBlock map[string]interface{}) map[string]interface{} {
		cfg := map[string]interface{}{}
		if root != nil {
			cfg["ai_service"] = root
		}
		if stepBlock != nil {
			cfg["workflow"] = map[string]interface{}{
				"steps": map[string]interface{}{
					"score_relevance": map[string]interface{}{
						"config": map[string]interface{}{"ai_service": stepBlock},
					},
				},
			}
		}
		return cfg
	}

	root := map[string]interface{}{
		"provider":        "anthropic",
		"api_key_env_var": "ANTHROPIC_API_KEY",
		"model":           "claude-sonnet-4-6",
		"max_tokens":      float64(4000),
	}

	cases := []struct {
		name        string
		agentConfig map[string]interface{}
		stepConfig  map[string]interface{}
		currentStep string
		want        map[string]interface{}
		wantSources []string
	}{
		{
			name:        "root only",
			agentConfig: agentWith(root, nil),
			currentStep: "score_relevance",
			want:        root,
			wantSources: []string{"root"},
		},
		{
			name: "step only",
			agentConfig: agentWith(nil, map[string]interface{}{
				"provider": "anthropic", "model": "claude-sonnet-5", "max_tokens": float64(8192),
			}),
			currentStep: "score_relevance",
			want: map[string]interface{}{
				"provider": "anthropic", "model": "claude-sonnet-5", "max_tokens": float64(8192),
			},
			wantSources: []string{"workflow_step"},
		},
		{
			name:        "runtime StepConfig only",
			agentConfig: map[string]interface{}{},
			stepConfig: map[string]interface{}{
				"ai_service": map[string]interface{}{"provider": "anthropic", "max_tokens": float64(1000)},
			},
			currentStep: "score_relevance",
			want:        map[string]interface{}{"provider": "anthropic", "max_tokens": float64(1000)},
			wantSources: []string{"step_config"},
		},
		{
			// THE 009 CASE: step overrides max_tokens, inherits everything else.
			name: "root plus step: step wins per key, omitted root keys survive",
			agentConfig: agentWith(root, map[string]interface{}{
				"max_tokens": float64(8192),
			}),
			currentStep: "score_relevance",
			want: map[string]interface{}{
				"provider":        "anthropic",
				"api_key_env_var": "ANTHROPIC_API_KEY",
				"model":           "claude-sonnet-4-6",
				"max_tokens":      float64(8192),
			},
			wantSources: []string{"root", "workflow_step"},
		},
		{
			// site-adoption-agent shape: step re-states model/provider, omits
			// max_tokens — root's cap must survive the overlay.
			name: "root plus step: step without max_tokens inherits root cap",
			agentConfig: agentWith(root, map[string]interface{}{
				"provider": "anthropic", "model": "claude-sonnet-4-6", "api_key_env_var": "ANTHROPIC_API_KEY",
			}),
			currentStep: "score_relevance",
			want:        root,
			wantSources: []string{"root", "workflow_step"},
		},
		{
			name: "all three sources: runtime StepConfig wins last",
			agentConfig: agentWith(root, map[string]interface{}{
				"max_tokens": float64(8192),
			}),
			stepConfig: map[string]interface{}{
				"ai_service": map[string]interface{}{"max_tokens": float64(16000)},
			},
			currentStep: "score_relevance",
			want: map[string]interface{}{
				"provider":        "anthropic",
				"api_key_env_var": "ANTHROPIC_API_KEY",
				"model":           "claude-sonnet-4-6",
				"max_tokens":      float64(16000),
			},
			wantSources: []string{"root", "workflow_step", "step_config"},
		},
		{
			name: "another step's block must not leak into this step",
			agentConfig: agentWith(root, map[string]interface{}{
				"max_tokens": float64(8192),
			}),
			currentStep: "verdict",
			want:        root,
			wantSources: []string{"root"},
		},
		{
			name:        "no block anywhere",
			agentConfig: map[string]interface{}{},
			currentStep: "score_relevance",
			want:        map[string]interface{}{},
			wantSources: nil,
		},
		{
			name:        "empty root block does not count as a source",
			agentConfig: agentWith(map[string]interface{}{}, map[string]interface{}{"max_tokens": float64(8192)}),
			currentStep: "score_relevance",
			want:        map[string]interface{}{"max_tokens": float64(8192)},
			wantSources: []string{"workflow_step"},
		},
		{
			name:        "nil agentConfig with nil stepConfig",
			agentConfig: nil,
			currentStep: "score_relevance",
			want:        map[string]interface{}{},
			wantSources: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, sources := resolveAIServiceConfig(tc.agentConfig, tc.stepConfig, tc.currentStep)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("merged config:\n got %#v\nwant %#v", got, tc.want)
			}
			if !reflect.DeepEqual(sources, tc.wantSources) {
				t.Errorf("sources: got %v want %v", sources, tc.wantSources)
			}
		})
	}
}

// TestResolveAIServiceConfig_DoesNotMutateInputs guards the overlay against
// writing through to the shared agent_definitions config map: the merged map
// must be a copy, or a runtime override would poison the cached definition.
func TestResolveAIServiceConfig_DoesNotMutateInputs(t *testing.T) {
	root := map[string]interface{}{"provider": "anthropic", "max_tokens": float64(4000)}
	agentConfig := map[string]interface{}{
		"ai_service": root,
		"workflow": map[string]interface{}{
			"steps": map[string]interface{}{
				"s1": map[string]interface{}{
					"config": map[string]interface{}{
						"ai_service": map[string]interface{}{"max_tokens": float64(8192)},
					},
				},
			},
		},
	}

	merged, _ := resolveAIServiceConfig(agentConfig, nil, "s1")
	merged["max_tokens"] = float64(999)
	merged["provider"] = "mutated"

	if root["max_tokens"] != float64(4000) || root["provider"] != "anthropic" {
		t.Errorf("root block mutated through the merged map: %#v", root)
	}
}
