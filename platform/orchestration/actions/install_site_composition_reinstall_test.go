// FILE: platform/orchestration/actions/install_site_composition_reinstall_test.go
//
// bugs_open/113 — the platform could COMPOSE a site that had nothing and could
// not RE-COMPOSE one that had the wrong thing.
//
// ai-agent-orchestration.com shared the LIGHT seed collection `professional-dark`
// (card_bg #ffffff) while its own spec was dark (#080B10), so 38 of 58 measured
// contrast failures on three pages were ink on a white card. The detector filed
// it correctly (work item 47ce091c, "shares its style collection with 3 other
// site(s) — needs its own collection") and sat `unresolved after 2 attempts`
// from 2026-08-06, because no handler could satisfy it: this action loud-failed
// on any site that already had a collection, and its own recommendation was for
// an operator to null `sites.style_collection_id` by hand.
//
// That manual route is what the flag replaces. Nulling the column leaves the
// site uncomposed until the re-resolve lands, and anything that renders in that
// window hits the composition loader's emergency fallback
// (render_css_composition_loader.go:144-158) and can deploy a `standard-brochure`
// stylesheet over a live site. With `allow_reinstall` the swap happens inside
// the action's existing transaction and the window never opens.
//
// THE TESTS COME IN A PAIR ON PURPOSE. A flag verified only on its permissive
// branch is satisfied by deleting the check — allowing everything would pass a
// permission-only test, and that is precisely the regression that matters here,
// because the unsafe direction is ON. So the default-OFF refusal is asserted
// too, and it is the more important of the two.
//
// Both tests stop at the guard rather than mocking the whole install. That is
// deliberate: the guard IS the change. Test B asserts only that the flag moved
// execution PAST the refusal, not that the install completed — claiming more
// than the mock can support is how a test becomes decoration.
//
// If you change this file, break the thing it guards and watch it fail. Both
// mutations were RUN on 2026-08-10 before this file was committed, and these are
// the results observed, not the results expected:
//   - `if !allowReinstall` → `if false`      → A and C FAIL, B passes
//   - `allowReinstall := false` (hardcoded)  → B FAILS, A and C pass
//
// The first mutation failing C as well as A is the useful part: C is what stops
// a malformed declaration switching the unsafe direction on, so it has to be
// sensitive to the same check.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

const (
	reinstallTestSiteID     = "2a8ebf9c-20a2-4c39-b191-840b012371da"
	reinstallTestExistingID = "3196d966-24ef-4415-9dc8-1afbc02166ca"
	reinstallTestPaletteID  = "11111111-1111-1111-1111-111111111111"
	reinstallTestLayoutID   = "22222222-2222-2222-2222-222222222222"
	reinstallTestTypoID     = "33333333-3333-3333-3333-333333333333"
)

// reinstallParams builds ActionParams for InstallSiteCompositionAction against a
// site that ALREADY has a style_collection_id. allowReinstall is passed through
// to the step config exactly as a workflow author would set it.
func reinstallParams(t *testing.T, allowReinstall interface{}) (ActionParams, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	// The site load is the only statement both branches reach. It reports an
	// EXISTING collection, which is the whole premise of these tests.
	mock.ExpectQuery("SELECT domain, style_collection_id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"domain", "style_collection_id"}).
			AddRow("ai-agent-orchestration.com", reinstallTestExistingID))

	// Data inputs live in collected_data — ExtractActionInputs treats config
	// STRINGS as path references, so putting ids in config would resolve to
	// nothing and every test would fail at extraction instead of at the guard.
	collected := map[string]interface{}{
		"site_id":                    reinstallTestSiteID,
		"selected_palette_id":        reinstallTestPaletteID,
		"selected_layout_id":         reinstallTestLayoutID,
		"selected_typography_set_id": reinstallTestTypoID,
	}
	// The switch is an author literal and lives in step config.
	// Absent means absent: the default-OFF case must not carry the key at all,
	// or it would test "explicit false" rather than "unset".
	cfg := map[string]interface{}{}
	if allowReinstall != nil {
		cfg["allow_reinstall"] = allowReinstall
	}

	return ActionParams{
			Context:       context.Background(),
			Logger:        zap.NewNop(),
			DB:            db,
			CollectedData: collected,
			StepConfig:    models.Step{Config: cfg},
			ExecutionContext: &orchtypes.ExecutionContext{
				OrchestrationID: "55555555-5555-5555-5555-555555555555",
				StepName:        "install_composition",
				Action:          "execute",
			},
		}, mock, func() {
			_ = db.Close()
		}
}

// A — the default. An existing collection is REFUSED when nothing asked for a
// reinstall. This is the test that keeps the unsafe direction off.
func TestInstallSiteComposition_RefusesExistingCollectionByDefault(t *testing.T) {
	params, _, done := reinstallParams(t, nil)
	defer done()

	_, err := InstallSiteCompositionAction(context.Background(), params)
	if err == nil {
		t.Fatalf("expected a refusal when the site already has a collection and " +
			"allow_reinstall was not set — got nil error, which means a live site's " +
			"composition would be replaced by a caller that never asked")
	}
	if !strings.Contains(err.Error(), "re-resolve not requested") {
		t.Fatalf("refused, but not for the reason under test: %v", err)
	}
	// The displaced id must be named, or an operator cannot roll back.
	if !strings.Contains(err.Error(), reinstallTestExistingID) {
		t.Errorf("refusal should name the existing collection id, got: %v", err)
	}
}

// B — the opt-in. The same site, the same existing collection, with the flag
// set, must get PAST the guard. It will fail later on the unmocked INSERT; what
// matters is that it fails somewhere else, for some other reason.
func TestInstallSiteComposition_AllowReinstallPassesTheGuard(t *testing.T) {
	params, _, done := reinstallParams(t, true)
	defer done()

	_, err := InstallSiteCompositionAction(context.Background(), params)
	if err != nil && strings.Contains(err.Error(), "re-resolve not requested") {
		t.Fatalf("allow_reinstall=true was ignored — still refused at the guard: %v", err)
	}
}

// C — a MALFORMED declaration must not switch the unsafe direction on. A string
// "true" is not a bool; GetBoolFieldLoud warns and falls back. For a flag whose
// permissive branch replaces a live site's stylesheet, "we could not parse it"
// has to mean "do not do it" — the opposite convention would let a typo in one
// workflow row repaint a site.
func TestInstallSiteComposition_MalformedFlagDoesNotEnableReinstall(t *testing.T) {
	params, _, done := reinstallParams(t, "true")
	defer done()

	_, err := InstallSiteCompositionAction(context.Background(), params)
	if err == nil || !strings.Contains(err.Error(), "re-resolve not requested") {
		t.Fatalf(`a non-bool allow_reinstall="true" enabled the reinstall path; `+
			`a malformed declaration must fall back to the SAFE branch, got: %v`, err)
	}
}
