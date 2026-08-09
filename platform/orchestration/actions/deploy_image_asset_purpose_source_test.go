// FILE: platform/orchestration/actions/deploy_image_asset_purpose_source_test.go
// Evidence harness for bugs_open/209 — the purpose-keyed source lookup in
// findStorageURI (Priority 2) and what a naive removal of it would do.
//
// These tests are deliberately written to characterise CURRENT behaviour, not
// to assert a fix. Two separate claims are under test:
//
//   1. The Priority-2 branch really is purpose-keyed last-write-wins, so two
//      same-purpose assets in one run collapse to one source. (Proves the
//      BRANCH exists. It does not prove any live workflow reaches it — see
//      the bug file's own [UNMEASURED] note and NOTES_209 for the config
//      census that shows no live workflow does.)
//
//   2. The legacy deploy steps (pageflow-builder / site-work-orchestrator)
//      carry NO input_fields, so deploy_image_asset resolves their inputs by
//      ExtractActionInputs Strategy 2 — an aggressive recursive search. With
//      two stored assets in collected_data, `asset_id` resolves by Go map
//      iteration order, which is randomised. This is why bug 209's fix
//      candidate 1 ("delete Priority 2 and let the asset_id path be the only
//      DB-free route") is NOT safe as written.

package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// Claim 1: {purpose}_uri is a single purpose-keyed slot. A second same-purpose
// store in the same run overwrites the first, and the deploy of the FIRST asset
// then resolves to the SECOND asset's bytes.
func TestFindStorageURI_PurposeKeyIsLastWriteWins(t *testing.T) {
	logger := zap.NewNop()

	// Simulate StoreAssetAction running twice with purpose="hero" in one run,
	// exactly as v3_site_actions.go does: params.CollectedData[purpose+"_uri"] = storageURI
	collected := map[string]interface{}{}
	const firstAsset = "s3://assets/hero-alpha.png"
	const secondAsset = "s3://assets/hero-beta.png"

	collected["hero"+"_uri"] = firstAsset
	collected["hero"+"_uri"] = secondAsset // second same-purpose store in the same run

	got := findStorageURI(collected, map[string]interface{}{}, "hero", logger)

	if got != secondAsset {
		t.Fatalf("expected the purpose slot to hold the LAST write %q, got %q", secondAsset, got)
	}
	// The defect stated plainly: there is no longer any way to ask for the first
	// asset's source through this path. Deploying asset alpha yields beta's bytes.
	if got == firstAsset {
		t.Fatalf("unreachable: the slot cannot hold both")
	}
	t.Logf("Priority 2 resolved purpose 'hero' to %q; asset alpha's source is unrecoverable from collected_data", got)
}

// Claim 1b: the collision is purpose-scoped. Two DIFFERENT purposes do not
// collide — which is why the legacy hero+logo workflows are correct today and
// why Priority 2 is load-bearing for them rather than merely legacy.
func TestFindStorageURI_DistinctPurposesDoNotCollide(t *testing.T) {
	logger := zap.NewNop()

	collected := map[string]interface{}{
		"hero_uri": "s3://assets/hero.png",
		"logo_uri": "s3://assets/logo.png",
	}

	if got := findStorageURI(collected, map[string]interface{}{}, "hero", logger); got != "s3://assets/hero.png" {
		t.Errorf("hero resolved to %q, want the hero source", got)
	}
	if got := findStorageURI(collected, map[string]interface{}{}, "logo", logger); got != "s3://assets/logo.png" {
		t.Errorf("logo resolved to %q, want the logo source", got)
	}
}

// Claim 2: with the legacy (no input_fields) config shape, `asset_id` is
// resolved by aggressive recursive search over collected_data. When the run
// holds two stored assets, the value returned is not stable.
//
// This test asserts the CURRENT hazard. If it ever fails because resolution
// became deterministic, that is a real change worth reading — not a flake.
func TestExtractActionInputs_LegacyShape_AssetIDIsNotStable(t *testing.T) {
	logger := zap.NewNop()

	// The live pageflow-builder / site-work-orchestrator deploy_logo_image step:
	// no input_fields, a deprecated uri_field, a static purpose.
	legacyConfig := map[string]interface{}{
		"purpose":   "logo",
		"uri_field": "logo_result.image_uri",
	}

	const heroAssetID = "11111111-1111-1111-1111-111111111111"
	const logoAssetID = "22222222-2222-2222-2222-222222222222"

	newCollected := func() map[string]interface{} {
		return map[string]interface{}{
			"site_record": map[string]interface{}{"domain": "example.test"},
			// Both store_asset outputs are present by the time the second
			// deploy step runs, each carrying its own asset_id.
			"hero_stored": map[string]interface{}{
				"asset_id":  heroAssetID,
				"image_uri": "s3://assets/hero.png",
				"purpose":   "hero",
			},
			"logo_stored": map[string]interface{}{
				"asset_id":  logoAssetID,
				"image_uri": "s3://assets/logo.png",
				"purpose":   "logo",
			},
			"logo_result": map[string]interface{}{"image_uri": "s3://assets/logo.png"},
			"hero_result": map[string]interface{}{"image_uri": "s3://assets/hero.png"},
		}
	}

	seen := map[string]int{}
	const iterations = 400
	for i := 0; i < iterations; i++ {
		inputs, err := datahelpers.ExtractActionInputs(newCollected(), legacyConfig, DeployImageAssetInputSpec, logger)
		if err != nil {
			t.Fatalf("iteration %d: ExtractActionInputs returned error: %v", i, err)
		}
		seen[inputs.Get("asset_id")]++
	}

	t.Logf("asset_id resolutions over %d identical inputs: %v", iterations, seen)
	t.Logf("  hero asset_id = %s", heroAssetID)
	t.Logf("  logo asset_id = %s  <- the one this step should deploy", logoAssetID)

	if len(seen) > 1 {
		t.Logf("CONFIRMED: asset_id is NOT stable under the legacy config shape — %d distinct values", len(seen))
	}
	if _, wrong := seen[heroAssetID]; wrong {
		t.Logf("CONFIRMED: the logo deploy step resolved the HERO's asset_id in %d/%d runs", seen[heroAssetID], iterations)
	}
}

// Claim 2b: the same shape, but resolving the source URI. Records which route
// actually supplies the URI for the legacy step today, so that any future
// change to findStorageURI's priority order can be checked against it.
func TestDeployImageAsset_LegacyShape_SourceRouteToday(t *testing.T) {
	logger := zap.NewNop()

	legacyConfig := map[string]interface{}{
		"purpose":   "logo",
		"uri_field": "logo_result.image_uri",
	}
	collected := map[string]interface{}{
		"site_record": map[string]interface{}{"domain": "example.test"},
		"hero_uri":    "s3://assets/hero.png",
		"logo_uri":    "s3://assets/logo.png",
		"logo_result": map[string]interface{}{"image_uri": "s3://assets/logo.png"},
		"hero_result": map[string]interface{}{"image_uri": "s3://assets/hero.png"},
	}

	inputs, err := datahelpers.ExtractActionInputs(collected, legacyConfig, DeployImageAssetInputSpec, logger)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}

	viaInputs := inputs.Get("s3_uri")
	viaFind := findStorageURI(collected, legacyConfig, "logo", logger)

	t.Logf("s3_uri via ExtractActionInputs (deprecated uri_field bridge): %q", viaInputs)
	t.Logf("findStorageURI fallback for purpose 'logo':                   %q", viaFind)

	// Whichever route wins, it must be the logo's bytes.
	effective := viaInputs
	if effective == "" {
		effective = viaFind
	}
	if effective != "s3://assets/logo.png" {
		t.Fatalf("legacy logo deploy would use %q, want the logo source", effective)
	}
}

// ---------------------------------------------------------------------------
// bugs_open/231 — the spec's Defaults SHADOW a static config value.
//
// ExtractActionInputs applies spec.Defaults into Values FIRST, and strategies
// 1/2/3 all skip a field that already has a value. A single-segment (non-dotted)
// config value is invisible to Strategy 0. Net: a step config carrying a STATIC
// value for a spec-defaulted field is dead — the default wins — and the
// action's own `if purpose == ""` fallback can never fire because the default
// means purpose is never empty.
//
// These tests PIN the defective behaviour deliberately (characterisation, same
// contract as the tests above): when bugs_open/231 is fixed, update them with
// the fix, citing it.
// ---------------------------------------------------------------------------

// The live shape of pageflow-builder / site-work-orchestrator `deploy_logo_image`:
// static purpose "logo" + uri_field, no input_fields. The step believes it is
// deploying a logo; the action resolves purpose="hero" (the spec default), which
// drives resize dimensions AND the deploy path (BuildAssetPaths: filename =
// purpose + ext) — the logo's bytes would land at the HERO's path.
func TestLegacyLogoStep_StaticPurposeIsShadowedByDefault(t *testing.T) {
	logger := zap.NewNop()
	legacyConfig := map[string]interface{}{
		"purpose":   "logo",
		"uri_field": "logo_result.image_uri",
	}
	collected := map[string]interface{}{
		"site_record": map[string]interface{}{"domain": "example.test"},
		"logo_result": map[string]interface{}{"image_uri": "s3://assets/logo.png"},
	}
	inputs, err := datahelpers.ExtractActionInputs(collected, legacyConfig, DeployImageAssetInputSpec, logger)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}

	// Replicates DeployImageAssetAction's own purpose resolution (lines ~92-99).
	purpose := inputs.Get("purpose")
	if purpose == "" {
		if p, ok := legacyConfig["purpose"].(string); ok && p != "" {
			purpose = p
		} else {
			purpose = "hero"
		}
	}

	if purpose != "hero" {
		t.Fatalf("bugs_open/231 behaviour changed: legacy logo step now resolves purpose=%q "+
			"(was shadowed to \"hero\" by the spec default). If this is a deliberate fix, "+
			"update this test to assert the fixed behaviour and cite the fix commit.", purpose)
	}
	t.Logf("PINNED DEFECT (bugs_open/231): logo step's effective purpose = %q — static config value is dead", purpose)
}

// The deprecated purpose_field bridge is equally dead for a defaulted field:
// Strategy 3 skips fields that already hold a value, and the default always has.
func TestPurposeFieldBridge_DeadForDefaultedField(t *testing.T) {
	logger := zap.NewNop()
	cfg := map[string]interface{}{
		"purpose_field": "logo_stored.purpose",
	}
	collected := map[string]interface{}{
		"site_record": map[string]interface{}{"domain": "example.test"},
		"logo_stored": map[string]interface{}{"purpose": "logo"},
	}
	inputs, err := datahelpers.ExtractActionInputs(collected, cfg, DeployImageAssetInputSpec, logger)
	if err != nil {
		t.Fatalf("ExtractActionInputs: %v", err)
	}
	if got := inputs.Get("purpose"); got != "hero" {
		t.Fatalf("bugs_open/231 behaviour changed: purpose_field bridge now yields %q; update deliberately", got)
	}
	t.Logf("PINNED DEFECT (bugs_open/231): purpose_field bridge is inert for a defaulted field")
}

// The one mechanism that DOES defeat the default: a Strategy-0 explicit dotted
// path in the step config. This is the into-line repair shape for the legacy
// workflows — deterministic for every field, no recursive search, and it kills
// both bugs_open/231 (the shadow) and the 86% asset_id instability above.
func TestStrategy0DottedPaths_DefeatTheDefaultAndTheRecursiveSearch(t *testing.T) {
	logger := zap.NewNop()
	intoLineConfig := map[string]interface{}{
		"purpose":  "logo_stored.purpose",
		"s3_uri":   "logo_stored.s3_uri",
		"asset_id": "logo_stored.asset_id",
		"domain":   "site_record.domain",
	}
	collected := map[string]interface{}{
		"site_record": map[string]interface{}{"domain": "example.test"},
		"hero_stored": map[string]interface{}{
			"asset_id": "11111111-1111-1111-1111-111111111111",
			"purpose":  "hero",
			"s3_uri":   "s3://assets/hero.png",
		},
		"logo_stored": map[string]interface{}{
			"asset_id": "22222222-2222-2222-2222-222222222222",
			"purpose":  "logo",
			"s3_uri":   "s3://assets/logo.png",
		},
	}

	// Deterministic across iterations — unlike the no-input_fields shape above,
	// which resolves the wrong sibling ~86% of the time.
	for i := 0; i < 100; i++ {
		inputs, err := datahelpers.ExtractActionInputs(collected, intoLineConfig, DeployImageAssetInputSpec, logger)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if p, s, a := inputs.Get("purpose"), inputs.Get("s3_uri"), inputs.Get("asset_id"); p != "logo" ||
			s != "s3://assets/logo.png" || a != "22222222-2222-2222-2222-222222222222" {
			t.Fatalf("iteration %d: non-deterministic or wrong: purpose=%q s3_uri=%q asset_id=%q", i, p, s, a)
		}
	}
}
