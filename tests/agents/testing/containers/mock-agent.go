package main

func main() {
	// platform/testing/mock_agent.go
	type MockAgent struct {
		Type      string
		Responses map[string]Response
	}

	func (m *MockAgent) Start() {
		// Listen on Kafka topic
		// Auto-respond based on configuration
	}

}
