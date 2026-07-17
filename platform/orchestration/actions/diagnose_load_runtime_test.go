package actions

import (
	"reflect"
	"testing"
)

// matchAgentTypes drives the agent-state autogather section of the diagnosis
// bundle (bugs_open/008+009 class: config-shaped bugs need state-tier
// evidence). Whole-token semantics matter: "generic" must not match inside
// "generically", and "content-creator" must not match inside
// "content-creator-hero" (the longer type is its own row and matches itself).
func TestMatchAgentTypes(t *testing.T) {
	types := []string{
		"diagnose-agent", "page-content-writer", "generic",
		"content-creator-hero", "content_researcher", "tool-generator",
	}
	cases := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "names one type verbatim",
			text: "the diagnose-agent definition had a step-level ai_service",
			want: []string{"diagnose-agent"},
		},
		{
			name: "several types across a long symptom",
			text: "page-content-writer generate_content at 2000; generic propose at 2048; tool-generator generate_tool_html",
			want: []string{"page-content-writer", "generic", "tool-generator"},
		},
		{
			name: "no substring match inside a longer word",
			text: "this text speaks generically about agents",
			want: nil,
		},
		{
			name: "shorter type does not match inside a longer hyphenated type",
			text: "the content-creator-hero agent runs at the default",
			want: []string{"content-creator-hero"},
		},
		{
			name: "underscore types match",
			text: "content_researcher declares max_tokens that never applies",
			want: []string{"content_researcher"},
		},
		{
			name: "case-insensitive",
			text: "Diagnose-Agent verdict calls logged 2048",
			want: []string{"diagnose-agent"},
		},
		{
			name: "empty text matches nothing",
			text: "",
			want: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := matchAgentTypes(c.text, types)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("matchAgentTypes(%q)\n got:  %v\n want: %v", c.text, got, c.want)
			}
		})
	}
}
