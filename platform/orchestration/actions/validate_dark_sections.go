package actions

// validate_dark_section.go
// Validates that dark-section components set --section-* CSS variables.
// Hook points:
//   1. RenderComponentAction — warn after rendering if dark component missing vars
//   2. SavePageSectionsAction — warn when saving components
//   3. Could be a standalone action in component creation workflows

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// DarkSectionContract defines the required --section-* variables
var DarkSectionContract = []string{
	"--section-text",
	"--section-text-muted",
	"--section-heading",
	"--section-surface",
	"--section-border",
}

// DarkBackgroundIndicators are CSS patterns that suggest a dark background
var DarkBackgroundIndicators = []string{
	"background: var(--color-primary",
	"background-color: var(--color-primary",
	"background: #1a1a2e",
	"background: #0f172a",
	"background: #1a202c",
	"linear-gradient(135deg, #1a1a2e",
	"linear-gradient(135deg, #0f172a",
	// Hero with background-image overlay is dark
	"linear-gradient(rgba(0,0,0,0",
}

// ValidateDarkSectionContract checks if a component's CSS follows the dark section contract.
// Returns a list of missing variables (empty = valid).
// isDarkSection can be passed from content_components.is_dark_section or auto-detected.
func ValidateDarkSectionContract(html string, isDarkSection bool, logger *zap.Logger) []string {
	if !isDarkSection {
		// Auto-detect: check if the HTML/CSS contains dark background indicators
		isDarkSection = looksLikeDarkSection(html)
	}

	if !isDarkSection {
		return nil // Not a dark section, nothing to validate
	}

	var missing []string
	for _, v := range DarkSectionContract {
		if !strings.Contains(html, v) {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 && logger != nil {
		logger.Warn("Dark section component missing --section-* variables",
			zap.Int("missing_count", len(missing)),
			zap.Strings("missing_vars", missing),
			zap.String("html_preview", truncateForLog(html, 200)),
		)
	}

	return missing
}

// looksLikeDarkSection checks CSS content for dark background patterns
func looksLikeDarkSection(html string) bool {
	lower := strings.ToLower(html)
	for _, indicator := range DarkBackgroundIndicators {
		if strings.Contains(lower, strings.ToLower(indicator)) {
			return true
		}
	}
	return false
}

// FormatDarkSectionWarning creates a human-readable warning for logging/responses
func FormatDarkSectionWarning(componentName string, missing []string) string {
	if len(missing) == 0 {
		return ""
	}
	return fmt.Sprintf(
		"Component '%s' has a dark background but is missing section context variables: %s. "+
			"Add these to the component's container CSS for proper text color inheritance. "+
			"See dark section template in styles.css.",
		componentName,
		strings.Join(missing, ", "),
	)
}

func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
