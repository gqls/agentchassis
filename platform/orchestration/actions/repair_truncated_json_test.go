package actions

import (
	"encoding/json"
	"testing"
)

// First tests for repairTruncatedJSON (bugs_open/138 round-3 fix). The two
// DESTROYED-to-SALVAGED cases are the defect: the old strings.Count balancing
// read brackets inside string values as structure.
func TestRepairTruncatedJSON(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string // "" means the repair must yield the discard arm
	}{
		{
			name: "clean truncation keeps the closed prefix",
			in:   `{"reviewer":"x","verdict":"object","objections":[{"edit":1,"problem":"too broad","severity":"low"}],"missing":[],"checks":[{"sql":"SELECT count(*) FROM site`,
			want: `{"reviewer":"x","verdict":"object","objections":[{"edit":1,"problem":"too broad","severity":"low"}],"missing":[]}`,
		},
		{
			name: "open bracket inside a kept string no longer draws a spurious closer",
			in:   `{"reviewer":"x","verdict":"object","objections":[{"edit":1,"problem":"slice a[0 access unchecked","severity":"low"}],"notes":"cut he`,
			want: `{"reviewer":"x","verdict":"object","objections":[{"edit":1,"problem":"slice a[0 access unchecked","severity":"low"}]}`,
		},
		{
			name: "close bracket inside a kept string no longer suppresses a needed closer",
			in:   `{"reviewer":"x","verdict":"object","objections":[{"edit":1,"problem":"stray ] in mapping","severity":"low"},{"edit":2,"problem":"cut mid obj`,
			want: `{"reviewer":"x","verdict":"object","objections":[{"edit":1,"problem":"stray ] in mapping","severity":"low"}]}`,
		},
		{
			name: "brace inside a string is not chosen as the cut point",
			in:   `{"reviewer":"x","verdict":"object","objections":[],"notes":"jsonb path {a} then the cut lands here in prose`,
			want: `{"reviewer":"x","verdict":"object","objections":[]}`,
		},
		{
			name: "escaped quote does not end the string",
			in:   `{"reviewer":"x","verdict":"object","objections":[{"edit":1,"problem":"says \"done[\" falsely","severity":"low"}],"notes":"cut`,
			want: `{"reviewer":"x","verdict":"object","objections":[{"edit":1,"problem":"says \"done[\" falsely","severity":"low"}]}`,
		},
		{
			name: "interleaved closers are appended innermost-first",
			in:   `{"a":[{"b":[1,2]},{"c":3`,
			want: `{"a":[{"b":[1,2]}]}`,
		},
		{
			name: "production fragment from the 2026-08-02 induced round is discarded whole",
			in:   `{"reviewer":"adoption_guardian","verdict":`,
			want: "",
		},
		{
			name: "cut inside the first string is discarded whole",
			in:   `{"reviewer":"adoption_gu`,
			want: "",
		},
		{
			name: "empty input",
			in:   "",
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := repairTruncatedJSON(tc.in)
			if got != tc.want {
				t.Fatalf("repairTruncatedJSON:\n  in   %q\n  got  %q\n  want %q", tc.in, got, tc.want)
			}
			if got == "" {
				return
			}
			var v map[string]interface{}
			if err := json.Unmarshal([]byte(got), &v); err != nil {
				t.Fatalf("repaired output does not unmarshal: %v\n  out %q", err, got)
			}
		})
	}
}
