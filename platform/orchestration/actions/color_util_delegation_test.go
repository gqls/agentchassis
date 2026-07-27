// FILE: platform/orchestration/actions/color_util_delegation_test.go
//
// Guards the one mistake in the platform/colour extraction that no other test
// would catch: wiring isDarkHex and isDarkColor to each other's implementation.
//
// The two are similarly named, both return bool, both take a hex string, and
// they agree on almost every colour — so a swap compiles, passes the palette
// tests, passes the renderer's tests, and silently inverts the light/dark
// classification on exactly the mid-greys where the decision is marginal and
// therefore most likely to matter.
//
// I raised this as the residual risk in the council submission and then wrote a
// probe over #000000/#666666/#888888/#ffffff to check it — on which the two
// functions AGREE, so the probe could not have detected a swap at all. It is
// recorded here because "I checked it" with a check that has no failing branch
// is the recurring defect in this repository's own WRONG_CALLS log, and I did
// it while writing the guard against it.
//
// The real discriminating window is six greys wide, found by scanning:
//
//	#767676 … #7b7b7b   isDarkHex = false, isDarkColor = true
//
// isDarkHex asks "does white contrast better than black here?" (crossover ≈
// #777777); isDarkColor asks "is relative luminance < 0.2?" (crossover ≈
// #7c7c7c). Between those two points they legitimately disagree, and that gap
// is the only place a swap is visible.

package actions

import "testing"

func TestColourWrappersAreNotSwapped(t *testing.T) {
	// Inside the discriminating window: the two MUST disagree, in this
	// direction. Swap the delegations and both assertions fail.
	const midGrey = "#787878"
	if isDarkHex(midGrey) {
		t.Errorf("isDarkHex(%s) = true, want false — black contrasts better than white here, "+
			"so this is NOT a background needing light text. If this fails, isDarkHex is "+
			"probably wired to colour.IsPerceptuallyDark.", midGrey)
	}
	if !isDarkColor(midGrey) {
		t.Errorf("isDarkColor(%s) = false, want true — relative luminance is below 0.2. "+
			"If this fails, isDarkColor is probably wired to colour.IsDark.", midGrey)
	}

	// Outside the window they agree, which is why the window is the only
	// useful test. Kept as the control: if these ever disagree, the extraction
	// has changed something real rather than merely moved it.
	for _, tc := range []struct {
		hex  string
		dark bool
	}{
		{"#000000", true},
		{"#080E1C", true}, // fundamentallyai.com's page background
		{"#ffffff", false},
		{"#86ADDE", false}, // its repaired primary — must read as light
	} {
		if got := isDarkHex(tc.hex); got != tc.dark {
			t.Errorf("isDarkHex(%s) = %v, want %v", tc.hex, got, tc.dark)
		}
		if got := isDarkColor(tc.hex); got != tc.dark {
			t.Errorf("isDarkColor(%s) = %v, want %v", tc.hex, got, tc.dark)
		}
	}
}

// TestColourWrapperArithmeticIsUnchanged pins the values the renderer's own
// behaviour depends on, so a future change inside platform/colour cannot alter
// what this package computes without failing here as well as there.
func TestColourWrapperArithmeticIsUnchanged(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want float64
	}{
		{"#000000", "#ffffff", 21.0},
		{"#E4EAF2", "#ffffff", 1.21},  // the defect that shipped
		{"#E4EAF2", "#132239", 13.19}, // the same pair after the repair
	} {
		got, err := wcagContrastRatio(tc.a, tc.b)
		if err != nil {
			t.Fatalf("wcagContrastRatio(%s,%s): %v", tc.a, tc.b, err)
		}
		if diff := got - tc.want; diff > 0.02 || diff < -0.02 {
			t.Errorf("wcagContrastRatio(%s,%s) = %.2f, want ≈%.2f", tc.a, tc.b, got, tc.want)
		}
	}
	if _, err := wcagContrastRatio("not-a-colour", "#ffffff"); err == nil {
		t.Error("an unparseable colour must error rather than return a confident ratio")
	}
}
