// FILE: platform/orchestration/actions/discovery_checks/check_unverified_claims_stats_test.go
//
// Guards for bugs_open/093's second call site. The scan engines themselves are
// tested in datahelpers; what is tested HERE is the part this file decides and
// the engines cannot: WHEN each scan runs, and how a finding is graded.
//
// The first test is the one that matters. It pins the asymmetry that made the
// claims layer a silent no-op on three sites for two days: a site_specs
// evidence_base row carrying only a writer_block parses to a nil EvidenceBase,
// and every consumer that treats nil as "not opted in" therefore switches
// itself OFF on a site that has explicitly opted IN. A green writer says
// nothing about a gate — each half has to be checked against its own consumer.

package discovery_checks

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func findingsByCheck(fs []unverifiedClaimFinding) map[string]unverifiedClaimFinding {
	out := make(map[string]unverifiedClaimFinding, len(fs))
	for _, f := range fs {
		out[f.Check] = f
	}
	return out
}

// A writer_block-only evidence_base row: opted in, nothing machine-readable.
// datahelpers.ParseEvidenceBase returns (nil, nil) for exactly this shape.
func TestWriterBlockOnlyRowStillAuditsStats(t *testing.T) {
	eb, err := datahelpers.ParseEvidenceBase([]byte(`{"audit_doc":"x","writer_block":"NUMBERS: ..."}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if eb != nil {
		t.Fatalf("premise of this test has changed: ParseEvidenceBase no longer returns nil for a "+
			"writer_block-only row (got %+v). Re-check every consumer that keys on nil.", eb)
	}

	content := map[string]interface{}{
		"stat1_value": "2,400",
		"stat1_label": "Gripper Models Indexed",
	}

	// rowExists=true is the whole point: nil eb must NOT mean "skip".
	got := scanStoredStatClaims("stat-grid", "main", content, eb, true)
	if len(got) != 1 {
		t.Fatalf("expected the figure to be reported on an opted-in site with no facts; got %d findings: %+v", len(got), got)
	}
	if got[0].Check != "unregistered_stat" {
		t.Errorf("check = %q, want unregistered_stat", got[0].Check)
	}
	// Nothing could actually be verified, so the grade must say so rather than
	// dressing an absence of evidence as evidence of absence.
	if got[0].Severity != "low" {
		t.Errorf("severity = %q, want low (no facts[] means nothing was checked)", got[0].Severity)
	}
	if !strings.Contains(got[0].Reason, "no machine-readable facts[]") {
		t.Errorf("reason does not name the gap that makes this unverifiable: %q", got[0].Reason)
	}

	// And the control: with the row absent, the register scan must stay off —
	// flagging every figure on a site that never opted in is noise.
	if off := scanStoredStatClaims("stat-grid", "main", content, nil, false); len(off) != 0 {
		t.Errorf("register scan ran on a site with no evidence_base row: %+v", off)
	}
}

// The unit lint compares a component against itself, so it must run whether or
// not the site has a register — the same scope it has in the build gate. Two
// sites with persisted stat fields have no evidence_base row at all.
func TestUnitLintRunsWithoutAnyRegister(t *testing.T) {
	content := map[string]interface{}{
		"stat1_value":  "2,400+", // magnitude marker...
		"stat1_suffix": "%",      // ...then a dimension: cannot mean anything
		"stat1_label":  "Gripper Models Indexed",
	}

	got := scanStoredStatClaims("system-stats", "main", content, nil, false)
	byCheck := findingsByCheck(got)

	f, ok := byCheck["stat_unit_impossible"]
	if !ok {
		t.Fatalf("the unit lint did not run without a register; got %+v", got)
	}
	if f.Severity != "high" {
		t.Errorf("severity = %q, want high — %q is a broken figure live on the site", f.Severity, f.Matched)
	}
	if f.Source != "content_data" {
		t.Errorf("source = %q, want content_data — a human fixing only the rendered HTML would have it reprinted by the next re-render", f.Source)
	}
	if _, leaked := byCheck["unregistered_stat"]; leaked {
		t.Errorf("register scan ran with no evidence_base row: %+v", got)
	}
}

// A figure a fact supports is not a finding, and one it does not is medium
// once the site has actually registered countables.
func TestRegisteredFactsGradeStatsAgainstTheRegister(t *testing.T) {
	eb, err := datahelpers.ParseEvidenceBase([]byte(
		`{"audit_doc":"x","facts":[{"claim":"gripper models indexed","value":170,"tolerance":"exact"}]}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if eb == nil || len(eb.Facts) != 1 {
		t.Fatalf("fixture did not parse into a one-fact register: %+v", eb)
	}

	supported := scanStoredStatClaims("stat-grid", "main", map[string]interface{}{
		"stat1_value": "170",
		"stat1_label": "Gripper Models Indexed",
	}, eb, true)
	if len(supported) != 0 {
		t.Errorf("a registered figure was reported: %+v", supported)
	}

	unsupported := scanStoredStatClaims("stat-grid", "main", map[string]interface{}{
		"stat1_value": "2,400",
		"stat1_label": "Gripper Models Indexed",
	}, eb, true)
	if len(unsupported) != 1 {
		t.Fatalf("expected the unsupported figure to be reported; got %+v", unsupported)
	}
	if unsupported[0].Severity != "medium" {
		t.Errorf("severity = %q, want medium — a machine checked this one and it failed", unsupported[0].Severity)
	}
}

// claimsSeverity/claimsPriority now read the per-finding grade. The HTML checks
// must keep the grades they had before that change.
func TestSurfaceSeverityIsTheWorstFinding(t *testing.T) {
	cases := []struct {
		name     string
		findings []unverifiedClaimFinding
		severity string
		priority int
	}{
		{"banned claim still high", []unverifiedClaimFinding{
			{Check: "banned_claim", Severity: "high"},
			{Check: "unregistered_stat", Severity: "low"},
		}, "high", 25},
		{"unregistered number still medium", []unverifiedClaimFinding{
			{Check: "unregistered_number", Severity: "medium"},
		}, "medium", 35},
		{"unverifiable stats alone are low", []unverifiedClaimFinding{
			{Check: "unregistered_stat", Severity: "low"},
			{Check: "stat_unit_mismatch", Severity: "low"},
		}, "low", 45},
		{"an ungraded finding is never silently demoted", []unverifiedClaimFinding{
			{Check: "something_new"},
		}, "medium", 35},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := claimsSeverity(tc.findings); got != tc.severity {
				t.Errorf("severity = %q, want %q", got, tc.severity)
			}
			if got := claimsPriority(tc.findings); got != tc.priority {
				t.Errorf("priority = %d, want %d", got, tc.priority)
			}
		})
	}
}

// A summary that reads "0 banned claim(s), 0 unregistered number(s), 1 …" for
// every item is what makes a review queue unreadable.
func TestSummaryNamesOnlyTheCategoriesThatFired(t *testing.T) {
	got := claimsSummary("about", []unverifiedClaimFinding{
		{Check: "stat_unit_impossible", Severity: "high"},
		{Check: "unregistered_stat", Severity: "low"},
	})
	want := "Unverified claims on about: 1 unregistered stat field(s), 1 suspect stat unit(s)"
	if got != want {
		t.Errorf("summary = %q, want %q", got, want)
	}
}
