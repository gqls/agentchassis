// Local proof for migration 686. Renders the OLD (live) article-body template
// and the NEW one against the REAL stored content_data of a live article, and
// asserts the three properties the migration claims. Each check is followed by
// a control that proves the check could have failed.
//
// Faithfulness: article-body's template uses no template functions at all (its
// only interpolation is {{.content}}), so stdlib text/template is a faithful
// stand-in for executeGoTemplate here. That is stated rather than assumed —
// a template WITH funcs would need the real funcmap.
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

const heroBlock = `{{if .hero_image_url}}<figure class="article-body__hero"><img src="{{.hero_image_url}}" alt="{{if .hero_image_alt}}{{.hero_image_alt}}{{end}}" loading="lazy" /></figure>{{end}}`

const heroCSS = `.article-body-section .article-body__hero{margin:0 0 2rem}.article-body-section .article-body__hero img{width:100%;height:auto;display:block;border-radius:var(--radius-md,8px)}`

func render(name, tmpl string, data map[string]interface{}) (string, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

var failures int

func check(n int, desc string, cond bool) {
	if cond {
		fmt.Printf("  PASS  %d. %s\n", n, desc)
	} else {
		fmt.Printf("  FAIL  %d. %s\n", n, desc)
		failures++
	}
}

func main() {
	oldB, err := os.ReadFile("article_body_old.html")
	if err != nil {
		panic(err)
	}
	old := strings.TrimRight(string(oldB), "\n")

	newT := strings.Replace(old, `<div class="container">`, `<div class="container">`+heroBlock, 1)
	newT = strings.Replace(newT, `</style>`, heroCSS+`</style>`, 1)

	realBody, err := os.ReadFile("article_content.txt")
	if err != nil {
		panic(err)
	}
	// The real stored content_data of /blog/womens-boxing-having-a-moment.html:
	// exactly one key, "content". This IS the shape of all 297 instances.
	noHero := map[string]interface{}{"content": string(realBody)}

	oldOut, err := render("old", old, noHero)
	if err != nil {
		panic(err)
	}
	newOut, err := render("new", newT, noHero)
	if err != nil {
		panic(err)
	}

	fmt.Println("CHECK GROUP A — an instance with no hero is unchanged")
	oldMarkup := oldOut[:strings.Index(oldOut, "<style>")]
	newMarkup := newOut[:strings.Index(newOut, "<style>")]
	check(1, "markup (everything before <style>) is byte-identical", oldMarkup == newMarkup)
	check(2, "the ONLY difference in the whole render is the added CSS",
		strings.Replace(newOut, heroCSS, "", 1) == oldOut)
	check(3, "added CSS is exactly 176 bytes", len(heroCSS) == 176)
	check(4, "no <figure> is emitted", !strings.Contains(newOut, "<figure"))
	check(5, "no <no value> leaks", !strings.Contains(newOut, "<no value>"))

	fmt.Println("CHECK GROUP B — hero present, with alt")
	withAlt := map[string]interface{}{
		"content":        "<p>body</p>",
		"hero_image_url": "/assets/images/content-hero-womens-boxing-having-a-moment.jpg",
		"hero_image_alt": "Two boxers touching gloves before a bout",
	}
	outB, err := render("b", newT, withAlt)
	if err != nil {
		panic(err)
	}
	check(6, "figure is emitted", strings.Contains(outB, `<figure class="article-body__hero">`))
	check(7, "src is the resolved asset path",
		strings.Contains(outB, `src="/assets/images/content-hero-womens-boxing-having-a-moment.jpg"`))
	check(8, "alt carries the text", strings.Contains(outB, `alt="Two boxers touching gloves before a bout"`))
	check(9, "figure precedes the body content", strings.Index(outB, "<figure") < strings.Index(outB, "<p>body</p>"))

	fmt.Println("CHECK GROUP C — hero present, alt ABSENT (the normal case for all 297 existing instances)")
	noAlt := map[string]interface{}{
		"content":        "<p>body</p>",
		"hero_image_url": "/assets/images/x.jpg",
	}
	outC, err := render("c", newT, noAlt)
	if err != nil {
		panic(err)
	}
	check(10, "image still renders", strings.Contains(outC, `src="/assets/images/x.jpg"`))
	check(11, "alt is empty, NOT the literal <no value>", strings.Contains(outC, `alt=""`))
	check(12, "no <no value> anywhere in the output", !strings.Contains(outC, "<no value>"))

	fmt.Println("CONTROLS — each proves the check above could have failed")
	// Control for 11/12: the UNGUARDED alt spelling, which is what a careless
	// version of this migration would have shipped.
	unguarded := strings.Replace(newT,
		`alt="{{if .hero_image_alt}}{{.hero_image_alt}}{{end}}"`,
		`alt="{{.hero_image_alt}}"`, 1)
	outCtl, err := render("ctl", unguarded, noAlt)
	if err != nil {
		panic(err)
	}
	check(13, "CONTROL: unguarded alt DOES emit <no value> (so checks 11/12 discriminate)",
		strings.Contains(outCtl, "<no value>"))
	// Control for 1/2: if the hero block were not guarded by {{if}}, the no-hero
	// render would differ in markup, not just CSS.
	unguardedBlock := strings.Replace(newT, heroBlock,
		`<figure class="article-body__hero"><img src="{{.hero_image_url}}"></figure>`, 1)
	outCtl2, err := render("ctl2", unguardedBlock, noHero)
	if err != nil {
		panic(err)
	}
	check(14, "CONTROL: an unguarded block DOES change the no-hero markup (so checks 1/2/4 discriminate)",
		strings.Contains(outCtl2, "<figure"))

	fmt.Printf("\nnew template length: %d (old %d, delta %d)\n", len(newT), len(old), len(newT)-len(old))
	if failures > 0 {
		fmt.Printf("RESULT: %d FAILED\n", failures)
		os.Exit(1)
	}
	fmt.Println("RESULT: 14/14 PASS")
}
