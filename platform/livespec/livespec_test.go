// FILE: platform/livespec/livespec_test.go
//
// livespec guards itself, and guards the class from regrowing.
//
// A registry that can hold prose, a dead probe, a duplicate key or a silently
// unchecked entry is migration 482's comment one rung up — so these tests exist
// before any of the guards that import it.
package livespec

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKeysAreUniqueAndKindsDeclared(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range Declarations {
		if d.Key == "" {
			t.Error("a Declaration has no Key — the auditor and the guards both address entries by key")
		}
		if seen[d.Key] {
			t.Errorf("duplicate Declaration key %q — one live object, one entry", d.Key)
		}
		seen[d.Key] = true
		if d.Kind == "" {
			t.Errorf("%s: no Kind — the auditor reports scope by kind, and a blank one makes 'probed N objects' unreadable", d.Key)
		}
		if d.Provenance == "" {
			t.Errorf("%s: no Provenance — an entry whose history nobody wrote down is how 482's comment happened", d.Key)
		}
	}
}

// A probe must be a single read-only SELECT. The phase-2 auditor runs these
// against production; a probe that is two statements, or that is not a SELECT, is
// a write hazard hiding in a declaration file.
func TestProbesAreSingleReadOnlySelects(t *testing.T) {
	for _, d := range Declarations {
		if d.ProbeSQL == "" {
			t.Errorf("%s: no ProbeSQL — nothing can ever check this entry", d.Key)
			continue
		}
		if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(d.ProbeSQL)), "SELECT") {
			t.Errorf("%s: ProbeSQL does not start with SELECT: %q", d.Key, d.ProbeSQL)
		}
		if strings.Contains(d.ProbeSQL, ";") {
			t.Errorf("%s: ProbeSQL contains ';' — one statement per probe", d.Key)
		}
	}
}

func TestFragmentsAreCoherent(t *testing.T) {
	for _, d := range Declarations {
		for _, f := range d.Fragments {
			if f.Text == "" {
				t.Errorf("%s: empty Fragment text", d.Key)
			}
			if f.Forbidden && (f.Min > 0 || f.Max > 0) {
				t.Errorf("%s: fragment %q is Forbidden AND carries Min/Max — say one thing", d.Key, f.Text)
			}
			if f.Max > 0 && f.Min > f.Max {
				t.Errorf("%s: fragment %q has Min %d > Max %d", d.Key, f.Text, f.Min, f.Max)
			}
		}
		if d.Mode == CountEqual && d.ExpectCount <= 0 {
			t.Errorf("%s: CountEqual with ExpectCount %d — a zero expectation passes on an empty probe", d.Key, d.ExpectCount)
		}
	}
}

// Every item type must look like an item type. The shape check is here because
// migration 482's own round proved it necessary: a regex that scooped up prose
// reported fourteen confusing roster errors instead of the one true cause.
func TestExclusionItemTypesHaveItemTypeShape(t *testing.T) {
	for _, itemType := range ClaimedItemTimeoutExclusions {
		if itemType == "" {
			t.Fatal("empty item_type in the exclusion list")
		}
		if itemType != strings.ToLower(itemType) || strings.ContainsAny(itemType, " '\"(),") {
			t.Errorf("%q is not an item_type — the declaration has picked up prose", itemType)
		}
	}
}

func TestExclusionClauseRoundTrips(t *testing.T) {
	clause := ClaimedItemTimeoutExclusionClause()
	if !strings.HasPrefix(clause, "item_type NOT IN (") || !strings.HasSuffix(clause, ")") {
		t.Fatalf("clause is not the SQL fragment the live pre_query carries: %q", clause)
	}
	for _, itemType := range ClaimedItemTimeoutExclusions {
		if !strings.Contains(clause, "'"+itemType+"'") {
			t.Errorf("clause omits %q", itemType)
		}
	}
	if got, want := strings.Count(clause, ","), len(ClaimedItemTimeoutExclusions)-1; got != want {
		t.Errorf("clause has %d separators for %d types", got, len(ClaimedItemTimeoutExclusions))
	}
}

// The cooldown boundary must stay non-strict. With "<" an item is claimable a
// moment before it is completable and the disagreement only shows up as a race.
func TestCooldownBoundaryIsNonStrict(t *testing.T) {
	if !strings.Contains(WorkItemRetryNotPendingAliased, "<= NOW()") {
		t.Errorf("declared boundary is not non-strict: %q", WorkItemRetryNotPendingAliased)
	}
	if strings.Contains(WorkItemRetryNotPendingAliased, "< NOW()") &&
		!strings.Contains(WorkItemRetryNotPendingAliased, "<= NOW()") {
		t.Error("declared boundary is STRICT")
	}
}

// GROWTH BOUNDARY (council: architecture). livespec is a registry of guarded live
// objects, not a general config store.
func TestRegistryStaysWithinItsGrowthBoundary(t *testing.T) {
	if len(Declarations) > MaxDeclarations {
		t.Fatalf("livespec holds %d declarations, boundary is %d — sprawl is a scope decision for a human, not a drift",
			len(Declarations), MaxDeclarations)
	}
}

// DECLARED-BUT-NEVER-READ IS VISIBLE, NOT IMPLIED (council: bug_historian, round 2).
// A field accepted but never read is indistinguishable from one that works. Every
// inert declaration must be counted, so adding one forces DeferredDeclarations up
// and the gap cannot grow quietly.
func TestInertDeclarationsAreCounted(t *testing.T) {
	var inert []string
	for _, d := range Declarations {
		if d.Phase == PhaseLiveAudit {
			inert = append(inert, d.Key)
		}
	}
	if len(inert) != DeferredDeclarations {
		t.Fatalf("%d declaration(s) are inert until the phase-2 auditor (%v), but DeferredDeclarations says %d.\n"+
			"Bump the constant deliberately — an inert entry that nobody counted reads as guarded when it is not.",
			len(inert), inert, DeferredDeclarations)
	}
}

func TestCompareFragmentsDetectsEachViolation(t *testing.T) {
	d := Declaration{
		Key:  "test",
		Mode: FragmentMatch,
		Fragments: []Fragment{
			{Text: "REQUIRED", Min: 1, Max: 1},
			{Text: "BANNED", Forbidden: true},
		},
	}
	if p := d.CompareFragments("... REQUIRED ..."); len(p) != 0 {
		t.Errorf("clean text reported problems: %v", p)
	}
	if p := d.CompareFragments("nothing here"); len(p) != 1 {
		t.Errorf("missing required fragment: want 1 problem, got %v", p)
	}
	if p := d.CompareFragments("REQUIRED REQUIRED"); len(p) != 1 {
		t.Errorf("duplicated fragment past Max: want 1 problem, got %v", p)
	}
	if p := d.CompareFragments("REQUIRED and BANNED"); len(p) != 1 {
		t.Errorf("forbidden fragment: want 1 problem, got %v", p)
	}
}

func TestCompareCountParsesTextAndRefusesNonIntegers(t *testing.T) {
	d := Declaration{Key: "test", Mode: CountEqual, ExpectCount: 3}
	if p := d.CompareCount("3"); len(p) != 0 {
		t.Errorf("matching count reported problems: %v", p)
	}
	if p := d.CompareCount(" 3\n"); len(p) != 0 {
		t.Errorf("whitespace-padded count should parse: %v", p)
	}
	if p := d.CompareCount("2"); len(p) != 1 {
		t.Errorf("drifted count: want 1 problem, got %v", p)
	}
	// A probe that returns nothing usable must be a PROBLEM, never a silent pass —
	// the nil-result-passes shape this estate keeps re-learning.
	if p := d.CompareCount(""); len(p) != 1 {
		t.Errorf("empty probe result must be a problem, got %v", p)
	}
}

// readCalls are the ways a test can pull a migration off disk.
var readCalls = map[string]bool{"ReadFile": true, "Glob": true, "ReadDir": true}

// migrationReaderAllowList names every test file that may still read a path under
// sql_for_agents, WITH THE REASON. An allow-list without reasons converts a live
// debt into a false all-clear, which is a documented landmine on this estate — so
// the reason is a required part of the entry, not a courtesy.
//
// The four guards converted by bugs_open/363 are deliberately ABSENT: they read
// livespec now. If one reappears here, this test fires, which is the point.
var migrationReaderAllowList = map[string]string{
	"doc_subjects_common_test.go": "scans the whole corpus for the NEWEST migration recreating each " +
		"subject_type CHECK — accumulation-aware, so it does not pin a stale file. Live tie is phase 2.",
	"links_shipped_predicate_test.go": "migration 302's file is the operator's paste source, so keeping " +
		"the canonical predicate spelled correctly THERE is a real repo property. Live tie is phase 2.",
	"v3_render_slot_name_test.go": "a seed lint: it asserts what the SEED says, which is a genuine repo " +
		"property. The seed is not the system, and the live tie is phase 2.",
	"write_experience_pattern_test.go": "schema DDL, where the accumulated migration corpus IS the " +
		"canonical channel — whole statements, checksummed runner, no replace()-of-live indirection.",
	"contact_info_no_fabrication_test.go": "uses a migration's template body as a RENDER FIXTURE, not as " +
		"a claim about a live object; the file names its live watcher itself.",
}

// TestNoNewMigrationFileReadersOutsideTheAllowList stops the class regrowing.
//
// It scans STRING LITERALS IN THE AST, never comments: a source-scanning test that
// reads comments makes comments load-bearing, and first occurrence wins. A file
// counts only if it BOTH mentions sql_for_agents in code AND calls a read — which
// is what separates a real reader from a file that merely names the path in prose
// or holds it in an unrelated fixture map.
func TestNoNewMigrationFileReadersOutsideTheAllowList(t *testing.T) {
	root := filepath.Join("..", "..", "platform")
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("cannot reach platform/ from the livespec package dir (%v) — this guard is inert until the path is fixed", err)
	}

	var offenders []string
	scanned := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return err
		}
		fset := token.NewFileSet()
		file, perr := parser.ParseFile(fset, path, nil, 0) // no ParseComments: comments must not count
		if perr != nil {
			return nil // an unparseable test file is the compiler's problem, not this guard's
		}
		scanned++

		namesMigrationPath, callsRead := false, false
		ast.Inspect(file, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.BasicLit:
				if v.Kind == token.STRING && strings.Contains(v.Value, "sql_for_agents") {
					namesMigrationPath = true
				}
			case *ast.CallExpr:
				if sel, ok := v.Fun.(*ast.SelectorExpr); ok && readCalls[sel.Sel.Name] {
					callsRead = true
				}
			}
			return true
		})

		if namesMigrationPath && callsRead {
			if _, allowed := migrationReaderAllowList[filepath.Base(path)]; !allowed {
				offenders = append(offenders, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking platform/: %v", err)
	}

	// A guard that scanned nothing proves nothing.
	if scanned == 0 {
		t.Fatal("scanned zero test files — the walk is broken and this guard is vacuous")
	}

	for _, o := range offenders {
		t.Errorf("%s reads a path under sql_for_agents and is not on the allow-list.\n"+
			"A migration is append-only history: the checksum in schema_migrations means an applied file "+
			"cannot change, so asserting its TEXT is an assertion that cannot fail, while the live object it "+
			"once described keeps moving (bugs_open/363).\n"+
			"Declare what the live object should contain in platform/livespec and assert against THAT — or, if "+
			"this file genuinely has a repo-side reason to read the corpus, add it to migrationReaderAllowList "+
			"WITH that reason.", o)
	}
}

// The allow-list must not rot into a list of names with no reasons.
func TestEveryAllowListEntryCarriesAReason(t *testing.T) {
	for file, reason := range migrationReaderAllowList {
		if len(strings.TrimSpace(reason)) < 40 {
			t.Errorf("allow-list entry %q has no real reason (%q) — that is how a live debt becomes a false all-clear", file, reason)
		}
	}
}
