// FILE: internal/adapters/webscrape/providers/firecrawl_scrape_payload_test.go
//
// Regression tests for bugs_open/101's [UNSETTLED] box: `only_main_content: false`
// was inexpressible on the /scrape path.
//
// These assert on the PAYLOAD WE SEND, deliberately. Asserting on scraped content
// would need a live Firecrawl and still could not tell "we sent false" apart from
// "Firecrawl happened to keep the footer" — the distinguishing input has to be
// observable in the request, so that is where the test looks.

package providers

import "testing"

// TestScrapePayloadOnlyMainContentIsExpressible is the regression proper.
// Before the fix the false case omitted the key, so Firecrawl's documented default
// (true — "excludes headers, navs, footers") applied and the caller got the exact
// opposite of what it asked for.
func TestScrapePayloadOnlyMainContentIsExpressible(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		wantPresent bool
		wantValue   bool
	}{
		{
			// The failing case. Three live steps pass exactly this:
			// site-scraper/scrape_site, site-adoption-agent/fetch_primary_css,
			// website-capture-firecrawl/scrape_main_page.
			name:        "explicit false is sent, not dropped",
			config:      map[string]interface{}{"only_main_content": false},
			wantPresent: true,
			wantValue:   false,
		},
		{
			// The positive control. If this ever stops passing, the test has
			// stopped exercising the key at all and the false case above would
			// look green for the wrong reason.
			name:        "explicit true is sent",
			config:      map[string]interface{}{"only_main_content": true},
			wantPresent: true,
			wantValue:   true,
		},
		{
			// Unchanged behaviour: say nothing, send nothing, let Firecrawl
			// default. This is what stops the fix being a behaviour change for
			// the ~1,150 live steps that never mention the key.
			name:        "absent stays absent",
			config:      map[string]interface{}{},
			wantPresent: false,
		},
		{
			// A non-bool must not be coerced into a silent true/false. Presence
			// of the WRONG TYPE is not presence of an instruction.
			name:        "non-bool is ignored, not coerced",
			config:      map[string]interface{}{"only_main_content": "false"},
			wantPresent: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := buildScrapePayload("https://example.com", tc.config)

			got, present := payload["onlyMainContent"]
			if present != tc.wantPresent {
				t.Fatalf("onlyMainContent present = %v, want %v (payload: %#v)",
					present, tc.wantPresent, payload)
			}
			if !tc.wantPresent {
				return
			}
			if got != tc.wantValue {
				t.Errorf("onlyMainContent = %v, want %v", got, tc.wantValue)
			}
		})
	}
}

// TestScrapePayloadPreservesExistingContract guards the extraction of
// buildScrapePayload out of Scrape: everything that was in the payload before must
// still be there, or the refactor has quietly changed behaviour for every caller.
func TestScrapePayloadPreservesExistingContract(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		payload := buildScrapePayload("https://example.com", map[string]interface{}{})

		if payload["url"] != "https://example.com" {
			t.Errorf("url = %v", payload["url"])
		}

		formats, ok := payload["formats"].([]interface{})
		if !ok {
			t.Fatalf("formats missing or wrong type: %#v", payload["formats"])
		}
		// 4 defaults + the screenshot object, which is on by default.
		if len(formats) != 5 {
			t.Errorf("default formats = %d entries, want 5: %#v", len(formats), formats)
		}
		shot, ok := formats[len(formats)-1].(map[string]interface{})
		if !ok || shot["type"] != "screenshot" {
			t.Errorf("expected trailing screenshot format object, got %#v", formats[len(formats)-1])
		}

		if _, present := payload["waitFor"]; present {
			t.Error("waitFor should be absent when unset")
		}
		if _, present := payload["maxAge"]; present {
			t.Error("maxAge should be absent when unset")
		}
	})

	t.Run("caller overrides", func(t *testing.T) {
		payload := buildScrapePayload("https://example.com", map[string]interface{}{
			"formats":            []interface{}{"markdown"},
			"capture_screenshot": false,
			"wait_for":           3000,
			"max_age":            float64(0),
		})

		formats, _ := payload["formats"].([]interface{})
		if len(formats) != 1 || formats[0] != "markdown" {
			t.Errorf("formats = %#v, want [markdown] with no screenshot appended", formats)
		}
		if payload["waitFor"] != 3000 {
			t.Errorf("waitFor = %v, want 3000", payload["waitFor"])
		}
		// maxAge 0 means "force a fresh scrape" — a real instruction, and the
		// same omit-when-falsy trap as only_main_content if written carelessly.
		if got, present := payload["maxAge"]; !present || got != 0 {
			t.Errorf("maxAge = %v present=%v, want 0 present=true", got, present)
		}
	})
}
