// FILE: platform/orchestration/actions/criteria_value_assertions.go
//
// Does this fence assert a NUMBER? — bugs_open/449.
//
// Every rung of the acceptance ladder judges what a page CONTAINS. One check
// type judges what it COMPUTES. Measured 2026-09-03 over `doc_plans`
// (`subject_type='tool' AND is_current`): `tool-generator` has written **186**
// current fences, **115** of them assert no expected value of any kind, and
// **zero** use `computed_values`. The newest carried `created_at` of that same
// day, so this is a live intake and not a standing backlog. The cause is one
// place — the type is absent from both fence-authoring prompts, so it is never
// a candidate — but the DAMAGE is elsewhere, and that is what this file is for:
//
//	a Tier-4 PASS on a fence that asserts no value means "the page loaded and
//	something appeared when we clicked". It is routinely read as "the
//	calculator works". Those are different claims and the record did not
//	distinguish them.
//
// So this is the single definition of "this fence compares a value", shared by
// the judge that reports a verdict (tool_acceptance_actions.go) and the write
// door that accepts a fence (write_doc_plan_action.go). One definition rather
// than two, for the reason criteriaFactsFromValue states beside it: two
// spellings of one rule is the drift class this package keeps catching.
//
// THREE GRADES, NOT TWO, AND THE MIDDLE ONE IS THE POINT.
//
//	exact    a `computed_values` check with a non-empty `expect_values` map.
//	         "#monthlyPayment reads exactly £303.44."
//	pattern  an `interaction` check with a non-empty `expect.text_matches`.
//	         "#monthlyPayment matches /£[\d,]+\.\d\d/."
//	none     neither, anywhere in the document.
//
// Collapsing `pattern` into `exact` would overclaim and collapsing it into
// `none` would underclaim, and both errors are live in the corpus. bugs_open/449
// §2 is explicit that its first version — "zero assert a computed value" — was
// too strong, because 63 of the fences do carry `text_matches` and some of those
// patterns are real values (`\$1000\.00`, `40.0%`). But `runComputedValues`'s own
// docstring is equally explicit the other way: a regexp loose enough to be worth
// authoring by hand, `/£[\d,]+\.\d\d/`, "is satisfied by any number whatsoever —
// including the £0.00 an unwired tool prints". Both are true. A pattern
// assertion is real evidence and weaker evidence, so it is reported as itself.
//
// DRIVESINPUTS IS READ OFF THE FENCE, NEVER OFF A CLASSIFIER.
//
// The tempting trigger for "this tool should assert a number" is "this tool is a
// calculator", which needs a classifier over tool kinds, and a guarantee
// conditional on a classifier inherits the classifier's gaps. The fence already
// contains the evidence: a check carrying a `fill` or `select` step has
// DECLARED that the tool takes input. A document that drives inputs and then
// asserts nothing about what came out is self-evidently incomplete, on its own
// terms, with no outside judgement about what the tool is. Measured the same
// day: 91 of the 186 drive inputs and 55 of those assert nothing.
//
// FAIL-OPEN, exactly like parseCriteriaFacts and parseNoAutoFix beside it. An
// unparseable or absent fence yields Parsed=false and grade `none`, and every
// caller must branch on Parsed before treating a zero as a finding — a fence
// nobody could read has not been shown to assert nothing. Tier 2 already reports
// that case separately as `criteria_unparseable`, and this must not quietly
// become a second, weaker report of the same thing.

package actions

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Assertion grades, in increasing strength. Strings rather than an int because
// they are written into action results and doc_note bodies that humans read.
const (
	criteriaAssertsNone    = "none"
	criteriaAssertsPattern = "pattern"
	criteriaAssertsExact   = "exact"
)

// criteriaValueDoc is the assertion-counter's view of a fence. Deliberately NOT
// folded into criteriaFactsDoc or acceptanceFenceFlags: those are the fan-out's
// and the judge's views, and a fence malformed in one of those keys must not
// change the answer to "does it compare a value".
type criteriaValueDoc struct {
	Checks []criteriaValueCheck `json:"checks"`
}

type criteriaValueCheck struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Steps []struct {
		Action string `json:"action"`
	} `json:"steps"`
	Expect struct {
		TextMatches string `json:"text_matches"`
	} `json:"expect"`
	ExpectValues map[string]string `json:"expect_values"`
}

// criteriaValueSummary is what one fence says about the values it compares.
type criteriaValueSummary struct {
	// Parsed is false when the fence was absent or would not decode. Branch on
	// it before reading anything else: a zero from an unreadable document is not
	// evidence that the document asserts nothing.
	Parsed bool

	// Exact counts `computed_values` checks carrying at least one expectation.
	// A `computed_values` check with an EMPTY expect_values map is NOT counted,
	// and that is not an oversight: the runner refuses it outright ("it would
	// assert nothing and pass on any page"), so counting it would credit a fence
	// for a check that can only ever fail.
	Exact int

	// Pattern counts `interaction` checks carrying a non-empty text_matches.
	Pattern int

	// DrivesInputs is true when any check carries a `fill` or `select` step —
	// the fence's own declaration that the tool takes input.
	DrivesInputs bool

	// AssertingIDs names the checks that compare something, exact first, so a
	// message can cite them rather than assert a count. Sorted by appearance,
	// which is the order a reader of the fence sees.
	AssertingIDs []string
}

// Grade returns none / pattern / exact — the strongest assertion present.
func (s criteriaValueSummary) Grade() string {
	switch {
	case s.Exact > 0:
		return criteriaAssertsExact
	case s.Pattern > 0:
		return criteriaAssertsPattern
	default:
		return criteriaAssertsNone
	}
}

// Total is every value-comparing check, exact and pattern together.
func (s criteriaValueSummary) Total() int { return s.Exact + s.Pattern }

// AssertsNoValue reports the state bugs_open/449 is about: a fence that was
// READ successfully and compares nothing. False for an unreadable fence, so a
// caller cannot mistake "could not tell" for "asserts nothing".
func (s criteriaValueSummary) AssertsNoValue() bool {
	return s.Parsed && s.Total() == 0
}

// DrivesButAssertsNothing is the sharp subset — the fence fills the tool's
// inputs and then never looks at what came out. This is the trigger the write
// door uses, because it needs no judgement about what kind of tool this is.
func (s criteriaValueSummary) DrivesButAssertsNothing() bool {
	return s.AssertsNoValue() && s.DrivesInputs
}

// summariseCriteriaValueAssertions reads a raw criteria fence (the contents of
// the ```criteria block, as extractCriteriaFence / extractCriteriaBlock return
// it) and reports what it compares. Fail-open: see the file header.
func summariseCriteriaValueAssertions(criteria string) criteriaValueSummary {
	var s criteriaValueSummary
	if strings.TrimSpace(criteria) == "" {
		return s
	}
	var doc criteriaValueDoc
	if err := json.Unmarshal([]byte(criteria), &doc); err != nil {
		return s
	}
	s.Parsed = true

	var exactIDs, patternIDs []string
	for _, ch := range doc.Checks {
		for _, st := range ch.Steps {
			// `click` and `reload` do not supply a value, so they do not make
			// the tool an input-taker; only fill and select do.
			if st.Action == "fill" || st.Action == "select" {
				s.DrivesInputs = true
				break
			}
		}
		switch ch.Type {
		case "computed_values":
			if len(ch.ExpectValues) > 0 {
				s.Exact++
				exactIDs = append(exactIDs, ch.ID)
			}
		case "interaction":
			if strings.TrimSpace(ch.Expect.TextMatches) != "" {
				s.Pattern++
				patternIDs = append(patternIDs, ch.ID)
			}
		}
	}
	s.AssertingIDs = append(exactIDs, patternIDs...)
	return s
}

// criteriaAssertionPhrase renders a summary for a doc_note or a log line, in
// the terms a human quoting a verdict needs. It always says what the verdict
// does NOT cover, because the failure this exists to stop is a true statement
// ("PASSED") being read as a wider one.
func criteriaAssertionPhrase(s criteriaValueSummary) string {
	if !s.Parsed {
		return "the fence could not be read, so what it asserts is UNKNOWN — this pass covers only the checks that ran"
	}
	switch s.Grade() {
	case criteriaAssertsExact:
		if s.Pattern > 0 {
			return "this fence compares " + plural(s.Exact, "exact value", "exact values") +
				" and " + plural(s.Pattern, "text pattern", "text patterns") + " (" +
				strings.Join(s.AssertingIDs, ", ") + ")"
		}
		return "this fence compares " + plural(s.Exact, "exact value", "exact values") +
			" (" + strings.Join(s.AssertingIDs, ", ") + ")"
	case criteriaAssertsPattern:
		return "this fence compares " + plural(s.Pattern, "text pattern", "text patterns") +
			" (" + strings.Join(s.AssertingIDs, ", ") + ") but no exact value — a pattern loose" +
			" enough to author by hand is satisfied by any number, including the one an" +
			" unwired tool prints, so treat this as weaker than an arithmetic check"
	default:
		if s.DrivesInputs {
			return "⚠ LIVENESS ONLY — this fence FILLS the tool's inputs and then asserts no value" +
				" of any kind, so the verdict says the tool responded and says NOTHING about whether" +
				" any number it printed is correct (bugs_open/449)"
		}
		return "⚠ LIVENESS ONLY — this fence asserts no value of any kind, so the verdict says the" +
			" page loads and responds and says NOTHING about what it computes (bugs_open/449)"
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}
