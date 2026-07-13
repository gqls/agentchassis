package actions

// FILE: platform/orchestration/actions/render_css_composition_helpers.go
//
// Pure helper functions for the composable-theme renderer introduced by
// 025_palette_layout_typography_migration (Phase 4). No DB, no logger,
// no side effects — deterministic functions over maps and JSONB bytes.
//
// Phase 4.1 of the migration: these helpers land without being wired
// into RenderCSSFromSpecAction. Phase 4.3 does the behavioural cutover.
// Landing them first lets us get unit-test coverage on the merge rules
// before any caller depends on them.
//
// Merge rules (per plan section 5):
//   - Palette core slots  (primary, secondary, accent, background,
//                         surface, text, text_muted, border):
//                         spec wins when the spec has a non-empty value
//   - Palette specialised slots (primary_hover, heading, hero_title,
//                         cta_bg, footer_text, etc.):
//                         theme always wins
//   - Typography: spec wins across the board (site font identity takes
//                 precedence over theme library's default)
//   - Structure tokens: layout always wins (spec does not contribute)

import (
	"encoding/json"
)

// corePaletteKeys names the palette slots where the design_spec's
// site-identity value wins over the theme's curated default.
//
// If a site's design_spec supplies `primary`, that primary applies
// regardless of which theme the site has selected. Specialised slots
// (everything not in this set) are owned by the theme — a theme like
// `bakery` has curated its `cta_bg`, `card_bg`, `heading` etc. to
// work together, and a site's partial spec shouldn't tear that apart.
//
// The set is deliberately small and stable. Adding a new core slot
// here changes the site/theme authority boundary, which is a plan-
// level decision, not a routine code change.
var corePaletteKeys = map[string]struct{}{
	"primary":    {},
	"secondary":  {},
	"accent":     {},
	"background": {},
	"surface":    {},
	"text":       {},
	"text_muted": {},
	"border":     {},
}

// isCorePaletteKey reports whether a palette slot is site-identity
// (spec wins) versus theme-owned (theme wins).
func isCorePaletteKey(key string) bool {
	_, ok := corePaletteKeys[key]
	return ok
}

// buildPaletteMap merges a theme's palette (from palettes.colours)
// with a design_spec's color_scheme, applying the core-vs-specialised
// authority rule.
//
// Inputs:
//   - themePalette: the theme's curated palette map; keys and values
//     both strings. Usually non-nil and populated (Phase 3 seeds all
//     14 themes' palettes), but the function tolerates nil.
//   - specPalette:  the raw color_scheme map from design_spec,
//     values typed as interface{} because they arrive from JSON
//     parsing. Non-string values are skipped.
//
// Returns a fresh map; inputs are not modified.
func buildPaletteMap(themePalette map[string]string, specPalette map[string]interface{}) map[string]string {
	out := make(map[string]string, len(themePalette)+len(specPalette))

	// 1. Start with every non-empty slot from the theme's palette.
	for k, v := range themePalette {
		if v != "" {
			out[k] = v
		}
	}

	// 2. Overlay the spec's core slots where the spec has a value.
	for k, raw := range specPalette {
		if !isCorePaletteKey(k) {
			continue
		}
		if v, ok := raw.(string); ok && v != "" {
			out[k] = v
		}
	}

	return out
}

// buildTypographyMap merges a typography_set's fonts with a design_spec's
// typography block. Spec wins across the board (not just core slots) —
// typography is a single-identity thing; a site that specifies a font
// family clearly wants it used.
//
// Both inputs may be nil or partial; returns a fresh map.
func buildTypographyMap(typeSetFonts map[string]string, specTypo map[string]interface{}) map[string]string {
	out := make(map[string]string, len(typeSetFonts)+len(specTypo))

	for k, v := range typeSetFonts {
		if v != "" {
			out[k] = v
		}
	}

	for k, raw := range specTypo {
		if v, ok := raw.(string); ok && v != "" {
			out[k] = v
		}
	}

	return out
}

// buildStructureMap extracts structure tokens (container widths, section
// paddings, radii, shadow tokens) from a layout's structure_tokens JSONB.
// Structure is layout-owned — design_spec does not contribute, so this
// function takes no spec input.
//
// Returns an empty (non-nil) map on empty input so callers can treat
// the result as always non-nil.
func buildStructureMap(rawJSON []byte) (map[string]string, error) {
	return jsonbToStringMap(rawJSON)
}

// jsonbToStringMap is the shared JSONB-to-flat-string-map converter
// used by buildStructureMap and the palette/typography loaders. Keys
// whose values aren't strings are silently skipped — the palette and
// structure_tokens schemas in Phase 3 seed only string values, so a
// non-string value indicates an unexpected row and dropping it is
// safer than failing the render.
//
// Returns an empty (non-nil) map on empty/nil input.
func jsonbToStringMap(rawJSON []byte) (map[string]string, error) {
	out := make(map[string]string)
	if len(rawJSON) == 0 {
		return out, nil
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(rawJSON, &parsed); err != nil {
		return out, err
	}

	for k, v := range parsed {
		if s, ok := v.(string); ok && s != "" {
			out[k] = s
		}
	}

	return out, nil
}

// makeMapLookupFunc returns the closure used as a template helper
// (palette / typo / token). Signature matches what text/template's
// FuncMap expects: a function of string arguments returning a string.
//
// Lookup semantics:
//   - key present and value non-empty → return the stored value
//   - key missing OR value empty      → return the fallback
//
// The second form lets layout templates remain valid CSS even when a
// palette or typography set omits a slot. Every {{palette "key"
// "fallback"}} call in the Phase 1 layout templates supplies an
// explicit fallback as required by the Template Helper Fallback
// contract (003_contracts_and_standards_v7.md).
func makeMapLookupFunc(m map[string]string) func(key, fallback string) string {
	return func(key, fallback string) string {
		if v, ok := m[key]; ok && v != "" {
			return v
		}
		return fallback
	}
}
