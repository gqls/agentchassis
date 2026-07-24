// FILE: platform/orchestration/datahelpers/citations_test.go
//
// The citation verifier's pure half. The matching rule to protect: forgiving
// about PRESENTATION (entities, unicode punctuation, whitespace, thousands
// separators), strict about CONTENT (words and numbers must match exactly).

package datahelpers

import (
	"strings"
	"testing"
)

func TestNormalizeForQuoteMatch(t *testing.T) {
	cases := map[string]string{
		// entities and nbsp
		"411&nbsp;million&nbsp;tonnes": "411 million tonnes",
		"supply &amp; demand":          "supply & demand",
		// curly punctuation and dashes
		"the market ‘grew’ — strongly": "the market 'grew' - strongly",
		"2019–2024":                    "2019-2024",
		// thousands separators, including nested
		"reached 1,234 sites":    "reached 1234 sites",
		"a 1,234,567 tonne year": "a 1234567 tonne year",
		// whitespace collapse + case
		"  Global   LNG\n\tTrade  ": "global lng trade",
	}
	for in, want := range cases {
		if got := NormalizeForQuoteMatch(in); got != want {
			t.Errorf("NormalizeForQuoteMatch(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestQuoteFoundInText(t *testing.T) {
	doc := VisibleTextFromHTML(`
		<html><head><style>.x{color:red}</style>
		<script>var quote = "global LNG trade reached 999 MT";</script></head>
		<body>
		  <h1>World LNG Report&nbsp;2025</h1>
		  <p>Against expectations, global LNG trade reached
		     411&nbsp;million tonnes in 2024 — a record year.</p>
		  <p>Deliveries rose to 1,234 cargoes.</p>
		  <pre>global LNG trade reached 555 MT (draft table)</pre>
		</body></html>`)

	// Presentation differences must not break a genuine match.
	for _, q := range []string{
		"global LNG trade reached 411 million tonnes in 2024",
		"Global LNG trade reached 411 MILLION TONNES in 2024",
		"reached 411&nbsp;million tonnes in 2024 – a record year", // entity + wrong-dash variant
		"rose to 1234 cargoes",                                    // separator-free form of 1,234
	} {
		if !QuoteFoundInText(q, doc) {
			t.Errorf("quote should match but did not: %q", q)
		}
	}

	// Content differences MUST break the match — this is the anti-hallucination
	// property: a fabricated or altered figure fails verification.
	for _, q := range []string{
		"global LNG trade reached 412 million tonnes in 2024", // wrong number
		"global LNG trade reached 411 million tonnes in 2023", // wrong year
		"global LNG trade reached 999 MT",                     // only exists in <script> — not visible text
		"global LNG trade reached 555 MT",                     // only exists in <pre> — excluded context
		"",                                                    // empty quote is never evidence
	} {
		if QuoteFoundInText(q, doc) {
			t.Errorf("quote should NOT match but did: %q", q)
		}
	}
}

// The markdown-table case (bugs_open/062 layer 3, 2026-07-24): a researcher
// quotes a pricing-table ROW from a scrape's markdown rendering; the verifier
// re-fetches HTML and flattens the same row to space-joined cells. The pipes
// are firecrawl's table syntax — presentation — and must fold away; the
// cells' words and numbers stay strict. The failing quote below is verbatim
// from the first live directory-researcher run that completed end-to-end
// (every candidate rejected on exactly this mismatch).
func TestQuoteFoundInTextMarkdownTableRow(t *testing.T) {
	doc := VisibleTextFromHTML(`
		<html><body><table>
		  <tr><th>Model</th><th>Input</th><th>Cached</th><th>Batch in</th><th>Output</th></tr>
		  <tr><td>gpt-5.6-sol</td><td>$5.00</td><td>$0.50</td><td>$6.25</td><td>$30.00</td></tr>
		  <tr><td>gpt-5.5</td><td>$1.25</td><td>$0.13</td><td>$1.56</td><td>$10.00</td></tr>
		</table></body></html>`)

	// The markdown rendering of the row, exactly as extracted live.
	if !QuoteFoundInText("gpt-5.6-sol | $5.00 | $0.50 | $6.25 | $30.00", doc) {
		t.Error("a markdown table-row quote must match the HTML-flattened row — pipes are presentation")
	}

	// Strict about content: a wrong price in the same row shape must fail.
	if QuoteFoundInText("gpt-5.6-sol | $4.00 | $0.50 | $6.25 | $30.00", doc) {
		t.Error("a table-row quote with an altered price must NOT match")
	}
	// And cells from DIFFERENT rows must not be stitchable into one quote.
	if QuoteFoundInText("gpt-5.6-sol | $1.25", doc) {
		t.Error("cells from different rows must NOT match as one quote")
	}
}

func TestParseCitation(t *testing.T) {
	// Not a citation source at all → nil, nil.
	c, err := ParseCitation(map[string]interface{}{"sql": "SELECT 1"})
	if c != nil || err != nil {
		t.Errorf("non-citation source: want nil,nil got %v,%v", c, err)
	}

	// Complete citation parses.
	c, err = ParseCitation(map[string]interface{}{"citation": map[string]interface{}{
		"publisher": "International Gas Union", "title": "World LNG Report 2025",
		"url": "https://example.org/r", "quote": "reached 411 MT", "accessed": "2026-07-20",
	}})
	if err != nil || c == nil || c.Publisher != "International Gas Union" {
		t.Fatalf("complete citation failed to parse: %v %v", c, err)
	}

	// Missing quote/url must error, naming the fields — an unverifiable
	// citation must never pass as evidence.
	_, err = ParseCitation(map[string]interface{}{"citation": map[string]interface{}{
		"publisher": "IGU",
	}})
	if err == nil || !strings.Contains(err.Error(), "url") || !strings.Contains(err.Error(), "quote") {
		t.Errorf("incomplete citation must name missing fields, got %v", err)
	}
}
