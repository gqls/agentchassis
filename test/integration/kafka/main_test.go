// test/integration/kafka/main_test.go
package kafka

import (
	"os"
	"testing"
)

// TestMain allows us to do setup/teardown for all tests
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Exit with the test result code
	os.Exit(code)
}
