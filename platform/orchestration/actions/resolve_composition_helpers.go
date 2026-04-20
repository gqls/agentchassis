// FILE: platform/orchestration/actions/resolve_composition_helpers.go
//
// Shared helpers used by site-design-planner's resolver actions:
//   - validate_composition_inputs_action.go
//   - resolve_composition_layout_action.go
//   - resolve_composition_typography_action.go
//   - resolve_composition_palette_action.go
//   - install_site_composition_action.go (future)
//
// These are pure utility — spec-aspect loading, classification tag extraction,
// signal-cascade walking, string conversion. No workflow logic.
//
// Designed to be called from any composition-resolver action. None of these
// functions write to the DB; they all read or transform.

package actions

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// loadSpecAspectFromContext loads a single spec aspect's data. Tries
// collected_data first (if the workflow already loaded it via
// read_site_spec or an earlier validate step), then falls back to a
// direct DB read via loadCurrentSpecData.
//
// Candidate paths checked in collected_data (in order):
//   - validated_inputs.<aspect>       (from validate_composition_inputs)
//   - site_specs.specs.<aspect>       (from read_site_spec all-aspects mode)
//   - <aspect>_spec.data              (from read_site_spec single-aspect mode)
//   - <aspect>                         (bare aspect name)
//
// Returns (data, found). found=false with nil data means no current spec row
// exists for this site+aspect.
func loadSpecAspectFromContext(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	aspect string,
	logger *zap.Logger,
) (map[string]interface{}, bool) {

	candidates := []string{
		"validated_inputs." + aspect,
		"site_specs.specs." + aspect,
		aspect + "_spec.data",
		aspect,
	}
	for _, p := range candidates {
		raw := datahelpers.ExtractNestedField(params.CollectedData, p)
		if raw == nil {
			continue
		}
		unwrapped := datahelpers.UnwrapDeep(raw, logger)
		if m, ok := unwrapped.(map[string]interface{}); ok && len(m) > 0 {
			return m, true
		}
	}

	// Fall back to DB. loadCurrentSpecData lives in
	// validate_composition_inputs_action.go (package-scope).
	data, found, err := loadCurrentSpecData(ctx, params.DB, siteID, aspect)
	if err != nil {
		logger.Warn("loadSpecAspectFromContext: DB read failed",
			zap.String("aspect", aspect),
			zap.Error(err),
		)
		return nil, false
	}
	return data, found
}

// readClassificationFromContext pulls `category` and `industry_tags` from
// the classification spec, preferring collected_data over a DB read.
//
// Used by every composition resolver that needs classification metadata
// (layout matching, industry tags on new palette/typography rows, etc.).
//
// If classification is absent, returns ("", nil). Callers should have
// been blocked earlier by validate_composition_inputs — this is a safety
// net, not an expected path.
func readClassificationFromContext(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	logger *zap.Logger,
) (category string, industryTags []string) {

	data, found := loadSpecAspectFromContext(ctx, params, siteID, "classification", logger)
	if !found {
		return "", nil
	}

	category, _ = data["category"].(string)

	if raw, ok := data["industry_tags"]; ok {
		switch v := raw.(type) {
		case []interface{}:
			for _, t := range v {
				if s, ok := t.(string); ok {
					industryTags = append(industryTags, s)
				}
			}
		case []string:
			industryTags = append(industryTags, v...)
		}
	}
	return category, industryTags
}

// extractReferenceValuesFromSpec pulls a reference_values sub-map out of
// a design spec that follows the documented schema:
//
//	{
//	  "<section>": {
//	    "character":        "...",
//	    "reference_values": { "font_family": "...", "heading_font": "..." },
//	    "guidance":         "..."
//	  }
//	}
//
// Where section is e.g. "typography" or "palette".
//
// Falls through to the flat shape if reference_values is absent — so
// simpler spec structures still work. Non-string values (arrays, nested
// maps, numbers) are filtered out; only string values survive.
//
// Returns an empty (non-nil) map if nothing is found.
func extractReferenceValuesFromSpec(spec map[string]interface{}, section string) map[string]string {
	out := make(map[string]string)
	if spec == nil {
		return out
	}

	sectionRaw, ok := spec[section]
	if !ok {
		return out
	}
	sectionMap, ok := sectionRaw.(map[string]interface{})
	if !ok {
		return out
	}

	// Prefer nested reference_values
	if refRaw, has := sectionMap["reference_values"]; has {
		if refMap, ok := refRaw.(map[string]interface{}); ok {
			return mapInterfaceToStrings(refMap)
		}
	}

	// Fall through: flat shape directly on the section map.
	// Filter out known-non-value keys (character, guidance, etc.) so we
	// don't accidentally treat descriptive prose as a reference value.
	reserved := map[string]struct{}{
		"character":        {},
		"guidance":         {},
		"notes":            {},
		"rationale":        {},
		"reference_values": {},
	}
	for k, v := range sectionMap {
		if _, skip := reserved[k]; skip {
			continue
		}
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				out[k] = s
			}
		}
	}
	return out
}

// mapInterfaceToStrings converts a map[string]interface{} to
// map[string]string, keeping only string values and trimming whitespace.
// Non-string values are dropped silently.
//
// Renamed from mapInterfaceToString (which still exists privately in
// resolve_composition_typography_action.go) to avoid a symbol collision
// at package scope. The typography action can migrate to using this
// version in a cleanup pass.
func mapInterfaceToStrings(in map[string]interface{}) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				out[k] = s
			}
		}
	}
	return out
}

// slugifyForCompositionName sanitises a domain string for use inside
// palette/typography/css_theme names. Output contains only a-z, 0-9,
// hyphen. Capped at 40 chars to leave headroom for the "palette-" /
// "typography-" / "theme-" prefix and any collision suffix added by
// resolveUniqueNameInTx (which adds up to ~6 chars).
//
// Empty input or input with no usable chars returns "".
func slugifyForCompositionName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ".", "-")
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z',
			c >= '0' && c <= '9',
			c == '-':
			out = append(out, c)
		}
	}
	// Collapse repeated hyphens
	collapsed := make([]byte, 0, len(out))
	prevHyphen := false
	for _, c := range out {
		if c == '-' {
			if prevHyphen {
				continue
			}
			prevHyphen = true
		} else {
			prevHyphen = false
		}
		collapsed = append(collapsed, c)
	}
	// Trim leading/trailing hyphens
	result := strings.Trim(string(collapsed), "-")
	const maxLen = 40
	if len(result) > maxLen {
		result = result[:maxLen]
		result = strings.TrimRight(result, "-")
	}
	return result
}

// toPGTextArrayLiteral converts a Go []string into a PostgreSQL array
// literal suitable for passing as a parameter to an INSERT/UPDATE targeting
// a text[] column, when used with an explicit ::text[] cast in the SQL.
//
// Format: '{"elem1","elem2","elem3"}' with per-element backslash-escaping of
// double-quote and backslash. An empty slice produces '{}'.
//
// This avoids a dependency on github.com/lib/pq's pq.Array helper, which the
// codebase does not currently use. Works identically under lib/pq and pgx's
// database/sql wrapper.
//
// NOTE: This is for passing to Postgres, not for display. Do NOT json.Marshal
// a []string for a text[] column — jsonb and text[] are different types.
func toPGTextArrayLiteral(tags []string) string {
	if len(tags) == 0 {
		return "{}"
	}
	var b strings.Builder
	b.WriteByte('{')
	for i, t := range tags {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('"')
		// Escape backslash first, then double-quote
		for j := 0; j < len(t); j++ {
			c := t[j]
			switch c {
			case '\\', '"':
				b.WriteByte('\\')
				b.WriteByte(c)
			default:
				b.WriteByte(c)
			}
		}
		b.WriteByte('"')
	}
	b.WriteByte('}')
	return b.String()
}
