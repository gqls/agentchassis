// Tests for parseAuditFindings (bugs_open/272): the parse switch used to
// handle a JSON string and a parsed array but not a parsed object, so
// site-review-agent — whose prompt asks for {"overall_score":..., "summary":
// ..., "findings":[...]} — could never file a finding when the LLM complied
// with its own instructions. Every shape an LLM result reaches collected_data
// in is pinned here.

package actions

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap"
)

func parsedJSON(t *testing.T, s string) interface{} {
	t.Helper()
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("bad fixture: %v", err)
	}
	return v
}

const oneFinding = `{"category":"gap","severity":"high","description":"No pricing page","current_value":"missing","suggestion":"Create pricing page","acceptance_test":"A page named pricing exists","page":"pricing","work_item_type":"needs_content_page","max_fix_attempts":2}`

func TestParseAuditFindingsShapes(t *testing.T) {
	wrappedObject := `{"overall_score": 6, "summary": "one paragraph", "findings": [` + oneFinding + `,` + oneFinding + `]}`
	bareArray := `[` + oneFinding + `]`

	cases := []struct {
		name         string
		raw          interface{}
		wantCount    int
		wantParseErr bool
	}{
		{
			// The defect: site-review-agent's prompted shape, already parsed
			// into collected_data. Fails on the pre-fix switch (0 findings).
			name:      "parsed object wrapping findings array",
			raw:       parsedJSON(t, wrappedObject),
			wantCount: 2,
		},
		{
			// bugs_closed/150's shape — the only one that produced items
			// before the fix. Must keep working.
			name:      "parsed bare array",
			raw:       parsedJSON(t, bareArray),
			wantCount: 1,
		},
		{
			name:      "JSON string holding a bare array",
			raw:       bareArray,
			wantCount: 1,
		},
		{
			name:      "JSON string holding a wrapped object",
			raw:       wrappedObject,
			wantCount: 2,
		},
		{
			name:      "fenced JSON string",
			raw:       "```json\n" + bareArray + "\n```",
			wantCount: 1,
		},
		{
			name:      "parsed object whose findings value is a JSON string",
			raw:       map[string]interface{}{"findings": bareArray},
			wantCount: 1,
		},
		{
			name:      "parsed object with no findings key",
			raw:       parsedJSON(t, `{"overall_score": 6, "summary": "s"}`),
			wantCount: 0,
		},
		{
			name:         "garbage string",
			raw:          "the model apologises instead of returning JSON",
			wantCount:    0,
			wantParseErr: true,
		},
		{
			name:      "unhandled scalar",
			raw:       float64(7),
			wantCount: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings, parseError := parseAuditFindings(tc.raw, zap.NewNop())
			if len(findings) != tc.wantCount {
				t.Errorf("got %d findings, want %d", len(findings), tc.wantCount)
			}
			if (parseError != "") != tc.wantParseErr {
				t.Errorf("parseError = %q, wantParseErr = %v", parseError, tc.wantParseErr)
			}
		})
	}
}

func TestParseAuditFindingsFieldMapping(t *testing.T) {
	raw := parsedJSON(t, `{"findings":[{
		"category":"cta","severity":"medium","description":"d","suggestion":"s",
		"page":"index","fix_type":"cta_improvement","affected_component":"hero",
		"current_value":"cv","acceptance_test":"at","max_fix_attempts":3,
		"affected_pages":["index","about"]}]}`)

	findings, parseError := parseAuditFindings(raw, zap.NewNop())
	if parseError != "" || len(findings) != 1 {
		t.Fatalf("got %d findings, parseError %q", len(findings), parseError)
	}
	f := findings[0]
	if f.Category != "cta" || f.Severity != "medium" || f.Description != "d" ||
		f.Suggestion != "s" || f.Page != "index" || f.FixType != "cta_improvement" ||
		f.AffectedComponent != "hero" || f.CurrentValue != "cv" ||
		f.AcceptanceTest != "at" || f.MaxFixAttempts != 3 {
		t.Errorf("field mapping lost a value: %+v", f)
	}
	if len(f.AffectedPages) != 2 || f.AffectedPages[0] != "index" || f.AffectedPages[1] != "about" {
		t.Errorf("affected_pages mapping lost a value: %v", f.AffectedPages)
	}
}
