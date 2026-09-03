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
// auditor-only declaration must be counted, so adding one forces the constant up
// and the set cannot grow quietly.
//
// Since phase 2 deployed (2026-08-23) these are no longer INERT — they are checked
// daily, just not by any Go test, because a unit test has no database.
func TestLiveAuditOnlyDeclarationsAreCounted(t *testing.T) {
	var auditorOnly []string
	for _, d := range Declarations {
		if d.Phase == PhaseLiveAudit {
			auditorOnly = append(auditorOnly, d.Key)
		}
	}
	if len(auditorOnly) != LiveAuditOnlyDeclarations {
		t.Fatalf("%d declaration(s) are checked ONLY by the daily auditor (%v), but "+
			"LiveAuditOnlyDeclarations says %d.\n"+
			"Bump the constant deliberately — an entry nobody counted reads as guarded by a test when it is not.",
			len(auditorOnly), auditorOnly, LiveAuditOnlyDeclarations)
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
		"subject_type CHECK — accumulation-aware, so it does not pin a stale file. Live tie is LIVE: " +
		"livespec constraint.doc_plans_subject_type_check / .doc_notes_subject_type_check.",
	"links_shipped_predicate_test.go": "migration 302's file is the operator's paste source, so keeping " +
		"the canonical predicate spelled correctly THERE is a real repo property. Live tie is LIVE: " +
		"livespec workflow.build-site-planner.load_existing_pages.",
	"v3_render_slot_name_test.go": "a seed lint: it asserts what the SEED says, which is a genuine repo " +
		"property. The seed is not the system; the live tie is LIVE: " +
		"livespec workflow.page-content-writer.slot_name_from.",
	"write_experience_pattern_test.go": "schema DDL, where the accumulated migration corpus IS the " +
		"canonical channel — whole statements, checksummed runner, no replace()-of-live indirection.",
	"contact_info_no_fabrication_test.go": "uses a migration's template body as a RENDER FIXTURE, not as " +
		"a claim about a live object; the file names its live watcher itself.",
	"rerender_reasons_test.go": "scans the WHOLE corpus for re-render reason literals and asserts each " +
		"is a declared value — accumulation-aware, pins no single file, and the live tie is LIVE " +
		"(livespec workflow.page-rerender.check_rerender_mode.reasons + .value_count). It exists to " +
		"close the window the ~24h auditor cadence leaves: migrations 460 and 473 each appended a " +
		"reason to the live gate on 2026-08-18 without telling the Go reader, and this is what would " +
		"have failed at COMMIT time on the day. bugs_open/404.",
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

// ── GAIN-BLINDNESS: A FragmentMatch DECLARATION CANNOT SEE AN ADDED VALUE ──────
//
// Measured 2026-08-25 (bugs_open/363), with the auditor deployed and running:
// FragmentMatch asserts each declared Fragment is PRESENT in the live text. So it
// catches the live object LOSING a declared value, and raises NOTHING when the
// live object GAINS an undeclared one — every declared fragment is still there,
// and the auditor prints the same "0 finding(s)" either way.
//
// That is this whole bug's founding shape: migration 357 declared two trigger
// bindings and 552 added a third, and the body-text comparison passed throughout.
//
// ⚠ NO PER-FRAGMENT Max CLOSES THIS. A Max bounds how often ONE declared value may
// appear; it says nothing about the SIZE OF THE SET, and a newly added value is in
// nobody's fragment list. Only a count assertion sees addition. Hence the rule:
//
//	every FragmentMatch declaration either has a paired "<key>.value_count"
//	CountEqual declaration, or a written waiver saying why its object is not an
//	enumerable vocabulary.
//
// The waiver is the same discipline as migrationReaderAllowList: a reason a human
// wrote, not a name on a list.
var gainBlindnessWaivers = map[string]string{
	"scheduled_task.claimed-item-timeout.exclusions": "ALREADY gain-visible, no pairing needed. Its single " +
		"fragment is the WHOLE rendered clause from ClaimedItemTimeoutExclusionClause(), terminator included: " +
		"\"item_type NOT IN ('a', ..., 'dark_section_audit')\". A 15th type added live makes the text read " +
		"\"..., 'dark_section_audit', 'new')\", so the declared substring — which ends at the closing paren — " +
		"stops matching and the auditor fires. Whole-clause fragments are self-bounding; per-value ones are not.",
	"scheduled_task.build-pipeline-trigger.retry_cooldown": "not an enumerable vocabulary. One boolean " +
		"predicate that is either spelled correctly or is not, plus a Forbidden fragment for the strict-'<' " +
		"spelling this bug's sibling was filed about. There is no set here whose size could grow.",
	"trigger_fn.site_component_history_archive": "a function BODY, not a list. Counting '::text' casts or " +
		"quoted literals in PL/pgSQL would be noise, not a vocabulary size. ⚠ STATED PLAINLY BECAUSE IT IS A " +
		"REAL RESIDUAL: a FOURTH verdict literal added to this function would be invisible to this entry. " +
		"classifySiteComponentArtefact is the Go mirror that would then disagree, and that is the tie we have.",
	"trigger_fn.page_component_artefact_archive": "same shape as its site_component twin — a function body. " +
		"Its dangerous growth direction is the TRIGGER SET rather than the body (357 declared 2 bindings, 552 " +
		"added a third), and that IS counted, by the separate CountEqual declaration " +
		"trigger_bindings.page_component_artefact_archive. Body-literal growth remains a stated residual.",
	"workflow.build-site-planner.load_existing_pages": "a single canonical predicate bounded Min 1 Max 1, " +
		"not a vocabulary. The Max already refuses a second occurrence, which is the only growth this " +
		"one-predicate declaration has.",
	"workflow.component-template-fixer.create_rerender": "a query BODY, not an enumerable vocabulary — " +
		"there is no set here whose SIZE could grow, so no count could be written. Its three fragments are " +
		"the three predicates that must hold: the reason stamp (Min 1 Max 1, so a second stamp is refused), " +
		"the owned-page guard, and the page-status filter migration 655 adds. ⚠ STATED PLAINLY BECAUSE IT IS " +
		"A REAL RESIDUAL: a rewrite that ADDS a predicate — narrowing the fan-out further — is invisible " +
		"here. That direction is the safe one for this query (it files fewer re-renders, never more), and " +
		"the dangerous direction is LOSS, which the Min:1 fragments catch. bugs_open/404.",
	"workflow.page-content-writer.prompt_item_shape": "two template SITES, not a vocabulary: there is no " +
		"set here whose size could grow, so no count could be written. Each is bounded Min 1 Max 1, so a " +
		"second copy of either directive stops the auditor, and the pre-437 flat spelling is declared " +
		"Forbidden, which catches the one regression that matters (a revert re-teaching the writer that a " +
		"nested array is a string). ⚠ STATED PLAINLY BECAUSE IT IS A REAL RESIDUAL: prompt text ADDED " +
		"elsewhere in this template — including another exemplar for another field shape — is invisible " +
		"here. The prompt is prose and cannot be counted; what is countable, and IS counted by " +
		"structured_item_shape_test.go, is the Go side that decides when these directives fire.",
	"workflow.page-content-writer.slot_name_from": "bounded Min 2 Max 2 as of 2026-08-25, so a third render " +
		"step setting slot_name_from stops the auditor. That is the growth direction that matters here; the " +
		"live object is two workflow steps, not an open-ended list of values.",
}

// EVERY FragmentMatch DECLARATION IS GAIN-VISIBLE OR EXPLICITLY WAIVED.
//
// Without this, the fix for bugs_open/363's own blind spot lasts exactly as long
// as the next person who adds a FragmentMatch entry without knowing about it.
func TestEveryFragmentMatchDeclarationIsGainVisibleOrWaived(t *testing.T) {
	paired := map[string]bool{}
	for _, d := range Declarations {
		if d.Mode == CountEqual {
			paired[d.Key] = true
		}
	}

	seen := map[string]bool{}
	for _, d := range Declarations {
		if d.Mode != FragmentMatch {
			continue
		}
		seen[d.Key] = true
		if paired[d.Key+".value_count"] {
			continue
		}
		reason, ok := gainBlindnessWaivers[d.Key]
		if !ok {
			t.Errorf("%s is FragmentMatch with no paired %q CountEqual declaration and no waiver.\n"+
				"FragmentMatch asserts PRESENCE, so it sees the live object LOSE a declared value and is BLIND "+
				"to it GAINING an undeclared one — and the auditor's clean run looks identical either way "+
				"(bugs_open/363, measured 2026-08-25).\n"+
				"Either add the paired count declaration, or add a waiver to gainBlindnessWaivers saying why "+
				"this object is not an enumerable vocabulary. A Max on a fragment does NOT close this: it "+
				"bounds one value's occurrences, never the size of the set.", d.Key, d.Key+".value_count")
			continue
		}
		if len(strings.TrimSpace(reason)) < 60 {
			t.Errorf("gain-blindness waiver for %s is too thin to be a reason (%q) — an allow-list without a "+
				"real reason converts a live debt into a false all-clear", d.Key, reason)
		}
	}

	// A waiver that outlives its declaration is the same defect one level up: it
	// reads as a considered decision about something that is no longer there.
	for key := range gainBlindnessWaivers {
		if !seen[key] {
			t.Errorf("gain-blindness waiver names %q, which is not a FragmentMatch declaration any more "+
				"(renamed, removed, or converted to CountEqual). Delete the waiver — a stale one hides "+
				"whatever now occupies that name.", key)
		}
		if paired[key+".value_count"] {
			t.Errorf("%s has BOTH a waiver and a paired count declaration. Keep the pairing and delete the "+
				"waiver: two answers to one question is how the wrong one survives.", key)
		}
	}

	// A guard that examined nothing proves nothing.
	if len(seen) == 0 {
		t.Fatal("no FragmentMatch declarations were examined — this guard is vacuous")
	}
}
