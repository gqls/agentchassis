// FILE: platform/orchestration/actions/resolve_composition_layout_action_test.go
//
// Regression test for the fix in bugs_open/113's tail (2026-09-02): this
// action used to reimplement its own classification-tag extraction
// (extractClassificationTags), which read only classData["category"] and
// classData["industry_tags"] with no fallback — the one caller, of the four
// composition resolvers, that lacked the identity/site_type derivation
// readClassificationFromContext already gives its other three callers
// (install_site_composition, resolve_composition_typography,
// resolve_composition_palette).
//
// ai-agent-orchestration.com is the real, live case this reproduces: its
// classification spec has always carried `site_type` but neither `category`
// nor `industry_tags` (a legacy classifier-output shape), while its identity
// spec carries real industry/sub_industry data that was sitting unused. The
// old code resolved with zero tags, fired `resolveLayoutByTags`'s hard
// fallback, and queued a needs_new_layout_candidate HITL item that then sat
// unactioned for three weeks.

package actions

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

const resolveLayoutTestSiteID = "2bafad00-0000-0000-0000-000000000001"

// TestResolveCompositionLayout_DerivesTagsFromIdentityWhenClassificationIsBare
// reproduces ai-agent-orchestration.com's real spec shapes: classification
// carries site_type only, identity carries real industry/sub_industry. Before
// the fix, this combination produced empty site_tags and an is_fallback
// resolution (MUTATION: reverting the call in ResolveCompositionLayoutAction
// back to the old extractClassificationTags makes this test fail — it drops
// straight to fallbackLayout, which queries a *different* single-row shape
// than the layouts table this test mocks, so the test fails loudly rather
// than passing vacuously).
func TestResolveCompositionLayout_DerivesTagsFromIdentityWhenClassificationIsBare(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// 1. Domain lookup (best-effort, for logging).
	mock.ExpectQuery(`SELECT domain FROM sites WHERE id = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).
			AddRow("ai-agent-orchestration.com"))

	// 2. readClassificationFromContext: classification spec — site_type only,
	// no category, no industry_tags. This is the real, live shape.
	mock.ExpectQuery(`SELECT data FROM site_specs`).
		WithArgs(sqlmock.AnyArg(), "classification").
		WillReturnRows(sqlmock.NewRows([]string{"data"}).
			AddRow(`{"site_type":"brochure","suggested_style":"professional-dark"}`))

	// 3. readClassificationFromContext step 2: identity spec — real data.
	mock.ExpectQuery(`SELECT data FROM site_specs`).
		WithArgs(sqlmock.AnyArg(), "identity").
		WillReturnRows(sqlmock.NewRows([]string{"data"}).
			AddRow(`{"industry":"Technology Services","sub_industry":"AI Infrastructure & Enterprise Software Engineering"}`))

	// 4. deriveSiteScheme: design_intent (not present).
	mock.ExpectQuery(`SELECT data FROM site_specs`).
		WithArgs(sqlmock.AnyArg(), "design_intent").
		WillReturnError(sql.ErrNoRows)

	// 5. deriveSiteScheme: classification again, for suggested_style.
	mock.ExpectQuery(`SELECT data FROM site_specs`).
		WithArgs(sqlmock.AnyArg(), "classification").
		WillReturnRows(sqlmock.NewRows([]string{"data"}).
			AddRow(`{"site_type":"brochure","suggested_style":"professional-dark"}`))

	// 6. The read-only transaction resolveLayoutByTags runs inside.
	mock.ExpectBegin()
	mock.ExpectQuery(`FROM layouts`).
		WillReturnRows(sqlmock.NewRows(
			[]string{"id", "name", "category", "industry_tags", "scheme", "description"}).
			AddRow("11111111-1111-1111-1111-111111111111", "technical-precise", "brochure",
				`["technology","services","engineering","software"]`, "dark", "engineered SaaS layout").
			AddRow("22222222-2222-2222-2222-222222222222", "brochure-formal", "brochure",
				`[]`, "", "generic fallback layout"))
	mock.ExpectCommit()

	params := ActionParams{
		Context: context.Background(),
		Logger:  zap.NewNop(),
		DB:      db,
		CollectedData: map[string]interface{}{
			"site_id": resolveLayoutTestSiteID,
		},
		StepConfig: models.Step{Config: map[string]interface{}{}},
		ExecutionContext: &orchtypes.ExecutionContext{
			OrchestrationID: "33333333-3333-3333-3333-333333333333",
			StepName:        "resolve_layout",
			Action:          "execute",
		},
	}

	result, err := ResolveCompositionLayoutAction(context.Background(), params)
	if err != nil {
		t.Fatalf("ResolveCompositionLayoutAction returned error: %v", err)
	}

	out, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	tags, _ := out["site_tags"].([]string)
	if len(tags) == 0 {
		t.Fatalf("site_tags is empty — classification-lacks-tags case did not fall back " +
			"to identity derivation; this is the exact regression the fix addresses")
	}

	foundTech := false
	for _, tag := range tags {
		if tag == "technology" {
			foundTech = true
		}
	}
	if !foundTech {
		t.Errorf("site_tags = %v — expected a tag derived from identity.industry "+
			"(\"Technology Services\" -> \"technology\"), none found", tags)
	}

	if isFallback, _ := out["is_fallback"].(bool); isFallback {
		t.Errorf("is_fallback = true with derived tags in play — expected the matcher "+
			"to score the mocked technical-precise layout instead; reason=%v", out["reason"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
