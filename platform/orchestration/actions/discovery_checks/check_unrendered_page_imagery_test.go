// FILE: platform/orchestration/actions/discovery_checks/check_unrendered_page_imagery_test.go
//
// The tests that matter here are the PIN and the PREDICATE tests, not the happy
// path — the ways this check can silently stop working are:
//
//   - the SQL key mirror drifting from imageryplan.ContentHeroKey (then the
//     population is empty for ever and the check reads as a healthy fleet);
//   - the lifecycle / asset-status / composed-page arms being deleted (then it
//     files rollups about retired pages or plan states);
//   - the rollup shapes changing so the dedup key stops holding one row per
//     state, or the retraction widening past its own key.
//
// The classifier and the item builder are pure, so their tests need no DB. The
// Run test uses sqlmock in the package's house style.

package discovery_checks

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
)

// TestContentHeroKeySQLMirrorStaysTrue pins the sweep's SQL spelling
// (`'content_hero_' || replace(p.name, '-', '_')`) to the Go helper it mirrors.
// A set-based sweep cannot call the helper per row; this is the contract that
// lets it not have to. If this fails, change the SQL in
// check_unrendered_page_imagery.go to match the helper — never the reverse.
func TestContentHeroKeySQLMirrorStaysTrue(t *testing.T) {
	for _, name := range []string{
		"tool-repayment",            // the fixture page
		"index",                     // no hyphens at all
		"tool-equity-release-guide", // several hyphens
		"already_underscored",       // underscores must survive untouched
		"mixed-name_shape",          // both separators at once
	} {
		sqlMirror := "content_hero_" + strings.ReplaceAll(name, "-", "_")
		if got := imageryplan.ContentHeroKey(name); got != sqlMirror {
			t.Errorf("ContentHeroKey(%q) = %q; the SQL mirror computes %q — the sweep would miss this page",
				name, got, sqlMirror)
		}
	}
}

func TestClassifyUnrenderedImagery(t *testing.T) {
	cases := []struct {
		referenced, capable, capableNonFragment bool
		want                                    string
	}{
		{true, true, true, ""},   // fulfilled — referenced wins over everything
		{true, false, false, ""}, // referenced even with no capable slot: fulfilled
		{false, true, true, "unwired"},
		{false, true, false, "fragment_slot"},
		{false, false, false, "no_image_slot"},
		// capableNonFragment without capable is unreachable from the SQL (the
		// third EXISTS is a strict subset of the second); the classifier still
		// answers it deterministically rather than panicking.
		{false, false, true, "unwired"},
	}
	for _, c := range cases {
		if got := classifyUnrenderedImagery(c.referenced, c.capable, c.capableNonFragment); got != c.want {
			t.Errorf("classify(ref=%v cap=%v capNF=%v) = %q, want %q",
				c.referenced, c.capable, c.capableNonFragment, got, c.want)
		}
	}
}

func unrenderedTestCtx(t *testing.T) DiscoveryCheckContext {
	t.Helper()
	return DiscoveryCheckContext{
		Ctx:       context.Background(),
		SiteID:    uuid.MustParse("62b5978e-4271-4589-8e00-4baebfc0447c"),
		Pipeline:  "design",
		AgentType: "design-discovery-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	}
}

// TestBuildUnrenderedImageryResultShapes covers the rollup contract in one pass:
// one item per non-empty state under the state-scoped dedup key, flag-only
// routing, the FULL count with truncated examples, and a narrow retraction for
// each empty state.
func TestBuildUnrenderedImageryResultShapes(t *testing.T) {
	dctx := unrenderedTestCtx(t)

	// 13 unwired members: one past the example cap, so count and examples must
	// diverge — a builder that counts the examples slice passes only at <=12.
	var unwired []unrenderedImageryPage
	for i := 0; i < 13; i++ {
		unwired = append(unwired, unrenderedImageryPage{
			PageName: "page-" + string(rune('a'+i)),
			AssetKey: "content_hero_page_" + string(rune('a'+i)),
			WebPath:  "/assets/images/content-hero-page-" + string(rune('a'+i)) + ".jpg",
		})
	}
	byState := map[string][]unrenderedImageryPage{
		"unwired":       unwired,
		"fragment_slot": {{PageName: "tool-x", AssetKey: "content_hero_tool_x", WebPath: "/assets/images/content-hero-tool-x.jpg"}},
		// no_image_slot deliberately empty -> must produce a retraction, not an item
	}

	res, err := buildUnrenderedImageryResult(dctx, byState, "2026-09-02")
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if len(res.WorkItems) != 2 {
		t.Fatalf("want 2 rollups (unwired, fragment_slot), got %d", len(res.WorkItems))
	}
	for _, wi := range res.WorkItems {
		if wi.ItemType != "unrendered_page_imagery" {
			t.Errorf("ItemType = %q", wi.ItemType)
		}
		if wi.HandlerAgent != "" {
			t.Errorf("HandlerAgent = %q — flag-only items must not be claimable", wi.HandlerAgent)
		}
		if wi.Status != "needs_human_review" {
			t.Errorf("Status = %q, want needs_human_review", wi.Status)
		}
		if !strings.HasPrefix(wi.ItemKey, "unrendered_page_imagery:") {
			t.Errorf("ItemKey = %q — must be state-scoped for one-open-rollup-per-state dedup", wi.ItemKey)
		}
	}

	var spec struct {
		State      string                  `json:"state"`
		Count      int                     `json:"count"`
		MeasuredAt string                  `json:"measured_at"`
		Examples   []unrenderedImageryPage `json:"examples"`
		Remedy     string                  `json:"remedy"`
	}
	if err := json.Unmarshal([]byte(res.WorkItems[0].SpecJSON), &spec); err != nil {
		t.Fatalf("spec unmarshal: %v", err)
	}
	if spec.State != "unwired" || spec.Count != 13 {
		t.Errorf("spec state/count = %q/%d, want unwired/13 (the FULL census, not the example slice)", spec.State, spec.Count)
	}
	if len(spec.Examples) != unrenderedImageryMaxExamples {
		t.Errorf("examples = %d, want capped at %d", len(spec.Examples), unrenderedImageryMaxExamples)
	}
	// A count without its date reads as current for ever (owner ruling 2026-08-22).
	if spec.MeasuredAt != "2026-09-02" {
		t.Errorf("measured_at = %q — the census date must travel with the count", spec.MeasuredAt)
	}
	if !strings.Contains(spec.Remedy, "bugs_open/412") {
		t.Errorf("unwired remedy must route the reader to the wiring owner; got %q", spec.Remedy)
	}

	if len(res.Resolved) != 1 {
		t.Fatalf("want exactly 1 retraction (no_image_slot empty), got %d", len(res.Resolved))
	}
	r := res.Resolved[0]
	if r.ItemKey != "unrendered_page_imagery:no_image_slot" || r.AllOfType {
		t.Errorf("retraction must be NARROW, by the emptied state's own key; got key=%q allOfType=%v", r.ItemKey, r.AllOfType)
	}
	if r.Reason == "" {
		t.Errorf("retraction must carry a reason — the runner refuses it otherwise")
	}
}

// TestUnrenderedImageryAllStatesEmptyRetractsAllAndFilesNothing — the healthy
// site. Three narrow retractions, zero items, zero findings.
func TestUnrenderedImageryAllStatesEmptyRetractsAllAndFilesNothing(t *testing.T) {
	res, err := buildUnrenderedImageryResult(unrenderedTestCtx(t), map[string][]unrenderedImageryPage{}, "2026-09-02")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Findings) != 0 {
		t.Fatalf("healthy site must file nothing; got %d items, %d findings", len(res.WorkItems), len(res.Findings))
	}
	if len(res.Resolved) != len(unrenderedImageryStates) {
		t.Fatalf("want %d retractions, got %d", len(unrenderedImageryStates), len(res.Resolved))
	}
}

// TestUnrenderedImageryRunPredicates drives Run through sqlmock and asserts the
// load-bearing SQL arms directly — each is a one-line deletion that leaves the
// happy path green:
//
//   - the population must demand asset status 'active', the ContentHeroKey
//     mirror, the lifecycle arm (PostureArmed registration), and the
//     composed-page guard;
//   - the classification must test the DEPLOYER-derived web path (not a
//     re-derivation) and build its fragment predicate from
//     InteractiveStructuralMarkers.
func TestUnrenderedImageryRunPredicates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	dctx := unrenderedTestCtx(t)
	dctx.DB = db

	pageID := uuid.New()
	popRe := regexp.MustCompile(`(?s)a\.status = 'active'.*content_hero_' \|\| replace\(p\.name, '-', '_'\).*p\.status = 'active'.*pc0\.build_status IS DISTINCT FROM 'removed'`)
	mock.ExpectQuery(popRe.String()).WithArgs(dctx.SiteID).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "asset_key", "purpose"}).
			AddRow(pageID.String(), "tool-repayment", "content_hero_tool_repayment", "content_hero"),
	)

	// The classification query must carry every fragment marker; build the
	// expectation from the exported list so a marker added there is demanded
	// here without editing this test.
	classifyParts := []string{`rc\.rendered_html LIKE '%' \|\| \$2 \|\| '%'`}
	for _, m := range InteractiveStructuralMarkers {
		classifyParts = append(classifyParts, regexp.QuoteMeta(m))
	}
	classifyRe := "(?s)" + strings.Join(classifyParts, ".*")
	// referenced=false, capable=true, capableNonFragment=false -> fragment_slot
	mock.ExpectQuery(classifyRe).
		WithArgs(pageID.String(), "/assets/images/content-hero-tool-repayment.jpg").
		WillReturnRows(sqlmock.NewRows([]string{"referenced", "capable", "capable_nonfragment"}).
			AddRow(false, true, false))

	res, err := (&UnrenderedPageImageryCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 rollup, got %d", len(res.WorkItems))
	}
	if res.WorkItems[0].ItemKey != "unrendered_page_imagery:fragment_slot" {
		t.Errorf("ItemKey = %q, want the fragment_slot rollup", res.WorkItems[0].ItemKey)
	}
	// The other two states were computed empty on a CLEAN run -> narrow retractions.
	if len(res.Resolved) != 2 {
		t.Errorf("want 2 retractions (unwired, no_image_slot), got %d", len(res.Resolved))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}

	// The WithArgs above is the deploy-path pin: Run must have asked about
	// storage.DeployedWebPath's answer for (content_hero_tool_repayment,
	// content_hero) — the hyphenated filename the DEPLOYER publishes. A
	// re-derivation from the purpose alone (the GAP 1 defect, register IMG-072)
	// would have asked about /assets/images/content-hero.jpg and failed the
	// argument match.
}

// TestUnrenderedImageryPopulationErrorFilesAndRetractsNothing — RFC_010's
// safety property from this check's side: an errored run must not reach the
// retraction path (the runner also skips Resolved on error; this pins that the
// check itself returns the error rather than a half-empty result).
func TestUnrenderedImageryPopulationErrorFilesAndRetractsNothing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	dctx := unrenderedTestCtx(t)
	dctx.DB = db

	mock.ExpectQuery(`FROM pages p`).WithArgs(dctx.SiteID).
		WillReturnError(context.DeadlineExceeded)

	res, err := (&UnrenderedPageImageryCheck{}).Run(dctx)
	if err == nil {
		t.Fatalf("want an error from a failed population sweep, got result %+v", res)
	}
}
