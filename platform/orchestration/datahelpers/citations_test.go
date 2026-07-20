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
