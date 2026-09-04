// FILE: platform/orchestration/actions/discovery_checks/check_empty_blog_test.go
//
// Settles the question bugs_open/460 left open: has BlogEmptyCheck STOPPED
// WORKING, or has it simply had nothing to file?
//
// 460 records that blog-content-planner ran 13 times and went silent on
// 2026-04-24, on two independent instruments, with the cause unestablished.
// This check is that agent's only driver. [MEASURED 2026-09-04, using the
// check's OWN predicate below, not an approximation of it] its selectable
// population is 4 sites with a blog hub, of which 0 have zero posts — so 0
// would fire today. The listing hubs this gate cannot see at all:
// section-index 62, news-index 11, entity-directory 8, against blog-index 4.
//
// A check that files nothing because its population is empty and a check that
// files nothing because it is broken are indistinguishable in production, and
// that ambiguity is what made 460 read as a fault. These tests remove it from
// the "broken" side: if the mechanism still fires on a qualifying site, then
// silence is population, and reviving the producer is not the question.
//
// They also PIN the swallowed-error behaviour at line 35-38 rather than leave
// it as folklore — see TestBlogEmptyCheck_QueryErrorIsIndistinguishableFromNoBlogPage.

package discovery_checks

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func blogCheckCtx(t *testing.T) (dctx DiscoveryCheckContext, mock sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(),
		Pipeline: "content", AgentType: "test", BatchID: uuid.New(),
		Logger: zap.NewNop(),
	}, mock
}

// THE LOAD-BEARING TEST. A site with a blog hub and zero posts must produce
// exactly one needs_blog_posts item routed at blog-content-planner. If this
// passes, the mechanism is alive and 460's silence is an empty population.
func TestBlogEmptyCheck_FilesWorkItemWhenHubHasNoPosts(t *testing.T) {
	dctx, mock := blogCheckCtx(t)
	blogPageID := uuid.New()

	mock.ExpectQuery("id::text").WithArgs(dctx.SiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(blogPageID.String(), "blog"))
	mock.ExpectQuery("COUNT").WithArgs(dctx.SiteID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	res, err := (&BlogEmptyCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 work item, got %d: %+v", len(res.WorkItems), res.WorkItems)
	}

	wi := res.WorkItems[0]
	if wi.ItemType != "needs_blog_posts" {
		t.Errorf("item_type = %q, want needs_blog_posts", wi.ItemType)
	}
	// The routing is the half that matters to 460: this is what names the
	// dormant agent as the handler.
	if wi.HandlerAgent != "blog-content-planner" {
		t.Errorf("handler_agent = %q, want blog-content-planner", wi.HandlerAgent)
	}
	if wi.ItemKey != "empty_blog:"+dctx.SiteID.String() {
		t.Errorf("item_key = %q, want the site-scoped empty_blog key", wi.ItemKey)
	}
	if wi.Status != "detected" || wi.Pipeline != "content" {
		t.Errorf("wrong shape: status=%q pipeline=%q", wi.Status, wi.Pipeline)
	}
	if wi.PageID == nil || *wi.PageID != blogPageID {
		t.Errorf("page_id = %v, want the blog hub's id %v", wi.PageID, blogPageID)
	}
	if len(res.Findings) != 1 {
		t.Errorf("want 1 finding alongside the work item, got %d", len(res.Findings))
	}
}

// The negative that makes the positive mean something: a populated blog must
// file nothing. Without this, the test above would pass on a check that filed
// unconditionally.
func TestBlogEmptyCheck_SilentWhenHubAlreadyHasPosts(t *testing.T) {
	dctx, mock := blogCheckCtx(t)

	mock.ExpectQuery("id::text").WithArgs(dctx.SiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(uuid.New().String(), "blog"))
	mock.ExpectQuery("COUNT").WithArgs(dctx.SiteID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	res, err := (&BlogEmptyCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("a populated blog must file nothing, got %d: %+v", len(res.WorkItems), res.WorkItems)
	}
}

// This is the production case for all 4 blog-hub sites as of 2026-09-04 and
// for every site that has no blog hub at all — i.e. the overwhelming majority.
func TestBlogEmptyCheck_SilentWhenSiteHasNoBlogHub(t *testing.T) {
	dctx, mock := blogCheckCtx(t)

	mock.ExpectQuery("id::text").WithArgs(dctx.SiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"})) // no rows

	res, err := (&BlogEmptyCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("no blog hub must file nothing, got %d", len(res.WorkItems))
	}
}

// PINS A REAL DEFECT rather than asserting it is fine. check_empty_blog.go:35-38
// returns an empty result on ANY error from the first query, not just
// sql.ErrNoRows — so a genuine database failure is indistinguishable from "this
// site has no blog page", and the check reports success either way.
//
// This is why "the check filed nothing" could never have been read as evidence
// about the population, and it is a second, independent reason 460's silence was
// ambiguous. The test asserts today's behaviour so the ambiguity is visible and
// a fix has something to change; it is NOT an endorsement. Whoever narrows this
// to sql.ErrNoRows should expect this test to fail and should update it.
func TestBlogEmptyCheck_QueryErrorIsIndistinguishableFromNoBlogPage(t *testing.T) {
	dctx, mock := blogCheckCtx(t)

	mock.ExpectQuery("id::text").WithArgs(dctx.SiteID).
		WillReturnError(errors.New("connection reset by peer"))

	res, err := (&BlogEmptyCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("current behaviour swallows the error; got err = %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Findings) != 0 {
		t.Fatalf("current behaviour returns an empty result; got %+v", res)
	}
	// If this ever starts failing, the swallow has been narrowed — that is an
	// improvement, and this test should be rewritten to assert the error is
	// surfaced rather than deleted.
}
