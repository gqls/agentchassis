// Tests for deriveRenderMode, which had none before 035 P1 added its third
// value. The ordering cases are the point: both wrong orderings fail SILENTLY —
// the row simply takes a path that renders a composite parent without its
// children — so nothing downstream would report them.
package actions

import "testing"

func TestDeriveRenderMode(t *testing.T) {
	cases := []struct {
		name   string
		schema string
		want   string
	}{
		{"empty schema", "", "template"},
		{"empty object", "{}", "template"},
		{"malformed json", `{"fields":`, "template"},
		{"no llm fields", `{"fields":{"title":{"source":"config"}}}`, "template"},
		{"an llm field", `{"fields":{"body":{"source":"llm"}}}`, "agent"},
		{"no fields block at all", `{"other":1}`, "template"},

		// 035 D3 — the third value.
		{
			name:   "slots block derives composite",
			schema: `{"slots":[{"key":"lead","function":"prose-block","required":true}]}`,
			want:   "composite",
		},
		{
			// THE ORDERING FALSIFIER, and D3's own worked example verbatim: a
			// composite that ALSO declares an llm field. With the llm-field loop
			// first this returns "agent" and the row never routes to the
			// composition build.
			name:   "slots WIN over an llm field (D3's worked example)",
			schema: `{"fields":{"standfirst":{"source":"llm","required":false}},"slots":[{"key":"lead","function":"prose-block","required":true}]}`,
			want:   "composite",
		},
		{
			// The second silent ordering failure: a composite with NO fields
			// block falls out of the missing-`fields` early return as "template"
			// unless slots are tested before it.
			name:   "slots with no fields block still derive composite",
			schema: `{"slots":[{"key":"lead","required":true}]}`,
			want:   "composite",
		},
		{
			// Opt-in with the unsafe side OFF: an empty or malformed slots block
			// is not a composite (RFC_022's test, 035 §7).
			name:   "empty slots array is not a composite",
			schema: `{"slots":[],"fields":{"body":{"source":"llm"}}}`,
			want:   "agent",
		},
		{
			name:   "slots entries with no key are not a composite",
			schema: `{"slots":[{"required":true}]}`,
			want:   "template",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := deriveRenderMode(c.schema); got != c.want {
				t.Errorf("deriveRenderMode(%s) = %q, want %q", c.schema, got, c.want)
			}
		})
	}
}
