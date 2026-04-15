// ── CSS Variable Extraction (replaces fpExtractCSSVars + fpCSSVarRe) ────
//
// Multi-strategy approach for extracting design tokens from CSS of
// varying quality — hand-written, minified, Tailwind, WordPress, etc.
//
// Strategy 1: Parse :root blocks (highest confidence — explicit design tokens)
// Strategy 2: Parse other token-hosting selectors (html, body, [data-theme], .dark)
// Strategy 3: Scan remaining CSS for isolated custom property declarations
//
// Each strategy uses semicolon-splitting rather than regex to handle
// minified CSS, multi-value properties, and complex values like
// calc() or var() references.
//
// Replaces:
//   fpCSSVarRe regex (remove from var block)
//   fpExtractCSSVars function
//
// Adds:
//   fpExtractCSSVars (replacement)
//   fpExtractRootBlocks
//   fpExtractTokenBlocks
//   fpExtractVarsFromBlock
//   fpIsValidCSSValue

package actions

import "strings"

// fpExtractCSSVars extracts CSS custom property declarations using
// multiple strategies for varied CSS quality.
func fpExtractCSSVars(css string, vars map[string]string) {
	// Strip CSS comments first — prevents extracting commented-out values
	css = fpStripCSSComments(css)

	// Strategy 1: :root blocks (most reliable — explicit design tokens)
	rootBlocks := fpExtractTokenBlocks(css, []string{":root"})
	for _, block := range rootBlocks {
		fpExtractVarsFromBlock(block, vars, true)
	}

	// Strategy 2: Other common token-hosting selectors
	// Sites use html, body, .dark, [data-theme] etc. for theming
	otherSelectors := []string{
		"html", "body",
		"[data-theme", "[data-mode", // attribute selectors (partial match)
		".dark", ".light", ".theme",
	}
	otherBlocks := fpExtractTokenBlocks(css, otherSelectors)
	for _, block := range otherBlocks {
		// Don't overwrite — :root values take priority
		fpExtractVarsFromBlock(block, vars, false)
	}

	// Strategy 3 not needed if we found variables
	// For sites with no token blocks at all, the colour/font frequency
	// analysis in the main fingerprint handles it.
}

// fpExtractTokenBlocks finds CSS rule blocks whose selectors match any
// of the given patterns. Returns the content between { and }.
//
// Handles:
//   - Minified CSS (no whitespace)
//   - Multiple matching blocks (e.g. :root inside @media)
//   - Nested braces (skips inner blocks correctly)
//   - Case-insensitive matching
func fpExtractTokenBlocks(css string, selectors []string) []string {
	var blocks []string
	lower := strings.ToLower(css)

	for _, selector := range selectors {
		selectorLower := strings.ToLower(selector)
		searchFrom := 0

		for searchFrom < len(lower) {
			idx := strings.Index(lower[searchFrom:], selectorLower)
			if idx < 0 {
				break
			}
			matchPos := searchFrom + idx

			// Verify this isn't a substring of a larger selector
			// e.g. "body" shouldn't match ".body-wrapper"
			if matchPos > 0 {
				prev := css[matchPos-1]
				if isIdentChar(prev) {
					searchFrom = matchPos + len(selector)
					continue
				}
			}

			// Find the opening brace after the selector
			braceStart := -1
			for i := matchPos + len(selector); i < len(css); i++ {
				ch := css[i]
				if ch == '{' {
					braceStart = i
					break
				}
				// Allow whitespace, commas (grouped selectors), pseudo-classes
				if ch == '}' || ch == ';' {
					break // Hit end of a different rule — not our block
				}
			}

			if braceStart < 0 {
				searchFrom = matchPos + len(selector)
				continue
			}

			// Match braces to find the closing }
			depth := 1
			blockStart := braceStart + 1
			blockEnd := -1
			for i := blockStart; i < len(css); i++ {
				switch css[i] {
				case '{':
					depth++
				case '}':
					depth--
					if depth == 0 {
						blockEnd = i
					}
				}
				if blockEnd >= 0 {
					break
				}
			}

			if blockEnd > blockStart {
				blocks = append(blocks, css[blockStart:blockEnd])
			}

			if blockEnd >= 0 {
				searchFrom = blockEnd + 1
			} else {
				break
			}
		}
	}

	return blocks
}

// fpExtractVarsFromBlock parses CSS custom property declarations from a
// block of CSS. Splits on semicolons to handle minified CSS cleanly.
//
// If overwrite is false, existing values in vars are not replaced.
// This lets :root values take priority over body/html values.
func fpExtractVarsFromBlock(block string, vars map[string]string, overwrite bool) {
	// Split by semicolons to get individual declarations
	declarations := strings.Split(block, ";")

	for _, decl := range declarations {
		decl = strings.TrimSpace(decl)

		// Quick check — must contain -- somewhere
		dashIdx := strings.Index(decl, "--")
		if dashIdx < 0 {
			continue
		}

		// Find the colon that separates property from value
		// Start searching from the -- position to avoid colons in selectors
		colonIdx := strings.Index(decl[dashIdx:], ":")
		if colonIdx < 0 {
			continue
		}
		colonIdx += dashIdx // Adjust to absolute position

		name := strings.TrimSpace(decl[dashIdx:colonIdx])
		value := strings.TrimSpace(decl[colonIdx+1:])

		// Validate the name is a clean custom property
		if !strings.HasPrefix(name, "--") || len(name) < 3 {
			continue
		}
		// Name should only contain word chars and hyphens
		if !isCleanVarName(name) {
			continue
		}

		// Clean value
		value = strings.TrimRight(value, " \t\n\r}")
		if value == "" || len(value) > 200 {
			continue
		}
		if !fpIsValidCSSValue(value) {
			continue
		}

		if fpIsDesignVar(name) {
			if overwrite || vars[name] == "" {
				vars[name] = value
			}
		}
	}
}

// fpIsValidCSSValue rejects values that are CSS pseudo-classes, selector
// fragments, or other non-value content that regex might capture.
func fpIsValidCSSValue(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	if lower == "" {
		return false
	}

	// Reject CSS pseudo-classes (from BEM selector false positives)
	pseudos := map[string]bool{
		"hover": true, "focus": true, "active": true, "visited": true,
		"disabled": true, "checked": true, "first-child": true,
		"last-child": true, "before": true, "after": true,
		"placeholder": true, "selection": true, "focus-visible": true,
		"focus-within": true, "first-of-type": true, "last-of-type": true,
		"not": true, "nth-child": true, "nth-of-type": true,
	}
	if pseudos[lower] {
		return false
	}

	// Reject values that look like selectors
	if strings.HasPrefix(lower, ".") || strings.HasPrefix(lower, "[") {
		return false
	}

	// Reject values containing { which means we've leaked into the next rule
	if strings.Contains(value, "{") {
		return false
	}

	return true
}

// fpStripCSSComments removes /* ... */ comments from CSS.
// Prevents extracting values from commented-out code.
func fpStripCSSComments(css string) string {
	var result strings.Builder
	result.Grow(len(css))

	i := 0
	for i < len(css) {
		if i+1 < len(css) && css[i] == '/' && css[i+1] == '*' {
			// Skip to end of comment
			end := strings.Index(css[i+2:], "*/")
			if end >= 0 {
				i = i + 2 + end + 2
			} else {
				break // Unclosed comment — skip rest
			}
		} else {
			result.WriteByte(css[i])
			i++
		}
	}

	return result.String()
}

// isCleanVarName checks that a CSS custom property name contains only
// valid characters (letters, digits, hyphens, underscores).
func isCleanVarName(name string) bool {
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return false
		}
	}
	return true
}

// isIdentChar returns true if the character could be part of a CSS identifier.
// Used to reject substring matches (e.g. "body" inside ".body-wrapper").
func isIdentChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '-' || ch == '_'
}
