// FILE: platform/orchestration/datahelpers/claims_global.go
//
// The fleet-wide banned-claim set (bugs_open/104, concept register CLM-010).
//
// WHY THIS EXISTS
//   `banned_claims` was per-site only, so a lesson learned on one site could not
//   reach any other and every new site was born unarmed. Measured 2026-07-28:
//   7 of 15 live sites carried a single pattern, and the two most exposed in the
//   estate — vetcomparison.uk (which published fabricated prices for 3,124 named
//   real vet practices) and idea.uk (the only site taking money) — carried none.
//   The decision to defer this was filed with a numeric trigger ("per-site only
//   until two sites have evidence bases") and no watcher; there are now nine.
//
// WHY IT IS NOT UNIONED INTO EvidenceBase AT PARSE TIME
//   `voicetells.go` does exactly that for the sibling voice layer
//   (`globalTellPhrases()` appended in the constructor), and it is the obvious
//   mirror — but EvidenceBase is marshalled BACK to site_specs by
//   refresh_evidence_base_action.go and evidence_citations.go. Seeding
//   eb.BannedClaims at parse time would silently persist the fleet-wide set into
//   every site's stored register through write paths that never intended to touch
//   it — the same trap claims.go already documents for EvidenceFact.Kind. So the
//   global set is held HERE, never inside a parsed EvidenceBase, and joined only
//   at scan time. `globalEvidence` is unexported so it cannot reach a writer.
//
// WHY IT SCANS WITH NO REGISTER AT ALL
//   ParseEvidenceBase returns nil for a site with no evidence_base row, and that
//   nil is a load-bearing opt-in signal several lanes read deliberately (see the
//   headers of validate_page_content_stats.go and check_unverified_claims_stats.go).
//   It is NOT changed here. Instead ScanAllBannedClaims is nil-safe: a site with
//   no register is still protected by the fleet-wide set, which is the whole
//   point — otherwise site sixteen is born unarmed exactly as site fifteen was.
//   The numeric scan stays strictly opt-in; only this half goes fleet-wide.
//
// THE MEASUREMENT THAT SHAPED THE SET, AND THE PATTERN THAT IS DELIBERATELY ABSENT
//   The candidate set (10 patterns, tested on oufe in sql_for_agents/226) was
//   dry-run through cmd/claimscan over the stored rendered_html of ALL 15 live
//   sites — 908 components. It produced 7 findings, and 4 were FALSE POSITIVES,
//   every one firing on a NEGATED sentence:
//
//     "Where manufacturer data has NOT been independently verified, that is
//      stated explicitly."                                    (robot-hands.com)
//     "When a figure CANNOT be independently verified, it is marked as
//      unverified rather than carried forward…"               (robot-hands.com)
//     "…are Spark's own assessment, NOT independently verified."   (vonc.com ×2)
//
//   At severity blocker that fails a whole page build for making precisely the
//   hedged disclosure this layer exists to encourage. All four came from ONE
//   pattern — `(fully|independently|externally|properly)
//   (verified|audited|fact.?checked)`.
//
//   > *** UPDATED 2026-07-29 — THE GUARD NOW EXISTS AND THE PATTERN IS BACK. ***
//   > This paragraph used to end "which is therefore NOT in the set below", with
//   > a DO-NOT-RE-ADD warning. Both are now spent: negatedClaimMatch in claims.go
//   > is the guard it asked for, all four sentences are regression fixtures, and
//   > the pattern is the last entry in the set, BARE. The original reasoning is
//   > kept above because it is why the guard is shaped the way it is: RE2 has no
//   > lookbehind, so the test cannot live in the pattern string. (Re-checked
//   > 2026-07-29: still no negation-guard prior art in the estate.
//   > voicetells.go:212 looks like one and is not — it CHECKS FOR
//   > defining-by-negation as a style tell.)
//   >
//   > AND THE SET NO LONGER SCANS CLEAN, which matters more than the guard.
//   > The 07-28 measurement's headline — "fires exactly once across 908
//   > components" — was an artefact of this pattern's absence. With it armed the
//   > set finds **2 further live overclaims**, both on robot-hands.com, both
//   > asserting its spec data is "independently verified". They were always
//   > there; the four false positives were the reason nobody could see them.
//   > A rebuild of those two components now FAILS until the copy is fixed. That
//   > is the gate working as designed, not a regression — but it is a real
//   > consequence, so it is written here and not only in a summary.
//   >
//   > The guard applies to EVERY banned-claim scan, per-site sets included, so
//   > there is one rule and not two that drift (CLM-004). What it suppresses is
//   > observable — ScanAllBannedClaimsWithSuppressed, printed by cmd/claimscan —
//   > because a silent suppressor and a dead gate look identical from outside.
//
//   With that pattern removed the remaining set fires exactly ONCE across all
//   908 components, on a true positive, and zero false positives on the other
//   907. The owner ruled that one sentence acceptable on 2026-07-28, so the
//   `every … is verified` pattern is narrowed here to claim/content nouns —
//   see its comment.
//
// THE LINE ALL OF THESE DRAW
//   A site may describe what it DOES; it may not claim what that GUARANTEES.
//   "We cite every figure and date it" is a keepable process commitment and must
//   pass. "A claim without a source does not appear here" asserts a completeness
//   nobody can verify, including us, and must not.

package datahelpers

import (
	"regexp"
	"sync"
)

// globalBannedClaims is the fleet-wide banned-claim set — claims NO site in the
// estate may make, about itself, regardless of sector or register. Per-site
// bans layer on top via site_specs (see ScanAllBannedClaims).
//
// The bar for adding one: it must be false-by-construction for every site we
// will ever run, not merely undesirable on the sites we run today. Anything
// narrower belongs in that site's own banned_claims.
//
// Before adding or widening ANY pattern here, dry-run it fleet-wide first:
//
//	go run ./cmd/claimscan -evidence <set.json> -no-global -components <tsv>
//
// and include the NEGATION of the pattern in the pass-list. The set below was
// tested 10-for-10 on fabrication shapes and 13-for-13 on legitimate sentences
// and still shipped a false-positive class, because no test sentence
// contradicted a banned phrase. That is the cheap check.
func globalBannedClaims() []BannedClaim {
	return []BannedClaim{
		{
			Pattern: `(claim|figure|number|statistic)s? without a[^.]{0,40}source[^.]{0,20}(does not|do not|doesn't|don't) appear`,
			Reason:  "completeness-of-exclusion: claims that NOTHING unsourced is published. Unverifiable by anyone, including us.",
		},
		{
			// LATENT FALSE-POSITIVE RISK, recorded 2026-07-28 and widened
			// 2026-07-29 at the council guardian's prompting: the three
			// completeness-of-exclusion patterns (this one and the two either side)
			// are ALL negative constructions — the same shape that produced all
			// four false positives on the excluded (fully|independently|...)
			// pattern. A legitimate sentence like "prices do not appear here
			// because they change daily" matches this one. Zero hits across the
			// whole 908-component live corpus today, so the family ships — but it
			// is where the next false positive will come from, and a negation
			// guard would let all four of these be stated more safely.
			Pattern: `(does not|doesn't|do not|don't) appear here`,
			Reason:  "completeness-of-exclusion, short form.",
		},
		{
			Pattern: `if we (can'?t|cannot)[^.]{0,60}(it |they )?(doesn'?t|don'?t) appear`,
			Reason:  "completeness-of-exclusion, conditional form.",
		},
		{
			// NARROWED from the oufe original `every[^.]{0,30}(is|are)
			// (verified|checked|confirmed|validated)`, which caught
			// leopardessconsulting's "Every component is verified against
			// production." — a claim about their own delivered work, which the
			// owner ruled acceptable on 2026-07-28. Requiring a claim/content
			// noun keeps the shape this layer is for ("every claim on this site
			// is verified") and drops claims about a site's own process output.
			Pattern: `every (claim|figure|number|statistic|statement|fact|price|source|citation|entry|detail)s?[^.]{0,30}(is|are) (verified|checked|confirmed|validated)`,
			Reason:  "verification-of-everything: a claim about outcomes, not process.",
		},
		{
			Pattern: `(you|readers?) can rely on (this|it|us|these|our)`,
			Reason:  "invites reliance — the opposite of the standing posture, and the negligent-misstatement exposure route.",
		},
		{
			// The oufe original carried the literal alternative `oufe` in this
			// list; removed, because a site name cannot be part of a fleet-wide
			// pattern. The generic self-referents cover it.
			Pattern: `(we|our|us|this site|this page|this analysis|this report|everything here|our (analysis|research|reporting|coverage|work)|all of (our|the) (figures|claims|numbers)) (is|are|remains?) (always )?(accurate|authoritative|definitive|error.?free|complete|reliable)\b`,
			Reason:  "self accuracy overclaim. Anchored to self, so 'the statute is authoritative' still passes.",
		},
		{
			Pattern: `guaranteed (accurate|correct|complete|up.?to.?date|reliable)`,
			Reason:  "accuracy guarantee.",
		},
		{
			Pattern: `(this|our) (discipline|method|process|standard) is not a disclaimer`,
			Reason:  "repudiates the caveat. Shipped live on oufe.com 2026-07-26.",
		},
		{
			Pattern: `never (wrong|inaccurate|invents?|fabricates?|makes? (a )?mistakes?)`,
			Reason:  "infallibility claim.",
		},
		{
			// RESTORED 2026-07-29, the day after it was excluded. This is the
			// pattern the header above says must not come back without a
			// negation guard; negatedClaimMatch in claims.go IS that guard, and
			// all four sentences it false-positived on are now regression
			// fixtures (claims_global_test.go) asserting they stay silent.
			//
			// KEPT BARE — and the first attempt to narrow it was WRONG, which is
			// worth more than the pattern. I first shipped this subject-anchored
			// (requiring a claim/content noun + is/are within N chars) to avoid
			// "our accounts are independently audited", a claim about a BUSINESS
			// that may be true. Dry-run over all 908 live components: the
			// anchored form matched **nothing at all** — 0 findings, 0
			// suppressions — while the bare form found **2 real overclaims**
			// (robot-hands.com how-it-works and gripper-catalog, both asserting
			// spec data is "independently verified" when nothing outside this
			// system checks it) and the guard removed all 4 negated sentences.
			// Narrowing by reasoning about a hypothetical false positive made
			// the pattern inert; the measurement is what caught it.
			//
			// KNOWN RESIDUAL, unmeasurable today: "our accounts are
			// independently audited" WOULD be a false positive. The phrase
			// occurs zero times in the estate (checked 2026-07-29: the only live
			// sentences of this whole shape are the six in the dry run, all
			// "independently verified"), so the guard for it would be a
			// mechanism rotting unexercised. If a finance or accountancy site
			// arrives, re-check here first — the cheap check is the corpus grep
			// in the RUNBOOK, not a new code path.
			Pattern: `(fully|independently|externally|properly) (verified|audited|fact.?checked)`,
			Reason:  "external-verification claim: asserts our content was checked by someone outside this system. Nothing does that.",
		},
	}
}

// globalEvidence holds the compiled fleet-wide set. It is a *EvidenceBase purely
// so the scan reuses ScanBannedClaims — ONE implementation of the matching rule,
// shared with the per-site path, rather than a second copy that can drift.
//
// It must never escape this package, never be handed to a caller that marshals
// an EvidenceBase, and never be mutated: ScanBannedClaims only reads
// BannedClaims[i].re, so concurrent scans are safe.
var (
	globalEvidenceOnce sync.Once
	globalEvidenceBase *EvidenceBase
)

func globalEvidence() *EvidenceBase {
	globalEvidenceOnce.Do(func() {
		bans := globalBannedClaims()
		for i := range bans {
			// Same fallback contract as ParseEvidenceBase: a pattern that will
			// not compile degrades to a literal case-insensitive substring
			// rather than silently dropping a ban.
			re, err := regexp.Compile("(?i)" + bans[i].Pattern)
			if err != nil {
				re = regexp.MustCompile("(?i)" + regexp.QuoteMeta(bans[i].Pattern))
			}
			bans[i].re = re
		}
		globalEvidenceBase = &EvidenceBase{BannedClaims: bans}
	})
	return globalEvidenceBase
}

// GlobalBannedClaimCount reports how many fleet-wide patterns are active. For
// operator tooling and tests that assert the set is wired at all — a silently
// empty global set and a working one are otherwise indistinguishable from
// outside.
func GlobalBannedClaimCount() int { return len(globalEvidence().BannedClaims) }

// ScanAllBannedClaims scans assertion blocks against the fleet-wide set PLUS the
// site's own patterns. This is the function every enforcement surface should
// call; eb may be nil, and a nil eb means "this site has no register", not "do
// not scan".
//
// A pattern present in both sets is reported once — a site whose register
// already carries a fleet-wide pattern verbatim (oufe does, via mig 226) must
// not produce two identical blockers for one sentence.
func ScanAllBannedClaims(blocks []string, eb *EvidenceBase) []ClaimFinding {
	findings, _ := ScanAllBannedClaimsWithSuppressed(blocks, eb)
	return findings
}

// ScanAllBannedClaimsWithSuppressed additionally returns the matches the
// negation guard REMOVED — phrases that are banned but were read as negated in
// their own clause ("where a figure has NOT been independently verified…").
//
// Nothing enforces the second list. It exists because a suppressor that leaves
// no trace is the failure mode this estate keeps rediscovering: a silent guard
// and a gate that has stopped matching are indistinguishable from outside, so
// the strongest pattern in the set would go back to being unfalsifiable the
// moment it stopped firing. cmd/claimscan prints it; an operator diffing a dry
// run can see exactly what the guard is doing on real copy.
func ScanAllBannedClaimsWithSuppressed(blocks []string, eb *EvidenceBase) (findings, suppressed []ClaimFinding) {
	findings = dedupeByPattern(
		globalEvidence().ScanBannedClaims(blocks),
		// Nil-safe: ScanBannedClaims returns nil for a nil receiver.
		eb.ScanBannedClaims(blocks),
	)

	// Suppressed = what the unguarded scan would have raised, minus what the
	// guarded scan actually raised. Derived by subtraction rather than by a
	// second guard implementation, so the two can never disagree.
	raised := make(map[string]bool, len(findings))
	for _, f := range findings {
		raised[f.Pattern] = true
	}
	for _, f := range dedupeByPattern(
		globalEvidence().ScanBannedClaimsIgnoringNegation(blocks),
		eb.ScanBannedClaimsIgnoringNegation(blocks),
	) {
		if raised[f.Pattern] {
			continue
		}
		suppressed = append(suppressed, f)
	}
	return findings, suppressed
}

// dedupeByPattern joins the fleet-wide and per-site results, reporting a pattern
// present in both exactly once — a site whose register already carries a
// fleet-wide pattern verbatim (oufe does, via mig 226) must not produce two
// identical blockers for one sentence.
func dedupeByPattern(global, site []ClaimFinding) []ClaimFinding {
	out := global
	seen := make(map[string]bool, len(global))
	for _, f := range global {
		seen[f.Pattern] = true
	}
	for _, f := range site {
		if seen[f.Pattern] {
			continue
		}
		seen[f.Pattern] = true
		out = append(out, f)
	}
	return out
}
