// FILE: platform/orchestration/actions/discovery_checks/check_placeholder_image_in_use_test.go
//
// bugs_open/210 (needs_logo slug). Pins three behaviour changes.
//
// WHY THE PATTERN IS TESTED DIRECTLY AND NOT THROUGH sqlmock. sqlmock does not
// execute SQL — it returns the rows the test hands it — so a mocked
// "cross-origin HTML in, no work item out" would be asserting the mock's own
// bookkeeping, not the predicate. The same trap as
// `a-mock-s-own-bookkeeping-cannot-assert-a-negative`. So the anchoring is
// tested as a regex against real markup, and the SQL that carries it was
// separately run against the live database over the real rendered_html of every
// site (RUNBOOK R5 / NOTES §8): it excludes fundamentallyai's cross-origin
// reference and still matches mortgagecalculator's six CSS url('…') ones.
//
// Go's RE2 and Postgres ARE agree on the two constructs used here — a bracket
// expression containing [:space:], and backslash-escaped literals — which is
// why this test is evidence about the deployed predicate and not merely about
// a Go string. Anything relying on a construct where they DIFFER would have to
// be proven in Postgres alone.

package discovery_checks

import (
	"context"
	"regexp"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// The live markup that filed the false positive. fundamentallyai.com renders a
// PARTNER's logo from the partner's own domain; the old unanchored
// `LIKE '%/assets/images/logo.png%'` read that as this site serving its own
// placeholder, and it was the only /assets/images/logo.png match in the fleet.
const liveCrossOriginLogoHTML = `<div class="portfolio-logo">` +
	`<img src="https://leopardessconsulting.co.uk/assets/images/logo.png" alt="Leopardess Consulting"></div>`

func TestPlaceholderImageInUse_CrossOriginReferenceIsNotOurPlaceholder(t *testing.T) {
	re := regexp.MustCompile(sameOriginPathPattern("/assets/images/logo.png"))

	if re.MatchString(liveCrossOriginLogoHTML) {
		t.Fatalf("cross-origin reference matched: the check would file a needs_logo item for a "+
			"partner logo served from the partner's own domain.\nHTML: %s\npattern: %s",
			liveCrossOriginLogoHTML, re.String())
	}

	// The substring really is present — without this the test above could pass
	// for the wrong reason (a pattern that matches nothing at all would also
	// "pass", and silencing the check is this fix's main risk).
	if !strings.Contains(liveCrossOriginLogoHTML, "/assets/images/logo.png") {
		t.Fatal("fixture no longer contains the path; the test has stopped testing anything")
	}
}

// The counterpart, and the one that makes the fix falsifiable: if the anchoring
// is wrong the check goes SILENT, which looks exactly like "fixed" because this
// check's true-positive rate is already near zero. Every form here is one the
// fleet actually renders.
func TestPlaceholderImageInUse_SameOriginFormsStillMatch(t *testing.T) {
	cases := []struct {
		name, path, html string
	}{
		{"double-quoted src", "/assets/images/logo.png",
			`<img src="/assets/images/logo.png" alt="logo">`},
		{"single-quoted src", "/assets/images/logo.png",
			`<img src='/assets/images/logo.png'>`},
		{"unquoted href", "/assets/images/logo.png",
			`<a href=/assets/images/logo.png>`},
		// mortgagecalculator.co.uk's live form — a true positive, six pages.
		{"css url() single-quoted inside gradient", "/assets/images/hero.jpg",
			`background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url('/assets/images/hero.jpg');`},
		{"css url() bare", "/assets/images/hero.jpg",
			`background-image: url(/assets/images/hero.jpg);`},
		{"srcset, space-separated", "/assets/images/logo.png",
			`<img srcset="/assets/images/logo-2x.png 2x, /assets/images/logo.png 1x">`},
		{"newline before the path", "/assets/images/hero.jpg",
			"background:\n\turl(\n\t/assets/images/hero.jpg\n\t);"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			re := regexp.MustCompile(sameOriginPathPattern(c.path))
			if !re.MatchString(c.html) {
				t.Fatalf("same-origin reference did NOT match — the check has been silenced for "+
					"this form.\nHTML: %s\npattern: %s", c.html, re.String())
			}
		})
	}
}

// A protocol-relative URL is still cross-origin. Included because it is the
// form most likely to be added later by a CDN change and read as "ours".
func TestPlaceholderImageInUse_ProtocolRelativeIsAlsoCrossOrigin(t *testing.T) {
	re := regexp.MustCompile(sameOriginPathPattern("/assets/images/logo.png"))
	html := `<img src="//cdn.partner.example/assets/images/logo.png">`
	if re.MatchString(html) {
		t.Fatalf("protocol-relative cross-origin reference matched: %s", html)
	}
}

// The path is regex-escaped, not interpolated raw: a '.' in the path must not
// behave as "any character", or /assets/images/logoXpng would count.
func TestPlaceholderImageInUse_PathIsRegexEscaped(t *testing.T) {
	re := regexp.MustCompile(sameOriginPathPattern("/assets/images/logo.png"))
	if re.MatchString(`<img src="/assets/images/logoXpng">`) {
		t.Fatal("the '.' in the path was treated as a wildcard — the path must be quoted")
	}
}

// TestPlaceholderImageInUse_VariantAssetDoesNotSuppressCanonicalDetection pins
// the 2026-08-12 fix: the asset-existence gate must ask about the CANONICAL
// asset (asset_key == purpose), not "any asset of this purpose". Found live
// on mortgagecalculator.co.uk — superseding the canonical hero asset left
// detection silently suppressed, because hero_about/hero_contact (page-named
// variants, same purpose="hero", different asset_key) were still active.
// hasActiveAssetForPurpose would have returned true here; the correct
// hasActiveAssetForAssetKey must return false, and the check must proceed to
// file a Finding rather than skip.
//
// sqlmock cannot verify the QUERY'S row-filtering semantics (it returns
// whatever rows the test hands it — see this file's header), so it is used
// here only to pin WHICH QUERY SHAPE the code issues (asset_key= , not
// purpose= ); a regression back to hasActiveAssetForPurpose would ask sqlmock
// a question it never expected and fail. The claim that the asset_key-scoped
// query itself returns the right answer against real data is proven
// separately in Postgres (NOTES §18): 2 (wrong, purpose-scoped) vs 0 (correct,
// asset_key-scoped) against this exact live row set.
func TestPlaceholderImageInUse_VariantAssetDoesNotSuppressCanonicalDetection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// 1. isPathReferencedInPages: the page renders the fallback.
	mock.ExpectQuery("FROM page_components").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// 2. hasActiveAssetForAssetKey: MUST be asset_key-scoped. A reversion to
	// hasActiveAssetForPurpose issues a query with no "asset_key" predicate,
	// which this expectation will not match, and the test fails on the
	// resulting sqlmock "call to Query was not expected" error.
	mock.ExpectQuery(`(?s)FROM assets.*asset_key\s*=\s*\$2`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// 3. loadImagePromptsForSite: no planned prompt (the ordinary case).
	mock.ExpectQuery("aspect = 'site_plan'").
		WillReturnRows(sqlmock.NewRows([]string{"data"}))

	// 4-5. DefaultBrandImagePrompt's reads: degrade gracefully, both empty.
	mock.ExpectQuery("FROM sites WHERE id").
		WillReturnRows(sqlmock.NewRows([]string{"domain", "name", "company_name", "tagline"}))
	mock.ExpectQuery("aspect = 'identity'").
		WillReturnRows(sqlmock.NewRows([]string{"data"}))

	res, err := (&PlaceholderImageInUseCheck{}).Run(DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.New(),
		Pipeline:  "design",
		AgentType: "test",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(res.Findings) != 1 {
		t.Fatalf("want 1 finding (detection must proceed past the variant-asset false positive), got %d",
			len(res.Findings))
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item filed, got %d", len(res.WorkItems))
	}
	if res.WorkItems[0].HandlerAgent != "image-build-handler" {
		t.Errorf("want the generation route, got handler %q", res.WorkItems[0].HandlerAgent)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet or unexpected query: %v", err)
	}
}
