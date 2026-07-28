package actions

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// injectPageJSONLD is the fleet's only source of schema.org markup: measured 2026-07-28,
// ZERO of 14 live sites emitted any application/ld+json. These tests pin the two things
// that make structured data safe rather than harmful — it must never assert something we
// do not know, and it must never be able to break out of the <script> it lives in.

func headWith(inner string) string {
	return "<head><title>x</title>" + inner + "</head>"
}

func extractLD(t *testing.T, head string) map[string]interface{} {
	t.Helper()
	const open = `<script type="application/ld+json">`
	i := strings.Index(head, open)
	if i < 0 {
		t.Fatalf("no ld+json block in head: %s", head)
	}
	rest := head[i+len(open):]
	j := strings.Index(rest, "</script>")
	if j < 0 {
		t.Fatalf("unterminated ld+json block")
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(rest[:j]), &out); err != nil {
		t.Fatalf("emitted JSON-LD does not parse: %v\npayload: %s", err, rest[:j])
	}
	return out
}

func TestInjectPageJSONLD_EmitsWebPage(t *testing.T) {
	page := &PageInfo{
		ID: uuid.New(), Name: "tourbillon", Title: "Tourbillon: qué es",
		URL: "/glosario/tourbillon.html", MetaDesc: "Definición del tourbillon.",
		Domain: "relojistas.com",
	}
	got := injectPageJSONLD(headWith(""), page, nil)
	d := extractLD(t, got)

	if d["@type"] != "WebPage" {
		t.Errorf("@type = %v, want WebPage", d["@type"])
	}
	want := "https://relojistas.com/glosario/tourbillon.html"
	if d["url"] != want || d["@id"] != want {
		t.Errorf("url/@id = %v/%v, want %s", d["url"], d["@id"], want)
	}
	if d["name"] != page.Title {
		t.Errorf("name = %v, want %q", d["name"], page.Title)
	}
	if d["description"] != page.MetaDesc {
		t.Errorf("description = %v, want %q", d["description"], page.MetaDesc)
	}
	part, ok := d["isPartOf"].(map[string]interface{})
	if !ok || part["@type"] != "WebSite" || part["url"] != "https://relojistas.com" {
		t.Errorf("isPartOf = %v, want a WebSite at the origin", d["isPartOf"])
	}
	// It must land inside the head, not after it.
	if strings.Index(got, "application/ld+json") > strings.Index(got, "</head>") {
		t.Error("JSON-LD emitted after </head>")
	}
}

func TestInjectPageJSONLD_OmitsDescriptionWhenAbsent(t *testing.T) {
	page := &PageInfo{Title: "Sin descripción", URL: "/x.html", Domain: "relojistas.com"}
	d := extractLD(t, injectPageJSONLD(headWith(""), page, nil))
	if _, present := d["description"]; present {
		t.Error("description key present when the page has none — an empty description is a claim we cannot make")
	}
}

// The failing branch, which is the whole point: emit NOTHING rather than a block that
// asserts something untrue. A JSON-LD naming a page with no title is machine-readable
// misinformation, and search engines act on it.
func TestInjectPageJSONLD_NoOpWhenNothingTruthfulToSay(t *testing.T) {
	cases := []struct {
		name string
		page *PageInfo
		head string
	}{
		{"no title", &PageInfo{URL: "/x.html", Domain: "relojistas.com"}, headWith("")},
		{"no domain", &PageInfo{Title: "T", URL: "/x.html"}, headWith("")},
		{"nil page", nil, headWith("")},
		{"empty head", &PageInfo{Title: "T", Domain: "d.com"}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := injectPageJSONLD(c.head, c.page, nil); got != c.head {
				t.Errorf("head was modified; want untouched.\n got: %s", got)
			}
		})
	}
}

func TestInjectPageJSONLD_Idempotent(t *testing.T) {
	page := &PageInfo{Title: "T", URL: "/x.html", Domain: "relojistas.com"}
	once := injectPageJSONLD(headWith(""), page, nil)
	twice := injectPageJSONLD(once, page, nil)
	if once != twice {
		t.Error("second call modified the head; a stored head may already carry a block")
	}
	if n := strings.Count(twice, "application/ld+json"); n != 1 {
		t.Errorf("%d ld+json blocks, want 1", n)
	}
}

// A page title is site content and can contain anything. If it could close the script
// element, this helper would be an HTML-injection vector on every page of every site.
func TestInjectPageJSONLD_CannotBreakOutOfTheScriptElement(t *testing.T) {
	page := &PageInfo{
		Title:  `</script><img src=x onerror=alert(1)>`,
		URL:    "/x.html",
		Domain: "relojistas.com",
	}
	got := injectPageJSONLD(headWith(""), page, nil)

	if strings.Contains(got, "</script><img") {
		t.Fatal("raw </script> reached the output — the script element can be closed by page content")
	}
	// And the payload must still be valid JSON carrying the real title.
	d := extractLD(t, got)
	if d["name"] != page.Title {
		t.Errorf("name = %v, want the original title round-tripped", d["name"])
	}
}
