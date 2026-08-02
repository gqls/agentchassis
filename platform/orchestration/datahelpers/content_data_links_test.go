package datahelpers

import (
	"reflect"
	"testing"
)

// realPages is the index every test resolves against: the shapes that actually
// occur live (a top-level page, a directory index, a nested tool page).
func realPages() PageURLIndex {
	return NewPageURLIndex([]string{
		"/index.html",
		"/about.html",
		"/tools.html",
		"/contact.html",
		"/cases/index.html",
		"/blog/thames-water.html",
		"/tools/tool-llm-cost-calculator.html",
	})
}

func TestIsURLFieldName(t *testing.T) {
	// Every live shape measured in production content_data on 2026-08-02, plus
	// the negatives that must NOT nominate a candidate.
	yes := []string{
		"url", "link_url", "cta_url", "primary_cta_url", "secondary_cta_url",
		"card1_link_url", "cta_link_url", "image_url", "video_urls", "href",
		"nav_link", "LINK_URL",
	}
	for _, n := range yes {
		if !IsURLFieldName(n) {
			t.Errorf("IsURLFieldName(%q) = false, want true", n)
		}
	}
	no := []string{"title", "body", "link_label", "cta_text", "urlish", "curl", "unlinked"}
	for _, n := range no {
		if IsURLFieldName(n) {
			t.Errorf("IsURLFieldName(%q) = true, want false", n)
		}
	}
}

// TestNestedArrayCardLinkIsFound is the bug in one test: info-card-grid keeps
// its destinations at cards[*].link_url, and every field-enumerating mechanism
// the platform had was blind to it. Depth must not hide a link.
func TestNestedArrayCardLinkIsFound(t *testing.T) {
	data := map[string]interface{}{
		"section_title": "Everything you need",
		"cards": []interface{}{
			map[string]interface{}{"title": "About", "link_url": "/about"},          // rewritable
			map[string]interface{}{"title": "Pricing", "link_url": "/pricing"},      // phantom
			map[string]interface{}{"title": "Contact", "link_url": "/contact.html"}, // fine
			map[string]interface{}{"title": "No link"},                              // no field at all
		},
	}

	got := RepairContentDataLinks(data, realPages())
	want := []ContentDataLinkFinding{
		{Path: "cards[0].link_url", Href: "/about", NewHref: "/about.html", Action: LinkRepairRewrite},
		{Path: "cards[1].link_url", Href: "/pricing", Action: ContentDataLinkPhantom},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings mismatch\n got: %+v\nwant: %+v", got, want)
	}

	cards := data["cards"].([]interface{})
	if v := cards[0].(map[string]interface{})["link_url"]; v != "/about.html" {
		t.Errorf("rewrite arm did not write back: card 0 link_url = %v, want /about.html", v)
	}
	// The phantom arm reports and leaves the value alone. Blanking it here would
	// destroy the authored intent in the SOURCE OF TRUTH, which is the asymmetry
	// content_data_links.go's header argues for; if that ever changes, this test
	// is the thing that should have to be edited deliberately.
	if v := cards[1].(map[string]interface{})["link_url"]; v != "/pricing" {
		t.Errorf("phantom arm mutated the source: card 1 link_url = %v, want /pricing untouched", v)
	}
	if v := cards[2].(map[string]interface{})["link_url"]; v != "/contact.html" {
		t.Errorf("a resolving link was perturbed: %v", v)
	}
}

// TestFlatNumberedCardFieldsAreFound covers case-studies-grid: five sibling
// cardN_link_url fields with no cardN_link_label, so the schema-derived pairing
// in ctafields.go cannot pair them and the hardcoded map does not list the
// component. 10 of the 51 live unresolved links are this shape.
func TestFlatNumberedCardFieldsAreFound(t *testing.T) {
	data := map[string]interface{}{
		"card1_title":    "Thames Water",
		"card1_link_url": "/cases/thames-water", // real page lives at /blog/thames-water.html
		"card2_link_url": "/cases",              // directory index — resolves
		"cta_link_url":   "/contact.html",
	}
	got := RepairContentDataLinks(data, realPages())
	want := []ContentDataLinkFinding{
		{Path: "card1_link_url", Href: "/cases/thames-water", Action: ContentDataLinkPhantom},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings mismatch\n got: %+v\nwant: %+v", got, want)
	}
}

// TestValueClassifierKeepsNonPageFieldsOut is the control for the design claim
// that this file needs no exclusion list. Every value below sits under a
// url-named key and must still be ignored — because of what the VALUE is, not
// because the field name was enumerated anywhere.
func TestValueClassifierKeepsNonPageFieldsOut(t *testing.T) {
	data := map[string]interface{}{
		"image_url":  "/images/hero.jpg",               // asset by prefix
		"logo_url":   "/brand/logo.svg",                // asset by extension
		"docs_url":   "https://example.com/docs",       // external
		"live_url":   "//cdn.example.com/x",            // protocol-relative external
		"mail_url":   "mailto:hi@example.com",          // mailto
		"anchor_url": "#approach",                      // in-page anchor
		"link_url":   "",                               // unset optional — the gated template renders nothing
		"video_urls": []interface{}{"https://v/1.mp4"}, // external, inside an array
	}
	before := len(data)
	if got := RepairContentDataLinks(data, realPages()); got != nil {
		t.Fatalf("expected no findings, got %+v", got)
	}
	if len(data) != before {
		t.Fatalf("data was mutated")
	}
	if data["image_url"] != "/images/hero.jpg" || data["link_url"] != "" {
		t.Fatalf("a non-page value was rewritten: %+v", data)
	}
}

// TestRewriteCarriesTheFragment: /tools#privacy is live on idea.uk. A rewrite
// swaps the PATH; dropping the tail would silently retarget the link.
func TestRewriteCarriesTheFragment(t *testing.T) {
	data := map[string]interface{}{"cta_url": "/tools#privacy"}
	got := RepairContentDataLinks(data, realPages())
	if len(got) != 1 || got[0].NewHref != "/tools.html#privacy" {
		t.Fatalf("got %+v, want one rewrite to /tools.html#privacy", got)
	}
	if data["cta_url"] != "/tools.html#privacy" {
		t.Fatalf("write-back = %v", data["cta_url"])
	}
}

// TestRewriteEmitsAStoredURLNeverAnAssembledOne: the index maps a normalised
// path back to the RAW stored pages.url, and that raw value is what must be
// written. bugs_closed/029 was a whole bug about an emitter assembling
// plausible URLs instead of citing real ones.
func TestRewriteEmitsAStoredURLNeverAnAssembledOne(t *testing.T) {
	index := NewPageURLIndex([]string{"/Tools/Guide.HTML"})
	data := map[string]interface{}{"link_url": "/tools/guide"}
	got := RepairContentDataLinks(data, index)
	if len(got) != 1 || got[0].NewHref != "/Tools/Guide.HTML" {
		t.Fatalf("got %+v, want the stored casing /Tools/Guide.HTML", got)
	}
}

// TestEmptyIndexIsANoOp pins the fail-open direction. An empty index means "the
// page set could not be read", never "this site has no pages" — judging against
// it would call every real link a phantom and rewrite nothing correctly.
func TestEmptyIndexIsANoOp(t *testing.T) {
	data := map[string]interface{}{"link_url": "/about"}
	if got := RepairContentDataLinks(data, NewPageURLIndex(nil)); got != nil {
		t.Fatalf("expected no findings against an empty index, got %+v", got)
	}
	if data["link_url"] != "/about" {
		t.Fatalf("data mutated against an empty index: %v", data["link_url"])
	}
}

// TestDeepNestingAndStringArrays: the walk must not stop at one level, and a
// bare string sitting directly in a url-named array inherits that array's name.
func TestDeepNestingAndStringArrays(t *testing.T) {
	data := map[string]interface{}{
		"groups": []interface{}{
			map[string]interface{}{
				"rows": []interface{}{
					map[string]interface{}{"link_url": "/about"},
				},
			},
		},
		"related_urls": []interface{}{"/about", "/nowhere"},
		"labels":       []interface{}{"/about"}, // NOT url-named — never judged
	}
	got := RepairContentDataLinks(data, realPages())
	want := []ContentDataLinkFinding{
		{Path: "groups[0].rows[0].link_url", Href: "/about", NewHref: "/about.html", Action: LinkRepairRewrite},
		{Path: "related_urls[0]", Href: "/about", NewHref: "/about.html", Action: LinkRepairRewrite},
		{Path: "related_urls[1]", Href: "/nowhere", Action: ContentDataLinkPhantom},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings mismatch\n got: %+v\nwant: %+v", got, want)
	}
	if data["labels"].([]interface{})[0] != "/about" {
		t.Fatalf("a non-url-named array was rewritten")
	}
}

// TestFindingOrderIsStable: Go map iteration is randomised, and a durable record
// that reshuffles between two identical runs cannot be diffed against the last
// one. Run repeatedly on a map wide enough for the randomisation to show.
func TestFindingOrderIsStable(t *testing.T) {
	var first []ContentDataLinkFinding
	for i := 0; i < 25; i++ {
		data := map[string]interface{}{
			"a_url": "/x1", "b_url": "/x2", "c_url": "/x3",
			"d_url": "/x4", "e_url": "/x5", "f_url": "/x6",
		}
		got := RepairContentDataLinks(data, realPages())
		if first == nil {
			first = got
			continue
		}
		if !reflect.DeepEqual(got, first) {
			t.Fatalf("ordering is not stable\nrun 0: %+v\nrun %d: %+v", first, i, got)
		}
	}
	if len(first) != 6 {
		t.Fatalf("expected 6 phantom findings, got %d", len(first))
	}
}

func TestCountContentDataLinkFindings(t *testing.T) {
	rewritten, phantom := CountContentDataLinkFindings([]ContentDataLinkFinding{
		{Action: LinkRepairRewrite}, {Action: ContentDataLinkPhantom}, {Action: LinkRepairRewrite},
	})
	if rewritten != 2 || phantom != 1 {
		t.Fatalf("got rewritten=%d phantom=%d, want 2/1", rewritten, phantom)
	}
}
