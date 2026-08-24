// FILE: platform/orchestration/datahelpers/claims_practice.go
//
// The fleet-wide PRACTICE-CLAIMS family: first-person statements that this
// business physically DOES or DID things — tests, buys, weighs, measures,
// inspects, receives samples — together with the one way a site is permitted to
// say so: a recorded operating-history attestation.
//
// WHY IT EXISTS (bugs_open/380). A greenfield site with no evidence register
// shipped a 1,486-word review methodology in the present tense ("Where we can,
// we buy the tool at the same price a reader would pay"; "We garden ourselves,
// and we test what we can get our hands on") for a business that has never
// tested, bought or been sent a product. The evidence register holds facts
// about the WORLD; these are claims about US, and no register — however full —
// can adjudicate them. The owner's ruling: aspirations must not be stated as
// present-tense practice; say only what is sourced.
//
// WHY A SEPARATE FAMILY, and why it is NOT unioned into
// ScanAllBannedClaimsWithSuppressed. The regulated family (claims_regulated.go)
// is unioned there because every one of its sentences is false by construction
// for an unattested site. "We test every tool" is NOT false by construction —
// a real reviews site does test tools — so the fleet-wide bar in
// claims_global.go ("false-by-construction for every site we will ever run")
// is not met, and every consumer of the union maps a finding to a REFUSAL
// (gate blocker, save-floor refusal, discovery high). Owner decision
// 2026-08-24: this family ships at default severity WARNING — record, never
// refuse — through its own entry point, so severity stays a consumer decision.
// Flipping it to a refusal fleet-wide is RFC_003 §8 Q1, an architecture-review
// decision; TestPracticeFamilyIsNotInTheRefusingUnion pins that it cannot
// happen by accident.
//
// PRECISION OVER RECALL, deliberately. The patterns anchor on physical-practice
// verbs with an object. They deliberately EXCLUDE the grammar of legitimate
// operating sites: use/used, compare, describe, explain, review, supply, serve,
// deliver, help, believe, recommend, "our clients/customers". P5 REQUIRES "us"
// — "Manufacturers do occasionally send review samples" (the garden-tools
// sentence, no "us") is left to the LLM auditor's cold arm (migration 597),
// which reported it. The deterministic family is the cheap, always-on
// instrument (claimscan prints PRACTICE lines); the auditor is the closer.
//
// CALIBRATION, run before shipping as claims_global.go's header requires
// (claims_practice_test.go carries the tables):
//   - 12 of 12 must-catch sentences matched (the four garden-tools sentences
//     the owner quoted lead the list);
//   - 20 of 20 must-allow sentences passed, INCLUDING the honest disclosures
//     ("Where we have not used a tool directly, we say so"), negations,
//     intentions ("We aim to test every tool eventually"), second and third
//     person, and copy-about-the-copy ("We describe the steel…");
//   - dry run over the WHOLE live corpus, 1,867 components across every
//     active/deployed site (cmd/claimscan, 2026-08-24, complete export —
//     the first attempt silently dropped 414 rows on a kubectl stream error,
//     so count exported rows against the DB before trusting a dry run):
//     **12 findings on 6 sites** — garden-tools.uk 7 (every one a sentence
//     the owner named as invented), leopardessconsulting.co.uk 1 ("We test
//     every workflow on our own sites" — TRUE practice, the attestation
//     case), loanandmortgagecalculator.co.uk 1 ("we test them" — true),
//     dartsonline.com 1 ("We test barrel profiles" on a site that "holds no
//     stock"), cookly.uk 1 ("How we test what goes up", register-less), and
//     idea.uk 1 ("how we test your idea" — a SERVICE description read as
//     physical practice: the one clear false positive, the bare-verb "we test"
//     shape). Zero hedge suppressions on live copy. At warning severity that
//     precision is acceptable and the false-positive shape is named here so
//     the next calibration starts from it.
//
// SUPPRESSORS ARE OBSERVABLE. A match whose claim is negated in its clause
// ("we do not test", via negatedClaimMatch) or hedged by a preceding
// conditional ("if we test a tool, we say how") is returned in `suppressed`,
// never silently dropped — a suppressor that leaves no trace is the failure
// mode this estate keeps rediscovering (claims_global.go, the "silent guard"
// note). RE2 has no lookbehind, so the hedge is a scan-then-filter.
//
// The attestation lives in the site_specs `evidence_base` aspect under
// `operating_history`, beside `regulated`. It is a RECORD, not a flag: who
// attested, when, and what they saw. ParseEvidenceBase keeps a base that holds
// ONLY this attestation non-nil (same reasoning as Regulated), and
// HasScannableRegister keeps such a base from arming the unregistered-number
// scan at error severity — the latent hazard the Regulated widening carried.

package datahelpers

import (
	"regexp"
	"strings"
	"sync"
)

// OperatingHistoryAttestation is a site's recorded permission to describe its
// own practice — that it tests, buys, measures, visits, receives samples — as
// present fact. Every field a person had to supply is required; a missing one
// means NOT attested, because a partial record is exactly the shape a
// hallucinating agent would produce.
type OperatingHistoryAttestation struct {
	AttestedBy string `json:"attested_by"`    // the person who checked it
	AttestedAt string `json:"attested_at"`    // RFC3339 or YYYY-MM-DD
	Evidence   string `json:"evidence"`       // what was seen: receipts, a workshop visit, sample correspondence
	Note       string `json:"note,omitempty"` // free text: what the practice actually is
}

// Attested reports whether this attestation is usable. Nil-safe.
func (o *OperatingHistoryAttestation) Attested() bool {
	if o == nil {
		return false
	}
	if strings.TrimSpace(o.AttestedBy) == "" || strings.TrimSpace(o.Evidence) == "" {
		return false
	}
	return parseAttestedAt(o.AttestedAt) != nil
}

// OperatingHistoryAttested is the nil-safe form used by the scan path: a site
// with no evidence base at all is, correctly, not attested.
func (eb *EvidenceBase) OperatingHistoryAttested() bool {
	if eb == nil {
		return false
	}
	return eb.OperatingHistory.Attested()
}

// practiceClaims is the family. All first-person, all anchored on a physical
// verb (and, where the verb is ambiguous, on an object). See the header before
// adding a verb, and re-run the fleet dry run: cmd/claimscan prints PRACTICE
// lines for exactly this set.
func practiceClaims() []BannedClaim {
	const reason = "practice claim: this site has no recorded operating history, so a first-person " +
		"statement of physical practice cannot be true yet. Record an operating_history " +
		"attestation (who attested, when, what evidence) in its evidence_base — or reframe " +
		"as intent ('we aim to') — and this stops being reported."
	return []BannedClaim{
		{ // P1: we [adverbs] test|buy|weigh|inspect|garden|cook …
			Pattern: `\b(?:we|our\s+(?:team|reviewers|testers|editors))\s+(?:(?:have|had|also|regularly|routinely|personally|always|usually|often|occasionally|actually|already|physically)\s+){0,3}(?:test|tests|tested|testing|trial|trials|trialled|trialed|buy|buys|bought|purchase|purchases|purchased|weigh|weighs|weighed|inspect|inspects|inspected|dismantle|dismantles|dismantled|garden|gardens|gardened|cook|cooks|cooked)\b`,
			Reason:  reason,
		},
		{ // P2: we measure <the|each|every …>  ("we measure success" must not trip)
			Pattern: `\b(?:we|our\s+team)\s+(?:(?:have|also|regularly|routinely|personally|often)\s+){0,2}measure[sd]?\s+(?:each|every|the|all|a|an|these|those|its|their)\b`,
			Reason:  reason,
		},
		{ // P3: we record the <weights|measurements|…>  ("we record your consent" must not trip)
			Pattern: `\b(?:we|our\s+team)\s+record(?:s|ed)?\s+(?:the\s+)?(?:\w+\s+){0,2}(?:used|weight|weights|measurement|measurements|reading|readings|result|results|dimension|dimensions|material|materials|spec|specs|specification|specifications)\b`,
			Reason:  reason,
		},
		{ // P4: we receive|get|accept|are sent [free|review|test|press] samples|units …
			Pattern: `\bwe\s+(?:sometimes\s+|occasionally\s+|regularly\s+|often\s+|do\s+)?(?:receive|get|accept|are\s+sent)\s+(?:free\s+|review\s+|test\s+|press\s+)?(?:samples|units|products|loaners?)\b`,
			Reason:  reason,
		},
		{ // P5: manufacturers|brands|… send|provide|lend US — "us" required
			Pattern: `\b(?:manufacturers|brands|suppliers|retailers|makers)\s+(?:do\s+)?(?:occasionally\s+|sometimes\s+|often\s+|regularly\s+|also\s+)?(?:send|sends|sent|provide|provides|provided|supply|supplies|supplied|lend|lends|lent)\s+us\b`,
			Reason:  reason,
		},
	}
}

var (
	practiceEvidenceOnce sync.Once
	practiceEvidenceBase *EvidenceBase
)

// practiceEvidence holds the compiled family. Same contract as
// regulatedEvidence: unexported, never mutated, never handed to a writer, and a
// pattern that fails to compile degrades to a literal substring rather than
// silently vanishing.
func practiceEvidence() *EvidenceBase {
	practiceEvidenceOnce.Do(func() {
		bans := practiceClaims()
		for i := range bans {
			re, err := regexp.Compile("(?i)" + bans[i].Pattern)
			if err != nil {
				re = regexp.MustCompile("(?i)" + regexp.QuoteMeta(bans[i].Pattern))
			}
			bans[i].re = re
		}
		practiceEvidenceBase = &EvidenceBase{BannedClaims: bans}
	})
	return practiceEvidenceBase
}

// PracticeClaimCount reports how many practice-family patterns are active.
// A family that compiles to nothing looks exactly like one that works.
func PracticeClaimCount() int { return len(practiceEvidence().BannedClaims) }

// practiceHedgeRe matches a conditional that turns a practice statement into a
// hypothetical: "if we test a tool, we say how". Applied to the ≤24 bytes
// preceding a match (RE2 has no lookbehind, so the scan filters after matching).
var practiceHedgeRe = regexp.MustCompile(`(?i)\b(?:if|when|whenever|before|once|should|unless|until)\s+$`)

// hedgedPracticeMatch reports whether the match starting at `start` in `block`
// is introduced by a conditional within the preceding 24 bytes.
func hedgedPracticeMatch(block string, start int) bool {
	lo := start - 24
	if lo < 0 {
		lo = 0
	}
	return practiceHedgeRe.MatchString(block[lo:start])
}

// ScanPracticeClaims scans blocks against the practice family ALONE, honouring
// the operating-history exemption. Findings carry Check == "practice_claim"
// so every consumer can tell them from banned claims and grade them itself.
func ScanPracticeClaims(blocks []string, eb *EvidenceBase) []ClaimFinding {
	findings, _ := ScanPracticeClaimsWithSuppressed(blocks, eb)
	return findings
}

// ScanPracticeClaimsWithSuppressed additionally returns the matches that the
// negation guard or the hedge filter REMOVED, each with its Reason set to why.
// cmd/claimscan prints them; an operator diffing a dry run can see exactly what
// the suppressors are doing to real copy.
func ScanPracticeClaimsWithSuppressed(blocks []string, eb *EvidenceBase) (findings, suppressed []ClaimFinding) {
	if eb.OperatingHistoryAttested() {
		return nil, nil
	}
	fam := practiceEvidence()
	live := make(map[string]*ClaimFinding)
	gone := make(map[string]*ClaimFinding)
	var liveOrder, goneOrder []string

	record := func(m map[string]*ClaimFinding, order *[]string, bc *BannedClaim, block string, loc []int, reason string) {
		if f, ok := m[bc.Pattern]; ok {
			f.Occurrences++
			return
		}
		m[bc.Pattern] = &ClaimFinding{
			Check:       "practice_claim",
			Matched:     block[loc[0]:loc[1]],
			Pattern:     bc.Pattern,
			Reason:      reason,
			Snippet:     claimSnippet(block, loc[0], loc[1]),
			Occurrences: 1,
		}
		*order = append(*order, bc.Pattern)
	}

	for _, block := range blocks {
		for i := range fam.BannedClaims {
			bc := &fam.BannedClaims[i]
			if bc.re == nil {
				continue
			}
			for _, loc := range bc.re.FindAllStringIndex(block, -1) {
				switch {
				case negatedClaimMatch(block, loc[0]):
					record(gone, &goneOrder, bc, block, loc, "negated in its own clause")
				case hedgedPracticeMatch(block, loc[0]):
					record(gone, &goneOrder, bc, block, loc, "hedged by a preceding conditional")
				default:
					record(live, &liveOrder, bc, block, loc, bc.Reason)
				}
			}
		}
	}
	for _, p := range liveOrder {
		findings = append(findings, *live[p])
	}
	for _, p := range goneOrder {
		suppressed = append(suppressed, *gone[p])
	}
	return findings, suppressed
}
