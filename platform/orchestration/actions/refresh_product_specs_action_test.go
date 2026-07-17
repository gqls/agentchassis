// FILE: platform/orchestration/actions/refresh_product_specs_action_test.go
//
// selectSpecRegion is the difference between the refresher working and the
// refresher returning {} for every product: the model can only be shown
// specRegionMaxChars of a page (~5.5 min of CPU inference), so WHICH chars get
// picked decides the outcome. These tests pin the behaviour that matters —
// a spec table buried under nav/marketing must still reach the model.

package actions

import (
	"strings"
	"testing"
)

// realisticPage mimics the shape of the manufacturer pages this action scrapes:
// a long nav/cookie/marketing preamble, the spec table far down, then footer.
func realisticPage(preambleChars int) string {
	var b strings.Builder
	b.WriteString("# Schunk EGP 40-N-S-B\n\n")
	for b.Len() < preambleChars {
		b.WriteString("Accept all cookies to continue browsing our product catalogue.\n")
		b.WriteString("Home | Products | Gripping Systems | Service | Careers | Contact us\n")
		b.WriteString("Discover our award-winning automation portfolio and success stories.\n")
	}
	b.WriteString(`
## Technical data

| Description | Value |
| --- | --- |
| Stroke per jaw | 3 mm |
| Gripping force | 140 N |
| Recommended workpiece weight | 0.7 kg |
| Weight | 0.19 kg |
| IP protection class | IP40 |
| Interface | Digital I/O |
| Nominal voltage | 24 V DC |
`)
	b.WriteString(strings.Repeat("Read more about our company history and vision.\n", 40))
	return b.String()
}

func TestSelectSpecRegion_FindsTableBuriedBelowTheFold(t *testing.T) {
	page := realisticPage(8000)
	got := selectSpecRegion(page, specRegionMaxChars)

	if len(got) > specRegionMaxChars {
		t.Fatalf("region is %d chars, over the %d budget", len(got), specRegionMaxChars)
	}
	// The whole point: the spec values must survive the cut.
	for _, want := range []string{"140 N", "3 mm", "0.19 kg", "IP40", "24 V DC"} {
		if !strings.Contains(got, want) {
			t.Errorf("spec value %q missing from selected region:\n%s", want, got)
		}
	}
	// And the head-of-page slice — the old behaviour — must NOT be what we send.
	if strings.Contains(got, "Accept all cookies") && !strings.Contains(got, "Gripping force") {
		t.Error("selected the cookie banner instead of the spec table")
	}
}

// The cases below are the REAL before/after pairs from the first working live
// run (robot-hands, 2026-07-17). The refresher invented nothing, but it traded
// hand-verified qualifiers for barer page-literal restatements — "6 mm per jaw"
// became "6 mm", which halves the stated stroke of a parallel gripper.
func TestSpecValueIsRestatement_LiveRegressions(t *testing.T) {
	cases := []struct {
		name              string
		existing, updated string
		wantSuppressed    bool
	}{
		// Degradations observed live — must be suppressed.
		{"stroke loses per-jaw", "6 mm per jaw", "6 mm", true},
		{"zimmer stroke loses per-jaw", "10 mm per jaw", "10 mm", true},
		{"payload loses meaning", "0.15 kg (recommended workpiece weight)", "0.15 kg", true},
		{"voltage loses DC", "24 V DC", "24 V", true},
		{"interface loses IO-Link", "I/O (IO-Link option)", "I/O", true},

		// Real updates — must still be written.
		{"genuine value change", "30 N", "45 N", false},
		{"enrichment adds equivalent", "11 kg", "11 kg (24.3 lb)", false},
		{"identical value", "30 N", "30 N", false},
		{"dash restyle is not a loss", "20–235 N", "20 to 235 N", false},
		{"different unit entirely", "IP30", "IP67", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := specValueIsRestatement(c.existing, c.updated); got != c.wantSuppressed {
				t.Errorf("specValueIsRestatement(%q, %q) = %v, want %v",
					c.existing, c.updated, got, c.wantSuppressed)
			}
		})
	}
}

func TestSpecValueIsRestatement_IgnoresSpacingAndDashStyle(t *testing.T) {
	// "20–235 N" and "20 - 235  n" are the same value written differently; the
	// richer-vs-barer test must not be fooled into calling that a loss.
	if specValueIsRestatement("20–235 N", "20 - 235  n") {
		t.Error("same value in a different style must not count as a restatement")
	}
}

func TestSelectSpecRegion_ShortPageUntouched(t *testing.T) {
	page := "| Gripping force | 140 N |\n| Weight | 0.19 kg |"
	if got := selectSpecRegion(page, specRegionMaxChars); got != page {
		t.Errorf("a page under budget must pass through unchanged, got:\n%s", got)
	}
}

func TestSelectSpecRegion_NeverExceedsBudget(t *testing.T) {
	// Signal-free page: degrades to head-of-page, but must still respect the
	// budget — an over-budget prompt is what blows the inference timeout.
	page := strings.Repeat("lorem ipsum dolor sit amet consectetur adipiscing\n", 500)
	got := selectSpecRegion(page, specRegionMaxChars)
	if len(got) > specRegionMaxChars {
		t.Fatalf("region is %d chars, over the %d budget", len(got), specRegionMaxChars)
	}
	if got == "" {
		t.Error("expected a head-of-page fallback, got nothing")
	}
}

func TestSelectSpecRegion_SingleLineLongerThanBudget(t *testing.T) {
	// Firecrawl can emit a page whose body is one unbroken line. Line-wise
	// selection would pick nothing here and send the model an empty page.
	page := "Schunk EGP 40-N-S-B " + strings.Repeat("gripping force 140 N stroke 3 mm ", 200)
	got := selectSpecRegion(page, specRegionMaxChars)
	if got == "" {
		t.Fatal("returned an empty region — the model would be asked about no text at all")
	}
	if len(got) > specRegionMaxChars {
		t.Fatalf("region is %d chars, over the %d budget", len(got), specRegionMaxChars)
	}
	if !strings.Contains(got, "140 N") {
		t.Errorf("spec content missing from fallback slice:\n%s", got)
	}
}

func TestSelectSpecRegion_PrefersDensestSpecRegion(t *testing.T) {
	// Two candidate regions: one with a single stray unit, one with a real
	// table. The table must win.
	page := "The gripper weighs about 5 kg in its shipping crate.\n" +
		strings.Repeat("Marketing copy about reliability and precision.\n", 60) +
		"| Stroke per jaw | 3 mm |\n| Gripping force | 140 N |\n| Nominal voltage | 24 V DC |\n" +
		"| IP protection class | IP40 |\n| Weight | 0.19 kg |\n" +
		strings.Repeat("More marketing copy about our global support network.\n", 60)

	got := selectSpecRegion(page, specRegionMaxChars)
	if !strings.Contains(got, "Gripping force") {
		t.Errorf("did not select the densest spec region:\n%s", got)
	}
}
