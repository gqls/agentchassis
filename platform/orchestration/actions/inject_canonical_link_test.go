// FILE: platform/orchestration/actions/inject_canonical_link_test.go
//
// Canonical + meta-description at assembly (portfolio_positioning seam 2,
// HANDOFF_2026-08-02). Measured 2026-08-02: zero canonicals fleet-wide
// (lendzy 0/15 on an any-attribute-order check — nothing on any platform path
// emitted one), and every stored head ships a blank
// `<meta name="description" content="">` on pages with no meta description.
// assemblePage is the single live assembly path (page-build-handler deploys
// through the page-rerender agent; no live agent uses assemble_page), so these
// two helpers close both gaps for builds and rerenders alike.
package actions

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const testStoredHead = `<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Loan rules, plainly</title>
    <meta name="description" content="">
    <link rel="stylesheet" href="/assets/css/styles.css">
</head>`

func canonicalTestPage() *PageInfo {
	return &PageInfo{
		ID:     uuid.New(),
		Name:   "about",
		Title:  "About",
		URL:    "/about.html",
		Domain: "lendzy.co.uk",
	}
}

func TestInjectCanonicalLinkEmitsPageIdentity(t *testing.T) {
	out := injectCanonicalLink(testStoredHead, canonicalTestPage(), zap.NewNop())
	want := `<link rel="canonical" href="https://lendzy.co.uk/about.html">`
	if !strings.Contains(out, want) {
		t.Fatalf("canonical missing or wrong.\nwant substring: %s\ngot:\n%s", want, out)
	}
	if idx := strings.Index(out, want); idx > strings.Index(out, "</head>") {
		t.Fatalf("canonical emitted outside <head>:\n%s", out)
	}
	// Identity must be byte-identical to what injectPageJSONLD asserts.
	jsonld := injectPageJSONLD(testStoredHead, canonicalTestPage(), zap.NewNop())
	if !strings.Contains(jsonld, `"@id":"https://lendzy.co.uk/about.html"`) {
		t.Fatalf("test premise broken: JSON-LD identity differs from canonical identity:\n%s", jsonld)
	}
}

func TestInjectCanonicalLinkIsIdempotentAndRespectsExisting(t *testing.T) {
	logger := zap.NewNop()

	once := injectCanonicalLink(testStoredHead, canonicalTestPage(), logger)
	twice := injectCanonicalLink(once, canonicalTestPage(), logger)
	if once != twice {
		t.Fatalf("second injection changed the head:\n%s", twice)
	}

	// A head already declaring a canonical — in either attribute order — is
	// left alone entirely.
	reversed := strings.Replace(testStoredHead, "<title>",
		`<link href="https://elsewhere.example/x" rel="canonical">`+"\n    <title>", 1)
	if out := injectCanonicalLink(reversed, canonicalTestPage(), logger); out != reversed {
		t.Fatalf("existing canonical (attribute-order-reversed) was not respected:\n%s", out)
	}
}

func TestInjectCanonicalLinkSkipsWhatItCannotAssertTruthfully(t *testing.T) {
	logger := zap.NewNop()

	for name, mutate := range map[string]func(*PageInfo){
		"no domain":         func(p *PageInfo) { p.Domain = "" },
		"empty url":         func(p *PageInfo) { p.URL = "" },
		"non-rooted url":    func(p *PageInfo) { p.URL = "about.html" },
		"fragment url":      func(p *PageInfo) { p.URL = "/tools.html#audience-check" },
		"querystringed url": func(p *PageInfo) { p.URL = "/tools.html?tab=1" },
	} {
		p := canonicalTestPage()
		mutate(p)
		if out := injectCanonicalLink(testStoredHead, p, logger); out != testStoredHead {
			t.Fatalf("%s: expected named skip (unchanged head), got:\n%s", name, out)
		}
	}
	if out := injectCanonicalLink("", canonicalTestPage(), logger); out != "" {
		t.Fatalf("empty head must stay empty, got: %s", out)
	}
	if out := injectCanonicalLink(testStoredHead, nil, logger); out != testStoredHead {
		t.Fatalf("nil page must be a no-op")
	}
}

func TestSpliceMetaDescriptionFillsTheBlankTag(t *testing.T) {
	out := spliceMetaDescription(testStoredHead, `Plain answers on UK loan rules — free, independent.`)
	want := `<meta name="description" content="Plain answers on UK loan rules — free, independent.">`
	if !strings.Contains(out, want) {
		t.Fatalf("description not filled.\nwant: %s\ngot:\n%s", want, out)
	}
	if strings.Contains(out, `content=""`) {
		t.Fatalf("blank tag survived the fill:\n%s", out)
	}
}

func TestSpliceMetaDescriptionEscapesAttributeBreakers(t *testing.T) {
	out := spliceMetaDescription(testStoredHead, `Loans: "cap" rules & fees`)
	want := `content="Loans: &quot;cap&quot; rules &amp; fees">`
	if !strings.Contains(out, want) {
		t.Fatalf("attribute escaping missing.\nwant: %s\ngot:\n%s", want, out)
	}
}

func TestSpliceMetaDescriptionDropsTheBlankTagWhenPageHasNone(t *testing.T) {
	out := spliceMetaDescription(testStoredHead, "")
	if strings.Contains(out, "description") {
		t.Fatalf("blank description tag must be REMOVED, not shipped (correct-or-absent):\n%s", out)
	}
	// Removal must take the whole indented line, not leave a blank one.
	if strings.Contains(out, "\n\n") {
		t.Fatalf("removal left an empty line behind:\n%s", out)
	}
	// Neighbours untouched.
	for _, keep := range []string{"<title>Loan rules, plainly</title>", `charset="UTF-8"`, "styles.css"} {
		if !strings.Contains(out, keep) {
			t.Fatalf("removal damaged a neighbour (%s):\n%s", keep, out)
		}
	}
}

func TestSpliceMetaDescriptionLegacyFallbackStillFires(t *testing.T) {
	// A head with no name="description" tag but some OTHER tag ending in
	// content="" — the shape the pre-change code filled (it replaced the FIRST
	// empty content attribute, whatever tag owned it). The targeted replace
	// misses; the legacy fallback must preserve that behaviour, warts
	// included. NOTE an attribute-order-reversed description tag
	// (`<meta content="" name="description">`) was NEVER covered — the legacy
	// literal requires `content="">` at tag end — so it is deliberately not
	// asserted here.
	exotic := strings.Replace(testStoredHead,
		`<meta name="description" content="">`,
		`<meta property="og:description" content="">`, 1)
	out := spliceMetaDescription(exotic, "Filled anyway")
	if !strings.Contains(out, `content="Filled anyway"`) {
		t.Fatalf("legacy fallback did not fire:\n%s", out)
	}
}
