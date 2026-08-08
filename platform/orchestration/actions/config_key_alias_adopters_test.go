// FILE: platform/orchestration/actions/config_key_alias_adopters_test.go
//
// bugs_open/136 — the `domain` -> `pipeline` rename landed in Go and never in
// the data, and only one of the three actions carrying it wrote a back-compat
// fallback. These tests pin the alias declarations that replaced that fallback
// and extended it to the other two.
//
// Two levels, because they die on different mutations. The spec tests catch
// "someone deleted the alias entry"; the sqlmock test catches "someone reverted
// the action body to the direct config read while leaving the declaration in
// place" — which the spec tests cannot see, and which is exactly the shape this
// bug is about: a declaration that has stopped describing the code.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// TestBug136RenameAliasesResolve calls ResolveConfigSetting exactly as each
// adopting action does, with the config its live callers actually carry.
//
// MUTATION PROOF: delete any spec's DeprecatedConfigKeys entry and that row
// falls back to the default, failing here.
func TestBug136RenameAliasesResolve(t *testing.T) {
	cases := []struct {
		name       string
		spec       datahelpers.ActionInputSpec
		canonical  string
		defaultVal string
		config     map[string]interface{}
		want       string
		wantVia    string
	}{
		{
			// completeness-discovery-agent, live. Until 2026-08-08 this resolved
			// to "design" and filed four rows under the wrong pipeline.
			name:      "run_discovery_checks honours check_domain",
			spec:      RunDiscoveryChecksInputSpec,
			canonical: "check_pipeline", defaultVal: "design",
			config: map[string]interface{}{"check_domain": "content", "checks": []interface{}{}},
			want:   "content", wantVia: "check_domain",
		},
		{
			// design-discovery-agent, live — asks for the same value the default
			// already gave. Included deliberately: it is the case that made the
			// defect invisible, and a test suite that only covers the changing
			// row cannot show that the unchanged ones stayed unchanged.
			name:      "run_discovery_checks: design agent value is unchanged",
			spec:      RunDiscoveryChecksInputSpec,
			canonical: "check_pipeline", defaultVal: "design",
			config: map[string]interface{}{"check_domain": "design"},
			want:   "design", wantVia: "check_domain",
		},
		{
			name:      "triage_detected_items honours target_domain",
			spec:      TriageDetectedItemsInputSpec,
			canonical: "target_pipeline", defaultVal: "build",
			config: map[string]interface{}{"target_domain": "content"},
			want:   "content", wantVia: "target_domain",
		},
		{
			// The shim's own truth table, preserved through the conversion.
			name:      "create_work_item: new name wins over old",
			spec:      CreateWorkItemInputSpec,
			canonical: "item_pipeline", defaultVal: "build",
			config: map[string]interface{}{"item_pipeline": "content", "item_domain": "build"},
			want:   "content", wantVia: "",
		},
		{
			name:      "create_work_item: old name alone still works",
			spec:      CreateWorkItemInputSpec,
			canonical: "item_pipeline", defaultVal: "build",
			config: map[string]interface{}{"item_domain": "content"},
			want:   "content", wantVia: "item_domain",
		},
		{
			name:      "create_work_item: neither set falls back to build",
			spec:      CreateWorkItemInputSpec,
			canonical: "item_pipeline", defaultVal: "build",
			config: map[string]interface{}{},
			want:   "build", wantVia: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, via := datahelpers.ResolveConfigSetting(
				tc.config, tc.spec, tc.canonical, tc.defaultVal, zap.NewNop())
			if got != tc.want {
				t.Errorf("value = %q, want %q", got, tc.want)
			}
			if via != tc.wantVia {
				t.Errorf("viaDeprecatedKey = %q, want %q", via, tc.wantVia)
			}
		})
	}
}

// TestTriageDetectedItemsHonoursTargetDomain drives the real action with the
// old key set, and asserts the value reaches the UPDATE.
//
// This is the only test in this file that would catch the action body being
// reverted to `config["target_pipeline"].(string)` while the spec still
// declares the alias — a declaration agreeing with nothing, which is the
// original bug wearing a fix.
//
// MUTATION PROOF: restore the direct config read in the action and the promoted
// pipeline reverts to "build", so the WithArgs expectation goes unmet.
func TestTriageDetectedItemsHonoursTargetDomain(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	// "content", not the "build" default — so a passing result cannot be the
	// default in disguise.
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(siteID, "content").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectQuery(`SELECT count\(\*\)[\s\S]*pipeline = \$2`).
		WithArgs(siteID, "content").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	params := ActionParams{
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "triage_findings"},
		StepConfig: models.Step{Config: map[string]interface{}{
			"target_domain": "content", // the OLD name, as improvement-loop writes it
		}},
		CollectedData: map[string]interface{}{"site_id": siteID.String()},
		Logger:        zap.NewNop(),
		DB:            db,
	}

	if _, err := TriageDetectedItemsAction(context.Background(), params); err != nil {
		t.Fatalf("action returned an error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the old key did not reach the UPDATE: %v", err)
	}
}

// TestExecuteVisionPromptDeclaresAIService pins the false-positive fix: the
// action reads ai_service through resolveAIServiceConfig, so an audit that
// calls it unknown is wrong about working config.
//
// MUTATION PROOF: remove "ai_service" from the spec's ConfigKeys and this fails.
func TestExecuteVisionPromptDeclaresAIService(t *testing.T) {
	unknown, checked := datahelpers.UnknownConfigKeys("execute_vision_prompt",
		map[string]interface{}{"ai_service": map[string]interface{}{"provider": "anthropic"}})
	if !checked {
		t.Fatal("checked = false; execute_vision_prompt sets CheckConfig")
	}
	if len(unknown) != 0 {
		t.Errorf("unknown = %v; ai_service is read via resolveAIServiceConfig", unknown)
	}

	// Negative control — the check must still be capable of reporting.
	unknown, _ = datahelpers.UnknownConfigKeys("execute_vision_prompt",
		map[string]interface{}{"not_a_real_key": "x"})
	if len(unknown) != 1 || unknown[0] != "not_a_real_key" {
		t.Errorf("negative control: unknown = %v, want [not_a_real_key]", unknown)
	}
}
