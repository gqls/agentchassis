// FILE: platform/orchestration/datahelpers/claims_attributed.go
//
// The ATTRIBUTED-BUT-UNCITED STATISTIC detector (bugs_open/123 candidate 3).
//
// WHAT IT IS FOR. The observed fabrication that filed 123 was, verbatim:
//
//	"Industry data shows that large language models experience hallucination
//	 rates between 3% and 10% depending on the task."
//
// No source. Invented range. Attributed to "industry data", which is the
// phrasing that makes it read as though it came from somewhere.
//
// WHY NO EXISTING SCAN CATCHES IT, checked pattern by pattern rather than
// assumed. All ten fleet-wide banned claims (claims_global.go:111-215) are
// SELF-accuracy overclaims — "guaranteed accurate", "independently verified",
// "you can rely on us", the completeness-of-exclusion family. A fabricated
// statistic is not a boast about us; it is a fact-shaped sentence with nothing
// behind it, so it matches none of them. And `ScanUnregisteredNumbers` is inert
// here twice over: it needs a register to compare against, and a site-less
// producer has none.
//
// WHY IT IS NOT A NEW FLEET-WIDE BANNED PATTERN. That set's stated bar is
// "false-by-construction for every site we will ever run" (claims_global.go:100).
// "Studies find that 40% of teams ship weekly" can be perfectly true and
// properly cited. This is a HEURISTIC, and heuristics do not belong at blocker
// severity — the same ruling `ScanUnregisteredNumbers` already carries
// (validate_page_content.go:1079-1082, "number extraction has false positives by
// design").
//
// WHAT IT DOES NOT DO, stated rather than implied: it CANNOT tell a true cited
// figure from a fabricated one. A citation is treated as present if a citation
// MARKER is present; verifying that the marker leads anywhere real is V5's job
// (`citations.go`, fetch the URL and assert the verbatim quote survives). This
// detector catches exactly one thing: a figure presented as sourced with no
// source at all.
//
// THE UNIT IS THE DOCUMENT, NOT THE BLOCK, AND THE CORPUS IS WHAT DECIDED THAT.
// This is the load-bearing design decision and it was got wrong twice before the
// measurement corrected it. Run over the complete live corpus (1,130 components,
// 2026-08-03) with a BLOCK-scoped citation test, the detector scored **0 true
// positives and 4 false positives** — every finding was a figure whose source is
// named in the very same sentence or the one before it:
//
//	"But Deloitte's 2026 banking research found that while 87% of banks say…"
//	"Thomson Reuters' Future of Professionals research found…"
//	"Salesforce's State of the Connected Customer research found only 42%…"
//	"The same Dice survey found 30% of tech professionals…"
//
// And an earlier candidate arm — a BARE "N% of <population>" with no cue at all —
// looked like nine true positives until each was read in its full paragraph
// rather than in a 100-character snippet. In context: "66% of people already use
// AI regularly" is preceded by a full citation (University of Melbourne, *Trust,
// attitudes and use of artificial intelligence*, April 2025); "at least 51% of
// people accepted" is the statutory DEFINITION of a representative APR, not a
// statistic at all; "the 78% of customers who…" is anaphoric to a figure cited
// earlier on the page. That arm is not in this file, and its absence is the
// finding.
//
// What both errors share: on long-form editorial copy the citation lives in a
// DIFFERENT BLOCK from the figure, so a block-local test cannot see it. The unit
// that can is the document — and for the producer this detector exists for, the
// document is exactly what it has: content-creator returns one free-standing
// piece of prose per request, with no earlier block anywhere that could be
// carrying the source. That is why the motivating fabrication is detectable and
// the estate's cited copy is not: the 212-word post contained no citation of any
// kind, anywhere in it.

package datahelpers

import "regexp"

// CheckAttributedUncitedStat is the ClaimFinding.Check value this scan emits.
// A FOURTH check value now exists in this layer; nothing routes it to the build
// gate, the claims floor or the discovery sweep (see the CLM-019 register entry).
const CheckAttributedUncitedStat = "attributed_uncited_stat"

// attributionCueRe matches a claim attributed to an UNNAMED evidence corpus.
//
// Deliberately narrow on the subject: a NAMED source ("according to the ONS",
// "the Bank of England reports") is a citation question for V5, not a fabrication
// signal — naming a checkable source is the behaviour this layer wants to
// encourage, so it must never be what trips the detector.
//
// The `(?:\w+\s+){0,3}?` gap is lazy and bounded so "research consistently
// shows" matches while "research into how compilers work shows" does not drag
// an unrelated verb into range.
var attributionCueRe = regexp.MustCompile(`(?i)\b(?:industry (?:data|research|reports?|figures?|benchmarks?|surveys?|analysis)|studies|research|surveys?|analysts?|experts?|the data)\s+(?:\w+\s+){0,3}?(?:show|shows|showing|find|finds|finding|found|suggest|suggests|suggesting|indicate|indicates|indicating|reveal|reveals|revealing|estimate|estimates|estimating|report|reports|reporting|confirm|confirms|agree|agrees|say|says)\b|\baccording to (?:industry|market|research|studies|surveys?|analysts?|experts?|the data)\b`)

// attributedStatRe is the figure that turns an attribution into a statistic.
//
// A PERCENTAGE OR A MULTIPLIER ONLY, never a bare number. This is the whole of
// the year defence, and it was learned from the corpus rather than reasoned:
// "more incidents in the first half of 2026 than in all of 2025" contains two
// numbers, no statistic, and matched the first draft twice. `isExcludedNumber`
// (claims.go:807) solves the same problem for the numeric scan by excluding
// composite tokens; requiring a unit is the stricter and simpler form of the
// same idea for a scan that is looking for a *statistic* specifically.
var attributedStatRe = regexp.MustCompile(`(?i)\d+(?:[.,]\d+)?\s*(?:%|percent\b|per cent\b)|\b\d+(?:\.\d+)?\s*(?:x|times)\b`)

// citationMarkerRe is what makes this a fabrication detector rather than a
// prose-style complaint: a bare or markdown-linked URL, an explicit "source:",
// a bracketed year, "cited", or a bracketed publisher in the estate's own
// news-listing style ("[VentureBeat] …", "[Reuters] …").
//
// The bracketed-publisher arm is measured, not speculative: without it the live
// `latest-news` component on two sites reported a false positive on a sentence
// that names its source in its first four characters.
var citationMarkerRe = regexp.MustCompile(`(?i)https?://|\[[^\]\n]{2,40}\]\([^)\n]+\)|\bsources?\s*:|\bcited\b|\(\s*(?:19|20)\d{2}\s*\)|\[[A-Z][^\]\n]{1,38}\]`)

// namedSourceRe matches an attribution to a NAMED party — a possessive proper
// noun ("Deloitte's research", "Thomson Reuters' survey") or "according to
// <Proper Noun>". Naming a checkable source is the behaviour this whole layer
// wants to encourage, so it must never be what trips the detector.
//
// It is deliberately NOT "any capitalised word": the live corpus is full of
// mid-sentence capitals that are not sources at all ("87% of organisations now
// use AI somewhere in their hiring process"), and treating those as citations
// would silence the detector on exactly the shape it exists for.
var namedSourceRe = regexp.MustCompile(`\b[A-Z][\w.&-]{1,30}['’]s?\s|\baccording to (?:the )?[A-Z]`)

// DocumentCarriesCitations reports whether a document as a whole shows any sign
// of sourcing its figures.
//
// Exported because it is the honest precondition for reading a
// ScanAttributedUncitedStats result, and a caller that scans a fragment without
// asking this question first will over-report — which is the mistake the corpus
// measurement caught (see the file header).
func DocumentCarriesCitations(blocks []string) bool {
	for _, b := range blocks {
		if citationMarkerRe.MatchString(b) || namedSourceRe.MatchString(b) {
			return true
		}
	}
	return false
}

// attributedStatWindowBytes bounds how far after the attribution cue a figure
// still counts as attributed BY it. Long enough for "…shows that large language
// models experience hallucination rates between 3% and 10%" (the motivating
// sentence, 62 bytes from cue to figure), short enough that an unrelated figure
// two sentences later is not swept in.
const attributedStatWindowBytes = 160

// ScanAttributedUncitedStats flags figures presented as externally sourced in a
// document that cites nothing at all. It needs NO evidence base, which is the
// point: it is the only scan in this layer that works for a producer with no
// site (bugs_open/123).
//
// PASS THE WHOLE DOCUMENT'S BLOCKS, not one component's. The first thing it does
// is ask whether the document cites anything anywhere; if it does, it reports
// nothing. That single line is the difference between 0-true/4-false and a
// detector worth having — see the file header for the four sentences that taught
// it.
//
// CALLERS MUST TREAT THESE AS RECORD-AND-ANNOTATE, NEVER AS A BLOCKER. Not a
// style preference: this is a heuristic over prose, the same class as
// ScanUnregisteredNumbers, which is never a blocker for the reason stated at
// validate_page_content.go:1079-1082.
//
// One finding per block. The block is the unit an author can review and rewrite;
// three findings on one sentence is noise, not three problems.
func ScanAttributedUncitedStats(blocks []string) []ClaimFinding {
	// A document that sources anything is not the fabrication shape. This is
	// deliberately generous — it clears a document on ONE citation — because the
	// cost of a false positive here is an author being told their cited work is
	// unsourced, which teaches them to ignore the check.
	if DocumentCarriesCitations(blocks) {
		return nil
	}

	var findings []ClaimFinding
	for _, block := range blocks {
		if block == "" {
			continue
		}
		for _, loc := range attributionCueRe.FindAllStringIndex(block, -1) {
			// The SAME negation implementation as every other scan in this
			// layer, deliberately: CLM-004's anti-drift argument is that there
			// is one matching rule here, not two that diverge.
			//
			// SCANNED FROM THE END OF THE MATCH (loc[1]), NOT THE START, and
			// that is not a detail. This cue spans subject-to-verb with a lazy
			// gap, so a negation lands INSIDE the match: "Industry data does not
			// show that 40% …" matches from "Industry", putting "does not"
			// within the matched span where a backwards scan from loc[0] cannot
			// see it. CLM-017 recorded the same shape for the fleet-wide
			// patterns ("their negation sits INSIDE the match; the guard only
			// reads text BEFORE it") — there it was harmless, here it inverted
			// the verdict on every denial, and three unit tests caught it.
			// Scanning back from the verb keeps the guard clause-local while
			// putting the negation in view.
			if negatedClaimMatch(block, loc[1]) {
				continue
			}
			window := block[loc[1]:minInt(len(block), loc[1]+attributedStatWindowBytes)]
			stat := attributedStatRe.FindString(window)
			if stat == "" {
				continue
			}
			findings = append(findings, ClaimFinding{
				Check:   CheckAttributedUncitedStat,
				Matched: block[loc[0]:loc[1]] + " … " + stat,
				// The reason must describe the check that actually ran. An
				// earlier draft said "no citation in the same block" while the
				// scan had already become document-scoped — a message
				// asserting something the code does not do is bugs_closed/133's
				// shape, and it would send an author looking in one paragraph
				// for a problem defined over the whole document.
				Reason:      "figure attributed to unnamed evidence, and this document cites no source anywhere — name the source, or drop the figure",
				Snippet:     claimSnippet(block, loc[0], loc[1]),
				Occurrences: 1,
			})
			break
		}
	}
	return findings
}
