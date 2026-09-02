// FILE: platform/orchestration/datahelpers/cta_eligibility_label_test.go
//
// Pins bugs_open/436's lever at the LABEL MATCH — constraint 2 of the 391
// review: the label match runs AHEAD of the positional pick at both writers,
// so an opt-out read only by the ranking has a hole exactly the shape of the
// original damage — a page the ranking refuses still wins through its own
// copy, which is how a wrong pick label-locked itself on 20 of 80 fields.
//
// AND IT PINS THE SHAPE OF THE FIX, exactly as the self-link test above does
// (TestBestLabelMatchForPageRefusesTheButtonsOwnPage): the opted-out page is
// REFUSED, not FILTERED OUT of the pool. Filtering was measured wrong for
// self-links — once the best candidate is gone, a single shared token lets a
// runner-up win, and most such writes were wrong. The runner-up sub-test here
// fails if someone "simplifies" the eligibility rule into a WHERE clause in
// CTALabelUniverseSQL or a pre-filter over the candidate slice.

package datahelpers

import "testing"

func eligibilityFixture(t *testing.T) (pages []LabelMatchCandidate) {
	t.Helper()
	mk := func(name, title, url string, eligible bool) LabelMatchCandidate {
		c, ok := CTALabelCandidateRow(name, name, title, "", url, "content", eligible)
		if !ok {
			t.Fatalf("fixture %q produced no candidate", name)
		}
		return c
	}
	// A discriminating pair, per the self-link test's lesson: the runner-up
	// SHARES a token with the label ("shapes"), so a pre-filter implementation
	// hands it the button and this file can tell the two implementations apart.
	return []LabelMatchCandidate{
		mk("flight-shapes", "Dart Flight Shapes Explained", "/blog/flight-shapes.html", false), // opted out
		mk("barrel-shapes", "Dart Barrel Shapes Explained", "/blog/barrel-shapes.html", true),
	}
}

func TestBestLabelMatchForPageRefusesAnOptedOutPage(t *testing.T) {
	pages := eligibilityFixture(t)
	const label = "Compare flight shapes"

	// The runner-up really would win if the opted-out page were filtered out.
	if got, ok, _ := BestLabelMatch(label, pages[1:]); !ok || got.URL != "/blog/barrel-shapes.html" {
		t.Fatalf("fixture is inert: with the opted-out page removed the runner-up must match, got %+v ok=%v", got, ok)
	}

	t.Run("copy naming the opted-out page is refused", func(t *testing.T) {
		got, ok, _ := BestLabelMatchForPage(label, pages, "some-other-page", "")
		if ok || got.URL != "" {
			t.Errorf("resolved to %q — an opted-out page won through its own copy, "+
				"which is bugs_open/436's lock-in hole", got.URL)
		}
	})
	t.Run("refuses rather than falling through to the runner-up", func(t *testing.T) {
		// Same call as above; asserted separately so a pre-filter rewrite
		// fails with this name in the output, pointing at the measured reason.
		got, ok, _ := BestLabelMatchForPage(label, pages, "some-other-page", "")
		if ok {
			t.Errorf("fell through to %q instead of refusing — dropping the best candidate "+
				"from the pool let noise win on 10 of 35 rows in the 2026-08-23 audit", got.URL)
		}
	})
	t.Run("flag off, the same copy resolves again", func(t *testing.T) {
		eligible := eligibilityFixture(t)
		eligible[0].IneligibleAsCTATarget = false
		got, ok, _ := BestLabelMatchForPage(label, eligible, "some-other-page", "")
		if !ok || got.URL != "/blog/flight-shapes.html" {
			t.Errorf("with eligibility restored the match must return: got %+v ok=%v", got, ok)
		}
	})
	t.Run("other pages' copy still resolves", func(t *testing.T) {
		got, ok, _ := BestLabelMatchForPage("Read the barrel shapes guide", pages, "some-other-page", "")
		if !ok || got.URL != "/blog/barrel-shapes.html" {
			t.Errorf("eligibility refusal must not disable ordinary matching: got %+v ok=%v", got, ok)
		}
	})
}

// TestJudgeCTALabelReportsAnIneligibleName pins the judge's new silence
// reason: detectors keep a first-class "copy names a page the estate ruled
// out" signal instead of it folding into names_nothing — and, critically, the
// judge no longer suggests (via Contradicts+Named) a repair the writers now
// refuse to perform, which is the 308 defect shape (188 findings naming a
// destination no writer could write).
func TestJudgeCTALabelReportsAnIneligibleName(t *testing.T) {
	pages := eligibilityFixture(t)

	j := JudgeCTALabel("Compare flight shapes", "/blog/barrel-shapes.html", pages, "some-other-page", "")
	if j.Verdict != CTALabelNoOpinion {
		t.Fatalf("verdict = %v, want no_opinion", j.Verdict)
	}
	if j.Silence != SilenceNamesIneligiblePage {
		t.Errorf("silence = %q, want names_ineligible_page", j.Silence)
	}
	if got := j.Silence.String(); got != "names_ineligible_page" {
		t.Errorf("Silence.String() = %q — this string lands in logs and later queries key on it", got)
	}

	t.Run("self-link stays the more specific reason", func(t *testing.T) {
		// A label naming the page it sits on, where that page is ALSO opted
		// out: the content defect ("the button should not be there") outranks
		// the eligibility one, matching the refusals' order in
		// BestLabelMatchForPage.
		j := JudgeCTALabel("Compare flight shapes", "/blog/barrel-shapes.html", pages, "flight-shapes", "")
		if j.Silence != SilenceNamesItsOwnPage {
			t.Errorf("silence = %q, want names_its_own_page", j.Silence)
		}
	})
}

// TestCTALabelCandidateRowCarriesEligibility pins the polarity at the one
// place the column crosses into Go. The field is inverted from the column on
// purpose (zero value = eligible); a sign flip here would silently refuse
// every label match fleet-wide, so the polarity gets its own assertion.
func TestCTALabelCandidateRowCarriesEligibility(t *testing.T) {
	c, ok := CTALabelCandidateRow("1", "guide", "The Guide", "", "/guide.html", "content", true)
	if !ok || c.IneligibleAsCTATarget {
		t.Errorf("eligible column true must produce IneligibleAsCTATarget=false (ok=%v, got %v)", ok, c.IneligibleAsCTATarget)
	}
	c, ok = CTALabelCandidateRow("2", "toy", "The Toy", "", "/toy.html", "content", false)
	if !ok || !c.IneligibleAsCTATarget {
		t.Errorf("eligible column false must produce IneligibleAsCTATarget=true (ok=%v, got %v)", ok, c.IneligibleAsCTATarget)
	}
}
