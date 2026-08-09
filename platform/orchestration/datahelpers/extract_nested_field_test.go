package datahelpers

import (
	"encoding/json"
	"testing"
)

// ExtractNestedField had no tests at all before this file, despite being the
// resolver behind ~460 call sites and every config-side dotted path in the
// fleet. The first block below is a characterisation harness: it pins the
// behaviour that already existed on 2026-08-09, so that the array-index
// addition can be shown to change nothing that previously resolved.
//
// The rule the array branch must not break: MAP ACCESS WINS. A map carrying a
// literal "0" key resolved before the change and must resolve identically
// after it. The new branch is reachable only where the old walk returned nil.

// --- existing behaviour (the safety net) --------------------------------

func TestExtractNestedField_ExistingBehaviour(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		path string
		want interface{}
	}{
		{
			name: "top-level key",
			data: map[string]interface{}{"domain": "example.co.uk"},
			path: "domain",
			want: "example.co.uk",
		},
		{
			name: "nested maps",
			data: map[string]interface{}{
				"input_data": map[string]interface{}{
					"site": map[string]interface{}{"domain": "example.co.uk"},
				},
			},
			path: "input_data.site.domain",
			want: "example.co.uk",
		},
		{
			name: "response auto-unwrap at one level",
			data: map[string]interface{}{
				"site_plan": map[string]interface{}{
					"status":   "complete",
					"response": map[string]interface{}{"needs_logo": true},
				},
			},
			path: "site_plan.needs_logo",
			want: true,
		},
		{
			name: "response auto-unwrap at two levels",
			data: map[string]interface{}{
				"site_plan": map[string]interface{}{
					"response": map[string]interface{}{
						"validated_plan": map[string]interface{}{"needs_logo": false},
					},
				},
			},
			path: "site_plan.validated_plan.needs_logo",
			want: false,
		},
		{
			name: "direct access beats the response wrapper",
			data: map[string]interface{}{
				"site_plan": map[string]interface{}{
					"needs_logo": "direct",
					"response":   map[string]interface{}{"needs_logo": "wrapped"},
				},
			},
			path: "site_plan.needs_logo",
			want: "direct",
		},
		{
			name: "missing key returns nil",
			data: map[string]interface{}{"input_data": map[string]interface{}{"domain": "x"}},
			path: "input_data.missing",
			want: nil,
		},
		{
			name: "missing intermediate key returns nil",
			data: map[string]interface{}{"input_data": map[string]interface{}{"domain": "x"}},
			path: "no_such.domain",
			want: nil,
		},
		{
			name: "empty path returns nil",
			data: map[string]interface{}{"domain": "x"},
			path: "",
			want: nil,
		},
		{
			name: "nil map returns nil",
			data: nil,
			path: "domain",
			want: nil,
		},
		{
			name: "nil map with empty path returns nil",
			data: nil,
			path: "",
			want: nil,
		},
		{
			name: "walking past a scalar returns nil",
			data: map[string]interface{}{"domain": "example.co.uk"},
			path: "domain.tld",
			want: nil,
		},
		{
			name: "response wrapper is not itself a map",
			data: map[string]interface{}{
				"site_plan": map[string]interface{}{"response": "not-a-map"},
			},
			path: "site_plan.needs_logo",
			want: nil,
		},
		{
			name: "whole subtree is returned when the path stops early",
			data: map[string]interface{}{
				"input_data": map[string]interface{}{"domain": "x"},
			},
			path: "input_data",
			want: map[string]interface{}{"domain": "x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractNestedField(tt.data, tt.path)
			if !jsonEqual(got, tt.want) {
				t.Fatalf("ExtractNestedField(%q) = %#v, want %#v", tt.path, got, tt.want)
			}
		})
	}
}

// TestExtractNestedField_LiteralZeroKeyOnMapWins is the load-bearing
// regression guard for the array-index addition: a MAP whose key happens to be
// the string "0" must keep resolving through map access, and must never be
// read as an index. Both directions are asserted — the map path resolves, and
// a sibling array in the same payload resolves by index — so the test cannot
// pass by the array branch simply being absent.
func TestExtractNestedField_LiteralZeroKeyOnMapWins(t *testing.T) {
	data := map[string]interface{}{
		"by_key": map[string]interface{}{
			"0": map[string]interface{}{"url": "from-map-key"},
			"1": map[string]interface{}{"url": "from-map-key-1"},
		},
		"by_index": []interface{}{
			map[string]interface{}{"url": "from-array-index"},
		},
	}

	if got := ExtractNestedField(data, "by_key.0.url"); got != "from-map-key" {
		t.Fatalf("map with literal \"0\" key: got %#v, want %q", got, "from-map-key")
	}
	if got := ExtractNestedField(data, "by_key.1.url"); got != "from-map-key-1" {
		t.Fatalf("map with literal \"1\" key: got %#v, want %q", got, "from-map-key-1")
	}
	if got := ExtractNestedField(data, "by_index.0.url"); got != "from-array-index" {
		t.Fatalf("array index: got %#v, want %q", got, "from-array-index")
	}
}

// TestExtractNestedField_ResponseUnwrapStaysMapOnly pins the auto-unwrap as
// map-only. A list under "response" must not be silently indexed into by the
// unwrap path; the unwrap looks up a KEY and nothing else.
func TestExtractNestedField_ResponseUnwrapStaysMapOnly(t *testing.T) {
	data := map[string]interface{}{
		"search_results": map[string]interface{}{
			"response": []interface{}{
				map[string]interface{}{"url": "should-not-be-reachable"},
			},
		},
	}
	if got := ExtractNestedField(data, "search_results.0.url"); got != nil {
		t.Fatalf("unwrap must not index a list under \"response\": got %#v, want nil", got)
	}
}

// --- new behaviour: array indexing --------------------------------------

// vetSearchPayload reproduces the shape vet-practice-verifier actually
// receives. The websearch adapter builds
//
//	{"results": [...], "query": ..., "total": ..., "provider": ..., "search_type": ..., "fallbacks": ...}
//
// (internal/adapters/websearch/adapter.go, sendResponse) and the chassis stores
// it under the call_agent "response" wrapper, so the configured path
// "search_results.results.0.url" must cross: map -> response unwrap -> array ->
// index -> map.
//
// Result element fields are providers.SearchResult's json tags: title, url,
// snippet, published_at, source.
func vetSearchPayload() map[string]interface{} {
	return map[string]interface{}{
		"search_results": map[string]interface{}{
			"status": "complete",
			"response": map[string]interface{}{
				"results": []interface{}{
					map[string]interface{}{
						"title":   "Oakwood Veterinary Surgery",
						"url":     "https://oakwoodvets.co.uk",
						"snippet": "Small animal practice in Leeds",
						"source":  "brave",
					},
					map[string]interface{}{
						"title":   "Oakwood Vets on Yell",
						"url":     "https://www.yell.com/oakwoodvets",
						"snippet": "Directory listing",
						"source":  "brave",
					},
				},
				"query":       "Oakwood Veterinary Surgery Leeds",
				"total":       2,
				"provider":    "brave",
				"search_type": "web",
				"fallbacks":   []interface{}{},
			},
		},
	}
}

func TestExtractNestedField_VetFallbackPathResolves(t *testing.T) {
	data := vetSearchPayload()

	// The exact string configured on the live vet-practice-verifier row
	// (default_config.workflow.steps.scrape_website.config.fallback_url_field,
	// verified against the live DB on 2026-08-09).
	const configured = "search_results.results.0.url"

	got := ExtractNestedField(data, configured)
	if got != "https://oakwoodvets.co.uk" {
		t.Fatalf("%s = %#v, want %q", configured, got, "https://oakwoodvets.co.uk")
	}

	if got := ExtractNestedField(data, "search_results.results.1.url"); got != "https://www.yell.com/oakwoodvets" {
		t.Fatalf("second element: got %#v", got)
	}
	if got := ExtractNestedField(data, "search_results.results.0.title"); got != "Oakwood Veterinary Surgery" {
		t.Fatalf("sibling field on an indexed element: got %#v", got)
	}
}

func TestExtractNestedField_ArrayIndexEdges(t *testing.T) {
	data := vetSearchPayload()

	tests := []struct {
		name string
		path string
		want interface{}
	}{
		{"index out of range", "search_results.results.2.url", nil},
		{"index far out of range", "search_results.results.99.url", nil},
		{"negative index", "search_results.results.-1.url", nil},
		{"non-numeric segment on an array", "search_results.results.url", nil},
		{"empty segment on an array", "search_results.results..url", nil},
		{"float segment on an array", "search_results.results.0_5.url", nil},
		{"missing field on an indexed element", "search_results.results.0.published_at", nil},
		{"index into an empty array", "search_results.fallbacks.0", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractNestedField(data, tt.path); !jsonEqual(got, tt.want) {
				t.Fatalf("ExtractNestedField(%q) = %#v, want %#v", tt.path, got, tt.want)
			}
		})
	}
}

// TestExtractNestedField_TrailingIndex covers a path that ENDS at an array
// element rather than descending through it, for both a scalar element and a
// map element.
func TestExtractNestedField_TrailingIndex(t *testing.T) {
	data := map[string]interface{}{
		"scalars": []interface{}{"first", "second"},
		"objects": []interface{}{map[string]interface{}{"url": "https://example.co.uk"}},
	}

	if got := ExtractNestedField(data, "scalars.1"); got != "second" {
		t.Fatalf("scalars.1 = %#v, want %q", got, "second")
	}
	want := map[string]interface{}{"url": "https://example.co.uk"}
	if got := ExtractNestedField(data, "objects.0"); !jsonEqual(got, want) {
		t.Fatalf("objects.0 = %#v, want %#v", got, want)
	}
}

// TestExtractNestedField_NestedArrays walks two array levels, which the vet
// path does not exercise but the segment rule permits.
func TestExtractNestedField_NestedArrays(t *testing.T) {
	data := map[string]interface{}{
		"pages": []interface{}{
			map[string]interface{}{
				"sections": []interface{}{
					map[string]interface{}{"heading": "Our team"},
				},
			},
		},
	}
	if got := ExtractNestedField(data, "pages.0.sections.0.heading"); got != "Our team" {
		t.Fatalf("pages.0.sections.0.heading = %#v, want %q", got, "Our team")
	}
}

// TestExtractNestedField_TypedNilSliceIsNotIndexable guards the one shape that
// looks like a slice and is not: []map[string]interface{} is a different type
// from []interface{}, which is all a JSON round-trip ever produces. Indexing
// must stay nil there rather than panicking.
func TestExtractNestedField_TypedNilSliceIsNotIndexable(t *testing.T) {
	data := map[string]interface{}{
		"results": []map[string]interface{}{{"url": "https://example.co.uk"}},
		"empty":   []interface{}(nil),
	}
	if got := ExtractNestedField(data, "results.0.url"); got != nil {
		t.Fatalf("[]map[string]interface{} must not index: got %#v, want nil", got)
	}
	if got := ExtractNestedField(data, "empty.0"); got != nil {
		t.Fatalf("nil slice must not index: got %#v, want nil", got)
	}
}

// TestExtractNestedField_FromRealJSON runs the vet path over a payload that
// has been through encoding/json, which is how it arrives at runtime: this is
// what guarantees the array is a []interface{} and not a typed slice.
func TestExtractNestedField_FromRealJSON(t *testing.T) {
	raw := `{
	  "search_results": {
	    "status": "complete",
	    "response": {
	      "results": [
	        {"title": "Oakwood Veterinary Surgery", "url": "https://oakwoodvets.co.uk", "snippet": "Leeds", "source": "brave"}
	      ],
	      "query": "Oakwood Veterinary Surgery Leeds",
	      "total": 1,
	      "provider": "brave",
	      "search_type": "web"
	    }
	  }
	}`

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := ExtractNestedField(data, "search_results.results.0.url"); got != "https://oakwoodvets.co.uk" {
		t.Fatalf("json round-trip: got %#v", got)
	}
}

// --- the string wrapper --------------------------------------------------

func TestExtractNestedFieldString_Delegation(t *testing.T) {
	data := vetSearchPayload()

	tests := []struct {
		name string
		path string
		want string
	}{
		{"resolves the vet fallback path", "search_results.results.0.url", "https://oakwoodvets.co.uk"},
		{"missing path is empty", "search_results.results.9.url", ""},
		{"empty path is empty", "", ""},
		{"non-string value is empty", "search_results.total", ""},
		{"a map value is empty", "search_results.results.0", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractNestedFieldString(data, tt.path); got != tt.want {
				t.Fatalf("ExtractNestedFieldString(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}

	if got := ExtractNestedFieldString(nil, "search_results.results.0.url"); got != "" {
		t.Fatalf("nil data: got %q, want empty", got)
	}
}

func TestExtractNestedFieldMap_OnIndexedElement(t *testing.T) {
	data := vetSearchPayload()

	m := ExtractNestedFieldMap(data, "search_results.results.0")
	if m == nil {
		t.Fatal("expected the indexed element as a map, got nil")
	}
	if m["url"] != "https://oakwoodvets.co.uk" {
		t.Fatalf("indexed element map: got %#v", m)
	}
	if got := ExtractNestedFieldMap(data, "search_results.results.0.url"); got != nil {
		t.Fatalf("a string value must not come back as a map: got %#v", got)
	}
}

// jsonEqual compares two decoded-JSON values structurally. Used rather than
// reflect.DeepEqual so that the expectations above read as the payloads they
// describe.
func jsonEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	ab, err1 := json.Marshal(a)
	bb, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(ab) == string(bb)
}
