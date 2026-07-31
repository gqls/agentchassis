// FILE: internal/adapters/browserrunner/computed_values_test.go
//
// Tests for the computed_values check — the one rung of the ladder that judges
// what a calculator COMPUTES rather than what it contains.
//
// The organising question for every test here is the one the check exists to
// answer: WOULD IT HAVE CAUGHT THE THING. A gate that cannot be shown failing on
// a real defect is not evidence of anything, and this whole check type was
// written because the existing tooling scored an inert page as working. So the
// central case is not the happy path — it is TestComputedValues_CatchesADivisor
// ChangeThatEveryOtherRungPasses, which drives the same fake page through the
// existing checks first and shows them all green on the wrong number.

package browserrunner

import (
	"errors"
	"strings"
	"testing"
)

// primed builds a fake page whose outputs read `texts`, with every named
// selector present exactly once (the shape a real settled tool page has).
func primed(texts map[string]string) *fakePage {
	counts := map[string]int{}
	for sel := range texts {
		counts[sel] = 1
	}
	return &fakePage{status: 200, texts: texts, counts: counts}
}

func TestComputedValues_PassesWhenEveryValueMatches(t *testing.T) {
	page := primed(map[string]string{
		"#monthlyPayment": "£303.44",
		"#totalInterest":  "£923.84",
	})
	ch := criteriaCheck{
		ID: "computes-defaults", Type: "computed_values",
		Steps: []criteriaStep{{Action: "fill", Selector: "#loanAmount", Value: "10000"}},
		ExpectValues: map[string]string{
			"#monthlyPayment": "£303.44",
			"#totalInterest":  "£923.84",
		},
	}

	pass, detail := runComputedValues(page, ch)
	if !pass {
		t.Fatalf("want pass on matching values, got fail: %s", detail)
	}
	if len(page.steps) != 1 || page.steps[0].Selector != "#loanAmount" {
		t.Errorf("the check must drive its input steps before reading: recorded %+v", page.steps)
	}
}

// TestComputedValues_CatchesADivisorChangeThatEveryOtherRungPasses is the
// point of the type. The page is IDENTICAL in every respect the rest of the
// ladder can see — same elements, same count, same money-shaped text, script
// present, nothing collapsed — and one arithmetic constant is wrong.
//
// Wrong-by-one-divisor is not a hypothetical: it is the exact mutation used to
// prove the lane harness could fail at all (12 → 11 months turns £303.44 into
// £330.61 on the same inputs).
func TestComputedValues_CatchesADivisorChangeThatEveryOtherRungPasses(t *testing.T) {
	const wrongByOneDivisor = "£330.61"
	page := primed(map[string]string{"#monthlyPayment": wrongByOneDivisor})

	// First: every rung that exists today, on this same page.
	if n := page.Count("#monthlyPayment"); n == 0 {
		t.Fatal("precondition: the element must be present, or this proves nothing")
	}
	loose := criteriaCheck{ID: "i", Type: "interaction",
		Expect: criteriaExpect{Selector: "#monthlyPayment", TextMatches: `£[\d,]+\.\d\d`}}
	if pass, _ := runInteraction(page, loose); !pass {
		t.Fatal("precondition: a hand-authorable text_matches regexp PASSES the wrong number — " +
			"if this ever fails, the premise of computed_values has changed and the check needs re-justifying")
	}

	// Then: the check under test, on the same page.
	ch := criteriaCheck{ID: "computes-defaults", Type: "computed_values",
		ExpectValues: map[string]string{"#monthlyPayment": "£303.44"}}
	pass, detail := runComputedValues(page, ch)
	if pass {
		t.Fatal("computed_values passed a wrong monthly payment — the check is inert")
	}
	if !strings.Contains(detail, wrongByOneDivisor) || !strings.Contains(detail, "£303.44") {
		t.Errorf("the failure must name BOTH the value read and the value expected, "+
			"or the fixer cannot tell a formula change from a formatting one: %s", detail)
	}
}

// A missing output FAILS rather than skips. At Tier 2 zero matches is the normal
// client-side-render state and must skip; here the page has settled in a real
// browser, so an absent element means the tool stopped reporting a number it
// used to report — which is the defect, not an inability to see it.
func TestComputedValues_MissingOutputElementFails(t *testing.T) {
	page := primed(map[string]string{"#monthlyPayment": "£303.44"})
	ch := criteriaCheck{ID: "c", Type: "computed_values",
		ExpectValues: map[string]string{
			"#monthlyPayment": "£303.44",
			"#totalInterest":  "£923.84", // never primed: absent
		}}

	pass, detail := runComputedValues(page, ch)
	if pass {
		t.Fatal("a vanished output element must fail, not pass on the outputs that remain")
	}
	if !strings.Contains(detail, "#totalInterest") || !strings.Contains(detail, "absent") {
		t.Errorf("failure must name the absent element: %s", detail)
	}
}

// A control the golden was captured with having gone means the comparison
// cannot be made — reporting a pass there would vouch for arithmetic that was
// never driven.
func TestComputedValues_UnusableInputControlFails(t *testing.T) {
	page := primed(map[string]string{"#monthlyPayment": "£303.44"})
	page.stepErr = map[string]error{"#loanAmount": errors.New("no element matches")}
	ch := criteriaCheck{ID: "c", Type: "computed_values",
		Steps:        []criteriaStep{{Action: "fill", Selector: "#loanAmount", Value: "10000"}},
		ExpectValues: map[string]string{"#monthlyPayment": "£303.44"}}

	pass, detail := runComputedValues(page, ch)
	if pass {
		t.Fatal("an input step that could not run must fail the check")
	}
	if !strings.Contains(detail, "#loanAmount") {
		t.Errorf("failure must name the control that could not be driven: %s", detail)
	}
}

// The vacuity guard, and the reason it is here: this check type exists because
// something that asserted nothing was being read as evidence. A computed_values
// check with no expect_values would do exactly that again.
func TestComputedValues_NoExpectValuesFailsRatherThanPassesVacuously(t *testing.T) {
	page := primed(map[string]string{"#monthlyPayment": "£303.44"})
	ch := criteriaCheck{ID: "c", Type: "computed_values"}

	pass, detail := runComputedValues(page, ch)
	if pass {
		t.Fatal("a check asserting nothing must never report a pass")
	}
	if !strings.Contains(detail, "assert") {
		t.Errorf("the detail must say what is wrong with the check itself: %s", detail)
	}
}

// Whitespace is the only latitude. Everything else — currency symbol, digit
// grouping, trailing zeros — is a real change on a money page.
func TestComputedValues_WhitespaceOnlyLatitude(t *testing.T) {
	cases := []struct {
		name, served, expected string
		wantPass               bool
	}{
		{"leading and trailing whitespace", "\n  £303.44\n ", "£303.44", true},
		{"internal whitespace run", "£303.44   per month", "£303.44 per month", true},
		{"currency symbol dropped", "303.44", "£303.44", false},
		{"digit grouping dropped", "£1303.44", "£1,303.44", false},
		{"float artefact", "£303.4400000000001", "£303.44", false},
		{"trailing zero dropped", "£303.4", "£303.40", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			page := primed(map[string]string{"#out": tc.served})
			ch := criteriaCheck{ID: "c", Type: "computed_values",
				ExpectValues: map[string]string{"#out": tc.expected}}
			pass, detail := runComputedValues(page, ch)
			if pass != tc.wantPass {
				t.Errorf("served %q vs expected %q: pass=%v want %v (%s)",
					tc.served, tc.expected, pass, tc.wantPass, detail)
			}
		})
	}
}

// A tool whose script failed to run reports EVERY output wrong. The message must
// stay readable and must say how many there were, or the one clause identifying
// the cause is buried.
func TestComputedValues_FailureMessageIsBoundedAndStable(t *testing.T) {
	texts := map[string]string{}
	expect := map[string]string{}
	for _, sel := range []string{"#a", "#b", "#c", "#d", "#e", "#f"} {
		texts[sel] = "£0.00"
		expect[sel] = "£123.45"
	}
	page := primed(texts)
	ch := criteriaCheck{ID: "c", Type: "computed_values", ExpectValues: expect}

	pass, detail := runComputedValues(page, ch)
	if pass {
		t.Fatal("every output wrong must fail")
	}
	if !strings.Contains(detail, "6 of 6") {
		t.Errorf("the message must report the full count, not only the sample: %s", detail)
	}
	if !strings.Contains(detail, "+3 more") {
		t.Errorf("the sample must be bounded and say what it omitted: %s", detail)
	}
	// Stability: the same defect must produce the same message every run, or a
	// constant fault reads as a flapping one.
	_, again := runComputedValues(primed(texts), ch)
	if detail != again {
		t.Errorf("message is not deterministic across runs:\n  %s\n  %s", detail, again)
	}
}

// The check must be reachable through the real dispatch path, not only by
// calling the function directly — the gap that makes a check look implemented
// while never running.
func TestComputedValues_IsDispatchedByEvaluateOnPage(t *testing.T) {
	page := primed(map[string]string{"#out": "£1.00"})
	doc := criteriaDoc{Checks: []criteriaCheck{{
		ID: "c", Type: "computed_values",
		ExpectValues: map[string]string{"#out": "£2.00"},
	}}}
	applicable, skipped := splitByProfile(doc, "desktop", "https://example.test/t")
	if len(applicable) != 1 {
		t.Fatalf("computed_values must be applicable at Tier 4, got applicable=%d skipped=%+v",
			len(applicable), skipped)
	}
	results := evaluateOnPage(page, doc, applicable, "desktop", "https://example.test/t")
	if len(results) != 1 || results[0].Pass {
		t.Fatalf("dispatch must run the check and report its failure, got %+v", results)
	}
}
