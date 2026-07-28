// FILE: platform/orchestration/datahelpers/fetch_provenance_test.go
//
// bugs_open/100 — provenance must come from the fetch, never from the model.

package datahelpers

import "testing"

func TestExtractFetchProvenanceReadsTheFetchRecord(t *testing.T) {
	tests := []struct {
		name      string
		scraped   interface{}
		wantURL   string
		wantName  string
		wantAt    string
		wantFound bool
	}{
		{
			// The shape the webscrape provider actually returns: url + captured_at
			// set beside the HTTP call.
			name: "provider result",
			scraped: map[string]interface{}{
				"url":              "https://www.arkvets.co.uk/terms",
				"captured_at":      "2026-07-28T16:04:05Z",
				"markdown_content": "…",
			},
			wantURL:   "https://www.arkvets.co.uk/terms",
			wantName:  "arkvets.co.uk", // www. stripped
			wantAt:    "2026-07-28T16:04:05Z",
			wantFound: true,
		},
		{
			// THE CONFIRMED LIVE SHAPE. The adapter replies
			// body:{success, body:{data: result}}; ResponseBody.Body is the inner
			// {data: result}, and that is what parseResponseBody stores under the
			// step's output_field. So collected_data["scraped_data"] is {data:{…}}.
			name: "coordinator-stored adapter response (live shape)",
			scraped: map[string]interface{}{
				"data": map[string]interface{}{
					"url":              "https://www.arkvets.co.uk/terms",
					"captured_at":      "2026-07-28T16:04:05Z",
					"markdown_content": "…",
				},
			},
			wantURL:   "https://www.arkvets.co.uk/terms",
			wantName:  "arkvets.co.uk",
			wantAt:    "2026-07-28T16:04:05Z",
			wantFound: true,
		},
		{
			// A deeper wrap, kept because the chain has several unwrap points. A
			// reader that only understood one shape would find nothing and report
			// "no provenance" — indistinguishable from a genuine absence.
			name: "adapter-wrapped result",
			scraped: map[string]interface{}{
				"body": map[string]interface{}{
					"data": map[string]interface{}{
						"url":         "https://example.com/about",
						"captured_at": "2026-07-28T16:00:00Z",
					},
				},
			},
			wantURL:   "https://example.com/about",
			wantName:  "example.com",
			wantAt:    "2026-07-28T16:00:00Z",
			wantFound: true,
		},
		{
			name:      "no fetch record",
			scraped:   map[string]interface{}{"markdown_content": "…"},
			wantFound: false,
		},
		{
			name:      "empty url is not provenance",
			scraped:   map[string]interface{}{"url": "   "},
			wantFound: false,
		},
		{
			name:      "nil input",
			scraped:   nil,
			wantFound: false,
		},
		{
			name:      "wrong type",
			scraped:   "https://example.com",
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prov, found := ExtractFetchProvenance(tc.scraped)
			if found != tc.wantFound {
				t.Fatalf("found = %v, want %v (prov: %+v)", found, tc.wantFound, prov)
			}
			if !tc.wantFound {
				if prov.SourceURL != "" {
					t.Errorf("not-found result must carry no URL, got %q", prov.SourceURL)
				}
				return
			}
			if prov.SourceURL != tc.wantURL {
				t.Errorf("SourceURL = %q, want %q", prov.SourceURL, tc.wantURL)
			}
			if prov.SourceName != tc.wantName {
				t.Errorf("SourceName = %q, want %q", prov.SourceName, tc.wantName)
			}
			if prov.CapturedAt != tc.wantAt {
				t.Errorf("CapturedAt = %q, want %q", prov.CapturedAt, tc.wantAt)
			}
			if prov.SourceType != SourceTypeWebsiteScrape {
				t.Errorf("SourceType = %q, want %q", prov.SourceType, SourceTypeWebsiteScrape)
			}
		})
	}
}

// TestExtractFetchProvenanceIgnoresModelClaims is the discriminating test for
// bugs_open/100: a model-authored source_url sitting in the verification result
// must never become provenance. The rejected fix (candidate 4) would pass every
// other test in this file — this is the only one that separates them.
func TestExtractFetchProvenanceIgnoresModelClaims(t *testing.T) {
	// What the LLM would emit if the prompt asked it to self-report, which is
	// exactly what 100 §"Why the obvious fix is WRONG" refuses.
	modelOutput := map[string]interface{}{
		"source_url":  "https://www.arkvets.co.uk/prices",
		"source_type": "website",
		"source_name": "Ark Veterinary Centre",
		"prices":      []interface{}{},
	}

	if prov, found := ExtractFetchProvenance(modelOutput); found {
		t.Errorf("model-claimed provenance was accepted as a fetch record: %+v", prov)
	}
}
