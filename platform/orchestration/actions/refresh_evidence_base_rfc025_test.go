// FILE: platform/orchestration/actions/refresh_evidence_base_rfc025_test.go
//
// RFC_025 (docs/agent_docs/docs024_key_docs_latest/architecture_review/
// RFC_025_artifact_sourced_facts_are_trusted_once_registered.md): stage 2's
// artifact_check, and stage 1's attestation staleness nudge.
//
// The induced-fault canary this file exists to prove (RFC_025 §5.3/§7, and
// bugs_open/161's own verification section): a fact shaped like the ACTUAL
// `gd-trials` case — claimed "10,000 Monte Carlo trials", cited "the figure is
// hard-coded in the shipped drop-rate tool JavaScript", artefact contains no
// randomness at all (bug file: `Math.random` count 0 in both tools; the real
// 10,000 is `Math.min(val, 10000)`, an input clamp) — must FLAG when
// artifact_check looks for the claimed technique and MUST NOT flag when it
// looks for what the artefact actually contains. The fixtures below are
// SYNTHETIC (a short, invented rendered_html snippet standing in for the real
// ~21.5KB component; the real component id is `b381f0db-...` per bugs_open/161
// — this file uses a made-up, well-formed UUID instead of the live one, since
// nothing here touches the real database).

package actions

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// gdTrialsLikeFact builds a fact shaped like the real, ALREADY-CORRECTED
// gd-trials fact (bugs_open/161's status block: claim retyped to "maximum
// attempts modelled per query", source now cites the real clamp), but with an
// artifact_check attached — which the real fact does not yet carry; RFC_025
// proposes attaching one, and this is what it would look like.
func gdTrialsLikeFact(componentID, pattern string, mustBePresent bool) map[string]interface{} {
	return map[string]interface{}{
		"id":    "gd-trials",
		"claim": "10,000 maximum attempts modelled per query",
		"value": float64(10000),
		"kind":  "metric",
		"source": map[string]interface{}{
			"artifact": "the figure is hard-coded in the shipped drop-rate tool JavaScript",
			"artifact_check": map[string]interface{}{
				"component_id":    componentID,
				"pattern":         pattern,
				"must_be_present": mustBePresent,
			},
		},
		"verified_at": "2026-07-24",
	}
}

// A synthetic stand-in for tool-drop-rate-simulator's real rendered_html
// (bugs_open/161): closed-form geometric computation, an input clamp at
// 10000, and — the decisive negative evidence the bug file could only
// establish from an untruncated export — NO Math.random anywhere.
const syntheticDropRateSimulatorHTML = `<script>
function computeSurvival(p, k) { return Math.pow(1 - p, k); }
function clampAttempts(val) { return Math.min(val, 10000); }
</script>`

const gdTrialsComponentID = "b381f0db-0000-4000-8000-000000000001" // synthetic; real id per bugs_open/161 truncated in the bug file itself

func TestArtifactCheck_MatchingPatternDoesNotFlag(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT rendered_html FROM page_components WHERE id = $1")).
		WithArgs(gdTrialsComponentID).
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).AddRow(syntheticDropRateSimulatorHTML))

	fact := gdTrialsLikeFact(gdTrialsComponentID, `Math\.min\(val,\s*10000\)`, true)
	entry := refreshArtifactCheckFact(context.Background(), db, fact, "2026-08-12")

	if entry.Outcome != "fresh" {
		t.Fatalf("a pattern that DOES match the real artefact must not flag: got outcome %q, detail %q", entry.Outcome, entry.Detail)
	}
	if entry.VerifiedAt != "2026-08-12" {
		t.Errorf("a fresh artifact_check must bump verified_at, got %q", entry.VerifiedAt)
	}
	if fact["verified_at"] != "2026-08-12" {
		t.Errorf("the fact map itself must be updated in place, got %v", fact["verified_at"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

// The motivating case, reproduced: this is what would have caught the ORIGINAL
// false "10,000 Monte Carlo trials" claim before it was corrected. Neither
// drop-rate tool contains any randomness (bugs_open/161, VERIFIED against the
// live rendered_html), so a check for the technique the claim asserted must
// drift, not pass.
func TestArtifactCheck_MismatchedPatternFlagsDrift(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT rendered_html FROM page_components WHERE id = $1")).
		WithArgs(gdTrialsComponentID).
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).AddRow(syntheticDropRateSimulatorHTML))

	// must_be_present:true for Math.random — exactly the assertion "this tool
	// uses randomness" that the original false claim implied and the real
	// artefact never supported.
	fact := gdTrialsLikeFact(gdTrialsComponentID, `Math\.random`, true)
	entry := refreshArtifactCheckFact(context.Background(), db, fact, "2026-08-12")

	if entry.Outcome != "drifted" {
		t.Fatalf("a pattern absent from the real artefact must flag as drifted: got outcome %q", entry.Outcome)
	}
	if entry.Detail == "" {
		t.Error("a drifted outcome must carry a human-readable detail")
	}
	if entry.VerifiedAt != "" {
		t.Error("a drifted fact must NOT be marked freshly verified")
	}
	if _, touched := fact["verified_at"]; fact["verified_at"] != "2026-07-24" || !touched {
		t.Errorf("a drifted fact's stored verified_at must be left for a human, got %v", fact["verified_at"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

// must_be_present:false is the mirror case — a claim that something is
// ABSENT (e.g. "this tool has no external network calls") must flag if the
// pattern is now found.
func TestArtifactCheck_MustBeAbsentButPresentFlagsDrift(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT rendered_html FROM page_components WHERE id = $1")).
		WithArgs(gdTrialsComponentID).
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).AddRow(syntheticDropRateSimulatorHTML))

	fact := gdTrialsLikeFact(gdTrialsComponentID, `Math\.min`, false) // asserts Math.min is ABSENT — it is not
	entry := refreshArtifactCheckFact(context.Background(), db, fact, "2026-08-12")

	if entry.Outcome != "drifted" {
		t.Fatalf("must_be_present:false with the pattern actually present must drift: got %q", entry.Outcome)
	}
}

// Resolution failure #1: the component_id does not resolve to any row — a
// deleted or mistyped component. RFC_017 (verifier registry fails CLOSED on
// error): a check that cannot run must never be reported as a pass.
func TestArtifactCheck_UnresolvedComponentFailsClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	missingID := "b381f0db-0000-4000-8000-0000000000ff"
	mock.ExpectQuery(regexp.QuoteMeta("SELECT rendered_html FROM page_components WHERE id = $1")).
		WithArgs(missingID).
		WillReturnError(sql.ErrNoRows)

	fact := gdTrialsLikeFact(missingID, `Math\.min\(val,\s*10000\)`, true)
	entry := refreshArtifactCheckFact(context.Background(), db, fact, "2026-08-12")

	if entry.Outcome != "error" {
		t.Fatalf("an unresolved component_id must fail CLOSED (RFC_017), got outcome %q — a check that could not run must never silently pass", entry.Outcome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

// Resolution failure #2: a malformed component_id (not even a well-formed
// id) must also fail closed, and — since it is caught before any query is
// issued — must never touch the database at all. Passing a nil *sql.DB
// proves that: a nil dereference would panic the test if the code tried.
func TestArtifactCheck_InvalidComponentIDFailsClosedWithoutTouchingDB(t *testing.T) {
	fact := gdTrialsLikeFact("not-a-uuid", `Math\.min\(val,\s*10000\)`, true)
	entry := refreshArtifactCheckFact(context.Background(), nil, fact, "2026-08-12")

	if entry.Outcome != "error" {
		t.Fatalf("a malformed component_id must fail CLOSED, got outcome %q", entry.Outcome)
	}
}

// Resolution failure #3: a pattern that does not compile as a regex must also
// fail closed, and — like the malformed id above — before any query, so a nil
// DB must not panic.
func TestArtifactCheck_InvalidRegexFailsClosedWithoutTouchingDB(t *testing.T) {
	fact := gdTrialsLikeFact(gdTrialsComponentID, `Math\.min(val,`, true) // unbalanced paren
	entry := refreshArtifactCheckFact(context.Background(), nil, fact, "2026-08-12")

	if entry.Outcome != "error" {
		t.Fatalf("a pattern that fails to compile must fail CLOSED, got outcome %q", entry.Outcome)
	}
}

// Resolution failure #4: artifact_check present but missing its required
// keys entirely — the shape parseArtifactCheck itself refuses.
func TestParseArtifactCheck_MissingFieldsRefused(t *testing.T) {
	cases := []map[string]interface{}{
		{"artifact_check": map[string]interface{}{"pattern": "x"}},      // no component_id
		{"artifact_check": map[string]interface{}{"component_id": "x"}}, // no pattern
		{"artifact_check": map[string]interface{}{}},                    // neither
		{"artifact_check": "not even an object"},                        // wrong shape entirely
	}
	for i, src := range cases {
		if _, err := parseArtifactCheck(src); err == nil {
			t.Errorf("case %d: expected parseArtifactCheck to refuse %+v, got no error", i, src)
		}
	}

	// The happy path, for contrast: all required fields present, and
	// must_be_present defaults to true when omitted (matching the RFC's own
	// worked example, which states it explicitly, but the field is documented
	// as optional).
	spec, err := parseArtifactCheck(map[string]interface{}{
		"artifact_check": map[string]interface{}{
			"component_id": gdTrialsComponentID,
			"pattern":      `Math\.min`,
		},
	})
	if err != nil {
		t.Fatalf("valid artifact_check refused: %v", err)
	}
	if !spec.MustBePresent {
		t.Error("must_be_present should default to true when omitted")
	}
}

// ============================================================================
// Stage 1 — attestation staleness nudge
// ============================================================================

func TestCheckAttestationStaleness(t *testing.T) {
	const today = "2026-08-12"

	cases := []struct {
		name       string
		verifiedAt string
		wantDue    bool
	}{
		{"recently dated, well inside the cadence", "2026-08-01", false},
		{"dated just under a year ago — well past 180 days", "2025-09-01", true},
		{"never dated at all — due immediately, silence is not freshness", "", true},
		{"unparseable verified_at treated the same as never dated", "some time last year", true},
	}
	for _, c := range cases {
		fact := map[string]interface{}{
			"id":    "x",
			"claim": "an attested claim",
			"source": map[string]interface{}{
				"attested_by": "owner, " + c.verifiedAt,
			},
			"verified_at": c.verifiedAt,
		}
		entry := checkAttestationStaleness(fact, today)
		gotDue := entry != nil
		if gotDue != c.wantDue {
			t.Errorf("%s: verifiedAt=%q due=%v, want %v (entry=%+v)", c.name, c.verifiedAt, gotDue, c.wantDue, entry)
			continue
		}
		if gotDue {
			if entry.Outcome != "attestation_due" {
				t.Errorf("%s: outcome = %q, want attestation_due", c.name, entry.Outcome)
			}
			// This is a NUDGE, never a re-attestation — it must not silently
			// bump verified_at, or the next pass would see the fact as fresh
			// again having actually checked nothing.
			if entry.VerifiedAt != "" {
				t.Errorf("%s: a staleness nudge must not set VerifiedAt, got %q", c.name, entry.VerifiedAt)
			}
			if fact["verified_at"] != c.verifiedAt {
				t.Errorf("%s: the stored fact's own verified_at must be untouched by a nudge, got %v", c.name, fact["verified_at"])
			}
		}
	}
}

// A fact with no artifact_check key must be completely unaffected by any of
// the above — this is RFC_025's unsafe-default-OFF requirement, checked at the
// unit level: refreshArtifactCheckFact is simply never reachable for such a
// fact (the caller in refreshOneSiteEvidence only calls it when the key is
// present), but parseArtifactCheck on its own must also refuse cleanly rather
// than inventing a check from nothing.
func TestParseArtifactCheck_AbsentKeyRefused(t *testing.T) {
	src := map[string]interface{}{"artifact": "some code path"} // no artifact_check at all
	if _, err := parseArtifactCheck(src); err == nil {
		t.Error("parseArtifactCheck must refuse a source with no artifact_check key")
	}
}
