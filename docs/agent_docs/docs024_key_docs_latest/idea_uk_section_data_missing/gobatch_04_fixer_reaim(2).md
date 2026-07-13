# Go batch — slice 3: fix_forced_text_colours re-aim (the contract as backstop)

Base: the uploaded `fix_forced_text_colours_action.go` (692 lines). The re-aim replaces
the `isDarkSection`-keyed decision core with a painting classifier + declaration
rewriter, per the paired-variable contract: what a section looks like derives from what
its own CSS paints; `is_dark_section` is metadata and never keys styling; painters
re-export `--section-*` as REFERENCES; non-painters declare nothing. The proven
literal-stripping machinery (removeChildTextColorsFromHTML and helpers) is unchanged.

Facts that shape the edits (verified in the file):
- The two selection queries do NOT filter on the flag — no query changes.
- `ensureSectionContractInHTML` / `injectSectionContract` / `sectionContractRe` have no
  callers outside the removed branch — they are deleted (this also removes the
  `logger.Debug` at :647).
- `isDarkSection` stays in both Scans and in the function signature (call sites
  unchanged); it is deliberately ignored inside — noted in code.
- `result.contractAdded` is REPURPOSED to count "declarations rewritten/removed";
  the callers' `contractsAdded` tally keeps its plumbing (name unchanged per the rule;
  meaning shift noted here).

## Pre-deploy confirmation (corrected — the palettes/css_variables guess was wrong)
The palette-band mapping uses `var(--color-primary-text, var(--color-background))`.
Two checks, either suffices; **absence does not block the deploy** — with no on-colour
token anywhere, the chain resolves to `--color-background`, the sensible on-primary
default for these schemes; the confirmation only decides the first link's spelling.

Zero-guess (the emitted tokens themselves), on a fresh deployed index.html:
```bash
grep -o '\-\-color-[a-z-]*' index.html | sort -u
# look for on-primary / primary-text among the emitted vars
```
DB-side (the join `loadSitePalette` itself uses — style_collections.color_palette):
```sql
SELECT sc.id, sc.name,
       (sc.color_palette::text LIKE '%on-primary%')   AS has_on_primary,
       (sc.color_palette::text LIKE '%primary-text%') AS has_primary_text
FROM style_collections sc
JOIN sites s ON s.style_collection_id = sc.id
WHERE s.domain = 'idea.uk';
-- layouts' variable set may also carry tokens: \d layouts first, then the same
-- LIKEs on its actual variables column.
```
**RESOLVED (2026-07-06):** the deployed page consumes `--color-primary-text` 18× and
`on-primary` 0×; the creator contract's item 7 lists `--color-primary-text` as the
available vocabulary. The three strings are switched to
`var(--color-primary-text, var(--color-background))` in the patched file. Definitions
live in `/assets/css/styles.css` (the page defines nothing itself) — one-line ship
check: `grep -c '\-\-color-primary-text\s*:' assets/css/styles.css` in the site repo;
if 0, the stylesheet omits a token the contract promises — slice-4 remit, not a blocker.

## Edit G — processComponentCSS: whole-function replacement
OLD (exact, :421 to the function's closing brace):
```go
func processComponentCSS(
	html, function string, isDarkSection bool,
	palette sitePalette, minContrast float64, logger *zap.Logger,
) cssFixResult {
	result := cssFixResult{html: html, worstRatio: 99.0}

	// Determine what the text and heading colors WILL BE after removing forced colors
	var expectedTextColor, expectedHeadingColor, bgColor string

	if isDarkSection {
		// Dark section: text gets --section-text (white-ish) or inherits body color
		// Headings get --section-heading (#fff) or --color-primary
		expectedTextColor = "#ffffff"    // --section-text resolves to ~white
		expectedHeadingColor = "#ffffff" // --section-heading resolves to white
		bgColor = palette.primary        // dark sections typically use primary as bg
	} else {
		// Light section: text inherits from body (--color-text)
		// Headings get --color-primary (fallback from --section-heading)
		expectedTextColor = palette.text
		expectedHeadingColor = palette.primary
		bgColor = palette.background
	}

	// Try to extract actual background color from the component's CSS
	if extracted := extractBgColor(html); extracted != "" {
		bgColor = extracted
	}

	// --- Contrast pre-check ---
	textRatio, textErr := wcagContrastRatio(expectedTextColor, bgColor)
	headingRatio, headingErr := wcagContrastRatio(expectedHeadingColor, bgColor)

	if textErr != nil || headingErr != nil {
		logger.Warn("processComponentCSS: contrast calc failed, skipping",
			zap.String("function", function),
			zap.Error(textErr),
		)
		result.skippedContrast = true
		result.worstRatio = 0
		return result
	}

	worst := math.Min(textRatio, headingRatio)
	result.worstRatio = math.Round(worst*100) / 100

	if worst < minContrast {
		logger.Warn("processComponentCSS: resulting contrast too low, skipping",
			zap.String("function", function),
			zap.Float64("text_ratio", textRatio),
			zap.Float64("heading_ratio", headingRatio),
			zap.Float64("min_required", minContrast),
			zap.String("bg", bgColor),
			zap.String("expected_text", expectedTextColor),
		)
		result.skippedContrast = true
		return result
	}

	// --- For dark sections missing --section-* contract, add it ---
	if isDarkSection {
		newHTML := ensureSectionContractInHTML(html, logger)
		if newHTML != html {
			html = newHTML
			result.contractAdded = true
		}
	}

	// --- Remove forced text colors from child element rules ---
	newHTML := removeChildTextColorsFromHTML(html)
	if newHTML != html {
		result.html = newHTML
		result.changed = true
		// Count how many declarations were removed
		result.colorsRemoved = countTextColorDeclarations(html) - countTextColorDeclarations(newHTML)
	}

	return result
}
```
NEW:
```go
func processComponentCSS(
	html, function string, isDarkSection bool,
	palette sitePalette, minContrast float64, logger *zap.Logger,
) cssFixResult {
	// NOTE (paired-variable re-aim): isDarkSection is deliberately IGNORED. A
	// section's appearance derives from what its own CSS paints; is_dark_section
	// is metadata only and must never key styling. The parameter is kept so the
	// two call sites and both Scans stay unchanged.
	_ = isDarkSection

	result := cssFixResult{html: html, worstRatio: 99.0}

	// --- Classify what the template's own CSS paints ---
	class, pairKind := classifySectionPainting(html)

	// --- Rewrite literal --section-* declarations to references per the contract:
	// pair painters re-export the pair; the ink model re-exports the ink;
	// palette/hex bands take the on-colour family; ambient painters must not
	// declare at all (their declarations are removed).
	newHTML := rewriteSectionDeclarationsInHTML(html, class, pairKind)
	if newHTML != html {
		html = newHTML
		result.html = html
		result.contractAdded = true // repurposed: declarations rewritten/removed
		result.changed = true
	}

	// --- Contrast pre-check for the literal-strip step (the safety property kept):
	// after stripping, text inherits the ambient chain — palette text on the
	// extracted/ambient background must clear minContrast or literals stay.
	expectedTextColor := palette.text
	expectedHeadingColor := palette.primary
	bgColor := palette.background
	if extracted := extractBgColor(html); extracted != "" {
		bgColor = extracted
	}
	textRatio, textErr := wcagContrastRatio(expectedTextColor, bgColor)
	headingRatio, headingErr := wcagContrastRatio(expectedHeadingColor, bgColor)
	if textErr != nil || headingErr != nil {
		logger.Warn("processComponentCSS: contrast calc failed, skipping strip",
			zap.String("function", function),
			zap.Error(textErr),
		)
		result.skippedContrast = true
		result.worstRatio = 0
		return result
	}
	worst := math.Min(textRatio, headingRatio)
	result.worstRatio = math.Round(worst*100) / 100
	if worst < minContrast {
		logger.Warn("processComponentCSS: ambient contrast too low, keeping literals",
			zap.String("function", function),
			zap.Float64("worst_ratio", result.worstRatio),
			zap.Float64("min_required", minContrast),
			zap.String("bg", bgColor),
		)
		result.skippedContrast = true
		return result
	}

	// --- Remove forced text colors from child element rules (unchanged) ---
	newHTML = removeChildTextColorsFromHTML(html)
	if newHTML != html {
		result.html = newHTML
		result.changed = true
		// Count how many declarations were removed
		result.colorsRemoved = countTextColorDeclarations(html) - countTextColorDeclarations(newHTML)
	}

	return result
}
```

## Edit G2 — delete the injector trio
Delete these, whole (no callers remain after Edit G; `go build` confirms):
- the `sectionContractRe` var (:418–419),
- `func ensureSectionContractInHTML(...)` (:611 header through its closing brace),
- `func injectSectionContract(...)` (:643 header through its closing brace — this
  removes the `logger.Debug` at :647).

## Edit G3 — NEW code: the classifier and rewriter
Insert directly after the existing regex vars (i.e. where `sectionContractRe` was):
```go
// paintClass describes what a template's own CSS paints, derived from the CSS
// itself — never from is_dark_section.
type paintClass int

const (
	paintAmbient     paintClass = iota // background/surface vars, or no background of its own
	paintPair                          // var(--color-<kind>-bg) band (cta/header/footer)
	paintInk                           // --hero-ink model (layered/image sections)
	paintPaletteBand                   // primary/secondary/accent band, or a literal hex band
)

var pairBgRe = regexp.MustCompile(`var\(--color-(cta|header|footer)-bg`)
var inkModelRe = regexp.MustCompile(`--hero-ink`)
var paletteBandBgRe = regexp.MustCompile(`background[^;{}]*var\(--color-(primary|secondary|accent)\b`)
var hexBandBgRe = regexp.MustCompile(`background(?:-color)?\s*:\s*[^;{}]*#[0-9a-fA-F]{3,8}`)

// sectionDeclRe matches a full --section-* custom-property declaration.
var sectionDeclRe = regexp.MustCompile(`--section-(bg|text|heading|muted|border|link)\s*:\s*[^;}]+;?`)

// classifySectionPainting inspects the template's CSS and returns its painting
// class; for paintPair it also returns the pair kind (cta/header/footer).
func classifySectionPainting(html string) (paintClass, string) {
	if m := pairBgRe.FindStringSubmatch(html); m != nil {
		return paintPair, m[1]
	}
	if inkModelRe.MatchString(html) {
		return paintInk, ""
	}
	if paletteBandBgRe.MatchString(html) || hexBandBgRe.MatchString(html) {
		return paintPaletteBand, ""
	}
	return paintAmbient, ""
}

// rewriteSectionDeclarationsInHTML converts literal --section-* declarations in
// <style> blocks to references appropriate to the painting class, and removes
// them entirely for ambient (non-painting) sections. Declarations that already
// reference vars or color-mix are left alone.
func rewriteSectionDeclarationsInHTML(html string, class paintClass, pairKind string) string {
	return styleBlockRe.ReplaceAllStringFunc(html, func(block string) string {
		m := styleBlockRe.FindStringSubmatch(block)
		if len(m) < 4 {
			return block
		}
		newCSS := sectionDeclRe.ReplaceAllStringFunc(m[2], func(decl string) string {
			sub := sectionDeclRe.FindStringSubmatch(decl)
			if len(sub) < 2 {
				return decl
			}
			name := sub[1]
			if strings.Contains(decl, "var(") || strings.Contains(decl, "color-mix(") {
				return decl // already a reference
			}
			switch class {
			case paintPair:
				switch name {
				case "bg":
					return "--section-bg: var(--color-" + pairKind + "-bg);"
				case "muted":
					return "--section-muted: color-mix(in srgb, var(--color-" + pairKind + "-text) 70%, transparent);"
				case "border":
					return "--section-border: color-mix(in srgb, var(--color-" + pairKind + "-text) 25%, transparent);"
				default: // text, heading, link
					return "--section-" + name + ": var(--color-" + pairKind + "-text);"
				}
			case paintInk:
				switch name {
				case "bg":
					return decl // the template paints its own layered background
				case "muted":
					return "--section-muted: color-mix(in srgb, var(--hero-ink) 75%, transparent);"
				case "border":
					return "--section-border: color-mix(in srgb, var(--hero-ink) 30%, transparent);"
				default:
					return "--section-" + name + ": var(--hero-ink);"
				}
			case paintPaletteBand:
				switch name {
				case "bg":
					return decl // the band's own background stands (hex bands are fix_hardcoded's territory)
				case "muted":
					return "--section-muted: color-mix(in srgb, var(--color-primary-text, var(--color-background)) 70%, transparent);"
				case "border":
					return "--section-border: color-mix(in srgb, var(--color-primary-text, var(--color-background)) 25%, transparent);"
				default:
					return "--section-" + name + ": var(--color-primary-text, var(--color-background));"
				}
			default: // paintAmbient — non-painters must not declare
				return ""
			}
		})
		if newCSS == m[2] {
			return block
		}
		return m[1] + newCSS + m[3]
	})
}
```
New identifiers introduced deliberately (noted per the rule): `paintClass`,
`paintAmbient/paintPair/paintInk/paintPaletteBand`, `pairBgRe`, `inkModelRe`,
`paletteBandBgRe`, `hexBandBgRe`, `sectionDeclRe`, `classifySectionPainting`,
`rewriteSectionDeclarationsInHTML`, and locals within them. Nothing existing renames.

## Coupling note (3b — no change this slice)
`fix_hardcoded_colours` keeps mapping dark background hex → `var(--color-primary)`,
which is coherent with `paintPaletteBand` here (a hex band becomes a primary band,
whose on-colour family this fixer then enforces). Loop ordering between the two
fixers is unchanged.

## Build + first run
`go build ./...`; rides the batch image. The fixer is improvement-loop-invoked — the
first live run is the tail decision's vehicle: point it at a non-idea.uk site first
(manual spawn per 016b) and read the returned `details` JSON before letting the loop
run it broadly.
