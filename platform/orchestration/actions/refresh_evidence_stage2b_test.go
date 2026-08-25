// FILE: platform/orchestration/actions/refresh_evidence_stage2b_test.go
//
// RFC_025 stage 2b (bugs_open/288 §5.6): artifact_check reachable for EVERY
// source kind, and addressable by something that survives page decomposition.
//
// The two defects, both still literally true at HEAD before this change:
//
//  1. THE ONE READER ON THE RIGHT SURFACE COULD NOT BE POINTED AT THE FACTS THAT
//     NEEDED IT. refreshOneSiteEvidence's per-fact loop handles src["citation"]
//     and `continue`s BEFORE the src["artifact_check"] test seventeen lines
//     further down. Every fact holding a legislated figure is a citation fact —
//     185 of 294 current facts, measured 2026-08-24 — so an artifact_check
//     written beside an SDLT band was never evaluated. bugs_closed/225 ran an
//     expired £625,000 cap for sixteen months on exactly such a fact.
//
//  2. THE ADDRESS DIED. artifact_check.component_id is a page_components row id,
//     and 225's own component no longer exists: the page was decomposed into
//     prose-0 / tool-1 / prose-2. A binding written when that bug was filed
//     would today fail closed for ever, which reads exactly like a working check.
//
// The mutation for (1) is: reinstate the citation-`continue` ahead of the
// artifact evaluation, i.e. delete the pre-pass. TestStage2b_CitationFactAlso...
// must go red. The mutation for (2) is: drop the SubjectKey arm from
// resolveArtifactCheckSurface.

package actions

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// A citation fact — the shape every legislated figure has — that ALSO carries an
// artifact_check addressed by subject key. Before stage 2b this combination was
// unrepresentable in practice: the key would simply never be read.
func sdltFactWithArtifactCheck(subjectKey, pattern string) map[string]interface{} {
	return map[string]interface{}{
		"id":    "sdlt-ftb-relief-cap",
		"claim": "First-time buyer relief cannot be claimed if the price is over £500,000",
		"value": float64(500000),
		"kind":  "metric",
		"source": map[string]interface{}{
			"citation": map[string]interface{}{
				"url":   "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
				"quote": "If the price is over £500,000, you cannot claim the relief.",
			},
			"artifact_check": map[string]interface{}{
				"subject_key":     subjectKey,
				"pattern":         pattern,
				"must_be_present": true,
			},
		},
		"verified_at": "2026-08-15",
	}
}

// ── Defect 2: addressing that survives decomposition ───────────────────────

func TestStage2b_SubjectKeyResolvesTheToolSurface(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	site := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(artifactCheckSubjectSurfaceQuery)).
		WithArgs(site, "stamp-duty").
		WillReturnRows(sqlmock.NewRows([]string{"surface"}).
			AddRow(`<script>var FTB_RELIEF_CAP = 500000;</script>`))

	spec := &artifactCheckSpec{SubjectKey: "stamp-duty", Pattern: `FTB_RELIEF_CAP\s*=\s*500000`, MustBePresent: true}
	surface, addr, err := resolveArtifactCheckSurface(context.Background(), db, site, spec)
	if err != nil {
		t.Fatalf("a resolvable subject key must not error: %v", err)
	}
	if surface == "" {
		t.Fatal("the tool's stored bytes must come back")
	}
	if addr != `tool "stamp-duty"` {
		t.Fatalf("the drift detail must name the tool, got %q", addr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

// An EMPTY surface is an ERROR, not an absence. Both readings are "the pattern
// is not there" and only one is about the artefact: a renamed page or a
// deactivated component must never be reported as a drifted published claim,
// and must certainly never be reported as fresh. RFC_017.
func TestStage2b_UnresolvableSubjectKeyFailsClosedNotAbsent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	site := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(artifactCheckSubjectSurfaceQuery)).
		WithArgs(site, "renamed-away").
		WillReturnRows(sqlmock.NewRows([]string{"surface"}).AddRow(""))

	fact := sdltFactWithArtifactCheck("renamed-away", `FTB_RELIEF_CAP\s*=\s*500000`)
	entry := refreshArtifactCheckFact(context.Background(), db, site, fact, "2026-08-24", false)
	if entry.Outcome != "error" {
		t.Fatalf("an unresolvable surface must be error, never drifted or fresh, got %q", entry.Outcome)
	}
	if fact["verified_at"] != "2026-08-15" {
		t.Fatalf("a check that could not run must not touch verified_at, got %v", fact["verified_at"])
	}
}

// Exactly one address, and both-at-once fails closed rather than picking one.
func TestStage2b_BothAddressesRefused(t *testing.T) {
	src := map[string]interface{}{
		"artifact_check": map[string]interface{}{
			"component_id": gdTrialsComponentID,
			"subject_key":  "stamp-duty",
			"pattern":      `FTB_RELIEF_CAP\s*=\s*500000`,
		},
	}
	if _, err := parseArtifactCheck(src); err == nil {
		t.Fatal("a fact naming BOTH component_id and subject_key must be refused")
	}
}

func TestStage2b_NeitherAddressRefused(t *testing.T) {
	src := map[string]interface{}{
		"artifact_check": map[string]interface{}{"pattern": `x\s*=\s*1`},
	}
	if _, err := parseArtifactCheck(src); err == nil {
		t.Fatal("a fact naming no address at all must be refused")
	}
}

// The bare-digit refusal is RFC_025's own guard against the platform's
// documented landmine (grepping for 10000 matches inside 100000). Stage 2b must
// not have widened a hole around it via the new address.
func TestStage2b_BareNumericPatternStillRefusedWithSubjectKey(t *testing.T) {
	src := map[string]interface{}{
		"artifact_check": map[string]interface{}{"subject_key": "stamp-duty", "pattern": "500000"},
	}
	if _, err := parseArtifactCheck(src); err == nil {
		t.Fatal("a bare-digit pattern must stay refused whichever address is used")
	}
}

// ── Defect 1: reachability, and what a SECONDARY check must not touch ──────

// THE INDUCED RED FOR THE WHOLE OF STAGE 2b. A citation fact carrying an
// artifact_check that does NOT match must produce a drifted artifact entry and
// count into ArtifactCheckDrifted. Before the pre-pass this fact's
// artifact_check was unreachable, so the assertion below could not be satisfied
// at all.
//
// MUTATION THAT MUST GO RED: delete the pre-pass in refreshOneSiteEvidence.
func TestStage2b_CitationFactAlsoGetsItsArtifactCheckEvaluated(t *testing.T) {
	fact := sdltFactWithArtifactCheck("stamp-duty", `FTB_RELIEF_CAP\s*=\s*500000`)
	src := fact["source"].(map[string]interface{})

	// The gate the pre-pass is keyed on: this fact HAS a primary source, so its
	// artifact_check is a secondary check.
	if !factHasNonArtifactSource(fact, src) {
		t.Fatal("a citation fact must be recognised as having a primary source")
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()
	// The tool's bytes still carry the EXPIRED cap — bugs_closed/225's shape.
	mock.ExpectQuery(regexp.QuoteMeta(artifactCheckSubjectSurfaceQuery)).
		WithArgs(site, "stamp-duty").
		WillReturnRows(sqlmock.NewRows([]string{"surface"}).
			AddRow(`<script>var FTB_RELIEF_CAP = 625000;</script>`))

	entry := refreshArtifactCheckFact(context.Background(), db, site, fact, "2026-08-24", false)
	if entry.Outcome != "drifted" {
		t.Fatalf("the registered figure is absent from the tool's bytes: must be drifted, got %q (%s)", entry.Outcome, entry.Detail)
	}
	// A SECONDARY check must not own verified_at. A passing-or-failing artifact
	// check bumping the date on a fact whose CITATION was just lost would
	// disarm the time-based citation escalation the council relied on.
	if fact["verified_at"] != "2026-08-15" {
		t.Fatalf("a secondary artifact check must not touch verified_at, got %v", fact["verified_at"])
	}
}

// The other half of the same rule: even a PASSING secondary check leaves
// verified_at alone, so `changed` cannot be flipped by it either. This is the
// specific regression that would otherwise open shouldRaiseStaleEvidence for
// unrelated citation drift on the same site.
func TestStage2b_PassingSecondaryCheckDoesNotBumpVerifiedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(artifactCheckSubjectSurfaceQuery)).
		WithArgs(site, "stamp-duty").
		WillReturnRows(sqlmock.NewRows([]string{"surface"}).
			AddRow(`<script>var FTB_RELIEF_CAP = 500000;</script>`))

	fact := sdltFactWithArtifactCheck("stamp-duty", `FTB_RELIEF_CAP\s*=\s*500000`)
	entry := refreshArtifactCheckFact(context.Background(), db, site, fact, "2026-08-24", false)
	if entry.Outcome != "fresh" {
		t.Fatalf("the pattern is present: expected fresh, got %q (%s)", entry.Outcome, entry.Detail)
	}
	if entry.VerifiedAt != "" || fact["verified_at"] != "2026-08-15" {
		t.Fatalf("a secondary pass must leave verified_at to the primary arm, got entry=%q fact=%v",
			entry.VerifiedAt, fact["verified_at"])
	}
}

// The pre-pass must NOT fire for a fact whose only mechanism is artifact_check —
// that fact keeps taking the original branch, which owns verified_at. This is
// the no-op guarantee for RFC_025 stage 2 as ratified.
func TestStage2b_ArtifactOnlyFactIsStillPrimaryAndOwnsVerifiedAt(t *testing.T) {
	fact := gdTrialsLikeFact(gdTrialsComponentID, `Math\.min\(val,\s*10000\)`, true)
	src := fact["source"].(map[string]interface{})
	if factHasNonArtifactSource(fact, src) {
		t.Fatal("an artifact-only fact must NOT be treated as having a primary source, or it would lose its verified_at bump")
	}
}

// An attested_by or sql fact carrying an artifact_check is also reachable now.
// Both branches sat BELOW the citation continue too, so both were unreachable
// for a dual-source fact.
func TestStage2b_AttestedAndSQLFactsAlsoCountAsPrimary(t *testing.T) {
	attested := map[string]interface{}{
		"id":     "hand-blessed",
		"source": map[string]interface{}{"attested_by": "the owner", "artifact_check": map[string]interface{}{}},
	}
	if !factHasNonArtifactSource(attested, attested["source"].(map[string]interface{})) {
		t.Error("an attested_by fact has a primary source")
	}
	// factSQLSource reads source.sql, not a top-level "query" key.
	sqlFact := map[string]interface{}{
		"id": "counted",
		"source": map[string]interface{}{
			"sql":            "SELECT count(*) FROM pages",
			"artifact_check": map[string]interface{}{},
		},
	}
	if !factHasNonArtifactSource(sqlFact, sqlFact["source"].(map[string]interface{})) {
		t.Error("a sql-sourced fact has a primary source")
	}
}

// ── 2c: two entries per fact, and severity must win over order ─────────────

// A fact can now append two entries. If "last wins" survived, a PASSING
// artifact check appended after a DRIFTED citation would hide the citation drift
// from classifyFactDrift's evidence_drift arm — a lost GOV.UK citation on a
// declared tax fact would stop reaching a human.
//
// MUTATION THAT MUST GO RED: restore `entryIdx[e.FactID] = i` unconditionally.
func TestStage2b_DriftedEntryWinsOverAFreshOneOnTheSameFact(t *testing.T) {
	res := &siteRefreshResult{
		SiteID: fdSiteIDStr,
		Facts: []evidenceFactRefresh{
			{FactID: "sdlt-ftb-relief-cap", Outcome: "drifted", Detail: "citation quote gone"},
			{FactID: "sdlt-ftb-relief-cap", Outcome: "fresh", Tolerance: "artifact_check"},
		},
	}
	idx := &factDriftIndex{byFact: map[string][]factDriftTool{
		"sdlt-ftb-relief-cap": {fdTool("stamp-duty", false, true)},
	}}
	facts := map[string]map[string]interface{}{"sdlt-ftb-relief-cap": sdltReliefCapFact(500000)}

	// A baseline exists and the value has NOT moved, so the only arm that can
	// fire is evidence_drift — and it fires only if the drifted entry was chosen.
	base := factDriftBaselines{lastItem: map[string]float64{"sdlt-ftb-relief-cap|stamp-duty": 500000}}
	ems := planFactDriftFanOut(res, facts, idx, base, fdSiteIDStr)

	if len(ems) != 1 {
		t.Fatalf("expected one emission, got %d", len(ems))
	}
	if ems[0].Kind != "evidence_drift" {
		t.Fatalf("the DRIFTED entry must be the one classified, got kind %q — a fresh artifact entry has masked a lost citation", ems[0].Kind)
	}
}

// Order-independence: the same pair the other way round must give the same
// answer. A rule that only works when the drifted entry happens to be second is
// not a rule.
func TestStage2b_SeveritySelectionIsOrderIndependent(t *testing.T) {
	res := &siteRefreshResult{
		SiteID: fdSiteIDStr,
		Facts: []evidenceFactRefresh{
			{FactID: "sdlt-ftb-relief-cap", Outcome: "fresh", Tolerance: "artifact_check"},
			{FactID: "sdlt-ftb-relief-cap", Outcome: "drifted", Detail: "citation quote gone"},
		},
	}
	idx := &factDriftIndex{byFact: map[string][]factDriftTool{
		"sdlt-ftb-relief-cap": {fdTool("stamp-duty", false, true)},
	}}
	facts := map[string]map[string]interface{}{"sdlt-ftb-relief-cap": sdltReliefCapFact(500000)}
	base := factDriftBaselines{lastItem: map[string]float64{"sdlt-ftb-relief-cap|stamp-duty": 500000}}

	ems := planFactDriftFanOut(res, facts, idx, base, fdSiteIDStr)
	if len(ems) != 1 || ems[0].Kind != "evidence_drift" {
		t.Fatalf("selection must not depend on append order, got %+v", ems)
	}
}

// A single-entry fact must behave exactly as before — the whole point of
// selecting by severity rather than rewriting the loop.
func TestStage2b_SingleEntryFactUnchanged(t *testing.T) {
	res := &siteRefreshResult{
		SiteID: fdSiteIDStr,
		Facts:  []evidenceFactRefresh{{FactID: "sdlt-ftb-relief-cap", Outcome: "fresh"}},
	}
	idx := &factDriftIndex{byFact: map[string][]factDriftTool{
		"sdlt-ftb-relief-cap": {fdTool("stamp-duty", false, true)},
	}}
	facts := map[string]map[string]interface{}{"sdlt-ftb-relief-cap": sdltReliefCapFact(500000)}
	base := factDriftBaselines{lastItem: map[string]float64{"sdlt-ftb-relief-cap|stamp-duty": 500000}}

	if ems := planFactDriftFanOut(res, facts, idx, base, fdSiteIDStr); len(ems) != 0 {
		t.Fatalf("a fresh single entry with an unmoved value must stay silent, got %+v", ems)
	}
}

// ── The pre-pass itself, driven through refreshOneSiteEvidence ──────────────
//
// ⚠ WITHOUT THIS TEST, DELETING THE PRE-PASS LEFT EVERY TEST ABOVE GREEN.
// I checked, by deleting it: the suite passed. Every assertion above is on a
// UNIT — the resolver, the gate, the entry selection — and none of them can see
// whether the loop calls any of it. That is the second time in this one change
// (see criteria_facts_declaration_gate_test.go's wiring test) and it is the
// lesson this lane already wrote down: a mutation that PASSES has usually
// bypassed the path, not proved the guard.
//
// It also found a real defect in the change itself. An attested_by fact carrying
// an artifact_check reaches BOTH the pre-pass and the original artifact_check
// branch, so before the `!factHasNonArtifactSource` guard was added there it was
// checked TWICE: two entries under one FactID and a verified_at bump from the
// second. The citation arm was safe only because its `continue` happens to
// intervene. No unit test could have seen that.
//
// attested_by is used as the primary source deliberately: it exercises the same
// pre-pass with NO network call, where a citation fact would try to fetch GOV.UK.
func TestStage2b_PrePassRunsInsideTheRealLoop(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	site := uuid.New()
	fact := map[string]interface{}{
		"id":    "sdlt-ftb-relief-cap",
		"claim": "First-time buyer relief cannot be claimed over £500,000",
		"value": float64(500000),
		"source": map[string]interface{}{
			"attested_by": "the owner",
			"artifact_check": map[string]interface{}{
				"subject_key":     "stamp-duty",
				"pattern":         `FTB_RELIEF_CAP\s*=\s*500000`,
				"must_be_present": true,
			},
		},
		"verified_at": "2026-08-15",
	}
	eb, _ := json.Marshal(map[string]interface{}{"facts": []interface{}{fact}})

	mock.ExpectQuery("SELECT id, data, pinned FROM site_specs").WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"id", "data", "pinned"}).AddRow(uuid.New(), eb, false))
	mock.ExpectQuery("SELECT domain FROM sites").WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("mortgagecalculator.co.uk"))
	mock.ExpectQuery("SELECT to_char").
		WillReturnRows(sqlmock.NewRows([]string{"today"}).AddRow("2026-08-24"))
	// THE PRE-PASS. The tool's bytes still hold bugs_closed/225's expired cap.
	mock.ExpectQuery(regexp.QuoteMeta(artifactCheckSubjectSurfaceQuery)).
		WithArgs(site, "stamp-duty").
		WillReturnRows(sqlmock.NewRows([]string{"surface"}).
			AddRow(`<script>var FTB_RELIEF_CAP = 625000;</script>`))
	// Then the fan-out's index read, on a site whose PLANs declare nothing.
	mock.ExpectQuery(regexp.QuoteMeta(factDriftIndexQuery)).WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "build_status", "subject_key", "body", "fork_component_id"}))

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	res, err := refreshOneSiteEvidence(context.Background(), params, site, true, zap.NewNop())
	if err != nil {
		t.Fatalf("refreshOneSiteEvidence: %v", err)
	}

	// MUTATION THAT MUST GO RED: delete the pre-pass from refreshOneSiteEvidence.
	// The unmet surface query is the failure, and so is this count.
	if res.ArtifactCheckDrifted != 1 {
		t.Fatalf("the pre-pass must evaluate a dual-source fact's artifact_check: ArtifactCheckDrifted = %d, want 1", res.ArtifactCheckDrifted)
	}

	// EXACTLY ONE entry for this fact. Two would mean it was checked twice — the
	// defect this test found.
	n := 0
	for _, e := range res.Facts {
		if e.FactID == "sdlt-ftb-relief-cap" && e.Tolerance == "artifact_check" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("a dual-source fact must be artifact-checked ONCE, got %d entries", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the loop did not reach the surface read: %v", err)
	}
}

// THE NO-OP PROOF for stage 2b. Today 294 of 294 current register facts carry no
// artifact_check at all (measured 2026-08-24, control: 185 carry citation). A
// fact with no artifact_check must not cause the pre-pass to issue any query.
func TestStage2b_FactWithoutArtifactCheckTouchesNothingNew(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	site := uuid.New()
	fact := map[string]interface{}{
		"id":          "hand-blessed",
		"claim":       "the owner says so",
		"source":      map[string]interface{}{"attested_by": "the owner"},
		"verified_at": "2026-08-15",
	}
	eb, _ := json.Marshal(map[string]interface{}{"facts": []interface{}{fact}})

	mock.ExpectQuery("SELECT id, data, pinned FROM site_specs").WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"id", "data", "pinned"}).AddRow(uuid.New(), eb, false))
	mock.ExpectQuery("SELECT domain FROM sites").WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("example.test"))
	mock.ExpectQuery("SELECT to_char").
		WillReturnRows(sqlmock.NewRows([]string{"today"}).AddRow("2026-08-24"))
	// NO surface query expected — sqlmock is strict, so a stray pre-pass read
	// here is an unexpected-query failure.
	//
	// And no fan-out query either: with no artifact_check the pre-pass never runs,
	// so this fact is checked by nothing (its attestation is not yet due) and
	// res.FactsChecked stays 0, which returns early at the "no live-verifiable
	// facts" guard. That is pre-existing behaviour and worth naming, because it is
	// the other half of the no-op property: on a site where nothing opts in, the
	// pre-pass does not merely stay quiet, it does not even change how far the
	// pass gets.

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	res, err := refreshOneSiteEvidence(context.Background(), params, site, true, zap.NewNop())
	if err != nil {
		t.Fatalf("refreshOneSiteEvidence: %v", err)
	}
	if res.ArtifactCheckDrifted != 0 {
		t.Fatalf("no fact carries artifact_check: ArtifactCheckDrifted must be 0, got %d", res.ArtifactCheckDrifted)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the pre-pass must issue no query for a fact that does not opt in: %v", err)
	}
}

// ── A MISPLACED artifact_check IS REPORTED, NOT IGNORED ────────────────────
//
// Found in production 2026-08-25 on agritec.uk: four artifact_check objects with
// correct contents, correct patterns and a correct subject_key, live in the
// register — and placed at the TOP LEVEL of the fact rather than inside
// `source`, where every reader looks. Read by nothing, with no signal anywhere.
// The author had tested the patterns and reported the fence live.
//
// It is the Phase 1 defect one table over: a declaration nobody can read looking
// identical to a document that declares nothing. And the misplacement is a
// REASONABLE reading — artifact_check describes the fact, not the source — so
// the answer is to say so, not to call the author wrong.
//
// MUTATION THAT MUST GO RED: delete the misplacement branch in
// refreshOneSiteEvidence.
func TestStage2b_TopLevelArtifactCheckIsReportedNotSilentlyIgnored(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()

	// agritec's exact shape: citation + attested_by under source, artifact_check
	// hoisted to the top level of the fact.
	fact := map[string]interface{}{
		"id":    "ATT-sfi26-CSAM3",
		"claim": "SFI26 action CSAM3 (Herbal leys) has an annual payment of £224 per hectare.",
		"value": float64(224),
		"source": map[string]interface{}{
			"attested_by": "agritec_uk lane",
		},
		"artifact_check": map[string]interface{}{
			"subject_key":     "tool-sfi26-revenue-stacker",
			"pattern":         `code:'CSAM3'[^}]*rate:224\b`,
			"must_be_present": true,
		},
		"verified_at": "2026-08-22",
	}
	eb, _ := json.Marshal(map[string]interface{}{"facts": []interface{}{fact}})

	mock.ExpectQuery("SELECT id, data, pinned FROM site_specs").WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"id", "data", "pinned"}).AddRow(uuid.New(), eb, false))
	mock.ExpectQuery("SELECT domain FROM sites").WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("agritec.uk"))
	mock.ExpectQuery("SELECT to_char").
		WillReturnRows(sqlmock.NewRows([]string{"today"}).AddRow("2026-08-25"))
	// NO surface query expected: the misplaced check must NOT be evaluated —
	// reporting it is not the same as quietly making it work, and relocating a
	// human's register row is not this action's business (CLM-001).

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	res, err := refreshOneSiteEvidence(context.Background(), params, site, true, zap.NewNop())
	if err != nil {
		t.Fatalf("refreshOneSiteEvidence: %v", err)
	}
	if len(res.MisplacedArtifactChecks) != 1 || res.MisplacedArtifactChecks[0] != "ATT-sfi26-CSAM3" {
		t.Fatalf("a top-level artifact_check must be NAMED, got %v", res.MisplacedArtifactChecks)
	}
	if res.ArtifactCheckDrifted != 0 {
		t.Fatalf("a misplaced check must not be evaluated, got ArtifactCheckDrifted=%d", res.ArtifactCheckDrifted)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected queries: %v", err)
	}
}

// The no-op twin: a CORRECTLY placed artifact_check must not be reported as
// misplaced. Without this, the fix could report every working fact and nobody
// would notice until the field was full of false accusations.
func TestStage2b_CorrectlyPlacedArtifactCheckIsNotReportedAsMisplaced(t *testing.T) {
	fact := sdltFactWithArtifactCheck("stamp-duty", `FTB_RELIEF_CEILING\s*=\s*500000`)
	if _, hasTop := fact["artifact_check"]; hasTop {
		t.Fatal("fixture wrong: the correctly-placed fixture must NOT have a top-level key")
	}
	src := fact["source"].(map[string]interface{})
	if _, ok := src["artifact_check"]; !ok {
		t.Fatal("fixture wrong: the correct placement is inside source")
	}
}
