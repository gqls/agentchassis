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
	"github.com/google/uuid"
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

// gdTrialsSiteID is the synthetic site this component belongs to, for every
// test below except the cross-site scoping test, which deliberately queries
// with a DIFFERENT site id to prove the join refuses.
var gdTrialsSiteID = uuid.MustParse("22222222-0000-4000-8000-000000000001")

// artifactCheckQuery is the real query refreshArtifactCheckFact issues,
// site-scoped via a join through pages (2026-08-12 council fix: the bare
// `WHERE id = $1` this replaced let a fact "verify" against another site's
// component).
const artifactCheckQuery = `SELECT pc.rendered_html FROM page_components pc JOIN pages p ON p.id = pc.page_id WHERE pc.id = $1 AND p.site_id = $2`

func TestArtifactCheck_MatchingPatternDoesNotFlag(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(artifactCheckQuery)).
		WithArgs(gdTrialsComponentID, gdTrialsSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).AddRow(syntheticDropRateSimulatorHTML))

	fact := gdTrialsLikeFact(gdTrialsComponentID, `Math\.min\(val,\s*10000\)`, true)
	entry := refreshArtifactCheckFact(context.Background(), db, gdTrialsSiteID, fact, "2026-08-12")

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

	mock.ExpectQuery(regexp.QuoteMeta(artifactCheckQuery)).
		WithArgs(gdTrialsComponentID, gdTrialsSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).AddRow(syntheticDropRateSimulatorHTML))

	// must_be_present:true for Math.random — exactly the assertion "this tool
	// uses randomness" that the original false claim implied and the real
	// artefact never supported.
	fact := gdTrialsLikeFact(gdTrialsComponentID, `Math\.random`, true)
	entry := refreshArtifactCheckFact(context.Background(), db, gdTrialsSiteID, fact, "2026-08-12")

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

	mock.ExpectQuery(regexp.QuoteMeta(artifactCheckQuery)).
		WithArgs(gdTrialsComponentID, gdTrialsSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).AddRow(syntheticDropRateSimulatorHTML))

	fact := gdTrialsLikeFact(gdTrialsComponentID, `Math\.min`, false) // asserts Math.min is ABSENT — it is not
	entry := refreshArtifactCheckFact(context.Background(), db, gdTrialsSiteID, fact, "2026-08-12")

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
	mock.ExpectQuery(regexp.QuoteMeta(artifactCheckQuery)).
		WithArgs(missingID, gdTrialsSiteID).
		WillReturnError(sql.ErrNoRows)

	fact := gdTrialsLikeFact(missingID, `Math\.min\(val,\s*10000\)`, true)
	entry := refreshArtifactCheckFact(context.Background(), db, gdTrialsSiteID, fact, "2026-08-12")

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
	entry := refreshArtifactCheckFact(context.Background(), nil, gdTrialsSiteID, fact, "2026-08-12")

	if entry.Outcome != "error" {
		t.Fatalf("a malformed component_id must fail CLOSED, got outcome %q", entry.Outcome)
	}
}

// Resolution failure #3: a pattern that does not compile as a regex must also
// fail closed, and — like the malformed id above — before any query, so a nil
// DB must not panic.
func TestArtifactCheck_InvalidRegexFailsClosedWithoutTouchingDB(t *testing.T) {
	fact := gdTrialsLikeFact(gdTrialsComponentID, `Math\.min(val,`, true) // unbalanced paren
	entry := refreshArtifactCheckFact(context.Background(), nil, gdTrialsSiteID, fact, "2026-08-12")

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

// Council objection, 2026-08-12 (editquality, HIGH): a bare numeric pattern
// substring-matches a larger number — the platform's own documented landmine
// ("grepping a tool for 10000 matches 100000"), cited in bugs_open/161's own
// "Traps this cost me" section, the motivating bug for this whole mechanism.
// parseArtifactCheck must refuse a pattern that is nothing but digits.
func TestParseArtifactCheck_BareNumericPatternRefused(t *testing.T) {
	cases := []string{"10000", "0", "123456789"}
	for _, pattern := range cases {
		_, err := parseArtifactCheck(map[string]interface{}{
			"artifact_check": map[string]interface{}{
				"component_id": gdTrialsComponentID,
				"pattern":      pattern,
			},
		})
		if err == nil {
			t.Errorf("bare numeric pattern %q must be refused — it would substring-match a larger number", pattern)
		}
	}

	// The exact failure this guards against, made concrete: a bare "10000" DOES
	// substring-match inside "100000" — this is what the parse-time refusal
	// above stops from ever being evaluated against real component content.
	if !regexp.MustCompile(`10000`).MatchString("100000") {
		t.Fatal("test premise wrong: expected a bare 10000 pattern to substring-match 100000")
	}

	// Patterns WITH surrounding context — even minimal — must still be accepted;
	// the guard is narrow on purpose (RFC_025's own worked example uses exactly
	// this shape).
	okPatterns := []string{`\b10000\b`, `Math\.min\(val,\s*10000\)`, `10000px`, `-10000`}
	for _, pattern := range okPatterns {
		if _, err := parseArtifactCheck(map[string]interface{}{
			"artifact_check": map[string]interface{}{
				"component_id": gdTrialsComponentID,
				"pattern":      pattern,
			},
		}); err != nil {
			t.Errorf("pattern with real surrounding context %q should NOT be refused, got: %v", pattern, err)
		}
	}
}

// Council objection, 2026-08-12 (editquality/guardian/compliance/debug_
// historian — raised independently by four reviewers): component_id was
// resolved with no check that it belongs to the fact's own site. A fact in
// site A's register could "verify" against site B's component and report a
// false PASS — the exact failure mode RFC_025 exists to close, reproduced
// inside its own fix. The join must refuse (fail closed) rather than match.
func TestArtifactCheck_CrossSiteComponentRefusedNotSilentlyMatched(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	otherSiteID := uuid.MustParse("33333333-0000-4000-8000-000000000002")

	// The component genuinely exists and its content genuinely matches the
	// pattern — but it belongs to a DIFFERENT site than the one asking. The
	// join's WHERE clause means this query correctly returns no rows for the
	// wrong site, exactly as it would for a component that does not exist at
	// all — sqlmock proves the query is actually site-scoped, not just that
	// the Go code claims to be.
	mock.ExpectQuery(regexp.QuoteMeta(artifactCheckQuery)).
		WithArgs(gdTrialsComponentID, otherSiteID).
		WillReturnError(sql.ErrNoRows)

	fact := gdTrialsLikeFact(gdTrialsComponentID, `Math\.min\(val,\s*10000\)`, true)
	entry := refreshArtifactCheckFact(context.Background(), db, otherSiteID, fact, "2026-08-12")

	if entry.Outcome != "error" {
		t.Fatalf("a component belonging to a DIFFERENT site must fail CLOSED, not verify: got outcome %q", entry.Outcome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations (the query must actually filter on site_id, not just accept any id): %v", err)
	}
}

// ============================================================================
// The stale_evidence gating decision (council objection, 2026-08-12:
// bug_historian/debug_historian/architecture — verified only by manual code
// reading, no test on the gating logic itself; also the architecture seat's
// finding that decoupling the raise from `changed` for the PRE-EXISTING
// citation branch was outside RFC_025's ratified scope, since fixed by
// scoping the new behaviour to res.ArtifactCheckDrifted specifically).
// ============================================================================

func TestShouldRaiseStaleEvidence(t *testing.T) {
	cases := []struct {
		name                 string
		drifted              int
		changed              bool
		artifactCheckDrifted int
		want                 bool
	}{
		{"nothing drifted, register changed for an unrelated reason (e.g. writer_block regen) — no raise", 0, true, 0, false},
		{"sql fact drifted, changed=true (sql drift always sets changed) — raises, exactly as always", 1, true, 0, true},
		{"citation fact drifted ALONE, nothing else changed — must NOT raise: pre-existing behaviour, out of RFC_025's ratified scope to alter", 1, false, 0, false},
		{"artifact_check fact drifted ALONE, nothing else changed — MUST raise: this is exactly the new capability RFC_025 was ratified for", 1, false, 1, true},
		{"artifact_check drifted AND register also changed — raises (both conditions independently sufficient)", 1, true, 1, true},
	}
	for _, c := range cases {
		res := &siteRefreshResult{Drifted: c.drifted, ArtifactCheckDrifted: c.artifactCheckDrifted}
		got := shouldRaiseStaleEvidence(res, c.changed)
		if got != c.want {
			t.Errorf("%s: shouldRaiseStaleEvidence(Drifted=%d, changed=%v, ArtifactCheckDrifted=%d) = %v, want %v",
				c.name, c.drifted, c.changed, c.artifactCheckDrifted, got, c.want)
		}
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
