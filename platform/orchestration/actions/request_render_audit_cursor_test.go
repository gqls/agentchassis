// FILE: platform/orchestration/actions/request_render_audit_cursor_test.go
//
// Tests for the coverage cursor (bugs_open/394).
//
// EVERY GUARD HERE IS PROVEN BY INDUCING THE FAULT, never by observing that
// nothing went wrong. Each test names the mutation that must make it fail, and
// the two that matter most are the ones a natural suite omits:
//
//  1. A priority page must NOT advance the cursor. If it does, the boundary
//     jumps past pages that were never in any window and they are audited by
//     nothing — a coverage hole created by the fix for a coverage hole. Every
//     single-run assertion passes either way; only a two-run union catches it.
//
//  2. The priority_* context keys must be present when ZERO. A key that appears
//     only on a bad run makes its absence ambiguous between "none happened" and
//     "the binary is too old to count it", which is the same class as the
//     LANDMINES entry about a zero-findings early return: green in every test
//     you would write, wrong in production.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// mkAuditRows builds n pages in the audit's own ordering: p000..p{n-1} at
// nav_order 100, which is the shape of the real motivating site (webdesign.co.uk
// has 94 tools at nav_order 100, ordered alphabetically by name).
func mkAuditRows(n int) []auditPageRow {
	out := make([]auditPageRow, 0, n)
	for i := 0; i < n; i++ {
		name := pageName(i)
		out = append(out, auditPageRow{
			URL:  "https://example.com/" + name + ".html",
			Path: "/" + name + ".html",
			Ord:  100,
			Name: name,
		})
	}
	return out
}

func pageName(i int) string {
	const digits = "0123456789"
	return "p" + string([]byte{digits[(i/100)%10], digits[(i/10)%10], digits[i%10]})
}

func firstName(rows []auditPageRow) string {
	if len(rows) == 0 {
		return ""
	}
	return rows[0].Name
}

func TestSelectAuditWindow_Table(t *testing.T) {
	rows := mkAuditRows(200)

	cases := []struct {
		name      string
		cur       *auditCursor
		n         int
		wantFirst string
		wantLen   int
		wantNext  string // "" means the cursor must be CLEARED
	}{
		// MUTATION: off-by-one the start index → wantFirst fails.
		{"no cursor takes the prefix", nil, 60, "p000", 60, "p059"},
		// MUTATION: use >= instead of > at the boundary → the cursor page repeats.
		{"mid cycle starts at the successor", &auditCursor{100, "p059"}, 60, "p060", 60, "p119"},
		// MUTATION: keep advancing instead of clearing → wantNext is non-empty and this fails.
		{"final window clears the cursor", &auditCursor{100, "p150"}, 60, "p151", 49, ""},
		// MUTATION: fall back to index 0 on an exact-match miss → the window restarts at p000.
		{"deleted cursor page continues, never restarts", &auditCursor{100, "p059x"}, 60, "p060", 60, "p119"},
		// MUTATION: return an empty window → upstream reads it as no_deployed_pages, a FALSE SKIP.
		{"cursor past the end restarts this run", &auditCursor{999, "zzz"}, 60, "p000", 60, "p059"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, next := selectAuditWindow(rows, c.cur, c.n, nil)
			if firstName(got) != c.wantFirst {
				t.Fatalf("window starts at %q, want %q", firstName(got), c.wantFirst)
			}
			if len(got) != c.wantLen {
				t.Fatalf("window length %d, want %d", len(got), c.wantLen)
			}
			if c.wantNext == "" {
				if next != nil {
					t.Fatalf("cursor should be CLEARED at the end of a cycle, got %+v", next)
				}
				return
			}
			if next == nil {
				t.Fatalf("cursor cleared mid-cycle — the next run would restart from the top")
			}
			if next.Name != c.wantNext {
				t.Fatalf("next cursor %q, want %q", next.Name, c.wantNext)
			}
		})
	}
}

// The union property, which is the whole point: consecutive windows must cover
// every page with no gap and no repeat.
//
// MUTATION: let the cursor advance by the window LENGTH rather than to the last
// page actually taken, and the skipped pages show up here as a gap.
func TestSelectAuditWindow_ConsecutiveRunsCoverEveryPageExactlyOnce(t *testing.T) {
	rows := mkAuditRows(146) // webdesign.co.uk as measured 2026-08-26
	seen := map[string]int{}
	var cur *auditCursor
	for run := 0; run < 3; run++ {
		win, next := selectAuditWindow(rows, cur, 60, nil)
		for _, r := range win {
			seen[r.Name]++
		}
		cur = next
		if cur == nil {
			break
		}
	}
	if len(seen) != len(rows) {
		t.Fatalf("union covered %d of %d pages — a cursor that does not converge is not a fix", len(seen), len(rows))
	}
	for name, n := range seen {
		if n != 1 {
			t.Fatalf("page %s audited %d times in one cycle — the window overlaps", name, n)
		}
	}
	if cur != nil {
		t.Fatalf("cycle did not clear its cursor after covering the site")
	}
}

// A priority page inside the scan range must be sent once, by the priority set,
// and must not consume a rotation slot twice.
//
// MUTATION: remove the `skip` check in the fill loop → the page appears in both
// the priority set and the rotation window, so the produced request carries a
// duplicate URL.
func TestSelectAuditWindow_SkipsPagesAlreadyInThePrioritySet(t *testing.T) {
	rows := mkAuditRows(200)
	skip := map[string]bool{"/p065.html": true}
	win, _ := selectAuditWindow(rows, &auditCursor{100, "p059"}, 60, skip)
	for _, r := range win {
		if r.Path == "/p065.html" {
			t.Fatalf("a page already carried by the priority set was also placed in the rotation window")
		}
	}
	if len(win) != 60 {
		t.Fatalf("skipping a page must not shrink the window: got %d, want 60", len(win))
	}
}

// THE ONE THAT ONLY A TWO-RUN ASSERTION CATCHES.
//
// A priority page far down the ordering must not drag the cursor to it. If it
// does, every page between the rotation window and that priority page is skipped
// by this run AND by every future run, because the boundary has moved past them.
//
// MUTATION: set the cursor from the last element of (priority ++ window) rather
// than from the rotation slice → run 2 starts past p140 and pages p118..p139 are
// audited by nothing. The gap assertion below is what fails.
func TestPriorityPagesDoNotAdvanceTheCursor(t *testing.T) {
	rows := mkAuditRows(200)
	skip := map[string]bool{"/p140.html": true} // priority page, far past the window

	win, next := selectAuditWindow(rows, &auditCursor{100, "p060"}, 57, skip)
	if next == nil {
		t.Fatalf("cursor cleared mid-cycle")
	}
	if next.Name != "p117" {
		t.Fatalf("cursor is at %q — a priority page beyond the window moved the boundary; "+
			"pages between the window and it would never be audited", next.Name)
	}

	// And prove the consequence directly: the next run must start at p118.
	win2, _ := selectAuditWindow(rows, next, 57, skip)
	if firstName(win2) != "p118" {
		t.Fatalf("next run starts at %q, want p118 — a gap has opened", firstName(win2))
	}
	_ = win
}

func TestCyclicFrom_RotatesSoAnOverBudgetDropIsNotAlwaysTheSamePages(t *testing.T) {
	rows := mkAuditRows(10)
	got := cyclicFrom(rows, &auditCursor{100, "p003"})
	if firstName(got) != "p004" {
		t.Fatalf("cyclic order starts at %q, want p004", firstName(got))
	}
	if len(got) != len(rows) {
		t.Fatalf("cyclic order dropped rows: %d of %d", len(got), len(rows))
	}
	// MUTATION: return rows unrotated → the same pages are dropped every run,
	// which is the deterministic-prefix disease reappearing inside its own fix.
	if firstName(cyclicFrom(rows, nil)) != "p000" {
		t.Fatalf("a nil cursor must leave the order alone")
	}
}

// The priority_* keys must be present when zero. See the file header.
//
// MUTATION: emit the keys only when non-zero → this fails, and in production the
// absence of a key becomes indistinguishable from an old binary.
func TestTruncationContext_ZeroPriorityKeysArePresentNotOmitted(t *testing.T) {
	window := mkAuditRows(3)
	ctx := truncationContext(true, 3, 146, 60, auditedPaths(nil, window), &auditCursor{100, "p002"}, window, priorityResult{})
	for _, k := range []string{"priority_paths", "priority_open_items", "priority_dropped", "priority_not_live", "coverage_mode", "window_first", "window_last", "cursor_cleared", "audited_paths"} {
		if _, ok := ctx[k]; !ok {
			t.Fatalf("cursor-mode context omits %q — absence must never be ambiguous with zero", k)
		}
	}
	if ctx["coverage_mode"] != "cursor" {
		t.Fatalf("coverage_mode = %v, want cursor", ctx["coverage_mode"])
	}
	// The three original keys keep their names and meanings, so existing
	// consumers are unaffected. MUTATION: rename any of them → this fails.
	for _, k := range []string{"pages_total", "pages_audited", "max_pages"} {
		if _, ok := ctx[k]; !ok {
			t.Fatalf("cursor-mode context dropped the pre-existing key %q", k)
		}
	}
}

// Prefix mode must be reported as prefix and must NOT carry cursor keys — a
// prefix row claiming a coverage window would be a false assertion, and the
// commissioned reader alarms on exactly this distinction.
func TestTruncationContext_PrefixModeCarriesNoCursorKeys(t *testing.T) {
	ctx := truncationContext(false, 2, 3, 2, nil, nil, nil, priorityResult{})
	if ctx["coverage_mode"] != "prefix" {
		t.Fatalf("coverage_mode = %v, want prefix", ctx["coverage_mode"])
	}
	for _, k := range []string{"window_first", "priority_paths", "audited_paths"} {
		if _, ok := ctx[k]; ok {
			t.Fatalf("prefix-mode context carries %q — it has no window to describe", k)
		}
	}
}

// The message must stop asserting the property the code no longer has — and must
// KEEP asserting it in the mode where it is still true.
func TestTruncationMessage_ModeSplit(t *testing.T) {
	const same = "the unaudited tail is the SAME pages every run"
	prefix := truncationMessage(false, 60, 146, "webdesign.co.uk", priorityResult{}, 0)
	if !strings.Contains(prefix, same) {
		t.Fatalf("prefix mode dropped a sentence that is still literally true: %q", prefix)
	}
	cursor := truncationMessage(true, 60, 146, "webdesign.co.uk", priorityResult{taken: mkAuditRows(3)}, 57)
	if strings.Contains(cursor, same) {
		t.Fatalf("cursor mode asserts a property the code no longer has: %q", cursor)
	}
	if !strings.Contains(cursor, "~3 runs") {
		t.Fatalf("cursor mode should estimate the cycle from the ROTATION size, not max_pages: %q", cursor)
	}
}

// A site that fits inside the cap must never touch the cursor table: no read, no
// write, and no truncation row. Proven by registering NO cursor expectations on
// the mock — an unexpected query fails the test.
//
// MUTATION: move the cursor read above the `total <= maxPages` guard → sqlmock
// reports a call it did not expect.
func TestRotateCoverage_SiteWithinCapNeverTouchesTheCursor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New().String()

	mock.ExpectQuery("FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"url", "ord", "name"}).
			AddRow("/index.html", 0, "index").
			AddRow("/about.html", 10, "about"))
	// Deliberately NO ExpectQuery for render_audit_page_cursor and no ExpectExec.

	core, logs := observer.New(zap.WarnLevel)
	producer := &capturingProducer{}
	params := renderAuditParams(ActionParams{
		DB:       db,
		Producer: producer,
		Logger:   zap.New(core),
	}, siteID)
	params.StepConfig = models.Step{Config: map[string]interface{}{
		"max_pages": 60, "rotate_coverage": true,
	}}

	if _, err := RequestRenderAuditAction(context.Background(), params); err != nil {
		t.Fatalf("action failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet or unexpected DB calls: %v", err)
	}
	for _, e := range logs.All() {
		if strings.Contains(e.Message, "TRUNCATED") {
			t.Fatalf("a site inside the cap must not report truncation: %q", e.Message)
		}
	}
}

// THE CALL-SITE TEST, and the reason it exists.
//
// The pure helper cannot pin this property. selectAuditWindow never sees the
// priority pages except as a `skip` set, so a unit test of it will agree with
// itself whatever the action does with the union. The property that matters —
// "the STORED cursor is the rotation window's last page, never a priority
// page" — is a property of the assembly at the call site, so it has to be
// asserted against the actual UPSERT.
//
// A first attempt at this mutated `out[len(out)-1]` to `rows[i-1]` inside the
// helper and the test still passed, because on that fixture the two are the same
// value. That was a VACUOUS MUTATION: the guard looked proven and was not.
// Recorded in WRONG_CALLS 2026-08-26.
//
// MUTATION that must fail this test: at the call site, write the cursor from the
// concatenated list — `saveAuditCursor(..., &auditCursor{Ord: last(priority ++
// window).Ord, Name: ...})`. The stored name becomes p140 and the WithArgs
// expectation below goes unmet.
func TestPriorityPageBeyondTheWindowDoesNotMoveTheStoredCursor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New().String()

	// 200 live pages at nav_order 100. The cap is 4, so the reserve is 2 and the
	// rotation gets 3 once one priority page is taken.
	pages := sqlmock.NewRows([]string{"url", "ord", "name"})
	for i := 0; i < 200; i++ {
		n := pageName(i)
		pages.AddRow("/"+n+".html", 100, n)
	}
	mock.ExpectQuery("FROM pages").WillReturnRows(pages)

	// Mid-cycle: the last run stopped at p060.
	mock.ExpectQuery("FROM render_audit_page_cursor").
		WillReturnRows(sqlmock.NewRows([]string{"after_nav_order", "after_name"}).AddRow(100, "p060"))

	// One open contrast_failure, on a page FAR past the window (p140).
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}).
			AddRow("contrast_failure:/p140.html#BUTTON.btn"))

	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(0, 1))

	// THE ASSERTION. The rotation slice is p061..p063 (3 slots, cap 4 minus the
	// one priority page), so the stored cursor must be p063 — NOT p140.
	mock.ExpectExec("INSERT INTO render_audit_page_cursor").
		WithArgs(siteID, "render-audit-agent", 100, "p063").
		WillReturnResult(sqlmock.NewResult(0, 1))

	producer := &capturingProducer{}
	params := renderAuditParams(ActionParams{
		DB:       db,
		Producer: producer,
		Logger:   zap.NewNop(),
	}, siteID)
	params.StepConfig = models.Step{Config: map[string]interface{}{
		"max_pages": 4, "rotate_coverage": true,
	}}

	if _, err := RequestRenderAuditAction(context.Background(), params); err != nil {
		t.Fatalf("action failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("cursor was not stored at the rotation window's last page: %v", err)
	}

	// And the priority page really was sent — otherwise this test would pass by
	// the priority set simply not working, which is the opposite defect.
	if !strings.Contains(string(producer.value), "/p140.html") {
		t.Fatalf("the finding-bearing page was not carried in the run at all")
	}
	// It must be FIRST, so a run the adapter abandons part-way still measures it.
	body := string(producer.value)
	if strings.Index(body, "/p140.html") > strings.Index(body, "/p061.html") {
		t.Fatalf("priority pages must be sent ahead of the rotation slice")
	}
}

// THE COLLISION THE COUNCIL FOUND (editquality, round 1, corr f67593f5).
//
// The first cut recovered the page path by splitting a contrast_failure key on
// its first '#'. Both halves of the hazard are LIVE, measured 2026-08-26:
//
//   - a selector may contain '#' — `…/index.html#BUTTON#c-tool-…` exists in
//     production (1 of 469 open rows), and the `describe` scheme emits
//     `tag#id.classes` by construction;
//   - a PAGE URL may contain '#' — idea.uk carries BOTH `/tools.html` and
//     `/tools.html#audience-check` as ACTIVE pages, with 35 open contrast rows.
//
// So the split turns a finding on `/tools.html#audience-check` into one on
// `/tools.html`, which is a REAL page on that site: the wrong page is
// prioritised, and nothing errors. This fixture is idea.uk's real shape.
//
// MUTATION: reinstate a first-'#' split to derive the path → `/tools.html` is
// selected instead of `/tools.html#audience-check` and this test fails.
func TestPriorityMatchIsNotFooledByAHashInThePageURLOrTheSelector(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New().String()

	live := []auditPageRow{
		{Path: "/tools.html", Ord: 10, Name: "tools"},
		{Path: "/tools.html#audience-check", Ord: 20, Name: "tools-audience-check"},
		{Path: "/tools/sfi26-revenue-stacker/index.html", Ord: 100, Name: "sfi26"},
	}

	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}).
			// a finding on the FRAGMENT page — must not be attributed to /tools.html
			AddRow("contrast_failure:/tools.html#audience-check#SPAN.eyebrow").
			// a selector that itself contains '#' — the live shape
			AddRow("contrast_failure:/tools/sfi26-revenue-stacker/index.html#BUTTON#c-tool-calc-btn").
			// an open row whose page is not live at all
			AddRow("contrast_failure:/retired.html#P.gone"))

	hit, total, notLive, err := pagesWithOpenContrastFindings(context.Background(), db, siteID, live)
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if total != 3 {
		t.Fatalf("open row count %d, want 3", total)
	}
	if !hit["/tools.html#audience-check"] {
		t.Fatalf("a finding on the fragment page was not attributed to it")
	}
	if hit["/tools.html"] {
		t.Fatalf("a finding on /tools.html#audience-check was mis-attributed to /tools.html — "+
			"this is the exact live collision on idea.uk, and it is silent: %v", hit)
	}
	if !hit["/tools/sfi26-revenue-stacker/index.html"] {
		t.Fatalf("a selector containing '#' broke the match")
	}
	if notLive != 1 {
		t.Fatalf("notLive = %d, want 1 (the retired page's row can never self-grade)", notLive)
	}
}

// The prefix this matcher builds must be the SAME string the grader builds, or
// the priority set and the retraction set are talking about different rows.
// Built by the shared composer on both sides, so this pins the composition
// rather than a copy of it.
//
// MUTATION: change the delimiter or the prefix spelling in workItemKey → fails
// here AND on the grading side, which is the point.
func TestPriorityPrefixIsTheGradersOwnComposition(t *testing.T) {
	const path = "/index.html"
	want := workItemKey("contrast_failure", path+"#") // exactly line 748's construction
	if want != "contrast_failure:/index.html#" {
		t.Fatalf("composer produced %q — the grader builds its audited-page prefix the same way", want)
	}
}
