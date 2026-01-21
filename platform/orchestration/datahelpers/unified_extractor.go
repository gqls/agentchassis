// FILE: platform/orchestration/datahelpers/unified_extractor.go
// THE MASTER EXTRACTOR - Uses all our helper functions together

package datahelpers

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ExtractFields is THE ONLY function you should use to extract fields
// It uses ALL our existing helpers in the right order
// When "input_data" is in fieldNames, this function:
//  1. Extracts core fields (domain, objective, model) to ROOT level -> {{.domain}} works
//  2. ALSO creates result["input_data"] map with same data -> {{.input_data.domain}} works
//
// This ensures backwards compatibility while supporting both template patterns.
func ExtractFields(
	collectedData map[string]interface{},
	fieldNames []string,
	logger *zap.Logger,
) map[string]interface{} {

	logger.Info("=== MASTER EXTRACTOR START ===",
		zap.Strings("requested_fields", fieldNames),
		zap.Strings("available_keys", getMapKeys(collectedData)),
	)

	result := make(map[string]interface{})

	// Special handling for "input_data" field
	if contains(fieldNames, "input_data") {
		logger.Info("Special case: extracting input_data (dual access pattern)")

		// Use aggressive search to get core data
		coreData := ExtractCoreInputData(collectedData, logger)

		// Create the input_data map for {{.input_data.domain}} access
		inputDataMap := make(map[string]interface{})

		// Populate both root level AND input_data map
		for k, v := range coreData {
			result[k] = v       // Root level: {{.domain}}
			inputDataMap[k] = v // Nested: {{.input_data.domain}}
			logger.Info("Added to both root and input_data",
				zap.String("key", k),
				zap.String("type", fmt.Sprintf("%T", v)),
			)
		}

		// Also try to get input_data map directly for any additional fields
		if existingInputMap := getInputDataMap(collectedData, logger); existingInputMap != nil {
			for k, v := range existingInputMap {
				if _, exists := result[k]; !exists {
					result[k] = v       // Root level
					inputDataMap[k] = v // Nested
					logger.Info("Added additional field from input_data map",
						zap.String("key", k),
					)
				}
			}
		}

		// Store the input_data map so {{.input_data.domain}} works
		result["input_data"] = inputDataMap
		logger.Info("Created input_data map for nested access",
			zap.Strings("input_data_keys", getMapKeys(inputDataMap)),
		)
	}

	// ========================================================================
	// Special handling for "reviewed_brief" - often deeply nested in input_data
	// This field is commonly needed for site planning and content generation
	// ========================================================================
	if contains(fieldNames, "reviewed_brief") {
		logger.Info("Special case: extracting reviewed_brief (aggressive recursive search)")

		// Use findFieldRecursive which searches through all nested structures
		if reviewedBrief := findFieldRecursive(collectedData, "reviewed_brief", 0, logger); reviewedBrief != nil {
			result["reviewed_brief"] = reviewedBrief

			// Log what we found for debugging
			if rbMap, ok := reviewedBrief.(map[string]interface{}); ok {
				logger.Info("✓ Found reviewed_brief",
					zap.Strings("keys", getMapKeys(rbMap)),
					zap.Bool("has_company_name", rbMap["company_name"] != nil),
					zap.Bool("has_about_us", rbMap["about_us"] != nil),
					zap.Bool("has_services", rbMap["services"] != nil),
				)
			} else {
				logger.Info("✓ Found reviewed_brief (non-map type)",
					zap.String("type", fmt.Sprintf("%T", reviewedBrief)))
			}
		} else {
			logger.Error("✗ Could not find reviewed_brief anywhere in collected data",
				zap.Strings("top_level_keys", getMapKeys(collectedData)))
		}
	}

	// ========================================================================
	// Special handling for "site_record" - also commonly needed
	// ========================================================================
	if contains(fieldNames, "site_record") {
		logger.Info("Special case: extracting site_record")

		if siteRecord := findFieldRecursive(collectedData, "site_record", 0, logger); siteRecord != nil {
			result["site_record"] = siteRecord
			logger.Info("✓ Found site_record")
		} else {
			logger.Warn("✗ Could not find site_record")
		}
	}

	// ========================================================================
	// Special handling for "current_page" - needed for content generation
	// ========================================================================
	if contains(fieldNames, "current_page") {
		logger.Info("Special case: extracting current_page")

		if currentPage := findFieldRecursive(collectedData, "current_page", 0, logger); currentPage != nil {
			result["current_page"] = currentPage
			if cpMap, ok := currentPage.(map[string]interface{}); ok {
				logger.Info("✓ Found current_page",
					zap.String("name", fmt.Sprintf("%v", cpMap["name"])),
					zap.String("title", fmt.Sprintf("%v", cpMap["title"])),
				)
			} else {
				logger.Info("✓ Found current_page (non-map type)")
			}
		} else {
			logger.Warn("✗ Could not find current_page")
		}
	}

	// ========================================================================
	// Track which fields we've already handled specially
	// ========================================================================
	speciallyHandled := map[string]bool{
		"input_data":     true,
		"reviewed_brief": true,
		"site_record":    true,
		"current_page":   true,
	}

	// Extract each specific field (skip already handled)
	for _, fieldName := range fieldNames {
		if speciallyHandled[fieldName] {
			continue // Already handled above
		}

		logger.Info(">>> Extracting field", zap.String("field", fieldName))

		value := extractSingleField(collectedData, fieldName, make(map[string]bool), logger)

		if value != nil {
			// Store with simple name (last part of path)
			parts := strings.Split(fieldName, ".")
			simpleKey := parts[len(parts)-1]
			result[simpleKey] = value

			logger.Info("✓ Field extracted",
				zap.String("requested", fieldName),
				zap.String("stored_as", simpleKey),
				zap.String("type", fmt.Sprintf("%T", value)),
			)
		} else {
			logger.Warn("✗ Field not found",
				zap.String("field", fieldName),
			)
		}
	}

	// Always ensure domain and objective exist at root level
	ensureCoreFields(result, collectedData, logger)

	// Also ensure they exist inside input_data if that map exists
	if inputDataMap, ok := result["input_data"].(map[string]interface{}); ok {
		syncCoreFieldsToInputData(result, inputDataMap, logger)
	}

	logger.Info("=== MASTER EXTRACTOR COMPLETE ===",
		zap.Int("fields_extracted", len(result)),
		zap.Strings("result_keys", getMapKeys(result)),
	)

	return result
}

// syncCoreFieldsToInputData ensures domain/objective/model exist in input_data map
// if they were recovered at root level by ensureCoreFields
func syncCoreFieldsToInputData(root map[string]interface{}, inputData map[string]interface{}, logger *zap.Logger) {
	coreFields := []string{"domain", "objective", "model"}

	for _, field := range coreFields {
		if _, existsInInputData := inputData[field]; !existsInInputData {
			if value, existsAtRoot := root[field]; existsAtRoot {
				inputData[field] = value
				logger.Info("Synced core field to input_data",
					zap.String("field", field),
				)
			}
		}
	}
}

// extractSingleField tries multiple strategies to find ONE field
func extractSingleField(
	data map[string]interface{},
	fieldName string,
	seen map[string]bool,
	logger *zap.Logger,
) interface{} {

	logger.Info("Trying extraction strategies", zap.String("field", fieldName))

	// Prevent infinite loops from circular aliases
	if seen[fieldName] {
		return nil // Already tried this field
	}
	seen[fieldName] = true

	// Strategy 1: Use FindByPath (handles unwrapping)
	if value := FindByPath(data, fieldName, logger); value != nil {
		logger.Info("Found via FindByPath", zap.String("field", fieldName))
		return value
	}

	// Strategy 2: Try with input_data prefix
	if !strings.HasPrefix(fieldName, "input_data.") {
		path := "input_data." + fieldName
		if value := FindByPath(data, path, logger); value != nil {
			logger.Info("Found via input_data prefix",
				zap.String("field", fieldName),
				zap.String("path", path))
			return value
		}
	}

	// Strategy 3: Look inside input_data map directly
	if inputMap := getInputDataMap(data, logger); inputMap != nil {
		if value, ok := inputMap[fieldName]; ok {
			logger.Info("Found in input_data map", zap.String("field", fieldName))
			return UnwrapDeep(value, logger)
		}
	}

	// Strategy 4: Aggressive recursive search
	logger.Info("Trying aggressive search", zap.String("field", fieldName))
	if value := findFieldRecursive(data, fieldName, 0, logger); value != nil {
		logger.Info("Found via aggressive search", zap.String("field", fieldName))
		return value
	}

	// Strategy 5: Check known aliases
	if alias := getFieldAlias(fieldName); alias != "" {
		logger.Info("Trying field alias",
			zap.String("field", fieldName),
			zap.String("alias", alias))
		return extractSingleField(data, alias, seen, logger)
	}

	logger.Info("Field name not found using aggressive search in extractSingleField.",
		zap.String("field", fieldName),
	)

	return nil
}

// findFieldRecursive does aggressive recursive search
func findFieldRecursive(
	data interface{},
	fieldName string,
	depth int,
	logger *zap.Logger,
) interface{} {
	if depth > 20 {
		return nil
	}

	// Handle maps
	if m, ok := data.(map[string]interface{}); ok {
		// Direct match
		if val, ok := m[fieldName]; ok {
			logger.Debug("Recursive: found direct match",
				zap.String("field", fieldName),
				zap.Int("depth", depth),
			)
			return UnwrapDeep(val, logger)
		}

		// Try unwrapping first
		unwrapped := tryUnwrapMapPatterns(m, logger)
		if unwrapped != nil {
			if result := findFieldRecursive(unwrapped, fieldName, depth+1, logger); result != nil {
				return result
			}
		}

		// Recurse into all values
		for key, val := range m {
			if result := findFieldRecursive(val, fieldName, depth+1, logger); result != nil {
				logger.Debug("Recursive: found through key",
					zap.String("through_key", key),
					zap.Int("depth", depth),
				)
				return result
			}
		}
	}

	// Handle slices
	if slice, ok := data.([]interface{}); ok {
		for _, val := range slice {
			if result := findFieldRecursive(val, fieldName, depth+1, logger); result != nil {
				return result
			}
		}
	}

	return nil
}

// tryUnwrapMapPatterns tries to unwrap common nesting patterns
func tryUnwrapMapPatterns(m map[string]interface{}, logger *zap.Logger) interface{} {
	// Pattern 1: {field}_result.result
	for key, val := range m {
		if strings.HasSuffix(key, "_result") {
			if resultMap, ok := val.(map[string]interface{}); ok {
				if result, ok := resultMap["result"]; ok {
					if parsed := tryParseJSON(result, logger); parsed != nil {
						return parsed
					}
					return result
				}
			}
		}
	}

	// Pattern 2: Direct result field
	if result, ok := m["result"]; ok {
		if parsed := tryParseJSON(result, logger); parsed != nil {
			return parsed
		}
		return result
	}

	// Pattern 3: input_data field
	if inputData, ok := m["input_data"]; ok {
		return inputData
	}

	return nil
}

// ensureCoreFields makes absolutely sure critical fields exist eg domain, objective
// These are fields that templates commonly reference but might not be explicitly requested
func ensureCoreFields(
	result map[string]interface{},
	source map[string]interface{},
	logger *zap.Logger,
) {
	logger.Info("Ensuring core fields present")

	// Check domain
	if _, hasDomain := result["domain"]; !hasDomain {
		logger.Warn("Domain missing from result, searching aggressively")
		if domain := FindDomainAggressive(source, logger); domain != "" {
			result["domain"] = domain
			logger.Info("✓ Recovered domain via aggressive search", zap.String("domain", domain))
		} else {
			// Try site_record.domain explicitly
			if siteRecord := findFieldRecursive(source, "site_record", 0, logger); siteRecord != nil {
				if srMap, ok := siteRecord.(map[string]interface{}); ok {
					if d, ok := srMap["domain"].(string); ok && d != "" {
						result["domain"] = d
						logger.Info("✓ Found domain in site_record.domain", zap.String("domain", d))
					}
				}
			}
		}
	}

	// Check objective
	if _, hasObjective := result["objective"]; !hasObjective {
		logger.Warn("Objective missing from result, searching aggressively")
		if objective := FindObjectiveAggressive(source, logger); objective != "" {
			result["objective"] = objective
			logger.Info("✓ Recovered objective via aggressive search",
				zap.Int("length", len(objective)))
		} else {
			// Build objective from context if we can't find one
			objective := buildObjectiveFromContext(result, source, logger)
			if objective != "" {
				result["objective"] = objective
				logger.Info("✓ Built objective from context", zap.String("objective", objective))
			} else {
				logger.Warn("✗ Could not find or build objective")
			}
		}
	}

	// Check model
	if _, hasModel := result["model"]; !hasModel {
		if model := findFieldRecursive(source, "model", 0, logger); model != nil {
			if modelStr, ok := model.(string); ok {
				result["model"] = modelStr
				logger.Info("✓ Recovered model via aggressive search", zap.String("model", modelStr))
			}
		}
	}

	// ADDED: Check current_page - commonly referenced in content generation templates
	if _, hasCurrentPage := result["current_page"]; !hasCurrentPage {
		if currentPage := findFieldRecursive(source, "current_page", 0, logger); currentPage != nil {
			result["current_page"] = currentPage
			if cpMap, ok := currentPage.(map[string]interface{}); ok {
				logger.Info("✓ Auto-recovered current_page",
					zap.String("name", fmt.Sprintf("%v", cpMap["name"])),
					zap.String("title", fmt.Sprintf("%v", cpMap["title"])),
				)
			} else {
				logger.Info("✓ Auto-recovered current_page (non-map)")
			}
		}
	}

	// ADDED: Check current_section - critical for loop iterations
	if _, hasCurrentSection := result["current_section"]; !hasCurrentSection {
		if currentSection := findFieldRecursive(source, "current_section", 0, logger); currentSection != nil {
			result["current_section"] = currentSection
			logger.Info("✓ Auto-recovered current_section")
		}
	}

	// ADDED: Check render_context - commonly needed for template rendering
	if _, hasRenderContext := result["render_context"]; !hasRenderContext {
		if renderContext := findFieldRecursive(source, "render_context", 0, logger); renderContext != nil {
			result["render_context"] = renderContext
			logger.Info("✓ Auto-recovered render_context")
		}
	}
}

// buildObjectiveFromContext constructs an objective string from available context
// Used when no explicit objective is found but we have enough context to build one
func buildObjectiveFromContext(result, source map[string]interface{}, logger *zap.Logger) string {
	var parts []string

	// Get section info
	sectionName := ""
	if cs, ok := result["current_section"].(map[string]interface{}); ok {
		if fn, ok := cs["function"].(string); ok && fn != "" {
			sectionName = fn
		} else if n, ok := cs["name"].(string); ok && n != "" {
			sectionName = n
		}
	}

	// Get company name
	companyName := ""
	if rb, ok := result["reviewed_brief"].(map[string]interface{}); ok {
		if cn, ok := rb["company_name"].(string); ok {
			companyName = cn
		}
	}

	// Get domain
	domain := ""
	if d, ok := result["domain"].(string); ok {
		domain = d
	}

	// Build objective based on what we have
	if sectionName != "" {
		parts = append(parts, fmt.Sprintf("Generate content for %s section", sectionName))
	}
	if companyName != "" {
		parts = append(parts, fmt.Sprintf("for %s", companyName))
	}
	if domain != "" && companyName == "" {
		parts = append(parts, fmt.Sprintf("for %s", domain))
	}

	if len(parts) > 0 {
		return strings.Join(parts, " ")
	}

	return ""
}

// getInputDataMap extracts the input_data map if it exists
func getInputDataMap(data map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	// Try direct lookup
	if inputData, ok := data["input_data"].(map[string]interface{}); ok {
		// Check for double nesting
		if nested, ok := inputData["input_data"].(map[string]interface{}); ok {
			logger.Debug("Unwrapped double-nested input_data")
			return nested
		}
		return inputData
	}

	// Try unwrapping data first
	unwrapped := UnwrapDeep(data, logger)
	if unwrappedMap, ok := unwrapped.(map[string]interface{}); ok {
		if inputData, ok := unwrappedMap["input_data"].(map[string]interface{}); ok {
			return inputData
		}
	}

	return nil
}

// getFieldAlias returns alternative names for common fields
// NOTE: Do NOT alias site_content<->content because "content" is too common
// (appears in sections[0].content, etc.) and causes wrong field extraction
func getFieldAlias(fieldName string) string {
	aliases := map[string]string{
		"site_architecture":  "architecture",
		"architecture":       "site_architecture",
		"domain_analysis":    "analysis",
		"analysis":           "domain_analysis",
		"available_builders": "builders",
		// Removed: "site_content": "content" - too dangerous, matches sections[0].content
		// Removed: "content": "site_content"
	}
	return aliases[fieldName]
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
