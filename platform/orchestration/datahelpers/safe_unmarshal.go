// FILE: platform/orchestration/datahelpers/safe_unmarshal.go
// Add this to your datahelpers package

package datahelpers

import (
	"encoding/json"

	"go.uber.org/zap"
)

// SafeUnmarshal attempts to unmarshal JSON data into dest without returning an error.
// If unmarshaling fails, dest is left unchanged and the function returns false.
// This is useful when you want to attempt parsing optional JSON fields without
// failing the entire operation.
func SafeUnmarshal(data []byte, dest interface{}) bool {
	if len(data) == 0 {
		return false
	}
	err := json.Unmarshal(data, dest)
	return err == nil
}

// SafeUnmarshalWithLog attempts to unmarshal JSON data into dest, logging any errors.
// Returns true if successful, false otherwise.
func SafeUnmarshalWithLog(data []byte, dest interface{}, logger *zap.Logger, fieldName string) bool {
	if len(data) == 0 {
		return false
	}
	err := json.Unmarshal(data, dest)
	if err != nil {
		if logger != nil {
			logger.Debug("SafeUnmarshal failed",
				zap.String("field", fieldName),
				zap.Error(err),
				zap.Int("data_length", len(data)),
			)
		}
		return false
	}
	return true
}

// SafeUnmarshalString attempts to unmarshal a JSON string into dest.
// Convenience wrapper for SafeUnmarshal that accepts string input.
func SafeUnmarshalString(data string, dest interface{}) bool {
	return SafeUnmarshal([]byte(data), dest)
}
