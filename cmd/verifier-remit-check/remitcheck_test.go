// Tests for the class detector.
//
// THE FIXTURE IS THE LIVE CENSUS, not an invention. Every key-set below is
// verbatim from clients_db on 2026-08-11 (the query is in
// RUNBOOK_verifier_producer_join.md, §"Find every verified item_type with more
// than one producer"). That matters: a clustering rule tested only against shapes
// chosen to exercise it will always look right, and this repo's WRONG_CALLS.md
// carries the entry — a fixture you COMPOSE to exercise a rule will exercise it.
// The one case that constrains the threshold from below (page_canonical_collision,
// 2 shared keys of 3) is real data, and it is why Jaccard is not used.
package main

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// registeredOn20260811 is what RegisteredVerifierItemTypes() returned that day.
var registeredOn20260811 = []string{
	"content_duplication", "dead_fragment_link", "decision_regression", "empty_section",
	"hardcoded_section_colors", "literal_markdown", "missing_conversion_path",
	"orphan_element_refs", "page_canonical_collision", "revenue_shape_cta",
	"truncated_component", "unbuilt_internal_link",
}

const designAuditShape = "acceptance_test,affected_component,audit_source,category,current_value," +
	"description,max_fix_attempts,original_domain,original_pipeline,page_name,suggestion"

// liveCensus20260811 — one entry per (item_type, audit_source, spec key-set).
func liveCensus20260811() []censusRow {
	return []censusRow{
		// dark_section_audit has NO verifier — present to prove an unregistered
		// type is not assessed at all, however many rows it has.
		{ItemType: "dark_section_audit", Label: "design-audit", Keyset: designAuditShape, Rows: 14},

		{ItemType: "decision_regression", Keyset: "assert,check,decision,page,pattern", Rows: 2},

		// One producer, two spec revisions: original_pipeline came and went.
		{ItemType: "empty_section", Keyset: "check,component_function,component_id,empty_pattern,original_pipeline,page_id,page_name,slot_name", Rows: 44},
		{ItemType: "empty_section", Keyset: "check,component_function,component_id,empty_pattern,page_id,page_name,slot_name", Rows: 20},

		// THE MOTIVATING CASE. Producer A's four revisions all contain
		// {check, components_found}; producer B shares NO key with any of them.
		{ItemType: "hardcoded_section_colors", Label: "design-audit", Keyset: designAuditShape, Rows: 8},
		{ItemType: "hardcoded_section_colors", Keyset: "check,components_found", Rows: 4},
		{ItemType: "hardcoded_section_colors", Keyset: "check,components_found,original_pipeline,out_of_remit,population", Rows: 3},
		{ItemType: "hardcoded_section_colors", Keyset: "check,components_found,out_of_remit,population", Rows: 1},
		{ItemType: "hardcoded_section_colors", Keyset: "check,components_found,original_pipeline", Rows: 1},

		{ItemType: "literal_markdown", Keyset: "check,findings,fix,original_pipeline,page_id,page_name,page_url", Rows: 32},
		{ItemType: "literal_markdown", Keyset: "check,findings,fix,page_id,page_name,page_url", Rows: 30},

		{ItemType: "missing_conversion_path", Keyset: "adopted_branch,check,missing,original_pipeline,primary_model,rule", Rows: 1},

		// The narrowest same-producer overlap in the whole fleet: 2 shared keys
		// of 3 (0.667). Under Jaccard this pair scores 0.167 and would be called
		// a second producer — which is why the coefficient is overlap, not Jaccard.
		{ItemType: "page_canonical_collision", Keyset: "active_count,bug,canonical_names,check,convention,decided_by,decision,group_key,members,path_keys,resolution_evidence", Rows: 2},
		{ItemType: "page_canonical_collision", Keyset: "canonical_names,path_keys,probe", Rows: 1},

		{ItemType: "truncated_component", Keyset: "check,component,component_id,component_level,fix,function,intact_version_available,unterminated", Rows: 2},
		{ItemType: "truncated_component", Keyset: "check,component,component_id,component_level,fix,function,intact_version_available,intact_version_number,unterminated", Rows: 1},

		{ItemType: "unbuilt_internal_link", Keyset: "check,fix,href,issue_type,occurrences,original_pipeline,page_id,page_name,slot_name,surface,target_page_id", Rows: 49},
	}
}

func noRemit(string) bool  { return false }
func allRemit(string) bool { return true }

// TestProducerFamiliesOnTheLiveCensus is the disconfirmability test: the predicate
// must return 2 where the shape genuinely exists and 1 everywhere else. If it
// could not return 2, the 1s would be an artefact of the predicate rather than a
// finding about the fleet (016b §9).
func TestProducerFamiliesOnTheLiveCensus(t *testing.T) {
	want := map[string]int{
		"hardcoded_section_colors": 2, // the only true two-producer type
		"empty_section":            1,
		"literal_markdown":         1,
		"page_canonical_collision": 1,
		"truncated_component":      1,
		"unbuilt_internal_link":    1,
		"decision_regression":      1,
		"missing_conversion_path":  1,
		"dark_section_audit":       1,
	}
	byType := map[string][]censusRow{}
	for _, r := range liveCensus20260811() {
		byType[r.ItemType] = append(byType[r.ItemType], r)
	}
	for itemType, wantN := range want {
		families, shapeless := producerFamilies(byType[itemType])
		if len(families) != wantN {
			t.Errorf("%s: got %d producer families, want %d — shapes: %v",
				itemType, len(families), wantN, byType[itemType])
		}
		if shapeless != 0 {
			t.Errorf("%s: %d shapeless rows in a fixture that has none", itemType, shapeless)
		}
	}
}

// TestFamilyRowsAreNotLost — a census that quietly drops rows is not a census.
func TestFamilyRowsAreNotLost(t *testing.T) {
	var rows []censusRow
	for _, r := range liveCensus20260811() {
		if r.ItemType == "hardcoded_section_colors" {
			rows = append(rows, r)
		}
	}
	families, shapeless := producerFamilies(rows)
	total := shapeless
	for _, f := range families {
		total += f.Rows
	}
	if total != 17 {
		t.Errorf("row count through the clustering = %d, want 17 (8 producer B + 9 producer A)", total)
	}
	// And the split must be the real one, not any 2-way split summing to 17.
	got := map[int]bool{}
	for _, f := range families {
		got[f.Rows] = true
	}
	if !got[8] || !got[9] {
		t.Errorf("families split %v, want one of 8 rows (design-audit) and one of 9 (the check's own)", got)
	}
}

// TestShapelessRowsAreReportedNotDropped — an empty spec carries no shape
// evidence, so it must not create a phantom producer, and must not vanish either.
func TestShapelessRowsAreReportedNotDropped(t *testing.T) {
	rows := []censusRow{
		{ItemType: "empty_section", Keyset: "check,page_id", Rows: 5},
		{ItemType: "empty_section", Keyset: "", Rows: 3},
	}
	families, shapeless := producerFamilies(rows)
	if len(families) != 1 {
		t.Errorf("got %d families, want 1 — an empty spec must not be a second producer", len(families))
	}
	if shapeless != 3 {
		t.Errorf("shapeless = %d, want 3", shapeless)
	}
}

// TestAuditSourceSplitsEvenWhenShapesMatch — the label axis earns its place: two
// producers writing the same keys under different audit_source labels are two
// producers. (The key-set axis alone would call them one.)
func TestAuditSourceSplitsEvenWhenShapesMatch(t *testing.T) {
	rows := []censusRow{
		{ItemType: "x", Label: "design-audit", Keyset: "a,b,c", Rows: 4},
		{ItemType: "x", Label: "content-audit", Keyset: "a,b,c", Rows: 4},
	}
	if families, _ := producerFamilies(rows); len(families) != 2 {
		t.Errorf("got %d families, want 2 — different audit_source labels are different producers", len(families))
	}
}

// TestDisjointShapesUnderOneLabelAreTwoProducers — the half the live census
// CANNOT pin, stated as such.
//
// Every multi-shape type in today's fleet either has one producer, or has a
// second one that labels itself with audit_source. So the live-census test above
// would stay green even if the key-set clustering were removed entirely: the
// label axis alone reproduces its answers. This fixture is therefore CONSTRUCTED,
// not observed — and it is the case the whole design turns on, because
// audit_source is optional and a converging producer need not write it. If this
// goes red, the check has been reduced to a distinct-audit_source count and would
// miss the very producer it was built for.
func TestDisjointShapesUnderOneLabelAreTwoProducers(t *testing.T) {
	rows := []censusRow{
		{ItemType: "x", Keyset: "check,components_found", Rows: 9},
		{ItemType: "x", Keyset: "finding,page_name,suggestion", Rows: 4},
	}
	if families, _ := producerFamilies(rows); len(families) != 2 {
		t.Errorf("got %d families, want 2 — two shapes sharing NO key are two producers "+
			"even when neither writes an audit_source", len(families))
	}
}

// TestFindingNeedsBothHalves is the mutation matrix in test form. This lane
// learned (NOTES 2026-08-10) that reverting one half of a fix can leave a guard
// green, so each half is flipped independently here rather than together.
func TestFindingNeedsBothHalves(t *testing.T) {
	census := liveCensus20260811()

	cases := []struct {
		name        string
		declares    func(string) bool
		wantFinding bool
	}{
		{"no remit declared anywhere → the two-producer type is a finding", noRemit, true},
		{"remit declared (today's live state) → answered, no finding", allRemit, false},
	}
	for _, tc := range cases {
		var findings []string
		for _, a := range assess(census, registeredOn20260811, tc.declares) {
			if a.Finding() {
				findings = append(findings, a.ItemType)
			}
		}
		if tc.wantFinding {
			if len(findings) != 1 || findings[0] != "hardcoded_section_colors" {
				t.Errorf("%s: findings %v, want exactly [hardcoded_section_colors]", tc.name, findings)
			}
		} else if len(findings) != 0 {
			t.Errorf("%s: findings %v, want none", tc.name, findings)
		}
	}

	// The other half: a SINGLE-producer type is never a finding, remit or not —
	// so the finding is not just "no remit declared".
	for _, declares := range []func(string) bool{noRemit, allRemit} {
		for _, a := range assess(census, registeredOn20260811, declares) {
			if a.ItemType != "hardcoded_section_colors" && a.Finding() {
				t.Errorf("%s became a finding with one producer family", a.ItemType)
			}
		}
	}
}

// TestUnregisteredTypesAreNotAssessed — dark_section_audit has 14 rows and no
// verifier. A type with no verifier has no wrong predicate to be graded by; that
// is a different gap with its own guard (verifier_coverage_test.go).
func TestUnregisteredTypesAreNotAssessed(t *testing.T) {
	for _, a := range assess(liveCensus20260811(), registeredOn20260811, noRemit) {
		if a.ItemType == "dark_section_audit" {
			t.Fatal("dark_section_audit was assessed, but it has no registered verifier")
		}
	}
}

// TestEveryRegisteredTypeIsAssessedEvenWithNoRows — "we looked and there was
// nothing to look at" must be visible as a row in the report, not as silence.
func TestEveryRegisteredTypeIsAssessedEvenWithNoRows(t *testing.T) {
	got := assess(liveCensus20260811(), registeredOn20260811, noRemit)
	if len(got) != len(registeredOn20260811) {
		t.Fatalf("assessed %d types, want all %d registered", len(got), len(registeredOn20260811))
	}
	for _, a := range got {
		if a.ItemType == "content_duplication" && a.Rows != 0 {
			t.Errorf("content_duplication has no rows in the census but reports %d", a.Rows)
		}
	}
}

// TestSuppressedIsReported — the positive control. A run that finds nothing must
// still be able to show that its census SAW the shape it exists to catch.
func TestSuppressedIsReported(t *testing.T) {
	report := render(assess(liveCensus20260811(), registeredOn20260811, allRemit), runOutcome{DryRun: true})
	if !strings.Contains(report, "hardcoded_section_colors") {
		t.Error("a clean report does not name the multi-producer type its remit answered — " +
			"a bare '0 findings' cannot be told apart from a check that looked at nothing")
	}
	if !strings.Contains(report, "Findings: **0**") {
		t.Errorf("expected a zero-finding report, got:\n%s", report)
	}
}

func TestOverlap(t *testing.T) {
	cases := []struct {
		a, b []string
		want float64
	}{
		{[]string{"check", "components_found"}, []string{"check", "components_found", "population"}, 1},
		{[]string{"check", "components_found"}, strings.Split(designAuditShape, ","), 0},
		{[]string{"canonical_names", "path_keys", "probe"}, []string{"active_count", "canonical_names", "path_keys"}, 2.0 / 3.0},
		{nil, []string{"a"}, 0},
		{[]string{"a"}, nil, 0},
	}
	for _, c := range cases {
		if got := overlap(c.a, c.b); got != c.want {
			t.Errorf("overlap(%v,%v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

// TestCensusQueryRefusesAnUnsafeItemType — the query is built by concatenation
// because both fetch routes must run the identical SQL text. That is only
// acceptable if the identifiers are proven, not assumed.
func TestCensusQueryRefusesAnUnsafeItemType(t *testing.T) {
	if _, err := censusQuery([]string{"empty_section", "x'; DROP TABLE site_work_items; --"}); err == nil {
		t.Fatal("censusQuery accepted an item_type that is not [a-z0-9_]+")
	}
	if _, err := censusQuery(nil); err == nil {
		t.Fatal("censusQuery accepted an empty type list")
	}
	q, err := censusQuery(registeredOn20260811)
	if err != nil {
		t.Fatalf("censusQuery on the live registry: %v", err)
	}
	for _, want := range []string{"'hardcoded_section_colors'", "audit_source", "jsonb_object_keys"} {
		if !strings.Contains(q, want) {
			t.Errorf("census query is missing %q:\n%s", want, q)
		}
	}
}

// TestStillAFindingDistinguishesAnsweredFromUnevaluated — the retraction rule.
// An open item whose subject type this run never evaluated must be left alone.
func TestStillAFindingDistinguishesAnsweredFromUnevaluated(t *testing.T) {
	assessments := assess(liveCensus20260811(), registeredOn20260811, noRemit)
	if !stillAFinding(assessments, itemKeyFor("hardcoded_section_colors")) {
		t.Error("the live finding was not recognised as still holding")
	}
	if stillAFinding(assessments, itemKeyFor("a_verifier_that_was_deleted")) {
		t.Error("an unevaluated subject type was treated as a finding — it must be left open and reported instead")
	}
}

// TestAccessorAgreesWithTheLiveRegistry pins the two halves of the question
// against the compiled-in registry, so this binary's premise cannot rot silently
// if a registration moves. It is also the check's own positive control in code:
// at least one type MUST declare a remit, or the suppression branch is untested
// against reality.
func TestAccessorAgreesWithTheLiveRegistry(t *testing.T) {
	if !discovery_checks.VerifierDeclaresRemit("hardcoded_section_colors") {
		t.Error("hardcoded_section_colors no longer declares a remit — WII-013 shipped one; " +
			"if it was deliberately removed, this detector will now file a finding on it (which is correct)")
	}
	if discovery_checks.VerifierDeclaresRemit("empty_section") {
		t.Error("empty_section declares a remit — the fixture and the register entry both say it does not")
	}
	if discovery_checks.VerifierDeclaresRemit("dark_section_audit") {
		t.Error("dark_section_audit has no verifier at all, so it cannot declare a remit")
	}
	declaring := 0
	for _, itemType := range discovery_checks.RegisteredVerifierItemTypes() {
		if discovery_checks.VerifierDeclaresRemit(itemType) {
			declaring++
		}
	}
	if declaring == 0 {
		t.Error("no registered verifier declares a remit — the suppression branch cannot be exercised")
	}
}
