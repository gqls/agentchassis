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
	if !contains(got, "at the counter. First we assess") {
		t.Errorf("intro does not read cleanly into the next sentence:\n%s", got[:min(300, len(got))])
	}
}

// An unpunctuated description must still get its full stop.
func TestReportIntroAddsAMissingFullStop(t *testing.T) {
	got := reportIntro("a comparison service for UK vets")
	if !contains(got, "for UK vets. First we assess") {
		t.Errorf("intro did not terminate the submitted fragment:\n%s", got[:min(300, len(got))])
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
