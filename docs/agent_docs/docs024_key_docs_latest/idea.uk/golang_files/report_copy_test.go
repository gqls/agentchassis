package main

// report_copy_test.go — copy defects caught by reading a real delivered report.
//
// All three shipped in `ord_1785090638951163875` (2026-07-26, the first run
// taken through approve → pay link). None is a crash and none would ever fail a
// build; they are only visible by reading the artefact, which is the argument
// for reading it.

import "testing"

// "…rather than finding out at the counter.. First we assess the idea…"
// The submitted description already ended in a full stop and the intro appended
// another.
func TestReportIntroDoesNotDoubleTheFullStop(t *testing.T) {
	submitted := "A price-transparency comparison service for UK veterinary practices: owners " +
		"get an upfront estimate for routine treatments before they book, rather than finding " +
		"out at the counter."
	got := reportIntro(submitted)
	if contains(got, "..") {
		t.Errorf("doubled full stop in the report intro:\n%s", excerptAround(got, ".."))
	}
	// The submitted fragment ends the first paragraph, so it must carry exactly
	// one stop and then the paragraph break — not run on into what follows.
	if !contains(got, "at the counter.\n\n") {
		t.Errorf("submitted fragment not cleanly terminated at the paragraph break:\n%s", got[:min(300, len(got))])
	}
}

// An unpunctuated description must still get its full stop.
func TestReportIntroAddsAMissingFullStop(t *testing.T) {
	got := reportIntro("a comparison service for UK vets")
	if !contains(got, "for UK vets.\n\n") {
		t.Errorf("intro did not terminate the submitted fragment:\n%s", got[:min(300, len(got))])
	}
}

// The intro is the reader's first sight of a £29 purchase. It used to be one
// block of four long sentences; this pins the structure so a later edit cannot
// quietly collapse it back into a wall of prose.
func TestReportIntroIsBrokenIntoParagraphs(t *testing.T) {
	paras := introParagraphs("a comparison service for UK vets")
	if len(paras) < 3 {
		t.Fatalf("want at least 3 intro paragraphs, got %d:\n%q", len(paras), paras)
	}
	for i, p := range paras {
		if p == "" {
			t.Errorf("intro paragraph %d is empty — check for a stray blank line", i)
		}
	}
	// The reader's own words belong in the opener, not buried in the middle.
	if !contains(paras[0], "for UK vets.") {
		t.Errorf("the submitted idea should be in the FIRST paragraph, got: %q", paras[0])
	}
}

func TestSentence(t *testing.T) {
	for in, want := range map[string]string{
		"ends in a stop.": "ends in a stop.",
		"needs one":       "needs one.",
		"a question?":     "a question?",
		"  padded  ":      "padded.",
		"":                "",
		"ends in a bang!": "ends in a bang!",
	} {
		if got := sentence(in); got != want {
			t.Errorf("sentence(%q) = %q, want %q", in, got, want)
		}
	}
}

// "What it's built on: The practice's price list…, using A form the receptionist
// fills in about the pet…" — a sentence-cased model field spliced mid-sentence.
func TestMidSentenceLowercasesASplicedFragment(t *testing.T) {
	got := midSentence("A form the receptionist fills in about the pet")
	if got != "a form the receptionist fills in about the pet" {
		t.Errorf("midSentence = %q, want the first letter lowered", got)
	}
}

// Acronyms and initialisms must survive — "AI", "UK", "CMA" appear constantly in
// these fields and lowering them would be a worse defect than the one being fixed.
func TestMidSentenceLeavesAcronymsAlone(t *testing.T) {
	for _, s := range []string{
		"AI that reads the practice's price list",
		"UK vet practices with published prices",
		"CMA guidance on price transparency",
		"already lowercase text",
	} {
		if got := midSentence(s); got != s {
			t.Errorf("midSentence(%q) = %q, want it unchanged", s, got)
		}
	}
}

func contains(hay, needle string) bool {
	for i := 0; i+len(needle) <= len(hay); i++ {
		if hay[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func excerptAround(s, needle string) string {
	for i := 0; i+len(needle) <= len(s); i++ {
		if s[i:i+len(needle)] == needle {
			return s[max(0, i-60):min(len(s), i+60)]
		}
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
