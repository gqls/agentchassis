// FILE: platform/orchestration/actions/resolve_composition_reference_helpers.go
//
// Helpers for reading a palette / typography signal out of a design_reference
// spec produced by extract_design_fingerprint (+ enrich_fingerprint_with_css).
//
// WHY THIS FILE EXISTS
// --------------------
// resolve_composition_palette_action.go and resolve_composition_typography_action.go
// both had a design_reference cascade slot that called
//   extractReferenceValuesFromSpec(ref, "palette" | "typography")
// which reads ref.palette.reference_values / ref.typography.reference_values.
//
// The fingerprint NEVER writes those keys. It writes:
//   - suggested_mapping : a flat map already keyed to our colour slots
//                         (text, primary, background, surface, text_muted,
//                          border, ...) plus font_family
//   - css_variables     : the source site's own --var names and values
//   - colors            : raw extracted colour lists (not slot-keyed)
//   - typography.fonts  : raw font families (array)
//
// So the design_reference slot could never fire. In practice the palette/typography
// for an adopted site comes from the GENERATED design_intent (which the adoption
// pipeline derives FROM this fingerprint and which carries the same colours plus
// accent/secondary the raw fingerprint lacks). That is the richer source and must
// stay ahead of the raw fingerprint — see the note in extractPaletteSignal.
//
// These helpers let the fingerprint act as a genuine FALLBACK: when the generated
// design_intent produced no palette/typography, an adopted design_reference can
// still seed the composition directly instead of dropping to the layout's default
// dark palette / sans-modern. Fresh (non-adopted) sites have no design_reference,
// so nothing here changes their behaviour.
//
// None of these functions read the DB; they only transform an already-loaded spec.

package actions

import "strings"

// paletteColourSlots is the set of canonical palette slot names the composition
// pipeline understands. Anything outside this set found in suggested_mapping
// (notably font_family) is not a colour and is dropped.
var paletteColourSlots = map[string]struct{}{
	"primary":    {},
	"secondary":  {},
	"accent":     {},
	"background": {},
	"surface":    {},
	"text":       {},
	"text_muted": {},
	"border":     {},
}

// cssVariableSynonyms maps common source CSS-variable names onto our canonical
// palette slots. Deliberately conservative: only well-understood names are
// mapped; anything unrecognised is ignored rather than guessed. Keys are matched
// after lowercasing and stripping a leading "--".
var cssVariableSynonyms = map[string]string{
	"primary":          "primary",
	"primary-color":    "primary",
	"primary-colour":   "primary",
	"color-primary":    "primary",
	"brand":            "primary",
	"brand-color":      "primary",
	"secondary":        "secondary",
	"secondary-color":  "secondary",
	"secondary-colour": "secondary",
	"accent":           "accent",
	"accent-color":     "accent",
	"accent-colour":    "accent",
	"highlight":        "accent",
	"background":       "background",
	"bg":               "background",
	"bg-color":         "background",
	"bg-colour":        "background",
	"background-color": "background",
	"surface":          "surface",
	"surface-color":    "surface",
	"card":             "surface",
	"card-bg":          "surface",
	"panel":            "surface",
	"text":             "text",
	"text-main":        "text",
	"text-color":       "text",
	"text-colour":      "text",
	"color-text":       "text",
	"foreground":       "text",
	"fg":               "text",
	"text-muted":       "text_muted",
	"text-dim":         "text_muted",
	"text-secondary":   "text_muted",
	"muted":            "text_muted",
	"border":           "border",
	"border-color":     "border",
	"border-colour":    "border",
}

// paletteFromDesignReference extracts palette colours from a design_reference
// spec produced by the fingerprint. Priority within the spec:
//
//  1. palette.reference_values  — forward-compat: if a fingerprint ever writes it
//  2. suggested_mapping         — the fingerprint's canonical slot-keyed mapping
//  3. css_variables             — source --var names, mapped via cssVariableSynonyms
//
// Only colour slots survive (font_family / heading_font are dropped — those are
// typography). Returns an empty (non-nil) map when nothing usable is found, so
// the caller's len()==0 check falls through to the next cascade slot.
func paletteFromDesignReference(ref map[string]interface{}) map[string]string {
	if ref == nil {
		return map[string]string{}
	}

	// 1. nested/flat palette.reference_values (kept for forward-compat)
	if c := dropNonColourKeys(extractReferenceValuesFromSpec(ref, "palette")); len(c) > 0 {
		return c
	}

	// 2. suggested_mapping — already keyed to our slots (text/primary/background/...)
	if raw, ok := ref["suggested_mapping"].(map[string]interface{}); ok {
		if c := dropNonColourKeys(mapInterfaceToStrings(raw)); len(c) > 0 {
			return c
		}
	}

	// 3. css_variables — map the source site's own var names onto our slots
	if raw, ok := ref["css_variables"].(map[string]interface{}); ok {
		if c := paletteFromCSSVariables(mapInterfaceToStrings(raw)); len(c) > 0 {
			return c
		}
	}

	return map[string]string{}
}

// dropNonColourKeys keeps only recognised colour-slot keys from a string map.
func dropNonColourKeys(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		if _, ok := paletteColourSlots[k]; !ok {
			continue
		}
		if v = strings.TrimSpace(v); v != "" {
			out[k] = v
		}
	}
	return out
}

// paletteFromCSSVariables maps a source site's --var names onto our palette
// slots using cssVariableSynonyms. A slot already filled by an earlier (more
// specific) variable is not overwritten. Font and unrecognised variables are
// skipped.
func paletteFromCSSVariables(vars map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range vars {
		name := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(k), "--"))
		slot, ok := cssVariableSynonyms[name]
		if !ok {
			continue
		}
		if _, exists := out[slot]; exists {
			continue
		}
		if v = strings.TrimSpace(v); v != "" {
			out[slot] = v
		}
	}
	return out
}

// typographyFromDesignReference extracts a font_family (and optional
// heading_font) from the fingerprint's real shape. Priority:
//
//  1. typography.reference_values        — forward-compat
//  2. suggested_mapping.font_family      — the fingerprint's canonical font stack
//  3. css_variables --font-family / --font — last resort
//
// Returns a map containing at least "font_family" when one is found, else an
// empty map (caller then lets resolveTypographySet apply its sans-modern default).
func typographyFromDesignReference(ref map[string]interface{}) map[string]string {
	if ref == nil {
		return map[string]string{}
	}

	// 1. typography.reference_values (forward-compat)
	if f := extractReferenceValuesFromSpec(ref, "typography"); strings.TrimSpace(f["font_family"]) != "" {
		return f
	}

	out := map[string]string{}

	// 2. suggested_mapping.font_family (and heading_font if present)
	if raw, ok := ref["suggested_mapping"].(map[string]interface{}); ok {
		sm := mapInterfaceToStrings(raw)
		if ff := strings.TrimSpace(sm["font_family"]); ff != "" {
			out["font_family"] = ff
		}
		if hf := strings.TrimSpace(sm["heading_font"]); hf != "" {
			out["heading_font"] = hf
		}
	}

	// 3. css_variables font fallback
	if out["font_family"] == "" {
		if raw, ok := ref["css_variables"].(map[string]interface{}); ok {
			for k, v := range mapInterfaceToStrings(raw) {
				name := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(k), "--"))
				if name == "font-family" || name == "font" || name == "body-font" || name == "font-body" {
					if v = strings.TrimSpace(v); v != "" {
						out["font_family"] = v
						break
					}
				}
			}
		}
	}

	return out
}
