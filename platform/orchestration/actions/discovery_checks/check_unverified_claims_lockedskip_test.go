// FILE: platform/orchestration/actions/discovery_checks/check_unverified_claims_lockedskip_test.go
//
// Discharges the `editquality` LOW objection from council round 7
// (b67eb26a-14ef-45d7-b755-3e489fd57ef0) and the same point raised
// independently by a second seat: the 2026-08-09 move of the locked-component
// filter out of SQL (`AND pc.locked_at IS NULL`) and into the Go loop was
// documented as leaving emitted output unchanged, and that was ASSERTED, never
// demonstrated. This check runs fleet-wide on the discovery rotation, so a
// regression here silently changes which pages get findings filed.
//
// WHAT IS DEMONSTRATED, precisely: that skipping a locked row inside the loop
// emits exactly what NOT RETURNING that row emitted — same findings, same
// order, same fields. The old SQL predicate's only effect was to withhold those
// rows, so feeding the current function the filtered row set reproduces the old
// behaviour exactly.
//
// WHAT IS NOT DEMONSTRATED, and cannot be from here: this does not replay the
// old BINARY. Both arms below execute today's ScanDeployedClaims. The claim it
// forecloses is the one the council actually doubted — that the Go skip sits
// early enough to be equivalent to absence.
//
// PROVEN BY MUTATION, not by a green run (2026-08-14, against a clean
// `git archive HEAD` tree so no other session's WIP could colour the result):
//
//	mutation applied to check_unverified_claims.go   caught by
//	----------------------------------------------   ----------------------------
//	drop the `continue` from the locked branch        tests 1 and 2
//	move the skip BELOW the scans (the exact fear)    tests 1 and 2
//	re-add `AND pc.locked_at IS NULL` to the page SQL test 3 only
//
// The third row is the informative one. Restoring the SQL predicate does NOT
// change emitted findings — which is precisely the equivalence the council
// asked to see — so the differential tests stay green and only the query-text
// contract notices. Test 3 is therefore a CONTRACT test, not a behavioural one:
// sqlmock does not execute predicates, so the behavioural consequence of that
// filter (locked rows never arriving at all) is modelled by arm A of tests 1
// and 2 instead, which feed the loop exactly the rows the old SQL returned.
//
// ⚠ ASYMMETRY, pinned by the third test rather than fixed here: only the
// page_components side moved to Go. The site_components query still carries
// `AND sc.locked_at IS NULL`, so locked CHROME rows are still dropped before
// anything can count them — the very shape the page-side move was made to end.
// Out of scope for this test; recorded so it is not mistaken for done.

package discovery_checks

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// lockedSkipComponent is one page_components row as ScanDeployedClaims selects
// it. Field order mirrors the SELECT so a column change breaks this loudly.
type lockedSkipComponent struct {
	pageID       string
	pageName     string
	slotName     string
	renderedHTML string
	component    string
	contentJSON  string
	pageType     string
	lockedAt     sql.NullTime
	updatedAt    sql.NullTime
}

// lockedSkipEvidence is a two-pattern register, so the locked component's claim
// and the unlocked component's claim are distinguishable in the output.
func lockedSkipEvidence(t *testing.T) *datahelpers.EvidenceBase {
	t.Helper()
	eb, err := datahelpers.ParseEvidenceBase([]byte(`{
		"audit_doc":"x",
		"facts":[],
		"banned_claims":[
			{"pattern":"award.winning","reason":"unlocked-surface marker"},
			{"pattern":"industry.leading","reason":"locked-surface marker"}
		]}`))
	if err != nil {
		t.Fatalf("evidence base fixture did not parse: %v", err)
	}
	if eb == nil || len(eb.BannedClaims) != 2 {
		t.Fatalf("fixture did not produce a two-pattern register: %+v", eb)
	}
	return eb
}

// runLockedSkipScan drives ScanDeployedClaims over exactly the rows given. The
// site_components arm returns nothing: this refactor did not touch it.
func runLockedSkipScan(t *testing.T, rows []lockedSkipComponent) ([]*ClaimsPageScan, []unverifiedClaimFinding) {
	t.Helper()
	scans, siteFindings, _ := runLockedSkipScanCapturingSQL(t, rows)
	return scans, siteFindings
}

// runLockedSkipScanCapturingSQL additionally returns the SQL the scan actually
// handed the driver, in order. Asserting on THAT rather than on the source file
// matters here: `AND pc.locked_at IS NULL` also appears in ScanDeployedClaims'
// own doc comment, where it describes the predicate that was REMOVED — a
// source-scanning test would match the prose and report the reverse of the truth.
func runLockedSkipScanCapturingSQL(t *testing.T, rows []lockedSkipComponent) ([]*ClaimsPageScan, []unverifiedClaimFinding, []string) {
	t.Helper()

	var seen []string
	// Always matches: query IDENTITY is established by sqlmock's ordering, and
	// query CONTENT is asserted explicitly by the caller against `seen`.
	recorder := sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		seen = append(seen, actualSQL)
		return nil
	})

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(recorder))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageRows := sqlmock.NewRows([]string{
		"page_id", "name", "slot_name", "rendered_html", "component",
		"content_data", "page_type", "locked_at", "updated_at",
	})
	for _, r := range rows {
		pageRows.AddRow(r.pageID, r.pageName, r.slotName, r.renderedHTML,
			r.component, r.contentJSON, r.pageType, r.lockedAt, r.updatedAt)
	}
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(pageRows)
	mock.ExpectQuery("FROM site_components sc").WillReturnRows(
		sqlmock.NewRows([]string{"slot_name", "rendered_html", "component", "content_data"}))

	scans, siteFindings, err := ScanDeployedClaims(
		context.Background(), db, uuid.New(), "", lockedSkipEvidence(t), true, zap.NewNop())
	if err != nil {
		t.Fatalf("ScanDeployedClaims: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations — the scan did not run the queries this test assumes: %v", err)
	}
	if len(seen) != 2 {
		t.Fatalf("expected exactly the page query then the site query; captured %d: %q", len(seen), seen)
	}
	return scans, siteFindings, seen
}

const (
	lockedSkipPageID   = "11111111-1111-1111-1111-111111111111"
	lockedSkipUnlocked = "<p>We are an award-winning studio.</p>"
	lockedSkipLocked   = "<p>We are an industry-leading studio.</p>"
)

// unlockedRow / lockedRow are the same page, two components, each carrying a
// DIFFERENT banned pattern so the output says which surface it came from.
func unlockedRow() lockedSkipComponent {
	return lockedSkipComponent{
		pageID: lockedSkipPageID, pageName: "about", slotName: "main",
		renderedHTML: lockedSkipUnlocked, component: "prose", contentJSON: "",
		pageType:  "standard",
		updatedAt: sql.NullTime{Time: time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC), Valid: true},
	}
}

func lockedRow() lockedSkipComponent {
	return lockedSkipComponent{
		pageID: lockedSkipPageID, pageName: "about", slotName: "pinned",
		renderedHTML: lockedSkipLocked, component: "prose", contentJSON: "",
		pageType:  "standard",
		lockedAt:  sql.NullTime{Time: time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC), Valid: true},
		updatedAt: sql.NullTime{Time: time.Date(2026, 8, 12, 18, 0, 0, 0, time.UTC), Valid: true},
	}
}

// The differential test. Arm A is what the old SQL predicate handed the loop
// (locked rows withheld); arm B is what today's query hands it (locked rows
// present, skipped in Go). Emitted findings must be identical.
func TestLockedSkipEmitsExactlyWhatTheSQLFilterEmitted(t *testing.T) {
	// Arm A — old behaviour: `AND pc.locked_at IS NULL` withheld the locked row.
	oldScans, oldSiteFindings := runLockedSkipScan(t, []lockedSkipComponent{unlockedRow()})
	// Arm B — current behaviour: both rows arrive, the loop skips the locked one.
	newScans, newSiteFindings := runLockedSkipScan(t, []lockedSkipComponent{unlockedRow(), lockedRow()})

	if len(oldScans) != 1 || len(newScans) != 1 {
		t.Fatalf("expected one page scan per arm; got old=%d new=%d", len(oldScans), len(newScans))
	}

	// THE DEMAND CONTROL. Two empty finding lists are trivially equal, and an
	// equivalence proved on nothing is the failure mode this whole test exists to
	// avoid. If the fixture stops tripping the scan, this test must die rather
	// than quietly certify emptiness.
	if len(oldScans[0].Findings) == 0 {
		t.Fatalf("fixture produced NO findings, so the equivalence below would be vacuous — " +
			"the banned-claim patterns no longer fire on the fixture copy")
	}

	if !reflect.DeepEqual(oldScans[0].Findings, newScans[0].Findings) {
		t.Errorf("emitted findings CHANGED when the locked row was returned to the loop.\n old: %+v\n new: %+v",
			oldScans[0].Findings, newScans[0].Findings)
	}
	if !reflect.DeepEqual(oldSiteFindings, newSiteFindings) {
		t.Errorf("site findings changed: old=%+v new=%+v", oldSiteFindings, newSiteFindings)
	}

	// Named rather than left to DeepEqual: the locked component's own claim must
	// never reach the output. This is the assertion that fails if the `continue`
	// is moved below the scans.
	for _, f := range newScans[0].Findings {
		if f.SlotName == "pinned" {
			t.Errorf("a LOCKED component produced a finding (%s / %q) — human-pinned content is never flagged",
				f.Check, f.Matched)
		}
	}

	// The examined-text record must not carry copy the scan did not read: the
	// revalidator's claim-granular gate searches it, and text from a locked slot
	// would make a live claim look present on a surface nobody audited.
	if _, leaked := newScans[0].ExaminedTextBySlot["pinned"]; leaked {
		t.Errorf("locked slot appears in ExaminedTextBySlot: %+v", newScans[0].ExaminedTextBySlot)
	}

	// The strictly-ADDED information — the whole point of the move. Old arm could
	// not report it at all.
	if newScans[0].ComponentsExamined != 1 || newScans[0].ComponentsSkippedLocked != 1 {
		t.Errorf("counts = examined %d / skipped-locked %d, want 1 / 1",
			newScans[0].ComponentsExamined, newScans[0].ComponentsSkippedLocked)
	}
	if oldScans[0].ComponentsSkippedLocked != 0 {
		t.Errorf("arm A must see no locked rows at all; got skipped-locked %d", oldScans[0].ComponentsSkippedLocked)
	}

	// NewestComponentUpdate is derived from EXAMINED rows only. The locked row is
	// deliberately the NEWER of the two, so a leak here is unmissable — and it
	// would be consequential: the copy-changed gate compares this against the
	// item's created_at to decide whether the page moved.
	if !newScans[0].NewestComponentUpdate.Equal(oldScans[0].NewestComponentUpdate) {
		t.Errorf("NewestComponentUpdate moved when the locked row was returned: old=%s new=%s — "+
			"a locked component's timestamp is not evidence the audited copy changed",
			oldScans[0].NewestComponentUpdate, newScans[0].NewestComponentUpdate)
	}
}

// The one place the two behaviours DO differ, pinned so it stays deliberate.
// Old: the page had no unlocked rows, so it produced no scan at all and was
// indistinguishable from a page that had been cleaned. New: the page appears
// with nothing examined, which is what lets the revalidator answer "cannot
// answer" instead of "the copy was fixed".
func TestPageWhoseOnlyComponentIsLockedIsVisibleAndUnanswerable(t *testing.T) {
	oldScans, _ := runLockedSkipScan(t, []lockedSkipComponent{})
	newScans, _ := runLockedSkipScan(t, []lockedSkipComponent{lockedRow()})

	if len(oldScans) != 0 {
		t.Fatalf("arm A premise wrong: a filtered-out row still produced %d page scan(s)", len(oldScans))
	}
	if len(newScans) != 1 {
		t.Fatalf("an all-locked page must still appear in the scan list; got %d page scans", len(newScans))
	}
	got := newScans[0]
	if got.ComponentsExamined != 0 || got.ComponentsSkippedLocked != 1 {
		t.Errorf("counts = examined %d / skipped-locked %d, want 0 / 1", got.ComponentsExamined, got.ComponentsSkippedLocked)
	}
	if len(got.Findings) != 0 {
		t.Errorf("an all-locked page emitted findings: %+v", got.Findings)
	}
	// len(Findings)==0 with Examined==0 is the state the revalidator must NOT read
	// as a clean page. Asserted here because the emit side is what produces it.
	if got.ComponentsExamined == 0 && len(got.Findings) == 0 && got.ComponentsSkippedLocked == 0 {
		t.Error("unreachable given the assertions above; kept as a tripwire if the counts are ever dropped")
	}
}

// Where each side does its locked filtering, asserted on the SQL the scan
// actually issued — not on the source file, and not on a comment.
//
// This states the CURRENT contract rather than the desired one. Two ways to
// fail, and each names its own remedy:
//   - the page query regains an SQL locked filter -> the all-locked page pinned
//     by the previous test goes invisible again, and "cannot answer" silently
//     becomes "the copy was fixed";
//   - the site query loses its SQL filter -> locked chrome rows start reaching
//     the loop, which does not count them, so they would be dropped silently in
//     a new place. That move needs a skipped-chrome counter alongside it.
func TestLockedFilteringHappensWhereEachSideClaimsItDoes(t *testing.T) {
	_, _, seen := runLockedSkipScanCapturingSQL(t, []lockedSkipComponent{unlockedRow()})
	pageSQL, siteSQL := seen[0], seen[1]

	// Premise check: the capture is in the order this test assumes. Without it,
	// the two assertions below could silently swap and both still pass.
	if !strings.Contains(pageSQL, "FROM page_components pc") || !strings.Contains(siteSQL, "FROM site_components sc") {
		t.Fatalf("captured queries are not (page, site) as assumed:\n [0] %s\n [1] %s", pageSQL, siteSQL)
	}

	if strings.Contains(pageSQL, "pc.locked_at IS NULL") {
		t.Errorf("the page query has regained its SQL locked filter — that re-hides the all-locked "+
			"page TestPageWhoseOnlyComponentIsLockedIsVisibleAndUnanswerable pins as visible.\n%s", pageSQL)
	}
	// It must still SELECT the column, or the loop has nothing to skip on.
	if !strings.Contains(pageSQL, "pc.locked_at") {
		t.Errorf("the page query no longer selects pc.locked_at, so the Go skip cannot fire:\n%s", pageSQL)
	}

	if !strings.Contains(siteSQL, "sc.locked_at IS NULL") {
		t.Errorf("the site_components query no longer filters locked rows in SQL. If that was "+
			"deliberate, the loop must now COUNT skipped chrome rows the way the page loop does — "+
			"otherwise this reintroduces the silent drop the page-side move was made to end.\n%s", siteSQL)
	}
}
