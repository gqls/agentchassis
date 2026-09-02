// FILE: platform/orchestration/actions/wire_page_hero_on_landing_test.go
//
// The tests that matter: the OFF default (the opt-in ruling's whole point), the
// deployer-derivation pin (IMG-072's lesson — the arg match fails if anyone
// re-derives the path from the purpose alone), and the three WHERE arms that
// are each a one-line deletion leaving the happy path green (hero-family,
// 357-fragment exclusion, the double value-gate).

package actions

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func TestHeroWireArmedDefaultsOff(t *testing.T) {
	if heroWireArmed(nil) {
		t.Fatalf("nil config must be OFF — the unsafe default is the armed one (owner ruling 2026-08-02 §2)")
	}
	if heroWireArmed(map[string]interface{}{}) {
		t.Fatalf("empty config must be OFF")
	}
	if heroWireArmed(map[string]interface{}{"wire_hero_on_landing": false}) {
		t.Fatalf("explicit false must be OFF")
	}
	if !heroWireArmed(map[string]interface{}{"wire_hero_on_landing": true}) {
		t.Fatalf("explicit true must arm the wire")
	}
}

func TestWirePageHeroOnLandingHappyPathPinsTheDeployerDerivation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	siteID := uuid.MustParse("62b5978e-4271-4589-8e00-4baebfc0447c")

	mock.ExpectBegin()
	mock.ExpectQuery(`FROM assets a`).WithArgs(siteID, "content_hero_tool_repayment").
		WillReturnRows(sqlmock.NewRows([]string{"purpose", "fallback"}).
			AddRow("content_hero", "/assets/images/hero-home.jpg"))

	// Every load-bearing WHERE arm, asserted on the emitted SQL: the
	// hero-family predicate, the fragment exclusion built from the SHARED
	// marker list, and BOTH halves of the value-gate.
	parts := []string{
		`cc\.function = 'hero' OR cc\.function LIKE 'hero-%'`,
		`OR cc\.function LIKE '%-hero' OR cc\.category = 'hero'`,
	}
	for _, m := range checks.InteractiveStructuralMarkers {
		parts = append(parts, regexp.QuoteMeta(m))
	}
	parts = append(parts,
		`content_data->>'hero_url', ''\) IN \('', \$3, \$5\)`,
		`content_data->>'background_image', ''\) IN \('', \$3, \$5\)`,
	)
	updateRe := "(?s)UPDATE page_components pc.*" + strings.Join(parts, ".*")

	// The WithArgs is the derivation pin: $4 must be the DEPLOYER's path for
	// (content_hero_tool_repayment, content_hero) — hyphenated, key-derived. A
	// purpose-alone re-derivation (the IMG-072 defect) would pass
	// /assets/images/content-hero.jpg and fail this match.
	mock.ExpectExec(updateRe).
		WithArgs(siteID, "tool-repayment", legacyHeroFallbackLiteral,
			"/assets/images/content-hero-tool-repayment.jpg", "/assets/images/hero-home.jpg").
		WillReturnResult(sqlmock.NewResult(0, 2))

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	got := wirePageHeroOnLanding(context.Background(), tx, siteID, "tool-repayment", zap.NewNop())
	if got != "wired:2" {
		t.Fatalf("disposition = %q, want wired:2", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestWirePageHeroOnLandingSkipsWhenNoContentHero(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	siteID := uuid.New()

	mock.ExpectBegin()
	// purpose '' = no active asset under the page's ContentHeroKey. No UPDATE
	// may follow — ExpectationsWereMet catches a write that happens anyway.
	mock.ExpectQuery(`FROM assets a`).WithArgs(siteID, "content_hero_logo_page").
		WillReturnRows(sqlmock.NewRows([]string{"purpose", "fallback"}).AddRow("", ""))

	tx, _ := db.BeginTx(context.Background(), nil)
	if got := wirePageHeroOnLanding(context.Background(), tx, siteID, "logo-page", zap.NewNop()); got != "skipped_no_content_hero" {
		t.Fatalf("disposition = %q, want skipped_no_content_hero", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestWirePageHeroOnLandingZeroRowsIsAnHonestSkip(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	siteID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(`FROM assets a`).WithArgs(siteID, "content_hero_about").
		WillReturnRows(sqlmock.NewRows([]string{"purpose", "fallback"}).
			AddRow("content_hero", ""))
	mock.ExpectExec(`UPDATE page_components pc`).
		WithArgs(siteID, "about", legacyHeroFallbackLiteral,
			"/assets/images/content-hero-about.jpg", "").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tx, _ := db.BeginTx(context.Background(), nil)
	if got := wirePageHeroOnLanding(context.Background(), tx, siteID, "about", zap.NewNop()); got != "skipped_no_eligible_component" {
		t.Fatalf("disposition = %q, want skipped_no_eligible_component (0 rows must not read as wired)", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
