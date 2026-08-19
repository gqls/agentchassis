// FILE: platform/orchestration/datahelpers/claims_regulated.go
//
// The fleet-wide guard on a site claiming a REGULATED FINANCIAL IDENTITY for
// itself — that it is FCA authorised, holds a firm reference number, or is a
// credit/insurance/mortgage intermediary — together with the one way a site is
// permitted to say so: a recorded attestation naming the firm and its FRN.
//
// WHY IT IS FLEET-WIDE AND NOT PER-SITE. Before this, the entire protection was
// one paragraph of prose in ONE agent's prompt (migration 464, register CGV-032,
// on domain-research-classifier only) plus hand-seeded regexes on exactly TWO of
// ~17 live sites. Measured 2026-08-19: `globalBannedClaims` carried 24 patterns
// and not one mentioned regulation, so **a newly built site was born with no
// protection at all** against "We are FCA authorised and regulated". A prompt
// rule binds the agent that reads the prompt; it does nothing when that agent is
// bypassed, re-run, or its answer edited downstream — and nothing at all for the
// strategist, planner or content writer, none of which carry the rule.
//
// WHY IT IS A SEPARATE FAMILY rather than six more entries in globalBannedClaims:
// because these are the only fleet-wide patterns that a site may legitimately be
// EXEMPT from. A genuinely regulated client must be able to state its status —
// that is a customer we want, not a risk to suppress. Keeping the family separate
// makes the exemption narrow by construction: it can only ever switch off these
// patterns, never the accuracy-overclaim set next door.
//
// THE DEFAULT IS THE SAFE ONE, deliberately (owner ruling 2026-08-02 §2: new
// authority on a shared seam ships as an opt-in field whose unsafe side is OFF).
// No attestation → the patterns apply → the claim is refused. An attestation is
// something a person had to write down, with an FRN in it.
//
// CALIBRATION, run before shipping as `claims_global.go`'s header requires:
//   - 8 of 8 must-catch first-person claims matched;
//   - 10 of 10 must-allow sentences passed, INCLUDING the ones that make this
//     hard — "Nationwide Building Society is authorised and regulated by the
//     Financial Conduct Authority" (third party), the bare string "authorised and
//     regulated by the Financial Conduct Authority" (which our own mortgage-lender
//     directory stores verbatim as a lender's regulator_status), and "We are not
//     authorised or regulated by the FCA" (negated);
//   - dry-run over the WHOLE live corpus, 1,758 components across every site, of
//     which 123 contain regulator language: **2 findings, both on one page**
//     (`loanzy.uk` `about-run1-broker`, status `archived`, serving 404) — which is
//     the preserved evidence of the credit-broker identity the owner cleared, i.e.
//     the exact incident this guard exists for. **Zero hits on anything serving.**
//
// The patterns are all FIRST-PERSON by design. A site describing someone else's
// regulatory status is doing its job — that is what the lender directory is for —
// so a pattern that matched the bare regulator phrase would fire on our own
// best content. That is why none of them do.

package datahelpers

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

// RegulatedAttestation is a site's recorded permission to describe itself as a
// regulated firm. It lives in the site_specs `evidence_base` aspect under
// `regulated`, alongside the facts and banned claims it belongs with.
//
// It is a RECORD, not a flag, and that is the point: "this site may say it is
// regulated" is a claim someone has to stand behind later, so the fields that
// make it auditable — who attested, when, and what they saw — are required, not
// optional. A boolean would be unfalsifiable six months from now.
type RegulatedAttestation struct {
	FirmName    string `json:"firm_name"`             // the authorised entity's registered name
	FRN         string `json:"frn"`                   // Financial Services Register firm reference number
	Regulator   string `json:"regulator,omitempty"`   // FCA (default), PRA, or both
	Permissions string `json:"permissions,omitempty"` // what the firm is actually authorised to do
	AttestedBy  string `json:"attested_by"`           // the person who checked it
	AttestedAt  string `json:"attested_at"`           // RFC3339 or YYYY-MM-DD
	Evidence    string `json:"evidence"`              // what proof was seen (e.g. "email 2026-08-19 + register entry")
}

// frnShape matches a Financial Services Register firm reference number: six or
// seven digits. Deliberately a SHAPE check and not a lookup — this package does
// no network I/O, and a shape check that can be read at a glance is honest about
// what it proves. Verifying the number against the live register is a separate
// job for whoever handles the email, and `Evidence` is where they record it.
var frnShape = regexp.MustCompile(`^\d{6,7}$`)

// Attested reports whether this site has a usable regulated attestation.
//
// Every field it checks is one a human had to supply, and a missing one means
// NOT attested. There is no partial credit: an attestation with no FRN, or one
// nobody signed, is exactly the shape a hallucinating agent would produce, so it
// must not open the gate.
func (r *RegulatedAttestation) Attested() bool {
	if r == nil {
		return false
	}
	if strings.TrimSpace(r.FirmName) == "" ||
		strings.TrimSpace(r.AttestedBy) == "" ||
		strings.TrimSpace(r.Evidence) == "" {
		return false
	}
	if !frnShape.MatchString(strings.TrimSpace(r.FRN)) {
		return false
	}
	return parseAttestedAt(r.AttestedAt) != nil
}

// parseAttestedAt accepts RFC3339 or a bare date. Returns nil when it cannot be
// read — an unparseable date is treated as absent rather than as "now", because
// defaulting an audit timestamp to the present is how an unsigned record starts
// looking signed.
func parseAttestedAt(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02", "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

// RegulatedAttested is the nil-safe form used by the scan path: a site with no
// evidence base at all is, correctly, not attested.
func (eb *EvidenceBase) RegulatedAttested() bool {
	if eb == nil {
		return false
	}
	return eb.Regulated.Attested()
}

// regulatedBannedClaims is the family. All first-person: the subject is the site
// itself. See the calibration note in this file's header before changing any of
// them, and re-run the fleet dry run — `claims_global.go`'s header explains why
// a pattern set that passes its own unit tests can still ship a false-positive
// class.
func regulatedBannedClaims() []BannedClaim {
	const reason = "regulated-identity claim: this site is not a recorded regulated firm. " +
		"Presenting as one is a claim about legal status, not a content choice. If the site IS " +
		"authorised, record a regulated attestation (firm name, FRN, who checked it and what they saw) " +
		"in its evidence_base and this stops being a blocker."
	return []BannedClaim{
		{
			Pattern: `\b(?:we|we're|we are)\s+(?:are\s+)?(?:an?\s+)?(?:fca[\s-]?)?(?:authorised|authorized)\s+and\s+regulated\b`,
			Reason:  reason,
		},
		{
			Pattern: `\b(?:we|our\s+(?:firm|company|business))\s+(?:is|are)\s+(?:authorised|authorized|regulated)\s+by\s+the\s+(?:fca|financial conduct authority|prudential regulation authority|pra)\b`,
			Reason:  reason,
		},
		{
			Pattern: `\b(?:we're|we are)\s+(?:fca|financially)\s+(?:authorised|authorized|regulated)\b`,
			Reason:  reason,
		},
		{
			Pattern: `\bour\s+(?:fca\s+)?(?:firm\s+reference(?:\s+number)?|frn|authorisation\s+number|authorization\s+number)\b`,
			Reason:  reason,
		},
		{
			Pattern: `\b(?:we|our\s+(?:firm|company))\s+(?:hold|holds|have|has)\s+(?:fca\s+)?(?:authorisation|authorization|permission\s+to)\b`,
			Reason:  reason,
		},
		{
			Pattern: `\b(?:we are|we're)\s+(?:an?\s+)?(?:appointed\s+representative|credit\s+broker|insurance\s+broker|mortgage\s+broker|mortgage\s+advis[eo]r)\b`,
			Reason:  reason,
		},
	}
}

var (
	regulatedEvidenceOnce sync.Once
	regulatedEvidenceBase *EvidenceBase
)

// regulatedEvidence holds the compiled family. Same contract as globalEvidence:
// unexported, never mutated, never handed to a writer, and a pattern that fails
// to compile degrades to a literal substring rather than silently vanishing.
func regulatedEvidence() *EvidenceBase {
	regulatedEvidenceOnce.Do(func() {
		bans := regulatedBannedClaims()
		for i := range bans {
			re, err := regexp.Compile("(?i)" + bans[i].Pattern)
			if err != nil {
				re = regexp.MustCompile("(?i)" + regexp.QuoteMeta(bans[i].Pattern))
			}
			bans[i].re = re
		}
		regulatedEvidenceBase = &EvidenceBase{BannedClaims: bans}
	})
	return regulatedEvidenceBase
}

// RegulatedClaimCount reports how many regulated-family patterns are active.
// For operator tooling and for tests that assert the family is wired at all — a
// silently empty family and a working one are indistinguishable from outside,
// which is the failure mode that let the gap this file closes go unnoticed.
func RegulatedClaimCount() int { return len(regulatedEvidence().BannedClaims) }
