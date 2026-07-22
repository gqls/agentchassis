// FILE: platform/orchestration/actions/directory_claims_test.go
//
// The pure decision this layer must never get wrong: a citation_lost claim
// (the source dropped it) must never be confused with a fetch_error claim
// (the source was merely unreachable), and a recovery must supersede
// history rather than being mistaken for "no change" — same property
// evidence_citations_test.go verifies for the V5 layer, re-checked here for
// the directory registry's own status-transition rule.

package actions

import "testing"

func TestClassifyDirectoryClaimOutcome(t *testing.T) {
	cases := []struct {
		name             string
		prevStatus       string
		outcome          citationVerifyOutcome
		wantStatus       string
		wantLabel        string
		wantIsTransition bool
	}{
		{
			name:             "still found is fresh, not a transition",
			prevStatus:       "found",
			outcome:          citationVerifyOutcome{Found: true},
			wantStatus:       "found",
			wantLabel:        "fresh",
			wantIsTransition: false,
		},
		{
			name:             "still citation_lost is fresh, not a transition",
			prevStatus:       "citation_lost",
			outcome:          citationVerifyOutcome{FailClass: "citation_lost"},
			wantStatus:       "citation_lost",
			wantLabel:        "fresh",
			wantIsTransition: false,
		},
		{
			name:             "found drops to citation_lost — a transition, not a recovery",
			prevStatus:       "found",
			outcome:          citationVerifyOutcome{FailClass: "citation_lost"},
			wantStatus:       "citation_lost",
			wantLabel:        "citation_lost",
			wantIsTransition: true,
		},
		{
			name:             "found drops to fetch_error — distinct from citation_lost",
			prevStatus:       "found",
			outcome:          citationVerifyOutcome{FailClass: "fetch_error"},
			wantStatus:       "fetch_error",
			wantLabel:        "fetch_error",
			wantIsTransition: true,
		},
		{
			name:             "citation_lost recovers to found",
			prevStatus:       "citation_lost",
			outcome:          citationVerifyOutcome{Found: true},
			wantStatus:       "found",
			wantLabel:        "recovered",
			wantIsTransition: true,
		},
		{
			name:             "fetch_error recovers to found",
			prevStatus:       "fetch_error",
			outcome:          citationVerifyOutcome{Found: true},
			wantStatus:       "found",
			wantLabel:        "recovered",
			wantIsTransition: true,
		},
		{
			name:             "fetch_error to citation_lost is still a non-recovery transition",
			prevStatus:       "fetch_error",
			outcome:          citationVerifyOutcome{FailClass: "citation_lost"},
			wantStatus:       "citation_lost",
			wantLabel:        "citation_lost",
			wantIsTransition: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotStatus, gotLabel, gotIsTransition := classifyDirectoryClaimOutcome(c.prevStatus, c.outcome)
			if gotStatus != c.wantStatus {
				t.Errorf("status: got %q, want %q", gotStatus, c.wantStatus)
			}
			if gotLabel != c.wantLabel {
				t.Errorf("label: got %q, want %q", gotLabel, c.wantLabel)
			}
			if gotIsTransition != c.wantIsTransition {
				t.Errorf("isTransition: got %v, want %v", gotIsTransition, c.wantIsTransition)
			}
		})
	}
}

func TestUnmarshalJSONObject(t *testing.T) {
	if got := unmarshalJSONObject([]byte(`{"url":"https://example.com","quote":"x"}`)); got["url"] != "https://example.com" {
		t.Errorf("expected url field to round-trip, got %v", got)
	}
	// Malformed input degrades to an empty map rather than panicking — the
	// caller's ParseCitation then reports it as invalid, a visible failure
	// rather than a crash.
	if got := unmarshalJSONObject([]byte(`not json`)); len(got) != 0 {
		t.Errorf("expected empty map for malformed input, got %v", got)
	}
}

func TestJSONObjectField(t *testing.T) {
	cand := map[string]interface{}{
		"entity_links": map[string]interface{}{"docs": "https://example.com/docs"},
	}
	if got := jsonObjectField(cand, "entity_links"); got != `{"docs":"https://example.com/docs"}` {
		t.Errorf("got %q", got)
	}
	// Absent key and wrong-shape value both default to "{}", never an error —
	// these are optional enrichments, not verifiable claims.
	if got := jsonObjectField(cand, "entity_attributes"); got != "{}" {
		t.Errorf("expected {} for absent key, got %q", got)
	}
	cand["entity_attributes"] = "not-an-object"
	if got := jsonObjectField(cand, "entity_attributes"); got != "{}" {
		t.Errorf("expected {} for wrong-shaped value, got %q", got)
	}
}
