// FILE: platform/orchestration/actions/discovery_checks/check_decision_guards_test.go
//
// bugfix_280: storedPageAssemblySQL used to omit chrome entirely (the same
// vestigial pages.rendered_header/footer columns bugs_open/270 documents),
// so a contains guard on header/footer content could never be satisfied
// (permanent false positive) and a not_contains guard on chrome content
// could never be violated (permanent false negative). The two behavioural
// tests below are demand controls for those exact shapes: fed a mocked
// "assembled" string that actually carries chrome, they prove the CALLER
// handles chrome correctly. TestDecisionGuardsAssemblyReadsSiteComponents
// is what proves the real SQL still produces that string in the first
// place, following check_missing_structure_test.go's rationale exactly.
package discovery_checks

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func newDecisionGuardsCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.MustParse("00ff3af5-dad8-4770-9f70-3edc267a3c92"),
		Pipeline:  "build",
		AgentType: "completeness-discovery-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	}, mock
}

func expectDecisionRow(mock sqlmock.Sqlmock, key, body string) {
	mock.ExpectQuery(`SELECT subject_key, body`).WillReturnRows(
		sqlmock.NewRows([]string{"subject_key", "body"}).AddRow(key, body))
}

func expectAssembledPage(mock sqlmock.Sqlmock, assembled string) {
	mock.ExpectQuery(`FROM pages pg`).WillReturnRows(
		sqlmock.NewRows([]string{"assembled"}).AddRow(assembled))
}

const chromeContainsGuardBody = "```guard\n" +
	`{"page": "index", "assert": "contains", "pattern": "site-nav-cta"}` +
	"\n```"

const chromeNotContainsGuardBody = "```guard\n" +
	`{"page": "index", "assert": "not_contains", "pattern": "old-banner-copy"}` +
	"\n```"

// TestDecisionGuardsContainsGuardSeesChromeContent is the demand control for
// bug 280's false-positive shape: a contains guard whose pattern lives only
// in the header must NOT fire once the assembled string actually carries
// chrome. Before the fix this pattern could never be present in the SQL's
// own output, so this case could not pass.
func TestDecisionGuardsContainsGuardSeesChromeContent(t *testing.T) {
	dctx, mock := newDecisionGuardsCtx(t)
	expectDecisionRow(mock, "D-TEST-chrome", chromeContainsGuardBody)
	expectAssembledPage(mock, `<header><a class="site-nav-cta" href="/x">Go</a></header><footer></footer><p>body</p>`)

	res, err := (&DecisionGuardsCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("a contains guard satisfied by chrome content must not fire; got %d work items", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestDecisionGuardsNotContainsGuardCatchesChromeRegression is the demand
// control for bug 280's false-negative shape: a not_contains guard on chrome
// content must fire when the assembled string actually carries that content.
// Before the fix, chrome was always absent from "assembled", so this guard
// could never see the regression it exists to catch.
func TestDecisionGuardsNotContainsGuardCatchesChromeRegression(t *testing.T) {
	dctx, mock := newDecisionGuardsCtx(t)
	expectDecisionRow(mock, "D-TEST-chrome-regression", chromeNotContainsGuardBody)
	expectAssembledPage(mock, `<header>old-banner-copy</header><footer></footer><p>body</p>`)

	res, err := (&DecisionGuardsCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("a not_contains guard violated by chrome content must fire; got %d work items", len(res.WorkItems))
	}
	if res.WorkItems[0].ItemType != "decision_regression" {
		t.Errorf("ItemType = %q, want decision_regression", res.WorkItems[0].ItemType)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestDecisionGuardsAssemblyReadsSiteComponents is the anti-regression test:
// it asserts on the SQL actually issued, because every behavioural test above
// stays green regardless of what predicate produced the mocked "assembled"
// row — see check_missing_structure_test.go's header for the same reasoning.
// This is bug 280 itself: the query must never again read
// pages.rendered_header / pages.rendered_footer, and must still gate on the
// page existing (FROM pages pg WHERE ...) so a nonexistent page still yields
// zero rows rather than a row with empty chrome — VerifyDecisionRegressionResolved
// relies on that to treat a missing page as an error, not a resolved guard.
func TestDecisionGuardsAssemblyReadsSiteComponents(t *testing.T) {
	issued := storedPageAssemblySQL

	for _, want := range []struct{ name, pattern string }{
		{"reads site_components for header", `site_components sc\s+WHERE sc\.site_id = \$1 AND sc\.slot_name = 'header'`},
		{"reads site_components for footer", `site_components sc\s+WHERE sc\.site_id = \$1 AND sc\.slot_name = 'footer'`},
		{"still reads page_components for body", `FROM page_components pc WHERE pc\.page_id = pg\.id`},
		{"still gates on the page existing", `FROM pages pg\s+WHERE pg\.site_id = \$1 AND pg\.name = \$2`},
	} {
		re := regexp.MustCompile(`(?s)` + want.pattern)
		if !re.MatchString(issued) {
			t.Errorf("%s missing from query: /%s/ did not match:\n%s", want.name, want.pattern, issued)
		}
	}
	if regexp.MustCompile(`pg\.rendered_(header|footer)`).MatchString(issued) {
		t.Errorf("query must not read pages.rendered_header/footer — that is the vestigial-column defect this fix removes:\n%s", issued)
	}
}
