package discovery_checks

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestEmptySectionVerdict(t *testing.T) {
	// The real defect this verifier was built for: gripper-detail's
	// product-details section — 8.7kB of e-commerce chrome with every value
	// empty, led by an empty <h1 class="pd-title"></h1>.
	hollowProductSection := `<section class="pd">` +
		`<h1 class="pd-title"></h1>` +
		`<span class="pd-price"></span>` +
		`<span class="pd-meta-val"></span>` +
		`<ul class="pd-features"><li></li><li></li><li></li><li></li></ul>` +
		`<button>Add to Cart</button><button>Buy Now</button>` +
		`</section>`

	cases := []struct {
		name         string
		html         string
		wantResolved bool
	}{
		{"empty string", "", false},
		{"whitespace only", "  \n\t ", false},
		{"minimal html", "<div></div>", false},
		{"empty heading shell", hollowProductSection, false},
		{"empty heading with attrs", `<section>` + strings.Repeat("x", 60) + `<h2 class="title"></h2></section>`, false},
		{"uppercase empty heading", `<section>` + strings.Repeat("x", 60) + `<H3></H3></section>`, false},
		{"runtime-fill shell exempt", `<div data-runtime-fill="lobby"><h2></h2></div>`, true},
		{"filled section", `<section><h1 class="pd-title">PG-90 Parallel Gripper</h1><span class="pd-price">£1,240</span></section>`, true},
		{"headless but substantial", `<section><p>` + strings.Repeat("real content ", 10) + `</p></section>`, true},
	}

	for _, tc := range cases {
		got := emptySectionVerdict(tc.html)
		if got.Resolved != tc.wantResolved {
			t.Errorf("%s: emptySectionVerdict resolved = %v (detail %q), want %v",
				tc.name, got.Resolved, got.Detail, tc.wantResolved)
		}
	}
}

func TestEmptyHeadingReMirrorsSQL(t *testing.T) {
	// The SQL pattern is '<(h[1-6])[^>]*>\s*</\1>'. RE2 has no backrefs, so
	// the Go mirror accepts mismatched heading levels — assert the deliberate
	// broadening so a future "fix" doesn't silently diverge the other way.
	if !emptyHeadingRe.MatchString(`<h2 class="a">   </h2>`) {
		t.Error("expected match on empty h2 with whitespace")
	}
	if !emptyHeadingRe.MatchString(`<h1></h3>`) {
		t.Error("expected match on mismatched empty heading pair (documented broadening)")
	}
	if emptyHeadingRe.MatchString(`<h2>Real title</h2>`) {
		t.Error("must not match a filled heading")
	}
}

// TestVerifyMissingComponentIsNotSuccess pins bugs_open/032. The verifier used to
// report Resolved:true when the page_components row was gone, on the assumption
// that a removed component cannot be empty. But a missing row is equally the
// signature of a rebuild silently deleting the component — so that branch
// recorded content-loss incidents as verified fixes, which is precisely the
// blind spot the completion gate exists to close.
//
// The contract now: never claim success from absence. Return an error, so the
// gate's fail-OPEN policy completes the item while recording under
// result._verification that verification could not be made. A false success
// becomes a visible unknown.
func TestVerifyMissingComponentIsNotSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	const componentID = "6f1b8a4e-6c2f-4a1e-9d3b-2f5c7a8e1b04"
	mock.ExpectQuery("SELECT rendered_html FROM page_components").
		WithArgs(componentID).
		WillReturnError(sql.ErrNoRows)

	res, err := VerifyEmptySectionResolved(
		context.Background(), db,
		VerifyTarget{Spec: map[string]interface{}{"component_id": componentID}},
		zap.NewNop(),
	)

	if err == nil {
		t.Fatal("a missing component must NOT verify as resolved: absence is equally deletion (bugs_open/032)")
	}
	if res.Resolved {
		t.Error("Resolved must be false when the component row is gone")
	}
	// The message has to say why it is unverifiable, or the recorded
	// _verification entry is as uninformative as the silent success it replaced.
	for _, want := range []string{"cannot verify", componentID, "indistinguishable"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must contain %q", err.Error(), want)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// A present-but-empty component must still fail closed — the ordinary path must
// not regress while fixing the absent one.
func TestVerifyPresentButEmptyStillFailsClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	const componentID = "6f1b8a4e-6c2f-4a1e-9d3b-2f5c7a8e1b04"
	mock.ExpectQuery("SELECT rendered_html FROM page_components").
		WithArgs(componentID).
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).AddRow(""))

	res, err := VerifyEmptySectionResolved(
		context.Background(), db,
		VerifyTarget{Spec: map[string]interface{}{"component_id": componentID}},
		zap.NewNop(),
	)
	if err != nil {
		t.Fatalf("a present component must verify without error: %v", err)
	}
	if res.Resolved {
		t.Error("an empty rendered_html must not verify as resolved")
	}
}

// ============================================================================
// Retraction guards (RFC_010 Decision 1 adoption)
// ============================================================================

const (
	retractPageID = "3f2a1c9e-5d4b-4e8a-9c1f-7b6d2e8a4f13"
	retractKey    = "empty_section:3f2a1c9e-5d4b-4e8a-9c1f-7b6d2e8a4f13:services"
)

// filledHTML passes emptySectionVerdict; hollowHTML does not.
var (
	filledHTML = `<section><p>` + strings.Repeat("real content ", 10) + `</p></section>`
	hollowHTML = `<div></div>`
)

// retractionRows builds the (item_key, component_id, rendered_html) result of
// the retraction observation query. A nil componentID models the LEFT JOIN miss.
func retractionRows(components ...interface{}) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"item_key", "id", "rendered_html"})
	if len(components) == 0 {
		return r
	}
	for i := 0; i < len(components); i += 2 {
		r.AddRow(retractKey, components[i], components[i+1])
	}
	return r
}

func runCheckWithNoFindings(t *testing.T, retraction *sqlmock.Rows) (*CheckResult, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	// No empty sections at all — the case the old early return swallowed.
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(sqlmock.NewRows([]string{
		"id", "page_id", "name", "slot_name", "function", "len", "pattern", "runtime_fill"}))
	mock.ExpectQuery("FROM site_work_items").WillReturnRows(retraction)

	check := &EmptySectionsCheck{}
	res, err := check.Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(), Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return res, mock, func() { db.Close() }
}

// The whole point: a site with ZERO empty sections is the site whose stale
// items most need closing. If someone reinstates the `len(sections) == 0`
// early return, this is the test that fails.
func TestRetractionRunsWhenThereAreNoFindingsAtAll(t *testing.T) {
	res, mock, done := runCheckWithNoFindings(t, retractionRows("c1", filledHTML))
	defer done()

	if len(res.Findings) != 0 || len(res.WorkItems) != 0 {
		t.Fatalf("expected no findings/work items, got %d/%d", len(res.Findings), len(res.WorkItems))
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("a healthy slot must retract even when the run found nothing: got %d resolved", len(res.Resolved))
	}
	if res.Resolved[0].ItemKey != retractKey {
		t.Errorf("ItemKey = %q, want %q", res.Resolved[0].ItemKey, retractKey)
	}
	if res.Resolved[0].ItemType != "empty_section" {
		t.Errorf("ItemType = %q, want empty_section", res.Resolved[0].ItemType)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The three refusals. Each models a population measured in the live queue on
// 2026-08-03; together they are 30 of 47 open items.
func TestRetractionRefusesEverythingItHasNotPositivelyObserved(t *testing.T) {
	cases := []struct {
		name        string
		rows        *sqlmock.Rows
		wantResolve bool
		why         string
	}{
		{
			name:        "slot holds no deployed component",
			rows:        retractionRows(nil, ""),
			wantResolve: false,
			why: "absence is equally a fix and a silently deleted component " +
				"(bugs_open/032) — 10 of 47 open items sit in this state",
		},
		{
			name:        "section still renders empty",
			rows:        retractionRows("c1", hollowHTML),
			wantResolve: false,
			why:         "the finding still reproduces",
		},
		{
			name:        "mixed slot: one of three still empty",
			rows:        retractionRows("c1", filledHTML, "c2", hollowHTML, "c3", filledHTML),
			wantResolve: false,
			why: "bugs_open/156 means (page_id, slot_name) is not unique; the " +
				"finding may still be true of the empty one",
		},
		{
			name:        "every component in a multi-component slot renders content",
			rows:        retractionRows("c1", filledHTML, "c2", filledHTML, "c3", filledHTML),
			wantResolve: true,
			why:         "nothing in the slot reproduces the finding",
		},
		{
			name:        "runtime-fill shell is healthy by design",
			rows:        retractionRows("c1", `<div data-runtime-fill="lobby"><h2></h2></div>`),
			wantResolve: true,
			why:         "the completion gate's own verdict function calls this resolved",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, _, done := runCheckWithNoFindings(t, tc.rows)
			defer done()
			got := len(res.Resolved) > 0
			if got != tc.wantResolve {
				t.Errorf("retracted = %v, want %v — %s", got, tc.wantResolve, tc.why)
			}
		})
	}
}

// AllOfType is the wide, destructive branch. This check observes one slot at a
// time and must never reach for it; a reviewer of a future edit should see this
// fail rather than discover the breadth in production.
func TestRetractionNeverUsesTheWideBranch(t *testing.T) {
	res, _, done := runCheckWithNoFindings(t, retractionRows("c1", filledHTML))
	defer done()

	if len(res.Resolved) == 0 {
		t.Fatal("expected a retraction to inspect")
	}
	for _, r := range res.Resolved {
		if r.AllOfType {
			t.Error("empty_section retraction must address one item_key, never AllOfType")
		}
		if r.ItemKey == "" {
			t.Error("ItemKey must be set — an empty key with AllOfType false is refused by the runner")
		}
		if r.Reason == "" {
			t.Error("Reason must be set — the runner refuses a retraction with no stated cause")
		}
	}
}

// A retraction fault must cost a missed CLOSURE, never a missed DEFECT. The
// runner skips a check's inserts when Run returns an error, so propagating this
// would trade the wrong way round.
func TestRetractionFailureDoesNotSuppressFindings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(
		sqlmock.NewRows([]string{"id", "page_id", "name", "slot_name", "function", "len", "pattern", "runtime_fill"}).
			AddRow(uuid.New().String(), pageID.String(), "services", "services", "services", 0, "empty_html", false))
	mock.ExpectQuery("FROM site_work_items").WillReturnError(sql.ErrConnDone)

	check := &EmptySectionsCheck{}
	res, err := check.Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(), Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("a retraction fault must not fail the check: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Errorf("the empty section must still be filed: got %d work items", len(res.WorkItems))
	}
	if len(res.Resolved) != 0 {
		t.Errorf("a failed observation must retract nothing: got %d", len(res.Resolved))
	}
}

// The retraction resolves page_id from the FIRST-CLASS COLUMN, falling back to
// the spec. Council 97923026, editquality (medium): a handler blind to the
// column the creator populates is a standing landmine on this table. Measured
// 58/58 agreement today, so this pins the PREFERENCE rather than a behaviour
// difference — if someone drops the COALESCE back to spec-only, the query text
// stops naming the column and this fails.
func TestRetractionPrefersTheFirstClassPageIDColumn(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM page_components pc").WillReturnRows(sqlmock.NewRows([]string{
		"id", "page_id", "name", "slot_name", "function", "len", "pattern", "runtime_fill"}))
	// The column must appear, coalesced ahead of the spec key.
	mock.ExpectQuery(`COALESCE\(page_id::text, spec->>'page_id'\)`).
		WillReturnRows(retractionRows("c1", filledHTML))

	check := &EmptySectionsCheck{}
	if _, err := check.Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(), Logger: zap.NewNop(),
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the retraction query must resolve page_id from the first-class column: %v", err)
	}
}
