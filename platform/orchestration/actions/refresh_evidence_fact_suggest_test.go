// FILE: platform/orchestration/actions/refresh_evidence_fact_suggest_test.go
//
// bugs_open/288 Phase 4 — adoption. The call-site test is written FIRST here,
// because three times in this change a guard's unit tests all passed while the
// call to it was gone or the fixture could not discriminate (WRONG_CALLS
// 2026-08-24). The mutations each test expects are named on it.

package actions

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func suggestEB(vals map[string]float64) map[string]interface{} {
	var facts []interface{}
	for id, v := range vals {
		facts = append(facts, map[string]interface{}{
			"id": id, "value": v, "claim": id, "kind": "metric",
		})
	}
	return map[string]interface{}{"facts": facts}
}

const suggestToolSurface = `
<section><p>Relief ends above &pound;500,000.</p></section>
<script>
  const FTB_RELIEF_CEILING = 500000;
  const SURCHARGE_FLOOR = 40000;
  var padding = 12;
</script>`

func TestFactSuggest_ProposesTheBindingsAlreadyVisibleInTheCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(factSuggestToolsQuery)).WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"subject_key", "name", "fragment"}).
			AddRow("stamp-duty", "tool-stamp-duty", suggestToolSurface))

	eb := suggestEB(map[string]float64{
		"sdlt-ftb-relief-cap":             500000, // in the code
		"sdlt-additional-surcharge-floor": 40000,  // in the code
		"sdlt-standard-top-rate":          12,     // BELOW the floor: must be refused, not "found"
		"sdlt-unrelated":                  777000, // not in the code
	})
	got := planFactBindingSuggestions(context.Background(), db, site, eb, map[string]bool{}, zap.NewNop())

	if len(got) != 1 {
		t.Fatalf("one tool should be suggested to, got %d", len(got))
	}
	ids := strings.Join(got[0].FactIDs, ",")
	if ids != "sdlt-additional-surcharge-floor,sdlt-ftb-relief-cap" {
		t.Fatalf("wrong bindings proposed: %s", ids)
	}
	// 12 is in the script twice over (`var padding = 12`) and must NOT be
	// proposed: measured false-positive rate at two digits is 3.79%.
	if strings.Contains(ids, "top-rate") {
		t.Error("a value below the measured distinctiveness floor must never be proposed")
	}
	if strings.Contains(ids, "unrelated") {
		t.Error("a value absent from the code must not be proposed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

// A tool that already declares is not a suggestion target — this is an adoption
// lever, not a re-audit. MUTATION: drop the declaring[subjectKey] skip.
func TestFactSuggest_SkipsAToolThatAlreadyDeclares(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(factSuggestToolsQuery)).WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"subject_key", "name", "fragment"}).
			AddRow("stamp-duty", "tool-stamp-duty", suggestToolSurface))

	got := planFactBindingSuggestions(context.Background(), db, site,
		suggestEB(map[string]float64{"sdlt-ftb-relief-cap": 500000}),
		map[string]bool{"stamp-duty": true}, zap.NewNop())
	if len(got) != 0 {
		t.Fatalf("a declaring tool must not be suggested to, got %+v", got)
	}
}

// PROSE MUST NOT PROPOSE A BINDING. The register's writer_line puts the figure
// in the copy, so a page whose only mention is prose would otherwise get a
// suggestion for a tool that does not encode it at all.
// MUTATION: probe `surface` instead of extractScriptText(surface).
func TestFactSuggest_ProseAloneProposesNothing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()
	proseOnly := `<p>Relief ends above &pound;500,000 (500000).</p><script>var a = 1;</script>`
	mock.ExpectQuery(regexp.QuoteMeta(factSuggestToolsQuery)).WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"subject_key", "name", "fragment"}).
			AddRow("stamp-duty", "tool-stamp-duty", proseOnly))

	got := planFactBindingSuggestions(context.Background(), db, site,
		suggestEB(map[string]float64{"sdlt-ftb-relief-cap": 500000}), map[string]bool{}, zap.NewNop())
	if len(got) != 0 {
		t.Fatalf("a figure that appears only in the COPY must not propose a code binding, got %+v", got)
	}
}

// The no-op population: a site with no register, or one whose values are all
// below the floor, must not even query.
func TestFactSuggest_NothingProbeableIssuesNoQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	for _, eb := range []map[string]interface{}{
		{},
		suggestEB(map[string]float64{"rate": 5, "pct": 12}),
	} {
		if got := planFactBindingSuggestions(context.Background(), db, uuid.New(), eb, map[string]bool{}, zap.NewNop()); got != nil {
			t.Fatalf("nothing probeable must yield nothing, got %+v", got)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("no query should have run: %v", err)
	}
}

func TestFactSuggest_BodyIsPasteReadyAndStatesItsOwnLimit(t *testing.T) {
	body := factBindingSuggestionBody(factBindingSuggestion{
		SubjectKey: "stamp-duty", PageName: "tool-stamp-duty",
		FactIDs: []string{"a", "b"}, Detail: []string{"a = 1", "b = 2"},
	})
	if !strings.Contains(body, `"facts": [`) || !strings.Contains(body, `"a",`) {
		t.Errorf("the note must carry a paste-ready declaration:\n%s", body)
	}
	for _, want := range []string{"co-occurrence, not role", "AGREE with the register", "never hand-edit"} {
		if !strings.Contains(body, want) {
			t.Errorf("the note must state its own limit (%q):\n%s", want, body)
		}
	}
}

// A dry run writes nothing — the induced-proof recipe depends on it.
func TestFactSuggest_DryRunWritesNothing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	n := writeFactBindingSuggestions(context.Background(), db, uuid.New(),
		[]factBindingSuggestion{{SubjectKey: "stamp-duty", FactIDs: []string{"a"}}}, true, zap.NewNop())
	if n != 0 {
		t.Fatalf("a dry run must write nothing, wrote %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("a dry run must not touch the database: %v", err)
	}
}

func TestFactSuggest_CooldownIsPerSubjectAndSite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()
	// (subject, site) both bound: one global PLAN resolves on many sites, so a
	// note about site A must not silence site B's different suggestion.
	mock.ExpectQuery("SELECT EXISTS").WithArgs("stamp-duty", site).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	n := writeFactBindingSuggestions(context.Background(), db, site,
		[]factBindingSuggestion{{SubjectKey: "stamp-duty", FactIDs: []string{"a"}}}, false, zap.NewNop())
	if n != 0 {
		t.Fatalf("a recent note must suppress the write, wrote %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

// ── The ambiguity guard, found MISSING by the first real production sweep ──
//
// 2026-08-25, agritec.uk: two facts (`CIT-3f1b219f15ec6a39`, `CIT-86c4010f7cdf820d`)
// both assert the SFI26 annual agreement cap of £100,000, and the suggester
// proposed BOTH for the single `100000` in the tool's script. A human cannot
// reconcile that binding — and `factProbeNotProbed`'s own comment already
// claimed "refused: … or ambiguous", for an arm that had never been written.
//
// The exact class this lane spent two days closing, committed by me while
// closing it: a comment promising a guard that does not exist.
//
// MUTATION THAT MUST GO RED: drop the `sharedValue[disp] > 1` branch.
func TestFactSuggest_TwoFactsSharingAValueAreReportedButNeverProposed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()

	// The real shape: two duplicate register facts at 100000, plus one clean fact.
	surface := `<script>const CAP = 100000; const FLOOR = 40000;</script>`
	mock.ExpectQuery(regexp.QuoteMeta(factSuggestToolsQuery)).WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"subject_key", "name", "fragment"}).
			AddRow("sfi26-stacker", "tool-sfi26-revenue-stacker", surface))

	got := planFactBindingSuggestions(context.Background(), db, site, suggestEB(map[string]float64{
		"CIT-duplicate-a": 100000,
		"CIT-duplicate-b": 100000,
		"clean-floor":     40000,
	}), map[string]bool{}, zap.NewNop())

	if len(got) != 1 {
		t.Fatalf("expected one suggestion, got %d", len(got))
	}
	s := got[0]
	// The clean one IS proposed.
	if len(s.FactIDs) != 1 || s.FactIDs[0] != "clean-floor" {
		t.Fatalf("only the unambiguous fact may be proposed, got %v", s.FactIDs)
	}
	// The colliding pair is REPORTED, not silently dropped — usually it means the
	// register has duplicates and the owner needs to see that.
	if len(s.Ambiguous) != 1 {
		t.Fatalf("the shared value must be reported once, got %v", s.Ambiguous)
	}
	for _, want := range []string{"CIT-duplicate-a", "CIT-duplicate-b", "cannot tell which"} {
		if !strings.Contains(s.Ambiguous[0], want) {
			t.Errorf("the ambiguity line must name both ids and say why: %s", s.Ambiguous[0])
		}
	}
	// And it must never reach the paste-ready fragment, because pasting it
	// declares two facts for one constant and every later pass owes a
	// reconciliation nobody can perform.
	body := factBindingSuggestionBody(s)
	if strings.Contains(body[strings.Index(body, `"facts": [`):], "CIT-duplicate") {
		t.Fatalf("an ambiguous id must not appear in the paste-ready declaration:\n%s", body)
	}
	if !strings.Contains(body, "AMBIGUOUS, NOT PROPOSED") {
		t.Errorf("the note must show the ambiguity to the reader:\n%s", body)
	}
}

// When EVERY match is ambiguous there is nothing to paste, and the note must not
// emit an empty `"facts": []` fragment that reads as "declare nothing".
func TestFactSuggest_AllAmbiguousEmitsNoPasteReadyFragment(t *testing.T) {
	s := factBindingSuggestion{
		SubjectKey: "sfi26-stacker", PageName: "tool-sfi26-revenue-stacker",
		Ambiguous: []string{"100,000 is carried by 2 facts on this site (a, b) — cannot tell which"},
	}
	body := factBindingSuggestionBody(s)
	if strings.Contains(body, `"facts": [`) {
		t.Fatalf("no paste-ready fragment when nothing is proposable:\n%s", body)
	}
	if !strings.Contains(body, "nothing is proposed") {
		t.Errorf("the note must say plainly that nothing is proposed:\n%s", body)
	}
}
