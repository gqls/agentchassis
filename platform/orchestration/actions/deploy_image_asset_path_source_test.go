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
