// FILE: platform/orchestration/actions/discovery_checks/check_missing_tools_test.go
//
// The content-to-tools ratio (growth_config.content_tools_ratio) exists because the
// original cooldown asks "has it been a while?" and never "has this site outgrown its
// tools?" — a site that published thirty guides in a fortnight looked identical to a
// dormant one, and both waited 30 days.
//
// The test that matters most is the OFF one: the key is opt-in, and every site in the
// estate has no growth_config row or no such key. If absence changed behaviour, this
// change would silently re-evaluate the whole fleet.
package discovery_checks

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// expectToolCount stubs step 1 (deployed tool count).
func expectToolCount(mock sqlmock.Sqlmock, siteID uuid.UUID, n int) {
	mock.ExpectQuery(`component_level = 'tool'`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(n))
}

// expectRatio stubs step 2's growth_config read. Pass -1 to simulate "no row",
// which is what every site in the estate looks like today.
func expectRatio(mock sqlmock.Sqlmock, siteID uuid.UUID, ratio int) {
	q := mock.ExpectQuery(`content_tools_ratio`).WithArgs(siteID)
	if ratio < 0 {
		q.WillReturnError(sql.ErrNoRows)
		return
	}
	q.WillReturnRows(sqlmock.NewRows([]string{"ratio"}).AddRow(ratio))
}

func expectArticleCount(mock sqlmock.Sqlmock, siteID uuid.UUID, n int) {
	mock.ExpectQuery(`page_type = ANY`).
		WithArgs(siteID, pq.Array(articlePageTypes)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(n))
}

// The first version of this change counted only ('blog-post','content'). The
// council gate's own read-only check found `guide` is a separate, real page_type
// — 52 deployed pages across 5 sites, the third-largest article-shaped type — so
// every one of them was excluded and those sites would have counted zero
// articles and never tripped the ratio.
//
// The failure mode is SILENT: an omitted type does not error, it just makes a
// site look emptier than it is. This test is the tripwire — a new article-shaped
// page_type must be added to articlePageTypes deliberately, and the excluded
// list is asserted too so that "tool" can never be counted as content the ratio
// is measured against (a site could otherwise satisfy its tool ratio by
// publishing tools).
func TestArticlePageTypes_CoversTheArticleShapedTypes(t *testing.T) {
	got := map[string]bool{}
	for _, pt := range articlePageTypes {
		got[pt] = true
	}
	for _, want := range []string{"blog-post", "guide", "content"} {
		if !got[want] {
			t.Errorf("articlePageTypes is missing %q — a site whose written content uses that type counts as empty and is never asked for tools", want)
		}
	}
	for _, mustNot := range []string{"tool", "game", "landing", "section-index", "blog-index", "news-index", "entity-directory"} {
		if got[mustNot] {
			t.Errorf("articlePageTypes must not contain %q: tools/games are what the ratio counts AGAINST, and the index types are navigational shells, not reading", mustNot)
		}
	}
}

// expectCooldown stubs step 3 and pins the cooldown the check chose — the whole
// point of the ratio is which number lands here.
func expectCooldown(mock sqlmock.Sqlmock, siteID uuid.UUID, days int, recent bool) {
	mock.ExpectQuery(`item_type = 'evaluate_tools'`).
		WithArgs(siteID, days).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(recent))
}

func runCheck(t *testing.T, db *sql.DB, siteID uuid.UUID) *CheckResult {
	t.Helper()
	res, err := (&MissingToolsCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return res
}

// No growth_config row — the state of every site in the estate right now. The
// article-count query must not run at all, and a site with tools must keep its
// 30-day cooldown.
func TestMissingTools_RatioAbsent_LeavesBehaviourUnchanged(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	expectToolCount(mock, siteID, 2)
	expectRatio(mock, siteID, -1) // no row
	// deliberately NO expectArticleCount: querying it would be the regression
	expectCooldown(mock, siteID, 30, true)

	res := runCheck(t, db, siteID)
	if len(res.WorkItems) != 0 {
		t.Fatalf("got %d work items, want 0 (recent evaluation exists)", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Ratio configured and the site is behind it: 18 published articles at 1-per-6
// implies 3 tools, only 1 is deployed. The cooldown must drop to 7 even though
// the site is not tool-less.
func TestMissingTools_BehindRatio_ShortensCooldownAndExplainsWhy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	expectToolCount(mock, siteID, 1)
	expectRatio(mock, siteID, 6)
	expectArticleCount(mock, siteID, 18)
	expectCooldown(mock, siteID, 7, false) // 7, not 30 — this is the behaviour change

	res := runCheck(t, db, siteID)
	if len(res.WorkItems) != 1 {
		t.Fatalf("got %d work items, want 1", len(res.WorkItems))
	}
	item := res.WorkItems[0]
	if item.Status != "detected" {
		t.Errorf("status = %q, want %q — routing is bugs_open/083's open decision, not this change's", item.Status, "detected")
	}
	for _, want := range []string{`"tools_expected":3`, `"tools_short":2`, `"content_tools_ratio":6`, `"published_articles":18`} {
		if !contains(item.SpecJSON, want) {
			t.Errorf("spec %s missing %s", item.SpecJSON, want)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Ratio configured and the site is AT or ahead of it: nothing changes, 30 days.
// Integer division is deliberate — 17 articles at 1-per-6 expects 2, not 2.83.
func TestMissingTools_MeetingRatio_KeepsLongCooldown(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	expectToolCount(mock, siteID, 2)
	expectRatio(mock, siteID, 6)
	expectArticleCount(mock, siteID, 17) // 17/6 = 2 expected, 2 deployed => not behind
	expectCooldown(mock, siteID, 30, true)

	res := runCheck(t, db, siteID)
	if len(res.WorkItems) != 0 {
		t.Fatalf("got %d work items, want 0", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// A tool-less site was already on 7 days. The ratio must not make that worse or
// change the message when it is not the reason.
func TestMissingTools_NoTools_UnaffectedByRatio(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	expectToolCount(mock, siteID, 0)
	expectRatio(mock, siteID, 0) // configured off explicitly
	expectCooldown(mock, siteID, 7, false)

	res := runCheck(t, db, siteID)
	if len(res.WorkItems) != 1 {
		t.Fatalf("got %d work items, want 1", len(res.WorkItems))
	}
	if contains(res.WorkItems[0].SpecJSON, "content_tools_ratio") {
		t.Errorf("spec %s should carry no ratio fields when the ratio is off", res.WorkItems[0].SpecJSON)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
