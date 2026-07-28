package datahelpers

import "testing"

// citedSource builds a valid citation source map for an observation.
func citedSource(url, quote string) map[string]interface{} {
	return map[string]interface{}{
		"citation": map[string]interface{}{
			"publisher": "Ofwat",
			"title":     "Final determination",
			"url":       url,
			"quote":     quote,
			"accessed":  "2026-07-28",
		},
	}
}

func goodSeries() *EvidenceFact {
	return &EvidenceFact{
		ID:    "SER-net-debt",
		Kind:  KindSeries,
		Claim: "Net debt at each financial year end",
		Observations: []Observation{
			{AsOf: "2024-03-31", Value: 15.2, VerifiedAt: "2026-07-28",
				Source: citedSource("https://example.org/ar2024", "net debt of 15.2")},
			{AsOf: "2023-03-31", Value: 14.0, VerifiedAt: "2026-07-28",
				Source: citedSource("https://example.org/ar2023", "net debt of 14.0")},
		},
	}
}

func TestSeriesObservationsSortedByAsOf(t *testing.T) {
	f := goodSeries()
	got := f.SeriesObservations()
	if len(got) != 2 {
		t.Fatalf("want 2 observations, got %d", len(got))
	}
	if got[0].AsOf != "2023-03-31" || got[1].AsOf != "2024-03-31" {
		t.Fatalf("not sorted by as_of: %q then %q", got[0].AsOf, got[1].AsOf)
	}
	// the caller must not be able to mutate the fact through the returned slice
	got[0].Value = 999
	if f.Observations[1].Value == 999 {
		t.Fatal("SeriesObservations returned an aliased slice")
	}
}

func TestMixedGranularityStillSortsChronologically(t *testing.T) {
	f := &EvidenceFact{Kind: KindSeries, Observations: []Observation{
		{AsOf: "2025", Value: 3, Source: citedSource("https://e/3", "q")},
		{AsOf: "2024-03", Value: 1, Source: citedSource("https://e/1", "q")},
		{AsOf: "2024-03-31", Value: 2, Source: citedSource("https://e/2", "q")},
	}}
	got := f.SeriesObservations()
	want := []string{"2024-03", "2024-03-31", "2025"}
	for i, w := range want {
		if got[i].AsOf != w {
			t.Fatalf("position %d: want %q, got %q", i, w, got[i].AsOf)
		}
	}
}

func TestValidateSeriesAcceptsAFullyCitedSeries(t *testing.T) {
	if p := goodSeries().ValidateSeries(); len(p) != 0 {
		t.Fatalf("a fully cited series must validate, got %v", p)
	}
}

// The load-bearing rejection: no inheritance. A parent fact with a perfectly
// good source must NOT rescue an observation that has none, because that is
// exactly how an interpolated point enters looking like data.
func TestObservationNeverInheritsTheParentSource(t *testing.T) {
	f := goodSeries()
	f.Source = EvidenceSource{Artifact: "the parent has impeccable provenance"}
	f.Observations = append(f.Observations, Observation{
		AsOf: "2025-03-31", Value: 16.9, VerifiedAt: "2026-07-28",
	})
	problems := f.ValidateSeries()
	if len(problems) == 0 {
		t.Fatal("an observation with no source of its own must be rejected")
	}
	found := false
	for _, p := range problems {
		if p.Index == 2 {
			found = true
		}
	}
	if !found {
		t.Fatalf("the unsourced observation was not the one rejected: %v", problems)
	}
}

func TestValidateSeriesRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		fact *EvidenceFact
	}{
		{"single observation is not a series", &EvidenceFact{Kind: KindSeries,
			Observations: []Observation{{AsOf: "2024", Value: 1, Source: citedSource("https://e/1", "q")}}}},
		{"unparseable as_of", &EvidenceFact{Kind: KindSeries, Observations: []Observation{
			{AsOf: "March 2024", Value: 1, Source: citedSource("https://e/1", "q")},
			{AsOf: "2025", Value: 2, Source: citedSource("https://e/2", "q")}}}},
		{"duplicate as_of", &EvidenceFact{Kind: KindSeries, Observations: []Observation{
			{AsOf: "2024", Value: 1, Source: citedSource("https://e/1", "q")},
			{AsOf: "2024", Value: 2, Source: citedSource("https://e/2", "q")}}}},
		{"citation missing its quote", &EvidenceFact{Kind: KindSeries, Observations: []Observation{
			{AsOf: "2024", Value: 1, Source: map[string]interface{}{"citation": map[string]interface{}{
				"url": "https://e/1", "publisher": "Ofwat"}}},
			{AsOf: "2025", Value: 2, Source: citedSource("https://e/2", "q")}}}},
		{"present but empty source object", &EvidenceFact{Kind: KindSeries, Observations: []Observation{
			{AsOf: "2024", Value: 1, Source: map[string]interface{}{"note": "trust me"}},
			{AsOf: "2025", Value: 2, Source: citedSource("https://e/2", "q")}}}},
		{"not a series at all", &EvidenceFact{Kind: "metric"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if p := tc.fact.ValidateSeries(); len(p) == 0 {
				t.Fatal("expected rejection, got none")
			}
		})
	}
}

func TestSQLSourcedObservationIsAccepted(t *testing.T) {
	f := &EvidenceFact{Kind: KindSeries, Observations: []Observation{
		{AsOf: "2026-07-01", Value: 10, Source: map[string]interface{}{"sql": "SELECT count(*) FROM sites"}},
		{AsOf: "2026-07-28", Value: 12, Source: map[string]interface{}{"sql": "SELECT count(*) FROM sites"}},
	}}
	if p := f.ValidateSeries(); len(p) != 0 {
		t.Fatalf("an internally-measured series must validate, got %v", p)
	}
}

// The integration that makes the honesty layer and the rendering layer agree:
// a value plotted from a series must not be reported as an unregistered number.
func TestSeriesValuesAreRegisteredNumbers(t *testing.T) {
	eb := &EvidenceBase{Facts: []EvidenceFact{{
		ID: "SER-debt", Kind: KindSeries, Claim: "Net debt at year end",
		ContextTerms: []string{"net debt"},
		Observations: []Observation{
			{AsOf: "2023-03-31", Value: 14, Source: citedSource("https://e/1", "q")},
			{AsOf: "2024-03-31", Value: 15.2, Source: citedSource("https://e/2", "q")},
		},
	}}}

	if !eb.numberSupported(15.2, "our net debt reached 15.2 at year end") {
		t.Fatal("a value present in the series must be supported")
	}
	if !eb.numberSupported(14, "net debt of 14 in the prior year") {
		t.Fatal("every observation registers its value, not just the latest")
	}
	if eb.numberSupported(99, "net debt of 99") {
		t.Fatal("a value absent from the series must NOT be supported")
	}
	// context terms scope the series exactly as they scope an ordinary fact
	if eb.numberSupported(15.2, "we have 15.2 million customers") {
		t.Fatal("context terms must stop a series blanket-supporting unrelated claims")
	}
}

// A series must not behave like a `gte` fact. Over a long series that would
// support very nearly every number on the page.
func TestSeriesDoesNotBlanketSupportSmallerValues(t *testing.T) {
	eb := &EvidenceBase{Facts: []EvidenceFact{{
		ID: "SER-debt", Kind: KindSeries, Tolerance: "gte",
		ContextTerms: []string{"net debt"},
		Observations: []Observation{
			{AsOf: "2023", Value: 100, Source: citedSource("https://e/1", "q")},
			{AsOf: "2024", Value: 200, Source: citedSource("https://e/2", "q")},
		},
	}}}
	if eb.numberSupported(37, "net debt of 37") {
		t.Fatal("a series must match exactly, even when the fact carries a gte tolerance")
	}
}

// A date written the way British English writes it must not be read as a
// business figure. The composite-token rule catches 2026-07-28; it cannot see
// "28 July 2026", which is the form the platform's own convention produces.
func TestWrittenOutDatesAreNotBusinessNumbers(t *testing.T) {
	eb := &EvidenceBase{}
	excluded := []string{
		"Verified on 28 July 2026. If we ever add analytics we will say so.",
		"The plan was sanctioned on 1 January 2025 by the court.",
		"Published 3rd March 2024 by the regulator.",
		"Reported July 28, 2026 in the annual accounts.",
	}
	for _, block := range excluded {
		if f := eb.ScanUnregisteredNumbers([]string{block}, ClaimSurface{}); len(f) != 0 {
			t.Errorf("date in %q flagged as a business number: %+v", block, f)
		}
	}
}

// The exclusion must not swallow a real figure that merely sits near a month.
func TestFigureNearAMonthIsStillScanned(t *testing.T) {
	eb := &EvidenceBase{}
	block := "We served 450 clients in the March quarter."
	if f := eb.ScanUnregisteredNumbers([]string{block}, ClaimSurface{}); len(f) == 0 {
		t.Fatal("a real business figure next to a month name must still be scanned")
	}
}

// Council round 1 on corr da40ddf0 found this hole: ValidateSeries enforced the
// per-observation source rule but numberSupported never called it, so an
// unsourced point still registered its value. The guarantee has to hold at the
// gate that decides, not only in a validator a caller may never run.
func TestUnsourcedObservationDoesNotRegisterItsValue(t *testing.T) {
	eb := &EvidenceBase{Facts: []EvidenceFact{{
		ID: "SER-debt", Kind: KindSeries, ContextTerms: []string{"net debt"},
		Source: EvidenceSource{Artifact: "the parent has impeccable provenance"},
		Observations: []Observation{
			{AsOf: "2023", Value: 14, Source: citedSource("https://e/1", "q")},
			{AsOf: "2024", Value: 99}, // no source at all
			{AsOf: "2025", Value: 77, Source: map[string]interface{}{"note": "trust me"}}, // present but empty
			{AsOf: "2026", Value: 55, Source: map[string]interface{}{ // citation missing its quote
				"citation": map[string]interface{}{"url": "https://e/x", "publisher": "Ofwat"}}},
		},
	}}}
	if !eb.numberSupported(14, "net debt of 14") {
		t.Fatal("the properly cited observation must still be supported")
	}
	for _, bad := range []float64{99, 77, 55} {
		if eb.numberSupported(bad, "net debt of something") {
			t.Errorf("value %v from an unsourced observation must NOT be registered", bad)
		}
	}
}
