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
//
// ---------------------------------------------------------------------------
// P6 — DOCUMENTARY diligence, added 2026-08-27 (bugs_open/414), and it breaks
// two of this family's own rules on purpose. Both departures are stated here
// because a reader who takes the rules above at face value will think P6 is a
// mistake.
//
// DEPARTURE 1: IT IS NOT FIRST-PERSON. Every pattern above requires "we" or
// "our team" and the header calls that design. P6 cannot, because the sentence
// that filed the bug has no subject at all:
//
//	"We explain what the FCA rulebook says a lender can and cannot do,
//	 checked against the FCA handbook, rule by rule, so you can hold your
//	 lender to it."                              (lendzy.co.uk /about.html)
//
// A first-person anchor would have been INERT on the motivating case — the
// same failure the fleet-wide set records when it narrowed a pattern by
// reasoning about a hypothetical false positive and measured 0 findings where
// the bare form found 2 real overclaims. What replaces the subject anchor is a
// four-part conjunction (verb + "against" + a named rulebook + an exhaustive
// "unit by unit" idiom), which is measured, not argued: see the constant's own
// comment for the 22 / 13 / 3-of-2,405 calibration.
//
// DEPARTURE 2: IT IS DOCUMENTARY, NOT PHYSICAL. "Checked against a handbook" is
// not weighing, buying or dismantling anything. It sits in this family because
// the family's real subject is *claims about us that no register can
// adjudicate* — the evidence base holds facts about the WORLD, and "we did this
// checking" is not one of them. That is the sentence in the header above, and it
// covers reading as well as testing.
//
// KNOWN COUPLING, and the reason it is written down rather than fixed: an
// `operating_history` attestation exempts the WHOLE family, so a site attesting
// that it really does test products would also switch off P6, which is about a
// different kind of work. [MEASURED 2026-08-27: **0** sites in the estate carry
// an operating_history attestation and 0 carry a regulated one, so the coupling
// is unexercised today.] TestPracticeDiligenceHonoursTheAttestation pins the
// behaviour so it is a decision on the record rather than a surprise. If a
// client ever needs one attestation without the other, the fix is a separate
// family with its own attestation, exactly as claims_regulated.go is separate
// from this file — cheap to do then, unfalsifiable to maintain now.
//
// WHY NOT THE REFUSING SET, where a completeness claim like "Everything on this
// site is checked" correctly lives (claims_global.go): "we check your policies
// against the FCA handbook, clause by clause" is a SERVICE a compliance
// consultancy really sells, so the fleet-wide bar — false-by-construction for
// every site we will ever run — is not met. And there is a sentence that settles
// it: the negation guard has no bare "no"/"nothing" cue, so
//
//	"Nothing here has been checked against the FCA handbook, rule by rule"
//
// — a correcting disclosure, close to what an honest repair might write — reads
// as un-negated. At blocker severity this layer would refuse the very sentence
// it exists to encourage. At warning it reports it, an operator reads it, and
// nothing breaks. Owner decision, 2026-08-27.

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
		{ // P6: DOCUMENTARY diligence — our copy was checked against a named
			// rulebook, exhaustively. See the header's P6 section for why this
			// pattern is not first-person and why the conjunction is the anchor.
			Pattern: diligenceAgainstRulebookPattern,
			Reason: "diligence overclaim: asserts this site's content was checked against a named " +
				"rulebook exhaustively, unit by unit. Nothing in this system performs such a check, " +
				"and no reader can verify that it happened. Say what the site DOES instead — name the " +
				"rule beside the figure and link it. If a person really did commission such a review, " +
				"record an operating_history attestation (who attested, when, what evidence) in the " +
				"site's evidence_base and this stops being reported. bugs_open/414.",
		},
	}
}

// diligenceAgainstRulebookPattern is P6, composed from named fragments because
// the whole thing on one line is unreadable and unreviewable. ONE alternation of
// the two word orders, not two entries: one sentence must yield one finding.
//
// THE CONJUNCTION IS THE PRECISION CONTROL, and it is the whole design. Measured
// over the complete live corpus (2,405 components with rendered_html and
// locked_at IS NULL, 2026-08-27) each half alone is unshippable:
//
//	exhaustive idiom alone ............. 22 components (legitimate guide copy:
//	                                     "work through it line by line")
//	verb + against + rulebook alone .... 13 components (legitimate imperative,
//	                                     live on the motivating site itself:
//	                                     "Check your loan against the FCA rules")
//	BOTH, same component ...............  3 components — the three planted ones,
//	                                     0 on the other 2,402
//
// RE2 HAS NO BACKREFERENCE, so "rule by clause" matches as readily as "rule by
// rule". That is accepted rather than worked around: any cross-pair of these
// units inside this conjunction is the same overclaim, and there is no RE2 form
// that says "the same word twice".
//
// TWO UNITS ARE DELIBERATELY ABSENT and their absence is pinned by tests.
// `step` — "step by step" is the commonest legitimate idiom in a how-to guide.
// `register` is absent from the RULEBOOK list for the same reason: "records
// verified against the Companies House register" is a live ATTESTED FACT on
// leopardessconsulting.co.uk, i.e. exactly the true, evidence-backed sentence
// this layer must never touch.
const (
	diligenceVerb = `(?:check\w*|verif\w*|audit\w*|validat\w*|cross.?check\w*|review\w*)`
	diligenceBook = `\b(?:handbook|rule.?book|sourcebook|the rules|regulations?|guidance|legislation|statutes?)\b`
	diligenceUnit = `(?:rule|line|section|clause|guide|page|point|word|entry|item|case|claim|figure)`

	diligenceIdiom   = `\b` + diligenceUnit + `\s+by\s+` + diligenceUnit + `\b`
	diligenceChecked = diligenceVerb + `[^.]{0,60}\bagainst\b[^.]{0,60}` + diligenceBook

	// Order A: "…checked against the FCA handbook, rule by rule".
	// Order B: "Rule by rule, our guides are audited against the regulations."
	diligenceAgainstRulebookPattern = diligenceChecked + `[^.]{0,60}` + diligenceIdiom +
		`|` + diligenceIdiom + `[^.]{0,60}` + diligenceChecked
)

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
var practiceHedgeRe = regexp.MustCompile(
	`(?i)\b(?:if|when|whenever|before|once|should|unless|until)\s+` +
		// OPTIONAL INTERVENING SUBJECT, added 2026-08-27 for P6 (bugs_open/414).
		// P1-P5 anchor ON the subject ("we test"), so a conditional sits directly
		// against the match and the bare form was enough. P6 anchors on the VERB,
		// deliberately — moving its anchor earlier would put the negation cue of
		// "our guides have NOT been checked…" outside negatedClaimMatch's
		// backwards window and break suppression that works today. So the subject
		// is skipped here instead: "If we check a guide against the handbook, rule
		// by rule, we say so" is a hypothetical, not a claim.
		//
		// This widens a SUPPRESSOR, which is the direction that can hide a real
		// finding, so the blast radius is stated rather than assumed: it can only
		// change patterns whose match does NOT start at the subject, and P6 is the
		// only one. For P1-P5 the cue already abuts the match, so every sentence
		// they hedge today they still hedge, and no sentence they catch today
		// becomes hedged — pinned by TestPracticeSuppressionsAreObservable and the
		// unchanged must-catch table.
		`(?:(?:we|you|they|our|its|their)\s+(?:[a-z]+\s+)?)?$`)

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
