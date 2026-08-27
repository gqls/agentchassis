package datahelpers

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// completeOperatingHistory is the only shape that opens the gate; tests that
// need a NON-attested site mutate one field of it, so each fails for the reason
// it names rather than because the fixture was never valid.
func completeOperatingHistory() *OperatingHistoryAttestation {
	return &OperatingHistoryAttestation{
		AttestedBy: "owner",
		AttestedAt: "2026-08-24",
		Evidence:   "receipts for six tools bought 2026-07; workshop visit 2026-08-20",
		Note:       "we buy and bench-test hand tools before recommending them",
	}
}

func practiceScan(t *testing.T, eb *EvidenceBase, sentence string) []ClaimFinding {
	t.Helper()
	return ScanPracticeClaims([]string{sentence}, eb)
}

// The garden-tools.uk sentences the owner quoted (bugs_open/380 §1) lead the
// list; the rest are the same grammar on the other register-less deployed
// sites (cookly.uk, gaswholesalers.com).
var practiceMustCatch = []string{
	"Where we can, we buy the tool at the same price a reader would pay",
	"We garden ourselves, and we test what we can get our hands on",
	"We weigh every tool on a kitchen scale",
	"We have tested more than forty spades",
	"Our team tests each product for a month",
	"We receive free samples from brands",
	"We measure the blade thickness with callipers",
	"We buy every product we review",
	"We cook every recipe at least twice",
	"We are sent review units",
	"We record the measurements in a shared log",
	"Suppliers regularly send us test units",
	// P6, the documentary-diligence conjunction (bugs_open/414). The first is
	// the served lendzy.co.uk sentence VERBATIM — a bare participle with no
	// subject, which is why P6 is not first-person anchored.
	"We explain what the FCA rulebook says a lender can and cannot do, checked against the FCA handbook, rule by rule, so you can hold your lender to it",
	"Our guides are checked against the FCA handbook, rule by rule",
	// Family generality: a different verb, rulebook and unit each time.
	"We verify our guidance against the Consumer Credit sourcebook, clause by clause",
	"Each page is audited against the regulations, section by section",
	// Order B — the idiom leads and the checking follows.
	"Rule by rule, our guides are cross-checked against the handbook",
}

// The grammar of a legitimate operating site, honest disclosures, negations,
// hedges, second/third person, and the copy-about-the-copy shapes the family
// must never convict (the fleet-wide bar, claims_global.go).
var practiceMustAllow = []string{
	"We help small businesses build their first site",
	"We believe plain language beats jargon",
	"We explain how each calculator works",
	"We recommend starting with the cheapest tier",
	"We link to retailers so you can check today's price",
	"Where we have not used a tool directly, we say so",
	"We do not test every tool we list",
	"We have not bought any of these products",
	"We don't receive samples from anyone",
	"We cannot test everything we cover",
	"If we test a tool, we say how",
	"We measure success by whether you found what you needed",
	"We record your consent when you accept cookies",
	"We compare lenders' published rates side by side",
	"We aim to test every tool eventually",
	"Nationwide tests its products before launch",
	"Manufacturers send review samples to publications",
	"How do you test the tools you feature?",
	"Every figure here comes from a manufacturer's stated specification",
	"We describe the steel, the handle material, and the grading standard",
	// bugs_open/414 — each of these is why a fragment of P6 is shaped as it is.
	// The live imperative on the motivating site itself: verb + against +
	// rulebook, addressed to the READER, with no exhaustive idiom.
	"Check your loan against the FCA rules",
	// The exhaustive idiom on its own is ordinary guide copy — 22 components in
	// the live corpus carry one (2026-08-27).
	"Work through your agreement line by line before you sign",
	"We walk you through the affordability rules, step by step",
	// `step` and `register` are deliberately outside P6's lists. The second
	// sentence is a live ATTESTED FACT on leopardessconsulting.co.uk.
	"Records verified against the Companies House register, entry by entry",
	// Negated and hedged forms of P6 itself.
	"Our guides have not been checked against the FCA handbook, rule by rule",
	"If we check a guide against the handbook, rule by rule, we say so",
	// Describing what we do, without claiming it was done exhaustively.
	"We name the rule beside every figure and link to the handbook so you can read it",
}

// A family that silently compiles to nothing looks exactly like one that works.
func TestPracticeFamilyIsWired(t *testing.T) {
	if n := PracticeClaimCount(); n == 0 {
		t.Fatalf("practice family is empty — the instrument is inert")
	}
	for i, b := range practiceEvidence().BannedClaims {
		if b.re == nil {
			t.Errorf("pattern %d did not compile and has no fallback", i)
		}
		if !strings.Contains(b.Reason, "operating_history attestation") {
			t.Errorf("pattern %d's reason does not tell the reader how to fix it: %q", i, b.Reason)
		}
	}
}

func TestEveryPracticePatternIsAValidRegex(t *testing.T) {
	for i, b := range practiceClaims() {
		if _, err := regexp.Compile("(?i)" + b.Pattern); err != nil {
			t.Errorf("pattern %d does not compile as RE2: %v\n%s", i, err, b.Pattern)
		}
	}
}

func TestPracticeClaimsAreReportedWithoutAttestation(t *testing.T) {
	for _, s := range practiceMustCatch {
		found := practiceScan(t, nil, s)
		if len(found) == 0 {
			t.Errorf("must-catch sentence produced no finding: %q", s)
			continue
		}
		for _, f := range found {
			if f.Check != "practice_claim" {
				t.Errorf("finding for %q is labelled %q, want practice_claim", s, f.Check)
			}
		}
	}
	// A register with facts but no attestation is still not attested.
	eb := &EvidenceBase{Facts: []EvidenceFact{{ID: "F1", Claim: "trading since 2024"}}}
	if len(practiceScan(t, eb, practiceMustCatch[0])) == 0 {
		t.Errorf("a register with facts but no operating_history must still report practice claims")
	}
}

func TestLegitimateFirstPersonCopyPasses(t *testing.T) {
	for _, s := range practiceMustAllow {
		if found := practiceScan(t, nil, s); len(found) != 0 {
			t.Errorf("must-allow sentence tripped %d finding(s): %q — first: %q", len(found), s, found[0].Matched)
		}
	}
}

func TestPracticeSuppressionsAreObservable(t *testing.T) {
	// "We do not test …" never MATCHES (the negator sits between "we" and the
	// verb, outside every pattern) — structural exclusion, asserted in the
	// must-allow table. The negation guard handles a negator BEFORE the match.
	cases := map[string]string{
		"Not once have we tested a tool we list": "negated",
		"If we test a tool, we say how":          "hedged",
	}
	for sentence, why := range cases {
		findings, suppressed := ScanPracticeClaimsWithSuppressed([]string{sentence}, nil)
		if len(findings) != 0 {
			t.Errorf("%q must not be a finding (%s)", sentence, why)
		}
		if len(suppressed) == 0 {
			t.Errorf("%q must appear in the suppressed list, or the %s guard is silent", sentence, why)
			continue
		}
		if !strings.Contains(suppressed[0].Reason, why) {
			t.Errorf("%q suppressed for %q, want a reason naming %s", sentence, suppressed[0].Reason, why)
		}
	}
}

func TestCompleteOperatingHistoryAttestationExemptsTheSite(t *testing.T) {
	eb := &EvidenceBase{OperatingHistory: completeOperatingHistory()}
	if !eb.OperatingHistoryAttested() {
		t.Fatalf("complete attestation not recognised")
	}
	for _, s := range practiceMustCatch {
		if found := practiceScan(t, eb, s); len(found) != 0 {
			t.Errorf("attested site still reported %q", s)
		}
	}
}

func TestIncompleteOperatingHistoryDoesNotExempt(t *testing.T) {
	mutations := map[string]func(o *OperatingHistoryAttestation){
		"no attested_by": func(o *OperatingHistoryAttestation) { o.AttestedBy = " " },
		"no evidence":    func(o *OperatingHistoryAttestation) { o.Evidence = "" },
		"bad date":       func(o *OperatingHistoryAttestation) { o.AttestedAt = "last week" },
		"no date":        func(o *OperatingHistoryAttestation) { o.AttestedAt = "" },
	}
	for name, mutate := range mutations {
		o := completeOperatingHistory()
		mutate(o)
		eb := &EvidenceBase{OperatingHistory: o}
		if eb.OperatingHistoryAttested() {
			t.Errorf("%s: attestation should not count", name)
		}
		if len(practiceScan(t, eb, practiceMustCatch[0])) == 0 {
			t.Errorf("%s: an incomplete attestation must not exempt the site", name)
		}
	}
	var nilAtt *OperatingHistoryAttestation
	if nilAtt.Attested() {
		t.Errorf("nil attestation reported as attested")
	}
}

// An attestation-only base must survive parsing (or the exemption could never
// fire) AND must not become scannable (or the number scan would arm at error
// severity against an empty fact list — the CGV-033 latent hazard).
func TestOperatingHistoryOnlyBaseSurvivesParsing(t *testing.T) {
	raw := []byte(`{"operating_history":{"attested_by":"owner","attested_at":"2026-08-24","evidence":"receipts"}}`)
	eb, err := ParseEvidenceBase(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if eb == nil {
		t.Fatalf("attestation-only base parsed to nil — the exemption can never fire")
	}
	if !eb.OperatingHistoryAttested() {
		t.Errorf("attestation not recognised after parse")
	}
	if eb.HasScannableRegister() {
		t.Errorf("attestation-only base must NOT be scannable")
	}
	// The regulated-only shape is guarded the same way now.
	reg, _ := ParseEvidenceBase([]byte(`{"regulated":{"firm_name":"X Ltd","frn":"305432","attested_by":"owner","attested_at":"2026-08-19","evidence":"register"}}`))
	if reg == nil || reg.HasScannableRegister() {
		t.Errorf("regulated-only base: nil=%v scannable=%v; want non-nil and not scannable", reg == nil, reg != nil && reg.HasScannableRegister())
	}
}

// Registers with content stay scannable exactly as before.
func TestRegistersWithContentStayScannable(t *testing.T) {
	facts := &EvidenceBase{Facts: []EvidenceFact{{ID: "F1", Claim: "x"}}}
	bans := &EvidenceBase{BannedClaims: []BannedClaim{{Pattern: "never", Reason: "r"}}}
	var none *EvidenceBase
	if !facts.HasScannableRegister() || !bans.HasScannableRegister() {
		t.Errorf("a register with facts or bans must be scannable")
	}
	if none.HasScannableRegister() || (&EvidenceBase{}).HasScannableRegister() {
		t.Errorf("nil / empty base must not be scannable")
	}
}

func TestOperatingHistoryRoundTripsThroughJSON(t *testing.T) {
	eb := &EvidenceBase{OperatingHistory: completeOperatingHistory()}
	b, err := json.Marshal(eb)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"operating_history"`) {
		t.Errorf("operating_history missing from JSON: %s", b)
	}
	back, err := ParseEvidenceBase(b)
	if err != nil || back == nil || !back.OperatingHistoryAttested() {
		t.Errorf("round trip lost the attestation: err=%v back=%+v", err, back)
	}
}

// THE MUTATION TEST. The practice family must not join the refusing union:
// every consumer of ScanAllBannedClaims maps a finding to a refusal (gate
// blocker, save-floor refusal, discovery high). That flip is RFC_003 §8 Q1 —
// an architecture-review decision (bugs_open/380) — and cannot happen by
// accident. If this fails, someone unioned it.
func TestPracticeFamilyIsNotInTheRefusingUnion(t *testing.T) {
	for _, s := range practiceMustCatch {
		found, _ := ScanAllBannedClaimsWithSuppressed([]string{s}, nil)
		for _, f := range found {
			if f.Check == "practice_claim" || strings.Contains(f.Reason, "operating_history") {
				t.Fatalf("the practice family must not join the refusing union — that flips warning to blocker "+
					"fleet-wide and is RFC_003 Q1, an architecture-review decision (bugs_open/380). Sentence %q produced %+v", s, f)
			}
		}
	}
}

// The exemption is narrow by construction: an operating-history attestation
// switches off ONLY the practice family, never the regulated or global sets.
func TestOperatingHistoryExemptionDoesNotWidenToRegulatedOrGlobal(t *testing.T) {
	eb := &EvidenceBase{OperatingHistory: completeOperatingHistory()}
	if got := ScanAllBannedClaims([]string{"We are authorised and regulated by the Financial Conduct Authority."}, eb); len(got) == 0 {
		t.Errorf("an operating-history attestation must not exempt the regulated family")
	}
}

// ---------------------------------------------------------------------------
// bugs_open/414 — P6's coupling and its two known residuals, each pinned by a
// test so it is a decision on the record and not a surprise to the next reader.
// ---------------------------------------------------------------------------

// THE COUPLING. An operating_history attestation exempts the WHOLE family, so a
// site attesting that it really does test products also switches off the
// documentary-diligence pattern, which is about a different kind of work.
// [MEASURED 2026-08-27: 0 sites in the estate carry either attestation, so this
// is unexercised today.] This test asserts the coupling EXISTS rather than
// pretending it does not: if it ever needs breaking, the fix is a separate
// family with its own attestation (claims_regulated.go is the shape), and this
// test is what will fail and say so.
func TestPracticeDiligenceHonoursTheAttestation(t *testing.T) {
	const s = "Our guides are checked against the FCA handbook, rule by rule"

	if f := practiceScan(t, &EvidenceBase{}, s); len(f) != 1 {
		t.Fatalf("unattested site: expected 1 diligence finding, got %d (%+v)", len(f), f)
	}
	attested := &EvidenceBase{OperatingHistory: completeOperatingHistory()}
	if f := practiceScan(t, attested, s); len(f) != 0 {
		t.Errorf("attested site: the whole family is exempt, so P6 must be silent too; got %+v", f)
	}
}

// RESIDUAL 1 — a THIRD-PARTY subject. P6 has no subject anchor (see the header),
// so a sentence about someone else's checking matches. Reported, not hidden:
// at WARNING severity the cost is a line in a report an operator reads, which is
// why this family is not in the refusing union. [MEASURED 2026-08-27: zero live
// components carry this shape.]
//
// RESIDUAL 2 — the CORRECTING DISCLOSURE. negationCueRe (claims.go) carries no
// bare "no"/"nothing"/"without" cue, so "Nothing here has been checked…" reads
// as un-negated. Adding those cues would change a guard SHARED by every claims
// family — a guarantee change, i.e. architecture-scope — so it is recorded here
// rather than fixed in passing.
func TestPracticeDiligenceKnownResiduals(t *testing.T) {
	residuals := []string{
		"The FCA checks firms against the handbook, rule by rule",
		"Nothing here has been checked against the FCA handbook, rule by rule",
	}
	for _, s := range residuals {
		if f := practiceScan(t, &EvidenceBase{}, s); len(f) == 0 {
			t.Errorf("residual no longer fires — if that was deliberate, delete it from this "+
				"list and say so in the header: %q", s)
		}
	}
}

// One sentence must yield ONE finding, not one per word order: P6 is a single
// entry with an internal alternation, and findings are keyed by pattern.
func TestPracticeDiligenceReportsOncePerSentence(t *testing.T) {
	f := practiceScan(t, &EvidenceBase{},
		"Rule by rule, our guides are checked against the FCA handbook, rule by rule")
	if len(f) != 1 {
		t.Fatalf("expected exactly 1 finding for one sentence, got %d (%+v)", len(f), f)
	}
}
