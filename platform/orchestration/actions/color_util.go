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
	"strings"

	"github.com/gqls/agentchassis/platform/colour"
	"go.uber.org/zap"
)

// ============================================================================
// Hex parsing and WCAG luminance
// ============================================================================

// The formulas below now live in platform/colour so that
// actions/discovery_checks can use them too — `actions` imports that package,
// so it can never import back, and a second copy of the WCAG maths is exactly
// the two-things-that-must-agree drift this codebase keeps paying for
// (bugs_open/109, bugs_open/113). These wrappers keep the unexported names so
// no call site in this package had to change.

// parseHexColor handles #rgb, #rrggbb, and #rrggbbaa forms. Alpha is ignored.
func parseHexColor(hex string) (r, g, b uint8, err error) {
	return colour.ParseHex(hex)
}

// relativeLuminance returns the WCAG relative luminance of an sRGB colour (0..1).
func relativeLuminance(r, g, b uint8) float64 {
	return colour.RelativeLuminance(r, g, b)
}

// wcagContrastRatio returns the WCAG contrast ratio between two hex colours:
// 1.0 for identical colours, 21.0 for pure black on pure white.
func wcagContrastRatio(hex1, hex2 string) (float64, error) {
	return colour.ContrastRatio(hex1, hex2)
}

// isDarkHex returns true when white contrasts better than black on the given
// colour — "does this background need light-on-dark text?". Self-consistent
// with the picker by construction, because both now read one implementation.
func isDarkHex(hex string) bool {
	return colour.IsDark(hex)
}

// isDarkColor returns true if a hex colour has relative luminance < 0.2.
// NOT a synonym for isDarkHex; see platform/colour for why they differ.
func isDarkColor(hex string) bool {
	return colour.IsPerceptuallyDark(hex)
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
//     cream that reads far better than pure white on a dark hero. (On a dark
//     site this is inert: the background against a dark section scores ~1:1
//     and is skipped.)
//   - "text" next: it is the palette's reading colour, and the answer the
//     palette's author actually chose.
//   - "accent", "text_muted", "secondary" as palette-character fallbacks.
//   - "primary" last because it usually shares a colour family with dark
//     backgrounds, so contrast is often poor, but worth checking.
//
// CORRECTED 2026-07-27 — "text_muted" used to sit AHEAD of "text", and that
// one line set every dark site's default body AND heading colour to the
// palette's DIMMEST reading colour. On fundamentallyai.com the emitted block
// was `--section-text: #7E91A8; --section-heading: #7E91A8` while #E4EAF2 sat
// unused in the same palette: headings rendered in the muted grey meant for
// de-emphasised captions. text_muted is by definition the de-emphasised
// colour; using it as the default de-emphasises everything.
var paletteTextPreference = []string{
	"background",
	"text",
	"accent",
	"text_muted",
	"secondary",
	"primary",
}

const (
	// Minimum contrast ratio for body text on section backgrounds: WCAG AA.
	//
	// RAISED 2026-07-27 from 3.0. The old value was set "deliberately loose to
	// preserve palette character", and the character it preserved was body copy
	// below the readability floor — a site can ship text at 3.0:1 and no check
	// downstream disagrees, because every check runs on the source rather than
	// the rendered page. With "text" now ahead of "text_muted" in the
	// preference order this rarely binds: a palette's own reading colour
	// clears 4.5 comfortably on its own background.
	sectionBodyMinContrast = 4.5

	// Minimum contrast ratio for section headings: WCAG AA for large text.
	// Headings are large and bold, so 3.0 is the standard's own allowance —
	// the previous 2.0 was below any published floor.
	sectionHeadingMinContrast = 3.0
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

		// A palette that names a `heading` slot has already answered this
		// question; the walk is only for palettes that don't. Added
		// 2026-07-27 — `heading` is defined by 15 of 31 palettes and was
		// consulted by nothing, so a curated heading colour lost to whatever
		// the preference walk happened to reach first.
		headingHex, headingSrc := "", ""
		if h, ok := palette["heading"]; ok && h != "" {
			if r, err := wcagContrastRatio(h, bg); err == nil && r >= sectionHeadingMinContrast {
				headingHex, headingSrc = h, "palette:heading"
			}
		}
		if headingHex == "" {
			headingHex, headingSrc = pickReadableOnBackground(bg, palette, sectionHeadingMinContrast)
		}
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
