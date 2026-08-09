// FILE: platform/orchestration/actions/discovery_checks/check_revenue_shape_test.go
//
// The lexicon's false-positive discipline is the load-bearing half: prose ABOUT
// hiring must never fire, only anchor/button TEXT. Plus the branch routing and
// the detector↔verifier no-drift property (both call scanHTMLForServiceCTA —
// asserted by exercising the verifier through the same fixtures).

package discovery_checks

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestScanHTMLForServiceCTA_LexiconDiscipline(t *testing.T) {
	cases := []struct {
		name string
		html string
		want int
	}{
		{"anchor with service CTA", `<a href="/contact.html">Start a Project</a>`, 1},
		{"button with service CTA", `<button class="btn">Get a Quote</button>`, 1},
		{"button with nested markup", `<button><span>Book a</span> <b>Consultation</b></button>`, 0},
		// ^ nested tags split the phrase across elements; the stripped text is
		// "Book a  Consultation" (double space) — asserting 0 records the KNOWN
		// LIMIT: multi-space renderings are invisible to the word-bounded regex.
		{"prose about hiring never fires", `<p>Many businesses hire us little by little, or start a project alone.</p>`, 0},
		{"case-insensitive", `<a href="/x">HIRE US</a>`, 1},
		{"phrase inside longer anchor text", `<a href="/x">Ready? Request a quote today</a>`, 1},
		{"our services deliberately absent", `<a href="/services.html">Our Services</a>`, 0},
		{"two distinct phrases dedupe by phrase", `<a href="/a">Get a quote</a><a href="/b">get a QUOTE</a><button>Work with us</button>`, 2},
		{"empty", ``, 0},
	}
	for _, tc := range cases {
		got := scanHTMLForServiceCTA(tc.html)
		if len(got) != tc.want {
			t.Errorf("%s: got %d hits %v, want %d", tc.name, len(got), got, tc.want)
		}
	}
}

func revenueCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.New(),
		Pipeline:  "content",
		AgentType: "quality-discovery-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	}, mock
}

func expectPrimaryModel(mock sqlmock.Sqlmock, domain string, primary interface{}) {
	mock.ExpectQuery(`SELECT s\.domain`).
		WillReturnRows(sqlmock.NewRows([]string{"domain", "primary"}).AddRow(domain, primary))
}

func TestRevenueShape_ToolsSiteWithServiceCTAFilesPerPage(t *testing.T) {
	dctx, mock := revenueCtx(t)
	pageID := uuid.New()
	expectPrimaryModel(mock, "webdesign.co.uk", "saas_tools")
	mock.ExpectQuery(`SELECT p\.id::text, p\.name`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "html", "locked"}).
			AddRow(pageID.String(), "index", `<a href="/contact.html" class="btn">Start a Project</a>`, 0).
			AddRow(uuid.New().String(), "tools", `<a href="/learn/">Read the guides</a>`, 0))
	mock.ExpectQuery(`SELECT string_agg\(rendered_html`).
		WillReturnRows(sqlmock.NewRows([]string{"agg"}).AddRow(`<nav><a href="/tools/">Tools</a></nav>`))

	res, err := (&RevenueShapeCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 item (offending page only), got %d", len(res.WorkItems))
	}
	wi := res.WorkItems[0]
	if wi.ItemType != "revenue_shape_cta" || wi.HandlerAgent != "component-template-fixer" {
		t.Errorf("routing: %s -> %s", wi.ItemType, wi.HandlerAgent)
	}
	if wi.ItemKey != "revenue_shape_cta:"+pageID.String() {
		t.Errorf("item_key = %q", wi.ItemKey)
	}
	if wi.PageID == nil || *wi.PageID != pageID {
		t.Error("page_id must be set — the verifier locates the defect by it")
	}
	// The clean page is a positive re-scan → narrow retraction by its key.
	if len(res.Resolved) != 1 || res.Resolved[0].ItemType != "revenue_shape_cta" {
		t.Errorf("clean page must retract narrowly, got %+v", res.Resolved)
	}
}

func TestRevenueShape_ChromeHitIsResidueNeverDispatchable(t *testing.T) {
	dctx, mock := revenueCtx(t)
	expectPrimaryModel(mock, "gamesdesign.co.uk", "display_advertising")
	mock.ExpectQuery(`SELECT p\.id::text, p\.name`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "html", "locked"}))
	mock.ExpectQuery(`SELECT string_agg\(rendered_html`).
		WillReturnRows(sqlmock.NewRows([]string{"agg"}).
			AddRow(`<header><a href="/contact.html">Book a consultation</a></header>`))

	res, err := (&RevenueShapeCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 capability_gap, got %d items", len(res.WorkItems))
	}
	wi := res.WorkItems[0]
	if wi.ItemType != "capability_gap" || wi.HandlerAgent != "" || wi.Status != "deferred" {
		t.Errorf("chrome residue must be undispatchable: type=%s handler=%q status=%s",
			wi.ItemType, wi.HandlerAgent, wi.Status)
	}
	if !strings.Contains(wi.SpecJSON, "chrome-fork-handler") {
		t.Errorf("gap must name the builder (A4.4 fork handler): %s", wi.SpecJSON)
	}
}

func TestRevenueShape_LeadGenWithoutConversionPathFiles(t *testing.T) {
	dctx, mock := revenueCtx(t)
	expectPrimaryModel(mock, "gaswholesalers.com", "lead_generation")
	// Contact page exists, has a form, but chrome does not link it.
	mock.ExpectQuery(`WITH contactish`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "has_form", "in_chrome"}).
			AddRow(uuid.New().String(), "contact", true, false))

	res, err := (&RevenueShapeCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 item, got %d", len(res.WorkItems))
	}
	wi := res.WorkItems[0]
	if wi.ItemType != "missing_conversion_path" || wi.HandlerAgent != "content-gap-planner" {
		t.Errorf("routing: %s -> %s", wi.ItemType, wi.HandlerAgent)
	}
	if !strings.Contains(wi.Summary, "not linked from the site chrome") {
		t.Errorf("summary must name WHICH half is missing: %s", wi.Summary)
	}
}

func TestRevenueShape_LeadGenWithWorkingPathRetracts(t *testing.T) {
	dctx, mock := revenueCtx(t)
	expectPrimaryModel(mock, "gaswholesalers.com", "lead_generation")
	mock.ExpectQuery(`WITH contactish`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "has_form", "in_chrome"}).
			AddRow(uuid.New().String(), "contact", true, true))

	res, err := (&RevenueShapeCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 1 {
		t.Errorf("working path must retract and file nothing: items=%d resolved=%d",
			len(res.WorkItems), len(res.Resolved))
	}
}

func TestRevenueShape_AffiliateIsOneCapabilityGap(t *testing.T) {
	dctx, mock := revenueCtx(t)
	expectPrimaryModel(mock, "vetcomparison.uk", "affiliate")

	res, err := (&RevenueShapeCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 || res.WorkItems[0].ItemType != "capability_gap" {
		t.Fatalf("affiliate must file exactly one capability_gap, got %+v", res.WorkItems)
	}
}

func TestRevenueShape_EmptyPrimaryModelIsSilent(t *testing.T) {
	// premise_incomplete owns the missing-premise finding; double-filing would
	// make two checks disagree about one defect.
	dctx, mock := revenueCtx(t)
	expectPrimaryModel(mock, "loancalculator.co.uk", nil)

	res, err := (&RevenueShapeCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Errorf("missing premise is not this check's finding: %+v", res)
	}
}

func TestRevenueShape_SponsoredListingsSilenceIsADecision(t *testing.T) {
	dctx, mock := revenueCtx(t)
	expectPrimaryModel(mock, "vonc.com", "sponsored_listings")
	res, err := (&RevenueShapeCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("no rule is stated for sponsored_listings in v1: %+v", res.WorkItems)
	}
}

// --- verifiers: same predicate as the detector, exercised on the same fixtures ---

func TestVerifyRevenueShapeCTA_StillPresentAndResolved(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	pageID := uuid.New()
	target := VerifyTarget{ItemID: uuid.New(), SiteID: uuid.New(), PageID: &pageID, ItemType: "revenue_shape_cta"}

	mock.ExpectQuery(`SELECT COALESCE\(string_agg`).
		WillReturnRows(sqlmock.NewRows([]string{"html"}).AddRow(`<a href="/x">Hire us</a>`))
	res, err := VerifyRevenueShapeCTAResolved(context.Background(), db, target, zap.NewNop())
	if err != nil || res.Resolved {
		t.Errorf("lexicon still present must not resolve (res=%+v err=%v)", res, err)
	}

	mock.ExpectQuery(`SELECT COALESCE\(string_agg`).
		WillReturnRows(sqlmock.NewRows([]string{"html"}).AddRow(`<a href="/learn/">Read the guides</a>`))
	res, err = VerifyRevenueShapeCTAResolved(context.Background(), db, target, zap.NewNop())
	if err != nil || !res.Resolved {
		t.Errorf("clean page must resolve (res=%+v err=%v)", res, err)
	}
}

func TestVerifyRevenueShapeCTA_NoPageIDIsAnErrorNotAPass(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	_, verr := VerifyRevenueShapeCTAResolved(context.Background(), db,
		VerifyTarget{ItemID: uuid.New(), SiteID: uuid.New(), ItemType: "revenue_shape_cta"}, zap.NewNop())
	if verr == nil {
		t.Fatal("a target the verifier cannot locate must error — fail-closed (RFC_017) turns that into a refusal, not a completion")
	}
}

func TestVerifyConversionPath_BothVerdicts(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	target := VerifyTarget{ItemID: uuid.New(), SiteID: uuid.New(), ItemType: "missing_conversion_path"}

	mock.ExpectQuery(`WITH contactish`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "has_form", "in_chrome"}).
			AddRow(uuid.New().String(), "contact", true, true))
	res, err := VerifyConversionPathResolved(context.Background(), db, target, zap.NewNop())
	if err != nil || !res.Resolved {
		t.Errorf("working path must resolve (res=%+v err=%v)", res, err)
	}

	mock.ExpectQuery(`WITH contactish`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "has_form", "in_chrome"}))
	res, err = VerifyConversionPathResolved(context.Background(), db, target, zap.NewNop())
	if err != nil || res.Resolved {
		t.Errorf("no contact page must not resolve (res=%+v err=%v)", res, err)
	}
}
