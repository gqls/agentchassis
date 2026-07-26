package datahelpers

import (
	"strings"
	"testing"
)

// Symbols are prefixed `stats…` so this file never collides with claims_test.go's
// testEvidenceBaseJSON / mustParseTestEB / scanAll.

const statsEvidenceBaseJSON = `{
  "facts": [
    {"id": "grippers", "claim": "gripper models indexed", "value": 10, "kind": "count",
     "tolerance": "exact", "context_terms": ["gripper model"],
     "source": {"sql": "SELECT count(*) FROM products"}, "verified_at": "2026-07-22"},
    {"id": "agents", "claim": "agent definitions", "value": 170, "kind": "count",
     "tolerance": "exact", "context_terms": ["agent definition"],
     "source": {"sql": "SELECT count(*) FROM agent_definitions"}, "verified_at": "2026-07-24"}
  ],
  "banned_claims": [],
  "allowed_entities": []
}`

func mustParseStatsEB(t *testing.T) *EvidenceBase {
	t.Helper()
	eb, err := ParseEvidenceBase([]byte(statsEvidenceBaseJSON))
	if err != nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	if eb == nil {
		t.Fatal("ParseEvidenceBase returned nil for a base with facts")
	}
	return eb
}

// TestStatCardHTMLSplitsValueFromLabel pins the mechanism that makes the HTML
// claims lane structurally blind to a stat block, and is the reason
// claims_stats.go exists at all.
//
// `div` is in assertionBlockElements, so extractAssertions flushes between the
// value and its label. The number's block is the bare figure, its claimWindow
// therefore contains no label text, and businessClaimContextRe can never match
// — the candidate is dropped before numberSupported is consulted.
//
// If someone later widens assertionBlockElements, this test fails and that
// change becomes a deliberate decision rather than an accident.
func TestStatCardHTMLSplitsValueFromLabel(t *testing.T) {
	const markup = `<div class="stat-value">170<span class="stat-suffix"></span></div>` +
		`<div class="stat-label">Agent Definitions</div>`

	blocks := ExtractAssertionText(markup)
	if len(blocks) != 2 {
		t.Fatalf("expected the stat card to split into 2 blocks, got %d: %q", len(blocks), blocks)
	}
	if blocks[0] != "170" {
		t.Errorf("expected the value to stand alone as %q, got %q", "170", blocks[0])
	}
	if blocks[1] != "Agent Definitions" {
		t.Errorf("expected the label in its own block, got %q", blocks[1])
	}

	// And therefore: the HTML numeric scan reports nothing, even though 170
	// with no supporting fact would be a finding if it could see the label.
	eb := mustParseStatsEB(t)
	if got := eb.ScanUnregisteredNumbers(blocks); len(got) != 0 {
		t.Fatalf("precondition changed: the HTML scan now sees a stat card (%d findings) — "+
			"claims_stats.go's rationale needs revisiting", len(got))
	}
}

// TestStatPairIsSeenViaContentData is the regression case: the same figure and
// label, read from content_data instead of HTML, IS audited.
func TestStatPairIsSeenViaContentData(t *testing.T) {
	eb := mustParseStatsEB(t)

	// The fabrication bug 043 recorded live on robot-hands.
	fabricated := map[string]interface{}{
		"stat1_value":  "2,400+",
		"stat1_suffix": "",
		"stat1_label":  "Gripper Models Indexed",
	}
	claims := ExtractStatClaims("system-stats", fabricated)
	if len(claims) != 1 {
		t.Fatalf("expected 1 stat claim, got %d: %+v", len(claims), claims)
	}
	if claims[0].Pairing != StatPairedExact || claims[0].Label != "Gripper Models Indexed" {
		t.Fatalf("expected an exact pairing to the label, got %+v", claims[0])
	}

	// The prose gate does NOT match this label — which is the point: the stat
	// lane must not be filtered by it.
	if businessClaimContextRe.MatchString(claims[0].Window()) {
		t.Fatal("test premise broken: businessClaimContextRe now matches " +
			"'Gripper Models Indexed', so this case no longer proves the gate is bypassed")
	}

	findings := eb.ScanStatClaims(claims)
	if len(findings) != 1 {
		t.Fatalf("expected the fabricated figure to be flagged, got %d findings", len(findings))
	}
	if findings[0].Check != "unregistered_stat" {
		t.Errorf("check = %q, want unregistered_stat", findings[0].Check)
	}
	if findings[0].Pattern != "system-stats.stat1_value" {
		t.Errorf("location = %q, want system-stats.stat1_value", findings[0].Pattern)
	}

	// The corrected, query-traced value passes.
	corrected := map[string]interface{}{
		"stat1_value":  "10",
		"stat1_suffix": "",
		"stat1_label":  "Gripper Models Indexed",
	}
	if got := eb.ScanStatClaims(ExtractStatClaims("system-stats", corrected)); len(got) != 0 {
		t.Fatalf("a registered figure was flagged: %+v", got)
	}
}

func TestStatFieldPairing(t *testing.T) {
	cases := []struct {
		name        string
		content     map[string]interface{}
		wantLabel   string
		wantPairing StatPairing
	}{
		{"system-stats", map[string]interface{}{
			"stat1_value": "170", "stat1_label": "Agent Definitions",
		}, "Agent Definitions", StatPairedExact},

		{"case-studies-grid", map[string]interface{}{
			"card3_stat_value": "1,267", "card3_stat_label": "work items completed",
		}, "work items completed", StatPairedExact},

		{"word index", map[string]interface{}{
			"stat_one_value": "39", "stat_one_label": "Published Figures",
		}, "Published Figures", StatPairedExact},

		{"name vocabulary", map[string]interface{}{
			"spec_4_value": "24", "spec_4_name": "Operating Voltage",
		}, "Operating Voltage", StatPairedExact},

		{"anchor fallback", map[string]interface{}{
			"row2_spark_value": "12", "row2_feature": "Concurrent runs",
		}, "Concurrent runs", StatPairedAnchor},

		// Refusals — never guess.
		// Two candidates under the same anchor. Note a sibling ending in a
		// DETAIL-role token (row2_note) would NOT create ambiguity: role tokens
		// are excluded from label candidates by design, so the pairing would
		// still resolve. Ambiguity needs two non-role strings.
		{"ambiguous anchor", map[string]interface{}{
			"row2_spark_value": "12", "row2_feature": "Concurrent runs", "row2_platform1": "Competitor A",
		}, "", StatUnpaired},

		{"no siblings", map[string]interface{}{
			"hero_value": "42",
		}, "", StatUnpaired},

		{"only numeric sibling", map[string]interface{}{
			"price_2_value": "42", "price_2_other": "99",
		}, "", StatUnpaired},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claims := ExtractStatClaims("c", tc.content)
			if len(claims) != 1 {
				t.Fatalf("expected exactly 1 claim, got %d: %+v", len(claims), claims)
			}
			if claims[0].Label != tc.wantLabel {
				t.Errorf("label = %q, want %q", claims[0].Label, tc.wantLabel)
			}
			if claims[0].Pairing != tc.wantPairing {
				t.Errorf("pairing = %q, want %q", claims[0].Pairing, tc.wantPairing)
			}
		})
	}
}

// TestStatClaimExclusions: values that are not quantitative business claims
// must never reach the scan.
func TestStatClaimExclusions(t *testing.T) {
	eb := mustParseStatsEB(t)

	for _, v := range []string{"", "—", "–", "-", "n/a", "Tracked & Published", "Vendor-Neutral", "Yes"} {
		claims := ExtractStatClaims("c", map[string]interface{}{
			"stat_1_value": v, "stat_1_label": "Something",
		})
		if len(claims) != 0 {
			t.Errorf("value %q produced a claim: %+v", v, claims)
		}
	}

	// Digit-bearing but structurally excluded by the shared claims engine.
	for _, v := range []string{"2019", "£29", "2026-07-20", "v1.0.1124"} {
		claims := ExtractStatClaims("c", map[string]interface{}{
			"stat_1_value": v, "stat_1_label": "Something",
		})
		if got := eb.ScanStatClaims(claims); len(got) != 0 {
			t.Errorf("value %q was flagged as a business figure: %+v", v, got)
		}
	}
}

// TestStatUnitLint is bugs_open/043's own table — the suffix findings it
// recorded live, plus the false-positive guards.
func TestStatUnitLint(t *testing.T) {
	cases := []struct {
		value, suffix, label string
		want                 string // "" = no finding
	}{
		// 043's live instances: a magnitude marker given a dimension.
		{"2,400+", "%", "Gripper Models Indexed", "stat_unit_impossible"},
		{"140+", "ms", "Manufacturers Covered", "stat_unit_impossible"},
		{"1,000s", "ms", "Concurrent Instances", "stat_unit_impossible"},

		// A dimension with nothing in the label licensing it.
		{"14,203", "%", "Takes Filed Today", "stat_unit_mismatch"},
		{"39", "x", "Published Figures Held", "stat_unit_mismatch"},
		{"10", "%", "Gripper Models Indexed", "stat_unit_mismatch"},

		// False-positive guards — these must stay silent.
		{"36.6", "%", "PRD Accuracy Gap", ""},    // "gap" licenses a rate
		{"8", "ms", "Avg. Response Time", ""},    // "time"/"response"
		{"500", "ms", "Median Latency", ""},      // "latency"
		{"2", "x", "Throughput Improvement", ""}, // "throughput"/"improvement"
		{"150", "+", "Clients Served", ""},       // "+" is a magnitude marker
		{"2.4", "M", "Records Processed", ""},    // "M" is a magnitude marker
		{"170", "", "Agent Definitions", ""},     // no suffix at all
		{"99.99", "%", "Uptime", ""},             // "uptime"
	}

	for _, tc := range cases {
		name := tc.value + "|" + tc.suffix + "|" + tc.label
		t.Run(name, func(t *testing.T) {
			claims := ExtractStatClaims("system-stats", map[string]interface{}{
				"stat1_value": tc.value, "stat1_suffix": tc.suffix, "stat1_label": tc.label,
			})
			findings := LintStatUnits(claims)
			if tc.want == "" {
				if len(findings) != 0 {
					t.Fatalf("expected no finding, got %+v", findings)
				}
				return
			}
			if len(findings) != 1 {
				t.Fatalf("expected 1 %s finding, got %d: %+v", tc.want, len(findings), findings)
			}
			if findings[0].Check != tc.want {
				t.Errorf("check = %q, want %q", findings[0].Check, tc.want)
			}
			if !strings.Contains(findings[0].Matched, tc.value) {
				t.Errorf("matched = %q, expected it to carry the rendered value", findings[0].Matched)
			}
		})
	}
}

// TestLintStatUnitsNeedsNoEvidenceBase pins the opt-in split as a type-level
// property: the unit lint compares a component against itself, so it is a free
// function and can run fleet-wide, while the numeric audit is a method on
// *EvidenceBase and cannot run without a register.
func TestLintStatUnitsNeedsNoEvidenceBase(t *testing.T) {
	claims := ExtractStatClaims("system-stats", map[string]interface{}{
		"stat1_value": "2,400+", "stat1_suffix": "%", "stat1_label": "Anything At All",
	})
	if got := LintStatUnits(claims); len(got) != 1 || got[0].Check != "stat_unit_impossible" {
		t.Fatalf("expected one stat_unit_impossible with no evidence base, got %+v", got)
	}

	// The numeric audit, by contrast, is inert without a register.
	var nilEB *EvidenceBase
	if got := nilEB.ScanStatClaims(claims); got != nil {
		t.Fatalf("ScanStatClaims on a nil evidence base returned %+v, want nil", got)
	}
}

// TestExtractStatClaimsIsDeterministic — findings must not reshuffle between
// runs, or a diff of two validation logs is unreadable.
func TestExtractStatClaimsIsDeterministic(t *testing.T) {
	content := map[string]interface{}{
		"stat1_value": "1", "stat1_label": "One",
		"stat2_value": "2", "stat2_label": "Two",
		"stat3_value": "3", "stat3_label": "Three",
		"stat4_value": "4", "stat4_label": "Four",
	}
	first := ExtractStatClaims("system-stats", content)
	for i := 0; i < 20; i++ {
		got := ExtractStatClaims("system-stats", content)
		if len(got) != len(first) {
			t.Fatalf("claim count varied: %d vs %d", len(got), len(first))
		}
		for j := range got {
			if got[j].ValueKey != first[j].ValueKey {
				t.Fatalf("order varied at %d: %q vs %q", j, got[j].ValueKey, first[j].ValueKey)
			}
		}
	}
}

// TestDisplayOrdinalsAreNotQuantities — 01/02/03 as step markers.
//
// Both cases below were found on live pages by the fleet sweep that built
// bugs_open/093's second call site, and both were being reported as published
// figures with no supporting fact. On a site that HAS registered facts that is
// `error` severity in the build gate, i.e. a page made unbuildable by its own
// step numbering — bugs_closed/073's shape on a new trigger.
func TestDisplayOrdinalsAreNotQuantities(t *testing.T) {
	got := ExtractStatClaims("process-steps", map[string]interface{}{
		"step1_number": "01", "step1_label": "Build Arguments, Not Answers",
		"step2_number": "02", "step2_label": "Know the Archetypes at Your Table",
		// The control: a real count in the same component must survive, and
		// so must a bare zero — zero is a count, and an honest one.
		"stat1_value": "8", "stat1_label": "Archetypes",
		"stat2_value": "0", "stat2_label": "Open Defects",
	})

	seen := map[string]bool{}
	for _, c := range got {
		seen[c.Value] = true
	}
	for _, ordinal := range []string{"01", "02"} {
		if seen[ordinal] {
			t.Errorf("display ordinal %q was extracted as a quantitative claim", ordinal)
		}
	}
	for _, real := range []string{"8", "0"} {
		if !seen[real] {
			t.Errorf("real count %q was dropped along with the ordinals; claims: %+v", real, got)
		}
	}
}

// TestTypographicRangeIsExcludedLikeAHyphenRange — "8–12 minutes".
//
// unitSuffixRe already spells `[-–]`, so typographic dashes were always meant
// to be in scope; the adjacency test beside it was byte-level and an en-dash
// is three bytes, so only the hyphen form was ever excluded. Found live on
// fundamentallyai's llm-cost-calculator ("Read time: 8–12 minutes") on a site
// with 15 registered facts.
func TestTypographicRangeIsExcludedLikeAHyphenRange(t *testing.T) {
	eb := &EvidenceBase{} // empty register: anything examined WILL be flagged
	for _, value := range []string{"8-12 minutes", "8–12 minutes", "8—12 minutes"} {
		claims := ExtractStatClaims("article-meta", map[string]interface{}{
			"read_time_value": value, "read_time_label": "Read time",
		})
		if len(claims) != 1 {
			t.Fatalf("%q: expected one extracted claim, got %+v", value, claims)
		}
		if f := eb.ScanStatClaims(claims); len(f) != 0 {
			t.Errorf("%q was reported as an unregistered business figure: %+v", value, f)
		}
	}

	// The control: a plain unsupported figure must still be reported, or this
	// exclusion has quietly switched the scan off rather than narrowed it.
	claims := ExtractStatClaims("stat-grid", map[string]interface{}{
		"stat1_value": "1,267", "stat1_label": "Work Items Completed",
	})
	if f := eb.ScanStatClaims(claims); len(f) != 1 {
		t.Errorf("the negative control stopped firing — exclusion is too wide: %+v", f)
	}
}
