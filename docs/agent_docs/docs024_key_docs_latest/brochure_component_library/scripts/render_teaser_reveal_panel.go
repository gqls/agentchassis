//go:build ignore

// Renders components/teaser-reveal-panel/template.html with sample_data.json
// through html/template exactly as the platform render path does, and asserts
// the behaviours that matter — including the failing branches.
//
// The assertions that earn their keep are the DEGRADED ones: an item with no
// body must produce NO control (the pattern's absence rule — never a dead
// control), and the cliffhanger must be marked in the data, not by an ellipsis,
// so a truncation checker can tell intent from damage.
//
// Usage: go run render_teaser_reveal_panel.go <template.html> <sample_data.json>
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: render_teaser_reveal_panel <template> <data>")
		os.Exit(2)
	}
	tb, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("read template:", err)
		os.Exit(1)
	}
	db, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Println("read data:", err)
		os.Exit(1)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(db, &data); err != nil {
		fmt.Println("parse data:", err)
		os.Exit(1)
	}
	tmpl, err := template.New("c").Parse(string(tb))
	if err != nil {
		fmt.Println("PARSE FAILED:", err)
		os.Exit(1)
	}
	var out strings.Builder
	if err := tmpl.Execute(&out, data); err != nil {
		fmt.Println("EXECUTE FAILED:", err)
		os.Exit(1)
	}
	html := out.String()

	// Assert against the MARKUP, never the whole document. The first version of
	// this harness counted ".trp__card" inside the <style> block and reported
	// four failures against a correct template — a check that cannot tell a CSS
	// rule from an element is measuring the wrong thing.
	markup := html
	if i := strings.Index(markup, "</style>"); i >= 0 {
		markup = markup[i+len("</style>"):]
	}
	if strings.Contains(markup, "<style") {
		fmt.Println("  FAIL style block still present in markup slice")
		os.Exit(1)
	}

	fail := 0
	check := func(name string, ok bool) {
		if ok {
			fmt.Println("  PASS", name)
		} else {
			fmt.Println("  FAIL", name)
			fail++
		}
	}

	fmt.Println("Rendered", len(html), "bytes")

	// The reveal is native: no JS required for the core behaviour.
	check("three cards rendered", strings.Count(markup, "trp__card\"") + strings.Count(markup, "trp__card trp__card--static") == 3)
	check("two openable <details>", strings.Count(markup, "<details class=\"trp__card\"") == 2)
	check("one static card, not a details", strings.Count(markup, "trp__card--static") == 1)

	// Degradation: the bodyless item must carry NO control and NO cliffhanger mark.
	// A missing static card means the degraded branch did not fire at all; that
	// is a failure to report, not a slice to panic on. (Found by mutating the
	// sample so the bodyless item gained a body — the mutant crashed the
	// harness instead of failing it.)
	staticPart := ""
	if staticIdx := strings.Index(markup, "trp__card--static"); staticIdx >= 0 {
		staticPart = markup[staticIdx:]
		if end := strings.Index(staticPart, "</article>"); end > 0 {
			staticPart = staticPart[:end]
		}
	}
	check("degraded branch fired at all", staticPart != "")
	check("bodyless item has no control", !strings.Contains(staticPart, "trp__control"))
	check("bodyless item is not marked as continuing", !strings.Contains(staticPart, "data-continues"))
	check("bodyless item keeps its prose", strings.Contains(staticPart, "no body at all"))

	// The cliffhanger is structural, not typographic.
	check("openable items marked data-continues", strings.Count(markup, "data-continues=\"true\"") == 2)
	check("no ellipsis anywhere", !strings.Contains(markup, "…") && !strings.Contains(markup, "..."))

	// Label fallback.
	check("explicit open_label used", strings.Contains(markup, ">Read the rest</span>"))
	check("default label applied to the item that omitted it", strings.Count(markup, "Read the rest</span>") == 2)

	// Bodies are in the DOM so the claims gate and crawlers can read them.
	check("body text present unconditionally", strings.Contains(markup, "claims gate and a crawler"))

	// Every card is addressable by key.
	check("keys emitted", strings.Count(markup, "data-trp-key=") == 3)
	check("pattern declared", strings.Contains(markup, "data-experience-pattern=\"teaser-detail-deeplink\""))

	// No unrendered template variables.
	check("no unrendered vars", !strings.Contains(html, "{{") && !strings.Contains(html, "ZgotmplZ"))

	// Images are optional per item: 2 of the 3 sample items carry one, the
	// middle item ("default-label") deliberately has none, and the panel must
	// not break either way — a card with no image_url must render NO <img> at
	// all, never a broken empty src.
	check("images rendered for the 2 items that have one", strings.Count(markup, "trp__media") == 2)
	check("alt text is never the hook restated", !strings.Contains(markup, `alt="A short complete first sentence."`))
	check("no empty img src for the item without an image", !strings.Contains(markup, `src=""`))

	if fail > 0 {
		fmt.Printf("\n%d CHECK(S) FAILED\n", fail)
		os.Exit(1)
	}
	fmt.Println("\nAll checks passed")
}
