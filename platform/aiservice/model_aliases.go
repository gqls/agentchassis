// ============================================================================
// FILE: platform/aiservice/model_aliases.go
//
// Resolves model aliases to actual Anthropic API model names.
// Update this file when Anthropic releases new model versions.
// ============================================================================

package aiservice

import (
	"go.uber.org/zap"
)

// ModelAliases maps user-friendly aliases to actual API model names
// Update these when Anthropic releases new model versions
var ModelAliases = map[string]string{
	// Claude 4.6 family
	"claude-sonnet-4-6": "claude-sonnet-4-6",
	"claude-opus-4-6":   "claude-opus-4-6",

	// Claude 4.5 family
	"claude-sonnet-4-5": "claude-sonnet-4-5-20250929",
	"claude-haiku-4-5":  "claude-haiku-4-5-20251001",
	"claude-opus-4-5":   "claude-opus-4-5-20251101",

	// Claude 4 family
	"claude-sonnet-4": "claude-sonnet-4-20250514",
	"claude-opus-4":   "claude-opus-4-20250514",

	// Claude 3.5 family
	"claude-3-5-sonnet": "claude-3-5-sonnet-20241022",
	"claude-3-5-haiku":  "claude-3-5-haiku-20241022",

	// Claude 3 family (legacy)
	"claude-3-opus":   "claude-3-opus-20240229",
	"claude-3-sonnet": "claude-3-sonnet-20240229",
	"claude-3-haiku":  "claude-3-haiku-20240307",
}

// ResolveModelAlias converts a model alias to the actual API model name.
// If the input is not an alias (already a full model name), returns it unchanged.
func ResolveModelAlias(model string, logger *zap.Logger) string {
	if resolved, isAlias := ModelAliases[model]; isAlias {
		if logger != nil {
			logger.Debug("Resolved model alias",
				zap.String("alias", model),
				zap.String("resolved", resolved),
			)
		}
		return resolved
	}
	return model
}

// GetAvailableAliases returns a list of all available model aliases
func GetAvailableAliases() []string {
	aliases := make([]string, 0, len(ModelAliases))
	for alias := range ModelAliases {
		aliases = append(aliases, alias)
	}
	return aliases
}

// IsValidModel checks if a model name is either a valid alias or a known full model name
func IsValidModel(model string) bool {
	// Check if it's an alias
	if _, isAlias := ModelAliases[model]; isAlias {
		return true
	}

	// Check if it's a known full model name (value in the map)
	for _, fullName := range ModelAliases {
		if model == fullName {
			return true
		}
	}

	// Could be a new model we don't know about - allow it through
	// The API will validate it
	return true
}
