# Build and push test harness image
make harness-build

# Run all tests in Kubernetes
make harness-run

# Run specific test suite
make harness-unit
make harness-integration
make harness-e2e
make harness-performance

# Open interactive shell for debugging
make harness-shell

# View test logs
make harness-logs

# Copy test results to local machine
make harness-results

# Clean up
make harness-clean