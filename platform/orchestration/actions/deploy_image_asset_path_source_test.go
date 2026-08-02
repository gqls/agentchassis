// FILE: platform/orchestration/actions/deploy_image_asset_path_source_test.go
//
// The drift sensor for bugs_open/168: deploy_image_asset must RESOLVE THROUGH
// storage.DeployedAssetPath, not re-implement it.

package actions

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/storage"
)

// TestDeployImageAssetResolvesThroughTheSharedDerivation is a SOURCE scan, and
// deliberately so.
//
// The value-level assertion a reader reaches for first — "the path the deployer
// derives equals the path a renderer resolves" — is now a tautology: both call
// one function, so it passes even if that function is wrong, and it would keep
// passing if someone re-introduced a private copy of the derivation guarded by a
// condition the test does not exercise. It would be a decoration.
//
// What actually failed before 2026-08-02 was not a value but a STRUCTURE: two
// implementations of one convention, kept in step by a doc comment claiming to
// "mirror" the other. They did agree; the defect was that nothing made them.
// So the property worth pinning is structural, and the only way to assert it is
// to read the source.
//
// If this fails because the file moved or the call was legitimately refactored,
// repoint it — do not delete it. Deleting it restores the exact arrangement that
// nearly shipped a fleet-wide false 404 (see check_image_url_404.go's audit).
func TestDeployImageAssetResolvesThroughTheSharedDerivation(t *testing.T) {
	const deployer = "deploy_image_asset_action.go"
	src, err := os.ReadFile(deployer)
	if err != nil {
		t.Fatalf("read %s: %v — if it moved, repoint this test rather than deleting it", deployer, err)
	}
	body := string(src)

	if !strings.Contains(body, "storage.DeployedAssetPath(") {
		t.Errorf("%s no longer calls storage.DeployedAssetPath.\n"+
			"It is the one derivation of where a generated asset is committed and served from; a "+
			"deployer that computes its own path can write a file no renderer references, and both "+
			"halves will look correct in isolation (bugs_open/168).", deployer)
	}

	// Re-implementing the derivation means spelling its parts. These belong
	// inside platform/storage, never at a call site.
	for _, banned := range []string{
		"storage.AssetKeyFilename(",
		"storage.BuildAssetPaths(",
	} {
		if strings.Contains(body, banned) {
			t.Errorf("%s calls %s directly.\n"+
				"That is the derivation being re-implemented at the call site — the shape bugs_open/168 "+
				"removed. Use storage.DeployedAssetPath, and if it cannot express what you need, widen "+
				"it there so every consumer moves with you.", deployer, banned)
		}
	}

	// The sensor must not pass because the strings it looks for are in a
	// comment. Assert the call appears in code, not prose.
	commentOnly := regexp.MustCompile(`(?m)^\s*(//|\*)`)
	var sawCall bool
	for _, line := range strings.Split(body, "\n") {
		if strings.Contains(line, "storage.DeployedAssetPath(") && !commentOnly.MatchString(line) {
			sawCall = true
			break
		}
	}
	if !sawCall {
		t.Errorf("%s mentions storage.DeployedAssetPath only in comments — this sensor would be "+
			"vacuous, which is how a guard survives the removal of the thing it guards", deployer)
	}
}

// TestDeployImageAssetRefusesBrandHeadPurposes pins bugs_open/179 finding B.
//
// This action must not be the writer for favicon/og_card. Once it resolves paths
// through storage.DeployedAssetPath, a brand-head purpose lands on the SAME path
// derive_brand_head_assets publishes — turning what used to be an inert orphan
// into an overwrite of the live artefact, committed to git before any provenance
// or lock guard runs.
//
// A source scan again, for the same reason as the sensor above and one more: the
// refusal must sit BEFORE the storage-URI resolution and the download, because a
// guard that fires after the git commit is not a guard (LANDMINES: "Guarding an
// asset's provenance UPSERT is not guarding the asset — the git commit already
// ran"). Position is the property, and position is structural.
func TestDeployImageAssetRefusesBrandHeadPurposes(t *testing.T) {
	const deployer = "deploy_image_asset_action.go"
	src, err := os.ReadFile(deployer)
	if err != nil {
		t.Fatalf("read %s: %v — if it moved, repoint this test rather than deleting it", deployer, err)
	}
	body := string(src)

	guard := strings.Index(body, "storage.IsBrandHeadPurpose(purpose)")
	if guard < 0 {
		t.Fatalf("%s no longer refuses brand-head purposes.\n"+
			"11 open undeployed_asset items carried purpose favicon/og_card with no mode when this "+
			"was written (two at status 'detected'), and asset-deployer's check_mode only diverts "+
			"mode=brand_head — so they reach this action and would overwrite the site's real "+
			"favicon or social card (bugs_open/179).", deployer)
	}

	// The guard must precede every irreversible step. sendGitCommitRequest is
	// the one that actually replaces the file in the site repo.
	for _, after := range []string{
		"DownloadOptimizeAndPrepare",
		"sendGitCommitRequest",
		"storage.DeployedAssetPath(",
	} {
		at := strings.Index(body, after)
		if at < 0 {
			t.Errorf("cannot find %q in %s — repoint this ordering assertion rather than dropping it",
				after, deployer)
			continue
		}
		if guard > at {
			t.Errorf("the brand-head refusal appears AFTER %q.\n"+
				"A guard that fires after the download or the git commit is not a guard — the "+
				"artefact is already replaced by then.", after)
		}
	}

	// A refusal is a RESULT, not an error: the work item must resolve rather
	// than retry for ever. The shape is the sibling's (ingest_staged_asset) and
	// this action's own no-storage-URI decline — the success flag false plus a
	// `reason` — NOT a bespoke key. An earlier draft asserted `"refused": true`
	// here; the council's reuse seat pointed out that key existed nowhere else
	// in the platform, so both the code and this assertion now pin the
	// convention instead of a local invention.
	tail := body[guard:]
	if end := strings.Index(tail, "\n\t}\n"); end > 0 {
		block := tail[:end]
		for _, want := range []string{`"deployed": false`, `"reason"`, "nil"} {
			if !strings.Contains(block, want) {
				t.Errorf("the brand-head branch does not return a house-style refusal: %s missing.\n"+
					"It must complete with the success flag false and a reason, not error — or the "+
					"item retries against a guard that will never let it through.", want)
			}
		}
		if strings.Contains(block, `"refused"`) {
			t.Errorf("the brand-head branch returns a bespoke \"refused\" key. The house style is the " +
				"action's own success flag set false plus a reason (ingest_staged_asset, same agent); " +
				"a second vocabulary for the same idea is what the reuse seat objected to.")
		}
	}
}

// TestBrandHeadAssetPathsAreAbsolute closes the round-3 advisory from the
// bug_historian seat: DeployedAssetPath falls through SILENTLY to the generic
// purpose/asset_key derivation when a BrandHeadAssetPaths value does not start
// with "/", so an author who adds a relative entry by mistake gets no signal and
// the artefact quietly resolves to the wrong place.
//
// A runtime log would be the wrong remedy — the map is a compile-time
// declaration, so the malformed case can be made impossible to ship rather than
// merely noisy when it runs. This test is that guard: it fails the build, which
// is the only signal that arrives before the mistake reaches a site.
func TestBrandHeadAssetPathsAreAbsolute(t *testing.T) {
	for purpose, published := range storage.BrandHeadAssetPaths {
		if !strings.HasPrefix(published, "/") {
			t.Errorf("BrandHeadAssetPaths[%q] = %q, which is not an absolute site path.\n"+
				"storage.DeployedAssetPath refuses to split such a value and falls through to the "+
				"generic purpose-derived path SILENTLY, so this artefact would resolve to "+
				"/assets/images/%s.<ext> and nothing would say so.", purpose, published, purpose)
		}
		if strings.HasSuffix(published, "/") || len(published) < 2 {
			t.Errorf("BrandHeadAssetPaths[%q] = %q has no filename component", purpose, published)
		}
	}
}
