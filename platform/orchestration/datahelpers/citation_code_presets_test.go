// FILE: platform/orchestration/datahelpers/citation_code_presets_test.go
//
// RFC_060 Q5 (owner-ruled 2026-09-03 §3f): per-site citation_codes, additive
// over the fleet-wide FCA default, plus named sector presets (veterinary,
// legal, medical) the owner named as imminent second/third consumers. See
// claims.go's compileCitationCodeRegexes and the EvidenceBase field
// comments for the design; this file is the behavioural proof.

package datahelpers

import "testing"

// scanWithEB mirrors claims_test.go's regNumberFindings, parameterised on
// the EvidenceBase so each case can carry its own CitationCodes/Presets.
func scanWithEB(t *testing.T, eb *EvidenceBase, sentence string) []ClaimFinding {
	t.Helper()
	eb.BannedClaims = []BannedClaim{{Pattern: "zzz-never-matches"}} // arms the scan, as regNumberFindings does
	eb.compileCitationCodeRegexes()
	return eb.ScanUnregisteredNumbers([]string{sentence}, ClaimSurface{})
}

func TestCitationCodePreset_VeterinaryExemptsRCVSAndVMD(t *testing.T) {
	eb := &EvidenceBase{CitationCodePresets: []string{"veterinary"}}
	cases := []string{
		"Our 12 clients were referred under RCVS 5 guidance on second opinions.",
		"We hold 8 records logged under VMD 3 for veterinary medicines.",
	}
	for _, s := range cases {
		if f := scanWithEB(t, eb, s); len(f) != 0 {
			t.Errorf("veterinary preset should exempt this regulatory citation, got findings: %q -> %+v", s, f)
		}
	}
}

func TestCitationCodePreset_LegalExemptsSRA(t *testing.T) {
	eb := &EvidenceBase{CitationCodePresets: []string{"legal"}}
	s := "We processed 9 complaints under SRA 7 conduct rules last year."
	if f := scanWithEB(t, eb, s); len(f) != 0 {
		t.Errorf("legal preset should exempt SRA citations, got: %+v", f)
	}
}

func TestCitationCodePreset_MedicalExemptsGMCMHRACQC(t *testing.T) {
	eb := &EvidenceBase{CitationCodePresets: []string{"medical"}}
	// ⚠ each sentence carries a businessClaimContextRe noun (clients/records/
	// processed) deliberately — an earlier draft used "clinicians"/"incidents"/
	// "team", none of which qualify, so the scan never reached the number for
	// either reason and the test passed vacuously regardless of whether the
	// exemption worked. Verified these fire by construction, per the file's
	// own established discipline (claims_test.go's mustCatch fixtures).
	cases := []string{
		"Our 6 clients are registered under GMC 2 fitness-to-practise rules.",
		"We logged 14 records under MHRA 9 yellow card guidance.",
		"We processed 5 inspections under CQC 1 fundamental standards.",
	}
	for _, s := range cases {
		if f := scanWithEB(t, eb, s); len(f) != 0 {
			t.Errorf("medical preset should exempt this citation, got findings: %q -> %+v", s, f)
		}
	}
}

// A preset NOT opted into must not exempt anything — a veterinary-only site
// still gets convicted for an unbacked SRA-shaped number, proving presets
// are genuinely additive per site, not a blanket widening.
func TestCitationCodePreset_NotOptedInDoesNotExempt(t *testing.T) {
	eb := &EvidenceBase{CitationCodePresets: []string{"veterinary"}}
	s := "We processed 9 complaints under SRA 7 conduct rules last year."
	if f := scanWithEB(t, eb, s); len(f) == 0 {
		t.Fatalf("a site that only opted into 'veterinary' must NOT get SRA's exemption for free")
	}
}

// An unrecognised preset name must contribute nothing — silently inert,
// never a panic, never treated as a literal code.
func TestCitationCodePreset_UnknownNameContributesNothing(t *testing.T) {
	eb := &EvidenceBase{CitationCodePresets: []string{"veterinery"}} // typo, deliberately
	s := "We processed 9 complaints under SRA 7 conduct rules last year."
	if f := scanWithEB(t, eb, s); len(f) == 0 {
		t.Fatalf("an unrecognised preset name must not silently grant any exemption")
	}
}

func TestCitationCodes_AdHocCodeIsExempted(t *testing.T) {
	eb := &EvidenceBase{CitationCodes: []string{"PSR"}} // Payment Systems Regulator, not in any preset
	s := "We logged 40 records under PSR 6 access rules each quarter."
	if f := scanWithEB(t, eb, s); len(f) != 0 {
		t.Errorf("an ad hoc CitationCodes entry should exempt its own citation, got: %+v", f)
	}
}

// RFC_060 §3f constraint 1: the matching rule does not change. A two-letter
// ad hoc code must be silently dropped — even site-declared, it is "too
// collidable" per fad209b92's own reasoning against bare FCA/PRA/FOS.
//
// ⚠ THE FIXTURE MUST CARRY A businessClaimContextRe NOUN ("records") — an
// earlier draft used "disputes"/"access", neither of which is in that list,
// so the scan never reached the number for EITHER reason and the test
// passed regardless of whether two-letter filtering worked. Confirmed this
// version actually exercises the path: TestCitationCodes_AdHocCodeIsExempted
// above proves the SAME sentence shape correctly finds zero findings when
// the code is real (PSR, 3 letters) — so a non-zero result here is
// PS-specific, not a fixture that can never fire.
func TestCitationCodes_TwoLetterCodeIsDropped(t *testing.T) {
	eb := &EvidenceBase{CitationCodes: []string{"PS"}}
	s := "We logged 40 records under PS 6 access rules each quarter."
	if f := scanWithEB(t, eb, s); len(f) == 0 {
		t.Fatalf("a two-letter ad hoc code must be dropped, not exempt a number")
	}
}

// Constraint 1, continued: case-sensitivity is preserved for NEW codes too,
// not just the fleet default — lowercase "rcvs" must not exempt anything.
func TestCitationCodePreset_CaseSensitivityPreservedForNewCodes(t *testing.T) {
	eb := &EvidenceBase{CitationCodePresets: []string{"veterinary"}}
	s := "Our 12 clients were referred under rcvs 5 guidance on second opinions."
	if f := scanWithEB(t, eb, s); len(f) == 0 {
		t.Fatalf("lowercase 'rcvs' must NOT be treated as a citation — case-sensitivity must hold for site-declared codes too")
	}
}

// And digit-adjacency: a bare regulator NAME with no following digit is a
// subject, not a citation, for new codes exactly as it already is for FCA.
func TestCitationCodePreset_BareNameWithNoDigitIsNotACitation(t *testing.T) {
	eb := &EvidenceBase{CitationCodePresets: []string{"veterinary"}}
	s := "Our 12 clients were referred under RCVS guidance on second opinions."
	if f := scanWithEB(t, eb, s); len(f) == 0 {
		t.Fatalf("'RCVS' with no digit following must NOT be treated as a citation")
	}
}

// A site declaring nothing must compile to patterns BYTE IDENTICAL to the
// fleet-wide default — the additive-only guarantee, asserted directly.
func TestCompileCitationCodeRegexes_EmptyDeclarationMatchesFleetDefault(t *testing.T) {
	eb := &EvidenceBase{}
	eb.compileCitationCodeRegexes()
	if eb.citationContextRe.String() != regulatoryCitationContextRe.String() {
		t.Errorf("a site with no CitationCodes/Presets must compile to the EXACT fleet default:\n got:  %s\n want: %s",
			eb.citationContextRe.String(), regulatoryCitationContextRe.String())
	}
	if eb.citationPrefixRe.String() != rulebookCitationPrefixRe.String() {
		t.Errorf("a site with no CitationCodes/Presets must compile to the EXACT fleet default:\n got:  %s\n want: %s",
			eb.citationPrefixRe.String(), rulebookCitationPrefixRe.String())
	}
}

// Two EvidenceBase instances must never leak state into each other — each
// site's codes are its own, even when compiled in the same test process.
func TestCompileCitationCodeRegexes_SitesDoNotLeakIntoEachOther(t *testing.T) {
	vet := &EvidenceBase{CitationCodePresets: []string{"veterinary"}}
	vet.compileCitationCodeRegexes()
	plain := &EvidenceBase{}
	plain.compileCitationCodeRegexes()

	if !vet.citationContextRe.MatchString("under RCVS 5") {
		t.Fatalf("the veterinary-opted-in base should match RCVS 5")
	}
	if plain.citationContextRe.MatchString("under RCVS 5") {
		t.Fatalf("a plain base must NOT match RCVS 5 just because a DIFFERENT EvidenceBase instance opted in")
	}
}

func TestKnownCitationCodePresets_ListsAllThree(t *testing.T) {
	got := KnownCitationCodePresets()
	want := map[string]bool{"veterinary": true, "legal": true, "medical": true}
	if len(got) != len(want) {
		t.Fatalf("expected %d presets, got %d: %v", len(want), len(got), got)
	}
	for _, name := range got {
		if !want[name] {
			t.Errorf("unexpected preset name %q", name)
		}
	}
}

// ParseEvidenceBase is the real production path (JSON in, compiled regexes
// out) — proving the wiring, not just the direct-construction unit tests
// above.
func TestParseEvidenceBase_CompilesCitationCodesFromJSON(t *testing.T) {
	data := []byte(`{
		"banned_claims": [{"pattern": "zzz-never-matches"}],
		"citation_code_presets": ["veterinary"],
		"citation_codes": ["PSR"]
	}`)
	eb, err := ParseEvidenceBase(data)
	if err != nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	if eb == nil {
		t.Fatalf("expected a non-nil base (it carries banned_claims)")
	}
	if eb.citationContextRe == nil {
		t.Fatalf("ParseEvidenceBase must compile the citation regexes, not leave them nil")
	}
	if !eb.citationContextRe.MatchString("under RCVS 5") {
		t.Errorf("the preset from JSON did not take effect")
	}
	if !eb.citationContextRe.MatchString("under PSR 6") {
		t.Errorf("the ad hoc code from JSON did not take effect")
	}
}
