// test/unit/helpers/test_uuid.go
package helpers

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// TestUUID generates a UUID that's recognizable as test data
// It replaces the first 8 characters with a test prefix
func TestUUID(prefix string) string {
	// Generate a normal UUID
	id := uuid.New().String()

	// UUIDs have format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	// We can safely replace the first segment with our test identifier

	// Ensure prefix is exactly 8 characters (pad or truncate)
	testPrefix := formatTestPrefix(prefix)

	// Replace first 8 characters with our test prefix
	parts := strings.Split(id, "-")
	parts[0] = testPrefix

	return strings.Join(parts, "-")
}

// TestUUIDWithType generates a test UUID with a specific type identifier
func TestUUIDWithType(testType string) string {
	switch testType {
	case "unit":
		return TestUUID("00000001") // 00000001-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	case "integration":
		return TestUUID("00000002") // 00000002-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	case "e2e":
		return TestUUID("00000003") // 00000003-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	default:
		return TestUUID("00000000") // 00000000-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	}
}

// formatTestPrefix ensures the prefix is exactly 8 hex characters
func formatTestPrefix(prefix string) string {
	// Remove any non-hex characters and convert to lowercase
	cleaned := strings.ToLower(prefix)

	// Common test prefixes mapped to hex
	switch cleaned {
	case "test":
		return "7e570000" // "test" in a recognizable hex pattern
	case "unit":
		return "00000001"
	case "intg":
		return "00000002"
	case "e2e":
		return "00000003"
	case "tmp":
		return "00000099"
	default:
		// Create a hex representation that's obviously test data
		// Use 0000 prefix followed by a hash of the input
		hash := fmt.Sprintf("%x", hashString(prefix))
		if len(hash) >= 4 {
			return "0000" + hash[:4]
		}
		return "00000000"
	}
}

// Simple hash function for consistent prefix generation
func hashString(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}

// IsTestUUID checks if a UUID is a test UUID
func IsTestUUID(id string) bool {
	if len(id) < 8 {
		return false
	}

	// Check for known test prefixes
	prefix := id[:8]
	testPrefixes := []string{
		"00000000", // general test
		"00000001", // unit test
		"00000002", // integration test
		"00000003", // e2e test
		"00000099", // temp test
		"7e570000", // "test" pattern
	}

	for _, tp := range testPrefixes {
		if prefix == tp {
			return true
		}
	}

	// Also check if it starts with multiple zeros (likely test data)
	return strings.HasPrefix(id, "0000")
}
