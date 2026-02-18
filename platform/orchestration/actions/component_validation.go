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
	"strings"

	"go.uber.org/zap"
)

var (
	kebabCaseRe         = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	dataComponentAttrRe = regexp.MustCompile(`data-component="([^"]+)"`)
)

// ValidateComponentFunction checks that a component function name follows
// the naming contract (kebab-case). Returns nil if valid.
func ValidateComponentFunction(function string) error {
	if function == "" {
		return nil
	}
	if !kebabCaseRe.MatchString(function) {
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
func NormalizeComponentFunction(function string) string {
	if function == "" {
		return ""
	}

	// Already valid kebab-case — fast path
	if kebabCaseRe.MatchString(function) {
		return function
	}

	// Replace underscores with hyphens
	result := strings.ReplaceAll(function, "_", "-")

	// Insert hyphen before uppercase runs (camelCase → kebab-case)
	var b strings.Builder
	for i, r := range result {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r + 32) // to lowercase
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
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
