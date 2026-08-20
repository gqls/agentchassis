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
	"sort"
	"strings"
)

// FormatContentDirection recursively formats the entire content_direction
// spec into a single readable text block. Handles any structure — strings,
// arrays, nested maps. Unknown fields are included automatically.
//
// ⚠ KEYS ARE SORTED, AND THAT IS LOAD-BEARING RATHER THAN TIDY. `range` over a Go
// map is randomised by design, so this function used to render an identical spec into
// a different text on every call — same content, no shared line order. Two costs, both
// measured 2026-08-19: a diff of two briefs reported ~100% changed whether or not
// anything had changed (which is how anyone verifies that a brief correction landed),
// and a diagnosis run read three such renderings of ONE unchanged `loanzy.uk` document
// as three different partial briefs and cited it as evidence of a different bug. If you
// need a semantic ordering rather than an alphabetical one, that is a deliberate change
// with an argument attached — but it must still be DETERMINISTIC.
func FormatContentDirection(spec map[string]interface{}) string {
	var sections []string

	for _, key := range sortedKeys(spec) {
		val := spec[key]
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
		// Sorted for the same reason as the top level — a nested map is just as
		// randomly ordered, and one unsorted level makes the whole output unstable.
		for _, subKey := range sortedKeys(v) {
			formatted := FormatSpecValue(HumaniseKey(subKey), v[subKey])
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

// sortedKeys returns a map's keys in a stable order, so the rendered brief is a
// function of the spec's CONTENT and nothing else.
func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
