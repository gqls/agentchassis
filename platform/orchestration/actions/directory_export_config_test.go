package actions

import "testing"

// Fail-closed defaults: no domain, no vertical, and attributed (scraped)
// per-business price publication OFF unless a site's config turns it on.
func TestParseDirectoryExportConfigFailClosedDefaults(t *testing.T) {
	ec := parseDirectoryExportConfig(map[string]interface{}{})
	if ec.Domain != "" {
		t.Fatalf("default domain must be empty, got %q", ec.Domain)
	}
	if ec.VerticalSlug != "" {
		t.Fatalf("default vertical must be empty, got %q", ec.VerticalSlug)
	}
	if ec.AttributedPrices {
		t.Fatal("attributed_prices must default to false")
	}
	if ec.AggMinN < 3 {
		t.Fatalf("aggregates min_n default must be >= 3, got %d", ec.AggMinN)
	}
}

func TestParseDirectoryExportConfigExplicit(t *testing.T) {
	ec := parseDirectoryExportConfig(map[string]interface{}{
		"vertical": "veterinary",
		"domain":   "vetcomparison.uk",
		"business_type_ilike": "%vet%",
		"outputs": map[string]interface{}{
			"directory_filename": "vet-full-index.json",
			"attributed_prices":  true,
			"aggregates":         map[string]interface{}{"min_n": float64(3)},
		},
	})
	if ec.VerticalSlug != "veterinary" || ec.Domain != "vetcomparison.uk" {
		t.Fatalf("explicit vertical/domain not honoured: %+v", ec)
	}
	if ec.DirectoryFilename != "vet-full-index.json" {
		t.Fatalf("directory filename not honoured: %q", ec.DirectoryFilename)
	}
	if !ec.AttributedPrices {
		t.Fatal("explicit attributed_prices=true not honoured")
	}
	if ec.AggMinN != 3 {
		t.Fatalf("min_n not honoured: %d", ec.AggMinN)
	}
}
