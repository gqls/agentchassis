// FILE: platform/orchestration/actions/discovery_checks/check_image_url_404_test.go
//
// bugs_open/128 — pins the FIXED predicate. Replaces
// check_image_url_404_masking_test.go, which pinned the defect deliberately and
// told the fixing session what it had to answer first. Both of its conditions
// were answered before this file was written:
//
//   - "a path-based predicate was measured and REFUTED — only 9 of 79 masked
//     paths have an assets row whose url/filename carries the basename".
//     True, and it is the wrong comparison. assets.url is an expiring presigned
//     S3 URL and filename/storage_path are mostly empty, so neither can say where
//     a file is SERVED from. storage.DeployedWebPath(asset_key, purpose) can, and
//     already does for every writer in the platform. Re-measured 2026-07-31 over
//     all 127 rendered image paths on 13 live sites with HTTP as ground truth:
//     the new predicate has 1 false positive and 0 misses, against the purpose
//     skip's 21 and 6.
//
//   - "unmasking activates knownPurposeMapping routing that has NEVER fired in
//     production." It cannot now: that branch is gone. It duplicated
//     check_placeholder_image_in_use exactly — same two paths, same purposes,
//     same needs_hero_image/needs_logo item types, same image-build-handler, same
//     precondition, both enabled on design-discovery-agent — and neither had ever
//     fired, for the same reason. TestImageURL404_NeverRoutesToAHandler is what
//     keeps it gone.
//
// The one residual false positive is deliberate and named in the check's header:
// a file committed by no asset row (webdesign.co.uk's legacy hero.jpg, 455KB,
// serving 200) is reported, because the database does not know it exists.

package discovery_checks

import (
	"context"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type assetRow struct{ key, purpose string }

// runImageURL404 drives the real check against mocked rows for both rendered
// surfaces and the site's active assets. Query order matters: page components,
// then chrome, then assets (only when a reference was found).
func runImageURL404(t *testing.T, pageHTML, chromeHTML string, assets []assetRow) *CheckResult {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageRows := sqlmock.NewRows([]string{"rendered_html"})
	if pageHTML != "" {
		pageRows.AddRow(pageHTML)
	}
	mock.ExpectQuery("FROM page_components").WillReturnRows(pageRows)

	chromeRows := sqlmock.NewRows([]string{"rendered_html"})
	if chromeHTML != "" {
		chromeRows.AddRow(chromeHTML)
	}
	mock.ExpectQuery("FROM site_components").WillReturnRows(chromeRows)

	arows := sqlmock.NewRows([]string{"asset_key", "purpose"})
	for _, a := range assets {
		arows.AddRow(a.key, a.purpose)
	}
	mock.ExpectQuery("FROM assets").WillReturnRows(arows)

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

func itemKeys(res *CheckResult) []string {
	out := make([]string, 0, len(res.WorkItems))
	for _, w := range res.WorkItems {
		out = append(out, w.ItemKey)
	}
	return out
}

// THE BUG, now reported. vonc.com owns six active hero assets and renders
// /assets/images/hero.jpg, which 404s. Under the purpose skip the site's OWN hero
// assets made that path unreportable; under the path predicate none of them
// deploys to it, so it is a finding. Six live sites were in this state.
func TestImageURL404_ReportsAPathNoActiveAssetDeploysTo(t *testing.T) {
	res := runImageURL404(t,
		`<div><img src="/assets/images/hero.jpg" alt="x"></div>`, "",
		[]assetRow{
			{"hero_home", "hero"}, {"hero_archetypes", "hero"}, {"icon_judge", "icon"},
		})

	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 work item, got %d (%v)", len(res.WorkItems), itemKeys(res))
	}
	w := res.WorkItems[0]
	if w.ItemType != "image_url_404" {
		t.Errorf("item_type = %q, want image_url_404", w.ItemType)
	}
	if w.ItemKey != "image_url_404:hero.jpg" {
		t.Errorf("item_key = %q, want image_url_404:hero.jpg", w.ItemKey)
	}
	if !strings.Contains(w.Summary, "no active asset deploys to that path") {
		t.Errorf("summary must say what the evidence IS, got %q", w.Summary)
	}
}

// THE REFUTATION THE OLD TEST CARRIED, honoured. 73 of the 79 masked paths serve
// 200, and a naive "does an assets row mention this basename" predicate would have
// flagged them. DeployedWebPath resolves each of these exactly, so all four are
// silent — hero variants, icon variants, content heroes and cards, which between
// them are most of the fleet's rendered imagery.
func TestImageURL404_SilentOnEveryPathAnAssetActuallyDeploysTo(t *testing.T) {
	res := runImageURL404(t, `
		<img src="/assets/images/hero-home.jpg">
		<img src="/assets/images/icon-catalyst.jpg">
		<img src="/assets/images/content-hero-tungsten-guide.jpg">
		<img src="/assets/images/card-barrel-weight.jpg">`, "",
		[]assetRow{
			{"hero_home", "hero"},
			{"icon_catalyst", "icon"},
			{"content_hero_tungsten_guide", "content_hero"},
			{"card_barrel_weight", "card"},
		})

	if len(res.WorkItems) != 0 {
		t.Fatalf("a working path must not be reported; got %v", itemKeys(res))
	}
}

// THE FLEET-WIDE FALSE POSITIVE THIS NEARLY SHIPPED. og_card publishes as
// og-card.png with a HYPHEN; DeployedWebPath alone returns og_card.png with an
// underscore, so routing brand-head purposes through it would report a 404 for
// the og card and the favicon of every site in the fleet — both referenced from
// the head, so on every page. storage.BrandHeadAssetPaths is the branch, and this
// is the test that keeps it.
func TestImageURL404_BrandHeadPathsResolveThroughTheirOwnMap(t *testing.T) {
	chrome := `<link rel="icon" href="/assets/images/favicon.png">` +
		`<meta property="og:image" content="/assets/images/og-card.png">`

	res := runImageURL404(t, "", chrome,
		[]assetRow{{"favicon", "favicon"}, {"og_card", "og_card"}})
	if len(res.WorkItems) != 0 {
		t.Fatalf("brand-head assets exist; nothing should be reported, got %v", itemKeys(res))
	}

	// idea.uk: the chrome references both and the site owns neither. Both 404 on
	// every page, and no finding existed before the chrome surface was scanned.
	res = runImageURL404(t, "", chrome, []assetRow{{"hero_home", "hero"}})
	if len(res.WorkItems) != 2 {
		t.Fatalf("want 2 work items, got %d (%v)", len(res.WorkItems), itemKeys(res))
	}
	for _, w := range res.WorkItems {
		if w.Severity != "high" {
			t.Errorf("%s: severity = %q, want high — chrome is on every page", w.ItemKey, w.Severity)
		}
		if !strings.Contains(w.Summary, "chrome") {
			t.Errorf("%s: summary should name the surface, got %q", w.ItemKey, w.Summary)
		}
	}
}

// THE EXTENSION IS PART OF THE IDENTITY. fundamentallyai.com stores its logo as
// asset_key 'logo' under purpose 'hero', so it deploys to logo.JPG (200) while a
// page also references logo.PNG (404). An extension-blind dedup key — which is
// what shipped until 2026-07-31 — collapses the two into one row and lets
// idx_swi_dedup drop the real finding (the failure mode of bugs_open/091).
func TestImageURL404_ExtensionDistinguishesTwoFilesWithOneBasename(t *testing.T) {
	res := runImageURL404(t,
		`<img src="/assets/images/logo.jpg"><img src="/assets/images/logo.png">`, "",
		[]assetRow{{"logo", "hero"}})

	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 work item (logo.png only), got %d (%v)",
			len(res.WorkItems), itemKeys(res))
	}
	if got := res.WorkItems[0].ItemKey; got != "image_url_404:logo.png" {
		t.Errorf("item_key = %q, want image_url_404:logo.png — the extension is the "+
			"only thing distinguishing the broken file from the working one", got)
	}
}

// AN <img> WITH NO SOURCE. It has no path, so every path predicate in this file is
// blind to it, and an HTTP checker would score it 200 because an empty src
// resolves against the current document. Structural detection is the only kind
// that works. Six of these shipped on leopardessconsulting.co.uk/blog.html.
func TestImageURL404_EmptyImgSrcIsCountedAndReportedOnce(t *testing.T) {
	res := runImageURL404(t,
		`<img src="" alt="a" loading="lazy"><img src='  ' alt="b"><img src="#" alt="c">`, "",
		[]assetRow{{"hero_home", "hero"}})

	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 work item, got %d (%v)", len(res.WorkItems), itemKeys(res))
	}
	w := res.WorkItems[0]
	if w.ItemKey != "image_url_404:empty-src" {
		t.Errorf("item_key = %q, want image_url_404:empty-src", w.ItemKey)
	}
	if !strings.Contains(w.Summary, "3 <img>") {
		t.Errorf("the count is the finding; summary = %q", w.Summary)
	}
	if !strings.Contains(w.SpecJSON, `"kind":"empty_src"`) {
		t.Errorf("spec must carry the kind so a reader can tell the two shapes apart: %s", w.SpecJSON)
	}
}

// A real <img> with a real src must not be mistaken for an empty one — the
// negative control for the pattern above.
func TestImageURL404_APopulatedSrcIsNotAnEmptyOne(t *testing.T) {
	res := runImageURL404(t,
		`<img src="/assets/images/hero-home.jpg" alt="a">`, "",
		[]assetRow{{"hero_home", "hero"}})
	if len(res.WorkItems) != 0 {
		t.Fatalf("nothing to report; got %v", itemKeys(res))
	}
}

// THE DE-DUPLICATION INVARIANT. check_placeholder_image_in_use owns "the fallback
// path is rendered and no asset of that purpose exists, so build one" and routes
// it to image-build-handler. This check must never file a second, competing repair
// for the same condition — which is exactly what its recognised-purpose branch did
// until 2026-07-31, under a different item key so the two would not even dedup.
func TestImageURL404_NeverRoutesToAHandler(t *testing.T) {
	// hero.jpg and logo.png are precisely placeholderPathMapping's two paths, and
	// this site owns no asset of either purpose — the condition under which the
	// deleted branch used to route.
	res := runImageURL404(t,
		`<img src="/assets/images/hero.jpg"><img src="/assets/images/logo.png">`, "",
		[]assetRow{{"icon_a", "icon"}})

	if len(res.WorkItems) != 2 {
		t.Fatalf("want 2 flag-only work items, got %d (%v)", len(res.WorkItems), itemKeys(res))
	}
	for _, w := range res.WorkItems {
		if w.HandlerAgent != "" {
			t.Errorf("%s: HandlerAgent = %q — this check is flag-only; generation is "+
				"check_placeholder_image_in_use's remit and duplicating it files two "+
				"work items for one repair", w.ItemKey, w.HandlerAgent)
		}
		if w.ItemType != "image_url_404" {
			t.Errorf("%s: item_type = %q — needs_hero_image/needs_logo belong to the "+
				"placeholder check", w.ItemKey, w.ItemType)
		}
	}
}

// The chrome surface and the page surface are one population: a path referenced
// on both is one finding, and it is labelled with both surfaces.
func TestImageURL404_ChromeAndPageAreOnePopulation(t *testing.T) {
	res := runImageURL404(t,
		`<img src="/assets/images/logo.png">`,
		`<img src="/assets/images/logo.png">`,
		[]assetRow{{"hero_home", "hero"}})

	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item, got %d (%v)", len(res.WorkItems), itemKeys(res))
	}
	if !strings.Contains(res.WorkItems[0].SpecJSON, `"surface":"page+chrome"`) {
		t.Errorf("spec should record both surfaces: %s", res.WorkItems[0].SpecJSON)
	}
}

// REGRESSION GUARD, from the acceptance set bugs_open/128 specified: the five
// finetuning.uk case-study images were the half the old check got RIGHT (their
// prefix "case" is not a purpose, so nothing masked them). They are live 404s and
// must stay reported.
func TestImageURL404_KeepsTheFindingsTheOldPredicateGotRight(t *testing.T) {
	res := runImageURL404(t,
		`<img src="/assets/images/case-study-legal-rag.jpg">`, "",
		[]assetRow{{"hero_home", "hero"}, {"logo", "logo"}})

	if len(res.WorkItems) != 1 || res.WorkItems[0].ItemKey != "image_url_404:case-study-legal-rag.jpg" {
		t.Fatalf("the previously-correct finding must survive the fix; got %v", itemKeys(res))
	}
}

// Two runs over the same data must emit the same sequence: work items are
// inserted in order and a shuffled sequence makes a diff of two runs unreadable.
func TestImageURL404_EmissionOrderIsDeterministic(t *testing.T) {
	html := `<img src="/assets/images/zeta.jpg"><img src="/assets/images/alpha.jpg">` +
		`<img src="/assets/images/mid.jpg">`
	first := itemKeys(runImageURL404(t, html, "", []assetRow{{"hero_home", "hero"}}))
	for i := 0; i < 5; i++ {
		got := itemKeys(runImageURL404(t, html, "", []assetRow{{"hero_home", "hero"}}))
		if strings.Join(got, ",") != strings.Join(first, ",") {
			t.Fatalf("emission order is not stable: %v then %v", first, got)
		}
	}
	if first[0] != "image_url_404:alpha.jpg" {
		t.Errorf("expected path-sorted emission, got %v", first)
	}
}

// THE THIRD SHAPE (2026-08-03). <img src="cpu"> is what an icon name looks like
// when a template renders it into an image slot. It is not an /assets/images/…
// path and it is not empty, so both older predicates were silent — and 31 of
// these were live across 2 sites when this test was written, every one a broken
// image with no finding anywhere.
func TestImageURL404_BareWordSrcIsReportedWithItsTokens(t *testing.T) {
	res := runImageURL404(t,
		`<img src="cpu" alt="Automation Department" class="member-photo">`+
			`<img src="network" alt="Agents Department" class="member-photo">`+
			`<img src="cpu" alt="Automation Department" class="member-photo">`, "",
		[]assetRow{{"hero_home", "hero"}})

	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 work item, got %d (%v)", len(res.WorkItems), itemKeys(res))
	}
	w := res.WorkItems[0]
	if w.ItemKey != "image_url_404:bare-token-src" {
		t.Errorf("item_key = %q, want image_url_404:bare-token-src", w.ItemKey)
	}
	// Three tags, two distinct tokens: the count is occurrences, the tokens are
	// the evidence that names the cause.
	if !strings.Contains(w.Summary, "3 <img>") {
		t.Errorf("count must be occurrences, not distinct tokens; summary = %q", w.Summary)
	}
	if !strings.Contains(w.SpecJSON, `"kind":"bare_token_src"`) {
		t.Errorf("spec must carry the kind: %s", w.SpecJSON)
	}
	if !strings.Contains(w.SpecJSON, `"tokens":["cpu","network"]`) {
		t.Errorf("tokens must be distinct and sorted: %s", w.SpecJSON)
	}
	if w.Severity != "high" {
		t.Errorf("severity = %q; a template defect repeats on every page that mounts it", w.Severity)
	}
}

// THE NEGATIVE CONTROLS. The predicate excludes '/', '.' and ':', so every shape
// of legitimate src must stay silent. Without this, the cheapest way to "fix" a
// false positive is to widen the pattern until it reports nothing — and a check
// that reports nothing looks exactly like a check that found nothing.
func TestImageURL404_LegitimateSrcsAreNotBareWords(t *testing.T) {
	for _, src := range []string{
		"/assets/images/hero-home.jpg",       // rooted path (and asset-backed)
		"hero.png",                           // relative filename — has a dot
		"../images/logo.svg",                 // relative path — has slashes
		"https://cdn.example.com/a.png",      // absolute URL — has a colon
		"//cdn.example.com/a.png",            // protocol-relative — has slashes
		"data:image/png;base64,iVBORw0KGgo=", // data URI — has a colon
	} {
		res := runImageURL404(t, `<img src="`+src+`" alt="x">`, "",
			[]assetRow{{"hero_home", "hero"}})
		for _, k := range itemKeys(res) {
			if k == "image_url_404:bare-token-src" {
				t.Errorf("src=%q was reported as a bare word", src)
			}
		}
	}
}

// The two "no usable source" shapes must stay separate items: they have
// different repairs (supply a source vs fix the template that wrote one), so
// folding them into one tally would send a reader to the wrong place.
func TestImageURL404_EmptyAndBareWordAreDistinctFindings(t *testing.T) {
	res := runImageURL404(t,
		`<img src="" alt="a"><img src="cpu" alt="b">`, "",
		[]assetRow{{"hero_home", "hero"}})

	keys := itemKeys(res)
	if len(keys) != 2 {
		t.Fatalf("want 2 distinct items, got %v", keys)
	}
	want := map[string]bool{"image_url_404:empty-src": false, "image_url_404:bare-token-src": false}
	for _, k := range keys {
		if _, ok := want[k]; !ok {
			t.Errorf("unexpected item key %q", k)
		}
		want[k] = true
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("missing item key %q", k)
		}
	}
}
