// Command contrastscan measures WCAG text contrast on LIVE pages, in a real
// browser, using computed style and the actual painted backdrop.
//
// WHY A BROWSER AND NOT THE STYLESHEET. bugs_open/122 was filed off a regex over
// each site's styles.css, reading `--color-primary` as "the link colour". That
// produced a table of four failing sites which was largely wrong: the link rule
// resolves to a different variable on nearly every site, several of the flagged
// colours were never the rendered colour, and the real defect (call-to-action
// buttons, not body links) was invisible to it. A stylesheet cannot tell you what
// is painted, because it cannot resolve the cascade, and the answer depends on
// ancestors, alpha and gradients. This runs the page instead.
//
// THREE GUARDS, each from a false positive this tool produced before it had them.
// Every one was caught by screenshotting the element and looking at it, which is
// the check to reach for when a number says a live page is broken and nobody has
// complained:
//
//  1. The element's ACTUAL background, not the page's. Comparing everything to
//     the page background flags every header that sits on its own bar. 122's own
//     first audit did this and reported the oufe header failing at 1.03 when it
//     scores 17.40; "fixing" it would have dropped it to 2.6.
//
//  2. A gradient or image backdrop is UNKNOWABLE from computed style, because
//     `background-image` leaves `backgroundColor` transparent. A naive ancestor
//     walk therefore falls straight through a coloured header to the body. This
//     reported vetcomparison's blue header with white nav as 1.05 against
//     near-white. Unknown is now reported as unknown; it is never guessed.
//
//  3. Alpha is not opacity. A layer with alpha > 0 is not opaque, and treating
//     rgba(255,255,255,0.1) as solid white reported a perfectly readable
//     white-on-purple button as 1.00. Every layer is alpha-composited outwards,
//     and the text colour is composited over the result too.
//
// A tool that over-reports on live sites is worse than no tool: it trains people
// to ignore it, and its findings get "fixed" into real regressions.
//
// RELATIONSHIP TO platform/colour.AuditPalette (026 phase 2b, live 2026-07-27).
// These are complementary and neither replaces the other. Do not delete one as a
// duplicate of the other; they audit different layers.
//
//	AuditPalette   reads the COMPOSED palette from
//	               site_specs.resolved_composition.palette_id. DB-only, no
//	               browser, microseconds, and it can run before anything is
//	               deployed. Its own insight is that intent != artefact, so it
//	               deliberately reads the composed row rather than design_intent.
//
//	contrastscan   reads the PAINTED PAGE. Slower (a real browser, seconds per
//	               page) and only works post-deploy.
//
// The gap that makes both necessary: a colour can be legible in the palette and
// illegible on the page, because chrome and component CSS carry hardcoded
// literals that are in no palette. The header CTA rule found on 2026-07-28 is
// exactly that — `color: white` hardcoded in site_components.rendered_html, on a
// button whose background is the site's accent. It is absent from the composed
// palette (checked), so no palette audit at any quality can see it, and it fails
// on five of six live sites.
//
// Rule of thumb: AuditPalette is the cheap gate that should run every build;
// this is the post-deploy witness for everything the palette does not govern.
//
// Usage:
//
//	go run ./cmd/contrastscan https://example.com/ https://example.com/about.html
//	go run ./cmd/contrastscan -selector "a,button" -json https://example.com/
//
// Exits non-zero when any measurable pair fails, so it can gate a deploy.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/mxschmitt/playwright-go"
)

// probe runs in the page. It returns measurable rows plus the elements whose
// backdrop could not be determined, so the caller can distinguish "passes" from
// "not checked" — collapsing those two is how an audit reports false comfort.
const probe = `(selector) => {
  const px = c => { const m = String(c).match(/[\d.]+/g); return m ? m.map(Number) : null; };
  const lum = rgb => {
    const s = rgb.slice(0,3).map(v => { v/=255; return v<=0.03928 ? v/12.92 : Math.pow((v+0.055)/1.055,2.4); });
    return 0.2126*s[0] + 0.7152*s[1] + 0.0722*s[2];
  };
  const ratio = (a,b) => { const l1=Math.max(lum(a),lum(b)), l2=Math.min(lum(a),lum(b)); return (l1+0.05)/(l2+0.05); };
  const over = (fg, bg) => {                       // fg composited over bg
    const a = fg.length > 3 ? fg[3] : 1;
    return [0,1,2].map(i => fg[i]*a + bg[i]*(1-a));
  };
  const backdrop = el => {
    const layers = [];
    let n = el;
    while (n && n !== document.documentElement) {
      const st = getComputedStyle(n);
      if (st.backgroundImage && st.backgroundImage !== 'none') return { c: null, sure: false };
      const c = px(st.backgroundColor);
      if (c) {
        const a = c.length > 3 ? c[3] : 1;
        if (a >= 0.999) {
          let acc = c.slice(0,3);
          for (let i = layers.length - 1; i >= 0; i--) acc = over(layers[i], acc);
          return { c: acc, sure: true };
        }
        if (a > 0) layers.push(c);
      }
      n = n.parentElement;
    }
    const bs = getComputedStyle(document.body);
    if (bs.backgroundImage && bs.backgroundImage !== 'none') return { c: null, sure: false };
    const bc = px(bs.backgroundColor);
    let acc = (bc && (bc.length < 4 || bc[3] >= 0.999)) ? bc.slice(0,3) : [255,255,255];
    for (let i = layers.length - 1; i >= 0; i--) acc = over(layers[i], acc);
    return { c: acc, sure: true };
  };

  const rows = [], unsure = [], seen = new Set();
  for (const el of document.querySelectorAll(selector)) {
    const r = el.getBoundingClientRect();
    if (r.width < 4 || r.height < 4) continue;
    const st = getComputedStyle(el);
    if (st.visibility === 'hidden' || st.display === 'none' || st.opacity === '0') continue;
    const text = (el.textContent || '').trim();
    if (!text) continue;
    const fg = px(st.color);
    if (!fg) continue;
    const b = backdrop(el);
    const label = text.slice(0,46) + ' [' + String(el.className || '').slice(0,44) + ']';
    if (!b.sure) { unsure.push(label); continue; }
    const fgc = over(fg, b.c);
    const size = parseFloat(st.fontSize), bold = parseInt(st.fontWeight,10) >= 700;
    const large = size >= 24 || (size >= 18.66 && bold);
    const need = large ? 3.0 : 4.5;
    const cr = ratio(fgc, b.c);
    const key = st.color + '|' + b.c.map(Math.round).join(',') + '|' + large;
    if (seen.has(key)) continue;
    seen.add(key);
    rows.push({ ratio: +cr.toFixed(2), need, pass: cr >= need, color: st.color,
                bg: 'rgb(' + b.c.map(Math.round).join(',') + ')', text: text.slice(0,46),
                cls: String(el.className || '').slice(0,44), large });
  }
  return { rows, unsure };
}`

type row struct {
	Ratio float64 `json:"ratio"`
	Need  float64 `json:"need"`
	Pass  bool    `json:"pass"`
	Color string  `json:"color"`
	Bg    string  `json:"bg"`
	Text  string  `json:"text"`
	Cls   string  `json:"cls"`
	Large bool    `json:"large"`
}

type pageResult struct {
	URL    string   `json:"url"`
	Rows   []row    `json:"rows"`
	Unsure []string `json:"unsure"`
	Err    string   `json:"error,omitempty"`
}

func main() {
	// DEFAULT SCOPE: every element that renders text, not just interactive ones.
	//
	// This default used to be "a, button, .btn, [role=button]". That is a false-green
	// generator: on 2026-07-28 it reported oufe.com's Thames page clean while a chart
	// eyebrow sat at 1.23 and a chart title at 1.29 (light text on a white card), both
	// plainly unreadable in a screenshot the owner sent back. Links and buttons are a
	// small fraction of the text on a page, and nothing about a contrast defect
	// confines itself to them.
	//
	// An UNDER-scoped audit is worse than an over-reporting one: over-reporting is
	// noisy and gets ignored, but under-reporting is trusted, and its green is used as
	// evidence that a page is fine.
	selector := flag.String("selector",
		"a, button, .btn, [role=button], h1, h2, h3, h4, h5, h6, p, li, span, td, th, "+
			"dt, dd, figcaption, label, summary, blockquote",
		"CSS selector for the elements to measure")
	asJSON := flag.Bool("json", false, "emit JSON instead of a report")
	flag.Parse()
	urls := flag.Args()
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "usage: contrastscan [-selector S] [-json] <url>...")
		os.Exit(2)
	}

	pw, err := playwright.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "playwright: ", err,
			"\n(run playwright.Install once if the driver is missing)")
		os.Exit(2)
	}
	defer pw.Stop()
	browser, err := pw.Chromium.Launch()
	if err != nil {
		fmt.Fprintln(os.Stderr, "chromium: ", err)
		os.Exit(2)
	}
	defer browser.Close()

	results := make([]pageResult, 0, len(urls))
	failures := 0

	for _, url := range urls {
		res := pageResult{URL: url}
		page, err := browser.NewPage(playwright.BrowserNewPageOptions{
			Viewport: &playwright.Size{Width: 1280, Height: 900},
		})
		if err != nil {
			res.Err = err.Error()
			results = append(results, res)
			continue
		}
		if _, err := page.Goto(url, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(45000),
		}); err != nil {
			res.Err = "navigation: " + err.Error()
			results = append(results, res)
			page.Close()
			continue
		}
		v, err := page.Evaluate(probe, *selector)
		page.Close()
		if err != nil {
			res.Err = "probe: " + err.Error()
			results = append(results, res)
			continue
		}
		b, _ := json.Marshal(v)
		var parsed struct {
			Rows   []row    `json:"rows"`
			Unsure []string `json:"unsure"`
		}
		_ = json.Unmarshal(b, &parsed)
		res.Rows, res.Unsure = parsed.Rows, parsed.Unsure
		results = append(results, res)
	}

	for _, res := range results {
		if !*asJSON {
			fmt.Printf("\n%s\n", res.URL)
			if res.Err != "" {
				fmt.Printf("  ERROR: %s\n", res.Err)
				continue
			}
		}
		worst := 999.0
		for _, r := range res.Rows {
			if r.Ratio < worst {
				worst = r.Ratio
			}
			if !r.Pass {
				failures++
				if !*asJSON {
					fmt.Printf("  FAIL %5.2f (needs %.1f)  %s on %s\n        text=%q class=%q\n",
						r.Ratio, r.Need, r.Color, r.Bg, r.Text, r.Cls)
				}
			}
		}
		if !*asJSON {
			if failures == 0 && len(res.Rows) > 0 {
				fmt.Printf("  %d measurable pair(s), all pass (worst %.2f)\n", len(res.Rows), worst)
			}
			if len(res.Unsure) > 0 {
				fmt.Printf("  %d NOT MEASURABLE (gradient/image backdrop, needs a visual check):\n", len(res.Unsure))
				for _, u := range res.Unsure {
					fmt.Printf("      %s\n", u)
				}
			}
		}
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(results)
	}
	if failures > 0 {
		os.Exit(1)
	}
}
