// FILE: platform/orchestration/actions/component_fallback_guard_test.go
//
// RFC_009 option B. Three things are proved here, and the second and third are the
// ones that make the first mean anything:
//
//  1. the rule agrees with the SHARED FIXTURE, which the Python lint also reads —
//     so the two implementations cannot drift silently;
//  2. it FIRES on the case that motivated it, recovered verbatim from the pre-fix
//     contact-info template — because "quiet on the corpus" is also what an inert
//     guard scores;
//  3. it is QUIET on the whole recorded write history — 347 writes, 0 findings,
//     measured 2026-08-03 — because a guard that refuses good work gets switched
//     off, and then it protects nothing.

package actions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fallbackFixture struct {
	Cases []struct {
		Name     string  `json:"name"`
		Template string  `json:"template"`
		Expect   *string `json:"expect"`
	} `json:"cases"`
}

// TestFabricatedFallbackMatchesSharedFixture is the parity pin. The same file is
// read by `check_placeholder_fallbacks.py --selftest`; if you change a pattern in
// one implementation and not the other, one of the two fails.
func TestFabricatedFallbackMatchesSharedFixture(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "component_fallback_fixtures.json"))
	if err != nil {
		t.Fatalf("reading the shared fixture: %v", err)
	}
	var fx fallbackFixture
	if err := json.Unmarshal(raw, &fx); err != nil {
		t.Fatalf("parsing the shared fixture: %v", err)
	}
	if len(fx.Cases) < 20 {
		t.Fatalf("the shared fixture has shrunk to %d cases — it is the only thing "+
			"pinning this rule to the Python lint; do not thin it", len(fx.Cases))
	}

	var fired, quiet int
	for _, c := range fx.Cases {
		got := fabricatedFallbacks(c.Template)
		switch {
		case c.Expect == nil:
			quiet++
			if len(got) != 0 {
				t.Errorf("%s\n  template: %s\n  want: allowed (a legitimate label)\n  got:  refused as %q",
					c.Name, c.Template, got[0].Shape)
			}
		default:
			fired++
			if len(got) == 0 {
				t.Errorf("%s\n  template: %s\n  want: refused as %q\n  got:  allowed",
					c.Name, c.Template, *c.Expect)
				continue
			}
			if got[0].Shape != *c.Expect {
				t.Errorf("%s\n  template: %s\n  want shape %q, got %q",
					c.Name, c.Template, *c.Expect, got[0].Shape)
			}
		}
	}
	// Both arms must be exercised. A fixture that only ever expects "allowed" would
	// pass against a guard that never fires.
	if fired == 0 || quiet == 0 {
		t.Fatalf("the fixture must exercise BOTH arms; fired=%d quiet=%d", fired, quiet)
	}
	t.Logf("shared fixture: %d must-refuse, %d must-allow", fired, quiet)
}

// contactInfoFabricatingTemplate is the pre-fix contact-info template, verbatim.
// FROZEN — it is the defect, not a maintained copy. Recovered from the before-image
// migration 287 wrote to migration_backups, which is the only reason it still
// exists: the live row was repaired, so the corpus this guard was calibrated
// against no longer contains the very bug it was written for.
const contactInfoFabricatingTemplate = `<div class="contact-grid">
  <div class="contact-card"><h3>Email</h3>
    <a href="mailto:{{if .email}}{{.email}}{{else}}info@example.com{{end}}">{{if .email}}{{.email}}{{else}}info@example.com{{end}}</a>
  </div>
  <div class="contact-card"><h3>Phone</h3>
    <a href="tel:{{if .phone}}{{.phone}}{{else}}+1234567890{{end}}">{{if .phone_display}}{{.phone_display}}{{else if .phone}}{{.phone}}{{else}}+1 (234) 567-890{{end}}</a>
  </div>
  <div class="contact-card"><h3>Hours</h3>
    <p>{{if .hours}}{{.hours}}{{else}}Monday – Friday, 9am – 6pm{{end}}</p>
  </div>
</div>`

// TestFabricatedFallbackRefusesTheMotivatingCase is the control that separates
// "correct" from "inert". Without it, every other assertion in this file is
// satisfied by a function that returns nil unconditionally.
func TestFabricatedFallbackRefusesTheMotivatingCase(t *testing.T) {
	found := fabricatedFallbacks(contactInfoFabricatingTemplate)
	if len(found) == 0 {
		t.Fatal("the pre-fix contact-info template was NOT refused — this guard cannot " +
			"catch the bug it was written for (bugs_open/140)")
	}

	shapes := map[string]bool{}
	for _, f := range found {
		shapes[f.Shape] = true
	}
	for _, want := range []string{"phone", "opening_hours", "email"} {
		if !shapes[want] {
			t.Errorf("expected the pre-fix template to be refused for %q; got %v", want, found)
		}
	}

	issue := fabricatedFallbackIssue(contactInfoFabricatingTemplate)
	if issue == "" {
		t.Fatal("fabricatedFallbackIssue returned no blocking issue for the known-bad template")
	}
	// The refusal has to tell the author what to do, or it just gets worked around.
	for _, want := range []string{"skip_field", "{{if .field}}", "bugs_open/140"} {
		if !strings.Contains(issue, want) {
			t.Errorf("the refusal message should mention %q so the author knows the fix; got:\n%s", want, issue)
		}
	}
	t.Logf("pre-fix template refused for %d fact(s): %v", len(found), found)
}

// TestFabricatedFallbackAllowsTheRepairedTemplate is the other half of the same
// boundary: the fix that closed bugs_open/140 must PASS this gate. A guard that
// refuses the repaired form would make the bug unfixable.
func TestFabricatedFallbackAllowsTheRepairedTemplate(t *testing.T) {
	repaired := contactInfoShippedTemplate(t) // parsed out of migration 287
	if got := fabricatedFallbacks(repaired); len(got) != 0 {
		t.Errorf("the REPAIRED contact-info template was refused as %v — the gate would "+
			"have blocked the fix for the bug it exists to prevent", got)
	}
}

// TestFabricatedFallbackQuietOnLegitimateShapes guards the exclusions that each
// cost something to learn, in the guard's own package rather than only in the
// fixture — so deleting a fixture case cannot silently remove the protection.
func TestFabricatedFallbackQuietOnLegitimateShapes(t *testing.T) {
	for _, tc := range []struct{ name, tpl string }{
		{"a constant rendered two ways invents nothing",
			`{{if .u}}<a href="{{.u}}">acme.co.uk</a>{{else}}acme.co.uk{{end}}`},
		{"an honest non-claim states no figure",
			`<p>{{if .price}}{{.price}}{{else}}Contact us for pricing{{end}}</p>`},
		{"a destination belongs to check_cta_gates.py",
			`<a href="{{if .url}}{{.url}}{{else}}/contact.html{{end}}">c</a>`},
		{"a gated field with no fallback is the CORRECT shape",
			`<p>{{if .phone}}{{.phone}}{{end}}</p>`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := fabricatedFallbacks(tc.tpl); len(got) != 0 {
				t.Errorf("refused %q as %q — this is a legitimate shape and refusing it "+
					"is how a guard gets switched off", tc.tpl, got[0].Shape)
			}
		})
	}
}

// TestFabricatedFallbackIsActuallyWiredIn closes the failure mode this package has
// a standing landmine about: a helper with no callers looks exactly like a finished
// refactor. Every other test here proves the RULE is correct; none of them would
// fail if the birth gate never called it.
//
// The action itself needs a DB and ActionParams, so the house pattern in
// store_generated_component_guard_test.go is to unit-test the decision logic and
// cover the full reject path in a Tier 2 integration test. That leaves exactly this
// gap, so this asserts the call site at the source level — crude, and decisive.
func TestFabricatedFallbackIsActuallyWiredIn(t *testing.T) {
	src, err := os.ReadFile("store_generated_component_action.go")
	if err != nil {
		t.Fatalf("reading the birth gate: %v", err)
	}
	body := string(src)

	if !strings.Contains(body, "fabricatedFallbackIssue(htmlTemplate)") {
		t.Fatal("StoreGeneratedComponentAction does not call fabricatedFallbackIssue — " +
			"the rule is implemented and tested but nothing invokes it, so a component " +
			"that invents a business fact would still be persisted (bugs_open/140)")
	}
	// ...and that its result reaches the refusal list rather than being computed and
	// dropped, which would be the same defect wearing a call site.
	idx := strings.Index(body, "fabricatedFallbackIssue(htmlTemplate)")
	window := body[idx:min(idx+220, len(body))]
	if !strings.Contains(window, "blockingIssues = append(blockingIssues, issue)") {
		t.Errorf("fabricatedFallbackIssue is called but its result does not reach "+
			"blockingIssues within the call site; found:\n%s", window)
	}
}
