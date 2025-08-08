3. Alternative: If you want to see the raw test output without gotestsum:
# Run with environment variable to disable gotestsum
kubectl -n ai-persona-system delete job test-harness 2>/dev/null || true
kubectl -n ai-persona-system create job test-harness \
--from=cronjob/test-harness-unit \
--env="USE_GOTESTSUM=false"

# Watch logs
kubectl -n ai-persona-system logs -l app=test-harness -f

Increase code coverage (currently at 43.5%):
# View detailed coverage report
kubectl -n ai-persona-system exec -it deployment/test-harness -- \
go test -coverprofile=coverage.out ./test/unit/...
kubectl -n ai-persona-system exec -it deployment/test-harness -- \
go tool cover -html=coverage.out -o coverage.html

make -C test build-test-harness
make -C test push-test-images
kubectl delete job test-harness -n ai-persona-system
make -C test harness-unit
make -C test harness-run TEST_SUITE=integration