// FILE: platform/orchestration/input_contracts/input_mapping.go
//
// CHANGE from current:
//   - ResolveInputMapping and ResolveInputMappingWithItem switch from
//     GetValueAtExactPath (literal path traversal) to datahelpers.FindByPath
//     (literal + .response auto-unwrap + UnwrapDeep fallback).
//
// Why:
//   The dev guide documents auto-unwrap behaviour as the canonical convention:
//   "datahelpers.ExtractNestedFieldString already does exactly this, with
//    .response auto-unwrapping as a bonus." But ResolveInputMapping (the
//   function call_agent uses for input_mapping) was using a different
//   resolver that didn't honour the convention. This patch closes that gap.
//
// Backward compatibility:
//   FindByPath tries direct lookup first, so any input_mapping path that
//   previously resolved literally still resolves literally. The only
//   behaviour change is that previously-failing paths that match through
//   .response auto-unwrap now resolve — which is the intent.
//
// Note: "found" was previously a bool from GetValueAtExactPath. FindByPath
// returns interface{}, where nil means not-found. The semantic is preserved.

package input_contracts

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// InputMapping defines explicit source paths for input fields
// Key = destination field name (what child receives)
// Value = source path in CollectedData (where to get it)
// Special value "$item" is used in fan_out to represent the current iteration item
type InputMapping map[string]string

// InputContract defines what an agent expects to receive
type InputContract struct {
	Required []string `json:"required"`
	Optional []string `json:"optional,omitempty"`
}

// OutputContract defines what an agent produces
type OutputContract struct {
	Produces []string `json:"produces"`
}

// ResolveInputMapping builds input data using explicit paths from the mapping.
// Path resolution uses datahelpers.FindByPath which supports:
//   - Literal dot-path traversal (primary)
//   - .response auto-unwrap for call_agent / spawn_agent response wrappers
//   - input_data prefix variations
//   - UnwrapDeep recursive unwrapping for deeply nested results
//
// Returns error if any required mapping path is not found (hard fail).
// Fields marked with "?" suffix on the destination are silently skipped if
// the source path doesn't resolve.
func ResolveInputMapping(
	collectedData map[string]interface{},
	mapping InputMapping,
	logger *zap.Logger,
) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for destField, sourcePath := range mapping {
		// Check if field is optional (ends with ?)
		isOptional := strings.HasSuffix(destField, "?")
		actualDestField := strings.TrimSuffix(destField, "?")

		// Handle special $item token (fan_out replaces this before calling)
		if sourcePath == "$item" {
			// This will be handled by the fan_out action which replaces $item
			// with the actual current item before calling ResolveInputMapping
			continue
		}

		// Handle empty source path
		if sourcePath == "" {
			logger.Warn("Empty source path in input_mapping",
				zap.String("dest_field", actualDestField))
			continue
		}

		// CHANGED: use FindByPath (auto-unwraps .response) instead of
		// GetValueAtExactPath (literal-only). See file header for rationale.
		value := datahelpers.FindByPath(collectedData, sourcePath, logger)

		if value == nil {
			if isOptional {
				// Optional field not found - just skip it
				logger.Debug("Optional field not found in input_mapping, skipping",
					zap.String("dest_field", actualDestField),
					zap.String("source_path", sourcePath))
				continue
			}
			// Required field not found - error
			availablePaths := ListAvailablePaths(collectedData, 2)
			return nil, fmt.Errorf(
				"input_mapping failed: source path '%s' not found for field '%s'\n"+
					"Available top-level paths: %v",
				sourcePath, actualDestField, availablePaths,
			)
		}

		result[actualDestField] = value
		logger.Debug("Resolved input mapping",
			zap.String("dest", actualDestField),
			zap.String("source", sourcePath),
			zap.Bool("optional", isOptional))
	}

	return result, nil
}

// ResolveInputMappingWithItem is like ResolveInputMapping but also handles
// the $item token by replacing it with the provided currentItem value.
// Uses the same FindByPath resolver as ResolveInputMapping.
func ResolveInputMappingWithItem(
	collectedData map[string]interface{},
	mapping InputMapping,
	currentItem interface{},
	logger *zap.Logger,
) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for destField, sourcePath := range mapping {
		// Check if field is optional (ends with ?)
		isOptional := strings.HasSuffix(destField, "?")
		actualDestField := strings.TrimSuffix(destField, "?")

		// Handle $item — pass through directly
		if sourcePath == "$item" {
			result[actualDestField] = currentItem
			logger.Debug("Resolved $item in input mapping",
				zap.String("dest", actualDestField))
			continue
		}

		// Handle empty source path
		if sourcePath == "" {
			logger.Warn("Empty source path in input_mapping (with_item)",
				zap.String("dest_field", actualDestField))
			continue
		}

		value := datahelpers.FindByPath(collectedData, sourcePath, logger)

		if value == nil {
			if isOptional {
				// Optional field not found - just skip it
				logger.Debug("Optional field not found in input_mapping, skipping",
					zap.String("dest_field", actualDestField),
					zap.String("source_path", sourcePath))
				continue
			}
			// Required field not found - error
			availablePaths := ListAvailablePaths(collectedData, 2)
			return nil, fmt.Errorf(
				"input_mapping (with_item) failed: source path '%s' not found for field '%s'\n"+
					"Available top-level paths: %v",
				sourcePath, actualDestField, availablePaths,
			)
		}

		result[actualDestField] = value
		logger.Debug("Resolved input mapping",
			zap.String("dest", actualDestField),
			zap.String("source", sourcePath),
			zap.Bool("optional", isOptional))
	}

	return result, nil
}

// GetValueAtExactPath retrieves a value at an exact dot-notation path.
// No fallbacks, no hunting - just the exact path specified.
// Returns (value, true) if found, (nil, false) if not found.
func GetValueAtExactPath(data map[string]interface{}, path string) (interface{}, bool) {
	if path == "" {
		return nil, false
	}

	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, exists := v[part]
			if !exists {
				return nil, false
			}
			current = val
		default:
			// Current value is not a map, can't traverse further
			return nil, false
		}
	}

	return current, true
}

// ValidateInputContract checks that data satisfies an agent's input contract.
// Returns error with details if required fields are missing.
func ValidateInputContract(
	agentType string,
	data map[string]interface{},
	contract *InputContract,
	logger *zap.Logger,
) error {
	if contract == nil {
		logger.Debug("No input contract defined, skipping validation",
			zap.String("agent_type", agentType))
		return nil
	}

	var missing []string

	for _, required := range contract.Required {
		// Check if field exists at top level
		if _, exists := data[required]; !exists {
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		providedFields := MapKeys(data)
		return fmt.Errorf(
			"contract violation for agent '%s': missing required fields: %v\n"+
				"Provided fields: %v\n"+
				"Hint: Check input_mapping in the step config",
			agentType, missing, providedFields,
		)
	}

	logger.Debug("Input contract validated successfully",
		zap.String("agent_type", agentType),
		zap.Int("required_count", len(contract.Required)),
		zap.Int("provided_count", len(data)))

	return nil
}

// ListAvailablePaths returns available paths in data for error messages.
// maxDepth controls how deep to traverse nested maps.
func ListAvailablePaths(data map[string]interface{}, maxDepth int) []string {
	var paths []string
	var walk func(prefix string, d map[string]interface{}, depth int)

	walk = func(prefix string, d map[string]interface{}, depth int) {
		for k, v := range d {
			// Skip internal/meta fields
			if strings.HasPrefix(k, "__") {
				continue
			}

			path := k
			if prefix != "" {
				path = prefix + "." + k
			}
			paths = append(paths, path)

			if depth < maxDepth {
				if nested, ok := v.(map[string]interface{}); ok {
					walk(path, nested, depth+1)
				}
			}
		}
	}

	walk("", data, 0)
	sort.Strings(paths)
	return paths
}

// MapKeys returns sorted keys from a map
func MapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// GetAgentInputContract retrieves the input contract for an agent type from the database.
func GetAgentInputContract(ctx context.Context, db *sql.DB, agentType string, logger *zap.Logger) (*InputContract, error) {
	if db == nil {
		logger.Debug("No database connection, skipping contract lookup",
			zap.String("agent_type", agentType))
		return nil, nil
	}

	var contractJSON sql.NullString
	query := `SELECT input_contract FROM agent_definitions WHERE type = $1 AND is_active = true ORDER BY version DESC LIMIT 1`

	err := db.QueryRowContext(ctx, query, agentType).Scan(&contractJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug("No agent definition found",
				zap.String("agent_type", agentType))
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query agent contract: %w", err)
	}

	if !contractJSON.Valid || contractJSON.String == "" || contractJSON.String == "null" {
		logger.Debug("No input contract defined for agent",
			zap.String("agent_type", agentType))
		return nil, nil
	}

	var contract InputContract
	if err := json.Unmarshal([]byte(contractJSON.String), &contract); err != nil {
		logger.Warn("Failed to parse input contract JSON",
			zap.String("agent_type", agentType),
			zap.Error(err))
		return nil, nil
	}

	logger.Debug("Loaded input contract",
		zap.String("agent_type", agentType),
		zap.Strings("required", contract.Required),
		zap.Strings("optional", contract.Optional))

	return &contract, nil
}

// ParseInputMapping extracts and validates the input_mapping from step config.
// (Unchanged — kept here for completeness of this file's surface.)
func ParseInputMapping(config map[string]interface{}) (InputMapping, bool) {
	raw, ok := config["input_mapping"]
	if !ok {
		return nil, false
	}

	switch v := raw.(type) {
	case map[string]interface{}:
		mapping := make(InputMapping)
		for k, val := range v {
			if strVal, ok := val.(string); ok {
				mapping[k] = strVal
			}
		}
		if len(mapping) == 0 {
			return nil, false
		}
		return mapping, true
	case map[string]string:
		mapping := make(InputMapping)
		for k, v := range v {
			mapping[k] = v
		}
		return mapping, true
	}
	return nil, false
}

// ConvertInputFieldsToMapping converts legacy input_fields array to input_mapping format.
// This is used for backward compatibility during migration.
// e.g., ["current_page", "site_record"] becomes {"current_page": "current_page", "site_record": "site_record"}
func ConvertInputFieldsToMapping(inputFields []interface{}, logger *zap.Logger) InputMapping {
	result := make(InputMapping)

	for _, f := range inputFields {
		fieldName, ok := f.(string)
		if !ok {
			continue
		}

		// For legacy input_fields, the source and dest are the same
		// The field name is used as-is for both
		result[fieldName] = fieldName
	}

	if len(result) > 0 {
		logger.Info("Converted legacy input_fields to input_mapping (consider updating workflow config)",
			zap.Int("field_count", len(result)))
	}

	return result
}
