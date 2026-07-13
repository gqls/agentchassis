// FILE: platform/orchestration/datahelpers/format_content_direction.go
//
// FormatContentDirection walks a content_direction spec (whatever structure
// the LLM produced) and formats it into a single readable text block.
// Stored as "formatted" in the spec. The content writer reads one field:
//   {{.site_specs.specs.content_direction.formatted}}
//
// This is a formatting utility, not an action. Called by:
//   - write_site_spec (when aspect == "content_direction")
//   - apply_adoption_plan (before writing the spec)
//   - any future agent that writes content_direction

package datahelpers

import (
	"strings"
)

// FormatContentDirection recursively formats the entire content_direction
// spec into a single readable text block. Handles any structure — strings,
// arrays, nested maps. Unknown fields are included automatically.
func FormatContentDirection(spec map[string]interface{}) string {
	var sections []string

	for key, val := range spec {
		if val == nil {
			continue
		}
		// Skip the formatted field itself
		if key == "formatted" {
			continue
		}
		formatted := FormatSpecValue(HumaniseKey(key), val)
		if formatted != "" {
			sections = append(sections, formatted)
		}
	}

	if len(sections) == 0 {
		return ""
	}
	return strings.Join(sections, "\n\n")
}

// FormatSpecValue formats a single value into readable text.
// Handles strings, arrays, and nested maps recursively.
func FormatSpecValue(label string, val interface{}) string {
	switch v := val.(type) {
	case string:
		if v == "" {
			return ""
		}
		return label + ": " + v

	case []interface{}:
		if len(v) == 0 {
			return ""
		}
		strs := InterfaceSliceToStrings(v)
		if len(strs) == 0 {
			return ""
		}
		lines := []string{label + ":"}
		for _, s := range strs {
			lines = append(lines, "- "+s)
		}
		return strings.Join(lines, "\n")

	case map[string]interface{}:
		if len(v) == 0 {
			return ""
		}
		var parts []string
		for subKey, subVal := range v {
			formatted := FormatSpecValue(HumaniseKey(subKey), subVal)
			if formatted != "" {
				parts = append(parts, formatted)
			}
		}
		if len(parts) == 0 {
			return ""
		}
		return label + ":\n" + strings.Join(parts, "\n")

	default:
		return ""
	}
}

// HumaniseKey converts snake_case to Title case labels
func HumaniseKey(key string) string {
	key = strings.ReplaceAll(key, "_", " ")
	if len(key) > 0 {
		return strings.ToUpper(key[:1]) + key[1:]
	}
	return key
}

// InterfaceSliceToStrings converts []interface{} to []string, skipping non-strings
func InterfaceSliceToStrings(slice []interface{}) []string {
	var out []string
	for _, v := range slice {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}
