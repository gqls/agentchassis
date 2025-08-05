package main

func main() {
	// platform/testing/containers.go
	func SetupTestKafka(t *testing.T) string {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
		Image: "confluentinc/cp-kafka:latest",
		// ... configuration
	}
		// Return broker address
	}

	func SetupTestPostgres(t *testing.T) string {
		// Similar for Postgres
	}
}
