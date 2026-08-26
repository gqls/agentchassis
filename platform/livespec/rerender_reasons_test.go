// FILE: platform/livespec/rerender_reasons_test.go
//
// bugs_open/404 — pins the re-render reason vocabulary's single definition.
//
// THE THIRD TEST IS THE ONE THAT MATTERS. It scans the migration corpus for
// re-render reason literals and asserts each is a declared value, so committing
// migration 460 (which appended `template_changed` to the live gate) with a
// three-value Go list would have failed HERE, naming the file and the value, on
// 2026-08-18 — instead of sitting latent until someone read the action to answer
// a council objection seven days later.
//
// It closes a window the daily auditor structurally cannot: the auditor compares
// the DECLARATION against the LIVE object once a morning, so it catches drift
// after it ships. This catches it at commit.
//
//	mutation                                              test that catches it
//	----------------------------------------------------  ---------------------------------------
//	delete a reason from the list                         CheckRerenderModeConditionClauseRoundTrips
//	scope literal_markdown (would SKIP dirty pages)        RerenderSectionReasonTableSemantics
//	give the legacy pair StampAlways (breaks REB-001)      RerenderSectionReasonTableSemantics
//	append a sixth reason to a .sql without declaring it   MigrationCorpusRerenderReasonLiteralsAreDeclared
package livespec

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCheckRerenderModeConditionClauseRoundTrips(t *testing.T) {
	got := CheckRerenderModeConditionClause()
	for _, r := range RerenderSectionReasons {
		if n := strings.Count(got, "input_data.spec.reason == '"+r.Name+"'"); n != 1 {
			t.Fatalf("%s appears %d times in the rendered condition, want exactly 1.\nrendered: %s",
				r.Name, n, got)
		}
	}
	if want, n := len(RerenderSectionReasons)-1, strings.Count(got, " OR "); n != want {
		t.Fatalf("separator count %d, want %d — the renderer must emit a flat disjunction, because "+
			"the live gate's condition is one and the Declaration asserts this string verbatim", n, want)
	}
	// The declared fragment IS this string; if they can differ the declaration is
	// asserting something nobody generated.
	d := MustGet("workflow.page-rerender.check_rerender_mode.reasons")
	if len(d.Fragments) != 1 || d.Fragments[0].Text != got {
		t.Fatalf("the gate Declaration's fragment must be exactly the renderer's output, or the " +
			"live tie is to a string with no source")
	}
}

// TestRerenderSectionReasonTableSemantics pins the JUDGEMENT in the table, with
// the reasoning in the failure text — the next author to edit it needs the why,
// not the assertion.
func TestRerenderSectionReasonTableSemantics(t *testing.T) {
	want := map[string]struct {
		scoped, stamp bool
		why           string
	}{
		"image_landed": {true, false,
			"REB-001's designed degrade: a reason without component_id falls back to assemble. " +
				"Preserving it byte-for-byte is not bugs_open/404's business"},
		"section_data_resolved": {true, false,
			"same REB-001 degrade, plus the action's own caution that a site-wide fan-out would " +
				"re-render unrelated sections"},
		"cta_links_stale": {false, true,
			"the CTA recompute is cheap and page-scoped, and a site-wide CTA repair has no single " +
				"triggering component to scope by"},
		"template_changed": {true, true,
			"component-caused, so scope when a component_id is given — but stamp EITHER WAY, because " +
				"assemble can never deliver a template change at all"},
		"literal_markdown": {false, true,
			"page-wide by meaning; scoping it to some component's dependents would silently SKIP " +
				"dirty pages, which is this bug's own shape"},
	}
	if len(RerenderSectionReasons) != len(want) {
		t.Fatalf("the vocabulary is %d values and this test knows %d — a new reason needs its "+
			"semantics decided HERE, not inferred at the call site",
			len(RerenderSectionReasons), len(want))
	}
	seen := map[string]bool{}
	for _, r := range RerenderSectionReasons {
		w, ok := want[r.Name]
		if !ok {
			t.Fatalf("undeclared semantics for %q", r.Name)
		}
		if seen[r.Name] {
			t.Fatalf("%q appears twice in the vocabulary", r.Name)
		}
		seen[r.Name] = true
		if r.ComponentScoped != w.scoped {
			t.Fatalf("%s ComponentScoped = %v, want %v — %s", r.Name, r.ComponentScoped, w.scoped, w.why)
		}
		if r.StampAlways != w.stamp {
			t.Fatalf("%s StampAlways = %v, want %v — %s", r.Name, r.StampAlways, w.stamp, w.why)
		}
		if r.Name != strings.ToLower(r.Name) || strings.TrimSpace(r.Name) != r.Name || r.Name == "" {
			t.Fatalf("reason %q must be a non-empty lower_snake token", r.Name)
		}
	}
}

// retiredRerenderReasons keeps the corpus lint from freezing the vocabulary.
// A value deliberately WITHDRAWN is listed here with its date, rather than making
// the list append-only for ever — which would be a new defect of this bug's own
// shape, one level up. Empty today, 2026-08-26.
var retiredRerenderReasons = map[string]string{}

var (
	// ⚠ BOTH QUOTE FORMS, AND THIS IS THE WHOLE TEST.
	//
	// A migration that edits the live condition writes it INSIDE a SQL string
	// literal, so its quotes are DOUBLED:
	//
	//	460:55   || ' OR input_data.spec.reason == '''template_changed'''
	//	473:78   'input_data.spec.reason == ''template_changed'' OR ... '
	//
	// while a PROSE COMMENT in the same file writes them singly:
	//
	//	460:23   --       OR input_data.spec.reason == 'template_changed'
	//
	// The first version of this regex accepted single quotes only. It matched 12
	// literals and PASSED — and every one of them was a comment. It could not see
	// either of the two executable lines it exists to catch. That is the landmine
	// "a source-scan test makes your COMMENTS load-bearing", and a guard that
	// cannot fail on its own motivating case. Caught by running the mutation,
	// never by reading the code.
	gateReasonLiteralRe = regexp.MustCompile(`input_data\.spec\.reason\s*==\s*'{1,2}([a-z0-9_]+)'{1,2}`)
	// A reason STAMPED into a spec by a producer's INSERT — same doubling.
	stampReasonLiteralRe = regexp.MustCompile(`'{1,2}reason'{1,2}\s*,\s*'{1,2}([a-z0-9_]+)'{1,2}`)
)

// corpusScanPositiveControls are values this scan MUST find, in files that
// definitely contain them, in their EXECUTABLE form. A regex that stops matching
// the real shape otherwise passes silently at full strength: the scan reports a
// healthy corpus because it read nothing, which is this bug one level up.
var corpusScanPositiveControls = map[string]string{
	"460_template_changed_rerender_reason.sql":   "template_changed",
	"473_literal_markdown_mechanical_repair.sql": "literal_markdown",
}

// TestMigrationCorpusRerenderReasonLiteralsAreDeclared is the commit-time half.
//
// ⚠ THE STAMP SCAN IS STATEMENT-SCOPED ON PURPOSE. A `'reason','template_truncated'`
// stamp inside a statement about `needs_component_regeneration` is a DIFFERENT
// vocabulary, and scooping it up here would make this test fail on correct code —
// which is how a guard gets weakened or deleted. So a stamp only counts when the
// statement it sits in also mentions 'page_rerender'.
func TestMigrationCorpusRerenderReasonLiteralsAreDeclared(t *testing.T) {
	dir := filepath.Join("..", "..", "docs", "agent_docs", "sql_for_agents")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migration corpus: %v", err)
	}

	declared := func(v string) bool {
		if _, ok := RerenderSectionReasonByName(v); ok {
			return true
		}
		_, retired := retiredRerenderReasons[v]
		return retired
	}

	checked := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		body := string(b)

		// (a) the GATE's own form — an allow-list disjunct.
		for _, m := range gateReasonLiteralRe.FindAllStringSubmatch(body, -1) {
			checked++
			if !declared(m[1]) {
				t.Errorf("%s writes the re-render gate value %q, which is NOT in "+
					"livespec.RerenderSectionReasons.\n"+
					"This is bugs_open/404 happening again: the live gate would know it and the Go "+
					"reader would not, so items carrying it would silently assemble.\n"+
					"Add it to the vocabulary (deciding ComponentScoped/StampAlways on its merits), "+
					"or to retiredRerenderReasons with a date if it is being withdrawn.", e.Name(), m[1])
			}
		}

		// (b) a producer STAMPING a reason onto a page_rerender item.
		for _, stmt := range strings.Split(body, ";") {
			if !strings.Contains(stmt, "page_rerender") {
				continue
			}
			for _, m := range stampReasonLiteralRe.FindAllStringSubmatch(stmt, -1) {
				checked++
				if !declared(m[1]) {
					t.Errorf("%s stamps reason %q onto a page_rerender item, and it is NOT in "+
						"livespec.RerenderSectionReasons — so page-rerender's gate will route it to "+
						"ASSEMBLE, re-shipping the stored HTML unchanged while completing green "+
						"(bugs_open/404).\n"+
						"If assemble is genuinely what you want, that is fine and it is what will "+
						"happen — but say so, because right now nobody can tell the two apart.",
						e.Name(), m[1])
				}
			}
		}
	}

	// NON-VACUITY, AND IT IS NOT ENOUGH ON ITS OWN — a count > 0 was true of the
	// first version of this test while it was matching only comments.
	if checked == 0 {
		t.Fatalf("the corpus scan matched NO reason literals at all — the regexes or the corpus path "+
			"are wrong, and a vacuous pass here is exactly the silent-green this test exists to "+
			"prevent (looked in %s)", dir)
	}

	// POSITIVE CONTROLS: the scan must find the real, executable literals in the
	// two migrations that caused this bug. If a later edit to the regex stops
	// matching the doubled-quote form, this fails LOUDLY instead of passing.
	for file, value := range corpusScanPositiveControls {
		b, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			t.Fatalf("positive control %s unreadable: %v", file, err)
		}
		found := false
		for _, m := range gateReasonLiteralRe.FindAllStringSubmatch(string(b), -1) {
			if m[1] == value {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("POSITIVE CONTROL FAILED: the scan cannot find %q in %s.\n"+
				"That file appends exactly that value to the live gate, so a scan that misses it "+
				"cannot catch the next one either — it will keep reporting a clean corpus while "+
				"blind. This is precisely how the first version of this test passed: it matched "+
				"the single-quoted PROSE COMMENTS and none of the doubled-quote executable SQL.",
				value, file)
		}
	}

	t.Logf("corpus scan: %d reason literal(s) checked, %d positive control(s) held",
		checked, len(corpusScanPositiveControls))
}

// TestReasonConstantsAreExactlyTheVocabulary ties the named constants and the
// table together in BOTH directions, so neither can drift from the other.
//
// It exists because the constants are used by readers OUTSIDE the item creator —
// the two per-value feature gates in rerender_page_sections_action.go — and the
// council objected, correctly, that leaving those as bare literals would reduce
// four copies of the vocabulary to two rather than one. With the constants named
// and pinned here, retiring or renaming a value breaks COMPILATION at every
// reader rather than silently disarming one of them.
func TestReasonConstantsAreExactlyTheVocabulary(t *testing.T) {
	consts := map[string]string{
		"ReasonImageLanded":         ReasonImageLanded,
		"ReasonSectionDataResolved": ReasonSectionDataResolved,
		"ReasonCTALinksStale":       ReasonCTALinksStale,
		"ReasonTemplateChanged":     ReasonTemplateChanged,
		"ReasonLiteralMarkdown":     ReasonLiteralMarkdown,
	}
	if len(consts) != len(RerenderSectionReasons) {
		t.Fatalf("%d named constants against %d vocabulary entries — a new reason needs a constant "+
			"too, or the readers outside create_rerender_items go back to bare literals",
			len(consts), len(RerenderSectionReasons))
	}
	for name, v := range consts {
		if _, ok := RerenderSectionReasonByName(v); !ok {
			t.Fatalf("%s = %q is not in RerenderSectionReasons — a constant naming a value the "+
				"vocabulary does not carry is a copy, which is what this bug is about", name, v)
		}
	}
	for _, r := range RerenderSectionReasons {
		found := false
		for _, v := range consts {
			if v == r.Name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("vocabulary value %q has no named constant, so a reader outside "+
				"create_rerender_items can only spell it as a bare literal", r.Name)
		}
	}
}
