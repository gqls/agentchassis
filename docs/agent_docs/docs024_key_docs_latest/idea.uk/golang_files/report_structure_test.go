package main

// report_structure_test.go — the assessment half's readability.
//
// Owner, 2026-07-28: "the report is text heavy and we could probably split the
// text up with subheadings a bit." The seven assessment fields are several
// hundred words each and used to render label-and-value inside ONE <p>, so each
// arrived as a wall with a bold run glued to the front.
//
// These pin the STRUCTURE, not the styling — a later restyle should be free, but
// silently collapsing the label back into the prose should not.

import (
	"strings"
	"testing"
)

func testAssessment() assessment {
	return assessment{
		IsAssessable: true,
		Reading:      "reads as a worked-through proposition",
		Problem:      "First paragraph about the problem.\n\nSecond paragraph with more detail.",
		NextStep:     "Spend fifty pounds testing one town before writing any software.",
	}
}

// The label must be its own element, not a bold run inside the prose.
func TestAssessmentLabelIsItsOwnBlock(t *testing.T) {
	h := renderHTML("an idea", "surveyors", "because it saves time", testAssessment(), nil, nil, nil, "")

	if strings.Contains(h, "The problem, and the evidence people have it</p>") == false &&
		strings.Contains(h, "The problem, and the evidence people have it<") == false {
		t.Fatalf("assessment label not found as its own element")
	}
	// The old shape: label and value in one paragraph, separated by "</span> ".
	if strings.Contains(h, "people have it:</span> First paragraph") {
		t.Error("label is still glued to the value inside one <p> — the wall is back")
	}
}

// A trailing colon reads as a label when the value follows on the same line and
// as a typo when it does not. Dropped in HTML, kept in text.
func TestAssessmentHeadingDropsTrailingColonInHTMLOnly(t *testing.T) {
	h := renderHTML("an idea", "surveyors", "because", testAssessment(), nil, nil, nil, "")
	if strings.Contains(h, "A considered next step:</p>") {
		t.Error("HTML subheading kept its trailing colon")
	}
	txt := render("an idea", "surveyors", "because", testAssessment(), nil, nil, nil, "")
	if !strings.Contains(txt, "A considered next step:") {
		t.Error("text label lost its colon — there the value really does follow it")
	}
}

// Model prose containing blank lines must become separate paragraphs in both
// renderers rather than one run with literal newlines in it.
func TestModelParagraphsSurviveIntoBothRenderers(t *testing.T) {
	h := renderHTML("an idea", "surveyors", "because", testAssessment(), nil, nil, nil, "")
	if !strings.Contains(h, "First paragraph about the problem.</p>") {
		t.Errorf("first model paragraph not closed as its own <p>")
	}
	if !strings.Contains(h, "Second paragraph with more detail.</p>") {
		t.Errorf("second model paragraph not rendered as its own <p>")
	}

	txt := render("an idea", "surveyors", "because", testAssessment(), nil, nil, nil, "")
	if !strings.Contains(txt, "   First paragraph about the problem.\n   Second paragraph with more detail.") {
		t.Errorf("text renderer did not keep the model's paragraphs on separate lines:\n%s", excerptAround(txt, "First paragraph"))
	}
}

// Without a blank line between entries the seven fields run together into one
// block in a plain-text client — the same wall, different renderer.
func TestTextEntriesAreSeparatedByABlankLine(t *testing.T) {
	txt := render("an idea", "surveyors", "because", testAssessment(), nil, nil, nil, "")
	if !strings.Contains(txt, "Second paragraph with more detail.\n\n") {
		t.Errorf("no blank line after an assessment entry:\n%s", excerptAround(txt, "Second paragraph"))
	}
}

// An empty field must render nothing at all — a bare heading with no prose under
// it looks like the report failed to generate something.
func TestEmptyAssessmentFieldRendersNoHeading(t *testing.T) {
	a := testAssessment()
	a.Exposed = "   " // whitespace only
	h := renderHTML("an idea", "surveyors", "because", a, nil, nil, nil, "")
	if strings.Contains(h, "Where it is exposed") {
		t.Error("an empty field rendered a dangling heading")
	}
}

// paragraphs() must never return an empty slice, or a caller ranging over it
// silently drops the field's entire content.
func TestParagraphsNeverReturnsEmpty(t *testing.T) {
	for _, in := range []string{"", "   ", "\n\n", "one", "one\n\ntwo"} {
		if got := paragraphs(in); len(got) == 0 {
			t.Errorf("paragraphs(%q) returned an empty slice", in)
		}
	}
	if got := paragraphs("one\n\n\n\ntwo"); len(got) != 2 {
		t.Errorf("paragraphs collapsed runs of blank lines wrongly: %q", got)
	}
}
