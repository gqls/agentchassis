// FILE: platform/orchestration/actions/css_templating.go
//
// Helpers for converting rendered CSS into Go-templated form so that
// the palette values in a :root block become {{.Primary}} etc. placeholders.
// Used by fork_theme_from_site to produce the css_template column value
// from an already-rendered css_content snapshot.
//
// Safety: operates only on the :root { ... } block. Hex values outside
// :root may be shadows, borders, or decorative highlights that are not
// palette references — touching them would corrupt unrelated styles.

package actions

import (
	"regexp"
	"strings"
)

// rootBlockRE finds the :root { ... } block at the top of a stylesheet.
// The pattern is permissive: leading whitespace, case-insensitive, and
// tolerates a newline between :root and the opening brace.
var rootBlockRE = regexp.MustCompile(`(?i):root\s*\{[^}]*\}`)

// paletteTemplateField maps a design_spec.color_scheme key to the Go template
// placeholder used in cssTemplateData (see render_css_from_spec_action.go).
// Keep this in sync with the struct — the pairing here is the contract.
var paletteTemplateField = map[string]string{
	"primary":    "{{.Primary}}",
	"secondary":  "{{.Secondary}}",
	"accent":     "{{.Accent}}",
	"background": "{{.Background}}",
	"surface":    "{{.Surface}}",
	"text":       "{{.Text}}",
	"text_muted": "{{.TextMuted}}",
	"border":     "{{.Border}}",
}

// typographyTemplateField maps a design_spec.typography key to its placeholder.
var typographyTemplateField = map[string]string{
	"font_family":  "{{.FontFamily}}",
	"heading_font": "{{.HeadingFont}}",
	"base_size":    "{{.BaseSize}}",
	"line_height":  "{{.LineHeight}}",
}

// spacingTemplateField maps a design_spec.spacing key to its placeholder.
var spacingTemplateField = map[string]string{
	"section_padding":     "{{.SectionPadding}}",
	"container_max_width": "{{.ContainerMaxWidth}}",
}

type cssRepl struct {
	from, to string
}

// TemplateCSSFromSpec replaces hardcoded palette/typography/spacing values
// inside the :root block of renderedCSS with their Go template placeholders.
// Everything outside :root is left unchanged.
//
// spec is the design_spec shape produced by the webdesign-agent's analyze_design
// step — typically { "color_scheme": {...}, "typography": {...}, "spacing": {...} }.
//
// Returns the templated CSS. If renderedCSS has no :root block, returns the
// input unchanged — the caller gets a template that will render identically
// for any palette (effectively a static theme stored in css_template form).
func TemplateCSSFromSpec(renderedCSS string, spec map[string]interface{}) string {
	rootBlock := rootBlockRE.FindString(renderedCSS)
	if rootBlock == "" {
		return renderedCSS
	}

	templatedRoot := rootBlock

	// Build the (hex → placeholder) substitutions from the spec, grouped by
	// section. Collect into a single slice so we can sort by length descending
	// before applying — this prevents a shorter hex from being replaced inside
	// a longer one (#1a1a2e inside #1a1a2eff), even though the :root block is
	// unlikely to contain #rrggbbaa forms in practice.
	var replacements []cssRepl

	collect := func(section string, fieldMap map[string]string) {
		raw, ok := spec[section].(map[string]interface{})
		if !ok {
			return
		}
		for key, placeholder := range fieldMap {
			val, ok := raw[key].(string)
			if !ok || val == "" {
				continue
			}
			replacements = append(replacements, cssRepl{from: val, to: placeholder})
		}
	}

	collect("color_scheme", paletteTemplateField)
	collect("typography", typographyTemplateField)
	collect("spacing", spacingTemplateField)

	// Sort by `from` length descending so longer matches replace first.
	sortReplacementsByLengthDesc(replacements)

	for _, r := range replacements {
		if r.from == "" {
			continue
		}
		templatedRoot = strings.ReplaceAll(templatedRoot, r.from, r.to)
	}

	// Swap the original :root block for the templated one.
	return strings.Replace(renderedCSS, rootBlock, templatedRoot, 1)
}

// sortReplacementsByLengthDesc sorts in place, longer `from` values first.
// A manual implementation avoids pulling in the sort package for a function
// with <20 entries in practice.
func sortReplacementsByLengthDesc(rs []cssRepl) {
	for i := 1; i < len(rs); i++ {
		for j := i; j > 0 && len(rs[j-1].from) < len(rs[j].from); j-- {
			rs[j-1], rs[j] = rs[j], rs[j-1]
		}
	}
}
