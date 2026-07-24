package actions

import (
	"strings"
	"testing"
)

// Parse-fidelity guard tests (bugs_open/061): a stored price must appear
// verbatim in the markdown retained as scrape evidence.

func TestMedPriceLiteralInMarkdown(t *testing.T) {
	cases := []struct {
		name     string
		markdown string // raw; commas stripped as the guard does
		price    float64
		want     bool
	}{
		{"two-decimal present", "Advocate for cats £17.75 per pipette", 17.75, true},
		{"two-decimal absent", "Advocate for cats £17.75 per pipette", 17.95, false},
		{"jammed against text", "£29.75Save £2.00", 29.75, true},
		{"embedded in longer number rejected", "special offer £117.95 today", 17.95, false},
		{"longer number, decimals continue", "price £17.950 units", 17.95, false},
		{"whole pound rendered without pence", "just £17 for members", 17.00, true},
		{"whole pound must not match pence price", "just £17.95 for members", 17.00, false},
		{"whole pound must be £-prefixed", "bottle size 17ml in stock", 17.00, false},
		{"single decimal rendering", "reduced to £17.4 this week", 17.40, true},
		{"single decimal must not match longer", "reduced to £17.45 this week", 17.40, false},
		{"comma-grouped thousands", "professional pack £1,234.56 only", 1234.56, true},
		{"empty markdown", "", 17.95, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mdNoCommas := strings.ReplaceAll(tc.markdown, ",", "")
			got := medPriceLiteralInMarkdown(mdNoCommas, tc.price)
			if got != tc.want {
				t.Errorf("medPriceLiteralInMarkdown(%q, %v) = %v, want %v",
					tc.markdown, tc.price, got, tc.want)
			}
		})
	}
}

// TestMedFilterVariantsByEvidence_FabricatedDropped reproduces the
// bugs_open/061 shape: the page (a category page) shows sale prices
// £17.75 / £29.75, the LLM fallback returned £17.95 / £29.95 with invented
// TVPs. Both variants must be dropped, none stored.
func TestMedFilterVariantsByEvidence_FabricatedDropped(t *testing.T) {
	markdown := `# Advocate®
Free standard delivery over £49. £4.49 under £49.
Advocate Spot-On Solution for Large Cats 80mg/8mg (4kg-8kg)
£17.75
Advocate Spot-On for Small Dogs
£29.75Save £2.00`

	fabricated := []medExtractedVariant{
		{SizeVariant: "Large Cats 80mg/8mg (4kg-8kg)", Price: 17.95, TypicalVetPrice: 23.60, InStock: true},
		{SizeVariant: "Small Dogs 10kg Pack of 3", Price: 29.95, TypicalVetPrice: 48.90, InStock: false},
	}

	kept, dropped := medFilterVariantsByEvidence(fabricated, markdown)
	if len(kept) != 0 {
		t.Errorf("fabricated variants kept: %+v", kept)
	}
	if len(dropped) != 2 {
		t.Errorf("dropped = %d, want 2", len(dropped))
	}
}

func TestMedFilterVariantsByEvidence_GenuineKept(t *testing.T) {
	markdown := `10ml bottle Price: £3.89 Regular Price: £14.09 (TVP) Save £10.20
100ml bottle Price: £16.73`

	variants := []medExtractedVariant{
		{SizeVariant: "10ml bottle", Price: 3.89, TypicalVetPrice: 14.09, InStock: true},
		{SizeVariant: "100ml bottle", Price: 16.73, InStock: true},
	}

	kept, dropped := medFilterVariantsByEvidence(variants, markdown)
	if len(dropped) != 0 {
		t.Errorf("genuine variants dropped: %+v", dropped)
	}
	if len(kept) != 2 {
		t.Fatalf("kept = %d, want 2", len(kept))
	}
	if kept[0].TypicalVetPrice != 14.09 {
		t.Errorf("verifiable TVP zeroed: got %v, want 14.09", kept[0].TypicalVetPrice)
	}
}

// A price that verifies but a TVP that does not: the variant is kept, the
// unverifiable TVP is zeroed rather than published.
func TestMedFilterVariantsByEvidence_UnverifiableTVPZeroed(t *testing.T) {
	markdown := `Metacam 10ml Price: £3.99`

	variants := []medExtractedVariant{
		{SizeVariant: "10ml", Price: 3.99, TypicalVetPrice: 12.50, InStock: true},
	}

	kept, dropped := medFilterVariantsByEvidence(variants, markdown)
	if len(dropped) != 0 || len(kept) != 1 {
		t.Fatalf("kept = %d, dropped = %d, want 1/0", len(kept), len(dropped))
	}
	if kept[0].TypicalVetPrice != 0 {
		t.Errorf("unverifiable TVP not zeroed: got %v", kept[0].TypicalVetPrice)
	}
	if kept[0].Price != 3.99 {
		t.Errorf("price mutated: got %v", kept[0].Price)
	}
}

// The comma-grouped case end-to-end through the filter (it strips commas
// from the markdown before matching).
func TestMedFilterVariantsByEvidence_CommaGrouped(t *testing.T) {
	markdown := `Professional dispensary pack Price: £1,234.56`

	variants := []medExtractedVariant{
		{SizeVariant: "dispensary pack", Price: 1234.56, InStock: true},
	}

	kept, dropped := medFilterVariantsByEvidence(variants, markdown)
	if len(kept) != 1 || len(dropped) != 0 {
		t.Errorf("kept = %d, dropped = %d, want 1/0", len(kept), len(dropped))
	}
}
