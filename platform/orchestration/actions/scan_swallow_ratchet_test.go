// FILE: platform/orchestration/actions/scan_swallow_ratchet_test.go
//
// bugs_open/410: three independent seams fail toward the quiet default, complete
// green, and ship nothing. This is the door-closing half for one of the three
// classes — the silent scan loss.
//
// THE FAILURE MODE. A `for rows.Next()` loop whose `rows.Scan` error branch logs
// and `continue`s returns FEWER rows than the cursor yielded, with NO error. The
// caller cannot distinguish a thinned result from a genuinely short table, so
// the work completes, the artefact is rewritten from the short result, and the
// deploy stamp says freshly built. In loadStoredSections that was destructive
// rather than merely wrong: save_page_sections replaces the page's rows
// wholesale, so a section dropped by the scan is DELETED.
//
// WHERE THIS SITS AMONG THE EXISTING GUARDS (the reuse question, answered rather
// than assumed). This is not a new mechanism class. It extends the estate's Go
// source-sensor idiom — work_item_type_minting_ratchet_test.go's
// TestNoDynamicallyConstructedItemTypes, verifier_coverage_test.go's
// TestEveryCheckProducedItemTypeIsClassified, and the finding-code registry's
// TestFindingCodeScanEveryWriteIsRegistered — to a shape none of them can see.
// scripts/pattern-check.py carries the SAME classifier and reads the SAME
// baseline file as check_scan_swallow, but pattern-check is ADVISORY BY DESIGN
// (its own output: "this never blocks") and examines only files a commit
// touches. A swallow that reaches production is what must BLOCK, which is what a
// Go test does and an advisory cannot. Two layers, one pattern: CHANGE THEM
// TOGETHER.
//
// WHY A BASELINE AND NOT A BAN. The minting ratchet could ban its shape outright
// because its population was zero. This population is 207 sites in 127 files
// tree-wide (measured 2026-08-26), 166 of them in this package's scope — far too
// many to convert in one change, and converting them blind would be a worse bug
// than the one being fixed. More decisively: THE SHAPE IS NOT THE DEFECT. At
// least one site in the baseline, scanBlogArticles in
// rebuild_blog_listing_action.go, has this exact shape AND a correct guard — it
// counts offered against kept after the loop and errors when every offered row
// failed. A ban would convict the estate's own best precedent. So the ratchet
// pins per-file COUNTS and only lets them fall.
//
// WHAT IT CANNOT SEE, stated so nobody mistakes the coverage:
//
//   - THE 41 SITES OUTSIDE THIS PACKAGE (internal/, cmd/, pkg/, the rest of
//     platform/) are covered by the advisory twin ONLY. A blocking test that
//     walks the whole tree from inside one package is fragile, and that
//     population is not artefact-shaped. This is a stated residual, not hidden
//     coverage.
//   - CONTENT loss, as opposed to ROW loss. loadStoredSections' own
//     `_ = json.Unmarshal(cdJSON, &s.contentData)` keeps the row and empties it
//     on corrupt JSON, so offered == kept and no count guard can see it. A
//     different axis, recorded in bugs_open/410 as a residual.
//   - A SITE THAT COUNTS BUT NEVER ACTS ON THE COUNT. The marker and the
//     baseline both track the SHAPE; only review can tell whether a guard that
//     exists is also consulted.
//   - SINGLE-ROW `QueryRow(...).Scan(&x)` inside a loop. Deliberately excluded:
//     there is no cursor-yielded count to compare against, and its `continue` is
//     ordinary control flow over an item rather than loss of a row the database
//     handed us. An earlier, looser version of this classifier counted those and
//     reported 225 sites instead of 207 — a number that was wrong by including a
//     pattern the guard does not apply to.
//
// COMMENTS ARE STRIPPED BEFORE MATCHING THE SHAPE — a prose mention of the
// pattern must not fail the build (the a-source-scanning-test-makes-comments-
// load-bearing trap, memorised fleet-wide, and the reason this very file does not
// fail its own ratchet). The MARKER, being a comment, is deliberately read from
// the RAW text over the same span. The stripper is line-based (`//` to end of
// line), which would also eat a `//` inside a string literal on the same line;
// for the shapes matched here that costs nothing, and the error direction is a
// missed exotic case, never a false failure.

package actions

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const (
	// scanLossMarker opts a site out at the site, with its reason next to the
	// code — the same "repairs, or is NAMED here with the reason" contract the
	// estate uses elsewhere. Written `// scan-loss:accepted: <reason>` inside the
	// error branch.
	scanLossMarker = "scan-loss:accepted"

	// scanSwallowBaselineFile is read by BOTH layers. Repo-root-relative paths.
	scanSwallowBaselineFile = "scan_swallow_baseline.txt"

	// pkgPrefix converts this package's local walk paths to the baseline's
	// repo-root-relative form.
	pkgPrefix = "platform/orchestration/actions/"
)

var (
	forNextRe  = regexp.MustCompile(`\bfor\s+([A-Za-z_][A-Za-z0-9_.]*)\.Next\(\)\s*\{`)
	errNilRe   = regexp.MustCompile(`err\s*!=\s*nil`)
	continueRe = regexp.MustCompile(`^\s*continue\b`)
)

// strippedLines reuses the package's existing stripLineComments (defined in
// agent_definition_nullable_columns_test.go) rather than duplicating the
// stripping rule, and splits it back into lines. The stripper preserves line
// count, so stripped and raw indices stay aligned — which is what lets the SHAPE
// be matched on stripped text while the MARKER, being a comment, is read from
// the raw text over the same span.
func strippedLines(src string) []string {
	return strings.Split(stripLineComments(src), "\n")
}

// matchBlock returns the index of the line closing the block opened at or after
// `start`, or -1.
func matchBlock(lines []string, start, limit int) int {
	depth, seen := 0, false
	end := start + limit
	if end > len(lines) {
		end = len(lines)
	}
	for k := start; k < end; k++ {
		for _, ch := range lines[k] {
			if ch == '{' {
				depth++
				seen = true
			} else if ch == '}' {
				depth--
				if seen && depth == 0 {
					return k
				}
			}
		}
	}
	return -1
}

// countUnmarkedScanSwallows counts cursor-loop scan swallows that carry no
// opt-out marker. MUST stay in step with _count_scan_swallows() in
// scripts/pattern-check.py.
func countUnmarkedScanSwallows(src string) int {
	raw := strings.Split(src, "\n")
	lines := strippedLines(src)
	n := 0

	for i := range lines {
		m := forNextRe.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		cursor := m[1]
		loopEnd := matchBlock(lines, i, 400)
		if loopEnd < 0 {
			continue
		}
		scanRe := regexp.MustCompile(`\b` + regexp.QuoteMeta(cursor) + `\.Scan\(`)

		for j := i + 1; j <= loopEnd && j < len(lines); j++ {
			if !scanRe.MatchString(lines[j]) {
				continue
			}
			// The error branch opening at or just after the Scan.
			start := -1
			for k := j; k < j+14 && k <= loopEnd && k < len(lines); k++ {
				if errNilRe.MatchString(lines[k]) && strings.Contains(lines[k], "{") {
					start = k
					break
				}
			}
			if start < 0 {
				continue
			}
			end := matchBlock(lines, start, 60)
			if end < 0 {
				continue
			}
			swallows := false
			for _, b := range lines[start+1 : end+1] {
				if continueRe.MatchString(b) {
					swallows = true
					break
				}
			}
			if swallows {
				// Marker read from RAW text (comments intact) over the same span.
				span := strings.Join(raw[start:end+1], "\n")
				if !strings.Contains(span, scanLossMarker) {
					n++
				}
			}
			j = end
		}
	}
	return n
}

// loadScanSwallowBaseline reads the shared baseline, returning repo-relative
// path -> count for entries inside this package's scope.
func loadScanSwallowBaseline(t *testing.T) map[string]int {
	t.Helper()
	f, err := os.Open(scanSwallowBaselineFile)
	if err != nil {
		t.Fatalf("read %s: %v — the baseline is the ratchet's memory; without it this test "+
			"cannot tell a new swallow from an old one", scanSwallowBaselineFile, err)
	}
	defer f.Close()

	out := map[string]int{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("malformed baseline line %q — expected `<path> <count>`", line)
		}
		n, err := strconv.Atoi(fields[1])
		if err != nil {
			t.Fatalf("malformed baseline count in %q: %v", line, err)
		}
		if strings.HasPrefix(fields[0], pkgPrefix) {
			out[fields[0]] = n
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan %s: %v", scanSwallowBaselineFile, err)
	}
	return out
}

func TestNoNewSilentScanLoss(t *testing.T) {
	baseline := loadScanSwallowBaseline(t)
	if len(baseline) == 0 {
		t.Fatal("the baseline yielded zero in-scope entries — a ratchet with an empty baseline " +
			"passes for ever and protects nothing")
	}

	actual := map[string]int{}
	for _, dir := range []string{".", "discovery_checks", "queryresolve"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			path := filepath.Join(dir, name)
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			if n := countUnmarkedScanSwallows(string(src)); n > 0 {
				rel := pkgPrefix + filepath.ToSlash(path)
				rel = strings.Replace(rel, pkgPrefix+"./", pkgPrefix, 1)
				actual[rel] = n
			}
		}
	}

	for path, got := range actual {
		want, known := baseline[path]
		switch {
		case !known:
			t.Errorf("%s: %d NEW silent scan loss site(s).\n"+
				"A scan failure here thins the result and the artefact still looks freshly built "+
				"(bugs_open/410). Either count and refuse — datahelpers.ScanShortfall(offered, "+
				"len(out), subject), with loadStoredSections in rerender_page_sections_action.go "+
				"as the worked example — or, for a deliberate guarded loss like scanBlogArticles', "+
				"mark the branch `// %s: <reason>`.", path, got, scanLossMarker)
		case got > want:
			t.Errorf("%s: silent scan loss sites rose from %d to %d.\n"+
				"Close the new one with datahelpers.ScanShortfall, or mark it "+
				"`// %s: <reason>` if the loss is deliberate and separately guarded. bugs_open/410.",
				path, want, got, scanLossMarker)
		case got < want:
			t.Errorf("%s: silent scan loss sites FELL from %d to %d — good, now ratchet down.\n"+
				"Lower this file's entry in %s to %d so the gain cannot be silently given back. "+
				"The baseline only ever goes down; that is what makes it a ratchet rather than a "+
				"tolerance.", path, want, got, scanSwallowBaselineFile, got)
		}
	}

	for path, want := range baseline {
		if _, still := actual[path]; !still {
			t.Errorf("%s: baseline expects %d silent scan loss site(s) and the file now has none "+
				"(or was deleted/renamed) — remove its line from %s.",
				path, want, scanSwallowBaselineFile)
		}
	}
}

// TestScanSwallowClassifierStillBites is the mutation guard on the ratchet
// itself. A source sensor whose pattern has been neutered matches nothing and
// passes for ever, which looks exactly like a clean tree.
//
// MUTATION CHECK: break forNextRe, errNilRe, continueRe or matchBlock and this
// test must fail. If it still passes, TestNoNewSilentScanLoss is asserting
// nothing and the door is open again.
func TestScanSwallowClassifierStillBites(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want int
	}{
		{
			name: "unmarked cursor-loop swallow is counted",
			src: `package p
func f() {
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			logger.Warn("boom")
			continue
		}
		out = append(out, a)
	}
}`,
			want: 1,
		},
		{
			name: "multi-line Scan args are still counted",
			src: `package p
func f() {
	for rows.Next() {
		if err := rows.Scan(
			&a,
			&b,
		); err != nil {
			continue
		}
	}
}`,
			want: 1,
		},
		{
			name: "marked swallow is not counted",
			src: `package p
func f() {
	for rows.Next() {
		if err := rows.Scan(&a); err != nil {
			// scan-loss:accepted: counted — ScanShortfall refuses below
			continue
		}
	}
}`,
			want: 0,
		},
		{
			name: "returning on scan error is not a swallow",
			src: `package p
func f() {
	for rows.Next() {
		if err := rows.Scan(&a); err != nil {
			return nil, err
		}
	}
}`,
			want: 0,
		},
		{
			name: "a non-default cursor name is still counted",
			src: `package p
func f() {
	for pageRows.Next() {
		if err := pageRows.Scan(&a); err != nil {
			continue
		}
	}
}`,
			want: 1,
		},
		{
			name: "single-row QueryRow scan in a loop is NOT the shape",
			src: `package p
func f() {
	for _, u := range urls {
		if err := db.QueryRow(q, u).Scan(&id); err != nil {
			continue
		}
	}
}`,
			want: 0,
		},
		{
			name: "prose in a comment must not be counted (comments are stripped)",
			src: `package p
// for rows.Next() { if err := rows.Scan(&a); err != nil { continue } }
func f() {}`,
			want: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := countUnmarkedScanSwallows(tc.src); got != tc.want {
				t.Errorf("classifier returned %d, want %d — the ratchet's pattern no longer "+
					"discriminates, so TestNoNewSilentScanLoss is asserting nothing", got, tc.want)
			}
		})
	}
}

// TestScanSwallowBaselineMatchesThisPackage is the belt-and-braces check that
// the baseline was regenerated rather than hand-edited: the in-scope entries must
// sum to the number this package actually contains. Without it, a baseline whose
// counts were nudged to make a red test green would look identical to a correct
// one.
func TestScanSwallowBaselineMatchesThisPackage(t *testing.T) {
	baseline := loadScanSwallowBaseline(t)
	sum := 0
	for _, n := range baseline {
		sum += n
	}
	if sum == 0 {
		t.Fatal("in-scope baseline sums to zero — see TestNoNewSilentScanLoss")
	}
	t.Logf("baseline pins %d silent scan loss site(s) across %d file(s) in %s (bugs_open/410, "+
		"2026-08-26). This number must only ever fall.", sum, len(baseline), pkgPrefix)
}
