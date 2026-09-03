// FILE: platform/orchestration/datahelpers/citation_rule_span_test.go
//
// RFC_060 §3d/Q6. The fixture below is shaped on the real CONC 6.7 case that
// motivated this file: 6.7.17 is the definitions rule and itself mentions
// the range "6.7.18 R to 6.7.23 R" inline; 6.7.23 is the substantive rule
// with its own heading further down the same page. A naive "find the id,
// take what follows" search would land inside 6.7.17's own text (the
// mention) rather than 6.7.23's heading — this is the trap the anchor on
// id+date+marker exists to avoid.

package datahelpers

import "testing"

const conc67Fixture = `CONC 6.7 Post contract: business practices ` +
	`CONC 6.7.17 01/04/2014 R In CONC 6.7.18 R to CONC 6.7.23 R "refinance" means to extend or vary a high-cost short-term credit agreement or to enter into a further such agreement. ` +
	`CONC 6.7.18 01/04/2014 R A firm must not exercise forbearance in a way that disguises problem debt. ` +
	`CONC 6.7.23 01/04/2014 R A firm must not refinance high-cost short-term credit (other than by exercising forbearance) on more than two occasions.`

func TestCitationRuleSpanFindsQuoteInItsOwnRulesSpan(t *testing.T) {
	found, applicable := CitationRuleSpan(conc67Fixture, "CONC 6.7.23",
		"must not refinance high-cost short-term credit (other than by exercising forbearance) on more than two occasions")
	if !applicable {
		t.Fatalf("expected applicable=true — the fixture carries rule headings")
	}
	if !found {
		t.Fatalf("expected found=true — the quote is genuinely inside CONC 6.7.23's own span")
	}
}

// TestCitationRuleSpanRejectsQuoteAttributedToWrongRule is the load-bearing
// case: the SAME quote (from CONC 6.7.23) checked against a citation
// labelled CONC 6.7.17 — the site's real mistake. Whole-page matching would
// pass this; span matching must not.
func TestCitationRuleSpanRejectsQuoteAttributedToWrongRule(t *testing.T) {
	found, applicable := CitationRuleSpan(conc67Fixture, "CONC 6.7.17",
		"must not refinance high-cost short-term credit (other than by exercising forbearance) on more than two occasions")
	if !applicable {
		t.Fatalf("expected applicable=true")
	}
	if found {
		t.Fatalf("the quote belongs to CONC 6.7.23, not CONC 6.7.17 — span match must reject this, " +
			"which is the entire reason this file exists")
	}
}

// TestCitationRuleSpanDoesNotTreatInlineMentionAsAHeading is the specific
// trap named in the design: CONC 6.7.17 mentions "CONC 6.7.18 R to CONC
// 6.7.23 R" inline, with no date between the id and the R marker. A naive
// split on the bare id would treat that mention as 6.7.18's own heading and
// silently span the wrong text from there.
func TestCitationRuleSpanDoesNotTreatInlineMentionAsAHeading(t *testing.T) {
	headings := ruleHeadingRe.FindAllString(conc67Fixture, -1)
	for _, h := range headings {
		if h == "CONC 6.7.18 R" || h == "CONC 6.7.23 R" {
			t.Fatalf("the inline cross-reference text %q matched the heading pattern — "+
				"it has no date, so this should be structurally impossible", h)
		}
	}
	if len(headings) != 3 {
		t.Fatalf("expected exactly 3 real headings (6.7.17, 6.7.18, 6.7.23), got %d: %v", len(headings), headings)
	}
}

// TestCitationRuleSpanNotApplicableWithoutHeadings is the fallback signal a
// legislation.gov.uk-style page (one rule per page, no heading pattern to
// split on) must produce — the caller falls back to whole-page matching,
// which is already correct there.
func TestCitationRuleSpanNotApplicableWithoutHeadings(t *testing.T) {
	pageText := "Consumer Credit Act 1974 Section 97. The debtor may request a settlement figure."
	_, applicable := CitationRuleSpan(pageText, "Consumer Credit Act 1974 s.97", "settlement figure")
	if applicable {
		t.Fatalf("a page with no rule-heading pattern must report applicable=false, not attempt a span match")
	}
}

// TestCitationRuleSpanNotApplicableWithEmptyRuleID covers a fact with no
// `rule` field on an FCA-Handbook-shaped page: no regression to whole-page
// behaviour for facts this fix does not yet cover.
func TestCitationRuleSpanNotApplicableWithEmptyRuleID(t *testing.T) {
	_, applicable := CitationRuleSpan(conc67Fixture, "", "must not refinance")
	if applicable {
		t.Fatalf("an empty ruleID must report applicable=false — nothing to span-match against")
	}
}

// TestCitationRuleSpanFailsWhenTargetRuleAbsentFromPage is the case that
// must NOT silently fall back to whole-page: headings exist, but none match
// the requested rule (wrong chapter fetched, or a typo'd rule id).
func TestCitationRuleSpanFailsWhenTargetRuleAbsentFromPage(t *testing.T) {
	found, applicable := CitationRuleSpan(conc67Fixture, "CONC 9.9.99", "anything")
	if !applicable {
		t.Fatalf("expected applicable=true — the page has headings, just not this one")
	}
	if found {
		t.Fatalf("a rule id absent from the page must never report found=true")
	}
}

// TestNormaliseRuleIDCollapsesWhitespaceAndCase pins the equality rule
// CitationRuleSpan's own matching depends on.
func TestNormaliseRuleIDCollapsesWhitespaceAndCase(t *testing.T) {
	cases := [][2]string{
		{"CONC 6.7.23", "conc  6.7.23"},
		{"CONC 6.7.23", "CONC\t6.7.23"},
	}
	for _, c := range cases {
		if normaliseRuleID(c[0]) != normaliseRuleID(c[1]) {
			t.Fatalf("normaliseRuleID(%q)=%q != normaliseRuleID(%q)=%q",
				c[0], normaliseRuleID(c[0]), c[1], normaliseRuleID(c[1]))
		}
	}
}
