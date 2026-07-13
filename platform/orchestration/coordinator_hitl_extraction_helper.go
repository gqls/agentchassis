// FILE: platform/orchestration/coordinator_hitl_extraction.go
// NEW FILE: HITL response extraction helpers
// Add this as a new file, then import it in coordinator.go

package orchestration

import (
	"fmt"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// extractHITLFormFields extracts only form field values from HITL response
// This prevents storing review/presentation data instead of actual form submissions
//
// Example:
//
//	Response: {"site_type": "brochure", "recommended_builder": "multipage-website-builder", "success": true}
//	Config: {"fields": [{"name": "site_type"}, {"name": "recommended_builder"}]}
//	Returns: {"site_type": "brochure", "recommended_builder": "multipage-website-builder"}
func extractHITLFormFields(
	responseData map[string]interface{},
	stepConfig map[string]interface{},
	logger *zap.Logger,
) map[string]interface{} {

	result := make(map[string]interface{})

	logger.Debug("Extracting HITL form fields",
		zap.Strings("response_keys", getMapKeys(responseData)),
	)

	// Get field definitions from step config
	fieldsConfig, ok := stepConfig["fields"]
	if !ok {
		logger.Debug("No 'fields' config found in HITL step, returning filtered response")
		// No fields defined, return entire response minus metadata
		for k, v := range responseData {
			if !isMetadataField(k) {
				result[k] = v
			}
		}
		return result
	}

	// Parse fields array
	fieldsArray, ok := fieldsConfig.([]interface{})
	if !ok {
		logger.Warn("Fields config is not an array",
			zap.String("type", fmt.Sprintf("%T", fieldsConfig)))
		// Fallback to filtered response
		for k, v := range responseData {
			if !isMetadataField(k) {
				result[k] = v
			}
		}
		return result
	}

	// Extract each defined form field from response
	extractedCount := 0
	for _, fieldDef := range fieldsArray {
		fieldMap, ok := fieldDef.(map[string]interface{})
		if !ok {
			continue
		}

		fieldName, ok := fieldMap["name"].(string)
		if !ok || fieldName == "" {
			continue
		}

		// Look for this field in the response
		if value, exists := responseData[fieldName]; exists {
			result[fieldName] = value
			extractedCount++
			logger.Debug("Extracted HITL form field",
				zap.String("field_name", fieldName),
				zap.String("value_type", fmt.Sprintf("%T", value)),
			)
		} else {
			logger.Debug("HITL form field not in response (may use default)",
				zap.String("field_name", fieldName),
			)
		}
	}

	logger.Info("HITL form field extraction complete",
		zap.Int("fields_defined", len(fieldsArray)),
		zap.Int("fields_extracted", extractedCount),
		zap.Strings("extracted_fields", getMapKeys(result)),
	)

	return result
}

// isMetadataField checks if a field name is metadata (not actual form data)
// These fields should not be stored in the output_field
func isMetadataField(fieldName string) bool {
	metadataFields := map[string]bool{
		// Standard HITL metadata
		"success":        true,
		"status":         true,
		"message":        true,
		"skipped":        true,
		"skip_reason":    true,
		"human_response": true,
		"timestamp":      true,
		"approved_by":    true,

		// Workflow control
		"await_response": true,
		"request_id":     true,
		"reply_to_topic": true,

		// Approval metadata
		"approved":      true,
		"comments":      true,
		"approval_type": true,
	}
	return metadataFields[fieldName]
}

// shouldExtractFormFields determines if this step needs HITL field extraction
func shouldExtractFormFields(step models.Step) bool {
	// List of actions that need form field extraction
	extractionActions := map[string]bool{
		"request_human_input": true,
		"await_approval":      true,
		// Add others if needed
	}
	return extractionActions[step.Action]
}
