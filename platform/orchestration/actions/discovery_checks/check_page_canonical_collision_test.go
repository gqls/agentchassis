// FILE: platform/orchestration/actions/discovery_checks/check_page_canonical_collision_test.go
//
// The fixtures reproduce the two REAL robot-hands.com collisions this check
// was built from (bugs_open/080, measured 2026-08-03), because each pair is
// caught by a DIFFERENT signal and only union-merging the two yields the right
// item count:
//
//	/news:            news [news-index] + news-index [section-index]
//	                  → the NAME signal (both canonicalise to "news-index");
//	                    the path signal also fires — one group, not two.
//	/gripper-catalog: gripper-catalog [content] + gripper-catalog-index
//	                  [section-index] → the PATH signal ONLY (a content row
//	                    canonicalises to itself).
package discovery_checks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func robotHandsFixture() []collisionPage {
	mk := func(name, url, pageType, status string) collisionPage {
		return collisionPage{
			ID: name, Name: name, URL: url, PageType: pageType,
			Build: "deployed", Status: status,
			PathKey:   pageURLPathKey(url),
			CanonName: collisionCanonName(name, pageType),
		}
	}
	return []collisionPage{
		mk("news", "/news.html", "news-index", "active"),
		mk("news-index", "/news/index.html", "section-index", "active"),
		mk("gripper-catalog", "/gripper-catalog.html", "content", "active"),
		mk("gripper-catalog-index", "/gripper-catalog/index.html", "section-index", "active"),
		// An innocent bystander — must join no group.
		mk("about", "/about.html", "content", "active"),
	}
}

func TestPageURLPathKey(t *testing.T) {
	cases := map[string]string{
		"/news/index.html":    "/news",
		"/news.html":          "/news",
		"/index.html":         "/", // the homepage is a signal, not a blank
		"/tools/x/index.html": "/tools/x",
		"/tools/x.html":       "/tools/x",
		"/tools.html#frag":    "/tools.html#frag", // fragment urls only collide with exact twins
	}
	for in, want := range cases {
		if got := pageURLPathKey(in); got != want {
			t.Errorf("pageURLPathKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCollisionCanonName_SkipsWhatCannotCarryASignal(t *testing.T) {
	if got := collisionCanonName("news", "news-index"); got != "news-index" {
		t.Errorf("news/news-index canonicalises to %q, want news-index", got)
	}
	// Legacy "index" page_type collapses to the homepage regardless of slug —
	// it must contribute NO name signal or every such row false-collides.
	if got := collisionCanonName("anything", "index"); got != "" {
		t.Errorf("page_type=index must be skipped, got %q", got)
	}
	// An uncanonicalisable identity contributes no signal rather than
	// colliding with every other empty triple.
	if got := collisionCanonName("tool-", "tool"); got != "" {
		t.Errorf("empty canonical triple must be skipped, got %q", got)
	}
}

// TestGroupCollisions_UnionMergesAcrossSignals is the acceptance shape: the
// live robot-hands fixture must yield EXACTLY TWO groups — a naive
// one-item-per-signal grouping yields three (news-by-name, news-by-path,
// gripper-by-path).
func TestGroupCollisions_UnionMergesAcrossSignals(t *testing.T) {
	groups := groupCollisions(robotHandsFixture())
	if len(groups) != 2 {
		t.Fatalf("got %d groups, want exactly 2 (union-merged)", len(groups))
	}
	for _, g := range groups {
		if len(g) != 2 {
			t.Errorf("group %q has %d members, want 2", collisionGroupKey(g), len(g))
		}
	}
	if k := collisionGroupKey(groups[0]); k != "/gripper-catalog" {
		t.Errorf("first group key = %q, want /gripper-catalog", k)
	}
	if k := collisionGroupKey(groups[1]); k != "/news" {
		t.Errorf("second group key = %q, want /news", k)
	}
}

// TestGroupCollisions_NameOnlyCollision: home and index live at different
// paths but canonicalise to one name — the collapse rule the helper exists for.
func TestGroupCollisions_NameOnlyCollision(t *testing.T) {
	pages := []collisionPage{
		{ID: "a", Name: "home", URL: "/home.html", PageType: "content", Status: "active",
			PathKey: pageURLPathKey("/home.html"), CanonName: collisionCanonName("home", "content")},
		{ID: "b", Name: "index", URL: "/index.html", PageType: "landing", Status: "active",
			PathKey: pageURLPathKey("/index.html"), CanonName: collisionCanonName("index", "landing")},
	}
	groups := groupCollisions(pages)
	if len(groups) != 1 {
		t.Fatalf("got %d groups, want 1 (home+index collapse)", len(groups))
	}
	// Two distinct paths → the key falls back to the shared canonical name.
	if k := collisionGroupKey(groups[0]); k != "index" {
		t.Errorf("group key = %q, want index", k)
	}
}

func TestGroupCollisions_NoCollisionsNoGroups(t *testing.T) {
	pages := []collisionPage{
		{ID: "a", Name: "about", URL: "/about.html", PageType: "content", Status: "active",
			PathKey: "/about", CanonName: "about"},
		{ID: "b", Name: "contact", URL: "/contact.html", PageType: "content", Status: "active",
			PathKey: "/contact", CanonName: "contact"},
	}
	if groups := groupCollisions(pages); len(groups) != 0 {
		t.Fatalf("got %d groups, want 0", len(groups))
	}
}

// ---------------------------------------------------------------------------
// Run — sqlmock
// ---------------------------------------------------------------------------

func canonicalCollisionCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.MustParse("00ff3af5-aaaa-bbbb-cccc-0123456789ab"),
		Pipeline:  "build",
		AgentType: "completeness-discovery-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	}, mock
}

func pageRows(pages []collisionPage) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "name", "url", "page_type", "build_status", "status"})
	for _, p := range pages {
		rows.AddRow(p.ID, p.Name, p.URL, p.PageType, p.Build, p.Status)
	}
	return rows
}

// TestRun_FilesExactlyTwoItemsOnRobotHandsShape is the live acceptance case:
// two union-merged groups, both with 2 active members → exactly 2 items with
// stable, path-derived keys.
func TestRun_FilesExactlyTwoItemsOnRobotHandsShape(t *testing.T) {
	dctx, mock := canonicalCollisionCtx(t)

	mock.ExpectQuery("FROM pages").WillReturnRows(pageRows(robotHandsFixture()))
	// Prior-ruling lookups, one per item-worthy group (alphabetical order).
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(dctx.SiteID, "page_canonical_collision:/gripper-catalog").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(dctx.SiteID, "page_canonical_collision:/news").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	// Open-item scan for the retraction seam — nothing open.
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}))

	res, err := (&PageCanonicalCollisionCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 2 {
		t.Fatalf("filed %d items, want exactly 2", len(res.WorkItems))
	}
	for _, wi := range res.WorkItems {
		if wi.Status != "needs_human_review" || wi.HandlerAgent != "" {
			t.Errorf("item %s: status=%q handler=%q — a collision is a decision, not a job",
				wi.ItemKey, wi.Status, wi.HandlerAgent)
		}
		var spec map[string]interface{}
		if err := json.Unmarshal([]byte(wi.SpecJSON), &spec); err != nil {
			t.Fatalf("spec is not JSON: %v", err)
		}
		// The verifier locates the defect through these — an item without them
		// is unverifiable.
		if _, ok := spec["path_keys"].([]interface{}); !ok {
			t.Errorf("item %s spec carries no path_keys", wi.ItemKey)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestRun_OneActiveClaimantIsFindingOnly: dartsonline's shape — an archived
// planned stray beside a live page needs no human decision.
func TestRun_OneActiveClaimantIsFindingOnly(t *testing.T) {
	dctx, mock := canonicalCollisionCtx(t)

	pages := []collisionPage{
		{ID: "a", Name: "guides", URL: "/guides.html", PageType: "landing", Build: "planned", Status: "archived"},
		{ID: "b", Name: "guides-index", URL: "/guides/index.html", PageType: "section-index", Build: "deployed", Status: "active"},
	}
	mock.ExpectQuery("FROM pages").WillReturnRows(pageRows(pages))
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}))

	res, err := (&PageCanonicalCollisionCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("filed %d items, want 0 — only one claimant is active", len(res.WorkItems))
	}
	if len(res.Findings) != 1 {
		t.Fatalf("got %d findings, want 1 — the group must still be reported", len(res.Findings))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestRun_PriorHumanRulingSuppressesRefiling: a wont_fix releases the dedup
// slot, so without this arm the same true-in-the-DB collision would re-file on
// every sweep, for ever.
func TestRun_PriorHumanRulingSuppressesRefiling(t *testing.T) {
	dctx, mock := canonicalCollisionCtx(t)

	pages := robotHandsFixture()[:2] // just the news pair
	mock.ExpectQuery("FROM pages").WillReturnRows(pageRows(pages))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(dctx.SiteID, "page_canonical_collision:/news").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true)) // ruled
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}))

	res, err := (&PageCanonicalCollisionCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("filed %d items, want 0 — a human already ruled on this group", len(res.WorkItems))
	}
	if len(res.Findings) != 1 || res.Findings[0]["suppressed_by_prior_ruling"] != true {
		t.Fatalf("the suppressed group must still surface as a marked finding: %+v", res.Findings)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestRun_RetractsStaleOpenItem: an open item whose group no longer has two
// active claimants was fixed out of band — the RFC_010 seam closes it.
func TestRun_RetractsStaleOpenItem(t *testing.T) {
	dctx, mock := canonicalCollisionCtx(t)

	pages := []collisionPage{
		{ID: "b", Name: "news-index", URL: "/news/index.html", PageType: "section-index", Build: "deployed", Status: "active"},
	}
	mock.ExpectQuery("FROM pages").WillReturnRows(pageRows(pages))
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"item_key"}).
			AddRow("page_canonical_collision:/news"))

	res, err := (&PageCanonicalCollisionCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Resolved) != 1 || res.Resolved[0].ItemKey != "page_canonical_collision:/news" {
		t.Fatalf("want the stale open item retracted, got %+v", res.Resolved)
	}
	// ItemType is REQUIRED by the runner (resolveWorkItems matches on it); an
	// entry without it resolves nothing, silently. This assertion exists because
	// exactly that shipped: the first live retraction left both items open while
	// this test was green — it pinned only the check's side of the contract.
	if res.Resolved[0].ItemType != "page_canonical_collision" {
		t.Fatalf("ResolvedFinding.ItemType = %q — the runner matches on it; empty resolves nothing", res.Resolved[0].ItemType)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// ---------------------------------------------------------------------------
// verifier
// ---------------------------------------------------------------------------

func verifyTargetWithSignature() VerifyTarget {
	return VerifyTarget{
		ItemID:   uuid.New(),
		SiteID:   uuid.New(),
		ItemType: "page_canonical_collision",
		Spec: map[string]interface{}{
			"path_keys":       []interface{}{"/news"},
			"canonical_names": []interface{}{"news-index"},
		},
	}
}

func TestVerify_UnresolvedWhileBothClaimantsActive(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"name", "url", "page_type"}).
			AddRow("news", "/news.html", "news-index").
			AddRow("news-index", "/news/index.html", "section-index"))

	res, err := VerifyPageCanonicalCollisionResolved(context.Background(), db, verifyTargetWithSignature(), zap.NewNop())
	if err != nil {
		t.Fatalf("verifier errored: %v", err)
	}
	if res.Resolved {
		t.Fatalf("resolved=true while both claimants are active: %s", res.Detail)
	}
}

func TestVerify_ResolvedWhenOneClaimantRetired(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// The stray is gone (retired/archived/deleted) — only one active claimant.
	mock.ExpectQuery("FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"name", "url", "page_type"}).
			AddRow("news-index", "/news/index.html", "news-index"))

	res, err := VerifyPageCanonicalCollisionResolved(context.Background(), db, verifyTargetWithSignature(), zap.NewNop())
	if err != nil {
		t.Fatalf("verifier errored: %v", err)
	}
	if !res.Resolved {
		t.Fatalf("resolved=false after the group lost its second claimant: %s", res.Detail)
	}
}

func TestVerify_VanishedSiteIsAnErrorNotASuccess(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	if _, err := VerifyPageCanonicalCollisionResolved(context.Background(), db, verifyTargetWithSignature(), zap.NewNop()); err == nil {
		t.Fatal("a vanished site must be an error — count=0 would be vacuously resolved")
	}
}

func TestVerify_MissingSignatureIsAnError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	target := verifyTargetWithSignature()
	target.Spec = map[string]interface{}{}
	if _, err := VerifyPageCanonicalCollisionResolved(context.Background(), db, target, zap.NewNop()); err == nil {
		t.Fatal("a spec without a group signature must be an error, not vacuously resolved")
	}
}
