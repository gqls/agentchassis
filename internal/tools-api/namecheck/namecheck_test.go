package namecheck

import (
	"strings"
	"testing"
)

// The product case. If these are refused, the feature is dead: the ruled audience
// (PLAN §11.1) is food, music and film canon, cities and generational habits,
// where naming things IS the argument. Every one of these names a real entity and
// says something unkind about it, and every one is opinion.
func TestOpinionAboutNamedThingsIsALLOWED(t *testing.T) {
	for _, s := range []string{
		"Christopher Nolan is overrated and Tenet proves it.",
		"The Beatles are the most boring band ever canonised.",
		"New York is a worse city than Chicago and everyone knows it.",
		"Manchester United have been derivative for a decade.",
		"Everything The Guardian publishes about food is wrong.",
		"Kubrick's later films are cold, airless and dull.",
		"Marmite is genuinely unpleasant and Unilever should stop.",
		"Radiohead peaked in 1997 and have coasted since.",
	} {
		if f := Scan(s); f != nil {
			t.Errorf("REFUSED an opinion, which kills the product:\n  %q\n  findings: %+v", s, f)
		}
	}
}

// The risk case. Each names an apparent person or business AND makes a checkable
// factual allegation about it — the shape RFC_020 exists to keep off a public URL.
func TestAllegationsAboutNamedEntitiesAreREFUSED(t *testing.T) {
	for _, s := range []string{
		"Christopher Nolan stole the script for Inception.",
		"Dr Palmer falsified the trial data.",
		"Everyone knows Aldridge Holdings laundered the money.",
		"@bigchef123 plagiarised that entire menu.",
		"The chef at Hollow Bone is a fraud who lied about his training.",
		"Mr Hayes was arrested for it and they covered it up.",
		"Sandra Whitcombe embezzled from the fund for years.",
	} {
		if f := Scan(s); f == nil {
			t.Errorf("ALLOWED an allegation about a named party:\n  %q", s)
		}
	}
}

// A word capitalised only because it opens a sentence is not a name.
//
// NOTE ON THIS TEST'S OWN HISTORY: it first asserted that "The Guardian lied
// about the numbers." must be ALLOWED, on the reasoning that the pair opens the
// sentence. The code refused it and the CODE WAS RIGHT — "The" is the opener,
// but "Guardian" is a name by any reading, and an allegation of dishonesty
// against a named publication is exactly what this exists to catch. The example
// was wrong, not the rule. Replaced with one where the whole run really is
// grammatical, and the original kept below as a positive case.
func TestSentenceOpeningCapitalsAreNotTreatedAsNames(t *testing.T) {
	for _, s := range []string{
		"This restaurant lied about its sourcing.",
		"They cheated us on the bill.",
		"Every supplier here is dishonest about provenance.",
	} {
		if f := Scan(s); f != nil {
			t.Errorf("a sentence-opening capital was read as a name:\n  %q\n  %+v", s, f)
		}
	}

	// The opener is skipped, but a real name AFTER it is not.
	if f := Scan("The Guardian lied about the numbers."); f == nil {
		t.Error("a named publication plus an allegation was allowed")
	}
	if f := Scan("It turns out The Guardian lied about the numbers."); f == nil {
		t.Error("a mid-sentence named entity with an allegation was allowed")
	}
}

// Proximity is load-bearing: an allegation about something else, far away in the
// text, must not attach itself to whatever name happened to appear.
func TestAllegationMustBeNEARTheName(t *testing.T) {
	near := "Christopher Nolan stole it."
	if Scan(near) == nil {
		t.Fatal("guard: the near case must be refused, or the far case proves nothing")
	}

	far := "Christopher Nolan directs films about time, memory, guilt, gravity, " +
		"dreams, magic, war and physics, and separately the whole banking industry " +
		"was corrupt."
	if f := Scan(far); f != nil {
		t.Errorf("an allegation %d+ words away was attached to a name: %+v", proximityWords, f)
	}
}

// The most damaging terms must not require a detectable name — a pronoun is
// enough to make them unpublishable.
func TestStandaloneTermsAreRefusedWithoutAName(t *testing.T) {
	for _, s := range []string{
		"he is a paedophile and everyone in that town knew",
		"she groomed those kids for years",
	} {
		if f := Scan(s); f == nil {
			t.Errorf("ALLOWED a standalone allegation with no name: %q", s)
		}
	}
}

// The AI's own output is checked on the same terms as the visitor's. RFC_020 §1.4:
// the verdict is the service's text, so trusting it for being ours is exactly the
// mistake — it is the part we are most clearly the author of.
func TestScanAllCoversTheModelsOwnText(t *testing.T) {
	visitorPosition := "Nolan is overrated."         // fine
	modelVerdict := "Agreed — Nolan plagiarised it." // not fine, and it is OUR text
	if f := ScanAll(visitorPosition, modelVerdict); f == nil {
		t.Fatal("ScanAll passed a round whose AI-authored half carries the allegation")
	}
	// And the visitor's half alone must still be clean, or the test above proves
	// nothing about WHICH half was caught.
	if f := Scan(visitorPosition); f != nil {
		t.Errorf("the visitor half should be clean on its own: %+v", f)
	}
}

// Mutation control. The proximity window is the fiddliest part and therefore the
// most plausible thing for a later edit to "simplify" away. This asserts the
// windowed and unwindowed behaviours genuinely DIFFER on the far case, so
// TestAllegationMustBeNEARTheName cannot be passing by accident.
func TestTheProximityTestsWouldCatchTheWindowBeingRemoved(t *testing.T) {
	far := "Christopher Nolan directs films about time, memory, guilt, gravity, " +
		"dreams, magic, war and physics, and separately the whole banking industry " +
		"was corrupt."

	occs := allegationOccurrences(strings.ToLower(far))
	if len(occs) == 0 {
		t.Fatal("guard: the far case must contain an allegation term, or this test proves nothing")
	}

	withWindow := Scan(far) != nil // expected false: the name is out of range

	// The weakened variant: identical detection, unbounded window.
	withoutWindow := false
	for _, o := range occs {
		if _, ok := nameNear(far, o, 1<<20); ok {
			withoutWindow = true
			break
		}
	}

	if withWindow == withoutWindow {
		t.Fatalf("window makes no difference on the far case (both %v), so the "+
			"proximity tests cannot detect it being removed", withWindow)
	}
}
