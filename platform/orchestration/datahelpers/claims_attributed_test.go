package datahelpers

import "strings"

import "testing"

// The motivating fabrication, VERBATIM from bugs_open/123. If this ever stops
// firing, the detector has stopped doing the only thing it was built for.
const bug123Fabrication = "Industry data shows that large language models experience hallucination rates between 3% and 10% depending on the task."

func TestBug123FabricationIsDetected(t *testing.T) {
	blocks := SplitPlainAssertionText(
		"Why agent systems need to recover from their own errors\n\n" +
			bug123Fabrication + "\n\n" +
			"Designing for recovery is therefore a first-class concern.\n")

	got := ScanAttributedUncitedStats(blocks)
	if len(got) != 1 {
		t.Fatalf("the bug's own fabrication must produce exactly 1 finding, got %d: %+v", len(got), got)
	}
	if got[0].Check != CheckAttributedUncitedStat {
		t.Errorf("check = %q, want %q", got[0].Check, CheckAttributedUncitedStat)
	}
	if !strings.Contains(got[0].Matched, "3%") && !strings.Contains(got[0].Matched, "10%") {
		t.Errorf("the finding must name the figure it objected to, got Matched=%q", got[0].Matched)
	}
}

// THE VACUOUS-PASS GUARD, and it is the reason this file exists in this shape.
//
// CLM-017's landmine: a test asserting that clean copy produces no findings had
// been passing for a reason unrelated to the guard it was supposed to prove —
// the only pattern that could match had been removed from the set. "0 findings"
// has two causes with opposite fixes: the scan looked and approved, or the scan
// is dead. Over the whole live corpus this detector reports ZERO, and that is
// only good news if it can still fire.
//
// So every must-pass case below is checked against the SAME text with its
// citation removed, and fails loudly if that does not fire.
func TestMustPassCasesAreNotVacuous(t *testing.T) {
	cases := []struct {
		name    string
		cited   string
		uncited string
	}{
		{
			// The exact live sentence that made the first, block-scoped version
			// of this detector score 0 true / 4 false on the corpus.
			name:    "possessive named source in the same sentence",
			cited:   "But Deloitte's 2026 banking research found that while 87% of banks say they could improve.",
			uncited: "But research found that while 87% of banks say they could improve.",
		},
		{
			name:    "bracketed publisher, the estate's news-listing style",
			cited:   "[VentureBeat] Writer released research showing token usage falls by 38% across models.",
			uncited: "Writer released research showing token usage falls by 38% across models.",
		},
		{
			name:    "a URL anywhere in the document",
			cited:   "Studies find that 40% of teams ship weekly.\n\nSee https://example.org/report for the data.",
			uncited: "Studies find that 40% of teams ship weekly.\n\nSee our earlier post for the data.",
		},
		{
			name:    "an explicit source: line",
			cited:   "Surveys suggest 30% of buyers compare on price.\n\nSource: the 2025 buyer panel.",
			uncited: "Surveys suggest 30% of buyers compare on price.\n\nOur own reading of the market.",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ScanAttributedUncitedStats(SplitPlainAssertionText(c.cited)); len(got) != 0 {
				t.Errorf("cited copy must produce no findings, got %d: %+v", len(got), got)
			}
			// The half that makes the half above mean something.
			if got := ScanAttributedUncitedStats(SplitPlainAssertionText(c.uncited)); len(got) == 0 {
				t.Fatalf("VACUOUS PASS: with the citation removed this text still produces no finding, "+
					"so the case above proves nothing about the citation test. text=%q", c.uncited)
			}
		})
	}
}

// The negation rule is SHARED, not reimplemented — CLM-004's anti-drift
// argument is that this layer has one matching rule, and a second one here would
// diverge from it silently.
//
// The cases below are chosen from the cues `negationCueRe` actually carries
// (claims.go:573-577). See TestBareNoIsAKnownResidualOfTheSharedGuard for the
// one this inherits and deliberately does not fix locally.
func TestNegatedAttributionIsSuppressedByTheSharedGuard(t *testing.T) {
	for _, negated := range []string{
		"Industry data does not show that 40% of teams do this.",
		"Studies have never found that 40% of teams do this.",
		"Research doesn't show that 40% of teams do this.",
	} {
		if got := ScanAttributedUncitedStats(SplitPlainAssertionText(negated)); len(got) != 0 {
			t.Errorf("a denied claim must not be a finding: %q gave %+v", negated, got)
		}
	}
	// Non-vacuous: the asserting form of the same sentence must fire, or the
	// cases above prove nothing about negation.
	asserted := "Industry data shows that 40% of teams do this."
	if got := ScanAttributedUncitedStats(SplitPlainAssertionText(asserted)); len(got) != 1 {
		t.Fatalf("VACUOUS PASS: the asserting form produces %d findings, so the negation cases above "+
			"prove nothing about negation", len(got))
	}
}

// A KNOWN RESIDUAL, INHERITED AND PINNED RATHER THAN PATCHED.
//
// "No industry data shows that 40% of teams do this" IS reported. That is not a
// defect in this detector — bare `no` is excluded from `negationCueRe` on
// purpose (CLM-017: it appears as an intensifier, "There are no exceptions:
// every claim is verified", and including it would silently disarm a real
// overclaim). Fixing it HERE would mean a second negation implementation in this
// layer, which is exactly what CLM-004 forbids.
//
// So it is pinned instead: this test fails the day the shared guard learns bare
// `no`, which is the day to delete it and move the sentence up into the test
// above — a deliberate act by whoever changes the shared rule, rather than a
// silent behaviour change discovered later.
func TestBareNoIsAKnownResidualOfTheSharedGuard(t *testing.T) {
	got := ScanAttributedUncitedStats(SplitPlainAssertionText(
		"No industry data shows that 40% of teams do this."))
	if len(got) == 0 {
		t.Fatalf("the shared negation guard now handles bare `no`. That is an IMPROVEMENT, not a " +
			"failure: delete this test and move the sentence into " +
			"TestNegatedAttributionIsSuppressedByTheSharedGuard.")
	}
}

// A bare number is not a statistic. Learned from the corpus: an early candidate
// matched "industry reported more AI-related incidents in the first half of 2026
// than in all of 2025" — twice — because the digits it required were THE YEAR.
func TestYearsAreNotStatistics(t *testing.T) {
	text := "Industry reports show more incidents in the first half of 2026 than in all of 2025."
	if got := ScanAttributedUncitedStats(SplitPlainAssertionText(text)); len(got) != 0 {
		t.Errorf("a year is not a statistic — got %+v", got)
	}
}

func TestCueWithoutAFigureIsNotAFinding(t *testing.T) {
	if got := ScanAttributedUncitedStats(SplitPlainAssertionText("Studies find that users prefer clarity.")); len(got) != 0 {
		t.Errorf("an attribution with no figure is not this detector's business — got %+v", got)
	}
}

// ---------------------------------------------------------------------------
// SplitPlainAssertionText
// ---------------------------------------------------------------------------

// The measurement that made this splitter necessary: ExtractAssertionText is an
// HTML parser and returns ONE fused block for markdown. Pinned as a property of
// the extractor so a future change to it is a deliberate act, in the style of
// claims_stats.go's TestStatCardHTMLSplitsValueFromLabel.
func TestExtractAssertionTextFusesMarkdownAndSplitterDoesNot(t *testing.T) {
	md := "## Why recovery matters\n\n" + bug123Fabrication + "\n\nA second paragraph.\n"

	if fused := ExtractAssertionText(md); len(fused) != 1 {
		t.Fatalf("PREMISE CHANGED: ExtractAssertionText returned %d blocks for markdown, not 1. "+
			"claims_plaintext.go exists because it returns one fused block; re-read that decision. blocks=%q", len(fused), fused)
	}

	blocks := SplitPlainAssertionText(md)
	if len(blocks) != 3 {
		t.Fatalf("splitter must keep the three lines apart, got %d: %q", len(blocks), blocks)
	}
	for _, b := range blocks {
		if strings.Contains(b, "Why recovery matters") && strings.Contains(b, "Industry data") {
			t.Errorf("heading fused to the following sentence — the whole point of this file: %q", b)
		}
	}
}

func TestSplitterDropsCodeAndKeepsLinkURLs(t *testing.T) {
	md := "Here is the finding.\n\n```\nfmt.Println(\"Studies find that 90% of X\")\n```\n\n" +
		"See [the report](https://example.org/r) for 42% of cases.\n" +
		"Inline `code with 50% of things` is not prose.\n"

	blocks := SplitPlainAssertionText(md)
	joined := strings.Join(blocks, " | ")

	if strings.Contains(joined, "fmt.Println") {
		t.Errorf("fenced code must not become assertion text: %q", joined)
	}
	if strings.Contains(joined, "code with 50%") {
		t.Errorf("inline code must not become assertion text: %q", joined)
	}
	// Load-bearing: the URL is the citation evidence the detector reads. Strip
	// it here and every cited figure looks uncited.
	if !strings.Contains(joined, "https://example.org/r") {
		t.Errorf("markdown link URLs must survive — they are the citation signal: %q", joined)
	}
}

// AssertionBlocks routes by format, and an unknown format must take the SAFE
// branch. Under-matching is a missed finding; over-matching fabricates one.
func TestAssertionBlocksRoutesUnknownFormatToPlain(t *testing.T) {
	md := "## Heading\n\nA sentence.\n"
	for _, format := range []string{"", "markdown", "plain_text", "nonsense"} {
		if got := AssertionBlocks(md, format); len(got) != 2 {
			t.Errorf("format %q must use the plain splitter (2 blocks), got %d: %q", format, len(got), got)
		}
	}
	if got := AssertionBlocks("<p>One.</p><p>Two.</p>", "HTML"); len(got) != 2 {
		t.Errorf("html (case-insensitive) must use the HTML extractor, got %d: %q", len(got), got)
	}
}
