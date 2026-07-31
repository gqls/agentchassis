package actions

import (
	"strings"
	"testing"
)

// The interactivity predicate decides whether a rebuild may replace a section.
// A false NEGATIVE silently destroys a hand-built tool; a false POSITIVE freezes
// ordinary content so it can never be improved. Both directions are asserted.

// A real calculator from loancalculator.co.uk: <input> fields plus an inline
// script, and NONE of the original three markers. Before the 2026-07-30 widening
// this returned false, so a page rebuild would have dropped all twelve.
const realCalculatorSection = `
<div class="card">
  <div class="input-grid">
    <div><label for="amount">Amount to Borrow (£)</label>
      <input type="number" id="amount" value="10000"></div>
    <div><label for="interest">Interest Rate (APR %)</label>
      <input type="number" id="interest" value="7.9"></div>
  </div>
  <div class="results-box"><div class="monthly-payment" id="monthly-display">£0.00</div></div>
</div>
<script>
  function calculateLoan() {
    const P = parseFloat(document.getElementById('amount').value);
    document.getElementById('monthly-display').innerText = P;
  }
  calculateLoan();
</script>`

func TestSectionHTMLIsInteractive(t *testing.T) {
	interactive := map[string]string{
		"a real calculator (the regression case)": realCalculatorSection,
		"canvas game":                  `<div class="game-wrap"><canvas id="c"></canvas></div>`,
		"game-container marker":        `<div class="game-container">x</div>`,
		"tool-page marker":             `<section class="tool-page">x</section>`,
		"data-tool marker":             `<div data-tool="scorer">x</div>`,
		"select driven by a script":    `<select id="s"><option>a</option></select><script>document.getElementById('s')</script>`,
		"inline oninput handler":       `<input id="q" oninput="recalc()"><script>function recalc(){}</script>`,
		"onclick button with a script": `<button onclick="go()">Go</button><script>function go(){}</script>`,
	}
	for name, html := range interactive {
		t.Run("interactive/"+name, func(t *testing.T) {
			if !sectionHTMLIsInteractive(html) {
				t.Errorf("NOT recognised as interactive — a rebuild would silently destroy this")
			}
		})
	}

	// Must stay editable: over-matching means the writer and improvement loops can
	// never touch these again, which defeats an evolving site.
	notInteractive := map[string]string{
		"plain prose":                 `<section><h2>How loans work</h2><p>You repay monthly.</p></section>`,
		"prose with a list":           `<div><ul><li>Fixed</li><li>Variable</li></ul></div>`,
		"a form with NO script":       `<form action="/subscribe"><input type="email" name="e"><button>Sign up</button></form>`,
		"a search box with no script": `<div><input type="search" placeholder="Search"></div>`,
		"a script with no controls":   `<div><p>Hello</p></div><script>console.log("analytics")</script>`,
		"an image card":               `<div class="card"><img src="/a.jpg" alt="x"><p>Caption</p></div>`,
	}
	for name, html := range notInteractive {
		t.Run("editable/"+name, func(t *testing.T) {
			if sectionHTMLIsInteractive(html) {
				t.Errorf("wrongly judged interactive — this content would be frozen and could never be improved")
			}
		})
	}
}

// The Go predicate and the SQL predicate are two spellings of one rule. Five
// hand-written copies used to exist; this asserts the surviving two agree, which
// is the drift class that produced idx_swi_dedup vs workItemTerminalStatuses.
func TestInteractivePredicateGoAndSQLAgree(t *testing.T) {
	sql := interactiveHTMLSQL("rendered_html")

	for _, m := range append(append([]string{}, interactiveStructuralMarkers...), interactiveControlMarkers...) {
		if !strings.Contains(sql, "'%"+m+"%'") {
			t.Errorf("marker %q is in the Go predicate but missing from the SQL — the two have drifted", m)
		}
	}
	// The script conjunction must be present, or SQL would match a bare control
	// where Go would not.
	if !strings.Contains(sql, "'%<script%'") {
		t.Error("SQL lost the <script> conjunction, so it would over-match relative to Go")
	}
	// Every marker referenced by SQL must exist in the Go slices.
	for _, tok := range strings.Split(sql, "'%") {
		if i := strings.Index(tok, "%'"); i > 0 {
			m := tok[:i]
			if m == "<script" {
				continue
			}
			found := false
			for _, k := range append(append([]string{}, interactiveStructuralMarkers...), interactiveControlMarkers...) {
				if k == m {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("SQL references marker %q that the Go predicate does not know", m)
			}
		}
	}
	if !strings.Contains(sql, "rendered_html ILIKE") {
		t.Errorf("column not interpolated: %s", sql)
	}
}
