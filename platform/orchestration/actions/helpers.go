package actions

// ExtractStepData checks if stepData contains a response field and extracts it.
// This is used throughout actions to handle both direct data and step responses.
func ExtractStepData(stepData interface{}) interface{} {
	if stepMap, ok := stepData.(map[string]interface{}); ok {
		// Check if this step has a response field
		if response, hasResponse := stepMap["response"]; hasResponse {
			return response
		}
		// No response field, return the step data itself
		return stepMap
	}
	// Not a map, return as-is
	return stepData
}

// GetMapKeys returns the keys of a map as a string slice
func GetMapKeys(m map[string]interface{}) []string {
	if m == nil {
		return []string{}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
