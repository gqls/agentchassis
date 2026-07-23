package actions

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// The export must never fall back to a default domain: a misconfigured task
// must fail closed rather than publish to a site we did not intend.
func TestParseMedExportConfigHasNoDefaultDomain(t *testing.T) {
	ec := parseMedExportConfig(map[string]interface{}{})
	if ec.Domain != "" {
		t.Fatalf("empty config produced default domain %q; exports must require an explicit domain", ec.Domain)
	}
}

func TestParseMedExportConfigUsesExplicitDomain(t *testing.T) {
	ec := parseMedExportConfig(map[string]interface{}{"domain": "example.uk"})
	if ec.Domain != "example.uk" {
		t.Fatalf("explicit domain not honoured: got %q", ec.Domain)
	}
}

// Publication policy: no price without provenance (source URL + capture date).
// The filter must DROP the deficient rows — exercising the failing branch, not
// just the happy path — and count them rather than vanishing them silently.
func TestFilterMedExportProvenanceDropsDeficientRows(t *testing.T) {
	good := medPriceExportRow{RetailerURL: "https://retailer.example/product", CollectedAt: time.Now(), Price: 9.99}
	noURL := medPriceExportRow{RetailerURL: "", CollectedAt: time.Now(), Price: 1.23}
	blankURL := medPriceExportRow{RetailerURL: "   ", CollectedAt: time.Now(), Price: 4.56}
	noDate := medPriceExportRow{RetailerURL: "https://retailer.example/other", Price: 7.89}

	kept, skipped := filterMedExportProvenance([]medPriceExportRow{good, noURL, blankURL, noDate})
	if len(kept) != 1 || kept[0].Price != good.Price {
		t.Fatalf("expected exactly the provenanced row to survive; kept=%d %+v", len(kept), kept)
	}
	if skipped != 3 {
		t.Fatalf("expected 3 skipped (no url, blank url, zero date); got %d", skipped)
	}
}

// The skip count must be visible in the published metadata even when zero —
// its PRESENCE is the deploy-time proof the provenance gate is in the binary.
func TestMedExportMetadataAlwaysCarriesProvenanceSkipCount(t *testing.T) {
	j, err := buildMedExportMetadata(nil, nil, medExportConfig{Domain: "example.uk"}, 0)
	if err != nil {
		t.Fatalf("metadata build failed: %v", err)
	}
	if !strings.Contains(j, `"skipped_missing_provenance": 0`) {
		t.Fatalf("metadata missing always-present skipped_missing_provenance field: %s", j)
	}
}

// The invented figure family must not reappear in any export output
// (LEGAL record 2026-07-15: vet_price_est was fabricated).
func TestMedExportOptionCarriesNoTypicalVetPrice(t *testing.T) {
	b, err := json.Marshal(medExportOption{Retailer: "R", URL: "https://r.example/p", Price: 1, CollectedAt: "2026-07-23"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "typical_vet_price") {
		t.Fatalf("typical_vet_price leaked back into export output: %s", b)
	}
}
