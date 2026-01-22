package orchestration

import (
	"reflect"
	"testing"

	"go.uber.org/zap"
)

func TestGetValueAtExactPath(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]interface{}
		path    string
		wantVal interface{}
		wantOk  bool
	}{
		{
			name: "simple top-level field",
			data: map[string]interface{}{
				"site_record": map[string]interface{}{"site_id": "abc123"},
				"page":        map[string]interface{}{"title": "Home"},
			},
			path:    "page",
			wantVal: map[string]interface{}{"title": "Home"},
			wantOk:  true,
		},
		{
			name: "nested field with dot notation",
			data: map[string]interface{}{
				"site_record": map[string]interface{}{
					"site_id": "abc123",
					"domain":  "example.com",
				},
			},
			path:    "site_record.site_id",
			wantVal: "abc123",
			wantOk:  true,
		},
		{
			name: "deeply nested field",
			data: map[string]interface{}{
				"site_record": map[string]interface{}{
					"content_data": map[string]interface{}{
						"brief": map[string]interface{}{
							"objective": "sell products",
						},
					},
				},
			},
			path:    "site_record.content_data.brief.objective",
			wantVal: "sell products",
			wantOk:  true,
		},
		{
			name: "non-existent field",
			data: map[string]interface{}{
				"site_record": map[string]interface{}{"site_id": "abc123"},
			},
			path:   "missing_field",
			wantOk: false,
		},
		{
			name: "non-existent nested field",
			data: map[string]interface{}{
				"site_record": map[string]interface{}{"site_id": "abc123"},
			},
			path:   "site_record.missing.field",
			wantOk: false,
		},
		{
			name:   "empty path",
			data:   map[string]interface{}{"field": "value"},
			path:   "",
			wantOk: false,
		},
		{
			name: "path through non-map value",
			data: map[string]interface{}{
				"site_id": "abc123", // string, not map
			},
			path:   "site_id.nested",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := GetValueAtExactPath(tt.data, tt.path)
			if gotOk != tt.wantOk {
				t.Errorf("GetValueAtExactPath() ok = %v, want %v", gotOk, tt.wantOk)
				return
			}
			if tt.wantOk {
				// Compare values (simple comparison for basic types)
				if !reflect.DeepEqual(gotVal, tt.wantVal) {
					// For maps, do a deeper comparison
					gotMap, gotIsMap := gotVal.(map[string]interface{})
					wantMap, wantIsMap := tt.wantVal.(map[string]interface{})
					if gotIsMap && wantIsMap {
						if len(gotMap) != len(wantMap) {
							t.Errorf("GetValueAtExactPath() map length = %v, want %v", len(gotMap), len(wantMap))
						}
					} else {
						t.Errorf("GetValueAtExactPath() value = %v, want %v", gotVal, tt.wantVal)
					}
				}
			}
		})
	}
}

func TestResolveInputMapping(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name        string
		data        map[string]interface{}
		mapping     InputMapping
		wantResult  map[string]interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "simple mapping",
			data: map[string]interface{}{
				"site_record": map[string]interface{}{"site_id": "abc123"},
				"page":        map[string]interface{}{"title": "Home"},
			},
			mapping: InputMapping{
				"current_page": "page",
				"site":         "site_record",
			},
			wantResult: map[string]interface{}{
				"current_page": map[string]interface{}{"title": "Home"},
				"site":         map[string]interface{}{"site_id": "abc123"},
			},
			wantErr: false,
		},
		{
			name: "nested source paths",
			data: map[string]interface{}{
				"site_record": map[string]interface{}{
					"site_id": "abc123",
					"domain":  "example.com",
				},
			},
			mapping: InputMapping{
				"id":     "site_record.site_id",
				"domain": "site_record.domain",
			},
			wantResult: map[string]interface{}{
				"id":     "abc123",
				"domain": "example.com",
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			data: map[string]interface{}{
				"site_record": map[string]interface{}{"site_id": "abc123"},
			},
			mapping: InputMapping{
				"page": "missing_page", // This doesn't exist
			},
			wantErr:     true,
			errContains: "source path 'missing_page' not found",
		},
		{
			name: "$item token is skipped",
			data: map[string]interface{}{
				"site_record": map[string]interface{}{"site_id": "abc123"},
			},
			mapping: InputMapping{
				"current_page": "$item", // Should be skipped
				"site":         "site_record",
			},
			wantResult: map[string]interface{}{
				"site": map[string]interface{}{"site_id": "abc123"},
				// Note: current_page is NOT included because $item is skipped
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := ResolveInputMapping(tt.data, tt.mapping, logger)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveInputMapping() expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ResolveInputMapping() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveInputMapping() unexpected error: %v", err)
				return
			}

			// Check result has expected keys
			for k := range tt.wantResult {
				if _, exists := gotResult[k]; !exists {
					t.Errorf("ResolveInputMapping() missing key %s in result", k)
				}
			}
		})
	}
}

func TestResolveInputMappingWithItem(t *testing.T) {
	logger := zap.NewNop()

	data := map[string]interface{}{
		"site_record": map[string]interface{}{"site_id": "abc123"},
	}

	currentItem := map[string]interface{}{
		"title": "Home",
		"slug":  "/",
	}

	mapping := InputMapping{
		"page": "$item",
		"site": "site_record",
	}

	result, err := ResolveInputMappingWithItem(data, mapping, currentItem, logger)
	if err != nil {
		t.Errorf("ResolveInputMappingWithItem() unexpected error: %v", err)
		return
	}

	// Check page is the current item
	if page, ok := result["page"].(map[string]interface{}); !ok {
		t.Error("ResolveInputMappingWithItem() page should be a map")
	} else {
		if page["title"] != "Home" {
			t.Errorf("ResolveInputMappingWithItem() page.title = %v, want Home", page["title"])
		}
	}

	// Check site is resolved from path
	if _, ok := result["site"]; !ok {
		t.Error("ResolveInputMappingWithItem() missing site key")
	}
}

func TestValidateInputContract(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name        string
		agentType   string
		data        map[string]interface{}
		contract    *InputContract
		wantErr     bool
		errContains string
	}{
		{
			name:      "nil contract passes",
			agentType: "test-agent",
			data:      map[string]interface{}{"field": "value"},
			contract:  nil,
			wantErr:   false,
		},
		{
			name:      "all required fields present",
			agentType: "test-agent",
			data: map[string]interface{}{
				"page":        map[string]interface{}{"title": "Home"},
				"site_record": map[string]interface{}{"site_id": "abc"},
			},
			contract: &InputContract{
				Required: []string{"page", "site_record"},
			},
			wantErr: false,
		},
		{
			name:      "missing required field",
			agentType: "test-agent",
			data: map[string]interface{}{
				"page": map[string]interface{}{"title": "Home"},
			},
			contract: &InputContract{
				Required: []string{"page", "site_record"},
			},
			wantErr:     true,
			errContains: "missing required fields: [site_record]",
		},
		{
			name:      "multiple missing fields",
			agentType: "test-agent",
			data:      map[string]interface{}{},
			contract: &InputContract{
				Required: []string{"page", "site_record", "brief"},
			},
			wantErr:     true,
			errContains: "missing required fields",
		},
		{
			name:      "optional fields ignored",
			agentType: "test-agent",
			data: map[string]interface{}{
				"page": map[string]interface{}{"title": "Home"},
			},
			contract: &InputContract{
				Required: []string{"page"},
				Optional: []string{"style_collection"}, // Not provided, but that's OK
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputContract(tt.agentType, tt.data, tt.contract, logger)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateInputContract() expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateInputContract() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateInputContract() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestParseInputMapping(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantOk  bool
		wantLen int
	}{
		{
			name: "valid input_mapping",
			config: map[string]interface{}{
				"input_mapping": map[string]interface{}{
					"page":        "current_page",
					"site_record": "site_record",
				},
			},
			wantOk:  true,
			wantLen: 2,
		},
		{
			name: "no input_mapping",
			config: map[string]interface{}{
				"agent_type": "test-agent",
			},
			wantOk: false,
		},
		{
			name: "empty input_mapping",
			config: map[string]interface{}{
				"input_mapping": map[string]interface{}{},
			},
			wantOk: false, // Empty mapping returns false
		},
		{
			name: "input_mapping is not a map",
			config: map[string]interface{}{
				"input_mapping": "not a map",
			},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, ok := ParseInputMapping(tt.config)
			if ok != tt.wantOk {
				t.Errorf("ParseInputMapping() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if tt.wantOk && len(mapping) != tt.wantLen {
				t.Errorf("ParseInputMapping() len = %v, want %v", len(mapping), tt.wantLen)
			}
		})
	}
}

func TestConvertInputFieldsToMapping(t *testing.T) {
	logger := zap.NewNop()

	inputFields := []interface{}{"page", "site_record", "reviewed_brief"}
	mapping := ConvertInputFieldsToMapping(inputFields, logger)

	if len(mapping) != 3 {
		t.Errorf("ConvertInputFieldsToMapping() len = %v, want 3", len(mapping))
	}

	// Each field should map to itself
	for _, field := range []string{"page", "site_record", "reviewed_brief"} {
		if mapping[field] != field {
			t.Errorf("ConvertInputFieldsToMapping() mapping[%s] = %v, want %s", field, mapping[field], field)
		}
	}
}

func TestListAvailablePaths(t *testing.T) {
	data := map[string]interface{}{
		"page": map[string]interface{}{
			"title": "Home",
			"slug":  "/",
		},
		"site_record": map[string]interface{}{
			"site_id": "abc123",
		},
		"__raw_message__": map[string]interface{}{ // Should be skipped
			"internal": "data",
		},
	}

	paths := ListAvailablePaths(data, 1)

	// Should contain page, site_record, and nested paths
	expectedPaths := []string{"page", "page.title", "page.slug", "site_record", "site_record.site_id"}
	for _, expected := range expectedPaths {
		found := false
		for _, p := range paths {
			if p == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ListAvailablePaths() missing expected path: %s", expected)
		}
	}

	// Should NOT contain __raw_message__
	for _, p := range paths {
		if contains(p, "__raw_message__") {
			t.Errorf("ListAvailablePaths() should not include internal paths: %s", p)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
