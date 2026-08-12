// FILE: platform/orchestration/actions/deploy_image_asset_identity_test.go
//
// The behavioural half of bugs_open/248: this action must not invent an
// asset_key from its own config, and must not treat a spec DEFAULT as if a
// caller had chosen it.
//
// Why these are worth pinning behaviourally and not just by source scan. The
// defect shipped 150+ page-visible 404s across 16 sites while every mechanical
// signal was green: the work item reported `complete`, the git commit succeeded,
// `success: true` came back, and the deployed file was a valid image of exactly
// the right content. Only the NAME was wrong, and nothing on the writing side
// ever reads the name back. So the assertions here are on the two inputs that
// determine the name, observed through the action's own returned reason string.
//
// The observable surface is the no-storage-URI decline, which names the purpose
// it resolved. That is deliberate: it is reached AFTER identity resolution and
// BEFORE anything irreversible, so these tests need no storage client and no
// producer — and reaching a write path with those nil would error rather than
// return, which is itself part of the ordering proof.

package actions

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// TestDeployImageAssetTakesPurposeFromTheAssetRowWhenTheRunStatesNone pins
// finding (b). The dispatcher's input_mapping carries no `purpose` key, so
// `input_data.purpose` resolves to nothing and the "hero" default stands — and
// because defaults are applied before the recursive search, nothing downstream
// can beat it. A logo then deploys with hero geometry under a .jpg extension.
//
// The row is consulted instead. Asserted through the decline's purpose, which
// would read "hero" before this fix and "logo" after it.
func TestDeployImageAssetTakesPurposeFromTheAssetRowWhenTheRunStatesNone(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	const assetID = "b99c5355-4b3a-430c-9294-56482726be34"

	// Identity first: what IS this asset.
	mock.ExpectQuery(`SELECT purpose, asset_key FROM assets`).
		WillReturnRows(sqlmock.NewRows([]string{"purpose", "asset_key"}).AddRow("logo", "logo"))
	// Then the source lookup, which finds no stored object — so the action
	// declines here, before any write machinery, and tells us what it resolved.
	mock.ExpectQuery(`SELECT storage_path, url FROM assets`).
		WillReturnRows(sqlmock.NewRows([]string{"storage_path", "url"}).AddRow("", ""))

	out, err := DeployImageAssetAction(context.Background(), ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		DB:               db,
		// Exactly the shape build-dispatch-loop produces: the spec carries the
		// truth, the mapping does not pass purpose, and the action's config
		// points at a path that therefore resolves to nothing.
		StepConfig: models.Step{Config: map[string]interface{}{
			"purpose":      "input_data.purpose",
			"asset_id":     "input_data.spec.asset_id",
			"input_fields": []interface{}{"purpose", "domain", "asset_key", "asset_id"},
		}},
		CollectedData: map[string]interface{}{
			"domain": "gaswholesalers.com",
			"input_data": map[string]interface{}{
				"domain": "gaswholesalers.com",
				"spec": map[string]interface{}{
					"asset_id": assetID,
					"purpose":  "logo",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("action errored instead of declining: %v", err)
	}

	res, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("result is %T, want map[string]interface{}", out)
	}
	reason, _ := res["reason"].(string)
	if !strings.Contains(reason, "logo") {
		t.Errorf("resolved purpose is not the asset row's.\nreason = %q\n"+
			"Want it to name %q. If it names \"hero\", the spec default is still winning: the row is "+
			"the only source of an asset's purpose that a dispatcher's input_mapping cannot drop, and "+
			"without it every work-item-dispatched logo deploys with hero geometry and a .jpg "+
			"extension (bugs_open/248 finding (b)).", reason, "logo")
	}
	if strings.Contains(reason, "hero") {
		t.Errorf("reason still names hero: %q — the default beat the row", reason)
	}
}

// TestDeployImageAssetAppliesTheRowPurposeBeforeTheBrandHeadRefusal is the
// ordering half, and the one that could do real damage if it regressed.
//
// derive_brand_head_assets is the writer for favicon/og_card; this action
// refuses them so an arbitrary image cannot overwrite a site's real favicon or
// social card (bugs_open/179 finding B). That refusal reads `purpose` — so if the
// row's purpose were applied AFTER it, a favicon row dispatched with no stated
// purpose would sail through the guard as a "hero" and overwrite the live
// artefact. Recovering an input must not move a guard's input out from under it.
func TestDeployImageAssetAppliesTheRowPurposeBeforeTheBrandHeadRefusal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT purpose, asset_key FROM assets`).
		WillReturnRows(sqlmock.NewRows([]string{"purpose", "asset_key"}).AddRow("favicon", "favicon"))

	out, err := DeployImageAssetAction(context.Background(), ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		DB:               db,
		StepConfig: models.Step{Config: map[string]interface{}{
			"purpose":      "input_data.purpose",
			"asset_id":     "input_data.spec.asset_id",
			"input_fields": []interface{}{"purpose", "domain", "asset_key", "asset_id"},
		}},
		CollectedData: map[string]interface{}{
			"domain": "example.com",
			"input_data": map[string]interface{}{
				"domain": "example.com",
				"spec":   map[string]interface{}{"asset_id": "3f2504e0-4f89-11d3-9a0c-0305e82c3301"},
			},
		},
	})
	if err != nil {
		t.Fatalf("action errored instead of refusing: %v", err)
	}
	res, _ := out.(map[string]interface{})
	reason, _ := res["reason"].(string)
	if !strings.HasPrefix(reason, "refused: purpose") {
		t.Fatalf("a favicon asset was NOT refused.\nreason = %q\n"+
			"The brand-head refusal reads `purpose`, so the row's purpose must be resolved before it. "+
			"Resolving it later would let a favicon through as a hero and overwrite the artefact "+
			"derive_brand_head_assets owns — committed to git before any lock guard runs.", reason)
	}
	if res["deployed"] != false {
		t.Errorf("deployed = %v, want false", res["deployed"])
	}
}

// TestDeployImageAssetNeverUsesConfigAssetKeyAsALiteralFilename pins finding (a)
// structurally, because the value it guards against is not observable in any
// result: the placeholder name only shows up in the git commit, which needs a
// storage client and a producer.
//
// This is a source scan, and the reason is the same one the sibling scans in
// deploy_image_asset_path_source_test.go give: the property is STRUCTURAL. What
// failed was not a value but an arrangement — a config key that holds a
// reference on every live caller being readable as a value. The scan is written
// to be non-vacuous: it asserts the read is absent from CODE, and separately
// that the ladder's two legitimate rungs are still present, so deleting the
// whole resolution would fail it rather than satisfy it.
func TestDeployImageAssetNeverUsesConfigAssetKeyAsALiteralFilename(t *testing.T) {
	body := readDeployerSource(t)

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") {
			continue // prose may name it; that is how the removal is documented
		}
		if strings.Contains(line, `config["asset_key"]`) {
			t.Errorf("the literal asset_key rung is back:\n\t%s\n"+
				"Every live caller sets config.asset_key to a dotted PATH "+
				"(asset-deployer: \"input_data.asset_key\"), so reading it as a filename publishes "+
				"files named after the path expression — `input-data.asset-key.jpg`. 150+ rows across "+
				"16 sites, each a page-visible 404, because readers derive the real name. There is no "+
				"caller-side discipline that makes this safe (bugs_open/248 finding (a)).", trimmed)
		}
	}

	// Non-vacuity: the legitimate rungs must still exist, or this test would
	// pass on an action that resolves no asset_key at all — which would silently
	// collapse every per-variant purpose onto one filename.
	for _, want := range []string{`config["asset_key_field"]`, `assetRowIdentity(`} {
		if !strings.Contains(body, want) {
			t.Errorf("%s is gone from the deployer.\n"+
				"Removing a legitimate asset_key source would make this scan vacuous AND collapse "+
				"multi-variant purposes (hero_about.jpg and hero.jpg) onto one path.", want)
		}
	}
}

// TestDeployImageAssetDiscardsADottedAssetKey pins the class guard rather than
// the one instance: whatever produces it, a key containing a dot is an
// unresolved path expression. Measured 2026-08-12: 478 asset rows, none empty,
// none containing a dot — so discarding cannot lose a real key.
func TestDeployImageAssetDiscardsADottedAssetKey(t *testing.T) {
	body := readDeployerSource(t)
	if !strings.Contains(body, `strings.Contains(assetKey, ".")`) {
		t.Errorf("the dotted-asset_key guard is gone.\n" +
			"It is the check that survives a NEW config making the same mistake in a different key — " +
			"the deleted literal rung was one instance of the class, not the class.")
	}
}

// readDeployerSource is shared by the scans above; a missing file must fail
// loudly rather than let a scan pass by finding nothing.
func readDeployerSource(t *testing.T) string {
	t.Helper()
	const deployer = "deploy_image_asset_action.go"
	src, err := os.ReadFile(deployer)
	if err != nil {
		t.Fatalf("read %s: %v — if it moved, repoint these scans rather than deleting them",
			deployer, err)
	}
	return string(src)
}
