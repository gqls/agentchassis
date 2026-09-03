// FILE: platform/orchestration/actions/fork_theme_composition_fit_test.go
//
// Tests for the layout-fit evidence and the widened library-gap predicate
// added by bugs_open/445. Before this change fork_theme_composition.go had NO
// test of any kind, so these are also the first tests of resolveLayoutByTags
// itself.
//
// WHAT THE BUG WAS, because it decides what these tests must pin. The
// library-growth signal fired only on `IsFallback` (no layout scores above
// zero anywhere) or `IsSchemeMismatch`. But the category, description and
// same-scheme bonuses are added to `total` INDEPENDENTLY of any tag matching,
// so `total > 0` is satisfiable with `tagScore == 0`. A layout matching NONE
// of a site's tags therefore counted as the library having answered.
// [MEASURED 2026-09-03] four live sites are recorded by this code as
// `tags 0.00` with lineage `library_match`, and exactly two
// needs_new_layout_candidate items exist across 63,007 work items ever
// written, BOTH from the degenerate no-tags-at-all arm.
//
// The fixtures below are the estate's own seed rows as of 2026-09-03, not
// invented shapes — a fixture you compose to exercise a rule will exercise it.
//
// MUTATIONS RUN, and the distinct subset each one kills (this repo requires a
// guard be proven by mutation, not by a mock's bookkeeping):
//
// RESULTS AS RUN, 2026-09-03 (not predicted — the (ii) prediction was WRONG and
// is corrected here rather than tidied away):
//
//	(i)   lmMinTagCoverage = 0  ->  kills TestOneAttractorTagIsAWeakFit and
//	      TestZeroTagOverlapStillCountsAsAGap ONLY; scheme, fallback and
//	      adequate-fit still pass. Proves the threshold is what those two rest
//	      on. (It also printed TagCoverage = 0.072 for the designblog fixture,
//	      reproducing that site's live measured 7% — an independent check that
//	      the Go formula matches the one validated against the fleet.)
//	(ii)  coverage denominator changed from the full site-term weight to
//	      best.tagScore (coverage becomes 1.0 whenever tagScore > 0)  ->  killed
//	      TestOneAttractorTagIsAWeakFit ALONE. I predicted it would also kill
//	      TestZeroTagOverlapStillCountsAsAGap; it does not, and the reason is
//	      worth keeping: that case has tagScore == 0, so the mutated expression
//	      is never evaluated and coverage stays 0. A mutation can miss a test
//	      by never reaching it — which is why the surviving test is not
//	      evidence the denominator is unimportant there.
//	(iii) LibraryGap() reverted to `IsFallback || IsSchemeMismatch`  ->  kills
//	      the two weak-fit cases; fallback and scheme cases still pass. Proves
//	      the new arm is what carries the fix and that the OLD arms are intact.
package actions

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// layoutFixtureRows returns the six-column shape resolveLayoutByTags scans,
// populated from the real active layouts.
func layoutFixtureRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "category", "industry_tags", "scheme", "description"}).
		AddRow("11111111-0000-0000-0000-000000000001", "magazine-grid", "editorial",
			`["publication","news","blog","opinion","long-form","editorial"]`, "",
			"Publication layout with featured article, main 2/3 + 1/3 sidebar grid, article cards. Suits news, opinion, long-form blogs, and curated content networks.").
		AddRow("11111111-0000-0000-0000-000000000002", "tool-portal-light", "interactive",
			`["interactive-platform","tool-portal","tools","interactive","founder-tools","editorial-publication"]`, "light",
			"Warm, flat, editorial multi-tool portal. Index pages present tools, guides and reports as bordered cards.").
		AddRow("11111111-0000-0000-0000-000000000003", "brochure-bold", "brochure",
			`["tech-startup","saas","fitness","landing-page","conversion"]`, "",
			"High-energy conversion layout with tall hero, gradient accents, oversized typography.").
		AddRow("11111111-0000-0000-0000-000000000004", "soft-editorial", "editorial",
			`["wellness","lifestyle","bakery","artisan","personal-brand","long-form"]`, "light",
			"Warm, reading-first layout with tinted background, serif display headings.").
		AddRow("11111111-0000-0000-0000-000000000005", "tool-portal-dark", "interactive",
			`["interactive-platform","tool-portal","tools","developer-tools","game-design"]`, "dark",
			"Dark developer-utility portal with index pages for tools, guides and playable prototypes.")
}

// resolveWithFixture runs resolveLayoutByTags against the fixture library.
func resolveWithFixture(t *testing.T, category string, tags []string, scheme string) *layoutResolution {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("FROM layouts").WillReturnRows(layoutFixtureRows())

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := resolveLayoutByTags(context.Background(), tx, category, tags, scheme, zap.NewNop())
	if err != nil {
		t.Fatalf("resolveLayoutByTags: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
	return res
}

// TestOneAttractorTagIsAWeakFit is designblog.co.uk's real shape: `editorial`
// canonicalises to `editorial-publication`, which magazine-grid carries, and
// the other nine tags match nothing in the library. The recorded live score
// was 3.05 (tags 2.30) at 7% coverage, and NO gap was raised.
func TestOneAttractorTagIsAWeakFit(t *testing.T) {
	res := resolveWithFixture(t, "editorial", []string{
		"editorial-blog", "content-hub", "design-publication", "digital-marketing",
		"web-design", "uk-business", "editorial-guides", "long-form-content",
		"content-platform",
	}, "light")

	if res.LayoutName != "magazine-grid" {
		t.Fatalf("selection changed: got %q, want magazine-grid — this change must NOT alter which layout wins", res.LayoutName)
	}
	if res.IsFallback {
		t.Errorf("IsFallback = true; the old predicate's whole problem is that this case is NOT a fallback")
	}
	if got := len(res.Fit.MatchedTerms); got != 1 {
		t.Errorf("MatchedTerms = %v (%d); want exactly 1 (editorial-publication)", res.Fit.MatchedTerms, got)
	}
	if res.Fit.TagCoverage >= lmMinTagCoverage {
		t.Errorf("TagCoverage = %.3f; want < %.2f", res.Fit.TagCoverage, lmMinTagCoverage)
	}
	if !res.IsWeakFit {
		t.Errorf("IsWeakFit = false at coverage %.3f", res.Fit.TagCoverage)
	}
	if !res.LibraryGap() {
		t.Errorf("LibraryGap() = false — this is the exact case bugs_open/445 was filed for")
	}
	if got := res.GapReason(); got != "weak_tag_fit" {
		t.Errorf("GapReason() = %q, want weak_tag_fit", got)
	}
	if len(res.Fit.UnmatchedTerms) == 0 {
		t.Errorf("UnmatchedTerms empty; it is the field that names the vocabulary the library cannot serve")
	}
}

// TestZeroTagOverlapStillCountsAsAGap is webdesign.uk's real shape: the live
// row records `score 0.75 (tags 0.00)` with lineage `library_match`. 0.75 is
// exactly lmCategoryMatchBonus — the category bonus alone, no tag overlap at
// all — which is how a total above zero was reached without matching anything.
func TestZeroTagOverlapStillCountsAsAGap(t *testing.T) {
	res := resolveWithFixture(t, "brochure", []string{
		"web-design", "digital-agency", "uk-business", "front-end-development",
	}, "light")

	if res.Fit.TagScore != 0 {
		t.Fatalf("fixture drifted: TagScore = %.2f, want 0 (this test is about a bonus-only match)", res.Fit.TagScore)
	}
	if res.Fit.Score <= 0 {
		t.Fatalf("fixture drifted: Score = %.2f, want > 0 — a zero total would fire the OLD predicate and prove nothing", res.Fit.Score)
	}
	if res.IsFallback {
		t.Errorf("IsFallback = true; the point of this case is that the old arms stay silent on it")
	}
	if res.Fit.TagCoverage != 0 {
		t.Errorf("TagCoverage = %.3f, want exactly 0", res.Fit.TagCoverage)
	}
	if !res.LibraryGap() || res.GapReason() != "weak_tag_fit" {
		t.Errorf("LibraryGap()=%v GapReason()=%q; want true/weak_tag_fit", res.LibraryGap(), res.GapReason())
	}
}

// TestAdequateFitIsNotAGap is the control. Without it a mutation that flags
// EVERY site would pass every other test in this file.
func TestAdequateFitIsNotAGap(t *testing.T) {
	res := resolveWithFixture(t, "interactive", []string{
		"interactive-platform", "tool-portal", "tools", "interactive",
	}, "light")

	if res.LayoutName != "tool-portal-light" {
		t.Fatalf("got %q, want tool-portal-light", res.LayoutName)
	}
	if res.Fit.TagCoverage < lmMinTagCoverage {
		t.Errorf("TagCoverage = %.3f; want >= %.2f for a well-served site", res.Fit.TagCoverage, lmMinTagCoverage)
	}
	if res.IsWeakFit || res.LibraryGap() {
		t.Errorf("a well-fitted site was flagged as a library gap (coverage %.3f) — the signal would cry wolf fleet-wide", res.Fit.TagCoverage)
	}
	if res.GapReason() != "" {
		t.Errorf("GapReason() = %q, want empty", res.GapReason())
	}
}

// TestSchemeGapArmStillFires pins the pre-existing arm. Mutation (iii) must
// leave this passing — that is what proves the fix ADDED an arm rather than
// replacing one.
//
// This needs its OWN fixture, and finding out why was worth the detour: with
// the shared fixture the arm cannot be reached at all, because the same-scheme
// bonus (lmSameSchemeBonus, 0.50) lifts ANY same-scheme layout to total > 0
// with zero tag overlap, and lmFirstEligible then takes it. That is not a
// contrivance of the test — it is the live mechanism behind soft-editorial
// scoring 0.50 as a permanent runner-up on 27 of 33 sites while matching none
// of their tags. So the scheme arm only fires when the library offers no
// same-scheme layout at all: here, a dark site against a light-only library.
func TestSchemeGapArmStillFires(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("FROM layouts").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "category", "industry_tags", "scheme", "description"}).
			AddRow("11111111-0000-0000-0000-000000000002", "tool-portal-light", "interactive",
				`["interactive-platform","tool-portal","founder-tools"]`, "light",
				"Warm, flat, editorial multi-tool portal.").
			AddRow("11111111-0000-0000-0000-000000000004", "soft-editorial", "editorial",
				`["wellness","lifestyle","artisan"]`, "light",
				"Warm, reading-first layout with tinted background."))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := resolveLayoutByTags(context.Background(), tx, "", []string{"founder-tools"}, "dark", zap.NewNop())
	if err != nil {
		t.Fatalf("resolveLayoutByTags: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}

	if !res.IsSchemeMismatch {
		t.Fatalf("IsSchemeMismatch = false; fixture drifted — a dark site against a light-only library must cross schemes")
	}
	if !res.LibraryGap() {
		t.Errorf("LibraryGap() = false on a scheme mismatch")
	}
	if got := res.GapReason(); got != "scheme_mismatch" {
		t.Errorf("GapReason() = %q, want scheme_mismatch (fallback and scheme must outrank weak_tag_fit)", got)
	}
}

// TestFitEvidenceIsRecordedNotInferred guards the distinction readFloatFromContext
// exists for downstream: a measured-zero coverage and an absent one must not be
// the same value. Here we assert the threshold travels with the score, so a
// later change to lmMinTagCoverage cannot silently re-interpret old rows.
func TestFitEvidenceIsRecordedNotInferred(t *testing.T) {
	res := resolveWithFixture(t, "editorial", []string{"content-hub"}, "light")

	if res.Fit.Threshold != lmMinTagCoverage {
		t.Errorf("Fit.Threshold = %.2f, want %.2f — the threshold must be recorded with the score it judged",
			res.Fit.Threshold, lmMinTagCoverage)
	}
	if res.Fit.SiteTermCount == 0 {
		t.Errorf("SiteTermCount = 0; the coverage denominator would be unreadable after the fact")
	}
}

// TestFallbackRecordsEveryTermAsUnmatched — a hard fallback matched nothing by
// construction, so the evidence must say so explicitly rather than arriving as
// an empty struct that reads identically to "not measured".
func TestFallbackRecordsEveryTermAsUnmatched(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	// No layout scores above zero for these terms.
	mock.ExpectQuery("FROM layouts").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "category", "industry_tags", "scheme", "description"}).
			AddRow("11111111-0000-0000-0000-000000000009", "media-grid", "media",
				`["video","audio","podcast"]`, "", "Thumbnail-dominant layout for media libraries."))
	mock.ExpectQuery("FROM layouts").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "scheme"}).
			AddRow("11111111-0000-0000-0000-00000000000f", "brochure-formal", ""))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := resolveLayoutByTags(context.Background(), tx,
		"veterinary", []string{"uk-farming", "livestock-health"}, "", zap.NewNop())
	if err != nil {
		t.Fatalf("resolveLayoutByTags: %v", err)
	}
	if !res.IsFallback {
		t.Fatalf("IsFallback = false; fixture drifted — nothing should have scored")
	}
	if got := res.GapReason(); got != "fallback" {
		t.Errorf("GapReason() = %q, want fallback", got)
	}
	if len(res.Fit.UnmatchedTerms) != res.Fit.SiteTermCount || res.Fit.SiteTermCount == 0 {
		t.Errorf("UnmatchedTerms = %v, SiteTermCount = %d; a fallback must record every term as unmatched",
			res.Fit.UnmatchedTerms, res.Fit.SiteTermCount)
	}
	var _ = sql.ErrNoRows
}
