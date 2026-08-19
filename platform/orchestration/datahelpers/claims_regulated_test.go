package datahelpers

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// completeAttestation is the only shape that opens the gate. Every other test
// that needs a NON-attested site mutates one field of this, so the tests fail
// for the reason they name rather than because the fixture was never valid.
func completeAttestation() *RegulatedAttestation {
	return &RegulatedAttestation{
		FirmName:    "Example Mortgages Ltd",
		FRN:         "305432",
		Regulator:   "FCA",
		Permissions: "mortgage arranging and advising",
		AttestedBy:  "owner",
		AttestedAt:  "2026-08-19",
		Evidence:    "email 2026-08-19 with register screenshot; FRN checked on the FS Register",
	}
}

func scanWith(t *testing.T, eb *EvidenceBase, sentence string) []ClaimFinding {
	t.Helper()
	return ScanAllBannedClaims([]string{sentence}, eb)
}

// TestRegulatedFamilyIsWired guards the failure mode that let this whole gap sit
// unnoticed: a family that silently compiles to nothing looks exactly like one
// that works. Assert the count directly.
func TestRegulatedFamilyIsWired(t *testing.T) {
	if n := RegulatedClaimCount(); n == 0 {
		t.Fatalf("regulated family is empty — the guard is inert")
	}
	for i, b := range regulatedEvidence().BannedClaims {
		if b.re == nil {
			t.Errorf("pattern %d did not compile and has no fallback", i)
		}
		if !strings.Contains(b.Reason, "regulated attestation") {
			t.Errorf("pattern %d's reason does not tell the reader how to fix it: %q", i, b.Reason)
		}
	}
}

func TestRegulatedClaimsAreBlockedWithoutAttestation(t *testing.T) {
	mustCatch := []string{
		"We are authorised and regulated by the Financial Conduct Authority.",
		"We're FCA regulated and here to help.",
		"Our FRN is 305432.",
		"Our firm reference number is 305432.",
		"We are a credit broker, not a lender.",
		"We are an appointed representative of Example Ltd.",
		"We hold FCA authorisation for consumer credit.",
		"Our company is regulated by the FCA.",
	}
	for _, s := range mustCatch {
		if got := scanWith(t, nil, s); len(got) == 0 {
			t.Errorf("NOT BLOCKED but should be: %q", s)
		}
	}
}

// TestLegitimateRegulatorLanguagePasses is the false-positive control, and it is
// the half that matters most: describing SOMEONE ELSE'S regulatory status is
// exactly what the mortgage-lender directory is for. A guard that broke it would
// break our best content while looking like it was working.
func TestLegitimateRegulatorLanguagePasses(t *testing.T) {
	mustAllow := []string{
		"Nationwide Building Society is authorised and regulated by the Financial Conduct Authority.",
		// The bare string below is stored verbatim by the directory as a
		// lender's regulator_status. If this ever starts failing, the guard has
		// begun eating the directory.
		"authorised and regulated by the Financial Conduct Authority",
		"Lenders are authorised and regulated by the FCA.",
		"Your lender is regulated by the FCA, so you can check them on the register.",
		"Check the FCA register to confirm a firm's authorisation before applying.",
		"This site is not a broker and arranges nothing.",
		"We are independent of the FCA and not affiliated with it.",
		"Mansfield Building Society is a UK building society regulated by the PRA.",
		"The Mortgage Lender is an intermediary-only specialist mortgage provider.",
		"We are not authorised or regulated by the Financial Conduct Authority.",
	}
	for _, s := range mustAllow {
		if got := scanWith(t, nil, s); len(got) != 0 {
			t.Errorf("FALSE POSITIVE on %q → %q", s, got[0].Pattern)
		}
	}
}

func TestCompleteAttestationExemptsTheSite(t *testing.T) {
	eb := &EvidenceBase{Regulated: completeAttestation()}
	claim := "We are authorised and regulated by the Financial Conduct Authority."
	if got := ScanAllBannedClaims([]string{claim}, eb); len(got) != 0 {
		t.Fatalf("an attested firm must be allowed to state its status; got %d finding(s)", len(got))
	}
}

// TestExemptionDoesNotWidenToTheRestOfTheFleetSet is the mutation that proves the
// exemption is narrow. A regulated attestation must switch off ONLY the
// regulated family — if it disabled the accuracy-overclaim patterns too, an
// attested site would quietly lose every other fleet-wide protection, and no
// test of the regulated family alone would notice.
func TestExemptionDoesNotWidenToTheRestOfTheFleetSet(t *testing.T) {
	unattested := ScanAllBannedClaims([]string{"Figures without a source do not appear here."}, nil)
	if len(unattested) == 0 {
		t.Skip("the chosen control sentence no longer matches a global pattern; pick another")
	}
	eb := &EvidenceBase{Regulated: completeAttestation()}
	attested := ScanAllBannedClaims([]string{"Figures without a source do not appear here."}, eb)
	if len(attested) == 0 {
		t.Fatal("attestation switched off a NON-regulated fleet-wide pattern — the exemption is too wide")
	}
}

// TestIncompleteAttestationsDoNotExempt is the negative control set. Each case
// removes exactly one required field from a fixture that is otherwise valid, so
// a pass here means that field is genuinely load-bearing. Without these, an
// attestation an agent invented — or a half-filled one — would open the gate.
func TestIncompleteAttestationsDoNotExempt(t *testing.T) {
	claim := "We are authorised and regulated by the Financial Conduct Authority."
	cases := map[string]func(*RegulatedAttestation){
		"no firm name":      func(r *RegulatedAttestation) { r.FirmName = "" },
		"no FRN":            func(r *RegulatedAttestation) { r.FRN = "" },
		"FRN not numeric":   func(r *RegulatedAttestation) { r.FRN = "FCA305432" },
		"FRN too short":     func(r *RegulatedAttestation) { r.FRN = "3054" },
		"FRN too long":      func(r *RegulatedAttestation) { r.FRN = "305432123" },
		"nobody attested":   func(r *RegulatedAttestation) { r.AttestedBy = "" },
		"no evidence":       func(r *RegulatedAttestation) { r.Evidence = "" },
		"no date":           func(r *RegulatedAttestation) { r.AttestedAt = "" },
		"unreadable date":   func(r *RegulatedAttestation) { r.AttestedAt = "last Tuesday" },
		"whitespace only":   func(r *RegulatedAttestation) { r.FirmName = "   " },
		"whitespace evidnc": func(r *RegulatedAttestation) { r.Evidence = "  " },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			att := completeAttestation()
			mutate(att)
			if att.Attested() {
				t.Fatalf("%s: attestation reported complete when it is not", name)
			}
			eb := &EvidenceBase{Regulated: att}
			if got := ScanAllBannedClaims([]string{claim}, eb); len(got) == 0 {
				t.Fatalf("%s: an incomplete attestation opened the gate", name)
			}
		})
	}
}

func TestNilAttestationIsNotAttested(t *testing.T) {
	var r *RegulatedAttestation
	if r.Attested() {
		t.Fatal("a nil attestation must not be attested")
	}
	var eb *EvidenceBase
	if eb.RegulatedAttested() {
		t.Fatal("a nil evidence base must not be attested")
	}
}

// TestAttestationOnlyEvidenceBaseSurvivesParsing guards a trap in ParseEvidenceBase:
// it returns nil for a base with nothing scannable, and before this change a site
// whose evidence_base contained ONLY an attestation parsed to nil — so the
// operator would have recorded an attestation that could never take effect, with
// no error to tell them.
func TestAttestationOnlyEvidenceBaseSurvivesParsing(t *testing.T) {
	raw, err := json.Marshal(map[string]interface{}{"regulated": completeAttestation()})
	if err != nil {
		t.Fatal(err)
	}
	eb, err := ParseEvidenceBase(raw)
	if err != nil {
		t.Fatal(err)
	}
	if eb == nil {
		t.Fatal("an attestation-only evidence base parsed to nil — the attestation would be silently discarded")
	}
	if !eb.RegulatedAttested() {
		t.Fatal("attestation did not survive the round trip")
	}
}

// TestAttestationRoundTripsThroughJSON pins the wire names, because the
// attestation is written by hand into site_specs and a renamed field would fail
// open — the site would simply read as unattested, which looks identical to a
// site that never had one.
func TestAttestationRoundTripsThroughJSON(t *testing.T) {
	raw := []byte(`{"facts":[],"banned_claims":[],"regulated":{
		"firm_name":"Example Mortgages Ltd","frn":"305432","regulator":"FCA",
		"permissions":"mortgage arranging","attested_by":"owner",
		"attested_at":"2026-08-19T10:00:00Z","evidence":"email + register check"}}`)
	eb, err := ParseEvidenceBase(raw)
	if err != nil {
		t.Fatal(err)
	}
	if eb == nil || !eb.RegulatedAttested() {
		t.Fatalf("attestation did not parse from its documented wire shape: %+v", eb)
	}
	if eb.Regulated.FRN != "305432" || eb.Regulated.FirmName != "Example Mortgages Ltd" {
		t.Fatalf("field mapping wrong: %+v", eb.Regulated)
	}
}

// ── Round 1 council objections (correlation aac38d5b), answered as tests ──────

// TestRegulatedOnlyBaseIsNotSafeToWriteBack answers the guardian's HIGH-severity
// objection, and it does so by PINNING THE HAZARD rather than claiming it is
// absent.
//
// The objection: LANDMINES names EvidenceBase/ParseEvidenceBase in a cluster
// warning that "parsing evidence_base through its own typed struct and writing it
// back DELETES every citation, writer_line, fact field not captured by the
// struct". Relaxing ParseEvidenceBase's nil-return means rows carrying ONLY an
// attestation now parse non-nil where before they parsed to nil — so IF any
// caller persisted the struct, this would newly arm that deletion for a class of
// rows previously immune.
//
// Enumerated 2026-08-19, every caller of ParseEvidenceBase in the tree:
//
//	validate_page_content.go:1157, validate_page_content_stats.go:143,
//	discovery_checks/check_unverified_claims.go:318, cmd/regcheck, cmd/claimscan.
//
// All five READ. None writes back. The two real writers
// (refresh_evidence_base_action.go:683, evidence_citations.go:350) work on
// map[string]interface{} and never touch the struct, which is what makes the
// estate safe today.
//
// A caller enumeration goes stale the moment someone adds a caller, so this test
// pins the LOSS instead: it asserts, concretely, what a write-back would destroy.
// If someone ever makes the struct lossless, this test fails and they must delete
// it deliberately — which is the right way for that decision to be made.
func TestRegulatedOnlyBaseIsNotSafeToWriteBack(t *testing.T) {
	raw := []byte(`{
	  "writer_block": "house rules the writer must follow",
	  "schema_notes": "notes the struct does not model",
	  "facts": [{"id":"f1","claim":"x","kind":"metric",
	             "source":{"type":"url","citation":{"url":"https://example.gov.uk","quote":"verbatim"}},
	             "writer_line":"how to phrase it","verified_at":"2026-08-19"}],
	  "banned_claims": [],
	  "regulated": {"firm_name":"Example Mortgages Ltd","frn":"305432","regulator":"FCA",
	                "attested_by":"owner","attested_at":"2026-08-19","evidence":"email + register"}
	}`)
	eb, err := ParseEvidenceBase(raw)
	if err != nil || eb == nil {
		t.Fatalf("parse: %v / %v", err, eb)
	}
	if !eb.RegulatedAttested() {
		t.Fatal("attestation should have survived PARSING — only writing back is lossy")
	}

	roundTripped, err := json.Marshal(eb)
	if err != nil {
		t.Fatal(err)
	}
	for _, lost := range []string{"writer_block", "schema_notes", "citation", "writer_line"} {
		if strings.Contains(string(roundTripped), lost) {
			t.Fatalf("%q survived a struct round trip — the documented landmine may have been "+
				"fixed. If so this test should be deleted DELIBERATELY, and the ParseEvidenceBase "+
				"caller list re-checked, not just made green.", lost)
		}
	}
	t.Log("confirmed lossy: writer_block, schema_notes, source.citation and fact.writer_line " +
		"are all destroyed by parse+marshal. No ParseEvidenceBase caller may persist the struct.")
}

// TestUnattestedSitesSeeOnlyRegulatedAdditions answers the guardian's medium
// objection: ScanAllBannedClaimsWithSuppressed is the choke point for the deploy
// gate, the save guard AND the post-deploy audit, so this change is fleet-wide on
// every scan, not just new builds.
//
// "Byte-identical output pre/post" cannot be tested after the fact without the old
// binary, so this pins the equivalent invariant, which is the one that actually
// matters: every finding the global+site scan would have produced is STILL
// produced, and anything new belongs to the regulated family. A site with no
// regulated language in its copy sees no change at all.
func TestUnattestedSitesSeeOnlyRegulatedAdditions(t *testing.T) {
	regulatedPatterns := map[string]bool{}
	for _, b := range regulatedEvidence().BannedClaims {
		regulatedPatterns[b.Pattern] = true
	}

	siteEB := &EvidenceBase{BannedClaims: []BannedClaim{
		{Pattern: `\bguaranteed (acceptance|approval)\b`, Reason: "site rule"},
	}}
	for i := range siteEB.BannedClaims {
		siteEB.BannedClaims[i].re = mustCompileCI(siteEB.BannedClaims[i].Pattern)
	}

	corpus := []string{
		"Figures without a source do not appear here.",
		"Guaranteed approval for every applicant.",
		"Nationwide Building Society is authorised and regulated by the Financial Conduct Authority.",
		"Rates change daily, so check with the lender.",
		"We are authorised and regulated by the Financial Conduct Authority.",
	}
	for _, s := range corpus {
		combined := ScanAllBannedClaims([]string{s}, siteEB)
		baseline := dedupeByPattern(
			globalEvidence().ScanBannedClaims([]string{s}),
			siteEB.ScanBannedClaims([]string{s}),
		)
		seen := map[string]bool{}
		for _, f := range combined {
			seen[f.Pattern] = true
		}
		// Nothing the pre-change scan raised may have gone missing.
		for _, b := range baseline {
			if !seen[b.Pattern] {
				t.Errorf("pre-existing finding LOST on %q: %s", s, b.Pattern)
			}
		}
		// Anything extra must belong to the regulated family and nothing else.
		base := map[string]bool{}
		for _, b := range baseline {
			base[b.Pattern] = true
		}
		for _, f := range combined {
			if !base[f.Pattern] && !regulatedPatterns[f.Pattern] {
				t.Errorf("NEW non-regulated finding on %q: %s", s, f.Pattern)
			}
		}
	}
}

// TestNegatedRegulatedClaimsNeverReachTheNegationGuard answers the edit-quality
// objection, and it CORRECTS the claim my calibration implied.
//
// The objection: all six patterns are subject-to-verb spans, the shape the
// documented negation-guard defect targets (negatedClaimMatch scans BACKWARDS
// from match start), and the calibration showed a negated sentence passing
// without showing the guard's backward scan was exercised.
//
// Measured answer: it was NOT exercised, and cannot be. On every negated form the
// UNGUARDED scan (ScanBannedClaimsIgnoringNegation) also returns zero — the
// patterns require "authorised and regulated" / "we are <role>" ADJACENCY, and an
// interposed "not" breaks the match before any guard is consulted. So these
// patterns are negation-safe BY CONSTRUCTION, not by the negation guard.
//
// That distinction is the maintenance hazard this test exists to hold: loosen any
// of these patterns to tolerate intervening words and they WILL start depending on
// the backward scan, which this file has never tested and which has a documented
// defect.
func TestNegatedRegulatedClaimsNeverReachTheNegationGuard(t *testing.T) {
	negated := []string{
		"We are not authorised and regulated by the Financial Conduct Authority.",
		"We are not a credit broker.",
		"We are not FCA regulated.",
		"We do not hold FCA authorisation.",
		"We are not, and have never been, authorised and regulated by the FCA.",
	}
	for _, s := range negated {
		if got := ScanAllBannedClaims([]string{s}, nil); len(got) != 0 {
			t.Errorf("negated sentence flagged: %q → %s", s, got[0].Pattern)
		}
		if raw := regulatedEvidence().ScanBannedClaimsIgnoringNegation([]string{s}); len(raw) != 0 {
			t.Errorf("pattern MATCHED %q and was only saved by the negation guard — these "+
				"patterns are supposed to miss negated forms outright. The guard has a documented "+
				"backward-scan defect; do not start relying on it.", s)
		}
	}
}

func mustCompileCI(p string) *regexp.Regexp { return regexp.MustCompile("(?i)" + p) }
