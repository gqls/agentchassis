// FILE: platform/orchestration/actions/conditional_branch_action.go
// Updated ConditionalBranchAction with support for:
// - String expressions with AND/OR operators
// - Dotted path resolution (e.g., "site_plan.validated_plan.needs_logo")
// - Backward compatibility with object-style conditions

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ConditionalBranchAction provides if/then/else logic in workflows
// Supports both object-style conditions and string expressions with AND/OR
//
// String expression format:
//
//	"field.path == value"
//	"field.path != value"
//	"field.path == true AND other.field == false"
//	"field.path == true OR other.field == true"
//
// Object format (legacy):
//
//	{"field": "path.to.field", "operator": "equals", "value": true}
func ConditionalBranchAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	// Get then/else steps
	thenStep, _ := config["then_step"].(string)
	elseStep, _ := config["else_step"].(string)

	var conditionMet bool
	var err error
	var conditionDebug string

	// Check if condition is a string (expression) or map (object)
	switch cond := config["condition"].(type) {
	case string:
		conditionDebug = cond
		// Parse string expression like "site_plan.needs_logo == true OR site_plan.needs_images == true"
		conditionMet, err = evaluateStringCondition(cond, params.CollectedData, params.Logger)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate condition '%s': %w", cond, err)
		}

	case map[string]interface{}:
		conditionDebug = fmt.Sprintf("%v", cond)
		// Legacy object-style condition
		conditionMet, err = evaluateObjectCondition(cond, params.CollectedData, params.Logger)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate condition: %w", err)
		}

	default:
		return nil, fmt.Errorf("condition is required for conditional_branch (got %T)", config["condition"])
	}

	// Determine next step based on condition
	var nextStep string
	if conditionMet {
		nextStep = thenStep
	} else {
		nextStep = elseStep
	}

	params.Logger.Info("Conditional branch evaluated",
		zap.String("condition", conditionDebug),
		zap.Bool("condition_met", conditionMet),
		zap.String("next_step", nextStep),
	)

	return map[string]interface{}{
		"condition_met":      conditionMet,
		"next_step_override": nextStep,
	}, nil
}

// evaluateStringCondition parses and evaluates string conditions like:
// - "site_plan.needs_logo == true"
// - "site_plan.needs_logo == true OR site_plan.needs_images == true"
// - "site_plan.needs_logo == true AND site_plan.needs_images == true"
//
// Precedence: AND binds tighter than OR (standard boolean precedence)
func evaluateStringCondition(expr string, data map[string]interface{}, logger *zap.Logger) (bool, error) {
	expr = strings.TrimSpace(expr)

	logger.Debug("Evaluating string condition", zap.String("expression", expr))

	// Handle OR (lower precedence - split and evaluate first)
	// We need to be careful to only split on " OR " (with spaces) to avoid matching inside field names
	if idx := strings.Index(expr, " OR "); idx >= 0 {
		parts := splitOnOperator(expr, " OR ")
		logger.Debug("Evaluating OR expression", zap.Int("parts", len(parts)))

		for i, part := range parts {
			result, err := evaluateStringCondition(strings.TrimSpace(part), data, logger)
			if err != nil {
				return false, fmt.Errorf("OR clause %d: %w", i, err)
			}
			if result {
				logger.Debug("OR short-circuit: true", zap.Int("at_part", i))
				return true, nil // Short-circuit OR
			}
		}
		return false, nil
	}

	// Handle AND (higher precedence)
	if idx := strings.Index(expr, " AND "); idx >= 0 {
		parts := splitOnOperator(expr, " AND ")
		logger.Debug("Evaluating AND expression", zap.Int("parts", len(parts)))

		for i, part := range parts {
			result, err := evaluateStringCondition(strings.TrimSpace(part), data, logger)
			if err != nil {
				return false, fmt.Errorf("AND clause %d: %w", i, err)
			}
			if !result {
				logger.Debug("AND short-circuit: false", zap.Int("at_part", i))
				return false, nil // Short-circuit AND
			}
		}
		return true, nil
	}

	// No AND/OR - evaluate single comparison
	return evaluateSingleComparison(expr, data, logger)
}

// splitOnOperator splits a string on an operator, but only at the top level
// (not inside nested expressions if we ever support parentheses)
func splitOnOperator(expr string, operator string) []string {
	// Simple split for now - can be enhanced to handle parentheses later
	return strings.Split(expr, operator)
}

// evaluateSingleComparison evaluates a single comparison expression
// Supports: ==, !=
func evaluateSingleComparison(expr string, data map[string]interface{}, logger *zap.Logger) (bool, error) {
	expr = strings.TrimSpace(expr)

	// Try != first (must check before == since == is a substring)
	if idx := strings.Index(expr, " != "); idx >= 0 {
		field := strings.TrimSpace(expr[:idx])
		expected := strings.TrimSpace(expr[idx+4:])

		actual := resolveFieldValue(field, data, logger)
		matches := compareValues(actual, expected, logger)

		logger.Debug("Evaluated != comparison",
			zap.String("field", field),
			zap.String("expected", expected),
			zap.Any("actual", actual),
			zap.Bool("matches", !matches),
		)

		return !matches, nil
	}

	// Try ==
	if idx := strings.Index(expr, " == "); idx >= 0 {
		field := strings.TrimSpace(expr[:idx])
		expected := strings.TrimSpace(expr[idx+4:])

		actual := resolveFieldValue(field, data, logger)
		matches := compareValues(actual, expected, logger)

		logger.Debug("Evaluated == comparison",
			zap.String("field", field),
			zap.String("expected", expected),
			zap.Any("actual", actual),
			zap.Bool("matches", matches),
		)

		return matches, nil
	}

	// No comparison operator found - treat as truthy check
	// e.g., "site_plan.needs_logo" is true if the field exists and is truthy
	actual := resolveFieldValue(expr, data, logger)
	isTruthy := valueIsTruthy(actual)

	logger.Debug("Evaluated truthy check",
		zap.String("field", expr),
		zap.Any("value", actual),
		zap.Bool("is_truthy", isTruthy),
	)

	return isTruthy, nil
}

// resolveFieldValue gets a value from collected data using a dotted path
// Uses datahelpers.FindByPath for traversal with fallbacks
func resolveFieldValue(fieldPath string, data map[string]interface{}, logger *zap.Logger) interface{} {
	logger.Debug("Resolving field value", zap.String("path", fieldPath))

	// Strategy 1: Use FindByPath from datahelpers (handles unwrapping)
	if value := datahelpers.FindByPath(data, fieldPath, logger); value != nil {
		logger.Debug("Found via FindByPath", zap.String("path", fieldPath))
		return value
	}

	// Strategy 2: Manual traversal for simple cases
	value := traverseDottedPath(fieldPath, data, logger)
	if value != nil {
		logger.Debug("Found via manual traversal", zap.String("path", fieldPath))
		return value
	}

	// Strategy 3: Try searching recursively for the last segment
	// e.g., if "site_plan.needs_logo" fails, try finding "needs_logo" within "site_plan"
	parts := strings.Split(fieldPath, ".")
	if len(parts) >= 2 {
		// Get the base object
		basePath := strings.Join(parts[:len(parts)-1], ".")
		baseObj := traverseDottedPath(basePath, data, logger)

		if baseMap, ok := baseObj.(map[string]interface{}); ok {
			lastKey := parts[len(parts)-1]
			// Search recursively within the base object
			if found := findKeyRecursive(baseMap, lastKey, 0, logger); found != nil {
				logger.Debug("Found via recursive search within base",
					zap.String("base", basePath),
					zap.String("key", lastKey),
				)
				return found
			}
		}
	}

	logger.Debug("Field not found", zap.String("path", fieldPath))
	return nil
}

// traverseDottedPath manually walks a dotted path through nested maps
func traverseDottedPath(path string, data map[string]interface{}, logger *zap.Logger) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil
			}
			current = val
		default:
			return nil
		}
	}

	return current
}

// findKeyRecursive searches for a key anywhere within a nested structure
func findKeyRecursive(data map[string]interface{}, key string, depth int, logger *zap.Logger) interface{} {
	if depth > 10 {
		return nil
	}

	// Direct match
	if val, ok := data[key]; ok {
		return val
	}

	// Recurse into nested maps
	for _, val := range data {
		if nestedMap, ok := val.(map[string]interface{}); ok {
			if found := findKeyRecursive(nestedMap, key, depth+1, logger); found != nil {
				return found
			}
		}
	}

	return nil
}

// compareValues compares an actual value against an expected string representation
func compareValues(actual interface{}, expected string, logger *zap.Logger) bool {
	// Handle null/nil
	if actual == nil {
		return expected == "null" || expected == "nil" || expected == ""
	}

	// Handle boolean "true"
	if expected == "true" {
		switch v := actual.(type) {
		case bool:
			return v == true
		case string:
			return strings.ToLower(v) == "true"
		case int, int64, float64:
			return v != 0
		}
		return false
	}

	// Handle boolean "false"
	if expected == "false" {
		switch v := actual.(type) {
		case bool:
			return v == false
		case string:
			return strings.ToLower(v) == "false"
		case int, int64:
			return v == 0
		case float64:
			return v == 0.0
		}
		return false
	}

	// Handle null check
	if expected == "null" || expected == "nil" {
		return actual == nil
	}

	// Handle quoted strings - remove quotes for comparison
	if (strings.HasPrefix(expected, "\"") && strings.HasSuffix(expected, "\"")) ||
		(strings.HasPrefix(expected, "'") && strings.HasSuffix(expected, "'")) {
		expected = expected[1 : len(expected)-1]
	}

	// String comparison (convert actual to string)
	actualStr := fmt.Sprintf("%v", actual)
	return actualStr == expected
}

// valueIsTruthy checks if a value is "truthy" (exists and is not false/empty/zero)
func valueIsTruthy(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != "" && strings.ToLower(v) != "false"
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0.0
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	default:
		return true // Non-nil unknown types are truthy
	}
}

// evaluateObjectCondition handles the legacy object-style condition format
// Format: {"field": "path.to.field", "operator": "equals", "value": true}
func evaluateObjectCondition(condition map[string]interface{}, data map[string]interface{}, logger *zap.Logger) (bool, error) {
	field, _ := condition["field"].(string)
	operator, _ := condition["operator"].(string)
	expectedValue := condition["value"]

	if field == "" {
		return false, fmt.Errorf("field is required in condition object")
	}

	// Default operator
	if operator == "" {
		operator = "equals"
	}

	// Get the field value using dotted path support
	fieldValue := resolveFieldValue(field, data, logger)

	logger.Debug("Evaluating object condition",
		zap.String("field", field),
		zap.String("operator", operator),
		zap.Any("expected", expectedValue),
		zap.Any("actual", fieldValue),
	)

	switch operator {
	case "equals", "==", "eq":
		return compareValues(fieldValue, fmt.Sprintf("%v", expectedValue), logger), nil

	case "not_equals", "!=", "ne", "neq":
		return !compareValues(fieldValue, fmt.Sprintf("%v", expectedValue), logger), nil

	case "exists":
		return fieldValue != nil, nil

	case "not_exists":
		return fieldValue == nil, nil

	case "contains":
		// Check if field value contains the expected value (for strings/arrays)
		return checkContains(fieldValue, expectedValue), nil

	case "greater_than", "gt", ">":
		return compareNumeric(fieldValue, expectedValue, ">"), nil

	case "less_than", "lt", "<":
		return compareNumeric(fieldValue, expectedValue, "<"), nil

	case "greater_equal", "gte", ">=":
		return compareNumeric(fieldValue, expectedValue, ">="), nil

	case "less_equal", "lte", "<=":
		return compareNumeric(fieldValue, expectedValue, "<="), nil

	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// checkContains checks if actual contains expected
func checkContains(actual, expected interface{}) bool {
	// String contains
	if actualStr, ok := actual.(string); ok {
		if expectedStr, ok := expected.(string); ok {
			return strings.Contains(actualStr, expectedStr)
		}
	}

	// Array contains
	if actualArr, ok := actual.([]interface{}); ok {
		for _, item := range actualArr {
			if fmt.Sprintf("%v", item) == fmt.Sprintf("%v", expected) {
				return true
			}
		}
	}

	return false
}

// compareNumeric performs numeric comparison
func compareNumeric(actual, expected interface{}, op string) bool {
	actualNum, _ := datahelpers.ToFloat64(actual)
	expectedNum, _ := datahelpers.ToFloat64(expected)

	switch op {
	case ">":
		return actualNum > expectedNum
	case "<":
		return actualNum < expectedNum
	case ">=":
		return actualNum >= expectedNum
	case "<=":
		return actualNum <= expectedNum
	default:
		return false
	}
}
