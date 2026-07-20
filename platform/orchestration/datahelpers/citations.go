// FILE: platform/orchestration/datahelpers/citations.go
//
// Citation evidence for the claims-verification layer (V5 —
// SPEC_V5_researched_citations.md). A `citation` is the fourth evidence source
// kind, alongside `sql`, `artifact` and `attested_by`: a fact whose ground
// truth is an external publication.
//
// The design's load-bearing idea: a citation is verified DETERMINISTICALLY —
// fetch the URL and assert that the stored VERBATIM QUOTE still appears in the
// document's visible text. A string comparison, exactly like the email check
// that started this programme. It kills hallucinated references at acquisition
// time (a fabricated source will not contain the quote) and makes scheduled
// re-verification free (the quote disappearing means the source changed, moved
// or was withdrawn).
//
// This file is the PURE half: parsing, normalisation and matching, with no
// network access, so the part that must never be wrong is testable offline
// with fixture HTML. Fetching lives in the actions package.
//
// Matching is deliberately forgiving about presentation and strict about
// content: publishers reformat pages, so whitespace, unicode punctuation, HTML
// entities and thousands separators are normalised on BOTH sides before
// comparison — but the words and numbers themselves must match exactly.

package datahelpers

import (
	stdhtml "html"
	"regexp"
	"strings"
)

// Citation records where an externally-sourced fact came from and the exact
// sentence relied upon.
type Citation struct {
	Publisher string `json:"publisher"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Published string `json:"published,omitempty"` // as printed by the source, e.g. "2025-06"
	Quote     string `json:"quote"`               // VERBATIM from the source — the load-bearing field
	Accessed  string `json:"accessed"`            // YYYY-MM-DD we last verified it
}

// ParseCitation extracts a citation from a fact's source map
// (fact["source"]["citation"]). Returns nil when the fact has no citation
// source. The error return names what is missing when a citation is present
// but unusable — a citation without a quote or URL cannot be verified and must
// never be treated as evidence.
func ParseCitation(source map[string]interface{}) (*Citation, error) {
	raw, ok := source["citation"].(map[string]interface{})
	if !ok {
		return nil, nil
	}
	c := &Citation{
		Publisher: GetStringField(raw, "publisher", ""),
		Title:     GetStringField(raw, "title", ""),
		URL:       strings.TrimSpace(GetStringField(raw, "url", "")),
		Published: GetStringField(raw, "published", ""),
		Quote:     strings.TrimSpace(GetStringField(raw, "quote", "")),
		Accessed:  GetStringField(raw, "accessed", ""),
	}
	var missing []string
	if c.URL == "" {
		missing = append(missing, "url")
	}
	if c.Quote == "" {
		missing = append(missing, "quote")
	}
	if c.Publisher == "" {
		missing = append(missing, "publisher")
	}
	if len(missing) > 0 {
		return c, &CitationInvalidError{Missing: missing}
	}
	return c, nil
}

// CitationInvalidError names the fields a citation is missing.
type CitationInvalidError struct{ Missing []string }

func (e *CitationInvalidError) Error() string {
	return "citation missing required field(s): " + strings.Join(e.Missing, ", ")
}

// ── Normalisation ───────────────────────────────────────────────────────────

var (
	quoteWSRe = regexp.MustCompile(`\s+`)
	// thousands separators inside numbers: 1,234 / 1,234,567 → 1234 / 1234567
	quoteThousandsRe = regexp.MustCompile(`(\d),(\d{3})`)
)

var quotePunctReplacer = strings.NewReplacer(
	"‘", "'", "’", "'", // curly single quotes
	"“", `"`, "”", `"`, // curly double quotes
	"–", "-", "—", "-", "−", "-", // en/em dash, minus
	" ", " ", " ", " ", " ", " ", // nbsp and thin spaces
	"…", "...", // ellipsis
)

// NormalizeForQuoteMatch reduces text to a form stable across publisher
// reformatting: entities decoded, unicode punctuation folded to ASCII,
// thousands separators removed from numbers, whitespace collapsed, lowercased.
// Applied to BOTH the stored quote and the fetched document, so "411&nbsp;MT"
// matches "411 MT" and "1,234" matches "1234" — but "411" never matches "412".
func NormalizeForQuoteMatch(s string) string {
	s = stdhtml.UnescapeString(s)
	s = quotePunctReplacer.Replace(s)
	for {
		out := quoteThousandsRe.ReplaceAllString(s, "$1$2")
		if out == s {
			break
		}
		s = out
	}
	s = quoteWSRe.ReplaceAllString(s, " ")
	return strings.ToLower(strings.TrimSpace(s))
}

// VisibleTextFromHTML reduces an HTML document to the text a reader sees,
// using the same assertion-text extraction as the claims scans (script, style,
// code and attributes excluded). For non-HTML plain text, pass it straight to
// QuoteFoundInText instead.
func VisibleTextFromHTML(htmlStr string) string {
	return strings.Join(ExtractAssertionText(htmlStr), " ")
}

// QuoteFoundInText reports whether the citation's quote appears in the
// document text, after normalising both sides. The caller supplies visible
// text (see VisibleTextFromHTML).
//
// An empty quote NEVER matches — a citation with nothing to check is not
// evidence, and returning true for it would let an unverifiable fact pose as a
// verified one.
func QuoteFoundInText(quote, docText string) bool {
	q := NormalizeForQuoteMatch(quote)
	if q == "" {
		return false
	}
	return strings.Contains(NormalizeForQuoteMatch(docText), q)
}
