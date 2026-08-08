// FILE: platform/orchestration/actions/validate_page_content_meta_scope_test.go
//
// The meta-commentary scan reads PROSE, not code — bugs_open/219.
//
// It used to substring-scan the whole assembled page at BLOCKER severity, so
// it convicted a human's documentation: three locked calculator components on
// loancalculator.co.uk explain themselves in a `/* … */` comment inside their
// <script>, and those notes contain "input_schema" because that is the name of
// the thing they document. The build failed before save_page_sections every
// time — car-finance was fired three times and failed identically — quoting a
// developer note written five days before the run that it blamed on the model.
//
// The predictor that established the mechanism: a voice rewrite was dispatched
// at all 12 pages carrying a locked tool component. Exactly the 3 whose
// template contains the string failed; the other 9 passed.
//
// THE SCRIPT CASE IS THE LOAD-BEARING ONE. 219's own leading fix candidate was
// "strip HTML comments before scanning", and that would have shipped INERT:
// measured 2026-08-08, the hit in all three templates lies between a <script>
// open and its </script> close, and two of the three contain no HTML comment
// at all. Scope was wrong, not comment syntax — hence ExtractAssertionText,
// which drops every non-assertion context at once.
//
// The "still blocks" half is the other side of the contract: re-scoping must
// not retire the rule. A false negative here ships an apology to readers,
// which is the failure the check was built for (robot-hands, 2026-07-14 —
// TestCheckMetaCommentary in validate_page_content_meta_test.go).

package actions

import (
	"strings"
	"testing"
)

// Verbatim from content_components.html_template for tool-car-finance-pcp-hp
// (the changelog note the validator quoted), inside the <script> it documents.
const carFinanceScriptNote = `<div class="calc-tool">
<script>
/* ARITHMETIC PRESERVED EXACTLY from the hand-built original
   (loancalculator.co.uk/tools/car-finance-calculator.html, 2026-07-31).

   THREE deliberate changes, none of which can move a number:
   1. No globals. ` + "`mode`" + ` was a bare ` + "`let`" + ` at page scope and ` + "`setType`" + ` was on
      window — two of these tools on one page would have shared one mode.
   2. No inline onclick/oninput. The mode buttons carry data-cf-mode and one
      delegated listener reads it.
   3. Labels, defaults and the currency symbol come from the input_schema.

   FIXED 2026-08-03 — 0% APR, as its own change. The whole calculation used to
   be wrapped in ` + "`if (apr > 0)`" + `, so an interest-free deal — a real and heavily
   advertised car finance product — computed nothing.
*/
const cfMode = 'hp';
</script>
<h2>Car Finance Calculator</h2>
<p>Work out the monthly cost of a PCP or HP agreement, including the balloon
payment, and see how much the finance adds to the price of the car.</p>
</div>`

func metaValues(issues []ValidationIssue) []string {
	vals := make([]string, 0, len(issues))
	for _, i := range issues {
		vals = append(vals, i.Value)
	}
	return vals
}

func TestMetaCommentaryIgnoresNonProseContexts(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{
			// The live case. A JS comment — NOT an HTML comment.
			"developer changelog in a script block",
			carFinanceScriptNote,
		},
		{
			// What 219 first believed the live case was. Also correct to pass.
			"template documentation in an HTML comment",
			`<section><!-- slot contract: input_schema drives the labels;
			 on_missing: skip_section when the query returns nothing -->
			 <h2>Repayment</h2><p>Enter the amount you want to borrow.</p></section>`,
		},
		{
			// A tool's own CSS naming its schema-driven hooks.
			"style block",
			`<div><style>/* widths come from input_schema */ .f{width:4rem}</style>
			 <p>Adjust the term to see the effect on total interest.</p></div>`,
		},
		{
			// Documentation shown to a reader ON PURPOSE, as marked-up code.
			"code sample",
			`<article><p>Every tool declares its inputs:</p>
			 <pre><code>input_schema: {amount: {required: true}}</code></pre></article>`,
		},
		{
			// Attributes are not assertions.
			"data attribute",
			`<div data-config="on_missing:skip_section"><p>Compare two loans side by side.</p></div>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := checkMetaCommentary(tc.html); len(got) != 0 {
				t.Errorf("non-prose context convicted as meta-commentary: %v", metaValues(got))
			}
		})
	}
}

// The negative control. A scan that cannot fail is not a check, and every case
// above is an absence — so each pattern family is induced in real copy here.
func TestMetaCommentaryStillBlocksVisibleCopy(t *testing.T) {
	cases := []struct {
		name    string
		html    string
		wantVal string
	}{
		{
			"schema vocabulary in a paragraph",
			`<section><h2>Repayment</h2><p>The input_schema for this section defines the
			 fields, so the amounts below are illustrative.</p></section>`,
			"input_schema",
		},
		{
			"pipeline vocabulary in a list item",
			`<ul><li>Rates are indicative; on_missing this row is omitted.</li></ul>`,
			"on_missing",
		},
		{
			"refusal prose in a paragraph",
			`<div><p>I don't have access to current lender rates, so this table is empty.</p></div>`,
			"i don't have access to",
		},
		{
			// The exact escape the placeholder scan had to be re-widened for
			// (35889819c): <head> is not an assertion context, but a meta
			// description IS prose a visitor reads. headProseBlocks re-adds it.
			"meta description",
			`<html><head><meta name="description" content="As an AI I cannot generate a summary of these rates."></head><body><p>Loan rates.</p></body></html>`,
			"as an ai",
		},
		{
			"title",
			`<html><head><title>The data schema for this page is unavailable</title></head><body><p>Loans.</p></body></html>`,
			"the data schema",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkMetaCommentary(tc.html)
			if len(got) == 0 {
				t.Fatalf("visible copy containing %q was not flagged", tc.wantVal)
			}
			found := false
			for _, issue := range got {
				if issue.Value == tc.wantVal {
					found = true
					if issue.Severity != "blocker" {
						t.Errorf("severity = %q, want blocker", issue.Severity)
					}
					if issue.Location == "" {
						t.Error("no location snippet — a blocker that cannot be located is not actionable")
					}
				}
			}
			if !found {
				t.Errorf("flagged %v, want a hit on %q", metaValues(got), tc.wantVal)
			}
		})
	}
}

// The check must not name a cause it has not established. It is a substring
// scan: it can see the vocabulary, never who typed it. On 2026-08-08 its old
// wording ("the model wrote about its task instead of doing it") was believed
// for an hour about a note a human wrote five days earlier.
func TestMetaCommentaryDescriptionDoesNotBlameTheModel(t *testing.T) {
	issues := checkMetaCommentary(`<p>The input_schema for this section is empty.</p>`)
	if len(issues) == 0 {
		t.Fatal("expected a hit to inspect")
	}
	for _, issue := range issues {
		if strings.Contains(strings.ToLower(issue.Description), "the model wrote") {
			t.Errorf("description asserts an unestablished cause: %q", issue.Description)
		}
		if !strings.Contains(issue.Description, issue.Value) {
			t.Errorf("description %q does not name the matched pattern %q", issue.Description, issue.Value)
		}
	}
}
