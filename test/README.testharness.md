# Run it from root and use -C test
make -C test build-test-harness
make -C test push-test-images
kubectl delete job test-harness -n ai-persona-system
make -C test harness-integration

make harness-unit
make harness-integration
make harness-e2e
make harness-performance

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

