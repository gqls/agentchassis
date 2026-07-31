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

// rawMarkerTest matches a marker test written as a bare string comparison OR as
// a compiled regexp — the shapes that carry no scope in their name.
//
// THE REGEXP ALTERNATIVE IS NOT DECORATION, and finding out it was missing is
// worth more than the gate itself. The first version of this test matched only
// Contains/HasPrefix/HasSuffix/Index, reported the tree clean, and I was one
// step from submitting a count of call sites derived from it — while
// rerender_single_page_action.go was testing the same marker through
// `regexp.MustCompile("(?i)data-runtime-fill")`. A gate that proves an absence
// only for the spellings it happens to search is this bug's own defect one level
// up. The manifest was therefore re-derived by grepping the LITERAL across the
// repo, not by trusting this pattern.
var rawMarkerTest = regexp.MustCompile(
	`(?i)((strings\.)?(Contains|HasPrefix|HasSuffix|Index)\s*\([^)]*"data-runtime-fill"` +
		"|regexp\\.MustCompile\\s*\\(\\s*[`\"][^`\"]*data-runtime-fill" +
		`|LIKE\s+'%data-runtime-fill%')`)

// thisFileOwnsTheMarker is where the raw literal is allowed to live.
const thisFileOwnsTheMarker = "runtime_fill.go"

// allowedRawMarkerSites are the sites deliberately left raw, each with the
// reason it cannot simply be renamed. An entry here is a decision on the record
// — which is the point. An allow-list with no escape hatch just pressures the
// next author into weakening the pattern instead.
// THE SQL ALTERNATIVE IN THE PATTERN ANSWERS THE OTHER HALF (council round 2,
// bug_historian): "the new lockstep gate only walks Go source — it cannot see
// [the SQL-side copies]. The mechanism remains generic and exploitable in the
// one place this round admits it did not look at closely enough to fix."
//
// Fair, and cheap to close, because every one of those copies is a SQL string
// EMBEDDED IN GO SOURCE — the same walk sees them once the pattern knows the SQL
// spelling. They are gated on the same terms as the Go ones: named here with a
// reason, or the build fails.
//
// AND THE QUESTION THAT SEAT ASKED, ANSWERED BY READING ALL FOUR rather than
// asserting it: every SQL copy computes the flag PER ROW — over a single
// pc.rendered_html or cc.html_template — and every one asks the SECTION question
// ("is this component/template a shell, so its emptiness is by design?"). None
// judges control liveness. They are the whole-input question at its correct
// granularity, not heuristic siblings of this bug.
var allowedRawMarkerSites = map[string]string{
	// --- Go, the section question -----------------------------------------
	// sectionHasVisibleContent asks the SECTION question ("keep this section in
	// the assembled page?"), so whole-input is already the right scope. It stays
	// a regexp rather than becoming HasRuntimeFillMarker because its (?i) makes
	// it the ONLY case-insensitive marker test in the tree: swapping it for the
	// case-sensitive predicate would be a silent behaviour change to the page
	// assembler, smuggled in under a scope fix. An unresolved divergence, not an
	// oversight (bugs_open/137, council round 2).
	"rerender_single_page_action.go": "section question; the tree's only (?i) test — see bugs_open/137",

	// emptySectionVerdict and its SQL twin both ask "is this section a shell, so
	// empty by design?", per component. Left untouched rather than renamed: the
	// rename changed no behaviour, and widening a diff to improve a comment costs
	// reviewers more than it gives them (council round 2).
	"check_empty_sections.go": "section question, per component (Go verdict + its SQL twin)",

	// --- Go, a WRITER ------------------------------------------------------
	// DropDeadURLControls REMOVES the control from rendered chrome, so it is a
	// WRITER: too narrow deletes chrome that was going to hydrate, too wide only
	// leaves a dead control visible — and visible is what emitChromeDeadControlItem
	// escalates. Same fail-safe direction as RepairPageLinks, and deliberately NOT
	// edited: chrome is shared across many sites, so the safest change is none
	// (council round 2, guardian).
	"render_site_components_action.go": "WRITER (drops the control); shared chrome — see link_repair.go's header",

	// --- SQL embedded in Go: all the section question, all per row ----------
	"check_required_fields_missing.go":      "SQL, per row: is this component a shell, so missing fields are by design?",
	"check_component_standards.go":          "SQL, per row: is this template a shell, so '<no value>' is the mechanism?",
	"check_component_template_corrupted.go": "SQL, per row: is this template a shell, so build-time emptiness is intended?",
}

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
			base := filepath.Base(path)
			if base == thisFileOwnsTheMarker || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			if _, allowed := allowedRawMarkerSites[base]; allowed {
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
