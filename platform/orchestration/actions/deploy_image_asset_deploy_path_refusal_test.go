// FILE: platform/orchestration/actions/deploy_image_asset_deploy_path_refusal_test.go
//
// The behavioural half of bugs_open/179 finding A. The source scan in
// deploy_image_asset_path_source_test.go pins that the guard EXISTS and sits
// before anything irreversible; these run the action and pin what it RETURNS.
//
// DB, StorageClient and Producer are all nil on purpose, and that is itself the
// ordering proof — the same construction as TestDeriveBrandHeadBothLockedRefuses
// in derive_brand_head_assets_test.go. Reaching the storage resolution or the git
// commit with nil dependencies would error; a clean refusal therefore also
// demonstrates the guard runs before any write machinery.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// TestDeployImageAssetRefusesExplicitDeployPath covers all three EXPLICIT
// sources. Each is a separate arm in the action, so each gets its own case: a
// single case would pass while two arms were deleted.
func TestDeployImageAssetRefusesExplicitDeployPath(t *testing.T) {
	cases := []struct {
		name       string
		config     map[string]interface{}
		collected  map[string]interface{}
		wantSource string
	}{
		{
			name:       "step config deploy_path",
			config:     map[string]interface{}{"deploy_path": "assets/images/chosen-by-caller.jpg"},
			collected:  map[string]interface{}{"domain": "example.com"},
			wantSource: "config.deploy_path",
		},
		{
			name:       "deprecated deploy_path_field",
			config:     map[string]interface{}{"deploy_path_field": "assets/images/legacy-alias.jpg"},
			collected:  map[string]interface{}{"domain": "example.com"},
			wantSource: "config.deploy_path_field",
		},
		{
			name:   "input_data deploy_path",
			config: map[string]interface{}{},
			collected: map[string]interface{}{
				"domain":     "example.com",
				"input_data": map[string]interface{}{"deploy_path": "assets/images/from-input-data.jpg"},
			},
			wantSource: "input_data.deploy_path",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := DeployImageAssetAction(context.Background(), ActionParams{
				Logger:           zap.NewNop(),
				ExecutionContext: &types.ExecutionContext{},
				StepConfig:       models.Step{Config: c.config},
				CollectedData:    c.collected,
			})

			// A refusal is a RESULT, not an error — the work item must resolve
			// rather than retry for ever against a guard that will never pass it.
			if err != nil {
				t.Fatalf("refusal returned an error, not a result: %v\n"+
					"An erroring action retries; the item then churns against a guard that can never "+
					"let it through. House style is the success flag false plus a reason.", err)
			}
			res, ok := out.(map[string]interface{})
			if !ok {
				t.Fatalf("result is %T, want map[string]interface{}", out)
			}
			if res["deployed"] != false {
				t.Errorf("deployed = %v, want false — the file must NOT be committed", res["deployed"])
			}
			if res["skipped"] != true {
				t.Errorf("skipped = %v, want true", res["skipped"])
			}
			reason, _ := res["reason"].(string)
			if !strings.HasPrefix(reason, "refused: deploy_path") {
				t.Errorf("reason = %q, want the refusal prefix %q.\n"+
					"The reason string is what distinguishes this decline from the no-storage-URI one; "+
					"there is deliberately no bespoke key to test instead.", reason, "refused: deploy_path")
			}
			if !strings.Contains(reason, c.wantSource) {
				t.Errorf("reason does not name the source %q: %q.\n"+
					"Naming which of the three arms fired is what makes the refusal debuggable — "+
					"otherwise an operator cannot tell WHERE the path came from.", c.wantSource, reason)
			}
		})
	}
}

// TestDeployImageAssetIgnoresAggressivelyFoundDeployPath is the negative control,
// and the reason the refusal is wired to explicit sources only.
//
// ExtractActionInputs resolves each declared field by a depth-20 recursive search
// of the whole of collected_data (datahelpers/unified_extractor.go). `deploy_path`
// was a declared optional input on the live asset-deployer row, so before this
// change a deploy_path key buried in ANY nested step result silently redirected
// the git commit — the caller need never have asked for it.
//
// Refusing on that same search would swap a silent hijack for a false denial, so
// the action ignores it. This test pins that: the fixture carries a nested
// deploy_path and the live pre-change input_fields declaration, and the action
// must decline for the ordinary missing-image reason, NOT the deploy_path refusal.
func TestDeployImageAssetIgnoresAggressivelyFoundDeployPath(t *testing.T) {
	out, err := DeployImageAssetAction(context.Background(), ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig: models.Step{Config: map[string]interface{}{
			// The live asset-deployer shape as it stood on 2026-08-04, including
			// the declaration this change prunes. A config that still declares the
			// field must not become a config that cannot deploy.
			"input_fields": []interface{}{"s3_uri", "deploy_path", "purpose", "domain", "asset_key"},
		}},
		CollectedData: map[string]interface{}{
			"domain": "example.com",
			// Buried two levels down, exactly where a sub-agent response lands.
			"prior_step_result": map[string]interface{}{
				"deploy_path": "assets/images/sneaky.jpg",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("result is %T, want map[string]interface{}", out)
	}
	reason, _ := res["reason"].(string)
	if strings.HasPrefix(reason, "refused: deploy_path") {
		t.Errorf("a deploy_path found only by the extractor's recursive search caused a REFUSAL: %q\n"+
			"That is a false denial of a legitimate deploy — a stray key in any nested step result "+
			"would now block deploys fleet-wide. Refuse on explicit sources only.", reason)
	}
	// Positive half: it must still decline for the real reason (no image source),
	// so this fixture is exercising the action rather than passing vacuously.
	if res["deployed"] != false {
		t.Errorf("deployed = %v, want false — there is no storage URI in this fixture", res["deployed"])
	}
	if reason == "" {
		t.Errorf("no reason given; the fixture may no longer reach the decline it is meant to test")
	}
}
