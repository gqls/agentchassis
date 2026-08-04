// FILE: platform/storage/asset_paths_construction_test.go
//
// The tree-wide half of bugs_open/179 finding A: an AssetPaths may only be
// constructed inside this package.
//
// Finding A was one call site (deploy_image_asset_action.go) replacing the
// derived AssetPaths with a caller-supplied path. Deleting that block fixes the
// instance. This fixes the CLASS: a hand-built AssetPaths anywhere outside
// platform/storage is, by construction, a written path that the readers — which
// all resolve (asset_key, purpose) through DeployedAssetPath — cannot derive.
//
// Why a source scan rather than an API change: AssetPaths is a plain struct with
// exported fields because callers legitimately READ them (FilePath, RelativeURL,
// Filename). Unexporting the fields or adding a constructor-only guard would
// break every reader to constrain a writer. The property wanted is "nobody
// outside this package DECIDES these values", and that is structural.

package storage

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// assetPathsLiteral matches a composite literal of the type, spelled either way
// a caller can spell it: `AssetPaths{` inside this package, `storage.AssetPaths{`
// outside it. It deliberately does NOT match `[]AssetPaths{` or `*AssetPaths`
// declarations — a slice of results returned from here is fine; deciding the
// field values elsewhere is not.
var assetPathsLiteral = regexp.MustCompile(`(?:\bstorage\.)?\bAssetPaths\{`)

// commentPrefix spots a line that is entirely a comment. A mention in prose is
// not a construction, and this test must not fire on the doc comments that
// explain the rule.
var commentPrefix = regexp.MustCompile(`^\s*(//|\*|/\*)`)

func TestAssetPathsAreOnlyConstructedInStorage(t *testing.T) {
	// Walked from this package's directory, so the test is location-independent
	// if the package moves.
	roots := []string{"../../platform", "../../internal", "../../pkg"}

	type hit struct {
		file string
		line int
		text string
	}
	var hits []hit

	for _, root := range roots {
		if _, err := os.Stat(root); err != nil {
			// A repo layout change should be loud, not silently reduce coverage:
			// a root that has moved makes this sensor blind to a whole tree.
			t.Errorf("cannot walk %s: %v — repoint this test rather than letting it scan less", root, err)
			continue
		}
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "vendor" || d.Name() == "testdata" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			// This package owns the type and is where every construction belongs.
			if strings.Contains(filepath.ToSlash(path), "platform/storage/") {
				return nil
			}
			src, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for i, line := range strings.Split(string(src), "\n") {
				if commentPrefix.MatchString(line) {
					continue
				}
				if assetPathsLiteral.MatchString(line) {
					hits = append(hits, hit{file: path, line: i + 1, text: strings.TrimSpace(line)})
				}
			}
			return nil
		})
		if err != nil {
			t.Errorf("walking %s: %v", root, err)
		}
	}

	for _, h := range hits {
		t.Errorf("%s:%d constructs an AssetPaths outside platform/storage:\n    %s\n"+
			"A hand-built AssetPaths is the deploy_path override's shape — a path this writer chose "+
			"and no reader can derive (bugs_closed/168, bugs_open/179 finding A). Derive it with "+
			"storage.DeployedAssetPath(asset_key, purpose); if that cannot express what you need, "+
			"widen it HERE so every consumer moves with you.", h.file, h.line, h.text)
	}
}

// TestAssetPathsConstructionScanCanActuallyFail guards the guard.
//
// The test above passes when the tree is clean, and would also pass if its regexp
// were broken, its roots were wrong, or its comment filter swallowed everything —
// three ways for a green result to mean "looked at nothing". This asserts the
// matcher fires on the shape it exists to catch and stays quiet on the shapes it
// must not, so the zero above is a measurement rather than an artefact.
func TestAssetPathsConstructionScanCanActuallyFail(t *testing.T) {
	mustMatch := []string{
		`	processed.Paths = storage.AssetPaths{`,
		`p := AssetPaths{FilePath: "x"}`,
		`return storage.AssetPaths{FilePath: deployPath, RelativeURL: "/" + deployPath}`,
	}
	for _, s := range mustMatch {
		if !assetPathsLiteral.MatchString(s) {
			t.Errorf("the scan would MISS a real construction: %q", s)
		}
	}

	// Each of these must be quiet, and for the stated reason. No escape clauses:
	// an assertion with a `continue` for the awkward cases is how a check ends up
	// proving nothing.
	mustNotMatch := []struct {
		line, why string
	}{
		{`func DeployedAssetPath(assetKey, purpose string) AssetPaths {`, "a return type, not a construction"},
		{`var paths []AssetPaths`, "a slice declaration"},
		{`p.FilePath = derived.FilePath`, "reading fields is what callers are FOR"},
		{`	return derived`, "passing on this package's own result"},
	}
	for _, c := range mustNotMatch {
		if assetPathsLiteral.MatchString(c.line) {
			t.Errorf("the scan FIRES on something that is not a construction (%s): %q", c.why, c.line)
		}
	}

	// The comment filter is a separate mechanism from the regexp, so it needs its
	// own case: this line DOES match the regexp and must still be ignored.
	const inProse = `// a storage.AssetPaths{ in prose is documentation, not code`
	if !assetPathsLiteral.MatchString(inProse) {
		t.Fatalf("fixture no longer exercises the comment filter — it must match the regexp to be a test of it")
	}
	if !commentPrefix.MatchString(inProse) {
		t.Errorf("the comment filter would let prose count as a construction: %q", inProse)
	}
}
