package discovery_checks

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// ── bugs_open/170: the pin predicate exists TWICE, and it has to ─────────────
//
// `deactivated_site_components` asks "does this site's chrome point at something
// that cannot serve chrome?" of BOTH stores — the per-site assignment
// (site_components) and the collection-level pin (style_collections). The pin half
// needs the same eligibility predicate the render and link paths use, which lives
// in package `actions` as chromePinEligibleSQL.
//
// This package CANNOT import it: `actions` imports `discovery_checks` to register
// the checks, so the dependency only runs one way. The predicate is therefore
// hand-typed here, which is exactly the drift class that produced bugs_open/118 in
// the first place — three hand-written answers to one question, each wrong
// differently, none aware of the others.
//
// So the copy is guarded instead of forbidden. This test reads the OTHER package's
// source and fails if the two have drifted apart. Same construction as the
// dedup-index ↔ Go-list lockstep: two artefacts that must stay identical, with a
// test standing between them.

const chromePinSourceFile = "../component_library.go"
const chromePinCheckFile = "check_integrity.go"

// chromeLevelsConst extracts the shared level whitelist from its single definition.
var chromeLevelsConst = regexp.MustCompile(`chromeComponentLevels\s*=\s*` + "`" + `([^` + "`" + `]*)` + "`")

func TestChromePinPredicateMatchesTheActionsPackage(t *testing.T) {
	src, err := os.ReadFile(chromePinSourceFile)
	if err != nil {
		t.Fatalf("read %s: %v — this lockstep has gone blind, so a pass means nothing", chromePinSourceFile, err)
	}

	m := chromeLevelsConst.FindSubmatch(src)
	if m == nil {
		t.Fatalf("could not find chromeComponentLevels in %s — either it was renamed (in which case "+
			"update this test) or this lockstep has gone blind", chromePinSourceFile)
	}
	wantLevels := strings.Fields(strings.ReplaceAll(string(m[1]), ",", " "))
	if len(wantLevels) == 0 {
		t.Fatal("chromeComponentLevels parsed as empty — the extraction is wrong, not the code")
	}

	// The predicate itself, so a clause added there is not silently missed here.
	// chromePinEligibleSQL omits `forked_from IS NULL` on purpose (a pin naming the
	// site's own fork is legitimate); if that ever changes, this check must change
	// with it, and the assertion below is what forces the conversation.
	pinFn := extractFuncBody(string(src), "func chromePinEligibleSQL(")
	if pinFn == "" {
		t.Fatal("chromePinEligibleSQL not found in the actions package — this lockstep has gone blind")
	}

	check, err := os.ReadFile(chromePinCheckFile)
	if err != nil {
		t.Fatalf("read %s: %v", chromePinCheckFile, err)
	}
	got := string(check)

	// The level lists as the CHECK writes them, not as the file happens to contain
	// them.
	//
	// > **CORRECTED before shipping — the first version searched the whole file.**
	// > Mutation-tested by narrowing the check's list to ('site','header'), it
	// > reported only 'head' as missing: `'footer'` still appeared elsewhere in
	// > check_integrity.go as a SLOT NAME (`SELECT 'footer', fc.name …`), so a
	// > file-wide containment test passed on a string that had nothing to do with
	// > the predicate. It caught that drift by luck, on the one level with no
	// > homonym. Scoped to the `component_level IN (…)` clauses, it cannot.
	clauses := componentLevelClauses(got)
	if len(clauses) == 0 {
		t.Fatal("no `component_level IN (…)` clause found in the pin check — either the check lost its " +
			"level filter entirely, or this lockstep has gone blind")
	}
	for _, clause := range clauses {
		have := strings.Fields(strings.ReplaceAll(clause, ",", " "))
		// 1. Every level in the shared whitelist must be in this clause.
		for _, lvl := range wantLevels {
			if !containsField(have, lvl) {
				t.Errorf("a pin-check level clause is missing %s, which chromeComponentLevels carries "+
					"(clause: %s). A narrower list here means the check silently stops reporting pins "+
					"the render path silently stops honouring — the two must agree or the detector "+
					"lies about the fix.", lvl, clause)
			}
		}
		// 2. …and no EXTRA level, which would make the check quieter than the code.
		for _, lvl := range have {
			if !containsField(wantLevels, lvl) {
				t.Errorf("a pin-check level clause admits %s, which the shared whitelist does not "+
					"(clause: %s). A pin at that level would be rejected by the code and reported as "+
					"fine by the check.", lvl, clause)
			}
		}
	}

	// 3. Both clauses of the predicate must be present.
	for _, clause := range []string{"is_active", "component_level"} {
		if !strings.Contains(pinFn, clause) {
			t.Fatalf("chromePinEligibleSQL no longer carries %s — this test's premise is stale", clause)
		}
		if !strings.Contains(got, clause) {
			t.Errorf("the pin check does not test %s, which chromePinEligibleSQL does", clause)
		}
	}

	// 4. The fork asymmetry, from the other side. If someone adds forked_from to
	// the shared predicate, this check must gain it too — and if someone adds it
	// HERE alone, the check starts reporting leopardessconsulting.co.uk's own
	// header as a defect while the code keeps serving it.
	srcHasFork := strings.Contains(pinFn, "forked_from")
	checkHasFork := strings.Contains(got, "forked_from")
	if srcHasFork != checkHasFork {
		t.Errorf("fork clause disagrees: chromePinEligibleSQL has forked_from=%v, the pin check has %v. "+
			"Whichever way round, the detector and the code now answer differently for a site pinned "+
			"to its own fork.", srcHasFork, checkHasFork)
	}
}

// Proof the lockstep can actually fail. Without it a broken extraction reports a
// clean tree for ever — the failure mode this file exists to prevent, applied to
// itself.
func TestChromePinLockstepFiresOnDrift(t *testing.T) {
	levels := chromeLevelsConst.FindSubmatch([]byte("const chromeComponentLevels = `'site', 'header'`"))
	if levels == nil {
		t.Fatal("the level extraction no longer matches its own declaration form — it has gone inert")
	}
	got := strings.Fields(strings.ReplaceAll(string(levels[1]), ",", " "))
	if len(got) != 2 || got[0] != "'site'" || got[1] != "'header'" {
		t.Fatalf("extraction returned %q, want ['site' 'header']", got)
	}

	// A check SQL missing one of them must be detectable by the same containment
	// test the real assertion uses.
	drifted := "WHERE component_level IN ('site')"
	if strings.Contains(drifted, "'header'") {
		t.Fatal("the containment test cannot distinguish a narrowed list — the assertion is vacuous")
	}

	// And the function-body extractor must return something for a real function.
	body := extractFuncBody("func a() {}\nfunc chromePinEligibleSQL(alias string) string {\n\treturn alias + \"is_active\"\n}\nfunc b() {}", "func chromePinEligibleSQL(")
	if !strings.Contains(body, "is_active") || strings.Contains(body, "func b()") {
		t.Fatalf("extractFuncBody returned %q — it must capture the function and stop at the next one", body)
	}
}

// componentLevelIn captures the contents of each `component_level IN (…)` clause.
var componentLevelIn = regexp.MustCompile(`component_level\s+IN\s*\(([^)]*)\)`)

// componentLevelClauses returns the level list from every such clause in src, so
// the assertion reads the predicate rather than the file.
func componentLevelClauses(src string) []string {
	var out []string
	for _, m := range componentLevelIn.FindAllStringSubmatch(src, -1) {
		out = append(out, strings.TrimSpace(m[1]))
	}
	return out
}

func containsField(fields []string, want string) bool {
	for _, f := range fields {
		if f == want {
			return true
		}
	}
	return false
}

// extractFuncBody returns the source of one function, from its declaration to the
// next top-level `func`.
func extractFuncBody(src, decl string) string {
	start := strings.Index(src, decl)
	if start < 0 {
		return ""
	}
	rest := src[start+len(decl):]
	end := strings.Index(rest, "\nfunc ")
	if end < 0 {
		return src[start:]
	}
	return src[start : start+len(decl)+end]
}
