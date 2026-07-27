package actions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The entries under test are HARVESTED fragments, not invented ones: shapes
// taken from what live implementations on vonc.com and fundamentallyai.com
// actually do (harvest/entries/*.json). A validator tested only against
// examples written to satisfy it proves nothing about the entries it will
// really see.

func validHarvestedEntry() map[string]interface{} {
	return map[string]interface{}{
		"name":         "feed-driven-teaser-list",
		"kind":         "component-contract",
		"display_name": "Feed-driven teaser list",
		"contract": map[string]interface{}{
			"triggers": []interface{}{
				map[string]interface{}{
					"when":             "a visitor activates a teaser row that has a detail body",
					"then":             "the detail body is revealed in place and the address bar carries the entry id",
					"destination_role": "self-state",
				},
				map[string]interface{}{
					"when": "a visitor activates a teaser row that has NO detail body",
					"then": "nothing happens, and no control was offered in the first place",
				},
			},
		},
		"requires_invariant": []interface{}{"no-inert-control"},
		"binding_schema": map[string]interface{}{
			"list_selector": map[string]interface{}{"type": "selector"},
			"row_selector":  map[string]interface{}{"type": "selector"},
		},
		"criteria_template": map[string]interface{}{
			"checks": []interface{}{
				map[string]interface{}{
					"id":       "list-present",
					"type":     "selector_exists",
					"selector": "{{binding.list_selector}}",
				},
				map[string]interface{}{
					"id":       "rows-present",
					"type":     "selector_count",
					"selector": "{{binding.row_selector}}",
				},
			},
		},
	}
}

func problemsFor(t *testing.T, entry map[string]interface{}) string {
	t.Helper()
	return strings.Join(validateExperiencePatternShape(entry), " | ")
}

func TestExperiencePatternShape_AcceptsAHarvestedEntry(t *testing.T) {
	if got := problemsFor(t, validHarvestedEntry()); got != "" {
		t.Fatalf("a harvested entry was refused: %s", got)
	}
}

func TestExperiencePatternShape_StatusIsNotWritable(t *testing.T) {
	// The rule that matters most: 'approved' is a council verdict and 'proven'
	// is a live green run. A writer that can assert either turns evidence into
	// a claim. Refusing beats ignoring — a silently dropped status lets the
	// caller believe it approved something.
	for _, status := range []string{"approved", "proven", "draft"} {
		entry := validHarvestedEntry()
		entry["status"] = status
		got := problemsFor(t, entry)
		if !strings.Contains(got, "status is not writable") {
			t.Errorf("status=%q was accepted or misreported: %s", status, got)
		}
	}
}

func TestExperiencePatternShape_DerivedCountsAreNotWritable(t *testing.T) {
	// Migration 230 refuses to store an entry as 'approved' with zero
	// executable checks. That constraint reads a column — so if an entry could
	// declare its own executable_checks, it would walk straight through the
	// gate the column exists to enforce. Three council seats raised the
	// underlying gap independently (corr bbdd2c5e); this is the half of the fix
	// that lives in Go.
	for _, derived := range []string{"executable_checks", "deferred_checks"} {
		entry := validHarvestedEntry()
		entry[derived] = 99
		got := problemsFor(t, entry)
		if !strings.Contains(got, derived+" is derived from validation") {
			t.Errorf("%s was accepted from the caller: %s", derived, got)
		}
	}
}

func TestExperienceDeferredRecords_EveryDeferralCarriesItsReason(t *testing.T) {
	// A deferral with no reason is indistinguishable from a check nobody could
	// be bothered to write, and the entire justification for allowing deferral
	// is that the clause stays in the record.
	v := ExperienceCriteriaValidation{
		Deferred: []ExperienceCriteriaIssue{
			{CheckID: "gauntlet-verdict-appears", Field: "expect_within_ms",
				Detail: "the API takes 8-23s; the runner asserts 300ms after the last step"},
		},
	}
	got := experienceDeferredRecords(v)
	if len(got) != 1 {
		t.Fatalf("expected one deferral, got %d", len(got))
	}
	for _, k := range []string{"check_id", "field", "reason"} {
		if got[0][k] == "" {
			t.Errorf("deferral record lost %q: %v", k, got[0])
		}
	}
}

func TestExperiencePatternShape_RefusesASiteSpecificDestination(t *testing.T) {
	// bugs_closed/045: a concrete value baked into a base entry is a static
	// fallback re-applied on every render, which a per-site binding cannot
	// override. The base entry names a ROLE; the page is bound per site.
	entry := validHarvestedEntry()
	contract := entry["contract"].(map[string]interface{})
	contract["triggers"] = []interface{}{
		map[string]interface{}{
			"when":        "a visitor activates a card",
			"then":        "the linked page loads",
			"destination": "/capabilities.html",
		},
	}
	got := problemsFor(t, entry)
	if !strings.Contains(got, "does not belong in a base entry") {
		t.Fatalf("a hard-coded destination was accepted: %s", got)
	}

	contract["triggers"] = []interface{}{
		map[string]interface{}{
			"when":             "a visitor activates a card",
			"then":             "the linked page loads",
			"destination_role": "/capabilities.html",
		},
	}
	if got := problemsFor(t, entry); !strings.Contains(got, "looks like a path or URL") {
		t.Fatalf("a URL wearing a role's name was accepted: %s", got)
	}
}

func TestExperiencePatternShape_TriggerNeedsAnObservableOutcome(t *testing.T) {
	// A trigger with no `then` is how "there is a button here" gets recorded as
	// behaviour — the exact defect the no-inert-control invariant exists for.
	entry := validHarvestedEntry()
	contract := entry["contract"].(map[string]interface{})
	contract["triggers"] = []interface{}{
		map[string]interface{}{"when": "a visitor clicks the arrow"},
	}
	if got := problemsFor(t, entry); !strings.Contains(got, "`then` is required") {
		t.Fatalf("a trigger with no observable outcome was accepted: %s", got)
	}
}

func TestExperiencePatternShape_ReportsEveryProblemAtOnce(t *testing.T) {
	// A writer that has to resubmit once per defect learns the contract one
	// refusal at a time; each round trip is an LLM call.
	entry := map[string]interface{}{
		"name":     "Not Kebab Case",
		"kind":     "widget",
		"status":   "approved",
		"contract": map[string]interface{}{},
	}
	problems := validateExperiencePatternShape(entry)
	if len(problems) < 5 {
		t.Fatalf("expected every problem at once, got %d: %v", len(problems), problems)
	}
	joined := strings.Join(problems, " | ")
	for _, want := range []string{"not kebab-case", "kind must be one of", "display_name is required",
		"status is not writable", "contract is required"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing problem %q in: %s", want, joined)
		}
	}
}

func TestChangedExperienceContractFields_CosmeticEditsDoNotDemote(t *testing.T) {
	prev := validHarvestedEntry()
	next := validHarvestedEntry()
	next["display_name"] = "Feed-driven teaser list (v2)"
	next["description"] = "a description that was not there before"
	if changed := changedExperienceContractFields(prev, next); len(changed) != 0 {
		t.Fatalf("cosmetic edits demoted an approved entry: %v", changed)
	}
}

func TestChangedExperienceContractFields_ClauseEditsDoDemote(t *testing.T) {
	prev := validHarvestedEntry()
	next := validHarvestedEntry()
	contract := next["contract"].(map[string]interface{})
	contract["triggers"] = append(contract["triggers"].([]interface{}),
		map[string]interface{}{"when": "the feed is empty", "then": "the section is not rendered"})

	changed := changedExperienceContractFields(prev, next)
	if len(changed) != 1 || changed[0] != "contract" {
		t.Fatalf("a contract change did not demote: %v", changed)
	}
}

func TestChangedExperienceContractFields_KeyOrderIsNotAChange(t *testing.T) {
	// A spurious demotion trains people to ignore the warning, which costs more
	// than the demotion saves. Round-tripping through JSON reorders map keys,
	// so comparison must be on canonical form.
	prev := validHarvestedEntry()
	raw, err := json.Marshal(prev)
	if err != nil {
		t.Fatal(err)
	}
	var next map[string]interface{}
	if err := json.Unmarshal(raw, &next); err != nil {
		t.Fatal(err)
	}
	if changed := changedExperienceContractFields(prev, next); len(changed) != 0 {
		t.Fatalf("a JSON round trip counted as a contract change: %v", changed)
	}
}

func TestChangedExperienceContractFields_AbsentFieldIsNotAnErasure(t *testing.T) {
	// A partial update must not silently empty a clause it never mentioned:
	// that would delete a rule and look like a no-op in the diff.
	prev := validHarvestedEntry()
	next := map[string]interface{}{
		"name":         prev["name"],
		"kind":         prev["kind"],
		"display_name": "renamed only",
	}
	if changed := changedExperienceContractFields(prev, next); len(changed) != 0 {
		t.Fatalf("fields absent from a partial update were treated as erased: %v", changed)
	}
}

func TestExperiencePatternColumns_MarshalsEveryJSONField(t *testing.T) {
	entry := validHarvestedEntry()
	for _, f := range experiencePatternJSONFields {
		if _, present := entry[f]; !present {
			entry[f] = []interface{}{"x"}
		}
	}
	cols, vals, err := experiencePatternColumns(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != len(vals) {
		t.Fatalf("columns and values disagree: %d vs %d", len(cols), len(vals))
	}
	got := map[string]bool{}
	for _, c := range cols {
		got[c] = true
	}
	for _, f := range experiencePatternJSONFields {
		if !got[f] {
			t.Errorf("jsonb column %q was not written", f)
		}
	}
	// Every jsonb value must reach the driver as a string of JSON, not as a Go
	// map — pgx would otherwise refuse the parameter at runtime, where it costs
	// an orchestration round trip to discover.
	for i, c := range cols {
		if !containsString(experiencePatternJSONFields, c) {
			continue
		}
		s, ok := vals[i].(string)
		if !ok {
			t.Errorf("column %q was passed as %T, not marshalled JSON", c, vals[i])
			continue
		}
		if !json.Valid([]byte(s)) {
			t.Errorf("column %q is not valid JSON: %s", c, s)
		}
	}
}

// TestExperiencePatternKinds_LockstepWithMigrationCheck holds the Go vocabulary
// to the database CHECK, the same idiom as
// TestValidDocSubjectTypes_LockstepWithMigrationCheck. A `kind` the DB accepts
// and this action rejects (or the reverse) is a split contract — the class that
// made migration 184's rows unreachable for months (bugs_closed/064).
func TestExperiencePatternKinds_LockstepWithMigrationCheck(t *testing.T) {
	migrationsDir := filepath.Join("..", "..", "..", "docs", "agent_docs", "sql_for_agents")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("cannot read migrations dir %s: %v", migrationsDir, err)
	}

	kindRE := regexp.MustCompile(`kind\s+text NOT NULL CHECK \(kind IN \(([^)]+)\)\)`)
	valueRE := regexp.MustCompile(`'([a-z-]+)'`)

	newest := -1
	var newestFile string
	var newestValues []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		numEnd := strings.IndexByte(name, '_')
		if numEnd <= 0 {
			continue
		}
		num, err := strconv.Atoi(name[:numEnd])
		if err != nil || num <= newest {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		m := kindRE.FindSubmatch(raw)
		if m == nil {
			continue
		}
		var values []string
		for _, v := range valueRE.FindAllStringSubmatch(string(m[1]), -1) {
			values = append(values, v[1])
		}
		newest, newestFile, newestValues = num, name, values
	}
	if newest < 0 {
		t.Fatal("no migration defining experience_patterns.kind found — if the CHECK moved or was reworded, update this test's regex rather than deleting it")
	}

	want := append([]string(nil), experiencePatternKinds...)
	got := append([]string(nil), newestValues...)
	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(want, ",") != strings.Join(got, ",") {
		t.Errorf("experiencePatternKinds and the CHECK in %s disagree:\n  Go: %v\n  DB: %v\nThey are ONE contract — move both together.",
			newestFile, want, got)
	}
}

// experiencePatternColumnsFromMigrations re-derives the table's columns from
// the migration files — the CREATE TABLE plus every later ADD COLUMN — so the
// tests below compare the Go lists against the schema rather than against
// another hand-maintained list.
func experiencePatternColumnsFromMigrations(t *testing.T) map[string]string {
	t.Helper()
	dir := filepath.Join("..", "..", "..", "docs", "agent_docs", "sql_for_agents")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("cannot read migrations dir %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names) // numeric prefixes, so lexical order is apply order here

	createRE := regexp.MustCompile(`(?s)CREATE TABLE IF NOT EXISTS experience_patterns \((.*?)\n\);`)
	colRE := regexp.MustCompile(`(?m)^\s{4}([a-z_]+)\s+(uuid|text|jsonb|integer|timestamptz)`)
	addRE := regexp.MustCompile(`ALTER TABLE experience_patterns\s+ADD COLUMN IF NOT EXISTS ([a-z_]+) ([a-z]+)`)

	cols := map[string]string{}
	for _, n := range names {
		raw, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			t.Fatalf("read %s: %v", n, err)
		}
		if m := createRE.FindSubmatch(raw); m != nil {
			for _, c := range colRE.FindAllStringSubmatch(string(m[1]), -1) {
				cols[c[1]] = c[2]
			}
		}
		for _, c := range addRE.FindAllStringSubmatch(string(raw), -1) {
			cols[c[1]] = c[2]
		}
	}
	if len(cols) == 0 {
		t.Fatal("no experience_patterns columns found in the migrations — if the table moved or the DDL was reworded, fix this parser rather than deleting the test")
	}
	return cols
}

// TestExperiencePatternColumns_EveryColumnIsClassified is the answer to the
// objection three council seats raised independently (corr 2e71f640): the
// demotion list was hand-maintained with no mechanical guard.
//
// Which fields are clause-bearing is a judgement, so it cannot be derived. What
// CAN be enforced is that the judgement was made at all: a column added to the
// table and classified into none of the four lists fails here. A silent
// omission becomes a required decision — which is the most a mechanical check
// can do for a question that is not mechanical.
func TestExperiencePatternColumns_EveryColumnIsClassified(t *testing.T) {
	cols := experiencePatternColumnsFromMigrations(t)

	classified := map[string]string{}
	for _, spec := range []struct {
		list  []string
		class string
	}{
		{experiencePatternContractFields, "contract"},
		{experiencePatternSelectionFields, "selection"},
		{experiencePatternCosmeticFields, "cosmetic"},
		{experiencePatternSystemFields, "system"},
	} {
		for _, f := range spec.list {
			if other, dup := classified[f]; dup {
				t.Errorf("column %q is classified twice (%s and %s) — the classes must partition, or demotion depends on list order",
					f, other, spec.class)
			}
			classified[f] = spec.class
		}
	}

	for col := range cols {
		if _, ok := classified[col]; !ok {
			t.Errorf("column %q exists in the migrations but is in NONE of experiencePatternContractFields / SelectionFields / CosmeticFields / SystemFields.\n"+
				"Classify it: does changing it invalidate an approval? If yes it belongs in the contract list, and leaving it out means a changed clause silently keeps 'approved'.", col)
		}
	}
	for col, class := range classified {
		if _, ok := cols[col]; !ok {
			t.Errorf("%q is classified as %s but is not a column of experience_patterns — a stale entry here makes the classification look complete when it is not", col, class)
		}
	}
}

// TestExperiencePatternJSONFields_CoversEveryJSONBColumn closes the gap
// bug_historian named: "a jsonb column added to experience_patterns but not to
// the hand-maintained marshalling list is silently never written — no error, no
// warning, no failed work item."
func TestExperiencePatternJSONFields_CoversEveryJSONBColumn(t *testing.T) {
	cols := experiencePatternColumnsFromMigrations(t)

	// deferred_checks is written from the validator's findings, not from the
	// caller's entry, so it is deliberately not in the marshalling list.
	writtenFromValidation := map[string]bool{"deferred_checks": true}

	for col, typ := range cols {
		if typ != "jsonb" || writtenFromValidation[col] {
			continue
		}
		if !containsString(experiencePatternJSONFields, col) {
			t.Errorf("jsonb column %q is not in experiencePatternJSONFields, so a caller can supply it and it will be SILENTLY never written", col)
		}
	}
	for _, f := range experiencePatternJSONFields {
		if cols[f] != "jsonb" {
			t.Errorf("experiencePatternJSONFields names %q, which is not a jsonb column of experience_patterns (got %q)", f, cols[f])
		}
	}
}

func TestMissingExperienceInvariants_NilDBIsNotAFalsePass(t *testing.T) {
	// The action must not report "no missing invariants" when it could not ask.
	// With no DB it returns nil, so the caller's own guard is what stands
	// between an entry and a dangling reference — this test pins the contract
	// so a future change to it is deliberate.
	missing, err := missingExperienceInvariants(t.Context(), nil, []interface{}{"no-such-invariant"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("expected the no-DB path to return nothing, got %v", missing)
	}
}

func TestStringSliceOf_HandlesBothJSONAndGoShapes(t *testing.T) {
	if got := stringSliceOf([]interface{}{"a", "", "b"}); strings.Join(got, ",") != "a,b" {
		t.Errorf("decoded JSON array: %v", got)
	}
	if got := stringSliceOf([]string{"a", "b"}); strings.Join(got, ",") != "a,b" {
		t.Errorf("Go slice: %v", got)
	}
	if got := stringSliceOf("not-an-array"); got != nil {
		t.Errorf("a bare string should not become a one-element list: %v", got)
	}
}
