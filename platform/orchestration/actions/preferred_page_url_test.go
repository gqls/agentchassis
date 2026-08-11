package actions

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// preferredPageURL is the single source of the URL a page asserts about itself
// (canonical href + JSON-LD @id/url). These tests pin bugs_open/251's fix and,
// more importantly, its LIMIT: only the site root is normalised, because
// directory URLs 404 on this hosting (measured 2026-08-11 across three live
// domains). A canonical pointing at a 404 is worse than the bug.

func TestPreferredPageURL_RootIsNormalised(t *testing.T) {
	got := preferredPageURL("example.com", "/index.html")
	if got != "https://example.com/" {
		t.Errorf("root: got %q, want https://example.com/", got)
	}
}

// The control bugs_open/251 names: a non-root page keeps its own full path,
// or the normalisation has gone too far.
func TestPreferredPageURL_NonRootKeepsItsPath(t *testing.T) {
	for _, url := range []string{
		"/legal.html",
		"/guides/index.html", // section index: /guides/ 404s live — must NOT normalise
		"/loans/index.html",
		"/indexes.html", // suffix trap: contains "index.html" but is not one
	} {
		got := preferredPageURL("example.com", url)
		want := "https://example.com" + url
		if got != want {
			t.Errorf("non-root %q: got %q, want %q", url, got, want)
		}
	}
}

// Both injectors must assert the SAME url for the same page. This used to be
// two byte-identical literals kept in sync by a comment; now it is one helper,
// and this test is what fails if anyone ever splits them again.
func TestCanonicalAndJSONLDAgreeOnRootURL(t *testing.T) {
	page := &PageInfo{
		ID: uuid.New(), Name: "index", Title: "Loans and mortgages, in one place",
		URL: "/index.html", Domain: "loanandmortgagecalculator.co.uk",
	}
	withCanonical := injectCanonicalLink(headWith(""), page, nil)
	const wantHref = `href="https://loanandmortgagecalculator.co.uk/"`
	if !strings.Contains(withCanonical, wantHref) {
		t.Errorf("canonical does not carry the bare-root href; head: %s", withCanonical)
	}
	if strings.Contains(withCanonical, "index.html") {
		t.Errorf("canonical still names index.html: %s", withCanonical)
	}

	d := extractLD(t, injectPageJSONLD(headWith(""), page, nil))
	want := "https://loanandmortgagecalculator.co.uk/"
	if d["url"] != want || d["@id"] != want {
		t.Errorf("JSON-LD url/@id = %v/%v, want %s", d["url"], d["@id"], want)
	}
}
