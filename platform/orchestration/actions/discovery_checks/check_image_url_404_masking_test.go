// FILE: platform/orchestration/actions/discovery_checks/check_image_url_404_masking_test.go
//
// bugs_open/128 — pins the purpose-vs-path blind spot, deliberately, as a
// DOCUMENTED LIMITATION rather than leaving it as folklore.
//
// `image_url_404` skips any rendered `/assets/images/<name>.<ext>` whose name (or
// its prefix before the first hyphen) matches an ACTIVE ASSET PURPOSE for the
// site. Purposes are names like "hero", "icon", "logo" — they are not paths. So a
// site that owns one hero asset, at any URL including an S3 one, makes every
// rendered `hero*` path unreportable regardless of whether it resolves.
//
// Measured fleet-wide 2026-07-29: 79 of 95 distinct rendered image paths across
// 13 sites are masked this way — 83% of the surface the check is named for — and
// SIX of the masked paths were serving live 404s (`/assets/images/hero.jpg` on
// gamesdesign, idea.uk, oufe, relojistas, vonc and webdesign).
//
// WHY THESE TESTS ASSERT THE DEFECT INSTEAD OF THE FIX. Switching the predicate to
// compare paths instead of purposes was measured and REFUTED: only 9 of those 79
// paths have any asset row whose url or filename carries the basename, while 73 of
// the 79 serve HTTP 200. Nothing in the database records which static files were
// deployed — `assets.url` is an S3 URL with `filename`/`storage_path` empty, and
// `v_site_assets` is a count view — so a DB-only check cannot answer "does this
// path resolve", and a naive path swap would flag ~70 working images. The skip is
// a crude false-positive suppressor that also suppresses the true positives.
//
// So if you are here because one of these tests failed: good. You are changing the
// masking, and these tests exist to make you read the two paragraphs above first,
// especially the next one.
//
// THE TRAP IF YOU UNMASK IT. `knownPurposeMapping` maps exactly two purposes —
// "hero" and "logo" — and routes them to `image-build-handler` with a real
// HandlerAgent, i.e. automatic image regeneration. Those are precisely the two
// purposes most likely to exist as assets, so the skip removes them before the
// routing branch is reached. Verified: of all 10 work items ever created under an
// `image_url_404:*` key fleet-wide, 10 are the flag-only `image_url_404` type and
// ZERO are `needs_hero_image` or `needs_logo`. **The recognised branch has never
// fired in production.** Unmasking therefore does not merely add findings — it
// activates a dormant fleet-wide auto-regeneration path for the first time. Cost
// that before switching it on.

package discovery_checks

import (
	"context"
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// runImageURL404 drives the real check against mocked rows: one rendered
// component, and the site's active asset purposes.
func runImageURL404(t *testing.T, html string, purposes []string, prompts *string) *CheckResult {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT pc.rendered_html").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).AddRow(html))

	prows := sqlmock.NewRows([]string{"purpose"})
	for _, p := range purposes {
		prows.AddRow(p)
	}
	mock.ExpectQuery("SELECT DISTINCT purpose").WillReturnRows(prows)

	// Only reached on the recognised branch; harmless to arm either way.
	if prompts == nil {
		mock.ExpectQuery("SELECT data->'image_prompts'").WillReturnError(sql.ErrNoRows)
	} else {
		mock.ExpectQuery("SELECT data->'image_prompts'").
			WillReturnRows(sqlmock.NewRows([]string{"image_prompts"}).AddRow([]byte(*prompts)))
	}

	res, err := (&ImageURL404Check{}).Run(DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.New(),
		Pipeline:  "design",
		AgentType: "test",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("ImageURL404Check.Run: %v", err)
	}
	return res
}

// THE BLIND SPOT. This is the fundamentallyai.com case from bugs_open/128: the
// site has four active purpose='hero' assets (all S3 URLs), /assets/images/hero.jpg
// serves 404, and the check reports nothing at all.
func TestImageURL404_PurposeMasksABrokenPath(t *testing.T) {
	res := runImageURL404(t,
		`<div><img src="/assets/images/hero.jpg" alt="x"></div>`,
		[]string{"hero", "icon", "og_card"}, nil)

	if len(res.WorkItems) != 0 || len(res.Findings) != 0 {
		t.Fatalf("DOCUMENTED LIMITATION CHANGED — hero.jpg is no longer masked by the "+
			"'hero' purpose. Read this file's header before updating it: unmasking "+
			"activates image-build-handler routing that has never fired in production. "+
			"got %d work items, %d findings", len(res.WorkItems), len(res.Findings))
	}
}

// The masking is WIDER than an exact purpose match: the prefix before the first
// hyphen is also tried, so one hero asset hides the entire hero-* path space.
// vonc.com renders hero-archetypes, hero-home, icon-catalyst, icon-judge and six
// more icon-* paths — every one invisible for this reason.
func TestImageURL404_PurposePrefixMasksTheWholeVariantSpace(t *testing.T) {
	for _, path := range []string{
		"/assets/images/hero-archetypes.jpg",
		"/assets/images/hero-home.jpg",
		"/assets/images/icon-catalyst.jpg",
	} {
		res := runImageURL404(t, `<img src="`+path+`">`, []string{"hero", "icon"}, nil)
		if len(res.WorkItems) != 0 {
			t.Errorf("%s: expected masked by purpose prefix, got %d work items",
				path, len(res.WorkItems))
		}
	}
}

// The half that WORKS. brand-illustration's prefix is "brand", which is not a
// purpose, so it is reported — as a flag-only item with no handler. Note what the
// finding actually means: "no asset purpose matches this basename". The path
// itself serves 200. That gap between the item's name and its meaning is the
// other half of bugs_open/128.
func TestImageURL404_UnmatchedPrefixIsReportedFlagOnly(t *testing.T) {
	res := runImageURL404(t,
		`<img src="/assets/images/brand-illustration.jpg">`,
		[]string{"hero", "icon", "og_card"}, nil)

	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 work item, got %d", len(res.WorkItems))
	}
	w := res.WorkItems[0]
	if w.ItemType != "image_url_404" {
		t.Errorf("item_type = %q, want image_url_404", w.ItemType)
	}
	if w.HandlerAgent != "" {
		t.Errorf("this branch is flag-only by design; HandlerAgent = %q", w.HandlerAgent)
	}
	if w.ItemKey != "image_url_404:brand-illustration" {
		t.Errorf("item_key = %q — dedup keys on the basename", w.ItemKey)
	}
}

// The routing branch is REACHABLE ONLY when the site owns no asset of that
// purpose — which is why it has never fired. Asserting it here proves the branch
// itself is sound, so a future fix knows the masking is the only thing standing
// between production and fleet-wide auto-regeneration.
func TestImageURL404_RecognisedBranchWorksButOnlyWithNoSuchPurpose(t *testing.T) {
	res := runImageURL404(t,
		`<img src="/assets/images/hero.jpg">`,
		[]string{"icon", "og_card"}, nil) // deliberately NO hero purpose

	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item, got %d", len(res.WorkItems))
	}
	w := res.WorkItems[0]
	if w.ItemType != "needs_hero_image" {
		t.Errorf("item_type = %q, want needs_hero_image", w.ItemType)
	}
	if w.HandlerAgent != "image-build-handler" {
		t.Errorf("HandlerAgent = %q, want image-build-handler — this is the dormant "+
			"auto-regeneration path", w.HandlerAgent)
	}
	if w.Severity != "high" {
		t.Errorf("severity = %q, want high", w.Severity)
	}
}

// Chrome is never scanned at all (bugs_open/128 defect 3): the reference query
// reads page_components only, so anything referenced solely by site_components —
// the surface that appears on EVERY page — cannot be seen. Pinned by asserting
// that the check issues no query against site_components: sqlmock's ordered
// expectations would fail on an unexpected one.
func TestImageURL404_NeverScansSiteComponents(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT pc.rendered_html").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).
			AddRow(`<img src="/assets/images/brand-illustration.jpg">`))
	mock.ExpectQuery("SELECT DISTINCT purpose").
		WillReturnRows(sqlmock.NewRows([]string{"purpose"}))
	// NO image_prompts expectation: brand-illustration takes the flag-only branch,
	// which never loads prompts. Arming it would leave an unmatched expectation and
	// this test would fail for a reason that has nothing to do with chrome.

	if _, err := (&ImageURL404Check{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(),
		Pipeline: "design", AgentType: "test", BatchID: uuid.New(), Logger: zap.NewNop(),
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// If a future change adds the chrome surface, this expectation set is
	// incomplete and the mock reports an unfulfilled/unexpected query — which is
	// the signal to update defect 3 in bugs_open/128 as fixed.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("query set changed — chrome scanning may have been added: %v", err)
	}
}
