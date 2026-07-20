// FILE: platform/orchestration/agent_error_log_test.go
//
// parseReportedConditions + the sender gate — the chassis half of the
// adapter-reported conditions contract (bugs_open/011 §4 residual). The
// input shapes here are what a response body looks like AFTER the Kafka
// JSON round-trip: []interface{} of map[string]interface{}, never the
// adapter's original typed slices.

package orchestration

import "testing"

func TestParseReportedConditionsAbsentIsSilent(t *testing.T) {
	// Absent field — the healthy case for every adapter that predates or
	// never uses the contract. Not malformed: nothing was attempted.
	got, malformed, _ := parseReportedConditions(map[string]interface{}{"image_uri": "s3://x"})
	if got != nil || malformed {
		t.Errorf("absent field parsed as (%v, malformed=%v), want (nil, false)", got, malformed)
	}
}

func TestParseReportedConditionsPresentButBrokenIsMalformed(t *testing.T) {
	// bug_historian, council round 5: "present but wrong type" must NOT
	// read as healthy — a contract break upstream (object instead of list,
	// a key reshape) has to surface, or the cure inherits the disease.
	for _, bad := range []interface{}{"warning", 42, map[string]interface{}{"code": "X"}} {
		got, malformed, _ := parseReportedConditions(map[string]interface{}{"reported_conditions": bad})
		if got != nil || !malformed {
			t.Errorf("non-list %T parsed as (%v, malformed=%v), want (nil, true)", bad, got, malformed)
		}
	}
	// A non-empty list yielding nothing usable is also a broken contract.
	got, malformed, _ := parseReportedConditions(map[string]interface{}{
		"reported_conditions": []interface{}{"junk", map[string]interface{}{}},
	})
	if got != nil || !malformed {
		t.Errorf("all-junk list parsed as (%v, malformed=%v), want (nil, true)", got, malformed)
	}
	// But an EMPTY list is a benign no-op, not a break.
	got, malformed, _ = parseReportedConditions(map[string]interface{}{
		"reported_conditions": []interface{}{},
	})
	if got != nil || malformed {
		t.Errorf("empty list parsed as (%v, malformed=%v), want (nil, false)", got, malformed)
	}
}

func TestParseReportedConditionsWellFormed(t *testing.T) {
	data := map[string]interface{}{
		"reported_conditions": []interface{}{
			map[string]interface{}{
				"code":     "UNROUTED_IMAGE_KIND",
				"severity": "warning",
				"message":  "image kind \"diagram\" is not in kindProviderRouting",
				"context":  map[string]interface{}{"kind": "diagram"},
			},
		},
	}
	got, malformed, _ := parseReportedConditions(data)
	if malformed {
		t.Fatal("well-formed input flagged malformed")
	}
	if len(got) != 1 {
		t.Fatalf("parsed %d conditions, want 1", len(got))
	}
	c := got[0]
	if c.Code != "UNROUTED_IMAGE_KIND" || c.Severity != "warning" {
		t.Errorf("parsed %+v, wrong code/severity", c)
	}
	if c.Context["kind"] != "diagram" {
		t.Errorf("context not carried: %v", c.Context)
	}
}

func TestParseReportedConditionsDefaultsAndSkips(t *testing.T) {
	data := map[string]interface{}{
		"reported_conditions": []interface{}{
			// No severity → warning; no message → code stands in.
			map[string]interface{}{"code": "SOMETHING"},
			// Junk entries are skipped, not fatal, as long as something usable remains.
			"not a map",
			map[string]interface{}{},
			map[string]interface{}{"severity": "error"}, // no code AND no message
		},
	}
	got, malformed, skipped := parseReportedConditions(data)
	if malformed {
		t.Fatal("partly-usable list flagged malformed — junk beside a good entry must degrade, not fail")
	}
	if len(got) != 1 {
		t.Fatalf("parsed %d conditions, want 1 (junk skipped)", len(got))
	}
	if skipped != 3 {
		t.Errorf("skipped = %d, want 3 — dropped entries must be COUNTED, not silently discarded", skipped)
	}
	if got[0].Severity != "warning" {
		t.Errorf("missing severity defaulted to %q, want warning", got[0].Severity)
	}
	if got[0].Message != "SOMETHING" {
		t.Errorf("missing message defaulted to %q, want the code", got[0].Message)
	}
}

func TestParseReportedConditionsMixedListReportsWhatItDropped(t *testing.T) {
	// bug_historian, council round 7 — the subtle case the round-5 fix
	// missed: a list where SOME entries parse and some are junk. Without a
	// count, the survivors make the response look wholly healthy while the
	// rest vanish. That is the same silence, one level down.
	data := map[string]interface{}{
		"reported_conditions": []interface{}{
			map[string]interface{}{"code": "UNROUTED_IMAGE_KIND"},
			"junk string",
			map[string]interface{}{"code": "UNRECOGNISED_PROVIDER_HINT"},
			42,
			map[string]interface{}{"code": "REFERENCE_ANCHORS_DROPPED"},
		},
	}
	got, malformed, skipped := parseReportedConditions(data)
	if malformed {
		t.Fatal("mixed list must not be flagged wholly malformed — three entries are usable")
	}
	if len(got) != 3 {
		t.Errorf("parsed %d conditions, want 3 survivors", len(got))
	}
	if skipped != 2 {
		t.Errorf("skipped = %d, want 2 — the caller cannot warn about losses it is not told about", skipped)
	}
}

func TestParseReportedConditionsIsCapped(t *testing.T) {
	// A misbehaving adapter must not turn one response into unbounded rows.
	var raw []interface{}
	for i := 0; i < maxReportedConditionsPerResponse*3; i++ {
		raw = append(raw, map[string]interface{}{"code": "FLOOD"})
	}
	got, malformed, _ := parseReportedConditions(map[string]interface{}{"reported_conditions": raw})
	if malformed {
		t.Fatal("capped flood flagged malformed")
	}
	if len(got) != maxReportedConditionsPerResponse {
		t.Errorf("parsed %d conditions, want cap %d", len(got), maxReportedConditionsPerResponse)
	}
}

func TestSenderGateIsAnExplicitAllowlist(t *testing.T) {
	// Guardian veto, council round 5: persistence must be provably a no-op
	// for every pipeline whose reporter has not been individually reviewed
	// in. The gate is the review record.
	if !senderMayReportConditions("image-generator") {
		t.Error("image-generator is the founding sanctioned reporter and must pass the gate")
	}
	for _, unsanctioned := range []string{"", "generic", "web-scraper", "copy-paste-adapter"} {
		if senderMayReportConditions(unsanctioned) {
			t.Errorf("agent type %q passes the gate without review — the allowlist is the contract", unsanctioned)
		}
	}
}
