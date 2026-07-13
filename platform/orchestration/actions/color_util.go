// FILE: platform/orchestration/actions/color_util.go
//
// Shared colour utilities used by the CSS renderer and any action that
// reasons about hex colours (contrast checks, dark/light classification,
// palette-aware text picking).
//
// All functions operate on CSS hex strings (#rgb, #rrggbb, #rrggbbaa) and
// follow the WCAG relative-luminance + contrast-ratio formulas.

package actions

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// ============================================================================
// Hex parsing and WCAG luminance
// ============================================================================

// parseHexColor handles #rgb, #rrggbb, and #rrggbbaa forms. Alpha is ignored.
func parseHexColor(hex string) (r, g, b uint8, err error) {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	switch len(hex) {
	case 3:
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	case 6:
		// already fine
	case 8:
		hex = hex[:6]
	default:
		return 0, 0, 0, fmt.Errorf("invalid hex color: #%s", hex)
	}

	rr, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}
	gg, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}
	bb, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}
	return uint8(rr), uint8(gg), uint8(bb), nil
}

// sRGBToLinear converts an sRGB channel value (0-255) to linear light.
func sRGBToLinear(c uint8) float64 {
	f := float64(c) / 255.0
	if f <= 0.03928 {
		return f / 12.92
	}
	return math.Pow((f+0.055)/1.055, 2.4)
}

// relativeLuminance returns the WCAG relative luminance of an sRGB colour (0..1).
func relativeLuminance(r, g, b uint8) float64 {
	return 0.2126*sRGBToLinear(r) + 0.7152*sRGBToLinear(g) + 0.0722*sRGBToLinear(b)
}

// wcagContrastRatio returns the WCAG contrast ratio between two hex colours:
// 1.0 for identical colours, 21.0 for pure black on pure white.
func wcagContrastRatio(hex1, hex2 string) (float64, error) {
	r1, g1, b1, err := parseHexColor(hex1)
	if err != nil {
		return 0, fmt.Errorf("parse fg color %q: %w", hex1, err)
	}
	r2, g2, b2, err := parseHexColor(hex2)
	if err != nil {
		return 0, fmt.Errorf("parse bg color %q: %w", hex2, err)
	}
	l1 := relativeLuminance(r1, g1, b1)
	l2 := relativeLuminance(r2, g2, b2)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05), nil
}

// ============================================================================
// Dark classification — self-consistent with the picker
// ============================================================================

// isDarkHex returns true when white contrasts better than black on the given
// colour. This is the semantically correct check for "does this background
// need light-on-dark text overrides?" and is self-consistent with the
// contrast-based picker. The crossover point sits around #777777, matching
// intuition.
func isDarkHex(hex string) bool {
	whiteRatio, err := wcagContrastRatio("#ffffff", hex)
	if err != nil {
		return false
	}
	blackRatio, _ := wcagContrastRatio("#000000", hex)
	return whiteRatio > blackRatio
}

// isDarkColor returns true if a hex color has relative luminance < 0.2
// (perceptually dark).
func isDarkColor(hex string) bool {
	r, g, b, err := parseHexColor(hex)
	if err != nil {
		return false
	}
	return relativeLuminance(r, g, b) < 0.2
}

// ============================================================================
// Hex → rgba conversion
// ============================================================================

// hexToRGBA converts a hex colour to an rgba() CSS string with the given alpha.
// Returns the original hex on parse failure so the caller can recover.
func hexToRGBA(hex string, alpha float64) string {
	r, g, b, err := parseHexColor(hex)
	if err != nil {
		return hex
	}
	return fmt.Sprintf("rgba(%d,%d,%d,%.2f)", r, g, b, alpha)
}

// ============================================================================
// Palette-aware text colour picker
// ============================================================================

// paletteTextPreference is the order in which palette colours are tried when
// picking a text/heading colour for a section background. The first colour
// whose contrast against the background meets minRatio is returned.
//
// Chosen to preserve palette character:
//   - "background" first because on light themes it's a warm off-white or
//     cream that reads far better than pure white on a dark hero.
//   - "text_muted", "text" next because they are the theme's reading colours
//     — on dark themes they are already light-on-dark.
//   - "accent", "secondary" as palette-character fallbacks.
//   - "primary" last because it usually shares a colour family with dark
//     backgrounds, so contrast is often poor, but worth checking.
var paletteTextPreference = []string{
	"background",
	"text_muted",
	"text",
	"accent",
	"secondary",
	"primary",
}

const (
	// Minimum contrast ratio for body text on section backgrounds. Below WCAG AA
	// body (4.5) but above WCAG AA large-text (3.0). Deliberately loose to preserve
	// palette character — picks softer palette colours over clinical pure-white fallback.
	sectionBodyMinContrast = 3.0

	// Minimum contrast ratio for section headings. Headings are large and bold so
	// readability holds at lower ratios; this lets accent/warm palette colours come
	// through on decorative titles.
	sectionHeadingMinContrast = 2.0
)

// pickReadableOnBackground returns a text colour (hex) suitable for rendering
// on bgHex. It walks paletteTextPreference and returns the first colour whose
// contrast ratio is >= minRatio. If no palette colour qualifies, it falls
// back to #ffffff or #000000 (whichever has higher contrast against bgHex).
//
// The second return value is a short label identifying which source was
// chosen, for logging.
func pickReadableOnBackground(
	bgHex string,
	palette map[string]string,
	minRatio float64,
) (hex, source string) {

	for _, key := range paletteTextPreference {
		val, ok := palette[key]
		if !ok || val == "" {
			continue
		}
		ratio, err := wcagContrastRatio(val, bgHex)
		if err != nil {
			continue
		}
		if ratio >= minRatio {
			return val, "palette:" + key
		}
	}

	// No palette colour qualifies — pick black or white based on which has
	// more contrast against the background.
	whiteRatio, _ := wcagContrastRatio("#ffffff", bgHex)
	blackRatio, _ := wcagContrastRatio("#000000", bgHex)
	if whiteRatio >= blackRatio {
		return "#ffffff", "fallback:white"
	}
	return "#000000", "fallback:black"
}

// buildSectionDefaults emits a renderer-owned block of --section-* CSS variable
// defaults. Text and heading colours are picked from the site's palette using
// pickReadableOnBackground so the output preserves the theme's character rather
// than falling back to clinical pure-white.
//
// When bgIsDark is true, a body-level block is emitted so any section without
// its own overrides inherits readable text colours via the
// var(--section-*, var(--color-*)) fallback pattern used in base theme rules.
//
// When surfaceIsDark is true, the same overrides are applied to the set of
// sections that commonly get painted with --color-surface across themes.
// Themes that don't use these section names are unaffected — the rules are inert.
//
// Themes MUST NOT declare --section-* defaults themselves; the renderer owns
// this contract. See 003_contracts_and_standards.md → CSS Theme Template Contract.
func buildSectionDefaults(
	bgHex, surfaceHex string,
	palette map[string]string,
	bgIsDark, surfaceIsDark bool,
	logger *zap.Logger,
) string {
	if !bgIsDark && !surfaceIsDark {
		return ""
	}

	// Shared builder: given the section's background hex, pick palette-aware
	// text/heading colours and emit the full --section-* override block.
	emitOverrides := func(bg string, context string) string {
		textHex, textSrc := pickReadableOnBackground(bg, palette, sectionBodyMinContrast)
		headingHex, headingSrc := pickReadableOnBackground(bg, palette, sectionHeadingMinContrast)
		mutedRGBA := hexToRGBA(textHex, 0.75)

		logger.Info("buildSectionDefaults: picked section colours",
			zap.String("context", context),
			zap.String("background", bg),
			zap.String("text", textHex),
			zap.String("text_source", textSrc),
			zap.String("heading", headingHex),
			zap.String("heading_source", headingSrc))

		return fmt.Sprintf(
			"  --section-text: %s;\n"+
				"  --section-text-muted: %s;\n"+
				"  --section-heading: %s;\n"+
				"  --section-surface: rgba(255,255,255,0.05);\n"+
				"  --section-border: rgba(255,255,255,0.2);",
			textHex, mutedRGBA, headingHex)
	}

	var sb strings.Builder
	sb.WriteString("\n\n/* ── Renderer-enforced section defaults ── */\n")
	sb.WriteString("/* Text/heading colours are picked from the site palette by contrast ratio. */\n")
	sb.WriteString("/* Themes MUST NOT declare --section-* defaults; the renderer owns this. */\n\n")

	if bgIsDark {
		sb.WriteString("body {\n")
		sb.WriteString(emitOverrides(bgHex, "body"))
		sb.WriteString("\n}\n\n")
	}

	if surfaceIsDark {
		sb.WriteString(".differentiators-section,\n")
		sb.WriteString(".features-section,\n")
		sb.WriteString(".faq-section,\n")
		sb.WriteString(".services-section,\n")
		sb.WriteString(".about-section {\n")
		sb.WriteString(emitOverrides(surfaceHex, "surface"))
		sb.WriteString("\n}\n")
	}

	return sb.String()
}
