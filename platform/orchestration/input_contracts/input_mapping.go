// FILE: platform/orchestration/input_contracts/input_mapping.go
//
// REVISION HISTORY:
//   v1: Switched ResolveInputMapping from GetValueAtExactPath to
//       datahelpers.FindByPath alone. This added .response auto-unwrap but
//       regressed null-tolerance: paths whose value is JSON null
//       (e.g. input_data.orchestration_id: null in a test trigger) failed
//       because FindByPath returns nil for both "not found" and "found null".
//
//   v2 (this file): Two-step resolution preserves both behaviours:
//     Step 1 — GetValueAtExactPath: returns (value, found). If the path
//              exists literally, store value as-is (may be nil — that's a
//              valid result, distinct from missing).
//     Step 2 — datahelpers.FindByPath: fallback when literal path missed,
//              for .response auto-unwrap, input_data prefix variations,
//              and UnwrapDeep recursive unwrapping.
//
// Why both:
//   The dev guide documents auto-unwrap as the canonical convention
//   ("datahelpers.ExtractNestedFieldString already does exactly this, with
//   .response auto-unwrapping as a bonus"). But FindByPath's interface{}
//   return value can't distinguish "found nil" from "not found". The
//   two-step approach gives literal-path-exists precedence (preserving
//   null tolerance) while still gaining auto-unwrap when literal misses.

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

// resolveSourcePath performs the two-step path resolution shared by
// ResolveInputMapping and ResolveInputMappingWithItem:
//
//	Step 1 — GetValueAtExactPath: literal lookup, null-tolerant.
//	         Returns (value, found) where found=true even if value is nil.
//	Step 2 — datahelpers.FindByPath: auto-unwrap fallback when the literal
//	         path doesn't exist. Returns nil for not-found.
//
// Returns (value, found). value may be nil when the literal path exists
// with a null value; that's distinct from "path not found".
func resolveSourcePath(
	collectedData map[string]interface{},
	sourcePath string,
	logger *zap.Logger,
) (interface{}, bool) {
	// Step 1: literal lookup. Found-with-nil is a valid result.
	if value, found := GetValueAtExactPath(collectedData, sourcePath); found {
		return value, true
	}

	// Step 2: auto-unwrap fallback via datahelpers.FindByPath.
	if value := datahelpers.FindByPath(collectedData, sourcePath, logger); value != nil {
		return value, true
	}

	return nil, false
}

// ResolveInputMapping builds input data using explicit paths from the mapping.
// Path resolution is two-step (see resolveSourcePath): literal lookup first
// for null-tolerance, then datahelpers.FindByPath for auto-unwrap of
// .response wrappers, input_data prefix variants, and UnwrapDeep results.
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

		// Two-step resolution: literal first (null-tolerant), then auto-unwrap.
		value, found := resolveSourcePath(collectedData, sourcePath, logger)

		if !found {
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

		// value may be nil here — that's valid if found=true (path exists, value is null)
		result[actualDestField] = value
		logger.Debug("Resolved input mapping",
			zap.String("dest", actualDestField),
			zap.String("source", sourcePath),
			zap.Bool("optional", isOptional),
			zap.Bool("value_nil", value == nil))
	}

	return result, nil
}

// ResolveInputMappingWithItem is like ResolveInputMapping but also handles
// the $item token by replacing it with the provided currentItem value.
// Uses the same two-step resolver (resolveSourcePath).
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

		// Two-step resolution: literal first (null-tolerant), then auto-unwrap.
		value, found := resolveSourcePath(collectedData, sourcePath, logger)

		if !found {
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
			zap.Bool("optional", isOptional),
			zap.Bool("value_nil", value == nil))
	}

	return result, nil
}

// GetValueAtExactPath retrieves a value at an exact dot-notation path.
// No fallbacks, no hunting - just the exact path specified.
// Returns (value, true) if found, (nil, false) if not found.
// Note: returns (nil, true) for paths that exist but resolve to nil/null —
// that's distinct from "not found". Used by resolveSourcePath as Step 1.
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

	// spec is the documented container for handler inputs (input_data.spec.*).
	// A required field may be delivered here rather than flattened top-level.
	spec, _ := data["spec"].(map[string]interface{})

	var missing []string
	var satisfiedViaSpec []string

	for _, required := range contract.Required {
		// 1) top-level (the historical check)
		if _, exists := data[required]; exists {
			continue
		}
		// 2) input_data.spec.* — the path handlers actually read (doc 003).
		if spec != nil {
			if _, inSpec := spec[required]; inSpec {
				satisfiedViaSpec = append(satisfiedViaSpec, required)
				continue
			}
		}
		missing = append(missing, required)
	}

	if len(missing) > 0 {
		providedFields := MapKeys(data)
		var specKeys []string
		if spec != nil {
			specKeys = MapKeys(spec)
		}
		return fmt.Errorf(
			"contract violation for agent '%s': missing required fields: %v\n"+
				"Provided fields: %v\n"+
				"Provided spec.* fields: %v\n"+
				"Hint: required fields must be present top-level or under input_data.spec.*; check input_mapping in the step config",
			agentType, missing, providedFields, specKeys,
		)
	}

	if len(satisfiedViaSpec) > 0 {
		logger.Info("Input contract satisfied (some fields via input_data.spec.*)",
			zap.String("agent_type", agentType),
			zap.Strings("via_spec", satisfiedViaSpec))
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
