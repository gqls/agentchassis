package datahelpers

import "testing"

func TestExtractAnchors(t *testing.T) {
	html := `
	<section class="hero">
	  <h1>Spark</h1>
	  <a href="/tools/gauntlet/index.html" class="btn">Enter <strong>the Gauntlet</strong></a>
	  <a class="btn secondary" href="/contact.html">
	     Get
	     Started
	  </a>
	  <a href="">Browse All</a>
	  <img src="/assets/x.png">
	  <a href="mailto:hi@example.com">Email us</a>
	</section>`

	anchors := ExtractAnchors(html)
	if len(anchors) != 4 {
		t.Fatalf("expected 4 anchors, got %d: %+v", len(anchors), anchors)
	}

	want := []Anchor{
		{Href: "/tools/gauntlet/index.html", Text: "Enter the Gauntlet"},
		{Href: "/contact.html", Text: "Get Started"},
		{Href: "", Text: "Browse All"},
		{Href: "mailto:hi@example.com", Text: "Email us"},
	}
	for i, w := range want {
		if anchors[i] != w {
			t.Errorf("anchor %d = %+v, want %+v", i, anchors[i], w)
		}
	}
}

func TestExtractAnchorsAgreesWithExtractHrefs(t *testing.T) {
	// Every anchor ExtractAnchors sees must carry a href ExtractHrefs also
	// sees — the two views of the same HTML must not diverge on hrefs.
	html := `<a href="/a.html">A</a><p><a data-x="1" href='/b.html'><span>B</span></a></p>`
	anchors := ExtractAnchors(html)
	hrefs := ExtractHrefs(html)
	if len(anchors) != len(hrefs) {
		t.Fatalf("anchors %d vs hrefs %d", len(anchors), len(hrefs))
	}
	for i := range anchors {
		if anchors[i].Href != hrefs[i] {
			t.Errorf("href %d: anchors=%q hrefs=%q", i, anchors[i].Href, hrefs[i])
		}
	}
}
