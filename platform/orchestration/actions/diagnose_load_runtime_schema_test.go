// Tests for the Schema-section visibility fix (owner ruling 2026-08-10, item 5:
// "let the diagnosis loop see the orchestration_states file").
//
// The defect these guard: gatherSchema applied a relevance include
// (site%|page%|content%|flow%) that selected 26 of 433 live public tables, and
// FIVE of the six tables this action renders evidence rows from fell outside it.
// The section was headed "Schema (live tables)" and said nothing about being
// filtered, so a filtered-out table and a non-existent table rendered
// identically. 090 run 074beb8a then guessed `WHERE id = …` on
// orchestration_states (the column is `orchestration_id`), got SQLSTATE 42703,
// and stopped at UNVERIFIABLE asking for a human to confirm the table's real id
// column. Two runs on bugs_open/236 died that way.
package actions

import (
	"context"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestSchemaAlwaysTablesCoverTablesThisActionQueries re-derives the always-list
// from the SOURCE and fails when a query names a table nobody added to it. This
// is what stops schemaAlwaysTables going stale the way the include list did —
// the include was correct when written and silently wrong once the gather grew
// sections that read other tables.
//
// Comments are STRIPPED before scanning. A source-scanning test otherwise makes
// prose load-bearing: this very file's comments name orchestration_states and
// llm_call_log, and a scan that read them would pass by quoting itself.
func TestSchemaAlwaysTablesCoverTablesThisActionQueries(t *testing.T) {
	src, err := os.ReadFile("diagnose_load_runtime_action.go")
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	code := stripGoComments(string(src))

	// Scan the BACKTICK literals only — that is where every SQL statement in this
	// file lives. Scanning all code matched prose inside ordinary string literals
	// ("…answered from the code_symbols index" yielded a table called "the"), and
	// widening notTables to absorb English words would have made the guard weaker
	// with every sentence added.
	var sql strings.Builder
	for _, lit := range regexp.MustCompile("(?s)`([^`]*)`").FindAllStringSubmatch(code, -1) {
		sql.WriteString(lit[1])
		sql.WriteString("\n")
	}

	// Anything the scan finds that is not a real base table in this schema.
	notTables := map[string]bool{
		"information_schema": true, // schema-qualified: information_schema.columns
	}

	found := map[string]bool{}
	re := regexp.MustCompile(`(?i)\b(?:FROM|JOIN)\s+([a-z_][a-z0-9_]*)`)
	for _, m := range re.FindAllStringSubmatch(sql.String(), -1) {
		tbl := strings.ToLower(m[1])
		if notTables[tbl] {
			continue
		}
		found[tbl] = true
	}
	if len(found) == 0 {
		// The scan silently finding nothing would make this test vacuous — the
		// failure mode where a "passing" guard proves only that it did not look.
		t.Fatal("source scan found no table references at all; the scan is broken, not the code")
	}

	// The scan can only see SQL in backticks. A query written with a
	// double-quoted string would be invisible to it and the test would PASS while
	// blind — the shape this estate files as "a check that answers the question it
	// encoded". Fail instead, and say what to do.
	// SELECT…FROM, not bare FROM: prose says "answered from the code_symbols
	// index" and would otherwise trip this on every sentence.
	quoted := regexp.MustCompile(`(?i)"[^"\n]*\bSELECT\b[^"\n]*\bFROM\s+[a-z_][a-z0-9_]*`)
	if m := quoted.FindString(stripBacktickLiterals(code)); m != "" {
		t.Errorf("SQL appears to be written in a double-quoted string, which this scan cannot read: %q\n"+
			"Move it to a backtick literal (as every other query here is) or the always-list guard is blind to it.", m)
	}

	always := map[string]bool{}
	for _, tbl := range schemaAlwaysTables {
		always[tbl] = true
	}
	var missing []string
	for tbl := range found {
		if !always[tbl] {
			missing = append(missing, tbl)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("these tables are queried by diagnose_load_runtime but are NOT in schemaAlwaysTables: %v\n"+
			"The bundle will render their rows while the Schema section hides their columns — the exact\n"+
			"shape that made run 074beb8a guess a column name and stop at UNVERIFIABLE. Add them to\n"+
			"schemaAlwaysTables, or to notTables here if the match is not a real table.", missing)
	}
}

// TestSchemaAlwaysTablesIsDeterministic guards the properties the always-list
// needs to be usable as evidence: sorted (so two runs render the section in the
// same order and can be diffed), no duplicates, and non-empty. An unsorted list
// would still work but makes two bundles of the same site gratuitously differ.
func TestSchemaAlwaysTablesIsDeterministic(t *testing.T) {
	if len(schemaAlwaysTables) == 0 {
		t.Fatal("schemaAlwaysTables is empty — the include filter alone decides the section again")
	}
	if !sort.StringsAreSorted(schemaAlwaysTables) {
		t.Errorf("schemaAlwaysTables is not sorted: %v", schemaAlwaysTables)
	}
	seen := map[string]bool{}
	for _, tbl := range schemaAlwaysTables {
		if seen[tbl] {
			t.Errorf("duplicate entry %q", tbl)
		}
		seen[tbl] = true
	}
	// The regression case itself, named explicitly: this entry is the whole
	// reason the mechanism exists, so it gets its own assertion rather than
	// relying on the derivation above to keep covering it.
	if !seen["orchestration_states"] {
		t.Error("orchestration_states missing — this is the table run 074beb8a could not address")
	}
}

// TestInputSpecDefaultMatchesAlwaysTables proves the spec default is the SAME
// list, not a copy that can drift. It is derived via stringsAsIface, so this
// asserts the derivation actually happened (a retyped literal would pass a
// length check but fail here once either side changed).
func TestInputSpecDefaultMatchesAlwaysTables(t *testing.T) {
	raw, ok := DiagnoseLoadRuntimeInputSpec.Defaults["schema_always_tables"]
	if !ok {
		t.Fatal("schema_always_tables missing from InputSpec Defaults — the action's contract does not carry it")
	}
	arr, ok := raw.([]interface{})
	if !ok {
		t.Fatalf("default is %T, want []interface{} (the shape config JSON arrives as)", raw)
	}
	if len(arr) != len(schemaAlwaysTables) {
		t.Fatalf("default has %d entries, schemaAlwaysTables has %d", len(arr), len(schemaAlwaysTables))
	}
	for i, v := range arr {
		if s, _ := v.(string); s != schemaAlwaysTables[i] {
			t.Errorf("entry %d: default %v, var %q", i, v, schemaAlwaysTables[i])
		}
	}
}

func TestSchemaFilterNotice(t *testing.T) {
	tests := []struct {
		name        string
		shown, tot  int
		full        bool
		wantEmpty   bool
		wantContain []string
	}{
		{
			name:  "filtered listing states coverage and how to reach an unlisted table",
			shown: 31, tot: 433,
			wantContain: []string{"31 of 433", "FILTERED", "NOT evidence it does not exist",
				"information_schema.columns", "you do not need a human"},
		},
		// schema_full asked for everything, so nothing was withheld and a notice
		// would be a false claim about the listing.
		{name: "schema_full suppresses", shown: 433, tot: 433, full: true, wantEmpty: true},
		{name: "nothing withheld suppresses", shown: 433, tot: 433, wantEmpty: true},
		// total=0 is the failed-count path in gatherSchema: degrade to silence
		// rather than print "31 of 0", which is worse than saying nothing.
		{name: "unknown total suppresses", shown: 31, tot: 0, wantEmpty: true},
		{name: "empty listing suppresses", shown: 0, tot: 433, wantEmpty: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := schemaFilterNotice(tc.shown, tc.tot, tc.full)
			if tc.wantEmpty {
				if got != "" {
					t.Errorf("want empty, got %q", got)
				}
				return
			}
			for _, want := range tc.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("notice missing %q\ngot: %s", want, got)
				}
			}
		})
	}
}

// TestGatherSchemaAlwaysListSurvivesTheFilters is the behavioural half: the
// always-list must beat BOTH the relevance include and the exclude denylist, and
// must sort first so tableCap truncation cannot reach it. Asserted against the
// SQL actually issued, because that is where the invariant lives — a unit test
// of the Go filtering alone would have passed on the broken version too.
func TestGatherSchemaAlwaysListSurvivesTheFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// The expectation IS the assertion: the query must order always-tables first
	// (so tableCap truncation cannot reach them), and must BIND the always-list as
	// a parameter rather than interpolate it. A mismatch on either fails the call.
	mock.ExpectQuery(`ORDER BY \(table_name = ANY\(\$3::text\[\]\)\) DESC`).
		WithArgs("%orchestration%", "site%", `{"orchestration_states"}`).
		WillReturnRows(sqlmock.NewRows([]string{"table_name", "column_name", "data_type"}).
			AddRow("orchestration_states", "orchestration_id", "uuid").
			AddRow("sites", "id", "uuid"))
	mock.ExpectQuery("count").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(433))

	out, err := gatherSchema(context.Background(), db,
		[]string{"%orchestration%"}, // denylist that WOULD drop it
		[]string{"site%"},           // include that does NOT match it
		[]string{"orchestration_states"},
		false, 120)
	if err != nil {
		t.Fatalf("gatherSchema: %v", err)
	}
	if !strings.Contains(out, "orchestration_states(orchestration_id uuid)") {
		t.Errorf("always-table absent from listing:\n%s", out)
	}
	if !strings.Contains(out, "2 of 433") {
		t.Errorf("notice missing or miscounted:\n%s", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

// stripGoComments removes // and /* */ comments so a source scan reads CODE
// only. Without it this file's own prose satisfies the scan.
func stripGoComments(src string) string {
	var b strings.Builder
	for i := 0; i < len(src); i++ {
		switch {
		case i+1 < len(src) && src[i] == '/' && src[i+1] == '/':
			for i < len(src) && src[i] != '\n' {
				i++
			}
			b.WriteByte('\n')
		case i+1 < len(src) && src[i] == '/' && src[i+1] == '*':
			i += 2
			for i+1 < len(src) && !(src[i] == '*' && src[i+1] == '/') {
				i++
			}
			i++
		default:
			b.WriteByte(src[i])
		}
	}
	return b.String()
}

// stripBacktickLiterals removes raw string literals, leaving the code that the
// SQL scan did NOT read — so the blind-spot check above looks only at what it
// could have missed.
func stripBacktickLiterals(src string) string {
	return regexp.MustCompile("(?s)`[^`]*`").ReplaceAllString(src, "``")
}

func TestStripGoComments(t *testing.T) {
	in := "a := 1 // FROM commented_table\n/* FROM block_table */\nrows := db.Query(`FROM real_table`)"
	got := stripGoComments(in)
	for _, bad := range []string{"commented_table", "block_table"} {
		if strings.Contains(got, bad) {
			t.Errorf("comment survived stripping: %q in %q", bad, got)
		}
	}
	if !strings.Contains(got, "real_table") {
		t.Errorf("code was stripped: %q", got)
	}
}
