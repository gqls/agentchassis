// FILE: platform/orchestration/actions/validate_page_content_fleetwide_claims_test.go
//
// bugs_open/104 — the WIRING, not the patterns.
//
// datahelpers/claims_global_test.go proves the fleet-wide set matches the right
// sentences. That is not the same claim as "the build gate scans a site that has
// no evidence_base row", which is the entire point of 104 and is a property of
// check 8's structure: the banned-claim scan had to move OUT of the `eb != nil`
// guard while the numeric scan stayed inside it.
//
// So these tests drive the real action with a database that returns NO
// evidence_base row — the unarmed-site case — and assert a blocker still comes
// out. Without this, the only evidence that the restructure works is that it
// compiles.

package actions

import (
	"context"
	"database/sql"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// unarmedSiteGate runs ValidatePageContentAction over html for a site whose
// evidence_base query returns sql.ErrNoRows. Every other DB-touching check is
// switched off by config, so loadEvidenceBase is the only query in play and the
// mock stays honest about what the gate actually asks the database for.
// It returns the action's error as well as its issues, because THE ERROR IS THE
// MECHANISM: on any blocker the action returns (nil, error) and the page build
// fails. A test that only inspected the issues list would report a pass on the
// case where the gate had refused the build, which is the outcome 104 wants.
func unarmedSiteGate(t *testing.T, html string) ([]ValidationIssue, error) {
	return unarmedSiteGateCfg(t, html, nil)
}

// unarmedSiteGateCfg is the same harness with config overrides, for the reversal
// lever. extra is merged over the defaults.
func unarmedSiteGateCfg(t *testing.T, html string, extra map[string]interface{}) ([]ValidationIssue, error) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// THE UNARMED SITE: no evidence_base row at all. vetcomparison.uk and
	// idea.uk are in exactly this state (verified 2026-07-28).
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnError(sql.ErrNoRows)

	cfg := map[string]interface{}{
		"site_id":              uuid.New().String(),
		"check_claims":         true,
		"check_internal_links": false,
		"check_emails":         false,
		"check_stat_claims":    false,
		"check_stat_units":     false,
	}
	for k, v := range extra {
		cfg[k] = v
	}

	res, err := ValidatePageContentAction(context.Background(), ActionParams{
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: cfg},
		CollectedData: map[string]interface{}{
			"page_content": map[string]interface{}{
				"response": map[string]interface{}{"page_html": html},
			},
		},
		DB:     db,
		Logger: zap.NewNop(),
	})
	if out, ok := res.(map[string]interface{}); ok {
		issues, _ := out["issues"].([]ValidationIssue)
		return issues, err
	}
	return nil, err
}

func claimsIssues(issues []ValidationIssue) []ValidationIssue {
	var out []ValidationIssue
	for _, i := range issues {
		if i.Category == "claims" {
			out = append(out, i)
		}
	}
	return out
}

// The 104 acceptance case, from the bug file's own "How to verify a fix": a page
// asserting "every claim on this site is verified" must fail the build on a site
// with NO register.
func TestUnarmedSiteStillBlocksFleetWideOverclaim(t *testing.T) {
	_, err := unarmedSiteGate(t, `<p>Every claim on this site is verified.</p>`)

	// The build FAILING is the outcome under test. Before bugs_open/104 this
	// returned no error at all on a site with no evidence_base row.
	if err == nil {
		t.Fatal("an unarmed site published an overclaim without failing the build — 104 is not fixed")
	}
	if !strings.Contains(err.Error(), "1 blockers") {
		t.Errorf("want the failure attributed to exactly 1 blocker, got: %v", err)
	}
}

// The other half of the induced pair the bug file demands, and the reason the
// negation-prone pattern was excluded: legitimate and hedged copy must still
// build on an unarmed site. A checker that fires on everything is as useless as
// one that fires on nothing.
func TestUnarmedSiteDoesNotBlockLegitimateCopy(t *testing.T) {
	for _, html := range []string{
		`<p>We cite each figure and date it.</p>`,
		`<p>Where a figure has not been independently verified, that is stated.</p>`,
		`<p>Every component is verified against production.</p>`, // owner ruling 2026-07-28
	} {
		got, err := unarmedSiteGate(t, html)
		if err != nil {
			t.Errorf("FALSE POSITIVE — legitimate copy failed the build for %q: %v", html, err)
		}
		if c := claimsIssues(got); len(c) != 0 {
			t.Errorf("FALSE POSITIVE at the gate for %q: %+v", html, c)
		}
	}
}

// The numeric half must NOT have been dragged fleet-wide with the banned half.
// It is opt-in precisely because its extractor has false positives, and it is
// filed at error rather than blocker for the same reason — so on an unarmed site
// it must stay silent even though the prose is full of business numbers.
func TestUnarmedSiteRaisesNoUnregisteredNumbers(t *testing.T) {
	got, err := unarmedSiteGate(t, `<p>We serve 4,200 clients across 37 departments.</p>`)
	if err != nil {
		t.Fatalf("business numbers must not fail an unarmed site's build: %v", err)
	}
	issues := claimsIssues(got)
	for _, i := range issues {
		if i.Type == "unregistered_number" {
			t.Errorf("the number scan must stay opt-in on a site with no register, got %+v", i)
		}
	}
	if len(issues) != 0 {
		t.Errorf("expected no claims issues at all on this copy, got %+v", issues)
	}
}

// Guards the join at the gate rather than in datahelpers: an ARMED site must get
// its own patterns AND the fleet-wide set out of the same call.
func TestArmedSiteGetsBothItsOwnAndTheFleetWideSet(t *testing.T) {
	eb, err := datahelpers.ParseEvidenceBase([]byte(claimsGateTestEB))
	if err != nil || eb == nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	blocks := datahelpers.ExtractAssertionText(
		`<p>We span eight departments.</p><p>Our reporting is always accurate.</p>`)

	got := checkBannedClaims(blocks, eb, true, "test-site", zap.NewNop())
	if len(got) != 2 {
		t.Fatalf("want 2 blockers (1 per-site + 1 fleet-wide), got %d: %+v", len(got), got)
	}
	for _, i := range got {
		if i.Severity != "blocker" {
			t.Errorf("both sources are blockers; got %q for %+v", i.Severity, i)
		}
	}
}

// ---------------------------------------------------------------------------
// The reversal lever (config check_claims_fleet_wide), asked for by the
// council's guardian seat: the set is enforced at blocker severity on every
// site, so withdrawing a bad pattern must not require a commit + build + roll.
// DB config is live immediately; these prove the flag actually reaches the scan.
// ---------------------------------------------------------------------------

func TestFleetWideSetCanBeWithdrawnByConfig(t *testing.T) {
	html := `<p>Every claim on this site is verified.</p>`

	// Default (absent key) and explicit true both enforce.
	if _, err := unarmedSiteGate(t, html); err == nil {
		t.Fatal("default must enforce — an off-by-default set would leave 104 live")
	}
	if _, err := unarmedSiteGateCfg(t, html, map[string]interface{}{
		"check_claims_fleet_wide": true,
	}); err == nil {
		t.Fatal("explicit true must enforce")
	}

	// Withdrawn: exactly the pre-104 behaviour — an unarmed site is scanned by
	// nothing, so the same page builds.
	got, err := unarmedSiteGateCfg(t, html, map[string]interface{}{
		"check_claims_fleet_wide": false,
	})
	if err != nil {
		t.Errorf("with the lever off an unarmed site must build as it did before 104: %v", err)
	}
	if c := claimsIssues(got); len(c) != 0 {
		t.Errorf("with the lever off there must be no claims findings, got %+v", c)
	}
}

// Withdrawing the fleet-wide set must NOT disarm a site's own audited patterns —
// those predate 104 and are not this lever's business.
func TestWithdrawingFleetWideSetKeepsPerSitePatterns(t *testing.T) {
	eb, err := datahelpers.ParseEvidenceBase([]byte(claimsGateTestEB))
	if err != nil || eb == nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	blocks := datahelpers.ExtractAssertionText(
		`<p>We span eight departments.</p><p>Our reporting is always accurate.</p>`)

	off := checkBannedClaims(blocks, eb, false, "test-site", zap.NewNop())
	if len(off) != 1 {
		t.Fatalf("lever off must keep the site's OWN pattern and drop only the fleet-wide one, got %d: %+v",
			len(off), off)
	}
	if off[0].Severity != "blocker" {
		t.Errorf("a per-site banned claim is still a blocker, got %q", off[0].Severity)
	}
}

// ---------------------------------------------------------------------------
// The negation guard must not be silent AT THE GATE.
//
// Raised by the council's architecture seat at medium (corr 8a41e1a5, round 1):
// the first version of this change made suppression observable only in
// cmd/claimscan — an offline tool someone has to think to run. That is the
// bugs_open/093 shape, in a change that cited 093. So the build gate logs what it
// dropped, and this pins it: if the log line goes away, a future guard bug becomes
// invisible in exactly the place it would do damage.
// ---------------------------------------------------------------------------

func TestGateLogsWhatTheNegationGuardSuppressed(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	// Verbatim live copy (robot-hands.com index/features) — honest, negated, and
	// it matches a fleet-wide pattern, so the guard is what keeps it quiet.
	blocks := datahelpers.ExtractAssertionText(
		"<p>Where manufacturer data has not been independently verified, that is stated explicitly.</p>")

	issues := checkBannedClaims(blocks, nil, true, "site-under-test", logger)
	if len(issues) != 0 {
		t.Fatalf("negated copy must not raise an issue at the gate, got %d", len(issues))
	}

	found := logs.FilterMessage("claims gate: banned-claim match suppressed as negated").All()
	if len(found) == 0 {
		t.Fatal("the gate suppressed a match and said nothing — a silent suppressor and a " +
			"dead gate are indistinguishable, which is the whole reason this change exists")
	}
	var sawSite bool
	for _, f := range found[0].Context {
		if f.Key == "site_id" && f.String == "site-under-test" {
			sawSite = true
		}
	}
	if !sawSite {
		t.Error("suppression log must name the site; without it the line cannot be traced to a build")
	}
}

// With the reversal lever OFF the gate runs the pre-104 scan, so it must not
// report suppressions for a scan it did not perform — otherwise pulling the lever
// produces log noise describing a mechanism that is switched off.
func TestGateReportsNoSuppressionWhenFleetWideSetIsDisabled(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	blocks := datahelpers.ExtractAssertionText(
		"<p>Where manufacturer data has not been independently verified, that is stated explicitly.</p>")

	if issues := checkBannedClaims(blocks, nil, false, "site-under-test", logger); len(issues) != 0 {
		t.Fatalf("lever off + no register must raise nothing, got %d", len(issues))
	}
	if n := logs.FilterMessage("claims gate: banned-claim match suppressed as negated").Len(); n != 0 {
		t.Errorf("lever is off, so no scan ran to suppress anything — got %d suppression log(s)", n)
	}
}
