package actions

// Converts the two colour fixers' floor exemptions from REASONED to MEASURED —
// the experiment single_slot_floors.go's disposition note prescribed: run each
// fixer's actual rendered-row transform over a fixture carrying class
// attributes and assert the component floor's own census (countComponentClasses)
// is unchanged. Two things keep the measurement honest:
//
//   - each transform must be PROVEN TO FIRE on the fixture (a class census that
//     never saw a rewrite measures nothing), and
//   - the census itself must be proven able to DETECT a dropped class in the
//     same run (an instrument that cannot fail cannot pass anything).

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"go.uber.org/zap"
)

// classAttrRe extracts whole class attributes, both quote styles — the identity
// check, stricter than the floor's count (a rewrite that swapped one class
// attribute for a different one would preserve the count and still fail this).
var classAttrRe = regexp.MustCompile(`(?i)\sclass\s*=\s*("[^"]*"|'[^']*')`)

// forcedTextFixture triggers fix_forced_text_colours' rendered-row transform
// (processComponentCSS): rules targeting text elements carry color:#hex
// declarations, no background is painted so the ambient contrast pre-check
// clears against defaultPalette (#333333 / #1a1a2e on #ffffff), and the .card a
// rule must SURVIVE (links keep explicit colour by design).
const forcedTextFixture = `<section class="tool-panel calculator" data-component="tool">
<style>
.card h2 { color: #101820; font-size: 2rem; }
.card p { color: #202830; line-height: 1.5; }
.card a { color: #0f3460; text-decoration: underline; }
</style>
<div class="card">
<h2 class="card-title">Quick check</h2>
<p class="lead">Some copy with a <a href="/x">link</a>.</p>
<div class="input-grid">
<label class="field-label">Amount</label>
<input class='field-input' type="number">
</div>
<button class="btn btn-primary">Calculate</button>
</div>
</section>`

// hardcodedFixture triggers checks.ReplaceHardcodedColors: a dark single
// background, a dark background-color, and an angled two-stop gradient, all
// inside the <style> block (the transform's stated scope).
const hardcodedFixture = `<section class="tool-panel dark-band" data-component="tool">
<style>
.tool-panel { background: #16213e; padding: 2rem; }
.tool-panel .band { background-color: #0f3460; }
.tool-panel .hero { background: linear-gradient(135deg, #1a1a2e, #0f3460); }
</style>
<div class="card">
<h2 class="card-title">Panel</h2>
<p class="lead">Copy.</p>
<button class="btn btn-primary">Go</button>
</div>
</section>`

func TestColourFixers_PreserveClassAttributes(t *testing.T) {
	t.Run("forced_text_colours", func(t *testing.T) {
		before := forcedTextFixture
		if got := countComponentClasses(before); got != 8 {
			t.Fatalf("fixture census = %d class attributes, expected 8 — fixture or census changed, re-derive before trusting this test", got)
		}

		result := processComponentCSS(before, "test-card", false, defaultPalette(), 4.5, zap.NewNop())

		// Controls: the transform FIRED, on exactly what it should touch.
		if !result.changed {
			t.Fatal("transform did not fire — the fixture no longer exercises processComponentCSS, so the class measurement below would be vacuous")
		}
		if result.skippedContrast {
			t.Fatal("contrast pre-check skipped the strip — the fixture no longer clears minContrast against defaultPalette")
		}
		if result.colorsRemoved != 2 {
			t.Fatalf("colorsRemoved = %d, want 2 (the h2 and p declarations)", result.colorsRemoved)
		}
		if strings.Contains(result.html, "#101820") || strings.Contains(result.html, "#202830") {
			t.Fatal("a text-element colour declaration survived the strip")
		}
		if !strings.Contains(result.html, "#0f3460") {
			t.Fatal("the .card a rule's colour was removed — links must keep explicit colour, and the fixture no longer proves the transform is selective")
		}

		// The measurement: the floor's own census, then strict identity.
		if b, a := countComponentClasses(before), countComponentClasses(result.html); a != b {
			t.Fatalf("countComponentClasses changed %d -> %d — the floor exemption's premise is false", b, a)
		}
		if b, a := classAttrRe.FindAllString(before, -1), classAttrRe.FindAllString(result.html, -1); !reflect.DeepEqual(b, a) {
			t.Fatalf("class attribute list changed:\nbefore: %v\nafter:  %v", b, a)
		}
	})

	t.Run("hardcoded_colours", func(t *testing.T) {
		before := hardcodedFixture
		if got := countComponentClasses(before); got != 5 {
			t.Fatalf("fixture census = %d class attributes, expected 5 — fixture or census changed, re-derive before trusting this test", got)
		}

		after := checks.ReplaceHardcodedColors(before)

		// Controls: all three rewrite patterns fired.
		if after == before {
			t.Fatal("transform did not fire — the fixture no longer exercises ReplaceHardcodedColors, so the class measurement below would be vacuous")
		}
		if n := strings.Count(after, "var(--color-primary)"); n != 3 {
			t.Fatalf("var(--color-primary) appears %d times, want 3 (single background, background-color, gradient first stop)", n)
		}
		if !strings.Contains(after, "var(--color-secondary)") {
			t.Fatal("gradient second stop was not rewritten")
		}
		if regexp.MustCompile(`background(-color)?:\s*#`).MatchString(after) {
			t.Fatal("a hardcoded background hex survived the rewrite")
		}

		if b, a := countComponentClasses(before), countComponentClasses(after); a != b {
			t.Fatalf("countComponentClasses changed %d -> %d — the floor exemption's premise is false", b, a)
		}
		if b, a := classAttrRe.FindAllString(before, -1), classAttrRe.FindAllString(after, -1); !reflect.DeepEqual(b, a) {
			t.Fatalf("class attribute list changed:\nbefore: %v\nafter:  %v", b, a)
		}
	})
}

// TestComponentClassCensus_DetectsRemovedClass proves the instrument the tests
// above rely on can actually fail: a rewrite that drops one class attribute
// must move the census. Without this, "census unchanged" would also pass if
// countComponentClasses matched nothing at all.
func TestComponentClassCensus_DetectsRemovedClass(t *testing.T) {
	mutated := strings.Replace(forcedTextFixture, ` class="lead"`, ``, 1)
	if mutated == forcedTextFixture {
		t.Fatal("mutation did not apply — fixture changed underneath this control")
	}
	b, a := countComponentClasses(forcedTextFixture), countComponentClasses(mutated)
	if a != b-1 {
		t.Fatalf("census did not see the dropped class attribute: %d -> %d, want a decrease of exactly 1", b, a)
	}
	// And the single-quoted form is counted too (the fixture carries one).
	if !strings.Contains(forcedTextFixture, `class='field-input'`) {
		t.Fatal("fixture lost its single-quoted class attribute — the census's quote-style coverage is no longer exercised here")
	}
}
