// FILE: platform/orchestration/datahelpers/action_inputs.go
// Standardized input extraction for all actions
// Replaces repetitive boilerplate in each action

package datahelpers

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ActionInputSpec declares what inputs an action needs
// Used to standardize extraction across all actions
type ActionInputSpec struct {
	// Required fields - action fails if not found
	Required []string

	// Optional fields - extracted if present, no error if missing
	Optional []string

	// Deprecated field mappings: old config key -> new field name
	// e.g., "site_id_field" -> "site_id"
	// Will log deprecation warning when old pattern is used
	Deprecated map[string]string

	// DefaultValues for optional fields
	Defaults map[string]interface{}
}

// ActionInputs is the result of extraction
type ActionInputs struct {
	// All extracted values (required + optional + defaults)
	Values map[string]interface{}

	// Which deprecated patterns were used (for logging)
	DeprecatedUsed []string

	// Which required fields were missing (before error)
	MissingRequired []string
}

// Get retrieves a string value, returns empty string if not found
func (ai *ActionInputs) Get(key string) string {
	if v, ok := ai.Values[key].(string); ok {
		return v
	}
	return ""
}

// GetMap retrieves a map value, returns nil if not found
func (ai *ActionInputs) GetMap(key string) map[string]interface{} {
	if v, ok := ai.Values[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}

// GetInt retrieves an int value, returns default if not found
func (ai *ActionInputs) GetInt(key string, defaultVal int) int {
	if v, ok := ai.Values[key].(float64); ok {
		return int(v)
	}
	if v, ok := ai.Values[key].(int); ok {
		return v
	}
	return defaultVal
}

// GetBool retrieves a bool value, returns default if not found
func (ai *ActionInputs) GetBool(key string, defaultVal bool) bool {
	if v, ok := ai.Values[key].(bool); ok {
		return v
	}
	return defaultVal
}

// Has checks if a key exists and is non-nil
func (ai *ActionInputs) Has(key string) bool {
	v, ok := ai.Values[key]
	return ok && v != nil
}

// GetRaw retrieves the raw interface{} value, returns nil if not found
func (ai *ActionInputs) GetRaw(key string) interface{} {
	return ai.Values[key]
}

// ExtractActionInputs extracts inputs according to spec
// Priority order:
//  1. input_fields from config (preferred pattern)
//  2. Deprecated *_field configs (with warning)
//  3. Direct values in input_data
//
// This centralizes the extraction logic that was previously duplicated
// in every action's boilerplate code.
func ExtractActionInputs(
	collectedData map[string]interface{},
	config map[string]interface{},
	spec ActionInputSpec,
	logger *zap.Logger,
) (*ActionInputs, error) {

	result := &ActionInputs{
		Values:         make(map[string]interface{}),
		DeprecatedUsed: []string{},
	}

	// Apply defaults first
	for k, v := range spec.Defaults {
		result.Values[k] = v
	}

	// Combine all fields we need to look for
	allFields := append([]string{}, spec.Required...)
	allFields = append(allFields, spec.Optional...)

	// Strategy 0: Resolve config values as EXPLICIT path references FIRST.
	// This must run before ExtractFields (Strategy 1/2) because ExtractFields
	// uses aggressive recursive search that can find stale values from
	// previous loop iterations (e.g. claim_result.work_item_id from iter 0).
	// When the config explicitly says "work_item_id": "current_item.id",
	// that explicit path should win over any aggressive search result.
	for _, field := range allFields {
		pathStr, ok := config[field].(string)
		if !ok || pathStr == "" {
			continue
		}
		// Only resolve multi-segment dot-paths (these are unambiguously path references)
		if strings.Contains(pathStr, ".") {
			value := ExtractNestedField(collectedData, pathStr)
			if value != nil {
				result.Values[field] = value
				logger.Info("Strategy 0: Resolved config path before ExtractFields",
					zap.String("field", field),
					zap.String("path", pathStr),
				)
			}
		}
	}

	// Strategy 1: Use input_fields if specified in config (preferred)
	if inputFields, ok := config["input_fields"].([]interface{}); ok {
		fieldNames := make([]string, len(inputFields))
		for i, f := range inputFields {
			fieldNames[i], _ = f.(string)
		}
		extracted := ExtractFields(collectedData, fieldNames, logger)
		for k, v := range extracted {
			// Don't overwrite values already resolved by Strategy 0 (explicit config paths)
			if _, alreadyResolved := result.Values[k]; alreadyResolved {
				continue
			}
			if v != nil {
				result.Values[k] = v
			}
		}
	} else {
		// Strategy 2: Try to extract all needed fields directly
		extracted := ExtractFields(collectedData, allFields, logger)
		for k, v := range extracted {
			// Don't overwrite values already resolved by Strategy 0 (explicit config paths)
			if _, alreadyResolved := result.Values[k]; alreadyResolved {
				continue
			}
			if v != nil {
				result.Values[k] = v
			}
		}
	}

	// Strategy 3: Check deprecated *_field patterns for any missing fields
	for oldKey, newField := range spec.Deprecated {
		// Only use deprecated pattern if we don't already have the value
		if _, hasValue := result.Values[newField]; hasValue {
			continue
		}

		if pathStr, ok := config[oldKey].(string); ok && pathStr != "" {
			value := ExtractNestedField(collectedData, pathStr)
			if value != nil {
				result.Values[newField] = value
				result.DeprecatedUsed = append(result.DeprecatedUsed, oldKey)

				logger.Warn("Using deprecated config pattern",
					zap.String("deprecated_key", oldKey),
					zap.String("path", pathStr),
					zap.String("use_instead", fmt.Sprintf("input_fields: [\"%s\"]", newField)),
				)
			}
		}
	}

	// Check for nested object access (backward compat)
	// e.g., if we need "page_id" but got "current_page" object containing "page_id"
	for _, field := range allFields {
		if _, hasValue := result.Values[field]; hasValue {
			continue
		}

		// Check common nested patterns
		nestedSources := []struct {
			parent string
			child  string
		}{
			{"current_page", field},
			{"rerender_pages", field},
			{"site_record", field},
			{"input_data", field},
		}

		for _, ns := range nestedSources {
			if parent, ok := result.Values[ns.parent].(map[string]interface{}); ok {
				if childVal, exists := parent[ns.child]; exists && childVal != nil {
					result.Values[field] = childVal
					break
				}
			}
		}
	}

	// Strategy 4: Resolve remaining config value references.
	// Dot-path references (e.g. "current_item.id") were already handled
	// by Strategy 0 above. This handles single-segment references
	// (e.g. "spec_data": "site_plan") and any dot-paths that Strategy 0
	// couldn't resolve (data may have been populated by Strategies 1-3).
	for _, field := range allFields {
		if _, hasValue := result.Values[field]; hasValue {
			continue
		}

		pathStr, ok := config[field].(string)
		if !ok || pathStr == "" {
			continue
		}

		// Multi-segment path (has dot): resolve via ExtractNestedField
		if strings.Contains(pathStr, ".") {
			value := ExtractNestedField(collectedData, pathStr)
			if value != nil {
				result.Values[field] = value
				logger.Debug("Resolved config value as dot-path",
					zap.String("field", field),
					zap.String("path", pathStr),
				)
			}
			continue
		}

		// Single-segment: check if it matches a top-level key in collectedData
		// e.g. config has "spec_data": "site_plan" and collectedData has "site_plan": {...}
		if val, exists := collectedData[pathStr]; exists && val != nil {
			result.Values[field] = val
			logger.Debug("Resolved config value as collected_data key",
				zap.String("field", field),
				zap.String("key", pathStr),
			)
		}
	}

	// Strategy 5: numeric and boolean config values, taken as LITERALS.
	//
	// Every strategy above reads step config as config[field].(string) and
	// treats that string as a REFERENCE to resolve against collectedData. A
	// JSON number or boolean fails the type assertion outright, so it never
	// reaches the action and the call site's Go fallback wins instead —
	// silently, while the config reads as though it were live. See
	// bugs_open/042: render_news_json carried max_age_hours 72 and
	// RenderNewsSectionAction defaulted to 72, so config and behaviour agreed
	// and nothing looked wrong until the value was changed to 720 and the
	// render kept behaving like 72. max_items had never been read either.
	//
	// Deliberately restricted to NON-STRING scalars. References are always
	// strings, so this cannot change how any existing reference resolves — it
	// only fills fields that are currently dropped on the floor. A string
	// literal that fails to resolve is left alone on purpose: taking it as its
	// own value would turn a broken reference into a silent literal and mask
	// real wiring bugs. Composite values (objects, arrays) are left alone too,
	// since there is no evidence they were ever intended as literals here.
	for _, field := range allFields {
		if _, hasValue := result.Values[field]; hasValue {
			continue
		}
		raw, exists := config[field]
		if !exists || raw == nil {
			continue
		}
		switch raw.(type) {
		case bool,
			float64, float32,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			json.Number:
			result.Values[field] = raw
			logger.Debug("Strategy 5: took literal scalar config value",
				zap.String("field", field),
			)
		}
	}

	// Validate required fields
	var missing []string
	for _, req := range spec.Required {
		val, exists := result.Values[req]
		if !exists || val == nil || val == "" {
			missing = append(missing, req)
		}
	}

	if len(missing) > 0 {
		result.MissingRequired = missing
		return result, fmt.Errorf("missing required fields: %v", missing)
	}

	// Log deprecation summary if any
	if len(result.DeprecatedUsed) > 0 {
		logger.Info("Action used deprecated config patterns",
			zap.Strings("deprecated_keys", result.DeprecatedUsed),
			zap.String("recommendation", "migrate to input_fields pattern"),
		)
	}

	return result, nil
}

// QuickExtract is a convenience wrapper for simple cases
// Returns map directly, panics on missing required (use in tests or simple actions)
func QuickExtract(
	collectedData map[string]interface{},
	config map[string]interface{},
	required []string,
	optional []string,
	logger *zap.Logger,
) (map[string]interface{}, error) {
	spec := ActionInputSpec{
		Required: required,
		Optional: optional,
	}
	result, err := ExtractActionInputs(collectedData, config, spec, logger)
	if err != nil {
		return nil, err
	}
	return result.Values, nil
}

// BuildDeprecationMap creates a Deprecated map from a list of field names
// Assumes old pattern was fieldName + "_field" suffix
// e.g., ["site_id", "page_id"] -> {"site_id_field": "site_id", "page_id_field": "page_id"}
func BuildDeprecationMap(fields []string) map[string]string {
	m := make(map[string]string)
	for _, f := range fields {
		m[f+"_field"] = f
	}
	return m
}

// ---- Input specification registry ----
// Actions can register their specs for documentation and validation

var actionInputSpecs = make(map[string]ActionInputSpec)

// RegisterActionInputSpec registers an action's input specification
// Used for documentation generation and contract validation
func RegisterActionInputSpec(actionName string, spec ActionInputSpec) {
	actionInputSpecs[actionName] = spec
}

// GetActionInputSpec retrieves a registered spec
func GetActionInputSpec(actionName string) (ActionInputSpec, bool) {
	spec, ok := actionInputSpecs[actionName]
	return spec, ok
}

// ListActionInputSpecNames returns the names of every registered spec.
// Used to check spec/registry parity — an action that declares inputs but has
// no registry entry is invisible to the workflow validator, which then reads it
// as a remote action and rejects the workflow (see registry_parity_test.go).
func ListActionInputSpecNames() []string {
	names := make([]string, 0, len(actionInputSpecs))
	for name := range actionInputSpecs {
		names = append(names, name)
	}
	return names
}

// GenerateInputContract creates an input_contract from a spec
// Can be used to auto-generate contract JSON for agent_definitions
func GenerateInputContract(spec ActionInputSpec) map[string]interface{} {
	contract := map[string]interface{}{
		"required": spec.Required,
	}
	if len(spec.Optional) > 0 {
		contract["optional"] = spec.Optional
	}
	if len(spec.Deprecated) > 0 {
		deprecated := make([]string, 0, len(spec.Deprecated))
		for k := range spec.Deprecated {
			deprecated = append(deprecated, k)
		}
		contract["deprecated"] = deprecated
	}
	return contract
}

// ---- Helper to migrate actions ----

// MigrationHelper logs what an action is receiving and suggests migration
// Use temporarily when converting actions to new pattern
func MigrationHelper(actionName string, collectedData, config map[string]interface{}, logger *zap.Logger) {
	logger.Info("=== Migration Helper ===",
		zap.String("action", actionName),
	)

	// Log what's in input_data
	if inputData, ok := collectedData["input_data"].(map[string]interface{}); ok {
		keys := make([]string, 0, len(inputData))
		for k := range inputData {
			keys = append(keys, k)
		}
		logger.Info("Available in input_data", zap.Strings("keys", keys))
	}

	// Log what's in config
	configKeys := make([]string, 0, len(config))
	for k := range config {
		configKeys = append(configKeys, k)
	}
	logger.Info("Config keys", zap.Strings("keys", configKeys))

	// Check for *_field patterns
	var fieldPatterns []string
	for k := range config {
		if strings.HasSuffix(k, "_field") {
			fieldPatterns = append(fieldPatterns, k)
		}
	}
	if len(fieldPatterns) > 0 {
		logger.Warn("Found deprecated *_field patterns - migrate to input_fields",
			zap.Strings("patterns", fieldPatterns),
		)
	}
}
