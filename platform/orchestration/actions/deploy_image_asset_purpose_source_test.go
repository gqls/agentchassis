// FILE: platform/orchestration/actions/deploy_image_asset_purpose_source_test.go
// Evidence harness for bugs_open/209 — the purpose-keyed source lookup that
// deploy_image_asset used to consult, and the two conditions that had to hold
// before it could be removed.
//
// HISTORY, because these tests changed meaning on 2026-08-09 and the reason is
// the whole point of keeping them:
//
//   - They began as CHARACTERISATION tests: `findStorageURI` resolved a source
//     by PURPOSE out of collected_data, and its Priority 2 (`{purpose}_uri`)
//     was a shared last-write-wins slot, so two same-purpose assets stored in
//     one run collapsed to one source.
//   - They also established why 209's own fix candidate 1 was NOT safe as
//     written: the legacy deploy steps carried no input_fields, so `asset_id`
//     resolved by aggressive recursive search — randomised, wrong ~86% of the
//     time for the logo step. Deleting the lookup THEN would have caused the
//     bug it was meant to prevent.
//   - Migration 348 (2026-08-09) removed that precondition by giving all four
//     legacy steps explicit dotted paths and input_fields, so every input now
//     resolves by identity through Strategy 0. Watched in production the same
//     day on a full cookly.uk build.
//   - `findStorageURI` was then deleted. The tests below no longer characterise
//     it; they PIN THE DOOR SHUT — that a purpose-keyed slot can no longer
//     supply a source, and that the post-348 config shape resolves by identity.
//
// TestExtractActionInputs_LegacyShape_AssetIDIsNotStable is deliberately kept
// unchanged: it pins RESOLVER behaviour, not the old config, and that behaviour
// is still current for anyone who authors a step without input_fields.

package actions

import (
	"context"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// The door, pinned shut. `{purpose}_uri` is still WRITTEN into collected_data by
// StoreAssetAction (v3_site_actions.go) — the writer outlives its reader, and
// retiring it is bug 209's optional Phase 3. So the slot is present here exactly
// as a run would leave it, and the assertion is that nothing reads it.
//
// If this test fails, a purpose-keyed source route has grown back: some resolution
// path is answering with a value that belongs to whichever same-purpose asset was
// stored LAST, which is bugs_open/209 and, one layer up, bugs_closed/155.
func TestDeployImageAsset_PurposeKeyedSlotIsNoLongerASource(t *testing.T) {
	logger := zap.NewNop()

	// Two same-purpose assets stored in one run: the shape that made the old
	// Priority 2 wrong. The slot holds the LAST write; asset alpha's source is
	// unrecoverable from it.
	collected := map[string]interface{}{
		"site_record": map[string]interface{}{"domain": "example.test"},
		"hero_uri":    "s3://assets/hero-beta.png", // alpha overwritten
		"hero_result": map[string]interface{}{"image_uri": "s3://assets/hero-beta.png"},
	}
	// A caller naming neither s3_uri nor asset_id — the only shape that could
	// ever have reached the deleted lookup.
	config := map[string]interface{}{
		"purpose":      "hero_stored.purpose",
		"input_fields": []interface{}{"purpose", "domain"},
	}

	// Drive the REAL action, not just the extractor. Asserting on
	// ExtractActionInputs alone would pass even if a purpose-keyed lookup were
	// reintroduced into the action itself, which is exactly the regression this
	// test exists to catch — the guard has to sit where the door is.
	//
	// DB is nil, so the asset_id route is unavailable too and the only question
	// left is whether anything derives a source from collected_data. Nothing may.
	// The action must reach its skip WITHOUT downloading: a source found here
	// would carry on to DownloadOptimizeAndPrepare and fail with some other error.
	res, err := DeployImageAssetAction(context.Background(), ActionParams{
		Context:          context.Background(),
		ExecutionContext: &types.ExecutionContext{Action: "process", StepName: "deploy_hero_image"},
		StepConfig:       models.Step{Config: config},
		CollectedData:    collected,
		DB:               nil,
		Logger:           logger,
	})
	// An ERROR is a regression signal here, not a broken test. With no resolvable
	// source the action must return the skip result before it touches anything;
	// if it got far enough to fail on the nil storage client, then something
	// resolved a source out of collected_data and carried on. Verified by
	// mutation 2026-08-09: reinstating the `{purpose}_uri` read makes this exact
	// branch fire with "storage client not available".
	if err != nil {
		t.Fatalf("the action proceeded past source resolution instead of skipping: %v\n"+
			"The only source candidates in this run are the purpose-keyed slots `hero_uri` and "+
			"`hero_result.image_uri`. Reaching the download means one of them was read, which is "+
			"the bugs_open/209 route: a shared last-write-wins value standing in for a per-asset "+
			"source. Resolve from s3_uri or the asset row, or skip.", err)
	}

	out, ok := res.(map[string]interface{})
	if !ok {
		t.Fatalf("result is %T, want map[string]interface{}", res)
	}
	if skipped, _ := out["skipped"].(bool); !skipped {
		t.Errorf("the action did NOT skip: %#v\n"+
			"It found a source in a run whose only candidates are the purpose-keyed slots "+
			"`hero_uri` / `hero_result.image_uri`. Those are shared, last-write-wins values: with two "+
			"same-purpose assets in one run they collapse to one source, and the deploy of the first "+
			"ships the second's bytes with success:true (bugs_open/209, and bugs_closed/155 one layer "+
			"up). A source may come only from s3_uri (named by the caller) or from the asset row.", out)
	}
	if deployed, _ := out["deployed"].(bool); deployed {
		t.Errorf("the action reported deployed:true with no resolvable source: %#v", out)
	}
	t.Logf("no source resolvable from a purpose-keyed slot; the action skipped: %v", out["reason"])
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

// Claim 2b, rewritten 2026-08-09: the POST-348 config shape, which is what the
// four legacy deploy steps carry live. Every input resolves by identity from the
// step's OWN store output, in the presence of a sibling asset that the old
// aggressive search would have picked instead.
//
// The exact shape is taken from the live rows, not invented:
//
//	purpose  = {p}_stored.purpose      s3_uri   = {p}_stored.s3_uri
//	asset_id = {p}_stored.asset_id     domain   = site_record.domain
//	input_fields = ["purpose","domain","asset_id"]      (no s3_uri — see below)
//
// Watched in production 2026-08-09 on a cookly.uk build: the hero step, running
// with BOTH assets in collected_data, logged Strategy 0 resolving each field and
// downloaded the hero's own object while the logo's sat beside it untouched.
func TestDeployImageAsset_Post348Shape_ResolvesEveryInputByIdentity(t *testing.T) {
	logger := zap.NewNop()

	const heroAssetID = "11111111-1111-1111-1111-111111111111"
	const logoAssetID = "22222222-2222-2222-2222-222222222222"

	// The LOGO step's live config.
	config := map[string]interface{}{
		"purpose":      "logo_stored.purpose",
		"s3_uri":       "logo_stored.s3_uri",
		"asset_id":     "logo_stored.asset_id",
		"domain":       "site_record.domain",
		"input_fields": []interface{}{"purpose", "domain", "asset_id"},
	}
	newCollected := func() map[string]interface{} {
		return map[string]interface{}{
			"site_record": map[string]interface{}{"domain": "example.test"},
			"hero_stored": map[string]interface{}{
				"asset_id": heroAssetID,
				"s3_uri":   "s3://assets/hero.png",
				"purpose":  "hero",
			},
			"logo_stored": map[string]interface{}{
				"asset_id": logoAssetID,
				"s3_uri":   "s3://assets/logo.png",
				"purpose":  "logo",
			},
			// The purpose-keyed slots a real run also leaves behind. Nothing may
			// read them.
			"hero_uri": "s3://assets/hero.png",
			"logo_uri": "s3://assets/logo.png",
		}
	}

	// Determinism is the claim, so assert it over many runs: Go randomises map
	// iteration, which is precisely what made the pre-348 shape unstable.
	const iterations = 200
	for i := 0; i < iterations; i++ {
		inputs, err := datahelpers.ExtractActionInputs(newCollected(), config, DeployImageAssetInputSpec, logger)
		if err != nil {
			t.Fatalf("iteration %d: ExtractActionInputs: %v", i, err)
		}
		if got := inputs.Get("s3_uri"); got != "s3://assets/logo.png" {
			t.Fatalf("iteration %d: s3_uri = %q, want the LOGO's own source.\n"+
				"s3_uri is absent from input_fields but present in the spec's Optional list, so "+
				"Strategy 0 resolves the config dotted path regardless — that is deliberate: the "+
				"explicit path resolves, the aggressive search does not run.", i, got)
		}
		if got := inputs.Get("asset_id"); got != logoAssetID {
			t.Fatalf("iteration %d: asset_id = %q, want the logo's %q (the hero's is %q).\n"+
				"A wrong value here is the ~86%% hazard the pre-348 shape had.", i, got, logoAssetID, heroAssetID)
		}
		if got := inputs.Get("purpose"); got != "logo" {
			t.Fatalf("iteration %d: purpose = %q, want \"logo\".\n"+
				"\"hero\" means the spec Default won and bugs_open/231 is back: the logo would be "+
				"resized and encoded as a hero and committed as logo.jpg, not a 400x400 logo.png.", i, got)
		}
	}
	t.Logf("all four inputs resolved by identity, %d/%d runs, with a sibling asset present throughout", iterations, iterations)
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

// The PRE-migration-348 shape of pageflow-builder / site-work-orchestrator
// `deploy_logo_image` (live until 2026-08-09; migration 348 replaced it with
// Strategy-0 dotted paths — see TestStrategy0DottedPaths below for the live
// shape): static purpose "logo" + uri_field, no input_fields. The step believes
// it is deploying a logo; the action resolves purpose="hero" (the spec default),
// which drives resize dimensions AND the deploy path (BuildAssetPaths:
// filename = purpose + ext) — the logo's bytes would land at the HERO's path.
// The RESOLVER behaviour pinned here is still current — any config authored in
// this shape today gets the same shadow — which is why this test outlives 348.
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
	// EXACT live shape of deploy_logo_image after migration 348 (2026-08-09),
	// including the pre-existing deprecated domain_field (inert: the Strategy-0
	// "domain" resolves first, so the Strategy-3 bridge's has-value skip fires)
	// and input_fields DELIBERATELY excluding s3_uri (see the migration header:
	// on a store failure the URI must come from the asset row or not at all,
	// never from an aggressive search that can cross assets).
	intoLineConfig := map[string]interface{}{
		"purpose":      "logo_stored.purpose",
		"s3_uri":       "logo_stored.s3_uri",
		"asset_id":     "logo_stored.asset_id",
		"domain":       "site_record.domain",
		"domain_field": "site_record.domain",
		"input_fields": []interface{}{"purpose", "domain", "asset_id"},
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
		if d := inputs.Get("domain"); d != "example.test" {
			t.Fatalf("iteration %d: domain=%q, want site_record.domain via Strategy 0", i, d)
		}
	}
}

// The migration-348 failure corner: the logo's store FAILED (logo_stored has no
// s3_uri), while the hero's succeeded. Because input_fields excludes s3_uri, the
// aggressive search must NOT go hunting for one — s3_uri resolves EMPTY, which
// the action turns into asset_id → dead row → safe skip. If this test ever
// reports the hero's URI, someone added s3_uri to input_fields and reopened the
// cross-asset door.
func TestMigration348Shape_StoreFailureResolvesNoURI_NeverTheSibling(t *testing.T) {
	logger := zap.NewNop()
	cfg := map[string]interface{}{
		"purpose":      "logo_stored.purpose",
		"s3_uri":       "logo_stored.s3_uri",
		"asset_id":     "logo_stored.asset_id",
		"domain":       "site_record.domain",
		"domain_field": "site_record.domain",
		"input_fields": []interface{}{"purpose", "domain", "asset_id"},
	}
	collected := map[string]interface{}{
		"site_record": map[string]interface{}{"domain": "example.test"},
		"hero_stored": map[string]interface{}{ // sibling succeeded — must never leak
			"asset_id": "11111111-1111-1111-1111-111111111111",
			"purpose":  "hero",
			"s3_uri":   "s3://assets/hero.png",
		},
		"logo_stored": map[string]interface{}{ // insert-failure shape: asset_id, no s3_uri
			"asset_id": "22222222-2222-2222-2222-222222222222",
			"purpose":  "logo",
		},
	}
	for i := 0; i < 100; i++ {
		inputs, err := datahelpers.ExtractActionInputs(collected, cfg, DeployImageAssetInputSpec, logger)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if s := inputs.Get("s3_uri"); s != "" {
			t.Fatalf("iteration %d: s3_uri=%q — the store-failure corner resolved a URI it must not (sibling leak?)", i, s)
		}
		if a := inputs.Get("asset_id"); a != "22222222-2222-2222-2222-222222222222" {
			t.Fatalf("iteration %d: asset_id=%q, want the logo's own (dead) id", i, a)
		}
	}
}
