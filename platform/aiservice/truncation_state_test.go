package aiservice

import "testing"

// The point of these tests is NOT that the arithmetic works — it is that
// "we could not tell" never collapses into "the answer was complete". Every
// case below that asserts TruncationUnknown is asserting the absence of the
// defect this type exists to prevent (bugs_open/305, council round 2).

func TestClassifyTruncation_states(t *testing.T) {
	cases := []struct {
		name           string
		outTok, maxTok int
		want           TruncationState
	}{
		{"reported and below", 447, 2000, TruncationBelow},
		{"reported and exactly at the ceiling", 2000, 2000, TruncationAtCeiling},
		{"reported and past the ceiling", 2100, 2000, TruncationAtCeiling},
		{"usage unreported (0) is NOT below", 0, 2000, TruncationUnknown},
		{"usage negative is NOT below", -1, 2000, TruncationUnknown},
		{"no ceiling recorded", 447, 0, TruncationUnknown},
		{"neither known", 0, 0, TruncationUnknown},
		{"a ceiling of 1 with 1 produced is still a cut", 1, 1, TruncationAtCeiling},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ClassifyTruncation(c.outTok, c.maxTok); got != c.want {
				t.Fatalf("ClassifyTruncation(%d, %d) = %v, want %v", c.outTok, c.maxTok, got, c.want)
			}
		})
	}
}

// The zero value is the load-bearing design decision: a caller that forgets a
// case, or reads a state off a zeroed struct, must land on the one that forces a
// decision. If someone reorders the const block so TruncationBelow is iota's
// zero, this fails.
func TestZeroValueIsUnknown(t *testing.T) {
	var s TruncationState
	if s != TruncationUnknown {
		t.Fatalf("zero value is %v, want TruncationUnknown — reordering the const block silently makes an unchecked state read as safe", s)
	}
	if s.Complete() {
		t.Fatal("the zero value reports Complete() — 'we could not tell' must never assert the answer finished")
	}
	if s.Truncated() {
		t.Fatal("the zero value reports Truncated() — unknown is not a positive claim either")
	}
}

// Truncated and Complete are deliberately NOT negations of each other. A caller
// writing `if !state.Truncated() { splice() }` is making the mistake this whole
// type exists to prevent, and this test pins the property that lets a reviewer
// catch it: there is exactly one state for which both are false.
func TestTruncatedAndCompleteAreNotComplementary(t *testing.T) {
	bothFalse := 0
	for _, s := range []TruncationState{TruncationUnknown, TruncationBelow, TruncationAtCeiling} {
		if s.Truncated() && s.Complete() {
			t.Fatalf("%v claims both cut and complete", s)
		}
		if !s.Truncated() && !s.Complete() {
			bothFalse++
		}
	}
	if bothFalse != 1 {
		t.Fatalf("%d states report neither cut nor complete, want exactly 1 (unknown)", bothFalse)
	}
}

func TestStringNamesEveryState(t *testing.T) {
	want := map[TruncationState]string{
		TruncationUnknown:   "unknown",
		TruncationBelow:     "below_ceiling",
		TruncationAtCeiling: "at_ceiling",
	}
	seen := map[string]bool{}
	for s, w := range want {
		got := s.String()
		if got != w {
			t.Fatalf("%d.String() = %q, want %q", int(s), got, w)
		}
		if seen[got] {
			t.Fatalf("two states share the log string %q — a census over logs could not tell them apart", got)
		}
		seen[got] = true
	}
}
