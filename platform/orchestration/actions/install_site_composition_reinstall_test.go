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
//
// ROUND 3 (2026-08-12) added F and reworked the spec read onto the platform's
// exact-path helper. Three more mutations RUN before commit, results observed:
//   - spec path → "input_data.NOPE"          → D FAILS alone
//   - `allowReinstall = false` after the
//     step-config read                       → B FAILS alone
//   - both silent-no-op strings collapsed
//     to "default(false)"                    → F FAILS alone (both sub-cases)
//
// Each mutation failing exactly ONE test is the point: it says the three
// channels — step config, work item spec, and the diagnostic that tells their
// absences apart — are independently load-bearing rather than one check wearing
// three names.
//
// F's THIRD sub-case exists because round 3's APPROVED verdict came with an
// advisory objection that was correct: the first draft returned a bare nil for
// both "nothing at that path" and "something at that path that is not an
// object", so the diagnostic reported "no input_data.spec in collected_data"
// for a spec that was demonstrably present. Mutation run after fixing it,
// result observed: collapsing the two reasons back into one fails
// F/spec_arrived_but_is_not_an_object ALONE. A diagnostic that lies about the
// rarer case is worse than no diagnostic, because it sends the next reader to
// the wrong half of the system with confidence.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
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
	return reinstallParamsFrom(t, allowReinstall, nil)
}

// reinstallParamsFrom is the same, but lets a test put the flag in the dispatching
// work item's spec (the PER-REQUEST channel) instead of the step config.
func reinstallParamsFrom(t *testing.T, stepCfgFlag interface{}, specFlag interface{}) (ActionParams, sqlmock.Sqlmock, func()) {
	t.Helper()
	allowReinstall := stepCfgFlag

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
	// The per-request channel: the dispatching work item's spec.
	if specFlag != nil {
		collected["input_data"] = map[string]interface{}{
			"spec": map[string]interface{}{"allow_reinstall": specFlag},
		}
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

// D — THE OBJECTION THE COUNCIL RAISED. Step config alone is an agent-definition
// edit, so it turns re-install on for EVERY composition install fleet-wide. The
// work item's spec is the only channel a SINGLE dispatch can set, and without it
// the flag cannot repair one site — which is the case it was built for.
func TestInstallSiteComposition_AllowReinstallFromWorkItemSpec(t *testing.T) {
	params, _, done := reinstallParamsFrom(t, nil, true)
	defer done()

	_, err := InstallSiteCompositionAction(context.Background(), params)
	if err != nil && strings.Contains(err.Error(), "re-resolve not requested") {
		t.Fatalf("allow_reinstall=true in the work item spec was ignored — the flag is "+
			"then only settable fleet-wide, which is the state it exists to prevent: %v", err)
	}
}

// E — the per-request channel must be exactly as strict as the step-config one.
// A malformed value here would be the easier mistake to make, because a work item
// spec is assembled by whatever queued it.
func TestInstallSiteComposition_MalformedSpecFlagDoesNotEnableReinstall(t *testing.T) {
	params, _, done := reinstallParamsFrom(t, nil, "yes")
	defer done()

	_, err := InstallSiteCompositionAction(context.Background(), params)
	if err == nil || !strings.Contains(err.Error(), "re-resolve not requested") {
		t.Fatalf(`a non-bool allow_reinstall="yes" in the work item spec enabled the `+
			`reinstall path; it must fall back to the SAFE branch: %v`, err)
	}
}

// F — the council's round-2 guardian objection, made ENFORCED rather than
// merely present: "ship this with a log line naming which branch resolved the
// flag so a silent no-op is diagnosable, not just safe."
//
// The two failure modes below both end in the same refusal and the same absent
// flag, and they have completely different fixes — "the spec never arrived"
// (dispatch/shape problem) versus "the spec arrived without the key" (whoever
// queued the item forgot it). GetBoolFieldLoud is deliberately silent on an
// ABSENT key, so without `resolved_from` nothing in the log tells them apart.
//
// Asserting the substrings is what stops the line rotting into decoration: a
// future edit that drops the distinction fails here rather than being noticed
// months later by someone debugging a flag that "does nothing".
func TestInstallSiteComposition_ResolvedFromDistinguishesTheSilentNoOps(t *testing.T) {
	for _, tc := range []struct {
		name       string
		spec       interface{} // nil = no input_data at all
		wantSubstr string
	}{
		{
			name:       "no spec reached the action",
			spec:       nil,
			wantSubstr: "no input_data.spec in collected_data",
		},
		{
			name:       "spec arrived but declares nothing",
			spec:       map[string]interface{}{"stage": "composition"},
			wantSubstr: "present but declares no true allow_reinstall",
		},
		{
			// Council b8e341b9 round 3, editquality (medium, and it was right):
			// an earlier draft returned a bare nil for BOTH "nothing at that
			// path" and "something there that is not an object", so this case
			// was reported as "no input_data.spec in collected_data" — a
			// diagnostic actively lying about a spec that was present. The type
			// must be named, because "who queued this item" is a different
			// investigation from "why did the spec not arrive".
			name:       "spec arrived but is not an object",
			spec:       "allow_reinstall=true",
			wantSubstr: "is present but is string, not an object",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params, _, done := reinstallParamsFrom(t, nil, nil)
			defer done()

			core, logs := observer.New(zap.InfoLevel)
			params.Logger = zap.New(core)
			if tc.spec != nil {
				params.CollectedData["input_data"] = map[string]interface{}{"spec": tc.spec}
			}

			_, err := InstallSiteCompositionAction(context.Background(), params)
			if err == nil || !strings.Contains(err.Error(), "re-resolve not requested") {
				t.Fatalf("premise broken — this case must still refuse: %v", err)
			}

			var got string
			for _, entry := range logs.FilterMessage(
				"InstallSiteCompositionAction: allow_reinstall resolved").All() {
				if v, ok := entry.ContextMap()["resolved_from"].(string); ok {
					got = v
				}
			}
			if got == "" {
				t.Fatalf("no allow_reinstall resolution was logged at all; a silent " +
					"no-op is then indistinguishable from a safe refusal")
			}
			if !strings.Contains(got, tc.wantSubstr) {
				t.Errorf("resolved_from does not identify this case:\n  got:  %q\n  want substring: %q",
					got, tc.wantSubstr)
			}
		})
	}
}

// ── approval (owner ruling 2026-08-12) ──────────────────────────────────────
//
// "yes approval needed but for now default that the human approves." So the
// replace path always RECORDS an approver, and today the default is a grant.
// These tests exercise the resolver directly rather than through the action,
// because the action's later statements are unmocked and the result map — where
// the approver is written — is never reached in this file's fixtures.
//
// G is the one that matters when the default is eventually tightened: it pins
// that an un-approved re-compose is distinguishable from an approved one, which
// is the only thing that makes the flip measurable before it is made.

func approverParams(t *testing.T, stepCfg map[string]interface{}, spec map[string]interface{}) ActionParams {
	t.Helper()
	collected := map[string]interface{}{}
	if spec != nil {
		collected["input_data"] = map[string]interface{}{"spec": spec}
	}
	if stepCfg == nil {
		stepCfg = map[string]interface{}{}
	}
	return ActionParams{
		Context:       context.Background(),
		Logger:        zap.NewNop(),
		CollectedData: collected,
		StepConfig:    models.Step{Config: stepCfg},
	}
}

// G — nobody named an approver, so the standing default is recorded AS the
// default, not as a person. If this ever returns "" or a human-looking name,
// the audit query that separates real approvals from inherited ones breaks.
func TestResolveReinstallApprover_DefaultsToTheStandingGrant(t *testing.T) {
	got := resolveReinstallApprover(
		approverParams(t, nil, nil), zap.NewNop(),
		uuid.MustParse(reinstallTestSiteID), "ai-agent-orchestration.com")
	if got != reinstallDefaultApprover {
		t.Fatalf("expected the standing default %q, got %q", reinstallDefaultApprover, got)
	}
}

// H — a named approver in the work item spec wins over the default. This is the
// channel a real approval queue will use.
func TestResolveReinstallApprover_NamedApproverWins(t *testing.T) {
	for _, key := range []string{"reinstall_approved_by", "approved_by"} {
		got := resolveReinstallApprover(
			approverParams(t, nil, map[string]interface{}{key: "uk@websy.uk"}),
			zap.NewNop(), uuid.MustParse(reinstallTestSiteID), "ai-agent-orchestration.com")
		if got != "uk@websy.uk" {
			t.Fatalf("spec key %q: expected the named approver, got %q", key, got)
		}
		if got == reinstallDefaultApprover {
			t.Fatalf("spec key %q: a named approval was recorded as the standing default", key)
		}
	}
}

// I — step config can name one too, and it outranks the spec. That is the
// "every install this agent does is pre-approved" case; it does not exist
// today, and pinning the precedence stops it being invented by accident.
func TestResolveReinstallApprover_StepConfigOutranksSpec(t *testing.T) {
	got := resolveReinstallApprover(
		approverParams(t,
			map[string]interface{}{"reinstall_approved_by": "step-config-approver"},
			map[string]interface{}{"approved_by": "spec-approver"}),
		zap.NewNop(), uuid.MustParse(reinstallTestSiteID), "ai-agent-orchestration.com")
	if got != "step-config-approver" {
		t.Fatalf("expected step config to outrank the spec, got %q", got)
	}
}
