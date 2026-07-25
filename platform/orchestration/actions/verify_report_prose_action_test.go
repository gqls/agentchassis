// FILE: platform/orchestration/actions/verify_report_prose_action_test.go
//
// The gate's job is to make fabrication a hard failure. Every test here is a
// verify-the-failing-branch test: the interesting assertions are the ones
// where the gate REJECTS.

package actions

import (
	"encoding/json"
	"strings"
	"testing"
)

// realScoring produces a genuine score_grippers output over the live-mirror
// fixture index, so prose tests bind against exactly what production will.
func realScoring(t *testing.T, mass float64, ipMin int) map[string]interface{} {
	t.Helper()
	req := reqFixture(mass, 40, 12, 2, 0.15, "0.15", 2, ipMin, true)
	return scoreGrippers(req, testGripperIndex())
}

func proseWith(summary, candidates string) map[string]interface{} {
	return map[string]interface{}{
		"summary_html":          "<p>" + summary + "</p>",
		"candidates_html":       "<p>" + candidates + "</p>",
		"integration_html":      "<p>Mount per the manufacturer instructions and verify grip on first articles.</p>",
		"vendor_questions_html": "<ul><li>Confirm the published figures for your part finish.</li></ul>",
	}
}

func TestProseGatePassesFactBoundProse(t *testing.T) {
	// NB: ipMin must be 0 here — with an IP requirement this application is
	// a genuine zero-match scenario (every candidate fails or goes unknown
	// on unpublished IP), and the gate then rightly demands the no-match
	// sentence. The first version of this test tripped exactly that.
	scoring := realScoring(t, 2.5, 0)
	// 200.0 N (fJaw), 235 N, 1520 N and 85 mm all appear in the fact block.
	prose := proseWith(
		"The friction-grip requirement works out at 200.0 N.",
		"The Robotiq 2F-85 publishes 20 to 235 N of gripping force and 85 mm of travel.")
	if v := verifyReportProse(prose, scoring, nil); len(v) != 0 {
		t.Errorf("clean prose rejected: %v", v)
	}
}

func TestProseGateRejectsInventedNumber(t *testing.T) {
	scoring := realScoring(t, 2.5, 54)
	prose := proseWith(
		"The requirement is 200.0 N.",
		"The Robotiq 2F-85 delivers 999 N of clamping force.") // 999 invented
	v := verifyReportProse(prose, scoring, nil)
	if len(v) == 0 {
		t.Fatal("invented number 999 passed the gate")
	}
	if !strings.Contains(strings.Join(v, "|"), `"999"`) {
		t.Errorf("violation should name the token: %v", v)
	}
}

func TestProseGateRejectsInventedSKU(t *testing.T) {
	scoring := realScoring(t, 2.5, 0)
	// "2F-140" is a real Robotiq sibling model that is NOT in this index —
	// exactly the fabrication shape the gate exists for. Its digits (140)
	// appear in the fact block via the 2FG7, so only the SKU check can
	// catch it.
	prose := proseWith(
		"The requirement is 200.0 N.",
		"Consider the Robotiq 2F-140 as an alternative with 140 N of force.")
	v := verifyReportProse(prose, scoring, nil)
	if len(v) == 0 {
		t.Fatal("invented SKU 2F-140 passed the gate")
	}
	if !strings.Contains(strings.Join(v, "|"), "2F-140") {
		t.Errorf("violation should name the SKU: %v", v)
	}
}

func TestProseGateNoMatchContract(t *testing.T) {
	scoring := realScoring(t, 500, 67) // nothing fits
	if scoring["match_count"].(int) != 0 {
		t.Fatal("fixture should produce zero matches")
	}

	// (a) Summary missing the mandatory sentence -> rejected.
	prose := proseWith("Unfortunately the options are limited.", "The index was assessed in full.")
	v := verifyReportProse(prose, scoring, nil)
	if !strings.Contains(strings.Join(v, "|"), "mandatory sentence") {
		t.Errorf("missing mandatory sentence not flagged: %v", v)
	}

	// (b) Sentence present but softened elsewhere -> rejected.
	prose = proseWith(
		noMatchSentence+" Consider widening the search.",
		"The Schmalz SGM-HP 50 nearly meets the requirement for your parts.")
	v = verifyReportProse(prose, scoring, nil)
	if !strings.Contains(strings.Join(v, "|"), "softens") {
		t.Errorf("softening language not flagged: %v", v)
	}

	// (c) Honest no-match prose -> passes.
	prose = proseWith(
		noMatchSentence+" The closest candidates are shown with the exact shortfalls.",
		"The Schmalz SGM-HP 50 publishes 385 N with the friction ring; the computed direct-hold requirement is beyond every published figure in this index.")
	if v := verifyReportProse(prose, scoring, nil); len(v) != 0 {
		t.Errorf("honest no-match prose rejected: %v", v)
	}
}

func TestProseGateRejectsEmptySection(t *testing.T) {
	scoring := realScoring(t, 2.5, 54)
	prose := proseWith("The requirement is 200.0 N.", "Assessment complete.")
	prose["integration_html"] = "<p>   </p>"
	v := verifyReportProse(prose, scoring, nil)
	if !strings.Contains(strings.Join(v, "|"), "integration_html is empty") {
		t.Errorf("empty section not flagged: %v", v)
	}
}

func TestProseGateFormattingToleranceAndContext(t *testing.T) {
	scoring := realScoring(t, 2.5, 0)
	// "1,520" must normalise to the fact block's 1520; the ISO mounting
	// string is request context the prose may echo even though scoring
	// never saw it.
	prose := proseWith(
		"The Zimmer Group unit publishes 1,520 N.",
		"Specify an ISO 9409-1-50-4-M6 flange when ordering.")
	v := verifyReportProse(prose, scoring, []string{"ISO 9409-1-50-4-M6"})
	if len(v) != 0 {
		t.Errorf("formatting/context tolerance failed: %v", v)
	}
	// Without the context, the mounting digits must be rejected — proving
	// the context path is what allowed them, not a hole in the gate.
	if v := verifyReportProse(prose, scoring, nil); len(v) == 0 {
		t.Error("mounting digits passed without context — the numeric gate has a hole")
	}
}

// --- truncation guard --------------------------------------------------------

// The marker is a SIBLING of the prose path, derived rather than configured so
// it cannot drift when prose_field changes. A wrong derivation here fails
// silently — it simply never finds a marker — so the mapping is asserted
// directly rather than only through the action.
func TestTruncationMarkerFieldDerivation(t *testing.T) {
	cases := []struct{ prose, override, want string }{
		{"report_prose.result", "", "report_prose.__truncated"},
		{"report_prose.result", "custom.path", "custom.path"},
		{"report_prose.result", "  ", "report_prose.__truncated"}, // blank is not an override
		{"flat", "", ""}, // no parent segment: no marker to find
	}
	for _, c := range cases {
		if got := truncationMarkerField(c.prose, c.override); got != c.want {
			t.Errorf("truncationMarkerField(%q,%q) = %q, want %q", c.prose, c.override, got, c.want)
		}
	}
}

// --- chart determinism -------------------------------------------------------

func TestHeadroomChartDeterministicAndHonest(t *testing.T) {
	req := reqFixture(2.5, 15, 12, 2, 0.15, "0.15", 2, 0, false)
	result := scoreGrippers(req, testGripperIndex())
	var cands []assessment
	remarshal(t, result["candidates"], &cands)

	a := renderHeadroomChart(cands)
	b := renderHeadroomChart(cands)
	if a != b {
		t.Fatal("chart SVG is not byte-stable across runs")
	}
	if !strings.Contains(a, "Zimmer Group") || !strings.Contains(a, "marginal threshold") {
		t.Error("chart missing candidate labels or reference captions")
	}
	if strings.Contains(a, "<script") {
		t.Error("chart must be script-free")
	}
	// A candidate with no comparable capacity figure must not get a bar.
	zeroHeadroom := []assessment{{Name: "Ghost", Verdict: "Insufficient data", Headroom: 0}}
	if svg := renderHeadroomChart(zeroHeadroom); svg != "" {
		t.Error("chart drew a bar for a candidate with no capacity figure")
	}
}

// --- report section renderer smoke ------------------------------------------

func TestRenderReportSection(t *testing.T) {
	scoring := realScoring(t, 2.5, 54)
	prose := proseWith(
		"The friction-grip requirement works out at 200.0 N.",
		"The Robotiq 2F-85 publishes 20 to 235 N.")

	html, err := renderReportSection(scoring, prose)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"(2.5 × 12 × 2) ÷ (0.15 × 2)",       // the substituted formula literal
		"Not published by manufacturer",     // unpublished figures stay visible
		"manufacturer specification",        // provenance anchors
		"data-component=\"report-dossier\"", // component contract
		"<svg",                              // the chart made it in
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered report missing %q", want)
		}
	}
	if strings.Contains(html, "<script") {
		t.Error("report section must be script-free")
	}

	// Empty prose at render time is a hard error, never a blank section.
	prose["integration_html"] = ""
	if _, err := renderReportSection(scoring, prose); err == nil {
		t.Error("empty prose section rendered without error")
	}
}

func remarshal(t *testing.T, in, out interface{}) {
	t.Helper()
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatal(err)
	}
}
