// FILE: platform/orchestration/actions/discovery_checks/check_undeployed_assets_test.go
//
// Guards for bugs_open/142. The happy path is not what matters here — the two
// defects this check was filed for both left every happy-path assertion green:
//
//   - the brand-head half looked in page_components, so it fired on 12 of 14
//     live sites whose artefacts were serving 200;
//   - its population was the assets table, so the 2 sites that genuinely had no
//     artefact could not be examined at all.
//
// Each of those is one line, and each of the lines that fixes it can be
// "tidied" back out by someone making the two halves look consistent. So they
// are asserted directly, with the reason attached.

package discovery_checks

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/storage"
)

func newUndeployedCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.MustParse("00ff3af5-dad8-4770-9f70-3edc267a3c92"),
		Pipeline:  "design",
		AgentType: "design-discovery-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	}, mock
}

func assetRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "purpose", "asset_type", "url"})
}

// expectPageAssetQuery stubs half 1 (page-referenced assets) returning nothing.
func expectPageAssetQuery(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("FROM assets a").WillReturnRows(assetRows())
}

// expectPopulationQuery stubs the brand-head population gate.
func expectPopulationQuery(mock sqlmock.Sqlmock, deployedComponents int, hasLogo bool) {
	mock.ExpectQuery("SELECT[\\s\\S]*count\\(\\*\\) FROM pages p").
		WillReturnRows(sqlmock.NewRows([]string{"count", "exists"}).
			AddRow(deployedComponents, hasLogo))
}

// expectPurposeQuery stubs the per-purpose existence probe.
// Columns: any active row · a row AT THE PUBLISHED PATH · head reference · sample url.
func expectPurposeQuery(mock sqlmock.Sqlmock, hasAsset, headRefs bool) {
	mock.ExpectQuery("FROM site_components sc").
		WillReturnRows(sqlmock.NewRows([]string{"exists", "exists", "exists", "url"}).
			AddRow(hasAsset, hasAsset, headRefs, publishedOrNull(hasAsset)))
}

// expectPurposeQueryProvenanceMismatch is the gamesdesign.co.uk / robot-hands.com
// shape: an active row exists, but it records the unresolved template literal
// instead of the published path.
func expectPurposeQueryProvenanceMismatch(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("FROM site_components sc").
		WillReturnRows(sqlmock.NewRows([]string{"exists", "exists", "exists", "url"}).
			AddRow(true, false, true, "/assets/images/input-data.asset-key.jpg"))
}

func publishedOrNull(hasAsset bool) interface{} {
	if hasAsset {
		return "/assets/images/favicon.png"
	}
	return nil
}

// TestBrandHeadGapIsRaisedWhenNoAssetRowExists is idea.uk as it actually was on
// 2026-07-31: 61 deployed page components, an active logo, NO og_card/favicon
// asset row, both files 404 on the wire, and a site head already advertising
// the card. Before this change the check could not see it — the site had no
// `assets` row, and the query's population was the `assets` table.
func TestBrandHeadGapIsRaisedWhenNoAssetRowExists(t *testing.T) {
	dctx, mock := newUndeployedCtx(t)
	expectPageAssetQuery(mock)
	expectPopulationQuery(mock, 61, true)
	expectPurposeQuery(mock, false, true) // favicon: no row, head points at it
	expectPurposeQuery(mock, false, true) // og_card: same

	res, err := (&UndeployedAssetsCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 2 {
		t.Fatalf("want 2 work items (favicon + og_card), got %d", len(res.WorkItems))
	}
	for _, wi := range res.WorkItems {
		if wi.ItemType != "needs_brand_head_assets" {
			t.Errorf("item_type = %q, want needs_brand_head_assets", wi.ItemType)
		}
		if wi.HandlerAgent != "asset-deployer" {
			t.Errorf("handler = %q, want asset-deployer (the agent that owns derive_brand_head_assets)", wi.HandlerAgent)
		}
		if wi.Status != "detected" {
			t.Errorf("status = %q, want detected — writing 'triaged' direct from a detector is bugs_open/083's landmine", wi.Status)
		}
		if !strings.Contains(wi.Summary, "The site head already points at it") {
			t.Errorf("summary should say the head advertises the missing artefact, got %q", wi.Summary)
		}
		// The claim must be about the RECORD, not about history. recordDerivedAsset
		// is best-effort (it warns and continues on failure), so a committed,
		// 200-serving artefact can have no row — "never generated" would be an
		// over-claim this check cannot support. Council gate 35d88a60,
		// bug_historian seat.
		if strings.Contains(wi.Summary, "never been generated") {
			t.Errorf("summary asserts an unverifiable history claim: %q\n"+
				"recordDerivedAsset is best-effort, so an absent row does NOT prove the artefact was "+
				"never derived. State the absence of the record, which is what was measured.", wi.Summary)
		}
		if !strings.Contains(wi.Summary, "No asset record publishes") {
			t.Errorf("summary should assert the absence of the RECORD, got %q", wi.Summary)
		}
	}
	// Sorted order is load-bearing: Go map iteration is randomised and the
	// item_key must be stable across runs or dedup breaks.
	if got := res.WorkItems[0].ItemKey; got != "needs_brand_head_assets:favicon" {
		t.Errorf("first item_key = %q, want needs_brand_head_assets:favicon (sorted)", got)
	}
}

// TestBrandHeadGapIsSilentWhenTheAssetRowExists is the false positive this bug
// was filed for. fundamentallyai.com, gamesdesign.co.uk, vonc.com and 9 others
// each have an active og_card row and serve og-card.png 200 — and the shipping
// check raised `undeployed_asset` for every one of them, for ever, because it
// looked for the reference in page_components where it can never be.
func TestBrandHeadGapIsSilentWhenTheAssetRowExists(t *testing.T) {
	dctx, mock := newUndeployedCtx(t)
	expectPageAssetQuery(mock)
	expectPopulationQuery(mock, 40, true)
	expectPurposeQuery(mock, true, true) // favicon: row exists
	expectPurposeQuery(mock, true, true) // og_card: row exists

	res, err := (&UndeployedAssetsCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("a site whose brand-head artefacts are present must raise nothing; got %d items: %+v",
			len(res.WorkItems), res.WorkItems)
	}
}

// TestProvenanceMismatchIsObservedButNeverFiled.
//
// gamesdesign.co.uk and robot-hands.com each hold exactly one active favicon and
// one active og_card row, both carrying the unresolved template literal
// `/assets/images/input-data.asset-key.jpg` rather than the published path — and
// both sites serve favicon.png and og-card.png 200 on the wire (probed
// 2026-07-31). So the artefact IS there and the row's provenance is wrong.
//
// Filing "has never been generated" against them would be a FALSE claim, and
// this check exists because the old one made false claims. It must surface the
// state and raise nothing.
func TestProvenanceMismatchIsObservedButNeverFiled(t *testing.T) {
	dctx, mock := newUndeployedCtx(t)
	expectPageAssetQuery(mock)
	expectPopulationQuery(mock, 66, true)
	expectPurposeQueryProvenanceMismatch(mock) // favicon
	expectPurposeQueryProvenanceMismatch(mock) // og_card

	res, err := (&UndeployedAssetsCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("a row with an unexpected url is evidence of NEITHER deployment nor absence — "+
			"raising a gap would assert something false; got %d items: %+v", len(res.WorkItems), res.WorkItems)
	}
	if len(res.Findings) != 2 {
		t.Fatalf("want 2 observations recorded, got %d", len(res.Findings))
	}
	for _, f := range res.Findings {
		if f["observation"] != "brand_head_provenance_url_unexpected" {
			t.Errorf("finding must be recorded as an observation, got %+v", f)
		}
		if f["actual_url"] != "/assets/images/input-data.asset-key.jpg" {
			t.Errorf("the observation must carry the url it actually saw, got %v", f["actual_url"])
		}
	}
}

// TestBrandHeadGapNeedsADeployedSurface: loancalculator.co.uk has pages but zero
// deployed components and no logo. Nothing is public, so there is nothing to
// serve a card from and no finding to file. Without this gate the check would
// file two items against every scaffold row in the fleet.
func TestBrandHeadGapNeedsADeployedSurface(t *testing.T) {
	dctx, mock := newUndeployedCtx(t)
	expectPageAssetQuery(mock)
	expectPopulationQuery(mock, 0, false)

	res, err := (&UndeployedAssetsCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("undeployed site must raise nothing, got %d", len(res.WorkItems))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the per-purpose probes must not run once the population gate fails: %v", err)
	}
}

// TestBlockedHandlerPreconditionIsNamedInTheItem. derive_brand_head_assets
// returns {"derived": false, "reason": "no active logo asset"} when the site has
// no active asset_key='logo' row. The finding is still raised — a check that
// silently drops the case it cannot route is the failing-branch-with-no-branch
// shape this file was filed about — but it must PREDICT the refusal rather than
// let the handler discover it (bugs_open/083 fix candidate 4).
func TestBlockedHandlerPreconditionIsNamedInTheItem(t *testing.T) {
	dctx, mock := newUndeployedCtx(t)
	expectPageAssetQuery(mock)
	expectPopulationQuery(mock, 12, false) // deployed, but no logo to derive from
	expectPurposeQuery(mock, false, false)
	expectPurposeQuery(mock, false, false)

	res, err := (&UndeployedAssetsCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 2 {
		t.Fatalf("want 2 items, got %d", len(res.WorkItems))
	}
	for _, wi := range res.WorkItems {
		if !strings.Contains(wi.Summary, "no active logo asset") {
			t.Errorf("summary must name the handler's blocking precondition, got %q", wi.Summary)
		}
	}
}

// TestBrandHeadPurposesAreExcludedFromThePageAssetQuery pins the exclusion. If
// it is removed, every site's favicon and og_card become permanent false
// positives again — which is exactly the state bugs_open/142 measured at 96
// findings across 14 of 14 sites.
func TestBrandHeadPurposesAreExcludedFromThePageAssetQuery(t *testing.T) {
	dctx, mock := newUndeployedCtx(t)
	mock.ExpectQuery(regexp.QuoteMeta(`NOT (COALESCE(a.purpose, '') = ANY($2::text[]))`)).
		WillReturnRows(assetRows())
	expectPopulationQuery(mock, 0, false)

	if _, err := (&UndeployedAssetsCheck{}).Run(dctx); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the page-asset query must exclude brand-head purposes: %v", err)
	}
}

// TestUnderscoreWildcardIsLoadBearing.
//
// READ THIS BEFORE "FIXING" THE LIKE PATTERN. `purpose` values contain
// underscores and SQL LIKE treats `_` as ANY CHARACTER, so the pattern built
// from `content_hero` matches the published filename `content-hero…`, which is
// spelled with a HYPHEN. Escaping the underscore is the obvious correctness fix
// and it is wrong: measured 2026-07-31, 38 of 38 `content_hero` assets read as
// deployed with the wildcard live and 0 of 38 with it escaped, so escaping
// manufactures 38 false findings.
//
// This asserts the raw concatenation survives. A test that merely exercised the
// query would stay green through the change, which is why it asserts the SQL.
func TestUnderscoreWildcardIsLoadBearing(t *testing.T) {
	src, err := os.ReadFile("check_undeployed_assets.go")
	if err != nil {
		t.Fatalf("read own source: %v", err)
	}
	body := string(src)

	const raw = `LIKE '%/assets/images/' || COALESCE(a.purpose, '') || '.%'`
	if !strings.Contains(body, raw) {
		t.Errorf("the unescaped LIKE concatenation is gone.\n"+
			"If it was escaped (replace(purpose,'_','\\_') or ESCAPE), that is the bugs_open/142 trap:\n"+
			"`_` matches the hyphen in the real filename, and escaping it turns 38 silent content_hero\n"+
			"assets into 38 false findings. Re-measure before changing this.\nwant substring: %s", raw)
	}
	for _, banned := range []string{`replace(a.purpose`, `ESCAPE '\'`, `'\_'`} {
		if strings.Contains(body, banned) {
			t.Errorf("found %q — the LIKE wildcard was escaped; see this test's comment", banned)
		}
	}
}

// TestSiteComponentsAreNotFilteredOnDeployedStatus.
//
// site_components.build_status is 'rendered' and NEVER 'deployed' — all 42 rows
// (3 slots x 14 sites) measured 2026-07-31. The page_components predicate five
// lines above DOES filter build_status='deployed', so "make the two consistent"
// is a natural and completely silent way to make the head probe match nothing.
func TestSiteComponentsAreNotFilteredOnDeployedStatus(t *testing.T) {
	src, err := os.ReadFile("check_undeployed_assets.go")
	if err != nil {
		t.Fatalf("read own source: %v", err)
	}
	body := string(src)
	idx := strings.Index(body, "FROM site_components sc")
	if idx == -1 {
		t.Fatal("the site_components probe is gone — the brand-head half no longer reads the head at all")
	}
	stanza := body[idx:]
	if end := strings.Index(stanza, "`"); end != -1 {
		stanza = stanza[:end]
	}
	if strings.Contains(stanza, "build_status") {
		t.Error("the site_components probe filters on build_status.\n" +
			"Those rows are 'rendered', never 'deployed' — this probe now matches nothing, silently.")
	}
}

// TestBrandHeadAssetPathsMatchTheDeriver is the drift sensor for the one
// declaration of these filenames.
//
// `derive_brand_head_assets_action.go` publishes favicon.png and og-card.png and
// spells both as literals; storage.BrandHeadAssetPaths carries the same strings
// so this check can ask "is it deployed?" without reconstructing a filename from
// the purpose (which is what got bugs_open/142 wrong — `og_card` publishes as
// `og-card.png`).
//
// The deriver is deliberately NOT edited to read the map: it was being actively
// worked for bugs_open/143 when this landed, and two sessions in one file is the
// one collision no hook can prevent. So the contract is pinned by reading its
// source instead — if it ever publishes a brand-head path the map does not
// carry, the build fails here rather than the check going quietly wrong.
func TestBrandHeadAssetPathsMatchTheDeriver(t *testing.T) {
	const deriver = "../derive_brand_head_assets_action.go"
	src, err := os.ReadFile(deriver)
	if err != nil {
		t.Fatalf("read %s: %v — if it moved, repoint this test rather than deleting it", deriver, err)
	}

	// The paths the deriver records against an asset row, e.g.
	//   recordDerivedAsset(ctx, params.DB, siteID, "og_card", "/assets/images/og-card.png", logger)
	recordRe := regexp.MustCompile(`recordDerivedAsset\([^)]*?"([a-z_]+)",\s*"(/[^"]+)"`)
	matches := recordRe.FindAllStringSubmatch(string(src), -1)
	if len(matches) == 0 {
		t.Fatalf("found no recordDerivedAsset call sites in %s — this sensor is now vacuous "+
			"and storage.BrandHeadAssetPaths is unpinned", deriver)
	}

	for _, m := range matches {
		purpose, path := m[1], m[2]
		want, ok := storage.BrandHeadAssetPaths[purpose]
		if !ok {
			t.Errorf("%s records a brand-head asset for purpose %q that storage.BrandHeadAssetPaths "+
				"does not carry.\nAdd it there, or this check cannot tell whether it is deployed.", deriver, purpose)
			continue
		}
		if want != path {
			t.Errorf("path drift for %q: the deriver publishes %q, storage.BrandHeadAssetPaths says %q.\n"+
				"These must be the same string — the check tests for the map's value in the rendered head.",
				purpose, path, want)
		}
	}

	if len(matches) != len(storage.BrandHeadAssetPaths) {
		t.Errorf("the deriver records %d brand-head assets but the map carries %d.\n"+
			"A map entry with no producer means this check will file a permanent, unfixable finding.",
			len(matches), len(storage.BrandHeadAssetPaths))
	}
}

// TestBrandHeadItemSpecCarriesTheRoutingMode pins the one spec key the whole
// dispatch depends on. asset-deployer's entry chain keys every conditional on
// spec.mode; an item without it falls through to deploy_image_asset, whose
// brand-head guard refuses it, and the refusal-as-result stamps the item
// 'complete' with nothing derived (bugs_open/131, 2026-08-17 contribution: 21
// items complete, four sites' artefacts 404 on the wire). The key looks
// redundant next to spec.purpose — it is not; it is the address.
func TestBrandHeadItemSpecCarriesTheRoutingMode(t *testing.T) {
	dctx, mock := newUndeployedCtx(t)
	expectPageAssetQuery(mock)
	expectPopulationQuery(mock, 61, true)
	expectPurposeQuery(mock, false, true)
	expectPurposeQuery(mock, false, true)

	res, err := (&UndeployedAssetsCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 2 {
		t.Fatalf("want 2 work items, got %d", len(res.WorkItems))
	}
	for _, wi := range res.WorkItems {
		var spec map[string]interface{}
		if err := json.Unmarshal([]byte(wi.SpecJSON), &spec); err != nil {
			t.Fatalf("spec is not JSON: %v", err)
		}
		if mode, _ := spec["mode"].(string); mode != "brand_head" {
			t.Errorf("spec.mode = %q, want brand_head — without it asset-deployer's chain "+
				"routes this item to a branch that can only refuse it, and the refusal completes "+
				"as success (bugs_open/131)", mode)
		}
		if purpose, _ := spec["purpose"].(string); purpose == "" {
			t.Errorf("spec.purpose is empty — the verifier and any reader of the item need it")
		}
	}
}

// TestBrandHeadVerifierIsRegisteredWithRemit pins the registration itself: the
// verifier must exist (so completion is gated at all) and must declare Grades
// (bugs_open/213 — two spec shapes already file under this item_type: the
// discovery producer's purpose shape and the hand-filed mode='brand_head'
// redrive shape).
func TestBrandHeadVerifierIsRegisteredWithRemit(t *testing.T) {
	v, _ := GetVerifier("needs_brand_head_assets")
	if v == nil {
		t.Fatal("no verifier registered for needs_brand_head_assets — completion is ungated " +
			"and a deploy-guard refusal will stamp items complete again (bugs_open/131)")
	}
	if !VerifierDeclaresRemit("needs_brand_head_assets") {
		t.Error("verifier declares no Grades remit — the next producer converging on this " +
			"item_type reproduces bugs_open/213")
	}
}

func brandHeadVerifyTarget(spec map[string]interface{}) VerifyTarget {
	return VerifyTarget{
		ItemID:   uuid.New(),
		SiteID:   uuid.MustParse("00ff3af5-dad8-4770-9f70-3edc267a3c92"),
		ItemType: "needs_brand_head_assets",
		Spec:     spec,
	}
}

// expectVerifyQuery stubs the verifier's two-EXISTS probe.
// Columns: a row AT THE PUBLISHED PATH · any active row.
func expectVerifyQuery(mock sqlmock.Sqlmock, atPath, anyRow bool) {
	mock.ExpectQuery("EXISTS[\\s\\S]*FROM assets a").
		WillReturnRows(sqlmock.NewRows([]string{"at_path", "any_row"}).AddRow(atPath, anyRow))
}

// TestVerifyBrandHeadBlocksCompletionWhenNothingWasDerived is the mis-route
// this verifier exists to catch: the handler run being completed was a
// deploy_image_asset refusal, nothing was derived, no row exists. Resolved
// must be false — before this verifier, that exact state stamped 'complete'
// 21 times.
func TestVerifyBrandHeadBlocksCompletionWhenNothingWasDerived(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	expectVerifyQuery(mock, false, false)

	res, err := VerifyBrandHeadAssetsResolved(context.Background(), db,
		brandHeadVerifyTarget(map[string]interface{}{"purpose": "og_card"}), zap.NewNop())
	if err != nil {
		t.Fatalf("verifier errored where it should have returned Resolved:false: %v", err)
	}
	if res.Resolved {
		t.Fatal("Resolved=true with no active row at the published path — this is the exact " +
			"false completion bugs_open/131 measured 21 times")
	}
	if !strings.Contains(res.Detail, "og_card") {
		t.Errorf("detail should name the purpose, got %q", res.Detail)
	}
}

// TestVerifyBrandHeadPassesOnTheDeriversOwnRecord: an active row recording the
// published path is the deriver's post-commit write — the same deploy evidence
// the producing check accepts (see the file header's "why the row is the
// evidence").
func TestVerifyBrandHeadPassesOnTheDeriversOwnRecord(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	expectVerifyQuery(mock, true, true)

	res, err := VerifyBrandHeadAssetsResolved(context.Background(), db,
		brandHeadVerifyTarget(map[string]interface{}{"purpose": "favicon"}), zap.NewNop())
	if err != nil {
		t.Fatalf("verifier errored on the resolved case: %v", err)
	}
	if !res.Resolved {
		t.Fatalf("Resolved=false with the deriver's own record present: %s", res.Detail)
	}
}

// TestVerifyBrandHeadCannotTellOnProvenanceMismatch: an active row at a
// DIFFERENT url is evidence of neither deployment nor absence (the
// bugs_open/152 shape; the producing check observes-and-declines there). The
// verifier must return an ERROR — fail-closed under RFC_017 — never a verdict
// in either direction.
func TestVerifyBrandHeadCannotTellOnProvenanceMismatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	expectVerifyQuery(mock, false, true)

	_, err = VerifyBrandHeadAssetsResolved(context.Background(), db,
		brandHeadVerifyTarget(map[string]interface{}{"purpose": "favicon"}), zap.NewNop())
	if err == nil {
		t.Fatal("expected a cannot-verify error on the provenance-mismatch shape; a verdict " +
			"either way would be a claim the row cannot support")
	}
	if !strings.Contains(err.Error(), "cannot verify") {
		t.Errorf("error should say verification could not decide, got: %v", err)
	}
}

// TestVerifyBrandHeadModeShapeChecksEveryPurpose: the hand-filed redrive shape
// (spec.mode='brand_head', no purpose) means "derive the lot" — the deriver
// regenerates every brand-head artefact in one run — so completion must be
// evidenced for every purpose, in the stable sorted order.
func TestVerifyBrandHeadModeShapeChecksEveryPurpose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	for range brandHeadPurposes() {
		expectVerifyQuery(mock, true, true)
	}

	res, err := VerifyBrandHeadAssetsResolved(context.Background(), db,
		brandHeadVerifyTarget(map[string]interface{}{"mode": "brand_head"}), zap.NewNop())
	if err != nil {
		t.Fatalf("verifier errored: %v", err)
	}
	if !res.Resolved {
		t.Fatalf("Resolved=false with every purpose evidenced: %s", res.Detail)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the mode shape must probe every brand-head purpose, not the first: %v", err)
	}
}

// TestGradesBrandHeadAssets pins the remit: both live spec shapes are graded,
// anything else is disclaimed with a reason an operator can act on.
func TestGradesBrandHeadAssets(t *testing.T) {
	cases := []struct {
		name   string
		spec   map[string]interface{}
		speaks bool
	}{
		{"discovery shape (purpose)", map[string]interface{}{"purpose": "og_card", "check": "undeployed_assets"}, true},
		{"redrive shape (mode only)", map[string]interface{}{"mode": "brand_head"}, true},
		{"foreign purpose", map[string]interface{}{"purpose": "hero"}, false},
		{"empty spec", map[string]interface{}{}, false},
	}
	for _, c := range cases {
		speaks, why := gradesBrandHeadAssets(brandHeadVerifyTarget(c.spec))
		if speaks != c.speaks {
			t.Errorf("%s: speaks=%v, want %v (%s)", c.name, speaks, c.speaks, why)
		}
		if !speaks && why == "" {
			t.Errorf("%s: a disclaimer must say why — the reason is shown to the operator "+
				"on the blocked item", c.name)
		}
	}
}
