package actions

import (
	"encoding/json"
	"strings"
	"testing"
)

// The two payload shapes below are the LIVE ones, copied from the database rather
// than composed, because a fixture written to exercise a rule will exercise it.
// Both come from the 2026-08-12 census in bugs_open/213 §B/§D.

// fixerEnvelope is what 4 of the 14 completed dark_section_audit rows carry: the
// color-variable-fixer response, both repair steps reporting zero.
func fixerEnvelope(fixTotal, textTotal interface{}) map[string]interface{} {
	return map[string]interface{}{
		"response_status":      "complete",
		"response_received_at": "2026-08-11T13:38:18Z",
		"response": map[string]interface{}{
			"fix_result": map[string]interface{}{
				"site_id":         "1368e337-dd1d-4799-bbb3-8221a1b79bcc",
				"total_fixed":     fixTotal,
				"templates_fixed": fixTotal,
				"rendered_fixed":  fixTotal,
				"needs_rerender":  false,
				"details":         nil,
			},
			"text_color_result": map[string]interface{}{
				"site_id":              "1368e337-dd1d-4799-bbb3-8221a1b79bcc",
				"total_fixed":          textTotal,
				"templates_fixed":      textTotal,
				"rendered_fixed":       textTotal,
				"contracts_added":      0,
				"skipped_low_contrast": 0,
				"needs_rerender":       false,
				"details":              nil,
			},
		},
	}
}

// designSystemPayload is what the OTHER 10 carry — a design-system spec with no
// response envelope at all. Why these rows exist is NOT ESTABLISHED
// (bugs_open/213 §D); the gate must abstain on them, not guess.
func designSystemPayload() map[string]interface{} {
	return map[string]interface{}{
		"spacing":      map[string]interface{}{"section_padding": "96px 0", "container_max_width": "1200px"},
		"typography":   map[string]interface{}{"base_size": "16px", "line_height": "1.6"},
		"color_scheme": map[string]interface{}{"text": "#E6EDF3", "primary": "#0D1117"},
		"design_notes": "Color rationale: the near-black blue-tinted background…",
	}
}

func TestHandlerReportedNoChange(t *testing.T) {
	cases := []struct {
		name           string
		itemType       string
		result         map[string]interface{}
		wantNoChange   bool
		wantUnknown    bool
		detailContains string
	}{
		{
			// THE OPT-IN DEFAULT. An item_type that has not asked for this gate takes
			// a map miss, so a payload that WOULD be blocked for dark_section_audit
			// passes untouched. This is the assertion that the owner's 2026-08-02
			// unsafe-default-OFF ruling is satisfied by construction, and it is first
			// deliberately: if it ever fails, every item type in the fleet is affected.
			name:         "unregistered item_type is inert even with all counters zero",
			itemType:     "hardcoded_section_colors",
			result:       fixerEnvelope(float64(0), float64(0)),
			wantNoChange: false,
			wantUnknown:  false,
		},
		{
			name:           "opted-in type, both counters zero as float64 (the JSON round-trip form)",
			itemType:       "dark_section_audit",
			result:         fixerEnvelope(float64(0), float64(0)),
			wantNoChange:   true,
			detailContains: "handler reported 0 changes at response.fix_result.total_fixed and response.text_color_result.total_fixed",
		},
		{
			// Same data, the form a Go action's own return value arrives in. A type
			// switch missing this would read "counter absent" for a counter that is
			// present and zero — reporting unknown shape, the opposite verdict.
			name:         "opted-in type, both counters zero as int",
			itemType:     "dark_section_audit",
			result:       fixerEnvelope(0, 0),
			wantNoChange: true,
		},
		{
			name:         "opted-in type, both counters zero as json.Number",
			itemType:     "dark_section_audit",
			result:       fixerEnvelope(json.Number("0"), json.Number("0")),
			wantNoChange: true,
		},
		{
			// The gate's whole job is to stay out of the way when work happened. One
			// non-zero counter anywhere is enough, even beside a zero.
			name:         "opted-in type, one counter non-zero — not this gate's business",
			itemType:     "dark_section_audit",
			result:       fixerEnvelope(float64(0), float64(3)),
			wantNoChange: false,
			wantUnknown:  false,
		},
		{
			name:         "opted-in type, both counters non-zero",
			itemType:     "dark_section_audit",
			result:       fixerEnvelope(float64(2), float64(5)),
			wantNoChange: false,
			wantUnknown:  false,
		},
		{
			// A partial payload must still be judged on what it DOES carry, or a
			// handler could escape the gate by dropping a field.
			name:     "opted-in type, one counter zero and the other absent → still blocked, absence named",
			itemType: "dark_section_audit",
			result: map[string]interface{}{
				"response": map[string]interface{}{
					"fix_result": map[string]interface{}{"total_fixed": float64(0)},
				},
			},
			wantNoChange:   true,
			detailContains: "no value present at response.text_color_result.total_fixed",
		},
		{
			// The live 10-of-14 case. Abstain and say what was there instead.
			name:           "opted-in type, payload is not this handler's → abstain, naming the keys found",
			itemType:       "dark_section_audit",
			result:         designSystemPayload(),
			wantNoChange:   false,
			wantUnknown:    true,
			detailContains: "color_scheme design_notes spacing typography",
		},
		{
			name:         "opted-in type, empty payload → abstain",
			itemType:     "dark_section_audit",
			result:       map[string]interface{}{},
			wantNoChange: false,
			wantUnknown:  true,
		},
		{
			// A path that runs into a non-map mid-way must not panic or be read as a
			// zero. total_fixed is nested under a string here.
			name:         "opted-in type, counter path blocked by a non-map → abstain",
			itemType:     "dark_section_audit",
			result:       map[string]interface{}{"response": "complete"},
			wantNoChange: false,
			wantUnknown:  true,
		},
		{
			// A counter present but of an unreadable type must abstain rather than be
			// silently treated as zero — "0" the string is not a measurement.
			name:     "opted-in type, counter is a string → abstain rather than assume zero",
			itemType: "dark_section_audit",
			result: map[string]interface{}{
				"response": map[string]interface{}{
					"fix_result":        map[string]interface{}{"total_fixed": "0"},
					"text_color_result": map[string]interface{}{"total_fixed": "0"},
				},
			},
			wantNoChange: false,
			wantUnknown:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			detail, noChange, unknown := handlerReportedNoChange(tc.itemType, tc.result)

			if noChange != tc.wantNoChange {
				t.Fatalf("noChange = %v, want %v (detail=%q unknown=%q)", noChange, tc.wantNoChange, detail, unknown)
			}
			if gotUnknown := unknown != ""; gotUnknown != tc.wantUnknown {
				t.Fatalf("unknownShape non-empty = %v, want %v (got %q)", gotUnknown, tc.wantUnknown, unknown)
			}
			if tc.detailContains != "" {
				haystack := detail + unknown
				if !strings.Contains(haystack, tc.detailContains) {
					t.Fatalf("expected text %q in detail/unknown, got detail=%q unknown=%q",
						tc.detailContains, detail, unknown)
				}
			}
			// A blocked completion must always carry the roster's evidence, because
			// this string is what an operator reads off the item's error column.
			if noChange && !strings.Contains(detail, noChangeGates[tc.itemType].Why) {
				t.Fatalf("blocking detail does not carry the rule's Why: %q", detail)
			}
			// The three outcomes are mutually exclusive by construction; assert it, so
			// a future edit cannot produce a block that also reports unknown shape.
			if noChange && unknown != "" {
				t.Fatalf("both noChange and unknownShape set: %q / %q", detail, unknown)
			}
		})
	}
}

// TestNoChangeGatesRosterCarriesItsEvidence guards the roster itself rather than
// the logic. Adding a type here is an assertion about somebody else's handler —
// that a zero-change run cannot be a repair for it — and the estate's rule is that
// such an assertion arrives with the measurement that licenses it.
func TestNoChangeGatesRosterCarriesItsEvidence(t *testing.T) {
	if len(noChangeGates) == 0 {
		t.Fatal("roster is empty — this gate is inert; if that is deliberate, delete the file rather than leaving a mechanism nothing drives")
	}
	for itemType, rule := range noChangeGates {
		if strings.TrimSpace(rule.Why) == "" {
			t.Errorf("%s: no Why — an operator reading a blocked item gets no reason, and no reviewer can check the claim", itemType)
		}
		if len(rule.CounterPaths) == 0 {
			t.Errorf("%s: no CounterPaths — the gate can never fire, so the entry only looks like protection", itemType)
		}
		for _, p := range rule.CounterPaths {
			if !strings.Contains(p, ".") {
				t.Errorf("%s: counter path %q is a bare key; the handler's counts live under its response envelope", itemType, p)
			}
		}
	}
}

// TestBlockedCompletionReasonDistinguishesNoChange asserts the fourth blocking
// cause gets its own sentence and reason code. The three existing causes are
// asserted alongside it deliberately: this function's whole history is a message
// that recorded a finding the verifier never made, and the way that regresses is
// one arm being made to answer for another.
func TestBlockedCompletionReasonDistinguishesNoChange(t *testing.T) {
	cases := []struct {
		name       string
		payload    map[string]interface{}
		wantReason string
		wantInMsg  string
	}{
		{
			name:       "no-change",
			payload:    map[string]interface{}{"status": "handler_reported_no_change", "detail": "handler reported 0 changes at X"},
			wantReason: "handler_reported_no_change",
			wantInMsg:  "reported it changed nothing",
		},
		{
			name:       "out of scope",
			payload:    map[string]interface{}{"status": "out_of_scope", "detail": "spec carries no check key"},
			wantReason: "verifier_scope_mismatch",
			wantInMsg:  "does not grade this item",
		},
		{
			name:       "verifier error",
			payload:    map[string]interface{}{"status": "error", "error": "dial tcp: refused"},
			wantReason: "verification_unavailable",
			wantInMsg:  "could not run",
		},
		{
			name:       "defect persists",
			payload:    map[string]interface{}{"status": "defect_persists", "detail": "3 components still match"},
			wantReason: "verification_failed",
			wantInMsg:  "still present",
		},
	}

	seen := map[string]string{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, reason := blockedCompletionReason(tc.payload)
			if reason != tc.wantReason {
				t.Errorf("reason = %q, want %q", reason, tc.wantReason)
			}
			if !strings.Contains(msg, tc.wantInMsg) {
				t.Errorf("message %q does not contain %q", msg, tc.wantInMsg)
			}
			if prev, dup := seen[reason]; dup {
				t.Errorf("reason code %q already used by %q — the codes are what census queries group by", reason, prev)
			}
			seen[reason] = tc.name
			// The detail must survive into the message: an operator reading the error
			// column gets this string and nothing else.
			if d, _ := tc.payload["detail"].(string); d != "" && !strings.Contains(msg, d) {
				t.Errorf("message dropped the detail %q: %q", d, msg)
			}
		})
	}
}
