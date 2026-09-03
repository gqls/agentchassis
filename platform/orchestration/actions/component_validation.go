// FILE: platform/orchestration/actions/component_validation.go
//
// Component Naming Contract enforcement.
// See 008_component_naming_contract.md for the full specification.
//
// Rules:
//   - content_components.function is the canonical identifier
//   - Always kebab-case: lowercase letters, digits, hyphens
//   - data-component attribute in html_template must equal function
//   - page_components.slot_name must store the kebab-case function value
//
// Integration points where these functions are called:
//   - GetComponentWithFallback (component_library.go):  NormalizeComponentFunction
//   - RenderComponentAction (v3_site_actions.go):       NormalizeComponentFunction
//   - LoadPageSectionComponentsAction (v3_site_actions): NormalizeSectionNames
//   - extractSectionsFromMetadata (save_page_sections): NormalizeComponentFunction
//   - saveSectionsExtractFromHTML (save_page_sections):  NormalizeComponentFunction
//   - enrichSectionsWithComponentIDs (save_page_sections): NormalizeComponentFunction

package actions

import (
	"fmt"
	"regexp"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var (
	dataComponentAttrRe = regexp.MustCompile(`data-component="([^"]+)"`)
)

// ValidateComponentFunction checks that a component function name follows
// the naming contract (kebab-case). Returns nil if valid.
func ValidateComponentFunction(function string) error {
	if function == "" {
		return nil
	}
	if !datahelpers.IsKebabCase(function) {
		suggested := NormalizeComponentFunction(function)
		return fmt.Errorf(
			"component function %q is not kebab-case. Suggested: %q",
			function, suggested,
		)
	}
	return nil
}

// ValidateComponentTemplate checks that html_template's data-component
// attribute matches the function value. Returns nil if valid or if
// template has no data-component attribute (headers/footers/heads don't have one).
func ValidateComponentTemplate(function, htmlTemplate string) error {
	if function == "" || htmlTemplate == "" {
		return nil
	}
	match := dataComponentAttrRe.FindStringSubmatch(htmlTemplate)
	if match == nil {
		return nil
	}
	if match[1] != function {
		return fmt.Errorf(
			"data-component=%q does not match function=%q",
			match[1], function,
		)
	}
	return nil
}

// NormalizeComponentFunction converts a function name to kebab-case.
//
//	"social_proof"    → "social-proof"
//	"call_to_action"  → "call-to-action"
//	"SocialProof"     → "social-proof"
//	"social-proof"    → "social-proof" (no-op)
//	""                → ""             (no-op)
//
// Delegates to datahelpers.NormalizeComponentFunction (moved down 2026-08-15,
// bugs_open/285) so the save guard's matchLockedRow here and the loader-side
// MergeLockedPageSlots in datahelpers normalise with ONE function — the same
// shape as pageComponentAgentWritableSQL → datahelpers.AgentWritableSQLFor.
func NormalizeComponentFunction(function string) string {
	return datahelpers.NormalizeComponentFunction(function)
}

// NormalizeSectionNames normalizes a slice of section names in-place,
// converting underscores to hyphens and fixing case.
// Logs any names that required conversion.
func NormalizeSectionNames(names []string, logger *zap.Logger) {
	for i, name := range names {
		normalized := NormalizeComponentFunction(name)
		if normalized != name {
			logger.Info("NormalizeSectionNames: Converted section name to kebab-case",
				zap.String("original", name),
				zap.String("normalized", normalized),
			)
			names[i] = normalized
		}
	}
}

// sectionLookupKeys returns the distinct identifiers a requested section name
// should be matched against when resolving it to a component: the raw name and
// its kebab-normalised form. Components are stored kebab-case
// (content_components.function, per the naming contract above), but a site plan
// or adoption flow may emit "call_to_action" or "SocialProof". Trying BOTH forms
// is a strict SUPERSET of trying the raw string alone — every section that
// resolved before still resolves, and "call_to_action" now also matches the
// stored "call-to-action".
//
// Returning both (rather than REPLACING raw with normalised) is load-bearing: a
// handful of live components carry a snake_case *name* whose function is a
// different kebab string — e.g. name "featured_article", function
// "featured-content". Those resolve ONLY by their raw name; normalise-then-
// look-up would silently drop them. See
// /bugs_open/041 (section-lookup-never-normalises).
func sectionLookupKeys(name string) []string {
	norm := NormalizeComponentFunction(name)
	if norm == "" || norm == name {
		return []string{name}
	}
	return []string{name, norm}
}

// sectionResolvedByFound reports whether a section name is satisfied by the set
// of component name/function values already loaded, matching under either its
// raw or normalised form.
func sectionResolvedByFound(found map[string]bool, name string) bool {
	for _, k := range sectionLookupKeys(name) {
		if found[k] {
			return true
		}
	}
	return false
}

// sectionLookupValueSet flattens sectionLookupKeys over many names into a
// de-duplicated slice suitable for a SQL IN-list. Order is preserved so the
// query args stay stable for a given input.
func sectionLookupValueSet(names []string) []string {
	seen := make(map[string]bool, len(names)*2)
	out := make([]string, 0, len(names)*2)
	for _, n := range names {
		for _, k := range sectionLookupKeys(n) {
			if !seen[k] {
				seen[k] = true
				out = append(out, k)
			}
		}
	}
	return out
}

// --- CSS custom-property vocabulary audit (R6f, 2026-07-06) -----------------
//
// Component templates consume CSS custom properties (var(--name)). The theme
// stylesheet, rendered deterministically by RenderCSSFromSpecAction, defines a
// canonical token set plus renderer-enforced compatibility aliases (see
// buildTokenAliases in render_css_from_spec_action.go). A template that
// references a name outside that set renders with the browser fallback (or
// nothing) — the R6f "dark-on-dark" failure. This audit is a DETECTION NET,
// not a gate: it logs unknown names so drift is visible, and NEVER rejects a
// template (vocabulary evolves; a false reject would block legitimate saves).

// canonicalCSSTokens is the set of custom-property names a component template
// may reference: the theme's defined vocabulary plus the compatibility aliases
// that render_css_from_spec guarantees. Keep in sync with the theme CSS
// template and tokenAliases. A name absent here is reported by
// AuditTemplateTokens (warn only).
var canonicalCSSTokens = map[string]struct{}{
	// palette / theme-defined
	"--color-primary": {}, "--color-primary-hover": {}, "--color-primary-text": {},
	"--color-secondary": {}, "--color-accent": {}, "--color-background": {},
	"--color-surface": {}, "--color-card-bg": {}, "--color-text": {},
	"--color-text-muted": {}, "--color-border": {}, "--color-cta-bg": {},
	"--color-cta-text": {}, "--color-header-bg": {}, "--color-header-text": {},
	"--color-footer-bg": {}, "--color-footer-text": {},
	"--section-text": {}, "--section-text-muted": {}, "--section-surface": {},
	"--section-border": {}, "--section-heading": {}, "--section-pad-y": {},
	"--section-pad-y-sm": {},
	"--radius":           {}, "--radius-sm": {}, "--radius-lg": {},
	"--shadow-sm": {}, "--shadow-md": {}, "--shadow-lg": {},
	"--container-max": {}, "--container-pad-x": {},
	"--font-body": {}, "--font-heading": {}, "--font-size-base": {},
	"--line-height-base": {}, "--grid-gap": {}, "--card-pad": {}, "--transition": {},
	// compatibility aliases guaranteed by buildTokenAliases (render_css_from_spec)
	"--border-radius": {}, "--shadow": {}, "--spacing-section": {},
	"--container-max-width": {}, "--primary-color": {}, "--secondary-color": {},
	"--accent-color": {}, "--color-heading": {}, "--color-white": {},
	"--color-error": {}, "--hero-ink": {},
	// renderer-owned legible-ink companions, emitted by buildLegibleInkDefaults
	// (palette_specialised_slots.go). They are the documented repair for a
	// palette colour that is illegible in the direction a component uses it,
	// so a template opting into one must not be audited as unknown drift.
	"--color-primary-ink": {}, "--color-accent-ink": {}, "--color-cta-bg-ink": {},
	"--color-accent-text": {},
}

var cssVarRefRe = regexp.MustCompile(`var\(\s*(--[a-z0-9-]+)`)

// AuditTemplateTokens returns the distinct CSS custom-property names a template
// references via var(--…) that are NOT in canonicalCSSTokens, in first-seen
// order. It never errors. When logger is non-nil and unknowns are found it logs
// one Warn with the function and the unknown names. Callers may use the return
// value for metrics; the template is always allowed to persist.
func AuditTemplateTokens(function, htmlTemplate string, logger *zap.Logger) []string {
	if htmlTemplate == "" {
		return nil
	}
	seen := make(map[string]struct{})
	var unknown []string
	for _, m := range cssVarRefRe.FindAllStringSubmatch(htmlTemplate, -1) {
		name := m[1]
		if _, ok := canonicalCSSTokens[name]; ok {
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		unknown = append(unknown, name)
	}
	if len(unknown) > 0 && logger != nil {
		logger.Warn("component template references non-canonical CSS custom properties",
			zap.String("function", function),
			zap.Strings("unknown_tokens", unknown),
			zap.Int("unknown_count", len(unknown)))
	}
	return unknown
}
