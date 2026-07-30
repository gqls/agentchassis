// FILE: platform/orchestration/actions/save_sections_claims_guard_test.go
//
// Contract for the CLAIMS FLOOR at the persistence seam (bugs_open/149 C1).
//
// What is pinned here is the FLOOR's own behaviour: which severity refuses a
// save, which severity only records, that the scan is per-section, and that the
// levers work. The scan SEMANTICS are not retested — datahelpers/claims_test.go
// and validate_page_content_claims_test.go own those, and duplicating them here
// would create the second implementation this file exists to avoid.
//
// Every assertion below checks that the mechanism FIRED, not merely that the run
// was quiet. A test that passes when the rule has been deleted is worse than no
// test: it reports coverage it does not have.

package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// claimsFloorTestEB carries one banned pattern and one registered fact, so the
// two severities can be provoked independently on the same register.
const claimsFloorTestEB = `{
  "audit_doc": "test",
  "facts": [
    {"id": "F1", "claim": "2767 records verified", "value": 2767,
     "kind": "metric", "source": {"sql": "SELECT 1"}, "verified_at": "2026-07-16", "tolerance": "exact"}
  ],
  "banned_claims": [
    {"pattern": "eight (functional |business )?(departments|functions|areas)", "reason": "test U1"}
  ]
}`

func mustParseFloorEB(t *testing.T) *datahelpers.EvidenceBase {
	t.Helper()
	eb, err := datahelpers.ParseEvidenceBase([]byte(claimsFloorTestEB))
	if err != nil || eb == nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	return eb
}

// TestClaimsFloorSeverityRule is the load-bearing one: a banned claim must land
// in `blockers` (which refuses the save) and an unregistered number must land in
// `errors` (which records and allows). If these two ever merge, the floor either
// stops protecting or starts blocking builds on a check with a known
// false-positive rate.
func TestClaimsFloorSeverityRule(t *testing.T) {
	eb := mustParseFloorEB(t)

	sections := []SectionData{
		{ComponentName: "hero", HTML: `<p>We span eight departments and serve 45 clients.</p>`},
	}

	blockers, errs := scanSectionClaims(sections, eb, true, datahelpers.ClaimSurface{}, zap.NewNop())

	if len(blockers) != 1 {
		t.Fatalf("expected exactly 1 blocker (the banned claim), got %d: %+v", len(blockers), blockers)
	}
	if blockers[0].Issue.Severity != "blocker" {
		t.Errorf("banned claim must carry severity blocker, got %q", blockers[0].Issue.Severity)
	}
	if blockers[0].Section != "hero" {
		t.Errorf("finding must name the section that carries it, got %q", blockers[0].Section)
	}

	if len(errs) != 1 {
		t.Fatalf("expected exactly 1 error (the unregistered number), got %d: %+v", len(errs), errs)
	}
	if errs[0].Issue.Severity != "error" {
		t.Errorf("unregistered number must carry severity error, got %q", errs[0].Issue.Severity)
	}
	if errs[0].Issue.Value != "45" {
		t.Errorf("expected the unregistered 45 to flag, got %q", errs[0].Issue.Value)
	}
}

// TestClaimsFloorScansEverySectionAndAttributes proves the scan is per-section
// rather than over a joined document, and that each finding names its own
// section. A whole-page scan would still find both claims — and would report
// them against nothing an operator could fix.
func TestClaimsFloorScansEverySectionAndAttributes(t *testing.T) {
	eb := mustParseFloorEB(t)

	sections := []SectionData{
		{ComponentName: "hero", HTML: `<p>Straightforward copy with nothing to flag.</p>`},
		{ComponentName: "about", HTML: `<p>We span eight departments.</p>`},
		{ComponentName: "cta", HTML: `<p>We also span eight functions.</p>`},
	}

	blockers, _ := scanSectionClaims(sections, eb, true, datahelpers.ClaimSurface{}, zap.NewNop())

	if len(blockers) != 2 {
		t.Fatalf("expected a finding from each of the two offending sections, got %d: %+v",
			len(blockers), blockers)
	}
	got := map[string]bool{}
	for _, b := range blockers {
		got[b.Section] = true
	}
	if !got["about"] || !got["cta"] {
		t.Errorf("both offending sections must be named; got %v", got)
	}
	if got["hero"] {
		t.Errorf("the clean section must not be implicated")
	}
}

// TestClaimsFloorFleetWideProtectsAnUnarmedSite is the bugs_closed/104 property
// at the new seam: a site with NO register is still scanned for banned claims.
// Four of the six agents that persist through this seam have no gate at all, so
// if this ever regresses, an unarmed site's known falsehoods have nothing looking
// at them on any path.
//
// nil eb is the unarmed case — not an empty register, which is a different thing.
func TestClaimsFloorFleetWideProtectsAnUnarmedSite(t *testing.T) {
	// A fleet-wide banned pattern, asserted by a site with no evidence_base.
	// "independently verified" is one of the live set and is the exact string
	// found on two robot-hands.com components in the 2026-07-30 fleet scan.
	sections := []SectionData{
		{ComponentName: "how-it-works", HTML: `<p>Every specification is independently verified.</p>`},
	}

	blockers, errs := scanSectionClaims(sections, nil, true, datahelpers.ClaimSurface{}, zap.NewNop())

	if len(blockers) == 0 {
		t.Fatal("an unarmed site (nil evidence base) must still be scanned by the fleet-wide " +
			"banned-claim set — bugs_closed/104. Zero findings here means the floor is blind " +
			"on every site without a register, which is most of them.")
	}
	if errs != nil {
		t.Errorf("the numeric half is opt-in on the register and must not run without one, got %+v", errs)
	}
}

// TestClaimsFloorFleetWideLeverDisarms pins the withdrawal path. The lever is the
// containment the council's guardian seat required for the fleet-wide set: it
// must genuinely turn the set off, or the kill switch is decorative.
//
// Asserted in BOTH directions in one test, because "0 findings with the lever
// off" alone would also pass if the scan had stopped working entirely.
func TestClaimsFloorFleetWideLeverDisarms(t *testing.T) {
	sections := []SectionData{
		{ComponentName: "how-it-works", HTML: `<p>Every specification is independently verified.</p>`},
	}

	armed, _ := scanSectionClaims(sections, nil, true, datahelpers.ClaimSurface{}, zap.NewNop())
	if len(armed) == 0 {
		t.Fatal("precondition failed: the fleet-wide set must find this with the lever ON, " +
			"otherwise the OFF assertion below proves nothing")
	}

	disarmed, _ := scanSectionClaims(sections, nil, false, datahelpers.ClaimSurface{}, zap.NewNop())
	if len(disarmed) != 0 {
		t.Errorf("check_claims_fleet_wide:false must restore the pre-104 behaviour exactly "+
			"(per-site patterns only, and an unarmed site scanned by nothing), got %+v", disarmed)
	}
}

// TestClaimsFloorEditorialSurfaceGatesOnlyTheNumericHalf is bugs_closed/102's
// boundary, re-pinned at this seam: on an editorial page type the prose-number
// scan returns nothing, while banned claims are scanned on EVERY page type
// because a known falsehood is one wherever it is written — the case that
// motivated that layer was found on a guide.
func TestClaimsFloorEditorialSurfaceGatesOnlyTheNumericHalf(t *testing.T) {
	eb := mustParseFloorEB(t)

	sections := []SectionData{
		{ComponentName: "body", HTML: `<p>We span eight departments and serve 45 clients.</p>`},
	}

	blockers, errs := scanSectionClaims(sections, eb, true,
		datahelpers.ClaimSurface{PageType: "guide"}, zap.NewNop())

	if len(blockers) != 1 {
		t.Errorf("a banned claim must be caught on an editorial page type too, got %+v", blockers)
	}
	if len(errs) != 0 {
		t.Errorf("prose numbers must not be scanned on an editorial page type (bugs_closed/102), got %+v", errs)
	}
}

// TestClaimsFloorIgnoresEmptySections guards the degrade: a section with no HTML
// must not be scanned and must not count. Cheap, but the alternative is a save
// refused because of a section that carries nothing.
func TestClaimsFloorIgnoresEmptySections(t *testing.T) {
	eb := mustParseFloorEB(t)

	blockers, errs := scanSectionClaims(
		[]SectionData{{ComponentName: "empty", HTML: ""}},
		eb, true, datahelpers.ClaimSurface{}, zap.NewNop())

	if len(blockers) != 0 || len(errs) != 0 {
		t.Errorf("an empty section must produce nothing, got %+v / %+v", blockers, errs)
	}
}

// TestClaimsFloorErrorNamesTheOffendingSection pins the refusal message. The
// operator reading a failed build needs to know WHICH component to open; a
// message saying only "this page has a banned claim" sends them to read the
// whole page.
func TestClaimsFloorErrorNamesTheOffendingSection(t *testing.T) {
	eb := mustParseFloorEB(t)

	sections := []SectionData{
		{ComponentName: "about-body", HTML: `<p>We span eight departments.</p>`},
	}
	blockers, _ := scanSectionClaims(sections, eb, true, datahelpers.ClaimSurface{}, zap.NewNop())
	if len(blockers) != 1 {
		t.Fatalf("precondition: expected 1 blocker, got %d", len(blockers))
	}

	// The message is assembled in claimsGuardBeforePersist; this asserts the
	// ingredients it formats are present and specific.
	if blockers[0].Section != "about-body" {
		t.Errorf("section name missing from the finding, got %q", blockers[0].Section)
	}
	if !strings.Contains(strings.ToLower(blockers[0].Issue.Value), "eight departments") {
		t.Errorf("the matched text must be carried so the refusal can quote it, got %q",
			blockers[0].Issue.Value)
	}
}
