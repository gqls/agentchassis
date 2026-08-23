package actions

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// bugs_open/328. A page that failed to build is still linked from the pages that
// did, so one blocked page becomes a visibly broken site. loanzy.uk's home page
// was deployed 2026-08-23 13:28:27Z carrying two anchors to /your-rights.html
// and one to /guides/index.html — both 404, both targets untouched since
// 2026-08-18 with no open work item. The information was in the pages table at
// render time and nothing consulted it.
//
// These tests pin the policy and its two escapes. The pass itself is pinned in
// datahelpers/link_suppress_test.go.

func refusedTargetsObserved() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.WarnLevel)
	return zap.New(core), logs
}

func refusedTargetRows(rows ...[3]interface{}) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"url", "refused", "shipped"})
	for _, row := range rows {
		r.AddRow(row[0], row[1], row[2])
	}
	return r
}

// TestLoaderUsesTheSharedPredicateAndNotASecondSpelling is the lockstep. The
// judgement "would an anchor here 404, and is the page not arriving" has ONE
// home, datahelpers.PageLinkRefusedPredicateFor. A hand-rolled `build_status <>
// 'deployed'` in this loader would be a second definition, which is the drift
// this estate keeps re-finding (bugs_open/185).
func TestLoaderUsesTheSharedPredicateAndNotASecondSpelling(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	// Matching on a fragment of the shared predicate means that if someone
	// replaces it with their own SQL, the expectation stops matching and the
	// query returns an error — so this test goes red rather than silently
	// approving a second spelling.
	mock.ExpectQuery(regexp.QuoteMeta("COALESCE(p.build_status, '') = 'planned'")).
		WithArgs(siteID).
		WillReturnRows(refusedTargetRows(
			[3]interface{}{"/your-rights.html", true, false},
			[3]interface{}{"/calculators.html", false, true},
		))

	logger, _ := refusedTargetsObserved()
	set, ok := loadRefusedLinkTargets(context.Background(), db, siteID, logger)
	if !ok {
		t.Fatal("loader reported an untrustworthy set on a healthy query")
	}
	if !set.Contains("/your-rights.html") {
		t.Error("the refused target is missing from the set")
	}
	if set.Contains("/calculators.html") {
		t.Error("a servable page was placed in the REFUSED set — the pass would strip a working link")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestZeroShippedPagesDisablesSuppression is the first-build escape, and it is
// the one that would have mattered most in the fleet: measured 2026-08-23, one
// live domain serves a parked-registrar redirect on every path while holding a
// full set of never-deployed page rows. Without this escape it is the site
// suppression hits hardest.
//
// ⚠ It keys on the SHIPPED-PAGE COUNT, never on the size of the refused set. An
// empty refused set is the normal healthy state of most sites and says nothing
// about whether the site has shipped.
func TestZeroShippedPagesDisablesSuppression(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery("SELECT p.url").WithArgs(siteID).
		WillReturnRows(refusedTargetRows(
			[3]interface{}{"/your-rights.html", true, false},
			[3]interface{}{"/guides/index.html", true, false},
		))

	logger, logs := refusedTargetsObserved()
	set, ok := loadRefusedLinkTargets(context.Background(), db, siteID, logger)

	if ok {
		t.Fatal("a site with NO shipped pages returned a trustworthy refused set — a first build would have its whole link graph stripped")
	}
	if len(set) != 0 {
		t.Errorf("expected a nil set on the escape, got %d entries", len(set))
	}
	if logs.FilterMessageSnippet("NO shipped pages").Len() == 0 {
		t.Error("the escape was taken silently; it must say so — an established site taking it is anomalous and must be findable")
	}
}

// TestFailedLookupSuppressesNothing — an infrastructure blip must not amputate a
// page's internal links. Fail-open, in the opposite direction from
// loadValidPagePaths' bool and for the same reason: there a failed load would
// strip every link, here it must strip none.
func TestFailedLookupSuppressesNothing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery("SELECT p.url").WithArgs(siteID).
		WillReturnError(errors.New("connection reset"))

	logger, logs := refusedTargetsObserved()
	set, ok := loadRefusedLinkTargets(context.Background(), db, siteID, logger)

	if ok || len(set) != 0 {
		t.Fatalf("a failed lookup produced a usable set (ok=%v, n=%d)", ok, len(set))
	}
	if logs.FilterMessageSnippet("lookup failed").Len() == 0 {
		t.Error("a failed lookup was swallowed silently")
	}
}

// TestSuppressionIsOffWithoutTheOptIn is the owner's RFC_010 §2 ruling made
// testable: new authority on a shared seam ships as an opt-in field whose unsafe
// default is OFF. This is the test a reviewer of the enabling migration can
// point at.
//
// ⚠ THE OBVIOUS VERSION OF THIS TEST IS VACUOUS, and I wrote it that way first.
// Registering NO query on the mock does not discriminate: with the flag wrongly
// defaulting ON, the loader runs, sqlmock returns an error for the unexpected
// query, the fail-open branch returns the html unchanged — and the assertion
// passes on the mutant. Flipping the default to `true` was measured to survive
// that test (2026-08-23).
//
// So the mock registers a query that WOULD find a refused target. Now the two
// worlds differ at the OUTPUT: flag OFF, the query is never consumed and the
// html is byte-identical; flag ON, the anchor disappears and this fails. The
// unconsumed expectation is deliberately not asserted — it is the control, not
// the claim.
func TestSuppressionIsOffWithoutTheOptIn(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT p.url").
		WillReturnRows(refusedTargetRows(
			[3]interface{}{"/your-rights.html", true, false},
			[3]interface{}{"/calculators.html", false, true},
		))

	html := `<p><a href="/your-rights.html">rights</a></p>`
	params := ActionParams{
		DB:         db,
		Logger:     zap.NewNop(),
		StepConfig: models.Step{Action: "assemble_page", Config: map[string]interface{}{}},
	}

	got := suppressUnshippedOutboundLinks(context.Background(), params, uuid.New(),
		"loanzy.uk", "index", "/index.html", html, zap.NewNop())

	if got != html {
		t.Errorf("suppression ran with no %s in the step config:\n got: %s\nwant: %s",
			suppressUnshippedLinksKey, got, html)
	}
}

// TestOptInSuppressesAndKeepsTheServableSibling is the end-to-end arm through
// the action-side glue, with the positive control in the same document.
func TestOptInSuppressesAndKeepsTheServableSibling(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery("SELECT p.url").WithArgs(siteID).
		WillReturnRows(refusedTargetRows(
			[3]interface{}{"/your-rights.html", true, false},
			[3]interface{}{"/calculators.html", false, true},
		))
	// The durable record is best-effort and writes through LogActionEntry; any
	// further statements are unexpected-but-tolerated by sqlmock's default
	// ordered matching only if registered, so the record write is not asserted
	// here — writeLinkSuppressionLog's own contract is best-effort.
	mock.MatchExpectationsInOrder(false)

	html := `<p><a href="/your-rights.html">rights</a> and <a href="/calculators.html">calculators</a></p>`
	params := ActionParams{
		DB:     db,
		Logger: zap.NewNop(),
		StepConfig: models.Step{
			Action: "assemble_page",
			Config: map[string]interface{}{suppressUnshippedLinksKey: true},
		},
	}

	got := suppressUnshippedOutboundLinks(context.Background(), params, siteID,
		"loanzy.uk", "index", "/index.html", html, zap.NewNop())

	if strings.Contains(got, `href="/your-rights.html"`) {
		t.Errorf("the refused anchor survived the opt-in path:\n%s", got)
	}
	if !strings.Contains(got, `<a href="/calculators.html">calculators</a>`) {
		t.Errorf("the servable sibling was stripped — the pass is over-reaching:\n%s", got)
	}
}

// TestSuppressionIsConfinedToTheOutboundSeams is the door, not the policy.
//
// Suppression must NEVER run at a seam that writes to the database. The
// persistence chokepoint writes page_components.rendered_html AND content_data,
// and suppressing there would delete the authored href — after which the link
// could never return when the target ships, which is the single property that
// makes this fix safe. A comment saying so is not a control on a tree this many
// sessions share (owner ruling, 2026-08-02), so it is asserted at the source.
func TestSuppressionIsConfinedToTheOutboundSeams(t *testing.T) {
	forbidden := []string{
		"save_sections_link_repair.go",
		"save_sections_content_data_links.go",
		"component_link_repair.go",
	}
	allowed := map[string]bool{
		// The outbound rerender seam — covers both rerender paths AND the build
		// path, whose deploy_page step calls the page-rerender agent by role.
		"rerender_link_repair.go": true,
		// The assemble seam — the initial-build and loop paths.
		"multipage_actions.go": true,
		// The definition itself, and this test.
		"refused_link_targets.go":      true,
		"refused_link_targets_test.go": true,
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}

	callers := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(".", e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		// The BARE IDENTIFIER, not `name(`. Matching the call syntax was
		// measured to miss a mutant that took the function as a value
		// (`_ = suppressUnshippedOutboundLinks`), and a scan that a plausible
		// mutant walks past is not a door. Comment lines are skipped, so
		// naming the helper in prose is not an accusation — the same
		// construction chrome_link_policy_test.go uses.
		if !sourceNamesIdentifier(string(src), "suppressUnshippedOutboundLinks") {
			continue
		}
		callers++
		if !allowed[e.Name()] {
			t.Errorf("%s calls suppressUnshippedOutboundLinks — suppression must run only at an outbound seam that writes nothing", e.Name())
		}
		for _, f := range forbidden {
			if e.Name() == f {
				t.Errorf("%s is a PERSISTENCE seam: suppressing there would delete the authored href from content_data and the link could never return", f)
			}
		}
	}

	// A scan that has stopped seeing is worse than no scan. If the helper is
	// renamed and this literal goes stale, the count drops to zero and this
	// fires rather than quietly passing.
	if callers < 3 {
		t.Fatalf("the source scan found %d files calling suppressUnshippedOutboundLinks; expected at least 3 (both seams plus its own definition) — the scan has gone blind", callers)
	}
}

// sourceNamesIdentifier reports whether the source references the identifier in
// CODE, ignoring comment lines. Split out so the scan above and its own
// synthetic-line test below read the same rule.
func sourceNamesIdentifier(src, ident string) bool {
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		if strings.Contains(line, ident) {
			return true
		}
	}
	return false
}

// TestTheSeamScanCanStillFire is the gone-blind guard on the scan above: a
// source scan that has stopped matching passes silently and for ever. These are
// synthetic lines, not files, so they prove the rule rather than the corpus.
func TestTheSeamScanCanStillFire(t *testing.T) {
	if !sourceNamesIdentifier("\tx := suppressUnshippedOutboundLinks(ctx)\n", "suppressUnshippedOutboundLinks") {
		t.Error("the scan does not match a plain call")
	}
	if !sourceNamesIdentifier("\t_ = suppressUnshippedOutboundLinks\n", "suppressUnshippedOutboundLinks") {
		t.Error("the scan does not match the function taken as a VALUE — the mutant that survived on 2026-08-23")
	}
	if sourceNamesIdentifier("// see suppressUnshippedOutboundLinks for the policy\n", "suppressUnshippedOutboundLinks") {
		t.Error("the scan accuses a comment; naming the helper in prose must not be a violation")
	}
}

// TestRefusedPredicateRefusesToProduceABareForm pins the alias contract at the
// consumer's altitude: the correlated subquery makes an unqualified `id` bind
// inside the EXISTS, which would refuse every page on the site.
func TestRefusedPredicateRefusesToProduceABareForm(t *testing.T) {
	if got := datahelpers.PageLinkRefusedPredicateFor(""); got != "FALSE" {
		t.Errorf("the bare form must degrade to FALSE (suppress nothing), got %q", got)
	}
	aliased := datahelpers.PageLinkRefusedPredicateFor("p")
	if !strings.Contains(aliased, "pc.page_id = p.id") {
		t.Errorf("the aliased form does not qualify the correlated join: %s", aliased)
	}
}
