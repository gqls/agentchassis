// FILE: platform/orchestration/datahelpers/runtime_fill_lockstep_test.go
//
// The gate the council's bug_historian seat asked for (bugs_open/137, round 1):
//
//	"No lint/grep gate is proposed to catch a future or existing raw
//	 strings.Contains(html,\"data-runtime-fill\") call site outside the ones
//	 enumerated here … a documentation fix for a code-shape problem."
//
// The objection is right, and it is the defect's own shape one level up: the
// reason eight copies of this line accumulated is that adding a ninth cost
// nothing and told nobody. A comment saying "use the named predicate" is the
// enforcement mechanism the architecture seat has already rejected once, in
// this very package's neighbour:
//
//	"A doc comment is not an enforcement mechanism — nothing stops a third
//	 future check type from being added confirm-style by someone who never
//	 reads that comment."
//
// So this reads the SOURCE TREE and fails the build for any raw marker test
// outside runtime_fill.go — the same source-lockstep technique as
// TestEveryStaticCheckTypeIsClassified. It does not judge which SCOPE a call
// site should use; that is a decision only the author can make. It forces the
// decision to be made through a NAMED predicate, so it is visible in review.

package datahelpers

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// rawMarkerTest matches a marker test written as a bare string comparison —
// the shape that has no scope in its name.
var rawMarkerTest = regexp.MustCompile(`(?i)(strings\.)?(Contains|HasPrefix|HasSuffix|Index)\s*\([^)]*"data-runtime-fill"`)

// thisFileOwnsTheMarker is where the raw literal is allowed to live.
const thisFileOwnsTheMarker = "runtime_fill.go"

func TestNoRawRuntimeFillMarkerTestOutsideThisPackagesPredicate(t *testing.T) {
	// Walk the two platform trees that hold every consumer. Repo-relative from
	// this package: ../../.. is the repo root.
	roots := []string{
		filepath.Clean("../.."),    // platform/
		filepath.Clean("../../.."), // repo root, to catch internal/ and pkg/
	}
	seen := map[string]bool{}
	var offenders []string

	for _, root := range roots {
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			// Skip vendored/third-party trees and this file's own owner.
			if strings.Contains(path, "/vendor/") || strings.Contains(path, "/node_modules/") {
				return nil
			}
			abs, _ := filepath.Abs(path)
			if seen[abs] {
				return nil
			}
			seen[abs] = true
			if filepath.Base(path) == thisFileOwnsTheMarker || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			src, rerr := os.ReadFile(path)
			if rerr != nil {
				return nil
			}
			for i, line := range strings.Split(string(src), "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "//") {
					continue // a comment quoting the old shape is not a call site
				}
				if rawMarkerTest.MatchString(line) {
					offenders = append(offenders, filepath.Clean(path)+":"+itoa(i+1)+"  "+trimmed)
				}
			}
			return nil
		})
	}

	if len(offenders) > 0 {
		t.Errorf(`a raw "data-runtime-fill" test was added outside %s.

Use a NAMED predicate, so the scope you intend is visible in review:
  HasRuntimeFillMarker(html)   — "is this SECTION a shell?"  (whole input)
  RuntimeFillSpans(html)       — "is this CONTROL alive?"    (per element, string callers)
  InRuntimeFillShell(sel)      — the same question for a goquery selection

Whichever you pick, say WHY in a comment beside it: bugs_open/137 is a whole bug
about this predicate's scope being decided by accident of what the caller passed.

offending line(s):
  %s`, thisFileOwnsTheMarker, strings.Join(offenders, "\n  "))
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
