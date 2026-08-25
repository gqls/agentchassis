// FILE: platform/orchestration/actions/rebuild_blog_listing_image_test.go
//
// The blog listing's item image must come from the SHARED projection, not from
// a hand-written blank (bugs_open/384 decision 3, 2026-08-25).
//
// THE DEFECT THESE GUARD. This action wrote `"image": ""` for every article it
// listed. That made it a SECOND WRITER of `content_data.articles` on a field
// the 384 seam exists to keep correct: leopardessconsulting.co.uk's blog page
// carries `blog-listing_pre_037`, which declares `articles` ← `query.blog_posts`
// AND renders `.image`, so it IS a 384 consumer. A card landing there makes the
// seam re-resolve the array with the real image; the next `rerender-pages` run
// blanked it again. Last writer won, and neither writer logged anything.
//
// ⚠ WHY THERE IS NO SOURCE-SCAN TEST HERE. The obvious cheap guard — grep the
// action's source for `"image": ""` and fail — would pass VACUOUSLY: the
// explanatory comment on blogPostsQuery quotes that exact literal while
// describing the defect, and a first-occurrence match finds the comment. A test
// whose needle matches your own prose cannot fail. So these assertions drive
// the real function with real rows instead.

package actions

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
)

// blogRowCols is blogPostsQuery's SELECT list in order. The last three come
// from queryresolve.PageImageProjectionSQL.
var blogRowCols = []string{
	"id", "name", "url", "title", "meta_description", "created_at", "content_length",
	"card_key", "hero_key", "hero_purpose",
}

// scanOneBlogRow drives the production scan with a single mock row.
func scanOneBlogRow(t *testing.T, cardKey, heroKey, heroPurpose string) map[string]interface{} {
	t.Helper()
	db, mock := newRetractMockDB(t)

	rows := sqlmock.NewRows(blogRowCols).AddRow(
		uuid.New(), "post-one", "/blog/post-one.html", "Post One | Acme Ltd",
		"A summary.", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), 4000,
		cardKey, heroKey, heroPurpose,
	)
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	got, err := db.Query(`SELECT`)
	if err != nil {
		t.Fatalf("mock query: %v", err)
	}
	defer got.Close()

	articles, err := scanBlogArticles(got, zap.NewNop())
	if err != nil {
		t.Fatalf("scanBlogArticles errored on a well-formed row: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("scanBlogArticles returned %d articles, want 1 — the Scan destination list no longer matches the projection's column count", len(articles))
	}
	return articles[0]
}

// The card crop wins when one exists. This is the assertion that goes RED if
// the projection is reverted to a literal "" — the mutation that motivated the
// whole fix.
func TestBlogListingUsesTheCardCropWhenOneExists(t *testing.T) {
	article := scanOneBlogRow(t, "card-post-one", "hero-post-one", "hero")

	image, _ := article["image"].(string)
	if image == "" {
		t.Fatalf("article image is empty though the row carries a card asset — the action is blanking the image the 384 seam repairs")
	}
	if !strings.Contains(image, "card-post-one") {
		t.Errorf("article image %q does not name the CARD key; card must win over hero", image)
	}
	if strings.Contains(image, "hero-post-one") {
		t.Errorf("article image %q used the hero while a card exists — preference order is card first", image)
	}
}

// The hero is the fallback, not a co-equal. Measured 2026-08-25: this is the
// branch loancalculator.co.uk takes on every listed tool page (0 of 10 have a
// card), so it is a live branch, not a theoretical one.
func TestBlogListingFallsBackToThePlanHero(t *testing.T) {
	article := scanOneBlogRow(t, "", "hero-post-one", "hero")

	image, _ := article["image"].(string)
	if image == "" {
		t.Fatalf("article image is empty though the row carries a plan hero — the fallback arm is not reached")
	}
	if !strings.Contains(image, "hero-post-one") {
		t.Errorf("article image %q does not name the hero key", image)
	}
}

// Neither candidate present must still yield "" — a listing item with no
// imagery is normal (42 of the fleet's tool-cta entries on 2026-08-25), and the
// templates gate on a falsy image. This is what makes the fix a no-op today:
// all 47 currently-listed articles take THIS branch.
func TestBlogListingYieldsEmptyImageWhenThePageHasNeither(t *testing.T) {
	article := scanOneBlogRow(t, "", "", "")

	if image, _ := article["image"].(string); image != "" {
		t.Errorf("article image is %q, want \"\" — a page with no card and no hero must project no image", image)
	}
}

// The scan contract, stated as its own failure. A projection that gains or
// reorders a column breaks Scan at RUNTIME, not compile time; scanBlogArticles
// logs and skips such a row, so the visible symptom is an EMPTY LISTING, which
// the caller then treats as "leave the existing listing alone" — a silent
// no-op. Pin the count so the drift fails here instead.
//
// ⚠ WHAT THIS ONE CANNOT SEE, measured by mutation 2026-08-25: delete the three
// image destinations from the Scan and this test still PASSES — a 9-column row
// into 7 destinations errors either way, so the skip it asserts happens for the
// wrong reason. The three tests above are what actually go red on that mutation
// (all three did). This test earns its place only for the OPPOSITE drift — a
// column ADDED to the projection without a matching destination — and the
// column count is the half of it that cannot pass vacuously.
func TestBlogListingScanContractMatchesTheProjection(t *testing.T) {
	if n := strings.Count(queryresolve.PageImageProjectionSQL, " AS "); n != 3 {
		t.Fatalf("PageImageProjectionSQL now yields %d columns, not 3 — scanBlogArticles' Scan destinations must change in the same commit", n)
	}
	// Drive it one column short: the row must be skipped, not silently accepted.
	db, mock := newRetractMockDB(t)
	short := blogRowCols[:len(blogRowCols)-1]
	mock.ExpectQuery(`SELECT`).WillReturnRows(
		sqlmock.NewRows(short).AddRow(
			uuid.New(), "post-one", "/blog/post-one.html", "Post One",
			"A summary.", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), 4000,
			"card-post-one", "hero-post-one",
		))
	got, err := db.Query(`SELECT`)
	if err != nil {
		t.Fatalf("mock query: %v", err)
	}
	defer got.Close()

	articles, err := scanBlogArticles(got, zap.NewNop())
	if len(articles) != 0 {
		t.Errorf("a row with too few columns produced %d articles — the Scan is not actually reading the image columns", len(articles))
	}
	if err == nil {
		t.Errorf("a projection/Scan divergence returned no error — the caller would read the empty set as 'no posts' and silently keep the stale listing")
	}
}

// THE SILENT-DEGRADATION PATH, closed (council round 170147b4, bug_historian,
// gating, 2026-08-25). The first cut logged-and-skipped EVERY unscannable row
// and returned an empty slice; the caller treats an empty article set as "leave
// the existing listing alone", so a projection-shape mismatch would have kept a
// stale listing, reported success, and told nobody. These two tests pin the
// distinction that fixes it — and the distinction is the point, because
// "no eligible posts" is a legitimate empty result that must NOT become an error.
func TestBlogListingErrorsWhenEveryRowFailsToScan(t *testing.T) {
	db, mock := newRetractMockDB(t)
	// Three rows offered, all one column short: a column was added to the
	// projection and the Scan destinations were not updated with it.
	short := blogRowCols[:len(blogRowCols)-1]
	rows := sqlmock.NewRows(short)
	for i := 0; i < 3; i++ {
		rows.AddRow(uuid.New(), "post", "/blog/post.html", "Post",
			"Summary.", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), 4000,
			"card-key", "hero-key")
	}
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
	got, qerr := db.Query(`SELECT`)
	if qerr != nil {
		t.Fatalf("mock query: %v", qerr)
	}
	defer got.Close()

	articles, err := scanBlogArticles(got, zap.NewNop())
	if err == nil {
		t.Fatalf("every row failed to scan and scanBlogArticles returned no error (%d articles) — the step would report success while the listing silently kept its stale contents", len(articles))
	}
	if !strings.Contains(err.Error(), "diverged") {
		t.Errorf("error does not name the cause (projection vs Scan divergence): %v", err)
	}
}

// ...and the other side of it: an EMPTY result set is not an error. A site with
// no eligible posts is ordinary, and turning that into a failure would break
// the caller's existing "keep the previous listing" path.
func TestBlogListingDoesNotErrorOnAnEmptyResultSet(t *testing.T) {
	db, mock := newRetractMockDB(t)
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows(blogRowCols))
	got, qerr := db.Query(`SELECT`)
	if qerr != nil {
		t.Fatalf("mock query: %v", qerr)
	}
	defer got.Close()

	articles, err := scanBlogArticles(got, zap.NewNop())
	if err != nil {
		t.Fatalf("an empty result set must not be an error — a site with no eligible posts is ordinary: %v", err)
	}
	if len(articles) != 0 {
		t.Errorf("got %d articles from an empty result set", len(articles))
	}
}

// One bad row among good ones is still tolerated: a single malformed post must
// not blank a live listing. This is what stops the fix above from becoming a
// new fragility.
func TestBlogListingSkipsOneBadRowButKeepsTheRest(t *testing.T) {
	db, mock := newRetractMockDB(t)
	rows := sqlmock.NewRows(blogRowCols).
		AddRow(uuid.New(), "good", "/blog/good.html", "Good Post", "Summary.",
			time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), 4000, "card-good", "", "").
		AddRow(uuid.New(), "bad", "/blog/bad.html", "Bad Post", "Summary.",
			"not-a-time", 4000, "card-bad", "", "")
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
	got, qerr := db.Query(`SELECT`)
	if qerr != nil {
		t.Fatalf("mock query: %v", qerr)
	}
	defer got.Close()

	articles, err := scanBlogArticles(got, zap.NewNop())
	if err != nil {
		t.Fatalf("one unscannable row among good ones must not fail the rebuild: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("got %d articles, want 1 — the good row must survive its bad neighbour", len(articles))
	}
	if title, _ := articles[0]["title"].(string); title != "Good Post" {
		t.Errorf("surviving article is %q, want \"Good Post\"", title)
	}
}

// The cap is SHARED with the resolver, not a second literal (council round
// 170147b4, guardian). Without it the two writers of one listing disagree:
// [MEASURED 2026-08-25] webdesign.co.uk has 40 eligible blog posts, so
// query.blog_posts yields 24 items and this action would have yielded 40.
func TestBlogPostsQueryCapsAtTheSharedListingLimit(t *testing.T) {
	want := "LIMIT " + strconv.Itoa(queryresolve.PageListingHardCap)
	if !strings.Contains(blogPostsQuery, want) {
		t.Errorf("blogPostsQuery must carry %q — an uncapped query pays PageImageJoinsSQL's per-row LATERAL unbounded AND can disagree with the resolver about the same listing:\n%s", want, blogPostsQuery)
	}
	if queryresolve.PageListingHardCap != 24 {
		t.Errorf("PageListingHardCap is %d — if the shared cap moves, confirm the resolvers and this action still agree", queryresolve.PageListingHardCap)
	}
}

// The query must SPLICE the shared fragments, not carry a copy of their text.
// Same reasoning as TestBlogPostsQueryUsesTheSharedArticleContract next door:
// two hand-maintained copies of one join is the drift class bugs_closed/023
// documents, and here the two copies would disagree about what a listing item's
// image IS.
func TestBlogPostsQueryUsesTheSharedImageProjection(t *testing.T) {
	if !strings.Contains(blogPostsQuery, queryresolve.PageImageProjectionSQL) {
		t.Errorf("blogPostsQuery must splice queryresolve.PageImageProjectionSQL verbatim.\nwant fragment:\n%s\ngot query:\n%s",
			queryresolve.PageImageProjectionSQL, blogPostsQuery)
	}
	if !strings.Contains(blogPostsQuery, queryresolve.PageImageJoinsSQL) {
		t.Errorf("blogPostsQuery must splice queryresolve.PageImageJoinsSQL verbatim — without the joins the projection selects columns that do not exist.\nwant fragment:\n%s\ngot query:\n%s",
			queryresolve.PageImageJoinsSQL, blogPostsQuery)
	}
}
