// FILE: platform/orchestration/actions/findingcodes_scan_test.go
//
// bugs_open/358. The source-side EARLY WARNING: a finding code written by this
// package must carry a disposition in the registry, and you find out when you
// commit rather than when the CronJob goes red the next morning.
//
// WHY THIS FILE EXISTS, AND IT IS AN EMBARRASSING REASON. cmd/config-key-audit/
// findingcodes.go has said since 2026-08-22 that "a conservative Go source scan
// is kept as an EARLY WARNING at commit time (findingcodes_scan_test.go in the
// actions package)". THAT FILE DID NOT EXIST. The claim shipped, went through a
// council round, and was quoted in the concept register, and nobody opened the
// path it named. What did exist was TestPackageErrorCodeConstantsAreRegistered,
// which walks a HAND-WRITTEN LIST of eleven constants — so it can only catch a
// code somebody remembered to add to it, which is the one case that does not
// need catching. That is a third hand-maintained roster inside the change whose
// register entry boasts of retiring two, and it went stale by ADDITION exactly
// as this estate's owner ruling of 2026-08-22 says a census does.
//
// It was found by the mechanism itself: the CronJob's first live run, 2026-08-24,
// exited 1 on LINK_CONTEXT_UNAVAILABLE — a code added to THIS PACKAGE two hours
// earlier, past a "source-side early warning" that could not see it.
//
// WHAT IT DOES. Parses every non-test .go file in the package and collects the
// value of every `ErrorCode:` field in a composite literal, resolving a constant
// identifier through the package's own const declarations. Each value found must
// be declared in the registry. No list to keep in step: adding a code to this
// package is what puts it in scope.
//
// ⚠ ITS LIMITS, STATED, because an early warning that is believed to be complete
// is worse than none. It sees `ErrorCode:` fields only. It does NOT see:
//   - codes passed POSITIONALLY, e.g. LogActionError(ctx, params, siteID, domain,
//     action, code, …) — the trap bugs_open/358's handoff calls a fourth
//     blindness beyond the constant one;
//   - codes built at runtime, or read from config;
//   - writers outside this package (four of the five INSERT paths bypass the
//     agenterrors seam entirely, one of them in SQL).
//
// THE AUTHORITY IS AND REMAINS THE LIVE TABLE — `config-key-audit --finding-codes`
// against SELECT DISTINCT error_code, which is blind to none of those. This is a
// convenience that shortens the feedback loop from "tomorrow morning, in the
// cluster" to "now, on your machine". Anything it misses is still caught there,
// within a day of the code first firing.
package actions

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// scanPackageErrorCodes returns every literal value assigned to an `ErrorCode:`
// field anywhere in the package's non-test sources, with constant identifiers
// resolved through the package's own `const x = "..."` declarations.
//
// Returns the codes and the files it actually parsed, because a scan that
// silently parsed nothing is indistinguishable from a clean one — the same
// vacuity trap the DB-side check refuses on.
func scanPackageErrorCodes(t *testing.T) (codes []string, filesParsed int) {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("cannot read the package directory: %v", err)
	}

	fset := token.NewFileSet()
	consts := map[string]string{}
	var files []*ast.File

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			// A file this package cannot parse is a compile error that will be
			// reported far more legibly by the build itself. Skipping keeps this
			// test from becoming a second, worse compiler.
			continue
		}
		files = append(files, f)
		filesParsed++
	}

	// Pass 1 — every string constant in the package, so an `ErrorCode:` naming a
	// constant resolves. This is bugs_open/358 §3.2's trap ("grep the CONSTANT,
	// not just the literal") applied to the scanner rather than to a human.
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			vs, ok := n.(*ast.ValueSpec)
			if !ok {
				return true
			}
			for i, id := range vs.Names {
				if i >= len(vs.Values) {
					continue
				}
				if lit, ok := vs.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if s, err := strconv.Unquote(lit.Value); err == nil {
						consts[id.Name] = s
					}
				}
			}
			return true
		})
	}

	// Pass 2 — every `ErrorCode:` field value.
	seen := map[string]bool{}
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			kv, ok := n.(*ast.KeyValueExpr)
			if !ok {
				return true
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "ErrorCode" {
				return true
			}
			switch v := kv.Value.(type) {
			case *ast.BasicLit:
				if v.Kind == token.STRING {
					if s, err := strconv.Unquote(v.Value); err == nil && s != "" {
						seen[s] = true
					}
				}
			case *ast.Ident:
				if s, ok := consts[v.Name]; ok && s != "" {
					seen[s] = true
				}
			}
			return true
		})
	}

	for c := range seen {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	return codes, filesParsed
}

// TestEveryErrorCodeThisPackageWritesIsRegistered is the early warning proper.
//
// It replaces TestPackageErrorCodeConstantsAreRegistered's hand-written list of
// eleven names. That list could only ever catch a code somebody remembered to
// add to it — the one case that does not need catching — and it had already gone
// stale twice over by the time this was written.
func TestEveryErrorCodeThisPackageWritesIsRegistered(t *testing.T) {
	codes, filesParsed := scanPackageErrorCodes(t)

	// VACUITY GUARD. A scan that parsed nothing, or found nothing, produces a
	// clean pass indistinguishable from a healthy package — and this test's own
	// subject is checks that cannot fail. The package has carried dozens of
	// ErrorCode: sites continuously; zero means the scanner broke, not that the
	// package stopped writing findings.
	if filesParsed < 50 {
		t.Fatalf("parsed only %d source files — refusing to report a clean scan over a package "+
			"this size. The scanner is broken, not the package.", filesParsed)
	}
	if len(codes) < 10 {
		t.Fatalf("found only %d ErrorCode: values across %d files — refusing to report a clean "+
			"scan. An empty or near-empty result here is an instrument failure.", len(codes), filesParsed)
	}

	// THE RATCHET. Failing flat over the codes this package already writes and
	// nobody has declared would make this red on the day it ships, and
	// findingcodes.go's own design refuses exactly that: "a check that fails from
	// day one over a pre-existing backlog is a check that gets ignored". The
	// crisp signal is "someone just added a code nobody declared", so the
	// pre-existing set is recorded in the registry's `_scan_baseline` and only a
	// code absent from BOTH the declarations and that list fails.
	//
	// Same shape, and for the same reason, as the unruled cap and as
	// component-source-vocabulary-check's shrinking baseline.
	baseline := scanBaseline(t)
	declared := map[string]bool{}
	for _, c := range findingCodeRoster(t) {
		declared[c] = true
	}

	var undeclaredNew []string
	for _, code := range codes {
		if declared[code] || baseline[code] {
			continue
		}
		undeclaredNew = append(undeclaredNew, code)
	}
	for _, code := range undeclaredNew {
		t.Errorf("error code %q is written by this package, is not declared in %s, and is not in "+
			"`_scan_baseline` — so it is NEW. Declare it (consumed / instrumented / human-evidence / "+
			"operational, or `unruled` if the decision is genuinely open) in the same commit that "+
			"adds it. That is the whole point of catching it here rather than in tomorrow's CronJob "+
			"run: LINK_CONTEXT_UNAVAILABLE reached the live table on 2026-08-24 past a source-side "+
			"early warning that could not see it. bugs_open/358.", code, findingCodeRegistryRelPath)
	}

	// THE OTHER DIRECTION, reported and never failed: a baseline entry that has
	// since been declared. An acknowledgement list that outlives its findings is
	// how a check goes green by going blind (the shared-output-fields ack list
	// records the same rule). Removing the line is a one-word edit that should
	// ride the declaring commit.
	for code := range baseline {
		if declared[code] {
			t.Logf("STALE BASELINE: %q is now declared — delete it from `_scan_baseline` in %s. "+
				"The list may only shrink.", code, findingCodeRegistryRelPath)
		}
	}
}

// scanBaseline reads the pre-existing undeclared set. A MISSING key is an error,
// not an empty set: with no baseline every pre-existing code reads as new, the
// test goes red on the day it ships, and the next author deletes it — which is
// the failure mode the ratchet exists to avoid, arriving by a different door.
func scanBaseline(t *testing.T) map[string]bool {
	t.Helper()
	path := filepath.Join(moduleRoot(t), findingCodeRegistryRelPath)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("finding-code registry unreadable at %s: %v", path, err)
	}
	var doc struct {
		Baseline []string `json:"_scan_baseline"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("finding-code registry is not valid JSON: %v", err)
	}
	if doc.Baseline == nil {
		t.Fatalf("%s declares no `_scan_baseline`. It is a RATCHET and it may not simply vanish: "+
			"an absent list makes every pre-existing code read as new. If the backlog is genuinely "+
			"empty, declare it as an empty array and say so.", findingCodeRegistryRelPath)
	}
	out := make(map[string]bool, len(doc.Baseline))
	for _, c := range doc.Baseline {
		out[c] = true
	}
	return out
}

// TestTheScanFindsCodesReachedThroughAConstant pins the half that a literal-only
// scanner gets wrong, and which is why the original census missed a real reader
// (bugs_open/358 §3.2). Both of the codes that caught this package out on
// 2026-08-24 are of exactly this shape — `ErrorCode: linkContextUnavailableCode`
// — so a scanner without constant resolution would have stayed silent on them
// while looking like it worked.
func TestTheScanFindsCodesReachedThroughAConstant(t *testing.T) {
	codes, _ := scanPackageErrorCodes(t)
	got := map[string]bool{}
	for _, c := range codes {
		got[c] = true
	}

	// These are written as `ErrorCode: <const>`, never as a literal at the write
	// site. If the scan stops resolving constants, these disappear and every
	// assertion above passes vacuously.
	for _, want := range []string{
		"LINK_CONTEXT_UNAVAILABLE",
		"CONTENT_LINK_SUPPRESSED_UNSHIPPED",
	} {
		if !got[want] {
			t.Errorf("the scan did not find %q, which this package writes as `ErrorCode: <const>` — "+
				"constant resolution has stopped working, and without it this whole file passes "+
				"vacuously", want)
		}
	}
}

// TestTheHandListHoldsOnlyWhatTheScanCannotSee keeps the two mechanisms in a
// CHECKED relationship rather than a hoped-for one.
//
// codesInvisibleToTheScan (finding_code_roster_test.go) exists only for codes
// this scanner structurally cannot find — today, exactly one, passed
// positionally rather than as an `ErrorCode:` field. The moment such a write is
// converted to a field, the scan finds it and the hand-written entry becomes
// redundant: a roster entry that outlives its reason is how a list starts
// growing again. This fails when that happens, and names the line to delete.
//
// It is deliberately a FAILURE and not a report, unlike the stale-baseline case
// above: a redundant entry here costs nothing today but re-establishes exactly
// the hand-maintained surface bugs_open/358 spent a round retiring, and the fix
// is deleting one line.
func TestTheHandListHoldsOnlyWhatTheScanCannotSee(t *testing.T) {
	codes, _ := scanPackageErrorCodes(t)
	found := map[string]bool{}
	for _, c := range codes {
		found[c] = true
	}

	for _, c := range codesInvisibleToTheScan {
		if found[c] {
			t.Errorf("%q is in codesInvisibleToTheScan but the scan DOES find it — its write site "+
				"must have become an `ErrorCode:` field. Delete the entry: the scan covers it now, "+
				"and a hand-written list that outlives its reason is the drift this replaced.", c)
		}
	}

	// THE CONTROL, so this test cannot pass by the list being empty and the scan
	// being broken at the same time. The list is expected to be small and
	// non-empty while any positional write survives; an empty list means every
	// write site is discoverable, which is a real improvement and should be
	// recorded deliberately rather than arrived at by accident.
	if len(codesInvisibleToTheScan) == 0 {
		t.Log("codesInvisibleToTheScan is EMPTY — every finding code this package writes is now " +
			"discoverable by the scan. If that is true, say so where the list was; if it is not, " +
			"the list was deleted rather than emptied.")
	}
}
