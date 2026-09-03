// FILE: platform/orchestration/datahelpers/citation_rule_span.go
//
// RFC_060 §3d/Q6 (2026-09-02, owner-ruled 2026-09-03 §3f): a citation can be
// substantively true, sourced from the right domain, and still name the
// WRONG rule. Root cause, measured against the live host: the FCA Handbook
// has no rule-level URL — a chapter page (e.g. CONC 6.7) holds dozens of
// rules, so a quote genuinely belonging to CONC 6.7.23 verifies perfectly
// against a citation labelled CONC 6.7.17 — same page, same bytes, same
// fetch. QuoteFoundInText answers "is this quote on THIS PAGE", which is
// correct for a news source and insufficient for a rulebook, where the page
// is a chapter and the citation is a line.
//
// The fix needs no new fetch: the rule id is IN the page text already
// retrieved. The visible text partitions cleanly on the rule's own heading
// — its id, the date it took effect, and its R/G status marker — which is
// what distinguishes a rule's own heading from a MENTION of it elsewhere:
// CONC 6.7.17 itself names the range "CONC 6.7.18 R to CONC 6.7.23 R"
// inline, with no date between the id and the marker, so that inline
// cross-reference does not match this file's heading pattern. Anchoring on
// id+date+marker rather than the bare id is what keeps a naive "find the id,
// take what follows" from landing inside a neighbour's mention and silently
// spanning the wrong text — the trap the lendzy_co_uk lane found measuring
// CONC 5A/6.7 directly (78 rules on CONC 5A, 54 on CONC 6.7, the pattern
// present on all of them).

package datahelpers

import (
	"regexp"
	"strings"
)

// ruleHeadingRe matches an FCA Handbook rule's OWN heading, never a bare
// mention of its id. Reuses regulatoryRulebookCodes (claims.go) so the
// vocabulary of recognised regulators cannot drift between the two checks —
// case-sensitive by the same reasoning fad209b92 already established: a
// real citation is capitalised, and case-insensitivity would let short codes
// like SUP or MAR swallow ordinary prose.
var ruleHeadingRe = regexp.MustCompile(
	`\b(` + regulatoryRulebookCodes + `\s+\d+[A-Z]?\.\d+\.\d+)\s+\d{2}/\d{2}/\d{4}\s+[RG]\b`)

// normaliseRuleID collapses whitespace and uppercases so "CONC  6.7.23",
// "conc 6.7.23" and "CONC 6.7.23" all compare equal. Folding case HERE is
// safe even though the heading match above is deliberately case-sensitive —
// comparing two ALREADY-MATCHED rule ids for equality is a different
// question from deciding whether text is a citation heading at all.
func normaliseRuleID(s string) string {
	return strings.ToUpper(strings.Join(strings.Fields(s), " "))
}

// CitationRuleSpan reports whether quote falls within the region of pageText
// belonging to the rule named by ruleID — the span from that rule's own
// heading up to the NEXT rule's heading, or the end of the text — rather
// than merely appearing anywhere on the page.
//
// applicable reports whether pageText carries the FCA Handbook's
// rule-heading shape at all. false means this page is not chapter-shaped
// (e.g. a legislation.gov.uk section page, one rule per page already) — the
// caller should fall back to ordinary whole-page matching, which is already
// correct there and untouched by this file.
//
// true with found=false means the page DOES have rule headings but NONE
// match ruleID — a real failure, not an absence of the mechanism. The
// caller must NOT fall back to whole-page matching in that case: doing so
// would silently re-enable the exact bug this function exists to catch (a
// quote that verifies on the page but belongs to a DIFFERENT rule).
func CitationRuleSpan(pageText, ruleID, quote string) (found bool, applicable bool) {
	headings := ruleHeadingRe.FindAllStringSubmatchIndex(pageText, -1)
	if len(headings) == 0 {
		return false, false
	}
	target := normaliseRuleID(ruleID)
	if target == "" {
		return false, false
	}
	for i, h := range headings {
		id := pageText[h[2]:h[3]]
		if normaliseRuleID(id) != target {
			continue
		}
		spanStart := h[0]
		spanEnd := len(pageText)
		if i+1 < len(headings) {
			spanEnd = headings[i+1][0]
		}
		return QuoteFoundInText(quote, pageText[spanStart:spanEnd]), true
	}
	return false, true
}
